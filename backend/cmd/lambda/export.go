package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/charts"
	"github.com/pickeringtech/FinOpsAggregator/internal/storage"
	"github.com/rs/zerolog/log"
)

// ExportRequest represents a request to export data.
type ExportRequest struct {
	// Type of export: "chart_graph", "chart_trend", "chart_waterfall", "csv", "json"
	Type string `json:"type"`

	// Common parameters
	Format string `json:"format,omitempty"` // png, svg, csv, json
	Date   string `json:"date,omitempty"`   // YYYY-MM-DD

	// For trend charts
	NodeID    string `json:"node_id,omitempty"`
	NodeName  string `json:"node_name,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Dimension string `json:"dimension,omitempty"`

	// For waterfall charts
	RunID string `json:"run_id,omitempty"`

	// Output configuration
	OutputKey string `json:"output_key,omitempty"` // Custom output key/filename
}

// ExportResponse represents the response from an export operation.
type ExportResponse struct {
	Success   bool   `json:"success"`
	OutputKey string `json:"output_key,omitempty"`
	OutputURL string `json:"output_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

// handleExport handles export requests.
func handleExport(ctx context.Context, request ExportRequest) (LambdaResponse, error) {
	log.Info().
		Str("type", request.Type).
		Str("format", request.Format).
		Msg("Processing export request")

	// Initialize blob storage for output
	storageURL := getEnvOrDefault("FINOPS_STORAGE_URL", "file://./exports")
	storagePrefix := getEnvOrDefault("FINOPS_STORAGE_PREFIX", "exports")

	blobStorage, err := storage.NewBlobStorage(ctx, storageURL, storagePrefix)
	if err != nil {
		return newErrorResponse(500, fmt.Errorf("failed to initialize storage: %w", err)), nil
	}
	defer blobStorage.Close()

	var response ExportResponse

	switch request.Type {
	case "chart_graph":
		response, err = exportGraphChart(ctx, request, blobStorage)
	case "chart_trend":
		response, err = exportTrendChart(ctx, request, blobStorage)
	case "chart_waterfall":
		response, err = exportWaterfallChart(ctx, request, blobStorage)
	default:
		return newErrorResponse(400, fmt.Errorf("unknown export type: %s", request.Type)), nil
	}

	if err != nil {
		log.Error().Err(err).Str("type", request.Type).Msg("Export failed")
		response = ExportResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	body, _ := json.Marshal(response)
	if response.Success {
		return newSuccessResponse(string(body)), nil
	}
	return newErrorResponse(500, fmt.Errorf(response.Error)), nil
}

// exportGraphChart exports a graph structure chart.
func exportGraphChart(ctx context.Context, request ExportRequest, blobStorage *storage.BlobStorage) (ExportResponse, error) {
	format := getFormatOrDefault(request.Format, "png")
	date := parseDateOrDefault(request.Date, time.Now())

	// Generate output key
	outputKey := request.OutputKey
	if outputKey == "" {
		outputKey = fmt.Sprintf("charts/graph-structure-%s.%s", date.Format("2006-01-02"), format)
	}

	// Create chart renderer
	renderer := charts.NewGraphRenderer(st)

	// Render to buffer
	var buf bytes.Buffer
	if err := renderer.RenderGraphStructure(ctx, date, &buf, format); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to render graph structure: %w", err)
	}

	// Write to storage
	contentType := storage.ContentTypeForExtension("." + format)
	if err := blobStorage.Write(ctx, outputKey, buf.Bytes(), contentType); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to write to storage: %w", err)
	}

	return ExportResponse{
		Success:   true,
		OutputKey: outputKey,
		OutputURL: blobStorage.GetURL(outputKey),
	}, nil
}

