# OpAMP Server - OSS Implementation

This package provides the OpAMP (Open Agent Management Protocol) server implementation for Lawrence OSS.

## Overview

The OpAMP server manages connections from OpenTelemetry Collector agents, handles configuration distribution, and tracks agent status.

## Architecture

```
┌─────────────────────────────────────┐
│         OpAMP Server                │
│                                     │
│  ┌──────────┐     ┌──────────────┐ │
│  │  Server  │────▶│   Agents     │ │
│  │          │     │   Registry   │ │
│  └──────────┘     └──────────────┘ │
│       │                    │        │
│       ▼                    ▼        │
│  ┌──────────┐     ┌──────────────┐ │
│  │Connection│     │    Agent     │ │
│  │ Handler  │────▶│   Instance   │ │
│  └──────────┘     └──────────────┘ │
│       │                    │        │
│       ▼                    ▼        │
│  ┌──────────┐     ┌──────────────┐ │
│  │  Status  │     │    Config    │ │
│  │ Updates  │     │ Distribution │ │
│  └──────────┘     └──────────────┘ │
└─────────────────────────────────────┘
```

## Files

### Core Components

1. **server.go** (271 lines)
   - Main OpAMP server
   - WebSocket connection handling
   - Message routing
   - Group management (simplified for OSS)
   - **Removed:** Authentication, backend API client, multi-tenancy

2. **agents.go** (152 lines)
   - Agent registry
   - Connection tracking
   - Agent lifecycle management
   - **Removed:** Organization-based filtering, Redis caching

### Agent Management

3. **agent_core.go** (140 lines)
   - Agent data structure
   - Core agent fields
   - Thread-safe operations
   - **Removed:** OrganizationID, UserID fields

4. **agent_status.go** (214 lines)
   - Status update processing
   - Health monitoring
   - Effective config tracking
   - **100% reusable** - No changes from SaaS version

5. **agent_config.go** (128 lines)
   - Config calculation
   - Config distribution
   - Config change detection
   - **100% reusable** - No changes from SaaS version

6. **agent_connection.go** (74 lines)
   - Message sending
   - Connection settings
   - Certificate handling
   - **100% reusable** - No changes from SaaS version

**Total:** 979 lines of production-tested code

## Key Features

### ✅ What's Included

- **Agent Registration**: Automatic agent discovery and registration
- **Config Distribution**: Push configs to agents via OpAMP
- **Status Monitoring**: Real-time agent health and status tracking
- **Group Support**: Agents can be organized into groups
- **TLS Support**: Optional TLS with client certificates
- **Reconnection Handling**: Automatic reconnection support

### ❌ What's Removed (vs SaaS)

- **Multi-Tenancy**: No organization-based separation
- **Authentication**: No JWT/token validation
- **Backend API**: No external backend calls for group resolution
- **Redis Caching**: In-memory only
- **Advanced Group Resolution**: Simplified local-only group management

## Usage Example

```go
package main

import (
    "context"
    "log"

    "github.com/getlawrence/lawrence-oss/internal/opamp"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()

    // Create agent registry
    agents := opamp.NewAgents(logger)

    // Create OpAMP server
    server, err := opamp.NewServer(agents, logger)
    if err != nil {
        log.Fatal(err)
    }

    // Start server on port 4320
    if err := server.Start(4320); err != nil {
        log.Fatal(err)
    }

    // ... run your application ...

    // Graceful shutdown
    server.Stop(context.Background())
}
```

## Configuration Management

### Default Config

The server provides a default OpenTelemetry Collector configuration:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  otlp:
    endpoint: localhost:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```

### Custom Configs

You can set custom configs per agent:

```go
// Get agent
agent := server.ListAgents()[agentID]

// Update config
err := server.UpdateConfig(agentID, customConfig, notifyChan)
```

## Group Management

### Group Attributes

Agents can specify group membership via OpAMP attributes:

```yaml
# In agent description
identifying_attributes:
  - key: "group.id"
    value: "production-collectors"
  - key: "group.name"
    value: "Production"
```

### Extracting Group Info

The server automatically extracts group information from:
- `group.id` or `service.group.id`
- `group.name` or `service.group.name`

## API Methods

### Server Methods

```go
// Start the OpAMP server
Start(port int) error

// Stop the server
Stop(ctx context.Context) error

// Get effective config for an agent
GetEffectiveConfig(agentId uuid.UUID) (string, error)

// Update agent config
UpdateConfig(agentId uuid.UUID, config map[string]interface{},
             notifyNextStatusUpdate chan<- struct{}) error

// List all agents
ListAgents() map[uuid.UUID]*Agent
```

### Agent Methods

```go
// Send message to agent
SendToAgent(msg *protobufs.ServerToAgent)

// Update agent status
UpdateStatus(statusMsg *protobufs.AgentToServer,
             response *protobufs.ServerToAgent)

// Set custom config
SetCustomConfig(config *protobufs.AgentConfigMap,
                notifyWhenConfigIsApplied chan<- struct{})

// Get clone for safe read access
CloneReadonly() *Agent
```

## Thread Safety

All agent operations are thread-safe:
- Agent registry uses `sync.RWMutex`
- Individual agents have their own `sync.RWMutex`
- Read-only clones for safe concurrent access

## Logging

Uses structured logging with zap:

```go
s.logger.Info("Agent joined group",
    zap.String("agentId", agent.InstanceIdStr),
    zap.String("groupId", groupID),
    zap.String("groupName", groupName))
```

## Testing

See `../../../test/opamp/` for integration tests.

## Future Enhancements

1. **Config Persistence**: Store configs in SQLite
2. **Group Configs**: Group-level config management
3. **Config Versioning**: Track config history
4. **Package Management**: OpAMP package distribution
5. **Advanced Health Checks**: Custom health probes

## Differences from SaaS Version

| Feature | SaaS | OSS |
|---------|------|-----|
| Multi-tenancy | ✅ | ❌ |
| Authentication | JWT/Tokens | None |
| Group Resolution | Backend API + Redis | Local extraction |
| Config Storage | PostgreSQL | In-memory (future: SQLite) |
| Organization Isolation | ✅ | ❌ |
| Backend Client | ✅ | ❌ |

## Dependencies

```go
require (
    github.com/google/uuid v1.6.0
    github.com/open-telemetry/opamp-go v0.16.0
    go.uber.org/zap v1.27.0
)
```

## License

Apache 2.0
