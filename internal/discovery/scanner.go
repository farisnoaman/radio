// Package discovery provides network scanning functionality to discover
// MikroTik RouterOS devices on the network.
//
// This package enables auto-discovery of MikroTik devices by:
// - Scanning IP ranges for open RouterOS API ports (8728/8729)
// - Detecting RouterOS devices using the RouterOS API protocol
// - Converting discovered devices to NAS entries for database storage
//
// Example usage:
//
//	scanner, err := discovery.NewScanner(discovery.Config{
//	    IPRange: "192.168.1.0/24",
//	    Ports:   []int{8728, 8729},
//	    Timeout: 2 * time.Second,
//	    Workers: 10,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	result := scanner.Scan(ctx)
//
//	for _, device := range result.Results {
//	    fmt.Printf("Found: %s (%s)\n", device.IP, device.DeviceInfo.Model)
//	}
package discovery

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/routeros"
)

const (
	// DefaultScanTimeout is the default timeout for scanning a single host.
	DefaultScanTimeout = 2 * time.Second
	// DefaultWorkers is the default number of concurrent workers.
	DefaultWorkers = 10
	// MaxConcurrentScans limits the maximum concurrent scans.
	MaxConcurrentScans = 100
)

// Config holds configuration for the network scanner.
type Config struct {
	IPRange string        // CIDR range to scan (e.g., "192.168.1.0/24")
	Ports   []int         // Ports to scan (default: 8728, 8729)
	Timeout time.Duration // Timeout per host
	Workers int           // Number of concurrent workers
}

// DeviceInfo holds information about a discovered RouterOS device.
type DeviceInfo struct {
	Identity  string // Device name from /system/identity
	BoardName string // Board name from /system/resource
	Version   string // RouterOS version
	Model     string // Device model (derived from board name)
	Serial    string // Serial number
}

// DiscoveryResult represents a single discovered device.
type DiscoveryResult struct {
	IP         string      // IP address of the device
	Port       int         // Port where RouterOS API was found
	IsRouterOS bool        // Whether the device is RouterOS
	DeviceInfo *DeviceInfo // RouterOS device information (if IsRouterOS is true)
	Timestamp  time.Time  // Time of discovery
	Error      string      // Error message if discovery failed
}

// ScanResult holds the results of a network scan.
type ScanResult struct {
	CIDR       string              // The CIDR range that was scanned
	StartedAt  time.Time          // When the scan started
	FinishedAt time.Time           // When the scan finished
	Duration   time.Duration      // Total scan duration
	Results    []*DiscoveryResult // Discovered devices
	FoundCount int                // Number of RouterOS devices found
}

// Scanner is a network scanner for discovering RouterOS devices.
type Scanner struct {
	config  Config
	ports   []int
	timeout time.Duration
	workers int
}

// NewScanner creates a new network scanner with the given configuration.
//
// The IPRange parameter should be a valid CIDR notation (e.g., "192.168.1.0/24").
// If Ports is not specified, default ports 8728 and 8729 are used.
// If Timeout is not specified, defaults to 2 seconds.
// If Workers is not specified, defaults to 10.
func NewScanner(config Config) (*Scanner, error) {
	if config.IPRange == "" {
		return nil, fmt.Errorf("IPRange is required")
	}

	scanner := &Scanner{
		config: config,
	}

	scanner.initDefaults()

	// Validate CIDR
	_, _, err := net.ParseCIDR(config.IPRange)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	return scanner, nil
}

// initDefaults sets default values for unspecified configuration.
func (s *Scanner) initDefaults() {
	if len(s.config.Ports) == 0 {
		s.ports = []int{8728, 8729}
	} else {
		s.ports = s.config.Ports
	}

	if s.config.Timeout <= 0 {
		s.timeout = DefaultScanTimeout
	} else {
		s.timeout = s.config.Timeout
	}

	if s.config.Workers <= 0 {
		s.workers = DefaultWorkers
	} else if s.config.Workers > MaxConcurrentScans {
		s.workers = MaxConcurrentScans
	} else {
		s.workers = s.config.Workers
	}
}

// Scan performs a network scan to discover RouterOS devices.
//
// The scan iterates through all IPs in the configured CIDR range and checks
// for open RouterOS API ports. For each open port, it attempts to detect
// if the device is a RouterOS and retrieves device information.
func (s *Scanner) Scan(ctx context.Context) *ScanResult {
	result := &ScanResult{
		CIDR:      s.config.IPRange,
		StartedAt: time.Now(),
		Results:   make([]*DiscoveryResult, 0),
	}

	// Parse the CIDR to get the network and IP count
	_, network, err := net.ParseCIDR(s.config.IPRange)
	if err != nil {
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(result.StartedAt)
		return result
	}

	// Generate all IPs in the range
	ips := generateIPs(network)

	// Create work channel
	ipChan := make(chan string, len(ips))
	for _, ip := range ips {
		ipChan <- ip
	}
	close(ipChan)

	// Results channel
	resultChan := make(chan *DiscoveryResult, len(ips))

	// Worker pool
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				select {
				case <-ctx.Done():
					return
				default:
					discoveryResult := s.scanHost(ctx, ip)
					if discoveryResult != nil {
						resultChan <- discoveryResult
					}
				}
			}
		}()
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for r := range resultChan {
		result.Results = append(result.Results, r)
		if r.IsRouterOS {
			result.FoundCount++
		}
	}

	result.FinishedAt = time.Now()
	result.Duration = result.FinishedAt.Sub(result.StartedAt)

	return result
}

