package adminapi

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/acs"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// cpePayload defines the CPE device update structure
type cpePayload struct {
	Status        *string `json:"status"`
	AutoProvision *bool   `json:"auto_provision"`
	ProfileID     *int64  `json:"profile_id"`
}

// registerCPERoutes registers CPE device routes
func registerCPERoutes() {
	webserver.ApiGET("/cpes", listCPEs)
	webserver.ApiGET("/cpes/:id", getCPE)
	webserver.ApiPUT("/cpes/:id", updateCPE)
	webserver.ApiDELETE("/cpes/:id", deleteCPE)
}

// listCPEs retrieves the CPE device list
func listCPEs(c echo.Context) error {
	page, pageSize := parsePagination(c)

	db := GetDB(c).Model(&acs.CPEDevice{})

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query CPE devices", err.Error())
	}

	var devices []acs.CPEDevice
	if err := db.
		Order("last_inform DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&devices).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query CPE devices", err.Error())
	}

	return paged(c, devices, total, page, pageSize)
}

// getCPE retrieves a single CPE device
func getCPE(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid CPE ID", nil)
	}

	var device acs.CPEDevice
	if err := GetDB(c).Where("id = ?", id).First(&device).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "CPE_NOT_FOUND", "CPE device not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query CPE device", err.Error())
	}

	return ok(c, device)
}

// updateCPE updates a CPE device
func updateCPE(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid CPE ID", nil)
	}

	var device acs.CPEDevice
	if err := GetDB(c).Where("id = ?", id).First(&device).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "CPE_NOT_FOUND", "CPE device not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query CPE device", err.Error())
	}

	var payload cpePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
	}

	updates := make(map[string]interface{})
	if payload.Status != nil {
		updates["status"] = *payload.Status
	}
	if payload.AutoProvision != nil {
		updates["auto_provision"] = *payload.AutoProvision
	}
	if payload.ProfileID != nil {
		updates["profile_id"] = *payload.ProfileID
	}

	if err := GetDB(c).Model(&device).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update CPE device", err.Error())
	}

	return ok(c, device)
}

// deleteCPE deletes a CPE device
func deleteCPE(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid CPE ID", nil)
	}

	if err := GetDB(c).Delete(&acs.CPEDevice{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete CPE device", err.Error())
	}

	return ok(c, nil)
}
