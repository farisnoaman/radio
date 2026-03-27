package adminapi

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/billing"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// registerBillingRoutes registers billing management routes
func registerBillingRoutes() {
	// Provider routes
	webserver.ApiGET("/billing/invoices", GetProviderInvoices)
	webserver.ApiGET("/billing/invoices/:id", GetProviderInvoice)
	webserver.ApiPOST("/billing/invoices/:id/pay", PayProviderInvoice)

	// Admin routes
	webserver.ApiGET("/admin/billing/plans", ListBillingPlans)
	webserver.ApiPOST("/admin/billing/plans", CreateBillingPlan)
	webserver.ApiPOST("/admin/billing/run", TriggerBillingCycle)
}

// GetProviderInvoices returns invoices for current tenant
func GetProviderInvoices(c echo.Context) error {
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	db := GetDB(c)

	var invoices []domain.ProviderInvoice
	err = db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&invoices).Error

	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch invoices", err)
	}

	return ok(c, invoices)
}

// GetProviderInvoice returns a specific provider invoice
func GetProviderInvoice(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "MISSING_ID", "Invoice ID is required", nil)
	}

	tenantID, _ := tenant.FromContext(c.Request().Context())
	db := GetDB(c)

	var invoice domain.ProviderInvoice
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&invoice).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Invoice not found", err)
	}

	return ok(c, invoice)
}

// PayProviderInvoice marks a provider invoice as paid
func PayProviderInvoice(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "MISSING_ID", "Invoice ID is required", nil)
	}

	tenantID, _ := tenant.FromContext(c.Request().Context())
	db := GetDB(c)

	var invoice domain.ProviderInvoice
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&invoice).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Invoice not found", err)
	}

	// Update invoice status
	now := time.Now()
	invoice.Status = "paid"
	invoice.PaidDate = &now
	db.Save(&invoice)

	return ok(c, invoice)
}

// ListBillingPlans returns all billing plans (admin only)
func ListBillingPlans(c echo.Context) error {
	// Verify platform admin
	if !IsPlatformAdmin(c) {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
	}

	db := GetDB(c)

	var plans []domain.BillingPlan
	err := db.Where("is_active = ?", true).Find(&plans).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch plans", err)
	}

	return ok(c, plans)
}

// CreateBillingPlan creates a new billing plan (admin only)
func CreateBillingPlan(c echo.Context) error {
	// Verify platform admin
	if !IsPlatformAdmin(c) {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
	}

	var req struct {
		Code          string  `json:"code" validate:"required"`
		Name          string  `json:"name" validate:"required"`
		BaseFee       float64 `json:"base_fee" validate:"required,min=0"`
		IncludedUsers int     `json:"included_users" validate:"required,min=0"`
		OverageFee    float64 `json:"overage_fee" validate:"required,min=0"`
		MaxUsers      int     `json:"max_users" validate:"required,min=0"`
		Features      string  `json:"features"`
	}

	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
	}

	db := GetDB(c)

	// Check if code already exists
	var count int64
	db.Model(&domain.BillingPlan{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "CODE_EXISTS", "Billing plan code already exists", nil)
	}

	plan := &domain.BillingPlan{
		Code:          req.Code,
		Name:          req.Name,
		BaseFee:       req.BaseFee,
		IncludedUsers: req.IncludedUsers,
		OverageFee:    req.OverageFee,
		MaxUsers:      req.MaxUsers,
		Features:      req.Features,
		IsActive:      true,
	}

	if err := db.Create(plan).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create billing plan", err)
	}

	return ok(c, plan)
}

// TriggerBillingCycle manually triggers the billing cycle (admin only)
func TriggerBillingCycle(c echo.Context) error {
	// Verify platform admin
	if !IsPlatformAdmin(c) {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
	}

	billingEngine := getBillingEngine(c)
	if billingEngine == nil {
		return fail(c, http.StatusInternalServerError, "ENGINE_NOT_FOUND", "Billing engine not initialized", nil)
	}

	if err := billingEngine.GenerateMonthlyInvoices(c.Request().Context()); err != nil {
		return fail(c, http.StatusInternalServerError, "BILLING_ERROR", "Failed to run billing", err)
	}

	return ok(c, map[string]string{"message": "Billing cycle triggered successfully"})
}

// getBillingEngine returns the billing engine from application context
func getBillingEngine(c echo.Context) *billing.BillingEngine {
	if engine, ok := c.Get("billingEngine").(*billing.BillingEngine); ok && engine != nil {
		return engine
	}
	return nil
}
