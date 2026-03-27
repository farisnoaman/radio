# Phase 5: Provider-Managed Backup System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement provider-controlled backup system with admin override capability for data protection and disaster recovery.

**Architecture:** Provider configures backup schedule → Automated backups run → Encrypted storage → Admin can override. Backup per-provider schema with isolation.

**Tech Stack:** pg_dump, pg_restore, GPG encryption, cron jobs, local/S3 storage

---

## Task 1: Create Backup Configuration Models

**Files:**
- Create: `internal/domain/backup.go`
- Create: `internal/domain/backup_test.go`

**Step 1: Write tests for backup models**

```go
// internal/domain/backup_test.go
package domain

import (
    "testing"
)

func TestBackupConfigModel(t *testing.T) {
    config := &BackupConfig{
        TenantID:          1,
        Enabled:           true,
        Schedule:          "daily",
        RetentionDays:     30,
        MaxBackups:        10,
        IncludeUserData:   true,
        IncludeAccounting: true,
        IncludeVouchers:   true,
        IncludeNas:        true,
        StorageLocation:   "local",
        EncryptionEnabled: true,
    }

    if config.TableName() != "mst_backup_config" {
        t.Errorf("Expected table name 'mst_backup_config', got '%s'", config.TableName())
    }
}

func TestBackupRecordModel(t *testing.T) {
    record := &BackupRecord{
        TenantID:    1,
        BackupType:  "automated",
        Status:      "pending",
        SchemaName:  "provider_1",
        StartedAt:   time.Now(),
    }

    if record.TableName() != "mst_backup_record" {
        t.Errorf("Expected table name 'mst_backup_record', got '%s'", record.TableName())
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain -run TestBackup -v`
Expected: FAIL with "undefined: BackupConfig"

**Step 3: Implement backup models**

```go
// internal/domain/backup.go
package domain

import "time"

type BackupConfig struct {
    ID                 int64     `json:"id" gorm:"primaryKey"`
    TenantID           int64     `json:"tenant_id" gorm:"uniqueIndex"`

    // Automated backup settings
    Enabled            bool      `json:"enabled"`
    Schedule           string    `json:"schedule"`           // daily, weekly
    RetentionDays      int       `json:"retention_days"`     // How long to keep
    MaxBackups         int       `json:"max_backups"`        // Max number of backups

    // Backup scope
    IncludeUserData    bool      `json:"include_user_data"`
    IncludeAccounting  bool      `json:"include_accounting"`
    IncludeVouchers    bool      `json:"include_vouchers"`
    IncludeNas         bool      `json:"include_nas"`

    // Storage settings
    StorageLocation    string    `json:"storage_location"`   // local, s3, gdrive
    StoragePath        string    `json:"storage_path"`
    EncryptionEnabled  bool      `json:"encryption_enabled"`
    EncryptionKey      string    `json:"encryption_key"`     // Encrypted at rest

    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}

func (BackupConfig) TableName() string {
    return "mst_backup_config"
}

type BackupRecord struct {
    ID              int64      `json:"id" gorm:"primaryKey"`
    TenantID        int64      `json:"tenant_id" gorm:"index"`
    BackupType      string     `json:"backup_type"`       // automated, manual, admin
    Status          string     `json:"status"`            // pending, running, completed, failed

    // Backup details
    SchemaName      string     `json:"schema_name"`
    FilePath        string     `json:"file_path"`
    FileSize        int64      `json:"file_size"`         // in bytes
    Checksum        string     `json:"checksum"`

    // Statistics
    TablesCount     int        `json:"tables_count"`
    RowsCount       int64      `json:"rows_count"`
    Duration        int        `json:"duration"`          // seconds

    // Error tracking
    ErrorMessage    string     `json:"error_message"`

    // Timestamps
    StartedAt       time.Time  `json:"started_at"`
    CompletedAt     *time.Time `json:"completed_at"`
    CreatedAt       time.Time  `json:"created_at"`
}

func (BackupRecord) TableName() string {
    return "mst_backup_record"
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain -run TestBackup -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/backup.go internal/domain/backup_test.go
git commit -m "feat(domain): add backup configuration and record models"
```

