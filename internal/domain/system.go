package domain

import (
	"time"
)

type SysConfig struct {
	ID        int64     `json:"id,string"   form:"id"`
	Sort      int       `json:"sort"  form:"sort"`
	Type      string    `gorm:"index" json:"type" form:"type"`
	Name      string    `gorm:"index" json:"name" form:"name"`
	Value     string    `json:"value" form:"value"`
	Remark    string    `json:"remark" form:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName Specify table name
func (SysConfig) TableName() string {
	return "sys_config"
}

type SysOpr struct {
	ID        int64     `json:"id,string" form:"id"`
	TenantID  int64     `gorm:"index" json:"tenant_id" form:"tenant_id"` // Tenant/Provider ID (0 = platform-wide)
	Realname  string    `json:"realname" form:"realname"`
	Mobile    string    `json:"mobile" form:"mobile"`
	Email     string    `json:"email" form:"email"`
	Username  string    `gorm:"index" json:"username" form:"username"`
	Password  string    `json:"password" form:"password"`
	Level     string    `gorm:"index" json:"level" form:"level"`
	Status    string    `gorm:"index" json:"status" form:"status"`
	RadiusUsername string `json:"radius_username" form:"radius_username" gorm:"index;size:255"`

	Remark    string    `json:"remark" form:"remark"`
	LastLogin time.Time `json:"last_login" form:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName Specify table name
func (SysOpr) TableName() string {
	return "sys_opr"
}

type SysOprLog struct {
	ID        int64     `json:"id,string"`
	OprName   string    `gorm:"index" json:"opr_name"`
	OprIp     string    `json:"opr_ip"`
	OptAction string    `gorm:"index" json:"opt_action"`
	OptDesc   string    `json:"opt_desc"`
	OptTime   time.Time `gorm:"index" json:"opt_time"`

}

// TableName Specify table name
func (SysOprLog) TableName() string {
	return "sys_opr_log"
}
