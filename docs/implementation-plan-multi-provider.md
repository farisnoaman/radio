# Multi-Provider Implementation Plan

> Detailed implementation roadmap for **Phase 1: Multi-Tenant Foundation**

---

## Overview

**Goal**: Transform RADIO from single-tenant to multi-tenant architecture supporting 100 providers × 1,000 concurrent users.

**Timeline**: 8 weeks total
- Phase 1 (Week 1-2): Database & Tenant Foundation
- Phase 2 (Week 3-4): API Multi-Tenancy
- Phase 3 (Week 5-6): Caching & Performance
- Phase 4 (Week 7-8): CI/CD & Coolify Integration

---

## Phase 1: Database & Tenant Foundation (Week 1-2)

### Week 1: Provider Model & Migration

#### Task 1.1: Create Provider Domain Model
**Files to Create:**
```
internal/domain/provider.go
```

**Implementation:**
```go
package domain

import (
    "time"
    "github.com/lib/pq"  // for JSON support
)

type Provider struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    Code        string     `json:"code" gorm:"uniqueIndex;size:50;not null"`
    Name        string     `json:"name" gorm:"size:255;not null"`
    Status      string     `json:"status" gorm:"size:20;default:'active'"`
    MaxUsers    int        `json:"max_users" gorm:"default:1000"`
    MaxNas      int        `json:"max_nas" gorm:"default:100"`
    Branding    string     `json:"branding" gorm:"type:text"`     // JSON
    Settings    string     `json:"settings" gorm:"type:text"`     // JSON
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Provider) TableName() string {
    return "mst_provider"
}
```

**Tests to Write:**
```
internal/domain/provider_test.go
```

---

#### Task 1.2: Create Tenant Context Package
**Files to Create:**
```
internal/tenant/context.go
internal/tenant/context_test.go
```

**Implementation:**
```go
package tenant

import (
    "context"
    "errors"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

var (
    ErrNoTenant      = errors.New("no tenant context")
    ErrInvalidTenant = errors.New("invalid tenant ID")
)

func FromContext(ctx context.Context) (int64, error) {
    tenantID, ok := ctx.Value(TenantIDKey).(int64)
    if !ok || tenantID == 0 {
        return 0, ErrNoTenant
    }
    return tenantID, nil
}

func WithTenantID(ctx context.Context, tenantID int64) context.Context {
    if tenantID <= 0 {
        panic("tenant ID must be positive")
    }
    return context.WithValue(ctx, TenantIDKey, tenantID)
}

func MustFromContext(ctx context.Context) int64 {
    tenantID, err := FromContext(ctx)
    if err != nil {
        panic(err)
    }
    return tenantID
}

func NewContext(parent context.Context, tenantID int64) (context.Context, func()) {
    cancel := func() {}
    if parent.Err() != nil {
        ctx, c := context.WithCancel(parent)
        cancel = c
    }
    return WithTenantID(ctx, tenantID), cancel
}
```

---

#### Task 1.3: Create Tenant Middleware
**Files to Create:**
```
internal/middleware/tenant.go
internal/middleware/tenant_test.go
```

**Implementation:**
```go
package middleware

import (
    "net/http"
    "strconv"
    
    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/tenant"
)

type TenantMiddlewareConfig struct {
    SkipPaths     []string
    DefaultTenant int64
}

func TenantMiddleware(config TenantMiddlewareConfig) echo.MiddlewareFunc {
    skipPaths := make(map[string]bool)
    for _, path := range config.SkipPaths {
        skipPaths[path] = true
    }

    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            path := c.Path()
            
            // Skip tenant check for public paths
            if skipPaths[path] {
                return next(c)
            }

            // Check for X-Tenant-ID header
            tenantHeader := c.Request().Header.Get("X-Tenant-ID")
            if tenantHeader != "" {
                tenantID, err := strconv.ParseInt(tenantHeader, 10, 64)
                if err != nil || tenantID <= 0 {
                    return echo.NewHTTPError(http.StatusBadRequest, "invalid X-Tenant-ID")
                }
                ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
                c.SetRequest(c.Request().WithContext(ctx))
            } else if config.DefaultTenant > 0 {
                // Use default tenant if configured
                ctx := tenant.WithTenantID(c.Request().Context(), config.DefaultTenant)
                c.SetRequest(c.Request().WithContext(ctx))
            }

            return next(c)
        }
    }
}
```

---

#### Task 1.4: Add tenant_id to Existing Models
**Files to Modify:**
```
internal/domain/radius.go
internal/domain/network.go
internal/domain/voucher.go
internal/domain/product.go
```

**Changes:**

