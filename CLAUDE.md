# CLAUDE.md - CS2 Log Saver Project Context

## 🚨 IMPORTANT: MCP Server Usage

**When implementing this project, ALWAYS use MCP servers when available:**
- **UI Components**: Use `mcp__shadcn-ui` to fetch all UI components
- **File Operations**: Use `mcp__filesystem` for file management
- **GitHub**: Use `mcp__github` for repository operations

**Never manually copy shadcn/ui components** - always fetch via MCP for latest versions.

## Project Overview
This is a **CS2 (Counter-Strike 2) log aggregation service** that receives, parses, stores, and visualizes logs from multiple CS2 game servers. The project prioritizes **shipping fast** over premature optimization.

## Core Requirements

### What This Project Does
1. **Receives logs** from multiple CS2 servers via HTTP endpoints
2. **Stores three types of logs**:
   - Raw logs (original, unmodified)
   - Parsed logs (structured data via cs2-log Go library)
   - Failed parse logs (logs that couldn't be parsed)
3. **Organizes logs by sessions**:
   - Server sessions (server start to stop)
   - Match sessions (individual games)
   - Game phases (warmup, live, halftime, overtime, post-match)
4. **Provides a web UI** for visualization and management
5. **Deploys to Coolify** for self-hosted production

## Critical Implementation Details

### Security Model - Dynamic IP Whitelist
**IMPORTANT**: CS2 servers CANNOT send authentication headers. Security is handled through:
- **Dynamic IP Whitelisting**: Managed through admin UI, stored in database
- **Fallback**: `ALLOWED_IPS` environment variable for initial setup
- **Rate Limiting**: Configured at proxy/reverse proxy level
- **Optional**: Obscure endpoint paths (e.g., `/logs/a7x9k2m4` instead of `/logs/server1`)

Never implement API key authentication - it won't work with CS2 servers.

### Admin Dashboard Features
- **IP Whitelist Management**: Add/remove/edit allowed IPs through web UI
- **Server Association**: Link IPs to specific server IDs
- **Real-time Updates**: Changes apply immediately without restart
- **Audit Log**: Track who made whitelist changes and when

### Tech Stack (Keep It Simple)
```yaml
Backend:
  - Language: Go 1.23+ (latest stable)
  - Framework: Gin (lightweight HTTP, requires Go 1.23+)
  - Database: PostgreSQL 17 (single DB for everything)
  - Parser: github.com/janstuemmel/cs2-log or github.com/joao-silva1007/cs2-log-re2 (performance fork)
  
Frontend:
  - Framework: Next.js 15.3 (App Router, React 19, Turbopack)
  - Runtime: Node.js 22 LTS
  - UI: shadcn/ui + Tailwind CSS v3
  - Charts: Recharts (minimal, only if needed)
  
Deployment:
  - Platform: Coolify (self-hosted PaaS)
  - Container: Docker + Docker Compose v2 (Compose Specification)
  - SSL: Automatic via Coolify/Caddy
```

### Project Structure (Clean Architecture)
```
nocs-log-saver/
├── backend/                        # Go backend service
│   ├── cmd/
│   │   └── server/                # Application entry points
│   │       └── main.go           
│   ├── internal/                  # Private application code
│   │   ├── domain/               # Core business logic (entities, interfaces)
│   │   │   ├── entities/         # Business entities
│   │   │   ├── repositories/     # Repository interfaces
│   │   │   └── services/         # Domain services interfaces
│   │   ├── application/          # Use cases / application services
│   │   │   ├── commands/         # Write operations
│   │   │   ├── queries/          # Read operations
│   │   │   └── dto/              # Data transfer objects
│   │   ├── infrastructure/       # External concerns
│   │   │   ├── persistence/      # Database implementation
│   │   │   ├── parser/           # CS2 log parser adapter
│   │   │   └── config/           # Configuration
│   │   └── interfaces/           # Interface adapters
│   │       ├── http/             # HTTP handlers
│   │       │   ├── handlers/     # Request handlers
│   │       │   ├── middleware/   # HTTP middleware
│   │       │   └── validators/   # Input validation
│   │       └── websocket/        # WebSocket handlers (future)
│   ├── pkg/                      # Public packages (can be imported)
│   │   ├── errors/              # Custom error types
│   │   └── utils/               # Shared utilities
│   └── Dockerfile
├── frontend/                      # Next.js UI
│   ├── app/                      # App Router pages
│   ├── components/               # React components
│   │   ├── ui/                  # Presentational components
│   │   └── features/            # Feature-specific components
│   ├── hooks/                   # Custom React hooks
│   ├── lib/                     # Utilities and helpers
│   ├── services/                # API client services
│   └── types/                   # TypeScript type definitions
├── docker-compose.yml            # Local dev + production
├── .env.example                  # Environment template
└── .claude/                      # Project documentation
```

## Implementation Priorities

### Phase 1: MVP (Ship in 2-3 weeks)
✅ **MUST Have**:
- Basic HTTP server accepting logs at `/logs/{server_id}`
- Dynamic IP whitelist with database storage
- Admin UI for managing whitelisted IPs
- Store raw logs in PostgreSQL
- Parse logs with cs2-log library
- Basic session detection
- Simple web UI to view logs
- Deploy to Coolify

❌ **NOT in MVP**:
- User authentication
- Real-time updates (just page refresh)
- Complex visualizations
- Performance optimization
- Message queues
- Microservices
- S3/cloud storage

### What Success Looks Like
1. **It works**: Receives and parses CS2 logs
2. **It's visible**: UI shows logs and sessions
3. **It's deployed**: Running on Coolify
4. **It's stable**: Handles 1-10 servers without crashing

## Database Schema (Simple)

```sql
-- Keep it simple, one database for everything
CREATE TABLE servers (
    id VARCHAR(50) PRIMARY KEY,        -- e.g., "server1", "competitive-us-east"
    name VARCHAR(100),
    ip_address VARCHAR(45),
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE raw_logs (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(50) REFERENCES servers(id),
    content TEXT NOT NULL,              -- Original log line
    received_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE parsed_logs (
    id SERIAL PRIMARY KEY,
    raw_log_id INTEGER REFERENCES raw_logs(id),
    server_id VARCHAR(50) REFERENCES servers(id),
    event_type VARCHAR(50),            -- kill, round_start, etc.
    event_data JSONB,                   -- Flexible parsed data
    session_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE failed_parses (
    id SERIAL PRIMARY KEY,
    raw_log_id INTEGER REFERENCES raw_logs(id),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE sessions (
    id VARCHAR(100) PRIMARY KEY,
    server_id VARCHAR(50) REFERENCES servers(id),
    map_name VARCHAR(100),
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active',
    metadata JSONB
);

CREATE TABLE ip_whitelist (
    id SERIAL PRIMARY KEY,
    ip_address VARCHAR(45) UNIQUE NOT NULL,  -- IPv4/IPv6
    server_id VARCHAR(50) REFERENCES servers(id),
    description VARCHAR(255),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(100)  -- For audit trail
);

-- Index for fast IP lookups
CREATE INDEX idx_ip_whitelist_ip ON ip_whitelist(ip_address) WHERE enabled = true;
```

## API Endpoints

### Log Ingestion (No Auth Required)
```
POST /logs/{server_id}
- IP whitelist protected
- Accepts: text/plain (line-delimited logs)
- Returns: 200 OK or 403 Forbidden
```

### Query APIs (Frontend Use)
```
GET /api/servers                    # List all servers
GET /api/logs?server_id=X&type=raw  # Get logs
GET /api/sessions                   # List sessions
GET /api/sessions/{id}              # Session details
GET /api/stats                      # Basic statistics
GET /health                         # Health check
```

### Admin APIs (Protected)
```
GET /api/admin/whitelist            # List all whitelisted IPs
POST /api/admin/whitelist           # Add new IP
PUT /api/admin/whitelist/{id}       # Update IP entry
DELETE /api/admin/whitelist/{id}    # Remove IP
GET /api/admin/whitelist/check/{ip} # Check if IP is whitelisted
```

## Environment Variables

```bash
# Required
DATABASE_URL=postgresql://user:pass@host:5432/dbname
ALLOWED_IPS=192.168.1.100,203.0.113.45  # CS2 server IPs
NEXT_PUBLIC_API_URL=https://api.domain.com

# Optional
BACKEND_PORT=9090
FRONTEND_PORT=6173
LOG_LEVEL=info
MAX_LOG_SIZE_MB=10
LOG_RETENTION_DAYS=30
```

## CS2 Server Configuration

CS2 servers should be configured to send logs:

```cfg
# In server.cfg or autoexec.cfg
log on
logaddress_add "https://api.yourdomain.com/logs/server1"

# No authentication possible - use IP whitelist
```

## Best Coding Practices & Principles

### SOLID Principles

#### 1. Single Responsibility Principle (SRP)
```go
// BAD: Handler doing too much
func HandleLog(c *gin.Context) {
    // Parse request, validate, save to DB, parse log, detect session...
}

// GOOD: Separated concerns
type LogHandler struct {
    validator  LogValidator
    storage    LogStorage
    parser     LogParser
    sessionSvc SessionService
}

func (h *LogHandler) Handle(c *gin.Context) {
    log := h.validator.Validate(c)
    h.storage.Save(log)
    go h.processAsync(log)
}
```

#### 2. Open/Closed Principle (OCP)
```go
// Define interfaces for extension
type LogParser interface {
    Parse(raw string) (*ParsedLog, error)
}

type LogStorage interface {
    Save(ctx context.Context, log *Log) error
    Get(ctx context.Context, id string) (*Log, error)
}

// Implementations can be swapped without changing core logic
type CS2LogParser struct{}
type PostgreSQLStorage struct{}
type S3Storage struct{} // Future extension
```

#### 3. Liskov Substitution Principle (LSP)
```go
// All storage implementations must honor the interface contract
type Storage interface {
    Save(ctx context.Context, data []byte) error
}

// Both implementations are interchangeable
type FileStorage struct{}
type DatabaseStorage struct{}
```

#### 4. Interface Segregation Principle (ISP)
```go
// BAD: Fat interface
type DataStore interface {
    SaveLog(log *Log) error
    GetLog(id string) (*Log, error)
    SaveSession(session *Session) error
    GetSession(id string) (*Session, error)
    SaveServer(server *Server) error
    // ... many more methods
}

// GOOD: Focused interfaces
type LogRepository interface {
    Save(ctx context.Context, log *Log) error
    Get(ctx context.Context, id string) (*Log, error)
}

type SessionRepository interface {
    Save(ctx context.Context, session *Session) error
    Get(ctx context.Context, id string) (*Session, error)
}
```

#### 5. Dependency Inversion Principle (DIP)
```go
// High-level modules depend on abstractions
type LogService struct {
    repo   LogRepository    // Interface, not concrete type
    parser LogParser        // Interface, not concrete type
}

// Inject dependencies
func NewLogService(repo LogRepository, parser LogParser) *LogService {
    return &LogService{repo: repo, parser: parser}
}
```

### DRY (Don't Repeat Yourself)

```go
// Shared validation logic
func ValidateIPAddress(ip string) error {
    if net.ParseIP(ip) == nil {
        return ErrInvalidIP
    }
    return nil
}

// Reusable error handling
func HandleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, ErrNotFound):
        c.JSON(404, gin.H{"error": "Not found"})
    case errors.Is(err, ErrUnauthorized):
        c.JSON(403, gin.H{"error": "Unauthorized"})
    default:
        c.JSON(500, gin.H{"error": "Internal error"})
    }
}
```

### Clean Code Patterns

#### Repository Pattern
```go
// internal/domain/repositories/log_repository.go
type LogRepository interface {
    Create(ctx context.Context, log *entities.Log) error
    FindByID(ctx context.Context, id string) (*entities.Log, error)
    FindByServerID(ctx context.Context, serverID string, limit int) ([]*entities.Log, error)
}

// internal/infrastructure/persistence/postgres_log_repository.go
type PostgresLogRepository struct {
    db *sql.DB
}

func (r *PostgresLogRepository) Create(ctx context.Context, log *entities.Log) error {
    // Implementation
}
```

#### Use Case Pattern (Application Services)
```go
// internal/application/commands/ingest_log_command.go
type IngestLogCommand struct {
    ServerID string
    Content  string
    IP       string
}

type IngestLogHandler struct {
    logRepo     repositories.LogRepository
    parser      services.LogParser
    whitelistSvc services.WhitelistService
}

func (h *IngestLogHandler) Handle(ctx context.Context, cmd IngestLogCommand) error {
    // 1. Check IP whitelist
    if !h.whitelistSvc.IsAllowed(cmd.IP) {
        return ErrIPNotWhitelisted
    }
    
    // 2. Save raw log
    log := &entities.Log{
        ServerID: cmd.ServerID,
        Content:  cmd.Content,
    }
    if err := h.logRepo.Create(ctx, log); err != nil {
        return err
    }
    
    // 3. Parse asynchronously
    go h.parseAsync(log)
    
    return nil
}
```

#### Factory Pattern
```go
// Create appropriate storage based on config
func NewLogStorage(config Config) LogStorage {
    switch config.StorageType {
    case "postgres":
        return NewPostgresStorage(config.DatabaseURL)
    case "file":
        return NewFileStorage(config.FilePath)
    default:
        return NewPostgresStorage(config.DatabaseURL)
    }
}
```

### Error Handling

```go
// Define domain errors
package errors

var (
    ErrLogNotFound      = errors.New("log not found")
    ErrIPNotWhitelisted = errors.New("IP not whitelisted")
    ErrParseFailed      = errors.New("failed to parse log")
)

// Wrap errors with context
func (s *LogService) GetLog(ctx context.Context, id string) (*Log, error) {
    log, err := s.repo.Get(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get log %s: %w", id, err)
    }
    return log, nil
}
```

### Testing Patterns

```go
// Use interfaces for easy mocking
type MockLogRepository struct {
    mock.Mock
}

func (m *MockLogRepository) Create(ctx context.Context, log *entities.Log) error {
    args := m.Called(ctx, log)
    return args.Error(0)
}

// Table-driven tests
func TestValidateIP(t *testing.T) {
    tests := []struct {
        name    string
        ip      string
        wantErr bool
    }{
        {"valid IPv4", "192.168.1.1", false},
        {"valid IPv6", "::1", false},
        {"invalid IP", "not-an-ip", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateIP(tt.ip)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateIP() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### MCP Server Integration

#### Using shadcn/ui Components via MCP

**IMPORTANT**: Always use the MCP shadcn server to fetch the latest component implementations. This ensures you're using the most up-to-date, optimized versions of shadcn/ui components.

```typescript
// When implementing UI components, use MCP tools:
// 1. mcp__shadcn-ui__list_components - List available components
// 2. mcp__shadcn-ui__get_component - Get component source code
// 3. mcp__shadcn-ui__get_component_demo - Get usage examples
// 4. mcp__shadcn-ui__get_component_metadata - Get component dependencies

// Example workflow:
// Step 1: Check available components
// Use: mcp__shadcn-ui__list_components

// Step 2: Get the component you need
// Use: mcp__shadcn-ui__get_component with componentName: "card"

// Step 3: Get usage examples
// Use: mcp__shadcn-ui__get_component_demo with componentName: "card"

// This ensures you always have:
// - Latest component versions
// - Proper TypeScript types
// - Accessibility features
// - Consistent styling with Tailwind
```

#### Common shadcn/ui Components for This Project

```typescript
// Essential components to fetch via MCP:
const requiredComponents = [
    'card',           // For server cards, log entries
    'table',          // For log tables, IP whitelist
    'button',         // Action buttons throughout
    'input',          // Form inputs
    'dialog',         // Modals for add/edit operations
    'toast',          // Notifications
    'tabs',           // For log types (raw/parsed/failed)
    'badge',          // Status indicators
    'alert',          // Error/warning messages
    'dropdown-menu',  // Context menus
    'data-table',     // Advanced table with sorting/filtering
    'form',           // Form handling with validation
];

// Always fetch these using MCP before implementing
```

### Frontend Best Practices

#### Component Composition
```typescript
// Composable components
const IPWhitelistManager: React.FC = () => {
    return (
        <Card>
            <CardHeader>
                <IPWhitelistToolbar onAdd={handleAdd} />
            </CardHeader>
            <CardContent>
                <IPWhitelistTable 
                    items={whitelist}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                />
            </CardContent>
        </Card>
    );
};
```

#### Custom Hooks for Logic Reuse
```typescript
// hooks/useIPWhitelist.ts
export function useIPWhitelist() {
    const [whitelist, setWhitelist] = useState<IPEntry[]>([]);
    const [loading, setLoading] = useState(false);
    
    const addIP = async (ip: string) => {
        // Implementation
    };
    
    const removeIP = async (id: string) => {
        // Implementation
    };
    
    return { whitelist, loading, addIP, removeIP };
}
```

#### Service Layer
```typescript
// services/api/whitelist.service.ts
export class WhitelistService {
    private api: ApiClient;
    
    async getAll(): Promise<IPEntry[]> {
        return this.api.get('/api/admin/whitelist');
    }
    
    async add(ip: string, description?: string): Promise<IPEntry> {
        return this.api.post('/api/admin/whitelist', { ip, description });
    }
    
    async remove(id: string): Promise<void> {
        return this.api.delete(`/api/admin/whitelist/${id}`);
    }
}
```

## Common Implementation Patterns

### Dynamic IP Whitelist Middleware (Go/Gin)
```go
func DynamicIPWhitelist(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        // Check database for whitelisted IP
        var exists bool
        err := db.QueryRow(
            "SELECT EXISTS(SELECT 1 FROM ip_whitelist WHERE ip_address = $1 AND enabled = true)",
            clientIP,
        ).Scan(&exists)
        
        if err != nil || !exists {
            // Fallback to environment variable if DB check fails
            if checkEnvWhitelist(clientIP) {
                c.Next()
                return
            }
            c.AbortWithStatus(403)
            return
        }
        
        c.Next()
    }
}
```

### Log Processing Flow
1. Receive raw log → Store immediately
2. Queue for parsing (async)
3. Try parse with cs2-log
4. If success → Store parsed + detect session
5. If fail → Store in failed_parses
6. Never lose data

### Session Detection Logic
- **Match Start**: Look for "World triggered Match_Start"
- **Match End**: Look for "World triggered Match_End" or map change
- **Round Events**: Track round_start, round_end for phases
- **Warmup**: Detect "mp_warmup_end" events

## MCP Server Usage Guidelines

### Available MCP Servers for This Project

**ALWAYS use MCP servers when available** for better code quality and consistency:

1. **mcp__shadcn-ui** - UI Component Library
   - Fetch latest shadcn/ui components
   - Get component demos and examples
   - Ensure TypeScript compatibility

2. **mcp__filesystem** - File Operations
   - Read/write project files
   - Maintain proper structure

3. **mcp__github** - Repository Management
   - Create issues and PRs
   - Manage branches

### Required MCP Usage for UI Components

```typescript
// IMPORTANT: When implementing ANY UI component:

