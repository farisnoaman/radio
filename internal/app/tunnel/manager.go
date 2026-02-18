package tunnel

import (
	"fmt"
	"sync"

	"github.com/talkincode/toughradius/v9/config"

	"go.uber.org/zap"
)

type DefaultTunnelManager struct {
	cfg     config.TunnelConfig
	service TunnelService
	mu      sync.RWMutex
}

func NewTunnelManager(cfg config.TunnelConfig) *DefaultTunnelManager {
	return &DefaultTunnelManager{
		cfg: cfg,
	}
}

func (m *DefaultTunnelManager) StartTunnel() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.service != nil {
		return fmt.Errorf("tunnel already running")
	}

	if !m.cfg.Enabled {
		return fmt.Errorf("tunnel is disabled")
	}

	var service TunnelService
	switch m.cfg.Type {
	case "cloudflare":
		service = NewCloudflareTunnel(m.cfg)
	default:
		return fmt.Errorf("unsupported tunnel type: %s", m.cfg.Type)
	}

	if err := service.Start(); err != nil {
		return err
	}

	m.service = service
	zap.S().Infof("Tunnel started: %s", m.cfg.Type)
	return nil
}

func (m *DefaultTunnelManager) StopTunnel() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.service == nil {
		return nil
	}

	if err := m.service.Stop(); err != nil {
		return err
	}

	m.service = nil
	zap.S().Info("Tunnel stopped")
	return nil
}

func (m *DefaultTunnelManager) GetStatus() (*TunnelStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.service == nil {
		return &TunnelStatus{
			Status: "stopped",
		}, nil
	}

	return m.service.Status()
}

func (m *DefaultTunnelManager) UpdateConfig(cfg config.TunnelConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = cfg
	return nil
}
