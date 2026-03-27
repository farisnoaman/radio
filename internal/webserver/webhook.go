package webserver

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebhookHandler struct {
	db *gorm.DB
}

func NewWebhookHandler(db *gorm.DB) *WebhookHandler {
	return &WebhookHandler{db: db}
}

func (h *WebhookHandler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")
	api.POST("/webhooks/device-metrics", h.HandleDeviceMetrics)
	api.POST("/webhooks/device-status", h.HandleDeviceStatus)
}

type DeviceMetricsPayload struct {
	DeviceIP  string        `json:"device_ip"`
	DeviceMAC string        `json:"device_mac"`
	Metrics   []MetricEntry `json:"metrics"`
	Timestamp string        `json:"timestamp"`
}

type MetricEntry struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type DeviceStatusPayload struct {
	DeviceIP  string `json:"device_ip"`
	DeviceMAC string `json:"device_mac"`
	Status    string `json:"status"`
	Online    bool   `json:"online"`
	LatencyMs int    `json:"latency_ms"`
	Timestamp string `json:"timestamp"`
}

func (h *WebhookHandler) HandleDeviceMetrics(c echo.Context) error {
	var payload DeviceMetricsPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	if payload.DeviceIP == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "device_ip required"})
	}

	device, err := h.findDeviceByIP(payload.DeviceIP)
	if err != nil {
		zap.S().Warnw("Webhook: device not found", "ip", payload.DeviceIP)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "device not found"})
	}

	now := time.Now()
	for _, m := range payload.Metrics {
		metric := domain.NetworkDeviceMetric{
			DeviceID:    device.ID,
			MetricType:  m.Type,
			Value:       m.Value,
			Unit:        m.Unit,
			Severity:    h.calculateSeverity(m.Type, m.Value),
			CollectedAt: now,
		}
		h.db.Create(&metric)
	}

	h.db.Model(&device).Updates(map[string]interface{}{
		"status":     domain.DeviceStatusOnline,
		"last_seen":  now,
		"updated_at": now,
	})

	zap.S().Debugw("Received device metrics",
		"device_id", device.ID,
		"ip", payload.DeviceIP,
		"metrics_count", len(payload.Metrics))

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *WebhookHandler) HandleDeviceStatus(c echo.Context) error {
	var payload DeviceStatusPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	if payload.DeviceIP == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "device_ip required"})
	}

	device, err := h.findDeviceByIP(payload.DeviceIP)
	if err != nil {
		zap.S().Warnw("Webhook: device not found", "ip", payload.DeviceIP)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "device not found"})
	}

	now := time.Now()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	if payload.Online {
		updates["status"] = domain.DeviceStatusOnline
		updates["last_seen"] = now
		if device.Status != domain.DeviceStatusOnline {
			updates["last_online"] = now
		}
	} else {
		updates["status"] = domain.DeviceStatusOffline
		if device.Status == domain.DeviceStatusOnline {
			updates["last_offline"] = now
			h.createOfflineAlert(device)
		}
	}

	h.db.Model(&device).Updates(updates)

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *WebhookHandler) findDeviceByIP(ip string) (*domain.NetworkDevice, error) {
	var device domain.NetworkDevice
	err := h.db.Where("ip_address = ?", ip).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (h *WebhookHandler) calculateSeverity(metricType string, value float64) string {
	switch metricType {
	case domain.MetricTypeCPU:
		if value >= 90 {
			return domain.DevSeverityCritical
		}
		if value >= 70 {
			return domain.DevSeverityWarning
		}
	case domain.MetricTypeMemory:
		if value >= 90 {
			return domain.DevSeverityCritical
		}
		if value >= 80 {
			return domain.DevSeverityWarning
		}
	case domain.MetricTypeDeviceTemp:
		if value >= 70 {
			return domain.DevSeverityCritical
		}
		if value >= 50 {
			return domain.DevSeverityWarning
		}
	case domain.MetricTypeDeviceVolt:
		if value <= 20 || value >= 26 {
			return domain.DevSeverityCritical
		}
		if value <= 22 || value >= 25 {
			return domain.DevSeverityWarning
		}
	}
	return domain.DevSeverityNormal
}

func (h *WebhookHandler) createOfflineAlert(device *domain.NetworkDevice) {
	var existing domain.NetworkDeviceAlert
	h.db.Where("device_id = ? AND status = ? AND alert_type = ?",
		device.ID, domain.DevAlertStatusActive, domain.AlertTypeOffline).First(&existing)

	if existing.ID > 0 {
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
	h.db.Create(&alert)
}

type MetricsAPIHandler struct {
	db *gorm.DB
}

func NewMetricsAPIHandler(db *gorm.DB) *MetricsAPIHandler {
	return &MetricsAPIHandler{db: db}
}

func (h *MetricsAPIHandler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")

	api.GET("/devices/:id/live-metrics", h.GetLiveMetrics)
	api.GET("/devices/:id/latest-metric", h.GetLatestMetric)
	api.GET("/tenants/:tenantId/alerts", h.GetTenantAlerts)
	api.PUT("/alerts/:id/acknowledge", h.AcknowledgeAlert)
	api.PUT("/alerts/:id/resolve", h.ResolveAlert)
}

func (h *MetricsAPIHandler) GetLiveMetrics(c echo.Context) error {
	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid device id"})
	}

	var device domain.NetworkDevice
	if err := h.db.First(&device, deviceID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "device not found"})
	}

	metricType := c.QueryParam("type")
	from := c.QueryParam("from")
	to := c.QueryParam("to")
	limit := 100

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	query := h.db.Model(&domain.NetworkDeviceMetric{}).Where("device_id = ?", deviceID)

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}
	if from != "" {
		query = query.Where("collected_at >= ?", from)
	}
	if to != "" {
		query = query.Where("collected_at <= ?", to)
	}

	var metrics []domain.NetworkDeviceMetric
	query.Order("collected_at DESC").Limit(limit).Find(&metrics)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  metrics,
		"count": len(metrics),
	})
}

