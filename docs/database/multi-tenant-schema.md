# Multi-Tenant Database Architecture

**Version:** 1.0
**Last Updated:** 2026-03-20
**Phase:** 1 - Database Schema & Migration

---

## Overview

This document describes the multi-tenant database architecture for the ToughRadius SaaS platform. The system supports 100+ providers (ISPs/WISPs) with 5000 users per provider (500K+ total users) while maintaining strong data isolation and optimal query performance.

---

## Architecture Principles

### Schema-Per-Provider Isolation

Each provider gets their own PostgreSQL schema (`provider_1`, `provider_2`, etc.) for:

- **Strong Data Isolation:** Provider data is physically separated at the schema level
- **Easy Backup/Restore:** Can backup or migrate individual providers
- **Scalability Path:** Large providers can be migrated to dedicated databases
- **Performance:** Queries are naturally scoped to a single schema
- **Security:** SQL injection in one provider cannot affect others

### Platform Schema

The `public` schema holds platform-wide data:

- **Provider Registry:** Metadata about all providers
- **Registration System:** Provider signup and approval workflow
- **System Configuration:** Platform-wide settings
- **System Operators:** Platform administrators

### Tenant ID Layering

All tables in provider schemas include a `tenant_id` column for:

- **Double Isolation:** Schema + tenant_id provides defense in depth
- **Future Flexibility:** Can aggregate queries across providers if needed
- **Indexing Strategy:** Composite indexes on `(tenant_id, field)` optimize queries
- **Migration Path:** Makes it easier to move data between schemas

---

## Schema Structure

### Platform Schema (`public`)

#### Provider Tables

**mst_provider**
```sql
CREATE TABLE mst_provider (
    id BIGINT PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    max_users INT DEFAULT 1000,
    max_nas INT DEFAULT 100,
    branding TEXT,              -- JSON: ProviderBranding
    settings TEXT,              -- JSON: ProviderSettings
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_provider_code ON mst_provider(code);
CREATE INDEX idx_provider_status ON mst_provider(status);
```

**mst_provider_registration**
```sql
CREATE TABLE mst_provider_registration (
    id BIGINT PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    address TEXT,
    business_type VARCHAR(100),
    expected_users INT,
    expected_nas INT,
    country VARCHAR(100),
    message TEXT,
    status VARCHAR(20) DEFAULT 'pending',  -- pending, approved, rejected
    reviewed_by BIGINT,
    reviewed_at TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_registration_status ON mst_provider_registration(status);
CREATE INDEX idx_registration_email ON mst_provider_registration(email);
```

#### System Tables

**sys_config**
- Platform-wide configuration (SMTP, RADIUS settings, etc.)

**sys_opr**
- System operators (both platform admins and provider admins)

**sys_opr_log**
- Audit log of administrative actions

---

### Provider Schemas (`provider_N`)

Each provider schema contains identical table structures:

#### User Management

**radius_user**
```sql
CREATE TABLE radius_user (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,                    -- Provider ID
    node_id BIGINT,
    profile_id BIGINT,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(255),
    realname VARCHAR(100),
    email VARCHAR(255),
    mobile VARCHAR(20),
    address VARCHAR(255),
    vlanid1 INT,
    vlanid2 INT,
    ip_addr INET,
    ipv6_addr INET,
    mac_addr MACADDR,
    bind_vlan INT DEFAULT 0,
    bind_mac INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'enabled',
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Indexes for tenant-scoped queries
CREATE INDEX idx_radius_user_tenant_status ON radius_user(tenant_id, status);
CREATE INDEX idx_radius_user_tenant_username ON radius_user(tenant_id, username);
CREATE INDEX idx_radius_user_mac_addr ON radius_user(mac_addr) WHERE mac_addr IS NOT NULL;
```

**radius_profile**
```sql
CREATE TABLE radius_profile (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'enabled',
    up_rate INT,
    down_rate INT,
    data_quota BIGINT,
    addr_pool VARCHAR(50),
    -- ... additional profile fields
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_radius_profile_tenant_status ON radius_profile(tenant_id, status);
```

#### Session Management

