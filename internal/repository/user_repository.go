package repository

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.RadiusUser, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user domain.RadiusUser
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND username = ?", tenantID, username).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.RadiusUser, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user domain.RadiusUser
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.RadiusUser) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	user.TenantID = tenantID
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*domain.RadiusUser, int64, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	var users []*domain.RadiusUser
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.RadiusUser{}).Where("tenant_id = ?", tenantID)

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("username LIKE ?", "%"+search+"%")
	}
	if nodeID, ok := filters["node_id"].(int64); ok && nodeID > 0 {
		query = query.Where("node_id = ?", nodeID)
	}
	if profileID, ok := filters["profile_id"].(int64); ok && profileID > 0 {
		query = query.Where("profile_id = ?", profileID)
	}

	query.Count(&total)

	err = query.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (r *UserRepository) Update(ctx context.Context, user *domain.RadiusUser) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, user.ID).
		Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&domain.RadiusUser{}).Error
}

func (r *UserRepository) Count(ctx context.Context, status string) (int64, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	query := r.db.WithContext(ctx).Model(&domain.RadiusUser{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err = query.Count(&count).Error
	return count, err
}

func (r *UserRepository) GetByNode(ctx context.Context, nodeID int64) ([]*domain.RadiusUser, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var users []*domain.RadiusUser
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND node_id = ?", tenantID, nodeID).
		Find(&users).Error
	return users, err
}

func (r *UserRepository) GetByProfile(ctx context.Context, profileID int64) ([]*domain.RadiusUser, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var users []*domain.RadiusUser
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND profile_id = ?", tenantID, profileID).
		Find(&users).Error
	return users, err
}