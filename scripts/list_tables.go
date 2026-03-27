package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("rundata/data/toughradius.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	rows, _ := db.Raw("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").Rows()
	defer rows.Close()

	fmt.Println("Tables in database:")
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Println("  -", name)
	}
}
