package allocate

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/graph"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// Engine performs cost allocation computations
type Engine struct {
	store      *store.Store
	builder    *graph.GraphBuilder
	strategies *StrategyResolver
}

// NewEngine creates a new allocation engine
func NewEngine(store *store.Store) *Engine {
	return &Engine{
		store:      store,
		builder:    graph.NewGraphBuilder(store),
		strategies: NewStrategyResolver(store),
	}
}

// AllocateForPeriod performs cost allocation for a date range
func (e *Engine) AllocateForPeriod(ctx context.Context, startDate, endDate time.Time, dimensions []string) (*models.AllocationOutput, error) {
	log.Info().
		Time("start_date", startDate).
		Time("end_date", endDate).
		Strs("dimensions", dimensions).
		Msg("Starting allocation computation")

	startTime := time.Now()

	// Create computation run
	run := &models.ComputationRun{
		ID:          uuid.New(),
		WindowStart: startDate,
		WindowEnd:   endDate,
		Status:      string(models.ComputationStatusRunning),
	}

	// Build graph for the first date to get hash
	firstGraph, err := e.builder.BuildForDate(ctx, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to build initial graph: %w", err)
	}
	run.GraphHash = firstGraph.Hash()

	// Save computation run
	if err := e.store.Runs.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create computation run: %w", err)
	}

	// Update status to running
	if err := e.store.Runs.UpdateStatus(ctx, run.ID, string(models.ComputationStatusRunning), nil); err != nil {
		log.Error().Err(err).Msg("Failed to update run status to running")
	}

	var allAllocations []models.AllocationResultByDimension
	var allContributions []models.ContributionResultByDimension
	summary := models.AllocationSummary{
		TotalDirectCost:   make(map[string]decimal.Decimal),
		TotalIndirectCost: make(map[string]decimal.Decimal),
		TotalCost:         make(map[string]decimal.Decimal),
	}

	// Process each day
	processedDays := 0
	log.Debug().
		Time("start_date", startDate).
		Time("end_date", endDate).
		Msg("Starting daily allocation loop")
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		log.Debug().Time("processing_date", date).Msg("Processing date")
		dayAllocations, dayContributions, err := e.allocateForDay(ctx, run.ID, date, dimensions)
		if err != nil {
			// Update run status to failed
			notes := fmt.Sprintf("Failed on date %s: %v", date.Format("2006-01-02"), err)
			if updateErr := e.store.Runs.UpdateStatus(ctx, run.ID, string(models.ComputationStatusFailed), &notes); updateErr != nil {
				log.Error().Err(updateErr).Msg("Failed to update run status to failed")
			}
			return nil, fmt.Errorf("failed to allocate for date %s: %w", date.Format("2006-01-02"), err)
		}

		allAllocations = append(allAllocations, dayAllocations...)
		allContributions = append(allContributions, dayContributions...)
		processedDays++

		// Update summary
		for _, allocation := range dayAllocations {
			dim := allocation.Dimension
			if _, exists := summary.TotalDirectCost[dim]; !exists {
				summary.TotalDirectCost[dim] = decimal.Zero
				summary.TotalIndirectCost[dim] = decimal.Zero
				summary.TotalCost[dim] = decimal.Zero
			}
			summary.TotalDirectCost[dim] = summary.TotalDirectCost[dim].Add(allocation.DirectAmount)
			summary.TotalIndirectCost[dim] = summary.TotalIndirectCost[dim].Add(allocation.IndirectAmount)
			summary.TotalCost[dim] = summary.TotalCost[dim].Add(allocation.TotalAmount)
		}
	}

	// Save results in batches
	if err := e.saveResultsInBatches(ctx, allAllocations, allContributions); err != nil {
		notes := fmt.Sprintf("Failed to save results: %v", err)
		if updateErr := e.store.Runs.UpdateStatus(ctx, run.ID, string(models.ComputationStatusFailed), &notes); updateErr != nil {
			log.Error().Err(updateErr).Msg("Failed to update run status to failed")
		}
		return nil, fmt.Errorf("failed to save results: %w", err)
	}

	// Update run status to completed
	if err := e.store.Runs.UpdateStatus(ctx, run.ID, string(models.ComputationStatusCompleted), nil); err != nil {
		log.Error().Err(err).Msg("Failed to update run status to completed")
	}

	// Complete summary
	summary.TotalNodes = len(firstGraph.Nodes())
	summary.TotalEdges = firstGraph.Stats().EdgeCount
	summary.ProcessedDays = processedDays
	summary.ProcessingTime = time.Since(startTime)

	// Calculate and log coverage metrics
	finalCostCentres := firstGraph.GetFinalCostCentres()
	finalCostCentreSet := make(map[uuid.UUID]bool)
	for _, id := range finalCostCentres {
		finalCostCentreSet[id] = true
	}

	// Calculate totals for logging
	totalDirectCost := decimal.Zero
	totalIndirectCost := decimal.Zero
	finalCostCentreTotal := decimal.Zero
	for _, alloc := range allAllocations {
		totalDirectCost = totalDirectCost.Add(alloc.DirectAmount)
		totalIndirectCost = totalIndirectCost.Add(alloc.IndirectAmount)
		if finalCostCentreSet[alloc.NodeID] {
			finalCostCentreTotal = finalCostCentreTotal.Add(alloc.TotalAmount)
		}
	}

	// Calculate coverage
	coveragePercent := 0.0
	if !totalDirectCost.IsZero() {
		ratio, _ := finalCostCentreTotal.Div(totalDirectCost).Float64()
		if ratio > 0 && ratio <= 1 {
			coveragePercent = ratio * 100
		}
	}

	log.Info().
		Str("run_id", run.ID.String()).
		Int("processed_days", processedDays).
		Int("allocations", len(allAllocations)).
		Int("contributions", len(allContributions)).
		Dur("processing_time", summary.ProcessingTime).
		Int("total_nodes", summary.TotalNodes).
		Int("total_edges", summary.TotalEdges).
		Int("final_cost_centres", len(finalCostCentres)).
		Str("total_direct_cost", totalDirectCost.StringFixed(2)).
		Str("total_indirect_cost", totalIndirectCost.StringFixed(2)).
		Str("final_cost_centre_total", finalCostCentreTotal.StringFixed(2)).
		Float64("coverage_percent", coveragePercent).
		Msg("Allocation computation completed")

	return &models.AllocationOutput{
		RunID:         run.ID,
		Allocations:   allAllocations,
		Contributions: allContributions,
		Summary:       summary,
	}, nil
}

