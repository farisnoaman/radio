package backup

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
)

type BackupScheduler struct {
	service *BackupService
}

func NewBackupScheduler(service *BackupService) *BackupScheduler {
	return &BackupScheduler{service: service}
}

// Start begins the automated backup scheduler
func (bs *BackupScheduler) Start(ctx context.Context) {
	// Check every hour for backups due
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bs.runDueBackups(ctx)
		}
	}
}

func (bs *BackupScheduler) runDueBackups(ctx context.Context) {
	// Get all enabled backup configs
	var configs []domain.BackupConfig
	bs.service.db.Where("enabled = ?", true).Find(&configs)

	for _, config := range configs {
		// Check if backup is due
		if bs.isBackupDue(&config) {
			_, err := bs.service.CreateBackup(ctx, config.TenantID, "automated")
			if err != nil {
				zap.S().Error("Automated backup failed",
					zap.Int64("tenant_id", config.TenantID),
					zap.Error(err))
			} else {
				zap.S().Info("Automated backup created",
					zap.Int64("tenant_id", config.TenantID),
					zap.String("schedule", config.Schedule))
			}
		}
	}
}

func (bs *BackupScheduler) isBackupDue(config *domain.BackupConfig) bool {
	// Get last backup for this provider
	var lastBackup domain.BackupRecord
	err := bs.service.db.Where("tenant_id = ? AND backup_type = ?", config.TenantID, "automated").
		Order("created_at DESC").
		First(&lastBackup).Error

	if err != nil {
		// No previous backup, run now
		return true
	}

	// Check if enough time has passed based on schedule
	now := time.Now()
	lastBackupTime := lastBackup.CreatedAt

	switch config.Schedule {
	case "daily":
		return now.Sub(lastBackupTime) >= 24*time.Hour
	case "weekly":
		return now.Sub(lastBackupTime) >= 7*24*time.Hour
	default:
		return false
	}
}
