-- Migration 004: Add RADIUS Proxy Tables
-- This migration adds tables for RADIUS proxy functionality

-- RADIUS Proxy Servers
CREATE TABLE IF NOT EXISTS radius_proxy_server (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    host VARCHAR(255) NOT NULL,
    auth_port INTEGER DEFAULT 1812 NOT NULL,
    acct_port INTEGER DEFAULT 1813 NOT NULL,
    secret VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'enabled',
    max_conns INTEGER DEFAULT 50,
    timeout_sec INTEGER DEFAULT 5,
    priority INTEGER DEFAULT 1,
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_proxy_server_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id)
);

CREATE INDEX idx_proxy_server_tenant ON radius_proxy_server(tenant_id);
CREATE INDEX idx_proxy_server_status ON radius_proxy_server(status);
CREATE INDEX idx_proxy_server_priority ON radius_proxy_server(priority);

-- RADIUS Proxy Realms
CREATE TABLE IF NOT EXISTS radius_proxy_realm (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    realm VARCHAR(255) NOT NULL,
    proxy_servers BIGINT[] NOT NULL,
    fallback_order INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'enabled',
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_proxy_realm_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id)
);

CREATE INDEX idx_proxy_realm_tenant ON radius_proxy_realm(tenant_id);
CREATE INDEX idx_proxy_realm_realm ON radius_proxy_realm(realm);
CREATE INDEX idx_proxy_realm_status ON radius_proxy_realm(status);

-- Proxy Request Logs
CREATE TABLE IF NOT EXISTS proxy_request_log (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    realm VARCHAR(255),
    username VARCHAR(255),
    server_id BIGINT,
    request_type VARCHAR(10),
    success BOOLEAN DEFAULT false,
    latency_ms INTEGER,
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_proxy_log_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_proxy_log_server FOREIGN KEY (server_id) REFERENCES radius_proxy_server(id)
);

CREATE INDEX idx_proxy_log_tenant ON proxy_request_log(tenant_id);
CREATE INDEX idx_proxy_log_realm ON proxy_request_log(realm);
CREATE INDEX idx_proxy_log_server ON proxy_request_log(server_id);
CREATE INDEX idx_proxy_log_created ON proxy_request_log(created_at DESC);
CREATE INDEX idx_proxy_log_success ON proxy_request_log(success);

-- Add comments for documentation
COMMENT ON TABLE radius_proxy_server IS 'Upstream RADIUS servers for proxying authentication requests';
COMMENT ON TABLE radius_proxy_realm IS 'Routing realms for proxy requests based on username suffix';
COMMENT ON TABLE proxy_request_log IS 'Audit log for all proxied RADIUS requests';