**radius_online**
```sql
CREATE TABLE radius_online (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    nas_id BIGINT,
    username VARCHAR(100),
    acct_session_id VARCHAR(100) UNIQUE,
    framed_ip_address INET,
    session_time BIGINT,
    input_octets BIGINT,
    output_octets BIGINT,
    -- ... additional session fields
    created_at TIMESTAMP
);

CREATE INDEX idx_radius_online_tenant ON radius_online(tenant_id);
CREATE INDEX idx_radius_online_session ON radius_online(acct_session_id);
CREATE INDEX idx_radius_online_username ON radius_online(username);
```

**radius_accounting**
```sql
CREATE TABLE radius_accounting (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    acct_session_id VARCHAR(100) UNIQUE,
    username VARCHAR(100),
    nas_id BIGINT,
    acct_start_time TIMESTAMP,
    acct_stop_time TIMESTAMP,
    acct_session_time BIGINT,
    input_octets BIGINT,
    output_octets BIGINT,
    -- ... additional accounting fields
    created_at TIMESTAMP
);

CREATE INDEX idx_radius_accounting_tenant_time ON radius_accounting(tenant_id, acct_start_time DESC);
CREATE INDEX idx_radius_accounting_session ON radius_accounting(acct_session_id);
```

#### Network Infrastructure

**net_nas**
```sql
CREATE TABLE net_nas (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    ip_addr INET NOT NULL,
    secret VARCHAR(100),
    coa_port INT DEFAULT 3799,
    status VARCHAR(20) DEFAULT 'enabled',
    -- ... additional NAS fields
    created_at TIMESTAMP
);

CREATE INDEX idx_nas_tenant ON net_nas(tenant_id);
CREATE INDEX idx_nas_tenant_status ON net_nas(tenant_id, status);
CREATE INDEX idx_nas_ip_addr ON net_nas(ip_addr);
```

**net_server** (MikroTik devices)
```sql
CREATE TABLE net_server (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    public_ip INET NOT NULL,
    ports VARCHAR(20),
    username VARCHAR(100),
    password VARCHAR(255),
    router_status VARCHAR(20),
    -- ... additional server fields
    created_at TIMESTAMP
);

CREATE INDEX idx_server_tenant ON net_server(tenant_id);
```

#### Voucher System

**voucher_batch**
```sql
CREATE TABLE voucher_batch (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    profile_id BIGINT,
    status VARCHAR(20) DEFAULT 'active',
    -- ... additional batch fields
    created_at TIMESTAMP
);

CREATE INDEX idx_voucher_batch_tenant ON voucher_batch(tenant_id);
CREATE INDEX idx_voucher_batch_tenant_status ON voucher_batch(tenant_id, status);
```

**voucher**
```sql
CREATE TABLE voucher (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    batch_id BIGINT,
    voucher_code VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'unused',
    -- ... additional voucher fields
    created_at TIMESTAMP
);

CREATE INDEX idx_voucher_tenant ON voucher(tenant_id);
CREATE INDEX idx_voucher_tenant_status ON voucher(tenant_id, status);
CREATE INDEX idx_voucher_code ON voucher(voucher_code);
```

**voucher_template**
```sql
CREATE TABLE voucher_template (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    -- ... template fields
    created_at TIMESTAMP
);

CREATE INDEX idx_voucher_template_tenant ON voucher_template(tenant_id);
```

#### Commercial Features

**product**
```sql
CREATE TABLE product (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    -- ... product fields
    created_at TIMESTAMP
);

CREATE INDEX idx_product_tenant ON product(tenant_id);
```

---

## Schema Management

### Creating a Provider Schema

When a new provider registers and is approved:

```go
package main

import (
    "github.com/talkincode/toughradius/v9/internal/migration"
    "gorm.io/gorm"
)

func ApproveProviderRegistration(db *gorm.DB, registrationID int64) error {
    // 1. Get registration
    var registration domain.ProviderRegistration
    db.First(&registration, registrationID)

    // 2. Create provider record
    provider := &domain.Provider{
        Code:     generateProviderCode(registration.CompanyName),
        Name:     registration.CompanyName,
        Status:   "active",
        MaxUsers: registration.ExpectedUsers,
    }
    db.Create(provider)

    // 3. Create provider schema
    migrator := migration.NewSchemaMigrator(db)
    if err := migrator.CreateProviderSchema(provider.ID); err != nil {
        return fmt.Errorf("failed to create provider schema: %w", err)
    }

    // 4. Initialize provider's default data
    if err := initializeProviderData(db, provider.ID); err != nil {
        return fmt.Errorf("failed to initialize provider data: %w", err)
    }

    // 5. Update registration status
    registration.Status = "approved"
    registration.ReviewedAt = time.Now()
    db.Save(&registration)

    return nil
}
```

