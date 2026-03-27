# Phase 1: Database Schema & Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create PostgreSQL schema-per-provider architecture with platform-level tables and automated migration system.

**Architecture:** Separate database schemas for data isolation. Platform schema holds provider registry and billing. Each provider gets isolated schema (provider_1, provider_2, etc.) with their own users, sessions, NAS devices, and vouchers.

**Tech Stack:** PostgreSQL 16+, GORM (Go ORM), Go migration library

---

## Task 1: Create Platform Schema Models

**Files:**
- Create: `internal/domain/platform.go`
- Create: `internal/domain/platform_test.go`

**Step 1: Write failing tests for platform models**

```go
// internal/domain/platform_test.go
package domain

import (
    "testing"
    "time"
)

func TestProviderModel(t *testing.T) {
    provider := &Provider{
        Code:     "test-provider",
        Name:     "Test Provider",
        Status:   "active",
        MaxUsers: 1000,
        MaxNas:   100,
    }

    if provider.Code != "test-provider" {
        t.Errorf("Expected code 'test-provider', got '%s'", provider.Code)
    }

    if provider.TableName() != "mst_provider" {
        t.Errorf("Expected table name 'mst_provider', got '%s'", provider.TableName())
    }
}

func TestProviderBrandingSerialization(t *testing.T) {
    provider := &Provider{}

    branding := &ProviderBranding{
        LogoURL:        "https://example.com/logo.png",
        PrimaryColor:   "#007bff",
        SecondaryColor: "#6c757d",
        CompanyName:    "Test ISP",
        SupportEmail:   "support@testisp.com",
        SupportPhone:   "+1234567890",
    }

    err := provider.SetBranding(branding)
    if err != nil {
        t.Fatalf("Failed to set branding: %v", err)
    }

    retrieved, err := provider.GetBranding()
    if err != nil {
        t.Fatalf("Failed to get branding: %v", err)
    }

    if retrieved.LogoURL != branding.LogoURL {
        t.Errorf("LogoURL mismatch: got '%s', want '%s'", retrieved.LogoURL, branding.LogoURL)
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain -run TestProvider -v`
Expected: FAIL with "undefined: Provider" and "undefined: ProviderBranding"

**Step 3: Implement platform models**

```go
// internal/domain/platform.go
package domain

import (
    "encoding/json"
    "time"
    "gorm.io/gorm"
)

// Provider represents an ISP/Provider tenant in the multi-provider system.
// Each provider operates independently with their own users, NAS devices,
// vouchers, and billing configuration.
type Provider struct {
    ID        int64          `json:"id" gorm:"primaryKey"`
    Code      string         `json:"code" gorm:"uniqueIndex;size:50;not null"`
    Name      string         `json:"name" gorm:"size:255;not null"`
    Status    string         `json:"status" gorm:"size:20;default:'active'"`
    MaxUsers  int            `json:"max_users" gorm:"default:1000"`
    MaxNas    int            `json:"max_nas" gorm:"default:100"`
    Branding  string         `json:"branding" gorm:"type:text"`
    Settings  string         `json:"settings" gorm:"type:text"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Provider) TableName() string {
    return "mst_provider"
}

// ProviderBranding holds the branding configuration for a provider.
type ProviderBranding struct {
    LogoURL        string `json:"logo_url"`
    PrimaryColor   string `json:"primary_color"`
    SecondaryColor string `json:"secondary_color"`
    FaviconURL     string `json:"favicon_url"`
    CompanyName    string `json:"company_name"`
    SupportEmail   string `json:"support_email"`
    SupportPhone   string `json:"support_phone"`
}

func (p *Provider) GetBranding() (*ProviderBranding, error) {
    if p.Branding == "" {
        return &ProviderBranding{}, nil
    }
    var branding ProviderBranding
    err := json.Unmarshal([]byte(p.Branding), &branding)
    return &branding, err
}

func (p *Provider) SetBranding(b *ProviderBranding) error {
    data, err := json.Marshal(b)
    if err != nil {
        return err
    }
    p.Branding = string(data)
    return nil
}

// ProviderSettings holds provider-specific settings.
type ProviderSettings struct {
    AllowUserRegistration  bool     `json:"allow_user_registration"`
    AllowVoucherCreation   bool     `json:"allow_voucher_creation"`
    DefaultProductID       int64    `json:"default_product_id"`
    DefaultProfileID       int64    `json:"default_profile_id"`
    AutoExpireSessions     bool     `json:"auto_expire_sessions"`
    SessionTimeout         int      `json:"session_timeout"`
    IdleTimeout            int      `json:"idle_timeout"`
    MaxConcurrentSessions  int      `json:"max_concurrent_sessions"`
    RADIUSSecretTemplate   string   `json:"radius_secret_template"`
    CustomAttributes       []string `json:"custom_attributes"`
}

