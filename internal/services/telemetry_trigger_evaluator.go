package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"go.uber.org/zap"
)

// TelemetryTriggerEvaluator evaluates telemetry data against active triggers
type TelemetryTriggerEvaluator struct {
	appStore     types.ApplicationStore
	logger       *zap.Logger
	mu           sync.RWMutex
	triggers     map[string]*types.WorkflowTrigger // workflowID -> trigger
	metricCache  map[string]*metricWindow          // key -> sliding window
	cacheCleanup *time.Ticker
}

// metricWindow represents a sliding window of metric values
type metricWindow struct {
	values []metricValue
	window time.Duration
	mu     sync.RWMutex
}

type metricValue struct {
	timestamp time.Time
	value     float64
}

// NewTelemetryTriggerEvaluator creates a new telemetry trigger evaluator
func NewTelemetryTriggerEvaluator(appStore types.ApplicationStore, logger *zap.Logger) *TelemetryTriggerEvaluator {
	evaluator := &TelemetryTriggerEvaluator{
		appStore:    appStore,
		logger:      logger,
		triggers:    make(map[string]*types.WorkflowTrigger),
		metricCache: make(map[string]*metricWindow),
	}

	// Start cache cleanup ticker (runs every minute)
	evaluator.cacheCleanup = time.NewTicker(1 * time.Minute)
	go evaluator.cleanupCache()

	return evaluator
}

// LoadActiveTriggers loads all active telemetry triggers from the store
func (e *TelemetryTriggerEvaluator) LoadActiveTriggers(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get all workflows with telemetry trigger type
	telemetryType := types.WorkflowTriggerTypeTelemetry
	activeStatus := types.WorkflowStatusActive

	workflows, err := e.appStore.ListWorkflows(ctx, types.WorkflowFilter{
		Type:   &telemetryType,
		Status: &activeStatus,
	})

	if err != nil {
		return fmt.Errorf("failed to list telemetry triggers: %w", err)
	}

	// Clear existing triggers
	e.triggers = make(map[string]*types.WorkflowTrigger)

	// Load triggers for each workflow
	for _, workflow := range workflows {
		trigger, err := e.appStore.GetWorkflowTrigger(ctx, workflow.ID)
		if err != nil {
			e.logger.Warn("Failed to get trigger for workflow",
				zap.String("workflow_id", workflow.ID),
				zap.Error(err))
			continue
		}

		if trigger != nil && trigger.Enabled && trigger.TelemetryConfig != nil {
			e.triggers[workflow.ID] = trigger
		}
	}

	e.logger.Info("Loaded active telemetry triggers",
		zap.Int("count", len(e.triggers)))

	return nil
}

// EvaluateLog evaluates a log entry against all log triggers
func (e *TelemetryTriggerEvaluator) EvaluateLog(ctx context.Context, log *otlp.LogData) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var matchingWorkflows []string

	for workflowID, trigger := range e.triggers {
		if trigger.TelemetryConfig == nil || trigger.TelemetryConfig.Type != types.TelemetryTriggerTypeLog {
			continue
		}

		if e.matchesLogTrigger(log, trigger.TelemetryConfig) {
			matchingWorkflows = append(matchingWorkflows, workflowID)
		}
	}

	return matchingWorkflows
}

// EvaluateMetric evaluates a metric against all metric triggers
func (e *TelemetryTriggerEvaluator) EvaluateMetric(ctx context.Context, metricName string, value float64, agentID string, serviceName string, timestamp time.Time) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var matchingWorkflows []string

	for workflowID, trigger := range e.triggers {
		if trigger.TelemetryConfig == nil || trigger.TelemetryConfig.Type != types.TelemetryTriggerTypeMetric {
			continue
		}

		config := trigger.TelemetryConfig

		// Check if metric name matches
		if config.MetricName != "" && config.MetricName != metricName {
			continue
		}

		// Check agent/service filters
		if config.AgentID != "" && config.AgentID != agentID {
			continue
		}
		if config.ServiceName != "" && config.ServiceName != serviceName {
			continue
		}

		// Add metric to cache and check threshold
		if e.checkMetricThreshold(workflowID, config, value, timestamp) {
			matchingWorkflows = append(matchingWorkflows, workflowID)
		}
	}

	return matchingWorkflows
}

