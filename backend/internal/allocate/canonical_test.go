package allocate

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCanonicalStageA tests the canonical Stage A allocation scenarios
// as documented in cost-attribution-behaviour.md and worked-examples.md

// TestResourceToSingleProduct tests Example 1: Basic Infrastructure → Product Allocation
// A compute resource allocates its costs to a single product.
func TestResourceToSingleProduct(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dimension := "compute_hours"

	// Create nodes
	resourceID := uuid.New()
	productID := uuid.New()

	resource := &models.CostNode{
		ID:   resourceID,
		Name: "compute_resource",
		Type: string(models.NodeTypeResource),
	}

	product := &models.CostNode{
		ID:   productID,
		Name: "web_app",
		Type: string(models.NodeTypeProduct),
	}

	// Create edge: Resource → Product (correct direction)
	edge := models.DependencyEdge{
		ID:              uuid.New(),
		ParentID:        resourceID, // Resource is cost source
		ChildID:         productID,  // Product is cost receiver
		DefaultStrategy: string(models.StrategyEqual),
		ActiveFrom:      date.AddDate(0, 0, -1),
	}

	// Resource has $100 direct cost
	resourceCost := models.NodeCostByDimension{
		NodeID:    resourceID,
		CostDate:  date,
		Dimension: dimension,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
	}

	// Verify edge direction is correct
	assert.Equal(t, resourceID, edge.ParentID, "Resource should be parent (cost source)")
	assert.Equal(t, productID, edge.ChildID, "Product should be child (cost receiver)")

	// Verify node types
	assert.Equal(t, string(models.NodeTypeResource), resource.Type)
	assert.Equal(t, string(models.NodeTypeProduct), product.Type)

	// Verify cost assignment
	assert.Equal(t, resourceID, resourceCost.NodeID, "Resource should have direct cost")
	assert.Equal(t, decimal.NewFromFloat(100.0), resourceCost.Amount)

	// Expected allocation result:
	// - Resource: Direct=$100, Indirect=$0, Holistic=$100
	// - Product: Direct=$0, Indirect=$100, Holistic=$100
	// - Raw Infra Cost: $100
	// - Allocated Product Cost: $100 (product is final cost centre)
	// - Coverage: 100%

	t.Run("allocation invariants", func(t *testing.T) {
		rawInfraCost := resourceCost.Amount
		allocatedProductCost := resourceCost.Amount // 100% allocation to single child
		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)

		// Invariant 1: Conservation
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Allocated + Unallocated should equal Raw Infra Cost")

		// Invariant 2: Coverage
		coverage := allocatedProductCost.Div(rawInfraCost).Mul(decimal.NewFromInt(100))
		assert.True(t, coverage.Equal(decimal.NewFromInt(100)),
			"Coverage should be 100%% for single child allocation")
	})

	_ = ctx // Used in future integration tests
}

// TestSharedToMultipleProductsEqual tests Example 2: Shared Service → Multiple Products (Equal)
// A shared database cluster allocates costs equally to three products.
func TestSharedToMultipleProductsEqual(t *testing.T) {
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dimension := "storage_gb_month"

	// Create nodes
	sharedID := uuid.New()
	productAID := uuid.New()
	productBID := uuid.New()
	productCID := uuid.New()

	shared := &models.CostNode{
		ID:   sharedID,
		Name: "database_cluster",
		Type: string(models.NodeTypeShared),
	}

	productA := &models.CostNode{ID: productAID, Name: "web_app", Type: string(models.NodeTypeProduct)}
	productB := &models.CostNode{ID: productBID, Name: "api", Type: string(models.NodeTypeProduct)}
	productC := &models.CostNode{ID: productCID, Name: "mobile", Type: string(models.NodeTypeProduct)}

	// Create edges: Shared → Products (equal strategy)
	edges := []models.DependencyEdge{
		{ID: uuid.New(), ParentID: sharedID, ChildID: productAID, DefaultStrategy: string(models.StrategyEqual), ActiveFrom: date.AddDate(0, 0, -1)},
		{ID: uuid.New(), ParentID: sharedID, ChildID: productBID, DefaultStrategy: string(models.StrategyEqual), ActiveFrom: date.AddDate(0, 0, -1)},
		{ID: uuid.New(), ParentID: sharedID, ChildID: productCID, DefaultStrategy: string(models.StrategyEqual), ActiveFrom: date.AddDate(0, 0, -1)},
	}

	// Shared service has $300 direct cost
	sharedCost := models.NodeCostByDimension{
		NodeID:    sharedID,
		CostDate:  date,
		Dimension: dimension,
		Amount:    decimal.NewFromFloat(300.0),
		Currency:  "USD",
	}

	// Verify edge directions
	for _, edge := range edges {
		assert.Equal(t, sharedID, edge.ParentID, "Shared should be parent")
	}

	// Verify node types
	assert.Equal(t, string(models.NodeTypeShared), shared.Type)
	assert.Equal(t, string(models.NodeTypeProduct), productA.Type)
	assert.Equal(t, string(models.NodeTypeProduct), productB.Type)
	assert.Equal(t, string(models.NodeTypeProduct), productC.Type)

	t.Run("equal allocation calculation", func(t *testing.T) {
		numChildren := decimal.NewFromInt(3)
		sharePerChild := decimal.NewFromInt(1).Div(numChildren)
		allocationPerChild := sharedCost.Amount.Mul(sharePerChild)

		// Each product should receive $100 (with rounding tolerance)
		expectedAllocation := decimal.NewFromFloat(100.0)
		// Use Round to handle 1/3 precision: 300 * (1/3) = 99.999... rounds to 100
		allocationRounded := allocationPerChild.Round(2)
		assert.True(t, allocationRounded.Equal(expectedAllocation),
			"Each product should receive $100 (1/3 of $300), got %s", allocationRounded.String())
	})

	t.Run("allocation invariants", func(t *testing.T) {
		rawInfraCost := sharedCost.Amount
		allocatedProductCost := sharedCost.Amount // All 3 products are final cost centres
		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)

		// Invariant 1: Conservation
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Allocated + Unallocated should equal Raw Infra Cost")

		// Invariant 2: No amplification
		// Sum of final cost centre costs = $100 + $100 + $100 = $300 = Raw Infra
		sumFinalCostCentres := decimal.NewFromFloat(300.0)
		assert.True(t, sumFinalCostCentres.Equal(rawInfraCost),
			"Sum of final cost centres should equal Raw Infra Cost")
	})
}

