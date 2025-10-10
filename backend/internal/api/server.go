package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pickeringtech/FinOpsAggregator/internal/store"
	"github.com/rs/zerolog/log"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// Server represents the HTTP API server
type Server struct {
	config  ServerConfig
	handler *Handler
	service *Service
	router  *gin.Engine
	server  *http.Server
}

// NewServer creates a new API server
func NewServer(config ServerConfig, store *store.Store) *Server {
	service := NewService(store)
	handler := NewHandler(service)
	router := SetupRouter(handler)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		config:  config,
		handler: handler,
		service: service,
		router:  router,
		server:  server,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Info().
		Str("host", s.config.Host).
		Int("port", s.config.Port).
		Msg("Starting HTTP API server")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping HTTP API server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	log.Info().Msg("HTTP API server stopped")
	return nil
}

// GetRouter returns the router for testing purposes
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
