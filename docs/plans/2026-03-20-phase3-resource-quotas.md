# Phase 3: Resource Quotas & Management Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement comprehensive resource quota system to prevent any provider from degrading performance for others.

**Architecture:** Quota enforcement at API layer (before DB), database layer (schema limits), and RADIUS layer (session limits). Real-time usage tracking with alerts.

**Tech Stack:** Redis for caching, GORM, Background workers, Prometheus for metrics

---

## Task 1: Create Quota Models and Service

**Files:**
- Create: `internal/domain/quota.go`
- Create: `internal/quota/service.go`
- Create: `internal/quota/service_test.go`

**Step 1: Write tests for quota service**

```go
// internal/quota/service_test.go
package quota

import (
    "context"
    "testing"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/tenant"
)

func TestCheckUserQuota(t *testing.T) {
    db := setupTestDB(t)
    service := NewQuotaService(db, nil)

    // Create provider with quota
    quota := &domain.ProviderQuota{
        TenantID:   1,
        MaxUsers:   100,
        MaxOnlineUsers: 50,
    }
    db.Create(quota)

    // Create 99 users
    for i := 0; i < 99; i++ {
        db.Create(&domain.RadiusUser{TenantID: 1, Username: fmt.Sprintf("user%d", i)})
    }

    ctx := tenant.WithTenantID(context.Background(), 1)

    // Should allow 100th user
    err := service.CheckUserQuota(ctx, 1)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    // Create 100th user
    db.Create(&domain.RadiusUser{TenantID: 1, Username: "user100"})

    // Should reject 101st user
    err = service.CheckUserQuota(ctx, 1)
    if err != ErrQuotaExceeded {
        t.Errorf("Expected ErrQuotaExceeded, got %v", err)
    }
}

func TestCheckSessionQuota(t *testing.T) {
    db := setupTestDB(t)
    service := NewQuotaService(db, nil)

    // Create provider with session quota
    quota := &domain.ProviderQuota{
        TenantID:       1,
        MaxOnlineUsers: 10,
    }
    db.Create(quota)

    // Create 10 online sessions
    for i := 0; i < 10; i++ {
        db.Create(&domain.RadiusOnline{
            TenantID: 1,
            Username: fmt.Sprintf("user%d", i),
            AcctSessionID: fmt.Sprintf("session%d", i),
        })
    }

    ctx := tenant.WithTenantID(context.Background(), 1)

    // Should reject new session
    err := service.CheckSessionQuota(ctx, 1)
    if err != ErrMaxSessionsExceeded {
        t.Errorf("Expected ErrMaxSessionsExceeded, got %v", err)
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/quota -run TestQuota -v`
Expected: FAIL with "undefined: quota service"

**Step 3: Implement quota models and service**

```go
// internal/domain/quota.go
package domain

import "time"

type ProviderQuota struct {
    ID               int64     `json:"id" gorm:"primaryKey"`
    TenantID         int64     `json:"tenant_id" gorm:"uniqueIndex"`

    // User Limits
    MaxUsers         int       `json:"max_users" gorm:"default:1000"`
    MaxOnlineUsers   int       `json:"max_online_users" gorm:"default:500"`

    // Device Limits
    MaxNAS           int       `json:"max_nas" gorm:"default:100"`
    MaxMikrotikDevices int     `json:"max_mikrotik_devices" gorm:"default:50"`

    // Storage Limits (GB)
    MaxStorage       int64     `json:"max_storage" gorm:"default:100"`
    MaxDailyBackups  int       `json:"max_daily_backups" gorm:"default:3"`

    // Bandwidth Limits (Gbps)
    MaxBandwidth     float64   `json:"max_bandwidth" gorm:"default:10"`

    // RADIUS Limits (requests per second)
    MaxAuthPerSecond int       `json:"max_auth_per_second" gorm:"default:100"`
    MaxAcctPerSecond int       `json:"max_acct_per_second" gorm:"default:200"`

    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

func (ProviderQuota) TableName() string {
    return "mst_provider_quota"
}

type ProviderUsage struct {
    ID               int64     `json:"id" gorm:"primaryKey"`
    TenantID         int64     `json:"tenant_id" gorm:"index"`

    // Current Usage
    CurrentUsers     int       `json:"current_users"`
    CurrentOnlineUsers int     `json:"current_online_users"`
    CurrentNAS       int       `json:"current_nas"`
    CurrentStorageGB float64   `json:"current_storage_gb"`
    CurrentBandwidth float64   `json:"current_bandwidth"`

    // Period Totals
    TotalAuthRequests int64    `json:"total_auth_requests"`
    TotalAcctRequests int64    `json:"total_acct_requests"`

    UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ProviderUsage) TableName() string {
    return "mst_provider_usage"
}
```

