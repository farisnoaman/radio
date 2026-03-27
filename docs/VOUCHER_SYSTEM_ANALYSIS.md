# Voucher System Implementation Analysis

## User Requirements (Complete Scenario)

### Product Definition
```
Silver Product:
- Data Quota: 300MB
- Time Limit: 5 hours
```

### Batch Configuration
```
Batch:
- Expiration Type: first_use (from first login)
- Validity Days: X days (e.g., 10 days)
- This is the TIME WINDOW to use the product limits
```

### Expected Behavior

1. **During Validity Window (Day 0 to Day X):**
   - User can login and use the service
   - Data quota (300MB) is tracked and enforced
   - Time limit (5 hours) is tracked and enforced
   - If quota exhausted: "quota/time finished"
   - If time limit exhausted: "time finished"

2. **After Validity Window (Day X+):**
   - Voucher is considered expired
   - User cannot login anymore
   - RADIUS should say: "card not found" or "user not exists"

---

## Current Implementation Analysis

### ✅ WHAT WORKS CORRECTLY

#### 1. Product Limits Are Defined
**Location:** `internal/domain/product.go:8-23`

```go
type Product struct {
    DataQuota       int64     // Data quota in MB
    ValiditySeconds int64     // Time limit in seconds
}
```

✅ Product can define 300MB quota and 5-hour time limit

---

#### 2. User Creation With Product Limits
**Location:** `internal/adminapi/vouchers.go:615-628`

```go
user := domain.RadiusUser{
    DataQuota: voucher.DataQuota,  // Inherited from product (300MB)
    // NOTE: No TimeQuota field in RadiusUser!
}
```

