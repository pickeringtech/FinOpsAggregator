package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/api"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the HTTP API server",
	Long:  "Start the HTTP API server for the FinOps aggregation tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting FinOps API server...")

		// Create server configuration from config
		serverConfig := api.ServerConfig{
			Host:         cfg.API.Host,
			Port:         cfg.API.Port,
			ReadTimeout:  cfg.API.ReadTimeout,
			WriteTimeout: cfg.API.WriteTimeout,
			IdleTimeout:  cfg.API.IdleTimeout,
		}

		// Create and start server
		server := api.NewServer(serverConfig, st)

		// Start server in a goroutine
		go func() {
			if err := server.Start(); err != nil {
				fmt.Printf("Server error: %v\n", err)
				os.Exit(1)
			}
		}()

		fmt.Printf("API server started on %s:%d\n", serverConfig.Host, serverConfig.Port)
		fmt.Println("Available endpoints:")
		fmt.Println("  GET /health                     - Health check")
		fmt.Println("  GET /api/v1/products/hierarchy  - Product hierarchy with costs")
		fmt.Println("  GET /api/v1/nodes/{nodeId}      - Individual node cost data")
		fmt.Println("  GET /api/v1/platform/services   - Platform and shared services")
		fmt.Println()
		fmt.Println("Press Ctrl+C to stop the server")

		// Wait for interrupt signal
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		fmt.Println("\nShutting down server...")

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		if err := server.Stop(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}

		fmt.Println("Server stopped")
		return nil
	},
}