### Schema Creation Details

**Function:** `SchemaMigrator.CreateProviderSchema(tenantID int64)`

**Operations:**
1. Generate schema name: `provider_{tenantID}`
2. Execute: `CREATE SCHEMA IF NOT EXISTS provider_{tenantID}`
3. Create all required tables in the new schema
4. Create tenant-scoped indexes
5. Set default search path for provider operations

**Example SQL:**
```sql
-- Create schema
CREATE SCHEMA IF NOT EXISTS provider_1;

-- Set search path
SET search_path TO provider_1, public;

-- Create tables
CREATE TABLE provider_1.radius_user (
    -- ... table definition
);

-- Create indexes
CREATE INDEX idx_radius_user_tenant_status
    ON provider_1.radius_user(tenant_id, status);
```

### Dropping a Provider Schema

**Function:** `SchemaMigrator.DropProviderSchema(tenantID int64)`

**Operations:**
1. Execute: `DROP SCHEMA IF EXISTS provider_{tenantID} CASCADE`
2. CASCADE automatically drops all tables and objects in the schema
3. Provider record in `mst_provider` is soft-deleted (deleted_at timestamp)

**Safety Considerations:**
- Always backup schema before dropping
- Verify provider has no unpaid invoices
- Check for any active sessions before dropping
- Log the deletion action for audit purposes

---

## Database Migrations

### Migration System

**File:** `cmd/migrate/main.go`
**Binary:** `cmd/migrate/migrate`

**Architecture:**
- Migration tracking table: `schema_migrations`
- Each migration has unique ID, up/down functions
- Migrations run in transaction (all or nothing)
- Record migration after successful completion

### Running Migrations

**Up (apply all pending migrations):**
```bash
./cmd/migrate/migrate \
    -action=up \
    -dsn="host=localhost user=toughradius password=secret dbname=toughradius port=5432 sslmode=require"
```

**Down (rollback last migration):**
```bash
./cmd/migrate/migrate \
    -action=down \
    -dsn="host=localhost user=toughradius password=secret dbname=toughradius port=5432 sslmode=require"
```

### Implemented Migrations

#### Migration 001: Create Platform Schema
**ID:** `001_create_platform_schema`

**Creates:**
- `mst_provider` table
- `mst_provider_registration` table
- `schema_migrations` tracking table

#### Migration 002: Add Tenant Indexes
**ID:** `002_add_tenant_indexes`

**Creates:**
- Composite indexes on `(tenant_id, field)` for all multi-tenant tables
- Uses `CREATE INDEX CONCURRENTLY` for production safety
- Optimize queries for tenant-scoped operations

**Indexes Created:**
```sql
-- RadiusUser
CREATE INDEX CONCURRENTLY idx_radius_user_tenant_status
    ON radius_user(tenant_id, status);
CREATE INDEX CONCURRENTLY idx_radius_user_tenant_username
    ON radius_user(tenant_id, username);

-- RadiusProfile
CREATE INDEX CONCURRENTLY idx_radius_profile_tenant_status
    ON radius_profile(tenant_id, status);

-- RadiusOnline
CREATE INDEX CONCURRENTLY idx_radius_online_tenant
    ON radius_online(tenant_id);

-- RadiusAccounting
CREATE INDEX CONCURRENTLY idx_radius_accounting_tenant_time
    ON radius_accounting(tenant_id, acct_start_time DESC);

-- NetNas
CREATE INDEX CONCURRENTLY idx_nas_tenant
    ON net_nas(tenant_id);
CREATE INDEX CONCURRENTLY idx_nas_tenant_status
    ON net_nas(tenant_id, status);

-- VoucherBatch
CREATE INDEX CONCURRENTLY idx_voucher_batch_tenant
    ON voucher_batch(tenant_id);
CREATE INDEX CONCURRENTLY idx_voucher_batch_tenant_status
    ON voucher_batch(tenant_id, status);

-- Voucher
CREATE INDEX CONCURRENTLY idx_voucher_tenant
    ON voucher(tenant_id);
CREATE INDEX CONCURRENTLY idx_voucher_tenant_status
    ON voucher(tenant_id, status);
```

