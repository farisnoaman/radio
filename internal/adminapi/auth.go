package adminapi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

const tokenTTL = 12 * time.Hour

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registerAuthRoutes() {
	webserver.ApiPOST("/auth/login", loginHandler)
	webserver.ApiPOST("/auth/portal/login", portalLoginHandler)
	webserver.ApiGET("/auth/me", currentUserHandler)
	webserver.ApiGET("/auth/portal/me", currentUserPortalHandler)
}

func loginHandler(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse login parameters", nil)
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		return fail(c, http.StatusBadRequest, "INVALID_CREDENTIALS", "Username and password cannot be empty", nil)
	}

	var operator domain.SysOpr
	err := GetDB(c).Where("username = ?", req.Username).First(&operator).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect username or password", nil)
	}
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query user", err.Error())
	}

	hashed := common.Sha256HashWithSalt(req.Password, common.GetSecretSalt())
	if hashed != operator.Password {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect username or password", nil)
	}

	token, err := issueToken(c, operator)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TOKEN_ERROR", "Failed to generate login token", nil)
	}

	go func(id int64) {
		GetDB(c).Model(&domain.SysOpr{}).Where("id = ?", id).Update("last_login", time.Now())
	}(operator.ID)

	// Mask password
	operator.Password = ""
	
	// Log login
	LogOperation(c, "operator_login", "Operator "+operator.Username+" logged in")

	return ok(c, map[string]interface{}{
		"token":        token,
		"user":         operator,
		"permissions":  []string{},
		"tokenExpires": time.Now().Add(tokenTTL).Unix(),
	})
}

func portalLoginHandler(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse login parameters", nil)
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		return fail(c, http.StatusBadRequest, "INVALID_CREDENTIALS", "Username and password cannot be empty", nil)
	}

	var user domain.RadiusUser
	err := GetDB(c).Where("username = ?", req.Username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect username or password", nil)
	}
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query user", err.Error())
	}

	// RADIUS users currently use plain text passwords (compatible with PAP)
	if req.Password != user.Password {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect username or password", nil)
	}

	if strings.EqualFold(user.Status, common.DISABLED) {
		return fail(c, http.StatusUnauthorized, "ACCOUNT_DISABLED", "Account has been disabled", nil)
	}

	token, err := issueUserToken(c, user)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TOKEN_ERROR", "Failed to generate login token", nil)
	}

	// Mask password
	user.Password = ""

	return ok(c, map[string]interface{}{
		"token":        token,
		"user":         user,
		"permissions":  []string{},
		"tokenExpires": time.Now().Add(tokenTTL).Unix(),
	})
}

func issueToken(c echo.Context, op domain.SysOpr) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", op.ID),
		"username": op.Username,
		"role":     op.Level,
		"exp":      now.Add(tokenTTL).Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Add(-1 * time.Minute).Unix(),
		"iss":      "toughradius",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetAppContext(c).Config().Web.Secret))
}

func issueUserToken(c echo.Context, user domain.RadiusUser) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", user.ID),
		"username": user.Username,
		"role":     "user",
		"exp":      now.Add(tokenTTL).Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Add(-1 * time.Minute).Unix(),
		"iss":      "toughradius",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetAppContext(c).Config().Web.Secret))
}

func currentUserHandler(c echo.Context) error {
	operator, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), nil)
	}
	return ok(c, map[string]interface{}{
		"user":        operator,
		"permissions": []string{},
	})
}

func currentUserPortalHandler(c echo.Context) error {
	user, err := resolveUserFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), nil)
	}
	return ok(c, map[string]interface{}{
		"user":        user,
		"permissions": []string{"user"},
	})
}

func resolveOperatorFromContext(c echo.Context) (*domain.SysOpr, error) {
	// Check for directly injected operator (for testing)
	if op, ok := c.Get("current_operator").(*domain.SysOpr); ok {
		return op, nil
	}

	userVal := c.Get("user")
	if userVal == nil {
		return nil, errors.New("no user in context")
	}

	token, ok := userVal.(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("invalid token type, got: %T", userVal)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Security: Do not allow RADIUS users (role: "user") to be resolved as Operators/Admins
	role, _ := claims["role"].(string)
	if role == "user" {
		return nil, errors.New("access denied: user is not an operator")
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, errors.New("invalid token subject")
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return nil, errors.New("invalid token id")
	}
	var operator domain.SysOpr
	err = GetDB(c).Where("id = ?", id).First(&operator).Error
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(operator.Status, common.DISABLED) {
		return nil, errors.New("account has been disabled")
	}
	operator.Password = ""
	return &operator, nil
}

func resolveUserFromContext(c echo.Context) (*domain.RadiusUser, error) {
	userVal := c.Get("user")
	if userVal == nil {
		return nil, errors.New("no user in context")
	}

	token, ok := userVal.(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("invalid token type, got: %T", userVal)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	role, _ := claims["role"].(string)
	if role != "user" {
		return nil, errors.New("access denied: account is not a RADIUS user")
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, errors.New("invalid token subject")
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return nil, errors.New("invalid token id")
	}
	var user domain.RadiusUser
	err = GetDB(c).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(user.Status, common.DISABLED) {
		return nil, errors.New("account has been disabled")
	}
	user.Password = ""
	return &user, nil
}

// GetOperator returns the current operator from the request context.
// Returns nil if no operator is authenticated.
func GetOperator(c echo.Context) *domain.SysOpr {
	opr, _ := resolveOperatorFromContext(c)
	return opr
}

// GetOperatorTenantID returns the tenant ID for the current operator.
// Returns 1 (default tenant) if no operator is authenticated.
func GetOperatorTenantID(c echo.Context) int64 {
	opr := GetOperator(c)
	if opr == nil {
		return 1
	}
	if opr.TenantID > 0 {
		return opr.TenantID
	}
	return 1
}

// GetOperatorTenantIDFromContext returns the tenant ID for the current operator from the global context.
// This is a helper for middleware that doesn't have direct access to the Echo context.
func GetOperatorTenantIDFromContext() int64 {
	return 1
}

// IsSuperAdmin checks if the current operator is a super admin.
func IsSuperAdmin(c echo.Context) bool {
	opr := GetOperator(c)
	if opr == nil {
		return false
	}
	return opr.Level == "super"
}

// CanAccessTenant checks if the current operator can access the specified tenant.
func CanAccessTenant(c echo.Context, tenantID int64) bool {
	opr := GetOperator(c)
	if opr == nil {
		return false
	}
	if opr.Level == "super" {
		return true
	}
	return opr.TenantID == tenantID
}
