package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// Helper to setup DB with system log table
func setupTestDBForSystemLogs(t *testing.T) *gorm.DB {
	// Re-using setupTestDB from test_helpers.go (assuming it's in the same package)
	db := setupTestDB(t)
	err := db.AutoMigrate(
		&domain.SysOprLog{},
	)
	assert.NoError(t, err)
	return db
}

func TestLogOperation(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForSystemLogs(t)

	t.Run("LogOperation Async Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Username: "admin"})

		// Trigger log
		LogOperation(c, "test_action", "test description")

		// Wait for goroutine to finish (simple sleep for test stability)
		time.Sleep(100 * time.Millisecond)

		// Verify log exists
		var log domain.SysOprLog
		err := db.First(&log, "opt_action = ?", "test_action").Error
		assert.NoError(t, err)
		assert.Equal(t, "admin", log.OprName)
		assert.Equal(t, "test_action", log.OptAction)
		assert.Equal(t, "test description", log.OptDesc)
	})

	t.Run("LogOperation System User Fallback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("db", db)
		// No operator in context

		LogOperation(c, "system_action", "system desc")

		time.Sleep(100 * time.Millisecond)

		var log domain.SysOprLog
		err := db.First(&log, "opt_action = ?", "system_action").Error
		assert.NoError(t, err)
		assert.Equal(t, "system", log.OprName)
	})
}

func TestListSystemLogs(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForSystemLogs(t)

	// Seed logs
	logs := []domain.SysOprLog{
		{OprName: "admin", OptAction: "login", OptDesc: "Admin login", OptTime: time.Now().Add(-1 * time.Hour)},
		{OprName: "admin", OptAction: "delete", OptDesc: "Deleted user", OptTime: time.Now()},
		{OprName: "operator", OptAction: "view", OptDesc: "View dashboard", OptTime: time.Now().Add(-2 * time.Hour)},
	}
	db.Create(&logs)

	t.Run("List Logs Success (Admin)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/logs?page=1&perPage=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})

		if assert.NoError(t, ListSystemLogs(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			
			var resp struct {
				Data []domain.SysOprLog `json:"data"`
				Meta Meta              `json:"meta"`
			}
			json.Unmarshal(rec.Body.Bytes(), &resp)
			
			assert.Len(t, resp.Data, 3)

		}
	})

	t.Run("List Logs Filter by Operator", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/logs?operator=operator", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})

		if assert.NoError(t, ListSystemLogs(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			
			var resp struct {
				Meta Meta `json:"meta"`
			}
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, int64(1), resp.Meta.Total)

		}
	})

	t.Run("List Logs Filter by Action", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/logs?action=login", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"})

		if assert.NoError(t, ListSystemLogs(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			
			var resp struct {
				Meta Meta `json:"meta"`
			}
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, int64(1), resp.Meta.Total)

		}
	})

	t.Run("List Logs Permission Denied (Operator)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/logs", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("db", db)
		c.Set("current_operator", &domain.SysOpr{ID: 2, Level: "operator", Username: "opr"})

		_ = ListSystemLogs(c)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
