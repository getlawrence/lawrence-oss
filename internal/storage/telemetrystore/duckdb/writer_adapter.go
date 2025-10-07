package duckdb

import (
	"context"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/otlp/parser"
)

// WriterAdapter adapts the DuckDB storage to implement both TelemetryWriter and AsyncTelemetryWriter interfaces
// This allows the OTLP receivers to write directly to telemetry storage
type WriterAdapter struct {
	storage *Storage
}

// NewWriterAdapter creates a new writer adapter
func NewWriterAdapter(storage *Storage) *WriterAdapter {
	return &WriterAdapter{
		storage: storage,
	}
}

// Sync methods (TelemetryWriter interface)

// WriteTraces writes trace data to storage synchronously
func (w *WriterAdapter) WriteTraces(ctx context.Context, traces []otlp.TraceData) error {
	return w.storage.WriteTracesFromOTLP(ctx, traces)
}

// WriteMetrics writes metric data to storage synchronously
func (w *WriterAdapter) WriteMetrics(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) error {
	return w.storage.WriteMetricsFromOTLP(ctx, sums, gauges, histograms)
}

// WriteLogs writes log data to storage synchronously
func (w *WriterAdapter) WriteLogs(ctx context.Context, logs []otlp.LogData) error {
	return w.storage.WriteLogsFromOTLP(ctx, logs)
}

// Async methods (AsyncTelemetryWriter interface)

// WriteTracesAsync writes trace data to storage asynchronously
func (w *WriterAdapter) WriteTracesAsync(data *parser.OTLPTracesData) error {
	return w.storage.WriteTracesFromOTLP(context.Background(), data.Traces)
}

// WriteMetricsAsync writes metric data to storage asynchronously
func (w *WriterAdapter) WriteMetricsAsync(data *parser.OTLPMetricsData) error {
	return w.storage.WriteMetricsFromOTLP(context.Background(), data.Sums, data.Gauges, data.Histograms)
}

// WriteLogsAsync writes log data to storage asynchronously
func (w *WriterAdapter) WriteLogsAsync(data *parser.OTLPLogsData) error {
	return w.storage.WriteLogsFromOTLP(context.Background(), data.Logs)
}