---

## Task 2: Create Backup Service

**Files:**
- Create: `internal/backup/service.go`
- Create: `internal/backup/service_test.go`

**Step 1: Write tests for backup service**

```go
// internal/backup/service_test.go
package backup

import (
    "context"
    "testing"

    "github.com/talkincode/toughradius/v9/internal/domain"
)

func TestCreateBackup(t *testing.T) {
    db := setupTestDB(t)
    service := NewBackupService(db, nil)

    // Create backup config
    config := &domain.BackupConfig{
        TenantID:   1,
        Enabled:    true,
        MaxBackups: 5,
    }
    db.Create(config)

    // Create backup
    record, err := service.CreateBackup(context.Background(), 1, "manual")
    if err != nil {
        t.Fatalf("Failed to create backup: %v", err)
    }

    if record.Status != "pending" {
        t.Errorf("Expected status 'pending', got '%s'", record.Status)
    }

    if record.TenantID != 1 {
        t.Errorf("Expected tenant ID 1, got %d", record.TenantID)
    }
}

func TestQuotaExceeded(t *testing.T) {
    db := setupTestDB(t)
    service := NewBackupService(db, nil)

    // Create backup config with max 1 backup
    config := &domain.BackupConfig{
        TenantID:   1,
        Enabled:    true,
        MaxBackups: 1,
    }
    db.Create(config)

    // Create first backup
    db.Create(&domain.BackupRecord{
        TenantID: 1,
        Status:   "completed",
    })

    // Try to create second backup (should fail)
    _, err := service.CreateBackup(context.Background(), 1, "manual")
    if err != ErrBackupQuotaExceeded {
        t.Errorf("Expected ErrBackupQuotaExceeded, got %v", err)
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/backup -run TestCreate -v`
Expected: FAIL with "undefined: NewBackupService"

**Step 3: Implement backup service**

```go
// internal/backup/service.go
package backup

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
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
    if err != nil {
        return nil, err
    }

    if config != nil && config.MaxBackups > 0 {
        var count int64
        bs.db.Model(&domain.BackupRecord{}).
            Where("tenant_id = ? AND status = ?", tenantID, "completed").
            Count(&count)

        if count >= config.MaxBackups {
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

    // Generate backup filename
    timestamp := time.Now().Format("20060102-150405")
    filename := fmt.Sprintf("backup_provider_%d_%s.sql", record.TenantID, timestamp)

    // Create backup directory
    backupDir := filepath.Join("/backups", fmt.Sprintf("provider_%d", record.TenantID))
    os.MkdirAll(backupDir, 0755)

    filepath := filepath.Join(backupDir, filename)

    // Execute pg_dump
    cmd := exec.Command("pg_dump",
        "-h", dbHost,
        "-U", dbUser,
        "-d", dbName,
        "-n", record.SchemaName,
        "-f", filepath,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        record.Status = "failed"
        record.ErrorMessage = string(output)
        bs.db.Save(record)
        zap.S().Error("Backup failed",
            zap.Int64("tenant_id", record.TenantID),
            zap.String("error", string(output)))
        return
    }

    // Get file info
    fileInfo, _ := os.Stat(filepath)
    record.FileSize = fileInfo.Size()
    record.FilePath = filepath

    // Calculate checksum
    record.Checksum = bs.calculateChecksum(filepath)

    // Encrypt if enabled
    if config != nil && config.EncryptionEnabled {
        encryptedPath, err := bs.encryptor.EncryptFile(filepath, config.EncryptionKey)
        if err != nil {
            record.Status = "failed"
            record.ErrorMessage = "Encryption failed: " + err.Error()
            bs.db.Save(record)
            return
        }
        record.FilePath = encryptedPath
        os.Remove(filepath) // Remove unencrypted file
    }

    // Update record
    now := time.Now()
    record.Status = "completed"
    record.Duration = int(now.Sub(start).Seconds())
    record.CompletedAt = &now
    bs.db.Save(record)

    zap.S().Info("Backup completed",
        zap.Int64("tenant_id", record.TenantID),
        zap.String("file", filepath),
        zap.Int64("size", record.FileSize))

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

    // Verify tenant access
    if !bs.isPlatformAdmin(ctx) {
        userTenantID, _ := tenant.FromContext(ctx)
        if userTenantID != tenantID {
            return fmt.Errorf("unauthorized")
        }
    }

    // Decrypt if encrypted
    backupPath := record.FilePath
    if record.EncryptionEnabled {
        decryptedPath, err := bs.encryptor.DecryptFile(backupPath, record.EncryptionKey)
        if err != nil {
            return err
        }
        defer os.Remove(decryptedPath)
        backupPath = decryptedPath
    }

    // Execute restore
    cmd := exec.Command("psql",
        "-h", dbHost,
        "-U", dbUser,
        "-d", dbName,
        "-f", backupPath,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("restore failed: %s", string(output))
    }

    zap.S().Info("Backup restored",
        zap.Int64("tenant_id", tenantID),
        zap.Int64("backup_id", backupID))

    return nil
}

// ListBackups lists backups for a provider (tenant-isolated)
func (bs *BackupService) ListBackups(
    ctx context.Context,
    tenantID int64,
) ([]domain.BackupRecord, error) {
    // Non-admin users can only see their own backups
    if !bs.isPlatformAdmin(ctx) {
        userTenantID, _ := tenant.FromContext(ctx)
        if userTenantID != tenantID {
            return nil, fmt.Errorf("unauthorized")
        }
    }

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
    // Verify platform admin
    if !bs.isPlatformAdmin(ctx) {
        return nil, fmt.Errorf("unauthorized: platform admin required")
    }

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
        // Delete file
        os.Remove(backup.FilePath)

        // Delete record
        bs.db.Delete(&backup)

        zap.S().Info("Old backup deleted",
            zap.Int64("tenant_id", tenantID),
            zap.String("file", backup.FilePath))
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

func (bs *BackupService) calculateChecksum(filepath string) string {
    data, _ := os.ReadFile(filepath)
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:])
}

func (bs *BackupService) isPlatformAdmin(ctx context.Context) bool {
    // Check if user is platform admin
    // Implementation depends on your auth system
    return false
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/backup -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/backup/service.go internal/backup/service_test.go
git commit -m "feat(backup): add provider backup service with encryption"
```

