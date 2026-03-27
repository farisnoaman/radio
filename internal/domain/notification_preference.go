package domain

import (
	"time"
)

// NotificationPreference defines user's notification settings
type NotificationPreference struct {
	ID                  int64     `json:"id" gorm:"primaryKey"`
	UserID              int64     `json:"user_id" gorm:"uniqueIndex"`
	EmailEnabled        bool      `json:"email_enabled" gorm:"default:true"`
	SMSEnabled          bool      `json:"sms_enabled" gorm:"default:false"`
	EmailThresholds     string    `json:"email_thresholds" gorm:"default:'80,90,100'"` // comma-separated
	SMSThresholds       string    `json:"sms_thresholds" gorm:"default:'100'"`           // comma-separated
	DailySummaryEnabled bool      `json:"daily_summary_enabled" gorm:"default:false"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User *RadiusUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for NotificationPreference
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

// GetEmailThresholds parses the comma-separated thresholds
func (p *NotificationPreference) GetEmailThresholds() []int {
	return parseThresholds(p.EmailThresholds)
}

// GetSMSThresholds parses the comma-separated thresholds
func (p *NotificationPreference) GetSMSThresholds() []int {
	return parseThresholds(p.SMSThresholds)
}

// ShouldSendEmailAt checks if email should be sent at given threshold
func (p *NotificationPreference) ShouldSendEmailAt(threshold int) bool {
	if !p.EmailEnabled {
		return false
	}
	thresholds := p.GetEmailThresholds()
	for _, t := range thresholds {
		if t == threshold {
			return true
		}
	}
	return false
}

// ShouldSendSMSAt checks if SMS should be sent at given threshold
func (p *NotificationPreference) ShouldSendSMSAt(threshold int) bool {
	if !p.SMSEnabled {
		return false
	}
	thresholds := p.GetSMSThresholds()
	for _, t := range thresholds {
		if t == threshold {
			return true
		}
	}
	return false
}
