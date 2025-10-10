# Generic API Design - Proper Separation of Concerns

## Problem with Previous Approach

The previous implementation created a `/api/v1/dashboard/summary` endpoint, which violated separation of concerns by:
- Embedding UI-specific concepts (dashboard) in the backend domain
- Creating tight coupling between frontend UI and backend API
- Making the backend less reusable for other clients

## New Approach: Generic, Reusable Endpoints

The backend now provides **generic, domain-focused endpoints** that can be composed by any client:

### 1. List Products with Costs
```
GET /api/v1/products?start_date=2024-01-01&end_date=2024-01-31&currency=USD&limit=5&offset=0
```

**Response:**
```json
{
  "nodes": [
    {
      "id": "uuid",
      "name": "product_name",
      "type": "product",
      "total_cost": "1234.56",
      "currency": "USD"
    }
  ],
  "total_count": 8,
  "limit": 5,
  "offset": 0
}
```

**Use cases:**
- Dashboard: Get top 5 products
- Product list page: Get all products with pagination
- Reports: Export product costs

### 2. List All Nodes with Costs
```
GET /api/v1/nodes?start_date=2024-01-01&end_date=2024-01-31&currency=USD&type=resource&limit=10&offset=0
```

**Response:** Same as products endpoint

**Use cases:**
- Filter by type (product, resource, platform, shared)
- Get all nodes across types
- Paginate through large datasets

### 3. Aggregate Costs by Type
```
GET /api/v1/costs/by-type?start_date=2024-01-01&end_date=2024-01-31&currency=USD
```

**Response:**
```json
{
  "aggregations": [
    {
      "type": "product",
      "total_cost": "5000.00",
      "node_count": 8
    },
    {
      "type": "resource",
      "total_cost": "3000.00",
      "node_count": 12
    }
  ],
  "total_cost": "8000.00",
  "currency": "USD"
}
```

**Use cases:**
- Dashboard: Pie chart of cost distribution
- Reports: Cost breakdown by type
- Analytics: Type-level analysis

### 4. Aggregate Costs by Custom Dimension
```
GET /api/v1/costs/by-dimension?start_date=2024-01-01&end_date=2024-01-31&currency=USD&key=environment
```

**Response:**
```json
{
  "dimension_key": "environment",
  "aggregations": [
    {
      "value": "production",
      "total_cost": "6000.00",
      "node_count": 15
    },
    {
      "value": "staging",
      "total_cost": "2000.00",
      "node_count": 5
    }
  ],
  "total_cost": "8000.00",
  "currency": "USD"
}
```

**Use cases:**
- Dashboard: Filter by environment/team/project
- Reports: Custom dimension analysis
- Chargeback: Cost allocation by business unit

## Frontend Composition Pattern

### Next.js API Routes (Server-Side Composition)

Create `/pages/api/dashboard/summary.ts`:

```typescript
import type { NextApiRequest, NextApiResponse } from 'next'

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  const { start_date, end_date, currency = 'USD' } = req.query
  
  // Compose multiple backend calls
  const [products, costsByType] = await Promise.all([
    fetch(`${BACKEND_URL}/api/v1/products?start_date=${start_date}&end_date=${end_date}&currency=${currency}&limit=5`),
    fetch(`${BACKEND_URL}/api/v1/costs/by-type?start_date=${start_date}&end_date=${end_date}&currency=${currency}`)
  ])
  
  const [productsData, costsByTypeData] = await Promise.all([
    products.json(),
    costsByType.json()
  ])
  
  // Transform for dashboard needs
  const summary = {
    top_products: productsData.nodes,
    cost_by_type: costsByTypeData.aggregations,
    total_cost: costsByTypeData.total_cost,
    product_count: costsByTypeData.aggregations.find(a => a.type === 'product')?.node_count || 0
  }
  
  res.json(summary)
}
```

### Client-Side with SWR

