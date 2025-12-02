import { useState, useEffect } from "react"
import { format } from "date-fns"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { api } from "@/lib/api"
import { formatCurrency } from "@/lib/utils"
import { useDateRange } from "@/context/date-range-context"
import type { NodeMetricsTimeSeriesResponse } from "@/types/api"
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts"

interface MetricsTimeSeriesChartProps {
  nodeId: string
  nodeName: string
}

interface ChartDataPoint {
  date: string
  cost: number
  [key: string]: string | number // For dynamic metric keys
}

const METRIC_COLORS = [
  "hsl(var(--chart-2))",
  "hsl(var(--chart-3))",
  "hsl(var(--chart-4))",
  "hsl(var(--chart-5))",
]

export function MetricsTimeSeriesChart({ nodeId, nodeName }: MetricsTimeSeriesChartProps) {
  const { dateRange } = useDateRange()
  const [data, setData] = useState<NodeMetricsTimeSeriesResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodeId, dateRange])

  const loadData = async () => {
    setLoading(true)
    setError(null)
    try {
      const params = {
        start_date: format(dateRange.from, "yyyy-MM-dd"),
        end_date: format(dateRange.to, "yyyy-MM-dd"),
        currency: "USD",
      }

      const response = await api.metrics.getTimeSeries(nodeId, params)
      setData(response)
    } catch (err) {
      console.error("Failed to load metrics time series:", err)
      setError("Failed to load metrics data")
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Cost & Usage Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-muted-foreground">Loading metrics...</div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Cost & Usage Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-destructive">{error}</div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (!data || ((data.cost_series?.length ?? 0) === 0 && (data.usage_series?.length ?? 0) === 0)) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Cost & Usage Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-muted-foreground">No data available for this period</div>
          </div>
        </CardContent>
      </Card>
    )
  }

  // Merge cost and usage data by date
  const chartData: ChartDataPoint[] = []
  const dateMap = new Map<string, ChartDataPoint>()

  // Add cost data (with null check)
  if (data.cost_series && Array.isArray(data.cost_series)) {
    for (const costPoint of data.cost_series) {
      const dateKey = costPoint.date.split("T")[0]
      if (!dateMap.has(dateKey)) {
        dateMap.set(dateKey, { date: dateKey, cost: 0 })
      }
      const point = dateMap.get(dateKey)!
      point.cost = parseFloat(costPoint.total_cost)
    }
  }

  // Add usage data (with null check)
  if (data.usage_series && Array.isArray(data.usage_series)) {
    for (const usagePoint of data.usage_series) {
      const dateKey = usagePoint.date.split("T")[0]
      if (!dateMap.has(dateKey)) {
        dateMap.set(dateKey, { date: dateKey, cost: 0 })
      }
      const point = dateMap.get(dateKey)!
      for (const [metric, value] of Object.entries(usagePoint.metrics)) {
        point[metric] = parseFloat(value)
      }
    }
  }

  // Sort by date
  const sortedDates = Array.from(dateMap.keys()).sort()
  for (const date of sortedDates) {
    chartData.push(dateMap.get(date)!)
  }

  // Get available metrics
  const availableMetrics = data.metrics || []

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Cost & Usage Over Time</CardTitle>
          <div className="flex gap-2">
            {availableMetrics.length > 0 && (
              <Badge variant="outline" className="text-xs">
                {availableMetrics.length} metric{availableMetrics.length !== 1 ? "s" : ""}
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis
              dataKey="date"
              className="text-xs"
              tickFormatter={(value) => {
                const date = new Date(value)
                return format(date, "MMM d")
              }}
            />
            <YAxis
              yAxisId="cost"
              orientation="left"
              className="text-xs"
              tickFormatter={(value) => `$${(value / 1000).toFixed(0)}k`}
            />
            {availableMetrics.length > 0 && (
              <YAxis
                yAxisId="usage"
                orientation="right"
                className="text-xs"
              />
            )}
            <Tooltip
              contentStyle={{
                backgroundColor: "hsl(var(--card))",
                border: "1px solid hsl(var(--border))",
                borderRadius: "8px",
              }}
              formatter={(value: number, name: string) => {
                if (name === "cost") {
                  return [formatCurrency(value, "USD"), "Cost"]
                }
                return [value.toLocaleString(), name]
              }}
              labelFormatter={(label) => format(new Date(label), "MMM d, yyyy")}
            />
            <Legend />
            <Line
              yAxisId="cost"
              type="monotone"
              dataKey="cost"
              stroke="hsl(var(--chart-1))"
              strokeWidth={2}
              dot={false}
              name="Cost"
            />
            {availableMetrics.map((metric, index) => (
              <Line
                key={metric}
                yAxisId="usage"
                type="monotone"
                dataKey={metric}
                stroke={METRIC_COLORS[index % METRIC_COLORS.length]}
                strokeWidth={2}
                dot={false}
                name={metric}
              />
            ))}
          </LineChart>
        </ResponsiveContainer>

        {/* Dimension breakdown */}
        {data.dimensions && data.dimensions.length > 0 && (
          <div className="mt-4 pt-4 border-t">
            <p className="text-sm font-medium mb-2">Cost Dimensions</p>
            <div className="flex flex-wrap gap-2">
              {data.dimensions.map((dim) => (
                <Badge key={dim} variant="secondary" className="text-xs">
                  {dim}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

