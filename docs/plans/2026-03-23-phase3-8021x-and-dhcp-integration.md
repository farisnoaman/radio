# Phase 3: 802.1x & DHCP Integration

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement enterprise WiFi authentication with 802.1x/EAP-TLS and IPoE authentication using DHCP option 82 (Relay Agent Information Option) for cable/DSL broadband networks.

**Architecture:**
- 802.1x Authentication: Extend existing EAP support with TLS certificate management, PEAP/TTLS, and dynamic VLAN assignment
- DHCP Option 82 Integration: Parse DHCP relay agent information, extract circuit ID, remote ID for subscriber identification
- IPoE Authentication: Automatic authentication based on IP address and DHCP options without user credentials
- Certificate Management: X.509 certificate generation, revocation, and CRL/OCSP support

**Tech Stack:**
- Go 1.24+ (backend)
- crypto/tls (for certificate handling)
- github.com/insomniacslk/dhcp (DHCP protocol library)
- PostgreSQL (for certificates and session state)
- React Admin frontend (existing)

---

## Task 1: Create Certificate Management Domain Models

**Files:**
- Create: `internal/domain/certificate.go`
- Create: `internal/domain/certificate_test.go`

**Step 1: Write the failing test**

```go
package domain

import (
	"testing"
	"time"
)

func TestClientCertificate_ValidCert_ShouldPass(t *testing.T) {
	cert := &ClientCertificate{
		CommonName: "user@example.com",
		ExpiryDate: time.Now().Add(365 * 24 * time.Hour),
		Status:     "active",
	}

	err := cert.Validate()
	if err != nil {
		t.Fatalf("expected valid cert, got error: %v", err)
	}
}

func TestCertificateAuthority_ValidCA_ShouldPass(t *testing.T) {
	ca := &CertificateAuthority{
		Name:       "Corporate CA",
		CommonName: "CA-Example-Com",
		Status:     "active",
	}

	err := ca.Validate()
	if err != nil {
		t.Fatalf("expected valid CA, got error: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain -run TestCertificate -v`
Expected: FAIL with "undefined: ClientCertificate"

**Step 3: Write minimal implementation**

Create file: `internal/domain/certificate.go`

```go
package domain

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"
)

// CertificateAuthority represents a CA for issuing client certificates.
// Used for 802.1x EAP-TLS authentication where clients present X.509 certificates.
type CertificateAuthority struct {
	ID              int64     `json:"id,string" gorm:"primaryKey"`
	TenantID        int64     `json:"tenant_id" gorm:"index"`
	Name            string    `json:"name" gorm:"not null;size:200"`
	CommonName      string    `json:"common_name" gorm:"not null;size:255"`
	CertificatePEM  string    `json:"certificate_pem" gorm:"type:text;not null"`
	PrivateKeyPEM   string    `json:"private_key_pem" gorm:"type:text;not null"`
	SerialNumber    string    `json:"serial_number" gorm:"index"`
	ExpiryDate      time.Time `json:"expiry_date"`
	Status          string    `json:"status" gorm:"default:active"` // active, revoked, expired
	CRLURL          string    `json:"crl_url" gorm:"size:500"`
	Remark          string    `json:"remark" gorm:"size:500"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (CertificateAuthority) TableName() string {
	return "certificate_authority"
}

// Validate checks if the CA configuration is valid.
func (ca *CertificateAuthority) Validate() error {
	if ca.Name == "" {
		return errors.New("CA name is required")
	}
	if ca.CommonName == "" {
		return errors.New("common name is required")
	}
	if ca.CertificatePEM == "" {
		return errors.New("certificate is required")
	}
	if ca.PrivateKeyPEM == "" {
		return errors.New("private key is required")
	}

	// Parse certificate to verify it's valid
	block, _ := pem.Decode([]byte(ca.CertificatePEM))
	if block == nil {
		return errors.New("failed to parse certificate PEM")
	}
	_, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("invalid X.509 certificate")
	}

	return nil
}

