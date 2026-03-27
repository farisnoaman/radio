#!/bin/bash
################################################################################
# Debug Voucher 675327 - Check database values
################################################################################

DB_PATH="/home/faris/Documents/lamees/radio/data/toughradius.db"

echo "========================================="
echo "Debugging Voucher: 675327"
echo "========================================="
echo ""

# Check if using PostgreSQL or SQLite
if command -v psql &> /dev/null; then
    echo "Using PostgreSQL..."
    DB_TYPE="postgres"
else
    echo "Using SQLite..."
    DB_TYPE="sqlite"
fi

echo ""
echo "1. VOUCHER INFO"
echo "---------------"
if [ "$DB_TYPE" = "sqlite" ]; then
    sqlite3 "$DB_PATH" "SELECT id, code, status, expire_time, data_quota, time_quota, activated_at, first_used_at FROM voucher WHERE code = '675327';" 2>/dev/null || echo "Query failed or database not found"
else
    echo "Please run this SQL query:"
    echo "SELECT id, code, status, expire_time, data_quota, time_quota, activated_at, first_used_at FROM voucher WHERE code = '675327';"
fi

echo ""
echo "2. VOUCHER BATCH INFO"
echo "--------------------"
if [ "$DB_TYPE" = "sqlite" ]; then
    sqlite3 "$DB_PATH" "SELECT id, name, product_id, expiration_type, validity_days, expire_time, activated_at FROM voucher_batch WHERE id IN (SELECT batch_id FROM voucher WHERE code = '675327');" 2>/dev/null
else
    echo "Please run this SQL query:"
    echo "SELECT id, name, product_id, expiration_type, validity_days, expire_time, activated_at FROM voucher_batch WHERE id IN (SELECT batch_id FROM voucher WHERE code = '675327');"
fi

echo ""
echo "3. PRODUCT INFO"
echo "--------------"
if [ "$DB_TYPE" = "sqlite" ]; then
    sqlite3 "$DB_PATH" "SELECT id, name, data_quota, validity_seconds FROM product WHERE id IN (SELECT product_id FROM voucher_batch WHERE id IN (SELECT batch_id FROM voucher WHERE code = '675327'));" 2>/dev/null
else
    echo "Please run this SQL query:"
    echo "SELECT id, name, data_quota, validity_seconds FROM product WHERE id IN (SELECT product_id FROM voucher_batch WHERE id IN (SELECT batch_id FROM voucher WHERE code = '675327'));"
fi

echo ""
echo "4. RADIUS USER (if created)"
echo "-------------------------"
if [ "$DB_TYPE" = "sqlite" ]; then
    sqlite3 "$DB_PATH" "SELECT id, username, status, expire_time, created_at FROM radius_user WHERE username = '675327';" 2>/dev/null
else
    echo "Please run this SQL query:"
    echo "SELECT id, username, status, expire_time, created_at FROM radius_user WHERE username = '675327';"
fi

echo ""
echo "5. CURRENT TIME"
echo "--------------"
date
echo ""

echo "========================================="
echo "What to check:"
echo "========================================="
echo "1. voucher.expire_time: Should be zero for first_use"
echo "2. batch.expiration_type: Should be 'first_use'"
echo "3. batch.validity_days: Should be 7 (for 7 days)"
echo "4. product.validity_seconds: Should be 604800 (7 days in seconds)"
echo "5. user.expire_time: Should be NOW + 7 days, not in the past!"
echo ""