// allocateForDay performs allocation for a single day
func (e *Engine) allocateForDay(ctx context.Context, runID uuid.UUID, date time.Time, dimensions []string) ([]models.AllocationResultByDimension, []models.ContributionResultByDimension, error) {
	log.Debug().Time("date", date).Msg("Processing allocation for day")

	// Step 1: Build allocation graph
	g, order, err := e.buildAllocationGraph(ctx, date)
	if err != nil {
		return nil, nil, err
	}

	// Step 2: Load direct costs
	costsByNode, err := e.loadDirectCosts(ctx, date, dimensions)
	if err != nil {
		return nil, nil, err
	}

	// Step 3: Initialize indirect cost accumulators
	indirectCosts := e.initializeIndirectCosts(g, dimensions)

	// Step 4: Perform allocation traversal
	allocations, contributions := e.performAllocationTraversal(ctx, runID, date, g, order, dimensions, costsByNode, indirectCosts)

	log.Debug().
		Time("date", date).
		Int("allocations", len(allocations)).
		Int("contributions", len(contributions)).
		Msg("Day allocation completed")

	// Step 5: Validate allocation invariants
	e.validateAllocationInvariants(ctx, g, allocations, contributions, costsByNode, indirectCosts, dimensions, date)

	return allocations, contributions, nil
}

// buildAllocationGraph builds the allocation graph and returns it with topological order
func (e *Engine) buildAllocationGraph(ctx context.Context, date time.Time) (*graph.Graph, []uuid.UUID, error) {
	log.Debug().Msg("Building allocation graph")
	g, err := e.builder.BuildForDate(ctx, date)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build graph")
		return nil, nil, fmt.Errorf("failed to build graph: %w", err)
	}

	log.Debug().Msg("Computing topological order")
	order, err := g.TopologicalSort()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get topological order")
		return nil, nil, fmt.Errorf("failed to get topological order: %w", err)
	}

	log.Debug().
		Int("total_nodes", len(order)).
		Int("total_edges", g.Stats().EdgeCount).
		Msg("Graph built successfully")

	return g, order, nil
}

