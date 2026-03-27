package domain

import (
	"time"
)

type DailySnapshot struct {
	ID                    int64     `json:"id" gorm:"primaryKey"`
	ProviderID            int64     `json:"provider_id" gorm:"index"`
	SnapshotDate          time.Time `json:"snapshot_date" gorm:"type:date"`
	TotalUsers            int       `json:"total_users"`
	ActiveUsers           int       `json:"active_users"`
	NewMonthlyUsers       int       `json:"new_monthly_users"`
	NewVoucherUsers       int       `json:"new_voucher_users"`
	TotalSessions         int       `json:"total_sessions"`
	ActiveSessions        int       `json:"active_sessions"`
	MonthlyDataUsedBytes  int64     `json:"monthly_data_used_bytes"`
	VoucherDataUsedBytes  int64     `json:"voucher_data_used_bytes"`
	ActiveNodes           int       `json:"active_nodes"`
	TotalNodes            int       `json:"total_nodes"`
	ActiveServers         int       `json:"active_servers"`
	TotalServers          int       `json:"total_servers"`
	ActiveCPEs            int       `json:"active_cpes"`
	TotalCPEs             int       `json:"total_cpes"`
	TotalAgents           int       `json:"total_agents"`
	TotalBatches          int       `json:"total_batches"`
	AgentRevenue          float64   `json:"agent_revenue"`
	MRR                   float64   `json:"mrr"`
	DeviceIssues          int       `json:"device_issues"`
	NetworkIssues         int       `json:"network_issues"`
	FraudAttempts         int       `json:"fraud_attempts"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (DailySnapshot) TableName() string {
	return "reporting_daily_snapshots"
}

type FraudLog struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	ProviderID int64     `json:"provider_id" gorm:"index"`
	VoucherID  int64     `json:"voucher_id"`
	UserID     int64     `json:"user_id"`
	IPAddress  string    `json:"ip_address" gorm:"type:varchar(45)"`
	EventType  string    `json:"event_type" gorm:"type:varchar(50)"`
	Details    string    `json:"details" gorm:"type:jsonb"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (FraudLog) TableName() string {
	return "reporting_fraud_log"
}

type ProviderNotificationPreference struct {
	ID                        int64     `json:"id" gorm:"primaryKey"`
	ProviderID                int64     `json:"provider_id" gorm:"uniqueIndex"`
	AlertPercentages          string    `json:"alert_percentages" gorm:"type:varchar(50);default:'70,85,100'"`
	AlertPercentagesEnabled   bool      `json:"alert_percentages_enabled" gorm:"default:true"`
	MaxUsersThreshold        int       `json:"max_users_threshold"`
	MaxDataBytesThreshold    int64     `json:"max_data_bytes_threshold"`
	AbsoluteAlertsEnabled    bool      `json:"absolute_alerts_enabled" gorm:"default:false"`
	AnomalyDetectionEnabled  bool      `json:"anomaly_detection_enabled" gorm:"default:false"`
	AnomalyThresholdPercent  int       `json:"anomaly_threshold_percent" gorm:"default:50"`
	EmailEnabled             bool      `json:"email_enabled" gorm:"default:true"`
	SMSEnabled               bool      `json:"sms_enabled" gorm:"default:false"`
	CreatedAt                time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ProviderNotificationPreference) TableName() string {
	return "provider_notification_preferences"
}

type NetworkIssue struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	ProviderID  int64      `json:"provider_id" gorm:"index"`
	DeviceType  string     `json:"device_type" gorm:"type:varchar(20)"`
	DeviceID    int64      `json:"device_id"`
	DeviceName  string     `json:"device_name" gorm:"type:varchar(255)"`
	IssueType   string     `json:"issue_type" gorm:"type:varchar(50)"`
	IssueDetails string     `json:"issue_details" gorm:"type:text"`
	Status      string     `json:"status" gorm:"type:varchar(20);default:'open'"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (NetworkIssue) TableName() string {
	return "reporting_network_issues"
}
