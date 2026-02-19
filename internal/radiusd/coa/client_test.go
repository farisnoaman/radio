package coa

import (
	"context"
	"net"
	"testing"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// TestNewClient tests the creation of a new CoA client with various configurations.
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected Config
	}{
		{
			name:   "default configuration",
			config: Config{},
			expected: Config{
				Timeout:    DefaultTimeout,
				RetryCount: DefaultRetryCount,
				RetryDelay: DefaultRetryDelay,
			},
		},
		{
			name: "custom configuration",
			config: Config{
				Timeout:    10 * time.Second,
				RetryCount: 5,
				RetryDelay: 1 * time.Second,
			},
			expected: Config{
				Timeout:    10 * time.Second,
				RetryCount: 5,
				RetryDelay: 1 * time.Second,
			},
		},
		{
			name: "zero timeout uses default",
			config: Config{
				Timeout:    0,
				RetryCount: 3,
				RetryDelay: 200 * time.Millisecond,
			},
			expected: Config{
				Timeout:    DefaultTimeout,
				RetryCount: 3,
				RetryDelay: 200 * time.Millisecond,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			if client == nil {
				t.Fatal("expected non-nil client")
			}
			if client.config.Timeout != tt.expected.Timeout {
				t.Errorf("timeout: got %v, want %v", client.config.Timeout, tt.expected.Timeout)
			}
			if client.config.RetryCount != tt.expected.RetryCount {
				t.Errorf("retry count: got %v, want %v", client.config.RetryCount, tt.expected.RetryCount)
			}
			if client.config.RetryDelay != tt.expected.RetryDelay {
				t.Errorf("retry delay: got %v, want %v", client.config.RetryDelay, tt.expected.RetryDelay)
			}
		})
	}
}