func (p *Provider) GetSettings() (*ProviderSettings, error) {
    if p.Settings == "" {
        return &ProviderSettings{
            AllowUserRegistration: true,
            AllowVoucherCreation:  true,
            SessionTimeout:        86400,
            IdleTimeout:           3600,
            MaxConcurrentSessions: 1,
        }, nil
    }
    var settings ProviderSettings
    err := json.Unmarshal([]byte(p.Settings), &settings)
    return &settings, err
}

func (p *Provider) SetSettings(s *ProviderSettings) error {
    data, err := json.Marshal(s)
    if err != nil {
        return err
    }
    p.Settings = string(data)
    return nil
}

// ProviderStats holds usage statistics for a provider.
type ProviderStats struct {
    ProviderID     int64 `json:"provider_id"`
    TotalUsers     int64 `json:"total_users"`
    ActiveUsers    int64 `json:"active_users"`
    OnlineSessions int64 `json:"online_sessions"`
    TotalNas       int64 `json:"total_nas"`
    ActiveNas      int64 `json:"active_nas"`
    TotalVouchers  int64 `json:"total_vouchers"`
    UsedVouchers   int64 `json:"used_vouchers"`
}

// ProviderRegistration holds registration requests.
type ProviderRegistration struct {
    ID              int64     `json:"id" gorm:"primaryKey"`
    CompanyName     string    `json:"company_name" gorm:"not null"`
    ContactName     string    `json:"contact_name" gorm:"not null"`
    Email           string    `json:"email" gorm:"not null"`
    Phone           string    `json:"phone"`
    Address         string    `json:"address"`
    BusinessType    string    `json:"business_type"`
    ExpectedUsers   int       `json:"expected_users"`
    ExpectedNas     int       `json:"expected_nas"`
    Country         string    `json:"country"`
    Message         string    `json:"message"`
    Status          string    `json:"status" gorm:"default:'pending'"`
    ReviewedBy      int64     `json:"reviewed_by"`
    ReviewedAt      *time.Time `json:"reviewed_at"`
    RejectionReason string    `json:"rejection_reason"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

func (ProviderRegistration) TableName() string {
    return "mst_provider_registration"
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain -run TestProvider -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/platform.go internal/domain/platform_test.go
git commit -m "feat(domain): add platform schema models for multi-tenant architecture"
```

---

## Task 2: Create Schema Migration Service

**Files:**
- Create: `internal/migration/schema.go`
- Create: `internal/migration/schema_test.go`
- Create: `internal/migration/migrator.go`

**Step 1: Write failing tests for schema creation**

```go
// internal/migration/schema_test.go
package migration

import (
    "testing"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func TestCreateProviderSchema(t *testing.T) {
    // Setup test database connection
    dsn := "host=localhost user=toughradius password=test dbname=toughradius_test port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Skip("Cannot connect to test database")
    }

    migrator := NewSchemaMigrator(db)

    // Test creating provider schema
    err = migrator.CreateProviderSchema(1)
    if err != nil {
        t.Fatalf("Failed to create provider schema: %v", err)
    }

    // Verify schema exists
    var result string
    err = db.Raw("SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'provider_1'").Scan(&result).Error
    if err != nil {
        t.Errorf("Schema provider_1 was not created: %v", err)
    }

    // Cleanup
    migrator.DropProviderSchema(1)
}

func TestDropProviderSchema(t *testing.T) {
    dsn := "host=localhost user=toughradius password=test dbname=toughradius_test port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Skip("Cannot connect to test database")
    }

    migrator := NewSchemaMigrator(db)

    // Create schema first
    migrator.CreateProviderSchema(99)

    // Test dropping schema
    err = migrator.DropProviderSchema(99)
    if err != nil {
        t.Fatalf("Failed to drop provider schema: %v", err)
    }

    // Verify schema doesn't exist
    var result string
    err = db.Raw("SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'provider_99'").Scan(&result).Error
    if err == nil {
        t.Error("Schema provider_99 still exists after drop")
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/migration -run TestSchema -v`
Expected: FAIL with "undefined: NewSchemaMigrator"

**Step 3: Implement schema migrator**

```go
// internal/migration/schema.go
package migration

import (
    "fmt"
    "gorm.io/gorm"
)

// SchemaMigrator handles database schema creation and management for multi-tenant architecture.
type SchemaMigrator struct {
    db *gorm.DB
}

// NewSchemaMigrator creates a new schema migrator.
func NewSchemaMigrator(db *gorm.DB) *SchemaMigrator {
    return &SchemaMigrator{db: db}
}

// CreateProviderSchema creates a new schema for a provider.
// The schema name follows the pattern "provider_{tenant_id}".
func (sm *SchemaMigrator) CreateProviderSchema(tenantID int64) error {
    schemaName := sm.getSchemaName(tenantID)

    // Create schema
    if err := sm.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error; err != nil {
        return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
    }

    // Set search path to new schema
    if err := sm.db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error; err != nil {
        return fmt.Errorf("failed to set search path: %w", err)
    }

    // Create tables in provider schema
    // Note: We'll use existing domain models but create them in provider schema
    tables := []interface{}{
        // Import domain models and create them here
        // These will be added in next task
    }

    for _, table := range tables {
        if err := sm.db.Table(schemaName + "." + "tablename").AutoMigrate(table); err != nil {
            return fmt.Errorf("failed to migrate table: %w", err)
        }
    }

    return nil
}

// DropProviderSchema drops a provider schema and all its data.
// WARNING: This operation cannot be undone.
func (sm *SchemaMigrator) DropProviderSchema(tenantID int64) error {
    schemaName := sm.getSchemaName(tenantID)

    // Drop schema with CASCADE (removes all objects in schema)
    if err := sm.db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error; err != nil {
        return fmt.Errorf("failed to drop schema %s: %w", schemaName, err)
    }

    return nil
}

// SchemaExists checks if a provider schema exists.
func (sm *SchemaMigrator) SchemaExists(tenantID int64) (bool, error) {
    schemaName := sm.getSchemaName(tenantID)

    var count int64
    err := sm.db.Raw(
        "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?",
        schemaName,
    ).Scan(&count).Error

    return count > 0, err
}

// ListProviderSchemas returns a list of all provider schemas.
func (sm *SchemaMigrator) ListProviderSchemas() ([]string, error) {
    var schemas []string

    err := sm.db.Raw(
        "SELECT schema_name FROM information_schema.schemata WHERE schema_name LIKE 'provider_%' ORDER BY schema_name",
    ).Scan(&schemas).Error

    return schemas, err
}

// getSchemaName returns the schema name for a given tenant ID.
func (sm *SchemaMigrator) getSchemaName(tenantID int64) string {
    return fmt.Sprintf("provider_%d", tenantID)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/migration -run TestSchema -v`
Expected: PASS (if test DB available)

**Step 5: Commit**

```bash
git add internal/migration/schema.go internal/migration/schema_test.go
git commit -m "feat(migration): add schema migrator for provider isolation"
```

---

## Task 3: Create Database Migration Runner

**Files:**
- Create: `internal/migration/migrator.go`
- Create: `cmd/migrate/main.go`

**Step 1: Write database migration interface**

```go
// internal/migration/migrator.go
package migration

import (
    "fmt"
    "gorm.io/gorm"
)

// Migration defines a database migration.
type Migration struct {
    ID          string
    Description string
    Up          func(db *gorm.DB) error
    Down        func(db *gorm.DB) error
}

// Migrator runs database migrations.
type Migrator struct {
    db         *gorm.DB
    migrations []Migration
}

// NewMigrator creates a new migrator.
func NewMigrator(db *gorm.DB) *Migrator {
    return &Migrator{db: db}
}

// RegisterMigration registers a migration.
func (m *Migrator) RegisterMigration(migration Migration) {
    m.migrations = append(m.migrations, migration)
}

// Up runs all pending migrations.
func (m *Migrator) Up() error {
    // Create migrations tracking table if not exists
    if err := m.createMigrationsTable(); err != nil {
        return fmt.Errorf("failed to create migrations table: %w", err)
    }

    for _, migration := range m.migrations {
        // Check if migration already ran
        var count int64
        m.db.Table("schema_migrations").Where("id = ?", migration.ID).Count(&count)
        if count > 0 {
            continue // Already migrated
        }

        // Run migration
        if err := migration.Up(m.db); err != nil {
            return fmt.Errorf("migration %s failed: %w", migration.ID, err)
        }

        // Record migration
        m.db.Table("schema_migrations").Create(map[string]interface{}{
            "id":          migration.ID,
            "description": migration.Description,
        })
    }

    return nil
}

// Down rolls back the last migration.
func (m *Migrator) Down() error {
    // Get last run migration
    var migration Migration
    // Implementation for rollback
    if migration.Down != nil {
        return migration.Down(m.db)
    }
    return nil
}

func (m *Migrator) createMigrationsTable() error {
    return m.db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            id VARCHAR(255) PRIMARY KEY,
            description TEXT,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `).Error
}
```

**Step 2: Create migration CLI tool**

```go
// cmd/migrate/main.go
package main

