package adminapi

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/discovery"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// discoveryScanPayload represents the network scan request structure
type discoveryScanPayload struct {
	IPRange string   `json:"ip_range" validate:"required"`
	Ports   []int    `json:"ports" validate:"omitempty"`
	Timeout *int     `json:"timeout" validate:"omitempty,gte=1"`
	Workers *int     `json:"workers" validate:"omitempty,gte=1,max=100"`
}

// discoveryAddPayload represents adding a discovered device to NAS
type discoveryAddPayload struct {
	IP      string `json:"ip" validate:"required,ip"`
	Secret  string `json:"secret" validate:"required,min=6,max=100"`
	Name    string `json:"name" validate:"omitempty,max=100"`
	Model   string `json:"model" validate:"omitempty,max=50"`
	Tags    string `json:"tags" validate:"omitempty,max=200"`
}

// DiscoveryResultResponse represents a discovered device for API response
type DiscoveryResultResponse struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	IsRouterOS bool    `json:"is_router_os"`
	Identity   string  `json:"identity,omitempty"`
	BoardName  string  `json:"board_name,omitempty"`
	Version    string  `json:"version,omitempty"`
	Model      string  `json:"model,omitempty"`
	Serial     string  `json:"serial,omitempty"`
	Timestamp  string  `json:"timestamp"`
	Error      string  `json:"error,omitempty"`
}

// ScanNetwork initiates a network scan to discover MikroTik devices
// @Summary scan network for MikroTik devices
// @Tags Discovery
// @Param payload body discoveryScanPayload true "Scan configuration"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/discovery/scan [post]
func ScanNetwork(c echo.Context) error {
	var payload discoveryScanPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, 400, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	// Validate CIDR format
	if !strings.Contains(payload.IPRange, "/") {
		return fail(c, 400, "INVALID_CIDR", "IPRange must be in CIDR format (e.g., 192.168.1.0/24)", nil)
	}

	// Build scanner config
	config := discovery.Config{
		IPRange: payload.IPRange,
		Ports:   payload.Ports,
	}

	if payload.Timeout != nil {
		config.Timeout = time.Duration(*payload.Timeout) * time.Second
	}

	if payload.Workers != nil {
		config.Workers = *payload.Workers
	}

	scanner, err := discovery.NewScanner(config)
	if err != nil {
		return fail(c, 400, "SCANNER_ERROR", "Failed to create scanner", err.Error())
	}

	// Run scan with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result := scanner.Scan(ctx)

	// Convert results to response format
	results := make([]DiscoveryResultResponse, 0, len(result.Results))
	for _, r := range result.Results {
		resp := DiscoveryResultResponse{
			IP:         r.IP,
			Port:       r.Port,
			IsRouterOS: r.IsRouterOS,
			Timestamp:  r.Timestamp.Format(time.RFC3339),
		}

		if r.DeviceInfo != nil {
			resp.Identity = r.DeviceInfo.Identity
			resp.BoardName = r.DeviceInfo.BoardName
			resp.Version = r.DeviceInfo.Version
			resp.Model = r.DeviceInfo.Model
			resp.Serial = r.DeviceInfo.Serial
		}

		if r.Error != "" {
			resp.Error = r.Error
		}

		results = append(results, resp)
	}

	return ok(c, map[string]interface{}{
		"cidr":        result.CIDR,
		"duration":    result.Duration.Seconds(),
		"found_count": result.FoundCount,
		"total_hosts": len(result.Results),
		"results":     results,
	})
}

// GetDiscoveryResult retrieves a previous scan result by ID
// @Summary get scan result by ID
// @Tags Discovery
// @Param id path int true "Scan ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/discovery/{id} [get]
func GetDiscoveryResult(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, 400, "INVALID_ID", "Invalid scan ID", nil)
	}

	// TODO: Store scan results in database and retrieve by ID
	// For now, return a placeholder
	return ok(c, map[string]interface{}{
		"id":         id,
		"message":    "Scan result storage not yet implemented",
		"scan_again": true,
	})
}

