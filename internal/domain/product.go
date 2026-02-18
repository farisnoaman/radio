package domain

import (
	"time"
)

// Product Commercial product/plan wrapping a partial RadiusProfile
type Product struct {
	ID              int64     `json:"id,string" form:"id"`
	RadiusProfileID int64     `json:"radius_profile_id,string" form:"radius_profile_id"` // Technical profile template
	Name            string    `json:"name" form:"name"`
	Price           float64   `json:"price" form:"price"`                     // Retail price
	CostPrice       float64   `json:"cost_price" form:"cost_price"`           // Cost to agent
	UpRate          int       `json:"up_rate" form:"up_rate"`                 // Upload rate in Kbps
	DownRate        int       `json:"down_rate" form:"down_rate"`             // Download rate in Kbps
	DataQuota       int64     `json:"data_quota" form:"data_quota"`           // Data quota in MB (0 = unlimited)
	ValiditySeconds int64     `json:"validity_seconds" form:"validity_seconds"` // Validity duration (0 = unlimited)
	Status          string    `json:"status" form:"status"`                   // enabled, disabled
	Color           string    `json:"color" form:"color"`                     // Display color
	Remark          string    `json:"remark" form:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Product) TableName() string {
	return "product"
}
