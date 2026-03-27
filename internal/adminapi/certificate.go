package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/certificate"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// caConfigPayload represents CA creation request.
type caConfigPayload struct {
	CommonName   string `json:"common_name" validate:"required,max=255"`
	Country      string `json:"country" validate:"required,size:2"`
	Organization string `json:"organization" validate:"max=200"`
	OU           string `json:"ou" validate:"max=200"`
	ExpiresDays  int    `json:"expires_days" validate:"gte=1,lte=3650"`
	KeySize      int    `json:"key_size" validate:"oneof=2048 4096"`
}

// clientCertConfigPayload represents client certificate request.
type clientCertConfigPayload struct {
	CommonName   string `json:"common_name" validate:"required,max=255"`
	UserID       int64  `json:"user_id" validate:"required"`
	CaID         int64  `json:"ca_id" validate:"required"`
	Country      string `json:"country" validate:"required,size:2"`
	Organization string `json:"organization" validate:"max=200"`
	OU           string `json:"ou" validate:"max=200"`
	ExpiresDays  int    `json:"expires_days" validate:"gte=1,lte=3650"`
	KeySize      int    `json:"key_size" validate:"oneof=2048 4096"`
}

// ListCAs retrieves all certificate authorities.
// @Summary list certificate authorities
// @Tags Certificate Management
// @Success 200 {object} Response
// @Router /api/v1/certificates/ca [get]
func ListCAs(c echo.Context) error {
	db := GetDB(c)
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	var cas []domain.CertificateAuthority
	err := db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&cas).Error

	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch CAs", err.Error())
	}

	return ok(c, cas)
}

// CreateCA creates a new certificate authority.
// @Summary create certificate authority
// @Tags Certificate Management
// @Param config body caConfigPayload true "CA configuration"
// @Success 201 {object} Response
// @Router /api/v1/certificates/ca [post]
func CreateCA(c echo.Context) error {
	var payload caConfigPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	db := GetDB(c)
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	certService := certificate.NewCertificateService(db)

	config := &certificate.CAConfig{
		CommonName:   payload.CommonName,
		Country:      payload.Country,
		Organization: payload.Organization,
		OU:           payload.OU,
		ExpiresIn:    time.Duration(payload.ExpiresDays) * 24 * time.Hour,
		KeySize:      payload.KeySize,
	}

	ca, certPEM, keyPEM, err := certService.GenerateCA(tenantID, config)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "CA_CREATE_FAILED", "Failed to generate CA", err.Error())
	}

	// Log the operation
	LogOperation(c, "create_ca", "Created CA: "+ca.Name)

	// Return CA with certificates for download
	return ok(c, map[string]interface{}{
		"ca":          ca,
		"cert_pem":    certPEM,
		"key_pem":     keyPEM,
		"download_url": "/api/v1/certificates/ca/" + strconv.FormatInt(ca.ID, 10) + "/download",
	})
}

// ListClientCertificates retrieves all client certificates.
// @Summary list client certificates
// @Tags Certificate Management
// @Param user_id query int false "Filter by user ID"
// @Success 200 {object} Response
// @Router /api/v1/certificates/client [get]
func ListClientCertificates(c echo.Context) error {
	db := GetDB(c)
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	query := db.Model(&domain.ClientCertificate{}).Where("tenant_id = ?", tenantID)

	if userID := c.QueryParam("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var certs []domain.ClientCertificate
	err := query.Order("created_at DESC").Find(&certs).Error

	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch certificates", err.Error())
	}

	return ok(c, certs)
}

