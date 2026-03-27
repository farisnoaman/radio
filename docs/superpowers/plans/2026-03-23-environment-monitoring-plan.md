# Environment Monitoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement environment monitoring for RouterOS devices with threshold-based alerting and notifications (dashboard, email, webhook).

**Architecture:** RouterOS API polling service that collects temperature, power, voltage, signal strength, fan status. Alert engine evaluates thresholds and triggers notifications. Uses existing device communication stack.

**Tech Stack:** Go, GORM, RouterOS API, existing notification infrastructure

---

### Task 1: Database Migration

**Files:**
- Create: `cmd/migrate/migrations/007_add_environment_monitoring_tables.sql`

- [ ] **Step 1: Create migration file**

```sql
-- Migration 007: Add Environment Monitoring Tables

-- Environment Metrics Table
CREATE TABLE environment_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    nas_id BIGINT NOT NULL,
    nas_name VARCHAR(128) NOT NULL DEFAULT '',
    metric_type VARCHAR(32) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(16) NOT NULL,
    severity VARCHAR(16) NOT NULL DEFAULT 'normal',
    collected_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Environment Alerts Table
CREATE TABLE environment_alerts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    nas_id BIGINT NOT NULL,
    metric_type VARCHAR(32) NOT NULL,
    threshold_type VARCHAR(8) NOT NULL,
    threshold_value DOUBLE PRECISION NOT NULL,
    alert_value DOUBLE PRECISION NOT NULL,
    severity VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'firing',
    notify_status VARCHAR(16) NOT NULL DEFAULT 'pending',
    fired_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    acknowledged_by VARCHAR(64),
    acknowledged_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Alert Configurations Table
CREATE TABLE alert_configs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    nas_id BIGINT NOT NULL,
    metric_type VARCHAR(32) NOT NULL,
    threshold_type VARCHAR(8) NOT NULL,
    threshold_value DOUBLE PRECISION NOT NULL,
    severity VARCHAR(16) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    notify_email BOOLEAN NOT NULL DEFAULT true,
    notify_webhook BOOLEAN NOT NULL DEFAULT false,
    webhook_url VARCHAR(2048) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_env_metrics_nas_time ON environment_metrics(nas_id, collected_at);
CREATE INDEX idx_env_metrics_tenant_time ON environment_metrics(tenant_id, metric_type, collected_at);
CREATE INDEX idx_env_alerts_status ON environment_alerts(status, fired_at);
CREATE INDEX idx_alert_configs_nas_metric ON alert_configs(nas_id, metric_type);
```

- [ ] **Step 2: Run migration**

```bash
cd cmd/migrate && go run main.go
# Or use: migrate -path=cmd/migrate/migrations -database="postgres://..." up
```

- [ ] **Step 3: Commit**

```bash
git add cmd/migrate/migrations/007_add_environment_monitoring_tables.sql
git commit -m "feat(migration): add environment monitoring tables"
```

---

### Task 2: Domain Models

**Files:**
- Create: `internal/domain/environment.go`

- [ ] **Step 1: Write domain models**

```go
package domain

import (
	"time"
)

const (
	MetricTypeTemperature    = "temperature"
	MetricTypePower         = "power"
	MetricTypeVoltage       = "voltage"
	MetricTypeSignalStrength = "signal_strength"
	MetricTypeFanSpeed     = "fan_speed"

	SeverityNormal   = "normal"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"

	ThresholdTypeMin = "min"
	ThresholdTypeMax = "max"

	AlertStatusFiring       = "firing"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusResolved     = "resolved"

	NotifyStatusPending = "pending"
	NotifyStatusSent    = "sent"
	NotifyStatusFailed  = "failed"
)

type EnvironmentMetric struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TenantID   string    `gorm:"size:36;not null;index" json:"tenant_id"`
	NasID      uint      `gorm:"not null;index" json:"nas_id"`
	NasName    string    `gorm:"size:128;default:''" json:"nas_name"`
	MetricType string    `gorm:"size:32;not null" json:"metric_type"`
	Value      float64   `gorm:"not null" json:"value"`
	Unit       string    `gorm:"size:16;not null" json:"unit"`
	Severity   string    `gorm:"size:16;not null;default:'normal'" json:"severity"`
	CollectedAt time.Time `gorm:"not null" json:"collected_at"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
}

