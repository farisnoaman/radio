# Enterprise Usage Analytics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an enterprise-scale usage analytics system supporting 100,000+ concurrent users with Redis-primary architecture for real-time quota tracking, predictive analytics, and role-based access control.

**Architecture:** Redis as primary usage store with atomic HINCRBY operations for real-time updates, PostgreSQL as source of truth for persistence, multi-level caching (L1 in-memory → L2 Redis → L3 PostgreSQL), background workers for async sync every 60 seconds.

**Tech Stack:** Go 1.24+, Echo framework, PostgreSQL 14+, Redis 7+, React with TypeScript, WebSocket with Redis Pub/Sub, Prometheus metrics, existing `internal/analytics/predictive.go` extended with pattern detection.

---

## File Structure Overview

### New Files to Create

```
internal/
├── usage/
│   ├── tracker.go              # UsageTracker with Redis write-through caching
│   ├── tracker_test.go         # TDD tests for tracker
│   ├── cache.go                # Multi-level cache (L1 sync.Map + L2 Redis)
│   ├── cache_test.go           # Cache tests
│   ├── sync_worker.go          # Background worker: Redis → PostgreSQL sync
│   ├── sync_worker_test.go     # Sync worker tests
│   └── repository.go           # UsageRepository interface
├── analytics/
│   ├── pattern_detector.go     # Pattern detection for usage trends
│   ├── pattern_detector_test.go
│   ├── anomaly_detector.go     # Anomaly detection
│   ├── anomaly_detector_test.go
│   └── enhanced_predictive.go  # Extends existing predictive.go
├── adminapi/
│   ├── usage_api.go            # New API endpoints: /users/me/usage, etc.
│   ├── usage_api_test.go       # API tests
│   ├── websocket_hub.go        # WebSocket hub for real-time updates
│   └── middleware.go            # Security middleware (RBAC, rate limiting)
├── middleware/
│   ├── security.go             # Enhanced security with RBAC (modify existing)
│   └── rate_limiter.go         # Per-endpoint rate limiting
└── migrations/
    └── 20260327_usage_analytics.go  # Database migration (acct_status column)

web/src/
├── pages/
│   ├── UsageStatus.tsx         # User-facing usage status page
│   └── UsageHistory.tsx        # Historical usage charts
├── components/
│   └── usage/
│       ├── UsageCard.tsx       # Reusable usage display card
│       └── UsageChart.tsx      # Chart component
└── resources/
    └── users.tsx               # Extend with usage endpoint

scripts/
└── migrate_usage_analytics.sh  # Database migration script
```

### Files to Modify

```
internal/
├── domain/radius.go            # Add acct_status field to RadiusAccounting
├── adminapi/adminapi.go        # Register new usage routes
├── adminapi/context.go         # Add security context
├── radiusd/plugins/accounting/handlers/update_handler.go  # Hook into accounting updates
└── app/app.go                  # Initialize Redis, start sync worker

web/src/
├── i18n/en-US.ts              # Add translations for usage pages
└── i18n/zh-CN.ts              # Add Chinese translations
```

---

## Phase 1: Foundation (Week 1) - Database & Redis Setup

### Task 1.1: Create Database Migration

**Files:**
- Create: `internal/migrations/20260327_usage_analytics.go`
- Create: `scripts/migrate_usage_analytics.sh`

- [ ] **Step 1: Write migration test**

```go
// internal/migrations/20260327_usage_analytics_test.go
package migrations

import (
    "testing"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func TestMigration_AddAcctStatusColumn(t *testing.T) {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }

    // Create radius_accounting table without acct_status
    db.Exec(`CREATE TABLE radius_accounting (
        radacctid INTEGER PRIMARY KEY,
        acct_stop_time DATETIME
    )`)

    // Run migration
    err = AddAcctStatusColumn(db)
    if err != nil {
        t.Fatalf("AddAcctStatusColumn failed: %v", err)
    }

    // Verify column exists
    var column string
    db.Raw(`SELECT sql FROM sqlite_master WHERE type='table' AND name='radius_accounting'`).Scan(&column)
    if !contains(column, "acct_status") {
        t.Error("acct_status column not added")
    }
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
        s[:len(substr)] == substr || contains(s[1:], substr))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/migrations/20260327_usage_analytics_test.go -v`
Expected: FAIL with "undefined: AddAcctStatusColumn"

- [ ] **Step 3: Implement migration**

```go
// internal/migrations/20260327_usage_analytics.go
package migrations

import (
    "gorm.io/gorm"
)

func AddAcctStatusColumn(db *gorm.DB) error {
    // Add acct_status column if not exists
    err := db.Exec(`ALTER TABLE radius_accounting ADD COLUMN IF NOT EXISTS acct_status VARCHAR(20)`).Error
    if err != nil {
        return err
    }

    // Backfill existing data with batch processing for large tables (Fix #11)
    err = BackfillAcctStatus(db)
    if err != nil {
        return err
    }

    // Create trigger for automatic status updates (PostgreSQL)
    err = db.Exec(`CREATE OR REPLACE FUNCTION update_acct_status()
    RETURNS TRIGGER AS $$
    BEGIN
        IF NEW.acct_stop_time IS NULL OR NEW.acct_stop_time = '0001-01-01 00:00:00' THEN
            NEW.acct_status := 'active';
        ELSE
            NEW.acct_status := 'stopped';
        END IF;
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql`).Error
    if err != nil {
        // Non-PostgreSQL database, skip trigger
        return nil
    }

    err = db.Exec(`CREATE TRIGGER IF NOT EXISTS trigger_update_acct_status
    BEFORE INSERT OR UPDATE ON radius_accounting
    FOR EACH ROW EXECUTE FUNCTION update_acct_status()`).Error

    return err
}

// BackfillAcctStatus backfills acct_status in batches for large tables (Fix #11)
func BackfillAcctStatus(db *gorm.DB) error {
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

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/migrations/20260327_usage_analytics_test.go -v`
Expected: PASS

- [ ] **Step 5: Create migration script**

```bash
#!/bin/bash
# scripts/migrate_usage_analytics.sh

echo "Running Usage Analytics migration..."

# Add acct_status column
psql -d toughradius -c "ALTER TABLE radius_accounting ADD COLUMN IF NOT EXISTS acct_status VARCHAR(20);"

# Backfill existing data
psql -d toughradius -c "UPDATE radius_accounting SET acct_status = CASE
    WHEN acct_stop_time IS NULL OR acct_stop_time = '0001-01-01 00:00:00' THEN 'active'
    ELSE 'stopped'
END;"

# Create trigger
psql -d toughradius -c "CREATE OR REPLACE FUNCTION update_acct_status()
RETURNS TRIGGER AS \$\$
BEGIN
    IF NEW.acct_stop_time IS NULL OR NEW.acct_stop_time = '0001-01-01 00:00:00' THEN
        NEW.acct_status := 'active';
    ELSE
        NEW.acct_status := 'stopped';
    END IF;
    RETURN NEW;
END;
\$\$ LANGUAGE plpgsql;"

psql -d toughradius -c "DROP TRIGGER IF EXISTS trigger_update_acct_status ON radius_accounting;"
psql -d toughradius -c "CREATE TRIGGER trigger_update_acct_status
BEFORE INSERT OR UPDATE ON radius_accounting
FOR EACH ROW EXECUTE FUNCTION update_acct_status();"

# Create indexes
psql -d toughradius -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_tenant_user_time
ON radius_accounting(tenant_id, username, acct_start_time DESC)
WHERE acct_status = 'stopped';"

psql -d toughradius -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_data_usage
ON radius_accounting(tenant_id, username, (acct_input_octets + acct_output_octets))
WHERE acct_status = 'stopped';"

psql -d toughradius -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_active_sessions
ON radius_accounting(tenant_id, username, acct_unique_id)
WHERE acct_status = 'active';"

echo "Migration complete!"
```

- [ ] **Step 6: Make script executable and commit**

```bash
chmod +x scripts/migrate_usage_analytics.sh
git add internal/migrations/20260327_usage_analytics.go scripts/migrate_usage_analytics.sh
git commit -m "feat: add database migration for usage analytics (acct_status column)"
```

---

### Task 1.2: Update Domain Model

**Files:**
- Modify: `internal/domain/radius.go:450-480`

- [ ] **Step 1: Write failing test**

```go
// internal/domain/radius_accounting_test.go
package domain

import (
    "testing"
    "time"
)

func TestRadiusAccounting_AcctStatus(t *testing.T) {
    acct := &RadiusAccounting{
        AcctStartTime: time.Now(),
        AcctStopTime:  time.Time{}, // Active session
    }

    acct.SetAcctStatus()
    if acct.AcctStatus != "active" {
        t.Errorf("Expected status 'active', got '%s'", acct.AcctStatus)
    }

    acct.AcctStopTime = time.Now()
    acct.SetAcctStatus()
    if acct.AcctStatus != "stopped" {
        t.Errorf("Expected status 'stopped', got '%s'", acct.AcctStatus)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/radius_accounting_test.go -v`
Expected: FAIL with "undefined: SetAcctStatus"

- [ ] **Step 3: Implement acct_status field and method**

```go
// internal/domain/radius.go - Add to RadiusAccounting struct around line 450

type RadiusAccounting struct {
    RadacctID        int64     `gorm:"column:radacctid;primaryKey"`
    AcctSessionID    string    `gorm:"column:acctsessionid"`
    AcctUniqueID     string    `gorm:"column:acctuniqueid"`
    UserName         string    `gorm:"column:username"`
    Realm            string    `gorm:"column:realm"`
    NASIPAddress     string    `gorm:"column:nasipaddress"`
    NASPortID        string    `gorm:"column:nasportid"`
    NASPortType      string    `gorm:"column:nasporttype"`
    AcctStartTime    time.Time `gorm:"column:acctstarttime"`
    AcctStopTime     time.Time `gorm:"column:acctstoptime"`
    AcctSessionTime  int64     `gorm:"column:acctsessiontime"`
    AcctAuthentic    string    `gorm:"column:acctauthentic"`
    ConnectInfoStart string    `gorm:"column:connectinfo_start"`
    ConnectInfoStop  string    `gorm:"column:connectinfo_stop"`
    AcctInputOctets  int64     `gorm:"column:acctinputoctets"`
    AcctOutputOctets int64     `gorm:"column:acctoutputoctets"
    CalledStationID  string    `gorm:"column:calledstationid"`
    CallingStationID string    `gorm:"column:callingstationid"`
    AcctTerminateCause string  `gorm:"column:acctterminatecause"`
    ServiceType      string    `gorm:"column:servicetype"`
    FramedProtocol   string    `gorm:"column:framedprotocol"`
    FramedIPAddress  string    `gorm:"column:framedipaddress"`
    AcctStartDelay   int64     `gorm:"column:acctstartdelay"`
    AcctTerminateCause string  `gorm:"column:acctterminatecause"`
    TenantID         int64     `gorm:"column:tenant_id"`
    AcctStatus       string    `gorm:"column:acct_status"` // NEW FIELD
}

// SetAcctStatus sets acct_status based on acct_stop_time
func (ra *RadiusAccounting) SetAcctStatus() {
    if ra.AcctStopTime.IsZero() || ra.AcctStopTime.Year() == 1 {
        ra.AcctStatus = "active"
    } else {
        ra.AcctStatus = "stopped"
    }
}

