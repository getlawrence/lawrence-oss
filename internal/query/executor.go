// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package query

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Executor executes Lawrence QL queries
type Executor struct {
	telemetryService services.TelemetryQueryService
	logger           *zap.Logger
}

// NewExecutor creates a new query executor
func NewExecutor(telemetryService services.TelemetryQueryService, logger *zap.Logger) *Executor {
	return &Executor{
		telemetryService: telemetryService,
		logger:           logger,
	}
}

// Execute executes a Lawrence QL query and returns results
func (e *Executor) Execute(ctx context.Context, query Query, execCtx *ExecutionContext) ([]QueryResult, *QueryMeta, error) {
	startTime := time.Now()

	// Use the visitor pattern to execute the query
	visitor := &ExecutorVisitor{
		executor: e,
		ctx:      ctx,
		execCtx:  execCtx,
	}

	results, err := query.Accept(visitor)
	if err != nil {
		return nil, nil, err
	}

	queryResults, ok := results.([]QueryResult)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected result type from query execution")
	}

	meta := &QueryMeta{
		ExecutionTime: time.Since(startTime),
		RowCount:      len(queryResults),
		QueryType:     fmt.Sprintf("%T", query),
		UsedRollups:   false, // TODO: detect rollup usage
	}

	return queryResults, meta, nil
}

// ExecutorVisitor implements QueryVisitor to execute queries
type ExecutorVisitor struct {
	executor *Executor
	ctx      context.Context
	execCtx  *ExecutionContext
}

// VisitTelemetryQuery executes a telemetry query
func (v *ExecutorVisitor) VisitTelemetryQuery(q *TelemetryQuery) (interface{}, error) {
	// Determine time range
	startTime, endTime := v.getTimeRange(q)

	switch q.Type {
	case TelemetryTypeMetrics:
		return v.executeMetricsQuery(q, startTime, endTime)
	case TelemetryTypeLogs:
		return v.executeLogsQuery(q, startTime, endTime)
	case TelemetryTypeTraces:
		return v.executeTracesQuery(q, startTime, endTime)
	default:
		return nil, fmt.Errorf("unknown telemetry type: %s", q.Type)
	}
}

// executeMetricsQuery executes a metrics query
func (v *ExecutorVisitor) executeMetricsQuery(q *TelemetryQuery, startTime, endTime time.Time) ([]QueryResult, error) {
	// Build metric query
	metricQuery := services.MetricQuery{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     q.Limit,
	}

	// Apply selectors
	for label, selector := range q.Selectors {
		switch label {
		case "agent_id":
			if v.execCtx.AgentID != nil {
				metricQuery.AgentID = v.execCtx.AgentID
			} else if selector.Operator == SelectorOpEqual {
				agentID, err := uuid.Parse(selector.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid agent_id: %v", err)
				}
				metricQuery.AgentID = &agentID
			}
		case "group_id":
			if v.execCtx.GroupID != nil {
				metricQuery.GroupID = v.execCtx.GroupID
			} else if selector.Operator == SelectorOpEqual {
				metricQuery.GroupID = &selector.Value
			}
		case "metric", "name", "metric_name":
			if selector.Operator == SelectorOpEqual {
				metricQuery.MetricName = &selector.Value
				v.executor.logger.Debug("Setting metric name filter", zap.String("metric_name", selector.Value))
			}
		}
	}

	// Execute query
	v.executor.logger.Debug("Querying metrics",
		zap.Any("metric_name", metricQuery.MetricName),
		zap.Time("start_time", metricQuery.StartTime),
		zap.Time("end_time", metricQuery.EndTime),
		zap.Int("limit", metricQuery.Limit))
	metrics, err := v.executor.telemetryService.QueryMetrics(v.ctx, metricQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %v", err)
	}

	v.executor.logger.Debug("Query returned metrics", zap.Int("count", len(metrics)))

	// Convert to QueryResults
	results := make([]QueryResult, 0, len(metrics))
	for _, metric := range metrics {
		// Apply additional filters (regex, etc.) - but exclude special selectors that were already applied in the query
		labelSelectors := make(map[string]*Selector)
		for label, selector := range q.Selectors {
			// Skip selectors that were already applied in the database query
			if label != "agent_id" && label != "group_id" && label != "metric" && label != "name" && label != "metric_name" {
				labelSelectors[label] = selector
			}
		}

		if !v.matchesSelectors(metric.Labels, labelSelectors) {
			continue
		}

		results = append(results, QueryResult{
			Type:      TelemetryTypeMetrics,
			Timestamp: metric.Timestamp,
			Labels:    metric.Labels,
			Value:     metric.Value,
			Data: map[string]interface{}{
				"name":        metric.Name,
				"type":        metric.Type,
				"agent_id":    metric.AgentID.String(),
				"config_hash": metric.ConfigHash,
			},
		})
	}

	return results, nil
}

