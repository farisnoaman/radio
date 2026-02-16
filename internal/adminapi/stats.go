package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

// BatchStats per-batch voucher stats
type BatchStats struct {
	ID             int64     `json:"id,string"`
	Name           string    `json:"name"`
	TotalVouchers  int64     `json:"total_vouchers"`
	UsedVouchers   int64     `json:"used_vouchers"`
	UnusedVouchers int64     `json:"unused_vouchers"`
	CreatedAt      time.Time `json:"created_at"`
}

// AgentStats response structure
type AgentStats struct {
	Balance            float64            `json:"balance"`
	TotalBatches       int64              `json:"total_batches"`
	TotalVouchers      int64              `json:"total_vouchers"`
	UsedVouchers       int64              `json:"used_vouchers"`
	UnusedVouchers     int64              `json:"unused_vouchers"`
	RecentTransactions []domain.WalletLog `json:"recent_transactions"`
	RecentBatches      []BatchStats       `json:"recent_batches"`
}

// GetAgentStats retrieves statistics for an agent
// @Summary get agent stats
// @Tags Agent
// @Param id path int true "Agent ID"
// @Success 200 {object} AgentStats
// @Router /api/v1/agents/{id}/stats [get]
func GetAgentStats(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	// Permission check: only self or admin
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	if currentUser.Level == "agent" && currentUser.ID != id {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Agents can only view their own stats", nil)
	}

	db := GetDB(c)
	var stats AgentStats

	// 1. Get Balance
	var wallet domain.AgentWallet
	db.Where("agent_id = ?", id).First(&wallet)
	stats.Balance = wallet.Balance

	// 2. Total Batches
	db.Model(&domain.VoucherBatch{}).Where("agent_id = ?", id).Count(&stats.TotalBatches)

	// 3. Voucher Counts
	db.Model(&domain.Voucher{}).Where("agent_id = ?", id).Count(&stats.TotalVouchers)
	db.Model(&domain.Voucher{}).Where("agent_id = ? AND status = ?", id, "used").Count(&stats.UsedVouchers)
	db.Model(&domain.Voucher{}).Where("agent_id = ? AND status = ?", id, "unused").Count(&stats.UnusedVouchers)

	// 4. Recent Transactions
	db.Where("agent_id = ?", id).Order("created_at DESC").Limit(10).Find(&stats.RecentTransactions)

	// 5. Recent Batches with per-batch stats
	var batches []domain.VoucherBatch
	db.Where("agent_id = ?", id).Order("id DESC").Limit(10).Find(&batches)

	for _, b := range batches {
		var bStats BatchStats
		bStats.ID = b.ID
		bStats.Name = b.Name
		bStats.CreatedAt = b.CreatedAt
		db.Model(&domain.Voucher{}).Where("batch_id = ?", b.ID).Count(&bStats.TotalVouchers)
		db.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", b.ID, "used").Count(&bStats.UsedVouchers)
		db.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", b.ID, "unused").Count(&bStats.UnusedVouchers)
		stats.RecentBatches = append(stats.RecentBatches, bStats)
	}

	return ok(c, stats)
}
