package domain

import (
	"time"
)

// Product represents a commercial product/plan that wraps a RadiusProfile.
type Product struct {
	ID              int64     `json:"id,string" form:"id"`
	TenantID        int64     `gorm:"index" json:"tenant_id" form:"tenant_id"` // Tenant/Provider ID
	RadiusProfileID int64     `json:"radius_profile_id,string" form:"radius_profile_id"`
	Name            string    `json:"name" form:"name"`
	Price           float64   `json:"price" form:"price"`
	CostPrice       float64   `json:"cost_price" form:"cost_price"`
	UpRate          int       `json:"up_rate" form:"up_rate"`
	DownRate        int       `json:"down_rate" form:"down_rate"`
	DataQuota       int64     `json:"data_quota" form:"data_quota"`                    // Data quota in MB (0 = unlimited)
	TimeQuota       int64     `json:"time_quota" form:"time_quota"`                    // Time quota in seconds (0 = unlimited)
	ValiditySeconds int64     `json:"validity_seconds" form:"validity_seconds"`          // Account validity period in seconds
	IdleTimeout     int       `json:"idle_timeout" form:"idle_timeout"`                  // Inactivity timeout in seconds
	SessionTimeout  int       `json:"session_timeout" form:"session_timeout"`            // Max session duration in seconds
	Status          string    `json:"status" form:"status"`
	Color           string    `json:"color" form:"color"`
	Remark          string    `json:"remark" form:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Product) TableName() string {
	return "product"
}
