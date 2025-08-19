# CS2 Log Saver - Development Makefile
.PHONY: help dev dev-backend dev-frontend server docker-up docker-down docker-rebuild clean setup check-prereqs test logs db-create db-connect db-test db-init db-shell db-query db-migrate db-seed db-reset db-status db-backup db-restore db-url db-setup install

# Default target - show help
help:
	@echo "CS2 Log Saver - Development Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  make check-prereqs  - Check if Go, Node, Docker are installed"
	@echo "  make setup          - Initial setup (install deps, copy env files)"
	@echo "  make dev            - Start all services in development mode"
	@echo ""
	@echo "Development Commands:"
	@echo "  make dev-backend    - Start backend server only (Go)"
	@echo "  make server         - Start backend server (alias for dev-backend)"
	@echo "  make dev-frontend   - Start frontend server only (Next.js)"
	@echo "  make dev-db         - Start PostgreSQL only"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-up      - Start all services with Docker Compose"
	@echo "  make docker-down    - Stop all Docker services"
	@echo "  make docker-rebuild - Rebuild and start Docker containers"
	@echo "  make docker-logs    - Show Docker logs (follow mode)"
	@echo ""
	@echo "Database Commands:"
	@echo "  make db-create      - Start PostgreSQL container"
	@echo "  make db-connect     - Open interactive PostgreSQL session"
	@echo "  make db-test        - Test database connection"
	@echo "  make db-query       - Run a single SQL query"
	@echo "  make db-init        - Initialize database schema"
	@echo "  make db-shell       - Open PostgreSQL shell (same as db-connect)"
	@echo "  make db-migrate     - Run database migrations"
	@echo "  make db-seed        - Seed database with initial data"
	@echo "  make db-reset       - Reset database (WARNING: destroys data)"
	@echo "  make db-status      - Show database status and row counts"
	@echo "  make db-backup      - Create database backup"
	@echo "  make db-restore     - Restore database from backup"
	@echo "  make db-url         - Show database connection strings"
	@echo "  make db-setup       - Complete database setup (create + init + seed)"
	@echo ""
	@echo "Utility Commands:"
	@echo "  make install        - Install all dependencies"
	@echo "  make test           - Run all tests"
	@echo "  make clean          - Clean build artifacts and temp files"
	@echo "  make logs           - Show all logs"
	@echo "  make lint           - Run linters"

# Initial setup
setup:
	@echo "🚀 Setting up CS2 Log Saver development environment..."
	@if not exist .env (copy .env.example .env && echo "✓ Created .env from .env.example") else (echo "✓ .env already exists")
	@if not exist frontend\.env.local (copy frontend\.env.local.example frontend\.env.local && echo "✓ Created frontend/.env.local") else (echo "✓ frontend/.env.local already exists")
	@$(MAKE) install
	@echo "✅ Setup complete! Run 'make dev' to start development servers"

# Check prerequisites
check-prereqs:
	@echo "🔍 Checking prerequisites..."
	@where go >NUL 2>&1 && (echo "✅ Go is installed: " && go version) || (echo "❌ Go is not installed. Please install from https://go.dev/dl/" && exit 1)
	@where node >NUL 2>&1 && (echo "✅ Node.js is installed: " && node --version) || (echo "❌ Node.js is not installed. Please install from https://nodejs.org/" && exit 1)
	@where npm >NUL 2>&1 && (echo "✅ NPM is installed: " && npm --version) || (echo "❌ NPM is not installed" && exit 1)
	@where docker >NUL 2>&1 && (echo "✅ Docker is installed: " && docker --version) || (echo "❌ Docker is not installed. Please install Docker Desktop" && exit 1)
	@echo ""
	@echo "✅ All prerequisites are installed!"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@echo "Checking for Go..."
	@where go >NUL 2>&1 && (cd backend && go mod download && echo "✅ Go dependencies installed") || echo "⚠️  Skipping Go dependencies (Go not installed)"
	@echo ""
	@echo "Checking for Node.js..."
	@where npm >NUL 2>&1 && (cd frontend && npm install && echo "✅ Node dependencies installed") || echo "⚠️  Skipping Node dependencies (Node/NPM not installed)"
	@echo ""
	@echo "✅ Installation complete (installed available dependencies)"

