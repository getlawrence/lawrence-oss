package receiver

import (
	"context"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/otlp/parser"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

// TraceService implements the OTLP Trace Service gRPC interface
type TraceService struct {
	coltracepb.UnimplementedTraceServiceServer
	writer TelemetryWriter
	parser *parser.OTLPParser
	logger *zap.Logger
}

// MetricsService implements the OTLP Metrics Service gRPC interface
type MetricsService struct {
	colmetricspb.UnimplementedMetricsServiceServer
	writer TelemetryWriter
	parser *parser.OTLPParser
	logger *zap.Logger
}

// LogsService implements the OTLP Logs Service gRPC interface
type LogsService struct {
	collogspb.UnimplementedLogsServiceServer
	writer TelemetryWriter
	parser *parser.OTLPParser
	logger *zap.Logger
}

// NewTraceService creates a new TraceService instance
func NewTraceService(writer TelemetryWriter, parser *parser.OTLPParser, logger *zap.Logger) *TraceService {
	return &TraceService{
		writer: writer,
		parser: parser,
		logger: logger,
	}
}

// NewMetricsService creates a new MetricsService instance
func NewMetricsService(writer TelemetryWriter, parser *parser.OTLPParser, logger *zap.Logger) *MetricsService {
	return &MetricsService{
		writer: writer,
		parser: parser,
		logger: logger,
	}
}

// NewLogsService creates a new LogsService instance
func NewLogsService(writer TelemetryWriter, parser *parser.OTLPParser, logger *zap.Logger) *LogsService {
	return &LogsService{
		writer: writer,
		parser: parser,
		logger: logger,
	}
}

// Export handles trace export requests via gRPC
func (s *TraceService) Export(ctx context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	start := time.Now()
	s.logger.Debug("Processing gRPC trace export request",
		zap.Int("resource_spans_count", len(req.ResourceSpans)))

	// Serialize the request to protobuf bytes
	data, err := proto.Marshal(req)
	if err != nil {
		s.logger.Error("Failed to marshal trace request", zap.Error(err))
		return &coltracepb.ExportTraceServiceResponse{
			PartialSuccess: &coltracepb.ExportTracePartialSuccess{
				RejectedSpans: int64(len(req.ResourceSpans)),
				ErrorMessage:  "Failed to serialize request",
			},
		}, nil
	}

	// Parse traces (OSS: use "default" as agent ID for now, will be enhanced later)
	traces, err := s.parser.ParseTraces(data, "default")
	if err != nil {
		s.logger.Error("Failed to parse trace data", zap.Error(err))
		return &coltracepb.ExportTraceServiceResponse{
			PartialSuccess: &coltracepb.ExportTracePartialSuccess{
				RejectedSpans: int64(len(req.ResourceSpans)),
				ErrorMessage:  "Failed to parse trace data",
			},
		}, nil
	}

	// Write to storage
	if err := s.writer.WriteTraces(ctx, traces); err != nil {
		s.logger.Error("Failed to write traces to storage", zap.Error(err))
		return &coltracepb.ExportTraceServiceResponse{
			PartialSuccess: &coltracepb.ExportTracePartialSuccess{
				RejectedSpans: int64(len(req.ResourceSpans)),
				ErrorMessage:  "Failed to write to storage",
			},
		}, nil
	}

	duration := time.Since(start)
	s.logger.Debug("Successfully processed trace export request",
		zap.Int("resource_spans_count", len(req.ResourceSpans)),
		zap.Int("traces_count", len(traces)),
		zap.Duration("duration", duration))

	return &coltracepb.ExportTraceServiceResponse{}, nil
}

// Export handles metrics export requests via gRPC
func (s *MetricsService) Export(ctx context.Context, req *colmetricspb.ExportMetricsServiceRequest) (*colmetricspb.ExportMetricsServiceResponse, error) {
	start := time.Now()
	s.logger.Debug("Processing gRPC metrics export request",
		zap.Int("resource_metrics_count", len(req.ResourceMetrics)))

	// Serialize the request to protobuf bytes
	data, err := proto.Marshal(req)
	if err != nil {
		s.logger.Error("Failed to marshal metrics request", zap.Error(err))
		return &colmetricspb.ExportMetricsServiceResponse{
			PartialSuccess: &colmetricspb.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(countMetricDataPoints(req.ResourceMetrics)),
				ErrorMessage:       "Failed to serialize request",
			},
		}, nil
	}

	// Parse metrics
	sums, gauges, histograms, err := s.parser.ParseMetrics(data, "default")
	if err != nil {
		s.logger.Error("Failed to parse metrics data", zap.Error(err))
		return &colmetricspb.ExportMetricsServiceResponse{
			PartialSuccess: &colmetricspb.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(countMetricDataPoints(req.ResourceMetrics)),
				ErrorMessage:       "Failed to parse metrics data",
			},
		}, nil
	}

	// Write to storage
	if err := s.writer.WriteMetrics(ctx, sums, gauges, histograms); err != nil {
		s.logger.Error("Failed to write metrics to storage", zap.Error(err))
		return &colmetricspb.ExportMetricsServiceResponse{
			PartialSuccess: &colmetricspb.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(countMetricDataPoints(req.ResourceMetrics)),
				ErrorMessage:       "Failed to write to storage",
			},
		}, nil
	}

	duration := time.Since(start)
	s.logger.Debug("Successfully processed metrics export request",
		zap.Int("resource_metrics_count", len(req.ResourceMetrics)),
		zap.Int("sum_count", len(sums)),
		zap.Int("gauge_count", len(gauges)),
		zap.Int("histogram_count", len(histograms)),
		zap.Duration("duration", duration))

	return &colmetricspb.ExportMetricsServiceResponse{}, nil
}