// exportTrendChart exports a cost trend chart.
func exportTrendChart(ctx context.Context, request ExportRequest, blobStorage *storage.BlobStorage) (ExportResponse, error) {
	format := getFormatOrDefault(request.Format, "png")

	// Parse node ID
	nodeID, err := resolveNodeID(ctx, request.NodeID, request.NodeName)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("failed to resolve node: %w", err)
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("invalid end_date: %w", err)
	}

	dimension := request.Dimension
	if dimension == "" {
		dimension = "instance_hours"
	}

	// Generate output key
	outputKey := request.OutputKey
	if outputKey == "" {
		outputKey = fmt.Sprintf("charts/cost-trend-%s-%s-%s-to-%s.%s",
			nodeID.String()[:8], dimension,
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), format)
	}

	// Create chart renderer
	renderer := charts.NewGraphRenderer(st)

	// Render to buffer
	var buf bytes.Buffer
	if err := renderer.RenderCostTrend(ctx, nodeID, startDate, endDate, dimension, &buf, format); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to render cost trend: %w", err)
	}

	// Write to storage
	contentType := storage.ContentTypeForExtension("." + format)
	if err := blobStorage.Write(ctx, outputKey, buf.Bytes(), contentType); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to write to storage: %w", err)
	}

	return ExportResponse{
		Success:   true,
		OutputKey: outputKey,
		OutputURL: blobStorage.GetURL(outputKey),
	}, nil
}

// exportWaterfallChart exports an allocation waterfall chart.
func exportWaterfallChart(ctx context.Context, request ExportRequest, blobStorage *storage.BlobStorage) (ExportResponse, error) {
	format := getFormatOrDefault(request.Format, "png")

	// Parse node ID
	nodeID, err := resolveNodeID(ctx, request.NodeID, request.NodeName)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("failed to resolve node: %w", err)
	}

	// Parse date
	date, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("invalid date: %w", err)
	}

	// Parse run ID
	runID, err := uuid.Parse(request.RunID)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("invalid run_id: %w", err)
	}

	// Generate output key
	outputKey := request.OutputKey
	if outputKey == "" {
		outputKey = fmt.Sprintf("charts/allocation-waterfall-%s-%s.%s",
			nodeID.String()[:8], date.Format("2006-01-02"), format)
	}

	// Create chart renderer
	renderer := charts.NewGraphRenderer(st)

	// Render to buffer
	var buf bytes.Buffer
	if err := renderer.RenderAllocationWaterfall(ctx, nodeID, date, runID, &buf, format); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to render waterfall: %w", err)
	}

	// Write to storage
	contentType := storage.ContentTypeForExtension("." + format)
	if err := blobStorage.Write(ctx, outputKey, buf.Bytes(), contentType); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to write to storage: %w", err)
	}

	return ExportResponse{
		Success:   true,
		OutputKey: outputKey,
		OutputURL: blobStorage.GetURL(outputKey),
	}, nil
}

// resolveNodeID resolves a node ID from either a UUID string or a node name.
func resolveNodeID(ctx context.Context, nodeIDStr, nodeName string) (uuid.UUID, error) {
	if nodeIDStr != "" {
		nodeID, err := uuid.Parse(nodeIDStr)
		if err == nil {
			return nodeID, nil
		}
		// If not a valid UUID, treat as name
		nodeName = nodeIDStr
	}

	if nodeName == "" {
		return uuid.Nil, fmt.Errorf("node_id or node_name required")
	}

	node, err := st.Nodes.GetByName(ctx, nodeName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("node not found: %s", nodeName)
	}

	return node.ID, nil
}

// getFormatOrDefault returns the format or a default value.
func getFormatOrDefault(format, defaultFormat string) string {
	if format == "" {
		return defaultFormat
	}
	// Normalize format
	ext := filepath.Ext(format)
	if ext != "" {
		return ext[1:] // Remove leading dot
	}
	return format
}

// parseDateOrDefault parses a date string or returns a default.
func parseDateOrDefault(dateStr string, defaultDate time.Time) time.Time {
	if dateStr == "" {
		return defaultDate
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return defaultDate
	}
	return date
}

