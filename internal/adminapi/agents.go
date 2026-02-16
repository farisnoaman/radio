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
	var agents []domain.SysOpr

	// Filter only agents
	query := db.Model(&domain.SysOpr{}).Where("level = ?", "agent")

	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		query = query.Where("username LIKE ? OR realname LIKE ?", "%"+name+"%", "%"+name+"%")
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order("id DESC").Limit(perPage).Offset(offset).Find(&agents)

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
	if err := tx.Model(&wallet).Update("balance", newBalance).Error; err != nil {
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

func registerAgentRoutes() {
	webserver.ApiGET("/agents", ListAgents)
	webserver.ApiPOST("/agents/:id/topup", TopupAgent)
	webserver.ApiGET("/agents/:id/wallet", GetAgentWallet)
}
