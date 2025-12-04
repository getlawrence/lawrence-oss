package receiver

import (
	"context"
	"fmt"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/services"
	"go.uber.org/zap"
)

// TriggerAwareWriter wraps a TelemetryWriter and evaluates triggers after writing
type TriggerAwareWriter struct {
	writer              TelemetryWriter
	evaluator           *services.TelemetryTriggerEvaluator
	workflowService     services.WorkflowService
	logger              *zap.Logger
}

// NewTriggerAwareWriter creates a new trigger-aware writer wrapper
func NewTriggerAwareWriter(
	writer TelemetryWriter,
	evaluator *services.TelemetryTriggerEvaluator,
	workflowService services.WorkflowService,
	logger *zap.Logger,
) *TriggerAwareWriter {
	return &TriggerAwareWriter{
		writer:          writer,
		evaluator:       evaluator,
		workflowService: workflowService,
		logger:          logger,
	}
}

// WriteTraces writes traces and evaluates triggers
func (w *TriggerAwareWriter) WriteTraces(ctx context.Context, traces []otlp.TraceData) error {
	// Write traces first
	if err := w.writer.WriteTraces(ctx, traces); err != nil {
		return err
	}

	// Traces don't trigger workflows currently (as per plan)
	return nil
}

// WriteMetrics writes metrics and evaluates triggers
func (w *TriggerAwareWriter) WriteMetrics(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) error {
	// Write metrics first
	if err := w.writer.WriteMetrics(ctx, sums, gauges, histograms); err != nil {
		return err
	}

	// Evaluate triggers for each metric
	for _, sum := range sums {
		w.evaluateMetricTrigger(ctx, sum.MetricName, sum.Value, sum.AgentID, sum.ServiceName, sum.TimeUnix)
	}

	for _, gauge := range gauges {
		w.evaluateMetricTrigger(ctx, gauge.MetricName, gauge.Value, gauge.AgentID, gauge.ServiceName, gauge.TimeUnix)
	}

	// Histograms could be evaluated too, but for simplicity, we'll skip them for now
	// They could use sum or count values if needed

	return nil
}

// WriteLogs writes logs and evaluates triggers
func (w *TriggerAwareWriter) WriteLogs(ctx context.Context, logs []otlp.LogData) error {
	// Write logs first
	if err := w.writer.WriteLogs(ctx, logs); err != nil {
		return err
	}

	// Evaluate triggers for each log
	for _, log := range logs {
		w.evaluateLogTrigger(ctx, &log)
	}

	return nil
}

// evaluateLogTrigger evaluates a log against triggers and executes matching workflows
func (w *TriggerAwareWriter) evaluateLogTrigger(ctx context.Context, log *otlp.LogData) {
	matchingWorkflows := w.evaluator.EvaluateLog(ctx, log)
	
	for _, workflowID := range matchingWorkflows {
		w.executeWorkflow(ctx, workflowID, w.buildLogMetadata(log))
	}
}

// evaluateMetricTrigger evaluates a metric against triggers and executes matching workflows
func (w *TriggerAwareWriter) evaluateMetricTrigger(ctx context.Context, metricName string, value float64, agentID string, serviceName string, timestamp time.Time) {
	matchingWorkflows := w.evaluator.EvaluateMetric(ctx, metricName, value, agentID, serviceName, timestamp)
	
	for _, workflowID := range matchingWorkflows {
		w.executeWorkflow(ctx, workflowID, w.buildMetricMetadata(metricName, value, agentID, serviceName, timestamp))
	}
}

// executeWorkflow executes a workflow with the given metadata
func (w *TriggerAwareWriter) executeWorkflow(ctx context.Context, workflowID string, metadata map[string]string) {
	metadata["source"] = "telemetry"
	metadata["trigger_type"] = "telemetry"

	execution, err := w.workflowService.ExecuteWorkflow(ctx, workflowID, metadata)
	if err != nil {
		w.logger.Warn("Failed to execute telemetry-triggered workflow",
			zap.String("workflow_id", workflowID),
			zap.Error(err))
		return
	}

	w.logger.Info("Executed telemetry-triggered workflow",
		zap.String("workflow_id", workflowID),
		zap.String("execution_id", execution.ID),
		zap.String("status", string(execution.Status)))
}

// buildLogMetadata builds metadata map from log data
func (w *TriggerAwareWriter) buildLogMetadata(log *otlp.LogData) map[string]string {
	metadata := make(map[string]string)
	metadata["agent_id"] = log.AgentID
	metadata["group_id"] = log.GroupID
	metadata["group_name"] = log.GroupName
	metadata["service_name"] = log.ServiceName
	metadata["severity"] = log.SeverityText
	metadata["severity_number"] = fmt.Sprintf("%d", log.SeverityNumber)
	metadata["log_body"] = log.Body
	metadata["log_timestamp"] = log.Timestamp.Format("2006-01-02T15:04:05Z07:00")
	
	if log.TraceId != "" {
		metadata["trace_id"] = log.TraceId
	}
	if log.SpanId != "" {
		metadata["span_id"] = log.SpanId
	}

	return metadata
}

// buildMetricMetadata builds metadata map from metric data
func (w *TriggerAwareWriter) buildMetricMetadata(metricName string, value float64, agentID string, serviceName string, timestamp time.Time) map[string]string {
	metadata := make(map[string]string)
	metadata["metric_name"] = metricName
	metadata["metric_value"] = fmt.Sprintf("%f", value)
	metadata["agent_id"] = agentID
	metadata["service_name"] = serviceName
	metadata["metric_timestamp"] = timestamp.Format("2006-01-02T15:04:05Z07:00")

	return metadata
}

