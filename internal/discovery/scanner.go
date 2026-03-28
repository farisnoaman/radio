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

	"github.com/go-routeros/routeros"
	"github.com/talkincode/toughradius/v9/internal/domain"
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
	IPRange  string        // CIDR range to scan (e.g., "192.168.1.0/24")
	Ports    []int         // Ports to scan (default: 8728, 8729)
	Timeout  time.Duration // Timeout per host
	Workers  int           // Number of concurrent workers
	Username string        // RouterOS username for authentication
	Password string        // RouterOS password for authentication
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
	config   Config
	ports    []int
	timeout  time.Duration
	workers  int
	username string
	password string
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

	s.username = s.config.Username
	s.password = s.config.Password
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

		addr := ip + ":" + fmt.Sprintf("%d", port)

		conn, err := net.DialTimeout("tcp4", addr, s.timeout)
		if err != nil {
			continue
		}
		conn.Close()

		result.Port = port
		isROS, deviceInfo := s.detectRouterOS(ctx, ip, port)
		result.IsRouterOS = isROS
		result.DeviceInfo = deviceInfo

		if isROS {
			return result
		}
	}

	return result
}

// detectRouterOS attempts to detect if a host is a RouterOS device
// and retrieves device information if it is.
func (s *Scanner) detectRouterOS(ctx context.Context, ip string, port int) (bool, *DeviceInfo) {
	addr := ip + ":" + fmt.Sprintf("%d", port)
	fmt.Printf("[Discovery] Testing %s\n", addr)

	// If credentials are provided, use authenticated detection
	if s.username != "" {
		fmt.Printf("[Discovery] Using credentials: username=%s\n", s.username)
		return s.detectRouterOSWithAuth(ctx, ip, port, s.username, s.password)
	}

	// Otherwise, try unauthenticated detection - just try to connect
	// Try to connect with empty credentials (will fail but proves it's RouterOS)
	client, err := routeros.Dial(addr, "", "")
	if err == nil {
		client.Close()
		fmt.Printf("[Discovery] Found RouterOS at %s\n", addr)
		return true, &DeviceInfo{
			Model: "Mikrotik Device",
		}
	}

	return false, nil
}

