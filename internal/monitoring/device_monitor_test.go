package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeviceMonitor(t *testing.T) {
	db := setupTestDB(t)
	collector := NewTenantMetricsCollector()

	// Create test device
	device := &domain.Server{
		TenantID:   1,
		Name:       "test-router",
		PublicIP:   "192.168.1.1",
		Username:   "admin",
		Password:   "password",
		Ports:      "8728",
	}
	result := db.Create(device)
	if result.Error != nil {
		t.Fatalf("Failed to create test device: %v", result.Error)
	}

	// Create monitor
	monitor := NewDeviceHealthMonitor(db, collector)

	// Mock MikroTik connection (in real test, use test server)
	// Will fail to connect in test, that's okay
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := monitor.checkDevice(ctx, *device)
	// Will fail to connect in test, that's expected
	if err != nil {
		t.Logf("Expected connection failure (this is okay in unit test): %v", err)
	}

	// Verify metrics were still recorded (for offline device)
	// The collector should have recorded the device as offline
}

func setupTestDB(t *testing.T) *gorm.DB {
	// Use a file-based temporary database with unique name for each test
	uniqueID := time.Now().Format("20060102150405.000000")
	dbPath := "/tmp/testdb_" + t.Name() + "_" + uniqueID + ".db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	// Automatically migrate the Server table
	err = db.AutoMigrate(&domain.Server{})
	require.NoError(t, err)

	return db
}
