package allocate

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// Strategy represents an allocation strategy
type Strategy struct {
	Type       models.AllocationStrategy  `json:"type"`
	Parameters map[string]interface{}     `json:"parameters"`
}

// StrategyResolver resolves allocation strategies for edges and dimensions
type StrategyResolver struct {
	store *store.Store
}

// NewStrategyResolver creates a new strategy resolver
func NewStrategyResolver(store *store.Store) *StrategyResolver {
	return &StrategyResolver{
		store: store,
	}
}

// ResolveStrategy resolves the allocation strategy for an edge and dimension
func (sr *StrategyResolver) ResolveStrategy(ctx context.Context, edge models.DependencyEdge, dimension string, date time.Time) (*Strategy, error) {
	// First, check for dimension-specific strategy override
	strategies, err := sr.store.Edges.GetStrategiesForEdge(ctx, edge.ID)
	if err != nil {
		log.Error().Err(err).Str("edge_id", edge.ID.String()).Msg("Failed to get edge strategies")
	} else {
		// Look for dimension-specific strategy
		for _, strategy := range strategies {
			if strategy.Dimension != nil && *strategy.Dimension == dimension {
				return &Strategy{
					Type:       models.AllocationStrategy(strategy.Strategy),
					Parameters: strategy.Parameters,
				}, nil
			}
		}
		
		// Look for default strategy override (dimension is null)
		for _, strategy := range strategies {
			if strategy.Dimension == nil {
				return &Strategy{
					Type:       models.AllocationStrategy(strategy.Strategy),
					Parameters: strategy.Parameters,
				}, nil
			}
		}
	}

	// Fall back to edge default strategy
	return &Strategy{
		Type:       models.AllocationStrategy(edge.DefaultStrategy),
		Parameters: edge.DefaultParameters,
	}, nil
}

// CalculateShare calculates the allocation share for a parent-child relationship
func (s *Strategy) CalculateShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	switch s.Type {
	case models.StrategyEqual:
		return s.calculateEqualShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyProportionalOn:
		return s.calculateProportionalShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyFixedPercent:
		return s.calculateFixedPercentShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyCappedProp:
		return s.calculateCappedProportionalShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyResidualToMax:
		return s.calculateResidualToMaxShare(ctx, store, parentID, childID, dimension, date)
	default:
		return decimal.Zero, fmt.Errorf("unknown strategy type: %s", s.Type)
	}
}

// calculateEqualShare calculates equal allocation among all parents
func (s *Strategy) calculateEqualShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get all parents of the child for this date
	edges, err := store.Edges.GetByChildID(ctx, childID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get parent edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	// Equal share among all parents
	return decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(len(edges)))), nil
}

// calculateProportionalShare calculates proportional allocation based on usage metric
func (s *Strategy) calculateProportionalShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the metric to use for proportional allocation
	metric, ok := s.Parameters["metric"].(string)
	if !ok {
		return decimal.Zero, fmt.Errorf("proportional_on strategy requires 'metric' parameter")
	}

	// Get all parents of the child
	edges, err := store.Edges.GetByChildID(ctx, childID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get parent edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	// Get usage values for all parents
	var totalUsage decimal.Decimal
	var parentUsage decimal.Decimal
	
	for _, edge := range edges {
		usage, err := store.Usage.GetByNodeAndDateRange(ctx, edge.ParentID, date, date, []string{metric})
		if err != nil {
			log.Error().Err(err).Str("node_id", edge.ParentID.String()).Str("metric", metric).Msg("Failed to get usage for proportional allocation")
			continue
		}
		
		var nodeUsage decimal.Decimal
		for _, u := range usage {
			if u.Metric == metric {
				nodeUsage = u.Value
				break
			}
		}
		
		totalUsage = totalUsage.Add(nodeUsage)
		if edge.ParentID == parentID {
			parentUsage = nodeUsage
		}
	}

	if totalUsage.IsZero() {
		// Fall back to equal allocation if no usage data
		return decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(len(edges)))), nil
	}

	return parentUsage.Div(totalUsage), nil
}

