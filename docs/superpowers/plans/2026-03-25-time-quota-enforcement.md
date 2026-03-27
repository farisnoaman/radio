# Time Quota Enforcement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the broken time quota enforcement system where users with 5-hour products see "9 days 59 hours" remaining time instead of "5 hours", and ensure time limits are properly enforced during the validity window.

**Architecture:** The voucher system has two independent time concepts:
1. **Validity Window (batch.ValidityDays):** Time period to use the voucher (e.g., 10 days from first login)
2. **Time Quota (product.ValiditySeconds):** Usage limit within that window (e.g., 5 hours total)

The current implementation confuses these two, treating the validity window as if it were the time quota. This plan separates them: TimeQuota becomes a tracked resource (like DataQuota) that gets depleted with each session, while the validity window remains a fixed expiration date.

**Tech Stack:** Go 1.21+, GORM, SQLite, RADIUS authentication pipeline with plugin-based checkers

---

## File Structure

### Files to Create:
- `internal/radiusd/plugins/auth/checkers/time_quota_checker.go` - New checker to enforce time limits
- `internal/radiusd/plugins/auth/checkers/time_quota_checker_test.go` - Tests for time quota checker

### Files to Modify:
- `internal/domain/radius.go` - Add TimeQuota field to RadiusUser struct
- `internal/radiusd/repository/gorm/accounting_repository.go` - Add GetTotalSessionTime method
- `internal/radiusd/repository/gorm/accounting_repository_test.go` - Tests for GetTotalSessionTime
- `internal/adminapi/vouchers.go` - Set TimeQuota when creating user from voucher
- `internal/radiusd/plugins/auth/checkers/expire_checker.go` - Change error message for expired users
- `internal/radiusd/errors/errors.go` - Add NewTimeQuotaError function
- `internal/radiusd/errors/errors_test.go` - Test NewTimeQuotaError
- `internal/radiusd/radius.go` - Register TimeQuotaChecker in pipeline
- `internal/radiusd/plugins/auth/pipeline.go` - Add TimeQuotaChecker to authentication flow

---

## Task 1: Add TimeQuota Field to RadiusUser

**IMPORTANT NOTE:** The RadiusUser struct **already has** VoucherBatchID and VoucherCode fields (lines 91-94). Only the TimeQuota field needs to be added.

**Files:**
- Modify: `internal/domain/radius.go:37-75`

- [ ] **Step 1: Add TimeQuota field to RadiusUser struct**

Locate the RadiusUser struct definition (around line 37). Add the TimeQuota field after DataQuota:

```go
type RadiusUser struct {
    ID              int64     `json:"id,string" form:"id"`
    TenantID        int64     `gorm:"index" json:"tenant_id" form:"tenant_id"`
    // ... other fields ...
    DataQuota       int64     `json:"data_quota" form:"data_quota"`                     // Data quota in MB (0 = unlimited)
    TimeQuota       int64     `json:"time_quota" form:"time_quota"`                     // Time quota in seconds (0 = unlimited) ← ADD THIS LINE
    Vlanid1         int       `json:"vlanid1" form:"vlanid1"`
    // ... remaining fields ...
}
```

- [ ] **Step 2: Verify code compiles**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go build -o /tmp/radio_test . 2>&1 | head -30`
Expected: No errors (compilation succeeds)

- [ ] **Step 3: Commit**

```bash
git add internal/domain/radius.go
git commit -m "feat(radius): add TimeQuota field to RadiusUser for tracking time usage limits

This adds a TimeQuota field (in seconds) to track cumulative session time,
independent from the account expiration time (ExpireTime).

