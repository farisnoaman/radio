package domain

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// RadiusProxyServer represents an upstream RADIUS server for proxying.
// The proxy forwards authentication and accounting requests to these servers
// based on realm-based routing rules.
//
// Example:
//	A proxy server might be an organization's Active Directory controller
//	or a third-party authentication service.
type RadiusProxyServer struct {
	ID         int64     `json:"id,string" gorm:"primaryKey"`
	TenantID   int64     `json:"tenant_id" gorm:"index"`
	Name       string    `json:"name" gorm:"size:200"`
	Host       string    `json:"host" gorm:"not null;size:255"` // IP or hostname
	AuthPort   int       `json:"auth_port" gorm:"default:1812"`
	AcctPort   int       `json:"acct_port" gorm:"default:1813"`
	Secret     string    `json:"secret" gorm:"not null;size:100"`
	Status     string    `json:"status" gorm:"default:enabled"` // enabled, disabled
	MaxConns   int       `json:"max_conns" gorm:"default:50"`    // Max concurrent connections
	TimeoutSec int       `json:"timeout_sec" gorm:"default:5"`  // Request timeout
	Priority   int       `json:"priority" gorm:"default:1"`     // Load balancing priority
	Remark     string    `json:"remark" gorm:"size:500"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (RadiusProxyServer) TableName() string {
	return "radius_proxy_server"
}

// Validate checks if the proxy server configuration is valid.
func (s *RadiusProxyServer) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.Host == "" {
		return errors.New("host is required")
	}
	if net.ParseIP(s.Host) == nil {
		// Not an IP address, could be a hostname
		// TODO: Add hostname validation
	}
	if s.Secret == "" {
		return errors.New("secret is required")
	}
	if s.AuthPort <= 0 || s.AuthPort > 65535 {
		return errors.New("invalid auth port")
	}
	if s.AcctPort <= 0 || s.AcctPort > 65535 {
		return errors.New("invalid acct port")
	}
	if s.TimeoutSec <= 0 || s.TimeoutSec > 60 {
		return errors.New("timeout must be between 1 and 60 seconds")
	}
	if s.MaxConns <= 0 || s.MaxConns > 1000 {
		return errors.New("max connections must be between 1 and 1000")
	}
	return nil
}

// GetAddress returns the network address for the proxy server.
func (s *RadiusProxyServer) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.AuthPort)
}

// RadiusProxyRealm defines a routing realm for proxy requests.
// Requests matching the realm suffix are forwarded to configured proxy servers.
//
// Example:
//	A realm "@example.com" forwards all requests with User-Name ending in
//	"@example.com" to the configured proxy servers.
type RadiusProxyRealm struct {
	ID            int64   `json:"id,string" gorm:"primaryKey"`
	TenantID      int64   `json:"tenant_id" gorm:"index"`
	Realm         string  `json:"realm" gorm:"not null;size:255;index"` // e.g., "example.com"
	ProxyServers  []int64 `json:"proxy_servers" gorm:"serializer:json"`   // Server IDs
	FallbackOrder int     `json:"fallback_order" gorm:"default:1"`        // Priority for failover
	Status        string  `json:"status" gorm:"default:enabled"`
	Remark        string  `json:"remark" gorm:"size:500"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (RadiusProxyRealm) TableName() string {
	return "radius_proxy_realm"
}

// Validate checks if the realm configuration is valid.
func (r *RadiusProxyRealm) Validate() error {
	if r.Realm == "" {
		return errors.New("realm is required")
	}
	if len(r.ProxyServers) == 0 {
		return errors.New("at least one proxy server is required")
	}
	if r.FallbackOrder < 1 {
		return errors.New("fallback order must be >= 1")
	}
	return nil
}

// ProxyRequestLog represents a logged proxied RADIUS request.
// Used for auditing and troubleshooting proxy operations.
type ProxyRequestLog struct {
	ID          int64     `json:"id,string" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index"`
	Realm       string    `json:"realm" gorm:"index"`
	Username    string    `json:"username" gorm:"size:255"`
	ServerID    int64     `json:"server_id" gorm:"index"`
	RequestType string    `json:"request_type"` // auth, acct
	Success     bool      `json:"success"`
	LatencyMs   int       `json:"latency_ms"`
	ErrorMsg    string    `json:"error_msg" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// TableName specifies the table name.
func (ProxyRequestLog) TableName() string {
	return "proxy_request_log"
}
