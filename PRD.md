# Product Requirements Document (PRD)
# CS2 Log Saver Service

## 1. Executive Summary

### 1.1 Project Overview
The CS2 Log Saver is a centralized log management service designed to receive, parse, store, and organize Counter-Strike 2 (CS2) server logs from multiple game servers. The service provides real-time log ingestion through dynamic HTTP endpoints, intelligent parsing using the cs2-log Go library, and comprehensive session-based organization.

### 1.2 Key Objectives
- **Centralized Log Collection**: Aggregate logs from multiple CS2 servers into a single, manageable system
- **Intelligent Parsing**: Automatically parse CS2 logs to extract meaningful game events and statistics
- **Session Management**: Organize logs by server sessions, match sessions, and game phases (warmup, live, intermission)
- **Data Preservation**: Store both raw and parsed logs, ensuring no data loss even when parsing fails
- **Real-time Processing**: Handle log streams in real-time with minimal latency
- **Scalability**: Support growing numbers of CS2 servers without performance degradation

## 2. Problem Statement

### 2.1 Current Challenges
- **Fragmented Log Management**: CS2 servers generate extensive logs that are typically stored locally, making centralized analysis difficult
- **Lack of Structure**: Raw CS2 logs are text-based and require parsing to extract meaningful information
- **Session Tracking**: Difficulty in organizing logs by game sessions, matches, and phases
- **Data Loss Risk**: Logs may be lost during server restarts or crashes without proper collection
- **Manual Analysis**: Without automated parsing, log analysis is time-consuming and error-prone

### 2.2 Solution Benefits
- Centralized repository for all CS2 server logs
- Automated parsing and categorization
- Historical data preservation for analysis and auditing
- Real-time visibility into server activities
- Foundation for advanced analytics and monitoring

## 3. Functional Requirements

### 3.1 Core Features

#### 3.1.1 Dynamic Endpoint Routing
- **Requirement**: Support multiple CS2 servers sending logs to unique endpoints
- **Implementation**: Dynamic route pattern `/logs/{server_id}` where server_id uniquely identifies each CS2 server
- **Validation**: Verify server_id against whitelist of authorized servers

#### 3.1.2 Log Reception
- **Protocol**: HTTP/HTTPS POST endpoints
- **Format**: Accept raw text logs, line-delimited or batch submissions
- **Buffering**: Implement request buffering for high-volume log streams
- **Acknowledgment**: Return confirmation receipts for received logs

#### 3.1.3 Log Storage Types

##### Raw Logs
- **Purpose**: Preserve original, unmodified log data
- **Storage**: Time-series database or file system with compression
- **Retention**: Configurable retention period (default: 90 days)
- **Indexing**: Timestamp and server_id based indexing

##### Parsed Logs
- **Purpose**: Store structured, parsed log data
- **Parser**: Integration with cs2-log Go library
- **Schema**: Structured format based on CS2 event types
- **Enrichment**: Add metadata like timestamps, server info, session IDs

##### Failed Parse Logs
- **Purpose**: Capture logs that couldn't be parsed
- **Storage**: Separate collection/directory for investigation
- **Metadata**: Include parse error details and timestamps
- **Retry Logic**: Implement retry mechanism for transient failures

#### 3.1.4 Session Management

##### Server Sessions
- **Definition**: Period from server start to server stop/restart
- **Tracking**: Monitor server lifecycle events
- **Metadata**: Server version, configuration, start/stop times

##### Match Sessions
- **Definition**: Individual competitive matches or games
- **Boundaries**: Match start/end events, map changes
- **Data**: Map name, team compositions, scores, duration

##### Game Phases
- **Warmup**: Pre-match preparation period
- **Live**: Active competitive gameplay
- **Halftime**: Mid-match break
- **Overtime**: Extended play periods
- **Post-match**: Results and statistics phase

### 3.2 Log Parsing Features

#### 3.2.1 Event Types
- Player actions (kills, deaths, assists, damage)
- Round events (start, end, MVP, bomb plants/defuses)
- Economic events (purchases, money awards)
- Team events (side switches, timeouts)
- Server events (map changes, configuration updates)
- Chat messages (team, all, admin)

#### 3.2.2 Parse Pipeline
1. **Ingestion**: Receive raw log lines
2. **Validation**: Check log format and structure
3. **Parsing**: Apply cs2-log library parsing rules
4. **Transformation**: Convert to structured format
5. **Enrichment**: Add contextual metadata
6. **Storage**: Route to appropriate storage based on parse result

## 4. Technical Architecture

### 4.1 System Components

#### 4.1.1 API Gateway
- **Technology**: Go HTTP server (net/http or Gin/Echo framework)
- **Load Balancing**: Support for horizontal scaling
- **Rate Limiting**: Per-server rate limits to prevent abuse
- **Authentication**: API key or token-based authentication

#### 4.1.2 Log Processing Pipeline
```
[CS2 Servers] → [API Gateway] → [Message Queue] → [Parser Service] → [Storage Layer]
                                        ↓
                                [Raw Log Storage]
```

