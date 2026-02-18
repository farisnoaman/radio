package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
)

func TestLocalBackupManager_SQLite(t *testing.T) {
	// Setup temporary directories
	tmpDir, err := os.MkdirTemp("", "toughradius_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	backupDir := filepath.Join(tmpDir, "backup")
	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(backupDir, 0755)

	// Create dummy sqlite db
	dbName := "test.db"
	dbPath := filepath.Join(dataDir, dbName)
	err = os.WriteFile(dbPath, []byte("dummy sqlite content"), 0644)
	assert.NoError(t, err)

	cfg := &config.AppConfig{
		System: config.SysConfig{
			Workdir: tmpDir,
		},
		Database: config.DBConfig{
			Type: "sqlite",
			Name: dbName,
		},
	}

	manager := NewLocalBackupManager(cfg)

	// Test CreateBackup
	filename, err := manager.CreateBackup()
	assert.NoError(t, err)
	assert.Contains(t, filename, "toughradius_")
	assert.Contains(t, filename, ".db")

	// Verify file exists in backup dir
	backupPath := filepath.Join(backupDir, filename)
	assert.FileExists(t, backupPath)

	// Test ListBackups
	backups, err := manager.ListBackups()
	assert.NoError(t, err)
	assert.Len(t, backups, 1)
	assert.Equal(t, filename, backups[0].FileName)
	assert.Equal(t, "sqlite", backups[0].Type)

	// Test PruneBackups
	// Create another dummy backup with older timestamp (manually)
	oldBackup := "toughradius_20200101120000.db"
	os.WriteFile(filepath.Join(backupDir, oldBackup), []byte("old content"), 0644)
	
	// Determine which file has older modification time (ListBackups uses modtime fallback if name parse fails, 
	// but here name parse should work)
	
	backups, _ = manager.ListBackups()
	assert.Len(t, backups, 2)

	// Keep 1, should delete the older one (2020...)
	err = manager.PruneBackups(1)
	assert.NoError(t, err)

	backups, _ = manager.ListBackups()
	assert.Len(t, backups, 1)
	assert.Equal(t, filename, backups[0].FileName) // The new one should remain

	// Test DeleteBackup
	err = manager.DeleteBackup(filename)
	assert.NoError(t, err)
	assert.NoFileExists(t, backupPath)
}
