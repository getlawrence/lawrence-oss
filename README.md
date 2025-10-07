# Lawrence OSS

An open-source OpenTelemetry agent management platform with built-in observability backend.

## Overview

Lawrence OSS manages OpenTelemetry collectors via OpAMP protocol and stores their telemetry data for analysis. It runs as a single Go binary or Docker container with embedded storage.

**What it does:**
- Manages OpenTelemetry collector agents remotely via OpAMP
- Ingests telemetry (traces, metrics, logs) via OTLP
- Stores agent configurations and telemetry data
- Provides a web UI for visualization and management

**How it works:**
1. Collectors connect to Lawrence's OpAMP server for remote management
2. Collectors send their internal telemetry to Lawrence's OTLP endpoint
3. Lawrence stores telemetry in DuckDB and serves it through a REST API
4. The React UI queries the API to display agents, telemetry, and topology

## Getting Started

### Run with Docker

Pull and run the latest release:

```bash
# Pull the image
docker pull ghcr.io/getlawrence/lawrence-oss:latest

# Run Lawrence
docker run -d \
  --name lawrence \
  -p 8080:8080 \
  -p 4320:4320 \
  -p 4317:4317 \
  -p 4318:4318 \
  -v lawrence-data:/data \
  ghcr.io/getlawrence/lawrence-oss:latest

# Access the UI
open http://localhost:8080
```

### Run with Docker Compose

```bash
# Clone the repository
git clone https://github.com/getlawrence/lawrence-oss.git
cd lawrence-oss

# Start Lawrence
docker compose up -d lawrence

# Access the UI
open http://localhost:8080
```

### Connect Your Agents

Configure your OpenTelemetry collector to connect to Lawrence:

```yaml
# collector-config.yaml
extensions:
  opamp:
    server:
      ws:
        endpoint: ws://localhost:4320/v1/opamp

  health_check:
    endpoint: 0.0.0.0:13133

exporters:
  otlp:
    endpoint: localhost:4317
    tls:
      insecure: true

service:
  extensions: [opamp, health_check]

  # Export internal collector telemetry to Lawrence
  telemetry:
    metrics:
      readers:
        - periodic:
            exporter:
              otlp:
                protocol: grpc
                endpoint: localhost:4317
                tls:
                  insecure: true

    logs:
      processors:
        - batch:
            exporter:
              otlp:
                protocol: grpc
                endpoint: localhost:4317
                tls:
                  insecure: true

  pipelines:
    # Your data pipelines here
    metrics:
      receivers: [otlp]
      exporters: [otlp]
```

Start your collector:
```bash
otelcol-contrib --config collector-config.yaml
```

## Architecture

Lawrence consists of several integrated components running in a single process:

**OpAMP Server** (port 4320)
- Accepts WebSocket connections from OpenTelemetry collectors
- Distributes configurations to connected agents
- Tracks agent status and capabilities

**OTLP Receiver** (ports 4317/4318)
- Ingests traces, metrics, and logs via gRPC and HTTP
- Stores raw telemetry in DuckDB
- Runs background rollup jobs for aggregated metrics

**Storage Layer**
- SQLite: Agent metadata, groups, configurations
- DuckDB: Raw telemetry and pre-aggregated rollups

**REST API** (port 8080)
- Serves agent data, telemetry queries, and topology
- Provides configuration management endpoints
- Supports Lawrence QL query language

**Web UI**
- React-based interface for agent management
- Query builder for telemetry exploration
- Topology visualization

## License

Apache 2.0
