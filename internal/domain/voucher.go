package domain

import (
	"time"
)

// VoucherBatch A batch of generated vouchers
type VoucherBatch struct {
	ID        int64     `json:"id,string" form:"id"`
	Name      string    `json:"name" form:"name"`
	ProductID int64     `json:"product_id,string" form:"product_id"`
	AgentID   int64     `json:"agent_id,string" form:"agent_id"` // Nullable if admin generated, but we use int64. 0 means system/admin.
	Count     int       `json:"count" form:"count"`
	Prefix    string    `json:"prefix" form:"prefix"`
	Remark    string     `json:"remark" form:"remark"`
	ExpireTime *time.Time `json:"expire_time"`
	IsDeleted  bool       `json:"is_deleted" gorm:"default:false"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (VoucherBatch) TableName() string {
	return "voucher_batch"
}

// Voucher Individual prepaid code
type Voucher struct {
	ID             int64     `json:"id,string" form:"id"`
	BatchID        int64     `json:"batch_id,string" form:"batch_id"`
	Code           string    `json:"code" gorm:"uniqueIndex" form:"code"` // The username/password
	RadiusUsername string    `json:"radius_username" form:"radius_username"` // Populated after activation
	Status         string    `json:"status" form:"status"` // unused, active, used, expired
	AgentID        int64     `json:"agent_id,string" form:"agent_id"`
	Price          float64   `json:"price" form:"price"`
	ActivatedAt    time.Time `json:"activated_at"`
	ExpireTime     time.Time `json:"expire_time"`
	IsDeleted      bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Voucher) TableName() string {
	return "voucher"
}
