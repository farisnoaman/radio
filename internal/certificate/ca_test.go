package certificate

import (
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCertificateTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&domain.CertificateAuthority{}, &domain.ClientCertificate{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestCertificateService_GenerateCA_ShouldSucceed(t *testing.T) {
	db := setupCertificateTestDB(t)
	service := NewCertificateService(db)

	ca, certPEM, keyPEM, err := service.GenerateCA(1, &CAConfig{
		CommonName: "Test CA",
		Country:    "US",
		ExpiresIn:  365 * 24 * time.Hour,
	})

	if err != nil {
		t.Fatalf("GenerateCA failed: %v", err)
	}

	if ca == nil {
		t.Fatal("expected CA to be created")
	}

	if certPEM == "" {
		t.Fatal("expected cert PEM")
	}

	if keyPEM == "" {
		t.Fatal("expected key PEM")
	}
}
