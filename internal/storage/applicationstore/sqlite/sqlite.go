package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// Storage implements the ApplicationStore interface using SQLite
type Storage struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string, logger *zap.Logger) (types.ApplicationStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	storage := &Storage{
		db:     db,
		logger: logger,
	}

	// Run migrations
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("SQLite storage initialized", zap.String("path", dbPath))
	return storage, nil
}

// migrate runs database migrations
func (s *Storage) migrate() error {
	// Create tables if they don't exist
	createTables := `
		CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			labels TEXT,
			status TEXT NOT NULL DEFAULT 'offline',
			last_seen DATETIME NOT NULL,
			group_id TEXT,
			group_name TEXT,
			version TEXT,
			capabilities TEXT,
			effective_config TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS groups (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			labels TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS configs (
			id TEXT PRIMARY KEY,
			name TEXT,
			agent_id TEXT,
			group_id TEXT,
			config_hash TEXT NOT NULL,
			content TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
			FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_agents_group_id ON agents(group_id);
		CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
		CREATE INDEX IF NOT EXISTS idx_configs_agent_id ON configs(agent_id);
		CREATE INDEX IF NOT EXISTS idx_configs_group_id ON configs(group_id);
		CREATE INDEX IF NOT EXISTS idx_configs_config_hash ON configs(config_hash);

		CREATE TABLE IF NOT EXISTS workflows (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			schedule_cron TEXT,
			schedule_timezone TEXT,
			webhook_url TEXT,
			webhook_secret TEXT,
			conditions TEXT,
			created_by TEXT,
			last_run DATETIME,
			next_run DATETIME,
			run_count INTEGER NOT NULL DEFAULT 0,
			error_count INTEGER NOT NULL DEFAULT 0,
			last_error TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS workflow_executions (
			id TEXT PRIMARY KEY,
			workflow_id TEXT NOT NULL,
			workflow_name TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			duration_ms INTEGER,
			actions_executed INTEGER NOT NULL DEFAULT 0,
			actions_succeeded INTEGER NOT NULL DEFAULT 0,
			actions_failed INTEGER NOT NULL DEFAULT 0,
			configs_created TEXT,
			error TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS delayed_action_queue (
			id TEXT PRIMARY KEY,
			workflow_id TEXT NOT NULL,
			execution_id TEXT NOT NULL,
			action TEXT NOT NULL,
			scheduled_for DATETIME NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			metadata TEXT,
			error TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE,
			FOREIGN KEY (execution_id) REFERENCES workflow_executions(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_workflows_type ON workflows(type);
		CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
		CREATE INDEX IF NOT EXISTS idx_workflows_next_run ON workflows(next_run);
		CREATE INDEX IF NOT EXISTS idx_delayed_actions_scheduled_for ON delayed_action_queue(scheduled_for);
		CREATE INDEX IF NOT EXISTS idx_delayed_actions_status ON delayed_action_queue(status);
		CREATE INDEX IF NOT EXISTS idx_workflow_executions_workflow_id ON workflow_executions(workflow_id);
		CREATE INDEX IF NOT EXISTS idx_workflow_executions_status ON workflow_executions(status);

		-- Normalized workflow tables
		CREATE TABLE IF NOT EXISTS workflow_triggers (
			workflow_id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			schedule_cron TEXT,
			schedule_timezone TEXT,
			webhook_url TEXT,
			webhook_secret TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS workflow_steps (
			id TEXT PRIMARY KEY,
			workflow_id TEXT NOT NULL,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			order_index INTEGER NOT NULL,
			position_x REAL,
			position_y REAL,
			config_json TEXT NOT NULL,
			retry_enabled INTEGER NOT NULL DEFAULT 0,
			retry_count INTEGER NOT NULL DEFAULT 0,
			retry_delay_ms INTEGER NOT NULL DEFAULT 0,
			continue_on_error INTEGER NOT NULL DEFAULT 0,
			depends_on TEXT, -- JSON array of step IDs
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS step_executions (
			id TEXT PRIMARY KEY,
			step_id TEXT NOT NULL,
			workflow_id TEXT NOT NULL,
			workflow_execution_id TEXT NOT NULL,
			step_name TEXT NOT NULL,
			step_type TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			duration_ms INTEGER,
			retry_attempt INTEGER NOT NULL DEFAULT 0,
			error TEXT,
			error_code TEXT,
			input_data TEXT, -- JSON
			output_data TEXT, -- JSON
			metadata TEXT, -- JSON
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (step_id) REFERENCES workflow_steps(id) ON DELETE CASCADE,
			FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE,
			FOREIGN KEY (workflow_execution_id) REFERENCES workflow_executions(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS step_execution_logs (
			id TEXT PRIMARY KEY,
			step_execution_id TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			data TEXT,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (step_execution_id) REFERENCES step_executions(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow_id ON workflow_steps(workflow_id);
		CREATE INDEX IF NOT EXISTS idx_workflow_steps_order ON workflow_steps(workflow_id, order_index);
		CREATE INDEX IF NOT EXISTS idx_step_executions_step_id ON step_executions(step_id);
		CREATE INDEX IF NOT EXISTS idx_step_executions_workflow_id ON step_executions(workflow_id);
		CREATE INDEX IF NOT EXISTS idx_step_executions_workflow_execution_id ON step_executions(workflow_execution_id);
		CREATE INDEX IF NOT EXISTS idx_step_executions_status ON step_executions(status);
		CREATE INDEX IF NOT EXISTS idx_step_execution_logs_step_execution_id ON step_execution_logs(step_execution_id);
		CREATE INDEX IF NOT EXISTS idx_step_execution_logs_timestamp ON step_execution_logs(step_execution_id, timestamp);
	`

	if _, err := s.db.Exec(createTables); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Run migrations for schema changes
	migrations := []string{
		// Add name column to configs table if it doesn't exist
		`ALTER TABLE configs ADD COLUMN name TEXT`,
		// Remove deprecated actions column from workflows table (SQLite 3.35.0+)
		// This will fail silently on older SQLite versions or if column doesn't exist
		`ALTER TABLE workflows DROP COLUMN actions`,
		// Remove deprecated flow_graph column from workflows table (SQLite 3.35.0+)
		// This will fail silently on older SQLite versions or if column doesn't exist
		`ALTER TABLE workflows DROP COLUMN flow_graph`,
		// Add workflow_id column to delayed_action_queue if it doesn't exist
		// This handles databases created before workflow_id was added
		`ALTER TABLE delayed_action_queue ADD COLUMN workflow_id TEXT`,
		// Add execution_id column to delayed_action_queue if it doesn't exist
		`ALTER TABLE delayed_action_queue ADD COLUMN execution_id TEXT`,
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			// Ignore errors for columns that already exist or other expected errors
			errStr := err.Error()
			if !isColumnExistsError(err) &&
				!strings.Contains(errStr, "duplicate column name") &&
				!strings.Contains(errStr, "column name already exists") &&
				!strings.Contains(errStr, "no such column") {
				s.logger.Debug("Migration skipped or failed", zap.String("migration", migration), zap.Error(err))
			}
		}
	}

	s.logger.Debug("Database migrations completed")
	return nil
}

