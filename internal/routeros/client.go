// Package routeros provides a client for interacting with MikroTik RouterOS devices
// via the RouterOS API protocol.
//
// This package enables auto-discovery of MikroTik devices on the network by connecting
// to the RouterOS API (TCP 8728 for unencrypted, TCP 8729 for SSL/TLS).
//
// The client supports:
// - Connecting to RouterOS devices
// - Authentication (login)
// - Querying system information (identity, resource, version)
// - Detecting RouterOS devices on the network
//
// Example usage:
//
//	client := routeros.NewClient(routeros.Config{
//	    Address:  "192.168.1.1",
//	    Username: "admin",
//	    Password: "password",
//	    UseTLS:   false,
//	})
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	if err := client.Login(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	info, err := client.GetSystemInfo(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Device: %s, Version: %s\n", info.BoardName, info.Version)
package routeros

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	// DefaultPort is the default unencrypted RouterOS API port.
	DefaultPort = "8728"
	// DefaultTLSPort is the default TLS-encrypted RouterOS API port.
	DefaultTLSPort = "8729"

	// Response codes from RouterOS API
	respDone = 0x14 // !do - command completed successfully
	respRe  = 0x11 // !re - reply data
	respTrap = 0x17 // !tr - command failed
	respFatal = 0x1a // !fatal - fatal error
)

// ResponseType represents the type of RouterOS API response.
type ResponseType int

const (
	ResponseUnknown ResponseType = iota
	ResponseDone
	ResponseRe
	ResponseTrap
	ResponseFatal
)

// Config holds configuration for the RouterOS client.
type Config struct {
	Address  string        // RouterOS device address (IP or hostname)
	Port     string        // API port (default: 8728)
	Username string        // Login username
	Password string        // Login password
	UseTLS   bool         // Use TLS/SSL connection
	Timeout  time.Duration // Connection timeout
}

// SystemInfo holds information retrieved from a RouterOS device.
type SystemInfo struct {
	Identity  string // System identity (from /system/identity)
	BoardName string // Board name (from /system/resource)
	Version   string // RouterOS version
	Model     string // Device model
	Serial    string // Serial number
}

// Client is a RouterOS API client for connecting to MikroTik devices.
type Client struct {
	config  Config
	conn    net.Conn
	tag     uint32
	loggedIn bool
}

// NewClient creates a new RouterOS client with the given configuration.
//
// If Port is not specified, default ports are used:
//   - 8728 for unencrypted connections
//   - 8729 for TLS connections
//
// If Timeout is not specified, defaults to 10 seconds.
func NewClient(config Config) *Client {
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Port == "" {
		if config.UseTLS {
			config.Port = DefaultTLSPort
		} else {
			config.Port = DefaultPort
		}
	}
	// If address doesn't include port, append default port
	if !strings.Contains(config.Address, ":") {
		config.Address = net.JoinHostPort(config.Address, config.Port)
	}
	return &Client{
		config: config,
		tag:    0,
	}
}

// Connect establishes a TCP connection to the RouterOS device.
//
// The context is used for cancellation and timeout control.
// Callers should always call Close() after Connect(), even if Connect() returns an error.
func (c *Client) Connect(ctx context.Context) error {
	addr := c.config.Address

	var conn net.Conn
	var err error

	if c.config.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		conn, err = tls.Dial("tcp4", addr, tlsConfig)
	} else {
		conn, err = net.DialTimeout("tcp4", addr, c.config.Timeout)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to RouterOS at %s: %w", addr, err)
	}

	conn.SetReadDeadline(time.Now().Add(c.config.Timeout))

	c.conn = conn
	return nil
}

// Close closes the connection to the RouterOS device.
func (c *Client) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.loggedIn = false
		return err
	}
	return nil
}

