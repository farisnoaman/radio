package domain

import (
	"time"
)

// Product represents a commercial product/plan that wraps a RadiusProfile.
// It defines the billing parameters including data quota, time validity,
// pricing, and bandwidth limits that will be applied to vouchers created
// from this product.
//
// The product serves as a template for voucher allocations - when a batch
// is created, vouchers inherit the DataQuota and ValiditySeconds from
// the selected product, not from the batch itself.
//
// Database table: product
//
// Lifecycle:
//   - Created via Admin API POST /api/v1/products
//   - Used as template when creating voucher batches
//   - Can be disabled but not deleted (vouchers may reference it)
type Product struct {
	ID              int64     `json:"id,string" form:"id"`
	RadiusProfileID int64     `json:"radius_profile_id,string" form:"radius_profile_id"`
	Name            string    `json:"name" form:"name"`
	Price           float64   `json:"price" form:"price"`
	CostPrice       float64   `json:"cost_price" form:"cost_price"`
	UpRate          int       `json:"up_rate" form:"up_rate"`
	DownRate        int       `json:"down_rate" form:"down_rate"`
	DataQuota       int64     `json:"data_quota" form:"data_quota"`
	ValiditySeconds int64     `json:"validity_seconds" form:"validity_seconds"`
	Status          string    `json:"status" form:"status"`
	Color           string    `json:"color" form:"color"`
	Remark          string    `json:"remark" form:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Product) TableName() string {
	return "product"
}
