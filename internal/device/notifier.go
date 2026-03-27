package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Notifier struct {
	db        *gorm.DB
	emailSvc  *service.SMTPEmailProvider
}

func NewNotifier(db *gorm.DB, emailSvc *service.SMTPEmailProvider) *Notifier {
	return &Notifier{
		db:       db,
		emailSvc: emailSvc,
	}
}

func (n *Notifier) SendAlertNotifications(ctx context.Context, alert domain.EnvironmentAlert, cfg domain.AlertConfig) {
	if cfg.NotifyEmail {
		n.sendEmail(ctx, alert, cfg)
	}
	if cfg.NotifyWebhook && cfg.WebhookURL != "" {
		n.sendWebhook(ctx, alert, cfg)
	}

	status := domain.NotifyStatusSent
	if err := n.db.Model(&alert).Update("notify_status", status).Error; err != nil {
		zap.S().Errorw("Failed to update notify status", "error", err)
	}
}

func (n *Notifier) sendEmail(ctx context.Context, alert domain.EnvironmentAlert, cfg domain.AlertConfig) {
	var nasName string
	var nas struct{ Name string }
	if err := n.db.Model(&domain.NetNas{}).Where("id = ?", alert.NasID).First(&nas).Error; err == nil {
		nasName = nas.Name
	}

	subject := fmt.Sprintf("[%s] Device Alert: %s", alert.Severity, alert.MetricType)
	body := fmt.Sprintf(`
Device Environment Alert

Metric: %s
Device: %s (%d)
Current Value: %.2f
Threshold: %s %.2f
Severity: %s
Time: %s

Please take action.
`, alert.MetricType, nasName, alert.NasID, alert.AlertValue, 
		alert.ThresholdType, alert.ThresholdValue, alert.Severity, alert.FiredAt.Format(time.RFC3339))

	if n.emailSvc != nil {
		if err := n.emailSvc.SendEmail("admin@provider.local", subject, body); err != nil {
			zap.S().Errorw("Failed to send alert email", "error", err)
		}
	}
}

func (n *Notifier) sendWebhook(ctx context.Context, alert domain.EnvironmentAlert, cfg domain.AlertConfig) {
	payload := map[string]interface{}{
		"alert_id":       alert.ID,
		"metric_type":    alert.MetricType,
		"nas_id":         alert.NasID,
		"alert_value":    alert.AlertValue,
		"threshold":      alert.ThresholdValue,
		"threshold_type": alert.ThresholdType,
		"severity":       alert.Severity,
		"fired_at":       alert.FiredAt,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", cfg.WebhookURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Errorw("Webhook failed", "url", cfg.WebhookURL, "error", err)
		n.db.Model(&domain.EnvironmentAlert{}).Where("id = ?", alert.ID).Update("notify_status", domain.NotifyStatusFailed)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		zap.S().Errorw("Webhook returned error", "status", resp.StatusCode)
		n.db.Model(&domain.EnvironmentAlert{}).Where("id = ?", alert.ID).Update("notify_status", domain.NotifyStatusFailed)
	}
}
