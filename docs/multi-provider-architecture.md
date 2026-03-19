# Multi-Provider Cloud Architecture for RADIO

> Scalable ISP Management System supporting **100 providers × 1,000 concurrent users = 100,000 total concurrent users**

---

## 1. Architecture Overview

### 1.1 Design Goals

| Metric | Target |
|--------|--------|
| Concurrent Users | 100,000 (100 providers × 1,000 users) |
| RADIUS Auth Throughput | ~2,000 req/s (200 req/s per provider × 100) |
| RADIUS Acct Throughput | ~5,000 req/s |
| API Requests | ~1,000 req/s |
| Database Connections | 200 pooled connections |
| Cache Hit Rate | >95% |

### 1.2 System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Load Balancer                                │
│                    (Cloudflare/HAProxy)                              │
└─────────────────────────────────────────────────────────────────────┘
                    │                    │
                    ▼                    ▼
┌──────────────────────────────────┐  ┌─────────────────────────────┐
│     Coolify Instance #1          │  │     Coolify Instance #2       │
│  ┌──────────────────────────┐   │  │  ┌─────────────────────────┐  │
│  │  RADIO Container (x2)     │   │  │  │  RADIO Container (x2)   │  │
│  │  ┌────────────────────┐   │   │  │  │                         │  │
│  │  │  Web/API (1816)    │   │   │  │  │  (Regional failover)    │  │
│  │  │  RADIUS Auth(1812) │   │   │  │  │                         │  │
│  │  │  RADIUS Acct(1813) │   │   │  │  │                         │  │
│  │  │  RadSec (2083)     │   │   │  │  │                         │  │
│  │  └────────────────────┘   │   │  │  └─────────────────────────┘  │
│  └──────────────────────────┘   │  └───────────────────────────────┘
└──────────────────────────────────┘
                    │                    │
                    ▼                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      PostgreSQL Cluster                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                │
│  │  Primary DB │◄─┤ Replica #1  │──┤ Replica #2  │                │
│  └─────────────┘  └─────────────┘  └─────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Redis Cluster                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                │
│  │  Primary    │──┤ Replica #1  │──┤ Sentinel    │                │
│  └─────────────┘  └─────────────┘  └─────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. Database Schema Changes

### 2.1 New Provider (Tenant) Model

```go
// internal/domain/tenant.go
package domain

import "time"

type Provider struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    Code        string     `json:"code" gorm:"uniqueIndex;size:50;not null"`  // e.g., "isp-alpha"
    Name        string     `json:"name" gorm:"size:255;not null"`             // e.g., "Alpha ISP"
    Status      string     `json:"status" gorm:"size:20;default:'active'"`   // active, suspended, inactive
    MaxUsers    int        `json:"max_users" gorm:"default:1000"`            // Max concurrent users
    MaxNas      int        `json:"max_nas" gorm:"default:100"`               // Max NAS devices
    Branding    string     `json:"branding" gorm:"type:text"`                // JSON branding config
    Settings    string     `json:"settings" gorm:"type:text"`               // JSON custom settings
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Provider) TableName() string {
    return "mst_provider"
}
```

### 2.2 Updated Core Models with tenant_id

```go
// All core tables get tenant_id FK

type RadiusUser struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index;not null"`         // NEW
    Username    string     `json:"username" gorm:"uniqueIndex:idx_tenant_username;not null"`
    Password    string     `json:"password" gorm:"not null"`
    Status      string     `json:"status" gorm:"default:'enabled';size:20"`
    // ... existing fields
}

type NetNas struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index;not null"`         // NEW
    Name        string     `json:"name" gorm:"size:255"`
    Identifier   string    `json:"identifier" gorm:"uniqueIndex:idx_tenant_nas_identifier;size:100"`
    Ipaddr      string     `json:"ipaddr" gorm:"size:50"`
    Secret      string     `json:"secret" gorm:"size:100"`
    // ... existing fields
}

