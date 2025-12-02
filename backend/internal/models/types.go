package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CostNode represents a node in the cost attribution graph
type CostNode struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	Name       string                 `json:"name" db:"name"`
	Type       string                 `json:"type" db:"type"`
	CostLabels map[string]interface{} `json:"cost_labels" db:"cost_labels"`
	IsPlatform bool                   `json:"is_platform" db:"is_platform"`
	Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
	ArchivedAt *time.Time             `json:"archived_at,omitempty" db:"archived_at"`
}

// DependencyEdge represents a dependency relationship between two nodes
type DependencyEdge struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	ParentID          uuid.UUID              `json:"parent_id" db:"parent_id"`
	ChildID           uuid.UUID              `json:"child_id" db:"child_id"`
	DefaultStrategy   string                 `json:"default_strategy" db:"default_strategy"`
	DefaultParameters map[string]interface{} `json:"default_parameters" db:"default_parameters"`
	ActiveFrom        time.Time              `json:"active_from" db:"active_from"`
	ActiveTo          *time.Time             `json:"active_to,omitempty" db:"active_to"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// EdgeStrategy represents a dimension-specific strategy override for an edge
type EdgeStrategy struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	EdgeID     uuid.UUID              `json:"edge_id" db:"edge_id"`
	Dimension  *string                `json:"dimension,omitempty" db:"dimension"`
	Strategy   string                 `json:"strategy" db:"strategy"`
	Parameters map[string]interface{} `json:"parameters" db:"parameters"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
}

