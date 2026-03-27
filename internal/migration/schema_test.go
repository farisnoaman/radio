package migration

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestCreateProviderSchema(t *testing.T) {
	// Setup test database connection
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL environment variable not set")
	}
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
	t.Cleanup(func() {
		migrator.DropProviderSchema(1)
	})
}

func TestDropProviderSchema(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL environment variable not set")
	}
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

func TestSchemaExists(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL environment variable not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Cannot connect to test database")
	}

	migrator := NewSchemaMigrator(db)

	// Test non-existent schema
	exists, _ := migrator.SchemaExists(999)
	if exists {
		t.Error("Expected schema 999 to not exist")
	}

	// Create schema
	migrator.CreateProviderSchema(999)

	// Test existing schema
	exists, _ = migrator.SchemaExists(999)
	if !exists {
		t.Error("Expected schema 999 to exist")
	}

	// Cleanup
	migrator.DropProviderSchema(999)
}
