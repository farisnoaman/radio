# Phase 1: Multi-Tenant Middleware Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement tenant-aware middleware that enforces data isolation across all API layers.

**Architecture:** Middleware intercepts all HTTP requests, extracts tenant context from authenticated user, and injects tenant_id into database queries. Multi-layer enforcement ensures no cross-tenant data leakage.

**Tech Stack:** Echo framework, Go context package, GORM scopes

---

## Task 1: Enhance Tenant Context System

**Files:**
- Modify: `internal/tenant/context.go` (enhance existing)
- Create: `internal/tenant/context_test.go`

**Step 1: Write tests for enhanced tenant context**

```go
// internal/tenant/context_test.go
package tenant

import (
    "context"
    "testing"
)

func TestFromContext(t *testing.T) {
    ctx := context.Background()

    // Test missing tenant
    _, err := FromContext(ctx)
    if err != ErrNoTenant {
        t.Errorf("Expected ErrNoTenant, got %v", err)
    }

    // Test valid tenant
    ctx = WithTenantID(ctx, 42)
    tenantID, err := FromContext(ctx)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if tenantID != 42 {
        t.Errorf("Expected tenant ID 42, got %d", tenantID)
    }
}

func TestMustFromContext(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Error("Expected panic when no tenant in context")
        }
    }()

    ctx := context.Background()
    MustFromContext(ctx)
}

func TestGetTenantIDOrDefault(t *testing.T) {
    ctx := context.Background()

    // Test default
    tenantID := GetTenantIDOrDefault(ctx)
    if tenantID != DefaultTenantID {
        t.Errorf("Expected default tenant ID %d, got %d", DefaultTenantID, tenantID)
    }

    // Test with tenant in context
    ctx = WithTenantID(ctx, 99)
    tenantID = GetTenantIDOrDefault(ctx)
    if tenantID != 99 {
        t.Errorf("Expected tenant ID 99, got %d", tenantID)
    }
}

func TestTenantContext(t *testing.T) {
    ctx := context.Background()

    tc, err := NewTenantContext(ctx, 1)
    if err != nil {
        t.Fatalf("Failed to create tenant context: %v", err)
    }

    if tc.TenantID != 1 {
        t.Errorf("Expected tenant ID 1, got %d", tc.TenantID)
    }

    extracted := MustFromContext(tc.Extract())
    if extracted != 1 {
        t.Errorf("Expected extracted tenant ID 1, got %d", extracted)
    }
}

func TestValidateTenantID(t *testing.T) {
    tests := []struct {
        name    string
        tenantID int64
        wantErr error
    }{
        {"valid", 1, nil},
        {"valid large", 999999, nil},
        {"invalid zero", 0, ErrInvalidTenant},
        {"invalid negative", -1, ErrInvalidTenant},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateTenantID(tt.tenantID)
            if err != tt.wantErr {
                t.Errorf("ValidateTenantID(%d) = %v, want %v", tt.tenantID, err, tt.wantErr)
            }
        })
    }
}

func TestTenantChecker(t *testing.T) {
    checker := NewTenantChecker()

    // Test IsSystemTenant
    if !checker.IsSystemTenant(DefaultTenantID) {
        t.Error("Expected default tenant to be system tenant")
    }
    if checker.IsSystemTenant(99) {
        t.Error("Expected tenant 99 not to be system tenant")
    }

    // Test CanAccess
    if !checker.CanAccess(DefaultTenantID, 1) {
        t.Error("System tenant should access any tenant")
    }
    if !checker.CanAccess(1, 1) {
        t.Error("Tenant should access own resources")
    }
    if checker.CanAccess(1, 2) {
        t.Error("Tenant should not access other tenant resources")
    }
}
```

**Step 2: Run tests to verify they pass (existing implementation should work)**

Run: `go test ./internal/tenant -v`
Expected: PASS (most tests should pass with existing code)

**Step 3: Enhance tenant context with additional methods**

