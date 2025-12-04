package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

// NodeWithCost represents a node with its aggregated cost
type NodeWithCost struct {
	ID       uuid.UUID       `json:"id"`
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	TotalCost decimal.Decimal `json:"total_cost"`
	Currency string          `json:"currency"`
}

// CostByType represents cost aggregated by node type
type CostByType struct {
	Type           string          `json:"type"`
	DirectCost     decimal.Decimal `json:"direct_cost"`
	IndirectCost   decimal.Decimal `json:"indirect_cost"`
	TotalCost      decimal.Decimal `json:"total_cost"`
	NodeCount      int             `json:"node_count"`
	Currency       string          `json:"currency"`
	PercentOfTotal float64         `json:"percent_of_total"`
}

// CostByDimension represents cost aggregated by a custom dimension
type CostByDimension struct {
	DimensionKey   string          `json:"dimension_key"`
	DimensionValue string          `json:"dimension_value"`
	TotalCost      decimal.Decimal `json:"total_cost"`
	NodeCount      int             `json:"node_count"`
	Currency       string          `json:"currency"`
}

// ListNodesWithCosts retrieves nodes with their aggregated costs
func (r *CostRepository) ListNodesWithCosts(ctx context.Context, startDate, endDate time.Time, currency string, nodeType string, limit int, offset int) ([]NodeWithCost, error) {
	// Build query with optional type filter
	queryBuilder := `
		WITH latest_run AS (
			SELECT id
			FROM computation_runs
			WHERE window_start <= $2 AND window_end >= $1
			  AND status = 'completed'
			ORDER BY created_at DESC
			LIMIT 1
		),
		node_costs AS (
			SELECT
				n.id,
				n.name,
				n.type,
				SUM(a.total_amount) as total_cost,
				$3 as currency
			FROM cost_nodes n
			JOIN allocation_results_by_dimension a ON n.id = a.node_id
			JOIN latest_run lr ON a.run_id = lr.id
			WHERE a.allocation_date >= $1
			  AND a.allocation_date <= $2`

	args := []interface{}{startDate, endDate, currency}
	argIndex := 4

	if nodeType != "" {
		queryBuilder += fmt.Sprintf(" AND n.type = $%d", argIndex)
		args = append(args, nodeType)
		argIndex++
	}

	queryBuilder += `
			GROUP BY n.id, n.name, n.type
		)
		SELECT id, name, type, total_cost, currency
		FROM node_costs
		ORDER BY total_cost DESC`

	if limit > 0 {
		queryBuilder += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		queryBuilder += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.Query(ctx, queryBuilder, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes with costs: %w", err)
	}
	defer rows.Close()

	var nodes []NodeWithCost
	for rows.Next() {
		var node NodeWithCost
		err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.Type,
			&node.TotalCost,
			&node.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating nodes: %w", err)
	}

	return nodes, nil
}

// GetCostsByType retrieves costs aggregated by node type
func (r *CostRepository) GetCostsByType(ctx context.Context, startDate, endDate time.Time, currency string) ([]CostByType, error) {
	// Query to aggregate costs by node type from the most recent computation run
	query := `
		WITH latest_run AS (
			SELECT id
			FROM computation_runs
			WHERE window_start <= $2 AND window_end >= $1
			  AND status = 'completed'
			ORDER BY created_at DESC
			LIMIT 1
		),
		type_costs AS (
			SELECT
				n.type,
				SUM(a.direct_amount) as direct_cost,
				SUM(a.indirect_amount) as indirect_cost,
				SUM(a.total_amount) as total_cost,
				COUNT(DISTINCT n.id) as node_count,
				$3 as currency
			FROM cost_nodes n
			JOIN allocation_results_by_dimension a ON n.id = a.node_id
			JOIN latest_run lr ON a.run_id = lr.id
			WHERE a.allocation_date >= $1
			  AND a.allocation_date <= $2
			GROUP BY n.type
		),
		grand_total AS (
			SELECT SUM(total_cost) as total FROM type_costs
		)
		SELECT
			tc.type,
			tc.direct_cost,
			tc.indirect_cost,
			tc.total_cost,
			tc.node_count,
			tc.currency,
			CASE WHEN gt.total > 0 THEN (tc.total_cost / gt.total * 100)::float8 ELSE 0 END as percent_of_total
		FROM type_costs tc, grand_total gt
		ORDER BY tc.total_cost DESC
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by type: %w", err)
	}
	defer rows.Close()

	var costsByType []CostByType
	for rows.Next() {
		var costByType CostByType
		err := rows.Scan(
			&costByType.Type,
			&costByType.DirectCost,
			&costByType.IndirectCost,
			&costByType.TotalCost,
			&costByType.NodeCount,
			&costByType.Currency,
			&costByType.PercentOfTotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost by type: %w", err)
		}
		costsByType = append(costsByType, costByType)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating costs by type: %w", err)
	}

	return costsByType, nil
}

// GetCostsByDimension retrieves costs aggregated by a custom dimension (e.g., tags, labels)
func (r *CostRepository) GetCostsByDimension(ctx context.Context, startDate, endDate time.Time, currency string, dimensionKey string) ([]CostByDimension, error) {
	// This aggregates by metadata/labels - for now, we'll aggregate by cost_labels
	query := `
		WITH latest_run AS (
			SELECT id
			FROM computation_runs
			WHERE window_start <= $2 AND window_end >= $1
			  AND status = 'completed'
			ORDER BY created_at DESC
			LIMIT 1
		),
		dimension_costs AS (
			SELECT
				$4 as dimension_key,
				COALESCE(n.cost_labels->>$4, 'untagged') as dimension_value,
				SUM(a.total_amount) as total_cost,
				COUNT(DISTINCT n.id) as node_count,
				$3 as currency
			FROM cost_nodes n
			JOIN allocation_results_by_dimension a ON n.id = a.node_id
			JOIN latest_run lr ON a.run_id = lr.id
			WHERE a.allocation_date >= $1
			  AND a.allocation_date <= $2
			GROUP BY dimension_value
		)
		SELECT dimension_key, dimension_value, total_cost, node_count, currency
		FROM dimension_costs
		ORDER BY total_cost DESC
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate, currency, dimensionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by dimension: %w", err)
	}
	defer rows.Close()

	var costsByDimension []CostByDimension
	for rows.Next() {
		var costByDim CostByDimension
		err := rows.Scan(
			&costByDim.DimensionKey,
			&costByDim.DimensionValue,
			&costByDim.TotalCost,
			&costByDim.NodeCount,
			&costByDim.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost by dimension: %w", err)
		}
		costsByDimension = append(costsByDimension, costByDim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating costs by dimension: %w", err)
	}

	return costsByDimension, nil
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



// GetTotalCostByDateRange retrieves the total of all allocated costs within a date range
// Uses allocation_results_by_dimension which contains the computed costs after allocation
func (r *CostRepository) GetTotalCostByDateRange(ctx context.Context, startDate, endDate time.Time, currency string) (decimal.Decimal, error) {
	query := `
		WITH latest_run AS (
			SELECT id
			FROM computation_runs
			WHERE window_start <= $2 AND window_end >= $1
			  AND status = 'completed'
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT COALESCE(SUM(total_amount), 0) as total
		FROM allocation_results_by_dimension a
		JOIN latest_run lr ON a.run_id = lr.id
		WHERE a.allocation_date >= $1
		  AND a.allocation_date <= $2
	`

	row := r.db.QueryRow(ctx, query, startDate, endDate)

	var total decimal.Decimal
	if err := row.Scan(&total); err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total allocated cost: %w", err)
	}

	return total, nil
}

// GetRawTotalCostByDateRange retrieves the total raw infrastructure cost within a date range
// Uses node_costs_by_dimension which contains the ingested pre-allocation spend
func (r *CostRepository) GetRawTotalCostByDateRange(ctx context.Context, startDate, endDate time.Time, currency string) (decimal.Decimal, error) {
	query := `
			SELECT COALESCE(SUM(amount), 0) as total
			FROM node_costs_by_dimension
			WHERE cost_date >= $1
			  AND cost_date <= $2
			  AND currency = $3
	`

	row := r.db.QueryRow(ctx, query, startDate, endDate, currency)

	var total decimal.Decimal
	if err := row.Scan(&total); err != nil {
		return decimal.Zero, fmt.Errorf("failed to get raw total cost: %w", err)
	}

	return total, nil
}

// GetPlatformAndSharedTotalCostByDateRange retrieves the total cost for platform and shared services within a date range
// This represents the "raw infrastructure cost" available for allocation to products
func (r *CostRepository) GetPlatformAndSharedTotalCostByDateRange(ctx context.Context, startDate, endDate time.Time, currency string) (decimal.Decimal, error) {
	query := `
			SELECT COALESCE(SUM(c.amount), 0) as total
			FROM node_costs_by_dimension c
			JOIN cost_nodes n ON c.node_id = n.id
			WHERE c.cost_date >= $1
			  AND c.cost_date <= $2
			  AND c.currency = $3
			  AND (n.is_platform = true OR n.type = 'shared')
			  AND n.archived_at IS NULL
	`

	row := r.db.QueryRow(ctx, query, startDate, endDate, currency)

	var total decimal.Decimal
	if err := row.Scan(&total); err != nil {
		return decimal.Zero, fmt.Errorf("failed to get platform and shared total cost: %w", err)
	}

	return total, nil
}

// GetTotalInfrastructureCostByDateRange retrieves the total infrastructure cost including
// all infrastructure-like nodes: Resource, Shared Service, Platform Service, and Infrastructure.
// This intentionally excludes product direct costs so that raw infrastructure reflects only
// the services that are allocated into products.
func (r *CostRepository) GetTotalInfrastructureCostByDateRange(ctx context.Context, startDate, endDate time.Time, currency string) (decimal.Decimal, error) {
	query := `
			SELECT COALESCE(SUM(c.amount), 0) as total
			FROM node_costs_by_dimension c
			JOIN cost_nodes n ON c.node_id = n.id
			WHERE c.cost_date >= $1
			  AND c.cost_date <= $2
			  AND c.currency = $3
			  AND (n.type IN ('resource', 'shared', 'platform', 'infrastructure') OR n.is_platform = true)
			  AND n.archived_at IS NULL
	`

	row := r.db.QueryRow(ctx, query, startDate, endDate, currency)

	var total decimal.Decimal
	if err := row.Scan(&total); err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total infrastructure cost: %w", err)
	}

	return total, nil
}


