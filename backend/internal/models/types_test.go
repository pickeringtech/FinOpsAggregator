package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCostNode_Validation(t *testing.T) {
	tests := []struct {
		name    string
		node    CostNode
		wantErr bool
	}{
		{
			name: "valid product node",
			node: CostNode{
				ID:          uuid.New(),
				Name:        "test-product",
				NodeType:    NodeTypeProduct,
				Description: "Test product",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty name should be invalid",
			node: CostNode{
				ID:       uuid.New(),
				Name:     "",
				NodeType: NodeTypeProduct,
			},
			wantErr: true,
		},
		{
			name: "invalid node type",
			node: CostNode{
				ID:       uuid.New(),
				Name:     "test",
				NodeType: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependencyEdge_Validation(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()

	tests := []struct {
		name    string
		edge    DependencyEdge
		wantErr bool
	}{
		{
			name: "valid edge",
			edge: DependencyEdge{
				ID:       uuid.New(),
				ParentID: parentID,
				ChildID:  childID,
				Strategy: StrategyProportionalOn,
				StrategyParams: map[string]interface{}{
					"dimension": "instance_hours",
				},
				EffectiveFrom: time.Now().AddDate(0, 0, -1),
				EffectiveTo:   nil,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "self-referencing edge should be invalid",
			edge: DependencyEdge{
				ID:       uuid.New(),
				ParentID: parentID,
				ChildID:  parentID, // Same as parent
				Strategy: StrategyEqual,
			},
			wantErr: true,
		},
		{
			name: "invalid strategy",
			edge: DependencyEdge{
				ID:       uuid.New(),
				ParentID: parentID,
				ChildID:  childID,
				Strategy: "invalid_strategy",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.edge.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNodeCostByDimension_Calculations(t *testing.T) {
	cost := NodeCostByDimension{
		ID:        uuid.New(),
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

	t.Run("validation", func(t *testing.T) {
		err := cost.Validate()
		assert.NoError(t, err)
	})

	t.Run("negative amount should be invalid", func(t *testing.T) {
		invalidCost := cost
		invalidCost.Amount = decimal.NewFromFloat(-10.0)
		err := invalidCost.Validate()
		assert.Error(t, err)
	})
}

func TestUsageMetric_Validation(t *testing.T) {
	tests := []struct {
		name    string
		metric  UsageMetric
		wantErr bool
	}{
		{
			name: "valid usage metric",
			metric: UsageMetric{
				ID:        uuid.New(),
				NodeID:    uuid.New(),
				UsageDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Dimension: "instance_hours",
				Value:     decimal.NewFromFloat(24.0),
				Unit:      "hours",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "negative value should be invalid",
			metric: UsageMetric{
				ID:        uuid.New(),
				NodeID:    uuid.New(),
				UsageDate: time.Now(),
				Dimension: "instance_hours",
				Value:     decimal.NewFromFloat(-5.0),
				Unit:      "hours",
			},
			wantErr: true,
		},
		{
			name: "empty dimension should be invalid",
			metric: UsageMetric{
				ID:        uuid.New(),
				NodeID:    uuid.New(),
				UsageDate: time.Now(),
				Dimension: "",
				Value:     decimal.NewFromFloat(10.0),
				Unit:      "hours",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAllocationStrategy_Validation(t *testing.T) {
	validStrategies := []AllocationStrategy{
		StrategyEqual,
		StrategyProportionalOn,
		StrategyFixedPercent,
		StrategyCappedProportional,
		StrategyResidualToMax,
	}

	for _, strategy := range validStrategies {
		t.Run(string(strategy), func(t *testing.T) {
			assert.True(t, strategy.IsValid())
		})
	}

	t.Run("invalid strategy", func(t *testing.T) {
		invalid := AllocationStrategy("invalid_strategy")
		assert.False(t, invalid.IsValid())
	})
}

func TestNodeType_Validation(t *testing.T) {
	validTypes := []NodeType{
		NodeTypeProduct,
		NodeTypeSharedResource,
		NodeTypePlatform,
		NodeTypeDirectResource,
	}

	for _, nodeType := range validTypes {
		t.Run(string(nodeType), func(t *testing.T) {
			assert.True(t, nodeType.IsValid())
		})
	}

	t.Run("invalid node type", func(t *testing.T) {
		invalid := NodeType("invalid_type")
		assert.False(t, invalid.IsValid())
	})
}

func TestComputationRun_StatusTransitions(t *testing.T) {
	run := ComputationRun{
		ID:        uuid.New(),
		Status:    RunStatusPending,
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("valid status transitions", func(t *testing.T) {
		validTransitions := map[RunStatus][]RunStatus{
			RunStatusPending:   {RunStatusRunning, RunStatusCancelled},
			RunStatusRunning:   {RunStatusCompleted, RunStatusFailed, RunStatusCancelled},
			RunStatusCompleted: {},
			RunStatusFailed:    {RunStatusPending}, // Can retry
			RunStatusCancelled: {RunStatusPending}, // Can restart
		}

		for fromStatus, toStatuses := range validTransitions {
			for _, toStatus := range toStatuses {
				t.Run(fmt.Sprintf("%s_to_%s", fromStatus, toStatus), func(t *testing.T) {
					testRun := run
					testRun.Status = fromStatus
					assert.True(t, testRun.CanTransitionTo(toStatus))
				})
			}
		}
	})

	t.Run("invalid status transitions", func(t *testing.T) {
		// Completed runs cannot transition to running
		completedRun := run
		completedRun.Status = RunStatusCompleted
		assert.False(t, completedRun.CanTransitionTo(RunStatusRunning))
	})
}
