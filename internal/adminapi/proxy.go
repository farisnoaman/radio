package adminapi

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/repository"
	"github.com/talkincode/toughradius/v9/internal/tenant"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// proxyServerPayload represents proxy server request payload.
type proxyServerPayload struct {
	Name       string `json:"name" validate:"required,max=200"`
	Host       string `json:"host" validate:"required,ip|fqdn"`
	AuthPort   int    `json:"auth_port" validate:"gte=1,lte=65535"`
	AcctPort   int    `json:"acct_port" validate:"gte=1,lte=65535"`
	Secret     string `json:"secret" validate:"required,min=6,max=100"`
	MaxConns   int    `json:"max_conns" validate:"gte=1,lte=1000"`
	TimeoutSec int    `json:"timeout_sec" validate:"gte=1,lte=60"`
	Priority   int    `json:"priority" validate:"gte=1"`
	Remark     string `json:"remark" validate:"max=500"`
}

// proxyRealmPayload represents proxy realm request payload.
type proxyRealmPayload struct {
	Realm         string  `json:"realm" validate:"required,max=255"`
	ProxyServers  []int64 `json:"proxy_servers" validate:"required,min=1"`
	FallbackOrder int     `json:"fallback_order" validate:"gte=1"`
	Remark        string  `json:"remark" validate:"max=500"`
}

// ListProxyServers retrieves all proxy servers.
// @Summary list RADIUS proxy servers
// @Tags RADIUS Proxy
// @Success 200 {object} ListResponse
// @Router /api/v1/radius-proxy/servers [get]
func ListProxyServers(c echo.Context) error {
	db := GetDB(c)
	repo := repository.NewProxyRepository(db)

	servers, err := repo.ListServers(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch servers", err.Error())
	}

	return ok(c, servers)
}

// CreateProxyServer creates a new proxy server.
// @Summary create RADIUS proxy server
// @Tags RADIUS Proxy
// @Param server body proxyServerPayload true "Server data"
// @Success 201 {object} domain.RadiusProxyServer
// @Router /api/v1/radius-proxy/servers [post]
func CreateProxyServer(c echo.Context) error {
	var payload proxyServerPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	server := &domain.RadiusProxyServer{
		Name:       payload.Name,
		Host:       payload.Host,
		AuthPort:   payload.AuthPort,
		AcctPort:   payload.AcctPort,
		Secret:     payload.Secret,
		MaxConns:   payload.MaxConns,
		TimeoutSec: payload.TimeoutSec,
		Priority:   payload.Priority,
		Remark:     payload.Remark,
		Status:     "enabled",
	}

	// Set defaults
	if server.AuthPort == 0 {
		server.AuthPort = 1812
	}
	if server.AcctPort == 0 {
		server.AcctPort = 1813
	}
	if server.MaxConns == 0 {
		server.MaxConns = 50
	}
	if server.TimeoutSec == 0 {
		server.TimeoutSec = 5
	}
	if server.Priority == 0 {
		server.Priority = 1
	}

	if err := server.Validate(); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid server configuration", err.Error())
	}

	db := GetDB(c)
	repo := repository.NewProxyRepository(db)

	if err := repo.CreateServer(c.Request().Context(), server); err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create server", err.Error())
	}

	return ok(c, server)
}

// ListProxyRealms retrieves all proxy realms.
// @Summary list RADIUS proxy realms
// @Tags RADIUS Proxy
// @Success 200 {object} ListResponse
// @Router /api/v1/radius-proxy/realms [get]
func ListProxyRealms(c echo.Context) error {
	db := GetDB(c)
	repo := repository.NewProxyRepository(db)

	crealms, err := repo.ListRealms(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch realms", err.Error())
	}

	return ok(c, crealms)
}

// CreateProxyRealm creates a new proxy realm.
// @Summary create RADIUS proxy realm
// @Tags RADIUS Proxy
// @Param realm body proxyRealmPayload true "Realm data"
// @Success 201 {object} domain.RadiusProxyRealm
// @Router /api/v1/radius-proxy/realms [post]
func CreateProxyRealm(c echo.Context) error {
	var payload proxyRealmPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	realm := &domain.RadiusProxyRealm{
		Realm:         payload.Realm,
		ProxyServers:  payload.ProxyServers,
		FallbackOrder: payload.FallbackOrder,
		Remark:        payload.Remark,
		Status:        "enabled",
	}

	// Set default
	if realm.FallbackOrder == 0 {
		realm.FallbackOrder = 1
	}

	if err := realm.Validate(); err != nil {
		return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid realm configuration", err.Error())
	}

	db := GetDB(c)
	repo := repository.NewProxyRepository(db)

	if err := repo.CreateRealm(c.Request().Context(), realm); err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create realm", err.Error())
	}

	return ok(c, realm)
}

// GetProxyLogs retrieves proxy request logs.
// @Summary get proxy request logs
// @Tags RADIUS Proxy
// @Param realm query string false "Filter by realm"
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Success 200 {object} ListResponse
// @Router /api/v1/radius-proxy/logs [get]
func GetProxyLogs(c echo.Context) error {
	db := GetDB(c)
	tenantID, err := tenant.FromContext(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant context required", nil)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	realm := c.QueryParam("realm")

	var total int64
	var logs []domain.ProxyRequestLog

	query := db.Model(&domain.ProxyRequestLog{}).Where("tenant_id = ?", tenantID)

	if realm != "" {
		query = query.Where("realm = ?", realm)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&logs)

	return ok(c, map[string]interface{}{
		"data":  logs,
		"total": total,
	})
}

// registerProxyRoutes registers RADIUS proxy routes.
func registerProxyRoutes() {
	webserver.ApiGET("/radius-proxy/servers", ListProxyServers)
	webserver.ApiPOST("/radius-proxy/servers", CreateProxyServer)
	webserver.ApiGET("/radius-proxy/realms", ListProxyRealms)
	webserver.ApiPOST("/radius-proxy/realms", CreateProxyRealm)
	webserver.ApiGET("/radius-proxy/logs", GetProxyLogs)
}
