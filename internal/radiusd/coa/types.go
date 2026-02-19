// Package coa implements RADIUS Change of Authorization (CoA) and Packet of Disconnect (PoD)
// functionality as defined in RFC 3576.
//
// This package provides a robust client for sending CoA-Request and Disconnect-Request
// messages to NAS devices, with support for vendor-specific attributes.
//
// Key components:
//   - Client: Configurable CoA client with timeout and retry support
//   - DisconnectRequest: Parameters for disconnecting a user session
//   - CoARequest: Parameters for modifying an active session
//   - CoAResponse: Response from NAS after CoA operation
//
// Usage:
//
//	client := coa.NewClient(coa.Config{
//	    Timeout:    5 * time.Second,
//	    RetryCount: 3,
//	})
//	resp := client.SendDisconnect(ctx, coa.DisconnectRequest{
//	    NASIP:        "192.168.1.1",
//	    NASPort:      3799,
//	    Secret:       "sharedsecret",
//	    Username:     "user@example.com",
//	    AcctSessionID: "session123",
//	})
package coa

import (
	"time"

	"layeh.com/radius"
)

// Default configuration values for CoA client.
const (
	// DefaultTimeout is the default timeout for CoA operations.
	DefaultTimeout = 5 * time.Second

	// DefaultRetryCount is the default number of retry attempts.
	DefaultRetryCount = 2

	// DefaultRetryDelay is the default delay between retries.
	DefaultRetryDelay = 500 * time.Millisecond

	// DefaultCoAPort is the default UDP port for CoA operations (RFC 3576).
	DefaultCoAPort = 3799
)

// Config holds the configuration for a CoA client.
//
// Fields:
//   - Timeout: Maximum time to wait for a response (default: 5s)
//   - RetryCount: Number of retry attempts on failure (default: 2)
//   - RetryDelay: Initial delay between retries, uses exponential backoff (default: 500ms)
type Config struct {
	// Timeout is the maximum duration to wait for a NAS response.
	// After this duration, the operation is considered failed.
	Timeout time.Duration

	// RetryCount is the number of additional attempts after initial failure.
	// Set to 0 for no retries.
	RetryCount int

	// RetryDelay is the initial delay between retry attempts.
	// Subsequent retries use exponential backoff (delay * 2^attempt).
	RetryDelay time.Duration
}

// DisconnectRequest contains the parameters for disconnecting a user session.
//
// This request sends a Disconnect-Request (Code 40) to the NAS, which should
// terminate the specified session and return a Disconnect-ACK (Code 41) or
// Disconnect-NAK (Code 42).
//
// Required fields: NASIP, Secret, at least one session identifier (Username or AcctSessionID)
//
// References:
//   - RFC 3576 Section 3: Disconnect-Request
type DisconnectRequest struct {
	// NASIP is the IP address of the NAS device.
	// Required for routing the CoA packet.
	NASIP string

	// NASPort is the UDP port for CoA operations on the NAS.
	// Default: 3799 (RFC 3576). Some vendors use different ports:
	//   - Cisco: 1700 (default) or 3799
	//   - Mikrotik: 3799
	//   - Huawei: 3799
	NASPort int

	// Secret is the RADIUS shared secret for authenticating with the NAS.
	// Must match the secret configured on the NAS device.
	Secret string

	// Username is the RADIUS User-Name attribute for session identification.
	// Used to identify which session to disconnect.
	Username string

	// AcctSessionID is the RADIUS Acct-Session-Id attribute (RFC 2866).
	// This is the most reliable way to identify a specific session.
	AcctSessionID string

	// NASIdentifier is an optional identifier for the NAS.
	// If empty, NASIP will be used for NAS-IP-Address attribute.
	NASIdentifier string

	// VendorCode specifies the vendor for vendor-specific attributes.
	// Common values: "mikrotik", "cisco", "huawei", "" (for standard RADIUS only)
	VendorCode string

	// VendorAttributes contains vendor-specific attributes to include in the request.
	// Keys are attribute names, values can be string, int, or []byte.
	VendorAttributes map[string]interface{}

	// Reason is an optional description of why the disconnect was initiated.
	// Used for logging and auditing purposes.
	Reason string
}

// CoARequest contains the parameters for modifying an active session.
//
// This request sends a CoA-Request (Code 43) to the NAS, which should
// modify the specified session attributes and return a CoA-ACK (Code 44) or
// CoA-NAK (Code 45).
//
// Required fields: NASIP, Secret, at least one session identifier
//
// References:
//   - RFC 3576 Section 4: CoA-Request
type CoARequest struct {
	// NASIP is the IP address of the NAS device.
	NASIP string

	// NASPort is the UDP port for CoA operations on the NAS.
	NASPort int

	// Secret is the RADIUS shared secret for authenticating with the NAS.
	Secret string

	// Username is the RADIUS User-Name attribute for session identification.
	Username string

	// AcctSessionID is the RADIUS Acct-Session-Id attribute for session identification.
	AcctSessionID string

	// NASIdentifier is an optional identifier for the NAS.
	NASIdentifier string

	// VendorCode specifies the vendor for vendor-specific attributes.
	VendorCode string

	// SessionTimeout is the new session timeout value in seconds.
	// Set to 0 to keep existing value.
	SessionTimeout int

	// UpRate is the new upload rate limit in Kbps.
	// Set to 0 to keep existing value.
	UpRate int

	// DownRate is the new download rate limit in Kbps.
	// Set to 0 to keep existing value.
	DownRate int

	// DataQuota is the new data quota in MB.
	// Set to 0 to keep existing value.
	DataQuota int64

	// VendorAttributes contains additional vendor-specific attributes.
	VendorAttributes map[string]interface{}

	// Reason is an optional description of the modification reason.
	Reason string
}

