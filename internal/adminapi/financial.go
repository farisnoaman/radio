package adminapi

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

type NetworkOverview struct {
	TotalBatches int64   `json:"total_batches"`
	TotalAgents  int64   `json:"total_agents"`
	TotalBalance float64 `json:"total_balance"`
}

type AgentPerformance struct {
	ID             int64   `json:"id,string"`
	Name           string  `json:"name"`
	Username       string  `json:"username"`
	Balance        float64 `json:"balance"`
	TotalVouchers  int64   `json:"total_vouchers"`
	UsedVouchers   int64   `json:"used_vouchers"`
	UnusedVouchers int64   `json:"unused_vouchers"`
}

type AgentPerformanceSummary struct {
	TotalAgents    int64   `json:"total_agents"`
	TotalBatches   int64   `json:"total_batches"`
	TotalVouchers  int64   `json:"total_vouchers"`
	TotalCost      float64 `json:"total_cost"`
	UsedCost       float64 `json:"used_cost"`
	UnusedCost     float64 `json:"unused_cost"`
}

type AdminBatchDetail struct {
	ID             int64     `json:"id,string"`
	Name           string    `json:"name"`
	ProductName    string    `json:"product_name"`
	Count          int       `json:"count"`
	UsedVouchers   int64     `json:"used_vouchers"`
	UnusedVouchers int64     `json:"unused_vouchers"`
	TotalCost      float64   `json:"total_cost"`
	CreatedAt      time.Time `json:"created_at"`
}

type AdminPerformance struct {
	TotalBatches   int64              `json:"total_batches"`
	TotalVouchers  int64              `json:"total_vouchers"`
	UsedVouchers   int64              `json:"used_vouchers"`
	UnusedVouchers int64              `json:"unused_vouchers"`
	TotalCost      float64            `json:"total_cost"`
	UsedCost       float64            `json:"used_cost"`
	UnusedCost     float64            `json:"unused_cost"`
	Batches        []AdminBatchDetail `json:"batches"`
}

type FinancialReport struct {
	Overview     NetworkOverview         `json:"overview"`
	AgentSummary AgentPerformanceSummary `json:"agent_summary"`
	Agents       []AgentPerformance      `json:"agents"`
	Admin        AdminPerformance        `json:"admin"`
	DateRange    struct {
		Start *string `json:"start"`
		End   *string `json:"end"`
	} `json:"date_range"`
}

