package adminapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// ListAgents retrieves the agent list (SysOpr with level=agent)
// @Summary get the agent list
// @Tags Agent
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Success 200 {object} ListResponse
// @Router /api/v1/agents [get]
func ListAgents(c echo.Context) error {
	db := GetDB(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	var total int64
	var agents []struct {
		domain.SysOpr
		Balance float64 `json:"balance"`
	}

	// Filter only agents and join with wallet
	query := db.Table("sys_opr").
		Select("sys_opr.*, COALESCE(agent_wallet.balance, 0) as balance").
		Joins("LEFT JOIN agent_wallet ON sys_opr.id = agent_wallet.agent_id").
		Where("sys_opr.level = ?", "agent")

	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		query = query.Where("sys_opr.username LIKE ? OR sys_opr.realname LIKE ?", "%"+name+"%", "%"+name+"%")
	}

	query.Count(&total)

	err := query.Offset((page - 1) * perPage).Limit(perPage).Order("sys_opr.id DESC").Find(&agents).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query agents", err.Error())
	}

	return paged(c, agents, total, page, perPage)
}

// AgentTopupRequest
type AgentTopupRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Remark string  `json:"remark"`
}

// BulkWalletRequest for bulk wallet operations
type BulkWalletRequest struct {
	AgentIDs  []int64 `json:"agent_ids" validate:"required,min=1,max=50"`
	Operation string  `json:"operation" validate:"required,oneof=deposit purchase refund set"`
	Amount    float64 `json:"amount" validate:"required"`
	Remark    string  `json:"remark"`
}

// BulkWalletResult for individual operation results
type BulkWalletResult struct {
	AgentID         int64   `json:"agent_id"`
	Success         bool    `json:"success"`
	PreviousBalance float64 `json:"previous_balance,omitempty"`
	NewBalance      float64 `json:"new_balance,omitempty"`
	Error           string  `json:"error,omitempty"`
}

// BulkWalletResponse for bulk operation summary
type BulkWalletResponse struct {
	TotalAgents    int                `json:"total_agents"`
	SuccessfulOps  int                `json:"successful_operations"`
	FailedOps      int                `json:"failed_operations"`
	Results        []BulkWalletResult `json:"results"`
}

// TopupAgent adds balance to agent wallet
// @Summary topup agent wallet
// @Tags Agent
// @Param id path int true "Agent ID"
// @Param topup body AgentTopupRequest true "Topup info"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/agents/{id}/topup [post]
func TopupAgent(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	var req AgentTopupRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Verify Agent Exists
	var agent domain.SysOpr
	if err := GetDB(c).First(&agent, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Agent not found", nil)
	}
	
	if agent.Level != "agent" {
		return fail(c, http.StatusBadRequest, "INVALID_ROLE", "User is not an agent", nil)
	}

	tx := GetDB(c).Begin()

	// 1. Get or Create Wallet
	var wallet domain.AgentWallet
	if err := tx.FirstOrCreate(&wallet, domain.AgentWallet{AgentID: id}).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to access wallet", err.Error())
	}

	// 2. Update Balance
	newBalance := wallet.Balance + req.Amount
	if err := tx.Model(&domain.AgentWallet{}).Where("agent_id = ?", id).Updates(map[string]interface{}{"balance": newBalance, "updated_at": time.Now()}).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update balance", err.Error())
	}

	// 3. Log Transaction
	log := domain.WalletLog{
		AgentID:     id,
		Type:        "deposit",
		Amount:      req.Amount,
		Balance:     newBalance,
		ReferenceID: "manual-" + common.UUID(),
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "LOG_FAILED", "Failed to create transaction log", err.Error())
	}

	tx.Commit()

	return ok(c, map[string]interface{}{
		"current_balance": newBalance,
		"message":         "Topup successful",
	})
}

