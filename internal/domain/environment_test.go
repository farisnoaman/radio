package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentMetric_TableName(t *testing.T) {
	metric := EnvironmentMetric{}
	assert.Equal(t, "environment_metrics", metric.TableName())
}

func TestEnvironmentMetric_Fields(t *testing.T) {
	now := time.Now()
	metric := EnvironmentMetric{
		ID:          1,
		TenantID:    "tenant-123",
		NasID:       100,
		NasName:     "Mikrotik-CCR",
		MetricType:  MetricTypeTemperature,
		Value:       65.5,
		Unit:        "C",
		Severity:    SeverityNormal,
		CollectedAt: now,
		CreatedAt:   now,
	}

	assert.Equal(t, uint(1), metric.ID)
	assert.Equal(t, "tenant-123", metric.TenantID)
	assert.Equal(t, uint(100), metric.NasID)
	assert.Equal(t, "Mikrotik-CCR", metric.NasName)
	assert.Equal(t, MetricTypeTemperature, metric.MetricType)
	assert.Equal(t, float64(65.5), metric.Value)
	assert.Equal(t, "C", metric.Unit)
	assert.Equal(t, SeverityNormal, metric.Severity)
}

func TestEnvironmentAlert_TableName(t *testing.T) {
	alert := EnvironmentAlert{}
	assert.Equal(t, "environment_alerts", alert.TableName())
}

func TestEnvironmentAlert_Fields(t *testing.T) {
	now := time.Now()
	resolvedAt := now.Add(1 * time.Hour)
	ackBy := "admin"
	ackAt := now.Add(30 * time.Minute)

	alert := EnvironmentAlert{
		ID:             1,
		TenantID:       "tenant-123",
		NasID:          100,
		MetricType:     MetricTypeTemperature,
		ThresholdType:  ThresholdTypeMax,
		ThresholdValue: 70,
		AlertValue:     85,
		Severity:       SeverityCritical,
		Status:         AlertStatusAcknowledged,
		NotifyStatus:   NotifyStatusSent,
		FiredAt:        now,
		ResolvedAt:     &resolvedAt,
		AcknowledgedBy: &ackBy,
		AcknowledgedAt: &ackAt,
		CreatedAt:      now,
	}

	assert.Equal(t, uint(1), alert.ID)
	assert.Equal(t, "tenant-123", alert.TenantID)
	assert.Equal(t, uint(100), alert.NasID)
	assert.Equal(t, MetricTypeTemperature, alert.MetricType)
	assert.Equal(t, ThresholdTypeMax, alert.ThresholdType)
	assert.Equal(t, float64(70), alert.ThresholdValue)
	assert.Equal(t, float64(85), alert.AlertValue)
	assert.Equal(t, SeverityCritical, alert.Severity)
	assert.Equal(t, AlertStatusAcknowledged, alert.Status)
	assert.Equal(t, NotifyStatusSent, alert.NotifyStatus)
	assert.NotNil(t, alert.ResolvedAt)
	assert.NotNil(t, alert.AcknowledgedBy)
	assert.Equal(t, "admin", *alert.AcknowledgedBy)
}

func TestAlertConfig_TableName(t *testing.T) {
	cfg := AlertConfig{}
	assert.Equal(t, "alert_configs", cfg.TableName())
}

func TestAlertConfig_Fields(t *testing.T) {
	now := time.Now()
	cfg := AlertConfig{
		ID:             1,
		TenantID:       "tenant-123",
		NasID:          100,
		MetricType:     MetricTypeTemperature,
		ThresholdType:  ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       SeverityWarning,
		Enabled:        true,
		NotifyEmail:    true,
		NotifyWebhook:  true,
		WebhookURL:     "https://example.com/webhook",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	assert.Equal(t, uint(1), cfg.ID)
	assert.Equal(t, "tenant-123", cfg.TenantID)
	assert.Equal(t, uint(100), cfg.NasID)
	assert.Equal(t, MetricTypeTemperature, cfg.MetricType)
	assert.Equal(t, ThresholdTypeMax, cfg.ThresholdType)
	assert.Equal(t, float64(70), cfg.ThresholdValue)
	assert.Equal(t, SeverityWarning, cfg.Severity)
	assert.Equal(t, true, cfg.Enabled)
	assert.Equal(t, true, cfg.NotifyEmail)
	assert.Equal(t, true, cfg.NotifyWebhook)
	assert.Equal(t, "https://example.com/webhook", cfg.WebhookURL)
}

func TestMetricType_Constants(t *testing.T) {
	assert.Equal(t, "temperature", MetricTypeTemperature)
	assert.Equal(t, "power", MetricTypePower)
	assert.Equal(t, "voltage", MetricTypeVoltage)
	assert.Equal(t, "signal_strength", MetricTypeSignalStrength)
	assert.Equal(t, "fan_speed", MetricTypeFanSpeed)
}

func TestSeverity_Constants(t *testing.T) {
	assert.Equal(t, "normal", SeverityNormal)
	assert.Equal(t, "warning", SeverityWarning)
	assert.Equal(t, "critical", SeverityCritical)
}

func TestThresholdType_Constants(t *testing.T) {
	assert.Equal(t, "min", ThresholdTypeMin)
	assert.Equal(t, "max", ThresholdTypeMax)
}

func TestAlertStatus_Constants(t *testing.T) {
	assert.Equal(t, "firing", AlertStatusFiring)
	assert.Equal(t, "acknowledged", AlertStatusAcknowledged)
	assert.Equal(t, "resolved", AlertStatusResolved)
}

func TestNotifyStatus_Constants(t *testing.T) {
	assert.Equal(t, "pending", NotifyStatusPending)
	assert.Equal(t, "sent", NotifyStatusSent)
	assert.Equal(t, "failed", NotifyStatusFailed)
}

func TestEnvironmentMetric_JSONTags(t *testing.T) {
	_ = EnvironmentMetric{}

	jsonTags := map[string]string{
		"ID":          "id",
		"TenantID":    "tenant_id",
		"NasID":       "nas_id",
		"NasName":     "nas_name",
		"MetricType":  "metric_type",
		"Value":       "value",
		"Unit":        "unit",
		"Severity":    "severity",
		"CollectedAt": "collected_at",
		"CreatedAt":   "created_at",
	}

	for field := range jsonTags {
		assert.NotEmpty(t, field)
	}
}

func TestEnvironmentAlert_PointerFields(t *testing.T) {
	alert := EnvironmentAlert{}

	assert.Nil(t, alert.ResolvedAt)
	assert.Nil(t, alert.AcknowledgedBy)
	assert.Nil(t, alert.AcknowledgedAt)

	later := time.Now()
	name := "operator"
	alert.ResolvedAt = &later
	alert.AcknowledgedBy = &name
	alert.AcknowledgedAt = &later

	assert.NotNil(t, alert.ResolvedAt)
	assert.NotNil(t, alert.AcknowledgedBy)
	assert.NotNil(t, alert.AcknowledgedAt)
	assert.Equal(t, "operator", *alert.AcknowledgedBy)
}
