# Dynatrace Metrics Integration

This document describes how Dynatrace metrics are ingested and used for cost allocation.

## 1. Supported Metric Types

| Dynatrace Metric | Internal Metric Name | Unit | Description |
|------------------|---------------------|------|-------------|
| `builtin:service.requestCount.total` | `http_requests` | count | Total HTTP request count |
| `builtin:service.response.time` | `http_duration_ms` | milliseconds | Request duration/latency |
| `builtin:service.cpu.time` | `cpu_time_ms` | milliseconds | CPU time consumed |
| `builtin:service.errors.total` | `error_count` | count | Total error count |
| `builtin:service.dbconnections.success` | `db_connections` | count | Database connection count |
| `builtin:service.keyRequest.count` | `key_requests` | count | Key request count |

## 2. Label-Based Filtering

Dynatrace metrics can include labels (dimensions) that enable segment-based allocation.

### 2.1 Supported Labels

- **`customer_id`**: Allocate costs based on per-customer usage
- **`service_name`**: Allocate based on calling service
- **`region`**: Allocate based on geographic region
- **`plan_tier`**: Allocate based on customer plan (free/pro/enterprise)
- **`environment`**: Allocate based on environment (prod/staging/dev)
- **`team`**: Allocate based on owning team

### 2.2 Example: Customer-Based Allocation

```
Product "API Gateway" ($10,000/day) → Customer Products
Dynatrace metrics with customer_id labels:
  - customer_001: 50,000 requests
  - customer_002: 30,000 requests
  - customer_003: 20,000 requests

Result with proportional_on(metric="http_requests", filter="customer_id"):
  - Customer 001 Product: $5,000/day (50%)
  - Customer 002 Product: $3,000/day (30%)
  - Customer 003 Product: $2,000/day (20%)
```

### 2.3 Example: Region-Based Allocation

```
Platform "Global CDN" ($50,000/day) → Regional Products
Dynatrace metrics with region labels:
  - us-east: 2,000,000 requests
  - eu-west: 1,500,000 requests
  - ap-south: 500,000 requests

Result with proportional_on(metric="http_requests", filter="region"):
  - US East Product: $25,000/day (50%)
  - EU West Product: $18,750/day (37.5%)
  - AP South Product: $6,250/day (12.5%)
```

## 3. Ingestion Format

### 3.1 Dynatrace Export JSON

Dynatrace metrics are ingested via export files in the following format:

```json
{
  "timeframe": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-02T00:00:00Z"
  },
  "metrics": [
    {
      "metricId": "builtin:service.requestCount.total",
      "data": [
        {
          "dimensions": {
            "customer_id": "customer_001",
            "service": "api-gateway"
          },
          "timestamps": [1704067200000],
          "values": [50000]
        },
        {
          "dimensions": {
            "customer_id": "customer_002",
            "service": "api-gateway"
          },
          "timestamps": [1704067200000],
          "values": [30000]
        }
      ]
    }
  ]
}
```

### 3.2 Internal Storage Format

Metrics are stored in `NodeUsageByDimension` with extended label support:

```go
type NodeUsageByDimension struct {
    NodeID    uuid.UUID       `json:"node_id"`
    UsageDate time.Time       `json:"usage_date"`
    Metric    string          `json:"metric"`
    Value     decimal.Decimal `json:"value"`
    Unit      string          `json:"unit"`
    Labels    map[string]string `json:"labels"` // Extended for label support
}
```

## 4. Ingestion Process

### 4.1 File-Based Ingestion

1. Place Dynatrace export files in the configured ingestion directory
2. Files are processed in chronological order
3. Metrics are mapped to internal metric names
4. Labels are preserved for segment-based allocation

### 4.2 API-Based Ingestion (Future)

Direct integration with Dynatrace API:

```
POST /api/v1/metrics/ingest
{
  "source": "dynatrace",
  "api_token": "...",
  "metric_ids": ["builtin:service.requestCount.total"],
  "timeframe": {"from": "now-1d", "to": "now"}
}
```

## 5. Metric Mapping Configuration

### 5.1 Default Mappings

```yaml
dynatrace:
  metric_mappings:
    "builtin:service.requestCount.total": "http_requests"
    "builtin:service.response.time": "http_duration_ms"
    "builtin:service.cpu.time": "cpu_time_ms"
    "builtin:service.errors.total": "error_count"
```

### 5.2 Custom Mappings

```yaml
dynatrace:
  custom_mappings:
    "custom:myapp.transactions": "app_transactions"
    "custom:myapp.revenue": "revenue_events"
```

## 6. Usage in Allocation Strategies

### 6.1 Basic Proportional Allocation

```json
{
  "strategy": "proportional_on",
  "parameters": {
    "metric": "http_requests"
  }
}
```

### 6.2 Segment-Filtered Allocation

```json
{
  "strategy": "proportional_on",
  "parameters": {
    "metric": "http_requests",
    "segment_filter": {
      "label": "customer_id",
      "values": ["customer_001", "customer_002"]
    }
  }
}
```

### 6.3 Weighted Average with Dynatrace Metrics

```json
{
  "strategy": "weighted_average",
  "parameters": {
    "metric": "http_duration_ms",
    "window_days": 7,
    "segment_filter": {
      "label": "plan_tier",
      "values": ["enterprise"]
    }
  }
}
```

## 7. Troubleshooting

### 7.1 Missing Metrics

If metrics are not appearing:
1. Check the Dynatrace export file format
2. Verify metric IDs match the mapping configuration
3. Check the ingestion logs for parsing errors

### 7.2 Label Mismatches

If segment-based allocation is not working:
1. Verify label names match exactly (case-sensitive)
2. Check that label values exist in the ingested data
3. Review the allocation debug endpoint for label details

### 7.3 Time Zone Issues

Dynatrace exports use UTC timestamps. Ensure:
1. All timestamps are in UTC
2. Date ranges align with allocation periods
3. No gaps in metric data for the allocation window

