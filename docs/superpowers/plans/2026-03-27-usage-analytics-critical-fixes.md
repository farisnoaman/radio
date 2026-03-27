# Critical Fixes for Usage Analytics Implementation Plan

**Date:** 2026-03-27
**Status:** Amendments to Implementation Plan
**Reviewers:** 4 Parallel Agents (Security, Performance, Code Quality, Architecture)

---

## Executive Summary

This document contains **15 critical fixes** identified by 4 specialized review agents that MUST be applied to the implementation plan before production deployment. These fixes address security vulnerabilities, performance blockers, code quality issues, and architectural gaps.

---

## 🔴 CRITICAL FIXES (Must Apply)

### Fix #1: Replace Redis KEYS with SCAN (Performance Blocker)

**Location:** Task 6.1, Line 2464
**Severity:** CRITICAL - Will cause Redis to block at scale
**Reviewer:** Performance Agent

**Problem:**
```go
// WRONG - Blocks Redis at 100K users
keys, err := sw.redis.Keys(ctx, "user:usage:*").Result()
```

**Solution:**
```go
// CORRECT - Use SCAN with cursor
func (sw *SyncWorker) SyncUsageToPostgreSQL(ctx context.Context) error {
    zap.L().Debug("Starting Redis → PostgreSQL sync")

    // Use SCAN instead of KEYS to avoid blocking Redis
    var keys []string
    var cursor uint64
    pattern := "user:usage:*"

    for {
        var batch []string
        var err error

        // SCAN with cursor (non-blocking)
        batch, cursor, err = sw.redis.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil && err != redis.Nil {
            zap.L().Error("SCAN failed", zap.Error(err))
            return err
        }

        keys = append(keys, batch...)

        if cursor == 0 {
            break // All keys scanned
        }
    }

    zap.L().Debug("Found usage keys", zap.Int("count", len(keys)))

    // Batch process (existing code)
    for i := 0; i < len(keys); i += sw.batchSize {
        end := min(i+sw.batchSize, len(keys))
        batch := keys[i:end]

        if err := sw.syncBatch(ctx, batch); err != nil {
            zap.L().Error("Batch sync failed",
                zap.Int("batch_start", i),
                zap.Error(err))
        }
    }

    zap.L().Info("Sync completed", zap.Int("total_keys", len(keys)))
    return nil
}
```

**Impact:** Prevents Redis blocking (1000ms+ latency spikes) at 100K users.

---

### Fix #2: Replace sync.Map with Bounded LRU Cache (Memory Safety)

**Location:** Task 2.3, Line 875
**Severity:** CRITICAL - Will cause OOM crashes
**Reviewer:** Performance Agent

**Problem:**
```go
// WRONG - Unbounded memory growth
l1Cache *sync.Map  // Will grow until OOM
```

**Solution:**
```go
// CORRECT - Bounded LRU cache
package cache

import (
    "github.com/hashicorp/golang-lru/v2"
    "context"
    "time"
)

type SessionCache struct {
    l1Cache *lru.Cache  // Bounded LRU (max 10K entries)
    l2Cache *redis.Client
    db      *gorm.DB
}

func NewSessionCache(l2Cache *redis.Client, db *gorm.DB) *SessionCache {
    // Create LRU cache with max 10,000 entries (~10MB memory)
    cache, _ := lru.NewEvicted[string, *UsageSnapshot](10000)

    return &SessionCache{
        l1Cache: cache,
        l2Cache: l2Cache,
        db:      db,
    }
}

func (sc *SessionCache) Get(ctx context.Context, username string, tenantID int64) (*UserUsage, error) {
    // Try L1 first (<1ms)
    if val, ok := sc.l1Cache.Get(username); ok {
        return val.(*UserUsage), nil
    }

    // Try L2 Redis (<10ms)
    usageKey := fmt.Sprintf("user:usage:%d:%s", tenantID, username)
    val, err := sc.l2Cache.Get(ctx, usageKey).Result()
    if err == nil {
        var usage UserUsage
        if err := json.Unmarshal([]byte(val), &usage); err == nil {
            // Populate L1 (evicts oldest if full)
            sc.l1Cache.Add(username, &usage)
            return &usage, nil
        }
    }

    // Fallback to DB (<100ms)
    // ... (existing DB query code)
}
```

