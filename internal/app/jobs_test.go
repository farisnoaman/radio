package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

func setupTestAppWithDB(t *testing.T) (*Application, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(domain.Tables...)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestJobApp",
			Location: "UTC",
		},
	}
	app := NewApplication(cfg)
	app.gormDB = db
	return app, db
}

func TestSchedSubscriptionRenewalTask(t *testing.T) {
	app, db := setupTestAppWithDB(t)

	// 1. Create Product
	product := domain.Product{
		Name:            "SubProduct",
		Price:           10.0,
		CostPrice:       8.0,
		ValiditySeconds: 3600,
	}
	db.Create(&product)

	// 2. Create Agent and Wallet
	agent := domain.SysOpr{Username: "testagent", Level: "agent"}
	db.Create(&agent)
	wallet := domain.AgentWallet{AgentID: agent.ID, Balance: 100.0}
	db.Create(&wallet)

	// 3. Create Voucher and RadiusUser
	voucher := domain.Voucher{
		Code:    "SUBTEST001",
		Status:  "active",
		AgentID: agent.ID,
	}
	db.Create(&voucher)

	now := time.Now().UTC()
	user := domain.RadiusUser{
		Username:   voucher.Code,
		ExpireTime: now.Add(1 * time.Hour),
	}
	db.Create(&user)

	// 4. Create Subscription due for renewal
	sub := domain.VoucherSubscription{
		VoucherCode:   voucher.Code,
		ProductID:     product.ID,
		AgentID:       agent.ID,
		IntervalDays:  30,
		Status:        "active",
		AutoRenew:     true,
		NextRenewalAt: now.Add(-1 * time.Hour), // Expired
	}
	db.Create(&sub)

	t.Run("Renewal Success", func(t *testing.T) {
		fmt.Printf("Initial now: %v\n", now)
		var userPre domain.RadiusUser
		db.First(&userPre, "username = ?", voucher.Code)
		fmt.Printf("Initial user expiry: %v\n", userPre.ExpireTime)

		app.SchedSubscriptionRenewalTask()

		// Verify Wallet deduction (CostPrice = 8.0)
		var updatedWallet domain.AgentWallet
		db.First(&updatedWallet, agent.ID)
		assert.Equal(t, 92.0, updatedWallet.Balance)

		// Verify RadiusUser extension (1h + 30 days)
		var updatedUser domain.RadiusUser
		db.First(&updatedUser, "username = ?", voucher.Code)
		fmt.Printf("Updated user expiry: %v\n", updatedUser.ExpireTime)
		
		expectedUserExpire := userPre.ExpireTime.Add(30 * 24 * time.Hour)
		assert.WithinDuration(t, expectedUserExpire, updatedUser.ExpireTime, time.Second)

		// Verify Subscription updated
		var updatedSub domain.VoucherSubscription
		db.First(&updatedSub, sub.ID)
		assert.Equal(t, 1, updatedSub.RenewalCount)
		assert.True(t, updatedSub.NextRenewalAt.After(now))
		
		expectedSubExpire := sub.NextRenewalAt.Add(30 * 24 * time.Hour)
		assert.WithinDuration(t, expectedSubExpire, updatedSub.NextRenewalAt, time.Minute)
	})

	t.Run("Renewal Skip due to insufficient balance", func(t *testing.T) {
		// Set balance to 0
		db.Model(&domain.AgentWallet{}).Where("agent_id = ?", agent.ID).Update("balance", 0)
		
		// Reset sub for renewal
		db.Model(&sub).Updates(map[string]interface{}{
			"next_renewal_at": now.Add(-1 * time.Hour),
			"renewal_count":   0,
		})

		app.SchedSubscriptionRenewalTask()

		// Verify RenewalCount still 0
		var subCheck domain.VoucherSubscription
		db.First(&subCheck, sub.ID)
		assert.Equal(t, 0, subCheck.RenewalCount)
	})
}
