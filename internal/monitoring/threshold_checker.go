package monitoring

import (
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ThresholdConfig struct {
	CPUWarning     float64
	CPUCritical    float64
	MemoryWarning  float64
	MemoryCritical float64
	TempWarning    float64
	TempCritical   float64
	VoltageMin     float64
	VoltageMax     float64
	SignalWarning  float64
	SignalCritical float64
}

var DefaultThresholdConfig = ThresholdConfig{
	CPUWarning:     70,
	CPUCritical:    90,
	MemoryWarning:  80,
	MemoryCritical: 90,
	TempWarning:    50,
	TempCritical:   70,
	VoltageMin:     22,
	VoltageMax:     25,
	SignalWarning:  -70,
	SignalCritical: -80,
}

type ThresholdChecker struct {
	db     *gorm.DB
	config ThresholdConfig
}

func NewThresholdChecker(db *gorm.DB, config *ThresholdConfig) *ThresholdChecker {
	if config == nil {
		config = &DefaultThresholdConfig
	}
	return &ThresholdChecker{
		db:     db,
		config: *config,
	}
}

func (tc *ThresholdChecker) CheckMetric(metric *domain.NetworkDeviceMetric, device *domain.NetworkDevice) *domain.NetworkDeviceAlert {
	var threshold, warning, critical float64
	var metricName string

	switch metric.MetricType {
	case domain.MetricTypeCPU:
		metricName = "CPU"
		warning = tc.config.CPUWarning
		critical = tc.config.CPUCritical
	case domain.MetricTypeMemory:
		metricName = "Memory"
		warning = tc.config.MemoryWarning
		critical = tc.config.MemoryCritical
	case domain.MetricTypeDeviceTemp:
		metricName = "Temperature"
		warning = tc.config.TempWarning
		critical = tc.config.TempCritical
	case domain.MetricTypeDeviceVolt:
		metricName = "Voltage"
		if metric.Value < tc.config.VoltageMin {
			threshold = tc.config.VoltageMin
		} else if metric.Value > tc.config.VoltageMax {
			threshold = tc.config.VoltageMax
			warning = tc.config.VoltageMax
			critical = 26
		}
	case domain.MetricTypeSignal:
		metricName = "Signal"
		if metric.Value < tc.config.SignalCritical {
			critical = tc.config.SignalCritical
		} else if metric.Value < tc.config.SignalWarning {
			warning = tc.config.SignalWarning
		}
	default:
		return nil
	}

	var severity string
	var shouldAlert bool

	if critical > 0 && (threshold > 0 && metric.Value <= threshold || metric.Value >= critical) {
		severity = domain.DevSeverityCritical
		shouldAlert = true
	} else if warning > 0 && metric.Value >= warning {
		severity = domain.DevSeverityWarning
		shouldAlert = true
	}

	if !shouldAlert {
		return nil
	}

	return &domain.NetworkDeviceAlert{
		DeviceID:       device.ID,
		TenantID:       device.TenantID,
		AlertType:      domain.AlertTypeThreshold,
		Severity:       severity,
		Message:        fmt.Sprintf("%s threshold exceeded: %.2f", metricName, metric.Value),
		MetricType:     metric.MetricType,
		MetricValue:    &metric.Value,
		ThresholdValue: &threshold,
		Status:         domain.DevAlertStatusActive,
	}
}

func (tc *ThresholdChecker) CheckMetricAndCreate(metric *domain.NetworkDeviceMetric, device *domain.NetworkDevice) {
	alert := tc.CheckMetric(metric, device)
	if alert == nil {
		return
	}

	var existing domain.NetworkDeviceAlert
	err := tc.db.Where(
		"device_id = ? AND status = ? AND metric_type = ? AND severity = ?",
		device.ID, domain.DevAlertStatusActive, metric.MetricType, alert.Severity,
	).First(&existing).Error

	if err == nil && existing.ID > 0 {
		return
	}

	if err := tc.db.Create(alert).Error; err != nil {
		zap.S().Errorw("Failed to create threshold alert",
			"device_id", device.ID,
			"metric_type", metric.MetricType,
			"error", err)
	} else {
		zap.S().Infow("Created threshold alert",
			"device_id", device.ID,
			"device_name", device.Name,
			"severity", alert.Severity,
			"metric_type", metric.MetricType,
			"value", metric.Value)
	}
}

func (tc *ThresholdChecker) ResolveStaleAlerts() (int64, error) {
	cutoff := time.Now().Add(-24 * time.Hour)

	var resolved int64
	err := tc.db.Model(&domain.NetworkDeviceAlert{}).
		Where("status = ? AND created_at < ?", domain.DevAlertStatusActive, cutoff).
		Updates(map[string]interface{}{
			"status":      domain.DevAlertStatusResolved,
			"resolved_at": time.Now(),
		}).Error

	if err != nil {
		return 0, err
	}

	return resolved, nil
}

func (tc *ThresholdChecker) GetAlertCounts(tenantID int64) (map[string]int64, error) {
	counts := make(map[string]int64)

	var total int64
	tc.db.Model(&domain.NetworkDeviceAlert{}).Where("tenant_id = ?", tenantID).Count(&total)
	counts["total"] = total

	var active int64
	tc.db.Model(&domain.NetworkDeviceAlert{}).
		Where("tenant_id = ? AND status = ?", tenantID, domain.DevAlertStatusActive).
		Count(&active)
	counts["active"] = active

	var critical int64
	tc.db.Model(&domain.NetworkDeviceAlert{}).
		Where("tenant_id = ? AND severity = ? AND status = ?", tenantID, domain.DevSeverityCritical, domain.DevAlertStatusActive).
		Count(&critical)
	counts["critical"] = critical

	var warning int64
	tc.db.Model(&domain.NetworkDeviceAlert{}).
		Where("tenant_id = ? AND severity = ? AND status = ?", tenantID, domain.DevSeverityWarning, domain.DevAlertStatusActive).
		Count(&warning)
	counts["warning"] = warning

	return counts, nil
}