// loadDirectCosts loads direct costs for all nodes on a given date
func (e *Engine) loadDirectCosts(ctx context.Context, date time.Time, dimensions []string) (map[uuid.UUID]map[string]decimal.Decimal, error) {
	directCosts, err := e.store.Costs.GetByDate(ctx, date, dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to load direct costs: %w", err)
	}

	// Organize costs by node and dimension
	costsByNode := make(map[uuid.UUID]map[string]decimal.Decimal)
	for _, cost := range directCosts {
		if costsByNode[cost.NodeID] == nil {
			costsByNode[cost.NodeID] = make(map[string]decimal.Decimal)
		}
		costsByNode[cost.NodeID][cost.Dimension] = cost.Amount
	}

	log.Debug().
		Int("nodes_with_costs", len(costsByNode)).
		Int("total_cost_records", len(directCosts)).
		Msg("Direct costs loaded")

	return costsByNode, nil
}

// initializeIndirectCosts initializes indirect cost accumulators for all nodes
func (e *Engine) initializeIndirectCosts(g *graph.Graph, dimensions []string) map[uuid.UUID]map[string]decimal.Decimal {
	indirectCosts := make(map[uuid.UUID]map[string]decimal.Decimal)
	for nodeID := range g.Nodes() {
		indirectCosts[nodeID] = make(map[string]decimal.Decimal)
		for _, dim := range dimensions {
			indirectCosts[nodeID][dim] = decimal.Zero
		}
	}
	return indirectCosts
}

// performAllocationTraversal traverses the graph in topological order and allocates costs
func (e *Engine) performAllocationTraversal(
	ctx context.Context,
	runID uuid.UUID,
	date time.Time,
	g *graph.Graph,
	order []uuid.UUID,
	dimensions []string,
	costsByNode map[uuid.UUID]map[string]decimal.Decimal,
	indirectCosts map[uuid.UUID]map[string]decimal.Decimal,
) ([]models.AllocationResultByDimension, []models.ContributionResultByDimension) {
	var allocations []models.AllocationResultByDimension
	var contributions []models.ContributionResultByDimension

	log.Debug().
		Int("total_nodes_to_process", len(order)).
		Msg("Starting allocation traversal")

	for _, nodeID := range order {
		// Process outgoing edges (allocate to children)
		nodeContributions := e.allocateFromNode(ctx, runID, date, g, nodeID, dimensions, costsByNode, indirectCosts)
		contributions = append(contributions, nodeContributions...)

		// Record allocation result for this node
		nodeAllocations := e.recordNodeAllocations(runID, date, nodeID, dimensions, costsByNode, indirectCosts)
		allocations = append(allocations, nodeAllocations...)
	}

	return allocations, contributions
}

// allocateFromNode allocates costs from a parent node to its children
func (e *Engine) allocateFromNode(
	ctx context.Context,
	runID uuid.UUID,
	date time.Time,
	g *graph.Graph,
	nodeID uuid.UUID,
	dimensions []string,
	costsByNode map[uuid.UUID]map[string]decimal.Decimal,
	indirectCosts map[uuid.UUID]map[string]decimal.Decimal,
) []models.ContributionResultByDimension {
	var contributions []models.ContributionResultByDimension

	edges := g.Edges(nodeID)
	if len(edges) == 0 {
		return contributions // No children to allocate to
	}

	log.Debug().
		Str("node_id", nodeID.String()).
		Int("edge_count", len(edges)).
		Msg("Processing node for allocation")

	for _, edge := range edges {
		childID := edge.ChildID

		for _, dim := range dimensions {
			contribution := e.calculateContribution(ctx, nodeID, childID, edge, dim, date, costsByNode, indirectCosts)
			if contribution == nil {
				continue
			}

			// Add to child's indirect costs
			indirectCosts[childID][dim] = indirectCosts[childID][dim].Add(contribution.ContributedAmount)

			// Record contribution
			contribution.RunID = runID
			contributions = append(contributions, *contribution)

			log.Debug().
				Str("parent_id", nodeID.String()).
				Str("child_id", childID.String()).
				Str("dimension", dim).
				Str("amount", contribution.ContributedAmount.String()).
				Msg("Created contribution record")
		}
	}

	return contributions
}

