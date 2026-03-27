package repository

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

// TenantScope returns a GORM scope that filters by tenant_id from context.
// Use this in all queries to ensure tenant isolation.
// Example: db.Scopes(TenantScope).Find(&users)
func TenantScope(db *gorm.DB) *gorm.DB {
	ctx := db.Statement.Context
	if ctx == nil {
		return db
	}

	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		// No tenant context, return empty result
		return db.Where("1 = 0")
	}

	return db.Where("tenant_id = ?", tenantID)
}

// AdminTenantScope allows platform admin to query specific tenant.
// Use this in admin APIs where admin can access any tenant data.
func AdminTenantScope(tenantID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// AllTenantsScope bypasses tenant filtering (platform admin only).
// Use with caution - only for platform-level aggregation queries.
func AllTenantsScope(db *gorm.DB) *gorm.DB {
	return db
}

// TenantScopeWithID returns a scope for a specific tenant ID.
// Use when you need to query a different tenant than the current context.
func TenantScopeWithID(tenantID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// WithTenant creates a new DB instance with tenant context.
// Convenience function for queries with tenant context.
func WithTenant(db *gorm.DB, tenantID int64) *gorm.DB {
	ctx := tenant.WithTenantID(context.Background(), tenantID)
	return db.WithContext(ctx).Scopes(TenantScope)
}
