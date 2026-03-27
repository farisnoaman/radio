package domain

import (
	"time"
)

// LoyaltyProfile represents a persistent aggregate of user usage and rewards.
// It survives voucher cleanup and tracks long-term loyalty metrics.
type LoyaltyProfile struct {
	ID                int64     `json:"id,string" form:"id"`
	TenantID          int64     `gorm:"index" json:"tenant_id" form:"tenant_id"` // Tenant/Provider ID

	// Identity
	IdentityKey       string    `gorm:"uniqueIndex" json:"identity_key"` // hash(MAC + TenantID)
	MacAddress        string    `gorm:"index" json:"mac_address"`        // Primary observed MAC

	// Lifetime metrics (never reset)
	TotalDataUsed     int64     `json:"total_data_used"` // Bytes
	TotalTimeUsed     int64     `json:"total_time_used"` // Seconds

	// Milestone tracking (used for rewards, threshold subtracted on award)
	MilestoneDataUsed int64     `json:"milestone_data_used"` // Bytes
	MilestoneTimeUsed int64     `json:"milestone_time_used"` // Seconds

	// Rewards
	Points            int64      `json:"points"`

	// Fraud prevention
	LastRedeemAt      *time.Time `json:"last_redeem_at"` // Tracks last redemption for daily rate limiting
	DailyRedeemCount  int        `json:"daily_redeem_count"` // Redemptions today (reset nightly)

	// Points expiry
	PointsExpiresAt   *time.Time `json:"points_expires_at"` // nil = never expire

	// Cached tier/badge
	Badge             string     `json:"badge"` // "Bronze", "Silver", "Gold", etc.

	// Concurrency control
	Version           int64      `gorm:"default:0" json:"version"` // Optimistic locking

	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (LoyaltyProfile) TableName() string {
	return "loyalty_profile"
}
