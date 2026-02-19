package acs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

// Provisioner handles automatic provisioning of CPE devices.
// It manages the creation of PPPoE credentials and links them to RADIUS users.
type Provisioner struct {
	repo     *Repository
	config   ProvisioningConfig
}

// NewProvisioner creates a new Provisioner instance.
//
// Parameters:
//   - repo: The repository for database operations
//   - config: The provisioning configuration
//
// Returns:
//   - *Provisioner: The provisioner instance
func NewProvisioner(repo *Repository, config ProvisioningConfig) *Provisioner {
	// Apply defaults
	if config.PasswordLength == 0 {
		config.PasswordLength = 12
	}
	if config.UsernamePrefix == "" {
		config.UsernamePrefix = "cpe_"
	}

	return &Provisioner{
		repo:   repo,
		config: config,
	}
}

// ProvisionResult contains the result of a provisioning operation.
type ProvisionResult struct {
	Device        *CPEDevice
	RadiusUser    *domain.RadiusUser
	PPPoEUsername string
	PPPoEPassword string
	Configs       []ParameterValueStruct
	Error         error
}

// AutoProvision handles the automatic provisioning flow for a CPE device.
// This is called when a CPE sends an Inform message and needs to be provisioned.
//
// Flow:
//  1. Check if auto-provisioning is enabled
//  2. Check if device already has a RADIUS user
//  3. Generate PPPoE credentials
//  4. Create RADIUS user
//  5. Update device with RADIUS user link
//  6. Generate configuration parameters to send to CPE
//
// Parameters:
//   - device: The CPE device to provision
//   - inform: The Inform message from the CPE
//
// Returns:
//   - *ProvisionResult: The provisioning result
func (p *Provisioner) AutoProvision(device *CPEDevice, inform *Inform) *ProvisionResult {
	// Check if auto-provisioning is enabled
	if !p.config.AutoProvision {
		return &ProvisionResult{
			Device: device,
			Error:  fmt.Errorf("auto-provisioning is disabled"),
		}
	}

	// Check if device is already provisioned
	if device.Status == DeviceStatusProvisioned && device.RadiusUserID != nil {
		// Already provisioned, return existing config
		return p.getExistingConfig(device)
	}

	// Generate PPPoE credentials
	username, password, err := p.GeneratePPPoECredentials(device)
	if err != nil {
		p.updateDeviceStatus(device.ID, DeviceStatusFailed, err.Error())
		return &ProvisionResult{
			Device: device,
			Error:  fmt.Errorf("failed to generate credentials: %w", err),
		}
	}

	// Create RADIUS user
	radiusUser, err := p.CreateRadiusUser(username, password, device)
	if err != nil {
		p.updateDeviceStatus(device.ID, DeviceStatusFailed, err.Error())
		return &ProvisionResult{
			Device:        device,
			PPPoEUsername: username,
			PPPoEPassword: password,
			Error:         fmt.Errorf("failed to create RADIUS user: %w", err),
		}
	}

	// Update device with RADIUS user link
	if err := p.repo.SetDeviceProvisioned(device.ID, radiusUser.ID, username, password); err != nil {
		return &ProvisionResult{
			Device:        device,
			RadiusUser:    radiusUser,
			PPPoEUsername: username,
			PPPoEPassword: password,
			Error:         fmt.Errorf("failed to update device: %w", err),
		}
	}

	// Generate configuration parameters
	configs := p.generateConfigParameters(username, password)

	return &ProvisionResult{
		Device:        device,
		RadiusUser:    radiusUser,
		PPPoEUsername: username,
		PPPoEPassword: password,
		Configs:       configs,
	}
}

// GeneratePPPoECredentials generates PPPoE username and password for a device.
//
// Parameters:
//   - device: The CPE device
//
// Returns:
//   - string: The generated username
//   - string: The generated password
//   - error: Error if generation fails
func (p *Provisioner) GeneratePPPoECredentials(device *CPEDevice) (string, string, error) {
	// Generate username: prefix + device serial (truncated) + random suffix
	serial := device.SerialNumber
	if len(serial) > 20 {
		serial = serial[:20]
	}

	randomSuffix, err := generateRandomString(4)
	if err != nil {
		return "", "", err
	}

	username := fmt.Sprintf("%s%s%s", p.config.UsernamePrefix, serial, randomSuffix)

	// Generate password
	password, err := generateRandomString(p.config.PasswordLength)
	if err != nil {
		return "", "", err
	}

	return username, password, nil
}

