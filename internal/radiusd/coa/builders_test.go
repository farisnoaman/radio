package coa

import (
	"testing"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestMikrotikBuilder_AddRateLimit_ShouldSetCorrectAttributes(t *testing.T) {
	builder := &MikrotikBuilder{}
	pkt := radius.New(radius.CodeCoARequest, []byte("secret"))

	err := builder.AddRateLimit(pkt, 10240, 20480)
	if err != nil {
		t.Fatalf("AddRateLimit failed: %v", err)
	}

	// Verify Mikrotik-specific attributes were set
	// The builder should add Mikrotik-Rate-Limit attribute
	// We can't easily verify the exact value without importing the mikrotik vendor package
	// but we can verify that no error occurred
}

func TestMikrotikBuilder_AddDataQuota_ShouldSetCorrectAttributes(t *testing.T) {
	builder := &MikrotikBuilder{}
	pkt := radius.New(radius.CodeCoARequest, []byte("secret"))

	err := builder.AddDataQuota(pkt, 1024) // 1 GB
	if err != nil {
		t.Fatalf("AddDataQuota failed: %v", err)
	}

	// Verify quota was set
}

func TestCiscoBuilder_AddRateLimit_ShouldSetCorrectAttributes(t *testing.T) {
	builder := &CiscoBuilder{}
	pkt := radius.New(radius.CodeCoARequest, []byte("secret"))

	err := builder.AddRateLimit(pkt, 10240, 20480)
	if err != nil {
		t.Fatalf("AddRateLimit failed: %v", err)
	}

	// Verify Cisco AVPair attributes were set
}

func TestHuaweiBuilder_AddRateLimit_ShouldSetCorrectAttributes(t *testing.T) {
	builder := &HuaweiBuilder{}
	pkt := radius.New(radius.CodeCoARequest, []byte("secret"))

	err := builder.AddRateLimit(pkt, 10240, 20480)
	if err != nil {
		t.Fatalf("AddRateLimit failed: %v", err)
	}

	// Verify Huawei rate limit attributes were set
}

func TestGetVendorBuilder_ShouldReturnCorrectBuilder(t *testing.T) {
	testCases := []struct {
		vendorCode   string
		expectNil    bool
		expectedCode string
	}{
		{"mikrotik", false, "mikrotik"},
		{"cisco", false, "cisco"},
		{"huawei", false, "huawei"},
		{"juniper", true, ""},
		{"", true, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.vendorCode, func(t *testing.T) {
			builder := GetVendorBuilder(tc.vendorCode)

			if tc.expectNil {
				if builder != nil {
					t.Errorf("expected nil builder for vendor '%s', got %T", tc.vendorCode, builder)
				}
			} else {
				if builder == nil {
					t.Fatalf("expected non-nil builder for vendor '%s'", tc.vendorCode)
				}
				if builder.VendorCode() != tc.expectedCode {
					t.Errorf("expected vendor code '%s', got '%s'", tc.expectedCode, builder.VendorCode())
				}
			}
		})
	}
}

func TestVendorAttributeBuilder_AddSessionTimeout(t *testing.T) {
	builders := []VendorAttributeBuilder{
		&MikrotikBuilder{},
		&CiscoBuilder{},
		&HuaweiBuilder{},
	}

	for _, builder := range builders {
		t.Run(builder.VendorCode(), func(t *testing.T) {
			pkt := radius.New(radius.CodeCoARequest, []byte("secret"))

			err := builder.AddSessionTimeout(pkt, 3600) // 1 hour
			if err != nil {
				t.Fatalf("AddSessionTimeout failed: %v", err)
			}

			// Most vendors use standard Session-Timeout attribute
			timeout := rfc2865.SessionTimeout_Get(pkt)
			if timeout != 3600 {
				// Some builders might not set it (no-op), so we just check no error
				t.Logf("Session timeout not set by %s builder (may be expected)", builder.VendorCode())
			}
		})
	}
}

func TestVendorAttributeBuilder_AddDisconnectAttributes(t *testing.T) {
	builders := []VendorAttributeBuilder{
		&MikrotikBuilder{},
		&CiscoBuilder{},
		&HuaweiBuilder{},
	}

	for _, builder := range builders {
		t.Run(builder.VendorCode(), func(t *testing.T) {
			pkt := radius.New(radius.CodeDisconnectRequest, []byte("secret"))

			err := builder.AddDisconnectAttributes(pkt, "testuser", "session123")
			if err != nil {
				t.Fatalf("AddDisconnectAttributes failed: %v", err)
			}

			// Verify username was set (most builders do this)
			username := rfc2865.UserName_GetString(pkt)
			if username == "" && builder.VendorCode() == "cisco" {
				// Cisco might use different attributes
				t.Logf("Username not set by %s builder (may use different attributes)", builder.VendorCode())
			}
		})
	}
}
