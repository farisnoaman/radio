package domain

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Provider represents an ISP/Provider tenant in the multi-provider system.
// Each provider operates independently with their own users, NAS devices,
// vouchers, and billing configuration.
type Provider struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	Code      string         `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Name      string         `json:"name" gorm:"size:255;not null"`
	Status    string         `json:"status" gorm:"size:20;default:'active'"`
	MaxUsers  int            `json:"max_users" gorm:"default:1000"`
	MaxNas    int            `json:"max_nas" gorm:"default:100"`
	Branding  string         `json:"branding" gorm:"type:text"`
	Settings  string         `json:"settings" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Provider) TableName() string {
	return "mst_provider"
}

// ProviderBranding holds the branding configuration for a provider.
// It is stored as JSON in the Branding field.
type ProviderBranding struct {
	LogoURL       string `json:"logo_url"`
	PrimaryColor  string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	FaviconURL    string `json:"favicon_url"`
	CompanyName   string `json:"company_name"`
	SupportEmail  string `json:"support_email"`
	SupportPhone  string `json:"support_phone"`
}

// GetBranding parses and returns the branding configuration.
func (p *Provider) GetBranding() (*ProviderBranding, error) {
	if p.Branding == "" {
		return &ProviderBranding{}, nil
	}
	var branding ProviderBranding
	err := json.Unmarshal([]byte(p.Branding), &branding)
	return &branding, err
}

// SetBranding sets the Branding field from a ProviderBranding struct.
func (p *Provider) SetBranding(b *ProviderBranding) error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	p.Branding = string(data)
	return nil
}

// ProviderSettings holds provider-specific settings.
// Stored as JSON in the Settings field.
type ProviderSettings struct {
	AllowUserRegistration bool     `json:"allow_user_registration"`
	AllowVoucherCreation  bool     `json:"allow_voucher_creation"`
	DefaultProductID      int64    `json:"default_product_id"`
	DefaultProfileID      int64    `json:"default_profile_id"`
	AutoExpireSessions    bool     `json:"auto_expire_sessions"`
	SessionTimeout        int      `json:"session_timeout"` // seconds
	IdleTimeout           int      `json:"idle_timeout"`    // seconds
	MaxConcurrentSessions int      `json:"max_concurrent_sessions"`
	RADIUSSecretTemplate  string   `json:"radius_secret_template"`
	CustomAttributes      []string `json:"custom_attributes"`
}

// GetSettings parses and returns the settings configuration.
func (p *Provider) GetSettings() (*ProviderSettings, error) {
	if p.Settings == "" {
		return &ProviderSettings{
			AllowUserRegistration: true,
			AllowVoucherCreation:  true,
			SessionTimeout:        86400,
			IdleTimeout:           3600,
			MaxConcurrentSessions: 1,
		}, nil
	}
	var settings ProviderSettings
	err := json.Unmarshal([]byte(p.Settings), &settings)
	return &settings, err
}

// SetSettings sets the Settings field from a ProviderSettings struct.
func (p *Provider) SetSettings(s *ProviderSettings) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	p.Settings = string(data)
	return nil
}

// IsActive returns true if the provider status is 'active'.
func (p *Provider) IsActive() bool {
	return p.Status == "active"
}

// IsSuspended returns true if the provider status is 'suspended'.
func (p *Provider) IsSuspended() bool {
	return p.Status == "suspended"
}

// ProviderStats holds usage statistics for a provider.
type ProviderStats struct {
	ProviderID    int64 `json:"provider_id"`
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	OnlineSessions int64 `json:"online_sessions"`
	TotalNas      int64 `json:"total_nas"`
	ActiveNas     int64 `json:"active_nas"`
	TotalVouchers int64 `json:"total_vouchers"`
	UsedVouchers  int64 `json:"used_vouchers"`
}
