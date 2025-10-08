package receiver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/metrics"
	"github.com/getlawrence/lawrence-oss/internal/otlp/parser"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

// HTTPServer represents the HTTP OTLP receiver server
type HTTPServer struct {
	server  *http.Server
	logger  *zap.Logger
	writer  TelemetryWriter
	parser  *parser.OTLPParser
	metrics *metrics.OTLPMetrics
	port    int
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(port int, writer TelemetryWriter, metricsInstance *metrics.OTLPMetrics, logger *zap.Logger) (*HTTPServer, error) {
	// Set Gin to release mode for better performance
	gin.SetMode(gin.ReleaseMode)

	// Create parser
	otlpParser := parser.NewOTLPParser(logger)

	// Create HTTP server
	s := &HTTPServer{
		logger:  logger,
		writer:  writer,
		parser:  otlpParser,
		metrics: metricsInstance,
		port:    port,
	}

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())

	// Setup routes
	s.setupRoutes(router)

	// Create HTTP server
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s, nil
}

// setupRoutes configures the HTTP server with routes
func (s *HTTPServer) setupRoutes(router *gin.Engine) {
	// Health check endpoints
	router.GET("/health", s.healthCheck)
	router.GET("/ready", s.readyCheck)

	// Standard OTLP HTTP endpoints
	router.POST("/v1/traces", s.handleOTLPTraces)
	router.POST("/v1/metrics", s.handleOTLPMetrics)
	router.POST("/v1/logs", s.handleOTLPLogs)

	s.logger.Info("OTLP HTTP routes registered")
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP OTLP receiver", zap.Int("port", s.port))

	// Start serving
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP OTLP receiver...")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("HTTP server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("HTTP server stopped gracefully")
	return nil
}

// handleOTLPTraces handles OTLP traces ingestion
func (s *HTTPServer) handleOTLPTraces(c *gin.Context) {
	start := time.Now()

	// Read raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.logger.Error("Failed to read traces request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Unmarshal to validate it's valid OTLP
	var req coltracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		s.logger.Error("Failed to unmarshal traces request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTLP traces data"})
		return
	}

	// Parse traces (agent ID will be extracted from service.instance.id)
	traces, err := s.parser.ParseTraces(body)
	if err != nil {
		s.logger.Error("Failed to parse trace data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse trace data"})
		return
	}

	// Create OTLP traces data for async writing
	otlpTracesData := &parser.OTLPTracesData{
		Traces: traces,
	}

	ctx := c.Request.Context()
	// Write to storage asynchronously
	if err := s.writer.WriteTraces(ctx, otlpTracesData.Traces); err != nil {
		s.logger.Error("Failed to write traces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write traces"})
		return
	}

	duration := time.Since(start)
	s.logger.Debug("Successfully processed traces request",
		zap.Int("body_size", len(body)),
		zap.Int("traces_count", len(traces)),
		zap.Duration("duration", duration))

	c.Status(http.StatusAccepted)
}

// handleOTLPMetrics handles OTLP metrics ingestion
func (s *HTTPServer) handleOTLPMetrics(c *gin.Context) {
	start := time.Now()

	// Read raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.logger.Error("Failed to read metrics request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Unmarshal to validate it's valid OTLP
	var req colmetricspb.ExportMetricsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		s.logger.Error("Failed to unmarshal metrics request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTLP metrics data"})
		return
	}

	// Parse metrics (agent ID will be extracted from service.instance.id)
	sums, gauges, histograms, err := s.parser.ParseMetrics(body)
	if err != nil {
		s.logger.Error("Failed to parse metrics data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse metrics data"})
		return
	}

	// Create OTLP metrics data for async writing
	otlpMetricsData := &parser.OTLPMetricsData{
		Sums:       sums,
		Gauges:     gauges,
		Histograms: histograms,
	}

	ctx := c.Request.Context()
	// Write to storage asynchronously
	if err := s.writer.WriteMetrics(ctx, otlpMetricsData.Sums, otlpMetricsData.Gauges, otlpMetricsData.Histograms); err != nil {
		s.logger.Error("Failed to write metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write metrics"})
		return
	}

	duration := time.Since(start)
	s.logger.Debug("Successfully processed metrics request",
		zap.Int("body_size", len(body)),
		zap.Int("sum_count", len(sums)),
		zap.Int("gauge_count", len(gauges)),
		zap.Int("histogram_count", len(histograms)),
		zap.Duration("duration", duration))

	c.Status(http.StatusAccepted)
}

// handleOTLPLogs handles OTLP logs ingestion
func (s *HTTPServer) handleOTLPLogs(c *gin.Context) {
	start := time.Now()

	// Read raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.logger.Error("Failed to read logs request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Unmarshal to validate it's valid OTLP
	var req collogspb.ExportLogsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		s.logger.Error("Failed to unmarshal logs request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTLP logs data"})
		return
	}

	// Parse logs (agent ID will be extracted from service.instance.id)
	logs, err := s.parser.ParseLogs(body)
	if err != nil {
		s.logger.Error("Failed to parse logs data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse logs data"})
		return
	}

	// Create OTLP logs data for async writing
	otlpLogsData := &parser.OTLPLogsData{
		Logs: logs,
	}

	ctx := c.Request.Context()
	// Write to storage asynchronously
	if err := s.writer.WriteLogs(ctx, otlpLogsData.Logs); err != nil {
		s.logger.Error("Failed to write logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write logs"})
		return
	}

	duration := time.Since(start)
	s.logger.Debug("Successfully processed logs request",
		zap.Int("body_size", len(body)),
		zap.Int("logs_count", len(logs)),
		zap.Duration("duration", duration))

	c.Status(http.StatusAccepted)
}

// healthCheck returns server health status
func (s *HTTPServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyCheck returns server readiness status
func (s *HTTPServer) readyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// corsMiddleware adds CORS headers
func (s *HTTPServer) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
