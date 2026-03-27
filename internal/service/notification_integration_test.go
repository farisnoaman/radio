package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

type mockEmailProvider struct {
	sent []struct {
		to      string
		subject string
		body    string
	}
}

func (m *mockEmailProvider) SendEmail(to, subject, body string) error {
	m.sent = append(m.sent, struct {
		to      string
		subject string
		body    string
	}{to, subject, body})
	return nil
}

func TestUsageAlertChecker_E2E_WithMockProvider(t *testing.T) {
	mockEmail := &mockEmailProvider{}
	notifier := &mockNotifier{emailProvider: mockEmail}
	checker := NewUsageAlertChecker(notifier)

	user := &domain.RadiusUser{
		ID:       1,
		Username: "testuser",
		Email:    "user@example.com",
		DataQuota: 10240,
	}

	usage := &UserUsage{
		UserID:    1,
		DataUsed:  8589934592,
		DataQuota: 10737418240,
	}

	pref := &domain.NotificationPreference{
		UserID:          1,
		EmailEnabled:    true,
		EmailThresholds: "80,90,100",
	}

	alerts := checker.CheckUserThresholds(user, usage, pref)

	assert.Len(t, alerts, 1)
	assert.Equal(t, 80, alerts[0].Threshold)
	assert.Equal(t, "email", alerts[0].AlertType)

	err := checker.SendAlert(user, alerts[0], usage)
	assert.NoError(t, err)
	assert.Len(t, mockEmail.sent, 1)
	assert.Equal(t, "user@example.com", mockEmail.sent[0].to)
}

type mockNotifier struct {
	emailProvider *mockEmailProvider
}

func (m *mockNotifier) SendUsageAlertEmail(data *NotificationData) error {
	subject := "Usage Alert"
	body := data.Username
	m.emailProvider.sent = append(m.emailProvider.sent, struct {
		to      string
		subject string
		body    string
	}{data.Email, subject, body})
	return nil
}

func (m *mockNotifier) SendUsageAlertSMS(data *NotificationData) error {
	return nil
}

func TestNotificationPreference_GetThresholds(t *testing.T) {
	pref := &domain.NotificationPreference{
		EmailThresholds: "80,90,100",
		SMSThresholds:   "100",
	}

	emailThresholds := pref.GetEmailThresholds()
	assert.Len(t, emailThresholds, 3)
	assert.Contains(t, emailThresholds, 80)
	assert.Contains(t, emailThresholds, 90)
	assert.Contains(t, emailThresholds, 100)

	smsThresholds := pref.GetSMSThresholds()
	assert.Len(t, smsThresholds, 1)
	assert.Contains(t, smsThresholds, 100)
}

func TestNotificationPreference_ShouldSendAt(t *testing.T) {
	pref := &domain.NotificationPreference{
		EmailEnabled:    true,
		EmailThresholds: "80,90,100",
	}

	assert.True(t, pref.ShouldSendEmailAt(80))
	assert.False(t, pref.ShouldSendEmailAt(85))
	assert.False(t, pref.ShouldSendEmailAt(75))
	assert.True(t, pref.ShouldSendEmailAt(90))
	assert.True(t, pref.ShouldSendEmailAt(100))

	pref.EmailEnabled = false
	assert.False(t, pref.ShouldSendEmailAt(80))
}

func TestUsageAlert_CanSendAlert(t *testing.T) {
	alert := &domain.UsageAlert{
		UserID:    1,
		Threshold: 80,
		AlertType: "email",
	}

	assert.True(t, alert.CanSendAlert())

	now := time.Now()
	alert.SentAt = &now
	assert.False(t, alert.CanSendAlert())
}

func TestUsageAlert_MarkAsSent(t *testing.T) {
	alert := &domain.UsageAlert{
		UserID:    1,
		Threshold: 80,
		AlertType: "email",
	}

	assert.Nil(t, alert.SentAt)
	alert.MarkAsSent()
	assert.NotNil(t, alert.SentAt)
}

func TestNotificationData_UsedPercent(t *testing.T) {
	data := &NotificationData{
		UsedGB:  8.0,
		QuotaGB: 10.0,
	}

	percent := data.UsedPercent()
	assert.InDelta(t, 80.0, percent, 0.1)
}

func TestNotificationData_UsedPercent_ZeroQuota(t *testing.T) {
	data := &NotificationData{
		UsedGB:  5.0,
		QuotaGB: 0,
	}

	percent := data.UsedPercent()
	assert.Equal(t, 0.0, percent)
}
