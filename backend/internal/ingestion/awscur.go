package ingestion

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/models"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// AWSCURIngester handles ingestion of AWS Cost and Usage Report CSV files
type AWSCURIngester struct {
	store           *store.Store
	columnMappings  map[string]string
	progressChan    chan IngestionProgress
	batchSize       int
	createMissingNodes bool
}

// IngestionProgress reports progress during ingestion
type IngestionProgress struct {
	RecordsProcessed int
	RecordsInserted  int
	RecordsSkipped   int
	RecordsErrored   int
	CurrentRecord    int
	TotalRecords     int
	CurrentFile      string
	Message          string
	Error            error
}

// AWSCURConfig configures the AWS CUR ingester
type AWSCURConfig struct {
	BatchSize          int
	CreateMissingNodes bool
	ProgressChan       chan IngestionProgress
}

// Standard AWS CUR column names
const (
	ColLineItemUsageStartDate    = "lineItem/UsageStartDate"
	ColLineItemUsageEndDate      = "lineItem/UsageEndDate"
	ColLineItemProductCode       = "lineItem/ProductCode"
	ColLineItemUsageType         = "lineItem/UsageType"
	ColLineItemOperation         = "lineItem/Operation"
	ColLineItemResourceId        = "lineItem/ResourceId"
	ColLineItemUnblendedCost     = "lineItem/UnblendedCost"
	ColLineItemBlendedCost       = "lineItem/BlendedCost"
	ColLineItemUsageAmount       = "lineItem/UsageAmount"
	ColProductProductName        = "product/ProductName"
	ColProductRegion             = "product/region"
	ColResourceTagsUserName      = "resourceTags/user:Name"
	ColResourceTagsUserProduct   = "resourceTags/user:Product"
	ColResourceTagsUserService   = "resourceTags/user:Service"
	ColResourceTagsUserCostCenter = "resourceTags/user:CostCenter"
)

// NewAWSCURIngester creates a new AWS CUR ingester
func NewAWSCURIngester(store *store.Store, config *AWSCURConfig) *AWSCURIngester {
	batchSize := 1000
	if config != nil && config.BatchSize > 0 {
		batchSize = config.BatchSize
	}

	createMissingNodes := false
	if config != nil {
		createMissingNodes = config.CreateMissingNodes
	}

	var progressChan chan IngestionProgress
	if config != nil {
		progressChan = config.ProgressChan
	}

	return &AWSCURIngester{
		store:              store,
		batchSize:          batchSize,
		createMissingNodes: createMissingNodes,
		progressChan:       progressChan,
		columnMappings: map[string]string{
			// Map AWS CUR columns to internal field names
			ColLineItemUsageStartDate:    "usage_start_date",
			ColLineItemProductCode:       "product_code",
			ColLineItemUsageType:         "usage_type",
			ColLineItemUnblendedCost:     "unblended_cost",
			ColLineItemResourceId:        "resource_id",
			ColResourceTagsUserProduct:   "product_tag",
			ColResourceTagsUserService:   "service_tag",
			ColResourceTagsUserCostCenter: "cost_center_tag",
		},
	}
}

// IngestFile ingests an AWS CUR CSV file
func (a *AWSCURIngester) IngestFile(ctx context.Context, filePath string) (*IngestionResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size for progress reporting
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	a.reportProgress(IngestionProgress{
		CurrentFile: filePath,
		Message:     fmt.Sprintf("Starting ingestion of %s (%.2f MB)", filePath, float64(stat.Size())/(1024*1024)),
	})

	return a.IngestReader(ctx, file, filePath)
}

