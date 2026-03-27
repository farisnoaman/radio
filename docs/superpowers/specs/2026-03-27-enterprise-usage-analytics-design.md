# Enterprise-Scale Usage Analytics System Design

**Date:** 2026-03-27
**Author:** Design Team
**Status:** Revised - Redis-Primary Architecture
**Version:** 2.0

---

## Executive Summary

This document outlines the design for an enterprise-scale usage analytics system supporting 100,000+ concurrent users with real-time session tracking, predictive analytics, and role-based access control. The system uses a **Redis-primary architecture** to achieve sub-second response times with real-time consistency while maintaining data persistence in PostgreSQL.

**Key Performance Targets:**
- <10ms response time for cached data (L1 in-memory cache)
- <50ms response time for Redis cache hits
- <100ms P99 response time for all API endpoints
- Support for 100,000+ concurrent users
- Real-time updates within 1 minute (5-10 seconds for active sessions)

**Key Security Features:**
- Role-based access control (User, Operator, Tenant Admin, Platform Admin)
- Tenant isolation enforcement with audit logging
- Encryption at rest and in transit
- GDPR compliance (data anonymization, export)
- Rate limiting per endpoint

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Data Model & Caching Strategy](#2-data-model--caching-strategy)
3. [API Endpoints](#3-api-endpoints)
4. [Real-time Session Tracking](#4-real-time-session-tracking--active-sessions)
5. [Predictive Analytics & Usage Patterns](#5-predictive-analytics--usage-patterns)
6. [Database Optimization](#6-database-optimization--indexing-strategy)
7. [Security Hardening](#7-security-hardening--compliance)
8. [Implementation Strategy](#8-implementation-strategy--migration-plan)

---

## 1. Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ User Status  │  │  Admin Online │  │  Accounting  │      │
│  │   Page       │  │    Page      │  │    Page      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼ (HTTP API + WebSocket)
┌─────────────────────────────────────────────────────────────┐
│                      API Layer (Go)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ /users/:id   │  │/radius/online│  │/radius/      │      │
│  │   /usage     │  │   /sessions  │  │ accounting   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                         │                   │                 │
│                   ┌─────┴─────┐       ┌─────┴─────┐           │
│                   │   Redis   │       │  Postgres   │           │
│                   │ (Primary)  │       │ (Persist)   │           │
│                   │   Store    │       │    Layer    │           │
│                   └───────────┘       └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼ (Async Replication)
┌─────────────────────────────────────────────────────────────┐
│                    Background Workers                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Usage Sync  │  │  Session     │  │  Analytics  │      │
│  │  Worker     │  │  Tracker     │  │  Reporter    │      │
│  │             │  │              │  │              │      │
│  │ Redis → PG  │  │  Redis → PG  │  │  Redis → PG  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

**Architecture Flow:**

1. **Accounting Update Arrives**
   - Write to PostgreSQL `radius_accounting` (source of truth)
   - **Immediately** update Redis with atomic HINCRBY (write-through)
   - Publish event to WebSocket hub for live updates

2. **User Requests Usage Data**
   - L1 cache check (<1ms)
   - Redis fetch (<10ms) - always consistent
   - PostgreSQL fallback only if Redis unavailable

3. **Background Workers (async)**
   - Sync Redis → PostgreSQL every 60 seconds
   - Generate daily aggregates
   - Cleanup old Redis keys

4. **Predictive Analytics**
   - Read from Redis sorted sets (hourly/daily data)
   - Extend existing `internal/analytics/predictive.go`
   - Cache predictions for 5 minutes

### Key Design Decisions

1. **Redis as primary usage store** - Atomic operations (HINCRBY) for real-time quota tracking
2. **PostgreSQL as source of truth** - Async replication from Redis for persistence and reporting
3. **Write-through caching** - Immediate Redis updates on accounting writes (no stale data)
4. **Multi-level caching** - L1 in-memory → L2 Redis → L3 PostgreSQL
5. **Sorted sets for time-series** - Efficient historical data storage in Redis
6. **Existing analytics reuse** - Extend `internal/analytics/predictive.go` instead of duplicating

### Technology Stack

- **Backend:** Go 1.24+ with Echo framework
- **Database:** PostgreSQL 14+ (source of truth, persistent storage)
- **Primary Store:** Redis 7+ (real-time usage tracking, atomic operations)
- **Cache:** Multi-level: L1 in-memory (sync.Map) → L2 Redis → L3 PostgreSQL
- **Frontend:** React with TypeScript, react-admin framework
- **Real-time:** WebSocket with Redis Pub/Sub (secured)
- **Analytics:** Extend existing `internal/analytics/predictive.go`
- **Monitoring:** Prometheus + Grafana

---

## 2. Data Model & Redis-Primary Storage Strategy

### Database Schema Additions

**First, add the missing `acct_status` column:**

```sql
-- Migration to add acct_status column
ALTER TABLE radius_accounting ADD COLUMN IF NOT EXISTS acct_status VARCHAR(20);

-- Backfill existing data
UPDATE radius_accounting
SET acct_status = CASE
    WHEN acct_stop_time IS NULL OR acct_stop_time = '0001-01-01 00:00:00' THEN 'active'
    ELSE 'stopped'
END;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_accounting_status
ON radius_accounting(tenant_id, username, acct_status);

-- Add trigger for automatic status updates
CREATE OR REPLACE FUNCTION update_acct_status()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.acct_stop_time IS NULL OR NEW.acct_stop_time = '0001-01-01 00:00:00' THEN
        NEW.acct_status := 'active';
    ELSE
        NEW.acct_status := 'stopped';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS acct_status_update ON radius_accounting;
CREATE TRIGGER acct_status_update
BEFORE INSERT OR UPDATE ON radius_accounting
FOR EACH ROW EXECUTE FUNCTION update_acct_status();
```

### Redis as Primary Usage Store

#### User Usage Statistics (Real-time)

```redis
# Primary usage data (TTL: 60 seconds, auto-refreshed on writes)
user:usage:{tenant_id}:{user_id} = Hash {
    "time_quota_total": 108000,
    "time_quota_used": 60780,           # Atomic HINCRBY updates
    "time_quota_remaining": 47220,      # Calculated on read
    "data_quota_total": 4294967296,
    "data_quota_used": 52428800,         # Atomic HINCRBY updates
    "data_quota_remaining": 4242538816,   # Calculated on read
    "session_count_today": 5,            # INCR on session start
    "last_accounting_update": "2026-03-27T02:30:00Z"
}

# Active session tracking (TTL: 300 seconds, auto-extend)
session:active:{tenant_id}:{username} = Hash {
    "session_id": "unique-session-id",
    "start_time": "2026-03-27T02:25:00Z",
    "duration_seconds": 300,
    "nas_ip_address": "192.168.1.20",
    "framed_ip_address": "10.0.0.100",
    "time_used_this_session": 300,
    "data_used_this_session": 52428800
}
```

#### Time-Series Data Storage (Sorted Sets)

```redis
# Hourly usage data (Sorted set by timestamp)
ZADD user:hourly:{tenant_id}:{user_id}:2026-03-27 {
    "1711591200": "3600",              # timestamp: seconds
    "1711594800": "1800",              # next hour
    "1711598400": "5400"               # next hour
}

# Daily aggregates (Sorted set by date)
ZADD user:daily:{tenant_id}:{user_id} {
    "1711591200": "43200",             # Mar 27: 43200 seconds
    "1711504800": "72000",             # Mar 26: 72000 seconds
    "1711418400": "50400"              # Mar 25: 50400 seconds
}
```

#### Real-time Write-Through Cache

```go
// internal/service/usage_tracker.go

type UsageTracker struct {
    redis    *redis.Client
    db       *gorm.DB
    eventBus *EventBus
}

// Write-through: Update both Redis and DB atomically
func (ut *UsageTracker) RecordAccounting(ctx context.Context, record *RadiusAccounting) error {
    // 1. Write to PostgreSQL (source of truth)
    if err := ut.db.Create(record).Error; err != nil {
        return err
    }

    // 2. Immediately update Redis (write-through cache)
    usageKey := fmt.Sprintf("user:usage:%d:%s", record.TenantID, record.Username)

    // Atomic increment operations
    pipe := ut.redis.Pipeline()
    pipe.HIncrBy(ctx, usageKey, "time_quota_used", record.AcctSessionTime)
    pipe.HIncrBy(ctx, usageKey, "data_quota_used", record.AcctInputOctets + record.AcctOutputOctets)
    pipe.Expire(ctx, usageKey, 60*time.Second)
    pipe.Exec(ctx)

    // 3. Add to hourly time-series sorted set
    hourlyKey := fmt.Sprintf("user:hourly:%d:%s:%s", record.TenantID, record.Username,
        time.Now().Format("2006-01-02"))
    timestamp := record.AcctStartTime.Unix()
    ut.redis.ZAdd(ctx, hourlyKey, redis.Z{Score: float64(record.AcctSessionTime), Member: fmt.Sprintf("%d", timestamp)})

    // 4. Publish event for WebSocket (real-time updates)
    ut.eventBus.Publish("usage:update", SessionEvent{
        EventType: "accounting_update",
        Username:  record.Username,
        TenantID:  record.TenantID,
        Duration:   record.AcctSessionTime,
        DataUsed:   record.AcctInputOctets + record.AcctOutputOctets,
    })

    return nil
}

// Read from Redis (always consistent)
func (ut *UsageTracker) GetUserUsage(ctx context.Context, tenantID int64, username string) (*UserUsage, error) {
    usageKey := fmt.Sprintf("user:usage:%d:%s", tenantID, username)

    // Try Redis first (<10ms)
    val, err := ut.redis.HGetAll(ctx, usageKey).Result()
    if err == nil {
        var usage UserUsage
        if err := json.Unmarshal([]byte(val), &usage); err == nil {
            // Calculate remaining quota
            usage.TimeQuotaRemaining = usage.TimeQuotaTotal - usage.TimeQuotaUsed
            usage.DataQuotaRemaining = usage.DataQuotaTotal - usage.DataQuotaUsed
            return &usage, nil
        }
    }

    // Redis unavailable - fall back to DB
    var user RadiusUser
    if err := ut.db.Where("tenant_id = ? AND username = ?", tenantID, username).First(&user).Error; err != nil {
        return nil, err
    }

    // Query both accounting (stopped) and online (active) tables
    var timeUsed, dataUsed int64

    ut.db.Model(&RadiusAccounting{}).
        Where("tenant_id = ? AND username = ? AND acct_status = ?", tenantID, username, "stopped").
        Select("COALESCE(SUM(acct_session_time), 0)").
        Scan(&timeUsed)

    ut.db.Model(&RadiusOnline{}).
        Where("tenant_id = ? AND username = ?", tenantID, username).
        Select("COALESCE(SUM(acct_session_time), 0)").
        Scan(&dataUsed)

    usage := &UserUsage{
        TimeQuotaTotal:     user.TimeQuota,
        TimeQuotaUsed:      timeUsed,
        TimeQuotaRemaining: user.TimeQuota - timeUsed,
        DataQuotaTotal:     user.DataQuota,
        DataQuotaUsed:      dataUsed,
        DataQuotaRemaining: user.DataQuota - dataUsed,
    }

    // Populate Redis for next time
    data, _ := json.Marshal(usage)
    ut.redis.Set(ctx, usageKey, data, 60*time.Second)

    return usage, nil
}
```

---

## 3. API Endpoints

### New Endpoints

#### GET /api/v1/users/me/usage

User's own usage statistics (self-access only)

**Response:**
```json
{
    "time_quota_total": 108000,
    "time_quota_used": 60780,
    "time_quota_remaining": 47220,
    "data_quota_total": 4294967296,
    "data_quota_used": 0,
    "data_quota_remaining": 4294967296,
    "sessions_today": 5,
    "prediction": {
        "quota_expires_at": "2026-04-15T14:30:00Z",
        "confidence": "high",
        "message": "Based on your usage patterns, you'll run out in 19 days"
    }
}
```

**Authentication:** Required (JWT token)
**Authorization:** Users can only access their own data
**Rate Limit:** 10 req/sec, burst 20
**Cache:** L1 + L2 cached (60s TTL)

#### GET /api/v1/users/:id/usage

Admin access to user usage statistics

**Response:** Same as above + user details

**Authentication:** Required (JWT token)
**Authorization:** Operators and Admins only
**Tenant Isolation:** Users must be in same tenant (unless Platform Admin)
**Rate Limit:** 5 req/sec, burst 10
**Cache:** L1 + L2 cached (60s TTL)

#### GET /api/v1/users/me/usage/history

Historical usage data with custom date ranges

**Query Parameters:**
- `from` (required): Start date (ISO 8601)
- `to` (required): End date (ISO 8601)
- `granularity` (optional): hourly, daily, weekly (default: daily)

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

**Authentication:** Required
**Rate Limit:** 2 req/sec, burst 5
**Cache:** L1 + L2 cached (varies by date range)

#### GET /api/v1/sessions/active

Real-time active sessions data

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
    ],
    "last_updated": "2026-03-27T02:30:00Z"
}
```

**Authentication:** Required
**Authorization:** Users see own sessions, Operators see tenant sessions, Admins see all
**Rate Limit:** 30 req/sec, burst 50
**Cache:** Redis real-time data (5s TTL)

#### GET /api/v1/users/me/usage/insights

Usage analytics and predictive insights

**Response:**
```json
{
    "quota_status": {
        "time_quota_total": 108000,
        "time_quota_remaining": 47220,
        "usage_percentage": 43.9
    },
    "prediction": {
        "quota_expires_at": "2026-04-15T14:30:00Z",
        "confidence": 0.87,
        "days_remaining": 19
    },
    "patterns": {
        "average_session_duration": 3600,
        "peak_usage_hours": [18, 19, 20, 21],
        "day_of_week_patterns": {
            "monday": 7200,
            "tuesday": 5400,
            "wednesday": 3600
        },
        "weekend_vs_weekday": {
            "weekend": 10800,
            "weekday": 5400
        }
    },
    "recommendations": [
        "⚠️ You've used 43% of your time quota.",
        "💡 Your peak usage is around 18:00. Consider off-peak downloads.",
        "📊 You use 2x more data on weekends."
    ],
    "anomalies": [
        {
            "timestamp": "2026-03-26T15:00:00Z",
            "type": "spike",
            "severity": "medium",
            "description": "Unusual spike in usage (2.3σ deviation)"
        }
    ]
}
```

**Authentication:** Required
**Rate Limit:** 1 req/sec, burst 3
**Cache:** L1 + L2 cached (5 min TTL)

#### GET /api/v1/tenants/:id/usage/stats

Aggregated tenant statistics (Admin only)

**Response:**
```json
{
    "total_users": 1250,
    "active_users_today": 890,
    "total_time_quota_pool": 135000000,
    "total_time_used": 67500000,
    "usage_percentage": 50,
    "top_consumers": [
        {"username": "user1", "time_used": 7200, "time_quota": 10800},
        {"username": "user2", "time_used": 5400, "time_quota": 7200}
    ]
}
```

**Authentication:** Required
**Authorization:** Tenant Admin and Platform Admin only
**Rate Limit:** 5 req/sec, burst 10
**Cache:** L1 + L2 cached (5 min TTL)

### Enhanced Existing Endpoints

#### GET /api/v1/users/:id

Augmented with usage statistics

**New Response Fields:**
```json
{
    "id": 123,
    "username": "889956",
    "expire_time": "2026-03-31T00:00:00Z",
    "time_quota": 108000,
    "data_quota": 4294967296,
    "time_used": 60780,
    "time_remaining": 47220,
    "data_used": 0,
    "data_remaining": 4294967296
}
```

#### GET /api/v1/radius/online

Enhanced with real-time session data

**New Response Fields:**
```json
{
    "total": 45,
    "online": [
        {
            "username": "889956",
            "nas_ip_address": "192.168.1.20",
            "acct_start_time": "2026-03-27T02:25:00Z",
            "session_duration": "5m 23s",
            "data_used_this_session": 52428800
        }
    ]
}
```

---

## 4. Real-time Session Tracking & Active Sessions

### Session Event Bus

```go
type SessionEvent struct {
    EventType     string    `json:"event_type"`     // start, update, end
    Username      string    `json:"username"`
    TenantID      int64     `json:"tenant_id"`
    SessionID     string    `json:"session_id"`
    NASIPAddress  string    `json:"nas_ip_address"`
    StartTime     time.Time `json:"start_time"`
    Duration      int64     `json:"duration_seconds"`
    DataUsed      int64     `json:"data_used"`
    Timestamp     time.Time `json:"timestamp"`
}

type SessionTracker struct {
    redis    *redis.Client
    eventBus *EventBus
    logger   *zap.Logger
}

func (st *SessionTracker) UpdateSession(ctx context.Context, event SessionEvent) error {
    key := fmt.Sprintf("sessions:active:%d", event.TenantID)

    sessionData := map[string]interface{}{
        "username":         event.Username,
        "session_id":       event.SessionID,
        "start_time":       event.StartTime,
        "duration_seconds": event.Duration,
        "nas_ip":           event.NASIPAddress,
        "last_update":      time.Now(),
    }

    // Store in Redis hash with TTL
    if err := st.redis.HSet(ctx, key, event.Username, sessionData).Err(); err != nil {
        return err
    }

    // Refresh TTL to 5 minutes on every update
    return st.redis.Expire(ctx, key, 5*time.Minute).Err()
}

func (st *SessionTracker) PublishEvent(ctx context.Context, event SessionEvent) error {
    return st.eventBus.Publish("session:events", event)
}
```

### WebSocket Hub for Live Updates

```go
type SessionHub struct {
    clients    map[string]*SessionClient
    register   chan *SessionClient
    unregister chan *SessionClient
    broadcast  chan SessionEvent
    redis      *redis.Client
}

func (sh *SessionHub) Run(ctx context.Context) {
    pubsub := sh.redis.Subscribe(ctx, "session:events")

    for {
        select {
        case msg := <-pubsub.Channel():
            var event SessionEvent
            json.Unmarshal([]byte(msg.Payload), &event)
            sh.broadcast <- event

        case client := <-sh.register:
            sh.clients[client.ID] = client

        case client := <-sh.unregister:
            delete(sh.clients, client.ID)
            close(client.Send)

        case event := <-sh.broadcast:
            for _, client := range sh.clients {
                if client.CanView(event.TenantID) {
                    client.Send <- event
                }
            }
        }
    }
}
```

### WebSocket Security (CRITICAL FIX)

```go
// SECURE WebSocket upgrader with origin checking
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // CRITICAL: Only allow same-origin requests
        origin := r.Header.Get("Origin")
        host := r.Host

        // Allow only same-origin or explicitly allowed domains
        allowedOrigins := []string{
            "https://your-domain.com",
            "https://admin.your-domain.com",
        }

        for _, allowed := range allowedOrigins {
            if origin == allowed {
                return true
            }
        }

        // Reject cross-origin requests
        zap.L().Warn("WebSocket rejected: cross-origin request",
            zap.String("origin", origin),
            zap.String("host", host))
        return false
    },
}

// Authentication middleware for WebSocket connections
func authenticateWebSocket(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Verify JWT token before upgrading
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
        })

        return next(c)
    }
}
```

### Multi-level Caching for Performance

```go
type SessionCache struct {
    l1Cache *sync.Map      // In-memory cache (hot data, <1ms access)
    l2Cache *redis.Client  // Redis cache (warm data, <10ms access)
    db      *gorm.DB       // PostgreSQL (cold data, <100ms access)
}

