package receiver

import (
	"context"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/metrics"
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
	writer      TelemetryWriter
	asyncWriter AsyncTelemetryWriter
	parser      *parser.OTLPParser
	logger      *zap.Logger
	metrics     *metrics.OTLPMetrics
}

// MetricsService implements the OTLP Metrics Service gRPC interface
type MetricsService struct {
	colmetricspb.UnimplementedMetricsServiceServer
	writer      TelemetryWriter
	asyncWriter AsyncTelemetryWriter
	parser      *parser.OTLPParser
	logger      *zap.Logger
	metrics     *metrics.OTLPMetrics
}

// LogsService implements the OTLP Logs Service gRPC interface
type LogsService struct {
	collogspb.UnimplementedLogsServiceServer
	writer      TelemetryWriter
	asyncWriter AsyncTelemetryWriter
	parser      *parser.OTLPParser
	logger      *zap.Logger
	metrics     *metrics.OTLPMetrics
}

// NewTraceService creates a new TraceService instance
func NewTraceService(writer TelemetryWriter, asyncWriter AsyncTelemetryWriter, parser *parser.OTLPParser, metricsInstance *metrics.OTLPMetrics, logger *zap.Logger) *TraceService {
	return &TraceService{
		writer:      writer,
		asyncWriter: asyncWriter,
		parser:      parser,
		logger:      logger,
		metrics:     metricsInstance,
	}
}

// NewMetricsService creates a new MetricsService instance
func NewMetricsService(writer TelemetryWriter, asyncWriter AsyncTelemetryWriter, parser *parser.OTLPParser, metricsInstance *metrics.OTLPMetrics, logger *zap.Logger) *MetricsService {
	return &MetricsService{
		writer:      writer,
		asyncWriter: asyncWriter,
		parser:      parser,
		logger:      logger,
		metrics:     metricsInstance,
	}
}

// NewLogsService creates a new LogsService instance
func NewLogsService(writer TelemetryWriter, asyncWriter AsyncTelemetryWriter, parser *parser.OTLPParser, metricsInstance *metrics.OTLPMetrics, logger *zap.Logger) *LogsService {
	return &LogsService{
		writer:      writer,
		asyncWriter: asyncWriter,
		parser:      parser,
		logger:      logger,
		metrics:     metricsInstance,
	}
}

