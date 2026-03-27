package dhcp

import (
	"testing"
)

func TestParseOption82_ValidData_ShouldExtractFields(t *testing.T) {
	// Example DHCP option 82 data:
	// Subopt 1 (Agent Circuit ID): "eth1/1/1:101"
	// Subopt 2 (Agent Remote ID): "switch02-hostname"
	data := []byte{0x01, 0x0C, 0x65, 0x74, 0x68, 0x31, 0x2F, 0x31, 0x2F, 0x31, 0x3A, 0x31, 0x30, 0x31,
		0x02, 0x11, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x30, 0x32, 0x2D, 0x68, 0x6F, 0x73, 0x74, 0x6E, 0x61, 0x6D, 0x65}

	circuitID, remoteID, err := ParseOption82(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if circuitID != "eth1/1/1:101" {
		t.Errorf("expected circuit ID 'eth1/1/1:101', got '%s'", circuitID)
	}

	if remoteID != "switch02-hostname" {
		t.Errorf("expected remote ID 'switch02-hostname', got '%s'", remoteID)
	}
}
