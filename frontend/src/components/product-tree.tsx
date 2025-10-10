import { useState } from "react"
import { ChevronRight, ChevronDown, Folder, FolderOpen, Box, Server, Share2 } from "lucide-react"
import { ProductNode, NodeType } from "@/types/api"
import { formatCurrency } from "@/lib/utils"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"

interface ProductTreeProps {
  nodes: ProductNode[]
  selectedNodeId?: string
  onNodeSelect: (node: ProductNode) => void
}

interface TreeNodeProps {
  node: ProductNode
  level: number
  selectedNodeId?: string
  onNodeSelect: (node: ProductNode) => void
}

function getNodeIcon(type: NodeType, isOpen: boolean) {
  switch (type) {
    case "product":
      return isOpen ? <FolderOpen className="h-4 w-4" /> : <Folder className="h-4 w-4" />
    case "resource":
      return <Box className="h-4 w-4" />
    case "platform":
      return <Server className="h-4 w-4" />
    case "shared":
      return <Share2 className="h-4 w-4" />
    default:
      return <Folder className="h-4 w-4" />
  }
}

function getNodeBadgeVariant(type: NodeType): "default" | "secondary" | "outline" {
  switch (type) {
    case "product":
      return "default"
    case "platform":
      return "secondary"
    default:
      return "outline"
  }
}

function TreeNode({ node, level, selectedNodeId, onNodeSelect }: TreeNodeProps) {
  const [isExpanded, setIsExpanded] = useState(true)
  const hasChildren = node.children && node.children.length > 0
  const isSelected = selectedNodeId === node.id

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
          {getNodeIcon(node.type, isExpanded)}
        </div>
        
        <span className="flex-1 text-sm font-medium truncate">{node.name}</span>
        
        <Badge variant={getNodeBadgeVariant(node.type)} className="text-xs">
          {node.type}
        </Badge>
        
        <div className="text-sm font-semibold text-right min-w-[100px]">
          {formatCurrency(node.holistic_costs.total, node.holistic_costs.currency)}
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

export function ProductTree({ nodes, selectedNodeId, onNodeSelect }: ProductTreeProps) {
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

