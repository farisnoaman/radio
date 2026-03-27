package domain

import (
	"time"
)

// LoyaltyConfig defines global or tenant-specific settings for point systems and anti-abuse.
type LoyaltyConfig struct {
	ID                       int64     `json:"id,string" form:"id"`
	TenantID                 int64     `gorm:"uniqueIndex" json:"tenant_id" form:"tenant_id"`

	MaxRedemptionsPerDay     int       `json:"max_redemptions_per_day"`
	MaxActiveRewards         int       `json:"max_active_rewards"`

	PointsExpiryDays         int       `json:"points_expiry_days"` // e.g. 90

	UpdatedAt                time.Time `json:"updated_at"`
}

func (LoyaltyConfig) TableName() string {
	return "loyalty_config"
}
