package tunnel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
)

// MockTunnelService for testing
type MockTunnelService struct {
	failStart bool
}

func (m *MockTunnelService) Start() error {
	if m.failStart {
		return errors.New("start failed")
	}
	return nil
}

func (m *MockTunnelService) Stop() error {
	return nil
}

func (m *MockTunnelService) Status() (*TunnelStatus, error) {
	return &TunnelStatus{Status: "running"}, nil
}

func (m *MockTunnelService) Type() string {
	return "mock"
}

func TestDefaultTunnelManager_StartStop(t *testing.T) {
	cfg := config.TunnelConfig{
		Enabled: true,
		Type:    "cloudflare",
		Token:   "dummy",
	}

	// Note: We can't easily inject MockTunnelService into DefaultTunnelManager
	// without changing the StartTunnel implementation to use a factory or interface.
	// For this test, we might only be able to test the validation logic
	// unless we refactor DefaultTunnelManager to accept a service factory.

	// However, for now, let's test the config validation and state.

	manager := NewTunnelManager(cfg)

	// Test Start with invalid type
	manager.cfg.Type = "invalid"
	err := manager.StartTunnel()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported tunnel type")

	// Test Start with disabled config
	manager.cfg.Enabled = false
	err = manager.StartTunnel()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestDefaultTunnelManager_UpdateConfig(t *testing.T) {
	manager := NewTunnelManager(config.TunnelConfig{Enabled: false})

	newCfg := config.TunnelConfig{
		Enabled: true,
		Type:    "cloudflare",
		Token:   "new-token",
	}

	err := manager.UpdateConfig(newCfg)
	assert.NoError(t, err)
	assert.Equal(t, true, manager.cfg.Enabled)
	assert.Equal(t, "new-token", manager.cfg.Token)
}
