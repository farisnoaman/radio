package domain

import (
	"testing"
	"time"
)

func TestNetFlowRecord_ValidRecord_ShouldPass(t *testing.T) {
	record := &NetFlowRecord{
		SourceAddr:    "192.168.1.100",
		DestAddr:      "8.8.8.8",
		SourcePort:    12345,
		DestPort:      53,
		Protocol:      17, // UDP
		Bytes:         1024,
		Packets:       10,
		FlowDuration:  5000, // milliseconds
		FirstSwitched: time.Now().Add(-5 * time.Second),
		LastSwitched:   time.Now(),
	}

	err := record.Validate()
	if err != nil {
		t.Fatalf("expected valid record, got error: %v", err)
	}
}

func TestNetFlowRecord_MissingAddresses_ShouldFail(t *testing.T) {
	record := &NetFlowRecord{
		Protocol:      17,
		FirstSwitched: time.Now(),
		LastSwitched:  time.Now(),
	}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected error for missing addresses, got nil")
	}
}

func TestNetFlowRecord_InvalidSourceIP_ShouldFail(t *testing.T) {
	record := &NetFlowRecord{
		SourceAddr:   "invalid-ip",
		DestAddr:     "8.8.8.8",
		Protocol:     17,
		FirstSwitched: time.Now(),
		LastSwitched:  time.Now(),
	}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected error for invalid source IP, got nil")
	}
}

func TestNetFlowRecord_InvalidDestIP_ShouldFail(t *testing.T) {
	record := &NetFlowRecord{
		SourceAddr:   "192.168.1.100",
		DestAddr:     "invalid-ip",
		Protocol:     17,
		FirstSwitched: time.Now(),
		LastSwitched:  time.Now(),
	}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected error for invalid destination IP, got nil")
	}
}

func TestNetFlowRecord_MissingProtocol_ShouldFail(t *testing.T) {
	record := &NetFlowRecord{
		SourceAddr:   "192.168.1.100",
		DestAddr:     "8.8.8.8",
		Protocol:     0, // Missing
		FirstSwitched: time.Now(),
		LastSwitched:  time.Now(),
	}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected error for missing protocol, got nil")
	}
}

func TestNetFlowRecord_MissingFirstSwitched_ShouldFail(t *testing.T) {
	record := &NetFlowRecord{
		SourceAddr:   "192.168.1.100",
		DestAddr:     "8.8.8.8",
		Protocol:     17,
		FirstSwitched: time.Time{}, // Zero
		LastSwitched:  time.Now(),
	}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected error for missing first switched, got nil")
	}
}

func TestNetFlowRecord_MissingLastSwitched_ShouldFail(t *testing.T) {
	record := &NetFlowRecord{
		SourceAddr:   "192.168.1.100",
		DestAddr:     "8.8.8.8",
		Protocol:     17,
		FirstSwitched: time.Now(),
		LastSwitched:  time.Time{}, // Zero
	}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected error for missing last switched, got nil")
	}
}

func TestNetFlowRecord_LastBeforeFirst_ShouldFail(t *testing.T) {
	now := time.Now()
	record := &NetFlowRecord{
		SourceAddr:   "192.168.1.100",
		DestAddr:     "8.8.8.8",
		Protocol:     17,
		FirstSwitched: now,
		LastSwitched:  now.Add(-1 * time.Second),
	}

	err := record.Validate()
	if err == nil {
		t.Fatal("expected error for last switched before first, got nil")
	}
}

func TestNetFlowRecord_ValidWithOnlyDestAddr_ShouldPass(t *testing.T) {
	record := &NetFlowRecord{
		DestAddr:      "8.8.8.8",
		Protocol:      17,
		FirstSwitched: time.Now().Add(-5 * time.Second),
		LastSwitched:  time.Now(),
	}

	err := record.Validate()
	if err != nil {
		t.Fatalf("expected valid record with only dest addr, got error: %v", err)
	}
}