#### 4.1.3 Storage Architecture
- **Raw Logs**: Object storage (S3-compatible) or time-series DB
- **Parsed Logs**: PostgreSQL 17 for structured data
- **Failed Logs**: Separate storage with retry queue
- **Session Metadata**: Redis for active sessions, PostgreSQL for historical

#### 4.1.4 Parser Service
- **Library**: github.com/janstuemmel/cs2-log or github.com/joao-silva1007/cs2-log-re2 integration
- **Concurrency**: Worker pool for parallel processing
- **Error Handling**: Graceful degradation on parse failures
- **Monitoring**: Parse success/failure metrics

### 4.2 Data Flow

1. **Log Reception**
   - CS2 server sends logs to `/logs/{server_id}`
   - API validates and authenticates request
   - Raw logs stored immediately

2. **Asynchronous Processing**
   - Logs queued for parsing
   - Parser service processes queue
   - Successful parses stored in structured DB
   - Failed parses logged with errors

3. **Session Association**
   - Parsed events associated with active sessions
   - Session boundaries detected from log events
   - Metadata updated in real-time

## 5. Data Models

### 5.1 Raw Log Schema
```json
{
  "id": "uuid",
  "server_id": "string",
  "timestamp": "ISO8601",
  "content": "string (raw log line)",
  "batch_id": "uuid (optional)",
  "received_at": "ISO8601"
}
```

### 5.2 Parsed Log Schema
```json
{
  "id": "uuid",
  "server_id": "string",
  "timestamp": "ISO8601",
  "event_type": "string",
  "session_id": "uuid",
  "match_id": "uuid (optional)",
  "round_number": "integer (optional)",
  "game_phase": "enum (warmup|live|halftime|overtime|post_match)",
  "event_data": {
    // Event-specific fields based on cs2-log parsing
  },
  "raw_log_id": "uuid (reference)",
  "parsed_at": "ISO8601"
}
```

### 5.3 Session Schema
```json
{
  "id": "uuid",
  "server_id": "string",
  "session_type": "enum (server|match)",
  "start_time": "ISO8601",
  "end_time": "ISO8601 (nullable)",
  "status": "enum (active|completed|terminated)",
  "metadata": {
    "map": "string (for match sessions)",
    "teams": ["array of team objects"],
    "config": "object (server configuration)",
    "final_score": "object (for completed matches)"
  }
}
```

### 5.4 Failed Parse Schema
```json
{
  "id": "uuid",
  "server_id": "string",
  "timestamp": "ISO8601",
  "raw_content": "string",
  "error_type": "string",
  "error_message": "string",
  "retry_count": "integer",
  "last_retry": "ISO8601",
  "resolved": "boolean"
}
```

## 6. API Specifications

### 6.1 Endpoints

#### 6.1.1 Log Ingestion
```
POST /logs/{server_id}
Headers:
  - Authorization: Bearer {api_key}
  - Content-Type: text/plain or application/json
Body: Raw log data (line-delimited or JSON array)
Response: 
  - 200 OK: {"received": true, "batch_id": "uuid", "line_count": 150}
  - 401 Unauthorized: {"error": "Invalid API key"}
  - 429 Too Many Requests: {"error": "Rate limit exceeded"}
```

#### 6.1.2 Query APIs
```
GET /logs/raw?server_id={id}&from={timestamp}&to={timestamp}
GET /logs/parsed?server_id={id}&event_type={type}&session_id={id}
GET /logs/failed?server_id={id}&unresolved=true
GET /sessions?server_id={id}&type={server|match}&status={active|completed}
GET /sessions/{session_id}/logs
```

#### 6.1.3 Management APIs
```
POST /servers/register
Body: {"server_id": "string", "name": "string", "ip": "string"}

PUT /logs/failed/{id}/retry
Response: {"retrying": true, "queue_position": 42}

GET /stats/parsing
Response: {"total_processed": 1000000, "success_rate": 0.98, "failed_count": 2000}
```

### 6.2 Authentication
- **Method**: Bearer token in Authorization header
- **Token Management**: Admin API for token generation/revocation
- **Scopes**: Read-only vs. write permissions per server

### 6.3 Rate Limiting
- **Default**: 1000 requests/minute per server_id
- **Burst**: Allow short bursts up to 100 requests/second
- **Headers**: Return rate limit status in response headers

## 7. Non-Functional Requirements

### 7.1 Performance
- **Throughput**: Handle 10,000+ log lines per second
- **Latency**: < 100ms for log ingestion acknowledgment
- **Parsing Speed**: Process logs within 5 seconds of receipt
- **Query Performance**: < 500ms for indexed queries

### 7.2 Scalability
- **Horizontal Scaling**: Support adding processing nodes
- **Storage Scaling**: Implement data partitioning by date/server
- **Queue Management**: Auto-scale based on queue depth

### 7.3 Reliability
- **Uptime**: 99.9% availability SLA
- **Data Durability**: No data loss after acknowledgment
- **Failover**: Automatic failover for critical components
- **Backup**: Daily backups with 30-day retention

