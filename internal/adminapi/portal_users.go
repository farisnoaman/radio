package adminapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerPortalUserRoutes() {
	webserver.ApiPOST("/portal/user/mac/unbind", UnbindPortalMac)
}

// UnbindPortalMac allows a user to reset their own MAC binding
func UnbindPortalMac(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found", nil)
	}

	db := GetDB(c)
	
	// Reset MAC address and BindMac setting
	updates := map[string]interface{}{
		"mac_addr":   "",
		"updated_at": time.Now(),
	}

	if err := db.Model(user).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to unbind MAC", err.Error())
	}

	LogOperation(c, "portal_unbind_mac", fmt.Sprintf("User %s reset their MAC binding", user.Username))

	return ok(c, map[string]string{"message": "MAC address unbinded successfully"})
}