type RadiusOnline struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index;not null"`         // NEW
    SessionID   string     `json:"session_id" gorm:"uniqueIndex;size:64"`
    Username    string     `json:"username" gorm:"index"`
    NasAddr     string     `json:"nas_addr" gorm:"size:50"`
    // ... existing fields
}

type Voucher struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index;not null"`         // NEW
    Code        string     `json:"code" gorm:"uniqueIndex;size:100;not null"`
    Status      string     `json:"status" gorm:"size:20;default:'active'"`
    // ... existing fields
}
```

### 2.3 Database Migration

```sql
-- Add tenant_id to all core tables
ALTER TABLE radius_user ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE net_nas ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE radius_online ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE radius_accounting ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE product ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE voucher ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE voucher_batch ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE radius_profile ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE net_node ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;

-- Add composite unique indexes
CREATE UNIQUE INDEX idx_tenant_username ON radius_user(tenant_id, username);
CREATE UNIQUE INDEX idx_tenant_nas_identifier ON net_nas(tenant_id, identifier);
CREATE INDEX idx_tenant_online_username ON radius_online(tenant_id, username);
```

---

## 3. Multi-Tenant Repository Pattern

### 3.1 Tenant Context

```go
// internal/tenant/context.go
package tenant

import (
    "context"
    "errors"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

var ErrNoTenant = errors.New("no tenant context")

func FromContext(ctx context.Context) (int64, error) {
    tenantID, ok := ctx.Value(TenantIDKey).(int64)
    if !ok || tenantID == 0 {
        return 0, ErrNoTenant
    }
    return tenantID, nil
}

func WithTenantID(ctx context.Context, tenantID int64) context.Context {
    return context.WithValue(ctx, TenantIDKey, tenantID)
}

func MustFromContext(ctx context.Context) int64 {
    tenantID, err := FromContext(ctx)
    if err != nil {
        panic(err)
    }
    return tenantID
}
```

### 3.2 Tenant-Aware Repository

```go
// internal/repository/user_repository.go
package repository

type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.RadiusUser, error) {
    tenantID, err := tenant.FromContext(ctx)
    if err != nil {
        return nil, err
    }

    var user domain.RadiusUser
    err = r.db.WithContext(ctx).
        Where("tenant_id = ? AND username = ?", tenantID, username).
        First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.RadiusUser) error {
    tenantID, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    user.TenantID = tenantID
    return r.db.WithContext(ctx).Create(user).Error
}
```

---

## 4. API Design

### 4.1 Provider Management API

```yaml
# Provider CRUD
POST   /api/v1/providers              # Create provider
GET    /api/v1/providers              # List providers (superadmin)
GET    /api/v1/providers/:id          # Get provider details
PUT    /api/v1/providers/:id          # Update provider
DELETE /api/v1/providers/:id          # Delete provider (soft)

# Provider-specific endpoints (uses X-Tenant-ID header)
GET    /api/v1/providers/me           # Get current provider context
PUT    /api/v1/providers/me/settings  # Update provider settings
```

### 4.2 Tenant-Scoped User API

```yaml
# Users - automatically scoped to X-Tenant-ID
GET    /api/v1/users                  # List users
POST   /api/v1/users                  # Create user
GET    /api/v1/users/:id              # Get user
PUT    /api/v1/users/:id              # Update user
DELETE /api/v1/users/:id              # Delete user

# Vouchers - automatically scoped to X-Tenant-ID
GET    /api/v1/vouchers               # List vouchers
POST   /api/v1/vouchers/batches      # Create voucher batch
GET    /api/v1/vouchers/batches/:id  # Get batch
```

### 4.3 Request Headers

```yaml
# Super Admin Operations (platform-wide)
Authorization: Bearer <super_admin_token>

# Provider Admin Operations (tenant-scoped)
Authorization: Bearer <provider_admin_token>
X-Tenant-ID: 123

