package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/storage/interfaces"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// Storage implements the AppStorage interface using SQLite
type Storage struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewStorage creates a new SQLite storage instance
func NewStorage(dbPath string, logger *zap.Logger) (*Storage, error) {
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
	var currentVersion int
	err := s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		// Table might not exist yet, run initial migration
		if _, err := s.db.Exec(InitialSchema); err != nil {
			return fmt.Errorf("failed to run initial migration: %w", err)
		}
		s.logger.Info("Applied initial schema migration")
		return nil
	}

	if currentVersion >= SchemaVersion {
		s.logger.Debug("Database schema is up to date", zap.Int("version", currentVersion))
		return nil
	}

	// Apply migrations
	for i := currentVersion; i < len(Migrations) && i < SchemaVersion; i++ {
		s.logger.Info("Applying migration", zap.Int("version", i+1))
		if _, err := s.db.Exec(Migrations[i]); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", i+1, err)
		}
	}

	return nil
}

// CreateAgent creates a new agent
func (s *Storage) CreateAgent(ctx context.Context, agent *interfaces.Agent) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert agent
	query := `
		INSERT INTO agents (
			id, instance_id_str, name, group_id, group_name, version,
			status, last_seen, started_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, query,
		agent.ID.String(),
		agent.ID.String(), // instance_id_str same as ID
		agent.Name,
		agent.GroupID,
		agent.GroupName,
		agent.Version,
		string(agent.Status),
		agent.LastSeen,
		agent.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert agent: %w", err)
	}

	// Insert capabilities
	if len(agent.Capabilities) > 0 {
		capQuery := `INSERT INTO agent_capabilities (agent_id, capability) VALUES (?, ?)`
		for _, cap := range agent.Capabilities {
			if _, err := tx.ExecContext(ctx, capQuery, agent.ID.String(), cap); err != nil {
				return fmt.Errorf("failed to insert capability: %w", err)
			}
		}
	}

	// Insert labels as attributes
	if len(agent.Labels) > 0 {
		attrQuery := `INSERT INTO agent_attributes (agent_id, key, value) VALUES (?, ?, ?)`
		for key, value := range agent.Labels {
			if _, err := tx.ExecContext(ctx, attrQuery, agent.ID.String(), key, value); err != nil {
				return fmt.Errorf("failed to insert attribute: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Created agent", zap.String("id", agent.ID.String()), zap.String("name", agent.Name))
	return nil
}

// GetAgent retrieves an agent by ID
func (s *Storage) GetAgent(ctx context.Context, id uuid.UUID) (*interfaces.Agent, error) {
	query := `
		SELECT id, name, group_id, group_name, version, status, last_seen, created_at, updated_at
		FROM agents
		WHERE id = ?
	`

	agent := &interfaces.Agent{}
	var groupID, groupName sql.NullString
	var lastSeen sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id.String()).Scan(
		&agent.ID,
		&agent.Name,
		&groupID,
		&groupName,
		&agent.Version,
		&agent.Status,
		&lastSeen,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found: %s", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if groupID.Valid {
		agent.GroupID = &groupID.String
	}
	if groupName.Valid {
		agent.GroupName = &groupName.String
	}
	if lastSeen.Valid {
		agent.LastSeen = lastSeen.Time
	}

	// Load capabilities
	caps, err := s.getAgentCapabilities(ctx, id)
	if err != nil {
		return nil, err
	}
	agent.Capabilities = caps

	// Load labels
	labels, err := s.getAgentAttributes(ctx, id)
	if err != nil {
		return nil, err
	}
	agent.Labels = labels

	return agent, nil
}

// getAgentCapabilities retrieves capabilities for an agent
func (s *Storage) getAgentCapabilities(ctx context.Context, agentID uuid.UUID) ([]string, error) {
	query := `SELECT capability FROM agent_capabilities WHERE agent_id = ?`
	rows, err := s.db.QueryContext(ctx, query, agentID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query capabilities: %w", err)
	}
	defer rows.Close()

	var capabilities []string
	for rows.Next() {
		var cap string
		if err := rows.Scan(&cap); err != nil {
			return nil, fmt.Errorf("failed to scan capability: %w", err)
		}
		capabilities = append(capabilities, cap)
	}

	return capabilities, nil
}

// getAgentAttributes retrieves attributes (labels) for an agent
func (s *Storage) getAgentAttributes(ctx context.Context, agentID uuid.UUID) (map[string]string, error) {
	query := `SELECT key, value FROM agent_attributes WHERE agent_id = ?`
	rows, err := s.db.QueryContext(ctx, query, agentID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query attributes: %w", err)
	}
	defer rows.Close()

	attributes := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}
		attributes[key] = value
	}

	return attributes, nil
}

// ListAgents retrieves all agents
func (s *Storage) ListAgents(ctx context.Context) ([]*interfaces.Agent, error) {
	query := `
		SELECT id, name, group_id, group_name, version, status, last_seen, created_at, updated_at
		FROM agents
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents: %w", err)
	}
	defer rows.Close()

	var agents []*interfaces.Agent
	for rows.Next() {
		agent := &interfaces.Agent{}
		var groupID, groupName sql.NullString
		var lastSeen sql.NullTime

		err := rows.Scan(
			&agent.ID,
			&agent.Name,
			&groupID,
			&groupName,
			&agent.Version,
			&agent.Status,
			&lastSeen,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		if groupID.Valid {
			agent.GroupID = &groupID.String
		}
		if groupName.Valid {
			agent.GroupName = &groupName.String
		}
		if lastSeen.Valid {
			agent.LastSeen = lastSeen.Time
		}

		// Load capabilities
		caps, err := s.getAgentCapabilities(ctx, agent.ID)
		if err != nil {
			return nil, err
		}
		agent.Capabilities = caps

		// Load labels
		labels, err := s.getAgentAttributes(ctx, agent.ID)
		if err != nil {
			return nil, err
		}
		agent.Labels = labels

		agents = append(agents, agent)
	}

	return agents, nil
}