### Creating New Migrations

**Template:**
```go
m.RegisterMigration(migration.Migration{
    ID:          "003_your_migration_name",
    Description: "Human-readable description of what this migration does",
    Up: func(db *gorm.DB) error {
        // Migration logic here
        return nil
    },
    Down: func(db *gorm.DB) error {
        // Rollback logic here
        return nil
    },
})
```

**Best Practices:**
- Use `CONCURRENTLY` for index creation in production
- Always write rollback logic in `Down` function
- Test migrations on staging database first
- Use transactions for data modifications
- Backward compatible changes first, then remove old code

---

## Query Patterns

### Tenant-Scoped Queries

**Using GORM Scopes:**
```go
// Repository pattern with automatic tenant filtering
func (r *UserRepository) ListUsers(ctx context.Context, tenantID int64) ([]domain.RadiusUser, error) {
    var users []domain.RadiusUser
    err := r.db.Scopes(TenantScope(tenantID)).
        Find(&users).Error
    return users, err
}

// Scope definition
func TenantScope(tenantID int64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}
```

**Raw SQL with Schema Prefix:**
```go
// Query specific provider schema
var users []RadiusUser
db.Table("provider_1.radius_user").
    Where("tenant_id = ? AND status = ?", 1, "enabled").
    Find(&users)
```

### Cross-Tenant Queries (Admin Only)

```go
// Platform admins can aggregate across providers
func (r *AdminRepository) GetProviderStats() ([]ProviderStats, error) {
    var stats []ProviderStats

    // Query all providers
    db.Table("mst_provider").
        Select("id, name, max_users").
        Where("status = ?", "active").
        Find(&stats)

    // Get user counts per provider
    for i := range stats {
        var count int64
        db.Table("radius_user").
            Where("tenant_id = ?", stats[i].ID).
            Count(&count)
        stats[i].CurrentUsers = count
    }

    return stats, nil
}
```

---

## Performance Considerations

### Index Strategy

**Composite Indexes (tenant_id, field):**
- Optimize tenant-scoped queries
- Support sorting and filtering
- Cover indexes for common query patterns

**Example Query Optimization:**
```sql
-- Query: Get enabled users for tenant 1
SELECT * FROM radius_user
WHERE tenant_id = 1 AND status = 'enabled'
ORDER BY username;

-- Index used: idx_radius_user_tenant_status
-- Index also covers: WHERE tenant_id = 1 AND status = 'enabled'
```

### Connection Pooling

**Recommended Settings:**
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    ConnPool: &stdsql.DB{
        MaxIdleConns: 10,
        MaxOpenConns: 100,
        ConnMaxLifetime: time.Hour,
    },
})
```

**Rationale:**
- 100 providers × 10 connections = 1000 max concurrent connections
- Each provider gets dedicated connection pool
- Prevents "noisy neighbor" problem at connection level

### Partitioning (Future Optimization)

For providers with >100K users, consider table partitioning:

```sql
-- Partition radius_user by tenant_id
CREATE TABLE radius_user (
    -- ... columns
) PARTITION BY LIST (tenant_id);

-- Create partition for large provider
CREATE TABLE radius_user_provider_1
    PARTITION OF radius_user
    FOR VALUES IN (1);
```

**Benefits:**
- Query pruning (only scan relevant partition)
- Faster vacuum and analyze
- Can archive old provider partitions

---

## Security Considerations

### SQL Injection Prevention

**Schema Names:**
```go
import "github.com/lib/pq"

// WRONG: Direct interpolation
query := fmt.Sprintf("CREATE SCHEMA %s", schemaName)

// CORRECT: Quote identifier
query := fmt.Sprintf("CREATE SCHEMA %s", pq.QuoteIdentifier(schemaName))
```

**User Input:**
```go
// WRONG: Direct interpolation
query := fmt.Sprintf("SELECT * FROM radius_user WHERE username = '%s'", userInput)

