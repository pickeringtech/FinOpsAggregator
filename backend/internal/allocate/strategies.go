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
	var share decimal.Decimal
	var err error

	switch s.Type {
	case models.StrategyEqual:
		share, err = s.calculateEqualShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyProportionalOn:
		share, err = s.calculateProportionalShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyFixedPercent:
		share, err = s.calculateFixedPercentShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyCappedProp:
		share, err = s.calculateCappedProportionalShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyResidualToMax:
		share, err = s.calculateResidualToMaxShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyWeightedAverage:
		share, err = s.calculateWeightedAverageShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyHybridFixedProp:
		share, err = s.calculateHybridFixedProportionalShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategyMinFloorProportional:
		share, err = s.calculateMinFloorProportionalShare(ctx, store, parentID, childID, dimension, date)
	case models.StrategySegmentFilteredProp:
		share, err = s.calculateSegmentFilteredProportionalShare(ctx, store, parentID, childID, dimension, date)
	default:
		return decimal.Zero, fmt.Errorf("unknown strategy type: %s", s.Type)
	}

	if err != nil {
		log.Debug().
			Err(err).
			Str("strategy", string(s.Type)).
			Str("parent_id", parentID.String()).
			Str("child_id", childID.String()).
			Str("dimension", dimension).
			Time("date", date).
			Msg("Strategy calculation failed")
		return decimal.Zero, err
	}

	log.Trace().
		Str("strategy", string(s.Type)).
		Str("parent_id", parentID.String()).
		Str("child_id", childID.String()).
		Str("dimension", dimension).
		Str("share", share.StringFixed(6)).
		Time("date", date).
		Msg("Strategy share calculated")

	return share, nil
}

// calculateEqualShare calculates equal allocation among all children
func (s *Strategy) calculateEqualShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get all children of the parent for this date (for top-down allocation)
	edges, err := store.Edges.GetByParentID(ctx, parentID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get child edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	// Equal share among all children
	return decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(len(edges)))), nil
}

// calculateProportionalShare calculates proportional allocation based on usage metric
func (s *Strategy) calculateProportionalShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the metric to use for proportional allocation
	metric, ok := s.Parameters["metric"].(string)
	if !ok {
		return decimal.Zero, fmt.Errorf("proportional_on strategy requires 'metric' parameter")
	}

	// Get all children of the parent (for top-down allocation)
	edges, err := store.Edges.GetByParentID(ctx, parentID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get child edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	// Get usage values for all children
	var totalUsage decimal.Decimal
	var childUsage decimal.Decimal

	for _, edge := range edges {
		usage, err := store.Usage.GetByNodeAndDateRange(ctx, edge.ChildID, date, date, []string{metric})
		if err != nil {
			log.Error().Err(err).Str("node_id", edge.ChildID.String()).Str("metric", metric).Msg("Failed to get usage for proportional allocation")
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
		if edge.ChildID == childID {
			childUsage = nodeUsage
		}
	}

	if totalUsage.IsZero() {
		// Fall back to equal allocation if no usage data
		return decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(len(edges)))), nil
	}

	return childUsage.Div(totalUsage), nil
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

// calculateWeightedAverageShare calculates allocation based on weighted average usage over a look-back window
func (s *Strategy) calculateWeightedAverageShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the metric to use for proportional allocation
	metric, ok := s.Parameters["metric"].(string)
	if !ok {
		return decimal.Zero, fmt.Errorf("weighted_average strategy requires 'metric' parameter")
	}

	// Get the look-back window (default 7 days)
	windowDays := 7
	if windowInterface, ok := s.Parameters["window_days"]; ok {
		switch v := windowInterface.(type) {
		case float64:
			windowDays = int(v)
		case int:
			windowDays = v
		}
	}

	// Get all children of the parent
	edges, err := store.Edges.GetByParentID(ctx, parentID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get child edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	// Calculate start date for look-back window
	startDate := date.AddDate(0, 0, -(windowDays - 1))

	// Get average usage values for all children over the window
	var totalAvgUsage decimal.Decimal
	var childAvgUsage decimal.Decimal

	for _, edge := range edges {
		usage, err := store.Usage.GetByNodeAndDateRange(ctx, edge.ChildID, startDate, date, []string{metric})
		if err != nil {
			log.Error().Err(err).Str("node_id", edge.ChildID.String()).Str("metric", metric).Msg("Failed to get usage for weighted_average allocation")
			continue
		}

		// Calculate average usage over the window
		var sumUsage decimal.Decimal
		var count int
		for _, u := range usage {
			if u.Metric == metric {
				sumUsage = sumUsage.Add(u.Value)
				count++
			}
		}

		var avgUsage decimal.Decimal
		if count > 0 {
			avgUsage = sumUsage.Div(decimal.NewFromInt(int64(count)))
		}

		totalAvgUsage = totalAvgUsage.Add(avgUsage)
		if edge.ChildID == childID {
			childAvgUsage = avgUsage
		}
	}

	if totalAvgUsage.IsZero() {
		// Fall back to equal allocation if no usage data
		return decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(len(edges)))), nil
	}

	return childAvgUsage.Div(totalAvgUsage), nil
}

