package domain

import "time"

type BillingPlan struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	Code          string    `json:"code" gorm:"uniqueIndex;size:50"`
	Name          string    `json:"name" gorm:"size:255"`
	BaseFee       float64   `json:"base_fee"`
	IncludedUsers int       `json:"included_users"`
	OverageFee    float64   `json:"overage_fee"`   // Per user over base
	MaxUsers      int       `json:"max_users"`
	Features      string    `json:"features"`      // JSON array
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (BillingPlan) TableName() string {
	return "mst_billing_plan"
}

type ProviderSubscription struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	TenantID       int64     `json:"tenant_id" gorm:"uniqueIndex"`
	PlanID         int64     `json:"plan_id"`
	Status         string    `json:"status"`         // active, suspended, canceled
	BaseFee        float64   `json:"base_fee"`
	OverageFee     float64   `json:"overage_fee"`
	BillingCycle   string    `json:"billing_cycle"`   // monthly, yearly
	NextBillingDate time.Time `json:"next_billing_date" gorm:"index"`
	CancelAt       time.Time `json:"cancel_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ProviderSubscription) TableName() string {
	return "mst_provider_subscription"
}

type ProviderInvoice struct {
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

func (ProviderInvoice) TableName() string {
	return "mst_provider_invoice"
}

// Calculate calculates invoice amounts based on usage
func (inv *ProviderInvoice) Calculate(sub *ProviderSubscription, plan *BillingPlan, currentUsers, overageSessions, storageGB int) {
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