func (EnvironmentMetric) TableName() string {
	return "environment_metrics"
}

type EnvironmentAlert struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	TenantID       string     `gorm:"size:36;not null;index" json:"tenant_id"`
	NasID          uint       `gorm:"not null;index" json:"nas_id"`
	MetricType     string     `gorm:"size:32;not null" json:"metric_type"`
	ThresholdType  string     `gorm:"size:8;not null" json:"threshold_type"`
	ThresholdValue float64    `gorm:"not null" json:"threshold_value"`
	AlertValue     float64    `gorm:"not null" json:"alert_value"`
	Severity       string     `gorm:"size:16;not null" json:"severity"`
	Status         string     `gorm:"size:16;not null;default:'firing'" json:"status"`
	NotifyStatus   string     `gorm:"size:16;not null;default:'pending'" json:"notify_status"`
	FiredAt        time.Time  `gorm:"not null" json:"fired_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	AcknowledgedBy *string    `gorm:"size:64" json:"acknowledged_by"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	CreatedAt      time.Time  `gorm:"not null" json:"created_at"`
}

func (EnvironmentAlert) TableName() string {
	return "environment_alerts"
}

type AlertConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"size:36;not null;index" json:"tenant_id"`
	NasID         uint      `gorm:"not null;index" json:"nas_id"`
	MetricType    string    `gorm:"size:32;not null" json:"metric_type"`
	ThresholdType string    `gorm:"size:8;not null" json:"threshold_type"`
	ThresholdValue float64  `gorm:"not null" json:"threshold_value"`
	Severity      string    `gorm:"size:16;not null" json:"severity"`
	Enabled       bool      `gorm:"not null;default:true" json:"enabled"`
	NotifyEmail   bool      `gorm:"not null;default:true" json:"notify_email"`
	NotifyWebhook bool      `gorm:"not null;default:false" json:"notify_webhook"`
	WebhookURL    string    `gorm:"size:2048;default:''" json:"webhook_url"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`
}

func (AlertConfig) TableName() string {
	return "alert_configs"
}
```

- [ ] **Step 2: Run go generate**

```bash
cd internal/domain && go generate
```

- [ ] **Step 3: Commit**

```bash
git add internal/domain/environment.go
git commit -m "feat(domain): add environment monitoring models"
```

---

### Task 3: Collector Service (RouterOS API)

**Files:**
- Create: `internal/device/env_monitor.go`

- [ ] **Step 1: Write collector service**