func (sc *SessionCache) GetSession(ctx context.Context, username string) (*Session, error) {
    // Try L1 first (<1ms)
    if sess, ok := sc.l1Cache.Load(username); ok {
        return sess.(*Session), nil
    }

    // Try L2 Redis (<10ms)
    val, err := sc.l2Cache.Get(ctx, fmt.Sprintf("session:%s", username)).Result()
    if err == nil {
        var session Session
        json.Unmarshal([]byte(val), &session)
        sc.l1Cache.Store(username, &session)
        return &session, nil
    }

    // Fallback to DB (<100ms)
    var session Session
    if err := sc.db.Where("username = ?", username).First(&session).Error; err != nil {
        return nil, err
    }

    // Populate caches
    sc.l1Cache.Store(username, &session)
    data, _ := json.Marshal(session)
    sc.l2Cache.Set(ctx, fmt.Sprintf("session:%s", username), data, 5*time.Minute)

    return &session, nil
}
```

### Connection Pooling (FIXED)

```go
// Calculate pool size based on Little's Law: L = λW
// λ = 100000 users * 10 req/sec/user = 1,000,000 requests/sec
// W = 10ms average query time
// Pool size = (1,000,000 * 0.010) / 0.8 = ~12,500 connections
// With 3 servers: ~4,200 connections per server

redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     500,              // Sufficient for 100K concurrent users
    MinIdleConns: 50,               // Keep 50 connections ready
    MaxRetries:   3,                // Retry failed commands
    DialTimeout:  50 * time.Millisecond,
    ReadTimeout:  100 * time.Millisecond,
    WriteTimeout: 100 * time.Millisecond,
    PoolTimeout:  4 * time.Second,   // Wait for connection before creating new one
    MaxRetries:   3,
})
```

---

## 5. Predictive Analytics & Usage Patterns

### Extend Existing Analytics Engine

**IMPORTANT:** We will extend the existing `internal/analytics/predictive.go` module instead of creating a duplicate.

```go
// internal/analytics/predictive_engine.go
// EXTENDS: internal/analytics/predictive.go (existing file)

type EnhancedPredictiveEngine struct {
    // Embed existing engine
    *analytics.PredictiveEngine

    // Add new capabilities
    patternDetector *PatternDetector
    anomalyDetector *AnomalyDetector
    cache          *lru.Cache
}

func NewEnhancedPredictiveEngine(db *gorm.DB) *EnhancedPredictiveEngine {
    return &EnhancedPredictiveEngine{
        PredictiveEngine: analytics.NewPredictiveEngine(db),
        patternDetector:  NewPatternDetector(),
        anomalyDetector: NewAnomalyDetector(),
        cache:          lru.New(10000),
    }
}

