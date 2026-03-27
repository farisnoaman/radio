package domain

import "time"

type BackupConfig struct {
	ID                int64     `json:"id" gorm:"primaryKey"`
	TenantID          int64     `json:"tenant_id" gorm:"uniqueIndex"`

	// Automated backup settings
	Enabled           bool      `json:"enabled"`
	Schedule          string    `json:"schedule"`           // daily, weekly
	RetentionDays     int       `json:"retention_days"`     // How long to keep
	MaxBackups        int       `json:"max_backups"`        // Max number of backups

	// Backup scope
	IncludeUserData    bool     `json:"include_user_data"`
	IncludeAccounting  bool     `json:"include_accounting"`
	IncludeVouchers    bool     `json:"include_vouchers"`
	IncludeNas         bool     `json:"include_nas"`

	// Storage settings
	StorageLocation    string    `json:"storage_location"`   // local, s3, gdrive
	StoragePath        string    `json:"storage_path"`
	EncryptionEnabled  bool      `json:"encryption_enabled"`
	EncryptionKey      string    `json:"encryption_key"`     // Encrypted at rest

	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (BackupConfig) TableName() string {
	return "mst_backup_config"
}

type BackupRecord struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	TenantID        int64      `json:"tenant_id" gorm:"index"`
	BackupType      string     `json:"backup_type"`       // automated, manual, admin
	Status          string     `json:"status"`            // pending, running, completed, failed

	// Backup details
	SchemaName      string     `json:"schema_name"`
	FilePath        string     `json:"file_path"`
	FileSize        int64      `json:"file_size"`         // in bytes
	Checksum        string     `json:"checksum"`

	// Statistics
	TablesCount     int        `json:"tables_count"`
	RowsCount       int64      `json:"rows_count"`
	Duration        int        `json:"duration"`          // seconds

	// Error tracking
	ErrorMessage    string     `json:"error_message"`

	// Timestamps
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (BackupRecord) TableName() string {
	return "mst_backup_record"
}
