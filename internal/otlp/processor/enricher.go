package processor

import (
	"context"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Enricher enriches telemetry data with group information based on agent lookups
type Enricher struct {
	agentService services.AgentService
	logger       *zap.Logger
}

// NewEnricher creates a new telemetry enricher
func NewEnricher(agentService services.AgentService, logger *zap.Logger) *Enricher {
	return &Enricher{
		agentService: agentService,
		logger:       logger,
	}
}

// EnrichTraces enriches traces with group information
func (e *Enricher) EnrichTraces(ctx context.Context, traces []otlp.TraceData) {
	for i := range traces {
		e.enrichTelemetry(ctx, &traces[i].AgentID, &traces[i].GroupID, &traces[i].GroupName)
	}
}

// EnrichMetrics enriches metrics with group information
func (e *Enricher) EnrichMetrics(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) {
	// Enrich sums
	for i := range sums {
		e.enrichTelemetry(ctx, &sums[i].AgentID, &sums[i].GroupID, &sums[i].GroupName)
	}

	// Enrich gauges
	for i := range gauges {
		e.enrichTelemetry(ctx, &gauges[i].AgentID, &gauges[i].GroupID, &gauges[i].GroupName)
	}

	// Enrich histograms
	for i := range histograms {
		e.enrichTelemetry(ctx, &histograms[i].AgentID, &histograms[i].GroupID, &histograms[i].GroupName)
	}
}

// EnrichLogs enriches logs with group information
func (e *Enricher) EnrichLogs(ctx context.Context, logs []otlp.LogData) {
	for i := range logs {
		e.enrichTelemetry(ctx, &logs[i].AgentID, &logs[i].GroupID, &logs[i].GroupName)
	}
}

// enrichTelemetry looks up the agent's group information and populates group_id and group_name
func (e *Enricher) enrichTelemetry(ctx context.Context, agentID *string, groupID *string, groupName *string) {
	// Skip if agentID is empty
	if agentID == nil || *agentID == "" || *agentID == "default" {
		return
	}

	// Parse agentID as UUID
	agentUUID, err := uuid.Parse(*agentID)
	if err != nil {
		e.logger.Debug("Failed to parse agent ID for enrichment",
			zap.String("agentID", *agentID),
			zap.Error(err))
		return
	}

	// Look up agent from service
	agent, err := e.agentService.GetAgent(ctx, agentUUID)
	if err != nil || agent == nil {
		// Agent not found - this can happen for telemetry sent before agent registers
		// This is debug level because it's expected in some scenarios
		e.logger.Debug("Agent not found for telemetry enrichment",
			zap.String("agentID", *agentID),
			zap.Error(err))
		return
	}

	// Populate group information if agent has a group
	if agent.GroupID != nil && *agent.GroupID != "" {
		*groupID = *agent.GroupID
	}

	if agent.GroupName != nil && *agent.GroupName != "" {
		*groupName = *agent.GroupName
	}
}