**Impact:** Prevents OOM crashes; limits memory to ~10MB per server instance.

---

### Fix #3: Implement JWT Authentication Middleware

**Location:** Before Task 3.1
**Severity:** CRITICAL - Authentication bypass vulnerability
**Reviewer:** Security Agent, Architecture Agent

**Add New Task 2.5: Implement JWT Authentication**

```go
// internal/middleware/auth.go

package middleware

import (
    "net/http"
    "strings"
    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
)

type SecurityContext struct {
    UserID   int64
    TenantID int64
    Role     string
    Username string
}

type JWTClaims struct {
    UserID   int64  `json:"user_id"`
    TenantID int64  `json:"tenant_id"`
    Role     string `json:"role"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

func JWTAuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Extract token from Authorization header
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "Missing authorization header",
                })
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "Invalid authorization format",
                })
            }

            // Validate token
            token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
                }
                return []byte(jwtSecret), nil
            })

            if err != nil || !token.Valid {
                zap.L().Warn("Invalid JWT token", zap.Error(err))
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "Invalid or expired token",
                })
            }

            // Extract claims
            if claims, ok := token.Claims.(*JWTClaims); ok {
                secCtx := SecurityContext{
                    UserID:   claims.UserID,
                    TenantID: claims.TenantID,
                    Role:     claims.Role,
                    Username: claims.Username,
                }
                c.Set("security", secCtx)
            }

            return next(c)
        }
    }
}
```

**Integration:** Apply to ALL API routes before `requirePermission`.

---

### Fix #4: Implement RBAC Permission System

**Location:** Before Task 7.1
**Severity:** CRITICAL - Authorization bypass vulnerability
**Reviewer:** Security Agent

**Add New Task 2.6: Implement RBAC System**

```go
// internal/middleware/rbac.go

package middleware

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
)

// Role permissions mapping
var rolePermissions = map[string][]string{
    "user": {
        "users:read:own",
        "usage:read:own",
        "sessions:read:own",
    },
    "operator": {
        "users:read:tenant",
        "usage:read:tenant",
        "sessions:read:tenant",
        "users:write:tenant",
    },
    "tenant_admin": {
        "users:read:tenant",
        "usage:read:tenant",
        "sessions:read:tenant",
        "users:write:tenant",
        "usage:aggregates:read:tenant",
    },
    "platform_admin": {
        "*", // All permissions
    },
}

func hasPermission(secCtx SecurityContext, requiredPermission string) bool {
    // Platform admins have all permissions
    if secCtx.Role == "platform_admin" {
        return true
    }

    permissions, exists := rolePermissions[secCtx.Role]
    if !exists {
        return false
    }

    for _, perm := range permissions {
        if perm == requiredPermission {
            return true
        }
    }
    return false
}

func requirePermission(permission string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            secCtx, ok := c.Get("security").(SecurityContext)
            if !ok {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "Not authenticated",
                })
            }

            if !hasPermission(secCtx, permission) {
                zap.L().Warn("Permission denied",
                    zap.Int64("user_id", secCtx.UserID),
                    zap.String("role", secCtx.Role),
                    zap.String("required", permission))
                return c.JSON(http.StatusForbidden, map[string]string{
                    "error": "Insufficient permissions",
                })
            }

            return next(c)
        }
    }
}
```

---

### Fix #5: Standardize Redis Data Structure (Hash Pattern)

**Location:** Task 2.1, Lines 586, 651
**Severity:** CRITICAL - Runtime incompatibility
**Reviewer:** Architecture Agent, Code Quality Agent

**Problem:** Plan uses both JSON and Hash patterns inconsistently.

**Solution:** Use Hash pattern throughout (as required by HINCRBY):

```go
// internal/usage/tracker.go - Replace GetUserUsage method

