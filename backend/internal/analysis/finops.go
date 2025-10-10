package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/shopspring/decimal"
)

// FinOpsAnalyzer provides comprehensive FinOps analysis and insights
type FinOpsAnalyzer struct {
	store *store.Store
}

// NewFinOpsAnalyzer creates a new FinOps analyzer
func NewFinOpsAnalyzer(store *store.Store) *FinOpsAnalyzer {
	return &FinOpsAnalyzer{
		store: store,
	}
}

// CostSummary represents cost analysis for a time period
type CostSummary struct {
	Period      string                    `json:"period"`
	StartDate   time.Time                 `json:"start_date"`
	EndDate     time.Time                 `json:"end_date"`
	TotalCost   decimal.Decimal           `json:"total_cost"`
	Currency    string                    `json:"currency"`
	ByNode      map[string]decimal.Decimal `json:"by_node"`
	ByDimension map[string]decimal.Decimal `json:"by_dimension"`
	TopCosts    []NodeCost                `json:"top_costs"`
	Trends      []DailyCost               `json:"trends"`
}

// NodeCost represents cost for a specific node
type NodeCost struct {
	NodeName  string          `json:"node_name"`
	NodeType  string          `json:"node_type"`
	Cost      decimal.Decimal `json:"cost"`
	Currency  string          `json:"currency"`
	Percentage float64        `json:"percentage"`
}

// DailyCost represents daily cost trend
type DailyCost struct {
	Date time.Time       `json:"date"`
	Cost decimal.Decimal `json:"cost"`
}

