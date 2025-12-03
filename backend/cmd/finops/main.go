package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/allocate"
	"github.com/pickeringtech/FinOpsAggregator/internal/api"
	"github.com/pickeringtech/FinOpsAggregator/internal/charts"
	"github.com/pickeringtech/FinOpsAggregator/internal/config"
	"github.com/pickeringtech/FinOpsAggregator/internal/demo"
	"github.com/pickeringtech/FinOpsAggregator/internal/graph"
	"github.com/pickeringtech/FinOpsAggregator/internal/ingestion"
	"github.com/pickeringtech/FinOpsAggregator/internal/logging"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/pickeringtech/FinOpsAggregator/internal/tui"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
	db      *store.DB
	st      *store.Store
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "finops",
	Short: "FinOps DAG Cost Attribution Tool",
	Long: `A dimension-aware FinOps aggregation tool that models cost attribution
as a weighted directed acyclic graph (DAG) and provides both TUI and API interfaces.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize logging
		logging.Init(cfg.Logging)

		// Initialize database
		db, err = store.NewDB(cfg.Postgres)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		// Initialize store
		st = store.NewStore(db)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if db != nil {
			db.Close()
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Add subcommands
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(allocateCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(demoCmd)
	rootCmd.AddCommand(apiCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from various sources",
}

	func runAllocation(ctx context.Context, st *store.Store, from, to string) error {
		startDate, err := time.Parse("2006-01-02", from)
		if err != nil {
			return fmt.Errorf("invalid start date format: %w", err)
		}

		endDate, err := time.Parse("2006-01-02", to)
		if err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}

		fmt.Printf("Running allocation from %s to %s\n", from, to)

		engine := allocate.NewEngine(st)
		result, err := engine.AllocateForPeriod(ctx, startDate, endDate, cfg.Compute.ActiveDimensions)
		if err != nil {
			return fmt.Errorf("allocation failed: %w", err)
		}

		fmt.Printf("Allocation completed successfully!\n")
		fmt.Printf("Run ID: %s\n", result.RunID)
		fmt.Printf("Processed %d days\n", result.Summary.ProcessedDays)
		fmt.Printf("Total allocations: %d\n", len(result.Allocations))
		fmt.Printf("Total contributions: %d\n", len(result.Contributions))
		fmt.Printf("Processing time: %v\n", result.Summary.ProcessingTime)

		return nil
	}


var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Graph operations and validation",
}

var allocateCmd = &cobra.Command{
	Use:   "allocate",
	Short: "Run cost allocation computations",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		return runAllocation(cmd.Context(), st, from, to)
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data and generate reports",
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate comprehensive FinOps reports",
	Long:  "Generate detailed FinOps reports with cost analysis, optimization insights, and recommendations",
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch terminal user interface",
	Long:  "Launch an interactive terminal user interface for FinOps cost analysis and optimization",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Launch TUI application
		tuiApp := tui.NewApp(st)
		return tuiApp.Run()
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze costs and generate insights",
	Long:  "Perform cost analysis and generate optimization insights for FinOps decision making",
}

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Demo data and examples",
}

var importCostsCmd = &cobra.Command{
	Use:   "costs [file]",
	Short: "Import cost data from AWS CUR CSV file",
	Long: `Import cost data from an AWS Cost and Usage Report (CUR) CSV file.

The CSV file should contain standard AWS CUR columns including:
  - lineItem/UsageStartDate
  - lineItem/UnblendedCost
  - lineItem/ProductCode
  - resourceTags/user:Product (optional, for mapping to products)
  - resourceTags/user:Service (optional, for mapping to services)

