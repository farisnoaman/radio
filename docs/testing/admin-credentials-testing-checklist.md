# Admin Credentials Management - Testing Checklist

**Feature:** Tenant Admin Credentials Management
**Date:** 2026-03-24
**Version:** v9.0.0

---

## 1. Backend API Tests

### 1.1 GET /api/v1/platform/providers/:id/admin-credentials

**Test Cases:**

- [ ] **TC-GET-001:** Retrieve credentials as platform admin
  - Setup: Create provider with custom admin credentials
  - Action: Call GET endpoint with provider ID
  - Expected: 200 OK with masked password (`********`)

- [ ] **TC-GET-002:** Retrieve credentials as tenant admin (should fail)
  - Setup: Login as tenant admin
  - Action: Call GET endpoint
  - Expected: 403 Forbidden

- [ ] **TC-GET-003:** Retrieve credentials for non-existent provider
  - Setup: Use invalid provider ID
  - Action: Call GET endpoint
  - Expected: 404 Not Found

- [ ] **TC-GET-004:** Retrieve credentials without authentication
  - Setup: No auth token
  - Action: Call GET endpoint
  - Expected: 401 Unauthorized

**cURL Commands:**
```bash
# Test GET as platform admin
curl -X GET http://localhost:1816/api/v1/platform/providers/1/admin-credentials \
  -H "Authorization: Bearer PLATFORM_ADMIN_TOKEN"

# Expected response:
{
  "username": "admin",
  "password": "********",
  "level": "admin",
  "status": "enabled"
}
```

### 1.2 PUT /api/v1/platform/providers/:id/admin-credentials

**Test Cases:**

- [ ] **TC-PUT-001:** Update username and password successfully
  - Setup: Create provider, login as platform admin
  - Action: Send PUT with new credentials
  - Expected: 200 OK with updated credentials

- [ ] **TC-PUT-002:** Update with invalid username (too short)
  - Setup: Use username < 3 characters
  - Action: Send PUT request
  - Expected: 400 Bad Request with validation error

- [ ] **TC-PUT-003:** Update with invalid username (too long)
  - Setup: Use username > 50 characters
  - Action: Send PUT request
  - Expected: 400 Bad Request with validation error

- [ ] **TC-PUT-004:** Update with invalid password (too short)
  - Setup: Use password < 6 characters
  - Action: Send PUT request
  - Expected: 400 Bad Request with validation error

- [ ] **TC-PUT-005:** Update with non-alphanumeric username
  - Setup: Use username with special characters
  - Action: Send PUT request
  - Expected: 400 Bad Request with validation error

- [ ] **TC-PUT-006:** Update credentials as tenant admin (should fail)
  - Setup: Login as tenant admin
  - Action: Send PUT request
  - Expected: 403 Forbidden

- [ ] **TC-PUT-007:** Update credentials for non-existent provider
  - Setup: Use invalid provider ID
  - Action: Send PUT request
  - Expected: 404 Not Found

**cURL Commands:**
```bash
# Test PUT update
curl -X PUT http://localhost:1816/api/v1/platform/providers/1/admin-credentials \
  -H "Authorization: Bearer PLATFORM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newadmin",
    "password": "newpass123"
  }'

# Expected response:
{
  "username": "newadmin",
  "password": "newpass123",
  "level": "admin",
  "status": "enabled"
}
```

### 1.3 POST /api/v1/platform/providers/:id/admin-credentials/reset

**Test Cases:**

- [ ] **TC-POST-001:** Reset to default credentials successfully
  - Setup: Create provider with custom credentials
  - Action: Send POST to reset endpoint
  - Expected: 200 OK with `admin` / `123456`

- [ ] **TC-POST-002:** Reset as tenant admin (should fail)
  - Setup: Login as tenant admin
  - Action: Send POST request
  - Expected: 403 Forbidden

- [ ] **TC-POST-003:** Reset credentials for non-existent provider
  - Setup: Use invalid provider ID
  - Action: Send POST request
  - Expected: 404 Not Found

- [ ] **TC-POST-004:** Verify audit log entry created
  - Setup: Reset credentials
  - Action: Query audit logs
  - Expected: Log entry with timestamp, operator, provider ID

