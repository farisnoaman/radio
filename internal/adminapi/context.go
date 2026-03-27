package adminapi

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

// GetAppContext gets the application context from echo context
func GetAppContext(c echo.Context) app.AppContext {
	return c.Get("appCtx").(app.AppContext) //nolint:errcheck // type assertion is safe for middleware-set context
}

// GetDB gets the database connection with request context for tenant isolation
func GetDB(c echo.Context) *gorm.DB {
	dbVal := c.Get("db")
	if dbVal != nil {
		if db, ok := dbVal.(*gorm.DB); ok && db != nil {
			return db.WithContext(c.Request().Context())
		}
	}
	return GetAppContext(c).DB().WithContext(c.Request().Context())
}

// GetConfig gets the configuration from echo context
func GetConfig(c echo.Context) *app.ConfigManager {
	return GetAppContext(c).ConfigMgr()
}

// GetTenantID gets the tenant ID from the request context.
// The tenant middleware stores this in the Go context (via tenant.WithTenantID).
func GetTenantID(c echo.Context) string {
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", tenantID)
}
