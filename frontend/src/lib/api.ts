import type {
  ProductHierarchyResponse,
  IndividualNodeResponse,
  PlatformServicesResponse,
  HealthResponse,
  NodeListResponse,
  CostsByTypeResponse,
  CostsByDimensionResponse,
} from "@/types/api"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message)
    this.name = "ApiError"
  }
}

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new ApiError(
        errorData.error || `HTTP ${response.status}: ${response.statusText}`,
        response.status,
        errorData.code
      )
    }

    return await response.json()
  } catch (error) {
    if (error instanceof ApiError) {
      throw error
    }
    throw new ApiError(
      error instanceof Error ? error.message : "Network error",
      0
    )
  }
}

export interface QueryParams {
  start_date: string
  end_date: string
  dimensions?: string[]
  include_trend?: boolean
  currency?: string
}

export const api = {
  health: {
    check: () => fetchApi<HealthResponse>("/health"),
  },

  products: {
    list: (params: QueryParams & { limit?: number; offset?: number }) => {
      const searchParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
      })

      if (params.currency) {
        searchParams.append("currency", params.currency)
      }
      if (params.limit !== undefined) {
        searchParams.append("limit", String(params.limit))
      }
      if (params.offset !== undefined) {
        searchParams.append("offset", String(params.offset))
      }

      return fetchApi<NodeListResponse>(
        `/api/v1/products?${searchParams.toString()}`
      )
    },

    getHierarchy: (params: QueryParams) => {
      const searchParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
      })
      
      if (params.dimensions?.length) {
        params.dimensions.forEach(dim => searchParams.append("dimensions", dim))
      }
      if (params.include_trend !== undefined) {
        searchParams.append("include_trend", String(params.include_trend))
      }
      if (params.currency) {
        searchParams.append("currency", params.currency)
      }

      return fetchApi<ProductHierarchyResponse>(
        `/api/v1/products/hierarchy?${searchParams.toString()}`
      )
    },
  },

  nodes: {
    list: (params: QueryParams & { type?: string; limit?: number; offset?: number }) => {
      const searchParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
      })

      if (params.currency) {
        searchParams.append("currency", params.currency)
      }
      if (params.type) {
        searchParams.append("type", params.type)
      }
      if (params.limit !== undefined) {
        searchParams.append("limit", String(params.limit))
      }
      if (params.offset !== undefined) {
        searchParams.append("offset", String(params.offset))
      }

      return fetchApi<NodeListResponse>(
        `/api/v1/nodes?${searchParams.toString()}`
      )
    },

    getDetails: (nodeId: string, params: QueryParams) => {
      const searchParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
      })

      if (params.dimensions?.length) {
        params.dimensions.forEach(dim => searchParams.append("dimensions", dim))
      }
      if (params.include_trend !== undefined) {
        searchParams.append("include_trend", String(params.include_trend))
      }
      if (params.currency) {
        searchParams.append("currency", params.currency)
      }

      return fetchApi<IndividualNodeResponse>(
        `/api/v1/nodes/${nodeId}?${searchParams.toString()}`
      )
    },
  },

  costs: {
    byType: (params: QueryParams) => {
      const searchParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
      })

      if (params.currency) {
        searchParams.append("currency", params.currency)
      }

      return fetchApi<CostsByTypeResponse>(
        `/api/v1/costs/by-type?${searchParams.toString()}`
      )
    },

    byDimension: (params: QueryParams & { key: string }) => {
      const searchParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
        key: params.key,
      })

      if (params.currency) {
        searchParams.append("currency", params.currency)
      }

      return fetchApi<CostsByDimensionResponse>(
        `/api/v1/costs/by-dimension?${searchParams.toString()}`
      )
    },
  },

  platform: {
    getServices: (params: QueryParams) => {
      const searchParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
      })
      
      if (params.dimensions?.length) {
        params.dimensions.forEach(dim => searchParams.append("dimensions", dim))
      }
      if (params.include_trend !== undefined) {
        searchParams.append("include_trend", String(params.include_trend))
      }
      if (params.currency) {
        searchParams.append("currency", params.currency)
      }

      return fetchApi<PlatformServicesResponse>(
        `/api/v1/platform/services?${searchParams.toString()}`
      )
    },
  },
}

