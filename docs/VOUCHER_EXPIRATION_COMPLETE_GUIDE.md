# Complete Voucher Expiration System Guide

## Overview

This guide explains the complete voucher expiration system after all fixes have been applied.

## Two Expiration Types

### 1. FIXED Expiration (From Creation)

**Use Case:** Event vouchers, promotional codes with fixed end dates

**Behavior:**
- Vouchers expire on a specific date (set when batch is created)
- All vouchers in batch expire on the same date
- User expiration = voucher expiration date
- Example: Event on Dec 31, 2027 → All vouchers expire on Dec 31, 2027

**Configuration:**
```
Expiration Type: fixed
Expire Time: 31/12/2027  (or leave blank for default 2999-12-31)
```

**Timeline Example:**
```
Day 0 (March 25, 2026):
  - Admin creates batch with fixed expiry: 31/12/2027
  - Vouchers created with ExpireTime = 2027-12-31

Day 100 (July 3, 2026):
  - User redeems voucher
  - User account created with ExpireTime = 2027-12-31

Day 500 (August 7, 2027):
  - User logs in → SUCCESS ✅ (not yet expired)

Day 650 (Dec 31, 2027):
  - User logs in → EXPIRED ❌
```

---

### 2. FIRST_USE Expiration (From Activation)

**Use Case:** Prepaid vouchers, time-limited access passes

**Behavior:**
- Vouchers can be activated anytime (default expiry: 2999-12-31)
- User expiration calculated at FIRST LOGIN
- Each user gets individual validity period
- Example: 7-day validity → Each user expires 7 days after THEIR first login

**Configuration:**
```
Expiration Type: first_use
Validity Days: 7  (number only, no text)
```

**Timeline Example:**
```
Day 0 (March 25, 2026):
  - Admin creates batch: 7 days validity
  - Vouchers created with ExpireTime = 2999-12-31 (can be activated anytime)

Day 1 (March 26, 2026):
  - Admin redeems vouchers
  - User accounts created with ExpireTime = 9999-12-31 (placeholder)

Day 2 (March 27, 2026):
  - User #1 logs in for first time
  - FirstUseActivator: Sets ExpireTime = April 3, 2026 (March 27 + 7 days)

Day 5 (March 30, 2026):
  - User #2 logs in for first time
  - FirstUseActivator: Sets ExpireTime = April 6, 2026 (March 30 + 7 days)

Day 9 (April 3, 2026):
  - User #1 logs in → EXPIRED ❌ (7 days elapsed)
  - User #2 logs in → SUCCESS ✅ (only 4 days elapsed)

Day 12 (April 6, 2026):
  - User #2 logs in → EXPIRED ❌ (7 days elapsed)
```

---

## Field Mappings

### Frontend → Backend

The frontend sends virtual fields that the backend transforms:

**Frontend Sends:**
```json
{
  "expiration_type": "first_use",
  "validity_value_virtual": 10,
  "validity_unit_virtual": "days"
}
```

**Backend Transforms To:**
```go
ValidityDays = 10
```

**Supported Units:**
- `days`: Direct conversion (10 days → 10 validity days)
- `hours`: Converted to days (240 hours → 10 validity days)
- `minutes`: Converted to days (14400 minutes → 10 validity days)

---

## Complete Flow

### FIXED Type Flow

```
1. ADMIN CREATES BATCH
   ├─ Expiration Type: fixed
   ├─ Expire Time: 31/12/2027
   └─ Vouchers created: ExpireTime = 2027-12-31

2. USER REDEEMS VOUCHER
   ├─ Voucher status: unused → active
   └─ User created: ExpireTime = 2027-12-31 (copied from voucher)

3. USER LOGS IN
   ├─ Authentication check: ExpireTime > NOW?
   └─ Result: SUCCESS (until 2027-12-31)
```

### FIRST_USE Type Flow

```
1. ADMIN CREATES BATCH
   ├─ Expiration Type: first_use
   ├─ Validity Days: 7
   └─ Vouchers created: ExpireTime = 2999-12-31

2. ADMIN REDEEMS VOUCHERS
   ├─ Voucher status: unused → active
   └─ User created: ExpireTime = 9999-12-31 (placeholder)

3. USER LOGS IN (FIRST TIME)
   ├─ FirstUseActivator detects: ExpireTime.Year() == 9999
   ├─ Calculates: NOW + 7 days
   ├─ Updates user: ExpireTime = NOW + 7 days
   └─ Logs: "voucher activated on first login"

4. USER LOGS IN (SUBSEQUENT TIMES)
   ├─ FirstUseActivator: ExpireTime.Year() < 9999 (skip)
   ├─ ExpireChecker: ExpireTime > NOW?
   └─ Result: SUCCESS (until 7 days elapsed)
```

---

## Database Schema

### voucher_batch Table

```sql
CREATE TABLE voucher_batch (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255),
    expiration_type VARCHAR(20),  -- "fixed" or "first_use"
    validity_days INTEGER,         -- Days for first_use type
    print_expire_time DATETIME,    -- When batch expires (or 2999-12-31)
    created_at DATETIME
);
```

### voucher Table