```go
// internal/quota/service.go
package quota

import (
    "context"
    "errors"
    "fmt"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/tenant"
    "gorm.io/gorm"
    "go.uber.org/zap"
)

var (
    ErrQuotaExceeded       = errors.New("quota exceeded")
    ErrMaxSessionsExceeded = errors.New("max sessions exceeded")
    ErrTenantLimitExceeded = errors.New("tenant limit exceeded")
)

type QuotaService struct {
    db    *gorm.DB
    cache *UsageCache
}

func NewQuotaService(db *gorm.DB, cache *UsageCache) *QuotaService {
    return &QuotaService{
        db:    db,
        cache: cache,
    }
}

// GetQuota retrieves quota for a tenant (with cache)
func (s *QuotaService) GetQuota(tenantID int64) (*domain.ProviderQuota, error) {
    // Try cache first
    if s.cache != nil {
        if quota := s.cache.GetQuota(tenantID); quota != nil {
            return quota, nil
        }
    }

    var quota domain.ProviderQuota
    err := s.db.Where("tenant_id = ?", tenantID).First(&quota).Error
    if err != nil {
        // Return default quota if not found
        return &domain.ProviderQuota{
            TenantID:         tenantID,
            MaxUsers:         1000,
            MaxOnlineUsers:   500,
            MaxNAS:           100,
            MaxStorage:       100,
            MaxAuthPerSecond: 100,
            MaxAcctPerSecond: 200,
        }, nil
    }

    if s.cache != nil {
        s.cache.SetQuota(tenantID, &quota)
    }

    return &quota, nil
}

// GetUsage retrieves current usage for a tenant
func (s *QuotaService) GetUsage(tenantID int64) (*domain.ProviderUsage, error) {
    // Try cache first
    if s.cache != nil {
        if usage := s.cache.GetUsage(tenantID); usage != nil {
            return usage, nil
        }
    }

    usage := &domain.ProviderUsage{TenantID: tenantID}

    // Count users
    s.db.Table("radius_user").Where("tenant_id = ?", tenantID).Count(&usage.CurrentUsers)

    // Count active users
    s.db.Table("radius_user").Where("tenant_id = ? AND status = ?", tenantID, "enabled").Count(&usage.CurrentOnlineUsers)

    // Count NAS devices
    s.db.Table("net_nas").Where("tenant_id = ?", tenantID).Count(&usage.CurrentNAS)

    // Count online sessions
    s.db.Table("radius_online").Where("tenant_id = ?", tenantID).Count(&usage.CurrentOnlineUsers)

    if s.cache != nil {
        s.cache.SetUsage(tenantID, usage)
    }

    return usage, nil
}

// CheckUserQuota checks if tenant can create more users
func (s *QuotaService) CheckUserQuota(ctx context.Context, tenantID int64) error {
    quota, err := s.GetQuota(tenantID)
    if err != nil {
        return err
    }

    usage, err := s.GetUsage(tenantID)
    if err != nil {
        return err
    }

    if usage.CurrentUsers >= quota.MaxUsers {
        zap.S().Warn("User quota exceeded",
            zap.Int64("tenant_id", tenantID),
            zap.Int("current", usage.CurrentUsers),
            zap.Int("limit", quota.MaxUsers))
        return ErrQuotaExceeded
    }

    return nil
}

// CheckSessionQuota checks if tenant can have more sessions
func (s *QuotaService) CheckSessionQuota(ctx context.Context, tenantID int64) error {
    quota, err := s.GetQuota(tenantID)
    if err != nil {
        return err
    }

    usage, err := s.GetUsage(tenantID)
    if err != nil {
        return err
    }

    if usage.CurrentOnlineUsers >= quota.MaxOnlineUsers {
        zap.S().Warn("Session quota exceeded",
            zap.Int64("tenant_id", tenantID),
            zap.Int("current", usage.CurrentOnlineUsers),
            zap.Int("limit", quota.MaxOnlineUsers))
        return ErrMaxSessionsExceeded
    }

    return nil
}

// UpdateUsage updates usage metrics for a tenant
func (s *QuotaService) UpdateUsage(tenantID int64) {
    usage, err := s.GetUsage(tenantID)
    if err != nil {
        return
    }

    // Save or update usage record
    s.db.Where("tenant_id = ?", tenantID).
        Assign(usage).
        FirstOrCreate(&domain.ProviderUsage{})
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/quota -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/quota.go internal/quota/service.go internal/quota/service_test.go
git commit -m "feat(quota): add resource quota enforcement system"
```