// Export handles logs export requests via gRPC
func (s *LogsService) Export(ctx context.Context, req *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	start := time.Now()
	s.logger.Debug("Processing gRPC logs export request",
		zap.Int("resource_logs_count", len(req.ResourceLogs)))

	// Serialize the request to protobuf bytes
	data, err := proto.Marshal(req)
	if err != nil {
		s.logger.Error("Failed to marshal logs request", zap.Error(err))
		return &collogspb.ExportLogsServiceResponse{
			PartialSuccess: &collogspb.ExportLogsPartialSuccess{
				RejectedLogRecords: int64(countLogRecords(req.ResourceLogs)),
				ErrorMessage:       "Failed to serialize request",
			},
		}, nil
	}

	// Parse logs
	logs, err := s.parser.ParseLogs(data, "default")
	if err != nil {
		s.logger.Error("Failed to parse logs data", zap.Error(err))
		return &collogspb.ExportLogsServiceResponse{
			PartialSuccess: &collogspb.ExportLogsPartialSuccess{
				RejectedLogRecords: int64(countLogRecords(req.ResourceLogs)),
				ErrorMessage:       "Failed to parse logs data",
			},
		}, nil
	}

	// Write to storage
	if err := s.writer.WriteLogs(ctx, logs); err != nil {
		s.logger.Error("Failed to write logs to storage", zap.Error(err))
		return &collogspb.ExportLogsServiceResponse{
			PartialSuccess: &collogspb.ExportLogsPartialSuccess{
				RejectedLogRecords: int64(countLogRecords(req.ResourceLogs)),
				ErrorMessage:       "Failed to write to storage",
			},
		}, nil
	}

	duration := time.Since(start)
	s.logger.Debug("Successfully processed logs export request",
		zap.Int("resource_logs_count", len(req.ResourceLogs)),
		zap.Int("logs_count", len(logs)),
		zap.Duration("duration", duration))

	return &collogspb.ExportLogsServiceResponse{}, nil
}

// countMetricDataPoints counts the total number of data points in resource metrics
func countMetricDataPoints(resourceMetrics []*metricspb.ResourceMetrics) int {
	count := 0
	for _, rm := range resourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				switch data := m.Data.(type) {
				case *metricspb.Metric_Gauge:
					count += len(data.Gauge.DataPoints)
				case *metricspb.Metric_Sum:
					count += len(data.Sum.DataPoints)
				case *metricspb.Metric_Histogram:
					count += len(data.Histogram.DataPoints)
				case *metricspb.Metric_ExponentialHistogram:
					count += len(data.ExponentialHistogram.DataPoints)
				case *metricspb.Metric_Summary:
					count += len(data.Summary.DataPoints)
				}
			}
		}
	}
	return count
}

// countLogRecords counts the total number of log records in resource logs
func countLogRecords(resourceLogs []*logspb.ResourceLogs) int {
	count := 0
	for _, rl := range resourceLogs {
		for _, sl := range rl.ScopeLogs {
			count += len(sl.LogRecords)
		}
	}
	return count
}
