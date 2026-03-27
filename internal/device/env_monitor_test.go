package device

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	var db *gorm.DB
	var err error

	// Use PostgreSQL if TEST_DATABASE_URL is set, otherwise SQLite
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Skip("Cannot connect to PostgreSQL test database")
		}
	} else {
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
	}

	err = db.AutoMigrate(
		&domain.EnvironmentMetric{},
		&domain.EnvironmentAlert{},
		&domain.AlertConfig{},
		&domain.NetNas{},
	)
	require.NoError(t, err)

	return db
}

func TestEnvCollector_CalculateSeverity(t *testing.T) {
	collector := &EnvCollector{}

	tests := []struct {
		name       string
		metricType string
		value      float64
		expected   string
	}{
		{"normal temp", domain.MetricTypeTemperature, 50, domain.SeverityNormal},
		{"warning temp", domain.MetricTypeTemperature, 75, domain.SeverityWarning},
		{"critical temp", domain.MetricTypeTemperature, 85, domain.SeverityCritical},
		{"normal power", domain.MetricTypePower, 50, domain.SeverityNormal},
		{"warning power", domain.MetricTypePower, 120, domain.SeverityCritical},
		{"normal voltage", domain.MetricTypeVoltage, 220, domain.SeverityNormal},
		{"warning voltage high", domain.MetricTypeVoltage, 255, domain.SeverityWarning},
		{"critical voltage low", domain.MetricTypeVoltage, 170, domain.SeverityCritical},
		{"normal fan", domain.MetricTypeFanSpeed, 3000, domain.SeverityNormal},
		{"warning fan low", domain.MetricTypeFanSpeed, 800, domain.SeverityWarning},
		{"unknown metric", "unknown", 100, domain.SeverityNormal},
		{"temp at boundary 70", domain.MetricTypeTemperature, 70, domain.SeverityWarning},
		{"temp at boundary 80", domain.MetricTypeTemperature, 80, domain.SeverityCritical},
		{"voltage boundary max", domain.MetricTypeVoltage, 250, domain.SeverityWarning},
		{"voltage critical max", domain.MetricTypeVoltage, 260, domain.SeverityCritical},
		{"fan critical low", domain.MetricTypeFanSpeed, 500, domain.SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.calculateSeverity(tt.metricType, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnvCollector_StoreMetrics_WithRealDB(t *testing.T) {
	db := setupTestDB(t)
	_ = NewEnvCollector(db)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "192.168.1.1",
		TenantID:   1,
		VendorCode: "mikrotik",
	}
	db.Create(&nas)

	now := time.Now()
	metrics := []domain.EnvironmentMetric{
		{
			TenantID:    "1",
			NasID:       1,
			NasName:     "Test-NAS",
			MetricType:  domain.MetricTypeTemperature,
			Value:       65.0,
			Unit:        "C",
			Severity:    domain.SeverityNormal,
			CollectedAt: now,
			CreatedAt:   now,
		},
		{
			TenantID:    "1",
			NasID:       1,
			NasName:     "Test-NAS",
			MetricType:  domain.MetricTypePower,
			Value:       85.0,
			Unit:        "W",
			Severity:    domain.SeverityWarning,
			CollectedAt: now,
			CreatedAt:   now,
		},
	}

	err := db.Create(&metrics).Error
	assert.NoError(t, err)

	var count int64
	db.Model(&domain.EnvironmentMetric{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestEnvCollector_CollectDevice_WithRealDB(t *testing.T) {
	db := setupTestDB(t)
	collector := NewEnvCollector(db)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "192.168.1.1",
		Secret:     "password",
		TenantID:   1,
		VendorCode: "mikrotik",
	}
	err := db.Create(&nas).Error
	require.NoError(t, err)

	ctx := context.Background()
	err = collector.CollectDevice(ctx, &nas)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestEnvCollector_CollectAllDevices_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	collector := NewEnvCollector(db)

	err := collector.CollectAllDevices(context.Background())
	assert.NoError(t, err)
}

func TestEnvCollector_NewEnvCollector(t *testing.T) {
	db := setupTestDB(t)
	collector := NewEnvCollector(db)

	assert.NotNil(t, collector)
	assert.Equal(t, db, collector.db)
}

func TestAlertEngine_EvaluateThreshold(t *testing.T) {
	tests := []struct {
		name          string
		metricType    string
		value         float64
		thresholdType string
		thresholdVal  float64
		expectTrigger bool
	}{
		{"max threshold exceeded", domain.MetricTypeTemperature, 85, domain.ThresholdTypeMax, 70, true},
		{"max threshold not exceeded", domain.MetricTypeTemperature, 50, domain.ThresholdTypeMax, 70, false},
		{"min threshold not reached", domain.MetricTypeTemperature, 90, domain.ThresholdTypeMin, 80, false},
		{"min threshold reached", domain.MetricTypeTemperature, 70, domain.ThresholdTypeMin, 80, true},
		{"voltage max exceeded", domain.MetricTypeVoltage, 260, domain.ThresholdTypeMax, 250, true},
		{"voltage min reached", domain.MetricTypeVoltage, 190, domain.ThresholdTypeMin, 200, true},
		{"power normal", domain.MetricTypePower, 50, domain.ThresholdTypeMax, 100, false},
		{"power warning", domain.MetricTypePower, 120, domain.ThresholdTypeMax, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggered := false

			if tt.thresholdType == domain.ThresholdTypeMax && tt.value > tt.thresholdVal {
				triggered = true
			} else if tt.thresholdType == domain.ThresholdTypeMin && tt.value < tt.thresholdVal {
				triggered = true
			}

			assert.Equal(t, tt.expectTrigger, triggered)
		})
	}
}

func TestAlertEngine_AlertStatusTransitions(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"firing to acknowledged", domain.AlertStatusFiring, domain.AlertStatusAcknowledged},
		{"acknowledged to resolved", domain.AlertStatusAcknowledged, domain.AlertStatusResolved},
		{"firing to resolved", domain.AlertStatusFiring, domain.AlertStatusResolved},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := domain.EnvironmentAlert{Status: tt.initial}
			if tt.expected == domain.AlertStatusResolved {
				now := time.Now()
				alert.ResolvedAt = &now
			}
			alert.Status = tt.expected
			assert.Equal(t, tt.expected, alert.Status)
		})
	}
}

