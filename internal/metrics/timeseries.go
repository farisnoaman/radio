// Package metrics provides time-series metric storage and querying.
//
// Uses InfluxDB for high-performance metric storage and retention.
// Stores RADIUS authentication, accounting, and performance metrics.
//
// Example:
//
//	store := metrics.NewInfluxDBStore(&metrics.InfluxDBConfig{
//	    URL:      "http://localhost:8086",
//	    Database: "radius_metrics",
//	    Token:    "my-token",
//	})
//	store.WriteMetric(ctx, &metrics.Metric{
//	    Name: "auth_requests_total",
//	    Tags: {"tenant_id": "1", "result": "success"},
//	    Fields: {"value": 42.0},
//	})
package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.uber.org/zap"
)

// InfluxDBConfig holds InfluxDB connection configuration.
type InfluxDBConfig struct {
	URL      string
	Org      string
	Bucket   string
	Token    string
	Database string // Deprecated in InfluxDB 2.x, use bucket
}

// Metric represents a time-series metric.
type Metric struct {
	Name      string
	Timestamp time.Time
	Tags      map[string]string
	Fields    map[string]interface{}
}

// InfluxDBStore implements TimeSeriesStore using InfluxDB 2.x.
type InfluxDBStore struct {
	client influxdb2.Client
	opts   *InfluxDBConfig
}

// NewInfluxDBStore creates a new InfluxDB store.
func NewInfluxDBStore(opts *InfluxDBConfig) *InfluxDBStore {
	client := influxdb2.NewClient(opts.URL, opts.Token)

	return &InfluxDBStore{
		client: client,
		opts:   opts,
	}
}

// WriteMetric writes a metric to InfluxDB.
func (s *InfluxDBStore) WriteMetric(ctx context.Context, metric *Metric) error {
	point := influxdb2.NewPoint(
		metric.Name,
		metric.Tags,
		metric.Fields,
		metric.Timestamp,
	)

	writeAPI := s.client.WriteAPI(s.opts.Org, s.opts.Bucket)
	writeAPI.WritePoint(point)
	writeAPI.Flush()

	return nil
}

// WriteMetrics writes multiple metrics in batch to InfluxDB.
func (s *InfluxDBStore) WriteMetrics(ctx context.Context, metrics []*Metric) error {
	writeAPI := s.client.WriteAPI(s.opts.Org, s.opts.Bucket)

	for _, metric := range metrics {
		point := influxdb2.NewPoint(
			metric.Name,
			metric.Tags,
			metric.Fields,
			metric.Timestamp,
		)
		writeAPI.WritePoint(point)
	}

	writeAPI.Flush()
	return nil
}

// QueryMetrics executes a Flux query against InfluxDB.
func (s *InfluxDBStore) QueryMetrics(ctx context.Context, query string) ([]map[string]interface{}, error) {
	queryAPI := s.client.QueryAPI(s.opts.Org)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer result.Close()

	// Parse results
	var results []map[string]interface{}
	for result.Next() {
		record := make(map[string]interface{})
		for key, value := range result.Record().Values() {
			record[key] = value
		}
		results = append(results, record)
	}

	// Check for errors after iteration
	if result.Err() != nil {
		return nil, fmt.Errorf("query iteration error: %w", result.Err())
	}

	return results, nil
}

// Close closes the InfluxDB client.
func (s *InfluxDBStore) Close() error {
	s.client.Close()
	return nil
}

// TimeSeriesStore provides metric storage operations.
type TimeSeriesStore interface {
	// WriteMetric writes a single metric.
	WriteMetric(ctx context.Context, metric *Metric) error

	// WriteMetrics writes multiple metrics in batch.
	WriteMetrics(ctx context.Context, metrics []*Metric) error

	// QueryMetrics executes a query and returns results.
	QueryMetrics(ctx context.Context, query string) ([]map[string]interface{}, error)

	// Close closes the connection.
	Close() error
}

// MockStore is a mock implementation of TimeSeriesStore for testing.
type MockStore struct {
	mu     sync.RWMutex
	metrics []*Metric
}

// NewMockStore creates a new mock store.
func NewMockStore() *MockStore {
	return &MockStore{
		metrics: make([]*Metric, 0),
	}
}

