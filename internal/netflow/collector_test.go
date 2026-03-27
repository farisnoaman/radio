package netflow

import (
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestCollector_ProcessNetFlowV9_ShouldSucceed(t *testing.T) {
	config := &CollectorConfig{
		ListenAddr:    ":2056",
		BufferSize:    1000,
		EnableLogging: false,
		BatchSize:     10,
		BatchTimeout:  5 * time.Second,
	}

	collector := NewCollector(nil, config)
	if collector == nil {
		t.Fatal("expected collector to be created")
	}

	// Test mock packet processing with data flow set
	packet := createMockNetFlowV9PacketWithData()
	count, err := collector.ProcessPacket(packet)
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if count == 0 {
		t.Error("expected at least one flow record")
	}
}

func TestCollector_ConfigDefaults_ShouldSetReasonableValues(t *testing.T) {
	config := &CollectorConfig{
		ListenAddr: ":0",
	}

	collector := NewCollector(nil, config)
	if collector.config.BufferSize == 0 {
		t.Error("expected default buffer size")
	}
	if collector.config.BatchSize == 0 {
		t.Error("expected default batch size")
	}
	if collector.config.BatchTimeout == 0 {
		t.Error("expected default batch timeout")
	}
}

func TestCollector_ValidateRecord_ShouldAcceptValid(t *testing.T) {
	record := &domain.NetFlowRecord{
		RouterID:      "router-1",
		SourceAddr:    "192.168.1.100",
		DestAddr:      "8.8.8.8",
		Protocol:      17,
		Bytes:         1024,
		Packets:       10,
		FlowDuration:  5000,
		FirstSwitched: time.Now().Add(-5 * time.Second),
		LastSwitched:  time.Now(),
	}

	err := record.Validate()
	if err != nil {
		t.Errorf("expected valid record, got error: %v", err)
	}
}

// Helper function to create mock NetFlow v9 packet
func createMockNetFlowV9Packet() []byte {
	// Simplified NetFlow v9 packet structure
	// Version (2) + Count (2) + SysUptime (4) + UnixSeconds (4) + UnixNanoseconds (4) +
	// FlowSequence (4) + SourceID (4) = 24 bytes header
	packet := make([]byte, 24)

	// Version: 9
	packet[0] = 0
	packet[1] = 9

	// Count: 1 flow set
	packet[2] = 0
	packet[3] = 1

	// Mock flow set data (template + data)
	flowSet := []byte{
		// Flow Set ID: 0 (Template)
		0, 0,
		// Length: 20
		0, 20,
		// Template ID: 256
		1, 0,
		// Field count: 2
		0, 2,
		// Field 1: Source IPv4 (type 8, length 4)
		0, 8, 0, 4,
		// Field 2: Destination IPv4 (type 12, length 4)
		0, 12, 0, 4,
	}

	return append(packet, flowSet...)
}

// Helper function to create mock NetFlow v9 packet with template and data
func createMockNetFlowV9PacketWithData() []byte {
	// NetFlow v9 packet header (24 bytes)
	packet := make([]byte, 24)

	// Version: 9
	packet[0] = 0
	packet[1] = 9

	// Count: 2 flow sets (template + data)
	packet[2] = 0
	packet[3] = 2

	// Template Flow Set
	templateFlowSet := []byte{
		// Flow Set ID: 0 (Template)
		0, 0,
		// Length: 20
		0, 20,
		// Template ID: 256
		1, 0,
		// Field count: 4
		0, 4,
		// Field 1: Source IPv4 (type 8, length 4)
		0, 8, 0, 4,
		// Field 2: Destination IPv4 (type 12, length 4)
		0, 12, 0, 4,
		// Field 3: Protocol (type 4, length 1)
		0, 4, 0, 1,
		// Field 4: Bytes (type 1, length 4)
		0, 1, 0, 4,
	}

	// Data Flow Set
	dataFlowSet := []byte{
		// Flow Set ID: 256 (Data)
		1, 0,
		// Length: 13
		0, 13,
		// Data: Source IP (192.168.1.100)
		192, 168, 1, 100,
		// Data: Dest IP (8.8.8.8)
		8, 8, 8, 8,
		// Data: Protocol (17 = UDP)
		17,
		// Data: Bytes (1024)
		0, 0, 4, 0,
	}

	packet = append(packet, templateFlowSet...)
	packet = append(packet, dataFlowSet...)

	return packet
}

func TestCollector_ParseHeader_ShouldExtractFields(t *testing.T) {
	packet := createMockNetFlowV9Packet()

	collector := NewCollector(nil, &CollectorConfig{})
	version, count, sourceID := collector.ParseHeader(packet)

	if version != 9 {
		t.Errorf("expected version 9, got %d", version)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
	if sourceID != 0 {
		t.Errorf("expected source ID 0, got %d", sourceID)
	}
}

func TestCollector_DecodeTemplate_ShouldStoreTemplate(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	// Mock template data
	templateData := []byte{
		// Template ID
		1, 0,
		// Field count
		0, 2,
		// Field 1: Source IPv4
		0, 8, 0, 4,
		// Field 2: Destination IPv4
		0, 12, 0, 4,
	}

	template := collector.decodeTemplate(templateData)
	if template == nil {
		t.Error("expected template to be created")
	}

	// Check template has 2 fields
	if len(template) != 1 {
		t.Errorf("expected 1 template, got %d", len(template))
	}
	fields, ok := template[256]
	if !ok {
		t.Fatal("expected template ID 256 to exist")
	}
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
}

func TestCollector_ProcessPacket_TooShort_ShouldFail(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	// Packet too short (less than 24 bytes header)
	packet := make([]byte, 10)

	_, err := collector.ProcessPacket(packet)
	if err == nil {
		t.Error("expected error for too short packet")
	}
}

func TestCollector_ProcessPacket_WrongVersion_ShouldReturnError(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	// Wrong version (version 5 instead of 9)
	packet := make([]byte, 24)
	packet[1] = 5 // Wrong version

	count, err := collector.ProcessPacket(packet)
	if err == nil {
		t.Error("expected error for wrong version")
	}
	if count != 0 {
		t.Error("expected 0 records for wrong version")
	}
}

func TestCollector_ProcessPacket_EmptyTemplate_ShouldReturnZero(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	packet := make([]byte, 24)
	packet[1] = 9 // Version 9
	packet[2] = 0
	packet[3] = 1 // Count: 1 flow set

	// Data flow set without template
	dataFlowSet := []byte{
		1, 0,     // Flow Set ID: 256 (Data)
		0, 10,    // Length
		192, 168, 1, 100, // Source IP
		8, 8, 8, 8,       // Dest IP
	}

	packet = append(packet, dataFlowSet...)

	count, err := collector.ProcessPacket(packet)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Error("expected 0 records when no template exists")
	}
}

func TestCollector_Shutdown_ShouldNotPanic(t *testing.T) {
	config := &CollectorConfig{
		ListenAddr: ":0", // Use random port
	}

	collector := NewCollector(nil, config)

	// Shutdown without starting should not panic
	err := collector.Shutdown()
	if err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}

func TestCollector_ParseField_AllFieldTypes(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	record := &domain.NetFlowRecord{
		FirstSwitched: time.Now(),
		LastSwitched:  time.Now(),
	}

	// Test Source IPv4 (field type 8)
	collector.parseField(record, 8, []byte{192, 168, 1, 100})
	if record.SourceAddr != "192.168.1.100" {
		t.Errorf("expected source IP 192.168.1.100, got %s", record.SourceAddr)
	}

	// Test Destination IPv4 (field type 12)
	collector.parseField(record, 12, []byte{8, 8, 8, 8})
	if record.DestAddr != "8.8.8.8" {
		t.Errorf("expected dest IP 8.8.8.8, got %s", record.DestAddr)
	}

	// Test Protocol (field type 4)
	collector.parseField(record, 4, []byte{17}) // UDP
	if record.Protocol != 17 {
		t.Errorf("expected protocol 17, got %d", record.Protocol)
	}

	// Test TOS (field type 5)
	collector.parseField(record, 5, []byte{128})
	if record.Tos != 128 {
		t.Errorf("expected TOS 128, got %d", record.Tos)
	}

	// Test TCP Flags (field type 6)
	collector.parseField(record, 6, []byte{0x18}) // SYN+ACK
	if record.TcpFlags != 0x18 {
		t.Errorf("expected TCP flags 0x18, got %d", record.TcpFlags)
	}

	// Test Bytes (field type 1) - 4 bytes
	collector.parseField(record, 1, []byte{0, 0, 16, 64}) // 4160 bytes
	if record.Bytes != 4160 {
		t.Errorf("expected bytes 4160, got %d", record.Bytes)
	}

	// Test Packets (field type 2) - 4 bytes
	collector.parseField(record, 2, []byte{0, 0, 0, 100}) // 100 packets
	if record.Packets != 100 {
		t.Errorf("expected packets 100, got %d", record.Packets)
	}

	// Test Source Port (field type 7)
	collector.parseField(record, 7, []byte{0x30, 0x39}) // 12345
	if record.SourcePort != 12345 {
		t.Errorf("expected source port 12345, got %d", record.SourcePort)
	}

	// Test Dest Port (field type 10)
	collector.parseField(record, 10, []byte{0, 53}) // 53
	if record.DestPort != 53 {
		t.Errorf("expected dest port 53, got %d", record.DestPort)
	}

	// Test Flow Duration (field type 152)
	collector.parseField(record, 152, []byte{0, 0, 0x13, 0x88}) // 5000 ms
	if record.FlowDuration != 5000 {
		t.Errorf("expected flow duration 5000, got %d", record.FlowDuration)
	}
}

func TestCollector_ProcessTemplateFlowSet(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})
	sourceID := uint32(123)

	// Template flow set data
	data := []byte{
		// Template ID
		1, 0,
		// Field count
		0, 2,
		// Field 1: Source IPv4 (type 8, length 4)
		0, 8, 0, 4,
		// Field 2: Destination IPv4 (type 12, length 4)
		0, 12, 0, 4,
	}

	collector.processTemplateFlowSet(data, sourceID)

	// Check template was stored
	collector.templateMux.RLock()
	templates := collector.templates[sourceID]
	collector.templateMux.RUnlock()

	if templates == nil {
		t.Fatal("expected templates to be stored for source ID")
	}

	if len(templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(templates))
	}
}

