package logging

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ArchivalManager struct {
	db     *gorm.DB
	config *config.AppConfig
}

func NewArchivalManager(db *gorm.DB, cfg *config.AppConfig) *ArchivalManager {
	return &ArchivalManager{
		db:     db,
		config: cfg,
	}
}

// ArchiveSystemLogs archives system operator logs older than days
func (m *ArchivalManager) ArchiveSystemLogs(days int) error {
	retentionDate := time.Now().AddDate(0, 0, -days)
	
	// First check if there are any logs to archive
	var count int64
	if err := m.db.Model(&domain.SysOprLog{}).Where("opt_time < ?", retentionDate).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check log count: %v", err)
	}

	if count == 0 {
		zap.S().Debug("No system logs to archive")
		return nil
	}

	archiveDir := filepath.Join(m.config.GetLogDir(), "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %v", err)
	}

	filename := fmt.Sprintf("sys_opr_log_%s.csv.gz", time.Now().Format("20060102150405"))
	archivePath := filepath.Join(archiveDir, filename)

	// Open file for writing
	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %v", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	csvWriter := csv.NewWriter(gzipWriter)
	defer csvWriter.Flush()

	// batch size
	batchSize := 1000
	offset := 0
	
	zap.S().Infof("Starting system log archival older than %v (%d logs total)", retentionDate, count)

	// Write header
	header := []string{"ID", "Operator", "Content", "IP", "Action", "Time"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	totalArchived := 0

	for {
		var logs []domain.SysOprLog
		// Find logs to archive
		if err := m.db.Where("opt_time < ?", retentionDate).Order("id ASC").Limit(batchSize).Offset(offset).Find(&logs).Error; err != nil {
			return fmt.Errorf("failed to fetch logs from database: %v", err)
		}

		if len(logs) == 0 {
			break
		}

		// Write to CSV
		for _, log := range logs {
			record := []string{
				strconv.FormatInt(log.ID, 10),
				log.OprName,
				log.OptDesc,
				log.OprIp,
				log.OptAction,
				log.OptTime.Format(time.RFC3339),
			}
			if err := csvWriter.Write(record); err != nil {
				return fmt.Errorf("failed to write CSV record: %v", err)
			}
		}

		totalArchived += len(logs)
		offset += len(logs)
	}

	if totalArchived > 0 {
		zap.S().Infof("Archived %d system logs to %s", totalArchived, archivePath)
		
		// Delete archived logs
		if err := m.db.Where("opt_time < ?", retentionDate).Delete(&domain.SysOprLog{}).Error; err != nil {
			zap.S().Errorf("Failed to delete archived logs from database: %v", err)
			return fmt.Errorf("failed to delete archived logs from database: %v", err)
		}
	}

	return nil
}
