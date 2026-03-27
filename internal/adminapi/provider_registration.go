package adminapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/migration"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

type CreateRegistrationRequest struct {
	CompanyName   string `json:"company_name" validate:"required"`
	ContactName   string `json:"contact_name" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	BusinessType  string `json:"business_type"`
	ExpectedUsers int    `json:"expected_users" validate:"min=1"`
	ExpectedNas   int    `json:"expected_nas" validate:"min=1"`
	Country       string `json:"country"`
	Message       string `json:"message"`
}

// CreateProviderRegistration handles public provider registration requests
func CreateProviderRegistration(c echo.Context) error {
	var req CreateRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
	}

	// Check if email already registered
	db := GetDB(c)
	var existingCount int64
	db.Model(&domain.ProviderRegistration{}).Where("email = ?", req.Email).Count(&existingCount)
	if existingCount > 0 {
		return fail(c, http.StatusConflict, "EMAIL_EXISTS", "Email already registered", nil)
	}

	// Create registration request
	registration := &domain.ProviderRegistration{
		CompanyName:   req.CompanyName,
		ContactName:   req.ContactName,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		BusinessType:  req.BusinessType,
		ExpectedUsers:  req.ExpectedUsers,
		ExpectedNas:   req.ExpectedNas,
		Country:       req.Country,
		Message:       req.Message,
		Status:        "pending",
	}

	if err := db.Create(registration).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create registration", nil)
	}

	// TODO: Send confirmation email

	return c.JSON(http.StatusCreated, registration)
}

// GetRegistrationStatus returns the status of a registration request
func GetRegistrationStatus(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "MISSING_ID", "Registration ID is required", nil)
	}

	db := GetDB(c)
	var registration domain.ProviderRegistration
	if err := db.Where("id = ?", id).First(&registration).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Registration not found", nil)
	}

	// Return only status (not full details for privacy)
	return ok(c, map[string]string{
		"status":  registration.Status,
		"message": getStatusMessage(registration.Status),
	})
}

func getStatusMessage(status string) string {
	switch status {
	case "pending":
		return "Your registration is pending review"
	case "approved":
		return "Your registration has been approved"
	case "rejected":
		return "Your registration has been rejected"
	default:
		return "Unknown status"
	}
}

type ApproveRegistrationRequest struct {
	ProviderCode string `json:"provider_code" validate:"required"`
	MaxUsers     int    `json:"max_users"`
	MaxNas       int    `json:"max_nas"`
	AdminUsername string `json:"admin_username" validate:"omitempty,min=3,max=50,alphanum"`
	AdminPassword string `json:"admin_password" validate:"omitempty,min=6"`
}

type RejectRegistrationRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// ApproveRegistration handles provider registration approval
func ApproveRegistration(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "MISSING_ID", "Registration ID is required", nil)
	}

	var req ApproveRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
	}

	db := GetDB(c)

	// Get registration
	var registration domain.ProviderRegistration
	if err := db.First(&registration, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Registration not found", nil)
	}

	if registration.Status != "pending" {
		return fail(c, http.StatusBadRequest, "INVALID_STATUS", "Registration is not pending", nil)
	}

	// Check if provider code already exists
	var existingCount int64
	db.Model(&domain.Provider{}).Where("code = ?", req.ProviderCode).Count(&existingCount)
	if existingCount > 0 {
		return fail(c, http.StatusConflict, "CODE_EXISTS", "Provider code already exists", nil)
	}

	// Set defaults
	maxUsers := req.MaxUsers
	if maxUsers == 0 {
		maxUsers = registration.ExpectedUsers
	}
	maxNas := req.MaxNas
	if maxNas == 0 {
		maxNas = registration.ExpectedNas
	}

	// Create provider
	provider := &domain.Provider{
		Code:     req.ProviderCode,
		Name:     registration.CompanyName,
		Status:   "active",
		MaxUsers: maxUsers,
		MaxNas:   maxNas,
	}

	if err := db.Create(provider).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create provider", nil)
	}

	// Create provider schema (only for PostgreSQL - skip for SQLite)
	// SQLite doesn't support schemas, so we skip this in development
	migrator := migration.NewSchemaMigrator(db)
	if err := migrator.CreateProviderSchema(provider.ID); err != nil {
		// Log the error but don't fail - schema creation is PostgreSQL-specific
		// For SQLite, we just continue without separate schemas
		fmt.Printf("Warning: Failed to create provider schema (this is expected for SQLite): %v\n", err)
	}

	// Create default admin operator for provider
	// Determine admin credentials - use custom if provided, otherwise defaults
	adminUsername := "admin"
	adminPassword := "123456" // Default password

	if req.AdminUsername != "" && req.AdminPassword != "" {
		// Use custom credentials provided during approval
		adminUsername = req.AdminUsername
		adminPassword = req.AdminPassword
	}

	opr := &domain.SysOpr{
		TenantID:  provider.ID,
		Realname:  registration.ContactName,
		Email:     registration.Email,
		Mobile:    registration.Phone,
		Username:  adminUsername,
		Password:  common.Sha256HashWithSalt(adminPassword, common.GetSecretSalt()),
		Level:     "admin",
		Status:    "enabled",
	}
	db.Create(opr)

	// Update registration status
	now := time.Now()
	db.Model(&registration).Updates(map[string]interface{}{
		"status":      "approved",
		"reviewed_by": GetOperatorID(c),
		"reviewed_at": &now,
	})

	// Send welcome email with credentials
	// TODO: Implement email service

	return ok(c, map[string]interface{}{
		"provider": provider,
		"admin":    opr,
	})
}

// RejectRegistration handles provider registration rejection
func RejectRegistration(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "MISSING_ID", "Registration ID is required", nil)
	}

	var req RejectRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err)
	}

	db := GetDB(c)

	// Get registration
	var registration domain.ProviderRegistration
	if err := db.First(&registration, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Registration not found", nil)
	}

	// Update registration status
	now := time.Now()
	err := db.Model(&registration).Updates(map[string]interface{}{
		"status":           "rejected",
		"rejection_reason": req.Reason,
		"reviewed_by":      GetOperatorID(c),
		"reviewed_at":      &now,
	}).Error

	if err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update registration", nil)
	}

	// Send rejection email
	// TODO: Implement email service

	return ok(c, map[string]string{"message": "Registration rejected"})
}

// ListRegistrations lists all registration requests (admin only)
func ListRegistrations(c echo.Context) error {
	db := GetDB(c)

	status := c.QueryParam("status")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := db.Model(&domain.ProviderRegistration{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var registrations []domain.ProviderRegistration
	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&registrations).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch registrations", nil)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":   registrations,
		"total":  total,
		"page":   page,
		"perPage": perPage,
	})
}

// GetRegistration returns a single registration request (admin only)
func GetRegistration(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return fail(c, http.StatusBadRequest, "MISSING_ID", "Registration ID is required", nil)
	}

	db := GetDB(c)
	var registration domain.ProviderRegistration
	if err := db.Where("id = ?", id).First(&registration).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Registration not found", nil)
	}

	return ok(c, registration)
}

// generateRandomPassword generates a random password for new provider admin
func generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// GetOperatorID gets the current operator ID from context
func GetOperatorID(c echo.Context) int64 {
	opr := GetOperator(c)
	if opr != nil {
		return opr.ID
	}
	return 0
}

// registerProviderRegistrationRoutes registers all provider registration routes
func registerProviderRegistrationRoutes() {
	webserver.ApiGET("/providers/registrations", ListRegistrations)
	webserver.ApiGET("/providers/registrations/:id", GetRegistration)
	webserver.ApiPOST("/providers/registrations", CreateProviderRegistration)      // Public (auth skipped via JwtSkipPrefix)
	webserver.ApiPOST("/providers/registrations/:id/approve", ApproveRegistration) // Admin only
	webserver.ApiPOST("/providers/registrations/:id/reject", RejectRegistration)   // Admin only
	webserver.ApiGET("/providers/registrations/status/:id", GetRegistrationStatus)  // Public (auth skipped via JwtSkipPrefix)
}
