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
	ID             int64     `json:"id,string" gorm:"primaryKey"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	Name           string    `json:"name" gorm:"not null;size:200"`
	CommonName     string    `json:"common_name" gorm:"not null;size:255"`
	CertificatePEM string    `json:"certificate_pem" gorm:"type:text;not null"`
	PrivateKeyPEM  string    `json:"private_key_pem" gorm:"type:text;not null"`
	SerialNumber   string    `json:"serial_number" gorm:"index"`
	ExpiryDate     time.Time `json:"expiry_date"`
	Status         string    `json:"status" gorm:"default:active"` // active, revoked, expired
	CRLURL         string    `json:"crl_url" gorm:"size:500"`
	Remark         string    `json:"remark" gorm:"size:500"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	ID                int64      `json:"id,string" gorm:"primaryKey"`
	TenantID          int64      `json:"tenant_id" gorm:"index"`
	UserID            int64      `json:"user_id" gorm:"index"`
	CaID              int64      `json:"ca_id" gorm:"index"`
	CommonName        string    `json:"common_name" gorm:"not null;size:255;index"`
	SerialNumber      string    `json:"serial_number" gorm:"index"`
	CertificatePEM    string    `json:"certificate_pem" gorm:"type:text;not null"`
	PrivateKeyPEM     string    `json:"private_key_pem" gorm:"type:text"` // Empty if user managed
	ExpiryDate        time.Time `json:"expiry_date"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	RevocationReason  string    `json:"revocation_reason" gorm:"size:500"`
	Status            string    `json:"status" gorm:"default:active"` // active, revoked, expired
	DeviceType        string    `json:"device_type" gorm:"size:50"`    // laptop, phone, iot
	MACAddress        string    `json:"mac_address" gorm:"size:17"`
	Remark            string    `json:"remark" gorm:"size:500"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
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
	ID             int64     `json:"id,string" gorm:"primaryKey"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	UserID         int64     `json:"user_id" gorm:"index"`
	CircuitID      string    `json:"circuit_id" gorm:"size:255;index"` // Agent Circuit ID suboption
	RemoteID       string    `json:"remote_id" gorm:"size:255;index"`  // Agent Remote ID suboption
	IPAddress      string    `json:"ip_address" gorm:"size:45;index"`   // Assigned IP address
	MACAddress     string    `json:"mac_address" gorm:"size:17"`        // Client MAC
	VendorSpecific string    `json:"vendor_specific" gorm:"type:text"`  // Vendor-specific suboptions
	LastSeen       time.Time `json:"last_seen"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	ID             int64     `json:"id,string" gorm:"primaryKey"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	UserID         int64     `json:"user_id" gorm:"index"`
	IPAddress      string    `json:"ip_address" gorm:"size:45;index"`
	MACAddress     string    `json:"mac_address" gorm:"size:17;index"`
	CircuitID      string    `json:"circuit_id" gorm:"size:255"`
	RemoteID       string    `json:"remote_id" gorm:"size:255"`
	SessionID      string    `json:"session_id" gorm:"size:64;index"`
	NasID          int64     `json:"nas_id" gorm:"index"`
	NasPort        string    `json:"nas_port" gorm:"size:50"`
	FramedIP       string    `json:"framed_ip" gorm:"size:45"`
	SessionStart   time.Time `json:"session_start"`
	SessionUpdate  time.Time `json:"session_update"`
	InputOctets    int64     `json:"input_octets"`
	OutputOctets   int64     `json:"output_octets"`
	Status         string    `json:"status" gorm:"default:active"` // active, terminated
	TerminateCause string    `json:"terminate_cause" gorm:"size:100"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