# Start everything in development mode
dev:
	@echo "🚀 Starting all development servers..."
	@echo "Starting PostgreSQL..."
	@docker compose up -d postgres
	@echo "⏳ Waiting for database to be ready..."
	@timeout /t 3 /nobreak >NUL 2>&1
	@echo ""
	@echo "📝 Instructions:"
	@echo "  Open 2 new terminal windows and run:"
	@echo "  Terminal 1: make dev-backend (or make server)"
	@echo "  Terminal 2: make dev-frontend"
	@echo ""
	@echo "Or use: make dev-all (runs in background)"

# Start all services in background (Windows)
dev-all:
	@echo "🚀 Starting all services in background..."
	@docker compose up -d postgres
	@timeout /t 3 /nobreak >NUL 2>&1
	@echo "Starting backend server..."
	@start /B cmd /c "cd backend && go run cmd/server/main.go > ../backend.log 2>&1"
	@echo "Starting frontend server..."
	@start /B cmd /c "cd frontend && npm run dev > ../frontend.log 2>&1"
	@echo "✅ Services started in background"
	@echo ""
	@echo "View logs with:"
	@echo "  make logs-dev"
	@echo ""
	@echo "Stop with:"
	@echo "  make dev-stop"

# Stop background dev services (Windows)
dev-stop:
	@echo "🛑 Stopping development services..."
	@taskkill /F /FI "WINDOWTITLE eq backend*" >NUL 2>&1 && echo "✓ Backend stopped" || echo "✓ Backend not running"
	@taskkill /F /FI "WINDOWTITLE eq frontend*" >NUL 2>&1 && echo "✓ Frontend stopped" || echo "✓ Frontend not running"
	@docker compose down
	@echo "✅ All services stopped"

# View dev logs
logs-dev:
	@echo "📋 Development logs (Ctrl+C to exit):"
	@type backend.log 2>NUL || echo "No backend logs yet"
	@echo ""
	@type frontend.log 2>NUL || echo "No frontend logs yet"

# Start backend in development mode
dev-backend:
	@echo "🔧 Starting backend server on port 9090..."
	@cd backend && go run cmd/server/main.go

# Alias for backend server
server: dev-backend

# Start frontend in development mode
dev-frontend:
	@echo "🎨 Starting frontend server on port 6173..."
	@cd frontend && npm run dev

# Start database only
dev-db:
	@echo "🗄️ Starting PostgreSQL database..."
	@docker compose up -d postgres
	@echo "✅ Database started on port 5432"

# Docker commands
docker-up:
	@echo "🐳 Starting all services with Docker Compose..."
	@docker compose up -d
	@echo "✅ Services started:"
	@echo "  - Frontend: http://localhost:6173"
	@echo "  - Backend:  http://localhost:9090"
	@echo "  - Admin:    http://localhost:6173/admin"

docker-down:
	@echo "🛑 Stopping Docker services..."
	@docker compose down
	@echo "✅ Services stopped"

docker-clean:
	@echo "🧹 COMPLETELY removing Docker containers and volumes..."
	@docker compose down -v
	@docker volume rm nocs-log-saver_postgres_data 2>NUL || echo "Volume already removed"
	@echo "✅ All Docker data removed"

docker-rebuild:
	@echo "🔨 COMPLETE REBUILD - This will delete everything and start fresh"
	@docker compose down -v
	@docker compose build --no-cache
	@docker compose up -d
	@echo "✅ Everything is ready!"
	@echo ""
	@echo "📋 Access the application:"
	@echo "   Frontend: http://localhost:6173"
	@echo "   Backend:  http://localhost:9090"
	@echo ""
	@echo "🔐 Login credentials:"
	@echo "   Username: admin"
	@echo "   Password: Admin123!"