// Login authenticates to the RouterOS device using the configured credentials.
//
// RouterOS API uses a challenge-response authentication mechanism:
// 1. Send /login with username only
// 2. Receive challenge (a hash value) in a !re response
// 3. Send /login with username, password, and response=MD5(password+challenge)
//
// Connect() must be called before Login().
// After successful login, the connection remains authenticated until Close() is called.
func (c *Client) Login(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("not connected: call Connect() first")
	}

	// Step 1: Send initial login request with username ONLY
	loginReq1 := c.buildLoginRequestSimple()
	fmt.Printf("[RouterOS] Login step 1 to %s: % x\n", c.config.Address, loginReq1)
	
	if _, err := c.conn.Write(loginReq1); err != nil {
		return fmt.Errorf("failed to send login: %w", err)
	}

	// Read response
	resp, err := c.readResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	fmt.Printf("[RouterOS] Login response from %s: type=%v, data=%s\n", c.config.Address, resp.Type, string(resp.Data))

	// If ResponseDone, login successful without challenge (old RouterOS versions)
	if resp.Type == ResponseDone {
		c.loggedIn = true
		return nil
	}

	// If ResponseTrap, credentials invalid
	if resp.Type == ResponseTrap {
		return errors.New("login failed: invalid credentials")
	}

	// If ResponseRe, it contains the challenge
	if resp.Type == ResponseRe {
		// Extract challenge from "ret=<hash>"
		challenge := extractValue(resp.Data, "ret")
		if challenge == "" {
			// Also check for "=ret" format
			challenge = extractValue(resp.Data, "=ret")
		}
		if challenge == "" {
			return errors.New("login failed: no challenge received")
		}

		fmt.Printf("[RouterOS] Got challenge: %s\n", challenge)

		// Step 2: Compute response and send
		loginReq2 := c.buildChallengeResponse(challenge)
		fmt.Printf("[RouterOS] Login step 2 to %s: % x\n", c.config.Address, loginReq2)

		if _, err := c.conn.Write(loginReq2); err != nil {
			return fmt.Errorf("failed to send challenge response: %w", err)
		}

		resp2, err := c.readResponse(ctx)
		if err != nil {
			return fmt.Errorf("failed to read challenge response: %w", err)
		}

		fmt.Printf("[RouterOS] Challenge result: type=%v, data=%s\n", c.config.Address, resp2.Type, string(resp2.Data))

		if resp2.Type == ResponseDone {
			c.loggedIn = true
			return nil
		}

		if resp2.Type == ResponseTrap {
			return errors.New("login failed: invalid credentials")
		}

		if resp2.Type == ResponseFatal {
			return errors.New("login failed: fatal error from RouterOS")
		}
	}

	return errors.New("login failed: unknown error")
}

// buildLoginRequestSimple sends only the username to get the challenge.
func (c *Client) buildLoginRequestSimple() []byte {
	var data []byte
	data = appendWord(data, "name="+c.config.Username)

	command := "/login"
	commandLen := len(command) + 1
	totalLen := commandLen + len(data)
	packet := make([]byte, 4+totalLen)
	binary.BigEndian.PutUint32(packet[0:4], uint32(commandLen))
	copy(packet[4:], command)
	packet[4+len(command)] = 0
	copy(packet[4+commandLen:], data)
	return packet
}

// buildChallengeResponse builds the second login step with the challenge response.
func (c *Client) buildChallengeResponse(challenge string) []byte {
	// RouterOS uses MD5 of: username + " " + password + challenge
	response := computeResponse(c.config.Username, c.config.Password, challenge)

	var data []byte
	data = appendWord(data, "name="+c.config.Username)
	data = appendWord(data, "response="+response)

	command := "/login"
	commandLen := len(command) + 1

	totalLen := commandLen + len(data)
	packet := make([]byte, 4+totalLen)

	binary.BigEndian.PutUint32(packet[0:4], uint32(commandLen))

	copy(packet[4:], command)
	packet[4+len(command)] = 0

	copy(packet[4+commandLen:], data)

	return packet
}

// computeResponse computes the MD5 hash for RouterOS challenge-response login.
func computeResponse(username, password, challenge string) string {
	input := username + " " + password + challenge
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

// GetSystemInfo retrieves system information from the RouterOS device.
//
// This method queries multiple API paths:
//   - /system/identity (device name)
//   - /system/resource (board name, version, serial)
//   - /system/package (to determine if it's RouterOS)
//
// Login() must be called successfully before calling this method.
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	if !c.loggedIn {
		return nil, errors.New("not logged in: call Login() first")
	}

	info := &SystemInfo{}

	// Get identity
	identity, err := c.runCommand(ctx, "/system/identity/get")
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	info.Identity = extractValue(identity, "name")

	// Get resource info
	resource, err := c.runCommand(ctx, "/system/resource/get")
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}
	info.BoardName = extractValue(resource, "board-name")
	info.Version = extractValue(resource, "version")
	info.Serial = extractValue(resource, "serial-number")

	// Determine model from board name
	info.Model = info.BoardName
	if info.BoardName != "" {
		// Common MikroTik models
		if strings.HasPrefix(info.BoardName, "RB") || strings.HasPrefix(info.BoardName, "hAP") ||
			strings.HasPrefix(info.BoardName, "CCR") || strings.HasPrefix(info.BoardName, "CRS") ||
			strings.HasPrefix(info.BoardName, "CSS") || strings.HasPrefix(info.BoardName, "cAP") {
			info.Model = "Mikrotik " + info.BoardName
		}
	}

	return info, nil
}

