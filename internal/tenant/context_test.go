package tenant

import (
	"context"
	"testing"
)

func TestFromContext(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantID    int64
		wantError error
	}{
		{
			name:      "valid tenant ID",
			ctx:       WithTenantID(context.Background(), 123),
			wantID:    123,
			wantError: nil,
		},
		{
			name:      "no tenant in context",
			ctx:       context.Background(),
			wantID:    0,
			wantError: ErrNoTenant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromContext(tt.ctx)
			if err != tt.wantError {
				t.Errorf("FromContext() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.wantID {
				t.Errorf("FromContext() = %v, want %v", got, tt.wantID)
			}
		})
	}
}

func TestWithTenantID(t *testing.T) {
	ctx := WithTenantID(context.Background(), 456)
	
	got, err := FromContext(ctx)
	if err != nil {
		t.Errorf("FromContext() error = %v", err)
		return
	}
	if got != 456 {
		t.Errorf("FromContext() = %v, want 456", got)
	}
}

func TestWithTenantIDPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("WithTenantID should panic for tenantID <= 0")
		}
	}()
	WithTenantID(context.Background(), 0)
}

func TestMustFromContext(t *testing.T) {
	ctx := WithTenantID(context.Background(), 789)
	got := MustFromContext(ctx)
	if got != 789 {
		t.Errorf("MustFromContext() = %v, want 789", got)
	}
}

func TestMustFromContextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustFromContext should panic when no tenant in context")
		}
	}()
	MustFromContext(context.Background())
}

func TestGetTenantIDOrDefault(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID int64
	}{
		{
			name:   "with tenant",
			ctx:    WithTenantID(context.Background(), 100),
			wantID: 100,
		},
		{
			name:   "without tenant returns default",
			ctx:    context.Background(),
			wantID: DefaultTenantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTenantIDOrDefault(tt.ctx)
			if got != tt.wantID {
				t.Errorf("GetTenantIDOrDefault() = %v, want %v", got, tt.wantID)
			}
		})
	}
}

func TestValidateTenantID(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		wantErr error
	}{
		{"valid positive", 1, nil},
		{"valid large", 999999, nil},
		{"zero", 0, ErrInvalidTenant},
		{"negative", -1, ErrInvalidTenant},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTenantID(tt.id)
			if err != tt.wantErr {
				t.Errorf("ValidateTenantID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTenantContext(t *testing.T) {
	ctx := context.Background()
	
	tc, err := NewTenantContext(ctx, 42)
	if err != nil {
		t.Errorf("NewTenantContext() error = %v", err)
		return
	}
	
	if tc.TenantID != 42 {
		t.Errorf("tc.TenantID = %v, want 42", tc.TenantID)
	}
	
	extractedCtx := tc.Extract()
	got, _ := FromContext(extractedCtx)
	if got != 42 {
		t.Errorf("FromContext() = %v, want 42", got)
	}
}

func TestNewTenantContextInvalid(t *testing.T) {
	_, err := NewTenantContext(context.Background(), 0)
	if err != ErrInvalidTenant {
		t.Errorf("NewTenantContext() error = %v, want %v", err, ErrInvalidTenant)
	}
}

func TestTenantChecker_IsSystemTenant(t *testing.T) {
	checker := NewTenantChecker()
	
	tests := []struct {
		id   int64
		want bool
	}{
		{1, true},
		{0, false},
		{-1, false},
		{999, false},
	}
	
	for _, tt := range tests {
		got := checker.IsSystemTenant(tt.id)
		if got != tt.want {
			t.Errorf("IsSystemTenant(%d) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestTenantChecker_CanAccess(t *testing.T) {
	checker := NewTenantChecker()
	
	tests := []struct {
		name    string
		source  int64
		target  int64
		canAccess bool
	}{
		{"same tenant", 5, 5, true},
		{"system accessing any", 1, 5, true},
		{"different tenants", 2, 3, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.CanAccess(tt.source, tt.target)
			if got != tt.canAccess {
				t.Errorf("CanAccess(%d, %d) = %v, want %v", tt.source, tt.target, got, tt.canAccess)
			}
		})
	}
}
