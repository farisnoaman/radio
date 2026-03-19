package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/radiusd/cache"
)

type CacheConfig struct {
	L1Enabled     bool
	L1MaxEntries  int
	L1UserTTL     time.Duration
	L1NasTTL      time.Duration
	L1SessionTTL  time.Duration

	L2Enabled    bool
	L2RedisConfig *RedisCacheConfig
	L2UserTTL     time.Duration
	L2NasTTL      time.Duration
	L2SessionTTL  time.Duration

	ProviderCount    int
	UsersPerProvider int
	ConcurrentPerProvider int
}

func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		L1Enabled:            true,
		L1MaxEntries:        10000,
		L1UserTTL:            10 * time.Second,
		L1NasTTL:             5 * time.Minute,
		L1SessionTTL:         2 * time.Second,

		L2Enabled:    true,
		L2RedisConfig: DefaultRedisCacheConfig(),
		L2UserTTL:     30 * time.Second,
		L2NasTTL:      10 * time.Minute,
		L2SessionTTL:  5 * time.Second,

		ProviderCount:         100,
		UsersPerProvider:      5000,
		ConcurrentPerProvider: 1500,
	}
}

type MultiLayerCache struct {
	config *CacheConfig
	L1     *TenantCache
	L2     *RedisCache
	mu     sync.RWMutex
}

type TenantCache struct {
	defaultTTL time.Duration
	mu         sync.RWMutex
	caches     map[int64]*CacheSet
	maxEntries int
}

type CacheSet struct {
	UserCache    *cache.TTLCache[interface{}]
	NasCache     *cache.TTLCache[interface{}]
	SessionCache *cache.TTLCache[int]
}

func NewTenantCache(defaultTTL time.Duration, maxEntries int) *TenantCache {
	if maxEntries <= 0 {
		maxEntries = 1024
	}
	return &TenantCache{
		defaultTTL: defaultTTL,
		caches:     make(map[int64]*CacheSet),
		maxEntries: maxEntries,
	}
}

