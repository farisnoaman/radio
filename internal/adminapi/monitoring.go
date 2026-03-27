package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/repository"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"gorm.io/gorm"
)

// registerMonitoringRoutes registers monitoring routes
func registerMonitoringRoutes() {
	// Provider routes (tenant-isolated)
	webserver.ApiGET("/monitoring/metrics", GetMonitoringMetrics)
	webserver.ApiGET("/monitoring/devices", GetDeviceHealth)
	webserver.ApiGET("/monitoring/sessions", GetSessionMetrics)

	// Admin routes (aggregated)
	webserver.ApiGET("/admin/monitoring/metrics", GetAggregatedMetrics)
	webserver.ApiGET("/admin/monitoring/provider/:id", GetProviderMetrics)
}

// GetMonitoringMetrics returns metrics for current tenant only
func GetMonitoringMetrics(c echo.Context) error {
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	db := GetDB(c)

	// Get user count
	var userCount int64
	db.Table("radius_user").Where("tenant_id = ?", tenantID).Count(&userCount)

	// Get online session count
	var sessionCount int64
	db.Table("radius_online").Where("tenant_id = ?", tenantID).Count(&sessionCount)

	// Get device count
	var deviceCount int64
	db.Table("net_nas").Where("tenant_id = ?", tenantID).Count(&deviceCount)

	// Get RADIUS stats for this tenant
	type RadiusStats struct {
		AuthSuccess int64
		AuthFailure int64
	}
	var radiusStats RadiusStats
	db.Raw(`
		SELECT COALESCE(SUM(auth_success), 0) as auth_success,
		       COALESCE(SUM(auth_failure), 0) as auth_failure
		FROM mst_provider_usage
		WHERE tenant_id = ?
	`, tenantID).Scan(&radiusStats)

	metrics := map[string]interface{}{
		"tenant_id":          tenantID,
		"total_users":        userCount,
		"online_sessions":    sessionCount,
		"total_devices":      deviceCount,
		"auth_success_total": radiusStats.AuthSuccess,
		"auth_failure_total": radiusStats.AuthFailure,
	}

	return ok(c, metrics)
}

// GetDeviceHealth returns device health for current tenant
func GetDeviceHealth(c echo.Context) error {
	db := GetDB(c)

	// Query devices with tenant isolation
	var devices []domain.Server
	err := db.Scopes(repository.TenantScope).Find(&devices).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch devices", err)
	}

	return ok(c, devices)
}

// GetSessionMetrics returns session metrics for current tenant
func GetSessionMetrics(c echo.Context) error {
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	db := GetDB(c)

	// Get active sessions for this tenant
	var sessions []domain.RadiusOnline
	err = db.Where("tenant_id = ?", tenantID).Find(&sessions).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch sessions", err)
	}

	// Get session statistics
	var stats struct {
		TotalSessions int64   `json:"total_sessions"`
		TotalInput    int64   `json:"total_input_bytes"`
		TotalOutput   int64   `json:"total_output_bytes"`
	}

	db.Table("radius_online").
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalSessions)

	db.Table("radius_online").
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(input_octets), 0) as total_input, COALESCE(SUM(output_octets), 0) as total_output").
		Scan(&stats)

	return ok(c, map[string]interface{}{
		"sessions": sessions,
		"stats":    stats,
	})
}

// GetAggregatedMetrics returns metrics for all providers (admin only)
func GetAggregatedMetrics(c echo.Context) error {
	// Verify platform admin
	if !IsPlatformAdmin(c) {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
	}

	db := GetDB(c)

	// Aggregate metrics across all tenants
	var providers []domain.Provider
	err := db.Find(&providers).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch providers", err)
	}

	var providerStats []map[string]interface{}
	for _, provider := range providers {
		stats := map[string]interface{}{
			"tenant_id":     provider.ID,
			"provider_code": provider.Code,
			"provider_name": provider.Name,
			"status":        provider.Status,
			"users":         getCount(db, "radius_user", provider.ID),
			"sessions":      getCount(db, "radius_online", provider.ID),
			"devices":       getCount(db, "net_nas", provider.ID),
		}
		providerStats = append(providerStats, stats)
	}

	// Calculate totals
	var totalUsers, totalSessions, totalDevices int64
	for _, stat := range providerStats {
		totalUsers += stat["users"].(int64)
		totalSessions += stat["sessions"].(int64)
		totalDevices += stat["devices"].(int64)
	}

	return ok(c, map[string]interface{}{
		"providers": providerStats,
		"totals": map[string]interface{}{
			"total_users":    totalUsers,
			"total_sessions": totalSessions,
			"total_devices":  totalDevices,
		},
	})
}

// GetProviderMetrics returns detailed metrics for a specific provider (admin only)
func GetProviderMetrics(c echo.Context) error {
	// Verify platform admin
	if !IsPlatformAdmin(c) {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
	}

	// Parse provider ID
	providerIDStr := c.Param("id")
	providerID, err := strconv.ParseInt(providerIDStr, 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", err)
	}

	db := GetDB(c)

	// Get provider details
	var provider domain.Provider
	err = db.First(&provider, providerID).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Provider not found", err)
	}

	// Get provider quota
	var quota domain.ProviderQuota
	err = db.Where("tenant_id = ?", providerID).First(&quota).Error
	if err != nil {
		// Return defaults if no quota set
		quota = domain.ProviderQuota{
			MaxUsers:       1000,
			MaxOnlineUsers: 500,
			MaxNAS:         100,
		}
	}

	// Get current usage
	var usage domain.ProviderUsage
	err = db.Where("tenant_id = ?", providerID).First(&usage).Error
	if err != nil {
		// Initialize usage if not exists
		usage = domain.ProviderUsage{}
	}

	// Calculate usage percentages
	userPercent := 0.0
	if quota.MaxUsers > 0 {
		userPercent = float64(usage.CurrentUsers) / float64(quota.MaxUsers) * 100
	}
	sessionPercent := 0.0
	if quota.MaxOnlineUsers > 0 {
		sessionPercent = float64(usage.CurrentOnlineUsers) / float64(quota.MaxOnlineUsers) * 100
	}

	return ok(c, map[string]interface{}{
		"provider": provider,
		"quota":    quota,
		"usage":    usage,
		"utilization": map[string]interface{}{
			"users_percent":    userPercent,
			"sessions_percent": sessionPercent,
		},
	})
}

// getCount returns the count of records in a table for a specific tenant
func getCount(db *gorm.DB, table string, tenantID int64) int64 {
	var count int64
	db.Table(table).Where("tenant_id = ?", tenantID).Count(&count)
	return count
}

// IsPlatformAdmin checks if the current operator is a platform admin
func IsPlatformAdmin(c echo.Context) bool {
	opr := GetOperator(c)
	if opr == nil {
		return false
	}
	// Platform admin has tenant_id = 0 or a specific admin role
	return opr.TenantID == 0 || opr.Level == "superadmin"
}
