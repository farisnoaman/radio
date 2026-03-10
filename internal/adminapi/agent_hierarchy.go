package adminapi

import (
	"fmt"
	"net/http"
	"strconv"
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
)

// AssignAgentToParentRequest represents the request to assign an agent to a parent agent
type AssignAgentToParentRequest struct {
	AgentID        int64   `json:"agent_id" validate:"required"`
	ParentID       *int64  `json:"parent_id"` // null for root agents
	CommissionRate float64 `json:"commission_rate" validate:"min=0,max=1"`
	Territory      string  `json:"territory"`
}

// AssignAgentToParent assigns an agent to a parent agent, establishing hierarchy.
// This creates or updates the agent hierarchy relationship and sets commission rates.
//
// Only super admins and admins can assign agents to parents.
// Agents cannot assign themselves or create cycles in the hierarchy.
//
// Parameters:
//   - agent_id: ID of the agent to assign (must exist and be level='agent')
//   - parent_id: ID of the parent agent, or null to make agent a root agent
//   - commission_rate: Percentage (0.0-1.0) parent earns on this agent's sales
//   - territory: Geographic territory assignment for the agent
//
// Returns:
//   - Success: Updated agent hierarchy information
//   - Error: Validation errors or permission denied
//
// Side effects:
//   - Creates/updates agent_hierarchy record
//   - Recalculates hierarchy levels for affected agents
//   - Logs the assignment operation
func AssignAgentToParent(c echo.Context) error {
	var req AssignAgentToParentRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get current user for authorization
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	if currentUser.Level != "super" && currentUser.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only admins can manage agent hierarchy", nil)
	}

	db := GetDB(c)

	// Verify agent exists and is actually an agent
	var agent domain.SysOpr
	if err := db.First(&agent, req.AgentID).Error; err != nil {
		return fail(c, http.StatusNotFound, "AGENT_NOT_FOUND", "Agent not found", nil)
	}
	if agent.Level != "agent" {
		return fail(c, http.StatusBadRequest, "INVALID_AGENT", "User is not an agent", nil)
	}

	// If parent specified, verify parent exists and is an agent
	if req.ParentID != nil {
		var parent domain.SysOpr
		if err := db.First(&parent, *req.ParentID).Error; err != nil {
			return fail(c, http.StatusNotFound, "PARENT_NOT_FOUND", "Parent agent not found", nil)
		}
		if parent.Level != "agent" {
			return fail(c, http.StatusBadRequest, "INVALID_PARENT", "Parent user is not an agent", nil)
		}

		// Prevent self-assignment
		if *req.ParentID == req.AgentID {
			return fail(c, http.StatusBadRequest, "SELF_ASSIGNMENT", "Agent cannot be its own parent", nil)
		}

		// Prevent cycles: check if assigning this parent would create a cycle
		if wouldCreateCycle(db, req.AgentID, *req.ParentID) {
			return fail(c, http.StatusBadRequest, "CYCLE_DETECTED", "This assignment would create a cycle in the hierarchy", nil)
		}
	}

	tx := db.Begin()

	// Create or update hierarchy record
	hierarchy := domain.AgentHierarchy{
		AgentID:        req.AgentID,
		ParentID:       req.ParentID,
		CommissionRate: req.CommissionRate,
		Territory:      req.Territory,
		Status:         "active",
		UpdatedAt:      time.Now(),
	}

	// Calculate the correct level
	level, err := calculateHierarchyLevel(tx, req.ParentID)
	if err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "LEVEL_CALCULATION_FAILED", "Failed to calculate hierarchy level", err.Error())
	}
	hierarchy.Level = level

	if err := tx.Save(&hierarchy).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "SAVE_FAILED", "Failed to save hierarchy", err.Error())
	}

	// Update levels for all descendants
	if err := updateDescendantLevels(tx, req.AgentID); err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "DESCENDANT_UPDATE_FAILED", "Failed to update descendant levels", err.Error())
	}

	tx.Commit()

	zap.L().Info("Agent assigned to parent",
		zap.Int64("agent_id", req.AgentID),
		zap.Any("parent_id", req.ParentID),
		zap.Float64("commission_rate", req.CommissionRate),
		zap.String("territory", req.Territory),
		zap.Int("level", hierarchy.Level))

	return ok(c, map[string]interface{}{
		"hierarchy": hierarchy,
		"message":   "Agent hierarchy updated successfully",
	})
}

