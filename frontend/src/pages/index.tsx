import { useState, useEffect } from "react"
import { format, subDays } from "date-fns"
import { DollarSign, TrendingUp, Package, Server } from "lucide-react"
import { CostCard } from "@/components/cost-card"
import { DateRangePicker } from "@/components/date-range-picker"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { api } from "@/lib/api"
import { formatCurrency } from "@/lib/utils"
import type { ProductHierarchyResponse, PlatformServicesResponse, ProductNode } from "@/types/api"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from "recharts"

export default function Dashboard() {
  const [dateRange, setDateRange] = useState({
    from: subDays(new Date(), 30),
    to: new Date(),
  })
  const [hierarchyData, setHierarchyData] = useState<ProductHierarchyResponse | null>(null)
  const [platformData, setPlatformData] = useState<PlatformServicesResponse | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dateRange])

  const loadData = async () => {
    setLoading(true)
    try {
      const params = {
        start_date: format(dateRange.from, "yyyy-MM-dd"),
        end_date: format(dateRange.to, "yyyy-MM-dd"),
        currency: "USD",
      }

      const [hierarchy, platform] = await Promise.all([
        api.products.getHierarchy(params),
        api.platform.getServices(params),
      ])

      setHierarchyData(hierarchy)
      setPlatformData(platform)
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }

  // ⚠️ TEMPORARY WORKAROUND - THIS SHOULD BE DONE IN THE BACKEND! ⚠️
  // TODO: Replace with /api/v1/dashboard/summary endpoint
  // See: BACKEND_API_REQUIREMENTS.md for details
  //
  // This client-side aggregation is NOT scalable and should be replaced
  // with proper database queries that return pre-aggregated data.
  //
  // Current issues:
  // - Flattening large hierarchies is slow
  // - Deduplication should not be needed if backend returns correct data
  // - Aggregation should happen in SQL, not JavaScript
  //
  // Flatten the hierarchy to get all nodes
  const flattenNodes = (nodes: ProductNode[]): ProductNode[] => {
    const result: ProductNode[] = []
    const traverse = (node: ProductNode) => {
      result.push(node)
      if (node.children) {
        node.children.forEach(traverse)
      }
    }
    nodes.forEach(traverse)
    return result
  }

  const allNodes = hierarchyData?.products ? flattenNodes(hierarchyData.products) : []

  // Deduplicate nodes by ID and create a map for quick lookup
  const uniqueNodesMap = new Map<string, ProductNode>()
  allNodes.forEach((node) => {
    if (!uniqueNodesMap.has(node.id)) {
      uniqueNodesMap.set(node.id, node)
    }
  })
  const uniqueNodes = Array.from(uniqueNodesMap.values())

  const topProducts = uniqueNodes
    .filter((p) => p.holistic_costs?.total) // Filter out nodes without costs
    .map((p) => ({
      id: p.id,
      name: p.name,
      cost: parseFloat(p.holistic_costs.total || "0"),
      currency: p.holistic_costs.currency,
    }))
    .sort((a, b) => b.cost - a.cost)
    .slice(0, 5)

  const costByType = uniqueNodes
    .filter((p) => p.holistic_costs?.total) // Filter out nodes without costs
    .reduce((acc, product) => {
      const type = product.type
      const cost = parseFloat(product.holistic_costs.total || "0")
      acc[type] = (acc[type] || 0) + cost
      return acc
    }, {} as Record<string, number>)

  const pieData = Object.entries(costByType).map(([name, value]) => ({
    name: name.charAt(0).toUpperCase() + name.slice(1),
    value,
  }))

  const COLORS = ["hsl(var(--chart-1))", "hsl(var(--chart-2))", "hsl(var(--chart-3))", "hsl(var(--chart-4))"]

  if (loading) {
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
            Overview of your cloud costs and resource allocation
          </p>
        </div>
        <DateRangePicker value={dateRange} onChange={setDateRange} />
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <CostCard
          title="Total Costs"
          amount={hierarchyData?.summary.total_cost || "0"}
          currency={hierarchyData?.summary.currency}
          subtitle={hierarchyData?.summary.period}
          icon={<DollarSign className="h-4 w-4" />}
        />
        <CostCard
          title="Product Count"
          amount={hierarchyData?.summary.product_count.toString() || "0"}
          subtitle="Active products"
          icon={<Package className="h-4 w-4" />}
          showCurrency={false}
        />
        <CostCard
          title="Platform Services"
          amount={platformData?.platform_services.length.toString() || "0"}
          subtitle="Shared infrastructure"
          icon={<Server className="h-4 w-4" />}
          showCurrency={false}
        />
        <CostCard
          title="Platform Costs"
          amount={platformData?.summary.total_cost || "0"}
          currency={platformData?.summary.currency}
          subtitle="Shared service costs"
          icon={<TrendingUp className="h-4 w-4" />}
        />
      </div>

      {/* Charts Row */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Top Products */}
        <Card>
          <CardHeader>
            <CardTitle>Top 5 Products by Cost</CardTitle>
          </CardHeader>
          <CardContent>
            {topProducts.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={topProducts}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis dataKey="name" className="text-xs" />
                  <YAxis className="text-xs" />
                  <Tooltip
                    formatter={(value: number) => formatCurrency(value, "USD")}
                    contentStyle={{
                      backgroundColor: "hsl(var(--card))",
                      border: "1px solid hsl(var(--border))",
                      borderRadius: "8px",
                    }}
                  />
                  <Bar dataKey="cost" fill="hsl(var(--primary))" radius={[8, 8, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                No product data available
              </div>
            )}
          </CardContent>
        </Card>

        {/* Cost Distribution */}
        <Card>
          <CardHeader>
            <CardTitle>Cost Distribution by Type</CardTitle>
          </CardHeader>
          <CardContent>
            {pieData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={pieData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    label={(entry: any) =>
                      `${entry.name} ${(entry.percent * 100).toFixed(0)}%`
                    }
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
                      backgroundColor: "hsl(var(--card))",
                      border: "1px solid hsl(var(--border))",
                      borderRadius: "8px",
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                No cost distribution data available
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Recent Products */}
      <Card>
        <CardHeader>
          <CardTitle>All Products</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {hierarchyData?.products && hierarchyData.products.length > 0 ? (
              hierarchyData.products.map((product) => (
                <div
                  key={product.id}
                  className="flex items-center justify-between p-4 rounded-lg border bg-card hover:bg-accent transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-4">
                    <div>
                      <p className="font-semibold">{product.name}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge variant="outline">{product.type}</Badge>
                        {product.children && product.children.length > 0 && (
                          <span className="text-xs text-muted-foreground">
                            {product.children.length} child nodes
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">Holistic Cost</p>
                    <p className="text-lg font-bold">
                      {formatCurrency(
                        product.holistic_costs?.total || "0",
                        product.holistic_costs?.currency
                      )}
                    </p>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                No products available
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
