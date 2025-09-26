package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
)

// ValidationResult represents the result of graph validation
type ValidationResult struct {
	Valid       bool                    `json:"valid"`
	Errors      []ValidationError       `json:"errors,omitempty"`
	Warnings    []ValidationWarning     `json:"warnings,omitempty"`
	Stats       GraphStats              `json:"stats"`
	Date        time.Time               `json:"date"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	NodeID      *uuid.UUID `json:"node_id,omitempty"`
	EdgeID      *uuid.UUID `json:"edge_id,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	NodeID      *uuid.UUID `json:"node_id,omitempty"`
	EdgeID      *uuid.UUID `json:"edge_id,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Validator validates graph structure and consistency
type Validator struct {
	store   *store.Store
	builder *GraphBuilder
}

// NewValidator creates a new graph validator
func NewValidator(store *store.Store) *Validator {
	return &Validator{
		store:   store,
		builder: NewGraphBuilder(store),
	}
}

// ValidateForDate validates the graph for a specific date
func (v *Validator) ValidateForDate(ctx context.Context, date time.Time) (*ValidationResult, error) {
	log.Info().Time("date", date).Msg("Starting graph validation")

	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Date:     date,
	}

	// Build the graph
	graph, err := v.builder.BuildForDate(ctx, date)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "graph_build_error",
			Message: fmt.Sprintf("Failed to build graph: %v", err),
		})
		return result, nil
	}

	result.Stats = graph.Stats()

	// Run validation checks
	v.validateDAGStructure(graph, result)
	v.validateNodeReferences(ctx, graph, result)
	v.validateEdgeConsistency(ctx, graph, result)
	v.validateIsolatedNodes(graph, result)
	v.validatePlatformNodes(graph, result)

	log.Info().
		Bool("valid", result.Valid).
		Int("errors", len(result.Errors)).
		Int("warnings", len(result.Warnings)).
		Msg("Graph validation completed")

	return result, nil
}

// validateDAGStructure validates that the graph is a valid DAG
func (v *Validator) validateDAGStructure(graph *Graph, result *ValidationResult) {
	if err := graph.ValidateDAG(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "cycle_detected",
			Message: "Graph contains cycles, which violates DAG requirements",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
	}

	// Check for self-loops (should be prevented by DB constraints, but double-check)
	for nodeID, edges := range graph.edges {
		for _, edge := range edges {
			if edge.ChildID == nodeID {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Type:    "self_loop",
					Message: "Node has an edge to itself",
					NodeID:  &nodeID,
					EdgeID:  &edge.ID,
				})
			}
		}
	}
}

// validateNodeReferences validates that all edge references point to existing nodes
func (v *Validator) validateNodeReferences(ctx context.Context, graph *Graph, result *ValidationResult) {
	// Get all edges for the date (including those that might reference missing nodes)
	edges, err := v.store.Edges.GetActiveEdgesForDate(ctx, graph.Date())
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "edge_query_error",
			Message: fmt.Sprintf("Failed to query edges: %v", err),
		})
		return
	}

	for _, edge := range edges {
		// Check parent exists
		if _, exists := graph.nodes[edge.ParentID]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "missing_parent_node",
				Message: "Edge references non-existent parent node",
				NodeID:  &edge.ParentID,
				EdgeID:  &edge.ID,
				Details: map[string]interface{}{
					"parent_id": edge.ParentID.String(),
					"child_id":  edge.ChildID.String(),
				},
			})
		}

		// Check child exists
		if _, exists := graph.nodes[edge.ChildID]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "missing_child_node",
				Message: "Edge references non-existent child node",
				NodeID:  &edge.ChildID,
				EdgeID:  &edge.ID,
				Details: map[string]interface{}{
					"parent_id": edge.ParentID.String(),
					"child_id":  edge.ChildID.String(),
				},
			})
		}
	}
}

// validateEdgeConsistency validates edge date ranges and overlaps
func (v *Validator) validateEdgeConsistency(ctx context.Context, graph *Graph, result *ValidationResult) {
	// Group edges by parent-child pair
	edgePairs := make(map[string][]uuid.UUID)
	
	for _, edges := range graph.edges {
		for _, edge := range edges {
			key := fmt.Sprintf("%s-%s", edge.ParentID.String(), edge.ChildID.String())
			edgePairs[key] = append(edgePairs[key], edge.ID)
		}
	}

	// Check for multiple active edges between same nodes
	for pairKey, edgeIDs := range edgePairs {
		if len(edgeIDs) > 1 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "multiple_active_edges",
				Message: "Multiple active edges found between same node pair",
				Details: map[string]interface{}{
					"pair":     pairKey,
					"edge_ids": edgeIDs,
					"count":    len(edgeIDs),
				},
			})
		}
	}
}

// validateIsolatedNodes identifies nodes with no connections
func (v *Validator) validateIsolatedNodes(graph *Graph, result *ValidationResult) {
	for nodeID, node := range graph.nodes {
		hasIncoming := len(graph.incoming[nodeID]) > 0
		hasOutgoing := len(graph.edges[nodeID]) > 0

		if !hasIncoming && !hasOutgoing {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "isolated_node",
				Message: "Node has no incoming or outgoing edges",
				NodeID:  &nodeID,
				Details: map[string]interface{}{
					"node_name": node.Name,
					"node_type": node.Type,
				},
			})
		}
	}
}

// validatePlatformNodes validates platform node configurations
func (v *Validator) validatePlatformNodes(graph *Graph, result *ValidationResult) {
	platformNodes := 0
	
	for nodeID, node := range graph.nodes {
		if node.IsPlatform {
			platformNodes++
			
			// Platform nodes should typically be leaf nodes (no outgoing edges)
			if len(graph.edges[nodeID]) > 0 {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:    "platform_node_has_children",
					Message: "Platform node has outgoing edges, which may indicate incorrect modeling",
					NodeID:  &nodeID,
					Details: map[string]interface{}{
						"node_name":    node.Name,
						"child_count":  len(graph.edges[nodeID]),
					},
				})
			}
		}
	}

	if platformNodes == 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "no_platform_nodes",
			Message: "No platform nodes found - consider marking shared infrastructure as platform nodes",
		})
	}
}

// ValidateCurrentGraph validates the graph for the current date
func (v *Validator) ValidateCurrentGraph(ctx context.Context) (*ValidationResult, error) {
	return v.ValidateForDate(ctx, time.Now())
}

// ValidateGraphHistory validates the graph across a date range
func (v *Validator) ValidateGraphHistory(ctx context.Context, startDate, endDate time.Time) ([]ValidationResult, error) {
	log.Info().
		Time("start_date", startDate).
		Time("end_date", endDate).
		Msg("Starting graph history validation")

	var results []ValidationResult
	
	// Validate for each day in the range
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		result, err := v.ValidateForDate(ctx, date)
		if err != nil {
			return nil, fmt.Errorf("failed to validate graph for date %s: %w", date.Format("2006-01-02"), err)
		}
		results = append(results, *result)
	}

	log.Info().
		Int("days_validated", len(results)).
		Msg("Graph history validation completed")

	return results, nil
}

// GetValidationSummary returns a summary of validation results
func GetValidationSummary(results []ValidationResult) ValidationSummary {
	summary := ValidationSummary{
		TotalDays:    len(results),
		ValidDays:    0,
		TotalErrors:  0,
		TotalWarnings: 0,
		ErrorTypes:   make(map[string]int),
		WarningTypes: make(map[string]int),
	}

	for _, result := range results {
		if result.Valid {
			summary.ValidDays++
		}
		
		summary.TotalErrors += len(result.Errors)
		summary.TotalWarnings += len(result.Warnings)
		
		for _, err := range result.Errors {
			summary.ErrorTypes[err.Type]++
		}
		
		for _, warn := range result.Warnings {
			summary.WarningTypes[warn.Type]++
		}
	}

	return summary
}

// ValidationSummary provides a summary of validation results across multiple days
type ValidationSummary struct {
	TotalDays     int            `json:"total_days"`
	ValidDays     int            `json:"valid_days"`
	TotalErrors   int            `json:"total_errors"`
	TotalWarnings int            `json:"total_warnings"`
	ErrorTypes    map[string]int `json:"error_types"`
	WarningTypes  map[string]int `json:"warning_types"`
}
