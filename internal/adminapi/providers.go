package adminapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/migration"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

func registerProviderRoutes() {
	// Admin routes (for /quotas page and platform management)
	webserver.ApiGET("/admin/providers", ListProviders)
	webserver.ApiPOST("/admin/providers", CreateProvider)
	webserver.ApiGET("/admin/providers/:id", GetProvider)
	webserver.ApiPUT("/admin/providers/:id", UpdateProvider)
	webserver.ApiDELETE("/admin/providers/:id", DeleteProvider)
	webserver.ApiGET("/admin/providers/me", GetCurrentProvider)
	webserver.ApiPUT("/admin/providers/me/settings", UpdateCurrentProviderSettings)

	// Legacy routes (for /providers page)
	webserver.ApiGET("/providers", ListProviders)
	webserver.ApiPOST("/providers", CreateProvider)
	webserver.ApiGET("/providers/:id", GetProvider)
	webserver.ApiPUT("/providers/:id", UpdateProvider)
	webserver.ApiDELETE("/providers/:id", DeleteProvider)
	webserver.ApiGET("/providers/me", GetCurrentProvider)
	webserver.ApiPUT("/providers/me/settings", UpdateCurrentProviderSettings)

	// Platform provider admin credentials management
	webserver.ApiGET("/platform/providers/:id/admin-credentials", GetProviderAdminCredentials)
	webserver.ApiPUT("/platform/providers/:id/admin-credentials", UpdateProviderAdminCredentials)
	webserver.ApiPOST("/platform/providers/:id/admin-credentials/reset", ResetProviderAdminCredentials)
}

type ProviderRequest struct {
	Code          string `json:"code" form:"code"`
	Name          string `json:"name" form:"name"`
	Status        string `json:"status" form:"status"`
	MaxUsers      int    `json:"max_users" form:"max_users"`
	MaxNas        int    `json:"max_nas" form:"max_nas"`
	Branding      string `json:"branding" form:"branding"`
	Settings      string `json:"settings" form:"settings"`
	AdminUsername string `json:"admin_username" validate:"omitempty,min=3,max=50,alphanum"`
	AdminPassword string `json:"admin_password" validate:"omitempty,min=6"`
}

type ProviderSettingsRequest struct {
	AllowUserRegistration  bool   `json:"allow_user_registration"`
	AllowVoucherCreation   bool   `json:"allow_voucher_creation"`
	DefaultProductID       int64  `json:"default_product_id"`
	DefaultProfileID       int64  `json:"default_profile_id"`
	AutoExpireSessions     bool   `json:"auto_expire_sessions"`
	SessionTimeout         int    `json:"session_timeout"`
	IdleTimeout            int    `json:"idle_timeout"`
	MaxConcurrentSessions  int    `json:"max_concurrent_sessions"`
}