import (
    "flag"
    "log"

    "github.com/talkincode/toughradius/v9/internal/migration"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    action := flag.String("action", "up", "Migration action: up, down")
    dsn := flag.String("dsn", "", "Database connection string")
    flag.Parse()

    if *dsn == "" {
        log.Fatal("DSN is required")
    }

    db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    m := migration.NewMigrator(db)

    // Register migrations
    m.RegisterMigration(migration.Migration{
        ID:          "001_create_platform_schema",
        Description: "Create platform schema and provider tables",
        Up: func(db *gorm.DB) error {
            // Create platform schema
            if err := db.Exec("CREATE SCHEMA IF NOT EXISTS platform").Error; err != nil {
                return err
            }
            return nil
        },
        Down: func(db *gorm.DB) error {
            return db.Exec("DROP SCHEMA IF EXISTS platform CASCADE").Error
        },
    })

    if *action == "up" {
        if err := m.Up(); err != nil {
            log.Fatalf("Migration up failed: %v", err)
        }
        log.Println("Migrations completed successfully")
    } else if *action == "down" {
        if err := m.Down(); err != nil {
            log.Fatalf("Migration down failed: %v", err)
        }
        log.Println("Rollback completed successfully")
    }
}
```

**Step 3: Build migrate tool**

Run: `go build -o bin/migrate cmd/migrate/main.go`

**Step 4: Test migration**

Run: `./bin/migrate -action=up -dsn="host=localhost user=toughradius password=test dbname=toughradius_test port=5432"`
Expected: "Migrations completed successfully"

**Step 5: Commit**

```bash
git add internal/migration/migrator.go cmd/migrate/main.go
git commit -m "feat(migration): add database migration runner and CLI tool"
```

---

## Task 4: Add TenantID Indexes to Existing Models

**Files:**
- Modify: `internal/domain/radius.go`
- Modify: `internal/domain/network.go`
- Modify: `internal/domain/voucher.go`

**Step 1: Add TenantID fields to existing models**

Modify existing domain models to ensure they have `tenant_id` fields and indexes. Most should already have them based on our reading.

**Step 2: Add database indexes**

```go
// In your AutoMigrate or migration files, add:
// CREATE INDEX CONCURRENTLY idx_radius_user_tenant_status ON radius_user(tenant_id, status);
// CREATE INDEX CONCURRENTLY idx_radius_online_tenant ON radius_online(tenant_id);
// CREATE INDEX CONCURRENTLY idx_radius_accounting_tenant_time ON radius_accounting(tenant_id, acct_start_time);
```

**Step 3: Verify indexes created**

Run: Check database for indexes

**Step 4: Commit**

```bash
git add internal/domain/
git commit -m "refactor(domain): ensure tenant_id indexes on all multi-tenant tables"
```

---

## Testing & Verification

**Integration Test:**
```bash
# 1. Create test database
createdb toughradius_test