// isColumnExistsError checks if the error is due to a column already existing
func isColumnExistsError(err error) bool {
	return err != nil && (err.Error() == "duplicate column name: name" ||
		err.Error() == "column name already exists")
}

// Agent management
func (s *Storage) CreateAgent(ctx context.Context, agent *types.Agent) error {
	labelsJSON, _ := json.Marshal(agent.Labels)
	capabilitiesJSON, _ := json.Marshal(agent.Capabilities)

	query := `
		INSERT INTO agents (id, name, labels, status, last_seen, group_id, group_name, version, capabilities, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		agent.ID.String(),
		agent.Name,
		string(labelsJSON),
		string(agent.Status),
		agent.LastSeen,
		agent.GroupID,
		agent.GroupName,
		agent.Version,
		string(capabilitiesJSON),
		agent.CreatedAt,
		agent.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	s.logger.Debug("Created agent", zap.String("agent_id", agent.ID.String()))
	return nil
}

func (s *Storage) GetAgent(ctx context.Context, id uuid.UUID) (*types.Agent, error) {
	query := `
		SELECT id, name, labels, status, last_seen, group_id, group_name, version, capabilities, effective_config, created_at, updated_at
		FROM agents WHERE id = ?
	`

	var agent types.Agent
	var labelsJSON, capabilitiesJSON string
	var agentIDStr string
	var effectiveConfig sql.NullString

	err := s.db.QueryRowContext(ctx, query, id.String()).Scan(
		&agentIDStr,
		&agent.Name,
		&labelsJSON,
		&agent.Status,
		&agent.LastSeen,
		&agent.GroupID,
		&agent.GroupName,
		&agent.Version,
		&capabilitiesJSON,
		&effectiveConfig,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	agent.ID = id
	_ = json.Unmarshal([]byte(labelsJSON), &agent.Labels)
	_ = json.Unmarshal([]byte(capabilitiesJSON), &agent.Capabilities)
	if effectiveConfig.Valid {
		agent.EffectiveConfig = effectiveConfig.String
	}

	return &agent, nil
}

func (s *Storage) ListAgents(ctx context.Context) ([]*types.Agent, error) {
	query := `
		SELECT id, name, labels, status, last_seen, group_id, group_name, version, capabilities, effective_config, created_at, updated_at
		FROM agents ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*types.Agent
	for rows.Next() {
		var agent types.Agent
		var labelsJSON, capabilitiesJSON string
		var agentIDStr string
		var effectiveConfig sql.NullString

		err := rows.Scan(
			&agentIDStr,
			&agent.Name,
			&labelsJSON,
			&agent.Status,
			&agent.LastSeen,
			&agent.GroupID,
			&agent.GroupName,
			&agent.Version,
			&capabilitiesJSON,
			&effectiveConfig,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		agent.ID, _ = uuid.Parse(agentIDStr)
		_ = json.Unmarshal([]byte(labelsJSON), &agent.Labels)
		_ = json.Unmarshal([]byte(capabilitiesJSON), &agent.Capabilities)
		if effectiveConfig.Valid {
			agent.EffectiveConfig = effectiveConfig.String
		}

		agents = append(agents, &agent)
	}

	return agents, nil
}

func (s *Storage) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status types.AgentStatus) error {
	query := `UPDATE agents SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, string(status), id.String())
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", id.String())
	}

	s.logger.Debug("Updated agent status", zap.String("agent_id", id.String()), zap.String("status", string(status)))
	return nil
}

func (s *Storage) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	query := `UPDATE agents SET last_seen = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, lastSeen, id.String())
	if err != nil {
		return fmt.Errorf("failed to update agent last seen: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", id.String())
	}

	return nil
}

func (s *Storage) UpdateAgentEffectiveConfig(ctx context.Context, id uuid.UUID, effectiveConfig string) error {
	query := `UPDATE agents SET effective_config = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, effectiveConfig, id.String())
	if err != nil {
		return fmt.Errorf("failed to update agent effective config: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", id.String())
	}

	s.logger.Debug("Updated agent effective config", zap.String("agent_id", id.String()))
	return nil
}

func (s *Storage) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM agents WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", id.String())
	}

	s.logger.Debug("Deleted agent", zap.String("agent_id", id.String()))
	return nil
}

// Group management
func (s *Storage) CreateGroup(ctx context.Context, group *types.Group) error {
	labelsJSON, _ := json.Marshal(group.Labels)

	query := `
		INSERT INTO groups (id, name, labels, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		group.ID,
		group.Name,
		string(labelsJSON),
		group.CreatedAt,
		group.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	s.logger.Debug("Created group", zap.String("group_id", group.ID))
	return nil
}

func (s *Storage) GetGroup(ctx context.Context, id string) (*types.Group, error) {
	query := `SELECT id, name, labels, created_at, updated_at FROM groups WHERE id = ?`

	var group types.Group
	var labelsJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID,
		&group.Name,
		&labelsJSON,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	_ = json.Unmarshal([]byte(labelsJSON), &group.Labels)
	return &group, nil
}

func (s *Storage) ListGroups(ctx context.Context) ([]*types.Group, error) {
	query := `SELECT id, name, labels, created_at, updated_at FROM groups ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	var groups []*types.Group
	for rows.Next() {
		var group types.Group
		var labelsJSON string

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&labelsJSON,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		_ = json.Unmarshal([]byte(labelsJSON), &group.Labels)
		groups = append(groups, &group)
	}

	return groups, nil
}

func (s *Storage) DeleteGroup(ctx context.Context, id string) error {
	query := `DELETE FROM groups WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("group not found: %s", id)
	}

	s.logger.Debug("Deleted group", zap.String("group_id", id))
	return nil
}

