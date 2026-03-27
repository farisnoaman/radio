# Comprehensive Tenant Isolation Audit Report

**Date:** 2026-03-24
**Audit Scope:** All admin API endpoints with tenant-scoped data
**Severity:** CRITICAL - Multiple data leakage vulnerabilities found

## Executive Summary

Found **CRITICAL SECURITY BREACHES** where tenant admins can view data from other tenants:
- Dashboard statistics (FIXED)
- Online sessions (CRITICAL)
- Accounting records (CRITICAL)
- Products (CRITICAL)
- Invoices (CRITICAL)

---

## Tables with TenantID Column (Must Enforce Isolation)

### Core RADIUS Tables
| Table | Has TenantID | Isolation Status | Notes |
|-------|-------------|------------------|-------|
| radius_profile | ✅ Yes | ✅ SECURE | ListProfiles uses TenantScope |
| radius_user | ✅ Yes | ✅ SECURE | listRadiusUsers uses TenantScope |
| radius_online | ✅ Yes | ❌ **LEAKING** | ListOnlineSessions NO tenant filter |
| radius_accounting | ✅ Yes | ❌ **LEAKING** | ListAccounting NO tenant filter |

### Network Tables
| Table | Has TenantID | Isolation Status | Notes |
|-------|-------------|------------------|-------|
| net_node | ✅ Yes | ⚠️ PARTIAL | Has manual tenant_id checks in WHERE |
| net_nas | ✅ Yes | ⚠️ PARTIAL | Has manual tenant_id checks in WHERE |

### Product Tables
| Table | Has TenantID | Isolation Status | Notes |
|-------|-------------|------------------|-------|
| product | ✅ Yes | ❌ **LEAKING** | ListProducts NO tenant filter |

### Voucher Tables
| Table | Has TenantID | Isolation Status | Notes |
|-------|-------------|------------------|-------|
| voucher_batch | ✅ Yes | ✅ SECURE | Manual tenant_id filtering |
| voucher | ✅ Yes | ✅ SECURE | Manual tenant_id filtering |

### Billing Tables
| Table | Has TenantID | Isolation Status | Notes |
|-------|-------------|------------------|-------|
| invoice | ❌ No | ⚠️ INDIRECT | Links to radius_user.username |
| billing_plan | ✅ Yes | ❓ UNKNOWN | Not audited yet |

### System Tables
| Table | Has TenantID | Isolation Status | Notes |
|-------|-------------|------------------|-------|
| sys_opr | ✅ Yes | ⚠️ CONTEXT | Role-based access needed |
| sys_config | ❌ No | N/A | Global configuration |

---

## Critical Security Issues