// IsRouterOS attempts to detect if the target is a RouterOS device.
//
// This method tries to connect and send a login request.
// If the target responds with RouterOS API protocol responses, it returns true.
func (c *Client) IsRouterOS(ctx context.Context) (bool, error) {
	if err := c.Connect(ctx); err != nil {
		return false, err
	}
	defer c.Close()

	// Send a simple /system/identity/get command to test if it's RouterOS
	cmd := c.BuildCommand("/system/identity/get")
	
	fmt.Printf("[RouterOS] Sending command to %s: % x\n", c.config.Address, cmd)

	n, err := c.conn.Write(cmd)
	if err != nil {
		return false, fmt.Errorf("write command failed: %w", err)
	}

	c.conn.SetReadDeadline(time.Now().Add(c.config.Timeout))

	buf := make([]byte, 2048)
	n, err = c.conn.Read(buf)
	if err != nil {
		return false, fmt.Errorf("read response failed: %w", err)
	}

	fmt.Printf("[RouterOS] Received response from %s: % x (len=%d)\n", c.config.Address, buf[:n], n)

	if n > 0 {
		respType := buf[0]
		if respType == respDone || respType == respRe || respType == respTrap || respType == respFatal {
			return true, nil
		}
	}

	return false, nil
}

// buildLoginRequest builds a RouterOS API login request packet.
func (c *Client) buildLoginRequest() []byte {
	var data []byte

	// Add username
	data = appendWord(data, "name="+c.config.Username)

	// Add password (for plain text auth)
	data = appendWord(data, "password="+c.config.Password)

	// Build the /login command with the data
	command := "/login"
	commandLen := len(command) + 1 // +1 for null

	// Total packet: 4 byte length + command + null + data
	totalLen := commandLen + len(data)
	packet := make([]byte, 4+totalLen)

	// Write length of first word (command)
	binary.BigEndian.PutUint32(packet[0:4], uint32(commandLen))

	// Write command
	copy(packet[4:], command)
	packet[4+len(command)] = 0

	// Write data
	copy(packet[4+commandLen:], data)

	return packet
}

// BuildCommand builds a RouterOS API command packet for the given path.
//
// Example:
//
//	cmd := client.BuildCommand("/system/identity/get")
func (c *Client) BuildCommand(path string) []byte {
	return c.buildCommand(path, nil)
}

// buildCommand builds a RouterOS API command with optional parameters.
func (c *Client) buildCommand(path string, params map[string]string) []byte {
	c.tag++

	// Build command data
	var data []byte

	// Add command path
	data = appendWord(data, path)

	// Add parameters
	for k, v := range params {
		data = appendWord(data, k+"="+v)
	}

	// Add .tag attribute if we have a tag
	if c.tag > 0 {
		data = appendWord(data, fmt.Sprintf(".tag=%d", c.tag))
	}

	// Build packet
	length := len(data)
	packet := make([]byte, 4+length)
	binary.BigEndian.PutUint32(packet[0:4], uint32(length))
	copy(packet[4:], data)

	return packet
}

// runCommand executes a RouterOS API command and returns the response.
func (c *Client) runCommand(ctx context.Context, path string) ([]byte, error) {
	cmd := c.buildCommand(path, nil)

	if _, err := c.conn.Write(cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Collect all responses
	var results []byte

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Set deadline for each read
		c.conn.SetReadDeadline(time.Now().Add(c.config.Timeout))

		resp, err := c.readResponse(ctx)
		if err != nil {
			if errors.Is(err, errors.New("no more responses")) {
				break
			}
			return nil, err
		}

		if resp.Type == ResponseDone {
			break
		}

		if resp.Type == ResponseRe {
			results = append(results, resp.Data...)
		}

		if resp.Type == ResponseTrap || resp.Type == ResponseFatal {
			return nil, fmt.Errorf("command failed: %s", string(resp.Data))
		}
	}

	return results, nil
}

// response holds a parsed RouterOS API response.
type response struct {
	Type ResponseType
	Data []byte
}

