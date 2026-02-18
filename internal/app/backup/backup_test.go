package backup

import (
	"testing"


	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
)

// MockInterface for testing (if we abstract GoogleDriveProvider later)
// For now, testing config loading and initialization logic without real API calls

func TestNewLocalBackupManager_WithGDrive(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Workdir: "/tmp/tr_test_backup",
		},
		Database: config.DBConfig{
			Type: "sqlite",
			Name: "test.db",

		},
		Backup: config.BackupConfig{
			Enabled: true,
			GoogleDrive: config.GoogleDriveConfig{
				Enabled:            true,
				ServiceAccountJSON: "{\"type\": \"service_account\"}", // Minimal valid-ish JSON
				FolderID:           "test-folder-id",
			},
		},
	}

	// We expect NewLocalBackupManager to initialize without panic
	// Note: NewGoogleDriveProvider might fail with invalid JSON credential, so we handle that case.
	// Since we can't easily mock the google auth lib here without cleaner interfaces, 
	// we accept that mgr.gdrive might be nil if validation fails, but the function should not crash.
	
	manager := NewLocalBackupManager(cfg)
	assert.NotNil(t, manager)
	assert.Equal(t, "/tmp/tr_test_backup/backup", manager.backupDir)
}

func TestLocalBackupManager_ConfigDisabled(t *testing.T) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Workdir: "/tmp/tr_test_backup",
		},

		Backup: config.BackupConfig{
			Enabled: true,
			GoogleDrive: config.GoogleDriveConfig{
				Enabled: false,
			},
		},
	}

	manager := NewLocalBackupManager(cfg)
	assert.Nil(t, manager.gdrive)
}