// Config management
func (s *Storage) CreateConfig(ctx context.Context, config *types.Config) error {
	query := `
		INSERT INTO configs (id, name, agent_id, group_id, config_hash, content, version, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		config.ID,
		config.Name,
		config.AgentID,
		config.GroupID,
		config.ConfigHash,
		config.Content,
		config.Version,
		config.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	s.logger.Debug("Created config", zap.String("config_id", config.ID))
	return nil
}

func (s *Storage) GetConfig(ctx context.Context, id string) (*types.Config, error) {
	query := `SELECT id, name, agent_id, group_id, config_hash, content, version, created_at FROM configs WHERE id = ?`

	var config types.Config
	var agentIDStr, groupIDStr sql.NullString
	var nameStr sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&nameStr,
		&agentIDStr,
		&groupIDStr,
		&config.ConfigHash,
		&config.Content,
		&config.Version,
		&config.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if nameStr.Valid {
		config.Name = nameStr.String
	}
	if agentIDStr.Valid {
		agentID, _ := uuid.Parse(agentIDStr.String)
		config.AgentID = &agentID
	}
	if groupIDStr.Valid {
		config.GroupID = &groupIDStr.String
	}

	return &config, nil
}

func (s *Storage) GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*types.Config, error) {
	query := `
		SELECT id, name, agent_id, group_id, config_hash, content, version, created_at
		FROM configs
		WHERE agent_id = ?
		ORDER BY version DESC, created_at DESC
		LIMIT 1
	`

	var config types.Config
	var agentIDStr, groupIDStr sql.NullString
	var nameStr sql.NullString

	err := s.db.QueryRowContext(ctx, query, agentID.String()).Scan(
		&config.ID,
		&nameStr,
		&agentIDStr,
		&groupIDStr,
		&config.ConfigHash,
		&config.Content,
		&config.Version,
		&config.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest config for agent: %w", err)
	}

	if nameStr.Valid {
		config.Name = nameStr.String
	}
	if agentIDStr.Valid {
		agentID, _ := uuid.Parse(agentIDStr.String)
		config.AgentID = &agentID
	}
	if groupIDStr.Valid {
		config.GroupID = &groupIDStr.String
	}

	return &config, nil
}

func (s *Storage) GetLatestConfigForGroup(ctx context.Context, groupID string) (*types.Config, error) {
	query := `
		SELECT id, name, agent_id, group_id, config_hash, content, version, created_at
		FROM configs
		WHERE group_id = ?
		ORDER BY version DESC, created_at DESC
		LIMIT 1
	`

	var config types.Config
	var agentIDStr, groupIDStr sql.NullString
	var nameStr sql.NullString

	err := s.db.QueryRowContext(ctx, query, groupID).Scan(
		&config.ID,
		&nameStr,
		&agentIDStr,
		&groupIDStr,
		&config.ConfigHash,
		&config.Content,
		&config.Version,
		&config.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest config for group: %w", err)
	}

	if nameStr.Valid {
		config.Name = nameStr.String
	}
	if agentIDStr.Valid {
		agentID, _ := uuid.Parse(agentIDStr.String)
		config.AgentID = &agentID
	}
	if groupIDStr.Valid {
		config.GroupID = &groupIDStr.String
	}

	return &config, nil
}

func (s *Storage) ListConfigs(ctx context.Context, filter types.ConfigFilter) (*types.ListConfigsResult, error) {
	// Build WHERE clause for counting and querying
	whereClause := `WHERE 1=1`
	args := []interface{}{}

	if filter.AgentID != nil {
		whereClause += ` AND agent_id = ?`
		args = append(args, filter.AgentID.String())
	}

	if filter.GroupID != nil {
		whereClause += ` AND group_id = ?`
		args = append(args, *filter.GroupID)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM configs ` + whereClause
	var totalCount int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count configs: %w", err)
	}

	// Build query for data
	query := `SELECT id, name, agent_id, group_id, config_hash, content, version, created_at FROM configs ` + whereClause
	query += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	defer rows.Close()

	var configs []*types.Config
	for rows.Next() {
		var config types.Config
		var agentIDStr, groupIDStr sql.NullString
		var nameStr sql.NullString

		err := rows.Scan(
			&config.ID,
			&nameStr,
			&agentIDStr,
			&groupIDStr,
			&config.ConfigHash,
			&config.Content,
			&config.Version,
			&config.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}

		if nameStr.Valid {
			config.Name = nameStr.String
		}
		if agentIDStr.Valid {
			agentID, _ := uuid.Parse(agentIDStr.String)
			config.AgentID = &agentID
		}
		if groupIDStr.Valid {
			config.GroupID = &groupIDStr.String
		}

		configs = append(configs, &config)
	}

	return &types.ListConfigsResult{
		Configs:    configs,
		TotalCount: totalCount,
	}, nil
}