func TestTrafficSummary_CalculateMetrics_ShouldReturnCorrect(t *testing.T) {
	summary := &TrafficSummary{
		TotalBytes:  1024000,
		TotalPackets: 10000,
		TotalFlows:   500,
		DurationSec:  300,
	}

	mbps := summary.GetMBPS()
	if mbps <= 0 {
		t.Error("expected positive Mbps")
	}

	pps := summary.GetPPS()
	if pps <= 0 {
		t.Error("expected positive PPS")
	}

	expectedMBPS := (float64(1024000) * 8 / float64(300)) / 1_000_000
	if mbps < expectedMBPS-0.1 || mbps > expectedMBPS+0.1 {
		t.Errorf("expected Mbps ~%.2f, got %.2f", expectedMBPS, mbps)
	}
}

func TestTrafficSummary_ZeroDuration_ShouldReturnZero(t *testing.T) {
	summary := &TrafficSummary{
		TotalBytes:  1024000,
		TotalPackets: 10000,
		DurationSec:  0,
	}

	if mbps := summary.GetMBPS(); mbps != 0 {
		t.Errorf("expected 0 Mbps for zero duration, got %.2f", mbps)
	}

	if pps := summary.GetPPS(); pps != 0 {
		t.Errorf("expected 0 PPS for zero duration, got %.2f", pps)
	}
}

func TestTrafficSummary_GetProtocolName(t *testing.T) {
	tests := []struct {
		protocol     uint8
		expectedName string
	}{
		{1, "ICMP"},
		{6, "TCP"},
		{17, "UDP"},
		{58, "IPv6-ICMP"},
		{99, "Protocol-99"},
		{255, "Protocol-255"},
	}

	for _, tt := range tests {
		record := &NetFlowRecord{Protocol: tt.protocol}
		if name := record.GetProtocolName(); name != tt.expectedName {
			t.Errorf("protocol %d: expected '%s', got '%s'", tt.protocol, tt.expectedName, name)
		}
	}
}

func TestNetFlowRecord_IsIPv6(t *testing.T) {
	tests := []struct {
		addr     string
		expected bool
	}{
		{"192.168.1.100", false},
		{"8.8.8.8", false},
		{"2001:db8::1", true},
		{"2001:4860:4860::8888", true},
		{"::1", true},
		{"fe80::1", true},
		{"", false}, // Empty address
	}

	for _, tt := range tests {
		record := &NetFlowRecord{SourceAddr: tt.addr}
		if result := record.IsIPv6(); result != tt.expected {
			t.Errorf("IsIPv6(%s): expected %v, got %v", tt.addr, tt.expected, result)
		}
	}
}

func TestNetFlowRecord_IsIPv6_WithInvalidIP_ShouldReturnFalse(t *testing.T) {
	record := &NetFlowRecord{
		SourceAddr: "not-an-ip",
	}

	if record.IsIPv6() {
		t.Error("expected false for invalid IP address")
	}
}

func TestTrafficSummary_GetGBHours(t *testing.T) {
	tests := []struct {
		bytes      uint64
		durationSec int
		expected   float64
	}{
		{1024 * 1024 * 1024, 3600, 1.0},     // 1 GB for 1 hour
		{2 * 1024 * 1024 * 1024, 3600, 2.0}, // 2 GB for 1 hour
		{1024 * 1024 * 1024, 1800, 0.5},     // 1 GB for 30 minutes
		{1024 * 1024 * 1024, 7200, 2.0},     // 1 GB for 2 hours
	}

	for _, tt := range tests {
		summary := &TrafficSummary{
			TotalBytes:  tt.bytes,
			DurationSec: tt.durationSec,
		}

		gbHours := summary.GetGBHours()
		if gbHours < tt.expected-0.01 || gbHours > tt.expected+0.01 {
			t.Errorf("bytes=%d duration=%d: expected GB-Hours ~%.2f, got %.2f",
				tt.bytes, tt.durationSec, tt.expected, gbHours)
		}
	}
}

func TestTrafficSummary_GetGBHours_ZeroDuration_ShouldReturnZero(t *testing.T) {
	summary := &TrafficSummary{
		TotalBytes:  1024 * 1024 * 1024,
		DurationSec: 0,
	}

	if gbHours := summary.GetGBHours(); gbHours != 0 {
		t.Errorf("expected 0 GB-Hours for zero duration, got %.2f", gbHours)
	}
}
