package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/analysis"
	"github.com/pickeringtech/FinOpsAggregator/internal/analyzer"
	"github.com/pickeringtech/FinOpsAggregator/internal/graph"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// Service provides business logic for the API
type Service struct {
	store                  *store.Store
	analyzer               *analysis.FinOpsAnalyzer
	recommendationAnalyzer *analyzer.RecommendationAnalyzer
	graphBuilder           *graph.GraphBuilder
}

// NewService creates a new API service
func NewService(store *store.Store) *Service {
	return &Service{
		store:                  store,
		analyzer:               analysis.NewFinOpsAnalyzer(store),
		recommendationAnalyzer: analyzer.NewRecommendationAnalyzer(store),
		graphBuilder:           graph.NewGraphBuilder(store),
	}
}

// GetProductHierarchy retrieves the product hierarchy with cost data
func (s *Service) GetProductHierarchy(ctx context.Context, req CostAttributionRequest) (*ProductHierarchyResponse, error) {
	// Build product hierarchy tree
	products, err := s.buildProductTree(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build product tree: %w", err)
	}

	// Build graph to identify final cost centres
	// Use the end date of the request period for graph structure
	g, err := s.graphBuilder.BuildForDate(ctx, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph for final cost centre detection: %w", err)
	}

	// Get final cost centres (product nodes with no outgoing product→product edges)
	finalCostCentreIDs := g.GetFinalCostCentres()
	finalCostCentreSet := make(map[uuid.UUID]bool)
	for _, id := range finalCostCentreIDs {
		finalCostCentreSet[id] = true
	}

	// Calculate total allocated costs as sum of final cost centre holistic costs only
	// This prevents double-counting when products roll up to other products
	totalAllocatedCost := decimal.Zero
	for _, product := range products {
		if finalCostCentreSet[product.ID] {
			totalAllocatedCost = totalAllocatedCost.Add(product.HolisticCosts.Total)
		}
	}

	// Get total infrastructure costs (platform + shared raw infrastructure only)
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	rawTotalCosts, err := s.store.Costs.GetTotalInfrastructureCostByDateRange(ctx, req.StartDate, req.EndDate, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get total infrastructure costs: %w", err)
	}

	// Calculate unallocated costs as the gap between raw infra spend and allocated product totals
	// Unallocated = Raw Infrastructure Cost - Allocated Product Cost
	unallocatedCost := rawTotalCosts.Sub(totalAllocatedCost)
	if unallocatedCost.LessThan(decimal.Zero) {
		// Clamp to zero - negative unallocated indicates data inconsistency
		unallocatedCost = decimal.Zero
	}

	// Get platform and shared nodes for unallocated breakdown
	unallocatedNode, err := s.buildUnallocatedNode(ctx, req, unallocatedCost)
	if err != nil {
		return nil, fmt.Errorf("failed to build unallocated node: %w", err)
	}

	// Track product count before adding unallocated node
	productCount := len(products)
	finalCostCentreCount := len(finalCostCentreIDs)

	// Add unallocated node to products list if there are unallocated costs
	if unallocatedCost.GreaterThan(decimal.Zero) {
		products = append(products, *unallocatedNode)
	}

	// Total cost shown in summary is sum of all product holistic costs (including unallocated)
	totalCostForSummary := totalAllocatedCost.Add(unallocatedCost)

	// Derive allocation coverage percentage based only on costs allocated to final cost centres
	// Coverage = Allocated Product Cost / Raw Infrastructure Cost
	coveragePercent := 0.0
	if !rawTotalCosts.IsZero() {
		// Clamp to [0, 100]
		ratio, _ := totalAllocatedCost.Div(rawTotalCosts).Float64()
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		coveragePercent = ratio * 100
	}

	// Perform sanity checks on the product tree totals
	sanityCheckPassed, sanityWarnings := s.performProductTreeSanityChecks(
		products,
		totalAllocatedCost,
		rawTotalCosts,
		unallocatedCost,
		finalCostCentreSet,
	)

	summary := CostSummary{
		TotalCost:                 totalCostForSummary,
		RawTotalCost:              rawTotalCosts,
		AllocationCoveragePercent: coveragePercent,
		Currency:                  currency,
		Period:                    fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		StartDate:                 req.StartDate,
		EndDate:                   req.EndDate,
		ProductCount:              productCount,
		FinalCostCentreCount:      finalCostCentreCount,
		SanityCheckPassed:         sanityCheckPassed,
		SanityCheckWarnings:       sanityWarnings,
	}

	return &ProductHierarchyResponse{
		Products: products,
		Summary:  summary,
	}, nil
}

// GetIndividualNode retrieves detailed cost data for a single node
func (s *Service) GetIndividualNode(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) (*IndividualNodeResponse, error) {
	// Get node details
	node, err := s.store.Nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get direct costs
	directCosts, err := s.getNodeDirectCosts(ctx, nodeID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct costs: %w", err)
	}

	// Get allocated costs (costs allocated to this node from others)
	allocatedCosts, err := s.getNodeAllocatedCosts(ctx, nodeID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocated costs: %w", err)
	}

	// Calculate total costs
	totalCosts := s.combineCostBreakdowns(directCosts, allocatedCosts)

	// Get dependencies
	dependencies, err := s.getNodeDependencies(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	// Get allocation details
	allocations, err := s.getNodeAllocations(ctx, nodeID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocations: %w", err)
	}

	nodeDetails := NodeDetails{
		ID:         node.ID,
		Name:       node.Name,
		Type:       node.Type,
		IsPlatform: node.IsPlatform,
		CostLabels: node.CostLabels,
		Metadata:   node.Metadata,
	}

	return &IndividualNodeResponse{
		Node:           nodeDetails,
		DirectCosts:    *directCosts,
		AllocatedCosts: *allocatedCosts,
		TotalCosts:     *totalCosts,
		Dependencies:   dependencies,
		Allocations:    allocations,
	}, nil
}

// GetPlatformServices retrieves platform and shared services cost data
func (s *Service) GetPlatformServices(ctx context.Context, req CostAttributionRequest) (*PlatformServicesResponse, error) {
	// Get platform nodes
	platformNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		IsPlatform: &[]bool{true}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get platform nodes: %w", err)
	}

	// Get shared service nodes (nodes of type "shared")
	sharedNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		Type: string(models.NodeTypeShared),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get shared nodes: %w", err)
	}

	// Build platform services
	platformServices := make([]PlatformService, 0, len(platformNodes))
	for _, node := range platformNodes {
		service, err := s.buildPlatformService(ctx, node, req)
		if err != nil {
			return nil, fmt.Errorf("failed to build platform service for node %s: %w", node.Name, err)
		}
		platformServices = append(platformServices, *service)
	}

	// Build shared services
	sharedServices := make([]SharedService, 0, len(sharedNodes))
	for _, node := range sharedNodes {
		service, err := s.buildSharedService(ctx, node, req)
		if err != nil {
			return nil, fmt.Errorf("failed to build shared service for node %s: %w", node.Name, err)
		}
		sharedServices = append(sharedServices, *service)
	}

	// Get weighted allocations
	weightedAllocations, err := s.getWeightedAllocations(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get weighted allocations: %w", err)
	}

	// Calculate summary
	summary, err := s.calculatePlatformSummary(ctx, req, len(platformNodes), len(sharedNodes))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate platform summary: %w", err)
	}

	return &PlatformServicesResponse{
		PlatformServices:    platformServices,
		SharedServices:      sharedServices,
		Summary:             *summary,
		WeightedAllocations: weightedAllocations,
	}, nil
}