```go
package device

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/routeros"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EnvCollector struct {
	db     *gorm.DB
	client *routeros.Client
}

func NewEnvCollector(db *gorm.DB) *EnvCollector {
	return &EnvCollector{db: db}
}

func (c *EnvCollector) CollectAllDevices(ctx context.Context) error {
	var nasList []domain.NetNas
	if err := c.db.Find(&nasList).Error; err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, nas := range nasList {
		wg.Add(1)
		go func(nas domain.NetNas) {
			defer wg.Done()
			if err := c.CollectDevice(ctx, &nas); err != nil {
				zap.S().Errorw("Failed to collect env metrics", "nas_id", nas.ID, "error", err)
			}
		}(nas)
	}
	wg.Wait()
	return nil
}

func (c *EnvCollector) CollectDevice(ctx context.Context, nas *domain.NetNas) error {
	client, err := routeros.Dial(ctx, nas.Ipaddr, nas.Username, nas.Password)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", nas.Ipaddr, err)
	}
	defer client.Close()

	now := time.Now()
	var metrics []domain.EnvironmentMetric

	// Temperature
	if temp, err := c.getTemperature(ctx, client); err == nil && temp > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:   nas.TenantID,
			NasID:      uint(nas.ID),
			NasName:    nas.Name,
			MetricType: domain.MetricTypeTemperature,
			Value:      temp,
			Unit:       "C",
			Severity:   c.calculateSeverity(domain.MetricTypeTemperature, temp),
			CollectedAt: now,
			CreatedAt:  now,
		})
	}

	// Power
	if power, err := c.getPower(ctx, client); err == nil && power > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:   nas.TenantID,
			NasID:      nas.ID,
			NasName:    nas.Ipaddr,
			MetricType: domain.MetricTypePower,
			Value:      power,
			Unit:       "W",
			Severity:   c.calculateSeverity(domain.MetricTypePower, power),
			CollectedAt: now,
			CreatedAt:  now,
		})
	}

	// Voltage
	if voltage, err := c.getVoltage(ctx, client); err == nil && voltage > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:   nas.TenantID,
			NasID:      uint(nas.ID),
			NasName:    nas.Name,
			MetricType: domain.MetricTypePower,
			Value:      power,
			Unit:       "W",
			Severity:   c.calculateSeverity(domain.MetricTypePower, power),
			CollectedAt: now,
			CreatedAt:  now,
		})

		// Voltage
		if voltage, err := c.getVoltage(ctx, client); err == nil && voltage > 0 {
			metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:   nas.TenantID,
			NasID:      uint(nas.ID),
			NasName:    nas.Name,
			MetricType: domain.MetricTypeVoltage,
			Value:      voltage,
			Unit:       "V",
			Severity:   c.calculateSeverity(domain.MetricTypeVoltage, voltage),
			CollectedAt: now,
			CreatedAt:  now,
		})
	}

	// Fan Speed
	if fan, err := c.getFanSpeed(ctx, client); err == nil && fan > 0 {
		metrics = append(metrics, domain.EnvironmentMetric{
			TenantID:   nas.TenantID,
			NasID:      uint(nas.ID),
			NasName:    nas.Name,
			MetricType: domain.MetricTypeFanSpeed,
			Value:      fan,
			Unit:       "RPM",
			Severity:   c.calculateSeverity(domain.MetricTypeFanSpeed, fan),
			CollectedAt: now,
			CreatedAt:  now,
		})
	}

	if len(metrics) > 0 {
		return c.db.Create(&metrics).Error
	}
	return nil
}

func (c *EnvCollector) getTemperature(ctx context.Context, client *routeros.Client) (float64, error) {
	// Try /system/health/print first
	resp, err := client.Run(ctx, []string{"/system/health/print"})
	if err == nil && len(resp) > 0 {
		for _, re := range resp {
			if v, ok := re.Get("cpu-temperature"); ok {
				return v.Float64()
			}
		}
	}
	// Fallback not available in RouterOS
	return 0, fmt.Errorf("temperature not available")
}

func (c *EnvCollector) getPower(ctx context.Context, client *routeros.Client) (float64, error) {
	resp, err := client.Run(ctx, []string{"/system/health/print"})
	if err != nil {
		return 0, err
	}
	for _, re := range resp {
		if v, ok := re.Get("power-consumption"); ok {
			return v.Float64()
		}
	}
	return 0, fmt.Errorf("power not available")
}

func (c *EnvCollector) getVoltage(ctx context.Context, client *routeros.Client) (float64, error) {
	resp, err := client.Run(ctx, []string{"/system/health/print"})
	if err != nil {
		return 0, err
	}
	for _, re := range resp {
		if v, ok := re.Get("voltage"); ok {
			return v.Float64()
		}
	}
	return 0, fmt.Errorf("voltage not available")
}

func (c *EnvCollector) getFanSpeed(ctx context.Context, client *routeros.Client) (float64, error) {
	resp, err := client.Run(ctx, []string{"/system/health/print"})
	if err != nil {
		return 0, err
	}
	for _, re := range resp {
		if v, ok := re.Get("fan1-speed"); ok {
			return v.Float64()
		}
	}
	return 0, fmt.Errorf("fan speed not available")
}

func (c *EnvCollector) calculateSeverity(metricType string, value float64) string {
	// Default thresholds - will be overridden by AlertConfig
	defaults := map[string]struct {
		warningMin, warningMax float64
		criticalMin, criticalMax float64
	}{
		domain.MetricTypeTemperature: {warningMax: 70, criticalMin: 80},
		domain.MetricTypePower:       {warningMax: 100, criticalMin: 150},
		domain.MetricTypeVoltage:     {warningMin: 200, warningMax: 250, criticalMin: 180, criticalMax: 260},
		domain.MetricTypeFanSpeed:    {warningMin: 1000},
	}

	if cfg, ok := defaults[metricType]; ok {
		// Check critical first (more severe)
		if value < cfg.criticalMin || (cfg.criticalMax > 0 && value > cfg.criticalMax) {
			return domain.SeverityCritical
		}
		// Then warning
		if value < cfg.warningMin || value > cfg.warningMax {
			return domain.SeverityWarning
		}
	}
	return domain.SeverityNormal
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/device/env_monitor.go
git commit -m "feat(device): add environment collector service"
```