func (h *MetricsAPIHandler) GetLatestMetric(c echo.Context) error {
	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid device id"})
	}

	metricType := c.QueryParam("type")

	query := h.db.Model(&domain.NetworkDeviceMetric{}).Where("device_id = ?", deviceID)
	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}

	var metric domain.NetworkDeviceMetric
	if err := query.Order("collected_at DESC").First(&metric).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "no metrics found"})
	}

	return c.JSON(http.StatusOK, metric)
}

func (h *MetricsAPIHandler) GetTenantAlerts(c echo.Context) error {
	tenantID, err := strconv.ParseInt(c.Param("tenantId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant id"})
	}

	status := c.QueryParam("status")
	severity := c.QueryParam("severity")
	page := 1
	pageSize := 50

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	query := h.db.Model(&domain.NetworkDeviceAlert{}).Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	var total int64
	query.Count(&total)

	var alerts []domain.NetworkDeviceAlert
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&alerts)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":      alerts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *MetricsAPIHandler) AcknowledgeAlert(c echo.Context) error {
	alertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid alert id"})
	}

	now := time.Now()
	result := h.db.Model(&domain.NetworkDeviceAlert{}).Where("id = ?", alertID).Updates(map[string]interface{}{
		"status":          domain.DevAlertStatusAcknowledged,
		"acknowledged_at": now,
	})

	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "alert not found"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *MetricsAPIHandler) ResolveAlert(c echo.Context) error {
	alertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid alert id"})
	}

	now := time.Now()
	result := h.db.Model(&domain.NetworkDeviceAlert{}).Where("id = ?", alertID).Updates(map[string]interface{}{
		"status":      domain.DevAlertStatusResolved,
		"resolved_at": now,
	})

	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "alert not found"})
	}

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func LookupDeviceByIP(db *gorm.DB, ip string) (*domain.NetworkDevice, error) {
	var device domain.NetworkDevice
	err := db.Where("ip_address = ?", ip).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func GetLatestMetrics(db *gorm.DB, deviceID int64, metricTypes []string) (map[string]domain.NetworkDeviceMetric, error) {
	result := make(map[string]domain.NetworkDeviceMetric)

	var metrics []domain.NetworkDeviceMetric
	query := db.Model(&domain.NetworkDeviceMetric{}).Where("device_id = ?", deviceID)

	if len(metricTypes) > 0 {
		query = query.Where("metric_type IN ?", metricTypes)
	}

	if err := query.Order("collected_at DESC").Find(&metrics).Error; err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	for _, m := range metrics {
		if !seen[m.MetricType] {
			result[m.MetricType] = m
			seen[m.MetricType] = true
		}
	}

	return result, nil
}

func PingHost(ip string, timeout time.Duration) (bool, float64) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "443"), timeout)
	if err != nil {
		conn, err = net.DialTimeout("tcp", net.JoinHostPort(ip, "80"), timeout)
		if err != nil {
			return false, 0
		}
	}
	defer conn.Close()
	return true, time.Since(start).Seconds() * 1000
}
