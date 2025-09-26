package charts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/s3blob"
	_ "gocloud.dev/blob/gcsblob"
)

// Exporter handles chart generation and export to various storage backends
type Exporter struct {
	store    *store.Store
	renderer *GraphRenderer
	bucket   *blob.Bucket
	prefix   string
}

// NewExporter creates a new chart exporter
func NewExporter(store *store.Store, storageURL, prefix string) (*Exporter, error) {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, storageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage bucket: %w", err)
	}

	return &Exporter{
		store:    store,
		renderer: NewGraphRenderer(store),
		bucket:   bucket,
		prefix:   prefix,
	}, nil
}

// Close closes the exporter and cleans up resources
func (e *Exporter) Close() error {
	return e.bucket.Close()
}

// ExportGraphStructure exports the DAG structure as an image
func (e *Exporter) ExportGraphStructure(ctx context.Context, date time.Time, filename, format string) error {
	log.Info().
		Time("date", date).
		Str("filename", filename).
		Str("format", format).
		Msg("Exporting graph structure")

	// Ensure format is supported
	if format != "png" && format != "svg" {
		return fmt.Errorf("unsupported format: %s (supported: png, svg)", format)
	}

	// Generate filename if not provided
	if filename == "" {
		filename = fmt.Sprintf("graph-structure-%s.%s", date.Format("2006-01-02"), format)
	}

	// Add prefix if configured
	if e.prefix != "" {
		filename = filepath.Join(e.prefix, filename)
	}

	// Create a temporary file to write to
	tempFile, err := os.CreateTemp("", "finops-chart-*."+format)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Render the graph
	if err := e.renderer.RenderGraphStructure(ctx, date, tempFile, format); err != nil {
		return fmt.Errorf("failed to render graph structure: %w", err)
	}

	// Reopen file for reading
	tempFile.Close()
	file, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to reopen temp file: %w", err)
	}
	defer file.Close()

	// Upload to storage
	writer, err := e.bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return fmt.Errorf("failed to create storage writer: %w", err)
	}
	defer writer.Close()

	// Set content type
	contentType := "image/png"
	if format == "svg" {
		contentType = "image/svg+xml"
	}
	writer.ContentType = contentType

	// Copy file to storage
	if _, err := file.WriteTo(writer); err != nil {
		return fmt.Errorf("failed to write to storage: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close storage writer: %w", err)
	}

	log.Info().
		Str("filename", filename).
		Str("format", format).
		Msg("Graph structure exported successfully")

	return nil
}

// ExportCostTrend exports a cost trend chart for a specific node
func (e *Exporter) ExportCostTrend(ctx context.Context, nodeID uuid.UUID, startDate, endDate time.Time, dimension, filename, format string) error {
	log.Info().
		Str("node_id", nodeID.String()).
		Time("start_date", startDate).
		Time("end_date", endDate).
		Str("dimension", dimension).
		Str("filename", filename).
		Str("format", format).
		Msg("Exporting cost trend chart")

	// Ensure format is supported
	if format != "png" && format != "svg" {
		return fmt.Errorf("unsupported format: %s (supported: png, svg)", format)
	}

	// Get node name for filename if not provided
	if filename == "" {
		node, err := e.store.Nodes.GetByID(ctx, nodeID)
		if err != nil {
			return fmt.Errorf("failed to get node: %w", err)
		}
		filename = fmt.Sprintf("cost-trend-%s-%s-%s-to-%s.%s",
			sanitizeFilename(node.Name),
			dimension,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
			format)
	}

	// Add prefix if configured
	if e.prefix != "" {
		filename = filepath.Join(e.prefix, filename)
	}

	// Create a temporary file to write to
	tempFile, err := os.CreateTemp("", "finops-chart-*."+format)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Render the chart
	if err := e.renderer.RenderCostTrend(ctx, nodeID, startDate, endDate, dimension, tempFile, format); err != nil {
		return fmt.Errorf("failed to render cost trend: %w", err)
	}

	// Reopen file for reading
	tempFile.Close()
	file, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to reopen temp file: %w", err)
	}
	defer file.Close()

	// Upload to storage
	writer, err := e.bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return fmt.Errorf("failed to create storage writer: %w", err)
	}
	defer writer.Close()

	// Set content type
	contentType := "image/png"
	if format == "svg" {
		contentType = "image/svg+xml"
	}
	writer.ContentType = contentType

	// Copy file to storage
	if _, err := file.WriteTo(writer); err != nil {
		return fmt.Errorf("failed to write to storage: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close storage writer: %w", err)
	}

	log.Info().
		Str("filename", filename).
		Str("format", format).
		Msg("Cost trend chart exported successfully")

	return nil
}

// ExportAllocationWaterfall exports a waterfall chart showing cost allocation breakdown
func (e *Exporter) ExportAllocationWaterfall(ctx context.Context, nodeID uuid.UUID, date time.Time, runID uuid.UUID, filename, format string) error {
	log.Info().
		Str("node_id", nodeID.String()).
		Time("date", date).
		Str("run_id", runID.String()).
		Str("filename", filename).
		Str("format", format).
		Msg("Exporting allocation waterfall chart")

	// Ensure format is supported
	if format != "png" && format != "svg" {
		return fmt.Errorf("unsupported format: %s (supported: png, svg)", format)
	}

	// Get node name for filename if not provided
	if filename == "" {
		node, err := e.store.Nodes.GetByID(ctx, nodeID)
		if err != nil {
			return fmt.Errorf("failed to get node: %w", err)
		}
		filename = fmt.Sprintf("allocation-waterfall-%s-%s.%s",
			sanitizeFilename(node.Name),
			date.Format("2006-01-02"),
			format)
	}

	// Add prefix if configured
	if e.prefix != "" {
		filename = filepath.Join(e.prefix, filename)
	}

	// Create a temporary file to write to
	tempFile, err := os.CreateTemp("", "finops-chart-*."+format)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Render the chart
	if err := e.renderer.RenderAllocationWaterfall(ctx, nodeID, date, runID, tempFile, format); err != nil {
		return fmt.Errorf("failed to render allocation waterfall: %w", err)
	}

	// Reopen file for reading
	tempFile.Close()
	file, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to reopen temp file: %w", err)
	}
	defer file.Close()

	// Upload to storage
	writer, err := e.bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return fmt.Errorf("failed to create storage writer: %w", err)
	}
	defer writer.Close()

	// Set content type
	contentType := "image/png"
	if format == "svg" {
		contentType = "image/svg+xml"
	}
	writer.ContentType = contentType

	// Copy file to storage
	if _, err := file.WriteTo(writer); err != nil {
		return fmt.Errorf("failed to write to storage: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close storage writer: %w", err)
	}

	log.Info().
		Str("filename", filename).
		Str("format", format).
		Msg("Allocation waterfall chart exported successfully")

	return nil
}

// GetStorageURL returns the public URL for a file (if supported by the storage backend)
func (e *Exporter) GetStorageURL(filename string) string {
	if e.prefix != "" {
		filename = filepath.Join(e.prefix, filename)
	}
	
	// For file:// storage, return local path
	// For cloud storage, this would need to be implemented based on the provider
	return filename
}

// ListExportedFiles lists all exported chart files
func (e *Exporter) ListExportedFiles(ctx context.Context) ([]string, error) {
	var files []string
	
	iter := e.bucket.List(&blob.ListOptions{
		Prefix: e.prefix,
	})
	
	for {
		obj, err := iter.Next(ctx)
		if err != nil {
			break
		}
		files = append(files, obj.Key)
	}
	
	return files, nil
}

// sanitizeFilename removes invalid characters from filenames
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := filename
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
