package adminapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/webserver"

)

func registerTunnelRoutes() {
	webserver.ApiGET("/system/tunnel/status", GetTunnelStatus)
	webserver.ApiPOST("/system/tunnel/start", StartTunnel)
	webserver.ApiPOST("/system/tunnel/stop", StopTunnel)
	webserver.ApiPOST("/system/tunnel/config", UpdateTunnelConfig)
}

func GetTunnelStatus(c echo.Context) error {
	status, err := GetAppContext(c).TunnelMgr().GetStatus()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TUNNEL_STATUS_ERROR", "Failed to get tunnel status", err.Error())
	}
	return ok(c, status)
}

func StartTunnel(c echo.Context) error {
	err := GetAppContext(c).TunnelMgr().StartTunnel()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TUNNEL_START_ERROR", "Failed to start tunnel", err.Error())
	}
	return ok(c, map[string]string{
		"message": "Tunnel started successfully",
	})
}

func StopTunnel(c echo.Context) error {
	err := GetAppContext(c).TunnelMgr().StopTunnel()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TUNNEL_STOP_ERROR", "Failed to stop tunnel", err.Error())
	}
	return ok(c, map[string]string{
		"message": "Tunnel stopped successfully",
	})
}

func UpdateTunnelConfig(c echo.Context) error {
	var cfg config.TunnelConfig
	if err := c.Bind(&cfg); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid config", err.Error())
	}

	appCtx := GetAppContext(c)
	err := appCtx.TunnelMgr().UpdateConfig(cfg)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TUNNEL_CONFIG_ERROR", "Failed to update config", err.Error())
	}

	// Persist config if possible, but for now just update in-memory
	// Ideally we'd update toughradius.yml or settings DB table

	return ok(c, map[string]string{
		"message": "Tunnel config updated successfully",
	})
}
