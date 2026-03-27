package domain

import (
	"time"
)

// LoyaltyIdentity provides a mapping between MAC addresses and loyalty profiles.
// This allows a single user (loyalty profile) to have multiple devices.
type LoyaltyIdentity struct {
	ID           int64     `json:"id,string" form:"id"`
	ProfileID    int64     `gorm:"index" json:"profile_id,string" form:"profile_id"`

	MacAddress   string    `gorm:"uniqueIndex" json:"mac_address"` // MAC address
	DeviceHash   string    `json:"device_hash"`                    // Optional fingerprint

	CreatedAt    time.Time `json:"created_at"`
}

func (LoyaltyIdentity) TableName() string {
	return "loyalty_identity"
}
