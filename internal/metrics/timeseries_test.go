package metrics

import (
	"context"
	"testing"
	"time"
)

func TestTimeSeriesStore_WriteMetric_ShouldSucceed(t *testing.T) {
	// Use mock store for testing
	store := NewMockStore()

	metric := &Metric{
		Name:      "auth_requests_total",
		Timestamp: time.Now(),
		Tags: map[string]string{
			"tenant_id": "1",
			"result":    "success",
		},
		Fields: map[string]interface{}{
			"value": 42.0,
		},
	}

	err := store.WriteMetric(context.Background(), metric)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Verify metric was stored
	metrics := store.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Name != "auth_requests_total" {
		t.Errorf("expected metric name 'auth_requests_total', got '%s'", metrics[0].Name)
	}
}

func TestMetricCollector_RecordAuth_ShouldStoreMetric(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record auth success
	collector.RecordAuth(ctx, 1, "success", nil)

	// Verify metric was written
	metrics := store.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Name != "auth_requests_total" {
		t.Errorf("expected metric name 'auth_requests_total', got '%s'", metrics[0].Name)
	}

	if metrics[0].Tags["result"] != "success" {
		t.Errorf("expected result tag 'success', got '%s'", metrics[0].Tags["result"])
	}
}

func TestMetricCollector_RecordAcctUpdate_ShouldStoreMetrics(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record accounting update
	collector.RecordAcctUpdate(ctx, 1, "session-123", 1024000, 2048000)

	// Verify metrics were written
	metrics := store.GetMetrics()
	if len(metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(metrics))
	}

	// Check input bytes metric
	inputFound := false
	outputFound := false
	for _, m := range metrics {
		if m.Name == "acct_input_bytes" {
			inputFound = true
			if m.Fields["value"] != float64(1024000) {
				t.Errorf("expected input bytes 1024000, got %v", m.Fields["value"])
			}
		}
		if m.Name == "acct_output_bytes" {
			outputFound = true
			if m.Fields["value"] != float64(2048000) {
				t.Errorf("expected output bytes 2048000, got %v", m.Fields["value"])
			}
		}
	}

	if !inputFound {
		t.Error("acct_input_bytes metric not found")
	}
	if !outputFound {
		t.Error("acct_output_bytes metric not found")
	}
}

func TestMetricCollector_RecordDeviceHealth_ShouldStoreMetric(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record device health
	collector.RecordDeviceHealth(ctx, 1, "device-1", "192.168.1.1", 75.5, 45.2, true)

	// Verify metric was written
	metrics := store.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Name != "device_health" {
		t.Errorf("expected metric name 'device_health', got '%s'", metrics[0].Name)
	}

	if metrics[0].Fields["cpu"] != 75.5 {
		t.Errorf("expected cpu 75.5, got %v", metrics[0].Fields["cpu"])
	}

	if metrics[0].Tags["status"] != "online" {
		t.Errorf("expected status 'online', got '%s'", metrics[0].Tags["status"])
	}
}

