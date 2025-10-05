package duckdb

import (
	"context"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/otlp/receiver"
)

// WriterAdapter adapts the DuckDB storage to implement the TelemetryWriter interface
// This allows the OTLP receivers to write directly to DuckDB storage
type WriterAdapter struct {
	storage *Storage
}

// NewWriterAdapter creates a new writer adapter
func NewWriterAdapter(storage *Storage) receiver.TelemetryWriter {
	return &WriterAdapter{
		storage: storage,
	}
}

// WriteTraces writes trace data to DuckDB
func (w *WriterAdapter) WriteTraces(ctx context.Context, traces []otlp.TraceData) error {
	return w.storage.WriteTracesFromOTLP(ctx, traces)
}

// WriteMetrics writes metric data to DuckDB
func (w *WriterAdapter) WriteMetrics(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) error {
	return w.storage.WriteMetricsFromOTLP(ctx, sums, gauges, histograms)
}

// WriteLogs writes log data to DuckDB
func (w *WriterAdapter) WriteLogs(ctx context.Context, logs []otlp.LogData) error {
	return w.storage.WriteLogsFromOTLP(ctx, logs)
}