// GetAgentWallet retrieves agent wallet info
// @Summary get agent wallet
// @Tags Agent
// @Param id path int true "Agent ID"
// @Success 200 {object} domain.AgentWallet
// @Router /api/v1/agents/{id}/wallet [get]
func GetAgentWallet(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	var wallet domain.AgentWallet
	if err := GetDB(c).FirstOrCreate(&wallet, domain.AgentWallet{AgentID: id}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to access wallet", err.Error())
	}

	return ok(c, wallet)
}

// GetAgentWalletTransactions retrieves paginated wallet transaction history with filtering
// @Summary get agent wallet transaction history
// @Tags Agent
// @Param id path int true "Agent ID"
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page (max 100)"
// @Param type query string false "Transaction type filter (deposit, purchase, refund, commission)"
// @Param start_date query string false "Start date (ISO8601 format)"
// @Param end_date query string false "End date (ISO8601 format)"
// @Param min_amount query number false "Minimum transaction amount"
// @Param max_amount query number false "Maximum transaction amount"
// @Success 200 {object} ListResponse
// @Router /api/v1/agents/{id}/wallet/transactions [get]
func GetAgentWalletTransactions(c echo.Context) error {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	// Verify agent exists and is an agent
	var agent domain.SysOpr
	if err := GetDB(c).Where("id = ? AND level = ?", agentID, "agent").First(&agent).Error; err != nil {
		return fail(c, http.StatusNotFound, "AGENT_NOT_FOUND", "Agent not found", nil)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := GetDB(c).Model(&domain.WalletLog{}).Where("agent_id = ?", agentID)

	// Apply filters
	if txType := strings.TrimSpace(c.QueryParam("type")); txType != "" {
		query = query.Where("type = ?", txType)
	}

	if startDate := strings.TrimSpace(c.QueryParam("start_date")); startDate != "" {
		if startTime, err := time.Parse(time.RFC3339, startDate); err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}

	if endDate := strings.TrimSpace(c.QueryParam("end_date")); endDate != "" {
		if endTime, err := time.Parse(time.RFC3339, endDate); err == nil {
			query = query.Where("created_at <= ?", endTime)
		}
	}

	if minAmountStr := strings.TrimSpace(c.QueryParam("min_amount")); minAmountStr != "" {
		if minAmount, err := strconv.ParseFloat(minAmountStr, 64); err == nil {
			query = query.Where("ABS(amount) >= ?", minAmount)
		}
	}

	if maxAmountStr := strings.TrimSpace(c.QueryParam("max_amount")); maxAmountStr != "" {
		if maxAmount, err := strconv.ParseFloat(maxAmountStr, 64); err == nil {
			query = query.Where("ABS(amount) <= ?", maxAmount)
		}
	}

	var total int64
	query.Count(&total)

	var transactions []domain.WalletLog
	offset := (page - 1) * perPage
	err = query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&transactions).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve transactions", err.Error())
	}

	return paged(c, transactions, total, page, perPage)
}

