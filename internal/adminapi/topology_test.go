package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"


	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestNetwork_Topology(t *testing.T) {
	// setupTestDB is defined in test_helpers.go
	db := setupTestDB(t)


	db.AutoMigrate(&domain.NetNode{}, &domain.NetNas{})

	// Seed data
	node := domain.NetNode{
		Name:      "Core Node",
		Latitude:  1.23,
		Longitude: 4.56,
	}
	db.Create(&node)

	nas := domain.NetNas{
		Name:   "NAS-1",
		NodeId: node.ID,
		Status: "active",
	}
	db.Create(&nas)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("db", db)

	if assert.NoError(t, GetNetworkTopology(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var resp struct {
			Data []*TopologyNode `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "Core Node", resp.Data[0].Name)
		assert.Equal(t, 1.23, resp.Data[0].Lat)
		assert.Len(t, resp.Data[0].Children, 1)
		assert.Equal(t, "NAS-1", resp.Data[0].Children[0].Name)

	}
}
