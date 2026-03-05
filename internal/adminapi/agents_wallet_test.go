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
)

func TestGetAgentWalletTransactions(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	agent := createAgent(t, db, "txagent")

	// Create wallet and some transactions
	wallet := domain.AgentWallet{AgentID: agent.ID, Balance: 1000.0, UpdatedAt: time.Now()}
	db.Create(&wallet)

	transactions := []domain.WalletLog{
		{AgentID: agent.ID, Type: "deposit", Amount: 500.0, Balance: 500.0, ReferenceID: "ref1", Remark: "Initial deposit", CreatedAt: time.Now().Add(-time.Hour)},
		{AgentID: agent.ID, Type: "purchase", Amount: -100.0, Balance: 400.0, ReferenceID: "ref2", Remark: "Purchase", CreatedAt: time.Now().Add(-time.Minute * 30)},
		{AgentID: agent.ID, Type: "commission", Amount: 50.0, Balance: 450.0, ReferenceID: "ref3", Remark: "Commission", CreatedAt: time.Now()},
	}
	for _, tx := range transactions {
		db.Create(&tx)
	}

	tests := []struct {
		name           string
		agentID        string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all transactions",
			agentID:        fmt.Sprintf("%d", agent.ID),
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Filter by type",
			agentID:        fmt.Sprintf("%d", agent.ID),
			queryParams:    "?type=deposit",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by date range",
			agentID:        fmt.Sprintf("%d", agent.ID),
			queryParams:    "?start_date=" + time.Now().Add(-time.Hour*2).Format(time.RFC3339) + "&end_date=" + time.Now().Add(time.Hour).Format(time.RFC3339),
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Invalid agent ID",
			agentID:        "invalid",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "Non-existent agent",
			agentID:        "99999",
			queryParams:    "",
			expectedStatus: http.StatusNotFound,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+tt.agentID+"/wallet/transactions"+tt.queryParams, nil)
			cctx := e.NewContext(req, rec)
			cctx.SetParamNames("id")
			cctx.SetParamValues(tt.agentID)
			cctx.Set("db", db)

			err := GetAgentWalletTransactions(cctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var transactions []domain.WalletLog
				err = json.Unmarshal(dataBytes, &transactions)
				require.NoError(t, err)
				assert.Len(t, transactions, tt.expectedCount)
			}
		})
	}
}