// ListProducts retrieves a flat list of products with their costs
func (s *Service) ListProducts(ctx context.Context, req CostAttributionRequest, limit, offset int) (*NodeListResponse, error) {
	nodes, err := s.store.Costs.ListNodesWithCosts(ctx, req.StartDate, req.EndDate, req.Currency, "product", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to response format
	nodeData := make([]NodeWithCostData, len(nodes))
	for i, node := range nodes {
		nodeData[i] = NodeWithCostData{
			ID:        node.ID,
			Name:      node.Name,
			Type:      node.Type,
			TotalCost: node.TotalCost,
			Currency:  node.Currency,
		}
	}

	return &NodeListResponse{
		Nodes:      nodeData,
		TotalCount: len(nodeData), // TODO: Get actual total count from DB
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// ListNodes retrieves a flat list of all nodes with their costs
func (s *Service) ListNodes(ctx context.Context, req CostAttributionRequest, nodeType string, limit, offset int) (*NodeListResponse, error) {
	nodes, err := s.store.Costs.ListNodesWithCosts(ctx, req.StartDate, req.EndDate, req.Currency, nodeType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Convert to response format
	nodeData := make([]NodeWithCostData, len(nodes))
	for i, node := range nodes {
		nodeData[i] = NodeWithCostData{
			ID:        node.ID,
			Name:      node.Name,
			Type:      node.Type,
			TotalCost: node.TotalCost,
			Currency:  node.Currency,
		}
	}

	return &NodeListResponse{
		Nodes:      nodeData,
		TotalCount: len(nodeData), // TODO: Get actual total count from DB
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// GetCostsByType retrieves costs aggregated by node type
func (s *Service) GetCostsByType(ctx context.Context, req CostAttributionRequest) (*CostsByTypeResponse, error) {
	costsByType, err := s.store.Costs.GetCostsByType(ctx, req.StartDate, req.EndDate, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by type: %w", err)
	}

	// Calculate total and convert to response format
	totalCost := decimal.Zero
	aggregations := make([]TypeAggregation, len(costsByType))
	for i, ct := range costsByType {
		totalCost = totalCost.Add(ct.TotalCost)
		aggregations[i] = TypeAggregation{
			Type:           ct.Type,
			DirectCost:     ct.DirectCost,
			IndirectCost:   ct.IndirectCost,
			TotalCost:      ct.TotalCost,
			NodeCount:      ct.NodeCount,
			PercentOfTotal: ct.PercentOfTotal,
		}
	}

	return &CostsByTypeResponse{
		Aggregations: aggregations,
		TotalCost:    totalCost,
		Currency:     req.Currency,
	}, nil
}

// GetCostsByDimension retrieves costs aggregated by a custom dimension
func (s *Service) GetCostsByDimension(ctx context.Context, req CostAttributionRequest, dimensionKey string) (*CostsByDimensionResponse, error) {
	costsByDim, err := s.store.Costs.GetCostsByDimension(ctx, req.StartDate, req.EndDate, req.Currency, dimensionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by dimension: %w", err)
	}

	// Calculate total and convert to response format
	totalCost := decimal.Zero
	aggregations := make([]DimensionAggregation, len(costsByDim))
	for i, cd := range costsByDim {
		totalCost = totalCost.Add(cd.TotalCost)
		aggregations[i] = DimensionAggregation{
			Value:     cd.DimensionValue,
			TotalCost: cd.TotalCost,
			NodeCount: cd.NodeCount,
		}
	}

	return &CostsByDimensionResponse{
		DimensionKey: dimensionKey,
		Aggregations: aggregations,
		TotalCost:    totalCost,
		Currency:     req.Currency,
	}, nil
}

// buildProductTree builds the product hierarchy tree
func (s *Service) buildProductTree(ctx context.Context, req CostAttributionRequest) ([]ProductNode, error) {
	// Get all product nodes
	productNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		Type: string(models.NodeTypeProduct),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get product nodes: %w", err)
	}

	products := make([]ProductNode, 0, len(productNodes))
	for _, node := range productNodes {
		product, err := s.buildProductNode(ctx, node, req)
		if err != nil {
			return nil, fmt.Errorf("failed to build product node %s: %w", node.Name, err)
		}
		products = append(products, *product)
	}

	return products, nil
}

// buildProductNode builds a single product node with its cost data
func (s *Service) buildProductNode(ctx context.Context, node models.CostNode, req CostAttributionRequest) (*ProductNode, error) {
	// Get direct costs
	directCosts, err := s.getNodeDirectCosts(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct costs: %w", err)
	}

	// Get holistic costs (direct + allocated)
	allocatedCosts, err := s.getNodeAllocatedCosts(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocated costs: %w", err)
	}
	holisticCosts := s.combineCostBreakdowns(directCosts, allocatedCosts)

	// Get shared service costs specifically
	sharedServiceCosts, err := s.getNodeSharedServiceCosts(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared service costs: %w", err)
	}

	// Get child nodes (dependencies)
	children, err := s.getProductChildren(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get product children: %w", err)
	}

	return &ProductNode{
		ID:                 node.ID,
		Name:               node.Name,
		Type:               node.Type,
		DirectCosts:        *directCosts,
		HolisticCosts:      *holisticCosts,
		SharedServiceCosts: *sharedServiceCosts,
		Children:           children,
		Metadata:           node.Metadata,
	}, nil
}

// getNodeDirectCosts retrieves direct costs for a node from allocation results
func (s *Service) getNodeDirectCosts(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) (*CostBreakdown, error) {
	// Use allocated costs from computation results instead of raw input costs
	allocatedCosts, err := s.store.Costs.GetAllocatedCostsByNodeAndDateRange(ctx, nodeID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get node costs: %w", err)
	}

	return s.buildCostBreakdownFromAllocations(allocatedCosts, req), nil
}

// buildCostBreakdown builds a cost breakdown from cost data
func (s *Service) buildCostBreakdown(costs []models.NodeCostByDimension, req CostAttributionRequest) *CostBreakdown {
	breakdown := &CostBreakdown{
		Total:      decimal.Zero,
		Currency:   req.Currency,
		Dimensions: make(map[string]decimal.Decimal),
		Trend:      []DailyCostPoint{},
	}

	if breakdown.Currency == "" {
		breakdown.Currency = "USD"
	}

	// Aggregate by dimension
	for _, cost := range costs {
		breakdown.Total = breakdown.Total.Add(cost.Amount)
		if existing, exists := breakdown.Dimensions[cost.Dimension]; exists {
			breakdown.Dimensions[cost.Dimension] = existing.Add(cost.Amount)
		} else {
			breakdown.Dimensions[cost.Dimension] = cost.Amount
		}
	}

	// Build trend if requested
	if req.IncludeTrend {
		breakdown.Trend = s.buildCostTrend(costs, req.StartDate, req.EndDate)
	}

	return breakdown
}

// buildCostBreakdownFromAllocations builds a cost breakdown from allocation results
// Uses DirectAmount to get only the node's own costs (not including costs allocated TO it)
func (s *Service) buildCostBreakdownFromAllocations(allocations []models.AllocationResultByDimension, req CostAttributionRequest) *CostBreakdown {
	breakdown := &CostBreakdown{
		Total:      decimal.Zero,
		Currency:   req.Currency,
		Dimensions: make(map[string]decimal.Decimal),
		Trend:      []DailyCostPoint{},
	}

	if breakdown.Currency == "" {
		breakdown.Currency = "USD"
	}

	// Aggregate by dimension using DirectAmount (node's own costs)
	for _, alloc := range allocations {
		breakdown.Total = breakdown.Total.Add(alloc.DirectAmount)
		if existing, exists := breakdown.Dimensions[alloc.Dimension]; exists {
			breakdown.Dimensions[alloc.Dimension] = existing.Add(alloc.DirectAmount)
		} else {
			breakdown.Dimensions[alloc.Dimension] = alloc.DirectAmount
		}
	}

	// Build trend if requested
	if req.IncludeTrend {
		breakdown.Trend = s.buildCostTrendFromAllocations(allocations, req.StartDate, req.EndDate, true)
	}

	return breakdown
}

// buildCostTrendFromAllocations builds daily cost trend data from allocations
// If useDirect is true, uses DirectAmount; otherwise uses TotalAmount
func (s *Service) buildCostTrendFromAllocations(allocations []models.AllocationResultByDimension, startDate, endDate time.Time, useDirect bool) []DailyCostPoint {
	// Group costs by date
	dailyCosts := make(map[time.Time]decimal.Decimal)
	for _, alloc := range allocations {
		amount := alloc.TotalAmount
		if useDirect {
			amount = alloc.DirectAmount
		}
		if existing, exists := dailyCosts[alloc.AllocationDate]; exists {
			dailyCosts[alloc.AllocationDate] = existing.Add(amount)
		} else {
			dailyCosts[alloc.AllocationDate] = amount
		}
	}

	// Build trend points
	var trend []DailyCostPoint
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		cost := decimal.Zero
		if c, exists := dailyCosts[date]; exists {
			cost = c
		}
		trend = append(trend, DailyCostPoint{
			Date: date,
			Cost: cost,
		})
	}

	return trend
}

// buildCostTrend builds daily cost trend data
func (s *Service) buildCostTrend(costs []models.NodeCostByDimension, startDate, endDate time.Time) []DailyCostPoint {
	// Group costs by date
	dailyCosts := make(map[time.Time]decimal.Decimal)
	for _, cost := range costs {
		if existing, exists := dailyCosts[cost.CostDate]; exists {
			dailyCosts[cost.CostDate] = existing.Add(cost.Amount)
		} else {
			dailyCosts[cost.CostDate] = cost.Amount
		}
	}

	// Build trend points
	var trend []DailyCostPoint
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		cost := decimal.Zero
		if dailyCost, exists := dailyCosts[date]; exists {
			cost = dailyCost
		}
		trend = append(trend, DailyCostPoint{
			Date: date,
			Cost: cost,
		})
	}

	return trend
}

// getNodeAllocatedCosts retrieves costs allocated to this node from other nodes
func (s *Service) getNodeAllocatedCosts(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) (*CostBreakdown, error) {
	// Get allocation results where this node is the target (receiving allocations)
	allocations, err := s.store.Runs.GetAllocationsByParentAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocations: %w", err)
	}

	// Convert allocations to cost breakdown format using indirect amounts
	var costs []models.NodeCostByDimension
	for _, allocation := range allocations {
		costs = append(costs, models.NodeCostByDimension{
			NodeID:    allocation.NodeID,
			CostDate:  allocation.AllocationDate,
			Dimension: allocation.Dimension,
			Amount:    allocation.IndirectAmount,
			Currency:  "USD", // TODO: Get from config
		})
	}

	return s.buildCostBreakdown(costs, req), nil
}

// getNodeSharedServiceCosts retrieves costs allocated from shared services specifically
func (s *Service) getNodeSharedServiceCosts(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) (*CostBreakdown, error) {
	// Get contributions where this node is the parent and child is a shared service
	contributions, err := s.store.Runs.GetContributionsByParentAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions: %w", err)
	}

	// Filter for shared services only
	var sharedCosts []models.NodeCostByDimension
	for _, contribution := range contributions {
		// Check if child is a shared service
		childNode, err := s.store.Nodes.GetByID(ctx, contribution.ChildID)
		if err != nil {
			continue // Skip if we can't get the child node
		}

		if childNode.Type == string(models.NodeTypeShared) || childNode.IsPlatform {
			sharedCosts = append(sharedCosts, models.NodeCostByDimension{
				NodeID:    contribution.ParentID,
				CostDate:  contribution.ContributionDate,
				Dimension: contribution.Dimension,
				Amount:    contribution.ContributedAmount,
				Currency:  "USD", // TODO: Get from config
			})
		}
	}

	return s.buildCostBreakdown(sharedCosts, req), nil
}

// combineCostBreakdowns combines two cost breakdowns
func (s *Service) combineCostBreakdowns(breakdown1, breakdown2 *CostBreakdown) *CostBreakdown {
	combined := &CostBreakdown{
		Total:      breakdown1.Total.Add(breakdown2.Total),
		Currency:   breakdown1.Currency,
		Dimensions: make(map[string]decimal.Decimal),
		Trend:      []DailyCostPoint{},
	}

	// Combine dimensions
	for dim, amount := range breakdown1.Dimensions {
		combined.Dimensions[dim] = amount
	}
	for dim, amount := range breakdown2.Dimensions {
		if existing, exists := combined.Dimensions[dim]; exists {
			combined.Dimensions[dim] = existing.Add(amount)
		} else {
			combined.Dimensions[dim] = amount
		}
	}

	// Combine trends (if both have trends)
	if len(breakdown1.Trend) > 0 && len(breakdown2.Trend) > 0 {
		trendMap := make(map[time.Time]decimal.Decimal)

		// Add first breakdown's trend
		for _, point := range breakdown1.Trend {
			trendMap[point.Date] = point.Cost
		}

		// Add second breakdown's trend
		for _, point := range breakdown2.Trend {
			if existing, exists := trendMap[point.Date]; exists {
				trendMap[point.Date] = existing.Add(point.Cost)
			} else {
				trendMap[point.Date] = point.Cost
			}
		}

		// Convert back to slice
		for date, cost := range trendMap {
			combined.Trend = append(combined.Trend, DailyCostPoint{
				Date: date,
				Cost: cost,
			})
		}
	}

	return combined
}

// getNodeDependencies retrieves node dependencies
func (s *Service) getNodeDependencies(ctx context.Context, nodeID uuid.UUID) ([]NodeDependency, error) {
	var dependencies []NodeDependency

	// Get parent edges (nodes that depend on this node)
	parentEdges, err := s.store.Edges.GetByChildID(ctx, nodeID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent edges: %w", err)
	}

	for _, edge := range parentEdges {
		parent, err := s.store.Nodes.GetByID(ctx, edge.ParentID)
		if err != nil {
			continue // Skip if we can't get the parent node
		}

		dependencies = append(dependencies, NodeDependency{
			ID:               parent.ID,
			Name:             parent.Name,
			Type:             parent.Type,
			RelationshipType: "parent",
			Strategy:         edge.DefaultStrategy,
		})
	}

	// Get child edges (nodes this node depends on)
	childEdges, err := s.store.Edges.GetByParentID(ctx, nodeID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get child edges: %w", err)
	}

	for _, edge := range childEdges {
		child, err := s.store.Nodes.GetByID(ctx, edge.ChildID)
		if err != nil {
			continue // Skip if we can't get the child node
		}

		dependencies = append(dependencies, NodeDependency{
			ID:               child.ID,
			Name:             child.Name,
			Type:             child.Type,
			RelationshipType: "child",
			Strategy:         edge.DefaultStrategy,
		})
	}

	return dependencies, nil
}

// getNodeAllocations retrieves allocation details for a node
func (s *Service) getNodeAllocations(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) ([]AllocationDetail, error) {
	var allocations []AllocationDetail

	// Get contributions where this node is the parent (receiving contributions)
	parentContributions, err := s.store.Runs.GetContributionsByParentAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent contributions: %w", err)
	}

	for _, contribution := range parentContributions {
		// Get the child node details
		childNode, err := s.store.Nodes.GetByID(ctx, contribution.ChildID)
		if err != nil {
			continue // Skip if we can't get the child node
		}

		allocations = append(allocations, AllocationDetail{
			FromNode: NodeReference{
				ID:   childNode.ID,
				Name: childNode.Name,
				Type: childNode.Type,
			},
			ToNode: NodeReference{
				ID:   nodeID,
				Name: "", // Will be filled by caller
				Type: "", // Will be filled by caller
			},
			Amount:         contribution.ContributedAmount,
			Currency:       "USD", // TODO: Get from config
			Dimension:      contribution.Dimension,
			Strategy:       "", // TODO: Get from edge strategy
			AllocationDate: contribution.ContributionDate,
		})
	}

	// Get contributions where this node is the child (contributing to others)
	childContributions, err := s.store.Runs.GetContributionsByChildAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get child contributions: %w", err)
	}

	for _, contribution := range childContributions {
		// Get the parent node details
		parentNode, err := s.store.Nodes.GetByID(ctx, contribution.ParentID)
		if err != nil {
			continue // Skip if we can't get the parent node
		}

		allocations = append(allocations, AllocationDetail{
			FromNode: NodeReference{
				ID:   nodeID,
				Name: "", // Will be filled by caller
				Type: "", // Will be filled by caller
			},
			ToNode: NodeReference{
				ID:   parentNode.ID,
				Name: parentNode.Name,
				Type: parentNode.Type,
			},
			Amount:         contribution.ContributedAmount,
			Currency:       "USD", // TODO: Get from config
			Dimension:      contribution.Dimension,
			Strategy:       "", // TODO: Get from edge strategy
			AllocationDate: contribution.ContributionDate,
		})
	}

	return allocations, nil
}

