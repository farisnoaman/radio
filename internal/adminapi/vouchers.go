package adminapi

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/coa"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/checkers"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// init function removed to avoid duplication/conflict if registerVoucherRoutes is used.

// ListVoucherBatches retrieves the voucher batch list
// @Summary get the voucher batch list
// @Tags Voucher
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Success 200 {object} ListResponse
// @Router /api/v1/voucher-batches [get]
func ListVoucherBatches(c echo.Context) error {
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	if tenantID == 0 {
		return fail(c, http.StatusBadRequest, "NO_TENANT", "Missing tenant context", nil)
	}

	db := GetDB(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 10000 {
		perPage = 100
	}

	sortField := c.QueryParam("sort")
	order := c.QueryParam("order")
	if sortField == "" {
		sortField = "id"
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	var total int64
	var batches []domain.VoucherBatch

	// Permission check
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	query := db.Model(&domain.VoucherBatch{}).Where("tenant_id = ?", tenantID)

	// Filter by agent: Agents can only see their own batches
	if currentUser.Level == "agent" {
		query = query.Where("agent_id = ? AND is_deleted = ?", currentUser.ID, false)
	} else {
		// Admin: show deleted batches if requested, or filter out by default
		if c.QueryParam("is_deleted") == "true" {
			query = query.Where("is_deleted = ?", true)
		} else if c.QueryParam("is_deleted") == "false" {
			query = query.Where("is_deleted = ?", false)
		}
	}

	// Filter by name
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&batches)

	return paged(c, batches, total, page, perPage)
}

// ListVouchers retrieves the voucher list
// @Summary get the voucher list
// @Tags Voucher
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param batch_id query int false "Batch ID"
// @Param status query string false "Status"
// @Success 200 {object} ListResponse
// @Router /api/v1/vouchers [get]
func ListVouchers(c echo.Context) error {
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	if tenantID == 0 {
		return fail(c, http.StatusBadRequest, "NO_TENANT", "Missing tenant context", nil)
	}

	db := GetDB(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 10000 {
		perPage = 100
	}

	var total int64
	var vouchers []domain.Voucher

	query := db.Model(&domain.Voucher{}).Where("tenant_id = ?", tenantID)

	if batchID := c.QueryParam("batch_id"); batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}

	if status := c.QueryParam("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if code := c.QueryParam("code"); code != "" {
		query = query.Where("code = ?", code)
	}

	// Search by SN (batchid-voucherid format), voucher ID, or exact code match
	if sn := c.QueryParam("sn"); sn != "" {
		parts := strings.Split(sn, "-")
		if len(parts) == 2 {
			// SN format: batchid-voucherid
			batchID, err1 := strconv.ParseInt(parts[0], 10, 64)
			voucherID, err2 := strconv.ParseInt(parts[1], 10, 64)
			if err1 == nil && err2 == nil {
				query = query.Where("batch_id = ? AND id = ?", batchID, voucherID)
			}
		} else if _, err := strconv.ParseInt(sn, 10, 64); err == nil {
			// Pure numeric input - search by voucher ID
			voucherID, _ := strconv.ParseInt(sn, 10, 64)
			query = query.Where("id = ?", voucherID)
		} else {
			// Exact code match
			query = query.Where("code = ?", sn)
		}
	}

	// Filter by agent: Agents can only see their own vouchers
	currentUser, err := resolveOperatorFromContext(c)
	if err == nil && currentUser.Level == "agent" {
		query = query.Where("agent_id = ?", currentUser.ID)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order("id DESC").Limit(perPage).Offset(offset).Find(&vouchers)

	return paged(c, vouchers, total, page, perPage)
}

// VoucherBatchRequest
type VoucherBatchRequest struct {
	Name        string `json:"name" validate:"required"`
	ProductID   string `json:"product_id" validate:"required"`
	Count       int    `json:"count" validate:"required,min=1,max=10000"`
	Prefix      string `json:"prefix" validate:"omitempty,max=10"`
	Length      int    `json:"length" validate:"omitempty,min=6,max=20"` // Length of random part
	Type        string `json:"type"`                                     // number, alpha, mixed
	Remark      string `json:"remark"`
	AgentID     string `json:"agent_id" validate:"omitempty"`               // Optional, if generated by agent
	ExpireTime  string `json:"expire_time"`                                 // ISO8601 string
	GeneratePIN bool   `json:"generate_pin"`                                // Generate PIN for vouchers
	PINLength   int    `json:"pin_length" validate:"omitempty,min=4,max=8"` // PIN length (default 4)
	// First-Use Expiration options
	ExpirationType string `json:"expiration_type"`                                   // "fixed" (default) or "first_use"
	ValidityDays   int    `json:"validity_days" validate:"omitempty,min=1,max=8760"` // Hours of validity for first_use type (1-8760 hours = 1-365 days)
}

// Local generation functions removed in favor of pkg/common

// CreateVoucherBatch generates vouchers
// @Summary create voucher batch
// @Tags Voucher
// @Param batch body VoucherBatchRequest true "Batch info"
// @Success 201 {object} domain.VoucherBatch
// @Router /api/v1/voucher-batches [post]
func CreateVoucherBatch(c echo.Context) error {
	var req VoucherBatchRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// DEBUG: Log request details
	zap.L().Info("CreateVoucherBatch: Request received",
		zap.String("name", req.Name),
		zap.String("expiration_type", req.ExpirationType),
		zap.Int("validity_days", req.ValidityDays),
		zap.String("expire_time", req.ExpireTime),
		zap.Int("count", req.Count))

	if req.Length == 0 {
		req.Length = 10
	}

	productID, _ := strconv.ParseInt(req.ProductID, 10, 64)

	// Validate Product
	var product domain.Product
	if err := GetDB(c).First(&product, productID).Error; err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_PRODUCT", "Product not found", nil)
	}

	// Securely get current user to enforce AgentID
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	// If user is an agent, force the AgentID to be their own ID
	if currentUser.Level == "agent" {
		req.AgentID = fmt.Sprintf("%d", currentUser.ID)
	}

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	// Start Transaction
	tx := GetDB(c).Begin()

	// Auto-generate batch name if not provided or matches default pattern
	// This ensures unique naming even with concurrent requests
	if req.Name == "" || isDefaultBatchNamePattern(req.Name) {
		newName, err := generateNextBatchName(tx, req.Name)
		if err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "NAME_ERROR", "Failed to generate batch name", err.Error())
		}
		req.Name = newName
	}

	agentID, _ := strconv.ParseInt(req.AgentID, 10, 64)
	var finalCost float64

	// Agent Wallet Logic
	if agentID > 0 {
		finalCost = product.CostPrice * float64(req.Count)
		// Fallback to retail price if cost price is not set
		if finalCost <= 0 && product.Price > 0 {
			finalCost = product.Price * float64(req.Count)
		}

		var wallet domain.AgentWallet
		if err := tx.Set("gorm:query_option", "FOR UPDATE").FirstOrCreate(&wallet, domain.AgentWallet{AgentID: agentID}).Error; err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to lock wallet", err.Error())
		}

		if wallet.Balance < finalCost {
			tx.Rollback()
			return fail(c, http.StatusPaymentRequired, "INSUFFICIENT_FUNDS", "Insufficient wallet balance", nil)
		}

		// Deduct Balance
		newBalance := wallet.Balance - finalCost
		if err := tx.Model(&domain.AgentWallet{}).Where("agent_id = ?", agentID).Updates(map[string]interface{}{"balance": newBalance, "updated_at": time.Now()}).Error; err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "WALLET_UPDATE_FAILED", "Failed to update balance", err.Error())
		}

		// Log Transaction
		log := domain.WalletLog{
			AgentID:     agentID,
			Type:        "purchase",
			Amount:      -finalCost,
			Balance:     newBalance,
			ReferenceID: "batch-" + common.UUID(),
			Remark:      fmt.Sprintf("generated %d vouchers from %s", req.Count, product.Name),
			CreatedAt:   time.Now(),
		}

		// We'll update ReferenceID with BatchID after batch creation, or just use UUID
		if err := tx.Create(&log).Error; err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "LOG_FAILED", "Failed to create transaction log", err.Error())
		}
	}

	batch := domain.VoucherBatch{
		TenantID:       tenantID, // Set tenant from context
		Name:           req.Name,
		ProductID:      productID,
		AgentID:        agentID,
		Count:          req.Count,
		Prefix:         req.Prefix,
		Remark:         req.Remark,
		GeneratePIN:    req.GeneratePIN,
		PINLength:      req.PINLength,
		ExpirationType: req.ExpirationType,
		ValidityDays:   req.ValidityDays,
		CreatedAt:      time.Now(),
	}

	// PrintExpireTime controls when vouchers can no longer be printed/activated
	// This is separate from the actual validity period which comes from Product
	if req.ExpireTime != "" {
		// Try multiple date formats to parse user input (including DD/MM/YYYY)
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04",
			"2006-01-02 15:04",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006/01/02 15:04",
			"2006/01/02T15:04",
			"02/01/2006 15:04", // DD/MM/YYYY HH:MM
			"02/01/2006T15:04", // DD/MM/YYYYTHH:MM
			"02-01-2006 15:04", // DD-MM-YYYY HH:MM
			"02-01-2006T15:04", // DD-MM-YYYYTHH:MM
		}
		var t time.Time
		var err error
		for _, format := range formats {
			t, err = time.Parse(format, req.ExpireTime)
			if err == nil {
				// VALIDATE: Expiry date must be in the future (strictly)
				if t.Before(time.Now()) {
					return fail(c, http.StatusBadRequest, "INVALID_EXPIRY",
						fmt.Sprintf("Voucher batch expiry date must be in the future. Provided date: %s, Current time: %s", t.Format("2006-01-02 15:04"), time.Now().Format("2006-01-02 15:04")), nil)
				}
				batch.PrintExpireTime = &t
				break
			}
		}
		if err != nil {
			// If all formats fail, return clear error instead of silently setting default
			return fail(c, http.StatusBadRequest, "INVALID_DATE_FORMAT",
				"Unable to parse expiry date. Please use format: YYYY-MM-DD HH:MM (e.g., 2026-12-31 23:59)", nil)
		}
	} else {
		// Set default expiry date when field is empty
		defaultExpiry := time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC)
		batch.PrintExpireTime = &defaultExpiry
	}

	if err := tx.Create(&batch).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create batch record", err.Error())
	}

	// Generate Vouchers
	vouchers := make([]domain.Voucher, 0, req.Count)

	var expireTime time.Time

	// Set voucher expiry to match batch expiry
	// Use user's batch expiry date, or default to 31/12/2999 if not set
	if batch.PrintExpireTime != nil && !batch.PrintExpireTime.IsZero() {
		// Use the batch expiry date set by user
		expireTime = *batch.PrintExpireTime
	} else {
		// Default to far-future date if no batch expiry set
		expireTime = time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC)
	}

	// Determine PIN length (default to 4 if not set)
	pinLength := req.PINLength
	if pinLength < 4 {
		pinLength = 4
	}

	for i := 0; i < req.Count; i++ {
		code := req.Prefix + common.GenerateVoucherCode(req.Length, req.Type)

		// Determine time quota: use product.TimeQuota if set, otherwise fallback to ValiditySeconds
		voucherTimeQuota := product.TimeQuota
		if voucherTimeQuota == 0 {
			// Fallback for existing products that don't have TimeQuota set yet
			// This maintains backward compatibility
			voucherTimeQuota = product.ValiditySeconds
		}

		zap.L().Info("CreateVoucherBatch: Voucher TimeQuota calculation",
			zap.Int64("product_id", product.ID),
			zap.Int64("product.TimeQuota", product.TimeQuota),
			zap.Int64("product.ValiditySeconds", product.ValiditySeconds),
			zap.Int64("voucher.TimeQuota (final)", voucherTimeQuota))

		voucher := domain.Voucher{
			TenantID:   tenantID, // Set tenant from context
			BatchID:    batch.ID,
			Code:       code,
			Status:     "unused",
			Price:      product.Price,
			AgentID:    agentID,
			ExpireTime: expireTime,
			RequirePIN: req.GeneratePIN,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),

			// Inherit allocations from Product
			DataQuota: product.DataQuota,
			TimeQuota: voucherTimeQuota, // Time quota from product (with fallback)
		}

		// Generate PIN if required
		if req.GeneratePIN {
			voucher.PIN = common.GeneratePIN(pinLength)
		}

		vouchers = append(vouchers, voucher)
	}

	// Batch Insert
	if err := tx.CreateInBatches(vouchers, 100).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "GENERATE_FAILED", "Failed to insert vouchers", err.Error())
	}

	tx.Commit()

	LogOperation(c, "create_voucher_batch", fmt.Sprintf("Created batch %s with %d vouchers", batch.Name, batch.Count))

	// Prepare response with clear expiry confirmation
	response := map[string]interface{}{
		"id":              batch.ID,
		"name":            batch.Name,
		"count":           batch.Count,
		"expire_time":     batch.PrintExpireTime,
		"expiration_type": batch.ExpirationType,
		"validity_days":   batch.ValidityDays,
		"message":         fmt.Sprintf("Voucher batch created successfully. All vouchers in this batch will expire on %s and cannot be redeemed after this date.", batch.PrintExpireTime.Format("2006-01-02 15:04")),
		"expiry_warning":  fmt.Sprintf("IMPORTANT: No voucher from batch '%s' will be valid after %s", batch.Name, batch.PrintExpireTime.Format("2006-01-02 15:04")),
	}

	return ok(c, response)
}