// CORRECT: Parameterized query
query := "SELECT * FROM radius_user WHERE username = ?"
db.Raw(query, userInput).Scan(&user)
```

### Tenant Isolation Enforcement

**Middleware:**
```go
func TenantMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Extract tenant from JWT
            tenantID := c.Get("tenant_id").(int64)

            // Add to context
            ctx := tenant.WithTenantID(c.Request().Context(), tenantID)
            c.SetRequest(c.Request().WithContext(ctx))

            return next(c)
        }
    }
}
```

**Repository Validation:**
```go
func (r *UserRepository) GetUser(ctx context.Context, id, tenantID int64) (*domain.RadiusUser, error) {
    var user domain.RadiusUser

    // Always filter by tenant_id
    err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).
        First(&user).Error

    if err != nil {
        return nil, err
    }

    // Double-check tenant match
    if user.TenantID != tenantID {
        return nil, errors.New("tenant mismatch")
    }

    return &user, nil
}
```

---

## Backup & Restore

### Provider-Level Backup

**Backup Single Provider:**
```bash
pg_dump -h localhost -U toughradius -d toughradius \
    -n provider_1 \
    -f provider_1_backup_$(date +%Y%m%d).sql
```

**Restore Provider:**
```bash
psql -h localhost -U toughradius -d toughradius \
    -f provider_1_backup_20260320.sql
```

### Full Platform Backup

**Backup All Providers:**
```bash
pg_dump -h localhost -U toughradius -d toughradius \
    -n public \
    -n provider_* \
    -f full_backup_$(date +%Y%m%d).sql
```

### Automated Backups

**Cron Job:**
```bash
# Daily provider backups at 2 AM
0 2 * * * /usr/local/bin/backup-provider.sh

# Weekly full backup on Sunday at 3 AM
0 3 * * 0 /usr/local/bin/backup-full.sh
```

---

## Monitoring & Maintenance

### Schema Size Monitoring

```sql
-- Monitor schema sizes per provider
SELECT
    schemaname,
    pg_size_pretty(pg_schema_size(schemaname)) AS size,
    pg_schema_size(schemaname) AS size_bytes
FROM pg_namespace
WHERE schemaname LIKE 'provider_%'
ORDER BY size_bytes DESC;
```

### Table Size Monitoring

```sql
-- Largest tables across all providers
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname LIKE 'provider_%'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;
```

### Index Maintenance

**Rebuild Indexes (CONCURRENTLY):**
```sql
-- Reindex specific table
REINDEX INDEX CONCURRENTLY idx_radius_user_tenant_status;

-- Reindex entire schema (PostgreSQL 14+)
REINDEX SCHEMA CONCURRENTLY provider_1;
```

**Vacuum Analyze:**
```sql
-- Run vacuum analyze on large provider
VACUUM ANALYZE provider_1.radius_user;
```

---

## Troubleshooting

### Common Issues

**Issue 1: Schema already exists**
```
Error: schema "provider_1" already exists
```
**Solution:** Check if provider was partially created, clean up before retry
```sql
DROP SCHEMA IF EXISTS provider_1 CASCADE;
```

**Issue 2: Migration lock timeout**
```
Error: timeout acquiring lock
```
**Solution:** Check for long-running transactions
```sql
SELECT pid, query, state
FROM pg_stat_activity
WHERE state = 'active'
ORDER BY query_start;
```

**Issue 3: Index creation fails**
```
Error: could not create index
```
**Solution:** Drop conflicting index first
```sql
DROP INDEX IF EXISTS idx_radius_user_tenant_status;
```

---

## Glossary

- **Provider**: ISP/WISP tenant using the platform
- **Tenant**: Synonym for provider (multi-tenant architecture)
- **Schema**: PostgreSQL namespace for organizing database objects
- **tenant_id**: Numeric identifier for a provider
- **Migration**: Database schema change script
- **CONCURRENTLY**: PostgreSQL option to create indexes without table locks
- **Search Path**: PostgreSQL schema search order (like PATH for executables)

---

## References

- [PostgreSQL Schema Documentation](https://www.postgresql.org/docs/current/ddl-schemas.html)
- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL Indexing Best Practices](https://www.postgresql.org/docs/current/indexes.html)
- [Database Migration Best Practices](https://martinfowler.com/articles/evodb.html)

---

**Document maintained by:** Engineering Team
**Last review:** 2026-03-20
**Next review:** After Phase 2 completion
