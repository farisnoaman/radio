package domain

import "time"

// VoucherTemplate represents a custom HTML template for printing vouchers.
// Templates can be per-user (private) or shared publicly.
//
// Template content supports variable placeholders:
//   - {{code}}      — voucher code
//   - {{price}}     — voucher price
//   - {{validity}}  — formatted validity period
//   - {{hotspot}}   — hotspot/network name
//   - {{link}}      — login link URL
//   - {{serial}}    — serial number (batchId-voucherId)
//   - {{qr}}        — QR code image tag
//   - {{product}}   — product name
//
// Database table: voucher_template
type VoucherTemplate struct {
	ID        int64     `json:"id,string" form:"id"`
	Name      string    `json:"name" form:"name" gorm:"not null"`
	Content   string    `json:"content" form:"content" gorm:"type:text;not null"` // HTML template content
	OwnerID   int64     `gorm:"index" json:"owner_id,string" form:"owner_id"`     // User who created this template
	IsPublic  bool      `json:"is_public" form:"is_public" gorm:"default:false"`  // Whether other users can see/use it
	IsDefault bool      `json:"is_default" form:"is_default" gorm:"default:false"` // System-provided default template
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (VoucherTemplate) TableName() string {
	return "voucher_template"
}
