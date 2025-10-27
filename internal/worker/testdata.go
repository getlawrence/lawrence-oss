package worker

import (
	"time"

	"google.golang.org/protobuf/proto"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// GenerateValidTraceData generates valid OTLP trace protobuf bytes
func GenerateValidTraceData() ([]byte, error) {
	now := time.Now()
	traceID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	parentSpanID := []byte{8, 7, 6, 5, 4, 3, 2, 1}

	request := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-service",
								},
							},
						},
						{
							Key: "service.instance.id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-agent-123",
								},
							},
						},
						{
							Key: "agent.group_id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "group-abc",
								},
							},
						},
						{
							Key: "agent.group_name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "TestGroup",
								},
							},
						},
					},
				},
				ScopeSpans: []*tracepb.ScopeSpans{
					{
						Scope: &commonpb.InstrumentationScope{
							Name:    "test-instrumentation",
							Version: "1.0.0",
						},
						Spans: []*tracepb.Span{
							{
								TraceId:           traceID,
								SpanId:            spanID,
								ParentSpanId:      parentSpanID,
								Name:              "test-span",
								Kind:              tracepb.Span_SPAN_KIND_SERVER,
								StartTimeUnixNano: uint64(now.Add(-time.Hour).UnixNano()),
								EndTimeUnixNano:   uint64(now.UnixNano()),
								Attributes: []*commonpb.KeyValue{
									{
										Key: "http.method",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{
												StringValue: "GET",
											},
										},
									},
									{
										Key: "http.status_code",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_IntValue{
												IntValue: 200,
											},
										},
									},
								},
								Status: &tracepb.Status{
									Code: tracepb.Status_STATUS_CODE_OK,
								},
							},
						},
					},
				},
			},
		},
	}

	return proto.Marshal(request)
}

// GenerateValidMetricsData generates valid OTLP metrics protobuf bytes with sum, gauge, and histogram
func GenerateValidMetricsData() ([]byte, error) {
	now := time.Now()
	startTime := now.Add(-time.Minute)

	request := &colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-service",
								},
							},
						},
						{
							Key: "service.instance.id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-agent-123",
								},
							},
						},
						{
							Key: "agent.group_id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "group-abc",
								},
							},
						},
						{
							Key: "agent.group_name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "TestGroup",
								},
							},
						},
					},
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name:    "test-instrumentation",
							Version: "1.0.0",
						},
						Metrics: []*metricspb.Metric{
							// Sum metric
							{
								Name:        "requests_total",
								Description: "Total requests",
								Unit:        "1",
								Data: &metricspb.Metric_Sum{
									Sum: &metricspb.Sum{
										DataPoints: []*metricspb.NumberDataPoint{
											{
												StartTimeUnixNano: uint64(startTime.UnixNano()),
												TimeUnixNano:      uint64(now.UnixNano()),
												Value: &metricspb.NumberDataPoint_AsInt{
													AsInt: 100,
												},
												Attributes: []*commonpb.KeyValue{
													{
														Key: "method",
														Value: &commonpb.AnyValue{
															Value: &commonpb.AnyValue_StringValue{
																StringValue: "GET",
															},
														},
													},
												},
											},
										},
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
										IsMonotonic:            true,
									},
								},
							},
							// Gauge metric
							{
								Name:        "cpu_usage",
								Description: "CPU usage percentage",
								Unit:        "1",
								Data: &metricspb.Metric_Gauge{
									Gauge: &metricspb.Gauge{
										DataPoints: []*metricspb.NumberDataPoint{
											{
												StartTimeUnixNano: uint64(startTime.UnixNano()),
												TimeUnixNano:      uint64(now.UnixNano()),
												Value: &metricspb.NumberDataPoint_AsDouble{
													AsDouble: 42.5,
												},
												Attributes: []*commonpb.KeyValue{
													{
														Key: "cpu",
														Value: &commonpb.AnyValue{
															Value: &commonpb.AnyValue_StringValue{
																StringValue: "0",
															},
														},
													},
												},
											},
										},
									},
								},
							},
							// Histogram metric
							{
								Name:        "request_duration",
								Description: "Request duration histogram",
								Unit:        "ms",
								Data: &metricspb.Metric_Histogram{
									Histogram: &metricspb.Histogram{
										DataPoints: []*metricspb.HistogramDataPoint{
											{
												StartTimeUnixNano: uint64(startTime.UnixNano()),
												TimeUnixNano:      uint64(now.UnixNano()),
												Count:             100,
												Sum:               proto.Float64(5000.0),
												BucketCounts:      []uint64{10, 20, 40, 30},
												ExplicitBounds:    []float64{10.0, 50.0, 100.0},
												Min:               proto.Float64(5.0),
												Max:               proto.Float64(200.0),
												Attributes: []*commonpb.KeyValue{
													{
														Key: "operation",
														Value: &commonpb.AnyValue{
															Value: &commonpb.AnyValue_StringValue{
																StringValue: "read",
															},
														},
													},
												},
											},
										},
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return proto.Marshal(request)
}

// GenerateValidLogsData generates valid OTLP logs protobuf bytes
func GenerateValidLogsData() ([]byte, error) {
	now := time.Now()
	traceID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	request := &collogspb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-service",
								},
							},
						},
						{
							Key: "service.instance.id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-agent-123",
								},
							},
						},
						{
							Key: "agent.group_id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "group-abc",
								},
							},
						},
						{
							Key: "agent.group_name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "TestGroup",
								},
							},
						},
					},
				},
				ScopeLogs: []*logspb.ScopeLogs{
					{
						Scope: &commonpb.InstrumentationScope{
							Name:    "test-instrumentation",
							Version: "1.0.0",
						},
						LogRecords: []*logspb.LogRecord{
							{
								TimeUnixNano:   uint64(now.UnixNano()),
								SeverityNumber: logspb.SeverityNumber_SEVERITY_NUMBER_INFO,
								SeverityText:   "INFO",
								Body: &commonpb.AnyValue{
									Value: &commonpb.AnyValue_StringValue{
										StringValue: "Test log message",
									},
								},
								Attributes: []*commonpb.KeyValue{
									{
										Key: "logger.name",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{
												StringValue: "com.example.Logger",
											},
										},
									},
									{
										Key: "event.name",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{
												StringValue: "test-event",
											},
										},
									},
								},
								TraceId: traceID,
								SpanId:  spanID,
								Flags:   uint32(1),
							},
						},
					},
				},
			},
		},
	}

	return proto.Marshal(request)
}

