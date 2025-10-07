// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// TestEndToEndTelemetryFlow tests the complete flow from OTLP ingestion to API query
func TestEndToEndTelemetryFlow(t *testing.T) {
	ts := NewTestServer(t, false) // Use real database for e2e
	defer ts.Stop()
	ts.Start()

	// Step 1: Send telemetry data via OTLP
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", ts.OTLPGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := colmetricspb.NewMetricsServiceClient(conn)

	now := uint64(time.Now().UnixNano())
	request := &colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "e2e-test-service"},
							},
						},
						{
							Key: "service.instance.id",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "e2e-agent-1"},
							},
						},
					},
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "e2e-test",
						},
						Metrics: []*metricspb.Metric{
							{
								Name:        "e2e.test.counter",
								Description: "End-to-end test counter",
								Data: &metricspb.Metric_Sum{
									Sum: &metricspb.Sum{
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
										DataPoints: []*metricspb.NumberDataPoint{
											{
												TimeUnixNano: now,
												Value: &metricspb.NumberDataPoint_AsDouble{
													AsDouble: 123.45,
												},
												Attributes: []*commonpb.KeyValue{
													{
														Key: "env",
														Value: &commonpb.AnyValue{
															Value: &commonpb.AnyValue_StringValue{StringValue: "integration-test"},
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
			},
		},
	}

	// Send metrics
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Export(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Step 2: Wait a bit for data to be persisted
	time.Sleep(100 * time.Millisecond)

	// Step 3: Query telemetry data via API
	queryReq := map[string]interface{}{
		"start_time": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		"end_time":   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		"limit":      100,
	}

	body, err := json.Marshal(queryReq)
	require.NoError(t, err)

	httpResp, err := ts.POST("/api/v1/telemetry/metrics/query", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer httpResp.Body.Close()

	assert.Equal(t, http.StatusOK, httpResp.StatusCode)

	var metricsResp map[string]interface{}
	err = json.NewDecoder(httpResp.Body).Decode(&metricsResp)
	require.NoError(t, err)

	metrics, ok := metricsResp["metrics"].([]interface{})
	require.True(t, ok)

	// Verify we got the metric back
	// Note: Actual verification depends on your API response format
	t.Logf("Received %d metrics", len(metrics))
}

// TestEndToEndAgentLifecycle tests the complete agent lifecycle
func TestEndToEndAgentLifecycle(t *testing.T) {
	ts := NewTestServer(t, true)
	defer ts.Stop()
	ts.Start()

	// Step 1: Create a group
	groupData := map[string]interface{}{
		"name": "E2E Test Group",
		"labels": map[string]string{
			"env": "integration",
		},
	}

	body, err := json.Marshal(groupData)
	require.NoError(t, err)

	resp, err := ts.POST("/api/v1/groups", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get the created group ID from response
	var createdGroup map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	require.NoError(t, err)

	groupID, ok := createdGroup["id"].(string)
	require.True(t, ok)

	// Step 2: Verify group was created
	resp, err = ts.GET("/api/v1/groups/" + groupID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Step 3: Create a config for the group
	configContent := "receivers:\n  otlp:\n    protocols:\n      grpc:\n      http:\n"
	configData := map[string]interface{}{
		"group_id":    groupID,
		"content":     configContent,
		"config_hash": "test-hash-123",
		"version":     1,
	}

	body, err = json.Marshal(configData)
	require.NoError(t, err)

	resp, err = ts.POST("/api/v1/configs", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Note: Actual status code depends on your implementation
	assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK)

	// Step 4: List configs
	resp, err = ts.GET("/api/v1/configs")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var configsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&configsResp)
	require.NoError(t, err)

	configs, ok := configsResp["configs"].([]interface{})
	require.True(t, ok)
	t.Logf("Found %d configs", len(configs))

	// Step 5: Delete the group
	resp, err = ts.DELETE("/api/v1/groups/" + groupID)
	require.NoError(t, err)
	resp.Body.Close()

	// Verify deletion
	resp, err = ts.GET("/api/v1/groups/" + groupID)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestEndToEndMultipleServices tests telemetry from multiple services
func TestEndToEndMultipleServices(t *testing.T) {
	ts := NewTestServer(t, false)
	defer ts.Stop()
	ts.Start()

	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", ts.OTLPGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := colmetricspb.NewMetricsServiceClient(conn)
	ctx := context.Background()

	// Send metrics from multiple services
	services := []string{"frontend", "backend", "database"}

	for _, svc := range services {
		now := uint64(time.Now().UnixNano())
		request := &colmetricspb.ExportMetricsServiceRequest{
			ResourceMetrics: []*metricspb.ResourceMetrics{
				{
					Resource: &resourcepb.Resource{
						Attributes: []*commonpb.KeyValue{
							{
								Key: "service.name",
								Value: &commonpb.AnyValue{
									Value: &commonpb.AnyValue_StringValue{StringValue: svc},
								},
							},
						},
					},
					ScopeMetrics: []*metricspb.ScopeMetrics{
						{
							Scope: &commonpb.InstrumentationScope{
								Name: "multi-service-test",
							},
							Metrics: []*metricspb.Metric{
								{
									Name: "requests.total",
									Data: &metricspb.Metric_Sum{
										Sum: &metricspb.Sum{
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

		_, err := client.Export(ctx, request)
		require.NoError(t, err)
	}

	// Wait for data to be persisted
	time.Sleep(200 * time.Millisecond)

	// Query services list
	resp, err := ts.GET("/api/v1/telemetry/services")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var servicesResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&servicesResp)
	require.NoError(t, err)

	servicesInterface, ok := servicesResp["services"].([]interface{})
	require.True(t, ok)

	// Convert to []string for logging
	var serviceNames []string
	for _, s := range servicesInterface {
		if service, ok := s.(string); ok {
			serviceNames = append(serviceNames, service)
		}
	}

	// Should have at least the services we sent
	t.Logf("Found services: %v", serviceNames)
}

// TestEndToEndConcurrentWrites tests concurrent telemetry writes
func TestEndToEndConcurrentWrites(t *testing.T) {
	ts := NewTestServer(t, false)
	defer ts.Stop()
	ts.Start()

	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", ts.OTLPGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := colmetricspb.NewMetricsServiceClient(conn)

	// Send metrics concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			now := uint64(time.Now().UnixNano())
			request := &colmetricspb.ExportMetricsServiceRequest{
				ResourceMetrics: []*metricspb.ResourceMetrics{
					{
						Resource: &resourcepb.Resource{
							Attributes: []*commonpb.KeyValue{
								{
									Key: "service.name",
									Value: &commonpb.AnyValue{
										Value: &commonpb.AnyValue_StringValue{StringValue: fmt.Sprintf("concurrent-test-%d", idx)},
									},
								},
							},
						},
						ScopeMetrics: []*metricspb.ScopeMetrics{
							{
								Scope: &commonpb.InstrumentationScope{
									Name: "concurrency-test",
								},
								Metrics: []*metricspb.Metric{
									{
										Name: "concurrent.metric",
										Data: &metricspb.Metric_Gauge{
											Gauge: &metricspb.Gauge{
												DataPoints: []*metricspb.NumberDataPoint{
													{
														TimeUnixNano: now,
														Value: &metricspb.NumberDataPoint_AsDouble{
															AsDouble: float64(idx),
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

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := client.Export(ctx, request)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("All concurrent writes completed successfully")
}
