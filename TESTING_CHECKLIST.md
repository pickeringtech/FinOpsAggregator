# Dashboard Fix - Testing Checklist

## Pre-Testing Setup

- [ ] Ensure database has computation runs with status='completed'
- [ ] Ensure there's data in `allocation_results_by_dimension` table
- [ ] Ensure there are nodes with type='product' in `cost_nodes` table
- [ ] Backend server is running on port 8080
- [ ] Frontend server is running on port 3000

## Backend Testing

### 1. Build Test
```bash
cd backend
go build -o /tmp/finops-test ./cmd/finops
```
**Expected**: ✅ Build succeeds with no errors

**Status**: ✅ PASSED

### 2. Start Backend Server
```bash
cd backend
./bin/finops serve --port 8080
```
**Expected**: Server starts and listens on port 8080

### 3. Test Health Endpoint
```bash
curl http://localhost:8080/health
```
**Expected**: Returns healthy status

### 4. Test Dashboard Summary Endpoint
```bash
curl "http://localhost:8080/api/v1/dashboard/summary?start_date=2025-01-01&end_date=2025-01-31&currency=USD"
```

**Expected Response Structure**:
```json
{
  "summary": {
    "total_cost": "...",
    "currency": "USD",
    "period": "...",
    "start_date": "...",
    "end_date": "...",
    "node_count": 0,
    "product_count": 0,
    "platform_node_count": 0
  },
  "top_products": [...],
  "cost_by_type": [...]
}
```

**Verify**:
- [ ] Response has all required fields
- [ ] `top_products` array has at most 5 items
- [ ] Each product has `id`, `name`, `type`, `total_cost`
- [ ] `cost_by_type` array has entries for each node type
- [ ] No duplicate products in `top_products`
- [ ] Costs are properly formatted as strings
- [ ] Currency is "USD"

### 5. Test with Different Date Ranges
```bash
# Last 7 days
curl "http://localhost:8080/api/v1/dashboard/summary?start_date=$(date -d '7 days ago' +%Y-%m-%d)&end_date=$(date +%Y-%m-%d)&currency=USD"

# Last 30 days
curl "http://localhost:8080/api/v1/dashboard/summary?start_date=$(date -d '30 days ago' +%Y-%m-%d)&end_date=$(date +%Y-%m-%d)&currency=USD"

# Last 90 days
curl "http://localhost:8080/api/v1/dashboard/summary?start_date=$(date -d '90 days ago' +%Y-%m-%d)&end_date=$(date +%Y-%m-%d)&currency=USD"
```

**Verify**:
- [ ] Different date ranges return different results
- [ ] No errors for any date range
- [ ] Costs change appropriately with date range

### 6. Test Error Cases
```bash
# Missing required parameters
curl "http://localhost:8080/api/v1/dashboard/summary"

# Invalid date format
curl "http://localhost:8080/api/v1/dashboard/summary?start_date=invalid&end_date=2025-01-31&currency=USD"

# End date before start date
curl "http://localhost:8080/api/v1/dashboard/summary?start_date=2025-01-31&end_date=2025-01-01&currency=USD"
```

**Expected**:
- [ ] Returns 400 Bad Request for missing/invalid parameters
- [ ] Error messages are clear and helpful

## Frontend Testing

### 1. Start Frontend Server
```bash
cd frontend
npm run dev
```
**Expected**: Server starts on port 3000

### 2. Open Dashboard
Navigate to: http://localhost:3000

**Verify**:
- [ ] Page loads without errors
- [ ] No console errors in browser DevTools
- [ ] Loading state appears briefly

### 3. Check Summary Cards
**Verify**:
- [ ] "Total Costs" card shows a value
- [ ] "Product Count" card shows a number
- [ ] "Platform Services" card shows a number
- [ ] "Platform Costs" card shows a value
- [ ] All currency values are properly formatted

### 4. Check Top 5 Products Chart
**Verify**:
- [ ] Bar chart is visible
- [ ] Shows at most 5 products
- [ ] **NO DUPLICATE PRODUCTS** ⭐ (This was the bug!)
- [ ] Each bar has a label (product name)
- [ ] Hovering shows tooltip with formatted cost
- [ ] Bars are sorted by cost (highest to lowest)
- [ ] Chart is responsive

