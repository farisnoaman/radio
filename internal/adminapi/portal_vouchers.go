package adminapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerPortalVoucherRoutes() {
	webserver.ApiPOST("/portal/vouchers/redeem", RedeemPortalVoucher)
}

// RedeemPortalVoucher allows a logged-in user to apply a voucher to their account
func RedeemPortalVoucher(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found", nil)
	}

	var req struct {
		Code string `json:"code" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid parameters", nil)
	}

	db := GetDB(c)
	tx := db.Begin()

	// 1. Find Voucher
	var voucher domain.Voucher
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("code = ?", req.Code).First(&voucher).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Voucher not found", nil)
	}

	if voucher.Status != "unused" {
		tx.Rollback()
		return fail(c, http.StatusConflict, "ALREADY_USED", "Voucher has already been used", nil)
	}

	if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
		tx.Rollback()
		return fail(c, http.StatusConflict, "EXPIRED", "Voucher has expired", nil)
	}

	// 2. Get Product/Profile
	var batch domain.VoucherBatch
	if err := tx.First(&batch, voucher.BatchID).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query batch", nil)
	}

	var product domain.Product
	if err := tx.First(&product, batch.ProductID).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query product", nil)
	}

	// 3. Apply to User
	// Logic: Extend expiration and add data quota
	now := time.Now()
	
	// Data Quota
	if product.DataQuota > 0 {
		user.DataQuota += product.DataQuota
	}

	// Expiration
	var currentExpire time.Time
	if user.ExpireTime.Before(now) {
		currentExpire = now
	} else {
		currentExpire = user.ExpireTime
	}
	
	if product.ValiditySeconds > 0 {
		user.ExpireTime = currentExpire.Add(time.Duration(product.ValiditySeconds) * time.Second)
	}

	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to update user", err.Error())
	}

	// 4. Update Voucher
	voucher.Status = "active"
	voucher.RadiusUsername = user.Username
	voucher.ActivatedAt = now
	voucher.FirstUsedAt = now

	if err := tx.Save(&voucher).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to update voucher", err.Error())
	}

	tx.Commit()

	LogOperation(c, "portal_redeem_voucher", fmt.Sprintf("User %s redeemed voucher %s", user.Username, voucher.Code))

	return ok(c, map[string]interface{}{
		"message": "Voucher redeemed successfully",
		"expire_time": user.ExpireTime,
		"data_quota": user.DataQuota,
	})
}