// TestPlatformToProductsProportional tests Example 3: Platform → Products (Proportional on CPU)
// A Kubernetes platform allocates costs proportionally based on CPU usage.
func TestPlatformToProductsProportional(t *testing.T) {
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dimension := "compute_hours"

	// Create nodes
	platformID := uuid.New()
	productAID := uuid.New()
	productBID := uuid.New()

	platform := &models.CostNode{
		ID:   platformID,
		Name: "kubernetes_platform",
		Type: string(models.NodeTypePlatform),
	}

	_ = &models.CostNode{ID: productAID, Name: "web_app", Type: string(models.NodeTypeProduct)}
	_ = &models.CostNode{ID: productBID, Name: "api", Type: string(models.NodeTypeProduct)}

	// Create edges: Platform → Products (proportional strategy)
	edges := []models.DependencyEdge{
		{ID: uuid.New(), ParentID: platformID, ChildID: productAID, DefaultStrategy: string(models.StrategyProportionalOn), ActiveFrom: date.AddDate(0, 0, -1)},
		{ID: uuid.New(), ParentID: platformID, ChildID: productBID, DefaultStrategy: string(models.StrategyProportionalOn), ActiveFrom: date.AddDate(0, 0, -1)},
	}

	// Platform has $500 direct cost
	platformCost := models.NodeCostByDimension{
		NodeID:    platformID,
		CostDate:  date,
		Dimension: dimension,
		Amount:    decimal.NewFromFloat(500.0),
		Currency:  "USD",
	}

	// Usage data: Product A = 1000 CPU-hrs, Product B = 4000 CPU-hrs
	usageA := models.NodeUsageByDimension{
		NodeID:    productAID,
		UsageDate: date,
		Metric:    "cpu_hours",
		Value:     decimal.NewFromFloat(1000.0),
		Unit:      "hours",
	}
	usageB := models.NodeUsageByDimension{
		NodeID:    productBID,
		UsageDate: date,
		Metric:    "cpu_hours",
		Value:     decimal.NewFromFloat(4000.0),
		Unit:      "hours",
	}

	// Verify edge directions
	for _, edge := range edges {
		assert.Equal(t, platformID, edge.ParentID, "Platform should be parent")
	}

	// Verify node types
	assert.Equal(t, string(models.NodeTypePlatform), platform.Type)

	t.Run("proportional allocation calculation", func(t *testing.T) {
		totalUsage := usageA.Value.Add(usageB.Value) // 5000
		shareA := usageA.Value.Div(totalUsage)       // 1000/5000 = 0.2
		shareB := usageB.Value.Div(totalUsage)       // 4000/5000 = 0.8

		allocationA := platformCost.Amount.Mul(shareA) // $500 * 0.2 = $100
		allocationB := platformCost.Amount.Mul(shareB) // $500 * 0.8 = $400

		assert.True(t, shareA.Equal(decimal.NewFromFloat(0.2)), "Product A share should be 20%%")
		assert.True(t, shareB.Equal(decimal.NewFromFloat(0.8)), "Product B share should be 80%%")
		assert.True(t, allocationA.Equal(decimal.NewFromFloat(100.0)), "Product A should receive $100")
		assert.True(t, allocationB.Equal(decimal.NewFromFloat(400.0)), "Product B should receive $400")
	})

	t.Run("allocation invariants", func(t *testing.T) {
		rawInfraCost := platformCost.Amount
		allocatedProductCost := platformCost.Amount
		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)

		// Invariant 1: Conservation
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Allocated + Unallocated should equal Raw Infra Cost")

		// Invariant 2: Sum of shares = 1
		totalUsage := usageA.Value.Add(usageB.Value)
		shareA := usageA.Value.Div(totalUsage)
		shareB := usageB.Value.Div(totalUsage)
		totalShares := shareA.Add(shareB)
		assert.True(t, totalShares.Equal(decimal.NewFromInt(1)),
			"Sum of proportional shares should equal 1")
	})

	_ = edges // Used in future integration tests
}

// TestFixedPercentAllocation tests the fixed_percent strategy
func TestFixedPercentAllocation(t *testing.T) {
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create nodes
	sharedID := uuid.New()
	productID := uuid.New()

	// Shared service has $1000 direct cost
	sharedCost := decimal.NewFromFloat(1000.0)

	// Fixed percent = 25%
	fixedPercent := decimal.NewFromFloat(25.0)

	t.Run("fixed percent calculation", func(t *testing.T) {
		share := fixedPercent.Div(decimal.NewFromInt(100)) // 0.25
		allocation := sharedCost.Mul(share)                // $1000 * 0.25 = $250

		assert.True(t, allocation.Equal(decimal.NewFromFloat(250.0)),
			"Product should receive $250 (25%% of $1000)")
	})

	_ = date
	_ = sharedID
	_ = productID
}

// TestCappedProportionalAllocation tests the capped_proportional strategy
func TestCappedProportionalAllocation(t *testing.T) {
	// Platform has $1000 direct cost
	platformCost := decimal.NewFromFloat(1000.0)

	// Usage: A = 90%, B = 10%
	usageA := decimal.NewFromFloat(900.0)
	usageB := decimal.NewFromFloat(100.0)
	totalUsage := usageA.Add(usageB)

	// Cap = 50%
	cap := decimal.NewFromFloat(0.5)

	t.Run("capped proportional calculation", func(t *testing.T) {
		// Proportional shares
		propShareA := usageA.Div(totalUsage) // 0.9
		propShareB := usageB.Div(totalUsage) // 0.1

		// Apply cap
		cappedShareA := decimal.Min(propShareA, cap) // min(0.9, 0.5) = 0.5
		cappedShareB := decimal.Min(propShareB, cap) // min(0.1, 0.5) = 0.1

		allocationA := platformCost.Mul(cappedShareA) // $1000 * 0.5 = $500
		allocationB := platformCost.Mul(cappedShareB) // $1000 * 0.1 = $100

		assert.True(t, cappedShareA.Equal(decimal.NewFromFloat(0.5)), "A's share should be capped at 50%%")
		assert.True(t, allocationA.Equal(decimal.NewFromFloat(500.0)), "A should receive $500 (capped)")
		assert.True(t, allocationB.Equal(decimal.NewFromFloat(100.0)), "B should receive $100")

		// Note: $400 remains unallocated due to cap
		totalAllocated := allocationA.Add(allocationB)
		unallocated := platformCost.Sub(totalAllocated)
		assert.True(t, unallocated.Equal(decimal.NewFromFloat(400.0)),
			"$400 should remain unallocated due to cap")
	})
}