// detectRouterOSWithAuth attempts to detect and get full info from RouterOS
// using provided credentials.
func (s *Scanner) detectRouterOSWithAuth(ctx context.Context, ip string, port int, username, password string) (bool, *DeviceInfo) {
	fmt.Printf("[Discovery] Attempting authenticated detection for %s:%d\n", ip, port)
	
	client, err := routeros.Dial(ip+":8728", username, password)
	if err != nil {
		fmt.Printf("[Discovery] Connect failed: %v\n", err)
		return false, nil
	}
	fmt.Printf("[Discovery] Connected successfully, getting system info...\n")
	defer client.Close()

	// Get system identity
	re, err := client.Run("/system/identity/print")
	if err != nil {
		return true, &DeviceInfo{
			Model: "Mikrotik Device",
		}
	}

	identity := ""
	if len(re.Re) > 0 {
		identity = re.Re[0].Map["name"]
	}

	// Get system resource
	re2, err := client.Run("/system/resource/print")
	if err != nil {
		return true, &DeviceInfo{
			Identity: identity,
			Model:    "Mikrotik Device",
		}
	}

	boardName := ""
	version := ""
	serial := ""
	if len(re2.Re) > 0 {
		boardName = re2.Re[0].Map["board-name"]
		version = re2.Re[0].Map["version"]
		serial = re2.Re[0].Map["serial-number"]
	}

	fmt.Printf("[Discovery] Found RouterOS: %s, %s, %s\n", identity, boardName, version)
	return true, &DeviceInfo{
		Identity:  identity,
		BoardName: boardName,
		Version:   version,
		Model:     boardName,
		Serial:    serial,
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

	fmt.Printf("[Discovery] CIDR mask: /%d, total IPs: %d\n", mask, numIPs)

	// Skip network address and broadcast address for networks larger than /31
	if numIPs > 2 && mask < 31 {
		start++
		numIPs -= 2
	}

	fmt.Printf("[Discovery] IPs to scan: %d\n", numIPs)

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

// NeighborInfo represents a network neighbor discovered via routing protocols.
type NeighborInfo struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac,omitempty"`
	Interface string `json:"interface"`
	Protocol  string `json:"protocol"` // OSPF, BGP, PPP, static
	RemoteID  string `json:"remote_id"`
	State     string `json:"state"` // full, active, established
}

// PPPProfileInfo represents a PPP profile configuration.
type PPPProfileInfo struct {
	Name                string `json:"name"`
	LocalAddress        string `json:"local_address"`
	RemoteAddressRange  string `json:"remote_address_range"`
	UsesCount           int    `json:"uses_count"`
}

// OSPFInfo represents OSPF routing information.
type OSPFInfo struct {
	InstanceID string         `json:"instance_id"`
	AreaID     string         `json:"area_id"`
	RouterID   string         `json:"router_id"`
	Neighbors  []NeighborInfo `json:"neighbors"`
	State      string         `json:"state"`
}

// DiscoverNeighbors discovers all network neighbors via Mikrotik RouterOS API.
// This retrieves:
// - Layer 2 neighbors (switches, APs, routers) via "/ip/neighbor/print"
// - OSPF routing neighbors via "/routing/ospf/neighbor/print"
// - BGP peers via "/routing/bgp/peer/print"
// - PPP connections via "/ppp/active/print"
func (s *Scanner) DiscoverNeighbors(
	ctx context.Context,
	ip, username, password string,
) ([]NeighborInfo, error) {
	// Use the go-routeros library which handles login correctly
	client, err := routeros.Dial(ip+":8728", username, password)
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	defer client.Close()

	var neighbors []NeighborInfo

	// Discover Layer 2 neighbors (switches, APs, routers, repeaters)
	ipNeighbors, err := s.getIPNeighborsRos(client)
	if err == nil {
		neighbors = append(neighbors, ipNeighbors...)
	}

	// Discover OSPF neighbors
	ospfNeighbors, err := s.getOSPFNeighborsRos(client)
	if err == nil {
		neighbors = append(neighbors, ospfNeighbors...)
	}

	// Discover BGP peers
	bgpNeighbors, err := s.getBGPNeighborsRos(client)
	if err == nil {
		neighbors = append(neighbors, bgpNeighbors...)
	}

	// Discover PPP connections
	pppNeighbors, err := s.getPPPNeighborsRos(client)
	if err == nil {
		neighbors = append(neighbors, pppNeighbors...)
	}

	return neighbors, nil
}

// DiscoverPPPProfiles retrieves PPP profile configurations from Mikrotik.
func (s *Scanner) DiscoverPPPProfiles(
	ctx context.Context,
	ip, username, password string,
) ([]PPPProfileInfo, error) {
	client, err := routeros.Dial(ip+":8728", username, password)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	profiles, err := s.getPPPProfilesRos(client)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

// DiscoverOSPF retrieves OSPF routing information from Mikrotik.
func (s *Scanner) DiscoverOSPF(
	ctx context.Context,
	ip, username, password string,
) (*OSPFInfo, error) {
	client, err := routeros.Dial(ip+":8728", username, password)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ospf, err := s.getOSPFInstanceRos(client)
	if err != nil {
		return nil, err
	}

	return ospf, nil
}

// getIPNeighborsRos retrieves Layer 2 neighbors using go-routeros library
func (s *Scanner) getIPNeighborsRos(client *routeros.Client) ([]NeighborInfo, error) {
	re, err := client.Run("/ip/neighbor/print")
	if err != nil {
		return nil, fmt.Errorf("ip neighbor print failed: %w", err)
	}

	var neighbors []NeighborInfo
	for _, item := range re.Re {
		neighbor := NeighborInfo{
			IP:        item.Map["address"],
			MAC:       item.Map["mac-address"],
			Interface: item.Map["interface"],
			Protocol:  "MNDP",
			RemoteID:  item.Map["identity"],
			State:     "discovered",
		}

		if dt, ok := item.Map["device-type"]; ok && dt != "" {
			neighbor.RemoteID = dt + " (" + neighbor.RemoteID + ")"
		}

		neighbors = append(neighbors, neighbor)
	}

	return neighbors, nil
}

// getOSPFNeighborsRos retrieves OSPF neighbors using go-routeros library
func (s *Scanner) getOSPFNeighborsRos(client *routeros.Client) ([]NeighborInfo, error) {
	re, err := client.Run("/routing/ospf/neighbor/print")
	if err != nil {
		return nil, fmt.Errorf("ospf neighbor print failed: %w", err)
	}

	var neighbors []NeighborInfo
	for _, item := range re.Re {
		neighbor := NeighborInfo{
			IP:        item.Map["address"],
			Interface: item.Map["interface"],
			Protocol:  "OSPF",
			RemoteID:  item.Map["router-id"],
			State:     item.Map["state"],
		}
		neighbors = append(neighbors, neighbor)
	}

	return neighbors, nil
}

// getBGPNeighborsRos retrieves BGP peers using go-routeros library
func (s *Scanner) getBGPNeighborsRos(client *routeros.Client) ([]NeighborInfo, error) {
	re, err := client.Run("/routing/bgp/peer/print")
	if err != nil {
		return nil, fmt.Errorf("bgp peer print failed: %w", err)
	}

	var neighbors []NeighborInfo
	for _, item := range re.Re {
		neighbor := NeighborInfo{
			IP:        item.Map["remote-address"],
			Interface: item.Map["interface"],
			Protocol:  "BGP",
			RemoteID:  item.Map["name"],
			State:     item.Map["state"],
		}
		neighbors = append(neighbors, neighbor)
	}

	return neighbors, nil
}

// getPPPNeighborsRos retrieves PPP connections using go-routeros library
func (s *Scanner) getPPPNeighborsRos(client *routeros.Client) ([]NeighborInfo, error) {
	re, err := client.Run("/ppp/active/print")
	if err != nil {
		return nil, fmt.Errorf("ppp active print failed: %w", err)
	}

	var neighbors []NeighborInfo
	for _, item := range re.Re {
		pType := item.Map["service"]
		if pType == "" {
			pType = "PPP"
		}

		neighbor := NeighborInfo{
			IP:        item.Map["address"],
			MAC:       item.Map["caller-id"],
			Interface: item.Map["interface"],
			Protocol:  pType,
			RemoteID:  item.Map["name"],
			State:     item.Map["state"],
		}
		neighbors = append(neighbors, neighbor)
	}

	return neighbors, nil
}

// getPPPProfiles retrieves PPP profiles from RouterOS device.
func (s *Scanner) getPPPProfilesRos(client *routeros.Client) ([]PPPProfileInfo, error) {
	re, err := client.Run("/ppp/profile/print")
	if err != nil {
		return nil, fmt.Errorf("ppp profile print failed: %w", err)
	}

	var profiles []PPPProfileInfo
	for _, item := range re.Re {
		profile := PPPProfileInfo{
			Name:               item.Map["name"],
			LocalAddress:       item.Map["local-address"],
			RemoteAddressRange: item.Map["remote-address"],
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// getOSPFInstance retrieves OSPF instance information from RouterOS device.
func (s *Scanner) getOSPFInstanceRos(client *routeros.Client) (*OSPFInfo, error) {
	re, err := client.Run("/routing/ospf/instance/print")
	if err != nil {
		return nil, fmt.Errorf("ospf instance print failed: %w", err)
	}

	ospf := &OSPFInfo{
		Neighbors: []NeighborInfo{},
	}

	for _, item := range re.Re {
		ospf.InstanceID = item.Map["instance"]
		ospf.AreaID = item.Map["area"]
		ospf.RouterID = item.Map["router-id"]
		ospf.State = item.Map["disabled"]
	}

	return ospf, nil
}
