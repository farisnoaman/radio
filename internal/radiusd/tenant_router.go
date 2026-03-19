package radiusd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/cache"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"go.uber.org/zap"
)

const (
	defaultNasCacheTTL  = 5 * time.Minute
	defaultNasCacheSize = 1024
)

type TenantRouter struct {
	db      TenantRouterDB
	cache   *cache.TTLCache[*TenantCacheEntry]
	cacheMu sync.RWMutex
}

type TenantRouterDB interface {
	GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error)
}

type TenantCacheEntry struct {
	TenantID int64
	Nas      *domain.NetNas
}

func NewTenantRouter(db TenantRouterDB) *TenantRouter {
	return &TenantRouter{
		db:    db,
		cache: cache.NewTTLCache[*TenantCacheEntry](defaultNasCacheTTL, defaultNasCacheSize),
	}
}

func (r *TenantRouter) GetTenantForNAS(ctx context.Context, nasIP, identifier string) (int64, error) {
	cacheKey := r.cacheKey(nasIP, identifier)

	r.cacheMu.RLock()
	if entry, ok := r.cache.Get(cacheKey); ok {
		r.cacheMu.RUnlock()
		return entry.TenantID, nil
	}
	r.cacheMu.RUnlock()

	nas, err := r.db.GetByIPOrIdentifier(ctx, nasIP, identifier)
	if err != nil {
		return 0, fmt.Errorf("NAS not found for IP %s: %w", nasIP, err)
	}

	entry := &TenantCacheEntry{
		TenantID: nas.TenantID,
		Nas:      nas,
	}

	r.cacheMu.Lock()
	r.cache.Set(cacheKey, entry)
	r.cacheMu.Unlock()

	return nas.TenantID, nil
}

func (r *TenantRouter) GetNASWithTenant(ctx context.Context, nasIP, identifier string) (*TenantContext, error) {
	cacheKey := r.cacheKey(nasIP, identifier)

	r.cacheMu.RLock()
	if entry, ok := r.cache.Get(cacheKey); ok {
		r.cacheMu.RUnlock()
		return &TenantContext{
			TenantID: entry.TenantID,
			Tenant:   tenant.WithTenantID(ctx, entry.TenantID),
			NAS:      entry.Nas,
		}, nil
	}
	r.cacheMu.RUnlock()

	nas, err := r.db.GetByIPOrIdentifier(ctx, nasIP, identifier)
	if err != nil {
		return nil, fmt.Errorf("NAS not found: %w", err)
	}

	entry := &TenantCacheEntry{
		TenantID: nas.TenantID,
		Nas:      nas,
	}

	r.cacheMu.Lock()
	r.cache.Set(cacheKey, entry)
	r.cacheMu.Unlock()

	return &TenantContext{
		TenantID: nas.TenantID,
		Tenant:   tenant.WithTenantID(ctx, nas.TenantID),
		NAS:      nas,
	}, nil
}

func (r *TenantRouter) InvalidateCache(nasIP, identifier string) {
	r.cacheMu.Lock()
	r.cache.Delete(r.cacheKey(nasIP, identifier))
	r.cacheMu.Unlock()
	zap.S().Debugf("Invalidated tenant cache for NAS: %s|%s", nasIP, identifier)
}

func (r *TenantRouter) InvalidateAll() {
	r.cacheMu.Lock()
	r.cache.Clear()
	r.cacheMu.Unlock()
	zap.S().Info("Invalidated all tenant cache entries")
}

func (r *TenantRouter) cacheKey(ip, identifier string) string {
	return fmt.Sprintf("%s|%s", ip, identifier)
}

type TenantContext struct {
	TenantID int64
	Tenant   context.Context
	NAS      *domain.NetNas
}

func GetTenantFromContext(ctx context.Context) (int64, error) {
	return tenant.FromContext(ctx)
}

func GetTenantOrDefault(ctx context.Context) int64 {
	return tenant.GetTenantIDOrDefault(ctx)
}

func MustGetTenant(ctx context.Context) int64 {
	return tenant.MustFromContext(ctx)
}