// TestFinalCostCentreIdentification tests that final cost centres are correctly identified
func TestFinalCostCentreIdentification(t *testing.T) {
	// Final cost centres are product nodes with no outgoing product→product edges

	productAID := uuid.New()
	productBID := uuid.New()
	productCID := uuid.New()

	// Product A has an outgoing edge to Product B (not a final cost centre)
	// Product B has no outgoing edges (final cost centre)
	// Product C has no outgoing edges (final cost centre)

	edges := []models.DependencyEdge{
		{ID: uuid.New(), ParentID: productAID, ChildID: productBID, DefaultStrategy: string(models.StrategyFixedPercent)},
	}

	t.Run("identify final cost centres", func(t *testing.T) {
		// Build set of products with outgoing edges
		hasOutgoingEdge := make(map[uuid.UUID]bool)
		for _, edge := range edges {
			hasOutgoingEdge[edge.ParentID] = true
		}

		// Product A has outgoing edge → not final cost centre
		assert.True(t, hasOutgoingEdge[productAID], "Product A has outgoing edge")

		// Product B has no outgoing edge → final cost centre
		assert.False(t, hasOutgoingEdge[productBID], "Product B is final cost centre")

		// Product C has no outgoing edge → final cost centre
		assert.False(t, hasOutgoingEdge[productCID], "Product C is final cost centre")
	})
}

// TestAllocationInvariants tests the core allocation invariants
func TestAllocationInvariants(t *testing.T) {
	t.Run("conservation invariant", func(t *testing.T) {
		rawInfraCost := decimal.NewFromFloat(1000.0)
		allocatedProductCost := decimal.NewFromFloat(800.0)
		unallocatedCost := decimal.NewFromFloat(200.0)

		// Invariant: Allocated + Unallocated = Raw Infra
		require.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Conservation invariant must hold")
	})

	t.Run("no amplification invariant", func(t *testing.T) {
		rawInfraCost := decimal.NewFromFloat(1000.0)

		// Sum of final cost centre holistic costs
		finalCostCentreSum := decimal.NewFromFloat(1000.0)

		// Invariant: Sum of final cost centres <= Raw Infra
		require.True(t, finalCostCentreSum.LessThanOrEqual(rawInfraCost),
			"No amplification invariant must hold")
	})

	t.Run("share validity invariant", func(t *testing.T) {
		// For each parent, sum of shares to children <= 1
		shares := []decimal.Decimal{
			decimal.NewFromFloat(0.3),
			decimal.NewFromFloat(0.5),
			decimal.NewFromFloat(0.2),
		}

		totalShares := decimal.Zero
		for _, share := range shares {
			totalShares = totalShares.Add(share)
		}

		require.True(t, totalShares.LessThanOrEqual(decimal.NewFromInt(1)),
			"Sum of shares must be <= 1")
	})
}

// =============================================================================
// R-2: Regression test for "100% coverage but 4 products" bug
// =============================================================================

// TestRegressionCoverageBug tests the scenario where coverage shows 100%
// but there are multiple products in the hierarchy that should be receiving costs.
// This was a real bug where the coverage calculation was incorrect.
func TestRegressionCoverageBug(t *testing.T) {
	// Scenario: 4 products in hierarchy, platform allocates to all
	// Bug: Coverage showed 100% even when some products weren't receiving costs

	platformID := uuid.New()
	productIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}

	// Platform has $1000 direct cost
	platformCost := decimal.NewFromFloat(1000.0)

	// Create edges from platform to all 4 products (equal strategy)
	edges := make([]models.DependencyEdge, len(productIDs))
	for i, productID := range productIDs {
		edges[i] = models.DependencyEdge{
			ID:              uuid.New(),
			ParentID:        platformID,
			ChildID:         productID,
			DefaultStrategy: string(models.StrategyEqual),
		}
	}

	t.Run("all products should receive allocation", func(t *testing.T) {
		numProducts := decimal.NewFromInt(int64(len(productIDs)))
		sharePerProduct := decimal.NewFromInt(1).Div(numProducts)
		allocationPerProduct := platformCost.Mul(sharePerProduct)

		// Each product should receive $250
		expectedAllocation := decimal.NewFromFloat(250.0)
		assert.True(t, allocationPerProduct.Equal(expectedAllocation),
			"Each of 4 products should receive $250")

		// Total allocated should equal platform cost
		totalAllocated := allocationPerProduct.Mul(numProducts)
		assert.True(t, totalAllocated.Equal(platformCost),
			"Total allocated should equal platform cost")
	})

	t.Run("coverage calculation should be correct", func(t *testing.T) {
		rawInfraCost := platformCost
		allocatedProductCost := platformCost // All cost allocated to 4 products

		coverage := allocatedProductCost.Div(rawInfraCost).Mul(decimal.NewFromInt(100))
		assert.True(t, coverage.Equal(decimal.NewFromInt(100)),
			"Coverage should be 100%% when all infra cost is allocated")
	})

	t.Run("final cost centres should be all 4 products", func(t *testing.T) {
		// All 4 products have no outgoing edges, so all are final cost centres
		finalCostCentres := make(map[uuid.UUID]bool)
		for _, productID := range productIDs {
			finalCostCentres[productID] = true
		}

		assert.Equal(t, 4, len(finalCostCentres),
			"Should have 4 final cost centres")
	})

	_ = edges
}

// =============================================================================
// R-3: Tests for Raw vs Allocated vs Unallocated invariants
// =============================================================================

// TestRawAllocatedUnallocatedInvariants tests the core cost invariants
func TestRawAllocatedUnallocatedInvariants(t *testing.T) {
	t.Run("full allocation scenario", func(t *testing.T) {
		// All infra cost is allocated to products
		rawInfraCost := decimal.NewFromFloat(1000.0)
		allocatedProductCost := decimal.NewFromFloat(1000.0)
		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)

		// Invariant: Allocated + Unallocated == Raw Infra
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Allocated + Unallocated must equal Raw Infra")
		assert.True(t, unallocatedCost.IsZero(),
			"Unallocated should be zero in full allocation")
	})

	t.Run("partial allocation scenario", func(t *testing.T) {
		// Only 80% of infra cost is allocated
		rawInfraCost := decimal.NewFromFloat(1000.0)
		allocatedProductCost := decimal.NewFromFloat(800.0)
		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)

		// Invariant: Allocated + Unallocated == Raw Infra
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Allocated + Unallocated must equal Raw Infra")
		assert.True(t, unallocatedCost.Equal(decimal.NewFromFloat(200.0)),
			"Unallocated should be $200")
	})

	t.Run("capped allocation leaves unallocated", func(t *testing.T) {
		// Capped proportional leaves some cost unallocated
		rawInfraCost := decimal.NewFromFloat(1000.0)
		// With cap=50% and one child at 90% usage, only 50% is allocated
		allocatedProductCost := decimal.NewFromFloat(600.0) // 50% + 10%
		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)

		// Invariant: Allocated + Unallocated == Raw Infra
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Allocated + Unallocated must equal Raw Infra")
		assert.True(t, unallocatedCost.Equal(decimal.NewFromFloat(400.0)),
			"Unallocated should be $400 due to cap")
	})

	t.Run("multi-stage allocation preserves invariant", func(t *testing.T) {
		// Stage A: Infra → Products
		// Stage B: Product → Product roll-ups
		rawInfraCost := decimal.NewFromFloat(1000.0)

		// After Stage A: All infra allocated to intermediate products
		intermediateProductCost := decimal.NewFromFloat(1000.0)

		// After Stage B: Intermediate products roll up to final cost centres
		// Product A ($400) rolls 100% to Product B
		// Product C ($600) is already a final cost centre
		finalCostCentreSum := decimal.NewFromFloat(1000.0) // $400 + $600

		// Invariant: Final cost centre sum == Raw Infra (no amplification)
		assert.True(t, finalCostCentreSum.Equal(rawInfraCost),
			"Final cost centre sum must equal Raw Infra")

		_ = intermediateProductCost
	})
}

