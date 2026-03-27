package domain

import (
	"time"
)

// UsageAlert tracks when usage threshold alerts were sent to users
type UsageAlert struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	UserID    int64      `json:"user_id" gorm:"index"`
	Threshold int        `json:"threshold"` // 80, 90, 100
	AlertType string     `json:"alert_type"` // email, sms
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User *RadiusUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for UsageAlert
func (UsageAlert) TableName() string {
	return "usage_alerts"
}

// CanSendAlert checks if alert hasn't been sent in the last 24 hours
func (a *UsageAlert) CanSendAlert() bool {
	if a.SentAt == nil {
		return true
	}

	// Don't send same alert type for same threshold within 24 hours
	hoursSinceLastSend := time.Since(*a.SentAt).Hours()
	return hoursSinceLastSend >= 24
}

// MarkAsSent marks the alert as sent
func (a *UsageAlert) MarkAsSent() {
	now := time.Now()
	a.SentAt = &now
}