### 7.4 Security
- **Encryption**: TLS 1.3 for data in transit
- **Authentication**: Secure API key management
- **Authorization**: Role-based access control
- **Audit Logging**: Track all API access and modifications
- **Data Privacy**: Comply with data protection regulations

### 7.5 Monitoring & Observability
- **Metrics**: Prometheus-compatible metrics endpoint
- **Logging**: Structured logging with log levels
- **Tracing**: Distributed tracing for request flow
- **Alerts**: Configurable alerts for failures and anomalies
- **Dashboard**: Grafana dashboards for system health

### 7.6 Operational Requirements
- **Deployment**: Containerized with Docker/Kubernetes support
- **Configuration**: Environment-based configuration
- **Health Checks**: Liveness and readiness probes
- **Graceful Shutdown**: Complete processing before termination

## 8. Implementation Roadmap

### 8.1 Phase 1: MVP (Weeks 1-4)
- [ ] Basic API gateway with dynamic routing
- [ ] Raw log storage implementation
- [ ] Simple cs2-log parser integration
- [ ] File-based storage for all log types
- [ ] Basic session detection (server sessions only)
- [ ] Simple authentication (API keys)

### 8.2 Phase 2: Core Features (Weeks 5-8)
- [ ] Message queue implementation
- [ ] Database storage for parsed logs
- [ ] Match session detection
- [ ] Failed parse handling with retry
- [ ] Basic query APIs
- [ ] Rate limiting

### 8.3 Phase 3: Advanced Features (Weeks 9-12)
- [ ] Game phase detection (warmup, live, etc.)
- [ ] Advanced querying with filters
- [ ] Performance optimizations
- [ ] Monitoring and metrics
- [ ] Admin dashboard
- [ ] Backup and recovery

### 8.4 Phase 4: Production Ready (Weeks 13-16)
- [ ] Security hardening
- [ ] Load testing and optimization
- [ ] Documentation and API specs
- [ ] Deployment automation
- [ ] SLA monitoring
- [ ] Disaster recovery testing

## 9. Success Metrics

### 9.1 Technical KPIs
- **Parse Success Rate**: > 95% of logs successfully parsed
- **Ingestion Latency**: P99 < 200ms
- **Query Response Time**: P95 < 500ms
- **System Uptime**: > 99.9%
- **Storage Efficiency**: < 20% storage overhead vs raw logs

### 9.2 Business KPIs
- **Server Coverage**: 100% of registered servers sending logs
- **Data Completeness**: < 0.1% data loss
- **API Adoption**: Active use by monitoring/analytics tools
- **Operational Cost**: < $X per million logs processed

### 9.3 Quality Metrics
- **Code Coverage**: > 80% test coverage
- **Bug Rate**: < 2 critical bugs per release
- **Recovery Time**: < 5 minutes for critical failures
- **Documentation**: 100% API endpoint documentation

## 10. Risk Assessment

### 10.1 Technical Risks
- **Risk**: cs2-log library parsing failures
  - **Mitigation**: Maintain raw logs, implement custom parsers for critical events

- **Risk**: Storage capacity exceeded
  - **Mitigation**: Implement data retention policies, compression, and archival

- **Risk**: Performance degradation at scale
  - **Mitigation**: Load testing, horizontal scaling strategy, caching layer

### 10.2 Operational Risks
- **Risk**: Server authentication compromise
  - **Mitigation**: Regular key rotation, IP whitelisting, audit logging

- **Risk**: Data loss during system failure
  - **Mitigation**: Redundant storage, real-time replication, backup strategy

## 11. Dependencies

### 11.1 External Libraries
- **cs2-log**: github.com/janstuemmel/cs2-log for CS2 log parsing (critical dependency)
- **Database Drivers**: PostgreSQL 17 Go drivers
- **HTTP Framework**: Gin (requires Go 1.23+) for API server
- **Message Queue**: RabbitMQ or NATS client libraries
- **Frontend**: Next.js 15.3 with React 19, Node.js 22 LTS

### 11.2 Infrastructure
- **Compute**: Kubernetes cluster or VM infrastructure
- **Storage**: Object storage (S3-compatible) and PostgreSQL 17 instances
- **Networking**: Load balancer, CDN for static assets
- **Monitoring**: Prometheus, Grafana, ELK stack
- **Container**: Docker with Compose v2 (Compose Specification)

## 12. Appendices

### A. Glossary
- **CS2**: Counter-Strike 2
- **Parse**: Convert raw log text to structured data
- **Session**: Logical grouping of related log entries
- **Warmup**: Pre-match practice period in CS2

### B. References
- CS2 Server Documentation
- cs2-log Library Documentation: https://github.com/path/to/cs2-log
- Go Best Practices
- RESTful API Design Guidelines

### C. Change Log
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2025-01-17 | System | Initial PRD creation |

---

**Document Status**: Draft  
**Last Updated**: 2025-01-17  
**Next Review**: TBD