package adminapi

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerSystemLogRoutes() {
	webserver.ApiGET("/system/logs", ListSystemLogs)
}

// ListSystemLogs retrieves the system operation logs
// @Summary get system logs
// @Tags System
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param operator query string false "Operator name"
// @Param action query string false "Action type"
// @Success 200 {object} ListResponse
// @Router /api/v1/system/logs [get]
func ListSystemLogs(c echo.Context) error {
	// Permission check
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	// Only super admin and admin can view logs
	if currentUser.Level != "super" && currentUser.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Access denied", nil)
	}

	page, perPage := parsePagination(c)
	
	db := GetDB(c).Model(&domain.SysOprLog{})

	// Filters
	if oprName := strings.TrimSpace(c.QueryParam("operator")); oprName != "" {
		db = db.Where("opr_name LIKE ?", "%"+oprName+"%")
	}

	if action := strings.TrimSpace(c.QueryParam("action")); action != "" {
		db = db.Where("opt_action LIKE ?", "%"+action+"%")
	}
	
	if keyword := strings.TrimSpace(c.QueryParam("keyword")); keyword != "" {
		db = db.Where("opt_desc LIKE ?", "%"+keyword+"%")
	}

	var total int64
	db.Count(&total)

	var logs []domain.SysOprLog
	offset := (page - 1) * perPage
	db.Order("id DESC").Limit(perPage).Offset(offset).Find(&logs)

	return paged(c, logs, total, page, perPage)
}
