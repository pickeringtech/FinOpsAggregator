package analysis

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
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

// AnalyzeCosts provides comprehensive cost analysis for a time period
func (fa *FinOpsAnalyzer) AnalyzeCosts(ctx context.Context, startDate, endDate time.Time) (*CostSummary, error) {
	// Get all costs for the period
	costs, err := fa.store.Costs.GetByDateRange(ctx, startDate, endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs: %w", err)
	}

	// Get all nodes for name mapping
	nodes, err := fa.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	nodeMap := make(map[uuid.UUID]string)
	nodeTypeMap := make(map[uuid.UUID]string)
	for _, node := range nodes {
		nodeMap[node.ID] = node.Name
		nodeTypeMap[node.ID] = node.Type
	}

	summary := &CostSummary{
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:   startDate,
		EndDate:     endDate,
		Currency:    "USD",
		ByNode:      make(map[string]decimal.Decimal),
		ByDimension: make(map[string]decimal.Decimal),
		Trends:      make([]DailyCost, 0),
	}

	totalCost := decimal.Zero
	dailyCosts := make(map[time.Time]decimal.Decimal)

	// Aggregate costs
	for _, cost := range costs {
		nodeName := nodeMap[cost.NodeID]
		if nodeName == "" {
			nodeName = cost.NodeID.String()
		}

		totalCost = totalCost.Add(cost.Amount)
		summary.ByNode[nodeName] = summary.ByNode[nodeName].Add(cost.Amount)
		summary.ByDimension[cost.Dimension] = summary.ByDimension[cost.Dimension].Add(cost.Amount)

		// Daily trends
		date := cost.CostDate.Truncate(24 * time.Hour)
		dailyCosts[date] = dailyCosts[date].Add(cost.Amount)
	}

	summary.TotalCost = totalCost

	// Create top costs list
	for nodeName, cost := range summary.ByNode {
		nodeType := "unknown"
		for nodeID, name := range nodeMap {
			if name == nodeName {
				nodeType = nodeTypeMap[nodeID]
				break
			}
		}

		percentage := 0.0
		if !totalCost.IsZero() {
			percentage = cost.Div(totalCost).InexactFloat64() * 100
		}

		summary.TopCosts = append(summary.TopCosts, NodeCost{
			NodeName:   nodeName,
			NodeType:   nodeType,
			Cost:       cost,
			Currency:   "USD",
			Percentage: percentage,
		})
	}

	// Sort top costs by amount
	sort.Slice(summary.TopCosts, func(i, j int) bool {
		return summary.TopCosts[i].Cost.GreaterThan(summary.TopCosts[j].Cost)
	})

	// Create daily trends
	for date, cost := range dailyCosts {
		summary.Trends = append(summary.Trends, DailyCost{
			Date: date,
			Cost: cost,
		})
	}

	// Sort trends by date
	sort.Slice(summary.Trends, func(i, j int) bool {
		return summary.Trends[i].Date.Before(summary.Trends[j].Date)
	})

	return summary, nil
}

// GenerateOptimizationInsights analyzes costs and provides optimization recommendations
func (fa *FinOpsAnalyzer) GenerateOptimizationInsights(ctx context.Context, startDate, endDate time.Time) ([]CostOptimizationInsight, error) {
	costs, err := fa.store.Costs.GetByDateRange(ctx, startDate, endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs: %w", err)
	}

	nodes, err := fa.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	nodeMap := make(map[uuid.UUID]string)
	for _, node := range nodes {
		nodeMap[node.ID] = node.Name
	}

	var insights []CostOptimizationInsight

	// Analyze cost patterns
	nodeCosts := make(map[string]map[string]decimal.Decimal) // node -> dimension -> cost
	for _, cost := range costs {
		nodeName := nodeMap[cost.NodeID]
		if nodeName == "" {
			continue
		}

		if nodeCosts[nodeName] == nil {
			nodeCosts[nodeName] = make(map[string]decimal.Decimal)
		}
		nodeCosts[nodeName][cost.Dimension] = nodeCosts[nodeName][cost.Dimension].Add(cost.Amount)
	}

	// Generate insights based on cost patterns
	for nodeName, dimensions := range nodeCosts {
		totalNodeCost := decimal.Zero
		for _, cost := range dimensions {
			totalNodeCost = totalNodeCost.Add(cost)
		}

		// High cost alert
		if totalNodeCost.GreaterThan(decimal.NewFromFloat(1000)) {
			insights = append(insights, CostOptimizationInsight{
				Type:        "high_cost",
				Severity:    "high",
				Title:       "High Cost Node",
				Description: fmt.Sprintf("Node %s has high costs that may need optimization", nodeName),
				NodeName:    nodeName,
				CurrentCost: totalNodeCost,
				PotentialSavings: totalNodeCost.Mul(decimal.NewFromFloat(0.15)), // Assume 15% potential savings
				Recommendation: "Review resource utilization and consider rightsizing or reserved instances",
			})
		}

		// Unused resources (very low cost might indicate unused resources)
		if totalNodeCost.LessThan(decimal.NewFromFloat(10)) && totalNodeCost.GreaterThan(decimal.Zero) {
			insights = append(insights, CostOptimizationInsight{
				Type:        "underutilized",
				Severity:    "medium",
				Title:       "Potentially Underutilized Resource",
				Description: fmt.Sprintf("Node %s has very low costs, may be underutilized", nodeName),
				NodeName:    nodeName,
				CurrentCost: totalNodeCost,
				PotentialSavings: totalNodeCost.Mul(decimal.NewFromFloat(0.8)), // Could save 80% if unused
				Recommendation: "Review if this resource is needed or can be terminated",
			})
		}

		// Storage optimization
		if storageCost, exists := dimensions["storage_gb_month"]; exists && storageCost.GreaterThan(decimal.NewFromFloat(100)) {
			insights = append(insights, CostOptimizationInsight{
				Type:        "storage_optimization",
				Severity:    "medium",
				Title:       "Storage Optimization Opportunity",
				Description: fmt.Sprintf("Node %s has high storage costs", nodeName),
				NodeName:    nodeName,
				Dimension:   "storage_gb_month",
				CurrentCost: storageCost,
				PotentialSavings: storageCost.Mul(decimal.NewFromFloat(0.25)), // 25% savings through optimization
				Recommendation: "Consider storage tiering, compression, or lifecycle policies",
			})
		}
	}

	// Sort insights by potential savings
	sort.Slice(insights, func(i, j int) bool {
		return insights[i].PotentialSavings.GreaterThan(insights[j].PotentialSavings)
	})

	return insights, nil
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
