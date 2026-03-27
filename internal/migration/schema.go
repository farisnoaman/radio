package migration

import (
	"fmt"

	"github.com/lib/pq"
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
	if err := sm.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schemaName))).Error; err != nil {
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
	if err := sm.db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(schemaName))).Error; err != nil {
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
