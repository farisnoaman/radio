package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate the RadiusAccounting table
	if err := db.AutoMigrate(&domain.RadiusAccounting{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestGormAccountingRepository_GetTotalSessionTime(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	repo := NewGormAccountingRepository(db)
	ctx := context.Background()

	// Create test accounting records
	accounting1 := &domain.RadiusAccounting{
		Username:        "testuser",
		AcctSessionTime: 3600, // 1 hour
		AcctInputTotal:  1000000,
		AcctOutputTotal: 2000000,
	}
	accounting2 := &domain.RadiusAccounting{
		Username:        "testuser",
		AcctSessionTime: 1800, // 30 minutes
		AcctInputTotal:  500000,
		AcctOutputTotal: 1000000,
	}
	err := repo.Create(ctx, accounting1)
	assert.Nil(t, err)
	err = repo.Create(ctx, accounting2)
	assert.Nil(t, err)

	// Test
	totalTime, err := repo.GetTotalSessionTime(ctx, "testuser")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, int64(5400), totalTime) // 3600 + 1800 = 5400 seconds (1.5 hours)
}

func TestGormAccountingRepository_GetTotalSessionTime_NoRecords(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	repo := NewGormAccountingRepository(db)
	ctx := context.Background()

	// Test with a user that has no records
	totalTime, err := repo.GetTotalSessionTime(ctx, "nonexistent")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, int64(0), totalTime)
}
