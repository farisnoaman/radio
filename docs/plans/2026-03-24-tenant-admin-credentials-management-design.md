# Tenant Admin Credentials Management - Design Document

**Date:** 2026-03-24
**Status:** Approved
**Author:** Claude Code (with user validation)

## Overview

Provide platform administrators with full control over tenant (provider) admin credentials, eliminating the current limitation where tenant admin passwords are randomly generated and can easily be lost.

## Problem Statement

**Current Issues:**
- Tenant admin passwords are randomly generated during provider approval
- Passwords shown only once in API response → easily lost
- No way to reset tenant admin passwords from platform interface
- Platform admins must access database directly to recover credentials
- No control over tenant admin credentials during provider creation

**User Impact:**
- Lost credentials require database access to recover
- No audit trail for credential changes
- Security risk with unknown passwords
- Poor user experience for platform administrators

## Solution

Add comprehensive tenant admin credentials management to the platform admin interface with:
1. **Set custom credentials** during provider creation (all paths)
2. **Reset credentials** anytime via provider detail page
3. **Default credentials** fallback (`admin` / `123456`)
4. **Full audit logging** of all credential changes

## Goals

✅ Set custom username/password during provider creation
✅ Reset tenant admin credentials via platform UI
✅ Default fallback credentials when not specified
✅ Standard validation on all inputs
✅ Complete audit trail of credential changes
✅ Masked password display (show only during creation/reset)

## Architecture

### Backend API

**New Endpoint:** `/api/v1/platform/providers/:id/admin-credentials`

**Operations:**
- **GET** - Retrieve current tenant admin credentials (password masked)
- **PUT** - Update tenant admin credentials
- **POST** - Reset to default credentials (`admin` / `123456`)

**Implementation File:** `internal/adminapi/provider_admin_credentials.go`

**Functions:**
- `GetProviderAdminCredentials(c echo.Context) error`
- `UpdateProviderAdminCredentials(c echo.Context) error`
- `ResetProviderAdminCredentials(c echo.Context) error`

**Validation:**
- Username: 3-50 characters, alphanumeric only
- Password: Minimum 6 characters, any characters allowed
- Both required when setting custom credentials

### Frontend Components

**New Components:**
1. **AdminCredentialsSection** - Form section for provider creation/edit
2. **AdminCredentialsCard** - Display card for provider detail page
3. **ResetPasswordDialog** - Modal dialog for password reset

**Placement:**
- **Provider Form:** Collapsible "Admin Credentials" section
- **Provider Detail Page:** Dedicated card with reset button

### Data Flow

**Provider Creation:**
```
User fills form → Backend creates provider → Creates/updates tenant admin → Returns credentials → Frontend displays success
```

**Password Reset:**
```
Click "Reset" → Modal opens → Enter new credentials → API call → Updates database → Logs change → Returns success
```

### Security Features

**Password Handling:**
- Passwords masked in GET responses (`********`)
- Full password only returned in POST/PUT (creation/update moments)
- Never log passwords in plain text
- Clipboard button for easy copying

**Access Control:**
- Only platform admins (`level='super'`) can access
- Tenant admins blocked from this endpoint
- Operators blocked from this endpoint

**Audit Logging:**
- Log all credential changes with:
  - Operator who made change
  - Timestamp
  - IP address
  - Provider affected
  - Old vs new username (password not logged)

**Rate Limiting:**
- 3 password resets per hour per IP
- 5 failed login attempts triggers 15-minute lockout

**Enhanced Security:**
- Password strength indicator (weak/medium/strong)
- Common password blacklist (reject "password", "123456")
- Force password change on first login if using defaults
- Password expiration reminder (90 days)
- Invalidate all sessions after password reset

### Error Handling

| Scenario | Error Code | Message |
|----------|------------|---------|
| Provider not found | `PROVIDER_NOT_FOUND` | "Provider does not exist" |
| Invalid username | `INVALID_USERNAME` | "Username must be 3-50 alphanumeric characters" |
| Invalid password | `INVALID_PASSWORD` | "Password must be at least 6 characters" |
| Passwords don't match | `PASSWORD_MISMATCH` | "Passwords do not match" |
| Username conflict | `USERNAME_EXISTS` | "An operator with this username already exists" |
| Not authorized | `FORBIDDEN` | "Only platform admins can manage provider admin credentials" |

### Default Credentials

When credentials are not provided:
- **Username:** `admin`
- **Password:** `123456`

**Behavior:**
- Auto-creates tenant admin if doesn't exist
- Updates existing admin if already exists
- Flags account for password change on first login

## Implementation Phases

### Phase 1: Backend API (HIGH)
- Create provider admin credentials API
- Implement CRUD operations
- Add validation and error handling
- Add audit logging

### Phase 2: Provider Integration (HIGH)
- Integrate with provider registration flow
- Integrate with provider approval flow
- Integrate with direct provider creation
- Add default credential fallback

### Phase 3: Frontend Components (MEDIUM)
- Create AdminCredentialsSection component
- Create AdminCredentialsCard component
- Create ResetPasswordDialog modal
- Add validation and error handling

### Phase 4: UI Integration (MEDIUM)
- Add to provider creation/edit forms
- Add to provider detail page
- Wire up API calls
- Add loading states and toasts

### Phase 5: Testing & Polish (LOW)
- Unit tests for API endpoints
- Integration tests for full flows
- Manual testing checklist
- Add password strength indicator
- Security testing

## Migration Strategy

**No Database Migration Required**
- Uses existing `sys_opr` table
- Links via existing `tenant_id` field
- Backward compatible with existing providers

**Rollout Plan:**
1. Deploy backend changes
2. Deploy frontend changes
3. Existing providers continue working
4. New providers get credential management
5. Manual updates available for existing providers

## Success Criteria

- [x] Platform admins can set credentials during provider creation
- [x] Platform admins can reset credentials via UI
- [x] Default credentials (admin/123456) work when not specified
- [x] Passwords are masked in GET responses
- [x] Full audit trail of credential changes
- [x] Validation prevents invalid inputs
- [x] Rate limiting prevents abuse
- [x] All three provider creation paths support custom credentials

**Status:** ✅ COMPLETE - All success criteria met (2026-03-24)

## Open Questions

- Should we implement password hashing in future? (Deferred - requires migration)
- Should we email credentials to provider contact? (Deferred - needs email service)
- Should we add 2FA for platform admin? (Out of scope)

## Related Documents

- Multi-Provider Architecture: `docs/plans/2026-03-19-multi-provider-saas-design.md`
- Provider Management: `docs/plans/2026-03-20-phase2-provider-management.md`
- Admin Credentials Guide: `docs/guides/admin-credentials-guide.md`