**cURL Commands:**
```bash
# Test POST reset
curl -X POST http://localhost:1816/api/v1/platform/providers/1/admin-credentials/reset \
  -H "Authorization: Bearer PLATFORM_ADMIN_TOKEN"

# Expected response:
{
  "username": "admin",
  "password": "123456",
  "level": "admin",
  "status": "enabled"
}
```

---

## 2. Frontend Component Tests

### 2.1 AdminCredentialsSection Component

**Test Cases:**

- [ ] **TC-UI-001:** Component renders in provider creation form
  - Setup: Navigate to Platform → Providers → Create
  - Expected: "Admin Credentials" section visible

- [ ] **TC-UI-002:** Default credentials radio button selected by default
  - Setup: Open provider creation form
  - Expected: "Use default credentials (admin/123456)" checked

- [ ] **TC-UI-003:** Custom credentials form appears when selected
  - Setup: Click "Set custom credentials"
  - Expected: Username and password fields appear

- [ ] **TC-UI-004:** Username validation shows error for < 3 chars
  - Setup: Enter 2-character username
  - Expected: "Username must be at least 3 characters" error

- [ ] **TC-UI-005:** Username validation shows error for > 50 chars
  - Setup: Enter 51-character username
  - Expected: "Username must be at most 50 characters" error

- [ ] **TC-UI-006:** Password validation shows error for < 6 chars
  - Setup: Enter 5-character password
  - Expected: "Password must be at least 6 characters" error

- [ ] **TC-UI-007:** Confirm password validation shows error when mismatch
  - Setup: Enter different passwords
  - Expected: "Passwords do not match" error

- [ ] **TC-UI-008:** Submit button disabled when validation fails
  - Setup: Enter invalid credentials
  - Expected: Submit button disabled or shows validation error

- [ ] **TC-UI-009:** Collapsible section expands/collapses
  - Setup: Click section header
  - Expected: Form toggles visibility

### 2.2 AdminCredentialsCard Component

**Test Cases:**

- [ ] **TC-UI-010:** Card renders on provider detail page
  - Setup: Navigate to Platform → Providers → [Click Provider]
  - Expected: "Admin Credentials" card visible

- [ ] **TC-UI-011:** Card shows current username correctly
  - Setup: View provider with custom admin username
  - Expected: Username displayed correctly

- [ ] **TC-UI-012:** Card shows masked password by default
  - Setup: View any provider
  - Expected: Password shows as `********`

- [ ] **TC-UI-013:** Copy button copies password to clipboard
  - Setup: Click copy button on password field
  - Expected: Password copied to clipboard

- [ ] **TC-UI-014:** Copy button copies username to clipboard
  - Setup: Click copy button on username field
  - Expected: Username copied to clipboard

- [ ] **TC-UI-015:** "Reset to Defaults" button visible
  - Setup: View provider detail page
  - Expected: Button visible and clickable

- [ ] **TC-UI-016:** "Update Credentials" button visible
  - Setup: View provider detail page
  - Expected: Button visible and clickable

### 2.3 ResetPasswordDialog Component

**Test Cases:**

- [ ] **TC-UI-017:** Dialog opens when "Update Credentials" clicked
  - Setup: Click "Update Credentials" button
  - Expected: Modal dialog opens

- [ ] **TC-UI-018:** Dialog shows current username
  - Setup: Open dialog
  - Expected: Current username pre-filled

- [ ] **TC-UI-019:** Dialog validates new password
  - Setup: Enter invalid password
  - Expected: Validation error shown

- [ ] **TC-UI-020:** Dialog validates confirm password match
  - Setup: Enter mismatched passwords
  - Expected: "Passwords do not match" error

- [ ] **TC-UI-021:** Cancel button closes dialog
  - Setup: Click Cancel
  - Expected: Dialog closes without saving

- [ ] **TC-UI-022:** Update button sends API request
  - Setup: Enter valid credentials, click Update
  - Expected: API call made, dialog closes on success

- [ ] **TC-UI-023:** Loading spinner shows during API call
  - Setup: Submit form
  - Expected: Spinner or loading state visible

- [ ] **TC-UI-024:** Success toast appears on successful update
  - Setup: Complete update successfully
  - Expected: Green toast with success message

- [ ] **TC-UI-025:** Error toast appears on failed update
  - Setup: Simulate API error
  - Expected: Red toast with error message

