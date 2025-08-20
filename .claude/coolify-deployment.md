# Coolify Deployment Guide

## Prerequisites

- [ ] Coolify instance running (v4.0+)
- [ ] GitHub repository created
- [ ] Domain/subdomain configured in DNS
- [ ] Basic understanding of Docker

## Step-by-Step Deployment

### Step 1: Prepare Your Repository

1. **Push code to GitHub:**
```bash
cd /path/to/nocs-log-saver
git init
git add .
git commit -m "Initial commit - CS2 Log Saver MVP"
git branch -M main
git remote add origin https://github.com/noueii/nocs-log-saver.git
git push -u origin main
```

2. **Ensure these files exist in repo root:**
- `docker-compose.yml` (for Coolify to detect)
- `.env.example` (for reference)

### Step 2: Create docker-compose.yml for Coolify

Create this in your repository root:

```yaml
# Compose Specification (no version field needed)
services:
  backend:
    build: 
      context: ./backend
      dockerfile: Dockerfile
    restart: unless-stopped
    ports:
      - "${BACKEND_PORT:-9090}:9090"
    environment:
      DATABASE_URL: ${DATABASE_URL}
      ALLOWED_IPS: ${ALLOWED_IPS}
      LOG_LEVEL: ${LOG_LEVEL:-info}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - cs2logs

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    restart: unless-stopped
    ports:
      - "${FRONTEND_PORT:-6173}:6173"
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
    depends_on:
      - backend
    networks:
      - cs2logs

networks:
  cs2logs:
    driver: bridge
```

### Step 3: Configure Coolify Project

1. **Log into Coolify Dashboard**

2. **Create New Project:**
   - Click "New Project"
   - Name: "CS2 Log Saver"
   - Description: "CS2 server log aggregation and parsing"

3. **Add New Resource:**
   - Select "Docker Compose"
   - Choose "Public Repository" (or Private with GitHub integration)
   - Repository URL: `https://github.com/noueii/nocs-log-saver`
   - Branch: `main`

### Step 4: Set Up PostgreSQL in Coolify

1. **Add PostgreSQL Service:**
   - Go to your project
   - Click "New Resource" → "Database"
   - Select "PostgreSQL"
   - Version: 17
   - Name: `cs2logs-db`

2. **Configure Database:**
```
Database Name: cs2logs
Username: cs2admin
Password: [Generate strong password]
Port: 5432
```

3. **Note the connection details** - You'll need the internal connection string

### Step 5: Configure Environment Variables

In Coolify, set these environment variables for your Docker Compose stack:

```bash
# Database (use Coolify's internal PostgreSQL URL)
DATABASE_URL=postgresql://cs2admin:YOUR_PASSWORD@cs2logs-db:5432/cs2logs

# Security Configuration (Initial IP Whitelist)
# After first deployment, manage IPs through the admin dashboard at /admin
# This is used as fallback if database is empty
ALLOWED_IPS=192.168.1.100,203.0.113.45,10.0.0.5
BACKEND_PORT=9090
FRONTEND_PORT=6173

# Frontend API URL (use your domain)
NEXT_PUBLIC_API_URL=https://api.yourdomain.com

# Logging
LOG_LEVEL=info
```

### Step 6: Configure Domains

1. **Backend API Domain:**
   - Domain: `api.yourdomain.com` or `yourdomain.com/api`
   - Port: 9090
   - Enable HTTPS (Coolify handles SSL automatically)

2. **Frontend Domain:**
   - Domain: `yourdomain.com` or `logs.yourdomain.com`
   - Port: 6173
   - Enable HTTPS

### Step 7: Build Configuration

1. **Set Build Configuration:**
   - Build Pack: Docker Compose
   - Base Directory: `/` (root)
   - Docker Compose File: `docker-compose.yml`

2. **Resource Limits (Optional but Recommended):**
```
Backend:
- CPU: 1 core
- Memory: 512MB

Frontend:
- CPU: 0.5 core
- Memory: 256MB

PostgreSQL:
- CPU: 1 core
- Memory: 1GB
```

### Step 8: Deploy

1. **Initial Deployment:**
   - Click "Deploy"
   - Watch build logs for errors
   - Wait for health checks to pass

2. **Verify Deployment:**
```bash
# Check backend health
curl https://api.yourdomain.com/health

# Check frontend
curl https://yourdomain.com
```

### Step 9: Database Migrations

After first deployment, run migrations:

1. **SSH into Coolify server or use Coolify's terminal:**
```bash
# Connect to backend container
docker exec -it [backend-container-id] sh

# Run migrations (adjust based on your migration tool)
./migrate up
```

### Step 10: Configure CS2 Servers

Configure your CS2 servers to send logs:

```cfg
# In CS2 server config
log on
logaddress_add_http "https://api.yourdomain.com/logs/server1"
# Note: CS2 servers don't support authentication headers
# Security is handled via IP whitelisting
```

## Environment Variables Reference

### Required Variables
```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/dbname

# IP Whitelist (comma-separated list of allowed server IPs)
ALLOWED_IPS=192.168.1.100,203.0.113.45

# Frontend API URL
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

### Optional Variables
```bash
# Ports
BACKEND_PORT=9090
FRONTEND_PORT=6173

