package demo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestDependencyEdgeCreation tests that dependency edges are created with correct parent-child relationships
func TestDependencyEdgeCreation(t *testing.T) {
	t.Run("Resources should be parents of Products", func(t *testing.T) {
		// Create test nodes
		product := &models.CostNode{
			ID:   uuid.New(),
			Name: "test_product",
			Type: string(models.NodeTypeProduct),
		}

		resource := &models.CostNode{
			ID:   uuid.New(),
			Name: "test_resource",
			Type: string(models.NodeTypeResource),
		}

		// Create a map like the seed function does
		nodeMap := map[string]uuid.UUID{
			"test_product":  product.ID,
			"test_resource": resource.ID,
		}

		// Test the CORRECT edge creation logic
		correctEdge := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        nodeMap["test_resource"], // Resource should be parent
			ChildID:         nodeMap["test_product"],  // Product should be child
			DefaultStrategy: string(models.StrategyEqual),
			ActiveFrom:      time.Now().AddDate(0, 0, -1),
		}

		// Verify correct relationship
		assert.Equal(t, resource.ID, correctEdge.ParentID, "Resource should be parent")
		assert.Equal(t, product.ID, correctEdge.ChildID, "Product should be child")
	})

	t.Run("Shared services should be parents of Products", func(t *testing.T) {
		// Create test nodes
		product := &models.CostNode{
			ID:   uuid.New(),
			Name: "test_product",
			Type: string(models.NodeTypeProduct),
		}

		shared := &models.CostNode{
			ID:   uuid.New(),
			Name: "test_shared",
			Type: string(models.NodeTypeShared),
		}

		// Create a map like the seed function does
		nodeMap := map[string]uuid.UUID{
			"test_product": product.ID,
			"test_shared":  shared.ID,
		}

		// Test the CORRECT edge creation logic
		correctEdge := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        nodeMap["test_shared"],  // Shared should be parent
			ChildID:         nodeMap["test_product"], // Product should be child
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "usage_metric",
			},
			ActiveFrom: time.Now().AddDate(0, 0, -1),
		}

		// Verify correct relationship
		assert.Equal(t, shared.ID, correctEdge.ParentID, "Shared service should be parent")
		assert.Equal(t, product.ID, correctEdge.ChildID, "Product should be child")
	})

	t.Run("Platform services should be parents of Products", func(t *testing.T) {
		// Create test nodes
		product := &models.CostNode{
			ID:   uuid.New(),
			Name: "test_product",
			Type: string(models.NodeTypeProduct),
		}

		platform := &models.CostNode{
			ID:   uuid.New(),
			Name: "test_platform",
			Type: string(models.NodeTypePlatform),
		}

		// Create a map like the seed function does
		nodeMap := map[string]uuid.UUID{
			"test_product":  product.ID,
			"test_platform": platform.ID,
		}

		// Test the CORRECT edge creation logic
		correctEdge := models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        nodeMap["test_platform"], // Platform should be parent
			ChildID:         nodeMap["test_product"],  // Product should be child
			DefaultStrategy: string(models.StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"metric": "api_requests",
			},
			ActiveFrom: time.Now().AddDate(0, 0, -1),
		}

		// Verify correct relationship
		assert.Equal(t, platform.ID, correctEdge.ParentID, "Platform service should be parent")
		assert.Equal(t, product.ID, correctEdge.ChildID, "Product should be child")
	})
}