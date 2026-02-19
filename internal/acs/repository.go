package acs

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// Repository provides database operations for TR-069 ACS entities.
// It handles CRUD operations for CPE devices and integrates with RADIUS users.
//
// All methods are safe for concurrent use as they rely on GORM's connection pool.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Repository instance.
//
// Parameters:
//   - db: GORM database connection
//
// Returns:
//   - *Repository: The repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetDB returns the underlying GORM database connection.
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// AutoMigrate runs auto-migration for ACS tables.
// This should be called during application startup.
//
// Returns:
//   - error: Migration error if any
func (r *Repository) AutoMigrate() error {
	return r.db.AutoMigrate(&CPEDevice{})
}

// CreateDevice creates a new CPE device record.
// If a device with the same serial number already exists, it returns the existing device.
//
// Parameters:
//   - device: The CPEDevice to create
//
// Returns:
//   - *CPEDevice: The created or existing device
//   - error: Database error if any
func (r *Repository) CreateDevice(device *CPEDevice) (*CPEDevice, error) {
	// Check if device already exists
	existing, err := r.GetDeviceBySerialNumber(device.SerialNumber)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Create new device
	if err := r.db.Create(device).Error; err != nil {
		return nil, err
	}
	return device, nil
}

// GetDeviceByID retrieves a CPE device by its primary key ID.
//
// Parameters:
//   - id: The device ID
//
// Returns:
//   - *CPEDevice: The device, or nil if not found
//   - error: Database error if any
func (r *Repository) GetDeviceByID(id int64) (*CPEDevice, error) {
	var device CPEDevice
	err := r.db.First(&device, id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetDeviceBySerialNumber retrieves a CPE device by its serial number.
//
// Parameters:
//   - serialNumber: The device serial number
//
// Returns:
//   - *CPEDevice: The device, or nil if not found
//   - error: Database error if any
func (r *Repository) GetDeviceBySerialNumber(serialNumber string) (*CPEDevice, error) {
	var device CPEDevice
	err := r.db.Where("serial_number = ?", serialNumber).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetDeviceByOUIAndSerial retrieves a CPE device by OUI and serial number combination.
//
// Parameters:
//   - oui: The Organization Unique Identifier
//   - serialNumber: The device serial number
//
// Returns:
//   - *CPEDevice: The device, or nil if not found
//   - error: Database error if any
func (r *Repository) GetDeviceByOUIAndSerial(oui, serialNumber string) (*CPEDevice, error) {
	var device CPEDevice
	err := r.db.Where("oui = ? AND serial_number = ?", oui, serialNumber).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// UpdateDevice updates a CPE device record.
//
// Parameters:
//   - device: The CPEDevice to update
//
// Returns:
//   - error: Database error if any
func (r *Repository) UpdateDevice(device *CPEDevice) error {
	return r.db.Save(device).Error
}

// DeleteDevice soft-deletes a CPE device record.
//
// Parameters:
//   - id: The device ID to delete
//
// Returns:
//   - error: Database error if any
func (r *Repository) DeleteDevice(id int64) error {
	return r.db.Delete(&CPEDevice{}, id).Error
}

// ListDevices retrieves all CPE devices with optional filtering.
//
// Parameters:
//   - status: Filter by status (empty string for all)
//   - limit: Maximum number of results (0 for no limit)
//   - offset: Result offset for pagination
//
// Returns:
//   - []*CPEDevice: List of devices
//   - error: Database error if any
func (r *Repository) ListDevices(status string, limit, offset int) ([]*CPEDevice, error) {
	var devices []*CPEDevice
	query := r.db.Model(&CPEDevice{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("created_at DESC").Find(&devices).Error
	return devices, err
}

// ListDevicesNeedingProvisioning retrieves all devices that need auto-provisioning.
//
// Returns:
//   - []*CPEDevice: List of devices needing provisioning
//   - error: Database error if any
func (r *Repository) ListDevicesNeedingProvisioning() ([]*CPEDevice, error) {
	var devices []*CPEDevice
	err := r.db.Where("status = ? AND auto_provision = ?",
		DeviceStatusPending, true).
		Find(&devices).Error
	return devices, err
}

// UpdateDeviceStatus updates the status of a CPE device.
//
// Parameters:
//   - id: The device ID
//   - status: The new status
//   - errorMsg: Error message (for failed status)
//
// Returns:
//   - error: Database error if any
func (r *Repository) UpdateDeviceStatus(id int64, status DeviceStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == DeviceStatusProvisioned {
		now := time.Now()
		updates["provisioned_at"] = &now
	}

	if errorMsg != "" {
		updates["last_error"] = errorMsg
	}

	return r.db.Model(&CPEDevice{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateDeviceLastInform updates the last inform time and IP address.
//
// Parameters:
//   - id: The device ID
//   - ip: The source IP address
//
// Returns:
//   - error: Database error if any
func (r *Repository) UpdateDeviceLastInform(id int64, ip string) error {
	now := time.Now()
	return r.db.Model(&CPEDevice{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_inform": &now,
		"last_ip":     ip,
		"updated_at":  now,
	}).Error
}

// SetDeviceProvisioned marks a device as provisioned with RADIUS user link.
//
// Parameters:
//   - id: The device ID
//   - radiusUserID: The RADIUS user ID
//   - pppoeUsername: The PPPoE username
//   - pppoePassword: The PPPoE password
//
// Returns:
//   - error: Database error if any
func (r *Repository) SetDeviceProvisioned(id int64, radiusUserID int64, pppoeUsername, pppoePassword string) error {
	now := time.Now()
	return r.db.Model(&CPEDevice{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          DeviceStatusProvisioned,
		"provisioned_at":  &now,
		"radius_user_id":  radiusUserID,
		"pppoe_username":  pppoeUsername,
		"pppoe_password":  pppoePassword,
		"updated_at":      now,
	}).Error
}

// CreateRadiusUser creates a new RADIUS user for auto-provisioning.
// This is used by the provisioning engine to create PPPoE accounts.
//
// Parameters:
//   - user: The RadiusUser to create
//
// Returns:
//   - *domain.RadiusUser: The created user
//   - error: Database error if any
func (r *Repository) CreateRadiusUser(user *domain.RadiusUser) (*domain.RadiusUser, error) {
	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetRadiusUserByUsername retrieves a RADIUS user by username.
//
// Parameters:
//   - username: The PPPoE username
//
// Returns:
//   - *domain.RadiusUser: The user, or nil if not found
//   - error: Database error if any
func (r *Repository) GetRadiusUserByUsername(username string) (*domain.RadiusUser, error) {
	var user domain.RadiusUser
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetRadiusProfileByID retrieves a RADIUS profile by ID.
//
// Parameters:
//   - id: The profile ID
//
// Returns:
//   - *domain.RadiusProfile: The profile, or nil if not found
//   - error: Database error if any
func (r *Repository) GetRadiusProfileByID(id int64) (*domain.RadiusProfile, error) {
	var profile domain.RadiusProfile
	err := r.db.First(&profile, id).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// CountDevicesByStatus counts devices by status.
//
// Parameters:
//   - status: The status to count (empty for all)
//
// Returns:
//   - int64: The count
//   - error: Database error if any
func (r *Repository) CountDevicesByStatus(status DeviceStatus) (int64, error) {
	var count int64
	query := r.db.Model(&CPEDevice{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

// GetDeviceWithRadiusUser retrieves a CPE device with its associated RADIUS user.
//
// Parameters:
//   - id: The device ID
//
// Returns:
//   - *CPEDevice: The device
//   - *domain.RadiusUser: The associated RADIUS user (if any)
//   - error: Database error if any
func (r *Repository) GetDeviceWithRadiusUser(id int64) (*CPEDevice, *domain.RadiusUser, error) {
	device, err := r.GetDeviceByID(id)
	if err != nil {
		return nil, nil, err
	}

	if device.RadiusUserID == nil {
		return device, nil, nil
	}

	user, err := r.GetRadiusUserByID(*device.RadiusUserID)
	if err != nil {
		return device, nil, err
	}

	return device, user, nil
}

// GetRadiusUserByID retrieves a RADIUS user by ID.
func (r *Repository) GetRadiusUserByID(id int64) (*domain.RadiusUser, error) {
	var user domain.RadiusUser
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SearchDevices searches for devices by various criteria.
//
// Parameters:
//   - query: Search query (matches serial number, manufacturer, product class)
//   - limit: Maximum results
//
// Returns:
//   - []*CPEDevice: Matching devices
//   - error: Database error if any
func (r *Repository) SearchDevices(query string, limit int) ([]*CPEDevice, error) {
	var devices []*CPEDevice
	searchPattern := "%" + query + "%"

	dbQuery := r.db.Where(
		"serial_number LIKE ? OR manufacturer LIKE ? OR product_class LIKE ?",
		searchPattern, searchPattern, searchPattern,
	)

	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}

	err := dbQuery.Order("created_at DESC").Find(&devices).Error
	return devices, err
}
