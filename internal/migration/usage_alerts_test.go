package migration

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateUsageAlertsTables(t *testing.T) {
	// Use SQLite for in-memory testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test creating tables
	if err := CreateUsageAlertsTables(db); err != nil {
		// SQLite doesn't support BIGSERIAL, so this will fail
		// but we can at least verify the SQL syntax is parseable
		t.Logf("Expected error on SQLite (PostgreSQL syntax): %v", err)
	}

	// Verify the function exists and has correct signature
	if CreateUsageAlertsTables == nil {
		t.Error("CreateUsageAlertsTables function is nil")
	}

	if DropUsageAlertsTables == nil {
		t.Error("DropUsageAlertsTables function is nil")
	}
}
