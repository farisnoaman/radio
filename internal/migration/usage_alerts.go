package migration

import (
	"fmt"

	"gorm.io/gorm"
)

// CreateUsageAlertsTables creates the usage_alerts and notification_preferences tables
func CreateUsageAlertsTables(db *gorm.DB) error {
	// Create usage_alerts table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS usage_alerts (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES radius_user(id) ON DELETE CASCADE,
			threshold INTEGER NOT NULL,
			alert_type VARCHAR(10) NOT NULL,
			sent_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_usage_alerts_user_threshold ON usage_alerts(user_id, threshold);
		CREATE INDEX IF NOT EXISTS idx_usage_alerts_sent_at ON usage_alerts(sent_at);
	`).Error; err != nil {
		return fmt.Errorf("failed to create usage_alerts table: %w", err)
	}

	// Create notification_preferences table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_preferences (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL UNIQUE REFERENCES radius_user(id) ON DELETE CASCADE,
			email_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			sms_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			email_thresholds VARCHAR(50) NOT NULL DEFAULT '80,90,100',
			sms_thresholds VARCHAR(50) NOT NULL DEFAULT '100',
			daily_summary_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`).Error; err != nil {
		return fmt.Errorf("failed to create notification_preferences table: %w", err)
	}

	return nil
}

// DropUsageAlertsTables drops the usage_alerts and notification_preferences tables
func DropUsageAlertsTables(db *gorm.DB) error {
	if err := db.Exec("DROP TABLE IF EXISTS usage_alerts CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop usage_alerts table: %w", err)
	}

	if err := db.Exec("DROP TABLE IF EXISTS notification_preferences CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop notification_preferences table: %w", err)
	}

	return nil
}