func (ut *UsageTracker) GetUserUsage(ctx context.Context, tenantID int64, username string) (*UserUsage, error) {
    usageKey := ut.usageKey(tenantID, username)

    // Try Redis hash first (L2 cache)
    vals, err := ut.redis.HGetAll(ctx, usageKey).Result()
    if err == nil && len(vals) > 0 {
        // Parse hash fields
        timeUsed, _ := strconv.ParseInt(vals["time_quota_used"], 10, 64)
        dataUsed, _ := strconv.ParseInt(vals["data_quota_used"], 10, 64)
        // ... parse other fields

        return &UserUsage{
            TenantID:       tenantID,
            Username:       username,
            TimeQuotaUsed:  timeUsed,
            DataQuotaUsed:   dataUsed,
            // ...
        }, nil
    }

    // Cache miss - calculate from PostgreSQL
    // ... (existing DB query code)

    // Store as hash in Redis (not JSON)
    pipe := ut.redis.Pipeline()
    pipe.HSet(ctx, usageKey, map[string]interface{}{
        "tenant_id":         usage.TenantID,
        "username":           usage.Username,
        "time_quota_total":   usage.TimeQuotaTotal,
        "time_quota_used":    usage.TimeQuotaUsed,
        "time_quota_remaining": usage.TimeQuotaRemaining,
        "data_quota_total":   usage.DataQuotaTotal,
        "data_quota_used":    usage.DataQuotaUsed,
        "data_quota_remaining": usage.DataQuotaRemaining,
    })
    pipe.Expire(ctx, usageKey, 60*time.Second)
    pipe.Exec(ctx)

    return usage, nil
}
```

---

### Fix #6: Add GDPR Compliance Features

**Location:** New Phase 11
**Severity:** CRITICAL - Legal compliance requirement
**Reviewer:** Security Agent, Architecture Agent

**Add Phase 11: GDPR Compliance (Week 11)**

```markdown
## Phase 11: GDPR Compliance (Week 11)

### Task 11.1: Implement Data Export Endpoint

**Files:**
- Create: `internal/adminapi/gdpr.go`
- Create: `internal/adminapi/gdpr_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestExportUserData_Success(t *testing.T) {
    secCtx := SecurityContext{
        UserID:   123,
        TenantID: 1,
        Username: "testuser",
    }

    data, err := ExportUserData(secCtx)
    if err != nil {
        t.Fatalf("ExportUserData failed: %v", err)
    }

    if len(data) == 0 {
        t.Error("Expected non-empty export")
    }
}
```

- [ ] **Step 2: Implement export endpoint**

```go
// GET /api/v1/users/me/gdpr/export
func (h *UsageAPIHandler) ExportUserData(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    // Collect all user data
    var userData struct {
        Username        string
        Email           string
        CreatedAt       time.Time
        AccountingRecords []RadiusAccounting
        Sessions        []RadiusOnline
    }

    h.db.Table("radius_user").
        Where("tenant_id = ? AND username = ?", secCtx.TenantID, secCtx.Username).
        First(&userData)

    h.db.Where("tenant_id = ? AND username = ?", secCtx.TenantID, secCtx.Username).
        Find(&userData.AccountingRecords)

    // Generate CSV
    c.Response().Header().Set("Content-Type", "text/csv")
    c.Response().Header().Set("Content-Disposition",
        fmt.Sprintf("attachment; filename=user_data_%s_%s.csv",
            secCtx.Username, time.Now().Format("20060102")))

    writer := csv.NewWriter(c.Response())
    defer writer.Flush()

    // Write CSV data
    // ... (CSV writing logic)

    return nil
}
```

### Task 11.2: Implement Right to be Forgotten

- [ ] **Step 1: Implement anonymization**

