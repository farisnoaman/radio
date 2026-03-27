package domain

import (
	"time"
)

// NetworkDevice represents a managed network device (router, AP, switch, etc.)
type NetworkDevice struct {
	ID         int64  `json:"id" gorm:"primaryKey"`
	TenantID   int64  `json:"tenant_id" gorm:"index"`
	LocationID *int64 `json:"location_id" gorm:"index"`

	// Device Identity
	Name            string `json:"name" gorm:"size:255"`
	DeviceType      string `json:"device_type" gorm:"size:50;index"`
	Vendor          string `json:"vendor" gorm:"size:100;index"`
	Model           string `json:"model" gorm:"size:100"`
	SerialNumber    string `json:"serial_number" gorm:"size:100"`
	FirmwareVersion string `json:"firmware_version" gorm:"size:100"`

	// Network
	IPAddress     string `json:"ip_address" gorm:"size:45"`
	MacAddress    string `json:"mac_address" gorm:"size:17"`
	SNMPPort      int    `json:"snmp_port" gorm:"default:161"`
	SNMPCommunity string `json:"-" gorm:"size:100"`

	// API Access (for MikroTik)
	APIEndpoint string `json:"api_endpoint" gorm:"size:500"`
	APIUsername string `json:"api_username" gorm:"size:100"`
	APIPassword string `json:"-" gorm:"size:255"`

	// Status
	Status      string     `json:"status" gorm:"size:20;index;default:unknown"`
	LastSeen    *time.Time `json:"last_seen"`
	LastOnline  *time.Time `json:"last_online"`
	LastOffline *time.Time `json:"last_offline"`

	// Settings
	PollingEnabled  bool `json:"polling_enabled" gorm:"default:true"`
	PollingInterval int  `json:"polling_interval" gorm:"default:60"`
	AlertOnOffline  bool `json:"alert_on_offline" gorm:"default:true"`

	// Metadata
	Tags   string `json:"tags" gorm:"size:500"`
	Remark string `json:"remark" gorm:"size:1000"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the database table name for network devices.
func (NetworkDevice) TableName() string {
	return "network_device"
}

// NetworkDeviceMetric represents a metric reading from a network device.
type NetworkDeviceMetric struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	DeviceID    int64     `json:"device_id" gorm:"index"`
	MetricType  string    `json:"metric_type" gorm:"size:50;index"`
	Value       float64   `json:"value" gorm:"type:decimal(15,4)"`
	Unit        string    `json:"unit" gorm:"size:20"`
	Severity    string    `json:"severity" gorm:"size:20;default:normal"`
	CollectedAt time.Time `json:"collected_at" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName specifies the database table name for device metrics.
func (NetworkDeviceMetric) TableName() string {
	return "network_device_metric"
}

// NetworkDeviceAlert represents an alert generated from a network device.
type NetworkDeviceAlert struct {
	ID             int64      `json:"id" gorm:"primaryKey"`
	DeviceID       int64      `json:"device_id" gorm:"index"`
	TenantID       int64      `json:"tenant_id" gorm:"index"`
	AlertType      string     `json:"alert_type" gorm:"size:50"`
	Severity       string     `json:"severity" gorm:"size:20"`
	Message        string     `json:"message" gorm:"type:text"`
	MetricType     string     `json:"metric_type" gorm:"size:50"`
	MetricValue    *float64   `json:"metric_value" gorm:"type:decimal(15,4)"`
	ThresholdValue *float64   `json:"threshold_value" gorm:"type:decimal(15,4)"`
	Status         string     `json:"status" gorm:"size:20;index;default:active"`
	AcknowledgedBy *int64     `json:"acknowledged_by"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// TableName specifies the database table name for device alerts.
func (NetworkDeviceAlert) TableName() string {
	return "network_device_alert"
}

// Device type constants
const (
	DeviceTypeRouter   = "router"
	DeviceTypeAP       = "ap"
	DeviceTypeBridge   = "bridge"
	DeviceTypeSwitch   = "switch"
	DeviceTypeFirewall = "firewall"
	DeviceTypeModem    = "modem"
	DeviceTypeOther    = "other"
)

// Device status constants
const (
	DeviceStatusOnline  = "online"
	DeviceStatusOffline = "offline"
	DeviceStatusUnknown = "unknown"
)

// Alert type constants
const (
	AlertTypeOffline   = "offline"
	AlertTypeOnline    = "online"
	AlertTypeThreshold = "threshold"
	AlertTypeError     = "error"
)

// Severity constants (aliases for environment metrics compatibility)
const (
	DevSeverityInfo     = "info"
	DevSeverityWarning  = "warning"
	DevSeverityCritical = "critical"
	DevSeverityNormal   = "normal"
)

// Alert status constants
const (
	DevAlertStatusActive       = "active"
	DevAlertStatusAcknowledged = "acknowledged"
	DevAlertStatusResolved     = "resolved"
)

// Metric type constants
const (
	MetricTypeCPU        = "cpu_load"
	MetricTypeMemory     = "memory"
	MetricTypeDeviceTemp = "temperature"
	MetricTypeDeviceVolt = "voltage"
	MetricTypeSignal     = "signal"
	MetricTypeUptime     = "uptime"
	MetricTypeDownload   = "download"
	MetricTypeUpload     = "upload"
	MetricTypeLatency    = "latency"
)

// Vendor constants
const (
	VendorMikroTik = "mikrotik"
	VendorUbiquiti = "ubiquiti"
	VendorTPLink   = "tplink"
	VendorCisco    = "cisco"
	VendorHuawei   = "huawei"
	VendorOther    = "other"
)
