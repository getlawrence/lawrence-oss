package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore"
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
func NewSQLiteStorage(dbPath string, logger *zap.Logger) (applicationstore.ApplicationStore, error) {
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
	// Check current schema version
	var version int
	err := s.db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

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
	`

	if _, err := s.db.Exec(createTables); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Update schema version
	if _, err := s.db.Exec("PRAGMA user_version = 1"); err != nil {
		return fmt.Errorf("failed to update schema version: %w", err)
	}

	s.logger.Debug("Database migrations completed")
	return nil
}

// Agent management
func (s *Storage) CreateAgent(ctx context.Context, agent *applicationstore.Agent) error {
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

func (s *Storage) GetAgent(ctx context.Context, id uuid.UUID) (*applicationstore.Agent, error) {
	query := `
		SELECT id, name, labels, status, last_seen, group_id, group_name, version, capabilities, created_at, updated_at
		FROM agents WHERE id = ?
	`

	var agent applicationstore.Agent
	var labelsJSON, capabilitiesJSON string
	var agentIDStr string

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

	return &agent, nil
}

func (s *Storage) ListAgents(ctx context.Context) ([]*applicationstore.Agent, error) {
	query := `
		SELECT id, name, labels, status, last_seen, group_id, group_name, version, capabilities, created_at, updated_at
		FROM agents ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*applicationstore.Agent
	for rows.Next() {
		var agent applicationstore.Agent
		var labelsJSON, capabilitiesJSON string
		var agentIDStr string

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
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		agent.ID, _ = uuid.Parse(agentIDStr)
		_ = json.Unmarshal([]byte(labelsJSON), &agent.Labels)
		_ = json.Unmarshal([]byte(capabilitiesJSON), &agent.Capabilities)

		agents = append(agents, &agent)
	}

	return agents, nil
}

func (s *Storage) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status applicationstore.AgentStatus) error {
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
func (s *Storage) CreateGroup(ctx context.Context, group *applicationstore.Group) error {
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

func (s *Storage) GetGroup(ctx context.Context, id string) (*applicationstore.Group, error) {
	query := `SELECT id, name, labels, created_at, updated_at FROM groups WHERE id = ?`

	var group applicationstore.Group
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

func (s *Storage) ListGroups(ctx context.Context) ([]*applicationstore.Group, error) {
	query := `SELECT id, name, labels, created_at, updated_at FROM groups ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	var groups []*applicationstore.Group
	for rows.Next() {
		var group applicationstore.Group
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
func (s *Storage) CreateConfig(ctx context.Context, config *applicationstore.Config) error {
	query := `
		INSERT INTO configs (id, agent_id, group_id, config_hash, content, version, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		config.ID,
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

func (s *Storage) GetConfig(ctx context.Context, id string) (*applicationstore.Config, error) {
	query := `SELECT id, agent_id, group_id, config_hash, content, version, created_at FROM configs WHERE id = ?`

	var config applicationstore.Config
	var agentIDStr, groupIDStr sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
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

	if agentIDStr.Valid {
		agentID, _ := uuid.Parse(agentIDStr.String)
		config.AgentID = &agentID
	}
	if groupIDStr.Valid {
		config.GroupID = &groupIDStr.String
	}

	return &config, nil
}

func (s *Storage) GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*applicationstore.Config, error) {
	query := `
		SELECT id, agent_id, group_id, config_hash, content, version, created_at
		FROM configs
		WHERE agent_id = ?
		ORDER BY version DESC, created_at DESC
		LIMIT 1
	`

	var config applicationstore.Config
	var agentIDStr, groupIDStr sql.NullString

	err := s.db.QueryRowContext(ctx, query, agentID.String()).Scan(
		&config.ID,
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

	if agentIDStr.Valid {
		agentID, _ := uuid.Parse(agentIDStr.String)
		config.AgentID = &agentID
	}
	if groupIDStr.Valid {
		config.GroupID = &groupIDStr.String
	}

	return &config, nil
}

func (s *Storage) GetLatestConfigForGroup(ctx context.Context, groupID string) (*applicationstore.Config, error) {
	query := `
		SELECT id, agent_id, group_id, config_hash, content, version, created_at
		FROM configs
		WHERE group_id = ?
		ORDER BY version DESC, created_at DESC
		LIMIT 1
	`

	var config applicationstore.Config
	var agentIDStr, groupIDStr sql.NullString

	err := s.db.QueryRowContext(ctx, query, groupID).Scan(
		&config.ID,
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

	if agentIDStr.Valid {
		agentID, _ := uuid.Parse(agentIDStr.String)
		config.AgentID = &agentID
	}
	if groupIDStr.Valid {
		config.GroupID = &groupIDStr.String
	}

	return &config, nil
}

func (s *Storage) ListConfigs(ctx context.Context, filter applicationstore.ConfigFilter) ([]*applicationstore.Config, error) {
	query := `SELECT id, agent_id, group_id, config_hash, content, version, created_at FROM configs WHERE 1=1`
	args := []interface{}{}

	if filter.AgentID != nil {
		query += ` AND agent_id = ?`
		args = append(args, filter.AgentID.String())
	}

	if filter.GroupID != nil {
		query += ` AND group_id = ?`
		args = append(args, *filter.GroupID)
	}

	query += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	defer rows.Close()

	var configs []*applicationstore.Config
	for rows.Next() {
		var config applicationstore.Config
		var agentIDStr, groupIDStr sql.NullString

		err := rows.Scan(
			&config.ID,
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

		if agentIDStr.Valid {
			agentID, _ := uuid.Parse(agentIDStr.String)
			config.AgentID = &agentID
		}
		if groupIDStr.Valid {
			config.GroupID = &groupIDStr.String
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	s.logger.Info("SQLite storage closed")
	return nil
}
