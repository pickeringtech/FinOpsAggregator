import { useState, useEffect } from "react"
import { format } from "date-fns"
import { ProductTree } from "@/components/product-tree"
import { NodeDetailPanel } from "@/components/node-detail-panel"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Info, HelpCircle } from "lucide-react"
import { api } from "@/lib/api"
import { useDateRange } from "@/context/date-range-context"
import { formatCurrency } from "@/lib/utils"
import type { ProductHierarchyResponse, ProductNode, IndividualNodeResponse } from "@/types/api"

export default function ProductsPage() {
  const { dateRange } = useDateRange()
  const [hierarchyData, setHierarchyData] = useState<ProductHierarchyResponse | null>(null)
  const [selectedNode, setSelectedNode] = useState<ProductNode | null>(null)
  const [nodeDetails, setNodeDetails] = useState<IndividualNodeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingDetails, setLoadingDetails] = useState(false)

  useEffect(() => {
    loadHierarchy()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dateRange])

  useEffect(() => {
    if (selectedNode) {
      loadNodeDetails(selectedNode.id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedNode, dateRange])

  const loadHierarchy = async () => {
    setLoading(true)
    try {
      const params = {
        start_date: format(dateRange.from, "yyyy-MM-dd"),
        end_date: format(dateRange.to, "yyyy-MM-dd"),
        currency: "USD",
      }

      const hierarchy = await api.products.getHierarchy(params)
      setHierarchyData(hierarchy)

      // Auto-select first product if none selected
      if (!selectedNode && hierarchy.products.length > 0) {
        setSelectedNode(hierarchy.products[0])
      }
    } catch (error) {
      console.error("Failed to load hierarchy:", error)
    } finally {
      setLoading(false)
    }
  }

  const loadNodeDetails = async (nodeId: string) => {
    setLoadingDetails(true)
    try {
      const params = {
        start_date: format(dateRange.from, "yyyy-MM-dd"),
        end_date: format(dateRange.to, "yyyy-MM-dd"),
        currency: "USD",
      }

      const details = await api.nodes.getDetails(nodeId, params)
      setNodeDetails(details)
    } catch (error) {
      console.error("Failed to load node details:", error)
    } finally {
      setLoadingDetails(false)
    }
  }

  const handleNodeSelect = (node: ProductNode) => {
    setSelectedNode(node)
  }

  if (loading) {
    return (
      <div className="container py-8">
        <div className="flex items-center justify-center h-96">
          <div className="text-lg text-muted-foreground">Loading products...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="container py-8 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold tracking-tight">Product Hierarchy</h1>
          <p className="text-muted-foreground mt-2">
            Explore product costs with direct and rolled-up allocations
          </p>
        </div>
      </div>

      {/* Cost Explanation */}
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>
          <strong>Holistic Product Cost</strong> ({formatCurrency(hierarchyData?.summary.total_cost || "0", hierarchyData?.summary.currency)}) is the sum of direct + indirect costs for all final cost centres (products with no child products).
          <br />
          <strong>Source Infrastructure Cost</strong> is the total direct cost from infrastructure nodes before allocation.
          <br />
          Allocation coverage shows what percentage of source infrastructure has been attributed to products.
        </AlertDescription>
      </Alert>

      {/* Summary Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex items-center justify-between pb-3">
            <CardTitle className="text-sm font-medium">Holistic Product Cost</CardTitle>
            <Dialog>
              <DialogTrigger asChild>
                <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground">
                  <HelpCircle className="h-4 w-4" />
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Understanding Cost Terminology</DialogTitle>
                  <DialogDescription asChild>
                    <div className="space-y-3 text-sm">
                      <p>
                        <strong>Direct Cost</strong> is the cost that originates on a node before any allocation occurs.
                        For infrastructure nodes, this is the cloud spend. For products, this is any direct charges.
                      </p>
                      <p>
                        <strong>Indirect Cost</strong> is the cost received via allocation from parent nodes
                        (infrastructure, platform, shared services).
                      </p>
                      <p>
                        <strong>Holistic Cost</strong> = Direct Cost + Indirect Cost. This is the total cost
                        attributed to a node after allocation.
                      </p>
                      <p>
                        <strong>Final Cost Centres</strong> are product nodes with no child products - the &quot;leaves&quot;
                        of the product tree. Summing their holistic costs gives the total product cost without double-counting.
                      </p>
                      <p>
                        <strong>Source Infrastructure</strong> is the total direct cost from all infrastructure nodes
                        (platform, shared, resource) before allocation.
                      </p>
                      <p className="text-muted-foreground">
                        The allocation coverage % shows how much of source infrastructure has been attributed to products.
                        Less than 100% indicates unallocated costs that need attention.
                      </p>
                    </div>
                  </DialogDescription>
                </DialogHeader>
              </DialogContent>
            </Dialog>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(hierarchyData?.summary.total_cost || "0", hierarchyData?.summary.currency)}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Sum of final cost centre holistic costs
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Source Infrastructure</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(hierarchyData?.summary.raw_total_cost || "0", hierarchyData?.summary.currency)}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">Direct cost from infrastructure nodes</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Allocation Coverage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {hierarchyData ? `${hierarchyData.summary.allocation_coverage_percent.toFixed(1)}%` : "0.0%"}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">% of source infra allocated to products</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Final Cost Centres</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{hierarchyData?.summary.product_count || 0}</div>
            <p className="mt-1 text-xs text-muted-foreground">Products with no child products</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Total Nodes</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{hierarchyData?.summary.node_count || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Period</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm font-medium">{hierarchyData?.summary.period}</div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content: Tree + Detail Panel */}
      <div className="grid gap-6 lg:grid-cols-[400px_1fr]">
        {/* Tree View */}
        <Card className="h-[calc(100vh-400px)]">
          <CardHeader>
            <CardTitle>Product Tree</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <ScrollArea className="h-[calc(100vh-480px)] px-6 pb-6">
              {hierarchyData && (
                <ProductTree
                  nodes={hierarchyData.products}
                  selectedNodeId={selectedNode?.id}
                  onNodeSelect={handleNodeSelect}
                />
              )}
            </ScrollArea>
          </CardContent>
        </Card>

        {/* Detail Panel */}
        <div className="h-[calc(100vh-400px)]">
          <ScrollArea className="h-full">
            {loadingDetails ? (
              <Card>
                <CardContent className="flex items-center justify-center h-96">
                  <div className="text-muted-foreground">Loading details...</div>
                </CardContent>
              </Card>
            ) : nodeDetails ? (
              <NodeDetailPanel nodeData={nodeDetails} />
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

