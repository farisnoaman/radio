package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/repository"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// nasTemplatePayload represents NAS template request payload.
type nasTemplatePayload struct {
	VendorCode string                     `json:"vendor_code" validate:"required"`
	Name       string                     `json:"name" validate:"required,max=200"`
	IsDefault  bool                       `json:"is_default"`
	Attributes []domain.TemplateAttribute `json:"attributes" validate:"required"`
	Remark     string                     `json:"remark" validate:"max=500"`
}

// ListNASTemplates retrieves all NAS templates for current tenant.
// @Summary list NAS templates
// @Tags NAS Template
// @Param vendor_code query string false "Filter by vendor code"
// @Success 200 {object} ListResponse
// @Router /api/v1/network/nas-templates [get]
func ListNASTemplates(c echo.Context) error {
	db := GetDB(c)
	repo := repository.NewNASTemplateRepository(db)

	vendorCode := c.QueryParam("vendor_code")

	var templates []*domain.NASTemplate
	var err error

	if vendorCode != "" {
		templates, err = repo.ListByVendor(c.Request().Context(), vendorCode)
	} else {
		// Get all templates for tenant
		tenantID, err := tenant.FromContext(c.Request().Context())
		if err != nil {
			return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
		}
		err = db.Where("tenant_id = ?", tenantID).Find(&templates).Error
	}

	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch templates", err.Error())
	}

	return ok(c, templates)
}

// CreateNASTemplate creates a new NAS template.
// @Summary create NAS template
// @Tags NAS Template
// @Param template body nasTemplatePayload true "Template data"
// @Success 201 {object} domain.NASTemplate
// @Router /api/v1/network/nas-templates [post]
func CreateNASTemplate(c echo.Context) error {
	var payload nasTemplatePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	template := &domain.NASTemplate{
		VendorCode: payload.VendorCode,
		Name:       payload.Name,
		IsDefault:  payload.IsDefault,
		Attributes: payload.Attributes,
		Remark:     payload.Remark,
		TenantID:   tenantID,
	}

	// Validate template
	if err := template.Validate(); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid template configuration", err.Error())
	}

	db := GetDB(c)
	if err := db.Create(template).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create template", err.Error())
	}

	return ok(c, template)
}

// GetNASTemplate retrieves a single template by ID.
// @Summary get NAS template
// @Tags NAS Template
// @Param id path int true "Template ID"
// @Success 200 {object} domain.NASTemplate
// @Router /api/v1/network/nas-templates/{id} [get]
func GetNASTemplate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid template ID", nil)
	}

	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	var template domain.NASTemplate
	err = db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&template).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Template not found", nil)
	}

	return ok(c, template)
}

// UpdateNASTemplate updates an existing template.
// @Summary update NAS template
// @Tags NAS Template
// @Param id path int true "Template ID"
// @Param template body nasTemplatePayload true "Template data"
// @Success 200 {object} domain.NASTemplate
// @Router /api/v1/network/nas-templates/{id} [put]
func UpdateNASTemplate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid template ID", nil)
	}

	var payload nasTemplatePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	var template domain.NASTemplate
	err = db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&template).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Template not found", nil)
	}

	// Update fields
	template.VendorCode = payload.VendorCode
	template.Name = payload.Name
	template.IsDefault = payload.IsDefault
	template.Attributes = payload.Attributes
	template.Remark = payload.Remark

	// Validate
	if err := template.Validate(); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid template configuration", err.Error())
	}

	if err := db.Save(&template).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update template", err.Error())
	}

	return ok(c, template)
}

// DeleteNASTemplate deletes a template.
// @Summary delete NAS template
// @Tags NAS Template
// @Param id path int true "Template ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/nas-templates/{id} [delete]
func DeleteNASTemplate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid template ID", nil)
	}

	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	result := db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&domain.NASTemplate{})
	if result.Error != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete template", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Template not found", nil)
	}

	return ok(c, map[string]interface{}{"message": "Template deleted successfully"})
}

// registerNASTemplateRoutes registers NAS template routes.
func registerNASTemplateRoutes() {
	webserver.ApiGET("/network/nas-templates", ListNASTemplates)
	webserver.ApiGET("/network/nas-templates/:id", GetNASTemplate)
	webserver.ApiPOST("/network/nas-templates", CreateNASTemplate)
	webserver.ApiPUT("/network/nas-templates/:id", UpdateNASTemplate)
	webserver.ApiDELETE("/network/nas-templates/:id", DeleteNASTemplate)
}
