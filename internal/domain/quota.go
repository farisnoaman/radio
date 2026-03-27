package domain

import "time"

type ProviderQuota struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"uniqueIndex"`

	// User Limits
	MaxUsers         int       `json:"max_users" gorm:"default:1000"`
	MaxOnlineUsers   int       `json:"max_online_users" gorm:"default:500"`

	// Device Limits
	MaxNAS           int       `json:"max_nas" gorm:"default:100"`
	MaxMikrotikDevices int     `json:"max_mikrotik_devices" gorm:"default:50"`

	// Storage Limits (GB)
	MaxStorage       int64     `json:"max_storage" gorm:"default:100"`
	MaxDailyBackups  int       `json:"max_daily_backups" gorm:"default:3"`

	// Bandwidth Limits (Gbps)
	MaxBandwidth     float64   `json:"max_bandwidth" gorm:"default:10"`

	// RADIUS Limits (requests per second)
	MaxAuthPerSecond int       `json:"max_auth_per_second" gorm:"default:100"`
	MaxAcctPerSecond int       `json:"max_acct_per_second" gorm:"default:200"`

	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (ProviderQuota) TableName() string {
	return "mst_provider_quota"
}

type ProviderUsage struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"index"`

	// Current Usage
	CurrentUsers     int       `json:"current_users"`
	CurrentOnlineUsers int     `json:"current_online_users"`
	CurrentNAS       int       `json:"current_nas"`
	CurrentStorageGB float64   `json:"current_storage_gb"`
	CurrentBandwidth float64   `json:"current_bandwidth"`

	// Period Totals
	TotalAuthRequests int64    `json:"total_auth_requests"`
	TotalAcctRequests int64    `json:"total_acct_requests"`

	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ProviderUsage) TableName() string {
	return "mst_provider_usage"
}
