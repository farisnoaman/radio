package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v9/config"
	"go.uber.org/zap"
)

// restoreMetadata tracks restore timestamps for backups
type restoreMetadata struct {
	RestoredAt map[string]string `json:"restored_at"` // backup_id -> timestamp
}

type LocalBackupManager struct {
	cfg            *config.AppConfig
	backupDir      string
	gdrive         *GoogleDriveProvider
	metadataFile   string
	restoreMeta    restoreMetadata
}

func NewLocalBackupManager(cfg *config.AppConfig) *LocalBackupManager {
	mgr := &LocalBackupManager{
		cfg:          cfg,
		backupDir:    cfg.GetBackupDir(),
		metadataFile: filepath.Join(cfg.GetBackupDir(), ".restore_metadata.json"),
		restoreMeta:  restoreMetadata{RestoredAt: make(map[string]string)},
	}

	// Load existing restore metadata
	mgr.loadRestoreMetadata()

	if cfg.Backup.GoogleDrive.Enabled {
		provider, err := NewGoogleDriveProvider(cfg.Backup.GoogleDrive.ServiceAccountJSON, cfg.Backup.GoogleDrive.FolderID)
		if err != nil {
			zap.S().Errorf("Failed to initialize Google Drive provider: %v", err)
		} else {
			mgr.gdrive = provider
		}
	}

	return mgr
}

// loadRestoreMetadata loads the restore timestamps from the metadata file
func (m *LocalBackupManager) loadRestoreMetadata() {
	data, err := os.ReadFile(m.metadataFile)
	if err != nil {
		zap.S().Debugf("No existing restore metadata file found at %s: %v", m.metadataFile, err)
		return // File doesn't exist or can't be read, use empty metadata
	}
	err = json.Unmarshal(data, &m.restoreMeta)
	if err != nil {
		zap.S().Warnf("Failed to parse restore metadata: %v", err)
	}
	if m.restoreMeta.RestoredAt == nil {
		m.restoreMeta.RestoredAt = make(map[string]string)
	}
	zap.S().Debugf("Loaded restore metadata with %d entries", len(m.restoreMeta.RestoredAt))
}

// saveRestoreMetadata saves the restore timestamps to the metadata file
func (m *LocalBackupManager) saveRestoreMetadata() {
	// Ensure backup directory exists
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		zap.S().Errorf("Failed to create backup directory %s: %v", m.backupDir, err)
		return
	}

	data, err := json.MarshalIndent(m.restoreMeta, "", "  ")
	if err != nil {
		zap.S().Errorf("Failed to marshal restore metadata: %v", err)
		return
	}
	err = os.WriteFile(m.metadataFile, data, 0644)
	if err != nil {
		zap.S().Errorf("Failed to save restore metadata to %s: %v", m.metadataFile, err)
	} else {
		zap.S().Infof("Saved restore metadata to %s", m.metadataFile)
	}
}

func (m *LocalBackupManager) CreateBackup() (string, error) {
	timestamp := time.Now().Format("20060102150405")
	var filename string
	var err error

	if m.cfg.Database.Type == "sqlite" {
		filename = fmt.Sprintf("toughradius_%s.db", timestamp)
		err = m.backupSQLite(filename)
	} else if m.cfg.Database.Type == "postgres" {
		filename = fmt.Sprintf("toughradius_%s.sql", timestamp)
		err = m.backupPostgres(filename)
	} else {
		return "", fmt.Errorf("unsupported database type: %s", m.cfg.Database.Type)
	}

	if err != nil {
		return "", err
	}

	// Upload to Google Drive if enabled
	if m.gdrive != nil {
		go func() {
			fullPath := filepath.Join(m.backupDir, filename)
			if err := m.gdrive.Upload(fullPath); err != nil {
				zap.S().Errorf("Failed to upload backup to Google Drive: %v", err)
			} else {
				zap.S().Infof("Uploaded backup to Google Drive: %s", filename)
			}
		}()
	}

	return filename, nil
}

func (m *LocalBackupManager) backupSQLite(filename string) error {
	srcPath := filepath.Join(m.cfg.GetDataDir(), m.cfg.Database.Name)
	dstPath := filepath.Join(m.backupDir, filename)

	// Simple file copy for now. 
	// In production with high concurrency, VACUUM INTO is better but requires SQLite 3.27+
	// For robust hot backup without VACUUM INTO, we'd need to use the connection to lock it.
	// Given we can't easily access the raw sql.DB connection here without passing it in,
	// we'll try a file copy. It might be slightly fuzzy if write happens, but acceptable for basic use.
	// Better approach: Use SQLite Online Backup API if possible, or VACUUM INTO.
	
	// Check if source exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("source database not found: %s", srcPath)
	}

	source, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func (m *LocalBackupManager) backupPostgres(filename string) error {
	dstPath := filepath.Join(m.backupDir, filename)
	
	// Use pg_dump
	// Requires pg_dump to be installed and in PATH
	cmd := exec.Command("pg_dump", 
		"-h", m.cfg.Database.Host,
		"-p", fmt.Sprintf("%d", m.cfg.Database.Port),
		"-U", m.cfg.Database.User,
		"-F", "c", // Custom format (compressed)
		"-b",      // Include large objects
		"-v",      // Verbose
		"-f", dstPath,
		m.cfg.Database.Name,
	)
	
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", m.cfg.Database.Passwd))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		zap.S().Errorf("pg_dump failed: %s", string(output))
		return fmt.Errorf("pg_dump failed: %v", err)
	}
	
	return nil
}