// VoucherRedeemRequest
type VoucherRedeemRequest struct {
	Code string `json:"code" validate:"required"`
	PIN  string `json:"pin"` // Optional PIN for PIN-protected vouchers
}

// VoucherExtendRequest - add time to voucher
type VoucherExtendRequest struct {
	Code         string `json:"code" validate:"required"`
	ValidityDays int    `json:"validity_days" validate:"required,min=1,max=365"`
}

// VoucherTopupRequest - add data/time to active voucher
type VoucherTopupRequest struct {
	VoucherCode string  `json:"voucher_code" validate:"required"`
	DataQuota   int64   `json:"data_quota"` // Data quota in MB
	TimeQuota   int64   `json:"time_quota"` // Time quota in seconds
	Price       float64 `json:"price"`      // Price paid for topup
}

// VoucherTopupListRequest - list topups for a voucher
type VoucherTopupListRequest struct {
	VoucherCode string `json:"voucher_code" validate:"required"`
}

// VoucherSubscriptionRequest - create subscription for auto-renewal
type VoucherSubscriptionRequest struct {
	VoucherCode  string `json:"voucher_code" validate:"required"`
	ProductID    string `json:"product_id" validate:"required"`
	IntervalDays int    `json:"interval_days" validate:"required,min=1,max=365"`
	AutoRenew    bool   `json:"auto_renew"`
}

// VoucherBundleRequest - create a bundle of vouchers
type VoucherBundleRequest struct {
	Name         string  `json:"name" validate:"required"`
	Description  string  `json:"description"`
	ProductID    string  `json:"product_id" validate:"required"`
	VoucherCount int     `json:"voucher_count" validate:"required,min=1,max=100"`
	Price        float64 `json:"price" validate:"required,min=0"`
}

// RedeemBundleRequest - redeem a bundle and get all voucher codes
type RedeemBundleRequest struct {
	BundleCode string `json:"bundle_code" validate:"required"`
	PIN        string `json:"pin"`
}

