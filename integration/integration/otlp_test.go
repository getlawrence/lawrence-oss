// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// TestOTLPGRPCTraces tests sending traces via gRPC
func TestOTLPGRPCTraces(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Connect to gRPC server
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", ts.OTLPGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := coltracepb.NewTraceServiceClient(conn)

	// Create test trace
	request := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "integration-test"},
							},
						},
						{
							Key: "service.instance.id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "test-agent"},
							},
						},
					},
				},
				ScopeSpans: []*tracepb.ScopeSpans{
					{
						Scope: &commonpb.InstrumentationScope{
							Name:    "test-scope",
							Version: "1.0.0",
						},
						Spans: []*tracepb.Span{
							{
								TraceId:           []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
								SpanId:            []byte{1, 2, 3, 4, 5, 6, 7, 8},
								Name:              "test-span",
								Kind:              tracepb.Span_SPAN_KIND_SERVER,
								StartTimeUnixNano: uint64(time.Now().UnixNano()),
								EndTimeUnixNano:   uint64(time.Now().Add(time.Second).UnixNano()),
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

	// Send traces
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Export(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOTLPHTTPTraces tests sending traces via HTTP
func TestOTLPHTTPTraces(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Create test trace
	request := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "integration-test-http"},
							},
						},
					},
				},
				ScopeSpans: []*tracepb.ScopeSpans{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-scope",
						},
						Spans: []*tracepb.Span{
							{
								TraceId:           []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
								SpanId:            []byte{1, 2, 3, 4, 5, 6, 7, 8},
								Name:              "http-test-span",
								StartTimeUnixNano: uint64(time.Now().UnixNano()),
								EndTimeUnixNano:   uint64(time.Now().Add(time.Second).UnixNano()),
							},
						},
					},
				},
			},
		},
	}

	// Marshal to protobuf
	data, err := proto.Marshal(request)
	require.NoError(t, err)

	// Send via HTTP POST
	url := fmt.Sprintf("http://localhost:%d/v1/traces", ts.OTLPHTTPPort)
	resp, err := http.Post(url, "application/x-protobuf", bytes.NewReader(data))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

// TestOTLPGRPCMetrics tests sending metrics via gRPC
func TestOTLPGRPCMetrics(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Connect to gRPC server
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", ts.OTLPGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := colmetricspb.NewMetricsServiceClient(conn)

	// Create test metrics
	now := uint64(time.Now().UnixNano())
	request := &colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "metric-test"},
							},
						},
					},
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-metrics",
						},
						Metrics: []*metricspb.Metric{
							{
								Name:        "test.counter",
								Description: "A test counter",
								Data: &metricspb.Metric_Sum{
									Sum: &metricspb.Sum{
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
										DataPoints: []*metricspb.NumberDataPoint{
											{
												TimeUnixNano: now,
												Value: &metricspb.NumberDataPoint_AsDouble{
													AsDouble: 42.0,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Send metrics
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Export(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOTLPHTTPMetrics tests sending metrics via HTTP
func TestOTLPHTTPMetrics(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	now := uint64(time.Now().UnixNano())
	request := &colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "http-metric-test"},
							},
						},
					},
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-metrics",
						},
						Metrics: []*metricspb.Metric{
							{
								Name: "test.gauge",
								Data: &metricspb.Metric_Gauge{
									Gauge: &metricspb.Gauge{
										DataPoints: []*metricspb.NumberDataPoint{
											{
												TimeUnixNano: now,
												Value: &metricspb.NumberDataPoint_AsDouble{
													AsDouble: 100.0,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := proto.Marshal(request)
	require.NoError(t, err)

	url := fmt.Sprintf("http://localhost:%d/v1/metrics", ts.OTLPHTTPPort)
	resp, err := http.Post(url, "application/x-protobuf", bytes.NewReader(data))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

// TestOTLPGRPCLogs tests sending logs via gRPC
func TestOTLPGRPCLogs(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", ts.OTLPGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := collogspb.NewLogsServiceClient(conn)

	now := uint64(time.Now().UnixNano())
	request := &collogspb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "log-test"},
							},
						},
					},
				},
				ScopeLogs: []*logspb.ScopeLogs{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-logs",
						},
						LogRecords: []*logspb.LogRecord{
							{
								TimeUnixNano:   now,
								SeverityNumber: logspb.SeverityNumber_SEVERITY_NUMBER_INFO,
								SeverityText:   "INFO",
								Body: &commonpb.AnyValue{
									Value: &commonpb.AnyValue_StringValue{StringValue: "Test log message"},
								},
							},
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Export(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOTLPHTTPLogs tests sending logs via HTTP
func TestOTLPHTTPLogs(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	now := uint64(time.Now().UnixNano())
	request := &collogspb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "http-log-test"},
							},
						},
					},
				},
				ScopeLogs: []*logspb.ScopeLogs{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-logs",
						},
						LogRecords: []*logspb.LogRecord{
							{
								TimeUnixNano: now,
								SeverityText: "ERROR",
								Body: &commonpb.AnyValue{
									Value: &commonpb.AnyValue_StringValue{StringValue: "Test error log"},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := proto.Marshal(request)
	require.NoError(t, err)

	url := fmt.Sprintf("http://localhost:%d/v1/logs", ts.OTLPHTTPPort)
	resp, err := http.Post(url, "application/x-protobuf", bytes.NewReader(data))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}