### 1. ❌ sessions.go - ListOnlineSessions (CRITICAL)
**File:** [internal/adminapi/sessions.go:74](sessions.go#L74)
**Severity:** CRITICAL - Data leakage
**Impact:** Tenant admins can see ALL online sessions from ALL tenants

**Code:**
```go
query := db.Model(&domain.RadiusOnline{})  // NO TENANT FILTER
```

**Fix Required:**
```go
query := db.Model(&domain.RadiusOnline{}).Scopes(repository.TenantScope)
```

---

### 2. ❌ accounting.go - ListAccounting (CRITICAL)
**File:** [internal/adminapi/accounting.go:31](accounting.go#L31)
**Severity:** CRITICAL - Data leakage
**Impact:** Tenant admins can see ALL accounting records from ALL tenants

**Code:**
```go
query := db.Model(&domain.RadiusAccounting{})  // NO TENANT FILTER
```

**Fix Required:**
```go
query := db.Model(&domain.RadiusAccounting{}).Scopes(repository.TenantScope)
```

---

### 3. ❌ products.go - ListProducts (CRITICAL)
**File:** [internal/adminapi/products.go:23](products.go#L23)
**Severity:** CRITICAL - Data leakage
**Impact:** Tenant admins can see ALL products from ALL tenants

**Code:**
```go
query := db.Model(&domain.Product{})  // NO TENANT FILTER
```

**Fix Required:**
```go
query := db.Model(&domain.Product{}).Scopes(repository.TenantScope)
```

---

### 4. ❌ invoices.go - ListInvoices (CRITICAL)
**File:** [internal/adminapi/invoices.go:28](invoices.go#L28)
**Severity:** CRITICAL - Data leakage
**Impact:** Admins can see ALL invoices from ALL tenants

**Code:**
```go
query := db.Model(&domain.Invoice{})  // NO TENANT FILTER
// Only filters by username for "user" role
// Admins see ALL invoices regardless of tenant
```

**Fix Required:**
Invoices need indirect tenant filtering through radius_user join:
```go
// For admin role: filter by tenant through radius_user
query = query.Joins("JOIN radius_user ON radius_user.username = invoice.username").
    Where("radius_user.tenant_id = ?", tenantID)
```

---

## Previously Fixed Issues ✅

### dashboard.go - GetDashboardStats (FIXED)
**Commit:** ba162108
**Fix:** Added `.Scopes(repository.TenantScope)` to main query

### profiles.go - UpdateProfile (FIXED)
**Commit:** 20d62451
**Fix:** Added tenant_id to uniqueness check

### nodes.go - updateNode (FIXED)
**Commit:** b521c5dd
**Fix:** Added tenant_id to name uniqueness check

### nas.go - UpdateNAS (FIXED)
**Commit:** b521c5dd
**Fix:** Added tenant_id to IP uniqueness check

---

## Manual Tenant ID Checks (Acceptable)

These endpoints use manual `tenant_id` in WHERE clauses instead of TenantScope:

### ✅ vouchers.go - ListVoucherBatches
```go
query := db.Model(&domain.VoucherBatch{}).Where("tenant_id = ?", tenantID)
```

### ✅ users.go - listRadiusUsers
Uses both TenantScope AND manual tenant_id in joins:
```go
base := db.Model(&domain.RadiusUser{}).
    Joins("LEFT JOIN (SELECT username, COUNT(1) AS count FROM radius_online WHERE tenant_id = ? GROUP BY username) ro...", tenantID).
    Scopes(repository.TenantScope)
```

### ✅ profiles.go - ListProfiles
```go
query := db.Model(&domain.RadiusProfile{}).Scopes(repository.TenantScope)
```

---

## Recommendations

### Immediate Actions (Priority 1 - CRITICAL)
1. **Fix sessions.go** - Add TenantScope to ListOnlineSessions
2. **Fix accounting.go** - Add TenantScope to ListAccounting
3. **Fix products.go** - Add TenantScope to ListProducts
4. **Fix invoices.go** - Add tenant filtering through radius_user join

### Code Review (Priority 2 - HIGH)
5. Audit ALL remaining admin API endpoints for missing tenant isolation
6. Check non-admin API endpoints (portal, public APIs)
7. Add automated tests for tenant isolation

### Architecture (Priority 3 - MEDIUM)
8. Consider adding middleware that automatically applies TenantScope to all queries
9. Add database-level row-level security (RLS) policies
10. Implement integration tests that create multiple tenants and verify data isolation

---

## Testing Checklist

After fixes, verify:
- [ ] Create tenant A and tenant B
- [ ] Create users in both tenants
- [ ] Create sessions in both tenants
- [ ] Login as tenant A admin
- [ ] Verify tenant A cannot see tenant B's:
  - [ ] Online sessions
  - [ ] Accounting records
  - [ ] Products
  - [ ] Invoices
  - [ ] Users
  - [ ] Profiles
  - [ ] NAS devices
  - [ ] Network nodes

---

## Summary Statistics

| Metric | Count |
|--------|-------|
| Tables with TenantID | 15+ |
| Endpoints audited | 30+ |
| Critical leaks found | 4 |
| Already secure | 6 |
| Partially secure | 2 |
| Fixed in this session | 4 |
| Remaining to fix | 4 |

---

## Generated By
Claude Code with systematic-debugging skill
Audit date: 2026-03-24
