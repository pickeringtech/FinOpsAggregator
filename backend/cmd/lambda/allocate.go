package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/allocate"
	"github.com/rs/zerolog/log"
)

// AllocateRequest represents a request to run cost allocation.
type AllocateRequest struct {
	// StartDate is the start of the allocation period (YYYY-MM-DD)
	StartDate string `json:"start_date"`
	// EndDate is the end of the allocation period (YYYY-MM-DD)
	EndDate string `json:"end_date"`
	// Dimensions is an optional list of dimensions to allocate (defaults to config)
	Dimensions []string `json:"dimensions,omitempty"`
}

// AllocateResponse represents the response from an allocation operation.
type AllocateResponse struct {
	Success        bool   `json:"success"`
	RunID          string `json:"run_id,omitempty"`
	ProcessedDays  int    `json:"processed_days,omitempty"`
	Allocations    int    `json:"allocations,omitempty"`
	Contributions  int    `json:"contributions,omitempty"`
	ProcessingTime string `json:"processing_time,omitempty"`
	Error          string `json:"error,omitempty"`
}

// handleAllocate handles allocation requests.
func handleAllocate(ctx context.Context, request AllocateRequest) (LambdaResponse, error) {
	log.Info().
		Str("start_date", request.StartDate).
		Str("end_date", request.EndDate).
		Strs("dimensions", request.Dimensions).
		Msg("Processing allocation request")

	// Parse dates
	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		return newErrorResponse(400, fmt.Errorf("invalid start_date: %w", err)), nil
	}

	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		return newErrorResponse(400, fmt.Errorf("invalid end_date: %w", err)), nil
	}

	// Validate date range
	if endDate.Before(startDate) {
		return newErrorResponse(400, fmt.Errorf("end_date must be after start_date")), nil
	}

	// Use provided dimensions or fall back to config
	dimensions := request.Dimensions
	if len(dimensions) == 0 {
		dimensions = cfg.Compute.ActiveDimensions
	}

	// Create allocation engine
	engine := allocate.NewEngine(st)

	// Run allocation
	result, err := engine.AllocateForPeriod(ctx, startDate, endDate, dimensions)
	if err != nil {
		log.Error().Err(err).Msg("Allocation failed")
		response := AllocateResponse{
			Success: false,
			Error:   err.Error(),
		}
		body, _ := json.Marshal(response)
		return newErrorResponse(500, fmt.Errorf(string(body))), nil
	}

	log.Info().
		Str("run_id", result.RunID.String()).
		Int("processed_days", result.Summary.ProcessedDays).
		Int("allocations", len(result.Allocations)).
		Int("contributions", len(result.Contributions)).
		Dur("processing_time", result.Summary.ProcessingTime).
		Msg("Allocation completed")

	response := AllocateResponse{
		Success:        true,
		RunID:          result.RunID.String(),
		ProcessedDays:  result.Summary.ProcessedDays,
		Allocations:    len(result.Allocations),
		Contributions:  len(result.Contributions),
		ProcessingTime: result.Summary.ProcessingTime.String(),
	}

	body, _ := json.Marshal(response)
	return newSuccessResponse(string(body)), nil
}

