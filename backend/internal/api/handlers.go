package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Handler provides HTTP handlers for the API
type Handler struct {
	service *Service
}

// NewHandler creates a new API handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: Get from build info
		Database:  "connected", // TODO: Check actual database connection
	}

	c.JSON(http.StatusOK, response)
}

// GetProductHierarchy handles requests for product hierarchy with cost data
func (h *Handler) GetProductHierarchy(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetProductHierarchy(c.Request.Context(), *req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get product hierarchy")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve product hierarchy")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetIndividualNode handles requests for individual node cost data
func (h *Handler) GetIndividualNode(c *gin.Context) {
	// Parse node ID from path parameter
	nodeIDStr := c.Param("nodeId")
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_node_id", "Invalid node ID format")
		return
	}

	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetIndividualNode(c.Request.Context(), nodeID, *req)
	if err != nil {
		log.Error().Err(err).Str("node_id", nodeID.String()).Msg("Failed to get individual node")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve node data")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPlatformServices handles requests for platform and shared services cost data
func (h *Handler) GetPlatformServices(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetPlatformServices(c.Request.Context(), *req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get platform services")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve platform services")
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListProducts handles requests for a flat list of products with costs
func (h *Handler) ListProducts(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.service.ListProducts(c.Request.Context(), *req, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list products")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve products")
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListNodes handles requests for a flat list of nodes with costs
func (h *Handler) ListNodes(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Parse filters and pagination
	nodeType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.service.ListNodes(c.Request.Context(), *req, nodeType, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list nodes")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve nodes")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCostsByType handles requests for costs aggregated by node type
func (h *Handler) GetCostsByType(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetCostsByType(c.Request.Context(), *req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get costs by type")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve costs by type")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCostsByDimension handles requests for costs aggregated by a custom dimension
func (h *Handler) GetCostsByDimension(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Get dimension key from query parameter
	dimensionKey := c.Query("key")
	if dimensionKey == "" {
		h.handleError(c, http.StatusBadRequest, "invalid_request", "dimension key is required")
		return
	}

	response, err := h.service.GetCostsByDimension(c.Request.Context(), *req, dimensionKey)
	if err != nil {
		log.Error().Err(err).Str("dimension_key", dimensionKey).Msg("Failed to get costs by dimension")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve costs by dimension")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRecommendations handles requests for cost optimization recommendations
func (h *Handler) GetRecommendations(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Optional node_id filter
	nodeIDStr := c.Query("node_id")
	var nodeID *uuid.UUID
	if nodeIDStr != "" {
		parsed, err := uuid.Parse(nodeIDStr)
		if err != nil {
			h.handleError(c, http.StatusBadRequest, "invalid_request", "invalid node_id format")
			return
		}
		nodeID = &parsed
	}

	response, err := h.service.GetRecommendations(c.Request.Context(), *req, nodeID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recommendations")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve recommendations")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardSummary handles requests for the dashboard summary with correct totals
func (h *Handler) GetDashboardSummary(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetDashboardSummary(c.Request.Context(), *req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get dashboard summary")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve dashboard summary")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetInfrastructureHierarchy handles requests for the infrastructure hierarchy view
func (h *Handler) GetInfrastructureHierarchy(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetInfrastructureHierarchy(c.Request.Context(), *req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get infrastructure hierarchy")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve infrastructure hierarchy")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNodeMetricsTimeSeries handles requests for node metrics over time
func (h *Handler) GetNodeMetricsTimeSeries(c *gin.Context) {
	nodeIDStr := c.Param("nodeId")
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_node_id", "Invalid node ID format")
		return
	}

	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetNodeMetricsTimeSeries(c.Request.Context(), nodeID, *req)
	if err != nil {
		log.Error().Err(err).Str("node_id", nodeIDStr).Msg("Failed to get node metrics time series")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve node metrics")
		return
	}

	c.JSON(http.StatusOK, response)
}

// parseCostAttributionRequest parses common request parameters
func (h *Handler) parseCostAttributionRequest(c *gin.Context) (*CostAttributionRequest, error) {
	req := &CostAttributionRequest{}

	// Parse start_date
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		// Default to 30 days ago
		req.StartDate = time.Now().AddDate(0, 0, -30)
	} else {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, err
		}
		req.StartDate = startDate
	}

	// Parse end_date
	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		// Default to today
		req.EndDate = time.Now()
	} else {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, err
		}
		req.EndDate = endDate
	}

	// Parse dimensions (comma-separated)
	dimensionsStr := c.Query("dimensions")
	if dimensionsStr != "" {
		req.Dimensions = parseCommaSeparated(dimensionsStr)
	}

	// Parse include_trend
	includeTrendStr := c.Query("include_trend")
	if includeTrendStr != "" {
		includeTrend, err := strconv.ParseBool(includeTrendStr)
		if err != nil {
			return nil, err
		}
		req.IncludeTrend = includeTrend
	}

	// Parse currency
	req.Currency = c.Query("currency")
	if req.Currency == "" {
		req.Currency = "USD"
	}

	return req, nil
}

// parseCommaSeparated parses a comma-separated string into a slice
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	
	var result []string
	for _, item := range splitAndTrim(s, ",") {
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		parts = append(parts, trimmed)
	}
	return parts
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// trimSpace trims whitespace from a string
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	// Trim leading whitespace
	for start < end && isSpace(s[start]) {
		start++
	}
	
	// Trim trailing whitespace
	for end > start && isSpace(s[end-1]) {
		end--
	}
	
	return s[start:end]
}

// isSpace checks if a character is whitespace
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// handleError handles API errors consistently
func (h *Handler) handleError(c *gin.Context, statusCode int, code, message string) {
	response := ErrorResponse{
		Error: message,
		Code:  code,
	}

	c.JSON(statusCode, response)
}

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.Info().
			Str("method", param.Method).
			Str("path", param.Path).
			Int("status", param.StatusCode).
			Dur("latency", param.Latency).
			Str("client_ip", param.ClientIP).
			Str("user_agent", param.Request.UserAgent()).
			Msg("HTTP request")
		return ""
	})
}

// GetAllocationReconciliation handles requests for allocation reconciliation/debug data
func (h *Handler) GetAllocationReconciliation(c *gin.Context) {
	req, err := h.parseCostAttributionRequest(c)
	if err != nil {
		h.handleError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	response, err := h.service.GetAllocationReconciliation(c.Request.Context(), *req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get allocation reconciliation")
		h.handleError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve reconciliation data")
		return
	}

	c.JSON(http.StatusOK, response)
}

// RecoveryMiddleware handles panics
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error().
			Interface("panic", recovered).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Msg("Panic recovered")

		response := ErrorResponse{
			Error: "Internal server error",
			Code:  "internal_error",
		}

		c.JSON(http.StatusInternalServerError, response)
	})
}