// RedeemVoucher activates a voucher and creates a Radius user
// @Summary redeem a voucher
// @Tags Voucher
// @Param redeem body VoucherRedeemRequest true "Redeem info"
// @Success 200 {object} domain.RadiusUser
// @Router /api/v1/vouchers/redeem [post]
func RedeemVoucher(c echo.Context) error {
	var req VoucherRedeemRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	tx := GetDB(c).Begin()

	// 1. Find Voucher
	var voucher domain.Voucher
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("code = ?", req.Code).First(&voucher).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusNotFound, "INVALID_CODE", "Voucher code not found", nil)
	}

	if voucher.Status != "unused" {
		tx.Rollback()
		return fail(c, http.StatusConflict, "VOUCHER_USED", "Voucher already used or expired", nil)
	}

	if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
		tx.Rollback()
		return fail(c, http.StatusConflict, "VOUCHER_EXPIRED", "Voucher has expired", nil)
	}

	// 1a. Validate PIN if required
	if voucher.RequirePIN {
		if req.PIN == "" {
			tx.Rollback()
			return fail(c, http.StatusBadRequest, "PIN_REQUIRED", "PIN is required for this voucher", nil)
		}
		if voucher.PIN != req.PIN {
			tx.Rollback()
			return fail(c, http.StatusUnauthorized, "INVALID_PIN", "Invalid PIN provided", nil)
		}
	}

	// 2. Get Product and Profile
	var product domain.Product
	var batch domain.VoucherBatch
	if err := tx.First(&batch, voucher.BatchID).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "BATCH_ERROR", "Voucher batch not found", err.Error())
	}

	// Check if batch has expired for printing/activation
	if batch.PrintExpireTime != nil && batch.PrintExpireTime.Before(time.Now()) {
		tx.Rollback()
		return fail(c, http.StatusConflict, "BATCH_EXPIRED", "Voucher batch has expired for printing/activation", nil)
	}

	if err := tx.First(&product, batch.ProductID).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "PRODUCT_ERROR", "Associated product not found", err.Error())
	}

	var profile domain.RadiusProfile
	if err := tx.First(&profile, product.RadiusProfileID).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "PROFILE_ERROR", "Associated profile not found", err.Error())
	}

	// 3. Create Radius User
	// User/Pass is the voucher code
	now := time.Now()

	// Calculate user expiration time based on batch expiration type
	var expireTime time.Time

	if batch.ExpirationType == "first_use" {
		// First-use expiration: The time window starts counting from the FIRST RADIUS LOGIN,
		// NOT from redemption time. We set ExpireTime to year 9999 as a placeholder signal
		// so that the first_use_activator plugin will detect this on first auth and
		// calculate the real expiration = first_login_time + ValidityDays (hours).
		// The voucher's ExpireTime (from batch) acts as the absolute deadline for activation.
		expireTime = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	} else {
		// Fixed expiration: Use the voucher's expiry date (set from batch expiry at creation)
		if !voucher.ExpireTime.IsZero() {
			expireTime = voucher.ExpireTime
		} else {
			// Fallback to product validity if voucher has no expiry set
			expireTime = now.Add(time.Duration(product.ValiditySeconds) * time.Second)
		}
	}

	// 519: user := domain.RadiusUser{ ... }
	// We need to decide which rates to use.
	// Use allocations from Voucher (inherited from Product at creation time)
	userUpRate := product.UpRate
	userDownRate := product.DownRate

	// Use DataQuota from Voucher (inherited from Product at batch creation)
	// This ensures quota is tied to the specific voucher, not the product
	userDataQuota := voucher.DataQuota
	userTimeQuota := voucher.TimeQuota

	if userUpRate == 0 {
		userUpRate = profile.UpRate
	}
	if userDownRate == 0 {
		userDownRate = profile.DownRate
	}
	if userDataQuota == 0 {
		userDataQuota = profile.DataQuota
	}
	// Fallback: if voucher.TimeQuota is 0 (created before migration), use product.ValiditySeconds
	if userTimeQuota == 0 {
		userTimeQuota = product.ValiditySeconds
	}

	zap.L().Info("RedeemVoucher: Creating user with TimeQuota",
		zap.String("voucher_code", voucher.Code),
		zap.Int64("voucher.TimeQuota", voucher.TimeQuota),
		zap.Int64("userTimeQuota (before fallback)", userTimeQuota),
		zap.Int64("product.ValiditySeconds", product.ValiditySeconds),
		zap.Int64("userTimeQuota (final)", userTimeQuota))

	user := domain.RadiusUser{
		TenantID:       tenant.GetTenantIDOrDefault(c.Request().Context()), // Set tenant from context
		Username:       voucher.Code,
		Password:       voucher.Code,
		ProfileId:      profile.ID,
		Status:         "enabled",
		ExpireTime:     expireTime,
		CreatedAt:      now,
		UpdatedAt:      now,
		UpRate:         userUpRate,
		DownRate:       userDownRate,
		DataQuota:      userDataQuota,
		AddrPool:       profile.AddrPool,
		TimeQuota:      userTimeQuota,   // Time quota from voucher (inherited from product at batch creation)
		VoucherBatchID: voucher.BatchID, // Link to voucher batch
		VoucherCode:    voucher.Code,    // Link to voucher code
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "USER_CREATE_FAILED", "Failed to create user", err.Error())
	}

	// 4. Update Voucher Status
	// Status "active" means the voucher has been redeemed and is in use
	// This allows tracking of data/time usage against the voucher
	// Only update specific fields - don't overwrite ExpireTime
	if err := tx.Model(&voucher).Updates(map[string]interface{}{
		"status":          "active",
		"radius_username": user.Username,
		"activated_at":    now,
		"first_used_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "VOUCHER_UPDATE_FAILED", "Failed to update voucher status", err.Error())
	}

	tx.Commit()

	LogOperation(c, "redeem_voucher", fmt.Sprintf("Redeemed voucher %s for user %s", voucher.Code, user.Username))

	// Calculate commissions for agent hierarchy
	if batch.AgentID != 0 {
		if err := CalculateCommissions(GetDB(c), batch.AgentID, voucher.ID, voucher.Price); err != nil {
			// Log error but don't fail the redemption
			zap.L().Error("failed to calculate commissions",
				zap.Error(err),
				zap.Int64("voucher_id", voucher.ID),
				zap.Int64("agent_id", batch.AgentID))
		}
	}

	return ok(c, user)
}

// ExtendVoucher adds time to a used voucher by extending the associated user's expiration.
// It retrieves the voucher, validates it's in "used" status, finds the associated RadiusUser,
// extends the user's expiration time by the specified days, and logs the extension.
//
// Parameters:
//   - Code: The voucher code to extend
//   - ValidityDays: Number of days to add (1-365)
//
// Returns:
//   - Success: Extended user object with updated expiration
//   - Error: Voucher not found, not used, or user not found
//
// Authorization: Admin only (check current_operator level)
func ExtendVoucher(c echo.Context) error {
	var req VoucherExtendRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	tx := GetDB(c).Begin()

	var voucher domain.Voucher
	if err := tx.Where("code = ?", req.Code).First(&voucher).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher code not found", nil)
	}

	// 2. Verify voucher status is "active"
	if voucher.Status != "active" {
		tx.Rollback()
		return fail(c, http.StatusConflict, "INVALID_STATUS", "Voucher must be in 'active' status to extend", nil)
	}

	// 3. Find associated RadiusUser
	var user domain.RadiusUser
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("username = ?", voucher.Code).First(&user).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "Associated user not found", err.Error())
	}

	// 4. Calculate new expiration time
	now := time.Now()
	var currentExpire time.Time

	// If user already expired, count from now; otherwise extend from current expiration
	if user.ExpireTime.Before(now) {
		currentExpire = now
	} else {
		currentExpire = user.ExpireTime
	}

	newExpire := currentExpire.Add(time.Duration(req.ValidityDays) * 24 * time.Hour)

	// 5. Update user's ExpireTime
	if err := tx.Model(&domain.RadiusUser{}).Where("id = ?", user.ID).Update("expire_time", newExpire).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to extend user expiration", err.Error())
	}

	// 6. Update voucher extension tracking
	voucher.ExtendedCount++
	voucher.LastExtendedAt = now
	if err := tx.Save(&voucher).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "VOUCHER_UPDATE_FAILED", "Failed to update voucher extension count", err.Error())
	}

	tx.Commit()

	LogOperation(c, "extend_voucher", fmt.Sprintf("Extended voucher %s by %d days", voucher.Code, req.ValidityDays))

	// Return updated user
	user.ExpireTime = newExpire
	return ok(c, map[string]interface{}{
		"voucher":    voucher,
		"user":       user,
		"old_expire": currentExpire,
		"new_expire": newExpire,
		"days_added": req.ValidityDays,
	})
}

