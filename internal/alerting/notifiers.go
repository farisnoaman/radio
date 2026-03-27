package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
}

type EmailNotifier struct {
	config EmailConfig
}

func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{config: config}
}

func (n *EmailNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	if alert.RecipientEmail == "" {
		return fmt.Errorf("recipient email is required")
	}

	subject := fmt.Sprintf("[%s] %s", alert.Severity, alert.RuleName)
	body := fmt.Sprintf("Alert: %s\n\nSeverity: %s\nMessage: %s\nTriggered: %s",
		alert.RuleName, alert.Severity, alert.Message, alert.TriggeredAt.Format(time.RFC3339))

	auth := smtp.PlainAuth("", n.config.SMTPUser, n.config.SMTPPassword, n.config.SMTPHost)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		n.config.FromAddress, alert.RecipientEmail, subject, body)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", n.config.SMTPHost, n.config.SMTPPort),
		auth,
		n.config.FromAddress,
		[]string{alert.RecipientEmail},
		[]byte(msg),
	)
	if err != nil {
		zap.S().Error("Failed to send email notification", zap.Error(err))
		return err
	}

	zap.S().Info("Email notification sent",
		zap.String("to", alert.RecipientEmail),
		zap.String("rule", alert.RuleName))
	return nil
}

func (n *EmailNotifier) Name() string {
	return "email"
}

type WebhookConfig struct {
	URL     string
	Headers map[string]string
	Timeout time.Duration
}

type WebhookNotifier struct {
	client *http.Client
	config WebhookConfig
}

func NewWebhookNotifier(config WebhookConfig) *WebhookNotifier {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &WebhookNotifier{
		client: &http.Client{Timeout: config.Timeout},
		config: config,
	}
}

func (n *WebhookNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	payload := map[string]interface{}{
		"alert_id":     alert.ID,
		"rule_name":    alert.RuleName,
		"severity":     alert.Severity,
		"message":      alert.Message,
		"triggered_at": alert.TriggeredAt,
		"tenant_id":    alert.TenantID,
		"status":       alert.Status,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.config.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		zap.S().Error("Failed to send webhook notification", zap.Error(err))
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	zap.S().Info("Webhook notification sent",
		zap.String("url", n.config.URL),
		zap.Int("status", resp.StatusCode))

	return nil
}

func (n *WebhookNotifier) Name() string {
	return "webhook"
}