package routeros

import (
	"testing"
	"time"
)

// TestNewClient tests creation of RouterOS client with default and custom config.
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantAddr string
		wantTLS  bool
	}{
		{
			name: "default config",
			config: Config{
				Address: "192.168.1.1",
			},
			wantAddr: "192.168.1.1:8728",
			wantTLS:  false,
		},
		{
			name: "custom port",
			config: Config{
				Address: "192.168.1.1:8729",
			},
			wantAddr: "192.168.1.1:8729",
			wantTLS:  false,
		},
		{
			name: "with TLS",
			config: Config{
				Address: "192.168.1.1",
				UseTLS:  true,
			},
			wantAddr: "192.168.1.1:8729",
			wantTLS:  true,
		},
		{
			name: "with credentials",
			config: Config{
				Address:  "192.168.1.1",
				Username: "admin",
				Password: "password",
			},
			wantAddr: "192.168.1.1:8728",
			wantTLS:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			if client == nil {
				t.Fatal("expected non-nil client")
			}
			if client.config.Address != tt.wantAddr {
				t.Errorf("address: got %v, want %v", client.config.Address, tt.wantAddr)
			}
			if client.config.UseTLS != tt.wantTLS {
				t.Errorf("UseTLS: got %v, want %v", client.config.UseTLS, tt.wantTLS)
			}
		})
	}
}

// TestClient_Timeout tests default timeout configuration.
func TestClient_Timeout(t *testing.T) {
	client := NewClient(Config{})
	if client.config.Timeout != 10*time.Second {
		t.Errorf("default timeout: got %v, want 10s", client.config.Timeout)
	}

	client = NewClient(Config{Timeout: 5 * time.Second})
	if client.config.Timeout != 5*time.Second {
		t.Errorf("custom timeout: got %v, want 5s", client.config.Timeout)
	}
}

// TestClient_Close tests closing an unconnected client.
func TestClient_Close(t *testing.T) {
	client := NewClient(Config{
		Address: "192.168.1.1",
	})

	// Closing unconnected client should not panic
	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestParseResponse tests parsing RouterOS API responses.
func TestParseResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantType ResponseType
	}{
		{
			name:     "done response",
			data:     []byte{0x14, 0x00, 0x00, 0x00},
			wantType: ResponseDone,
		},
		{
			name:     "re response",
			data:     []byte{0x11, 0x00, 0x00, 0x00},
			wantType: ResponseRe,
		},
		{
			name:     "trap response",
			data:     []byte{0x17, 0x00, 0x00, 0x00},
			wantType: ResponseTrap,
		},
		{
			name:     "fatal response",
			data:     []byte{0x1a, 0x00, 0x00, 0x00},
			wantType: ResponseFatal,
		},
		{
			name:     "empty data",
			data:     []byte{},
			wantType: ResponseUnknown,
		},
		{
			name:     "unknown response type",
			data:     []byte{0xff, 0x00, 0x00, 0x00},
			wantType: ResponseUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := parseResponse(tt.data)
			if resp.Type != tt.wantType {
				t.Errorf("parseResponse() type = %v, want %v", resp.Type, tt.wantType)
			}
		})
	}
}

// TestBuildLoginRequest tests building RouterOS login request.
func TestBuildLoginRequest(t *testing.T) {
	client := NewClient(Config{
		Username: "admin",
		Password: "password123",
	})

	req := client.buildLoginRequest()
	if len(req) == 0 {
		t.Error("expected non-empty login request")
	}

	// Verify packet has 4-byte length header
	if len(req) < 4 {
		t.Errorf("login request too short: %d bytes", len(req))
	}
}

// TestBuildCommand tests building RouterOS API command.
func TestBuildCommand(t *testing.T) {
	client := NewClient(Config{})

	cmd := client.BuildCommand("/system/identity/get")
	if len(cmd) == 0 {
		t.Error("expected non-empty command")
	}

	// Verify packet has 4-byte length header
	if len(cmd) < 4 {
		t.Errorf("command too short: %d bytes", len(cmd))
	}
}

// TestExtractValue tests extracting values from response data.
func TestExtractValue(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      string
		wantVal  string
	}{
		{
			name:    "simple value",
			data:    []byte("name=RouterOS\x00"),
			key:     "name",
			wantVal: "RouterOS",
		},
		{
			name:    "value with equals",
			data:    []byte("key=value=with=equals\x00"),
			key:     "key",
			wantVal: "value=with=equals",
		},
		{
			name:    "missing key",
			data:    []byte("other=value\x00"),
			key:     "name",
			wantVal: "",
		},
		{
			name:    "empty data",
			data:    []byte{},
			key:     "name",
			wantVal: "",
		},
		{
			name:    "board name with hyphen",
			data:    []byte("board-name=RB750GS\x00"),
			key:     "board-name",
			wantVal: "RB750GS",
		},
		{
			name:    "version",
			data:    []byte("version=7.13.5\x00"),
			key:     "version",
			wantVal: "7.13.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractValue(tt.data, tt.key)
			if got != tt.wantVal {
				t.Errorf("extractValue(%q, %q) = %q, want %q", string(tt.data), tt.key, got, tt.wantVal)
			}
		})
	}
}

// TestAppendWord tests building RouterOS word format data.
func TestAppendWord(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		wantLen int
	}{
		{
			name:    "simple word",
			word:    "test",
			wantLen: 9, // 4 bytes length + 5 bytes (test + null)
		},
		{
			name:    "empty word",
			word:    "",
			wantLen: 5, // 4 bytes length + 1 byte (null)
		},
		{
			name:    "key=value",
			word:    "name=value",
			wantLen: 15, // 4 bytes length + 11 bytes (name=value + null)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendWord(nil, tt.word)
			if len(result) != tt.wantLen {
				t.Errorf("appendWord(%q) length = %d, want %d", tt.word, len(result), tt.wantLen)
			}
		})
	}
}

// TestSystemInfo tests the SystemInfo struct.
func TestSystemInfo(t *testing.T) {
	info := &SystemInfo{
		Identity:  "MyRouter",
		BoardName: "RB750GS",
		Version:   "7.13.5",
		Model:     "Mikrotik RB750GS",
		Serial:   "1234567890",
	}

	if info.Identity != "MyRouter" {
		t.Errorf("Identity: got %q, want %q", info.Identity, "MyRouter")
	}
	if info.BoardName != "RB750GS" {
		t.Errorf("BoardName: got %q, want %q", info.BoardName, "RB750GS")
	}
	if info.Version != "7.13.5" {
		t.Errorf("Version: got %q, want %q", info.Version, "7.13.5")
	}
}

// TestClient_State tests client state tracking.
func TestClient_State(t *testing.T) {
	client := NewClient(Config{})

	// Initial state
	if client.loggedIn {
		t.Error("expected not logged in initially")
	}
	if client.conn != nil {
		t.Error("expected nil connection initially")
	}

	// After setting connection (simulated)
	client.conn = nil
	client.loggedIn = false

	// Verify Close handles nil connection
	err := client.Close()
	if err != nil {
		t.Errorf("Close() with nil conn: got error %v", err)
	}
}
