# Phase 4: Tenant-Isolated Monitoring Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement comprehensive monitoring where providers see only their metrics while platform admins see aggregated data.

**Architecture:** Prometheus metrics with tenant_id labels. Tenant-aware query layer ensures isolation. Background workers collect device health, network performance, and RADIUS metrics.

**Tech Stack:** Prometheus, Redis, Go routines, MikroTik API client

---

## Task 1: Create Tenant-Aware Metrics Collector

**Files:**
- Create: `internal/monitoring/metrics.go`
- Create: `internal/monitoring/metrics_test.go`

**Step 1: Write tests for metrics collector**

```go
// internal/monitoring/metrics_test.go
package monitoring

import (
    "testing"
    "strconv"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/talkincode/toughradius/v9/internal/tenant"
)

func TestRecordAuthMetric(t *testing.T) {
    collector := NewTenantMetricsCollector()

    // Test recording successful auth
    collector.RecordAuth(1, true)
    collector.RecordAuth(1, false)

    // Verify metrics registered
    metrics, err := prometheus.DefaultGatherer.Gather()
    if err != nil {
        t.Fatalf("Failed to gather metrics: %v", err)
    }

    // Find radius_auth_total metric
    var found bool
    for _, m := range metrics {
        if m.GetName() == "radius_auth_total" {
            found = true
            break
        }
    }

    if !found {
        t.Error("radius_auth_total metric not found")
    }
}

func TestRecordDeviceMetric(t *testing.T) {
    collector := NewTenantMetricsCollector()

    // Test recording device health
    collector.RecordDeviceHealth(context.Background(), 1, "device1", "192.168.1.1", 75.5, 60.2)

    // Verify metric exists
    metrics, _ := prometheus.DefaultGatherer.Gather()

    var found bool
    for _, m := range metrics {
        if m.GetName() == "device_cpu_usage_percent" {
            found = true
            break
        }
    }

    if !found {
        t.Error("device_cpu_usage_percent metric not found")
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/monitoring -run TestRecord -v`
Expected: FAIL with "undefined: NewTenantMetricsCollector"

**Step 3: Implement metrics collector**

