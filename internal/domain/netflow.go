package domain

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// NetFlowRecord represents a single flow record from NetFlow v9 export.
// NetFlow provides network traffic flow information for analysis and accounting.
//
// Key fields:
//   - SourceAddr/DestAddr: IP addresses of the flow endpoints
//   - SourcePort/DestPort: Port numbers (L4)
//   - Protocol: IP protocol number (6=TCP, 17=UDP, etc.)
//   - Bytes: Number of bytes transferred in the flow
//   - Packets: Number of packets in the flow
//   - FlowDuration: Flow duration in milliseconds
type NetFlowRecord struct {
	ID              int64     `json:"id,string" gorm:"primaryKey"`
	TenantID        int64     `json:"tenant_id" gorm:"index"`
	RouterID        string    `json:"router_id" gorm:"size:100;index"`     // NetFlow exporter ID
	SourceAddr      string    `json:"source_addr" gorm:"size:45;index"`    // IPv4 or IPv6
	DestAddr        string    `json:"dest_addr" gorm:"size:45;index"`
	SourcePort      uint16    `json:"source_port"`
	DestPort        uint16    `json:"dest_port"`
	Protocol        uint8     `json:"protocol" gorm:"index"`                 // IP protocol number
	Tos             uint8     `json:"tos"`                                   // Type of Service
	TcpFlags        uint8     `json:"tcp_flags"`
	Bytes           uint64    `json:"bytes" gorm:"index"`
	Packets         uint64    `json:"packets"`
	FlowDuration    uint32    `json:"flow_duration_ms"`                    // Duration in ms
	FirstSwitched   time.Time `json:"first_switched"`
	LastSwitched    time.Time `json:"last_switched"`
	IngressInterface uint32   `json:"ingress_interface"`
	EgressInterface  uint32   `json:"egress_interface"`
	Direction        uint8    `json:"direction"`                            // 0=ingress, 1=egress
	IPv6FlowLabel    uint32   `json:"ipv6_flow_label"`                       // For IPv6 flows
	BgpNextHop       string    `json:"bgp_next_hop" gorm:"size:45"`
	BgpPrevHop       string    `json:"bgp_prev_hop" gorm:"size:45"`
	MplsLabelTop     uint32   `json:"mpls_label_top"`
	ApplicationID    uint16   `json:"application_id"`                        // Cisco/Avaya CCA
	ApplicationName  string    `json:"application_name" gorm:"size:100"`      // DPI result
	VrfID            uint32   `json:"vrf_id"`                                 // VRF instance
	UserID           int64    `json:"user_id" gorm:"index"`                  // Associated user
	SessionID        string   `json:"session_id" gorm:"size:64;index"`       // RADIUS session
	CreatedAt        time.Time `json:"created_at" gorm:"index"`
}

// TableName specifies the table name.
func (NetFlowRecord) TableName() string {
	return "netflow_record"
}

// Validate checks if the NetFlow record is valid.
func (r *NetFlowRecord) Validate() error {
	if r.SourceAddr == "" && r.DestAddr == "" {
		return errors.New("at least source or destination address is required")
	}
	if r.Protocol == 0 {
		return errors.New("protocol is required")
	}
	if r.FirstSwitched.IsZero() {
		return errors.New("first switched time is required")
	}
	if r.LastSwitched.IsZero() {
		return errors.New("last switched time is required")
	}
	if r.LastSwitched.Before(r.FirstSwitched) {
		return errors.New("last switched must be after first switched")
	}

	// Validate IP addresses
	if r.SourceAddr != "" {
		if net.ParseIP(r.SourceAddr) == nil {
			return fmt.Errorf("invalid source IP: %s", r.SourceAddr)
		}
	}
	if r.DestAddr != "" {
		if net.ParseIP(r.DestAddr) == nil {
			return fmt.Errorf("invalid destination IP: %s", r.DestAddr)
		}
	}

	return nil
}

// GetProtocolName returns the protocol name (TCP, UDP, ICMP, etc.).
func (r *NetFlowRecord) GetProtocolName() string {
	switch r.Protocol {
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 58:
		return "IPv6-ICMP"
	default:
		return fmt.Sprintf("Protocol-%d", r.Protocol)
	}
}

// IsIPv6 returns true if this is an IPv6 flow.
func (r *NetFlowRecord) IsIPv6() bool {
	if r.SourceAddr != "" {
		if ip := net.ParseIP(r.SourceAddr); ip != nil {
			return ip.To4() == nil // IPv6 if no IPv4 representation
		}
	}
	return false
}

// TrafficSummary represents aggregated traffic statistics for a time period.
type TrafficSummary struct {
	ID              int64     `json:"id,string" gorm:"primaryKey"`
	TenantID        int64     `json:"tenant_id" gorm:"index"`
	RouterID        string    `json:"router_id" gorm:"size:100"`
	UserID          int64     `json:"user_id" gorm:"index"`
	SessionID       string    `json:"session_id" gorm:"size:64;index"`
	SourceSubnet    string    `json:"source_subnet" gorm:"size:64"`
	DestSubnet      string    `json:"dest_subnet" gorm:"size:64"`
	ApplicationName string    `json:"application_name" gorm:"size:100"`
	Protocol        uint8     `json:"protocol"`
	TotalBytes      uint64    `json:"total_bytes"`
	TotalPackets    uint64    `json:"total_packets"`
	TotalFlows      uint64    `json:"total_flows"`
	DurationSec     int       `json:"duration_sec"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (TrafficSummary) TableName() string {
	return "traffic_summary"
}

// GetMBPS calculates average throughput in Mbps.
func (s *TrafficSummary) GetMBPS() float64 {
	if s.DurationSec == 0 {
		return 0
	}
	bps := float64(s.TotalBytes) * 8 / float64(s.DurationSec)
	return bps / 1_000_000 // Convert to Mbps
}

// GetPPS calculates average packets per second.
func (s *TrafficSummary) GetPPS() float64 {
	if s.DurationSec == 0 {
		return 0
	}
	return float64(s.TotalPackets) / float64(s.DurationSec)
}

// GetGBHours calculates total data in GB-hours.
func (s *TrafficSummary) GetGBHours() float64 {
	hours := float64(s.DurationSec) / 3600
	gb := float64(s.TotalBytes) / (1024 * 1024 * 1024)
	return gb * hours
}
