import { useState, useEffect } from "react"
import { format, subDays } from "date-fns"
import { Server, Share2 } from "lucide-react"
import { DateRangePicker } from "@/components/date-range-picker"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { api } from "@/lib/api"
import { formatCurrency } from "@/lib/utils"
import type { PlatformServicesResponse } from "@/types/api"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts"

export default function PlatformPage() {
  const [dateRange, setDateRange] = useState({
    from: subDays(new Date(), 30),
    to: new Date(),
  })
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

      const platform = await api.platform.getServices(params)
      setPlatformData(platform)
    } catch (error) {
      console.error("Failed to load platform data:", error)
    } finally {
      setLoading(false)
    }
  }

  const platformChartData = platformData?.platform_services.map((service) => ({
    name: service.name,
    direct: parseFloat(service.direct_costs.total || "0"),
    allocated: (service.allocated_to || []).reduce((sum, target) => sum + parseFloat(target.amount || "0"), 0),
  })) || []

  const sharedChartData = platformData?.shared_services.map((service) => ({
    name: service.name,
    direct: parseFloat(service.direct_costs.total || "0"),
    allocated: (service.weighted_targets || []).reduce((sum, target) => sum + parseFloat(target.amount || "0"), 0),
  })) || []

  if (loading) {
    return (
      <div className="container py-8">
        <div className="flex items-center justify-center h-96">
          <div className="text-lg text-muted-foreground">Loading platform services...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="container py-8 space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold tracking-tight">Platform & Shared Services</h1>
          <p className="text-muted-foreground mt-2">
            Infrastructure and shared service cost allocation
          </p>
        </div>
        <DateRangePicker value={dateRange} onChange={setDateRange} />
      </div>

      {/* Summary */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Total Platform Cost</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(platformData?.summary.total_cost || "0", platformData?.summary.currency)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Platform Services</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{platformData?.platform_services.length || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Shared Services</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{platformData?.shared_services.length || 0}</div>
          </CardContent>
        </Card>
      </div>

      {/* Platform Services */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Server className="h-5 w-5 text-primary" />
            <CardTitle>Platform Services</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Chart */}
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={platformChartData}>
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
              <Bar dataKey="direct" fill="hsl(var(--chart-1))" name="Direct" radius={[4, 4, 0, 0]} />
              <Bar dataKey="allocated" fill="hsl(var(--chart-2))" name="Allocated" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>

          <Separator />

          {/* Service List */}
          <div className="space-y-3">
            {platformData?.platform_services.map((service) => (
              <div
                key={service.id}
                className="p-4 rounded-lg border bg-card hover:bg-accent transition-colors"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-semibold">{service.name}</h3>
                      <Badge variant="secondary">{service.type}</Badge>
                    </div>
                    {service.metadata?.description && (
                      <p className="text-sm text-muted-foreground">{service.metadata.description}</p>
                    )}
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">Direct Cost</p>
                    <p className="text-xl font-bold">
                      {formatCurrency(
                        service.direct_costs.total || "0",
                        service.direct_costs.currency
                      )}
                    </p>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4 mt-3 pt-3 border-t">
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Direct</p>
                    <p className="text-sm font-semibold">
                      {formatCurrency(
                        service.direct_costs.total || "0",
                        service.direct_costs.currency
                      )}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Allocated To</p>
                    <p className="text-sm font-semibold">
                      {service.allocated_to.length} target(s)
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Dimensions</p>
                    <p className="text-sm font-semibold">
                      {Object.keys(service.direct_costs.dimensions || {}).length}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Shared Services */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Share2 className="h-5 w-5 text-primary" />
            <CardTitle>Shared Services</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Chart */}
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={sharedChartData}>
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
              <Bar dataKey="direct" fill="hsl(var(--chart-3))" name="Direct" radius={[4, 4, 0, 0]} />
              <Bar dataKey="allocated" fill="hsl(var(--chart-4))" name="Allocated" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>

          <Separator />

          {/* Service List */}
          <div className="space-y-3">
            {platformData?.shared_services.map((service) => (
              <div
                key={service.id}
                className="p-4 rounded-lg border bg-card hover:bg-accent transition-colors"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-semibold">{service.name}</h3>
                      <Badge variant="outline">{service.type}</Badge>
                    </div>
                    {service.metadata?.description && (
                      <p className="text-sm text-muted-foreground">{service.metadata.description}</p>
                    )}
                    {service.weighted_targets && service.weighted_targets.length > 0 && (
                      <div className="flex items-center gap-2 mt-2">
                        <span className="text-xs text-muted-foreground">
                          Allocated to {service.weighted_targets.length} target(s)
                        </span>
                      </div>
                    )}
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">Direct Cost</p>
                    <p className="text-xl font-bold">
                      {formatCurrency(
                        service.direct_costs.total || "0",
                        service.direct_costs.currency
                      )}
                    </p>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4 mt-3 pt-3 border-t">
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Direct</p>
                    <p className="text-sm font-semibold">
                      {formatCurrency(
                        service.direct_costs.total || "0",
                        service.direct_costs.currency
                      )}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Allocated To</p>
                    <p className="text-sm font-semibold">
                      {service.weighted_targets.length} target(s)
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Dimensions</p>
                    <p className="text-sm font-semibold">
                      {Object.keys(service.direct_costs.dimensions || {}).length}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

    </div>
  )
}