---

## Task 2: Create Redis Cache for Quota/Usage

**Files:**
- Create: `internal/quota/cache.go`
- Create: `internal/quota/cache_test.go`

**Step 1: Implement Redis cache**

```go
// internal/quota/cache.go
package quota

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/talkincode/toughradius/v9/internal/domain"
)

type UsageCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewUsageCache(client *redis.Client) *UsageCache {
    return &UsageCache{
        client: client,
        ttl:    5 * time.Minute, // Cache for 5 minutes
    }
}

// GetQuota retrieves quota from cache
func (c *UsageCache) GetQuota(tenantID int64) *domain.ProviderQuota {
    ctx := context.Background()
    key := c.quotaKey(tenantID)

    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil
    }

    var quota domain.ProviderQuota
    if err := json.Unmarshal(data, &quota); err != nil {
        return nil
    }

    return &quota
}

// SetQuota stores quota in cache
func (c *UsageCache) SetQuota(tenantID int64, quota *domain.ProviderQuota) {
    ctx := context.Background()
    key := c.quotaKey(tenantID)

    data, _ := json.Marshal(quota)
    c.client.Set(ctx, key, data, c.ttl)
}

// GetUsage retrieves usage from cache
func (c *UsageCache) GetUsage(tenantID int64) *domain.ProviderUsage {
    ctx := context.Background()
    key := c.usageKey(tenantID)

    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil
    }

    var usage domain.ProviderUsage
    if err := json.Unmarshal(data, &usage); err != nil {
        return nil
    }

    return &usage
}

// SetUsage stores usage in cache
func (c *UsageCache) SetUsage(tenantID int64, usage *domain.ProviderUsage) {
    ctx := context.Background()
    key := c.usageKey(tenantID)

    data, _ := json.Marshal(usage)
    c.client.Set(ctx, key, data, c.ttl)
}

// Invalidate clears cache for a tenant
func (c *UsageCache) Invalidate(tenantID int64) {
    ctx := context.Background()
    c.client.Del(ctx, c.quotaKey(tenantID), c.usageKey(tenantID))
}

func (c *UsageCache) quotaKey(tenantID int64) string {
    return fmt.Sprintf("quota:%d", tenantID)
}

func (c *UsageCache) usageKey(tenantID int64) string {
    return fmt.Sprintf("usage:%d", tenantID)
}
```

**Step 2: Commit**

```bash
git add internal/quota/cache.go internal/quota/cache_test.go
git commit -m "feat(quota): add Redis caching for quota and usage metrics"
```

---

## Task 3: Integrate Quota Checks into APIs

**Files:**
- Modify: `internal/adminapi/users.go`
- Modify: `internal/radiusd/auth.go`

**Step 1: Add quota check to user creation**

```go
// In internal/adminapi/users.go - CreateUser function

func CreateUser(c echo.Context) error {
    tenantID, _ := tenant.FromContext(c.Request().Context())

    // Check quota before creating user
    quotaService := GetQuotaService(c)
    if err := quotaService.CheckUserQuota(c.Request().Context(), tenantID); err != nil {
        if errors.Is(err, quota.ErrQuotaExceeded) {
            return fail(c, http.StatusForbidden, "QUOTA_EXCEEDED", "User quota exceeded", nil)
        }
        return fail(c, http.StatusInternalServerError, "QUOTA_ERROR", "Failed to check quota", nil)
    }

    // Proceed with user creation
    // ... existing code
}
```

