package domain

import (
	"time"
)

// LoyaltyRule defines the logic for awarding points based on usage thresholds.
type LoyaltyRule struct {
	ID                   int64     `json:"id,string" form:"id"`
	TenantID             int64     `gorm:"index" json:"tenant_id" form:"tenant_id"` // Tenant/Provider ID

	DataThreshold        int64     `json:"data_threshold"` // bytes
	TimeThreshold        int64     `json:"time_threshold"` // seconds

	RequireBoth          bool      `json:"require_both"`  // true = AND logic, false = OR logic

	PointsAwarded        int64     `json:"points_awarded"`

	CreatedAt            time.Time `json:"created_at"`
}

func (LoyaltyRule) TableName() string {
	return "loyalty_rule"
}
