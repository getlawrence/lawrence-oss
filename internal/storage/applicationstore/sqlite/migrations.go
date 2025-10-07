package sqlite

const SchemaVersion = 1

// InitialSchema creates the initial SQLite database schema
const InitialSchema = `
-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER PRIMARY KEY,
	applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Agents table
CREATE TABLE IF NOT EXISTS agents (
	id TEXT PRIMARY KEY,
	instance_id_str TEXT NOT NULL,
	name TEXT,
	group_id TEXT,
	group_name TEXT,
	version TEXT,
	status TEXT NOT NULL DEFAULT 'offline',
	last_seen DATETIME,
	started_at DATETIME NOT NULL,
	effective_config TEXT,
	custom_config TEXT,
	remote_config_status TEXT,
	health_status TEXT,
	error_message TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (group_id) REFERENCES groups(id)
);

-- Groups table
CREATE TABLE IF NOT EXISTS groups (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	config TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Configs table (config history for agents and groups)
CREATE TABLE IF NOT EXISTS configs (
	id TEXT PRIMARY KEY,
	agent_id TEXT,
	group_id TEXT,
	config_type TEXT NOT NULL, -- 'agent' or 'group'
	config_body TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	created_by TEXT,
	FOREIGN KEY (agent_id) REFERENCES agents(id),
	FOREIGN KEY (group_id) REFERENCES groups(id),
	CHECK (
		(config_type = 'agent' AND agent_id IS NOT NULL AND group_id IS NULL) OR
		(config_type = 'group' AND group_id IS NOT NULL AND agent_id IS NULL)
	)
);

-- Agent capabilities table
CREATE TABLE IF NOT EXISTS agent_capabilities (
	agent_id TEXT NOT NULL,
	capability TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (agent_id, capability),
	FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Agent attributes table
CREATE TABLE IF NOT EXISTS agent_attributes (
	agent_id TEXT NOT NULL,
	key TEXT NOT NULL,
	value TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (agent_id, key),
	FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_agents_group_id ON agents(group_id);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_last_seen ON agents(last_seen);
CREATE INDEX IF NOT EXISTS idx_configs_agent_id ON configs(agent_id);
CREATE INDEX IF NOT EXISTS idx_configs_group_id ON configs(group_id);
CREATE INDEX IF NOT EXISTS idx_configs_created_at ON configs(created_at DESC);

-- Triggers to update updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_agents_timestamp
AFTER UPDATE ON agents
BEGIN
	UPDATE agents SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_groups_timestamp
AFTER UPDATE ON groups
BEGIN
	UPDATE groups SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_agent_attributes_timestamp
AFTER UPDATE ON agent_attributes
BEGIN
	UPDATE agent_attributes SET updated_at = CURRENT_TIMESTAMP
	WHERE agent_id = NEW.agent_id AND key = NEW.key;
END;

-- Insert initial schema version
INSERT OR IGNORE INTO schema_version (version) VALUES (1);
`

// Migrations is a list of all schema migrations
var Migrations = []string{
	InitialSchema,
}