// Workflow management
func (s *Storage) CreateWorkflow(ctx context.Context, workflow *types.Workflow) error {
	var scheduleCron, scheduleTimezone sql.NullString
	if workflow.Schedule != nil {
		scheduleCron = sql.NullString{String: workflow.Schedule.CronExpression, Valid: true}
		scheduleTimezone = sql.NullString{String: workflow.Schedule.Timezone, Valid: true}
	}

	query := `
		INSERT INTO workflows (
			id, name, description, type, status, schedule_cron, schedule_timezone,
			webhook_url, webhook_secret, created_by,
			last_run, next_run, run_count, error_count, last_error, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var lastRun, nextRun sql.NullTime
	if workflow.LastRun != nil {
		lastRun = sql.NullTime{Time: *workflow.LastRun, Valid: true}
	}
	if workflow.NextRun != nil {
		nextRun = sql.NullTime{Time: *workflow.NextRun, Valid: true}
	}

	_, err := s.db.ExecContext(ctx, query,
		workflow.ID,
		workflow.Name,
		workflow.Description,
		string(workflow.Type),
		string(workflow.Status),
		scheduleCron,
		scheduleTimezone,
		workflow.WebhookURL,
		workflow.WebhookSecret,
		workflow.CreatedBy,
		lastRun,
		nextRun,
		workflow.RunCount,
		workflow.ErrorCount,
		workflow.LastError,
		workflow.CreatedAt,
		workflow.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	s.logger.Debug("Workflow created", zap.String("id", workflow.ID), zap.String("name", workflow.Name))
	return nil
}

func (s *Storage) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	query := `
		SELECT id, name, description, type, status, schedule_cron, schedule_timezone,
			webhook_url, webhook_secret, created_by,
			last_run, next_run, run_count, error_count, last_error, created_at, updated_at
		FROM workflows
		WHERE id = ?
	`

	var workflow types.Workflow
	var scheduleCron, scheduleTimezone, webhookURL, webhookSecret, createdBy, lastError sql.NullString
	var lastRun, nextRun sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&workflow.ID,
		&workflow.Name,
		&workflow.Description,
		&workflow.Type,
		&workflow.Status,
		&scheduleCron,
		&scheduleTimezone,
		&webhookURL,
		&webhookSecret,
		&createdBy,
		&lastRun,
		&nextRun,
		&workflow.RunCount,
		&workflow.ErrorCount,
		&lastError,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if scheduleCron.Valid && scheduleTimezone.Valid {
		workflow.Schedule = &types.ScheduleConfig{
			CronExpression: scheduleCron.String,
			Timezone:       scheduleTimezone.String,
		}
	}

	if webhookURL.Valid {
		workflow.WebhookURL = webhookURL.String
	}
	if webhookSecret.Valid {
		workflow.WebhookSecret = webhookSecret.String
	}
	if createdBy.Valid {
		workflow.CreatedBy = createdBy.String
	}
	if lastError.Valid {
		workflow.LastError = lastError.String
	}
	if lastRun.Valid {
		workflow.LastRun = &lastRun.Time
	}
	if nextRun.Valid {
		workflow.NextRun = &nextRun.Time
	}

	return &workflow, nil
}

func (s *Storage) ListWorkflows(ctx context.Context, filter types.WorkflowFilter) ([]*types.Workflow, error) {
	query := `
		SELECT id, name, description, type, status, schedule_cron, schedule_timezone,
			webhook_url, webhook_secret, created_by,
			last_run, next_run, run_count, error_count, last_error, created_at, updated_at
		FROM workflows
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.Type != nil {
		query += " AND type = ?"
		args = append(args, string(*filter.Type))
	}

	if filter.Status != nil {
		query += " AND status = ?"
		args = append(args, string(*filter.Status))
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*types.Workflow
	for rows.Next() {
		var workflow types.Workflow
		var scheduleCron, scheduleTimezone, webhookURL, webhookSecret, createdBy, lastError sql.NullString
		var lastRun, nextRun sql.NullTime

		err := rows.Scan(
			&workflow.ID,
			&workflow.Name,
			&workflow.Description,
			&workflow.Type,
			&workflow.Status,
			&scheduleCron,
			&scheduleTimezone,
			&webhookURL,
			&webhookSecret,
			&createdBy,
			&lastRun,
			&nextRun,
			&workflow.RunCount,
			&workflow.ErrorCount,
			&lastError,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		if scheduleCron.Valid && scheduleTimezone.Valid {
			workflow.Schedule = &types.ScheduleConfig{
				CronExpression: scheduleCron.String,
				Timezone:       scheduleTimezone.String,
			}
		}

		if webhookURL.Valid {
			workflow.WebhookURL = webhookURL.String
		}
		if webhookSecret.Valid {
			workflow.WebhookSecret = webhookSecret.String
		}
		if createdBy.Valid {
			workflow.CreatedBy = createdBy.String
		}
		if lastError.Valid {
			workflow.LastError = lastError.String
		}
		if lastRun.Valid {
			workflow.LastRun = &lastRun.Time
		}
		if nextRun.Valid {
			workflow.NextRun = &nextRun.Time
		}

		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}

func (s *Storage) UpdateWorkflow(ctx context.Context, workflow *types.Workflow) error {
	var scheduleCron, scheduleTimezone sql.NullString
	if workflow.Schedule != nil {
		scheduleCron = sql.NullString{String: workflow.Schedule.CronExpression, Valid: true}
		scheduleTimezone = sql.NullString{String: workflow.Schedule.Timezone, Valid: true}
	}

	var lastRun, nextRun sql.NullTime
	if workflow.LastRun != nil {
		lastRun = sql.NullTime{Time: *workflow.LastRun, Valid: true}
	}
	if workflow.NextRun != nil {
		nextRun = sql.NullTime{Time: *workflow.NextRun, Valid: true}
	}

	query := `
		UPDATE workflows SET
			name = ?, description = ?, type = ?, status = ?,
			schedule_cron = ?, schedule_timezone = ?,
			webhook_url = ?, webhook_secret = ?,
			created_by = ?,
			last_run = ?, next_run = ?,
			run_count = ?, error_count = ?, last_error = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query,
		workflow.Name,
		workflow.Description,
		string(workflow.Type),
		string(workflow.Status),
		scheduleCron,
		scheduleTimezone,
		workflow.WebhookURL,
		workflow.WebhookSecret,
		workflow.CreatedBy,
		lastRun,
		nextRun,
		workflow.RunCount,
		workflow.ErrorCount,
		workflow.LastError,
		workflow.UpdatedAt,
		workflow.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	s.logger.Debug("Workflow updated", zap.String("id", workflow.ID))
	return nil
}

func (s *Storage) DeleteWorkflow(ctx context.Context, id string) error {
	query := "DELETE FROM workflows WHERE id = ?"

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	s.logger.Debug("Workflow deleted", zap.String("id", id))
	return nil
}

// Workflow execution tracking
func (s *Storage) CreateWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	configsCreatedJSON, _ := json.Marshal(execution.ConfigsCreated)
	metadataJSON, _ := json.Marshal(execution.Metadata)

	var completedAt sql.NullTime
	if execution.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *execution.CompletedAt, Valid: true}
	}

	var durationMs sql.NullInt64
	if execution.DurationMs != nil {
		durationMs = sql.NullInt64{Int64: *execution.DurationMs, Valid: true}
	}

	query := `
		INSERT INTO workflow_executions (
			id, workflow_id, workflow_name, status, started_at, completed_at, duration_ms,
			actions_executed, actions_succeeded, actions_failed,
			configs_created, error, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		execution.ID,
		execution.WorkflowID,
		execution.WorkflowName,
		string(execution.Status),
		execution.StartedAt,
		completedAt,
		durationMs,
		execution.ActionsExecuted,
		execution.ActionsSucceeded,
		execution.ActionsFailed,
		string(configsCreatedJSON),
		execution.Error,
		string(metadataJSON),
		execution.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create workflow execution: %w", err)
	}

	s.logger.Debug("Workflow execution created", zap.String("id", execution.ID))
	return nil
}

func (s *Storage) GetWorkflowExecution(ctx context.Context, id string) (*types.WorkflowExecution, error) {
	query := `
		SELECT id, workflow_id, workflow_name, status, started_at, completed_at, duration_ms,
			actions_executed, actions_succeeded, actions_failed,
			configs_created, error, metadata, created_at
		FROM workflow_executions
		WHERE id = ?
	`

	var execution types.WorkflowExecution
	var configsCreatedJSON, errorStr, metadataJSON string
	var completedAt sql.NullTime
	var durationMs sql.NullInt64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&execution.ID,
		&execution.WorkflowID,
		&execution.WorkflowName,
		&execution.Status,
		&execution.StartedAt,
		&completedAt,
		&durationMs,
		&execution.ActionsExecuted,
		&execution.ActionsSucceeded,
		&execution.ActionsFailed,
		&configsCreatedJSON,
		&errorStr,
		&metadataJSON,
		&execution.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow execution: %w", err)
	}

	if err := json.Unmarshal([]byte(configsCreatedJSON), &execution.ConfigsCreated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configs_created: %w", err)
	}

	if metadataJSON != "" && metadataJSON != "null" {
		if err := json.Unmarshal([]byte(metadataJSON), &execution.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}
	if durationMs.Valid {
		execution.DurationMs = &durationMs.Int64
	}
	execution.Error = errorStr

	return &execution, nil
}

func (s *Storage) ListWorkflowExecutions(ctx context.Context, filter types.WorkflowExecutionFilter) ([]*types.WorkflowExecution, error) {
	query := `
		SELECT id, workflow_id, workflow_name, status, started_at, completed_at, duration_ms,
			actions_executed, actions_succeeded, actions_failed,
			configs_created, error, metadata, created_at
		FROM workflow_executions
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.WorkflowID != nil {
		query += " AND workflow_id = ?"
		args = append(args, *filter.WorkflowID)
	}

	if filter.Status != nil {
		query += " AND status = ?"
		args = append(args, string(*filter.Status))
	}

	query += " ORDER BY started_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow executions: %w", err)
	}
	defer rows.Close()

	var executions []*types.WorkflowExecution
	for rows.Next() {
		var execution types.WorkflowExecution
		var configsCreatedJSON, errorStr, metadataJSON string
		var completedAt sql.NullTime
		var durationMs sql.NullInt64

		err := rows.Scan(
			&execution.ID,
			&execution.WorkflowID,
			&execution.WorkflowName,
			&execution.Status,
			&execution.StartedAt,
			&completedAt,
			&durationMs,
			&execution.ActionsExecuted,
			&execution.ActionsSucceeded,
			&execution.ActionsFailed,
			&configsCreatedJSON,
			&errorStr,
			&metadataJSON,
			&execution.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow execution: %w", err)
		}

		if err := json.Unmarshal([]byte(configsCreatedJSON), &execution.ConfigsCreated); err != nil {
			return nil, fmt.Errorf("failed to unmarshal configs_created: %w", err)
		}

		if metadataJSON != "" && metadataJSON != "null" {
			if err := json.Unmarshal([]byte(metadataJSON), &execution.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if completedAt.Valid {
			execution.CompletedAt = &completedAt.Time
		}
		if durationMs.Valid {
			execution.DurationMs = &durationMs.Int64
		}
		execution.Error = errorStr

		executions = append(executions, &execution)
	}

	return executions, nil
}

func (s *Storage) UpdateWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	configsCreatedJSON, _ := json.Marshal(execution.ConfigsCreated)
	metadataJSON, _ := json.Marshal(execution.Metadata)

	var completedAt sql.NullTime
	if execution.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *execution.CompletedAt, Valid: true}
	}

	var durationMs sql.NullInt64
	if execution.DurationMs != nil {
		durationMs = sql.NullInt64{Int64: *execution.DurationMs, Valid: true}
	}

	query := `
		UPDATE workflow_executions SET
			status = ?, completed_at = ?, duration_ms = ?,
			actions_executed = ?, actions_succeeded = ?, actions_failed = ?,
			configs_created = ?, error = ?, metadata = ?
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query,
		string(execution.Status),
		completedAt,
		durationMs,
		execution.ActionsExecuted,
		execution.ActionsSucceeded,
		execution.ActionsFailed,
		string(configsCreatedJSON),
		execution.Error,
		string(metadataJSON),
		execution.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update workflow execution: %w", err)
	}

	s.logger.Debug("Trigger execution updated", zap.String("id", execution.ID))
	return nil
}

// CreateDelayedAction creates a new delayed action in the queue
func (s *Storage) CreateDelayedAction(ctx context.Context, action *types.DelayedActionQueue) error {
	actionJSON, err := json.Marshal(action.Action)
	if err != nil {
		return fmt.Errorf("failed to marshal action: %w", err)
	}

	metadataJSON, err := json.Marshal(action.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO delayed_action_queue (
			id, workflow_id, execution_id, action, scheduled_for, status, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, action.ID, action.WorkflowID, action.ExecutionID, string(actionJSON), action.ScheduledFor,
		action.Status, string(metadataJSON), action.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create delayed action: %w", err)
	}

	s.logger.Debug("Delayed action created", zap.String("id", action.ID))
	return nil
}

// GetDelayedAction retrieves a delayed action by ID
func (s *Storage) GetDelayedAction(ctx context.Context, id string) (*types.DelayedActionQueue, error) {
	var action types.DelayedActionQueue
	var actionJSON, metadataJSON string
	var completedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, workflow_id, execution_id, action, scheduled_for, status, metadata, error, created_at, completed_at
		FROM delayed_action_queue
		WHERE id = ?
	`, id).Scan(
		&action.ID, &action.WorkflowID, &action.ExecutionID, &actionJSON, &action.ScheduledFor,
		&action.Status, &metadataJSON, &action.Error, &action.CreatedAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get delayed action: %w", err)
	}

	if err := json.Unmarshal([]byte(actionJSON), &action.Action); err != nil {
		return nil, fmt.Errorf("failed to unmarshal action: %w", err)
	}

	if metadataJSON != "" && metadataJSON != "null" {
		if err := json.Unmarshal([]byte(metadataJSON), &action.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if completedAt.Valid {
		action.CompletedAt = &completedAt.Time
	}

	return &action, nil
}

// ListPendingDelayedActions retrieves all pending delayed actions that are due
func (s *Storage) ListPendingDelayedActions(ctx context.Context) ([]*types.DelayedActionQueue, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, workflow_id, execution_id, action, scheduled_for, status, metadata, error, created_at, completed_at
		FROM delayed_action_queue
		WHERE status = 'pending' AND scheduled_for <= datetime('now')
		ORDER BY scheduled_for ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending delayed actions: %w", err)
	}
	defer rows.Close()

	var actions []*types.DelayedActionQueue
	for rows.Next() {
		var action types.DelayedActionQueue
		var actionJSON, metadataJSON string
		var completedAt sql.NullTime

		if err := rows.Scan(
			&action.ID, &action.WorkflowID, &action.ExecutionID, &actionJSON, &action.ScheduledFor,
			&action.Status, &metadataJSON, &action.Error, &action.CreatedAt, &completedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan delayed action: %w", err)
		}

		if err := json.Unmarshal([]byte(actionJSON), &action.Action); err != nil {
			return nil, fmt.Errorf("failed to unmarshal action: %w", err)
		}

		if metadataJSON != "" && metadataJSON != "null" {
			if err := json.Unmarshal([]byte(metadataJSON), &action.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if completedAt.Valid {
			action.CompletedAt = &completedAt.Time
		}

		actions = append(actions, &action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate delayed actions: %w", err)
	}

	return actions, nil
}

// UpdateDelayedAction updates a delayed action
func (s *Storage) UpdateDelayedAction(ctx context.Context, action *types.DelayedActionQueue) error {
	metadataJSON, err := json.Marshal(action.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE delayed_action_queue
		SET status = ?, error = ?, completed_at = ?, metadata = ?
		WHERE id = ?
	`, action.Status, action.Error, action.CompletedAt, string(metadataJSON), action.ID)

	if err != nil {
		return fmt.Errorf("failed to update delayed action: %w", err)
	}

	s.logger.Debug("Delayed action updated", zap.String("id", action.ID))
	return nil
}

// DeleteDelayedAction deletes a delayed action
func (s *Storage) DeleteDelayedAction(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM delayed_action_queue WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete delayed action: %w", err)
	}

	s.logger.Debug("Delayed action deleted", zap.String("id", id))
	return nil
}

// Workflow triggers (normalized)
func (s *Storage) CreateWorkflowTrigger(ctx context.Context, trigger *types.WorkflowTrigger) error {
	var scheduleCron, scheduleTimezone sql.NullString
	if trigger.Schedule != nil {
		scheduleCron = sql.NullString{String: trigger.Schedule.CronExpression, Valid: true}
		scheduleTimezone = sql.NullString{String: trigger.Schedule.Timezone, Valid: true}
	}

	var webhookURL, webhookSecret sql.NullString
	if trigger.WebhookURL != "" {
		webhookURL = sql.NullString{String: trigger.WebhookURL, Valid: true}
	}
	if trigger.WebhookSecret != "" {
		webhookSecret = sql.NullString{String: trigger.WebhookSecret, Valid: true}
	}

	enabled := 0
	if trigger.Enabled {
		enabled = 1
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO workflow_triggers (
			workflow_id, type, schedule_cron, schedule_timezone,
			webhook_url, webhook_secret, enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, trigger.WorkflowID, string(trigger.Type), scheduleCron, scheduleTimezone,
		webhookURL, webhookSecret, enabled, trigger.CreatedAt, trigger.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create workflow trigger: %w", err)
	}

	s.logger.Debug("Workflow trigger created", zap.String("workflow_id", trigger.WorkflowID))
	return nil
}

func (s *Storage) GetWorkflowTrigger(ctx context.Context, workflowID string) (*types.WorkflowTrigger, error) {
	var trigger types.WorkflowTrigger
	var scheduleCron, scheduleTimezone, webhookURL, webhookSecret sql.NullString
	var enabled int

	err := s.db.QueryRowContext(ctx, `
		SELECT workflow_id, type, schedule_cron, schedule_timezone,
			webhook_url, webhook_secret, enabled, created_at, updated_at
		FROM workflow_triggers
		WHERE workflow_id = ?
	`, workflowID).Scan(
		&trigger.WorkflowID, &trigger.Type, &scheduleCron, &scheduleTimezone,
		&webhookURL, &webhookSecret, &enabled, &trigger.CreatedAt, &trigger.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow trigger: %w", err)
	}

	if scheduleCron.Valid && scheduleTimezone.Valid {
		trigger.Schedule = &types.ScheduleConfig{
			CronExpression: scheduleCron.String,
			Timezone:       scheduleTimezone.String,
		}
	}

	if webhookURL.Valid {
		trigger.WebhookURL = webhookURL.String
	}
	if webhookSecret.Valid {
		trigger.WebhookSecret = webhookSecret.String
	}
	trigger.Enabled = enabled == 1

	return &trigger, nil
}

func (s *Storage) UpdateWorkflowTrigger(ctx context.Context, trigger *types.WorkflowTrigger) error {
	var scheduleCron, scheduleTimezone sql.NullString
	if trigger.Schedule != nil {
		scheduleCron = sql.NullString{String: trigger.Schedule.CronExpression, Valid: true}
		scheduleTimezone = sql.NullString{String: trigger.Schedule.Timezone, Valid: true}
	}

	var webhookURL, webhookSecret sql.NullString
	if trigger.WebhookURL != "" {
		webhookURL = sql.NullString{String: trigger.WebhookURL, Valid: true}
	}
	if trigger.WebhookSecret != "" {
		webhookSecret = sql.NullString{String: trigger.WebhookSecret, Valid: true}
	}

	enabled := 0
	if trigger.Enabled {
		enabled = 1
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE workflow_triggers SET
			type = ?, schedule_cron = ?, schedule_timezone = ?,
			webhook_url = ?, webhook_secret = ?, enabled = ?, updated_at = ?
		WHERE workflow_id = ?
	`, string(trigger.Type), scheduleCron, scheduleTimezone,
		webhookURL, webhookSecret, enabled, trigger.UpdatedAt, trigger.WorkflowID)

	if err != nil {
		return fmt.Errorf("failed to update workflow trigger: %w", err)
	}

	s.logger.Debug("Workflow trigger updated", zap.String("workflow_id", trigger.WorkflowID))
	return nil
}

func (s *Storage) DeleteWorkflowTrigger(ctx context.Context, workflowID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_triggers WHERE workflow_id = ?`, workflowID)
	if err != nil {
		return fmt.Errorf("failed to delete workflow trigger: %w", err)
	}

	s.logger.Debug("Workflow trigger deleted", zap.String("workflow_id", workflowID))
	return nil
}

// Workflow steps (normalized)
func (s *Storage) CreateWorkflowStep(ctx context.Context, step *types.WorkflowStep) error {
	var positionX, positionY sql.NullFloat64
	if step.PositionX != nil {
		positionX = sql.NullFloat64{Float64: *step.PositionX, Valid: true}
	}
	if step.PositionY != nil {
		positionY = sql.NullFloat64{Float64: *step.PositionY, Valid: true}
	}

	dependsOnJSON, _ := json.Marshal(step.DependsOn)

	retryEnabled := 0
	if step.RetryEnabled {
		retryEnabled = 1
	}
	continueOnError := 0
	if step.ContinueOnError {
		continueOnError = 1
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO workflow_steps (
			id, workflow_id, type, name, description, order_index,
			position_x, position_y, config_json,
			retry_enabled, retry_count, retry_delay_ms, continue_on_error,
			depends_on, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, step.ID, step.WorkflowID, string(step.Type), step.Name, step.Description, step.Order,
		positionX, positionY, step.ConfigJSON,
		retryEnabled, step.RetryCount, step.RetryDelayMs, continueOnError,
		string(dependsOnJSON), step.CreatedAt, step.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create workflow step: %w", err)
	}

	s.logger.Debug("Workflow step created", zap.String("id", step.ID))
	return nil
}

func (s *Storage) GetWorkflowStep(ctx context.Context, id string) (*types.WorkflowStep, error) {
	var step types.WorkflowStep
	var stepType string
	var positionX, positionY sql.NullFloat64
	var dependsOnJSON string
	var retryEnabled, continueOnError int

	err := s.db.QueryRowContext(ctx, `
		SELECT id, workflow_id, type, name, description, order_index,
			position_x, position_y, config_json,
			retry_enabled, retry_count, retry_delay_ms, continue_on_error,
			depends_on, created_at, updated_at
		FROM workflow_steps
		WHERE id = ?
	`, id).Scan(
		&step.ID, &step.WorkflowID, &stepType, &step.Name, &step.Description, &step.Order,
		&positionX, &positionY, &step.ConfigJSON,
		&retryEnabled, &step.RetryCount, &step.RetryDelayMs, &continueOnError,
		&dependsOnJSON, &step.CreatedAt, &step.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow step: %w", err)
	}

	step.Type = types.FlowNodeType(stepType)
	if positionX.Valid {
		x := positionX.Float64
		step.PositionX = &x
	}
	if positionY.Valid {
		y := positionY.Float64
		step.PositionY = &y
	}
	step.RetryEnabled = retryEnabled == 1
	step.ContinueOnError = continueOnError == 1

	if dependsOnJSON != "" && dependsOnJSON != "null" {
		if err := json.Unmarshal([]byte(dependsOnJSON), &step.DependsOn); err != nil {
			return nil, fmt.Errorf("failed to unmarshal depends_on: %w", err)
		}
	}

	return &step, nil
}

func (s *Storage) ListWorkflowSteps(ctx context.Context, workflowID string) ([]*types.WorkflowStep, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, workflow_id, type, name, description, order_index,
			position_x, position_y, config_json,
			retry_enabled, retry_count, retry_delay_ms, continue_on_error,
			depends_on, created_at, updated_at
		FROM workflow_steps
		WHERE workflow_id = ?
		ORDER BY order_index ASC
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow steps: %w", err)
	}
	defer rows.Close()

	var steps []*types.WorkflowStep
	for rows.Next() {
		var step types.WorkflowStep
		var stepType string
		var positionX, positionY sql.NullFloat64
		var dependsOnJSON string
		var retryEnabled, continueOnError int

		if err := rows.Scan(
			&step.ID, &step.WorkflowID, &stepType, &step.Name, &step.Description, &step.Order,
			&positionX, &positionY, &step.ConfigJSON,
			&retryEnabled, &step.RetryCount, &step.RetryDelayMs, &continueOnError,
			&dependsOnJSON, &step.CreatedAt, &step.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workflow step: %w", err)
		}

		step.Type = types.FlowNodeType(stepType)
		if positionX.Valid {
			x := positionX.Float64
			step.PositionX = &x
		}
		if positionY.Valid {
			y := positionY.Float64
			step.PositionY = &y
		}
		step.RetryEnabled = retryEnabled == 1
		step.ContinueOnError = continueOnError == 1

		if dependsOnJSON != "" && dependsOnJSON != "null" {
			if err := json.Unmarshal([]byte(dependsOnJSON), &step.DependsOn); err != nil {
				return nil, fmt.Errorf("failed to unmarshal depends_on: %w", err)
			}
		}

		steps = append(steps, &step)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate workflow steps: %w", err)
	}

	return steps, nil
}

func (s *Storage) UpdateWorkflowStep(ctx context.Context, step *types.WorkflowStep) error {
	var positionX, positionY sql.NullFloat64
	if step.PositionX != nil {
		positionX = sql.NullFloat64{Float64: *step.PositionX, Valid: true}
	}
	if step.PositionY != nil {
		positionY = sql.NullFloat64{Float64: *step.PositionY, Valid: true}
	}

	dependsOnJSON, _ := json.Marshal(step.DependsOn)

	retryEnabled := 0
	if step.RetryEnabled {
		retryEnabled = 1
	}
	continueOnError := 0
	if step.ContinueOnError {
		continueOnError = 1
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE workflow_steps SET
			type = ?, name = ?, description = ?, order_index = ?,
			position_x = ?, position_y = ?, config_json = ?,
			retry_enabled = ?, retry_count = ?, retry_delay_ms = ?, continue_on_error = ?,
			depends_on = ?, updated_at = ?
		WHERE id = ?
	`, string(step.Type), step.Name, step.Description, step.Order,
		positionX, positionY, step.ConfigJSON,
		retryEnabled, step.RetryCount, step.RetryDelayMs, continueOnError,
		string(dependsOnJSON), step.UpdatedAt, step.ID)

	if err != nil {
		return fmt.Errorf("failed to update workflow step: %w", err)
	}

	s.logger.Debug("Workflow step updated", zap.String("id", step.ID))
	return nil
}

func (s *Storage) DeleteWorkflowStep(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_steps WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow step: %w", err)
	}

	s.logger.Debug("Workflow step deleted", zap.String("id", id))
	return nil
}

func (s *Storage) DeleteWorkflowSteps(ctx context.Context, workflowID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_steps WHERE workflow_id = ?`, workflowID)
	if err != nil {
		return fmt.Errorf("failed to delete workflow steps: %w", err)
	}

	s.logger.Debug("Workflow steps deleted", zap.String("workflow_id", workflowID))
	return nil
}

// Step executions
func (s *Storage) CreateStepExecution(ctx context.Context, execution *types.StepExecution) error {
	var completedAt sql.NullTime
	if execution.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *execution.CompletedAt, Valid: true}
	}

	var durationMs sql.NullInt64
	if execution.DurationMs != nil {
		durationMs = sql.NullInt64{Int64: *execution.DurationMs, Valid: true}
	}

	inputDataJSON, _ := json.Marshal(execution.InputData)
	outputDataJSON, _ := json.Marshal(execution.OutputData)
	metadataJSON, _ := json.Marshal(execution.Metadata)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO step_executions (
			id, step_id, workflow_id, workflow_execution_id,
			step_name, step_type, status, started_at, completed_at, duration_ms,
			retry_attempt, error, error_code,
			input_data, output_data, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, execution.ID, execution.StepID, execution.WorkflowID, execution.WorkflowExecutionID,
		execution.StepName, string(execution.StepType), string(execution.Status),
		execution.StartedAt, completedAt, durationMs,
		execution.RetryAttempt, execution.Error, execution.ErrorCode,
		string(inputDataJSON), string(outputDataJSON), string(metadataJSON), execution.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create step execution: %w", err)
	}

	s.logger.Debug("Step execution created", zap.String("id", execution.ID))
	return nil
}