// BulkWalletOperation performs bulk wallet operations for multiple agents
// @Summary perform bulk wallet operations
// @Tags Agent
// @Param operation body BulkWalletRequest true "Bulk operation details"
// @Success 200 {object} BulkWalletResponse
// @Router /api/v1/agents/wallet/bulk [post]
func BulkWalletOperation(c echo.Context) error {
	// Permission check - only admins can perform bulk operations
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if currentOpr.Level != "super" && currentOpr.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only admins can perform bulk wallet operations", nil)
	}

	var req BulkWalletRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if len(req.AgentIDs) == 0 {
		return fail(c, http.StatusBadRequest, "NO_AGENTS", "At least one agent ID must be specified", nil)
	}

	if len(req.AgentIDs) > 50 {
		return fail(c, http.StatusBadRequest, "TOO_MANY_AGENTS", "Cannot process more than 50 agents at once", nil)
	}

	db := GetDB(c)
	tx := db.Begin()

	results := make([]BulkWalletResult, 0, len(req.AgentIDs))

	for _, agentID := range req.AgentIDs {
		result := BulkWalletResult{AgentID: agentID}

		// Verify agent exists and is an agent
		var agent domain.SysOpr
		if err := tx.Where("id = ? AND level = ?", agentID, "agent").First(&agent).Error; err != nil {
			result.Success = false
			result.Error = "Agent not found"
			results = append(results, result)
			continue
		}

		// Get or create wallet
		var wallet domain.AgentWallet
		if err := tx.FirstOrCreate(&wallet, domain.AgentWallet{AgentID: agentID}).Error; err != nil {
			result.Success = false
			result.Error = "Failed to access wallet"
			results = append(results, result)
			continue
		}

		// Calculate new balance
		newBalance := wallet.Balance + req.Amount
		if req.Operation == "set" {
			newBalance = req.Amount
		}

		// Update wallet balance
		if err := tx.Model(&domain.AgentWallet{}).Where("agent_id = ?", agentID).
			Updates(map[string]interface{}{"balance": newBalance, "updated_at": time.Now()}).Error; err != nil {
			result.Success = false
			result.Error = "Failed to update balance"
			results = append(results, result)
			continue
		}

		// Log transaction
		log := domain.WalletLog{
			AgentID:     agentID,
			Type:        req.Operation,
			Amount:      req.Amount,
			Balance:     newBalance,
			ReferenceID: "bulk-" + common.UUID(),
			Remark:      req.Remark,
			CreatedAt:   time.Now(),
		}

		if err := tx.Create(&log).Error; err != nil {
			result.Success = false
			result.Error = "Failed to create transaction log"
			results = append(results, result)
			continue
		}

		result.Success = true
		result.PreviousBalance = wallet.Balance
		result.NewBalance = newBalance
		results = append(results, result)
	}

	if err := tx.Commit().Error; err != nil {
		return fail(c, http.StatusInternalServerError, "COMMIT_FAILED", "Failed to commit bulk operation", err.Error())
	}

	response := BulkWalletResponse{
		TotalAgents:    len(req.AgentIDs),
		SuccessfulOps:  0,
		FailedOps:      0,
		Results:        results,
	}

	for _, result := range results {
		if result.Success {
			response.SuccessfulOps++
		} else {
			response.FailedOps++
		}
	}

	return ok(c, response)
}

// GetWalletBalanceAlerts retrieves agents with low balance alerts
// @Summary get wallet balance alerts
// @Tags Agent
// @Param threshold query number false "Balance threshold for alerts (default: 100)"
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page (max 100)"
// @Success 200 {object} ListResponse
// @Router /api/v1/agents/wallet/alerts [get]
func GetWalletBalanceAlerts(c echo.Context) error {
	thresholdStr := strings.TrimSpace(c.QueryParam("threshold"))
	threshold := 100.0 // default threshold
	if thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 64); err == nil && t >= 0 {
			threshold = t
		}
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int64
	var alerts []struct {
		domain.SysOpr
		Balance float64 `json:"balance"`
	}

	query := GetDB(c).Table("sys_opr").
		Select("sys_opr.*, COALESCE(agent_wallet.balance, 0) as balance").
		Joins("LEFT JOIN agent_wallet ON sys_opr.id = agent_wallet.agent_id").
		Where("sys_opr.level = ? AND COALESCE(agent_wallet.balance, 0) < ?", "agent", threshold)

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("COALESCE(agent_wallet.balance, 0) ASC").Limit(perPage).Offset(offset).Find(&alerts).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve balance alerts", err.Error())
	}

	return paged(c, alerts, total, page, perPage)
}