// calculateContribution calculates the contribution from parent to child for a dimension
func (e *Engine) calculateContribution(
	ctx context.Context,
	parentID, childID uuid.UUID,
	edge models.DependencyEdge,
	dim string,
	date time.Time,
	costsByNode map[uuid.UUID]map[string]decimal.Decimal,
	indirectCosts map[uuid.UUID]map[string]decimal.Decimal,
) *models.ContributionResultByDimension {
	// Calculate parent's holistic cost for this dimension
	parentDirectDim := decimal.Zero
	if costsByNode[parentID] != nil {
		parentDirectDim = costsByNode[parentID][dim]
	}
	parentIndirectDim := indirectCosts[parentID][dim]
	parentTotalDim := parentDirectDim.Add(parentIndirectDim)

	if parentTotalDim.IsZero() {
		return nil // No cost to allocate
	}

	// Resolve and apply allocation strategy
	strategy, err := e.strategies.ResolveStrategy(ctx, edge, dim, date)
	if err != nil {
		log.Error().
			Err(err).
			Str("edge_id", edge.ID.String()).
			Str("dimension", dim).
			Msg("Failed to resolve strategy, using equal allocation")
		strategy = &Strategy{
			Type:       models.StrategyEqual,
			Parameters: make(map[string]interface{}),
		}
	}

	// Calculate allocation share
	share, err := strategy.CalculateShare(ctx, e.store, parentID, childID, dim, date)
	if err != nil {
		log.Error().
			Err(err).
			Str("strategy", string(strategy.Type)).
			Str("parent_id", parentID.String()).
			Str("child_id", childID.String()).
			Str("dimension", dim).
			Msg("Failed to calculate share, skipping")
		return nil
	}

	// Calculate contribution amount
	contribution := parentTotalDim.Mul(share)
	if contribution.IsZero() {
		return nil
	}

	return &models.ContributionResultByDimension{
		ParentID:          parentID,
		ChildID:           childID,
		ContributionDate:  date,
		Dimension:         dim,
		ContributedAmount: contribution,
		Path:              []uuid.UUID{parentID, childID},
	}
}

// recordNodeAllocations records the allocation results for a node
func (e *Engine) recordNodeAllocations(
	runID uuid.UUID,
	date time.Time,
	nodeID uuid.UUID,
	dimensions []string,
	costsByNode map[uuid.UUID]map[string]decimal.Decimal,
	indirectCosts map[uuid.UUID]map[string]decimal.Decimal,
) []models.AllocationResultByDimension {
	var allocations []models.AllocationResultByDimension

	for _, dim := range dimensions {
		direct := decimal.Zero
		if costsByNode[nodeID] != nil {
			direct = costsByNode[nodeID][dim]
		}
		indirect := indirectCosts[nodeID][dim]
		total := direct.Add(indirect)

		allocations = append(allocations, models.AllocationResultByDimension{
			RunID:          runID,
			NodeID:         nodeID,
			AllocationDate: date,
			Dimension:      dim,
			DirectAmount:   direct,
			IndirectAmount: indirect,
			TotalAmount:    total,
		})
	}

	return allocations
}

