package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	ID             int64  `gorm:"column:id"`
	Name           string `gorm:"column:name"`
	TimeQuota      int64  `gorm:"column:time_quota"`
	ValiditySeconds int64 `gorm:"column:validity_seconds"`
}

type Voucher struct {
	ID        int64  `gorm:"column:id"`
	Code      string `gorm:"column:code"`
	TimeQuota int64  `gorm:"column:time_quota"`
}

type RadiusUser struct {
	ID        int64  `gorm:"column:id"`
	Username  string `gorm:"column:username"`
	TimeQuota int64  `gorm:"column:time_quota"`
	ExpireTime string `gorm:"column:expire_time"`
}

func (Product) TableName() string { return "product" }
func (Voucher) TableName() string { return "voucher" }
func (RadiusUser) TableName() string { return "radius_user" }

func main() {
	db, err := gorm.Open(sqlite.Open("rundata/data/toughradius.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("=== RECENT PRODUCTS ===")
	var products []Product
	db.Order("id desc").Limit(5).Find(&products)
	for _, p := range products {
		fmt.Printf("ID: %d | Name: %s | TimeQuota: %d | ValiditySeconds: %d\n",
			p.ID, p.Name, p.TimeQuota, p.ValiditySeconds)
	}

	fmt.Println("\n=== RECENT VOUCHERS ===")
	var vouchers []Voucher
	db.Order("id desc").Limit(5).Find(&vouchers)
	for _, v := range vouchers {
		fmt.Printf("ID: %d | Code: %s | TimeQuota: %d\n",
			v.ID, v.Code, v.TimeQuota)
	}

	fmt.Println("\n=== RECENT USERS ===")
	var users []RadiusUser
	db.Order("id desc").Limit(5).Find(&users)
	for _, u := range users {
		fmt.Printf("ID: %d | Username: %s | TimeQuota: %d | ExpireTime: %s\n",
			u.ID, u.Username, u.TimeQuota, u.ExpireTime)
	}
}
