# CS2 Log Saver - Lean Implementation Plan

## Goal: Ship Working MVP in 2-3 Weeks

### Core Principle: Build Fast, Ship Fast, Iterate Later

## Tech Stack (Simple & Proven)

### Backend
- **Language**: Go 1.23+ (latest stable)
- **Framework**: Gin (lightweight, fast, requires Go 1.23+)
- **Database**: PostgreSQL 17 (single DB for all data)
- **Parser**: github.com/janstuemmel/cs2-log or github.com/joao-silva1007/cs2-log-re2
- **Storage**: Local filesystem → PostgreSQL (migrate to S3 later if needed)

### Frontend  
- **Framework**: Next.js 15.3 (App Router, React 19, Turbopack)
- **Runtime**: Node.js 22 LTS
- **UI Library**: shadcn/ui + Tailwind CSS v3
- **Data Fetching**: Native fetch with SWR for caching
- **Charts**: Recharts (if needed, keep minimal)

### Deployment
- **Container**: Docker + Docker Compose v2 (Compose Specification)
- **Platform**: Coolify (self-hosted PaaS)
- **Database**: PostgreSQL 17 via Coolify
- **SSL**: Automatic via Coolify/Caddy

## Project Structure

```
nocs-log-saver/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers.go       # HTTP handlers
│   │   │   └── middleware.go     # Auth middleware
│   │   ├── db/
│   │   │   ├── models.go         # Database models
│   │   │   └── migrations/       # SQL migrations
│   │   ├── parser/
│   │   │   └── cs2.go            # cs2-log integration
│   │   └── storage/
│   │       └── logs.go           # Log storage logic
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── frontend/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx              # Dashboard
│   │   ├── servers/
│   │   │   └── page.tsx          # Server list
│   │   ├── logs/
│   │   │   └── page.tsx          # Log viewer
│   │   └── sessions/
│   │       ├── page.tsx          # Session list
│   │       └── [id]/
│   │           └── page.tsx      # Session details
│   ├── components/
│   │   ├── ui/                   # shadcn/ui components
│   │   ├── LogTable.tsx
│   │   ├── ServerCard.tsx
│   │   └── SessionTimeline.tsx
│   ├── lib/
│   │   ├── api.ts                # API client
│   │   └── utils.ts
│   ├── package.json
│   └── Dockerfile
│
├── docker-compose.yml             # Local dev + Coolify deployment
├── .env.example
├── README.md
└── .claude/                       # Project documentation
```

## Architecture Guidelines

### Backend Architecture (Clean Architecture)
- **Domain Layer**: Core business logic, entities, interfaces
- **Application Layer**: Use cases, DTOs, application services  
- **Infrastructure Layer**: Database, external services, parsers
- **Interface Layer**: HTTP handlers, middleware, validators

### Key Principles
- **Dependency Injection**: All dependencies passed via constructors
- **Interface-based Design**: Program to interfaces, not implementations
- **Repository Pattern**: Abstract data access behind interfaces
- **Service Layer**: Business logic separated from handlers
- **Error Handling**: Consistent error types and wrapping

### Testing Strategy
- **Unit Tests**: Test individual functions/methods in isolation
- **Integration Tests**: Test API endpoints with real database
- **Mock Dependencies**: Use interfaces for easy mocking
- **Table-driven Tests**: Group similar test cases

## Week 1: Backend Core (Days 1-5)

### Day 1-2: Foundation
```bash
# Initialize Go module (requires Go 1.23+)
go mod init github.com/noueii/nocs-log-saver

# Install dependencies
go get github.com/gin-gonic/gin@latest
go get github.com/lib/pq
go get github.com/joho/godotenv
# Install CS2 log parser
go get github.com/janstuemmel/cs2-log
```

**Tasks:**
- [ ] Setup Go project structure following clean architecture
- [ ] Create Gin server with dependency injection
- [ ] PostgreSQL repository implementation with interfaces
- [ ] `/logs/{server_id}` POST endpoint with proper validation
- [ ] Store raw logs using repository pattern
- [ ] Dynamic IP whitelist service with caching
- [ ] Admin endpoints with proper authorization