✅ User gets DataQuota from product
❌ User does NOT get TimeQuota (field doesn't exist in RadiusUser)

---

#### 3. First-Use Expiration Window
**Location:** `internal/radiusd/plugins/auth/checkers/first_use_activator.go:75-94`

```go
validityDuration := time.Duration(batch.ValidityDays) * 24 * time.Hour
newExpire := now.Add(validityDuration)  // NOW + X days
```

✅ User's ExpireTime is set to NOW + validity days (e.g., NOW + 10 days)

---

#### 4. Data Quota Enforcement
**Location:** `internal/radiusd/plugins/auth/checkers/quota_checker.go:31-51`

```go
totalUsage, _ := c.accountingRepo.GetTotalUsage(ctx, user.Username)
if totalUsage >= user.DataQuota*1024*1024 {
    return errors.NewUserQuotaError()  // "quota exceeded"
}
```

✅ Data quota is checked on every login
✅ Returns "quota exceeded" when DataQuota is exhausted

---

### ❌ WHAT DOESN'T WORK

#### 1. Time Quota Not Enforced
**Problem:** There is NO time quota checker in the authentication pipeline

**Evidence:**
- Product has `ValiditySeconds` field (5 hours)
- But RadiusUser does NOT have a TimeQuota field
- No checker enforces time limit during authentication

**Impact:** User can use more than 5 hours within the 10-day window

---

#### 2. Wrong Error Message After Expiration
**Problem:** After validity window, RADIUS says "user expired" instead of "user not exists"

**Current Flow:**
```
1. User tries to login after 10 days
2. ExpireChecker: user.ExpireTime < NOW → true
3. Returns: NewUserExpiredError() → "user expired"
```

**User Requirement:**
```
1. User tries to login after 10 days
2. Should return: NewUserNotExistsError() → "user not exists" (like card not found)
```

**Location:** `internal/radiusd/plugins/auth/checkers/expire_checker.go:22-30`

---

#### 3. No Cleanup of Expired Users
**Problem:** Expired users remain in database forever

**Current Behavior:**
- After 10 days, user expires but still exists in database
- Status remains "enabled"
- Only ExpireTime is in the past

**User Requirement:**
- After 10 days, user should be deleted or marked as "not found"
- Subsequent logins should return "card not found"

---

#### 4. TimeQuota Field Missing
**Problem:** RadiusUser doesn't have a TimeQuota field

**Evidence:** `internal/domain/radius.go:37-75`

**Current Fields:**
```go
DataQuota  int64  // Data quota in MB
ExpireTime time.Time  // Account expiration
// NO TimeQuota field!
```

**Impact:** Cannot track or enforce time limit (5 hours)

---

## Gap Analysis Summary

| Requirement | Status | Gap |
|-------------|--------|-----|
| Product defines data quota (300MB) | ✅ Works | None |
| Product defines time limit (5 hours) | ✅ Defined | ❌ Not enforced |
| User gets product limits on creation | ✅ Works | ❌ Time quota not copied |
| Batch validity window (X days) | ✅ Works | None |
| Data quota enforced during window | ✅ Works | None |
| Time quota enforced during window | ❌ Missing | No time quota checker |
| After window: "card not found" | ❌ Wrong error | Returns "user expired" |
| Cleanup after expiration | ❌ Missing | Users never deleted |

---

## Authentication Pipeline Order

**Current Order (from code):**

1. **Order 5:** FirstUseActivator - Activates first-use users on first login
2. **Order 10:** ExpireChecker - Checks if user expired
3. **Order 15:** QuotaChecker - Checks data quota
4. Other checkers...

**What Happens After 10 Days:**

```
1. User tries to login
2. FirstUseActivator: Skips (already activated)
3. ExpireChecker: user.ExpireTime < NOW → TRUE
4. Returns: "user expired" ❌
   User expects: "user not exists" or "card not found"
```

---

## Root Causes

### Root Cause #1: Time Quota Not Tracked
**Why:** RadiusUser struct doesn't have a TimeQuota field

**Impact:** Cannot enforce the 5-hour time limit

**Evidence:**
- Product.ValiditySeconds exists (5 hours)
- But user.DataQuota exists, user.TimeQuota does NOT

---

### Root Cause #2: Wrong Error After Expiration
**Why:** ExpireChecker returns "user expired" instead of "user not exists"

**User Perspective:**
- "card not found" = The voucher code no longer exists in system
- "user expired" = The user exists but their time is up

**Business Logic:**
- After validity window, the voucher should be considered "invalid"
- Invalid vouchers should return "not found" error

---

### Root Cause #3: No Time Quota Checker
**Why:** No checker in authentication pipeline enforces time limit

**Current Checkers:**
- ExpireChecker: Account expiration
- QuotaChecker: Data quota only
- StatusChecker: User status

**Missing:**
- TimeQuotaChecker: Would check accumulated session time

---

## Proposed Fixes

### Fix #1: Add TimeQuota Field to RadiusUser
**File:** `internal/domain/radius.go`

**Add field:**
```go
TimeQuota int64  `json:"time_quota" form:"time_quota"`  // Time quota in seconds
```

**Update user creation:**
```go
user := domain.RadiusUser{
    DataQuota: product.DataQuota,
    TimeQuota: product.ValiditySeconds,  // 5 hours = 18000 seconds
}
```

---

### Fix #2: Create TimeQuotaChecker
**File:** `internal/radiusd/plugins/auth/checkers/time_quota_checker.go` (NEW)

**Logic:**
```go
func (c *TimeQuotaChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
    user := authCtx.User
    if user == nil || user.TimeQuota <= 0 {
        return nil
    }

    // Get total session time from accounting
    totalTime, err := c.accountingRepo.GetTotalSessionTime(ctx, user.Username)
    if err != nil {
        return nil
    }

    if totalTime >= user.TimeQuota {
        return errors.NewTimeQuotaError()  // "time quota exceeded"
    }

    return nil
}
```

**Order:** 16 (after QuotaChecker at 15)

---

### Fix #3: Change Expired User Error Message
**Option A:** Delete expired users immediately after expiration
- Pros: User truly doesn't exist
- Cons: Data loss, can't track usage history

**Option B:** Change ExpireChecker to return "not exists" error
- Pros: Preserves data, easy to implement
- Cons: User still exists in database

**Option C:** Delete user in ExpireChecker when expired
- Pros: Clean database, correct error
- Cons: Need to recreate user if voucher extended

**Recommended:** Option B (change error message)

**File:** `internal/radiusd/plugins/auth/checkers/expire_checker.go:22-30`

**Change:**
```go
// OLD:
if user.ExpireTime.Before(time.Now()) {
    return errors.NewUserExpiredError()  // "user expired"
}

// NEW:
if user.ExpireTime.Before(time.Now()) {
    return errors.NewUserNotExistsError()  // "user not exists"
}
```

---

### Fix #4: Add Accounting Method for Total Session Time
**File:** `internal/radiusd/repository/gorm/accounting_repository.go`

**Add method:**
```go
func (r *GormAccountingRepository) GetTotalSessionTime(ctx context.Context, username string) (int64, error) {
    var total int64
    err := r.db.WithContext(ctx).
        Model(&domain.RadiusAccounting{}).
        Where("username = ?", username).
        Select("SUM(acct_session_time)").
        Scan(&total).Error
    return total, err
}
```

---

## Test Scenario Verification

### Scenario: Silver Product, 10-Day Window

**Setup:**
```
Product: Silver
- DataQuota: 300MB
- ValiditySeconds: 18000 (5 hours)

Batch:
- ExpirationType: first_use
- ValidityDays: 10
```

**Test Cases:**

1. **Day 0: User logs in first time**
   - ✅ FirstUseActivator sets ExpireTime = NOW + 10 days
   - ✅ User can login
   - ❌ TimeQuota not set (missing field)

2. **Day 5: User has used 250MB, 4 hours**
   - ✅ Can login (quota not exhausted)
   - ❌ Time quota not checked (missing checker)

3. **Day 7: User has used 350MB**
   - ✅ Cannot login (data quota exceeded)
   - ✅ Returns "quota exceeded"

4. **Day 7: User has used 6 hours**
   - ❌ Can still login (time quota not enforced)
   - ❌ Should return "time quota exceeded"

5. **Day 11: User tries to login (10 days expired)**
   - ❌ Returns "user expired"
   - ✅ Should return "user not exists"

---

## Priority Ranking

| Fix | Priority | Complexity | Impact |
|-----|----------|------------|--------|
| Fix #3: Change error message | HIGH | Low | User experience |
| Fix #1: Add TimeQuota field | HIGH | Low | Enable tracking |
| Fix #4: Add GetTotalSessionTime | HIGH | Low | Data retrieval |
| Fix #2: Create TimeQuotaChecker | HIGH | Medium | Enforcement |

---

## Next Steps

1. Implement Fix #3 (quickest win - user experience)
2. Implement Fix #1 + Fix #4 (infrastructure)
3. Implement Fix #2 (enforcement)
4. Test complete scenario
5. Verify all requirements met
