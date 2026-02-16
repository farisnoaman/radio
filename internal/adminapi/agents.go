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
	webserver.ApiGET("/agents/:id/stats", GetAgentStats)
}
