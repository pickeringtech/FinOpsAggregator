// API Response Types based on OpenAPI specification

export interface CostBreakdown {
  total: string
  currency: string
  dimensions: Record<string, string>
}

export interface CostSummary {
  total_cost: string
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
  source_node_id: string
  target_node_id: string
  amount: string
  strategy: AllocationStrategy
  weight: number
  dimension?: string
}

export interface IndividualNodeResponse {
  node: NodeDetails
  direct_costs: CostBreakdown
  allocated_costs: CostBreakdown
  total_costs: CostBreakdown
  dependencies: NodeDependency[]
  allocations?: AllocationDetail[]
}

export interface PlatformService {
  id: string
  name: string
  type: "platform"
  direct_costs: CostBreakdown
  allocated_costs: CostBreakdown
  total_costs: CostBreakdown
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
  allocated_costs: CostBreakdown
  total_costs: CostBreakdown
  allocation_targets?: string[]
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
  weighted_allocations: WeightedAllocation[]
}

export interface HealthResponse {
  status: string
  timestamp: string
  version: string
  database: string
}

