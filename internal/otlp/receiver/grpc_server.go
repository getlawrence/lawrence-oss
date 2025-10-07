package receiver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/metrics"
	"github.com/getlawrence/lawrence-oss/internal/otlp/parser"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // Register gzip compressor
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

// GRPCServer represents the gRPC OTLP receiver server
type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	logger   *zap.Logger
	port     int
}

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(port int, writer TelemetryWriter, asyncWriter AsyncTelemetryWriter, metricsInstance *metrics.OTLPMetrics, logger *zap.Logger) (*GRPCServer, error) {
	// Create parser
	otlpParser := parser.NewOTLPParser(logger)

	// Create gRPC server with keepalive settings
	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    10 * time.Second,
			Timeout: 5 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// Register OTLP services
	traceService := NewTraceService(writer, asyncWriter, otlpParser, metricsInstance, logger)
	metricsService := NewMetricsService(writer, asyncWriter, otlpParser, metricsInstance, logger)
	logsService := NewLogsService(writer, asyncWriter, otlpParser, metricsInstance, logger)

	coltracepb.RegisterTraceServiceServer(server, traceService)
	colmetricspb.RegisterMetricsServiceServer(server, metricsService)
	collogspb.RegisterLogsServiceServer(server, logsService)

	// Enable gRPC reflection for debugging
	reflection.Register(server)

	return &GRPCServer{
		server: server,
		logger: logger,
		port:   port,
	}, nil
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	// Listen on the gRPC port
	address := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	s.listener = listener
	s.logger.Info("Starting gRPC OTLP receiver", zap.String("address", address))

	// Start serving
	go func() {
		if err := s.server.Serve(listener); err != nil {
			s.logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully stops the gRPC server
func (s *GRPCServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping gRPC OTLP receiver...")

	// Graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		s.logger.Warn("gRPC server shutdown timeout, forcing stop")
		s.server.Stop()
		return ctx.Err()
	}
}

// GetPort returns the port the server is listening on
func (s *GRPCServer) GetPort() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return fmt.Sprintf("%d", s.port)
}
