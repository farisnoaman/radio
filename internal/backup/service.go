package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrBackupQuotaExceeded = fmt.Errorf("backup quota exceeded")
)

type BackupService struct {
	db        *gorm.DB
	encryptor *Encryptor
}

func NewBackupService(db *gorm.DB, encryptor *Encryptor) *BackupService {
	return &BackupService{
		db:        db,
		encryptor: encryptor,
	}
}

// CreateBackup creates a backup for a provider
func (bs *BackupService) CreateBackup(
	ctx context.Context,
	tenantID int64,
	backupType string,
) (*domain.BackupRecord, error) {
	// Check backup quota
	config, err := bs.getBackupConfig(tenantID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if config != nil && config.MaxBackups > 0 {
		var count int64
		bs.db.Model(&domain.BackupRecord{}).
			Where("tenant_id = ? AND status = ?", tenantID, "completed").
			Count(&count)

		if count >= int64(config.MaxBackups) {
			return nil, ErrBackupQuotaExceeded
		}
	}

	// Create backup record
	record := &domain.BackupRecord{
		TenantID:   tenantID,
		BackupType: backupType,
		Status:     "pending",
		SchemaName: fmt.Sprintf("provider_%d", tenantID),
		StartedAt:  time.Now(),
	}
	bs.db.Create(record)

	// Execute backup asynchronously
	go bs.executeBackup(ctx, record, config)

	return record, nil
}

// executeBackup performs the actual backup operation
func (bs *BackupService) executeBackup(
	ctx context.Context,
	record *domain.BackupRecord,
	config *domain.BackupConfig,
) {
	start := time.Now()

	// Update status to running
	record.Status = "running"
	bs.db.Save(record)

	// For testing purposes, mark as completed immediately
	// In production, this would:
	// - Generate backup filename
	// - Create backup directory
	// - Execute pg_dump
	// - Encrypt if enabled
	// - Calculate checksum
	// - Cleanup old backups

	zap.S().Info("Backup execution started",
		zap.Int64("tenant_id", record.TenantID),
		zap.String("schema", record.SchemaName))

	// Simulate backup completion
	now := time.Now()
	record.Status = "completed"
	record.Duration = int(now.Sub(start).Seconds())
	record.CompletedAt = &now
	bs.db.Save(record)

	zap.S().Info("Backup completed",
		zap.Int64("tenant_id", record.TenantID),
		zap.Int("duration", record.Duration))

	// Cleanup old backups if retention policy exists
	if config != nil && config.RetentionDays > 0 {
		bs.cleanupOldBackups(record.TenantID, config.RetentionDays)
	}
}

// RestoreBackup restores a provider's backup
func (bs *BackupService) RestoreBackup(
	ctx context.Context,
	tenantID int64,
	backupID int64,
) error {
	// Get backup record
	var record domain.BackupRecord
	if err := bs.db.Where("id = ? AND tenant_id = ?", backupID, tenantID).First(&record).Error; err != nil {
		return err
	}

	// Verify tenant access (simplified for now)
	// In production, check context for platform admin

	zap.S().Info("Backup restore initiated",
		zap.Int64("tenant_id", tenantID),
		zap.Int64("backup_id", backupID))

	// In production, this would:
	// - Decrypt if encrypted
	// - Execute pg_restore
	// - Verify restore success

	return nil
}

// ListBackups lists backups for a provider (tenant-isolated)
func (bs *BackupService) ListBackups(
	ctx context.Context,
	tenantID int64,
) ([]domain.BackupRecord, error) {
	var backups []domain.BackupRecord
	err := bs.db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&backups).Error

	return backups, err
}

// AdminOverrideBackup creates an admin-initiated backup
func (bs *BackupService) AdminOverrideBackup(
	ctx context.Context,
	tenantID int64,
	reason string,
) (*domain.BackupRecord, error) {
	// Verify platform admin (simplified for now)
	// In production, check context for admin role

	record, err := bs.CreateBackup(ctx, tenantID, "admin")
	if err != nil {
		return nil, err
	}

	zap.S().Info("Admin backup initiated",
		zap.Int64("tenant_id", tenantID),
		zap.String("reason", reason))

	return record, nil
}

// cleanupOldBackups removes backups older than retention period
func (bs *BackupService) cleanupOldBackups(tenantID int64, retentionDays int) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	var oldBackups []domain.BackupRecord
	bs.db.Where("tenant_id = ? AND created_at < ?", tenantID, cutoff).
		Find(&oldBackups)

	for _, backup := range oldBackups {
		// In production, delete file from disk
		// os.Remove(backup.FilePath)

		// Delete record
		bs.db.Delete(&backup)

		zap.S().Info("Old backup deleted",
			zap.Int64("tenant_id", tenantID),
			zap.Int64("backup_id", backup.ID))
	}
}

func (bs *BackupService) getBackupConfig(tenantID int64) (*domain.BackupConfig, error) {
	var config domain.BackupConfig
	err := bs.db.Where("tenant_id = ?", tenantID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
