package handlers

import (
	"testing"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

func TestTLSHandler_Name_ShouldReturnEAPTLS(t *testing.T) {
	handler := NewTLSHandler(nil)

	if handler.Name() != EAPMethodTLS {
		t.Errorf("expected name '%s', got '%s'", EAPMethodTLS, handler.Name())
	}
}

func TestTLSHandler_EAPType_ShouldReturn13(t *testing.T) {
	handler := NewTLSHandler(nil)

	if handler.EAPType() != eap.TypeTLS {
		t.Errorf("expected EAP type %d, got %d", eap.TypeTLS, handler.EAPType())
	}
}

func TestTLSHandler_CanHandle_WithTLSMessage_ShouldReturnTrue(t *testing.T) {
	handler := NewTLSHandler(nil)

	ctx := &eap.EAPContext{
		EAPMessage: &eap.EAPMessage{
			Type: eap.TypeTLS,
		},
	}

	if !handler.CanHandle(ctx) {
		t.Error("expected handler to handle EAP-TLS messages")
	}
}

func TestTLSHandler_CanHandle_WithMD5Message_ShouldReturnFalse(t *testing.T) {
	handler := NewTLSHandler(nil)

	ctx := &eap.EAPContext{
		EAPMessage: &eap.EAPMessage{
			Type: eap.TypeMD5Challenge,
		},
	}

	if handler.CanHandle(ctx) {
		t.Error("expected handler not to handle EAP-MD5 messages")
	}
}

func TestTLSHandler_CanHandle_WithNilMessage_ShouldReturnFalse(t *testing.T) {
	handler := NewTLSHandler(nil)

	ctx := &eap.EAPContext{
		EAPMessage: nil,
	}

	if handler.CanHandle(ctx) {
		t.Error("expected handler not to handle nil messages")
	}
}

func TestTLSHandler_buildTLSStartRequest_ShouldCreateValidEAPTLS(t *testing.T) {
	handler := NewTLSHandler(nil)

	eapData := handler.buildTLSStartRequest(1)

	// Check minimum length
	if len(eapData) < 6 {
		t.Fatalf("expected EAP data length >= 6, got %d", len(eapData))
	}

	// Check EAP code (request)
	if eapData[0] != eap.CodeRequest {
		t.Errorf("expected EAP code %d, got %d", eap.CodeRequest, eapData[0])
	}

	// Check identifier
	if eapData[1] != 1 {
		t.Errorf("expected identifier 1, got %d", eapData[1])
	}

	// Check EAP type (TLS)
	if eapData[4] != eap.TypeTLS {
		t.Errorf("expected EAP type %d, got %d", eap.TypeTLS, eapData[4])
	}

	// Check start flag
	if eapData[5] != 0x00 {
		t.Errorf("expected start flag 0x00, got 0x%02x", eapData[5])
	}
}

func TestTLSHandler_isClientCertificateMessage_WithCertMessage_ShouldReturnTrue(t *testing.T) {
	handler := NewTLSHandler(nil)

	// Simulated TLS certificate message
	// Content Type: 22 (Handshake)
	// Handshake Type: 11 (Certificate)
	data := []byte{0x16, 0x03, 0x01, 0x00, 0x05, 0x0B}

	if !handler.isClientCertificateMessage(data) {
		t.Error("expected to identify as certificate message")
	}
}

func TestTLSHandler_isClientCertificateMessage_WithNonCertMessage_ShouldReturnFalse(t *testing.T) {
	handler := NewTLSHandler(nil)

	// Different handshake type (ClientHello = 1)
	data := []byte{0x16, 0x03, 0x01, 0x00, 0x5, 0x01}

	if handler.isClientCertificateMessage(data) {
		t.Error("expected not to identify as certificate message")
	}
}

func TestTLSHandler_isClientCertificateMessage_WithEmptyData_ShouldReturnFalse(t *testing.T) {
	handler := NewTLSHandler(nil)

	data := []byte{}

	if handler.isClientCertificateMessage(data) {
		t.Error("expected not to identify empty data as certificate message")
	}
}

func TestTLSHandler_extractCertificate_WithPEMData_ShouldReturnCertDER(t *testing.T) {
	handler := NewTLSHandler(nil)

	// PEM-encoded certificate (simplified for testing)
	pemData := []byte("-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJAKHHCgVZU41ZMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBnRl\nc3RDQTAeFw0yNDAxMDEwMDAwMDBaFw0yNTAxMDEwMDAwMDBaMBExDzANBgNVBAMM\n-----END CERTIFICATE-----")

	// Wrap in TLS message format
	data := make([]byte, 9)
	data[5] = 0x0B // Certificate handshake type
	data = append(data, pemData...)

	certDER := handler.extractCertificate(data)

	if certDER == nil {
		t.Error("expected to extract certificate, got nil")
	}
}

func TestTLSHandler_extractCertificate_WithShortData_ShouldReturnNil(t *testing.T) {
	handler := NewTLSHandler(nil)

	data := []byte{0x01, 0x02, 0x03}

	certDER := handler.extractCertificate(data)

	if certDER != nil {
		t.Error("expected nil for short data, got non-nil")
	}
}

func TestTLSHandler_extractCertificate_WithEmptyData_ShouldReturnNil(t *testing.T) {
	handler := NewTLSHandler(nil)

	certDER := handler.extractCertificate([]byte{})

	if certDER != nil {
		t.Error("expected nil for empty data, got non-nil")
	}
}