docker-logs:
	@docker compose logs -f

# Database commands
db-create:
	@echo "🗄️ Starting PostgreSQL container..."
	@docker compose up -d postgres
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@echo "   This may take 10-30 seconds on first run..."
	@timeout /t 10 /nobreak >NUL 2>&1
	@docker compose exec -T postgres pg_isready -U cs2admin >NUL 2>&1 || timeout /t 5 /nobreak >NUL 2>&1
	@docker compose exec -T postgres pg_isready -U cs2admin >NUL 2>&1 || timeout /t 5 /nobreak >NUL 2>&1
	@docker compose exec -T postgres pg_isready -U cs2admin >NUL 2>&1 || timeout /t 5 /nobreak >NUL 2>&1
	@echo "✅ PostgreSQL container is running!"
	@echo ""
	@echo "📝 Database Information:"
	@echo "    Database: cs2logs"
	@echo "    User: cs2admin"
	@echo "    Password: localpass123 (or DB_PASSWORD env var)"
	@echo "    Port: 5432"
	@echo ""
	@echo "Run 'make db-connect' to test the connection"

db-connect:
	@echo "🔌 Connecting to PostgreSQL interactive session..."
	@echo "📝 Type \q to exit, \? for help, \dt to list tables"
	@echo ""
	@docker compose exec postgres psql -U cs2admin -d cs2logs

db-test:
	@echo "🔌 Testing database connection..."
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT version();" 2>NUL && echo "✅ Database connection successful" || (echo "❌ Connection failed. Run 'make db-create' first" && exit 1)

db-init:
	@echo "📋 Initializing database schema..."
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE TABLE IF NOT EXISTS servers (id VARCHAR(50) PRIMARY KEY, name VARCHAR(100), ip_address VARCHAR(45), api_key VARCHAR(255), last_seen TIMESTAMP, created_at TIMESTAMP DEFAULT NOW());"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE TABLE IF NOT EXISTS raw_logs (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), server_id VARCHAR(50) REFERENCES servers(id), content TEXT NOT NULL, received_at TIMESTAMP DEFAULT NOW());"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE TABLE IF NOT EXISTS parsed_logs (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), raw_log_id UUID REFERENCES raw_logs(id), server_id VARCHAR(50) REFERENCES servers(id), event_type VARCHAR(50), event_data JSONB, game_time VARCHAR(20), session_id VARCHAR(100), created_at TIMESTAMP DEFAULT NOW());"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE TABLE IF NOT EXISTS failed_parses (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), raw_log_id UUID REFERENCES raw_logs(id), error_message TEXT, retry_count INTEGER DEFAULT 0, last_retry TIMESTAMP, resolved BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT NOW());"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE TABLE IF NOT EXISTS sessions (id VARCHAR(100) PRIMARY KEY, server_id VARCHAR(50) REFERENCES servers(id), map_name VARCHAR(100), started_at TIMESTAMP, ended_at TIMESTAMP, status VARCHAR(20) DEFAULT 'active', metadata JSONB);"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE INDEX IF NOT EXISTS idx_raw_logs_server_id ON raw_logs(server_id);"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE INDEX IF NOT EXISTS idx_parsed_logs_session_id ON parsed_logs(session_id);"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE INDEX IF NOT EXISTS idx_parsed_logs_event_type ON parsed_logs(event_type);"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE INDEX IF NOT EXISTS idx_raw_logs_received_at ON raw_logs(received_at DESC);"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "CREATE INDEX IF NOT EXISTS idx_sessions_server_id ON sessions(server_id);"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"
	@echo "✅ Database schema initialized"

