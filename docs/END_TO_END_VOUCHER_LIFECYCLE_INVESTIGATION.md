# End-to-End Integration Investigation Report

## Summary
- **Can current code handle Scenario 2b?** PARTIAL
- **Logic Gaps Found:** 3 CRITICAL
- **Root Cause:** Confusion between voucher-level `ExpireTime` (activation deadline) and user-level `ExpireTime` (validity window end)

## Critical Finding

The system has **TWO INDEPENDENT TIME LIMITS** that are currently being **CONFUSED**:

1. **Voucher.ExpireTime** = Activation deadline (when voucher must be activated by)
2. **RadiusUser.ExpireTime** = Service expiry (when user's access actually ends)
3. **RadiusUser.TimeQuota** = Accumulated session time limit (hours from product)

**The code is mixing these up, causing Scenario 2b to FAIL.**

---

## Detailed Flow Analysis

### 1. Batch Creation Flow

**File:** `/home/faris/Documents/lamees/radio/internal/adminapi/vouchers.go`

**Function:** `CreateVoucherBatch` (lines 203-425)

**Scenario 2b Input:**
- Product: 4.9GB/24hrs
- Expiry Type: `first_use`
- Validity: 50 hours
- Voucher Expiry: 30/12/2028

**What Actually Happens:**

```go
// Line 308-321: Create batch record
batch := domain.VoucherBatch{
    ExpirationType: req.ExpirationType,  // "first_use"
    ValidityDays:   req.ValidityDays,    // 50 (HOURS!)
}

// Line 325-358: Set PrintExpireTime (activation deadline)
if req.ExpireTime != "" {
    batch.PrintExpireTime = &parsedTime  // 30/12/2028
} else {
    defaultExpiry := time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC)
    batch.PrintExpireTime = &defaultExpiry
}

// Line 368-378: Set voucher ExpireTime
expireTime := *batch.PrintExpireTime  // Copies activation deadline to voucher!

// Line 389-404: Create vouchers
voucher := domain.Voucher{
    ExpireTime:  expireTime,  // 30/12/2028 (activation deadline)
    TimeQuota:   product.ValiditySeconds,  // 86400 (24 hours from product)
    DataQuota:   product.DataQuota,        // 4.9GB
}
```

**✅ CORRECT:** Both `ValidityDays` (50 hours) AND `TimeQuota` (86400 seconds) are stored.

**❌ WRONG:** `voucher.ExpireTime` is set to the activation deadline (30/12/2028), not zero.

---

### 2. User Activation Flow (First Login)

**File:** `/home/faris/Documents/lamees/radio/internal/adminapi/vouchers.go`

**Function:** `RedeemVoucher` (lines 481-648)

**What Happens When User Logs In:**

```go
// Line 554-573: Calculate user expiration
if batch.ExpirationType == "first_use" {
    if batch.ValidityDays > 0 {
        windowSeconds := batch.ValidityDays * 3600  // 50 * 3600 = 180000 seconds
        expireTime = now.Add(time.Duration(windowSeconds) * time.Second)
        // expireTime = 2025-01-15 10:00 + 50 hours = 2025-01-17 12:00
    } else {
        expireTime = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
    }
}

// Line 595-611: Create RadiusUser
user := domain.RadiusUser{
    ExpireTime:  expireTime,  // 2025-01-17 12:00 ✅ CORRECT
    TimeQuota:   product.ValiditySeconds,  // 86400 seconds ✅ CORRECT
    DataQuota:   voucher.DataQuota,        // 4.9GB ✅ CORRECT
}
```

**✅ CORRECT:** User is created with BOTH:
- `ExpireTime = 2025-01-17 12:00` (50 hours from first login)
- `TimeQuota = 86400` (24 hours from product)

---

### 3. Authentication Validation Flow

**Execution Order** (from `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/init.go`):

```
Order 3:  VoucherAuthChecker
Order 5:  FirstUseActivator
Order 10: ExpireChecker
Order 15: QuotaChecker
Order 16: TimeQuotaChecker
```

#### 3.1 VoucherAuthChecker (Order 3)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/voucher_auth.go`

```go
// Line 93-99: Check voucher expiration
if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
    return error("voucher_expired")
}

// Line 109-115: Check voucher time quota
if voucher.TimeQuota > 0 && voucher.TimeUsed >= voucher.TimeQuota {
    return error("voucher_time_quota_exceeded")
}
```

**❌ CRITICAL BUG:** This checks `voucher.ExpireTime` which is the **activation deadline** (30/12/2028), NOT the service expiry!

After activation, this check becomes meaningless. The voucher should be checked against `user.ExpireTime`, not `voucher.ExpireTime`.

#### 3.2 FirstUseActivator (Order 5)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/first_use_activator.go`

```go
// Line 46-50: Detect pending activation
if user.ExpireTime.Year() < 9999 {
    return nil  // Already activated
}

// Line 83-104: Calculate new expiration
actualValidityDays := batch.ValidityDays  // 50 hours
if actualValidityDays <= 0 {
    actualValidityDays = 48  // Default fallback
}
validityDuration := time.Duration(actualValidityDays) * time.Hour
newExpire := now.Add(validityDuration)  // now + 50 hours

// Line 107: Update user
userRepo.UpdateField(ctx, user.Username, "expire_time", newExpire)
```

**✅ CORRECT:** Activates first-use vouchers and sets `user.ExpireTime = now + validity_window`

**⚠️ WARNING:** Has fallback to 48 hours if `ValidityDays` is 0, which shouldn't happen with proper input.

#### 3.3 ExpireChecker (Order 10)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/expire_checker.go`

```go
// Line 22-29: Check user expiration
if user.ExpireTime.Before(time.Now()) {
    return error("user_not_exists")  // Wrong error message!
}
```

**✅ CORRECT LOGIC:** Checks if `user.ExpireTime < now`

**❌ WRONG ERROR:** Returns "user_not_exists" instead of "account_expired"

#### 3.4 TimeQuotaChecker (Order 16)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/time_quota_checker.go`

```go
// Line 31-50: Check accumulated session time
totalTime, err := c.accountingRepo.GetTotalSessionTime(ctx, user.Username)
if totalTime >= user.TimeQuota {
    return error("time_quota_exceeded")
}
```

**✅ CORRECT:** Checks if accumulated session time exceeds `user.TimeQuota`

**❌ WRONG DATA SOURCE:** Uses `user.TimeQuota` (86400 seconds from product), which tracks TOTAL time, not the validity window.

---

### 4. Session Accounting Flow

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/accounting/handlers/voucher_quota_sync.go`

**Function:** `VoucherQuotaSyncHandler.Handle` (lines 29-80)

**What Happens During Each Session Update/Stop:**

```go
// Line 58: Get session time from RADIUS packet
acctSessionTime := int64(rfc2866.AcctSessionTime_Get(r.Packet))

// Line 60-63: Update voucher usage
updates := map[string]interface{}{
    "data_used": gorm.Expr("data_used + ?", dataUsedMB),
    "time_used": gorm.Expr("time_used + ?", acctSessionTime),
}

// Line 65-67: Apply to VOUCHER, not RadiusUser!
db.Model(&domain.Voucher{}).
    Where("code = ?", user.VoucherCode).
    Updates(updates)
```

**✅ CORRECT:** Accumulates session time in `voucher.TimeUsed`

**❌ CRITICAL GAP:** `RadiusUser.TimeQuota` is NEVER updated during accounting!

---

## Critical Logic Flaws

### FLAW 1: Voucher vs User TimeQuota Confusion

**Location:** Multiple files

**Problem:**
- `Voucher.TimeQuota` = Product's validity seconds (86400 for 24-hour product)
- `RadiusUser.TimeQuota` = Same value, copied during activation
- `TimeQuotaChecker` checks `user.TimeQuota` against accounting records
- `VoucherQuotaSyncHandler` updates `voucher.TimeUsed`, NOT `user.TimeQuota`

**Impact:**
- TimeQuotaChecker queries accounting tables for total session time
- VoucherQuotaSyncHandler updates voucher table
- These are TWO DIFFERENT data sources that can get out of sync!

**Scenario 2b Walkthrough:**

1. User created with `user.TimeQuota = 86400` (24 hours)
2. User uses 12 hours across multiple sessions
3. Accounting shows: `acct_session_time = 43200` seconds (12 hours)
4. Voucher shows: `time_used = 43200` seconds (12 hours)
5. TimeQuotaChecker checks: `43200 >= 86400` → FALSE → Allows login ✅
6. On `2025-01-17 12:00`: ExpireChecker checks `user.ExpireTime < now` → TRUE → Blocks login ✅

**This part actually WORKS!** But only because TimeQuotaChecker and VoucherQuotaSyncHandler are inconsistent.

---

### FLAW 2: Wrong Expiration Check in VoucherAuthChecker

**Location:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/voucher_auth.go:93-99`

**Problem:**
```go
if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
    return error("voucher_expired")
}
```

This checks the **activation deadline** (30/12/2028), not the service expiry!

**Impact:**
- After activation, `voucher.ExpireTime` is meaningless
- Should check `user.ExpireTime` instead
- But this checker runs BEFORE FirstUseActivator, so it doesn't see the activated expiry

**Scenario 2b Impact:**
- User activates on `2025-01-15 10:00`
- `user.ExpireTime = 2025-01-17 12:00`
- `voucher.ExpireTime = 30/12/2028`
- On `2025-01-18`, VoucherAuthChecker sees `30/12/2028 > now` → PASSES ✅
- But ExpireChecker sees `2025-01-17 12:00 < now` → BLOCKS ✅

**Result:** It works, but for the WRONG REASON!

---

### FLAW 3: TimeQuotaChecker Checks Wrong Field

**Location:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/time_quota_checker.go:46-48`

**Problem:**
```go
if totalTime >= user.TimeQuota {
    return error("time_quota_exceeded")
}
```

This checks `user.TimeQuota` (product validity: 86400 seconds), which is meant to track TOTAL AVAILABLE TIME, not the validity window.

**Expected Behavior for Scenario 2b:**
- User has `TimeQuota = 86400` seconds (24 hours total)
- User has `ExpireTime = 2025-01-17 12:00` (50-hour window)
- After using 12 hours, user should still have access until `2025-01-17 12:00`
- After using 24 hours, user should be blocked even if within the 50-hour window

**Actual Behavior:**
- TimeQuotaChecker blocks after 24 hours of usage
- ExpireChecker blocks after 50 hours from activation
- WHICHEVER COMES FIRST wins

**This is actually CORRECT behavior for a dual-limit system!**

---

## Scenario 2b Test Results

### Step-by-Step Walkthrough

**Input:**
- Product: 4.9GB/24hrs
- Expiry Type: `first_use`
- Validity: 50 hours
- Voucher Expiry: 30/12/2028

**1. Batch Creation (2025-01-01)**
```
VoucherBatch {
    ExpirationType: "first_use",
    ValidityDays: 50,  // hours
    PrintExpireTime: 30/12/2028,
}

Voucher {
    ExpireTime: 30/12/2028,  // Activation deadline
    TimeQuota: 86400,        // 24 hours from product
    DataQuota: 5060,         // 4.9GB
}
```
✅ Both time limits stored correctly

**2. User Activation (2025-01-15 10:00)**
```
RadiusUser {
    ExpireTime: 2025-01-17 12:00,  // now + 50 hours
    TimeQuota: 86400,              // 24 hours from product
    DataQuota: 5060,               // 4.9GB
}
```
✅ User created with correct expiration

**3. First Login Attempt (2025-01-15 10:00)**

Checker execution:
- Order 3: VoucherAuthChecker
  - `voucher.ExpireTime = 30/12/2028 > now` → PASS ✅
  - `voucher.TimeQuota = 86400, voucher.TimeUsed = 0` → PASS ✅
- Order 5: FirstUseActivator
  - `user.ExpireTime.Year() = 2025 < 9999` → SKIP (already activated) ✅
- Order 10: ExpireChecker
  - `user.ExpireTime = 2025-01-17 12:00 > now` → PASS ✅
- Order 16: TimeQuotaChecker
  - `totalTime = 0 < user.TimeQuota = 86400` → PASS ✅

Result: **LOGIN ALLOWED** ✅

**4. Session Usage (12 hours across multiple sessions)**

Accounting updates:
- Session 1: 3 hours → `acct_session_time = 10800`
- Session 2: 5 hours → `acct_session_time = 18000`
- Session 3: 4 hours → `acct_session_time = 14400`
- Total: `acct_session_time = 43200` seconds (12 hours)

Voucher updates:
- `voucher.TimeUsed = 43200` seconds (12 hours)

✅ Time quota tracked correctly

**5. Login Before Expiry (2025-01-17 11:00)**

Checker execution:
- Order 10: ExpireChecker
  - `user.ExpireTime = 2025-01-17 12:00 > now` → PASS ✅
- Order 16: TimeQuotaChecker
  - `totalTime = 43200 < user.TimeQuota = 86400` → PASS ✅

Result: **LOGIN ALLOWED** ✅

**6. Login After Expiry (2025-01-17 12:01)**

Checker execution:
- Order 10: ExpireChecker
  - `user.ExpireTime = 2025-01-17 12:00 < now` → **FAIL** ❌

Result: **LOGIN BLOCKED** ✅

**7. Login After 24 Hours of Usage (hypothetical, 2025-01-16 10:00)**

If user used 24 hours before the 50-hour window:
- `totalTime = 86400 >= user.TimeQuota = 86400` → **FAIL** ❌

Result: **LOGIN BLOCKED** ✅

---

## Verdict

### Scenario 2b Works, But By Accident

**Why It Works:**
1. ExpireChecker (Order 10) enforces the 50-hour window
2. TimeQuotaChecker (Order 16) enforces the 24-hour quota
3. WHICHEVER LIMIT IS REACHED FIRST blocks access

**Why It's Fragile:**
1. VoucherAuthChecker checks the wrong expiration field
2. TimeQuotaChecker and VoucherQuotaSyncHandler operate on different tables
3. Error messages are misleading ("user_not_exists" instead of "account_expired")

**The Two Time Limits ARE Independent:**
- `user.ExpireTime` = Hard deadline (50-hour window)
- `user.TimeQuota` = Usage limit (24 hours total)

This is actually the CORRECT behavior for a dual-limit voucher system!

---

## Recommendations

### 1. Fix VoucherAuthChecker (CRITICAL)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/voucher_auth.go`

**Change Line 93-99:**
```go
// BEFORE:
if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
    return error("voucher_expired")
}

// AFTER:
// Don't check voucher.ExpireTime here - that's the activation deadline
// Let ExpireChecker handle the actual service expiry via user.ExpireTime
// Only check if voucher hasn't been activated yet
if voucher.Status == "unused" && !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
    return radiuserrors.NewAuthErrorWithStage(
        "voucher_activation_expired",
        "Voucher activation deadline has passed. Please contact administrator.",
        "voucher_auth",
    )
}
```

### 2. Fix ExpireChecker Error Message (HIGH)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/expire_checker.go`

**Change Line 25-27:**
```go
// BEFORE:
if user.ExpireTime.Before(time.Now()) {
    return errors.NewUserNotExistsError()
}

// AFTER:
if user.ExpireTime.Before(time.Now()) {
    return errors.NewAccountExpiredError()
}
```

### 3. Remove Fallback in FirstUseActivator (MEDIUM)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/first_use_activator.go`

**Change Line 95-100:**
```go
// BEFORE:
actualValidityDays := batch.ValidityDays
if actualValidityDays <= 0 {
    zap.L().Warn("first_use_activator: ValidityDays is 0, using default 48 hours")
    actualValidityDays = 48 // Default to 48 hours
}

// AFTER:
actualValidityDays := batch.ValidityDays
if actualValidityDays <= 0 {
    zap.L().Error("first_use_activator: ValidityDays is 0, cannot activate voucher",
        zap.String("username", user.Username),
        zap.Int64("batch_id", batch.ID))
    return errors.NewAuthErrorWithStage(
        "voucher_configuration_error",
        "Voucher validity window not configured. Please contact administrator.",
        "first_use_activator",
    )
}
```

### 4. Add Logging for Debugging (LOW)

**File:** `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/expire_checker.go`

**Add detailed logging:**
```go
func (c *ExpireChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
    user := authCtx.User

    zap.L().Debug("expire_checker: checking user expiration",
        zap.String("username", user.Username),
        zap.Time("expire_time", user.ExpireTime),
        zap.Time("now", time.Now()),
        zap.Bool("is_expired", user.ExpireTime.Before(time.Now())))

    if user.ExpireTime.Before(time.Now()) {
        zap.L().Info("expire_checker: user account expired",
            zap.String("username", user.Username),
            zap.Time("expire_time", user.ExpireTime))
        return errors.NewAccountExpiredError()
    }

    return nil
}
```

---

## Summary

The current code **DOES** handle Scenario 2b correctly, but due to a mix of correct logic and accidental bugs. The two independent time limits work as expected:

1. **Validity Window (50 hours)**: Enforced by `ExpireChecker` via `user.ExpireTime`
2. **Time Quota (24 hours)**: Enforced by `TimeQuotaChecker` via `user.TimeQuota`

However, the code quality is poor due to:
- Confusing variable names (`ExpireTime` used for two different purposes)
- Wrong error messages
- Inconsistent data sources (voucher vs user tables)
- Dangerous fallbacks in first-use activation

The recommended fixes will make the code more maintainable and easier to understand, while preserving the correct dual-limit behavior.
