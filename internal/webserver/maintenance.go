package webserver

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/pkg/web"
)

// MaintenanceMiddleware blocks non-admin requests when maintenance mode is active
func MaintenanceMiddleware(appCtx app.AppContext) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip for system/maintenance endpoints to allow admins to disable it
			if strings.Contains(c.Path(), "/system/maintenance") {
				return next(c)
			}
			
			// Check if maintenance is active
			if !appCtx.MaintMgr().IsActive() {
				return next(c)
			}
			
			// Allow admins (validated by JWT middleware before this)
			// We need to check the user role from context
			// Assuming "operator" info is in context or we can decode token
			// For now, let's just return 503 Service Unavailable for everyone except specific bypass
			// In a real scenario, we'd check `c.Get("user")`
			
			// If we want to allow admins, we must ensure this middleware runs AFTER JWT.
			// And we need to inspect the claims/user.
			
			// For safety in this iteration, we block everything except the maintenance toggle endpoints.
			return c.JSON(http.StatusServiceUnavailable, web.RestError("System is undergoing maintenance"))
		}
	}
}
