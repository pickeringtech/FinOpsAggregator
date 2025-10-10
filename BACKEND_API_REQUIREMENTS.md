# Backend API Requirements for Dashboard

## Problem Statement

The frontend dashboard currently needs to display:
1. **Top 5 Products by Cost** - A bar chart showing the 5 products with highest costs
2. **Cost Distribution by Type** - A pie chart showing costs grouped by node type (product, resource, platform, shared)

Currently, the frontend is attempting to:
- Flatten the product hierarchy tree
- Deduplicate nodes by ID
- Aggregate costs by type
- Sort and filter to get top products

**This is wrong!** All aggregation should happen in the database with proper SQL queries, not in the frontend or even in Go application logic.

## Required New Endpoint

### `GET /api/v1/dashboard/summary`

A new endpoint that returns pre-aggregated dashboard data calculated entirely in the database.

#### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| start_date | string (date) | Yes | Start date (ISO 8601: YYYY-MM-DD) |
| end_date | string (date) | Yes | End date (ISO 8601: YYYY-MM-DD) |
| currency | string | No | Currency code (default: USD) |
| limit | integer | No | Number of top products to return (default: 5) |

#### Response Schema

```yaml
DashboardSummaryResponse:
  type: object
  required:
    - summary
    - top_products
    - cost_by_type
  properties:
    summary:
      $ref: '#/components/schemas/CostSummary'
      description: Overall cost summary for the period
    
    top_products:
      type: array
      description: Top N products by holistic cost (sorted descending)
      items:
        type: object
        required:
          - id
          - name
          - type
          - holistic_cost
        properties:
          id:
            type: string
            format: uuid
            description: Unique node identifier
          name:
            type: string
            description: Product/node name
          type:
            type: string
            enum: [product, resource, platform, shared]
            description: Node type
          holistic_cost:
            type: object
            required:
              - total
              - currency
            properties:
              total:
                type: string
                description: Total holistic cost as decimal string
                example: "1234.56"
              currency:
                type: string
                description: Currency code
                example: "USD"
    
    cost_by_type:
      type: array
      description: Cost aggregated by node type
      items:
        type: object
        required:
          - type
          - total_cost
          - node_count
        properties:
          type:
            type: string
            enum: [product, resource, platform, shared]
            description: Node type
          total_cost:
            type: object
            required:
              - total
              - currency
            properties:
              total:
                type: string
                description: Total cost for this type
                example: "5678.90"
              currency:
                type: string
                example: "USD"
          node_count:
            type: integer
            description: Number of nodes of this type
            example: 12
```

#### Example Response

```json
{
  "summary": {
    "total_cost": "24567.21",
    "currency": "USD",
    "period": "30 days",
    "start_date": "2025-04-10",
    "end_date": "2025-05-10",
    "node_count": 45,
    "product_count": 8,
    "platform_node_count": 3
  },
  "top_products": [
    {
      "id": "35522c6e-f85a-4cf1-9efe-6543fe7d4dda",
      "name": "payments_service",
      "type": "product",
      "holistic_cost": {
        "total": "8234.56",
        "currency": "USD"
      }
    },
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "user_service",
      "type": "product",
      "holistic_cost": {
        "total": "6123.45",
        "currency": "USD"
      }
    },
    {
      "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
      "name": "analytics_pipeline",
      "type": "product",
      "holistic_cost": {
        "total": "4567.89",
        "currency": "USD"
      }
    },
    {
      "id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
      "name": "notification_service",
      "type": "product",
      "holistic_cost": {
        "total": "3456.78",
        "currency": "USD"
      }
    },
    {
      "id": "d4e5f6a7-b8c9-0123-def1-234567890123",
      "name": "search_service",
      "type": "product",
      "holistic_cost": {
        "total": "2184.53",
        "currency": "USD"
      }
    }
  ],
  "cost_by_type": [
    {
      "type": "product",
      "total_cost": {
        "total": "18234.56",
        "currency": "USD"
      },
      "node_count": 8
    },
    {
      "type": "resource",
      "total_cost": {
        "total": "4123.45",
        "currency": "USD"
      },
      "node_count": 32
    },
    {
      "type": "platform",
      "total_cost": {
        "total": "1567.89",
        "currency": "USD"
      },
      "node_count": 3
    },
    {
      "type": "shared",
      "total_cost": {
        "total": "641.31",
        "currency": "USD"
      },
      "node_count": 2
    }
  ]
}
```

## Database Implementation Requirements

### SQL Queries Needed

#### 1. Top Products Query

```sql
-- Get top N products by holistic cost
-- This should use the existing holistic_costs table/view
-- and aggregate at the node level

SELECT 
    n.id,
    n.name,
    n.type,
    SUM(hc.cost) as total_holistic_cost,
    hc.currency
FROM nodes n
JOIN holistic_costs hc ON n.id = hc.node_id
WHERE hc.date >= $1 
  AND hc.date <= $2
  AND hc.currency = $3
GROUP BY n.id, n.name, n.type, hc.currency
ORDER BY total_holistic_cost DESC
LIMIT $4;
```

#### 2. Cost by Type Query

```sql
-- Aggregate costs by node type
SELECT 
    n.type,
    SUM(hc.cost) as total_cost,
    COUNT(DISTINCT n.id) as node_count,
    hc.currency
FROM nodes n
JOIN holistic_costs hc ON n.id = hc.node_id
WHERE hc.date >= $1 
  AND hc.date <= $2
  AND hc.currency = $3
GROUP BY n.type, hc.currency
ORDER BY total_cost DESC;
```

### Important Notes

1. **Use Existing Tables**: Leverage the existing `holistic_costs` table/view that already has pre-calculated holistic costs
2. **Date Range Filtering**: Filter at the database level using the date range parameters
3. **Currency Handling**: Filter by currency in the WHERE clause
4. **Aggregation**: Use SUM() and GROUP BY for all aggregations
5. **No Application Logic**: The Go handler should just call these queries and return the results
6. **Performance**: Add indexes if needed on (node_id, date, currency) for the holistic_costs table

## Implementation Checklist

Backend team should:

- [ ] Add the new endpoint to `internal/api/handlers.go` (or appropriate file)
- [ ] Create database queries in `internal/store/` (or appropriate package)
- [ ] Update the OpenAPI specification (`api-specification.yaml`)
- [ ] Add appropriate error handling (400, 500)
- [ ] Add logging for the new endpoint
- [ ] Test with various date ranges and currencies
- [ ] Verify performance with large datasets
- [ ] Document any new database indexes needed

## Frontend Changes Required

Once the backend endpoint is ready, the frontend will:

1. Remove all flattening/deduplication logic
2. Remove aggregation logic
3. Call the new `/api/v1/dashboard/summary` endpoint
4. Directly use the returned `top_products` and `cost_by_type` arrays for charts
5. Update the API client types

## Benefits

✅ **Performance**: Database aggregation is much faster than application-level
✅ **Scalability**: Can handle large datasets efficiently
✅ **Correctness**: Database ensures data consistency and accuracy
✅ **Maintainability**: Business logic in one place (database)
✅ **Caching**: Can add caching layer at API level if needed
✅ **Separation of Concerns**: Frontend just displays data, backend handles logic

## Timeline

This is a **blocking issue** for the dashboard. The frontend currently has a workaround with client-side aggregation, but this should be replaced ASAP with the proper backend endpoint.

Estimated effort: 2-4 hours for backend implementation + testing

