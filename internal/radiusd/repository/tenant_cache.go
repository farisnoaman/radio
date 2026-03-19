package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/radiusd/cache"
)

type TenantCacheConfig struct {
	UserTTL        time.Duration
	NasTTL         time.Duration
	SessionTTL     time.Duration
	MaxEntries     int
}

func DefaultTenantCacheConfig() *TenantCacheConfig {
	return &TenantCacheConfig{
		UserTTL:    10 * time.Second,
		NasTTL:     5 * time.Minute,
		SessionTTL: 2 * time.Second,
		MaxEntries: 4096,
	}
}

type TenantCache struct {
	config *TenantCacheConfig
	mu     sync.RWMutex
	caches map[int64]*TenantCacheSet
}

type TenantCacheSet struct {
	UserCache   *cache.TTLCache[interface{}]
	NasCache    *cache.TTLCache[interface{}]
	SessionCache *cache.TTLCache[int]
}

func NewTenantCache(config *TenantCacheConfig) *TenantCache {
	if config == nil {
		config = DefaultTenantCacheConfig()
	}
	return &TenantCache{
		config: config,
		caches: make(map[int64]*TenantCacheSet),
	}
}

func (tc *TenantCache) GetCache(tenantID int64) *TenantCacheSet {
	tc.mu.RLock()
	cacheSet, ok := tc.caches[tenantID]
	tc.mu.RUnlock()

	if ok {
		return cacheSet
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Double-check after acquiring write lock
	if cacheSet, ok := tc.caches[tenantID]; ok {
		return cacheSet
	}

	cacheSet = &TenantCacheSet{
		UserCache:    cache.NewTTLCache[interface{}](tc.config.UserTTL, tc.config.MaxEntries),
		NasCache:     cache.NewTTLCache[interface{}](tc.config.NasTTL, tc.config.MaxEntries),
		SessionCache: cache.NewTTLCache[int](tc.config.SessionTTL, tc.config.MaxEntries),
	}
	tc.caches[tenantID] = cacheSet
	return cacheSet
}

func (tc *TenantCache) UserCacheKey(tenantID int64, username string) string {
	return fmt.Sprintf("tenant:%d:user:%s", tenantID, username)
}

func (tc *TenantCache) NasCacheKey(tenantID int64, nasIP string) string {
	return fmt.Sprintf("tenant:%d:nas:%s", tenantID, nasIP)
}

func (tc *TenantCache) SessionCacheKey(tenantID int64, username string) string {
	return fmt.Sprintf("tenant:%d:session:%s", tenantID, username)
}

func (tc *TenantCache) GetUser(tenantID int64, username string) (interface{}, bool) {
	cacheSet := tc.GetCache(tenantID)
	return cacheSet.UserCache.Get(tc.UserCacheKey(tenantID, username))
}

func (tc *TenantCache) SetUser(tenantID int64, username string, user interface{}) {
	cacheSet := tc.GetCache(tenantID)
	cacheSet.UserCache.Set(tc.UserCacheKey(tenantID, username), user)
}

func (tc *TenantCache) DeleteUser(tenantID int64, username string) {
	cacheSet := tc.GetCache(tenantID)
	cacheSet.UserCache.Delete(tc.UserCacheKey(tenantID, username))
}

func (tc *TenantCache) GetNas(tenantID int64, nasIP string) (interface{}, bool) {
	cacheSet := tc.GetCache(tenantID)
	return cacheSet.NasCache.Get(tc.NasCacheKey(tenantID, nasIP))
}

func (tc *TenantCache) SetNas(tenantID int64, nasIP string, nas interface{}) {
	cacheSet := tc.GetCache(tenantID)
	cacheSet.NasCache.Set(tc.NasCacheKey(tenantID, nasIP), nas)
}

func (tc *TenantCache) DeleteNas(tenantID int64, nasIP string) {
	cacheSet := tc.GetCache(tenantID)
	cacheSet.NasCache.Delete(tc.NasCacheKey(tenantID, nasIP))
}

func (tc *TenantCache) GetSessionCount(tenantID int64, username string) (int, bool) {
	cacheSet := tc.GetCache(tenantID)
	return cacheSet.SessionCache.Get(tc.SessionCacheKey(tenantID, username))
}

func (tc *TenantCache) SetSessionCount(tenantID int64, username string, count int) {
	cacheSet := tc.GetCache(tenantID)
	cacheSet.SessionCache.Set(tc.SessionCacheKey(tenantID, username), count)
}

func (tc *TenantCache) IncrementSessionCount(tenantID int64, username string) int {
	cacheSet := tc.GetCache(tenantID)
	key := tc.SessionCacheKey(tenantID, username)
	count, _ := cacheSet.SessionCache.Get(key)
	cacheSet.SessionCache.Set(key, count+1)
	return count + 1
}

func (tc *TenantCache) DecrementSessionCount(tenantID int64, username string) int {
	cacheSet := tc.GetCache(tenantID)
	key := tc.SessionCacheKey(tenantID, username)
	count, _ := cacheSet.SessionCache.Get(key)
	if count > 0 {
		cacheSet.SessionCache.Set(key, count-1)
		return count - 1
	}
	return 0
}

func (tc *TenantCache) Clear(tenantID int64) {
	tc.mu.Lock()
	delete(tc.caches, tenantID)
	tc.mu.Unlock()
}

func (tc *TenantCache) ClearAll() {
	tc.mu.Lock()
	tc.caches = make(map[int64]*TenantCacheSet)
	tc.mu.Unlock()
}