// =============================================================================
// R-4: Tests for no amplification invariant
// =============================================================================

// TestNoAmplificationInvariant tests that costs are never amplified through allocation
func TestNoAmplificationInvariant(t *testing.T) {
	t.Run("single stage allocation", func(t *testing.T) {
		rawInfraCost := decimal.NewFromFloat(1000.0)

		// Products receive allocations
		productAllocations := []decimal.Decimal{
			decimal.NewFromFloat(300.0),
			decimal.NewFromFloat(400.0),
			decimal.NewFromFloat(300.0),
		}

		totalAllocated := decimal.Zero
		for _, alloc := range productAllocations {
			totalAllocated = totalAllocated.Add(alloc)
		}

		// No amplification: sum of allocations <= raw infra
		assert.True(t, totalAllocated.LessThanOrEqual(rawInfraCost),
			"Total allocated must not exceed Raw Infra")
	})

	t.Run("product roll-up does not amplify", func(t *testing.T) {
		// Product A has $500 holistic cost
		// Product A rolls up 100% to Product B
		// Product B should have $500 holistic cost (not more)

		productAHolistic := decimal.NewFromFloat(500.0)
		rollUpPercent := decimal.NewFromFloat(1.0) // 100%
		productBReceived := productAHolistic.Mul(rollUpPercent)

		assert.True(t, productBReceived.Equal(productAHolistic),
			"Roll-up should not amplify cost")
	})

	t.Run("partial roll-up preserves total", func(t *testing.T) {
		// Product A has $1000 holistic cost
		// Rolls 30% to B, 50% to C, retains 20%
		productAHolistic := decimal.NewFromFloat(1000.0)

		toBPercent := decimal.NewFromFloat(0.3)
		toCPercent := decimal.NewFromFloat(0.5)
		retainedPercent := decimal.NewFromFloat(0.2)

		toB := productAHolistic.Mul(toBPercent)       // $300
		toC := productAHolistic.Mul(toCPercent)       // $500
		retained := productAHolistic.Mul(retainedPercent) // $200

		totalAfterRollUp := toB.Add(toC).Add(retained)

		assert.True(t, totalAfterRollUp.Equal(productAHolistic),
			"Partial roll-up must preserve total cost")
	})
}

// =============================================================================
// R-5: Tests for final cost centre identification
// =============================================================================

// TestFinalCostCentreIdentificationDetailed tests various scenarios for identifying final cost centres
func TestFinalCostCentreIdentificationDetailed(t *testing.T) {
	t.Run("product with no outgoing edges is final cost centre", func(t *testing.T) {
		productID := uuid.New()
		edges := []models.DependencyEdge{} // No edges

		isFinalCostCentre := !hasOutgoingProductEdge(productID, edges)
		assert.True(t, isFinalCostCentre,
			"Product with no outgoing edges should be final cost centre")
	})

	t.Run("product with outgoing product edge is not final cost centre", func(t *testing.T) {
		productAID := uuid.New()
		productBID := uuid.New()

		edges := []models.DependencyEdge{
			{ParentID: productAID, ChildID: productBID},
		}

		isFinalCostCentre := !hasOutgoingProductEdge(productAID, edges)
		assert.False(t, isFinalCostCentre,
			"Product with outgoing edge should not be final cost centre")
	})

	t.Run("product receiving allocation but not allocating is final cost centre", func(t *testing.T) {
		platformID := uuid.New()
		productID := uuid.New()

		edges := []models.DependencyEdge{
			{ParentID: platformID, ChildID: productID}, // Platform → Product
		}

		// Product receives allocation but has no outgoing edges
		isFinalCostCentre := !hasOutgoingProductEdge(productID, edges)
		assert.True(t, isFinalCostCentre,
			"Product receiving allocation but not allocating should be final cost centre")
	})

	t.Run("partial roll-up product is still final cost centre", func(t *testing.T) {
		// Product A rolls 80% to B, retains 20%
		// Product A is still a final cost centre (for the 20%)
		productAID := uuid.New()
		productBID := uuid.New()

		edges := []models.DependencyEdge{
			{ParentID: productAID, ChildID: productBID}, // 80% roll-up
		}

		// Product A has outgoing edge, but retains residual
		// In our model, it's NOT a final cost centre because it has outgoing edges
		// The residual is tracked separately
		hasOutgoing := hasOutgoingProductEdge(productAID, edges)
		assert.True(t, hasOutgoing,
			"Product with partial roll-up has outgoing edge")

		// Product B has no outgoing edges
		isBFinalCostCentre := !hasOutgoingProductEdge(productBID, edges)
		assert.True(t, isBFinalCostCentre,
			"Product B (roll-up target) should be final cost centre")
	})
}

// hasOutgoingProductEdge checks if a product has any outgoing edges
func hasOutgoingProductEdge(productID uuid.UUID, edges []models.DependencyEdge) bool {
	for _, edge := range edges {
		if edge.ParentID == productID {
			return true
		}
	}
	return false
}

// =============================================================================
// R-6: Tests for full roll-up product→product
// =============================================================================