// BulkActivateVouchers activates all vouchers in a batch
// This creates RadiusUser records automatically so users can authenticate directly
func BulkActivateVouchers(c echo.Context) error {
	id := c.Param("id")
	batchID, _ := strconv.ParseInt(id, 10, 64)

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	db := GetDB(c)
	var batch domain.VoucherBatch
	if err := db.First(&batch, batchID).Error; err != nil {
		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
	}

	var product domain.Product
	if err := db.First(&product, batch.ProductID).Error; err != nil {
		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
	}

	var profile domain.RadiusProfile
	if err := db.First(&profile, product.RadiusProfileID).Error; err != nil {
		return fail(c, http.StatusNotFound, "PROFILE_NOT_FOUND", "Profile not found", err.Error())
	}

	now := time.Now()
	zap.L().Debug("bulk activate vouchers",
		zap.Int64("batch_id", batchID),
		zap.String("expiration_type", batch.ExpirationType),
		zap.Int64("product_id", batch.ProductID),
		zap.Int64("product_validity_seconds", product.ValiditySeconds),
		zap.Time("batch_print_expire_time", func() time.Time {
			if batch.PrintExpireTime != nil {
				return *batch.PrintExpireTime
			}
			return time.Time{}
		}()),
	)

	updates := map[string]interface{}{
		"status":       "active",
		"activated_at": now,
	}

	expireTime := now.AddDate(1, 0, 0)

	// Calculate expiration for fixed type
	if batch.ExpirationType != "first_use" {
		if batch.PrintExpireTime != nil && !batch.PrintExpireTime.IsZero() {
			expireTime = *batch.PrintExpireTime
		} else if product.ValiditySeconds > 0 {
			expireTime = now.Add(time.Duration(product.ValiditySeconds) * time.Second)
		} else {
			// Default: set expire time to 30 days from now if no validity specified
			expireTime = now.AddDate(0, 0, 30)
		}
		updates["expire_time"] = expireTime
	} else {
		// ensure expire_time is zero for first_use
		updates["expire_time"] = time.Time{}
		// For first_use, set expiry to far future - will be calculated on first login
		expireTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	zap.L().Debug("bulk activate vouchers updates",
		zap.Any("updates", updates),
	)

	// Get all unused vouchers for this batch
	var vouchers []domain.Voucher
	if err := db.Where("batch_id = ? AND status = ?", batchID, "unused").Find(&vouchers).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query vouchers", err.Error())
	}

	if len(vouchers) == 0 {
		return fail(c, http.StatusConflict, "NO_VOUCHERS", "No unused vouchers in this batch", nil)
	}

	tx := db.Begin()

	// Update vouchers status
	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", batchID, "unused").Updates(updates).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to activate vouchers", err.Error())
	}

	// Create RadiusUser for each voucher
	upRate := product.UpRate
	downRate := product.DownRate
	dataQuota := product.DataQuota
	timeQuota := product.TimeQuota
	idleTimeout := product.IdleTimeout
	sessionTimeout := product.SessionTimeout
	if upRate == 0 {
		upRate = profile.UpRate
	}
	if downRate == 0 {
		downRate = profile.DownRate
	}
	if dataQuota == 0 {
		dataQuota = profile.DataQuota
	}
	if timeQuota == 0 {
		timeQuota = product.ValiditySeconds
	}

	for _, voucher := range vouchers {
		user := domain.RadiusUser{
			TenantID:       tenantID,
			Username:       voucher.Code,
			Password:       voucher.Code,
			ProfileId:      profile.ID,
			Status:         "enabled",
			ExpireTime:     expireTime,
			CreatedAt:      now,
			UpdatedAt:      now,
			UpRate:         upRate,
			DownRate:       downRate,
			DataQuota:      dataQuota,
			TimeQuota:      timeQuota,
			IdleTimeout:    idleTimeout,
			SessionTimeout: sessionTimeout,
			AddrPool:       profile.AddrPool,
			VoucherBatchID: batchID,
			VoucherCode:    voucher.Code,
		}

		if err := tx.Create(&user).Error; err != nil {
			zap.L().Error("Failed to create RadiusUser for voucher",
				zap.String("voucher_code", voucher.Code),
				zap.Error(err))
		}
	}

	// Also update the batch activation timestamp
	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", batchID).Update("activated_at", &now).Error; err != nil {
		zap.L().Error("Failed to update batch activated_at", zap.Int64("batch_id", batchID), zap.Error(err))
	}

	if err := tx.Commit().Error; err != nil {
		return fail(c, http.StatusInternalServerError, "COMMIT_FAILED", "Failed to commit transaction", err.Error())
	}

	// Add batch to active cache
	if cache := checkers.GetVoucherBatchCache(); cache != nil {
		cache.AddBatch(batchID)
	}

	LogOperation(c, "bulk_activate_vouchers", fmt.Sprintf("Activated batch %d with %d vouchers", batchID, len(vouchers)))

	return ok(c, map[string]interface{}{
		"activated_count": len(vouchers),
		"batch_id":        batchID,
		"expire_time":     expireTime,
	})
}

// BulkDeactivateVouchers deactivates all active vouchers in a batch
// Supports graceful disconnect with grace period notification
func BulkDeactivateVouchers(c echo.Context) error {
	id := c.Param("id")
	batchID, _ := strconv.ParseInt(id, 10, 64)

	// Parse grace period from query params (default 5 minutes)
	gracePeriodStr := c.QueryParam("grace_period")
	graceDuration := 5 * time.Minute
	if gracePeriodStr != "" {
		if gp, err := strconv.Atoi(gracePeriodStr); err == nil && gp >= 0 {
			graceDuration = time.Duration(gp) * time.Second
		}
	}

	// Parse reason from query params
	reason := c.QueryParam("reason")
	if reason == "" {
		reason = "batch_deactivated_by_admin"
	}

	db := GetDB(c)

	// Find all active vouchers in this batch to get their codes (usernames)
	var vouchers []domain.Voucher
	if err := db.Where("batch_id = ? AND status = ?", id, "active").Find(&vouchers).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query vouchers", err.Error())
	}

	if len(vouchers) == 0 {
		return fail(c, http.StatusConflict, "NO_ACTIVE_VOUCHERS", "No active vouchers in this batch", nil)
	}

	// Find all online users from this batch
	var onlineSessions []domain.RadiusOnline
	if err := db.Joins("JOIN radius_user ON radius_user.username = radius_online.username").
		Where("radius_user.voucher_batch_id = ?", batchID).
		Find(&onlineSessions).Error; err != nil {
		zap.L().Error("Failed to query online sessions for batch",
			zap.Int64("batch_id", batchID),
			zap.Error(err))
	}

	// If grace period > 0, send warning and schedule disconnect
	if graceDuration > 0 && len(onlineSessions) > 0 {
		// Send grace period warning via CoA (Session-Timeout update)
		// Note: This is a best-effort warning, actual disconnect happens after grace period
		go func() {
			// Get a fresh DB connection for the goroutine
			// This is a simplified approach - in production you'd want proper connection management
			sendGracePeriodWarnings(db, batchID, graceDuration, reason)
		}()

		// Schedule actual disconnect after grace period
		go func() {
			time.Sleep(graceDuration)
			disconnectBatchUsersAsync(db, batchID, reason)
		}()

		zap.L().Info("Grace period disconnect scheduled",
			zap.Int64("batch_id", batchID),
			zap.Duration("grace_period", graceDuration),
			zap.Int("online_users", len(onlineSessions)),
			zap.String("reason", reason))

		// Update voucher status immediately but don't disconnect yet
		tx := db.Begin()
		if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", batchID, "active").Update("status", "deactivating").Error; err != nil {
			tx.Rollback()
			return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update voucher status", err.Error())
		}
		tx.Commit()

		// Remove batch from active cache immediately to prevent new connections
		if cache := checkers.GetVoucherBatchCache(); cache != nil {
			cache.RemoveBatch(batchID)
		}

		return ok(c, map[string]interface{}{
			"status":        "grace_period_scheduled",
			"batch_id":      batchID,
			"grace_period":  graceDuration.Seconds(),
			"online_users":  len(onlineSessions),
			"disconnect_at": time.Now().Add(graceDuration).Format(time.RFC3339),
			"message":       "Users will be disconnected after grace period",
		})
	}

	// Immediate disconnect (grace period = 0)
	// Disconnect all online users
	for _, session := range onlineSessions {
		if err := DisconnectSession(c, session); err != nil {
			zap.L().Error("Failed to disconnect voucher user",
				zap.String("username", session.Username),
				zap.Error(err))
		}
	}

	tx := db.Begin()

	// Update voucher status to unused
	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status IN ?", batchID, []string{"active", "deactivating"}).Update("status", "unused").Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to deactivate vouchers", err.Error())
	}

	// Clear the batch activation timestamp
	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", batchID).Update("activated_at", nil).Error; err != nil {
		zap.L().Error("Failed to clear batch activated_at", zap.Int64("batch_id", batchID), zap.Error(err))
	}

	if err := tx.Commit().Error; err != nil {
		return fail(c, http.StatusInternalServerError, "COMMIT_FAILED", "Failed to commit transaction", err.Error())
	}

	// Remove batch from active cache
	if cache := checkers.GetVoucherBatchCache(); cache != nil {
		cache.RemoveBatch(batchID)
	}

	LogOperation(c, "bulk_deactivate_vouchers", fmt.Sprintf("Deactivated batch %d, disconnected %d users", batchID, len(onlineSessions)))

	return ok(c, map[string]interface{}{
		"status":             "deactivated",
		"batch_id":           batchID,
		"deactivated_count":  len(vouchers),
		"disconnected_count": len(onlineSessions),
	})
}

