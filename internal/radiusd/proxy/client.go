package proxy

import (
	"context"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
)

// ProxyClient handles forwarding RADIUS requests to upstream servers.
type ProxyClient struct {
	timeout time.Duration
}

// NewProxyClient creates a new proxy client.
func NewProxyClient(timeout time.Duration) *ProxyClient {
	return &ProxyClient{timeout: timeout}
}

// Forward forwards a RADIUS request to an upstream server.
func (c *ProxyClient) Forward(
	ctx context.Context,
	pkt *radius.Packet,
	server *domain.RadiusProxyServer,
) (*radius.Packet, error) {
	// Create RADIUS client
	client := &radius.Client{
		Retry: c.timeout,
	}

	// Build server address
	addr := fmt.Sprintf("%s:%d", server.Host, server.AuthPort)
	if pkt.Code == radius.CodeAccountingRequest {
		addr = fmt.Sprintf("%s:%d", server.Host, server.AcctPort)
	}

	// Set secret
	pkt.Secret = []byte(server.Secret)

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(server.TimeoutSec)*time.Second)
	defer cancel()

	// Exchange packet with upstream server
	response, err := client.Exchange(timeoutCtx, pkt, addr)
	if err != nil {
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}

	return response, nil
}

// extractRealm extracts the realm suffix from a username.
// Example: "user@example.com" -> "example.com"
func extractRealm(username string) string {
	for i := len(username) - 1; i >= 0; i-- {
		if username[i] == '@' {
			return username[i+1:]
		}
	}
	return ""
}
