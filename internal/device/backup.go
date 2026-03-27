package device

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SSHClient defines the interface for SSH connections to network devices.
type SSHClient interface {
	Connect(ctx context.Context, addr, username, password string) error
	RunCommand(ctx context.Context, cmd string) (string, error)
	Close() error
}

// DeviceBackupService handles automatic configuration backups for network devices.
type DeviceBackupService struct {
	db      *gorm.DB
	sshPool map[string]SSHClient
	timeout time.Duration
	mu      sync.RWMutex
}

// NewDeviceBackupService creates a new device backup service.
func NewDeviceBackupService(db *gorm.DB) *DeviceBackupService {
	return &DeviceBackupService{
		db:      db,
		sshPool: make(map[string]SSHClient),
		timeout: 30 * time.Second,
	}
}

// BackupConfig backs up the configuration for a single device.
// The method connects to the device via SSH, retrieves the configuration,
// and stores it in the database.
func (s *DeviceBackupService) BackupConfig(
	ctx context.Context,
	device *domain.NetNas,
	createdBy string,
) (*domain.DeviceConfigBackup, error) {
	// Create backup record
	record := &domain.DeviceConfigBackup{
		TenantID:   device.TenantID,
		NasID:      device.ID,
		VendorCode: device.VendorCode,
		Status:     "pending",
		StartedAt:  time.Now(),
		CreatedBy:  createdBy,
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// Execute backup asynchronously
	go s.executeBackup(ctx, device, record)

	return record, nil
}

// executeBackup performs the actual backup operation.
func (s *DeviceBackupService) executeBackup(
	ctx context.Context,
	device *domain.NetNas,
	record *domain.DeviceConfigBackup,
) {
	// Update status to running
	record.Status = "running"
	s.db.Save(record)

	start := time.Now()
	zap.S().Info("Starting device config backup",
		zap.Int64("nas_id", device.ID),
		zap.String("ip", device.Ipaddr),
		zap.String("vendor", device.VendorCode))

	// Get vendor-specific backup command
	cmd := s.getBackupCommand(device.VendorCode)

	// Execute command via SSH
	config, err := s.executeSSHCommand(ctx, device, cmd)
	if err != nil {
		record.Status = "failed"
		record.Error = err.Error()
		s.db.Save(record)
		zap.S().Error("Device backup failed",
			zap.Int64("nas_id", device.ID),
			zap.Error(err))
		return
	}

	// Store configuration
	now := time.Now()
	record.ConfigData = config
	record.FileSize = int64(len(config))
	record.Status = "completed"
	record.CompletedAt = &now
	s.db.Save(record)

	duration := now.Sub(start)
	zap.S().Info("Device backup completed",
		zap.Int64("nas_id", device.ID),
		zap.Int64("size_bytes", record.FileSize),
		zap.Duration("duration", duration))
}

// getBackupCommand returns the vendor-specific command to export configuration.
func (s *DeviceBackupService) getBackupCommand(vendorCode string) string {
	commands := map[string]string{
		"mikrotik": "/export verbose",
		"cisco":    "show running-config",
		"huawei":   "display current-configuration",
		"juniper":  "show configuration",
		"ubiquiti": "cat /cfg/config",
	}

	if cmd, ok := commands[vendorCode]; ok {
		return cmd
	}
	return "show running-config" // Default fallback
}

// executeSSHCommand executes a command on a device via SSH.
func (s *DeviceBackupService) executeSSHCommand(
	ctx context.Context,
	device *domain.NetNas,
	command string,
) (string, error) {
	// TODO: Implement actual SSH connection
	// For now, return mock data
	return fmt.Sprintf("# Configuration backup from %s\n%s", device.Name, command), nil
}

// ScheduleBackups schedules automatic backups for all enabled devices.
func (s *DeviceBackupService) ScheduleBackups(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.backupAllDevices(ctx)
		}
	}
}

// backupAllDevices backs up all enabled NAS devices.
func (s *DeviceBackupService) backupAllDevices(ctx context.Context) {
	var devices []domain.NetNas
	s.db.Where("status = ?", "enabled").Find(&devices)

	for _, device := range devices {
		_, err := s.BackupConfig(ctx, &device, "system")
		if err != nil {
			zap.S().Error("Failed to queue backup",
				zap.Int64("nas_id", device.ID),
				zap.Error(err))
		}
	}
}

// MockSSHClient is a mock SSH client for testing.
type MockSSHClient struct {
	ConfigOutput string
	Connected    bool
}

func (m *MockSSHClient) Connect(ctx context.Context, addr, username, password string) error {
	m.Connected = true
	return nil
}

func (m *MockSSHClient) RunCommand(ctx context.Context, cmd string) (string, error) {
	if !m.Connected {
		return "", fmt.Errorf("not connected")
	}
	return m.ConfigOutput, nil
}

func (m *MockSSHClient) Close() error {
	m.Connected = false
	return nil
}