// GetAgentHierarchy retrieves the hierarchy information for a specific agent
func GetAgentHierarchy(c echo.Context) error {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	var hierarchy domain.AgentHierarchy
	if err := GetDB(c).Where("agent_id = ?", agentID).First(&hierarchy).Error; err != nil {
		// Return empty hierarchy if not found (agent might be root)
		hierarchy = domain.AgentHierarchy{
			AgentID: agentID,
			Status:  "active",
		}
	}

	return ok(c, hierarchy)
}

// GetAgentSubAgents retrieves all direct sub-agents of a given agent
func GetAgentSubAgents(c echo.Context) error {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	// Pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	var subAgents []struct {
		domain.AgentHierarchy
		AgentName  string `json:"agent_name"`
		AgentEmail string `json:"agent_email"`
	}

	baseQuery := GetDB(c).Table("agent_hierarchy").
		Select("agent_hierarchy.*, sys_opr.realname as agent_name, sys_opr.email as agent_email").
		Joins("JOIN sys_opr ON agent_hierarchy.agent_id = sys_opr.id").
		Where("agent_hierarchy.parent_id = ? AND agent_hierarchy.status = ?", agentID, "active").
		Order("agent_hierarchy.created_at DESC")

	// Get total count
	var total int64
	GetDB(c).Model(&domain.AgentHierarchy{}).
		Where("parent_id = ? AND status = ?", agentID, "active").
		Count(&total)

	// Get paginated results
	if err := baseQuery.Offset(offset).Limit(perPage).Find(&subAgents).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query sub-agents", err.Error())
	}

	return paged(c, subAgents, total, page, perPage)
}

// GetAgentHierarchyTree retrieves the complete hierarchy tree for an agent
func GetAgentHierarchyTree(c echo.Context) error {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	tree, err := buildHierarchyTree(GetDB(c), agentID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TREE_BUILD_FAILED", "Failed to build hierarchy tree", err.Error())
	}

	return ok(c, tree)
}

// UpdateAgentCommissionRate updates the commission rate for a specific agent relationship
func UpdateAgentCommissionRate(c echo.Context) error {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	var req struct {
		CommissionRate float64 `json:"commission_rate" validate:"min=0,max=1"`
	}
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get current user for authorization
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	if currentUser.Level != "super" && currentUser.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only admins can update commission rates", nil)
	}

	if err := GetDB(c).Model(&domain.AgentHierarchy{}).
		Where("agent_id = ?", agentID).
		Update("commission_rate", req.CommissionRate).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update commission rate", err.Error())
	}

	zap.L().Info("Agent commission rate updated",
		zap.Int64("agent_id", agentID),
		zap.Float64("commission_rate", req.CommissionRate))

	return ok(c, map[string]interface{}{
		"message": "Commission rate updated successfully",
	})
}

// Helper functions

// wouldCreateCycle checks if assigning parentID to agentID would create a cycle
func wouldCreateCycle(db interface{}, agentID, parentID int64) bool {
	// Walk up the hierarchy from parentID, checking if we encounter agentID
	currentID := parentID
	for currentID != 0 {
		if currentID == agentID {
			return true // Cycle detected
		}

		var hierarchy domain.AgentHierarchy
		if err := db.(*gorm.DB).Where("agent_id = ?", currentID).First(&hierarchy).Error; err != nil {
			break // No parent, end of chain
		}
		if hierarchy.ParentID == nil {
			break // Root agent
		}
		currentID = *hierarchy.ParentID
	}
	return false
}

// calculateHierarchyLevel calculates the correct level for an agent based on its parent.
// If the parent has no hierarchy record yet (i.e. is a root-level agent), we treat
// the parent's level as 0. This allows assigning the first child to a freshly created
// agent without requiring a pre-existing hierarchy entry for the parent.
func calculateHierarchyLevel(db interface{}, parentID *int64) (int, error) {
	if parentID == nil {
		return 0, nil // Root level
	}

	var parentHierarchy domain.AgentHierarchy
	err := db.(*gorm.DB).Where("agent_id = ?", *parentID).First(&parentHierarchy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// parent has no hierarchy entry yet; treat as root
			return 1, nil
		}
		// unexpected database error
		return 0, err
	}

	return parentHierarchy.Level + 1, nil
}

// updateDescendantLevels recursively updates hierarchy levels for all descendants
func updateDescendantLevels(db interface{}, agentID int64) error {
	// Get all direct children
	var children []domain.AgentHierarchy
	if err := db.(*gorm.DB).Where("parent_id = ?", agentID).Find(&children).Error; err != nil {
		return err
	}

	// Update each child's level and recurse
	for _, child := range children {
		newLevel, err := calculateHierarchyLevel(db, child.ParentID)
		if err != nil {
			return err
		}

		if err := db.(*gorm.DB).Model(&child).Update("level", newLevel).Error; err != nil {
			return err
		}

		// Recurse to children
		if err := updateDescendantLevels(db, child.AgentID); err != nil {
			return err
		}
	}

	return nil
}

