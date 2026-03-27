-- Migration 005: Add Certificate and IPoE Tables
-- This migration adds tables for 802.1x EAP-TLS certificate management
-- and IPoE (IP over Ethernet) authentication using DHCP option 82

-- Certificate Authorities
CREATE TABLE IF NOT EXISTS certificate_authority (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    common_name VARCHAR(255) NOT NULL,
    certificate_pem TEXT NOT NULL,
    private_key_pem TEXT NOT NULL,
    serial_number VARCHAR(255) UNIQUE,
    expiry_date TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    crl_url VARCHAR(500),
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ca_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id)
);

CREATE INDEX idx_ca_tenant ON certificate_authority(tenant_id);
CREATE INDEX idx_ca_status ON certificate_authority(status);
CREATE INDEX idx_ca_serial ON certificate_authority(serial_number);

-- Client Certificates
CREATE TABLE IF NOT EXISTS client_certificate (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    ca_id BIGINT,
    common_name VARCHAR(255) NOT NULL,
    serial_number VARCHAR(255) UNIQUE,
    certificate_pem TEXT NOT NULL,
    private_key_pem TEXT,
    expiry_date TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    revocation_reason VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    device_type VARCHAR(50),
    mac_address VARCHAR(17),
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_client_cert_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_client_cert_user FOREIGN KEY (user_id) REFERENCES radius_user(id),
    CONSTRAINT fk_client_cert_ca FOREIGN KEY (ca_id) REFERENCES certificate_authority(id)
);

CREATE INDEX idx_client_cert_tenant ON client_certificate(tenant_id);
CREATE INDEX idx_client_cert_user ON client_certificate(user_id);
CREATE INDEX idx_client_cert_cn ON client_certificate(common_name);
CREATE INDEX idx_client_cert_serial ON client_certificate(serial_number);
CREATE INDEX idx_client_cert_status ON client_certificate(status);

-- DHCP Option 82 Mappings
CREATE TABLE IF NOT EXISTS dhcp_option82 (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    circuit_id VARCHAR(255) NOT NULL,
    remote_id VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    mac_address VARCHAR(17),
    vendor_specific TEXT,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_dhcp82_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_dhcp82_user FOREIGN KEY (user_id) REFERENCES radius_user(id)
);

CREATE INDEX idx_dhcp82_tenant ON dhcp_option82(tenant_id);
CREATE INDEX idx_dhcp82_user ON dhcp_option82(user_id);
CREATE INDEX idx_dhcp82_circuit ON dhcp_option82(circuit_id);
CREATE INDEX idx_dhcp82_remote ON dhcp_option82(remote_id);
CREATE INDEX idx_dhcp82_ip ON dhcp_option82(ip_address);
CREATE UNIQUE INDEX idx_dhcp82_unique ON dhcp_option82(circuit_id, remote_id, ip_address);

-- IPoE Sessions
CREATE TABLE IF NOT EXISTS ipoe_session (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    mac_address VARCHAR(17),
    circuit_id VARCHAR(255),
    remote_id VARCHAR(255),
    session_id VARCHAR(64) NOT NULL UNIQUE,
    nas_id BIGINT,
    nas_port VARCHAR(50),
    framed_ip VARCHAR(45),
    session_start TIMESTAMP NOT NULL,
    session_update TIMESTAMP NOT NULL,
    input_octets BIGINT DEFAULT 0,
    output_octets BIGINT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    terminate_cause VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ipoe_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_ipoe_user FOREIGN KEY (user_id) REFERENCES radius_user(id),
    CONSTRAINT fk_ipoe_nas FOREIGN KEY (nas_id) REFERENCES net_nas(id)
);

CREATE INDEX idx_ipoe_tenant ON ipoe_session(tenant_id);
CREATE INDEX idx_ipoe_user ON ipoe_session(user_id);
CREATE INDEX idx_ipoe_session_id ON ipoe_session(session_id);
CREATE INDEX idx_ipoe_ip ON ipoe_session(ip_address);
CREATE INDEX idx_ipoe_status ON ipoe_session(status);

-- Add comments for documentation
COMMENT ON TABLE certificate_authority IS 'X.509 Certificate Authorities for issuing client certificates in 802.1x EAP-TLS authentication';
COMMENT ON TABLE client_certificate IS 'X.509 client certificates issued to users/devices for 802.1x authentication';
COMMENT ON TABLE dhcp_option82 IS 'DHCP option 82 (Relay Agent Information) mappings for IPoE authentication';
COMMENT ON TABLE ipoe_session IS 'IPoE (IP over Ethernet) authentication sessions for cable/DSL broadband networks';
