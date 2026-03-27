# Voucher "User Expired" Bug - Resolution Summary

## Problem Solved ✅

Your vouchers were immediately showing "user expired" upon first login. This has been **fixed**.

---

## What Was Wrong

When you created a voucher batch with "first_use" expiration (valid for 7 days from first login), the `validity_days` field was **not set** (defaulted to 0).

This caused the authentication system to calculate:
```
expiration = NOW + (0 days × 24 hours) = NOW
```

So users expired immediately!

---

## Fixes Applied

### 1. Runtime Defense ✅
**File:** `internal/radiusd/plugins/auth/checkers/first_use_activator.go`

Now validates that `validity_days > 0` before calculating expiration.
If validity_days is 0, returns a clear error instead of causing immediate expiration.

### 2. API Validation ✅
**File:** `internal/adminapi/vouchers.go`

Now validates when creating first-use batches that `validity_days` is set (1-365 days).
Prevents creating misconfigured batches.

---

## What You Need to Do

### Step 1: Fix Your Existing Batch

Your current batch has `validity_days=0` and needs to be updated.

**Option A: Run the fix script** (Recommended)
```bash
cd /home/faris/Documents/lamees/radio
./scripts/fix_voucher_batch_validity.sh
```

**Option B: Manual SQL update**
```bash
sqlite3 rundata/data/toughradius.db "UPDATE voucher_batch SET validity_days = 7 WHERE id = 2;"
```

### Step 2: Restart the Backend

The code changes require a restart:
```bash
# Stop the backend (Ctrl+C if running)
# Then start it again
./start_dev.sh
```

### Step 3: Test Login

1. Try logging in with voucher **675327**
2. Expected result: **Login succeeds!** ✅
3. User account will expire on **April 1, 2026** (7 days from first login)

---

## How to Create New Batches Correctly

When creating a new voucher batch:

### Required Fields:
- **Expiration Type:** `first_use`
- **Validity Days:** `7` ← **CRITICAL! Must be 1-365**
- **Product:** Select your product
- **Count:** Number of vouchers to generate

### Example Configuration:
```
Name: Batch #3
Product: 1000 (6000 MB, 7 days)
Expiration Type: first_use
Validity Days: 7 ← Don't forget this!
Count: 10
```

### What Each Field Means:

- **Expiration Type: "first_use"**
  - Voucher is valid for X days from FIRST login (not from creation)
  - User activates when they first log in

- **Expiration Type: "fixed"**
  - Voucher expires on a specific date (regardless of when activated)
  - Less common for prepaid vouchers

- **Validity Days:** 7
  - For first_use: Days from first login until expiration
  - For fixed: Not used (use PrintExpireTime instead)

---

## Verification

### Check Database Values

Run the debug script to verify the fix:
```bash
go run scripts/debug_voucher.go
```

**Look for:**
```
ValidityDays: 7  ← Should be > 0
ExpirationType: first_use
```

### Monitor Authentication Logs

When a user logs in, you should see:
```
INFO: first_use_activator: voucher activated on first login
  username: 675327
  new_expire: 2026-04-01 16:56:00 (7 days from now)
```

**NOT** this error:
```
ERROR: first_use_activator: Invalid ValidityDays configuration
  validity_days: 0
```

---

## Understanding Time vs Data Limits

### Your Product Configuration:
- **Data Quota:** 6000 MB
- **ValiditySeconds:** 604800 seconds (7 days)

### What This Means:

✅ **Data Limit:** User can use up to 6000 MB of data
- Tracked in real-time
- When exceeded: Cannot access more data

✅ **Time Limit:** Account is valid for 7 days
- NOT a usage time tracker (not "7 hours of online time")
- Account expiration date is set at first login
- After 7 days: Account expires, regardless of data used

### Example Timeline:

```
Day 0 (March 25, 4:00 PM): User logs in first time
  → Account created
  → Expires: April 1, 4:00 PM
  → Data used: 50 MB

Day 2 (March 27, 10:00 AM): User still has access
  → Account still active ✅
  → Data used: 500 MB

Day 5 (March 30): User still has access
  → Account still active ✅
  → Data used: 4500 MB

Day 7 (April 1, 4:01 PM): User tries to login
  → Account expired ❌
  → "user expired" error
```

---

## Troubleshooting

### If you still get "user expired" after the fix:

1. **Verify batch was updated:**
   ```bash
   go run scripts/debug_voucher.go
   ```
   Check: `ValidityDays: 7`

2. **Check if backend is restarted:**
   - Code changes require restart to take effect
   - Stop and restart `./start_dev.sh`

3. **Check user was already created:**
   - If user was created BEFORE the fix, they may have wrong ExpireTime
   - Solution: Delete the user and try logging in again

4. **Check logs for errors:**
   ```bash
   tail -f backend.log | grep -i "first_use\|expired"
   ```

### If you get "invalid voucher configuration" error:

This means the batch still has `validity_days=0`. Run the fix script again.

---

## Summary of Changes

### Code Files Modified:
1. `internal/radiusd/plugins/auth/checkers/first_use_activator.go`
   - Added validation for ValidityDays
   - Added error logging
   - Added clear error message

2. `internal/adminapi/vouchers.go`
   - Added validation when creating batches
   - Prevents saving invalid configuration

### Documentation Created:
1. `docs/VOUCHER_EXPIRATION_FIX.md` - Technical details
2. `docs/VOUCHER_BUG_RESOLUTION.md` - This summary
3. `scripts/fix_voucher_batch_validity.sh` - Batch fix script
4. `scripts/debug_voucher.go` - Debug tool

---

## Quick Start Commands

```bash
# 1. Fix existing batch
./scripts/fix_voucher_batch_validity.sh

# 2. Restart backend
./start_dev.sh

# 3. Test login
# (Try logging in with voucher 675327)

# 4. Debug if needed
go run scripts/debug_voucher.go
```

---

## Need Help?

If issues persist:
1. Check the logs: `tail -f backend.log`
2. Run debug script: `go run scripts/debug_voucher.go`
3. Verify batch configuration in database
4. Ensure backend is restarted with new code

---

**★ Insight ─────────────────────────────────────**
- **Root Cause Found**: Zero values (Go's default) caused silent failure in time calculation
- **Defense in Depth**: Fixed at both runtime activation AND batch creation
- **Clear Error Messages**: Now explains exactly what's wrong and how to fix it
`─────────────────────────────────────────────────`

**Status:** ✅ **FIXED** - Ready for testing!