---

## Task 3: Create Encryption Service

**Files:**
- Create: `internal/backup/encryptor.go`

**Step 1: Implement file encryption**

```go
// internal/backup/encryptor.go
package backup

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "fmt"
    "io"
    "os"
)

type Encryptor struct{}

func NewEncryptor() *Encryptor {
    return &Encryptor{}
}

// EncryptFile encrypts a file using AES-GCM
func (e *Encryptor) EncryptFile(filepath, key string) (string, error) {
    // Read file
    data, err := os.ReadFile(filepath)
    if err != nil {
        return "", err
    }

    // Create cipher block
    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    // Create GCM mode
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    // Generate nonce
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    // Encrypt and authenticate
    ciphertext := gcm.Seal(nonce, nonce, data, nil)

    // Write encrypted file
    encryptedPath := filepath + ".enc"
    if err := os.WriteFile(encryptedPath, ciphertext, 0600); err != nil {
        return "", err
    }

    return encryptedPath, nil
}

// DecryptFile decrypts an encrypted file
func (e *Encryptor) DecryptFile(filepath, key string) (string, error) {
    // Read encrypted file
    data, err := os.ReadFile(filepath)
    if err != nil {
        return "", err
    }

    // Create cipher block
    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    // Create GCM mode
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    // Extract nonce
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]

    // Decrypt
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    // Write decrypted file
    decryptedPath := filepath[:len(filepath)-4] // Remove .enc
    if err := os.WriteFile(decryptedPath, plaintext, 0600); err != nil {
        return "", err
    }

    return decryptedPath, nil
}
```

**Step 2: Commit**

```bash
git add internal/backup/encryptor.go
git commit -m "feat(backup): add AES-GCM file encryption for backups"
```

---

## Task 4: Create Backup APIs

**Files:**
- Modify: `internal/adminapi/backup.go` (enhance existing)

**Step 1: Enhance existing backup APIs**

The existing `backup.go` should be enhanced with provider isolation:

