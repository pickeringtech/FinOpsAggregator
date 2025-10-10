package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/analysis"
	"github.com/pickeringtech/FinOpsAggregator/internal/reports"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

func init() {
	// Add analyze subcommands
	analyzeCmd.AddCommand(analyzeCostsCmd)
	analyzeCmd.AddCommand(analyzeOptimizationCmd)
	analyzeCmd.AddCommand(analyzeProductsCmd)
	analyzeCmd.AddCommand(analyzeEfficiencyCmd)
	
	// Add report subcommands
	reportCmd.AddCommand(reportGenerateCmd)
	reportCmd.AddCommand(reportExportCmd)
	
	// Add flags
	analyzeCostsCmd.Flags().StringP("from", "f", "", "Start date (YYYY-MM-DD)")
	analyzeCostsCmd.Flags().StringP("to", "t", "", "End date (YYYY-MM-DD)")
	analyzeCostsCmd.Flags().StringP("format", "o", "table", "Output format (table, json)")
	analyzeCostsCmd.Flags().IntP("top", "n", 10, "Number of top items to show")
	
	analyzeOptimizationCmd.Flags().StringP("from", "f", "", "Start date (YYYY-MM-DD)")
	analyzeOptimizationCmd.Flags().StringP("to", "t", "", "End date (YYYY-MM-DD)")
	analyzeOptimizationCmd.Flags().StringP("format", "o", "table", "Output format (table, json)")
	analyzeOptimizationCmd.Flags().StringP("severity", "s", "", "Filter by severity (high, medium, low)")

	analyzeProductsCmd.Flags().StringP("from", "f", "", "Start date (YYYY-MM-DD)")
	analyzeProductsCmd.Flags().StringP("to", "t", "", "End date (YYYY-MM-DD)")
	analyzeProductsCmd.Flags().StringP("format", "o", "table", "Output format (table, json)")
	
	reportGenerateCmd.Flags().StringP("from", "f", "", "Start date (YYYY-MM-DD)")
	reportGenerateCmd.Flags().StringP("to", "t", "", "End date (YYYY-MM-DD)")
	reportGenerateCmd.Flags().StringP("output", "o", "finops-report.html", "Output filename")
	reportGenerateCmd.Flags().StringP("format", "F", "html", "Output format (html, json)")
}

var analyzeCostsCmd = &cobra.Command{
	Use:   "costs",
	Short: "Analyze cost breakdown and trends",
	Long:  "Analyze cost breakdown by nodes and dimensions with trend analysis",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse flags
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		format, _ := cmd.Flags().GetString("format")
		topN, _ := cmd.Flags().GetInt("top")
		
		// Set default date range (last 30 days)
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		
		if fromStr != "" {
			var err error
			startDate, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				return fmt.Errorf("invalid start date: %w", err)
			}
		}
		
		if toStr != "" {
			var err error
			endDate, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				return fmt.Errorf("invalid end date: %w", err)
			}
		}
		
		// Analyze costs
		analyzer := analysis.NewFinOpsAnalyzer(st)
		summary, err := analyzer.AnalyzeCosts(context.Background(), startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to analyze costs: %w", err)
		}
		
		// Output results
		switch format {
		case "json":
			return outputJSON(summary)
		default:
			return outputCostTable(summary, topN)
		}
	},
}

var analyzeOptimizationCmd = &cobra.Command{
	Use:   "optimization",
	Short: "Generate cost optimization insights",
	Long:  "Analyze costs and generate actionable optimization recommendations",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse flags
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		format, _ := cmd.Flags().GetString("format")
		severity, _ := cmd.Flags().GetString("severity")
		
		// Set default date range (last 30 days)
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		
		if fromStr != "" {
			var err error
			startDate, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				return fmt.Errorf("invalid start date: %w", err)
			}
		}
		
		if toStr != "" {
			var err error
			endDate, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				return fmt.Errorf("invalid end date: %w", err)
			}
		}
		
		// Generate insights
		analyzer := analysis.NewFinOpsAnalyzer(st)
		insights, err := analyzer.GenerateOptimizationInsights(context.Background(), startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to generate insights: %w", err)
		}
		
		// Filter by severity if specified
		if severity != "" {
			var filtered []analysis.CostOptimizationInsight
			for _, insight := range insights {
				if insight.Severity == severity {
					filtered = append(filtered, insight)
				}
			}
			insights = filtered
		}
		
		// Output results
		switch format {
		case "json":
			return outputJSON(insights)
		default:
			return outputOptimizationTable(insights)
		}
	},
}

var analyzeProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "Analyze costs by Product",
	Long:  "Analyze cost breakdown by Products and their dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse flags
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		format, _ := cmd.Flags().GetString("format")

		// Set default date range (last 30 days)
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)

		if fromStr != "" {
			var err error
			startDate, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				return fmt.Errorf("invalid start date: %w", err)
			}
		}

		if toStr != "" {
			var err error
			endDate, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				return fmt.Errorf("invalid end date: %w", err)
			}
		}

		// Analyze product costs
		analyzer := analysis.NewFinOpsAnalyzer(st)
		summary, err := analyzer.AnalyzeProductCosts(context.Background(), startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to analyze product costs: %w", err)
		}

		// Output results
		switch format {
		case "json":
			return outputJSON(summary)
		default:
			return outputProductTable(summary)
		}
	},
}

var analyzeEfficiencyCmd = &cobra.Command{
	Use:   "efficiency",
	Short: "Analyze allocation efficiency",
	Long:  "Analyze how efficiently costs are being allocated across the DAG",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Allocation efficiency analysis coming soon...")
		return nil
	},
}

var reportGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate comprehensive FinOps report",
	Long:  "Generate a comprehensive FinOps report with cost analysis, insights, and recommendations",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse flags
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		output, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		
		// Set default date range (last 30 days)
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		
		if fromStr != "" {
			var err error
			startDate, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				return fmt.Errorf("invalid start date: %w", err)
			}
		}
		
		if toStr != "" {
			var err error
			endDate, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				return fmt.Errorf("invalid end date: %w", err)
			}
		}
		
		// Generate report
		generator := reports.NewReportGenerator(st)
		report, err := generator.GenerateReport(context.Background(), startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}
		
		// Export report
		switch format {
		case "json":
			err = generator.ExportReportJSON(report, output)
		default:
			err = generator.ExportReportHTML(report, output)
		}
		
		if err != nil {
			return fmt.Errorf("failed to export report: %w", err)
		}
		
		fmt.Printf("✅ Report generated: %s\n", output)
		fmt.Printf("📊 Period: %s\n", report.Period)
		fmt.Printf("💰 Total Cost: $%s\n", report.Summary.TotalCost.String())
		fmt.Printf("💡 Insights: %d optimization opportunities\n", len(report.Insights))
		
		totalSavings := decimal.Zero
		for _, insight := range report.Insights {
			totalSavings = totalSavings.Add(insight.PotentialSavings)
		}
		fmt.Printf("💵 Potential Savings: $%s\n", totalSavings.String())
		
		return nil
	},
}

var reportExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export existing report data",
	Long:  "Export cost data in various formats for external analysis",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Report export functionality coming soon...")
		return nil
	},
}

// Helper functions for output formatting

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputCostTable(summary *analysis.CostSummary, topN int) error {
	fmt.Printf("📊 Cost Analysis - %s\n", summary.Period)
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")
	
	fmt.Printf("💰 Total Cost: $%s\n", summary.TotalCost.String())
	fmt.Printf("📅 Period: %s to %s\n", summary.StartDate.Format("2006-01-02"), summary.EndDate.Format("2006-01-02"))
	fmt.Printf("🏗️  Nodes: %d\n", len(summary.ByNode))
	fmt.Printf("📏 Dimensions: %d\n\n", len(summary.ByDimension))
	
	// Top cost nodes
	fmt.Printf("🔝 Top %d Cost Nodes:\n", min(topN, len(summary.TopCosts)))
	fmt.Printf("───────────────────────────────────────────────────────────────\n")
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Rank\tNode Name\tType\tCost\tPercentage")
	fmt.Fprintln(w, "────\t─────────\t────\t────\t──────────")
	
	for i, node := range summary.TopCosts {
		if i >= topN {
			break
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t$%s\t%.1f%%\n", 
			i+1, node.NodeName, node.NodeType, node.Cost.String(), node.Percentage)
	}
	w.Flush()
	
	// Cost by dimension
	fmt.Printf("\n📏 Cost by Dimension:\n")
	fmt.Printf("───────────────────────────────────────────────────────────────\n")
	
	// Sort dimensions by cost
	type dimCost struct {
		dimension string
		cost      decimal.Decimal
		percentage float64
	}
	
	var dimensions []dimCost
	for dim, cost := range summary.ByDimension {
		percentage := cost.Div(summary.TotalCost).InexactFloat64() * 100
		dimensions = append(dimensions, dimCost{dim, cost, percentage})
	}
	
	sort.Slice(dimensions, func(i, j int) bool {
		return dimensions[i].cost.GreaterThan(dimensions[j].cost)
	})
	
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Dimension\tCost\tPercentage")
	fmt.Fprintln(w, "─────────\t────\t──────────")
	
	for _, dim := range dimensions {
		fmt.Fprintf(w, "%s\t$%s\t%.1f%%\n", dim.dimension, dim.cost.String(), dim.percentage)
	}
	w.Flush()
	
	return nil
}