```typescript
import useSWR from 'swr'

function Dashboard() {
  const { data: summary } = useSWR(
    `/api/dashboard/summary?start_date=${startDate}&end_date=${endDate}`,
    fetcher,
    {
      refreshInterval: 60000, // Refresh every minute
      revalidateOnFocus: true
    }
  )
  
  // Or fetch directly from backend with multiple SWR hooks
  const { data: products } = useSWR(
    `${BACKEND_URL}/api/v1/products?start_date=${startDate}&end_date=${endDate}&limit=5`
  )
  
  const { data: costsByType } = useSWR(
    `${BACKEND_URL}/api/v1/costs/by-type?start_date=${startDate}&end_date=${endDate}`
  )
  
  return (
    <div>
      <TopProductsChart data={products?.nodes} />
      <CostDistributionChart data={costsByType?.aggregations} />
    </div>
  )
}
```

## Benefits of This Approach

### 1. Separation of Concerns
- ✅ Backend provides generic, reusable data endpoints
- ✅ Frontend composes data as needed for specific UIs
- ✅ No UI concepts leak into backend domain

### 2. Flexibility
- ✅ Same endpoints can be used by dashboard, reports, mobile app, CLI
- ✅ Easy to add new UI views without backend changes
- ✅ Clients can compose data differently for their needs

### 3. Caching & Performance
- ✅ SWR provides automatic caching and revalidation
- ✅ Multiple components can share same SWR cache
- ✅ Background revalidation keeps data fresh
- ✅ Optimistic updates and mutations

### 4. Scalability
- ✅ Pagination support (limit/offset)
- ✅ Filtering support (type, dimension)
- ✅ Can add more generic endpoints as needed

### 5. Testability
- ✅ Backend endpoints are simple and focused
- ✅ Frontend composition logic is testable
- ✅ Can mock individual endpoints easily

## Migration Path

### Step 1: Update Frontend API Client
Add new generic endpoint methods to `lib/api.ts`

### Step 2: Create Next.js API Routes (Optional)
Create server-side composition routes in `/pages/api/`

### Step 3: Update Dashboard Component
Use SWR to fetch from either:
- Next.js API routes (server-side composition)
- Backend directly (client-side composition)

### Step 4: Remove Old Dashboard-Specific Code
Clean up any dashboard-specific backend code

## Example: Dashboard Implementation

```typescript
// pages/api/dashboard/summary.ts
export default async function handler(req, res) {
  const params = new URLSearchParams(req.query)
  
  const [products, costsByType, platformServices] = await Promise.all([
    backendFetch(`/api/v1/products?${params}&limit=5`),
    backendFetch(`/api/v1/costs/by-type?${params}`),
    backendFetch(`/api/v1/platform/services?${params}`)
  ])
  
  res.json({
    top_products: products.nodes,
    cost_distribution: costsByType.aggregations,
    summary: {
      total_cost: costsByType.total_cost,
      product_count: costsByType.aggregations.find(a => a.type === 'product')?.node_count,
      platform_count: platformServices.platform_services.length
    }
  })
}

// pages/index.tsx
function Dashboard() {
  const [dateRange, setDateRange] = useState({ from: ..., to: ... })
  
  const { data, error, isLoading } = useSWR(
    `/api/dashboard/summary?start_date=${format(dateRange.from)}&end_date=${format(dateRange.to)}`,
    fetcher,
    { refreshInterval: 60000 }
  )
  
  if (isLoading) return <Loading />
  if (error) return <Error />
  
  return (
    <div>
      <SummaryCards data={data.summary} />
      <TopProductsChart data={data.top_products} />
      <CostDistributionChart data={data.cost_distribution} />
    </div>
  )
}
```

## Key Principles

1. **Backend = Domain Logic**: Provide generic, reusable data operations
2. **Frontend = Presentation Logic**: Compose and transform data for specific UIs
3. **API Routes = Composition Layer**: Optional server-side data composition
4. **SWR = Caching Layer**: Automatic caching, revalidation, and state management

This design keeps the backend clean and domain-focused while giving the frontend full flexibility to compose data as needed for different UIs.

