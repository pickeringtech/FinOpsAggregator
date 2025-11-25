// API Response Types based on OpenAPI specification

export interface CostBreakdown {
  total: string
  currency: string
  dimensions: Record<string, string>
}

export interface CostSummary {
  // Allocated product total (including the Unallocated bucket on the Products page)
  total_cost: string
  // Raw infrastructure spend from node_costs_by_dimension for the same date range
  raw_total_cost: string
  // 0-100: how much of the raw spend is covered by allocated totals
  allocation_coverage_percent: number
  currency: string
  period: string
  start_date: string
  end_date: string
  node_count: number
  product_count: number
  platform_node_count: number
}

export type NodeType = "product" | "resource" | "shared" | "platform"

export interface ProductNode {
  id: string
  name: string
  type: NodeType
  direct_costs: CostBreakdown
  holistic_costs: CostBreakdown
  shared_service_costs: CostBreakdown
  children?: ProductNode[]
  metadata?: {
    description?: string
    [key: string]: unknown
  }
}

export interface ProductHierarchyResponse {
  products: ProductNode[]
  summary: CostSummary
}

export interface NodeDetails {
  id: string
  name: string
  type: NodeType
  is_platform: boolean
  cost_labels?: Record<string, string>
  metadata?: {
    description?: string
    [key: string]: unknown
  }
}

export type RelationshipType = "parent" | "child" | "peer"
export type AllocationStrategy = "equal" | "proportional" | "weighted" | "usage_based"

export interface NodeDependency {
  id: string
  name: string
  type: NodeType
  relationship_type: RelationshipType
  strategy: AllocationStrategy
}

export interface AllocationDetail {
  from_node: {
    id: string
    name: string
    type: string
  }
  to_node: {
    id: string
    name: string
    type: string
  }
  amount: string
  currency: string
  dimension: string
  strategy: string
  allocation_date: string
}

export interface IndividualNodeResponse {
  node: NodeDetails
  direct_costs: CostBreakdown
  allocated_costs: CostBreakdown
  total_costs: CostBreakdown
  dependencies: NodeDependency[]
  allocations?: AllocationDetail[]
}

export interface AllocationTarget {
  node_id: string
  node_name: string
  node_type: string
  amount: string
  currency: string
  percentage: number
}

export interface WeightedTarget {
  node_id: string
  node_name: string
  node_type: string
  weight: string
  amount: string
  currency: string
  percentage: number
}

export interface PlatformService {
  id: string
  name: string
  type: "platform"
  direct_costs: CostBreakdown
  allocated_to: AllocationTarget[]
  metadata?: {
    description?: string
    [key: string]: unknown
  }
}

export interface SharedService {
  id: string
  name: string
  type: "shared"
  direct_costs: CostBreakdown
  weighted_targets: WeightedTarget[]
  metadata?: {
    description?: string
    [key: string]: unknown
  }
}

export interface WeightedAllocation {
  source_service_id: string
  target_node_id: string
  allocation_amount: string
  weight: number
  strategy: AllocationStrategy
  dimension?: string
}

export interface PlatformServicesResponse {
  platform_services: PlatformService[]
  shared_services: SharedService[]
  summary: CostSummary
  weighted_allocations?: WeightedAllocation[]
}

export interface HealthResponse {
  status: string
  timestamp: string
  version: string
  database: string
}

// Generic API Types for cost aggregation
export interface NodeListResponse {
  nodes: NodeWithCost[]
  total_count: number
  limit: number
  offset: number
}

export interface NodeWithCost {
  id: string
  name: string
  type: string
  total_cost: string
  currency: string
}

export interface CostsByTypeResponse {
  aggregations: TypeAggregation[]
  total_cost: string
  currency: string
}

// Cost Optimization Recommendations
export type RecommendationType = "downsize" | "rightsize" | "unused" | "overprovisioned"
export type RecommendationSeverity = "low" | "medium" | "high"

export interface CostRecommendation {
  id: string
  node_id: string
  node_name: string
  node_type: string
  type: RecommendationType
  severity: RecommendationSeverity
  title: string
  description: string
  current_cost: string
  potential_savings: string
  currency: string
  metric: string
  current_value: string
  peak_value: string
  average_value: string
  utilization_percent: string
  recommended_action: string
  analysis_period: string
  start_date: string
  end_date: string
  created_at: string
}

export interface RecommendationsResponse {
  recommendations: CostRecommendation[]
  total_savings: string
  currency: string
  high_severity_count: number
  medium_severity_count: number
  low_severity_count: number
}

export interface TypeAggregation {
  type: string
  total_cost: string
  node_count: number
}

export interface CostsByDimensionResponse {
  dimension_key: string
  aggregations: DimensionAggregation[]
  total_cost: string
  currency: string
}

export interface DimensionAggregation {
  value: string
  total_cost: string
  node_count: number
}