```go
// internal/tenant/context.go
package tenant

import (
    "context"
    "errors"
)

type contextKey string

const (
    TenantIDKey contextKey = "tenant_id"
    DefaultTenantID int64 = 1
)

var (
    ErrNoTenant       = errors.New("no tenant context found in request")
    ErrInvalidTenant  = errors.New("invalid tenant ID")
    ErrTenantMismatch = errors.New("tenant ID mismatch")
)

// FromContext extracts the tenant ID from a context.
// Returns ErrNoTenant if no tenant ID is present.
func FromContext(ctx context.Context) (int64, error) {
    tenantID, ok := ctx.Value(TenantIDKey).(int64)
    if !ok || tenantID <= 0 {
        return 0, ErrNoTenant
    }
    return tenantID, nil
}

// WithTenantID returns a new context with the specified tenant ID.
// Panics if tenantID is not positive.
func WithTenantID(ctx context.Context, tenantID int64) context.Context {
    if tenantID <= 0 {
        panic("tenant ID must be positive")
    }
    return context.WithValue(ctx, TenantIDKey, tenantID)
}

// MustFromContext extracts the tenant ID from context.
// Panics if no tenant ID is present.
func MustFromContext(ctx context.Context) int64 {
    tenantID, err := FromContext(ctx)
    if err != nil {
        panic(err)
    }
    return tenantID
}

// GetTenantIDOrDefault returns the tenant ID from context or the default.
func GetTenantIDOrDefault(ctx context.Context) int64 {
    tenantID, err := FromContext(ctx)
    if err != nil {
        return DefaultTenantID
    }
    return tenantID
}

// ValidateTenantID checks if the tenant ID is valid (positive).
func ValidateTenantID(tenantID int64) error {
    if tenantID <= 0 {
        return ErrInvalidTenant
    }
    return nil
}

// TenantContext wraps a context with tenant information.
type TenantContext struct {
    TenantID int64
    Context  context.Context
}

// NewTenantContext creates a new TenantContext.
func NewTenantContext(ctx context.Context, tenantID int64) (*TenantContext, error) {
    if err := ValidateTenantID(tenantID); err != nil {
        return nil, err
    }
    return &TenantContext{
        TenantID: tenantID,
        Context:  WithTenantID(ctx, tenantID),
    }, nil
}

// Extract extracts the tenant context from the provided context.
func (tc *TenantContext) Extract() context.Context {
    return tc.Context
}

// TenantChecker provides methods to check tenant-related conditions.
type TenantChecker struct{}

// NewTenantChecker creates a new TenantChecker.
func NewTenantChecker() *TenantChecker {
    return &TenantChecker{}
}

// IsSystemTenant checks if the tenant ID represents the system tenant.
func (c *TenantChecker) IsSystemTenant(tenantID int64) bool {
    return tenantID == DefaultTenantID
}

// CanAccess checks if a request from sourceTenant can access targetTenant resources.
func (c *TenantChecker) CanAccess(sourceTenantID, targetTenantID int64) bool {
    if c.IsSystemTenant(sourceTenantID) {
        return true
    }
    return sourceTenantID == targetTenantID
}

// TenantScope returns a GORM scope for tenant isolation.
func TenantScope(tenantID int64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}
```

**Step 4: Run tests to verify enhancements work**

Run: `go test ./internal/tenant -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tenant/context.go internal/tenant/context_test.go
git commit -m "feat(tenant): add TenantScope helper for GORM queries"
```

---

## Task 2: Create Tenant Isolation Middleware

**Files:**
- Modify: `internal/middleware/tenant.go` (enhance existing)
- Create: `internal/middleware/tenant_test.go`

**Step 1: Write tests for tenant middleware**

