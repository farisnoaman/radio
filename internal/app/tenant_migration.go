package app

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (a *Application) MigrateTenantSupport() error {
	db := a.gormDB
	
	// Check if mst_provider table exists
	hasProviderTable := db.Migrator().HasTable(&domain.Provider{})
	if !hasProviderTable {
		if err := db.Migrator().AutoMigrate(&domain.Provider{}); err != nil {
			zap.S().Errorf("failed to create provider table: %v", err)
			return err
		}
		
		// Create default provider
		defaultProvider := &domain.Provider{
			Code:     "default",
			Name:     "Default Provider",
			Status:   "active",
			MaxUsers: 1000,
			MaxNas:   100,
		}
		if err := db.Create(defaultProvider).Error; err != nil {
			zap.S().Errorf("failed to create default provider: %v", err)
			return err
		}
		zap.S().Info("Created default provider with ID: 1")
	}

	// Add tenant_id column to existing tables if they don't have it
	if err := addTenantIDToTable(db, "radius_user"); err != nil {
		zap.S().Warnf("radius_user tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "radius_online"); err != nil {
		zap.S().Warnf("radius_online tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "radius_accounting"); err != nil {
		zap.S().Warnf("radius_accounting tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "radius_profile"); err != nil {
		zap.S().Warnf("radius_profile tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "net_nas"); err != nil {
		zap.S().Warnf("net_nas tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "net_node"); err != nil {
		zap.S().Warnf("net_node tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "product"); err != nil {
		zap.S().Warnf("product tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "voucher_batch"); err != nil {
		zap.S().Warnf("voucher_batch tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "voucher"); err != nil {
		zap.S().Warnf("voucher tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "voucher_topup"); err != nil {
		zap.S().Warnf("voucher_topup tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "voucher_subscription"); err != nil {
		zap.S().Warnf("voucher_subscription tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "voucher_bundle"); err != nil {
		zap.S().Warnf("voucher_bundle tenant_id migration: %v", err)
	}
	if err := addTenantIDToTable(db, "sys_opr"); err != nil {
		zap.S().Warnf("sys_opr tenant_id migration: %v", err)
	}

	return nil
}

func addTenantIDToTable(db *gorm.DB, tableName string) error {
	hasColumn, err := hasColumn(db, tableName, "tenant_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		// Add tenant_id column with default value of 1
		// For SQLite, we need to use raw SQL
		if db.Dialector.Name() == "sqlite" {
			if err := db.Exec("ALTER TABLE " + tableName + " ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1").Error; err != nil {
				return err
			}
		} else {
			// For PostgreSQL
			if err := db.Exec("ALTER TABLE " + tableName + " ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1").Error; err != nil {
				return err
			}
			// Add index
			if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_" + tableName + "_tenant_id ON " + tableName + "(tenant_id)").Error; err != nil {
				zap.S().Warnf("failed to create index for %s: %v", tableName, err)
			}
		}
		zap.S().Infof("Added tenant_id column to %s table", tableName)
	}

	return nil
}

func hasColumn(db *gorm.DB, tableName, columnName string) (bool, error) {
	var count int64
	if db.Dialector.Name() == "sqlite" {
		err := db.Raw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?", tableName, columnName).Scan(&count).Error
		if err != nil {
			return false, err
		}
	} else {
		err := db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?", tableName, columnName).Scan(&count).Error
		if err != nil {
			return false, err
		}
	}
	return count > 0, nil
}