func ListProviders(c echo.Context) error {
	var providers []*domain.Provider
	var total int64

	db := GetDB(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := db.Model(&domain.Provider{})

	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if search := c.QueryParam("search"); search != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	offset := (page - 1) * perPage
	query.Order("id DESC").Offset(offset).Limit(perPage).Find(&providers)

	// Transform providers to include aliases for frontend compatibility
	var transformedProviders []map[string]interface{}
	for _, p := range providers {
		tp := map[string]interface{}{
			"id":              p.ID,
			"code":            p.Code,
			"provider_code":   p.Code, // Alias
			"name":            p.Name,
			"provider_name":   p.Name, // Alias
			"status":          p.Status,
			"max_users":       p.MaxUsers,
			"max_nas":         p.MaxNas,
			"branding":        p.Branding,
			"settings":        p.Settings,
			"created_at":      p.CreatedAt,
			"updated_at":      p.UpdatedAt,
		}
		// Parse settings and add to response
		if settings, err := p.GetSettings(); err == nil && settings != nil {
			tp["allow_user_registration"] = settings.AllowUserRegistration
			tp["allow_voucher_creation"] = settings.AllowVoucherCreation
			tp["default_product_id"] = settings.DefaultProductID
			tp["default_profile_id"] = settings.DefaultProfileID
			tp["auto_expire_sessions"] = settings.AutoExpireSessions
			tp["session_timeout"] = settings.SessionTimeout
			tp["idle_timeout"] = settings.IdleTimeout
			tp["max_concurrent_sessions"] = settings.MaxConcurrentSessions
		}

		// Get real usage statistics from database
		var totalUsers, activeUsers, onlineSessions, totalNas, activeNas int64
		db.Model(&domain.RadiusUser{}).Where("tenant_id = ?", p.ID).Count(&totalUsers)
		db.Model(&domain.RadiusUser{}).Where("tenant_id = ? AND status = ?", p.ID, common.ENABLED).Count(&activeUsers)
		db.Model(&domain.RadiusOnline{}).Where("tenant_id = ?", p.ID).Count(&onlineSessions)
		db.Model(&domain.NetNas{}).Where("tenant_id = ?", p.ID).Count(&totalNas)
		db.Model(&domain.NetNas{}).Where("tenant_id = ? AND status = ?", p.ID, common.ENABLED).Count(&activeNas)

		// Add usage statistics
		tp["usage"] = map[string]interface{}{
			"current_users":         totalUsers,
			"current_active_users": activeUsers,
			"current_online_users": onlineSessions,
			"current_nas":          totalNas,
			"current_active_nas":   activeNas,
		}

		// Add utilization percentages
		maxUsers := p.MaxUsers
		if maxUsers <= 0 {
			maxUsers = 1 // Avoid division by zero
		}
		maxNas := p.MaxNas
		if maxNas <= 0 {
			maxNas = 1 // Avoid division by zero
		}
		tp["utilization"] = map[string]interface{}{
			"users_percent":   float64(totalUsers) / float64(maxUsers) * 100,
			"sessions_percent": float64(onlineSessions) / float64(maxUsers) * 100,
			"nas_percent":     float64(totalNas) / float64(maxNas) * 100,
		}

		transformedProviders = append(transformedProviders, tp)
	}

	return paged(c, transformedProviders, total, page, perPage)
}

func CreateProvider(c echo.Context) error {
	var req ProviderRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	if req.Code == "" {
		return fail(c, http.StatusBadRequest, "MISSING_CODE", "Provider code is required", nil)
	}
	if req.Name == "" {
		return fail(c, http.StatusBadRequest, "MISSING_NAME", "Provider name is required", nil)
	}
	if req.MaxUsers <= 0 {
		req.MaxUsers = 1000
	}
	if req.MaxNas <= 0 {
		req.MaxNas = 100
	}
	if req.Status == "" {
		req.Status = "active"
	}

	db := GetDB(c)

	provider := &domain.Provider{
		Code:     req.Code,
		Name:     req.Name,
		Status:   req.Status,
		MaxUsers: req.MaxUsers,
		MaxNas:   req.MaxNas,
		Branding: req.Branding,
		Settings: req.Settings,
	}

	if err := db.Create(provider).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create provider: "+err.Error(), nil)
	}

	// Create provider schema (only for PostgreSQL - skip for SQLite)
	migrator := migration.NewSchemaMigrator(db)
	if err := migrator.CreateProviderSchema(provider.ID); err != nil {
		// Log the error but don't fail - schema creation is PostgreSQL-specific
		fmt.Printf("Warning: Failed to create provider schema (this is expected for SQLite): %v\n", err)
	}

	// Create tenant admin with custom or default credentials
	adminUsername := "admin"
	adminPassword := "123456"

	if req.AdminUsername != "" && req.AdminPassword != "" {
		adminUsername = req.AdminUsername
		adminPassword = req.AdminPassword
	}

	opr := &domain.SysOpr{
		TenantID:  provider.ID,
		Realname:  "Provider Administrator",
		Username:  adminUsername,
		Password:  common.Sha256HashWithSalt(adminPassword, common.GetSecretSalt()),
		Level:     "admin",
		Status:    "enabled",
	}
	if err := db.Create(opr).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "ADMIN_CREATE_FAILED", "Failed to create admin", err.Error())
	}

	return ok(c, map[string]interface{}{
		"provider": provider,
		"admin": map[string]interface{}{
			"username": opr.Username,
			"password": adminPassword, // Return plain text password for display
			"level":    opr.Level,
			"status":   opr.Status,
		},
		"message": "Provider created successfully. Please save the admin password now.",
	})
}

func GetProvider(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", nil)
	}

	var provider domain.Provider
	if err := GetDB(c).First(&provider, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Provider not found", nil)
	}

	db := GetDB(c)

	// Parse settings to include in response
	settings, _ := provider.GetSettings()

	// Get real usage statistics from database
	var totalUsers, activeUsers, onlineSessions, totalNas, activeNas int64
	db.Model(&domain.RadiusUser{}).Where("tenant_id = ?", id).Count(&totalUsers)
	db.Model(&domain.RadiusUser{}).Where("tenant_id = ? AND status = ?", id, common.ENABLED).Count(&activeUsers)
	db.Model(&domain.RadiusOnline{}).Where("tenant_id = ?", id).Count(&onlineSessions)
	db.Model(&domain.NetNas{}).Where("tenant_id = ?", id).Count(&totalNas)
	db.Model(&domain.NetNas{}).Where("tenant_id = ? AND status = ?", id, common.ENABLED).Count(&activeNas)

	// Build response with expanded settings
	response := map[string]interface{}{
		"id":             provider.ID,
		"code":           provider.Code,
		"provider_code":  provider.Code, // Alias for frontend compatibility
		"name":           provider.Name,
		"provider_name":  provider.Name, // Alias for frontend compatibility
		"status":         provider.Status,
		"max_users":      provider.MaxUsers,
		"max_nas":        provider.MaxNas,
		"branding":       provider.Branding,
		"settings":       provider.Settings,
		"created_at":     provider.CreatedAt,
		"updated_at":     provider.UpdatedAt,
	}

	// Add settings fields directly to response for easier frontend access
	if settings != nil {
		response["allow_user_registration"] = settings.AllowUserRegistration
		response["allow_voucher_creation"] = settings.AllowVoucherCreation
		response["default_product_id"] = settings.DefaultProductID
		response["default_profile_id"] = settings.DefaultProfileID
		response["auto_expire_sessions"] = settings.AutoExpireSessions
		response["session_timeout"] = settings.SessionTimeout
		response["idle_timeout"] = settings.IdleTimeout
		response["max_concurrent_sessions"] = settings.MaxConcurrentSessions
	}

	// Add usage statistics
	response["usage"] = map[string]interface{}{
		"current_users":         totalUsers,
		"current_active_users": activeUsers,
		"current_online_users": onlineSessions,
		"current_nas":          totalNas,
		"current_active_nas":   activeNas,
	}

	// Add utilization percentages
	maxUsers := provider.MaxUsers
	if maxUsers <= 0 {
		maxUsers = 1 // Avoid division by zero
	}
	maxNas := provider.MaxNas
	if maxNas <= 0 {
		maxNas = 1 // Avoid division by zero
	}
	response["utilization"] = map[string]interface{}{
		"users_percent":    float64(totalUsers) / float64(maxUsers) * 100,
		"sessions_percent": float64(onlineSessions) / float64(maxUsers) * 100,
		"nas_percent":     float64(totalNas) / float64(maxNas) * 100,
	}

	return ok(c, response)
}

