-- Migration 003: Add Device Management Tables
-- This migration adds tables for NAS templates, device backups, and speed tests

-- NAS Templates
CREATE TABLE IF NOT EXISTS nas_template (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    vendor_code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    attributes JSONB NOT NULL,
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_nas_template_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id)
);

CREATE INDEX idx_nas_template_tenant ON nas_template(tenant_id);
CREATE INDEX idx_nas_template_vendor ON nas_template(vendor_code);

-- Device Configuration Backups
CREATE TABLE IF NOT EXISTS device_config_backup (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    nas_id BIGINT NOT NULL,
    vendor_code VARCHAR(50) NOT NULL,
    config_data TEXT NOT NULL,
    file_size BIGINT,
    status VARCHAR(20) DEFAULT 'pending',
    error TEXT,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_device_backup_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_device_backup_nas FOREIGN KEY (nas_id) REFERENCES net_nas(id)
);

CREATE INDEX idx_device_backup_tenant ON device_config_backup(tenant_id);
CREATE INDEX idx_device_backup_nas ON device_config_backup(nas_id);
CREATE INDEX idx_device_backup_status ON device_config_backup(status);

-- Speed Test Results
CREATE TABLE IF NOT EXISTS speed_test_result (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    nas_id BIGINT NOT NULL,
    test_server VARCHAR(200),
    upload_mbps DECIMAL(10,2),
    download_mbps DECIMAL(10,2),
    latency_ms DECIMAL(10,2),
    jitter_ms DECIMAL(10,2),
    packet_loss_percent DECIMAL(5,2),
    test_duration_sec INTEGER,
    status VARCHAR(20) DEFAULT 'running',
    error TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_speedtest_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_speedtest_nas FOREIGN KEY (nas_id) REFERENCES net_nas(id)
);

CREATE INDEX idx_speedtest_tenant ON speed_test_result(tenant_id);
CREATE INDEX idx_speedtest_nas ON speed_test_result(nas_id);
CREATE INDEX idx_speedtest_created ON speed_test_result(created_at DESC);

-- Add comments for documentation
COMMENT ON TABLE nas_template IS 'Vendor-specific RADIUS attribute templates for different NAS equipment';
COMMENT ON TABLE device_config_backup IS 'Automatic configuration backups for network devices';
COMMENT ON TABLE speed_test_result IS 'Network speed test results from bandwidth testing tools';
