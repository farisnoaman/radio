package repository

import (
	"context"
	"errors"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"gorm.io/gorm"
)

// NASTemplateRepository handles database operations for NAS templates.
type NASTemplateRepository struct {
	db *gorm.DB
}

// NewNASTemplateRepository creates a new NAS template repository.
func NewNASTemplateRepository(db *gorm.DB) *NASTemplateRepository {
	return &NASTemplateRepository{db: db}
}

// Create creates a new NAS template with tenant isolation.
func (r *NASTemplateRepository) Create(ctx context.Context, template *domain.NASTemplate) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	template.TenantID = tenantID

	return r.db.WithContext(ctx).Create(template).Error
}

// GetByID retrieves a template by ID with tenant isolation.
func (r *NASTemplateRepository) GetByID(ctx context.Context, id int64) (*domain.NASTemplate, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var template domain.NASTemplate
	err = r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// ListByVendor returns all templates for a specific vendor code.
func (r *NASTemplateRepository) ListByVendor(ctx context.Context, vendorCode string) ([]*domain.NASTemplate, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var templates []*domain.NASTemplate
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND vendor_code = ?", tenantID, vendorCode).
		Order("is_default DESC, name ASC").
		Find(&templates).Error

	return templates, err
}

// GetDefaultTemplate returns the default template for a vendor.
func (r *NASTemplateRepository) GetDefaultTemplate(ctx context.Context, vendorCode string) (*domain.NASTemplate, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var template domain.NASTemplate
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND vendor_code = ? AND is_default = ?", tenantID, vendorCode, true).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// Update updates an existing template.
func (r *NASTemplateRepository) Update(ctx context.Context, template *domain.NASTemplate) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", template.ID, tenantID).
		Updates(template).Error
}

// Delete deletes a template by ID.
func (r *NASTemplateRepository) Delete(ctx context.Context, id int64) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.NASTemplate{}).Error
}