func (m *LocalBackupManager) ListBackups() ([]BackupInfo, error) {
	// Reload metadata to get latest restore timestamps
	m.loadRestoreMetadata()

	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		// Filter relevant files
		if !strings.HasPrefix(name, "toughradius_") {
			continue
		}

		// Parse timestamp from filename
		// Format: toughradius_20230101120000.db / .sql
		parts := strings.Split(strings.TrimSuffix(name, filepath.Ext(name)), "_")
		if len(parts) < 2 {
			continue
		}
		
		tsStr := parts[len(parts)-1]
		createdAt, err := time.Parse("20060102150405", tsStr)
		if err != nil {
			// Fallback to file mod time
			createdAt = info.ModTime()
		}

		dbType := "sqlite"
		if strings.HasSuffix(name, ".sql") {
			dbType = "postgres"
		}

		// Check if this backup has been restored
		var restoredAt *time.Time
		if restoredStr, ok := m.restoreMeta.RestoredAt[name]; ok {
			t, err := time.Parse(time.RFC3339, restoredStr)
			if err == nil {
				restoredAt = &t
				zap.S().Debugf("Backup %s has restored_at: %s", name, restoredStr)
			}
		}

		backups = append(backups, BackupInfo{
			ID:         name,
			FileName:   name,
			Size:       info.Size(),
			CreatedAt:  createdAt,
			RestoredAt: restoredAt,
			Type:       dbType,
		})
	}

	// Sort by CreatedAt desc
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

func (m *LocalBackupManager) GetBackup(id string) (string, error) {
	path := filepath.Join(m.backupDir, id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("backup not found")
	}
	return path, nil
}

func (m *LocalBackupManager) DeleteBackup(id string) error {
	path := filepath.Join(m.backupDir, id)
	return os.Remove(path)
}

func (m *LocalBackupManager) RestoreBackup(id string) error {
	path := filepath.Join(m.backupDir, id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("backup not found")
	}

	zap.S().Infof("Starting restore of backup: %s from path: %s", id, path)

	var restoreErr error
	if m.cfg.Database.Type == "sqlite" {
		restoreErr = m.restoreSQLite(path)
	} else if m.cfg.Database.Type == "postgres" {
		restoreErr = m.restorePostgres(path)
	} else {
		return fmt.Errorf("unsupported database type")
	}

	if restoreErr != nil {
		zap.S().Errorf("Failed to restore backup %s: %v", id, restoreErr)
		return restoreErr
	}

	// Record the restore timestamp
	restoreTime := time.Now().Format(time.RFC3339)
	m.restoreMeta.RestoredAt[id] = restoreTime
	zap.S().Infof("Recording restore timestamp for %s: %s", id, restoreTime)
	m.saveRestoreMetadata()

	zap.S().Infof("Successfully restored backup: %s", id)
	return nil
}

func (m *LocalBackupManager) restoreSQLite(backupPath string) error {
	dstPath := filepath.Join(m.cfg.GetDataDir(), m.cfg.Database.Name)
	
	// Create a temporary backup of current state just in case
	tmpBackup := dstPath + ".pre_restore"
	_ = copyFile(dstPath, tmpBackup)
	
	// Overwrite DB file
	err := copyFile(backupPath, dstPath)
	if err != nil {
		// Try to restore old state
		_ = copyFile(tmpBackup, dstPath)
		return err
	}
	
	_ = os.Remove(tmpBackup)
	return nil
}

func (m *LocalBackupManager) restorePostgres(backupPath string) error {
	// Use pg_restore
	cmd := exec.Command("pg_restore",
		"-h", m.cfg.Database.Host,
		"-p", fmt.Sprintf("%d", m.cfg.Database.Port),
		"-U", m.cfg.Database.User,
		"-d", m.cfg.Database.Name,
		"-c", // Clean (drop) database objects before creating
		"-v", // Verbose
		backupPath,
	)
	
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", m.cfg.Database.Passwd))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		zap.S().Errorf("pg_restore failed: %s", string(output))
		return fmt.Errorf("pg_restore failed: %v", err)
	}
	
	return nil
}

func (m *LocalBackupManager) PruneBackups(keep int) error {
	backups, err := m.ListBackups()
	if err != nil {
		return err
	}

	if len(backups) <= keep {
		return nil
	}

	// Delete excess backups (backups are sorted desc, so delete from keep index onwards)
	for i := keep; i < len(backups); i++ {
		err := m.DeleteBackup(backups[i].ID)
		if err != nil {
			zap.S().Warnf("failed to delete old backup %s: %v", backups[i].ID, err)
		} else {
			zap.S().Infof("pruned old backup: %s", backups[i].ID)
		}
	}
	
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