```go
// internal/domain/radius.go - Add to RadiusUser
type RadiusUser struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index;not null"`  // NEW
    Username    string     `json:"username" gorm:"uniqueIndex:idx_tenant_username;size:255;not null"`
    Password    string     `json:"password" gorm:"not null"`
    // ... existing fields
}

// Add composite index in TableName() or via GORM callback
```

```go
// internal/domain/network.go - Add to NetNas
type NetNas struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index;not null"`  // NEW
    Name        string     `json:"name" gorm:"size:255"`
    Identifier  string     `json:"identifier" gorm:"uniqueIndex:idx_tenant_nas_identifier;size:100"`
    Ipaddr      string     `json:"ipaddr" gorm:"size:50"`
    Secret      string     `json:"secret" gorm:"size:100"`
    // ... existing fields
}
```

```go
// internal/domain/voucher.go - Add to Voucher
type Voucher struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index;not null"`  // NEW
    Code        string     `json:"code" gorm:"uniqueIndex;size:100;not null"`
    Status      string     `json:"status" gorm:"size:20;default:'active'"`
    // ... existing fields
}
```

---

#### Task 1.5: Create Database Migration
**Files to Create:**
```
migrations/001_add_tenant_support.sql
internal/app/migration.go
```

**SQL Migration:**
```sql
-- migrations/001_add_tenant_support.sql