// CreateRadiusUser creates a RADIUS user for the provisioned device.
//
// Parameters:
//   - username: The PPPoE username
//   - password: The PPPoE password
//   - device: The CPE device
//
// Returns:
//   - *domain.RadiusUser: The created RADIUS user
//   - error: Error if creation fails
func (p *Provisioner) CreateRadiusUser(username, password string, device *CPEDevice) (*domain.RadiusUser, error) {
	// Get profile if specified
	var profile *domain.RadiusProfile
	if p.config.DefaultProfileID > 0 {
		var err error
		profile, err = p.repo.GetRadiusProfileByID(p.config.DefaultProfileID)
		if err != nil {
			// Log warning but continue with default values
			profile = nil
		}
	}

	// Build RADIUS user
	radiusUser := &domain.RadiusUser{
		Username: username,
		Password: password,
		Status:   "enabled",
		NodeId:   p.config.DefaultNodeID,
		Remark:   fmt.Sprintf("Auto-provisioned for CPE: %s (%s)", device.SerialNumber, device.Manufacturer),
	}

	// Apply profile settings if available
	if profile != nil {
		radiusUser.ProfileId = profile.ID
		radiusUser.AddrPool = profile.AddrPool
		radiusUser.UpRate = profile.UpRate
		radiusUser.DownRate = profile.DownRate
		radiusUser.DataQuota = profile.DataQuota
		radiusUser.ActiveNum = profile.ActiveNum
		radiusUser.ProfileLinkMode = 1 // Dynamic link - use real-time from profile
	}

	// Create the user
	createdUser, err := p.repo.CreateRadiusUser(radiusUser)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// generateConfigParameters generates the SetParameterValues for PPPoE configuration.
func (p *Provisioner) generateConfigParameters(username, password string) []ParameterValueStruct {
	return NewSetParameterValuesForPPPoE(
		username,
		password,
		p.config.PPPoEUsernamePath,
		p.config.PPPoEPasswordPath,
	)
}

// getExistingConfig returns configuration for an already provisioned device.
func (p *Provisioner) getExistingConfig(device *CPEDevice) *ProvisionResult {
	if device.RadiusUserID == nil {
		return &ProvisionResult{
			Device: device,
			Error:  fmt.Errorf("device has no linked RADIUS user"),
		}
	}

	user, err := p.repo.GetRadiusUserByID(*device.RadiusUserID)
	if err != nil {
		return &ProvisionResult{
			Device: device,
			Error:  fmt.Errorf("failed to get RADIUS user: %w", err),
		}
	}

	configs := p.generateConfigParameters(user.Username, user.Password)

	return &ProvisionResult{
		Device:        device,
		RadiusUser:    user,
		PPPoEUsername: user.Username,
		PPPoEPassword: user.Password,
		Configs:       configs,
	}
}

// updateDeviceStatus updates the device status in the database.
func (p *Provisioner) updateDeviceStatus(id int64, status DeviceStatus, errMsg string) {
	_ = p.repo.UpdateDeviceStatus(id, status, errMsg)
}

// generateRandomString generates a cryptographically secure random string.
//
// Parameters:
//   - length: The length of the string to generate
//
// Returns:
//   - string: The generated string
//   - error: Error if generation fails
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// HandleInform processes an Inform message from a CPE device.
// This is the main entry point for processing CPE Inform messages.
//
// Parameters:
//   - inform: The Inform message from the CPE
//   - sourceIP: The source IP address of the CPE
//
// Returns:
//   - *CPEDevice: The CPE device (created or updated)
//   - *ProvisionResult: Provisioning result (may be nil if no provisioning needed)
//   - error: Error if processing fails
func (p *Provisioner) HandleInform(inform *Inform, sourceIP string) (*CPEDevice, *ProvisionResult, error) {
	// Extract device info from Inform
	deviceID := inform.DeviceId

	// Try to find existing device
	device, err := p.repo.GetDeviceBySerialNumber(deviceID.SerialNumber)
	if err != nil && err.Error() != "record not found" {
		return nil, nil, fmt.Errorf("failed to lookup device: %w", err)
	}

	now := time.Now()

	// Create new device if not found
	if device == nil {
		device = &CPEDevice{
			SerialNumber:     deviceID.SerialNumber,
			OUI:             deviceID.OUI,
			Manufacturer:    deviceID.Manufacturer,
			ProductClass:    deviceID.ProductClass,
			Status:          DeviceStatusPending,
			AutoProvision:  p.config.AutoProvision,
			ProfileID:      int64Ptr(p.config.DefaultProfileID),
			LastInform:     &now,
			LastIP:          sourceIP,
			SoftwareVersion: extractSoftwareVersion(inform),
		}

		device, err = p.repo.CreateDevice(device)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create device: %w", err)
		}
	} else {
		// Update existing device
		_ = p.repo.UpdateDeviceLastInform(device.ID, sourceIP)

		// Update software version if available
		if swVersion := extractSoftwareVersion(inform); swVersion != "" {
			device.SoftwareVersion = swVersion
		}
		device.LastIP = sourceIP
	}

	// Check if device needs provisioning
	var provisionResult *ProvisionResult

	// Check for boot events (new device or reboot)
	if IsBootEvent(inform) && device.NeedsProvisioning() {
		// Auto-provision new devices on boot
		result := p.AutoProvision(device, inform)
		provisionResult = result
		if result.Error == nil {
			device.Status = DeviceStatusProvisioned
		}
	} else if device.Status == DeviceStatusPending && device.AutoProvision {
		// Also provision on periodic if still pending
		result := p.AutoProvision(device, inform)
		provisionResult = result
		if result.Error == nil {
			device.Status = DeviceStatusProvisioned
		}
	}

	return device, provisionResult, nil
}