func TestCollector_DecodeDataFlowSet_NoTemplate_ShouldReturnNil(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	data := []byte{
		192, 168, 1, 100, // Source IP
		8, 8, 8, 8,       // Dest IP
	}

	records := collector.decodeDataFlowSet(data, 999, time.Now())
	if records != nil {
		t.Error("expected nil records when no template exists")
	}
}

func TestCollector_DecodeDataFlowSet_ValidTemplate_ShouldReturnRecords(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})
	sourceID := uint32(1)

	// First, store a template with all required fields for validation
	collector.templateMux.Lock()
	if collector.templates[sourceID] == nil {
		collector.templates[sourceID] = make(map[uint16][]TemplateField)
	}
	collector.templates[sourceID][256] = []TemplateField{
		{Type: 8, Length: 4},  // Source IPv4
		{Type: 12, Length: 4}, // Dest IPv4
		{Type: 4, Length: 1},  // Protocol (required for validation)
	}
	collector.templateMux.Unlock()

	// Now decode data flow set with protocol
	data := []byte{
		192, 168, 1, 100, // Source IP
		8, 8, 8, 8,       // Dest IP
		17,               // Protocol (UDP)
	}

	records := collector.decodeDataFlowSet(data, sourceID, time.Now())
	if records == nil {
		t.Fatal("expected records to be decoded")
	}

	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}

	if records[0].SourceAddr != "192.168.1.100" {
		t.Errorf("expected source IP 192.168.1.100, got %s", records[0].SourceAddr)
	}

	if records[0].DestAddr != "8.8.8.8" {
		t.Errorf("expected dest IP 8.8.8.8, got %s", records[0].DestAddr)
	}

	if records[0].Protocol != 17 {
		t.Errorf("expected protocol 17, got %d", records[0].Protocol)
	}
}

func TestCollector_GetTrafficStats_ShouldReturnError(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	// No database configured
	_, err := collector.GetTrafficStats(nil, 1, "day")
	if err == nil {
		t.Error("expected error when database not configured")
	}
}

func TestCollector_GetTrafficStats_ValidTimeRanges(t *testing.T) {
	// Test that time range parsing works
	collector := NewCollector(nil, &CollectorConfig{})

	validRanges := []string{"hour", "day", "week", "month"}
	for _, rangeVal := range validRanges {
		// Should not panic on valid ranges
		_, err := collector.GetTrafficStats(nil, 1, rangeVal)
		if err == nil {
			t.Error("expected error when database not configured")
		}
	}
}

func TestCollector_GetTrafficStats_InvalidTimeRange(t *testing.T) {
	collector := NewCollector(nil, &CollectorConfig{})

	// Invalid time range should default to "day"
	_, err := collector.GetTrafficStats(nil, 1, "invalid")
	if err == nil {
		t.Error("expected error when database not configured")
	}
}


