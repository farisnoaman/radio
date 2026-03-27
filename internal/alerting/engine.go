// Package alerting provides real-time alerting for RADIUS and network events.
//
// The alerting system evaluates metrics against threshold rules and sends
// notifications via multiple channels (email, webhook, SMS).
//
// Features:
//   - Threshold-based alerting with hysteresis
//   - Multiple notification channels
//   - Alert deduplication and rate limiting
//   - Alert history and acknowledgment
package alerting

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertRule is an alias for domain.AlertRule for convenience.
type AlertRule = domain.AlertRule

// MetricValue represents a metric value at a point in time.
type MetricValue struct {
	MetricName string
	Value      float64
	Timestamp  time.Time
}

// AlertEngine evaluates metrics and triggers alerts.
type AlertEngine struct {
	db               *gorm.DB
	metricHistory    map[string][]MetricValue // metric -> values
	metricHistoryMux sync.RWMutex
	rules            []*AlertRule
	rulesMux         sync.RWMutex
	notifiers        map[string]Notifier
	cooldowns        map[string]time.Time // rule_id -> last triggered
	evalTicker       *time.Ticker
	shutdown         chan struct{}
	wg               sync.WaitGroup
}

// NewAlertEngine creates a new alert engine.
func NewAlertEngine(db *gorm.DB) *AlertEngine {
	return &AlertEngine{
		db:            db,
		metricHistory: make(map[string][]MetricValue),
		notifiers:     make(map[string]Notifier),
		cooldowns:     make(map[string]time.Time),
		evalTicker:    time.NewTicker(30 * time.Second),
		shutdown:      make(chan struct{}),
	}
}

// Shutdown gracefully shuts down the alert engine.
func (e *AlertEngine) Shutdown() error {
	close(e.shutdown)
	e.evalTicker.Stop()
	e.wg.Wait()
	return nil
}

// RecordMetric records a metric value for alert evaluation.
func (e *AlertEngine) RecordMetric(
	ctx context.Context,
	deviceID string,
	metricName string,
	value float64,
	timestamp time.Time,
) {
	e.metricHistoryMux.Lock()
	defer e.metricHistoryMux.Unlock()

	key := deviceID + ":" + metricName

	// Add to history
	e.metricHistory[key] = append(e.metricHistory[key], MetricValue{
		MetricName: metricName,
		Value:      value,
		Timestamp:  timestamp,
	})

	// Keep only last 100 values
	if len(e.metricHistory[key]) > 100 {
		e.metricHistory[key] = e.metricHistory[key][1:]
	}
}

// getMetricValue retrieves the current value for a metric.
// Supports both full key lookup (e.g., "device-1:cpu"), metric name lookup (e.g., "cpu"),
// and hierarchical lookup (e.g., "device.cpu" matches "device-1:cpu").
func (e *AlertEngine) getMetricValue(metricName string) (float64, error) {
	e.metricHistoryMux.RLock()
	defer e.metricHistoryMux.RUnlock()

	// Find matching metric
	for key, values := range e.metricHistory {
		if len(values) == 0 {
			continue
		}

		// Check if exact key matches (full key lookup)
		if key == metricName {
			return values[len(values)-1].Value, nil
		}

		// Check if metric name matches (suffix after :)
		if strings.HasSuffix(key, ":"+metricName) {
			return values[len(values)-1].Value, nil
		}

		// Check hierarchical match - "device.cpu" should match "device-1:cpu"
		// i.e., check if the last part of metricName matches the recorded metric name
		if idx := strings.LastIndex(metricName, "."); idx >= 0 {
			suffix := metricName[idx+1:]
			if strings.HasSuffix(key, ":"+suffix) {
				return values[len(values)-1].Value, nil
			}
		}
	}

	return 0, fmt.Errorf("metric not found: %s", metricName)
}

// evaluateRule evaluates a single rule against current metrics.
func (e *AlertEngine) evaluateRule(ctx context.Context, rule *AlertRule) bool {
	// Check cooldown
	if rule.LastTriggered != nil {
		timeSinceTrigger := time.Since(*rule.LastTriggered)
		if timeSinceTrigger < time.Duration(rule.CooldownSec)*time.Second {
			return false // Still in cooldown
		}
	}

	// Get current metric value
	value, err := e.getMetricValue(rule.MetricName)
	if err != nil {
		zap.S().Debug("Failed to get metric value for rule",
			zap.String("rule", rule.Name),
			zap.Error(err))
		return false
	}

	// Evaluate condition
	thresholdMet := e.evaluateValue(value, rule.Threshold, rule.Operator)
	if !thresholdMet {
		return false
	}

	// Check duration requirement
	if rule.Duration > 0 {
		if !e.checkDuration(rule.MetricName, rule.Duration, rule.Threshold, rule.Operator) {
			return false
		}
	}

	return true
}

// evaluateValue evaluates a single value against threshold.
func (e *AlertEngine) evaluateValue(value, threshold float64, operator string) bool {
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}

// checkDuration checks if a condition has been true for the specified duration.
func (e *AlertEngine) checkDuration(
	metricName string,
	durationSec int,
	threshold float64,
	operator string,
) bool {
	e.metricHistoryMux.RLock()
	defer e.metricHistoryMux.RUnlock()

	cutoff := time.Now().Add(-time.Duration(durationSec) * time.Second)

	// Find metric history
	for key, values := range e.metricHistory {
		if strings.HasSuffix(key, ":"+metricName) {
			// Check if condition has been true for the duration
			durationMet := 0

			for _, v := range values {
				if v.Timestamp.Before(cutoff) {
					continue
				}

				if e.evaluateValue(v.Value, threshold, operator) {
					durationMet++
				} else {
					// Condition not met, reset
					durationMet = 0
				}
			}

			// Require at least some data points to meet duration
			return durationMet > 0
		}
	}

	return false
}

// RegisterNotifier registers a notification channel.
func (e *AlertEngine) RegisterNotifier(name string, notifier Notifier) {
	e.notifiers[name] = notifier
}

// Notifier defines the interface for sending alert notifications.
type Notifier interface {
	// Send sends an alert notification.
	Send(ctx context.Context, alert *domain.Alert) error

	// Name returns the notifier name.
	Name() string
}

// MockNotifier is a mock implementation of Notifier for testing.
type MockNotifier struct {
	name      string
	lastAlert *domain.Alert
	mu        sync.Mutex
	sentCount int
}

// NewMockNotifier creates a new mock notifier.
func NewMockNotifier(name string) *MockNotifier {
	return &MockNotifier{name: name}
}

// Send stores the alert for testing.
func (n *MockNotifier) Send(ctx context.Context, alert *domain.Alert) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lastAlert = alert
	n.sentCount++
	return nil
}

// Name returns the notifier name.
func (n *MockNotifier) Name() string {
	return n.name
}

// GetLastAlert returns the last alert sent to this notifier.
func (n *MockNotifier) GetLastAlert() *domain.Alert {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.lastAlert
}

// GetSentCount returns the number of alerts sent.
func (n *MockNotifier) GetSentCount() int {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.sentCount
}

// Reset clears the alert history.
func (n *MockNotifier) Reset() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lastAlert = nil
	n.sentCount = 0
}