-- Create provider table
CREATE TABLE IF NOT EXISTS mst_provider (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    max_users INTEGER DEFAULT 1000,
    max_nas INTEGER DEFAULT 100,
    branding TEXT,
    settings TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Insert default provider
INSERT INTO mst_provider (code, name, status) VALUES ('default', 'Default Provider', 'active');

-- Add tenant_id to core tables
ALTER TABLE radius_user ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE net_nas ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE radius_online ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE radius_accounting ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE product ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE voucher ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE voucher_batch ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE radius_profile ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE net_node ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1;
ALTER TABLE sys_opr ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT 1;

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_status ON radius_user(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_online_tenant_user ON radius_online(tenant_id, username);
CREATE INDEX IF NOT EXISTS idx_acct_tenant_time ON radius_accounting(tenant_id, acct_start_time);
CREATE INDEX IF NOT EXISTS idx_voucher_tenant_status ON voucher(tenant_id, status);

-- Drop old unique indexes and create composite
-- For SQLite compatibility, we handle this in the app layer
```

**Go Migration Handler:**
```go
// internal/app/migration.go
func (app *Application) RunMigrations() error {
    // Run tenant migration
    if err := app.migrateTenantSupport(); err != nil {
        return fmt.Errorf("tenant migration failed: %w", err)
    }
    return nil
}

func (app *Application) migrateTenantSupport() error {
    // Check if tenant columns exist
    hasTenant := app.db.Migrator().HasColumn(&domain.RadiusUser{}, "tenant_id")
    if hasTenant {
        return nil // Already migrated
    }

    // Run migration based on DB type
    if app.config.DatabaseType() == "sqlite" {
        return app.migrateTenantSQLite()
    }
    return app.migrateTenantPostgres()
}
```

---

#### Task 1.6: Update Repository Layer
**Files to Create:**
```
internal/repository/user_repository.go
internal/repository/nas_repository.go
internal/repository/voucher_repository.go
```

**User Repository:**
```go
package repository

import (
    "context"
    
    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/tenant"
    "gorm.io/gorm"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
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

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*domain.RadiusUser, int64, error) {
    tenantID, err := tenant.FromContext(ctx)
    if err != nil {
        return nil, 0, err
    }

    var users []*domain.RadiusUser
    var total int64

    query := r.db.WithContext(ctx).Model(&domain.RadiusUser{}).Where("tenant_id = ?", tenantID)
    query.Count(&total)
    
    err = query.Offset(offset).Limit(limit).Find(&users).Error
    return users, total, err
}

func (r *UserRepository) Update(ctx context.Context, user *domain.RadiusUser) error {
    tenantID, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    return r.db.WithContext(ctx).
        Where("tenant_id = ? AND id = ?", tenantID, user.ID).
        Saves(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
    tenantID, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    return r.db.WithContext(ctx).
        Where("tenant_id = ? AND id = ?", tenantID, id).
        Delete(&domain.RadiusUser{}).Error
}
```

**NAS Repository:**
```go
package repository

import (
    "context"
    
    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/tenant"
    "gorm.io/gorm"
)

type NasRepository struct {
    db *gorm.DB
}

func NewNasRepository(db *gorm.DB) *NasRepository {
    return &NasRepository{db: db}
}

func (r *NasRepository) GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error) {
    var nas domain.NetNas
    
    // If we have tenant context, scope the query
    tenantID, err := tenant.FromContext(ctx)
    if err != nil {
        // Fallback: search all tenants (for system operations)
        err = r.db.WithContext(ctx).
            Where("ipaddr = ? OR identifier = ?", ip, identifier).
            First(&nas).Error
    } else {
        err = r.db.WithContext(ctx).
            Where("tenant_id = ? AND (ipaddr = ? OR identifier = ?)", tenantID, ip, identifier).
            First(&nas).Error
    }
    
    if err != nil {
        return nil, err
    }
    return &nas, nil
}

func (r *NasRepository) GetByTenantAndIP(ctx context.Context, tenantID int64, ip, identifier string) (*domain.NetNas, int64, error) {
    var nas domain.NetNas
    
    err := r.db.WithContext(ctx).
        Where("ipaddr = ? OR identifier = ?", ip, identifier).
        First(&nas).Error
    if err != nil {
        return nil, 0, err
    }
    return &nas, nas.TenantID, nil
}
```

---

### Week 2: RADIUS Tenant Integration

#### Task 2.1: Create Tenant Router for RADIUS
**Files to Create:**
```
internal/radiusd/tenant_router.go
internal/radiusd/tenant_router_test.go
```

**Implementation:**
```go
package radiusd

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/pkg/cache"
    "gorm.io/gorm"
)

type TenantRouter struct {
    db    *gorm.DB
    cache *cache.TTLCache
    mu    sync.RWMutex
}

func NewTenantRouter(db *gorm.DB) *TenantRouter {
    return &TenantRouter{
        db:    db,
        cache: cache.NewTTLCache(5 * time.Minute),
    }
}

func (r *TenantRouter) GetTenantForNAS(ctx context.Context, nasIP, identifier string) (int64, error) {
    cacheKey := r.cacheKey(nasIP, identifier)
    
    // Check cache
    r.mu.RLock()
    if tenantID, ok := r.cache.Get(cacheKey); ok {
        r.mu.RUnlock()
        return tenantID.(int64), nil
    }
    r.mu.RUnlock()

    // Lookup in database
    var nas domain.NetNas
    query := r.db.WithContext(ctx).Model(&domain.NetNas{})
    
    if nasIP != "" {
        query = query.Where("ipaddr = ?", nasIP)
    }
    if identifier != "" {
        query = query.Or("identifier = ?", identifier)
    }
    
    err := query.First(&nas).Error
    if err != nil {
        return 0, fmt.Errorf("NAS not found: %w", err)
    }

    // Cache result
    r.mu.Lock()
    r.cache.Set(cacheKey, nas.TenantID)
    r.mu.Unlock()

    return nas.TenantID, nil
}

func (r *TenantRouter) InvalidateCache(nasIP, identifier string) {
    r.mu.Lock()
    r.cache.Delete(r.cacheKey(nasIP, identifier))
    r.mu.Unlock()
}

func (r *TenantRouter) cacheKey(ip, identifier string) string {
    return fmt.Sprintf("nas:%s:%s", ip, identifier)
}
```

---

#### Task 2.2: Update Auth Service for Tenant Context
**Files to Modify:**
```
internal/radiusd/auth_service.go
internal/radiusd/radius_auth.go
```

**Changes:**
```go
// Add tenant context to auth request
func (s *AuthService) ServeRADIUS(ctx context.Context, pkt *radius.Packet) (*radius.Packet, error) {
    // Extract tenant from NAS context
    nasIP := pkt.Source.IP.String()
    username := extractUsername(pkt)
    
    tenantID, err := s.tenantRouter.GetTenantForNAS(ctx, nasIP, "")
    if err != nil {
        log.Printf("NAS not found for IP: %s", nasIP)
        return s.reject(ErrNasNotFound)
    }

    // Add tenant context for all downstream operations
    ctx = tenant.WithTenantID(ctx, tenantID)

    // Get user with tenant scope
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return s.reject(ErrUserNotFound)
        }
        return nil, err
    }

    // Continue with authentication...
    return s.authenticateUser(ctx, user, pkt)
}
```

---

#### Task 2.3: Update Acct Service for Tenant Context
**Files to Modify:**
```
internal/radiusd/acct_service.go
```

**Changes:**
```go
func (s *AcctService) ServeRADIUS(ctx context.Context, pkt *radius.Packet) error {
    nasIP := pkt.Source.IP.String()
    
    tenantID, err := s.tenantRouter.GetTenantForNAS(ctx, nasIP, "")
    if err != nil {
        return fmt.Errorf("NAS not found: %w", err)
    }

    ctx = tenant.WithTenantID(ctx, tenantID)
    
    // Process accounting with tenant context
    return s.processAccounting(ctx, pkt)
}
```

---

#### Task 2.4: Create Tenant-Scoped Cache
**Files to Create:**
```
internal/cache/tenant_cache.go
internal/cache/tenant_cache_test.go
```

**Implementation:**
```go
package cache

