package repository

import (
	"context"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate tables
	db.AutoMigrate(&domain.RadiusUser{})

	return db
}

func TestTenantScope(t *testing.T) {
	db := setupTestDB(t)

	// Create test users for different tenants
	users := []domain.RadiusUser{
		{Username: "user1", TenantID: 1, Status: "enabled"},
		{Username: "user2", TenantID: 1, Status: "enabled"},
		{Username: "user3", TenantID: 2, Status: "enabled"},
	}
	db.Create(&users)

	// Create context with tenant ID
	ctx := tenant.WithTenantID(context.Background(), 1)

	// Query with tenant scope
	var results []domain.RadiusUser
	err := db.WithContext(ctx).Scopes(TenantScope).Find(&results).Error
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should only return users from tenant 1
	if len(results) != 2 {
		t.Errorf("Expected 2 users, got %d", len(results))
	}

	for _, user := range results {
		if user.TenantID != 1 {
			t.Errorf("Expected tenant ID 1, got %d", user.TenantID)
		}
	}
}

func TestTenantScopeWithAdmin(t *testing.T) {
	db := setupTestDB(t)

	// Create test users
	users := []domain.RadiusUser{
		{Username: "user1", TenantID: 1, Status: "enabled"},
		{Username: "user2", TenantID: 2, Status: "enabled"},
	}
	db.Create(&users)

	// Create context with platform admin (no tenant filter)
	ctx := context.Background()

	// Admin query bypasses tenant scope
	var results []domain.RadiusUser
	err := db.WithContext(ctx).Scopes(AdminTenantScope(1)).Find(&results).Error
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should return users from tenant 1 (admin specified)
	if len(results) != 1 {
		t.Errorf("Expected 1 user, got %d", len(results))
	}

	if results[0].TenantID != 1 {
		t.Errorf("Expected tenant ID 1, got %d", results[0].TenantID)
	}
}

func TestTenantScopeNoContext(t *testing.T) {
	db := setupTestDB(t)

	// Create test users
	users := []domain.RadiusUser{
		{Username: "user1", TenantID: 1, Status: "enabled"},
	}
	db.Create(&users)

	// Query without tenant context
	var results []domain.RadiusUser
	err := db.Scopes(TenantScope).Find(&results).Error
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should return empty result (no tenant context)
	if len(results) != 0 {
		t.Errorf("Expected 0 users without tenant context, got %d", len(results))
	}
}

func TestAllTenantsScope(t *testing.T) {
	db := setupTestDB(t)

	// Create test users
	users := []domain.RadiusUser{
		{Username: "user1", TenantID: 1, Status: "enabled"},
		{Username: "user2", TenantID: 2, Status: "enabled"},
	}
	db.Create(&users)

	// Query bypassing tenant filter
	var results []domain.RadiusUser
	err := db.Scopes(AllTenantsScope).Find(&results).Error
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should return all users
	if len(results) != 2 {
		t.Errorf("Expected 2 users, got %d", len(results))
	}
}

func TestTenantScopeWithID(t *testing.T) {
	db := setupTestDB(t)

	// Create test users
	users := []domain.RadiusUser{
		{Username: "user1", TenantID: 1, Status: "enabled"},
		{Username: "user2", TenantID: 2, Status: "enabled"},
	}
	db.Create(&users)

	// Query with specific tenant ID
	var results []domain.RadiusUser
	err := db.Scopes(TenantScopeWithID(2)).Find(&results).Error
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should return users from tenant 2
	if len(results) != 1 {
		t.Errorf("Expected 1 user, got %d", len(results))
	}

	if results[0].TenantID != 2 {
		t.Errorf("Expected tenant ID 2, got %d", results[0].TenantID)
	}
}

func TestWithTenant(t *testing.T) {
	db := setupTestDB(t)

	// Create test users
	users := []domain.RadiusUser{
		{Username: "user1", TenantID: 1, Status: "enabled"},
		{Username: "user2", TenantID: 2, Status: "enabled"},
	}
	db.Create(&users)

	// Query using WithTenant convenience function
	var results []domain.RadiusUser
	err := WithTenant(db, 1).Find(&results).Error
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should return users from tenant 1
	if len(results) != 1 {
		t.Errorf("Expected 1 user, got %d", len(results))
	}

	if results[0].TenantID != 1 {
		t.Errorf("Expected tenant ID 1, got %d", results[0].TenantID)
	}
}