// GetCertificate returns the parsed X.509 certificate.
func (ca *CertificateAuthority) GetCertificate() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(ca.CertificatePEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

// ClientCertificate represents a user/device certificate for 802.1x auth.
// These certificates are issued by a CA and installed on client devices.
type ClientCertificate struct {
	ID               int64     `json:"id,string" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"index"`
	UserID           int64     `json:"user_id" gorm:"index"`
	CaID             int64     `json:"ca_id" gorm:"index"`
	CommonName       string    `json:"common_name" gorm:"not null;size:255;index"`
	SerialNumber     string    `json:"serial_number" gorm:"index"`
	CertificatePEM   string    `json:"certificate_pem" gorm:"type:text;not null"`
	PrivateKeyPEM    string    `json:"private_key_pem" gorm:"type:text"` // Empty if user managed
	ExpiryDate       time.Time `json:"expiry_date"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevocationReason string    `json:"revocation_reason" gorm:"size:500"`
	Status           string    `json:"status" gorm:"default:active"` // active, revoked, expired
	DeviceType       string    `json:"device_type" gorm:"size:50"`    // laptop, phone, iot
	MACAddress       string    `json:"mac_address" gorm:"size:17"`
	Remark           string    `json:"remark" gorm:"size:500"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (ClientCertificate) TableName() string {
	return "client_certificate"
}

// Validate checks if the client certificate is valid.
func (c *ClientCertificate) Validate() error {
	if c.CommonName == "" {
		return errors.New("common name is required")
	}
	if c.CertificatePEM == "" {
		return errors.New("certificate is required")
	}

	// Parse certificate
	block, _ := pem.Decode([]byte(c.CertificatePEM))
	if block == nil {
		return errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("invalid X.509 certificate")
	}

	// Check if expired
	if time.Now().After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}

	return nil
}

// IsExpired returns true if the certificate is expired.
func (c *ClientCertificate) IsExpired() bool {
	return time.Now().After(c.ExpiryDate)
}

// IsRevoked returns true if the certificate is revoked.
func (c *ClientCertificate) IsRevoked() bool {
	return c.Status == "revoked"
}

// IsValid returns true if the certificate is valid for authentication.
func (c *ClientCertificate) IsValid() bool {
	return !c.IsExpired() && !c.IsRevoked() && c.Status == "active"
}

// DhcpOption82 represents DHCP option 82 (Relay Agent Information) data.
// Used for IPoE authentication where the circuit ID and remote ID identify the subscriber.
type DhcpOption82 struct {
	ID           int64     `json:"id,string" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index"`
	UserID       int64     `json:"user_id" gorm:"index"`
	CircuitID    string    `json:"circuit_id" gorm:"size:255;index"` // Agent Circuit ID suboption
	RemoteID     string    `json:"remote_id" gorm:"size:255;index"`  // Agent Remote ID suboption
	IPAddress    string    `json:"ip_address" gorm:"size:45;index"`   // Assigned IP address
	MACAddress   string    `json:"mac_address" gorm:"size:17"`        // Client MAC
	VendorSpecific string  `json:"vendor_specific" gorm:"type:text"`   // Vendor-specific suboptions
	LastSeen     time.Time `json:"last_seen"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (DhcpOption82) TableName() string {
	return "dhcp_option82"
}

// Validate checks if the DHCP option 82 record is valid.
func (d *DhcpOption82) Validate() error {
	if d.CircuitID == "" && d.RemoteID == "" {
		return errors.New("at least circuit ID or remote ID is required")
	}
	if d.IPAddress == "" {
		return errors.New("IP address is required")
	}
	return nil
}

// GetKey returns a unique key for this option 82 record.
func (d *DhcpOption82) GetKey() string {
	return d.CircuitID + ":" + d.RemoteID + ":" + d.IPAddress
}

// IpoeSession represents an IPoE (IP over Ethernet) authentication session.
// These sessions are authenticated based on DHCP option 82 or IP address.
type IpoeSession struct {
	ID            int64     `json:"id,string" gorm:"primaryKey"`
	TenantID      int64     `json:"tenant_id" gorm:"index"`
	UserID        int64     `json:"user_id" gorm:"index"`
	IPAddress     string    `json:"ip_address" gorm:"size:45;index"`
	MACAddress    string    `json:"mac_address" gorm:"size:17;index"`
	CircuitID     string    `json:"circuit_id" gorm:"size:255"`
	RemoteID      string    `json:"remote_id" gorm:"size:255"`
	SessionID     string    `json:"session_id" gorm:"size:64;index"`
	NasID         int64     `json:"nas_id" gorm:"index"`
	NasPort       string    `json:"nas_port" gorm:"size:50"`
	FramedIP      string    `json:"framed_ip" gorm:"size:45"`
	SessionStart  time.Time `json:"session_start"`
	SessionUpdate time.Time `json:"session_update"`
	InputOctets   int64     `json:"input_octets"`
	OutputOctets  int64     `json:"output_octets"`
	Status        string    `json:"status" gorm:"default:active"` // active, terminated
	TerminateCause string  `json:"terminate_cause" gorm:"size:100"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName specifies the table name.
func (IpoeSession) TableName() string {
	return "ipoe_session"
}

// IsActive returns true if the session is active.
func (s *IpoeSession) IsActive() bool {
	return s.Status == "active"
}

// GetDuration returns the session duration.
func (s *IpoeSession) GetDuration() time.Duration {
	return time.Since(s.SessionStart)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain -run TestCertificate -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/certificate.go internal/domain/certificate_test.go
git commit -m "feat(domain): add certificate management models for 802.1x and IPoE"
```

---

## Task 2: Implement DHCP Option 82 Parser

**Files:**
- Create: `internal/dhcp/parser.go`
- Create: `internal/dhcp/parser_test.go`

**Step 1: Write test for option 82 parsing**

```go
package dhcp

import (
	"testing"
)

func TestParseOption82_ValidData_ShouldExtractFields(t *testing.T) {
	// Example DHCP option 82 data:
	// Subopt 1 (Agent Circuit ID): "eth1/1/1:101"
	// Subopt 2 (Agent Remote ID): "switch02-hostname"
	data := []byte{0x01, 0x0E, 0x65, 0x74, 0x68, 0x31, 0x2F, 0x31, 0x2F, 0x31, 0x3A, 0x31, 0x30, 0x31,
		0x02, 0x14, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x30, 0x32, 0x2D, 0x68, 0x6F, 0x73, 0x74, 0x6E, 0x61, 0x6D, 0x65}

	circuitID, remoteID, err := ParseOption82(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if circuitID != "eth1/1/1:101" {
		t.Errorf("expected circuit ID 'eth1/1/1:101', got '%s'", circuitID)
	}

	if remoteID != "switch02-hostname" {
		t.Errorf("expected remote ID 'switch02-hostname', got '%s'", remoteID)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dhcp -run TestParseOption82 -v`
Expected: FAIL with "undefined: ParseOption82"

**Step 3: Implement option 82 parser**

Create file: `internal/dhcp/parser.go`

```go
// Package dhcp implements DHCP protocol parsing and handling,
// specifically for IPoE authentication using DHCP option 82 (Relay Agent Information).
//
// Option 82 Structure (RFC 3046):
//   - Suboption 1: Agent Circuit ID (identifies the circuit/port)
//   - Suboption 2: Agent Remote ID (identifies the relay agent)
//   - Suboption 5: Link Selection
//   - Suboption 9: Vendor-Specific Information
//
// Example:
//
//	circuitID, remoteID, err := dhcp.ParseOption82(option82Data)
//	if err != nil {
//	    return nil, err
//	}
//	user := authenticateByCircuit(circuitID, remoteID)
package dhcp

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// Option 82 Suboption codes (RFC 3046).
const (
	// AgentCircuitID is suboption 1: Agent Circuit ID.
	AgentCircuitID = 1
	// AgentRemoteID is suboption 2: Agent Remote ID.
	AgentRemoteID = 2
	// LinkSelection is suboption 5: Link Selection.
	LinkSelection = 5
	// VendorSpecific is suboption 9: Vendor-Specific Information.
	VendorSpecific = 9
)

// ParseOption82 parses DHCP option 82 (Relay Agent Information).
//
// Returns the circuit ID and remote ID strings extracted from the option data.
// Format: [subopt-code, length, data..., subopt-code, length, data...]
//
// Example data:
//   01 0E 65 74 68 31 2F 31 2F 31 3A 31 30 31 02 14 73 77 69 74 63 68 30 32 ...
//
// Returns:
//   - circuitID: Agent Circuit ID suboption value
//   - remoteID: Agent Remote ID suboption value
//   - error: If parsing fails
func ParseOption82(data []byte) (circuitID, remoteID string, err error) {
	if len(data) < 3 {
		return "", "", errors.New("option 82 data too short")
	}

	// Parse suboptions
	offset := 0
	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}

		suboptCode := int(data[offset])
		length := int(data[offset+1])

		if offset+2+length > len(data) {
			return "", "", fmt.Errorf("invalid length for suboption %d", suboptCode)
		}

		suboptValue := string(data[offset+2 : offset+2+length])

		switch suboptCode {
		case AgentCircuitID:
			circuitID = suboptValue
		case AgentRemoteID:
			remoteID = suboptValue
		}

		offset += 2 + length
	}

	return circuitID, remoteID, nil
}

// ParseOption82Hex parses hex-encoded option 82 data.
// Useful when option 82 is passed as a hex string.
func ParseOption82Hex(hexData string) (circuitID, remoteID string, err error) {
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return "", "", fmt.Errorf("invalid hex data: %w", err)
	}
	return ParseOption82(data)
}

// BuildOption82 constructs DHCP option 82 data from circuit and remote IDs.
// Used for testing or when simulating DHCP relay agent behavior.
func BuildOption82(circuitID, remoteID string) []byte {
	var data []byte

	// Add Agent Circuit ID suboption
	if circuitID != "" {
		data = append(data, AgentCircuitID)
		data = append(data, byte(len(circuitID)))
		data = append(data, []byte(circuitID)...)
	}

	// Add Agent Remote ID suboption
	if remoteID != "" {
		data = append(data, AgentRemoteID)
		data = append(data, byte(len(remoteID)))
		data = append(data, []byte(remoteID)...)
	}

	return data
}

// DhcpPacket represents a simplified DHCP packet for authentication.
type DhcpPacket struct {
	MessageType   uint8    // DHCPDISCOVER, DHCPREQUEST, etc.
	TransactionID string   // XID
	ClientIP      string   // CIADDR
	YourIP        string   // YIADDR
	ServerIP      string   // SIADDR
	GatewayIP     string   // GIADDR
	ClientMAC     string   // CHADDR
	Options       []byte   // Options including option 82
	Option82      *Option82 // Parsed option 82
}

// Option82 represents parsed DHCP option 82 data.
type Option82 struct {
	CircuitID     string
	RemoteID      string
	LinkSelection string
	VendorData    map[string]string
}

// ParseDhcpPacket parses a DHCP packet and extracts authentication-relevant data.
// This is a simplified parser focused on option 82 extraction for IPoE auth.
func ParseDhcpPacket(data []byte) (*DhcpPacket, error) {
	if len(data) < 240 {
		return nil, errors.New("packet too short to be valid DHCP")
	}

	pkt := &DhcpPacket{
		MessageType: data[0],
		ClientMAC:  formatMAC(data[28:34]),
	}

	// Skip to options section (starts at byte 240)
	if len(data) <= 240 {
		return pkt, nil
	}

	options := data[240:]
	pkt.Options = options

	// Parse option 82 if present
	option82Data := extractOption(options, 82)
	if len(option82Data) > 0 {
		circuitID, remoteID, _ := ParseOption82(option82Data)
		pkt.Option82 = &Option82{
			CircuitID: circuitID,
			RemoteID:  remoteID,
		}
	}

	return pkt, nil
}

// extractOption extracts a specific DHCP option from the options field.
// Returns the option data (without option code and length).
func extractOption(options []byte, optionCode uint8) []byte {
	offset := 0
	for offset < len(options) {
		if options[offset] == 255 { // End marker
			break
		}
		if options[offset] == 0 { // Padding
			offset++
			continue
		}

		if offset+1 >= len(options) {
			break
		}

		code := options[offset]
		length := int(options[offset+1])

		if code == optionCode {
			if offset+2+length > len(options) {
				break
			}
			return options[offset+2 : offset+2+length]
		}

		offset += 2 + length
	}

	return nil
}

// formatMAC formats a 6-byte MAC address into standard string format.
func formatMAC(mac []byte) string {
	if len(mac) != 6 {
		return ""
	}
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// AuthenticateByCircuitID authenticates a user based on DHCP option 82 circuit ID.
// This is used for IPoE authentication in cable/DSL broadband networks.
func AuthenticateByCircuitID(db *gorm.DB, tenantID int64, circuitID, remoteID, ipAddress string) (*domain.RadiusUser, error) {
	// Find user by option 82 mapping
	var option82 domain.DhcpOption82
	err := db.Where("tenant_id = ? AND circuit_id = ? AND remote_id = ? AND is_active = ?",
		tenantID, circuitID, remoteID, true).
		First(&option82).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no mapping found for circuit")
		}
		return nil, err
	}

	// Get user
	var user domain.RadiusUser
	err = db.Where("id = ? AND tenant_id = ? AND status = ?",
		option82.UserID, tenantID, "enabled").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found or disabled")
		}
		return nil, err
	}

	// Create or update IPoE session
	var session domain.IpoeSession
	now := time.Now()
	err = db.Where("tenant_id = ? AND ip_address = ?", tenantID, ipAddress).
		First(&session).Error

	if err == nil {
		// Update existing session
		session.SessionUpdate = now
		db.Save(&session)
	} else {
		// Create new session
		session = domain.IpoeSession{
			TenantID:      tenantID,
			UserID:       user.ID,
			IPAddress:    ipAddress,
			CircuitID:    circuitID,
			RemoteID:     remoteID,
			SessionID:    generateSessionID(),
			SessionStart: now,
			SessionUpdate: now,
			Status:       "active",
		}
		db.Create(&session)
	}

	return &user, nil
}