// CreateAgent creates a new agent
// @Summary create an agent
// @Tags Agent
// @Param agent body operatorPayload true "Agent info"
// @Success 201 {object} domain.SysOpr
// @Router /api/v1/agents [post]
func CreateAgent(c echo.Context) error {
	// Permission check
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if currentOpr.Level != "super" && currentOpr.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only admins can create agents", nil)
	}

	var payload operatorPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse agent parameters", nil)
	}

	payload.Username = strings.TrimSpace(payload.Username)
	payload.Password = strings.TrimSpace(payload.Password)

	// Validate required fields
	if payload.Username == "" {
		return fail(c, http.StatusBadRequest, "MISSING_USERNAME", "Username is required", nil)
	}
	if payload.Password == "" {
		return fail(c, http.StatusBadRequest, "MISSING_PASSWORD", "Password is required", nil)
	}
	if payload.Realname == "" {
		return fail(c, http.StatusBadRequest, "MISSING_REALNAME", "Real name is required", nil)
	}

	// Validate Username format
	if len(payload.Username) < 3 || len(payload.Username) > 30 {
		return fail(c, http.StatusBadRequest, "INVALID_USERNAME", "Username length must be between 3 and 30 characters", nil)
	}

	// Validate Password length
	if len(payload.Password) < 6 || len(payload.Password) > 50 {
		return fail(c, http.StatusBadRequest, "INVALID_PASSWORD", "Password length must be between 6 and 50 characters", nil)
	}

	// Check Username exists
	var exists int64
	GetDB(c).Model(&domain.SysOpr{}).Where("username = ?", payload.Username).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "USERNAME_EXISTS", "Username already exists", nil)
	}

	hashedPassword := common.Sha256HashWithSalt(payload.Password, common.GetSecretSalt())

	operator := domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  payload.Username,
		Password:  hashedPassword,
		Realname:  payload.Realname,
		Mobile:    payload.Mobile,
		Email:     payload.Email,
		Level:     "agent", // Force level to agent
		Status:    "enabled",
		Remark:    payload.Remark,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := GetDB(c).Create(&operator).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create agent", err.Error())
	}

	// Initialize wallet
	if err := GetDB(c).Create(&domain.AgentWallet{AgentID: operator.ID, Balance: 0, UpdatedAt: time.Now()}).Error; err != nil {
		// Log error but don't fail, wallet can be created on topup
		c.Logger().Warn("Failed to initialize agent wallet", err)
	}

	operator.Password = ""
	return ok(c, operator)
}

func registerAgentRoutes() {
	webserver.ApiGET("/agents", ListAgents)
	webserver.ApiPOST("/agents", CreateAgent)
	webserver.ApiPOST("/agents/:id/topup", TopupAgent)
	webserver.ApiGET("/agents/:id/wallet", GetAgentWallet)
	webserver.ApiGET("/agents/:id/wallet/transactions", GetAgentWalletTransactions)
	webserver.ApiPOST("/agents/wallet/bulk", BulkWalletOperation)
	webserver.ApiGET("/agents/wallet/alerts", GetWalletBalanceAlerts)
	webserver.ApiGET("/agents/:id/stats", GetAgentStats)

	// Agent Self-Service Portal Routes
	webserver.ApiGET("/agent/dashboard", GetAgentDashboard)
	webserver.ApiGET("/agent/wallet", GetMyWallet)
	webserver.ApiGET("/agent/wallet/transactions", GetMyWalletTransactions)
	webserver.ApiGET("/agent/commissions", GetMyCommissions)
	webserver.ApiPOST("/agent/payout-request", CreatePayoutRequest)
	webserver.ApiGET("/agent/payout-requests", ListMyPayoutRequests)
	webserver.ApiDELETE("/agent/payout-requests/:id", CancelMyPayoutRequest)
	webserver.ApiGET("/agent/sub-agents", GetMySubAgents)
}

// ---------------------------------------------------------------------------
// Agent Self-Service Portal Handlers
// ----------------------------------------------------------------------------