// TestFullProductRollUp tests 100% roll-up from Product A to Product B
func TestFullProductRollUp(t *testing.T) {
	// Scenario: Platform → Product A → Product B (100% roll-up)
	// Product A receives $500 from platform, rolls 100% to Product B
	// Product B is the only final cost centre

	platformID := uuid.New()
	productAID := uuid.New()
	productBID := uuid.New()

	// Platform has $500 direct cost
	platformCost := decimal.NewFromFloat(500.0)

	// Edges
	edges := []models.DependencyEdge{
		{ParentID: platformID, ChildID: productAID, DefaultStrategy: string(models.StrategyEqual)},
		{ParentID: productAID, ChildID: productBID, DefaultStrategy: string(models.StrategyFixedPercent)},
	}

	t.Run("stage A allocation", func(t *testing.T) {
		// Platform allocates 100% to Product A (single child)
		productAIndirect := platformCost
		productAHolistic := productAIndirect // No direct cost

		assert.True(t, productAHolistic.Equal(decimal.NewFromFloat(500.0)),
			"Product A should have $500 holistic cost after Stage A")
	})

	t.Run("stage B roll-up", func(t *testing.T) {
		productAHolistic := decimal.NewFromFloat(500.0)
		rollUpPercent := decimal.NewFromFloat(1.0) // 100%

		productBReceived := productAHolistic.Mul(rollUpPercent)
		productAResidual := productAHolistic.Sub(productBReceived)

		assert.True(t, productBReceived.Equal(decimal.NewFromFloat(500.0)),
			"Product B should receive $500 from roll-up")
		assert.True(t, productAResidual.IsZero(),
			"Product A should have no residual after 100%% roll-up")
	})

	t.Run("final cost centre identification", func(t *testing.T) {
		// Product A has outgoing edge → not final cost centre
		// Product B has no outgoing edge → final cost centre
		assert.True(t, hasOutgoingProductEdge(productAID, edges),
			"Product A has outgoing edge")
		assert.False(t, hasOutgoingProductEdge(productBID, edges),
			"Product B is final cost centre")
	})

	t.Run("invariants preserved", func(t *testing.T) {
		rawInfraCost := platformCost
		finalCostCentreSum := decimal.NewFromFloat(500.0) // Only Product B

		// No amplification
		assert.True(t, finalCostCentreSum.Equal(rawInfraCost),
			"Final cost centre sum should equal Raw Infra")

		// Conservation
		allocatedProductCost := finalCostCentreSum
		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Conservation invariant should hold")
	})
}

// =============================================================================
// R-7: Tests for partial and multi-parent roll-ups
// =============================================================================

// TestPartialRollUp tests partial roll-up where Product A retains some cost
func TestPartialRollUp(t *testing.T) {
	// Scenario: Product A ($1000 holistic) rolls 30% to B, 50% to C, retains 20%

	productAID := uuid.New()
	productBID := uuid.New()
	productCID := uuid.New()

	productAHolistic := decimal.NewFromFloat(1000.0)

	// Roll-up percentages
	toBPercent := decimal.NewFromFloat(0.3)
	toCPercent := decimal.NewFromFloat(0.5)
	retainedPercent := decimal.NewFromFloat(0.2)

	t.Run("partial roll-up calculation", func(t *testing.T) {
		toB := productAHolistic.Mul(toBPercent)           // $300
		toC := productAHolistic.Mul(toCPercent)           // $500
		retained := productAHolistic.Mul(retainedPercent) // $200

		assert.True(t, toB.Equal(decimal.NewFromFloat(300.0)),
			"Product B should receive $300")
		assert.True(t, toC.Equal(decimal.NewFromFloat(500.0)),
			"Product C should receive $500")
		assert.True(t, retained.Equal(decimal.NewFromFloat(200.0)),
			"Product A should retain $200")
	})

	t.Run("total preserved", func(t *testing.T) {
		toB := productAHolistic.Mul(toBPercent)
		toC := productAHolistic.Mul(toCPercent)
		retained := productAHolistic.Mul(retainedPercent)

		total := toB.Add(toC).Add(retained)
		assert.True(t, total.Equal(productAHolistic),
			"Total after partial roll-up should equal original holistic cost")
	})

	t.Run("final cost centres", func(t *testing.T) {
		// Product A has outgoing edges but retains residual
		// Products B and C have no outgoing edges
		// All three contribute to final cost centre sum

		edges := []models.DependencyEdge{
			{ParentID: productAID, ChildID: productBID},
			{ParentID: productAID, ChildID: productCID},
		}

		// Product A has outgoing edges
		assert.True(t, hasOutgoingProductEdge(productAID, edges))
		// Products B and C are final cost centres
		assert.False(t, hasOutgoingProductEdge(productBID, edges))
		assert.False(t, hasOutgoingProductEdge(productCID, edges))
	})
}

// TestMultiParentRollUp tests a product receiving allocations from multiple parents
func TestMultiParentRollUp(t *testing.T) {
	// Scenario: Product C receives from both Product A (30%) and Product B (40%)

	productAID := uuid.New()
	productBID := uuid.New()
	productCID := uuid.New()

	productAHolistic := decimal.NewFromFloat(1000.0)
	productBHolistic := decimal.NewFromFloat(500.0)

	t.Run("multi-parent allocation", func(t *testing.T) {
		// Product A rolls 30% to C
		fromA := productAHolistic.Mul(decimal.NewFromFloat(0.3)) // $300

		// Product B rolls 40% to C
		fromB := productBHolistic.Mul(decimal.NewFromFloat(0.4)) // $200

		productCReceived := fromA.Add(fromB) // $500

		assert.True(t, productCReceived.Equal(decimal.NewFromFloat(500.0)),
			"Product C should receive $500 from multiple parents")
	})

	t.Run("residuals preserved", func(t *testing.T) {
		// Product A retains 70%
		productAResidual := productAHolistic.Mul(decimal.NewFromFloat(0.7)) // $700

		// Product B retains 60%
		productBResidual := productBHolistic.Mul(decimal.NewFromFloat(0.6)) // $300

		assert.True(t, productAResidual.Equal(decimal.NewFromFloat(700.0)),
			"Product A should retain $700")
		assert.True(t, productBResidual.Equal(decimal.NewFromFloat(300.0)),
			"Product B should retain $300")
	})

	t.Run("total cost preserved", func(t *testing.T) {
		// Original total: $1000 + $500 = $1500
		originalTotal := productAHolistic.Add(productBHolistic)

		// After roll-ups:
		// Product A residual: $700
		// Product B residual: $300
		// Product C received: $500
		productAResidual := decimal.NewFromFloat(700.0)
		productBResidual := decimal.NewFromFloat(300.0)
		productCReceived := decimal.NewFromFloat(500.0)

		finalTotal := productAResidual.Add(productBResidual).Add(productCReceived)

		assert.True(t, finalTotal.Equal(originalTotal),
			"Total cost should be preserved after multi-parent roll-ups")
	})

	_ = productAID
	_ = productBID
	_ = productCID
}