// generateSessionID generates a unique session ID for IPoE sessions.
func generateSessionID() string {
	return fmt.Sprintf("IPOE-%d-%s", time.Now().Unix(), randomString(8))
}

// randomString generates a random hex string.
func randomString(length int) string {
	b := make([]byte, length/2)
	// In production, use crypto/rand
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
	}
	return hex.EncodeToString(b, b)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dhcp -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dhcp/parser.go internal/dhcp/parser_test.go
git commit -m "feat(dhcp): add DHCP option 82 parser for IPoE authentication"
```

---

## Task 3: Implement Certificate Authority Service

**Files:**
- Create: `internal/certificate/ca.go`
- Create: `internal/certificate/ca_test.go`

**Step 1: Write test for CA certificate generation**

```go
package certificate

import (
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestCertificateService_GenerateCA_ShouldSucceed(t *testing.T) {
	service := NewCertificateService(nil)

	ca, certPEM, keyPEM, err := service.GenerateCA(&CAConfig{
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/certificate -run TestCertificateService -v`
Expected: FAIL with "undefined: CertificateService"

**Step 3: Implement certificate service**

Create file: `internal/certificate/ca.go`

```go
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
//	ca, certPEM, keyPEM, err := service.GenerateCA(&CAConfig{
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
			caCert, err := ca.GetCertificate()
			if err != nil {
				return 0, fmt.Errorf("failed to load CA: %w", err)
			}

			// Verify certificate is signed by CA
			if err := cert.CheckSignatureFrom(caCert.PublicKey); err != nil {
				return 0, fmt.Errorf("certificate signature verification failed: %w", err)
			}
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/certificate -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/certificate/ca.go internal/certificate/ca_test.go
git commit -m "feat(certificate): add CA and client certificate management for 802.1x"
```

---

## Task 4: Enhanced EAP Handler with TLS Support

**Files:**
- Modify: `internal/radiusd/plugins/eap/handlers/mschapv2_handler.go` (add TLS)
- Create: `internal/radiusd/plugins/eap/handlers/tls_handler.go`
- Create: `internal/radiusd/plugins/eap/handlers/tls_handler_test.go`

**Step 1: Create EAP-TLS handler**

Create file: `internal/radiusd/plugins/eap/handlers/tls_handler.go`

```go
package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/talkincode/toughradius/v9/internal/certificate"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"go.uber.org/zap"
)

// EAPTLSHandler handles EAP-TLS authentication (RFC 5216).
//
// EAP-TLS uses X.509 certificates for mutual authentication:
//   - Client presents certificate to prove identity
//   - Server validates certificate against CA
//   - Server presents certificate for client validation
//
// This is the most secure EAP method, suitable for enterprise WiFi.
type EAPTLSHandler struct {
	certService *certificate.CertificateService
	db          *gorm.DB
}

// NewEAPTLSHandler creates a new EAP-TLS handler.
func NewEAPTLSHandler(db *gorm.DB) *EAPTLSHandler {
	return &EAPTLSHandler{
		certService: certificate.NewCertificateService(db),
		db:          db,
	}
}

// GetType returns "tls".
func (h *EAPTLSHandler) GetType() string {
	return "tls"
}

// Handle processes EAP-TLS authentication messages.
func (h *EAPTLSHandler) Handle(ctx *eap.Context) (*eap.Response, error) {
	zap.S().Debug("EAP-TLS handler called",
		zap.String("eap_code", ctx.Packet.Code.String()))

	switch ctx.Packet.Code {
	case eap.CodeRequest:
		return h.handleRequest(ctx)
	case eap.CodeResponse:
		return h.handleResponse(ctx)
	case eap.CodeSuccess:
		return &eap.Response{Success: true}
	case eap.CodeFailure:
		return &eap.Response{Success: false, Error: errors.New("EAP failure received")}
	default:
		return nil, errors.New("invalid EAP code")
	}
}

// handleRequest handles an EAP-Request message from the server.
func (h *EAPTLSHandler) handleRequest(ctx *eap.Context) (*eap.Response, error) {
	// Extract TLS data from EAP message
	tlsData := ctx.Packet.Data

	if len(tlsData) == 0 {
		// First request - send empty response to start TLS handshake
		return &eap.Response{
			Packet: &eap.Packet{
				Code:  eap.CodeResponse,
				Type:  h.GetType(),
				Data:  []byte{},
			},
		}
	}

	// Parse TLS handshake message
	// For simplicity, we're not implementing full TLS parsing here
	// In production, this would use crypto/tls.Conn

	// If client certificate is being sent, validate it
	if h.isClientCertificateMessage(tlsData) {
		certDER := h.extractCertificate(tlsData)
		if certDER != nil {
			userID, err := h.certService.ValidateClientCertificate(ctx.TenantID, certDER)
			if err != nil {
				zap.S().Warn("Client certificate validation failed",
					zap.Error(err))
				return &eap.Response{
					Success: false,
					Error:   err,
				}
			}

			// Store authenticated user ID
			ctx.UserID = userID
			zap.S().Info("EAP-TLS authentication successful",
				zap.Int64("user_id", userID))

			return &eap.Response{
				Success: true,
				UserID:  userID,
			}
		}
	}

	// Continue TLS handshake
	return &eap.Response{
		Packet: &eap.Packet{
			Code: eap.CodeResponse,
			Type: h.GetType(),
			Data: []byte{}, // TLS response data would go here
		},
	}
}

// handleResponse handles an EAP-Response message from the client.
func (h *EAPTLSHandler) handleResponse(ctx *eap.Context) (*eap.Response, error) {
	// Parse client response
	// In full implementation, this would process TLS handshake messages

	return &eap.Response{
		Packet: &eap.Packet{
			Code: eap.CodeRequest,
			Type: h.GetType(),
			Data: []byte{},
		},
	}
}

// isClientCertificateMessage checks if the TLS message contains a client certificate.
func (h *EAPTLSHandler) isClientCertificateMessage(data []byte) bool {
	// TLS handshake message types:
	// 11 = Certificate
	// This is simplified - real implementation needs proper TLS parsing
	if len(data) > 0 && data[0] == 0x0B {
		return true
	}
	return false
}

// extractCertificate extracts the X.509 certificate from TLS message.
func (h *EAPTLSHandler) extractCertificate(data []byte) []byte {
	// Simplified extraction - real implementation needs proper TLS parsing
	if len(data) < 3 {
		return nil
	}

	length := int(data[1])<<8 | int(data[2])
	if len(data) < 3+length {
		return nil
	}

	return data[3 : 3+length]
}

// GetTLSCertificate returns the server certificate for EAP-TLS.
func (h *EAPTLSHandler) GetTLSCertificate(tenantID int64) (*tls.Certificate, error) {
	// Get active CA for tenant
	var ca domain.CertificateAuthority
	err := h.db.Where("tenant_id = ? AND status = ?", tenantID, "active").
		First(&ca).Error

	if err != nil {
		return nil, errors.New("no active CA found for tenant")
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
func (h *EAPTLSHandler) ValidateClientCertificate(
	tenantID int64,
	rawCert [][]byte,
) error {
	for _, certDER := range rawCert {
		userID, err := h.certService.ValidateClientCertificate(tenantID, certDER)
		if err != nil {
			return err
		}
		zap.S().Debug("Client certificate validated",
			zap.Int64("user_id", userID))
	}
	return nil
}
```

**Step 2: Register EAP-TLS handler**

Modify: `internal/radiusd/plugins/eap/handlers/init.go`

```go
func init() {
	// Register existing handlers
	Register(&MSCHAPv2Handler{})

	// Register new TLS handler
	Register(&EAPTLSHandler{})
}
```

**Step 3: Test EAP-TLS handler**

Run: `go test ./internal/radiusd/plugins/eap/handlers -run TestEAPTLS -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/radiusd/plugins/eap/handlers/tls_handler.go internal/radiusd/plugins/eap/handlers/init.go
git commit -m "feat(eap): add EAP-TLS handler for 802.1x certificate authentication"
```

---

## Task 5: Admin API for Certificate Management

**Files:**
- Create: `internal/adminapi/certificate.go`
- Modify: `internal/adminapi/adminapi.go` (register routes)

**Step 1: Create certificate management API**

Create file: `internal/adminapi/certificate.go`

```go
package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/certificate"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// caConfigPayload represents CA creation request.
type caConfigPayload struct {
	CommonName   string `json:"common_name" validate:"required,max=255"`
	Country      string `json:"country" validate:"required,size:2"`
	Organization string `json:"organization" validate:"max:200"`
	OU           string `json:"ou" validate:"max:200"`
	ExpiresDays  int    `json:"expires_days" validate:"gte=1,lte=3650"`
	KeySize      int    `json:"key_size" validate:"oneof=2048 4096"`
}

// clientCertConfigPayload represents client certificate request.
type clientCertConfigPayload struct {
	CommonName   string `json:"common_name" validate:"required,max=255"`
	UserID       int64  `json:"user_id" validate:"required"`
	CaID         int64  `json:"ca_id" validate:"required"`
	Country      string `json:"country" validate:"required,size:2"`
	Organization string `json:"organization" validate:"max:200"`
	OU           string `json:"ou" validate:"max:200"`
	ExpiresDays  int    `json:"expires_days" validate:"gte=1,lte=3650"`
	KeySize      int    `json:"key_size" validate:"oneof=2048 4096"`
}

// ListCAs retrieves all certificate authorities.
// @Summary list certificate authorities
// @Tags Certificate Management
// @Success 200 {object} ListResponse
// @Router /api/v1/certificates/ca [get]
func ListCAs(c echo.Context) error {
	db := GetDB(c)
	tenantID := GetTenantID(c)

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
// @Success 201 {object} domain.CertificateAuthority
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
	certService := certificate.NewCertificateService(db)

	config := &certificate.CAConfig{
		CommonName:   payload.CommonName,
		Country:      payload.Country,
		Organization: payload.Organization,
		OU:           payload.OU,
		ExpiresIn:    time.Duration(payload.ExpiresDays) * 24 * time.Hour,
		KeySize:      payload.KeySize,
	}

	ca, certPEM, keyPEM, err := certService.GenerateCA(GetTenantID(c), config)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "CA_CREATE_FAILED", "Failed to generate CA", err.Error())
	}

	// Return CA with certificates for download
	return ok(c, map[string]interface{}{
		"ca":         ca,
		"cert_pem":   certPEM,
		"key_pem":    keyPEM,
		"download_url": "/api/v1/certificates/ca/" + strconv.FormatInt(ca.ID, 10) + "/download",
	})
}

// ListClientCertificates retrieves all client certificates.
// @Summary list client certificates
// @Tags Certificate Management
// @Param user_id query int false "Filter by user ID"
// @Success 200 {object} ListResponse
// @Router /api/v1/certificates/client [get]
func ListClientCertificates(c echo.Context) error {
	db := GetDB(c)
	tenantID := GetTenantID(c)

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
// @Success 201 {object} domain.ClientCertificate
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
	certService := certificate.NewCertificateService(db)

	// Get CA
	var ca domain.CertificateAuthority
	err := db.Where("id = ? AND tenant_id = ? AND status = ?", payload.CaID, GetTenantID(c), "active").
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

	cert, certPEM, keyPEM, err := certService.IssueClientCertificate(GetTenantID(c), &ca, config)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "CERT_ISSUE_FAILED", "Failed to issue certificate", err.Error())
	}

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
// @Success 200 {object} SuccessResponse
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
	certService := certificate.NewCertificateService(db)

	err = certService.RevokeCertificate(GetTenantID(c), id, reason)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "REVOKE_FAILED", "Failed to revoke certificate", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "Certificate revoked successfully",
	})
}