func UpdateProvider(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", nil)
	}

	var provider domain.Provider
	if err := GetDB(c).First(&provider, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Provider not found", nil)
	}

	var req ProviderRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.Status != "" {
		provider.Status = req.Status
	}
	if req.MaxUsers > 0 {
		provider.MaxUsers = req.MaxUsers
	}
	if req.MaxNas > 0 {
		provider.MaxNas = req.MaxNas
	}
	if req.Branding != "" {
		provider.Branding = req.Branding
	}
	if req.Settings != "" {
		provider.Settings = req.Settings
	}

	if err := GetDB(c).Save(&provider).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update provider: "+err.Error(), nil)
	}

	return ok(c, provider)
}

func DeleteProvider(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", nil)
	}

	if id == 1 {
		return fail(c, http.StatusBadRequest, "CANNOT_DELETE", "Cannot delete default provider", nil)
	}

	var provider domain.Provider
	if err := GetDB(c).First(&provider, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Provider not found", nil)
	}

	if err := GetDB(c).Delete(&provider).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete provider: "+err.Error(), nil)
	}

	return ok(c, map[string]interface{}{"message": "Provider deleted successfully"})
}

func GetCurrentProvider(c echo.Context) error {
	opr := GetOperator(c)
	if opr == nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", nil)
	}

	var provider domain.Provider
	if err := GetDB(c).First(&provider, opr.TenantID).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Provider not found", nil)
	}

	return ok(c, provider)
}

func UpdateCurrentProviderSettings(c echo.Context) error {
	opr := GetOperator(c)
	if opr == nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", nil)
	}

	var provider domain.Provider
	if err := GetDB(c).First(&provider, opr.TenantID).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Provider not found", nil)
	}

	var req ProviderSettingsRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	settings := &domain.ProviderSettings{
		AllowUserRegistration:  req.AllowUserRegistration,
		AllowVoucherCreation:   req.AllowVoucherCreation,
		DefaultProductID:       req.DefaultProductID,
		DefaultProfileID:       req.DefaultProfileID,
		AutoExpireSessions:     req.AutoExpireSessions,
		SessionTimeout:        req.SessionTimeout,
		IdleTimeout:           req.IdleTimeout,
		MaxConcurrentSessions: req.MaxConcurrentSessions,
	}

	if err := provider.SetSettings(settings); err != nil {
		return fail(c, http.StatusInternalServerError, "SETTINGS_FAILED", "Failed to set settings: "+err.Error(), nil)
	}

	if err := GetDB(c).Save(&provider).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update provider: "+err.Error(), nil)
	}

	return ok(c, provider)
}

func GetProviderStats(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", nil)
	}

	db := GetDB(c)
	stats := &domain.ProviderStats{
		ProviderID: id,
	}

	db.Model(&domain.RadiusUser{}).Where("tenant_id = ?", id).Count(&stats.TotalUsers)
	db.Model(&domain.RadiusUser{}).Where("tenant_id = ? AND status = ?", id, common.ENABLED).Count(&stats.ActiveUsers)
	db.Model(&domain.RadiusOnline{}).Where("tenant_id = ?", id).Count(&stats.OnlineSessions)
	db.Model(&domain.NetNas{}).Where("tenant_id = ?", id).Count(&stats.TotalNas)
	db.Model(&domain.NetNas{}).Where("tenant_id = ? AND status = ?", id, common.ENABLED).Count(&stats.ActiveNas)
	db.Model(&domain.VoucherBatch{}).Where("tenant_id = ?", id).Count(&stats.TotalVouchers)

	return ok(c, stats)
}