import (
    "fmt"
    "sync"
    "time"
)

type TenantCache struct {
    defaultTTL time.Duration
    mu         sync.RWMutex
    caches     map[int64]*TTLCache
}

func NewTenantCache(defaultTTL time.Duration) *TenantCache {
    return &TenantCache{
        defaultTTL: defaultTTL,
        caches:     make(map[int64]*TTLCache),
    }
}

func (c *TenantCache) GetCache(tenantID int64) *TTLCache {
    c.mu.RLock()
    cache, ok := c.caches[tenantID]
    c.mu.RUnlock()
    
    if ok {
        return cache
    }

    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Double-check after acquiring write lock
    if cache, ok := c.caches[tenantID]; ok {
        return cache
    }

    cache = NewTTLCache(c.defaultTTL)
    c.caches[tenantID] = cache
    return cache
}

func (c *TenantCache) UserCacheKey(tenantID int64, username string) string {
    return fmt.Sprintf("tenant:%d:user:%s", tenantID, username)
}

func (c *TenantCache) NasCacheKey(tenantID int64, nasIP string) string {
    return fmt.Sprintf("tenant:%d:nas:%s", tenantID, nasIP)
}

func (c *TenantCache) SessionCacheKey(tenantID int64, username string) string {
    return fmt.Sprintf("tenant:%d:session:%s", tenantID, username)
}

func (c *TenantCache) Get(tenantID int64, key string) (interface{}, bool) {
    return c.GetCache(tenantID).Get(key)
}

func (c *TenantCache) Set(tenantID int64, key string, value interface{}) {
    c.GetCache(tenantID).Set(key, value)
}

func (c *TenantCache) Delete(tenantID int64, key string) {
    c.GetCache(tenantID).Delete(key)
}

func (c *TenantCache) Clear(tenantID int64) {
    c.mu.Lock()
    delete(c.caches, tenantID)
    c.mu.Unlock()
}
```

---

## Phase 2: API Multi-Tenancy (Week 3-4)

### Week 3: Provider Management API

#### Task 3.1: Create Provider CRUD API
**Files to Create:**
```
internal/adminapi/providers.go
internal/adminapi/providers_test.go
```

**Endpoints:**
```go
func registerProviderRoutes() {
    // Super admin only
    webserver.ApiGET("/providers", ListProviders)
    webserver.ApiPOST("/providers", CreateProvider)
    webserver.ApiGET("/providers/:id", GetProvider)
    webserver.ApiPUT("/providers/:id", UpdateProvider)
    webserver.ApiDELETE("/providers/:id", DeleteProvider)
    
    // Provider admin
    webserver.ApiGET("/providers/me", GetCurrentProvider)
    webserver.ApiPUT("/providers/me", UpdateCurrentProvider)
}

type ProviderRequest struct {
    Code     string `json:"code" validate:"required,max=50"`
    Name     string `json:"name" validate:"required,max=255"`
    MaxUsers int    `json:"max_users" validate:"min=0,max=100000"`
    MaxNas   int    `json:"max_nas" validate:"min=0,max=10000"`
    Branding string `json:"branding"`
    Settings string `json:"settings"`
}

func CreateProvider(c echo.Context) error {
    var req ProviderRequest
    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
    }

    provider := &domain.Provider{
        Code:     req.Code,
        Name:     req.Name,
        MaxUsers: req.MaxUsers,
        MaxNas:   req.MaxNas,
        Branding: req.Branding,
        Settings: req.Settings,
        Status:   "active",
    }

    if err := GetDB(c).Create(provider).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create provider", nil)
    }

    return ok(c, provider)
}
```

---

#### Task 3.2: Update User API with Tenant Scope
**Files to Modify:**
```
internal/adminapi/users.go
```

**Changes:**
```go
func ListUsers(c echo.Context) error {
    // Get tenant from context
    tenantID, err := tenant.FromContext(c.Request().Context())
    if err != nil {
        return fail(c, http.StatusBadRequest, "NO_TENANT", "Missing tenant context", nil)
    }

    page, _ := strconv.Atoi(c.QueryParam("page"))
    perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
    if page < 1 { page = 1 }
    if perPage < 1 || perPage > 100 { perPage = 20 }

    var users []*domain.RadiusUser
    var total int64

    db := GetDB(c).Model(&domain.RadiusUser{}).Where("tenant_id = ?", tenantID)
    
    // Apply filters
    if status := c.QueryParam("status"); status != "" {
        db = db.Where("status = ?", status)
    }
    if search := c.QueryParam("search"); search != "" {
        db = db.Where("username LIKE ?", "%"+search+"%")
    }

    db.Count(&total)
    offset := (page - 1) * perPage
    db.Offset(offset).Limit(perPage).Order("id DESC").Find(&users)

    return paged(c, users, total, page, perPage)
}