// GetAllocatedCostsByNodeAndDateRange retrieves allocated costs for a node from allocation results
// This uses the computed allocation results rather than raw input costs
// Returns direct_amount (node's own costs) in TotalAmount field
func (r *CostRepository) GetAllocatedCostsByNodeAndDateRange(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time) ([]models.AllocationResultByDimension, error) {
	query := `
		WITH latest_run AS (
			SELECT id
			FROM computation_runs
			WHERE window_start <= $3 AND window_end >= $2
			  AND status = 'completed'
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT a.run_id, a.node_id, a.allocation_date, a.dimension, a.direct_amount, a.indirect_amount, a.total_amount, a.created_at, a.updated_at
		FROM allocation_results_by_dimension a
		JOIN latest_run lr ON a.run_id = lr.id
		WHERE a.node_id = $1
		  AND a.allocation_date >= $2
		  AND a.allocation_date <= $3
		ORDER BY a.allocation_date, a.dimension
	`

	rows, err := r.db.Query(ctx, query, nodeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocated costs: %w", err)
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

// AllocationFromNode represents an allocation from a parent node to a child node
type AllocationFromNode struct {
	ParentID  uuid.UUID
	ChildID   uuid.UUID
	Amount    decimal.Decimal
	Dimension string
	Strategy  string
	Date      time.Time
}

// DetailedCostRecord represents a detailed cost record with node information
type DetailedCostRecord struct {
	NodeID       uuid.UUID       `json:"node_id"`
	NodeName     string          `json:"node_name"`
	NodeType     string          `json:"node_type"`
	Date         time.Time       `json:"date"`
	Dimension    string          `json:"dimension"`
	DirectCost   decimal.Decimal `json:"direct_cost"`
	IndirectCost decimal.Decimal `json:"indirect_cost"`
	TotalCost    decimal.Decimal `json:"total_cost"`
	Currency     string          `json:"currency"`
}

// RawCostRecord represents a raw ingested cost record with node information
type RawCostRecord struct {
	NodeID    uuid.UUID              `json:"node_id"`
	NodeName  string                 `json:"node_name"`
	NodeType  string                 `json:"node_type"`
	Date      time.Time              `json:"date"`
	Dimension string                 `json:"dimension"`
	Amount    decimal.Decimal        `json:"amount"`
	Currency  string                 `json:"currency"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// GetAllocationsFromNode retrieves all allocations FROM a specific node TO its children
// This shows how much cost has been allocated from this infrastructure node to products
func (r *CostRepository) GetAllocationsFromNode(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time) ([]AllocationFromNode, error) {
	query := `
		WITH latest_run AS (
			SELECT id
			FROM computation_runs
			WHERE window_start <= $2 AND window_end >= $3
			  AND status = 'completed'
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT
			c.parent_id,
			c.child_id,
			c.contributed_amount,
			c.dimension,
			COALESCE(e.default_strategy, 'unknown') as strategy,
			c.contribution_date
		FROM contribution_results_by_dimension c
		JOIN latest_run lr ON c.run_id = lr.id
		LEFT JOIN dependency_edges e ON e.parent_id = c.parent_id AND e.child_id = c.child_id
		WHERE c.parent_id = $1
		  AND c.contribution_date >= $2
		  AND c.contribution_date <= $3
		ORDER BY c.contribution_date, c.child_id, c.dimension
	`

	rows, err := r.db.Query(ctx, query, nodeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocations from node: %w", err)
	}
	defer rows.Close()

	var allocations []AllocationFromNode
	for rows.Next() {
		var alloc AllocationFromNode
		err := rows.Scan(
			&alloc.ParentID,
			&alloc.ChildID,
			&alloc.Amount,
			&alloc.Dimension,
			&alloc.Strategy,
			&alloc.Date,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allocation: %w", err)
		}
		allocations = append(allocations, alloc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating allocations: %w", err)
	}

	return allocations, nil
}

// GetDetailedCostRecords retrieves detailed cost records (not aggregated) with node information
func (r *CostRepository) GetDetailedCostRecords(ctx context.Context, startDate, endDate time.Time, currency string, nodeType string) ([]DetailedCostRecord, error) {
	// Query to get detailed allocation results with node information
	queryBuilder := `
		WITH latest_run AS (
			SELECT id
			FROM computation_runs
			WHERE window_start <= $2 AND window_end >= $1
			  AND status = 'completed'
			ORDER BY created_at DESC
			LIMIT 1
		)
		SELECT
			n.id,
			n.name,
			n.type,
			a.allocation_date,
			a.dimension,
			a.direct_amount,
			a.indirect_amount,
			a.total_amount,
			$3 as currency
		FROM cost_nodes n
		JOIN allocation_results_by_dimension a ON n.id = a.node_id
		JOIN latest_run lr ON a.run_id = lr.id
		WHERE a.allocation_date >= $1
		  AND a.allocation_date <= $2`

	args := []interface{}{startDate, endDate, currency}
	argIndex := 4

	if nodeType != "" {
		queryBuilder += fmt.Sprintf(" AND n.type = $%d", argIndex)
		args = append(args, nodeType)
		argIndex++
	}

	queryBuilder += `
		ORDER BY n.name, a.allocation_date, a.dimension`

	rows, err := r.db.Query(ctx, queryBuilder, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed cost records: %w", err)
	}
	defer rows.Close()

	var records []DetailedCostRecord
	for rows.Next() {
		var record DetailedCostRecord
		err := rows.Scan(
			&record.NodeID,
			&record.NodeName,
			&record.NodeType,
			&record.Date,
			&record.Dimension,
			&record.DirectCost,
			&record.IndirectCost,
			&record.TotalCost,
			&record.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan detailed cost record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating detailed cost records: %w", err)
	}

	return records, nil
}

// GetRawCostRecords retrieves raw ingested cost records with node information
func (r *CostRepository) GetRawCostRecords(ctx context.Context, startDate, endDate time.Time, currency string, nodeType string) ([]RawCostRecord, error) {
	// Query to get raw cost data from node_costs_by_dimension with node information
	queryBuilder := `
		SELECT
			n.id,
			n.name,
			n.type,
			c.cost_date,
			c.dimension,
			c.amount,
			c.currency,
			c.metadata
		FROM cost_nodes n
		JOIN node_costs_by_dimension c ON n.id = c.node_id
		WHERE c.cost_date >= $1
		  AND c.cost_date <= $2
		  AND c.currency = $3`

	args := []interface{}{startDate, endDate, currency}
	argIndex := 4

	if nodeType != "" {
		queryBuilder += fmt.Sprintf(" AND n.type = $%d", argIndex)
		args = append(args, nodeType)
		argIndex++
	}

	queryBuilder += `
		ORDER BY n.name, c.cost_date, c.dimension`

	rows, err := r.db.Query(ctx, queryBuilder, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw cost records: %w", err)
	}
	defer rows.Close()

	var records []RawCostRecord
	for rows.Next() {
		var record RawCostRecord
		var metadataJSON []byte
		err := rows.Scan(
			&record.NodeID,
			&record.NodeName,
			&record.NodeType,
			&record.Date,
			&record.Dimension,
			&record.Amount,
			&record.Currency,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan raw cost record: %w", err)
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &record.Metadata); err != nil {
				// If metadata parsing fails, set to empty map
				record.Metadata = make(map[string]interface{})
			}
		} else {
			record.Metadata = make(map[string]interface{})
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating raw cost records: %w", err)
	}

	return records, nil
}

// ProductHierarchyRecord represents a flattened product hierarchy record for CSV export
type ProductHierarchyRecord struct {
	ProductID              uuid.UUID       `json:"product_id"`
	ProductName            string          `json:"product_name"`
	ProductType            string          `json:"product_type"`
	Date                   time.Time       `json:"date"`
	Dimension              string          `json:"dimension"`
	DirectCost             decimal.Decimal `json:"direct_cost"`
	IndirectCost           decimal.Decimal `json:"indirect_cost"`
	TotalCost              decimal.Decimal `json:"total_cost"`
	SharedServiceCost      decimal.Decimal `json:"shared_service_cost"`
	Currency               string          `json:"currency"`
	DownstreamNodeID       *uuid.UUID      `json:"downstream_node_id,omitempty"`
	DownstreamNodeName     *string         `json:"downstream_node_name,omitempty"`
	DownstreamNodeType     *string         `json:"downstream_node_type,omitempty"`
	ContributedAmount      *decimal.Decimal `json:"contributed_amount,omitempty"`
	AllocationStrategy     *string         `json:"allocation_strategy,omitempty"`
	Metadata               map[string]interface{} `json:"metadata"`
}

// GetProductHierarchyRecords retrieves flattened product hierarchy records with downstream relationships
func (r *CostRepository) GetProductHierarchyRecords(ctx context.Context, startDate, endDate time.Time, currency string) ([]ProductHierarchyRecord, error) {
	// Simplified query to get product hierarchy with downstream relationships
	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			p.type as product_type,
			a.allocation_date,
			a.dimension,
			a.direct_amount,
			a.indirect_amount,
			a.total_amount,
			COALESCE(ssc.shared_service_amount, 0) as shared_service_cost,
			$3 as currency,
			c.child_id as downstream_node_id,
			cn.name as downstream_node_name,
			cn.type as downstream_node_type,
			c.contributed_amount,
			COALESCE(e.default_strategy, 'unknown') as allocation_strategy,
			p.metadata as product_metadata
		FROM cost_nodes p
		JOIN allocation_results_by_dimension a ON p.id = a.node_id
		LEFT JOIN contribution_results_by_dimension c ON (
			p.id = c.parent_id
			AND a.allocation_date = c.contribution_date
			AND a.dimension = c.dimension
			AND a.run_id = c.run_id
		)
		LEFT JOIN cost_nodes cn ON c.child_id = cn.id
		LEFT JOIN dependency_edges e ON e.parent_id = c.parent_id AND e.child_id = c.child_id
		LEFT JOIN (
			SELECT
				c2.parent_id as product_id,
				c2.contribution_date,
				c2.dimension,
				SUM(c2.contributed_amount) as shared_service_amount
			FROM contribution_results_by_dimension c2
			JOIN cost_nodes cn2 ON c2.child_id = cn2.id
			WHERE cn2.type IN ('shared', 'platform')
			  AND c2.contribution_date >= $1
			  AND c2.contribution_date <= $2
			GROUP BY c2.parent_id, c2.contribution_date, c2.dimension
		) ssc ON (
			p.id = ssc.product_id
			AND a.allocation_date = ssc.contribution_date
			AND a.dimension = ssc.dimension
		)
		WHERE p.type = 'product'
		  AND a.allocation_date >= $1
		  AND a.allocation_date <= $2
		ORDER BY p.name, a.allocation_date, a.dimension, cn.name
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get product hierarchy records: %w", err)
	}
	defer rows.Close()

	var records []ProductHierarchyRecord
	for rows.Next() {
		var record ProductHierarchyRecord
		var metadataJSON []byte
		var downstreamNodeID, downstreamNodeName, downstreamNodeType, contributedAmount, allocationStrategy interface{}

		err := rows.Scan(
			&record.ProductID,
			&record.ProductName,
			&record.ProductType,
			&record.Date,
			&record.Dimension,
			&record.DirectCost,
			&record.IndirectCost,
			&record.TotalCost,
			&record.SharedServiceCost,
			&record.Currency,
			&downstreamNodeID,
			&downstreamNodeName,
			&downstreamNodeType,
			&contributedAmount,
			&allocationStrategy,
			&metadataJSON,
		)


		if err != nil {
			return nil, fmt.Errorf("failed to scan product hierarchy record: %w", err)
		}

		// Handle nullable downstream node fields
		if downstreamNodeID != nil {
			// Handle different types for downstream node ID
			switch v := downstreamNodeID.(type) {
			case uuid.UUID:
				record.DownstreamNodeID = &v
			case string:
				if parsed, err := uuid.Parse(v); err == nil {
					record.DownstreamNodeID = &parsed
				}
			case []byte:
				if len(v) == 16 {
					if parsed, err := uuid.FromBytes(v); err == nil {
						record.DownstreamNodeID = &parsed
					}
				}
			case [16]uint8:
				// Handle PostgreSQL UUID as [16]uint8
				byteSlice := v[:]
				if parsed, err := uuid.FromBytes(byteSlice); err == nil {
					record.DownstreamNodeID = &parsed
				}
			}
		}
		if downstreamNodeName != nil {
			if name, ok := downstreamNodeName.(string); ok {
				record.DownstreamNodeName = &name
			}
		}
		if downstreamNodeType != nil {
			if nodeType, ok := downstreamNodeType.(string); ok {
				record.DownstreamNodeType = &nodeType
			}
		}
		if contributedAmount != nil {
			// Handle different types for contributed amount
			switch v := contributedAmount.(type) {
			case decimal.Decimal:
				record.ContributedAmount = &v
			case pgtype.Numeric:
				// Convert pgtype.Numeric to decimal.Decimal
				if v.Valid {
					// Convert to string first, then to decimal
					str := v.Int.String()
					if v.Exp < 0 {
						// Handle decimal places
						exp := int(-v.Exp)
						if len(str) <= exp {
							// Pad with zeros if needed
							str = "0." + fmt.Sprintf("%0*s", exp, str)
						} else {
							// Insert decimal point
							pos := len(str) - exp
							str = str[:pos] + "." + str[pos:]
						}
					}
					if parsed, err := decimal.NewFromString(str); err == nil {
						record.ContributedAmount = &parsed
					}
				}
			case string:
				if parsed, err := decimal.NewFromString(v); err == nil {
					record.ContributedAmount = &parsed
				}
			case float64:
				parsed := decimal.NewFromFloat(v)
				record.ContributedAmount = &parsed
			case int64:
				parsed := decimal.NewFromInt(v)
				record.ContributedAmount = &parsed
			}
		}
		if allocationStrategy != nil {
			if strategy, ok := allocationStrategy.(string); ok {
				record.AllocationStrategy = &strategy
			}
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &record.Metadata); err != nil {
				// If metadata parsing fails, set to empty map
				record.Metadata = make(map[string]interface{})
			}
		} else {
			record.Metadata = make(map[string]interface{})
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product hierarchy records: %w", err)
	}

	return records, nil
}
