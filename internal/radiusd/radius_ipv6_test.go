package radiusd

import (
	"encoding/binary"
	"testing"

	"layeh.com/radius"
)

func TestParseIPv6Attributes_ShouldExtractAllFields(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))

	// Add Framed-IPv6-Prefix attribute (97)
	// Format: tag (1) + prefix length (1) + prefix (16 bytes)
	ipv6Prefix := make([]byte, 18)
	ipv6Prefix[0] = 0 // tag
	ipv6Prefix[1] = 64 // prefix length
	// 2001:db8::1
	copy(ipv6Prefix[2:18], []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	packet.Attributes.Add(97, ipv6Prefix)

	// Add Framed-Interface-Id attribute (96)
	// Format: tag (1) + iftype (1) + ifindex (4)
	interfaceId := make([]byte, 6)
	interfaceId[0] = 0 // tag
	interfaceId[1] = 6 // iftype (Ethernet)
	binary.BigEndian.PutUint32(interfaceId[2:6], 1234) // ifindex
	packet.Attributes.Add(96, interfaceId)

	attrs := ParseIPv6Attributes(packet)

	if attrs.FramedIPv6Prefix == "" {
		t.Error("expected IPv6 prefix to be parsed")
	}
	if attrs.FramedIPv6PrefixLen != 64 {
		t.Errorf("expected prefix length 64, got %d", attrs.FramedIPv6PrefixLen)
	}
	if attrs.FramedInterfaceId == "" {
		t.Error("expected interface ID to be parsed")
	}
}

func TestGetFramedIPv6Address_ShouldParseAddress(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))

	// Add Framed-IPv6-Address attribute (168)
	// Format: tag (1) + length (1) + address (16)
	ipv6Addr := make([]byte, 17)
	ipv6Addr[0] = 0 // tag
	ipv6Addr[1] = 16 // length
	// 2001:db8::1
	copy(ipv6Addr[1:17], []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	packet.Attributes.Add(168, ipv6Addr)

	addr := GetFramedIPv6Address(packet)
	if addr == "" {
		t.Error("expected IPv6 address to be parsed")
	}
}

func TestGetDelegatedIPv6Prefix_ShouldParsePrefix(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))

	// Add Delegated-IPv6-Prefix attribute (123)
	// Format: tag (1) + prefix length (1) + prefix (variable)
	delegatedPrefix := make([]byte, 10)
	delegatedPrefix[0] = 0 // tag
	delegatedPrefix[1] = 56 // prefix length
	// 2001:db8:1::/56
	copy(delegatedPrefix[2:10], []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x01, 0, 0, 0, 0})
	packet.Attributes.Add(123, delegatedPrefix)

	prefix := GetDelegatedIPv6Prefix(packet)
	if prefix == "" {
		t.Error("expected delegated IPv6 prefix to be parsed")
	}
}

func TestParseIPv6Attributes_EmptyPacket_ShouldNotPanic(t *testing.T) {
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))

	attrs := ParseIPv6Attributes(packet)

	// Should return empty attributes without panicking
	if attrs.FramedIPv6Prefix != "" {
		t.Error("expected empty IPv6 prefix")
	}
	if attrs.FramedIPv6PrefixLen != 0 {
		t.Error("expected zero prefix length")
	}
	if attrs.FramedInterfaceId != "" {
		t.Error("expected empty interface ID")
	}
}
