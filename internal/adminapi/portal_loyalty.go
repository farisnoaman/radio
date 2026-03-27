package adminapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"gorm.io/gorm"
)

func registerPortalLoyaltyRoutes() {
	webserver.ApiGET("/portal/loyalty", GetPortalLoyaltyStatus)
	webserver.ApiPOST("/portal/loyalty/redeem", RedeemLoyaltyPoints)
}

func GetPortalLoyaltyStatus(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found", nil)
	}

	appCtx := GetAppContext(c)
	loyaltySvc := appCtx.LoyaltyService()
	db := GetDB(c)

	if loyaltySvc == nil {
		return fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Loyalty service not injected", nil)
	}

	identityKey := loyaltySvc.GenerateIdentityKey(user.MacAddr, user.TenantID)

	var profile domain.LoyaltyProfile
	if err := db.Where("identity_key = ?", identityKey).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Profile not created yet (no usage events processed after the feature was turned on)
			profile = domain.LoyaltyProfile{
				TenantID:    user.TenantID,
				IdentityKey: identityKey,
				MacAddress:  user.MacAddr,
				Badge:       "None",
				Points:      0,
			}
		} else {
			return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch loyalty profile", nil)
		}
	}

	// Fetch active rules to help the frontend display upcoming thresholds
	var rules []domain.LoyaltyRule
	if err := db.Where("tenant_id = ?", user.TenantID).Find(&rules).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch loyalty rules", nil)
	}

	return ok(c, map[string]interface{}{
		"profile": profile,
		"rules":   rules,
	})
}

func RedeemLoyaltyPoints(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found", nil)
	}

	var req struct {
		PointsToRedeem int64 `json:"points" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid parameters", nil)
	}

	if req.PointsToRedeem <= 0 {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Points to redeem must be greater than zero", nil)
	}

	appCtx := GetAppContext(c)
	loyaltySvc := appCtx.LoyaltyService()
	db := GetDB(c)

	identityKey := loyaltySvc.GenerateIdentityKey(user.MacAddr, user.TenantID)

	err = db.Transaction(func(tx *gorm.DB) error {
		var profile domain.LoyaltyProfile
		if err := tx.Where("identity_key = ?", identityKey).First(&profile).Error; err != nil {
			return err
		}

		if profile.Points < req.PointsToRedeem {
			return errors.New("INSUFFICIENT_POINTS")
		}

		// --- Fraud Prevention: daily rate limiting ---
		const maxDailyRedemptions = 3
		now := time.Now()
		if profile.LastRedeemAt != nil {
			// Check if last redemption was today
			lastDate := profile.LastRedeemAt.Truncate(24 * time.Hour)
			today := now.Truncate(24 * time.Hour)
			if lastDate.Equal(today) && profile.DailyRedeemCount >= maxDailyRedemptions {
				return errors.New("DAILY_LIMIT_EXCEEDED")
			}
			// If it was a different day, the background job resets the counter.
			// As a safeguard, reset inline too if it crossed midnight.
			if !lastDate.Equal(today) {
				profile.DailyRedeemCount = 0
			}
		}

		// Deduct points and update fraud prevention counters using optimistic locking
		result := tx.Model(&profile).
			Where("id = ? AND version = ?", profile.ID, profile.Version).
			Updates(map[string]interface{}{
				"points":             gorm.Expr("points - ?", req.PointsToRedeem),
				"version":            gorm.Expr("version + 1"),
				"last_redeem_at":     now,
				"daily_redeem_count": gorm.Expr("daily_redeem_count + 1"),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("CONCURRENCY_ERROR")
		}

		// Record in reward history for audit trail
		tx.Create(&domain.LoyaltyReward{
			ProfileID:  profile.ID,
			RewardType: "DATA_QUOTA",
			PointsCost: req.PointsToRedeem,
			RedeemedAt: now,
		})

		// Apply reward: 10 GB per 20 points
		rewardDataBytes := (req.PointsToRedeem / 20) * 10 * 1024 * 1024 * 1024
		if rewardDataBytes > 0 {
			if err := tx.Model(user).Update("data_quota", gorm.Expr("data_quota + ?", rewardDataBytes)).Error; err != nil {
				return err
			}
		}

		LogOperation(c, "loyalty_points_redeem", "User redeemed loyalty points")
		return nil
	})

	if err != nil {
		switch err.Error() {
		case "INSUFFICIENT_POINTS":
			return fail(c, http.StatusBadRequest, "INSUFFICIENT_POINTS", "Not enough loyalty points available.", nil)
		case "DAILY_LIMIT_EXCEEDED":
			return fail(c, http.StatusTooManyRequests, "DAILY_LIMIT_EXCEEDED", "You have reached the maximum of 3 redemptions per day. Please try again tomorrow.", nil)
		default:
			return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to process redemption", err.Error())
		}
	}

	return ok(c, map[string]interface{}{
		"message": "Successfully redeemed loyalty points.",
	})
}
