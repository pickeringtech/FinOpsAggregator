# Component Guide

## Component Hierarchy

```
App (_app.tsx)
└── MainLayout
    ├── Header (Navigation)
    └── Page Content
        ├── Dashboard (/)
        ├── Products (/products)
        └── Platform (/platform)
```

## Component Details

### Layout Components

#### `MainLayout`
**Location**: `src/components/layout/main-layout.tsx`
**Purpose**: Wraps all pages with consistent layout
**Props**: `{ children: ReactNode }`
**Features**:
- Renders Header
- Provides main content area
- Handles responsive layout

#### `Header`
**Location**: `src/components/layout/header.tsx`
**Purpose**: Top navigation bar
**Features**:
- Logo and branding
- Navigation links (Dashboard, Products, Platform)
- Active page highlighting
- Responsive design

### Reusable Components

#### `CostCard`
**Location**: `src/components/cost-card.tsx`
**Purpose**: Display cost metrics
**Props**:
```typescript
{
  title: string
  amount: string
  currency?: string
  subtitle?: string
  trend?: { value: number; direction: "up" | "down" | "neutral" }
  icon?: ReactNode
}
```
**Usage**:
```tsx
<CostCard
  title="Total Costs"
  amount="24567.21"
  currency="USD"
  subtitle="Last 30 days"
  icon={<DollarSign />}
/>
```

#### `ProductTree`
**Location**: `src/components/product-tree.tsx`
**Purpose**: Hierarchical tree view of products
**Props**:
```typescript
{
  nodes: ProductNode[]
  selectedNodeId?: string
  onNodeSelect: (node: ProductNode) => void
}
```
**Features**:
- Recursive rendering
- Expand/collapse functionality
- Node type icons
- Inline cost display
- Selection highlighting

#### `NodeDetailPanel`
**Location**: `src/components/node-detail-panel.tsx`
**Purpose**: Detailed node information
**Props**:
```typescript
{
  nodeData: IndividualNodeResponse
}
```
**Displays**:
- Node header with metadata
- Cost summary (3 types)
- Cost dimensions
- Dependencies
- Allocations
- Cost labels

#### `DateRangePicker`
**Location**: `src/components/date-range-picker.tsx`
**Purpose**: Date range selection
**Props**:
```typescript
{
  value: { from: Date; to: Date }
  onChange: (range: { from: Date; to: Date }) => void
}
```
**Features**:
- Current range display
- Quick presets (7, 14, 30, 90 days)
- Formatted date display

### UI Primitives (shadcn/ui)

#### `Button`
**Location**: `src/components/ui/button.tsx`
**Variants**: default, destructive, outline, secondary, ghost, link
**Sizes**: default, sm, lg, icon
**Usage**:
```tsx
<Button variant="outline" size="sm">Click me</Button>
```

#### `Card`
**Location**: `src/components/ui/card.tsx`
**Sub-components**: CardHeader, CardTitle, CardDescription, CardContent, CardFooter
**Usage**:
```tsx
<Card>
  <CardHeader>
    <CardTitle>Title</CardTitle>
  </CardHeader>
  <CardContent>Content here</CardContent>
</Card>
```

#### `Badge`
**Location**: `src/components/ui/badge.tsx`
**Variants**: default, secondary, destructive, outline
**Usage**:
```tsx
<Badge variant="outline">product</Badge>
```

#### `Separator`
**Location**: `src/components/ui/separator.tsx`
**Orientations**: horizontal, vertical
**Usage**:
```tsx
<Separator orientation="horizontal" />
```

#### `ScrollArea`
**Location**: `src/components/ui/scroll-area.tsx`
**Purpose**: Scrollable container
**Usage**:
```tsx
<ScrollArea className="h-96">
  {/* Content */}
</ScrollArea>
```

## Page Components

### Dashboard (`src/pages/index.tsx`)

**Structure**:
```
Container
├── Header Section
│   ├── Title & Description
│   └── DateRangePicker
├── Summary Cards (Grid)
│   ├── CostCard (Total Costs)
│   ├── CostCard (Product Count)
│   ├── CostCard (Platform Services)
│   └── CostCard (Platform Costs)
├── Charts Row (Grid)
│   ├── Card (Top 5 Products - Bar Chart)
│   └── Card (Cost Distribution - Pie Chart)
└── Products List
    └── Card (All Products)
```

**State**:
- `dateRange`: Selected date range
- `hierarchyData`: Product hierarchy response
- `platformData`: Platform services response
- `loading`: Loading state

### Products (`src/pages/products.tsx`)

