package device

import (
	"context"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SpeedTestService manages network speed tests on devices.
type SpeedTestService struct {
	db      *gorm.DB
	timeout time.Duration
}

// NewSpeedTestService creates a new speed test service.
func NewSpeedTestService(db *gorm.DB) *SpeedTestService {
	return &SpeedTestService{
		db:      db,
		timeout: 60 * time.Second,
	}
}

// RunSpeedTest executes a speed test on the specified device.
// For Mikrotik devices, this uses the built-in bandwidth test tool.
func (s *SpeedTestService) RunSpeedTest(
	ctx context.Context,
	device *domain.NetNas,
	createdBy string,
) (*domain.SpeedTestResult, error) {
	// Create result record
	result := &domain.SpeedTestResult{
		TenantID:   device.TenantID,
		NasID:      device.ID,
		Status:     "running",
		CreatedBy:  createdBy,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(result).Error; err != nil {
		return nil, fmt.Errorf("failed to create result record: %w", err)
	}

	// Execute test asynchronously
	go s.executeSpeedTest(ctx, device, result)

	return result, nil
}

// executeSpeedTest performs the actual speed test.
func (s *SpeedTestService) executeSpeedTest(
	ctx context.Context,
	device *domain.NetNas,
	result *domain.SpeedTestResult,
) {
	start := time.Now()

	zap.S().Info("Starting speed test",
		zap.Int64("nas_id", device.ID),
		zap.String("ip", device.Ipaddr),
		zap.String("vendor", device.VendorCode))

	// Execute vendor-specific test
	switch device.VendorCode {
	case "mikrotik":
		s.runMikrotikSpeedTest(ctx, device, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("unsupported vendor: %s", device.VendorCode)
		s.db.Save(result)
	}

	duration := int(time.Since(start).Seconds())
	result.TestDurationSec = duration
	s.db.Save(result)

	zap.S().Info("Speed test completed",
		zap.Int64("nas_id", device.ID),
		zap.Float64("upload_mbps", result.UploadMbps),
		zap.Float64("download_mbps", result.DownloadMbps))
}

// runMikrotikSpeedTest executes Mikrotik's built-in bandwidth test.
func (s *SpeedTestService) runMikrotikSpeedTest(
	ctx context.Context,
	device *domain.NetNas,
	result *domain.SpeedTestResult,
) {
	// TODO: Implement actual Mikrotik bandwidth test via RouterOS API
	// Command: /tool bandwidth-test [find] test-server=1.2.3.4

	// For now, set mock results
	result.UploadMbps = 95.5
	result.DownloadMbps = 485.2
	result.LatencyMs = 12.3
	result.JitterMs = 2.1
	result.PacketLoss = 0.0
	result.Status = "completed"

	s.db.Save(result)
}

// GetTestHistory returns speed test history for a device.
func (s *SpeedTestService) GetTestHistory(
	ctx context.Context,
	nasID int64,
	limit int,
) ([]*domain.SpeedTestResult, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var results []*domain.SpeedTestResult
	err = s.db.Where("tenant_id = ? AND nas_id = ?", tenantID, nasID).
		Order("created_at DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}
