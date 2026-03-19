package repository

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

type VoucherRepository struct {
	db *gorm.DB
}

func NewVoucherRepository(db *gorm.DB) *VoucherRepository {
	return &VoucherRepository{db: db}
}

func (r *VoucherRepository) GetByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var voucher domain.Voucher
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND code = ?", tenantID, code).
		First(&voucher).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *VoucherRepository) GetByID(ctx context.Context, id int64) (*domain.Voucher, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var voucher domain.Voucher
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&voucher).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *VoucherRepository) Create(ctx context.Context, voucher *domain.Voucher) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	voucher.TenantID = tenantID
	return r.db.WithContext(ctx).Create(voucher).Error
}

func (r *VoucherRepository) CreateBatch(ctx context.Context, vouchers []*domain.Voucher) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}

	for _, v := range vouchers {
		v.TenantID = tenantID
	}
	return r.db.WithContext(ctx).Create(vouchers).Error
}

func (r *VoucherRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*domain.Voucher, int64, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	var vouchers []*domain.Voucher
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Voucher{}).Where("tenant_id = ?", tenantID)

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("code LIKE ?", "%"+search+"%")
	}
	if batchID, ok := filters["batch_id"].(int64); ok && batchID > 0 {
		query = query.Where("batch_id = ?", batchID)
	}
	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	query.Count(&total)

	err = query.Offset(offset).Limit(limit).Order("id DESC").Find(&vouchers).Error
	return vouchers, total, err
}

func (r *VoucherRepository) Update(ctx context.Context, voucher *domain.Voucher) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, voucher.ID).
		Save(voucher).Error
}

func (r *VoucherRepository) Delete(ctx context.Context, id int64) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&domain.Voucher{}).Error
}

func (r *VoucherRepository) Count(ctx context.Context, status string) (int64, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Voucher{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err = query.Count(&count).Error
	return count, err
}

func (r *VoucherRepository) GetByBatch(ctx context.Context, batchID int64) ([]*domain.Voucher, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var vouchers []*domain.Voucher
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND batch_id = ?", tenantID, batchID).
		Find(&vouchers).Error
	return vouchers, err
}

type VoucherBatchRepository struct {
	db *gorm.DB
}

func NewVoucherBatchRepository(db *gorm.DB) *VoucherBatchRepository {
	return &VoucherBatchRepository{db: db}
}

func (r *VoucherBatchRepository) GetByID(ctx context.Context, id int64) (*domain.VoucherBatch, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var batch domain.VoucherBatch
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&batch).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

func (r *VoucherBatchRepository) Create(ctx context.Context, batch *domain.VoucherBatch) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	batch.TenantID = tenantID
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *VoucherBatchRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*domain.VoucherBatch, int64, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	var batches []*domain.VoucherBatch
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.VoucherBatch{}).Where("tenant_id = ?", tenantID)

	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	query.Count(&total)

	err = query.Offset(offset).Limit(limit).Order("id DESC").Find(&batches).Error
	return batches, total, err
}

func (r *VoucherBatchRepository) Update(ctx context.Context, batch *domain.VoucherBatch) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, batch.ID).
		Save(batch).Error
}

func (r *VoucherBatchRepository) Delete(ctx context.Context, id int64) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&domain.VoucherBatch{}).Error
}
