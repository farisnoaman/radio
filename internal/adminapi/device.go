package adminapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/mikrotik"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
)

var deviceTypes = []map[string]string{
	{"id": "router", "name": "Router"},
	{"id": "ap", "name": "Access Point"},
	{"id": "bridge", "name": "Bridge"},
	{"id": "switch", "name": "Switch"},
	{"id": "firewall", "name": "Firewall"},
	{"id": "modem", "name": "Modem"},
	{"id": "other", "name": "Other"},
}

var deviceVendors = []map[string]string{
	{"id": "mikrotik", "name": "MikroTik"},
	{"id": "ubiquiti", "name": "Ubiquiti"},
	{"id": "tplink", "name": "TP-Link"},
	{"id": "cisco", "name": "Cisco"},
	{"id": "huawei", "name": "Huawei"},
	{"id": "other", "name": "Other"},
}

var metricTypes = []map[string]string{
	{"id": "cpu_load", "name": "CPU Load", "unit": "%"},
	{"id": "memory", "name": "Memory", "unit": "%"},
	{"id": "temperature", "name": "Temperature", "unit": "°C"},
	{"id": "voltage", "name": "Voltage", "unit": "V"},
	{"id": "signal", "name": "Signal", "unit": "dBm"},
	{"id": "latency", "name": "Latency", "unit": "ms"},
	{"id": "uptime", "name": "Uptime", "unit": "hours"},
}

func registerDeviceRoutes() {
	webserver.ApiGET("/network/devices", listDevices)
	webserver.ApiPOST("/network/devices", createDevice)
	webserver.ApiGET("/network/devices/:id", getDevice)
	webserver.ApiPUT("/network/devices/:id", updateDevice)
	webserver.ApiDELETE("/network/devices/:id", deleteDevice)
	webserver.ApiPOST("/network/devices/:id/reboot", rebootDevice)
	webserver.ApiGET("/network/devices/:id/metrics", getDeviceMetrics)
	webserver.ApiGET("/network/devices/:id/alerts", getDeviceAlerts)
	webserver.ApiGET("/network/devices/overview", getDevicesOverview)

	webserver.ApiGET("/locations", listLocations)
	webserver.ApiPOST("/locations", createLocation)
	webserver.ApiPUT("/locations/:id", updateLocation)
	webserver.ApiDELETE("/locations/:id", deleteLocation)

	webserver.ApiGET("/devices/types", getDeviceTypes)
	webserver.ApiGET("/devices/vendors", getDeviceVendors)
	webserver.ApiGET("/metrics/types", getMetricTypes)
}

func listDevices(c echo.Context) error {
	db := GetDB(c)
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	page, pageSize := parsePagination(c)
	sortField := c.QueryParam("sort")
	if sortField == "" {
		sortField = "updated_at"
	}
	order := strings.ToUpper(c.QueryParam("order"))
	if order != "ASC" {
		order = "DESC"
	}

	var total int64
	var devices []domain.NetworkDevice
	query := db.Model(&domain.NetworkDevice{}).Where("tenant_id = ?", tenantID)

	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		if strings.EqualFold(db.Name(), "postgres") {
			query = query.Where("name ILIKE ?", "%"+name+"%")
		} else {
			query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
		}
	}
	if deviceType := strings.TrimSpace(c.QueryParam("type")); deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}
	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if vendor := strings.TrimSpace(c.QueryParam("vendor")); vendor != "" {
		query = query.Where("vendor = ?", vendor)
	}
	if loc := strings.TrimSpace(c.QueryParam("location_id")); loc != "" {
		if lid, err := strconv.ParseInt(loc, 10, 64); err == nil {
			query = query.Where("location_id = ?", lid)
		}
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order(sortField + " " + order).Limit(pageSize).Offset(offset).Find(&devices)

	return paged(c, devices, total, page, pageSize)
}

func getDevice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}

	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var device domain.NetworkDevice
	if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}
	return ok(c, device)
}

func createDevice(c echo.Context) error {
	var device domain.NetworkDevice
	if err := c.Bind(&device); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var count int64
	GetDB(c).Model(&domain.NetworkDevice{}).Where("tenant_id = ? AND ip_address = ?", tenantID, device.IPAddress).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "IPADDR_EXISTS", "IP address already exists", nil)
	}

	device.TenantID = tenantID
	if device.Status == "" {
		device.Status = domain.DeviceStatusUnknown
	}
	if device.SNMPPort == 0 {
		device.SNMPPort = 161
	}
	device.PollingEnabled = true
	device.AlertOnOffline = true

	if err := GetDB(c).Create(&device).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create device", err.Error())
	}

	return ok(c, device)
}

func updateDevice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var device domain.NetworkDevice
	if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	delete(updates, "id")
	delete(updates, "tenant_id")
	delete(updates, "created_at")

	if ipAddr, ok := updates["ip_address"].(string); ok && ipAddr != device.IPAddress {
		var count int64
		GetDB(c).Model(&domain.NetworkDevice{}).Where("tenant_id = ? AND ip_address = ? AND id != ?", tenantID, ipAddr, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "IPADDR_EXISTS", "IP address already exists", nil)
		}
	}

	if err := GetDB(c).Model(&device).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update device", err.Error())
	}

	return ok(c, device)
}

