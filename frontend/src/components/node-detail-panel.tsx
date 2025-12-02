import { IndividualNodeResponse } from "@/types/api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { formatCurrency } from "@/lib/utils"
import { Box, Server, Share2, Folder, ArrowRight } from "lucide-react"
import { MetricsTimeSeriesChart } from "@/components/metrics-time-series-chart"

interface NodeDetailPanelProps {
  nodeData: IndividualNodeResponse
}

export function NodeDetailPanel({ nodeData }: NodeDetailPanelProps) {
  const { node, direct_costs, allocated_costs, total_costs, dependencies, allocations } = nodeData

  const getNodeIcon = () => {
    switch (node.type) {
      case "product":
        return <Folder className="h-5 w-5" />
      case "resource":
        return <Box className="h-5 w-5" />
      case "platform":
        return <Server className="h-5 w-5" />
      case "shared":
        return <Share2 className="h-5 w-5" />
    }
  }

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
                {node.is_platform && <Badge variant="secondary">Platform</Badge>}
              </div>
            </div>
          </div>
        </CardHeader>
        {node.metadata?.description && (
          <CardContent>
            <p className="text-sm text-muted-foreground">{node.metadata.description}</p>
          </CardContent>
        )}
      </Card>

      {/* Cost Summary */}
      <Card>
        <CardHeader>
          <CardTitle>Cost Summary</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-3 gap-4">
            <div>
              <p className="text-sm text-muted-foreground mb-1">Direct Costs</p>
              <p className="text-2xl font-bold">
                {formatCurrency(direct_costs.total, direct_costs.currency)}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground mb-1">Allocated Costs</p>
              <p className="text-2xl font-bold">
                {formatCurrency(allocated_costs.total, allocated_costs.currency)}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground mb-1">Total Costs</p>
              <p className="text-2xl font-bold text-primary">
                {formatCurrency(total_costs.total, total_costs.currency)}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Cost Dimensions */}
      <Card>
        <CardHeader>
          <CardTitle>Cost Dimensions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Object.entries(direct_costs.dimensions).map(([dimension, value]) => (
              <div key={dimension} className="flex justify-between items-center">
                <span className="text-sm font-medium">{dimension.replace(/_/g, " ")}</span>
                <span className="text-sm font-semibold">
                  {formatCurrency(value, direct_costs.currency)}
                </span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Dependencies */}
      {dependencies.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Dependencies</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {dependencies.map((dep) => (
                <div
                  key={dep.id}
                  className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <Badge variant="outline">{dep.relationship_type}</Badge>
                    <span className="font-medium">{dep.name}</span>
                    <Badge variant="secondary" className="text-xs">
                      {dep.type}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <span>{dep.strategy}</span>
                    <ArrowRight className="h-4 w-4" />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Allocations */}
      {allocations && allocations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Cost Allocations</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {allocations.map((allocation, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between p-3 rounded-lg border bg-card"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-sm font-medium">{allocation.from_node.name}</span>
                      <ArrowRight className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm font-medium">{allocation.to_node.name}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">
                        {allocation.strategy}
                      </Badge>
                      {allocation.dimension && (
                        <span className="text-xs text-muted-foreground">
                          {allocation.dimension}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-semibold">
                      {formatCurrency(allocation.amount, allocation.currency)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Cost Labels */}
      {node.cost_labels && Object.keys(node.cost_labels).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Cost Labels</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3">
              {Object.entries(node.cost_labels).map(([key, value]) => (
                <div key={key} className="flex flex-col">
                  <span className="text-xs text-muted-foreground uppercase">{key}</span>
                  <span className="text-sm font-medium">{value}</span>
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