---

### Task 4: Alert Engine

**Files:**
- Create: `internal/device/alert_engine.go`

- [ ] **Step 1: Write alert engine**

```go
package device

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AlertEngine struct {
	db           *gorm.DB
	notifier     *Notifier
	cooldownMins int
}

func NewAlertEngine(db *gorm.DB, notifier *Notifier) *AlertEngine {
	return &AlertEngine{
		db:           db,
		notifier:     notifier,
		cooldownMins: 5,
	}
}

func (e *AlertEngine) ProcessMetrics(ctx context.Context, metrics []domain.EnvironmentMetric) error {
	for _, metric := range metrics {
		if err := e.evaluateThreshold(ctx, metric); err != nil {
			zap.S().Errorw("Failed to evaluate threshold", "metric", metric.MetricType, "error", err)
		}
	}
	return nil
}

func (e *AlertEngine) evaluateThreshold(ctx context.Context, metric domain.EnvironmentMetric) error {
	var configs []domain.AlertConfig
	if err := e.db.Where("nas_id = ? AND metric_type = ? AND enabled = ?", 
		metric.NasID, metric.MetricType, true).Find(&configs).Error; err != nil {
		return err
	}

	for _, cfg := range configs {
		triggered := false
		alertValue := metric.Value

		if cfg.ThresholdType == domain.ThresholdTypeMax && metric.Value > cfg.ThresholdValue {
			triggered = true
		} else if cfg.ThresholdType == domain.ThresholdTypeMin && metric.Value < cfg.ThresholdValue {
			triggered = true
		}

		if triggered {
			if err := e.createOrUpdateAlert(ctx, metric, cfg, alertValue); err != nil {
				return err
			}
		} else {
			// Auto-resolve existing firing alerts
			e.resolveAlert(ctx, metric.NasID, metric.MetricType)
		}
	}
	return nil
}

func (e *AlertEngine) createOrUpdateAlert(ctx context.Context, metric domain.EnvironmentMetric, cfg domain.AlertConfig, alertValue float64) error {
	var existing domain.EnvironmentAlert
	err := e.db.Where("nas_id = ? AND metric_type = ? AND status = ?", 
		metric.NasID, metric.MetricType, domain.AlertStatusFiring).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		alert := domain.EnvironmentAlert{
			TenantID:       metric.TenantID,
			NasID:          metric.NasID,
			MetricType:     metric.MetricType,
			ThresholdType:  cfg.ThresholdType,
			ThresholdValue: cfg.ThresholdValue,
			AlertValue:     alertValue,
			Severity:       cfg.Severity,
			Status:         domain.AlertStatusFiring,
			NotifyStatus:   domain.NotifyStatusPending,
			FiredAt:        time.Now(),
			CreatedAt:      time.Now(),
		}
		if err := e.db.Create(&alert).Error; err != nil {
			return err
		}
		// Send notifications
		go e.notifier.SendAlertNotifications(ctx, alert, cfg)
	} else if err == nil {
		// Update existing with new value
		existing.AlertValue = alertValue
		e.db.Save(&existing)
	}
	return nil
}

func (e *AlertEngine) resolveAlert(ctx context.Context, nasID uint, metricType string) error {
	now := time.Now()
	return e.db.Model(&domain.EnvironmentAlert{}).
		Where("nas_id = ? AND metric_type = ? AND status = ?", nasID, metricType, domain.AlertStatusFiring).
		Updates(map[string]interface{}{
			"status":       domain.AlertStatusResolved,
			"resolved_at":  now,
		}).Error
}

func (e *AlertEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := e.processAllMetrics(ctx); err != nil {
				zap.S().Errorw("Alert engine error", "error", err)
			}
		}
	}
}

func (e *AlertEngine) processAllMetrics(ctx context.Context) error {
	var latestMetrics []domain.EnvironmentMetric
	subQuery := e.db.Model(&domain.EnvironmentMetric{}).
		Select("nas_id, metric_type, MAX(collected_at) as max_collected").
		Group("nas_id, metric_type")

	if err := e.db.Table("environment_metrics", func(db *gorm.DB) *gorm.DB {
		return db.Raw(`
			SELECT em.* FROM environment_metrics em
			INNER JOIN (?) tm ON em.nas_id = tm.nas_id AND em.metric_type = tm.metric_type AND em.collected_at = tm.max_collected
		`, subQuery)
	}).Find(&latestMetrics).Error; err != nil {
		return err
	}

	return e.ProcessMetrics(ctx, latestMetrics)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/device/alert_engine.go
git commit -m "feat(device): add alert engine"
```