func TestBulkWalletOperation(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	// Create test agents
	agent1 := createAgent(t, db, "bulk1")
	agent2 := createAgent(t, db, "bulk2")
	agent3 := createAgent(t, db, "bulk3")

	// Create wallets with initial balances
	wallets := []domain.AgentWallet{
		{AgentID: agent1.ID, Balance: 100.0, UpdatedAt: time.Now()},
		{AgentID: agent2.ID, Balance: 200.0, UpdatedAt: time.Now()},
		{AgentID: agent3.ID, Balance: 300.0, UpdatedAt: time.Now()},
	}
	for _, wallet := range wallets {
		db.Create(&wallet)
	}

	admin := &domain.SysOpr{ID: 1, Level: "admin", Username: "admin"}

	tests := []struct {
		name           string
		request        BulkWalletRequest
		currentUser    *domain.SysOpr
		expectedStatus int
		expectedSuccess int
		expectedFail    int
	}{
		{
			name: "Bulk deposit success",
			request: BulkWalletRequest{
				AgentIDs:  []int64{agent1.ID, agent2.ID},
				Operation: "deposit",
				Amount:    50.0,
				Remark:    "Bulk deposit test",
			},
			currentUser:    admin,
			expectedStatus: http.StatusOK,
			expectedSuccess: 2,
			expectedFail:    0,
		},
		{
			name: "Bulk set balance",
			request: BulkWalletRequest{
				AgentIDs:  []int64{agent3.ID},
				Operation: "set",
				Amount:    500.0,
				Remark:    "Set balance test",
			},
			currentUser:    admin,
			expectedStatus: http.StatusOK,
			expectedSuccess: 1,
			expectedFail:    0,
		},
		{
			name: "Unauthorized user",
			request: BulkWalletRequest{
				AgentIDs:  []int64{agent1.ID},
				Operation: "deposit",
				Amount:    10.0,
			},
			currentUser:    &domain.SysOpr{ID: 2, Level: "agent", Username: "agent"},
			expectedStatus: http.StatusForbidden,
			expectedSuccess: 0,
			expectedFail:    0,
		},
		{
			name: "Invalid operation",
			request: BulkWalletRequest{
				AgentIDs:  []int64{agent1.ID},
				Operation: "invalid",
				Amount:    10.0,
			},
			currentUser:    admin,
			expectedStatus: http.StatusBadRequest,
			expectedSuccess: 0,
			expectedFail:    0,
		},
		{
			name: "Too many agents",
			request: BulkWalletRequest{
				AgentIDs:  make([]int64, 51), // 51 agents
				Operation: "deposit",
				Amount:    10.0,
			},
			currentUser:    admin,
			expectedStatus: http.StatusBadRequest,
			expectedSuccess: 0,
			expectedFail:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonReq, err := json.Marshal(tt.request)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/wallet/bulk", strings.NewReader(string(jsonReq)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			cctx := e.NewContext(req, rec)
			cctx.Set("db", db)
			cctx.Set("current_operator", tt.currentUser)

			err = BulkWalletOperation(cctx)
			
			// For validation errors (invalid operation, too many agents), 
			// the handler returns an error directly
			if tt.name == "Invalid operation" || tt.name == "Too many agents" {
				// Validation errors return an error, not a response
				assert.Error(t, err)
				return
			}
			
			// For other cases, handler returns nil and we check response code
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var bulkResp BulkWalletResponse
				err = json.Unmarshal(dataBytes, &bulkResp)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedSuccess, bulkResp.SuccessfulOps)
				assert.Equal(t, tt.expectedFail, bulkResp.FailedOps)
				assert.Len(t, bulkResp.Results, len(tt.request.AgentIDs))
			}
		})
	}
}

func TestGetWalletBalanceAlerts(t *testing.T) {
	e := setupTestEcho()
	db := setupTestDBForAgents(t)

	// Create test agents with different balances
	agent1 := createAgent(t, db, "lowbalance1")
	agent2 := createAgent(t, db, "lowbalance2")
	agent3 := createAgent(t, db, "highbalance")

	// Create wallets
	wallets := []domain.AgentWallet{
		{AgentID: agent1.ID, Balance: 50.0, UpdatedAt: time.Now()},  // Below default threshold (100)
		{AgentID: agent2.ID, Balance: 75.0, UpdatedAt: time.Now()},  // Below default threshold
		{AgentID: agent3.ID, Balance: 200.0, UpdatedAt: time.Now()}, // Above threshold
	}
	for _, wallet := range wallets {
		db.Create(&wallet)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Default threshold (100)",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // agent1 (50) and agent2 (75) are below 100
		},
		{
			name:           "Custom threshold (80)",
			queryParams:    "?threshold=80",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // agent1 (50) and agent2 (75) are below 80
		},
		{
			name:           "High threshold (300)",
			queryParams:    "?threshold=300",
			expectedStatus: http.StatusOK,
			expectedCount:  3, // all agents
		},
		{
			name:           "Pagination",
			queryParams:    "?page=1&perPage=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/wallet/alerts"+tt.queryParams, nil)
			cctx := e.NewContext(req, rec)
			cctx.Set("db", db)

			err := GetWalletBalanceAlerts(cctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var alerts []map[string]interface{}
				err = json.Unmarshal(dataBytes, &alerts)
				require.NoError(t, err)
				assert.Len(t, alerts, tt.expectedCount)
			}
		})
	}
}