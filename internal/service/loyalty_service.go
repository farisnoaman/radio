package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

type UsageEvent struct {
	TenantID  int64
	Mac       string
	DataUsed  int64 // Bytes
	TimeUsed  int64 // Seconds
	Timestamp time.Time
}

type LoyaltyService struct {
	db *gorm.DB
}

func NewLoyaltyService(db *gorm.DB) *LoyaltyService {
	return &LoyaltyService{db: db}
}

// GenerateIdentityKey creates a unique hash for a MAC and Tenant combination.
func (s *LoyaltyService) GenerateIdentityKey(mac string, tenantID int64) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d", mac, tenantID)))
	return hex.EncodeToString(h.Sum(nil))
}

// ProcessUsageEvent handles real-time usage updates and awards points.
func (s *LoyaltyService) ProcessUsageEvent(ctx context.Context, event UsageEvent) error {
	if event.Mac == "" {
		return nil
	}

	identityKey := s.GenerateIdentityKey(event.Mac, event.TenantID)

	return s.db.Transaction(func(tx *gorm.DB) error {
		var profile domain.LoyaltyProfile
		err := tx.Where("identity_key = ?", identityKey).First(&profile).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			profile = domain.LoyaltyProfile{
				TenantID:    event.TenantID,
				IdentityKey: identityKey,
				MacAddress:  event.Mac,
				Badge:       "None",
			}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
			// Also create initial identity mapping
			identity := domain.LoyaltyIdentity{
				ProfileID:  profile.ID,
				MacAddress: event.Mac,
			}
			tx.Create(&identity)
		} else if err != nil {
			return err
		}

		// Atomic Update with Optimistic Locking
		// total_data_used and total_time_used are lifetime counters.
		// milestone counters are for reward calculation.
		result := tx.Model(&profile).
			Where("id = ? AND version = ?", profile.ID, profile.Version).
			Updates(map[string]interface{}{
				"total_data_used":     gorm.Expr("total_data_used + ?", event.DataUsed),
				"total_time_used":     gorm.Expr("total_time_used + ?", event.TimeUsed),
				"milestone_data_used": gorm.Expr("milestone_data_used + ?", event.DataUsed),
				"milestone_time_used": gorm.Expr("milestone_time_used + ?", event.TimeUsed),
				"version":             gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("loyalty profile optimistic lock failure")
		}

		// Refresh profile to check rules
		if err := tx.First(&profile, profile.ID).Error; err != nil {
			return err
		}

		return s.applyRules(tx, &profile)
	})
}

func (s *LoyaltyService) applyRules(tx *gorm.DB, profile *domain.LoyaltyProfile) error {
	var rules []domain.LoyaltyRule
	if err := tx.Where("tenant_id = ?", profile.TenantID).Find(&rules).Error; err != nil {
		return err
	}

	// Default Rule if none configured
	if len(rules) == 0 {
		rules = []domain.LoyaltyRule{
			{
				DataThreshold: 250 * 1024 * 1024 * 1024, // 250 GB
				TimeThreshold: 100 * 3600,              // 100 Hours
				RequireBoth:   false,                   // OR logic by default for 250GB
				PointsAwarded: 20,
			},
			{
				DataThreshold: 100 * 1024 * 1024 * 1024, // 100 GB
				TimeThreshold: 100 * 3600,              // 100 Hours
				RequireBoth:   true,                    // AND logic
				PointsAwarded: 20,
			},
		}
	}

	awardedPoints := int64(0)
	for _, rule := range rules {
		condition := false
		if rule.RequireBoth {
			condition = profile.MilestoneDataUsed >= rule.DataThreshold && profile.MilestoneTimeUsed >= rule.TimeThreshold
		} else {
			// If rule only has one threshold set, use it. Otherwise OR.
			if rule.DataThreshold > 0 && rule.TimeThreshold > 0 {
				condition = profile.MilestoneDataUsed >= rule.DataThreshold || profile.MilestoneTimeUsed >= rule.TimeThreshold
			} else if rule.DataThreshold > 0 {
				condition = profile.MilestoneDataUsed >= rule.DataThreshold
			} else {
				condition = profile.MilestoneTimeUsed >= rule.TimeThreshold
			}
		}

		if condition {
			awardedPoints += rule.PointsAwarded
			// Subtract threshold to preserve overflow
			tx.Model(profile).Updates(map[string]interface{}{
				"milestone_data_used": gorm.Expr("milestone_data_used - ?", rule.DataThreshold),
				"milestone_time_used": gorm.Expr("milestone_time_used - ?", rule.TimeThreshold),
			})
		}
	}

	if awardedPoints > 0 {
		tx.Model(profile).Update("points", gorm.Expr("points + ?", awardedPoints))
	}

	// Badge Calculation
	newBadge := s.calculateBadge(profile.TotalDataUsed, profile.TotalTimeUsed)
	if newBadge != profile.Badge {
		tx.Model(profile).Update("badge", newBadge)
	}

	return nil
}

func (s *LoyaltyService) calculateBadge(totalData, totalTime int64) string {
	// Totals are in Bytes and Seconds
	gb := totalData / (1024 * 1024 * 1024)
	hours := totalTime / 3600

	if gb >= 1000 && hours >= 200 {
		return "Gold"
	}
	if gb >= 500 && hours >= 70 {
		return "Silver"
	}
	if gb >= 200 && hours >= 50 {
		return "Bronze"
	}
	return "None"
}

// Redeem converts points into a reward voucher.
func (s *LoyaltyService) Redeem(ctx context.Context, profileID int64, rewardType string) (*domain.Voucher, error) {
	// This will be implemented when we have the reward definitions and voucher generation logic.
	return nil, errors.New("not implemented")
}
