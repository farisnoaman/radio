# Billing Profile (RadiusProfile) Not Found - Fix Guide

## Issue Description

**Error Message:** "Associated billing profile not found"

**When it happens:** When trying to create a RadiusUser, you need to select a "Billing Profile" (RadiusProfile), but the system can't find any profiles.

## Root Cause

**Same Bug Pattern:** RadiusProfile objects were being created **without setting `TenantID`**, causing:
1. Profiles created with `tenant_id = 0`
2. User creation queries: `WHERE tenant_id = ? AND id = ?`
3. Profiles with `tenant_id = 0` are filtered out
4. Result: No profiles found ❌

**Code Location:** [internal/adminapi/profiles.go:276-306](internal/adminapi/profiles.go#L276-L306)

## The Fix

### Changes Made

1. **CreateProfile** - Now sets TenantID from context
2. **ListProfiles** - Now uses `repository.TenantScope` for filtering
3. **Name uniqueness check** - Now checks within tenant scope

```diff
+ // Get tenant ID from context
+ tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

  profile := req.toRadiusProfile()
+ profile.TenantID = tenantID // Set tenant from context

- GetDB(c).Model(&domain.RadiusProfile{}).Where("name = ?", profile.Name).Count(&count)
+ GetDB(c).Model(&domain.RadiusProfile{}).Where("tenant_id = ? AND name = ?", tenantID, profile.Name).Count(&count)

  query := db.Model(&domain.RadiusProfile{})
+ query := db.Model(&domain.RadiusProfile{}).Scopes(repository.TenantScope)
```

## Solution for Existing Data

If you have existing RadiusProfiles with `tenant_id = 0`, you need to fix them:

### Option 1: SQL Update (Quick Fix)

```sql
-- First, check what profiles exist with tenant_id = 0
SELECT id, name, tenant_id FROM radius_profile WHERE tenant_id = 0;

-- Update all profiles to belong to a specific tenant (replace YOUR_TENANT_ID)
-- WARNING: This assigns ALL profiles to one tenant - only do this if you have one tenant
UPDATE radius_profile
SET tenant_id = YOUR_TENANT_ID
WHERE tenant_id = 0;

-- Verify the update
SELECT id, name, tenant_id FROM radius_profile WHERE tenant_id = YOUR_TENANT_ID;
```

### Option 2: Manual Re-creation (Clean)

If you have multiple tenants or want to do this properly:

1. **Login as tenant admin**
2. **Go to Profiles page** (in the UI)
3. **Create new profiles** for each tenant
4. **Delete old profiles** with `tenant_id = 0`

```sql
-- List profiles with tenant_id = 0
SELECT id, name FROM radius_profile WHERE tenant_id = 0;

-- Delete old profiles (BE CAREFUL - make sure you created replacements first)
DELETE FROM radius_profile WHERE tenant_id = 0;
```

### Option 3: Tenant-Specific Migration (Best for Multi-Tenant)

If you have multiple tenants and need to assign profiles correctly:

```sql
-- Step 1: Check your tenants
SELECT id, name FROM tenant;

-- Step 2: For each tenant, update their profiles
-- Replace TENANT_ID with actual tenant ID from step 1
-- You might need to check which products belong to which tenant first

-- View products to understand tenant-product relationships
SELECT id, name, tenant_id FROM product;

-- Update profiles based on product relationships
-- This assigns profiles to the same tenant as their associated products
UPDATE radius_profile rp
SET tenant_id = (
    SELECT DISTINCT p.tenant_id
    FROM product p
    WHERE p.radius_profile_id = rp.id
    LIMIT 1
)
WHERE rp.tenant_id = 0;

-- Verify
SELECT id, name, tenant_id FROM radius_profile ORDER BY tenant_id, id;
```

## How to Create a Billing Profile (RadiusProfile)

### Via UI (Recommended)

1. **Login as Tenant Admin** or Super Admin
2. **Navigate to:** Profiles → Radius Profiles
3. **Click "Create Profile"** button
4. **Fill in the form:**
   - **Name:** e.g., "Default Profile", "Premium Plan"
   - **Status:** Enabled
   - **Address Pool:** Your IP pool name (optional)
   - **Active Sessions:** Max concurrent sessions (e.g., 1)
   - **Upload Rate:** Upload speed in Kbps (e.g., 1024)
   - **Download Rate:** Download speed in Kbps (e.g., 1024)
   - **Data Quota:** Data limit in MB (0 = unlimited)
   - **Remark:** Description (optional)
5. **Click "Save"**

### Via API

```bash
# Create a new profile
curl -X POST http://your-server/api/v1/radius-profiles \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Default Profile",
    "status": "enabled",
    "active_num": 1,
    "up_rate": 1024,
    "down_rate": 1024,
    "data_quota": 0,
    "addr_pool": "pool1"
  }'
```

## Verify the Fix

### 1. Check Profiles Exist for Your Tenant

```sql
-- Replace YOUR_TENANT_ID with your actual tenant ID
SELECT id, name, status, tenant_id
FROM radius_profile
WHERE tenant_id = YOUR_TENANT_ID;
```

**Expected:** Should see at least one profile with your tenant_id

### 2. Try Creating a User

1. Go to Users page
2. Click "Create User"
3. **Billing Profile** dropdown should now show profiles
4. Fill in user details and select a profile
5. Click "Save"

**Expected:** User is created successfully without the "Associated billing profile not found" error

### 3. Verify User Creation

```sql
-- Check that user was created with correct profile
SELECT u.id, u.username, u.profile_id, u.tenant_id, p.name as profile_name
FROM radius_user u
LEFT JOIN radius_profile p ON u.profile_id = p.id
WHERE u.tenant_id = YOUR_TENANT_ID
ORDER BY u.id DESC
LIMIT 5;
```

## Common Issues & Solutions

### Issue 1: "No profiles in dropdown"

**Cause:** No profiles exist for your tenant (tenant_id)

**Solution:**
```sql
-- Create a default profile for your tenant
INSERT INTO radius_profile (name, status, tenant_id, active_num, up_rate, down_rate, data_quota, created_at)
VALUES ('Default Profile', 'enabled', YOUR_TENANT_ID, 1, 1024, 1024, 0, datetime('now'));
```

### Issue 2: "Profile name already exists" (but you don't see it)

**Cause:** Old profile with `tenant_id = 0` exists with the same name

**Solution:**
```sql
-- Delete or update old profiles
-- Option A: Delete
DELETE FROM radius_profile WHERE tenant_id = 0;

-- Option B: Update to your tenant
UPDATE radius_profile SET tenant_id = YOUR_TENANT_ID WHERE tenant_id = 0;
```

### Issue 3: Can see profiles but they're wrong (from other tenants)

**Cause:** ListProfiles wasn't filtering by tenant (fixed in our update)

**Solution:** Restart backend to load the updated code with TenantScope filtering

## Testing Checklist

After applying the fix:

- [ ] Backend restarted with updated code
- [ ] Old profiles updated to correct tenant_id (SQL run)
- [ ] Can see profiles in Profiles page
- [ ] Can create new profile via UI
- [ ] Billing Profile dropdown shows profiles when creating user
- [ ] Can create user successfully
- [ ] User can login with created credentials
- [ ] Database query confirms correct tenant_id on all records

## Related Fixes

This is part of a broader fix for missing TenantID across the system:
- ✅ VoucherBatch creation
- ✅ Voucher creation
- ✅ RadiusUser creation (batch activation)
- ✅ RadiusUser creation (individual redemption)
- ✅ VoucherBundle creation
- ✅ VoucherTopup creation
- ✅ VoucherSubscription creation
- ✅ **RadiusProfile creation** (this fix)

## Prevention

To prevent this in the future:

1. **Database Constraint:**
```sql
ALTER TABLE radius_profile
ADD CONSTRAINT chk_tenant_id_not_zero CHECK (tenant_id > 0);
```

2. **Code Review:** Always check that TenantID is set when creating domain objects

3. **Testing:** Add integration tests that verify tenant_id is set correctly

---

**Fixed By:** Claude Code (Systematic Debugging)
**Date:** 2026-03-24
**Files Modified:** `internal/adminapi/profiles.go`
**Status:** Ready for Deployment
