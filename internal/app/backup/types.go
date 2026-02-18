package backup

import (
	"time"
)

// BackupInfo represents metadata about a backup
type BackupInfo struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"` // "sqlite" or "postgres"
}

// BackupManager defines the interface for backup operations
type BackupManager interface {
	// CreateBackup creates a new backup immediately
	CreateBackup() (string, error)
	
	// ListBackups returns a list of available backups
	ListBackups() ([]BackupInfo, error)
	
	// GetBackup returns the path to a specific backup file
	GetBackup(id string) (string, error)
	
	// DeleteBackup removes a backup
	DeleteBackup(id string) error
	
	// RestoreBackup restores the database from a backup
	// Warning: This will stop the application or require a restart
	RestoreBackup(id string) error
	
	// PruneBackups deletes old backups exceeding the count limit
	PruneBackups(keep int) error
}
