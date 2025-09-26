# Chart Generation Guide

The FinOps DAG Cost Attribution Tool includes comprehensive chart generation capabilities to visualize your cost data and graph structure.

## Quick Start

### 1. Generate Demo Charts
```bash
make demo-charts
```

This creates sample charts using the demo data including:
- Graph structure visualization
- Cost trend charts for key nodes
- Allocation breakdown charts

### 2. Generate All Charts
```bash
make charts-all
```

### 3. Generate Specific Chart Types
```bash
make charts-graph    # Graph structure only
make charts-trends   # Cost trends only
```

## Chart Types

### 1. Graph Structure Chart
Visualizes the DAG structure showing nodes and their dependencies.

**CLI Command:**
```bash
./bin/finops export chart graph --format png --out graph-structure.png
```

**Features:**
- Hierarchical layout based on dependency levels
- Color-coded nodes by type
- Edge visualization showing cost flow
- Supports PNG and SVG formats

### 2. Cost Trend Charts
Shows cost trends over time for specific nodes and dimensions.

**CLI Command:**
```bash
./bin/finops export chart trend \
  --node product_p \
  --dimension instance_hours \
  --from 2024-01-01 \
  --to 2024-01-31 \
  --format png \
  --out trend-product_p-instance_hours.png
```

**Features:**
- Time series visualization
- Multiple cost dimensions
- Configurable date ranges
- Automatic scaling and formatting

### 3. Allocation Waterfall Charts
Shows how costs are allocated from direct costs through the dependency chain.

**CLI Command:**
```bash
./bin/finops export chart waterfall \
  --node product_p \
  --date 2024-01-15 \
  --run <allocation-run-id> \
  --format png \
  --out waterfall-product_p.png
```

**Features:**
- Direct vs indirect cost breakdown
- Dimension-wise allocation visualization
- Cumulative cost flow representation

## Output Formats

### PNG (Default)
- High-quality raster images
- Good for reports and presentations
- Smaller file sizes
- Universal compatibility

### SVG
- Vector graphics format
- Scalable without quality loss
- Editable in design tools
- Ideal for web display

## Storage Backends

Charts can be exported to various storage backends:

### Local Filesystem (Default)
```yaml
storage:
  url: file://./charts
  prefix: "finops-charts"
```

### AWS S3
```yaml
storage:
  url: s3://my-bucket?region=us-east-1
  prefix: "finops-charts"
```

### Google Cloud Storage
```yaml
storage:
  url: gs://my-bucket
  prefix: "finops-charts"
```

## Configuration

### Chart Settings
```yaml
charts:
  out_dir: ./charts

storage:
  url: file://./charts
  prefix: "finops-charts"
```

### Environment Variables
```bash
export FINOPS_STORAGE_URL="s3://my-bucket?region=us-east-1"
export FINOPS_STORAGE_PREFIX="production-charts"
```

## Advanced Usage

### Batch Chart Generation
Use the provided script for batch operations:

```bash
# Generate all demo charts
./scripts/generate-charts.sh demo

# Generate trends with custom date range
./scripts/generate-charts.sh trends --start 2024-01-01 --end 2024-12-31

# Generate as SVG format
./scripts/generate-charts.sh all --format svg

# Custom output directory
./scripts/generate-charts.sh all --dir /tmp/my-charts
```

### Programmatic Access
```go
// Create chart exporter
exporter, err := charts.NewExporter(store, "file://./charts", "my-prefix")
if err != nil {
    return err
}
defer exporter.Close()

// Export graph structure
err = exporter.ExportGraphStructure(ctx, time.Now(), "graph.png", "png")

// Export cost trend
err = exporter.ExportCostTrend(ctx, nodeID, startDate, endDate, "instance_hours", "trend.png", "png")

// Export allocation waterfall
err = exporter.ExportAllocationWaterfall(ctx, nodeID, date, runID, "waterfall.png", "png")
```

## Chart Customization

### Colors and Styling
Charts use a consistent color scheme:
- **Blue**: Primary data series, nodes
- **Red**: Edges, indirect costs
- **Green**: Totals, completed allocations
- **Yellow**: Warnings, partial data

### Layout Options
- **Hierarchical**: Nodes arranged by dependency levels
- **Circular**: Nodes arranged in a circle (fallback)
- **Time Series**: Chronological arrangement for trends

## Troubleshooting

### Common Issues

#### 1. No Data Found
```
Error: no cost data found for node product_p
```
**Solution**: Ensure data is loaded and the node exists:
```bash
./bin/finops demo seed
./bin/finops graph validate
```

#### 2. Invalid Date Format
```
Error: invalid date format
```
**Solution**: Use YYYY-MM-DD format:
```bash
--from 2024-01-01 --to 2024-01-31
```

#### 3. Storage Permission Issues
```
Error: failed to write to storage
```
**Solution**: Check storage permissions and credentials:
```bash
# For S3
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret

# For GCS
export GOOGLE_APPLICATION_CREDENTIALS=path/to/service-account.json
```

### Debug Mode
Enable debug logging to troubleshoot chart generation:
```bash
export FINOPS_LOGGING_LEVEL=debug
./bin/finops export chart graph --format png
```

## Integration Examples

### CI/CD Pipeline
```yaml
# .github/workflows/charts.yml
- name: Generate Charts
  run: |
    make build
    make demo-seed
    make charts-all
    
- name: Upload Charts
  uses: actions/upload-artifact@v3
  with:
    name: finops-charts
    path: charts/
```

### Scheduled Reports
```bash
#!/bin/bash
# daily-charts.sh
export FINOPS_STORAGE_URL="s3://reports-bucket"
export FINOPS_STORAGE_PREFIX="daily-$(date +%Y-%m-%d)"

./bin/finops export chart graph --format png
./scripts/generate-charts.sh trends --start $(date -d '30 days ago' +%Y-%m-%d) --end $(date +%Y-%m-%d)
```

## Performance Notes

- Graph structure charts: ~2-5 seconds for 100 nodes
- Trend charts: ~1-3 seconds per node/dimension
- Waterfall charts: ~1-2 seconds per allocation
- Memory usage: ~50MB for typical datasets
- PNG files: 100-500KB typical size
- SVG files: 50-200KB typical size

## Next Steps

1. **Custom Chart Types**: Extend the chart system with custom visualizations
2. **Interactive Charts**: Add web-based interactive charts
3. **Real-time Updates**: Implement live chart updates
4. **Dashboard Integration**: Embed charts in web dashboards
5. **Export Automation**: Set up automated chart generation and distribution
