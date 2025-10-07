// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package metrics

// OTLPMetrics tracks metrics for OTLP receivers
type OTLPMetrics struct {
	// Trace metrics
	TracesReceived         Counter `metric:"otlp_traces_received_total" tags:"component=otlp,signal=traces" help:"Total number of trace spans received"`
	TracesProcessed        Counter `metric:"otlp_traces_processed_total" tags:"component=otlp,signal=traces" help:"Total number of trace spans successfully processed"`
	TracesErrors           Counter `metric:"otlp_traces_errors_total" tags:"component=otlp,signal=traces" help:"Total number of trace processing errors"`
	TraceProcessDuration   Timer   `metric:"otlp_trace_process_duration_seconds" tags:"component=otlp,signal=traces" help:"Trace processing duration in seconds"`
	TraceBytes             Counter `metric:"otlp_trace_bytes_total" tags:"component=otlp,signal=traces" help:"Total bytes of trace data received"`

	// Metric metrics
	MetricsReceived        Counter `metric:"otlp_metrics_received_total" tags:"component=otlp,signal=metrics" help:"Total number of metric data points received"`
	MetricsProcessed       Counter `metric:"otlp_metrics_processed_total" tags:"component=otlp,signal=metrics" help:"Total number of metric data points successfully processed"`
	MetricsErrors          Counter `metric:"otlp_metrics_errors_total" tags:"component=otlp,signal=metrics" help:"Total number of metric processing errors"`
	MetricProcessDuration  Timer   `metric:"otlp_metric_process_duration_seconds" tags:"component=otlp,signal=metrics" help:"Metric processing duration in seconds"`
	MetricBytes            Counter `metric:"otlp_metric_bytes_total" tags:"component=otlp,signal=metrics" help:"Total bytes of metric data received"`

	// Log metrics
	LogsReceived           Counter `metric:"otlp_logs_received_total" tags:"component=otlp,signal=logs" help:"Total number of log records received"`
	LogsProcessed          Counter `metric:"otlp_logs_processed_total" tags:"component=otlp,signal=logs" help:"Total number of log records successfully processed"`
	LogsErrors             Counter `metric:"otlp_logs_errors_total" tags:"component=otlp,signal=logs" help:"Total number of log processing errors"`
	LogProcessDuration     Timer   `metric:"otlp_log_process_duration_seconds" tags:"component=otlp,signal=logs" help:"Log processing duration in seconds"`
	LogBytes               Counter `metric:"otlp_log_bytes_total" tags:"component=otlp,signal=logs" help:"Total bytes of log data received"`

	// gRPC receiver metrics
	GRPCRequestsTotal      Counter `metric:"otlp_grpc_requests_total" tags:"component=otlp,protocol=grpc" help:"Total number of gRPC requests received"`
	GRPCRequestErrors      Counter `metric:"otlp_grpc_request_errors_total" tags:"component=otlp,protocol=grpc" help:"Total number of gRPC request errors"`
	GRPCRequestDuration    Timer   `metric:"otlp_grpc_request_duration_seconds" tags:"component=otlp,protocol=grpc" help:"gRPC request duration in seconds"`
	GRPCActiveConnections  Gauge   `metric:"otlp_grpc_active_connections" tags:"component=otlp,protocol=grpc" help:"Current number of active gRPC connections"`

	// HTTP receiver metrics
	HTTPRequestsTotal      Counter `metric:"otlp_http_requests_total" tags:"component=otlp,protocol=http" help:"Total number of HTTP requests received"`
	HTTPRequestErrors      Counter `metric:"otlp_http_request_errors_total" tags:"component=otlp,protocol=http" help:"Total number of HTTP request errors"`
	HTTPRequestDuration    Timer   `metric:"otlp_http_request_duration_seconds" tags:"component=otlp,protocol=http" help:"HTTP request duration in seconds"`

	// Storage metrics
	StorageWriteLatency    Timer   `metric:"otlp_storage_write_duration_seconds" tags:"component=otlp" help:"Storage write duration in seconds"`
	StorageWriteErrors     Counter `metric:"otlp_storage_write_errors_total" tags:"component=otlp" help:"Total number of storage write errors"`
	StorageBatchSize       Histogram `metric:"otlp_storage_batch_size" tags:"component=otlp" help:"Distribution of storage batch sizes" buckets:"1,10,50,100,500,1000,5000"`

	// Parser metrics
	ParserErrors           Counter `metric:"otlp_parser_errors_total" tags:"component=otlp" help:"Total number of parsing errors"`
	ParserDuration         Timer   `metric:"otlp_parser_duration_seconds" tags:"component=otlp" help:"Parser processing duration in seconds"`
}

// NewOTLPMetrics creates and initializes OTLP metrics
func NewOTLPMetrics(factory Factory) *OTLPMetrics {
	metrics := &OTLPMetrics{}
	MustInit(metrics, factory, nil)
	return metrics
}
