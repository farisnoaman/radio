package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"

	"github.com/talkincode/toughradius/v9/internal/certificate"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	EAPMethodTLS = "eap-tls"
)

// TLSHandler handles EAP-TLS authentication (RFC 5216).
//
// EAP-TLS uses X.509 certificates for mutual authentication:
//   - Client presents certificate to prove identity
//   - Server validates certificate against CA
//   - Server presents certificate for client validation
//
// This is the most secure EAP method, suitable for enterprise WiFi.
type TLSHandler struct {
	certService *certificate.CertificateService
	db          *gorm.DB
}

// NewTLSHandler creates a new EAP-TLS handler.
func NewTLSHandler(db *gorm.DB) *TLSHandler {
	return &TLSHandler{
		certService: certificate.NewCertificateService(db),
		db:          db,
	}
}

// Name returns the handler name.
func (h *TLSHandler) Name() string {
	return EAPMethodTLS
}

// EAPType returns the EAP type code (13 for TLS).
func (h *TLSHandler) EAPType() uint8 {
	return eap.TypeTLS
}

// CanHandle checks whether this handler can process the EAP message.
func (h *TLSHandler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeTLS
}

// HandleIdentity handles EAP-Response/Identity and starts TLS handshake.
func (h *TLSHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	// For EAP-TLS, we start the TLS handshake
	// In a full implementation, this would use crypto/tls with a custom connection

	// Generate state for tracking the TLS session
	stateID := common.UUID()
	username := rfc2865.UserName_GetString(ctx.Request.Packet)

	state := &eap.EAPState{
		Username:  username,
		StateID:   stateID,
		Method:    EAPMethodTLS,
		Success:   false,
		Data:      make(map[string]interface{}),
	}

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}

	// Send TLS start request (empty EAP-TLS message)
	eapData := h.buildTLSStartRequest(ctx.EAPMessage.Identifier)

	// Create RADIUS Access-Challenge response
	response := ctx.Request.Response(radius.CodeAccessChallenge)

	// Set the State attribute
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck

	// Set the EAP-Message and Message-Authenticator
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)

	// Send response
	return true, ctx.ResponseWriter.Write(response)
}

// HandleResponse handles EAP-Response with TLS handshake data.
func (h *TLSHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	// Get state
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	// Parse TLS data from EAP message
	tlsData := ctx.EAPMessage.Data

	// Check if this is a client certificate message
	if h.isClientCertificateMessage(tlsData) {
		// Extract and validate client certificate
		certDER := h.extractCertificate(tlsData)
		if certDER != nil {
			// Get tenant ID from context (assuming it's stored somewhere accessible)
			// For now, we'll use a default approach
			tenantID := h.getTenantID(ctx)

			userID, err := h.certService.ValidateClientCertificate(tenantID, certDER)
			if err != nil {
				return false, err
			}

			// Store authenticated user ID in state
			state.Data["user_id"] = userID
			state.Success = true
			_ = ctx.StateManager.SetState(stateID, state) //nolint:errcheck

			return true, nil
		}
	}

	// For now, accept the TLS handshake
	// In a full implementation, this would process all TLS handshake messages
	state.Data["tls_handshake_complete"] = true
	_ = ctx.StateManager.SetState(stateID, state) //nolint:errcheck

	return true, nil
}

// buildTLSStartRequest creates an EAP-TLS start request.
func (h *TLSHandler) buildTLSStartRequest(identifier uint8) []byte {
	// EAP-TLS start request: Code (1) | Identifier (1) | Length (2) | Type (13) | Flags (0)
	// Flags: 0x00 indicates start of TLS handshake
	totalLen := 5 + 1 // EAP header (5) + Type (1) + Flags (1)

	buffer := make([]byte, totalLen)
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	buffer[2] = byte(totalLen >> 8)
	buffer[3] = byte(totalLen)
	buffer[4] = eap.TypeTLS
	buffer[5] = 0x00 // Start flag

	return buffer
}

// isClientCertificateMessage checks if the TLS message contains a client certificate.
// TLS handshake message types:
// 0 = ClientHello, 1 = ClientHello, 2 = ServerHello, 11 = Certificate, etc.
func (h *TLSHandler) isClientCertificateMessage(data []byte) bool {
	// This is a simplified check
	// In a full implementation, proper TLS message parsing is required
	if len(data) < 1 {
		return false
	}

	// Check for TLS handshake type
	// TLS record layer: Content Type (1) | Version (2) | Length (2) | Handshake data
	// Handshake data: Type (1) | Length (3) | ...
	if len(data) >= 6 {
		contentType := data[0]
		handshakeType := data[5]

		// Content Type 22 = Handshake
		// Handshake Type 11 = Certificate
		if contentType == 22 && handshakeType == 11 {
			return true
		}
	}

	return false
}

// extractCertificate extracts the X.509 certificate from TLS certificate message.
func (h *TLSHandler) extractCertificate(data []byte) []byte {
	// Simplified certificate extraction
	// In a full implementation, this would parse the TLS Certificate message properly
	// which can contain multiple certificates in a chain

	if len(data) < 10 {
		return nil
	}

	// Skip TLS record layer header (5 bytes) and handshake message header (4 bytes)
	// to get to the certificate data
	certDataStart := 9

	// Try to parse as PEM
	block, _ := pem.Decode(data[certDataStart:])
	if block != nil {
		return block.Bytes
	}

	// Try to parse as DER directly
	// This assumes the certificate starts at a certain offset
	// Real implementation would properly parse TLS handshake messages
	if len(data) > certDataStart+3 {
		certLength := int(data[certDataStart])<<16 | int(data[certDataStart+1])<<8 | int(data[certDataStart+2])
		if len(data) >= certDataStart+3+certLength {
			return data[certDataStart+3 : certDataStart+3+certLength]
		}
	}

	return nil
}

// getTenantID extracts the tenant ID from the context.
func (h *TLSHandler) getTenantID(ctx *eap.EAPContext) int64 {
	// Try to get tenant ID from user
	if ctx.User != nil {
		return ctx.User.TenantID
	}

	// Try to get from NAS
	if ctx.NAS != nil {
		return ctx.NAS.TenantID
	}

	// Default to tenant 1 (should be properly configured in production)
	return 1
}

// GetTLSCertificate returns the server certificate for EAP-TLS.
func (h *TLSHandler) GetTLSCertificate(tenantID int64) (*tls.Certificate, error) {
	// Get active CA for tenant
	var ca domain.CertificateAuthority
	err := h.db.Where("tenant_id = ? AND status = ?", tenantID, "active").
		First(&ca).Error

	if err != nil {
		return nil, err
	}

	// Load certificate
	cert, err := tls.X509KeyPair(
		[]byte(ca.CertificatePEM),
		[]byte(ca.PrivateKeyPEM),
	)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// ValidateClientCertificate validates a client certificate against the CA.
func (h *TLSHandler) ValidateClientCertificate(
	tenantID int64,
	rawCert [][]byte,
) (int64, error) {
	for _, certDER := range rawCert {
		userID, err := h.certService.ValidateClientCertificate(tenantID, certDER)
		if err != nil {
			return 0, err
		}
		return userID, nil
	}
	return 0, x509.ErrUnsupportedAlgorithm
}
