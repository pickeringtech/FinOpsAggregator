package store

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
)

// RunRepository handles computation run operations
type RunRepository struct {
	*BaseRepository
}

// NewRunRepository creates a new run repository
func NewRunRepository(db *DB) *RunRepository {
	return &RunRepository{
		BaseRepository: NewBaseRepository(db.pool, db.sb),
	}
}

// NewRunRepositoryWithTx creates a new run repository with a transaction
func NewRunRepositoryWithTx(tx pgx.Tx, sb squirrel.StatementBuilderType) *RunRepository {
	return &RunRepository{
		BaseRepository: NewBaseRepository(tx, sb),
	}
}

// Create creates a new computation run
func (r *RunRepository) Create(ctx context.Context, run *models.ComputationRun) error {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}

	query := r.QueryBuilder().
		Insert("computation_runs").
		Columns("id", "window_start", "window_end", "graph_hash", "status", "notes").
		Values(run.ID, run.WindowStart, run.WindowEnd, run.GraphHash, run.Status, run.Notes).
		Suffix("RETURNING created_at, updated_at")

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&run.CreatedAt, &run.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create computation run: %w", err)
	}

	return nil
}

// GetByID retrieves a computation run by ID
func (r *RunRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ComputationRun, error) {
	query := r.QueryBuilder().
		Select("id", "created_at", "updated_at", "window_start", "window_end", "graph_hash", "status", "notes").
		From("computation_runs").
		Where(squirrel.Eq{"id": id})

	row := r.QueryRow(ctx, query)

	var run models.ComputationRun
	err := row.Scan(
		&run.ID,
		&run.CreatedAt,
		&run.UpdatedAt,
		&run.WindowStart,
		&run.WindowEnd,
		&run.GraphHash,
		&run.Status,
		&run.Notes,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("computation run not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get computation run: %w", err)
	}

	return &run, nil
}

// List retrieves computation runs with optional filtering
func (r *RunRepository) List(ctx context.Context, filters RunFilters) ([]models.ComputationRun, error) {
	query := r.QueryBuilder().
		Select("id", "created_at", "updated_at", "window_start", "window_end", "graph_hash", "status", "notes").
		From("computation_runs")

	// Apply filters
	if filters.Status != "" {
		query = query.Where(squirrel.Eq{"status": filters.Status})
	}
	if !filters.WindowStart.IsZero() {
		query = query.Where(squirrel.GtOrEq{"window_start": filters.WindowStart})
	}
	if !filters.WindowEnd.IsZero() {
		query = query.Where(squirrel.LtOrEq{"window_end": filters.WindowEnd})
	}
	if filters.GraphHash != "" {
		query = query.Where(squirrel.Eq{"graph_hash": filters.GraphHash})
	}

	// Apply ordering
	query = query.OrderBy("created_at DESC")

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(uint64(filters.Limit))
	}
	if filters.Offset > 0 {
		query = query.Offset(uint64(filters.Offset))
	}

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list computation runs: %w", err)
	}
	defer rows.Close()

	var runs []models.ComputationRun
	for rows.Next() {
		var run models.ComputationRun

		err := rows.Scan(
			&run.ID,
			&run.CreatedAt,
			&run.UpdatedAt,
			&run.WindowStart,
			&run.WindowEnd,
			&run.GraphHash,
			&run.Status,
			&run.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan computation run: %w", err)
		}

		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating computation runs: %w", err)
	}

	return runs, nil
}

// UpdateStatus updates the status of a computation run
func (r *RunRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, notes *string) error {
	query := r.QueryBuilder().
		Update("computation_runs").
		Set("status", status).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING updated_at")

	if notes != nil {
		query = query.Set("notes", *notes)
	}

	row := r.QueryRow(ctx, query)
	var updatedAt time.Time
	if err := row.Scan(&updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("computation run not found: %s", id)
		}
		return fmt.Errorf("failed to update computation run status: %w", err)
	}

	return nil
}

// Delete deletes a computation run and all associated results
func (r *RunRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := r.QueryBuilder().
		Delete("computation_runs").
		Where(squirrel.Eq{"id": id})

	tag, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete computation run: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("computation run not found: %s", id)
	}

	return nil
}

// SaveAllocationResults saves allocation results for a computation run
func (r *RunRepository) SaveAllocationResults(ctx context.Context, results []models.AllocationResultByDimension) error {
	if len(results) == 0 {
		return nil
	}

	query := r.QueryBuilder().
		Insert("allocation_results_by_dimension").
		Columns("run_id", "node_id", "allocation_date", "dimension", "direct_amount", "indirect_amount", "total_amount")

	for _, result := range results {
		query = query.Values(
			result.RunID,
			result.NodeID,
			result.AllocationDate,
			result.Dimension,
			result.DirectAmount,
			result.IndirectAmount,
			result.TotalAmount,
		)
	}

	_, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to save allocation results: %w", err)
	}

	return nil
}