db-seed:
	@echo "🌱 Seeding database with initial data..."
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "INSERT INTO servers (id, name, ip_address, api_key) VALUES ('testserver', 'Test Server', '127.0.0.1', 'test-api-key-123'), ('server1', 'Production Server 1', '192.168.1.100', 'prod-api-key-456') ON CONFLICT (id) DO NOTHING;"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT id, name, ip_address FROM servers;"
	@echo "✅ Database seeded with initial data"

db-shell:
	@echo "🗄️ Opening PostgreSQL shell (same as db-connect)..."
	@docker compose exec postgres psql -U cs2admin -d cs2logs

db-query:
	@echo "📝 Enter SQL query (or press Ctrl+C to cancel):"
	@set /p query="SQL> " && docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "%%query%%"

db-migrate:
	@echo "📝 Running database migrations..."
	@cd backend && go run cmd/migrate/main.go up

db-reset:
	@echo "⚠️  WARNING: This will delete all data!"
	@read -p "Are you sure? (y/N) " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker compose down -v
	@docker compose up -d postgres
	@sleep 3
	@$(MAKE) db-create
	@$(MAKE) db-init
	@echo "✅ Database reset complete"

db-status:
	@echo "📊 Database Status:"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "\dt+" 2>NUL || echo "❌ Database not running"
	@echo ""
	@echo "📈 Table Row Counts:"
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT CONCAT('servers: ', COUNT(*)) FROM servers;" 2>NUL || echo ""
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT CONCAT('raw_logs: ', COUNT(*)) FROM raw_logs;" 2>NUL || echo ""
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT CONCAT('parsed_logs: ', COUNT(*)) FROM parsed_logs;" 2>NUL || echo ""
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT CONCAT('sessions: ', COUNT(*)) FROM sessions;" 2>NUL || echo ""
	@docker compose exec -T postgres psql -U cs2admin -d cs2logs -c "SELECT CONCAT('failed_parses: ', COUNT(*)) FROM failed_parses;" 2>NUL || echo "No data available"

db-backup:
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	@docker compose exec -T postgres pg_dump -U cs2admin -d cs2logs > backups/cs2logs_backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup saved to backups/cs2logs_backup_$$(date +%Y%m%d_%H%M%S).sql"

db-restore:
	@echo "📥 Restoring database from backup..."
	@read -p "Enter backup file path (e.g., backups/cs2logs_backup_20250119_120000.sql): " backup_file; \
	if [ -f "$$backup_file" ]; then \
		docker compose exec -T postgres psql -U cs2admin -d cs2logs < $$backup_file && \
		echo "✅ Database restored from $$backup_file"; \
	else \
		echo "❌ Backup file not found: $$backup_file"; \
	fi

# Database connection string helper
db-url:
	@echo "📋 Database Connection URLs:"
	@echo ""
	@echo "Local development:"
	@echo "  postgresql://cs2admin:localpass123@localhost:5432/cs2logs"
	@echo ""
	@echo "Docker internal:"
	@echo "  postgresql://cs2admin:localpass123@postgres:5432/cs2logs"
	@echo ""
	@echo "Go application (with sslmode):"
	@echo "  postgres://cs2admin:localpass123@localhost:5432/cs2logs?sslmode=disable"

# Complete database setup
db-setup:
	@echo "🚀 Setting up complete database..."
	@$(MAKE) db-create
	@$(MAKE) db-init
	@$(MAKE) db-seed
	@$(MAKE) db-status
	@echo "✅ Database setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make dev-backend' to start the backend"
	@echo "  2. Run 'make dev-frontend' to start the frontend"
	@echo "  3. Or run 'make docker-up' to start everything"

# Testing
test:
	@echo "🧪 Running tests..."
	@cd backend && go test ./...
	@cd frontend && npm test

test-backend:
	@echo "🧪 Running backend tests..."
	@cd backend && go test -v ./...

test-frontend:
	@echo "🧪 Running frontend tests..."
	@cd frontend && npm test

# Linting
lint:
	@echo "🔍 Running linters..."
	@cd backend && go fmt ./... && go vet ./...
	@cd frontend && npm run lint

