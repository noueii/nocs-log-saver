# CS2 Log Saver

A comprehensive log aggregation and analysis system for Counter-Strike 2 servers.

## Quick Start - ONE COMMAND

```bash
make docker-rebuild
```

That's it. Everything is set up automatically. No additional steps needed.

### Access the Application
- **Frontend**: http://localhost:6173
- **Username**: admin
- **Password**: Admin123!

## Features

- **Log Collection**: Receive and store logs from multiple CS2 servers
- **IP Whitelisting**: Dynamic IP whitelist management through admin dashboard
- **Real-time Monitoring**: View active game sessions and server status
- **Log Analysis**: Parse and analyze CS2 server logs
- **Session Tracking**: Track match sessions, rounds, and game phases
- **Web Dashboard**: Modern UI for visualization and management

## Quick Start

### Prerequisites

- Docker & Docker Compose v2
- Node.js 22 LTS (for local development)
- Go 1.23+ (for local development)
- Make (optional, for easier development)

### Setup with Make (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/noueii/nocs-log-saver.git
cd nocs-log-saver
```

2. Run initial setup:
```bash
make setup
```

3. Start development servers:
```bash
make dev
```

### Manual Setup

1. Clone the repository:
```bash
git clone https://github.com/noueii/nocs-log-saver.git
cd nocs-log-saver
```

2. Copy environment configuration:
```bash
cp .env.example .env
cp frontend/.env.local.example frontend/.env.local
```

3. Start with Docker:
```bash
docker-compose up -d
```

### Access Points

- Frontend: http://localhost:6173
- Backend API: http://localhost:9090
- Admin Panel: http://localhost:6173/admin

### CS2 Server Configuration

Add the following to your CS2 server configuration:

```
log on
logaddress_add "http://your-domain.com/logs/YOUR_SERVER_ID"
```

Replace:
- `your-domain.com` with your actual domain
- `YOUR_SERVER_ID` with a unique identifier for your server

**Important**: Make sure your CS2 server's IP is whitelisted in the admin panel.

## Development

### Using Make Commands

```bash
# Show all available commands
make help

# Start all services (database, backend, frontend)
make dev

# Start individual services
make dev-backend    # Backend only (port 9090)
make dev-frontend   # Frontend only (port 6173)
make dev-db        # Database only

# Docker operations
make docker-up      # Start with Docker Compose
make docker-down    # Stop all containers
make docker-rebuild # Rebuild containers

# Database operations
make db-shell      # Open PostgreSQL shell
make db-migrate    # Run migrations
make db-reset      # Reset database (WARNING: destroys data)

# Other useful commands
make test          # Run all tests
make lint          # Run linters
make logs          # Show all logs
make status        # Check service status
```

### Manual Development

#### Backend Development

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

### Hot Reloading (Optional)

For backend hot reloading, install Air:
```bash
go install github.com/cosmtrek/air@latest
make watch-backend
```

## Architecture

- **Backend**: Go with Gin framework, clean architecture
- **Frontend**: Next.js 15 with React 19, Tailwind CSS v4
- **Database**: PostgreSQL 17
- **Deployment**: Docker Compose, Coolify-ready

## API Endpoints

- `POST /logs/:server_id` - Receive logs from CS2 servers
- `GET /api/admin/whitelist` - Get IP whitelist
- `POST /api/admin/whitelist` - Add IP to whitelist
- `DELETE /api/admin/whitelist/:id` - Remove IP from whitelist
- `GET /api/servers` - Get connected servers
- `GET /api/logs` - Get stored logs
- `GET /api/stats` - Get system statistics

## Security

- IP-based whitelisting (CS2 servers cannot authenticate)
- Dynamic whitelist management through admin UI
- All server IPs must be explicitly whitelisted

## License

MIT