package quota

import (
	"context"
	"errors"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

var (
	ErrQuotaExceeded       = errors.New("quota exceeded")
	ErrMaxSessionsExceeded = errors.New("max sessions exceeded")
	ErrTenantLimitExceeded = errors.New("tenant limit exceeded")
)

type QuotaService struct {
	db    *gorm.DB
	cache *UsageCache
}

func NewQuotaService(db *gorm.DB, cache *UsageCache) *QuotaService {
	return &QuotaService{
		db:    db,
		cache: cache,
	}
}

// GetQuota retrieves quota for a tenant (with cache)
func (s *QuotaService) GetQuota(tenantID int64) (*domain.ProviderQuota, error) {
	// Try cache first
	if s.cache != nil {
		if quota := s.cache.GetQuota(tenantID); quota != nil {
			return quota, nil
		}
	}

	var quota domain.ProviderQuota
	err := s.db.Where("tenant_id = ?", tenantID).First(&quota).Error
	if err != nil {
		// Return default quota if not found
		return &domain.ProviderQuota{
			TenantID:         tenantID,
			MaxUsers:         1000,
			MaxOnlineUsers:   500,
			MaxNAS:           100,
			MaxStorage:       100,
			MaxAuthPerSecond: 100,
			MaxAcctPerSecond: 200,
		}, nil
	}

	if s.cache != nil {
		s.cache.SetQuota(tenantID, &quota)
	}

	return &quota, nil
}

// GetUsage retrieves current usage for a tenant
func (s *QuotaService) GetUsage(tenantID int64) (*domain.ProviderUsage, error) {
	// Try cache first
	if s.cache != nil {
		if usage := s.cache.GetUsage(tenantID); usage != nil {
			return usage, nil
		}
	}

	usage := &domain.ProviderUsage{TenantID: tenantID}

	// Count users - use int64 then convert
	var userCount, onlineUserCount, nasCount, sessionCount int64
	s.db.Table("radius_user").Where("tenant_id = ?", tenantID).Count(&userCount)
	s.db.Table("radius_user").Where("tenant_id = ? AND status = ?", tenantID, "enabled").Count(&onlineUserCount)
	s.db.Table("net_nas").Where("tenant_id = ?", tenantID).Count(&nasCount)
	s.db.Table("radius_online").Where("tenant_id = ?", tenantID).Count(&sessionCount)

	usage.CurrentUsers = int(userCount)
	usage.CurrentOnlineUsers = int(onlineUserCount)
	usage.CurrentNAS = int(nasCount)
	// Note: sessionCount overwrites onlineUserCount if we want online sessions instead
	// For now, keep onlineUserCount as enabled users

	if s.cache != nil {
		s.cache.SetUsage(tenantID, usage)
	}

	return usage, nil
}

// CheckUserQuota checks if tenant can create more users
func (s *QuotaService) CheckUserQuota(ctx context.Context, tenantID int64) error {
	quota, err := s.GetQuota(tenantID)
	if err != nil {
		return err
	}

	usage, err := s.GetUsage(tenantID)
	if err != nil {
		return err
	}

	if usage.CurrentUsers >= quota.MaxUsers {
		zap.S().Warn("User quota exceeded",
			zap.Int64("tenant_id", tenantID),
			zap.Int("current", usage.CurrentUsers),
			zap.Int("limit", quota.MaxUsers))
		return ErrQuotaExceeded
	}

	return nil
}

// CheckSessionQuota checks if tenant can have more sessions
func (s *QuotaService) CheckSessionQuota(ctx context.Context, tenantID int64) error {
	quota, err := s.GetQuota(tenantID)
	if err != nil {
		return err
	}

	usage, err := s.GetUsage(tenantID)
	if err != nil {
		return err
	}

	if usage.CurrentOnlineUsers >= quota.MaxOnlineUsers {
		zap.S().Warn("Session quota exceeded",
			zap.Int64("tenant_id", tenantID),
			zap.Int("current", usage.CurrentOnlineUsers),
			zap.Int("limit", quota.MaxOnlineUsers))
		return ErrMaxSessionsExceeded
	}

	return nil
}

// UpdateUsage updates usage metrics for a tenant
func (s *QuotaService) UpdateUsage(tenantID int64) {
	usage, err := s.GetUsage(tenantID)
	if err != nil {
		return
	}

	// Save or update usage record
	s.db.Where("tenant_id = ?", tenantID).
		Assign(usage).
		FirstOrCreate(&domain.ProviderUsage{})
}
