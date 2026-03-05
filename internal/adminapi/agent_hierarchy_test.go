package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// ---------------------------------------------------------------------------
// Helper for hierarchy tests
// ---------------------------------------------------------------------------

// setupTestDBForAgents returns a database prepared for agent-related tests.
// It simply reuses setupTestDB but includes the new hierarchy tables as well.
func setupTestDBForAgents(t *testing.T) *gorm.DB {
	return setupTestDB(t)
}

// createAgent creates a simple agent record in the database.
// It returns the created SysOpr object.
func createAgent(t *testing.T, db *gorm.DB, username string) domain.SysOpr {
	agent := domain.SysOpr{
		ID:       common.UUIDint64(),
		Username: username,
		Password: "pass",
		Level:    "agent",
		Status:   "enabled",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&agent).Error)
	return agent
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAssignAgentToParent_Success(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	// create two agents
	agent1 := createAgent(t, db, "agent1")
	agent2 := createAgent(t, db, "agent2")

	// admin operator
	admin := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	reqBody := AssignAgentToParentRequest{
		AgentID:        agent2.ID,
		ParentID:       &agent1.ID,
		CommissionRate: 0.05,
		Territory:      "RegionA",
	}
	jsonReq, err := json.Marshal(reqBody)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/agents/assign-parent", strings.NewReader(string(jsonReq))), rec)
	c.Set("db", db)
	c.Set("current_operator", admin)
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	err = AssignAgentToParent(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// verify hierarchy record
	var h domain.AgentHierarchy
	req := db.Where("agent_id = ?", agent2.ID).First(&h)
	require.NoError(t, req.Error)
	assert.Equal(t, agent1.ID, *h.ParentID)
	assert.Equal(t, 1, h.Level)
	assert.Equal(t, 0.05, h.CommissionRate)
	assert.Equal(t, "RegionA", h.Territory)
}

func TestAssignAgentToParent_Unauthorized(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	agent := createAgent(t, db, "agentX")
	nonAdmin := &domain.SysOpr{ID: 2, Level: "agent", Username: "agentX"}

	reqBody := AssignAgentToParentRequest{
		AgentID:        agent.ID,
		ParentID:       nil,
		CommissionRate: 0.1,
	}
	jsonReq, err := json.Marshal(reqBody)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/agents/assign-parent", strings.NewReader(string(jsonReq))), rec)
	c.Set("db", db)
	c.Set("current_operator", nonAdmin)
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	err = AssignAgentToParent(c)
	assert.NoError(t, err) // Handler returns nil error after writing JSON response
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAssignAgentToParent_SelfAssignment(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	agent := createAgent(t, db, "agent_self")
	admin := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	reqBody := AssignAgentToParentRequest{
		AgentID:        agent.ID,
		ParentID:       &agent.ID,
		CommissionRate: 0.1,
	}
	jsonReq, err := json.Marshal(reqBody)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/agents/assign-parent", strings.NewReader(string(jsonReq))), rec)
	c.Set("db", db)
	c.Set("current_operator", admin)
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	err = AssignAgentToParent(c)
	assert.NoError(t, err) // Handler returns nil error after writing JSON response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// no hierarchy record should exist
	var count int64
	db.Model(&domain.AgentHierarchy{}).Where("agent_id = ?", agent.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestAssignAgentToParent_CycleDetection(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	// create three agents: a -> b -> c
	a := createAgent(t, db, "agentA")
	b := createAgent(t, db, "agentB")
	c := createAgent(t, db, "agentC")

	// set up existing hierarchy a->b and b->c
	db.Create(&domain.AgentHierarchy{AgentID: b.ID, ParentID: &a.ID, Level: 1, CommissionRate: 0.05, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	db.Create(&domain.AgentHierarchy{AgentID: c.ID, ParentID: &b.ID, Level: 2, CommissionRate: 0.05, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()})

	admin := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	// try to assign parent of 'a' to 'c' which would create a cycle
	reqBody := AssignAgentToParentRequest{AgentID: a.ID, ParentID: &c.ID, CommissionRate: 0.1}
	jsonReq, err := json.Marshal(reqBody)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	cctx := e.NewContext(httptest.NewRequest(http.MethodPost, "/api/v1/agents/assign-parent", strings.NewReader(string(jsonReq))), rec)
	cctx.Set("db", db)
	cctx.Set("current_operator", admin)
	cctx.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	err = AssignAgentToParent(cctx)
	assert.NoError(t, err) // Handler returns nil error after writing JSON response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// ensure hierarchy of 'a' remains unchanged (nil parent)
	var h domain.AgentHierarchy
	hErr := db.Where("agent_id = ?", a.ID).First(&h).Error
	assert.Error(t, hErr) // should not exist
}

func TestGetAgentHierarchyEndpoints(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	// create root agent and children
	root := createAgent(t, db, "root")
	child1 := createAgent(t, db, "child1")
	child2 := createAgent(t, db, "child2")

	// create hierarchy records - root has no parent (parent_id is null)
	var rootParentID *int64 = nil
	db.Create(&domain.AgentHierarchy{AgentID: root.ID, ParentID: rootParentID, Level: 0, CommissionRate: 0.10, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	db.Create(&domain.AgentHierarchy{AgentID: child1.ID, ParentID: &root.ID, Level: 1, CommissionRate: 0.05, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	db.Create(&domain.AgentHierarchy{AgentID: child2.ID, ParentID: &root.ID, Level: 1, CommissionRate: 0.06, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()})

	// query hierarchy for root
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/agents/%d/hierarchy", root.ID), nil)
	cctx := e.NewContext(req, rec)
	cctx.SetParamNames("id")
	cctx.SetParamValues(fmt.Sprintf("%d", root.ID))
	cctx.Set("db", db)
	cctx.SetPath("/agents/:id/hierarchy")
	cctx.SetParamNames("id")
	cctx.SetParamValues(fmt.Sprintf("%d", root.ID))

	err := GetAgentHierarchy(cctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Extract the hierarchy data
	dataBytes, _ := json.Marshal(response.Data)
	var h domain.AgentHierarchy
	err = json.Unmarshal(dataBytes, &h)
	require.NoError(t, err)
	assert.Equal(t, root.ID, h.AgentID)

	// query sub-agents list
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/agents/%d/sub-agents", root.ID), nil)
	cctx2 := e.NewContext(req2, rec2)
	cctx2.SetParamNames("id")
	cctx2.SetParamValues(fmt.Sprintf("%d", root.ID))
	cctx2.Set("db", db)
	cctx2.SetPath("/agents/:id/sub-agents")
	cctx2.SetParamNames("id")
	cctx2.SetParamValues(fmt.Sprintf("%d", root.ID))

	err = GetAgentSubAgents(cctx2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)
	var response2 Response
	err = json.Unmarshal(rec2.Body.Bytes(), &response2)
	require.NoError(t, err)

	// Extract the sub-agents data
	dataBytes2, _ := json.Marshal(response2.Data)
	var subs []map[string]interface{}
	err = json.Unmarshal(dataBytes2, &subs)
	require.NoError(t, err)
	assert.Len(t, subs, 2)

	// query hierarchy tree
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/agents/%d/hierarchy-tree", root.ID), nil)
	cctx3 := e.NewContext(req3, rec3)
	cctx3.SetParamNames("id")
	cctx3.SetParamValues(fmt.Sprintf("%d", root.ID))
	cctx3.Set("db", db)
	cctx3.SetPath("/agents/:id/hierarchy-tree")
	cctx3.SetParamNames("id")
	cctx3.SetParamValues(fmt.Sprintf("%d", root.ID))

	err = GetAgentHierarchyTree(cctx3)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec3.Code)
	var tree Response
	err = json.Unmarshal(rec3.Body.Bytes(), &tree)
	require.NoError(t, err)

	// The tree response is wrapped in Response.Data
	treeData, _ := json.Marshal(tree.Data)
	var treeMap map[string]interface{}
	err = json.Unmarshal(treeData, &treeMap)
	require.NoError(t, err)

	// Verify the tree has children
	if children, ok := treeMap["children"]; ok {
		assert.Len(t, children, 2)
	}
}

func TestUpdateAgentCommissionRate(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	agent := createAgent(t, db, "rateagent")
	hierarchy := domain.AgentHierarchy{AgentID: agent.ID, Level: 0, CommissionRate: 0.01, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(&hierarchy)

	admin := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	req := struct {
		CommissionRate float64 `json:"commission_rate"`
	}{CommissionRate: 0.20}
	jsonReq, err := json.Marshal(req)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	cctx := e.NewContext(httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/agents/%d/commission-rate", agent.ID), strings.NewReader(string(jsonReq))), rec)
	cctx.Set("db", db)
	cctx.Set("current_operator", admin)
	cctx.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	cctx.SetPath("/agents/:id/commission-rate")
	cctx.SetParamNames("id")
	cctx.SetParamValues(fmt.Sprintf("%d", agent.ID))

	err = UpdateAgentCommissionRate(cctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var updated domain.AgentHierarchy
	db.Where("agent_id = ?", agent.ID).First(&updated)
	assert.Equal(t, 0.20, updated.CommissionRate)
}

func TestUpdateAgentCommissionRate_Unauthorized(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	agent := createAgent(t, db, "rateagent2")
	hierarchy := domain.AgentHierarchy{AgentID: agent.ID, Level: 0, CommissionRate: 0.01, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(&hierarchy)

	nonAdmin := &domain.SysOpr{ID: 2, Level: "agent", Username: "agent2"}

	req := struct {
		CommissionRate float64 `json:"commission_rate"`
	}{CommissionRate: 0.50}
	jsonReq, err := json.Marshal(req)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	cctx := e.NewContext(httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/agents/%d/commission-rate", agent.ID), strings.NewReader(string(jsonReq))), rec)
	cctx.Set("db", db)
	cctx.Set("current_operator", nonAdmin)
	cctx.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	cctx.SetPath("/agents/:id/commission-rate")
	cctx.SetParamNames("id")
	cctx.SetParamValues(fmt.Sprintf("%d", agent.ID))

	err = UpdateAgentCommissionRate(cctx)
	assert.NoError(t, err) // Handler returns nil error after writing JSON response
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
