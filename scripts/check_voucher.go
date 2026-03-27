// how to run it
// export PATH=/home/faris/go/go/bin:$PATH && cd /home/faris/Documents/lamees/radio && go run scripts/check_voucher.go
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RadiusUser struct {
	ID        int64  `gorm:"primaryKey;column:id"`
	Username  string `gorm:"column:username"`
	TimeQuota int64  `gorm:"column:time_quota"`
}

func (RadiusUser) TableName() string {
	return "radius_user"
}

type RadiusAccounting struct {
	ID              int64  `gorm:"primaryKey;column:id"`
	Username        string `gorm:"column:username"`
	AcctSessionId   string `gorm:"column:acct_session_id"`
	AcctSessionTime int    `gorm:"column:acct_session_time"`
}

func (RadiusAccounting) TableName() string {
	return "radius_accounting"
}

func main() {
	db, err := gorm.Open(sqlite.Open("rundata/data/toughradius.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	voucher := "5748524737"

	var user RadiusUser
	if err := db.Where("username = ?", voucher).First(&user).Error; err != nil {
		log.Fatalf("failed to find voucher %s: %v", voucher, err)
	}

	fmt.Printf("Voucher: %s\n", user.Username)
	fmt.Printf("Time Quota: %d seconds (%d minutes)\n", user.TimeQuota, user.TimeQuota/60)

	var sessions []RadiusAccounting
	if err := db.Where("username = ?", voucher).Find(&sessions).Error; err != nil {
		log.Fatalf("failed to find accounting records: %v", err)
	}

	var totalTime int
	fmt.Println("\nPast Sessions:")
	for _, s := range sessions {
		fmt.Printf("- Session %s: %d seconds\n", s.AcctSessionId, s.AcctSessionTime)
		totalTime += s.AcctSessionTime
	}

	fmt.Printf("\nTotal Used Time: %d seconds (%d minutes)\n", totalTime, totalTime/60)
	
	remaining := user.TimeQuota - int64(totalTime)
	fmt.Printf("Remaining Time: %d seconds (%d minutes)\n", remaining, remaining/60)
}