```go
// DELETE /api/v1/users/me/gdpr/delete
func (h *UsageAPIHandler) DeleteUserData(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    err := h.db.Transaction(func(tx *gorm.DB) error {
        // Anonymize radius_user
        tx.Table("radius_user").
            Where("tenant_id = ? AND username = ?", secCtx.TenantID, secCtx.Username).
            Updates(map[string]interface{}{
                "username": fmt.Sprintf("deleted_%d_%s", secCtx.UserID, generateRandomString(8)),
                "email":    fmt.Sprintf("deleted_%d@example.com", secCtx.UserID),
                "phone":    "",
                "realname": "Deleted User",
            })

        // Anonymize radius_accounting
        tx.Table("radius_accounting").
            Where("tenant_id = ? AND username = ?", secCtx.TenantID, secCtx.Username).
            Update("username", fmt.Sprintf("deleted_%d", secCtx.UserID))

        // Delete from Redis cache
        h.redis.Del(c.Request().Context(),
            fmt.Sprintf("user:usage:%d:%s", secCtx.TenantID, secCtx.Username))

        return nil
    })

    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to delete data",
        })
    }

    // Log consent withdrawal
    zap.L().Info("GDPR deletion requested",
        zap.Int64("user_id", secCtx.UserID),
        zap.String("username", secCtx.Username))

    return c.JSON(http.StatusOK, map[string]string{
        "message": "Data anonymized",
    })
}
```

### Task 11.3: Implement Data Retention Policy

- [ ] **Step 1: Create retention job**

```go
// internal/jobs/retention_policy.go

func EnforceRetentionPolicy(db *gorm.DB) error {
    // Delete accounting records older than 12 months
    cutoffDate := time.Now().AddDate(-1, 0, 0)

    result := db.
        Where("acct_stop_time < ? AND acct_status = ?", cutoffDate, "stopped").
        Delete(&RadiusAccounting{})

    zap.L().Info("Enforced data retention policy",
        zap.Int64("deleted_count", result.RowsAffected))

    return result.Error
}
```
```

---

### Fix #7: Add Input Validation Package

**Location:** Before Task 3.1
**Severity:** HIGH - SQL injection and DoS vulnerabilities
**Reviewer:** Security Agent

**Add New Task 2.7: Implement Input Validation**

```go
// internal/validator/usage_validator.go

package validator

import (
    "time"
    "github.com/go-playground/validator/v10"
)

type UsageHistoryRequest struct {
    From        string `json:"from" validate:"required,rfc3339"`
    To          string `json:"to" validate:"required,rfc3339"`
    Granularity string `json:"granularity" validate:"oneof=hourly daily weekly monthly"`
}

var validate = validator.New()

func ValidateUsageHistoryRequest(from, to, granularity string) (time.Time, time.Time, error) {
    req := UsageHistoryRequest{
        From:        from,
        To:          to,
        Granularity: granularity,
    }

    if err := validate.Struct(req); err != nil {
        return time.Time{}, time.Time{}, fmt.Errorf("validation failed: %w", err)
    }

    // Parse dates
    fromDate, err := time.Parse(time.RFC3339, from)
    if err != nil {
        return time.Time{}, time.Time{}, fmt.Errorf("invalid from date: %w", err)
    }

    toDate, err := time.Parse(time.RFC3339, to)
    if err != nil {
        return time.Time{}, time.Time{}, fmt.Errorf("invalid to date: %w", err)
    }

    // Validate range (max 1 year)
    if toDate.Sub(fromDate) > 365*24*time.Hour {
        return time.Time{}, time.Time{}, fmt.Errorf("date range too large (max 1 year)")
    }

    // Validate order
    if toDate.Before(fromDate) {
        return time.Time{}, time.Time{}, fmt.Errorf("to date must be after from date")
    }

    // Validate not in future
    if toDate.After(time.Now()) {
        return time.Time{}, time.Time{}, fmt.Errorf("to date cannot be in future")
    }

    return fromDate, toDate, nil
}
```

---

### Fix #8: Add Automatic Tenant Isolation

**Location:** Task 7.1
**Severity:** CRITICAL - Cross-tenant data leakage
**Reviewer:** Security Agent, Architecture Agent

**Modify Task 7.1 to add:**

```go
// internal/middleware/tenant_filter.go

package middleware

import (
    "github.com/labstack/echo/v4"
    "gorm.io/gorm"
)

