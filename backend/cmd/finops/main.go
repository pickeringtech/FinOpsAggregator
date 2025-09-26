package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/allocate"
	"github.com/pickeringtech/FinOpsAggregator/internal/config"
	"github.com/pickeringtech/FinOpsAggregator/internal/demo"
	"github.com/pickeringtech/FinOpsAggregator/internal/graph"
	"github.com/pickeringtech/FinOpsAggregator/internal/logging"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
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
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(demoCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from various sources",
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
		result, err := engine.AllocateForPeriod(context.Background(), startDate, endDate, cfg.Compute.ActiveDimensions)
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
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data and generate reports",
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch terminal user interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Launching TUI...")
		// TODO: Implement TUI launch
		return nil
	},
}

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Demo data and examples",
}

func init() {
	// Import subcommands
	importCmd.AddCommand(&cobra.Command{
		Use:   "costs [file]",
		Short: "Import cost data from CSV",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Importing costs from %s\n", args[0])
			// TODO: Implement cost import
			return nil
		},
	})

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
		Use:   "trend",
		Short: "Generate trend chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			node, _ := cmd.Flags().GetString("node")
			out, _ := cmd.Flags().GetString("out")
			fmt.Printf("Generating trend chart for %s to %s\n", node, out)
			// TODO: Implement trend chart
			return nil
		},
	})

	chartCmd.AddCommand(&cobra.Command{
		Use:   "waterfall",
		Short: "Generate waterfall chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			node, _ := cmd.Flags().GetString("node")
			out, _ := cmd.Flags().GetString("out")
			fmt.Printf("Generating waterfall chart for %s to %s\n", node, out)
			// TODO: Implement waterfall chart
			return nil
		},
	})

	// Chart flags
	for _, subCmd := range chartCmd.Commands() {
		subCmd.Flags().String("node", "", "Node ID to chart")
		subCmd.Flags().String("out", "", "Output file path")
		subCmd.MarkFlagRequired("node")
		subCmd.MarkFlagRequired("out")
	}

	exportCmd.AddCommand(chartCmd)

	exportCmd.AddCommand(&cobra.Command{
		Use:   "csv",
		Short: "Export data to CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("out")
			labels, _ := cmd.Flags().GetString("labels")
			fmt.Printf("Exporting CSV to %s with labels %s\n", out, labels)
			// TODO: Implement CSV export
			return nil
		},
	})

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
	synthCmd := demoCmd.Commands()[1] // synth command
	synthCmd.Flags().Int("nodes", 1000, "Number of nodes")
	synthCmd.Flags().Int("edges", 3000, "Number of edges")
	synthCmd.Flags().Int("days", 30, "Number of days")
	synthCmd.Flags().Int("dimensions", 6, "Number of dimensions")
}
