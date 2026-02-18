package adminapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"

)

func registerPrivacyRoutes() {
	webserver.ApiGET("/privacy/user/:username/export", ExportUserData)
	webserver.ApiPOST("/privacy/user/:username/anonymize", AnonymizeUser)
}

// UserDataExport represents the structure of exported user data
type UserDataExport struct {
	User          domain.RadiusUser          `json:"user"`
	Vouchers      []domain.Voucher           `json:"vouchers"`
	Accounting    []domain.RadiusAccounting  `json:"accounting"`
	Subscriptions []domain.VoucherSubscription `json:"subscriptions"`
	ExportedAt    time.Time                  `json:"exported_at"`
}

// ExportUserData exports all data related to a user
// @Summary export user data
// @Tags Privacy
// @Param username path string true "Username"
// @Success 200 {object} UserDataExport
// @Router /api/v1/privacy/user/{username}/export [get]
func ExportUserData(c echo.Context) error {
	username := c.Param("username")
	db := GetDB(c)

	var user domain.RadiusUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", err.Error())
	}

	// Fetch related data
	var vouchers []domain.Voucher
	db.Where("radius_username = ?", username).Find(&vouchers)

	var accounting []domain.RadiusAccounting
	db.Where("username = ?", username).Limit(1000).Order("acct_start_time DESC").Find(&accounting)

	var subscriptions []domain.VoucherSubscription
	db.Where("voucher_code = ?", username).Find(&subscriptions) // Assuming username is voucher code in many cases

	export := UserDataExport{
		User:          user,
		Vouchers:      vouchers,
		Accounting:    accounting,
		Subscriptions: subscriptions,
		ExportedAt:    time.Now(),
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=user_export_%s.json", username))
	return c.JSON(http.StatusOK, export)
}

// AnonymizeUser anonymizes user PII
// @Summary anonymize user data
// @Tags Privacy
// @Param username path string true "Username"
// @Success 200 {object} map[string]string
// @Router /api/v1/privacy/user/{username}/anonymize [post]
func AnonymizeUser(c echo.Context) error {
	username := c.Param("username")
	db := GetDB(c)

	var user domain.RadiusUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", err.Error())
	}

	tx := db.Begin()

	// Anonymize PII fields
	anonName := fmt.Sprintf("anon-%d", user.ID)
	updates := map[string]interface{}{
		"realname": "Anonymized User",
		"mobile":   "",
		"email":    "",
		"address":  "",
		"remark":   "User requested anonymization on " + time.Now().Format(time.RFC3339),
		// We might keep the username if it's the identifier, or change it if it's PII (e.g. mobile number)
		// If username is PII (like mobile), we should change it, but that breaks Radius auth if not handled.
		// Detailed strategy: If username looks like mobile/email, randomize it.
		// For now, we assume username is technical ID, but Realname/Email/Mobile are PII.
	}
	
	// If Username looks like an email or mobile, anonymize it too, but this changes login
	if strings.Contains(username, "@") || (len(username) > 8) {
		updates["username"] = anonName
		updates["password"] = common.UUID() // Reset password
		
		// Also update vouchers linked to this username
		tx.Model(&domain.Voucher{}).Where("radius_username = ?", username).Update("radius_username", anonName)
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to anonymize user", err.Error())
	}

	// Anonymize SysOprLog if any (optional, usually operator logs refer to admin actions, not user data)

	tx.Commit()

	LogOperation(c, "anonymize_user", fmt.Sprintf("Anonymized user %s (ID: %d)", username, user.ID))

	return ok(c, map[string]string{
		"message": "User anonymized successfully",
		"new_username": user.Username, // Return in case it changed
	})
}
