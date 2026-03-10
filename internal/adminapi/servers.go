package adminapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// serverPayload represents the Server creation request structure
type serverPayload struct {
	Name          string `json:"name" validate:"required,min=1,max=255"`
	PublicIP      string `json:"public_ip" validate:"omitempty,max=64"`
	Secret        string `json:"secret" validate:"omitempty,max=128"`
	Username      string `json:"username" validate:"omitempty,max=128"`
	Password      string `json:"password" validate:"omitempty,max=128"`
	Ports         string `json:"ports" validate:"omitempty,max=128"`
	RouterLimit   string `json:"router_limit" validate:"omitempty,max=255"`
	DBHost        string `json:"db_host" validate:"omitempty,max=128"`
	DBPort        int    `json:"db_port"`
	DBName        string `json:"db_name" validate:"omitempty,max=128"`
	DBUsername    string `json:"db_username" validate:"omitempty,max=128"`
	DBPassword    string `json:"db_password" validate:"omitempty,max=128"`
	RouterStatus  string `json:"router_status" validate:"omitempty,max=32"`
	OnlineHotspot int    `json:"online_hotspot"`
	OnlinePPPoE   int    `json:"online_pppoe"`
}

// ListServers retrieves the server list
// @Summary get the server list
// @Tags Servers
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Success 200 {object} ListResponse
// @Router /api/v1/network/servers [get]
func ListServers(c echo.Context) error {
	db := GetDB(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
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
	var servers []domain.Server

	query := db.Model(&domain.Server{})

	// Filter by name (case-insensitive)
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		if strings.EqualFold(db.Name(), "postgres") { //nolint:staticcheck
			query = query.Where("name ILIKE ?", "%"+name+"%")
		} else {
			query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
		}
	}

	// Filter by public_ip
	if publicIP := strings.TrimSpace(c.QueryParam("public_ip")); publicIP != "" {
		query = query.Where("public_ip LIKE ?", publicIP+"%")
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&servers)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  servers,
		"total": total,
	})
}

// GetServer fetches a single server
// @Summary get server detail
// @Tags Servers
// @Param id path int true "Server ID"
// @Success 200 {object} domain.Server
// @Router /api/v1/network/servers/{id} [get]
func GetServer(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Server ID", nil)
	}

	var server domain.Server
	if err := GetDB(c).First(&server, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Server not found", nil)
	}

	return ok(c, server)
}

// CreateServer creates a server
// @Summary create a server
// @Tags Servers
// @Param server body serverPayload true "Server information"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/servers [post]
func CreateServer(c echo.Context) error {
	var payload serverPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	// Validate the request payload
	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	server := domain.Server{
		Name:          payload.Name,
		PublicIP:      payload.PublicIP,
		Secret:        payload.Secret,
		Username:      payload.Username,
		Password:      payload.Password,
		Ports:         payload.Ports,
		RouterLimit:   payload.RouterLimit,
		DBHost:        payload.DBHost,
		DBPort:        payload.DBPort,
		DBName:        payload.DBName,
		DBUsername:    payload.DBUsername,
		DBPassword:    payload.DBPassword,
		RouterStatus:  payload.RouterStatus,
		OnlineHotspot: payload.OnlineHotspot,
		OnlinePPPoE:   payload.OnlinePPPoE,
	}

	if err := GetDB(c).Create(&server).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create Server", err.Error())
	}

	return ok(c, server)
}

// UpdateServer updates a server
// @Summary update a server
// @Tags Servers
// @Param id path int true "Server ID"
// @Param server body serverPayload true "Server information"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/servers/{id} [put]
func UpdateServer(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Server ID", nil)
	}

	var server domain.Server
	if err := GetDB(c).First(&server, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Server not found", nil)
	}

	var payload serverPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	// Validate the request payload
	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	// Update fields
	server.Name = payload.Name
	server.PublicIP = payload.PublicIP
	server.Secret = payload.Secret
	server.Username = payload.Username
	server.Password = payload.Password
	server.Ports = payload.Ports
	server.RouterLimit = payload.RouterLimit
	server.DBHost = payload.DBHost
	server.DBPort = payload.DBPort
	server.DBName = payload.DBName
	server.DBUsername = payload.DBUsername
	server.DBPassword = payload.DBPassword
	server.RouterStatus = payload.RouterStatus
	server.OnlineHotspot = payload.OnlineHotspot
	server.OnlinePPPoE = payload.OnlinePPPoE

	if err := GetDB(c).Save(&server).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update Server", err.Error())
	}

	return ok(c, server)
}

// DeleteServer deletes a server
// @Summary delete a server
// @Tags Servers
// @Param id path int true "Server ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/servers/{id} [delete]
func DeleteServer(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid Server ID", nil)
	}

	if err := GetDB(c).Delete(&domain.Server{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete Server", err.Error())
	}

	return ok(c, map[string]interface{}{
		"message": "Deletion successful",
	})
}

// registerServerRoutes registers Server routes
func registerServerRoutes() {
	webserver.ApiGET("/network/servers", ListServers)
	webserver.ApiGET("/network/servers/:id", GetServer)
	webserver.ApiPOST("/network/servers", CreateServer)
	webserver.ApiPUT("/network/servers/:id", UpdateServer)
	webserver.ApiDELETE("/network/servers/:id", DeleteServer)
}
