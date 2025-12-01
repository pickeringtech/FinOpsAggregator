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

	log.Info().
		Str("run_id", run.ID.String()).
		Int("processed_days", processedDays).
		Int("allocations", len(allAllocations)).
		Int("contributions", len(allContributions)).
		Dur("processing_time", summary.ProcessingTime).
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

	// Build graph for this date
	log.Debug().Msg("About to build graph")
	g, err := e.builder.BuildForDate(ctx, date)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build graph")
		return nil, nil, fmt.Errorf("failed to build graph: %w", err)
	}
	log.Debug().Msg("Graph built successfully")

	// Get topological order (reverse for allocation)
	log.Debug().Msg("About to call TopologicalSort")
	order, err := g.TopologicalSort()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get topological order")
		return nil, nil, fmt.Errorf("failed to get topological order: %w", err)
	}

	log.Debug().
		Int("total_nodes", len(order)).
		Msg("Got topological order")

	// Load direct costs for all nodes
	directCosts, err := e.store.Costs.GetByDate(ctx, date, dimensions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load direct costs: %w", err)
	}

	// Organize costs by node and dimension
	costsByNode := make(map[uuid.UUID]map[string]decimal.Decimal)
	for _, cost := range directCosts {
		if costsByNode[cost.NodeID] == nil {
			costsByNode[cost.NodeID] = make(map[string]decimal.Decimal)
		}
		costsByNode[cost.NodeID][cost.Dimension] = cost.Amount
	}

	// Initialize indirect costs
	indirectCosts := make(map[uuid.UUID]map[string]decimal.Decimal)
	for nodeID := range g.Nodes() {
		indirectCosts[nodeID] = make(map[string]decimal.Decimal)
		for _, dim := range dimensions {
			indirectCosts[nodeID][dim] = decimal.Zero
		}
	}

	var allocations []models.AllocationResultByDimension
	var contributions []models.ContributionResultByDimension

	// Process nodes in forward topological order for top-down allocation
	log.Debug().
		Int("total_nodes_to_process", len(order)).
		Msg("Starting node processing loop")

	for i := 0; i < len(order); i++ {
		nodeID := order[i]

		// Get outgoing edges for this node
		edges := g.Edges(nodeID)

		// Get parent costs for this node
		parentDirectCosts := costsByNode[nodeID]
		parentIndirectCosts := indirectCosts[nodeID]

		log.Debug().
			Str("node_id", nodeID.String()).
			Int("edge_count", len(edges)).
			Interface("parent_direct_costs", parentDirectCosts).
			Interface("parent_indirect_costs", parentIndirectCosts).
			Msg("Processing node for allocation")

		if len(edges) == 0 {
			continue // No children to allocate to
		}

		for _, edge := range edges {
			childID := edge.ChildID

			// Process each dimension
			for _, dim := range dimensions {
				// Get parent's cost for this dimension
				parentDirectDim := decimal.Zero
				if costsByNode[nodeID] != nil {
					parentDirectDim = costsByNode[nodeID][dim]
				}
				parentIndirectDim := indirectCosts[nodeID][dim]
				parentTotalDim := parentDirectDim.Add(parentIndirectDim)

				if parentTotalDim.IsZero() {
					continue // No cost to allocate for this dimension
				}

				// Resolve allocation strategy for this edge and dimension
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
				share, err := strategy.CalculateShare(ctx, e.store, nodeID, childID, dim, date)
				if err != nil {
					log.Error().
						Err(err).
						Str("strategy", string(strategy.Type)).
						Str("parent_id", nodeID.String()).
						Str("child_id", childID.String()).
						Str("dimension", dim).
						Msg("Failed to calculate share, using zero")
					continue
				}

				log.Debug().
					Str("parent_id", nodeID.String()).
					Str("child_id", childID.String()).
					Str("dimension", dim).
					Str("strategy", string(strategy.Type)).
					Str("share", share.String()).
					Str("parent_total", parentTotalDim.String()).
					Msg("Calculated allocation share")

				// Calculate contribution amount (parent cost allocated to child)
				contribution := parentTotalDim.Mul(share)

				// Add to child's indirect costs
				indirectCosts[childID][dim] = indirectCosts[childID][dim].Add(contribution)
				
				// Record contribution
				if !contribution.IsZero() {
					contributions = append(contributions, models.ContributionResultByDimension{
						RunID:             runID,
						ParentID:          nodeID,
						ChildID:           childID,
						ContributionDate:  date,
						Dimension:         dim,
						ContributedAmount: contribution,
						Path:              []uuid.UUID{nodeID, childID}, // Simple path for now
					})

					log.Debug().
						Str("parent_id", nodeID.String()).
						Str("child_id", childID.String()).
						Str("dimension", dim).
						Str("amount", contribution.String()).
						Msg("Created contribution record")
				} else {
					log.Debug().
						Str("parent_id", nodeID.String()).
						Str("child_id", childID.String()).
						Str("dimension", dim).
						Msg("Skipped zero contribution")
				}
			}
		}
		
		// Record allocation for this node
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
	}

	log.Debug().
		Time("date", date).
		Int("allocations", len(allocations)).
		Int("contributions", len(contributions)).
		Msg("Day allocation completed")

	return allocations, contributions, nil
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