// getProductChildren retrieves child nodes for a product with their contribution amounts
func (s *Service) getProductChildren(ctx context.Context, productID uuid.UUID, req CostAttributionRequest) ([]ProductNode, error) {
	// Get contributions where this product is the parent (receiving contributions from children)
	contributions, err := s.store.Runs.GetContributionsByParentAndDateRange(ctx, productID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions: %w", err)
	}

	// Group contributions by child node
	childContributions := make(map[uuid.UUID][]models.ContributionResultByDimension)
	for _, contribution := range contributions {
		childContributions[contribution.ChildID] = append(childContributions[contribution.ChildID], contribution)
	}

	var children []ProductNode
	for childID, contribs := range childContributions {
		// Get child node
		childNode, err := s.store.Nodes.GetByID(ctx, childID)
		if err != nil {
			continue // Skip if we can't get the child node
		}

		// Build child node with contribution amounts (not full direct costs)
		child, err := s.buildChildNodeWithContributions(ctx, *childNode, contribs, req)
		if err != nil {
			continue // Skip if we can't build the child node
		}

		children = append(children, *child)
	}

	return children, nil
}

// buildChildNodeWithContributions builds a child node showing only the costs contributed to the parent
func (s *Service) buildChildNodeWithContributions(ctx context.Context, node models.CostNode, contributions []models.ContributionResultByDimension, req CostAttributionRequest) (*ProductNode, error) {
	// Convert contributions to cost breakdown format
	var costs []models.NodeCostByDimension
	for _, contribution := range contributions {
		costs = append(costs, models.NodeCostByDimension{
			NodeID:    contribution.ChildID,
			CostDate:  contribution.ContributionDate,
			Dimension: contribution.Dimension,
			Amount:    contribution.ContributedAmount,
			Currency:  "USD", // TODO: Get from config
		})
	}

	contributedCosts := s.buildCostBreakdown(costs, req)

	// For children, we show the contributed amount as both direct and holistic
	// since we're showing what this child contributed to the parent
	return &ProductNode{
		ID:                 node.ID,
		Name:               node.Name,
		Type:               node.Type,
		DirectCosts:        *contributedCosts,
		HolisticCosts:      *contributedCosts,
		SharedServiceCosts: CostBreakdown{Total: decimal.Zero, Currency: req.Currency, Dimensions: make(map[string]decimal.Decimal)},
		Children:           []ProductNode{}, // Don't recurse further for children
		Metadata:           node.Metadata,
	}, nil
}

// buildPlatformService builds a platform service with its cost data
func (s *Service) buildPlatformService(ctx context.Context, node models.CostNode, req CostAttributionRequest) (*PlatformService, error) {
	// Get direct costs
	directCosts, err := s.getNodeDirectCosts(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct costs: %w", err)
	}

	// Get allocation targets (where this platform service allocates costs to)
	allocatedTo, err := s.getPlatformAllocationTargets(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation targets: %w", err)
	}

	return &PlatformService{
		ID:          node.ID,
		Name:        node.Name,
		Type:        node.Type,
		DirectCosts: *directCosts,
		AllocatedTo: allocatedTo,
		Metadata:    node.Metadata,
	}, nil
}

