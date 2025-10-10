package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(handler *Handler) *gin.Engine {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(CORSMiddleware())
	router.Use(LoggingMiddleware())
	router.Use(RecoveryMiddleware())

	// Health check endpoint
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Product hierarchy endpoints
		products := v1.Group("/products")
		{
			products.GET("/hierarchy", handler.GetProductHierarchy)
			products.GET("", handler.ListProducts) // New: flat list of products
		}

		// Individual node endpoints
		nodes := v1.Group("/nodes")
		{
			nodes.GET("/:nodeId", handler.GetIndividualNode)
			nodes.GET("", handler.ListNodes) // New: flat list of all nodes
		}

		// Cost aggregation endpoints
		costs := v1.Group("/costs")
		{
			costs.GET("/by-type", handler.GetCostsByType)       // New: aggregate by node type
			costs.GET("/by-dimension", handler.GetCostsByDimension) // New: aggregate by any dimension
		}

		// Platform and shared services endpoints
		platform := v1.Group("/platform")
		{
			platform.GET("/services", handler.GetPlatformServices)
		}
	}

	return router
}
