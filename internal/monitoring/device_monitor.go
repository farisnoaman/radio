package monitoring

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/go-routeros/routeros/proto"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeviceHealthMonitor monitors MikroTik device health across all tenants
type DeviceHealthMonitor struct {
	db        *gorm.DB
	metrics   *TenantMetricsCollector
	connPool  map[string]*routeros.Client
	poolMutex sync.RWMutex

	// Performance tuning
	maxConcurrent int
	checkInterval time.Duration
	timeout       time.Duration
}

// NewDeviceHealthMonitor creates a new device health monitor
func NewDeviceHealthMonitor(db *gorm.DB, metrics *TenantMetricsCollector) *DeviceHealthMonitor {
	return &DeviceHealthMonitor{
		db:            db,
		metrics:       metrics,
		connPool:      make(map[string]*routeros.Client),
		maxConcurrent: 50,
		checkInterval: 30 * time.Second,
		timeout:       5 * time.Second,
	}
}

// Run starts the monitoring loop
func (m *DeviceHealthMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAllDevices(ctx)
		}
	}
}

// checkAllDevices checks all devices across all providers efficiently
func (m *DeviceHealthMonitor) checkAllDevices(ctx context.Context) {
	var devices []domain.Server
	result := m.db.Find(&devices)
	if result.Error != nil {
		zap.S().Error("Failed to fetch devices for monitoring", zap.Error(result.Error))
		return
	}

	// Use semaphore to limit concurrent checks
	sem := make(chan struct{}, m.maxConcurrent)
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		sem <- struct{}{}

		go func(d domain.Server) {
			defer wg.Done()
			defer func() { <-sem }()

			deviceCtx, cancel := context.WithTimeout(ctx, m.timeout)
			defer cancel()

			if err := m.checkDevice(deviceCtx, d); err != nil {
				zap.S().Debug("Device check failed",
					zap.String("device", d.Name),
					zap.String("ip", d.PublicIP),
					zap.Error(err))
			}
		}(device)
	}

	wg.Wait()
}

// checkDevice checks a single device's health
func (m *DeviceHealthMonitor) checkDevice(ctx context.Context, device domain.Server) error {
	client, err := m.getDeviceConnection(device)
	if err != nil {
		// Device offline - record offline status
		m.metrics.RecordDeviceHealth(ctx, device.TenantID,
			strconv.FormatInt(device.ID, 10), device.PublicIP, 0, 0, false)

		// Update database to reflect offline status
		m.db.Model(&device).Updates(map[string]interface{}{
			"router_status": "offline",
			"updated_at":    time.Now(),
		})

		zap.S().Debug("Device offline",
			zap.String("device", device.Name),
			zap.String("ip", device.PublicIP),
			zap.Error(err))
		return err
	}

	// Get system resources
	reply, err := client.Run("/system/resource/print")
	if err != nil {
		return fmt.Errorf("failed to get system resources: %w", err)
	}

	if len(reply.Re) > 0 {
		cpu := parseCPU(reply.Re[0])
		memory := parseMemory(reply.Re[0])

		// Record metrics with tenant isolation
		m.metrics.RecordDeviceHealth(ctx, device.TenantID,
			strconv.FormatInt(device.ID, 10), device.PublicIP, cpu, memory, true)

		// Update database
		m.db.Model(&device).Updates(map[string]interface{}{
			"router_status": "online",
			"updated_at":    time.Now(),
		})

		zap.S().Debug("Device health checked",
			zap.String("device", device.Name),
			zap.Int64("tenant_id", device.TenantID),
			zap.Float64("cpu", cpu),
			zap.Float64("memory", memory))
	}

	return nil
}

// getDeviceConnection gets or creates a pooled connection
func (m *DeviceHealthMonitor) getDeviceConnection(device domain.Server) (*routeros.Client, error) {
	m.poolMutex.RLock()
	client, exists := m.connPool[device.PublicIP]
	m.poolMutex.RUnlock()

	if exists {
		return client, nil
	}

	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	// Double-check after acquiring write lock
	if client, exists := m.connPool[device.PublicIP]; exists {
		return client, nil
	}

	address := fmt.Sprintf("%s:%s", device.PublicIP, device.Ports)
	client, err := routeros.Dial(address, device.Username, device.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device %s: %w", address, err)
	}

	m.connPool[device.PublicIP] = client
	return client, nil
}

// Close closes all pooled connections
func (m *DeviceHealthMonitor) Close() {
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	for ip, client := range m.connPool {
		client.Close() // Close() in newer API doesn't return error
		zap.S().Debug("Closed device connection",
			zap.String("ip", ip))
	}

	m.connPool = make(map[string]*routeros.Client)
}

// parseCPU extracts CPU load from MikroTik response
func parseCPU(re *proto.Sentence) float64 {
	// Try to get cpu-load from sentence attributes
	if len(re.List) > 0 {
		for _, attr := range re.List {
			if attr.Key == "cpu-load" {
				var f float64
				fmt.Sscanf(attr.Value, "%f", &f)
				return f
			}
		}
	}
	return 0
}

// parseMemory calculates memory usage percentage from MikroTik response
func parseMemory(re *proto.Sentence) float64 {
	// MikroTik returns memory as "free-memory" and "total-memory"
	// Calculate used memory percentage
	var freeMB, totalMB int64
	hasFree := false
	hasTotal := false

	if len(re.List) > 0 {
		for _, attr := range re.List {
			if attr.Key == "free-memory" {
				fmt.Sscanf(attr.Value, "%d", &freeMB)
				hasFree = true
			}
			if attr.Key == "total-memory" {
				fmt.Sscanf(attr.Value, "%d", &totalMB)
				hasTotal = true
			}
		}
	}

	if hasFree && hasTotal && totalMB > 0 {
		used := totalMB - freeMB
		return float64(used) / float64(totalMB) * 100
	}
	return 0
}

