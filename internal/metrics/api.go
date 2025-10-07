// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package metrics

// APIMetrics tracks metrics for the REST API server
type APIMetrics struct {
	// Request metrics
	RequestCount       Counter `metric:"api_requests_total" tags:"component=api" help:"Total number of API requests"`
	RequestErrors      Counter `metric:"api_request_errors_total" tags:"component=api" help:"Total number of API request errors"`
	RequestDuration    Timer   `metric:"api_request_duration_seconds" tags:"component=api" help:"API request duration in seconds"`

	// Agent endpoint metrics
	AgentGetCount      Counter `metric:"api_agent_get_total" tags:"component=api,endpoint=get_agent" help:"Total agent GET requests"`
	AgentListCount     Counter `metric:"api_agent_list_total" tags:"component=api,endpoint=list_agents" help:"Total agent LIST requests"`
	AgentCreateCount   Counter `metric:"api_agent_create_total" tags:"component=api,endpoint=create_agent" help:"Total agent CREATE requests"`
	AgentUpdateCount   Counter `metric:"api_agent_update_total" tags:"component=api,endpoint=update_agent" help:"Total agent UPDATE requests"`
	AgentDeleteCount   Counter `metric:"api_agent_delete_total" tags:"component=api,endpoint=delete_agent" help:"Total agent DELETE requests"`

	// Group endpoint metrics
	GroupGetCount      Counter `metric:"api_group_get_total" tags:"component=api,endpoint=get_group" help:"Total group GET requests"`
	GroupListCount     Counter `metric:"api_group_list_total" tags:"component=api,endpoint=list_groups" help:"Total group LIST requests"`
	GroupCreateCount   Counter `metric:"api_group_create_total" tags:"component=api,endpoint=create_group" help:"Total group CREATE requests"`
	GroupUpdateCount   Counter `metric:"api_group_update_total" tags:"component=api,endpoint=update_group" help:"Total group UPDATE requests"`
	GroupDeleteCount   Counter `metric:"api_group_delete_total" tags:"component=api,endpoint=delete_group" help:"Total group DELETE requests"`

	// Config endpoint metrics
	ConfigGetCount     Counter `metric:"api_config_get_total" tags:"component=api,endpoint=get_config" help:"Total config GET requests"`
	ConfigListCount    Counter `metric:"api_config_list_total" tags:"component=api,endpoint=list_configs" help:"Total config LIST requests"`
	ConfigCreateCount  Counter `metric:"api_config_create_total" tags:"component=api,endpoint=create_config" help:"Total config CREATE requests"`
	ConfigUpdateCount  Counter `metric:"api_config_update_total" tags:"component=api,endpoint=update_config" help:"Total config UPDATE requests"`
	ConfigDeleteCount  Counter `metric:"api_config_delete_total" tags:"component=api,endpoint=delete_config" help:"Total config DELETE requests"`

	// Telemetry query metrics
	TelemetryQueryCount    Counter `metric:"api_telemetry_query_total" tags:"component=api,endpoint=query_telemetry" help:"Total telemetry query requests"`
	TelemetryQueryDuration Timer   `metric:"api_telemetry_query_duration_seconds" tags:"component=api" help:"Telemetry query duration in seconds"`

	// Topology metrics
	TopologyQueryCount     Counter `metric:"api_topology_query_total" tags:"component=api,endpoint=query_topology" help:"Total topology query requests"`
	TopologyQueryDuration  Timer   `metric:"api_topology_query_duration_seconds" tags:"component=api" help:"Topology query duration in seconds"`

	// Health check metrics
	HealthCheckCount   Counter `metric:"api_health_check_total" tags:"component=api,endpoint=health" help:"Total health check requests"`
}

// NewAPIMetrics creates and initializes API metrics
func NewAPIMetrics(factory Factory) *APIMetrics {
	metrics := &APIMetrics{}
	MustInit(metrics, factory, nil)
	return metrics
}