// SaveContributionResults saves contribution results for a computation run
func (r *RunRepository) SaveContributionResults(ctx context.Context, results []models.ContributionResultByDimension) error {
	if len(results) == 0 {
		return nil
	}

	query := r.QueryBuilder().
		Insert("contribution_results_by_dimension").
		Columns("run_id", "parent_id", "child_id", "contribution_date", "dimension", "contributed_amount", "path")

	for _, result := range results {
		pathJSON := "[]"
		if len(result.Path) > 0 {
			// Convert UUID slice to JSON array
			pathStr := "["
			for i, id := range result.Path {
				if i > 0 {
					pathStr += ","
				}
				pathStr += fmt.Sprintf(`"%s"`, id.String())
			}
			pathStr += "]"
			pathJSON = pathStr
		}

		query = query.Values(
			result.RunID,
			result.ParentID,
			result.ChildID,
			result.ContributionDate,
			result.Dimension,
			result.ContributedAmount,
			pathJSON,
		)
	}

	_, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to save contribution results: %w", err)
	}

	return nil
}

// GetAllocationResults retrieves allocation results for a computation run
func (r *RunRepository) GetAllocationResults(ctx context.Context, runID uuid.UUID, filters AllocationResultFilters) ([]models.AllocationResultByDimension, error) {
	query := r.QueryBuilder().
		Select("run_id", "node_id", "allocation_date", "dimension", "direct_amount", "indirect_amount", "total_amount", "created_at", "updated_at").
		From("allocation_results_by_dimension").
		Where(squirrel.Eq{"run_id": runID})

	// Apply filters
	if filters.NodeID != uuid.Nil {
		query = query.Where(squirrel.Eq{"node_id": filters.NodeID})
	}
	if !filters.StartDate.IsZero() {
		query = query.Where(squirrel.GtOrEq{"allocation_date": filters.StartDate})
	}
	if !filters.EndDate.IsZero() {
		query = query.Where(squirrel.LtOrEq{"allocation_date": filters.EndDate})
	}
	if len(filters.Dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": filters.Dimensions})
	}

	query = query.OrderBy("node_id, allocation_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation results: %w", err)
	}
	defer rows.Close()

	var results []models.AllocationResultByDimension
	for rows.Next() {
		var result models.AllocationResultByDimension

		err := rows.Scan(
			&result.RunID,
			&result.NodeID,
			&result.AllocationDate,
			&result.Dimension,
			&result.DirectAmount,
			&result.IndirectAmount,
			&result.TotalAmount,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allocation result: %w", err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating allocation results: %w", err)
	}

	return results, nil
}

// GetContributionResults retrieves contribution results for a computation run
func (r *RunRepository) GetContributionResults(ctx context.Context, runID uuid.UUID, filters ContributionResultFilters) ([]models.ContributionResultByDimension, error) {
	query := r.QueryBuilder().
		Select("run_id", "parent_id", "child_id", "contribution_date", "dimension", "contributed_amount", "path", "created_at", "updated_at").
		From("contribution_results_by_dimension").
		Where(squirrel.Eq{"run_id": runID})

	// Apply filters
	if filters.ParentID != uuid.Nil {
		query = query.Where(squirrel.Eq{"parent_id": filters.ParentID})
	}
	if filters.ChildID != uuid.Nil {
		query = query.Where(squirrel.Eq{"child_id": filters.ChildID})
	}
	if !filters.StartDate.IsZero() {
		query = query.Where(squirrel.GtOrEq{"contribution_date": filters.StartDate})
	}
	if !filters.EndDate.IsZero() {
		query = query.Where(squirrel.LtOrEq{"contribution_date": filters.EndDate})
	}
	if len(filters.Dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": filters.Dimensions})
	}

	query = query.OrderBy("parent_id, child_id, contribution_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contribution results: %w", err)
	}
	defer rows.Close()

	var results []models.ContributionResultByDimension
	for rows.Next() {
		var result models.ContributionResultByDimension
		var pathJSON string

		err := rows.Scan(
			&result.RunID,
			&result.ParentID,
			&result.ChildID,
			&result.ContributionDate,
			&result.Dimension,
			&result.ContributedAmount,
			&pathJSON,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contribution result: %w", err)
		}

		// TODO: Parse pathJSON back to []uuid.UUID
		// For now, leaving it empty
		result.Path = []uuid.UUID{}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contribution results: %w", err)
	}

	return results, nil
}

// RunFilters represents filtering options for listing computation runs
type RunFilters struct {
	Status      string
	WindowStart time.Time
	WindowEnd   time.Time
	GraphHash   string
	Limit       int
	Offset      int
}

// AllocationResultFilters represents filtering options for allocation results
type AllocationResultFilters struct {
	NodeID     uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
	Dimensions []string
}

// ContributionResultFilters represents filtering options for contribution results
type ContributionResultFilters struct {
	ParentID   uuid.UUID
	ChildID    uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
	Dimensions []string
}