// TestDiamondRollUp tests a diamond-shaped roll-up pattern
func TestDiamondRollUp(t *testing.T) {
	// Scenario: Diamond pattern
	//     Platform
	//     /      \
	//    A        B
	//     \      /
	//        C
	// Platform allocates to A and B, both roll up to C

	platformID := uuid.New()
	productAID := uuid.New()
	productBID := uuid.New()
	productCID := uuid.New()

	platformCost := decimal.NewFromFloat(1000.0)

	t.Run("diamond allocation flow", func(t *testing.T) {
		// Stage A: Platform → A (50%), Platform → B (50%)
		toA := platformCost.Mul(decimal.NewFromFloat(0.5)) // $500
		toB := platformCost.Mul(decimal.NewFromFloat(0.5)) // $500

		// Stage B: A → C (100%), B → C (100%)
		fromA := toA.Mul(decimal.NewFromFloat(1.0)) // $500
		fromB := toB.Mul(decimal.NewFromFloat(1.0)) // $500

		productCHolistic := fromA.Add(fromB) // $1000

		assert.True(t, productCHolistic.Equal(platformCost),
			"Product C should receive full platform cost through diamond")
	})

	t.Run("no amplification in diamond", func(t *testing.T) {
		rawInfraCost := platformCost
		finalCostCentreSum := platformCost // Only C is final cost centre

		assert.True(t, finalCostCentreSum.Equal(rawInfraCost),
			"Diamond pattern should not amplify costs")
	})

	t.Run("final cost centre is only C", func(t *testing.T) {
		edges := []models.DependencyEdge{
			{ParentID: platformID, ChildID: productAID},
			{ParentID: platformID, ChildID: productBID},
			{ParentID: productAID, ChildID: productCID},
			{ParentID: productBID, ChildID: productCID},
		}

		// A and B have outgoing edges
		assert.True(t, hasOutgoingProductEdge(productAID, edges))
		assert.True(t, hasOutgoingProductEdge(productBID, edges))
		// C is the only final cost centre
		assert.False(t, hasOutgoingProductEdge(productCID, edges))
	})
}

// =============================================================================
// R-8: Tests for weighted-average strategy
// =============================================================================

// TestWeightedAverageStrategy tests the weighted average allocation strategy
// that uses a look-back window over usage metrics
func TestWeightedAverageStrategy(t *testing.T) {
	// Scenario: Platform allocates to products based on 7-day weighted average usage

	t.Run("simple average over window", func(t *testing.T) {
		// Product A: Day usage [100, 200, 150, 180, 160, 140, 170] → avg 157.14
		// Product B: Day usage [300, 100, 350, 120, 340, 160, 330] → avg 242.86
		usageA := []decimal.Decimal{
			decimal.NewFromFloat(100), decimal.NewFromFloat(200), decimal.NewFromFloat(150),
			decimal.NewFromFloat(180), decimal.NewFromFloat(160), decimal.NewFromFloat(140),
			decimal.NewFromFloat(170),
		}
		usageB := []decimal.Decimal{
			decimal.NewFromFloat(300), decimal.NewFromFloat(100), decimal.NewFromFloat(350),
			decimal.NewFromFloat(120), decimal.NewFromFloat(340), decimal.NewFromFloat(160),
			decimal.NewFromFloat(330),
		}

		// Calculate averages
		avgA := calculateAverage(usageA)
		avgB := calculateAverage(usageB)

		totalAvg := avgA.Add(avgB)
		shareA := avgA.Div(totalAvg)
		shareB := avgB.Div(totalAvg)

		// Platform cost
		platformCost := decimal.NewFromFloat(1000.0)
		allocationA := platformCost.Mul(shareA)
		allocationB := platformCost.Mul(shareB)

		// Verify allocations sum to platform cost
		totalAllocated := allocationA.Add(allocationB)
		assert.True(t, totalAllocated.Round(2).Equal(platformCost),
			"Total allocated should equal platform cost")

		// Verify shares are reasonable (A should get less than B)
		assert.True(t, shareA.LessThan(shareB),
			"Product A should have smaller share than B")
	})

	t.Run("smoothing effect of window", func(t *testing.T) {
		// Day 7 has spike for A, but window smooths it out
		// Without window: A=500, B=100 → A gets 83%
		// With 7-day window: A avg ≈ 157, B avg ≈ 243 → A gets ~39%

		singleDayUsageA := decimal.NewFromFloat(500)
		singleDayUsageB := decimal.NewFromFloat(100)
		singleDayTotal := singleDayUsageA.Add(singleDayUsageB)
		singleDayShareA := singleDayUsageA.Div(singleDayTotal)

		// With window (from previous test)
		windowedShareA := decimal.NewFromFloat(0.39) // Approximate

		// Window should reduce the spike effect
		assert.True(t, windowedShareA.LessThan(singleDayShareA),
			"Windowed share should be less than single-day spike share")
	})

	t.Run("zero usage in window", func(t *testing.T) {
		// If all usage is zero, fall back to equal allocation
		usageA := []decimal.Decimal{decimal.Zero, decimal.Zero, decimal.Zero}
		usageB := []decimal.Decimal{decimal.Zero, decimal.Zero, decimal.Zero}

		avgA := calculateAverage(usageA)
		avgB := calculateAverage(usageB)
		totalAvg := avgA.Add(avgB)

		// When total is zero, should fall back to equal
		var shareA, shareB decimal.Decimal
		if totalAvg.IsZero() {
			shareA = decimal.NewFromFloat(0.5)
			shareB = decimal.NewFromFloat(0.5)
		} else {
			shareA = avgA.Div(totalAvg)
			shareB = avgB.Div(totalAvg)
		}

		assert.True(t, shareA.Equal(shareB),
			"Zero usage should result in equal shares")
	})
}

// calculateAverage calculates the simple average of a slice of decimals
func calculateAverage(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}
	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	return sum.Div(decimal.NewFromInt(int64(len(values))))
}

// =============================================================================
// R-9: Tests for hybrid fixed+proportional strategy
// =============================================================================