lint-backend:
	@echo "🔍 Linting backend code..."
	@cd backend && go fmt ./... && go vet ./...

lint-frontend:
	@echo "🔍 Linting frontend code..."
	@cd frontend && npm run lint

# Build commands
build:
	@echo "🏗️ Building production artifacts..."
	@$(MAKE) build-backend
	@$(MAKE) build-frontend

build-backend:
	@echo "🏗️ Building backend..."
	@cd backend && go build -o bin/server cmd/server/main.go
	@echo "✅ Backend built: backend/bin/server"

build-frontend:
	@echo "🏗️ Building frontend..."
	@cd frontend && npm run build
	@echo "✅ Frontend built: frontend/.next"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf backend/bin
	@rm -rf backend/tmp
	@rm -rf frontend/.next
	@rm -rf frontend/node_modules/.cache
	@echo "✅ Clean complete"

# Show logs
logs:
	@docker compose logs -f --tail=100

logs-backend:
	@docker compose logs -f backend --tail=100

logs-frontend:
	@docker compose logs -f frontend --tail=100

logs-db:
	@docker compose logs -f postgres --tail=100

# Production commands
prod-build:
	@echo "📦 Building for production..."
	@docker compose -f docker-compose.yml build

prod-up:
	@echo "🚀 Starting production services..."
	@docker compose -f docker-compose.yml up -d
	@echo "✅ Production services started"

# Development shortcuts
.PHONY: b f d
b: dev-backend
f: dev-frontend  
d: dev-db

# Status check
status:
	@echo "📊 Service Status:"
	@docker compose ps
	@echo ""
	@echo "🌐 URLs:"
	@echo "  Frontend: http://localhost:6173"
	@echo "  Backend:  http://localhost:9090/health"
	@echo "  Admin:    http://localhost:6173/admin"

# Send test log (for testing)
test-log:
	@echo "📤 Sending test log to backend..."
	@curl -X POST http://localhost:9090/logs/testserver \
		-H "Content-Type: text/plain" \
		-d 'L 01/17/2025 - 12:00:00: "Player<1><STEAM_1:0:123456><CT>" killed "Enemy<2><STEAM_1:0:654321><T>" with "ak47"' \
		&& echo "\n✅ Test log sent" || echo "\n❌ Failed to send test log"

# Watch for changes and restart
watch-backend:
	@echo "👁️ Watching backend for changes..."
	@cd backend && air || (go install github.com/cosmtrek/air@latest && air)

watch-frontend:
	@cd frontend && npm run dev

# Port check
check-ports:
	@echo "🔍 Checking if required ports are available..."
	@lsof -i :9090 >/dev/null 2>&1 && echo "❌ Port 9090 (backend) is in use" || echo "✅ Port 9090 (backend) is available"
	@lsof -i :6173 >/dev/null 2>&1 && echo "❌ Port 6173 (frontend) is in use" || echo "✅ Port 6173 (frontend) is available"
	@lsof -i :5432 >/dev/null 2>&1 && echo "❌ Port 5432 (postgres) is in use" || echo "✅ Port 5432 (postgres) is available"

# Environment info
info:
	@echo "ℹ️  Environment Information:"
	@echo "  Go version:      $$(go version 2>/dev/null || echo 'Not installed')"
	@echo "  Node version:    $$(node --version 2>/dev/null || echo 'Not installed')"
	@echo "  NPM version:     $$(npm --version 2>/dev/null || echo 'Not installed')"
	@echo "  Docker version:  $$(docker --version 2>/dev/null || echo 'Not installed')"
	@echo "  Docker Compose:  $$(docker compose version 2>/dev/null || echo 'Not installed')"
	@echo ""
	@echo "📁 Project Structure:"
	@echo "  Backend:  $(CURDIR)/backend"
	@echo "  Frontend: $(CURDIR)/frontend"
	@echo "  Docs:     $(CURDIR)/.claude"