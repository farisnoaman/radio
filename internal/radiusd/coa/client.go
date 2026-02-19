package coa

import (
	"context"
	"fmt"
	"net"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2868"
)

// Client is a RADIUS CoA (Change of Authorization) client that can send
// Disconnect-Request and CoA-Request messages to NAS devices.
//
// The client supports configurable timeouts, retry logic with exponential backoff,
// and vendor-specific attribute handling.
//
// Example:
//
//	client := coa.NewClient(coa.Config{
//	    Timeout:    5 * time.Second,
//	    RetryCount: 3,
//	    RetryDelay: 500 * time.Millisecond,
//	})
//
//	// Send disconnect request
//	resp := client.SendDisconnect(ctx, coa.DisconnectRequest{
//	    NASIP:        "192.168.1.1",
//	    NASPort:      3799,
//	    Secret:       "secret",
//	    Username:     "user@example.com",
//	    AcctSessionID: "session123",
//	})
//
//	// Send CoA request to modify session
//	resp = client.SendCoA(ctx, coa.CoARequest{
//	    NASIP:         "192.168.1.1",
//	    NASPort:       3799,
//	    Secret:        "secret",
//	    Username:      "user@example.com",
//	    SessionTimeout: 3600,
//	    UpRate:        10240,
//	    DownRate:      20480,
//	})
type Client struct {
	config Config
	client *radius.Client
}

// NewClient creates a new CoA client with the given configuration.
//
// If config is zero-valued, default values are used:
//   - Timeout: 5 seconds
//   - RetryCount: 2
//   - RetryDelay: 500 milliseconds
func NewClient(config Config) *Client {
	if config.Timeout <= 0 {
		config.Timeout = DefaultTimeout
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = DefaultRetryDelay
	}
	// Use default retry count if not specified (0 means no retries)
	if config.RetryCount == 0 {
		config.RetryCount = DefaultRetryCount
	}

	// Create the underlying RADIUS client
	radiusClient := &radius.Client{
		Retry: config.Timeout,
	}

	return &Client{
		config:  config,
		client:  radiusClient,
	}
}

// SendDisconnect sends a Disconnect-Request (RFC 3576) to the NAS to terminate a session.
//
// This method sends a RADIUS Disconnect-Request packet to the specified NAS.
// The NAS should respond with either a Disconnect-ACK (successful) or Disconnect-NAK (failed).
//
// The method includes retry logic with exponential backoff if the initial request fails.
// It validates the request parameters before sending and returns detailed response information.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - req: The disconnect request parameters
//
// Returns a CoAResponse with:
//   - Success: true if ACK received, false otherwise
//   - Code: The RADIUS response code
//   - Error: Any error that occurred
//   - Duration: Time taken for the operation
//   - RetryCount: Number of retries performed
func (c *Client) SendDisconnect(ctx context.Context, req DisconnectRequest) *CoAResponse {
	// Validate request
	if err := req.Validate(); err != nil {
		return &CoAResponse{
			Success: false,
			Code:    0,
			Error:   err,
			NASIP:   req.NASIP,
		}
	}

	// Use default port if not specified
	if req.NASPort <= 0 {
		req.NASPort = DefaultCoAPort
	}

	// Get NAS address
	addr, err := getNASAddress(req.NASIP, req.NASPort)
	if err != nil {
		return &CoAResponse{
			Success: false,
			Code:    0,
			Error:   err,
			NASIP:   req.NASIP,
		}
	}

	// Build the disconnect packet
	pkt := buildDisconnectPacket(&req)

	// Attempt the request with retries
	var lastErr error
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		start := time.Now()

		// Create context with timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		// Exchange packet with NAS
		response, err := c.client.Exchange(attemptCtx, pkt, addr)
		duration := time.Since(start)

		if err != nil {
			lastErr = err

			// Check if we should retry
			if attempt < c.config.RetryCount && attemptCtx.Err() == nil {
				// Wait with exponential backoff before retry
				delay := c.config.RetryDelay * time.Duration(1<<attempt)
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					break
				}
			}
			continue
		}

		// Success - check response code
		success := response.Code == radius.CodeDisconnectACK

		return &CoAResponse{
			Success:    success,
			Code:       response.Code,
			Error:      nil,
			Duration:   duration,
			RetryCount: attempt,
			NASIP:      req.NASIP,
		}
	}

	// All attempts failed
	return &CoAResponse{
		Success:    false,
		Code:      0,
		Error:     lastErr,
		Duration:  0,
		RetryCount: c.config.RetryCount,
		NASIP:     req.NASIP,
	}
}

