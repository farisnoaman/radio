# Phase 2: Provider Management Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build complete provider lifecycle management - registration, approval, provisioning, and CRUD operations.

**Architecture:** Public registration form → Admin approval workflow → Automated schema provisioning → Welcome email. RESTful APIs for all provider operations with tenant-aware access control.

**Tech Stack:** Echo framework, GORM, SMTP for emails, Go validation

---

## Task 1: Create Provider Registration API

**Files:**
- Create: `internal/adminapi/provider_registration.go`
- Create: `internal/adminapi/provider_registration_test.go`

**Step 1: Write failing tests for registration**

```go
// internal/adminapi/provider_registration_test.go
package adminapi

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/domain"
)

func TestCreateProviderRegistration(t *testing.T) {
    // Setup
    e := echo.New()
    reqBody := map[string]interface{}{
        "company_name":  "Test ISP LLC",
        "contact_name":  "John Doe",
        "email":         "john@testisp.com",
        "phone":         "+1234567890",
        "business_type": "WISP",
        "expected_users": 500,
        "expected_nas":  10,
        "country":       "US",
        "message":       "We want to join your platform",
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest(http.MethodPost, "/api/v1/public/register", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // Execute
    err := CreateProviderRegistration(c)

    // Assert
    if err != nil {
        t.Fatalf("Handler returned error: %v", err)
    }

    if rec.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", rec.Code)
    }

    var response map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &response)

    if response["status"] != "pending" {
        t.Errorf("Expected status 'pending', got '%v'", response["status"])
    }
}

func TestCreateProviderRegistrationValidation(t *testing.T) {
    tests := []struct {
        name       string
        payload    map[string]interface{}
        expectCode int
    }{
        {
            name: "missing company name",
            payload: map[string]interface{}{
                "contact_name": "John Doe",
                "email":        "john@test.com",
            },
            expectCode: http.StatusBadRequest,
        },
        {
            name: "invalid email",
            payload: map[string]interface{}{
                "company_name": "Test ISP",
                "contact_name": "John Doe",
                "email":        "not-an-email",
            },
            expectCode: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e := echo.New()
            body, _ := json.Marshal(tt.payload)

            req := httptest.NewRequest(http.MethodPost, "/api/v1/public/register", bytes.NewReader(body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)

            err := CreateProviderRegistration(c)

            if err == nil {
                t.Error("Expected validation error")
            }
        })
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/adminapi -run TestProviderRegistration -v`
Expected: FAIL with "undefined: CreateProviderRegistration"

**Step 3: Implement registration API**

```go
// internal/adminapi/provider_registration.go
package adminapi

import (
    "net/http"
    "regexp"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "gopkg.in/go-playground/validator.v9"
)

var (
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func registerProviderRegistrationRoutes() {
    // Public routes (no auth required)
    webserver.PublicPOST("/public/register", CreateProviderRegistration)
    webserver.PublicGET("/public/register/:id/status", GetRegistrationStatus)

    // Admin routes
    webserver.ApiGET("/admin/registrations", ListRegistrations)
    webserver.ApiGET("/admin/registrations/:id", GetRegistration)
    webserver.ApiPOST("/admin/registrations/:id/approve", ApproveRegistration)
    webserver.ApiPOST("/admin/registrations/:id/reject", RejectRegistration)
}

type CreateRegistrationRequest struct {
    CompanyName    string `json:"company_name" validate:"required"`
    ContactName    string `json:"contact_name" validate:"required"`
    Email          string `json:"email" validate:"required,email"`
    Phone          string `json:"phone"`
    Address        string `json:"address"`
    BusinessType   string `json:"business_type"`
    ExpectedUsers  int    `json:"expected_users" validate:"min=1"`
    ExpectedNas    int    `json:"expected_nas" validate:"min=1"`
    Country        string `json:"country"`
    Message        string `json:"message"`
}

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
        CompanyName:    req.CompanyName,
        ContactName:    req.ContactName,
        Email:          req.Email,
        Phone:          req.Phone,
        Address:        req.Address,
        BusinessType:   req.BusinessType,
        ExpectedUsers:  req.ExpectedUsers,
        ExpectedNas:    req.ExpectedNas,
        Country:        req.Country,
        Message:        req.Message,
        Status:         "pending",
    }

    if err := db.Create(registration).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create registration", nil)
    }

    // TODO: Send confirmation email

    return ok(c, registration)
}

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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/adminapi -run TestProviderRegistration -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adminapi/provider_registration.go internal/adminapi/provider_registration_test.go
git commit -m "feat(adminapi): add provider registration API with validation"
```

---

## Task 2: Implement Admin Registration Approval Workflow

**Files:**
- Modify: `internal/adminapi/provider_registration.go` (add approval functions)

**Step 1: Write tests for approval workflow**

