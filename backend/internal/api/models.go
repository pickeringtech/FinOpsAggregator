package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductHierarchyResponse represents the product hierarchy tree with cost data
type ProductHierarchyResponse struct {
	Products []ProductNode `json:"products"`
	Summary  CostSummary   `json:"summary"`
}

// InfrastructureHierarchyResponse represents the infrastructure hierarchy tree with cost data
// This mirrors ProductHierarchyResponse but for platform/shared/resource nodes
type InfrastructureHierarchyResponse struct {
	Infrastructure []InfrastructureNode `json:"infrastructure"`
	Summary        InfraSummary         `json:"summary"`
}

// InfrastructureNode represents a node in the infrastructure hierarchy
type InfrastructureNode struct {
	ID              uuid.UUID               `json:"id"`
	Name            string                  `json:"name"`
	Type            string                  `json:"type"` // platform, shared, resource, infrastructure
	IsPlatform      bool                    `json:"is_platform"`
	DirectCosts     CostBreakdown           `json:"direct_costs"`      // Cost originating on this node
	AllocatedCosts  CostBreakdown           `json:"allocated_costs"`   // Cost allocated OUT to products
	UnallocatedCost decimal.Decimal         `json:"unallocated_cost"`  // Direct - Allocated (what's left)
	AllocationPct   float64                 `json:"allocation_pct"`    // % of direct cost that's been allocated
	AllocatedTo     []InfraAllocationTarget `json:"allocated_to"`      // Products this node allocates to
	Children        []InfrastructureNode    `json:"children,omitempty"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// InfraAllocationTarget represents a product that receives allocation from an infrastructure node
type InfraAllocationTarget struct {
	ID       uuid.UUID       `json:"id"`
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
	Strategy string          `json:"strategy"`
	Percent  float64         `json:"percent"` // % of parent's cost allocated to this target
}

// InfraSummary provides summary for infrastructure hierarchy
type InfraSummary struct {
	TotalDirectCost      decimal.Decimal `json:"total_direct_cost"`      // Sum of all infra direct costs
	TotalAllocatedCost   decimal.Decimal `json:"total_allocated_cost"`   // Sum of all costs allocated to products
	TotalUnallocatedCost decimal.Decimal `json:"total_unallocated_cost"` // Direct - Allocated
	AllocationPct        float64         `json:"allocation_pct"`         // Overall allocation percentage
	Currency             string          `json:"currency"`
	Period               string          `json:"period"`
	StartDate            time.Time       `json:"start_date"`
	EndDate              time.Time       `json:"end_date"`
	PlatformCount        int             `json:"platform_count"`
	SharedCount          int             `json:"shared_count"`
	ResourceCount        int             `json:"resource_count"`
	TotalNodeCount       int             `json:"total_node_count"`
}

// NodeMetricsTimeSeriesResponse represents cost and usage metrics over time for a node
type NodeMetricsTimeSeriesResponse struct {
	NodeID     uuid.UUID              `json:"node_id"`
	NodeName   string                 `json:"node_name"`
	NodeType   string                 `json:"node_type"`
	Period     string                 `json:"period"`
	StartDate  time.Time              `json:"start_date"`
	EndDate    time.Time              `json:"end_date"`
	Currency   string                 `json:"currency"`
	CostSeries []DailyCostDataPoint   `json:"cost_series"`
	UsageSeries []DailyUsageDataPoint `json:"usage_series"`
	Dimensions []string               `json:"dimensions"` // Available cost dimensions
	Metrics    []string               `json:"metrics"`    // Available usage metrics
}

// DailyCostDataPoint represents cost data for a single day with dimension breakdown
type DailyCostDataPoint struct {
	Date       time.Time                  `json:"date"`
	TotalCost  decimal.Decimal            `json:"total_cost"`
	Dimensions map[string]decimal.Decimal `json:"dimensions"`
}

// DailyUsageDataPoint represents usage metrics for a single day
type DailyUsageDataPoint struct {
	Date    time.Time                  `json:"date"`
	Metrics map[string]decimal.Decimal `json:"metrics"` // metric_name -> value
}

// ProductNode represents a product in the hierarchy with its cost data
type ProductNode struct {
	ID                uuid.UUID         `json:"id"`
	Name              string            `json:"name"`
	Type              string            `json:"type"`
	DirectCosts       CostBreakdown     `json:"direct_costs"`
	HolisticCosts     CostBreakdown     `json:"holistic_costs"`
	SharedServiceCosts CostBreakdown    `json:"shared_service_costs"`
	Children          []ProductNode     `json:"children,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// CostBreakdown represents cost data broken down by dimensions
type CostBreakdown struct {
	Total      decimal.Decimal            `json:"total"`
	Currency   string                     `json:"currency"`
	Dimensions map[string]decimal.Decimal `json:"dimensions"`
	Trend      []DailyCostPoint           `json:"trend,omitempty"`
}

// DailyCostPoint represents a single point in a cost trend
type DailyCostPoint struct {
	Date time.Time       `json:"date"`
	Cost decimal.Decimal `json:"cost"`
}

// CostSummary provides overall cost summary information
// TotalCost here represents the allocated product total shown on the Products page
// RawTotalCost represents the underlying raw infrastructure spend for the same date range
// AllocationCoveragePercent indicates how much of the raw spend has been allocated into actual products
type CostSummary struct {
	TotalCost                 decimal.Decimal `json:"total_cost"`                  // allocated product total (including unallocated bucket)
	RawTotalCost              decimal.Decimal `json:"raw_total_cost"`              // raw infra spend from node_costs_by_dimension
	AllocationCoveragePercent float64         `json:"allocation_coverage_percent"` // 0-100, share of raw spend allocated to real products (excluding unallocated bucket)
	Currency                  string          `json:"currency"`
	Period                    string          `json:"period"`
	StartDate                 time.Time       `json:"start_date"`
	EndDate                   time.Time       `json:"end_date"`
	NodeCount                 int             `json:"node_count"`
	ProductCount              int             `json:"product_count"`
	FinalCostCentreCount      int             `json:"final_cost_centre_count"` // count of product nodes with no outgoing product edges
	PlatformNodeCount         int             `json:"platform_node_count"`
	SanityCheckPassed         bool            `json:"sanity_check_passed"`          // true if product tree totals match allocated cost
	SanityCheckWarnings       []string        `json:"sanity_check_warnings,omitempty"` // warnings if sanity checks fail
}

// IndividualNodeResponse represents detailed cost data for a single node
type IndividualNodeResponse struct {
	Node         NodeDetails     `json:"node"`
	DirectCosts  CostBreakdown   `json:"direct_costs"`
	AllocatedCosts CostBreakdown `json:"allocated_costs"`
	TotalCosts   CostBreakdown   `json:"total_costs"`
	Dependencies []NodeDependency `json:"dependencies"`
	Allocations  []AllocationDetail `json:"allocations"`
}

// NodeDetails represents basic node information
type NodeDetails struct {
	ID         uuid.UUID              `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	IsPlatform bool                   `json:"is_platform"`
	CostLabels map[string]interface{} `json:"cost_labels"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NodeDependency represents a dependency relationship
type NodeDependency struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	RelationshipType string    `json:"relationship_type"` // "parent" or "child"
	Strategy         string    `json:"strategy"`
}

// AllocationDetail represents how costs were allocated to/from this node
type AllocationDetail struct {
	FromNode      NodeReference   `json:"from_node"`
	ToNode        NodeReference   `json:"to_node"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Dimension     string          `json:"dimension"`
	Strategy      string          `json:"strategy"`
	AllocationDate time.Time      `json:"allocation_date"`
}

// NodeReference represents a lightweight node reference
type NodeReference struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

// PlatformServicesResponse represents platform and shared services cost data
type PlatformServicesResponse struct {
	PlatformServices []PlatformService `json:"platform_services"`
	SharedServices   []SharedService   `json:"shared_services"`
	Summary          CostSummary       `json:"summary"`
	WeightedAllocations []WeightedAllocation `json:"weighted_allocations"`
}

// PlatformService represents a platform service with its cost data
type PlatformService struct {
	ID            uuid.UUID         `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	DirectCosts   CostBreakdown     `json:"direct_costs"`
	AllocatedTo   []AllocationTarget `json:"allocated_to"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SharedService represents a shared service with weighted allocation
type SharedService struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	DirectCosts     CostBreakdown     `json:"direct_costs"`
	WeightedTargets []WeightedTarget  `json:"weighted_targets"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AllocationTarget represents where costs are allocated to
type AllocationTarget struct {
	NodeID     uuid.UUID       `json:"node_id"`
	NodeName   string          `json:"node_name"`
	NodeType   string          `json:"node_type"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	Percentage float64         `json:"percentage"`
}

// WeightedTarget represents a target with weighted allocation
type WeightedTarget struct {
	NodeID     uuid.UUID       `json:"node_id"`
	NodeName   string          `json:"node_name"`
	NodeType   string          `json:"node_type"`
	Weight     decimal.Decimal `json:"weight"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	Percentage float64         `json:"percentage"`
}

// WeightedAllocation represents a weighted allocation from shared services
type WeightedAllocation struct {
	SharedServiceID   uuid.UUID       `json:"shared_service_id"`
	SharedServiceName string          `json:"shared_service_name"`
	TargetNodeID      uuid.UUID       `json:"target_node_id"`
	TargetNodeName    string          `json:"target_node_name"`
	Weight            decimal.Decimal `json:"weight"`
	Amount            decimal.Decimal `json:"amount"`
	Currency          string          `json:"currency"`
	Dimension         string          `json:"dimension"`
	Strategy          string          `json:"strategy"`
}

// CostAttributionRequest represents request parameters for cost attribution queries
type CostAttributionRequest struct {
	StartDate    time.Time `json:"start_date" form:"start_date"`
	EndDate      time.Time `json:"end_date" form:"end_date"`
	Dimensions   []string  `json:"dimensions,omitempty" form:"dimensions"`
	IncludeTrend bool      `json:"include_trend,omitempty" form:"include_trend"`
	Currency     string    `json:"currency,omitempty" form:"currency"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// NodeListResponse represents a list of nodes with costs
type NodeListResponse struct {
	Nodes      []NodeWithCostData `json:"nodes"`
	TotalCount int                `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// NodeWithCostData represents a node with its cost information
type NodeWithCostData struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	TotalCost decimal.Decimal `json:"total_cost"`
	Currency  string         `json:"currency"`
}

// CostsByTypeResponse represents costs aggregated by node type
type CostsByTypeResponse struct {
	Aggregations []TypeAggregation `json:"aggregations"`
	TotalCost    decimal.Decimal   `json:"total_cost"`
	Currency     string            `json:"currency"`
}

// TypeAggregation represents cost aggregated by a single type
type TypeAggregation struct {
	Type           string          `json:"type"`
	DirectCost     decimal.Decimal `json:"direct_cost"`
	IndirectCost   decimal.Decimal `json:"indirect_cost"`
	TotalCost      decimal.Decimal `json:"total_cost"`
	NodeCount      int             `json:"node_count"`
	PercentOfTotal float64         `json:"percent_of_total"`
}

// CostsByDimensionResponse represents costs aggregated by a custom dimension
type CostsByDimensionResponse struct {
	DimensionKey string                  `json:"dimension_key"`
	Aggregations []DimensionAggregation  `json:"aggregations"`
	TotalCost    decimal.Decimal         `json:"total_cost"`
	Currency     string                  `json:"currency"`
}

// DimensionAggregation represents cost aggregated by a dimension value
type DimensionAggregation struct {
	Value     string          `json:"value"`
	TotalCost decimal.Decimal `json:"total_cost"`
	NodeCount int             `json:"node_count"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Database  string    `json:"database"`
}

// RecommendationsResponse represents cost optimization recommendations
type RecommendationsResponse struct {
	Recommendations   []CostRecommendation `json:"recommendations"`
	TotalSavings      decimal.Decimal      `json:"total_savings"`
	Currency          string               `json:"currency"`
	HighSeverityCount int                  `json:"high_severity_count"`
	MediumSeverityCount int                `json:"medium_severity_count"`
	LowSeverityCount  int                  `json:"low_severity_count"`
}

// CostRecommendation represents a cost optimization recommendation
type CostRecommendation struct {
	ID                 uuid.UUID       `json:"id"`
	NodeID             uuid.UUID       `json:"node_id"`
	NodeName           string          `json:"node_name"`
	NodeType           string          `json:"node_type"`
	Type               string          `json:"type"`
	Severity           string          `json:"severity"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	CurrentCost        decimal.Decimal `json:"current_cost"`
	PotentialSavings   decimal.Decimal `json:"potential_savings"`
	Currency           string          `json:"currency"`
	Metric             string          `json:"metric"`
	CurrentValue       decimal.Decimal `json:"current_value"`
	PeakValue          decimal.Decimal `json:"peak_value"`
	AverageValue       decimal.Decimal `json:"average_value"`
	UtilizationPercent decimal.Decimal `json:"utilization_percent"`
	RecommendedAction  string          `json:"recommended_action"`
	AnalysisPeriod     string          `json:"analysis_period"`
	StartDate          time.Time       `json:"start_date"`
	EndDate            time.Time       `json:"end_date"`
	CreatedAt          time.Time       `json:"created_at"`
}

// AllocationReconciliationResponse provides debug information for allocation reconciliation
type AllocationReconciliationResponse struct {
	Period                    string                     `json:"period"`
	StartDate                 time.Time                  `json:"start_date"`
	EndDate                   time.Time                  `json:"end_date"`
	Currency                  string                     `json:"currency"`
	RawInfrastructureCost     decimal.Decimal            `json:"raw_infrastructure_cost"`
	AllocatedProductCost      decimal.Decimal            `json:"allocated_product_cost"`
	UnallocatedCost           decimal.Decimal            `json:"unallocated_cost"`
	CoveragePercent           float64                    `json:"coverage_percent"`
	ConservationDelta         decimal.Decimal            `json:"conservation_delta"`
	ConservationValid         bool                       `json:"conservation_valid"`
	FinalCostCentres          []FinalCostCentreDetail    `json:"final_cost_centres"`
	InfrastructureNodes       []InfrastructureNodeDetail `json:"infrastructure_nodes"`
	InvariantViolations       []InvariantViolation       `json:"invariant_violations"`
	GraphStats                GraphStatistics            `json:"graph_stats"`
}

// FinalCostCentreDetail provides detail about a final cost centre
type FinalCostCentreDetail struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	HolisticCost decimal.Decimal `json:"holistic_cost"`
	DirectCost   decimal.Decimal `json:"direct_cost"`
	IndirectCost decimal.Decimal `json:"indirect_cost"`
}

// InfrastructureNodeDetail provides detail about an infrastructure node
type InfrastructureNodeDetail struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	DirectCost decimal.Decimal `json:"direct_cost"`
	IsPlatform bool            `json:"is_platform"`
}

