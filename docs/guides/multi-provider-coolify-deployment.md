# Multi-Provider RADIO Deployment on Coolify

> Deploy RADIO with multi-layer caching for **100 providers × 5000 users × 1500 concurrent**

---

## Prerequisites

1. **Coolify Instance** with:
   - Docker 24+
   - PostgreSQL 16+
   - Redis 7+
   - At least 4GB RAM, 4 vCPUs

2. **GitHub Repository** with RADIO source code

3. **GitHub Secrets** configured:
   - `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN`
   - `COOLIFY_SECRET`
   - `COOLIFY_STAGING_WEBHOOK_URL`
   - `COOLIFY_PRODUCTION_WEBHOOK_URL`
   - `STAGING_URL`

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Actions                            │
│         (Build → Push → Trigger Coolify)                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   Coolify Instance                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  RADIO #1   │  │  RADIO #2   │  │  RADIO #N   │        │
│  │  (L1+L2)    │  │  (L1+L2)    │  │  (L1+L2)    │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         │                │                │                  │
└─────────┼────────────────┼────────────────┼──────────────────┘
          │                │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL Cluster                        │
│                  (Shared Database)                          │
└─────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Redis Cluster                          │
│              (L2 Cache - Shared Cache)                      │
└─────────────────────────────────────────────────────────────┘
```

---

## Deployment Steps

### 1. Create PostgreSQL Database

```bash
# SSH into Coolify server
ssh coolify@your-server

# Create database
docker exec -it postgres psql -U postgres -c "CREATE DATABASE toughradius;"
docker exec -it postgres psql -U postgres -c "CREATE USER toughradius WITH PASSWORD 'strong-password';"
docker exec -it postgres psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE toughradius TO toughradius;"

# Create Redis database (optional, using default DB 0)
# Redis is configured in docker-compose.yml
```

### 2. Create Coolify Application

1. Go to **Coolify Dashboard**
2. Click **"New Resource"** → **"Application"**
3. Connect GitHub repository
4. Select branch (main for staging, release/* for production)
5. Configure build:
   - **Build Pack**: Dockerfile
   - **Dockerfile Path**: `./Dockerfile`
6. Add environment variables:

```env
# System Configuration
TOUGHRADIUS_SYSTEM_WORKER_DIR=/var/toughradius
TOUGHRADIUS_SYSTEM_APPID=RADIO
TOUGHRADIUS_SYSTEM_DOMAIN=https://radius.your-domain.com
TOUGHRADIUS_LOGGER_MODE=production
TOUGHRADIUS_SYSTEM_DEBUG=false

# Database Configuration
TOUGHRADIUS_DB_TYPE=postgres
TOUGHRADIUS_DB_HOST=postgres
TOUGHRADIUS_DB_PORT=5432
TOUGHRADIUS_DB_NAME=toughradius
TOUGHRADIUS_DB_USER=toughradius
TOUGHRADIUS_DB_PASSWD=strong-password
TOUGHRADIUS_DB_MAX_CONN=200
TOUGHRADIUS_DB_IDLE_CONN=20

# Multi-Tenant Configuration
TOUGHRADIUS_MULTITENANT_ENABLED=true

# Cache Configuration (L1 + L2)
TOUGHRADIUS_CACHE_L1_ENABLED=true
TOUGHRADIUS_CACHE_L1_MAX_ENTRIES=10000
TOUGHRADIUS_CACHE_L2_ENABLED=true
TOUGHRADIUS_REDIS_HOST=redis
TOUGHRADIUS_REDIS_PORT=6379
TOUGHRADIUS_CACHE_USER_TTL=30
TOUGHRADIUS_CACHE_NAS_TTL=300
TOUGHRADIUS_CACHE_SESSION_TTL=5