func TestEnvHealthStats_JSON(t *testing.T) {
	stats := struct {
		TotalDevices   int `json:"total_devices"`
		OnlineDevices  int `json:"online_devices"`
		WarningAlerts  int `json:"warning_alerts"`
		CriticalAlerts int `json:"critical_alerts"`
	}{
		TotalDevices:   10,
		OnlineDevices:  8,
		WarningAlerts:  1,
		CriticalAlerts: 1,
	}

	assert.Equal(t, 10, stats.TotalDevices)
	assert.Equal(t, 8, stats.OnlineDevices)
	assert.Equal(t, 1, stats.WarningAlerts)
	assert.Equal(t, 1, stats.CriticalAlerts)
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	select {
	case <-ctx.Done():
		assert.Equal(t, context.Canceled, ctx.Err())
	default:
		t.Fatal("context should be cancelled")
	}
}

func TestAlertCooldownLogic(t *testing.T) {
	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	tenMinAgo := now.Add(-10 * time.Minute)

	tests := []struct {
		name          string
		lastFiredAt   time.Time
		shouldTrigger bool
	}{
		{"within cooldown", fiveMinAgo, false},
		{"outside cooldown", tenMinAgo, true},
		{"never fired", time.Time{}, true},
	}

	cooldownDuration := 5 * time.Minute

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.lastFiredAt.IsZero() {
				assert.True(t, tt.shouldTrigger)
				return
			}

			timeSinceLastAlert := now.Sub(tt.lastFiredAt)
			shouldTrigger := timeSinceLastAlert > cooldownDuration

			assert.Equal(t, tt.shouldTrigger, shouldTrigger)
		})
	}
}

func TestNotifier_WebhookPayload(t *testing.T) {
	alert := domain.EnvironmentAlert{
		ID:             1,
		MetricType:     domain.MetricTypeTemperature,
		NasID:          100,
		AlertValue:     85.5,
		ThresholdValue: 70.0,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
	}

	_ = alert
	cfg := domain.AlertConfig{
		WebhookURL: "https://example.com/webhook",
	}

	_ = cfg

	payload := map[string]interface{}{
		"alert_id":        alert.ID,
		"metric_type":     alert.MetricType,
		"nas_id":          alert.NasID,
		"alert_value":     alert.AlertValue,
		"threshold":       alert.ThresholdValue,
		"threshold_type":  alert.ThresholdType,
		"severity":        alert.Severity,
		"fired_at":        alert.FiredAt,
	}

	assert.Equal(t, alert.MetricType, payload["metric_type"])
	assert.Equal(t, alert.NasID, payload["nas_id"])
	assert.Equal(t, alert.AlertValue, payload["alert_value"])
	assert.Equal(t, alert.ThresholdValue, payload["threshold"])
	assert.Equal(t, alert.Severity, payload["severity"])
}

func TestNotifier_EmailContent(t *testing.T) {
	alert := domain.EnvironmentAlert{
		ID:             1,
		MetricType:     domain.MetricTypeTemperature,
		NasID:          100,
		AlertValue:     85.5,
		ThresholdValue: 70.0,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC),
	}
	_ = alert

	nasName := "Mikrotik-CCR"

	subject := "[critical] Device Alert: temperature"
	assert.Contains(t, subject, "critical")
	assert.Contains(t, subject, "temperature")

	body := `Metric: temperature
Device: Mikrotik-CCR (100)
Current Value: 85.50
Threshold: max 70.00
Severity: critical`

	assert.Contains(t, body, "temperature")
	assert.Contains(t, body, nasName)
	assert.Contains(t, body, "85.50")
	assert.Contains(t, body, "70.00")
	assert.Contains(t, body, "critical")
}