// 1. First check if shadcn/ui has it
await mcp__shadcn-ui__list_components()

// 2. Fetch the component
await mcp__shadcn-ui__get_component({ 
    componentName: "data-table" 
})

// 3. Get usage examples
await mcp__shadcn-ui__get_component_demo({ 
    componentName: "data-table" 
})

// 4. Check dependencies
await mcp__shadcn-ui__get_component_metadata({ 
    componentName: "data-table" 
})
```

### UI Components Required from MCP

For the CS2 Log Saver project, fetch these components via MCP:

```typescript
// Admin Dashboard
- card (server status cards)
- data-table (IP whitelist management)
- dialog (add/edit IP modals)
- form (IP entry forms)
- button (actions)
- badge (status indicators)
- toast (notifications)

// Log Viewer
- table (log display)
- tabs (raw/parsed/failed)
- select (filter options)
- input (search)
- scroll-area (log scrolling)

// Session View
- timeline (match progress)
- progress (round progress)
- separator (visual breaks)
- accordion (collapsible sections)
```

## Development Workflow

### Local Development
```bash
# Start everything (use docker compose v2)
docker compose up -d

# Backend only
cd backend && go run cmd/server/main.go

# Frontend only  
cd frontend && npm run dev

# View logs
docker compose logs -f backend
```

### Testing Log Ingestion
```bash
# Send test log (from whitelisted IP)
curl -X POST http://localhost:9090/logs/testserver \
  -H "Content-Type: text/plain" \
  -d 'L 01/17/2025 - 12:00:00: "Player<1><STEAM_1:0:123456><CT>" killed "Enemy<2><STEAM_1:0:654321><T>" with "ak47"'
