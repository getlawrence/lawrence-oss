package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore"
)

// TelemetryQueryServiceImpl implements the TelemetryQueryService interface
type TelemetryQueryServiceImpl struct {
	telemetryReader telemetrystore.Reader
	agentService    AgentService
	logger          *zap.Logger
}

// NewTelemetryQueryService creates a new telemetry query service
func NewTelemetryQueryService(telemetryReader telemetrystore.Reader, agentService AgentService, logger *zap.Logger) TelemetryQueryService {
	return &TelemetryQueryServiceImpl{
		telemetryReader: telemetryReader,
		agentService:    agentService,
		logger:          logger,
	}
}

// QueryMetrics queries metrics data
func (s *TelemetryQueryServiceImpl) QueryMetrics(ctx context.Context, query MetricQuery) ([]Metric, error) {
	// Convert service query to storage query
	storageQuery := telemetrystore.MetricQuery{
		AgentID:    query.AgentID,
		GroupID:    query.GroupID,
		MetricName: query.MetricName,
		StartTime:  query.StartTime,
		EndTime:    query.EndTime,
		Limit:      query.Limit,
	}

	storageMetrics, err := s.telemetryReader.QueryMetrics(ctx, storageQuery)
	if err != nil {
		return nil, err
	}

	// Convert storage metrics to service metrics
	metrics := make([]Metric, len(storageMetrics))
	for i, metric := range storageMetrics {
		metrics[i] = Metric{
			Timestamp:        metric.Timestamp,
			AgentID:          metric.AgentID,
			GroupID:          metric.GroupID,
			ServiceName:      metric.ServiceName,
			ConfigHash:       metric.ConfigHash,
			Name:             metric.Name,
			Value:            metric.Value,
			MetricAttributes: metric.MetricAttributes,
			Labels:           metric.Labels,
			Type:             MetricType(metric.Type),
		}
	}

	return metrics, nil
}

// QueryLogs queries logs data
func (s *TelemetryQueryServiceImpl) QueryLogs(ctx context.Context, query LogQuery) ([]Log, error) {
	// Convert service query to storage query
	storageQuery := telemetrystore.LogQuery{
		AgentID:   query.AgentID,
		GroupID:   query.GroupID,
		Severity:  query.Severity,
		Search:    query.Search,
		StartTime: query.StartTime,
		EndTime:   query.EndTime,
		Limit:     query.Limit,
	}

	storageLogs, err := s.telemetryReader.QueryLogs(ctx, storageQuery)
	if err != nil {
		return nil, err
	}

	// Convert storage logs to service logs
	logs := make([]Log, len(storageLogs))
	for i, log := range storageLogs {
		logs[i] = Log{
			Timestamp:      log.Timestamp,
			AgentID:        log.AgentID,
			GroupID:        log.GroupID,
			ServiceName:    log.ServiceName,
			SeverityText:   log.SeverityText,
			SeverityNumber: log.SeverityNumber,
			Body:           log.Body,
			TraceID:        log.TraceID,
		SpanID:         log.SpanID,
		LogAttributes:  log.LogAttributes,
		ConfigHash:     log.ConfigHash,
		}
	}

	return logs, nil
}

// QueryTraces queries traces data
func (s *TelemetryQueryServiceImpl) QueryTraces(ctx context.Context, query TraceQuery) ([]Trace, error) {
	// Convert service query to storage query
	storageQuery := telemetrystore.TraceQuery{
		AgentID:   query.AgentID,
		GroupID:   query.GroupID,
		TraceID:   query.TraceID,
		StartTime: query.StartTime,
		EndTime:   query.EndTime,
		Limit:     query.Limit,
	}

	storageTraces, err := s.telemetryReader.QueryTraces(ctx, storageQuery)
	if err != nil {
		return nil, err
	}

	// Convert storage traces to service traces
	traces := make([]Trace, len(storageTraces))
	for i, trace := range storageTraces {
		traces[i] = Trace{
			Timestamp:     trace.Timestamp,
			AgentID:       trace.AgentID,
			ConfigHash:    trace.ConfigHash,
			TraceID:       trace.TraceID,
			SpanID:        trace.SpanID,
			ParentSpanID:  trace.ParentSpanID,
			Name:          trace.Name,
			Duration:      trace.Duration,
			StatusCode:    trace.StatusCode,
			StatusMessage: trace.StatusMessage,
			Attributes:    trace.Attributes,
		}
	}

	return traces, nil
}

// QueryRaw executes a raw SQL query
func (s *TelemetryQueryServiceImpl) QueryRaw(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return s.telemetryReader.QueryRaw(ctx, query, args...)
}

