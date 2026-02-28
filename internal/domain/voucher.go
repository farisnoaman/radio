package domain

import (
	"time"
)

// VoucherBatch represents a batch of generated vouchers that can be sold
// to customers. Each batch is linked to a Product which defines the actual
// allocations (data quota, time validity) that vouchers will inherit.
//
// The batch controls printing/validity deadlines - vouchers cannot be printed
// or activated after PrintExpireTime. The actual allocations come from the
// Product at voucher creation time.
//
// Database table: voucher_batch
//
// Lifecycle:
//   - Created via Admin API POST /api/v1/voucher-batches
//   - Generates multiple Voucher records on creation
//   - Can be activated/deactivated in bulk
//   - Supports soft delete with restore capability
type VoucherBatch struct {
	ID           int64      `json:"id,string" form:"id"`
	Name         string     `json:"name" form:"name"`
	ProductID    int64      `gorm:"index" json:"product_id,string" form:"product_id"`
	AgentID      int64      `gorm:"index" json:"agent_id,string" form:"agent_id"`
	Count        int        `json:"count" form:"count"`
	Prefix       string     `json:"prefix" form:"prefix"`
	Remark      string     `json:"remark" form:"remark"`
	PrintExpireTime  *time.Time `json:"print_expire_time" gorm:"column:expire_time"`  // Deadline for printing/activating vouchers (maps to existing expire_time column)
	GeneratePIN bool        `json:"generate_pin" form:"generate_pin"`
	PINLength    int        `json:"pin_length" form:"pin_length"`
	ExpirationType string    `json:"expiration_type" form:"expiration_type"` // "fixed" or "first_use"
	ValidityDays   int       `json:"validity_days" form:"validity_days"`
	IsDeleted      bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`
}

func (VoucherBatch) TableName() string {
	return "voucher_batch"
}

// Voucher represents an individual prepaid access code that customers can
// redeem to get internet access. Each voucher is linked to a Product which
// defines the allocations (data quota, time validity) that this voucher provides.
//
// Vouchers track both their allocated quota and actual usage. Expiration can
// occur via either: (1) time-based ExpireTime reached, or (2) quota-based
// when DataUsed >= DataQuota or TimeUsed >= TimeQuota.
//
// Database table: voucher
//
// Lifecycle:
//   - Created when batch is generated (inherits allocations from Product)
//   - Status transitions: unused → active → used/expired
//   - Quota usage tracked via RADIUS accounting
//   - Soft-deleted after grace period when expired
type Voucher struct {
	ID              int64     `json:"id,string" form:"id"`
	BatchID         int64     `gorm:"index" json:"batch_id,string" form:"batch_id"`
	Code            string    `json:"code" gorm:"uniqueIndex" form:"code"`
	RadiusUsername  string    `gorm:"index" json:"radius_username" form:"radius_username"`
	Status          string    `gorm:"index" json:"status" form:"status"` // unused, active, used, expired
	AgentID         int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
	Price           float64   `json:"price" form:"price"`
	ActivatedAt     time.Time `json:"activated_at"`
	ExpireTime      time.Time `gorm:"index" json:"expire_time"`

	// Allocations from Product (set at voucher creation)
	DataQuota int64 `json:"data_quota" form:"data_quota"` // MB (0 = unlimited)
	TimeQuota int64 `json:"time_quota" form:"time_quota"` // seconds (0 = unlimited)

	// Actual usage tracking
	DataUsed int64 `json:"data_used" form:"data_used"` // MB used
	TimeUsed int64 `json:"time_used" form:"time_used"` // seconds used

	// Grace period tracking for quota-based expiration
	QuotaExpiredAt *time.Time `json:"quota_expired_at"`

	ExtendedCount   int       `json:"extended_count"`
	LastExtendedAt  time.Time `json:"last_extended_at"`
	FirstUsedAt     time.Time `json:"first_used_at"`
	PIN             string    `json:"-"`                  // Hashed PIN (never exposed via JSON)
	RequirePIN      bool      `json:"require_pin" form:"require_pin"`
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