**Example Structure:**
```go
// internal/domain/entities/log.go
type Log struct {
    ID        string
    ServerID  string
    Content   string
    CreatedAt time.Time
}

// internal/domain/repositories/log_repository.go
type LogRepository interface {
    Save(ctx context.Context, log *Log) error
    FindByID(ctx context.Context, id string) (*Log, error)
}

// internal/infrastructure/persistence/postgres_log_repo.go
type PostgresLogRepository struct {
    db *sql.DB
}

func (r *PostgresLogRepository) Save(ctx context.Context, log *entities.Log) error {
    query := `INSERT INTO raw_logs (server_id, content) VALUES ($1, $2)`
    _, err := r.db.ExecContext(ctx, query, log.ServerID, log.Content)
    return err
}

// internal/application/services/log_service.go
type LogService struct {
    repo   repositories.LogRepository
    parser services.Parser
}

func NewLogService(repo repositories.LogRepository, parser services.Parser) *LogService {
    return &LogService{repo: repo, parser: parser}
}

// internal/interfaces/http/handlers/log_handler.go
type LogHandler struct {
    logService *application.LogService
    validator  *validators.LogValidator
}

func (h *LogHandler) IngestLog(c *gin.Context) {
    // Validate input
    // Call service
    // Return response
}
```

### Day 3-4: Parse & Store
```go
// Integrate cs2-log library
import "github.com/janstuemmel/cs2-log"
```

**Tasks:**
- [ ] Integrate cs2-log parser
- [ ] Parse incoming logs asynchronously
- [ ] Store parsed logs in structured format
- [ ] Capture failed parses with error details
- [ ] Basic session detection (match start/end)

### Day 5: Dockerize
**docker-compose.yml:**
```yaml
# Using Compose Specification (no version field needed)
services:
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: cs2logs
      POSTGRES_USER: cs2admin
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cs2admin"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://cs2admin:${DB_PASSWORD}@postgres:5432/cs2logs
      ALLOWED_IPS: ${ALLOWED_IPS}  # Comma-separated list of whitelisted IPs
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  postgres_data:
```

## Week 2: Frontend UI (Days 6-10)

### Day 6-7: Next.js Setup
```bash
# Create Next.js 15 app with latest features
npx create-next-app@latest frontend --typescript --tailwind --app --turbopack
cd frontend
npx shadcn@latest init
npx shadcn@latest add table card button input
```

**Pages to Build:**
1. **Dashboard** (`/`) - Overview stats
2. **Servers** (`/servers`) - List of connected servers
3. **Logs** (`/logs`) - Browse raw/parsed logs
4. **Sessions** (`/sessions`) - Match sessions list
5. **Session Detail** (`/sessions/[id]`) - Single match view
6. **Admin** (`/admin`) - Settings and IP whitelist management

### Day 8-9: Core Components
**Essential Components:**
```typescript
// components/LogTable.tsx
- Paginated table for logs
- Search/filter by server
- Show parse status

// components/ServerCard.tsx  
- Server status (online/offline)
- Last seen timestamp
- Log count
- Associated IPs

// components/SessionTimeline.tsx
- Match phases (warmup/live/halftime)
- Round progression
- Basic scores

// components/admin/IPWhitelistManager.tsx
- List whitelisted IPs
- Add/Edit/Delete IPs
- Enable/Disable IPs
- Link IPs to servers
```

### Day 10: API Integration
```typescript
// lib/api.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = {
  getLogs: (serverId?: string) => 
    fetch(`${API_BASE}/api/logs?server_id=${serverId}`).then(r => r.json()),
  
  getSessions: () =>
    fetch(`${API_BASE}/api/sessions`).then(r => r.json()),
    
  getServers: () =>
    fetch(`${API_BASE}/api/servers`).then(r => r.json())
};
```

## Week 3: Deploy to Coolify (Days 11-15)

### Day 11: Production Prep
**Tasks:**
- [ ] Add health check endpoints
- [ ] Environment variable validation
- [ ] Basic error handling
- [ ] CORS configuration
- [ ] Update docker-compose for production

### Day 12-13: Coolify Setup
1. **Push to GitHub:**
```bash
git init
git add .
git commit -m "Initial MVP"
git remote add origin https://github.com/noueii/nocs-log-saver
git push -u origin main
```

2. **Coolify Configuration:**
- Add new project in Coolify
- Connect GitHub repository
- Set up PostgreSQL service
- Configure environment variables
- Set up domain/subdomain

### Day 14: Testing & Fixes
- [ ] Test log ingestion from real CS2 server
- [ ] Verify parsing works correctly
- [ ] Check UI displays data properly
- [ ] Fix critical bugs only

### Day 15: Documentation
- [ ] Basic README with setup instructions
- [ ] API documentation (simple markdown)
- [ ] Coolify deployment guide

## Database Schema (Simple)