// calculateFixedPercentShare calculates fixed percentage allocation
func (s *Strategy) calculateFixedPercentShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the fixed percentage
	percentInterface, ok := s.Parameters["percent"]
	if !ok {
		return decimal.Zero, fmt.Errorf("fixed_percent strategy requires 'percent' parameter")
	}

	var percent decimal.Decimal
	switch v := percentInterface.(type) {
	case float64:
		percent = decimal.NewFromFloat(v)
	case string:
		var err error
		percent, err = decimal.NewFromString(v)
		if err != nil {
			return decimal.Zero, fmt.Errorf("invalid percent value: %v", v)
		}
	default:
		return decimal.Zero, fmt.Errorf("percent parameter must be float64 or string, got %T", v)
	}

	// Convert percentage to decimal (e.g., 25% -> 0.25)
	if percent.GreaterThan(decimal.NewFromInt(1)) {
		percent = percent.Div(decimal.NewFromInt(100))
	}

	return percent, nil
}

// calculateCappedProportionalShare calculates proportional allocation with a cap
func (s *Strategy) calculateCappedProportionalShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// First calculate proportional share
	proportionalShare, err := s.calculateProportionalShare(ctx, store, parentID, childID, dimension, date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to calculate proportional share: %w", err)
	}

	// Get the cap
	capInterface, ok := s.Parameters["cap"]
	if !ok {
		return proportionalShare, nil // No cap, return proportional share
	}

	var cap decimal.Decimal
	switch v := capInterface.(type) {
	case float64:
		cap = decimal.NewFromFloat(v)
	case string:
		var err error
		cap, err = decimal.NewFromString(v)
		if err != nil {
			return decimal.Zero, fmt.Errorf("invalid cap value: %v", v)
		}
	default:
		return decimal.Zero, fmt.Errorf("cap parameter must be float64 or string, got %T", v)
	}

	// Convert percentage to decimal if needed
	if cap.GreaterThan(decimal.NewFromInt(1)) {
		cap = cap.Div(decimal.NewFromInt(100))
	}

	// Return the minimum of proportional share and cap
	if proportionalShare.LessThan(cap) {
		return proportionalShare, nil
	}
	return cap, nil
}

// calculateResidualToMaxShare calculates allocation for the node with maximum usage
func (s *Strategy) calculateResidualToMaxShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the metric to use for finding max usage
	metric, ok := s.Parameters["metric"].(string)
	if !ok {
		return decimal.Zero, fmt.Errorf("residual_to_max strategy requires 'metric' parameter")
	}

	// Get all parents of the child
	edges, err := store.Edges.GetByChildID(ctx, childID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get parent edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	// Find the parent with maximum usage
	var maxUsage decimal.Decimal
	var maxUsageParentID uuid.UUID
	
	for _, edge := range edges {
		usage, err := store.Usage.GetByNodeAndDateRange(ctx, edge.ParentID, date, date, []string{metric})
		if err != nil {
			log.Error().Err(err).Str("node_id", edge.ParentID.String()).Str("metric", metric).Msg("Failed to get usage for residual_to_max allocation")
			continue
		}
		
		var nodeUsage decimal.Decimal
		for _, u := range usage {
			if u.Metric == metric {
				nodeUsage = u.Value
				break
			}
		}
		
		if nodeUsage.GreaterThan(maxUsage) {
			maxUsage = nodeUsage
			maxUsageParentID = edge.ParentID
		}
	}

	// Calculate shares for other parents first (using proportional)
	if parentID != maxUsageParentID {
		// Use proportional allocation for non-max parents
		return s.calculateProportionalShare(ctx, store, parentID, childID, dimension, date)
	}

	// For the max usage parent, calculate residual
	var totalOtherShares decimal.Decimal
	for _, edge := range edges {
		if edge.ParentID != maxUsageParentID {
			share, err := s.calculateProportionalShare(ctx, store, edge.ParentID, childID, dimension, date)
			if err != nil {
				log.Error().Err(err).Str("parent_id", edge.ParentID.String()).Msg("Failed to calculate proportional share for residual calculation")
				continue
			}
			totalOtherShares = totalOtherShares.Add(share)
		}
	}

	// Residual share is what's left after other allocations
	residualShare := decimal.NewFromInt(1).Sub(totalOtherShares)
	if residualShare.LessThan(decimal.Zero) {
		residualShare = decimal.Zero
	}

	return residualShare, nil
}