// buildHierarchyTree builds a nested tree structure of the agent's hierarchy
func buildHierarchyTree(db interface{}, agentID int64) (map[string]interface{}, error) {
	// Get agent info
	var agent domain.SysOpr
	if err := db.(*gorm.DB).First(&agent, agentID).Error; err != nil {
		return nil, err
	}

	// Get hierarchy info
	var hierarchy domain.AgentHierarchy
	hierarchyErr := db.(*gorm.DB).Where("agent_id = ?", agentID).First(&hierarchy).Error

	// Build tree node
	node := map[string]interface{}{
		"id":       agent.ID,
		"name":     agent.Realname,
		"username": agent.Username,
		"email":    agent.Email,
		"level":    0,
		"territory": "",
		"commission_rate": 0.0,
		"children": []map[string]interface{}{},
	}

	if hierarchyErr == nil {
		node["level"] = hierarchy.Level
		node["territory"] = hierarchy.Territory
		node["commission_rate"] = hierarchy.CommissionRate
	}

	// Get children
	var children []struct {
		domain.AgentHierarchy
		AgentName  string `json:"agent_name"`
		AgentEmail string `json:"agent_email"`
	}

	if err := db.(*gorm.DB).Table("agent_hierarchy").
		Select("agent_hierarchy.*, sys_opr.realname as agent_name, sys_opr.email as agent_email").
		Joins("JOIN sys_opr ON agent_hierarchy.agent_id = sys_opr.id").
		Where("agent_hierarchy.parent_id = ? AND agent_hierarchy.status = ?", agentID, "active").
		Find(&children).Error; err != nil {
		return nil, err
	}

	// Recursively build children
	for _, child := range children {
		childTree, err := buildHierarchyTree(db, child.AgentID)
		if err != nil {
			return nil, err
		}
		children := node["children"].([]map[string]interface{})
		children = append(children, childTree)
		node["children"] = children
	}

	return node, nil
}