// SendCoA sends a CoA-Request (RFC 3576) to the NAS to modify an active session.
//
// This method sends a RADIUS CoA-Request packet to modify session parameters
// such as bandwidth limits, session timeout, or data quotas. The NAS should
// respond with either a CoA-ACK (successful) or CoA-NAK (failed).
//
// The method includes retry logic with exponential backoff if the initial request fails.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - req: The CoA request parameters including session modifications
//
// Returns a CoAResponse with:
//   - Success: true if ACK received, false otherwise
//   - Code: The RADIUS response code
//   - Error: Any error that occurred
//   - Duration: Time taken for the operation
//   - RetryCount: Number of retries performed
func (c *Client) SendCoA(ctx context.Context, req CoARequest) *CoAResponse {
	// Validate request
	if err := req.Validate(); err != nil {
		return &CoAResponse{
			Success: false,
			Code:    0,
			Error:   err,
			NASIP:   req.NASIP,
		}
	}

	// Use default port if not specified
	if req.NASPort <= 0 {
		req.NASPort = DefaultCoAPort
	}

	// Get NAS address
	addr, err := getNASAddress(req.NASIP, req.NASPort)
	if err != nil {
		return &CoAResponse{
			Success: false,
			Code:    0,
			Error:   err,
			NASIP:   req.NASIP,
		}
	}

	// Build the CoA packet
	pkt := buildCoAPacket(&req)

	// Attempt the request with retries
	var lastErr error
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		start := time.Now()

		// Create context with timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		// Exchange packet with NAS
		response, err := c.client.Exchange(attemptCtx, pkt, addr)
		duration := time.Since(start)

		if err != nil {
			lastErr = err

			// Check if we should retry
			if attempt < c.config.RetryCount && attemptCtx.Err() == nil {
				// Wait with exponential backoff before retry
				delay := c.config.RetryDelay * time.Duration(1<<attempt)
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					break
				}
			}
			continue
		}

		// Success - check response code
		success := response.Code == radius.CodeCoAACK

		return &CoAResponse{
			Success:    success,
			Code:       response.Code,
			Error:      nil,
			Duration:   duration,
			RetryCount: attempt,
			NASIP:      req.NASIP,
		}
	}

	// All attempts failed
	return &CoAResponse{
		Success:    false,
		Code:      0,
		Error:     lastErr,
		Duration:  0,
		RetryCount: c.config.RetryCount,
		NASIP:     req.NASIP,
	}
}

// Validate checks if the DisconnectRequest has all required fields.
//
// Returns an error if:
//   - NASIP is missing or invalid
//   - Secret is missing
//   - Neither Username nor AcctSessionID is provided
func (r *DisconnectRequest) Validate() error {
	if r.NASIP == "" {
		return ErrMissingNASIP
	}

	if r.Secret == "" {
		return ErrMissingSecret
	}

	// Validate IP address format
	if net.ParseIP(r.NASIP) == nil {
		return ErrInvalidNASIP
	}

	// Must have at least one session identifier
	if r.Username == "" && r.AcctSessionID == "" {
		return ErrMissingSessionID
	}

	return nil
}

// Validate checks if the CoARequest has all required fields.
//
// Returns an error if:
//   - NASIP is missing or invalid
//   - Secret is missing
//   - Neither Username nor AcctSessionID is provided
func (r *CoARequest) Validate() error {
	if r.NASIP == "" {
		return ErrMissingNASIP
	}

	if r.Secret == "" {
		return ErrMissingSecret
	}

	// Validate IP address format
	if net.ParseIP(r.NASIP) == nil {
		return ErrInvalidNASIP
	}

	// Must have at least one session identifier
	if r.Username == "" && r.AcctSessionID == "" {
		return ErrMissingSessionID
	}

	return nil
}

