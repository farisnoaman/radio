package service

import (
	"fmt"
)

type EmailProvider interface {
	SendEmail(to, subject, body string) error
}

type SMSProvider interface {
	SendSMS(to, message string) error
}

type NotificationData struct {
	Email     string
	Phone     string
	Username  string
	Threshold int
	UsedGB    float64
	QuotaGB   float64
}

func (d *NotificationData) UsedPercent() float64 {
	if d.QuotaGB == 0 {
		return 0
	}
	return (d.UsedGB / d.QuotaGB) * 100
}

type NotificationService struct {
	emailProvider EmailProvider
	smsProvider   SMSProvider
}

func NewNotificationService(email EmailProvider, sms SMSProvider) *NotificationService {
	return &NotificationService{
		emailProvider: email,
		smsProvider:   sms,
	}
}

func (s *NotificationService) SendUsageAlertEmail(data *NotificationData) error {
	if s.emailProvider == nil {
		return fmt.Errorf("email provider not configured")
	}

	subject := fmt.Sprintf("Usage Alert: %d%% Data Quota Used", data.Threshold)

	remaining := data.QuotaGB - data.UsedGB
	if remaining < 0 {
		remaining = 0
	}

	body := fmt.Sprintf(`Dear %s,

You have used %d%% of your monthly data quota.

Usage Details:
- Data Used: %.2f GB
- Monthly Quota: %.2f GB
- Remaining: %.2f GB

Please monitor your usage to avoid service interruption.

Best regards,
Your Service Provider
`, data.Username, data.Threshold, data.UsedGB, data.QuotaGB, remaining)

	return s.emailProvider.SendEmail(data.Email, subject, body)
}

func (s *NotificationService) SendUsageAlertSMS(data *NotificationData) error {
	if s.smsProvider == nil {
		return fmt.Errorf("SMS provider not configured")
	}

	message := fmt.Sprintf("Alert: You've used %d%% of your data quota (%.2f/%.2f GB). Login to portal for details.",
		data.Threshold, data.UsedGB, data.QuotaGB)

	return s.smsProvider.SendSMS(data.Phone, message)
}