// AddDiscoveredDevice adds a discovered RouterOS device to NAS table
// @Summary add discovered device to NAS
// @Tags Discovery
// @Param payload body discoveryAddPayload true "Device information"
// @Success 200 {object} domain.NetNas
// @Router /api/v1/network/discovery [post]
func AddDiscoveredDevice(c echo.Context) error {
	var payload discoveryAddPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, 400, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	// Check if IP already exists
	var count int64
	GetDB(c).Model(&domain.NetNas{}).Where("ipaddr = ?", payload.IP).Count(&count)
	if count > 0 {
		return fail(c, 409, "IP_EXISTS", "Device with this IP already exists in NAS", nil)
	}

	// Set default name if not provided
	name := payload.Name
	if name == "" {
		name = "Mikrotik-" + payload.IP
	}

	// Set default model if not provided
	model := payload.Model
	if model == "" {
		model = "Mikrotik"
	}

	device := domain.NetNas{
		Name:       name,
		Ipaddr:    payload.IP,
		Secret:    payload.Secret,
		CoaPort:   3799,
		Model:     model,
		VendorCode: "mikrotik",
		Status:    "enabled",
		Tags:      payload.Tags,
		Remark:    "Added via auto-discovery",
	}

	if err := GetDB(c).Create(&device).Error; err != nil {
		return fail(c, 500, "CREATE_FAILED", "Failed to add device to NAS", err.Error())
	}

	return ok(c, device)
}

// AddDiscoveredDevices bulk adds multiple discovered devices to NAS table
// @Summary bulk add discovered devices
// @Tags Discovery
// @Param devices body []discoveryAddPayload true "Array of devices"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/network/discovery/bulk [post]
func AddDiscoveredDevices(c echo.Context) error {
	var payloads []discoveryAddPayload
	if err := c.Bind(&payloads); err != nil {
		return fail(c, 400, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	if len(payloads) == 0 {
		return fail(c, 400, "EMPTY_REQUEST", "No devices provided", nil)
	}

	var addedDevices []domain.NetNas
	var errors []string

	for i, payload := range payloads {
		// Validate IP
		if payload.IP == "" {
			errors = append(errors, "Device at index "+strconv.Itoa(i)+": IP is required")
			continue
		}

		// Validate secret
		if payload.Secret == "" {
			errors = append(errors, "Device at index "+strconv.Itoa(i)+": Secret is required")
			continue
		}

		// Check if IP already exists
		var count int64
		GetDB(c).Model(&domain.NetNas{}).Where("ipaddr = ?", payload.IP).Count(&count)
		if count > 0 {
			errors = append(errors, "Device at index "+strconv.Itoa(i)+": IP "+payload.IP+" already exists")
			continue
		}

		// Set default name if not provided
		name := payload.Name
		if name == "" {
			name = "Mikrotik-" + payload.IP
		}

		// Set default model if not provided
		model := payload.Model
		if model == "" {
			model = "Mikrotik"
		}

		device := domain.NetNas{
			Name:       name,
			Ipaddr:    payload.IP,
			Secret:    payload.Secret,
			CoaPort:   3799,
			Model:     model,
			VendorCode: "mikrotik",
			Status:    "enabled",
			Tags:      payload.Tags,
			Remark:    "Added via auto-discovery",
		}

		if err := GetDB(c).Create(&device).Error; err != nil {
			errors = append(errors, "Device at index "+strconv.Itoa(i)+": "+err.Error())
			continue
		}

		addedDevices = append(addedDevices, device)
	}

	return ok(c, map[string]interface{}{
		"added_count": len(addedDevices),
		"added":       addedDevices,
		"errors":      errors,
	})
}

// registerDiscoveryRoutes registers network discovery routes
func registerDiscoveryRoutes() {
	webserver.ApiPOST("/network/discovery/scan", ScanNetwork)
	webserver.ApiGET("/network/discovery/:id", GetDiscoveryResult)
	webserver.ApiPOST("/network/discovery", AddDiscoveredDevice)
	webserver.ApiPOST("/network/discovery/bulk", AddDiscoveredDevices)
}