// sendGracePeriodWarnings sends CoA requests to warn users about impending disconnect
func sendGracePeriodWarnings(db *gorm.DB, batchID int64, graceDuration time.Duration, reason string) {
	if db == nil {
		zap.L().Error("Cannot send grace period warnings: database not available")
		return
	}

	graceSeconds := int(graceDuration.Seconds())

	var onlineSessions []domain.RadiusOnline
	if err := db.Joins("JOIN radius_user ON radius_user.username = radius_online.username").
		Where("radius_user.voucher_batch_id = ?", batchID).
		Find(&onlineSessions).Error; err != nil {
		zap.L().Error("Failed to query online sessions for grace period warning",
			zap.Int64("batch_id", batchID),
			zap.Error(err))
		return
	}

	for _, session := range onlineSessions {
		var nas domain.NetNas
		if err := db.Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
			zap.L().Warn("NAS not found for grace period warning",
				zap.String("nas_addr", session.NasAddr))
			continue
		}

		coaPort := nas.CoaPort
		if coaPort <= 0 {
			coaPort = 3799
		}

		vendorCode := nas.VendorCode
		if vendorCode == "" {
			vendorCode = "generic"
		}

		coaReq := coa.CoARequest{
			NASIP:          nas.Ipaddr,
			NASPort:        coaPort,
			Secret:         nas.Secret,
			Username:       session.Username,
			AcctSessionID:  session.AcctSessionId,
			VendorCode:     vendorCode,
			SessionTimeout: graceSeconds,
			Reason:         reason,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client := coa.NewClient(coa.Config{Timeout: 5 * time.Second, RetryCount: 2})
		resp := client.SendCoA(ctx, coaReq)
		cancel()

		if resp.Success {
			zap.L().Info("Grace period warning sent",
				zap.String("username", session.Username),
				zap.Int("grace_seconds", graceSeconds))
		} else {
			zap.L().Warn("Failed to send grace period warning",
				zap.String("username", session.Username),
				zap.Error(resp.Error))
		}
	}
}

// disconnectBatchUsersAsync disconnects all users from a batch after grace period
func disconnectBatchUsersAsync(db *gorm.DB, batchID int64, reason string) {
	if db == nil {
		zap.L().Error("Cannot disconnect batch users: database not available")
		return
	}

	zap.L().Info("Starting batch user disconnect",
		zap.Int64("batch_id", batchID),
		zap.String("reason", reason))

	// Find all online sessions for this batch
	var onlineSessions []domain.RadiusOnline
	if err := db.Joins("JOIN radius_user ON radius_user.username = radius_online.username").
		Where("radius_user.voucher_batch_id = ?", batchID).
		Find(&onlineSessions).Error; err != nil {
		zap.L().Error("Failed to query online sessions for disconnect",
			zap.Int64("batch_id", batchID),
			zap.Error(err))
		return
	}

	disconnectedCount := 0
	for _, session := range onlineSessions {
		var nas domain.NetNas
		if err := db.Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
			zap.L().Warn("NAS not found for disconnect",
				zap.String("nas_addr", session.NasAddr))
			continue
		}

		coaPort := nas.CoaPort
		if coaPort <= 0 {
			coaPort = 3799
		}

		vendorCode := nas.VendorCode
		if vendorCode == "" {
			vendorCode = "generic"
		}

		discReq := coa.DisconnectRequest{
			NASIP:         nas.Ipaddr,
			NASPort:       coaPort,
			Secret:        nas.Secret,
			Username:      session.Username,
			AcctSessionID: session.AcctSessionId,
			VendorCode:    vendorCode,
			Reason:        reason,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client := coa.NewClient(coa.Config{Timeout: 5 * time.Second, RetryCount: 2})
		resp := client.SendDisconnect(ctx, discReq)
		cancel()

		if resp.Success {
			// Delete session from database
			db.Delete(&domain.RadiusOnline{}, "acct_session_id = ?", session.AcctSessionId)
			disconnectedCount++
			zap.L().Info("User disconnected",
				zap.String("username", session.Username))
		} else {
			zap.L().Warn("Failed to disconnect user",
				zap.String("username", session.Username),
				zap.Error(resp.Error))
		}
	}

	// Update voucher status to unused
	db.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", batchID, "deactivating").Update("status", "unused")

	// Clear batch activation timestamp
	db.Model(&domain.VoucherBatch{}).Where("id = ?", batchID).Update("activated_at", nil)

	// Remove from active cache
	if cache := checkers.GetVoucherBatchCache(); cache != nil {
		cache.RemoveBatch(batchID)
	}

	zap.L().Info("Batch disconnect completed",
		zap.Int64("batch_id", batchID),
		zap.Int("disconnected_count", disconnectedCount),
		zap.String("reason", reason))
}

// DeleteVoucherBatch soft deletes a batch and its vouchers
func DeleteVoucherBatch(c echo.Context) error {
	id := c.Param("id")
	tx := GetDB(c).Begin()
	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ?", id).Update("is_deleted", true).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "DELETE_VOUCHERS_FAILED", "Failed to delete vouchers", err.Error())
	}
	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", id).Update("is_deleted", true).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "DELETE_BATCH_FAILED", "Failed to delete batch", err.Error())
	}
	tx.Commit()
	return ok(c, nil)
}

// RestoreVoucherBatch restores a soft-deleted batch
func RestoreVoucherBatch(c echo.Context) error {
	id := c.Param("id")
	tx := GetDB(c).Begin()
	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ?", id).Update("is_deleted", false).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "RESTORE_VOUCHERS_FAILED", "Failed to restore vouchers", err.Error())
	}
	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", id).Update("is_deleted", false).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "RESTORE_BATCH_FAILED", "Failed to restore batch", err.Error())
	}
	tx.Commit()
	return ok(c, nil)
}

// TransferVouchersRequest - transfer vouchers between agents
type TransferVouchersRequest struct {
	BatchID   string `json:"batch_id" validate:"required"`
	ToAgentID string `json:"to_agent_id" validate:"required"`
}

// TransferVouchers transfers vouchers from one agent to another.
// Only unused vouchers can be transferred. Admin can transfer any vouchers,
// while agents can only transfer their own vouchers.
//
// Parameters:
//   - batch_id: ID of the voucher batch to transfer
//   - to_agent_id: ID of the target agent
//
// Returns:
//   - transferred_count: Number of vouchers transferred
func TransferVouchers(c echo.Context) error {
	var req TransferVouchersRequest
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

	batchID, _ := strconv.ParseInt(req.BatchID, 10, 64)
	toAgentID, _ := strconv.ParseInt(req.ToAgentID, 10, 64)

	db := GetDB(c)

	// Verify target agent exists
	var targetAgent domain.SysOpr
	if err := db.First(&targetAgent, toAgentID).Error; err != nil {
		return fail(c, http.StatusNotFound, "AGENT_NOT_FOUND", "Target agent not found", nil)
	}

	// Get the batch
	var batch domain.VoucherBatch
	if err := db.First(&batch, batchID).Error; err != nil {
		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Voucher batch not found", nil)
	}

	// Authorization check
	if currentUser.Level == "agent" {
		// Agents can only transfer their own vouchers
		if batch.AgentID != currentUser.ID {
			return fail(c, http.StatusForbidden, "FORBIDDEN", "You can only transfer your own vouchers", nil)
		}
	}

	// Check if transfer is to the same agent
	if batch.AgentID == toAgentID {
		return fail(c, http.StatusBadRequest, "SAME_AGENT", "Source and target agents are the same", nil)
	}

	tx := db.Begin()

	// Count unused vouchers to transfer
	var count int64
	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ? AND is_deleted = ?", batchID, "unused", false).Count(&count).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count vouchers", err.Error())
	}

	if count == 0 {
		tx.Rollback()
		return fail(c, http.StatusConflict, "NO_VOUCHERS", "No unused vouchers to transfer", nil)
	}

	// Update vouchers to new agent
	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ? AND is_deleted = ?", batchID, "unused", false).Updates(map[string]interface{}{"agent_id": toAgentID}).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "TRANSFER_FAILED", "Failed to transfer vouchers", err.Error())
	}

	// Update batch agent_id
	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", batchID).Update("agent_id", toAgentID).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "BATCH_UPDATE_FAILED", "Failed to update batch agent", err.Error())
	}

	tx.Commit()

	zap.L().Info("Vouchers transferred",
		zap.Int64("batch_id", batchID),
		zap.Int64("from_agent", batch.AgentID),
		zap.Int64("to_agent", toAgentID),
		zap.Int64("count", count))

	return ok(c, map[string]interface{}{
		"transferred_count": count,
		"batch_id":          batchID,
		"from_agent":        batch.AgentID,
		"to_agent":          toAgentID,
	})
}

