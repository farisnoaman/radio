package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

type MockNotifier struct{}

func (m *MockNotifier) SendUsageAlertEmail(data *NotificationData) error {
	return nil
}

func (m *MockNotifier) SendUsageAlertSMS(data *NotificationData) error {
	return nil
}

func TestUsageAlertChecker_CheckUserThresholds(t *testing.T) {
	user := &domain.RadiusUser{
		ID:       1,
		Username: "testuser",
		Email:    "user@example.com",
	}

	usage := &UserUsage{
		UserID:    1,
		DataUsed:  8.5 * 1024 * 1024 * 1024,
		DataQuota: 10 * 1024 * 1024 * 1024,
	}

	pref := &domain.NotificationPreference{
		UserID:          1,
		EmailEnabled:    true,
		EmailThresholds: "80,90,100",
	}

	notifier := &MockNotifier{}
	checker := NewUsageAlertChecker(notifier)

	alerts := checker.CheckUserThresholds(user, usage, pref)

	assert.Len(t, alerts, 1)
	assert.Equal(t, 80, alerts[0].Threshold)
	assert.Equal(t, "email", alerts[0].AlertType)
}

func TestUsageAlertChecker_MultipleThresholds(t *testing.T) {
	user := &domain.RadiusUser{
		ID:       1,
		Username: "testuser",
		Email:    "user@example.com",
	}

	usage := &UserUsage{
		UserID:    1,
		DataUsed:  8589934592, // 8 GB = 80%
		DataQuota: 10 * 1024 * 1024 * 1024,
	}

	pref := &domain.NotificationPreference{
		UserID:          1,
		EmailEnabled:    true,
		EmailThresholds: "80,90,100",
	}

	notifier := &MockNotifier{}
	checker := NewUsageAlertChecker(notifier)

	alerts := checker.CheckUserThresholds(user, usage, pref)

	// 8/10 = 80%, should trigger 80% threshold
	assert.Len(t, alerts, 1)
	assert.Equal(t, 80, alerts[0].Threshold)
}

func TestUsageAlertChecker_SMSThreshold(t *testing.T) {
	user := &domain.RadiusUser{
		ID:       1,
		Username: "testuser",
		Email:    "user@example.com",
		Mobile:   "+1234567890",
	}

	usage := &UserUsage{
		UserID:    1,
		DataUsed:  10860378112, // 10.1 GB in bytes
		DataQuota: 10 * 1024 * 1024 * 1024,
	}

	pref := &domain.NotificationPreference{
		UserID:          1,
		EmailEnabled:    false,
		SMSEnabled:      true,
		SMSThresholds:   "100",
	}

	notifier := &MockNotifier{}
	checker := NewUsageAlertChecker(notifier)

	alerts := checker.CheckUserThresholds(user, usage, pref)

	assert.Len(t, alerts, 1)
	assert.Equal(t, 100, alerts[0].Threshold)
	assert.Equal(t, "sms", alerts[0].AlertType)
}

func TestUsageAlertChecker_DisabledNotifications(t *testing.T) {
	user := &domain.RadiusUser{
		ID:       1,
		Username: "testuser",
		Email:    "user@example.com",
	}

	usage := &UserUsage{
		UserID:    1,
		DataUsed:  9 * 1024 * 1024 * 1024,
		DataQuota: 10 * 1024 * 1024 * 1024,
	}

	pref := &domain.NotificationPreference{
		UserID:          1,
		EmailEnabled:    false,
		SMSEnabled:      false,
	}

	notifier := &MockNotifier{}
	checker := NewUsageAlertChecker(notifier)

	alerts := checker.CheckUserThresholds(user, usage, pref)

	assert.Len(t, alerts, 0)
}

func TestUsageAlertChecker_ZeroQuota(t *testing.T) {
	user := &domain.RadiusUser{
		ID:       1,
		Username: "testuser",
		Email:    "user@example.com",
	}

	usage := &UserUsage{
		UserID:    1,
		DataUsed:  0,
		DataQuota: 0,
	}

	pref := &domain.NotificationPreference{
		UserID:          1,
		EmailEnabled:    true,
		EmailThresholds: "80,90,100",
	}

	notifier := &MockNotifier{}
	checker := NewUsageAlertChecker(notifier)

	alerts := checker.CheckUserThresholds(user, usage, pref)

	assert.Len(t, alerts, 0)
}