func (c *TenantCache) getCacheSet(tenantID int64) *CacheSet {
	c.mu.RLock()
	cs, ok := c.caches[tenantID]
	c.mu.RUnlock()

	if ok {
		return cs
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if cs, ok := c.caches[tenantID]; ok {
		return cs
	}

	cs = &CacheSet{
		UserCache:    cache.NewTTLCache[interface{}](c.defaultTTL, c.maxEntries),
		NasCache:     cache.NewTTLCache[interface{}](c.defaultTTL*3, c.maxEntries),
		SessionCache: cache.NewTTLCache[int](time.Second, c.maxEntries*2),
	}
	c.caches[tenantID] = cs
	return cs
}

func (c *TenantCache) GetUser(tenantID int64, username string) (interface{}, bool) {
	cs := c.getCacheSet(tenantID)
	return cs.UserCache.Get(c.userKey(tenantID, username))
}

func (c *TenantCache) SetUser(tenantID int64, username string, value interface{}) {
	cs := c.getCacheSet(tenantID)
	cs.UserCache.Set(c.userKey(tenantID, username), value)
}

func (c *TenantCache) DeleteUser(tenantID int64, username string) {
	cs := c.getCacheSet(tenantID)
	cs.UserCache.Delete(c.userKey(tenantID, username))
}

func (c *TenantCache) GetNAS(tenantID int64, nasIP string) (interface{}, bool) {
	cs := c.getCacheSet(tenantID)
	return cs.NasCache.Get(c.nasKey(tenantID, nasIP))
}

func (c *TenantCache) SetNAS(tenantID int64, nasIP string, value interface{}) {
	cs := c.getCacheSet(tenantID)
	cs.NasCache.Set(c.nasKey(tenantID, nasIP), value)
}

func (c *TenantCache) DeleteNAS(tenantID int64, nasIP string) {
	cs := c.getCacheSet(tenantID)
	cs.NasCache.Delete(c.nasKey(tenantID, nasIP))
}

func (c *TenantCache) GetSessionCount(tenantID int64, username string) (int, bool) {
	cs := c.getCacheSet(tenantID)
	return cs.SessionCache.Get(c.sessionKey(tenantID, username))
}

func (c *TenantCache) SetSessionCount(tenantID int64, username string, count int) {
	cs := c.getCacheSet(tenantID)
	cs.SessionCache.Set(c.sessionKey(tenantID, username), count)
}

func (c *TenantCache) IncrementSessionCount(tenantID int64, username string) int {
	cs := c.getCacheSet(tenantID)
	key := c.sessionKey(tenantID, username)
	count, _ := cs.SessionCache.Get(key)
	cs.SessionCache.Set(key, count+1)
	return count + 1
}

func (c *TenantCache) DecrementSessionCount(tenantID int64, username string) int {
	cs := c.getCacheSet(tenantID)
	key := c.sessionKey(tenantID, username)
	count, _ := cs.SessionCache.Get(key)
	if count > 0 {
		cs.SessionCache.Set(key, count-1)
		return count - 1
	}
	return 0
}

func (c *TenantCache) ClearTenant(tenantID int64) {
	c.mu.Lock()
	delete(c.caches, tenantID)
	c.mu.Unlock()
}

func (c *TenantCache) ClearAll() {
	c.mu.Lock()
	c.caches = make(map[int64]*CacheSet)
	c.mu.Unlock()
}

func (c *TenantCache) userKey(tenantID int64, username string) string {
	return fmt.Sprintf("u:%d:%s", tenantID, username)
}

func (c *TenantCache) nasKey(tenantID int64, nasIP string) string {
	return fmt.Sprintf("n:%d:%s", tenantID, nasIP)
}

func (c *TenantCache) sessionKey(tenantID int64, username string) string {
	return fmt.Sprintf("s:%d:%s", tenantID, username)
}

func NewMultiLayerCache(config *CacheConfig) (*MultiLayerCache, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	mlc := &MultiLayerCache{
		config: config,
	}

	if config.L1Enabled {
		mlc.L1 = NewTenantCache(config.L1UserTTL, config.L1MaxEntries)
	}

	if config.L2Enabled {
		redisCache, err := NewRedisCache(config.L2RedisConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %w", err)
		}
		mlc.L2 = redisCache
	}

	return mlc, nil
}

func (c *MultiLayerCache) Close() error {
	if c.L2 != nil {
		return c.L2.Close()
	}
	return nil
}

func (c *MultiLayerCache) IsL1Enabled() bool {
	return c.config.L1Enabled
}

func (c *MultiLayerCache) IsL2Enabled() bool {
	return c.config.L2Enabled
}

func (c *MultiLayerCache) GetUser(ctx context.Context, tenantID int64, username string) (*CachedUser, error) {
	if c.L1 != nil {
		if data, ok := c.L1.GetUser(tenantID, username); ok {
			if user, ok := data.(*CachedUser); ok {
				return user, nil
			}
		}
	}

	if c.L2 != nil {
		user, err := c.L2.GetUser(ctx, tenantID, username)
		if err == nil && user != nil {
			if c.L1 != nil {
				c.L1.SetUser(tenantID, username, user)
			}
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (c *MultiLayerCache) SetUser(ctx context.Context, tenantID int64, username string, user *CachedUser) error {
	if c.L1 != nil {
		c.L1.SetUser(tenantID, username, user)
	}

	if c.L2 != nil {
		return c.L2.SetUser(ctx, tenantID, username, user, c.config.L2UserTTL)
	}

	return nil
}

func (c *MultiLayerCache) DeleteUser(ctx context.Context, tenantID int64, username string) error {
	if c.L1 != nil {
		c.L1.DeleteUser(tenantID, username)
	}

	if c.L2 != nil {
		return c.L2.DeleteUser(ctx, tenantID, username)
	}

	return nil
}

func (c *MultiLayerCache) GetNAS(ctx context.Context, tenantID int64, nasIP string) (*CachedNAS, error) {
	if c.L1 != nil {
		if data, ok := c.L1.GetNAS(tenantID, nasIP); ok {
			if nas, ok := data.(*CachedNAS); ok {
				return nas, nil
			}
		}
	}

	if c.L2 != nil {
		nas, err := c.L2.GetNAS(ctx, tenantID, nasIP)
		if err == nil && nas != nil {
			if c.L1 != nil {
				c.L1.SetNAS(tenantID, nasIP, nas)
			}
			return nas, nil
		}
	}

	return nil, fmt.Errorf("NAS not found")
}

func (c *MultiLayerCache) SetNAS(ctx context.Context, tenantID int64, nasIP string, nas *CachedNAS) error {
	if c.L1 != nil {
		c.L1.SetNAS(tenantID, nasIP, nas)
	}

	if c.L2 != nil {
		return c.L2.SetNAS(ctx, tenantID, nasIP, nas, c.config.L2NasTTL)
	}

	return nil
}

func (c *MultiLayerCache) DeleteNAS(ctx context.Context, tenantID int64, nasIP string) error {
	if c.L1 != nil {
		c.L1.DeleteNAS(tenantID, nasIP)
	}

	if c.L2 != nil {
		return c.L2.DeleteNAS(ctx, tenantID, nasIP)
	}

	return nil
}

func (c *MultiLayerCache) IncrementSessionCount(ctx context.Context, tenantID int64, username string) (int64, error) {
	if c.L1 != nil {
		c.L1.IncrementSessionCount(tenantID, username)
	}

	if c.L2 != nil {
		return c.L2.IncrementSessionCount(ctx, tenantID, username)
	}

	return 0, nil
}

func (c *MultiLayerCache) DecrementSessionCount(ctx context.Context, tenantID int64, username string) (int64, error) {
	if c.L1 != nil {
		c.L1.DecrementSessionCount(tenantID, username)
	}

	if c.L2 != nil {
		return c.L2.DecrementSessionCount(ctx, tenantID, username)
	}

	return 0, nil
}

func (c *MultiLayerCache) GetSessionCount(ctx context.Context, tenantID int64, username string) (int64, error) {
	if c.L1 != nil {
		if count, ok := c.L1.GetSessionCount(tenantID, username); ok {
			return int64(count), nil
		}
	}

	if c.L2 != nil {
		return c.L2.GetSessionCount(ctx, tenantID, username)
	}

	return 0, nil
}

func (c *MultiLayerCache) ClearTenantCache(tenantID int64) {
	if c.L1 != nil {
		c.L1.ClearTenant(tenantID)
	}
}

func (c *MultiLayerCache) ClearAllCaches(ctx context.Context) {
	if c.L1 != nil {
		c.L1.ClearAll()
	}

	if c.L2 != nil && ctx != nil {
		c.L2.DeleteTenantKeys(ctx, -1)
	}
}

func (c *MultiLayerCache) InvalidateUser(ctx context.Context, tenantID int64, username string) {
	c.DeleteUser(ctx, tenantID, username)
}

func (c *MultiLayerCache) InvalidateNAS(ctx context.Context, tenantID int64, nasIP string) {
	c.DeleteNAS(ctx, tenantID, nasIP)
}

func (c *MultiLayerCache) InvalidateTenant(tenantID int64) {
	c.ClearTenantCache(tenantID)
}

type CacheStats struct {
	L1Enabled     bool
	L2Enabled     bool
	L1CacheCount  int
	MemoryUsageMB float64
	HitRate       float64
}

func (c *MultiLayerCache) GetStats() *CacheStats {
	stats := &CacheStats{
		L1Enabled: c.config.L1Enabled,
		L2Enabled: c.config.L2Enabled,
	}

	if c.L1 != nil {
		stats.L1CacheCount = len(c.L1.caches)
	}

	return stats
}
