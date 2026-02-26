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
	"github.com/talkincode/toughradius/v9/internal/acs"
	"gorm.io/gorm"
)

// createTestCPEDevice creates test CPE device data
func createTestCPEDevice(db *gorm.DB, serialNumber string, status acs.DeviceStatus) *acs.CPEDevice {
	// Ensure table exists
	db.AutoMigrate(&acs.CPEDevice{})
	
	now := time.Now()
	device := &acs.CPEDevice{
		SerialNumber:     serialNumber,
		OUI:             "001122",
		Manufacturer:    "Test Manufacturer",
		ProductClass:    "Router",
		Status:          status,
		AutoProvision:   true,
		LastInform:      &now,
		LastIP:          "192.168.1.100",
		SoftwareVersion: "1.0.0",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	db.Create(device)
	return device
}

func TestListCPEs(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Migrate CPEDevice table
	db.AutoMigrate(&acs.CPEDevice{})

	// Create test data
	createTestCPEDevice(db, "SN001", acs.DeviceStatusProvisioned)
	createTestCPEDevice(db, "SN002", acs.DeviceStatusPending)
	createTestCPEDevice(db, "SN003", acs.DeviceStatusFailed)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "List all CPE devices - default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 1, resp.Meta.Page)
			},
		},
		{
			name:           "Paginated query - page 1 with pageSize 2",
			queryParams:    "?page=1&pageSize=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 1, resp.Meta.Page)
				assert.Equal(t, 2, resp.Meta.PageSize)
			},
		},
		{
			name:           "Paginated query - page 2",
			queryParams:    "?page=2&pageSize=2",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 2, resp.Meta.Page)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/cpes"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := listCPEs(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// Convert the response data to a slice of CPE devices
			dataBytes, _ := json.Marshal(response.Data)
			var devices []acs.CPEDevice
			_ = json.Unmarshal(dataBytes, &devices) //nolint:errcheck
			response.Data = devices

			assert.Len(t, devices, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetCPE(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Migrate CPEDevice table
	db.AutoMigrate(&acs.CPEDevice{})

	// Create test data
	device := createTestCPEDevice(db, "SN001", acs.DeviceStatusProvisioned)

	tests := []struct {
		name           string
		cpeID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get existing CPE device",
			cpeID:          "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get missing CPE device",
			cpeID:          "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "CPE_NOT_FOUND",
		},
		{
			name:           "Invalid ID",
			cpeID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/cpes/"+tt.cpeID, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.cpeID)

			err := getCPE(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultDevice acs.CPEDevice
				_ = json.Unmarshal(dataBytes, &resultDevice) //nolint:errcheck

				assert.Equal(t, device.SerialNumber, resultDevice.SerialNumber)
				assert.Equal(t, device.Manufacturer, resultDevice.Manufacturer)
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse) //nolint:errcheck
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateCPE(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Migrate CPEDevice table
	db.AutoMigrate(&acs.CPEDevice{})

	// Create test data
	_ = createTestCPEDevice(db, "SN001", acs.DeviceStatusPending)

	tests := []struct {
		name           string
		cpeID          string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *acs.CPEDevice)
	}{
		{
			name:   "Successfully update CPE status",
			cpeID:  "1",
			requestBody: `{
				"status": "provisioned"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, device *acs.CPEDevice) {
				assert.Equal(t, acs.DeviceStatus("provisioned"), device.Status)
			},
		},
		{
			name:   "Update auto_provision flag",
			cpeID:  "1",
			requestBody: `{
				"auto_provision": false
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, device *acs.CPEDevice) {
				assert.Equal(t, false, device.AutoProvision)
			},
		},
		{
			name:   "Update profile_id",
			cpeID:  "1",
			requestBody: `{
				"profile_id": 123
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, device *acs.CPEDevice) {
				assert.Equal(t, int64(123), *device.ProfileID)
			},
		},
		{
			name:   "Multiple field update",
			cpeID:  "1",
			requestBody: `{
				"status": "disabled",
				"auto_provision": false,
				"profile_id": 456
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, device *acs.CPEDevice) {
				assert.Equal(t, acs.DeviceStatus("disabled"), device.Status)
				assert.Equal(t, false, device.AutoProvision)
				assert.Equal(t, int64(456), *device.ProfileID)
			},
		},
		{
			name:           "CPE not found",
			cpeID:          "999",
			requestBody:    `{"status": "provisioned"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "CPE_NOT_FOUND",
		},
		{
			name:           "Invalid ID",
			cpeID:          "invalid",
			requestBody:    `{"status": "provisioned"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:           "Invalid JSON",
			cpeID:          "1",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name:   "Partial update - only status",
			cpeID:  "1",
			requestBody: `{
				"status": "failed"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, device *acs.CPEDevice) {
				assert.Equal(t, acs.DeviceStatus("failed"), device.Status)
				// ProfileID should remain unchanged (was set to 456 in previous test)
				assert.NotNil(t, device.ProfileID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/cpes/"+tt.cpeID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.cpeID)

			err := updateCPE(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedDevice acs.CPEDevice
				_ = json.Unmarshal(dataBytes, &updatedDevice) //nolint:errcheck

				if tt.checkResult != nil {
					tt.checkResult(t, &updatedDevice)
				}
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse) //nolint:errcheck
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteCPE(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Migrate CPEDevice table
	db.AutoMigrate(&acs.CPEDevice{})

	// Create test data
	_ = createTestCPEDevice(db, "SN001", acs.DeviceStatusProvisioned)
	_ = createTestCPEDevice(db, "SN002", acs.DeviceStatusPending)

	tests := []struct {
		name           string
		cpeID          string
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:           "Successfully delete CPE device",
			cpeID:          "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "Delete non-existent CPE",
			cpeID:          "999",
			expectedStatus: http.StatusOK, // GORM Delete does not return error for non-existent
			checkDeleted:   false,
		},
		{
			name:           "Invalid ID",
			cpeID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Recreate devices for delete test
			if tt.name == "Successfully delete CPE device" {
				db.Exec("DELETE FROM cpe_device")
				_ = createTestCPEDevice(db, "SN001", acs.DeviceStatusProvisioned)
				_ = createTestCPEDevice(db, "SN002", acs.DeviceStatusPending)
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/cpes/"+tt.cpeID, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.cpeID)

			err := deleteCPE(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK && tt.checkDeleted {
				// Validate the device has been deleted
				var count int64
				db.Model(&acs.CPEDevice{}).Where("id = ?", tt.cpeID).Count(&count)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}

// TestCPEEdgeCases tests edge cases
func TestCPEEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Migrate CPEDevice table
	db.AutoMigrate(&acs.CPEDevice{})

	t.Run("List CPE devices ordered by last_inform DESC", func(t *testing.T) {
		// Create devices with different last_inform times
		now := time.Now()
		oldTime := now.Add(-time.Hour)
		
		device1 := &acs.CPEDevice{
			SerialNumber:  "SN001",
			Status:        acs.DeviceStatusProvisioned,
			LastInform:    &oldTime,
			AutoProvision: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		device2 := &acs.CPEDevice{
			SerialNumber:  "SN002",
			Status:        acs.DeviceStatusProvisioned,
			LastInform:    &now,
			AutoProvision: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		db.Create(device1)
		db.Create(device2)

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cpes", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := listCPEs(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var devices []acs.CPEDevice
		_ = json.Unmarshal(dataBytes, &devices)

		// Most recent should be first
		assert.Equal(t, "SN002", devices[0].SerialNumber)
		assert.Equal(t, "SN001", devices[1].SerialNumber)
	})

	t.Run("Get CPE after update reflects changes", func(t *testing.T) {
		// Clear and recreate to ensure clean state
		db.Exec("DELETE FROM cpe_device")
		device := createTestCPEDevice(db, "SN-UPDATE", acs.DeviceStatusPending)

		// Update the device
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/cpes/"+strconv.FormatInt(device.ID, 10), strings.NewReader(`{"status": "provisioned"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(device.ID, 10))

		err := updateCPE(c)
		require.NoError(t, err)

		// Now get the device
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/cpes/"+strconv.FormatInt(device.ID, 10), nil)
		rec2 := httptest.NewRecorder()
		c2 := CreateTestContext(e, db, req2, rec2, appCtx)
		c2.SetParamNames("id")
		c2.SetParamValues(strconv.FormatInt(device.ID, 10))

		err = getCPE(c2)
		require.NoError(t, err)

		var response Response
	_ = json.Unmarshal(rec2.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var resultDevice acs.CPEDevice
		_ = json.Unmarshal(dataBytes, &resultDevice)

		assert.Equal(t, acs.DeviceStatusProvisioned, resultDevice.Status)
		assert.Equal(t, device.SerialNumber, resultDevice.SerialNumber)
	})
}
