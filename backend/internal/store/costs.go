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
	"github.com/shopspring/decimal"
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

	query = query.OrderBy("total_amount DESC")

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

	return summaries, nil
}

// GetCostOverviewByDateRange retrieves aggregated cost overview data optimized for TUI
func (r *CostRepository) GetCostOverviewByDateRange(ctx context.Context, startDate, endDate time.Time) (*CostOverview, error) {
	// Single query to get all overview data with JOINs
	query := `
		WITH node_totals AS (
			SELECT
				c.node_id,
				n.name as node_name,
				n.type as node_type,
				SUM(c.amount) as total_cost
			FROM node_costs_by_dimension c
			JOIN cost_nodes n ON c.node_id = n.id
			WHERE c.cost_date >= $1 AND c.cost_date <= $2
			GROUP BY c.node_id, n.name, n.type
		),
		dimension_totals AS (
			SELECT
				-- Extract base dimension name by removing service instance suffixes
				CASE
					WHEN dimension ~ '_s[0-9]+_r[0-9]+$' THEN
						regexp_replace(dimension, '_s[0-9]+_r[0-9]+$', '')
					ELSE dimension
				END as base_dimension,
				SUM(amount) as total_cost
			FROM node_costs_by_dimension
			WHERE cost_date >= $1 AND cost_date <= $2
			GROUP BY CASE
				WHEN dimension ~ '_s[0-9]+_r[0-9]+$' THEN
					regexp_replace(dimension, '_s[0-9]+_r[0-9]+$', '')
				ELSE dimension
			END
		),
		daily_totals AS (
			SELECT
				cost_date::timestamp as cost_date,
				SUM(amount) as daily_cost
			FROM node_costs_by_dimension
			WHERE cost_date >= $1 AND cost_date <= $2
			GROUP BY cost_date
			ORDER BY cost_date
		)
		SELECT
			(SELECT SUM(total_cost) FROM node_totals) as total_cost,
			(SELECT COUNT(*) FROM node_totals) as node_count,
			(SELECT COUNT(DISTINCT base_dimension) FROM dimension_totals) as dimension_count,
			(SELECT json_agg(json_build_object('node_id', node_id, 'node_name', node_name, 'node_type', node_type, 'cost', total_cost) ORDER BY total_cost DESC) FROM node_totals) as top_nodes,
			(SELECT json_agg(json_build_object('dimension', base_dimension, 'cost', total_cost) ORDER BY total_cost DESC) FROM dimension_totals) as dimensions,
			(SELECT json_agg(json_build_object('date', cost_date, 'cost', daily_cost) ORDER BY cost_date) FROM daily_totals) as daily_trend
	`

	row := r.db.QueryRow(ctx, query, startDate, endDate)

	var overview CostOverview
	var topNodesJSON, dimensionsJSON, dailyTrendJSON []byte

	err := row.Scan(
		&overview.TotalCost,
		&overview.NodeCount,
		&overview.DimensionCount,
		&topNodesJSON,
		&dimensionsJSON,
		&dailyTrendJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost overview: %w", err)
	}

	// Parse JSON results
	if err := json.Unmarshal(topNodesJSON, &overview.TopNodes); err != nil {
		return nil, fmt.Errorf("failed to parse top nodes: %w", err)
	}
	if err := json.Unmarshal(dimensionsJSON, &overview.Dimensions); err != nil {
		return nil, fmt.Errorf("failed to parse dimensions: %w", err)
	}
	if err := json.Unmarshal(dailyTrendJSON, &overview.DailyTrend); err != nil {
		return nil, fmt.Errorf("failed to parse daily trend: %w", err)
	}

	return &overview, nil
}

