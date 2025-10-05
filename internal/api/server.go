package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/api/handlers"
	"github.com/getlawrence/lawrence-oss/internal/storage"
)

// Server represents the HTTP API server
type Server struct {
	router     *gin.Engine
	storage    *storage.Container
	logger     *zap.Logger
	httpServer *http.Server
}

// NewServer creates a new API server
func NewServer(storage *storage.Container, logger *zap.Logger) *Server {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))
	
	server := &Server{
		router:  router,
		storage: storage,
		logger:  logger,
	}
	
	// Register routes
	server.registerRoutes()
	
	return server
}

// Start starts the HTTP server
func (s *Server) Start(port string) error {
	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: s.router,
	}
	
	s.logger.Info("Starting HTTP API server", zap.String("port", port))
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP API server")
	
	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	return s.httpServer.Shutdown(shutdownCtx)
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	// Initialize handlers
	agentHandlers := handlers.NewAgentHandlers(s.storage, s.logger)
	configHandlers := handlers.NewConfigHandlers(s.storage, s.logger)
	telemetryHandlers := handlers.NewTelemetryHandlers(s.storage, s.logger)
	groupHandlers := handlers.NewGroupHandlers(s.storage, s.logger)
	topologyHandlers := handlers.NewTopologyHandlers(s.storage, s.logger)
	healthHandlers := handlers.NewHealthHandlers(s.storage, s.logger)

	// Health check
	s.router.GET("/health", healthHandlers.HandleHealth)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Agent routes
		agents := v1.Group("/agents")
		{
			agents.GET("", agentHandlers.HandleGetAgents)
			agents.GET("/stats", agentHandlers.HandleGetAgentStats) // Must come before /:id
			agents.GET("/:id", agentHandlers.HandleGetAgent)
			agents.PATCH("/:id/group", agentHandlers.HandleUpdateAgentGroup)
		}

		// Config routes
		configs := v1.Group("/configs")
		{
			configs.GET("", configHandlers.HandleGetConfigs)
			configs.POST("", configHandlers.HandleCreateConfig)
			configs.POST("/validate", configHandlers.HandleValidateConfig) // Must come before /:id
			configs.GET("/versions", configHandlers.HandleGetConfigVersions)
			configs.GET("/:id", configHandlers.HandleGetConfig)
			configs.PUT("/:id", configHandlers.HandleUpdateConfig)
			configs.DELETE("/:id", configHandlers.HandleDeleteConfig)
		}

		// Telemetry routes
		telemetry := v1.Group("/telemetry")
		{
			telemetry.POST("/metrics/query", telemetryHandlers.HandleQueryMetrics)
			telemetry.POST("/logs/query", telemetryHandlers.HandleQueryLogs)
			telemetry.POST("/traces/query", telemetryHandlers.HandleQueryTraces)
			telemetry.GET("/overview", telemetryHandlers.HandleGetTelemetryOverview)
			telemetry.GET("/services", telemetryHandlers.HandleGetServices)
		}

		// Group routes
		groups := v1.Group("/groups")
		{
			groups.GET("", groupHandlers.HandleGetGroups)
			groups.POST("", groupHandlers.HandleCreateGroup)
			groups.GET("/:id", groupHandlers.HandleGetGroup)
			groups.PUT("/:id", groupHandlers.HandleUpdateGroup)
			groups.DELETE("/:id", groupHandlers.HandleDeleteGroup)
			groups.POST("/:id/config", groupHandlers.HandleAssignConfig)
			groups.GET("/:id/config", groupHandlers.HandleGetGroupConfig)
			groups.GET("/:id/agents", groupHandlers.HandleGetGroupAgents)
		}

		// Topology routes
		topology := v1.Group("/topology")
		{
			topology.GET("", topologyHandlers.HandleGetTopology)
			topology.GET("/agent/:id", topologyHandlers.HandleGetAgentTopology)
			topology.GET("/group/:id", topologyHandlers.HandleGetGroupTopology)
		}
	}

	// Serve static files for the UI
	s.router.Static("/assets", "./ui/dist/assets")

	// SPA catch-all route - must be last
	s.router.NoRoute(func(c *gin.Context) {
		// Check if file exists
		filePath := filepath.Join("./ui/dist", c.Request.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
			return
		}

		// Serve index.html for all other routes (SPA routing)
		c.File("./ui/dist/index.html")
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// loggingMiddleware adds request logging
func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
		)
		return ""
	})
}