```go
// internal/middleware/tenant_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/tenant"
)

func TestTenantMiddleware(t *testing.T) {
    e := echo.New()

    // Test with X-Tenant-ID header
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("X-Tenant-ID", "42")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    middleware := TenantMiddleware(TenantMiddlewareConfig{
        DefaultTenant: 1,
    })

    handler := middleware(func(c echo.Context) error {
        tenantID, err := tenant.FromContext(c.Request().Context())
        if err != nil {
            return err
        }
        return c.String(200, string(rune(tenantID)))
    })

    err := handler(c)
    if err != nil {
        t.Fatalf("Handler failed: %v", err)
    }

    tenantID, _ := tenant.FromContext(c.Request().Context())
    if tenantID != 42 {
        t.Errorf("Expected tenant ID 42, got %d", tenantID)
    }
}

func TestTenantMiddlewareInvalidHeader(t *testing.T) {
    e := echo.New()

    // Test with invalid X-Tenant-ID header
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("X-Tenant-ID", "invalid")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    middleware := TenantMiddleware(TenantMiddlewareConfig{})

    handler := middleware(func(c echo.Context) error {
        return c.String(200, "ok")
    })

    err := handler(c)
    if err == nil {
        t.Error("Expected error for invalid tenant ID")
    }
}

func TestTenantMiddlewareDefaultTenant(t *testing.T) {
    e := echo.New()

    // Test without X-Tenant-ID header (should use default)
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    middleware := TenantMiddleware(TenantMiddlewareConfig{
        DefaultTenant: 1,
    })

    handler := middleware(func(c echo.Context) error {
        tenantID, err := tenant.FromContext(c.Request().Context())
        if err != nil {
            return err
        }
        return c.String(200, string(rune(tenantID)))
    })

    err := handler(c)
    if err != nil {
        t.Fatalf("Handler failed: %v", err)
    }

    tenantID, _ := tenant.FromContext(c.Request().Context())
    if tenantID != 1 {
        t.Errorf("Expected default tenant ID 1, got %d", tenantID)
    }
}

func TestTenantMiddlewareSkipPaths(t *testing.T) {
    e := echo.New()

    // Test skip paths
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetPath("/health")

    middleware := TenantMiddleware(TenantMiddlewareConfig{
        SkipPaths:     []string{"/health", "/metrics"},
        DefaultTenant: 1,
    })

    handler := middleware(func(c echo.Context) error {
        // Should not have tenant context
        _, err := tenant.FromContext(c.Request().Context())
        if err == nil {
            return c.String(200, "has tenant")
        }
        return c.String(200, "no tenant")
    })

    err := handler(c)
    if err != nil {
        t.Fatalf("Handler failed: %v", err)
    }

    if rec.Body.String() != "no tenant" {
        t.Error("Expected skipped path to not have tenant context")
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/middleware -run TestTenant -v`
Expected: FAIL (tests fail as middleware needs implementation)

**Step 3: Implement tenant middleware**

```go
// internal/middleware/tenant.go
package middleware

import (
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/tenant"
)

const (
    TenantIDHeader = "X-Tenant-ID"
)

// TenantMiddlewareConfig configures the tenant middleware.
type TenantMiddlewareConfig struct {
    SkipPaths     []string
    DefaultTenant int64
}

// TenantMiddleware extracts tenant ID from request headers and adds it to context.
// It enforces that all requests have a valid tenant context for data isolation.
func TenantMiddleware(config TenantMiddlewareConfig) echo.MiddlewareFunc {
    skipPaths := make(map[string]bool)
    for _, path := range config.SkipPaths {
        skipPaths[path] = true
    }

    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            path := c.Path()

            // Skip tenant check for specified paths
            if skipPaths[path] || skipPaths[c.Request().URL.Path] {
                return next(c)
            }

            // Extract tenant ID from header
            tenantHeader := c.Request().Header.Get(TenantIDHeader)
            if tenantHeader != "" {
                tenantID, err := strconv.ParseInt(tenantHeader, 10, 64)
                if err != nil || tenantID <= 0 {
                    return echo.NewHTTPError(http.StatusBadRequest, "invalid X-Tenant-ID header")
                }

                // Validate tenant ID
                if err := tenant.ValidateTenantID(tenantID); err != nil {
                    return echo.NewHTTPError(http.StatusBadRequest, err.Error())
                }

                // Add tenant to context
                ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
                c.SetRequest(c.Request().WithContext(ctx))
            } else if config.DefaultTenant > 0 {
                // Use default tenant
                ctx := tenant.WithTenantID(c.Request().Context(), config.DefaultTenant)
                c.SetRequest(c.Request().WithContext(ctx))
            } else {
                // No tenant context available
                return echo.NewHTTPError(http.StatusUnauthorized, "missing tenant identification")
            }

            return next(c)
        }
    }
}

// TenantMiddlewareFromOperator creates middleware that extracts tenant ID from operator.
// Use this when operator authentication is done and tenant_id is available from operator object.
func TenantMiddlewareFromOperator(getTenantIDFunc func() (int64, error)) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            tenantID, err := getTenantIDFunc()
            if err == nil && tenantID > 0 {
                if err := tenant.ValidateTenantID(tenantID); err != nil {
                    return echo.NewHTTPError(http.StatusBadRequest, err.Error())
                }
                ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
                c.SetRequest(c.Request().WithContext(ctx))
            }
            return next(c)
        }
    }
}

// RequireTenant checks if request has tenant context and returns error if not.
// Use this in individual handlers that require tenant context.
func RequireTenant() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            _, err := tenant.FromContext(c.Request().Context())
            if err != nil {
                return echo.NewHTTPError(http.StatusUnauthorized, "tenant context required")
            }
            return next(c)
        }
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/middleware -run TestTenant -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/middleware/tenant.go internal/middleware/tenant_test.go
git commit -m "feat(middleware): add tenant isolation middleware with skip paths support"
```

