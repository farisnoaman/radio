package maintenance

import (
	"context"
	"sync"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MaintenanceManager struct {
	db       *gorm.DB
	mu       sync.RWMutex
	isActive bool
}

func NewMaintenanceManager(db *gorm.DB) *MaintenanceManager {
	m := &MaintenanceManager{
		db: db,
	}
	// TODO: Load initial state from DB settings
	return m
}

func (m *MaintenanceManager) Enable() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isActive = true
	zap.S().Info("Maintenance mode enabled")
	return nil
}

func (m *MaintenanceManager) Disable() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isActive = false
	zap.S().Info("Maintenance mode disabled")
	return nil
}

func (m *MaintenanceManager) IsActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isActive
}

// DrainSessions disconnects all active sessions
// In a real implementation this would iterate over active sessions and send CoA Disconnect-Requests
func (m *MaintenanceManager) DrainSessions(ctx context.Context) error {
	zap.S().Info("Starting session drain...")
	
	// Example logic:
	// 1. Get all online sessions
	var sessions []domain.RadiusOnline
	if err := m.db.Find(&sessions).Error; err != nil {
		return err
	}
	
	count := 0
	for _, s := range sessions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 2. Call Disconnect (Needs access to RADIUS client/adminapi functions)
			// For now, we just log. In a full implementation, we'd inject a CoAClientservice.
			zap.S().Debugf("Draining session: %s", s.Username)
			count++
		}
	}
	
	zap.S().Infof("Drained %d sessions", count)
	return nil
}
