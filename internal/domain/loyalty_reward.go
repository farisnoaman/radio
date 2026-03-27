package domain

import (
	"encoding/json"
	"time"
)

// LoyaltyReward represents a redemption event or reward choice for a loyalty profile.
type LoyaltyReward struct {
	ID           int64           `json:"id,string" form:"id"`
	ProfileID    int64           `gorm:"index" json:"profile_id,string" form:"profile_id"`

	RewardType   string          `json:"reward_type" form:"reward_type"` // e.g., "10GB_30D"
	PointsCost   int64           `json:"points_cost" form:"points_cost"`

	// Flexible metadata for reward configuration
	Metadata     json.RawMessage `gorm:"type:jsonb" json:"metadata"`

	RedeemedAt   time.Time       `json:"redeemed_at"`
}

func (LoyaltyReward) TableName() string {
	return "loyalty_reward"
}
