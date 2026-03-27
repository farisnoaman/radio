package gorm

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

// GormVoucherRepository is the GORM implementation of the voucher repository
type GormVoucherRepository struct {
	db *gorm.DB
}

// NewGormVoucherRepository creates a voucher repository instance
func NewGormVoucherRepository(db *gorm.DB) repository.VoucherRepository {
	return &GormVoucherRepository{db: db}
}

func (r *GormVoucherRepository) GetByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	var voucher domain.Voucher
	query := r.db.WithContext(ctx).Where("code = ?", code)
	if tenantID, err := tenant.FromContext(ctx); err == nil && tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&voucher).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *GormVoucherRepository) UpdateFirstUsedAt(ctx context.Context, code string, firstUsedAt, expireTime time.Time) error {
	updates := map[string]interface{}{
		"first_used_at": firstUsedAt,
		"expire_time":   expireTime,
	}
	query := r.db.WithContext(ctx).Model(&domain.Voucher{}).Where("code = ?", code)
	if tenantID, err := tenant.FromContext(ctx); err == nil && tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return query.Updates(updates).Error
}

func (r *GormVoucherRepository) GetBatchByID(ctx context.Context, batchID int64) (*domain.VoucherBatch, error) {
	var batch domain.VoucherBatch
	query := r.db.WithContext(ctx).Where("id = ?", batchID)
	if tenantID, err := tenant.FromContext(ctx); err == nil && tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&batch).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}
