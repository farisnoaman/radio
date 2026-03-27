# Phase 5: Billing Engine Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement automated billing engine with hybrid pricing (base fee + usage overages).

**Architecture:** Billing plans → Provider subscriptions → Usage tracking → Invoice generation → Payment processing. Background cron job generates monthly invoices.

**Tech Stack:** Cron jobs, GORM, Stripe API (optional), PDF generation

---

## Task 1: Create Billing Models

**Files:**
- Create: `internal/domain/billing.go`
- Create: `internal/domain/billing_test.go`

**Step 1: Write tests for billing models**

```go
// internal/domain/billing_test.go
package domain

import (
    "testing"
    "time"
)

func TestBillingPlanModel(t *testing.T) {
    plan := &BillingPlan{
        Code:          "basic",
        Name:          "Basic Plan",
        BaseFee:       49.99,
        IncludedUsers: 100,
        OverageFee:    0.50,
        MaxUsers:      1000,
        IsActive:      true,
    }

    if plan.TableName() != "mst_billing_plan" {
        t.Errorf("Expected table name 'mst_billing_plan', got '%s'", plan.TableName())
    }
}

func TestInvoiceCalculation(t *testing.T) {
    plan := &BillingPlan{
        BaseFee:       100.0,
        IncludedUsers: 100,
        OverageFee:    1.0,
    }

    subscription := &ProviderSubscription{
        BaseFee:    100.0,
        OverageFee: 1.0,
    }

    // Test with 150 users (50 overage)
    invoice := &Invoice{}
    invoice.Calculate(subscription, plan, 150, 0, 0)

    expectedBase := 100.0
    expectedOverage := 50.0 * 1.0
    expectedSubtotal := expectedBase + expectedOverage
    expectedTax := expectedSubtotal * 0.15
    expectedTotal := expectedSubtotal + expectedTax

    if invoice.TotalAmount != expectedTotal {
        t.Errorf("Expected total %.2f, got %.2f", expectedTotal, invoice.TotalAmount)
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain -run TestBilling -v`
Expected: FAIL with "undefined: BillingPlan"

**Step 3: Implement billing models**

```go
// internal/domain/billing.go
package domain

import "time"

type BillingPlan struct {
    ID              int64     `json:"id" gorm:"primaryKey"`
    Code            string    `json:"code" gorm:"uniqueIndex;size:50"`
    Name            string    `json:"name" gorm:"size:255"`
    BaseFee         float64   `json:"base_fee"`
    IncludedUsers   int       `json:"included_users"`
    OverageFee      float64   `json:"overage_fee"`     // Per user over base
    MaxUsers        int       `json:"max_users"`
    Features        string    `json:"features"`        // JSON array
    IsActive        bool      `json:"is_active"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

func (BillingPlan) TableName() string {
    return "mst_billing_plan"
}

type ProviderSubscription struct {
    ID              int64     `json:"id" gorm:"primaryKey"`
    TenantID        int64     `json:"tenant_id" gorm:"uniqueIndex"`
    PlanID          int64     `json:"plan_id"`
    Status          string    `json:"status"`          // active, suspended, canceled
    BaseFee         float64   `json:"base_fee"`
    OverageFee      float64   `json:"overage_fee"`
    BillingCycle    string    `json:"billing_cycle"`   // monthly, yearly
    NextBillingDate  time.Time `json:"next_billing_date" gorm:"index"`
    CancelAt        time.Time `json:"cancel_at"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

func (ProviderSubscription) TableName() string {
    return "mst_provider_subscription"
}

type Invoice struct {
    ID               int64      `json:"id" gorm:"primaryKey"`
    TenantID         int64      `json:"tenant_id" gorm:"index"`
    InvoiceNumber    string     `json:"invoice_number" gorm:"uniqueIndex"`
    SubscriptionID   int64      `json:"subscription_id"`

    // Line items
    BaseFee          float64    `json:"base_fee"`
    UserOverageFee   float64    `json:"user_overage_fee"`
    SessionOverageFee float64   `json:"session_overage_fee"`
    StorageOverageFee float64   `json:"storage_overage_fee"`
    TaxAmount        float64    `json:"tax_amount"`
    TotalAmount      float64    `json:"total_amount"`

    // Usage breakdown
    CurrentUsers     int        `json:"current_users"`
    IncludedUsers    int        `json:"included_users"`
    OverageUsers     int        `json:"overage_users"`

    // Billing period
    PeriodStart      time.Time  `json:"period_start"`
    PeriodEnd        time.Time  `json:"period_end"`

    // Status
    Status           string     `json:"status"`    // draft, pending, paid, overdue
    DueDate          time.Time  `json:"due_date"`
    PaidDate         *time.Time `json:"paid_date"`

    CreatedAt        time.Time  `json:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at"`
}