func CreateUser(c echo.Context) error {
    tenantID, err := tenant.FromContext(c.Request().Context())
    if err != nil {
        return fail(c, http.StatusBadRequest, "NO_TENANT", "Missing tenant context", nil)
    }

    var req UserRequest
    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", nil)
    }

    user := &domain.RadiusUser{
        TenantID: tenantID,  // Set tenant from context
        Username: req.Username,
        Password: req.Password,
        // ... other fields
    }

    if err := GetDB(c).Create(user).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create user", nil)
    }

    return ok(c, user)
}
```

---

#### Task 3.3: Update Voucher API with Tenant Scope
**Files to Modify:**
```
internal/adminapi/vouchers.go
```

**Changes:** Similar to user API, add tenant scope to all queries

---

#### Task 3.4: Create Operator with Tenant Role
**Files to Modify:**
```
internal/domain/system.go
```

**Changes:**
```go
type SysOpr struct {
    ID          int64      `json:"id" gorm:"primaryKey"`
    TenantID    int64      `json:"tenant_id" gorm:"index"`  // NEW: for provider admins
    Username    string     `json:"username" gorm:"uniqueIndex;size:50;not null"`
    Password    string     `json:"-" gorm:"size:100"`
    Realname    string     `json:"realname" gorm:"size:100"`
    Role        string     `json:"role" gorm:"size:20"`  // super, admin, operator, agent
    TenantID    int64      `json:"tenant_id" gorm:"index"`  // NEW
    // ...
}

const (
    RoleSuper    = "super"    // Platform-wide admin
    RoleAdmin    = "admin"    // Provider admin
    RoleOperator = "operator" // Provider operator
    RoleAgent    = "agent"    // Provider agent
)
```

---

### Week 4: Frontend Tenant Support

#### Task 4.1: Add Provider Selector to Frontend
**Files to Modify:**
```
web/src/App.tsx
web/src/providers/authProvider.ts
web/src/providers/dataProvider.ts
```

**Changes:**
```tsx
// Add provider context to API requests
const dataProvider = {
  ...defaultDataProvider,
  getList: (resource, params) => {
    const tenantId = getCurrentTenantId();
    return defaultDataProvider.getList(resource, {
      ...params,
      meta: { ...params.meta, tenant_id: tenantId }
    });
  },
  // Apply tenant_id header to all requests
};

// Add tenant selector dropdown
const ProviderSelector = () => {
  const { data: providers } = useListControllersContext();
  const [selectedProvider, setSelectedProvider] = useState(null);
  
  return (
    <Select
      value={selectedProvider}
      onChange={setSelectedProvider}
      label="Provider"
    >
      {providers?.map(p => (
        <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>
      ))}
    </Select>
  );
};
```

---

## Phase 3: Caching & Performance (Week 5-6)

### Week 5: Redis Integration

#### Task 5.1: Add Redis Configuration
**Files to Create:**
```
internal/cache/redis.go
```

**Implementation:**
```go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
    prefix string
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
        PoolSize: 100,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis connection failed: %w", err)
    }

    return &RedisCache{
        client: client,
        prefix: "radio:",
    }, nil
}

func (c *RedisCache) UserKey(tenantID int64, username string) string {
    return fmt.Sprintf("%stenant:%d:user:%s", c.prefix, tenantID, username)
}

func (c *RedisCache) GetUser(ctx context.Context, tenantID int64, username string) (*CachedUser, error) {
    key := c.UserKey(tenantID, username)
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }

    var user CachedUser
    if err := json.Unmarshal(data, &user); err != nil {
        return nil, err
    }
    return &user, nil
}

