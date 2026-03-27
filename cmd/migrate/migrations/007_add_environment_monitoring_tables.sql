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
