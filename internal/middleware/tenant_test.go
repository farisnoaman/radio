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
			expectError:    true, // Now returns error when no tenant context
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
				// "empty header no default" returns 401, others return 400
				expectedStatus := http.StatusBadRequest
				if tt.name == "empty header no default" {
					expectedStatus = http.StatusUnauthorized
				}
				if he.Code != expectedStatus {
					t.Errorf("Expected status %d, got %d", expectedStatus, he.Code)
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

func TestRequireTenant(t *testing.T) {
	e := echo.New()

	t.Run("with tenant context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Add tenant to context
		ctx := tenant.WithTenantID(req.Context(), 789)
		c.SetRequest(req.WithContext(ctx))

		middleware := RequireTenant()
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
			t.Error("Handler should have been called")
		}
	})

	t.Run("without tenant context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		middleware := RequireTenant()
		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		if err == nil {
			t.Error("Expected error when no tenant context")
		}

		he, ok := err.(*echo.HTTPError)
		if !ok {
			t.Errorf("Expected HTTPError, got %T", err)
		}
		if he.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", he.Code)
		}
	})
}