func TestTimeSeriesStore_QueryMetrics_ShouldReturnResults(t *testing.T) {
	store := NewMockStore()

	// Write some test metrics
	metric := &Metric{
		Name:      "test_metric",
		Timestamp: time.Now(),
		Tags: map[string]string{
			"tenant_id": "1",
		},
		Fields: map[string]interface{}{
			"value": 100.0,
		},
	}

	store.WriteMetric(context.Background(), metric)

	// Query metrics
	results, err := store.QueryMetrics(context.Background(), "test_metric")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestMetricCollector_RecordAuth_WithError(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record auth failure
	collector.RecordAuth(ctx, 1, "failure", nil)

	// Verify metric was written
	metrics := store.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Tags["result"] != "failure" {
		t.Errorf("expected result tag 'failure', got '%s'", metrics[0].Tags["result"])
	}
}

func TestMetricCollector_RecordDeviceHealth_Offline(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record offline device
	collector.RecordDeviceHealth(ctx, 1, "device-2", "192.168.1.2", 0, 0, false)

	// Verify metric was written
	metrics := store.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Tags["status"] != "offline" {
		t.Errorf("expected status 'offline', got '%s'", metrics[0].Tags["status"])
	}

	if metrics[0].Fields["online"] != false {
		t.Errorf("expected online false, got %v", metrics[0].Fields["online"])
	}
}

func TestMetricCollector_GetAuthRate_ShouldReturnSum(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record multiple auth requests
	collector.RecordAuth(ctx, 1, "success", nil)
	collector.RecordAuth(ctx, 1, "success", nil)
	collector.RecordAuth(ctx, 1, "failure", nil)

	// Get auth rate
	rate, err := collector.GetAuthRate(ctx, 1, "1h")
	if err != nil {
		t.Fatalf("GetAuthRate failed: %v", err)
	}

	// Should return sum of all values (3.0)
	if rate != 3.0 {
		t.Errorf("expected auth rate 3.0, got %v", rate)
	}
}

func TestMetricCollector_GetAuthRate_NoMetrics_ShouldReturnZero(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Get auth rate with no metrics
	rate, err := collector.GetAuthRate(ctx, 1, "1h")
	if err != nil {
		t.Fatalf("GetAuthRate failed: %v", err)
	}

	if rate != 0 {
		t.Errorf("expected auth rate 0, got %v", rate)
	}
}

func TestTimeSeriesStore_Close_ShouldNotPanic(t *testing.T) {
	store := NewMockStore()

	// Close should not panic
	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestMetricCollector_RecordAuth_WriteError_ShouldHandleGracefully(t *testing.T) {
	store := NewFailingStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record auth - should handle error gracefully without panicking
	collector.RecordAuth(ctx, 1, "success", nil)

	// Test passes if no panic occurs
}

func TestMetricCollector_RecordAcctUpdate_WriteError_ShouldHandleGracefully(t *testing.T) {
	store := NewFailingStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record accounting update - should handle error gracefully
	collector.RecordAcctUpdate(ctx, 1, "session-123", 1024000, 2048000)

	// Test passes if no panic occurs
}

func TestMetricCollector_RecordDeviceHealth_WriteError_ShouldHandleGracefully(t *testing.T) {
	store := NewFailingStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record device health - should handle error gracefully
	collector.RecordDeviceHealth(ctx, 1, "device-1", "192.168.1.1", 75.5, 45.2, true)

	// Test passes if no panic occurs
}

func TestMetricCollector_GetAuthRate_WrongTenant_ShouldReturnZero(t *testing.T) {
	store := NewMockStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Record auth for tenant 1
	collector.RecordAuth(ctx, 1, "success", nil)

	// Query for tenant 2 (should return 0)
	rate, err := collector.GetAuthRate(ctx, 2, "1h")
	if err != nil {
		t.Fatalf("GetAuthRate failed: %v", err)
	}

	if rate != 0 {
		t.Errorf("expected auth rate 0 for different tenant, got %v", rate)
	}
}

func TestTimeSeriesStore_WriteMetrics_ErrorOnFirstMetric_ShouldReturnError(t *testing.T) {
	store := NewFailingStore()

	metrics := []*Metric{
		{Name: "test1", Timestamp: time.Now(), Tags: map[string]string{}, Fields: map[string]interface{}{}},
		{Name: "test2", Timestamp: time.Now(), Tags: map[string]string{}, Fields: map[string]interface{}{}},
	}

	err := store.WriteMetrics(context.Background(), metrics)
	if err == nil {
		t.Error("expected error when writing metrics fails")
	}
}

func TestFailingStore_QueryMetrics_ShouldReturnError(t *testing.T) {
	store := NewFailingStore()

	results, err := store.QueryMetrics(context.Background(), "any_query")
	if err == nil {
		t.Error("expected error from QueryMetrics, got nil")
	}

	if results != nil {
		t.Errorf("expected nil results on error, got %v", results)
	}
}

func TestFailingStore_Close_ShouldNotPanic(t *testing.T) {
	store := NewFailingStore()

	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestTimeSeriesStore_WriteMetrics_EmptyList_ShouldSucceed(t *testing.T) {
	store := NewMockStore()

	err := store.WriteMetrics(context.Background(), []*Metric{})
	if err != nil {
		t.Errorf("WriteMetrics with empty list failed: %v", err)
	}
}

func TestMetricCollector_GetAuthRate_QueryError_ShouldReturnError(t *testing.T) {
	store := NewFailingStore()
	collector := NewMetricCollector(store)

	ctx := context.Background()

	// Get auth rate - should return error when query fails
	rate, err := collector.GetAuthRate(ctx, 1, "1h")
	if err == nil {
		t.Error("expected error from GetAuthRate on query failure, got nil")
	}

	if rate != 0 {
		t.Errorf("expected rate 0 on query error, got %v", rate)
	}
}

func TestTimeSeriesStore_WriteMetrics_SecondMetricFails_ShouldReturnError(t *testing.T) {
	store := NewConditionalFailingStore()

	metrics := []*Metric{
		{Name: "test1", Timestamp: time.Now(), Tags: map[string]string{}, Fields: map[string]interface{}{}},
		{Name: "test2", Timestamp: time.Now(), Tags: map[string]string{}, Fields: map[string]interface{}{}},
	}

	err := store.WriteMetrics(context.Background(), metrics)
	if err == nil {
		t.Error("expected error when second metric write fails")
	}
}
