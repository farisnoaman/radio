package backup

import (
	"context"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBackupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate tables
	err = db.AutoMigrate(
		&domain.BackupConfig{},
		&domain.BackupRecord{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	return db
}

func TestCreateBackup(t *testing.T) {
	db := setupBackupTestDB(t)
	service := NewBackupService(db, nil)

	// Create backup config
	config := &domain.BackupConfig{
		TenantID:   1,
		Enabled:    true,
		MaxBackups: 5,
	}
	db.Create(config)

	// Create backup
	record, err := service.CreateBackup(context.Background(), 1, "manual")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if record.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", record.Status)
	}

	if record.TenantID != 1 {
		t.Errorf("Expected tenant ID 1, got %d", record.TenantID)
	}
}

func TestQuotaExceeded(t *testing.T) {
	db := setupBackupTestDB(t)
	service := NewBackupService(db, nil)

	// Create backup config with max 1 backup (use tenant ID 2 to avoid conflict)
	config := &domain.BackupConfig{
		TenantID:   2,
		Enabled:    true,
		MaxBackups: 1,
	}
	db.Create(config)

	// Create first backup
	db.Create(&domain.BackupRecord{
		TenantID: 2,
		Status:   "completed",
	})

	// Try to create second backup (should fail)
	_, err := service.CreateBackup(context.Background(), 2, "manual")
	if err != ErrBackupQuotaExceeded {
		t.Errorf("Expected ErrBackupQuotaExceeded, got %v", err)
	}
}
