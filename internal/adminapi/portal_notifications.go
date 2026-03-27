package adminapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

type NotificationPreferencesRequest struct {
	EmailEnabled        bool   `json:"email_enabled"`
	SMSEnabled         bool   `json:"sms_enabled"`
	EmailThresholds    string `json:"email_thresholds"`
	SMSThresholds      string `json:"sms_thresholds"`
	DailySummaryEnabled bool   `json:"daily_summary_enabled"`
}

func registerPortalNotificationRoutes() {
	webserver.ApiGET("/portal/preferences/notifications", getNotificationPreferencesHandler)
	webserver.ApiPUT("/portal/preferences/notifications", updateNotificationPreferencesHandler)
	webserver.ApiGET("/portal/alerts/history", getAlertHistoryHandler)
}

func getNotificationPreferencesHandler(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), nil)
	}

	var pref domain.NotificationPreference
	err = GetDB(c).Where("user_id = ?", user.ID).First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		pref = domain.NotificationPreference{
			UserID:              user.ID,
			EmailEnabled:        true,
			SMSEnabled:          false,
			EmailThresholds:     "80,90,100",
			SMSThresholds:       "100",
			DailySummaryEnabled: false,
		}
		if err := GetDB(c).Create(&pref).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create preferences", err.Error())
		}
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch preferences", err.Error())
	}

	return ok(c, pref)
}

func updateNotificationPreferencesHandler(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), nil)
	}

	var req NotificationPreferencesRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if req.EmailThresholds == "" {
		req.EmailThresholds = "80,90,100"
	}
	if req.SMSThresholds == "" {
		req.SMSThresholds = "100"
	}

	var pref domain.NotificationPreference
	err = GetDB(c).Where("user_id = ?", user.ID).First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		pref = domain.NotificationPreference{
			UserID:              user.ID,
			EmailEnabled:        req.EmailEnabled,
			SMSEnabled:          req.SMSEnabled,
			EmailThresholds:     req.EmailThresholds,
			SMSThresholds:       req.SMSThresholds,
			DailySummaryEnabled: req.DailySummaryEnabled,
		}
		if err := GetDB(c).Create(&pref).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create preferences", err.Error())
		}
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch preferences", err.Error())
	} else {
		pref.EmailEnabled = req.EmailEnabled
		pref.SMSEnabled = req.SMSEnabled
		pref.EmailThresholds = req.EmailThresholds
		pref.SMSThresholds = req.SMSThresholds
		pref.DailySummaryEnabled = req.DailySummaryEnabled
		if err := GetDB(c).Save(&pref).Error; err != nil {
			return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update preferences", err.Error())
		}
	}

	return ok(c, pref)
}

func getAlertHistoryHandler(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), nil)
	}

	var alerts []domain.UsageAlert
	err = GetDB(c).Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(50).
		Find(&alerts).Error

	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch alerts", err.Error())
	}

	return ok(c, alerts)
}