# RADIUS Requests (auto-detected from NAS)
# tenant_id extracted from NAS device lookup
```

### 4.4 Role-Based Access Control

| Role | Scope | Capabilities |
|------|-------|---------------|
| super | Platform | All providers, system config |
| admin | Provider | All provider resources |
| operator | Provider | User management, vouchers |
| agent | Provider | Vouchers only |

---

## 5. RADIUS Multi-Tenant Processing

### 5.1 Tenant-Aware NAS Lookup

```go
// internal/radiusd/tenant_router.go
package radiusd

type TenantRouter struct {
    cache  *cache.TTLCache
    db     *gorm.DB
}

func (r *TenantRouter) GetTenantForNAS(ctx context.Context, nasIP, identifier string) (int64, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("nas:%s:%s", nasIP, identifier)
    if tenantID, ok := r.cache.Get(cacheKey); ok {
        return tenantID.(int64), nil
    }

    // Lookup NAS in database
    var nas domain.NetNas
    err := r.db.WithContext(ctx).
        Where("ipaddr = ? OR identifier = ?", nasIP, identifier).
        First(&nas).Error
    if err != nil {
        return 0, err
    }

    // Cache result (5 minute TTL)
    r.cache.Set(cacheKey, nas.TenantID, 5*time.Minute)
    return nas.TenantID, nil
}
```

### 5.2 Auth Service with Tenant Context

```go
// internal/radiusd/auth_service.go
func (s *AuthService) ServeRADIUS(ctx context.Context, pkt *radius.Packet) (*radius.Packet, error) {
    // Extract tenant from NAS context
    nasIP := pkt.Source.IP.String()
    tenantID, err := s.tenantRouter.GetTenantForNAS(ctx, nasIP, "")
    if err != nil {
        return s.reject(ErrNasNotFound)
    }

    // Add tenant context for all downstream operations
    ctx = tenant.WithTenantID(ctx, tenantID)

    // Continue with existing auth logic...
    return s.authenticate(ctx, pkt)
}
```

---

## 6. Caching Strategy (Redis)

### 6.1 Cache Key Structure

```
# User cache: tenant-scoped
radio:tenant:{tenant_id}:user:{username} -> User JSON

# NAS cache: tenant-scoped
radio:tenant:{tenant_id}:nas:{ip_or_id} -> NAS JSON

# Session count cache
radio:tenant:{tenant_id}:session_count:{username} -> count

# Voucher cache
radio:tenant:{tenant_id}:voucher:{code} -> Voucher JSON
```

### 6.2 Redis Configuration

```yaml
# docker-compose.yml additions
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data
  command: redis-server --appendonly yes --maxmemory 2gb --maxmemory-policy allkeys-lru

# Cache TTL settings
cache:
  user_ttl: 10s
  nas_ttl: 5m
  session_count_ttl: 2s
  voucher_ttl: 2m
```

---

## 7. Coolify Auto-Deploy Integration

### 7.1 GitHub Actions Workflow

```yaml
# .github/workflows/multi-provider-deploy.yml
name: Build and Deploy Multi-Provider

on:
  push:
    branches:
      - main
      - 'release/**'
    tags:
      - 'v*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      image: ${{ steps.meta.outputs.tags }}
    
    steps:
      - uses: actions/checkout@v4

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=sha

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            BUILD_VERSION=${{ github.sha }}

  deploy-coolify:
    needs: build
    runs-on: ubuntu-latest
    environment: production
    
    steps:
      - name: Trigger Coolify Redeploy
        run: |
          curl -X POST "${{ secrets.COOLIFY_WEBHOOK_URL }}" \
            -H "Content-Type: application/json" \
            -H "X-Coolify-Secret: ${{ secrets.COOLIFY_SECRET }}" \
            -d '{
              "deployment_id": "${{ github.run_id }}",
              "image": "${{ needs.build.outputs.image }}",
              "branch": "${{ github.ref_name }}",
              "commit": "${{ github.sha }}"
            }'

  deploy-staging:
    if: github.ref == 'refs/heads/main'
    needs: build
    runs-on: ubuntu-latest
    environment: staging
    
    steps:
      - name: Trigger Staging Deploy
        run: |
          curl -X POST "${{ secrets.COOLIFY_STAGING_WEBHOOK_URL }}" \
            -H "Content-Type: application/json" \
            -d '{"image": "${{ needs.build.outputs.image }}"}'
