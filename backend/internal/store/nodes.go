package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
)

// NodeRepository handles cost node operations
type NodeRepository struct {
	*BaseRepository
}

// NewNodeRepository creates a new node repository
func NewNodeRepository(db *DB) *NodeRepository {
	return &NodeRepository{
		BaseRepository: NewBaseRepository(db.pool, db.sb),
	}
}

// NewNodeRepositoryWithTx creates a new node repository with a transaction
func NewNodeRepositoryWithTx(tx pgx.Tx, sb squirrel.StatementBuilderType) *NodeRepository {
	return &NodeRepository{
		BaseRepository: NewBaseRepository(tx, sb),
	}
}

// Create creates a new cost node
func (r *NodeRepository) Create(ctx context.Context, node *models.CostNode) error {
	if node.ID == uuid.Nil {
		node.ID = uuid.New()
	}

	costLabelsJSON, err := json.Marshal(node.CostLabels)
	if err != nil {
		return fmt.Errorf("failed to marshal cost labels: %w", err)
	}

	metadataJSON, err := json.Marshal(node.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := r.QueryBuilder().
		Insert("cost_nodes").
		Columns("id", "name", "type", "cost_labels", "is_platform", "metadata").
		Values(node.ID, node.Name, node.Type, costLabelsJSON, node.IsPlatform, metadataJSON).
		Suffix("RETURNING created_at, updated_at")

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&node.CreatedAt, &node.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	return nil
}

// GetByID retrieves a cost node by ID
func (r *NodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CostNode, error) {
	query := r.QueryBuilder().
		Select("id", "name", "type", "cost_labels", "is_platform", "metadata", "created_at", "updated_at", "archived_at").
		From("cost_nodes").
		Where(squirrel.Eq{"id": id})

	row := r.QueryRow(ctx, query)

	var node models.CostNode
	var costLabelsJSON, metadataJSON []byte

	err := row.Scan(
		&node.ID,
		&node.Name,
		&node.Type,
		&costLabelsJSON,
		&node.IsPlatform,
		&metadataJSON,
		&node.CreatedAt,
		&node.UpdatedAt,
		&node.ArchivedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("node not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if err := json.Unmarshal(costLabelsJSON, &node.CostLabels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cost labels: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &node.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &node, nil
}

// GetByName retrieves a cost node by name
func (r *NodeRepository) GetByName(ctx context.Context, name string) (*models.CostNode, error) {
	query := r.QueryBuilder().
		Select("id", "name", "type", "cost_labels", "is_platform", "metadata", "created_at", "updated_at", "archived_at").
		From("cost_nodes").
		Where(squirrel.Eq{"name": name}).
		Where(squirrel.Eq{"archived_at": nil})

	row := r.QueryRow(ctx, query)

	var node models.CostNode
	var costLabelsJSON, metadataJSON []byte

	err := row.Scan(
		&node.ID,
		&node.Name,
		&node.Type,
		&costLabelsJSON,
		&node.IsPlatform,
		&metadataJSON,
		&node.CreatedAt,
		&node.UpdatedAt,
		&node.ArchivedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("node not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if err := json.Unmarshal(costLabelsJSON, &node.CostLabels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cost labels: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &node.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &node, nil
}

// List retrieves all cost nodes with optional filtering
func (r *NodeRepository) List(ctx context.Context, filters NodeFilters) ([]models.CostNode, error) {
	query := r.QueryBuilder().
		Select("id", "name", "type", "cost_labels", "is_platform", "metadata", "created_at", "updated_at", "archived_at").
		From("cost_nodes")

	// Apply filters
	if filters.Type != "" {
		query = query.Where(squirrel.Eq{"type": filters.Type})
	}
	if filters.IsPlatform != nil {
		query = query.Where(squirrel.Eq{"is_platform": *filters.IsPlatform})
	}
	if !filters.IncludeArchived {
		query = query.Where(squirrel.Eq{"archived_at": nil})
	}

	// Apply ordering
	query = query.OrderBy("name ASC")

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(uint64(filters.Limit))
	}
	if filters.Offset > 0 {
		query = query.Offset(uint64(filters.Offset))
	}

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []models.CostNode
	for rows.Next() {
		var node models.CostNode
		var costLabelsJSON, metadataJSON []byte

		err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.Type,
			&costLabelsJSON,
			&node.IsPlatform,
			&metadataJSON,
			&node.CreatedAt,
			&node.UpdatedAt,
			&node.ArchivedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		if err := json.Unmarshal(costLabelsJSON, &node.CostLabels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cost labels: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &node.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		nodes = append(nodes, node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating nodes: %w", err)
	}

	return nodes, nil
}

// Update updates an existing cost node
func (r *NodeRepository) Update(ctx context.Context, node *models.CostNode) error {
	costLabelsJSON, err := json.Marshal(node.CostLabels)
	if err != nil {
		return fmt.Errorf("failed to marshal cost labels: %w", err)
	}

	metadataJSON, err := json.Marshal(node.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := r.QueryBuilder().
		Update("cost_nodes").
		Set("name", node.Name).
		Set("type", node.Type).
		Set("cost_labels", costLabelsJSON).
		Set("is_platform", node.IsPlatform).
		Set("metadata", metadataJSON).
		Where(squirrel.Eq{"id": node.ID}).
		Suffix("RETURNING updated_at")

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&node.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("node not found: %s", node.ID)
		}
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

// Delete soft deletes a cost node by setting archived_at
func (r *NodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := r.QueryBuilder().
		Update("cost_nodes").
		Set("archived_at", "now()").
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"archived_at": nil})

	tag, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("node not found or already deleted: %s", id)
	}

	return nil
}

// NodeFilters represents filtering options for listing nodes
type NodeFilters struct {
	Type            string
	IsPlatform      *bool
	IncludeArchived bool
	Limit           int
	Offset          int
}
