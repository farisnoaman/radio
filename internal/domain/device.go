package domain

import "time"

// DeviceConfigBackup represents a device configuration backup record.
type DeviceConfigBackup struct {
	ID          int64      `json:"id,string" gorm:"primaryKey"`
	TenantID    int64      `json:"tenant_id" gorm:"index"`
	NasID       int64      `json:"nas_id" gorm:"index"`
	VendorCode  string     `json:"vendor_code" gorm:"index"`
	ConfigData  string     `json:"config_data" gorm:"type:text"` // Encrypted config
	FileSize    int64      `json:"file_size"`
	Status      string     `json:"status" gorm:"default:pending"` // pending, running, completed, failed
	Error       string     `json:"error,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName specifies the table name for DeviceConfigBackup.
func (DeviceConfigBackup) TableName() string {
	return "device_config_backup"
}

// SpeedTestResult represents a network speed test result.
type SpeedTestResult struct {
	ID              int64     `json:"id,string" gorm:"primaryKey"`
	TenantID        int64     `json:"tenant_id" gorm:"index"`
	NasID           int64     `json:"nas_id" gorm:"index"`
	TestServer      string    `json:"test_server"`
	UploadMbps      float64   `json:"upload_mbps"`
	DownloadMbps    float64   `json:"download_mbps"`
	LatencyMs       float64   `json:"latency_ms"`
	JitterMs        float64   `json:"jitter_ms"`
	PacketLoss      float64   `json:"packet_loss_percent"`
	TestDurationSec int       `json:"test_duration_sec"`
	Status          string    `json:"status"` // running, completed, failed
	Error           string    `json:"error,omitempty"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}

// TableName specifies the table name.
func (SpeedTestResult) TableName() string {
	return "speed_test_result"
}
