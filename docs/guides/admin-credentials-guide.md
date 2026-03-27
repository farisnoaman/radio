# Admin Credentials Guide - ToughRADIUS

## Quick Reference

| Admin Type | Default Username | Default Password | Access Level |
|------------|------------------|------------------|--------------|
| **Super Admin** | `admin` | `toughradius` | Platform-wide access |
| **Provider Admin** | `admin` | Randomly generated | Per-provider access |

---

## 1. Super Admin Credentials

### Default Credentials

**Username:** `admin`
**Password:** `toughradius`
**Access URL:** http://localhost:1816

These are automatically created when the system starts (see `internal/app/initdb.go`).

### How to Reset Super Admin Password

#### Method 1: Using the Reset Script (Recommended)

```bash
# Navigate to project root
cd /home/faris/Documents/lamees/radio

# Reset to default password
./scripts/reset-admin-password.sh

# Reset to custom password
./scripts/reset-admin-password.sh "your_new_password"
```

#### Method 2: Using Command-Line Tool

```bash
# Build and run the reset tool
cd /home/faris/Documents/lamees/radio
cd cmd/reset-password
go build -o ../../reset-password .
cd ../..
./reset-password -c toughradius.yml -u admin -p "new_password"

# Clean up
rm -f reset-password
```

#### Method 3: Direct SQL Update (Emergency Only)

```sql
-- Update password (requires hashing)
-- This is NOT recommended unless you understand the security implications
UPDATE sys_opr
SET password = 'toughradius'  -- This will be hashed on next login
WHERE username = 'admin';
```

### Verify Super Admin Login

```bash
# Check if super admin exists
sqlite3 data/radius.db "SELECT id, username, level, status FROM sys_opr WHERE level = 'super';"

# Expected output:
# id|username|level|status|
# 1|admin|super|enabled|
```

---

## 2. Provider Admin Credentials

Provider admins are **automatically created** when a new provider is approved.

### What Happens During Provider Creation

When you approve a provider registration (see `internal/adminapi/provider_registration.go:183-194`):

1. Provider account is created
2. **Default admin operator is automatically created** with:
   - **Username:** `admin`
   - **Password:** Randomly generated (8 characters)
   - **Level:** `admin`
   - **Status:** `enabled`
3. Credentials are **returned in the API response**

### How to Get Provider Admin Password

#### Step 1: Check the Approval Response

When you approve a provider registration, the response includes the admin credentials:

```json
{
  "provider": {
    "id": 123,
    "name": "Test ISP LLC",
    "code": "test-isp"
  },
  "admin": {
    "username": "admin",
    "password": "random8chars",  // ← This is the only time you'll see this!
    "level": "admin",
    "status": "enabled"
  }
}
```

**⚠️ IMPORTANT:** Save the password immediately! It's only shown once during creation.

#### Step 2: Check Database (If Lost)

If you didn't save the password, you can find it in the database:

```sql
-- Find provider admin (replace PROVIDER_ID with actual provider ID)
SELECT id, username, password, level, tenant_id
FROM sys_opr
WHERE tenant_id = PROVIDER_ID AND level = 'admin';

-- Example: If provider ID is 2
SELECT id, username, password, level, tenant_id
FROM sys_opr
WHERE tenant_id = 2 AND level = 'admin';
```

The password will be in plain text (you can copy it and login).

#### Step 3: Reset Provider Admin Password

If you need to reset a provider admin password:

```bash
# Use the reset script with provider context
# (This requires knowing the provider ID)

# Option 1: Direct SQL update to known password
UPDATE sys_opr
SET password = 'new_password_here'
WHERE tenant_id = PROVIDER_ID AND level = 'admin';

# Option 2: Use command-line tool (specify provider admin username)
./reset-password -c toughradius.yml -u admin -p "new_password"
```

### Login as Provider Admin

1. **Access URL:** http://localhost:1816
2. **Username:** `admin`
3. **Password:** [Use password from approval response or database]

**Note:** Provider admins can only access resources within their tenant (provider).

---

## 3. Creating Additional Admin Users

### Via Web UI