func (c *RedisCache) SetUser(ctx context.Context, tenantID int64, username string, user *CachedUser, ttl time.Duration) error {
    key := c.UserKey(tenantID, username)
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) DeleteUser(ctx context.Context, tenantID int64, username string) error {
    key := c.UserKey(tenantID, username)
    return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) IncrementSessionCount(ctx context.Context, tenantID int64, username string) (int64, error) {
    key := fmt.Sprintf("%stenant:%d:session:%s", c.prefix, tenantID, username)
    return c.client.Incr(ctx, key).Result()
}
```

---

#### Task 5.2: Update Config for Redis
**Files to Modify:**
```
config/config.go
```

**Changes:**
```go
type CacheConfig struct {
    Type            string        `yaml:"cache_type"`  // memory, redis
    RedisURL        string        `yaml:"redis_url"`
    RedisPassword   string        `yaml:"redis_password"`
    RedisDB         int           `yaml:"redis_db"`
    UserTTL         time.Duration `yaml:"user_ttl"`
    NasTTL          time.Duration `yaml:"nas_ttl"`
    SessionCountTTL time.Duration `yaml:"session_count_ttl"`
}

func (c *Config) LoadCacheConfig() *CacheConfig {
    return &CacheConfig{
        Type:            getEnv("TOUGHRADIUS_CACHE_TYPE", "memory"),
        RedisURL:        getEnv("TOUGHRADIUS_REDIS_URL", "localhost:6379"),
        RedisPassword:   getEnv("TOUGHRADIUS_REDIS_PASSWORD", ""),
        RedisDB:         getEnvInt("TOUGHRADIUS_REDIS_DB", 0),
        UserTTL:         getEnvDuration("TOUGHRADIUS_CACHE_USER_TTL", 10*time.Second),
        NasTTL:          getEnvDuration("TOUGHRADIUS_CACHE_NAS_TTL", 5*time.Minute),
        SessionCountTTL: getEnvDuration("TOUGHRADIUS_CACHE_SESSION_TTL", 2*time.Second),
    }
}
```

---

### Week 6: Performance Optimization

#### Task 6.1: Update docker-compose.yml
**Files to Modify:**
```
docker-compose.yml
```

**Changes:**
```yaml
version: '3.8'
services:
  toughradius:
    image: farisnoaman/toughradius:latest
    ports:
      - "1816:1816"
      - "1812:1812"
      - "1813:1813"
    environment:
      - TOUGHRADIUS_DB_TYPE=postgres
      - TOUGHRADIUS_DB_HOST=postgres
      - TOUGHRADIUS_DB_PORT=5432
      - TOUGHRADIUS_DB_NAME=toughradius
      - TOUGHRADIUS_DB_USER=toughradius
      - TOUGHRADIUS_DB_PASSWD=${DB_PASSWORD}
      - TOUGHRADIUS_DB_MAX_CONN=200
      - TOUGHRADIUS_DB_IDLE_CONN=20
      - TOUGHRADIUS_CACHE_TYPE=redis
      - TOUGHRADIUS_REDIS_URL=redis://redis:6379/0
    depends_on:
      - postgres
      - redis
    volumes:
      - toughradius_data:/var/toughradius
    restart: unless-stopped

  postgres:
    image: postgres:16
    environment:
      - POSTGRES_DB=toughradius
      - POSTGRES_USER=toughradius
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgres.conf:/etc/postgresql/postgresql.conf
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes --maxmemory 2gb --maxmemory-policy allkeys-lru
    restart: unless-stopped

volumes:
  toughradius_data:
  postgres_data:
  redis_data:
```

---

#### Task 6.2: PostgreSQL Tuning
**Files to Create:**
```
postgres.conf
```

**Settings:**
```conf
# Connection settings
max_connections = 200
superuser_reserved_connections = 5

# Memory settings
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 512MB
work_mem = 16MB

# Write settings
wal_buffers = 64MB
checkpoint_completion_target = 0.9
max_wal_size = 4GB

# Parallel queries
max_worker_processes = 8
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
max_parallel_maintenance_workers = 4

# Logging
log_min_duration_statement = 1000
log_connections = on
log_disconnections = on

# Performance
random_page_cost = 1.1
effective_io_concurrency = 200
```

---

## Phase 4: CI/CD & Coolify Integration (Week 7-8)

### Week 7: GitHub Actions Workflow

#### Task 7.1: Update GitHub Actions Workflow
**Files to Modify:**
```
.github/workflows/docker-build.yml
```

**Changes:**
```yaml
name: Build and Deploy Multi-Provider RADIO