// TestDisconnectRequest_Validate tests validation of disconnect requests.
func TestDisconnectRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     DisconnectRequest
		wantErr error
	}{
		{
			name: "valid request with username",
			req: DisconnectRequest{
				NASIP:    "192.168.1.1",
				NASPort:  3799,
				Secret:   "secret",
				Username: "testuser",
			},
			wantErr: nil,
		},
		{
			name: "valid request with session id",
			req: DisconnectRequest{
				NASIP:          "192.168.1.1",
				NASPort:        3799,
				Secret:         "secret",
				AcctSessionID:  "session123",
			},
			wantErr: nil,
		},
		{
			name: "missing NAS IP",
			req: DisconnectRequest{
				Secret:    "secret",
				Username:  "testuser",
			},
			wantErr: ErrMissingNASIP,
		},
		{
			name: "missing secret",
			req: DisconnectRequest{
				NASIP:    "192.168.1.1",
				Username: "testuser",
			},
			wantErr: ErrMissingSecret,
		},
		{
			name: "missing session identifier",
			req: DisconnectRequest{
				NASIP:  "192.168.1.1",
				Secret: "secret",
			},
			wantErr: ErrMissingSessionID,
		},
		{
			name: "invalid NAS IP",
			req: DisconnectRequest{
				NASIP:    "invalid-ip",
				Secret:   "secret",
				Username: "testuser",
			},
			wantErr: ErrInvalidNASIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("error: got %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCoARequest_Validate tests validation of CoA requests.
func TestCoARequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CoARequest
		wantErr error
	}{
		{
			name: "valid request",
			req: CoARequest{
				NASIP:         "192.168.1.1",
				NASPort:       3799,
				Secret:        "secret",
				Username:      "testuser",
				SessionTimeout: 3600,
				UpRate:        10240,
				DownRate:      20480,
			},
			wantErr: nil,
		},
		{
			name: "missing NAS IP",
			req: CoARequest{
				Secret:   "secret",
				Username: "testuser",
			},
			wantErr: ErrMissingNASIP,
		},
		{
			name: "missing secret",
			req: CoARequest{
				NASIP:    "192.168.1.1",
				Username: "testuser",
			},
			wantErr: ErrMissingSecret,
		},
		{
			name: "missing session identifier",
			req: CoARequest{
				NASIP:  "192.168.1.1",
				Secret: "secret",
			},
			wantErr: ErrMissingSessionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("error: got %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestBuildDisconnectPacket tests the construction of disconnect packets.
func TestBuildDisconnectPacket(t *testing.T) {
	tests := []struct {
		name         string
		req          DisconnectRequest
		wantUsername string
		wantSessionID string
	}{
		{
			name: "with username and session id",
			req: DisconnectRequest{
				NASIP:         "192.168.1.1",
				Secret:        "secret",
				Username:      "testuser@example.com",
				AcctSessionID: "session123",
			},
			wantUsername:  "testuser@example.com",
			wantSessionID: "session123",
		},
		{
			name: "with username only",
			req: DisconnectRequest{
				NASIP:    "192.168.1.1",
				Secret:   "secret",
				Username: "testuser",
			},
			wantUsername:  "testuser",
			wantSessionID: "",
		},
		{
			name: "with session id only",
			req: DisconnectRequest{
				NASIP:         "192.168.1.1",
				Secret:        "secret",
				AcctSessionID: "session456",
			},
			wantUsername:  "",
			wantSessionID: "session456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt := buildDisconnectPacket(&tt.req)

			if pkt.Code != radius.CodeDisconnectRequest {
				t.Errorf("packet code: got %v, want %v", pkt.Code, radius.CodeDisconnectRequest)
			}

			username := rfc2865.UserName_GetString(pkt)
			if username != tt.wantUsername {
				t.Errorf("username: got %q, want %q", username, tt.wantUsername)
			}

			sessionID := rfc2866.AcctSessionID_GetString(pkt)
			if sessionID != tt.wantSessionID {
				t.Errorf("session id: got %q, want %q", sessionID, tt.wantSessionID)
			}
		})
	}
}

// TestBuildCoAPacket tests the construction of CoA packets.
func TestBuildCoAPacket(t *testing.T) {
	tests := []struct {
		name              string
		req               CoARequest
		wantUsername      string
		wantSessionID     string
		wantSessionTimeout int
	}{
		{
			name: "with all attributes",
			req: CoARequest{
				NASIP:          "192.168.1.1",
				Secret:         "secret",
				Username:       "testuser",
				AcctSessionID:  "session123",
				SessionTimeout: 3600,
				UpRate:         10240,
				DownRate:       20480,
			},
			wantUsername:      "testuser",
			wantSessionID:     "session123",
			wantSessionTimeout: 3600,
		},
		{
			name: "with username only",
			req: CoARequest{
				NASIP:    "192.168.1.1",
				Secret:   "secret",
				Username: "testuser",
			},
			wantUsername:  "testuser",
			wantSessionID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt := buildCoAPacket(&tt.req)

			if pkt.Code != radius.CodeCoARequest {
				t.Errorf("packet code: got %v, want %v", pkt.Code, radius.CodeCoARequest)
			}

			username := rfc2865.UserName_GetString(pkt)
			if username != tt.wantUsername {
				t.Errorf("username: got %q, want %q", username, tt.wantUsername)
			}

			sessionID := rfc2866.AcctSessionID_GetString(pkt)
			if sessionID != tt.wantSessionID {
				t.Errorf("session id: got %q, want %q", sessionID, tt.wantSessionID)
			}
		})
	}
}

// mockCoAServer creates a mock RADIUS server for testing CoA operations.
type mockCoAServer struct {
	listener   *net.UDPConn
	secret     string
	response   radius.Code
	closeChan  chan struct{}
	errorChan  chan error
}

// newMockCoAServer creates and starts a mock CoA server.
func newMockCoAServer(secret string, response radius.Code) (*mockCoAServer, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	server := &mockCoAServer{
		listener:  listener,
		secret:    secret,
		response:  response,
		closeChan: make(chan struct{}),
		errorChan: make(chan error, 1),
	}

	go server.serve()

	return server, nil
}

func (s *mockCoAServer) serve() {
	buf := make([]byte, 4096)
	for {
		select {
		case <-s.closeChan:
			return
		default:
			s.listener.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, remoteAddr, err := s.listener.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				select {
				case s.errorChan <- err:
				default:
				}
				return
			}

			// Parse the incoming packet
			pkt := buf[:n]
			radiusPkt, err := radius.Parse(pkt, []byte(s.secret))
			if err != nil {
				select {
				case s.errorChan <- err:
				default:
				}
				continue
			}

			// Create response packet
			resp := radius.New(s.response, []byte(s.secret))
			resp.Identifier = radiusPkt.Identifier

			// Copy authenticator for response calculation
			copy(resp.Authenticator[:], radiusPkt.Authenticator[:])

			// Send response
			respBytes, err := resp.Encode()
			if err != nil {
				select {
				case s.errorChan <- err:
				default:
				}
				continue
			}

			s.listener.WriteToUDP(respBytes, remoteAddr)
		}
	}
}

func (s *mockCoAServer) addr() *net.UDPAddr {
	return s.listener.LocalAddr().(*net.UDPAddr)
}

func (s *mockCoAServer) close() {
	close(s.closeChan)
	s.listener.Close()
}

// TestSendDisconnect_Integration tests the SendDisconnect method with a mock server.
func TestSendDisconnect_Integration(t *testing.T) {
	secret := "testsecret"

	tests := []struct {
		name     string
		response radius.Code
		wantSuccess bool
		wantCode   radius.Code
	}{
		{
			name:        "successful disconnect ACK",
			response:    radius.CodeDisconnectACK,
			wantSuccess: true,
			wantCode:    radius.CodeDisconnectACK,
		},
		{
			name:        "disconnect NAK",
			response:    radius.CodeDisconnectNAK,
			wantSuccess: false,
			wantCode:    radius.CodeDisconnectNAK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := newMockCoAServer(secret, tt.response)
			if err != nil {
				t.Fatalf("failed to create mock server: %v", err)
			}
			defer server.close()

			addr := server.addr()
			client := NewClient(Config{
				Timeout:    2 * time.Second,
				RetryCount: 0,
			})

			req := DisconnectRequest{
				NASIP:         addr.IP.String(),
				NASPort:       addr.Port,
				Secret:        secret,
				Username:      "testuser",
				AcctSessionID: "session123",
			}

			ctx := context.Background()
			resp := client.SendDisconnect(ctx, req)

			if resp.Success != tt.wantSuccess {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSuccess)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("code: got %v, want %v", resp.Code, tt.wantCode)
			}

			if resp.Error != nil && tt.wantSuccess {
				t.Errorf("unexpected error: %v", resp.Error)
			}
		})
	}
}

