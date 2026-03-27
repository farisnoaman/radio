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