// matchesLogTrigger checks if a log matches a log trigger configuration
func (e *TelemetryTriggerEvaluator) matchesLogTrigger(log *otlp.LogData, config *types.TelemetryTriggerConfig) bool {
	// Check agent filter
	if config.AgentID != "" && config.AgentID != log.AgentID {
		return false
	}

	// Check service filter
	if config.ServiceName != "" && config.ServiceName != log.ServiceName {
		return false
	}

	// Check severity filter
	if config.Severity != "" {
		severityMatch := false
		switch strings.ToLower(config.Severity) {
		case "error":
			severityMatch = log.SeverityNumber >= 17 // ERROR level in OTLP
		case "warn", "warning":
			severityMatch = log.SeverityNumber >= 13 && log.SeverityNumber < 17 // WARN level
		case "info":
			severityMatch = log.SeverityNumber >= 9 && log.SeverityNumber < 13 // INFO level
		}
		if !severityMatch {
			return false
		}
	}

	// Check pattern match
	if config.Pattern != "" {
		matched, err := e.matchPattern(log.Body, config.Pattern)
		if err != nil || !matched {
			return false
		}
	}

	return true
}

// checkMetricThreshold checks if a metric value exceeds the threshold over the time window
func (e *TelemetryTriggerEvaluator) checkMetricThreshold(workflowID string, config *types.TelemetryTriggerConfig, value float64, timestamp time.Time) bool {
	// Parse time window
	windowDuration, err := time.ParseDuration(config.TimeWindow)
	if err != nil {
		e.logger.Warn("Invalid time window for metric trigger",
			zap.String("workflow_id", workflowID),
			zap.String("time_window", config.TimeWindow),
			zap.Error(err))
		return false
	}

	// Create cache key
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", workflowID, config.MetricName, config.AgentID, config.ServiceName)

	// Get or create metric window
	e.mu.RUnlock()
	e.mu.Lock()
	window, exists := e.metricCache[cacheKey]
	if !exists {
		window = &metricWindow{
			values: make([]metricValue, 0),
			window: windowDuration,
		}
		e.metricCache[cacheKey] = window
	}
	e.mu.Unlock()
	e.mu.RLock()

	// Add new value to window
	window.mu.Lock()
	window.values = append(window.values, metricValue{
		timestamp: timestamp,
		value:     value,
	})

	// Remove old values outside the window
	cutoff := timestamp.Add(-windowDuration)
	validStart := 0
	for i, v := range window.values {
		if v.timestamp.After(cutoff) {
			validStart = i
			break
		}
	}
	if validStart > 0 {
		window.values = window.values[validStart:]
	}
	window.mu.Unlock()

	// Check if any value in the window exceeds threshold
	window.mu.RLock()
	defer window.mu.RUnlock()

	for _, v := range window.values {
		if e.compareValue(v.value, config.Operator, config.Threshold) {
			return true
		}
	}

	return false
}

// compareValue compares a value against a threshold using the specified operator
func (e *TelemetryTriggerEvaluator) compareValue(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	default:
		return false
	}
}

// matchPattern performs pattern matching (supports simple string contains or regex)
func (e *TelemetryTriggerEvaluator) matchPattern(text, pattern string) (bool, error) {
	// Try regex first
	if strings.HasPrefix(pattern, "^") || strings.Contains(pattern, ".*") || strings.Contains(pattern, "\\") {
		matched, err := regexp.MatchString(pattern, text)
		return matched, err
	}

	// Otherwise, simple string contains
	return strings.Contains(text, pattern), nil
}

// cleanupCache periodically cleans up old metric windows
func (e *TelemetryTriggerEvaluator) cleanupCache() {
	for range e.cacheCleanup.C {
		e.mu.Lock()
		now := time.Now()
		for key, window := range e.metricCache {
			window.mu.Lock()
			// Remove windows that haven't been updated in 2x their window duration
			if len(window.values) > 0 {
				lastUpdate := window.values[len(window.values)-1].timestamp
				if now.Sub(lastUpdate) > window.window*2 {
					delete(e.metricCache, key)
				}
			} else {
				// Remove empty windows older than 1 hour
				delete(e.metricCache, key)
			}
			window.mu.Unlock()
		}
		e.mu.Unlock()
	}
}

// Stop stops the evaluator and cleans up resources
func (e *TelemetryTriggerEvaluator) Stop() {
	if e.cacheCleanup != nil {
		e.cacheCleanup.Stop()
	}
}
