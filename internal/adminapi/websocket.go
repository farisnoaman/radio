package adminapi

import (
	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/app/websocket"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerWebsocketRoutes() {
	webserver.GET("/api/v1/dashboard/ws", DashboardWebSocket)
}

func DashboardWebSocket(c echo.Context) error {
	hub := GetAppContext(c).WsHub()
	websocket.ServeWs(hub, c.Response(), c.Request())
	return nil
}