func (s *Storage) GetStepExecution(ctx context.Context, id string) (*types.StepExecution, error) {
	var execution types.StepExecution
	var stepType, status string
	var completedAt sql.NullTime
	var durationMs sql.NullInt64
	var inputDataJSON, outputDataJSON, metadataJSON string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, step_id, workflow_id, workflow_execution_id,
			step_name, step_type, status, started_at, completed_at, duration_ms,
			retry_attempt, error, error_code,
			input_data, output_data, metadata, created_at
		FROM step_executions
		WHERE id = ?
	`, id).Scan(
		&execution.ID, &execution.StepID, &execution.WorkflowID, &execution.WorkflowExecutionID,
		&execution.StepName, &stepType, &status,
		&execution.StartedAt, &completedAt, &durationMs,
		&execution.RetryAttempt, &execution.Error, &execution.ErrorCode,
		&inputDataJSON, &outputDataJSON, &metadataJSON, &execution.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get step execution: %w", err)
	}

	execution.StepType = types.FlowNodeType(stepType)
	execution.Status = types.StepStatus(status)

	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}
	if durationMs.Valid {
		execution.DurationMs = &durationMs.Int64
	}

	if inputDataJSON != "" && inputDataJSON != "null" {
		if err := json.Unmarshal([]byte(inputDataJSON), &execution.InputData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input_data: %w", err)
		}
	}
	if outputDataJSON != "" && outputDataJSON != "null" {
		if err := json.Unmarshal([]byte(outputDataJSON), &execution.OutputData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output_data: %w", err)
		}
	}
	if metadataJSON != "" && metadataJSON != "null" {
		if err := json.Unmarshal([]byte(metadataJSON), &execution.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &execution, nil
}

func (s *Storage) ListStepExecutions(ctx context.Context, filter types.StepExecutionFilter) ([]*types.StepExecution, error) {
	query := `
		SELECT id, step_id, workflow_id, workflow_execution_id,
			step_name, step_type, status, started_at, completed_at, duration_ms,
			retry_attempt, error, error_code,
			input_data, output_data, metadata, created_at
		FROM step_executions
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.WorkflowID != nil {
		query += " AND workflow_id = ?"
		args = append(args, *filter.WorkflowID)
	}
	if filter.WorkflowExecutionID != nil {
		query += " AND workflow_execution_id = ?"
		args = append(args, *filter.WorkflowExecutionID)
	}
	if filter.StepID != nil {
		query += " AND step_id = ?"
		args = append(args, *filter.StepID)
	}
	if filter.Status != nil {
		query += " AND status = ?"
		args = append(args, string(*filter.Status))
	}

	query += " ORDER BY started_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list step executions: %w", err)
	}
	defer rows.Close()

	var executions []*types.StepExecution
	for rows.Next() {
		var execution types.StepExecution
		var stepType, status string
		var completedAt sql.NullTime
		var durationMs sql.NullInt64
		var inputDataJSON, outputDataJSON, metadataJSON string

		if err := rows.Scan(
			&execution.ID, &execution.StepID, &execution.WorkflowID, &execution.WorkflowExecutionID,
			&execution.StepName, &stepType, &status,
			&execution.StartedAt, &completedAt, &durationMs,
			&execution.RetryAttempt, &execution.Error, &execution.ErrorCode,
			&inputDataJSON, &outputDataJSON, &metadataJSON, &execution.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan step execution: %w", err)
		}

		execution.StepType = types.FlowNodeType(stepType)
		execution.Status = types.StepStatus(status)

		if completedAt.Valid {
			execution.CompletedAt = &completedAt.Time
		}
		if durationMs.Valid {
			execution.DurationMs = &durationMs.Int64
		}

		if inputDataJSON != "" && inputDataJSON != "null" {
			if err := json.Unmarshal([]byte(inputDataJSON), &execution.InputData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input_data: %w", err)
			}
		}
		if outputDataJSON != "" && outputDataJSON != "null" {
			if err := json.Unmarshal([]byte(outputDataJSON), &execution.OutputData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal output_data: %w", err)
			}
		}
		if metadataJSON != "" && metadataJSON != "null" {
			if err := json.Unmarshal([]byte(metadataJSON), &execution.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		executions = append(executions, &execution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate step executions: %w", err)
	}

	return executions, nil
}

func (s *Storage) UpdateStepExecution(ctx context.Context, execution *types.StepExecution) error {
	var completedAt sql.NullTime
	if execution.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *execution.CompletedAt, Valid: true}
	}

	var durationMs sql.NullInt64
	if execution.DurationMs != nil {
		durationMs = sql.NullInt64{Int64: *execution.DurationMs, Valid: true}
	}

	inputDataJSON, _ := json.Marshal(execution.InputData)
	outputDataJSON, _ := json.Marshal(execution.OutputData)
	metadataJSON, _ := json.Marshal(execution.Metadata)

	_, err := s.db.ExecContext(ctx, `
		UPDATE step_executions SET
			status = ?, completed_at = ?, duration_ms = ?,
			retry_attempt = ?, error = ?, error_code = ?,
			input_data = ?, output_data = ?, metadata = ?
		WHERE id = ?
	`, string(execution.Status), completedAt, durationMs,
		execution.RetryAttempt, execution.Error, execution.ErrorCode,
		string(inputDataJSON), string(outputDataJSON), string(metadataJSON), execution.ID)

	if err != nil {
		return fmt.Errorf("failed to update step execution: %w", err)
	}

	s.logger.Debug("Step execution updated", zap.String("id", execution.ID))
	return nil
}