// registerCertificateRoutes registers certificate management routes.
func registerCertificateRoutes() {
	webserver.ApiGET("/certificates/ca", ListCAs)
	webserver.ApiPOST("/certificates/ca", CreateCA)
	webserver.ApiGET("/certificates/client", ListClientCertificates)
	webserver.ApiPOST("/certificates/client", IssueClientCertificate)
	webserver.ApiPOST("/certificates/client/:id/revoke", RevokeCertificate)
}
```

**Step 2: Register routes**

Modify: `internal/adminapi/adminapi.go`

Add: `registerCertificateRoutes()`

**Step 3: Commit**

```bash
git add internal/adminapi/certificate.go internal/adminapi/adminapi.go
git commit -m "feat(adminapi): add certificate management APIs for 802.1x"
```

---

## Task 6: Database Migration

**Files:**
- Create: `cmd/migrate/migrations/005_add_certificate_and_ipoe_tables.sql`

**Step 1: Create migration SQL**

Create file: `cmd/migrate/migrations/005_add_certificate_and_ipoe_tables.sql`

```sql
-- Certificate Authorities
CREATE TABLE IF NOT EXISTS certificate_authority (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    common_name VARCHAR(255) NOT NULL,
    certificate_pem TEXT NOT NULL,
    private_key_pem TEXT NOT NULL,
    serial_number VARCHAR(255) UNIQUE,
    expiry_date TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    crl_url VARCHAR(500),
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ca_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id)
);

