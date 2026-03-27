package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

type ProviderNotificationService struct {
	db            *gorm.DB
	emailProvider EmailProvider
}

func NewProviderNotificationService(db *gorm.DB, emailProvider EmailProvider) *ProviderNotificationService {
	return &ProviderNotificationService{db: db, emailProvider: emailProvider}
}

type NotificationCheck struct {
	CurrentUsagePercent float64
	TotalUsers         int
	ActiveUsers        int
	DataUsedGB         float64
}

func (s *ProviderNotificationService) CheckThresholds(providerID int64, check NotificationCheck) error {
	pref, err := s.getOrCreatePreferences(providerID)
	if err != nil {
		return err
	}

	if pref.AlertPercentagesEnabled {
		if err := s.checkPercentageThresholds(providerID, pref, check.CurrentUsagePercent); err != nil {
			return err
		}
	}

	if pref.AbsoluteAlertsEnabled {
		if err := s.checkAbsoluteThresholds(providerID, pref, check); err != nil {
			return err
		}
	}

	if pref.AnomalyDetectionEnabled {
		if err := s.checkAnomalyDetection(providerID, pref, check); err != nil {
			return err
		}
	}

	return nil
}

func (s *ProviderNotificationService) checkPercentageThresholds(providerID int64, pref *domain.ProviderNotificationPreference, usagePercent float64) error {
	thresholds := parseThresholds(pref.AlertPercentages)

	for _, threshold := range thresholds {
		if usagePercent >= float64(threshold) {
			if !s.shouldSendAlert(providerID, "percentage", threshold) {
				continue
			}

			subject := fmt.Sprintf("Usage Alert: %d%% of Plan Limit", threshold)
			body := fmt.Sprintf("Your provider has used %.1f%% of its plan limit.", usagePercent)

			if s.emailProvider != nil && pref.EmailEnabled {
				s.emailProvider.SendEmail("admin@provider.local", subject, body)
			}

			s.recordAlertSent(providerID, "percentage", threshold)
		}
	}

	return nil
}

func (s *ProviderNotificationService) checkAbsoluteThresholds(providerID int64, pref *domain.ProviderNotificationPreference, check NotificationCheck) error {
	if pref.MaxUsersThreshold > 0 && check.TotalUsers >= pref.MaxUsersThreshold {
		if s.shouldSendAlert(providerID, "max_users", pref.MaxUsersThreshold) {
			subject := fmt.Sprintf("User Limit Alert: %d Users", check.TotalUsers)
			if s.emailProvider != nil && pref.EmailEnabled {
				s.emailProvider.SendEmail("admin@provider.local", subject, "You have reached your maximum user limit.")
			}
			s.recordAlertSent(providerID, "max_users", pref.MaxUsersThreshold)
		}
	}

	return nil
}

func (s *ProviderNotificationService) checkAnomalyDetection(providerID int64, pref *domain.ProviderNotificationPreference, check NotificationCheck) error {
	threshold := pref.AnomalyThresholdPercent

	if check.CurrentUsagePercent >= float64(threshold) {
		if s.shouldSendAlert(providerID, "anomaly", threshold) {
			subject := "Anomaly Detection Alert"
			body := fmt.Sprintf("Unusual activity detected: %.1f%% usage deviation from baseline.", check.CurrentUsagePercent)
			if s.emailProvider != nil && pref.EmailEnabled {
				s.emailProvider.SendEmail("admin@provider.local", subject, body)
			}
			s.recordAlertSent(providerID, "anomaly", threshold)
		}
	}

	return nil
}

func (s *ProviderNotificationService) shouldSendAlert(providerID int64, alertType string, threshold int) bool {
	var count int64
	yesterday := time.Now().Add(-24 * time.Hour)
	s.db.Model(&domain.UsageAlert{}).
		Where("user_id = ? AND alert_type = ? AND threshold = ? AND sent_at > ?",
			providerID, alertType, threshold, yesterday).
		Count(&count)
	return count == 0
}

func (s *ProviderNotificationService) recordAlertSent(providerID int64, alertType string, threshold int) {
	alert := &domain.UsageAlert{
		UserID:    providerID,
		Threshold: threshold,
		AlertType: alertType,
	}
	now := time.Now()
	alert.SentAt = &now
	s.db.Create(alert)
}

func (s *ProviderNotificationService) getOrCreatePreferences(providerID int64) (*domain.ProviderNotificationPreference, error) {
	var pref domain.ProviderNotificationPreference
	err := s.db.Where("provider_id = ?", providerID).First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		pref = domain.ProviderNotificationPreference{
			ProviderID:              providerID,
			AlertPercentages:        "70,85,100",
			AlertPercentagesEnabled: true,
		}
		err = s.db.Create(&pref).Error
	}

	return &pref, err
}

func (s *ProviderNotificationService) GetPreferences(providerID int64) (*domain.ProviderNotificationPreference, error) {
	var pref domain.ProviderNotificationPreference
	err := s.db.Where("provider_id = ?", providerID).First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		pref = domain.ProviderNotificationPreference{
			ProviderID:              providerID,
			AlertPercentages:        "70,85,100",
			AlertPercentagesEnabled: true,
		}
		err = s.db.Create(&pref).Error
	}

	return &pref, err
}

type NotificationPreferenceUpdate struct {
	AlertPercentages            string `json:"alert_percentages"`
	AlertPercentagesEnabled    bool   `json:"alert_percentages_enabled"`
	MaxUsersThreshold          int    `json:"max_users_threshold"`
	AbsoluteAlertsEnabled      bool   `json:"absolute_alerts_enabled"`
	AnomalyDetectionEnabled    bool   `json:"anomaly_detection_enabled"`
	AnomalyThresholdPercent    int    `json:"anomaly_threshold_percent"`
	EmailEnabled               bool   `json:"email_enabled"`
	SMSEnabled                bool   `json:"sms_enabled"`
}

func (s *ProviderNotificationService) UpdatePreferences(providerID int64, req *NotificationPreferenceUpdate) error {
	pref, err := s.GetPreferences(providerID)
	if err != nil {
		return err
	}

	pref.AlertPercentages = req.AlertPercentages
	pref.AlertPercentagesEnabled = req.AlertPercentagesEnabled
	pref.MaxUsersThreshold = req.MaxUsersThreshold
	pref.AbsoluteAlertsEnabled = req.AbsoluteAlertsEnabled
	pref.AnomalyDetectionEnabled = req.AnomalyDetectionEnabled
	pref.AnomalyThresholdPercent = req.AnomalyThresholdPercent
	pref.EmailEnabled = req.EmailEnabled
	pref.SMSEnabled = req.SMSEnabled

	return s.db.Save(pref).Error
}

func parseThresholds(s string) []int {
	if s == "" {
		return []int{}
	}
	var result []int
	for _, part := range strings.Split(s, ",") {
		val, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && val > 0 {
			result = append(result, val)
		}
	}
	return result
}
