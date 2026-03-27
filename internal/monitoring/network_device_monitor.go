package monitoring

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NetworkDeviceMonitor struct {
	db            *gorm.DB
	pollInterval  time.Duration
	pingInterval  time.Duration
	maxConcurrent int
	timeout       time.Duration
	snmpTimeout   time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func NewNetworkDeviceMonitor(db *gorm.DB) *NetworkDeviceMonitor {
	return &NetworkDeviceMonitor{
		db:            db,
		pollInterval:  60 * time.Second,
		pingInterval:  30 * time.Second,
		maxConcurrent: 100,
		timeout:       5 * time.Second,
		snmpTimeout:   3 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

func (m *NetworkDeviceMonitor) Start(ctx context.Context) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runMonitoringLoop(ctx)
	}()

	zap.S().Info("Network device monitor started")
}

func (m *NetworkDeviceMonitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	zap.S().Info("Network device monitor stopped")
}

func (m *NetworkDeviceMonitor) runMonitoringLoop(ctx context.Context) {
	pingTicker := time.NewTicker(m.pingInterval)
	defer pingTicker.Stop()

	pollTicker := time.NewTicker(m.pollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			m.checkAllDeviceStatus(ctx)
		case <-pollTicker.C:
			m.pollAllDevices(ctx)
		}
	}
}

func (m *NetworkDeviceMonitor) checkAllDeviceStatus(ctx context.Context) {
	var devices []domain.NetworkDevice
	if err := m.db.Where("polling_enabled = ?", true).Find(&devices).Error; err != nil {
		zap.S().Error("Failed to fetch devices for status check", zap.Error(err))
		return
	}

	sem := make(chan struct{}, m.maxConcurrent)
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		sem <- struct{}{}

		go func(d domain.NetworkDevice) {
			defer wg.Done()
			defer func() { <-sem }()

			m.checkDeviceStatus(ctx, d)
		}(device)
	}

	wg.Wait()
}

func (m *NetworkDeviceMonitor) checkDeviceStatus(ctx context.Context, device domain.NetworkDevice) {
	online, latency := m.pingDevice(device.IPAddress)

	now := time.Now()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	if online {
		updates["status"] = domain.DeviceStatusOnline
		updates["last_seen"] = now
		if device.Status != domain.DeviceStatusOnline {
			updates["last_online"] = now
		}
	} else {
		updates["status"] = domain.DeviceStatusOffline
		if device.Status == domain.DeviceStatusOnline {
			updates["last_offline"] = now
		}

		if device.AlertOnOffline {
			m.createOfflineAlert(ctx, device)
		}
	}

	m.db.Model(&device).Updates(updates)

	zap.S().Debug("Device status checked",
		zap.String("device", device.Name),
		zap.String("ip", device.IPAddress),
		zap.Bool("online", online),
		zap.Float64("latency_ms", latency))
}

func (m *NetworkDeviceMonitor) pingDevice(ip string) (bool, float64) {
	_, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	start := time.Now()
	conn, err := net.DialTimeout("tcp", ip+":443", m.timeout)
	latency := time.Since(start).Seconds() * 1000

	if err != nil {
		conn, err = net.DialTimeout("tcp", ip+":80", m.timeout)
		if err != nil {
			return false, 0
		}
	}
	defer conn.Close()

	return true, latency
}

func (m *NetworkDeviceMonitor) createOfflineAlert(ctx context.Context, device domain.NetworkDevice) {
	var existing domain.NetworkDeviceAlert
	notFound := m.db.Where("device_id = ? AND status = ? AND alert_type = ?",
		device.ID, domain.DevAlertStatusActive, domain.AlertTypeOffline).First(&existing).Error

	if notFound == nil {
		return
	}

	alert := domain.NetworkDeviceAlert{
		DeviceID:  device.ID,
		TenantID:  device.TenantID,
		AlertType: domain.AlertTypeOffline,
		Severity:  domain.DevSeverityCritical,
		Message:   fmt.Sprintf("Device %s (%s) is offline", device.Name, device.IPAddress),
		Status:    domain.DevAlertStatusActive,
	}

	m.db.Create(&alert)

	zap.S().Info("Created offline alert",
		zap.String("device", device.Name),
		zap.Int64("device_id", device.ID))
}

