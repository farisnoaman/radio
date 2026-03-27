package repository

import (
	"context"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProxyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&domain.RadiusProxyServer{}, &domain.RadiusProxyRealm{}, &domain.ProxyRequestLog{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestProxyRepository_CreateServer_ShouldSucceed(t *testing.T) {
	db := setupProxyTestDB(t)
	repo := NewProxyRepository(db)

	// Create context with tenant
	ctx := tenant.WithTenantID(context.Background(), 1)

	server := &domain.RadiusProxyServer{
		Name:      "Test Proxy",
		Host:      "192.168.1.10",
		AuthPort:  1812,
		AcctPort:  1813,
		Secret:    "testsecret",
	}

	err := repo.CreateServer(ctx, server)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if server.ID == 0 {
		t.Fatal("expected ID to be set")
	}
	if server.TenantID != 1 {
		t.Errorf("expected tenant ID 1, got %d", server.TenantID)
	}
}

func TestProxyRepository_ListServers_ShouldSucceed(t *testing.T) {
	db := setupProxyTestDB(t)
	repo := NewProxyRepository(db)

	ctx := tenant.WithTenantID(context.Background(), 1)

	// Create test servers
	servers := []*domain.RadiusProxyServer{
		{Name: "Proxy 1", Host: "192.168.1.10", AuthPort: 1812, AcctPort: 1813, Secret: "secret1", Status: "enabled", Priority: 2},
		{Name: "Proxy 2", Host: "192.168.1.11", AuthPort: 1812, AcctPort: 1813, Secret: "secret2", Status: "enabled", Priority: 1},
	}

	for _, s := range servers {
		if err := repo.CreateServer(ctx, s); err != nil {
			t.Fatalf("failed to create server: %v", err)
		}
	}

	// List servers
	result, err := repo.ListServers(ctx)
	if err != nil {
		t.Fatalf("failed to list servers: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 servers, got %d", len(result))
	}

	// Should be ordered by priority
	if result[0].Name != "Proxy 2" {
		t.Errorf("expected first server to be Proxy 2 (priority 1), got %s", result[0].Name)
	}
}

func TestProxyRepository_FindRealmForUsername_ShouldMatch(t *testing.T) {
	db := setupProxyTestDB(t)
	repo := NewProxyRepository(db)

	ctx := tenant.WithTenantID(context.Background(), 1)

	// Create test realm
	realm := &domain.RadiusProxyRealm{
		Realm:         "example.com",
		ProxyServers:  []int64{1, 2},
		FallbackOrder: 1,
		Status:        "enabled",
	}

	if err := repo.CreateRealm(ctx, realm); err != nil {
		t.Fatalf("failed to create realm: %v", err)
	}

	// Test realm extraction
	testCases := []struct {
		username   string
		shouldFind bool
	}{
		{"user@example.com", true},
		{"test@example.com", true},
		{"user@other.com", false},
		{"nouser", false},
	}

	for _, tc := range testCases {
		result, err := repo.FindRealmForUsername(ctx, tc.username)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tc.shouldFind && result == nil {
			t.Errorf("expected to find realm for %s", tc.username)
		}
		if !tc.shouldFind && result != nil {
			t.Errorf("expected no realm for %s, found %s", tc.username, result.Realm)
		}
	}
}
