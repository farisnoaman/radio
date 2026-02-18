package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

// setupTestDB is defined in test_helpers.go


func TestPrivacy_ExportAnonymize(t *testing.T) {
	db := setupTestDB(t)


	// Seed data
	user := domain.RadiusUser{
		Username: "testuser",
		Realname: "John Doe",
		Mobile:   "1234567890",
		Status:   "enabled",
		CreatedAt: time.Now(),
	}
	db.Create(&user)

	e := echo.New()

	// Test Export
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/privacy/user/:username/export")
	c.SetParamNames("username")
	c.SetParamValues("testuser")
	c.Set("db", db) // Inject DB

	if assert.NoError(t, ExportUserData(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var export UserDataExport
		err := json.Unmarshal(rec.Body.Bytes(), &export)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", export.User.Username)
	}

	// Test Anonymize
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/privacy/user/:username/anonymize")
	c.SetParamNames("username")
	c.SetParamValues("testuser")
	c.Set("db", db)
	// Mock operator for log
	c.Set("current_operator", &domain.SysOpr{ID: 1, Username: "admin"})

	if assert.NoError(t, AnonymizeUser(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var updatedUser domain.RadiusUser
		db.First(&updatedUser, user.ID)
		assert.Equal(t, "Anonymized User", updatedUser.Realname)
		assert.Empty(t, updatedUser.Mobile)
	}
}
