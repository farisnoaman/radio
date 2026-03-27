package main

import (
	"flag"
	"fmt"
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
		log.Fatal("DSN is required. Usage: ./migrate -action=up -dsn=\"host=localhost user=toughradius password=test dbname=toughradius_test port=5432 sslmode=disable\"")
	}

	log.Printf("Connecting to database...")
	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Connected successfully")
	m := migration.NewMigrator(db)

	// Register migrations
	m.RegisterMigration(migration.Migration{
		ID:          "001_create_platform_schema",
		Description: "Create platform schema and provider tables",
		Up: func(db *gorm.DB) error {
			// Create platform schema
			if err := db.Exec("CREATE SCHEMA IF NOT EXISTS platform").Error; err != nil {
				return fmt.Errorf("failed to create platform schema: %w", err)
			}
			log.Println("Created platform schema")
			return nil
		},
		Down: func(db *gorm.DB) error {
			return db.Exec("DROP SCHEMA IF EXISTS platform CASCADE").Error
		},
	})

	// Add more migrations here as needed
	// m.RegisterMigration(migration.Migration{
	//     ID:          "002_add_providers_table",
	//     Description: "Create providers table",
	//     Up: func(db *gorm.DB) error {
	//         // Migration logic here
	//         return nil
	//     },
	//     Down: func(db *gorm.DB) error {
	//         // Rollback logic here
	//         return nil
	//     },
	// })

	// Migration 002: Add tenant_id indexes for multi-tenant query performance
	m.RegisterMigration(migration.Migration{
		ID:          "002_add_tenant_indexes",
		Description: "Add tenant_id indexes for multi-tenant query performance",
		Up: func(db *gorm.DB) error {
			// RadiusUser indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_radius_user_tenant_status
				ON radius_user(tenant_id, status)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_radius_user_tenant_status: %w", err)
			}

			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_radius_user_tenant_username
				ON radius_user(tenant_id, username)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_radius_user_tenant_username: %w", err)
			}

			// RadiusProfile indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_radius_profile_tenant_status
				ON radius_profile(tenant_id, status)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_radius_profile_tenant_status: %w", err)
			}

			// RadiusOnline indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_radius_online_tenant
				ON radius_online(tenant_id)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_radius_online_tenant: %w", err)
			}

			// RadiusAccounting indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_radius_accounting_tenant_time
				ON radius_accounting(tenant_id, acct_start_time)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_radius_accounting_tenant_time: %w", err)
			}

			// NetNas indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_net_nas_tenant
				ON net_nas(tenant_id)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_net_nas_tenant: %w", err)
			}

			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_net_nas_tenant_status
				ON net_nas(tenant_id, status)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_net_nas_tenant_status: %w", err)
			}

			// VoucherBatch indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_voucher_batch_tenant
				ON voucher_batch(tenant_id)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_voucher_batch_tenant: %w", err)
			}

			// Voucher indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_voucher_tenant_status
				ON voucher(tenant_id, status)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_voucher_tenant_status: %w", err)
			}

			// VoucherTopup indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_voucher_topup_tenant_status
				ON voucher_topup(tenant_id, status)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_voucher_topup_tenant_status: %w", err)
			}

			// VoucherSubscription indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_voucher_subscription_tenant_status
				ON voucher_subscription(tenant_id, status)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_voucher_subscription_tenant_status: %w", err)
			}

			// VoucherBundle indexes
			if err := db.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_voucher_bundle_tenant_status
				ON voucher_bundle(tenant_id, status)
			`).Error; err != nil {
				return fmt.Errorf("failed to create idx_voucher_bundle_tenant_status: %w", err)
			}

			log.Println("Successfully created all tenant_id indexes")
			return nil
		},
		Down: func(db *gorm.DB) error {
			log.Println("Dropping tenant_id indexes...")
			// Rollback indexes
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_radius_user_tenant_status`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_radius_user_tenant_username`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_radius_profile_tenant_status`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_radius_online_tenant`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_radius_accounting_tenant_time`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_net_nas_tenant`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_net_nas_tenant_status`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_voucher_batch_tenant`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_voucher_tenant_status`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_voucher_topup_tenant_status`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_voucher_subscription_tenant_status`)
			db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_voucher_bundle_tenant_status`)
			log.Println("Successfully dropped all tenant_id indexes")
			return nil
		},
	})

	// Migration 003: Create usage alerts and notification preferences tables
	m.RegisterMigration(migration.Migration{
		ID:          "003_create_usage_alerts_tables",
		Description: "Create usage_alerts and notification_preferences tables",
		Up: func(db *gorm.DB) error {
			log.Println("Creating usage_alerts and notification_preferences tables...")
			if err := migration.CreateUsageAlertsTables(db); err != nil {
				return err
			}
			log.Println("Successfully created usage_alerts and notification_preferences tables")
			return nil
		},
		Down: func(db *gorm.DB) error {
			log.Println("Dropping usage_alerts and notification_preferences tables...")
			if err := migration.DropUsageAlertsTables(db); err != nil {
				return err
			}
			log.Println("Successfully dropped usage_alerts and notification_preferences tables")
			return nil
		},
	})

	// Migration 004: Create reporting tables for provider dashboard
	m.RegisterMigration(migration.Migration{
		ID:          "004_create_reporting_tables",
		Description: "Create reporting tables for provider dashboard",
		Up: func(db *gorm.DB) error {
			log.Println("Creating reporting tables...")
			if err := migration.CreateReportingTables(db); err != nil {
				return err
			}
			log.Println("Successfully created reporting tables")
			return nil
		},
		Down: func(db *gorm.DB) error {
			log.Println("Dropping reporting tables...")
			if err := migration.DropReportingTables(db); err != nil {
				return err
			}
			log.Println("Successfully dropped reporting tables")
			return nil
		},
	})

	// Migration 005: Add certificate and IPoE tables for 802.1x and DHCP authentication
	m.RegisterMigration(migration.Migration{
		ID:          "005_add_certificate_and_ipoe_tables",
		Description: "Add certificate and IPoE tables for 802.1x and DHCP authentication",
		Up: func(db *gorm.DB) error {
			log.Println("Creating certificate and IPoE tables...")

			if err := migration.CreateCertificateAndIPoeTables(db); err != nil {
				return fmt.Errorf("failed to create certificate and IPoE tables: %w", err)
			}

			log.Println("Successfully created certificate and IPoE tables")
			return nil
		},
		Down: func(db *gorm.DB) error {
			log.Println("Dropping certificate and IPoE tables...")

			if err := migration.DropCertificateAndIPoeTables(db); err != nil {
				return fmt.Errorf("failed to drop certificate and IPoE tables: %w", err)
			}

			log.Println("Successfully dropped certificate and IPoE tables")
			return nil
		},
	})

	// Migration 006: Add RADIUS proxy tables for realm-based routing
	m.RegisterMigration(migration.Migration{
		ID:          "006_add_radius_proxy_tables",
		Description: "Add RADIUS proxy tables for realm-based routing",
		Up: func(db *gorm.DB) error {
			log.Println("Creating RADIUS proxy tables...")

			if err := migration.CreateProxyTables(db); err != nil {
				return fmt.Errorf("failed to create RADIUS proxy tables: %w", err)
			}

			log.Println("Successfully created RADIUS proxy tables")
			return nil
		},
		Down: func(db *gorm.DB) error {
			log.Println("Dropping RADIUS proxy tables...")

			if err := migration.DropProxyTables(db); err != nil {
				return fmt.Errorf("failed to drop RADIUS proxy tables: %w", err)
			}

			log.Println("Successfully dropped RADIUS proxy tables")
			return nil
		},
	})

	if *action == "up" {
		log.Println("Running migrations up...")
		if err := m.Up(); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migrations completed successfully")
	} else if *action == "down" {
		log.Println("Running migration rollback...")
		if err := m.Down(); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Rollback completed successfully")
	} else {
		log.Fatalf("Unknown action: %s. Use 'up' or 'down'", *action)
	}
}
