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

// EdgeRepository handles dependency edge operations
type EdgeRepository struct {
	*BaseRepository
}

// NewEdgeRepository creates a new edge repository
func NewEdgeRepository(db *DB) *EdgeRepository {
	return &EdgeRepository{
		BaseRepository: NewBaseRepository(db.pool, db.sb),
	}
}

// NewEdgeRepositoryWithTx creates a new edge repository with a transaction
func NewEdgeRepositoryWithTx(tx pgx.Tx, sb squirrel.StatementBuilderType) *EdgeRepository {
	return &EdgeRepository{
		BaseRepository: NewBaseRepository(tx, sb),
	}
}

// Create creates a new dependency edge
func (r *EdgeRepository) Create(ctx context.Context, edge *models.DependencyEdge) error {
	if edge.ID == uuid.Nil {
		edge.ID = uuid.New()
	}

	parametersJSON, err := json.Marshal(edge.DefaultParameters)
	if err != nil {
		return fmt.Errorf("failed to marshal default parameters: %w", err)
	}

	query := r.QueryBuilder().
		Insert("dependency_edges").
		Columns("id", "parent_id", "child_id", "default_strategy", "default_parameters", "active_from", "active_to").
		Values(edge.ID, edge.ParentID, edge.ChildID, edge.DefaultStrategy, parametersJSON, edge.ActiveFrom, edge.ActiveTo).
		Suffix("RETURNING created_at, updated_at")

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&edge.CreatedAt, &edge.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create edge: %w", err)
	}

	return nil
}

// GetByID retrieves a dependency edge by ID
func (r *EdgeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.DependencyEdge, error) {
	query := r.QueryBuilder().
		Select("id", "parent_id", "child_id", "default_strategy", "default_parameters", "active_from", "active_to", "created_at", "updated_at").
		From("dependency_edges").
		Where(squirrel.Eq{"id": id})

	row := r.QueryRow(ctx, query)

	var edge models.DependencyEdge
	var parametersJSON []byte

	err := row.Scan(
		&edge.ID,
		&edge.ParentID,
		&edge.ChildID,
		&edge.DefaultStrategy,
		&parametersJSON,
		&edge.ActiveFrom,
		&edge.ActiveTo,
		&edge.CreatedAt,
		&edge.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("edge not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get edge: %w", err)
	}

	if err := json.Unmarshal(parametersJSON, &edge.DefaultParameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal default parameters: %w", err)
	}

	return &edge, nil
}

// GetActiveEdgesForDate retrieves all active edges for a specific date
func (r *EdgeRepository) GetActiveEdgesForDate(ctx context.Context, date time.Time) ([]models.DependencyEdge, error) {
	query := r.QueryBuilder().
		Select("id", "parent_id", "child_id", "default_strategy", "default_parameters", "active_from", "active_to", "created_at", "updated_at").
		From("dependency_edges").
		Where(squirrel.LtOrEq{"active_from": date}).
		Where(squirrel.Or{
			squirrel.Eq{"active_to": nil},
			squirrel.GtOrEq{"active_to": date},
		}).
		OrderBy("parent_id, child_id")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active edges: %w", err)
	}
	defer rows.Close()

	var edges []models.DependencyEdge
	for rows.Next() {
		var edge models.DependencyEdge
		var parametersJSON []byte

		err := rows.Scan(
			&edge.ID,
			&edge.ParentID,
			&edge.ChildID,
			&edge.DefaultStrategy,
			&parametersJSON,
			&edge.ActiveFrom,
			&edge.ActiveTo,
			&edge.CreatedAt,
			&edge.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}

		if err := json.Unmarshal(parametersJSON, &edge.DefaultParameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default parameters: %w", err)
		}

		edges = append(edges, edge)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating edges: %w", err)
	}

	return edges, nil
}

// GetByParentID retrieves all edges where the given node is the parent
func (r *EdgeRepository) GetByParentID(ctx context.Context, parentID uuid.UUID, date *time.Time) ([]models.DependencyEdge, error) {
	query := r.QueryBuilder().
		Select("id", "parent_id", "child_id", "default_strategy", "default_parameters", "active_from", "active_to", "created_at", "updated_at").
		From("dependency_edges").
		Where(squirrel.Eq{"parent_id": parentID})

	if date != nil {
		query = query.
			Where(squirrel.LtOrEq{"active_from": *date}).
			Where(squirrel.Or{
				squirrel.Eq{"active_to": nil},
				squirrel.GtOrEq{"active_to": *date},
			})
	}

	query = query.OrderBy("child_id")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges by parent: %w", err)
	}
	defer rows.Close()

	var edges []models.DependencyEdge
	for rows.Next() {
		var edge models.DependencyEdge
		var parametersJSON []byte

		err := rows.Scan(
			&edge.ID,
			&edge.ParentID,
			&edge.ChildID,
			&edge.DefaultStrategy,
			&parametersJSON,
			&edge.ActiveFrom,
			&edge.ActiveTo,
			&edge.CreatedAt,
			&edge.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}

		if err := json.Unmarshal(parametersJSON, &edge.DefaultParameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default parameters: %w", err)
		}

		edges = append(edges, edge)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating edges: %w", err)
	}

	return edges, nil
}

