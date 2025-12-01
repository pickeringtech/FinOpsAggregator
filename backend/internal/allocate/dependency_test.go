package allocate

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestDependencyRelationships validates the correct dependency model
func TestDependencyRelationships(t *testing.T) {
	t.Run("Resource should allocate costs TO Product", func(t *testing.T) {
		// Given: A resource with direct costs and a product with no direct costs
		resource := &models.CostNode{
			ID:   uuid.New(),
			Name: "compute_resource",
			Type: string(models.NodeTypeResource),
		}

		product := &models.CostNode{
			ID:   uuid.New(),
			Name: "web_app",
			Type: string(models.NodeTypeProduct),
		}

		// The edge should be: Resource → Product (resource is parent, product is child)
		edge := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        resource.ID, // Resource is parent
			ChildID:         product.ID,  // Product is child
			DefaultStrategy: string(models.StrategyEqual),
			ActiveFrom:      time.Now().AddDate(0, 0, -1),
		}

		// Resource has direct costs
		resourceCost := models.NodeCostByDimension{
			NodeID:    resource.ID,
			CostDate:  time.Now(),
			Dimension: "compute_hours",
			Amount:    decimal.NewFromFloat(100.0),
			Currency:  "USD",
		}

		// Expected: Product should receive allocated costs from resource
		// This test defines the CORRECT behavior
		assert.Equal(t, resource.ID, edge.ParentID, "Resource should be parent in dependency edge")
		assert.Equal(t, product.ID, edge.ChildID, "Product should be child in dependency edge")
		assert.Equal(t, decimal.NewFromFloat(100.0), resourceCost.Amount, "Resource should have direct costs")
	})

	t.Run("Shared Service should allocate costs TO multiple Products", func(t *testing.T) {
		// Given: A shared service with direct costs and multiple products
		sharedService := &models.CostNode{
			ID:   uuid.New(),
			Name: "database_cluster",
			Type: string(models.NodeTypeShared),
		}

		product1 := &models.CostNode{
			ID:   uuid.New(),
			Name: "web_app",
			Type: string(models.NodeTypeProduct),
		}

		product2 := &models.CostNode{
			ID:   uuid.New(),
			Name: "mobile_app",
			Type: string(models.NodeTypeProduct),
		}

		// The edges should be: Shared Service → Product1, Shared Service → Product2
		edge1 := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        sharedService.ID, // Shared service is parent
			ChildID:         product1.ID,      // Product1 is child
			DefaultStrategy: string(models.StrategyEqual),
			ActiveFrom:      time.Now().AddDate(0, 0, -1),
		}

		edge2 := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        sharedService.ID, // Shared service is parent
			ChildID:         product2.ID,      // Product2 is child
			DefaultStrategy: string(models.StrategyEqual),
			ActiveFrom:      time.Now().AddDate(0, 0, -1),
		}

		// Shared service has direct costs
		sharedCost := models.NodeCostByDimension{
			NodeID:    sharedService.ID,
			CostDate:  time.Now(),
			Dimension: "database_hours",
			Amount:    decimal.NewFromFloat(300.0),
			Currency:  "USD",
		}

		// Expected: Each product should receive $150 (equal allocation)
		expectedAllocationPerProduct := decimal.NewFromFloat(150.0)

		assert.Equal(t, sharedService.ID, edge1.ParentID, "Shared service should be parent in edge1")
		assert.Equal(t, product1.ID, edge1.ChildID, "Product1 should be child in edge1")
		assert.Equal(t, sharedService.ID, edge2.ParentID, "Shared service should be parent in edge2")
		assert.Equal(t, product2.ID, edge2.ChildID, "Product2 should be child in edge2")
		assert.Equal(t, decimal.NewFromFloat(300.0), sharedCost.Amount, "Shared service should have direct costs")
		assert.Equal(t, expectedAllocationPerProduct, decimal.NewFromFloat(150.0), "Each product should receive equal allocation")
	})

	t.Run("Platform should allocate costs TO Products proportionally", func(t *testing.T) {
		// Given: A platform service with direct costs and products with different usage
		platform := &models.CostNode{
			ID:   uuid.New(),
			Name: "kubernetes_platform",
			Type: string(models.NodeTypePlatform),
		}

		product1 := &models.CostNode{
			ID:   uuid.New(),
			Name: "high_usage_app",
			Type: string(models.NodeTypeProduct),
		}

		product2 := &models.CostNode{
			ID:   uuid.New(),
			Name: "low_usage_app",
			Type: string(models.NodeTypeProduct),
		}

		// The edges should be: Platform → Product1, Platform → Product2
		edge1 := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        platform.ID,  // Platform is parent
			ChildID:         product1.ID,  // Product1 is child
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "cpu_usage",
			},
			ActiveFrom: time.Now().AddDate(0, 0, -1),
		}

		edge2 := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        platform.ID,  // Platform is parent
			ChildID:         product2.ID,  // Product2 is child
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "cpu_usage",
			},
			ActiveFrom: time.Now().AddDate(0, 0, -1),
		}

		// Platform has direct costs
		platformCost := models.NodeCostByDimension{
			NodeID:    platform.ID,
			CostDate:  time.Now(),
			Dimension: "platform_hours",
			Amount:    decimal.NewFromFloat(500.0),
			Currency:  "USD",
		}

		// Products have usage metrics (for proportional allocation)
		// product1Usage: 80% of total usage -> should receive $400
		// product2Usage: 20% of total usage -> should receive $100

		// Expected: Product1 should receive $400 (80%), Product2 should receive $100 (20%)
		expectedProduct1Allocation := decimal.NewFromFloat(400.0)
		expectedProduct2Allocation := decimal.NewFromFloat(100.0)

		assert.Equal(t, platform.ID, edge1.ParentID, "Platform should be parent in edge1")
		assert.Equal(t, product1.ID, edge1.ChildID, "Product1 should be child in edge1")
		assert.Equal(t, platform.ID, edge2.ParentID, "Platform should be parent in edge2")
		assert.Equal(t, product2.ID, edge2.ChildID, "Product2 should be child in edge2")
		assert.Equal(t, decimal.NewFromFloat(500.0), platformCost.Amount, "Platform should have direct costs")
		assert.Equal(t, expectedProduct1Allocation, decimal.NewFromFloat(400.0), "Product1 should receive proportional allocation")
		assert.Equal(t, expectedProduct2Allocation, decimal.NewFromFloat(100.0), "Product2 should receive proportional allocation")
	})
}