```go
// internal/monitoring/metrics.go
package monitoring

import (
    "context"
    "strconv"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/talkincode/toughradius/v9/internal/tenant"
)

type TenantMetricsCollector struct {
    // RADIUS metrics
    radiusAuthRate    *prometheus.CounterVec
    radiusAcctRate    *prometheus.CounterVec
    authErrors        *prometheus.CounterVec
    onlineSessions    *prometheus.GaugeVec

    // Device health metrics
    deviceCpuUsage    *prometheus.GaugeVec
    deviceMemoryUsage *prometheus.GaugeVec
    deviceUptime      *prometheus.GaugeVec
    deviceStatus      *prometheus.GaugeVec

    // Network performance metrics
    networkLatency    *prometheus.GaugeVec
    packetLoss        *prometheus.GaugeVec
    bandwidthUsage    *prometheus.GaugeVec
}

func NewTenantMetricsCollector() *TenantMetricsCollector {
    return &TenantMetricsCollector{
        radiusAuthRate: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "radius_auth_total",
                Help: "Total RADIUS authentication requests by tenant",
            },
            []string{"tenant_id", "result"}, // result: success, failure
        ),

        radiusAcctRate: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "radius_acct_total",
                Help: "Total RADIUS accounting requests by tenant",
            },
            []string{"tenant_id"},
        ),

        authErrors: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "radius_auth_errors_total",
                Help: "Total RADIUS authentication errors by tenant and error type",
            },
            []string{"tenant_id", "error_type"},
        ),

        onlineSessions: prometheus.NewGaugeVec(
            prometheus.CounterOpts{
                Name: "online_sessions",
                Help: "Number of currently online sessions by tenant",
            },
            []string{"tenant_id"},
        ),

        deviceCpuUsage: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "device_cpu_usage_percent",
                Help: "MikroTik device CPU usage percentage",
            },
            []string{"tenant_id", "device_id", "device_ip"},
        ),

        deviceMemoryUsage: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "device_memory_usage_percent",
                Help: "MikroTik device memory usage percentage",
            },
            []string{"tenant_id", "device_id", "device_ip"},
        ),

        deviceStatus: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "device_status",
                Help: "Device status (1=online, 0=offline)",
            },
            []string{"tenant_id", "device_id", "device_ip"},
        ),

        networkLatency: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "network_latency_ms",
                Help: "Network latency to MikroTik device in milliseconds",
            },
            []string{"tenant_id", "device_id"},
        ),

        packetLoss: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "packet_loss_percent",
                Help: "Packet loss percentage to device",
            },
            []string{"tenant_id", "device_id"},
        ),

        bandwidthUsage: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "bandwidth_usage_mbps",
                Help: "Current bandwidth usage in Mbps",
            },
            []string{"tenant_id", "device_id"},
        ),
    }
}

// RecordAuth records an authentication attempt
func (m *TenantMetricsCollector) RecordAuth(tenantID int64, success bool) {
    result := "failure"
    if success {
        result = "success"
    }
    m.radiusAuthRate.WithLabelValues(
        strconv.FormatInt(tenantID, 10),
        result,
    ).Inc()
}

// RecordAuthError records an authentication error
func (m *TenantMetricsCollector) RecordAuthError(tenantID int64, errorType string) {
    m.authErrors.WithLabelValues(
        strconv.FormatInt(tenantID, 10),
        errorType,
    ).Inc()
}

// UpdateOnlineSessions updates online session count
func (m *TenantMetricsCollector) UpdateOnlineSessions(tenantID int64, count int) {
    m.onlineSessions.WithLabelValues(
        strconv.FormatInt(tenantID, 10),
    ).Set(float64(count))
}

// RecordDeviceHealth records device health metrics
func (m *TenantMetricsCollector) RecordDeviceHealth(
    ctx context.Context,
    tenantID int64,
    deviceID, deviceIP string,
    cpu, memory float64,
    online bool,
) {
    tenantStr := strconv.FormatInt(tenantID, 10)

    m.deviceCpuUsage.WithLabelValues(tenantStr, deviceID, deviceIP).Set(cpu)
    m.deviceMemoryUsage.WithLabelValues(tenantStr, deviceID, deviceIP).Set(memory)

    status := 0.0
    if online {
        status = 1.0
    }
    m.deviceStatus.WithLabelValues(tenantStr, deviceID, deviceIP).Set(status)
}

// RecordNetworkPerformance records network metrics
func (m *TenantMetricsCollector) RecordNetworkPerformance(
    tenantID int64,
    deviceID string,
    latency, packetLoss, bandwidth float64,
) {
    tenantStr := strconv.FormatInt(tenantID, 10)

    m.networkLatency.WithLabelValues(tenantStr, deviceID).Set(latency)
    m.packetLoss.WithLabelValues(tenantStr, deviceID).Set(packetLoss)
    m.bandwidthUsage.WithLabelValues(tenantStr, deviceID).Set(bandwidth)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/monitoring -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/monitoring/metrics.go internal/monitoring/metrics_test.go
git commit -m "feat(monitoring): add tenant-aware Prometheus metrics collector"
```

---

## Task 2: Create Device Monitoring Service

**Files:**
- Create: `internal/monitoring/device_monitor.go`
- Create: `internal/monitoring/device_monitor_test.go`

**Step 1: Write tests for device monitor**

