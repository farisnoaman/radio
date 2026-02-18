package adminapi

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerMaintenanceRoutes() {
	webserver.ApiGET("/system/maintenance", GetMaintenanceStatus)
	webserver.ApiPOST("/system/maintenance/enable", EnableMaintenance)
	webserver.ApiPOST("/system/maintenance/disable", DisableMaintenance)
}

func GetMaintenanceStatus(c echo.Context) error {
	active := GetAppContext(c).MaintMgr().IsActive()
	return ok(c, map[string]bool{
		"active": active,
	})
}

func EnableMaintenance(c echo.Context) error {
	drain := c.QueryParam("drain") == "true"
	
	err := GetAppContext(c).MaintMgr().Enable()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "MAINTENANCE_ERROR", "Failed to enable maintenance mode", err.Error())
	}
	
	if drain {
		go func() {
			// Run drain in background with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			_ = GetAppContext(c).MaintMgr().DrainSessions(ctx)
		}()
	}
	
	return ok(c, map[string]interface{}{
		"message": "Maintenance mode enabled",
		"drain_started": drain,
	})
}

func DisableMaintenance(c echo.Context) error {
	err := GetAppContext(c).MaintMgr().Disable()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "MAINTENANCE_ERROR", "Failed to disable maintenance mode", err.Error())
	}
	
	return ok(c, map[string]string{
		"message": "Maintenance mode disabled",
	})
}
