package adminapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestCreateServer(t *testing.T) {
	c, db, _ := CreateTestContextWithApp(t, httptest.NewRequest(http.MethodPost, "/", nil), httptest.NewRecorder())

	payload := serverPayload{
		Name:          "Test Mikrotik",
		PublicIP:      "192.168.1.1",
		Secret:        "radius_secret",
		Username:      "api_user",
		Password:      "api_pass",
		RouterLimit:   "100M/100M",
		DBHost:        "10.0.0.1",
		DBPort:        3306,
		DBName:        "radiusdb",
		DBUsername:    "dbuser",
		DBPassword:    "dbpass",
		RouterStatus:  "online",
		OnlineHotspot: 5,
		OnlinePPPoE:   12,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/network/servers", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.SetRequest(req)
	rec := httptest.NewRecorder()
	c = c.Echo().NewContext(req, rec)
	c.Set("db", db)

	err := CreateServer(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var res struct {
		Data domain.Server `json:"data"`
	}
	t.Logf("Response body: %s", rec.Body.String())
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	serverResp := res.Data
	if serverResp.ID == 0 {
		// Just in case it's not wrapped in {"data": ...}
		err = json.Unmarshal(rec.Body.Bytes(), &serverResp)
		require.NoError(t, err)
	}

	assert.Equal(t, "Test Mikrotik", serverResp.Name)
	assert.Equal(t, "192.168.1.1", serverResp.PublicIP)
	assert.Equal(t, "100M/100M", serverResp.RouterLimit)
	assert.Equal(t, 12, serverResp.OnlinePPPoE)

	// Verify in DB
	var dbServer domain.Server
	db.First(&dbServer, serverResp.ID)
	assert.Equal(t, "radiusdb", dbServer.DBName)
}