// InvariantViolation describes a detected invariant violation
type InvariantViolation struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	NodeID      string `json:"node_id,omitempty"`
	Dimension   string `json:"dimension,omitempty"`
	Expected    string `json:"expected,omitempty"`
	Actual      string `json:"actual,omitempty"`
}

// GraphStatistics provides statistics about the allocation graph
type GraphStatistics struct {
	TotalNodes         int `json:"total_nodes"`
	ProductNodes       int `json:"product_nodes"`
	InfrastructureNodes int `json:"infrastructure_nodes"`
	FinalCostCentres   int `json:"final_cost_centres"`
	TotalEdges         int `json:"total_edges"`
	MaxDepth           int `json:"max_depth"`
}

// DashboardSummaryResponse provides the correct totals for the dashboard
// This uses final cost centres (product nodes with no outgoing product edges)
// to calculate the true "Total Product Cost" without double-counting
type DashboardSummaryResponse struct {
	// TotalProductCost is the sum of holistic costs for final cost centres only
	// This is the correct total that should be displayed on the dashboard
	TotalProductCost decimal.Decimal `json:"total_product_cost"`

	// RawInfrastructureCost is the total pre-allocation spend from infrastructure nodes
	RawInfrastructureCost decimal.Decimal `json:"raw_infrastructure_cost"`

	// AllocationCoveragePercent shows what % of raw infra cost is allocated to products
	AllocationCoveragePercent float64 `json:"allocation_coverage_percent"`

	// UnallocatedCost is the gap between raw infra and allocated product costs
	UnallocatedCost decimal.Decimal `json:"unallocated_cost"`

	// CostsByType provides breakdown by node type (for pie charts, etc.)
	CostsByType []TypeAggregation `json:"costs_by_type"`

	// Counts
	ProductCount         int `json:"product_count"`
	FinalCostCentreCount int `json:"final_cost_centre_count"`
	PlatformCount        int `json:"platform_count"`
	SharedCount          int `json:"shared_count"`
	ResourceCount        int `json:"resource_count"`

	// Metadata
	Currency  string    `json:"currency"`
	Period    string    `json:"period"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