func deleteDevice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	result := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&domain.NetworkDevice{})
	if result.Error != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete device", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	return ok(c, map[string]interface{}{"success": true})
}

func getDeviceMetrics(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var device domain.NetworkDevice
	if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	page, pageSize := parsePagination(c)
	metricType := c.QueryParam("type")
	from := c.QueryParam("from")
	to := c.QueryParam("to")

	var total int64
	var metrics []domain.NetworkDeviceMetric
	query := GetDB(c).Model(&domain.NetworkDeviceMetric{}).Where("device_id = ?", id)

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}
	if from != "" {
		query = query.Where("collected_at >= ?", from)
	}
	if to != "" {
		query = query.Where("collected_at <= ?", to)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order("collected_at DESC").Limit(pageSize).Offset(offset).Find(&metrics)

	return paged(c, metrics, total, page, pageSize)
}

func getDeviceAlerts(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var device domain.NetworkDevice
	if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	page, pageSize := parsePagination(c)
	status := c.QueryParam("status")

	var total int64
	var alerts []domain.NetworkDeviceAlert
	query := GetDB(c).Model(&domain.NetworkDeviceAlert{}).Where("device_id = ?", id)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&alerts)

	return paged(c, alerts, total, page, pageSize)
}

func getDevicesOverview(c echo.Context) error {
	db := GetDB(c)
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var stats struct {
		Total   int64 `json:"total"`
		Online  int64 `json:"online"`
		Offline int64 `json:"offline"`
		Unknown int64 `json:"unknown"`
	}

	base := db.Model(&domain.NetworkDevice{}).Where("tenant_id = ?", tenantID)
	base.Count(&stats.Total)
	base.Where("status = ?", domain.DeviceStatusOnline).Count(&stats.Online)
	base.Where("status = ?", domain.DeviceStatusOffline).Count(&stats.Offline)
	base.Where("status = ?", domain.DeviceStatusUnknown).Count(&stats.Unknown)

	return ok(c, stats)
}

func listLocations(c echo.Context) error {
	db := GetDB(c)
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	page, pageSize := parsePagination(c)

	var total int64
	var locations []domain.Location
	query := db.Model(&domain.Location{}).Where("tenant_id = ?", tenantID)

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order("name ASC").Limit(pageSize).Offset(offset).Find(&locations)

	return paged(c, locations, total, page, pageSize)
}

func createLocation(c echo.Context) error {
	var location domain.Location
	if err := c.Bind(&location); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)
	location.TenantID = tenantID

	if err := GetDB(c).Create(&location).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create location", err.Error())
	}

	return ok(c, location)
}

func updateLocation(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid location ID", nil)
	}
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var location domain.Location
	if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&location).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Location not found", nil)
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	delete(updates, "id")
	delete(updates, "tenant_id")
	delete(updates, "created_at")

	if err := GetDB(c).Model(&location).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update location", err.Error())
	}

	return ok(c, location)
}

func deleteLocation(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid location ID", nil)
	}
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	result := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&domain.Location{})
	if result.Error != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete location", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Location not found", nil)
	}

	return ok(c, map[string]interface{}{"success": true})
}

func getDeviceTypes(c echo.Context) error {
	return ok(c, deviceTypes)
}

func getDeviceVendors(c echo.Context) error {
	return ok(c, deviceVendors)
}

func getMetricTypes(c echo.Context) error {
	return ok(c, metricTypes)
}

func rebootDevice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}
	tenantID, _ := strconv.ParseInt(GetTenantID(c), 10, 64)

	var device domain.NetworkDevice
	if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	if device.Vendor != domain.VendorMikroTik {
		return fail(c, http.StatusBadRequest, "UNSUPPORTED_DEVICE", "Reboot only supported for MikroTik devices", nil)
	}

	host := device.APIEndpoint
	if host == "" {
		host = device.IPAddress
	}
	if host == "" {
		return fail(c, http.StatusBadRequest, "NO_ENDPOINT", "Device has no API endpoint or IP address configured", nil)
	}

	if device.APIUsername == "" || device.APIPassword == "" {
		return fail(c, http.StatusBadRequest, "NO_CREDENTIALS", "Device has no API credentials configured", nil)
	}

	client := mikrotik.NewClient(host, device.APIUsername, device.APIPassword)

	zap.S().Infow("Initiating device reboot",
		"device_id", device.ID,
		"device_name", device.Name,
		"host", host,
		"namespace", "adminapi")

	if err := client.Reboot(); err != nil {
		zap.S().Errorw("Failed to reboot device",
			"device_id", device.ID,
			"device_name", device.Name,
			"host", host,
			"error", err.Error(),
			"namespace", "adminapi")
		return fail(c, http.StatusInternalServerError, "REBOOT_FAILED", fmt.Sprintf("Failed to reboot device: %s", err.Error()), nil)
	}

	zap.S().Infow("Device reboot command sent successfully",
		"device_id", device.ID,
		"device_name", device.Name,
		"namespace", "adminapi")

	return ok(c, map[string]interface{}{
		"success":     true,
		"message":     "Reboot command sent to device",
		"device_id":   device.ID,
		"device_name": device.Name,
	})
}
