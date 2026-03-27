#!/bin/bash
################################################################################
# Fix Existing Voucher Batch - Set ValidityDays
#
# This script updates existing first_use voucher batches that have
# ValidityDays=0 to use the correct value (7 days).
################################################################################

DB_PATH="rundata/data/toughradius.db"
NEW_VALIDITY_DAYS=7

echo "========================================="
echo "Fix Voucher Batch ValidityDays"
echo "========================================="
echo ""
echo "This will update ALL first_use voucher batches with ValidityDays=0"
echo "to have ValidityDays=$NEW_VALIDITY_DAYS"
echo ""
echo "Database: $DB_PATH"
echo ""

read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

echo ""
echo "Checking for affected batches..."
echo ""

# Check which batches will be updated
export PATH=$PATH:/home/faris/go/go/bin
go run -c 'package main; import ("fmt"; "gorm.io/driver/sqlite"; "gorm.io/gorm"); func main() { db, _ := gorm.Open(sqlite.Open("'$DB_PATH'"), &gorm.Config{}); rows, _ := db.Raw("SELECT id, name, expiration_type, validity_days FROM voucher_batch WHERE expiration_type = '\''first_use'\'' AND validity_days = 0").Rows(); defer rows.Close(); fmt.Println("Affected batches:"); for rows.Next() { var id int; var name, expType string; var days int; rows.Scan(&id, &name, &expType, &days); fmt.Printf("  ID %d: %s (expiration_type=%s, validity_days=%d)\n", id, name, expType, days) } }' 2>&1

echo ""
read -p "Update these batches to validity_days=$NEW_VALIDITY_DAYS? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Update the batches
echo ""
echo "Updating batches..."

go run -c 'package main; import ("fmt"; "gorm.io/driver/sqlite"; "gorm.io/gorm"); func main() { db, _ := gorm.Open(sqlite.Open("'$DB_PATH'"), &gorm.Config{}); result := db.Exec("UPDATE voucher_batch SET validity_days = '$NEW_VALIDITY_DAYS' WHERE expiration_type = '\''first_use'\'' AND validity_days = 0"); if result.Error != nil { fmt.Println("Error:", result.Error) } else { fmt.Printf("Updated %d rows\n", result.RowsAffected) } }' 2>&1

echo ""
echo "Verifying update..."

go run -c 'package main; import ("fmt"; "gorm.io/driver/sqlite"; "gorm.io/gorm"); func main() { db, _ := gorm.Open(sqlite.Open("'$DB_PATH'"), &gorm.Config{}); var count int64; db.Model(&struct{TableName() string}{TableName: "voucher_batch"}{}).Where("expiration_type = '\''first_use'\'' AND validity_days = 0").Count(&count); if count > 0 { fmt.Printf("ERROR: %d batches still have validity_days=0\n", count) } else { fmt.Println("✓ All first_use batches now have validity_days > 0") } }' 2>&1

echo ""
echo "========================================="
echo "Fix Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Test login with voucher 675327"
echo "2. Monitor logs for successful activation"
echo "3. Verify user expires 7 days from first login"
echo ""
