import { useState, useEffect } from "react"
import { format, subDays } from "date-fns"
import { ProductTree } from "@/components/product-tree"
import { NodeDetailPanel } from "@/components/node-detail-panel"
import { DateRangePicker } from "@/components/date-range-picker"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Info, HelpCircle } from "lucide-react"
import { api } from "@/lib/api"
import { formatCurrency } from "@/lib/utils"
import type { ProductHierarchyResponse, ProductNode, IndividualNodeResponse } from "@/types/api"

export default function ProductsPage() {
  const [dateRange, setDateRange] = useState({
    from: subDays(new Date(), 30),
    to: new Date(),
  })
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
        <DateRangePicker value={dateRange} onChange={setDateRange} />
      </div>

      {/* Cost Explanation */}
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>
          <strong>Allocated Product Cost</strong> ({formatCurrency(hierarchyData?.summary.total_cost || "0", hierarchyData?.summary.currency)}) is the sum of all product holistic costs, including the <strong>Unallocated Platform &amp; Shared Costs</strong> bucket.
          <br />
          <strong>Raw Infrastructure Cost</strong> is the total pre-allocation spend for the same date range (all nodes in the graph).
          <br />
          Allocation coverage shows what percentage of raw infrastructure spend has been allocated into products.
        </AlertDescription>
      </Alert>

      {/* Summary Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex items-center justify-between pb-3">
            <CardTitle className="text-sm font-medium">Allocated Product Cost</CardTitle>
            <Dialog>
              <DialogTrigger asChild>
                <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground">
                  <HelpCircle className="h-4 w-4" />
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>How are these totals calculated?</DialogTitle>
                  <DialogDescription asChild>
                    <div className="space-y-3 text-sm">
                      <p>
                        <strong>Allocated Product Cost</strong> is the sum of holistic costs for all products in the tree, plus a synthetic
                        <em>Unallocated Platform &amp; Shared Costs</em> node which represents platform/shared spend that has not yet been
                        allocated to any specific product.
                      </p>
                      <p>
                        <strong>Raw Infrastructure Cost</strong> comes directly from ingested billing data (the <code>node_costs_by_dimension</code> table)
                        for the same date range. This is the total underlying spend before any allocation.
                      </p>
                      <p>
                        The <strong>allocation coverage %</strong> tells you how much of that raw spend is currently represented in
                        product totals. If coverage is less than 100%, it usually means either:
                      </p>
                      <ul className="list-disc pl-6 space-y-1">
                        <li>you have only run allocation for part of the selected period, or</li>
                        <li>some infrastructure nodes are not yet connected into the allocation graph.</li>
                      </ul>
                      <p className="text-muted-foreground">
                        On the Platform &amp; Shared Services page you see raw platform/shared spend. On this Products page you see
                        how much of that (plus other infra) has actually been attributed to products.
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
              Includes Unallocated Platform &amp; Shared bucket
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Raw Infrastructure Cost</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(hierarchyData?.summary.raw_total_cost || "0", hierarchyData?.summary.currency)}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">Pre-allocation infra spend for this date range</p>
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
            <p className="mt-1 text-xs text-muted-foreground">Allocated vs raw infrastructure cost</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Products</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{hierarchyData?.summary.product_count || 0}</div>
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