// buildSharedService builds a shared service with weighted allocation
func (s *Service) buildSharedService(ctx context.Context, node models.CostNode, req CostAttributionRequest) (*SharedService, error) {
	// Get direct costs
	directCosts, err := s.getNodeDirectCosts(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct costs: %w", err)
	}

	// Get weighted targets
	weightedTargets, err := s.getSharedServiceWeightedTargets(ctx, node.ID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get weighted targets: %w", err)
	}

	return &SharedService{
		ID:              node.ID,
		Name:            node.Name,
		Type:            node.Type,
		DirectCosts:     *directCosts,
		WeightedTargets: weightedTargets,
		Metadata:        node.Metadata,
	}, nil
}

// getPlatformAllocationTargets gets allocation targets for a platform service
func (s *Service) getPlatformAllocationTargets(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) ([]AllocationTarget, error) {
	// Get contributions where this node is the child (contributing to others)
	contributions, err := s.store.Runs.GetContributionsByChildAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions: %w", err)
	}

	// Group by parent node
	targetMap := make(map[uuid.UUID]*AllocationTarget)
	totalAmount := decimal.Zero

	for _, contribution := range contributions {
		if target, exists := targetMap[contribution.ParentID]; exists {
			target.Amount = target.Amount.Add(contribution.ContributedAmount)
		} else {
			// Get parent node details
			parentNode, err := s.store.Nodes.GetByID(ctx, contribution.ParentID)
			if err != nil {
				continue // Skip if we can't get the parent node
			}

			targetMap[contribution.ParentID] = &AllocationTarget{
				NodeID:   contribution.ParentID,
				NodeName: parentNode.Name,
				NodeType: parentNode.Type,
				Amount:   contribution.ContributedAmount,
				Currency: "USD", // TODO: Get from config
			}
		}
		totalAmount = totalAmount.Add(contribution.ContributedAmount)
	}

	// Calculate percentages and convert to slice
	var targets []AllocationTarget
	for _, target := range targetMap {
		if !totalAmount.IsZero() {
			target.Percentage = target.Amount.Div(totalAmount).InexactFloat64() * 100
		}
		targets = append(targets, *target)
	}

	return targets, nil
}

// getSharedServiceWeightedTargets gets weighted targets for a shared service
func (s *Service) getSharedServiceWeightedTargets(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) ([]WeightedTarget, error) {
	// Get contributions where this node is the child (contributing to others)
	contributions, err := s.store.Runs.GetContributionsByChildAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions: %w", err)
	}

	// Group by parent node
	targetMap := make(map[uuid.UUID]*WeightedTarget)
	totalAmount := decimal.Zero

	for _, contribution := range contributions {
		if target, exists := targetMap[contribution.ParentID]; exists {
			target.Amount = target.Amount.Add(contribution.ContributedAmount)
		} else {
			// Get parent node details
			parentNode, err := s.store.Nodes.GetByID(ctx, contribution.ParentID)
			if err != nil {
				continue // Skip if we can't get the parent node
			}

			// TODO: Get actual weight from edge strategy parameters
			weight := decimal.NewFromFloat(1.0) // Default weight

			targetMap[contribution.ParentID] = &WeightedTarget{
				NodeID:   contribution.ParentID,
				NodeName: parentNode.Name,
				NodeType: parentNode.Type,
				Weight:   weight,
				Amount:   contribution.ContributedAmount,
				Currency: "USD", // TODO: Get from config
			}
		}
		totalAmount = totalAmount.Add(contribution.ContributedAmount)
	}

	// Calculate percentages and convert to slice
	var targets []WeightedTarget
	for _, target := range targetMap {
		if !totalAmount.IsZero() {
			target.Percentage = target.Amount.Div(totalAmount).InexactFloat64() * 100
		}
		targets = append(targets, *target)
	}

	return targets, nil
}

// getWeightedAllocations gets all weighted allocations from shared services
func (s *Service) getWeightedAllocations(ctx context.Context, req CostAttributionRequest) ([]WeightedAllocation, error) {
	// Get all shared service nodes
	sharedNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		Type: string(models.NodeTypeShared),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get shared nodes: %w", err)
	}

	var allocations []WeightedAllocation
	for _, sharedNode := range sharedNodes {
		// Get contributions from this shared service
		contributions, err := s.store.Runs.GetContributionsByChildAndDateRange(ctx, sharedNode.ID, req.StartDate, req.EndDate, req.Dimensions)
		if err != nil {
			continue // Skip if we can't get contributions
		}

		for _, contribution := range contributions {
			// Get target node details
			targetNode, err := s.store.Nodes.GetByID(ctx, contribution.ParentID)
			if err != nil {
				continue // Skip if we can't get the target node
			}

			// TODO: Get actual weight and strategy from edge
			weight := decimal.NewFromFloat(1.0) // Default weight
			strategy := "equal"                 // Default strategy

			allocations = append(allocations, WeightedAllocation{
				SharedServiceID:   sharedNode.ID,
				SharedServiceName: sharedNode.Name,
				TargetNodeID:      targetNode.ID,
				TargetNodeName:    targetNode.Name,
				Weight:            weight,
				Amount:            contribution.ContributedAmount,
				Currency:          "USD", // TODO: Get from config
				Dimension:         contribution.Dimension,
				Strategy:          strategy,
			})
		}
	}

	return allocations, nil
}

// calculatePlatformSummary calculates summary for platform services
func (s *Service) calculatePlatformSummary(ctx context.Context, req CostAttributionRequest, platformCount, sharedCount int) (*CostSummary, error) {
	// Get total costs for platform and shared services
	platformNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		IsPlatform: &[]bool{true}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get platform nodes: %w", err)
	}

	sharedNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		Type: string(models.NodeTypeShared),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get shared nodes: %w", err)
	}

	totalCost := decimal.Zero
	nodeCount := 0

	// Calculate total costs for platform nodes
	for _, node := range platformNodes {
		costs, err := s.store.Costs.GetByNodeAndDateRange(ctx, node.ID, req.StartDate, req.EndDate, req.Dimensions)
		if err != nil {
			continue // Skip if we can't get costs
		}
		for _, cost := range costs {
			totalCost = totalCost.Add(cost.Amount)
		}
		nodeCount++
	}

	// Calculate total costs for shared nodes
	for _, node := range sharedNodes {
		costs, err := s.store.Costs.GetByNodeAndDateRange(ctx, node.ID, req.StartDate, req.EndDate, req.Dimensions)
		if err != nil {
			continue // Skip if we can't get costs
		}
		for _, cost := range costs {
			totalCost = totalCost.Add(cost.Amount)
		}
		nodeCount++
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	return &CostSummary{
		TotalCost:         totalCost,
		Currency:          currency,
		Period:            fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		NodeCount:         nodeCount,
		PlatformNodeCount: platformCount,
	}, nil
}