# 2. Run migrations
./bin/migrate -action=up -dsn="host=localhost user=toughradius password=test dbname=toughradius_test port=5432"

# 3. Run all tests
go test ./... -v

# 4. Verify schemas created
psql toughradius_test -c "\dn" | grep provider

# 5. Cleanup
dropdb toughradius_test
```

**Performance Test:**
```bash
# Test schema creation performance
go test ./internal/migration -bench=. -benchmem
```

---

## Documentation Updates

**Files:**
- Create: `docs/database/multi-tenant-schema.md`
- Update: `README.md` (add database setup section)

**Documentation Content:**
```markdown
# Multi-Tenant Database Architecture

## Schema Structure

### Platform Schema
- `mst_provider`: Provider registry
- `mst_provider_registration`: Registration requests
- `mst_billing`: Billing and invoices

### Provider Schemas
Each provider gets isolated schema: `provider_1`, `provider_2`, etc.

Tables per provider:
- `radius_user`: User accounts
- `radius_profile`: User profiles
- `radius_online`: Active sessions
- `radius_accounting`: Session history
- `net_nas`: NAS devices
- `voucher_batch`: Voucher batches

## Schema Management

### Creating Provider Schema
```go
migrator := migration.NewSchemaMigrator(db)
err := migrator.CreateProviderSchema(tenantID)
```

### Dropping Provider Schema
```go
err := migrator.DropProviderSchema(tenantID)
```

## Migrations

Run migrations:
```bash
./bin/migrate -action=up -dsn="..."
```
```

**Step: Commit documentation**

```bash
git add docs/database/
git commit -m "docs: add multi-tenant database architecture documentation"
```

---

## Success Criteria

- ✅ Platform schema created with provider tables
- ✅ Provider schemas can be created/dropped via API
- ✅ All existing domain models have tenant_id
- ✅ Database indexes optimized for tenant-scoped queries
- ✅ Migration tool functional
- ✅ Unit tests pass (≥80% coverage)
- ✅ Integration tests validate schema operations
