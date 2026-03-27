package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&domain.LoyaltyProfile{},
		&domain.LoyaltyIdentity{},
		&domain.LoyaltyRule{},
	)
	require.NoError(t, err)

	return db
}

func TestLoyaltyService_IdentityGenerator(t *testing.T) {
	s := NewLoyaltyService(nil)
	key1 := s.GenerateIdentityKey("00:11:22:33:44:55", 1)
	key2 := s.GenerateIdentityKey("00:11:22:33:44:55", 1)
	key3 := s.GenerateIdentityKey("00:11:22:33:44:55", 2)
	
	assert.Equal(t, key1, key2, "Same MAC and TenantID should generate the same hash")
	assert.NotEqual(t, key1, key3, "Different TenantID should generate a different hash")
}

func TestLoyaltyService_ProcessUsageEvent(t *testing.T) {
	db := setupTestDB(t)
	s := NewLoyaltyService(db)
	ctx := context.Background()
	tenantID := int64(1)
	mac := "AA:BB:CC:DD:EE:FF"

	// Create a test rule: 1 GB data OR 1 hour time -> 10 points
	err := db.Create(&domain.LoyaltyRule{
		TenantID:      tenantID,
		Name:          "Test Rule",
		DataThreshold: 1024 * 1024 * 1024, // 1 GB
		TimeThreshold: 3600,               // 1 Hour
		RequireBoth:   false,
		PointsAwarded: 10,
	}).Error
	require.NoError(t, err)

	// --- Step 1: Initial event that doesn't trigger the rule ---
	err = s.ProcessUsageEvent(ctx, UsageEvent{
		TenantID: tenantID,
		Mac:      mac,
		DataUsed: 500 * 1024 * 1024, // 500 MB
		TimeUsed: 1800,              // 30 min
	})
	require.NoError(t, err)

	var profile domain.LoyaltyProfile
	db.First(&profile)
	assert.Equal(t, int64(0), profile.Points, "Not enough usage to get points")
	assert.Equal(t, "None", profile.Badge)
	assert.Equal(t, int64(500*1024*1024), profile.TotalDataUsed)
	assert.Equal(t, int64(1800), profile.TotalTimeUsed)
	assert.Equal(t, int64(500*1024*1024), profile.MilestoneDataUsed)

	// --- Step 2: Triggering the threshold ---
	// Add 600 MB more data -> 1.1 GB total (over the 1 GB threshold)
	err = s.ProcessUsageEvent(ctx, UsageEvent{
		TenantID: tenantID,
		Mac:      mac,
		DataUsed: 600 * 1024 * 1024,
		TimeUsed: 0,
	})
	require.NoError(t, err)

	db.First(&profile, profile.ID)
	assert.Equal(t, int64(10), profile.Points, "Should award 10 points for crossing 1GB")
	// Overflow after subtracting 1GB: (500 + 600)MB - 1GB = 100MB
	assert.Equal(t, int64(100*1024*1024), profile.MilestoneDataUsed, "Milestone tracker should preserve 100MB overflow")
	assert.Equal(t, int64(1100*1024*1024), profile.TotalDataUsed, "Total data should be 1.1GB")

	// --- Step 3: Triggering Bronze Badge ---
	// 200GB and 50 Hours total needed for Bronze Badge
	err = s.ProcessUsageEvent(ctx, UsageEvent{
		TenantID: tenantID,
		Mac:      mac,
		DataUsed: 200 * 1024 * 1024 * 1024, // + 200 GB
		TimeUsed: 50 * 3600,                // + 50 Hours
	})
	require.NoError(t, err)

	db.First(&profile, profile.ID)
	assert.Equal(t, "Bronze", profile.Badge, "User should reach Bronze Badge status")

	// --- Step 4: Triggering Silver Badge ---
	// 500GB and 70 Hours total for Silver. (We already have 200GB+ and 50H+, so adding 300GB+ and 20H+)
	err = s.ProcessUsageEvent(ctx, UsageEvent{
		TenantID: tenantID,
		Mac:      mac,
		DataUsed: 300 * 1024 * 1024 * 1024, // + 300 GB
		TimeUsed: 20 * 3600,                // + 20 Hours
	})
	require.NoError(t, err)

	db.First(&profile, profile.ID)
	assert.Equal(t, "Silver", profile.Badge, "User should reach Silver Badge status")
}
