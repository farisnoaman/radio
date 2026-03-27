package gorm

import (
	"context"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

// GormUserRepository is the GORM implementation of the user repository
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a user repository instance
func NewGormUserRepository(db *gorm.DB) repository.UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) GetByUsername(ctx context.Context, username string) (*domain.RadiusUser, error) {
	var user domain.RadiusUser
	query := r.db.WithContext(ctx).Where("username = ?", username)
	if tenantID, err := tenant.FromContext(ctx); err == nil && tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) GetByMacAddr(ctx context.Context, macAddr string) (*domain.RadiusUser, error) {
	var user domain.RadiusUser
	query := r.db.WithContext(ctx).Where("mac_addr = ?", macAddr)
	if tenantID, err := tenant.FromContext(ctx); err == nil && tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) UpdateMacAddr(ctx context.Context, username, macAddr string) error {
	tenantID, _ := tenant.FromContext(ctx)
	query := r.db.WithContext(ctx).Model(&domain.RadiusUser{}).Where("username = ?", username)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return query.Update("mac_addr", macAddr).Error
}

func (r *GormUserRepository) UpdateVlanId(ctx context.Context, username string, vlanId1, vlanId2 int) error {
	updates := map[string]interface{}{
		"vlanid1": vlanId1,
		"vlanid2": vlanId2,
	}
	tenantID, _ := tenant.FromContext(ctx)
	query := r.db.WithContext(ctx).Model(&domain.RadiusUser{}).Where("username = ?", username)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return query.Updates(updates).Error
}

func (r *GormUserRepository) UpdateLastOnline(ctx context.Context, username string) error {
	tenantID, _ := tenant.FromContext(ctx)
	query := r.db.WithContext(ctx).Model(&domain.RadiusUser{}).Where("username = ?", username)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return query.Update("last_online", time.Now()).Error
}

func (r *GormUserRepository) UpdateField(ctx context.Context, username string, field string, value interface{}) error {
	tenantID, _ := tenant.FromContext(ctx)
	query := r.db.WithContext(ctx).Model(&domain.RadiusUser{}).Where("username = ?", username)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return query.Update(field, value).Error
}