// ApplyTenantFilter automatically adds tenant WHERE clause
func ApplyTenantFilter(db *gorm.DB, c echo.Context) *gorm.DB {
    secCtx, ok := c.Get("security").(SecurityContext)
    if !ok {
        return db
    }

    // Platform admins can see all data (with audit logging)
    if secCtx.Role == "platform_admin" {
        zap.L().Warn("Platform admin accessing all tenants",
            zap.Int64("admin_id", secCtx.UserID),
            zap.String("path", c.Path()))
        return db
    }

    // Enforce tenant isolation
    return db.Where("tenant_id = ?", secCtx.TenantID)
}

// Usage in queries:
query = h.db.Table("radius_accounting").
    Select("*").
    Scopes(ApplyTenantFilter(c)).  // AUTOMATIC filtering
    Where("acct_status = ?", "active")
```

---

### Fix #9: Increase Redis Connection Pool Size

**Location:** Task 1.3, Line 428
**Severity:** HIGH - Connection exhaustion at scale
**Reviewer:** Performance Agent

**Change:**
```go
// internal/app/redis.go - Modify NewRedisClient

func NewRedisClient(config *RedisConfig) (*redis.Client, error) {
    if config.PoolSize == 0 {
        config.PoolSize = 2000  // Increased from 500 for 100K users
    }
    if config.MinIdleConns == 0 {
        config.MinIdleConns = 100  // Increased from 50
    }
    if config.MaxRetries == 0 {
        config.MaxRetries = 3
    }

    client := redis.NewClient(&redis.Options{
        Addr:         config.Addr,
        Password:     config.Password,
        PoolSize:     config.PoolSize,
        MinIdleConns: config.MinIdleConns,
        MaxRetries:   config.MaxRetries,
        DialTimeout:  50 * time.Millisecond,
        ReadTimeout:  100 * time.Millisecond,
        WriteTimeout: 100 * time.Millisecond,
        PoolTimeout:  4 * time.Second,
    })

    // ... rest of code
}
```

---

### Fix #10: Combine N+1 Queries

**Location:** Task 2.1, Lines 602-621
**Severity:** MEDIUM - Performance optimization
**Reviewer:** Performance Agent, Code Quality Agent

**Replace:**
```go
// BEFORE: Two separate queries
ut.db.Model(&RadiusAccounting{}).
    Where("tenant_id = ? AND username = ? AND acct_status = ?", tenantID, username, "stopped").
    Select("COALESCE(SUM(acct_session_time), 0)").
    Scan(&timeUsed)

ut.db.Model(&RadiusAccounting{}).
    Where("tenant_id = ? AND username = ? AND acct_status = ?", tenantID, username, "stopped").
    Select("COALESCE(SUM(acct_input_octets + acct_output_octets), 0)").
    Scan(&dataUsed)

// AFTER: Single combined query
var result struct {
    TimeUsed int64 `gorm:"column:time_used"`
    DataUsed int64 `gorm:"column:data_used"`
}

ut.db.Model(&RadiusAccounting{}).
    Where("tenant_id = ? AND username = ? AND acct_status = ?", tenantID, username, "stopped").
    Select(`
        COALESCE(SUM(acct_session_time), 0) as time_used,
        COALESCE(SUM(acct_input_octets + acct_output_octets), 0) as data_used
    `).
    Scan(&result)

timeUsed = result.TimeUsed
dataUsed = result.DataUsed
```

---

### Fix #11: Add acct_status Migration with Backfill Strategy

**Location:** Task 1.2
**Severity:** HIGH - Data integrity risk
**Reviewer:** Code Quality Agent

**Add to Task 1.2 Step 3:**

```go
// internal/migrations/20260327_usage_analytics.go - Add backfill function