```go
func TestApproveRegistration(t *testing.T) {
    // Setup: Create pending registration
    db := setupTestDB(t)
    registration := &domain.ProviderRegistration{
        CompanyName: "Test ISP",
        ContactName: "John Doe",
        Email:       "john@test.com",
        Status:      "pending",
    }
    db.Create(registration)

    // Create admin context
    e := echo.New()
    req := httptest.NewRequest(http.MethodPost, "/admin/registrations/1/approve", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetPath("/admin/registrations/:id/approve")
    c.SetParamNames("id")
    c.SetParamValues("1")

    // Execute
    err := ApproveRegistration(c)

    // Assert
    if err != nil {
        t.Fatalf("Handler returned error: %v", err)
    }

    // Verify provider created
    var provider domain.Provider
    db.First(&provider, 1)
    if provider.Name != "Test ISP" {
        t.Errorf("Provider name mismatch")
    }

    // Verify registration updated
    db.First(&registration, 1)
    if registration.Status != "approved" {
        t.Errorf("Registration status not updated")
    }
}

func TestRejectRegistration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    registration := &domain.ProviderRegistration{
        CompanyName: "Test ISP",
        Status:      "pending",
    }
    db.Create(registration)

    // Create request
    e := echo.New()
    reqBody := map[string]string{"reason": "Does not meet requirements"}
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/admin/registrations/1/reject", bytes.NewReader(body))
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetPath("/admin/registrations/:id/reject")
    c.SetParamNames("id")
    c.SetParamValues("1")

    // Execute
    err := RejectRegistration(c)

    // Assert
    if err != nil {
        t.Fatalf("Handler returned error: %v", err)
    }

    // Verify registration rejected
    db.First(&registration, 1)
    if registration.Status != "rejected" {
        t.Errorf("Registration not rejected")
    }

    if registration.RejectionReason != "Does not meet requirements" {
        t.Errorf("Rejection reason not saved")
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/adminapi -run TestApprove -v`
Expected: FAIL with "undefined: ApproveRegistration"

**Step 3: Implement approval workflow**

```go
// Add to internal/adminapi/provider_registration.go

type ApproveRegistrationRequest struct {
    ProviderCode string `json:"provider_code" validate:"required"`
    MaxUsers     int    `json:"max_users"`
    MaxNas       int    `json:"max_nas"`
    PlanID       int64  `json:"plan_id"`
}

type RejectRegistrationRequest struct {
    Reason string `json:"reason" validate:"required"`
}

func ApproveRegistration(c echo.Context) error {
    id := c.Param("id")
    if id == "" {
        return fail(c, http.StatusBadRequest, "MISSING_ID", "Registration ID is required", nil)
    }

    var req ApproveRegistrationRequest
    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
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

    // Create provider
    provider := &domain.Provider{
        Code:     req.ProviderCode,
        Name:     registration.CompanyName,
        Status:   "active",
        MaxUsers: req.MaxUsers,
        MaxNas:   req.MaxNas,
    }

    if err := db.Create(provider).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create provider", nil)
    }

    // Create provider schema
    migrator := migration.NewSchemaMigrator(db)
    if err := migrator.CreateProviderSchema(provider.ID); err != nil {
        // Rollback provider creation
        db.Delete(provider)
        return fail(c, http.StatusInternalServerError, "SCHEMA_FAILED", "Failed to create provider schema", nil)
    }

    // Create default admin operator for provider
    opr := &domain.SysOpr{
        TenantID:  provider.ID,
        Realname:  registration.ContactName,
        Email:     registration.Email,
        Mobile:    registration.Phone,
        Username:  "admin",
        Password:  generateRandomPassword(),
        Level:     "admin",
        Status:    "enabled",
    }
    db.Create(opr)

    // Update registration status
    now := time.Now()
    db.Model(&registration).Updates(map[string]interface{}{
        "status":      "approved",
        "reviewed_by": GetOperator(c).ID,
        "reviewed_at": &now,
    })

    // Send welcome email with credentials
    // TODO: Implement email service

    return ok(c, map[string]interface{}{
        "provider": provider,
        "admin":    opr,
    })
}

func RejectRegistration(c echo.Context) error {
    id := c.Param("id")
    if id == "" {
        return fail(c, http.StatusBadRequest, "MISSING_ID", "Registration ID is required", nil)
    }

    var req RejectRegistrationRequest
    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
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
        "reviewed_by":      GetOperator(c).ID,
        "reviewed_at":      &now,
    }).Error

    if err != nil {
        return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update registration", nil)
    }

    // Send rejection email
    // TODO: Implement email service

    return ok(c, map[string]string{"message": "Registration rejected"})
}

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
    query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&registrations)

    return paged(c, registrations, total, page, perPage)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/adminapi -run TestApprove -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adminapi/provider_registration.go
git commit -m "feat(adminapi): add registration approval workflow with schema provisioning"
```

---

## Task 3: Implement Provider CRUD APIs

**Files:**
- Modify: `internal/adminapi/providers.go` (enhance existing)

**Step 1: Write tests for provider CRUD**

