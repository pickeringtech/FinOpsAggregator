import { useState, useEffect } from "react"
import { format } from "date-fns"
import { InfrastructureTree } from "@/components/infrastructure-tree"
import { InfrastructureDetailPanel } from "@/components/infrastructure-detail-panel"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { Info, HelpCircle, Server, Share2, Box } from "lucide-react"
import { api } from "@/lib/api"
import { useDateRange } from "@/context/date-range-context"
import { formatCurrency } from "@/lib/utils"
import type { InfrastructureHierarchyResponse, InfrastructureNode } from "@/types/api"

export default function PlatformPage() {
  const { dateRange } = useDateRange()
  const [hierarchyData, setHierarchyData] = useState<InfrastructureHierarchyResponse | null>(null)
  const [selectedNode, setSelectedNode] = useState<InfrastructureNode | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadHierarchy()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dateRange])

  const loadHierarchy = async () => {
    setLoading(true)
    try {
      const params = {
        start_date: format(dateRange.from, "yyyy-MM-dd"),
        end_date: format(dateRange.to, "yyyy-MM-dd"),
        currency: "USD",
      }

      const hierarchy = await api.infrastructure.getHierarchy(params)
      setHierarchyData(hierarchy)

      // Auto-select first node if none selected
      if (!selectedNode && hierarchy.infrastructure && hierarchy.infrastructure.length > 0) {
        setSelectedNode(hierarchy.infrastructure[0])
      }
    } catch (error) {
      console.error("Failed to load infrastructure hierarchy:", error)
    } finally {
      setLoading(false)
    }
  }

  const handleNodeSelect = (node: InfrastructureNode) => {
    setSelectedNode(node)
  }

  if (loading) {
    return (
      <div className="container py-8">
        <div className="flex items-center justify-center h-96">
          <div className="text-lg text-muted-foreground">Loading infrastructure...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="container py-8 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold tracking-tight">Infrastructure & Shared Services</h1>
          <p className="text-muted-foreground mt-2">
            Explore infrastructure costs and how they flow to products
          </p>
        </div>
      </div>

      {/* Cost Explanation */}
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>
          <strong>Direct Cost</strong> is the cost originating on each infrastructure node before allocation.
          <br />
          <strong>Allocated Cost</strong> shows how much has been attributed to products.
          <br />
          <strong>Unallocated Cost</strong> represents costs not yet attributed to any product (data quality issue to address).
        </AlertDescription>
      </Alert>

      {/* Summary Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex items-center justify-between pb-3">
            <CardTitle className="text-sm font-medium">Total Direct Cost</CardTitle>
            <Dialog>
              <DialogTrigger asChild>
                <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground">
                  <HelpCircle className="h-4 w-4" />
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Understanding Infrastructure Costs</DialogTitle>
                  <DialogDescription asChild>
                    <div className="space-y-3 text-sm">
                      <p>
                        <strong>Direct Cost</strong> is the total cost originating on infrastructure nodes
                        (platform services, shared services, and resources) before any allocation.
                      </p>
                      <p>
                        <strong>Allocated Cost</strong> is the portion of direct cost that has been
                        attributed to products through the allocation graph.
                      </p>
                      <p>
                        <strong>Unallocated Cost</strong> is the gap - costs that haven&apos;t been
                        attributed to any product yet. This typically indicates:
                      </p>
                      <ul className="list-disc pl-6 space-y-1">
                        <li>Missing allocation edges in the graph</li>
                        <li>Allocation not yet run for this period</li>
                        <li>New infrastructure not yet connected</li>
                      </ul>
                    </div>
                  </DialogDescription>
                </DialogHeader>
              </DialogContent>
            </Dialog>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(hierarchyData?.summary.total_direct_cost || "0", hierarchyData?.summary.currency)}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Pre-allocation infrastructure spend
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Allocated to Products</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {formatCurrency(hierarchyData?.summary.total_allocated_cost || "0", hierarchyData?.summary.currency)}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">Successfully attributed</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Unallocated</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {formatCurrency(hierarchyData?.summary.total_unallocated_cost || "0", hierarchyData?.summary.currency)}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">Not yet attributed</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Allocation Coverage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {hierarchyData ? `${hierarchyData.summary.allocation_pct.toFixed(1)}%` : "0.0%"}
            </div>
            <Progress value={hierarchyData?.summary.allocation_pct || 0} className="h-2 mt-2" />
          </CardContent>
        </Card>
      </div>

      {/* Node Type Counts */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Server className="h-4 w-4" />
              Platform Services
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{hierarchyData?.summary.platform_count || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Share2 className="h-4 w-4" />
              Shared Services
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{hierarchyData?.summary.shared_count || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Box className="h-4 w-4" />
              Resources
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{hierarchyData?.summary.resource_count || 0}</div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content: Tree + Detail Panel */}
      <div className="grid gap-6 lg:grid-cols-[400px_1fr]">
        {/* Tree View */}
        <Card className="h-[calc(100vh-500px)]">
          <CardHeader>
            <CardTitle>Infrastructure Tree</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <ScrollArea className="h-[calc(100vh-580px)] px-6 pb-6">
              {hierarchyData && hierarchyData.infrastructure && (
                <InfrastructureTree
                  nodes={hierarchyData.infrastructure}
                  selectedNodeId={selectedNode?.id}
                  onNodeSelect={handleNodeSelect}
                />
              )}
              {hierarchyData && !hierarchyData.infrastructure && (
                <div className="text-center text-muted-foreground py-8">
                  No infrastructure nodes found. Run an allocation to see infrastructure data.
                </div>
              )}
            </ScrollArea>
          </CardContent>
        </Card>

        {/* Detail Panel */}
        <div className="h-[calc(100vh-500px)]">
          <ScrollArea className="h-full">
            {selectedNode ? (
              <InfrastructureDetailPanel node={selectedNode} />
            ) : (
              <Card>
                <CardContent className="flex items-center justify-center h-96">
                  <div className="text-muted-foreground">Select a node to view details</div>
                </CardContent>
              </Card>
            )}
          </ScrollArea>
        </div>
      </div>
    </div>
  )
}