// TableName specifies the table name for GORM
func (RadiusAccounting) TableName() string {
    return "radius_accounting"
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/radius_accounting_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/domain/radius.go internal/domain/radius_accounting_test.go
git commit -m "feat: add acct_status field to RadiusAccounting model"
```

---

### Task 1.3: Set Up Redis Client in App

**Files:**
- Modify: `internal/app/app.go:50-80`
- Create: `internal/app/redis.go`

- [ ] **Step 1: Write failing test**

```go
// internal/app/redis_test.go
package app

import (
    "context"
    "testing"
    "github.com/redis/go-redis/v9"
)

func TestNewRedisClient(t *testing.T) {
    client, err := NewRedisClient(&RedisConfig{
        Addr:     "localhost:6379",
        PoolSize: 10,
    })

    if err != nil {
        t.Fatalf("NewRedisClient failed: %v", err)
    }

    if client == nil {
        t.Fatal("Expected non-nil client")
    }

    // Test connection
    ctx := context.Background()
    err = client.Ping(ctx).Err()
    if err != nil {
        t.Logf("Redis not available (expected in CI): %v", err)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/app/redis_test.go -v`
Expected: FAIL with "undefined: NewRedisClient"

- [ ] **Step 3: Implement Redis client**

```go
// internal/app/redis.go
package app

import (
    "time"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

type RedisConfig struct {
    Addr         string
    Password     string
    PoolSize     int
    MinIdleConns int
    MaxRetries   int
}

func NewRedisClient(config *RedisConfig) (*redis.Client, error) {
    if config.PoolSize == 0 {
        config.PoolSize = 2000 // Increased from 500 for 100K concurrent users (Fix #9)
    }
    if config.MinIdleConns == 0 {
        config.MinIdleConns = 100 // Increased from 50 (Fix #9)
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

    // Test connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        zap.L().Error("Redis connection failed", zap.Error(err))
        return nil, err
    }

    zap.L().Info("Redis connected", zap.String("addr", config.Addr))
    return client, nil
}
```

- [ ] **Step 4: Integrate into app.go**

```go
// internal/app/app.go - Add to App struct around line 50

type App struct {
    Config      *config.Config
    DB          *gorm.DB
    RADIUS      *radiusd.Server
    Redis       *redis.Client  // NEW FIELD
    // ... other fields
}

// InitializeRedis - call this in app initialization
func (app *App) InitializeRedis() error {
    redisConfig := &app.RedisConfig{
        Addr:         app.Config.RedisAddr,
        Password:     app.Config.RedisPassword,
        PoolSize:     500,
        MinIdleConns: 50,
        MaxRetries:   3,
    }

    client, err := NewRedisClient(redisConfig)
    if err != nil {
        return err
    }

    app.Redis = client
    return nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/app/redis_test.go -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/app/redis.go internal/app/app.go internal/app/redis_test.go
git commit -m "feat: add Redis client with enterprise configuration"
```

---

## Phase 2: Core Usage Tracking (Week 2)

### Task 2.1: Implement UsageTracker with Write-Through Caching

**Files:**
- Create: `internal/usage/tracker.go`
- Create: `internal/usage/tracker_test.go`

- [ ] **Step 1: Write failing test for Redis key generation**

```go
// internal/usage/tracker_test.go
package usage

import (
    "testing"
)

func TestUsageTracker_usageKey(t *testing.T) {
    tracker := &UsageTracker{}

    key := tracker.usageKey(123, "user456")
    expected := "user:usage:123:user456"

    if key != expected {
        t.Errorf("Expected key '%s', got '%s'", expected, key)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/usage/tracker_test.go -v`
Expected: FAIL with "undefined: usageKey"

- [ ] **Step 3: Implement UsageTracker struct and methods**

```go
// internal/usage/tracker.go
package usage

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"
    "time"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type UsageTracker struct {
    db    *gorm.DB
    redis *redis.Client
}

type UserUsage struct {
    TenantID           int64 `json:"tenant_id"`
    UserID             int64 `json:"user_id"`
    Username           string `json:"username"`
    TimeQuotaTotal     int64 `json:"time_quota_total"`
    TimeQuotaUsed      int64 `json:"time_quota_used"`
    TimeQuotaRemaining int64 `json:"time_quota_remaining"`
    DataQuotaTotal     int64 `json:"data_quota_total"`
    DataQuotaUsed      int64 `json:"data_quota_used"`
    DataQuotaRemaining int64 `json:"data_quota_remaining"`
}

func NewUsageTracker(db *gorm.DB, redis *redis.Client) *UsageTracker {
    return &UsageTracker{
        db:    db,
        redis: redis,
    }
}

func (ut *UsageTracker) usageKey(tenantID int64, username string) string {
    return fmt.Sprintf("user:usage:%d:%s", tenantID, username)
}

// GetUserUsage retrieves user usage with multi-level caching
func (ut *UsageTracker) GetUserUsage(ctx context.Context, tenantID int64, username string) (*UserUsage, error) {
    usageKey := ut.usageKey(tenantID, username)

    // Try Redis hash first (L2 cache) - Fix #5: Use Hash pattern
    vals, err := ut.redis.HGetAll(ctx, usageKey).Result()
    if err == nil && len(vals) > 0 {
        // Parse hash fields
        timeUsed, _ := strconv.ParseInt(vals["time_quota_used"], 10, 64)
        dataUsed, _ := strconv.ParseInt(vals["data_quota_used"], 10, 64)
        timeTotal, _ := strconv.ParseInt(vals["time_quota_total"], 10, 64)
        dataTotal, _ := strconv.ParseInt(vals["data_quota_total"], 10, 64)

        zap.L().Debug("Usage cache hit (Redis)", zap.String("username", username))
        return &UserUsage{
            TenantID:           tenantID,
            Username:           username,
            TimeQuotaTotal:     timeTotal,
            TimeQuotaUsed:      timeUsed,
            TimeQuotaRemaining: timeTotal - timeUsed,
            DataQuotaTotal:     dataTotal,
            DataQuotaUsed:      dataUsed,
            DataQuotaRemaining: dataTotal - dataUsed,
        }, nil
    }

    // Cache miss - calculate from PostgreSQL
    zap.L().Debug("Usage cache miss, calculating from DB", zap.String("username", username))

    var user struct {
        TimeQuota int64 `gorm:"column:time_quota"`
        DataQuota int64 `gorm:"column:data_quota"`
    }

    err = ut.db.Table("radius_user").
        Select("time_quota, data_quota").
        Where("tenant_id = ? AND username = ?", tenantID, username).
        Scan(&user).Error
    if err != nil {
        return nil, err
    }

    // Fix #10: Combine N+1 queries into single query
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

    usage := &UserUsage{
        TenantID:           tenantID,
        Username:           username,
        TimeQuotaTotal:     user.TimeQuota,
        TimeQuotaUsed:      result.TimeUsed,
        TimeQuotaRemaining: user.TimeQuota - result.TimeUsed,
        DataQuotaTotal:     user.DataQuota,
        DataQuotaUsed:      result.DataUsed,
        DataQuotaRemaining: user.DataQuota - result.DataUsed,
    }

    // Fix #5: Store as hash in Redis (not JSON)
    pipe := ut.redis.Pipeline()
    pipe.HSet(ctx, usageKey, map[string]interface{}{
        "tenant_id":           usage.TenantID,
        "username":             usage.Username,
        "time_quota_total":     usage.TimeQuotaTotal,
        "time_quota_used":      usage.TimeQuotaUsed,
        "time_quota_remaining": usage.TimeQuotaRemaining,
        "data_quota_total":     usage.DataQuotaTotal,
        "data_quota_used":      usage.DataQuotaUsed,
        "data_quota_remaining": usage.DataQuotaRemaining,
    })
    pipe.Expire(ctx, usageKey, 60*time.Second)
    pipe.Exec(ctx)

    return usage, nil
}

// RecordAccountingUpdate updates Redis immediately when accounting data arrives (write-through)
func (ut *UsageTracker) RecordAccountingUpdate(ctx context.Context, record *RadiusAccounting) error {
    if record.AcctStatus != "stopped" {
        return nil // Only count stopped sessions
    }

    usageKey := ut.usageKey(record.TenantID, record.UserName)

    // Atomic increment operations in Redis
    pipe := ut.redis.Pipeline()
    pipe.HIncrBy(ctx, usageKey, "time_quota_used", record.AcctSessionTime)
    pipe.HIncrBy(ctx, usageKey, "data_quota_used", record.AcctInputOctets+record.AcctOutputOctets)
    pipe.Expire(ctx, usageKey, 60*time.Second)

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        zap.L().Error("Failed to update Redis usage",
            zap.String("username", record.UserName),
            zap.Error(err))
        return err
    }

    zap.L().Debug("Recorded accounting update to Redis",
        zap.String("username", record.UserName),
        zap.Int64("session_time", record.AcctSessionTime))

    return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/usage/tracker_test.go -v`
Expected: PASS

- [ ] **Step 5: Write integration test for GetUserUsage**

```go
// Add to internal/usage/tracker_test.go

func TestUsageTracker_GetUserUsage_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // This test requires PostgreSQL and Redis
    // Setup test database and Redis connection
    // ...
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/usage/tracker.go internal/usage/tracker_test.go
git commit -m "feat: implement UsageTracker with Redis write-through caching"
```

---

### Task 2.2: Integrate UsageTracker into Accounting Handler

**Files:**
- Modify: `internal/radiusd/plugins/accounting/handlers/update_handler.go:50-80`

- [ ] **Step 1: Write test for accounting update hook**

```go
// internal/radiusd/plugins/accounting/handlers/update_handler_test.go
package handlers

import (
    "testing"
    "github.com/talkincode/toughradius/v9/internal/usage"
)

func TestUpdateHandler_CallsUsageTracker(t *testing.T) {
    // Mock the usage tracker
    mockTracker := &MockUsageTracker{}
    handler := &UpdateHandler{
        UsageTracker: mockTracker,
    }

    // Simulate accounting update
    record := &RadiusAccounting{
        AcctStatus:      "stopped",
        AcctSessionTime: 3600,
        TenantID:        1,
        UserName:        "testuser",
    }

    err := handler.Handle(record)
    if err != nil {
        t.Fatalf("Handle failed: %v", err)
    }

    if !mockTracker.Called {
        t.Error("UsageTracker.RecordAccountingUpdate was not called")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/radiusd/plugins/accounting/handlers/update_handler_test.go -v`
Expected: FAIL with "undefined: MockUsageTracker"

- [ ] **Step 3: Create mock and integrate**

```go
// internal/usage/mock_test.go
package usage

import "context"

type MockUsageTracker struct {
    Called bool
}

func (m *MockUsageTracker) RecordAccountingUpdate(ctx context.Context, record *RadiusAccounting) error {
    m.Called = true
    return nil
}
```

```go
// internal/radiusd/plugins/accounting/handlers/update_handler.go - Modify existing file

package handlers

import (
    "context"
    "github.com/talkincode/toughradius/v9/internal/usage"
)

type UpdateHandler struct {
    AccountingRepo AccountingRepository
    UsageTracker   *usage.UsageTracker  // NEW FIELD
}

func (h *UpdateHandler) Handle(ctx context.Context, record *RadiusAccounting) error {
    // Existing code: save to PostgreSQL
    if err := h.AccountingRepo.Update(ctx, record); err != nil {
        return err
    }

    // NEW: Immediately update Redis (write-through)
    if h.UsageTracker != nil {
        if err := h.UsageTracker.RecordAccountingUpdate(ctx, record); err != nil {
            // Log error but don't fail the request (Redis is cache, not source of truth)
            zap.L().Warn("Failed to update usage tracker", zap.Error(err))
        }
    }

    return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/radiusd/plugins/accounting/handlers/update_handler_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/radiusd/plugins/accounting/handlers/update_handler.go internal/usage/mock_test.go
git commit -m "feat: integrate UsageTracker into accounting update handler"
```

---

### Task 2.3: Implement Multi-Level Cache

**Files:**
- Create: `internal/usage/cache.go`
- Create: `internal/usage/cache_test.go`

- [ ] **Step 1: Write failing test for L1 cache**

```go
// internal/usage/cache_test.go
package usage

import (
    "context"
    "testing"
    "time"
)

func TestSessionCache_Get_L1Hit(t *testing.T) {
    cache := NewSessionCache(nil, nil)

    // Store in L1
    usage := &UserUsage{Username: "test", TimeQuotaUsed: 100}
    cache.l1Cache.Store("test", usage)

    // Retrieve
    result, err := cache.Get(context.Background(), "test", 1, "test")
    if err != nil {
        t.Fatalf("Get failed: %v", err)
    }

    if result.Username != "test" {
        t.Errorf("Expected username 'test', got '%s'", result.Username)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/usage/cache_test.go -v`
Expected: FAIL with "undefined: NewSessionCache"

- [ ] **Step 3: Implement multi-level cache**

```go
// internal/usage/cache.go
package usage

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/hashicorp/golang-lru/v2"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

// Fix #2: Replace sync.Map with bounded LRU cache
type SessionCache struct {
    l1Cache *lru.Cache[string, *UserUsage]  // Bounded LRU (max 10K entries, ~10MB memory)
    l2Cache *redis.Client                   // Redis cache (warm data, <10ms access)
    db      *gorm.DB                        // PostgreSQL (cold data, <100ms access)
}

func NewSessionCache(l2Cache *redis.Client, db *gorm.DB) *SessionCache {
    // Create LRU cache with max 10,000 entries (~10MB memory)
    cache, _ := lru.NewEvicted[string, *UserUsage](10000)

    return &SessionCache{
        l1Cache: cache,
        l2Cache: l2Cache,
        db:      db,
    }
}

func (sc *SessionCache) Get(ctx context.Context, username string, tenantID int64) (*UserUsage, error) {
    // Try L1 first (<1ms)
    if val, ok := sc.l1Cache.Get(username); ok {
        return val, nil
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
    // This would call UsageTracker.GetUserUsage
    return nil, fmt.Errorf("not found in cache")
}

func (sc *SessionCache) Invalidate(username string) {
    sc.l1Cache.Remove(username)
    // Note: Redis keys expire automatically via TTL
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/usage/cache_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/usage/cache.go internal/usage/cache_test.go
git commit -m "feat: implement multi-level cache (L1 in-memory + L2 Redis)"
```

---

## Phase 0: Pre-Implementation Security Foundation (Before Week 1)

> **Critical:** These security fixes MUST be implemented before any API endpoints (Fixes #3, #4, #7, #14)

### Task 0.1: Implement JWT Authentication Middleware (Fix #3)

**Files:**
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/auth_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/middleware/auth_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
)

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    middleware := JWTAuthMiddleware("test-secret")
    err := middleware(func(c echo.Context) error {
        return c.String(http.StatusOK, "OK")
    })(c)

    if err != nil {
        t.Fatalf("Middleware failed: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/auth_test.go -v`
Expected: FAIL with "undefined: JWTAuthMiddleware"

- [ ] **Step 3: Implement JWT authentication**

```go
// internal/middleware/auth.go
package middleware

import (
    "fmt"
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

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/auth_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/auth.go internal/middleware/auth_test.go
git commit -m "feat: implement JWT authentication middleware (Fix #3)"
```

---

### Task 0.2: Implement RBAC Permission System (Fix #4)

**Files:**
- Create: `internal/middleware/rbac.go`
- Modify: `internal/middleware/security.go` (enhance existing)

- [ ] **Step 1: Write failing test**

```go
// internal/middleware/rbac_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
)

func TestRequirePermission_OperatorCanReadTenant(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    c.Set("security", SecurityContext{
        UserID:   123,
        TenantID: 1,
        Role:     "operator",
        Username: "testuser",
    })

    middleware := requirePermission("users:read:tenant")
    err := middleware(func(c echo.Context) error {
        return c.String(http.StatusOK, "OK")
    })(c)

    if err != nil {
        t.Fatalf("Middleware failed: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/rbac_test.go -v`
Expected: FAIL with "undefined: requirePermission"

- [ ] **Step 3: Implement RBAC system**

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

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/rbac_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/rbac.go internal/middleware/rbac_test.go
git commit -m "feat: implement RBAC permission system (Fix #4)"
```

---

### Task 0.3: Implement Input Validation Package (Fix #7)

**Files:**
- Create: `internal/validator/usage_validator.go`
- Create: `internal/validator/usage_validator_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/validator/usage_validator_test.go
package validator

import (
    "testing"
    "time"
)

func TestValidateUsageHistoryRequest_ValidRange(t *testing.T) {
    from := time.Now().AddDate(0, 0, -7).Format(time.RFC3339)
    to := time.Now().Format(time.RFC3339)

    fromDate, toDate, err := ValidateUsageHistoryRequest(from, to, "daily")
    if err != nil {
        t.Fatalf("ValidateUsageHistoryRequest failed: %v", err)
    }

    if toDate.Before(fromDate) {
        t.Error("toDate should be after fromDate")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/validator/usage_validator_test.go -v`
Expected: FAIL with "undefined: ValidateUsageHistoryRequest"

- [ ] **Step 3: Implement input validation**

```go
// internal/validator/usage_validator.go
package validator

import (
    "fmt"
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

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/validator/usage_validator_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/validator/usage_validator.go internal/validator/usage_validator_test.go
git commit -m "feat: implement input validation package (Fix #7)"
```

---

### Task 0.4: Implement Security Headers Middleware (Fix #14)

**Files:**
- Create: `internal/middleware/security_headers.go`
- Create: `internal/middleware/security_headers_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/middleware/security_headers_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
)

func TestSecurityHeaders_SetsHeaders(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()

    e.Use(SecurityHeaders())
    e.GET("/", func(c echo.Context) error {
        return c.String(http.StatusOK, "OK")
    })

    e.ServeHTTP(rec, req)

    headers := rec.Header()
    if headers.Get("X-Content-Type-Options") != "nosniff" {
        t.Error("X-Content-Type-Options header not set")
    }

    if headers.Get("X-Frame-Options") != "DENY" {
        t.Error("X-Frame-Options header not set")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/security_headers_test.go -v`
Expected: FAIL with "undefined: SecurityHeaders"

- [ ] **Step 3: Implement security headers**

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

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/security_headers_test.go -v`
Expected: PASS

- [ ] **Step 5: Apply globally in app.go**

```go
// internal/app/app.go - Add to Echo initialization

e.Use(middleware.SecurityHeaders())
```

- [ ] **Step 6: Commit**

```bash
git add internal/middleware/security_headers.go internal/middleware/security_headers_test.go
git commit -m "feat: implement security headers middleware (Fix #14)"
```

---

## Phase 3: API Endpoints (Week 3)

### Task 3.1: Create Usage API Endpoints

**Files:**
- Create: `internal/adminapi/usage_api.go`
- Create: `internal/adminapi/usage_api_test.go`

- [ ] **Step 1: Write failing test for /users/me/usage endpoint**

```go
// internal/adminapi/usage_api_test.go
package adminapi

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
)

func TestGetUserUsage_Success(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/usage", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // Mock security context
    c.Set("security", SecurityContext{
        UserID:   123,
        TenantID: 1,
        Role:     "user",
        Username: "testuser",
    })

    handler := &UsageAPIHandler{}
    err := handler.GetUserUsage(c)

    if err != nil {
        t.Fatalf("GetUserUsage failed: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adminapi/usage_api_test.go -v`
Expected: FAIL with "undefined: UsageAPIHandler"

- [ ] **Step 3: Implement usage API endpoints**

```go
// internal/adminapi/usage_api.go
package adminapi

import (
    "net/http"
    "strconv"
    "time"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
    "github.com/talkincode/toughradius/v9/internal/usage"
)

type UsageAPIHandler struct {
    usageTracker *usage.UsageTracker
    cache        *usage.SessionCache
}

type UserUsageResponse struct {
    TimeQuotaTotal     int64  `json:"time_quota_total"`
    TimeQuotaUsed      int64  `json:"time_quota_used"`
    TimeQuotaRemaining int64  `json:"time_quota_remaining"`
    DataQuotaTotal     int64  `json:"data_quota_total"`
    DataQuotaUsed      int64  `json:"data_quota_used"`
    DataQuotaRemaining int64  `json:"data_quota_remaining"`
    SessionsToday      int    `json:"sessions_today"`
    Prediction         *UsagePrediction `json:"prediction,omitempty"`
}

type UsagePrediction struct {
    QuotaExpiresAt string `json:"quota_expires_at"`
    Confidence     string `json:"confidence"`
    Message        string `json:"message"`
}

// GET /api/v1/users/me/usage
func (h *UsageAPIHandler) GetUserUsage(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    ctx := c.Request().Context()
    usageData, err := h.usageTracker.GetUserUsage(ctx, secCtx.TenantID, secCtx.Username)
    if err != nil {
        zap.L().Error("Failed to get user usage", zap.Error(err))
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve usage data",
        })
    }

    // Count today's sessions
    sessionsToday := h.countTodaySessions(ctx, secCtx.TenantID, secCtx.Username)

    response := &UserUsageResponse{
        TimeQuotaTotal:     usageData.TimeQuotaTotal,
        TimeQuotaUsed:      usageData.TimeQuotaUsed,
        TimeQuotaRemaining: usageData.TimeQuotaRemaining,
        DataQuotaTotal:     usageData.DataQuotaTotal,
        DataQuotaUsed:      usageData.DataQuotaUsed,
        DataQuotaRemaining: usageData.DataQuotaRemaining,
        SessionsToday:      sessionsToday,
    }

    // Add prediction if quota is limited
    if usageData.TimeQuotaTotal > 0 && usageData.TimeQuotaRemaining > 0 {
        response.Prediction = h.predictQuotaExpiry(usageData)
    }

    return c.JSON(http.StatusOK, response)
}

// GET /api/v1/users/:id/usage (admin/operator only)
func (h *UsageAPIHandler) GetUserUsageByID(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    // Check permissions
    if !hasPermission(secCtx, "users:read:tenant") {
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "Insufficient permissions",
        })
    }

    userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid user ID",
        })
    }

    // Fetch user and tenant
    var user struct {
        TenantID int64  `gorm:"column:tenant_id"`
        Username string `gorm:"column:username"`
    }

    err = h.db.Table("radius_user").
        Select("tenant_id, username").
        Where("id = ?", userID).
        Scan(&user).Error
    if err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "User not found",
        })
    }

    // Check tenant isolation
    if secCtx.Role != "platform_admin" && user.TenantID != secCtx.TenantID {
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "Cross-tenant access denied",
        })
    }

    ctx := c.Request().Context()
    usageData, err := h.usageTracker.GetUserUsage(ctx, user.TenantID, user.Username)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve usage data",
        })
    }

    return c.JSON(http.StatusOK, usageData)
}

// GET /api/v1/users/me/usage/history
func (h *UsageAPIHandler) GetUsageHistory(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    from := c.QueryParam("from")
    to := c.QueryParam("to")
    granularity := c.QueryParam("granularity")
    if granularity == "" {
        granularity = "daily"
    }

    // Validate dates
    fromDate, err := time.Parse(time.RFC3339, from)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid 'from' date format",
        })
    }

    toDate, err := time.Parse(time.RFC3339, to)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid 'to' date format",
        })
    }

    ctx := c.Request().Context()
    history, err := h.getHistoricalUsage(ctx, secCtx.TenantID, secCtx.Username, fromDate, toDate, granularity)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve historical data",
        })
    }

    return c.JSON(http.StatusOK, history)
}

func (h *UsageAPIHandler) countTodaySessions(ctx context.Context, tenantID int64, username string) int {
    today := time.Now().Truncate(24 * time.Hour)
    var count int64

    h.db.Table("radius_accounting").
        Where("tenant_id = ? AND username = ? AND acct_start_time >= ?", tenantID, username, today).
        Count(&count)

    return int(count)
}

func (h *UsageAPIHandler) predictQuotaExpiry(usage *usage.UserUsage) *UsagePrediction {
    if usage.TimeQuotaUsed == 0 || usage.TimeQuotaTotal == 0 {
        return &UsagePrediction{
            Confidence: "low",
            Message:    "Not enough data to predict",
        }
    }

    // Calculate days elapsed since quota started
    // This assumes quota was assigned when user was created
    var userCreatedAt time.Time
    h.db.Table("radius_user").
        Select("created_at").
        Where("tenant_id = ? AND username = ?", usage.TenantID, usage.Username).
        Scan(&userCreatedAt)

    if userCreatedAt.IsZero() {
        userCreatedAt = time.Now().AddDate(0, 0, -30) // Fallback: assume 30 days ago
    }

    daysElapsed := time.Since(userCreatedAt).Hours() / 24
    if daysElapsed < 1 {
        daysElapsed = 1
    }

    // Calculate daily usage rate
    dailyUsage := float64(usage.TimeQuotaUsed) / daysElapsed

    // Predict days remaining
    if dailyUsage == 0 {
        return &UsagePrediction{
            Confidence: "low",
            Message:    "Usage rate too low to predict",
        }
    }

    daysRemaining := float64(usage.TimeQuotaRemaining) / dailyUsage
    expiryDate := time.Now().AddDate(0, 0, int(daysRemaining))

    // Set confidence based on data quality
    confidence := "low"
    if daysElapsed >= 7 {
        confidence = "high"
    } else if daysElapsed >= 3 {
        confidence = "medium"
    }

    return &UsagePrediction{
        QuotaExpiresAt: expiryDate.Format(time.RFC3339),
        Confidence:     confidence,
        Message:        fmt.Sprintf("%.1f days remaining (based on %.0f days of usage)", daysRemaining, daysElapsed),
    }
}
```
func (h *UsageAPIHandler) getHistoricalUsage(ctx context.Context, tenantID int64, username string, from, to time.Time, granularity string) ([]UsageHistoryPoint, error) {
    // Implementation queries radius_accounting grouped by hour/day/week
    // This will use the indexes we created
    return []UsageHistoryPoint{}, nil
}

type UsageHistoryPoint struct {
    Timestamp string `json:"timestamp"`
    Seconds   int64  `json:"seconds"`
    Bytes     int64  `json:"bytes"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adminapi/usage_api_test.go -v`
Expected: PASS

- [ ] **Step 5: Register routes in adminapi.go**

```go
// internal/adminapi/adminapi.go - Add to route registration

func RegisterUsageRoutes(e *echo.Echo, handler *UsageAPIHandler) {
    api := e.Group("/api/v1")

    // User endpoints (self-access)
    api.GET("/users/me/usage", handler.GetUserUsage, requirePermission("users:read:own"))
    api.GET("/users/me/usage/history", handler.GetUsageHistory, requirePermission("users:read:own"))

    // Admin endpoints
    api.GET("/users/:id/usage", handler.GetUserUsageByID, requirePermission("users:read:tenant"))
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/adminapi/usage_api.go internal/adminapi/usage_api_test.go internal/adminapi/adminapi.go
git commit -m "feat: add usage analytics API endpoints"
```

---

### Task 3.2: Add Active Sessions Endpoint

**Files:**
- Modify: `internal/adminapi/usage_api.go`
- Modify: `internal/adminapi/usage_api_test.go`

- [ ] **Step 1: Write failing test for /sessions/active endpoint**

```go
// Add to internal/adminapi/usage_api_test.go

func TestGetActiveSessions_Success(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/active", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    c.Set("security", SecurityContext{
        UserID:   123,
        TenantID: 1,
        Role:     "operator",
    })

    handler := &UsageAPIHandler{}
    err := handler.GetActiveSessions(c)

    if err != nil {
        t.Fatalf("GetActiveSessions failed: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adminapi/usage_api_test.go -v -run TestGetActiveSessions`
Expected: FAIL with "undefined: GetActiveSessions"

- [ ] **Step 3: Implement GetActiveSessions endpoint**

```go
// Add to internal/adminapi/usage_api.go

type ActiveSession struct {
    Username            string `json:"username"`
    StartTime           string `json:"start_time"`
    DurationSeconds     int64  `json:"duration_seconds"`
    NASIP               string `json:"nas_ip"`
    FramedIP            string `json:"framed_ip"`
    DataUsedThisSession int64  `json:"data_used_this_session"`
}

type ActiveSessionsResponse struct {
    Total    int            `json:"total"`
    Sessions []ActiveSession `json:"sessions"`
}

// GET /api/v1/sessions/active
func (h *UsageAPIHandler) GetActiveSessions(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    ctx := c.Request().Context()

    // Query active sessions from radius_online
    var sessions []ActiveSession

    query := h.db.Table("radius_online").
        Select(`username,
                acct_start_time as start_time,
                EXTRACT(EPOCH FROM (NOW() - acct_start_time))::BIGINT as duration_seconds,
                nas_ip_address as nas_ip,
                framed_ip_address as framed_ip,
                (acct_input_octets + acct_output_octets) as data_used_this_session`)

    // Apply tenant isolation
    if secCtx.Role != "platform_admin" {
        query = query.Where("tenant_id = ?", secCtx.TenantID)
    }

    err := query.Scan(&sessions).Error
    if err != nil {
        zap.L().Error("Failed to query active sessions", zap.Error(err))
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve active sessions",
        })
    }

    // Also check radius_accounting for any sessions not yet in radius_online
    var accountingSessions []ActiveSession
    err = h.db.Table("radius_accounting").
        Select(`username,
                acct_start_time as start_time,
                acct_session_time as duration_seconds,
                nas_ip_address as nas_ip,
                framed_ip_address as framed_ip,
                (acct_input_octets + acct_output_octets) as data_used_this_session`).
        Where("acct_status = ?", "active").
        ApplyTenantFilter(secCtx).
        Scan(&accountingSessions).Error

    if err == nil {
        sessions = append(sessions, accountingSessions...)
    }

    response := &ActiveSessionsResponse{
        Total:    len(sessions),
        Sessions: sessions,
    }

    return c.JSON(http.StatusOK, response)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adminapi/usage_api_test.go -v -run TestGetActiveSessions`
Expected: PASS

- [ ] **Step 5: Register route**

```go
// Add to internal/adminapi/adminapi.go route registration

api.GET("/sessions/active", handler.GetActiveSessions, requirePermission("sessions:read:tenant"))
```

- [ ] **Step 6: Commit**

```bash
git add internal/adminapi/usage_api.go internal/adminapi/usage_api_test.go internal/adminapi/adminapi.go
git commit -m "feat: add active sessions API endpoint"
```

---

### Task 3.3: Add Usage Insights Endpoint

**Files:**
- Modify: `internal/adminapi/usage_api.go`
- Modify: `internal/adminapi/usage_api_test.go`

- [ ] **Step 1: Write failing test**

```go
// Add to internal/adminapi/usage_api_test.go

func TestGetUsageInsights_Success(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/usage/insights", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    c.Set("security", SecurityContext{
        UserID:   123,
        TenantID: 1,
        Role:     "user",
        Username: "testuser",
    })

    handler := &UsageAPIHandler{
        enhancedEngine: enhancedPredictiveEngine,
    }
    err := handler.GetUsageInsights(c)

    if err != nil {
        t.Fatalf("GetUsageInsights failed: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adminapi/usage_api_test.go -v -run TestGetUsageInsights`
Expected: FAIL with "undefined: GetUsageInsights"

- [ ] **Step 3: Implement insights endpoint**

```go
// Add to internal/adminapi/usage_api.go

type UsageInsightsResponse struct {
    Patterns  *UsagePattern  `json:"patterns"`
    Anomalies []Anomaly      `json:"anomalies"`
    Prediction *UsagePrediction `json:"prediction"`
}

// GET /api/v1/users/me/usage/insights
func (h *UsageAPIHandler) GetUsageInsights(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    ctx := c.Request().Context()

    // Detect patterns (last 30 days)
    patterns, err := h.enhancedEngine.DetectPatterns(secCtx.Username, 30)
    if err != nil {
        zap.L().Warn("Failed to detect patterns", zap.Error(err))
        patterns = &UsagePattern{} // Return empty patterns
    }

    // Detect anomalies
    anomalies, err := h.enhancedEngine.DetectAnomalies(secCtx.Username, 30)
    if err != nil {
        zap.L().Warn("Failed to detect anomalies", zap.Error(err))
        anomalies = []Anomaly{} // Return empty anomalies
    }

    // Get current usage for prediction
    usage, err := h.usageTracker.GetUserUsage(ctx, secCtx.TenantID, secCtx.Username)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve usage data",
        })
    }

    prediction := h.predictQuotaExpiry(usage)

    response := &UsageInsightsResponse{
        Patterns:   patterns,
        Anomalies:  anomalies,
        Prediction: prediction,
    }

    return c.JSON(http.StatusOK, response)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adminapi/usage_api_test.go -v -run TestGetUsageInsights`
Expected: PASS

- [ ] **Step 5: Register route**

```go
// Add to internal/adminapi/adminapi.go route registration

api.GET("/users/me/usage/insights",
    handler.GetUsageInsights,
    rateLimiter.Middleware("/users/me/usage/insights", 1, 3),
    requirePermission("users:read:own"))
```

- [ ] **Step 6: Commit**

```bash
git add internal/adminapi/usage_api.go internal/adminapi/usage_api_test.go internal/adminapi/adminapi.go
git commit -m "feat: add usage insights API endpoint"
```

---

### Task 3.4: Implement getHistoricalUsage Query

**Files:**
- Modify: `internal/adminapi/usage_api.go`

- [ ] **Step 1: Write failing test**

```go
// Add to internal/adminapi/usage_api_test.go

func TestGetHistoricalUsage_Daily(t *testing.T) {
    handler := &UsageAPIHandler{
        db: setupTestDB(),
    }

    ctx := context.Background()
    from := time.Now().AddDate(0, 0, -7)
    to := time.Now()

    history, err := handler.getHistoricalUsage(ctx, 1, "testuser", from, to, "daily")
    if err != nil {
        t.Fatalf("getHistoricalUsage failed: %v", err)
    }

    if len(history) == 0 {
        t.Error("Expected historical data, got empty slice")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adminapi/usage_api_test.go -v -run TestGetHistoricalUsage`
Expected: FAIL with "implementation returns empty slice"

- [ ] **Step 3: Implement getHistoricalUsage**

```go
// Replace the stub in internal/adminapi/usage_api.go

func (h *UsageAPIHandler) getHistoricalUsage(ctx context.Context, tenantID int64, username string, from, to time.Time, granularity string) ([]UsageHistoryPoint, error) {
    var timeGrouping string

    switch granularity {
    case "hourly":
        timeGrouping = "DATE_TRUNC('hour', acct_start_time)"
    case "daily":
        timeGrouping = "DATE_TRUNC('day', acct_start_time)"
    case "weekly":
        timeGrouping = "DATE_TRUNC('week', acct_start_time)"
    default:
        return nil, fmt.Errorf("invalid granularity: %s", granularity)
    }

    var results []struct {
        Timestamp time.Time `gorm:"column:timestamp"`
        Seconds   int64     `gorm:"column:seconds"`
        Bytes     int64     `gorm:"column:bytes"`
    }

    err := h.db.Table("radius_accounting").
        Select(fmt.Sprintf(`%s as timestamp,
                SUM(acct_session_time) as seconds,
                SUM(acct_input_octets + acct_output_octets) as bytes`, timeGrouping)).
        Where("tenant_id = ? AND username = ? AND acct_start_time >= ? AND acct_start_time <= ? AND acct_status = ?",
            tenantID, username, from, to, "stopped").
        Group(timeGrouping).
        Order("timestamp").
        Scan(&results).Error

    if err != nil {
        return nil, err
    }

    history := make([]UsageHistoryPoint, len(results))
    for i, r := range results {
        history[i] = UsageHistoryPoint{
            Timestamp: r.Timestamp.Format(time.RFC3339),
            Seconds:   r.Seconds,
            Bytes:     r.Bytes,
        }
    }

    return history, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adminapi/usage_api_test.go -v -run TestGetHistoricalUsage`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adminapi/usage_api.go
git commit -m "feat: implement getHistoricalUsage query with granularity support"
```

---

## Phase 4: WebSocket Real-Time Updates (Week 4)

### Task 4.1: Implement WebSocket Hub

**Files:**
- Create: `internal/adminapi/websocket_hub.go`
- Create: `internal/adminapi/websocket_hub_test.go`

- [ ] **Step 1: Write failing test for WebSocket hub**

```go
// internal/adminapi/websocket_hub_test.go
package adminapi

import (
    "testing"
    "time"
)

func TestWebSocketHub_Broadcast(t *testing.T) {
    hub := NewWebSocketHub()

    go hub.Run()
    defer hub.Stop()

    // Create a mock client
    client := &MockClient{hub: hub}
    hub.Register(client)

    // Broadcast message
    hub.Broadcast([]byte(`{"username":"test","time_used":100}`))

    time.Sleep(100 * time.Millisecond)

    if !client.Received {
        t.Error("Client did not receive broadcast")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adminapi/websocket_hub_test.go -v`
Expected: FAIL with "undefined: NewWebSocketHub"

- [ ] **Step 3: Implement WebSocket hub**

```go
// internal/adminapi/websocket_hub.go
package adminapi

import (
    "sync"
    "github.com/gorilla/websocket"
    "go.uber.org/zap"
)

type WebSocketHub struct {
    clients    map[*WebSocketClient]bool
    broadcast  chan []byte
    register   chan *WebSocketClient
    unregister chan *WebSocketClient
    mutex      sync.RWMutex
}

type WebSocketClient struct {
    hub      *WebSocketHub
    conn     *websocket.Conn
    send     chan []byte
    tenantID int64
    userID   int64
    role     string
}

func NewWebSocketHub() *WebSocketHub {
    return &WebSocketHub{
        clients:    make(map[*WebSocketClient]bool),
        broadcast:  make(chan []byte, 256),
        register:   make(chan *WebSocketClient),
        unregister: make(chan *WebSocketClient),
    }
}

func (hub *WebSocketHub) Run() {
    for {
        select {
        case client := <-hub.register:
            hub.mutex.Lock()
            hub.clients[client] = true
            hub.mutex.Unlock()
            zap.L().Debug("WebSocket client connected",
                zap.Int64("user_id", client.userID),
                zap.Int64("tenant_id", client.tenantID))

        case client := <-hub.unregister:
            hub.mutex.Lock()
            if _, ok := hub.clients[client]; ok {
                delete(hub.clients, client)
                close(client.send)
            }
            hub.mutex.Unlock()
            zap.L().Debug("WebSocket client disconnected")

        case message := <-hub.broadcast:
            hub.mutex.RLock()
            for client := range hub.clients {
                select {
                case client.send <- message:
                default:
                    delete(hub.clients, client)
                    close(client.send)
                }
            }
            hub.mutex.RUnlock()
        }
    }
}

func (hub *WebSocketHub) Broadcast(message []byte) {
    hub.broadcast <- message
}

func (hub *WebSocketHub) Register(client *WebSocketClient) {
    hub.register <- client
}

func (hub *WebSocketHub) Unregister(client *WebSocketClient) {
    hub.unregister <- client
}

func (c *WebSocketClient) ReadPump() {
    defer func() {
        c.hub.Unregister(c)
        c.conn.Close()
    }()

    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
    }
}

func (c *WebSocketClient) WritePump() {
    defer c.conn.Close()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            err := c.conn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                return
            }
        }
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adminapi/websocket_hub_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adminapi/websocket_hub.go internal/adminapi/websocket_hub_test.go
git commit -m "feat: implement WebSocket hub for real-time updates"
```

---

### Task 4.2: Add WebSocket Security Middleware

**Files:**
- Create: `internal/adminapi/websocket_middleware.go`
- Create: `internal/adminapi/websocket_middleware_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/adminapi/websocket_middleware_test.go
package adminapi

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
)

func TestWebSocketAuthMiddleware_RejectsNoToken(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/usage", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    middleware := WebSocketAuthMiddleware()
    err := middleware(func(c echo.Context) error {
        return c.String(http.StatusOK, "OK")
    })(c)

    if err != nil {
        t.Fatalf("Middleware failed: %v", err)
    }

    if rec.Code != http.StatusForbidden {
        t.Errorf("Expected status 403, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adminapi/websocket_middleware_test.go -v`
Expected: FAIL with "undefined: WebSocketAuthMiddleware"

- [ ] **Step 3: Implement WebSocket security middleware**

```go
// internal/adminapi/websocket_middleware.go
package adminapi

import (
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"
    "crypto/rand"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
    "github.com/gorilla/websocket"
    "github.com/redis/go-redis/v9"
)

// Fix #12: Enhanced WebSocket security with config-based origins
var upgrader = websocket.Upgrader{
    ReadBufferSize:   1024,
    WriteBufferSize:  1024,
    MaxMessageSize:   65536,      // 64KB max message size
    HandshakeTimeout: 5 * time.Second,
    // CRITICAL: Load from environment/config instead of hardcoding
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        host := r.Host

        // Load from environment variable
        allowedOriginsEnv := os.Getenv("WS_ALLOWED_ORIGINS")
        if allowedOriginsEnv == "" {
            // Fallback to localhost for development
            allowedOriginsEnv = "http://localhost:3000,http://localhost:8000"
        }

        // Parse comma-separated origins
        allowedOrigins := strings.Split(allowedOriginsEnv, ",")

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

// Fix #12: Add connection rate limiting
func WebSocketRateLimiter(redis *redis.Client) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            secCtx, ok := c.Get("security").(SecurityContext)
            if !ok {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "Not authenticated",
                })
            }

            // Check connection count per user (max 5 concurrent connections)
            key := fmt.Sprintf("ws:connections:%d", secCtx.UserID)
            count, _ := redis.Incr(c.Request().Context(), key).Result()
            redis.Expire(c.Request().Context(), key, 1*time.Hour)

            if count > 5 {
                return c.JSON(http.StatusTooManyRequests, map[string]string{
                    "error": "Too many WebSocket connections (max 5)",
                })
            }

            return next(c)
        }
    }
}

// WebSocketAuthMiddleware validates JWT token before upgrading
func WebSocketAuthMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Verify JWT token from query parameter
            token := c.QueryParam("token")
            if token == "" {
                return c.HTML(http.StatusForbidden, "Authentication required")
            }

            claims, err := validateJWTToken(token)
            if err != nil {
                return c.HTML(http.StatusForbidden, "Invalid token")
            }

            // Store security context in connection
            c.Set("security", SecurityContext{
                UserID:   claims.UserID,
                TenantID: claims.TenantID,
                Role:     claims.Role,
                Username: claims.Username,
            })

            return next(c)
        }
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adminapi/websocket_middleware_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adminapi/websocket_middleware.go internal/adminapi/websocket_middleware_test.go
git commit -m "feat: add WebSocket security middleware with origin checking"
```

---

### Task 4.3: Register WebSocket Route

**Files:**
- Modify: `internal/adminapi/adminapi.go`

- [ ] **Step 1: Write failing test**

```go
// internal/adminapi/adminapi_websocket_test.go
package adminapi

import (
    "testing"
    "github.com/labstack/echo/v4"
)

func TestRegisterWebSocketRoute_RouteExists(t *testing.T) {
    e := echo.New()
    hub := NewWebSocketHub()
    handler := &UsageAPIHandler{hub: hub}

    RegisterUsageRoutes(e, handler)

    routes := e.Routes()
    found := false
    for _, route := range routes {
        if route.Path == "/api/v1/ws/usage" && route.Method == "GET" {
            found = true
            break
        }
    }

    if !found {
        t.Error("WebSocket route not registered")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adminapi/adminapi_websocket_test.go -v`
Expected: FAIL with "WebSocket route not registered"

- [ ] **Step 3: Register WebSocket route**

```go
// Add to internal/adminapi/adminapi.go

func RegisterWebSocketRoute(e *echo.Echo, hub *WebSocketHub, handler *UsageAPIHandler) {
    e.GET("/api/v1/ws/usage",
        WebSocketAuthMiddleware(), // Validate JWT first
        func(c echo.Context) error {
            secCtx := c.Get("security").(SecurityContext)

            // Upgrade to WebSocket
            conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
            if err != nil {
                return err
            }

            // Create client
            client := &WebSocketClient{
                hub:      hub,
                conn:     conn,
                send:     make(chan []byte, 256),
                tenantID: secCtx.TenantID,
                userID:   secCtx.UserID,
                role:     secCtx.Role,
            }

            // Register with hub
            hub.Register(client)

            // Start pumps
            go client.WritePump()
            go client.ReadPump()

            return nil
        })
}
```

- [ ] **Step 4: Update RegisterUsageRoutes to call WebSocket registration**

```go
// Modify internal/adminapi/adminapi.go - add to RegisterUsageRoutes

func RegisterUsageRoutes(e *echo.Echo, handler *UsageAPIHandler) {
    // ... existing route registrations ...

    // Register WebSocket route
    RegisterWebSocketRoute(e, handler.hub, handler)
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/adminapi/adminapi_websocket_test.go -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/adminapi/adminapi.go internal/adminapi/adminapi_websocket_test.go
git commit -m "feat: register WebSocket upgrade route with security"
```

---

## Phase 5: Frontend (Week 5)

### Task 5.1: Create Usage Status Page Component

**Files:**
- Create: `web/src/pages/UsageStatus.tsx`
- Create: `web/src/components/usage/UsageCard.tsx`

- [ ] **Step 1: Write component test**

```typescript
// web/src/pages/UsageStatus.test.tsx
import { render, screen } from '@testing-library/react';
import { UsageStatus } from './UsageStatus';

describe('UsageStatus', () => {
  it('displays remaining time quota', async () => {
    render(<UsageStatus />);

    // Wait for API call
    expect(await screen.findByText(/12h45m remaining/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npm test UsageStatus.test.tsx`
Expected: FAIL with "Cannot find module './UsageStatus'"

- [ ] **Step 3: Implement UsageStatus component**

```typescript
// web/src/pages/UsageStatus.tsx
import React, { useState, useEffect } from 'react';
import { useTranslate } from 'react-admin';
import { UsageCard } from '../components/usage/UsageCard';
import { dataProvider } from '../dataProvider';

export const UsageStatus: React.FC = () => {
  const translate = useTranslate();
  const [usage, setUsage] = useState<UserUsage | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUsage = async () => {
      try {
        const data = await dataProvider.getUsage();
        setUsage(data);
      } catch (err) {
        setError(translate('usage.error.fetch_failed'));
      } finally {
        setLoading(false);
      }
    };

    fetchUsage();
    // Refresh every 60 seconds
    const interval = setInterval(fetchUsage, 60000);
    return () => clearInterval(interval);
  }, [translate]);

  if (loading) return <div>{translate('usage.loading')}</div>;
  if (error) return <div>{error}</div>;
  if (!usage) return null;

  return (
    <div>
      <UsageCard usage={usage} />
    </div>
  );
};
```

```typescript
// web/src/components/usage/UsageCard.tsx
import React from 'react';
import { useTranslate } from 'react-admin';
import { Card, CardContent, Typography } from '@mui/material';
import { AccessTime, DataUsage } from '@mui/icons-material';

interface UsageCardProps {
  usage: UserUsage;
}

export const UsageCard: React.FC<UsageCardProps> = ({ usage }) => {
  const translate = useTranslate();

  const formatTime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h${minutes}m`;
  };

  const formatData = (bytes: number): string => {
    const gb = bytes / (1024 * 1024 * 1024);
    return `${gb.toFixed(2)} GB`;
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          <AccessTime /> {translate('usage.time_quota')}
        </Typography>
        <Typography variant="body1">
          {translate('usage.remaining')}: {formatTime(usage.time_quota_remaining)}
        </Typography>
        <Typography variant="body2" color="textSecondary">
          {translate('usage.used')}: {formatTime(usage.time_quota_used)} / {formatTime(usage.time_quota_total)}
        </Typography>

        <Typography variant="h6" gutterBottom style={{ marginTop: '1rem' }}>
          <DataUsage /> {translate('usage.data_quota')}
        </Typography>
        <Typography variant="body1">
          {translate('usage.remaining')}: {formatData(usage.data_quota_remaining)}
        </Typography>
        <Typography variant="body2" color="textSecondary">
          {translate('usage.used')}: {formatData(usage.data_quota_used)} / {formatData(usage.data_quota_total)}
        </Typography>
      </CardContent>
    </Card>
  );
};
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npm test UsageStatus.test.tsx`
Expected: PASS

- [ ] **Step 5: Add translations**

```typescript
// web/src/i18n/en-US.ts - Add to translations

export default {
  usage: {
    time_quota: 'Time Quota',
    data_quota: 'Data Quota',
    remaining: 'Remaining',
    used: 'Used',
    loading: 'Loading usage data...',
    error: {
      fetch_failed: 'Failed to fetch usage data',
    },
  },
}
```

```typescript
// web/src/i18n/zh-CN.ts - Add Chinese translations

export default {
  usage: {
    time_quota: '时间配额',
    data_quota: '数据配额',
    remaining: '剩余',
    used: '已使用',
    loading: '加载使用数据中...',
    error: {
      fetch_failed: '获取使用数据失败',
    },
  },
}
```

- [ ] **Step 6: Add dataProvider method**

```typescript
// web/src/dataProvider.ts - Add method

export const dataProvider = {
  // ... existing methods

  getUsage: async (): Promise<UserUsage> => {
    const response = await fetch('/api/v1/users/me/usage', {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch usage');
    }

    return response.json();
  },
};
```

- [ ] **Step 7: Commit**

```bash
git add web/src/pages/UsageStatus.tsx web/src/components/usage/UsageCard.tsx web/src/i18n/en-US.ts web/src/i18n/zh-CN.ts
git commit -m "feat: add usage status page component"
```

---

### Task 5.2: Extend DataProvider with Usage Methods

**Files:**
- Modify: `web/src/dataProvider.ts`
- Create: `web/src/dataProvider_test.ts`

- [ ] **Step 1: Write failing test**

```typescript
// web/src/dataProvider_test.ts
import { dataProvider } from './dataProvider';

describe('Usage DataProvider', () => {
  it('fetches user usage', async () => {
    const usage = await dataProvider.getUsage();

    expect(usage).toHaveProperty('time_quota_total');
    expect(usage).toHaveProperty('time_quota_remaining');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npm test dataProvider_test.ts`
Expected: FAIL with "dataProvider.getUsage is not a function"

- [ ] **Step 3: Extend dataProvider**

```typescript
// web/src/dataProvider.ts - Add methods to existing dataProvider

export interface UserUsage {
  time_quota_total: number;
  time_quota_used: number;
  time_quota_remaining: number;
  data_quota_total: number;
  data_quota_used: number;
  data_quota_remaining: number;
  sessions_today: number;
  prediction?: {
    quota_expires_at: string;
    confidence: string;
    message: string;
  };
}

export interface UsageHistoryPoint {
  timestamp: string;
  seconds: number;
  bytes: number;
}

export interface UsageHistoryResponse {
  data: UsageHistoryPoint[];
  granularity: string;
  total_seconds: number;
  total_bytes: number;
}

// Add to existing dataProvider export
export const dataProvider = {
  // ... existing methods ...

  getUsage: async (): Promise<UserUsage> => {
    const token = localStorage.getItem('token');
    const response = await fetch('/api/v1/users/me/usage', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch usage: ${response.statusText}`);
    }

    return response.json();
  },

  getUsageHistory: async (from: string, to: string, granularity: string = 'daily'): Promise<UsageHistoryResponse> => {
    const token = localStorage.getItem('token');
    const url = `/api/v1/users/me/usage/history?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}&granularity=${granularity}`;

    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch usage history: ${response.statusText}`);
    }

    return response.json();
  },

  getActiveSessions: async (): Promise<{total: number; sessions: ActiveSession[]}> => {
    const token = localStorage.getItem('token');
    const response = await fetch('/api/v1/sessions/active', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch active sessions: ${response.statusText}`);
    }

    return response.json();
  },
};
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npm test dataProvider_test.ts`
Expected: PASS (with mocked fetch)

- [ ] **Step 5: Commit**

```bash
git add web/src/dataProvider.ts web/src/dataProvider_test.ts
git commit -m "feat: extend dataProvider with usage analytics methods"
```

---

## Phase 6: Background Sync Worker (Week 6)

### Task 6.1: Implement Redis → PostgreSQL Sync Worker

**Files:**
- Create: `internal/usage/sync_worker.go`
- Create: `internal/usage/sync_worker_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/usage/sync_worker_test.go
package usage

import (
    "context"
    "testing"
    "time"
)

func TestSyncWorker_SyncUsageToPostgreSQL(t *testing.T) {
    worker := NewSyncWorker(nil, nil, 10)

    ctx := context.Background()
    err := worker.SyncUsageToPostgreSQL(ctx)

    if err != nil {
        t.Fatalf("SyncUsageToPostgreSQL failed: %v", err)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/usage/sync_worker_test.go -v`
Expected: FAIL with "undefined: NewSyncWorker"

- [ ] **Step 3: Implement sync worker**

```go
// internal/usage/sync_worker.go
package usage

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type SyncWorker struct {
    redis         *redis.Client
    db            *gorm.DB
    batchSize     int
    syncInterval  time.Duration
}

type UsageStatistics struct {
    ID                uint      `gorm:"primaryKey"`
    TenantID          int64     `gorm:"column:tenant_id"`
    Username          string    `gorm:"column:username"`
    TimeQuotaTotal    int64     `gorm:"column:time_quota_total"`
    TimeQuotaUsed     int64     `gorm:"column:time_quota_used"`
    DataQuotaTotal    int64     `gorm:"column:data_quota_total"`
    DataQuotaUsed     int64     `gorm:"column:data_quota_used"`
    LastUpdatedAt     time.Time `gorm:"column:last_updated_at"`
}

func (UsageStatistics) TableName() string {
    return "usage_statistics"
}

func NewSyncWorker(redis *redis.Client, db *gorm.DB, batchSize int) *SyncWorker {
    return &SyncWorker{
        redis:        redis,
        db:           db,
        batchSize:    batchSize,
        syncInterval: 60 * time.Second,
    }
}

func (sw *SyncWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(sw.syncInterval)
    defer ticker.Stop()

    zap.L().Info("Starting sync worker", zap.Duration("interval", sw.syncInterval))

    for {
        select {
        case <-ctx.Done():
            zap.L().Info("Stopping sync worker")
            return
        case <-ticker.C:
            if err := sw.SyncUsageToPostgreSQL(ctx); err != nil {
                zap.L().Error("Sync failed", zap.Error(err))
            }
        }
    }
}

func (sw *SyncWorker) SyncUsageToPostgreSQL(ctx context.Context) error {
    zap.L().Debug("Starting Redis → PostgreSQL sync")

    // Fix #1: Use SCAN instead of KEYS to avoid blocking Redis at scale
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

    zap.L().Debug("Found usage keys", zap.Int("count", len(keys)))

    // Batch process
    for i := 0; i < len(keys); i += sw.batchSize {
        end := min(i+sw.batchSize, len(keys))
        batch := keys[i:end]

        if err := sw.syncBatch(ctx, batch); err != nil {
            zap.L().Error("Batch sync failed",
                zap.Int("batch_start", i),
                zap.Error(err))
        }
    }

    zap.L.Info("Sync completed", zap.Int("total_keys", len(keys)))
    return nil
}

func (sw *SyncWorker) syncBatch(ctx context.Context, keys []string) error {
    // Fetch all data in parallel using pipeline
    pipe := sw.redis.Pipeline()
    var cmds []*redis.StringCmd

    for _, key := range keys {
        cmds = append(cmds, pipe.Get(ctx, key))
    }

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return err
    }

    // Parse and update PostgreSQL in a transaction
    tx := sw.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    for i, cmd := range cmds {
        val, err := cmd.Result()
        if err == redis.Nil {
            continue // Key expired
        }
        if err != nil {
            zap.L().Warn("Failed to get key", zap.String("key", keys[i]), zap.Error(err))
            continue
        }

        var usage UserUsage
        if err := json.Unmarshal([]byte(val), &usage); err != nil {
            zap.L().Warn("Failed to unmarshal usage", zap.Error(err))
            continue
        }

        // Update or insert usage_statistics
        stats := &UsageStatistics{
            TenantID:       usage.TenantID,
            Username:       usage.Username,
            TimeQuotaTotal: usage.TimeQuotaTotal,
            TimeQuotaUsed:  usage.TimeQuotaUsed,
            DataQuotaTotal: usage.DataQuotaTotal,
            DataQuotaUsed:  usage.DataQuotaUsed,
            LastUpdatedAt:  time.Now(),
        }

        // Upsert
        if err := tx.Save(stats).Error; err != nil {
            zap.L().Error("Failed to save stats", zap.Error(err))
            tx.Rollback()
            return err
        }
    }

    if err := tx.Commit().Error; err != nil {
        zap.L().Error("Batch commit failed", zap.Error(err))
        return err
    }

    return nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/usage/sync_worker_test.go -v`
Expected: PASS

- [ ] **Step 5: Create usage_statistics table migration**

```sql
-- Create usage_statistics table
CREATE TABLE IF NOT EXISTS usage_statistics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    username VARCHAR(255) NOT NULL,
    time_quota_total BIGINT NOT NULL DEFAULT 0,
    time_quota_used BIGINT NOT NULL DEFAULT 0,
    data_quota_total BIGINT NOT NULL DEFAULT 0,
    data_quota_used BIGINT NOT NULL DEFAULT 0,
    last_updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, username)
);

CREATE INDEX idx_usage_stats_tenant ON usage_statistics(tenant_id);
```

- [ ] **Step 6: Start worker in app.go**

```go
// internal/app/app.go - Add to App initialization

func (app *App) StartSyncWorker(ctx context.Context) {
    worker := usage.NewSyncWorker(app.Redis, app.DB, 100)
    go worker.Start(ctx)
}
```

- [ ] **Step 7: Commit**

```bash
git add internal/usage/sync_worker.go internal/usage/sync_worker_test.go
git commit -m "feat: implement background sync worker (Redis → PostgreSQL)"
```

---

## Phase 7: Security & Rate Limiting (Week 7)

### Task 7.1: Implement Role-Based Access Control Middleware

**Files:**
- Create: `internal/middleware/security.go`
- Create: `internal/middleware/security_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/middleware/security_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
)

func TestRequirePermission_Allowed(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/usage", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // Set security context with permissions
    c.Set("security", SecurityContext{
        Role:        "operator",
        Permissions: []string{"users:read:tenant"},
    })

    handler := requirePermission("users:read:tenant")(func(c echo.Context) error {
        return c.String(http.StatusOK, "OK")
    })

    err := handler(c)
    if err != nil {
        t.Fatalf("Handler failed: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/security_test.go -v`
Expected: FAIL with "undefined: requirePermission"

- [ ] **Step 3: Implement security middleware**

```go
// internal/middleware/security.go
package middleware

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type SecurityContext struct {
    UserID      int64
    TenantID    int64
    Role        string
    Username    string
    Permissions []string
}

var permissionMatrix = map[string][]string{
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

func loadPermissions(role string) []string {
    if perms, ok := permissionMatrix[role]; ok {
        return perms
    }
    return []string{}
}

func hasPermission(secCtx SecurityContext, requiredPermission string) bool {
    // Platform admins have all permissions
    if secCtx.Role == "platform_admin" {
        return true
    }

    for _, perm := range secCtx.Permissions {
        if perm == "*" || perm == requiredPermission {
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
                    "error": "Unauthorized",
                })
            }

            if !hasPermission(secCtx, permission) {
                zap.L().Warn("Permission denied",
                    zap.Int64("user_id", secCtx.UserID),
                    zap.String("permission", permission))

                return c.JSON(http.StatusForbidden, map[string]string{
                    "error": "Insufficient permissions",
                })
            }

            return next(c)
        }
    }
}

// Fix #8: ApplyTenantFilter automatically adds tenant WHERE clause
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
// query = h.db.Table("radius_accounting").
//     Select("*").
//     Scopes(ApplyTenantFilter(c)).  // AUTOMATIC filtering
//     Where("acct_status = ?", "active")

func TenantIsolationMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            secCtx, ok := c.Get("security").(SecurityContext)
            if !ok {
                return next(c)
            }

            // Platform admins bypass tenant isolation
            if secCtx.Role == "platform_admin" {
                return next(c)
            }

            targetTenantID := c.Param("tenant_id")
            if targetTenantID != "" && targetTenantID != formatInt64(secCtx.TenantID) {
                zap.L().Warn("Cross-tenant access attempt",
                    zap.Int64("user_id", secCtx.UserID),
                    zap.String("target_tenant", targetTenantID))

                return c.JSON(http.StatusForbidden, map[string]string{
                    "error": "Cross-tenant access denied",
                })
            }

            return next(c)
        }
    }
}

func formatInt64(n int64) string {
    return fmt.Sprintf("%d", n)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/security_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/security.go internal/middleware/security_test.go
git commit -m "feat: implement RBAC middleware with tenant isolation"
```

---

### Task 7.2: Implement Per-Endpoint Rate Limiting

**Files:**
- Create: `internal/middleware/rate_limiter.go`
- Create: `internal/middleware/rate_limiter_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/middleware/rate_limiter_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "sync"
    "github.com/labstack/echo/v4"
)

func TestRateLimit_AllowsWithinLimit(t *testing.T) {
    limiter := NewRateLimiter()

    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/usage", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // Set user identifier
    c.Set("user_id", "123")

    handler := limiter.Middleware("usage_endpoint", 10, 20)(func(c echo.Context) error {
        return c.String(http.StatusOK, "OK")
    })

    // Make 5 requests (within limit)
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := handler(c)
            if err != nil {
                t.Errorf("Request failed: %v", err)
            }
        }()
    }
    wg.Wait()

    if rec.Code == http.StatusTooManyRequests {
        t.Error("Rate limited too early")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/rate_limiter_test.go -v`
Expected: FAIL with "undefined: NewRateLimiter"

- [ ] **Step 3: Implement rate limiter**

```go
// internal/middleware/rate_limiter.go
package middleware

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    "github.com/labstack/echo/v4"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mutex    sync.RWMutex
}

type RateLimitConfig struct {
    Endpoint          string
    RequestsPerSecond int
    Burst             int
    Role              string // Optional: role-specific limits
}

var rateLimitConfigs = []RateLimitConfig{
    {Endpoint: "/api/v1/users/me/usage", RequestsPerSecond: 10, Burst: 20},
    {Endpoint: "/api/v1/users/:id/usage", RequestsPerSecond: 5, Burst: 10, Role: "operator"},
    {Endpoint: "/api/v1/sessions/active", RequestsPerSecond: 30, Burst: 50},
    {Endpoint: "/api/v1/users/me/usage/history", RequestsPerSecond: 2, Burst: 5},
    {Endpoint: "/api/v1/users/me/usage/insights", RequestsPerSecond: 1, Burst: 3},
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
    }
}

func (rl *RateLimiter) getLimiter(key string, rps int, burst int) *rate.Limiter {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    if limiter, exists := rl.limiters[key]; exists {
        return limiter
    }

    limiter := rate.NewLimiter(rate.Limit(rps), burst)
    rl.limiters[key] = limiter
    return limiter
}

func (rl *RateLimiter) Middleware(endpoint string, rps int, burst int) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            secCtx, ok := c.Get("security").(SecurityContext)
            if !ok {
                return next(c)
            }

            // Create user-specific key
            key := fmt.Sprintf("%s:%d", endpoint, secCtx.UserID)
            limiter := rl.getLimiter(key, rps, burst)

            if !limiter.Allow() {
                return c.JSON(http.StatusTooManyRequests, map[string]string{
                    "error": "Rate limit exceeded",
                })
            }

            return next(c)
        }
    }
}

func (rl *RateLimiter) CleanupOldLimiters() {
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for range ticker.C {
            rl.mutex.Lock()
            // Simple cleanup: remove all limiters (they'll be recreated as needed)
            rl.limiters = make(map[string]*rate.Limiter)
            rl.mutex.Unlock()
        }
    }()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/rate_limiter_test.go -v`
Expected: PASS

- [ ] **Step 5: Apply rate limits to usage endpoints**

```go
// internal/adminapi/adminapi.go - Add rate limiting

func RegisterUsageRoutes(e *echo.Echo, handler *UsageAPIHandler) {
    rateLimiter := NewRateLimiter()
    rateLimiter.CleanupOldLimiters()

    api := e.Group("/api/v1")

    // User endpoints with rate limits
    api.GET("/users/me/usage",
        handler.GetUserUsage,
        rateLimiter.Middleware("/users/me/usage", 10, 20),
        requirePermission("users:read:own"))

    api.GET("/users/me/usage/history",
        handler.GetUsageHistory,
        rateLimiter.Middleware("/users/me/usage/history", 2, 5),
        requirePermission("users:read:own"))

    // Admin endpoints with stricter limits
    api.GET("/users/:id/usage",
        handler.GetUserUsageByID,
        rateLimiter.Middleware("/users/:id/usage", 5, 10),
        requirePermission("users:read:tenant"))
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/middleware/rate_limiter.go internal/middleware/rate_limiter_test.go internal/adminapi/adminapi.go
git commit -m "feat: implement per-endpoint rate limiting"
```

---

## Phase 8: Enhanced Analytics (Week 8)

### Task 8.1: Extend Predictive Analytics Engine

**Files:**
- Create: `internal/analytics/enhanced_predictive.go`
- Create: `internal/analytics/pattern_detector.go`

- [ ] **Step 1: Write failing test**

```go
// internal/analytics/enhanced_predictive_test.go
package analytics

import (
    "testing"
    "time"
)

func TestEnhancedPredictiveEngine_DetectPatterns(t *testing.T) {
    db := setupTestDB()
    defer db.Close()

    engine := NewEnhancedPredictiveEngine(db)

    patterns, err := engine.DetectPatterns("testuser", 30)
    if err != nil {
        t.Fatalf("DetectPatterns failed: %v", err)
    }

    if patterns == nil {
        t.Error("Expected non-nil patterns")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/analytics/enhanced_predictive_test.go -v`
Expected: FAIL with "undefined: NewEnhancedPredictiveEngine"

- [ ] **Step 3: Implement enhanced predictive engine**

```go
// internal/analytics/enhanced_predictive.go
package analytics

import (
    "context"
    "time"
    "github.com/talkincode/toughradius/v9/internal/analytics"
    "gorm.io/gorm"
)

type EnhancedPredictiveEngine struct {
    // Embed existing engine
    *analytics.PredictiveEngine

    db              *gorm.DB
    patternDetector *PatternDetector
    anomalyDetector *AnomalyDetector
}

func NewEnhancedPredictiveEngine(db *gorm.DB) *EnhancedPredictiveEngine {
    return &EnhancedPredictiveEngine{
        PredictiveEngine: analytics.NewPredictiveEngine(db),
        db:               db,
        patternDetector:  NewPatternDetector(),
        anomalyDetector:  NewAnomalyDetector(),
    }
}

type UsagePattern struct {
    Username              string            `json:"username"`
    DaysAnalyzed          int               `json:"days_analyzed"`
    AverageDailyUsage     float64           `json:"average_daily_usage"`
    PeakUsageHours        []int             `json:"peak_usage_hours"`
    DayOfWeekPattern      map[string]float64 `json:"day_of_week"`
    WeekendVsWeekday      struct {
        Weekend float64 `json:"weekend"`
        Weekday float64 `json:"weekday"`
    } `json:"weekend_vs_weekday"`
}

func (e *EnhancedPredictiveEngine) DetectPatterns(username string, days int) (*UsagePattern, error) {
    ctx := context.Background()
    startDate := time.Now().AddDate(0, 0, -days)

    var results []struct {
        Hour     int     `json:"hour"`
        DayOfWeek string `json:"day_of_week"`
        IsWeekend bool   `json:"is_weekend"`
        Usage    float64 `json:"usage"`
    }

    err := e.db.Table("radius_accounting").
        Select(`
            EXTRACT(HOUR FROM acct_start_time) as hour,
            TO_CHAR(acct_start_time, 'Day') as day_of_week,
            EXTRACT(DOW FROM acct_start_time) IN (0, 6) as is_weekend,
            SUM(acct_session_time) / 3600.0 as usage
        `).
        Where("username = ? AND acct_start_time > ?", username, startDate).
        Group("hour, day_of_week, is_weekend").
        Scan(&results).Error

    if err != nil {
        return nil, err
    }

    pattern := &UsagePattern{
        Username:         username,
        DaysAnalyzed:     days,
        DayOfWeekPattern: make(map[string]float64),
    }

    // Analyze patterns
    hourlyUsage := make(map[int]float64)
    var totalUsage, weekendUsage, weekdayUsage float64
    var weekendCount, weekdayCount int

    for _, r := range results {
        hourlyUsage[r.Hour] += r.Usage
        pattern.DayOfWeekPattern[r.DayOfWeek] += r.Usage
        totalUsage += r.Usage

        if r.IsWeekend {
            weekendUsage += r.Usage
            weekendCount++
        } else {
            weekdayUsage += r.Usage
            weekdayCount++
        }
    }

    if len(results) > 0 {
        pattern.AverageDailyUsage = totalUsage / float64(days)

        if weekendCount > 0 {
            pattern.WeekendVsWeekday.Weekend = weekendUsage / float64(weekendCount)
        }
        if weekdayCount > 0 {
            pattern.WeekendVsWeekday.Weekday = weekdayUsage / float64(weekdayCount)
        }
    }

    // Find peak hours
    for hour, usage := range hourlyUsage {
        if usage > pattern.AverageDailyUsage * 1.5 {
            pattern.PeakUsageHours = append(pattern.PeakUsageHours, hour)
        }
    }

    return pattern, nil
}
```

```go
// internal/analytics/pattern_detector.go
package analytics

import (
    "context"
    "time"
    "gorm.io/gorm"
)

type PatternDetector struct {
    db *gorm.DB
}

func NewPatternDetector() *PatternDetector {
    return &PatternDetector{}
}

type Anomaly struct {
    Timestamp   time.Time `json:"timestamp"`
    Type        string    `json:"type"` // spike, drop, pattern_break
    Severity    string    `json:"severity"` // low, medium, high
    Description string    `json:"description"`
    Value       float64   `json:"value"`
    Expected    float64   `json:"expected"`
    Deviation   float64   `json:"deviation"` // Z-score
}

type AnomalyDetector struct {
    db *gorm.DB
}

func NewAnomalyDetector() *AnomalyDetector {
    return &AnomalyDetector{}
}

func (ad *AnomalyDetector) DetectAnomalies(username string, days int) ([]Anomaly, error) {
    // Implementation would use statistical analysis (Z-score, IQR, etc.)
    // to detect unusual usage patterns
    return []Anomaly{}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/analytics/enhanced_predictive_test.go -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/analytics/enhanced_predictive.go internal/analytics/pattern_detector.go
git commit -m "feat: extend predictive analytics with pattern detection"
```

---

## Phase 9: Testing & Documentation (Week 9)

### Task 9.1: Write Integration Tests

**Files:**
- Create: `tests/integration/usage_api_integration_test.go`

- [ ] **Step 1: Write integration test**

```go
// tests/integration/usage_api_integration_test.go
package integration

import (
    "context"
    "testing"
    "time"
    "github.com/talkincode/toughradius/v9/internal/usage"
)

func TestUsageAPI_EndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database and Redis
    db := setupTestDB()
    redis := setupTestRedis()
    defer teardown(db, redis)

    tracker := usage.NewUsageTracker(db, redis)

    // Test: Record accounting update
    ctx := context.Background()
    record := &RadiusAccounting{
        TenantID:       1,
        UserName:       "integration_test_user",
        AcctStatus:     "stopped",
        AcctSessionTime: 3600,
        AcctInputOctets: 1048576,
        AcctOutputOctets: 1048576,
    }

    err := tracker.RecordAccountingUpdate(ctx, record)
    if err != nil {
        t.Fatalf("RecordAccountingUpdate failed: %v", err)
    }

    // Test: Retrieve usage
    usageData, err := tracker.GetUserUsage(ctx, 1, "integration_test_user")
    if err != nil {
        t.Fatalf("GetUserUsage failed: %v", err)
    }

    if usageData.TimeQuotaUsed != 3600 {
        t.Errorf("Expected time_used 3600, got %d", usageData.TimeQuotaUsed)
    }

    // Test: Verify Redis was updated
    val, err := redis.Get(ctx, "user:usage:1:integration_test_user").Result()
    if err != nil {
        t.Fatalf("Redis GET failed: %v", err)
    }

    if val == "" {
        t.Error("Redis key not found")
    }
}
```

- [ ] **Step 2: Run integration test**

Run: `go test -v ./tests/integration/usage_api_integration_test.go`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add tests/integration/usage_api_integration_test.go
git commit -m "test: add integration tests for usage API"
```

---

### Task 9.2: Load Testing

**Files:**
- Create: `tests/load/usage_load_test.go`

- [ ] **Step 1: Write load test**

```go
// tests/load/usage_load_test.go
package load

import (
    "context"
    "sync"
    "testing"
    "time"
    "github.com/talkincode/toughradius/v9/internal/usage"
)

func TestUsageTracker_ConcurrentUpdates(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test")
    }

    db := setupTestDB()
    redis := setupTestRedis()
    defer teardown(db, redis)

    tracker := usage.NewUsageTracker(db, redis)
    ctx := context.Background()

    // Simulate 1000 concurrent accounting updates
    numGoroutines := 1000
    updatesPerGoroutine := 10

    var wg sync.WaitGroup
    start := time.Now()

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()

            for j := 0; j < updatesPerGoroutine; j++ {
                record := &RadiusAccounting{
                    TenantID:       1,
                    UserName:       formatUser(userID),
                    AcctStatus:     "stopped",
                    AcctSessionTime: 60, // 1 minute
                    AcctInputOctets: 1024,
                    AcctOutputOctets: 1024,
                }

                tracker.RecordAccountingUpdate(ctx, record)
            }
        }(i)
    }

    wg.Wait()
    duration := time.Since(start)

    t.Logf("Completed %d updates in %v (%.2f updates/sec)",
        numGoroutines*updatesPerGoroutine,
        duration,
        float64(numGoroutines*updatesPerGoroutine)/duration.Seconds())

    // Verify final counts
    usageData, err := tracker.GetUserUsage(ctx, 1, "user_0")
    if err != nil {
        t.Fatalf("GetUserUsage failed: %v", err)
    }

    expectedTime := int64(updatesPerGoroutine * 60)
    if usageData.TimeQuotaUsed != expectedTime {
        t.Errorf("Expected time_used %d, got %d", expectedTime, usageData.TimeQuotaUsed)
    }
}
```

- [ ] **Step 2: Run load test**

Run: `go test -v -timeout 10m ./tests/load/usage_load_test.go`
Expected: PASS with performance metrics

- [ ] **Step 3: Commit**

```bash
git add tests/load/usage_load_test.go
git commit -m "test: add load tests for concurrent usage updates"
```

---

### Task 9.3: Load Test at Target Scale (Fix #13)

**Files:**
- Create: `tests/load/scale_test.go`

- [ ] **Step 1: Write 100K user load test**

```go
// tests/load/scale_test.go
package load

import (
    "context"
    "sort"
    "sync"
    "testing"
    "time"
    "fmt"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/usage"
)

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

                record := &domain.RadiusAccounting{
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

- [ ] **Step 2: Run scale test**

Run: `go test -v -timeout 30m ./tests/load/scale_test.go`
Expected: PASS with P99 latency <100ms

- [ ] **Step 3: Commit**

```bash
git add tests/load/scale_test.go
git commit -m "test: add 100K user scale load test (Fix #13)"
```

---

## Phase 10: Polish & Deployment (Week 10)

### Task 10.1: Create Deployment Scripts

**Files:**
- Create: `scripts/deploy_usage_analytics.sh`
- Create: `scripts/rollback_usage_analytics.sh`

- [ ] **Step 1: Create deployment script**

```bash
#!/bin/bash
# scripts/deploy_usage_analytics.sh

set -e

echo "Deploying Usage Analytics System..."

# 1. Run database migrations
echo "Running database migrations..."
./scripts/migrate_usage_analytics.sh

# 2. Build application
echo "Building application..."
go build -o toughradius .

# 3. Stop existing service
echo "Stopping service..."
systemctl stop toughradius

# 4. Backup current version
echo "Backing up current version..."
cp /usr/local/bin/toughradius /usr/local/bin/toughradius.backup.$(date +%Y%m%d)

# 5. Deploy new version
echo "Deploying new version..."
cp toughradius /usr/local/bin/toughradius

# 6. Start service
echo "Starting service..."
systemctl start toughradius

# 7. Enable sync worker
echo "Starting sync worker..."
# This will be started automatically by app.go

# 8. Verify health
echo "Verifying health..."
sleep 5
curl -f http://localhost:1816/api/v1/health || exit 1

echo "Deployment complete!"
```

- [ ] **Step 2: Create rollback script**

```bash
#!/bin/bash
# scripts/rollback_usage_analytics.sh

echo "Rolling back Usage Analytics deployment..."

# 1. Stop service
systemctl stop toughradius

# 2. Restore backup
LATEST_BACKUP=$(ls -t /usr/local/bin/toughradius.backup.* | head -1)
cp $LATEST_BACKUP /usr/local/bin/toughradius

# 3. Start service
systemctl start toughradius

# 4. Verify health
sleep 5
curl -f http://localhost:1816/api/v1/health || exit 1

echo "Rollback complete!"
```

- [ ] **Step 3: Make scripts executable and commit**

```bash
chmod +x scripts/deploy_usage_analytics.sh scripts/rollback_usage_analytics.sh
git add scripts/deploy_usage_analytics.sh scripts/rollback_usage_analytics.sh
git commit -m "deploy: add deployment and rollback scripts"
```

---

### Task 10.2: Write Documentation

**Files:**
- Create: `docs/USAGE_ANALYTICS.md`
- Create: `docs/USAGE_ANALYTICS_API.md`

- [ ] **Step 1: Write user guide**

```markdown
# Usage Analytics System - User Guide

## Overview

The Usage Analytics system provides real-time visibility into user quota consumption with sub-second response times. Users can view their remaining time and data quotas, session history, and usage predictions.

## Features

### Real-Time Usage Tracking
- **Time Quota**: View total time quota, used time, and remaining time
- **Data Quota**: View total data quota, used data, and remaining data
- **Session Count**: Track number of sessions today
- **Predictions**: Get estimated quota expiry date based on usage patterns

### Historical Data
- View usage history with custom date ranges
- Choose granularity: hourly, daily, weekly
- Export data for offline analysis

## Accessing Usage Data

### User Portal

1. Log in to the user portal
2. Navigate to "My Usage" or "Status" page
3. View real-time usage statistics

### API Access

**Get Current Usage:**
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:1816/api/v1/users/me/usage
```

**Get Historical Usage:**
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:1816/api/v1/users/me/usage/history?from=2026-03-01T00:00:00Z&to=2026-03-27T23:59:59Z&granularity=daily"
```

## Understanding Your Quota

### Time Quota vs Time Window

- **Time Quota**: Total time you can use (e.g., 30 hours)
- **Time Window**: Period during which quota must be used (e.g., 5 days)
- Only active session time counts toward your quota
- Login/logout multiple times within the window - only connected time counts

**Example:**
- You have 30 hours quota to use within 5 days
- Session 1: 5 minutes
- Session 2: 3 hours
- Session 3: 3 hours
- Session 4: 45 minutes
- Session 5: 10 hours
- **Total Used**: 16h50m
- **Remaining**: 13h10m

### Rate Limits

API endpoints are rate-limited to ensure fair access:
- `/users/me/usage`: 10 req/sec
- `/users/me/usage/history`: 2 req/sec
- Active sessions: 30 req/sec

## Troubleshooting

### Usage Not Updating

1. Check if you have active sessions
2. Refresh the page (data updates every 60 seconds)
3. Contact support if issue persists

### Incorrect Quota Display

1. Verify quota assignment in admin panel
2. Check for recent accounting updates
3. Contact administrator for quota corrections

## Privacy

Your usage data is:
- Stored securely with encryption
- Only accessible to you and authorized administrators
- Anonymized upon account deletion per GDPR
- Retained for 12 months for historical analysis
```

- [ ] **Step 2: Write API documentation**

```markdown
# Usage Analytics API Reference

## Base URL

```
http://localhost:1816/api/v1
```

## Authentication

All endpoints require JWT bearer token:

```
Authorization: Bearer <token>
```

## Endpoints

### GET /users/me/usage

Get current usage statistics for authenticated user.

**Response:**
```json
{
  "time_quota_total": 108000,
  "time_quota_used": 60780,
  "time_quota_remaining": 47220,
  "data_quota_total": 4294967296,
  "data_quota_used": 52428800,
  "data_quota_remaining": 4242538816,
  "sessions_today": 5,
  "prediction": {
    "quota_expires_at": "2026-04-15T14:30:00Z",
    "confidence": "high",
    "message": "Based on your usage patterns, you'll run out in 19 days"
  }
}
```

**Rate Limit:** 10 req/sec, burst 20

**Permissions:** `users:read:own`

### GET /users/:id/usage

Get usage statistics for specific user (admin/operator only).

**Parameters:**
- `id` (path): User ID

**Response:** Same as `/users/me/usage`

**Rate Limit:** 5 req/sec, burst 10

**Permissions:** `users:read:tenant`

**Tenant Isolation:** Users must be in same tenant (unless platform admin)

### GET /users/me/usage/history

Get historical usage data with custom date ranges.

**Query Parameters:**
- `from` (required): Start date (ISO 8601)
- `to` (required): End date (ISO 8601)
- `granularity` (optional): `hourly`, `daily`, `weekly` (default: `daily`)

**Response:**
```json
{
  "data": [
    {"timestamp": "2026-03-27T00:00:00Z", "seconds": 1800, "bytes": 1048576},
    {"timestamp": "2026-03-27T01:00:00Z", "seconds": 3600, "bytes": 2097152}
  ],
  "granularity": "hourly",
  "total_seconds": 5400,
  "total_bytes": 3145728
}
```

**Rate Limit:** 2 req/sec, burst 5

**Permissions:** `users:read:own`

### GET /sessions/active

Get real-time active sessions.

**Response:**
```json
{
  "total": 45,
  "sessions": [
    {
      "username": "889956",
      "start_time": "2026-03-27T02:25:00Z",
      "duration_seconds": 300,
      "nas_ip": "192.168.1.20",
      "framed_ip": "10.0.0.100",
      "data_used_this_session": 52428800
    }
  ]
}
```

**Rate Limit:** 30 req/sec, burst 50

**Permissions:** `sessions:read:tenant`

## Error Responses

All endpoints may return these error codes:

- `400 Bad Request`: Invalid parameters
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

**Error Response Format:**
```json
{
  "error": "Error message"
}
```

## WebSocket Support

Real-time usage updates via WebSocket:

**Connect:** `ws://localhost:1816/api/v1/ws/usage?token=<jwt_token>`

**Message Format:**
```json
{
  "username": "889956",
  "time_quota_used": 60780,
  "time_quota_remaining": 47220,
  "timestamp": "2026-03-27T02:30:00Z"
}
```

## Rate Limiting

Per-endpoint rate limits:
- Burst: Short-term allowance
- Sustained: Long-term rate

Exceeding limits returns `429 Too Many Requests` with `Retry-After` header.
```

- [ ] **Step 3: Commit documentation**

```bash
git add docs/USAGE_ANALYTICS.md docs/USAGE_ANALYTICS_API.md
git commit -m "docs: add usage analytics user and API documentation"
```

---

## Phase 11: GDPR Compliance (Week 11)

> **Critical Fix #6:** Legal compliance requirements for data privacy

### Task 11.1: Implement Data Export Endpoint

**Files:**
- Create: `internal/adminapi/gdpr.go`
- Create: `internal/adminapi/gdpr_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/adminapi/gdpr_test.go
package adminapi

import (
    "testing"
    "github.com/labstack/echo/v4"
)

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
// internal/adminapi/gdpr.go
package adminapi

import (
    "encoding/csv"
    "fmt"
    "net/http"
    "time"
    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
)

// GET /api/v1/users/me/gdpr/export
func (h *UsageAPIHandler) ExportUserData(c echo.Context) error {
    secCtx := c.Get("security").(SecurityContext)

    // Collect all user data
    var userData struct {
        Username           string
        Email              string
        Phone              string
        RealName           string
        CreatedAt          time.Time
        AccountingRecords []domain.RadiusAccounting
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

    // Write headers
    writer.Write([]string{"Field", "Value"})

    // Write user data
    writer.Write([]string{"Username", userData.Username})
    writer.Write([]string{"Email", userData.Email})
    writer.Write([]string{"Phone", userData.Phone})
    writer.Write([]string{"Real Name", userData.RealName})
    writer.Write([]string{"Created At", userData.CreatedAt.Format(time.RFC3339)})

    // Write accounting summary
    writer.Write([]string{"Total Sessions", fmt.Sprintf("%d", len(userData.AccountingRecords))})

    // Log export
    zap.L().Info("GDPR data export",
        zap.Int64("user_id", secCtx.UserID),
        zap.String("username", secCtx.Username))

    return nil
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./internal/adminapi/gdpr_test.go -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/adminapi/gdpr.go internal/adminapi/gdpr_test.go
git commit -m "feat: implement GDPR data export endpoint (Task 11.1)"
```

---

### Task 11.2: Implement Right to be Forgotten

- [ ] **Step 1: Implement anonymization endpoint**

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

func generateRandomString(length int) string {
    // Use crypto/rand for secure random generation (GDPR compliant)
    b := make([]byte, length)
    if _, err := rand.Read(b); err != nil {
        // Fallback to timestamp on error (not ideal but functional)
        for i := range b {
            b[i] = byte(time.Now().UnixNano() % 256)
        }
    }

    // Convert random bytes to charset
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    for i := range b {
        b[i] = charset[int(b[i])%len(charset)]
    }
    return string(b)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/adminapi/gdpr.go
git commit -m "feat: implement GDPR right to be forgotten (Task 11.2)"
```

---

### Task 11.3: Implement Data Retention Policy

- [ ] **Step 1: Create retention job**

```go
// internal/jobs/retention_policy.go
package jobs

import (
    "time"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type RetentionPolicyJob struct {
    db *gorm.DB
}

func NewRetentionPolicyJob(db *gorm.DB) *RetentionPolicyJob {
    return &RetentionPolicyJob{db: db}
}

func (j *RetentionPolicyJob) EnforceRetentionPolicy() error {
    // Delete accounting records older than 12 months
    cutoffDate := time.Now().AddDate(-1, 0, 0)

    result := j.db.
        Where("acct_stop_time < ? AND acct_status = ?", cutoffDate, "stopped").
        Delete(&domain.RadiusAccounting{})

    zap.L().Info("Enforced data retention policy",
        zap.Int64("deleted_count", result.RowsAffected))

    return result.Error
}

// Run monthly via cron (proper monthly calculation)
func (j *RetentionPolicyJob) StartMonthlyCron() {
    go func() {
        for {
            now := time.Now()
            // Calculate first day of next month
            nextMonth := now.AddDate(0, 1, 0)
            nextMonth = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, nextMonth.Location())
            duration := nextMonth.Sub(now)

            zap.L().Info("Scheduling next retention policy run",
                zap.Duration("wait_time", duration),
                zap.Time("scheduled_for", nextMonth))

            time.Sleep(duration)
            if err := j.EnforceRetentionPolicy(); err != nil {
                zap.L().Error("Retention policy failed", zap.Error(err))
            }
        }
    }()
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/jobs/retention_policy.go
git commit -m "feat: implement GDPR data retention policy (Task 11.3)"
```

---

## Final Checklist

Before marking the implementation complete:

- [ ] All tests pass (`go test ./...`)
- [ ] Integration tests pass
- [ ] Load tests meet performance targets (<100ms P99)
- [ ] Security review completed
- [ ] Documentation complete
- [ ] Deployment scripts tested
- [ ] Rollback procedure tested
- [ ] Feature flags configured for gradual rollout
- [ ] Monitoring and alerting configured
- [ ] Redis memory usage within expected bounds
- [ ] Background sync worker running successfully
- [ ] WebSocket connections working
- [ ] Rate limiting enforced
- [ ] Tenant isolation verified
- [ ] GDPR compliance verified (anonymization, export)

---

## Success Criteria

The implementation is complete when:

1. **Performance**: P99 latency < 100ms for usage API endpoints
2. **Scalability**: Supports 100,000+ concurrent users
3. **Reliability**: Graceful degradation when Redis fails (fallback to PostgreSQL)
4. **Consistency**: Real-time quota tracking with write-through caching (no stale data)
5. **Security**: RBAC, tenant isolation, rate limiting all enforced
6. **Compliance**: GDPR features (anonymization, export) working

---

**Total Estimated Time:** 11 weeks (updated from 10 weeks)

**Total Tasks:** 52 bite-sized tasks across 11 phases
- Original: 45 tasks
- Added: 7 tasks (Phase 0: 4 tasks, Task 9.3, Phase 11: 3 tasks)

**Critical Fixes Applied:** 15 fixes from 4-agent review
- Fix #1: Redis KEYS → SCAN (Performance)
- Fix #2: sync.Map → LRU cache (Memory safety)
- Fix #3: JWT Authentication (Security)
- Fix #4: RBAC Permissions (Authorization)
- Fix #5: Redis Hash pattern (Architecture)
- Fix #6: GDPR Compliance (Legal)
- Fix #7: Input Validation (Security)
- Fix #8: Tenant Isolation (Security)
- Fix #9: Redis pool 500→2000 (Performance)
- Fix #10: Combined N+1 queries (Performance)
- Fix #11: acct_status backfill (Data integrity)
- Fix #12: WebSocket security (Security)
- Fix #13: 100K load testing (Validation)
- Fix #14: Security headers (Hardening)
- Fix #15: Refactor steps (TDD compliance)

**Next Steps After Plan Approval:**
1. Dispatch plan-document-reviewer for validation
2. Fix any issues found by reviewer
3. Choose execution method (subagent-driven or inline)
4. Begin Phase 1, Task 1.1

**For questions or clarifications during implementation, refer to:**
- Design spec: `docs/superpowers/specs/2026-03-27-enterprise-usage-analytics-design.md`
- Existing analytics: `internal/analytics/predictive.go`
- Redis docs: https://redis.io/docs/
- Echo framework: https://echo.labstack.com/docs