// GetByChildID retrieves all edges where the given node is the child
func (r *EdgeRepository) GetByChildID(ctx context.Context, childID uuid.UUID, date *time.Time) ([]models.DependencyEdge, error) {
	query := r.QueryBuilder().
		Select("id", "parent_id", "child_id", "default_strategy", "default_parameters", "active_from", "active_to", "created_at", "updated_at").
		From("dependency_edges").
		Where(squirrel.Eq{"child_id": childID})

	if date != nil {
		query = query.
			Where(squirrel.LtOrEq{"active_from": *date}).
			Where(squirrel.Or{
				squirrel.Eq{"active_to": nil},
				squirrel.GtOrEq{"active_to": *date},
			})
	}

	query = query.OrderBy("parent_id")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges by child: %w", err)
	}
	defer rows.Close()

	var edges []models.DependencyEdge
	for rows.Next() {
		var edge models.DependencyEdge
		var parametersJSON []byte

		err := rows.Scan(
			&edge.ID,
			&edge.ParentID,
			&edge.ChildID,
			&edge.DefaultStrategy,
			&parametersJSON,
			&edge.ActiveFrom,
			&edge.ActiveTo,
			&edge.CreatedAt,
			&edge.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}

		if err := json.Unmarshal(parametersJSON, &edge.DefaultParameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default parameters: %w", err)
		}

		edges = append(edges, edge)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating edges: %w", err)
	}

	return edges, nil
}

// Update updates an existing dependency edge
func (r *EdgeRepository) Update(ctx context.Context, edge *models.DependencyEdge) error {
	parametersJSON, err := json.Marshal(edge.DefaultParameters)
	if err != nil {
		return fmt.Errorf("failed to marshal default parameters: %w", err)
	}

	query := r.QueryBuilder().
		Update("dependency_edges").
		Set("parent_id", edge.ParentID).
		Set("child_id", edge.ChildID).
		Set("default_strategy", edge.DefaultStrategy).
		Set("default_parameters", parametersJSON).
		Set("active_from", edge.ActiveFrom).
		Set("active_to", edge.ActiveTo).
		Where(squirrel.Eq{"id": edge.ID}).
		Suffix("RETURNING updated_at")

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&edge.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("edge not found: %s", edge.ID)
		}
		return fmt.Errorf("failed to update edge: %w", err)
	}

	return nil
}

// Delete deletes a dependency edge
func (r *EdgeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := r.QueryBuilder().
		Delete("dependency_edges").
		Where(squirrel.Eq{"id": id})

	tag, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete edge: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("edge not found: %s", id)
	}

	return nil
}

// GetStrategiesForEdge retrieves all dimension-specific strategies for an edge
func (r *EdgeRepository) GetStrategiesForEdge(ctx context.Context, edgeID uuid.UUID) ([]models.EdgeStrategy, error) {
	query := r.QueryBuilder().
		Select("id", "edge_id", "dimension", "strategy", "parameters", "created_at", "updated_at").
		From("edge_strategies").
		Where(squirrel.Eq{"edge_id": edgeID}).
		OrderBy("dimension")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get edge strategies: %w", err)
	}
	defer rows.Close()

	var strategies []models.EdgeStrategy
	for rows.Next() {
		var strategy models.EdgeStrategy
		var parametersJSON []byte

		err := rows.Scan(
			&strategy.ID,
			&strategy.EdgeID,
			&strategy.Dimension,
			&strategy.Strategy,
			&parametersJSON,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge strategy: %w", err)
		}

		if err := json.Unmarshal(parametersJSON, &strategy.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal strategy parameters: %w", err)
		}

		strategies = append(strategies, strategy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating edge strategies: %w", err)
	}

	return strategies, nil
}

// CreateStrategy creates a new edge strategy
func (r *EdgeRepository) CreateStrategy(ctx context.Context, strategy *models.EdgeStrategy) error {
	if strategy.ID == uuid.Nil {
		strategy.ID = uuid.New()
	}

	parametersJSON, err := json.Marshal(strategy.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal strategy parameters: %w", err)
	}

	query := r.QueryBuilder().
		Insert("edge_strategies").
		Columns("id", "edge_id", "dimension", "strategy", "parameters").
		Values(strategy.ID, strategy.EdgeID, strategy.Dimension, strategy.Strategy, parametersJSON).
		Suffix("RETURNING created_at, updated_at")

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&strategy.CreatedAt, &strategy.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create edge strategy: %w", err)
	}

	return nil
}
