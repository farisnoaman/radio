package domain

import (
	"testing"
	"time"
)

func TestBackupConfigModel(t *testing.T) {
	config := &BackupConfig{
		TenantID:          1,
		Enabled:           true,
		Schedule:          "daily",
		RetentionDays:     30,
		MaxBackups:        10,
		IncludeUserData:   true,
		IncludeAccounting: true,
		IncludeVouchers:   true,
		IncludeNas:        true,
		StorageLocation:   "local",
		EncryptionEnabled: true,
	}

	if config.TableName() != "mst_backup_config" {
		t.Errorf("Expected table name 'mst_backup_config', got '%s'", config.TableName())
	}
}

func TestBackupRecordModel(t *testing.T) {
	record := &BackupRecord{
		TenantID:   1,
		BackupType: "automated",
		Status:     "pending",
		SchemaName: "provider_1",
		StartedAt:  time.Now(),
	}

	if record.TableName() != "mst_backup_record" {
		t.Errorf("Expected table name 'mst_backup_record', got '%s'", record.TableName())
	}
}