// TestSendCoA_Integration tests the SendCoA method with a mock server.
func TestSendCoA_Integration(t *testing.T) {
	secret := "testsecret"

	tests := []struct {
		name       string
		response   radius.Code
		wantSuccess bool
		wantCode   radius.Code
	}{
		{
			name:        "successful CoA ACK",
			response:    radius.CodeCoAACK,
			wantSuccess: true,
			wantCode:    radius.CodeCoAACK,
		},
		{
			name:        "CoA NAK",
			response:    radius.CodeCoANAK,
			wantSuccess: false,
			wantCode:    radius.CodeCoANAK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := newMockCoAServer(secret, tt.response)
			if err != nil {
				t.Fatalf("failed to create mock server: %v", err)
			}
			defer server.close()

			addr := server.addr()
			client := NewClient(Config{
				Timeout:    2 * time.Second,
				RetryCount: 0,
			})

			req := CoARequest{
				NASIP:          addr.IP.String(),
				NASPort:        addr.Port,
				Secret:         secret,
				Username:       "testuser",
				AcctSessionID:  "session123",
				SessionTimeout: 3600,
				UpRate:         10240,
				DownRate:       20480,
			}

			ctx := context.Background()
			resp := client.SendCoA(ctx, req)

			if resp.Success != tt.wantSuccess {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSuccess)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("code: got %v, want %v", resp.Code, tt.wantCode)
			}
		})
	}
}

// TestSendDisconnect_Timeout tests timeout handling.
func TestSendDisconnect_Timeout(t *testing.T) {
	// Create a server that doesn't respond
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	serverAddr := listener.LocalAddr().(*net.UDPAddr)

	// Don't read from the listener, simulating a non-responsive server

	client := NewClient(Config{
		Timeout:    500 * time.Millisecond,
		RetryCount: 0,
	})

	req := DisconnectRequest{
		NASIP:         serverAddr.IP.String(),
		NASPort:       serverAddr.Port,
		Secret:        "secret",
		Username:      "testuser",
		AcctSessionID: "session123",
	}

	ctx := context.Background()
	resp := client.SendDisconnect(ctx, req)

	if resp.Success {
		t.Error("expected failure for timeout")
	}

	if resp.Error == nil {
		t.Error("expected timeout error")
	}

	// Duration should be at least close to timeout
	// Allow some tolerance for fast failures
	if resp.Duration > 1*time.Second {
		t.Errorf("expected duration <= 1s for fast failure, got %v", resp.Duration)
	}
}

// TestSendDisconnect_Validation tests that validation errors are returned properly.
func TestSendDisconnect_Validation(t *testing.T) {
	client := NewClient(Config{})

	tests := []struct {
		name    string
		req     DisconnectRequest
		wantErr error
	}{
		{
			name: "missing NAS IP",
			req: DisconnectRequest{
				Secret:   "secret",
				Username: "testuser",
			},
			wantErr: ErrMissingNASIP,
		},
		{
			name: "missing secret",
			req: DisconnectRequest{
				NASIP:    "192.168.1.1",
				Username: "testuser",
			},
			wantErr: ErrMissingSecret,
		},
		{
			name: "missing session identifier",
			req: DisconnectRequest{
				NASIP:  "192.168.1.1",
				Secret: "secret",
			},
			wantErr: ErrMissingSessionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := client.SendDisconnect(context.Background(), tt.req)
			if resp.Success {
				t.Error("expected failure")
			}
			if resp.Error == nil {
				t.Error("expected error")
			}
		})
	}
}