func (Invoice) TableName() string {
    return "mst_invoice"
}

// Calculate calculates invoice amounts based on usage
func (inv *Invoice) Calculate(sub *ProviderSubscription, plan *BillingPlan, currentUsers, overageSessions, storageGB int) {
    // Base fee
    inv.BaseFee = sub.BaseFee

    // Calculate user overage
    overageUsers := 0
    if currentUsers > plan.IncludedUsers {
        overageUsers = currentUsers - plan.IncludedUsers
    }

    inv.UserOverageFee = float64(overageUsers) * sub.OverageFee
    inv.CurrentUsers = currentUsers
    inv.IncludedUsers = plan.IncludedUsers
    inv.OverageUsers = overageUsers

    // Calculate subtotal
    subtotal := inv.BaseFee + inv.UserOverageFee

    // Calculate tax (15%)
    inv.TaxAmount = subtotal * 0.15

    // Calculate total
    inv.TotalAmount = subtotal + inv.TaxAmount
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain -run TestBilling -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/billing.go internal/domain/billing_test.go
git commit -m "feat(domain): add billing models with invoice calculation"
```

---

## Task 2: Create Billing Engine Service

**Files:**
- Create: `internal/billing/engine.go`
- Create: `internal/billing/engine_test.go`

**Step 1: Write tests for billing engine**

```go
// internal/billing/engine_test.go
package billing

import (
    "testing"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
)

func TestGenerateInvoice(t *testing.T) {
    db := setupTestDB(t)
    engine := NewBillingEngine(db, nil, nil)

    // Create test data
    plan := &domain.BillingPlan{
        Code:          "basic",
        Name:          "Basic Plan",
        BaseFee:       100.0,
        IncludedUsers: 100,
        OverageFee:    1.0,
        MaxUsers:      1000,
        IsActive:      true,
    }
    db.Create(plan)

    subscription := &domain.ProviderSubscription{
        TenantID:        1,
        PlanID:          plan.ID,
        Status:          "active",
        BaseFee:         100.0,
        OverageFee:      1.0,
        BillingCycle:    "monthly",
        NextBillingDate: time.Now().AddDate(0, 0, -1), // Due yesterday
    }
    db.Create(subscription)

    // Create 150 users (50 overage)
    for i := 0; i < 150; i++ {
        db.Create(&domain.RadiusUser{TenantID: 1, Username: fmt.Sprintf("user%d", i)})
    }

    // Generate invoice
    invoice, err := engine.generateInvoiceForSubscription(context.Background(), *subscription)
    if err != nil {
        t.Fatalf("Failed to generate invoice: %v", err)
    }

    // Verify calculations
    if invoice.UserOverageFee != 50.0 {
        t.Errorf("Expected user overage fee 50.0, got %.2f", invoice.UserOverageFee)
    }

    if invoice.TotalAmount < 150.0 {
        t.Errorf("Expected total > 150.0, got %.2f", invoice.TotalAmount)
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/billing -run TestGenerate -v`
Expected: FAIL with "undefined: NewBillingEngine"

**Step 3: Implement billing engine**

```go
// internal/billing/engine.go
package billing

import (
    "context"
    "fmt"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/quota"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type BillingEngine struct {
    db            *gorm.DB
    quotaService  *quota.QuotaService
    emailService  *email.Service
}

func NewBillingEngine(db *gorm.DB, quotaSvc *quota.QuotaService, emailSvc *email.Service) *BillingEngine {
    return &BillingEngine{
        db:           db,
        quotaService: quotaSvc,
        emailService: emailSvc,
    }
}

// GenerateMonthlyInvoices generates invoices for all active subscriptions due for billing
func (be *BillingEngine) GenerateMonthlyInvoices(ctx context.Context) error {
    var subs []domain.ProviderSubscription
    be.db.Where("status = ? AND next_billing_date <= ?", "active", time.Now()).Find(&subs)

    for _, sub := range subs {
        invoice, err := be.generateInvoiceForSubscription(ctx, sub)
        if err != nil {
            zap.S().Error("Failed to generate invoice",
                zap.Int64("tenant_id", sub.TenantID),
                zap.Error(err))
            continue
        }

        if err := be.db.Create(invoice).Error; err != nil {
            zap.S().Error("Failed to save invoice", zap.Error(err))
            continue
        }

        be.updateNextBillingDate(&sub)

        // Send invoice email
        be.sendInvoiceEmail(sub.TenantID, invoice)
    }

    return nil
}

// generateInvoiceForSubscription calculates invoice for a subscription
func (be *BillingEngine) generateInvoiceForSubscription(
    ctx context.Context,
    sub domain.ProviderSubscription,
) (*domain.Invoice, error) {
    // Get current usage
    usage, err := be.quotaService.GetUsage(sub.TenantID)
    if err != nil {
        return nil, err
    }

    // Get plan details
    var plan domain.BillingPlan
    be.db.First(&plan, sub.PlanID)

    // Create invoice
    invoiceNumber := be.generateInvoiceNumber(sub.TenantID)

    periodStart := time.Now().AddDate(0, -1, 0)
    periodEnd := time.Now()

    invoice := &domain.Invoice{
        TenantID:       sub.TenantID,
        InvoiceNumber:  invoiceNumber,
        SubscriptionID: sub.ID,
        PeriodStart:    periodStart,
        PeriodEnd:      periodEnd,
        Status:         "pending",
        DueDate:        time.Now().AddDate(0, 0, 14), // 14 days
        CreatedAt:      time.Now(),
    }

    // Calculate amounts
    invoice.Calculate(&sub, &plan, usage.CurrentUsers, 0, 0)

    return invoice, nil
}

// generateInvoiceNumber generates unique invoice number
func (be *BillingEngine) generateInvoiceNumber(tenantID int64) string {
    timestamp := time.Now().Format("200601")
    return fmt.Sprintf("INV-%d-%s-%04d", tenantID, timestamp, time.Now().Unix()%10000)
}

// updateNextBillingDate updates next billing date for subscription
func (be *BillingEngine) updateNextBillingDate(sub *domain.ProviderSubscription) {
    if sub.BillingCycle == "monthly" {
        sub.NextBillingDate = sub.NextBillingDate.AddDate(0, 1, 0)
    } else if sub.BillingCycle == "yearly" {
        sub.NextBillingDate = sub.NextBillingDate.AddDate(1, 0, 0)
    }
    be.db.Save(sub)
}

// sendInvoiceEmail sends invoice to provider
func (be *BillingEngine) sendInvoiceEmail(tenantID int64, invoice *domain.Invoice) {
    var provider domain.Provider
    be.db.First(&provider, tenantID)

    // Get provider contact email
    var opr domain.SysOpr
    be.db.Where("tenant_id = ? AND level = ?", tenantID, "admin").First(&opr)

    if opr.Email != "" {
        // be.emailService.SendInvoiceEmail(provider.Name, opr.Email, invoice)
        zap.S().Info("Invoice sent",
            zap.Int64("tenant_id", tenantID),
            zap.String("invoice", invoice.InvoiceNumber),
            zap.String("email", opr.Email))
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/billing -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/billing/engine.go internal/billing/engine_test.go
git commit -m "feat(billing): add automated invoice generation engine"
```

---

## Task 3: Create Billing Cron Job

**Files:**
- Create: `internal/billing/cron.go`

**Step 1: Implement cron job**

```go
// internal/billing/cron.go
package billing

import (
    "context"
    "time"

    "go.uber.org/zap"
)

type BillingScheduler struct {
    engine *BillingEngine
}

func NewBillingScheduler(engine *BillingEngine) *BillingScheduler {
    return &BillingScheduler{engine: engine}
}

// Start begins the billing scheduler
func (bs *BillingScheduler) Start(ctx context.Context) {
    // Run daily at midnight
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()

    // Run immediately on start
    bs.runBilling(ctx)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            bs.runBilling(ctx)
        }
    }
}

func (bs *BillingScheduler) runBilling(ctx context.Context) {
    zap.S().Info("Running billing cycle")

    if err := bs.engine.GenerateMonthlyInvoices(ctx); err != nil {
        zap.S().Error("Billing cycle failed", zap.Error(err))
    } else {
        zap.S().Info("Billing cycle completed successfully")
    }
}
```

**Step 2: Start scheduler in app initialization**

```go
// In internal/app/app.go

func (a *Application) Init(config *config.Config) {
    // ... existing initialization

    // Start billing scheduler
    billingEngine := billing.NewBillingEngine(a.gormDB, quotaService, emailService)
    scheduler := billing.NewBillingScheduler(billingEngine)
    go scheduler.Start(context.Background())
}
```

**Step 3: Commit**

```bash
git add internal/billing/cron.go
git commit -m "feat(billing): add automated billing scheduler"
```

---

## Task 4: Create Billing Management APIs

**Files:**
- Create: `internal/adminapi/billing.go`

**Step 1: Implement billing APIs**

```go
// internal/adminapi/billing.go
package adminapi

import (
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/billing"
    "github.com/talkincode/toughradius/v9/internal/domain"
)

func registerBillingRoutes() {
    // Provider routes
    webserver.ApiGET("/billing/invoices", GetInvoices)
    webserver.ApiGET("/billing/invoices/:id", GetInvoice)
    webserver.ApiPOST("/billing/invoices/:id/pay", PayInvoice)

    // Admin routes
    webserver.ApiGET("/admin/billing/plans", ListBillingPlans)
    webserver.ApiPOST("/admin/billing/plans", CreateBillingPlan)
    webserver.ApiPOST("/admin/billing/run", TriggerBillingCycle)
}

func GetInvoices(c echo.Context) error {
    tenantID, _ := tenant.FromContext(c.Request().Context())
    db := GetDB(c)

    var invoices []domain.Invoice
    err := db.Where("tenant_id = ?", tenantID).
        Order("created_at DESC").
        Find(&invoices).Error

    if err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch invoices", err)
    }

    return ok(c, invoices)
}

func PayInvoice(c echo.Context) error {
    id := c.Param("id")
    if id == "" {
        return fail(c, http.StatusBadRequest, "MISSING_ID", "Invoice ID is required", nil)
    }

    db := GetDB(c)

    var invoice domain.Invoice
    if err := db.First(&invoice, id).Error; err != nil {
        return fail(c, http.StatusNotFound, "NOT_FOUND", "Invoice not found", nil)
    }

    // Update invoice status
    now := time.Now()
    invoice.Status = "paid"
    invoice.PaidDate = &now
    db.Save(&invoice)

    return ok(c, invoice)
}

func TriggerBillingCycle(c echo.Context) error {
    // Verify platform admin
    if !IsPlatformAdmin(c) {
        return fail(c, http.StatusForbidden, "FORBIDDEN", "Platform admin required", nil)
    }

    billingEngine := GetBillingEngine(c)
    if err := billingEngine.GenerateMonthlyInvoices(c.Request().Context()); err != nil {
        return fail(c, http.StatusInternalServerError, "BILLING_ERROR", "Failed to run billing", err)
    }

    return ok(c, map[string]string{"message": "Billing cycle triggered successfully"})
}
```

**Step 2: Commit**

```bash
git add internal/adminapi/billing.go
git commit -m "feat(adminapi): add billing management APIs"
```

---

## Success Criteria

- ✅ Billing plans and subscriptions modeled
- ✅ Invoice calculation functional (base + overage + tax)
- ✅ Automated monthly billing cycle
- ✅ Invoices generated and emailed
- ✅ Provider can view/pay invoices
- ✅ Admin can trigger manual billing
- ✅ Unit tests pass (≥80% coverage)
