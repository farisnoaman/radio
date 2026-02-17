package gorm

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
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
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&voucher).Error
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
	return r.db.WithContext(ctx).
		Model(&domain.Voucher{}).
		Where("code = ?", code).
		Updates(updates).Error
}

func (r *GormVoucherRepository) GetBatchByID(ctx context.Context, batchID int64) (*domain.VoucherBatch, error) {
	var batch domain.VoucherBatch
	err := r.db.WithContext(ctx).First(&batch, batchID).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}
