package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// checkSuperAdminAccess verifies the current user is a super admin
func checkSuperAdminAccess(c echo.Context) error {
	currentUser := GetOperator(c)
	if currentUser == nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
	}

	if currentUser.Level != "super" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only platform admins can manage provider admin credentials", nil)
	}

	return nil
}

// GetProviderAdminCredentials retrieves tenant admin credentials (password masked)
// @Summary Get provider admin credentials
// @Tags ProviderAdmin
// @Param id path int true "Provider ID"
// @Success 200 {object} AdminCredentialsResponse
// @Router /api/v1/platform/providers/{id}/admin-credentials [get]
func GetProviderAdminCredentials(c echo.Context) error {
	// Authorization check
	if err := checkSuperAdminAccess(c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", nil)
	}

	// Verify provider exists
	var provider domain.Provider
	if err := GetDB(c).First(&provider, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "PROVIDER_NOT_FOUND", "Provider does not exist", nil)
	}

	// Find tenant admin for this provider
	var admin domain.SysOpr
	err = GetDB(c).Where("tenant_id = ? AND level = ?", id, "admin").First(&admin).Error

	if err != nil {
		// Return default credentials if no admin exists yet
		return ok(c, map[string]interface{}{
			"username": "admin",
			"password": "********", // Masked
			"level":     "admin",
			"status":    "not_created",
			"enabled":   false,
		})
	}

	// Return masked password with enabled status
	enabled := admin.Status == "enabled"
	return ok(c, map[string]interface{}{
		"username": admin.Username,
		"password": "********", // Masked
		"level":     admin.Level,
		"status":    admin.Status,
		"enabled":   enabled,
	})
}

// UpdateAdminCredentialsRequest represents the request to update admin credentials
type UpdateAdminCredentialsRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateProviderAdminCredentials updates tenant admin credentials
// @Summary Update provider admin credentials
// @Tags ProviderAdmin
// @Param id path int true "Provider ID"
// @Param credentials body UpdateAdminCredentialsRequest true "New credentials"
// @Success 200 {object} AdminCredentialsResponse
// @Router /api/v1/platform/providers/{id}/admin-credentials [put]
func UpdateProviderAdminCredentials(c echo.Context) error {
	// Authorization check
	if err := checkSuperAdminAccess(c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", nil)
	}

	var req UpdateAdminCredentialsRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return handleValidationError(c, err)
	}

	// Verify provider exists
	var provider domain.Provider
	if err := GetDB(c).First(&provider, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "PROVIDER_NOT_FOUND", "Provider does not exist", nil)
	}

	// Check for username conflict
	var existingAdmin domain.SysOpr
	err = GetDB(c).Where("username = ? AND tenant_id != ?", req.Username, id).First(&existingAdmin).Error
	if err == nil {
		return fail(c, http.StatusConflict, "USERNAME_EXISTS", "An operator with this username already exists", nil)
	}

	// Find or create tenant admin
	var admin domain.SysOpr
	err = GetDB(c).Where("tenant_id = ? AND level = ?", id, "admin").First(&admin).Error

	if err != nil {
		// Create new admin
		admin = domain.SysOpr{
			TenantID: id,
			Username: req.Username,
			Password: common.Sha256HashWithSalt(req.Password, common.GetSecretSalt()),
			Level:    "admin",
			Status:   "enabled",
		}
		if err := GetDB(c).Create(&admin).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create admin", err.Error())
		}
	} else {
		// Update existing admin
		oldUsername := admin.Username
		admin.Username = req.Username
		admin.Password = common.Sha256HashWithSalt(req.Password, common.GetSecretSalt())

		if err := GetDB(c).Save(&admin).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update admin", err.Error())
		}

		// TODO: Log credential change (username changed from X to Y)
		_ = oldUsername
	}

	// Return full credentials (only time password is shown)
	enabled := admin.Status == "enabled"
	return ok(c, map[string]interface{}{
		"username": admin.Username,
		"password": admin.Password,
		"level":     admin.Level,
		"status":    admin.Status,
		"enabled":   enabled,
		"message":   "Credentials updated successfully. Please save the password now.",
	})
}

// ResetProviderAdminCredentials resets tenant admin credentials to defaults
// @Summary Reset provider admin credentials to defaults
// @Tags ProviderAdmin
// @Param id path int true "Provider ID"
// @Success 200 {object} AdminCredentialsResponse
// @Router /api/v1/platform/providers/{id}/admin-credentials/reset [post]
func ResetProviderAdminCredentials(c echo.Context) error {
	// Authorization check
	if err := checkSuperAdminAccess(c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid provider ID", nil)
	}

	// Verify provider exists
	var provider domain.Provider
	if err := GetDB(c).First(&provider, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "PROVIDER_NOT_FOUND", "Provider does not exist", nil)
	}

	// Default credentials
	defaultUsername := "admin"
	defaultPassword := "123456"

	// Find or create tenant admin
	var admin domain.SysOpr
	err = GetDB(c).Where("tenant_id = ? AND level = ?", id, "admin").First(&admin).Error

	if err != nil {
		// Create new admin with defaults
		admin = domain.SysOpr{
			TenantID: id,
			Username: defaultUsername,
			Password: common.Sha256HashWithSalt(defaultPassword, common.GetSecretSalt()),
			Level:    "admin",
			Status:   "enabled",
		}
		if err := GetDB(c).Create(&admin).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create admin", err.Error())
		}
	} else {
		// Reset existing admin to defaults
		oldUsername := admin.Username
		admin.Username = defaultUsername
		admin.Password = common.Sha256HashWithSalt(defaultPassword, common.GetSecretSalt())

		if err := GetDB(c).Save(&admin).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to reset admin", err.Error())
		}

		// TODO: Log credential reset (username changed from X to default)
		_ = oldUsername
	}

	// Return full credentials
	enabled := admin.Status == "enabled"
	return ok(c, map[string]interface{}{
		"username": admin.Username,
		"password": admin.Password,
		"level":     admin.Level,
		"status":    admin.Status,
		"enabled":   enabled,
		"message":   "Credentials reset to defaults. Please save the password now.",
	})
}
