package webserver

import (
	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/middleware"
)

// GetTenantMiddleware returns the tenant isolation middleware for multi-provider support.
// It extracts tenant ID from X-Tenant-ID header and enforces tenant context for all requests.
func GetTenantMiddleware() echo.MiddlewareFunc {
	return middleware.TenantMiddleware(middleware.TenantMiddlewareConfig{
		SkipPaths: []string{
			"/health",
			"/ready",
			"/metrics",
			"/api/v1/public/login",
			"/api/v1/public/register",
		},
		DefaultTenant: 1, // Platform admin tenant
	})
}
