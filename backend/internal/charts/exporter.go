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
)

// Exporter handles chart generation and export to local files
type Exporter struct {
	store    *store.Store
	renderer *GraphRenderer
	prefix   string
}

// NewExporter creates a new chart exporter for local file output
func NewExporter(store *store.Store, storageURL, prefix string) (*Exporter, error) {
	// Note: storageURL parameter is ignored, we write directly to local files

	return &Exporter{
		store:    store,
		renderer: NewGraphRenderer(store),
		prefix:   prefix,
	}, nil
}

// Close closes the exporter and cleans up resources
func (e *Exporter) Close() error {
	// No cleanup needed for direct file writing
	return nil
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
		// Create directory if it doesn't exist
		if err := os.MkdirAll(e.prefix, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		filename = filepath.Join(e.prefix, filename)
	}

	// Create the output file directly
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Render the graph directly to the file
	if err := e.renderer.RenderGraphStructure(ctx, date, file, format); err != nil {
		return fmt.Errorf("failed to render graph structure: %w", err)
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

	// Create directory if needed
	if e.prefix != "" {
		if err := os.MkdirAll(e.prefix, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create the output file directly
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Render the chart directly to the file
	if err := e.renderer.RenderCostTrend(ctx, nodeID, startDate, endDate, dimension, file, format); err != nil {
		return fmt.Errorf("failed to render cost trend: %w", err)
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

	// Create directory if needed
	if e.prefix != "" {
		if err := os.MkdirAll(e.prefix, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create the output file directly
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Render the chart directly to the file
	if err := e.renderer.RenderAllocationWaterfall(ctx, nodeID, date, runID, file, format); err != nil {
		return fmt.Errorf("failed to render allocation waterfall: %w", err)
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

// ListExportedFiles lists all exported chart files in the local directory
func (e *Exporter) ListExportedFiles(ctx context.Context) ([]string, error) {
	var files []string

	// Use the prefix directory or current directory
	searchDir := "."
	if e.prefix != "" {
		searchDir = e.prefix
	}

	// Check if directory exists
	if _, err := os.Stat(searchDir); os.IsNotExist(err) {
		return files, nil // Return empty list if directory doesn't exist
	}

	// Read directory contents
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Filter for chart files (png, svg)
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".svg") {
				if e.prefix != "" {
					files = append(files, filepath.Join(e.prefix, name))
				} else {
					files = append(files, name)
				}
			}
		}
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
