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

	// Get table schema
	fmt.Println("voucher_batch table schema:")
	fmt.Println("================================")

	// Get column information
	rows, err := db.Raw("PRAGMA table_info(voucher_batch)").Rows()
	if err != nil {
		fmt.Println("Error getting table info:", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-20s %-15s %-10s %-10s\n", "Column", "Type", "NotNull", "Default")
	fmt.Println("------------------------------------------------------------")

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var dfltValue interface{}
		var pk int

		rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)

		defaultStr := "NULL"
		if dfltValue != nil {
			defaultStr = fmt.Sprintf("%v", dfltValue)
		}

		notNullStr := "NO"
		if notNull == 0 {
			notNullStr = "YES"
		}

		pkStr := ""
		if pk == 1 {
			pkStr = " PK"
		}

		fmt.Printf("%-20s %-15s %-10s %-10s%s\n", name, dataType, notNullStr, defaultStr, pkStr)
	}

	// Check for any indexes
	fmt.Println("\nIndexes:")
	fmt.Println("================================")
	indexRows, _ := db.Raw("PRAGMA index_list(voucher_batch)").Rows()
	defer indexRows.Close()

	for indexRows.Next() {
		var seq int
		var name, origin string
		var partial int
		indexRows.Scan(&seq, &name, &origin, &partial)
		fmt.Printf("  - %s (%s)\n", name, origin)
	}
}