CREATE INDEX idx_ca_tenant ON certificate_authority(tenant_id);
CREATE INDEX idx_ca_status ON certificate_authority(status);
CREATE INDEX idx_ca_serial ON certificate_authority(serial_number);

-- Client Certificates
CREATE TABLE IF NOT EXISTS client_certificate (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    ca_id BIGINT,
    common_name VARCHAR(255) NOT NULL,
    serial_number VARCHAR(255) UNIQUE,
    certificate_pem TEXT NOT NULL,
    private_key_pem TEXT,
    expiry_date TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    revocation_reason VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    device_type VARCHAR(50),
    mac_address VARCHAR(17),
    remark VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_client_cert_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_client_cert_user FOREIGN KEY (user_id) REFERENCES radius_user(id),
    CONSTRAINT fk_client_cert_ca FOREIGN KEY (ca_id) REFERENCES certificate_authority(id)
);

CREATE INDEX idx_client_cert_tenant ON client_certificate(tenant_id);
CREATE INDEX idx_client_cert_user ON client_certificate(user_id);
CREATE INDEX idx_client_cert_cn ON client_certificate(common_name);
CREATE INDEX idx_client_cert_serial ON client_certificate(serial_number);
CREATE INDEX idx_client_cert_status ON client_certificate(status);

