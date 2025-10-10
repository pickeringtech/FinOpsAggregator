# FinOps Aggregator - Quick Start Guide

## Prerequisites

Make sure you have the backend API running on port 8080. You can verify this by running:

```bash
curl http://localhost:8080/health
```

You should see a response like:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-10T10:30:00Z",
  "version": "1.0.0",
  "database": "connected"
}
```

## Installation & Setup

1. **Install Node.js** (if not already installed):

```bash
# Using mise (recommended)
mise use node@22

# Or download from https://nodejs.org/
```

2. **Install dependencies**:

```bash
cd frontend
npm install
```

3. **Configure environment** (already done):

The `.env.local` file is already configured to point to `http://localhost:8080`.

4. **Start the development server**:

```bash
npm run dev
```

5. **Open your browser**:

Navigate to [http://localhost:3000](http://localhost:3000)

## What You'll See

### Dashboard (/)
- **Total costs** across all products
- **Product count** and platform services summary
- **Top 5 products** by cost (bar chart)
- **Cost distribution** by type (pie chart)
- **All products list** with quick stats

### Products (/products)
- **Left Panel**: Interactive tree view
  - Click to expand/collapse nodes
  - See inline cost summaries
  - Color-coded badges for node types
- **Right Panel**: Detailed node information
  - Direct, allocated, and total costs
  - Cost dimensions breakdown
  - Dependencies and relationships
  - Allocation details

### Platform & Shared (/platform)
- **Platform Services**: Infrastructure costs with charts
- **Shared Services**: Shared resource costs
- **Weighted Allocations**: Cost distribution details

## Date Range Selection

Use the date range picker in the top-right corner:
- Click the calendar button to see the current range
- Use quick presets: 7, 14, 30, or 90 days
- Data automatically refreshes when you change the range

## Features to Explore

1. **Tree Navigation**: Click on any product in the tree view to see its details
2. **Cost Breakdown**: View direct vs allocated vs total costs
3. **Dependencies**: See how costs flow between nodes
4. **Allocations**: Understand weighted cost distribution
5. **Charts**: Interactive visualizations with hover tooltips

## Troubleshooting

### "Failed to load data" error
- Verify backend is running: `curl http://localhost:8080/health`
- Check `.env.local` has correct API URL
- Look for CORS errors in browser console

### Port 3000 already in use
```bash
# Use a different port
PORT=3001 npm run dev
```

### Build errors
```bash
# Clear cache and rebuild
rm -rf .next
npm run dev
```

## Production Build

To create a production build:

```bash
npm run build
npm run start
```

The production server will run on port 3000.

## Next Steps

- Explore different date ranges to see cost trends
- Click through the product hierarchy
- Compare direct vs holistic costs
- Review platform service allocations
- Check out the weighted allocation details

## Support

For issues or questions:
1. Check the main README.md for detailed documentation
2. Review the API specification at `../api-specification.yaml`
3. Check browser console for error messages
4. Verify backend logs for API issues

Enjoy exploring your FinOps data! 🚀

