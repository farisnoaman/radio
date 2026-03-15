package billing

import (
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a fresh in-memory SQLite database for each test,
// auto-migrating the required tables. This avoids CGO by using the
// pure-Go glebarez/sqlite driver (imported via the gorm sqlite wrapper).
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	err = db.AutoMigrate(&domain.RadiusUser{}, &domain.Invoice{}, &domain.RadiusProfile{})
	if err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	return db
}

// createTestPostpaidUser is a helper that inserts a postpaid user with a known NextBillingDate.
func createTestPostpaidUser(t *testing.T, db *gorm.DB, username string, monthlyFee float64, nextBilling time.Time) {
	t.Helper()
	user := domain.RadiusUser{
		ID:                 1,
		Username:           username,
		Password:           "test",
		Status:             "enabled",
		BillingType:        domain.BillingTypePostpaid,
		SubscriptionStatus: domain.SubscriptionActive,
		NextBillingDate:    nextBilling,
		MonthlyFee:         monthlyFee,
		ProfileId:          100,
		ExpireTime:         time.Now().AddDate(10, 0, 0), // Far future
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
}

func TestGenerateInvoices_UserDueForBilling_ShouldCreateInvoice(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	// User whose billing date is in the past (due for invoice)
	pastBilling := time.Now().AddDate(0, 0, -1)
	createTestPostpaidUser(t, db, "postpaid_user1", 25.00, pastBilling)

	generated, err := engine.GenerateInvoices(time.Now())
	if err != nil {
		t.Fatalf("GenerateInvoices failed: %v", err)
	}
	if generated != 1 {
		t.Errorf("expected 1 invoice generated, got %d", generated)
	}

	// Verify invoice was created in the database
	var invoices []domain.Invoice
	db.Where("username = ?", "postpaid_user1").Find(&invoices)
	if len(invoices) != 1 {
		t.Fatalf("expected 1 invoice in DB, got %d", len(invoices))
	}
	if invoices[0].Amount != 25.00 {
		t.Errorf("expected amount 25.00, got %f", invoices[0].Amount)
	}
	if invoices[0].Status != domain.InvoiceUnpaid {
		t.Errorf("expected status 'unpaid', got '%s'", invoices[0].Status)
	}

	// Verify NextBillingDate was advanced by cycleDays
	var updatedUser domain.RadiusUser
	db.Where("username = ?", "postpaid_user1").First(&updatedUser)
	expectedNext := pastBilling.AddDate(0, 1, 0)
	if !updatedUser.NextBillingDate.Equal(expectedNext) {
		t.Errorf("expected NextBillingDate %v, got %v", expectedNext, updatedUser.NextBillingDate)
	}
}

func TestGenerateInvoices_UserNotYetDue_ShouldNotCreateInvoice(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	// User whose billing date is in the future (not yet due)
	futureBilling := time.Now().AddDate(0, 0, 10)
	createTestPostpaidUser(t, db, "postpaid_future", 15.00, futureBilling)

	generated, err := engine.GenerateInvoices(time.Now())
	if err != nil {
		t.Fatalf("GenerateInvoices failed: %v", err)
	}
	if generated != 0 {
		t.Errorf("expected 0 invoices generated, got %d", generated)
	}
}

func TestGenerateInvoices_PrepaidUser_ShouldBeIgnored(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	// Create a prepaid user — should never get an invoice
	user := domain.RadiusUser{
		ID:          2,
		Username:    "prepaid_user",
		Password:    "test",
		Status:      "enabled",
		BillingType: domain.BillingTypePrepaid,
		ExpireTime:  time.Now().AddDate(1, 0, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(&user)

	generated, err := engine.GenerateInvoices(time.Now())
	if err != nil {
		t.Fatalf("GenerateInvoices failed: %v", err)
	}
	if generated != 0 {
		t.Errorf("expected 0 invoices for prepaid user, got %d", generated)
	}
}

func TestEnforceOverdueSuspensions_OverdueInvoice_ShouldSuspendUser(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	createTestPostpaidUser(t, db, "overdue_user", 30.00, time.Now().AddDate(0, 1, 0))

	// Manually create an overdue invoice: due_date is in the past
	invoice := domain.Invoice{
		ID:                 100,
		Username:           "overdue_user",
		Amount:             30.00,
		IssueDate:          time.Now().AddDate(0, 0, -10),
		DueDate:            time.Now().AddDate(0, 0, -3), // 3 days past due
		Status:             domain.InvoiceUnpaid,
		BillingPeriodStart: time.Now().AddDate(0, 0, -40),
		BillingPeriodEnd:   time.Now().AddDate(0, 0, -10),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	db.Create(&invoice)

	suspended, err := engine.EnforceOverdueSuspensions(time.Now())
	if err != nil {
		t.Fatalf("EnforceOverdueSuspensions failed: %v", err)
	}
	if suspended != 1 {
		t.Errorf("expected 1 user suspended, got %d", suspended)
	}

	// Verify the invoice is now overdue
	var updatedInvoice domain.Invoice
	db.First(&updatedInvoice, 100)
	if updatedInvoice.Status != domain.InvoiceOverdue {
		t.Errorf("expected invoice status 'overdue', got '%s'", updatedInvoice.Status)
	}

	// Verify the user is now suspended
	var updatedUser domain.RadiusUser
	db.Where("username = ?", "overdue_user").First(&updatedUser)
	if updatedUser.SubscriptionStatus != domain.SubscriptionSuspended {
		t.Errorf("expected subscription_status 'suspended', got '%s'", updatedUser.SubscriptionStatus)
	}
}

func TestEnforceOverdueSuspensions_NoOverdueInvoices_NothingSuspended(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	createTestPostpaidUser(t, db, "good_user", 10.00, time.Now().AddDate(0, 1, 0))

	// Invoice not yet due
	invoice := domain.Invoice{
		ID:        200,
		Username:  "good_user",
		Amount:    10.00,
		IssueDate: time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 7), // Still 7 days to pay
		Status:    domain.InvoiceUnpaid,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&invoice)

	suspended, err := engine.EnforceOverdueSuspensions(time.Now())
	if err != nil {
		t.Fatalf("EnforceOverdueSuspensions failed: %v", err)
	}
	if suspended != 0 {
		t.Errorf("expected 0 suspensions, got %d", suspended)
	}
}

func TestPayInvoice_ShouldMarkPaidAndReactivateUser(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	// Create a suspended user with an overdue invoice
	user := domain.RadiusUser{
		ID:                 3,
		Username:           "suspended_user",
		Password:           "test",
		Status:             "enabled",
		BillingType:        domain.BillingTypePostpaid,
		SubscriptionStatus: domain.SubscriptionSuspended,
		MonthlyFee:         20.00,
		ExpireTime:         time.Now().AddDate(10, 0, 0),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	db.Create(&user)

	invoice := domain.Invoice{
		ID:        300,
		Username:  "suspended_user",
		Amount:    20.00,
		IssueDate: time.Now().AddDate(0, 0, -15),
		DueDate:   time.Now().AddDate(0, 0, -8),
		Status:    domain.InvoiceOverdue,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&invoice)

	err := engine.PayInvoice(300)
	if err != nil {
		t.Fatalf("PayInvoice failed: %v", err)
	}

	// Verify invoice is paid
	var updatedInvoice domain.Invoice
	db.First(&updatedInvoice, 300)
	if updatedInvoice.Status != domain.InvoicePaid {
		t.Errorf("expected invoice status 'paid', got '%s'", updatedInvoice.Status)
	}
	if updatedInvoice.PaidAt.IsZero() {
		t.Error("expected PaidAt to be set")
	}

	// Verify user is reactivated (no remaining unpaid/overdue invoices)
	var updatedUser domain.RadiusUser
	db.Where("username = ?", "suspended_user").First(&updatedUser)
	if updatedUser.SubscriptionStatus != domain.SubscriptionActive {
		t.Errorf("expected subscription_status 'active' after payment, got '%s'", updatedUser.SubscriptionStatus)
	}
}

func TestPayInvoice_AlreadyPaid_ShouldBeNoOp(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	invoice := domain.Invoice{
		ID:        400,
		Username:  "paid_user",
		Amount:    10.00,
		Status:    domain.InvoicePaid,
		PaidAt:    time.Now().AddDate(0, 0, -1),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&invoice)

	err := engine.PayInvoice(400)
	if err != nil {
		t.Fatalf("PayInvoice on already-paid should not error, got: %v", err)
	}
}

func TestNewBillingEngine_ClampMinimumValues(t *testing.T) {
	db := setupTestDB(t)

	engine := NewBillingEngine(db, 0)
	if engine.dueDateDays != 7 {
		t.Errorf("expected dueDateDays clamped to 7, got %d", engine.dueDateDays)
	}
}

func TestGenerateEarlyInvoice(t *testing.T) {
	db := setupTestDB(t)
	engine := NewBillingEngine(db, 7)

	// Create user with future billing date
	futureBilling := time.Now().AddDate(0, 1, 0)
	createTestPostpaidUser(t, db, "ali", 50.00, futureBilling)

	err := engine.GenerateEarlyInvoice("ali")
	if err != nil {
		t.Fatalf("GenerateEarlyInvoice failed: %v", err)
	}

	// Verify invoice
	var invoices []domain.Invoice
	db.Where("username = ?", "ali").Find(&invoices)
	if len(invoices) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(invoices))
	}

	// Verify date adjustment
	var user domain.RadiusUser
	db.Where("username = ?", "ali").First(&user)
	expectedDate := futureBilling.AddDate(0, 1, 0)
	if !user.NextBillingDate.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, user.NextBillingDate)
	}
}

