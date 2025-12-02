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

// UpsertWithLabels creates or updates a node usage record with labels
func (r *UsageRepository) UpsertWithLabels(ctx context.Context, usage *models.NodeUsageByDimension) error {
	query := r.QueryBuilder().
		Insert("node_usage_by_dimension").
		Columns("node_id", "usage_date", "metric", "value", "unit", "labels", "source").
		Values(usage.NodeID, usage.UsageDate, usage.Metric, usage.Value, usage.Unit, usage.Labels, usage.Source).
		Suffix(`ON CONFLICT (node_id, usage_date, metric)
			DO UPDATE SET
				value = EXCLUDED.value,
				unit = EXCLUDED.unit,
				labels = EXCLUDED.labels,
				source = EXCLUDED.source,
				updated_at = now()
			RETURNING created_at, updated_at`)

	row := r.QueryRow(ctx, query)
	if err := row.Scan(&usage.CreatedAt, &usage.UpdatedAt); err != nil {
		return fmt.Errorf("failed to upsert usage with labels: %w", err)
	}

	return nil
}

// QueryWithOptions retrieves usage metrics with flexible filtering options
// Supports filtering by labels using JSONB operators
func (r *UsageRepository) QueryWithOptions(ctx context.Context, opts models.UsageQueryOptions) ([]models.NodeUsageByDimension, error) {
	query := r.QueryBuilder().
		Select("node_id", "usage_date", "metric", "value", "unit", "labels", "source", "created_at", "updated_at").
		From("node_usage_by_dimension").
		Where(squirrel.GtOrEq{"usage_date": opts.StartDate}).
		Where(squirrel.LtOrEq{"usage_date": opts.EndDate})

	// Filter by node IDs
	if len(opts.NodeIDs) > 0 {
		query = query.Where(squirrel.Eq{"node_id": opts.NodeIDs})
	}

	// Filter by metrics
	if len(opts.Metrics) > 0 {
		query = query.Where(squirrel.Eq{"metric": opts.Metrics})
	}

	// Filter by source
	if opts.Source != "" {
		query = query.Where(squirrel.Eq{"source": opts.Source})
	}

	// Apply label filters using JSONB operators
	for _, filter := range opts.LabelFilters {
		switch filter.Operator {
		case "eq":
			if len(filter.Values) > 0 {
				// labels->>'key' = 'value'
				query = query.Where(squirrel.Expr("labels->>? = ?", filter.Key, filter.Values[0]))
			}
		case "neq":
			if len(filter.Values) > 0 {
				// labels->>'key' != 'value' OR labels->>'key' IS NULL
				query = query.Where(squirrel.Or{
					squirrel.Expr("labels->>? != ?", filter.Key, filter.Values[0]),
					squirrel.Expr("labels->>? IS NULL", filter.Key),
				})
			}
		case "in":
			if len(filter.Values) > 0 {
				// labels->>'key' IN ('value1', 'value2', ...)
				query = query.Where(squirrel.Eq{fmt.Sprintf("labels->>'%s'", filter.Key): filter.Values})
			}
		case "not_in":
			if len(filter.Values) > 0 {
				// labels->>'key' NOT IN ('value1', 'value2', ...) OR labels->>'key' IS NULL
				query = query.Where(squirrel.Or{
					squirrel.NotEq{fmt.Sprintf("labels->>'%s'", filter.Key): filter.Values},
					squirrel.Expr("labels->>? IS NULL", filter.Key),
				})
			}
		case "exists":
			// labels ? 'key'
			query = query.Where(squirrel.Expr("labels ?? ?", filter.Key))
		case "not_exists":
			// NOT (labels ? 'key')
			query = query.Where(squirrel.Expr("NOT (labels ?? ?)", filter.Key))
		}
	}

	query = query.OrderBy("node_id, usage_date, metric")

	rows, err := r.QueryRows(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage with options: %w", err)
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
			&usage.Labels,
			&usage.Source,
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

// GetSummaryByLabelValue retrieves usage summaries grouped by a specific label value
// Useful for getting usage breakdown by customer_id, environment, etc.
func (r *UsageRepository) GetSummaryByLabelValue(ctx context.Context, opts models.UsageQueryOptions) (map[string][]UsageSummary, error) {
	if opts.GroupByLabel == "" {
		return nil, fmt.Errorf("group_by_label is required")
	}

	// Build query with label grouping
	query := fmt.Sprintf(`
		SELECT
			node_id,
			metric,
			unit,
			labels->>'%s' as label_value,
			SUM(value) as total_value,
			AVG(value) as avg_value,
			COUNT(*) as day_count
		FROM node_usage_by_dimension
		WHERE usage_date >= $1 AND usage_date <= $2
		  AND labels->>'%s' IS NOT NULL
	`, opts.GroupByLabel, opts.GroupByLabel)

	args := []interface{}{opts.StartDate, opts.EndDate}
	argIdx := 3

	// Add node filter
	if len(opts.NodeIDs) > 0 {
		placeholders := make([]string, len(opts.NodeIDs))
		for i, id := range opts.NodeIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		query += fmt.Sprintf(" AND node_id IN (%s)", joinStrings(placeholders, ", "))
	}

	// Add metric filter
	if len(opts.Metrics) > 0 {
		placeholders := make([]string, len(opts.Metrics))
		for i, m := range opts.Metrics {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, m)
			argIdx++
		}
		query += fmt.Sprintf(" AND metric IN (%s)", joinStrings(placeholders, ", "))
	}

	query += fmt.Sprintf(" GROUP BY node_id, metric, unit, labels->>'%s' ORDER BY label_value, node_id, metric", opts.GroupByLabel)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary by label: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]UsageSummary)
	for rows.Next() {
		var summary UsageSummary
		var labelValue string

		err := rows.Scan(
			&summary.NodeID,
			&summary.Metric,
			&summary.Unit,
			&labelValue,
			&summary.TotalValue,
			&summary.AvgValue,
			&summary.DayCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage summary: %w", err)
		}

		result[labelValue] = append(result[labelValue], summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage summaries: %w", err)
	}

	return result, nil
}

// BulkUpsertWithLabels efficiently inserts or updates multiple usage records with labels
func (r *UsageRepository) BulkUpsertWithLabels(ctx context.Context, usages []models.NodeUsageByDimension) error {
	if len(usages) == 0 {
		return nil
	}

	query := r.QueryBuilder().
		Insert("node_usage_by_dimension").
		Columns("node_id", "usage_date", "metric", "value", "unit", "labels", "source")

	for _, usage := range usages {
		query = query.Values(usage.NodeID, usage.UsageDate, usage.Metric, usage.Value, usage.Unit, usage.Labels, usage.Source)
	}

	query = query.Suffix(`ON CONFLICT (node_id, usage_date, metric)
		DO UPDATE SET
			value = EXCLUDED.value,
			unit = EXCLUDED.unit,
			labels = EXCLUDED.labels,
			source = EXCLUDED.source,
			updated_at = now()`)

	_, err := r.ExecQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to bulk upsert usage with labels: %w", err)
	}

	return nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