---

### Task 5: Notification Service

**Files:**
- Create: `internal/device/notifier.go`

- [ ] **Step 1: Write notification service**

```go
package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Notifier struct {
	db        *gorm.DB
	emailSvc  *service.SMTPEmailProvider
}

func NewNotifier(db *gorm.DB, emailSvc *service.SMTPEmailProvider) *Notifier {
	return &Notifier{
		db:       db,
		emailSvc: emailSvc,
	}
}

func (n *Notifier) SendAlertNotifications(ctx context.Context, alert domain.EnvironmentAlert, cfg domain.AlertConfig) {
	if cfg.NotifyEmail {
		n.sendEmail(ctx, alert, cfg)
	}
	if cfg.NotifyWebhook && cfg.WebhookURL != "" {
		n.sendWebhook(ctx, alert, cfg)
	}

	// Update notify status
	status := domain.NotifyStatusSent
	if err := n.db.Model(&alert).Update("notify_status", status).Error; err != nil {
		zap.S().Errorw("Failed to update notify status", "error", err)
	}
}

func (n *Notifier) sendEmail(ctx context.Context, alert domain.EnvironmentAlert, cfg domain.AlertConfig) {
	// Get NAS name
	var nas struct{ Name string }
	n.db.Model(&domain.NetNas{}).Where("id = ?", alert.NasID).First(&nas)

	subject := fmt.Sprintf("[%s] Device Alert: %s", alert.Severity, alert.MetricType)
	body := fmt.Sprintf(`
Device Environment Alert

Metric: %s
Device: %s (%d)
Current Value: %.2f
Threshold: %s %.2f
Severity: %s
Time: %s

Please take action.
`, alert.MetricType, nas.Name, alert.NasID, alert.AlertValue, 
		alert.ThresholdType, alert.ThresholdValue, alert.Severity, alert.FiredAt.Format(time.RFC3339))

	if n.emailSvc != nil {
		if err := n.emailSvc.SendEmail("admin@provider.local", subject, body); err != nil {
			zap.S().Errorw("Failed to send alert email", "error", err)
		}
	}
}

func (n *Notifier) sendWebhook(ctx context.Context, alert domain.EnvironmentAlert, cfg domain.AlertConfig) {
	payload := map[string]interface{}{
		"alert_id":        alert.ID,
		"metric_type":     alert.MetricType,
		"nas_id":          alert.NasID,
		"alert_value":     alert.AlertValue,
		"threshold":       alert.ThresholdValue,
		"threshold_type":  alert.ThresholdType,
		"severity":        alert.Severity,
		"fired_at":        alert.FiredAt,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", cfg.WebhookURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Errorw("Webhook failed", "url", cfg.WebhookURL, "error", err)
		n.db.Model(&domain.EnvironmentAlert{}).Where("id = ?", alert.ID).Update("notify_status", domain.NotifyStatusFailed)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		zap.S().Errorw("Webhook returned error", "status", resp.StatusCode)
		n.db.Model(&domain.EnvironmentAlert{}).Where("id = ?", alert.ID).Update("notify_status", domain.NotifyStatusFailed)
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/device/notifier.go
git commit -m "feat(device): add notification service"
```

