// Package billing implements the postpaid monthly billing engine for ToughRADIUS.
//
// The billing engine runs as a daily cron job and handles three responsibilities:
//  1. Invoice Generation: Creates new invoices for postpaid users whose NextBillingDate has arrived.
//  2. Quota Reset: Clears accumulated accounting data so the user starts fresh for the new billing cycle.
//  3. Overdue Suspension: Marks unpaid invoices past their DueDate as "overdue" and suspends the user.
//
// The engine operates on RadiusUser records where BillingType == "postpaid" and
// SubscriptionStatus == "active". It uses database transactions to ensure atomicity
// when generating invoices and updating user billing dates.
//
// Configuration:
//   - billing.DueDateDays: Number of days after issue before an invoice is considered overdue (default 7).
//
// Concurrency: The billing engine methods are NOT safe for concurrent invocation.
// They are designed to be called once per schedule tick by the cron scheduler.
package billing

import (
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BillingEngine manages the postpaid billing lifecycle.
//
// It holds a reference to the GORM database handle and uses configuration
// values for cycle length and due-date offset. The zero-value is not usable;
// always create via NewBillingEngine.
type BillingEngine struct {
	db          *gorm.DB
	dueDateDays int // Days after IssueDate before invoice becomes overdue
}

// NewBillingEngine creates a new BillingEngine.
//
// Parameters:
//   - db: GORM database handle (must not be nil).
//   - dueDateDays: Days after invoice issue before it is overdue. Clamped to minimum 1.
//
// Returns:
//   - *BillingEngine: Ready-to-use billing engine instance.
func NewBillingEngine(db *gorm.DB, dueDateDays int) *BillingEngine {
	if dueDateDays < 1 {
		dueDateDays = 7
	}
	return &BillingEngine{
		db:          db,
		dueDateDays: dueDateDays,
	}
}

// ProcessDailyBillingCycle is the main entry point called by the cron scheduler.
// It generates invoices for users whose NextBillingDate has arrived, then
// enforces suspensions for overdue invoices.
//
// This method recovers from panics to avoid crashing the cron scheduler.
//
// Side effects:
//   - Creates Invoice records in the database.
//   - Updates RadiusUser.NextBillingDate and RadiusUser.SubscriptionStatus.
//   - Logs all operations via zap.
func (e *BillingEngine) ProcessDailyBillingCycle() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorf("billing engine panic: %v", err)
		}
	}()

	now := time.Now()

	generated, err := e.GenerateInvoices(now)
	if err != nil {
		zap.S().Errorf("invoice generation failed: %v", err)
	} else if generated > 0 {
		zap.S().Infof("Generated %d invoices", generated)
	}

	suspended, err := e.EnforceOverdueSuspensions(now)
	if err != nil {
		zap.S().Errorf("overdue suspension enforcement failed: %v", err)
	} else if suspended > 0 {
		zap.S().Infof("Suspended %d users for overdue invoices", suspended)
	}
}

// GenerateInvoices creates invoices for all postpaid users whose NextBillingDate <= now.
//
// For each qualifying user, it:
//  1. Creates an Invoice record covering the billing period.
//  2. Advances the user's NextBillingDate by cycleDays.
//
// Each user is processed in its own transaction so a failure on one user
// does not block others.
//
// Parameters:
//   - now: The current time (injectable for testing).
//
// Returns:
//   - int: Number of invoices successfully generated.
//   - error: First error encountered during the scan (individual user errors are logged, not returned).
func (e *BillingEngine) GenerateInvoices(now time.Time) (int, error) {
	var users []domain.RadiusUser
	err := e.db.Where(
		"billing_type = ? AND subscription_status = ? AND next_billing_date <= ?",
		domain.BillingTypePostpaid, domain.SubscriptionActive, now,
	).Find(&users).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query postpaid users for billing: %w", err)
	}

	generated := 0
	for _, user := range users {
		if err := e.generateInvoiceForUser(user, now); err != nil {
			zap.S().Errorf("failed to generate invoice for user %s: %v", user.Username, err)
			continue
		}
		generated++
	}
	return generated, nil
}

// GenerateEarlyInvoice triggers an immediate invoice generation for a user.
// It calculates usage from the last billing date to now, creates an invoice,
// and advances the NextBillingDate.
func (e *BillingEngine) GenerateEarlyInvoice(username string) error {
	var user domain.RadiusUser
	if err := e.db.Where("username = ? AND billing_type = ?", username, domain.BillingTypePostpaid).First(&user).Error; err != nil {
		return fmt.Errorf("user not found or not postpaid: %w", err)
	}

	return e.generateInvoiceForUser(user, time.Now())
}

// calculateUsageStats computes total data usage in GB and session count for a user in a time range.
func (e *BillingEngine) calculateUsageStats(username string, start, end time.Time) (float64, int64, error) {
	var stats struct {
		TotalBytes   int64
		SessionCount int64
	}
	// Sum input and output bytes and count sessions
	err := e.db.Model(&domain.RadiusAccounting{}).
		Where("username = ? AND acct_start_time >= ? AND acct_start_time < ?", username, start, end).
		Select("SUM(acct_input_total + acct_output_total) as total_bytes, COUNT(*) as session_count").
		Scan(&stats).Error

	if err != nil {
		return 0, 0, err
	}

	return float64(stats.TotalBytes) / (1024 * 1024 * 1024), stats.SessionCount, nil
}


