package quota

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
)

type AlertService struct {
	quotaService *QuotaService
}

func NewAlertService(quotaService *QuotaService) *AlertService {
	return &AlertService{
		quotaService: quotaService,
	}
}

// CheckQuotaUsage monitors all providers and sends alerts
func (s *AlertService) CheckQuotaUsage() {
	var providers []domain.Provider
	s.quotaService.db.Find(&providers)

	for _, provider := range providers {
		quota, _ := s.quotaService.GetQuota(provider.ID)
		usage, _ := s.quotaService.GetUsage(provider.ID)

		// Calculate usage percentages
		userPercent := float64(usage.CurrentUsers) / float64(quota.MaxUsers) * 100
		sessionPercent := float64(usage.CurrentOnlineUsers) / float64(quota.MaxOnlineUsers) * 100

		// Send alerts if approaching limits
		if userPercent > 80 {
			s.sendQuotaWarning(provider.ID, "users", userPercent, usage.CurrentUsers, quota.MaxUsers)
		}
		if sessionPercent > 80 {
			s.sendQuotaWarning(provider.ID, "sessions", sessionPercent, usage.CurrentOnlineUsers, quota.MaxOnlineUsers)
		}
	}
}

func (s *AlertService) sendQuotaWarning(tenantID int64, resourceType string, percent float64, current, limit int) {
	zap.S().Warn("Quota warning",
		zap.Int64("tenant_id", tenantID),
		zap.String("resource", resourceType),
		zap.Float64("percent", percent),
		zap.Int("current", current),
		zap.Int("limit", limit))
}

// StartBackgroundMonitoring starts periodic quota checks
func (s *AlertService) StartBackgroundMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CheckQuotaUsage()
		}
	}()
}