```

### 7.2 Coolify Environment Variables

```bash
# Core Configuration
TOUGHRADIUS_SYSTEM_DOMAIN=https://radius.your-platform.com
TOUGHRADIUS_WEB_SECRET=your-random-secret-min-32-chars
TOUGHRADIUS_LOGGER_MODE=production
TOUGHRADIUS_SYSTEM_DEBUG=false

# Multi-Tenant Configuration
TOUGHRADIUS_MULTITENANT_ENABLED=true
TOUGHRADIUS_CACHE_TYPE=redis
TOUGHRADIUS_REDIS_URL=redis://redis:6379/0
TOUGHRADIUS_REDIS_PASSWORD=

# Database Configuration
TOUGHRADIUS_DB_TYPE=postgres
TOUGHRADIUS_DB_HOST=postgres
TOUGHRADIUS_DB_PORT=5432
TOUGHRADIUS_DB_NAME=toughradius
TOUGHRADIUS_DB_USER=toughradius
TOUGHRADIUS_DB_PASSWD=strong-password
TOUGHRADIUS_DB_MAX_CONN=200
TOUGHRADIUS_DB_IDLE_CONN=20

# RADIUS Configuration
TOUGHRADIUS_RADIUSD_ENABLED=true
TOUGHRADIUS_RADIUSD_AUTH_PORT=1812
TOUGHRADIUS_RADIUSD_ACCT_PORT=1813
TOUGHRADIUS_RADIUSD_RADSEC_PORT=2083

# Cache Configuration
TOUGHRADIUS_CACHE_USER_TTL=10s
TOUGHRADIUS_CACHE_NAS_TTL=5m
TOUGHRADIUS_CACHE_SESSION_TTL=2s
```

### 7.3 Coolify Deployment Settings

| Setting | Value |
|---------|-------|
| Build Pack | Dockerfile |
| Dockerfile Path | ./Dockerfile |
| Port Mappings | 1816:1816, 1812:1812/udp, 1813:1813/udp, 2083:2083 |
| Volumes | radio_data:/var/toughradius |
| Health Check | http://localhost:1816/ready |

---

## 8. Scaling Configuration

### 8.1 PostgreSQL Tuning for 100K Users

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
max_parallel_workers = 8

-- Critical indexes for multi-tenant queries
CREATE INDEX CONCURRENTLY idx_users_tenant_status ON radius_user(tenant_id, status);
CREATE INDEX CONCURRENTLY idx_online_tenant_user ON radius_online(tenant_id, username);
CREATE INDEX CONCURRENTLY idx_acct_tenant_time ON radius_accounting(tenant_id, acct_start_time);
CREATE INDEX CONCURRENTLY idx_voucher_tenant_status ON voucher(tenant_id, status);
```

### 8.2 Go Runtime Configuration

```yaml
# toughradius.yml
system:
  worker_dir: /var/toughradius
  appid: RADIO

runtime:
  gomaxprocs: 0  # Auto-detect (0 = numCPU)
  max_open_conns: 200
  max_idle_conns: 20
  conn_max_lifetime: 30m

radiusd:
  worker_pool_size: 100
  request_timeout: 30s
  queue_size: 10000

cache:
  redis_url: redis://redis:6379/0
  local_cache_enabled: true
  local_cache_size: 10000
```

---