// GetRecommendations retrieves cost optimization recommendations
func (s *Service) GetRecommendations(ctx context.Context, req CostAttributionRequest, nodeID *uuid.UUID) (*RecommendationsResponse, error) {
	var recommendations []models.CostRecommendation
	var err error

	if nodeID != nil {
		// Get recommendations for specific node
		recommendations, err = s.recommendationAnalyzer.AnalyzeNode(ctx, *nodeID, req.StartDate, req.EndDate)
	} else {
		// Get recommendations for all nodes
		recommendations, err = s.recommendationAnalyzer.AnalyzeAllNodes(ctx, req.StartDate, req.EndDate)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to analyze recommendations: %w", err)
	}

	// Convert to API models and count by severity
	apiRecommendations := make([]CostRecommendation, len(recommendations))
	highCount := 0
	mediumCount := 0
	lowCount := 0
	totalSavings := decimal.Zero

	for i, rec := range recommendations {
		apiRecommendations[i] = CostRecommendation{
			ID:                 rec.ID,
			NodeID:             rec.NodeID,
			NodeName:           rec.NodeName,
			NodeType:           rec.NodeType,
			Type:               string(rec.Type),
			Severity:           string(rec.Severity),
			Title:              rec.Title,
			Description:        rec.Description,
			CurrentCost:        rec.CurrentCost,
			PotentialSavings:   rec.PotentialSavings,
			Currency:           rec.Currency,
			Metric:             rec.Metric,
			CurrentValue:       rec.CurrentValue,
			PeakValue:          rec.PeakValue,
			AverageValue:       rec.AverageValue,
			UtilizationPercent: rec.UtilizationPercent,
			RecommendedAction:  rec.RecommendedAction,
			AnalysisPeriod:     rec.AnalysisPeriod,
			StartDate:          rec.StartDate,
			EndDate:            rec.EndDate,
			CreatedAt:          rec.CreatedAt,
		}

		totalSavings = totalSavings.Add(rec.PotentialSavings)

		switch rec.Severity {
		case models.RecommendationSeverityHigh:
			highCount++
		case models.RecommendationSeverityMedium:
			mediumCount++
		case models.RecommendationSeverityLow:
			lowCount++
		}
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	return &RecommendationsResponse{
		Recommendations:     apiRecommendations,
		TotalSavings:        totalSavings,
		Currency:            currency,
		HighSeverityCount:   highCount,
		MediumSeverityCount: mediumCount,
		LowSeverityCount:    lowCount,
	}, nil
}



// buildUnallocatedNode creates a synthetic node representing unallocated platform/shared costs
func (s *Service) buildUnallocatedNode(ctx context.Context, req CostAttributionRequest, unallocatedCost decimal.Decimal) (*ProductNode, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Get platform and shared nodes to show as children
	platformNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		IsPlatform: &[]bool{true}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get platform nodes: %w", err)
	}

	sharedNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{
		Type: string(models.NodeTypeShared),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get shared nodes: %w", err)
	}

	// Build children showing unallocated portions
	children := make([]ProductNode, 0)

	// Add platform nodes
	for _, node := range platformNodes {
		directCosts, err := s.getNodeDirectCosts(ctx, node.ID, req)
		if err != nil {
			continue
		}

		// Get how much was allocated to products
		allocatedToProducts, err := s.getNodeAllocatedToProducts(ctx, node.ID, req)
		if err != nil {
			continue
		}

		// Calculate unallocated portion
		unallocatedPortion := directCosts.Total.Sub(allocatedToProducts)

		if unallocatedPortion.GreaterThan(decimal.Zero) {
			children = append(children, ProductNode{
				ID:   node.ID,
				Name: node.Name,
				Type: node.Type,
				DirectCosts: CostBreakdown{
					Total:      unallocatedPortion,
					Currency:   currency,
					Dimensions: map[string]decimal.Decimal{},
				},
				HolisticCosts: CostBreakdown{
					Total:      unallocatedPortion,
					Currency:   currency,
					Dimensions: map[string]decimal.Decimal{},
				},
				SharedServiceCosts: CostBreakdown{
					Total:      decimal.Zero,
					Currency:   currency,
					Dimensions: map[string]decimal.Decimal{},
				},
				Metadata: node.Metadata,
			})
		}
	}

	// Add shared nodes
	for _, node := range sharedNodes {
		directCosts, err := s.getNodeDirectCosts(ctx, node.ID, req)
		if err != nil {
			continue
		}

		// Get how much was allocated to products
		allocatedToProducts, err := s.getNodeAllocatedToProducts(ctx, node.ID, req)
		if err != nil {
			continue
		}

		// Calculate unallocated portion
		unallocatedPortion := directCosts.Total.Sub(allocatedToProducts)

		if unallocatedPortion.GreaterThan(decimal.Zero) {
			children = append(children, ProductNode{
				ID:   node.ID,
				Name: node.Name,
				Type: node.Type,
				DirectCosts: CostBreakdown{
					Total:      unallocatedPortion,
					Currency:   currency,
					Dimensions: map[string]decimal.Decimal{},
				},
				HolisticCosts: CostBreakdown{
					Total:      unallocatedPortion,
					Currency:   currency,
					Dimensions: map[string]decimal.Decimal{},
				},
				SharedServiceCosts: CostBreakdown{
					Total:      decimal.Zero,
					Currency:   currency,
					Dimensions: map[string]decimal.Decimal{},
				},
				Metadata: node.Metadata,
			})
		}
	}

	return &ProductNode{
		ID:   uuid.New(),
		Name: "Unallocated Platform & Shared Costs",
		Type: "unallocated",
		DirectCosts: CostBreakdown{
			Total:      unallocatedCost,
			Currency:   currency,
			Dimensions: map[string]decimal.Decimal{},
		},
		HolisticCosts: CostBreakdown{
			Total:      unallocatedCost,
			Currency:   currency,
			Dimensions: map[string]decimal.Decimal{},
		},
		SharedServiceCosts: CostBreakdown{
			Total:      decimal.Zero,
			Currency:   currency,
			Dimensions: map[string]decimal.Decimal{},
		},
		Children: children,
		Metadata: map[string]interface{}{
			"description": "Platform and shared service costs that are not allocated to any product",
		},
	}, nil
}

// getNodeAllocatedToProducts calculates how much of a node's cost was allocated to products
func (s *Service) getNodeAllocatedToProducts(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) (decimal.Decimal, error) {
	// Get all contributions where this node is the child (contributing to products)
	contributions, err := s.store.Runs.GetContributionsByChildAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get contributions: %w", err)
	}

	total := decimal.Zero
	for _, contrib := range contributions {
		// Only count contributions to product nodes
		parentNode, err := s.store.Nodes.GetByID(ctx, contrib.ParentID)
		if err != nil {
			continue
		}
		if parentNode.Type == string(models.NodeTypeProduct) {
			total = total.Add(contrib.ContributedAmount)
		}
	}

	return total, nil
}

