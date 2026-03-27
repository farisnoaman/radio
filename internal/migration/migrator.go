package migration

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Migration defines a single database migration step.
//
// The ID field should follow the format "NNN_description" where NNN is a sequence number
// (e.g., "001_create_platform_schema"). The description should clearly explain what
// the migration does.
//
// The Up function contains the logic to apply the migration (CREATE TABLE, ALTER COLUMN, etc.).
// It receives a GORM DB instance which may be used within a transaction.
//
// The Down function contains rollback logic to undo the migration. If the migration
// cannot be rolled back (e.g., DROP SCHEMA), Down can be nil. This is acceptable for
// irreversible operations.
type Migration struct {
	ID          string
	Description string
	Up          func(db *gorm.DB) error
	Down        func(db *gorm.DB) error
}

// Migrator runs database migrations with tracking.
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
	log.Println("Starting database migrations...")

	// Create migrations tracking table if not exists
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	for _, migration := range m.migrations {
		// Check if migration already ran
		var count int64
		m.db.Table("schema_migrations").Where("id = ?", migration.ID).Count(&count)
		if count > 0 {
			log.Printf("Migration %s already applied, skipping", migration.ID)
			continue // Already migrated
		}

		// Run migration
		log.Printf("Applying migration: %s - %s", migration.ID, migration.Description)
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.ID, err)
		}

		// Record migration
		if err := m.db.Table("schema_migrations").Create(map[string]interface{}{
			"id":          migration.ID,
			"description": migration.Description,
		}).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
		}
		log.Printf("Successfully applied migration: %s", migration.ID)
	}

	log.Println("All migrations completed successfully")
	return nil
}

// Down rolls back the last migration.
func (m *Migrator) Down() error {
	if len(m.migrations) == 0 {
		return fmt.Errorf("no migrations registered")
	}

	// Get the last migration
	lastMigration := m.migrations[len(m.migrations)-1]

	// Check if it was applied
	var count int64
	m.db.Table("schema_migrations").Where("id = ?", lastMigration.ID).Count(&count)
	if count == 0 {
		return fmt.Errorf("migration %s was not applied", lastMigration.ID)
	}

	// Run rollback
	log.Printf("Rolling back migration: %s - %s", lastMigration.ID, lastMigration.Description)
	if lastMigration.Down != nil {
		if err := lastMigration.Down(m.db); err != nil {
			return fmt.Errorf("rollback failed for migration %s: %w", lastMigration.ID, err)
		}
	}

	// Remove from tracking
	m.db.Table("schema_migrations").Where("id = ?", lastMigration.ID).Delete(nil)
	log.Printf("Successfully rolled back migration: %s", lastMigration.ID)

	return nil
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist.
func (m *Migrator) createMigrationsTable() error {
	return m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id VARCHAR(255) PRIMARY KEY,
			description TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

// MigrationRunner handles running database migrations for the platform.
type MigrationRunner struct {
	db         *gorm.DB
	migrator *SchemaMigrator
}

// NewMigrationRunner creates a new migration runner.
func NewMigrationRunner(db *gorm.DB) *MigrationRunner {
	return &MigrationRunner{
		db:         db,
		migrator: NewSchemaMigrator(db),
	}
}

// RunMigrations executes all pending database migrations.
func (mr *MigrationRunner) RunMigrations() error {
	log.Println("Starting database migrations...")

	// Auto-migrate platform models (shared tables)
	if err := mr.migratePlatformModels(); err != nil {
		return fmt.Errorf("failed to migrate platform models: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// migratePlatformModels migrates the platform-level models.
// These are shared across all providers.
func (mr *MigrationRunner) migratePlatformModels() error {
	log.Println("Migrating platform models...")

	// Platform models will be migrated here
	// These models are shared across all tenants
	// We'll populate this in the next task

	return nil
}

// CreateProviderSchema creates a new provider schema with all required tables.
func (mr *MigrationRunner) CreateProviderSchema(tenantID int64) error {
	log.Printf("Creating schema for provider %d...", tenantID)

	if err := mr.migrator.CreateProviderSchema(tenantID); err != nil {
		return fmt.Errorf("failed to create provider schema: %w", err)
	}

	log.Printf("Successfully created schema for provider %d", tenantID)
	return nil
}

// DropProviderSchema drops a provider schema and all its data.
func (mr *MigrationRunner) DropProviderSchema(tenantID int64) error {
	log.Printf("Dropping schema for provider %d...", tenantID)

	if err := mr.migrator.DropProviderSchema(tenantID); err != nil {
		return fmt.Errorf("failed to drop provider schema: %w", err)
	}

	log.Printf("Successfully dropped schema for provider %d", tenantID)
	return nil
}