// CalculateCommissions calculates and records commissions for an agent hierarchy
// when a voucher is redeemed. Traverses up the hierarchy and distributes commissions
// based on predefined rates per level.
//
// Commission rates (hardcoded for Phase 1):
//   - Level 1 (direct agent): 5% of voucher price
//   - Level 2 (parent): 2% of voucher price
//   - Level 3 (grandparent): 1% of voucher price
//
// Parameters:
//   - db: Database connection
//   - agentID: The agent who sold the voucher (starting point)
//   - voucherID: ID of the redeemed voucher
//   - voucherPrice: Price of the voucher for commission calculation
//
// Side effects:
//   - Creates CommissionLog entries for each level
//   - Updates or creates CommissionSummary records for monthly totals
//   - Credits agent wallets with commission amounts
func CalculateCommissions(db *gorm.DB, agentID int64, voucherID int64, voucherPrice float64) error {
	if agentID == 0 || voucherPrice <= 0 {
		return nil // No commissions for system vouchers or free vouchers
	}

	// Commission rates per level (0-based index)
	commissionRates := []float64{0.05, 0.02, 0.01} // 5%, 2%, 1%
	maxLevels := len(commissionRates)

	currentAgentID := agentID
	level := 0

	for level < maxLevels {
		// Get agent details
		var agent domain.SysOpr
		if err := db.First(&agent, currentAgentID).Error; err != nil {
			zap.L().Error("failed to find agent for commission",
				zap.Error(err), zap.Int64("agent_id", currentAgentID))
			break // Stop if agent not found
		}

		if agent.Level != "agent" {
			break // Only agents get commissions, not admins
		}

		// Calculate commission for this level
		rate := commissionRates[level]
		commissionAmount := voucherPrice * rate

		if commissionAmount <= 0 {
			break
		}

		now := time.Now()
		periodKey := now.Format("2006-01") // YYYY-MM format

		// Determine commission type based on level
		commissionType := "direct_sale"
		if level == 1 {
			commissionType = "referral"
		} else if level >= 2 {
			commissionType = "override"
		}

		// Use transaction for atomicity
		tx := db.Begin()

		// Create commission log
		commissionLog := domain.CommissionLog{
			AgentID:       currentAgentID,
			SourceAgentID: agentID, // Original selling agent
			VoucherID:     &voucherID,
			Amount:        commissionAmount,
			Type:          commissionType,
			Level:         level,
			CreatedAt:     now,
		}

		if err := tx.Create(&commissionLog).Error; err != nil {
			tx.Rollback()
			zap.L().Error("failed to create commission log",
				zap.Error(err), zap.Int64("agent_id", currentAgentID))
			return err
		}

		// Update or create commission summary
		var summary domain.CommissionSummary
		err := tx.Where("agent_id = ? AND period = ?", currentAgentID, periodKey).
			First(&summary).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			tx.Rollback()
			zap.L().Error("failed to find commission summary",
				zap.Error(err), zap.Int64("agent_id", currentAgentID))
			return err
		}

		if err == gorm.ErrRecordNotFound {
			// Create new summary
			summary = domain.CommissionSummary{
				AgentID:          currentAgentID,
				Period:           periodKey,
				TotalEarned:      commissionAmount,
				PendingAmount:    commissionAmount,
				TransactionCount: 1,
				LastUpdated:      now,
			}
			// Set the appropriate sales type
			switch commissionType {
			case "direct_sale":
				summary.DirectSales = commissionAmount
			case "referral":
				summary.ReferralSales = commissionAmount
			case "override":
				summary.OverrideSales = commissionAmount
			}
			if err := tx.Create(&summary).Error; err != nil {
				tx.Rollback()
				zap.L().Error("failed to create commission summary",
					zap.Error(err), zap.Int64("agent_id", currentAgentID))
				return err
			}
		} else {
			// Update existing summary
			summary.TotalEarned += commissionAmount
			summary.PendingAmount += commissionAmount
			summary.TransactionCount += 1
			summary.LastUpdated = now
			// Update the appropriate sales type
			switch commissionType {
			case "direct_sale":
				summary.DirectSales += commissionAmount
			case "referral":
				summary.ReferralSales += commissionAmount
			case "override":
				summary.OverrideSales += commissionAmount
			}
			if err := tx.Save(&summary).Error; err != nil {
				tx.Rollback()
				zap.L().Error("failed to update commission summary",
					zap.Error(err), zap.Int64("agent_id", currentAgentID))
				return err
			}
		}

		// Credit agent wallet
		var wallet domain.AgentWallet
		if err := tx.Where("agent_id = ?", currentAgentID).First(&wallet).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create wallet if not exists
				wallet = domain.AgentWallet{
					AgentID:    currentAgentID,
					Balance:    commissionAmount,
					UpdatedAt:  now,
				}
				if err := tx.Create(&wallet).Error; err != nil {
					tx.Rollback()
					zap.L().Error("failed to create agent wallet",
						zap.Error(err), zap.Int64("agent_id", currentAgentID))
					return err
				}
			} else {
				tx.Rollback()
				zap.L().Error("failed to find agent wallet",
					zap.Error(err), zap.Int64("agent_id", currentAgentID))
				return err
			}
		} else {
			// Update existing wallet
			wallet.Balance += commissionAmount
			wallet.UpdatedAt = now
			if err := tx.Save(&wallet).Error; err != nil {
				tx.Rollback()
				zap.L().Error("failed to update agent wallet",
					zap.Error(err), zap.Int64("agent_id", currentAgentID))
				return err
			}
		}

		tx.Commit()

		zap.L().Info("commission credited",
			zap.Int64("agent_id", currentAgentID),
			zap.Float64("amount", commissionAmount),
			zap.String("type", commissionType),
			zap.Int("level", level),
			zap.Int64("voucher_id", voucherID))

		// Move to parent agent
		var hierarchy domain.AgentHierarchy
		if err := db.Where("agent_id = ? AND status = ?", currentAgentID, "active").
			First(&hierarchy).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				break // No parent, end of hierarchy
			}
			zap.L().Error("failed to find agent hierarchy",
				zap.Error(err), zap.Int64("agent_id", currentAgentID))
			break
		}

		if hierarchy.ParentID == nil {
			break // Root agent, no more parents
		}

		currentAgentID = *hierarchy.ParentID
		level++
	}

	return nil
}

