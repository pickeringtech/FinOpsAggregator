import { InfrastructureNode } from "@/types/api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { formatCurrency } from "@/lib/utils"
import { Box, Server, Share2, Database, ArrowRight, AlertTriangle, CheckCircle } from "lucide-react"
import { cn } from "@/lib/utils"
import { MetricsTimeSeriesChart } from "@/components/metrics-time-series-chart"

interface InfrastructureDetailPanelProps {
  node: InfrastructureNode
}

export function InfrastructureDetailPanel({ node }: InfrastructureDetailPanelProps) {
  const getNodeIcon = () => {
    switch (node.type) {
      case "platform":
        return <Server className="h-5 w-5" />
      case "shared":
        return <Share2 className="h-5 w-5" />
      case "resource":
        return <Box className="h-5 w-5" />
      case "infrastructure":
        return <Database className="h-5 w-5" />
      default:
        return <Box className="h-5 w-5" />
    }
  }

  // Determine allocation status
  const allocationStatus = node.allocation_pct >= 90 
    ? { label: "Fully Allocated", color: "text-green-600", icon: CheckCircle }
    : node.allocation_pct >= 50 
      ? { label: "Partially Allocated", color: "text-yellow-600", icon: AlertTriangle }
      : { label: "Under-Allocated", color: "text-red-600", icon: AlertTriangle }

  const StatusIcon = allocationStatus.icon

  return (
    <div className="space-y-4">
      {/* Node Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="text-primary">{getNodeIcon()}</div>
            <div className="flex-1">
              <CardTitle className="text-2xl">{node.name}</CardTitle>
              <div className="flex items-center gap-2 mt-2">
                <Badge>{node.type}</Badge>
                {node.is_platform && <Badge variant="secondary">Platform Service</Badge>}
              </div>
            </div>
          </div>
        </CardHeader>
        {node.metadata?.description && (
          <CardContent>
            <p className="text-sm text-muted-foreground">{String(node.metadata.description)}</p>
          </CardContent>
        )}
      </Card>

      {/* Allocation Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <StatusIcon className={cn("h-5 w-5", allocationStatus.color)} />
            Allocation Status
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <span className={cn("text-lg font-semibold", allocationStatus.color)}>
              {allocationStatus.label}
            </span>
            <span className="text-2xl font-bold">{node.allocation_pct.toFixed(1)}%</span>
          </div>
          <Progress value={node.allocation_pct} className="h-3" />
          <div className="grid grid-cols-3 gap-4 pt-2">
            <div>
              <p className="text-sm text-muted-foreground mb-1">Direct Cost</p>
              <p className="text-xl font-bold">
                {formatCurrency(node.direct_costs.total, node.direct_costs.currency)}
              </p>
              <p className="text-xs text-muted-foreground">Cost originating here</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground mb-1">Allocated Out</p>
              <p className="text-xl font-bold text-green-600">
                {formatCurrency(node.allocated_costs.total, node.allocated_costs.currency)}
              </p>
              <p className="text-xs text-muted-foreground">Sent to products</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground mb-1">Unallocated</p>
              <p className={cn("text-xl font-bold", parseFloat(node.unallocated_cost) > 0 ? "text-red-600" : "text-green-600")}>
                {formatCurrency(node.unallocated_cost, node.direct_costs.currency)}
              </p>
              <p className="text-xs text-muted-foreground">Not yet attributed</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Cost Dimensions */}
      {Object.keys(node.direct_costs.dimensions).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Cost Breakdown by Dimension</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {Object.entries(node.direct_costs.dimensions).map(([dimension, value]) => (
                <div key={dimension} className="flex justify-between items-center">
                  <span className="text-sm font-medium">{dimension.replace(/_/g, " ")}</span>
                  <span className="text-sm font-semibold">
                    {formatCurrency(value, node.direct_costs.currency)}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Allocation Targets */}
      {node.allocated_to && node.allocated_to.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Allocated To Products</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {node.allocated_to.map((target) => (
                <div
                  key={target.id}
                  className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <ArrowRight className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">{target.name}</span>
                    <Badge variant="secondary" className="text-xs">
                      {target.type}
                    </Badge>
                    <Badge variant="outline" className="text-xs">
                      {target.strategy}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-sm text-muted-foreground">
                      {target.percent.toFixed(1)}%
                    </span>
                    <span className="text-sm font-semibold min-w-[100px] text-right">
                      {formatCurrency(target.amount, target.currency)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* No Allocations Warning */}
      {(!node.allocated_to || node.allocated_to.length === 0) && (
        <Card className="border-yellow-500/50 bg-yellow-500/5">
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <AlertTriangle className="h-5 w-5 text-yellow-600" />
              <div>
                <p className="font-medium text-yellow-600">No Allocation Targets</p>
                <p className="text-sm text-muted-foreground">
                  This infrastructure node has no products configured to receive its costs.
                  Consider setting up allocation edges to attribute these costs.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Metadata */}
      {node.metadata && Object.keys(node.metadata).filter(k => k !== 'description').length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Metadata</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3">
              {Object.entries(node.metadata)
                .filter(([key]) => key !== 'description')
                .map(([key, value]) => (
                  <div key={key} className="flex flex-col">
                    <span className="text-xs text-muted-foreground uppercase">{key}</span>
                    <span className="text-sm font-medium">{String(value)}</span>
                  </div>
                ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Metrics Over Time Chart */}
      <MetricsTimeSeriesChart nodeId={node.id} nodeName={node.name} />
    </div>
  )
}