```

## Deployment Checklist

- [ ] Set ALLOWED_IPS with actual CS2 server IPs
- [ ] Configure PostgreSQL in Coolify
- [ ] Set up domains with SSL
- [ ] Test log ingestion from real CS2 server
- [ ] Verify parsing works
- [ ] Check UI displays data
- [ ] Monitor for 24 hours

## Performance Expectations

### MVP Targets
- Handle 10,000 log lines/second
- Support 1-10 CS2 servers
- < 100ms ingestion latency
- < 500ms query response
- 99% parse success rate

### Don't Optimize Until
- Actual performance issues observed
- More than 10 servers connected
- Database > 10GB
- Users complain about speed

## Code Quality Standards

### Naming Conventions

#### Go
```go
// Packages: lowercase, no underscores
package logparser

// Interfaces: noun + "er" suffix
type Logger interface {}
type Parser interface {}

// Structs: PascalCase
type LogEntry struct {}

// Methods/Functions: PascalCase for exported, camelCase for private
func ParseLog(raw string) (*Log, error) {}
func validateInput(input string) error {}

// Constants: PascalCase or ALL_CAPS for groups
const MaxRetries = 3
const (
    StatusPending   = "pending"
    StatusCompleted = "completed"
)
```

#### TypeScript/React
```typescript
// Components: PascalCase
const LogViewer: React.FC = () => {}

