package monitoring

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	AuthResultSuccess = "success"
	AuthResultFailure = "failure"
)

type TenantMetricsCollector struct {
	// RADIUS metrics
	radiusAuthRate    *prometheus.CounterVec
	radiusAcctRate    *prometheus.CounterVec
	authErrors        *prometheus.CounterVec
	onlineSessions    *prometheus.GaugeVec

	// Device health metrics
	deviceCpuUsage    *prometheus.GaugeVec
	deviceMemoryUsage *prometheus.GaugeVec
	// deviceUptime - Device uptime in seconds (set by device monitor in Task 2)
	deviceUptime      *prometheus.GaugeVec
	deviceStatus      *prometheus.GaugeVec

	// Network performance metrics
	networkLatency    *prometheus.GaugeVec
	packetLoss        *prometheus.GaugeVec
	bandwidthUsage    *prometheus.GaugeVec
}

func NewTenantMetricsCollector() *TenantMetricsCollector {
	return &TenantMetricsCollector{
		radiusAuthRate: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "radius_auth_total",
				Help: "Total RADIUS authentication requests by tenant",
			},
			[]string{"tenant_id", "result"}, // result: success, failure
		),

		radiusAcctRate: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "radius_acct_total",
				Help: "Total RADIUS accounting requests by tenant",
			},
			[]string{"tenant_id"},
		),

		authErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "radius_auth_errors_total",
				Help: "Total RADIUS authentication errors by tenant and error type",
			},
			[]string{"tenant_id", "error_type"},
		),

		onlineSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "online_sessions",
				Help: "Number of currently online sessions by tenant",
			},
			[]string{"tenant_id"},
		),

		deviceCpuUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "device_cpu_usage_percent",
				Help: "MikroTik device CPU usage percentage",
			},
			[]string{"tenant_id", "device_id", "device_ip"},
		),

		deviceMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "device_memory_usage_percent",
				Help: "MikroTik device memory usage percentage",
			},
			[]string{"tenant_id", "device_id", "device_ip"},
		),

		deviceUptime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "device_uptime_seconds",
				Help: "MikroTik device uptime in seconds",
			},
			[]string{"tenant_id", "device_id", "device_ip"},
		),

		deviceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "device_status",
				Help: "Device status (1=online, 0=offline)",
			},
			[]string{"tenant_id", "device_id", "device_ip"},
		),

		networkLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "network_latency_ms",
				Help: "Network latency to MikroTik device in milliseconds",
			},
			[]string{"tenant_id", "device_id"},
		),

		packetLoss: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "packet_loss_percent",
				Help: "Packet loss percentage to device",
			},
			[]string{"tenant_id", "device_id"},
		),

		bandwidthUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bandwidth_usage_mbps",
				Help: "Current bandwidth usage in Mbps",
			},
			[]string{"tenant_id", "device_id"},
		),
	}
}

// RecordAuth records an authentication attempt with tenant isolation.
// Parameters:
//   - tenantID: The provider/tenant ID
//   - success: true if authentication succeeded, false otherwise
func (m *TenantMetricsCollector) RecordAuth(tenantID int64, success bool) {
	result := AuthResultFailure
	if success {
		result = AuthResultSuccess
	}
	m.radiusAuthRate.WithLabelValues(
		strconv.FormatInt(tenantID, 10),
		result,
	).Inc()
}

// RecordAuthError records an authentication error with tenant isolation.
// Parameters:
//   - tenantID: The provider/tenant ID
//   - errorType: The type of authentication error that occurred
func (m *TenantMetricsCollector) RecordAuthError(tenantID int64, errorType string) {
	m.authErrors.WithLabelValues(
		strconv.FormatInt(tenantID, 10),
		errorType,
	).Inc()
}

// UpdateOnlineSessions updates the current online session count for a tenant.
// Parameters:
//   - tenantID: The provider/tenant ID
//   - count: The number of currently active sessions
func (m *TenantMetricsCollector) UpdateOnlineSessions(tenantID int64, count int) {
	m.onlineSessions.WithLabelValues(
		strconv.FormatInt(tenantID, 10),
	).Set(float64(count))
}

// RecordDeviceHealth records device health metrics with tenant isolation.
// Parameters:
//   - ctx: Context for the operation (currently unused but reserved for future use)
//   - tenantID: The provider/tenant ID
//   - deviceID: Unique identifier for the device
//   - deviceIP: IP address of the device
//   - cpu: CPU usage percentage (0-100)
//   - memory: Memory usage percentage (0-100)
//   - online: Device online status (true=online, false=offline)
func (m *TenantMetricsCollector) RecordDeviceHealth(
	ctx context.Context,
	tenantID int64,
	deviceID, deviceIP string,
	cpu, memory float64,
	online bool,
) {
	// Input validation
	if cpu < 0 || cpu > 100 {
		zap.S().Warn("Invalid CPU percentage", zap.Float64("cpu", cpu))
		return
	}
	if memory < 0 || memory > 100 {
		zap.S().Warn("Invalid memory percentage", zap.Float64("memory", memory))
		return
	}
	if deviceID == "" {
		zap.S().Warn("Empty device ID")
		return
	}

	tenantStr := strconv.FormatInt(tenantID, 10)

	m.deviceCpuUsage.WithLabelValues(tenantStr, deviceID, deviceIP).Set(cpu)
	m.deviceMemoryUsage.WithLabelValues(tenantStr, deviceID, deviceIP).Set(memory)

	status := 0.0
	if online {
		status = 1.0
	}
	m.deviceStatus.WithLabelValues(tenantStr, deviceID, deviceIP).Set(status)
}

// RecordDeviceUptime records device uptime in seconds with tenant isolation.
// Parameters:
//   - ctx: Context for the operation (currently unused but reserved for future use)
//   - tenantID: The provider/tenant ID
//   - deviceID: Unique identifier for the device
//   - deviceIP: IP address of the device
//   - uptime: Uptime in seconds
func (m *TenantMetricsCollector) RecordDeviceUptime(
	ctx context.Context,
	tenantID int64,
	deviceID, deviceIP string,
	uptime float64,
) {
	if deviceID == "" {
		zap.S().Warn("Empty device ID")
		return
	}
	if uptime < 0 {
		zap.S().Warn("Invalid uptime value", zap.Float64("uptime", uptime))
		return
	}

	tenantStr := strconv.FormatInt(tenantID, 10)
	m.deviceUptime.WithLabelValues(tenantStr, deviceID, deviceIP).Set(uptime)
}

// RecordNetworkPerformance records network performance metrics with tenant isolation.
// Parameters:
//   - tenantID: The provider/tenant ID
//   - deviceID: Unique identifier for the device
//   - latency: Network latency in milliseconds
//   - packetLoss: Packet loss percentage
//   - bandwidth: Current bandwidth usage in Mbps
func (m *TenantMetricsCollector) RecordNetworkPerformance(
	tenantID int64,
	deviceID string,
	latency, packetLoss, bandwidth float64,
) {
	tenantStr := strconv.FormatInt(tenantID, 10)

	m.networkLatency.WithLabelValues(tenantStr, deviceID).Set(latency)
	m.packetLoss.WithLabelValues(tenantStr, deviceID).Set(packetLoss)
	m.bandwidthUsage.WithLabelValues(tenantStr, deviceID).Set(bandwidth)
}
