package domain

import (
	"testing"
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
	invoice := &ProviderInvoice{}
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