```sql
CREATE TABLE voucher (
    id INTEGER PRIMARY KEY,
    code VARCHAR(50),
    batch_id INTEGER,
    status VARCHAR(20),             -- unused, active, used, expired
    expire_time DATETIME,           -- Voucher expiry (2999-12-31 or user-specified)
    first_used_at DATETIME,        -- When user first logged in
    created_at DATETIME
);
```

### radius_user Table

```sql
CREATE TABLE radius_user (
    id INTEGER PRIMARY KEY,
    username VARCHAR(50),
    expire_time DATETIME,           -- User expiry
                                  -- FIXED: Same as voucher expiry
                                  -- FIRST_USE: First login + validity days
    created_at DATETIME
);
```

---

## Testing Scenarios

### Test 1: Fixed Expiration

**Setup:**
```bash
# Create batch
Expiration Type: fixed
Expire Time: 25/03/2027 (1 year from now)
Count: 5
```

**Verification:**
```sql
SELECT code, expire_time FROM voucher WHERE batch_id = X;
-- Expected: All vouchers have expire_time = 2027-03-25
```

**Expected Behavior:**
- ✅ Login works today (March 25, 2026)
- ✅ Login works in 11 months (February 25, 2027)
- ❌ Login fails on March 26, 2027 (expired)

---

### Test 2: First-Use Expiration

**Setup:**
```bash
# Create batch
Expiration Type: first_use
Validity Days: 7
Count: 5
```

**Verification:**
```sql
SELECT code, expire_time FROM voucher WHERE batch_id = X;
-- Expected: All vouchers have expire_time = 2999-12-31

SELECT username, expire_time FROM radius_user WHERE username LIKE 'CODE%';
-- Expected: Users have expire_time = 9999-12-31 (before first login)
```

**Expected Behavior:**
```
Day 0: User #1 logs in → User expires on Day 7
Day 2: User #2 logs in → User expires on Day 9
Day 5: User #3 logs in → User expires on Day 12
```

---

## Troubleshooting

### Issue: "Invalid first_use batch configuration"

**Cause:** ValidityDays is 0 or not set

**Solution:**
- Check that you entered a number (e.g., "7") without text
- Check logs for: "CreateVoucherBatch request received"
- Verify validity_value_virtual and validity_unit_virtual are sent

### Issue: "user expired" immediately after first login

**Cause:** FirstUseActivator not running or batch.ValidityDays is 0

**Solution:**
1. Check logs for: "first_use_activator: voucher activated on first login"
2. Verify batch.ValidityDays > 0 in database
3. Check that FirstUseActivator is registered in auth pipeline

### Issue: All users expire at the same time (first_use type)

**Cause:** RedeemVoucher is setting ExpireTime at redemption instead of 9999

**Solution:**
- Check that vouchers.go line 587 sets ExpireTime to 9999-12-31 for first_use
- Verify FirstUseActivator is updating ExpireTime on first login

### Issue: Vouchers show wrong expiry date

**Cause:** ExpireTime not set correctly at batch creation

**Solution:**
- Check vouchers.go lines 388-403 for voucher ExpireTime calculation
- Verify PrintExpireTime is parsed correctly (lines 345-378)
- Check that default is 2999-12-31 when not specified

---

## Log Examples

### Successful Batch Creation (FIRST_USE)

```
INFO CreateVoucherBatch request received
  expiration_type: first_use
  validity_days: 0
  validity_value_virtual: 7
  validity_unit_virtual: days

INFO Transformed virtual fields to ValidityDays
  validity_value_virtual: 7
  validity_unit_virtual: days
  calculated_validity_days: 7

INFO Parsed PrintExpireTime
  input: (empty)
  parsed: 2999-12-31 00:00:00
  format_used: (default)
```

### Successful First Login Activation

```
INFO first_use_activator: voucher activated on first login
  username: CODE12345
  activated_at: 2026-03-25T15:30:00
  new_expire: 2026-04-01T15:30:00
  validity_days: 7
```

### Error Cases

```
ERROR Invalid first_use batch configuration
  validity_days: 0
  expiration_type: first_use
  batch_name: Batch #5
  → FIX: Enter Validity Days as a number (1-365)
```

---

## Code Changes Summary

### Files Modified

1. **internal/adminapi/vouchers.go**
   - Lines 190-196: Added virtual field support
   - Lines 227-247: Added transformation logic
   - Lines 388-403: Fixed voucher ExpireTime at batch creation
   - Lines 415-442: Updated voucher creation loop
   - Lines 577-597: Fixed RedeemVoucher logic

2. **internal/radiusd/plugins/auth/checkers/first_use_activator.go**
   - Lines 78-91: Added ValidityDays validation
   - Lines 93-94: Calculate expiration from first login

---

## Quick Reference

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `expiration_type` | string | "fixed" or "first_use" | "fixed" |
| `validity_days` | int | Days from first login (first_use only) | 0 |
| `expire_time` | string | Fixed expiry date (fixed only) | 2999-12-31 |
| `validity_value_virtual` | int | Numeric value from frontend | - |
| `validity_unit_virtual` | string | Unit from frontend (days/hours/minutes) | "days" |

---

**Last Updated:** 2026-03-25
**Status:** ✅ All fixes applied and tested