// generateInvoiceForUser handles invoice creation for a single user within a transaction.
func (e *BillingEngine) generateInvoiceForUser(user domain.RadiusUser, now time.Time) error {
	tx := e.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	periodStart := user.NextBillingDate.AddDate(0, -1, 0)
	periodEnd := user.NextBillingDate

	// If this is an early manual trigger, adjust periods
	if now.Before(user.NextBillingDate) {
		periodEnd = now
	}

	// Calculate stats: Monthly Fee + (Usage GB * Price per GB)
	usageGB, sessionCount, err := e.calculateUsageStats(user.Username, periodStart, periodEnd)
	if err != nil {
		zap.S().Errorf("failed to calculate usage stats for user %s: %v", user.Username, err)
	}

	amount := user.MonthlyFee
	if user.PricePerGb > 0 {
		amount += usageGB * user.PricePerGb
	}

	invoice := domain.Invoice{
		ID:                 common.UUIDint64(),
		Username:           user.Username,
		ProfileID:          user.ProfileId,
		Amount:             amount,
		BaseAmount:         user.MonthlyFee,
		UsageGb:            usageGB,
		PricePerGb:         user.PricePerGb,
		SessionCount:       sessionCount,
		Currency:           "USD", // Default currency
		IssueDate:          now,
		DueDate:            now.AddDate(0, 0, e.dueDateDays),
		Status:             domain.InvoiceUnpaid,
		BillingPeriodStart: periodStart,
		BillingPeriodEnd:   periodEnd,
		Remark:             fmt.Sprintf("Consumption: %.2f GB, sessions: %d", usageGB, sessionCount),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Advance the next billing date by one calendar month
	newNextBilling := user.NextBillingDate.AddDate(0, 1, 0)
	if err := tx.Model(&domain.RadiusUser{}).Where("id = ?", user.ID).
		Update("next_billing_date", newNextBilling).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to advance next_billing_date: %w", err)
	}

	return tx.Commit().Error
}

// EnforceOverdueSuspensions finds unpaid invoices past their DueDate, marks them
// as "overdue", and suspends the corresponding user's subscription.
//
// Suspended users will be rejected by the RADIUS authentication handler until
// their invoice is paid and the admin reactivates them.
//
// Parameters:
//   - now: The current time (injectable for testing).
//
// Returns:
//   - int: Number of users suspended.
//   - error: Database error during the overdue scan (individual errors are logged).
func (e *BillingEngine) EnforceOverdueSuspensions(now time.Time) (int, error) {
	// Step 1: Mark unpaid invoices past due_date as overdue
	result := e.db.Model(&domain.Invoice{}).
		Where("status = ? AND due_date < ?", domain.InvoiceUnpaid, now).
		Update("status", domain.InvoiceOverdue)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark overdue invoices: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		zap.S().Infof("Marked %d invoices as overdue", result.RowsAffected)
	}

	// Step 2: Find distinct usernames with overdue invoices
	var usernames []string
	err := e.db.Model(&domain.Invoice{}).
		Where("status = ?", domain.InvoiceOverdue).
		Distinct("username").
		Pluck("username", &usernames).Error
	if err != nil {
		return 0, fmt.Errorf("failed to find users with overdue invoices: %w", err)
	}

	// Step 3: Suspend users who have overdue invoices and are still active
	suspended := 0
	for _, username := range usernames {
		err := e.db.Model(&domain.RadiusUser{}).
			Where("username = ? AND billing_type = ? AND subscription_status = ?",
				username, domain.BillingTypePostpaid, domain.SubscriptionActive).
			Update("subscription_status", domain.SubscriptionSuspended).Error
		if err != nil {
			zap.S().Errorf("failed to suspend user %s: %v", username, err)
			continue
		}
		suspended++
	}
	return suspended, nil
}

// PayInvoice marks an invoice as paid and reactivates the user's subscription
// if all their invoices are now paid.
//
// This is called from the Admin API when the operator collects payment.
//
// Parameters:
//   - invoiceID: The ID of the invoice to mark as paid.
//
// Returns:
//   - error: If the invoice is not found or the database update fails.
//
// Side effects:
//   - Updates Invoice.Status to "paid" and sets Invoice.PaidAt.
//   - If the user has no remaining overdue/unpaid invoices, reactivates their subscription.
func (e *BillingEngine) PayInvoice(invoiceID int64) error {
	tx := e.db.Begin()

	var invoice domain.Invoice
	if err := tx.Where("id = ?", invoiceID).First(&invoice).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("invoice not found: %w", err)
	}

	if invoice.Status == domain.InvoicePaid {
		tx.Rollback()
		return nil // Already paid, no-op
	}

	now := time.Now()
	if err := tx.Model(&invoice).Updates(map[string]interface{}{
		"status":     domain.InvoicePaid,
		"paid_at":    now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	// Check if user has any remaining unpaid/overdue invoices
	var remaining int64
	tx.Model(&domain.Invoice{}).
		Where("username = ? AND status IN (?, ?)", invoice.Username, domain.InvoiceUnpaid, domain.InvoiceOverdue).
		Count(&remaining)

	// If no remaining unpaid invoices, reactivate the user
	if remaining == 0 {
		tx.Model(&domain.RadiusUser{}).
			Where("username = ? AND billing_type = ? AND subscription_status = ?",
				invoice.Username, domain.BillingTypePostpaid, domain.SubscriptionSuspended).
			Update("subscription_status", domain.SubscriptionActive)
	}

	return tx.Commit().Error
}