// saveResultsInBatches saves allocation and contribution results in batches
func (e *Engine) saveResultsInBatches(ctx context.Context, allocations []models.AllocationResultByDimension, contributions []models.ContributionResultByDimension) error {
	const batchSize = 1000

	// Save allocations in batches
	for i := 0; i < len(allocations); i += batchSize {
		end := i + batchSize
		if end > len(allocations) {
			end = len(allocations)
		}
		
		batch := allocations[i:end]
		if err := e.store.Runs.SaveAllocationResults(ctx, batch); err != nil {
			return fmt.Errorf("failed to save allocation batch %d-%d: %w", i, end, err)
		}
	}

	// Save contributions in batches
	for i := 0; i < len(contributions); i += batchSize {
		end := i + batchSize
		if end > len(contributions) {
			end = len(contributions)
		}

		batch := contributions[i:end]
		if err := e.store.Runs.SaveContributionResults(ctx, batch); err != nil {
			return fmt.Errorf("failed to save contribution batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

// validateAllocationInvariants checks allocation invariants and logs warnings if violated
// Invariants checked:
// 1. Share validity: For each parent, sum of shares to children <= 1
// 2. Conservation: Total cost allocated from a node <= node's holistic cost
// 3. No amplification: Sum of final cost centre costs <= Raw Infrastructure Cost
func (e *Engine) validateAllocationInvariants(
	ctx context.Context,
	g *graph.Graph,
	allocations []models.AllocationResultByDimension,
	contributions []models.ContributionResultByDimension,
	costsByNode map[uuid.UUID]map[string]decimal.Decimal,
	indirectCosts map[uuid.UUID]map[string]decimal.Decimal,
	dimensions []string,
	date time.Time,
) {
	// Tolerance for floating point comparisons (0.01% of total)
	tolerance := decimal.NewFromFloat(0.0001)

	// Build allocation map by node and dimension
	allocationMap := make(map[uuid.UUID]map[string]decimal.Decimal)
	for _, alloc := range allocations {
		if allocationMap[alloc.NodeID] == nil {
			allocationMap[alloc.NodeID] = make(map[string]decimal.Decimal)
		}
		allocationMap[alloc.NodeID][alloc.Dimension] = alloc.TotalAmount
	}

	// Build contribution map: parent -> dimension -> total contributed
	contributionsByParent := make(map[uuid.UUID]map[string]decimal.Decimal)
	for _, contrib := range contributions {
		if contributionsByParent[contrib.ParentID] == nil {
			contributionsByParent[contrib.ParentID] = make(map[string]decimal.Decimal)
		}
		contributionsByParent[contrib.ParentID][contrib.Dimension] = contributionsByParent[contrib.ParentID][contrib.Dimension].Add(contrib.ContributedAmount)
	}

	// Invariant 1: Check share validity (sum of shares <= 1)
	// This is implicitly checked by ensuring total contributed <= parent's holistic cost
	for parentID, dimContribs := range contributionsByParent {
		for dim, totalContributed := range dimContribs {
			parentHolistic := decimal.Zero
			if costsByNode[parentID] != nil {
				parentHolistic = costsByNode[parentID][dim]
			}
			if indirectCosts[parentID] != nil {
				parentHolistic = parentHolistic.Add(indirectCosts[parentID][dim])
			}

			if !parentHolistic.IsZero() {
				shareSum := totalContributed.Div(parentHolistic)
				if shareSum.GreaterThan(decimal.NewFromInt(1).Add(tolerance)) {
					log.Warn().
						Str("parent_id", parentID.String()).
						Str("dimension", dim).
						Str("share_sum", shareSum.StringFixed(4)).
						Str("total_contributed", totalContributed.StringFixed(2)).
						Str("parent_holistic", parentHolistic.StringFixed(2)).
						Time("date", date).
						Msg("INVARIANT VIOLATION: Sum of allocation shares exceeds 1")
				}
			}
		}
	}

	// Invariant 2: Calculate raw infrastructure cost and final cost centre totals
	rawInfraCost := make(map[string]decimal.Decimal)
	finalCostCentreTotal := make(map[string]decimal.Decimal)

	// Get final cost centres
	finalCostCentres := g.GetFinalCostCentres()
	finalCostCentreSet := make(map[uuid.UUID]bool)
	for _, id := range finalCostCentres {
		finalCostCentreSet[id] = true
	}

	// Calculate raw infrastructure cost (direct costs on infra-like nodes)
	for nodeID, dimCosts := range costsByNode {
		node, exists := g.Nodes()[nodeID]
		if !exists {
			continue
		}
		// Check if this is an infrastructure-like node
		nodeType := models.NodeType(node.Type)
		if nodeType == models.NodeTypeResource || nodeType == models.NodeTypeShared ||
			nodeType == models.NodeTypePlatform || nodeType == models.NodeTypeInfra ||
			node.IsPlatform {
			for dim, cost := range dimCosts {
				rawInfraCost[dim] = rawInfraCost[dim].Add(cost)
			}
		}
	}

	// Calculate final cost centre totals (holistic costs)
	for _, alloc := range allocations {
		if finalCostCentreSet[alloc.NodeID] {
			finalCostCentreTotal[alloc.Dimension] = finalCostCentreTotal[alloc.Dimension].Add(alloc.TotalAmount)
		}
	}

	// Invariant 3: No amplification - final cost centre sum <= raw infra cost
	for dim, rawCost := range rawInfraCost {
		finalTotal := finalCostCentreTotal[dim]
		if finalTotal.GreaterThan(rawCost.Add(rawCost.Mul(tolerance))) {
			log.Warn().
				Str("dimension", dim).
				Str("raw_infra_cost", rawCost.StringFixed(2)).
				Str("final_cost_centre_total", finalTotal.StringFixed(2)).
				Str("difference", finalTotal.Sub(rawCost).StringFixed(2)).
				Time("date", date).
				Msg("INVARIANT VIOLATION: Final cost centre total exceeds raw infrastructure cost (amplification detected)")
		}
	}

	// Log summary of invariant checks
	log.Debug().
		Int("final_cost_centres", len(finalCostCentres)).
		Interface("raw_infra_cost", rawInfraCost).
		Interface("final_cost_centre_total", finalCostCentreTotal).
		Time("date", date).
		Msg("Allocation invariant checks completed")
}
