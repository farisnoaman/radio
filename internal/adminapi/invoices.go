package adminapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app/billing"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ListInvoices retrieves the invoice list
// @Summary get the invoice list
// @Tags Invoice
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Param username query string false "Filter by username"
// @Param status query string false "Filter by status"
// @Success 200 {object} ListResponse
// @Router /api/v1/invoices [get]
func ListInvoices(c echo.Context) error {
	db := GetDB(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	sortField := c.QueryParam("sort")
	order := c.QueryParam("order")
	if sortField == "" {
		sortField = "id"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var invoices []domain.Invoice

	query := db.Model(&domain.Invoice{})

	// Authentication context check
	if userVal := c.Get("user"); userVal != nil {
		if token, ok := userVal.(*jwt.Token); ok {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				role, _ := claims["role"].(string)
				tokenUsername, _ := claims["username"].(string)
				if role == "user" {
					// Enforce filtering by own username for portal users
					query = query.Where("username = ?", tokenUsername)
				} else {
					// Admin/Operator can filter by any username
					if username := strings.TrimSpace(c.QueryParam("username")); username != "" {
						query = query.Where("username = ?", username)
					}
				}
			}
		}
	}

	// Filter by status (available for everyone)
	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total using a cloned session to avoid corrupting the query for Find
	countQuery := query.Session(&gorm.Session{NewDB: true}).Model(&domain.Invoice{})
	countQuery.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&invoices)

	return paged(c, invoices, total, page, perPage)
}

// GetInvoice retrieves a single invoice
// @Summary get invoice detail
// @Tags Invoice
// @Param id path int true "Invoice ID"
// @Success 200 {object} domain.Invoice
// @Router /api/v1/invoices/{id} [get]
func GetInvoice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid invoice ID", nil)
	}

	var invoice domain.Invoice
	if err := GetDB(c).First(&invoice, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Invoice not found", nil)
	}

	// Security: Verify ownership for RADIUS users
	if userVal := c.Get("user"); userVal != nil {
		if token, ok := userVal.(*jwt.Token); ok {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				role, _ := claims["role"].(string)
				tokenUsername, _ := claims["username"].(string)
				if role == "user" && invoice.Username != tokenUsername {
					return fail(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to view this invoice", nil)
				}
			}
		}
	}

	return ok(c, invoice)
}

// PayInvoice endpoint records a manual payment for an invoice
// @Summary pay an invoice
// @Tags Invoice
// @Param id path int true "Invoice ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/invoices/{id}/pay [post]
func PayInvoice(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid invoice ID", nil)
	}

	// Security: RADIUS users cannot manually record payments
	if userVal := c.Get("user"); userVal != nil {
		if token, ok := userVal.(*jwt.Token); ok {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				role, _ := claims["role"].(string)
				if role == "user" {
					return fail(c, http.StatusForbidden, "FORBIDDEN", "Portal users cannot record manual payments", nil)
				}
			}
		}
	}

	// Billing Engine config - using defaults as there is no specific config lookup needed just for marking paid
	engine := billing.NewBillingEngine(GetDB(c), 7)
	
	if err := engine.PayInvoice(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fail(c, http.StatusNotFound, "NOT_FOUND", "Invoice not found", nil)
		}
		return fail(c, http.StatusInternalServerError, "PAY_FAILED", "Failed to pay invoice", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "Invoice marked as paid successfully",
	})
}

// GenerateUserInvoice allows an admin to trigger an early invoice for a user
func GenerateUserInvoice(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return fail(c, http.StatusBadRequest, "INVALID_USERNAME", "Username is required", nil)
	}

	zap.S().Infof("Admin triggering manual invoice generation for user: %s", username)
	engine := billing.NewBillingEngine(GetDB(c).Debug(), 7)
	if err := engine.GenerateEarlyInvoice(username); err != nil {
		zap.S().Errorf("Manual invoice generation failed for user %s: %v", username, err)
		return fail(c, http.StatusInternalServerError, "BILLING_FAILED", "Failed to generate invoice", err.Error())
	}

	zap.S().Infof("Manual invoice generation succeeded for user: %s", username)
	return ok(c, map[string]interface{}{
		"message": "Invoice generated successfully and next billing date advanced",
	})
}



func registerInvoiceRoutes() {
	webserver.ApiGET("/radius/invoices", ListInvoices)
	webserver.ApiGET("/radius/invoices/:id", GetInvoice)
	webserver.ApiPOST("/radius/invoices/:id/pay", PayInvoice)
	webserver.ApiPOST("/radius/users/:username/bill", GenerateUserInvoice)
}