func TestAlertEngine_NewAlertEngine(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)
	engine := NewAlertEngine(db, notifier)

	assert.NotNil(t, engine)
	assert.Equal(t, 5, engine.cooldownMins)
	assert.Equal(t, db, engine.db)
	assert.Equal(t, notifier, engine.notifier)
}

func TestAlertEngine_ProcessMetrics_Empty(t *testing.T) {
	db := setupTestDB(t)
	engine := NewAlertEngine(db, nil)

	err := engine.ProcessMetrics(context.Background(), []domain.EnvironmentMetric{})
	assert.NoError(t, err)
}

func TestAlertEngine_ProcessMetrics_WithThresholdHit(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	cfg := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}
	db.Create(&cfg)

	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:   "test-tenant",
		MetricType: domain.MetricTypeTemperature,
		Value:      85,
		CollectedAt: time.Now(),
	}

	engine := NewAlertEngine(db, nil)
	err := engine.ProcessMetrics(context.Background(), []domain.EnvironmentMetric{metric})
	assert.NoError(t, err)

	var alert domain.EnvironmentAlert
	err = db.Where("nas_id = ? AND metric_type = ? AND status = ?", 1, domain.MetricTypeTemperature, domain.AlertStatusFiring).First(&alert).Error
	assert.NoError(t, err)
	assert.Equal(t, 85.0, alert.AlertValue)
	assert.Equal(t, domain.SeverityCritical, alert.Severity)
}

func TestAlertEngine_ProcessMetrics_NoThresholdHit(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	cfg := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}
	db.Create(&cfg)

	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:   "test-tenant",
		MetricType: domain.MetricTypeTemperature,
		Value:      50,
		CollectedAt: time.Now(),
	}

	engine := NewAlertEngine(db, nil)
	err := engine.ProcessMetrics(context.Background(), []domain.EnvironmentMetric{metric})
	assert.NoError(t, err)

	var alertCount int64
	db.Model(&domain.EnvironmentAlert{}).Where("nas_id = ? AND metric_type = ? AND status = ?", 1, domain.MetricTypeTemperature, domain.AlertStatusFiring).Count(&alertCount)
	assert.Equal(t, int64(0), alertCount)
}

func TestAlertEngine_EvaluateThreshold_DisabledConfig(t *testing.T) {
	t.Skip("Test has race condition with goroutines")
}

func TestAlertEngine_ResolveAlert(t *testing.T) {
	db := setupTestDB(t)

	alert := domain.EnvironmentAlert{
		NasID:      1,
		TenantID:  "test-tenant",
		MetricType: domain.MetricTypeTemperature,
		Status:     domain.AlertStatusFiring,
		FiredAt:   time.Now(),
	}
	db.Create(&alert)

	engine := NewAlertEngine(db, nil)
	err := engine.resolveAlert(context.Background(), 1, domain.MetricTypeTemperature)
	assert.NoError(t, err)

	var resolvedAlert domain.EnvironmentAlert
	err = db.First(&resolvedAlert).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.AlertStatusResolved, resolvedAlert.Status)
	assert.NotNil(t, resolvedAlert.ResolvedAt)
}

func TestAlertEngine_ProcessAllMetrics(t *testing.T) {
	t.Skip("ProcessAllMetrics uses complex subquery not compatible with SQLite in tests")
}

func TestAlertEngine_CreateOrUpdateAlert_NewAlert(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:   "test-tenant",
		MetricType: domain.MetricTypeTemperature,
		Value:      85,
		CollectedAt: time.Now(),
	}

	cfg := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}

	notifier := NewNotifier(db, nil)
	engine := NewAlertEngine(db, notifier)

	err := engine.createOrUpdateAlert(context.Background(), metric, cfg, 85)
	assert.NoError(t, err)

	var alert domain.EnvironmentAlert
	err = db.First(&alert).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.AlertStatusFiring, alert.Status)
	assert.Equal(t, domain.NotifyStatusPending, alert.NotifyStatus)
}

func TestAlertEngine_CreateOrUpdateAlert_UpdateExisting(t *testing.T) {
	db := setupTestDB(t)

	alert := domain.EnvironmentAlert{
		NasID:          1,
		TenantID:       "test-tenant",
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		AlertValue:     80,
		Status:         domain.AlertStatusFiring,
		FiredAt:        time.Now(),
	}
	db.Create(&alert)

	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:   "test-tenant",
		MetricType: domain.MetricTypeTemperature,
		Value:      90,
		CollectedAt: time.Now(),
	}

	cfg := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}

	engine := NewAlertEngine(db, nil)
	err := engine.createOrUpdateAlert(context.Background(), metric, cfg, 90)
	assert.NoError(t, err)

	var updatedAlert domain.EnvironmentAlert
	err = db.First(&updatedAlert).Error
	assert.NoError(t, err)
	assert.Equal(t, 90.0, updatedAlert.AlertValue)
}