---

## 3. Integration Tests

### 3.1 Provider Creation Flow

**Test Cases:**

- [ ] **TC-INT-001:** Create provider with default credentials
  - Setup: Platform → Providers → Create
  - Action: Fill provider info, use default credentials
  - Expected: Provider created with admin/123456

- [ ] **TC-INT-002:** Create provider with custom credentials
  - Setup: Platform → Providers → Create
  - Action: Fill provider info, set custom credentials
  - Expected: Provider created with custom admin

- [ ] **TC-INT-003:** Verify admin can login with new credentials
  - Setup: Create provider with custom credentials
  - Action: Login as tenant admin
  - Expected: Successful login, tenant-scoped access

- [ ] **TC-INT-004:** Verify old credentials don't work after update
  - Setup: Update provider admin credentials
  - Action: Try login with old credentials
  - Expected: Login fails

### 3.2 Provider Registration Approval Flow

**Test Cases:**

- [ ] **TC-INT-005:** Approve registration with default credentials
  - Setup: Pending provider registration
  - Action: Approve without changing credentials
  - Expected: Provider created with admin/123456

- [ ] **TC-INT-006:** Approve registration with custom credentials
  - Setup: Pending provider registration
  - Action: Approve with custom credentials
  - Expected: Provider created with custom admin

- [ ] **TC-INT-007:** Verify credentials shown in approval response
  - Setup: Approve registration
  - Action: Check API response
  - Expected: Response includes username and password

### 3.3 Provider Detail Page Flow

**Test Cases:**

- [ ] **TC-INT-008:** View provider admin credentials
  - Setup: Navigate to provider detail page
  - Action: Look at Admin Credentials card
  - Expected: Current credentials displayed (password masked)

- [ ] **TC-INT-009:** Reset credentials to defaults
  - Setup: Provider with custom credentials
  - Action: Click "Reset to Defaults"
  - Expected: Credentials reset to admin/123456

- [ ] **TC-INT-010:** Update credentials via detail page
  - Setup: Navigate to provider detail page
  - Action: Click "Update Credentials", enter new credentials
  - Expected: Credentials updated successfully

- [ ] **TC-INT-011:** Copy credentials to clipboard
  - Setup: View provider detail page
  - Action: Click copy buttons
  - Expected: Credentials copied, can paste elsewhere

---

## 4. Security Tests

### 4.1 Authorization Tests

**Test Cases:**

- [ ] **TC-SEC-001:** Platform admin can access all endpoints
  - Setup: Login as super admin
  - Action: Call GET, PUT, POST endpoints
  - Expected: All requests succeed

- [ ] **TC-SEC-002:** Tenant admin cannot access endpoints
  - Setup: Login as tenant admin
  - Action: Call GET, PUT, POST endpoints
  - Expected: All requests return 403 Forbidden

- [ ] **TC-SEC-003:** Regular operator cannot access endpoints
  - Setup: Login as operator
  - Action: Call GET, PUT, POST endpoints
  - Expected: All requests return 403 Forbidden

- [ ] **TC-SEC-004:** Unauthenticated requests rejected
  - Setup: No auth token
  - Action: Call any endpoint
  - Expected: 401 Unauthorized

### 4.2 Password Security Tests

**Test Cases:**

- [ ] **TC-SEC-005:** Password not logged in plain text
  - Setup: Update credentials
  - Action: Check backend logs
  - Expected: No password in logs (only masked or hash)

- [ ] **TC-SEC-006:** Password masked in GET response
  - Setup: Call GET endpoint
  - Action: Check response
  - Expected: Password shows as `********`

- [ ] **TC-SEC-007:** Password only shown once during creation
  - Setup: Create provider
  - Action: Check response
  - Expected: Password shown in creation response only

- [ ] **TC-SEC-008:** SQL injection attempts blocked
  - Setup: Enter SQL injection in username
  - Action: Submit form
  - Expected: Validation error or sanitized input

- [ ] **TC-SEC-009:** XSS attempts blocked
  - Setup: Enter XSS payload in username
  - Action: Submit form
  - Expected: Input sanitized or validation error

### 4.3 Validation Tests

**Test Cases:**

- [ ] **TC-SEC-010:** Username length validation enforced
  - Setup: Enter 2-char and 51-char usernames
  - Action: Submit form
  - Expected: Validation errors for both

