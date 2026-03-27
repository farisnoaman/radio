// Package dhcp implements DHCP protocol parsing and handling,
// specifically for IPoE authentication using DHCP option 82 (Relay Agent Information).
//
// Option 82 Structure (RFC 3046):
//   - Suboption 1: Agent Circuit ID (identifies the circuit/port)
//   - Suboption 2: Agent Remote ID (identifies the relay agent)
//   - Suboption 5: Link Selection
//   - Suboption 9: Vendor-Specific Information
//
// Example:
//
//	circuitID, remoteID, err := dhcp.ParseOption82(option82Data)
//	if err != nil {
//	    return nil, err
//	}
//	user := authenticateByCircuit(circuitID, remoteID)
package dhcp

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// Option 82 Suboption codes (RFC 3046).
const (
	// AgentCircuitID is suboption 1: Agent Circuit ID.
	AgentCircuitID = 1
	// AgentRemoteID is suboption 2: Agent Remote ID.
	AgentRemoteID = 2
	// LinkSelection is suboption 5: Link Selection.
	LinkSelection = 5
	// VendorSpecific is suboption 9: Vendor-Specific Information.
	VendorSpecific = 9
)

// ParseOption82 parses DHCP option 82 (Relay Agent Information).
//
// Returns the circuit ID and remote ID strings extracted from the option data.
// Format: [subopt-code, length, data..., subopt-code, length, data...]
//
// Example data:
//   01 0E 65 74 68 31 2F 31 2F 31 3A 31 30 31 02 14 73 77 69 74 63 68 30 32 ...
//
// Returns:
//   - circuitID: Agent Circuit ID suboption value
//   - remoteID: Agent Remote ID suboption value
//   - error: If parsing fails
func ParseOption82(data []byte) (circuitID, remoteID string, err error) {
	if len(data) < 3 {
		return "", "", errors.New("option 82 data too short")
	}

	// Parse suboptions
	offset := 0
	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}

		suboptCode := int(data[offset])
		length := int(data[offset+1])

		if offset+2+length > len(data) {
			return "", "", fmt.Errorf("invalid length for suboption %d", suboptCode)
		}

		suboptValue := string(data[offset+2 : offset+2+length])

		switch suboptCode {
		case AgentCircuitID:
			circuitID = suboptValue
		case AgentRemoteID:
			remoteID = suboptValue
		}

		offset += 2 + length
	}

	return circuitID, remoteID, nil
}

// ParseOption82Hex parses hex-encoded option 82 data.
// Useful when option 82 is passed as a hex string.
func ParseOption82Hex(hexData string) (circuitID, remoteID string, err error) {
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return "", "", fmt.Errorf("invalid hex data: %w", err)
	}
	return ParseOption82(data)
}

// BuildOption82 constructs DHCP option 82 data from circuit and remote IDs.
// Used for testing or when simulating DHCP relay agent behavior.
func BuildOption82(circuitID, remoteID string) []byte {
	var data []byte

	// Add Agent Circuit ID suboption
	if circuitID != "" {
		data = append(data, AgentCircuitID)
		data = append(data, byte(len(circuitID)))
		data = append(data, []byte(circuitID)...)
	}

	// Add Agent Remote ID suboption
	if remoteID != "" {
		data = append(data, AgentRemoteID)
		data = append(data, byte(len(remoteID)))
		data = append(data, []byte(remoteID)...)
	}

	return data
}

// DhcpPacket represents a simplified DHCP packet for authentication.
type DhcpPacket struct {
	MessageType uint8  // DHCPDISCOVER, DHCPREQUEST, etc.
	TransactionID string // XID
	ClientIP      string // CIADDR
	YourIP        string // YIADDR
	ServerIP      string // SIADDR
	GatewayIP     string // GIADDR
	ClientMAC     string // CHADDR
	Options       []byte // Options including option 82
	Option82      *Option82 // Parsed option 82
}

// Option82 represents parsed DHCP option 82 data.
type Option82 struct {
	CircuitID     string
	RemoteID      string
	LinkSelection string
	VendorData    map[string]string
}

// ParseDhcpPacket parses a DHCP packet and extracts authentication-relevant data.
// This is a simplified parser focused on option 82 extraction for IPoE auth.
func ParseDhcpPacket(data []byte) (*DhcpPacket, error) {
	if len(data) < 240 {
		return nil, errors.New("packet too short to be valid DHCP")
	}

	pkt := &DhcpPacket{
		MessageType: data[0],
		ClientMAC:   formatMAC(data[28:34]),
	}

	// Skip to options section (starts at byte 240)
	if len(data) <= 240 {
		return pkt, nil
	}

	options := data[240:]
	pkt.Options = options

	// Parse option 82 if present
	option82Data := extractOption(options, 82)
	if len(option82Data) > 0 {
		circuitID, remoteID, _ := ParseOption82(option82Data)
		pkt.Option82 = &Option82{
			CircuitID: circuitID,
			RemoteID:  remoteID,
		}
	}

	return pkt, nil
}

// extractOption extracts a specific DHCP option from the options field.
// Returns the option data (without option code and length).
func extractOption(options []byte, optionCode uint8) []byte {
	offset := 0
	for offset < len(options) {
		if options[offset] == 255 { // End marker
			break
		}
		if options[offset] == 0 { // Padding
			offset++
			continue
		}

		if offset+1 >= len(options) {
			break
		}

		code := options[offset]
		length := int(options[offset+1])

		if code == optionCode {
			if offset+2+length > len(options) {
				break
			}
			return options[offset+2 : offset+2+length]
		}

		offset += 2 + length
	}

	return nil
}

// formatMAC formats a 6-byte MAC address into standard string format.
func formatMAC(mac []byte) string {
	if len(mac) != 6 {
		return ""
	}
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}