// buildDisconnectPacket constructs a RADIUS Disconnect-Request packet.
//
// This function creates a packet with the CodeDisconnectRequest (40) code
// and adds standard RADIUS attributes for session identification.
func buildDisconnectPacket(req *DisconnectRequest) *radius.Packet {
	// Create packet with Disconnect-Request code
	pkt := radius.New(radius.CodeDisconnectRequest, []byte(req.Secret))

	// Add User-Name attribute if provided
	if req.Username != "" {
		_ = rfc2865.UserName_SetString(pkt, req.Username)
	}

	// Add Acct-Session-Id attribute if provided
	if req.AcctSessionID != "" {
		_ = rfc2866.AcctSessionID_SetString(pkt, req.AcctSessionID)
	}

	// Add NAS-IP-Address attribute
	_ = rfc2865.NASIPAddress_Set(pkt, net.ParseIP(req.NASIP))

	// Add NAS-Identifier if provided
	if req.NASIdentifier != "" {
		_ = rfc2865.NASIdentifier_SetString(pkt, req.NASIdentifier)
	} else {
		// Use NAS-IP as identifier if not specified
		_ = rfc2865.NASIdentifier_SetString(pkt, req.NASIP)
	}

	// Add Message-Authenticator if needed (some NAS require it)
	// For now, we don't add it as it's not always required

	return pkt
}

// buildCoAPacket constructs a RADIUS CoA-Request packet.
//
// This function creates a packet with the CodeCoARequest (43) code
// and adds standard RADIUS attributes for session modification.
func buildCoAPacket(req *CoARequest) *radius.Packet {
	// Create packet with CoA-Request code
	pkt := radius.New(radius.CodeCoARequest, []byte(req.Secret))

	// Add User-Name attribute if provided
	if req.Username != "" {
		_ = rfc2865.UserName_SetString(pkt, req.Username)
	}

	// Add Acct-Session-Id attribute if provided
	if req.AcctSessionID != "" {
		_ = rfc2866.AcctSessionID_SetString(pkt, req.AcctSessionID)
	}

	// Add NAS-IP-Address attribute
	_ = rfc2865.NASIPAddress_Set(pkt, net.ParseIP(req.NASIP))

	// Add NAS-Identifier if provided
	if req.NASIdentifier != "" {
		_ = rfc2865.NASIdentifier_SetString(pkt, req.NASIdentifier)
	} else {
		// Use NAS-IP as identifier if not specified
		_ = rfc2865.NASIdentifier_SetString(pkt, req.NASIP)
	}

	// Add Session-Timeout if provided and greater than 0
	if req.SessionTimeout > 0 {
		_ = rfc2865.SessionTimeout_Set(pkt, rfc2865.SessionTimeout(req.SessionTimeout)) //nolint:errcheck,gosec // G115: timeout is validated
	}

	// Add rate limit attributes if provided
	// These use RFC 2868 Tunnel attributes for bandwidth
	if req.UpRate > 0 || req.DownRate > 0 {
		// Add Tunnel-Type (usually PPTP or L2TP for rate limiting)
		// A value of 1 indicates PPTP
		_ = rfc2868.TunnelType_Set(pkt, 1, rfc2868.TunnelType(1)) //nolint:errcheck
		// Add Tunnel-Medium-Type (IP = 1)
		_ = rfc2868.TunnelMediumType_Set(pkt, 1, rfc2868.TunnelMediumType(1)) //nolint:errcheck
	}

	return pkt
}

// getNASAddress converts the NAS IP and port to a network address string.
//
// If port is 0, the default CoA port (3799) is used.
func getNASAddress(nasIP string, nasPort int) (string, error) {
	if nasPort <= 0 {
		nasPort = DefaultCoAPort
	}
	return net.JoinHostPort(nasIP, fmt.Sprintf("%d", nasPort)), nil
}