// calculateHybridFixedProportionalShare calculates allocation with a fixed baseline plus proportional variable
func (s *Strategy) calculateHybridFixedProportionalShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the fixed percentage (portion allocated equally)
	fixedPercentInterface, ok := s.Parameters["fixed_percent"]
	if !ok {
		return decimal.Zero, fmt.Errorf("hybrid_fixed_proportional strategy requires 'fixed_percent' parameter")
	}

	var fixedPercent decimal.Decimal
	switch v := fixedPercentInterface.(type) {
	case float64:
		fixedPercent = decimal.NewFromFloat(v)
	case string:
		var err error
		fixedPercent, err = decimal.NewFromString(v)
		if err != nil {
			return decimal.Zero, fmt.Errorf("invalid fixed_percent value: %v", v)
		}
	default:
		return decimal.Zero, fmt.Errorf("fixed_percent parameter must be float64 or string, got %T", v)
	}

	// Convert percentage to decimal if needed (e.g., 40 -> 0.40)
	if fixedPercent.GreaterThan(decimal.NewFromInt(1)) {
		fixedPercent = fixedPercent.Div(decimal.NewFromInt(100))
	}

	// Get all children of the parent
	edges, err := store.Edges.GetByParentID(ctx, parentID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get child edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	numChildren := decimal.NewFromInt(int64(len(edges)))

	// Fixed portion: split equally among all children
	fixedShare := fixedPercent.Div(numChildren)

	// Variable portion: split proportionally
	variablePercent := decimal.NewFromInt(1).Sub(fixedPercent)

	// Calculate proportional share for the variable portion
	proportionalShare, err := s.calculateProportionalShare(ctx, store, parentID, childID, dimension, date)
	if err != nil {
		// Fall back to equal for variable portion if proportional fails
		proportionalShare = decimal.NewFromInt(1).Div(numChildren)
	}

	variableShare := variablePercent.Mul(proportionalShare)

	return fixedShare.Add(variableShare), nil
}

// calculateMinFloorProportionalShare calculates allocation with a minimum floor per child
func (s *Strategy) calculateMinFloorProportionalShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the minimum floor percentage per child
	minFloorInterface, ok := s.Parameters["min_floor_percent"]
	if !ok {
		return decimal.Zero, fmt.Errorf("min_floor_proportional strategy requires 'min_floor_percent' parameter")
	}

	var minFloorPercent decimal.Decimal
	switch v := minFloorInterface.(type) {
	case float64:
		minFloorPercent = decimal.NewFromFloat(v)
	case string:
		var err error
		minFloorPercent, err = decimal.NewFromString(v)
		if err != nil {
			return decimal.Zero, fmt.Errorf("invalid min_floor_percent value: %v", v)
		}
	default:
		return decimal.Zero, fmt.Errorf("min_floor_percent parameter must be float64 or string, got %T", v)
	}

	// Convert percentage to decimal if needed (e.g., 10 -> 0.10)
	if minFloorPercent.GreaterThan(decimal.NewFromInt(1)) {
		minFloorPercent = minFloorPercent.Div(decimal.NewFromInt(100))
	}

	// Get all children of the parent
	edges, err := store.Edges.GetByParentID(ctx, parentID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get child edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	numChildren := decimal.NewFromInt(int64(len(edges)))

	// Check if total floor exceeds 100%
	totalFloor := minFloorPercent.Mul(numChildren)
	if totalFloor.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		// Fall back to equal allocation
		return decimal.NewFromInt(1).Div(numChildren), nil
	}

	// Calculate remainder after floor allocations
	remainder := decimal.NewFromInt(1).Sub(totalFloor)

	// Calculate proportional share for the remainder
	proportionalShare, err := s.calculateProportionalShare(ctx, store, parentID, childID, dimension, date)
	if err != nil {
		// Fall back to equal for remainder if proportional fails
		proportionalShare = decimal.NewFromInt(1).Div(numChildren)
	}

	// Total share = floor + (remainder * proportional share)
	return minFloorPercent.Add(remainder.Mul(proportionalShare)), nil
}

