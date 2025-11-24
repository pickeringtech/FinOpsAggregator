package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/shopspring/decimal"
)

// RecommendationAnalyzer analyzes usage metrics and generates cost optimization recommendations
type RecommendationAnalyzer struct {
	store *store.Store
}

// NewRecommendationAnalyzer creates a new recommendation analyzer
func NewRecommendationAnalyzer(store *store.Store) *RecommendationAnalyzer {
	return &RecommendationAnalyzer{
		store: store,
	}
}

// AnalyzeNode analyzes a single node and generates recommendations
func (a *RecommendationAnalyzer) AnalyzeNode(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time) ([]models.CostRecommendation, error) {
	// Get node details
	node, err := a.store.Nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get usage metrics for the node
	usageData, err := a.store.Usage.GetByNodeAndDateRange(ctx, nodeID, startDate, endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage data: %w", err)
	}

	// Get cost data for the node
	costData, err := a.store.Costs.GetByNodeAndDateRange(ctx, nodeID, startDate, endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost data: %w", err)
	}

	// Calculate total cost
	totalCost := decimal.Zero
	for _, cost := range costData {
		totalCost = totalCost.Add(cost.Amount)
	}

	// Analyze usage patterns and generate recommendations
	recommendations := []models.CostRecommendation{}

	// Group usage by metric
	metricUsage := make(map[string][]models.NodeUsageByDimension)
	for _, usage := range usageData {
		metricUsage[usage.Metric] = append(metricUsage[usage.Metric], usage)
	}

	// Analyze each metric
	for metric, usages := range metricUsage {
		if rec := a.analyzeMetric(node, metric, usages, totalCost, startDate, endDate); rec != nil {
			recommendations = append(recommendations, *rec)
		}
	}

	return recommendations, nil
}

// analyzeMetric analyzes a specific metric and generates a recommendation if needed
func (a *RecommendationAnalyzer) analyzeMetric(
	node *models.CostNode,
	metric string,
	usages []models.NodeUsageByDimension,
	totalCost decimal.Decimal,
	startDate, endDate time.Time,
) *models.CostRecommendation {
	if len(usages) == 0 {
		return nil
	}

	// Calculate statistics
	var sum, peak, min decimal.Decimal
	peak = decimal.Zero
	min = usages[0].Value

	for _, usage := range usages {
		sum = sum.Add(usage.Value)
		if usage.Value.GreaterThan(peak) {
			peak = usage.Value
		}
		if usage.Value.LessThan(min) {
			min = usage.Value
		}
	}

	average := sum.Div(decimal.NewFromInt(int64(len(usages))))
	current := usages[len(usages)-1].Value // Most recent value

	// Calculate utilization percentage (peak vs current)
	var utilization decimal.Decimal
	if !peak.IsZero() {
		utilization = average.Div(peak).Mul(decimal.NewFromInt(100))
	}

	// Determine if recommendation is needed
	// Low utilization: average is less than 50% of peak
	lowUtilizationThreshold := decimal.NewFromFloat(50.0)
	veryLowUtilizationThreshold := decimal.NewFromFloat(30.0)

	if utilization.LessThan(veryLowUtilizationThreshold) && !peak.IsZero() {
		// High severity: very low utilization
		potentialSavings := totalCost.Mul(decimal.NewFromFloat(0.4)) // Estimate 40% savings

		return &models.CostRecommendation{
			ID:                 uuid.New(),
			NodeID:             node.ID,
			NodeName:           node.Name,
			NodeType:           node.Type,
			Type:               models.RecommendationTypeDownsize,
			Severity:           models.RecommendationSeverityHigh,
			Title:              fmt.Sprintf("Downsize %s - Very Low Utilization", node.Name),
			Description:        fmt.Sprintf("The %s metric shows very low utilization (%.1f%%). Peak usage is %.2f but average is only %.2f. Consider downsizing to reduce costs.", metric, utilization, peak, average),
			CurrentCost:        totalCost,
			PotentialSavings:   potentialSavings,
			Currency:           "USD",
			Metric:             metric,
			CurrentValue:       current,
			PeakValue:          peak,
			AverageValue:       average,
			UtilizationPercent: utilization,
			RecommendedAction:  fmt.Sprintf("Reduce capacity to match average usage of %.2f %s", average, usages[0].Unit),
			AnalysisPeriod:     fmt.Sprintf("%d days", int(endDate.Sub(startDate).Hours()/24)),
			StartDate:          startDate,
			EndDate:            endDate,
			CreatedAt:          time.Now(),
		}
	} else if utilization.LessThan(lowUtilizationThreshold) && !peak.IsZero() {
		// Medium severity: low utilization
		potentialSavings := totalCost.Mul(decimal.NewFromFloat(0.2)) // Estimate 20% savings

		return &models.CostRecommendation{
			ID:                 uuid.New(),
			NodeID:             node.ID,
			NodeName:           node.Name,
			NodeType:           node.Type,
			Type:               models.RecommendationTypeRightsize,
			Severity:           models.RecommendationSeverityMedium,
			Title:              fmt.Sprintf("Rightsize %s - Low Utilization", node.Name),
			Description:        fmt.Sprintf("The %s metric shows low utilization (%.1f%%). Peak usage is %.2f but average is only %.2f. Consider rightsizing to optimize costs.", metric, utilization, peak, average),
			CurrentCost:        totalCost,
			PotentialSavings:   potentialSavings,
			Currency:           "USD",
			Metric:             metric,
			CurrentValue:       current,
			PeakValue:          peak,
			AverageValue:       average,
			UtilizationPercent: utilization,
			RecommendedAction:  fmt.Sprintf("Adjust capacity to better match average usage of %.2f %s", average, usages[0].Unit),
			AnalysisPeriod:     fmt.Sprintf("%d days", int(endDate.Sub(startDate).Hours()/24)),
			StartDate:          startDate,
			EndDate:            endDate,
			CreatedAt:          time.Now(),
		}
	}

	return nil
}

// AnalyzeAllNodes analyzes all nodes and generates recommendations
func (a *RecommendationAnalyzer) AnalyzeAllNodes(ctx context.Context, startDate, endDate time.Time) ([]models.CostRecommendation, error) {
	// Get all nodes
	nodes, err := a.store.Nodes.List(ctx, store.NodeFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var allRecommendations []models.CostRecommendation

	for _, node := range nodes {
		recommendations, err := a.AnalyzeNode(ctx, node.ID, startDate, endDate)
		if err != nil {
			// Log error but continue with other nodes
			continue
		}
		allRecommendations = append(allRecommendations, recommendations...)
	}

	return allRecommendations, nil
}

// GetRecommendationsByNode gets recommendations for a specific node
func (a *RecommendationAnalyzer) GetRecommendationsByNode(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time) ([]models.CostRecommendation, error) {
	return a.AnalyzeNode(ctx, nodeID, startDate, endDate)
}

// GetRecommendationsBySeverity filters recommendations by severity
func (a *RecommendationAnalyzer) GetRecommendationsBySeverity(recommendations []models.CostRecommendation, severity models.RecommendationSeverity) []models.CostRecommendation {
	var filtered []models.CostRecommendation
	for _, rec := range recommendations {
		if rec.Severity == severity {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

// CalculateTotalPotentialSavings calculates total potential savings from all recommendations
func (a *RecommendationAnalyzer) CalculateTotalPotentialSavings(recommendations []models.CostRecommendation) decimal.Decimal {
	total := decimal.Zero
	for _, rec := range recommendations {
		total = total.Add(rec.PotentialSavings)
	}
	return total
}

