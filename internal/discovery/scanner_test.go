package discovery

import (
	"context"
	"net"
	"testing"
	"time"
)

// TestNewScanner tests creating a new network scanner.
func TestNewScanner(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				IPRange:   "192.168.1.0/24",
				Ports:     []int{8728, 8729},
				Timeout:   2 * time.Second,
				Workers:   10,
			},
			wantErr: false,
		},
		{
			name: "invalid CIDR",
			config: Config{
				IPRange: "invalid",
				Timeout: 2 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "empty ports uses default",
			config: Config{
				IPRange: "192.168.1.0/24",
				Timeout: 2 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, err := NewScanner(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScanner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && scanner == nil {
				t.Error("expected non-nil scanner")
			}
		})
	}
}

// TestParseCIDR tests CIDR parsing.
func TestParseCIDR(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		wantIP   string
		wantNum  int
		wantErr  bool
	}{
		{
			name:    "24-bit mask",
			cidr:    "192.168.1.0/24",
			wantIP:  "192.168.1",
			wantNum: 256,
			wantErr: false,
		},
		{
			name:    "30-bit mask",
			cidr:    "192.168.1.0/30",
			wantIP:  "192.168.1",
			wantNum: 4,
			wantErr: false,
		},
		{
			name:    "16-bit mask",
			cidr:    "10.0.0.0/16",
			wantIP:  "10.0.0",
			wantNum: 65536,
			wantErr: false,
		},
		{
			name:    "invalid CIDR",
			cidr:    "192.168.1.0/33",
			wantErr: true,
		},
		{
			name:    "invalid format",
			cidr:    "192.168.1.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network, num, err := parseCIDR(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCIDR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if network != tt.wantIP {
				t.Errorf("network = %v, want %v", network, tt.wantIP)
			}
			if num != tt.wantNum {
				t.Errorf("num = %v, want %v", num, tt.wantNum)
			}
		})
	}
}

// TestScanHost tests scanning a single host.
func TestScanHost(t *testing.T) {
	// Create a simple echo server for testing
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	// Handle connection in goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	scanner := &Scanner{
		ports:   []int{8728},
		timeout: 1 * time.Second,
	}

	ctx := context.Background()
	result := scanner.scanHost(ctx, listener.Addr().String())
	
	// Should not panic and should return a result
	_ = result
}

// TestDefaultPorts tests default port configuration.
func TestDefaultPorts(t *testing.T) {
	scanner := &Scanner{
		ports:   nil,
		timeout: 1 * time.Second,
	}
	scanner.initDefaults()

	if len(scanner.ports) == 0 {
		t.Error("expected default ports to be set")
	}

	// Check default ports include RouterOS API
	found := false
	for _, p := range scanner.ports {
		if p == 8728 || p == 8729 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected default ports to include 8728 or 8729")
	}
}

// TestDiscoveryResult tests the DiscoveryResult struct.
func TestDiscoveryResult(t *testing.T) {
	result := &DiscoveryResult{
		IP:         "192.168.1.1",
		Port:       8728,
		IsRouterOS: true,
		DeviceInfo: &DeviceInfo{
			Identity:  "MyRouter",
			BoardName: "RB750GS",
			Version:   "7.13.5",
			Model:     "Mikrotik RB750GS",
			Serial:   "123456",
		},
		Timestamp: time.Now(),
	}

	if result.IP != "192.168.1.1" {
		t.Errorf("IP = %v, want %v", result.IP, "192.168.1.1")
	}
	if !result.IsRouterOS {
		t.Error("expected IsRouterOS to be true")
	}
	if result.DeviceInfo.BoardName != "RB750GS" {
		t.Errorf("BoardName = %v, want %v", result.DeviceInfo.BoardName, "RB750GS")
	}
}

// TestScanResult tests the ScanResult struct.
func TestScanResult(t *testing.T) {
	result := &ScanResult{
		CIDR:      "192.168.1.0/24",
		StartedAt: time.Now(),
		Results: []*DiscoveryResult{
			{IP: "192.168.1.1", IsRouterOS: true},
			{IP: "192.168.1.2", IsRouterOS: false},
		},
	}

	if result.CIDR != "192.168.1.0/24" {
		t.Errorf("CIDR = %v, want %v", result.CIDR, "192.168.1.0/24")
	}
	if len(result.Results) != 2 {
		t.Errorf("Results length = %v, want %v", len(result.Results), 2)
	}
	result.FinishedAt = time.Now()
	if result.FinishedAt.Before(result.StartedAt) {
		t.Error("FinishedAt should be after StartedAt")
	}
}

// TestDiscoveryResult_ToNAS tests converting discovery result to NAS model.
func TestDiscoveryResult_ToNAS(t *testing.T) {
	result := &DiscoveryResult{
		IP:         "192.168.1.1",
		Port:       8728,
		IsRouterOS: true,
		DeviceInfo: &DeviceInfo{
			Identity:  "Office-Router",
			BoardName: "RB750GS",
			Version:   "7.13.5",
			Model:     "Mikrotik RB750GS",
			Serial:   "ABC123",
		},
		Timestamp: time.Now(),
	}

	nas := result.ToNAS("secret123")
	if nas.Ipaddr != "192.168.1.1" {
		t.Errorf("Ipaddr = %v, want %v", nas.Ipaddr, "192.168.1.1")
	}
	if nas.Name != "Office-Router" {
		t.Errorf("Name = %v, want %v", nas.Name, "Office-Router")
	}
	if nas.Model != "Mikrotik RB750GS" {
		t.Errorf("Model = %v, want %v", nas.Model, "Mikrotik RB750GS")
	}
	if nas.VendorCode != "mikrotik" {
		t.Errorf("VendorCode = %v, want %v", nas.VendorCode, "mikrotik")
	}
	if nas.Secret != "secret123" {
		t.Errorf("Secret = %v, want %v", nas.Secret, "secret123")
	}
}
