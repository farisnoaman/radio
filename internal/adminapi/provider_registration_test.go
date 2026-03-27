package adminapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateProviderRegistration(t *testing.T) {
	// Setup
	e := echo.New()
	reqBody := map[string]interface{}{
		"company_name":   "Test ISP LLC",
		"contact_name":   "John Doe",
		"email":          "john@testisp.com",
		"phone":          "+1234567890",
		"business_type":  "WISP",
		"expected_users": 500,
		"expected_nas":   10,
		"country":        "US",
		"message":        "We want to join your platform",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := CreateProviderRegistration(c)

	// Assert
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["status"] != "pending" {
		t.Errorf("Expected status 'pending', got '%v'", response["status"])
	}
}

func TestCreateProviderRegistrationValidation(t *testing.T) {
	tests := []struct {
		name       string
		payload    map[string]interface{}
		expectCode int
	}{
		{
			name: "missing company name",
			payload: map[string]interface{}{
				"contact_name": "John Doe",
				"email":        "john@test.com",
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			payload: map[string]interface{}{
				"company_name": "Test ISP",
				"contact_name": "John Doe",
				"email":        "not-an-email",
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			body, _ := json.Marshal(tt.payload)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/public/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := CreateProviderRegistration(c)

			if err == nil {
				t.Error("Expected validation error")
			}
		})
	}
}

func TestGetRegistrationStatus(t *testing.T) {
	// This test would require a database setup
	// For now, we'll skip it
	t.Skip("Requires database setup")
}