Use --allocate to automatically run cost allocation after import.
Use --create-nodes to automatically create missing nodes from AWS product codes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		allocateAfter, _ := cmd.Flags().GetBool("allocate")
		createNodes, _ := cmd.Flags().GetBool("create-nodes")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		// Create progress channel for reporting
		progressChan := make(chan ingestion.IngestionProgress, 100)
		defer close(progressChan)

		// Start progress reporter goroutine
		done := make(chan struct{})
		go func() {
			for progress := range progressChan {
				if progress.Error != nil {
					fmt.Printf("  ERROR: %v\n", progress.Error)
				} else if progress.Message != "" {
					fmt.Printf("  %s\n", progress.Message)
				}
			}
			close(done)
		}()

		fmt.Printf("Importing AWS CUR costs from %s\n", filePath)
		fmt.Println("----------------------------------------")

		// Create ingester
		ingester := ingestion.NewAWSCURIngester(st, &ingestion.AWSCURConfig{
			BatchSize:          1000,
			CreateMissingNodes: createNodes,
			ProgressChan:       progressChan,
		})

		// Run ingestion
		ctx := cmd.Context()
		result, err := ingester.IngestFile(ctx, filePath)
		if err != nil {
			return fmt.Errorf("ingestion failed: %w", err)
		}

		// Wait for progress reporter to finish
		<-done

		fmt.Println("----------------------------------------")
		fmt.Printf("Ingestion completed!\n")
		fmt.Printf("  Records processed: %d\n", result.RecordsProcessed)
		fmt.Printf("  Records inserted:  %d\n", result.RecordsInserted)
		fmt.Printf("  Records skipped:   %d\n", result.RecordsSkipped)
		fmt.Printf("  Duration:          %v\n", result.Duration)

		if len(result.Errors) > 0 {
			fmt.Printf("  Errors (%d):\n", len(result.Errors))
			for i, e := range result.Errors {
				if i >= 5 {
					fmt.Printf("    ... and %d more errors\n", len(result.Errors)-5)
					break
				}
				fmt.Printf("    - %s\n", e)
			}
		}

		// Run allocation if requested
		if allocateAfter {
			fmt.Println()
			fmt.Println("Running allocation...")
			if err := runAllocation(ctx, st, from, to); err != nil {
				return fmt.Errorf("allocation failed: %w", err)
			}
		}

		return nil
	},
}