// Step execution logs
func (s *Storage) CreateStepExecutionLog(ctx context.Context, log *types.StepExecutionLog) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO step_execution_logs (
			id, step_execution_id, level, message, data, timestamp
		) VALUES (?, ?, ?, ?, ?, ?)
	`, log.ID, log.StepExecutionID, string(log.Level), log.Message, log.Data, log.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to create step execution log: %w", err)
	}

	return nil
}

func (s *Storage) ListStepExecutionLogs(ctx context.Context, stepExecutionID string, limit int) ([]*types.StepExecutionLog, error) {
	query := `
		SELECT id, step_execution_id, level, message, data, timestamp
		FROM step_execution_logs
		WHERE step_execution_id = ?
		ORDER BY timestamp ASC
	`
	args := []interface{}{stepExecutionID}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list step execution logs: %w", err)
	}
	defer rows.Close()

	var logs []*types.StepExecutionLog
	for rows.Next() {
		var log types.StepExecutionLog
		var level string

		if err := rows.Scan(
			&log.ID, &log.StepExecutionID, &level, &log.Message, &log.Data, &log.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan step execution log: %w", err)
		}

		log.Level = types.LogLevel(level)
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate step execution logs: %w", err)
	}

	return logs, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	s.logger.Info("SQLite storage closed")
	return nil
}
