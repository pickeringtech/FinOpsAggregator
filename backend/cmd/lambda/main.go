// Package main provides AWS Lambda handlers for FinOps Aggregator.
// The handler type is determined by the FINOPS_LAMBDA_HANDLER environment variable.
//
// This binary supports two modes:
// 1. AWS Lambda runtime mode (default) - invoked by AWS Lambda
// 2. Local/CLI mode - reads event from stdin, useful for Docker testing
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pickeringtech/FinOpsAggregator/internal/config"
	"github.com/pickeringtech/FinOpsAggregator/internal/logging"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
)

// HandlerType represents the type of Lambda handler to use.
type HandlerType string

const (
	HandlerImportAWSCUR    HandlerType = "import_awscur"
	HandlerImportDynatrace HandlerType = "import_dynatrace"
	HandlerExport          HandlerType = "export"
	HandlerAllocate        HandlerType = "allocate"
)

var (
	cfg *config.Config
	db  *store.DB
	st  *store.Store
)

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load("")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logging
	logging.Init(cfg.Logging)

	// Initialize database connection
	db, err = store.NewDB(cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize store
	st = store.NewStore(db)

	// Determine handler type from environment
	handlerType := HandlerType(os.Getenv("FINOPS_LAMBDA_HANDLER"))
	if handlerType == "" {
		// Try to infer from FINOPS_IMPORT_SOURCE for backward compatibility
		importSource := os.Getenv("FINOPS_IMPORT_SOURCE")
		switch importSource {
		case "aws_cur":
			handlerType = HandlerImportAWSCUR
		case "dynatrace":
			handlerType = HandlerImportDynatrace
		default:
			log.Fatal().Msg("FINOPS_LAMBDA_HANDLER environment variable not set")
		}
	}

	log.Info().
		Str("handler", string(handlerType)).
		Msg("Starting Lambda handler")

	// Check if we're running in local/CLI mode (stdin has data or LOCAL_MODE is set)
	localMode := os.Getenv("LOCAL_MODE") == "true" || os.Getenv("_LAMBDA_SERVER_PORT") == ""

	if localMode {
		// Local mode: read event from stdin and invoke handler directly
		runLocalMode(handlerType)
	} else {
		// Lambda runtime mode: use AWS Lambda SDK
		switch handlerType {
		case HandlerImportAWSCUR:
			lambda.Start(handleImportAWSCUR)
		case HandlerImportDynatrace:
			lambda.Start(handleImportDynatrace)
		case HandlerExport:
			lambda.Start(handleExport)
		case HandlerAllocate:
			lambda.Start(handleAllocate)
		default:
			log.Fatal().Str("handler", string(handlerType)).Msg("Unknown handler type")
		}
	}
}

// runLocalMode runs the handler in local/CLI mode, reading event from stdin.
func runLocalMode(handlerType HandlerType) {
	ctx := context.Background()

	// Read event from stdin
	reader := bufio.NewReader(os.Stdin)
	eventData, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		log.Fatal().Err(err).Msg("Failed to read event from stdin")
	}

	// If no event data, use empty JSON object
	if len(eventData) == 0 {
		eventData = []byte("{}")
	}

	log.Debug().
		Str("handler", string(handlerType)).
		RawJSON("event", eventData).
		Msg("Processing local event")

	var response interface{}

	switch handlerType {
	case HandlerImportAWSCUR:
		var event events.S3Event
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Fatal().Err(err).Msg("Failed to parse S3 event")
		}
		response, err = handleImportAWSCUR(ctx, event)

	case HandlerImportDynatrace:
		var event events.S3Event
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Fatal().Err(err).Msg("Failed to parse S3 event")
		}
		response, err = handleImportDynatrace(ctx, event)

	case HandlerExport:
		var request ExportRequest
		if err := json.Unmarshal(eventData, &request); err != nil {
			log.Fatal().Err(err).Msg("Failed to parse export request")
		}
		response, err = handleExport(ctx, request)

	case HandlerAllocate:
		var request AllocateRequest
		if err := json.Unmarshal(eventData, &request); err != nil {
			log.Fatal().Err(err).Msg("Failed to parse allocate request")
		}
		response, err = handleAllocate(ctx, request)

	default:
		log.Fatal().Str("handler", string(handlerType)).Msg("Unknown handler type")
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler returned error")
		os.Exit(1)
	}

	// Output response as JSON
	output, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(output))
}

// initStorage initializes blob storage for the handler.
func initStorage(ctx context.Context) error {
	// Storage is initialized per-request if needed
	return nil
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns the environment variable as a boolean.
func getEnvBool(key string) bool {
	value := os.Getenv(key)
	return value == "true" || value == "1" || value == "yes"
}

// LambdaResponse represents a standard Lambda response.
type LambdaResponse struct {
	StatusCode int               `json:"statusCode"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// newSuccessResponse creates a successful Lambda response.
func newSuccessResponse(body string) LambdaResponse {
	return LambdaResponse{
		StatusCode: 200,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// newErrorResponse creates an error Lambda response.
func newErrorResponse(statusCode int, err error) LambdaResponse {
	return LambdaResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

