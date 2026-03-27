package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

type FraudDetectionService struct {
	db *gorm.DB
}

func NewFraudDetectionService(db *gorm.DB) *FraudDetectionService {
	return &FraudDetectionService{db: db}
}

type FraudRule struct {
	Name      string
	Threshold int
	Window    time.Duration
	Action    string
}

var fraudRules = []FraudRule{
	{Name: "ip_activation_limit", Threshold: 5, Window: time.Hour, Action: "flag"},
	{Name: "same_voucher_multi_use", Threshold: 2, Window: time.Hour, Action: "quarantine"},
	{Name: "rapid_successive", Threshold: 10, Window: time.Minute, Action: "block"},
}

func (s *FraudDetectionService) CheckAndRecord(providerID, voucherID, userID int64, ipAddress string) ([]string, error) {
	var triggeredRules []string

	for _, rule := range fraudRules {
		count, err := s.countRecentEvents(providerID, ipAddress, voucherID, rule)
		if err != nil {
			return nil, err
		}

		if count >= int64(rule.Threshold) {
			triggeredRules = append(triggeredRules, rule.Name)

			details, _ := json.Marshal(map[string]interface{}{
				"rule":      rule.Name,
				"count":     count,
				"threshold": rule.Threshold,
				"voucher_id": voucherID,
				"user_id":   userID,
			})

			fraudLog := &domain.FraudLog{
				ProviderID: providerID,
				VoucherID:   voucherID,
				UserID:     userID,
				IPAddress:  ipAddress,
				EventType:  rule.Name,
				Details:    string(details),
			}
			s.db.Create(fraudLog)

			switch rule.Action {
			case "quarantine":
				s.quarantineVoucher(voucherID)
			case "block":
				s.blockIP(providerID, ipAddress)
			}
		}
	}

	return triggeredRules, nil
}

func (s *FraudDetectionService) countRecentEvents(providerID int64, ipAddress string, voucherID int64, rule FraudRule) (int64, error) {
	since := time.Now().Add(-rule.Window)

	var count int64
	switch rule.Name {
	case "ip_activation_limit":
		s.db.Model(&domain.FraudLog{}).
			Where("provider_id = ? AND ip_address = ? AND created_at > ?", providerID, ipAddress, since).
			Count(&count)
	case "same_voucher_multi_use":
		s.db.Model(&domain.FraudLog{}).
			Where("provider_id = ? AND voucher_id = ? AND event_type = ? AND created_at > ?",
				providerID, voucherID, "same_voucher_multi_use", since).
			Count(&count)
	case "rapid_successive":
		s.db.Model(&domain.FraudLog{}).
			Where("provider_id = ? AND ip_address = ? AND created_at > ?", providerID, ipAddress, since).
			Count(&count)
	}

	return count, nil
}

func (s *FraudDetectionService) quarantineVoucher(voucherID int64) {
	s.db.Model(&domain.Voucher{}).Where("id = ?", voucherID).Update("status", "quarantined")
}

func (s *FraudDetectionService) blockIP(providerID int64, ipAddress string) {
	fmt.Printf("Blocking IP %s for provider %d\n", ipAddress, providerID)
}

func (s *FraudDetectionService) GetFraudLogs(providerID int64, limit int) ([]domain.FraudLog, error) {
	var logs []domain.FraudLog
	err := s.db.Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
