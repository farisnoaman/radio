package domain

import (
	"time"
)

// VoucherBatch A batch of generated vouchers
type VoucherBatch struct {
	ID           int64      `json:"id,string" form:"id"`
	Name         string     `json:"name" form:"name"`
	ProductID    int64      `gorm:"index" json:"product_id,string" form:"product_id"`
	AgentID      int64      `gorm:"index" json:"agent_id,string" form:"agent_id"` // Nullable if admin generated, but we use int64. 0 means system/admin.
	Count        int        `json:"count" form:"count"`
	Prefix       string     `json:"prefix" form:"prefix"`
	Remark      string     `json:"remark" form:"remark"`
	ExpireTime  *time.Time `json:"expire_time"`
	GeneratePIN bool        `json:"generate_pin" form:"generate_pin"` // Whether to generate PIN for vouchers
	PINLength    int        `json:"pin_length" form:"pin_length"`    // Length of PIN (default 4)
	// First-Use Expiration: voucher expires X days after first login instead of from creation
	ExpirationType string    `json:"expiration_type" form:"expiration_type"` // "fixed" or "first_use"
	ValidityDays   int       `json:"validity_days" form:"validity_days"`     // Days of validity (for first_use type)
	IsDeleted      bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`

}

func (VoucherBatch) TableName() string {
	return "voucher_batch"
}

// Voucher Individual prepaid code
type Voucher struct {
	ID              int64     `json:"id,string" form:"id"`
	BatchID         int64     `gorm:"index" json:"batch_id,string" form:"batch_id"`
	Code            string    `json:"code" gorm:"uniqueIndex" form:"code"` // The username/password
	RadiusUsername  string    `gorm:"index" json:"radius_username" form:"radius_username"` // Populated after activation
	Status          string    `gorm:"index" json:"status" form:"status"` // unused, active, used, expired
	AgentID         int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
	Price           float64   `json:"price" form:"price"`
	ActivatedAt     time.Time `json:"activated_at"`
	ExpireTime      time.Time `gorm:"index" json:"expire_time"`

	ExtendedCount   int       `json:"extended_count"`     // Times extended
	LastExtendedAt  time.Time `json:"last_extended_at"`  // Last extension timestamp
	FirstUsedAt     time.Time `json:"first_used_at"`     // First login timestamp for first_use expiration
	PIN             string    `json:"-"`                  // Hashed PIN (never exposed via JSON)
	RequirePIN      bool      `json:"require_pin"`        // Whether PIN is required for redemption
	IsDeleted       bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Voucher) TableName() string {
	return "voucher"
}

// VoucherTopup represents additional data quota that can be added to an active voucher
type VoucherTopup struct {
	ID              int64     `json:"id,string" form:"id"`
	VoucherID       int64     `json:"voucher_id,string" form:"voucher_id"`
	VoucherCode     string    `gorm:"index" json:"voucher_code" form:"voucher_code"`
	AgentID         int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
	DataQuota       int64     `json:"data_quota" form:"data_quota"`      // Additional data in MB
	TimeQuota       int64     `json:"time_quota" form:"time_quota"`      // Additional time in seconds
	Price           float64   `json:"price" form:"price"`                // Price paid for topup
	Status          string    `gorm:"index" json:"status" form:"status"`               // active, used, expired
	ActivatedAt     time.Time `json:"activated_at"`
	ExpireTime      time.Time `json:"expire_time"`
	IsDeleted       bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`

	UpdatedAt       time.Time `json:"updated_at"`
}

func (VoucherTopup) TableName() string {
	return "voucher_topup"
}

// VoucherSubscription represents a recurring subscription for automatic voucher renewal
type VoucherSubscription struct {
	ID              int64     `json:"id,string" form:"id"`
	VoucherCode     string    `gorm:"index" json:"voucher_code" form:"voucher_code"`
	ProductID       int64     `json:"product_id,string" form:"product_id"`
	AgentID         int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
	IntervalDays    int       `json:"interval_days" form:"interval_days"` // Days between renewals
	Status          string    `gorm:"index" json:"status" form:"status"`                 // active, paused, cancelled

	AutoRenew       bool      `json:"auto_renew" form:"auto_renew"`        // Whether to auto-renew
	LastRenewalAt   time.Time `json:"last_renewal_at"`
	NextRenewalAt   time.Time `json:"next_renewal_at"`
	RenewalCount    int       `json:"renewal_count"`                       // Number of times renewed
	IsDeleted       bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (VoucherSubscription) TableName() string {
	return "voucher_subscription"
}

// VoucherBundle represents a package of multiple vouchers sold together
type VoucherBundle struct {
	ID          int64     `json:"id,string" form:"id"`
	Name        string    `json:"name" form:"name"`
	Description string    `json:"description" form:"description"`
	AgentID     int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
	ProductID   int64     `json:"product_id,string" form:"product_id"`
	VoucherCount int      `json:"voucher_count" form:"voucher_count"` // Number of vouchers in bundle
	Price       float64   `json:"price" form:"price"`                  // Bundle price
	Status      string    `gorm:"index" json:"status" form:"status"`                 // active, inactive

	IsDeleted   bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (VoucherBundle) TableName() string {
	return "voucher_bundle"
}

// VoucherBundleItem represents a voucher included in a bundle
type VoucherBundleItem struct {
	ID         int64     `json:"id,string" form:"id"`
	BundleID   int64     `json:"bundle_id,string" form:"bundle_id"`
	VoucherID  int64     `json:"voucher_id,string" form:"voucher_id"`
	VoucherCode string   `json:"voucher_code" form:"voucher_code"`
	Status     string    `json:"status" form:"status"` // available, sold, used
	CreatedAt  time.Time `json:"created_at"`
}

func (VoucherBundleItem) TableName() string {
	return "voucher_bundle_item"
}
