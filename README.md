# Lawrence All-in-One OSS

A lightweight, self-contained observability platform for OpenTelemetry. Run everything in a single binary or Docker container.

## Features

- 🚀 **Single Binary Deployment** - Everything in one executable
- 📊 **OTLP Ingestion** - Native support for traces, metrics, and logs
- 🔧 **OpAMP Agent Management** - Manage and configure OpenTelemetry collectors
- 💾 **Embedded Storage** - SQLite for app data, DuckDB for telemetry
- 📈 **Pre-aggregated Rollups** - Fast queries for historical data
- 🎨 **Built-in UI** - React-based web interface
- 🐳 **Docker Ready** - Single container deployment

## Quick Start

### Using Docker

```bash
docker run -p 8080:8080 -p 4317:4317 -p 4318:4318 \
  -v $(pwd)/data:/data \
  lawrence/all-in-one:latest
```

### Using Binary

```bash
# Download latest release
curl -LO https://github.com/getlawrence/lawrence-oss/releases/latest/download/lawrence-oss

# Run
./lawrence-oss

# Or with custom config
./lawrence-oss --config ./lawrence.yaml
```

### Building from Source

```bash
# Build UI
make ui

# Build binary
make build

# Run
./bin/lawrence
```

## Architecture

```
┌─────────────────────────────┐
│ Lawrence All-in-One         │
│-----------------------------│
│ Go single process           │
│                             │
│ ┌───────────┐  ┌──────────┐ │
│ │ OTLP      │  │ OpAMP    │ │
│ │ Collector │  │ Server   │ │
│ └───────────┘  └──────────┘ │
│        │           │         │
│ ┌──────────────────────────┐ │
│ │ Storage                  │ │
│ │  - AppStorage (SQLite)   │ │
│ │  - TelemetryStorage      │ │
│ │    (DuckDB)              │ │
│ └──────────────────────────┘ │
│        │                     │
│ ┌───────────┐                │
│ │ Backend   │                │
│ │ API       │                │
│ └───────────┘                │
│        │                      │
│ ┌───────────┐                 │
│ │ Web UI    │                 │
│ │ (React)   │                 │
│ └───────────┘                 │
└─────────────────────────────┘
```

## Configuration

Create a `lawrence.yaml` configuration file:

```yaml
server:
  http_port: 8080
  opamp_port: 4320

otlp:
  grpc_endpoint: 0.0.0.0:4317
  http_endpoint: 0.0.0.0:4318

storage:
  app:
    type: sqlite
    path: ./data/app.db
  telemetry:
    type: duckdb
    path: ./data/telemetry.db

retention:
  raw_metrics: 24h
  raw_logs: 24h
  rollups_1m: 7d
  rollups_5m: 30d

rollups:
  enabled: true
  interval_1m: "*/1 * * * *"
  interval_5m: "*/5 * * * *"
```

## API Endpoints

### Agents
- `GET /api/v1/agents` - List all agents
- `GET /api/v1/agents/:id` - Get agent details

### Configs
- `GET /api/v1/configs` - List configs
- `GET /api/v1/configs/:id` - Get config
- `POST /api/v1/configs` - Create config

### Telemetry
- `POST /api/v1/metrics/query` - Query metrics
- `POST /api/v1/logs/query` - Query logs
- `GET /api/v1/metrics/rollups` - Get rollups

## Development

### Prerequisites
- Go 1.23+
- Node.js 20+
- Make

### Commands

```bash
# Install dependencies
make deps

# Build UI
make ui

# Build Go binary
make build

# Run locally
make run

# Run tests
make test

# Build Docker image
make docker

# Clean build artifacts
make clean
```

## Storage

### SQLite (App Data)
- Agents
- Groups
- Configs
- Config history

### DuckDB (Telemetry Data)
- Raw metrics/logs (recent)
- Pre-aggregated rollups
- Agent-level rollups
- Group-level rollups

## License

Apache 2.0

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)

## Support

- 📖 [Documentation](https://docs.getlawrence.com)
- 💬 [Discord Community](https://discord.gg/lawrence)
- 🐛 [Issue Tracker](https://github.com/getlawrence/lawrence-oss/issues)