Related: #time-quota-enforcement"
```

---

## Task 2: Add GetTotalSessionTime Method

**Files:**
- Modify: `internal/radiusd/repository/interfaces.go`
- Modify: `internal/radiusd/repository/gorm/accounting_repository.go`
- Test: `internal/radiusd/repository/gorm/accounting_repository_test.go`

- [ ] **Step 0: Add GetTotalSessionTime to AccountingRepository interface**

Update `internal/radiusd/repository/interfaces.go` to add the method signature to the interface (after line 66):

```go
// AccountingRepository defines accounting record operations
type AccountingRepository interface {
    // Create Create accounting record
    Create(ctx context.Context, accounting *domain.RadiusAccounting) error

    // UpdateStop updates stop time and traffic counters
    UpdateStop(ctx context.Context, sessionId string, accounting *domain.RadiusAccounting) error
    // GetTotalUsage calculates total traffic usage for a user (input + output total)
    GetTotalUsage(ctx context.Context, username string) (int64, error)

    // GetTotalSessionTime retrieves total accumulated session time for a user
    GetTotalSessionTime(ctx context.Context, username string) (int64, error)  // ← ADD THIS
}
```

- [ ] **Step 1: Write failing test for GetTotalSessionTime**

Add to `accounting_repository_test.go`:

```go
func TestGormAccountingRepository_GetTotalSessionTime(t *testing.T) {
    // Setup test database
    db, _ := setupTestDB(t)
    repo := NewGormAccountingRepository(db)
    ctx := context.Background()

    // Create test accounting records
    accounting1 := &domain.RadiusAccounting{
        Username:        "testuser",
        AcctSessionTime: 3600,  // 1 hour
        AcctInputOctets:  1000000,
        AcctOutputOctets: 2000000,
    }
    accounting2 := &domain.RadiusAccounting{
        Username:        "testuser",
        AcctSessionTime: 1800,  // 30 minutes
        AcctInputOctets:  500000,
        AcctOutputOctets: 1000000,
    }
    repo.Create(ctx, accounting1)
    repo.Create(ctx, accounting2)

    // Test
    totalTime, err := repo.GetTotalSessionTime(ctx, "testuser")

    // Assert
    assert.Nil(t, err)
    assert.Equal(t, int64(5400), totalTime)  // 3600 + 1800 = 5400 seconds (1.5 hours)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go test ./internal/radiusd/repository/gorm/... -run TestGormAccountingRepository_GetTotalSessionTime -v`
Expected: FAIL with "undefined: GetTotalSessionTime"

- [ ] **Step 3: Implement GetTotalSessionTime method**

Add to `accounting_repository.go` (after GetTotalUsage method, around line 65):

```go
// GetTotalSessionTime retrieves the total accumulated session time for a user
func (r *GormAccountingRepository) GetTotalSessionTime(ctx context.Context, username string) (int64, error) {
    var total int64
    err := r.db.WithContext(ctx).
        Model(&domain.RadiusAccounting{}).
        Where("username = ?", username).
        Select("COALESCE(SUM(acct_session_time), 0)").
        Scan(&total).Error
    return total, err
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go test ./internal/radiusd/repository/gorm/... -run TestGormAccountingRepository_GetTotalSessionTime -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/radiusd/repository/gorm/accounting_repository.go internal/radiusd/repository/gorm/accounting_repository_test.go
git commit -m "feat(accounting): add GetTotalSessionTime to aggregate user session time

This sums up acct_session_time from all accounting records for a user,
enabling enforcement of time-based quotas (e.g., 5-hour limits).

Tests: Add testGetTotalSessionTime verifying aggregation logic

Related: #time-quota-enforcement"
```

---

## Task 3: Add NewTimeQuotaError Function

**Files:**
- Modify: `internal/radiusd/errors/errors.go`
- Test: `internal/radiusd/errors/errors_test.go`

- [ ] **Step 1: Write failing test for NewTimeQuotaError**

Add to `errors_test.go`:

```go
func TestNewTimeQuotaError(t *testing.T) {
    err := NewTimeQuotaError()
    assert.NotNil(t, err)

    authErr, ok := GetAuthError(err)
    assert.True(t, ok)
    assert.Equal(t, app.MetricsRadiusRejectQuota, authErr.MetricsType)
    assert.Equal(t, "time quota exceeded", authErr.Message)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go test ./internal/radiusd/errors/... -run TestNewTimeQuotaError -v`
Expected: FAIL with "undefined: NewTimeQuotaError"

- [ ] **Step 3: Implement NewTimeQuotaError function**

Add to `errors.go` (after NewUserQuotaError, around line 234):

```go
// NewTimeQuotaError creates an error for users who have exceeded their time quota
func NewTimeQuotaError() error {
    return NewAuthError(app.MetricsRadiusRejectQuota, "time quota exceeded")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go test ./internal/radiusd/errors/... -run TestNewTimeQuotaError -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/radiusd/errors/errors.go internal/radiusd/errors/errors_test.go
git commit -m "feat(errors): add NewTimeQuotaError for time limit enforcement

Returns 'time quota exceeded' error when users exceed their allowed
total session time (e.g., 5 hours across all sessions).

Tests: Add TestNewTimeQuotaError verifying error type and message

Related: #time-quota-enforcement"
```

---

## Task 4: Create TimeQuotaChecker

**Files:**
- Create: `internal/radiusd/plugins/auth/checkers/time_quota_checker.go`
- Test: `internal/radiusd/plugins/auth/checkers/time_quota_checker_test.go`

- [ ] **Step 1: Write failing test for TimeQuotaChecker**

Create `time_quota_checker_test.go`:

```go
package checkers

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
    "github.com/talkincode/toughradius/v9/internal/radiusd/repository/gorm"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
)

func TestTimeQuotaChecker_Check(t *testing.T) {
    // Setup in-memory SQLite database
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    assert.NoError(t, err)

    // Migrate schema
    err = db.AutoMigrate(&domain.RadiusAccounting{})
    assert.NoError(t, err)

    // Create real repository (not mocks)
    repo := gormrepo.NewGormAccountingRepository(db)
    checker := NewTimeQuotaChecker(repo)
    ctx := context.Background()

    t.Run("allows login when time quota not set", func(t *testing.T) {
        user := &domain.RadiusUser{
            Username:  "testuser",
            TimeQuota: 0,  // No time quota
        }
        authCtx := &auth.AuthContext{User: user}

        err := checker.Check(ctx, authCtx)
        assert.Nil(t, err)
    })

    t.Run("allows login when time quota not exceeded", func(t *testing.T) {
        // Create accounting record with 1 hour used
        accounting := &domain.RadiusAccounting{
            Username:        "testuser2",
            AcctSessionTime:  3600,  // 1 hour
            AcctInputOctets:  1000000,
            AcctOutputOctets: 2000000,
        }
        err = repo.Create(ctx, accounting)
        assert.NoError(t, err)

        user := &domain.RadiusUser{
            Username:  "testuser2",
            TimeQuota: 18000,  // 5 hours
        }
        authCtx := &auth.AuthContext{User: user}

        err = checker.Check(ctx, authCtx)
        assert.Nil(t, err)  // Should allow (1 hour < 5 hours)
    })

    t.Run("rejects login when time quota exceeded", func(t *testing.T) {
        // Create accounting records totaling 5.5 hours
        accounting1 := &domain.RadiusAccounting{
            Username:        "testuser3",
            AcctSessionTime:  18000, // 5 hours
            AcctInputOctets:  1000000,
            AcctOutputOctets: 2000000,
        }
        accounting2 := &domain.RadiusAccounting{
            Username:        "testuser3",
            AcctSessionTime:  1800,  // 30 minutes
            AcctInputOctets:  500000,
            AcctOutputOctets: 1000000,
        }
        err = repo.Create(ctx, accounting1)
        assert.NoError(t, err)
        err = repo.Create(ctx, accounting2)
        assert.NoError(t, err)

        user := &domain.RadiusUser{
            Username:  "testuser3",
            TimeQuota: 18000,  // 5 hours
        }
        authCtx := &auth.Context{User: user}

        err = checker.Check(ctx, authCtx)
        assert.NotNil(t, err)  // Should reject (5.5 hours > 5 hours)
    })
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go test ./internal/radiusd/plugins/auth/checkers/... -run TestTimeQuotaChecker -v`
Expected: FAIL with "undefined: NewTimeQuotaChecker" or import errors

- [ ] **Step 3: Implement TimeQuotaChecker**

Create `time_quota_checker.go`:

```go
package checkers

import (
    "context"

    "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
    "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
    "github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// TimeQuotaChecker checks whether the user has exceeded their time quota
type TimeQuotaChecker struct {
    accountingRepo repository.AccountingRepository
}

// NewTimeQuotaChecker creates a time quota checker instance
func NewTimeQuotaChecker(accountingRepo repository.AccountingRepository) *TimeQuotaChecker {
    return &TimeQuotaChecker{
        accountingRepo: accountingRepo,
    }
}

func (c *TimeQuotaChecker) Name() string {
    return "time_quota"
}

func (c *TimeQuotaChecker) Order() int {
    return 16 // Execute after QuotaChecker (15)
}

func (c *TimeQuotaChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
    user := authCtx.User
    if user == nil || user.TimeQuota <= 0 {
        return nil  // No time quota configured, allow login
    }

    // Get total session time from accounting records
    totalTime, err := c.accountingRepo.GetTotalSessionTime(ctx, user.Username)
    if err != nil {
        // Log error but allow login on check failure
        // (safer to allow than to block on accounting errors)
        return nil
    }

    // TimeQuota is in seconds, totalTime is in seconds
    if totalTime >= user.TimeQuota {
        return errors.NewTimeQuotaError()
    }

    return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go test ./internal/radiusd/plugins/auth/checkers/... -run TestTimeQuotaChecker -v`
Expected: PASS (after fixing any mock setup issues)

- [ ] **Step 5: Commit**

```bash
git add internal/radiusd/plugins/auth/checkers/time_quota_checker.go internal/radiusd/plugins/auth/checkers/time_quota_checker_test.go
git commit -m "feat(auth): add TimeQuotaChecker to enforce time usage limits

This checker runs during authentication to verify users haven't exceeded
their total allowed session time (e.g., 5 hours across all sessions).

Order: 16 (after QuotaChecker at 15)

Tests: Add testTimeQuotaChecker with coverage for:
- No time quota configured (allow)
- Within quota (allow)
- Exceeded quota (reject with error)

Related: #time-quota-enforcement"
```

---

## Task 5: Update Voucher Redemption to Set Voucher Linkage Fields

**CRITICAL BUG:** Voucher expiry is not enforced because VoucherBatchID and VoucherCode are not set on user creation!

**IMPORTANT:** The VoucherBatchID and VoucherCode fields **already exist** in the RadiusUser schema (lines 91-94). They just need to be **populated** when creating users. Only the TimeQuota field was added in Task 1.

**Files:**
- Modify: `internal/adminapi/vouchers.go:656-669`

- [ ] **Step 1: Locate user creation code in RedeemVoucher**

Find the user creation section in RedeemVoucher function (around line 656).

- [ ] **Step 2: Populate VoucherLinkage fields and TimeQuota**

Locate the RadiusUser initialization. Add VoucherBatchID, VoucherCode, and TimeQuota fields:

```go
user := domain.RadiusUser{
    TenantID:        tenant.GetTenantIDOrDefault(c.Request().Context()),
    Username:        voucher.Code,
    Password:        voucher.Code,
    ProfileId:       profile.ID,
    Status:          "enabled",
    ExpireTime:      expireTime,
    CreatedAt:       now,
    UpdatedAt:       now,
    UpRate:          userUpRate,
    DownRate:        userDownRate,
    DataQuota:       userDataQuota,
    TimeQuota:       product.ValiditySeconds,  // ← ADD THIS: Time quota from product
    VoucherBatchID:  voucher.BatchID,           // ← ADD THIS: Link to voucher batch
    VoucherCode:     voucher.Code,              // ← ADD THIS: Link to voucher code
    AddrPool:        profile.AddrPool,
}
```

**WHY THIS CRITICAL FIX IS NEEDED:**

Without VoucherBatchID and VoucherCode:
- VoucherAuthChecker is bypassed (checks `if user.VoucherBatchID == 0`)
- Voucher expiry is NEVER validated during authentication
- Users can login AFTER voucher expiry date

With these fields set:
- VoucherAuthChecker runs and validates voucher expiry on every login
- Voucher data/time quotas are enforced
- Batch-level controls work correctly

- [ ] **Step 3: Verify code compiles**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go build -o /tmp/radio_test . 2>&1 | head -30`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/adminapi/vouchers.go
git commit -m "fix(vouchers): set voucher linkage fields and TimeQuota when creating users

CRITICAL FIX: This fixes voucher expiry enforcement!

Previously, VoucherBatchID and VoucherCode were not set when creating
users from vouchers, causing VoucherAuthChecker to be completely bypassed
during authentication. This meant:
- Voucher expiry dates were never validated after account creation
- Users could login indefinitely even after voucher expiry
- Voucher-level controls were ineffective

Changes:
1. Add VoucherBatchID field (links user to voucher batch)
2. Add VoucherCode field (links user to specific voucher)
3. Add TimeQuota field (enforces time limits from product)
4. Enables VoucherAuthChecker to properly validate voucher expiry

This ensures:
- Voucher expiry is enforced on every login (not just redemption)
- Voucher data/time quotas work correctly
- Batch-level controls are active

Related: #time-quota-enforcement #voucher-expiry-fix"
```

---

## Task 6: Change Expired User Error Message

**Files:**
- Modify: `internal/radiusd/plugins/auth/checkers/expire_checker.go:22-30`

- [ ] **Step 1: Locate error return in ExpireChecker**

Find the Check method in ExpireChecker (around line 22).

- [ ] **Step 2: Change error from NewUserExpiredError to NewUserNotExistsError**

Update the code:

```go
// OLD CODE:
if user.ExpireTime.Before(time.Now()) {
    return errors.NewUserExpiredError()
}

// NEW CODE:
if user.ExpireTime.Before(time.Now()) {
    return errors.NewUserNotExistsError()
}
```

- [ ] **Step 3: Verify code compiles**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go build -o /tmp/radio_test . 2>&1 | head -30`
Expected: No errors

- [ ] **Step 4: Update ExpireChecker test if needed**

Check if `expire_checker_test.go` exists and update assertions to expect NewUserNotExistsError instead of NewUserExpiredError.

- [ ] **Step 5: Run tests**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go test ./internal/radiusd/plugins/auth/checkers/... -run TestExpireChecker -v`
Expected: PASS (after updating test expectations)

- [ ] **Step 6: Commit**

```bash
git add internal/radiusd/plugins/auth/checkers/expire_checker.go internal/radiusd/plugins/auth/checkers/expire_checker_test.go
git commit -m "fix(auth): return 'user not exists' for expired users instead of 'user expired'

After the validity window (e.g., 10 days), expired users should appear
as if they don't exist, matching the 'card not found' requirement.

This changes ExpireChecker to return NewUserNotExistsError instead of
NewUserExpiredError when user.ExpireTime < NOW.

Related: #time-quota-enforcement"
```

---

## Task 7: Register TimeQuotaChecker in Authentication Pipeline

**Files:**
- Modify: `internal/radiusd/plugins/init.go`

- [ ] **Step 1: Locate checker registration in init.go**

Open `internal/radiusd/plugins/init.go` and find where authentication checkers are registered (search for `NewQuotaChecker` or `quotaChecker :=`).

- [ ] **Step 2: Add TimeQuotaChecker instantiation**

Add checker creation following the existing pattern (look for where QuotaChecker is created):

```go
// Find the section where checkers are instantiated
// Add TimeQuotaChecker after QuotaChecker:

timeQuotaChecker := checkers.NewTimeQuotaChecker(s.AccountingRepo)
```

- [ ] **Step 3: Add TimeQuotaChecker to checker list**

Find where checkers are added to the authentication service. Add TimeQuotaChecker to the list:

```go
// Look for where checkers are added (e.g., checkers: []auth.Checker{...})
// Add TimeQuotaChecker to the list:

timeQuotaChecker,
```

- [ ] **Step 4: Verify checker order**

Ensure the pipeline order is:
1. Order 5: FirstUseActivator
2. Order 10: ExpireChecker
3. Order 15: QuotaChecker
4. Order 16: TimeQuotaChecker ← should come after QuotaChecker

- [ ] **Step 5: Verify code compiles**

Run: `export PATH=$PATH:/home/faris/go/go/bin && go build -o /tmp/radio_test . 2>&1 | head -30`
Expected: No errors

- [ ] **Step 6: Commit**

```bash
git add internal/radiusd/plugins/init.go
git commit -m "feat(auth): register TimeQuotaChecker in authentication pipeline

Add TimeQuotaChecker to the authentication flow to enforce time-based
quotas during user login.

Pipeline Order: 16 (after QuotaChecker at 15)

Related: #time-quota-enforcement"
```

---

## Task 8: End-to-End Integration Test

**Files:**
- Test: Create manual test or integration test

- [ ] **Step 1: Create test product with time quota**

Run SQL or use admin UI:
```sql
INSERT INTO product (name, data_quota, validity_seconds, status)
VALUES ('5-Hour Test', 300, 18000, 'active');
-- 300MB data quota, 18000 seconds = 5 hours
```

- [ ] **Step 2: Create test batch**

Use admin UI or API:
```json
{
  "name": "Time Quota Test Batch",
  "product_id": "<product_id_from_step_1>",
  "count": 5,
  "expiration_type": "first_use",
  "validity_days": 10
}
```

- [ ] **Step 3: Redeem a voucher**

Redeem one voucher from the batch (e.g., TEST001)

- [ ] **Step 4: Verify TimeQuota was set**

Run SQL:
```sql
SELECT username, data_quota, time_quota, expire_time
FROM radius_user
WHERE username = 'TEST001';
```

Expected:
```
username: TEST001
data_quota: 300
time_quota: 18000  ← Should be 18000 seconds (5 hours)
expire_time: 9999-12-31 (will be updated on first login)
```

- [ ] **Step 5: First login - activate user**

Login with TEST001. Check logs for FirstUseActivator:
```bash
tail -f backend.log | grep first_use_activator
```

Expected: User ExpireTime updated to NOW + 10 days

- [ ] **Step 6: Check status page shows correct remaining time**

Login again, check status page. Should show:
- **Remaining Time:** ~5 hours (NOT 10 days!)
- **Remaining Data:** ~300MB

- [ ] **Step 7: Test time quota enforcement**

Use the voucher for ~5 hours across multiple sessions. After exceeding 5 hours total:
- Login should be rejected
- Error should be "time quota exceeded"

- [ ] **Step 8: Test data quota still works**

After resetting, use 300MB of data. Login should be rejected with "quota exceeded"

- [ ] **Step 9: Test voucher expiry enforcement**

Create a batch with specific expiry date (e.g., yesterday):
```json
{
  "name": "Expiry Test Batch",
  "product_id": "<product_id>",
  "count": 1,
  "expire_time": "2026-03-24 00:00:00",  // Yesterday
  "expiration_type": "fixed"
}
```

Try to redeem the voucher. Expected:
```json
{
  "error": "VOUCHER_EXPIRED",
  "message": "Voucher has expired"
}
```

- [ ] **Step 10: Test voucher expiry enforced during login**

Manually update a voucher's expiry to past and create a user:
```sql
-- Create voucher with expired date
INSERT INTO voucher (code, batch_id, status, expire_time)
VALUES ('EXPIRED001', 1, 'active', '2026-03-24 00:00:00');

-- Create user with voucher linkage
INSERT INTO radius_user (username, voucher_batch_id, voucher_code, status)
VALUES ('EXPIRED001', 1, 'EXPIRED001', 'enabled');
```

Try to login with EXPIRED001. Expected:
- Login rejected
- Error: "Voucher has expired" (from VoucherAuthChecker)

- [ ] **Step 11: Test validity window expiration**

Wait for 10 days to pass, or manually update ExpireTime to past:
```sql
UPDATE radius_user SET expire_time = datetime('now', '-1 day') WHERE username = 'TEST001';
```

Login should be rejected with "user not exists" (NOT "user expired")

- [ ] **Step 12: Document test results**

Create test report documenting:
- Time quota correctly enforced after 5 hours
- Data quota correctly enforced after 300MB
- Validity window correctly enforced after 10 days
- Voucher expiry enforced on every login
- Error messages match requirements

---

## Verification Checklist

After completing all tasks:

### Database Schema
- [ ] RadiusUser has TimeQuota field
- [ ] RadiusUser has VoucherBatchID field
- [ ] RadiusUser has VoucherCode field
- [ ] Users created from vouchers have all three fields set correctly

### Authentication Pipeline
- [ ] VoucherAuthChecker is registered and running
- [ ] VoucherAuthChecker checks user.VoucherBatchID != 0 (not bypassed)
- [ ] TimeQuotaChecker is registered in pipeline
- [ ] TimeQuotaChecker runs at Order 16 (after QuotaChecker)
- [ ] GetTotalSessionTime returns correct sum of acct_session_time

### Voucher Expiry Enforcement (CRITICAL)
- [ ] Voucher expiry date is checked on EVERY login (not just redemption)
- [ ] After voucher expiry date, login rejected with "Voucher has expired"
- [ ] User-created vouchers with expiry dates work correctly
- [ ] Default 2999-12-31 expiry allows indefinite activation

### Time Quota Enforcement
- [ ] User with 5-hour quota can use up to 5 hours across multiple sessions
- [ ] After 5 hours, login rejected with "time quota exceeded"
- [ ] Status page shows remaining time in hours (not days)
- [ ] Data quota (300MB) still enforced independently

### Validity Window Expiration
- [ ] After validity window (10 days), returns "user not exists"

### Code Quality
- [ ] All tests pass: `go test ./...`
- [ ] Code compiles without warnings
- [ ] No regression in existing quota expiration checks

---

## Rollback Plan

If issues arise:

1. **Revert Task 7** (Remove TimeQuotaChecker from pipeline)
2. **Revert Task 5** (Don't set TimeQuota in voucher redemption)
3. **Revert Task 6** (Restore original error message)

Tasks 1-4 can remain safely (database schema, repository method, error function) as they don't affect behavior until checker is registered.

---

## Related Documentation

- **Analysis:** `docs/VOUCHER_SYSTEM_ANALYSIS.md` - Detailed gap analysis
- **Complete Guide:** `docs/VOUCHER_EXPIRATION_COMPLETE_GUIDE.md` - System overview
- **Bug Resolution:** `docs/VOUCHER_BUG_RESOLUTION.md` - Previous fixes

---

**Estimated Total Time:** 2-3 hours (including testing)

**Risk Level:** Medium (adds new enforcement, may affect existing users if they have time quotas set)

**Migration Required:** No database migration needed (TimeQuota field uses default 0 value for existing users)
