package alerting

import (
	"context"
	"testing"
	"time"
)

func TestAlertRule_Validate_ShouldPass(t *testing.T) {
	rule := &AlertRule{
		Name:       "High CPU Alert",
		MetricName: "device.cpu",
		Operator:   ">",
		Threshold:  80.0,
		Duration:   300,
		Severity:   "warning",
	}

	err := rule.Validate()
	if err != nil {
		t.Errorf("expected valid rule, got error: %v", err)
	}
}

func TestAlertRule_MissingName_ShouldFail(t *testing.T) {
	rule := &AlertRule{
		MetricName: "device.cpu",
		Operator:   ">",
		Threshold:  80.0,
		Duration:   300,
		Severity:   "warning",
	}

	err := rule.Validate()
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestAlertRule_InvalidOperator_ShouldFail(t *testing.T) {
	rule := &AlertRule{
		Name:       "High CPU Alert",
		MetricName: "device.cpu",
		Operator:   "invalid",
		Threshold:  80.0,
		Duration:   300,
		Severity:   "warning",
	}

	err := rule.Validate()
	if err == nil {
		t.Error("expected error for invalid operator, got nil")
	}
}

func TestAlertRule_InvalidSeverity_ShouldFail(t *testing.T) {
	rule := &AlertRule{
		Name:       "High CPU Alert",
		MetricName: "device.cpu",
		Operator:   ">",
		Threshold:  80.0,
		Duration:   300,
		Severity:   "invalid",
	}

	err := rule.Validate()
	if err == nil {
		t.Error("expected error for invalid severity, got nil")
	}
}

func TestAlertEngine_New_ShouldInitialize(t *testing.T) {
	engine := NewAlertEngine(nil)

	if engine == nil {
		t.Fatal("expected engine to be created")
	}

	if engine.evalTicker == nil {
		t.Error("expected eval ticker to be initialized")
	}

	if engine.shutdown == nil {
		t.Error("expected shutdown channel to be initialized")
	}
}

func TestAlertEngine_RecordMetric_ShouldStoreValue(t *testing.T) {
	engine := NewAlertEngine(nil)
	ctx := context.Background()

	timestamp := time.Now()
	engine.RecordMetric(ctx, "device-1", "cpu", 75.5, timestamp)

	// Verify metric was stored
	value, err := engine.getMetricValue("device-1:cpu")
	if err != nil {
		t.Errorf("expected to find metric, got error: %v", err)
	}

	if value != 75.5 {
		t.Errorf("expected value 75.5, got %v", value)
	}
}

func TestAlertEngine_RecordMetric_ShouldLimitHistorySize(t *testing.T) {
	engine := NewAlertEngine(nil)
	ctx := context.Background()

	timestamp := time.Now()

	// Record 150 values
	for i := 0; i < 150; i++ {
		engine.RecordMetric(ctx, "device-1", "cpu", float64(i), timestamp.Add(time.Duration(i)*time.Second))
	}

	// Verify history is limited to 100
	engine.metricHistoryMux.RLock()
	history := engine.metricHistory["device-1:cpu"]
	engine.metricHistoryMux.RUnlock()

	if len(history) != 100 {
		t.Errorf("expected history size 100, got %d", len(history))
	}
}

func TestAlertEngine_EvaluateRule_ShouldTrigger(t *testing.T) {
	engine := NewAlertEngine(nil)
	ctx := context.Background()

	rule := &AlertRule{
		Name:       "High CPU Alert",
		MetricName: "device.cpu",
		Operator:   ">",
		Threshold:  80.0,
		Duration:   0, // No duration requirement
		Severity:   "warning",
	}

	// Record metric above threshold
	engine.RecordMetric(ctx, "device-1", "cpu", 85.0, time.Now())

	// Small delay to ensure metric is recorded
	time.Sleep(10 * time.Millisecond)

	triggered := engine.evaluateRule(ctx, rule)
	if !triggered {
		t.Error("expected alert to trigger")
	}
}

func TestAlertEngine_EvaluateRule_BelowThreshold_ShouldNotTrigger(t *testing.T) {
	engine := NewAlertEngine(nil)
	ctx := context.Background()

	rule := &AlertRule{
		Name:       "High CPU Alert",
		MetricName: "device.cpu",
		Operator:   ">",
		Threshold:  80.0,
		Duration:   0,
		Severity:   "warning",
	}

	// Record metric below threshold
	engine.RecordMetric(ctx, "device-1", "cpu", 75.0, time.Now())

	time.Sleep(10 * time.Millisecond)

	triggered := engine.evaluateRule(ctx, rule)
	if triggered {
		t.Error("expected alert not to trigger")
	}
}

func TestAlertEngine_EvaluateValue_GreaterThan(t *testing.T) {
	engine := NewAlertEngine(nil)

	result := engine.evaluateValue(85.0, 80.0, ">")
	if !result {
		t.Error("expected 85.0 > 80.0 to be true")
	}

	result = engine.evaluateValue(75.0, 80.0, ">")
	if result {
		t.Error("expected 75.0 > 80.0 to be false")
	}
}

func TestAlertEngine_EvaluateValue_LessThan(t *testing.T) {
	engine := NewAlertEngine(nil)

	result := engine.evaluateValue(75.0, 80.0, "<")
	if !result {
		t.Error("expected 75.0 < 80.0 to be true")
	}

	result = engine.evaluateValue(85.0, 80.0, "<")
	if result {
		t.Error("expected 85.0 < 80.0 to be false")
	}
}

func TestAlertEngine_EvaluateValue_GreaterThanOrEqual(t *testing.T) {
	engine := NewAlertEngine(nil)

	result := engine.evaluateValue(80.0, 80.0, ">=")
	if !result {
		t.Error("expected 80.0 >= 80.0 to be true")
	}

	result = engine.evaluateValue(85.0, 80.0, ">=")
	if !result {
		t.Error("expected 85.0 >= 80.0 to be true")
	}

	result = engine.evaluateValue(75.0, 80.0, ">=")
	if result {
		t.Error("expected 75.0 >= 80.0 to be false")
	}
}

func TestAlertEngine_EvaluateValue_LessThanOrEqual(t *testing.T) {
	engine := NewAlertEngine(nil)

	result := engine.evaluateValue(80.0, 80.0, "<=")
	if !result {
		t.Error("expected 80.0 <= 80.0 to be true")
	}

	result = engine.evaluateValue(75.0, 80.0, "<=")
	if !result {
		t.Error("expected 75.0 <= 80.0 to be true")
	}

	result = engine.evaluateValue(85.0, 80.0, "<=")
	if result {
		t.Error("expected 85.0 <= 80.0 to be false")
	}
}

func TestAlertEngine_EvaluateValue_Equal(t *testing.T) {
	engine := NewAlertEngine(nil)

	result := engine.evaluateValue(80.0, 80.0, "==")
	if !result {
		t.Error("expected 80.0 == 80.0 to be true")
	}

	result = engine.evaluateValue(75.0, 80.0, "==")
	if result {
		t.Error("expected 75.0 == 80.0 to be false")
	}
}

func TestAlertEngine_Shutdown_ShouldNotPanic(t *testing.T) {
	engine := NewAlertEngine(nil)

	// Shutdown without starting should not panic
	err := engine.Shutdown()
	if err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}

func TestAlertEngine_RegisterNotifier_ShouldStore(t *testing.T) {
	engine := NewAlertEngine(nil)

	notifier := &MockNotifier{name: "test"}
	engine.RegisterNotifier("test", notifier)

	_, ok := engine.notifiers["test"]
	if !ok {
		t.Error("expected notifier to be registered")
	}
}

func TestAlertEngine_GetMetricValue_NotFound_ShouldReturnError(t *testing.T) {
	engine := NewAlertEngine(nil)

	_, err := engine.getMetricValue("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent metric")
	}
}
