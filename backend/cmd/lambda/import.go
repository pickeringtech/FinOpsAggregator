package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/ingestion"
	"github.com/pickeringtech/FinOpsAggregator/internal/storage"
	"github.com/rs/zerolog/log"
)

// buildS3URL constructs an S3 URL with optional LocalStack endpoint configuration.
// For LocalStack, it adds the endpoint and path-style parameters.
func buildS3URL(bucket string) string {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	if endpoint == "" {
		// Production: use standard S3 URL
		return fmt.Sprintf("s3://%s", bucket)
	}

	// LocalStack: add endpoint and path-style parameters
	// GoCloud s3blob uses: endpoint, disable_https, use_path_style (or s3ForcePathStyle for V1 compat)
	return fmt.Sprintf("s3://%s?endpoint=%s&disable_https=true&s3ForcePathStyle=true", bucket, endpoint)
}

// handleImportAWSCUR handles S3 events for AWS CUR file imports.
func handleImportAWSCUR(ctx context.Context, s3Event events.S3Event) (LambdaResponse, error) {
	log.Info().
		Int("records", len(s3Event.Records)).
		Msg("Processing AWS CUR import event")

	var totalProcessed, totalInserted, totalSkipped int
	var errors []string

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		log.Info().
			Str("bucket", bucket).
			Str("key", key).
			Msg("Processing S3 object")

		result, err := processAWSCURFile(ctx, bucket, key)
		if err != nil {
			log.Error().Err(err).
				Str("bucket", bucket).
				Str("key", key).
				Msg("Failed to process AWS CUR file")
			errors = append(errors, fmt.Sprintf("%s/%s: %v", bucket, key, err))
			continue
		}

		totalProcessed += result.RecordsProcessed
		totalInserted += result.RecordsInserted
		totalSkipped += result.RecordsSkipped
		errors = append(errors, result.Errors...)
	}

	response := map[string]interface{}{
		"message":           "AWS CUR import completed",
		"records_processed": totalProcessed,
		"records_inserted":  totalInserted,
		"records_skipped":   totalSkipped,
		"errors":            errors,
	}

	body, _ := json.Marshal(response)
	return newSuccessResponse(string(body)), nil
}

// processAWSCURFile processes a single AWS CUR file from S3.
func processAWSCURFile(ctx context.Context, bucket, key string) (*ingestion.IngestionResult, error) {
	// Initialize storage to read from S3
	storageURL := buildS3URL(bucket)
	blobStorage, err := storage.NewBlobStorage(ctx, storageURL, "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer blobStorage.Close()

	// Read the file from S3
	reader, err := blobStorage.ReadStream(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from S3: %w", err)
	}
	defer reader.Close()

	// Create ingester
	createNodes := getEnvBool("FINOPS_IMPORT_CREATE_NODES")
	ingester := ingestion.NewAWSCURIngester(st, &ingestion.AWSCURConfig{
		BatchSize:          1000,
		CreateMissingNodes: createNodes,
	})

	// Process the file
	sourceName := fmt.Sprintf("s3://%s/%s", bucket, key)
	result, err := ingester.IngestReader(ctx, reader, sourceName)
	if err != nil {
		return nil, fmt.Errorf("ingestion failed: %w", err)
	}

	log.Info().
		Str("source", sourceName).
		Int("processed", result.RecordsProcessed).
		Int("inserted", result.RecordsInserted).
		Int("skipped", result.RecordsSkipped).
		Dur("duration", result.Duration).
		Msg("AWS CUR file processed")

	return result, nil
}

// handleImportDynatrace handles S3 events for Dynatrace metrics file imports.
func handleImportDynatrace(ctx context.Context, s3Event events.S3Event) (LambdaResponse, error) {
	log.Info().
		Int("records", len(s3Event.Records)).
		Msg("Processing Dynatrace import event")

	var totalProcessed, totalInserted, totalSkipped int
	var errors []string

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		log.Info().
			Str("bucket", bucket).
			Str("key", key).
			Msg("Processing S3 object")

		result, err := processDynatraceFile(ctx, bucket, key)
		if err != nil {
			log.Error().Err(err).
				Str("bucket", bucket).
				Str("key", key).
				Msg("Failed to process Dynatrace file")
			errors = append(errors, fmt.Sprintf("%s/%s: %v", bucket, key, err))
			continue
		}

		totalProcessed += result.RecordsProcessed
		totalInserted += result.RecordsInserted
		totalSkipped += result.RecordsSkipped
		errors = append(errors, result.Errors...)
	}

	response := map[string]interface{}{
		"message":           "Dynatrace import completed",
		"records_processed": totalProcessed,
		"records_inserted":  totalInserted,
		"records_skipped":   totalSkipped,
		"errors":            errors,
	}

	body, _ := json.Marshal(response)
	return newSuccessResponse(string(body)), nil
}

// processDynatraceFile processes a single Dynatrace metrics file from S3.
func processDynatraceFile(ctx context.Context, bucket, key string) (*ingestion.IngestionResult, error) {
	// Initialize storage to read from S3
	storageURL := buildS3URL(bucket)
	blobStorage, err := storage.NewBlobStorage(ctx, storageURL, "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer blobStorage.Close()

	// Read the file from S3
	reader, err := blobStorage.ReadStream(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from S3: %w", err)
	}
	defer reader.Close()

	// Create ingester
	ingester := ingestion.NewDynatraceIngester(st)

	// For Dynatrace, we need a node ID. Try to get it from the filename or metadata.
	// The filename convention is: dynatrace/<node-id>.json or dynatrace/<node-name>.json
	nodeID, err := extractNodeIDFromKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to determine node ID from key %s: %w", key, err)
	}

	// Process the file
	result, err := ingester.IngestReader(ctx, reader, nodeID)
	if err != nil {
		return nil, fmt.Errorf("ingestion failed: %w", err)
	}

	log.Info().
		Str("bucket", bucket).
		Str("key", key).
		Str("node_id", nodeID.String()).
		Int("processed", result.RecordsProcessed).
		Int("inserted", result.RecordsInserted).
		Dur("duration", result.Duration).
		Msg("Dynatrace file processed")

	return result, nil
}

// extractNodeIDFromKey extracts the node ID from an S3 key.
// Expected format: dynatrace/<node-id>.json or dynatrace/<node-name>.json
func extractNodeIDFromKey(ctx context.Context, key string) (uuid.UUID, error) {
	// Extract filename from key
	var filename string
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '/' {
			filename = key[i+1:]
			break
		}
	}
	if filename == "" {
		filename = key
	}

	// Remove .json extension
	if len(filename) > 5 && filename[len(filename)-5:] == ".json" {
		filename = filename[:len(filename)-5]
	}

	// Try to parse as UUID
	nodeID, err := uuid.Parse(filename)
	if err == nil {
		return nodeID, nil
	}

	// Try to find node by name
	node, err := st.Nodes.GetByName(ctx, filename)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not find node by name %s: %w", filename, err)
	}

	return node.ID, nil
}