// extractSoftwareVersion extracts software version from Inform parameters.
func extractSoftwareVersion(inform *Inform) string {
	// Try common parameter paths
	paths := []string{
		"Device.DeviceInfo.SoftwareVersion",
		"Device.DeviceInfo.FirmwareVersion",
		"InternetGatewayDevice.DeviceInfo.SoftwareVersion",
		"InternetGatewayDevice.DeviceInfo.FirmwareVersion",
	}

	for _, path := range paths {
		if val := GetParameterValue(inform, path); val != "" {
			return val
		}
	}

	return ""
}

// int64Ptr returns a pointer to an int64.
func int64Ptr(v int64) *int64 {
	return &v
}

// ProvisionDeviceManual manually provisions a device with specific credentials.
// This is useful for admin-initiated provisioning.
//
// Parameters:
//   - deviceID: The device ID to provision
//   - username: The PPPoE username (if empty, auto-generate)
//   - password: The PPPoE password (if empty, auto-generate)
//
// Returns:
//   - *ProvisionResult: The provisioning result
func (p *Provisioner) ProvisionDeviceManual(deviceID int64, username, password string) (*ProvisionResult, error) {
	device, err := p.repo.GetDeviceByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Generate credentials if not provided
	if username == "" || password == "" {
		var err error
		username, password, err = p.GeneratePPPoECredentials(device)
		if err != nil {
			return nil, fmt.Errorf("failed to generate credentials: %w", err)
		}
	}

	// Check if user already exists with this username
	existing, _ := p.repo.GetRadiusUserByUsername(username)
	if existing != nil {
		return nil, fmt.Errorf("username already exists: %s", username)
	}

	// Create RADIUS user
	radiusUser, err := p.CreateRadiusUser(username, password, device)
	if err != nil {
		p.updateDeviceStatus(device.ID, DeviceStatusFailed, err.Error())
		return nil, fmt.Errorf("failed to create RADIUS user: %w", err)
	}

	// Update device with RADIUS user link
	if err := p.repo.SetDeviceProvisioned(device.ID, radiusUser.ID, username, password); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// Generate configuration parameters
	configs := p.generateConfigParameters(username, password)

	return &ProvisionResult{
		Device:        device,
		RadiusUser:    radiusUser,
		PPPoEUsername: username,
		PPPoEPassword: password,
		Configs:       configs,
	}, nil
}

// ReprovisionDevice reprovisions a device with new credentials.
// This is useful for password rotation or credential changes.
//
// Parameters:
//   - deviceID: The device ID to reprovision
//
// Returns:
//   - *ProvisionResult: The provisioning result
//   - error: Error if provisioning fails
func (p *Provisioner) ReprovisionDevice(deviceID int64) (*ProvisionResult, error) {
	device, err := p.repo.GetDeviceByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Delete existing RADIUS user if linked
	if device.RadiusUserID != nil {
		// Note: In a real implementation, we'd delete the user
		// For now, just create a new user
	}

	// Generate new credentials
	username, password, err := p.GeneratePPPoECredentials(device)
	if err != nil {
		return nil, fmt.Errorf("failed to generate credentials: %w", err)
	}

	// Create new RADIUS user
	radiusUser, err := p.CreateRadiusUser(username, password, device)
	if err != nil {
		return nil, fmt.Errorf("failed to create RADIUS user: %w", err)
	}

	// Update device
	if err := p.repo.SetDeviceProvisioned(device.ID, radiusUser.ID, username, password); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// Generate configuration parameters
	configs := p.generateConfigParameters(username, password)

	return &ProvisionResult{
		Device:        device,
		RadiusUser:    radiusUser,
		PPPoEUsername: username,
		PPPoEPassword: password,
		Configs:       configs,
	}, nil
}
