# Refactoring Complete: Generic API Design

## What Was Wrong

The previous implementation created a `/api/v1/dashboard/summary` endpoint, which violated separation of concerns:
- ❌ Backend had UI-specific concepts (dashboard) in the domain layer
- ❌ Tight coupling between frontend UI and backend API
- ❌ Backend was not reusable for other clients (mobile, CLI, reports)
- ❌ No caching strategy on the frontend

## What Was Fixed

### Backend: Generic, Reusable Endpoints

Removed dashboard-specific endpoint and created **4 generic endpoints**:

#### 1. `GET /api/v1/products`
List products with aggregated costs (flat list, no hierarchy)

**Parameters:**
- `start_date`, `end_date`, `currency` (required)
- `limit`, `offset` (optional, for pagination)

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

#### 2. `GET /api/v1/nodes`
List all nodes with aggregated costs (any type)

**Parameters:**
- `start_date`, `end_date`, `currency` (required)
- `type` (optional: product, resource, platform, shared)
- `limit`, `offset` (optional)

**Response:** Same as products endpoint

**Use cases:**
- Filter by type
- Get all nodes across types
- Paginate through large datasets

#### 3. `GET /api/v1/costs/by-type`
Aggregate costs by node type

**Parameters:**
- `start_date`, `end_date`, `currency` (required)

**Response:**
```json
{
  "aggregations": [
    {
      "type": "product",
      "total_cost": "5000.00",
      "node_count": 8
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

#### 4. `GET /api/v1/costs/by-dimension`
Aggregate costs by custom dimension (tags, labels)

**Parameters:**
- `start_date`, `end_date`, `currency` (required)
- `key` (required: dimension key like "environment", "team", "project")

**Response:**
```json
{
  "dimension_key": "environment",
  "aggregations": [
    {
      "value": "production",
      "total_cost": "6000.00",
      "node_count": 15
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

### Frontend: Composition Layer

#### Next.js API Route (Server-Side Composition)

Created `/pages/api/dashboard/summary.ts` that:
- ✅ Composes multiple backend API calls in parallel
- ✅ Transforms data for dashboard-specific needs
- ✅ Keeps UI logic in the frontend
- ✅ Provides a clean interface for the dashboard component

#### SWR for Caching

Created new dashboard implementation (`index-new.tsx`) that:
- ✅ Uses SWR for automatic caching and revalidation
- ✅ Refreshes data every minute
- ✅ Revalidates on focus
- ✅ Provides loading and error states
- ✅ Shares cache across components

## Files Changed

### Backend
1. **backend/internal/api/router.go**
   - Removed `/api/v1/dashboard/summary` route
   - Added `/api/v1/products` (list)
   - Added `/api/v1/nodes` (list)
   - Added `/api/v1/costs/by-type`
   - Added `/api/v1/costs/by-dimension`

2. **backend/internal/store/costs.go**
   - Renamed `DashboardTopProduct` → `NodeWithCost`
   - Renamed `DashboardCostByType` → `CostByType`
   - Added `CostByDimension` type
   - Renamed `GetTopProductsByHolisticCost` → `ListNodesWithCosts`
   - Renamed `GetCostByType` → `GetCostsByType`
   - Added `GetCostsByDimension` method
   - Added pagination support (limit/offset)
   - Added type filtering

3. **backend/internal/api/models.go**
   - Removed dashboard-specific types
   - Added `NodeListResponse`
   - Added `NodeWithCostData`
   - Added `CostsByTypeResponse`
   - Added `TypeAggregation`
   - Added `CostsByDimensionResponse`
   - Added `DimensionAggregation`

4. **backend/internal/api/service.go**
   - Removed `GetDashboardSummary` method
   - Added `ListProducts` method
   - Added `ListNodes` method
   - Added `GetCostsByType` method
   - Added `GetCostsByDimension` method

5. **backend/internal/api/handlers.go**
   - Removed `GetDashboardSummary` handler
   - Added `ListProducts` handler
   - Added `ListNodes` handler
   - Added `GetCostsByType` handler
   - Added `GetCostsByDimension` handler

### Frontend
1. **frontend/src/types/api.ts**
   - Removed dashboard-specific types
   - Added `NodeListResponse`
   - Added `NodeWithCost`
   - Added `CostsByTypeResponse`
   - Added `TypeAggregation`
   - Added `CostsByDimensionResponse`
   - Added `DimensionAggregation`

2. **frontend/src/lib/api.ts**
   - Removed `dashboard.getSummary`
   - Added `products.list`
   - Added `nodes.list`
   - Added `costs.byType`
   - Added `costs.byDimension`

3. **frontend/src/pages/api/dashboard/summary.ts** (NEW)
   - Server-side composition of generic endpoints
   - Transforms data for dashboard needs
   - Parallel API calls for performance

4. **frontend/src/pages/index-new.tsx** (NEW)
   - Uses SWR for data fetching
   - Automatic caching and revalidation
   - Clean, simple component code

## Testing

### Backend Endpoints

```bash
# Test products list
curl "http://localhost:8080/api/v1/products?start_date=2024-01-01&end_date=2024-01-31&currency=USD&limit=5" | jq .

# Test nodes list
curl "http://localhost:8080/api/v1/nodes?start_date=2024-01-01&end_date=2024-01-31&currency=USD&type=resource&limit=10" | jq .

# Test costs by type
curl "http://localhost:8080/api/v1/costs/by-type?start_date=2024-01-01&end_date=2024-01-31&currency=USD" | jq .

# Test costs by dimension
curl "http://localhost:8080/api/v1/costs/by-dimension?start_date=2024-01-01&end_date=2024-01-31&currency=USD&key=environment" | jq .
```

### Frontend

1. Install SWR:
   ```bash
   cd frontend
   npm install swr
   # or
   yarn add swr
   # or
   pnpm add swr
   ```

2. Replace `index.tsx` with `index-new.tsx`:
   ```bash
   mv src/pages/index.tsx src/pages/index-old.tsx
   mv src/pages/index-new.tsx src/pages/index.tsx
   ```

3. Start the frontend:
   ```bash
   npm run dev
   ```

4. Visit http://localhost:3000

## Benefits

### 1. Separation of Concerns ✅
- Backend provides generic, reusable data operations
- Frontend composes data for specific UIs
- No UI concepts in backend domain

### 2. Flexibility ✅
- Same endpoints for dashboard, reports, mobile, CLI
- Easy to add new UI views without backend changes
- Clients compose data differently for their needs

### 3. Caching & Performance ✅
- SWR provides automatic caching
- Background revalidation keeps data fresh
- Optimistic updates and mutations
- Reduced backend load

### 4. Scalability ✅
- Pagination support (limit/offset)
- Filtering support (type, dimension)
- Can add more generic endpoints as needed

### 5. Testability ✅
- Backend endpoints are simple and focused
- Frontend composition logic is testable
- Can mock individual endpoints easily

## Next Steps

1. **Install SWR** in the frontend
2. **Replace dashboard** with the new SWR-based implementation
3. **Test the new endpoints** with the provided curl commands
4. **Add more generic endpoints** as needed (e.g., `/api/v1/costs/by-tag`)
5. **Create more Next.js API routes** for other UI compositions

## Key Principles

1. **Backend = Domain Logic**: Provide generic, reusable data operations
2. **Frontend = Presentation Logic**: Compose and transform data for specific UIs
3. **API Routes = Composition Layer**: Optional server-side data composition
4. **SWR = Caching Layer**: Automatic caching, revalidation, and state management

This design keeps the backend clean and domain-focused while giving the frontend full flexibility to compose data as needed for different UIs.

