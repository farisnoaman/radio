package migration

import (
	"fmt"

	"gorm.io/gorm"
)

// CreateReportingTables creates all reporting tables for the provider dashboard
func CreateReportingTables(db *gorm.DB) error {
	// Create daily snapshots table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS reporting_daily_snapshots (
			id BIGSERIAL PRIMARY KEY,
			provider_id BIGINT NOT NULL,
			snapshot_date DATE NOT NULL,
			total_users INT DEFAULT 0,
			active_users INT DEFAULT 0,
			new_monthly_users INT DEFAULT 0,
			new_voucher_users INT DEFAULT 0,
			total_sessions INT DEFAULT 0,
			active_sessions INT DEFAULT 0,
			monthly_data_used_bytes BIGINT DEFAULT 0,
			voucher_data_used_bytes BIGINT DEFAULT 0,
			active_nodes INT DEFAULT 0,
			total_nodes INT DEFAULT 0,
			active_servers INT DEFAULT 0,
			total_servers INT DEFAULT 0,
			active_cpes INT DEFAULT 0,
			total_cpes INT DEFAULT 0,
			total_agents INT DEFAULT 0,
			total_batches INT DEFAULT 0,
			agent_revenue DECIMAL(15,2) DEFAULT 0,
			mrr DECIMAL(15,2) DEFAULT 0,
			device_issues INT DEFAULT 0,
			network_issues INT DEFAULT 0,
			fraud_attempts INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(provider_id, snapshot_date)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create reporting_daily_snapshots table: %w", err)
	}

	// Create indexes for snapshots
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_daily_snapshots_provider_date ON reporting_daily_snapshots(provider_id, snapshot_date)
	`).Error; err != nil {
		return fmt.Errorf("failed to create snapshots index: %w", err)
	}

	// Create fraud log table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS reporting_fraud_log (
			id BIGSERIAL PRIMARY KEY,
			provider_id BIGINT NOT NULL,
			voucher_id BIGINT,
			user_id BIGINT,
			ip_address VARCHAR(45),
			event_type VARCHAR(50) NOT NULL,
			details JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create reporting_fraud_log table: %w", err)
	}

	// Create indexes for fraud log
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_fraud_log_provider ON reporting_fraud_log(provider_id);
		CREATE INDEX IF NOT EXISTS idx_fraud_log_ip_time ON reporting_fraud_log(ip_address, created_at)
	`).Error; err != nil {
		return fmt.Errorf("failed to create fraud log indexes: %w", err)
	}

	// Create provider notification preferences table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS provider_notification_preferences (
			id BIGSERIAL PRIMARY KEY,
			provider_id BIGINT NOT NULL UNIQUE,
			alert_percentages VARCHAR(50) DEFAULT '70,85,100',
			alert_percentages_enabled BOOLEAN DEFAULT TRUE,
			max_users_threshold INT,
			max_data_bytes_threshold BIGINT,
			absolute_alerts_enabled BOOLEAN DEFAULT FALSE,
			anomaly_detection_enabled BOOLEAN DEFAULT FALSE,
			anomaly_threshold_percent INT DEFAULT 50,
			email_enabled BOOLEAN DEFAULT TRUE,
			sms_enabled BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create provider_notification_preferences table: %w", err)
	}

	// Create network issues table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS reporting_network_issues (
			id BIGSERIAL PRIMARY KEY,
			provider_id BIGINT NOT NULL,
			device_type VARCHAR(20) NOT NULL,
			device_id BIGINT,
			device_name VARCHAR(255),
			issue_type VARCHAR(50) NOT NULL,
			issue_details TEXT,
			status VARCHAR(20) DEFAULT 'open',
			resolved_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create reporting_network_issues table: %w", err)
	}

	// Create indexes for network issues
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_network_issues_provider ON reporting_network_issues(provider_id);
		CREATE INDEX IF NOT EXISTS idx_network_issues_status ON reporting_network_issues(status, created_at)
	`).Error; err != nil {
		return fmt.Errorf("failed to create network issues indexes: %w", err)
	}

	return nil
}

// DropReportingTables drops all reporting tables
func DropReportingTables(db *gorm.DB) error {
	if err := db.Exec("DROP TABLE IF EXISTS reporting_daily_snapshots CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop reporting_daily_snapshots table: %w", err)
	}

	if err := db.Exec("DROP TABLE IF EXISTS reporting_fraud_log CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop reporting_fraud_log table: %w", err)
	}

	if err := db.Exec("DROP TABLE IF EXISTS provider_notification_preferences CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop provider_notification_preferences table: %w", err)
	}

	if err := db.Exec("DROP TABLE IF EXISTS reporting_network_issues CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop reporting_network_issues table: %w", err)
	}

	return nil
}