// IngestReader ingests AWS CUR data from a reader
func (a *AWSCURIngester) IngestReader(ctx context.Context, reader io.Reader, sourceName string) (*IngestionResult, error) {
	result := &IngestionResult{
		Source:    "aws_cur",
		StartTime: time.Now(),
	}

	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	// Read header row
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Build column index map
	colIndex := make(map[string]int)
	for i, header := range headers {
		colIndex[header] = i
	}

	// Validate required columns exist
	requiredCols := []string{ColLineItemUsageStartDate, ColLineItemUnblendedCost}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("required column %s not found in CSV", col)
		}
	}

	// Cache for node lookups
	nodeCache := make(map[string]*models.CostNode)

	// Batch for bulk insert
	var costBatch []models.NodeCostByDimension
	recordNum := 0

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", recordNum+1, err))
			result.RecordsSkipped++
			continue
		}

		recordNum++
		result.RecordsProcessed++

		// Parse the record
		cost, err := a.parseRecord(ctx, record, colIndex, nodeCache)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", recordNum, err))
			result.RecordsSkipped++
			continue
		}

		if cost != nil {
			costBatch = append(costBatch, *cost)
		}

		// Batch insert when batch is full
		if len(costBatch) >= a.batchSize {
			if err := a.store.Costs.BulkUpsert(ctx, costBatch); err != nil {
				return nil, fmt.Errorf("failed to bulk insert costs: %w", err)
			}
			result.RecordsInserted += len(costBatch)
			
			a.reportProgress(IngestionProgress{
				RecordsProcessed: result.RecordsProcessed,
				RecordsInserted:  result.RecordsInserted,
				RecordsSkipped:   result.RecordsSkipped,
				CurrentRecord:    recordNum,
				Message:          fmt.Sprintf("Processed %d records, inserted %d", result.RecordsProcessed, result.RecordsInserted),
			})
			
			costBatch = costBatch[:0]
		}
	}

	// Insert remaining records
	if len(costBatch) > 0 {
		if err := a.store.Costs.BulkUpsert(ctx, costBatch); err != nil {
			return nil, fmt.Errorf("failed to bulk insert final costs: %w", err)
		}
		result.RecordsInserted += len(costBatch)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	a.reportProgress(IngestionProgress{
		RecordsProcessed: result.RecordsProcessed,
		RecordsInserted:  result.RecordsInserted,
		RecordsSkipped:   result.RecordsSkipped,
		Message:          fmt.Sprintf("Ingestion complete: %d processed, %d inserted, %d skipped", result.RecordsProcessed, result.RecordsInserted, result.RecordsSkipped),
	})

	log.Info().
		Int("processed", result.RecordsProcessed).
		Int("inserted", result.RecordsInserted).
		Int("skipped", result.RecordsSkipped).
		Dur("duration", result.Duration).
		Msg("AWS CUR ingestion completed")

	return result, nil
}

// parseRecord parses a single CSV record into a NodeCostByDimension
func (a *AWSCURIngester) parseRecord(ctx context.Context, record []string, colIndex map[string]int, nodeCache map[string]*models.CostNode) (*models.NodeCostByDimension, error) {
	// Get usage start date
	usageStartDateStr := a.getColumn(record, colIndex, ColLineItemUsageStartDate)
	if usageStartDateStr == "" {
		return nil, fmt.Errorf("missing usage start date")
	}

	usageDate, err := a.parseDate(usageStartDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid usage start date: %w", err)
	}

	// Get cost amount
	costStr := a.getColumn(record, colIndex, ColLineItemUnblendedCost)
	if costStr == "" {
		return nil, nil // Skip records with no cost
	}

	cost, err := decimal.NewFromString(costStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cost value: %w", err)
	}

	// Skip zero-cost records
	if cost.IsZero() {
		return nil, nil
	}

	// Determine the node to associate this cost with
	node, err := a.resolveNode(ctx, record, colIndex, nodeCache)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node: %w", err)
	}
	if node == nil {
		return nil, nil // Skip if no node found and not creating missing nodes
	}

	// Build dimension from usage type or product code
	dimension := a.buildDimension(record, colIndex)

	// Build metadata
	metadata := a.buildMetadata(record, colIndex)

	return &models.NodeCostByDimension{
		NodeID:    node.ID,
		CostDate:  usageDate,
		Dimension: dimension,
		Amount:    cost,
		Currency:  "USD",
		Metadata:  metadata,
	}, nil
}

