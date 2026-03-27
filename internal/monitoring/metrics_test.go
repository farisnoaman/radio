package monitoring

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// setupTestCollector creates a new collector and registers all its metrics with a custom registry.
// This helper function reduces test code duplication.
func setupTestCollector(t *testing.T) (*prometheus.Registry, *TenantMetricsCollector) {
	registry := prometheus.NewRegistry()
	collector := NewTenantMetricsCollector()

	// Register all collector metrics with the custom registry
	registry.MustRegister(
		collector.radiusAuthRate,
		collector.radiusAcctRate,
		collector.authErrors,
		collector.onlineSessions,
		collector.deviceCpuUsage,
		collector.deviceMemoryUsage,
		collector.deviceUptime,
		collector.deviceStatus,
		collector.networkLatency,
		collector.packetLoss,
		collector.bandwidthUsage,
	)

	return registry, collector
}

func TestRecordAuthMetric(t *testing.T) {
	registry, collector := setupTestCollector(t)

	// Test recording successful auth
	collector.RecordAuth(1, true)
	collector.RecordAuth(1, false)

	// Verify metrics registered
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Find radius_auth_total metric
	var found bool
	for _, m := range metrics {
		if m.GetName() == "radius_auth_total" {
			found = true

			// Verify we have the right labels
			if len(m.Metric) < 2 {
				t.Error("Expected at least 2 metric samples (success + failure)")
			}

			// Check that tenant_id label exists
			for _, metric := range m.Metric {
				hasTenantID := false
				hasResult := false
				for _, label := range metric.Label {
					if label.GetName() == "tenant_id" && label.GetValue() == "1" {
						hasTenantID = true
					}
					if label.GetName() == "result" {
						hasResult = true
					}
				}
				if !hasTenantID {
					t.Error("Expected tenant_id label to be '1'")
				}
				if !hasResult {
					t.Error("Expected result label")
				}
			}
			break
		}
	}

	if !found {
		t.Error("radius_auth_total metric not found")
	}
}

func TestRecordDeviceMetric(t *testing.T) {
	registry, collector := setupTestCollector(t)

	// Test recording device health
	collector.RecordDeviceHealth(context.Background(), 1, "device1", "192.168.1.1", 75.5, 60.2, true)

	// Verify metric exists
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, m := range metrics {
		if m.GetName() == "device_cpu_usage_percent" {
			found = true

			// Verify the CPU value was recorded
			if len(m.Metric) != 1 {
				t.Errorf("Expected 1 metric sample, got %d", len(m.Metric))
			}

			metric := m.Metric[0]
			gauge := metric.Gauge
			if gauge == nil {
				t.Error("Expected gauge type")
			} else if gauge.GetValue() != 75.5 {
				t.Errorf("Expected CPU value 75.5, got %f", gauge.GetValue())
			}

			// Verify labels
			labels := make(map[string]string)
			for _, label := range metric.Label {
				labels[label.GetName()] = label.GetValue()
			}

			if labels["tenant_id"] != "1" {
				t.Errorf("Expected tenant_id '1', got '%s'", labels["tenant_id"])
			}
			if labels["device_id"] != "device1" {
				t.Errorf("Expected device_id 'device1', got '%s'", labels["device_id"])
			}
			if labels["device_ip"] != "192.168.1.1" {
				t.Errorf("Expected device_ip '192.168.1.1', got '%s'", labels["device_ip"])
			}

			break
		}
	}

	if !found {
		t.Error("device_cpu_usage_percent metric not found")
	}
}
