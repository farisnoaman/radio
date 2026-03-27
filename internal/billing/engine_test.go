package billing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/quota"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate tables
	err = db.AutoMigrate(
		&domain.BillingPlan{},
		&domain.ProviderSubscription{},
		&domain.ProviderInvoice{},
		&domain.RadiusUser{},
		&domain.ProviderQuota{},
		&domain.ProviderUsage{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	return db
}

func TestGenerateInvoice(t *testing.T) {
	db := setupTestDB(t)

	// Create quota service for billing engine (without cache for testing)
	quotaService := quota.NewQuotaService(db, nil)
	engine := NewBillingEngine(db, quotaService, nil)

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
	invoice, err := engine.GenerateInvoiceForSubscription(context.Background(), *subscription)
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

	if invoice.CurrentUsers != 150 {
		t.Errorf("Expected 150 users, got %d", invoice.CurrentUsers)
	}
}