// RefundUnusedVouchers refunds unused vouchers to agent wallet
func RefundUnusedVouchers(c echo.Context) error {
	id := c.Param("id")
	db := GetDB(c)
	zap.L().Info("Refund attempt", zap.String("batch_id", id))

	var batch domain.VoucherBatch
	if err := db.First(&batch, id).Error; err != nil {
		zap.L().Error("Refund failed: batch not found", zap.String("batch_id", id), zap.Error(err))
		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
	}

	if batch.AgentID <= 0 {
		zap.L().Warn("Refund failed: no agent linked", zap.String("batch_id", id), zap.Int64("agent_id", batch.AgentID))
		return fail(c, http.StatusBadRequest, "NO_AGENT", "This batch was not generated by an agent", nil)
	}

	tx := db.Begin()

	var vouchers []domain.Voucher
	if err := tx.Where("batch_id = ? AND status = ?", id, "unused").Find(&vouchers).Error; err != nil {
		zap.L().Error("Refund failed: voucher query error", zap.String("batch_id", id), zap.Error(err))
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to find unused vouchers", err.Error())
	}

	if len(vouchers) == 0 {
		zap.L().Warn("Refund failed: no unused vouchers found", zap.String("batch_id", id))
		tx.Rollback()
		return fail(c, http.StatusBadRequest, "NO_UNUSED", "No unused vouchers to refund", nil)
	}

	var totalRefund float64
	// Fetch product to get the cost price (same logic as creation)
	var product domain.Product
	if err := tx.First(&product, batch.ProductID).Error; err != nil {
		zap.L().Error("Refund failed: product not found", zap.String("batch_id", id), zap.Int64("product_id", batch.ProductID), zap.Error(err))
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "PRODUCT_ERROR", "Associated product not found", err.Error())
	}

	unitCost := product.CostPrice
	if unitCost <= 0 && product.Price > 0 {
		unitCost = product.Price
	}

	totalRefund = unitCost * float64(len(vouchers))

	// 1. Mark vouchers as deleted/refunded
	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", id, "unused").Updates(map[string]interface{}{
		"status":     "refunded",
		"is_deleted": true,
	}).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update voucher status", err.Error())
	}

	// 2. Refund Wallet
	var wallet domain.AgentWallet
	if err := tx.Set("gorm:query_option", "FOR UPDATE").FirstOrCreate(&wallet, domain.AgentWallet{AgentID: batch.AgentID}).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to lock wallet", err.Error())
	}

	newBalance := wallet.Balance + totalRefund
	if err := tx.Model(&domain.AgentWallet{}).Where("agent_id = ?", batch.AgentID).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "REFUND_FAILED", "Failed to update wallet balance", err.Error())
	}

	// 3. Log Transaction
	log := domain.WalletLog{
		AgentID:     batch.AgentID,
		Type:        "refund",
		Amount:      totalRefund,
		Balance:     newBalance,
		ReferenceID: fmt.Sprintf("refund-%d", batch.ID),
		Remark:      fmt.Sprintf("refunded %d unused vouchers (unit cost: %.2f) from batch %s", len(vouchers), unitCost, batch.Name),
		CreatedAt:   time.Now(),
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "LOG_FAILED", "Failed to create log", err.Error())
	}

	tx.Commit()

	LogOperation(c, "refund_vouchers", fmt.Sprintf("Refunded %d vouchers from batch %s", len(vouchers), batch.Name))

	return ok(c, map[string]interface{}{"refunded_count": len(vouchers), "refund_amount": totalRefund})
}

// ExportVoucherBatch exports batch vouchers to CSV
// ExportVoucherBatch exports batch vouchers to CSV
func ExportVoucherBatch(c echo.Context) error {
	id := c.Param("id")

	var batch domain.VoucherBatch
	if err := GetDB(c).First(&batch, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
	}

	var product domain.Product
	if err := GetDB(c).First(&product, batch.ProductID).Error; err != nil {
		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
	}

	var vouchers []domain.Voucher
	if err := GetDB(c).Where("batch_id = ?", id).Find(&vouchers).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to find vouchers", err.Error())
	}

	nowStr := time.Now().Format("02012006")
	filename := fmt.Sprintf("%s-%s-%d-%s.csv", batch.Name, product.Name, batch.Count, nowStr)
	// Sanitize filename
	filename = strings.ReplaceAll(filename, " ", "_")

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().WriteHeader(http.StatusOK)

	writer := csv.NewWriter(c.Response())
	defer writer.Flush()

	header := []string{
		"batch_id", "code", "radius_user", "status", "agent_id", "price",
		"activated_at", "expire_time", "extended_count", "last_extended_at",
		"is_deleted", "created_at", "updated_at",
	}

	if err := writer.Write(header); err != nil {
		zap.L().Error("Failed to write CSV header", zap.Error(err))
		return fail(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to write CSV header", err.Error())
	}

	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	}

	for _, v := range vouchers {
		row := []string{
			fmt.Sprintf("%d", v.BatchID),
			v.Code,
			v.RadiusUsername,
			v.Status,
			fmt.Sprintf("%d", v.AgentID),
			fmt.Sprintf("%.2f", v.Price),
			formatTime(v.ActivatedAt),
			formatTime(v.ExpireTime),
			fmt.Sprintf("%d", v.ExtendedCount),
			formatTime(v.LastExtendedAt),
			strconv.FormatBool(v.IsDeleted),
			formatTime(v.CreatedAt),
			formatTime(v.UpdatedAt),
		}
		if err := writer.Write(row); err != nil {
			zap.L().Error("Failed to write CSV row", zap.Int64("voucher_id", v.ID), zap.Error(err))
			return fail(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to write CSV row", err.Error())
		}
	}
	return nil
}

// PrintVoucherBatch returns all vouchers for a batch without pagination limits
// Used specifically for printing large batches
// @Summary print voucher batch
// @Tags Voucher
// @Param id path int true "Batch ID"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/voucher-batches/{id}/print [get]
func PrintVoucherBatch(c echo.Context) error {
	id := c.Param("id")

	var batch domain.VoucherBatch
	if err := GetDB(c).First(&batch, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
	}

	// Permission check: Agents can only print their own batches
	currentUser, err := resolveOperatorFromContext(c)
	if err == nil && currentUser.Level == "agent" {
		if batch.AgentID != currentUser.ID {
			return fail(c, http.StatusForbidden, "FORBIDDEN", "Cannot print batch not owned by you", nil)
		}
	}

	var vouchers []domain.Voucher
	if err := GetDB(c).Where("batch_id = ? AND is_deleted = ?", id, false).Order("id ASC").Find(&vouchers).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "PRINT_FAILED", "Failed to fetch vouchers", err.Error())
	}

	result := make([]map[string]interface{}, len(vouchers))
	for i, v := range vouchers {
		result[i] = map[string]interface{}{
			"id":           v.ID,
			"code":         v.Code,
			"status":       v.Status,
			"price":        v.Price,
			"expire_time":  v.ExpireTime,
			"activated_at": v.ActivatedAt,
		}
	}

	return c.JSON(http.StatusOK, result)
}

// CreateVoucherTopup adds data/time quota to an active voucher
// @Summary create voucher topup
// @Tags Voucher
// @Param topup body VoucherTopupRequest true "Topup info"
// @Success 200 {object} domain.VoucherTopup
// @Router /api/v1/vouchers/topup [post]
func CreateVoucherTopup(c echo.Context) error {
	var req VoucherTopupRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	db := GetDB(c)

	// 1. Find the voucher
	var voucher domain.Voucher
	if err := db.Where("code = ?", req.VoucherCode).First(&voucher).Error; err != nil {
		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher not found", err.Error())
	}

	// 2. Validate voucher is active
	if voucher.Status != "active" {
		return fail(c, http.StatusConflict, "VOUCHER_NOT_ACTIVE", "Voucher must be active to add topup", nil)
	}

	// 3. Get current user to check agent permissions
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	// Determine agent ID
	var agentID int64
	if currentUser.Level == "agent" {
		agentID = currentUser.ID
		// Agent can only add topup to their own vouchers
		if voucher.AgentID != agentID {
			return fail(c, http.StatusForbidden, "FORBIDDEN", "Cannot add topup to voucher not owned by you", nil)
		}
	}

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	// 4. Create topup record
	now := time.Now()
	topup := domain.VoucherTopup{
		TenantID:    tenantID, // Set tenant from context
		VoucherID:   voucher.ID,
		VoucherCode: voucher.Code,
		AgentID:     agentID,
		DataQuota:   req.DataQuota,
		TimeQuota:   req.TimeQuota,
		Price:       req.Price,
		Status:      "active",
		ActivatedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 5. If there's a time quota, extend the expiry
	if req.TimeQuota > 0 {
		newExpireTime := voucher.ExpireTime.Add(time.Duration(req.TimeQuota) * time.Second)
		topup.ExpireTime = newExpireTime
		if err := db.Model(&voucher).Update("expire_time", newExpireTime).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to extend voucher expiry", err.Error())
		}
	}

	if err := db.Create(&topup).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create topup", err.Error())
	}

	zap.L().Info("Voucher topup created",
		zap.String("voucher_code", voucher.Code),
		zap.Int64("data_quota", req.DataQuota),
		zap.Int64("time_quota", req.TimeQuota))

	return ok(c, topup)
}