func (epe *EnhancedPredictiveEngine) PredictWithPatterns(username string, timeQuota int64) (*EnhancedPrediction, error) {
    // Reuse existing prediction logic
    basePrediction, err := epe.PredictiveEngine.Predict(username, timeQuota)
    if err != nil {
        return nil, err
    }

    // Add pattern analysis
    patterns, _ := epe.patternDetector.DetectPatterns(username, timeQuota)

    // Detect anomalies
    anomalies := epe.anomalyDetector.DetectAnomalies(username, patterns)

    return &EnhancedPrediction{
        BasePrediction: basePrediction,
        Patterns:      patterns,
        Anomalies:     anomalies,
        Recommendations: epe.generateRecommendations(basePrediction, patterns),
    }, nil
}
```

### Prediction Models (Extended)

1. **Linear Regression** - (existing, reused) - Fits a line through usage data points
2. **Moving Average** - (existing, reused) - Calculates average of last N days
3. **Exponential Smoothing** - (existing, reused) - Weights recent data more heavily
4. **Pattern-Based** - (NEW) - Uses detected patterns for prediction
5. **Anomaly-Aware** - (NEW) - Adjusts predictions based on detected anomalies

### Pattern Detection

```go
type UsagePatterns struct {
    HourOfDayPattern      map[int]float64   `json:"hour_of_day"`
    DayOfWeekPattern      map[string]float64 `json:"day_of_week"`
    AverageSessionDuration float64          `json:"avg_session_duration"`
    PeakUsageHours        []int             `json:"peak_hours"`
    LowUsageHours         []int             `json:"low_hours"`
    WeekendVsWeekday      struct {
        Weekend float64 `json:"weekend"`
        Weekday float64 `json:"weekday"`
    } `json:"weekend_vs_weekday"`
}
```

### Anomaly Detection

```go
type Anomaly struct {
    Timestamp   time.Time `json:"timestamp"`
    Type        string    `json:"type"` // spike, drop, pattern_break
    Severity    string    `json:"severity"` // low, medium, high
    Description string    `json:"description"`
    Value       float64   `json:"value"`
    Expected    float64   `json:"expected"`
    Deviation   float64   `json:"deviation"` // Z-score
}
```

---

## 6. Database Optimization & Indexing Strategy

### Critical Indexes

```sql
-- Primary accounting query index
CREATE INDEX CONCURRENTLY idx_accounting_tenant_user_time
ON radius_accounting(tenant_id, username, acct_start_time DESC)
WHERE acct_status = 'stopped';

-- Historical aggregation index (for usage_statistics table queries)
CREATE INDEX CONCURRENTLY idx_accounting_hourly_aggregate
ON radius_accounting(
    tenant_id,
    username,
    DATE_TRUNC('hour', acct_start_time)
)
WHERE acct_status = 'stopped';

