package middleware

import (
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
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

			var tenantID int64
			var err error

			// First, try to get tenant_id from X-Tenant-ID header
			tenantHeader := c.Request().Header.Get(TenantIDHeader)
			if tenantHeader != "" {
				tenantID, err = strconv.ParseInt(tenantHeader, 10, 64)
				if err != nil || tenantID <= 0 {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid X-Tenant-ID header")
				}
			} else {
				// If no header, try to extract from JWT token
				userVal := c.Get("user")
				if userVal != nil {
					if token, ok := userVal.(*jwt.Token); ok {
						if claims, ok := token.Claims.(jwt.MapClaims); ok {
							if tidVal, ok := claims["tenant_id"]; ok {
								switch v := tidVal.(type) {
								case float64:
									tenantID = int64(v)
								case string:
									tenantID, err = strconv.ParseInt(v, 10, 64)
									if err != nil {
										return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant_id in token")
									}
								}
							}
						}
					}
				}
			}

			// If we have a tenant ID, validate and set it
			if tenantID > 0 {
				// Validate tenant ID
				if err := tenant.ValidateTenantID(tenantID); err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, err.Error())
				}

				ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
				c.SetRequest(c.Request().WithContext(ctx))
			} else if config.DefaultTenant > 0 {
				// Use default tenant if configured
				ctx := tenant.WithTenantID(c.Request().Context(), config.DefaultTenant)
				c.SetRequest(c.Request().WithContext(ctx))
			} else {
				// No tenant context available
				return echo.NewHTTPError(http.StatusUnauthorized, "missing tenant identification")
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
				if err := tenant.ValidateTenantID(tenantID); err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, err.Error())
				}
				ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	}
}

// RequireTenant checks if request has tenant context and returns error if not.
// Use this in individual handlers that require tenant context.
func RequireTenant() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, err := tenant.FromContext(c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "tenant context required")
			}
			return next(c)
		}
	}
}