// executeLogsQuery executes a logs query
func (v *ExecutorVisitor) executeLogsQuery(q *TelemetryQuery, startTime, endTime time.Time) ([]QueryResult, error) {
	// Build log query
	logQuery := services.LogQuery{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     q.Limit,
	}

	// Apply selectors
	for label, selector := range q.Selectors {
		switch label {
		case "agent_id":
			if v.execCtx.AgentID != nil {
				logQuery.AgentID = v.execCtx.AgentID
			} else if selector.Operator == SelectorOpEqual {
				agentID, err := uuid.Parse(selector.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid agent_id: %v", err)
				}
				logQuery.AgentID = &agentID
			}
		case "group_id":
			if v.execCtx.GroupID != nil {
				logQuery.GroupID = v.execCtx.GroupID
			} else if selector.Operator == SelectorOpEqual {
				logQuery.GroupID = &selector.Value
			}
		case "severity", "level":
			if selector.Operator == SelectorOpEqual {
				logQuery.Severity = &selector.Value
			}
		case "body", "message", "search":
			if selector.Operator == SelectorOpEqual || selector.Operator == SelectorOpRegex {
				logQuery.Search = &selector.Value
			}
		}
	}

	// Execute query
	logs, err := v.executor.telemetryService.QueryLogs(v.ctx, logQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %v", err)
	}

	// Convert to QueryResults
	results := make([]QueryResult, 0, len(logs))
	for _, log := range logs {
		// Apply additional filters
		if !v.matchesSelectors(log.Attributes, q.Selectors) {
			continue
		}

		results = append(results, QueryResult{
			Type:      TelemetryTypeLogs,
			Timestamp: log.Timestamp,
			Labels:    log.Attributes,
			Value:     log.Body,
			Data: map[string]interface{}{
				"severity":    log.Severity,
				"agent_id":    log.AgentID.String(),
				"config_hash": log.ConfigHash,
			},
		})
	}

	return results, nil
}

// executeTracesQuery executes a traces query
func (v *ExecutorVisitor) executeTracesQuery(q *TelemetryQuery, startTime, endTime time.Time) ([]QueryResult, error) {
	// Build trace query
	traceQuery := services.TraceQuery{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     q.Limit,
	}

	// Apply selectors
	for label, selector := range q.Selectors {
		switch label {
		case "agent_id":
			if v.execCtx.AgentID != nil {
				traceQuery.AgentID = v.execCtx.AgentID
			} else if selector.Operator == SelectorOpEqual {
				agentID, err := uuid.Parse(selector.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid agent_id: %v", err)
				}
				traceQuery.AgentID = &agentID
			}
		case "group_id":
			if v.execCtx.GroupID != nil {
				traceQuery.GroupID = v.execCtx.GroupID
			} else if selector.Operator == SelectorOpEqual {
				traceQuery.GroupID = &selector.Value
			}
		case "trace_id":
			if selector.Operator == SelectorOpEqual {
				traceQuery.TraceID = &selector.Value
			}
		}
	}

	// Execute query
	traces, err := v.executor.telemetryService.QueryTraces(v.ctx, traceQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %v", err)
	}

	// Convert to QueryResults
	results := make([]QueryResult, 0, len(traces))
	for _, trace := range traces {
		// Apply additional filters
		if !v.matchesSelectors(trace.Attributes, q.Selectors) {
			continue
		}

		results = append(results, QueryResult{
			Type:      TelemetryTypeTraces,
			Timestamp: trace.Timestamp,
			Labels:    trace.Attributes,
			Value:     trace.Duration,
			Data: map[string]interface{}{
				"trace_id":       trace.TraceID,
				"span_id":        trace.SpanID,
				"parent_span_id": trace.ParentSpanID,
				"name":           trace.Name,
				"status_code":    trace.StatusCode,
				"status_message": trace.StatusMessage,
				"agent_id":       trace.AgentID.String(),
				"config_hash":    trace.ConfigHash,
			},
		})
	}

	return results, nil
}

// VisitBinaryOp executes a binary operation
func (v *ExecutorVisitor) VisitBinaryOp(b *BinaryOp) (interface{}, error) {
	// Execute left and right queries
	leftResults, err := b.Left.Accept(v)
	if err != nil {
		return nil, err
	}
	leftQueryResults, ok := leftResults.([]QueryResult)
	if !ok {
		return nil, fmt.Errorf("left operand did not return query results")
	}

	rightResults, err := b.Right.Accept(v)
	if err != nil {
		return nil, err
	}
	rightQueryResults, ok := rightResults.([]QueryResult)
	if !ok {
		return nil, fmt.Errorf("right operand did not return query results")
	}

	// Apply binary operation
	switch b.Operator {
	case BinaryOpAdd:
		// Merge results
		return append(leftQueryResults, rightQueryResults...), nil
	case BinaryOpAnd:
		// Intersection based on timestamp and labels
		return v.intersectResults(leftQueryResults, rightQueryResults), nil
	case BinaryOpOr:
		// Union - just merge
		return append(leftQueryResults, rightQueryResults...), nil
	default:
		return nil, fmt.Errorf("unsupported binary operator: %s", b.Operator)
	}
}

