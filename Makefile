# CS2 Log Saver - Development Makefile
.PHONY: help dev dev-backend dev-frontend docker-up docker-down docker-rebuild clean setup test logs db-shell install

# Default target - show help
help:
	@echo "CS2 Log Saver - Development Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  make setup          - Initial setup (install deps, copy env files)"
	@echo "  make dev            - Start all services in development mode"
	@echo ""
	@echo "Development Commands:"
	@echo "  make dev-backend    - Start backend server only (Go)"
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
	@echo "  make db-shell       - Open PostgreSQL shell"
	@echo "  make db-migrate     - Run database migrations"
	@echo "  make db-reset       - Reset database (WARNING: destroys data)"
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
	@cp -n .env.example .env 2>/dev/null || echo "✓ .env already exists"
	@cp -n frontend/.env.local.example frontend/.env.local 2>/dev/null || echo "✓ frontend/.env.local already exists"
	@$(MAKE) install
	@echo "✅ Setup complete! Run 'make dev' to start development servers"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@cd backend && go mod download
	@cd frontend && npm install
	@echo "✅ Dependencies installed"

# Start everything in development mode
dev:
	@echo "🚀 Starting all development servers..."
	@echo "Starting PostgreSQL..."
	@docker compose up -d postgres
	@echo "⏳ Waiting for database to be ready..."
	@sleep 3
	@echo ""
	@echo "📝 Instructions:"
	@echo "  Open 2 new terminal windows and run:"
	@echo "  Terminal 1: make dev-backend"
	@echo "  Terminal 2: make dev-frontend"
	@echo ""
	@echo "Or use: make dev-all (runs in background)"

# Start all services in background
dev-all:
	@echo "🚀 Starting all services in background..."
	@docker compose up -d postgres
	@sleep 3
	@cd backend && go run cmd/server/main.go > ../backend.log 2>&1 & echo $$! > ../backend.pid
	@cd frontend && npm run dev > ../frontend.log 2>&1 & echo $$! > ../frontend.pid
	@echo "✅ Services started in background"
	@echo "  Backend PID: $$(cat backend.pid 2>/dev/null)"
	@echo "  Frontend PID: $$(cat frontend.pid 2>/dev/null)"
	@echo ""
	@echo "View logs with:"
	@echo "  make logs-dev"
	@echo ""
	@echo "Stop with:"
	@echo "  make dev-stop"

# Stop background dev services
dev-stop:
	@echo "🛑 Stopping development services..."
	@-kill $$(cat backend.pid 2>/dev/null) 2>/dev/null && rm -f backend.pid && echo "✓ Backend stopped"
	@-kill $$(cat frontend.pid 2>/dev/null) 2>/dev/null && rm -f frontend.pid && echo "✓ Frontend stopped"
	@docker compose down
	@echo "✅ All services stopped"

# View dev logs
logs-dev:
	@echo "📋 Development logs (Ctrl+C to exit):"
	@tail -f backend.log frontend.log

# Start backend in development mode
dev-backend:
	@echo "🔧 Starting backend server on port 9090..."
	@cd backend && go run cmd/server/main.go

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

docker-rebuild:
	@echo "🔨 Rebuilding Docker containers..."
	@docker compose down
	@docker compose build --no-cache
	@docker compose up -d
	@echo "✅ Containers rebuilt and started"

docker-logs:
	@docker compose logs -f

# Database commands
db-shell:
	@echo "🗄️ Opening PostgreSQL shell..."
	@docker compose exec postgres psql -U cs2admin -d cs2logs

db-migrate:
	@echo "📝 Running database migrations..."
	@cd backend && go run cmd/migrate/main.go up

db-reset:
	@echo "⚠️  WARNING: This will delete all data!"
	@read -p "Are you sure? (y/N) " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker compose down -v
	@docker compose up -d postgres
	@sleep 3
	@$(MAKE) db-migrate
	@echo "✅ Database reset complete"

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