func BackfillAcctStatus(db *gorm.DB) error {
    // For large tables, backfill in batches
    batchSize := 1000
    var maxID int64

    for {
        var count int64

        // Process batch
        result := db.Exec(`
            UPDATE radius_accounting
            SET acct_status = CASE
                WHEN acct_stop_time IS NULL OR acct_stop_time = '0001-01-01 00:00:00' THEN 'active'
                ELSE 'stopped'
            END
            WHERE radacctid > ? AND radacctid <= ? AND acct_status IS NULL
        `, maxID, maxID+batchSize)

        if result.Error != nil {
            return result.Error
        }

        count = result.RowsAffected
        if count == 0 {
            break // No more rows to update
        }

        maxID += batchSize
        zap.L().Info("Backfilled acct_status",
            zap.Int64("processed", maxID))
    }

    return nil
}
```

---

### Fix #12: Enhance WebSocket Security

**Location:** Task 4.2
**Severity:** HIGH - Security vulnerabilities
**Reviewer:** Security Agent, Code Quality Agent

**Add to Task 4.2:**

```go
// internal/adminapi/websocket_middleware.go - Additions

import (
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:   1024,
    WriteBufferSize:  1024,
    MaxMessageSize:   65536, // 64KB max message size
    HandshakeTimeout: 5 * time.Second,
    // CRITICAL: Load from config, not hardcode
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        host := r.Host

        // Load from environment/config
        allowedOrigins := []string{
            getEnv("WS_ALLOWED_ORIGINS", ""),
        }

        for _, allowed := range allowedOrigins {
            if origin == allowed {
                return true
            }
        }

        zap.L().Warn("WebSocket rejected: cross-origin request",
            zap.String("origin", origin),
            zap.String("host", host))
        return false
    },
}

// Add connection rate limiting
func WebSocketRateLimiter(redis *redis.Client) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            secCtx, _ := c.Get("security").(SecurityContext)

            // Check connection count per user
            key := fmt.Sprintf("ws:connections:%d", secCtx.UserID)
            count, _ := redis.Incr(c.Request().Context(), key).Result()
            redis.Expire(c.Request().Context(), key, 1*time.Hour)

            if count > 5 { // Max 5 concurrent WebSocket connections
                return c.JSON(http.StatusTooManyRequests, map[string]string{
                    "error": "Too many WebSocket connections",
                })
            }

            return next(c)
        }
    }
}
```

---

### Fix #13: Add Load Testing at Scale

**Location:** New Task 9.3
**Severity:** MEDIUM - Validation of performance claims
**Reviewer:** Performance Agent, Code Quality Agent

**Add New Task 9.3: Scale Load Testing**

```markdown
### Task 9.3: Load Test at Target Scale

**Files:**
- Create: `tests/load/scale_test.go`

- [ ] **Step 1: Write 100K user load test**

```go
// tests/load/scale_test.go

func TestUsageTracker_100KConcurrentUsers(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test")
    }

    db := setupTestDB()
    redis := setupTestRedis()
    defer teardown(db, redis)

    tracker := usage.NewUsageTracker(db, redis)
    ctx := context.Background()

    numUsers := 100000
    reqPerUser := 10
    var latencies []time.Duration

    start := time.Now()

    // Run concurrent updates
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 1000) // Limit concurrency

    for i := 0; i < numUsers; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release

            for j := 0; j < reqPerUser; j++ {
                reqStart := time.Now()

                record := &RadiusAccounting{
                    TenantID:       1,
                    UserName:       fmt.Sprintf("user_%d", userID),
                    AcctStatus:     "stopped",
                    AcctSessionTime: 60, // 1 minute
                    AcctInputOctets: 1024,
                    AcctOutputOctets: 1024,
                }

                tracker.RecordAccountingUpdate(ctx, record)

                latencies = append(latencies, time.Since(reqStart))
            }
        }(i)
    }

    wg.Wait()
    duration := time.Since(start)

    // Calculate percentiles
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })

    p50 := latencies[len(latencies)*50/100]
    p95 := latencies[len(latencies)*95/100]
    p99 := latencies[len(latencies)*99/100]

    t.Logf("Completed %d updates in %v", numUsers*reqPerUser, duration)
    t.Logf("Latencies: P50=%v, P95=%v, P99=%v", p50, p95, p99)

    // Assert performance targets
    if p99 > 100*time.Millisecond {
        t.Errorf("P99 latency %v exceeds target of 100ms", p99)
    }
}
```
```

---

### Fix #14: Add Security Headers Middleware

**Location:** Before Task 3.1
**Severity:** MEDIUM - Security hardening
**Reviewer:** Security Agent

**Add New Task 2.8: Implement Security Headers**

```go
// internal/middleware/security_headers.go