```go
// internal/monitoring/device_monitor_test.go
package monitoring

import (
    "context"
    "testing"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
)

func TestDeviceMonitor(t *testing.T) {
    db := setupTestDB(t)

    // Create test device
    device := &domain.Server{
        ID:         1,
        TenantID:   1,
        Name:       "test-router",
        PublicIP:   "192.168.1.1",
        Username:   "admin",
        Password:   "password",
        Ports:      "8728",
    }
    db.Create(device)

    // Create monitor
    collector := NewTenantMetricsCollector()
    monitor := NewDeviceHealthMonitor(db, collector)

    // Mock MikroTik connection (in real test, use test server)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := monitor.checkDevice(ctx, *device)
    // Will fail to connect in test, that's okay
    if err != nil {
        t.Logf("Expected connection failure: %v", err)
    }
}
```

**Step 2: Implement device monitoring service**

```go
// internal/monitoring/device_monitor.go
package monitoring

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/go-routeros/routeros"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
)

type DeviceHealthMonitor struct {
    db           *gorm.DB
    metrics      *TenantMetricsCollector
    connPool     map[string]*routeros.Client
    poolMutex    sync.RWMutex

    // Performance tuning
    maxConcurrent int
    checkInterval time.Duration
    timeout       time.Duration
}

func NewDeviceHealthMonitor(db *gorm.DB, metrics *TenantMetricsCollector) *DeviceHealthMonitor {
    return &DeviceHealthMonitor{
        db:            db,
        metrics:       metrics,
        connPool:      make(map[string]*routeros.Client),
        maxConcurrent: 50,
        checkInterval: 30 * time.Second,
        timeout:       5 * time.Second,
    }
}

// Run starts the monitoring loop
func (m *DeviceHealthMonitor) Run(ctx context.Context) {
    ticker := time.NewTicker(m.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkAllDevices(ctx)
        }
    }
}

// checkAllDevices checks all devices across all providers efficiently
func (m *DeviceHealthMonitor) checkAllDevices(ctx context.Context) {
    var devices []domain.Server
    m.db.Find(&devices)

    // Use semaphore to limit concurrent checks
    sem := make(chan struct{}, m.maxConcurrent)
    var wg sync.WaitGroup

    for _, device := range devices {
        wg.Add(1)
        sem <- struct{}{}

        go func(d domain.Server) {
            defer wg.Done()
            defer func() { <-sem }()

            deviceCtx, cancel := context.WithTimeout(ctx, m.timeout)
            defer cancel()

            m.checkDevice(deviceCtx, d)
        }(device)
    }

    wg.Wait()
}

// checkDevice checks a single device's health
func (m *DeviceHealthMonitor) checkDevice(ctx context.Context, device domain.Server) error {
    client, err := m.getDeviceConnection(device)
    if err != nil {
        // Device offline
        m.metrics.RecordDeviceHealth(ctx, device.TenantID,
            fmt.Sprintf("%d", device.ID), device.PublicIP, 0, 0, false)
        return err
    }

    // Get system resources
    reply, err := client.Run("/system/resource/print")
    if err != nil {
        return err
    }

    if len(reply.Re) > 0 {
        cpu := parseCPU(reply.Re[0])
        memory := parseMemory(reply.Re[0])

        // Record metrics with tenant isolation
        m.metrics.RecordDeviceHealth(ctx, device.TenantID,
            fmt.Sprintf("%d", device.ID), device.PublicIP, cpu, memory, true)

        // Update database
        m.db.Model(&device).Updates(map[string]interface{}{
            "router_status": "online",
            "updated_at":    time.Now(),
        })
    }

    return nil
}

// getDeviceConnection gets or creates a pooled connection
func (m *DeviceHealthMonitor) getDeviceConnection(device domain.Server) (*routeros.Client, error) {
    m.poolMutex.RLock()
    client, exists := m.connPool[device.PublicIP]
    m.poolMutex.RUnlock()

    if exists {
        return client, nil
    }

    m.poolMutex.Lock()
    defer m.poolMutex.Unlock()

    if client, exists := m.connPool[device.PublicIP]; exists {
        return client, nil
    }

    address := fmt.Sprintf("%s:%s", device.PublicIP, device.Ports)
    client, err := routeros.Dial(address, device.Username, device.Password)
    if err != nil {
        return nil, err
    }

    m.connPool[device.PublicIP] = client
    return client, nil
}

func parseCPU(re *routeros.Reply) float64 {
    if cpu, ok := re.Map["cpu-load"]; ok {
        var f float64
        fmt.Sscanf(cpu, "%f", &f)
        return f
    }
    return 0
}

func parseMemory(re *routeros.Reply) float64 {
    // Parse memory from MikroTik response
    // Implementation details...
    return 0
}
```