// RequestPayout allows an agent to request a payout from their wallet balance.
// The request must be approved by an admin before funds are transferred.
//
// Parameters:
//   - amount: Amount to withdraw (must be <= wallet balance)
//   - payment_method: How to receive payment (bank_transfer, paypal, etc.)
//   - payment_details: Payment method specific details
//   - notes: Optional notes from the agent
//
// Returns:
//   - Success: Created payout request with pending status
//   - Error: Insufficient balance, invalid amount, etc.
func RequestPayout(c echo.Context) error {
	// Get agent from context
	agentID, err := getAgentIDFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to identify agent", nil)
	}

	var req struct {
		Amount         float64 `json:"amount" validate:"required,min=0.01"`
		PaymentMethod  string  `json:"payment_method" validate:"required,oneof=bank_transfer paypal crypto cash check"`
		PaymentDetails string  `json:"payment_details" validate:"required"`
		Notes          string  `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Check wallet balance
	var wallet domain.AgentWallet
	if err := GetDB(c).Where("agent_id = ?", agentID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fail(c, http.StatusBadRequest, "INSUFFICIENT_BALANCE", "No wallet found", nil)
		}
		return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to check balance", err.Error())
	}

	if wallet.Balance < req.Amount {
		return fail(c, http.StatusBadRequest, "INSUFFICIENT_BALANCE",
			fmt.Sprintf("Requested amount %.2f exceeds available balance %.2f", req.Amount, wallet.Balance), nil)
	}

	// Create payout request
	payoutRequest := domain.PayoutRequest{
		AgentID:        agentID,
		Amount:         req.Amount,
		PaymentMethod:  req.PaymentMethod,
		PaymentDetails: req.PaymentDetails,
		Status:         "pending",
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := GetDB(c).Create(&payoutRequest).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create payout request", err.Error())
	}

	LogOperation(c, "request_payout", fmt.Sprintf("Agent %d requested payout of %.2f", agentID, req.Amount))

	return ok(c, payoutRequest)
}

// ApprovePayout allows admins to approve or reject payout requests.
// Approved payouts create a PayoutLog and deduct from wallet balance.
//
// Parameters:
//   - request_id: ID of the payout request
//   - action: "approve" or "reject"
//   - admin_notes: Optional notes from admin
//   - transaction_id: For approved payouts, external payment reference
//
// Returns:
//   - Success: Updated payout request with new status
//   - Error: Permission denied, invalid request, etc.
func ApprovePayout(c echo.Context) error {
	// Check admin permissions
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user", nil)
	}

	if currentOpr.Level != "super" && currentOpr.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only admins can approve payouts", nil)
	}

	var req struct {
		RequestID    int64  `json:"request_id" validate:"required"`
		Action       string `json:"action" validate:"required,oneof=approve reject"`
		AdminNotes   string `json:"admin_notes"`
		TransactionID string `json:"transaction_id"` // For approved payouts
	}

	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Find payout request
	var payoutRequest domain.PayoutRequest
	if err := GetDB(c).First(&payoutRequest, req.RequestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fail(c, http.StatusNotFound, "NOT_FOUND", "Payout request not found", nil)
		}
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to find payout request", err.Error())
	}

	if payoutRequest.Status != "pending" {
		return fail(c, http.StatusBadRequest, "INVALID_STATUS",
			fmt.Sprintf("Payout request is already %s", payoutRequest.Status), nil)
	}

	now := time.Now()
	payoutRequest.ReviewedBy = &currentOpr.ID
	payoutRequest.ReviewedAt = &now
	payoutRequest.AdminNotes = req.AdminNotes
	payoutRequest.UpdatedAt = now

	// Use transaction for atomicity
	tx := GetDB(c).Begin()

	if req.Action == "approve" {
		// Check wallet balance again (in case it changed)
		var wallet domain.AgentWallet
		if err := tx.Where("agent_id = ?", payoutRequest.AgentID).First(&wallet).Error; err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to check wallet", err.Error())
		}

		if wallet.Balance < payoutRequest.Amount {
			tx.Rollback()
			return fail(c, http.StatusBadRequest, "INSUFFICIENT_BALANCE",
				"Wallet balance insufficient for payout", nil)
		}

		// Deduct from wallet
		wallet.Balance -= payoutRequest.Amount
		wallet.UpdatedAt = now
		if err := tx.Save(&wallet).Error; err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "WALLET_UPDATE_FAILED", "Failed to update wallet", err.Error())
		}

		// Create payout log
		payoutLog := domain.PayoutLog{
			AgentID:         payoutRequest.AgentID,
			PayoutRequestID: payoutRequest.ID,
			Amount:          payoutRequest.Amount,
			PaymentMethod:   payoutRequest.PaymentMethod,
			PaymentDetails:  payoutRequest.PaymentDetails,
			TransactionID:   req.TransactionID,
			ProcessedBy:     currentOpr.ID,
			ProcessedAt:     now,
			Notes:           req.AdminNotes,
		}

		if err := tx.Create(&payoutLog).Error; err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "LOG_CREATE_FAILED", "Failed to create payout log", err.Error())
		}

		payoutRequest.Status = "completed"

		LogOperation(c, "approve_payout", fmt.Sprintf("Approved payout of %.2f for agent %d", payoutRequest.Amount, payoutRequest.AgentID))

	} else { // reject
		payoutRequest.Status = "rejected"
		LogOperation(c, "reject_payout", fmt.Sprintf("Rejected payout request %d for agent %d", payoutRequest.ID, payoutRequest.AgentID))
	}

	if err := tx.Save(&payoutRequest).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update payout request", err.Error())
	}

	tx.Commit()

	return ok(c, payoutRequest)
}

// ListPayoutRequests retrieves payout requests with filtering and pagination.
// Agents see only their own requests, admins see all requests.
//
// Query parameters:
//   - status: Filter by status (pending, approved, rejected, completed)
//   - agent_id: Admin filter by specific agent
//   - page, per_page: Pagination
//
// Returns:
//   - List of payout requests with agent information
func ListPayoutRequests(c echo.Context) error {
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user", nil)
	}

	isAdmin := currentOpr.Level == "super" || currentOpr.Level == "admin"

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	query := GetDB(c).Table("payout_request").
		Select(`payout_request.*,
		        sys_opr.username as agent_username,
		        sys_opr.realname as agent_realname`).
		Joins("JOIN sys_opr ON payout_request.agent_id = sys_opr.id")

	// Non-admin agents can only see their own requests
	if !isAdmin {
		query = query.Where("payout_request.agent_id = ?", currentOpr.ID)
	}

	// Filters
	if status := c.QueryParam("status"); status != "" {
		query = query.Where("payout_request.status = ?", status)
	}

	if agentID := c.QueryParam("agent_id"); agentID != "" && isAdmin {
		query = query.Where("payout_request.agent_id = ?", agentID)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get results
	var requests []map[string]interface{}
	if err := query.Order("payout_request.created_at DESC").
		Limit(perPage).Offset(offset).
		Find(&requests).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve payout requests", err.Error())
	}

	return paged(c, requests, total, page, perPage)
}

// GetPayoutRequest retrieves details of a specific payout request.
// Agents can only view their own requests, admins can view any request.
//
// Parameters:
//   - id: Payout request ID
//
// Returns:
//   - Payout request details with agent information
func GetPayoutRequest(c echo.Context) error {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid payout request ID", nil)
	}

	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user", nil)
	}

	isAdmin := currentOpr.Level == "super" || currentOpr.Level == "admin"

	var request map[string]interface{}
	query := GetDB(c).Table("payout_request").
		Select(`payout_request.*,
		        sys_opr.username as agent_username,
		        sys_opr.realname as agent_realname`).
		Joins("JOIN sys_opr ON payout_request.agent_id = sys_opr.id").
		Where("payout_request.id = ?", requestID)

	// Non-admin agents can only see their own requests
	if !isAdmin {
		query = query.Where("payout_request.agent_id = ?", currentOpr.ID)
	}

	if err := query.First(&request).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fail(c, http.StatusNotFound, "NOT_FOUND", "Payout request not found", nil)
		}
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve payout request", err.Error())
	}

	return ok(c, request)
}

// CancelPayoutRequest allows agents to cancel their own pending payout requests.
//
// Parameters:
//   - id: Payout request ID
//
// Returns:
//   - Success: Updated payout request with cancelled status
//   - Error: Permission denied, invalid status, etc.
func CancelPayoutRequest(c echo.Context) error {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid payout request ID", nil)
	}

	// Get agent from context
	agentID, err := getAgentIDFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to identify agent", nil)
	}

	// Find and validate payout request
	var payoutRequest domain.PayoutRequest
	if err := GetDB(c).Where("id = ? AND agent_id = ?", requestID, agentID).First(&payoutRequest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fail(c, http.StatusNotFound, "NOT_FOUND", "Payout request not found", nil)
		}
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to find payout request", err.Error())
	}

	if payoutRequest.Status != "pending" {
		return fail(c, http.StatusBadRequest, "INVALID_STATUS",
			fmt.Sprintf("Cannot cancel payout request with status: %s", payoutRequest.Status), nil)
	}

	// Update status to cancelled
	payoutRequest.Status = "cancelled"
	payoutRequest.UpdatedAt = time.Now()

	if err := GetDB(c).Save(&payoutRequest).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to cancel payout request", err.Error())
	}

	LogOperation(c, "cancel_payout", fmt.Sprintf("Agent %d cancelled payout request %d", agentID, requestID))

	return ok(c, payoutRequest)
}

// GetAgentPayoutHistory retrieves completed payouts for an agent.
// Includes payout details and transaction information.
//
// Parameters:
//   - page, per_page: Pagination
//   - start_date, end_date: Date range filter
//
// Returns:
//   - List of completed payouts with transaction details
func GetAgentPayoutHistory(c echo.Context) error {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	// Check permissions - agents can only view their own history, admins can view any
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user", nil)
	}

	isAdmin := currentOpr.Level == "super" || currentOpr.Level == "admin"
	if !isAdmin && currentOpr.ID != agentID {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Can only view own payout history", nil)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	query := GetDB(c).Table("payout_log").
		Select(`payout_log.*,
		        payout_request.notes as request_notes,
		        payout_request.admin_notes as admin_notes,
		        sys_opr.username as processed_by_username`).
		Joins("JOIN payout_request ON payout_log.payout_request_id = payout_request.id").
		Joins("LEFT JOIN sys_opr ON payout_log.processed_by = sys_opr.id").
		Where("payout_log.agent_id = ?", agentID)

	// Date range filter
	if startDate := c.QueryParam("start_date"); startDate != "" {
		if startTime, err := time.Parse(time.RFC3339, startDate); err == nil {
			query = query.Where("payout_log.processed_at >= ?", startTime)
		}
	}

	if endDate := c.QueryParam("end_date"); endDate != "" {
		if endTime, err := time.Parse(time.RFC3339, endDate); err == nil {
			query = query.Where("payout_log.processed_at <= ?", endTime)
		}
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get results
	var payouts []map[string]interface{}
	if err := query.Order("payout_log.processed_at DESC").
		Limit(perPage).Offset(offset).
		Find(&payouts).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve payout history", err.Error())
	}

	return paged(c, payouts, total, page, perPage)
}

// GetAgentPerformance returns performance metrics for an agent.
// Includes sales data, commission earnings, and key performance indicators.
//
// Parameters:
//   - period: Time period (daily, weekly, monthly, yearly)
//   - start_date, end_date: Custom date range
//
// Returns:
//   - Comprehensive performance metrics and analytics
func GetAgentPerformance(c echo.Context) error {
	agentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid agent ID", nil)
	}

	// Check permissions - agents can only view their own performance, admins can view any
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user", nil)
	}

	isAdmin := currentOpr.Level == "super" || currentOpr.Level == "admin"
	if !isAdmin && currentOpr.ID != agentID {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Can only view own performance", nil)
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "monthly"
	}

	startDate, endDate := getDateRange(period, c.QueryParam("start_date"), c.QueryParam("end_date"))

	// Calculate performance metrics
	metrics, err := calculateAgentPerformance(GetDB(c), agentID, startDate, endDate)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "CALCULATION_FAILED", "Failed to calculate performance", err.Error())
	}

	return ok(c, metrics)
}

// calculateAgentPerformance computes comprehensive performance metrics for an agent
func calculateAgentPerformance(db *gorm.DB, agentID int64, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Voucher sales metrics
	var voucherStats struct {
		TotalSold     int64   `json:"total_sold"`
		TotalRevenue  float64 `json:"total_revenue"`
		ActiveUsers   int64   `json:"active_users"`
		AvgOrderValue float64 `json:"avg_order_value"`
	}

	// Count vouchers sold by this agent
	if err := db.Table("voucher").
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("voucher_batch.agent_id = ? AND voucher.status = 'active' AND voucher.activated_at BETWEEN ? AND ?",
			agentID, startDate, endDate).
		Select("COUNT(*) as total_sold, COALESCE(SUM(voucher.price), 0) as total_revenue").
		Scan(&voucherStats).Error; err != nil {
		return nil, err
	}

	// Count active users from sold vouchers
	if err := db.Table("radius_user").
		Joins("JOIN voucher ON radius_user.username = voucher.code").
		Joins("JOIN voucher_batch ON voucher.batch_id = voucher_batch.id").
		Where("voucher_batch.agent_id = ? AND radius_user.status = 'enabled' AND voucher.activated_at BETWEEN ? AND ?",
			agentID, startDate, endDate).
		Count(&voucherStats.ActiveUsers).Error; err != nil {
		return nil, err
	}

	if voucherStats.TotalSold > 0 {
		voucherStats.AvgOrderValue = voucherStats.TotalRevenue / float64(voucherStats.TotalSold)
	}

	// Commission metrics
	var commissionStats struct {
		TotalEarned     float64 `json:"total_earned"`
		TotalPaid       float64 `json:"total_paid"`
		PendingAmount   float64 `json:"pending_amount"`
		DirectSales     float64 `json:"direct_sales"`
		ReferralSales   float64 `json:"referral_sales"`
		OverrideSales   float64 `json:"override_sales"`
	}

	periodKey := startDate.Format("2006-01")
	if err := db.Table("commission_summary").
		Where("agent_id = ? AND period = ?", agentID, periodKey).
		First(&commissionStats).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Wallet balance
	var wallet domain.AgentWallet
	db.Where("agent_id = ?", agentID).First(&wallet)

	// Recent activity (last 10 transactions)
	var recentActivity []map[string]interface{}
	db.Table("commission_log").
		Select("commission_log.amount, commission_log.type, commission_log.created_at, voucher.code as voucher_code").
		Joins("LEFT JOIN voucher ON commission_log.voucher_id = voucher.id").
		Where("commission_log.agent_id = ?", agentID).
		Order("commission_log.created_at DESC").
		Limit(10).
		Find(&recentActivity)

	return map[string]interface{}{
		"agent_id":         agentID,
		"period":           periodKey,
		"date_range": map[string]string{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
		"voucher_sales":    voucherStats,
		"commissions":      commissionStats,
		"wallet_balance":   wallet.Balance,
		"recent_activity":  recentActivity,
		"performance_score": calculatePerformanceScore(voucherStats, commissionStats),
	}, nil
}

// calculatePerformanceScore computes a performance score (0-100) based on metrics
func calculatePerformanceScore(voucherStats interface{}, commissionStats interface{}) float64 {
	// Simple scoring algorithm - can be enhanced with ML models later
	score := 0.0

	// This is a placeholder - implement based on business requirements
	// For now, return a basic score
	return score
}

// getDateRange returns start and end dates based on period
func getDateRange(period, startParam, endParam string) (time.Time, time.Time) {
	now := time.Now()

	switch period {
	case "daily":
		start := now.Truncate(24 * time.Hour)
		return start, start.AddDate(0, 0, 1)
	case "weekly":
		start := now.AddDate(0, 0, -int(now.Weekday()))
		return start, start.AddDate(0, 0, 7)
	case "yearly":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return start, start.AddDate(1, 0, 0)
	case "monthly":
		fallthrough
	default:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 1, 0)
	}
}

// getAgentIDFromContext extracts agent ID from JWT context
func getAgentIDFromContext(c echo.Context) (int64, error) {
	// This would extract from JWT claims - placeholder implementation
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return 0, err
	}
	return currentOpr.ID, nil
}

// Register agent hierarchy routes
func registerAgentHierarchyRoutes() {
	webserver.ApiPOST("/agents/:id/assign-parent", AssignAgentToParent)
	webserver.ApiGET("/agents/:id/hierarchy", GetAgentHierarchy)
	webserver.ApiGET("/agents/:id/sub-agents", GetAgentSubAgents)
	webserver.ApiGET("/agents/:id/hierarchy-tree", GetAgentHierarchyTree)
	webserver.ApiPUT("/agents/:id/commission-rate", UpdateAgentCommissionRate)

	// Payout processing routes
	webserver.ApiPOST("/agents/payout/request", RequestPayout)
	webserver.ApiPOST("/agents/payout/:id/cancel", CancelPayoutRequest)
	webserver.ApiGET("/agents/:id/payout-history", GetAgentPayoutHistory)
	webserver.ApiGET("/payout/requests/:id", GetPayoutRequest)
	webserver.ApiPOST("/admin/payout/approve", ApprovePayout)
	webserver.ApiGET("/payout/requests", ListPayoutRequests)

	// Performance analytics routes
	webserver.ApiGET("/agents/:id/performance", GetAgentPerformance)

	// Root agents list
	webserver.ApiGET("/agents/roots", GetRootAgents)
}

// GetRootAgents retrieves all root-level agents (no parent)
func GetRootAgents(c echo.Context) error {
	var agents []struct {
		domain.AgentHierarchy
		AgentName  string `json:"agent_name"`
		AgentEmail string `json:"agent_email"`
	}

	query := GetDB(c).Table("agent_hierarchy").
		Select("agent_hierarchy.*, sys_opr.realname as agent_name, sys_opr.email as agent_email").
		Joins("JOIN sys_opr ON agent_hierarchy.agent_id = sys_opr.id").
		Where("agent_hierarchy.parent_id IS NULL AND agent_hierarchy.status = ?", "active")

	if err := query.Find(&agents).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query root agents", err.Error())
	}

	return ok(c, agents)
}
