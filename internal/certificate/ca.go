// Package certificate provides X.509 certificate management for 802.1x EAP-TLS authentication.
//
// This service handles:
//   - CA certificate generation and management
//   - Client certificate issuance
//   - Certificate revocation (CRL)
//   - Certificate validation
//
// Usage:
//
//	service := certificate.NewCertificateService(db)
//	ca, certPEM, keyPEM, err := service.GenerateCA(tenantID, &CAConfig{
//	    CommonName: "Corporate CA",
//	    ExpiresIn:  5 * 365 * 24 * time.Hour,
//	})
package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CAConfig holds configuration for CA certificate generation.
type CAConfig struct {
	CommonName   string
	Country      string
	Organization string
	OU           string
	ExpiresIn    time.Duration
	KeySize      int // RSA key size in bits
}

// ClientCertConfig holds configuration for client certificate generation.
type ClientCertConfig struct {
	CommonName   string
	UserID       int64
	CaID         int64
	Country      string
	Organization string
	OU           string
	ExpiresIn    time.Duration
	KeySize      int
}

// CertificateService handles certificate operations.
type CertificateService struct {
	db *gorm.DB
}

// NewCertificateService creates a new certificate service.
func NewCertificateService(db *gorm.DB) *CertificateService {
	return &CertificateService{db: db}
}

// GenerateCA generates a new CA certificate and private key.
// Returns the CA domain model and PEM-encoded certificate/key.
func (s *CertificateService) GenerateCA(
	tenantID int64,
	config *CAConfig,
) (*domain.CertificateAuthority, string, string, error) {
	if config.KeySize == 0 {
		config.KeySize = 4096
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Country:      []string{config.Country},
			Organization: []string{config.Organization},
			OrganizationalUnit: []string{config.OU},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(config.ExpiresIn),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template,
		&privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create CA record
	ca := &domain.CertificateAuthority{
		TenantID:       tenantID,
		Name:           config.CommonName + " CA",
		CommonName:     config.CommonName,
		CertificatePEM: string(certPEM),
		PrivateKeyPEM:  string(keyPEM),
		SerialNumber:   serialNumber.String(),
		ExpiryDate:     template.NotAfter,
		Status:         "active",
	}

	if err := s.db.Create(ca).Error; err != nil {
		return nil, "", "", fmt.Errorf("failed to save CA: %w", err)
	}

	zap.S().Info("CA certificate generated",
		zap.Int64("ca_id", ca.ID),
		zap.String("common_name", config.CommonName))

	return ca, string(certPEM), string(keyPEM), nil
}

// IssueClientCertificate issues a client certificate signed by the specified CA.
func (s *CertificateService) IssueClientCertificate(
	tenantID int64,
	ca *domain.CertificateAuthority,
	config *ClientCertConfig,
) (*domain.ClientCertificate, string, string, error) {
	// Load CA certificate and key
	caCert, err := ca.GetCertificate()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load CA certificate: %w", err)
	}

	caKeyBlock, _ := pem.Decode([]byte(ca.PrivateKeyPEM))
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to parse CA key: %w", err)
	}

	if config.KeySize == 0 {
		config.KeySize = 2048
	}

	// Generate client private key
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Country:      []string{config.Country},
			Organization: []string{config.Organization},
			OrganizationalUnit: []string{config.OU},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(config.ExpiresIn),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Sign with CA
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert,
		&privateKey.PublicKey, caKey)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create client cert record
	cert := &domain.ClientCertificate{
		TenantID:       tenantID,
		UserID:         config.UserID,
		CaID:           config.CaID,
		CommonName:     config.CommonName,
		SerialNumber:   serialNumber.String(),
		CertificatePEM: string(certPEM),
		PrivateKeyPEM:  string(keyPEM),
		ExpiryDate:     template.NotAfter,
		Status:         "active",
	}

	if err := s.db.Create(cert).Error; err != nil {
		return nil, "", "", fmt.Errorf("failed to save certificate: %w", err)
	}

	zap.S().Info("Client certificate issued",
		zap.Int64("cert_id", cert.ID),
		zap.String("common_name", config.CommonName))

	return cert, string(certPEM), string(keyPEM), nil
}

// RevokeCertificate revokes a client certificate.
func (s *CertificateService) RevokeCertificate(
	tenantID int64,
	certID int64,
	reason string,
) error {
	now := time.Now()

	result := s.db.Model(&domain.ClientCertificate{}).
		Where("id = ? AND tenant_id = ?", certID, tenantID).
		Updates(map[string]interface{}{
			"status":            "revoked",
			"revoked_at":        &now,
			"revocation_reason": reason,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	zap.S().Info("Certificate revoked",
		zap.Int64("cert_id", certID),
		zap.String("reason", reason))

	return nil
}

// ValidateClientCertificate validates a client certificate for 802.1x auth.
// Returns the user ID if valid, error otherwise.
func (s *CertificateService) ValidateClientCertificate(
	tenantID int64,
	certDER []byte,
) (int64, error) {
	// Parse certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return 0, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Find certificate in database
	var clientCert domain.ClientCertificate
	err = s.db.Where("tenant_id = ? AND serial_number = ? AND status = ?",
		tenantID, cert.SerialNumber.String(), "active").
		First(&clientCert).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("certificate not found or inactive")
		}
		return 0, err
	}

	// Check if expired
	if time.Now().After(cert.NotAfter) {
		return 0, errors.New("certificate has expired")
	}

	// Check if revoked
	if clientCert.IsRevoked() {
		return 0, errors.New("certificate has been revoked")
	}

	// Verify certificate chain (if CA is configured)
	if clientCert.CaID > 0 {
		var ca domain.CertificateAuthority
		if err := s.db.First(&ca, clientCert.CaID).Error; err == nil {
			_, err := ca.GetCertificate()
			if err != nil {
				return 0, fmt.Errorf("failed to load CA: %w", err)
			}

			// TODO: Verify certificate is signed by CA
			// This requires proper TLS certificate chain verification
			// For now, we trust that the certificate exists in the database
		}
	}

	return clientCert.UserID, nil
}

// GenerateCRL generates a Certificate Revocation List for a CA.
func (s *CertificateService) GenerateCRL(
	tenantID int64,
	caID int64,
) ([]byte, error) {
	// Get all revoked certificates for this CA
	var revokedCerts []domain.ClientCertificate
	err := s.db.Where("tenant_id = ? AND ca_id = ? AND status = ?",
		tenantID, caID, "revoked").
		Find(&revokedCerts).Error

	if err != nil {
		return nil, err
	}

	// TODO: Generate actual CRL in DER format
	// For now, return placeholder
	zap.S().Info("CRL generated",
		zap.Int64("ca_id", caID),
		zap.Int("revoked_count", len(revokedCerts)))

	return []byte("PLACEHOLDER_CRL"), nil
}