// IssueClientCertificate issues a new client certificate.
// @Summary issue client certificate
// @Tags Certificate Management
// @Param config body clientCertConfigPayload true "Certificate configuration"
// @Success 201 {object} Response
// @Router /api/v1/certificates/client [post]
func IssueClientCertificate(c echo.Context) error {
	var payload clientCertConfigPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	db := GetDB(c)
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	certService := certificate.NewCertificateService(db)

	// Get CA
	var ca domain.CertificateAuthority
	err := db.Where("id = ? AND tenant_id = ? AND status = ?", payload.CaID, tenantID, "active").
		First(&ca).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "CA_NOT_FOUND", "CA not found or inactive", nil)
	}

	config := &certificate.ClientCertConfig{
		CommonName:   payload.CommonName,
		UserID:       payload.UserID,
		CaID:         payload.CaID,
		Country:      payload.Country,
		Organization: payload.Organization,
		OU:           payload.OU,
		ExpiresIn:    time.Duration(payload.ExpiresDays) * 24 * time.Hour,
		KeySize:      payload.KeySize,
	}

	cert, certPEM, keyPEM, err := certService.IssueClientCertificate(tenantID, &ca, config)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "CERT_ISSUE_FAILED", "Failed to issue certificate", err.Error())
	}

	// Log the operation
	LogOperation(c, "issue_cert", "Issued certificate: "+cert.CommonName)

	return ok(c, map[string]interface{}{
		"certificate": cert,
		"cert_pem":    certPEM,
		"key_pem":     keyPEM,
	})
}

// RevokeCertificate revokes a client certificate.
// @Summary revoke client certificate
// @Tags Certificate Management
// @Param id path int true "Certificate ID"
// @Param reason body map[string]string true "Revocation reason"
// @Success 200 {object} Response
// @Router /api/v1/certificates/client/{id}/revoke [post]
func RevokeCertificate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid certificate ID", nil)
	}

	var payload map[string]string
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	reason := payload["reason"]
	if reason == "" {
		reason = "Revoked by administrator"
	}

	db := GetDB(c)
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	certService := certificate.NewCertificateService(db)

	err = certService.RevokeCertificate(tenantID, id, reason)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "REVOKE_FAILED", "Failed to revoke certificate", err.Error())
	}

	// Log the operation
	LogOperation(c, "revoke_cert", "Revoked certificate ID: "+strconv.FormatInt(id, 10))

	return ok(c, map[string]interface{}{
		"message": "Certificate revoked successfully",
	})
}

// GetCACertificate retrieves a CA certificate for download.
// @Summary get CA certificate
// @Tags Certificate Management
// @Param id path int true "CA ID"
// @Success 200 {object} Response
// @Router /api/v1/certificates/ca/{id} [get]
func GetCACertificate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid CA ID", nil)
	}

	db := GetDB(c)
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	var ca domain.CertificateAuthority
	err = db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&ca).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "CA_NOT_FOUND", "CA not found", nil)
	}

	return ok(c, ca)
}

// GetClientCertificate retrieves a client certificate.
// @Summary get client certificate
// @Tags Certificate Management
// @Param id path int true "Certificate ID"
// @Success 200 {object} Response
// @Router /api/v1/certificates/client/{id} [get]
func GetClientCertificate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid certificate ID", nil)
	}

	db := GetDB(c)
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	var cert domain.ClientCertificate
	err = db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&cert).Error
	if err != nil {
		return fail(c, http.StatusNotFound, "CERT_NOT_FOUND", "Certificate not found", nil)
	}

	return ok(c, cert)
}

// registerCertificateRoutes registers certificate management routes.
func registerCertificateRoutes() {
	// CA management
	webserver.ApiGET("/certificates/ca", ListCAs)
	webserver.ApiPOST("/certificates/ca", CreateCA)
	webserver.ApiGET("/certificates/ca/:id", GetCACertificate)

	// Client certificate management
	webserver.ApiGET("/certificates/client", ListClientCertificates)
	webserver.ApiPOST("/certificates/client", IssueClientCertificate)
	webserver.ApiGET("/certificates/client/:id", GetClientCertificate)
	webserver.ApiPOST("/certificates/client/:id/revoke", RevokeCertificate)
}
