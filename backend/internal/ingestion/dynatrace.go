package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// DynatraceIngester handles ingestion of Dynatrace metric exports
type DynatraceIngester struct {
	store          *store.Store
	metricMappings map[string]MetricMapping
}

// MetricMapping defines how a Dynatrace metric maps to internal metrics
type MetricMapping struct {
	InternalName string
	Unit         string
}

// DynatraceExport represents the structure of a Dynatrace metric export file
type DynatraceExport struct {
	Timeframe DynatraceTimeframe `json:"timeframe"`
	Metrics   []DynatraceMetric  `json:"metrics"`
}

// DynatraceTimeframe represents the time range of the export
type DynatraceTimeframe struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// DynatraceMetric represents a single metric in the export
type DynatraceMetric struct {
	MetricID string                `json:"metricId"`
	Data     []DynatraceDataPoint  `json:"data"`
}

// DynatraceDataPoint represents a single data point with dimensions
type DynatraceDataPoint struct {
	Dimensions map[string]string `json:"dimensions"`
	Timestamps []int64           `json:"timestamps"`
	Values     []float64         `json:"values"`
}

// NewDynatraceIngester creates a new Dynatrace ingester with default metric mappings
func NewDynatraceIngester(store *store.Store) *DynatraceIngester {
	return &DynatraceIngester{
		store: store,
		metricMappings: map[string]MetricMapping{
			"builtin:service.requestCount.total":   {InternalName: "http_requests", Unit: "count"},
			"builtin:service.response.time":        {InternalName: "http_duration_ms", Unit: "milliseconds"},
			"builtin:service.cpu.time":             {InternalName: "cpu_time_ms", Unit: "milliseconds"},
			"builtin:service.errors.total":         {InternalName: "error_count", Unit: "count"},
			"builtin:service.dbconnections.success": {InternalName: "db_connections", Unit: "count"},
			"builtin:service.keyRequest.count":     {InternalName: "key_requests", Unit: "count"},
		},
	}
}

// AddMetricMapping adds a custom metric mapping
func (d *DynatraceIngester) AddMetricMapping(dynatraceMetric string, internalName string, unit string) {
	d.metricMappings[dynatraceMetric] = MetricMapping{
		InternalName: internalName,
		Unit:         unit,
	}
}

// IngestFile ingests a Dynatrace export file
func (d *DynatraceIngester) IngestFile(ctx context.Context, filePath string, nodeID uuid.UUID) (*IngestionResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return d.IngestReader(ctx, file, nodeID)
}

// IngestReader ingests Dynatrace metrics from a reader
func (d *DynatraceIngester) IngestReader(ctx context.Context, reader io.Reader, nodeID uuid.UUID) (*IngestionResult, error) {
	var export DynatraceExport
	if err := json.NewDecoder(reader).Decode(&export); err != nil {
		return nil, fmt.Errorf("failed to decode Dynatrace export: %w", err)
	}

	return d.processExport(ctx, &export, nodeID)
}

// IngestDirectory ingests all Dynatrace export files from a directory
func (d *DynatraceIngester) IngestDirectory(ctx context.Context, dirPath string, nodeMapping map[string]uuid.UUID) (*IngestionResult, error) {
	result := &IngestionResult{
		Source:    "dynatrace",
		StartTime: time.Now(),
	}

	files, err := filepath.Glob(filepath.Join(dirPath, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	for _, file := range files {
		// Try to determine node from filename or use default
		baseName := filepath.Base(file)
		nodeID, ok := nodeMapping[baseName]
		if !ok {
			// Try without extension
			nameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]
			nodeID, ok = nodeMapping[nameWithoutExt]
			if !ok {
				log.Warn().Str("file", file).Msg("No node mapping found for file, skipping")
				result.Errors = append(result.Errors, fmt.Sprintf("no node mapping for %s", baseName))
				continue
			}
		}

		fileResult, err := d.IngestFile(ctx, file, nodeID)
		if err != nil {
			log.Error().Err(err).Str("file", file).Msg("Failed to ingest file")
			result.Errors = append(result.Errors, fmt.Sprintf("failed to ingest %s: %v", baseName, err))
			continue
		}

		result.RecordsProcessed += fileResult.RecordsProcessed
		result.RecordsInserted += fileResult.RecordsInserted
		result.RecordsUpdated += fileResult.RecordsUpdated
		result.RecordsSkipped += fileResult.RecordsSkipped
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// processExport processes a Dynatrace export and stores the metrics
func (d *DynatraceIngester) processExport(ctx context.Context, export *DynatraceExport, nodeID uuid.UUID) (*IngestionResult, error) {
	result := &IngestionResult{
		Source:    "dynatrace",
		StartTime: time.Now(),
	}

	var usages []models.NodeUsageByDimension

	for _, metric := range export.Metrics {
		mapping, ok := d.metricMappings[metric.MetricID]
		if !ok {
			log.Debug().Str("metric_id", metric.MetricID).Msg("No mapping found for metric, skipping")
			continue
		}

		for _, dataPoint := range metric.Data {
			for i, timestamp := range dataPoint.Timestamps {
				if i >= len(dataPoint.Values) {
					continue
				}

				result.RecordsProcessed++

				// Convert timestamp (milliseconds) to time
				usageDate := time.Unix(timestamp/1000, (timestamp%1000)*1000000).UTC()
				// Truncate to day for daily aggregation
				usageDate = time.Date(usageDate.Year(), usageDate.Month(), usageDate.Day(), 0, 0, 0, 0, time.UTC)

				usage := models.NodeUsageByDimension{
					NodeID:    nodeID,
					UsageDate: usageDate,
					Metric:    mapping.InternalName,
					Value:     decimal.NewFromFloat(dataPoint.Values[i]),
					Unit:      mapping.Unit,
					Labels:    dataPoint.Dimensions,
					Source:    "dynatrace",
				}

				usages = append(usages, usage)
			}
		}
	}

	// Bulk insert the usage records
	if len(usages) > 0 {
		if err := d.store.Usage.BulkUpsertWithLabels(ctx, usages); err != nil {
			return nil, fmt.Errorf("failed to store usage records: %w", err)
		}
		result.RecordsInserted = len(usages)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Info().
		Int("processed", result.RecordsProcessed).
		Int("inserted", result.RecordsInserted).
		Dur("duration", result.Duration).
		Msg("Dynatrace ingestion completed")

	return result, nil
}

// IngestionResult represents the result of an ingestion operation
type IngestionResult struct {
	Source           string        `json:"source"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	Duration         time.Duration `json:"duration"`
	RecordsProcessed int           `json:"records_processed"`
	RecordsInserted  int           `json:"records_inserted"`
	RecordsUpdated   int           `json:"records_updated"`
	RecordsSkipped   int           `json:"records_skipped"`
	Errors           []string      `json:"errors,omitempty"`
}

// GetDefaultMetricMappings returns the default Dynatrace metric mappings
func GetDefaultMetricMappings() map[string]MetricMapping {
	return map[string]MetricMapping{
		"builtin:service.requestCount.total":   {InternalName: "http_requests", Unit: "count"},
		"builtin:service.response.time":        {InternalName: "http_duration_ms", Unit: "milliseconds"},
		"builtin:service.cpu.time":             {InternalName: "cpu_time_ms", Unit: "milliseconds"},
		"builtin:service.errors.total":         {InternalName: "error_count", Unit: "count"},
		"builtin:service.dbconnections.success": {InternalName: "db_connections", Unit: "count"},
		"builtin:service.keyRequest.count":     {InternalName: "key_requests", Unit: "count"},
	}
}

