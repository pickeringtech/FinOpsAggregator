# Temporary Frontend Workaround

## ⚠️ WARNING: This is NOT the correct solution!

The dashboard currently implements client-side data aggregation as a **temporary workaround**. This is **NOT** how it should work in production.

## What's Wrong

The frontend (`src/pages/index.tsx`) currently:

1. ✗ Flattens the entire product hierarchy tree
2. ✗ Deduplicates nodes by ID
3. ✗ Aggregates costs by type
4. ✗ Sorts and filters to get top products

**All of this should be done in the database with SQL queries!**

## Why This is Bad

### Performance Issues
- Flattening large hierarchies (1000+ nodes) is slow in JavaScript
- Client downloads entire hierarchy just to show top 5 products
- Unnecessary network bandwidth usage
- Browser memory consumption for large datasets

### Scalability Issues
- Won't scale to enterprise-level data (10,000+ nodes)
- Every user's browser does the same expensive computation
- No caching possible at this level

### Correctness Issues
- Deduplication shouldn't be needed if backend returns correct data
- Risk of inconsistent results if logic differs from backend
- Harder to maintain business logic in multiple places

### Architecture Issues
- Violates separation of concerns
- Frontend should display data, not process it
- Backend has the database - use it!

## The Correct Solution

### Backend Should Provide

A new endpoint: `GET /api/v1/dashboard/summary`

This endpoint should return:
```json
{
  "summary": { /* overall stats */ },
  "top_products": [
    /* Top 5 products pre-sorted by cost */
  ],
  "cost_by_type": [
    /* Costs aggregated by type (product, resource, etc.) */
  ]
}
```

### Database Queries

All aggregation should happen in SQL:

```sql
-- Top products
SELECT n.id, n.name, n.type, SUM(hc.cost) as total
FROM nodes n
JOIN holistic_costs hc ON n.id = hc.node_id
WHERE hc.date BETWEEN $1 AND $2
GROUP BY n.id, n.name, n.type
ORDER BY total DESC
LIMIT 5;

-- Cost by type
SELECT n.type, SUM(hc.cost) as total, COUNT(DISTINCT n.id) as count
FROM nodes n
JOIN holistic_costs hc ON n.id = hc.node_id
WHERE hc.date BETWEEN $1 AND $2
GROUP BY n.type;
```

### Frontend Should Do

```typescript
// Simple, clean, correct
const response = await api.dashboard.getSummary(params)
setTopProducts(response.top_products)  // Already sorted!
setCostByType(response.cost_by_type)   // Already aggregated!
```

No flattening, no deduplication, no aggregation - just display the data!

## Current Status

### What Works (Temporarily)
- ✓ Dashboard displays charts with data
- ✓ Top 5 products shown (after deduplication)
- ✓ Cost distribution pie chart works
- ✓ No crashes or errors

### What's Wrong
- ✗ Client-side aggregation (should be server-side)
- ✗ Inefficient data processing
- ✗ Not scalable
- ✗ Violates architecture principles

## Action Items

### For Backend Team
See: `BACKEND_API_REQUIREMENTS.md`

1. Create `/api/v1/dashboard/summary` endpoint
2. Implement SQL queries for aggregation
3. Update OpenAPI specification
4. Test with various date ranges
5. Verify performance with large datasets

**Estimated effort**: 2-4 hours

### For Frontend Team (After Backend is Ready)

1. Remove flattening logic from `src/pages/index.tsx`
2. Remove deduplication logic
3. Remove aggregation logic
4. Add new API client method for dashboard summary
5. Update types to match new endpoint
6. Test with real backend data

**Estimated effort**: 1 hour

## Timeline

- **Current**: Temporary workaround in place (functional but not correct)
- **Target**: Backend endpoint implemented within 1 sprint
- **Cleanup**: Frontend updated immediately after backend is ready

## Testing

### Current Workaround
```bash
cd frontend
npm run dev
# Visit http://localhost:3000
# Charts should display (but using client-side aggregation)
```

### After Backend Fix
```bash
# Backend should be running with new endpoint
curl http://localhost:8080/api/v1/dashboard/summary?start_date=2025-04-10&end_date=2025-05-10

# Frontend should call this endpoint instead
```

## Code Location

The temporary workaround code is in:
- `frontend/src/pages/index.tsx` (lines 50-95)
- Look for the comment: `⚠️ TEMPORARY WORKAROUND`

## References

- `BACKEND_API_REQUIREMENTS.md` - Detailed backend requirements
- `frontend/src/pages/index.tsx` - Current workaround implementation
- `api-specification.yaml` - Current API spec (needs update)

## Questions?

If you're reading this and wondering "why is this here?":
- This is a **known issue**
- It's **documented** and **tracked**
- It's **temporary** until backend endpoint is ready
- It **works** but it's **not the right way**

Don't copy this pattern for other features!