// GetAgentDashboard returns the agent's dashboard data including wallet balance,
// commission summary, and sub-agent count.
// @Summary get agent dashboard
// @Tags Agent Portal
// @Success 200 {object} SuccessResponse
// @Router /api/v1/agent/dashboard [get]
func GetAgentDashboard(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	db := GetDB(c)

	// Get wallet balance
	var wallet domain.AgentWallet
	err = db.FirstOrCreate(&wallet, domain.AgentWallet{AgentID: agent.ID}).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get wallet", err.Error())
	}

	// Get commission summary
	var totalEarned, totalPaid, pendingAmount float64
	db.Model(&domain.CommissionLog{}).Where("agent_id = ? AND status IN ?", agent.ID, []string{"pending", "payable", "paid"}).Select("COALESCE(SUM(amount), 0)").Scan(&totalEarned)
	db.Model(&domain.CommissionLog{}).Where("agent_id = ? AND status = ?", agent.ID, "paid").Select("COALESCE(SUM(amount), 0)").Scan(&totalPaid)
	pendingAmount = totalEarned - totalPaid

	// Get sub-agent count
	var subAgentCount int64
	db.Model(&domain.AgentHierarchy{}).Where("parent_id = ? AND status = ?", agent.ID, "active").Count(&subAgentCount)

	return ok(c, map[string]interface{}{
		"agent": map[string]interface{}{
			"id":         agent.ID,
			"username":   agent.Username,
			"realname":   agent.Realname,
			"level":      agent.Level,
			"status":     agent.Status,
		},
		"wallet": map[string]interface{}{
			"balance":     wallet.Balance,
			"updated_at":  wallet.UpdatedAt,
		},
		"commissions": map[string]interface{}{
			"total_earned":   totalEarned,
			"total_paid":     totalPaid,
			"pending_amount": pendingAmount,
		},
		"sub_agents": map[string]interface{}{
			"count": subAgentCount,
		},
	})
}

// GetMyWallet returns the current agent's wallet information.
// @Summary get my wallet
// @Tags Agent Portal
// @Success 200 {object} SuccessResponse
// @Router /api/v1/agent/wallet [get]
func GetMyWallet(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	var wallet domain.AgentWallet
	err = GetDB(c).FirstOrCreate(&wallet, domain.AgentWallet{AgentID: agent.ID}).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get wallet", err.Error())
	}

	return ok(c, wallet)
}

// GetMyWalletTransactions returns the current agent's wallet transaction history.
// @Summary get my wallet transactions
// @Tags Agent Portal
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Success 200 {object} ListResponse
// @Router /api/v1/agent/wallet/transactions [get]
func GetMyWalletTransactions(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := GetDB(c).Model(&domain.WalletLog{}).Where("agent_id = ?", agent.ID)

	var total int64
	query.Count(&total)

	var transactions []domain.WalletLog
	offset := (page - 1) * perPage
	err = query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&transactions).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve transactions", err.Error())
	}

	return paged(c, transactions, total, page, perPage)
}

// GetMyCommissions returns the current agent's commission history.
// @Summary get my commissions
// @Tags Agent Portal
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param status query string false "Filter by status (pending, payable, paid)"
// @Success 200 {object} ListResponse
// @Router /api/v1/agent/commissions [get]
func GetMyCommissions(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := GetDB(c).Model(&domain.CommissionLog{}).Where("agent_id = ?", agent.ID)

	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var commissions []domain.CommissionLog
	offset := (page - 1) * perPage
	err = query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&commissions).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve commissions", err.Error())
	}

	return paged(c, commissions, total, page, perPage)
}

// PayoutRequestInput defines the payout request structure
type PayoutRequestInput struct {
	Amount         float64 `json:"amount" validate:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=bank_transfer paypal crypto cash check"`
	PaymentDetails string  `json:"payment_details" validate:"required"`
	Notes         string  `json:"notes"`
}