# Logging
LOG_LEVEL=debug|info|warn|error

# Performance
MAX_CONNECTIONS=100
REQUEST_TIMEOUT=30

# Storage
MAX_LOG_SIZE_MB=10
LOG_RETENTION_DAYS=30
```

## Security Configuration

### IP Whitelisting Setup
Since CS2 servers cannot send authentication headers, we use IP-based security:

1. **Initial Setup - Find your CS2 server IPs:**
```bash
# From your CS2 server
curl ifconfig.me
```

2. **Configure initial ALLOWED_IPS in Coolify:**
```bash
ALLOWED_IPS=203.0.113.45,198.51.100.12,192.0.2.33
```

3. **After Deployment - Use Admin Dashboard:**
- Navigate to `https://yourdomain.com/admin`
- Go to "IP Whitelist" section
- Add/remove/edit allowed IPs
- Associate IPs with specific servers
- Changes apply immediately without restart

3. **Optional: Use Cloudflare for additional protection:**
- Add your domain to Cloudflare
- Enable DDoS protection
- Set up rate limiting rules
- Use Cloudflare's IP ranges for internal whitelisting

### Rate Limiting with Coolify/Caddy
Add to your Caddy configuration:
```
rate_limit {
    zone dynamic_zone {
        key {remote_host}
        events 100
        window 60s
    }
}
```

## Monitoring in Coolify

### Health Checks
Coolify automatically monitors these endpoints:
- Backend: `http://backend:9090/health`
- Frontend: `http://frontend:6173/api/health`

### Logs
View logs in Coolify dashboard:
1. Go to your deployment
2. Click "Logs" tab
3. Filter by service (backend/frontend/database)

### Metrics
Coolify provides basic metrics:
- CPU usage
- Memory usage
- Network I/O
- Disk usage

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check DATABASE_URL format
   - Verify PostgreSQL service is running
   - Check network connectivity between services

2. **Frontend Can't Reach Backend**
   - Verify NEXT_PUBLIC_API_URL is correct
   - Check CORS settings in backend
   - Ensure backend health check passes

3. **SSL/HTTPS Issues**
   - Let Coolify manage SSL certificates
   - Ensure domain DNS points to Coolify server
   - Wait for certificate provisioning (can take 5-10 minutes)

4. **Build Failures**
   - Check Dockerfile syntax
   - Verify all dependencies are installed
   - Review build logs in Coolify

### Debug Commands

```bash
# Check service status (Docker Compose v2)
docker compose ps

# View backend logs
docker compose logs backend --tail 100

# Test database connection
docker compose exec backend sh
psql $DATABASE_URL -c "SELECT 1"

# Check API endpoint (from whitelisted IP)
curl -X POST https://api.yourdomain.com/logs/test \
  -H "Content-Type: text/plain" \
  -d "test log entry"
```

## Backup Strategy

### Database Backups
1. **Configure in Coolify:**
   - Go to PostgreSQL service
   - Enable automatic backups
   - Set schedule (daily recommended)
   - Configure retention (7-30 days)

2. **Manual Backup:**
```bash
docker exec [postgres-container] pg_dump cs2logs > backup_$(date +%Y%m%d).sql
```

### Application Backups
- Code is in GitHub (version controlled)
- Logs stored in database (backed up above)
- Consider backing up environment variables

## Updating Your Application

### Deploy Updates
1. **Push changes to GitHub:**
```bash
git add .
git commit -m "Update: description"
git push origin main
```

2. **In Coolify:**
   - Go to your deployment
   - Click "Redeploy" or enable auto-deploy
   - Monitor deployment logs

### Rolling Updates
Coolify supports zero-downtime deployments:
1. New containers are started
2. Health checks must pass
3. Traffic switches to new containers
4. Old containers are stopped

## Performance Tuning

### Quick Optimizations
1. **Enable Coolify's built-in caching**
2. **Set appropriate resource limits**
3. **Use PostgreSQL connection pooling**
4. **Enable gzip compression**

### Scaling
When ready to scale:
1. **Horizontal Scaling:**
   - Increase replica count in Coolify
   - Add load balancer

2. **Vertical Scaling:**
   - Increase CPU/Memory limits
   - Upgrade Coolify server

## Security Checklist

- [ ] Strong DATABASE_URL password
- [ ] Configure IP whitelist for CS2 servers
- [ ] HTTPS enabled (automatic with Coolify)
- [ ] Environment variables not in code
- [ ] Regular backups configured
- [ ] Monitor logs for suspicious activity
- [ ] Keep Docker images updated
- [ ] Rate limiting configured at proxy level
- [ ] Consider Cloudflare for additional DDoS protection

## Support Resources

- **Coolify Documentation**: https://coolify.io/docs
- **Coolify Discord**: https://discord.gg/coolify
- **Project Issues**: https://github.com/noueii/nocs-log-saver/issues

---

**Pro Tips:**
- Start with minimal resources and scale up as needed
- Use Coolify's preview deployments for testing
- Enable notifications for deployment status
- Keep your Coolify instance updated