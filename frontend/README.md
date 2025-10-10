# FinOps Aggregator Frontend

A modern, enterprise-grade FinOps dashboard built with Next.js, TypeScript, and shadcn/ui. This application provides comprehensive cost visibility and allocation tracking for cloud infrastructure and services.

## Features

### 🎯 Core Capabilities

- **Interactive Dashboard** - Real-time overview of cloud costs with visual analytics
- **Product Hierarchy View** - Tree-based navigation with expandable/collapsible nodes
- **Cost Attribution** - Direct, allocated, and holistic cost breakdowns
- **Platform Services** - Dedicated view for shared infrastructure costs
- **Cost Allocations** - Weighted distribution and allocation tracking
- **Responsive Design** - Optimized for desktop, tablet, and mobile devices

### 📊 Visualizations

- Bar charts for top products and services
- Pie charts for cost distribution by type
- Interactive tree view with inline cost summaries
- Detailed cost dimension breakdowns
- Allocation flow visualizations

### 🎨 User Experience

- Clean, professional interface designed for FinOps engineers and Finance teams
- Intuitive navigation with persistent header
- Date range selection with quick presets (7, 14, 30, 90 days)
- Real-time data loading with loading states
- Hover effects and smooth transitions
- Color-coded badges for node types

## Tech Stack

- **Framework**: Next.js 15 (Pages Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4
- **UI Components**: shadcn/ui (Radix UI primitives)
- **Charts**: Recharts
- **Icons**: Lucide React
- **Date Handling**: date-fns

## Getting Started

### Prerequisites

- Node.js 18+ (or use mise: `mise use node@22`)
- Backend API running on port 8080

### Installation

1. Install dependencies:

```bash
npm install
```

2. Configure environment variables:

```bash
cp .env.local.example .env.local
```

Edit `.env.local` if your API is running on a different port:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

3. Start the development server:

```bash
npm run dev
```

4. Open [http://localhost:3000](http://localhost:3000) in your browser

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/              # shadcn/ui base components
│   │   ├── layout/          # Layout components
│   │   ├── cost-card.tsx
│   │   ├── product-tree.tsx
│   │   ├── node-detail-panel.tsx
│   │   └── date-range-picker.tsx
│   ├── lib/
│   │   ├── api.ts           # API client
│   │   └── utils.ts         # Utilities
│   ├── pages/
│   │   ├── index.tsx        # Dashboard
│   │   ├── products.tsx     # Product hierarchy
│   │   └── platform.tsx     # Platform/shared
│   ├── styles/
│   │   └── globals.css
│   └── types/
│       └── api.ts
└── package.json
```

## Pages

### Dashboard (`/`)
- Overview metrics and charts
- Top 5 products by cost
- Cost distribution visualizations

### Products (`/products`)
- Interactive tree view (left panel)
- Detailed node information (right panel)
- Cost breakdowns and dependencies

### Platform & Shared (`/platform`)
- Platform services with charts
- Shared services section
- Weighted allocations

## API Integration

Connects to backend at `http://localhost:8080`:

- `GET /api/v1/products/hierarchy` - Product hierarchy
- `GET /api/v1/nodes/{nodeId}` - Node details
- `GET /api/v1/platform/services` - Platform services

## Development

```bash
npm run dev      # Development server
npm run build    # Production build
npm run start    # Production server
npm run lint     # Lint code
```

## Customization

### Theme Colors

Edit `src/styles/globals.css`:

```css
:root {
  --primary: 221.2 83.2% 53.3%;
  /* ... more colors */
}
```

### Date Presets

Modify `src/components/date-range-picker.tsx`

## Troubleshooting

### API Connection
1. Verify backend: `curl http://localhost:8080/health`
2. Check `.env.local`
3. Check browser console for errors

### Build Issues
```bash
rm -rf .next
npm run dev
```

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers

## License

MIT