func outputOptimizationTable(insights []analysis.CostOptimizationInsight) error {
	if len(insights) == 0 {
		fmt.Println("✅ No optimization opportunities found!")
		return nil
	}
	
	fmt.Printf("💡 Cost Optimization Insights (%d opportunities)\n", len(insights))
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")
	
	totalSavings := decimal.Zero
	for _, insight := range insights {
		totalSavings = totalSavings.Add(insight.PotentialSavings)
	}
	
	fmt.Printf("💵 Total Potential Savings: $%s\n\n", totalSavings.String())
	
	// Group by severity
	severityGroups := make(map[string][]analysis.CostOptimizationInsight)
	for _, insight := range insights {
		severityGroups[insight.Severity] = append(severityGroups[insight.Severity], insight)
	}
	
	severityOrder := []string{"high", "medium", "low"}
	severityIcons := map[string]string{"high": "🔴", "medium": "🟡", "low": "🟢"}
	
	for _, severity := range severityOrder {
		if len(severityGroups[severity]) == 0 {
			continue
		}
		
		fmt.Printf("%s %s Priority (%d items):\n", severityIcons[severity], 
			map[string]string{"high": "High", "medium": "Medium", "low": "Low"}[severity],
			len(severityGroups[severity]))
		fmt.Printf("───────────────────────────────────────────────────────────────\n")
		
		for i, insight := range severityGroups[severity] {
			fmt.Printf("%d. %s\n", i+1, insight.Title)
			fmt.Printf("   Node: %s\n", insight.NodeName)
			if insight.Dimension != "" {
				fmt.Printf("   Dimension: %s\n", insight.Dimension)
			}
			fmt.Printf("   Current Cost: $%s\n", insight.CurrentCost.String())
			fmt.Printf("   Potential Savings: $%s\n", insight.PotentialSavings.String())
			fmt.Printf("   Recommendation: %s\n", insight.Recommendation)
			fmt.Println()
		}
	}
	
	return nil
}

func outputProductTable(summary *analysis.ProductCostSummary) error {
	fmt.Printf("🏢 Product Cost Analysis - %s\n", summary.Period)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("💰 Total Cost: %s\n", summary.TotalCost.StringFixed(2))
	fmt.Printf("📅 Period: %s\n", summary.Period)
	fmt.Printf("🏢 Products: %d\n", summary.ProductCount)
	fmt.Println()

	if len(summary.Products) == 0 {
		fmt.Println("No product costs found for the specified period.")
		return nil
	}

	fmt.Printf("🏢 Products by Cost:\n")
	fmt.Println("───────────────────────────────────────────────────────────────")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Rank\tProduct\tTotal Cost\tPercentage\tDependencies\n")
	fmt.Fprintf(w, "────\t───────\t──────────\t──────────\t────────────\n")

	for i, product := range summary.Products {
		fmt.Fprintf(w, "%d\t%s\t$%s\t%.1f%%\t%d nodes\n",
			i+1,
			product.ProductName,
			product.TotalCost.StringFixed(2),
			product.Percentage,
			product.DependentNodeCount,
		)
	}
	w.Flush()
	fmt.Println()

	// Show detailed breakdown for each product
	for _, product := range summary.Products {
		fmt.Printf("📊 %s - Cost Breakdown:\n", product.ProductName)
		fmt.Println("───────────────────────────────────────────────────────────────")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Node Name\tType\tCost\tPercentage\n")
		fmt.Fprintf(w, "─────────\t────\t────\t──────────\n")

		for _, node := range product.CostBreakdown {
			if node.Cost.IsPositive() {
				fmt.Fprintf(w, "%s\t%s\t$%s\t%.1f%%\n",
					node.NodeName,
					node.NodeType,
					node.Cost.StringFixed(2),
					node.Percentage,
				)
			}
		}
		w.Flush()
		fmt.Println()
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
