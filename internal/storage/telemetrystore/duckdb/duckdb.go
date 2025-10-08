package duckdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore/types"
	"github.com/google/uuid"
	_ "github.com/marcboeker/go-duckdb"
	"go.uber.org/zap"
)

// Storage implements the TelemetryStorage interface using DuckDB
type Storage struct {
	db     *sql.DB
	logger *zap.Logger
}

// Reader implements the types.Reader interface
type Reader struct {
	*Storage
}

// Writer implements the types.Writer interface
type Writer struct {
	*Storage
}

// NewDuckDBReader creates a new DuckDB reader instance
func NewDuckDBReader(dbPath string, logger *zap.Logger) (types.Reader, error) {
	storage, err := NewStorage(dbPath, logger)
	if err != nil {
		return nil, err
	}
	return &Reader{Storage: storage}, nil
}

// NewDuckDBWriter creates a new DuckDB writer instance
func NewDuckDBWriter(dbPath string, logger *zap.Logger) (types.Writer, error) {
	storage, err := NewStorage(dbPath, logger)
	if err != nil {
		return nil, err
	}
	return &Writer{Storage: storage}, nil
}

// NewStorage creates a new DuckDB storage instance
func NewStorage(dbPath string, logger *zap.Logger) (*Storage, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(30 * time.Minute)

	storage := &Storage{
		db:     db,
		logger: logger,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.Info("DuckDB storage initialized", zap.String("path", dbPath))
	return storage, nil
}

// initSchema creates the DuckDB tables
func (s *Storage) initSchema() error {
	if _, err := s.db.Exec(TelemetrySchema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	s.logger.Debug("DuckDB schema initialized")
	return nil
}

// WriteMetrics writes metric data to DuckDB
func (s *Storage) WriteMetrics(ctx context.Context, metrics []types.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Group metrics by type for batch insertion
	var sums, gauges []types.Metric
	for _, m := range metrics {
		switch m.Type {
		case types.MetricTypeCounter:
			sums = append(sums, m)
		case types.MetricTypeGauge:
			gauges = append(gauges, m)
		}
	}

	// Write sums
	if len(sums) > 0 {
		if err := s.writeMetricSums(ctx, sums); err != nil {
			return err
		}
	}

	// Write gauges
	if len(gauges) > 0 {
		if err := s.writeMetricGauges(ctx, gauges); err != nil {
			return err
		}
	}

	s.logger.Debug("Wrote metrics to DuckDB", zap.Int("count", len(metrics)))
	return nil
}

// writeMetricSums writes sum metrics to DuckDB
func (s *Storage) writeMetricSums(ctx context.Context, sums []types.Metric) error {
	query := `
		INSERT INTO metrics_sum (
			timestamp, agent_id, service_name, metric_name, value, metric_attributes
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range sums {
		attrsJSON, _ := json.Marshal(m.Labels)
		serviceName := m.Labels["service.name"]
		if serviceName == "" {
			serviceName = "unknown"
		}

		_, err = stmt.ExecContext(ctx,
			m.Timestamp,
			m.AgentID.String(),
			serviceName,
			m.Name,
			m.Value,
			string(attrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert sum metric: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// writeMetricGauges writes gauge metrics to DuckDB
func (s *Storage) writeMetricGauges(ctx context.Context, gauges []types.Metric) error {
	query := `
		INSERT INTO metrics_gauge (
			timestamp, agent_id, service_name, metric_name, value, metric_attributes
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range gauges {
		attrsJSON, _ := json.Marshal(m.Labels)
		serviceName := m.Labels["service.name"]
		if serviceName == "" {
			serviceName = "unknown"
		}

		_, err = stmt.ExecContext(ctx,
			m.Timestamp,
			m.AgentID.String(),
			serviceName,
			m.Name,
			m.Value,
			string(attrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert gauge metric: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WriteLogs writes log data to DuckDB
func (s *Storage) WriteLogs(ctx context.Context, logs []types.Log) error {
	if len(logs) == 0 {
		return nil
	}

	query := `
		INSERT INTO logs (
			timestamp, agent_id, service_name, severity_text, body, log_attributes
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, log := range logs {
		attrsJSON, _ := json.Marshal(log.Attributes)
		serviceName := log.Attributes["service.name"]
		if serviceName == "" {
			serviceName = "unknown"
		}

		_, err = stmt.ExecContext(ctx,
			log.Timestamp,
			log.AgentID.String(),
			serviceName,
			log.Severity,
			log.Body,
			string(attrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert log: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Wrote logs to DuckDB", zap.Int("count", len(logs)))
	return nil
}

// WriteTraces writes trace data to DuckDB
func (s *Storage) WriteTraces(ctx context.Context, traces []types.Trace) error {
	if len(traces) == 0 {
		return nil
	}

	query := `
		INSERT INTO traces (
			timestamp, agent_id, trace_id, span_id, parent_span_id,
			service_name, span_name, duration, status_code, status_message,
			span_attributes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, trace := range traces {
		attrsJSON, _ := json.Marshal(trace.Attributes)
		serviceName := trace.Attributes["service.name"]
		if serviceName == "" {
			serviceName = "unknown"
		}

		var parentSpanID interface{}
		if trace.ParentSpanID != nil {
			parentSpanID = *trace.ParentSpanID
		}

		_, err = stmt.ExecContext(ctx,
			trace.Timestamp,
			trace.AgentID.String(),
			trace.TraceID,
			trace.SpanID,
			parentSpanID,
			serviceName,
			trace.Name,
			trace.Duration,
			trace.StatusCode,
			trace.StatusMessage,
			string(attrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert trace: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Wrote traces to DuckDB", zap.Int("count", len(traces)))
	return nil
}

// QueryMetrics queries metrics from DuckDB
func (s *Storage) QueryMetrics(ctx context.Context, query types.MetricQuery) ([]types.Metric, error) {
	sqlQuery := `
		SELECT timestamp, agent_id, group_id, service_name, metric_name, value, metric_attributes
		FROM (
			SELECT timestamp, agent_id, group_id, service_name, metric_name, value, metric_attributes FROM metrics_sum
			UNION ALL
			SELECT timestamp, agent_id, group_id, service_name, metric_name, value, metric_attributes FROM metrics_gauge
		) AS all_metrics
		WHERE timestamp >= ? AND timestamp <= ?
	`
	args := []interface{}{query.StartTime, query.EndTime}

	if query.AgentID != nil {
		sqlQuery += ` AND agent_id = ?`
		args = append(args, query.AgentID.String())
	}

	if query.MetricName != nil {
		sqlQuery += ` AND metric_name = ?`
		args = append(args, *query.MetricName)
	}

	sqlQuery += ` ORDER BY timestamp DESC`

	if query.Limit > 0 {
		sqlQuery += ` LIMIT ?`
		args = append(args, query.Limit)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []types.Metric
	for rows.Next() {
		var m types.Metric
		var agentIDStr string
		var groupID sql.NullString
		var serviceName string
		var attrsJSON string

		err := rows.Scan(&m.Timestamp, &agentIDStr, &groupID, &serviceName, &m.Name, &m.Value, &attrsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}

		m.AgentID, _ = uuid.Parse(agentIDStr)
		m.ServiceName = serviceName
		if groupID.Valid {
			m.GroupID = &groupID.String
		}
		if attrsJSON != "" {
			_ = json.Unmarshal([]byte(attrsJSON), &m.MetricAttributes)
			
			// Convert MetricAttributes to Labels (map[string]string)
			m.Labels = make(map[string]string)
			for key, value := range m.MetricAttributes {
				m.Labels[key] = fmt.Sprintf("%v", value)
			}
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// QueryLogs queries logs from DuckDB
func (s *Storage) QueryLogs(ctx context.Context, query types.LogQuery) ([]types.Log, error) {
	sqlQuery := `
		SELECT timestamp, agent_id, group_id, service_name, severity_text, severity_number, 
		       body, trace_id, span_id, log_attributes
		FROM logs
		WHERE timestamp >= ? AND timestamp <= ?
	`
	args := []interface{}{query.StartTime, query.EndTime}

	if query.AgentID != nil {
		sqlQuery += ` AND agent_id = ?`
		args = append(args, query.AgentID.String())
	}

	if query.Severity != nil {
		sqlQuery += ` AND severity_text = ?`
		args = append(args, *query.Severity)
	}

	if query.Search != nil {
		sqlQuery += ` AND body LIKE ?`
		args = append(args, "%"+*query.Search+"%")
	}

	sqlQuery += ` ORDER BY timestamp DESC`

	if query.Limit > 0 {
		sqlQuery += ` LIMIT ?`
		args = append(args, query.Limit)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []types.Log
	for rows.Next() {
		var l types.Log
		var agentIDStr string
		var groupID, traceID, spanID sql.NullString
		var serviceName string
		var attrsJSON string

		err := rows.Scan(&l.Timestamp, &agentIDStr, &groupID, &serviceName, 
			&l.SeverityText, &l.SeverityNumber, &l.Body, &traceID, &spanID, &attrsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}

		l.AgentID, _ = uuid.Parse(agentIDStr)
		l.ServiceName = serviceName
		if groupID.Valid {
			l.GroupID = &groupID.String
		}
		if traceID.Valid {
			l.TraceID = &traceID.String
		}
		if spanID.Valid {
			l.SpanID = &spanID.String
		}
		if attrsJSON != "" {
			_ = json.Unmarshal([]byte(attrsJSON), &l.LogAttributes)
			
			// Convert LogAttributes to deprecated Attributes field (map[string]string)
			l.Attributes = make(map[string]string)
			for key, value := range l.LogAttributes {
				l.Attributes[key] = fmt.Sprintf("%v", value)
			}
		}
		
		// Set deprecated Severity field from SeverityText
		l.Severity = l.SeverityText

		logs = append(logs, l)
	}

	return logs, nil
}

// QueryTraces queries traces from DuckDB
func (s *Storage) QueryTraces(ctx context.Context, query types.TraceQuery) ([]types.Trace, error) {
	sqlQuery := `
		SELECT timestamp, agent_id, trace_id, span_id, parent_span_id,
		       span_name, duration, status_code, status_message, span_attributes
		FROM traces
		WHERE timestamp >= ? AND timestamp <= ?
	`
	args := []interface{}{query.StartTime, query.EndTime}

	if query.AgentID != nil {
		sqlQuery += ` AND agent_id = ?`
		args = append(args, query.AgentID.String())
	}

	if query.TraceID != nil {
		sqlQuery += ` AND trace_id = ?`
		args = append(args, *query.TraceID)
	}

	sqlQuery += ` ORDER BY timestamp DESC`

	if query.Limit > 0 {
		sqlQuery += ` LIMIT ?`
		args = append(args, query.Limit)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %w", err)
	}
	defer rows.Close()

	var traces []types.Trace
	for rows.Next() {
		var t types.Trace
		var agentIDStr string
		var parentSpanID sql.NullString
		var attrsJSON string

		err := rows.Scan(
			&t.Timestamp, &agentIDStr, &t.TraceID, &t.SpanID, &parentSpanID,
			&t.Name, &t.Duration, &t.StatusCode, &t.StatusMessage, &attrsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace: %w", err)
		}

		t.AgentID, _ = uuid.Parse(agentIDStr)
		if parentSpanID.Valid {
			t.ParentSpanID = &parentSpanID.String
		}
		_ = json.Unmarshal([]byte(attrsJSON), &t.Attributes)

		traces = append(traces, t)
	}

	return traces, nil
}

// CreateRollups creates pre-aggregated rollup data
func (s *Storage) CreateRollups(ctx context.Context, window time.Time, interval types.RollupInterval) error {
	var tableName string
	var windowDuration time.Duration

	switch interval {
	case types.RollupInterval1m:
		tableName = "rollups_1m"
		windowDuration = 1 * time.Minute
	case types.RollupInterval5m:
		tableName = "rollups_5m"
		windowDuration = 5 * time.Minute
	case types.RollupInterval1h:
		tableName = "rollups_1h"
		windowDuration = 1 * time.Hour
	case types.RollupInterval1d:
		tableName = "rollups_1d"
		windowDuration = 24 * time.Hour
	default:
		return fmt.Errorf("invalid rollup interval: %s", interval)
	}

	windowStart := window.Truncate(windowDuration)
	windowEnd := windowStart.Add(windowDuration)

	// Create rollups from sum metrics
	query := fmt.Sprintf(`
		INSERT INTO %s (window_start, agent_id, group_id, metric_name, count, sum, avg, min, max)
		SELECT
			? as window_start,
			agent_id,
			group_id,
			metric_name,
			COUNT(*) as count,
			SUM(value) as sum,
			AVG(value) as avg,
			MIN(value) as min,
			MAX(value) as max
		FROM metrics_sum
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY agent_id, group_id, metric_name
		ON CONFLICT (window_start, agent_id, group_id, metric_name)
		DO UPDATE SET
			count = EXCLUDED.count,
			sum = EXCLUDED.sum,
			avg = EXCLUDED.avg,
			min = EXCLUDED.min,
			max = EXCLUDED.max
	`, tableName)

	_, err := s.db.ExecContext(ctx, query, windowStart, windowStart, windowEnd)
	if err != nil {
		return fmt.Errorf("failed to create rollups: %w", err)
	}

	s.logger.Debug("Created rollups", zap.String("interval", string(interval)), zap.Time("window", windowStart))
	return nil
}

// QueryRollups queries rollup data
func (s *Storage) QueryRollups(ctx context.Context, query types.RollupQuery) ([]types.Rollup, error) {
	var tableName string
	switch query.Interval {
	case types.RollupInterval1m:
		tableName = "rollups_1m"
	case types.RollupInterval5m:
		tableName = "rollups_5m"
	case types.RollupInterval1h:
		tableName = "rollups_1h"
	case types.RollupInterval1d:
		tableName = "rollups_1d"
	default:
		return nil, fmt.Errorf("invalid rollup interval: %s", query.Interval)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT window_start, agent_id, group_id, metric_name, count, sum, avg, min, max
		FROM %s
		WHERE window_start >= ? AND window_start <= ?
	`, tableName)
	args := []interface{}{query.StartTime, query.EndTime}

	if query.AgentID != nil {
		sqlQuery += ` AND agent_id = ?`
		args = append(args, query.AgentID.String())
	}

	if query.GroupID != nil {
		sqlQuery += ` AND group_id = ?`
		args = append(args, *query.GroupID)
	}

	if query.MetricName != nil {
		sqlQuery += ` AND metric_name = ?`
		args = append(args, *query.MetricName)
	}

	sqlQuery += ` ORDER BY window_start DESC`

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rollups: %w", err)
	}
	defer rows.Close()

	var rollups []types.Rollup
	for rows.Next() {
		var r types.Rollup
		var agentIDStr, groupIDStr sql.NullString

		err := rows.Scan(
			&r.WindowStart, &agentIDStr, &groupIDStr, &r.MetricName,
			&r.Count, &r.Sum, &r.Avg, &r.Min, &r.Max,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollup: %w", err)
		}

		if agentIDStr.Valid {
			agentID, _ := uuid.Parse(agentIDStr.String)
			r.AgentID = &agentID
		}
		if groupIDStr.Valid {
			r.GroupID = &groupIDStr.String
		}
		r.Interval = query.Interval

		rollups = append(rollups, r)
	}

	return rollups, nil
}

// QueryRaw executes a raw SQL query and returns results as a map
func (s *Storage) QueryRaw(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare result
	var results []map[string]interface{}

	// Process rows
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row into value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = values[i]
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// CleanupOldData removes old telemetry data based on retention policy
func (s *Storage) CleanupOldData(ctx context.Context, retention time.Duration) error {
	cutoffTime := time.Now().Add(-retention)

	tables := []string{"metrics_sum", "metrics_gauge", "metrics_histogram", "logs", "traces"}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE timestamp < ?", table)
		result, err := s.db.ExecContext(ctx, query, cutoffTime)
		if err != nil {
			return fmt.Errorf("failed to cleanup %s: %w", table, err)
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			s.logger.Info("Cleaned up old data", zap.String("table", table), zap.Int64("rows", rows))
		}
	}

	return nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	s.logger.Info("DuckDB storage closed")
	return nil
}

// WriteTracesFromOTLP writes trace data from OTLP parser format
func (s *Storage) WriteTracesFromOTLP(ctx context.Context, traces []otlp.TraceData) error {
	if len(traces) == 0 {
		return nil
	}

	query := `
		INSERT INTO traces (
			timestamp, agent_id, group_id, group_name, trace_id, span_id, parent_span_id,
			service_name, span_name, duration, status_code, resource_attributes, span_attributes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, trace := range traces {
		resourceAttrsJSON, _ := json.Marshal(trace.ResourceAttributes)
		spanAttrsJSON, _ := json.Marshal(trace.SpanAttributes)

		var parentSpanID interface{}
		if trace.ParentSpanId != "" {
			parentSpanID = trace.ParentSpanId
		}

		_, err = stmt.ExecContext(ctx,
			trace.Timestamp,
			trace.AgentID,
			trace.GroupID,
			trace.GroupName,
			trace.TraceId,
			trace.SpanId,
			parentSpanID,
			trace.ServiceName,
			trace.SpanName,
			trace.Duration,
			trace.StatusCode,
			string(resourceAttrsJSON),
			string(spanAttrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert trace: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Wrote OTLP traces to DuckDB", zap.Int("count", len(traces)))
	return nil
}

// WriteLogsFromOTLP writes log data from OTLP parser format
func (s *Storage) WriteLogsFromOTLP(ctx context.Context, logs []otlp.LogData) error {
	if len(logs) == 0 {
		return nil
	}

	query := `
		INSERT INTO logs (
			timestamp, agent_id, group_id, group_name, service_name,
			severity_text, severity_number, body, trace_id, span_id,
			resource_attributes, log_attributes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, log := range logs {
		resourceAttrsJSON, _ := json.Marshal(log.ResourceAttributes)
		logAttrsJSON, _ := json.Marshal(log.LogAttributes)

		var traceID, spanID interface{}
		if log.TraceId != "" {
			traceID = log.TraceId
		}
		if log.SpanId != "" {
			spanID = log.SpanId
		}

		_, err = stmt.ExecContext(ctx,
			log.Timestamp,
			log.AgentID,
			log.GroupID,
			log.GroupName,
			log.ServiceName,
			log.SeverityText,
			log.SeverityNumber,
			log.Body,
			traceID,
			spanID,
			string(resourceAttrsJSON),
			string(logAttrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert log: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Wrote OTLP logs to DuckDB", zap.Int("count", len(logs)))
	return nil
}

// WriteMetricsFromOTLP writes metric data from OTLP parser format
func (s *Storage) WriteMetricsFromOTLP(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) error {
	s.logger.Debug("WriteMetricsFromOTLP called",
		zap.Int("sums", len(sums)),
		zap.Int("gauges", len(gauges)),
		zap.Int("histograms", len(histograms)))

	// Write sums
	if len(sums) > 0 {
		s.logger.Debug("Writing sum metrics", zap.Int("count", len(sums)))
		if err := s.writeOTLPSums(ctx, sums); err != nil {
			s.logger.Error("Failed to write sum metrics", zap.Error(err))
			return err
		}
		s.logger.Debug("Successfully wrote sum metrics")
	}

	// Write gauges
	if len(gauges) > 0 {
		s.logger.Debug("Writing gauge metrics", zap.Int("count", len(gauges)))
		if err := s.writeOTLPGauges(ctx, gauges); err != nil {
			s.logger.Error("Failed to write gauge metrics", zap.Error(err))
			return err
		}
		s.logger.Debug("Successfully wrote gauge metrics")
	}

	// Write histograms
	if len(histograms) > 0 {
		s.logger.Debug("Writing histogram metrics", zap.Int("count", len(histograms)))
		if err := s.writeOTLPHistograms(ctx, histograms); err != nil {
			s.logger.Error("Failed to write histogram metrics", zap.Error(err))
			return err
		}
		s.logger.Debug("Successfully wrote histogram metrics")
	}

	s.logger.Debug("Wrote OTLP metrics to DuckDB",
		zap.Int("sums", len(sums)),
		zap.Int("gauges", len(gauges)),
		zap.Int("histograms", len(histograms)))
	return nil
}

func (s *Storage) writeOTLPSums(ctx context.Context, sums []otlp.MetricSumData) error {
	if len(sums) == 0 {
		return nil
	}

	query := `
		INSERT INTO metrics_sum (
			timestamp, agent_id, group_id, group_name, service_name,
			metric_name, metric_description, value, resource_attributes, metric_attributes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	s.logger.Debug("Starting writeOTLPSums transaction", zap.Int("count", len(sums)))

	// Process in smaller batches to avoid overwhelming the database
	batchSize := 10
	for i := 0; i < len(sums); i += batchSize {
		end := i + batchSize
		if end > len(sums) {
			end = len(sums)
		}

		batch := sums[i:end]
		s.logger.Debug("Processing batch", zap.Int("start", i), zap.Int("end", end), zap.Int("batch_size", len(batch)))

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			s.logger.Error("Failed to begin transaction", zap.Error(err))
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			_ = tx.Rollback()
			s.logger.Error("Failed to prepare statement", zap.Error(err))
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		for j, m := range batch {
			resourceAttrsJSON, _ := json.Marshal(m.ResourceAttributes)
			metricAttrsJSON, _ := json.Marshal(m.Attributes)

			_, err = stmt.ExecContext(ctx,
				m.TimeUnix,
				m.AgentID,
				m.GroupID,
				m.GroupName,
				m.ServiceName,
				m.MetricName,
				m.MetricDescription,
				m.Value,
				string(resourceAttrsJSON),
				string(metricAttrsJSON),
			)
			if err != nil {
				stmt.Close()
				_ = tx.Rollback()
				s.logger.Error("Failed to insert sum metric", zap.Error(err), zap.Int("batch_index", j), zap.String("metric_name", m.MetricName))
				return fmt.Errorf("failed to insert sum metric: %w", err)
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			s.logger.Error("Failed to commit transaction", zap.Error(err))
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		s.logger.Debug("Batch committed successfully", zap.Int("batch_size", len(batch)))
	}

	s.logger.Debug("All batches completed successfully")
	return nil
}

func (s *Storage) writeOTLPGauges(ctx context.Context, gauges []otlp.MetricGaugeData) error {
	query := `
		INSERT INTO metrics_gauge (
			timestamp, agent_id, group_id, group_name, service_name,
			metric_name, metric_description, value, resource_attributes, metric_attributes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range gauges {
		resourceAttrsJSON, _ := json.Marshal(m.ResourceAttributes)
		metricAttrsJSON, _ := json.Marshal(m.Attributes)

		_, err = stmt.ExecContext(ctx,
			m.TimeUnix,
			m.AgentID,
			m.GroupID,
			m.GroupName,
			m.ServiceName,
			m.MetricName,
			m.MetricDescription,
			m.Value,
			string(resourceAttrsJSON),
			string(metricAttrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert gauge metric: %w", err)
		}
	}

	return tx.Commit()
}

func (s *Storage) writeOTLPHistograms(ctx context.Context, histograms []otlp.MetricHistogramData) error {
	query := `
		INSERT INTO metrics_histogram (
			timestamp, agent_id, group_id, group_name, service_name,
			metric_name, metric_description, count, sum, min, max,
			bucket_counts, explicit_bounds, resource_attributes, metric_attributes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range histograms {
		resourceAttrsJSON, _ := json.Marshal(m.ResourceAttributes)
		metricAttrsJSON, _ := json.Marshal(m.Attributes)

		// Convert slices to arrays for DuckDB
		bucketCountsStr := fmt.Sprintf("[%s]", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(m.BucketCounts)), ","), "[]"))
		explicitBoundsStr := fmt.Sprintf("[%s]", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(m.ExplicitBounds)), ","), "[]"))

		_, err = stmt.ExecContext(ctx,
			m.TimeUnix,
			m.AgentID,
			m.GroupID,
			m.GroupName,
			m.ServiceName,
			m.MetricName,
			m.MetricDescription,
			m.Count,
			m.Sum,
			m.Min,
			m.Max,
			bucketCountsStr,
			explicitBoundsStr,
			string(resourceAttrsJSON),
			string(metricAttrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert histogram metric: %w", err)
		}
	}

	return tx.Commit()
}

// Writer interface implementations for types.Writer
// These methods directly use OTLP parsed types

// WriteTraces implements types.Writer
func (w *Writer) WriteTraces(ctx context.Context, traces []otlp.TraceData) error {
	return w.Storage.WriteTracesFromOTLP(ctx, traces)
}

// WriteMetrics implements types.Writer
func (w *Writer) WriteMetrics(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) error {
	return w.Storage.WriteMetricsFromOTLP(ctx, sums, gauges, histograms)
}

// WriteLogs implements types.Writer
func (w *Writer) WriteLogs(ctx context.Context, logs []otlp.LogData) error {
	return w.Storage.WriteLogsFromOTLP(ctx, logs)
}
