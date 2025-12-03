# CSV Export Functionality

This document describes how to use the CSV export functionality that has been added to the FinOps Aggregator.

## Quick Start

### Export Scripts (Recommended)

**For regular use (fast, small files):**
```bash
./export-csv-summary.sh
```

**For detailed analysis (large files with individual records):**
```bash
./export-csv-data.sh
```

Both scripts export to the `exports/` directory which is git-ignored.

## Overview

The CSV export functionality allows you to export various types of financial and cost data from the FinOps system in CSV format. This is useful for:

- Data analysis in spreadsheet applications
- Integration with external reporting tools
- Backup and archival of cost data
- Custom reporting and visualization

## Available Export Methods

### 1. API Endpoint

**Endpoint:** `GET /api/v1/export/csv`

**Parameters:**
- `type` (required): Type of data to export
  - `products` - Products with aggregated costs (summary view)
  - `nodes` - Nodes with aggregated costs and metadata (summary view)
  - `costs_by_type` - Cost breakdown by node type (summary view)
  - `recommendations` - Cost optimization recommendations
  - `detailed_costs` - **NEW**: Individual cost records by date/dimension (detailed view)
  - `raw_costs` - **NEW**: Original ingested cost data with metadata (detailed view)
- `start_date` (required): Start date in YYYY-MM-DD format
- `end_date` (required): End date in YYYY-MM-DD format
- `currency` (optional): Currency code (default: USD)
- `node_type` (optional): Filter by node type (for nodes, detailed_costs, raw_costs exports)
- `node_id` (optional): Filter by specific node ID (for recommendations export)

**Example requests:**
```bash
# Export products with aggregated costs (summary)
curl "http://localhost:8080/api/v1/export/csv?type=products&start_date=2024-12-01&end_date=2026-12-31" -o products.csv

# Export compute nodes (summary)
curl "http://localhost:8080/api/v1/export/csv?type=nodes&node_type=compute&start_date=2024-12-01&end_date=2026-12-31" -o nodes.csv

# Export cost breakdown by type (summary)
curl "http://localhost:8080/api/v1/export/csv?type=costs_by_type&start_date=2024-12-01&end_date=2026-12-31" -o costs.csv

# Export detailed cost records (individual records by date/dimension)
curl "http://localhost:8080/api/v1/export/csv?type=detailed_costs&start_date=2024-12-01&end_date=2026-12-31" -o detailed_costs.csv

# Export raw ingested cost data (original data with metadata)
curl "http://localhost:8080/api/v1/export/csv?type=raw_costs&start_date=2024-12-01&end_date=2026-12-31" -o raw_costs.csv

# Export detailed costs for specific node type
curl "http://localhost:8080/api/v1/export/csv?type=detailed_costs&node_type=product&start_date=2024-12-01&end_date=2026-12-31" -o product_detailed_costs.csv

# Export recommendations for specific node
curl "http://localhost:8080/api/v1/export/csv?type=recommendations&node_id=123e4567-e89b-12d3-a456-426614174000&start_date=2024-12-01&end_date=2026-12-31" -o recommendations.csv
```

### 2. CLI Command

**Command:** `finops export csv`

**Flags:**
- `--type` (default: products): Export type (products, nodes, costs_by_type, recommendations)
- `--start-date`: Start date (YYYY-MM-DD)
- `--end-date`: End date (YYYY-MM-DD)
- `--currency` (default: USD): Currency
- `--node-type`: Node type filter (for nodes export)
- `--node-id`: Node ID filter (for recommendations export)
- `--out`: Output file (default: stdout)

**Example commands:**
```bash
# Export products to file
finops export csv --type=products --start-date=2024-01-01 --end-date=2024-01-31 --out=products.csv

# Export nodes to stdout
finops export csv --type=nodes --node-type=compute --start-date=2024-01-01 --end-date=2024-01-31

# Export recommendations for specific node
finops export csv --type=recommendations --node-id=123e4567-e89b-12d3-a456-426614174000 --out=recommendations.csv
```

### 3. Lambda Function

The Lambda export function now supports CSV exports alongside chart exports.

**Request format:**
```json
{
  "type": "csv",
  "csv_type": "products",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "currency": "USD",
  "node_type": "compute",
  "output_key": "exports/products_2024-01-01_to_2024-01-31.csv"
}
```

## CSV Output Formats

### Products Export
Columns: Product ID, Product Name, Total Cost, Currency, Start Date, End Date, Node Count, Cost per Node

### Nodes Export  
Columns: Node ID, Node Name, Node Type, Total Cost, Currency, Start Date, End Date, Dimensions (JSON)

### Costs by Type Export
Columns: Cost Type, Total Amount, Currency, Start Date, End Date, Node Count, Average per Node

### Recommendations Export
Columns: Recommendation ID, Node ID, Node Name, Type, Description, Potential Savings, Currency, Priority, Created Date

## Testing

To test the CSV export functionality in the docker-compose environment:

1. Start the development environment:
   ```bash
   docker-compose -f docker-compose.dev.yml up -d
   ```

2. Run the test script:
   ```bash
   chmod +x test-csv-export.sh
   ./test-csv-export.sh
   ```

3. Check the generated CSV files for results.

## Storage

In the docker-compose environment, CSV exports are stored in LocalStack S3 emulation. The storage configuration uses:
- Endpoint: http://localstack:4566
- Bucket: finops-exports
- Prefix: exports/

For production deployments, configure the `FINOPS_STORAGE_URL` environment variable to point to your actual S3 bucket or other blob storage.
