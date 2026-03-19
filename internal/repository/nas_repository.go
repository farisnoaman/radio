package repository

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

type NasRepository struct {
	db *gorm.DB
}

func NewNasRepository(db *gorm.DB) *NasRepository {
	return &NasRepository{db: db}
}

func (r *NasRepository) GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		var nas domain.NetNas
		err = r.db.WithContext(ctx).
			Where("ipaddr = ? OR identifier = ?", ip, identifier).
			First(&nas).Error
		return &nas, err
	}

	var nas domain.NetNas
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND (ipaddr = ? OR identifier = ?)", tenantID, ip, identifier).
		First(&nas).Error
	return &nas, err
}

func (r *NasRepository) GetByID(ctx context.Context, id int64) (*domain.NetNas, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var nas domain.NetNas
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&nas).Error
	return &nas, err
}

func (r *NasRepository) GetByTenantAndIP(ctx context.Context, tenantID int64, ip, identifier string) (*domain.NetNas, int64, error) {
	var nas domain.NetNas

	err := r.db.WithContext(ctx).
		Where("ipaddr = ? OR identifier = ?", ip, identifier).
		First(&nas).Error
	if err != nil {
		return nil, 0, err
	}
	return &nas, nas.TenantID, nil
}

func (r *NasRepository) Create(ctx context.Context, nas *domain.NetNas) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	nas.TenantID = tenantID
	return r.db.WithContext(ctx).Create(nas).Error
}

func (r *NasRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*domain.NetNas, int64, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	var nasList []*domain.NetNas
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.NetNas{}).Where("tenant_id = ?", tenantID)

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("name LIKE ? OR ipaddr LIKE ? OR identifier LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if nodeID, ok := filters["node_id"].(int64); ok && nodeID > 0 {
		query = query.Where("node_id = ?", nodeID)
	}

	query.Count(&total)

	err = query.Offset(offset).Limit(limit).Order("id DESC").Find(&nasList).Error
	return nasList, total, err
}

func (r *NasRepository) Update(ctx context.Context, nas *domain.NetNas) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, nas.ID).
		Save(nas).Error
}

func (r *NasRepository) Delete(ctx context.Context, id int64) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&domain.NetNas{}).Error
}

func (r *NasRepository) Count(ctx context.Context, status string) (int64, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	query := r.db.WithContext(ctx).Model(&domain.NetNas{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err = query.Count(&count).Error
	return count, err
}