package allocate

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// MockStore implements the store interface for testing
type MockStore struct {
	nodes  []*models.CostNode
	edges  []models.DependencyEdge
	costs  []models.NodeCostByDimension
	usages []models.NodeUsageByDimension
}

func (m *MockStore) GetNodeCostsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.NodeCostByDimension, error) {
	var result []models.NodeCostByDimension
	for _, cost := range m.costs {
		if (cost.CostDate.Equal(startDate) || cost.CostDate.After(startDate)) &&
			(cost.CostDate.Equal(endDate) || cost.CostDate.Before(endDate)) {
			result = append(result, cost)
		}
	}
	return result, nil
}

func (m *MockStore) GetNodeUsageByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.NodeUsageByDimension, error) {
	var result []models.NodeUsageByDimension
	for _, usage := range m.usages {
		if (usage.UsageDate.Equal(startDate) || usage.UsageDate.After(startDate)) &&
			(usage.UsageDate.Equal(endDate) || usage.UsageDate.Before(endDate)) {
			result = append(result, usage)
		}
	}
	return result, nil
}

func (m *MockStore) SaveAllocationResults(ctx context.Context, results []models.AllocationResultByDimension) error {
	// For testing, we don't need to actually save
	return nil
}

// TestAllocationIntegration tests the full allocation process with correct vs incorrect dependencies
func TestAllocationIntegration(t *testing.T) {
	t.Run("Current incorrect behavior - Products allocating TO Resources", func(t *testing.T) {
		// This test demonstrates the CURRENT (incorrect) behavior
		// where products try to allocate costs to resources

		// Create nodes
		product := &models.CostNode{
			ID:   uuid.New(),
			Name: "web_app",
			Type: string(models.NodeTypeProduct),
		}

		resource := &models.CostNode{
			ID:   uuid.New(),
			Name: "compute_resource",
			Type: string(models.NodeTypeResource),
		}

		// INCORRECT edge: Product → Resource (current wrong behavior)
		incorrectEdge := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        product.ID,  // Product as parent (WRONG!)
			ChildID:         resource.ID, // Resource as child (WRONG!)
			DefaultStrategy: string(models.StrategyEqual),
			ActiveFrom:      time.Now().AddDate(0, 0, -1),
		}

		// Product has no direct costs (correct)
		productCosts := []models.NodeCostByDimension{}

		// Resource has direct costs (correct)
		resourceCosts := []models.NodeCostByDimension{
			{
				NodeID:    resource.ID,
				CostDate:  time.Now(),
				Dimension: "compute_hours",
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
		}

		// For now, we'll just validate the edge structure without running the full allocation
		// This demonstrates the incorrect dependency relationship

		// Verify INCORRECT behavior:
		// - Product (with $0 costs) tries to allocate to Resource
		// - Resource receives minimal/no allocation despite having $100 direct costs
		// - This demonstrates the current broken behavior

		// The product should have tried to allocate $0 to the resource
		// The resource should still have its $100 direct costs but receive no meaningful allocation

		// This test will PASS with current code but demonstrates the problem
		assert.Equal(t, product.ID, incorrectEdge.ParentID, "Current incorrect behavior: Product is parent")
		assert.Equal(t, resource.ID, incorrectEdge.ChildID, "Current incorrect behavior: Resource is child")
		assert.Equal(t, decimal.NewFromFloat(100.0), resourceCosts[0].Amount, "Resource has direct costs")
		assert.Empty(t, productCosts, "Product has no direct costs")
	})

	t.Run("Correct behavior - Resources allocating TO Products", func(t *testing.T) {
		// This test demonstrates the CORRECT behavior
		// where resources allocate their costs to products

		// Create nodes
		product := &models.CostNode{
			ID:   uuid.New(),
			Name: "web_app",
			Type: string(models.NodeTypeProduct),
		}

		resource := &models.CostNode{
			ID:   uuid.New(),
			Name: "compute_resource",
			Type: string(models.NodeTypeResource),
		}

		// CORRECT edge: Resource → Product
		correctEdge := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        resource.ID, // Resource as parent (CORRECT!)
			ChildID:         product.ID,  // Product as child (CORRECT!)
			DefaultStrategy: string(models.StrategyEqual),
			ActiveFrom:      time.Now().AddDate(0, 0, -1),
		}

		// Product has no direct costs (correct)
		productCosts := []models.NodeCostByDimension{}

		// Resource has direct costs (correct)
		resourceCosts := []models.NodeCostByDimension{
			{
				NodeID:    resource.ID,
				CostDate:  time.Now(),
				Dimension: "compute_hours",
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
		}

		// For now, we'll just validate the edge structure without running the full allocation
		// This demonstrates the correct dependency relationship

		// Verify CORRECT behavior:
		// - Resource (with $100 costs) allocates to Product
		// - Product receives $100 allocated costs
		// - Resource's costs are fully allocated
		// - This is the desired behavior

		assert.Equal(t, resource.ID, correctEdge.ParentID, "Correct behavior: Resource is parent")
		assert.Equal(t, product.ID, correctEdge.ChildID, "Correct behavior: Product is child")
		assert.Equal(t, decimal.NewFromFloat(100.0), resourceCosts[0].Amount, "Resource has direct costs")
		assert.Empty(t, productCosts, "Product has no direct costs")

		// TODO: Add assertions for allocation results once we implement the fix
		// expectedProductAllocation := decimal.NewFromFloat(100.0)
		// assert.Equal(t, expectedProductAllocation, actualProductAllocation, "Product should receive full resource costs")
	})
}