---

## Task 3: Create Tenant-Aware Query Builder

**Files:**
- Create: `internal/repository/tenant_scope.go`
- Create: `internal/repository/tenant_scope_test.go`

**Step 1: Write tests for tenant-aware query builder**

```go
// internal/repository/tenant_scope_test.go
package repository

import (
    "context"
    "testing"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/tenant"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to open test database: %v", err)
    }

    // Migrate tables
    db.AutoMigrate(&domain.RadiusUser{})

    return db
}

func TestTenantScope(t *testing.T) {
    db := setupTestDB(t)

    // Create test users for different tenants
    users := []domain.RadiusUser{
        {Username: "user1", TenantID: 1, Status: "enabled"},
        {Username: "user2", TenantID: 1, Status: "enabled"},
        {Username: "user3", TenantID: 2, Status: "enabled"},
    }
    db.Create(&users)

    // Create context with tenant ID
    ctx := tenant.WithTenantID(context.Background(), 1)

    // Query with tenant scope
    var results []domain.RadiusUser
    err := db.WithContext(ctx).Scopes(TenantScope).Find(&results).Error
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    // Should only return users from tenant 1
    if len(results) != 2 {
        t.Errorf("Expected 2 users, got %d", len(results))
    }

    for _, user := range results {
        if user.TenantID != 1 {
            t.Errorf("Expected tenant ID 1, got %d", user.TenantID)
        }
    }
}

func TestTenantScopeWithAdmin(t *testing.T) {
    db := setupTestDB(t)

    // Create test users
    users := []domain.RadiusUser{
        {Username: "user1", TenantID: 1, Status: "enabled"},
        {Username: "user2", TenantID: 2, Status: "enabled"},
    }
    db.Create(&users)

    // Create context with platform admin (no tenant filter)
    ctx := context.Background()

    // Admin query bypasses tenant scope
    var results []domain.RadiusUser
    err := db.WithContext(ctx).Scopes(AdminTenantScope(1)).Find(&results).Error
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    // Should return users from tenant 1 (admin specified)
    if len(results) != 1 {
        t.Errorf("Expected 1 user, got %d", len(results))
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/repository -run TestTenantScope -v`
Expected: FAIL with "undefined: TenantScope"

**Step 3: Implement tenant-aware query scopes**

```go
// internal/repository/tenant_scope.go
package repository

import (
    "context"

    "github.com/talkincode/toughradius/v9/internal/tenant"
    "gorm.io/gorm"
)

// TenantScope returns a GORM scope that filters by tenant_id from context.
// Use this in all queries to ensure tenant isolation.
// Example: db.Scopes(TenantScope).Find(&users)
func TenantScope(db *gorm.DB) *gorm.DB {
    ctx := db.Statement.Context
    if ctx == nil {
        return db
    }

    tenantID, err := tenant.FromContext(ctx)
    if err != nil {
        // No tenant context, return empty result
        return db.Where("1 = 0")
    }

    return db.Where("tenant_id = ?", tenantID)
}

// AdminTenantScope allows platform admin to query specific tenant.
// Use this in admin APIs where admin can access any tenant data.
func AdminTenantScope(tenantID int64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}

// AllTenantsScope bypasses tenant filtering (platform admin only).
// Use with caution - only for platform-level aggregation queries.
func AllTenantsScope(db *gorm.DB) *gorm.DB {
    return db
}

// TenantScopeWithID returns a scope for a specific tenant ID.
// Use when you need to query a different tenant than the current context.
func TenantScopeWithID(tenantID int64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}

// WithTenant creates a new DB instance with tenant context.
// Convenience function for queries with tenant context.
func WithTenant(db *gorm.DB, tenantID int64) *gorm.DB {
    ctx := tenant.WithTenantID(context.Background(), tenantID)
    return db.WithContext(ctx).Scopes(TenantScope)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/repository -run TestTenantScope -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/repository/tenant_scope.go internal/repository/tenant_scope_test.go
git commit -m "feat(repository): add tenant-aware query scopes for data isolation"
```