// CoAResponse contains the result of a CoA operation.
//
// Fields:
//   - Success: Whether the operation was acknowledged by the NAS
//   - Code: The RADIUS response code (ACK or NAK)
//   - Error: Any error that occurred during the operation
//   - Duration: Time taken for the operation
//   - RetryCount: Number of retries attempted
type CoAResponse struct {
	// Success indicates whether the NAS acknowledged the request.
	// True for ACK responses, false for NAK or errors.
	Success bool

	// Code is the RADIUS response code received from the NAS.
	// Possible values:
	//   - radius.CodeDisconnectACK (41): Disconnect successful
	//   - radius.CodeDisconnectNAK (42): Disconnect rejected
	//   - radius.CodeCoAACK (44): CoA successful
	//   - radius.CodeCoANAK (45): CoA rejected
	Code radius.Code

	// Error contains any error that occurred during the operation.
	// This includes network errors, timeout errors, and packet parsing errors.
	// Nil if the operation completed successfully.
	Error error

	// Duration is the total time taken for the operation,
	// including any retries.
	Duration time.Duration

	// RetryCount is the number of retry attempts made.
	// 0 if the first attempt succeeded.
	RetryCount int

	// NASIP is the IP address of the NAS that was contacted.
	NASIP string

	// Message is an optional message from the NAS (Error-Cause attribute).
	Message string
}

// IsACK returns true if the response is an acknowledgment (success).
func (r *CoAResponse) IsACK() bool {
	return r.Success && (r.Code == radius.CodeDisconnectACK || r.Code == radius.CodeCoAACK)
}

// IsNAK returns true if the response is a negative acknowledgment.
func (r *CoAResponse) IsNAK() bool {
	return !r.Success && (r.Code == radius.CodeDisconnectNAK || r.Code == radius.CodeCoANAK)
}

// IsTimeout returns true if the operation timed out.
func (r *CoAResponse) IsTimeout() bool {
	return r.Error != nil && r.Error.Error() == "context deadline exceeded"
}

// VendorAttributeBuilder defines the interface for vendor-specific CoA attribute construction.
//
// Implementations of this interface handle the vendor-specific formatting of
// rate limits, session timeouts, and other attributes for CoA requests.
//
// Each vendor (Mikrotik, Cisco, Huawei) has different attribute formats and naming
// conventions. This interface abstracts those differences.
type VendorAttributeBuilder interface {
	// VendorCode returns the vendor identifier string (e.g., "mikrotik", "cisco").
	VendorCode() string

	// AddRateLimit adds bandwidth rate limit attributes to the packet.
	//
	// Parameters:
	//   - pkt: The RADIUS packet to add attributes to
	//   - upRate: Upload rate in Kbps (0 = no limit)
	//   - downRate: Download rate in Kbps (0 = no limit)
	//
	// Returns an error if attribute encoding fails.
	AddRateLimit(pkt *radius.Packet, upRate, downRate int) error

	// AddSessionTimeout adds session timeout attribute to the packet.
	//
	// Parameters:
	//   - pkt: The RADIUS packet to add the attribute to
	//   - timeout: Session timeout in seconds
	//
	// Returns an error if attribute encoding fails.
	AddSessionTimeout(pkt *radius.Packet, timeout int) error

	// AddDataQuota adds data quota attributes to the packet.
	//
	// Some vendors support data quota limits via vendor-specific attributes.
	//
	// Parameters:
	//   - pkt: The RADIUS packet to add attributes to
	//   - quotaMB: Data quota in megabytes
	//
	// Returns an error if attribute encoding fails or vendor doesn't support quota.
	AddDataQuota(pkt *radius.Packet, quotaMB int64) error

	// AddDisconnectAttributes adds vendor-specific attributes for disconnect requests.
	//
	// Some vendors require additional attributes in disconnect requests.
	//
	// Parameters:
	//   - pkt: The RADIUS packet to add attributes to
	//   - username: The username of the session to disconnect
	//   - acctSessionID: The accounting session ID
	//
	// Returns an error if attribute encoding fails.
	AddDisconnectAttributes(pkt *radius.Packet, username, acctSessionID string) error
}

// Error definitions for CoA operations.
var (
	// ErrMissingNASIP indicates that NASIP is required but not provided.
	ErrMissingNASIP = &CoAError{Code: "MISSING_NAS_IP", Message: "NAS IP address is required"}

	// ErrMissingSecret indicates that Secret is required but not provided.
	ErrMissingSecret = &CoAError{Code: "MISSING_SECRET", Message: "RADIUS secret is required"}

	// ErrMissingSessionID indicates that no session identifier was provided.
	ErrMissingSessionID = &CoAError{Code: "MISSING_SESSION_ID", Message: "Username or AcctSessionID is required"}

	// ErrInvalidNASIP indicates that the NAS IP address is invalid.
	ErrInvalidNASIP = &CoAError{Code: "INVALID_NAS_IP", Message: "Invalid NAS IP address"}

	// ErrUnknownVendor indicates that the vendor code is not recognized.
	ErrUnknownVendor = &CoAError{Code: "UNKNOWN_VENDOR", Message: "Unknown vendor code"}

	// ErrTimeout indicates that the operation timed out.
	ErrTimeout = &CoAError{Code: "TIMEOUT", Message: "Operation timed out"}

	// ErrNAKReceived indicates that the NAS rejected the request.
	ErrNAKReceived = &CoAError{Code: "NAK_RECEIVED", Message: "NAS rejected the request"}
)

// CoAError represents a CoA operation error with a code and message.
type CoAError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *CoAError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap returns the underlying cause of the error.
func (e *CoAError) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is for comparison.
func (e *CoAError) Is(target error) bool {
	t, ok := target.(*CoAError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}
