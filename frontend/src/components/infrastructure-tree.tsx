import { useState } from "react"
import { ChevronRight, ChevronDown, Server, Share2, Box, Database } from "lucide-react"
import { InfrastructureNode } from "@/types/api"
import { formatCurrency } from "@/lib/utils"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"

interface InfrastructureTreeProps {
  nodes: InfrastructureNode[]
  selectedNodeId?: string
  onNodeSelect: (node: InfrastructureNode) => void
}

interface TreeNodeProps {
  node: InfrastructureNode
  level: number
  selectedNodeId?: string
  onNodeSelect: (node: InfrastructureNode) => void
}

function getNodeIcon(type: string) {
  switch (type) {
    case "platform":
      return <Server className="h-4 w-4" />
    case "shared":
      return <Share2 className="h-4 w-4" />
    case "resource":
      return <Box className="h-4 w-4" />
    case "infrastructure":
      return <Database className="h-4 w-4" />
    default:
      return <Box className="h-4 w-4" />
  }
}

function getNodeBadgeVariant(type: string): "default" | "secondary" | "outline" | "destructive" {
  switch (type) {
    case "platform":
      return "default"
    case "shared":
      return "secondary"
    case "resource":
      return "outline"
    default:
      return "outline"
  }
}

function TreeNode({ node, level, selectedNodeId, onNodeSelect }: TreeNodeProps) {
  const [isExpanded, setIsExpanded] = useState(true)
  const hasChildren = node.children && node.children.length > 0
  const isSelected = selectedNodeId === node.id

  // Calculate allocation status color
  const allocationColor = node.allocation_pct >= 90 
    ? "text-green-600" 
    : node.allocation_pct >= 50 
      ? "text-yellow-600" 
      : "text-red-600"

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-2 py-2 px-3 rounded-md cursor-pointer hover:bg-accent transition-colors",
          isSelected && "bg-accent"
        )}
        style={{ paddingLeft: `${level * 1.5 + 0.75}rem` }}
        onClick={() => onNodeSelect(node)}
      >
        {hasChildren ? (
          <button
            onClick={(e) => {
              e.stopPropagation()
              setIsExpanded(!isExpanded)
            }}
            className="p-0 hover:bg-transparent"
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
          </button>
        ) : (
          <div className="w-4" />
        )}
        
        <div className="text-muted-foreground">
          {getNodeIcon(node.type)}
        </div>
        
        <span className="flex-1 text-sm font-medium truncate">{node.name}</span>
        
        <Badge variant={getNodeBadgeVariant(node.type)} className="text-xs">
          {node.type}
        </Badge>
        
        {/* Allocation percentage indicator */}
        <div className="flex items-center gap-2 min-w-[80px]">
          <Progress value={node.allocation_pct} className="h-2 w-12" />
          <span className={cn("text-xs font-medium", allocationColor)}>
            {node.allocation_pct.toFixed(0)}%
          </span>
        </div>
        
        <div className="text-sm font-semibold text-right min-w-[100px]">
          {formatCurrency(node.direct_costs.total, node.direct_costs.currency)}
        </div>
      </div>
      
      {hasChildren && isExpanded && (
        <div>
          {node.children!.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              level={level + 1}
              selectedNodeId={selectedNodeId}
              onNodeSelect={onNodeSelect}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function InfrastructureTree({ nodes, selectedNodeId, onNodeSelect }: InfrastructureTreeProps) {
  return (
    <div className="space-y-1">
      {nodes.map((node) => (
        <TreeNode
          key={node.id}
          node={node}
          level={0}
          selectedNodeId={selectedNodeId}
          onNodeSelect={onNodeSelect}
        />
      ))}
    </div>
  )
}

