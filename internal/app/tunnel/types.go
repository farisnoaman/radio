package tunnel

import (
	"time"

	"github.com/talkincode/toughradius/v9/config"

)

// TunnelConfig is aliased from config package
type TunnelConfig = config.TunnelConfig


// TunnelStatus represents the current status of a tunnel
type TunnelStatus struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"` // "running", "stopped", "error"
	Error     string    `json:"error,omitempty"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TunnelService defines the interface for different tunnel implementations
type TunnelService interface {
	Start() error
	Stop() error
	Status() (*TunnelStatus, error)
	Type() string
}

// TunnelManager defines the interface for managing tunnels
type TunnelManager interface {
	StartTunnel() error
	StopTunnel() error
	GetStatus() (*TunnelStatus, error)
	UpdateConfig(cfg TunnelConfig) error
	GetConfig() TunnelConfig
}