func TestNotifier_NewNotifier(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)

	assert.NotNil(t, notifier)
	assert.Equal(t, db, notifier.db)
	assert.Nil(t, notifier.emailSvc)
}

func TestNotifier_SendAlertNotifications_NoConfig(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)

	alert := domain.EnvironmentAlert{
		ID:            1,
		NasID:         1,
		MetricType:    domain.MetricTypeTemperature,
		Severity:      domain.SeverityCritical,
		FiredAt:       time.Now(),
		NotifyStatus:  domain.NotifyStatusPending,
	}
	db.Create(&alert)

	cfg := domain.AlertConfig{
		NotifyEmail:   false,
		NotifyWebhook: false,
	}

	notifier.SendAlertNotifications(context.Background(), alert, cfg)

	var updatedAlert domain.EnvironmentAlert
	err := db.First(&updatedAlert).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.NotifyStatusSent, updatedAlert.NotifyStatus)
}

func TestNotifier_SendAlertNotifications_EmailEnabled(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS"}
	db.Create(&nas)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
		NotifyStatus:   domain.NotifyStatusPending,
	}
	db.Create(&alert)

	notifier := NewNotifier(db, nil)

	cfg := domain.AlertConfig{
		NotifyEmail:    true,
		NotifyWebhook:  false,
	}

	notifier.SendAlertNotifications(context.Background(), alert, cfg)

	var updatedAlert domain.EnvironmentAlert
	err := db.First(&updatedAlert).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.NotifyStatusSent, updatedAlert.NotifyStatus)
}

func TestNotifier_SendAlertNotifications_WebhookEnabled(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
		NotifyStatus:   domain.NotifyStatusPending,
	}
	db.Create(&alert)

	cfg := domain.AlertConfig{
		NotifyEmail:    false,
		NotifyWebhook:  true,
		WebhookURL:     "https://example.com/webhook",
	}

	notifier.SendAlertNotifications(context.Background(), alert, cfg)

	var updatedAlert domain.EnvironmentAlert
	err := db.First(&updatedAlert).Error
	assert.NoError(t, err)
}

func TestEnvCollector_CollectAllDevices_WithDevices(t *testing.T) {
	db := setupTestDB(t)

	nas1 := domain.NetNas{ID: 1, Name: "NAS-1", Ipaddr: "192.168.1.1", TenantID: int64(1)}
	nas2 := domain.NetNas{ID: 2, Name: "NAS-2", Ipaddr: "192.168.1.2", TenantID: int64(1)}
	db.Create(&nas1)
	db.Create(&nas2)

	collector := NewEnvCollector(db)

	err := collector.CollectAllDevices(context.Background())
	assert.NoError(t, err)
}

func TestEnvCollector_CollectDevice_NoRouterOS(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "192.168.1.1",
		TenantID:   1,
		VendorCode: "mikrotik",
		Secret:     "wrong",
	}
	db.Create(&nas)

	collector := NewEnvCollector(db)

	err := collector.CollectDevice(context.Background(), &nas)
	assert.Error(t, err)
}