---

### Task 6: Admin API Endpoints

**Files:**
- Create: `internal/adminapi/env_monitoring.go`

- [ ] **Step 1: Write API handlers**

```go
package adminapi

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

func EnvMetricHandlers(db *gorm.DB, webserver *echo.Echo) {
	group := webserver.Group("/api/v1/network/nas/:id")

	group.GET("/metrics", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var metrics []domain.EnvironmentMetric
		if err := db.Where("nas_id = ?", nasID).
			Order("collected_at DESC").Limit(100).Find(&metrics).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"metrics": metrics,
			"count":   len(metrics),
		})
	})

	group.GET("/metrics/history", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))
		metricType := c.QueryParam("type")
		days := 7
		if d := c.QueryParam("days"); d != "" {
			// parse days
		}

		since := time.Now().AddDate(0, 0, -days)

		var metrics []domain.EnvironmentMetric
		query := db.Where("nas_id = ? AND collected_at > ?", nasID, since)
		if metricType != "" {
			query = query.Where("metric_type = ?", metricType)
		}
		if err := query.Order("collected_at DESC").Find(&metrics).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"metrics": metrics,
			"count":   len(metrics),
		})
	})

	group.GET("/alerts", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var alerts []domain.EnvironmentAlert
		if err := db.Where("nas_id = ?", nasID).
			Order("fired_at DESC").Limit(100).Find(&alerts).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"alerts": alerts,
			"count":  len(alerts),
		})
	})

	group.GET("/alerts/config", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var configs []domain.AlertConfig
		if err := db.Where("nas_id = ?", nasID).Find(&configs).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"configs": configs,
			"count":   len(configs),
		})
	})

	group.PUT("/alerts/config", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))

		var req struct {
			MetricType    string  `json:"metric_type"`
			ThresholdType string  `json:"threshold_type"`
			ThresholdValue float64 `json:"threshold_value"`
			Severity      string  `json:"severity"`
			Enabled       bool    `json:"enabled"`
			NotifyEmail   bool    `json:"notify_email"`
			NotifyWebhook bool    `json:"notify_webhook"`
			WebhookURL    string  `json:"webhook_url"`
		}
		if err := c.Bind(&req); err != nil {
			return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		}

		// Upsert config
		var cfg domain.AlertConfig
		err := db.Where("nas_id = ? AND metric_type = ?", nasID, req.MetricType).First(&cfg).Error
		if err == gorm.ErrRecordNotFound {
			cfg = domain.AlertConfig{
				TenantID:      c.Get("tenant_id").(string),
				NasID:         nasID,
				MetricType:    req.MetricType,
				ThresholdType: req.ThresholdType,
				ThresholdValue: req.ThresholdValue,
				Severity:      req.Severity,
				Enabled:       req.Enabled,
				NotifyEmail:   req.NotifyEmail,
				NotifyWebhook: req.NotifyWebhook,
				WebhookURL:    req.WebhookURL,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := db.Create(&cfg).Error; err != nil {
				return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
			}
			return c.JSON(http.StatusCreated, map[string]interface{}{"config": cfg})
		}

		cfg.ThresholdType = req.ThresholdType
		cfg.ThresholdValue = req.ThresholdValue
		cfg.Severity = req.Severity
		cfg.Enabled = req.Enabled
		cfg.NotifyEmail = req.NotifyEmail
		cfg.NotifyWebhook = req.NotifyWebhook
		cfg.WebhookURL = req.WebhookURL
		cfg.UpdatedAt = time.Now()

		if err := db.Save(&cfg).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"config": cfg})
	})

	group.POST("/alerts/:alertId/ack", func(c echo.Context) error {
		nasID := uint(c.Get("nas_id").(float64))
		alertID := uint(c.Get("alert_id").(float64))
		username := c.Get("username").(string)

		var alert domain.EnvironmentAlert
		if err := db.Where("id = ? AND nas_id = ?", alertID, nasID).First(&alert).Error; err != nil {
			return fail(c, http.StatusNotFound, "NOT_FOUND", "Alert not found")
		}

		alert.Status = domain.AlertStatusAcknowledged
		alert.AcknowledgedBy = &username
		now := time.Now()
		alert.AcknowledgedAt = &now

		if err := db.Save(&alert).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"alert": alert})
	})
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/adminapi/env_monitoring.go
git commit -m "feat(adminapi): add environment monitoring endpoints"
```

