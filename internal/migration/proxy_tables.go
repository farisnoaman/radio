package migration

import (
	"fmt"
	"gorm.io/gorm"
)

// CreateProxyTables creates all tables for RADIUS proxy functionality.
func CreateProxyTables(db *gorm.DB) error {
	// RADIUS Proxy Servers table
	if err := db.Exec(`
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
		)`).Error; err != nil {
		return fmt.Errorf("failed to create radius_proxy_server table: %w", err)
	}

	// RADIUS Proxy Realms table
	if err := db.Exec(`
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
		)`).Error; err != nil {
		return fmt.Errorf("failed to create radius_proxy_realm table: %w", err)
	}

	// Proxy Request Logs table
	if err := db.Exec(`
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
		)`).Error; err != nil {
		return fmt.Errorf("failed to create proxy_request_log table: %w", err)
	}

	// Create indexes for radius_proxy_server
	serverIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_proxy_server_tenant ON radius_proxy_server(tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_server_status ON radius_proxy_server(status)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_server_priority ON radius_proxy_server(priority)",
	}

	for _, idx := range serverIndexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create indexes for radius_proxy_realm
	realmIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_proxy_realm_tenant ON radius_proxy_realm(tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_realm_realm ON radius_proxy_realm(realm)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_realm_status ON radius_proxy_realm(status)",
	}

	for _, idx := range realmIndexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create indexes for proxy_request_log
	logIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_proxy_log_tenant ON proxy_request_log(tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_log_realm ON proxy_request_log(realm)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_log_server ON proxy_request_log(server_id)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_log_created ON proxy_request_log(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_log_success ON proxy_request_log(success)",
	}

	for _, idx := range logIndexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// DropProxyTables drops all RADIUS proxy tables.
func DropProxyTables(db *gorm.DB) error {
	// Drop tables in reverse order due to foreign keys
	tables := []string{
		"proxy_request_log",
		"radius_proxy_realm",
		"radius_proxy_server",
	}

	for _, table := range tables {
		if err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