// getColumn safely gets a column value from a record
func (a *AWSCURIngester) getColumn(record []string, colIndex map[string]int, colName string) string {
	if idx, ok := colIndex[colName]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

// parseDate parses AWS CUR date formats
func (a *AWSCURIngester) parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// Truncate to day
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// resolveNode finds or creates the node to associate costs with
func (a *AWSCURIngester) resolveNode(ctx context.Context, record []string, colIndex map[string]int, nodeCache map[string]*models.CostNode) (*models.CostNode, error) {
	// Try to find node by product tag first
	productTag := a.getColumn(record, colIndex, ColResourceTagsUserProduct)
	if productTag != "" {
		if node, ok := nodeCache[productTag]; ok {
			return node, nil
		}
		node, err := a.store.Nodes.GetByName(ctx, productTag)
		if err == nil {
			nodeCache[productTag] = node
			return node, nil
		}
	}

	// Try service tag
	serviceTag := a.getColumn(record, colIndex, ColResourceTagsUserService)
	if serviceTag != "" {
		if node, ok := nodeCache[serviceTag]; ok {
			return node, nil
		}
		node, err := a.store.Nodes.GetByName(ctx, serviceTag)
		if err == nil {
			nodeCache[serviceTag] = node
			return node, nil
		}
	}

	// Try cost center tag
	costCenterTag := a.getColumn(record, colIndex, ColResourceTagsUserCostCenter)
	if costCenterTag != "" {
		if node, ok := nodeCache[costCenterTag]; ok {
			return node, nil
		}
		node, err := a.store.Nodes.GetByName(ctx, costCenterTag)
		if err == nil {
			nodeCache[costCenterTag] = node
			return node, nil
		}
	}

	// Fall back to product code (AWS service name)
	productCode := a.getColumn(record, colIndex, ColLineItemProductCode)
	if productCode != "" {
		nodeName := "aws_" + strings.ToLower(productCode)
		if node, ok := nodeCache[nodeName]; ok {
			return node, nil
		}
		node, err := a.store.Nodes.GetByName(ctx, nodeName)
		if err == nil {
			nodeCache[nodeName] = node
			return node, nil
		}

		// Create missing node if configured
		if a.createMissingNodes {
			newNode := &models.CostNode{
				ID:         uuid.New(),
				Name:       nodeName,
				Type:       string(models.NodeTypeResource),
				IsPlatform: false,
				CostLabels: map[string]interface{}{
					"aws_product_code": productCode,
				},
				Metadata: map[string]interface{}{
					"source":     "aws_cur_import",
					"created_by": "auto",
				},
			}
			if err := a.store.Nodes.Create(ctx, newNode); err != nil {
				return nil, fmt.Errorf("failed to create node: %w", err)
			}
			nodeCache[nodeName] = newNode
			log.Info().Str("node_name", nodeName).Msg("Created new node from AWS CUR import")
			return newNode, nil
		}
	}

	return nil, nil
}

// buildDimension creates a dimension string from the record
func (a *AWSCURIngester) buildDimension(record []string, colIndex map[string]int) string {
	usageType := a.getColumn(record, colIndex, ColLineItemUsageType)
	if usageType != "" {
		// Normalize usage type to a dimension name
		// e.g., "USW2-BoxUsage:t3.medium" -> "box_usage"
		parts := strings.Split(usageType, ":")
		if len(parts) > 0 {
			// Take the part after the region prefix
			dimPart := parts[0]
			if idx := strings.Index(dimPart, "-"); idx != -1 {
				dimPart = dimPart[idx+1:]
			}
			return strings.ToLower(strings.ReplaceAll(dimPart, "-", "_"))
		}
	}

	productCode := a.getColumn(record, colIndex, ColLineItemProductCode)
	if productCode != "" {
		return strings.ToLower(productCode) + "_cost"
	}

	return "unclassified"
}

// buildMetadata creates metadata from the record
func (a *AWSCURIngester) buildMetadata(record []string, colIndex map[string]int) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Add relevant fields to metadata
	if v := a.getColumn(record, colIndex, ColLineItemProductCode); v != "" {
		metadata["product_code"] = v
	}
	if v := a.getColumn(record, colIndex, ColLineItemUsageType); v != "" {
		metadata["usage_type"] = v
	}
	if v := a.getColumn(record, colIndex, ColLineItemOperation); v != "" {
		metadata["operation"] = v
	}
	if v := a.getColumn(record, colIndex, ColLineItemResourceId); v != "" {
		metadata["resource_id"] = v
	}
	if v := a.getColumn(record, colIndex, ColProductRegion); v != "" {
		metadata["region"] = v
	}
	if v := a.getColumn(record, colIndex, ColResourceTagsUserName); v != "" {
		metadata["resource_name"] = v
	}

	metadata["source"] = "aws_cur"

	return metadata
}

// reportProgress sends progress updates if a channel is configured
func (a *AWSCURIngester) reportProgress(progress IngestionProgress) {
	if a.progressChan != nil {
		select {
		case a.progressChan <- progress:
		default:
			// Don't block if channel is full
		}
	}
}

