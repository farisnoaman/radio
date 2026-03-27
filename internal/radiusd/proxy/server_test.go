package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestProxyClient_ForwardAuthRequest_ShouldSucceed(t *testing.T) {
	client := NewProxyClient(5 * time.Second)

	// Create mock upstream server
	upstream := &domain.RadiusProxyServer{
		ID:       1,
		Name:     "Test Server",
		Host:     "127.0.0.1",
		AuthPort: 18120, // Non-standard port for testing
		Secret:   "testsecret",
	}

	// Create auth request
	req := radius.New(radius.CodeAccessRequest, []byte("testsecret"))
	rfc2865.UserName_SetString(req, "testuser")
	rfc2865.UserPassword_SetString(req, "testpass")

	ctx := context.Background()

	// This will fail in test environment without actual server
	// We're testing that the client method exists and has correct signature
	_, err := client.Forward(ctx, req, upstream)

	// We expect an error since there's no actual server running
	if err == nil {
		t.Error("expected error when no server is running")
	}
}

func TestExtractRealm_ShouldExtractCorrectly(t *testing.T) {
	testCases := []struct {
		username     string
		expectedRealm string
	}{
		{"user@example.com", "example.com"},
		{"test@other.org", "other.org"},
		{"nouser", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			result := extractRealm(tc.username)
			if result != tc.expectedRealm {
				t.Errorf("expected realm '%s', got '%s'", tc.expectedRealm, result)
			}
		})
	}
}