```go
// In internal/adminapi/backup.go

func ListBackups(c echo.Context) error {
    tenantID, _ := tenant.FromContext(c.Request().Context())
    backupSvc := GetBackupService(c)

    backups, err := backupSvc.ListBackups(c.Request().Context(), tenantID)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "BACKUP_LIST_ERROR", "Failed to list backups", err)
    }

    return ok(c, backups)
}

func CreateBackup(c echo.Context) error {
    tenantID, _ := tenant.FromContext(c.Request().Context())
    backupSvc := GetBackupService(c)

    record, err := backupSvc.CreateBackup(c.Request().Context(), tenantID, "manual")
    if err != nil {
        if err == backup.ErrBackupQuotaExceeded {
            return fail(c, http.StatusForbidden, "QUOTA_EXCEEDED", "Backup quota exceeded", nil)
        }
        return fail(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "Failed to create backup", err)
    }

    return ok(c, map[string]string{
        "message":  "Backup created successfully",
        "backup_id": fmt.Sprintf("%d", record.ID),
    })
}

func RestoreBackup(c echo.Context) error {
    id := c.Param("id")
    if id == "" {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Backup ID is required", nil)
    }

    backupID, _ := strconv.ParseInt(id, 10, 64)
    tenantID, _ := tenant.FromContext(c.Request().Context())

    backupSvc := GetBackupService(c)
    err := backupSvc.RestoreBackup(c.Request().Context(), tenantID, backupID)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "BACKUP_RESTORE_ERROR", "Failed to restore backup", err)
    }

    return ok(c, map[string]string{
        "message": "Backup restored successfully",
    })
}

// Admin route: Override backup
func AdminCreateBackup(c echo.Context) error {
    var req struct {
        TenantID int64  `json:"tenant_id"`
        Reason    string `json:"reason"`
    }

    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
    }

    backupSvc := GetBackupService(c)
    record, err := backupSvc.AdminOverrideBackup(c.Request().Context(), req.TenantID, req.Reason)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "Failed to create backup", err)
    }

    return ok(c, record)
}
```

**Step 2: Register new admin route**

```go
// In internal/adminapi/backup.go - registerBackupRoutes function

func registerBackupRoutes() {
    // Provider routes (already exist)
    webserver.ApiGET("/system/backup", ListBackups)
    webserver.ApiPOST("/system/backup", CreateBackup)
    webserver.ApiDELETE("/system/backup/:id", DeleteBackup)
    webserver.ApiPOST("/system/backup/:id/restore", RestoreBackup)
    webserver.ApiGET("/system/backup/:id/download", DownloadBackup)

    // New admin route
    webserver.ApiPOST("/admin/system/backup", AdminCreateBackup)
}
```

**Step 3: Commit**

```bash
git add internal/adminapi/backup.go
git commit -m "feat(adminapi): add provider-isolated backup APIs with admin override"
```

---

## Task 5: Create Automated Backup Scheduler

**Files:**
- Create: `internal/backup/scheduler.go`

**Step 1: Implement backup scheduler**

```go
// internal/backup/scheduler.go
package backup

import (
    "context"
    "time"

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
            bs.service.CreateBackup(ctx, config.TenantID, "automated")
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
```

**Step 2: Start scheduler in app initialization**

```go
// In internal/app/app.go

func (a *Application) Init(config *config.Config) {
    // ... existing initialization

    // Start backup scheduler
    backupSvc := backup.NewBackupService(a.gormDB, backup.NewEncryptor())
    backupScheduler := backup.NewBackupScheduler(backupSvc)
    go backupScheduler.Start(context.Background())
}
```

**Step 3: Commit**

```bash
git add internal/backup/scheduler.go
git commit -m "feat(backup): add automated backup scheduler"
```

---

## Success Criteria

- ✅ Backup configuration models
- ✅ Provider-controlled backup creation
- ✅ Automated backup scheduling (daily/weekly)
- ✅ AES-GCM encryption for backups
- ✅ Admin override capability
- ✅ Backup isolation per provider
- ✅ Automated cleanup of old backups
- ✅ Unit tests pass (≥80% coverage)
