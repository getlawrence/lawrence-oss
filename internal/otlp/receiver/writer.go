package receiver

import (
	"context"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/otlp/parser"
)

// TelemetryWriter defines the interface for writing telemetry data to storage
type TelemetryWriter interface {
	// WriteTraces writes trace data to storage
	WriteTraces(ctx context.Context, traces []otlp.TraceData) error

	// WriteMetrics writes metric data to storage
	WriteMetrics(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) error

	// WriteLogs writes log data to storage
	WriteLogs(ctx context.Context, logs []otlp.LogData) error
}

// AsyncTelemetryWriter defines the interface for asynchronous writing of telemetry data
type AsyncTelemetryWriter interface {
	// WriteTracesAsync writes trace data to storage asynchronously
	WriteTracesAsync(data *parser.OTLPTracesData) error

	// WriteMetricsAsync writes metric data to storage asynchronously
	WriteMetricsAsync(data *parser.OTLPMetricsData) error

	// WriteLogsAsync writes log data to storage asynchronously
	WriteLogsAsync(data *parser.OTLPLogsData) error
}
