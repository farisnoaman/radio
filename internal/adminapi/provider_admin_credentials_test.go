package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// createTestProvider creates a test provider for testing
func createTestProvider(db *gorm.DB, code string, name string) *domain.Provider {
	provider := &domain.Provider{
		ID:       common.UUIDint64(),
		Code:     code,
		Name:     name,
		Status:   "active",
		MaxUsers: 1000,
		MaxNas:   100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(provider)
	return provider
}

// createTestProviderAdmin creates a test admin for a provider
func createTestProviderAdmin(db *gorm.DB, providerID int64, username string, password string) *domain.SysOpr {
	admin := &domain.SysOpr{
		ID:        common.UUIDint64(),
		TenantID:  providerID,
		Username:  username,
		Password:  common.Sha256HashWithSalt(password, common.GetSecretSalt()),
		Level:     "admin",
		Status:    "enabled",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(admin)
	return admin
}

// TestGetProviderAdminCredentials tests retrieving provider admin credentials
func TestGetProviderAdminCredentials(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	_ = db.AutoMigrate(&domain.Provider{}) //nolint:errcheck

	tests := []struct {
		name           string
		setup          func(*testing.T, *gorm.DB) int64
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "Get credentials for provider with existing admin",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				createTestProviderAdmin(db, provider.ID, "provideradmin", "AdminPass123")
				return provider.ID
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "provideradmin", resp["username"])
				assert.Equal(t, "********", resp["password"])
				assert.Equal(t, "admin", resp["level"])
				assert.Equal(t, "enabled", resp["status"])
			},
		},
		{
			name: "Get credentials for provider without admin returns defaults",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov2", "Provider 2")
				return provider.ID
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "admin", resp["username"])
				assert.Equal(t, "********", resp["password"])
				assert.Equal(t, "admin", resp["level"])
				assert.Equal(t, "not_created", resp["status"])
			},
		},
		{
			name: "Get credentials for non-existent provider returns 404",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				return 999999
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "PROVIDER_NOT_FOUND",
		},
		{
			name: "Get credentials with invalid provider ID",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				return 0 // Will cause parse error
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerID := tt.setup(t, db)

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/platform/providers/"+strconv.FormatInt(providerID, 10)+"/admin-credentials", nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(providerID, 10))

			err := GetProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				if tt.checkResponse != nil {
					tt.checkResponse(t, response)
				}
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestUpdateProviderAdminCredentials tests updating provider admin credentials
func TestUpdateProviderAdminCredentials(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	_ = db.AutoMigrate(&domain.Provider{}) //nolint:errcheck

	tests := []struct {
		name           string
		setup          func(*testing.T, *gorm.DB) int64
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, map[string]interface{})
		verifyDB       func(*testing.T, *gorm.DB, int64)
	}{
		{
			name: "Update credentials for existing provider admin",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				createTestProviderAdmin(db, provider.ID, "oldadmin", "OldPass123")
				return provider.ID
			},
			requestBody: `{
				"username": "newadmin",
				"password": "NewSecurePass456"
			}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "newadmin", resp["username"])
				assert.NotEmpty(t, resp["password"])
				assert.Equal(t, "admin", resp["level"])
				assert.Equal(t, "enabled", resp["status"])
				assert.Contains(t, resp["message"], "Credentials updated successfully")
			},
			verifyDB: func(t *testing.T, db *gorm.DB, providerID int64) {
				var admin domain.SysOpr
				err := db.Where("tenant_id = ? AND level = ?", providerID, "admin").First(&admin).Error
				require.NoError(t, err)
				assert.Equal(t, "newadmin", admin.Username)
				expectedHash := common.Sha256HashWithSalt("NewSecurePass456", common.GetSecretSalt())
				assert.Equal(t, expectedHash, admin.Password)
			},
		},
		{
			name: "Create new admin credentials for provider without admin",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov2", "Provider 2")
				return provider.ID
			},
			requestBody: `{
				"username": "firstadmin",
				"password": "FirstPass123"
			}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "firstadmin", resp["username"])
				assert.NotEmpty(t, resp["password"])
				assert.Equal(t, "admin", resp["level"])
				assert.Equal(t, "enabled", resp["status"])
			},
			verifyDB: func(t *testing.T, db *gorm.DB, providerID int64) {
				var admin domain.SysOpr
				err := db.Where("tenant_id = ? AND level = ?", providerID, "admin").First(&admin).Error
				require.NoError(t, err)
				assert.Equal(t, "firstadmin", admin.Username)
			},
		},
		{
			name: "Username conflict with another provider's admin",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider1 := createTestProvider(db, "prov1", "Provider 1")
				createTestProviderAdmin(db, provider1.ID, "conflictuser", "Pass123")
				provider2 := createTestProvider(db, "prov2", "Provider 2")
				return provider2.ID
			},
			requestBody: `{
				"username": "conflictuser",
				"password": "NewPass123"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name: "Non-existent provider returns 404",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				return 999999
			},
			requestBody: `{
				"username": "testadmin",
				"password": "TestPass123"
			}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "PROVIDER_NOT_FOUND",
		},
		{
			name:           "Invalid provider ID",
			setup:          func(t *testing.T, db *gorm.DB) int64 { return 0 },
			requestBody:    `{"username": "test", "password": "TestPass123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name: "Username too short",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				return provider.ID
			},
			requestBody: `{
				"username": "ab",
				"password": "TestPass123"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USERNAME",
		},
		{
			name: "Password too short",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				return provider.ID
			},
			requestBody: `{
				"username": "testadmin",
				"password": "12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PASSWORD",
		},
		{
			name: "Missing username",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				return provider.ID
			},
			requestBody:    `{"password": "TestPass123"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing password",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				return provider.ID
			},
			requestBody:    `{"username": "testadmin"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid JSON",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				return provider.ID
			},
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerID := tt.setup(t, db)

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/platform/providers/"+strconv.FormatInt(providerID, 10)+"/admin-credentials", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(providerID, 10))

			err := UpdateProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				if tt.checkResponse != nil {
					tt.checkResponse(t, response)
				}
				if tt.verifyDB != nil {
					tt.verifyDB(t, db, providerID)
				}
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestResetProviderAdminCredentials tests resetting provider admin credentials to defaults
func TestResetProviderAdminCredentials(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	_ = db.AutoMigrate(&domain.Provider{}) //nolint:errcheck

	defaultUsername := "admin"
	defaultPassword := "123456"

	tests := []struct {
		name           string
		setup          func(*testing.T, *gorm.DB) int64
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, map[string]interface{})
		verifyDB       func(*testing.T, *gorm.DB, int64)
	}{
		{
			name: "Reset credentials for provider with existing admin",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov1", "Provider 1")
				createTestProviderAdmin(db, provider.ID, "customadmin", "CustomPass123")
				return provider.ID
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, defaultUsername, resp["username"])
				assert.NotEmpty(t, resp["password"])
				assert.Equal(t, "admin", resp["level"])
				assert.Equal(t, "enabled", resp["status"])
				assert.Contains(t, resp["message"], "Credentials reset to defaults")
			},
			verifyDB: func(t *testing.T, db *gorm.DB, providerID int64) {
				var admin domain.SysOpr
				err := db.Where("tenant_id = ? AND level = ?", providerID, "admin").First(&admin).Error
				require.NoError(t, err)
				assert.Equal(t, defaultUsername, admin.Username)
				expectedHash := common.Sha256HashWithSalt(defaultPassword, common.GetSecretSalt())
				assert.Equal(t, expectedHash, admin.Password)
			},
		},
		{
			name: "Reset creates admin for provider without one",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				provider := createTestProvider(db, "prov2", "Provider 2")
				return provider.ID
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, defaultUsername, resp["username"])
				assert.NotEmpty(t, resp["password"])
				assert.Equal(t, "admin", resp["level"])
				assert.Equal(t, "enabled", resp["status"])
			},
			verifyDB: func(t *testing.T, db *gorm.DB, providerID int64) {
				var admin domain.SysOpr
				err := db.Where("tenant_id = ? AND level = ?", providerID, "admin").First(&admin).Error
				require.NoError(t, err)
				assert.Equal(t, defaultUsername, admin.Username)
			},
		},
		{
			name: "Reset for non-existent provider returns 404",
			setup: func(t *testing.T, db *gorm.DB) int64 {
				return 999999
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "PROVIDER_NOT_FOUND",
		},
		{
			name:           "Invalid provider ID",
			setup:          func(t *testing.T, db *gorm.DB) int64 { return 0 },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerID := tt.setup(t, db)

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/platform/providers/"+strconv.FormatInt(providerID, 10)+"/admin-credentials/reset", nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(providerID, 10))

			err := ResetProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				if tt.checkResponse != nil {
					tt.checkResponse(t, response)
				}
				if tt.verifyDB != nil {
					tt.verifyDB(t, db, providerID)
				}
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestProviderAdminCredentialsAuthorization tests that only super admins can access these endpoints
func TestProviderAdminCredentialsAuthorization(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	_ = db.AutoMigrate(&domain.Provider{}) //nolint:errcheck

	provider := createTestProvider(db, "prov1", "Provider 1")

	tests := []struct {
		name           string
		operatorLevel  string
		expectedStatus int
	}{
		{
			name:           "Super admin can access",
			operatorLevel:  "super",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Regular admin cannot access",
			operatorLevel:  "admin",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Operator cannot access",
			operatorLevel:  "operator",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" (GET)", func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/platform/providers/"+strconv.FormatInt(provider.ID, 10)+"/admin-credentials", nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(provider.ID, 10))

			// Override the operator level
			c.Set("current_operator", &domain.SysOpr{
				ID:       1,
				Username: "testuser",
				Level:    tt.operatorLevel,
				Status:   "enabled",
			})

			err := GetProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})

		t.Run(tt.name+" (PUT)", func(t *testing.T) {
			e := setupTestEcho()
			requestBody := `{"username": "testadmin", "password": "TestPass123"}`
			req := httptest.NewRequest(http.MethodPut, "/api/v1/platform/providers/"+strconv.FormatInt(provider.ID, 10)+"/admin-credentials", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(provider.ID, 10))

			// Override the operator level
			c.Set("current_operator", &domain.SysOpr{
				ID:       1,
				Username: "testuser",
				Level:    tt.operatorLevel,
				Status:   "enabled",
			})

			err := UpdateProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})

		t.Run(tt.name+" (POST reset)", func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/platform/providers/"+strconv.FormatInt(provider.ID, 10)+"/admin-credentials/reset", nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(provider.ID, 10))

			// Override the operator level
			c.Set("current_operator", &domain.SysOpr{
				ID:       1,
				Username: "testuser",
				Level:    tt.operatorLevel,
				Status:   "enabled",
			})

			err := ResetProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestProviderAdminCredentialsIntegration tests the full workflow
func TestProviderAdminCredentialsIntegration(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	_ = db.AutoMigrate(&domain.Provider{}) //nolint:errcheck

	t.Run("Full workflow: get, update, reset", func(t *testing.T) {
		// Create a provider
		provider := createTestProvider(db, "prov1", "Provider 1")

		e := setupTestEcho()

		// Step 1: Get credentials (should return defaults)
		t.Run("Step 1: Get initial credentials", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/platform/providers/"+strconv.FormatInt(provider.ID, 10)+"/admin-credentials", nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(provider.ID, 10))

			err := GetProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "admin", response["username"])
			assert.Equal(t, "not_created", response["status"])
		})

		// Step 2: Update credentials
		t.Run("Step 2: Update credentials", func(t *testing.T) {
			requestBody := `{"username": "myadmin", "password": "MySecurePass123"}`
			req := httptest.NewRequest(http.MethodPut, "/api/v1/platform/providers/"+strconv.FormatInt(provider.ID, 10)+"/admin-credentials", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(provider.ID, 10))

			err := UpdateProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "myadmin", response["username"])
			assert.NotEmpty(t, response["password"])
		})

		// Step 3: Get credentials again (should return masked password)
		t.Run("Step 3: Get credentials after update", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/platform/providers/"+strconv.FormatInt(provider.ID, 10)+"/admin-credentials", nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(provider.ID, 10))

			err := GetProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "myadmin", response["username"])
			assert.Equal(t, "********", response["password"])
			assert.Equal(t, "enabled", response["status"])
		})

		// Step 4: Reset credentials
		t.Run("Step 4: Reset credentials to defaults", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/platform/providers/"+strconv.FormatInt(provider.ID, 10)+"/admin-credentials/reset", nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(provider.ID, 10))

			err := ResetProviderAdminCredentials(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "admin", response["username"])
		})

		// Step 5: Verify in database
		t.Run("Step 5: Verify database state", func(t *testing.T) {
			var admin domain.SysOpr
			err := db.Where("tenant_id = ? AND level = ?", provider.ID, "admin").First(&admin).Error
			require.NoError(t, err)
			assert.Equal(t, "admin", admin.Username)
			expectedHash := common.Sha256HashWithSalt("123456", common.GetSecretSalt())
			assert.Equal(t, expectedHash, admin.Password)
		})
	})
}