// ListVoucherTopups lists all topups for a voucher
// @Summary list voucher topups
// @Tags Voucher
// @Param voucher_code query string true "Voucher code"
// @Success 200 {object} ListResponse
// @Router /api/v1/vouchers/topups [get]
func ListVoucherTopups(c echo.Context) error {
	voucherCode := c.QueryParam("voucher_code")
	if voucherCode == "" {
		return fail(c, http.StatusBadRequest, "VOUCHER_CODE_REQUIRED", "Voucher code is required", nil)
	}

	db := GetDB(c)
	var topups []domain.VoucherTopup
	if err := db.Where("voucher_code = ?", voucherCode).Order("created_at DESC").Find(&topups).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query topups", err.Error())
	}

	return ok(c, topups)
}

// CreateVoucherSubscription creates a subscription for automatic voucher renewal
// @Summary create voucher subscription
// @Tags Voucher
// @Param subscription body VoucherSubscriptionRequest true "Subscription info"
// @Success 200 {object} domain.VoucherSubscription
// @Router /api/v1/vouchers/subscriptions [post]
func CreateVoucherSubscription(c echo.Context) error {
	var req VoucherSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	db := GetDB(c)

	// 1. Find the voucher
	var voucher domain.Voucher
	if err := db.Where("code = ?", req.VoucherCode).First(&voucher).Error; err != nil {
		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher not found", err.Error())
	}

	// 2. Validate voucher is active
	if voucher.Status != "active" {
		return fail(c, http.StatusConflict, "VOUCHER_NOT_ACTIVE", "Voucher must be active to create subscription", nil)
	}

	// 3. Parse product ID
	productID, err := strconv.ParseInt(req.ProductID, 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_PRODUCT", "Invalid product ID", err.Error())
	}

	// 4. Check product exists
	var product domain.Product
	if err := db.First(&product, productID).Error; err != nil {
		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
	}

	// 5. Get current user
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	var agentID int64
	if currentUser.Level == "agent" {
		agentID = currentUser.ID
		if voucher.AgentID != agentID {
			return fail(c, http.StatusForbidden, "FORBIDDEN", "Cannot create subscription for voucher not owned by you", nil)
		}
	}

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	// 6. Check if subscription already exists
	var existing domain.VoucherSubscription
	if err := db.Where("voucher_code = ? AND status = ?", req.VoucherCode, "active").First(&existing).Error; err == nil {
		return fail(c, http.StatusConflict, "SUBSCRIPTION_EXISTS", "Active subscription already exists for this voucher", nil)
	}

	now := time.Now()
	subscription := domain.VoucherSubscription{
		TenantID:      tenantID, // Set tenant from context
		VoucherCode:   req.VoucherCode,
		ProductID:     productID,
		AgentID:       agentID,
		IntervalDays:  req.IntervalDays,
		Status:        "active",
		AutoRenew:     req.AutoRenew,
		LastRenewalAt: now,
		NextRenewalAt: now.Add(time.Duration(req.IntervalDays) * 24 * time.Hour),
		RenewalCount:  0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := db.Create(&subscription).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create subscription", err.Error())
	}

	zap.L().Info("Voucher subscription created",
		zap.String("voucher_code", req.VoucherCode),
		zap.Int("interval_days", req.IntervalDays),
		zap.Bool("auto_renew", req.AutoRenew))

	return ok(c, subscription)
}

// ListVoucherSubscriptions lists subscriptions for a voucher
// @Summary list voucher subscriptions
// @Tags Voucher
// @Param voucher_code query string true "Voucher code"
// @Success 200 {object} ListResponse
// @Router /api/v1/vouchers/subscriptions [get]
func ListVoucherSubscriptions(c echo.Context) error {
	voucherCode := c.QueryParam("voucher_code")
	if voucherCode == "" {
		return fail(c, http.StatusBadRequest, "VOUCHER_CODE_REQUIRED", "Voucher code is required", nil)
	}

	db := GetDB(c)
	var subscriptions []domain.VoucherSubscription
	if err := db.Where("voucher_code = ?", voucherCode).Order("created_at DESC").Find(&subscriptions).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query subscriptions", err.Error())
	}

	return ok(c, subscriptions)
}

// CancelVoucherSubscription cancels an active subscription
// @Summary cancel voucher subscription
// @Tags Voucher
// @Param id path string true "Subscription ID"
// @Success 200 {object} domain.VoucherSubscription
// @Router /api/v1/vouchers/subscriptions/{id}/cancel [post]
func CancelVoucherSubscription(c echo.Context) error {
	id := c.Param("id")
	subID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID", err.Error())
	}

	db := GetDB(c)

	var subscription domain.VoucherSubscription
	if err := db.Where("id = ? AND status = ?", subID, "active").First(&subscription).Error; err != nil {
		return fail(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Active subscription not found", err.Error())
	}

	// Verify ownership
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	if currentUser.Level == "agent" && subscription.AgentID != currentUser.ID {
		return fail(c, http.StatusForbidden, "FORBIDDEN", "Cannot cancel subscription not owned by you", nil)
	}

	subscription.Status = "cancelled"
	subscription.UpdatedAt = time.Now()

	if err := db.Save(&subscription).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to cancel subscription", err.Error())
	}

	zap.L().Info("Voucher subscription cancelled",
		zap.Int64("subscription_id", subID),
		zap.String("voucher_code", subscription.VoucherCode))

	return ok(c, subscription)
}

// CreateVoucherBundle creates a bundle of vouchers
// @Summary create voucher bundle
// @Tags Voucher
// @Param bundle body VoucherBundleRequest true "Bundle info"
// @Success 200 {object} domain.VoucherBundle
// @Router /api/v1/voucher-bundles [post]
func CreateVoucherBundle(c echo.Context) error {
	var req VoucherBundleRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	db := GetDB(c)

	// Parse product ID
	productID, err := strconv.ParseInt(req.ProductID, 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_PRODUCT", "Invalid product ID", err.Error())
	}

	// Check product exists
	var product domain.Product
	if err := db.First(&product, productID).Error; err != nil {
		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
	}

	// Get current user
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	var agentID int64
	if currentUser.Level == "agent" {
		agentID = currentUser.ID
	}

	// Get tenant ID from context
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())

	now := time.Now()
	// Generate unique bundle code
	bundleCode := "BUNDLE-" + common.UUID()[:8]

	bundle := domain.VoucherBundle{
		TenantID:     tenantID, // Set tenant from context
		Name:         req.Name,
		Description:  req.Description,
		AgentID:      agentID,
		ProductID:    productID,
		VoucherCount: req.VoucherCount,
		Price:        req.Price,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := db.Create(&bundle).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create bundle", err.Error())
	}

	// Generate vouchers for the bundle
	vouchers, err := generateVouchersForBundle(db, product, bundle, req.VoucherCount, agentID, tenantID)
	if err != nil {
		db.Delete(&bundle)
		return fail(c, http.StatusInternalServerError, "GENERATE_FAILED", "Failed to generate vouchers", err.Error())
	}

	zap.L().Info("Voucher bundle created",
		zap.String("name", req.Name),
		zap.Int("voucher_count", req.VoucherCount),
		zap.Float64("price", req.Price))

	return ok(c, map[string]interface{}{
		"bundle":      bundle,
		"bundle_code": bundleCode,
		"vouchers":    vouchers,
	})
}

// generateVouchersForBundle generates vouchers for a bundle
func generateVouchersForBundle(db *gorm.DB, product domain.Product, bundle domain.VoucherBundle, count int, agentID int64, tenantID int64) ([]domain.Voucher, error) {
	vouchers := make([]domain.Voucher, 0, count)

	for i := 0; i < count; i++ {
		code := common.GenerateVoucherCode(12, "mixed")

		voucher := domain.Voucher{
			TenantID:   tenantID,  // Set tenant from context
			BatchID:    bundle.ID, // Link to bundle
			Code:       code,
			Status:     "unused",
			Price:      product.Price,
			AgentID:    agentID,
			ExpireTime: time.Time{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		vouchers = append(vouchers, voucher)
	}

	if err := db.CreateInBatches(vouchers, 100).Error; err != nil {
		return nil, err
	}

	return vouchers, nil
}

// ListVoucherBundles lists all voucher bundles
// @Summary list voucher bundles
// @Tags Voucher
// @Success 200 {object} ListResponse
// @Router /api/v1/voucher-bundles [get]
func ListVoucherBundles(c echo.Context) error {
	db := GetDB(c)

	var bundles []domain.VoucherBundle
	query := db.Model(&domain.VoucherBundle{})

	// Filter by agent
	currentUser, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
	}

	if currentUser.Level == "agent" {
		query = query.Where("agent_id = ?", currentUser.ID)
	}

	if err := query.Order("created_at DESC").Find(&bundles).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query bundles", err.Error())
	}

	return ok(c, bundles)
}

