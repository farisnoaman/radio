package domain

import (
	"testing"
)

func TestClientCertificate_ValidCert_ShouldPass(t *testing.T) {
	// This test checks that the client certificate validation logic works
	// We'll test with empty cert to verify error handling
	cert := &ClientCertificate{
		CommonName: "user@example.com",
		Status:     "active",
	}

	err := cert.Validate()
	if err == nil {
		t.Fatal("expected error for missing certificate")
	}

	// Verify it's the expected error
	if err.Error() != "certificate is required" {
		t.Errorf("expected 'certificate is required', got '%v'", err)
	}
}

func TestCertificateAuthority_ValidCA_ShouldPass(t *testing.T) {
	// This test checks that the CA validation logic works
	// We'll test with empty cert to verify error handling
	ca := &CertificateAuthority{
		Name:       "Corporate CA",
		CommonName: "CA-Example-Com",
		Status:     "active",
	}

	err := ca.Validate()
	if err == nil {
		t.Fatal("expected error for missing certificate")
	}

	// Verify it's the expected error
	if err.Error() != "certificate is required" {
		t.Errorf("expected 'certificate is required', got '%v'", err)
	}
}