// readResponse reads and parses a single RouterOS API response.
func (c *Client) readResponse(ctx context.Context) (*response, error) {
	// Read 4-byte length header
	header := make([]byte, 4)
	_, err := c.conn.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read response header: %w", err)
	}

	length := binary.BigEndian.Uint32(header)
	if length == 0 {
		return &response{Type: ResponseDone}, nil
	}

	// Read response data
	data := make([]byte, length)
	_, err = c.conn.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read response data: %w", err)
	}

	resp := &response{
		Type: parseResponseType(data),
		Data: data,
	}

	return resp, nil
}

// parseResponseType determines the response type from the first byte.
func parseResponseType(data []byte) ResponseType {
	if len(data) == 0 {
		return ResponseUnknown
	}

	switch data[0] {
	case respDone:
		return ResponseDone
	case respRe:
		return ResponseRe
	case respTrap:
		return ResponseTrap
	case respFatal:
		return ResponseFatal
	default:
		return ResponseUnknown
	}
}

// parseResponse parses a raw RouterOS API response.
func parseResponse(data []byte) *response {
	return &response{
		Type: parseResponseType(data),
		Data: data,
	}
}

// appendWord appends a word (null-terminated string) to the data.
// RouterOS API uses word format: [length (4 bytes)][string][null]
func appendWord(data []byte, word string) []byte {
	wordLen := len(word) + 1 // +1 for null terminator
	length := 4 + wordLen

	newData := make([]byte, len(data)+length)
	copy(newData, data)

	// Write length
	binary.BigEndian.PutUint32(newData[len(data):len(data)+4], uint32(wordLen))

	// Write string and null terminator
	copy(newData[len(data)+4:], word)
	newData[len(data)+4+len(word)] = 0

	return newData
}

// RunCommand executes a RouterOS API command and returns parsed response.
//
// This method is more flexible than runCommand - it parses the response
// into a slice of maps, where each map represents a row of data.
// The command path uses spaces instead of slashes (e.g., "ip neighbor print")
func (c *Client) RunCommand(ctx context.Context, path string) ([]map[string]string, error) {
	if !c.loggedIn {
		return nil, errors.New("not logged in: call Login() first")
	}

	data, err := c.runCommand(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to run command %s: %w", path, err)
	}

	// Parse response data into a list of maps
	// Response format: key1=value1\0key2=value2\0... (repeated for each item)
	return parseResponseList(data), nil
}

// parseResponseList parses RouterOS API response data into a list of maps.
func parseResponseList(data []byte) []map[string]string {
	if len(data) == 0 {
		return nil
	}

	var results []map[string]string

	// Split by looking for the pattern of repeated keys
	// Each entry starts after a null byte followed by a word
	entries := splitEntries(data)

	for _, entry := range entries {
		if len(entry) == 0 {
			continue
		}

		row := make(map[string]string)
		pairs := strings.Split(string(entry), "\x00")

		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			// Find the first = to split key/value
			eqIdx := strings.Index(pair, "=")
			if eqIdx > 0 {
				key := pair[:eqIdx]
				value := pair[eqIdx+1:]
				row[key] = value
			}
		}

		if len(row) > 0 {
			results = append(results, row)
		}
	}

	return results
}

// splitEntries splits the response data into individual entries.
func splitEntries(data []byte) [][]byte {
	var entries [][]byte

	// Look for the pattern where a word appears at the start after null bytes
	// This is a simplified approach - RouterOS responses have repeated structure
	current := make([]byte, 0)

	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			if len(current) > 0 {
				endIdx := i + 20
				if endIdx > len(data) {
					endIdx = len(data)
				}
				lookAhead := string(data[i:endIdx])
				if !strings.Contains(lookAhead, "=") || strings.HasPrefix(lookAhead, ".") {
					entries = append(entries, current)
					current = make([]byte, 0)
					continue
				}
			}
		}
		current = append(current, data[i])
	}

	// Don't forget the last entry
	if len(current) > 0 {
		entries = append(entries, current)
	}

	return entries
}

// extractValue extracts a value from RouterOS API response data.
//
// Response data format: key=value\0key=value\0
func extractValue(data []byte, key string) string {
	if len(data) == 0 {
		return ""
	}

	search := key + "="
	idx := strings.Index(string(data), search)
	if idx == -1 {
		// Try without trailing =
		search = key + "="
		idx = strings.Index(string(data), search)
		if idx == -1 {
			return ""
		}
	}

	// Find value start
	valueStart := idx + len(search)

	// Find value end (null terminator or end of data)
	valueEnd := len(data)
	for i := valueStart; i < len(data); i++ {
		if data[i] == 0 {
			valueEnd = i
			break
		}
	}

	return string(data[valueStart:valueEnd])
}