// CostOptimizationInsight represents a cost optimization recommendation
type CostOptimizationInsight struct {
	Type        string          `json:"type"`
	Severity    string          `json:"severity"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	NodeName    string          `json:"node_name"`
	Dimension   string          `json:"dimension"`
	CurrentCost decimal.Decimal `json:"current_cost"`
	PotentialSavings decimal.Decimal `json:"potential_savings"`
	Recommendation string     `json:"recommendation"`
}

// AllocationEfficiency represents how efficiently costs are being allocated
type AllocationEfficiency struct {
	NodeName           string  `json:"node_name"`
	DirectCostRatio    float64 `json:"direct_cost_ratio"`
	IndirectCostRatio  float64 `json:"indirect_cost_ratio"`
	AllocationAccuracy float64 `json:"allocation_accuracy"`
	EfficiencyScore    float64 `json:"efficiency_score"`
}

// AnalyzeCosts provides comprehensive cost analysis for a time period using optimized database queries
func (fa *FinOpsAnalyzer) AnalyzeCosts(ctx context.Context, startDate, endDate time.Time) (*CostSummary, error) {
	// Use optimized database query instead of loading all data into memory
	overview, err := fa.store.Costs.GetCostOverviewByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost overview: %w", err)
	}

	// Convert database result to analysis format
	summary := &CostSummary{
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:   startDate,
		EndDate:     endDate,
		Currency:    "USD",
		TotalCost:   overview.TotalCost,
		ByNode:      make(map[string]decimal.Decimal),
		ByDimension: make(map[string]decimal.Decimal),
		TopCosts:    []NodeCost{},
		Trends:      []DailyCost{},
	}

	// Convert top nodes
	for _, node := range overview.TopNodes {
		percentage := 0.0
		if !overview.TotalCost.IsZero() {
			percentage = node.Cost.Div(overview.TotalCost).InexactFloat64() * 100
		}

		summary.TopCosts = append(summary.TopCosts, NodeCost{
			NodeName:   node.NodeName,
			NodeType:   node.NodeType,
			Cost:       node.Cost,
			Currency:   "USD",
			Percentage: percentage,
		})
		summary.ByNode[node.NodeName] = node.Cost
	}

	// Convert dimensions
	for _, dim := range overview.Dimensions {
		summary.ByDimension[dim.Dimension] = dim.Cost
	}

	// Convert daily trend
	for _, daily := range overview.DailyTrend {
		summary.Trends = append(summary.Trends, DailyCost{
			Date: daily.Date,
			Cost: daily.Cost,
		})
	}

	return summary, nil
}

// GenerateOptimizationInsights analyzes costs and provides optimization recommendations using optimized database queries
func (fa *FinOpsAnalyzer) GenerateOptimizationInsights(ctx context.Context, startDate, endDate time.Time) ([]CostOptimizationInsight, error) {
	// Use optimized database query instead of loading all data into memory
	dbInsights, err := fa.store.Costs.GetOptimizationInsights(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization insights: %w", err)
	}

	// Convert database results to analysis format
	var insights []CostOptimizationInsight
	for _, dbInsight := range dbInsights {
		insights = append(insights, CostOptimizationInsight{
			Type:             dbInsight.Type,
			Severity:         dbInsight.Severity,
			Title:            dbInsight.Title,
			Description:      dbInsight.Description,
			NodeName:         dbInsight.NodeName,
			CurrentCost:      dbInsight.CurrentCost,
			PotentialSavings: dbInsight.PotentialSavings,
			Recommendation:   getRecommendationForType(dbInsight.Type),
		})
	}

	return insights, nil
}

// getRecommendationForType returns appropriate recommendations for each insight type
func getRecommendationForType(insightType string) string {
	switch insightType {
	case "high_cost":
		return "Review resource utilization and consider rightsizing or reserved instances"
	case "compute_optimization":
		return "Consider rightsizing instances or using spot instances"
	case "storage_optimization":
		return "Consider storage tiering, compression, or lifecycle policies"
	default:
		return "Review resource configuration and usage patterns"
	}
}

// AnalyzeAllocationEfficiency analyzes how efficiently costs are being allocated
func (fa *FinOpsAnalyzer) AnalyzeAllocationEfficiency(ctx context.Context, startDate, endDate time.Time) ([]AllocationEfficiency, error) {
	// This would analyze allocation runs to determine efficiency
	// For now, return mock data to demonstrate the concept
	
	nodes, err := fa.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var efficiency []AllocationEfficiency
	for _, node := range nodes {
		// Mock efficiency calculation - in reality this would analyze allocation patterns
		efficiency = append(efficiency, AllocationEfficiency{
			NodeName:           node.Name,
			DirectCostRatio:    0.7,  // 70% direct costs
			IndirectCostRatio:  0.3,  // 30% allocated costs
			AllocationAccuracy: 0.85, // 85% allocation accuracy
			EfficiencyScore:    0.82, // Overall efficiency score
		})
	}

	return efficiency, nil
}

// AnalyzeProductCosts provides cost analysis grouped by Products and their dependencies
func (fa *FinOpsAnalyzer) AnalyzeProductCosts(ctx context.Context, startDate, endDate time.Time) (*ProductCostSummary, error) {
	// Use optimized database query to get product-level cost rollups
	overview, err := fa.store.Costs.GetProductCostOverview(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get product cost overview: %w", err)
	}

	// Convert database result to analysis format
	summary := &ProductCostSummary{
		Period:       fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:    startDate,
		EndDate:      endDate,
		Currency:     "USD",
		TotalCost:    overview.TotalCost,
		ProductCount: overview.ProductCount,
		Products:     []ProductCostDetail{},
	}

	// Convert products
	for _, product := range overview.Products {
		percentage := 0.0
		if !overview.TotalCost.IsZero() {
			percentage = product.TotalCost.Div(overview.TotalCost).InexactFloat64() * 100
		}

		// Convert cost breakdown
		var breakdown []NodeCostDetail
		for _, node := range product.CostBreakdown {
			nodePercentage := 0.0
			if !product.TotalCost.IsZero() {
				nodePercentage = node.Cost.Div(product.TotalCost).InexactFloat64() * 100
			}

			breakdown = append(breakdown, NodeCostDetail{
				NodeName:   node.NodeName,
				NodeType:   node.NodeType,
				Cost:       node.Cost,
				Currency:   "USD",
				Percentage: nodePercentage,
			})
		}

		summary.Products = append(summary.Products, ProductCostDetail{
			ProductName:        product.ProductName,
			TotalCost:          product.TotalCost,
			Currency:           "USD",
			Percentage:         percentage,
			DependentNodeCount: product.DependentNodeCount,
			CostBreakdown:      breakdown,
		})
	}

	return summary, nil
}

// ProductCostSummary represents cost analysis grouped by Products
type ProductCostSummary struct {
	Period       string                `json:"period"`
	StartDate    time.Time             `json:"start_date"`
	EndDate      time.Time             `json:"end_date"`
	TotalCost    decimal.Decimal       `json:"total_cost"`
	Currency     string                `json:"currency"`
	ProductCount int                   `json:"product_count"`
	Products     []ProductCostDetail   `json:"products"`
}

// ProductCostDetail represents detailed cost information for a product
type ProductCostDetail struct {
	ProductName        string            `json:"product_name"`
	TotalCost          decimal.Decimal   `json:"total_cost"`
	Currency           string            `json:"currency"`
	Percentage         float64           `json:"percentage"`
	DependentNodeCount int               `json:"dependent_node_count"`
	CostBreakdown      []NodeCostDetail  `json:"cost_breakdown"`
}

// NodeCostDetail represents cost detail for a node within a product
type NodeCostDetail struct {
	NodeName   string          `json:"node_name"`
	NodeType   string          `json:"node_type"`
	Cost       decimal.Decimal `json:"cost"`
	Currency   string          `json:"currency"`
	Percentage float64         `json:"percentage"`
}