// GetAllocationsByParentAndDateRange retrieves allocations where the given node is the parent (receiving allocations)
func (r *RunRepository) GetAllocationsByParentAndDateRange(ctx context.Context, parentID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.AllocationResultByDimension, error) {
	query := r.QueryBuilder().
		Select("run_id", "node_id", "allocation_date", "dimension", "direct_amount", "indirect_amount", "total_amount", "created_at", "updated_at").
		From("allocation_results_by_dimension").
		Where(squirrel.Eq{"node_id": parentID}).
		Where(squirrel.GtOrEq{"allocation_date": startDate}).
		Where(squirrel.LtOrEq{"allocation_date": endDate})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("allocation_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocations by parent: %w", err)
	}
	defer rows.Close()

	var results []models.AllocationResultByDimension
	for rows.Next() {
		var result models.AllocationResultByDimension

		err := rows.Scan(
			&result.RunID,
			&result.NodeID,
			&result.AllocationDate,
			&result.Dimension,
			&result.DirectAmount,
			&result.IndirectAmount,
			&result.TotalAmount,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allocation result: %w", err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating allocation results: %w", err)
	}

	return results, nil
}

// GetAllocationsByChildAndDateRange retrieves allocations where the given node is contributing to others
// This is a bit more complex as we need to find allocations where this node's costs were allocated
func (r *RunRepository) GetAllocationsByChildAndDateRange(ctx context.Context, childID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.AllocationResultByDimension, error) {
	// This would require joining with contribution results to find where this node contributed
	// For now, we'll return an empty slice and implement this later if needed
	return []models.AllocationResultByDimension{}, nil
}

// GetContributionsByParentAndDateRange retrieves contributions where the given node is the parent (receiving contributions)
func (r *RunRepository) GetContributionsByParentAndDateRange(ctx context.Context, parentID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.ContributionResultByDimension, error) {
	query := r.QueryBuilder().
		Select("run_id", "parent_id", "child_id", "contribution_date", "dimension", "contributed_amount", "path", "created_at", "updated_at").
		From("contribution_results_by_dimension").
		Where(squirrel.Eq{"parent_id": parentID}).
		Where(squirrel.GtOrEq{"contribution_date": startDate}).
		Where(squirrel.LtOrEq{"contribution_date": endDate})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("contribution_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions by parent: %w", err)
	}
	defer rows.Close()

	var results []models.ContributionResultByDimension
	for rows.Next() {
		var result models.ContributionResultByDimension
		var pathJSON string

		err := rows.Scan(
			&result.RunID,
			&result.ParentID,
			&result.ChildID,
			&result.ContributionDate,
			&result.Dimension,
			&result.ContributedAmount,
			&pathJSON,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contribution result: %w", err)
		}

		// TODO: Parse pathJSON back to []uuid.UUID
		result.Path = []uuid.UUID{}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contribution results: %w", err)
	}

	return results, nil
}

// GetContributionsByChildAndDateRange retrieves contributions where the given node is the child (contributing to others)
func (r *RunRepository) GetContributionsByChildAndDateRange(ctx context.Context, childID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.ContributionResultByDimension, error) {
	query := r.QueryBuilder().
		Select("run_id", "parent_id", "child_id", "contribution_date", "dimension", "contributed_amount", "path", "created_at", "updated_at").
		From("contribution_results_by_dimension").
		Where(squirrel.Eq{"child_id": childID}).
		Where(squirrel.GtOrEq{"contribution_date": startDate}).
		Where(squirrel.LtOrEq{"contribution_date": endDate})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("contribution_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions by child: %w", err)
	}
	defer rows.Close()

	var results []models.ContributionResultByDimension
	for rows.Next() {
		var result models.ContributionResultByDimension
		var pathJSON string

		err := rows.Scan(
			&result.RunID,
			&result.ParentID,
			&result.ChildID,
			&result.ContributionDate,
			&result.Dimension,
			&result.ContributedAmount,
			&pathJSON,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contribution result: %w", err)
		}

		// TODO: Parse pathJSON back to []uuid.UUID
		result.Path = []uuid.UUID{}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contribution results: %w", err)
	}

	return results, nil
}

// GetAllocationsByNodeAndDateRange retrieves allocation results for a specific node within a date range
func (r *RunRepository) GetAllocationsByNodeAndDateRange(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.AllocationResultByDimension, error) {
	query := r.QueryBuilder().
		Select("run_id", "node_id", "allocation_date", "dimension", "direct_amount", "indirect_amount", "total_amount", "created_at", "updated_at").
		From("allocation_results_by_dimension").
		Where(squirrel.Eq{"node_id": nodeID}).
		Where(squirrel.GtOrEq{"allocation_date": startDate}).
		Where(squirrel.LtOrEq{"allocation_date": endDate})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("allocation_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocations by node and date range: %w", err)
	}
	defer rows.Close()

	var results []models.AllocationResultByDimension
	for rows.Next() {
		var result models.AllocationResultByDimension

		err := rows.Scan(
			&result.RunID,
			&result.NodeID,
			&result.AllocationDate,
			&result.Dimension,
			&result.DirectAmount,
			&result.IndirectAmount,
			&result.TotalAmount,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allocation result: %w", err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating allocation results: %w", err)
	}

	return results, nil
}