// VisitFunctionCall executes a function call
func (v *ExecutorVisitor) VisitFunctionCall(f *FunctionCall) (interface{}, error) {
	if len(f.Args) == 0 {
		return nil, fmt.Errorf("function %s requires at least one argument", f.Name)
	}

	// Execute the argument query
	argResults, err := f.Args[0].Accept(v)
	if err != nil {
		return nil, err
	}
	queryResults, ok := argResults.([]QueryResult)
	if !ok {
		return nil, fmt.Errorf("function argument did not return query results")
	}

	// Apply function
	function, ok := GetFunction(f.Name)
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", f.Name)
	}

	return function.Apply(queryResults)
}

// VisitAggregation executes an aggregation query
func (v *ExecutorVisitor) VisitAggregation(a *Aggregation) (interface{}, error) {
	// Execute the inner query
	queryResults, err := a.Query.Accept(v)
	if err != nil {
		return nil, err
	}
	results, ok := queryResults.([]QueryResult)
	if !ok {
		return nil, fmt.Errorf("aggregation query did not return query results")
	}

	// Apply aggregation
	function, ok := GetFunction(a.Function)
	if !ok {
		return nil, fmt.Errorf("unknown aggregation function: %s", a.Function)
	}

	// Group by labels if specified
	if len(a.By) > 0 {
		return v.groupAndAggregate(results, a.By, function)
	}

	// No grouping - aggregate all
	return function.Apply(results)
}

// getTimeRange determines the time range for the query
func (v *ExecutorVisitor) getTimeRange(q *TelemetryQuery) (time.Time, time.Time) {
	now := time.Now()

	// Use execution context times if provided
	if v.execCtx.StartTime != nil && v.execCtx.EndTime != nil {
		return *v.execCtx.StartTime, *v.execCtx.EndTime
	}

	// Use query times if provided
	if q.StartTime != nil && q.EndTime != nil {
		return *q.StartTime, *q.EndTime
	}

	// Use duration from query
	if q.Duration > 0 {
		return now.Add(-q.Duration), now
	}

	// Default to last 5 minutes
	return now.Add(-5 * time.Minute), now
}

// matchesSelectors checks if labels match all selectors
func (v *ExecutorVisitor) matchesSelectors(labels map[string]string, selectors map[string]*Selector) bool {
	for label, selector := range selectors {
		value, ok := labels[label]

		switch selector.Operator {
		case SelectorOpEqual:
			if !ok || value != selector.Value {
				return false
			}
		case SelectorOpNotEqual:
			if ok && value == selector.Value {
				return false
			}
		case SelectorOpRegex:
			if !ok {
				return false
			}
			matched, err := regexp.MatchString(selector.Value, value)
			if err != nil || !matched {
				return false
			}
		case SelectorOpNotRegex:
			if !ok {
				continue
			}
			matched, err := regexp.MatchString(selector.Value, value)
			if err != nil || matched {
				return false
			}
		}
	}
	return true
}

// intersectResults returns results that appear in both sets based on timestamp and labels
func (v *ExecutorVisitor) intersectResults(left, right []QueryResult) []QueryResult {
	results := []QueryResult{}

	// Create a map for faster lookup
	rightMap := make(map[string]QueryResult)
	for _, r := range right {
		key := v.resultKey(r)
		rightMap[key] = r
	}

	// Find intersections
	for _, l := range left {
		key := v.resultKey(l)
		if _, exists := rightMap[key]; exists {
			results = append(results, l)
		}
	}

	return results
}

// resultKey creates a unique key for a query result
func (v *ExecutorVisitor) resultKey(r QueryResult) string {
	return fmt.Sprintf("%s_%d", r.Type, r.Timestamp.UnixNano())
}

// groupAndAggregate groups results by labels and applies aggregation
func (v *ExecutorVisitor) groupAndAggregate(results []QueryResult, byLabels []string, function *Function) ([]QueryResult, error) {
	// Group results by specified labels
	groups := make(map[string][]QueryResult)

	for _, result := range results {
		// Create group key from specified labels
		key := ""
		for i, label := range byLabels {
			if i > 0 {
				key += "|"
			}
			if val, ok := result.Labels[label]; ok {
				key += val
			}
		}

		groups[key] = append(groups[key], result)
	}

	// Apply aggregation to each group
	aggregated := []QueryResult{}
	for _, group := range groups {
		groupResult, err := function.Apply(group)
		if err != nil {
			return nil, err
		}
		if groupResults, ok := groupResult.([]QueryResult); ok {
			aggregated = append(aggregated, groupResults...)
		}
	}

	return aggregated, nil
}
