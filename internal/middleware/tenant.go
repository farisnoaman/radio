package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/tenant"
)

const (
	TenantIDHeader = "X-Tenant-ID"
)

type TenantMiddlewareConfig struct {
	SkipPaths     []string
	DefaultTenant int64
}

func TenantMiddleware(config TenantMiddlewareConfig) echo.MiddlewareFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()

			if skipPaths[path] || skipPaths[c.Request().URL.Path] {
				return next(c)
			}

			tenantHeader := c.Request().Header.Get(TenantIDHeader)
			if tenantHeader != "" {
				tenantID, err := strconv.ParseInt(tenantHeader, 10, 64)
				if err != nil || tenantID <= 0 {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid X-Tenant-ID header")
				}
				ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
				c.SetRequest(c.Request().WithContext(ctx))
			} else if config.DefaultTenant > 0 {
				ctx := tenant.WithTenantID(c.Request().Context(), config.DefaultTenant)
				c.SetRequest(c.Request().WithContext(ctx))
			}

			return next(c)
		}
	}
}

func TenantMiddlewareFromOperator(getTenantIDFunc func() (int64, error)) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID, err := getTenantIDFunc()
			if err == nil && tenantID > 0 {
				ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	}
}
