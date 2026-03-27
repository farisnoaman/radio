package domain

import "time"

// Server represents a remote server or Mikrotik router managed by the system.
// It maps to the "net_server" table in the database and stores API credentials,
// database connection info, and monitoring statistics.
type Server struct {
	ID            int64     `json:"id,string" form:"id"`
	TenantID      int64     `gorm:"index" json:"tenant_id" form:"tenant_id"` // Tenant/Provider ID
	Name          string    `json:"name" form:"name" gorm:"size:255;not null"`
	PublicIP      string    `json:"public_ip" form:"public_ip" gorm:"size:64"`
	Secret        string    `json:"secret" form:"secret" gorm:"size:128"`
	Username      string    `json:"username" form:"username" gorm:"size:128"`
	Password      string    `json:"password" form:"password" gorm:"size:128"`
	Ports         string    `json:"ports" form:"ports" gorm:"size:128"`
	RouterLimit   string    `json:"router_limit" form:"router_limit" gorm:"size:255"`

	// Database Information fields
	DBHost        string    `json:"db_host" form:"db_host" gorm:"size:128"`
	DBPort        int       `json:"db_port" form:"db_port"`
	DBName        string    `json:"db_name" form:"db_name" gorm:"size:128"`
	DBUsername    string    `json:"db_username" form:"db_username" gorm:"size:128"`
	DBPassword    string    `json:"db_password" form:"db_password" gorm:"size:128"`

	// Monitoring / Status fields
	RouterStatus  string    `json:"router_status" form:"router_status" gorm:"size:32"`
	OnlineHotspot int       `json:"online_hotspot" form:"online_hotspot"`
	OnlinePPPoE   int       `json:"online_pppoe" form:"online_pppoe"`

	CreatedAt     time.Time `json:"created_at" form:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" form:"updated_at"`
}

// TableName returns the table name for the Server model
func (Server) TableName() string {
	return "net_server"
}
