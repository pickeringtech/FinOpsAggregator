package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/analysis"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
)

// ReportGenerator generates comprehensive FinOps reports
type ReportGenerator struct {
	store    *store.Store
	analyzer *analysis.FinOpsAnalyzer
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(store *store.Store) *ReportGenerator {
	return &ReportGenerator{
		store:    store,
		analyzer: analysis.NewFinOpsAnalyzer(store),
	}
}

// FinOpsReport represents a comprehensive FinOps report
type FinOpsReport struct {
	GeneratedAt        time.Time                           `json:"generated_at"`
	Period             string                              `json:"period"`
	Summary            *analysis.CostSummary               `json:"summary"`
	Insights           []analysis.CostOptimizationInsight  `json:"insights"`
	Efficiency         []analysis.AllocationEfficiency     `json:"efficiency"`
	Recommendations    []string                            `json:"recommendations"`
	ExecutiveSummary   string                              `json:"executive_summary"`
}

// GenerateReport creates a comprehensive FinOps report
func (rg *ReportGenerator) GenerateReport(ctx context.Context, startDate, endDate time.Time) (*FinOpsReport, error) {
	// Analyze costs
	summary, err := rg.analyzer.AnalyzeCosts(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze costs: %w", err)
	}

	// Generate optimization insights
	insights, err := rg.analyzer.GenerateOptimizationInsights(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate insights: %w", err)
	}

	// Analyze allocation efficiency
	efficiency, err := rg.analyzer.AnalyzeAllocationEfficiency(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze efficiency: %w", err)
	}

	// Generate recommendations
	recommendations := rg.generateRecommendations(summary, insights)

	// Generate executive summary
	execSummary := rg.generateExecutiveSummary(summary, insights)

	report := &FinOpsReport{
		GeneratedAt:      time.Now(),
		Period:           summary.Period,
		Summary:          summary,
		Insights:         insights,
		Efficiency:       efficiency,
		Recommendations:  recommendations,
		ExecutiveSummary: execSummary,
	}

	return report, nil
}

// ExportReportJSON exports the report as JSON
func (rg *ReportGenerator) ExportReportJSON(report *FinOpsReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// ExportReportHTML exports the report as HTML
func (rg *ReportGenerator) ExportReportHTML(report *FinOpsReport, filename string) error {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>FinOps Cost Attribution Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .metric { display: inline-block; margin: 10px; padding: 15px; background: #e8f4f8; border-radius: 5px; }
        .insight { margin: 10px 0; padding: 15px; border-left: 4px solid #007acc; background: #f9f9f9; }
        .high-severity { border-left-color: #d32f2f; }
        .medium-severity { border-left-color: #f57c00; }
        .low-severity { border-left-color: #388e3c; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f5f5f5; }
        .cost { font-weight: bold; color: #d32f2f; }
        .savings { font-weight: bold; color: #388e3c; }
    </style>
</head>
<body>
    <div class="header">
        <h1>FinOps Cost Attribution Report</h1>
        <p><strong>Period:</strong> {{.Period}}</p>
        <p><strong>Generated:</strong> {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
    </div>

    <div class="section">
        <h2>Executive Summary</h2>
        <p>{{.ExecutiveSummary}}</p>
    </div>

    <div class="section">
        <h2>Cost Overview</h2>
        <div class="metric">
            <h3>Total Cost</h3>
            <div class="cost">${{.Summary.TotalCost}}</div>
        </div>
        <div class="metric">
            <h3>Number of Nodes</h3>
            <div>{{len .Summary.ByNode}}</div>
        </div>
        <div class="metric">
            <h3>Cost Dimensions</h3>
            <div>{{len .Summary.ByDimension}}</div>
        </div>
    </div>

    <div class="section">
        <h2>Top Cost Nodes</h2>
        <table>
            <tr>
                <th>Node Name</th>
                <th>Type</th>
                <th>Cost</th>
                <th>Percentage</th>
            </tr>
            {{range .Summary.TopCosts}}
            <tr>
                <td>{{.NodeName}}</td>
                <td>{{.NodeType}}</td>
                <td class="cost">${{.Cost}}</td>
                <td>{{printf "%.1f" .Percentage}}%</td>
            </tr>
            {{end}}
        </table>
    </div>

    <div class="section">
        <h2>Cost by Dimension</h2>
        <table>
            <tr>
                <th>Dimension</th>
                <th>Cost</th>
            </tr>
            {{range $dim, $cost := .Summary.ByDimension}}
            <tr>
                <td>{{$dim}}</td>
                <td class="cost">${{$cost}}</td>
            </tr>
            {{end}}
        </table>
    </div>

    <div class="section">
        <h2>Optimization Insights</h2>
        {{range .Insights}}
        <div class="insight {{.Severity}}-severity">
            <h3>{{.Title}}</h3>
            <p><strong>Node:</strong> {{.NodeName}}</p>
            <p><strong>Current Cost:</strong> <span class="cost">${{.CurrentCost}}</span></p>
            <p><strong>Potential Savings:</strong> <span class="savings">${{.PotentialSavings}}</span></p>
            <p><strong>Description:</strong> {{.Description}}</p>
            <p><strong>Recommendation:</strong> {{.Recommendation}}</p>
        </div>
        {{end}}
    </div>

    <div class="section">
        <h2>Key Recommendations</h2>
        <ul>
            {{range .Recommendations}}
            <li>{{.}}</li>
            {{end}}
        </ul>
    </div>

    <div class="section">
        <h2>Allocation Efficiency</h2>
        <table>
            <tr>
                <th>Node</th>
                <th>Direct Cost Ratio</th>
                <th>Indirect Cost Ratio</th>
                <th>Allocation Accuracy</th>
                <th>Efficiency Score</th>
            </tr>
            {{range .Efficiency}}
            <tr>
                <td>{{.NodeName}}</td>
                <td>{{printf "%.1f" (mul .DirectCostRatio 100)}}%</td>
                <td>{{printf "%.1f" (mul .IndirectCostRatio 100)}}%</td>
                <td>{{printf "%.1f" (mul .AllocationAccuracy 100)}}%</td>
                <td>{{printf "%.1f" (mul .EfficiencyScore 100)}}%</td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>`

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
	}

	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := t.Execute(file, report); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// generateRecommendations creates actionable recommendations based on analysis
func (rg *ReportGenerator) generateRecommendations(summary *analysis.CostSummary, insights []analysis.CostOptimizationInsight) []string {
	var recommendations []string

	// High-level recommendations based on total cost
	if summary.TotalCost.GreaterThan(summary.TotalCost.Mul(summary.TotalCost.Div(summary.TotalCost))) {
		recommendations = append(recommendations, "Implement cost monitoring alerts to track spending trends")
	}

	// Recommendations based on insights
	highCostNodes := 0
	underutilizedNodes := 0
	storageOptimizations := 0

	for _, insight := range insights {
		switch insight.Type {
		case "high_cost":
			highCostNodes++
		case "underutilized":
			underutilizedNodes++
		case "storage_optimization":
			storageOptimizations++
		}
	}

	if highCostNodes > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d high-cost nodes for rightsizing opportunities", highCostNodes))
	}

	if underutilizedNodes > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Investigate %d potentially underutilized resources for termination", underutilizedNodes))
	}

	if storageOptimizations > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Implement storage optimization for %d nodes with high storage costs", storageOptimizations))
	}

	// General recommendations
	recommendations = append(recommendations, "Set up automated cost allocation runs for regular reporting")
	recommendations = append(recommendations, "Implement tagging strategy for better cost attribution")
	recommendations = append(recommendations, "Consider reserved instances for predictable workloads")

	return recommendations
}

// generateExecutiveSummary creates an executive summary
func (rg *ReportGenerator) generateExecutiveSummary(summary *analysis.CostSummary, insights []analysis.CostOptimizationInsight) string {
	totalSavings := summary.TotalCost.Mul(summary.TotalCost.Div(summary.TotalCost)).Sub(summary.TotalCost)
	for _, insight := range insights {
		totalSavings = totalSavings.Add(insight.PotentialSavings)
	}

	return fmt.Sprintf(
		"During the period %s, total cloud costs were $%s across %d nodes and %d cost dimensions. "+
		"Analysis identified %d optimization opportunities with potential savings of $%s. "+
		"The top cost driver is %s, representing %.1f%% of total spend. "+
		"Key focus areas include cost optimization, resource rightsizing, and improved allocation accuracy.",
		summary.Period,
		summary.TotalCost.String(),
		len(summary.ByNode),
		len(summary.ByDimension),
		len(insights),
		totalSavings.String(),
		func() string {
			if len(summary.TopCosts) > 0 {
				return summary.TopCosts[0].NodeName
			}
			return "unknown"
		}(),
		func() float64 {
			if len(summary.TopCosts) > 0 {
				return summary.TopCosts[0].Percentage
			}
			return 0.0
		}(),
	)
}