## 9. Monitoring & Observability

### 9.1 Key Metrics

```yaml
# Prometheus metrics endpoint
curl localhost:1816/metrics

# Key metrics to monitor:
# - radius_auth_requests_total{tenant_id, result}
# - radius_acct_requests_total{tenant_id, result}
# - active_sessions{tenant_id}
# - cache_hit_ratio{tenant_id, cache_type}
# - db_query_duration{query_type}
# - api_request_duration{endpoint, status}
```

### 9.2 Health Checks

```yaml
# /ready endpoint
GET /ready
Response: {"status": "healthy", "tenants": 100, "sessions": 50000}

# /ready/radius endpoint  
GET /ready/radius
Response: {"auth_port": "up", "acct_port": "up", "radsec_port": "up"}
```

---

## 10. Implementation Phases

### Phase 1: Multi-Tenant Foundation (Week 1-2)
- [ ] Add `tenant_id` to all core tables
- [ ] Create Provider model and migration
- [ ] Implement tenant context middleware
- [ ] Update repositories with tenant scoping
- [ ] Add tenant-aware NAS lookup

### Phase 2: API Multi-Tenancy (Week 3-4)
- [ ] Add `X-Tenant-ID` header support
- [ ] Create provider admin roles
- [ ] Update all API endpoints with tenant scoping
- [ ] Add provider management API
- [ ] Update frontend with provider selector

### Phase 3: Caching & Performance (Week 5-6)
- [ ] Integrate Redis for distributed caching
- [ ] Implement cache invalidation strategy
- [ ] Tune PostgreSQL for multi-tenant workload
- [ ] Add connection pooling per tenant

### Phase 4: CI/CD & Deployment (Week 7-8)
- [ ] Update GitHub Actions workflow
- [ ] Configure Coolify for auto-deploy
- [ ] Add staging environment
- [ ] Implement rollback strategy
- [ ] Documentation update

---

## 11. File Changes Summary

### New Files

| File | Purpose |
|------|---------|
| `internal/domain/tenant.go` | Provider/tenant model |
| `internal/tenant/context.go` | Tenant context utilities |
| `internal/repository/user_repository.go` | Tenant-aware user repo |
| `internal/repository/nas_repository.go` | Tenant-aware NAS repo |
| `internal/radiusd/tenant_router.go` | Tenant routing for RADIUS |
| `internal/middleware/tenant.go` | Tenant middleware |
| `internal/adminapi/providers.go` | Provider CRUD API |
| `internal/cache/redis.go` | Redis cache implementation |
| `internal/config/multitenant.go` | Multi-tenant config |

### Modified Files

| File | Changes |
|------|---------|
| `internal/domain/radius.go` | Add tenant_id to all models |
| `internal/domain/network.go` | Add tenant_id to NAS, Node |
| `internal/domain/voucher.go` | Add tenant_id to vouchers |
| `internal/radiusd/auth_service.go` | Tenant-aware auth |
| `internal/radiusd/acct_service.go` | Tenant-aware accounting |
| `internal/adminapi/adminapi.go` | Register tenant routes |
| `internal/adminapi/users.go` | Tenant-scoped user API |
| `internal/adminapi/vouchers.go` | Tenant-scoped voucher API |
| `config/config.go` | Add multi-tenant config |
| `docker-compose.yml` | Add Redis service |
| `.github/workflows/docker-build.yml` | Add Coolify webhook |
| `Dockerfile` | Multi-stage build optimization |

---

## 12. Estimated Performance

| Component | Current | Target | Change |
|-----------|---------|--------|--------|
| Database | 1 instance | Primary + 2 replicas | HA |
| Cache | In-memory | Redis cluster | Scalability |
| Concurrent Users | ~5,000 | 100,000 | 20x |
| Auth req/s | ~100 | 2,000 | 20x |
| Response Time (p99) | ~50ms | <100ms | Acceptable |
