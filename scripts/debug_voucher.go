package main

import (
	"fmt"
	"time"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Voucher struct {
	ID          int64     `gorm:"column:id"`
	Code        string    `gorm:"column:code"`
	Status      string    `gorm:"column:status"`
	ExpireTime  time.Time `gorm:"column:expire_time"`
	DataQuota   int64     `gorm:"column:data_quota"`
	TimeQuota   int64     `gorm:"column:time_quota"`
	ActivatedAt time.Time `gorm:"column:activated_at"`
	FirstUsedAt time.Time `gorm:"column:first_used_at"`
	BatchID     int64     `gorm:"column:batch_id"`
}

func (Voucher) TableName() string {
	return "voucher"
}

type VoucherBatch struct {
	ID             int64      `gorm:"column:id"`
	Name           string     `gorm:"column:name"`
	ProductID      int64      `gorm:"column:product_id"`
	ExpirationType string     `gorm:"column:expiration_type"`
	ValidityDays   int        `gorm:"column:validity_days"`
	PrintExpireTime *time.Time `gorm:"column:expire_time"`
	ActivatedAt    *time.Time `gorm:"column:activated_at"`
}

func (VoucherBatch) TableName() string {
	return "voucher_batch"
}

type Product struct {
	ID              int64  `gorm:"column:id"`
	Name            string `gorm:"column:name"`
	DataQuota       int64  `gorm:"column:data_quota"`
	ValiditySeconds int64  `gorm:"column:validity_seconds"`
}

func (Product) TableName() string {
	return "product"
}

type RadiusUser struct {
	ID         int64     `gorm:"column:id"`
	Username   string    `gorm:"column:username"`
	Status     string    `gorm:"column:status"`
	ExpireTime time.Time `gorm:"column:expire_time"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (RadiusUser) TableName() string {
	return "radius_user"
}

func main() {
	db, err := gorm.Open(sqlite.Open("rundata/data/toughradius.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect: " + err.Error())
	}

	fmt.Println("=========================================")
	fmt.Println("Debugging Voucher: 675327")
	fmt.Println("=========================================")
	fmt.Println("")
	fmt.Println("Current time:", time.Now().Format("2006-01-02 15:04:05 PM MST"))
	fmt.Println("")

	// 1. Voucher info
	fmt.Println("1. VOUCHER INFO")
	fmt.Println("---------------")
	var voucher Voucher
	result := db.Table("voucher").Where("code = ?", "675327").First(&voucher)
	if result.Error != nil {
		fmt.Println("Error:", result.Error)
	} else {
		fmt.Printf("ID: %d\n", voucher.ID)
		fmt.Printf("Code: %s\n", voucher.Code)
		fmt.Printf("Status: %s\n", voucher.Status)
		fmt.Printf("ExpireTime: %s\n", formatTime(voucher.ExpireTime))
		fmt.Printf("ExpireTime.IsZero: %v\n", voucher.ExpireTime.IsZero())
		fmt.Printf("DataQuota: %d MB\n", voucher.DataQuota)
		fmt.Printf("TimeQuota: %d seconds (%.2f hours)\n", voucher.TimeQuota, float64(voucher.TimeQuota)/3600)
		fmt.Printf("ActivatedAt: %s\n", formatTime(voucher.ActivatedAt))
		fmt.Printf("FirstUsedAt: %s\n", formatTime(voucher.FirstUsedAt))
		fmt.Printf("BatchID: %d\n", voucher.BatchID)
	}
	fmt.Println("")

	// 2. Batch info
	fmt.Println("2. VOUCHER BATCH INFO")
	fmt.Println("--------------------")
	var batch VoucherBatch
	result = db.Table("voucher_batch").Where("id = ?", voucher.BatchID).First(&batch)
	if result.Error != nil {
		fmt.Println("Error:", result.Error)
	} else {
		fmt.Printf("ID: %d\n", batch.ID)
		fmt.Printf("Name: %s\n", batch.Name)
		fmt.Printf("ProductID: %d\n", batch.ProductID)
		fmt.Printf("ExpirationType: %s\n", batch.ExpirationType)
		fmt.Printf("ValidityDays: %d\n", batch.ValidityDays)
		if batch.PrintExpireTime != nil {
			fmt.Printf("PrintExpireTime: %s\n", batch.PrintExpireTime.Format("2006-01-02 15:04:05 PM MST"))
		} else {
			fmt.Printf("PrintExpireTime: (null)\n")
		}
		if batch.ActivatedAt != nil {
			fmt.Printf("ActivatedAt: %s\n", batch.ActivatedAt.Format("2006-01-02 15:04:05 PM MST"))
		} else {
			fmt.Printf("ActivatedAt: (null)\n")
		}
	}
	fmt.Println("")

	// 3. Product info
	fmt.Println("3. PRODUCT INFO")
	fmt.Println("--------------")
	var product Product
	result = db.Table("product").Where("id = ?", batch.ProductID).First(&product)
	if result.Error != nil {
		fmt.Println("Error:", result.Error)
	} else {
		fmt.Printf("ID: %d\n", product.ID)
		fmt.Printf("Name: %s\n", product.Name)
		fmt.Printf("DataQuota: %d MB\n", product.DataQuota)
		fmt.Printf("ValiditySeconds: %d seconds (%.2f days)\n", product.ValiditySeconds, float64(product.ValiditySeconds)/86400)
	}
	fmt.Println("")

	// 4. Radius user (if created)
	fmt.Println("4. RADIUS USER (if created)")
	fmt.Println("-------------------------")
	var user RadiusUser
	result = db.Table("radius_user").Where("username = ?", "675327").First(&user)
	if result.Error != nil {
		fmt.Println("✓ No user created yet (expected for first login)")
	} else {
		fmt.Printf("ID: %d\n", user.ID)
		fmt.Printf("Username: %s\n", user.Username)
		fmt.Printf("Status: %s\n", user.Status)
		fmt.Printf("ExpireTime: %s\n", formatTime(user.ExpireTime))
		fmt.Printf("CreatedAt: %s\n", formatTime(user.CreatedAt))

		now := time.Now()
		if user.ExpireTime.Before(now) {
			fmt.Printf("✗ BUG: User ExpireTime is in the PAST!\n")
			fmt.Printf("  User ExpireTime: %s\n", formatTime(user.ExpireTime))
			fmt.Printf("  Current time: %s\n", now.Format("2006-01-02 15:04:05 PM MST"))
			fmt.Printf("  Time difference: %v\n", now.Sub(user.ExpireTime))
		} else {
			fmt.Printf("✓ User ExpireTime is in the future\n")
		}
	}
	fmt.Println("")

	// 5. Analysis
	fmt.Println("5. ANALYSIS")
	fmt.Println("---------")
	now := time.Now()

	if voucher.Status == "unused" {
		fmt.Println("✓ Voucher is unused (correct)")
	} else if voucher.Status == "active" {
		fmt.Printf("⚠ Voucher status is: %s (already used)\n", voucher.Status)
	} else {
		fmt.Printf("✗ Voucher status is: %s (unexpected)\n", voucher.Status)
	}

	if batch.ExpirationType == "first_use" {
		fmt.Println("✓ Expiration type is 'first_use' (correct)")
		if batch.ValidityDays > 0 {
			expectedExpire := now.AddDate(0, 0, batch.ValidityDays)
			fmt.Printf("  Expected user ExpireTime after login: %s\n", expectedExpire.Format("2006-01-02 15:04:05 PM MST"))
		}
	} else {
		fmt.Printf("⚠ Expiration type is: %s (expected 'first_use' for 7 days from first login)\n", batch.ExpirationType)
	}

	fmt.Println("")
	fmt.Println("=========================================")
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "0001-01-01 00:00:00 +0000 UTC (ZERO TIME)"
	}
	return t.Format("2006-01-02 15:04:05 PM MST")
}
