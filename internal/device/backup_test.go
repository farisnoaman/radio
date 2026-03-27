package device

import (
	"context"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeviceBackup_BackupMikrotikConfig_ShouldSucceed(t *testing.T) {
	// Setup in-memory test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto-migrate the DeviceConfigBackup table
	if err := db.AutoMigrate(&domain.DeviceConfigBackup{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create backup service
	backup := NewDeviceBackupService(db)

	ctx := context.Background()
	device := &domain.NetNas{
		ID:         1,
		TenantID:   1,
		Ipaddr:     "192.168.1.1",
		VendorCode: "mikrotik",
		Name:       "Test Router",
	}

	record, err := backup.BackupConfig(ctx, device, "test-user")
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Wait a moment for async backup to complete
	time.Sleep(200 * time.Millisecond)

	// Refresh record from database
	if err := db.First(&domain.DeviceConfigBackup{}, record.ID).Error; err != nil {
		t.Fatalf("Failed to fetch record: %v", err)
	}

	if record.ID == 0 {
		t.Fatal("expected backup record ID to be set")
	}

	if record.Status != "completed" {
		t.Fatalf("expected status completed, got %s", record.Status)
	}

	if record.FileSize == 0 {
		t.Fatal("expected file size to be set")
	}
}