// GetAllocationReconciliation provides debug information for allocation reconciliation
// This endpoint helps diagnose allocation issues by showing:
// - Raw infrastructure cost vs allocated product cost
// - Final cost centre breakdown
// - Infrastructure node breakdown
// - Invariant violations
func (s *Service) GetAllocationReconciliation(ctx context.Context, req CostAttributionRequest) (*AllocationReconciliationResponse, error) {
	// Build graph for the date range
	g, err := s.graphBuilder.BuildForDate(ctx, req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	// Get raw infrastructure cost
	rawInfraCost, err := s.store.Costs.GetTotalInfrastructureCostByDateRange(ctx, req.StartDate, req.EndDate, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw infrastructure cost: %w", err)
	}

	// Get final cost centres
	finalCostCentreIDs := g.GetFinalCostCentres()
	finalCostCentreSet := make(map[uuid.UUID]bool)
	for _, id := range finalCostCentreIDs {
		finalCostCentreSet[id] = true
	}

	// Calculate allocated product cost (sum of final cost centre holistic costs)
	allocatedProductCost := decimal.Zero
	var finalCostCentres []FinalCostCentreDetail

	for _, nodeID := range finalCostCentreIDs {
		node, exists := g.Nodes()[nodeID]
		if !exists {
			continue
		}

		// Get allocation results for this node
		allocations, err := s.store.Runs.GetAllocationsByNodeAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
		if err != nil {
			continue
		}

		var directCost, indirectCost decimal.Decimal
		for _, alloc := range allocations {
			directCost = directCost.Add(alloc.DirectAmount)
			indirectCost = indirectCost.Add(alloc.IndirectAmount)
		}
		holisticCost := directCost.Add(indirectCost)
		allocatedProductCost = allocatedProductCost.Add(holisticCost)

		finalCostCentres = append(finalCostCentres, FinalCostCentreDetail{
			ID:           nodeID,
			Name:         node.Name,
			HolisticCost: holisticCost,
			DirectCost:   directCost,
			IndirectCost: indirectCost,
		})
	}

	// Get infrastructure nodes
	var infraNodes []InfrastructureNodeDetail
	infraNodeIDs := g.GetInfrastructureNodes()
	for _, nodeID := range infraNodeIDs {
		node, exists := g.Nodes()[nodeID]
		if !exists {
			continue
		}

		// Get direct costs for this node
		costs, err := s.store.Costs.GetByNodeAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
		if err != nil {
			continue
		}

		var directCost decimal.Decimal
		for _, cost := range costs {
			directCost = directCost.Add(cost.Amount)
		}

		infraNodes = append(infraNodes, InfrastructureNodeDetail{
			ID:         nodeID,
			Name:       node.Name,
			Type:       node.Type,
			DirectCost: directCost,
			IsPlatform: node.IsPlatform,
		})
	}

	// Calculate unallocated cost
	unallocatedCost := rawInfraCost.Sub(allocatedProductCost)
	if unallocatedCost.LessThan(decimal.Zero) {
		unallocatedCost = decimal.Zero
	}

	// Calculate coverage
	coveragePercent := 0.0
	if !rawInfraCost.IsZero() {
		ratio, _ := allocatedProductCost.Div(rawInfraCost).Float64()
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		coveragePercent = ratio * 100
	}

	// Check conservation invariant
	conservationDelta := rawInfraCost.Sub(allocatedProductCost.Add(unallocatedCost))
	conservationValid := conservationDelta.Abs().LessThan(decimal.NewFromFloat(0.01))

	// Check for invariant violations
	var violations []InvariantViolation

	// Check if allocated > raw (amplification)
	if allocatedProductCost.GreaterThan(rawInfraCost) {
		violations = append(violations, InvariantViolation{
			Type:        "amplification",
			Description: "Allocated product cost exceeds raw infrastructure cost",
			Expected:    rawInfraCost.StringFixed(2),
			Actual:      allocatedProductCost.StringFixed(2),
		})
	}

	// Check conservation
	if !conservationValid {
		violations = append(violations, InvariantViolation{
			Type:        "conservation",
			Description: "Conservation invariant violated: Raw != Allocated + Unallocated",
			Expected:    rawInfraCost.StringFixed(2),
			Actual:      allocatedProductCost.Add(unallocatedCost).StringFixed(2),
		})
	}

	// Build graph statistics
	stats := g.Stats()
	graphStats := GraphStatistics{
		TotalNodes:          stats.NodeCount,
		ProductNodes:        len(g.GetProductNodes()),
		InfrastructureNodes: len(infraNodeIDs),
		FinalCostCentres:    len(finalCostCentreIDs),
		TotalEdges:          stats.EdgeCount,
		MaxDepth:            stats.MaxDepth,
	}

	return &AllocationReconciliationResponse{
		Period:                fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		Currency:             req.Currency,
		RawInfrastructureCost: rawInfraCost,
		AllocatedProductCost:  allocatedProductCost,
		UnallocatedCost:       unallocatedCost,
		CoveragePercent:       coveragePercent,
		ConservationDelta:     conservationDelta,
		ConservationValid:     conservationValid,
		FinalCostCentres:      finalCostCentres,
		InfrastructureNodes:   infraNodes,
		InvariantViolations:   violations,
		GraphStats:            graphStats,
	}, nil
}

// performProductTreeSanityChecks validates that product tree totals are consistent
// with the calculated allocated product cost. Returns true if all checks pass,
// along with any warning messages.
func (s *Service) performProductTreeSanityChecks(
	products []ProductNode,
	allocatedProductCost decimal.Decimal,
	rawInfraCost decimal.Decimal,
	unallocatedCost decimal.Decimal,
	finalCostCentreSet map[uuid.UUID]bool,
) (bool, []string) {
	var warnings []string
	allPassed := true

	// Tolerance threshold: 1% of raw infrastructure cost or $0.01, whichever is larger
	tolerance := rawInfraCost.Mul(decimal.NewFromFloat(0.01))
	minTolerance := decimal.NewFromFloat(0.01)
	if tolerance.LessThan(minTolerance) {
		tolerance = minTolerance
	}

	// Check 1: Sum of final cost centre holistic costs should equal allocated product cost
	finalCostCentreSum := decimal.Zero
	for _, product := range products {
		if finalCostCentreSet[product.ID] {
			finalCostCentreSum = finalCostCentreSum.Add(product.HolisticCosts.Total)
		}
	}

	diff := finalCostCentreSum.Sub(allocatedProductCost).Abs()
	if diff.GreaterThan(tolerance) {
		allPassed = false
		warnings = append(warnings, fmt.Sprintf(
			"Final cost centre sum (%s) differs from allocated product cost (%s) by %s (tolerance: %s)",
			finalCostCentreSum.StringFixed(2),
			allocatedProductCost.StringFixed(2),
			diff.StringFixed(2),
			tolerance.StringFixed(2),
		))
		log.Warn().
			Str("final_cost_centre_sum", finalCostCentreSum.StringFixed(2)).
			Str("allocated_product_cost", allocatedProductCost.StringFixed(2)).
			Str("difference", diff.StringFixed(2)).
			Str("tolerance", tolerance.StringFixed(2)).
			Msg("SANITY CHECK FAILED: Final cost centre sum differs from allocated product cost")
	}

	// Check 2: Allocated + Unallocated should equal Raw Infrastructure Cost (conservation)
	totalCost := allocatedProductCost.Add(unallocatedCost)
	conservationDiff := totalCost.Sub(rawInfraCost).Abs()
	if conservationDiff.GreaterThan(tolerance) {
		allPassed = false
		warnings = append(warnings, fmt.Sprintf(
			"Conservation violated: Allocated (%s) + Unallocated (%s) = %s differs from Raw Infra (%s) by %s",
			allocatedProductCost.StringFixed(2),
			unallocatedCost.StringFixed(2),
			totalCost.StringFixed(2),
			rawInfraCost.StringFixed(2),
			conservationDiff.StringFixed(2),
		))
		log.Warn().
			Str("allocated", allocatedProductCost.StringFixed(2)).
			Str("unallocated", unallocatedCost.StringFixed(2)).
			Str("total", totalCost.StringFixed(2)).
			Str("raw_infra", rawInfraCost.StringFixed(2)).
			Str("difference", conservationDiff.StringFixed(2)).
			Msg("SANITY CHECK FAILED: Conservation invariant violated")
	}

	// Check 3: Allocated product cost should not exceed raw infrastructure cost (no amplification)
	if allocatedProductCost.GreaterThan(rawInfraCost.Add(tolerance)) {
		allPassed = false
		amplification := allocatedProductCost.Sub(rawInfraCost)
		warnings = append(warnings, fmt.Sprintf(
			"Amplification detected: Allocated product cost (%s) exceeds raw infrastructure cost (%s) by %s",
			allocatedProductCost.StringFixed(2),
			rawInfraCost.StringFixed(2),
			amplification.StringFixed(2),
		))
		log.Warn().
			Str("allocated", allocatedProductCost.StringFixed(2)).
			Str("raw_infra", rawInfraCost.StringFixed(2)).
			Str("amplification", amplification.StringFixed(2)).
			Msg("SANITY CHECK FAILED: Cost amplification detected")
	}

	// Check 4: No product should have negative holistic cost
	for _, product := range products {
		if product.HolisticCosts.Total.LessThan(decimal.Zero) {
			allPassed = false
			warnings = append(warnings, fmt.Sprintf(
				"Product '%s' has negative holistic cost: %s",
				product.Name,
				product.HolisticCosts.Total.StringFixed(2),
			))
			log.Warn().
				Str("product_id", product.ID.String()).
				Str("product_name", product.Name).
				Str("holistic_cost", product.HolisticCosts.Total.StringFixed(2)).
				Msg("SANITY CHECK FAILED: Negative holistic cost detected")
		}
	}

	if allPassed {
		log.Debug().
			Str("allocated_product_cost", allocatedProductCost.StringFixed(2)).
			Str("raw_infra_cost", rawInfraCost.StringFixed(2)).
			Int("final_cost_centres", len(finalCostCentreSet)).
			Msg("Product tree sanity checks passed")
	}

	return allPassed, warnings
}

// GetDashboardSummary returns the correct totals for the dashboard
// This uses final cost centres to calculate the true "Total Product Cost" without double-counting
func (s *Service) GetDashboardSummary(ctx context.Context, req CostAttributionRequest) (*DashboardSummaryResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Build graph to identify final cost centres
	g, err := s.graphBuilder.BuildForDate(ctx, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph for final cost centre detection: %w", err)
	}

	// Get final cost centres (product nodes with no outgoing product edges)
	finalCostCentreIDs := g.GetFinalCostCentres()
	finalCostCentreSet := make(map[uuid.UUID]bool)
	for _, id := range finalCostCentreIDs {
		finalCostCentreSet[id] = true
	}

	// Get all products with their holistic costs using the existing buildProductTree method
	products, err := s.buildProductTree(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get products with holistic costs: %w", err)
	}

	// Calculate total product cost from final cost centres only
	totalProductCost := decimal.Zero
	for _, product := range products {
		if finalCostCentreSet[product.ID] {
			totalProductCost = totalProductCost.Add(product.HolisticCosts.Total)
		}
	}

	// Get raw infrastructure cost
	rawInfraCost, err := s.store.Costs.GetTotalInfrastructureCostByDateRange(ctx, req.StartDate, req.EndDate, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw infrastructure cost: %w", err)
	}

	// Calculate unallocated cost and coverage
	unallocatedCost := rawInfraCost.Sub(totalProductCost)
	if unallocatedCost.LessThan(decimal.Zero) {
		unallocatedCost = decimal.Zero
	}

	var coveragePercent float64
	if !rawInfraCost.IsZero() {
		coveragePercent, _ = totalProductCost.Div(rawInfraCost).Mul(decimal.NewFromInt(100)).Float64()
		if coveragePercent > 100 {
			coveragePercent = 100
		}
	}

	// Get costs by type for breakdown
	costsByType, err := s.store.Costs.GetCostsByType(ctx, req.StartDate, req.EndDate, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs by type: %w", err)
	}

	// Convert to API response format
	typeAggregations := make([]TypeAggregation, 0, len(costsByType))
	var productCount, platformCount, sharedCount, resourceCount int
	for _, ct := range costsByType {
		typeAggregations = append(typeAggregations, TypeAggregation{
			Type:           ct.Type,
			DirectCost:     ct.DirectCost,
			IndirectCost:   ct.IndirectCost,
			TotalCost:      ct.TotalCost,
			NodeCount:      ct.NodeCount,
			PercentOfTotal: ct.PercentOfTotal,
		})
		switch ct.Type {
		case "product":
			productCount = ct.NodeCount
		case "platform":
			platformCount = ct.NodeCount
		case "shared":
			sharedCount = ct.NodeCount
		case "resource", "infrastructure":
			resourceCount += ct.NodeCount
		}
	}

	return &DashboardSummaryResponse{
		TotalProductCost:          totalProductCost,
		RawInfrastructureCost:     rawInfraCost,
		AllocationCoveragePercent: coveragePercent,
		UnallocatedCost:           unallocatedCost,
		CostsByType:               typeAggregations,
		ProductCount:              productCount,
		FinalCostCentreCount:      len(finalCostCentreIDs),
		PlatformCount:             platformCount,
		SharedCount:               sharedCount,
		ResourceCount:             resourceCount,
		Currency:                  currency,
		Period:                    fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		StartDate:                 req.StartDate,
		EndDate:                   req.EndDate,
	}, nil
}

// GetInfrastructureHierarchy retrieves the infrastructure hierarchy with cost and allocation data
// This shows platform, shared, and resource nodes with their direct costs and how much has been allocated to products
func (s *Service) GetInfrastructureHierarchy(ctx context.Context, req CostAttributionRequest) (*InfrastructureHierarchyResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Get all infrastructure-type nodes
	platformNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{IsPlatform: &[]bool{true}[0]})
	if err != nil {
		return nil, fmt.Errorf("failed to get platform nodes: %w", err)
	}

	sharedNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{Type: string(models.NodeTypeShared)})
	if err != nil {
		return nil, fmt.Errorf("failed to get shared nodes: %w", err)
	}

	resourceNodes, err := s.store.Nodes.List(ctx, store.NodeFilters{Type: string(models.NodeTypeResource)})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource nodes: %w", err)
	}

	// Build infrastructure nodes with cost data
	var infraNodes []InfrastructureNode
	totalDirectCost := decimal.Zero
	totalAllocatedCost := decimal.Zero

	// Process platform nodes
	for _, node := range platformNodes {
		infraNode, err := s.buildInfrastructureNode(ctx, node, req)
		if err != nil {
			log.Warn().Err(err).Str("node", node.Name).Msg("Failed to build infrastructure node")
			continue
		}
		infraNodes = append(infraNodes, *infraNode)
		totalDirectCost = totalDirectCost.Add(infraNode.DirectCosts.Total)
		totalAllocatedCost = totalAllocatedCost.Add(infraNode.AllocatedCosts.Total)
	}

	// Process shared nodes (avoid duplicates with platform)
	platformIDs := make(map[uuid.UUID]bool)
	for _, n := range platformNodes {
		platformIDs[n.ID] = true
	}
	for _, node := range sharedNodes {
		if platformIDs[node.ID] {
			continue // Already processed as platform
		}
		infraNode, err := s.buildInfrastructureNode(ctx, node, req)
		if err != nil {
			log.Warn().Err(err).Str("node", node.Name).Msg("Failed to build infrastructure node")
			continue
		}
		infraNodes = append(infraNodes, *infraNode)
		totalDirectCost = totalDirectCost.Add(infraNode.DirectCosts.Total)
		totalAllocatedCost = totalAllocatedCost.Add(infraNode.AllocatedCosts.Total)
	}

	// Process resource nodes
	for _, node := range resourceNodes {
		if platformIDs[node.ID] {
			continue
		}
		infraNode, err := s.buildInfrastructureNode(ctx, node, req)
		if err != nil {
			log.Warn().Err(err).Str("node", node.Name).Msg("Failed to build infrastructure node")
			continue
		}
		infraNodes = append(infraNodes, *infraNode)
		totalDirectCost = totalDirectCost.Add(infraNode.DirectCosts.Total)
		totalAllocatedCost = totalAllocatedCost.Add(infraNode.AllocatedCosts.Total)
	}

	// Calculate summary
	totalUnallocatedCost := totalDirectCost.Sub(totalAllocatedCost)
	if totalUnallocatedCost.LessThan(decimal.Zero) {
		totalUnallocatedCost = decimal.Zero
	}

	var allocationPct float64
	if !totalDirectCost.IsZero() {
		allocationPct, _ = totalAllocatedCost.Div(totalDirectCost).Mul(decimal.NewFromInt(100)).Float64()
	}

	return &InfrastructureHierarchyResponse{
		Infrastructure: infraNodes,
		Summary: InfraSummary{
			TotalDirectCost:      totalDirectCost,
			TotalAllocatedCost:   totalAllocatedCost,
			TotalUnallocatedCost: totalUnallocatedCost,
			AllocationPct:        allocationPct,
			Currency:             currency,
			Period:               fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
			StartDate:            req.StartDate,
			EndDate:              req.EndDate,
			PlatformCount:        len(platformNodes),
			SharedCount:          len(sharedNodes),
			ResourceCount:        len(resourceNodes),
			TotalNodeCount:       len(infraNodes),
		},
	}, nil
}