```go
// Add to internal/adminapi/providers_test.go

func TestGetProvider(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    provider := &domain.Provider{
        Code:     "test-isp",
        Name:     "Test ISP",
        Status:   "active",
        MaxUsers: 1000,
        MaxNas:   100,
    }
    db.Create(provider)

    // Create request
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/providers/1", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetPath("/api/v1/providers/:id")
    c.SetParamNames("id")
    c.SetParamValues("1")

    // Execute
    err := GetProvider(c)

    // Assert
    if err != nil {
        t.Fatalf("Handler returned error: %v", err)
    }

    var response map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &response)

    if response["name"] != "Test ISP" {
        t.Errorf("Provider name mismatch")
    }
}

func TestUpdateProviderBranding(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    provider := &domain.Provider{
        Code:    "test-isp",
        Name:    "Test ISP",
        Status:  "active",
    }
    db.Create(provider)

    // Create request
    reqBody := map[string]string{
        "logo_url":        "https://test.com/logo.png",
        "primary_color":   "#007bff",
        "secondary_color": "#6c757d",
        "company_name":    "Test ISP LLC",
    }
    body, _ := json.Marshal(reqBody)

    e := echo.New()
    req := httptest.NewRequest(http.MethodPut, "/api/v1/providers/1", bytes.NewReader(body))
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetPath("/api/v1/providers/:id")
    c.SetParamNames("id")
    c.SetParamValues("1")

    // Execute
    err := UpdateProvider(c)

    // Assert
    if err != nil {
        t.Fatalf("Handler returned error: %v", err)
    }

    // Verify branding saved
    db.First(&provider, 1)
    branding, _ := provider.GetBranding()
    if branding.LogoURL != "https://test.com/logo.png" {
        t.Errorf("Branding not saved correctly")
    }
}
```

**Step 2: Run tests to verify they work with existing implementation**

Run: `go test ./internal/adminapi -run TestGetProvider -v`
Expected: Tests should work with existing `providers.go`

**Step 3: Enhance provider CRUD with branding and settings**

The existing `providers.go` already has basic CRUD. Ensure branding and settings are properly handled (they should be based on our reading).

**Step 4: Commit any improvements**

```bash
git add internal/adminapi/providers.go internal/adminapi/providers_test.go
git commit -m "refactor(adminapi): enhance provider CRUD with branding support"
```

---

## Task 4: Implement Email Service for Notifications

**Files:**
- Create: `internal/email/service.go`
- Create: `internal/email/templates.go`

**Step 1: Write email service interface**

```go
// internal/email/service.go
package email

import (
    "bytes"
    "html/template"
    "net/smtp"
)

type EmailService struct {
    smtpHost     string
    smtpPort     string
    smtpUsername string
    smtpPassword string
    fromAddress  string
}

type Email struct {
    To      []string
    Subject string
    Body    string
}

func NewEmailService(host, port, username, password, from string) *EmailService {
    return &EmailService{
        smtpHost:     host,
        smtpPort:     port,
        smtpUsername: username,
        smtpPassword: password,
        fromAddress:  from,
    }
}

func (es *EmailService) SendWelcomeEmail(providerName, to, username, password string) error {
    body, err := es.renderWelcomeTemplate(providerName, username, password)
    if err != nil {
        return err
    }

    email := &Email{
        To:      []string{to},
        Subject: "Welcome to RADIUS Platform",
        Body:    body,
    }

    return es.send(email)
}

func (es *EmailService) SendRejectionEmail(providerName, to, reason string) error {
    body, err := es.renderRejectionTemplate(providerName, reason)
    if err != nil {
        return err
    }

    email := &Email{
        To:      []string{to},
        Subject: "Registration Status Update",
        Body:    body,
    }

    return es.send(email)
}

func (es *EmailService) send(email *Email) error {
    auth := smtp.PlainAuth("", es.smtpUsername, es.smtpPassword, es.smtpHost)
    addr := es.smtpHost + ":" + es.smtpPort

    return smtp.SendMail(addr, auth, es.fromAddress, email.To, []byte(email.Body))
}

func (es *EmailService) renderWelcomeTemplate(providerName, username, password string) (string, error) {
    tmpl := `<!DOCTYPE html>
<html>
<body>
    <h2>Welcome to RADIUS Platform!</h2>
    <p>Dear {{.ProviderName}},</p>
    <p>Your registration has been approved. Here are your credentials:</p>
    <p>Username: {{.Username}}</p>
    <p>Password: {{.Password}}</p>
    <p>Please log in and change your password.</p>
</body>
</html>`

    t, err := template.New("welcome").Parse(tmpl)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    err = t.Execute(&buf, map[string]string{
        "ProviderName": providerName,
        "Username":     username,
        "Password":     password,
    })

    return buf.String(), err
}
```

**Step 2: Integrate email service into approval workflow**

Update `ApproveRegistration` and `RejectRegistration` to send emails.

**Step 3: Commit**

```bash
git add internal/email/
git commit -m "feat(email): add email service for registration notifications"
```

---

## Success Criteria

- ✅ Public can submit registration requests
- ✅ Admin can approve/reject registrations
- ✅ Provider schema auto-created on approval
- ✅ Welcome emails sent with credentials
- ✅ Provider CRUD operations functional
- ✅ Branding customization works
- ✅ Unit tests pass (≥80% coverage)