on:
  push:
    branches:
      - main
      - 'release/**'
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        default: 'staging'
        type: choice
        options:
          - staging
          - production

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}
  MULTITENANT_ENABLED: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
        
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    outputs:
      image: ${{ steps.meta.outputs.tags }}
      version: ${{ steps.version.outputs.version }}
    
    steps:
      - uses: actions/checkout@v4

      - name: Extract version
        id: version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          if [ "$VERSION" = "$GITHUB_REF" ]; then
            if [ "${{ github.event_name }}" = "workflow_dispatch" ]; then
              VERSION="${{ inputs.environment }}-$(date +%Y%m%d%H%M%S)"
            else
              VERSION=$(date +%Y%m%d)-$(echo ${{ github.sha }} | cut -c1-7)
            fi
          fi
          echo "version=${VERSION}" >> $GITHUB_OUTPUT
          echo "image=${REGISTRY}/${IMAGE_NAME}:${VERSION}" >> $GITHUB_OUTPUT

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

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=sha
            type=raw,value=${{ steps.version.outputs.version }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            BUILD_VERSION=${{ github.sha }}
            MULTITENANT_ENABLED=true

  deploy-staging:
    if: github.ref == 'refs/heads/main' || github.event.inputs.environment == 'staging'
    needs: build
    runs-on: ubuntu-latest
    environment: staging
    
    steps:
      - name: Trigger Coolify Staging Deploy
        run: |
          curl -X POST "${{ secrets.COOLIFY_STAGING_WEBHOOK_URL }}" \
            -H "Content-Type: application/json" \
            -H "X-Coolify-Secret: ${{ secrets.COOLIFY_SECRET }}" \
            -d '{
              "deployment_id": "${{ github.run_id }}",
              "image": "${{ needs.build.outputs.image }}",
              "version": "${{ needs.build.outputs.version }}",
              "branch": "${{ github.ref_name }}",
              "commit": "${{ github.sha }}",
              "commit_message": "${{ github.event.head_commit.message }}",
              "triggered_by": "github_actions"
            }'

      - name: Wait for deployment
        run: |
          echo "Waiting for Coolify to pull and deploy new image..."
          sleep 60
          
      - name: Verify deployment
        run: |
          echo "Checking deployment status..."
          curl -f "${{ secrets.STAGING_URL }}/ready" || exit 1

  deploy-production:
    if: startsWith(github.ref, 'refs/tags/v') || github.event.inputs.environment == 'production'
    needs: build
    runs-on: ubuntu-latest
    environment: production
    
    steps:
      - name: Trigger Coolify Production Deploy
        run: |
          curl -X POST "${{ secrets.COOLIFY_PRODUCTION_WEBHOOK_URL }}" \
            -H "Content-Type: application/json" \
            -H "X-Coolify-Secret: ${{ secrets.COOLIFY_SECRET }}" \
            -d '{
              "deployment_id": "${{ github.run_id }}",
              "image": "${{ needs.build.outputs.image }}",
              "version": "${{ needs.build.outputs.version }}",
              "tag": "${{ github.ref_name }}",
              "commit": "${{ github.sha }}",
              "triggered_by": "github_actions"
            }'

      - name: Notify success
        if: success()
        run: |
          echo "Deployment triggered successfully"
          
      - name: Notify failure
        if: failure()
        run: |
          echo "Deployment failed. Check Coolify dashboard."
```

---

### Week 8: Coolify Configuration & Documentation

#### Task 8.1: Create Coolify Deployment Guide
**Files to Create:**
```
docs/guides/multi-provider-coolify-deployment.md
```

**Content:**
```markdown
# Multi-Provider RADIO Deployment on Coolify

## Prerequisites

1. Coolify instance with:
   - Docker 24+
   - PostgreSQL 16+
   - Redis 7+
   - At least 4GB RAM, 4 vCPUs

2. GitHub repository with RADIO source

## Deployment Steps

### 1. Create PostgreSQL Database

```bash
# SSH into Coolify server
ssh coolify@your-server

# Create database
docker exec -it postgres psql -U postgres -c "CREATE DATABASE toughradius;"
docker exec -it postgres psql -U postgres -c "CREATE USER toughradius WITH PASSWORD 'strong-password';"
docker exec -it postgres psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE toughradius TO toughradius;"
```

### 2. Configure Redis

Redis is automatically configured via docker-compose in Coolify.

### 3. Create Coolify Application