**Step 2: Add quota check to RADIUS authentication**

```go
// In internal/radiusd/auth.go - AuthenticateUser function

func (s *AuthService) AuthenticateUser(username, password string) (*domain.RadiusUser, error) {
    user, err := s.userRepo.GetByUsername(username)
    if err != nil {
        return nil, err
    }

    // Check session quota
    quotaService := s.quotaService
    if err := quotaService.CheckSessionQuota(context.Background(), user.TenantID); err != nil {
        if errors.Is(err, quota.ErrMaxSessionsExceeded) {
            return nil, errors.New("max_sessions_exceeded")
        }
    }

    // Continue with authentication
    // ... existing code
}
```

**Step 3: Commit**

```bash
git add internal/adminapi/users.go internal/radiusd/auth.go
git commit -m "feat(quota): integrate quota checks into user creation and authentication"
```

---

## Task 4: Create Quota Alert System

**Files:**
- Create: `internal/quota/alert.go`
- Create: `internal/quota/alert_test.go`

**Step 1: Implement quota monitoring and alerts**

```go
// internal/quota/alert.go
package quota

import (
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
)

type AlertService struct {
    quotaService *QuotaService
    emailService *email.Service
}

func NewAlertService(quotaService *QuotaService, emailService *email.Service) *AlertService {
    return &AlertService{
        quotaService: quotaService,
        emailService: emailService,
    }
}

// CheckQuotaUsage monitors all providers and sends alerts
func (s *AlertService) CheckQuotaUsage() {
    var providers []domain.Provider
    s.quotaService.db.Find(&providers)

    for _, provider := range providers {
        quota, _ := s.quotaService.GetQuota(provider.ID)
        usage, _ := s.quotaService.GetUsage(provider.ID)

        // Calculate usage percentages
        userPercent := float64(usage.CurrentUsers) / float64(quota.MaxUsers) * 100
        sessionPercent := float64(usage.CurrentOnlineUsers) / float64(quota.MaxOnlineUsers) * 100

        // Send alerts if approaching limits
        if userPercent > 80 {
            s.sendQuotaWarning(provider.ID, "users", userPercent, usage.CurrentUsers, quota.MaxUsers)
        }
        if sessionPercent > 80 {
            s.sendQuotaWarning(provider.ID, "sessions", sessionPercent, usage.CurrentOnlineUsers, quota.MaxOnlineUsers)
        }
    }
}

func (s *AlertService) sendQuotaWarning(tenantID int64, resourceType string, percent float64, current, limit int) {
    zap.S().Warn("Quota warning",
        zap.Int64("tenant_id", tenantID),
        zap.String("resource", resourceType),
        zap.Float64("percent", percent),
        zap.Int("current", current),
        zap.Int("limit", limit))

    // Get provider contact info
    var provider domain.Provider
    s.quotaService.db.First(&provider, tenantID)

    // Send email alert
    // s.emailService.SendQuotaAlert(provider.Name, email, resourceType, percent)
}

// StartBackgroundMonitoring starts periodic quota checks
func (s *AlertService) StartBackgroundMonitoring(interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            s.CheckQuotaUsage()
        }
    }()
}
```

**Step 2: Start monitoring in app initialization**

```go
// In internal/app/app.go - Init function

func (a *Application) Init(config *config.Config) {
    // ... existing initialization

    // Start quota monitoring
    alertService := quota.NewAlertService(quotaService, emailService)
    alertService.StartBackgroundMonitoring(15 * time.Minute)
}
```

**Step 3: Commit**

```bash
git add internal/quota/alert.go internal/quota/alert_test.go
git commit -m "feat(quota): add quota monitoring and alert system"
```

---

## Success Criteria

- ✅ Quota models and service implemented
- ✅ User and session quotas enforced
- ✅ Redis caching for quota/usage metrics
- ✅ API calls check quotas before operations
- ✅ RADIUS authentication enforces session limits
- ✅ Alert system sends warnings at 80% capacity
- ✅ Unit tests pass (≥80% coverage)