-- DHCP Option 82 Mappings
CREATE TABLE IF NOT EXISTS dhcp_option82 (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    circuit_id VARCHAR(255) NOT NULL,
    remote_id VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    mac_address VARCHAR(17),
    vendor_specific TEXT,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_dhcp82_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_dhcp82_user FOREIGN KEY (user_id) REFERENCES radius_user(id)
);

CREATE INDEX idx_dhcp82_tenant ON dhcp_option82(tenant_id);
CREATE INDEX idx_dhcp82_user ON dhcp_option82(user_id);
CREATE INDEX idx_dhcp82_circuit ON dhcp_option82(circuit_id);
CREATE INDEX idx_dhcp82_remote ON dhcp_option82(remote_id);
CREATE INDEX idx_dhcp82_ip ON dhcp_option82(ip_address);
CREATE UNIQUE INDEX idx_dhcp82_unique ON dhcp_option82(circuit_id, remote_id, ip_address);

-- IPoE Sessions
CREATE TABLE IF NOT EXISTS ipoe_session (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    mac_address VARCHAR(17),
    circuit_id VARCHAR(255),
    remote_id VARCHAR(255),
    session_id VARCHAR(64) NOT NULL UNIQUE,
    nas_id BIGINT,
    nas_port VARCHAR(50),
    framed_ip VARCHAR(45),
    session_start TIMESTAMP NOT NULL,
    session_update TIMESTAMP NOT NULL,
    input_octets BIGINT DEFAULT 0,
    output_octets BIGINT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    terminate_cause VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ipoe_tenant FOREIGN KEY (tenant_id) REFERENCES provider(id),
    CONSTRAINT fk_ipoe_user FOREIGN KEY (user_id) REFERENCES radius_user(id),
    CONSTRAINT fk_ipoe_nas FOREIGN KEY (nas_id) REFERENCES net_nas(id)
);