// CreateRollups creates rollups for the given time window
func (s *TelemetryQueryServiceImpl) CreateRollups(ctx context.Context, window time.Time, interval RollupInterval) error {
	return s.telemetryReader.CreateRollups(ctx, window, telemetrystore.RollupInterval(interval))
}

// QueryRollups queries rollup data
func (s *TelemetryQueryServiceImpl) QueryRollups(ctx context.Context, query RollupQuery) ([]Rollup, error) {
	// Convert service query to storage query
	storageQuery := telemetrystore.RollupQuery{
		AgentID:    query.AgentID,
		GroupID:    query.GroupID,
		MetricName: query.MetricName,
		StartTime:  query.StartTime,
		EndTime:    query.EndTime,
		Interval:   telemetrystore.RollupInterval(query.Interval),
	}

	storageRollups, err := s.telemetryReader.QueryRollups(ctx, storageQuery)
	if err != nil {
		return nil, err
	}

	// Convert storage rollups to service rollups
	rollups := make([]Rollup, len(storageRollups))
	for i, rollup := range storageRollups {
		rollups[i] = Rollup{
			WindowStart: rollup.WindowStart,
			AgentID:     rollup.AgentID,
			GroupID:     rollup.GroupID,
			MetricName:  rollup.MetricName,
			Count:       rollup.Count,
			Sum:         rollup.Sum,
			Avg:         rollup.Avg,
			Min:         rollup.Min,
			Max:         rollup.Max,
			Interval:    RollupInterval(rollup.Interval),
		}
	}

	return rollups, nil
}

// CleanupOldData cleans up old telemetry data
func (s *TelemetryQueryServiceImpl) CleanupOldData(ctx context.Context, retention time.Duration) error {
	return s.telemetryReader.CleanupOldData(ctx, retention)
}

// GetTelemetryOverview gets the telemetry overview
func (s *TelemetryQueryServiceImpl) GetTelemetryOverview(ctx context.Context) (*TelemetryOverview, error) {
	// Get counts from telemetry store
	var metricsCount, logsCount, tracesCount int64

	// Count metrics
	metricsQuery := `SELECT COUNT(*) FROM (SELECT 1 FROM metrics_sum UNION ALL SELECT 1 FROM metrics_gauge) AS all_metrics`
	if rows, err := s.telemetryReader.QueryRaw(ctx, metricsQuery); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			metricsCount = count
		}
	}

	// Count logs
	logsQuery := `SELECT COUNT(*) FROM logs`
	if rows, err := s.telemetryReader.QueryRaw(ctx, logsQuery); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			logsCount = count
		}
	}

	// Count traces
	tracesQuery := `SELECT COUNT(*) FROM traces`
	if rows, err := s.telemetryReader.QueryRaw(ctx, tracesQuery); err == nil && len(rows) > 0 {
		if count, ok := rows[0]["count"].(int64); ok {
			tracesCount = count
		}
	}

	// Get active agents count
	agents, err := s.agentService.ListAgents(ctx)
	activeAgents := 0
	if err == nil {
		for _, agent := range agents {
			if agent.Status == AgentStatusOnline {
				activeAgents++
			}
		}
	}

	// Get unique services
	services, err := s.GetServices(ctx)
	if err != nil {
		s.logger.Error("Failed to get services for overview", zap.Error(err))
		services = []string{}
	}

	return &TelemetryOverview{
		TotalMetrics: metricsCount,
		TotalLogs:    logsCount,
		TotalTraces:  tracesCount,
		ActiveAgents: activeAgents,
		Services:     services,
		LastUpdated:  time.Now(),
	}, nil
}

// GetServices gets the list of unique services
func (s *TelemetryQueryServiceImpl) GetServices(ctx context.Context) ([]string, error) {
	// Get unique services from all telemetry tables
	query := `
		SELECT DISTINCT service_name FROM (
			SELECT service_name FROM metrics_sum
			UNION
			SELECT service_name FROM metrics_gauge
			UNION
			SELECT service_name FROM logs
			UNION
			SELECT service_name FROM traces
		) AS all_services
		WHERE service_name IS NOT NULL AND service_name != ''
		ORDER BY service_name
	`

	rows, err := s.telemetryReader.QueryRaw(ctx, query)
	if err != nil {
		return nil, err
	}

	var services []string
	for _, row := range rows {
		if svc, ok := row["service_name"].(string); ok && svc != "" {
			services = append(services, svc)
		}
	}

	return services, nil
}