1. Go to Coolify Dashboard
2. Click "New Resource" → "Application"
3. Connect GitHub repository
4. Select branch (main for staging, release/* for production)
5. Configure build:
   - Build Pack: Dockerfile
   - Dockerfile Path: ./Dockerfile
6. Configure environment variables:

```env
# Database
TOUGHRADIUS_DB_TYPE=postgres
TOUGHRADIUS_DB_HOST=postgres
TOUGHRADIUS_DB_PORT=5432
TOUGHRADIUS_DB_NAME=toughradius
TOUGHRADIUS_DB_USER=toughradius
TOUGHRADIUS_DB_PASSWD=strong-password
TOUGHRADIUS_DB_MAX_CONN=200
TOUGHRADIUS_DB_IDLE_CONN=20

# Cache
TOUGHRADIUS_CACHE_TYPE=redis
TOUGHRADIUS_REDIS_URL=redis://redis:6379/0

# Multi-Tenant
TOUGHRADIUS_MULTITENANT_ENABLED=true

# System
TOUGHRADIUS_SYSTEM_DOMAIN=https://your-domain.com
TOUGHRADIUS_WEB_SECRET=your-32-char-secret
TOUGHRADIUS_LOGGER_MODE=production
TOUGHRADIUS_SYSTEM_DEBUG=false
```

### 4. Configure Ports

| Internal Port | External Port | Protocol |
|--------------|---------------|----------|
| 1816 | 1816 | TCP |
| 1812 | 1812 | UDP |
| 1813 | 1813 | UDP |
| 2083 | 2083 | TCP |

### 5. Configure Volumes

| Source | Destination |
|--------|-------------|
| radio_data | /var/toughradius |

### 6. Set Up Health Check

```
URL: http://localhost:1816/ready
Interval: 30s
Timeout: 10s
Retries: 3
```

### 7. Configure Webhook

Add GitHub webhook for auto-deploy:

1. In Coolify, go to Application → Deployments → Webhooks
2. Copy the webhook URL
3. In GitHub, go to Settings → Webhooks → Add webhook
4. Payload URL: `<coolify-webhook-url>`
5. Content type: application/json
6. Events: Push, Tag

## Auto-Deploy Configuration

### GitHub Secrets

Add these to GitHub repository secrets:

| Secret | Description |
|--------|-------------|
| COOLIFY_SECRET | Secret for webhook authentication |
| COOLIFY_STAGING_WEBHOOK_URL | Staging deploy webhook URL |
| COOLIFY_PRODUCTION_WEBHOOK_URL | Production deploy webhook URL |

### Triggering Deploys

**Automatic:**
- Push to main branch → Deploys to staging
- Push to release/* branch → Deploys to staging
- New tag v* → Deploys to production

**Manual:**
- Go to GitHub Actions → "Build and Deploy"
- Click "Run workflow"
- Select environment

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

### Metrics Endpoint

```bash
curl https://your-domain.com/metrics
```

Prometheus-format metrics including:
- radius_auth_requests_total
- radius_acct_requests_total
- active_sessions
- cache_hit_ratio
```

---

#### Task 8.2: Update Environment Variables Documentation
**Files to Modify:**
```
docs/environment-variables.md
```

**Add Multi-Tenant Variables:**
```markdown
## Multi-Tenant Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| TOUGHRADIUS_MULTITENANT_ENABLED | false | Enable multi-tenant mode |
| TOUGHRADIUS_CACHE_TYPE | memory | Cache type: memory, redis |
| TOUGHRADIUS_REDIS_URL | localhost:6379 | Redis server URL |
| TOUGHRADIUS_REDIS_PASSWORD | - | Redis password |
| TOUGHRADIUS_REDIS_DB | 0 | Redis database number |
| TOUGHRADIUS_CACHE_USER_TTL | 10s | User cache TTL |
| TOUGHRADIUS_CACHE_NAS_TTL | 5m | NAS cache TTL |
| TOUGHRADIUS_CACHE_SESSION_TTL | 2s | Session count cache TTL |
```

---

## Testing Strategy

### Unit Tests
```bash
# Run all unit tests
go test ./... -v -race

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests
```bash
# Run integration tests
go test ./internal/... -tags=integration -v

# Test multi-tenant scenarios
go test ./internal/tenant/... -v
go test ./internal/radiusd/... -run Tenant -v
```

### Load Testing
```bash
# Install wrk
brew install wrk

# Test RADIUS auth throughput
wrk -t10 -c100 -d30s http://localhost:1816/metrics

# Test with specific tenant
curl -H "X-Tenant-ID: 1" http://localhost:1816/api/v1/users
```

---

## Rollback Strategy

### Docker Image Rollback
```bash
# List available tags
docker images farisnoaman/toughradius

# Pull specific version
docker pull farisnoaman/toughradius:v1.0.0

# Update Coolify to use specific image
# Go to Coolify → Application → Environment → Change image tag
```

### Database Rollback
```bash
# Restore from backup
docker exec -it toughradius-app-1 toughradius -restore /var/toughradius/backup/backup_20240101.sql
```

---

## Success Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Concurrent users | 100,000 | - |
| Auth latency (p99) | <100ms | - |
| Acct latency (p99) | <50ms | - |
| Cache hit rate | >95% | - |
| API latency (p99) | <200ms | - |
| Uptime | 99.9% | - |
