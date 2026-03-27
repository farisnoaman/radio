package cache

import (
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

type VoucherBatchCache struct {
	ttl     time.Duration
	mu      sync.RWMutex
	batches map[int64]time.Time
	db      *gorm.DB
}

func NewVoucherBatchCache(ttl time.Duration) *VoucherBatchCache {
	return &VoucherBatchCache{
		ttl:     ttl,
		batches: make(map[int64]time.Time),
	}
}

func NewVoucherBatchCacheWithDB(ttl time.Duration, db *gorm.DB) *VoucherBatchCache {
	return &VoucherBatchCache{
		ttl:     ttl,
		batches: make(map[int64]time.Time),
		db:      db,
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
	exp, ok := c.batches[batchID]
	c.mu.RUnlock()

	if ok && !time.Now().After(exp) {
		return true
	}

	if c.db != nil {
		var batch domain.VoucherBatch
		err := c.db.Where("id = ? AND activated_at IS NOT NULL AND is_deleted = ?", batchID, false).First(&batch).Error
		if err == nil {
			c.AddBatch(batchID)
			return true
		}
	}

	return false
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
