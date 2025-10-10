# Install SWR and Complete the Migration

## Current Status

✅ Backend is running with new generic endpoints
✅ Dashboard code has been updated to use SWR
❌ SWR package is not installed yet

## Step 1: Install SWR

You need to install the SWR package. Run ONE of these commands:

```bash
# If you have npm
cd frontend
npm install swr

# If you have yarn
cd frontend
yarn add swr

# If you have pnpm
cd frontend
pnpm add swr
```

## Step 2: Verify Installation

After installing, check that SWR is in your package.json:

```bash
cd frontend
grep swr package.json
```

You should see something like:
```json
"swr": "^2.2.4"
```

## Step 3: Restart Frontend Dev Server

If your frontend dev server is running, restart it:

```bash
# Stop the current server (Ctrl+C)
# Then start it again
cd frontend
npm run dev
```

## Step 4: Test the Dashboard

Visit http://localhost:3000

You should now see:
- ✅ No more errors
- ✅ Dashboard loads with data
- ✅ Top 5 Products chart (no duplicates!)
- ✅ Cost Distribution pie chart
- ✅ Summary cards with counts

## What Changed

### Old Dashboard (index-old-backup.tsx)
- Used `useEffect` and `useState` for data fetching
- Called `api.dashboard.getSummary()` (which no longer exists)
- Manual loading state management
- No caching

### New Dashboard (index.tsx)
- Uses `useSWR` for data fetching
- Calls Next.js API route `/api/dashboard/summary`
- Automatic loading/error states
- Automatic caching and revalidation
- Refreshes every 60 seconds
- Revalidates when you focus the tab

## How It Works

```
┌─────────────┐
│  Dashboard  │
│  Component  │
└──────┬──────┘
       │ useSWR('/api/dashboard/summary?...')
       ↓
┌─────────────────────────┐
│  Next.js API Route      │
│  /api/dashboard/summary │
└──────┬──────────────────┘
       │ Parallel requests
       ├─→ GET /api/v1/products?limit=5
       └─→ GET /api/v1/costs/by-type
       ↓
┌─────────────┐
│   Backend   │
│  (Go API)   │
└─────────────┘
```

### Benefits

1. **Separation of Concerns**
   - Backend: Generic, reusable endpoints
   - Next.js API: Composition layer
   - Dashboard: Presentation layer

2. **Caching**
   - SWR caches responses automatically
   - Reduces backend load
   - Faster page loads

3. **Revalidation**
   - Background updates every 60 seconds
   - Revalidates when you focus the tab
   - Always shows fresh data

4. **Error Handling**
   - Automatic error states
   - Retry logic built-in
   - Better UX

## Troubleshooting

### Error: "Cannot find module 'swr'"

**Solution**: Install SWR (see Step 1 above)

### Error: "Cannot read properties of undefined (reading 'getSummary')"

**Solution**: Make sure you replaced the old dashboard:
```bash
cd frontend/src/pages
ls -la index*.tsx
# You should see:
# - index.tsx (the new one)
# - index-old-backup.tsx (the old one)
```

### Error: "404 Not Found" on /api/dashboard/summary

**Solution**: Make sure the Next.js API route exists:
```bash
ls -la frontend/src/pages/api/dashboard/summary.ts
```

If it doesn't exist, it should have been created. Check the file.

### Dashboard shows $0 for all costs

**Solution**: This is expected if there's no cost data in the database. The important thing is:
- ✅ No errors
- ✅ No duplicates in the charts
- ✅ Correct structure

To add test data, see the database migration scripts.

## Next Steps

After SWR is installed and working:

1. **Test the generic endpoints directly**:
   ```bash
   # Top 5 products
   curl "http://localhost:8080/api/v1/products?start_date=2024-01-01&end_date=2024-01-31&currency=USD&limit=5" | jq .
   
   # Costs by type
   curl "http://localhost:8080/api/v1/costs/by-type?start_date=2024-01-01&end_date=2024-01-31&currency=USD" | jq .
   ```

2. **Create more compositions**:
   - Add more Next.js API routes for other pages
   - Use the same generic backend endpoints
   - Compose data differently for different UIs

3. **Add more generic endpoints**:
   - `/api/v1/costs/by-tag`
   - `/api/v1/costs/by-team`
   - `/api/v1/costs/trend` (time series)

4. **Use SWR in other components**:
   - Product list page
   - Node details page
   - Reports page

## Summary

The refactoring is complete! You just need to:
1. Install SWR: `npm install swr`
2. Restart frontend dev server
3. Refresh your browser

The dashboard will now use:
- ✅ Generic backend endpoints
- ✅ Next.js API route for composition
- ✅ SWR for caching and revalidation
- ✅ No more duplicates!
- ✅ Proper separation of concerns!