-- Active session lookup index
CREATE INDEX CONCURRENTLY idx_accounting_active_sessions
ON radius_accounting(tenant_id, username, acct_unique_id)
WHERE acct_status = 'active';

-- Data quota tracking index
CREATE INDEX CONCURRENTLY idx_accounting_data_usage
ON radius_accounting(tenant_id, username, (acct_input_octets + acct_output_octets))
WHERE acct_status = 'stopped';

-- Partial index for recent records (hot data)
CREATE INDEX CONCURRENTLY idx_accounting_recent
ON radius_accounting(tenant_id, username, acct_start_time DESC)
WHERE acct_start_time > NOW() - INTERVAL '30 days';

-- Tenant-level stats index
CREATE INDEX CONCURRENTLY idx_accounting_tenant_stats
ON radius_accounting(tenant_id, acct_start_time)
INCLUDE (username, acct_session_time, acct_input_octets, acct_output_octets);
```

### Table Partitioning

```sql
-- Monthly partitioning for radius_accounting
CREATE TABLE radius_accounting_y2026m01 PARTITION OF radius_accounting
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE radius_accounting_y2026m02 PARTITION OF radius_accounting
FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

-- Automatic partition creation
CREATE OR REPLACE FUNCTION create_monthly_partitions()
RETURNS void AS $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN 0..2 LOOP
        start_date := DATE_TRUNC('month', NOW() + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        partition_name := 'radius_accounting_y' || TO_CHAR(start_date, 'YYYY') || 'm' || TO_CHAR(start_date, 'MM');

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF radius_accounting
            FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

### Redis Memory Planning & Configuration

**Memory Requirements for 100K Concurrent Users:**

```go
// Per-user memory calculation:
// - Hash with 6 fields × 8 bytes (int64) = 48 bytes
// - Redis overhead ~80 bytes per key
// - Total per user: ~130 bytes
// - 100,000 users: ~13 MB for active usage data

// Hourly time-series (30-day retention):
// - 30 days × 24 hours × 8 bytes (score) × 50 bytes (member) = 288 KB per user
// - For 10K active users: ~2.8 GB
// - Compressed with Redis 7+ data compression: ~1.4 GB

// Daily aggregates (1-year retention):
// - 365 days × 8 bytes × 50 bytes = 144 KB per user
// - For 100K users: ~14.4 GB
// - Compressed: ~7 GB

// Total Redis memory: ~10 GB for full historical data
```

**Redis Configuration Optimization:**

```conf
# redis.conf for enterprise usage analytics

# Memory management
maxmemory 12gb
maxmemory-policy allkeys-lru
save ""  # Disable RDB snapshots (use AOF instead)

# AOF persistence (write-only from Redis)
appendonly yes
appendfsync everysec
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb

# Replication for HA
replica-read-only yes
replica-serve-stale-data no

# Performance tuning
tcp-backlog 511
timeout 0
tcp-keepalive 300
maxclients 10000

# Activerehashing for better key distribution
activerehashing yes

# Latency monitoring
latency-monitor-threshold 100
```

### Background Sync Worker (Redis → PostgreSQL)

```go
type SyncWorker struct {
    redis       *redis.Client
    db          *gorm.DB
    batchSize   int
    syncInterval time.Duration
}

func (sw *SyncWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(sw.syncInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            sw.syncUsageToPostgreSQL(ctx)
        }
    }
}

func (sw *SyncWorker) syncUsageToPostgreSQL(ctx context.Context) {
    // Get all user usage keys
    keys, err := sw.redis.Keys(ctx, "user:usage:*").Result()
    if err != nil {
        zap.L().Error("failed to get usage keys", zap.Error(err))
        return
    }

    // Batch process
    for i := 0; i < len(keys); i += sw.batchSize {
        end := min(i+sw.batchSize, len(keys))
        batch := keys[i:end]

        sw.syncBatch(ctx, batch)
    }
}

func (sw *SyncWorker) syncBatch(ctx context.Context, keys []string) {
    pipe := sw.redis.Pipeline()

    // Fetch all data in parallel
    var cmds []*redis.StringCmd
    for _, key := range keys {
        cmds = append(cmds, pipe.Get(ctx, key))
    }

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        zap.L().Error("pipeline failed", zap.Error(err))
        return
    }

    // Parse and update PostgreSQL
    tx := sw.db.Begin()
    for i, cmd := range cmds {
        var usage UserUsage
        if err := json.Unmarshal([]byte(cmd.Val()), &usage); err != nil {
            continue
        }

        // Update usage_statistics table
        tx.Save(&usage)
    }

    if err := tx.Commit().Error; err != nil {
        zap.L().Error("batch commit failed", zap.Error(err))
        tx.Rollback()
    }
}
```

### Query Performance Monitoring

```go
type PerformanceMonitor struct {
    slowQueryThreshold time.Duration
    metrics            *prometheus.Registry
}

func (pm *PerformanceMonitor) RecordQuery(ctx context.Context, query string, duration time.Duration) {
    // Track query duration
    pm.metrics.With(prometheus.Labels{
        "query_type": pm.categorizeQuery(query),
    }).Observe(duration.Seconds())

    // Alert on slow queries
    if duration > pm.slowQueryThreshold {
        zap.L().Warn("slow query detected",
            zap.String("query", query),
            zap.Duration("duration", duration))
    }
}
```

---

## 7. Security Hardening & Compliance

### Role-Based Access Control

```go
type SecurityContext struct {
    UserID      int64
    TenantID    int64
    Role        string // user, operator, tenant_admin, platform_admin
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
```

### Security Middleware

```go
func SessionSecurityMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := getJWTToken(c)
            claims, err := validateToken(token)
            if err != nil {
                return fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid authentication", nil)
            }

            secCtx := SecurityContext{
                UserID:   claims.UserID,
                TenantID: claims.TenantID,
                Role:     claims.Role,
            }
            secCtx.Permissions = loadPermissions(secCtx.Role)
            c.Set("security", secCtx)

            return next(c)
        }
    }
}

func requirePermission(permission string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            secCtx := c.Get("security").(SecurityContext)

            if !hasPermission(secCtx, permission) {
                logAuditEvent(c, "permission_denied", permission)
                return fail(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions", nil)
            }

            return next(c)
        }
    }
}

func TenantIsolationMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            secCtx := c.Get("security").(SecurityContext)

            if secCtx.Role == "platform_admin" {
                return next(c)
            }

            targetTenantID := c.Param("tenant_id")
            if targetTenantID != "" && targetTenantID != fmt.Sprint(secCtx.TenantID) {
                logAuditEvent(c, "cross_tenant_attempt", targetTenantID)
                return fail(c, http.StatusForbidden, "CROSS_TENANT", "Cross-tenant access denied", nil)
            }

            return next(c)
        }
    }
}
```

### Rate Limiting

```go
var rateLimitConfigs = []RateLimitConfig{
    {Endpoint: "/api/v1/users/me/usage", RequestsPerSecond: 10, Burst: 20},
    {Endpoint: "/api/v1/users/:id/usage", RequestsPerSecond: 5, Burst: 10, Role: "operator"},
    {Endpoint: "/api/v1/sessions/active", RequestsPerSecond: 30, Burst: 50},
    {Endpoint: "/api/v1/users/me/usage/history", RequestsPerSecond: 2, Burst: 5},
    {Endpoint: "/api/v1/users/me/usage/insights", RequestsPerSecond: 1, Burst: 3},
}
```

### Encryption Service

```go
type EncryptionService struct {
    key []byte
}

func (es *EncryptionService) EncryptUsageData(data *UsageData) (string, error) {
    plaintext, _ := json.Marshal(data)

    block, err := aes.NewCipher(es.key)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

### GDPR Compliance

```go
type PrivacyService struct {
    db *gorm.DB
}

// Right to be forgotten
func (ps *PrivacyService) AnonymizeUserData(ctx context.Context, userID int64) error {
    hash := sha256.Sum256([]byte(fmt.Sprintf("user-%d", userID)))
    anonymousUsername := fmt.Sprintf("anon-%x", hash[:8])

    // Anonymize in PostgreSQL
    ps.db.Exec("UPDATE radius_accounting SET username = ? WHERE user_id = ?",
        anonymousUsername, userID)
    ps.db.Exec("UPDATE radius_online SET username = ? WHERE user_id = ?",
        anonymousUsername, userID)

    // Delete from Redis cache (all patterns)
    redisClient.Del(ctx, fmt.Sprintf("user:usage:*:%d", userID))
    redisClient.Del(ctx, fmt.Sprintf("session:active:*:%d", userID))

    auditLogger.LogPrivacyAction(ctx, "anonymize", userID, "User requested data deletion")

    return nil
}

// Data export
func (ps *PrivacyService) ExportUserData(ctx context.Context, userID int64) (*UserDataExport, error) {
    var usage []UsageRecord
    var sessions []SessionRecord

    ps.db.Where("user_id = ?", userID).Find(&usage)
    ps.db.Where("user_id = ?", userID).Find(&sessions)

    auditLogger.LogPrivacyAction(ctx, "export", userID, "User exported personal data")

    return &UserDataExport{
        Usage:    usage,
        Sessions: sessions,
        ExportedAt: time.Now(),
    }, nil
}
```

### Audit Logging

```go
type AuditLogger struct {
    db *gorm.DB
}

type AuditLog struct {
    UserID      int64     `gorm:"column:user_id"`
    TenantID    int64     `gorm:"column:tenant_id"`
    Role        string    `gorm:"column:role"`
    Resource    string    `gorm:"column:resource"`
    Action      string    `gorm:"column:action"`
    IPAddress   string    `gorm:"column:ip_address"`
    UserAgent   string    `gorm:"column:user_agent"`
    Timestamp   time.Time `gorm:"column:timestamp"`
}

func (al *AuditLogger) LogAccess(ctx context.Context, secCtx SecurityContext, resource string, action string) {
    audit := AuditLog{
        UserID:    secCtx.UserID,
        TenantID:  secCtx.TenantID,
        Role:      secCtx.Role,
        Resource:  resource,
        Action:    action,
        IPAddress: getClientIP(ctx),
        UserAgent: getUserAgent(ctx),
        Timestamp: time.Now(),
    }

    go al.db.Create(&audit)
}
```

---

## 8. Implementation Strategy & Migration Plan

### 5-Phase Implementation (10 weeks)

```
PHASE 1: Foundation (Weeks 1-2) ████████████░░░░░░░░░░░░░░░░░░░ 30%
  ✓ Set up Redis infrastructure
  ✓ Create PostgreSQL indexes and acct_status column
  ✓ Implement Redis usage tracker with write-through caching
  ✓ Set up monitoring & alerting
  ✓ Security audit & hardening

PHASE 2: Core Features (Weeks 3-5) ████░░░░░░░░░░░░░░░░░░░░░░░░ 15%
  ✓ User usage API endpoints
  ✓ Frontend status page components
  ✓ Cache layer implementation
  ✓ Role-based access control
  ✓ Audit logging system

PHASE 3: Analytics (Weeks 6-7) ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░   0%
  ✓ Predictive analytics engine
  ✓ Pattern detection service
  ✓ Insights API endpoints
  ✓ Historical data endpoints
  ✓ Anomaly detection

PHASE 4: Real-time (Weeks 8-9) ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░   0%
  ✓ WebSocket support
  ✓ Live activity feed
  ✓ Real-time session tracking
  ✓ Push notifications
  ✓ Mobile app API support

PHASE 5: Polish (Week 10) ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░   0%
  ✓ Performance optimization
  ✓ Load testing (100K concurrent users)
  ✓ Security penetration testing
  ✓ Documentation completion
  ✓ User training materials
```

### Feature Flags for Gradual Rollout

```go
var featureFlags = map[string]FeatureFlag{
    "usage_analytics": {
        Name:       "Usage Analytics Dashboard",
        Enabled:    true,
        Percentage: 10, // Start with 10% of users
        Whitelist:  []string{"user123", "user456"},
    },
    "predictive_insights": {
        Name:       "Predictive Usage Insights",
        Enabled:    false,
        Percentage: 0,
    },
    "realtime_sessions": {
        Name:       "Real-time Session Tracking",
        Enabled:    true,
        Percentage: 5, // Start with 5% of users
    },
}

func IsFeatureEnabled(feature string, userID string) bool {
    flag, exists := featureFlags[feature]
    if !exists || !flag.Enabled {
        return false
    }

    for _, whitelisted := range flag.Whitelist {
        if whitelisted == userID {
            return true
        }
    }

    for _, blacklisted := range flag.Blacklist {
        if blacklisted == userID {
            return false
        }
    }

    if flag.Percentage > 0 {
        hash := sha256.Sum256([]byte(userID + feature))
        rolloutValue := int(hash[0]) % 100
        return rolloutValue < flag.Percentage
    }

    return flag.Enabled
}
```

### Database Migration Strategy

```go
// cmd/migrate/main.go

func main() {
    fmt.Println("Starting migration...")

    // 1. Add acct_status column to radius_accounting
    fmt.Println("Adding acct_status column...")
    addAccountingStatusColumn()

    // 2. Create indexes concurrently (non-blocking)
    fmt.Println("Creating indexes...")
    createIndexesConcurrently()

    // 3. Set up table partitioning
    fmt.Println("Setting up partitioning...")
    setupPartitioning()

    // 4. Set up Redis infrastructure
    fmt.Println("Setting up Redis...")
    setupRedisInfrastructure()

    // 5. Backfill existing data to Redis
    fmt.Println("Backfilling historical data to Redis...")
    backfillHistoricalDataToRedis()

    // 6. Verify data integrity
    fmt.Println("Verifying data integrity...")
    verifyDataIntegrity()

    fmt.Println("Migration complete!")
}

func addAccountingStatusColumn() {
    // Add acct_status column if not exists
    db.Exec(`ALTER TABLE radius_accounting ADD COLUMN IF NOT EXISTS acct_status VARCHAR(20)`)

    // Backfill existing data
    db.Exec(`UPDATE radius_accounting SET acct_status = CASE
        WHEN acct_stop_time IS NULL OR acct_stop_time = '0001-01-01 00:00:00' THEN 'active'
        ELSE 'stopped'
    END`)

    // Create trigger for automatic status updates
    db.Exec(`CREATE OR REPLACE FUNCTION update_acct_status()
    RETURNS TRIGGER AS $$
    BEGIN
        IF NEW.acct_stop_time IS NULL OR NEW.acct_stop_time = '0001-01-01 00:00:00' THEN
            NEW.acct_status := 'active';
        ELSE
            NEW.acct_status := 'stopped';
        END IF;
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql`)

    db.Exec(`CREATE TRIGGER trigger_update_acct_status
    BEFORE INSERT OR UPDATE ON radius_accounting
    FOR EACH ROW EXECUTE FUNCTION update_acct_status()`)
}

func createIndexesConcurrently() {
    indexes := []string{
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_tenant_user_time ON radius_accounting(tenant_id, username, acct_start_time DESC) WHERE acct_status = 'stopped'",
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_data_usage ON radius_accounting(tenant_id, username, (acct_input_octets + acct_output_octets)) WHERE acct_status = 'stopped'",
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_active_sessions ON radius_accounting(tenant_id, username, acct_unique_id) WHERE acct_status = 'active'",
    }

    for _, indexSQL := range indexes {
        fmt.Println("Creating index:", indexSQL)
        db.Exec(indexSQL)
    }
}

func setupRedisInfrastructure() {
    // Configure Redis for enterprise usage analytics
    redisConf := &redis.Options{
        Addr:         "localhost:6379",
        PoolSize:     500,
        MinIdleConns: 50,
        MaxRetries:   3,
        DialTimeout:  50 * time.Millisecond,
        ReadTimeout:  100 * time.Millisecond,
        WriteTimeout: 100 * time.Millisecond,
        PoolTimeout:  4 * time.Second,
    }

    redisClient := redis.NewClient(redisConf)

    // Test connection
    if err := redisClient.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    fmt.Println("Redis infrastructure ready")
}

func backfillHistoricalDataToRedis() {
    // Backfill last 30 days of usage data to Redis
    var users []User
    db.Where("created_at > ?", time.Now().AddDate(0, 0, -30)).Find(&users)

    pipe := redisClient.Pipeline()
    for _, user := range users {
        // Calculate usage from PostgreSQL
        var timeUsed, dataUsed int64
        db.Model(&RadiusAccounting{}).
            Where("tenant_id = ? AND username = ? AND acct_status = ?", user.TenantID, user.Username, "stopped").
            Select("COALESCE(SUM(acct_session_time), 0)").
            Scan(&timeUsed)

        db.Model(&RadiusAccounting{}).
            Where("tenant_id = ? AND username = ? AND acct_status = ?", user.TenantID, user.Username, "stopped").
            Select("COALESCE(SUM(acct_input_octets + acct_output_octets), 0)").
            Scan(&dataUsed)

        // Set in Redis
        usageKey := fmt.Sprintf("user:usage:%d:%s", user.TenantID, user.Username)
        pipe.HSet(context.Background(), usageKey, map[string]interface{}{
            "time_quota_total":   user.TimeQuota,
            "time_quota_used":    timeUsed,
            "data_quota_total":   user.DataQuota,
            "data_quota_used":    dataUsed,
        })
        pipe.Expire(context.Background(), usageKey, 60*time.Second)
    }

    pipe.Exec(context.Background())
    fmt.Println("Backfilled Redis with historical data")
}
```

### Rollback Procedure

```bash
#!/bin/bash
# scripts/rollback_usage_analytics.sh

echo "Rolling back Usage Analytics deployment..."

# 1. Disable feature flags
redis-cli SET feature:usage_analytics false
redis-cli SET feature:predictive_insights false
redis-cli SET feature:realtime_sessions false

# 2. Stop new services
systemctl stop session-tracker
systemctl stop analytics-worker

# 3. Restore previous version
git checkout HEAD~1

# 4. Rebuild and restart
go build -o toughradius .
systemctl restart toughradius

# 5. Drop new indexes (optional - keep for performance)
psql -d toughradius -c "DROP INDEX IF EXISTS idx_accounting_tenant_user_time;"

# 6. Verify system health
curl -f http://localhost:1816/api/v1/health || exit 1

echo "Rollback complete!"
```

### Performance Testing

```go
func LoadTestUsageAPI(t *testing.T) {
    numUsers := 100000
    requestsPerUser := 10

    var wg sync.WaitGroup
    results := make(chan time.Duration, numUsers*requestsPerUser)

    startTime := time.Now()

    for i := 0; i < numUsers; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()

            for j := 0; j < requestsPerUser; j++ {
                reqStart := time.Now()

                resp, err := http.Get(fmt.Sprintf("http://localhost:1816/api/v1/users/me/usage?user_id=%d", userID))

                if err != nil {
                    t.Errorf("Request failed: %v", err)
                    continue
                }
                resp.Body.Close()

                reqDuration := time.Since(reqStart)
                results <- reqDuration
            }
        }(i)
    }

    wg.Wait()
    close(results)

    totalDuration := time.Since(startTime)

    // Analyze results
    var durations []time.Duration
    for d := range results {
        durations = append(durations, d)
    }

    sort.Slice(durations, func(i, j int) bool {
        return durations[i] < durations[j]
    })

    p50 := durations[len(durations)/2]
    p95 := durations[int(float64(len(durations))*0.95)]
    p99 := durations[int(float64(len(durations))*0.99)]

    fmt.Printf("Load Test Results:\n")
    fmt.Printf("Total Requests: %d\n", numUsers*requestsPerUser)
    fmt.Printf("Total Time: %v\n", totalDuration)
    fmt.Printf("P50: %v\n", p50)
    fmt.Printf("P95: %v\n", p95)
    fmt.Printf("P99: %v\n", p99)

    // Assertions
    assert.Less(t, p50, 10*time.Millisecond, "P50 should be <10ms")
    assert.Less(t, p95, 50*time.Millisecond, "P95 should be <50ms")
    assert.Less(t, p99, 100*time.Millisecond, "P99 should be <100ms")
}
```

### Monitoring & Alerting

```go
type MetricsCollector struct {
    prometheus *prometheus.Registry
}

