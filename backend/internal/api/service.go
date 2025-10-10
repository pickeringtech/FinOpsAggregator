package api

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/analysis"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/shopspring/decimal"
)

// Service provides business logic for the API
type Service struct {
	store    *store.Store
	analyzer *analysis.FinOpsAnalyzer
}

// NewService creates a new API service
func NewService(store *store.Store) *Service {
	return &Service{
		store:    store,
		analyzer: analysis.NewFinOpsAnalyzer(store),
	}
}

// GetProductHierarchy retrieves the product hierarchy with cost data
func (s *Service) GetProductHierarchy(ctx context.Context, req CostAttributionRequest) (*ProductHierarchyResponse, error) {
	// Get product cost analysis
	productSummary, err := s.analyzer.AnalyzeProductCosts(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze product costs: %w", err)
	}

	// Build product hierarchy tree
	products, err := s.buildProductTree(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build product tree: %w", err)
	}

	summary := CostSummary{
		TotalCost:    productSummary.TotalCost,
		Currency:     productSummary.Currency,
		Period:       productSummary.Period,
		StartDate:    productSummary.StartDate,
		EndDate:      productSummary.EndDate,
		ProductCount: productSummary.ProductCount,
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
			Type:      ct.Type,
			TotalCost: ct.TotalCost,
			NodeCount: ct.NodeCount,
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

// getNodeDirectCosts retrieves direct costs for a node
func (s *Service) getNodeDirectCosts(ctx context.Context, nodeID uuid.UUID, req CostAttributionRequest) (*CostBreakdown, error) {
	costs, err := s.store.Costs.GetByNodeAndDateRange(ctx, nodeID, req.StartDate, req.EndDate, req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to get node costs: %w", err)
	}

	return s.buildCostBreakdown(costs, req), nil
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