- [ ] **TC-SEC-011:** Password length validation enforced
  - Setup: Enter 5-char password
  - Action: Submit form
  - Expected: Validation error

- [ ] **TC-SEC-012:** Username character validation enforced
  - Setup: Enter username with special chars (!@#$%)
  - Action: Submit form
  - Expected: Validation error

- [ ] **TC-SEC-013:** Required fields validated
  - Setup: Submit empty username or password
  - Action: Submit form
  - Expected: Validation errors

### 4.4 Rate Limiting Tests

**Test Cases:**

- [ ] **TC-SEC-014:** Rate limiting enforced on password reset
  - Setup: Attempt 4 resets in 1 hour
  - Action: Send POST requests
  - Expected: 4th request returns 429 Too Many Requests

- [ ] **TC-SEC-015:** Failed login attempts trigger lockout
  - Setup: Attempt 6 failed logins
  - Action: Try login again
  - Expected: Account locked for 15 minutes

---

## 5. Database Verification Tests

**Test Cases:**

- [ ] **TC-DB-001:** Verify tenant admin created in sys_opr table
  - Setup: Create new provider
  - Action: Query database
  - Expected: Row in sys_opr with tenant_id = provider_id

- [ ] **TC-DB-002:** Verify username updated correctly
  - Setup: Update admin username
  - Action: Query database
  - Expected: username field updated

- [ ] **TC-DB-003:** Verify password updated correctly
  - Setup: Update admin password
  - Action: Query database
  - Expected: password field updated

- [ ] **TC-DB-004:** Verify level set to 'admin'
  - Setup: Create provider
  - Action: Query database
  - Expected: level = 'admin'

- [ ] **TC-DB-005:** Verify status set to 'enabled'
  - Setup: Create provider
  - Action: Query database
  - Expected: status = 'enabled'

- [ ] **TC-DB-006:** Verify tenant_id matches provider_id
  - Setup: Create provider
  - Action: Query database
  - Expected: sys_opr.tenant_id = provider.id

**SQL Verification Queries:**
```sql
-- Check tenant admin exists
SELECT id, username, level, status, tenant_id
FROM sys_opr
WHERE tenant_id = 1 AND level = 'admin';

-- Check username updated
SELECT username FROM sys_opr
WHERE tenant_id = 1 AND level = 'admin';

-- Check all provider admins
SELECT
  o.id,
  o.username,
  o.level,
  o.status,
  p.name as provider_name
FROM sys_opr o
LEFT JOIN provider p ON o.tenant_id = p.id
WHERE o.level = 'admin';
```

---

## 6. Performance Tests

**Test Cases:**

- [ ] **TC-PERF-001:** GET response time < 200ms
  - Setup: Call GET endpoint
  - Action: Measure response time
  - Expected: < 200ms

- [ ] **TC-PERF-002:** PUT response time < 300ms
  - Setup: Call PUT endpoint
  - Action: Measure response time
  - Expected: < 300ms

- [ ] **TC-PERF-003:** POST response time < 300ms
  - Setup: Call POST endpoint
  - Action: Measure response time
  - Expected: < 300ms

- [ ] **TC-PERF-004:** Concurrent requests handled correctly
  - Setup: Send 10 simultaneous requests
  - Action: Monitor for errors
  - Expected: All requests succeed

- [ ] **TC-PERF-005:** UI renders without lag
  - Setup: Open provider detail page
  - Action: Measure render time
  - Expected: < 500ms to full render

---

## 7. Audit Log Tests

**Test Cases:**

- [ ] **TC-AUDIT-001:** Credential creation logged
  - Setup: Create provider with custom credentials
  - Action: Check audit logs
  - Expected: Log entry with action, operator, timestamp

- [ ] **TC-AUDIT-002:** Credential update logged
  - Setup: Update admin credentials
  - Action: Check audit logs
  - Expected: Log entry with old/new username (no password)

- [ ] **TC-AUDIT-003:** Credential reset logged
  - Setup: Reset to defaults
  - Action: Check audit logs
  - Expected: Log entry with reset action

- [ ] **TC-AUDIT-004:** IP address logged
  - Setup: Make credential change
  - Action: Check audit logs
  - Expected: IP address recorded

- [ ] **TC-AUDIT-005:** Operator information logged
  - Setup: Make credential change
  - Action: Check audit logs
  - Expected: Operator ID and username recorded

---

## 8. Cross-Browser Tests

**Test Cases:**

- [ ] **TC-BROWSER-001:** Chrome - All features work
  - Setup: Open in Chrome
  - Action: Test all features
  - Expected: No issues

- [ ] **TC-BROWSER-002:** Firefox - All features work
  - Setup: Open in Firefox
  - Action: Test all features
  - Expected: No issues

- [ ] **TC-BROWSER-003:** Safari - All features work
  - Setup: Open in Safari
  - Action: Test all features
  - Expected: No issues

- [ ] **TC-BROWSER-004:** Edge - All features work
  - Setup: Open in Edge
  - Action: Test all features
  - Expected: No issues

---

## 9. Edge Cases

**Test Cases:**

- [ ] **TC-EDGE-001:** Create provider with same username as existing
  - Setup: Provider A has admin "user1"
  - Action: Create Provider B with admin "user1"
  - Expected: Error or different tenant_id allowed

- [ ] **TC-EDGE-002:** Update credentials while admin logged in
  - Setup: Admin is logged into tenant interface
  - Action: Update their credentials
  - Expected: Session invalidated or remains valid

- [ ] **TC-EDGE-003:** Reset credentials multiple times rapidly
  - Setup: Click reset button 5 times quickly
  - Action: Monitor system
  - Expected: Rate limiting kicks in

- [ ] **TC-EDGE-004:** Very long password (1000 chars)
  - Setup: Enter 1000-character password
  - Action: Submit form
  - Expected: Accepted or reasonable max enforced

- [ ] **TC-EDGE-005:** Unicode characters in username
  - Setup: Enter username with emojis or accents
  - Action: Submit form
  - Expected: Validation error or sanitized

---

## 10. Regression Tests

**Test Cases:**

- [ ] **TC-REG-001:** Existing providers still work
  - Setup: Provider created before feature
  - Action: Access provider
  - Expected: No issues, can still manage

- [ ] **TC-REG-002:** Existing admin login still works
  - Setup: Old admin credentials
  - Action: Login
  - Expected: Successful login

- [ ] **TC-REG-003:** Provider creation without credentials still works
  - Setup: Create provider (old way)
  - Action: Don't specify credentials
  - Expected: Defaults used, works fine

- [ ] **TC-REG-004:** Other provider features unaffected
  - Setup: Use other provider features
  - Action: Test edit, delete, etc.
  - Expected: All work normally

---

## Test Execution Summary

### Manual Testing Required

- All frontend component tests (TC-UI-001 to TC-UI-025)
- All integration tests (TC-INT-001 to TC-INT-011)
- All cross-browser tests (TC-BROWSER-001 to TC-BROWSER-004)

### Automated Testing Recommended

- All backend API tests (TC-GET-001 to TC-POST-004)
- All security tests (TC-SEC-001 to TC-SEC-015)
- All database verification tests (TC-DB-001 to TC-DB-006)
- Performance tests (TC-PERF-001 to TC-PERF-005)

### Test Priority

**HIGH (Must Pass):**
- TC-GET-001, TC-PUT-001, TC-POST-001
- TC-UI-001, TC-UI-002, TC-UI-003, TC-UI-010
- TC-INT-001, TC-INT-002, TC-INT-003
- TC-SEC-001, TC-SEC-002, TC-SEC-005, TC-SEC-006
- TC-DB-001, TC-DB-006

**MEDIUM (Should Pass):**
- All validation tests
- All integration tests
- All audit log tests

**LOW (Nice to Have):**
- Performance tests
- Cross-browser tests
- Edge cases

---

## Sign-Off

**Tester:** _______________________ **Date:** _____________

**Test Environment:**
- Backend Version: _____________
- Frontend Version: _____________
- Database: ____________________
- Browser: _____________________

**Results:**
- Total Tests: _______
- Passed: _______
- Failed: _______
- Blocked: _______

**Overall Status:** [ ] PASS [ ] FAIL [ ] PARTIAL

**Notes:**
___________________________________________________________________________
___________________________________________________________________________
___________________________________________________________________________

---

**Last Updated:** 2026-03-24
**Test Suite Version:** 1.0