// GetOptimizationInsights generates optimization insights using database queries
func (r *CostRepository) GetOptimizationInsights(ctx context.Context, startDate, endDate time.Time) ([]OptimizationInsight, error) {
	query := `
		WITH node_costs AS (
			SELECT
				c.node_id,
				n.name as node_name,
				-- Extract base dimension name by removing service instance suffixes
				CASE
					WHEN c.dimension ~ '_s[0-9]+_r[0-9]+$' THEN
						regexp_replace(c.dimension, '_s[0-9]+_r[0-9]+$', '')
					ELSE c.dimension
				END as base_dimension,
				SUM(c.amount) as total_cost,
				AVG(c.amount) as avg_daily_cost,
				COUNT(*) as day_count
			FROM node_costs_by_dimension c
			JOIN cost_nodes n ON c.node_id = n.id
			WHERE c.cost_date >= $1 AND c.cost_date <= $2
			GROUP BY c.node_id, n.name, CASE
				WHEN c.dimension ~ '_s[0-9]+_r[0-9]+$' THEN
					regexp_replace(c.dimension, '_s[0-9]+_r[0-9]+$', '')
				ELSE c.dimension
			END
		),
		high_cost_nodes AS (
			SELECT
				node_id,
				node_name,
				SUM(total_cost) as node_total,
				'high_cost' as insight_type,
				'high' as severity,
				'High Cost Node' as title,
				'Node has high costs that may need optimization' as description,
				SUM(total_cost) * 0.15 as potential_savings
			FROM node_costs
			GROUP BY node_id, node_name
			HAVING SUM(total_cost) > 1000
		),
		storage_optimization AS (
			SELECT
				node_id,
				node_name,
				total_cost as node_total,
				'storage_optimization' as insight_type,
				'medium' as severity,
				'Storage Optimization Opportunity' as title,
				'Node has high storage costs' as description,
				total_cost * 0.25 as potential_savings
			FROM node_costs
			WHERE base_dimension = 'storage_gb_month' AND total_cost > 100
		),
		compute_optimization AS (
			SELECT
				node_id,
				node_name,
				total_cost as node_total,
				'compute_optimization' as insight_type,
				'medium' as severity,
				'Compute Optimization Opportunity' as title,
				'Node has high compute costs' as description,
				total_cost * 0.20 as potential_savings
			FROM node_costs
			WHERE base_dimension = 'instance_hours' AND total_cost > 500
		)
		SELECT * FROM high_cost_nodes
		UNION ALL
		SELECT * FROM storage_optimization
		UNION ALL
		SELECT * FROM compute_optimization
		ORDER BY potential_savings DESC
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization insights: %w", err)
	}
	defer rows.Close()

	var insights []OptimizationInsight
	for rows.Next() {
		var insight OptimizationInsight
		err := rows.Scan(
			&insight.NodeID,
			&insight.NodeName,
			&insight.CurrentCost,
			&insight.Type,
			&insight.Severity,
			&insight.Title,
			&insight.Description,
			&insight.PotentialSavings,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan optimization insight: %w", err)
		}
		insights = append(insights, insight)
	}

	return insights, nil
}

// GetProductCostOverview retrieves cost overview grouped by Products and their dependencies
func (r *CostRepository) GetProductCostOverview(ctx context.Context, startDate, endDate time.Time) (*ProductCostOverview, error) {
	query := `
		WITH RECURSIVE product_costs AS (
			-- Get direct costs for all nodes
			SELECT
				n.id as node_id,
				n.name as node_name,
				n.type as node_type,
				SUM(c.amount) as direct_cost,
				CASE WHEN n.type = 'product' THEN n.id ELSE NULL END as product_id,
				CASE WHEN n.type = 'product' THEN n.name ELSE NULL END as product_name
			FROM cost_nodes n
			LEFT JOIN node_costs_by_dimension c ON n.id = c.node_id
				AND c.cost_date >= $1 AND c.cost_date <= $2
			GROUP BY n.id, n.name, n.type
		),
		product_dependencies AS (
			-- Get all nodes that belong to each product through dependency edges
			SELECT DISTINCT
				p.id as product_id,
				p.name as product_name,
				n.id as dependent_node_id,
				n.name as dependent_node_name,
				n.type as dependent_node_type,
				pc.direct_cost
			FROM cost_nodes p
			JOIN dependency_edges e ON p.id = e.parent_id
			JOIN cost_nodes n ON e.child_id = n.id
			JOIN product_costs pc ON n.id = pc.node_id
			WHERE p.type = 'product'

			UNION

			-- Include the products themselves
			SELECT
				p.id as product_id,
				p.name as product_name,
				p.id as dependent_node_id,
				p.name as dependent_node_name,
				p.type as dependent_node_type,
				pc.direct_cost
			FROM cost_nodes p
			JOIN product_costs pc ON p.id = pc.node_id
			WHERE p.type = 'product'
		),
		product_totals AS (
			SELECT
				product_id,
				product_name,
				SUM(direct_cost) as total_cost,
				COUNT(DISTINCT dependent_node_id) as dependent_node_count,
				json_agg(
					json_build_object(
						'node_name', dependent_node_name,
						'node_type', dependent_node_type,
						'cost', direct_cost
					) ORDER BY direct_cost DESC
				) as cost_breakdown
			FROM product_dependencies
			GROUP BY product_id, product_name
		)
		SELECT
			(SELECT SUM(total_cost) FROM product_totals) as total_cost,
			(SELECT COUNT(*) FROM product_totals) as product_count,
			(SELECT json_agg(
				json_build_object(
					'product_id', product_id,
					'product_name', product_name,
					'total_cost', total_cost,
					'dependent_node_count', dependent_node_count,
					'cost_breakdown', cost_breakdown
				) ORDER BY total_cost DESC
			) FROM product_totals) as products
	`

	row := r.db.QueryRow(ctx, query, startDate, endDate)

	var overview ProductCostOverview
	var productsJSON []byte

	err := row.Scan(
		&overview.TotalCost,
		&overview.ProductCount,
		&productsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product cost overview: %w", err)
	}

	// Parse JSON results
	if err := json.Unmarshal(productsJSON, &overview.Products); err != nil {
		return nil, fmt.Errorf("failed to parse products: %w", err)
	}

	return &overview, nil
}

// ProductCostOverview represents cost overview grouped by Products
type ProductCostOverview struct {
	TotalCost    decimal.Decimal `json:"total_cost"`
	ProductCount int             `json:"product_count"`
	Products     []ProductCost   `json:"products"`
}

// ProductCost represents a product's total cost including dependencies
type ProductCost struct {
	ProductID           uuid.UUID             `json:"product_id"`
	ProductName         string                `json:"product_name"`
	TotalCost           decimal.Decimal       `json:"total_cost"`
	DependentNodeCount  int                   `json:"dependent_node_count"`
	CostBreakdown       []ProductNodeCost     `json:"cost_breakdown"`
}

// ProductNodeCost represents cost breakdown by node within a product
type ProductNodeCost struct {
	NodeName string          `json:"node_name"`
	NodeType string          `json:"node_type"`
	Cost     decimal.Decimal `json:"cost"`
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
	NodeID      uuid.UUID       `db:"node_id"`
	Dimension   string          `db:"dimension"`
	Currency    string          `db:"currency"`
	TotalAmount decimal.Decimal `db:"total_amount"`
	DayCount    int             `db:"day_count"`
}

// CostOverview represents optimized cost overview data for TUI
type CostOverview struct {
	TotalCost      decimal.Decimal `json:"total_cost"`
	NodeCount      int             `json:"node_count"`
	DimensionCount int             `json:"dimension_count"`
	TopNodes       []TopNodeCost   `json:"top_nodes"`
	Dimensions     []DimensionCost `json:"dimensions"`
	DailyTrend     []DailyCost     `json:"daily_trend"`
}

// TopNodeCost represents a node's cost summary
type TopNodeCost struct {
	NodeID   uuid.UUID       `json:"node_id"`
	NodeName string          `json:"node_name"`
	NodeType string          `json:"node_type"`
	Cost     decimal.Decimal `json:"cost"`
}

// DimensionCost represents cost by dimension
type DimensionCost struct {
	Dimension string          `json:"dimension"`
	Cost      decimal.Decimal `json:"cost"`
}

// DailyCost represents daily cost trend
type DailyCost struct {
	Date time.Time       `json:"date"`
	Cost decimal.Decimal `json:"cost"`
}

// UnmarshalJSON handles custom date parsing for DailyCost
func (dc *DailyCost) UnmarshalJSON(data []byte) error {
	type Alias DailyCost
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(dc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try multiple date formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var parseErr error
	for _, format := range formats {
		if dc.Date, parseErr = time.Parse(format, aux.Date); parseErr == nil {
			break
		}
	}

	return parseErr
}

// OptimizationInsight represents database-generated optimization insights
type OptimizationInsight struct {
	NodeID           uuid.UUID       `json:"node_id"`
	NodeName         string          `json:"node_name"`
	Type             string          `json:"type"`
	Severity         string          `json:"severity"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	CurrentCost      decimal.Decimal `json:"current_cost"`
	PotentialSavings decimal.Decimal `json:"potential_savings"`
}