// Hooks: camelCase with "use" prefix
const useLogData = () => {}

// Types/Interfaces: PascalCase with "I" or "T" prefix (optional)
interface ILogEntry {}
type TLogStatus = 'pending' | 'completed'

// Functions: camelCase
const fetchLogs = async () => {}

// Constants: SCREAMING_SNAKE_CASE or PascalCase
const MAX_RETRIES = 3
const ApiEndpoints = {
    LOGS: '/api/logs',
    SESSIONS: '/api/sessions'
}
```

### Code Organization

#### Keep Functions Small
```go
// BAD: Function doing too much
func ProcessLog(raw string) error {
    // 50+ lines of parsing, validation, storage, etc.
}

// GOOD: Small, focused functions
func ProcessLog(raw string) error {
    log, err := parseLog(raw)
    if err != nil {
        return fmt.Errorf("parse: %w", err)
    }
    
    if err := validateLog(log); err != nil {
        return fmt.Errorf("validate: %w", err)
    }
    
    return storeLog(log)
}
```

#### Avoid Deep Nesting
```go
// BAD: Deep nesting
if condition1 {
    if condition2 {
        if condition3 {
            // Do something
        }
    }
}

// GOOD: Early returns
if !condition1 {
    return nil
}
if !condition2 {
    return nil
}
if !condition3 {
    return nil
}
// Do something
```

### Documentation

```go
// Package logparser provides utilities for parsing CS2 server logs.
package logparser