func init() {
	// Import costs command flags
	importCostsCmd.Flags().Bool("allocate", false, "Run cost allocation after import")
	importCostsCmd.Flags().Bool("create-nodes", false, "Create missing nodes from AWS product codes")
	importCostsCmd.Flags().String("from", time.Now().AddDate(0, 0, -30).Format("2006-01-02"), "Start date for allocation (YYYY-MM-DD)")
	importCostsCmd.Flags().String("to", time.Now().Format("2006-01-02"), "End date for allocation (YYYY-MM-DD)")

	// Import subcommands
	importCmd.AddCommand(importCostsCmd)

	importCmd.AddCommand(&cobra.Command{
		Use:   "usage [file]",
		Short: "Import usage data from CSV",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Importing usage from %s\n", args[0])
			// TODO: Implement usage import
			return nil
		},
	})

	// Graph subcommands
	graphCmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate graph structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Validating graph...")

			validator := graph.NewValidator(st)
			result, err := validator.ValidateCurrentGraph(context.Background())
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Printf("Graph validation completed\n")
			fmt.Printf("Valid: %t\n", result.Valid)
			fmt.Printf("Nodes: %d\n", result.Stats.NodeCount)
			fmt.Printf("Edges: %d\n", result.Stats.EdgeCount)
			fmt.Printf("Roots: %d\n", result.Stats.RootCount)
			fmt.Printf("Leaves: %d\n", result.Stats.LeafCount)

			if len(result.Errors) > 0 {
				fmt.Printf("\nErrors (%d):\n", len(result.Errors))
				for _, err := range result.Errors {
					fmt.Printf("  - %s: %s\n", err.Type, err.Message)
				}
			}

			if len(result.Warnings) > 0 {
				fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
				for _, warn := range result.Warnings {
					fmt.Printf("  - %s: %s\n", warn.Type, warn.Message)
				}
			}

			return nil
		},
	})

	// Allocate flags
	allocateCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	allocateCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	allocateCmd.MarkFlagRequired("from")
	allocateCmd.MarkFlagRequired("to")

	// Export subcommands
	chartCmd := &cobra.Command{
		Use:   "chart",
		Short: "Export charts",
	}

	chartCmd.AddCommand(&cobra.Command{
		Use:   "graph",
		Short: "Generate graph structure chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("out")
			format, _ := cmd.Flags().GetString("format")
			date, _ := cmd.Flags().GetString("date")

			// Parse date
			var chartDate time.Time
			var err error
			if date != "" {
				chartDate, err = time.Parse("2006-01-02", date)
				if err != nil {
					return fmt.Errorf("invalid date format: %w", err)
				}
			} else {
				chartDate = time.Now()
			}

			// Create exporter
			exporter, err := charts.NewExporter(st, cfg.Storage.URL, cfg.Storage.Prefix)
			if err != nil {
				return fmt.Errorf("failed to create chart exporter: %w", err)
			}
			defer exporter.Close()

			// Export graph structure
			if err := exporter.ExportGraphStructure(context.Background(), chartDate, out, format); err != nil {
				return fmt.Errorf("failed to export graph structure: %w", err)
			}

			fmt.Printf("Graph structure chart exported to: %s\n", out)
			return nil
		},
	})

	chartCmd.AddCommand(&cobra.Command{
		Use:   "trend",
		Short: "Generate trend chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeStr, _ := cmd.Flags().GetString("node")
			out, _ := cmd.Flags().GetString("out")
			format, _ := cmd.Flags().GetString("format")
			dimension, _ := cmd.Flags().GetString("dimension")
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")

			// Parse node ID
			nodeID, err := uuid.Parse(nodeStr)
			if err != nil {
				// Try to find node by name
				node, err := st.Nodes.GetByName(context.Background(), nodeStr)
				if err != nil {
					return fmt.Errorf("invalid node ID or name: %s", nodeStr)
				}
				nodeID = node.ID
			}

			// Parse dates
			startDate, err := time.Parse("2006-01-02", from)
			if err != nil {
				return fmt.Errorf("invalid start date format: %w", err)
			}

			endDate, err := time.Parse("2006-01-02", to)
			if err != nil {
				return fmt.Errorf("invalid end date format: %w", err)
			}

			// Create exporter
			exporter, err := charts.NewExporter(st, cfg.Storage.URL, cfg.Storage.Prefix)
			if err != nil {
				return fmt.Errorf("failed to create chart exporter: %w", err)
			}
			defer exporter.Close()

			// Export trend chart
			if err := exporter.ExportCostTrend(context.Background(), nodeID, startDate, endDate, dimension, out, format); err != nil {
				return fmt.Errorf("failed to export cost trend: %w", err)
			}

			fmt.Printf("Cost trend chart exported to: %s\n", out)
			return nil
		},
	})

	chartCmd.AddCommand(&cobra.Command{
		Use:   "waterfall",
		Short: "Generate waterfall chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeStr, _ := cmd.Flags().GetString("node")
			out, _ := cmd.Flags().GetString("out")
			format, _ := cmd.Flags().GetString("format")
			date, _ := cmd.Flags().GetString("date")
			runStr, _ := cmd.Flags().GetString("run")

			// Parse node ID
			nodeID, err := uuid.Parse(nodeStr)
			if err != nil {
				// Try to find node by name
				node, err := st.Nodes.GetByName(context.Background(), nodeStr)
				if err != nil {
					return fmt.Errorf("invalid node ID or name: %s", nodeStr)
				}
				nodeID = node.ID
			}

			// Parse run ID
			runID, err := uuid.Parse(runStr)
			if err != nil {
				return fmt.Errorf("invalid run ID: %s", runStr)
			}

			// Parse date
			chartDate, err := time.Parse("2006-01-02", date)
			if err != nil {
				return fmt.Errorf("invalid date format: %w", err)
			}

			// Create exporter
			exporter, err := charts.NewExporter(st, cfg.Storage.URL, cfg.Storage.Prefix)
			if err != nil {
				return fmt.Errorf("failed to create chart exporter: %w", err)
			}
			defer exporter.Close()

			// Export waterfall chart
			if err := exporter.ExportAllocationWaterfall(context.Background(), nodeID, chartDate, runID, out, format); err != nil {
				return fmt.Errorf("failed to export allocation waterfall: %w", err)
			}

			fmt.Printf("Allocation waterfall chart exported to: %s\n", out)
			return nil
		},
	})

	// Add flags directly to each command

	// Graph command flags
	graphCmd := chartCmd.Commands()[0]
	graphCmd.Flags().String("format", "png", "Output format (png, svg)")
	graphCmd.Flags().String("out", "", "Output file path (optional, auto-generated if not provided)")
	graphCmd.Flags().String("date", "", "Date for graph structure (YYYY-MM-DD, defaults to today)")

	// Trend command flags
	trendCmd := chartCmd.Commands()[1]
	trendCmd.Flags().String("format", "png", "Output format (png, svg)")
	trendCmd.Flags().String("out", "", "Output file path (optional, auto-generated if not provided)")
	trendCmd.Flags().String("node", "", "Node ID or name")
	trendCmd.Flags().String("dimension", "instance_hours", "Cost dimension")
	trendCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	trendCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	trendCmd.MarkFlagRequired("node")
	trendCmd.MarkFlagRequired("from")
	trendCmd.MarkFlagRequired("to")

	// Waterfall command flags
	waterfallCmd := chartCmd.Commands()[2]
	waterfallCmd.Flags().String("format", "png", "Output format (png, svg)")
	waterfallCmd.Flags().String("out", "", "Output file path (optional, auto-generated if not provided)")
	waterfallCmd.Flags().String("node", "", "Node ID or name")
	waterfallCmd.Flags().String("date", "", "Date for allocation (YYYY-MM-DD)")
	waterfallCmd.Flags().String("run", "", "Allocation run ID")
	waterfallCmd.MarkFlagRequired("node")
	waterfallCmd.MarkFlagRequired("date")
	waterfallCmd.MarkFlagRequired("run")

	exportCmd.AddCommand(chartCmd)

	csvCmd := &cobra.Command{
		Use:   "csv",
		Short: "Export data to CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCSVExport(cmd, args)
		},
	}

	// Add flags for CSV export
	csvCmd.Flags().String("type", "products", "Export type: products, nodes, costs_by_type, recommendations")
	csvCmd.Flags().String("node-type", "", "Node type filter (for nodes export)")
	csvCmd.Flags().String("node-id", "", "Node ID filter (for recommendations export)")
	csvCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	csvCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	csvCmd.Flags().String("currency", "USD", "Currency")
	csvCmd.Flags().String("out", "", "Output file (default: stdout)")

	exportCmd.AddCommand(csvCmd)

	// Demo subcommands
	demoCmd.AddCommand(&cobra.Command{
		Use:   "seed",
		Short: "Load demo seed data",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Loading demo seed data...")

			seeder := demo.NewSeeder(st)

			// Seed basic DAG structure
			if err := seeder.SeedBasicDAG(context.Background()); err != nil {
				return fmt.Errorf("failed to seed DAG: %w", err)
			}

			// Seed cost data
			if err := seeder.SeedCostData(context.Background()); err != nil {
				return fmt.Errorf("failed to seed cost data: %w", err)
			}

			// Seed usage data
			if err := seeder.SeedUsageData(context.Background()); err != nil {
				return fmt.Errorf("failed to seed usage data: %w", err)
			}

			fmt.Println("Demo seed data loaded successfully!")
			return nil
		},
	})

	demoCmd.AddCommand(&cobra.Command{
		Use:   "allocate",
		Short: "Run allocation for the full demo seed period",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Align with demo seed range: 12 months past to 12 months future
			now := time.Now()
			from := now.AddDate(0, -12, 0).Format("2006-01-02")
			to := now.AddDate(0, 12, 0).Format("2006-01-02")
			fmt.Printf("Running demo allocation from %s to %s...\n", from, to)
			return runAllocation(cmd.Context(), st, from, to)
		},
	})

	demoCmd.AddCommand(&cobra.Command{
		Use:   "synth",
		Short: "Generate synthetic data",
		RunE: func(cmd *cobra.Command, args []string) error {
			nodes, _ := cmd.Flags().GetInt("nodes")
			edges, _ := cmd.Flags().GetInt("edges")
			days, _ := cmd.Flags().GetInt("days")
			dimensions, _ := cmd.Flags().GetInt("dimensions")
			fmt.Printf("Generating synthetic data: %d nodes, %d edges, %d days, %d dimensions\n",
				nodes, edges, days, dimensions)
			// TODO: Implement synthetic data generation
			return nil
		},
	})

	// Demo synth flags
	synthCmd := demoCmd.Commands()[2] // synth command
	synthCmd.Flags().Int("nodes", 1000, "Number of nodes")
	synthCmd.Flags().Int("edges", 3000, "Number of edges")
	synthCmd.Flags().Int("days", 30, "Number of days")
	synthCmd.Flags().Int("dimensions", 6, "Number of dimensions")
}

