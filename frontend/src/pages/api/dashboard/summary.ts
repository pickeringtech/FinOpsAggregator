import type { NextApiRequest, NextApiResponse } from 'next'

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface DashboardSummary {
  top_products: Array<{
    id: string
    name: string
    type: string
    total_cost: string
    currency: string
  }>
  platform_nodes: Array<{
    id: string
    name: string
    type: string
    total_cost: string
    currency: string
  }>
  resource_nodes: Array<{
    id: string
    name: string
    type: string
    total_cost: string
    currency: string
  }>
  shared_nodes: Array<{
    id: string
    name: string
    type: string
    total_cost: string
    currency: string
  }>
  cost_by_type: Array<{
    type: string
    total_cost: string
    node_count: number
  }>
  summary: {
    total_cost: string
    currency: string
    product_count: number
    platform_count: number
    resource_count: number
    shared_count: number
  }
}

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse<DashboardSummary | { error: string }>
) {
  if (req.method !== 'GET') {
    return res.status(405).json({ error: 'Method not allowed' })
  }

  const { start_date, end_date, currency = 'USD' } = req.query

  if (!start_date || !end_date) {
    return res.status(400).json({ error: 'start_date and end_date are required' })
  }

  try {
    // Compose multiple backend API calls in parallel
    const [productsResponse, platformResponse, resourceResponse, sharedResponse, costsByTypeResponse] = await Promise.all([
      fetch(
        `${BACKEND_URL}/api/v1/products?start_date=${start_date}&end_date=${end_date}&currency=${currency}&limit=5`
      ),
      fetch(
        `${BACKEND_URL}/api/v1/nodes?start_date=${start_date}&end_date=${end_date}&currency=${currency}&type=platform`
      ),
      fetch(
        `${BACKEND_URL}/api/v1/nodes?start_date=${start_date}&end_date=${end_date}&currency=${currency}&type=resource`
      ),
      fetch(
        `${BACKEND_URL}/api/v1/nodes?start_date=${start_date}&end_date=${end_date}&currency=${currency}&type=shared`
      ),
      fetch(
        `${BACKEND_URL}/api/v1/costs/by-type?start_date=${start_date}&end_date=${end_date}&currency=${currency}`
      ),
    ])

    if (!productsResponse.ok || !platformResponse.ok || !resourceResponse.ok || !sharedResponse.ok || !costsByTypeResponse.ok) {
      throw new Error('Failed to fetch data from backend')
    }

    const [productsData, platformData, resourceData, sharedData, costsByTypeData] = await Promise.all([
      productsResponse.json(),
      platformResponse.json(),
      resourceResponse.json(),
      sharedResponse.json(),
      costsByTypeResponse.json(),
    ])

    // Transform and compose the data for dashboard needs
    const summary: DashboardSummary = {
      top_products: productsData.nodes,
      platform_nodes: platformData.nodes || [],
      resource_nodes: resourceData.nodes || [],
      shared_nodes: sharedData.nodes || [],
      cost_by_type: costsByTypeData.aggregations,
      summary: {
        total_cost: costsByTypeData.total_cost,
        currency: costsByTypeData.currency,
        product_count:
          costsByTypeData.aggregations.find((a: any) => a.type === 'product')?.node_count || 0,
        platform_count:
          costsByTypeData.aggregations.find((a: any) => a.type === 'platform')?.node_count || 0,
        resource_count:
          costsByTypeData.aggregations.find((a: any) => a.type === 'resource')?.node_count || 0,
        shared_count:
          costsByTypeData.aggregations.find((a: any) => a.type === 'shared')?.node_count || 0,
      },
    }

    res.status(200).json(summary)
  } catch (error) {
    console.error('Dashboard API error:', error)
    res.status(500).json({ error: 'Failed to fetch dashboard data' })
  }
}

