package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

func registerProviderRoutes() {
	webserver.ApiGET("/providers", ListProviders)
	webserver.ApiPOST("/providers", CreateProvider)
	webserver.ApiGET("/providers/:id", GetProvider)
	webserver.ApiPUT("/providers/:id", UpdateProvider)
	webserver.ApiDELETE("/providers/:id", DeleteProvider)
	webserver.ApiGET("/providers/me", GetCurrentProvider)
	webserver.ApiPUT("/providers/me/settings", UpdateCurrentProviderSettings)
}

type ProviderRequest struct {
	Code     string `json:"code" form:"code"`
	Name     string `json:"name" form:"name"`
	Status   string `json:"status" form:"status"`
	MaxUsers int    `json:"max_users" form:"max_users"`
	MaxNas   int    `json:"max_nas" form:"max_nas"`
	Branding string `json:"branding" form:"branding"`
	Settings string `json:"settings" form:"settings"`
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

	return paged(c, providers, total, page, perPage)
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

	provider := &domain.Provider{
		Code:     req.Code,
		Name:     req.Name,
		Status:   req.Status,
		MaxUsers: req.MaxUsers,
		MaxNas:   req.MaxNas,
		Branding: req.Branding,
		Settings: req.Settings,
	}

	if err := GetDB(c).Create(provider).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create provider: "+err.Error(), nil)
	}

	return ok(c, provider)
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

	return ok(c, provider)
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
