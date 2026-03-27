# Environment Monitoring Design

## Overview

Implement environment monitoring for network devices to track temperature, power, voltage, signal strength, and fan status. Supports threshold-based alerting with dashboard, email, and webhook notifications.

## Scope

- **Phase 1**: RouterOS API polling (Mikrotik devices)
- **Phase 2**: SNMP support (future enhancement)

## Data Model

### Domain: EnvironmentMetric

Represents a single environment metric reading.

```
EnvironmentMetric {
    ID              uint
    TenantID        string
    NasID           uint
    NasName         string    // For dashboard display
    MetricType      string    // "temperature", "power", "voltage", "signal_strength", "fan_speed"
    Value           float64
    Unit            string    // "C", "W", "V", "dBm", "RPM"
    Severity        string    // "normal", "warning", "critical"
    CollectedAt     time.Time
    CreatedAt       time.Time
}
```

### Domain: EnvironmentAlert

Represents an alert triggered by threshold violation.

```
EnvironmentAlert {
    ID              uint
    TenantID        string
    NasID           uint
    MetricType      string
    ThresholdType   string    // "min", "max"
    ThresholdValue  float64
    AlertValue      float64
    Severity        string    // "warning", "critical"
    Status          string    // "firing", "acknowledged", "resolved"
    NotifyStatus    string    // "pending", "sent", "failed"
    FiredAt         time.Time
    ResolvedAt      *time.Time
    AcknowledgedBy  *string
    AcknowledgedAt  *time.Time
    CreatedAt       time.Time
}
```

### Domain: AlertConfig

Configuration for alert thresholds per device.

```
AlertConfig {
    ID              uint
    TenantID        string
    NasID           uint
    MetricType      string
    ThresholdType   string    // "min" or "max"
    ThresholdValue  float64
    Severity        string    // "warning", "critical"
    Enabled         bool
    NotifyEmail     bool
    NotifyWebhook   bool
    WebhookURL      string    // HTTPS only, validated
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Environment Monitor                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐     │
│  │   Scheduler  │───▶│   Collector  │───▶│   Database   │     │
│  │  (cron job)  │    │  (RouterOS)  │    │  (metrics)   │     │
│  └──────────────┘    └──────────────┘    └──────────────┘     │
│         │                    │                    │            │
│         │              ┌─────▼─────┐              │            │
│         │              │  Alert    │              │            │
│         │              │  Engine   │              │            │
│         │              └─────┬─────┘              │            │
│         │                    │                    │            │
│         ▼                    ▼                    ▼            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Notification Queue                          │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                   │  │
│  │  │Dashboard│  │  Email  │  │ Webhook │                   │  │
│  │  └─────────┘  └─────────┘  └─────────┘                   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Collector Service (`internal/device/env_monitor.go`)

- Polls device health via RouterOS API
- Endpoints: `/system/health`, `/system/resource`, `/interface` (for wireless signal)
- Supports parallel batch collection
- Calculates severity based on thresholds

### 2. Alert Engine (`internal/device/alert_engine.go`)

- Evaluates thresholds after each collection
- Creates alerts on threshold violation
- Auto-resolves when metric returns to normal
- Queues notifications asynchronously

### 3. Notification Service (`internal/device/notifier.go`)

- **Dashboard**: Updates alert status in real-time via WebSocket
- **Email**: Sends email using existing email infrastructure
- **Webhook**: POSTs JSON payload to configured URLs

### 4. Scheduler

- Runs via built-in cron (per-device interval, default 5 minutes)
- Can be configured per device or globally
- Skips disabled alerts

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/network/nas/:id/metrics` | Get latest metrics for device |
| GET | `/api/v1/network/nas/:id/metrics/history` | Get metric history |
| GET | `/api/v1/network/nas/:id/alerts` | Get alerts for device |
| GET | `/api/v1/network/nas/:id/alerts/config` | Get alert config |
| PUT | `/api/v1/network/nas/:id/alerts/config` | Update alert config |
| POST | `/api/v1/network/nas/:id/alerts/:alertId/ack` | Acknowledge alert |
| GET | `/api/v1/dashboard/stats` | Get all devices health overview (reuse existing pattern with query param `?type=env`)

## RouterOS API Integration

### Temperature

```go
// Request: /system/health/print
// Response: cpu-temperature, board-temperature (hardware-dependent - CCR series, cloud routers)
// Fallback: /system/resource/print if /system/health unavailable (returns cpu-load, not temp)
```

### Power/Voltage

```go
// Request: /system/health/print
// Response: voltage, power-consumption
```

### Signal Strength (Wireless)

```go
// Request: /interface/wireless/registration-table/print
// Note: signal is per-station, not general interface. Query specific interface or iterate stations.
```

### Fan Status

```go
// Request: /system/health/print
// Response: fan1-speed, fan2-speed, fan-mode
```

## Migration

Add new tables:

```sql
CREATE TABLE environment_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    nas_id BIGINT NOT NULL,
    metric_type VARCHAR(32) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(16) NOT NULL,
    severity VARCHAR(16) NOT NULL DEFAULT 'normal',
    collected_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

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
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_env_metrics_nas_time ON environment_metrics(nas_id, collected_at);
CREATE INDEX idx_env_metrics_tenant_time ON environment_metrics(tenant_id, metric_type, collected_at);
CREATE INDEX idx_env_alerts_status ON environment_alerts(status, fired_at);
```

## Acceptance Criteria

1. ✅ Can collect temperature, power, voltage, signal strength, fan speed from RouterOS devices
2. ✅ Threshold-based alerts trigger when values exceed min/max limits
3. ✅ Alerts appear in dashboard with severity indicators
4. ✅ Email notifications sent for critical alerts
5. ✅ Webhook notifications POST JSON to configured URLs
6. ✅ Alerts can be acknowledged by admins
7. ✅ Auto-resolve when metric returns to normal
8. ✅ Metrics history queryable for last 7 days
9. ✅ Health overview shows all devices status at glance
10. ✅ Alert configuration CRUD via API

## Security Considerations

- **Tenant Isolation**: All API endpoints enforce `tenant_id` from JWT claims
- **Webhook URL Validation**: Only HTTPS allowed, max 2048 chars, URL format validated
- **API Rate Limiting**: Polling endpoint rate-limited to prevent abuse
- **Credential Encryption**: Device credentials stored encrypted at rest

## Edge Cases

- **Unsupported Hardware**: When `/system/health` unavailable, mark metric as "not_supported" not error
- **Device Offline**: Skip metrics collection, mark device as "offline" in health status
- **Alert Storm Prevention**: Cooldown period of 5 minutes between same alert re-triggers
- **Data Retention**: Auto-purge metrics older than 7 days (configurable)

