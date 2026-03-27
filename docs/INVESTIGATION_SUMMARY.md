# Voucher System Investigation Summary

## Quick Findings

**Can Scenario 2b work?** YES - but by accident, not design.

**The Two Independent Time Limits:**
1. `user.ExpireTime` = Validity window (50 hours from first login)
2. `user.TimeQuota` = Time quota (24 hours total from product)

**Critical Flaws Found:** 3

---

## Execution Flow Summary

### Authentication Checkers (in order)

```
Order 3:  VoucherAuthChecker    → Checks voucher activation deadline
Order 5:  FirstUseActivator     → Activates first-use vouchers on first login
Order 10: ExpireChecker         → Checks if user expired (validity window)
Order 16: TimeQuotaChecker      → Checks if time quota exceeded
```

### Scenario 2b Walkthrough

**Input:** Product=4.9GB/24hrs, Type=first_use, Validity=50hrs, Expiry=30/12/2028

1. **Voucher Created:**
   - `voucher.ExpireTime = 30/12/2028` (activation deadline)
   - `voucher.TimeQuota = 86400` (24 hours from product)

2. **User Activated (2025-01-15 10:00):**
   - `user.ExpireTime = 2025-01-17 12:00` (now + 50 hours)
   - `user.TimeQuota = 86400` (24 hours from product)

3. **Login Checks:**
   - Order 10: `2025-01-17 12:00 > now` → PASS
   - Order 16: `43200 < 86400` → PASS (after 12 hours usage)

4. **After 50 Hours (2025-01-17 12:01):**
   - Order 10: `2025-01-17 12:00 < now` → BLOCK ✅

5. **After 24 Hours Usage (whenever):**
   - Order 16: `86400 >= 86400` → BLOCK ✅

**Result:** WHICHEVER LIMIT COMES FIRST blocks access.

---

## The 3 Critical Flaws

### FLAW 1: VoucherAuthChecker Checks Wrong Field

**File:** `internal/radiusd/plugins/auth/checkers/voucher_auth.go:93-99`

**Problem:** Checks `voucher.ExpireTime` (activation deadline) instead of letting `ExpireChecker` handle service expiry.

**Impact:** Confusing code, but works because `ExpireChecker` runs later.

**Fix:** Only check activation deadline for unused vouchers.

---

### FLAW 2: ExpireChecker Returns Wrong Error

**File:** `internal/radiusd/plugins/auth/checkers/expire_checker.go:25-27`

**Problem:** Returns "user_not_exists" instead of "account_expired".

**Impact:** Misleading error messages for debugging.

**Fix:** Return proper "account_expired" error.

---

### FLAW 3: FirstUseActivator Has Dangerous Fallback

**File:** `internal/radiusd/plugins/auth/checkers/first_use_activator.go:95-100`

**Problem:** Defaults to 48 hours if `ValidityDays` is 0.

**Impact:** Masks configuration errors.

**Fix:** Fail fast with error instead of silent fallback.

---

## Why It Works Despite The Flaws

The dual-limit system works because:

1. **ExpireChecker** (Order 10) enforces the validity window
2. **TimeQuotaChecker** (Order 16) enforces the time quota
3. Both checkers run independently
4. WHICHEVER FAILS FIRST blocks access

This is actually **CORRECT BEHAVIOR** for a dual-limit voucher system!

---

## Recommended Fixes Priority

### HIGH PRIORITY
1. Fix VoucherAuthChecker to only check activation deadline
2. Fix ExpireChecker error message

### MEDIUM PRIORITY
3. Remove FirstUseActivator fallback

### LOW PRIORITY
4. Add debug logging for troubleshooting

---

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     BATCH CREATION                           │
├─────────────────────────────────────────────────────────────┤
│  VoucherBatch.ExpirationType = "first_use"                  │
│  VoucherBatch.ValidityDays = 50 (hours)                     │
│  VoucherBatch.PrintExpireTime = 30/12/2028                  │
│                                                             │
│  Voucher.ExpireTime = 30/12/2028 (activation deadline)      │
│  Voucher.TimeQuota = 86400 (from product)                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  USER ACTIVATION (First Login)               │
├─────────────────────────────────────────────────────────────┤
│  FirstUseActivator (Order 5):                               │
│    - Detects user.ExpireTime.Year() == 9999                 │
│    - Calculates: now + 50 hours                             │
│    - Updates user.ExpireTime = 2025-01-17 12:00             │
│                                                             │
│  RadiusUser created:                                        │
│    - ExpireTime = 2025-01-17 12:00 (validity window)        │
│    - TimeQuota = 86400 (usage limit)                        │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│              AUTHENTICATION (Each Login Attempt)             │
├─────────────────────────────────────────────────────────────┤
│  Order 3:  VoucherAuthChecker                               │
│            → Checks voucher status & batch                  │
│                                                             │
│  Order 5:  FirstUseActivator                                │
│            → Skips if already activated                     │
│                                                             │
│  Order 10: ExpireChecker                                    │
│            → user.ExpireTime < now ?                        │
│            → 2025-01-17 12:00 < now ?                       │
│                                                             │
│  Order 16: TimeQuotaChecker                                 │
│            → totalTime >= user.TimeQuota ?                  │
│            → acct_session_time >= 86400 ?                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│              ACCOUNTING (Each Session Update)                │
├─────────────────────────────────────────────────────────────┤
│  VoucherQuotaSyncHandler:                                   │
│    - Reads AcctSessionTime from RADIUS packet               │
│    - Updates voucher.TimeUsed += sessionTime                │
│    - Updates voucher.DataUsed += sessionBytes               │
│                                                             │
│  TimeQuotaChecker queries:                                  │
│    - SUM(acct_session_time) from accounting records         │
│    - Compares against user.TimeQuota                        │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Takeaways

1. **The system WORKS** for Scenario 2b, but code quality is poor
2. **Two independent limits** are correctly enforced by different checkers
3. **Naming confusion** between voucher-level and user-level expiration
4. **Fixes are recommended** for maintainability, not functionality
5. **Dual-limit behavior** is actually correct - first limit reached blocks

---

## Files Referenced

- `/home/faris/Documents/lamees/radio/internal/adminapi/vouchers.go`
- `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/voucher_auth.go`
- `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/first_use_activator.go`
- `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/expire_checker.go`
- `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/auth/checkers/time_quota_checker.go`
- `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/accounting/handlers/voucher_quota_sync.go`
- `/home/faris/Documents/lamees/radio/internal/radiusd/plugins/init.go`
- `/home/faris/Documents/lamees/radio/internal/domain/voucher.go`
- `/home/faris/Documents/lamees/radio/internal/domain/radius.go`
- `/home/faris/Documents/lamees/radio/internal/domain/product.go`

---

**Investigation Date:** 2026-03-27
**Investigator:** Claude (End-to-End Integration Investigator)
**Status:** COMPLETE