// buildInfrastructureNode builds a single infrastructure node with cost and allocation data
func (s *Service) buildInfrastructureNode(ctx context.Context, node models.CostNode, req CostAttributionRequest) (*InfrastructureNode, error) {
	// Get direct costs from node_costs_by_dimension
	directCosts, err := s.store.Costs.GetByNodeAndDateRange(ctx, node.ID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct costs: %w", err)
	}
	directBreakdown := s.buildCostBreakdown(directCosts, req)

	// Get allocations FROM this node TO products
	allocations, err := s.store.Costs.GetAllocationsFromNode(ctx, node.ID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocations: %w", err)
	}

	// Build allocated costs breakdown and allocation targets
	allocatedTotal := decimal.Zero
	allocatedDimensions := make(map[string]decimal.Decimal)
	targetMap := make(map[uuid.UUID]*InfraAllocationTarget)

	for _, alloc := range allocations {
		allocatedTotal = allocatedTotal.Add(alloc.Amount)
		if _, ok := allocatedDimensions[alloc.Dimension]; !ok {
			allocatedDimensions[alloc.Dimension] = decimal.Zero
		}
		allocatedDimensions[alloc.Dimension] = allocatedDimensions[alloc.Dimension].Add(alloc.Amount)

		// Build allocation target
		if target, ok := targetMap[alloc.ChildID]; ok {
			target.Amount = target.Amount.Add(alloc.Amount)
		} else {
			// Get target node info
			targetNode, err := s.store.Nodes.GetByID(ctx, alloc.ChildID)
			if err != nil {
				continue
			}
			targetMap[alloc.ChildID] = &InfraAllocationTarget{
				ID:       targetNode.ID,
				Name:     targetNode.Name,
				Type:     targetNode.Type,
				Amount:   alloc.Amount,
				Currency: req.Currency,
				Strategy: alloc.Strategy,
			}
		}
	}

	// Convert target map to slice and calculate percentages
	var allocatedTo []InfraAllocationTarget
	for _, target := range targetMap {
		if !directBreakdown.Total.IsZero() {
			target.Percent, _ = target.Amount.Div(directBreakdown.Total).Mul(decimal.NewFromInt(100)).Float64()
		}
		allocatedTo = append(allocatedTo, *target)
	}

	// Calculate unallocated cost
	unallocatedCost := directBreakdown.Total.Sub(allocatedTotal)
	if unallocatedCost.LessThan(decimal.Zero) {
		unallocatedCost = decimal.Zero
	}

	var allocationPct float64
	if !directBreakdown.Total.IsZero() {
		allocationPct, _ = allocatedTotal.Div(directBreakdown.Total).Mul(decimal.NewFromInt(100)).Float64()
	}

	return &InfrastructureNode{
		ID:         node.ID,
		Name:       node.Name,
		Type:       node.Type,
		IsPlatform: node.IsPlatform,
		DirectCosts: *directBreakdown,
		AllocatedCosts: CostBreakdown{
			Total:      allocatedTotal,
			Currency:   req.Currency,
			Dimensions: allocatedDimensions,
		},
		UnallocatedCost: unallocatedCost,
		AllocationPct:   allocationPct,
		AllocatedTo:     allocatedTo,
		Metadata:        node.Metadata,
	}, nil
}

// GetNodeMetricsTimeSeries retrieves cost and usage metrics over time for a specific node
func (s *Service) GetNodeMetricsTimeSeries(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) (*NodeMetricsTimeSeriesResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Get node info
	node, err := s.store.Nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get daily costs
	costs, err := s.store.Costs.GetByNodeAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get costs: %w", err)
	}

	// Get daily usage (nil metrics means all metrics)
	usage, err := s.store.Usage.GetByNodeAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	// Build cost series grouped by date
	costByDate := make(map[string]*DailyCostDataPoint)
	dimensionSet := make(map[string]bool)
	for _, c := range costs {
		dateKey := c.CostDate.Format("2006-01-02")
		if _, ok := costByDate[dateKey]; !ok {
			costByDate[dateKey] = &DailyCostDataPoint{
				Date:       c.CostDate,
				TotalCost:  decimal.Zero,
				Dimensions: make(map[string]decimal.Decimal),
			}
		}
		costByDate[dateKey].TotalCost = costByDate[dateKey].TotalCost.Add(c.Amount)
		if _, ok := costByDate[dateKey].Dimensions[c.Dimension]; !ok {
			costByDate[dateKey].Dimensions[c.Dimension] = decimal.Zero
		}
		costByDate[dateKey].Dimensions[c.Dimension] = costByDate[dateKey].Dimensions[c.Dimension].Add(c.Amount)
		dimensionSet[c.Dimension] = true
	}

	// Convert to sorted slice
	var costSeries []DailyCostDataPoint
	for _, dp := range costByDate {
		costSeries = append(costSeries, *dp)
	}

	// Build usage series grouped by date
	usageByDate := make(map[string]*DailyUsageDataPoint)
	metricSet := make(map[string]bool)
	for _, u := range usage {
		dateKey := u.UsageDate.Format("2006-01-02")
		if _, ok := usageByDate[dateKey]; !ok {
			usageByDate[dateKey] = &DailyUsageDataPoint{
				Date:    u.UsageDate,
				Metrics: make(map[string]decimal.Decimal),
			}
		}
		usageByDate[dateKey].Metrics[u.Metric] = u.Value
		metricSet[u.Metric] = true
	}

	// Convert to sorted slice
	var usageSeries []DailyUsageDataPoint
	for _, dp := range usageByDate {
		usageSeries = append(usageSeries, *dp)
	}

	// Extract dimension and metric names
	var dimensions []string
	for d := range dimensionSet {
		dimensions = append(dimensions, d)
	}
	var metrics []string
	for m := range metricSet {
		metrics = append(metrics, m)
	}

	return &NodeMetricsTimeSeriesResponse{
		NodeID:      node.ID,
		NodeName:    node.Name,
		NodeType:    node.Type,
		Period:      fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Currency:    currency,
		CostSeries:  costSeries,
		UsageSeries: usageSeries,
		Dimensions:  dimensions,
		Metrics:     metrics,
	}, nil
}