// UpdateAgentStatus updates the status of an agent
func (s *Storage) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status interfaces.AgentStatus) error {
	query := `UPDATE agents SET status = ? WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, string(status), id.String())
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("agent not found: %s", id.String())
	}

	s.logger.Debug("Updated agent status", zap.String("id", id.String()), zap.String("status", string(status)))
	return nil
}

// UpdateAgentLastSeen updates the last seen timestamp of an agent
func (s *Storage) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	query := `UPDATE agents SET last_seen = ? WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, lastSeen, id.String())
	if err != nil {
		return fmt.Errorf("failed to update agent last seen: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("agent not found: %s", id.String())
	}

	return nil
}

// DeleteAgent deletes an agent
func (s *Storage) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM agents WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("agent not found: %s", id.String())
	}

	s.logger.Debug("Deleted agent", zap.String("id", id.String()))
	return nil
}

// CreateGroup creates a new group
func (s *Storage) CreateGroup(ctx context.Context, group *interfaces.Group) error {
	// Serialize labels to JSON
	labelsJSON, err := json.Marshal(group.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	query := `
		INSERT INTO groups (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query,
		group.ID,
		group.Name,
		string(labelsJSON), // Store labels in description field as JSON
		group.CreatedAt,
		group.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert group: %w", err)
	}

	s.logger.Debug("Created group", zap.String("id", group.ID), zap.String("name", group.Name))
	return nil
}

// GetGroup retrieves a group by ID
func (s *Storage) GetGroup(ctx context.Context, id string) (*interfaces.Group, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM groups WHERE id = ?`

	group := &interfaces.Group{}
	var description sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID,
		&group.Name,
		&description,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Parse labels from description field
	if description.Valid && description.String != "" {
		if err := json.Unmarshal([]byte(description.String), &group.Labels); err != nil {
			s.logger.Warn("Failed to unmarshal group labels", zap.Error(err))
			group.Labels = make(map[string]string)
		}
	} else {
		group.Labels = make(map[string]string)
	}

	return group, nil
}

// ListGroups retrieves all groups
func (s *Storage) ListGroups(ctx context.Context) ([]*interfaces.Group, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM groups ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []*interfaces.Group
	for rows.Next() {
		group := &interfaces.Group{}
		var description sql.NullString

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&description,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		// Parse labels
		if description.Valid && description.String != "" {
			if err := json.Unmarshal([]byte(description.String), &group.Labels); err != nil {
				s.logger.Warn("Failed to unmarshal group labels", zap.Error(err))
				group.Labels = make(map[string]string)
			}
		} else {
			group.Labels = make(map[string]string)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// DeleteGroup deletes a group
func (s *Storage) DeleteGroup(ctx context.Context, id string) error {
	query := `DELETE FROM groups WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("group not found: %s", id)
	}

	s.logger.Debug("Deleted group", zap.String("id", id))
	return nil
}

// CreateConfig creates a new configuration
func (s *Storage) CreateConfig(ctx context.Context, config *interfaces.Config) error {
	var configType string
	var agentID, groupID sql.NullString

	if config.AgentID != nil {
		configType = "agent"
		agentID.String = config.AgentID.String()
		agentID.Valid = true
	} else if config.GroupID != nil {
		configType = "group"
		groupID.String = *config.GroupID
		groupID.Valid = true
	} else {
		return fmt.Errorf("config must have either agent_id or group_id")
	}

	query := `
		INSERT INTO configs (id, agent_id, group_id, config_type, config_body, version, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		config.ID,
		agentID,
		groupID,
		configType,
		config.Content,
		config.Version,
		config.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert config: %w", err)
	}

	s.logger.Debug("Created config", zap.String("id", config.ID), zap.String("type", configType))
	return nil
}

// GetConfig retrieves a configuration by ID
func (s *Storage) GetConfig(ctx context.Context, id string) (*interfaces.Config, error) {
	query := `SELECT id, agent_id, group_id, config_body, version, created_at FROM configs WHERE id = ?`

	config := &interfaces.Config{}
	var agentID, groupID sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&agentID,
		&groupID,
		&config.Content,
		&config.Version,
		&config.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if agentID.Valid {
		parsedID, err := uuid.Parse(agentID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse agent ID: %w", err)
		}
		config.AgentID = &parsedID
	}
	if groupID.Valid {
		config.GroupID = &groupID.String
	}

	return config, nil
}

// GetLatestConfigForAgent retrieves the latest configuration for an agent
func (s *Storage) GetLatestConfigForAgent(ctx context.Context, agentID uuid.UUID) (*interfaces.Config, error) {
	query := `
		SELECT id, agent_id, group_id, config_body, version, created_at
		FROM configs
		WHERE agent_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	config := &interfaces.Config{}
	var agentIDStr, groupID sql.NullString

	err := s.db.QueryRowContext(ctx, query, agentID.String()).Scan(
		&config.ID,
		&agentIDStr,
		&groupID,
		&config.Content,
		&config.Version,
		&config.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No config found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest config: %w", err)
	}

	if agentIDStr.Valid {
		parsedID, err := uuid.Parse(agentIDStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse agent ID: %w", err)
		}
		config.AgentID = &parsedID
	}
	if groupID.Valid {
		config.GroupID = &groupID.String
	}

	return config, nil
}

// GetLatestConfigForGroup retrieves the latest configuration for a group
func (s *Storage) GetLatestConfigForGroup(ctx context.Context, groupID string) (*interfaces.Config, error) {
	query := `
		SELECT id, agent_id, group_id, config_body, version, created_at
		FROM configs
		WHERE group_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	config := &interfaces.Config{}
	var agentID, groupIDStr sql.NullString

	err := s.db.QueryRowContext(ctx, query, groupID).Scan(
		&config.ID,
		&agentID,
		&groupIDStr,
		&config.Content,
		&config.Version,
		&config.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No config found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest config: %w", err)
	}

	if agentID.Valid {
		parsedID, err := uuid.Parse(agentID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse agent ID: %w", err)
		}
		config.AgentID = &parsedID
	}
	if groupIDStr.Valid {
		config.GroupID = &groupIDStr.String
	}

	return config, nil
}

// ListConfigs retrieves configurations based on filter
func (s *Storage) ListConfigs(ctx context.Context, filter interfaces.ConfigFilter) ([]*interfaces.Config, error) {
	query := `
		SELECT id, agent_id, group_id, config_body, version, created_at
		FROM configs
		WHERE 1=1
	`
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
		return nil, fmt.Errorf("failed to query configs: %w", err)
	}
	defer rows.Close()

	var configs []*interfaces.Config
	for rows.Next() {
		config := &interfaces.Config{}
		var agentID, groupID sql.NullString

		err := rows.Scan(
			&config.ID,
			&agentID,
			&groupID,
			&config.Content,
			&config.Version,
			&config.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}

		if agentID.Valid {
			parsedID, err := uuid.Parse(agentID.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse agent ID: %w", err)
			}
			config.AgentID = &parsedID
		}
		if groupID.Valid {
			config.GroupID = &groupID.String
		}

		configs = append(configs, config)
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
