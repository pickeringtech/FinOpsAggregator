// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/charts"
	"github.com/pickeringtech/FinOpsAggregator/internal/config"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
)

func main() {
	fmt.Println("🧪 Direct Chart Generation Test")
	fmt.Println("===============================")

	// Test 1: Create a no-data chart without any database
	fmt.Println("Test 1: No-data chart (no database required)")
	
	// Create a nil store to simulate no database
	renderer := charts.NewGraphRenderer(nil)
	
	// Create output file
	file, err := os.Create("test-direct-no-data.png")
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		return
	}
	defer file.Close()
	
	// Try to render - this should create a no-data chart
	err = renderer.RenderNoDataChart(context.Background(), "Test: No database connection", file, "png")
	if err != nil {
		fmt.Printf("❌ Failed to render no-data chart: %v\n", err)
		return
	}
	
	// Check file size
	info, err := file.Stat()
	if err != nil {
		fmt.Printf("❌ Failed to get file info: %v\n", err)
		return
	}
	
	fmt.Printf("✅ No-data chart created: %d bytes\n", info.Size())
	file.Close()

	// Test 2: Try with a real database connection (if available)
	fmt.Println("\nTest 2: Chart with database (if available)")
	
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("⚠️  Config load failed: %v\n", err)
		fmt.Println("Skipping database test")
		return
	}
	
	// Try to connect to database
	db, err := store.NewDB(cfg.Database)
	if err != nil {
		fmt.Printf("⚠️  Database connection failed: %v\n", err)
		fmt.Println("This is expected if database is not configured")
		return
	}
	defer db.Close()
	
	// Create store
	st := store.NewStore(db)
	
	// Create renderer with real store
	rendererWithDB := charts.NewGraphRenderer(st)
	
	// Try to render graph structure
	file2, err := os.Create("test-direct-with-db.png")
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		return
	}
	defer file2.Close()
	
	err = rendererWithDB.RenderGraphStructure(context.Background(), time.Now(), file2, "png")
	if err != nil {
		fmt.Printf("⚠️  Graph structure render failed: %v\n", err)
		fmt.Println("This might be expected if no data exists")
	} else {
		info2, _ := file2.Stat()
		fmt.Printf("✅ Graph structure chart created: %d bytes\n", info2.Size())
	}
	
	fmt.Println("\n🎉 Direct chart test complete!")
	fmt.Println("Check the generated files:")
	fmt.Println("  - test-direct-no-data.png")
	fmt.Println("  - test-direct-with-db.png (if database was available)")
}