// TestSendDisconnect_Retry tests retry logic.
func TestSendDisconnect_Retry(t *testing.T) {
	// Create a server that doesn't respond
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	serverAddr := listener.LocalAddr().(*net.UDPAddr)

	client := NewClient(Config{
		Timeout:    200 * time.Millisecond,
		RetryCount: 2,
		RetryDelay: 100 * time.Millisecond,
	})

	req := DisconnectRequest{
		NASIP:         serverAddr.IP.String(),
		NASPort:       serverAddr.Port,
		Secret:        "secret",
		Username:      "testuser",
		AcctSessionID: "session123",
	}

	start := time.Now()
	resp := client.SendDisconnect(context.Background(), req)
	elapsed := time.Since(start)

	if resp.Success {
		t.Error("expected failure")
	}

	// With 2 retries, we should have at least 3 attempts (initial + 2 retries)
	// Each attempt takes ~200ms timeout
	// Total should be at least 600ms
	if elapsed < 600*time.Millisecond {
		t.Errorf("expected at least 600ms with retries, got %v", elapsed)
	}

	if resp.RetryCount != 2 {
		t.Errorf("expected 2 retries, got %d", resp.RetryCount)
	}
}

// TestCoAResponse_Methods tests CoAResponse helper methods.
func TestCoAResponse_Methods(t *testing.T) {
	tests := []struct {
		name      string
		resp      CoAResponse
		isACK     bool
		isNAK     bool
		isTimeout bool
	}{
		{
			name: "disconnect ACK",
			resp: CoAResponse{
				Success: true,
				Code:    radius.CodeDisconnectACK,
			},
			isACK:     true,
			isNAK:     false,
			isTimeout: false,
		},
		{
			name: "disconnect NAK",
			resp: CoAResponse{
				Success: false,
				Code:    radius.CodeDisconnectNAK,
			},
			isACK:     false,
			isNAK:     true,
			isTimeout: false,
		},
		{
			name: "CoA ACK",
			resp: CoAResponse{
				Success: true,
				Code:    radius.CodeCoAACK,
			},
			isACK:     true,
			isNAK:     false,
			isTimeout: false,
		},
		{
			name: "CoA NAK",
			resp: CoAResponse{
				Success: false,
				Code:    radius.CodeCoANAK,
			},
			isACK:     false,
			isNAK:     true,
			isTimeout: false,
		},
		{
			name: "timeout error",
			resp: CoAResponse{
				Success: false,
				Error:   context.DeadlineExceeded,
			},
			isACK:     false,
			isNAK:     false,
			isTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.IsACK() != tt.isACK {
				t.Errorf("IsACK: got %v, want %v", tt.resp.IsACK(), tt.isACK)
			}
			if tt.resp.IsNAK() != tt.isNAK {
				t.Errorf("IsNAK: got %v, want %v", tt.resp.IsNAK(), tt.isNAK)
			}
			if tt.resp.IsTimeout() != tt.isTimeout {
				t.Errorf("IsTimeout: got %v, want %v", tt.resp.IsTimeout(), tt.isTimeout)
			}
		})
	}
}

// TestGetNASAddress tests the getNASAddress helper function.
func TestGetNASAddress(t *testing.T) {
	tests := []struct {
		name     string
		nasIP    string
		nasPort  int
		wantAddr string
		wantErr  bool
	}{
		{
			name:     "valid IP and port",
			nasIP:    "192.168.1.1",
			nasPort:  3799,
			wantAddr: "192.168.1.1:3799",
			wantErr:  false,
		},
		{
			name:     "default port",
			nasIP:    "192.168.1.1",
			nasPort:  0,
			wantAddr: "192.168.1.1:3799",
			wantErr:  false,
		},
		{
			name:     "empty IP",
			nasIP:    "",
			nasPort:  3799,
			wantAddr: ":3799",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := getNASAddress(tt.nasIP, tt.nasPort)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if addr != tt.wantAddr {
					t.Errorf("address: got %q, want %q", addr, tt.wantAddr)
				}
			}
		})
	}
}

// TestContextCancellation tests that operations respect context cancellation.
func TestContextCancellation(t *testing.T) {
	client := NewClient(Config{
		Timeout:    5 * time.Second,
		RetryCount: 0,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := DisconnectRequest{
		NASIP:         "192.168.1.1",
		NASPort:       3799,
		Secret:        "secret",
		Username:      "testuser",
		AcctSessionID: "session123",
	}

	resp := client.SendDisconnect(ctx, req)

	if resp.Success {
		t.Error("expected failure due to cancelled context")
	}

	if resp.Error == nil {
		t.Error("expected error from cancelled context")
	}
}