**Structure**:
```
Container
├── Header Section
│   ├── Title & Description
│   └── DateRangePicker
├── Summary Stats (Grid)
│   ├── Card (Total Cost)
│   ├── Card (Products)
│   ├── Card (Total Nodes)
│   └── Card (Period)
└── Main Content (Grid)
    ├── Card (Tree View)
    │   └── ScrollArea
    │       └── ProductTree
    └── Detail Panel
        └── ScrollArea
            └── NodeDetailPanel
```

**State**:
- `dateRange`: Selected date range
- `hierarchyData`: Product hierarchy response
- `selectedNode`: Currently selected node
- `nodeDetails`: Detailed node information
- `loading`: Loading state
- `loadingDetails`: Detail loading state

### Platform (`src/pages/platform.tsx`)

**Structure**:
```
Container
├── Header Section
│   ├── Title & Description
│   └── DateRangePicker
├── Summary Cards (Grid)
│   ├── Card (Total Platform Cost)
│   ├── Card (Platform Services)
│   └── Card (Shared Services)
├── Platform Services
│   └── Card
│       ├── Bar Chart
│       ├── Separator
│       └── Service List
├── Shared Services
│   └── Card
│       ├── Bar Chart
│       ├── Separator
│       └── Service List
└── Weighted Allocations
    └── Card
        └── Allocation List
```

**State**:
- `dateRange`: Selected date range
- `platformData`: Platform services response
- `loading`: Loading state

## Utility Functions

### `formatCurrency(amount, currency)`
**Location**: `src/lib/utils.ts`
**Purpose**: Format numbers as currency
**Example**: `formatCurrency("1234.56", "USD")` → "$1,234.56"

### `formatNumber(value)`
**Location**: `src/lib/utils.ts`
**Purpose**: Format numbers with decimals
**Example**: `formatNumber("1234.5678")` → "1,234.57"

### `formatDate(date)`
**Location**: `src/lib/utils.ts`
**Purpose**: Format dates
**Example**: `formatDate(new Date())` → "Jan 10, 2025"

### `cn(...inputs)`
**Location**: `src/lib/utils.ts`
**Purpose**: Merge Tailwind classes
**Example**: `cn("text-sm", isActive && "font-bold")`

## API Client

### `api.products.getHierarchy(params)`
**Returns**: `ProductHierarchyResponse`
**Endpoint**: `GET /api/v1/products/hierarchy`

### `api.nodes.getDetails(nodeId, params)`
**Returns**: `IndividualNodeResponse`
**Endpoint**: `GET /api/v1/nodes/{nodeId}`

### `api.platform.getServices(params)`
**Returns**: `PlatformServicesResponse`
**Endpoint**: `GET /api/v1/platform/services`

### `api.health.check()`
**Returns**: `HealthResponse`
**Endpoint**: `GET /health`

## Styling Patterns

### Container
```tsx
<div className="container py-8 space-y-8">
  {/* Content */}
</div>
```

### Grid Layouts
```tsx
{/* 4 columns on large screens */}
<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
  {/* Items */}
</div>

{/* 2 columns */}
<div className="grid gap-4 md:grid-cols-2">
  {/* Items */}
</div>

{/* Fixed + flexible */}
<div className="grid gap-6 lg:grid-cols-[400px_1fr]">
  {/* Fixed width left, flexible right */}
</div>
```

### Spacing
- `space-y-{n}`: Vertical spacing between children
- `gap-{n}`: Grid/flex gap
- `p-{n}`: Padding
- `m-{n}`: Margin

### Common Patterns
```tsx
{/* Loading state */}
{loading && (
  <div className="flex items-center justify-center h-96">
    <div className="text-lg text-muted-foreground">Loading...</div>
  </div>
)}

{/* Hover effect */}
<div className="hover:bg-accent transition-colors cursor-pointer">
  {/* Content */}
</div>

{/* Responsive text */}
<h1 className="text-2xl md:text-3xl lg:text-4xl font-bold">
  Title
</h1>
```

## Best Practices

1. **Always use TypeScript types** - No implicit any
2. **Use shadcn/ui components** - Consistent styling
3. **Format currency/numbers** - Use utility functions
4. **Handle loading states** - Show user feedback
5. **Handle errors** - Try/catch with console.error
6. **Use semantic HTML** - Proper heading hierarchy
7. **Responsive design** - Mobile-first approach
8. **Accessibility** - Proper ARIA labels
9. **Component composition** - Small, reusable components
10. **Consistent spacing** - Use Tailwind spacing scale