---

## Task 4: Integrate Tenant Middleware into Web Server

**Files:**
- Modify: `internal/webserver/server.go`
- Modify: `internal/adminapi/adminapi.go`

**Step 1: Register tenant middleware in web server**

```go
// In internal/webserver/server.go

import (
    "github.com/talkincode/toughradius/v9/internal/middleware"
)

// Add to server initialization
func Init(app *app.Application) {
    e := echo.New()

    // Configure tenant middleware
    tenantMiddleware := middleware.TenantMiddleware(middleware.TenantMiddlewareConfig{
        SkipPaths: []string{
            "/health",
            "/ready",
            "/metrics",
            "/api/v1/public/login",
            "/api/v1/public/register",
        },
        DefaultTenant: 1, // Platform admin tenant
    })

    // Apply tenant middleware globally
    e.Use(tenantMiddleware)

    // ... rest of server initialization
}
```

**Step 2: Update admin API to use tenant scopes**

```go
// In internal/adminapi/users.go (and other API files)

import (
    "github.com/talkincode/toughradius/v9/internal/repository"
)

// Update query functions to use tenant scope
func GetUsers(c echo.Context) error {
    db := GetDB(c)

    // Automatically applies tenant filter from context
    var users []domain.RadiusUser
    err := db.Scopes(repository.TenantScope).Find(&users).Error
    if err != nil {
        return fail(c, 500, "DATABASE_ERROR", "Failed to fetch users", err)
    }

    return ok(c, users)
}
```

**Step 3: Test tenant isolation**

```bash
# Test as tenant 1
curl -H "X-Tenant-ID: 1" http://localhost:1816/api/v1/users

# Test as tenant 2 (should get different users)
curl -H "X-Tenant-ID: 2" http://localhost:1816/api/v1/users

# Test without tenant header (should get error)
curl http://localhost:1816/api/v1/users
```

**Step 4: Commit**

```bash
git add internal/webserver/server.go internal/adminapi/
git commit -m "feat(webserver): integrate tenant middleware for request isolation"
```

---

## Testing & Verification

**Integration Test:**
```bash
# 1. Start server with tenant middleware
./toughradius -c toughradius.yml

# 2. Test tenant isolation
curl -H "X-Tenant-ID: 1" http://localhost:1816/api/v1/users
curl -H "X-Tenant-ID: 2" http://localhost:1816/api/v1/users

# 3. Verify cross-tenant access blocked
curl -H "X-Tenant-ID: 1" http://localhost:1816/api/v1/users/999  # user from tenant 2
# Should return 404 or 403
```

**Security Test:**
```go
// Test attempts to bypass tenant isolation
func TestCrossTenantAccessBlocked(t *testing.T) {
    // Create user in tenant 1
    // Try to access it with tenant 2 context
    // Should fail with authorization error
}
```

---

## Documentation Updates

**Files:**
- Create: `docs/multi-tenant/tenant-isolation.md`

**Documentation Content:**
```markdown
# Tenant Isolation Guide

## Middleware

All requests must include tenant identification:

```bash
curl -H "X-Tenant-ID: 1" http://localhost:1816/api/v1/users
```

## Query Scopes

Use tenant scopes in all database queries:

```go
// Automatic from request context
db.Scopes(repository.TenantScope).Find(&users)

// Platform admin querying specific tenant
db.Scopes(repository.AdminTenantScope(tenantID)).Find(&users)
```

## Security Rules

1. All API requests must have tenant context (except public endpoints)
2. Database queries automatically filter by tenant_id
3. Platform admins can query across all tenants
4. Cross-tenant access is blocked at middleware and database layers
```

**Step: Commit documentation**

```bash
git add docs/multi-tenant/
git commit -m "docs: add tenant isolation guide for developers"
```

---

## Success Criteria

- ✅ Tenant context properly extracted from requests
- ✅ Middleware enforces tenant presence in requests
- ✅ Database queries automatically scoped to tenant
- ✅ Cross-tenant access blocked at all layers
- ✅ Unit tests pass (≥80% coverage)
- ✅ Integration tests validate isolation