1. **Login as Super Admin** or **Provider Admin**
2. **Navigate to:** Operators → Create Operator
3. **Fill in the form:**
   - **Username:** (e.g., "john_admin")
   - **Password:** (e.g., "secure_password123")
   - **Real Name:** (e.g., "John Doe")
   - **Email:** (e.g., "john@example.com")
   - **Mobile:** (e.g., "+1234567890")
   - **Level:** Select from dropdown:
     - `super` - Platform admin (all providers)
     - `admin` - Provider admin (one provider)
     - `operator` - Limited access
     - `agent` - Agent access
   - **Status:** Enabled
4. **Click "Save"**

### Via API

```bash
# Create new operator
curl -X POST http://localhost:1816/api/v1/operators \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_admin",
    "password": "secure_password123",
    "realname": "John Doe",
    "email": "john@example.com",
    "mobile": "+1234567890",
    "level": "admin",
    "status": "enabled"
  }'
```

### Via SQL (Not Recommended)

```sql
-- Create new operator
INSERT INTO sys_opr (
  id,
  username,
  password,
  realname,
  email,
  mobile,
  level,
  status,
  tenant_id,
  created_at
)
VALUES (
  lower(hex(randomblob(16))),  -- Generate ID
  'john_admin',
  'secure_password123',
  'John Doe',
  'john@example.com',
  '+1234567890',
  'admin',
  'enabled',
  1,  -- Replace with your tenant_id
  datetime('now')
);
```

---

## 4. Common Issues & Solutions

### Issue 1: "Cannot login with admin/toughradius"

**Possible Causes:**
1. Backend not started
2. Database not initialized
3. Password was changed

**Solution:**
```bash
# 1. Check if backend is running
curl http://localhost:1816/api/v1/system/status

# 2. Check database has admin account
sqlite3 data/radius.db "SELECT username, level, status FROM sys_opr WHERE username = 'admin';"

# 3. Reset password
./scripts/reset-admin-password.sh
```

### Issue 2: "Provider admin password lost"

**Solution:**
```sql
-- Find the provider ID first
SELECT id, name, code FROM provider;

-- Then find admin for that provider
SELECT username, password FROM sys_opr
WHERE tenant_id = PROVIDER_ID AND level = 'admin';

-- Copy the password and login
```

### Issue 3: "Need to change provider admin password"

**Solution:**
```sql
-- Update to new password
UPDATE sys_opr
SET password = 'new_secure_password'
WHERE tenant_id = PROVIDER_ID AND level = 'admin';
```

### Issue 4: "Creating new provider - what are the admin credentials?"

**Answer:** Provider admin credentials are **automatically generated** when you approve the provider registration. The password is random and shown **only once** in the approval response.

**To get credentials:**
1. Check the API response when approving registration
2. Or query database: `SELECT * FROM sys_opr WHERE tenant_id = PROVIDER_ID;`

---

## Managing Provider Admin Credentials (NEW)

### Via Platform UI

Platform administrators can now manage provider admin credentials directly:

1. **During Provider Creation:**
   - Navigate to Platform → Providers → Create Provider
   - Expand "Admin Credentials" section
   - Choose custom credentials or use defaults
   - Save the provider

2. **On Provider Detail Page:**
   - Navigate to Platform → Providers → Click on provider
   - View "Admin Credentials" card
   - Click "Reset to Defaults" for admin/123456
   - Click "Update Credentials" for custom credentials
   - Use copy buttons to save credentials

3. **During Registration Approval:**
   - Navigate to Platform → Provider Registrations
   - Click "Approve" on pending registration
   - Set custom credentials in approval form
   - Submit to create provider with admin

### Via API

**Get Provider Admin Credentials:**
```bash
curl -X GET http://localhost:1816/api/v1/platform/providers/1/admin-credentials \
  -H "Authorization: Bearer PLATFORM_ADMIN_TOKEN"

# Response:
{
  "username": "admin",
  "password": "********",  # Masked for security
  "level": "admin",
  "status": "enabled"
}
```

**Update Provider Admin Credentials:**
```bash
curl -X PUT http://localhost:1816/api/v1/platform/providers/1/admin-credentials \
  -H "Authorization: Bearer PLATFORM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newadmin",
    "password": "newpass123"
  }'

# Response:
{
  "username": "newadmin",
  "password": "newpass123",  # Only shown once on update
  "level": "admin",
  "status": "enabled"
}
```

**Reset to Default Credentials:**
```bash
curl -X POST http://localhost:1816/api/v1/platform/providers/1/admin-credentials/reset \
  -H "Authorization: Bearer PLATFORM_ADMIN_TOKEN"

# Response:
{
  "username": "admin",
  "password": "123456",  # Default credentials
  "level": "admin",
  "status": "enabled"
}
```

