package radiusd

import (
	"encoding/binary"
	"fmt"
	"net"

	"layeh.com/radius"
)

// IPv6Attributes holds parsed IPv6-specific RADIUS attributes.
type IPv6Attributes struct {
	FramedIPv6Prefix    string // RFC 3162 attribute 97
	FramedIPv6PrefixLen int    // Prefix length
	FramedInterfaceId   string // RFC 3162 attribute 96
}

// ParseIPv6Attributes extracts IPv6-specific attributes from RADIUS packet.
func ParseIPv6Attributes(pkt *radius.Packet) IPv6Attributes {
	attrs := IPv6Attributes{}

	// Framed-IPv6-Prefix (RFC 3162, attribute 97)
	// This attribute includes the prefix length
	if v, len := getFramedIPv6Prefix(pkt); v != "" {
		attrs.FramedIPv6Prefix = v
		attrs.FramedIPv6PrefixLen = len
	}

	// Framed-Interface-Id (RFC 3162, attribute 96)
	if v := getFramedInterfaceId(pkt); v != "" {
		attrs.FramedInterfaceId = v
	}

	return attrs
}

// getFramedIPv6Prefix extracts Framed-IPv6-Prefix attribute (97).
// Format: tag (1 byte) + prefix length (1 byte) + prefix (variable)
// Returns: prefix string and prefix length
func getFramedIPv6Prefix(pkt *radius.Packet) (string, int) {
	attr := pkt.Attributes.Get(97)
	if len(attr) < 3 {
		return "", 0
	}

	prefixLen := int(attr[1])
	if len(attr) < 2+prefixLen/8 {
		return "", 0
	}

	// Convert prefix bytes to IPv6 string
	prefixBytes := attr[2:]
	if len(prefixBytes) < 16 {
		// Pad to 16 bytes if needed
		padded := make([]byte, 16)
		copy(padded, prefixBytes)
		prefixBytes = padded
	}

	ip := net.IP(prefixBytes)
	return ip.String(), prefixLen
}

// getFramedIPv6PrefixLen extracts Framed-IPv6-Prefix-Length attribute (98).
func getFramedIPv6PrefixLen(pkt *radius.Packet) int {
	attr := pkt.Attributes.Get(98)
	if len(attr) != 1 {
		return 0
	}
	return int(attr[0])
}

// getFramedInterfaceId extracts Framed-Interface-Id attribute (96).
// Format: tag (1 byte) + iftype (1 byte) + ifindex (4 bytes)
func getFramedInterfaceId(pkt *radius.Packet) string {
	attr := pkt.Attributes.Get(96)
	if len(attr) < 6 {
		return ""
	}

	ifType := attr[1]
	ifIndex := binary.BigEndian.Uint32(attr[2:6])

	return fmt.Sprintf("%d/%d", ifType, ifIndex)
}

// GetFramedIPv6Address extracts Framed-IPv6-Address attribute (168).
func GetFramedIPv6Address(pkt *radius.Packet) string {
	attr := pkt.Attributes.Get(168)
	if len(attr) < 17 { // tag (1) + length (1) + address (16)
		return ""
	}

	ip := net.IP(attr[1:17])
	return ip.String()
}

// GetDelegatedIPv6Prefix extracts Delegated-IPv6-Prefix attribute (123).
func GetDelegatedIPv6Prefix(pkt *radius.Packet) string {
	attr := pkt.Attributes.Get(123)
	if len(attr) < 3 {
		return ""
	}

	prefixLen := int(attr[1])
	prefixBytes := attr[2:]

	// Calculate how many bytes we need
	prefixByteLen := (prefixLen + 7) / 8
	if len(prefixBytes) < prefixByteLen {
		return ""
	}

	// Pad to 16 bytes for IPv6
	padded := make([]byte, 16)
	copy(padded, prefixBytes)

	ip := net.IP(padded)
	return fmt.Sprintf("%s/%d", ip.String(), prefixLen)
}
