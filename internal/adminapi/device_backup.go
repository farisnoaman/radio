package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/device"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/discovery"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// BackupDeviceConfig triggers an immediate config backup for a device.
// @Summary backup device configuration
// @Tags Device Management
// @Param id path int true "NAS Device ID"
// @Success 200 {object} domain.DeviceConfigBackup
// @Router /api/v1/network/nas/{id}/backup [post]
func BackupDeviceConfig(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}

	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	var nas domain.NetNas
	err = db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&nas).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	backupService := device.NewDeviceBackupService(db)
	record, err := backupService.BackupConfig(c.Request().Context(), &nas, GetOperator(c).Username)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "BACKUP_FAILED", "Failed to start backup", err.Error())
	}

	return ok(c, record)
}

// ListDeviceBackups retrieves backup history for a device.
// @Summary list device backups
// @Tags Device Management
// @Param id path int true "NAS Device ID"
// @Success 200 {object} ListResponse
// @Router /api/v1/network/nas/{id}/backups [get]
func ListDeviceBackups(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}

	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	var backups []domain.DeviceConfigBackup
	err = db.Where("nas_id = ? AND tenant_id = ?", id, tenantID).
		Order("created_at DESC").
		Find(&backups).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch backups", err.Error())
	}

	return ok(c, backups)
}

// RunSpeedTest triggers a speed test on a device.
// @Summary run device speed test
// @Tags Device Management
// @Param id path int true "NAS Device ID"
// @Success 200 {object} device.SpeedTestResult
// @Router /api/v1/network/nas/{id}/speedtest [post]
func RunSpeedTest(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}

	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	var nas domain.NetNas
	err = db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&nas).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	speedTestService := device.NewSpeedTestService(db)
	result, err := speedTestService.RunSpeedTest(c.Request().Context(), &nas, GetOperator(c).Username)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TEST_FAILED", "Failed to start speed test", err.Error())
	}

	return ok(c, result)
}

// GetSpeedTestHistory retrieves speed test history for a device.
// @Summary get speed test history
// @Tags Device Management
// @Param id path int true "NAS Device ID"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {object} ListResponse
// @Router /api/v1/network/nas/{id}/speedtest/history [get]
func GetSpeedTestHistory(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	speedTestService := device.NewSpeedTestService(GetDB(c))
	results, err := speedTestService.GetTestHistory(c.Request().Context(), id, limit)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch history", err.Error())
	}

	return ok(c, results)
}

// DiscoverNeighbors triggers neighbor discovery for a device.
// @Summary discover device neighbors
// @Tags Device Management
// @Param id path int true "NAS Device ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/nas/{id}/neighbors [get]
func DiscoverNeighbors(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid device ID", nil)
	}

	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	var nas domain.NetNas
	err = db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&nas).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Device not found", nil)
	}

	// Get device credentials from query parameters
	username := c.QueryParam("username")
	password := c.QueryParam("password")

	if username == "" || password == "" {
		return fail(c, http.StatusBadRequest, "MISSING_CREDENTIALS", "SSH credentials required", nil)
	}

	scanner, err := discovery.NewScanner(discovery.Config{
		IPRange:  nas.Ipaddr + "/32",
		Timeout:  10 * time.Second,
	})
	if err != nil {
		return fail(c, http.StatusInternalServerError, "SCANNER_ERROR", "Failed to create scanner", err.Error())
	}

	neighbors, err := scanner.DiscoverNeighbors(c.Request().Context(), nas.Ipaddr, username, password)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DISCOVERY_FAILED", "Neighbor discovery failed", err.Error())
	}

	return ok(c, map[string]interface{}{
		"neighbors": neighbors,
		"count":     len(neighbors),
	})
}

// registerDeviceManagementRoutes registers device management routes.
func registerDeviceManagementRoutes() {
	webserver.ApiPOST("/network/nas/:id/backup", BackupDeviceConfig)
	webserver.ApiGET("/network/nas/:id/backups", ListDeviceBackups)
	webserver.ApiPOST("/network/nas/:id/speedtest", RunSpeedTest)
	webserver.ApiGET("/network/nas/:id/speedtest/history", GetSpeedTestHistory)
	webserver.ApiGET("/network/nas/:id/neighbors", DiscoverNeighbors)
}