// Export handles trace export requests via gRPC
func (s *TraceService) Export(ctx context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	start := time.Now()
	s.logger.Debug("Processing gRPC trace export request",
		zap.Int("resource_spans_count", len(req.ResourceSpans)))

	// Track gRPC request
	if s.metrics != nil {
		s.metrics.GRPCRequestsTotal.Inc(1)
	}

	// Serialize the request to protobuf bytes
	data, err := proto.Marshal(req)
	if err != nil {
		s.logger.Error("Failed to marshal trace request", zap.Error(err))
		if s.metrics != nil {
			s.metrics.GRPCRequestErrors.Inc(1)
			s.metrics.TracesErrors.Inc(1)
		}
		return &coltracepb.ExportTraceServiceResponse{
			PartialSuccess: &coltracepb.ExportTracePartialSuccess{
				RejectedSpans: int64(len(req.ResourceSpans)),
				ErrorMessage:  "Failed to serialize request",
			},
		}, nil
	}

	// Track bytes received
	if s.metrics != nil {
		s.metrics.TraceBytes.Inc(int64(len(data)))
	}

	// Parse traces (agent ID will be extracted from service.instance.id)
	traces, err := s.parser.ParseTraces(data)
	if err != nil {
		s.logger.Error("Failed to parse trace data", zap.Error(err))
		if s.metrics != nil {
			s.metrics.ParserErrors.Inc(1)
			s.metrics.TracesErrors.Inc(1)
		}
		return &coltracepb.ExportTraceServiceResponse{
			PartialSuccess: &coltracepb.ExportTracePartialSuccess{
				RejectedSpans: int64(len(req.ResourceSpans)),
				ErrorMessage:  "Failed to parse trace data",
			},
		}, nil
	}

	// Track received and processed traces
	if s.metrics != nil {
		s.metrics.TracesReceived.Inc(int64(len(traces)))
	}

	// Create OTLP traces data for async writing
	otlpTracesData := &parser.OTLPTracesData{
		Traces: traces,
	}

	// Write to storage asynchronously
	writeStart := time.Now()
	if err := s.asyncWriter.WriteTracesAsync(otlpTracesData); err != nil {
		s.logger.Error("Failed to queue traces for async writing", zap.Error(err))
		if s.metrics != nil {
			s.metrics.StorageWriteErrors.Inc(1)
			s.metrics.TracesErrors.Inc(1)
		}
		return &coltracepb.ExportTraceServiceResponse{
			PartialSuccess: &coltracepb.ExportTracePartialSuccess{
				RejectedSpans: int64(len(req.ResourceSpans)),
				ErrorMessage:  "Failed to queue for storage",
			},
		}, nil
	}

	// Track storage write latency and processed traces
	if s.metrics != nil {
		s.metrics.StorageWriteLatency.Record(time.Since(writeStart))
		s.metrics.TracesProcessed.Inc(int64(len(traces)))
	}

	duration := time.Since(start)
	s.logger.Debug("Successfully processed trace export request",
		zap.Int("resource_spans_count", len(req.ResourceSpans)),
		zap.Int("traces_count", len(traces)),
		zap.Duration("duration", duration))

	// Track request duration
	if s.metrics != nil {
		s.metrics.GRPCRequestDuration.Record(duration)
		s.metrics.TraceProcessDuration.Record(duration)
	}

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

	// Parse metrics (agent ID will be extracted from service.instance.id)
	sums, gauges, histograms, err := s.parser.ParseMetrics(data)
	if err != nil {
		s.logger.Error("Failed to parse metrics data", zap.Error(err))
		return &colmetricspb.ExportMetricsServiceResponse{
			PartialSuccess: &colmetricspb.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(countMetricDataPoints(req.ResourceMetrics)),
				ErrorMessage:       "Failed to parse metrics data",
			},
		}, nil
	}

	// Create OTLP metrics data for async writing
	otlpMetricsData := &parser.OTLPMetricsData{
		Sums:       sums,
		Gauges:     gauges,
		Histograms: histograms,
	}

	// Write to storage asynchronously
	if err := s.asyncWriter.WriteMetricsAsync(otlpMetricsData); err != nil {
		s.logger.Error("Failed to queue metrics for async writing", zap.Error(err))
		return &colmetricspb.ExportMetricsServiceResponse{
			PartialSuccess: &colmetricspb.ExportMetricsPartialSuccess{
				RejectedDataPoints: int64(countMetricDataPoints(req.ResourceMetrics)),
				ErrorMessage:       "Failed to queue for storage",
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

	// Parse logs (agent ID will be extracted from service.instance.id)
	logs, err := s.parser.ParseLogs(data)
	if err != nil {
		s.logger.Error("Failed to parse logs data", zap.Error(err))
		return &collogspb.ExportLogsServiceResponse{
			PartialSuccess: &collogspb.ExportLogsPartialSuccess{
				RejectedLogRecords: int64(countLogRecords(req.ResourceLogs)),
				ErrorMessage:       "Failed to parse logs data",
			},
		}, nil
	}

	// Create OTLP logs data for async writing
	otlpLogsData := &parser.OTLPLogsData{
		Logs: logs,
	}

	// Write to storage asynchronously
	if err := s.asyncWriter.WriteLogsAsync(otlpLogsData); err != nil {
		s.logger.Error("Failed to queue logs for async writing", zap.Error(err))
		return &collogspb.ExportLogsServiceResponse{
			PartialSuccess: &collogspb.ExportLogsPartialSuccess{
				RejectedLogRecords: int64(countLogRecords(req.ResourceLogs)),
				ErrorMessage:       "Failed to queue for storage",
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