package middleware

import (
    "github.com/labstack/echo/v4"
)

func SecurityHeaders() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Prevent MIME sniffing
            c.Response().Header().Set("X-Content-Type-Options", "nosniff")

            // Prevent clickjacking
            c.Response().Header().Set("X-Frame-Options", "DENY")

            // Enable XSS protection
            c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

            // Force HTTPS
            c.Response().Header().Set("Strict-Transport-Security",
                "max-age=31536000; includeSubDomains")

            // CSP
            c.Response().Header().Set("Content-Security-Policy",
                "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

            return next(c)
        }
    }
}
```

**Apply globally:**
```go
e.Use(middleware.SecurityHeaders())
```

---

### Fix #15: Add Refactor Step to All Tasks

**Location:** All tasks
**Severity:** MEDIUM - TDD compliance
**Reviewer:** Code Quality Agent

**Update all tasks to include refactor step:**

```markdown
- [ ] **Step 4: Run test to verify it passes**

Run: `pytest tests/path/test.py::test_name -v`
Expected: PASS

- [ ] **Step 5: Refactor** (NEW STEP)
   - Extract magic numbers to named constants
   - Improve code organization
   - Add comments for complex logic
   - Ensure naming conventions

- [ ] **Step 6: Run tests again after refactor**

Run: `pytest tests/path/test.py::test_name -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add tests/path/test.py src/path/file.py
git commit -m "feat: add specific feature"
```
```

---

## 📋 Implementation Order for Fixes

### Phase 0: Pre-Implementation (Before Week 1)

1. Fix #3: JWT Authentication Middleware
2. Fix #4: RBAC Permission System
3. Fix #7: Input Validation Package
4. Fix #8: Automatic Tenant Isolation
5. Fix #14: Security Headers Middleware

### Phase 1: Foundation (Week 1 - Modify Existing Tasks)

6. Fix #5: Standardize Redis Data Structure (Task 2.1)
7. Fix #2: Replace sync.Map with LRU (Task 2.3)
8. Fix #9: Increase Connection Pool (Task 1.3)
9. Fix #11: acct_status Backfill Strategy (Task 1.2)

### Phase 2: Core Features (Week 2-3 - Modify Existing Tasks)

10. Fix #10: Combine N+1 Queries (Task 2.1)
11. Fix #1: Replace KEYS with SCAN (Task 6.1)
12. Fix #12: WebSocket Security (Task 4.2)

### Phase 3: Testing (Week 9 - Add New Tasks)

13. Fix #13: Add Scale Load Testing (Task 9.3)

### Phase 4: Compliance (Week 11 - New Phase)

14. Fix #6: GDPR Compliance Features (New Phase 11)

### Phase 5: Code Quality (All Tasks)

15. Fix #15: Add Refactor Step (All tasks)

---

## ✅ Validation Checklist

After applying all fixes, the plan should meet:

- [ ] **Security**: JWT auth, RBAC, tenant isolation, input validation, GDPR
- [ ] **Performance**: <100ms P99, 100K concurrent users, no blocking operations
- [ ] **Code Quality**: Genuine TDD, proper error handling, bounded caches
- [ ] **Architecture**: Consistent Redis patterns, all spec features implemented
- [ ] **Compliance**: GDPR data export, anonymization, retention policy

---

## 📊 Updated Plan Statistics

- **Original Tasks**: 45 tasks
- **New Tasks Added**: 7 tasks
- **Tasks Modified**: 15 tasks
- **Total Tasks**: 52 tasks
- **New Timeline**: 11 weeks (was 10 weeks)

---

## 🚀 Next Steps

1. **Apply all 15 fixes** to the implementation plan
2. **Re-dispatch reviewers** to validate fixes
3. **Get final approval** to proceed with implementation
4. **Begin Phase 0** (pre-implementation fixes)

---

**Document Version:** 1.0
**Last Updated:** 2026-03-27
**Status:** Ready for Application to Plan
