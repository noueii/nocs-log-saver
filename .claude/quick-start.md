# Quick Start Guide - CS2 Log Saver

## Get Running in 15 Minutes

### Prerequisites
- Go 1.23+ (latest stable)
- Node.js 22+ (LTS version)
- Docker & Docker Compose v2
- Git

### 1. Clone and Setup (2 min)

```bash
# Clone the repo (once it exists)
git clone https://github.com/noueii/nocs-log-saver.git
cd nocs-log-saver

# Copy environment variables
cp .env.example .env
```

### 2. Configure Environment (2 min)

Edit `.env` file:
```bash
# Database
DATABASE_URL=postgresql://cs2admin:localpass123@localhost:5432/cs2logs

# IP Whitelist (initial setup only - manage through admin UI after first run)
# Used as fallback if database is empty
ALLOWED_IPS=127.0.0.1,::1,localhost

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080

# Ports
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

### 3. Start Everything with Docker (3 min)

```bash
# Start all services (Docker Compose v2)
docker compose up -d

# Check if everything is running
docker compose ps

# View logs
docker compose logs -f
```

Services will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

### 4. Initialize Database (2 min)

```bash
# Run migrations (first time only)
docker compose exec backend /app/migrate up

# Or if developing locally without Docker:
cd backend
go run cmd/migrate/main.go up
```

### 5. Test Log Ingestion (2 min)

Send a test log to verify everything works:

```bash
# Send test log (from localhost, which is whitelisted)
curl -X POST http://localhost:8080/logs/testserver \
  -H "Content-Type: text/plain" \
  -d 'L 01/17/2025 - 12:00:00: "Player<1><STEAM_1:0:123456><CT>" killed "Enemy<2><STEAM_1:0:654321><T>" with "ak47"'

# Check if log was received
curl http://localhost:8080/api/logs?server_id=testserver
```

### 6. Access the UI (1 min)

Open http://localhost:3000 in your browser:
- Dashboard shows overview stats
- Servers page lists connected servers
- Logs page shows raw and parsed logs
- Sessions page shows active/completed matches
- Admin page for managing IP whitelist (add/remove/edit allowed IPs)

## Development Setup (Without Docker)

### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Run database migrations
go run cmd/migrate/main.go up

# Run the server
go run cmd/server/main.go

# Or use air for hot reload
go install github.com/cosmtrek/air@latest
air
```

### Frontend Development

```bash
cd frontend

# Install dependencies (Node.js 22 LTS)
npm install

# Run development server with Turbopack
npm run dev --turbo

# Build for production
npm run build
```

### Database Only

```bash
# Start just PostgreSQL 17
docker compose up -d postgres

# Connect with psql
docker compose exec postgres psql -U cs2admin -d cs2logs
```

## Project Structure Overview

```
.
├── backend/              # Go backend service
│   ├── cmd/server/      # Main application
│   └── internal/        # Core logic
├── frontend/            # Next.js UI
│   ├── app/            # App router pages
│   └── components/     # React components
├── docker-compose.yml   # Local development
└── .env                # Environment config
```

## Common Tasks

### View Logs
```bash
# Backend logs
docker compose logs -f backend

# Frontend logs
docker compose logs -f frontend

# Database logs
docker compose logs -f postgres
```

### Reset Database
```bash
# Stop services
docker compose down

# Remove volumes (deletes all data)
docker compose down -v

# Start fresh
docker compose up -d
```

### Run Tests
```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test
```

### Build for Production
```bash
# Build images
docker compose build

# Or build individually
docker build -t cs2-backend ./backend
docker build -t cs2-frontend ./frontend
```

## Configure CS2 Server

Add to your CS2 server config:

```cfg
# Enable logging
log on

# Send logs to your backend (adjust URL and server ID)
logaddress_add_http "http://localhost:8080/logs/myserver"

# Note: CS2 servers don't support authentication headers
# Make sure to add your server's IP to ALLOWED_IPS environment variable
```

Or use webhook/HTTP log forwarder:
```bash
# Example with a log forwarder script
tail -f /path/to/cs2/logs/server.log | while read line; do
  curl -X POST http://localhost:8080/logs/server1 \
    -H "Content-Type: text/plain" \
    -d "$line"
done
```

## Quick Troubleshooting

### Backend won't start
```bash
# Check port 8080 is free
lsof -i :8080

# Check database connection
docker-compose exec backend ping postgres
```

### Frontend won't connect to backend
```bash
# Check NEXT_PUBLIC_API_URL in .env
# Make sure backend is running
curl http://localhost:8080/health
```

### Database connection issues
```bash
# Check PostgreSQL is running
docker compose ps postgres

# Test connection
docker compose exec postgres pg_isready

# View PostgreSQL logs
docker compose logs postgres
```

### Logs not parsing
```bash
# Check parser logs
docker compose logs backend | grep parser

# Verify cs2-log library is installed
cd backend && go list -m all | grep cs2-log
```

### Logs rejected (403 Forbidden)
```bash
# Check if IP is whitelisted
echo $ALLOWED_IPS

# Find your server's IP
# From CS2 server:
curl ifconfig.me

# Add IP to ALLOWED_IPS in .env
# Then restart backend:
docker compose restart backend
```

## Useful Commands Cheatsheet

```bash
# Start all services (Docker Compose v2)
docker compose up -d

# Stop all services
docker compose down

# Restart a service
docker compose restart backend

# View logs (all services)
docker compose logs -f

# View specific service logs
docker compose logs -f backend

# Execute command in container
docker compose exec backend sh

# Check service health
docker compose ps

# Rebuild after code changes
docker compose up -d --build

# Remove everything (including data)
docker compose down -v
```

## Next Steps

1. **Send Real Logs**: Configure your CS2 servers to send logs
2. **Customize**: Modify parsing rules in `backend/internal/parser/`
3. **Add Features**: Check `.claude/implementation-plan.md` for roadmap
4. **Deploy**: Follow `.claude/coolify-deployment.md` for production

## Getting Help

- Check logs first: `docker compose logs -f`
- Review `.env` configuration
- Ensure all services are running: `docker compose ps`
- Open an issue: https://github.com/noueii/nocs-log-saver/issues

---

**Quick Tips:**
- Keep `.env` file private (never commit it)
- Use `127.0.0.1,::1,localhost` in ALLOWED_IPS for local testing
- Add your CS2 server IPs to ALLOWED_IPS before deploying
- Frontend auto-refreshes in development
- Database data persists in Docker volumes
- Consider using random/obscure paths like `/logs/a7x9k2m4` for extra security