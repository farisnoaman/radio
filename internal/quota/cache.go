package quota

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

type UsageCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewUsageCache(client *redis.Client) *UsageCache {
	return &UsageCache{
		client: client,
		ttl:    5 * time.Minute, // Cache for 5 minutes
	}
}

// GetQuota retrieves quota from cache
func (c *UsageCache) GetQuota(tenantID int64) *domain.ProviderQuota {
	ctx := context.Background()
	key := c.quotaKey(tenantID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}

	var quota domain.ProviderQuota
	if err := json.Unmarshal(data, &quota); err != nil {
		return nil
	}

	return &quota
}

// SetQuota stores quota in cache
func (c *UsageCache) SetQuota(tenantID int64, quota *domain.ProviderQuota) {
	ctx := context.Background()
	key := c.quotaKey(tenantID)

	data, _ := json.Marshal(quota)
	c.client.Set(ctx, key, data, c.ttl)
}

// GetUsage retrieves usage from cache
func (c *UsageCache) GetUsage(tenantID int64) *domain.ProviderUsage {
	ctx := context.Background()
	key := c.usageKey(tenantID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}

	var usage domain.ProviderUsage
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil
	}

	return &usage
}

// SetUsage stores usage in cache
func (c *UsageCache) SetUsage(tenantID int64, usage *domain.ProviderUsage) {
	ctx := context.Background()
	key := c.usageKey(tenantID)

	data, _ := json.Marshal(usage)
	c.client.Set(ctx, key, data, c.ttl)
}

// Invalidate clears cache for a tenant
func (c *UsageCache) Invalidate(tenantID int64) {
	ctx := context.Background()
	c.client.Del(ctx, c.quotaKey(tenantID), c.usageKey(tenantID))
}

func (c *UsageCache) quotaKey(tenantID int64) string {
	return fmt.Sprintf("quota:%d", tenantID)
}

func (c *UsageCache) usageKey(tenantID int64) string {
	return fmt.Sprintf("usage:%d", tenantID)
}