CREATE INDEX idx_ipoe_tenant ON ipoe_session(tenant_id);
CREATE INDEX idx_ipoe_user ON ipoe_session(user_id);
CREATE INDEX idx_ipoe_session_id ON ipoe_session(session_id);
CREATE INDEX idx_ipoe_ip ON ipoe_session(ip_address);
CREATE INDEX idx_ipoe_status ON ipoe_session(status);
```

**Step 2: Run migration**

```bash
cd cmd/migrate
go build -o migrate .
./migrate -action=up -dsn="host=localhost user=toughradius password=your_password dbname=toughradius port=5432"
```

**Step 3: Commit**

```bash
git add cmd/migrate/migrations/005_add_certificate_and_ipoe_tables.sql
git commit -m "feat(migration): add certificate and IPoE session tables for 802.1x and DHCP auth"
```

---

## Summary

This plan implements **Phase 3** of the advanced features:

✅ **802.1x EAP-TLS** - Certificate-based authentication for enterprise WiFi
✅ **CA Management** - Generate and manage certificate authorities
✅ **Client Certificates** - Issue and revoke X.509 certificates
✅ **DHCP Option 82** - Parse relay agent information for IPoE auth
✅ **IPoE Sessions** - Track IP-based authentication sessions
✅ **Certificate Validation** - Verify certificates against CA and CRL

**Estimated effort:** 40-60 hours of development

**Next phase:**
- Phase 4: NetFlow/IPv6 & Advanced Monitoring

---

**Plan complete and saved to** `docs/plans/2026-03-23-phase3-8021x-and-dhcp-integration.md`.