// ParseLog parses a raw CS2 log line and returns a structured Log.
// It returns an error if the log format is invalid or unrecognized.
//
// Example:
//   log, err := ParseLog("L 01/17/2025 - 12:00:00: Player killed Enemy")
//   if err != nil {
//       // Handle error
//   }
func ParseLog(raw string) (*Log, error) {
    // Implementation
}
```

### Performance Considerations

```go
// Use context for cancellation
func ProcessLogs(ctx context.Context, logs []string) error {
    for _, log := range logs {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Process log
        }
    }
    return nil
}

// Preallocate slices when size is known
logs := make([]Log, 0, expectedCount)

// Use string builder for concatenation
var sb strings.Builder
sb.WriteString("prefix")
sb.WriteString(variable)
result := sb.String()
```

## Common Pitfalls to Avoid

1. **Don't add authentication** - CS2 servers can't use it
2. **Don't over-engineer** - Ship simple, iterate later
3. **Don't optimize prematurely** - Wait for real bottlenecks
4. **Don't lose raw logs** - Always store original data
5. **Don't block on parsing** - Process asynchronously
6. **Don't forget IP whitelist** - Primary security mechanism
7. **Don't ignore error handling** - Always handle and log errors appropriately
8. **Don't skip tests** - Write tests for critical paths
9. **Don't hardcode values** - Use configuration and environment variables
10. **Don't ignore SOLID principles** - Maintain clean, maintainable code

## Quick Commands Reference

```bash
# Build and run (Docker Compose v2 syntax)
docker compose up --build

# Deploy to production
git push origin main  # Coolify auto-deploys

# Check logs
docker compose logs -f backend

# Database console
docker compose exec postgres psql -U cs2admin -d cs2logs

# Run migrations
docker compose exec backend /app/migrate up

# Clean restart
docker compose down -v && docker compose up
```

## Success Metrics

Track these to know if the project is working:
1. **Uptime**: > 99% availability
2. **Parse Rate**: > 95% successful parses
3. **Data Loss**: 0% after acknowledgment
4. **Response Time**: < 500ms for UI queries
5. **Storage Growth**: < 1GB per million logs

## When to Ask for Help

Contact the team if:
- Parse success rate < 90%
- Losing log data
- Database queries > 1 second
- Can't connect CS2 servers
- Coolify deployment fails

## Future Enhancements (After MVP)

Only consider these after shipping and gathering feedback:
- Real-time WebSocket updates
- Advanced match analytics
- Player statistics tracking
- Heat maps and visualizations
- Export functionality
- Email alerts
- S3 storage migration
- Horizontal scaling

---

**Remember**: The goal is to ship a working product in 2-3 weeks. Keep it simple, make it work, deploy it, then iterate based on real usage. Perfect is the enemy of shipped.