// PublicVoucherStatus returns public voucher status without authentication
// @Summary get public voucher status
// @Tags Voucher
// @Param code query string true "Voucher code"
// @Success 200 {object} map[string]interface{}
// @Router /public/vouchers/status [get]
func PublicVoucherStatus(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return fail(c, http.StatusBadRequest, "CODE_REQUIRED", "Voucher code is required", nil)
	}

	db := GetDB(c)

	var voucher domain.Voucher
	if err := db.Where("code = ?", code).First(&voucher).Error; err != nil {
		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher not found", nil)
	}

	// Return only public information
	return ok(c, map[string]interface{}{
		"status":       voucher.Status,
		"expire_time":  voucher.ExpireTime,
		"activated_at": voucher.ActivatedAt,
	})
}

// isDefaultBatchNamePattern checks if the batch name matches the default pattern
// (e.g., "Batch #" or empty, or any name ending with # without a number)
func isDefaultBatchNamePattern(name string) bool {
	if name == "" {
		return true
	}
	// Match patterns like "Batch #", "الباتش #", "批次 #", or names ending with #
	return name == "Batch #" || name == "الباتش #" || name == "批次 #" || name == "#" ||
		(len(name) >= 2 && name[len(name)-1] == '#' && !isNumeric(name[:len(name)-1]))
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// extractBatchNumber extracts the trailing number from a batch name
// Returns 0 if no number is found
func extractBatchNumber(name string) int {
	if len(name) == 0 {
		return 0
	}
	// Find trailing digits
	var digits string
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] >= '0' && name[i] <= '9' {
			digits = string(name[i]) + digits
		} else {
			break
		}
	}
	if digits == "" {
		return 0
	}
	var num int
	fmt.Sscanf(digits, "%d", &num)
	return num
}

// generateNextBatchName generates the next batch name based on existing batches
// It uses a database lock to prevent race conditions when multiple agents create batches simultaneously
func generateNextBatchName(tx *gorm.DB, suggestedName string) (string, error) {
	// Get the last batch with a numeric suffix, using FOR UPDATE to lock
	var lastBatch domain.VoucherBatch
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Order("id DESC").
		First(&lastBatch).Error

	var nextNumber int
	if err != nil {
		// No batches exist, start from 1
		if err == gorm.ErrRecordNotFound {
			nextNumber = 1
		} else {
			return "", err
		}
	} else {
		// Extract number from last batch name
		nextNumber = extractBatchNumber(lastBatch.Name)
		if nextNumber == 0 {
			// No number in last name, use ID + 1 as fallback
			nextNumber = int(lastBatch.ID) + 1
		} else {
			nextNumber = nextNumber + 1
		}
	}

	// Determine the prefix based on suggested name or use default
	prefix := "Batch #"
	if suggestedName != "" && suggestedName != "Batch #" {
		// Try to extract prefix from suggested name (everything except trailing digits)
		var prefixBuilder string
		for i := len(suggestedName) - 1; i >= 0; i-- {
			if suggestedName[i] >= '0' && suggestedName[i] <= '9' {
				break
			}
			prefixBuilder = string(suggestedName[i]) + prefixBuilder
		}
		if prefixBuilder != "" {
			prefix = prefixBuilder
		}
	}

	return fmt.Sprintf("%s%d", prefix, nextNumber), nil
}

func registerVoucherRoutes() {
	webserver.ApiGET("/voucher-batches", ListVoucherBatches)
	webserver.ApiGET("/vouchers", ListVouchers)
	webserver.ApiGET("/vouchers/check", CheckVoucherUsage)
	webserver.ApiPOST("/voucher-batches", CreateVoucherBatch)
	webserver.ApiPOST("/voucher-batches/:id/activate", BulkActivateVouchers)
	webserver.ApiPOST("/voucher-batches/:id/deactivate", BulkDeactivateVouchers)
	webserver.ApiDELETE("/voucher-batches/:id", DeleteVoucherBatch)
	webserver.ApiPOST("/voucher-batches/:id/restore", RestoreVoucherBatch)
	webserver.ApiPOST("/voucher-batches/:id/refund", RefundUnusedVouchers)
	webserver.ApiGET("/voucher-batches/:id/export", ExportVoucherBatch)
	webserver.ApiGET("/voucher-batches/:id/print", PrintVoucherBatch)
	webserver.ApiPOST("/voucher-batches/:id/transfer", TransferVouchers)
	webserver.ApiPOST("/vouchers/redeem", RedeemVoucher, webserver.RateLimitMiddleware(rate.Limit(5.0/60.0), 5))
	webserver.ApiPOST("/vouchers/extend", ExtendVoucher)
	webserver.ApiPOST("/vouchers/topup", CreateVoucherTopup)
	webserver.ApiGET("/vouchers/topups", ListVoucherTopups)
	webserver.ApiPOST("/vouchers/subscriptions", CreateVoucherSubscription)
	webserver.ApiGET("/vouchers/subscriptions", ListVoucherSubscriptions)
	webserver.ApiPOST("/vouchers/subscriptions/:id/cancel", CancelVoucherSubscription)
	webserver.ApiPOST("/voucher-bundles", CreateVoucherBundle)
	webserver.ApiGET("/voucher-bundles", ListVoucherBundles)
	webserver.ApiPOST("/vouchers/bulk/delete", BulkDeleteVouchers)
	webserver.ApiPOST("/vouchers/bulk/status", BulkUpdateVoucherStatus)

	// Public routes (no authentication required) - use root group
	webserver.GET("/public/vouchers/status", PublicVoucherStatus)
}

func BulkDeleteVouchers(c echo.Context) error {
	var req struct {
		IDs []int64 `json:"ids" validate:"required,min=1"`
	}
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
	}

	if err := GetDB(c).Delete(&domain.Voucher{}, req.IDs).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to delete vouchers", err.Error())
	}

	return ok(c, map[string]interface{}{
		"count":   len(req.IDs),
		"message": "Vouchers deleted successfully",
	})
}

func BulkUpdateVoucherStatus(c echo.Context) error {
	var req struct {
		IDs    []int64 `json:"ids" validate:"required,min=1"`
		Status string  `json:"status" validate:"required,oneof=active used expired disabled"`
	}
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
	}

	if err := GetDB(c).Model(&domain.Voucher{}).Where("id IN ?", req.IDs).Update("status", req.Status).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to update voucher status", err.Error())
	}

	return ok(c, map[string]interface{}{
		"count":   len(req.IDs),
		"message": "Vouchers updated successfully",
	})
}

// CheckVoucherUsage retrieves the usage summary for a specific voucher
// @Summary check voucher usage summary and session history
// @Tags Voucher
// @Param code query string true "Voucher code"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/vouchers/check [get]
func CheckVoucherUsage(c echo.Context) error {
	tenantID := tenant.GetTenantIDOrDefault(c.Request().Context())
	if tenantID == 0 {
		return fail(c, http.StatusBadRequest, "NO_TENANT", "Missing tenant context", nil)
	}

	code := c.QueryParam("code")
	if code == "" {
		return fail(c, http.StatusBadRequest, "INVALID_CODE", "Voucher code is required", nil)
	}

	db := GetDB(c)

	var user domain.RadiusUser
	if err := db.Where("username = ? AND tenant_id = ?", code, tenantID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fail(c, http.StatusNotFound, "NOT_FOUND", "Voucher not found", nil)
		}
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query voucher", err.Error())
	}

	var sessions []domain.RadiusAccounting
	if err := db.Where("username = ? AND tenant_id = ?", code, tenantID).Order("acct_start_time DESC").Find(&sessions).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query accounting records", err.Error())
	}

	var totalTime int64
	var totalData int64
	for _, s := range sessions {
		totalTime += int64(s.AcctSessionTime)
		totalData += s.AcctInputTotal + s.AcctOutputTotal
	}

	remainingTime := user.TimeQuota - totalTime
	if user.TimeQuota > 0 && remainingTime < 0 {
		remainingTime = 0
	}

	dataQuotaBytes := user.DataQuota * 1024 * 1024
	remainingData := dataQuotaBytes - totalData
	if user.DataQuota > 0 && remainingData < 0 {
		remainingData = 0
	}

	return ok(c, map[string]interface{}{
		"username":        user.Username,
		"status":          user.Status,
		"time_quota":      user.TimeQuota,
		"used_time":       totalTime,
		"remaining_time":  remainingTime,
		"data_quota":      user.DataQuota,
		"used_data":       totalData,
		"remaining_data":  remainingData,
		"idle_timeout":    user.IdleTimeout,
		"session_timeout": user.SessionTimeout,
		"sessions":        sessions,
	})
}