### 5. Check Cost Distribution Chart
**Verify**:
- [ ] Pie chart is visible
- [ ] Shows different node types (Product, Resource, Platform, Shared)
- [ ] Each slice has a label with percentage
- [ ] Hovering shows tooltip with formatted cost
- [ ] Colors are distinct for each type
- [ ] Percentages add up to 100%
- [ ] Chart is responsive

### 6. Check All Products List
**Verify**:
- [ ] List shows products
- [ ] Each product has name and total cost
- [ ] Costs are properly formatted
- [ ] **NO DUPLICATE PRODUCTS** ⭐
- [ ] Hover effect works

### 7. Test Date Range Picker
**Verify**:
- [ ] Can select different date ranges
- [ ] Quick presets work (Last 7 days, Last 30 days, etc.)
- [ ] Charts update when date range changes
- [ ] Loading state appears during update
- [ ] Data changes appropriately

### 8. Test Responsive Design
**Verify**:
- [ ] Desktop view (1920x1080): All elements visible
- [ ] Tablet view (768x1024): Layout adjusts properly
- [ ] Mobile view (375x667): Charts stack vertically

### 9. Browser Console Check
Open DevTools Console and verify:
- [ ] No JavaScript errors
- [ ] No failed network requests
- [ ] API calls to `/api/v1/dashboard/summary` succeed
- [ ] Response data structure matches expected format

### 10. Network Tab Check
Open DevTools Network tab and verify:
- [ ] Only 2 API calls on page load:
  - `/api/v1/dashboard/summary`
  - `/api/v1/platform/services`
- [ ] No calls to `/api/v1/products/hierarchy` (old endpoint)
- [ ] Response times are reasonable (< 1 second)
- [ ] Response sizes are small (< 100KB)

## Performance Testing

### Backend Performance
```bash
# Install Apache Bench if not available
# sudo apt-get install apache2-utils

# Test with 100 requests, 10 concurrent
ab -n 100 -c 10 "http://localhost:8080/api/v1/dashboard/summary?start_date=2025-01-01&end_date=2025-01-31&currency=USD"
```

**Verify**:
- [ ] Average response time < 100ms
- [ ] No failed requests
- [ ] Consistent response times

### Frontend Performance
**Verify**:
- [ ] Page loads in < 2 seconds
- [ ] Charts render smoothly
- [ ] No lag when interacting with date picker
- [ ] No memory leaks (check DevTools Memory tab)

## Regression Testing

### Other Pages Still Work
- [ ] Products page (`/products`) still works
- [ ] Platform page (`/platform`) still works
- [ ] Navigation between pages works

### Old Functionality Preserved
- [ ] Date range picker still works everywhere
- [ ] Cost formatting is consistent
- [ ] Error handling works

## Data Validation

### Compare Old vs New
If possible, compare the old dashboard (with duplicates) to the new one:

**Old Dashboard Issues**:
- ❌ Products appeared twice
- ❌ Costs were doubled
- ❌ Slow with large datasets

**New Dashboard**:
- ✅ Each product appears once
- ✅ Costs are correct
- ✅ Fast even with large datasets

### Manual Verification
Pick a specific product and verify:
- [ ] It appears only once in the top products chart
- [ ] Its cost matches the database value
- [ ] It's included in the correct type aggregation

## Sign-Off

- [ ] All backend tests pass
- [ ] All frontend tests pass
- [ ] No duplicate products in charts ⭐
- [ ] Performance is acceptable
- [ ] No regressions in other features
- [ ] Code is properly documented
- [ ] Ready for production deployment

## Notes

Record any issues found during testing:

```
Issue 1: [Description]
Resolution: [How it was fixed]

Issue 2: [Description]
Resolution: [How it was fixed]
```

## Deployment Checklist

Before deploying to production:
- [ ] All tests pass
- [ ] Database indexes are in place
- [ ] Monitoring is configured
- [ ] Rollback plan is ready
- [ ] Documentation is updated
- [ ] Team is notified