---

### Task 7: Dashboard Health Endpoint

**Files:**
- Modify: `internal/adminapi/dashboard.go` (or create separate)

- [ ] **Step 1: Add health overview endpoint**

```go
// Add to existing dashboard or new file
func GetEnvHealthOverview(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tenantID := c.Get("tenant_id").(string)

		type HealthSummary struct {
			TotalDevices   int `json:"total_devices"`
			OnlineDevices  int `json:"online_devices"`
			WarningAlerts int `json:"warning_alerts"`
			CriticalAlerts int `json:"critical_alerts"`
		}

		var summary HealthSummary

		// Count NAS devices
		db.Model(&domain.NetNas{}).Where("tenant_id = ?", tenantID).Count(&summary.TotalDevices)

		// Count firing alerts by severity
		var alertCounts []struct {
			Severity string
			Count    int
		}
		db.Model(&domain.EnvironmentAlert{}).
			Where("tenant_id = ? AND status = ?", tenantID, domain.AlertStatusFiring).
			Group("severity").
			Select("severity, COUNT(*) as count").
			Scan(&alertCounts)

		for _, ac := range alertCounts {
			if ac.Severity == domain.SeverityWarning {
				summary.WarningAlerts = ac.Count
			} else if ac.Severity == domain.SeverityCritical {
				summary.CriticalAlerts = ac.Count
			}
		}

		// Get latest metric status per device
		type LatestMetric struct {
			NasID   uint
			Severity string
		}
		var latest []LatestMetric
		subQuery := db.Model(&domain.EnvironmentMetric{}).
			Select("nas_id, MAX(collected_at) as max_collected").
			Group("nas_id")
		
		db.Raw(`
			SELECT em.nas_id, em.severity 
			FROM environment_metrics em
			INNER JOIN (?) lm ON em.nas_id = lm.nas_id AND em.collected_at = lm.max_collected
		`, subQuery).Find(&latest)

		summary.OnlineDevices = len(latest)

		return c.JSON(http.StatusOK, summary)
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/adminapi/dashboard.go
git commit -m "feat(adminapi): add environment health overview"
```

---

### Task 8: Service Integration (Main)

**Files:**
- Modify: `main.go` or service initialization

- [ ] **Step 1: Initialize services**

```go
// In main.go after db initialization
envCollector := device.NewEnvCollector(db)
notifier := device.NewNotifier(db, smtpHost, smtpPort, fromEmail)
alertEngine := device.NewAlertEngine(db, notifier)

// Start background tasks
go envCollector.CollectAllDevices(context.Background())
go alertEngine.Run(context.Background())
```

- [ ] **Step 2: Commit**

```bash
git add main.go
git commit -m "feat: integrate environment monitoring services"
```

---

### Task 9: Tests

**Files:**
- Create: `internal/device/env_monitor_test.go`

- [ ] **Step 1: Write unit tests**

```go
package device

import (
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestEnvCollector_CalculateSeverity(t *testing.T) {
	collector := &EnvCollector{}

	tests := []struct {
		name       string
		metricType string
		value      float64
		expected   string
	}{
		{"normal temp", domain.MetricTypeTemperature, 50, domain.SeverityNormal},
		{"warning temp", domain.MetricTypeTemperature, 75, domain.SeverityWarning},
		{"critical temp", domain.MetricTypeTemperature, 85, domain.SeverityCritical},
		{"normal power", domain.MetricTypePower, 50, domain.SeverityNormal},
		{"warning power", domain.MetricTypePower, 120, domain.SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.calculateSeverity(tt.metricType, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd internal/device && go test -v -run TestEnvCollector
```

- [ ] **Step 3: Commit**

```bash
git add internal/device/env_monitor_test.go
git commit -m "test(device): add environment monitoring tests"
```
