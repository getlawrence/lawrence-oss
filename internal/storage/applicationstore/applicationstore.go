package applicationstore

// Re-export types from the types package for convenience
import "github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/types"

// Type aliases for convenience
type ApplicationStore = types.ApplicationStore
type Agent = types.Agent
type AgentStatus = types.AgentStatus
type Group = types.Group
type Config = types.Config
type ConfigFilter = types.ConfigFilter

// Re-export constants
const (
	AgentStatusOnline  = types.AgentStatusOnline
	AgentStatusOffline = types.AgentStatusOffline
	AgentStatusError   = types.AgentStatusError
)
