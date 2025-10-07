package duckdb

// TelemetrySchema defines the DuckDB schema for telemetry data
const TelemetrySchema = `
-- Metrics tables
CREATE TABLE IF NOT EXISTS metrics_sum (
	timestamp TIMESTAMP NOT NULL,
	agent_id VARCHAR NOT NULL,
	group_id VARCHAR,
	group_name VARCHAR,
	service_name VARCHAR NOT NULL,
	metric_name VARCHAR NOT NULL,
	metric_description VARCHAR,
	value DOUBLE NOT NULL,
	resource_attributes JSON,
	metric_attributes JSON
);

CREATE TABLE IF NOT EXISTS metrics_gauge (
	timestamp TIMESTAMP NOT NULL,
	agent_id VARCHAR NOT NULL,
	group_id VARCHAR,
	group_name VARCHAR,
	service_name VARCHAR NOT NULL,
	metric_name VARCHAR NOT NULL,
	metric_description VARCHAR,
	value DOUBLE NOT NULL,
	resource_attributes JSON,
	metric_attributes JSON
);

CREATE TABLE IF NOT EXISTS metrics_histogram (
	timestamp TIMESTAMP NOT NULL,
	agent_id VARCHAR NOT NULL,
	group_id VARCHAR,
	group_name VARCHAR,
	service_name VARCHAR NOT NULL,
	metric_name VARCHAR NOT NULL,
	metric_description VARCHAR,
	count BIGINT NOT NULL,
	sum DOUBLE NOT NULL,
	min DOUBLE,
	max DOUBLE,
	bucket_counts BIGINT[],
	explicit_bounds DOUBLE[],
	resource_attributes JSON,
	metric_attributes JSON
);

-- Logs table
CREATE TABLE IF NOT EXISTS logs (
	timestamp TIMESTAMP NOT NULL,
	agent_id VARCHAR NOT NULL,
	group_id VARCHAR,
	group_name VARCHAR,
	service_name VARCHAR NOT NULL,
	severity_text VARCHAR,
	severity_number INTEGER,
	body VARCHAR,
	trace_id VARCHAR,
	span_id VARCHAR,
	resource_attributes JSON,
	log_attributes JSON
);

-- Traces table
CREATE TABLE IF NOT EXISTS traces (
	timestamp TIMESTAMP NOT NULL,
	agent_id VARCHAR NOT NULL,
	group_id VARCHAR,
	group_name VARCHAR,
	trace_id VARCHAR NOT NULL,
	span_id VARCHAR NOT NULL,
	parent_span_id VARCHAR,
	service_name VARCHAR NOT NULL,
	span_name VARCHAR NOT NULL,
	span_kind VARCHAR,
	duration BIGINT NOT NULL,
	status_code VARCHAR,
	status_message VARCHAR,
	resource_attributes JSON,
	span_attributes JSON,
	events JSON,
	links JSON
);

-- Rollup tables for pre-aggregated data
CREATE TABLE IF NOT EXISTS rollups_1m (
	window_start TIMESTAMP NOT NULL,
	agent_id VARCHAR,
	group_id VARCHAR,
	metric_name VARCHAR NOT NULL,
	count BIGINT NOT NULL,
	sum DOUBLE NOT NULL,
	avg DOUBLE NOT NULL,
	min DOUBLE NOT NULL,
	max DOUBLE NOT NULL,
	PRIMARY KEY (window_start, agent_id, group_id, metric_name)
);

CREATE TABLE IF NOT EXISTS rollups_5m (
	window_start TIMESTAMP NOT NULL,
	agent_id VARCHAR,
	group_id VARCHAR,
	metric_name VARCHAR NOT NULL,
	count BIGINT NOT NULL,
	sum DOUBLE NOT NULL,
	avg DOUBLE NOT NULL,
	min DOUBLE NOT NULL,
	max DOUBLE NOT NULL,
	PRIMARY KEY (window_start, agent_id, group_id, metric_name)
);

CREATE TABLE IF NOT EXISTS rollups_1h (
	window_start TIMESTAMP NOT NULL,
	agent_id VARCHAR,
	group_id VARCHAR,
	metric_name VARCHAR NOT NULL,
	count BIGINT NOT NULL,
	sum DOUBLE NOT NULL,
	avg DOUBLE NOT NULL,
	min DOUBLE NOT NULL,
	max DOUBLE NOT NULL,
	PRIMARY KEY (window_start, agent_id, group_id, metric_name)
);

CREATE TABLE IF NOT EXISTS rollups_1d (
	window_start TIMESTAMP NOT NULL,
	agent_id VARCHAR,
	group_id VARCHAR,
	metric_name VARCHAR NOT NULL,
	count BIGINT NOT NULL,
	sum DOUBLE NOT NULL,
	avg DOUBLE NOT NULL,
	min DOUBLE NOT NULL,
	max DOUBLE NOT NULL,
	PRIMARY KEY (window_start, agent_id, group_id, metric_name)
);
`
