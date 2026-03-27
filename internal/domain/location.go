package domain

import "time"

type Location struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	AgentID   *int64    `json:"agent_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:255"`
	Address   string    `json:"address" gorm:"size:500"`
	Region    string    `json:"region" gorm:"size:100"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timezone  string    `json:"timezone" gorm:"size:50"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Location) TableName() string {
	return "mst_location"
}