// runCSVExport handles CSV export command
func runCSVExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse flags
	exportType, _ := cmd.Flags().GetString("type")
	nodeType, _ := cmd.Flags().GetString("node-type")
	nodeIDStr, _ := cmd.Flags().GetString("node-id")
	startDateStr, _ := cmd.Flags().GetString("start-date")
	endDateStr, _ := cmd.Flags().GetString("end-date")
	currency, _ := cmd.Flags().GetString("currency")
	outputFile, _ := cmd.Flags().GetString("out")

	// Parse dates
	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -30) // Default to 30 days ago
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return fmt.Errorf("invalid start date format: %w", err)
		}
	}

	if endDateStr == "" {
		endDate = time.Now() // Default to today
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}
	}

	// Parse node ID if provided
	var nodeID *uuid.UUID
	if nodeIDStr != "" {
		parsed, err := uuid.Parse(nodeIDStr)
		if err != nil {
			return fmt.Errorf("invalid node ID format: %w", err)
		}
		nodeID = &parsed
	}

	// Create API service
	service := api.NewService(st)

	// Create cost attribution request
	req := api.CostAttributionRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Currency:  currency,
	}

	// Determine output writer
	var writer *os.File
	if outputFile == "" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer writer.Close()
	}

	// Export based on type
	switch exportType {
	case "products":
		err = service.ExportProductsToCSV(ctx, req, writer)
	case "nodes":
		err = service.ExportNodesToCSV(ctx, req, nodeType, writer)
	case "costs_by_type":
		err = service.ExportCostsByTypeToCSV(ctx, req, writer)
	case "recommendations":
		err = service.ExportRecommendationsToCSV(ctx, req, nodeID, writer)
	case "detailed_costs":
		err = service.ExportDetailedCostsToCSV(ctx, req, nodeType, writer)
	case "raw_costs":
		err = service.ExportRawCostsToCSV(ctx, req, nodeType, writer)
	default:
		return fmt.Errorf("unsupported export type: %s. Supported types: products, nodes, costs_by_type, recommendations, detailed_costs, raw_costs", exportType)
	}

	if err != nil {
		return fmt.Errorf("failed to export CSV: %w", err)
	}

	if outputFile != "" {
		fmt.Printf("CSV exported to %s\n", outputFile)
	}

	return nil
}
