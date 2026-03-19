package tenant

import (
	"context"
	"errors"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	DefaultTenantID int64 = 1
)

var (
	ErrNoTenant      = errors.New("no tenant context found in request")
	ErrInvalidTenant = errors.New("invalid tenant ID")
	ErrTenantMismatch = errors.New("tenant ID mismatch")
)

// FromContext extracts the tenant ID from a context.
// Returns ErrNoTenant if no tenant ID is present.
func FromContext(ctx context.Context) (int64, error) {
	tenantID, ok := ctx.Value(TenantIDKey).(int64)
	if !ok || tenantID <= 0 {
		return 0, ErrNoTenant
	}
	return tenantID, nil
}

// WithTenantID returns a new context with the specified tenant ID.
// Panics if tenantID is not positive.
func WithTenantID(ctx context.Context, tenantID int64) context.Context {
	if tenantID <= 0 {
		panic("tenant ID must be positive")
	}
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// MustFromContext extracts the tenant ID from context.
// Panics if no tenant ID is present.
func MustFromContext(ctx context.Context) int64 {
	tenantID, err := FromContext(ctx)
	if err != nil {
		panic(err)
	}
	return tenantID
}

// GetTenantIDOrDefault returns the tenant ID from context or the default.
func GetTenantIDOrDefault(ctx context.Context) int64 {
	tenantID, err := FromContext(ctx)
	if err != nil {
		return DefaultTenantID
	}
	return tenantID
}

// ValidateTenantID checks if the tenant ID is valid (positive).
func ValidateTenantID(tenantID int64) error {
	if tenantID <= 0 {
		return ErrInvalidTenant
	}
	return nil
}

// TenantContext wraps a context with tenant information.
type TenantContext struct {
	TenantID int64
	Context  context.Context
}

// NewTenantContext creates a new TenantContext.
func NewTenantContext(ctx context.Context, tenantID int64) (*TenantContext, error) {
	if err := ValidateTenantID(tenantID); err != nil {
		return nil, err
	}
	return &TenantContext{
		TenantID: tenantID,
		Context:  WithTenantID(ctx, tenantID),
	}, nil
}

// Extract extracts the tenant context from the provided context.
func (tc *TenantContext) Extract() context.Context {
	return tc.Context
}

// TenantChecker provides methods to check tenant-related conditions.
type TenantChecker struct{}

// NewTenantChecker creates a new TenantChecker.
func NewTenantChecker() *TenantChecker {
	return &TenantChecker{}
}

// IsSystemTenant checks if the tenant ID represents the system tenant.
func (c *TenantChecker) IsSystemTenant(tenantID int64) bool {
	return tenantID == DefaultTenantID
}

// CanAccess checks if a request from sourceTenant can access targetTenant resources.
func (c *TenantChecker) CanAccess(sourceTenantID, targetTenantID int64) bool {
	if c.IsSystemTenant(sourceTenantID) {
		return true
	}
	return sourceTenantID == targetTenantID
}
