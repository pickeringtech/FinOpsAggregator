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

// UsageRepository handles node usage operations
type UsageRepository struct {
	*BaseRepository
}

// NewUsageRepository creates a new usage repository
func NewUsageRepository(db *DB) *UsageRepository {
	return &UsageRepository{
		BaseRepository: NewBaseRepository(db.pool, db.sb),
	}
}

// NewUsageRepositoryWithTx creates a new usage repository with a transaction
func NewUsageRepositoryWithTx(tx pgx.Tx, sb squirrel.StatementBuilderType) *UsageRepository {
	return &UsageRepository{
		BaseRepository: NewBaseRepository(tx, sb),
	}
}

// Upsert creates or updates a node usage record
func (r *UsageRepository) Upsert(ctx context.Context, usage *models.NodeUsageByDimension) error {
	query := r.QueryBuilder().
		Insert("node_usage_by_dimension").
		Columns("node_id", "usage_date", "metric", "value", "unit").
		Values(usage.NodeID, usage.UsageDate, usage.Metric, usage.Value, usage.Unit).
		Suffix(`ON CONFLICT (node_id, usage_date, metric) 
			DO UPDATE SET 
				value = EXCLUDED.value,
				unit = EXCLUDED.unit,
				updated_at = now()
			RETURNING created_at, updated_at`)

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&usage.CreatedAt, &usage.UpdatedAt); err != nil {
		return fmt.Errorf("failed to upsert usage: %w", err)
	}

	return nil
}

// GetByNodeAndDateRange retrieves usage for a node within a date range
func (r *UsageRepository) GetByNodeAndDateRange(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, metrics []string) ([]models.NodeUsageByDimension, error) {
	query := r.QueryBuilder().
		Select("node_id", "usage_date", "metric", "value", "unit", "created_at", "updated_at").
		From("node_usage_by_dimension").
		Where(squirrel.Eq{"node_id": nodeID}).
		Where(squirrel.GtOrEq{"usage_date": startDate}).
		Where(squirrel.LtOrEq{"usage_date": endDate})

	if len(metrics) > 0 {
		query = query.Where(squirrel.Eq{"metric": metrics})
	}

	query = query.OrderBy("usage_date, metric")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by node and date range: %w", err)
	}
	defer rows.Close()

	var usages []models.NodeUsageByDimension
	for rows.Next() {
		var usage models.NodeUsageByDimension

		err := rows.Scan(
			&usage.NodeID,
			&usage.UsageDate,
			&usage.Metric,
			&usage.Value,
			&usage.Unit,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage: %w", err)
		}

		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage: %w", err)
	}

	return usages, nil
}

// GetByDateRange retrieves all usage within a date range
func (r *UsageRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, metrics []string) ([]models.NodeUsageByDimension, error) {
	query := r.QueryBuilder().
		Select("node_id", "usage_date", "metric", "value", "unit", "created_at", "updated_at").
		From("node_usage_by_dimension").
		Where(squirrel.GtOrEq{"usage_date": startDate}).
		Where(squirrel.LtOrEq{"usage_date": endDate})

	if len(metrics) > 0 {
		query = query.Where(squirrel.Eq{"metric": metrics})
	}

	query = query.OrderBy("node_id, usage_date, metric")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by date range: %w", err)
	}
	defer rows.Close()

	var usages []models.NodeUsageByDimension
	for rows.Next() {
		var usage models.NodeUsageByDimension

		err := rows.Scan(
			&usage.NodeID,
			&usage.UsageDate,
			&usage.Metric,
			&usage.Value,
			&usage.Unit,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage: %w", err)
		}

		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage: %w", err)
	}

	return usages, nil
}

// GetByDate retrieves all usage for a specific date
func (r *UsageRepository) GetByDate(ctx context.Context, date time.Time, metrics []string) ([]models.NodeUsageByDimension, error) {
	query := r.QueryBuilder().
		Select("node_id", "usage_date", "metric", "value", "unit", "created_at", "updated_at").
		From("node_usage_by_dimension").
		Where(squirrel.Eq{"usage_date": date})

	if len(metrics) > 0 {
		query = query.Where(squirrel.Eq{"metric": metrics})
	}

	query = query.OrderBy("node_id, metric")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by date: %w", err)
	}
	defer rows.Close()

	var usages []models.NodeUsageByDimension
	for rows.Next() {
		var usage models.NodeUsageByDimension

		err := rows.Scan(
			&usage.NodeID,
			&usage.UsageDate,
			&usage.Metric,
			&usage.Value,
			&usage.Unit,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage: %w", err)
		}

		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage: %w", err)
	}

	return usages, nil
}

// GetSummaryByNodeAndDateRange retrieves usage summaries aggregated by node and metric
func (r *UsageRepository) GetSummaryByNodeAndDateRange(ctx context.Context, startDate, endDate time.Time, metrics []string) ([]UsageSummary, error) {
	query := r.QueryBuilder().
		Select("node_id", "metric", "unit", "SUM(value) as total_value", "AVG(value) as avg_value", "COUNT(*) as day_count").
		From("node_usage_by_dimension").
		Where(squirrel.GtOrEq{"usage_date": startDate}).
		Where(squirrel.LtOrEq{"usage_date": endDate}).
		GroupBy("node_id", "metric", "unit")

	if len(metrics) > 0 {
		query = query.Where(squirrel.Eq{"metric": metrics})
	}

	query = query.OrderBy("node_id, metric")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}
	defer rows.Close()

	var summaries []UsageSummary
	for rows.Next() {
		var summary UsageSummary

		err := rows.Scan(
			&summary.NodeID,
			&summary.Metric,
			&summary.Unit,
			&summary.TotalValue,
			&summary.AvgValue,
			&summary.DayCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage summary: %w", err)
		}

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage summaries: %w", err)
	}

	return summaries, nil
}

// Delete deletes usage records for a node within a date range
func (r *UsageRepository) Delete(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, metrics []string) error {
	query := r.QueryBuilder().
		Delete("node_usage_by_dimension").
		Where(squirrel.Eq{"node_id": nodeID}).
		Where(squirrel.GtOrEq{"usage_date": startDate}).
		Where(squirrel.LtOrEq{"usage_date": endDate})

	if len(metrics) > 0 {
		query = query.Where(squirrel.Eq{"metric": metrics})
	}

	tag, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete usage: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no usage found to delete")
	}

	return nil
}

// BulkUpsert efficiently inserts or updates multiple usage records
func (r *UsageRepository) BulkUpsert(ctx context.Context, usages []models.NodeUsageByDimension) error {
	if len(usages) == 0 {
		return nil
	}

	query := r.QueryBuilder().
		Insert("node_usage_by_dimension").
		Columns("node_id", "usage_date", "metric", "value", "unit")

	for _, usage := range usages {
		query = query.Values(usage.NodeID, usage.UsageDate, usage.Metric, usage.Value, usage.Unit)
	}

	query = query.Suffix(`ON CONFLICT (node_id, usage_date, metric) 
		DO UPDATE SET 
			value = EXCLUDED.value,
			unit = EXCLUDED.unit,
			updated_at = now()`)

	_, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to bulk upsert usage: %w", err)
	}

	return nil
}

// UsageSummary represents aggregated usage data
type UsageSummary struct {
	NodeID     uuid.UUID `db:"node_id"`
	Metric     string    `db:"metric"`
	Unit       string    `db:"unit"`
	TotalValue string    `db:"total_value"` // Using string to handle decimal precision
	AvgValue   string    `db:"avg_value"`   // Using string to handle decimal precision
	DayCount   int       `db:"day_count"`
}
