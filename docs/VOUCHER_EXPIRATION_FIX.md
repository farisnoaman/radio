# Voucher Expiration Bug - Fix Summary

**Date:** 2026-03-25
**Issue:** Vouchers with "first_use" expiration immediately show "user expired" on first login
**Status:** ✅ FIXED

---

## Root Cause

When creating voucher batches with `expiration_type="first_use"`, if `validity_days` was not set or was 0, the `FirstUseActivator` would calculate:

```go
validityDuration = time.Duration(0) * 24 * time.Hour = 0
newExpire = now.Add(0) = NOW
```

This caused users to expire immediately upon first login.

## Bug Flow

```
1. Create batch (ValidityDays=0)
   ↓
2. Redeem voucher → User created with ExpireTime=9999-12-31
   ↓
3. User logs in → FirstUseActivator sets ExpireTime=NOW
   ↓
4. ExpireChecker checks: NOW < NOW? → TRUE
   ↓
5. Error: "user expired"
```

## Fixes Applied

### Fix #1: FirstUseActivator Defense ✅

**File:** `internal/radiusd/plugins/auth/checkers/first_use_activator.go`

**What:** Added validation to check if `ValidityDays <= 0` before calculating expiration

**Result:** Returns clear error message instead of causing immediate expiration

```go
if batch.ValidityDays <= 0 {
    return radiuserrors.NewAuthErrorWithStage(
        "invalid_voucher_configuration",
        "Voucher configuration error: Invalid validity period...",
        "first_use_activator",
    )
}
```

### Fix #2: CreateVoucherBatch Validation ✅

**File:** `internal/adminapi/vouchers.go`

**What:** Added validation when creating first_use batches to ensure ValidityDays is set

**Result:** Prevents creating misconfigured batches at the API level

```go
if batch.ExpirationType == "first_use" {
    if batch.ValidityDays <= 0 {
        return fail(c, http.StatusBadRequest, "INVALID_CONFIGURATION",
            "First-use expiration type requires ValidityDays to be set (1-365 days)...",
            nil)
    }
}
```

---

## Testing the Fix

### Test 1: Verify Old Batch Still Fails (Expected)

Your existing batch with `ValidityDays=0` will now give a clear error:

```
Error: Voucher configuration error: Invalid validity period.
Please contact administrator to ensure the voucher batch is configured
with a valid validity period (1-365 days).
```

### Test 2: Create New Batch (Correct)

When creating a new batch:

1. **Fill in the form:**
   - Expiration Type: `first_use`
   - **Validity Days:** `7` ← **CRITICAL: Must be set!**
   - Other fields as needed

2. **Expected:** Batch created successfully ✅

3. **Redeem voucher:** User created with ExpireTime=9999-12-31 ✅

4. **Login:**
   - FirstUseActivator sets ExpireTime=NOW+7days ✅
   - ExpireChecker passes ✅
   - Login successful ✅

### Test 3: Attempt to Create Batch Without ValidityDays

1. **Fill in the form:**
   - Expiration Type: `first_use`
   - Validity Days: (leave empty or 0)

2. **Expected:** API returns error:
   ```
   Status: 400 Bad Request
   Error: INVALID_CONFIGURATION
   Message: First-use expiration type requires ValidityDays to be set (1-365 days)
   ```

---

## How to Fix Your Existing Vouchers

### Option 1: Update the Batch (Recommended)

**Using SQL:**
```sql
UPDATE voucher_batch
SET validity_days = 7
WHERE expiration_type = 'first_use'
  AND validity_days = 0;
```

**Using API (if available):**
```bash
# Update batch to set validity_days
PATCH /api/v1/voucher-batches/{batch_id}
{
  "validity_days": 7
}
```

### Option 2: Create New Batch

1. Delete the old batch (or mark as deleted)
2. Create new batch with `validity_days=7`
3. Generate new vouchers
4. Test login

---

## Verification Steps

### 1. Check Database

```bash
cd /home/faris/Documents/lamees/radio
go run scripts/debug_voucher.go
```

**Expected output:**
```
ValidityDays: 7  ← Should be > 0
```

### 2. Check Logs

When logging in, you should see:

**BEFORE fix:**
```
ERROR: first_use_activator: Invalid ValidityDays configuration
  validity_days: 0
```

**AFTER fix (with ValidityDays=7):**
```
INFO: first_use_activator: voucher activated on first login
  username: 675327
  new_expire: 2026-04-01 16:56:00 (7 days from now)
```

### 3. Test Login Flow

1. **Create batch** with validity_days=7
2. **Generate voucher**
3. **Redeem voucher** → Creates user
4. **Login via Mikrotik** using voucher code as username/password
5. **Expected:** Login succeeds ✅

---

## Technical Details

### Authentication Pipeline Order

```
1. VoucherAuthChecker (Order: 3)
   - Validates batch is active
   - Validates voucher status
   - Checks data/time quotas

2. FirstUseActivator (Order: 5) ← OUR FIX
   - Activates first-use vouchers
   - Sets correct expiration time
   - NOW: Validates ValidityDays > 0

3. ExpireChecker (Order: 10)
   - Checks if user is expired
   - NOW: Receives correct expiration time
```

### Why Two Fixes?

**FirstUseActivator:** Defense at runtime - prevents silent failures
**CreateVoucherBatch:** Defense at creation - catches bugs early

This is **defense in depth** - multiple layers of validation.

---

## Files Changed

1. `internal/radiusd/plugins/auth/checkers/first_use_activator.go`
   - Added ValidityDays validation
   - Added error logging
   - Added clear error message

2. `internal/adminapi/vouchers.go`
   - Added batch creation validation
   - Prevents saving invalid configuration

---

## Data Quota vs Time Quota Clarification

### Product Configuration

- **DataQuota:** 6000 MB (data cap)
- **ValiditySeconds:** 604800 seconds (7 days of account validity)

### What This Means

- ✅ User can use up to 6000 MB of data
- ✅ User account is valid for 7 days from first login
- ❌ "7 hours" is NOT a usage time limit - it's the account validity duration

### Quota Enforcement

- **Data quota:** Tracked in `Voucher.DataUsed` ✅
- **Time quota:** NOT tracked (no TimeUsed counter) ⚠️

The system correctly enforces data quota but uses validity period for time limiting.

---

## Summary

✅ **Bug Fixed:** FirstUseActivator now validates ValidityDays
✅ **API Protection:** Cannot create batches without ValidityDays
✅ **Clear Errors:** Helpful error messages for debugging
✅ **Defense in Depth:** Multiple validation layers

**Next Steps:**
1. Update your existing batch to set `validity_days=7`
2. Create new batch with proper configuration
3. Test login flow
4. Monitor logs for successful activations

---

**★ Insight ─────────────────────────────────────**
- **Zero Values Matter:** Go's zero value (0 for int) can be catastrophic when used in calculations
- **Validate Early:** Catching bad data at creation is better than failing at runtime
- **Clear Errors:** Error messages should guide users to the exact problem and solution
`─────────────────────────────────────────────────`
