package domain

import (
	"encoding/json"
	"testing"
	"time"
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

func TestProviderRegistrationTableName(t *testing.T) {
	reg := ProviderRegistration{}
	if got := reg.TableName(); got != "mst_provider_registration" {
		t.Errorf("TableName() = %v, want mst_provider_registration", got)
	}
}

func TestProviderRegistrationModel(t *testing.T) {
	now := time.Now()
	reg := &ProviderRegistration{
		ID:            1,
		CompanyName:   "Test ISP LLC",
		ContactName:   "John Doe",
		Email:         "john@testisp.com",
		Phone:         "+1234567890",
		Address:       "123 Main St, City, Country",
		BusinessType:  "ISP",
		ExpectedUsers: 500,
		ExpectedNas:   10,
		Country:       "US",
		Message:       "We would like to join your platform",
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if reg.CompanyName != "Test ISP LLC" {
		t.Errorf("Expected CompanyName 'Test ISP LLC', got '%s'", reg.CompanyName)
	}

	if reg.Email != "john@testisp.com" {
		t.Errorf("Expected Email 'john@testisp.com', got '%s'", reg.Email)
	}

	if reg.Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", reg.Status)
	}

	if reg.TableName() != "mst_provider_registration" {
		t.Errorf("Expected table name 'mst_provider_registration', got '%s'", reg.TableName())
	}
}

func TestProviderRegistrationWithReview(t *testing.T) {
	now := time.Now()
	reviewTime := now.Add(24 * time.Hour)
	reg := &ProviderRegistration{
		ID:              1,
		CompanyName:     "Test ISP LLC",
		ContactName:     "John Doe",
		Email:           "john@testisp.com",
		Status:          "approved",
		ReviewedBy:      100,
		ReviewedAt:      &reviewTime,
		RejectionReason: "",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if reg.Status != "approved" {
		t.Errorf("Expected Status 'approved', got '%s'", reg.Status)
	}

	if reg.ReviewedBy != 100 {
		t.Errorf("Expected ReviewedBy 100, got %d", reg.ReviewedBy)
	}

	if reg.ReviewedAt == nil {
		t.Error("Expected ReviewedAt to be set, got nil")
	}

	if reg.RejectionReason != "" {
		t.Errorf("Expected empty RejectionReason, got '%s'", reg.RejectionReason)
	}
}

func TestProviderRegistrationRejected(t *testing.T) {
	now := time.Now()
	reviewTime := now.Add(24 * time.Hour)
	reg := &ProviderRegistration{
		ID:              2,
		CompanyName:     "Bad ISP",
		ContactName:     "Jane Doe",
		Email:           "jane@badisp.com",
		Status:          "rejected",
		ReviewedBy:      100,
		ReviewedAt:      &reviewTime,
		RejectionReason: "Incomplete documentation",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if reg.Status != "rejected" {
		t.Errorf("Expected Status 'rejected', got '%s'", reg.Status)
	}

	if reg.RejectionReason != "Incomplete documentation" {
		t.Errorf("Expected RejectionReason 'Incomplete documentation', got '%s'", reg.RejectionReason)
	}
}

func TestProviderBrandingSerialization(t *testing.T) {
	provider := &Provider{}

	branding := &ProviderBranding{
		LogoURL:        "https://example.com/logo.png",
		PrimaryColor:   "#007bff",
		SecondaryColor: "#6c757d",
		CompanyName:    "Test ISP",
		SupportEmail:   "support@testisp.com",
		SupportPhone:   "+1234567890",
	}

	err := provider.SetBranding(branding)
	if err != nil {
		t.Fatalf("Failed to set branding: %v", err)
	}

	retrieved, err := provider.GetBranding()
	if err != nil {
		t.Fatalf("Failed to get branding: %v", err)
	}

	if retrieved.LogoURL != branding.LogoURL {
		t.Errorf("LogoURL mismatch: got '%s', want '%s'", retrieved.LogoURL, branding.LogoURL)
	}
}