### Important Notes

- **Default Credentials:** `admin` / `123456`
- **Password Security:** Passwords are masked in GET responses (`********`)
- **One-Time Display:** Full passwords only shown during creation/update
- **Access Control:** Only platform admins (`super` level) can manage tenant admin credentials
- **Audit Logging:** All credential changes are logged with operator, timestamp, and IP
- **Validation Requirements:**
  - Username: 3-50 alphanumeric characters
  - Password: Minimum 6 characters

---

## 5. Security Best Practices

### Immediate Actions After Setup

1. ✅ **Change super admin password**
   ```bash
   ./scripts/reset-admin-password.sh "your_secure_password_here"
   ```

2. ✅ **Save provider admin passwords** securely
   - Use password manager (e.g., LastPass, 1Password)
   - Store in encrypted vault
   - Never share via email/chat

3. ✅ **Create separate admin accounts** for different people
   - Don't share super admin account
   - Create individual operator accounts
   - Use appropriate access levels

### Password Requirements

Current system does **NOT** enforce password complexity (you should add this):

- **Minimum length:** 8 characters (recommended)
- **Include:** Uppercase, lowercase, numbers, symbols
- **Avoid:** Dictionary words, common patterns

### Access Level Guidelines

| Level | Use Case | Permissions |
|-------|----------|-------------|
| `super` | Platform owner | Full access to all providers, system settings |
| `admin` | Provider owner | Full access within their provider only |
| `operator` | Support staff | Limited access (view/edit users, view reports) |
| `agent` | Sales agents | Vouchers, commissions only |

---

## 6. Troubleshooting Commands

### Check All Admins

```sql
-- List all operators with their levels
SELECT
  id,
  username,
  realname,
  level,
  status,
  tenant_id,
  created_at
FROM sys_opr
ORDER BY level, tenant_id;
```

### Check Provider Admins

```sql
-- List all provider admins
SELECT
  o.id,
  o.username,
  o.realname,
  o.email,
  p.name as provider_name,
  p.code as provider_code
FROM sys_opr o
LEFT JOIN provider p ON o.tenant_id = p.id
WHERE o.level = 'admin'
ORDER BY p.name;
```

### Count Admins by Type

```sql
-- Count admins by level
SELECT
  level,
  COUNT(*) as count
FROM sys_opr
WHERE status = 'enabled'
GROUP BY level;
```

### Find All Admins for Specific Provider

```sql
-- Replace PROVIDER_ID with actual provider ID
SELECT
  username,
  realname,
  email,
  mobile,
  level,
  status
FROM sys_opr
WHERE tenant_id = PROVIDER_ID
ORDER BY level, username;
```

---

## 7. API Reference

### Login Endpoint

```bash
# Login and get token
curl -X POST http://localhost:1816/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "toughradius"
  }'

# Response:
# {
#   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "user": {
#     "id": 1,
#     "username": "admin",
#     "level": "super",
#     ...
#   }
# }
```

### Get Current User Info

```bash
# Get current operator info
curl http://localhost:1816/api/v1/operators/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 8. Summary

### Super Admin (Platform Admin)
- **Username:** `admin`
- **Default Password:** `toughradius`
- **Reset:** `./scripts/reset-admin-password.sh`
- **Access:** All providers, system settings

### Provider Admin (Tenant Admin)
- **Username:** `admin` (per provider)
- **Password:** Randomly generated during provider creation
- **View in:** Database (`sys_opr` table) or approval response
- **Access:** Single provider only

### Additional Admins
- **Create:** Via UI (Operators page) or API
- **Levels:** `super`, `admin`, `operator`, `agent`
- **Recommendation:** Create individual accounts for each person

---

## Quick Start Checklist

- [ ] Login with default super admin credentials
- [ ] Change super admin password immediately
- [ ] Save new password securely
- [ ] Create provider(s) as needed
- [ ] Save provider admin credentials from approval response
- [ ] Create individual operator accounts for team members
- [ ] Set appropriate access levels for each operator

---

**Need Help?** Check backend logs:
```bash
tail -f backend.log | grep -i "admin\|operator\|login"
```

**Last Updated:** 2026-03-24
**Version:** v9.0.0