// scanHost scans a single host for RouterOS API.
func (s *Scanner) scanHost(ctx context.Context, ip string) *DiscoveryResult {
	result := &DiscoveryResult{
		IP:        ip,
		Timestamp: time.Now(),
	}

	for _, port := range s.ports {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		// Try to connect to the port
		addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
		
		dialer := &net.Dialer{
			Timeout: s.timeout,
		}

		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			continue
		}
		conn.Close()

		// Port is open, try to detect RouterOS
		result.Port = port
		isROS, deviceInfo := s.detectRouterOS(ctx, ip, port)
		result.IsRouterOS = isROS
		result.DeviceInfo = deviceInfo

		if isROS {
			return result
		}
	}

	// Not a RouterOS device
	return result
}

// detectRouterOS attempts to detect if a host is a RouterOS device
// and retrieves device information if it is.
func (s *Scanner) detectRouterOS(ctx context.Context, ip string, port int) (bool, *DeviceInfo) {
	// Create RouterOS client
	client := routeros.NewClient(routeros.Config{
		Address:  ip,
		Port:     fmt.Sprintf("%d", port),
		Username: "", // Will try anonymous detection
		Password: "",
		Timeout:  s.timeout,
	})

	// Try to detect RouterOS
	isROS, err := client.IsRouterOS(ctx)
	if err != nil || !isROS {
		return false, nil
	}

	// If we just want to detect, return basic info
	return true, &DeviceInfo{
		Model: "Mikrotik Device",
	}
}

// detectRouterOSWithAuth attempts to detect and get full info from RouterOS
// using provided credentials.
func (s *Scanner) detectRouterOSWithAuth(ctx context.Context, ip string, port int, username, password string) (bool, *DeviceInfo) {
	client := routeros.NewClient(routeros.Config{
		Address:  ip,
		Port:     fmt.Sprintf("%d", port),
		Username: username,
		Password: password,
		Timeout:  s.timeout,
	})

	if err := client.Connect(ctx); err != nil {
		return false, nil
	}
	defer client.Close()

	if err := client.Login(ctx); err != nil {
		// Login failed but it's RouterOS
		return true, &DeviceInfo{
			Model: "Mikrotik Device (auth required)",
		}
	}

	info, err := client.GetSystemInfo(ctx)
	if err != nil {
		return true, &DeviceInfo{
			Model: "Mikrotik Device",
		}
	}

	return true, &DeviceInfo{
		Identity:  info.Identity,
		BoardName: info.BoardName,
		Version:   info.Version,
		Model:     info.Model,
		Serial:    info.Serial,
	}
}

// generateIPs generates all IP addresses in the given network.
func generateIPs(network *net.IPNet) []string {
	ips := make([]string, 0)

	// Convert IP to uint32
	ip := network.IP.To4()
	if ip == nil {
		return ips
	}

	start := ipToUint32(ip)
	mask, _ := network.Mask.Size()
	numIPs := 1 << (32 - mask)

	// Skip network address and broadcast address for /24 and smaller
	start++
	if numIPs > 2 {
		numIPs -= 2
	}

	for i := 0; i < numIPs; i++ {
		ip := uint32ToIP(start + uint32(i))
		ips = append(ips, ip.String())
	}

	return ips
}

// parseCIDR parses a CIDR string and returns the network prefix and number of IPs.
func parseCIDR(cidr string) (string, int, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", 0, err
	}

	ip := network.IP.To4()
	if ip == nil {
		return "", 0, fmt.Errorf("invalid IP address")
	}

	maskSize, _ := network.Mask.Size()
	numIPs := 1 << (32 - maskSize)

	return ip.String()[:strings.LastIndex(ip.String(), ".")], numIPs, nil
}

// ipToUint32 converts an IPv4 address to a uint32.
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// uint32ToIp converts a uint32 to an IPv4 address.
func uint32ToIP(v uint32) net.IP {
	return net.IP{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
}

// ToNAS converts a DiscoveryResult to a NetNas domain model for database storage.
func (r *DiscoveryResult) ToNAS(secret string) *domain.NetNas {
	nas := &domain.NetNas{
		Ipaddr:     r.IP,
		CoaPort:    3799,
		VendorCode: "mikrotik",
		Status:     "enabled",
		Tags:       "discovered",
	}

	if r.DeviceInfo != nil {
		nas.Name = r.DeviceInfo.Identity
		nas.Model = r.DeviceInfo.Model
		if nas.Name == "" {
			nas.Name = r.DeviceInfo.Model
		}
	}

	if nas.Name == "" {
		nas.Name = fmt.Sprintf("Mikrotik-%s", r.IP)
	}

	nas.Secret = secret

	return nas
}