func (mc *MetricsCollector) RegisterMetrics() {
    cacheHits := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"cache_level", "cache_type"},
    )

    responseTime := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_response_time_seconds",
            Help:    "API response time in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"endpoint", "method"},
    )

    activeSessions := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "active_sessions_total",
            Help: "Number of active sessions",
        },
        []string{"tenant_id"},
    )

    mc.prometheus.MustRegister(cacheHits)
    mc.prometheus.MustRegister(responseTime)
    mc.prometheus.MustRegister(activeSessions)
}

func (mc *MetricsCollector) SetupAlerts() {
    prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "cache_miss_rate_high",
        Help: "Alert when cache miss rate exceeds 30%",
    }, func() float64 {
        missRate := calculateCacheMissRate()
        if missRate > 0.3 {
            return 1
        }
        return 0
    })

    prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "slow_queries_detected",
        Help: "Alert when queries exceed 100ms",
    }, func() float64 {
        if hasSlowQueries() {
            return 1
        }
        return 0
    })
}
```

---

## Key Files to Create/Modify

### Backend (Go)

- `internal/service/session_tracker.go` - Session tracking service
- `internal/service/usage_insights.go` - Insights aggregation service
- `internal/analytics/predictive_engine.go` - Prediction models
- `internal/analytics/pattern_detector.go` - Pattern detection
- `internal/analytics/anomaly_detector.go` - Anomaly detection
- `internal/cache/multilevel_cache.go` - Multi-level caching
- `internal/middleware/permissions.go` - RBAC middleware
- `internal/middleware/ratelimit.go` - Rate limiting
- `internal/service/privacy_service.go` - GDPR compliance
- `internal/jobs/materialized_view_refresher.go` - MV refresh jobs
- `internal/websocket/session_hub.go` - WebSocket hub
- `internal/adminapi/usage_api.go` - Usage API endpoints
- `internal/adminapi/insights_api.go` - Insights API endpoints

### Frontend (TypeScript/React)

- `web/src/resources/users/UsageQuotaField.tsx` - Time quota display component
- `web/src/resources/users/DataQuotaField.tsx` - Data quota display component
- `web/src/resources/users/UsageHistoryChart.tsx` - Historical usage chart
- `web/src/resources/users/UsageInsights.tsx` - Insights display
- `web/src/resources/users/ActiveSessionsList.tsx` - Active sessions list

### Database (SQL)

- `migrations/20260327_create_materialized_views.sql` - MV definitions
- `migrations/20260327_create_usage_indexes.sql` - Performance indexes
- `migrations/20260327_setup_partitioning.sql` - Table partitioning

---

## Success Criteria

The implementation will be considered successful when:

1. **Performance**
   - ✅ P50 API response time <10ms for cached data
   - ✅ P95 API response time <50ms for Redis cache hits
   - ✅ P99 API response time <100ms for all requests
   - ✅ System handles 100,000 concurrent users without degradation

2. **Security**
   - ✅ All endpoints properly enforce RBAC
   - ✅ Tenant isolation working correctly (verified by penetration testing)
   - ✅ Audit logging captures all usage data access
   - ✅ Rate limiting prevents abuse

3. **Functionality**
   - ✅ Users see accurate remaining time quota (not time window)
   - ✅ Real-time active sessions display updates within 1 minute
   - ✅ Predictive insights have >80% accuracy
   - ✅ Historical data queries support custom date ranges

4. **Reliability**
   - ✅ 99.9% uptime for usage analytics endpoints
   - ✅ Graceful degradation when Redis is unavailable
   - ✅ Automatic failover for cache misses
   - ✅ Rollback procedure tested and documented

---

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|------------|------------|
| Redis downtime causes performance degradation | High | Medium | Multi-level cache with DB fallback, Redis HA with replication |
| Redis memory exhaustion at scale | High | Low | Memory planning with LRU eviction, monitoring and alerts |
| Cache stampede during high traffic | Medium | Medium | Randomized cache expiration, request coalescing |
| Prediction models inaccurate | Low | Medium | Ensemble multiple models, continuous validation |
| Cross-tenant data leakage | Critical | Low | Comprehensive RBAC + audit logging |
| Redis sync lag causes inconsistencies | Medium | Low | Write-through caching, background worker every 60s |

---

## Next Steps

1. ✅ Design document approved
2. ⏳ Dispatch spec-document-reviewer for validation
3. �<arg_value> Iterate on spec based on feedback
4. ⏳ User reviews and approves final spec
5. ⏳ Invoke writing-plans skill for implementation plan
6. ⏳ Begin Phase 1 implementation

---

**Document Status:** Draft - Ready for Review
**Last Updated:** 2026-03-27
**Version:** 1.0
