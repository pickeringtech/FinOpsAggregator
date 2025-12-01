package allocate

import (
	"context"
	"testing"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/config"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndAllocationCoverage validates that the fixed dependency relationships
// result in proper allocation coverage
func TestEndToEndAllocationCoverage(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires a real database connection
	// We'll use the same database as the running system
	cfg := config.PostgresConfig{
		DSN: "postgresql://finops:finops@postgres:5432/finops?sslmode=disable",
	}

	db, err := store.NewDB(cfg)
	require.NoError(t, err, "Failed to connect to database")
	defer db.Close()

	st := store.NewStore(db)
	ctx := context.Background()

	t.Run("Dependency relationships are correct", func(t *testing.T) {
		// Query dependency edges to verify they're correct
		edges, err := st.Edges.GetActiveEdgesForDate(ctx, time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		// Count edges by parent-child type combinations
		edgeCounts := make(map[string]int)
		for _, edge := range edges {
			parentNode, err := st.Nodes.GetByID(ctx, edge.ParentID)
			require.NoError(t, err)

			childNode, err := st.Nodes.GetByID(ctx, edge.ChildID)
			require.NoError(t, err)

			key := parentNode.Type + " → " + childNode.Type
			edgeCounts[key]++
		}

		// Verify correct relationships exist
		assert.Greater(t, edgeCounts["resource → product"], 0, "Should have resource → product edges")
		assert.Greater(t, edgeCounts["shared → product"], 0, "Should have shared → product edges")
		assert.Greater(t, edgeCounts["platform → product"], 0, "Should have platform → product edges")

		// Verify incorrect relationships don't exist
		assert.Equal(t, 0, edgeCounts["product → resource"], "Should NOT have product → resource edges")
		assert.Equal(t, 0, edgeCounts["product → shared"], "Should NOT have product → shared edges")
		assert.Equal(t, 0, edgeCounts["product → platform"], "Should NOT have product → platform edges")

		t.Logf("Edge counts: %+v", edgeCounts)
	})

	t.Run("Raw costs exist in infrastructure nodes", func(t *testing.T) {
		startDate := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC)

		// Get all nodes
		nodes, err := st.Nodes.List(ctx, store.NodeFilters{IncludeArchived: false})
		require.NoError(t, err)

		totalRawCosts := make(map[string]float64)

		for _, node := range nodes {
			costs, err := st.Costs.GetByNodeAndDateRange(ctx, node.ID, startDate, endDate, nil)
			require.NoError(t, err)

			var nodeTotalCost float64
			for _, cost := range costs {
				costFloat, _ := cost.Amount.Float64()
				nodeTotalCost += costFloat
			}

			if nodeTotalCost > 0 {
				totalRawCosts[node.Type] += nodeTotalCost
			}
		}

		t.Logf("Raw costs by node type: %+v", totalRawCosts)

		// Infrastructure should have the majority of costs
		assert.Greater(t, totalRawCosts["resource"], 0.0, "Resources should have costs")
		assert.Greater(t, totalRawCosts["shared"], 0.0, "Shared services should have costs")
		assert.Greater(t, totalRawCosts["platform"], 0.0, "Platform services should have costs")
	})
}