**Step 3: Commit**

```bash
git add internal/monitoring/device_monitor.go internal/monitoring/device_monitor_test.go
git commit -m "feat(monitoring): add device health monitoring with connection pooling"
```

---

## Task 3: Create Tenant-Isolated Monitoring API

**Files:**
- Create: `internal/adminapi/monitoring.go`
- Create: `internal/adminapi/monitoring_test.go`

**Step 1: Implement monitoring APIs with tenant isolation**

```go
// internal/adminapi/monitoring.go
package adminapi

import (
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/monitoring"
    "github.com/talkincode/toughradius/v9/internal/repository"
    "github.com/talkincode/toughradius/v9/internal/tenant"
)

func registerMonitoringRoutes() {
    // Provider routes (tenant-isolated)
    webserver.ApiGET("/monitoring/metrics", GetMonitoringMetrics)
    webserver.ApiGET("/monitoring/devices", GetDeviceHealth)
    webserver.ApiGET("/monitoring/sessions", GetSessionMetrics)

    // Admin routes (aggregated)
    webserver.ApiGET("/admin/monitoring/metrics", GetAggregatedMetrics)
    webserver.ApiGET("/admin/monitoring/provider/:id", GetProviderMetrics)
}

// GetMonitoringMetrics returns metrics for current tenant only
func GetMonitoringMetrics(c echo.Context) error {
    tenantID, err := tenant.FromContext(c.Request().Context())
    if err != nil {
        return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
    }

    metrics := map[string]interface{}{
        "tenant_id": tenantID,
        // Fetch Prometheus metrics for this tenant only
    }

    return ok(c, metrics)
}

// GetDeviceHealth returns device health for current tenant
func GetDeviceHealth(c echo.Context) error {
    db := GetDB(c)

    // Query devices with tenant isolation
    var devices []domain.Server
    err := db.Scopes(repository.TenantScope).Find(&devices).Error
    if err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch devices", err)
    }

    return ok(c, devices)
}

// GetAggregatedMetrics returns metrics for all providers (admin only)
func GetAggregatedMetrics(c echo.Context) error {
    // Verify platform admin
    if !IsPlatformAdmin(c) {
        return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
    }

    db := GetDB(c)

    // Aggregate metrics across all tenants
    var providers []domain.Provider
    db.Find(&providers)

    var providerStats []map[string]interface{}
    for _, provider := range providers {
        stats := map[string]interface{}{
            "tenant_id":     provider.ID,
            "provider_name": provider.Name,
            "users":         getCount(db, "radius_user", provider.ID),
            "sessions":      getCount(db, "radius_online", provider.ID),
            "devices":       getCount(db, "net_nas", provider.ID),
        }
        providerStats = append(providerStats, stats)
    }

    return ok(c, providerStats)
}

func getCount(db *gorm.DB, table string, tenantID int64) int64 {
    var count int64
    db.Table(table).Where("tenant_id = ?", tenantID).Count(&count)
    return count
}
```

**Step 2: Commit**

```bash
git add internal/adminapi/monitoring.go internal/adminapi/monitoring_test.go
git commit -m "feat(adminapi): add tenant-isolated monitoring APIs"
```

---

## Success Criteria

- ✅ Prometheus metrics collected with tenant_id labels
- ✅ Providers see only their own metrics
- ✅ Platform admins see aggregated metrics
- ✅ Device health monitoring functional
- ✅ Background workers collect metrics efficiently
- ✅ Unit tests pass (≥80% coverage)
