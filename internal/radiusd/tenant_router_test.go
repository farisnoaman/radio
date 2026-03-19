package radiusd

import (
	"context"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

type mockTenantRouterDB struct {
	nas *domain.NetNas
	err error
}

func (m *mockTenantRouterDB) GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.nas, nil
}

func TestTenantRouter_GetTenantForNAS(t *testing.T) {
	db := &mockTenantRouterDB{
		nas: &domain.NetNas{
			ID:       1,
			TenantID: 42,
			Ipaddr:   "192.168.1.1",
		},
	}
	router := NewTenantRouter(db)

	t.Run("successful tenant lookup", func(t *testing.T) {
		tenantID, err := router.GetTenantForNAS(context.Background(), "192.168.1.1", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if tenantID != 42 {
			t.Errorf("expected tenantID 42, got %d", tenantID)
		}
	})

	t.Run("cache hit", func(t *testing.T) {
		tenantID, err := router.GetTenantForNAS(context.Background(), "192.168.1.1", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if tenantID != 42 {
			t.Errorf("expected tenantID 42, got %d", tenantID)
		}
	})

	t.Run("NAS not found", func(t *testing.T) {
		db.err = context.DeadlineExceeded
		_, err := router.GetTenantForNAS(context.Background(), "192.168.1.99", "")
		if err == nil {
			t.Error("expected error for NAS not found")
		}
		db.err = nil
	})
}

func TestTenantRouter_GetNASWithTenant(t *testing.T) {
	db := &mockTenantRouterDB{
		nas: &domain.NetNas{
			ID:       1,
			TenantID: 123,
			Ipaddr:   "10.0.0.1",
			Name:     "Test NAS",
		},
	}
	router := NewTenantRouter(db)

	ctx, err := router.GetNASWithTenant(context.Background(), "10.0.0.1", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if ctx.TenantID != 123 {
		t.Errorf("expected TenantID 123, got %d", ctx.TenantID)
	}

	if ctx.NAS == nil {
		t.Error("expected NAS to be set")
	}

	if ctx.NAS.Name != "Test NAS" {
		t.Errorf("expected NAS name 'Test NAS', got %s", ctx.NAS.Name)
	}
}

func TestTenantRouter_InvalidateCache(t *testing.T) {
	db := &mockTenantRouterDB{
		nas: &domain.NetNas{
			ID:       1,
			TenantID: 1,
			Ipaddr:   "172.16.0.1",
		},
	}
	router := NewTenantRouter(db)

	router.GetTenantForNAS(context.Background(), "172.16.0.1", "")

	router.InvalidateCache("172.16.0.1", "")

	db.nas.TenantID = 2
	tenantID, _ := router.GetTenantForNAS(context.Background(), "172.16.0.1", "")
	if tenantID != 2 {
		t.Errorf("expected updated tenantID 2, got %d", tenantID)
	}
}

func TestTenantRouter_InvalidateAll(t *testing.T) {
	db := &mockTenantRouterDB{
		nas: &domain.NetNas{
			ID:       1,
			TenantID: 1,
			Ipaddr:   "172.16.0.1",
		},
	}
	router := NewTenantRouter(db)

	router.GetTenantForNAS(context.Background(), "172.16.0.1", "")
	router.GetTenantForNAS(context.Background(), "172.16.0.2", "")

	router.InvalidateAll()

	db.nas.TenantID = 99
	tenantID, _ := router.GetTenantForNAS(context.Background(), "172.16.0.1", "")
	if tenantID != 99 {
		t.Errorf("expected tenantID 99, got %d", tenantID)
	}
}

func TestNewTenantRouter(t *testing.T) {
	db := &mockTenantRouterDB{}
	router := NewTenantRouter(db)

	if router == nil {
		t.Error("expected non-nil router")
	}

	if router.db == nil {
		t.Error("expected non-nil db")
	}

	if router.cache == nil {
		t.Error("expected non-nil cache")
	}
}

func TestTenantCacheEntry(t *testing.T) {
	entry := &TenantCacheEntry{
		TenantID: 10,
		Nas: &domain.NetNas{
			ID:       5,
			TenantID: 10,
			Ipaddr:   "192.168.100.1",
		},
	}

	if entry.TenantID != 10 {
		t.Errorf("expected TenantID 10, got %d", entry.TenantID)
	}

	if entry.Nas == nil {
		t.Error("expected non-nil NAS")
	}
}

func TestTenantContextStruct(t *testing.T) {
	ctx := context.Background()
	nas := &domain.NetNas{
		ID:       1,
		TenantID: 55,
	}

	tc := &TenantContext{
		TenantID: 55,
		Tenant:   ctx,
		NAS:      nas,
	}

	if tc.TenantID != 55 {
		t.Errorf("expected TenantID 55, got %d", tc.TenantID)
	}

	if tc.NAS != nas {
		t.Error("expected NAS to be set")
	}
}

func TestCacheKey(t *testing.T) {
	router := &TenantRouter{}
	
	key := router.cacheKey("192.168.1.1", "router1")
	expected := "192.168.1.1|router1"
	if key != expected {
		t.Errorf("expected key %s, got %s", expected, key)
	}

	key = router.cacheKey("10.0.0.1", "")
	expected = "10.0.0.1|"
	if key != expected {
		t.Errorf("expected key %s, got %s", expected, key)
	}
}
