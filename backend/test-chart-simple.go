package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/charts"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/shopspring/decimal"
)

// MockStore for testing
type MockStore struct{}

type MockNodeRepository struct{}
type MockCostRepository struct{}

func (m *MockNodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CostNode, error) {
	return &models.CostNode{
		ID:   id,
		Name: "test-node",
		Type: "product",
	}, nil
}

func (m *MockCostRepository) GetByNodeAndDateRange(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimensions []string) ([]models.NodeCostByDimension, error) {
	// Return some test data
	return []models.NodeCostByDimension{
		{
			NodeID:    nodeID,
			CostDate:  startDate,
			Dimension: dimensions[0],
			Amount:    decimal.NewFromFloat(100.0),
			Currency:  "USD",
		},
		{
			NodeID:    nodeID,
			CostDate:  startDate.AddDate(0, 0, 1),
			Dimension: dimensions[0],
			Amount:    decimal.NewFromFloat(150.0),
			Currency:  "USD",
		},
	}, nil
}

// Create a mock store that satisfies the interface
type TestStore struct {
	Nodes *MockNodeRepository
	Costs *MockCostRepository
}

func main() {
	fmt.Println("Testing chart generation...")

	// Create mock store
	store := &TestStore{
		Nodes: &MockNodeRepository{},
		Costs: &MockCostRepository{},
	}

	// Create renderer
	renderer := charts.NewGraphRenderer(store)

	// Test 1: No data chart
	fmt.Println("Test 1: No data chart")
	var buf bytes.Buffer
	err := renderer.RenderNoDataChart(context.Background(), "Test message", &buf, "png")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: Generated %d bytes\n", buf.Len())
	}

	// Test 2: Cost trend chart
	fmt.Println("Test 2: Cost trend chart")
	nodeID := uuid.New()
	buf.Reset()
	err = renderer.RenderCostTrend(
		context.Background(),
		nodeID,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		"instance_hours",
		&buf,
		"png",
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: Generated %d bytes\n", buf.Len())
		
		// Write to file for inspection
		if err := os.WriteFile("test-trend.png", buf.Bytes(), 0644); err != nil {
			fmt.Printf("Failed to write file: %v\n", err)
		} else {
			fmt.Println("Wrote test-trend.png")
		}
	}

	fmt.Println("Chart generation test complete!")
}
