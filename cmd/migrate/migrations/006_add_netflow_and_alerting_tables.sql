-- Migration 006: Add NetFlow and Alerting Tables
-- This migration adds tables for NetFlow v9 traffic analysis and real-time alerting

-- NetFlow Records
CREATE TABLE IF NOT EXISTS netflow_record (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    router_id VARCHAR(100) NOT NULL,
    source_addr VARCHAR(45) NOT NULL,
    dest_addr VARCHAR(45) NOT NULL,
    source_port SMALLINT,
    dest_port SMALLINT,
    protocol SMALLINT,
    tos SMALLINT,
    tcp_flags SMALLINT,
    bytes BIGINT NOT NULL,
    packets BIGINT NOT NULL,
    flow_duration_ms INT NOT NULL,
    first_switched TIMESTAMP NOT NULL,
    last_switched TIMESTAMP NOT NULL,
    ingress_interface INT,
    egress_interface INT,
    direction SMALLINT,
    ipv6_flow_label INT,
    bgp_next_hop VARCHAR(45),
    bgp_prev_hop VARCHAR(45),
    mpls_label_top INT,
    application_id SMALLINT,
    application_name VARCHAR(100),
    vrf_id INT,
    user_id BIGINT,
    session_id VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_netflow_tenant ON netflow_record(tenant_id);
CREATE INDEX idx_netflow_router ON netflow_record(router_id);
CREATE INDEX idx_netflow_source ON netflow_record(source_addr);
CREATE INDEX idx_netflow_dest ON netflow_record(dest_addr);
CREATE INDEX idx_netflow_user ON netflow_record(user_id);
CREATE INDEX idx_netflow_session ON netflow_record(session_id);
CREATE INDEX idx_netflow_protocol ON netflow_record(protocol);
CREATE INDEX idx_netflow_created ON netflow_record(created_at DESC);

-- Traffic Summary (aggregated statistics)
CREATE TABLE IF NOT EXISTS traffic_summary (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    router_id VARCHAR(100),
    user_id BIGINT,
    session_id VARCHAR(64),
    source_subnet VARCHAR(64),
    dest_subnet VARCHAR(64),
    application_name VARCHAR(100),
    protocol SMALLINT,
    total_bytes BIGINT NOT NULL,
    total_packets BIGINT NOT NULL,
    total_flows BIGINT NOT NULL,
    duration_sec INT NOT NULL,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_traffic_summary_tenant ON traffic_summary(tenant_id);
CREATE INDEX idx_traffic_summary_user ON traffic_summary(user_id);
CREATE INDEX idx_traffic_summary_app ON traffic_summary(application_name);
CREATE INDEX idx_traffic_summary_created ON traffic_summary(created_at DESC);

-- Alert Rules
CREATE TABLE IF NOT EXISTS alert_rule (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    operator VARCHAR(10) NOT NULL,
    threshold FLOAT NOT NULL,
    duration INTEGER NOT NULL,
    severity VARCHAR(20) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    notification_channels JSONB,
    message_template TEXT,
    last_triggered TIMESTAMP,
    trigger_count INTEGER DEFAULT 0,
    cooldown_sec INTEGER DEFAULT 300,
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_rule_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id)
);

CREATE INDEX idx_alert_rule_tenant ON alert_rule(tenant_id);
CREATE INDEX idx_alert_rule_enabled ON alert_rule(enabled);
CREATE INDEX idx_alert_rule_severity ON alert_rule(severity);

-- Alerts (triggered alert instances)
CREATE TABLE IF NOT EXISTS alert (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    rule_id BIGINT NOT NULL,
    rule_name VARCHAR(200) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    triggered_at TIMESTAMP NOT NULL,
    acknowledged_at TIMESTAMP,
    acknowledged_by VARCHAR(100),
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(100),
    recipient_email VARCHAR(200),
    notification_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_alert_rule FOREIGN KEY (rule_id) REFERENCES alert_rule(id)
);

CREATE INDEX idx_alert_tenant ON alert(tenant_id);
CREATE INDEX idx_alert_rule ON alert(rule_id);
CREATE INDEX idx_alert_status ON alert(status);
CREATE INDEX idx_alert_severity ON alert(severity);
CREATE INDEX idx_alert_created ON alert(created_at DESC);