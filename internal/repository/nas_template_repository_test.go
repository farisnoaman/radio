package repository

import (
	"context"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
)

func TestNASTemplateRepository_CreateTemplate_ShouldSucceed(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)

	// Auto-migrate the NASTemplate table
	if err := db.AutoMigrate(&domain.NASTemplate{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	repo := NewNASTemplateRepository(db)
	ctx := tenant.WithTenantID(context.Background(), 1)

	template := &domain.NASTemplate{
		VendorCode: "huawei",
		Name:       "Test Template",
		Attributes: []domain.TemplateAttribute{
			{AttrName: "Framed-IP-Address", VendorAttr: "Huawei-IP-Address", ValueType: "ipaddr"},
		},
	}

	err := repo.Create(ctx, template)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	if template.ID == 0 {
		t.Fatal("expected template ID to be set")
	}
}