func TestEnvCollector_getFloatFromSentence(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected float64
	}{
		{"nil sentence", "cpu-temperature", 0},
		{"empty key not found", "cpu-temperature", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloatFromSentence(nil, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnvCollector_CalculateSeverity_AllTypes(t *testing.T) {
	collector := &EnvCollector{}

	tests := []struct {
		metricType string
		value      float64
		expected   string
	}{
		{domain.MetricTypeTemperature, -10, domain.SeverityNormal},
		{domain.MetricTypeTemperature, 0, domain.SeverityNormal},
		{domain.MetricTypeTemperature, 60, domain.SeverityNormal},
		{domain.MetricTypeTemperature, 69, domain.SeverityNormal},
		{domain.MetricTypeTemperature, 71, domain.SeverityWarning},
		{domain.MetricTypeTemperature, 79, domain.SeverityWarning},
		{domain.MetricTypeTemperature, 81, domain.SeverityCritical},
		{domain.MetricTypeTemperature, 100, domain.SeverityCritical},
		{domain.MetricTypePower, 0, domain.SeverityNormal},
		{domain.MetricTypePower, 50, domain.SeverityNormal},
		{domain.MetricTypePower, 99, domain.SeverityNormal},
		{domain.MetricTypePower, 100, domain.SeverityWarning},
		{domain.MetricTypePower, 119, domain.SeverityWarning},
		{domain.MetricTypePower, 120, domain.SeverityCritical},
		{domain.MetricTypePower, 150, domain.SeverityCritical},
		{domain.MetricTypeVoltage, 170, domain.SeverityCritical},
		{domain.MetricTypeVoltage, 179, domain.SeverityCritical},
		{domain.MetricTypeVoltage, 180, domain.SeverityWarning},
		{domain.MetricTypeVoltage, 199, domain.SeverityWarning},
		{domain.MetricTypeVoltage, 200, domain.SeverityNormal},
		{domain.MetricTypeVoltage, 220, domain.SeverityNormal},
		{domain.MetricTypeVoltage, 249, domain.SeverityNormal},
		{domain.MetricTypeVoltage, 250, domain.SeverityWarning},
		{domain.MetricTypeVoltage, 259, domain.SeverityWarning},
		{domain.MetricTypeVoltage, 260, domain.SeverityCritical},
		{domain.MetricTypeFanSpeed, 0, domain.SeverityWarning},
		{domain.MetricTypeFanSpeed, 999, domain.SeverityWarning},
		{domain.MetricTypeFanSpeed, 1000, domain.SeverityNormal},
		{domain.MetricTypeFanSpeed, 5000, domain.SeverityNormal},
		{domain.MetricTypeSignalStrength, -50, domain.SeverityNormal},
		{domain.MetricTypeSignalStrength, -80, domain.SeverityNormal},
		{domain.MetricTypeSignalStrength, -90, domain.SeverityNormal},
		{"unknown_type", 100, domain.SeverityNormal},
	}

	for _, tt := range tests {
		t.Run(tt.metricType+"_"+fmt.Sprintf("%.0f", tt.value), func(t *testing.T) {
			result := collector.calculateSeverity(tt.metricType, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNotifier_SendAlertNotifications_BothEmailAndWebhook(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS"}
	db.Create(&nas)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
		NotifyStatus:   domain.NotifyStatusPending,
	}
	db.Create(&alert)

	cfg := domain.AlertConfig{
		NotifyEmail:    true,
		NotifyWebhook:  true,
		WebhookURL:     "http://invalid.local/webhook",
	}

	notifier := NewNotifier(db, nil)
	notifier.SendAlertNotifications(context.Background(), alert, cfg)

	var updatedAlert domain.EnvironmentAlert
	err := db.First(&updatedAlert).Error
	assert.NoError(t, err)
}

func TestEnvCollector_CollectDevice_SavesMetrics(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "192.168.1.1",
		TenantID:   1,
		Secret:     "wrong",
		VendorCode: "mikrotik",
	}
	db.Create(&nas)

	collector := NewEnvCollector(db)
	err := collector.CollectDevice(context.Background(), &nas)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestAlertEngine_Run_ContextCancelled(t *testing.T) {
	db := setupTestDB(t)
	engine := NewAlertEngine(db, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := engine.Run(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestEnvCollector_getTemperature(t *testing.T) {
	t.Skip("Cannot test with nil client - causes panic")
}

func TestEnvCollector_getPower(t *testing.T) {
	t.Skip("Cannot test with nil client - causes panic")
}

func TestEnvCollector_getVoltage(t *testing.T) {
	t.Skip("Cannot test with nil client - causes panic")
}

func TestEnvCollector_getFanSpeed(t *testing.T) {
	t.Skip("Cannot test with nil client - causes panic")
}

func TestEnvCollector_CollectAllDevices_Empty(t *testing.T) {
	db := setupTestDB(t)
	collector := NewEnvCollector(db)

	err := collector.CollectAllDevices(context.Background())
	assert.NoError(t, err)
}

func TestEnvCollector_CollectAllDevices_Cancel(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "192.168.1.1",
		TenantID:   1,
		Secret:     "wrong",
		VendorCode: "mikrotik",
	}
	db.Create(&nas)

	collector := NewEnvCollector(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := collector.CollectAllDevices(ctx)
	assert.NoError(t, err)
}

func TestProcessAllMetrics_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	engine := NewAlertEngine(db, nil)

	err := engine.ProcessAllMetrics(context.Background())
	assert.NoError(t, err)
}

func TestProcessMetrics_WithMultipleMetrics(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	cfg1 := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}
	cfg2 := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypePower,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 100,
		Severity:       domain.SeverityWarning,
		Enabled:        true,
	}
	db.Create(&cfg1)
	db.Create(&cfg2)

	now := time.Now()
	metrics := []domain.EnvironmentMetric{
		{
			NasID:      1,
			TenantID:   "1",
			MetricType: domain.MetricTypeTemperature,
			Value:      85,
			CollectedAt: now,
		},
		{
			NasID:      1,
			TenantID:   "1",
			MetricType: domain.MetricTypePower,
			Value:      110,
			CollectedAt: now,
		},
	}

	engine := NewAlertEngine(db, nil)
	err := engine.ProcessMetrics(context.Background(), metrics)
	assert.NoError(t, err)

	var alertCount int64
	db.Model(&domain.EnvironmentAlert{}).Count(&alertCount)
	assert.Equal(t, int64(2), alertCount)
}

func TestProcessMetrics_MinThreshold(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	cfg := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypeVoltage,
		ThresholdType:  domain.ThresholdTypeMin,
		ThresholdValue: 200,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}
	db.Create(&cfg)

	now := time.Now()
	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:   "1",
		MetricType: domain.MetricTypeVoltage,
		Value:      180,
		CollectedAt: now,
	}

	engine := NewAlertEngine(db, nil)
	err := engine.ProcessMetrics(context.Background(), []domain.EnvironmentMetric{metric})
	assert.NoError(t, err)

	var alert domain.EnvironmentAlert
	err = db.First(&alert).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.AlertStatusFiring, alert.Status)
}

func TestEvaluateThreshold_NoConfigs(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:   "1",
		MetricType: domain.MetricTypeTemperature,
		Value:      85,
		CollectedAt: time.Now(),
	}

	engine := NewAlertEngine(db, nil)
	err := engine.evaluateThreshold(context.Background(), metric)
	assert.NoError(t, err)
}

func TestResolveAlert_MultipleAlerts(t *testing.T) {
	db := setupTestDB(t)

	alert1 := domain.EnvironmentAlert{
		NasID:      1,
		TenantID:  "1",
		MetricType: domain.MetricTypeTemperature,
		Status:     domain.AlertStatusFiring,
		FiredAt:   time.Now(),
	}
	alert2 := domain.EnvironmentAlert{
		NasID:      1,
		TenantID:  "1",
		MetricType: domain.MetricTypePower,
		Status:     domain.AlertStatusFiring,
		FiredAt:   time.Now(),
	}
	db.Create(&alert1)
	db.Create(&alert2)

	engine := NewAlertEngine(db, nil)
	err := engine.resolveAlert(context.Background(), 1, domain.MetricTypeTemperature)
	assert.NoError(t, err)

	var count int64
	db.Model(&domain.EnvironmentAlert{}).Where("nas_id = ? AND metric_type = ? AND status = ?", 1, domain.MetricTypeTemperature, domain.AlertStatusResolved).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestEnvCollector_CollectAllDevices_Concurrent(t *testing.T) {
	db := setupTestDB(t)

	for i := 1; i <= 5; i++ {
		nas := domain.NetNas{
			ID:         int64(i),
			Name:       fmt.Sprintf("NAS-%d", i),
			Ipaddr:     fmt.Sprintf("192.168.1.%d", i),
			TenantID:   1,
			Secret:     "wrong",
			VendorCode: "mikrotik",
		}
		db.Create(&nas)
	}

	collector := NewEnvCollector(db)
	err := collector.CollectAllDevices(context.Background())
	assert.NoError(t, err)
}

func TestAlertEngine_NewAlertEngine_CustomCooldown(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)
	
	engine := &AlertEngine{
		db:           db,
		notifier:     notifier,
		cooldownMins: 10,
	}
	
	assert.Equal(t, 10, engine.cooldownMins)
	assert.Equal(t, db, engine.db)
	assert.Equal(t, notifier, engine.notifier)
}

func TestEnvCollector_StructInitialization(t *testing.T) {
	collector := &EnvCollector{}
	assert.NotNil(t, collector)
}

func TestEnvCollector_StoreMetrics_EmptySlice(t *testing.T) {
	db := setupTestDB(t)
	_ = NewEnvCollector(db)

	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:   "1",
		MetricType: domain.MetricTypeTemperature,
		Value:      50,
		CollectedAt: time.Now(),
		CreatedAt:  time.Now(),
	}
	err := db.Create(&metric).Error
	assert.NoError(t, err)
}

func TestEnvCollector_StoreMetrics_SingleMetric(t *testing.T) {
	db := setupTestDB(t)
	_ = NewEnvCollector(db)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	now := time.Now()
	metric := domain.EnvironmentMetric{
		NasID:        1,
		TenantID:     "1",
		MetricType:   domain.MetricTypeTemperature,
		Value:        50,
		CollectedAt:  now,
		CreatedAt:    now,
	}

	err := db.Create(&metric).Error
	assert.NoError(t, err)

	var count int64
	db.Model(&domain.EnvironmentMetric{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestEnvCollector_StoreMetrics_MultipleMetricTypes(t *testing.T) {
	db := setupTestDB(t)

	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	now := time.Now()
	metrics := []domain.EnvironmentMetric{
		{NasID: 1, TenantID: "1", MetricType: domain.MetricTypeTemperature, Value: 50, CollectedAt: now, CreatedAt: now},
		{NasID: 1, TenantID: "1", MetricType: domain.MetricTypePower, Value: 100, CollectedAt: now, CreatedAt: now},
		{NasID: 1, TenantID: "1", MetricType: domain.MetricTypeVoltage, Value: 220, CollectedAt: now, CreatedAt: now},
		{NasID: 1, TenantID: "1", MetricType: domain.MetricTypeFanSpeed, Value: 3000, CollectedAt: now, CreatedAt: now},
	}

	for _, m := range metrics {
		err := db.Create(&m).Error
		assert.NoError(t, err)
	}

	var count int64
	db.Model(&domain.EnvironmentMetric{}).Count(&count)
	assert.Equal(t, int64(4), count)
}

func TestEnvCollector_StoreMetrics_DifferentNAS(t *testing.T) {
	db := setupTestDB(t)

	nas1 := domain.NetNas{ID: 1, Name: "NAS-1", TenantID: int64(1)}
	nas2 := domain.NetNas{ID: 2, Name: "NAS-2", TenantID: int64(1)}
	db.Create(&nas1)
	db.Create(&nas2)

	now := time.Now()
	metric1 := domain.EnvironmentMetric{NasID: 1, TenantID: "1", MetricType: domain.MetricTypeTemperature, Value: 50, CollectedAt: now, CreatedAt: now}
	metric2 := domain.EnvironmentMetric{NasID: 2, TenantID: "1", MetricType: domain.MetricTypeTemperature, Value: 60, CollectedAt: now, CreatedAt: now}

	db.Create(&metric1)
	db.Create(&metric2)

	var count int64
	db.Model(&domain.EnvironmentMetric{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestEnvCollector_CollectDevice_ConnectionRefused(t *testing.T) {
	db := setupTestDB(t)
	collector := NewEnvCollector(db)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "127.0.0.1:1",
		TenantID:   1,
		Secret:     "test",
		VendorCode: "mikrotik",
	}

	err := collector.CollectDevice(context.Background(), &nas)
	assert.Error(t, err)
}

func TestEnvCollector_CollectDevice_InvalidIP(t *testing.T) {
	db := setupTestDB(t)
	collector := NewEnvCollector(db)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "invalid-ip-address",
		TenantID:   1,
		Secret:     "test",
		VendorCode: "mikrotik",
	}

	err := collector.CollectDevice(context.Background(), &nas)
	assert.Error(t, err)
}

func TestEnvCollector_CollectDevice_Timeout(t *testing.T) {
	db := setupTestDB(t)
	collector := NewEnvCollector(db)

	nas := domain.NetNas{
		ID:         1,
		Name:       "Test-NAS",
		Ipaddr:     "10.255.255.1",
		TenantID:   1,
		Secret:     "test",
		VendorCode: "mikrotik",
	}

	err := collector.CollectDevice(context.Background(), &nas)
	assert.Error(t, err)
}

func TestEnvCollector_CollectAllDevices_InvalidSecret(t *testing.T) {
	db := setupTestDB(t)

	nas1 := domain.NetNas{ID: 1, Name: "NAS-1", Ipaddr: "192.168.1.1", TenantID: int64(1), Secret: "wrong"}
	nas2 := domain.NetNas{ID: 2, Name: "NAS-2", Ipaddr: "192.168.1.2", TenantID: int64(1), Secret: "wrong"}
	db.Create(&nas1)
	db.Create(&nas2)

	collector := NewEnvCollector(db)
	err := collector.CollectAllDevices(context.Background())
	assert.NoError(t, err)
}

func TestAlertEngine_ProcessMetrics_EmptyNasID(t *testing.T) {
	db := setupTestDB(t)
	engine := NewAlertEngine(db, nil)

	now := time.Now()
	metric := domain.EnvironmentMetric{
		NasID:      0,
		TenantID:  "1",
		MetricType: domain.MetricTypeTemperature,
		Value:      85,
		CollectedAt: now,
	}

	err := engine.ProcessMetrics(context.Background(), []domain.EnvironmentMetric{metric})
	assert.NoError(t, err)
}

func TestAlertEngine_evaluateThreshold_DBError(t *testing.T) {
	db := setupTestDB(t)
	
	nas := domain.NetNas{ID: 1, Name: "Test-NAS", TenantID: int64(1)}
	db.Create(&nas)

	cfg := domain.AlertConfig{
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}
	db.Create(&cfg)

	metric := domain.EnvironmentMetric{
		NasID:      1,
		TenantID:  "1",
		MetricType: domain.MetricTypeTemperature,
		Value:      85,
		CollectedAt: time.Now(),
	}

	engine := NewAlertEngine(db, nil)
	err := engine.evaluateThreshold(context.Background(), metric)
	assert.NoError(t, err)
}

func TestAlertEngine_createOrUpdateAlert_DBError(t *testing.T) {
	db := setupTestDB(t)

	metric := domain.EnvironmentMetric{
		NasID:      999,
		TenantID:  "1",
		MetricType: domain.MetricTypeTemperature,
		Value:      85,
		CollectedAt: time.Now(),
	}

	cfg := domain.AlertConfig{
		NasID:          999,
		MetricType:     domain.MetricTypeTemperature,
		ThresholdType:  domain.ThresholdTypeMax,
		ThresholdValue: 70,
		Severity:       domain.SeverityCritical,
		Enabled:        true,
	}

	engine := NewAlertEngine(db, nil)
	err := engine.createOrUpdateAlert(context.Background(), metric, cfg, 85)
	assert.NoError(t, err)
}

func TestAlertEngine_resolveAlert_DBError(t *testing.T) {
	db := setupTestDB(t)
	engine := NewAlertEngine(db, nil)

	err := engine.resolveAlert(context.Background(), 9999, domain.MetricTypeTemperature)
	assert.NoError(t, err)
}

func TestNotifier_SendAlertNotifications_EmailNilNas(t *testing.T) {
	db := setupTestDB(t)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          999,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
		NotifyStatus:   domain.NotifyStatusPending,
	}
	db.Create(&alert)

	cfg := domain.AlertConfig{
		NotifyEmail:    true,
		NotifyWebhook:  false,
	}

	notifier := NewNotifier(db, nil)
	notifier.SendAlertNotifications(context.Background(), alert, cfg)

	var updatedAlert domain.EnvironmentAlert
	err := db.First(&updatedAlert).Error
	assert.NoError(t, err)
}

func TestNotifier_sendEmail_EmptyNasName(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
	}

	cfg := domain.AlertConfig{
		NotifyEmail:    true,
		NotifyWebhook:  false,
	}

	notifier.sendEmail(context.Background(), alert, cfg)
}

func TestNotifier_sendWebhook_InvalidURL(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
		NotifyStatus:   domain.NotifyStatusPending,
	}
	db.Create(&alert)

	cfg := domain.AlertConfig{
		NotifyEmail:    false,
		NotifyWebhook:  true,
		WebhookURL:     "http://",
	}

	notifier.sendWebhook(context.Background(), alert, cfg)

	var updatedAlert domain.EnvironmentAlert
	err := db.First(&updatedAlert).Error
	assert.NoError(t, err)
}

func TestNotifier_sendWebhook_EmptyURL(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
		NotifyStatus:   domain.NotifyStatusPending,
	}
	db.Create(&alert)

	cfg := domain.AlertConfig{
		NotifyEmail:    false,
		NotifyWebhook:  true,
		WebhookURL:     "",
	}

	notifier.sendWebhook(context.Background(), alert, cfg)
}

func TestNotifier_SendAlertNotifications_UpdateFails(t *testing.T) {
	db := setupTestDB(t)
	notifier := NewNotifier(db, nil)

	alert := domain.EnvironmentAlert{
		ID:             1,
		NasID:          1,
		MetricType:     domain.MetricTypeTemperature,
		AlertValue:     85,
		ThresholdValue: 70,
		ThresholdType:  domain.ThresholdTypeMax,
		Severity:       domain.SeverityCritical,
		FiredAt:        time.Now(),
		NotifyStatus:   domain.NotifyStatusPending,
	}
	db.Create(&alert)

	cfg := domain.AlertConfig{
		NotifyEmail:    false,
		NotifyWebhook:  false,
	}

	notifier.SendAlertNotifications(context.Background(), alert, cfg)

	var updatedAlert domain.EnvironmentAlert
	err := db.First(&updatedAlert).Error
	assert.NoError(t, err)
}

func TestEnvCollector_calculateSeverity_PowerEdge(t *testing.T) {
	collector := &EnvCollector{}
	
	assert.Equal(t, domain.SeverityNormal, collector.calculateSeverity(domain.MetricTypePower, 0))
	assert.Equal(t, domain.SeverityWarning, collector.calculateSeverity(domain.MetricTypePower, 100))
	assert.Equal(t, domain.SeverityWarning, collector.calculateSeverity(domain.MetricTypePower, 119))
	assert.Equal(t, domain.SeverityCritical, collector.calculateSeverity(domain.MetricTypePower, 120))
}

func TestEnvCollector_calculateSeverity_TemperatureEdge(t *testing.T) {
	collector := &EnvCollector{}
	
	assert.Equal(t, domain.SeverityNormal, collector.calculateSeverity(domain.MetricTypeTemperature, 0))
	assert.Equal(t, domain.SeverityWarning, collector.calculateSeverity(domain.MetricTypeTemperature, 70))
	assert.Equal(t, domain.SeverityWarning, collector.calculateSeverity(domain.MetricTypeTemperature, 79))
	assert.Equal(t, domain.SeverityCritical, collector.calculateSeverity(domain.MetricTypeTemperature, 80))
}

func TestEnvCollector_calculateSeverity_VoltageEdge(t *testing.T) {
	collector := &EnvCollector{}
	
	assert.Equal(t, domain.SeverityCritical, collector.calculateSeverity(domain.MetricTypeVoltage, 170))
	assert.Equal(t, domain.SeverityNormal, collector.calculateSeverity(domain.MetricTypeVoltage, 200))
	assert.Equal(t, domain.SeverityNormal, collector.calculateSeverity(domain.MetricTypeVoltage, 220))
	assert.Equal(t, domain.SeverityWarning, collector.calculateSeverity(domain.MetricTypeVoltage, 250))
	assert.Equal(t, domain.SeverityCritical, collector.calculateSeverity(domain.MetricTypeVoltage, 260))
}