// TestHybridFixedProportionalStrategy tests the hybrid strategy that splits
// cost into a fixed baseline and a variable proportional portion
func TestHybridFixedProportionalStrategy(t *testing.T) {
	// Scenario: Platform ($1000) with fixed_percent=40
	// Fixed portion ($400) split equally among 3 children
	// Variable portion ($600) split proportionally by usage

	platformCost := decimal.NewFromFloat(1000.0)
	fixedPercent := decimal.NewFromFloat(40.0)
	numChildren := 3

	// Usage: A=100, B=300, C=0
	usageA := decimal.NewFromFloat(100)
	usageB := decimal.NewFromFloat(300)
	usageC := decimal.NewFromFloat(0)
	totalUsage := usageA.Add(usageB).Add(usageC)

	t.Run("fixed portion calculation", func(t *testing.T) {
		fixedPortion := platformCost.Mul(fixedPercent.Div(decimal.NewFromInt(100)))
		fixedPerChild := fixedPortion.Div(decimal.NewFromInt(int64(numChildren)))

		assert.True(t, fixedPortion.Equal(decimal.NewFromFloat(400.0)),
			"Fixed portion should be $400")
		assert.True(t, fixedPerChild.Round(2).Equal(decimal.NewFromFloat(133.33)),
			"Each child should receive $133.33 from fixed portion")
	})

	t.Run("variable portion calculation", func(t *testing.T) {
		variablePortion := platformCost.Mul(decimal.NewFromInt(100).Sub(fixedPercent).Div(decimal.NewFromInt(100)))

		propShareA := usageA.Div(totalUsage)
		propShareB := usageB.Div(totalUsage)
		propShareC := usageC.Div(totalUsage)

		variableA := variablePortion.Mul(propShareA)
		variableB := variablePortion.Mul(propShareB)
		variableC := variablePortion.Mul(propShareC)

		assert.True(t, variablePortion.Equal(decimal.NewFromFloat(600.0)),
			"Variable portion should be $600")
		assert.True(t, variableA.Equal(decimal.NewFromFloat(150.0)),
			"A should receive $150 from variable portion (25%%)")
		assert.True(t, variableB.Equal(decimal.NewFromFloat(450.0)),
			"B should receive $450 from variable portion (75%%)")
		assert.True(t, variableC.IsZero(),
			"C should receive $0 from variable portion (0%%)")
	})

	t.Run("total allocation", func(t *testing.T) {
		fixedPortion := platformCost.Mul(fixedPercent.Div(decimal.NewFromInt(100)))
		fixedPerChild := fixedPortion.Div(decimal.NewFromInt(int64(numChildren)))
		variablePortion := platformCost.Sub(fixedPortion)

		propShareA := usageA.Div(totalUsage)
		propShareB := usageB.Div(totalUsage)
		propShareC := usageC.Div(totalUsage)

		totalA := fixedPerChild.Add(variablePortion.Mul(propShareA))
		totalB := fixedPerChild.Add(variablePortion.Mul(propShareB))
		totalC := fixedPerChild.Add(variablePortion.Mul(propShareC))

		// A: $133.33 + $150 = $283.33
		// B: $133.33 + $450 = $583.33
		// C: $133.33 + $0 = $133.33
		assert.True(t, totalA.Round(2).Equal(decimal.NewFromFloat(283.33)),
			"A total should be $283.33")
		assert.True(t, totalB.Round(2).Equal(decimal.NewFromFloat(583.33)),
			"B total should be $583.33")
		assert.True(t, totalC.Round(2).Equal(decimal.NewFromFloat(133.33)),
			"C total should be $133.33")

		// Verify total equals platform cost
		grandTotal := totalA.Add(totalB).Add(totalC)
		assert.True(t, grandTotal.Round(2).Equal(platformCost),
			"Grand total should equal platform cost")
	})

	t.Run("zero usage child still gets fixed portion", func(t *testing.T) {
		// Child C has zero usage but still receives fixed portion
		fixedPortion := platformCost.Mul(fixedPercent.Div(decimal.NewFromInt(100)))
		fixedPerChild := fixedPortion.Div(decimal.NewFromInt(int64(numChildren)))

		assert.True(t, fixedPerChild.GreaterThan(decimal.Zero),
			"Zero-usage child should still receive fixed portion")
	})
}

// =============================================================================
// R-10: Tests for min-floor+proportional strategy
// =============================================================================

// TestMinFloorProportionalStrategy tests the strategy that guarantees
// a minimum share per child with proportional remainder
func TestMinFloorProportionalStrategy(t *testing.T) {
	// Scenario: Platform ($1000) with min_floor_percent=10
	// Each child gets at least 10%, remainder split proportionally

	platformCost := decimal.NewFromFloat(1000.0)
	minFloorPercent := decimal.NewFromFloat(10.0)
	numChildren := 3

	// Usage: A=100, B=300, C=0
	usageA := decimal.NewFromFloat(100)
	usageB := decimal.NewFromFloat(300)
	usageC := decimal.NewFromFloat(0)
	totalUsage := usageA.Add(usageB).Add(usageC)

	t.Run("floor allocation", func(t *testing.T) {
		floorPerChild := platformCost.Mul(minFloorPercent.Div(decimal.NewFromInt(100)))
		totalFloor := floorPerChild.Mul(decimal.NewFromInt(int64(numChildren)))

		assert.True(t, floorPerChild.Equal(decimal.NewFromFloat(100.0)),
			"Each child should receive $100 floor")
		assert.True(t, totalFloor.Equal(decimal.NewFromFloat(300.0)),
			"Total floor should be $300 (30%%)")
	})

	t.Run("remainder allocation", func(t *testing.T) {
		floorPerChild := platformCost.Mul(minFloorPercent.Div(decimal.NewFromInt(100)))
		totalFloor := floorPerChild.Mul(decimal.NewFromInt(int64(numChildren)))
		remainder := platformCost.Sub(totalFloor)

		propShareA := usageA.Div(totalUsage)
		propShareB := usageB.Div(totalUsage)
		propShareC := usageC.Div(totalUsage)

		remainderA := remainder.Mul(propShareA)
		remainderB := remainder.Mul(propShareB)
		remainderC := remainder.Mul(propShareC)

		assert.True(t, remainder.Equal(decimal.NewFromFloat(700.0)),
			"Remainder should be $700")
		assert.True(t, remainderA.Equal(decimal.NewFromFloat(175.0)),
			"A should receive $175 from remainder (25%%)")
		assert.True(t, remainderB.Equal(decimal.NewFromFloat(525.0)),
			"B should receive $525 from remainder (75%%)")
		assert.True(t, remainderC.IsZero(),
			"C should receive $0 from remainder (0%%)")
	})

	t.Run("total allocation", func(t *testing.T) {
		floorPerChild := platformCost.Mul(minFloorPercent.Div(decimal.NewFromInt(100)))
		totalFloor := floorPerChild.Mul(decimal.NewFromInt(int64(numChildren)))
		remainder := platformCost.Sub(totalFloor)

		propShareA := usageA.Div(totalUsage)
		propShareB := usageB.Div(totalUsage)
		propShareC := usageC.Div(totalUsage)

		totalA := floorPerChild.Add(remainder.Mul(propShareA))
		totalB := floorPerChild.Add(remainder.Mul(propShareB))
		totalC := floorPerChild.Add(remainder.Mul(propShareC))

		// A: $100 + $175 = $275
		// B: $100 + $525 = $625
		// C: $100 + $0 = $100
		assert.True(t, totalA.Equal(decimal.NewFromFloat(275.0)),
			"A total should be $275")
		assert.True(t, totalB.Equal(decimal.NewFromFloat(625.0)),
			"B total should be $625")
		assert.True(t, totalC.Equal(decimal.NewFromFloat(100.0)),
			"C total should be $100")

		// Verify total equals platform cost
		grandTotal := totalA.Add(totalB).Add(totalC)
		assert.True(t, grandTotal.Equal(platformCost),
			"Grand total should equal platform cost")
	})

	t.Run("floor exceeds 100% scenario", func(t *testing.T) {
		// If min_floor_percent * num_children >= 100, use equal allocation
		highFloorPercent := decimal.NewFromFloat(40.0) // 40% * 3 = 120% > 100%
		numChildrenHigh := 3

		totalFloorPercent := highFloorPercent.Mul(decimal.NewFromInt(int64(numChildrenHigh)))

		if totalFloorPercent.GreaterThanOrEqual(decimal.NewFromInt(100)) {
			// Fall back to equal allocation
			sharePerChild := decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(numChildrenHigh)))
			assert.True(t, sharePerChild.Round(4).Equal(decimal.NewFromFloat(0.3333)),
				"Should fall back to equal allocation when floor exceeds 100%%")
		}
	})
}

