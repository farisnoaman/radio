package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/tenant"
)

func TestTenantMiddleware(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		headerValue    string
		defaultTenant  int64
		expectedTenant int64
		expectError    bool
	}{
		{
			name:           "valid tenant from header",
			headerValue:    "123",
			defaultTenant:  0,
			expectedTenant: 123,
			expectError:    false,
		},
		{
			name:           "invalid tenant header",
			headerValue:    "invalid",
			defaultTenant:  0,
			expectedTenant: 0,
			expectError:    true,
		},
		{
			name:           "negative tenant header",
			headerValue:    "-1",
			defaultTenant:  0,
			expectedTenant: 0,
			expectError:    true,
		},
		{
			name:           "empty header with default",
			headerValue:    "",
			defaultTenant:  1,
			expectedTenant: 1,
			expectError:    false,
		},
		{
			name:           "empty header no default",
			headerValue:    "",
			defaultTenant:  0,
			expectedTenant: 0,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.headerValue != "" {
				req.Header.Set(TenantIDHeader, tt.headerValue)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := TenantMiddleware(TenantMiddlewareConfig{
				DefaultTenant: tt.defaultTenant,
				SkipPaths:    []string{"/skip"},
			})

			var capturedTenant int64
			handler := middleware(func(c echo.Context) error {
				capturedTenant, _ = tenant.FromContext(c.Request().Context())
				return c.String(http.StatusOK, "OK")
			})

			err := handler(c)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				he, ok := err.(*echo.HTTPError)
				if !ok {
					t.Errorf("Expected HTTPError, got %T", err)
				}
				if he.Code != http.StatusBadRequest {
					t.Errorf("Expected status 400, got %d", he.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if capturedTenant != tt.expectedTenant {
					t.Errorf("Captured tenant = %d, want %d", capturedTenant, tt.expectedTenant)
				}
			}
		})
	}
}

func TestTenantMiddlewareSkipPath(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/skip", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := TenantMiddleware(TenantMiddlewareConfig{
		SkipPaths: []string{"/skip"},
	})

	var called bool
	handler := middleware(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !called {
		t.Error("Handler should have been called for skip path")
	}
}

func TestTenantMiddlewareFromOperator(t *testing.T) {
	e := echo.New()

	t.Run("with operator tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		getTenantID := func() (int64, error) {
			return 456, nil
		}

		middleware := TenantMiddlewareFromOperator(getTenantID)

		var capturedTenant int64
		handler := middleware(func(c echo.Context) error {
			capturedTenant, _ = tenant.FromContext(c.Request().Context())
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if capturedTenant != 456 {
			t.Errorf("Captured tenant = %d, want 456", capturedTenant)
		}
	})

	t.Run("operator returns zero", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		getTenantID := func() (int64, error) {
			return 0, nil
		}

		middleware := TenantMiddlewareFromOperator(getTenantID)

		handler := middleware(func(c echo.Context) error {
			_, err := tenant.FromContext(c.Request().Context())
			if err == nil {
				t.Error("Expected error when no tenant context")
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("operator returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		getTenantID := func() (int64, error) {
			return 0, echo.NewHTTPError(http.StatusUnauthorized, "no tenant")
		}

		middleware := TenantMiddlewareFromOperator(getTenantID)

		handler := middleware(func(c echo.Context) error {
			_, err := tenant.FromContext(c.Request().Context())
			if err == nil {
				t.Error("Expected error when no tenant context")
			}
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