// GetFinancialReport retrieves the financial performance report
// @Summary get financial report
// @Tags Financial
// @Success 200 {object} FinancialReport
// @Router /api/v1/financial/report [get]
func GetFinancialReport(c echo.Context) error {
	// Permission check: only admin/super
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}
	if currentUser.Level == "agent" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Agents cannot view financial reports", nil)
	}

	db := GetDB(c)
	var report FinancialReport
	report.Agents = make([]AgentPerformance, 0)
	
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	if startDateStr != "" {
		report.DateRange.Start = &startDateStr
	}
	if endDateStr != "" {
		report.DateRange.End = &endDateStr
	}

	// 1. Network Overview
	db.Model(&domain.VoucherBatch{}).Count(&report.Overview.TotalBatches)
	db.Model(&domain.SysOpr{}).Where("level = ?", "agent").Count(&report.Overview.TotalAgents)
	
	// Total Agent Balance (Current, not historical)
	var totalBalance float64
	db.Model(&domain.AgentWallet{}).Select("COALESCE(sum(balance), 0)").Scan(&totalBalance)
	report.Overview.TotalBalance = totalBalance

	// 2. Agent Performance
	var agents []domain.SysOpr
	db.Where("level = ?", "agent").Find(&agents)

	for _, agent := range agents {
		var ap AgentPerformance
		ap.ID = agent.ID
		ap.Name = agent.Realname
		ap.Username = agent.Username
		
		var wallet domain.AgentWallet
		db.Where("agent_id = ?", agent.ID).First(&wallet)
		ap.Balance = wallet.Balance

		// Voucher stats for this agent
		vQuery := db.Model(&domain.Voucher{}).Where("agent_id = ?", agent.ID)
		
		// Apply date filter to voucher creation/usage
		if startDateStr != "" {
			vQuery = vQuery.Where("created_at >= ?", startDateStr)
		}
		if endDateStr != "" {
			vQuery = vQuery.Where("created_at <= ?", endDateStr)
		}
		
		vQuery.Count(&ap.TotalVouchers)
		
		// We need independent queries or separate counts because Go GORM chaining modifies the query struct
		// Clone query for used/unused

		// Note: GORM's Count() modifies the query, so we re-build or use Session
		// Re-building is safer/simpler here
		
		// Used
		qUsed := db.Model(&domain.Voucher{}).Where("agent_id = ? AND status = ?", agent.ID, "used")
		if startDateStr != "" { qUsed = qUsed.Where("created_at >= ?", startDateStr) }
		if endDateStr != "" { qUsed = qUsed.Where("created_at <= ?", endDateStr) }
		qUsed.Count(&ap.UsedVouchers)

		// Unused
		qUnused := db.Model(&domain.Voucher{}).Where("agent_id = ? AND status = ?", agent.ID, "unused")
		if startDateStr != "" { qUnused = qUnused.Where("created_at >= ?", startDateStr) }
		if endDateStr != "" { qUnused = qUnused.Where("created_at <= ?", endDateStr) }
		qUnused.Count(&ap.UnusedVouchers)

		report.Agents = append(report.Agents, ap)
	}

	// --- 2.1 Agent Summary Metrics ---
	report.AgentSummary.TotalAgents = report.Overview.TotalAgents
	
	agentBatchQuery := db.Model(&domain.VoucherBatch{}).Where("agent_id != 0 AND agent_id IS NOT NULL")
	if startDateStr != "" { agentBatchQuery = agentBatchQuery.Where("created_at >= ?", startDateStr) }
	if endDateStr != "" { agentBatchQuery = agentBatchQuery.Where("created_at <= ?", endDateStr) }
	agentBatchQuery.Count(&report.AgentSummary.TotalBatches)

	agentVoucherQuery := db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("voucher_batch.agent_id != 0 AND voucher_batch.agent_id IS NOT NULL")
	
	if startDateStr != "" { agentVoucherQuery = agentVoucherQuery.Where("voucher.created_at >= ?", startDateStr) }
	if endDateStr != "" { agentVoucherQuery = agentVoucherQuery.Where("voucher.created_at <= ?", endDateStr) }
	agentVoucherQuery.Count(&report.AgentSummary.TotalVouchers)

	// Agent Costs
	agentVoucherQuery.Select("COALESCE(SUM(voucher.price), 0)").Scan(&report.AgentSummary.TotalCost)

	qAgentUsedCost := db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("(voucher_batch.agent_id != 0 AND voucher_batch.agent_id IS NOT NULL) AND voucher.status = ?", "used")
	if startDateStr != "" { qAgentUsedCost = qAgentUsedCost.Where("voucher.created_at >= ?", startDateStr) }
	if endDateStr != "" { qAgentUsedCost = qAgentUsedCost.Where("voucher.created_at <= ?", endDateStr) }
	qAgentUsedCost.Select("COALESCE(SUM(voucher.price), 0)").Scan(&report.AgentSummary.UsedCost)

	qAgentUnusedCost := db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("(voucher_batch.agent_id != 0 AND voucher_batch.agent_id IS NOT NULL) AND voucher.status = ?", "unused")
	if startDateStr != "" { qAgentUnusedCost = qAgentUnusedCost.Where("voucher.created_at >= ?", startDateStr) }
	if endDateStr != "" { qAgentUnusedCost = qAgentUnusedCost.Where("voucher.created_at <= ?", endDateStr) }
	qAgentUnusedCost.Select("COALESCE(SUM(voucher.price), 0)").Scan(&report.AgentSummary.UnusedCost)

	// 3. Admin Performance (Batches created by system/admin)
	// Admin batches have agent_id = 0 or NULL
	
	adminBatchQuery := db.Model(&domain.VoucherBatch{}).Where("agent_id = 0 OR agent_id IS NULL")
	if startDateStr != "" {
		adminBatchQuery = adminBatchQuery.Where("created_at >= ?", startDateStr)
	}
	if endDateStr != "" {
		adminBatchQuery = adminBatchQuery.Where("created_at <= ?", endDateStr)
	}
	adminBatchQuery.Count(&report.Admin.TotalBatches)

	// Admin Vouchers
	// Join with voucher_batch to filter by batch's agent_id
	baseAdminVouchers := db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("voucher_batch.agent_id = 0 OR voucher_batch.agent_id IS NULL")

	if startDateStr != "" {
		baseAdminVouchers = baseAdminVouchers.Where("voucher.created_at >= ?", startDateStr)
	}
	if endDateStr != "" {
		baseAdminVouchers = baseAdminVouchers.Where("voucher.created_at <= ?", endDateStr)
	}

	baseAdminVouchers.Count(&report.Admin.TotalVouchers)
	
	// Used
	qAdminUsed := db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("(voucher_batch.agent_id = 0 OR voucher_batch.agent_id IS NULL) AND voucher.status = ?", "used")
	if startDateStr != "" { qAdminUsed = qAdminUsed.Where("voucher.created_at >= ?", startDateStr) }
	if endDateStr != "" { qAdminUsed = qAdminUsed.Where("voucher.created_at <= ?", endDateStr) }
	qAdminUsed.Count(&report.Admin.UsedVouchers)

	// Unused
	qAdminUnused := db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("(voucher_batch.agent_id = 0 OR voucher_batch.agent_id IS NULL) AND voucher.status = ?", "unused")
	if startDateStr != "" { qAdminUnused = qAdminUnused.Where("voucher.created_at >= ?", startDateStr) }
	if endDateStr != "" { qAdminUnused = qAdminUnused.Where("voucher.created_at <= ?", endDateStr) }
	qAdminUnused.Count(&report.Admin.UnusedVouchers)

	// --- 4. Cost and Batch Detailed Metrics for Admin ---

	// Overall Costs
	db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("voucher_batch.agent_id = 0 OR voucher_batch.agent_id IS NULL").
		Select("COALESCE(SUM(voucher.price), 0)").Scan(&report.Admin.TotalCost)

	db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("(voucher_batch.agent_id = 0 OR voucher_batch.agent_id IS NULL) AND voucher.status = ?", "used").
		Select("COALESCE(SUM(voucher.price), 0)").Scan(&report.Admin.UsedCost)

	db.Model(&domain.Voucher{}).
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("(voucher_batch.agent_id = 0 OR voucher_batch.agent_id IS NULL) AND voucher.status = ?", "unused").
		Select("COALESCE(SUM(voucher.price), 0)").Scan(&report.Admin.UnusedCost)

	// Detailed Batch List
	var adminBatches []domain.VoucherBatch
	adminBatchQuery = db.Model(&domain.VoucherBatch{}).Where("agent_id = 0 OR agent_id IS NULL").Order("created_at DESC")
	if startDateStr != "" { adminBatchQuery = adminBatchQuery.Where("created_at >= ?", startDateStr) }
	if endDateStr != "" { adminBatchQuery = adminBatchQuery.Where("created_at <= ?", endDateStr) }
	adminBatchQuery.Find(&adminBatches)

	report.Admin.Batches = make([]AdminBatchDetail, 0)
	for _, b := range adminBatches {
		var detail AdminBatchDetail
		detail.ID = b.ID
		detail.Name = b.Name
		detail.Count = b.Count
		detail.CreatedAt = b.CreatedAt

		var product domain.Product
		db.Where("id = ?", b.ProductID).First(&product)
		detail.ProductName = product.Name

		db.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", b.ID, "used").Count(&detail.UsedVouchers)
		db.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", b.ID, "unused").Count(&detail.UnusedVouchers)
		db.Model(&domain.Voucher{}).Where("batch_id = ?").Select("COALESCE(SUM(price), 0)").Scan(&detail.TotalCost)

		report.Admin.Batches = append(report.Admin.Batches, detail)
	}

	return ok(c, report)
}

func registerFinancialRoutes() {
	webserver.ApiGET("/financial/report", GetFinancialReport)
}