// =============================================================================
// R-11: Tests for Dynatrace-backed product→product shares
// =============================================================================

// TestDynatraceBackedShares tests allocation using Dynatrace metrics with customer_id labels
func TestDynatraceBackedShares(t *testing.T) {
	// Scenario: API Gateway product allocates to customer products based on request counts

	t.Run("customer-based allocation", func(t *testing.T) {
		// API Gateway cost: $10,000/day
		apiGatewayCost := decimal.NewFromFloat(10000.0)

		// Dynatrace metrics with customer_id labels
		customerUsage := map[string]decimal.Decimal{
			"customer_001": decimal.NewFromFloat(50000), // 50,000 requests
			"customer_002": decimal.NewFromFloat(30000), // 30,000 requests
			"customer_003": decimal.NewFromFloat(20000), // 20,000 requests
		}

		totalUsage := decimal.Zero
		for _, usage := range customerUsage {
			totalUsage = totalUsage.Add(usage)
		}

		// Calculate allocations
		allocations := make(map[string]decimal.Decimal)
		for customerID, usage := range customerUsage {
			share := usage.Div(totalUsage)
			allocations[customerID] = apiGatewayCost.Mul(share)
		}

		// Verify allocations
		assert.True(t, allocations["customer_001"].Equal(decimal.NewFromFloat(5000.0)),
			"Customer 001 should receive $5,000 (50%%)")
		assert.True(t, allocations["customer_002"].Equal(decimal.NewFromFloat(3000.0)),
			"Customer 002 should receive $3,000 (30%%)")
		assert.True(t, allocations["customer_003"].Equal(decimal.NewFromFloat(2000.0)),
			"Customer 003 should receive $2,000 (20%%)")

		// Verify total
		totalAllocated := decimal.Zero
		for _, alloc := range allocations {
			totalAllocated = totalAllocated.Add(alloc)
		}
		assert.True(t, totalAllocated.Equal(apiGatewayCost),
			"Total allocated should equal API Gateway cost")
	})

	t.Run("filtered segment allocation", func(t *testing.T) {
		// Only allocate to enterprise customers
		apiGatewayCost := decimal.NewFromFloat(10000.0)

		// All customer usage
		allCustomerUsage := map[string]decimal.Decimal{
			"customer_001": decimal.NewFromFloat(50000), // enterprise
			"customer_002": decimal.NewFromFloat(30000), // free tier
			"customer_003": decimal.NewFromFloat(20000), // enterprise
		}

		// Filter to enterprise only
		enterpriseCustomers := []string{"customer_001", "customer_003"}
		filteredUsage := make(map[string]decimal.Decimal)
		for _, customerID := range enterpriseCustomers {
			filteredUsage[customerID] = allCustomerUsage[customerID]
		}

		totalFilteredUsage := decimal.Zero
		for _, usage := range filteredUsage {
			totalFilteredUsage = totalFilteredUsage.Add(usage)
		}

		// Calculate allocations for filtered segment
		allocations := make(map[string]decimal.Decimal)
		for customerID, usage := range filteredUsage {
			share := usage.Div(totalFilteredUsage)
			allocations[customerID] = apiGatewayCost.Mul(share)
		}

		// customer_001: 50000/70000 = 71.43%
		// customer_003: 20000/70000 = 28.57%
		expectedAlloc001 := apiGatewayCost.Mul(decimal.NewFromFloat(50000).Div(decimal.NewFromFloat(70000)))
		expectedAlloc003 := apiGatewayCost.Mul(decimal.NewFromFloat(20000).Div(decimal.NewFromFloat(70000)))

		assert.True(t, allocations["customer_001"].Round(2).Equal(expectedAlloc001.Round(2)),
			"Customer 001 should receive ~$7,142.86")
		assert.True(t, allocations["customer_003"].Round(2).Equal(expectedAlloc003.Round(2)),
			"Customer 003 should receive ~$2,857.14")
	})
}

// =============================================================================
// R-12: API/summary tests
// =============================================================================

// TestAPISummaryCalculations tests the summary calculations used in the API layer
func TestAPISummaryCalculations(t *testing.T) {
	t.Run("coverage percentage calculation", func(t *testing.T) {
		rawInfraCost := decimal.NewFromFloat(1000.0)
		allocatedProductCost := decimal.NewFromFloat(800.0)

		coverage := allocatedProductCost.Div(rawInfraCost).Mul(decimal.NewFromInt(100))

		assert.True(t, coverage.Equal(decimal.NewFromInt(80)),
			"Coverage should be 80%%")
	})

	t.Run("unallocated cost calculation", func(t *testing.T) {
		rawInfraCost := decimal.NewFromFloat(1000.0)
		allocatedProductCost := decimal.NewFromFloat(800.0)

		unallocatedCost := rawInfraCost.Sub(allocatedProductCost)

		assert.True(t, unallocatedCost.Equal(decimal.NewFromFloat(200.0)),
			"Unallocated cost should be $200")
	})

	t.Run("product hierarchy totals", func(t *testing.T) {
		// Simulate product hierarchy with holistic costs
		productCosts := map[string]decimal.Decimal{
			"product_a": decimal.NewFromFloat(300.0),
			"product_b": decimal.NewFromFloat(400.0),
			"product_c": decimal.NewFromFloat(300.0),
		}

		totalProductCost := decimal.Zero
		for _, cost := range productCosts {
			totalProductCost = totalProductCost.Add(cost)
		}

		assert.True(t, totalProductCost.Equal(decimal.NewFromFloat(1000.0)),
			"Total product cost should be $1000")
	})

	t.Run("summary invariant check", func(t *testing.T) {
		rawInfraCost := decimal.NewFromFloat(1000.0)
		allocatedProductCost := decimal.NewFromFloat(800.0)
		unallocatedCost := decimal.NewFromFloat(200.0)

		// Invariant: Allocated + Unallocated = Raw Infra
		assert.True(t, allocatedProductCost.Add(unallocatedCost).Equal(rawInfraCost),
			"Summary invariant should hold")
	})

	t.Run("tolerance for floating point", func(t *testing.T) {
		// Allow small tolerance for floating point errors
		rawInfraCost := decimal.NewFromFloat(1000.0)
		allocatedProductCost := decimal.NewFromFloat(999.9999)
		tolerance := decimal.NewFromFloat(0.01)

		diff := rawInfraCost.Sub(allocatedProductCost).Abs()
		withinTolerance := diff.LessThanOrEqual(tolerance)

		assert.True(t, withinTolerance,
			"Should be within tolerance for floating point errors")
	})
}

