import { useState, useEffect } from "react"
import { format, subDays } from "date-fns"
import { DollarSign, TrendingUp, Package, Server } from "lucide-react"
import { CostCard } from "@/components/cost-card"
import { DateRangePicker } from "@/components/date-range-picker"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { api } from "@/lib/api"
import { formatCurrency } from "@/lib/utils"
import type { DashboardSummaryResponse, PlatformServicesResponse } from "@/types/api"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from "recharts"

export default function Dashboard() {
  const [dateRange, setDateRange] = useState({
    from: subDays(new Date(), 30),
    to: new Date(),
  })
  const [dashboardData, setDashboardData] = useState<DashboardSummaryResponse | null>(null)
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

      const [dashboard, platform] = await Promise.all([
        api.dashboard.getSummary(params),
        api.platform.getServices(params),
      ])

      setDashboardData(dashboard)
      setPlatformData(platform)
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }

  // Prepare chart data from the pre-aggregated backend response
  const topProducts = dashboardData?.top_products.map((p) => ({
    id: p.id,
    name: p.name,
    cost: parseFloat(p.total_cost.total || "0"),
    currency: p.total_cost.currency,
  })) || []

  const pieData = dashboardData?.cost_by_type.map((ct) => ({
    name: ct.type.charAt(0).toUpperCase() + ct.type.slice(1),
    value: parseFloat(ct.total_cost.total || "0"),
  })) || []

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
          amount={dashboardData?.summary.total_cost || "0"}
          currency={dashboardData?.summary.currency}
          subtitle={dashboardData?.summary.period}
          icon={<DollarSign className="h-4 w-4" />}
        />
        <CostCard
          title="Product Count"
          amount={dashboardData?.summary.product_count.toString() || "0"}
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

      {/* All Products Summary */}
      <Card>
        <CardHeader>
          <CardTitle>All Products</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {dashboardData?.top_products && dashboardData.top_products.length > 0 ? (
              dashboardData.top_products.map((product) => (
                <div
                  key={product.id}
                  className="flex items-center justify-between p-4 rounded-lg border bg-card hover:bg-accent transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-4">
                    <div>
                      <p className="font-semibold">{product.name}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">Total Cost</p>
                    <p className="text-lg font-bold">
                      {formatCurrency(
                        product.total_cost.total || "0",
                        product.total_cost.currency
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
