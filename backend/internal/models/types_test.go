package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCostNode_Structure(t *testing.T) {
	// Test that CostNode can be created with valid fields
	t.Run("valid product node", func(t *testing.T) {
		node := CostNode{
			ID:         uuid.New(),
			Name:       "test-product",
			Type:       string(NodeTypeProduct),
			CostLabels: map[string]interface{}{"env": "prod"},
			IsPlatform: false,
			Metadata:   map[string]interface{}{"description": "Test product"},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "test-product", node.Name)
		assert.Equal(t, string(NodeTypeProduct), node.Type)
	})

	t.Run("platform node", func(t *testing.T) {
		node := CostNode{
			ID:         uuid.New(),
			Name:       "kubernetes",
			Type:       string(NodeTypePlatform),
			IsPlatform: true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		assert.True(t, node.IsPlatform)
		assert.Equal(t, string(NodeTypePlatform), node.Type)
	})
}

func TestDependencyEdge_Structure(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()

	t.Run("valid edge", func(t *testing.T) {
		edge := DependencyEdge{
			ID:              uuid.New(),
			ParentID:        parentID,
			ChildID:         childID,
			DefaultStrategy: string(StrategyProportionalOn),
			DefaultParameters: map[string]interface{}{
				"dimension": "instance_hours",
			},
			ActiveFrom: time.Now().AddDate(0, 0, -1),
			ActiveTo:   nil,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		assert.NotEmpty(t, edge.ID)
		assert.Equal(t, parentID, edge.ParentID)
		assert.Equal(t, childID, edge.ChildID)
		assert.Equal(t, string(StrategyProportionalOn), edge.DefaultStrategy)
	})

	t.Run("edge with fixed percent strategy", func(t *testing.T) {
		edge := DependencyEdge{
			ID:              uuid.New(),
			ParentID:        parentID,
			ChildID:         childID,
			DefaultStrategy: string(StrategyFixedPercent),
			DefaultParameters: map[string]interface{}{
				"percent": 0.25,
			},
			ActiveFrom: time.Now(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		assert.Equal(t, string(StrategyFixedPercent), edge.DefaultStrategy)
		assert.Equal(t, 0.25, edge.DefaultParameters["percent"])
	})
}

func TestNodeCostByDimension_Structure(t *testing.T) {
	cost := NodeCostByDimension{
		NodeID:    uuid.New(),
		CostDate:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Dimension: "instance_hours",
		Amount:    decimal.NewFromFloat(100.50),
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("amount conversion", func(t *testing.T) {
		amount, exact := cost.Amount.Float64()
		assert.True(t, exact)
		assert.Equal(t, 100.50, amount)
	})

	t.Run("structure fields", func(t *testing.T) {
		assert.NotEmpty(t, cost.NodeID)
		assert.Equal(t, "instance_hours", cost.Dimension)
		assert.Equal(t, "USD", cost.Currency)
	})
}

func TestNodeUsageByDimension_Structure(t *testing.T) {
	t.Run("valid usage metric", func(t *testing.T) {
		usage := NodeUsageByDimension{
			NodeID:    uuid.New(),
			UsageDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Metric:    "instance_hours",
			Value:     decimal.NewFromFloat(24.0),
			Unit:      "hours",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		assert.NotEmpty(t, usage.NodeID)
		assert.Equal(t, "instance_hours", usage.Metric)
		assert.Equal(t, "hours", usage.Unit)
		value, _ := usage.Value.Float64()
		assert.Equal(t, 24.0, value)
	})
}

func TestAllocationStrategy_Validation(t *testing.T) {
	validStrategies := []AllocationStrategy{
		StrategyEqual,
		StrategyProportionalOn,
		StrategyFixedPercent,
		StrategyCappedProp,
		StrategyResidualToMax,
		StrategyWeightedAverage,
		StrategyHybridFixedProp,
		StrategyMinFloorProportional,
		StrategySegmentFilteredProp,
	}

	for _, strategy := range validStrategies {
		t.Run(string(strategy), func(t *testing.T) {
			assert.True(t, IsValidStrategy(string(strategy)))
		})
	}

	t.Run("invalid strategy", func(t *testing.T) {
		assert.False(t, IsValidStrategy("invalid_strategy"))
	})
}

func TestNodeType_Constants(t *testing.T) {
	// Test that node type constants are defined correctly
	t.Run("product type", func(t *testing.T) {
		assert.Equal(t, NodeType("product"), NodeTypeProduct)
	})

	t.Run("platform type", func(t *testing.T) {
		assert.Equal(t, NodeType("platform"), NodeTypePlatform)
	})

	t.Run("shared type", func(t *testing.T) {
		assert.Equal(t, NodeType("shared"), NodeTypeShared)
	})

	t.Run("resource type", func(t *testing.T) {
		assert.Equal(t, NodeType("resource"), NodeTypeResource)
	})
}

func TestComputationRun_Structure(t *testing.T) {
	run := ComputationRun{
		ID:          uuid.New(),
		Status:      "pending",
		WindowStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		WindowEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		GraphHash:   "abc123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("structure fields", func(t *testing.T) {
		assert.NotEmpty(t, run.ID)
		assert.Equal(t, "pending", run.Status)
		assert.Equal(t, "abc123", run.GraphHash)
		assert.True(t, run.WindowEnd.After(run.WindowStart))
	})
}