// GenerateLargeTraceData generates a large OTLP trace payload for stress testing
func GenerateLargeTraceData(count int) ([]byte, error) {
	traceID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	now := time.Now()

	spans := make([]*tracepb.Span, count)
	for i := 0; i < count; i++ {
		spanID := make([]byte, 8)
		spanID[0] = byte(i)
		spanID[1] = byte(i >> 8)

		spans[i] = &tracepb.Span{
			TraceId:           traceID,
			SpanId:            spanID,
			Name:              "large-test-span",
			Kind:              tracepb.Span_SPAN_KIND_SERVER,
			StartTimeUnixNano: uint64(now.Add(-time.Minute).UnixNano()),
			EndTimeUnixNano:   uint64(now.UnixNano()),
			Attributes: []*commonpb.KeyValue{
				{
					Key: "span.number",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_IntValue{
							IntValue: int64(i),
						},
					},
				},
			},
			Status: &tracepb.Status{
				Code: tracepb.Status_STATUS_CODE_OK,
			},
		}
	}

	request := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-service",
								},
							},
						},
						{
							Key: "service.instance.id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "test-agent-123",
								},
							},
						},
					},
				},
				ScopeSpans: []*tracepb.ScopeSpans{
					{
						Scope: &commonpb.InstrumentationScope{
							Name:    "test-instrumentation",
							Version: "1.0.0",
						},
						Spans: spans,
					},
				},
			},
		},
	}

	return proto.Marshal(request)
}

// GenerateInvalidData generates invalid protobuf data for error testing
func GenerateInvalidData() []byte {
	return []byte("invalid protobuf data")
}

// GenerateEmptyTraceData generates an empty trace export request
func GenerateEmptyTraceData() ([]byte, error) {
	request := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{},
	}
	return proto.Marshal(request)
}
