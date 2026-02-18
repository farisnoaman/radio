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
	archiveDir := filepath.Join(m.config.GetLogDir(), "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("sys_opr_log_%s.csv.gz", time.Now().Format("20060102150405"))
	filepath := filepath.Join(archiveDir, filename)

	// Open file for writing
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	csvWriter := csv.NewWriter(gzipWriter)
	defer csvWriter.Flush()

	// batch size
	batchSize := 1000
	offset := 0
	
	zap.S().Infof("Starting system log archival older than %v", retentionDate)

	// Write header
	header := []string{"ID", "Operator", "Content", "IP", "Action", "Time"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	totalArchived := 0

	for {
		var logs []domain.SysOprLog
		// Find logs to archive
		if err := m.db.Where("opt_time < ?", retentionDate).Order("id ASC").Limit(batchSize).Offset(offset).Find(&logs).Error; err != nil {
			return err
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
				return err
			}
		}

		totalArchived += len(logs)
		offset += len(logs)
	}

	if totalArchived > 0 {
		zap.S().Infof("Archived %d system logs to %s", totalArchived, filepath)
		
		// Delete archived logs
		// For safety, we only delete logs up to the last ID we processed, or simply by time again.
		// Deleting by time is safer in concurrent environments, slightly risks deleting logs inserted *during* archival 
		// if timestamps match exactly, but highly unlikely for historical logs.
		if err := m.db.Where("opt_time < ?", retentionDate).Delete(&domain.SysOprLog{}).Error; err != nil {
			zap.S().Errorf("Failed to delete archived logs: %v", err)
			return err
		}
	} else {
		// If no logs found, remove the empty file
		os.Remove(filepath)
	}

	return nil
}
