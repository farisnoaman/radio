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
	db           *gorm.DB
	quotaService *quota.QuotaService
	emailService interface{} // email.Service - placeholder for future implementation
}

func NewBillingEngine(db *gorm.DB, quotaSvc *quota.QuotaService, emailSvc interface{}) *BillingEngine {
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
		invoice, err := be.GenerateInvoiceForSubscription(ctx, sub)
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

// GenerateInvoiceForSubscription calculates invoice for a subscription
func (be *BillingEngine) GenerateInvoiceForSubscription(
	ctx context.Context,
	sub domain.ProviderSubscription,
) (*domain.ProviderInvoice, error) {
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

	invoice := &domain.ProviderInvoice{
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
func (be *BillingEngine) sendInvoiceEmail(tenantID int64, invoice *domain.ProviderInvoice) {
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
