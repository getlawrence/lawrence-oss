# Lawrence OSS Web Interface

A modern React-based web interface for the Lawrence OSS observability platform.

## Features

- **Agent Inventory**: View and manage OpenTelemetry agents
- **Group Management**: Organize agents into groups
- **Configuration Management**: Create and manage agent configurations
- **Telemetry Explorer**: Explore metrics, logs, and traces (coming soon)

## Development

### Prerequisites

- Node.js 18+
- pnpm (recommended) or npm

### Setup

1. Install dependencies:

   ```bash
   pnpm install
   ```

2. Start the development server:

   ```bash
   pnpm dev
   ```

3. Open http://localhost:5173 in your browser

### Environment Variables

Create a `.env` file in the web directory:

```env
VITE_BACKEND_URL=http://localhost:8080
VITE_APP_BASE_URL=http://localhost:5173
```

## API Integration

The web interface connects to the Lawrence OSS backend API running on port 8080. Make sure the backend is running before starting the web interface.

### Available Endpoints

- `GET /api/v1/agents` - List all agents
- `GET /api/v1/agents/stats` - Get agent statistics
- `GET /api/v1/groups` - List all groups
- `POST /api/v1/groups` - Create new group
- `GET /api/v1/configs` - List configurations
- `POST /api/v1/configs` - Create new configuration

## Build

To build for production:

```bash
pnpm build
```

The built files will be in the `dist/` directory.

## Tech Stack

- **React 19** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Styling
- **Radix UI** - Component library
- **SWR** - Data fetching
- **React Router** - Navigation
