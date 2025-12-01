import { format } from "date-fns"
import useSWR from "swr"
import { DollarSign, TrendingUp, Package, Server } from "lucide-react"
import { CostCard } from "@/components/cost-card"
import { useDateRange } from "@/context/date-range-context"
import type { TypeAggregation } from "@/types/api"
import { RecommendationsPanel } from "@/components/recommendations-panel"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatCurrency } from "@/lib/utils"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from "recharts"
import { RecommendationsResponse } from "@/types/api"

// Fetcher for SWR
const fetcher = (url: string) => fetch(url).then((res) => res.json())

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

export default function Dashboard() {
  const { dateRange } = useDateRange()

  // Fetch data directly from backend API
  const startDate = format(dateRange.from, "yyyy-MM-dd")
  const endDate = format(dateRange.to, "yyyy-MM-dd")
  const backendUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

  // Fetch all data in parallel
  const { data: productsData } = useSWR(
    `${backendUrl}/api/v1/products?start_date=${startDate}&end_date=${endDate}&currency=USD&limit=5`,
    fetcher
  )
  const { data: platformData } = useSWR(
    `${backendUrl}/api/v1/nodes?start_date=${startDate}&end_date=${endDate}&currency=USD&type=platform`,
    fetcher
  )
  const { data: resourceData } = useSWR(
    `${backendUrl}/api/v1/nodes?start_date=${startDate}&end_date=${endDate}&currency=USD&type=resource`,
    fetcher
  )
  const { data: sharedData } = useSWR(
    `${backendUrl}/api/v1/nodes?start_date=${startDate}&end_date=${endDate}&currency=USD&type=shared`,
    fetcher
  )
  const { data: costsByTypeData } = useSWR(
    `${backendUrl}/api/v1/costs/by-type?start_date=${startDate}&end_date=${endDate}&currency=USD`,
    fetcher
  )

  // Compose dashboard data from individual responses
  const data: DashboardSummary | undefined = productsData && costsByTypeData ? {
    top_products: productsData.nodes || [],
    platform_nodes: platformData?.nodes || [],
    resource_nodes: resourceData?.nodes || [],
    shared_nodes: sharedData?.nodes || [],
    cost_by_type: costsByTypeData.aggregations || [],
    summary: {
      total_cost: costsByTypeData.total_cost || "0",
      currency: costsByTypeData.currency || "USD",
      product_count:
        costsByTypeData.aggregations?.find((a: TypeAggregation) => a.type === "product")?.node_count || 0,
      platform_count:
        costsByTypeData.aggregations?.find((a: TypeAggregation) => a.type === "platform")?.node_count || 0,
      resource_count:
        costsByTypeData.aggregations?.find((a: TypeAggregation) => a.type === "resource")?.node_count || 0,
      shared_count:
        costsByTypeData.aggregations?.find((a: TypeAggregation) => a.type === "shared")?.node_count || 0,
    },
  } : undefined

  const isLoading = !productsData || !costsByTypeData
  const error: Error | null = null

  // Fetch recommendations
  const { data: recommendationsData } = useSWR<RecommendationsResponse>(
    `${backendUrl}/api/v1/recommendations?start_date=${startDate}&end_date=${endDate}&currency=USD`,
    fetcher,
    {
      refreshInterval: 300000, // Refresh every 5 minutes
      revalidateOnFocus: false,
    }
  )

  // Prepare chart data
  const topProducts = (data?.top_products || []).map((p) => ({
    id: p.id,
    name: p.name,
    cost: parseFloat(p.total_cost || "0"),
    currency: p.currency,
  }))

  const pieData = (data?.cost_by_type || []).map((ct) => ({
    name: ct.type.charAt(0).toUpperCase() + ct.type.slice(1),
    value: parseFloat(ct.total_cost || "0"),
  }))

  const COLORS = ["hsl(var(--chart-1))", "hsl(var(--chart-2))", "hsl(var(--chart-3))", "hsl(var(--chart-4))"]

  if (error) {
    return (
      <div className="container py-8">
        <div className="flex items-center justify-center h-96">
          <div className="text-lg text-destructive">Error loading dashboard</div>
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="container py-8">
        <div className="flex items-center justify-center h-96">
          <div className="text-lg text-muted-foreground">Loading dashboard...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="container py-8 space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold tracking-tight">FinOps Dashboard</h1>
          <p className="text-muted-foreground mt-2">
            Cost attribution and analysis for {format(dateRange.from, "MMM d, yyyy")} - {format(dateRange.to, "MMM d, yyyy")}
          </p>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <CostCard
          title="Total Product Cost"
          amount={data?.summary?.total_cost || "0"}
          currency={data?.summary?.currency || "USD"}
          subtitle="Sum of all product holistic costs"
          icon={<DollarSign className="h-4 w-4" />}
        />
        <CostCard
          title="Products"
          amount={data?.summary?.product_count?.toString() || "0"}
          subtitle="Active products"
          icon={<Package className="h-4 w-4" />}
          showCurrency={false}
        />
        <CostCard
          title="Platform Services"
          amount={data?.summary?.platform_count?.toString() || "0"}
          subtitle="Platform nodes"
          icon={<Server className="h-4 w-4" />}
          showCurrency={false}
        />
        <CostCard
          title="Resources"
          amount={data?.summary?.resource_count?.toString() || "0"}
          subtitle="Resource nodes"
          icon={<TrendingUp className="h-4 w-4" />}
          showCurrency={false}
        />
      </div>

      {/* Charts */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Top Products Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Top 5 Products by Cost</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={topProducts}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip
                  formatter={(value: number) => formatCurrency(value, "USD")}
                  labelStyle={{ color: "hsl(var(--foreground))" }}
                  contentStyle={{
                    backgroundColor: "hsl(var(--background))",
                    border: "1px solid hsl(var(--border))",
                  }}
                />
                <Bar dataKey="cost" fill="hsl(var(--chart-1))" />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Cost Distribution Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Cost Distribution by Type</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name}: ${(Number(percent) * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(value: number) => formatCurrency(value, "USD")}
                  contentStyle={{
                    backgroundColor: "hsl(var(--background))",
                    border: "1px solid hsl(var(--border))",
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Products Table */}
      <Card>
        <CardHeader>
          <CardTitle>Product Costs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {topProducts.map((product) => (
              <div key={product.id} className="flex items-center justify-between border-b pb-2">
                <div>
                  <div className="font-medium">{product.name}</div>
                  <div className="text-sm text-muted-foreground">Product</div>
                </div>
                <div className="text-right">
                  <div className="font-semibold">{formatCurrency(product.cost, product.currency)}</div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Cost Optimization Recommendations */}
      {recommendationsData && (
        <RecommendationsPanel
          recommendations={recommendationsData.recommendations}
          totalSavings={recommendationsData.total_savings}
          currency={recommendationsData.currency}
          highCount={recommendationsData.high_severity_count}
          mediumCount={recommendationsData.medium_severity_count}
          lowCount={recommendationsData.low_severity_count}
        />
      )}

      {/* Platform and Shared Services Section */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Platform Services */}
        <Card>
          <CardHeader>
            <CardTitle>Platform Services</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {data?.platform_nodes && data.platform_nodes.length > 0 ? (
                data.platform_nodes.map((node) => (
                  <div key={node.id} className="flex items-center justify-between border-b pb-2">
                    <div>
                      <div className="font-medium">{node.name}</div>
                      <div className="text-sm text-muted-foreground">Platform</div>
                    </div>
                    <div className="text-right">
                      <div className="font-semibold">{formatCurrency(parseFloat(node.total_cost || "0"), node.currency)}</div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-sm text-muted-foreground">No platform services found</div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Shared Services */}
        <Card>
          <CardHeader>
            <CardTitle>Shared Services</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {data?.shared_nodes && data.shared_nodes.length > 0 ? (
                data.shared_nodes.map((node) => (
                  <div key={node.id} className="flex items-center justify-between border-b pb-2">
                    <div>
                      <div className="font-medium">{node.name}</div>
                      <div className="text-sm text-muted-foreground">Shared</div>
                    </div>
                    <div className="text-right">
                      <div className="font-semibold">{formatCurrency(parseFloat(node.total_cost || "0"), node.currency)}</div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-sm text-muted-foreground">No shared services found</div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Resource Costs */}
      <Card>
        <CardHeader>
          <CardTitle>Resource Costs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {data?.resource_nodes && data.resource_nodes.length > 0 ? (
              data.resource_nodes.map((node) => (
                <div key={node.id} className="flex items-center justify-between border-b pb-2">
                  <div>
                    <div className="font-medium">{node.name}</div>
                    <div className="text-sm text-muted-foreground">Resource</div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">{formatCurrency(parseFloat(node.total_cost || "0"), node.currency)}</div>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-sm text-muted-foreground">No resources found</div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