// calculateSegmentFilteredProportionalShare calculates proportional allocation based on usage metrics
// filtered by specific label values (e.g., customer_id, environment, plan_tier).
// This enables segment-based cost allocation using Dynatrace or other labelled metrics.
//
// Parameters:
//   - metric: The usage metric to use for proportional allocation (required)
//   - segment_filter.label: The label key to filter on (e.g., "customer_id")
//   - segment_filter.values: Array of label values to include (OR semantics)
//   - segment_filter.operator: Filter operator ("eq", "in", "exists", etc.)
func (s *Strategy) calculateSegmentFilteredProportionalShare(ctx context.Context, store *store.Store, parentID, childID uuid.UUID, dimension string, date time.Time) (decimal.Decimal, error) {
	// Get the metric to use for proportional allocation
	metric, ok := s.Parameters["metric"].(string)
	if !ok {
		return decimal.Zero, fmt.Errorf("segment_filtered_proportional strategy requires 'metric' parameter")
	}

	// Parse segment filter from parameters
	segmentFilter, err := s.parseSegmentFilter()
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse segment filter: %w", err)
	}

	// Get all children of the parent (for top-down allocation)
	edges, err := store.Edges.GetByParentID(ctx, parentID, &date)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get child edges: %w", err)
	}

	if len(edges) == 0 {
		return decimal.Zero, nil
	}

	// Build usage query options with label filters
	opts := models.UsageQueryOptions{
		Metrics:   []string{metric},
		StartDate: date,
		EndDate:   date,
	}

	// Add label filter if specified
	if segmentFilter != nil {
		opts.LabelFilters = []models.UsageLabelFilter{*segmentFilter}
	}

	// Get usage values for all children with label filtering
	var totalUsage decimal.Decimal
	var childUsage decimal.Decimal

	for _, edge := range edges {
		opts.NodeIDs = []uuid.UUID{edge.ChildID}

		usage, err := store.Usage.QueryWithOptions(ctx, opts)
		if err != nil {
			log.Error().Err(err).Str("node_id", edge.ChildID.String()).Str("metric", metric).Msg("Failed to get filtered usage for segment allocation")
			continue
		}

		// Sum all matching usage values (may have multiple records with different label values)
		var nodeUsage decimal.Decimal
		for _, u := range usage {
			if u.Metric == metric {
				nodeUsage = nodeUsage.Add(u.Value)
			}
		}

		totalUsage = totalUsage.Add(nodeUsage)
		if edge.ChildID == childID {
			childUsage = nodeUsage
		}
	}

	if totalUsage.IsZero() {
		// Fall back to equal allocation if no usage data matches the filter
		log.Debug().
			Str("parent_id", parentID.String()).
			Str("metric", metric).
			Msg("No filtered usage data found, falling back to equal allocation")
		return decimal.NewFromInt(1).Div(decimal.NewFromInt(int64(len(edges)))), nil
	}

	return childUsage.Div(totalUsage), nil
}

// parseSegmentFilter parses the segment_filter parameter from strategy parameters
func (s *Strategy) parseSegmentFilter() (*models.UsageLabelFilter, error) {
	filterParam, ok := s.Parameters["segment_filter"]
	if !ok {
		return nil, nil // No filter specified, return nil (no filtering)
	}

	filterMap, ok := filterParam.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("segment_filter must be an object")
	}

	filter := &models.UsageLabelFilter{
		Operator: "in", // Default operator
	}

	// Parse label key
	if label, ok := filterMap["label"].(string); ok {
		filter.Key = label
	} else {
		return nil, fmt.Errorf("segment_filter.label is required")
	}

	// Parse operator
	if op, ok := filterMap["operator"].(string); ok {
		filter.Operator = op
	}

	// Parse values
	if values, ok := filterMap["values"].([]interface{}); ok {
		for _, v := range values {
			if str, ok := v.(string); ok {
				filter.Values = append(filter.Values, str)
			}
		}
	} else if value, ok := filterMap["value"].(string); ok {
		// Support single value as well
		filter.Values = []string{value}
		if filter.Operator == "in" {
			filter.Operator = "eq"
		}
	}

	return filter, nil
}
