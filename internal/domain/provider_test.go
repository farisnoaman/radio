package domain

import (
	"encoding/json"
	"testing"
)

func TestProviderTableName(t *testing.T) {
	p := Provider{}
	if got := p.TableName(); got != "mst_provider" {
		t.Errorf("TableName() = %v, want mst_provider", got)
	}
}

func TestProvider_GetBranding(t *testing.T) {
	p := Provider{}

	// Test empty branding
	branding, err := p.GetBranding()
	if err != nil {
		t.Errorf("GetBranding() error = %v", err)
	}
	if branding == nil {
		t.Error("GetBranding() returned nil")
	}

	// Test with branding JSON
	brandingJSON := `{"logo_url": "https://example.com/logo.png", "primary_color": "#FF0000", "company_name": "Test ISP"}`
	p.Branding = brandingJSON

	branding, err = p.GetBranding()
	if err != nil {
		t.Errorf("GetBranding() error = %v", err)
	}
	if branding.LogoURL != "https://example.com/logo.png" {
		t.Errorf("branding.LogoURL = %v, want https://example.com/logo.png", branding.LogoURL)
	}
	if branding.PrimaryColor != "#FF0000" {
		t.Errorf("branding.PrimaryColor = %v, want #FF0000", branding.PrimaryColor)
	}
	if branding.CompanyName != "Test ISP" {
		t.Errorf("branding.CompanyName = %v, want Test ISP", branding.CompanyName)
	}

	// Test invalid JSON
	p.Branding = "invalid json"
	_, err = p.GetBranding()
	if err == nil {
		t.Error("GetBranding() expected error for invalid JSON")
	}
}

func TestProvider_SetBranding(t *testing.T) {
	p := Provider{}

	branding := &ProviderBranding{
		LogoURL:        "https://example.com/newlogo.png",
		PrimaryColor:   "#00FF00",
		SecondaryColor: "#0000FF",
		CompanyName:    "New ISP",
		SupportEmail:   "support@newisp.com",
		SupportPhone:   "+1234567890",
	}

	err := p.SetBranding(branding)
	if err != nil {
		t.Errorf("SetBranding() error = %v", err)
	}

	// Verify by unmarshaling
	var result ProviderBranding
	err = json.Unmarshal([]byte(p.Branding), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal branding: %v", err)
	}
	if result.LogoURL != branding.LogoURL {
		t.Errorf("LogoURL = %v, want %v", result.LogoURL, branding.LogoURL)
	}
	if result.PrimaryColor != branding.PrimaryColor {
		t.Errorf("PrimaryColor = %v, want %v", result.PrimaryColor, branding.PrimaryColor)
	}
}

func TestProvider_GetSettings(t *testing.T) {
	p := Provider{}

	// Test empty settings returns defaults
	settings, err := p.GetSettings()
	if err != nil {
		t.Errorf("GetSettings() error = %v", err)
	}
	if !settings.AllowUserRegistration {
		t.Error("GetSettings() AllowUserRegistration should be true by default")
	}
	if settings.SessionTimeout != 86400 {
		t.Errorf("GetSettings() SessionTimeout = %v, want 86400", settings.SessionTimeout)
	}

	// Test with settings JSON
	settingsJSON := `{"allow_user_registration": false, "max_concurrent_sessions": 3, "session_timeout": 3600}`
	p.Settings = settingsJSON

	settings, err = p.GetSettings()
	if err != nil {
		t.Errorf("GetSettings() error = %v", err)
	}
	if settings.AllowUserRegistration {
		t.Error("GetSettings() AllowUserRegistration should be false")
	}
	if settings.MaxConcurrentSessions != 3 {
		t.Errorf("GetSettings() MaxConcurrentSessions = %v, want 3", settings.MaxConcurrentSessions)
	}

	// Test invalid JSON
	p.Settings = "invalid json"
	_, err = p.GetSettings()
	if err == nil {
		t.Error("GetSettings() expected error for invalid JSON")
	}
}

func TestProvider_SetSettings(t *testing.T) {
	p := Provider{}

	settings := &ProviderSettings{
		AllowUserRegistration: true,
		AllowVoucherCreation:  false,
		MaxConcurrentSessions: 5,
		SessionTimeout:        7200,
		IdleTimeout:          1800,
	}

	err := p.SetSettings(settings)
	if err != nil {
		t.Errorf("SetSettings() error = %v", err)
	}

	// Verify by unmarshaling
	var result ProviderSettings
	err = json.Unmarshal([]byte(p.Settings), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal settings: %v", err)
	}
	if result.AllowUserRegistration != settings.AllowUserRegistration {
		t.Errorf("AllowUserRegistration = %v, want %v", result.AllowUserRegistration, settings.AllowUserRegistration)
	}
	if result.MaxConcurrentSessions != settings.MaxConcurrentSessions {
		t.Errorf("MaxConcurrentSessions = %v, want %v", result.MaxConcurrentSessions, settings.MaxConcurrentSessions)
	}
}

func TestProvider_IsActive(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"active", true},
		{"suspended", false},
		{"inactive", false},
		{"", false},
	}

	for _, tt := range tests {
		p := Provider{Status: tt.status}
		got := p.IsActive()
		if got != tt.want {
			t.Errorf("IsActive() for status %q = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestProvider_IsSuspended(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"suspended", true},
		{"active", false},
		{"inactive", false},
		{"", false},
	}

	for _, tt := range tests {
		p := Provider{Status: tt.status}
		got := p.IsSuspended()
		if got != tt.want {
			t.Errorf("IsSuspended() for status %q = %v, want %v", tt.status, got, tt.want)
		}
	}
}