// NodeCostByDimension represents direct costs for a node on a specific date and dimension
type NodeCostByDimension struct {
	NodeID    uuid.UUID              `json:"node_id" db:"node_id"`
	CostDate  time.Time              `json:"cost_date" db:"cost_date"`
	Dimension string                 `json:"dimension" db:"dimension"`
	Amount    decimal.Decimal        `json:"amount" db:"amount"`
	Currency  string                 `json:"currency" db:"currency"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// NodeUsageByDimension represents usage metrics for a node on a specific date
type NodeUsageByDimension struct {
	NodeID    uuid.UUID              `json:"node_id" db:"node_id"`
	UsageDate time.Time              `json:"usage_date" db:"usage_date"`
	Metric    string                 `json:"metric" db:"metric"`
	Value     decimal.Decimal        `json:"value" db:"value"`
	Unit      string                 `json:"unit" db:"unit"`
	Labels    map[string]string      `json:"labels,omitempty" db:"labels"`     // Optional labels for filtering (e.g. customer_id, environment)
	Source    string                 `json:"source,omitempty" db:"source"`     // Source of the metric (e.g. "dynatrace", "prometheus", "manual")
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// UsageLabelFilter represents a filter for querying usage metrics by labels
type UsageLabelFilter struct {
	Key      string   `json:"key"`      // Label key to filter on (e.g. "customer_id")
	Values   []string `json:"values"`   // Values to match (OR semantics)
	Operator string   `json:"operator"` // "eq", "neq", "in", "not_in", "exists", "not_exists"
}

// UsageQueryOptions represents options for querying usage metrics
type UsageQueryOptions struct {
	NodeIDs      []uuid.UUID        `json:"node_ids,omitempty"`
	Metrics      []string           `json:"metrics,omitempty"`
	StartDate    time.Time          `json:"start_date"`
	EndDate      time.Time          `json:"end_date"`
	LabelFilters []UsageLabelFilter `json:"label_filters,omitempty"`
	Source       string             `json:"source,omitempty"`
	GroupByLabel string             `json:"group_by_label,omitempty"` // Group results by this label (e.g. "customer_id")
}

// ComputationRun represents a single allocation computation run
type ComputationRun struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	WindowStart time.Time  `json:"window_start" db:"window_start"`
	WindowEnd   time.Time  `json:"window_end" db:"window_end"`
	GraphHash   string     `json:"graph_hash" db:"graph_hash"`
	Status      string     `json:"status" db:"status"`
	Notes       *string    `json:"notes,omitempty" db:"notes"`
}

// AllocationResultByDimension represents the allocation result for a node on a specific date and dimension
type AllocationResultByDimension struct {
	RunID          uuid.UUID       `json:"run_id" db:"run_id"`
	NodeID         uuid.UUID       `json:"node_id" db:"node_id"`
	AllocationDate time.Time       `json:"allocation_date" db:"allocation_date"`
	Dimension      string          `json:"dimension" db:"dimension"`
	DirectAmount   decimal.Decimal `json:"direct_amount" db:"direct_amount"`
	IndirectAmount decimal.Decimal `json:"indirect_amount" db:"indirect_amount"`
	TotalAmount    decimal.Decimal `json:"total_amount" db:"total_amount"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// ContributionResultByDimension represents how much a child contributed to a parent
type ContributionResultByDimension struct {
	RunID             uuid.UUID       `json:"run_id" db:"run_id"`
	ParentID          uuid.UUID       `json:"parent_id" db:"parent_id"`
	ChildID           uuid.UUID       `json:"child_id" db:"child_id"`
	ContributionDate  time.Time       `json:"contribution_date" db:"contribution_date"`
	Dimension         string          `json:"dimension" db:"dimension"`
	ContributedAmount decimal.Decimal `json:"contributed_amount" db:"contributed_amount"`
	Path              []uuid.UUID     `json:"path" db:"path"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// ComputationStatus represents the status of a computation run
type ComputationStatus string

const (
	ComputationStatusPending   ComputationStatus = "pending"
	ComputationStatusRunning   ComputationStatus = "running"
	ComputationStatusCompleted ComputationStatus = "completed"
	ComputationStatusFailed    ComputationStatus = "failed"
)

// NodeType represents different types of cost nodes
type NodeType string

const (
	NodeTypeProduct     NodeType = "product"
	NodeTypeService     NodeType = "service"
	NodeTypeResource    NodeType = "resource"
	NodeTypePlatform    NodeType = "platform"
	NodeTypeInfra       NodeType = "infrastructure"
	NodeTypeShared      NodeType = "shared"
)

// AllocationStrategy represents different cost allocation strategies
type AllocationStrategy string

const (
	StrategyProportionalOn         AllocationStrategy = "proportional_on"
	StrategyEqual                  AllocationStrategy = "equal"
	StrategyFixedPercent           AllocationStrategy = "fixed_percent"
	StrategyCappedProp             AllocationStrategy = "capped_proportional"
	StrategyResidualToMax          AllocationStrategy = "residual_to_max"
	StrategyWeightedAverage        AllocationStrategy = "weighted_average"
	StrategyHybridFixedProp        AllocationStrategy = "hybrid_fixed_proportional"
	StrategyMinFloorProportional   AllocationStrategy = "min_floor_proportional"
	StrategySegmentFilteredProp    AllocationStrategy = "segment_filtered_proportional"
)

// IsValidStrategy checks if a strategy string is a valid allocation strategy
func IsValidStrategy(s string) bool {
	switch AllocationStrategy(s) {
	case StrategyProportionalOn, StrategyEqual, StrategyFixedPercent,
		StrategyCappedProp, StrategyResidualToMax, StrategyWeightedAverage,
		StrategyHybridFixedProp, StrategyMinFloorProportional, StrategySegmentFilteredProp:
		return true
	default:
		return false
	}
}

// Dimension represents common cost dimensions
type Dimension string

const (
	DimensionInstanceHours      Dimension = "instance_hours"
	DimensionStorageGBMonth     Dimension = "storage_gb_month"
	DimensionEgressGB           Dimension = "egress_gb"
	DimensionIOPS               Dimension = "iops"
	DimensionBackupsGBMonth     Dimension = "backups_gb_month"
	DimensionRequests           Dimension = "requests"
	DimensionComputeHours       Dimension = "compute_hours"
	DimensionNetworkGB          Dimension = "network_gb"
)

// Common dimensions slice for iteration
var CommonDimensions = []Dimension{
	DimensionInstanceHours,
	DimensionStorageGBMonth,
	DimensionEgressGB,
	DimensionIOPS,
	DimensionBackupsGBMonth,
	DimensionRequests,
	DimensionComputeHours,
	DimensionNetworkGB,
}

// AllocationInput represents input data for allocation computation
type AllocationInput struct {
	Nodes       []CostNode                `json:"nodes"`
	Edges       []DependencyEdge          `json:"edges"`
	Strategies  []EdgeStrategy            `json:"strategies"`
	Costs       []NodeCostByDimension     `json:"costs"`
	Usage       []NodeUsageByDimension    `json:"usage"`
	WindowStart time.Time                 `json:"window_start"`
	WindowEnd   time.Time                 `json:"window_end"`
	Dimensions  []string                  `json:"dimensions"`
}

// AllocationOutput represents the result of allocation computation
type AllocationOutput struct {
	RunID         uuid.UUID                        `json:"run_id"`
	Allocations   []AllocationResultByDimension    `json:"allocations"`
	Contributions []ContributionResultByDimension  `json:"contributions"`
	Summary       AllocationSummary                `json:"summary"`
}

// AllocationSummary provides high-level statistics about an allocation run
type AllocationSummary struct {
	TotalNodes        int                        `json:"total_nodes"`
	TotalEdges        int                        `json:"total_edges"`
	ProcessedDays     int                        `json:"processed_days"`
	TotalDirectCost   map[string]decimal.Decimal `json:"total_direct_cost"`
	TotalIndirectCost map[string]decimal.Decimal `json:"total_indirect_cost"`
	TotalCost         map[string]decimal.Decimal `json:"total_cost"`
	ProcessingTime    time.Duration              `json:"processing_time"`
}

// Custom JSON marshaling for JSONB fields
func (cn *CostNode) MarshalJSON() ([]byte, error) {
	type Alias CostNode
	return json.Marshal(&struct {
		*Alias
		CostLabels json.RawMessage `json:"cost_labels"`
		Metadata   json.RawMessage `json:"metadata"`
	}{
		Alias:      (*Alias)(cn),
		CostLabels: mustMarshalJSON(cn.CostLabels),
		Metadata:   mustMarshalJSON(cn.Metadata),
	})
}

func mustMarshalJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return data
}