// CreatePayoutRequest allows an agent to request a payout.
// @Summary create payout request
// @Tags Agent Portal
// @Param request body PayoutRequestInput true "Payout request details"
// @Success 201 {object} SuccessResponse
// @Router /api/v1/agent/payout-request [post]
func CreatePayoutRequest(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	var req PayoutRequestInput
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	db := GetDB(c)

	// Check wallet balance
	var wallet domain.AgentWallet
	if err := db.First(&wallet, "agent_id = ?", agent.ID).Error; err != nil {
		return fail(c, http.StatusNotFound, "WALLET_NOT_FOUND", "Wallet not found", nil)
	}

	if wallet.Balance < req.Amount {
		return fail(c, http.StatusBadRequest, "INSUFFICIENT_BALANCE", "Insufficient wallet balance", nil)
	}

	// Create payout request
	payoutReq := domain.PayoutRequest{
		AgentID:        agent.ID,
		Amount:         req.Amount,
		PaymentMethod:  req.PaymentMethod,
		PaymentDetails: req.PaymentDetails,
		Notes:          req.Notes,
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.Create(&payoutReq).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to create payout request", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message":        "Payout request created successfully",
		"payout_request": payoutReq,
	})
}

// ListMyPayoutRequests returns the current agent's payout requests.
// @Summary list my payout requests
// @Tags Agent Portal
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Success 200 {object} ListResponse
// @Router /api/v1/agent/payout-requests [get]
func ListMyPayoutRequests(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := GetDB(c).Model(&domain.PayoutRequest{}).Where("agent_id = ?", agent.ID)

	var total int64
	query.Count(&total)

	var requests []domain.PayoutRequest
	offset := (page - 1) * perPage
	err = query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&requests).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve payout requests", err.Error())
	}

	return paged(c, requests, total, page, perPage)
}

// CancelMyPayoutRequest allows an agent to cancel their own pending payout request.
// This is the self-service portal version that uses resolveOperatorFromContext.
// @Summary cancel my payout request
// @Tags Agent Portal
// @Param id path int true "Payout request ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/agent/payout-requests/{id} [delete]
func CancelMyPayoutRequest(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid payout request ID", nil)
	}

	db := GetDB(c)

	// Find the payout request
	var payoutReq domain.PayoutRequest
	if err := db.First(&payoutReq, "id = ? AND agent_id = ?", id, agent.ID).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Payout request not found", nil)
	}

	// Only pending requests can be cancelled
	if payoutReq.Status != "pending" {
		return fail(c, http.StatusBadRequest, "INVALID_STATUS", "Only pending payout requests can be cancelled", nil)
	}

	// Update status to cancelled
	payoutReq.Status = "cancelled"
	payoutReq.UpdatedAt = time.Now()

	if err := db.Save(&payoutReq).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to cancel payout request", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "Payout request cancelled successfully",
	})
}

// GetMySubAgents returns the current agent's sub-agents (downline).
// @Summary get my sub-agents
// @Tags Agent Portal
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Success 200 {object} ListResponse
// @Router /api/v1/agent/sub-agents [get]
func GetMySubAgents(c echo.Context) error {
	agent, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	if agent.Level != "agent" {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Only agents can access this resource", nil)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	db := GetDB(c)

	// Get sub-agents from hierarchy
	var subAgents []struct {
		domain.AgentHierarchy
		domain.SysOpr
		Balance float64 `json:"balance"`
	}

	query := db.Table("agent_hierarchy").
		Select("agent_hierarchy.*, sys_opr.id, sys_opr.username, sys_opr.realname, sys_opr.status, COALESCE(agent_wallet.balance, 0) as balance").
		Joins("LEFT JOIN sys_opr ON agent_hierarchy.agent_id = sys_opr.id").
		Joins("LEFT JOIN agent_wallet ON agent_hierarchy.agent_id = agent_wallet.agent_id").
		Where("agent_hierarchy.parent_id = ? AND agent_hierarchy.status = ?", agent.ID, "active")

	var total int64
	query.Count(&total)

	offset := (page - 1) * perPage
	err = query.Offset(offset).Limit(perPage).Order("agent_hierarchy.created_at DESC").Find(&subAgents).Error
	if err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve sub-agents", err.Error())
	}

	return paged(c, subAgents, total, page, perPage)
}