// CSV Export Methods

// ExportProductsToCSV exports products with costs to CSV format
func (s *Service) ExportProductsToCSV(ctx context.Context, req CostAttributionRequest, writer io.Writer) error {
	// Get products data
	products, err := s.store.Costs.ListNodesWithCosts(ctx, req.StartDate, req.EndDate, req.Currency, "product", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list products: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Name", "Type", "Total Cost", "Currency", "Period Start", "Period End"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, product := range products {
		row := []string{
			product.ID.String(),
			product.Name,
			product.Type,
			product.TotalCost.StringFixed(2),
			product.Currency,
			req.StartDate.Format("2006-01-02"),
			req.EndDate.Format("2006-01-02"),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportNodesToCSV exports all nodes with costs to CSV format
func (s *Service) ExportNodesToCSV(ctx context.Context, req CostAttributionRequest, nodeType string, writer io.Writer) error {
	// Get nodes data
	nodes, err := s.store.Costs.ListNodesWithCosts(ctx, req.StartDate, req.EndDate, req.Currency, nodeType, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Name", "Type", "Total Cost", "Currency", "Period Start", "Period End"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, node := range nodes {
		row := []string{
			node.ID.String(),
			node.Name,
			node.Type,
			node.TotalCost.StringFixed(2),
			node.Currency,
			req.StartDate.Format("2006-01-02"),
			req.EndDate.Format("2006-01-02"),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportCostsByTypeToCSV exports costs aggregated by type to CSV format
func (s *Service) ExportCostsByTypeToCSV(ctx context.Context, req CostAttributionRequest, writer io.Writer) error {
	// Get costs by type data
	costsByType, err := s.store.Costs.GetCostsByType(ctx, req.StartDate, req.EndDate, req.Currency)
	if err != nil {
		return fmt.Errorf("failed to get costs by type: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"Type", "Direct Cost", "Indirect Cost", "Total Cost", "Node Count", "Percent of Total", "Currency", "Period Start", "Period End"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, ct := range costsByType {
		row := []string{
			ct.Type,
			ct.DirectCost.StringFixed(2),
			ct.IndirectCost.StringFixed(2),
			ct.TotalCost.StringFixed(2),
			strconv.Itoa(ct.NodeCount),
			strconv.FormatFloat(ct.PercentOfTotal, 'f', 2, 64),
			req.Currency,
			req.StartDate.Format("2006-01-02"),
			req.EndDate.Format("2006-01-02"),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportRecommendationsToCSV exports cost optimization recommendations to CSV format
func (s *Service) ExportRecommendationsToCSV(ctx context.Context, req CostAttributionRequest, nodeID *uuid.UUID, writer io.Writer) error {
	// Get recommendations data
	var recommendations []models.CostRecommendation
	var err error

	if nodeID != nil {
		recommendations, err = s.recommendationAnalyzer.AnalyzeNode(ctx, *nodeID, req.StartDate, req.EndDate)
	} else {
		recommendations, err = s.recommendationAnalyzer.AnalyzeAllNodes(ctx, req.StartDate, req.EndDate)
	}

	if err != nil {
		return fmt.Errorf("failed to analyze recommendations: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"ID", "Node ID", "Node Name", "Node Type", "Type", "Severity", "Title", "Description",
		"Current Cost", "Potential Savings", "Currency", "Metric", "Current Value", "Peak Value",
		"Average Value", "Utilization Percent", "Recommended Action", "Analysis Period",
		"Start Date", "End Date", "Created At",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, rec := range recommendations {
		row := []string{
			rec.ID.String(),
			rec.NodeID.String(),
			rec.NodeName,
			rec.NodeType,
			string(rec.Type),
			string(rec.Severity),
			rec.Title,
			rec.Description,
			rec.CurrentCost.StringFixed(2),
			rec.PotentialSavings.StringFixed(2),
			rec.Currency,
			rec.Metric,
			rec.CurrentValue.StringFixed(2),
			rec.PeakValue.StringFixed(2),
			rec.AverageValue.StringFixed(2),
			rec.UtilizationPercent.StringFixed(2),
			rec.RecommendedAction,
			rec.AnalysisPeriod,
			rec.StartDate.Format("2006-01-02"),
			rec.EndDate.Format("2006-01-02"),
			rec.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportDetailedCostsToCSV exports detailed cost records (not aggregated) to CSV format
func (s *Service) ExportDetailedCostsToCSV(ctx context.Context, req CostAttributionRequest, nodeType string, writer io.Writer) error {
	// Get detailed cost data from allocation results
	costs, err := s.store.Costs.GetDetailedCostRecords(ctx, req.StartDate, req.EndDate, req.Currency, nodeType)
	if err != nil {
		return fmt.Errorf("failed to get detailed cost records: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"Node ID", "Node Name", "Node Type", "Date", "Dimension",
		"Direct Cost", "Indirect Cost", "Total Cost", "Currency",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, cost := range costs {
		row := []string{
			cost.NodeID.String(),
			cost.NodeName,
			cost.NodeType,
			cost.Date.Format("2006-01-02"),
			cost.Dimension,
			cost.DirectCost.StringFixed(2),
			cost.IndirectCost.StringFixed(2),
			cost.TotalCost.StringFixed(2),
			cost.Currency,
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportRawCostsToCSV exports raw ingested cost data (pre-allocation) to CSV format
func (s *Service) ExportRawCostsToCSV(ctx context.Context, req CostAttributionRequest, nodeType string, writer io.Writer) error {
	// Get raw cost data from node_costs_by_dimension
	costs, err := s.store.Costs.GetRawCostRecords(ctx, req.StartDate, req.EndDate, req.Currency, nodeType)
	if err != nil {
		return fmt.Errorf("failed to get raw cost records: %w", err)
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"Node ID", "Node Name", "Node Type", "Date", "Dimension",
		"Amount", "Currency", "Metadata",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, cost := range costs {
		metadataStr := ""
		if cost.Metadata != nil {
			if metadataBytes, err := json.Marshal(cost.Metadata); err == nil {
				metadataStr = string(metadataBytes)
			}
		}

		row := []string{
			cost.NodeID.String(),
			cost.NodeName,
			cost.NodeType,
			cost.Date.Format("2006-01-02"),
			cost.Dimension,
			cost.Amount.StringFixed(2),
			cost.Currency,
			metadataStr,
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportProductHierarchyToCSV exports product hierarchy with downstream relationships to CSV
func (s *Service) ExportProductHierarchyToCSV(ctx context.Context, req CostAttributionRequest, writer io.Writer) error {
	// Get product hierarchy records
	records, err := s.store.Costs.GetProductHierarchyRecords(ctx, req.StartDate, req.EndDate, req.Currency)
	if err != nil {
		return fmt.Errorf("failed to get product hierarchy records: %w", err)
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"Product ID",
		"Product Name",
		"Product Type",
		"Date",
		"Dimension",
		"Direct Cost",
		"Indirect Cost",
		"Total Cost",
		"Shared Service Cost",
		"Currency",
		"Downstream Node ID",
		"Downstream Node Name",
		"Downstream Node Type",
		"Contributed Amount",
		"Allocation Strategy",
		"Product Description",
		"Product Metadata JSON",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write records
	for _, record := range records {
		// Extract description from metadata
		description := ""
		if desc, ok := record.Metadata["description"].(string); ok {
			description = desc
		}

		// Convert metadata to JSON string
		metadataJSON := "{}"
		if len(record.Metadata) > 0 {
			if jsonBytes, err := json.Marshal(record.Metadata); err == nil {
				metadataJSON = string(jsonBytes)
			}
		}

		// Handle nullable fields
		downstreamNodeID := ""
		if record.DownstreamNodeID != nil {
			downstreamNodeID = record.DownstreamNodeID.String()
		}

		downstreamNodeName := ""
		if record.DownstreamNodeName != nil {
			downstreamNodeName = *record.DownstreamNodeName
		}

		downstreamNodeType := ""
		if record.DownstreamNodeType != nil {
			downstreamNodeType = *record.DownstreamNodeType
		}

		contributedAmount := ""
		if record.ContributedAmount != nil {
			contributedAmount = record.ContributedAmount.StringFixed(2)
		}

		allocationStrategy := ""
		if record.AllocationStrategy != nil {
			allocationStrategy = *record.AllocationStrategy
		}

		row := []string{
			record.ProductID.String(),
			record.ProductName,
			record.ProductType,
			record.Date.Format("2006-01-02"),
			record.Dimension,
			record.DirectCost.StringFixed(2),
			record.IndirectCost.StringFixed(2),
			record.TotalCost.StringFixed(2),
			record.SharedServiceCost.StringFixed(2),
			record.Currency,
			downstreamNodeID,
			downstreamNodeName,
			downstreamNodeType,
			contributedAmount,
			allocationStrategy,
			description,
			metadataJSON,
		}

		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
