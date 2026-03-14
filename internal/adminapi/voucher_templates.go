package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
)

// ListVoucherTemplates returns templates visible to the current user.
// Users see: their own templates + all public templates + all default templates.
// @Summary list voucher templates
// @Tags VoucherTemplate
// @Success 200 {array} domain.VoucherTemplate
// @Router /api/v1/voucher-templates [get]
func ListVoucherTemplates(c echo.Context) error {
	db := GetDB(c)

	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	var templates []domain.VoucherTemplate
	// Show: own templates OR public templates OR default templates
	if err := db.Where("owner_id = ? OR is_public = ? OR is_default = ?", currentUser.ID, true, true).
		Order("is_default DESC, created_at DESC").
		Find(&templates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query templates", err.Error())
	}

	return ok(c, templates)
}

// CreateVoucherTemplate creates a new custom template for the current user.
// @Summary create voucher template
// @Tags VoucherTemplate
// @Param template body domain.VoucherTemplate true "Template info"
// @Success 200 {object} domain.VoucherTemplate
// @Router /api/v1/voucher-templates [post]
func CreateVoucherTemplate(c echo.Context) error {
	var tmpl domain.VoucherTemplate
	if err := c.Bind(&tmpl); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if tmpl.Name == "" || tmpl.Content == "" {
		return fail(c, http.StatusBadRequest, "MISSING_FIELDS", "Name and content are required", nil)
	}

	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	now := time.Now()
	tmpl.ID = 0
	tmpl.OwnerID = currentUser.ID
	tmpl.IsDefault = false // Only system can set defaults
	tmpl.CreatedAt = now
	tmpl.UpdatedAt = now

	db := GetDB(c)
	if err := db.Create(&tmpl).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create template", err.Error())
	}

	zap.L().Info("Voucher template created",
		zap.String("name", tmpl.Name),
		zap.Int64("owner_id", tmpl.OwnerID),
		zap.Bool("is_public", tmpl.IsPublic))

	return ok(c, tmpl)
}

// UpdateVoucherTemplate updates an existing template owned by the current user.
// @Summary update voucher template
// @Tags VoucherTemplate
// @Param id path int true "Template ID"
// @Param template body domain.VoucherTemplate true "Template info"
// @Success 200 {object} domain.VoucherTemplate
// @Router /api/v1/voucher-templates/{id} [put]
func UpdateVoucherTemplate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid template ID", err.Error())
	}

	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	db := GetDB(c)

	var existing domain.VoucherTemplate
	if err := db.First(&existing, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Template not found", err.Error())
	}

	// Only the owner (or admin) can update
	if existing.OwnerID != currentUser.ID && currentUser.Level != "admin" && currentUser.Level != "super" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "You can only edit your own templates", nil)
	}

	// Cannot edit system defaults
	if existing.IsDefault {
		return fail(c, http.StatusForbidden, "SYSTEM_TEMPLATE", "Cannot modify system default templates", nil)
	}

	var update domain.VoucherTemplate
	if err := c.Bind(&update); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	existing.Name = update.Name
	existing.Content = update.Content
	existing.IsPublic = update.IsPublic
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update template", err.Error())
	}

	return ok(c, existing)
}

// DeleteVoucherTemplate deletes a template owned by the current user.
// @Summary delete voucher template
// @Tags VoucherTemplate
// @Param id path int true "Template ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/voucher-templates/{id} [delete]
func DeleteVoucherTemplate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid template ID", err.Error())
	}

	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	db := GetDB(c)

	var existing domain.VoucherTemplate
	if err := db.First(&existing, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Template not found", err.Error())
	}

	// Only the owner (or admin) can delete
	if existing.OwnerID != currentUser.ID && currentUser.Level != "admin" && currentUser.Level != "super" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "You can only delete your own templates", nil)
	}

	if existing.IsDefault {
		return fail(c, http.StatusForbidden, "SYSTEM_TEMPLATE", "Cannot delete system default templates", nil)
	}

	if err := db.Delete(&existing).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete template", err.Error())
	}

	return ok(c, map[string]interface{}{"message": "Template deleted"})
}

func registerVoucherTemplateRoutes() {
	webserver.ApiGET("/voucher-templates", ListVoucherTemplates)
	webserver.ApiPOST("/voucher-templates", CreateVoucherTemplate)
	webserver.ApiPUT("/voucher-templates/:id", UpdateVoucherTemplate)
	webserver.ApiDELETE("/voucher-templates/:id", DeleteVoucherTemplate)
}
