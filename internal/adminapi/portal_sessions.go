package adminapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"go.uber.org/zap"
)

func registerPortalSessionRoutes() {
	webserver.ApiGET("/portal/usage", GetPortalUsage)
	webserver.ApiGET("/portal/sessions", ListPortalSessions)
	webserver.ApiDELETE("/portal/sessions/:id", TerminatePortalSession)
}

// GetPortalUsage returns usage statistics for the current portal user
func GetPortalUsage(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found", nil)
	}

	var stats struct {
		DataUsed     int64     `json:"data_used"`   // Bytes
		TimeUsed     int       `json:"time_used"`   // Seconds consumed
		TimeQuota    int64     `json:"time_quota"`  // Seconds allocated (from product)
		DataQuota    int64     `json:"data_quota"`  // MB
		ExpireTime   time.Time `json:"expire_time"` // Validity window end date
		Status       string    `json:"status"`
		Username     string    `json:"username"`
		MonthlyFee   float64   `json:"monthly_fee"`
		NextBillDate time.Time `json:"next_bill_date"`
		OnlineCount  int       `json:"online_count"`
		MacAddr      string    `json:"mac_addr"`
		BindMac      int       `json:"bind_mac"`
	}

	db := GetDB(c)

	// Get aggregation from accounting
	var usage struct {
		TotalInput  int64 `gorm:"column:input"`
		TotalOutput int64 `gorm:"column:output"`
		TotalTime   int   `gorm:"column:duration"`
	}
	db.Model(&domain.RadiusAccounting{}).
		Select("SUM(acct_input_total) as input, SUM(acct_output_total) as output, SUM(acct_session_time) as duration").
		Where("username = ?", user.Username).
		Scan(&usage)

	// Get online count
	var onlineCount int64
	db.Model(&domain.RadiusOnline{}).Where("username = ?", user.Username).Count(&onlineCount)

	stats.DataUsed = usage.TotalInput + usage.TotalOutput
	stats.TimeUsed = usage.TotalTime
	stats.TimeQuota = user.TimeQuota // ← ADD THIS: Total time allocated by product
	stats.DataQuota = user.DataQuota
	stats.ExpireTime = user.ExpireTime
	stats.Status = user.Status
	stats.Username = user.Username
	stats.MonthlyFee = user.MonthlyFee
	stats.NextBillDate = user.NextBillingDate
	stats.OnlineCount = int(onlineCount)
	stats.MacAddr = user.MacAddr
	stats.BindMac = user.BindMac

	zap.L().Info("GetPortalUsage: Returning stats",
		zap.String("username", user.Username),
		zap.Int64("user.TimeQuota", user.TimeQuota),
		zap.Int64("stats.TimeQuota", stats.TimeQuota),
		zap.Int("stats.TimeUsed", stats.TimeUsed),
		zap.Time("user.ExpireTime", user.ExpireTime))

	return ok(c, stats)
}

// ListPortalSessions lists active sessions for the current portal user
func ListPortalSessions(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found", nil)
	}

	var sessions []domain.RadiusOnline
	if err := GetDB(c).Where("username = ?", user.Username).Find(&sessions).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to query sessions", err.Error())
	}

	return ok(c, sessions)
}

// TerminatePortalSession allows a user to disconnect their own session
func TerminatePortalSession(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found", nil)
	}

	sessionID := c.Param("id")
	db := GetDB(c)

	var session domain.RadiusOnline
	if err := db.Where("id = ? AND username = ?", sessionID, user.Username).First(&session).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found or access denied", nil)
	}

	// Delete online session record
	if err := db.Delete(&session).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to terminate session record", err.Error())
	}

	// Send CoA Disconnect
	if err := DisconnectSession(c, session); err != nil {
		// Log but don't fail the response
		fmt.Printf("Portal: Failed to disconnect session %v: %v\n", sessionID, err)
	}

	return ok(c, map[string]string{"message": "Session disconnected"})
}