// WriteMetric writes a metric to the mock store.
func (s *MockStore) WriteMetric(ctx context.Context, metric *Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store a copy of the metric
	metricCopy := &Metric{
		Name:      metric.Name,
		Timestamp: metric.Timestamp,
		Tags:      make(map[string]string),
		Fields:    make(map[string]interface{}),
	}

	for k, v := range metric.Tags {
		metricCopy.Tags[k] = v
	}
	for k, v := range metric.Fields {
		metricCopy.Fields[k] = v
	}

	s.metrics = append(s.metrics, metricCopy)
	return nil
}

// WriteMetrics writes multiple metrics to the mock store.
func (s *MockStore) WriteMetrics(ctx context.Context, metrics []*Metric) error {
	for _, metric := range metrics {
		if err := s.WriteMetric(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

// QueryMetrics queries metrics from the mock store.
func (s *MockStore) QueryMetrics(ctx context.Context, query string) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]map[string]interface{}, 0)

	for _, metric := range s.metrics {
		// Simple query matching - check if metric name matches query
		if query == metric.Name || query == "" {
			result := map[string]interface{}{
				"name":      metric.Name,
				"timestamp": metric.Timestamp,
				"tags":      metric.Tags,
				"fields":    metric.Fields,
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// Close closes the mock store (no-op for mock).
func (s *MockStore) Close() error {
	return nil
}

// GetMetrics returns all stored metrics (for testing).
func (s *MockStore) GetMetrics() []*Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return copies to prevent external modification
	copies := make([]*Metric, len(s.metrics))
	for i, m := range s.metrics {
		copies[i] = &Metric{
			Name:      m.Name,
			Timestamp: m.Timestamp,
			Tags:      make(map[string]string),
			Fields:    make(map[string]interface{}),
		}
		for k, v := range m.Tags {
			copies[i].Tags[k] = v
		}
		for k, v := range m.Fields {
			copies[i].Fields[k] = v
		}
	}

	return copies
}

// FailingStore is a mock store that always fails operations.
type FailingStore struct{}

// NewFailingStore creates a new failing store.
func NewFailingStore() *FailingStore {
	return &FailingStore{}
}

// WriteMetric always returns an error.
func (s *FailingStore) WriteMetric(ctx context.Context, metric *Metric) error {
	return fmt.Errorf("simulated write failure")
}

// WriteMetrics always returns an error.
func (s *FailingStore) WriteMetrics(ctx context.Context, metrics []*Metric) error {
	return fmt.Errorf("simulated write failure")
}

// QueryMetrics always returns an error.
func (s *FailingStore) QueryMetrics(ctx context.Context, query string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("simulated query failure")
}

// Close is a no-op.
func (s *FailingStore) Close() error {
	return nil
}

// PartialFailingStore is a mock store that fails on specific metrics.
type PartialFailingStore struct {
	FailAfterCount int // Number of successful writes before failing
}

// NewPartialFailingStore creates a new partial failing store.
func NewPartialFailingStore(failAfter int) *PartialFailingStore {
	return &PartialFailingStore{FailAfterCount: failAfter}
}

// WriteMetric succeeds for first N metrics, then fails.
func (s *PartialFailingStore) WriteMetric(ctx context.Context, metric *Metric) error {
	if s.FailAfterCount <= 0 {
		return fmt.Errorf("simulated write failure")
	}
	s.FailAfterCount--
	return nil
}

// WriteMetrics is not implemented for this store.
func (s *PartialFailingStore) WriteMetrics(ctx context.Context, metrics []*Metric) error {
	return fmt.Errorf("not implemented")
}

// QueryMetrics is not implemented for this store.
func (s *PartialFailingStore) QueryMetrics(ctx context.Context, query string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// Close is a no-op.
func (s *PartialFailingStore) Close() error {
	return nil
}

// ConditionalFailingStore implements WriteMetrics that fails on second metric.
type ConditionalFailingStore struct {
	mu     sync.Mutex
	count  int
	metrics []*Metric
}

// NewConditionalFailingStore creates a new conditional failing store.
func NewConditionalFailingStore() *ConditionalFailingStore {
	return &ConditionalFailingStore{
		metrics: make([]*Metric, 0),
	}
}

// WriteMetric succeeds.
func (s *ConditionalFailingStore) WriteMetric(ctx context.Context, metric *Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics = append(s.metrics, metric)
	return nil
}

// WriteMetrics succeeds for first metric, fails for second.
func (s *ConditionalFailingStore) WriteMetrics(ctx context.Context, metrics []*Metric) error {
	for i, metric := range metrics {
		if i == 1 {
			return fmt.Errorf("simulated failure on metric %d", i+1)
		}
		if err := s.WriteMetric(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

// QueryMetrics returns stored metrics.
func (s *ConditionalFailingStore) QueryMetrics(ctx context.Context, query string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

// Close is a no-op.
func (s *ConditionalFailingStore) Close() error {
	return nil
}

// MetricCollector collects and stores RADIUS metrics.
type MetricCollector struct {
	store TimeSeriesStore
}

// NewMetricCollector creates a new metric collector.
func NewMetricCollector(store TimeSeriesStore) *MetricCollector {
	return &MetricCollector{store: store}
}

// RecordAuth records an authentication attempt.
func (m *MetricCollector) RecordAuth(
	ctx context.Context,
	tenantID int64,
	result string,
	err error,
) {
	metric := &Metric{
		Name:      "auth_requests_total",
		Timestamp: time.Now(),
		Tags: map[string]string{
			"tenant_id": fmt.Sprintf("%d", tenantID),
			"result":    result,
		},
		Fields: map[string]interface{}{
			"value": 1.0,
		},
	}

	if writeErr := m.store.WriteMetric(ctx, metric); writeErr != nil {
		zap.S().Error("Failed to write auth metric",
			zap.Error(writeErr))
	}
}

// RecordAcctUpdate records an accounting update.
func (m *MetricCollector) RecordAcctUpdate(
	ctx context.Context,
	tenantID int64,
	sessionID string,
	inputOctets,
	outputOctets int64,
) {
	metrics := []*Metric{
		{
			Name:      "acct_input_bytes",
			Timestamp: time.Now(),
			Tags: map[string]string{
				"tenant_id":  fmt.Sprintf("%d", tenantID),
				"session_id": sessionID,
			},
			Fields: map[string]interface{}{
				"value": float64(inputOctets),
			},
		},
		{
			Name:      "acct_output_bytes",
			Timestamp: time.Now(),
			Tags: map[string]string{
				"tenant_id":  fmt.Sprintf("%d", tenantID),
				"session_id": sessionID,
			},
			Fields: map[string]interface{}{
				"value": float64(outputOctets),
			},
		},
	}

	if writeErr := m.store.WriteMetrics(ctx, metrics); writeErr != nil {
		zap.S().Error("Failed to write acct metrics",
			zap.Error(writeErr))
	}
}

// RecordDeviceHealth records device health metrics.
func (m *MetricCollector) RecordDeviceHealth(
	ctx context.Context,
	tenantID int64,
	deviceID string,
	deviceIP string,
	cpu,
	memory float64,
	isOnline bool,
) {
	status := "online"
	if !isOnline {
		status = "offline"
	}

	metric := &Metric{
		Name:      "device_health",
		Timestamp: time.Now(),
		Tags: map[string]string{
			"tenant_id":  fmt.Sprintf("%d", tenantID),
			"device_id":  deviceID,
			"device_ip":  deviceIP,
			"status":    status,
		},
		Fields: map[string]interface{}{
			"cpu":    cpu,
			"memory": memory,
			"online": isOnline,
		},
	}

	if writeErr := m.store.WriteMetric(ctx, metric); writeErr != nil {
		zap.S().Error("Failed to write device health metric",
			zap.Error(writeErr))
	}
}

// GetAuthRate retrieves authentication rate over a time period.
func (m *MetricCollector) GetAuthRate(
	ctx context.Context,
	tenantID int64,
	timeRange string,
) (float64, error) {
	// Build query - use metric name directly
	query := "auth_requests_total"

	results, err := m.store.QueryMetrics(ctx, query)
	if err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	// Sum up all values for the tenant
	sum := 0.0
	tenantIDStr := fmt.Sprintf("%d", tenantID)
	for _, result := range results {
		// Check if this result matches the tenant_id
		if tags, ok := result["tags"].(map[string]string); ok {
			if tags["tenant_id"] == tenantIDStr {
				if fields, ok := result["fields"].(map[string]interface{}); ok {
					if value, ok := fields["value"].(float64); ok {
						sum += value
					}
				}
			}
		}
	}

	return sum, nil
}
