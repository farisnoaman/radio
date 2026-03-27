package device

import (
	"context"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSpeedTest_RunMikrotikSpeedTest_ShouldReturnResults(t *testing.T) {
	// Setup in-memory test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto-migrate the SpeedTestResult table
	if err := db.AutoMigrate(&domain.SpeedTestResult{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	service := NewSpeedTestService(db)

	ctx := context.Background()
	device := &domain.NetNas{
		ID:         1,
		TenantID:   1,
		Ipaddr:     "192.168.1.1",
		VendorCode: "mikrotik",
	}

	result, err := service.RunSpeedTest(ctx, device, "test-user")
	if err != nil {
		t.Fatalf("speed test failed: %v", err)
	}

	// Wait for async test to complete
	time.Sleep(200 * time.Millisecond)

	// Refresh from database
	if err := db.First(&domain.SpeedTestResult{}, result.ID).Error; err != nil {
		t.Fatalf("Failed to fetch result: %v", err)
	}

	if result.UploadMbps == 0 {
		t.Error("expected upload speed")
	}

	if result.DownloadMbps == 0 {
		t.Error("expected download speed")
	}
}