func (m *NetworkDeviceMonitor) pollAllDevices(ctx context.Context) {
	var devices []domain.NetworkDevice
	if err := m.db.Where("polling_enabled = ? AND vendor IN ?", true, []string{
		domain.VendorUbiquiti, domain.VendorTPLink, domain.VendorCisco,
	}).Find(&devices).Error; err != nil {
		zap.S().Error("Failed to fetch devices for polling", zap.Error(err))
		return
	}

	sem := make(chan struct{}, m.maxConcurrent)
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		sem <- struct{}{}

		go func(d domain.NetworkDevice) {
			defer wg.Done()
			defer func() { <-sem }()

			m.pollDeviceMetrics(ctx, d)
		}(device)
	}

	wg.Wait()
}

func (m *NetworkDeviceMonitor) pollDeviceMetrics(ctx context.Context, device domain.NetworkDevice) {
	var metrics []domain.NetworkDeviceMetric

	switch device.Vendor {
	case domain.VendorUbiquiti:
		metrics = m.pollUbiquitiDevice(device)
	case domain.VendorTPLink:
		metrics = m.pollTPLinkDevice(device)
	default:
		metrics = m.pollGenericDevice(device)
	}

	for i := range metrics {
		m.db.Create(&metrics[i])
	}

	if len(metrics) > 0 {
		zap.S().Debug("Polled device metrics",
			zap.String("device", device.Name),
			zap.Int("count", len(metrics)))
	}
}

func (m *NetworkDeviceMonitor) pollUbiquitiDevice(device domain.NetworkDevice) []domain.NetworkDeviceMetric {
	var metrics []domain.NetworkDeviceMetric
	now := time.Now()

	online, _ := m.pingDevice(device.IPAddress)
	if !online {
		return metrics
	}

	metrics = append(metrics, domain.NetworkDeviceMetric{
		DeviceID:    device.ID,
		MetricType:  domain.MetricTypeUptime,
		Value:       float64(time.Since(now).Milliseconds()),
		Unit:        "ms",
		Severity:    domain.DevSeverityNormal,
		CollectedAt: now,
	})

	return metrics
}

func (m *NetworkDeviceMonitor) pollTPLinkDevice(device domain.NetworkDevice) []domain.NetworkDeviceMetric {
	var metrics []domain.NetworkDeviceMetric
	now := time.Now()

	online, _ := m.pingDevice(device.IPAddress)
	if !online {
		return metrics
	}

	metrics = append(metrics, domain.NetworkDeviceMetric{
		DeviceID:    device.ID,
		MetricType:  domain.MetricTypeUptime,
		Value:       float64(time.Since(now).Milliseconds()),
		Unit:        "ms",
		Severity:    domain.DevSeverityNormal,
		CollectedAt: now,
	})

	return metrics
}

func (m *NetworkDeviceMonitor) pollGenericDevice(device domain.NetworkDevice) []domain.NetworkDeviceMetric {
	var metrics []domain.NetworkDeviceMetric
	now := time.Now()

	online, latency := m.pingDevice(device.IPAddress)

	severity := domain.DevSeverityNormal
	if latency > 500 {
		severity = domain.DevSeverityWarning
	}

	metrics = append(metrics, domain.NetworkDeviceMetric{
		DeviceID:    device.ID,
		MetricType:  domain.MetricTypeLatency,
		Value:       latency,
		Unit:        "ms",
		Severity:    severity,
		CollectedAt: now,
	})

	if !online {
		metrics = append(metrics, domain.NetworkDeviceMetric{
			DeviceID:    device.ID,
			MetricType:  domain.MetricTypeUptime,
			Value:       0,
			Unit:        "",
			Severity:    domain.DevSeverityCritical,
			CollectedAt: now,
		})
	}

	return metrics
}

func (m *NetworkDeviceMonitor) CleanupOldMetrics(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := m.db.Where("collected_at < ?", cutoff).Delete(&domain.NetworkDeviceMetric{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", result.Error)
	}

	zap.S().Info("Cleaned up old metrics",
		zap.Int64("deleted", result.RowsAffected),
		zap.Int("retention_days", retentionDays))

	return nil
}
