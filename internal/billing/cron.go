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