# Scale Configuration
TOUGHRADIUS_CACHE_PROVIDER_COUNT=100
TOUGHRADIUS_CACHE_USERS_PER_PROVIDER=5000
TOUGHRADIUS_CACHE_CONCURRENT_PER_PROVIDER=1500
```

### 3. Configure Ports

| Internal Port | External Port | Protocol | Purpose |
|---------------|---------------|----------|---------|
| 1816 | 1816 | TCP | Web/Admin API |
| 1812 | 1812 | UDP | RADIUS Auth |
| 1813 | 1813 | UDP | RADIUS Acct |
| 2083 | 2083 | TCP | RadSec |

### 4. Configure Volumes

| Source | Destination |
|--------|------------|
| `radio_data` | `/var/toughradius` |

### 5. Set Up Health Check

```
URL: http://localhost:1816/ready
Interval: 30s
Timeout: 10s
Retries: 3
```

### 6. Configure Webhook for Auto-Deploy

1. In Coolify, go to **Application → Deployments → Webhooks**
2. Copy the webhook URL
3. In GitHub, go to **Settings → Webhooks → Add webhook**:
   - Payload URL: `<coolify-webhook-url>`
   - Content type: `application/json`
   - Events: **Push**, **Tag**

---

## Auto-Deploy Configuration

### GitHub Secrets

| Secret | Description |
|--------|-------------|
| `DOCKERHUB_USERNAME` | Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token |
| `COOLIFY_SECRET` | Secret for webhook authentication |
| `COOLIFY_STAGING_WEBHOOK_URL` | Staging deploy webhook URL |
| `COOLIFY_PRODUCTION_WEBHOOK_URL` | Production deploy webhook URL |
| `STAGING_URL` | Staging environment URL |

### Deployment Triggers

| Event | Action |
|-------|--------|
| Push to `main` | Deploys to staging |
| Push to `release/*` | Deploys to staging |
| New tag `v*` | Deploys to production |
| Manual workflow dispatch | Deploys to selected environment |

---

## Performance Tuning

### PostgreSQL Configuration

```sql
-- postgresql.conf
max_connections = 200
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 512MB
work_mem = 16MB
random_page_cost = 1.1
effective_io_concurrency = 200
max_worker_processes = 8
max_parallel_workers_per_gather = 4
```

### Redis Configuration

```bash
# In docker-compose.yml or Redis config
redis-server \
  --maxmemory 1gb \
  --maxmemory-policy allkeys-lru \
  --tcp-backlog 511 \
  --timeout 0 \
  --tcp-keepalive 300
```

### L1 Cache (Local Memory)

```env
# Per-container memory cache
TOUGHRADIUS_CACHE_L1_ENABLED=true
TOUGHRADIUS_CACHE_L1_MAX_ENTRIES=10000
```

---

## Monitoring

### Health Check Endpoint

```bash
curl https://your-domain.com/ready
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "24h",
  "database": "connected",
  "cache": "connected",
  "tenants": 10,
  "sessions": 5000
}
```

### Prometheus Metrics

```bash
curl https://your-domain.com/metrics
```

Key metrics:
- `radius_auth_requests_total{tenant_id, result}`
- `radius_acct_requests_total{tenant_id, result}`
- `active_sessions{tenant_id}`
- `cache_hit_ratio{cache_type}`

---

## Scaling Guide

### Horizontal Scaling (Multiple Containers)

For higher concurrency, run multiple RADIO containers behind a load balancer:

1. **Sticky Sessions**: Not required (stateless + Redis)
2. **Database Connection Pooling**: 200 connections shared
3. **Redis**: Single point for L2 cache

### Vertical Scaling

| Concurrent Users | RAM | vCPUs | Redis |
|-----------------|-----|-------|-------|
| 10,000 | 4GB | 2 | 512MB |
| 50,000 | 8GB | 4 | 1GB |
| 150,000 | 16GB | 8 | 2GB |

---

## Rollback Strategy

### Docker Image Rollback

```bash
# List available tags
docker images farisnoaman/toughradius

# Pull specific version
docker pull farisnoaman/toughradius:v1.0.0

# Update Coolify to use specific image tag
# Go to: Coolify → Application → Environment → Change image tag
```

### Database Rollback

```bash
# Restore from backup
docker exec -it toughradius-app-1 toughradius -restore /var/toughradius/backup/backup_20240101.sql
```

---

## Success Metrics

| Metric | Target | Notes |
|--------|--------|-------|
| Concurrent Users | 150,000 | 100 providers × 1500 |
| Auth Latency (p99) | <50ms | L1 cache hit |
| Acct Latency (p99) | <20ms | L1 cache hit |
| Cache Hit Rate | >95% | L1 + L2 combined |
| API Latency (p99) | <100ms | Web API |
| Uptime | 99.9% | Multi-container HA |
