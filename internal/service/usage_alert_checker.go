package service

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

type UserUsage struct {
	UserID    int64
	DataUsed  int64
	DataQuota int64
	TimeUsed  int64
}

type Notifier interface {
	SendUsageAlertEmail(data *NotificationData) error
	SendUsageAlertSMS(data *NotificationData) error
}

type UsageAlertChecker struct {
	notifier Notifier
}

func NewUsageAlertChecker(notifier Notifier) *UsageAlertChecker {
	return &UsageAlertChecker{notifier: notifier}
}

func (c *UsageAlertChecker) CheckUserThresholds(
	user *domain.RadiusUser,
	usage *UserUsage,
	pref *domain.NotificationPreference,
) []*domain.UsageAlert {
	if usage.DataQuota == 0 {
		return nil
	}

	percent := int((float64(usage.DataUsed) / float64(usage.DataQuota)) * 100)
	alerts := make([]*domain.UsageAlert, 0)

	if pref.EmailEnabled {
		for _, threshold := range pref.GetEmailThresholds() {
			if percent >= threshold && percent < threshold+10 {
				alert := &domain.UsageAlert{
					UserID:    user.ID,
					Threshold: threshold,
					AlertType: "email",
				}
				alerts = append(alerts, alert)
			}
		}
	}

	if pref.SMSEnabled {
		for _, threshold := range pref.GetSMSThresholds() {
			if percent >= threshold && percent < threshold+10 {
				alert := &domain.UsageAlert{
					UserID:    user.ID,
					Threshold: threshold,
					AlertType: "sms",
				}
				alerts = append(alerts, alert)
			}
		}
	}

	return alerts
}

func (c *UsageAlertChecker) SendAlert(user *domain.RadiusUser, alert *domain.UsageAlert, usage *UserUsage) error {
	usedGB := float64(usage.DataUsed) / (1024 * 1024 * 1024)
	quotaGB := float64(usage.DataQuota) / (1024 * 1024 * 1024)

	data := &NotificationData{
		Email:     user.Email,
		Phone:     user.Mobile,
		Username:  user.Username,
		Threshold: alert.Threshold,
		UsedGB:    usedGB,
		QuotaGB:   quotaGB,
	}

	if alert.AlertType == "email" {
		return c.notifier.SendUsageAlertEmail(data)
	} else if alert.AlertType == "sms" {
		return c.notifier.SendUsageAlertSMS(data)
	}

	return fmt.Errorf("unknown alert type: %s", alert.AlertType)
}