```sql
-- Servers
CREATE TABLE servers (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100),
    ip_address VARCHAR(45),
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Raw logs
CREATE TABLE raw_logs (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(50) REFERENCES servers(id),
    content TEXT NOT NULL,
    received_at TIMESTAMP DEFAULT NOW()
);

-- Parsed logs
CREATE TABLE parsed_logs (
    id SERIAL PRIMARY KEY,
    raw_log_id INTEGER REFERENCES raw_logs(id),
    server_id VARCHAR(50) REFERENCES servers(id),
    event_type VARCHAR(50),
    event_data JSONB,
    game_time VARCHAR(20),
    session_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Failed parses
CREATE TABLE failed_parses (
    id SERIAL PRIMARY KEY,
    raw_log_id INTEGER REFERENCES raw_logs(id),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Sessions
CREATE TABLE sessions (
    id VARCHAR(100) PRIMARY KEY,
    server_id VARCHAR(50) REFERENCES servers(id),
    map_name VARCHAR(100),
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active',
    metadata JSONB
);

-- IP Whitelist (Dynamic management)
CREATE TABLE ip_whitelist (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(45) UNIQUE NOT NULL,
    server_id VARCHAR(50) REFERENCES servers(id),
    description VARCHAR(255),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(100)
);

CREATE INDEX idx_ip_whitelist_ip ON ip_whitelist(ip_address) WHERE enabled = true;
```

## Security Without Authentication

### IP Whitelisting
Since CS2 servers cannot authenticate, we use network-level security:
- **Allowed IPs**: Environment variable with comma-separated list
- **Middleware**: Check request IP against whitelist
- **Reject**: Return 403 for unauthorized IPs
- **Optional**: Use obscure/random paths like `/logs/a7x9k2m4` instead of predictable names

### Rate Limiting
- **Per-IP limits**: Max requests per minute
- **Burst protection**: Short-term spike handling
- **At proxy level**: Configure in Coolify/Caddy/nginx

### Additional Security
- **Cloudflare**: Optional DDoS protection
- **Firewall rules**: Network-level blocking
- **Log rotation**: Prevent storage exhaustion

## API Endpoints (MVP)

### Ingestion
- `POST /logs/{server_id}` - Receive logs from CS2 server (IP whitelist protected)

### Query APIs  
- `GET /api/servers` - List all servers
- `GET /api/logs?server_id=X&type=raw|parsed` - Get logs
- `GET /api/sessions?status=active|completed` - Get sessions
- `GET /api/sessions/{id}` - Get session details
- `GET /api/stats` - Basic statistics

### Admin APIs
- `GET /api/admin/whitelist` - List all whitelisted IPs
- `POST /api/admin/whitelist` - Add new IP to whitelist
- `PUT /api/admin/whitelist/{id}` - Update IP entry
- `DELETE /api/admin/whitelist/{id}` - Remove IP from whitelist
- `GET /api/admin/whitelist/check/{ip}` - Check if IP is whitelisted

### Health
- `GET /health` - Health check for Coolify

## What We're Shipping (MVP Features)

### ✅ MUST Have
- Receive logs from multiple CS2 servers
- Store raw logs
- Parse logs with cs2-log library
- Store parsed and failed logs
- Basic session detection
- Web UI to view logs
- Server list with status
- Session/match list
- Admin dashboard with IP whitelist management
- Deploy to Coolify

### ❌ NOT in MVP (Add Later)
- User authentication (IP whitelist only)
- Real-time updates (just refresh)
- Complex visualizations
- Advanced search/filtering
- Export functionality
- Email alerts
- Performance optimization
- Horizontal scaling
- Microservices
- Message queues

## Success Criteria for MVP

1. **It Works**: Can receive and parse CS2 logs
2. **It's Visible**: UI shows logs and sessions
3. **It's Deployed**: Running on Coolify
4. **It's Stable**: Doesn't crash under normal load (1-10 servers)
5. **It's Simple**: <5000 lines of code total

## Quick Commands

```bash
# Local development (Docker Compose v2)
docker compose up

# Run backend (requires Go 1.23+)
cd backend && go run cmd/server/main.go

# Run frontend (with Turbopack)
cd frontend && npm run dev --turbo

# Build for production
docker compose -f docker-compose.prod.yml build

# Deploy to Coolify (after git push)
# Coolify will auto-deploy from GitHub webhook
```

## Next Steps After MVP

Once deployed and working:
1. Gather user feedback
2. Monitor for critical issues
3. Add most requested features
4. Optimize only proven bottlenecks
5. Scale only when needed

---

**Remember**: Perfect is the enemy of shipped. Get it working, get it deployed, iterate based on real usage.