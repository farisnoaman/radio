package cache

import (
	"sync"
	"time"
)

type VoucherBatchCache struct {
	ttl     time.Duration
	mu      sync.RWMutex
	batches map[int64]time.Time
}

func NewVoucherBatchCache(ttl time.Duration) *VoucherBatchCache {
	return &VoucherBatchCache{
		ttl:     ttl,
		batches: make(map[int64]time.Time),
	}
}

func (c *VoucherBatchCache) AddBatch(batchID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batches[batchID] = time.Now().Add(c.ttl)
}

func (c *VoucherBatchCache) RemoveBatch(batchID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.batches, batchID)
}

func (c *VoucherBatchCache) IsActive(batchID int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	exp, ok := c.batches[batchID]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		return false
	}
	return true
}

func (c *VoucherBatchCache) RemoveExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, exp := range c.batches {
		if now.After(exp) {
			delete(c.batches, id)
		}
	}
}

func (c *VoucherBatchCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batches = make(map[int64]time.Time)
}
