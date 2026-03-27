package device

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AlertEngine struct {
	db           *gorm.DB
	notifier     *Notifier
	cooldownMins int
}

func NewAlertEngine(db *gorm.DB, notifier *Notifier) *AlertEngine {
	return &AlertEngine{
		db:           db,
		notifier:     notifier,
		cooldownMins: 5,
	}
}

func (e *AlertEngine) ProcessMetrics(ctx context.Context, metrics []domain.EnvironmentMetric) error {
	for _, metric := range metrics {
		if err := e.evaluateThreshold(ctx, metric); err != nil {
			zap.S().Errorw("Failed to evaluate threshold", "metric", metric.MetricType, "error", err)
		}
	}
	return nil
}

func (e *AlertEngine) evaluateThreshold(ctx context.Context, metric domain.EnvironmentMetric) error {
	var configs []domain.AlertConfig
	if err := e.db.Where("nas_id = ? AND metric_type = ? AND enabled = ?", 
		metric.NasID, metric.MetricType, true).Find(&configs).Error; err != nil {
		return err
	}

	for _, cfg := range configs {
		triggered := false
		alertValue := metric.Value

		if cfg.ThresholdType == domain.ThresholdTypeMax && metric.Value > cfg.ThresholdValue {
			triggered = true
		} else if cfg.ThresholdType == domain.ThresholdTypeMin && metric.Value < cfg.ThresholdValue {
			triggered = true
		}

		if triggered {
			if err := e.createOrUpdateAlert(ctx, metric, cfg, alertValue); err != nil {
				return err
			}
		} else {
			e.resolveAlert(ctx, metric.NasID, metric.MetricType)
		}
	}
	return nil
}

func (e *AlertEngine) createOrUpdateAlert(ctx context.Context, metric domain.EnvironmentMetric, cfg domain.AlertConfig, alertValue float64) error {
	var existing domain.EnvironmentAlert
	err := e.db.Where("nas_id = ? AND metric_type = ? AND status = ?", 
		metric.NasID, metric.MetricType, domain.AlertStatusFiring).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		alert := domain.EnvironmentAlert{
			TenantID:       metric.TenantID,
			NasID:          metric.NasID,
			MetricType:     metric.MetricType,
			ThresholdType:  cfg.ThresholdType,
			ThresholdValue: cfg.ThresholdValue,
			AlertValue:     alertValue,
			Severity:       cfg.Severity,
			Status:         domain.AlertStatusFiring,
			NotifyStatus:   domain.NotifyStatusPending,
			FiredAt:        time.Now(),
			CreatedAt:      time.Now(),
		}
		if err := e.db.Create(&alert).Error; err != nil {
			return err
		}
		if e.notifier != nil {
			go e.notifier.SendAlertNotifications(ctx, alert, cfg)
		}
	} else if err == nil {
		existing.AlertValue = alertValue
		e.db.Save(&existing)
	}
	return nil
}

func (e *AlertEngine) resolveAlert(ctx context.Context, nasID uint, metricType string) error {
	now := time.Now()
	return e.db.Model(&domain.EnvironmentAlert{}).
		Where("nas_id = ? AND metric_type = ? AND status = ?", nasID, metricType, domain.AlertStatusFiring).
		Updates(map[string]interface{}{
			"status":      domain.AlertStatusResolved,
			"resolved_at": now,
		}).Error
}

func (e *AlertEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := e.ProcessAllMetrics(ctx); err != nil {
				zap.S().Errorw("Alert engine error", "error", err)
			}
		}
	}
}

func (e *AlertEngine) ProcessAllMetrics(ctx context.Context) error {
	var latestMetrics []domain.EnvironmentMetric

	subQuery := e.db.Model(&domain.EnvironmentMetric{}).
		Select("nas_id, metric_type, MAX(collected_at) as max_collected").
		Group("nas_id, metric_type")

	if err := e.db.Raw(`
		SELECT em.* FROM environment_metrics em
		INNER JOIN (?) tm ON em.nas_id = tm.nas_id AND em.metric_type = tm.metric_type AND em.collected_at = tm.max_collected
	`, subQuery).Find(&latestMetrics).Error; err != nil {
		return err
	}

	return e.ProcessMetrics(ctx, latestMetrics)
}
