package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
)

// CostRepository handles node cost operations
type CostRepository struct {
	*BaseRepository
}

// NewCostRepository creates a new cost repository
func NewCostRepository(db *DB) *CostRepository {
	return &CostRepository{
		BaseRepository: NewBaseRepository(db.pool, db.sb),
	}
}

// NewCostRepositoryWithTx creates a new cost repository with a transaction
func NewCostRepositoryWithTx(tx pgx.Tx, sb squirrel.StatementBuilderType) *CostRepository {
	return &CostRepository{
		BaseRepository: NewBaseRepository(tx, sb),
	}
}

// Upsert creates or updates a node cost record
func (r *CostRepository) Upsert(ctx context.Context, cost *models.NodeCostByDimension) error {
	metadataJSON, err := json.Marshal(cost.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := r.QueryBuilder().
		Insert("node_costs_by_dimension").
		Columns("node_id", "cost_date", "dimension", "amount", "currency", "metadata").
		Values(cost.NodeID, cost.CostDate, cost.Dimension, cost.Amount, cost.Currency, metadataJSON).
		Suffix(`ON CONFLICT (node_id, cost_date, dimension) 
			DO UPDATE SET 
				amount = EXCLUDED.amount,
				currency = EXCLUDED.currency,
				metadata = EXCLUDED.metadata,
				updated_at = now()
			RETURNING created_at, updated_at`)

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&cost.CreatedAt, &cost.UpdatedAt); err != nil {
		return fmt.Errorf("failed to upsert cost: %w", err)
	}

	return nil
}

// GetByNodeAndDateRange retrieves costs for a node within a date range
func (r *CostRepository) GetByNodeAndDateRange(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.NodeCostByDimension, error) {
	query := r.QueryBuilder().
		Select("node_id", "cost_date", "dimension", "amount", "currency", "metadata", "created_at", "updated_at").
		From("node_costs_by_dimension").
		Where(squirrel.Eq{"node_id": nodeID}).
		Where(squirrel.GtOrEq{"cost_date": startDate}).
		Where(squirrel.LtOrEq{"cost_date": endDate})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("cost_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by node and date range: %w", err)
	}
	defer rows.Close()

	var costs []models.NodeCostByDimension
	for rows.Next() {
		var cost models.NodeCostByDimension
		var metadataJSON []byte

		err := rows.Scan(
			&cost.NodeID,
			&cost.CostDate,
			&cost.Dimension,
			&cost.Amount,
			&cost.Currency,
			&metadataJSON,
			&cost.CreatedAt,
			&cost.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &cost.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		costs = append(costs, cost)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating costs: %w", err)
	}

	return costs, nil
}

// GetByDateRange retrieves all costs within a date range
func (r *CostRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, dimensions []string) ([]models.NodeCostByDimension, error) {
	query := r.QueryBuilder().
		Select("node_id", "cost_date", "dimension", "amount", "currency", "metadata", "created_at", "updated_at").
		From("node_costs_by_dimension").
		Where(squirrel.GtOrEq{"cost_date": startDate}).
		Where(squirrel.LtOrEq{"cost_date": endDate})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("node_id, cost_date, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by date range: %w", err)
	}
	defer rows.Close()

	var costs []models.NodeCostByDimension
	for rows.Next() {
		var cost models.NodeCostByDimension
		var metadataJSON []byte

		err := rows.Scan(
			&cost.NodeID,
			&cost.CostDate,
			&cost.Dimension,
			&cost.Amount,
			&cost.Currency,
			&metadataJSON,
			&cost.CreatedAt,
			&cost.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &cost.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		costs = append(costs, cost)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating costs: %w", err)
	}

	return costs, nil
}

// GetByDate retrieves all costs for a specific date
func (r *CostRepository) GetByDate(ctx context.Context, date time.Time, dimensions []string) ([]models.NodeCostByDimension, error) {
	query := r.QueryBuilder().
		Select("node_id", "cost_date", "dimension", "amount", "currency", "metadata", "created_at", "updated_at").
		From("node_costs_by_dimension").
		Where(squirrel.Eq{"cost_date": date})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("node_id, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by date: %w", err)
	}
	defer rows.Close()

	var costs []models.NodeCostByDimension
	for rows.Next() {
		var cost models.NodeCostByDimension
		var metadataJSON []byte

		err := rows.Scan(
			&cost.NodeID,
			&cost.CostDate,
			&cost.Dimension,
			&cost.Amount,
			&cost.Currency,
			&metadataJSON,
			&cost.CreatedAt,
			&cost.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &cost.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		costs = append(costs, cost)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating costs: %w", err)
	}

	return costs, nil
}

// GetSummaryByNodeAndDateRange retrieves cost summaries aggregated by node and dimension
func (r *CostRepository) GetSummaryByNodeAndDateRange(ctx context.Context, startDate, endDate time.Time, dimensions []string) ([]CostSummary, error) {
	query := r.QueryBuilder().
		Select("node_id", "dimension", "currency", "SUM(amount) as total_amount", "COUNT(*) as day_count").
		From("node_costs_by_dimension").
		Where(squirrel.GtOrEq{"cost_date": startDate}).
		Where(squirrel.LtOrEq{"cost_date": endDate}).
		GroupBy("node_id", "dimension", "currency")

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	query = query.OrderBy("node_id, dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %w", err)
	}
	defer rows.Close()

	var summaries []CostSummary
	for rows.Next() {
		var summary CostSummary

		err := rows.Scan(
			&summary.NodeID,
			&summary.Dimension,
			&summary.Currency,
			&summary.TotalAmount,
			&summary.DayCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost summary: %w", err)
		}

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cost summaries: %w", err)
	}

	return summaries, nil
}

// Delete deletes cost records for a node within a date range
func (r *CostRepository) Delete(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimensions []string) error {
	query := r.QueryBuilder().
		Delete("node_costs_by_dimension").
		Where(squirrel.Eq{"node_id": nodeID}).
		Where(squirrel.GtOrEq{"cost_date": startDate}).
		Where(squirrel.LtOrEq{"cost_date": endDate})

	if len(dimensions) > 0 {
		query = query.Where(squirrel.Eq{"dimension": dimensions})
	}

	tag, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete costs: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no costs found to delete")
	}

	return nil
}

// BulkUpsert efficiently inserts or updates multiple cost records
func (r *CostRepository) BulkUpsert(ctx context.Context, costs []models.NodeCostByDimension) error {
	if len(costs) == 0 {
		return nil
	}

	query := r.QueryBuilder().
		Insert("node_costs_by_dimension").
		Columns("node_id", "cost_date", "dimension", "amount", "currency", "metadata")

	for _, cost := range costs {
		metadataJSON, err := json.Marshal(cost.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		query = query.Values(cost.NodeID, cost.CostDate, cost.Dimension, cost.Amount, cost.Currency, metadataJSON)
	}

	query = query.Suffix(`ON CONFLICT (node_id, cost_date, dimension) 
		DO UPDATE SET 
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			metadata = EXCLUDED.metadata,
			updated_at = now()`)

	_, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to bulk upsert costs: %w", err)
	}

	return nil
}

// CostSummary represents aggregated cost data
type CostSummary struct {
	NodeID      uuid.UUID `db:"node_id"`
	Dimension   string    `db:"dimension"`
	Currency    string    `db:"currency"`
	TotalAmount string    `db:"total_amount"` // Using string to handle decimal precision
	DayCount    int       `db:"day_count"`
}
