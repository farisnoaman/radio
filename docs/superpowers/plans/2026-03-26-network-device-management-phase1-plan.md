# Network Device Management - Phase 1 Implementation Plan

**Date:** 2026-03-26
**Status:** Draft
**Phase:** 1 - Core Infrastructure

---

## 1. Overview

Phase 1 implements the core infrastructure for network device management:
- Database tables for devices, locations, metrics, and alerts
- Basic CRUD API with tenant isolation
- Device list and detail UI
- Location management

---

## 2. Deliverables

### 2.1 Backend

| # | Task | File(s) | Effort |
|---|------|---------|--------|
| 1 | Database migration | `migrations/xxx_network_devices.sql` | Low |
| 2 | Domain models | `internal/domain/network_device.go` | Low |
| 3 | Repository | `internal/repository/network_device.go` | Medium |
| 4 | API endpoints | `internal/adminapi/device.go` | Medium |
| 5 | Register routes | `internal/adminapi/adminapi.go` | Low |
| 6 | Location API | `internal/adminapi/location.go` | Medium |

### 2.2 Frontend

| # | Task | File(s) | Effort |
|---|------|---------|--------|
| 1 | Device list page | `web/src/pages/Devices.tsx` | Medium |
| 2 | Device detail page | `web/src/pages/DeviceDetail.tsx` | Medium |
| 3 | Location management | `web/src/pages/Locations.tsx` | Medium |
| 4 | API hooks | `web/src/hooks/useDevices.ts` | Low |
| 5 | Menu update | `web/src/components/CustomMenu.tsx` | Low |

---

## 3. Database Migration

### 3.1 Migration File

Create: `migrations/20260326120000_network_devices.sql`

```sql
-- Location table
CREATE TABLE IF NOT EXISTS agent_location (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES net_tenant(id) ON DELETE CASCADE,
    agent_id BIGINT REFERENCES sys_opr(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    region VARCHAR(100),
    coordinates VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_agent_location_tenant ON agent_location(tenant_id);
CREATE INDEX IF NOT EXISTS idx_agent_location_agent ON agent_location(agent_id);

-- Network device table
CREATE TABLE IF NOT EXISTS network_device (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES net_tenant(id) ON DELETE CASCADE,
    location_id BIGINT REFERENCES agent_location(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL DEFAULT 'router',
    vendor VARCHAR(100),
    model VARCHAR(100),
    serial_number VARCHAR(100),
    firmware_version VARCHAR(100),
    ip_address VARCHAR(45) NOT NULL,
    mac_address VARCHAR(17),
    snmp_port INTEGER DEFAULT 161,
    snmp_community VARCHAR(100),
    api_endpoint VARCHAR(500),
    api_username VARCHAR(100),
    api_password VARCHAR(255),
    status VARCHAR(20) DEFAULT 'unknown',
    last_seen TIMESTAMP,
    last_online TIMESTAMP,
    last_offline TIMESTAMP,
    polling_enabled BOOLEAN DEFAULT true,
    polling_interval INTEGER DEFAULT 60,
    alert_on_offline BOOLEAN DEFAULT true,
    tags TEXT,
    remark TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_network_device_tenant ON network_device(tenant_id);
CREATE INDEX IF NOT EXISTS idx_network_device_location ON network_device(location_id);
CREATE INDEX IF NOT EXISTS idx_network_device_status ON network_device(status);
CREATE INDEX IF NOT EXISTS idx_network_device_type ON network_device(device_type);
CREATE INDEX IF NOT EXISTS idx_network_device_ip_tenant ON network_device(ip_address, tenant_id);

-- Device metric table
CREATE TABLE IF NOT EXISTS network_device_metric (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES network_device(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL,
    value DECIMAL(15,4) NOT NULL,
    unit VARCHAR(20),
    severity VARCHAR(20) DEFAULT 'normal',
    collected_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_metric_device ON network_device_metric(device_id);
CREATE INDEX IF NOT EXISTS idx_device_metric_type ON network_device_metric(metric_type);
CREATE INDEX IF NOT EXISTS idx_device_metric_time ON network_device_metric(collected_at);
CREATE INDEX IF NOT EXISTS idx_device_metric_cleanup ON network_device_metric(collected_at);

-- Device alert table
CREATE TABLE IF NOT EXISTS network_device_alert (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES network_device(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES net_tenant(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metric_type VARCHAR(50),
    metric_value DECIMAL(15,4),
    threshold_value DECIMAL(15,4),
    status VARCHAR(20) DEFAULT 'active',
    acknowledged_by BIGINT REFERENCES sys_opr(id),
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_alert_tenant ON network_device_alert(tenant_id);
CREATE INDEX IF NOT EXISTS idx_device_alert_device ON network_device_alert(device_id);
CREATE INDEX IF NOT EXISTS idx_device_alert_status ON network_device_alert(status);
```

---

## 4. Backend Implementation

### 4.1 Domain Models

**File:** `internal/domain/network_device.go`

```go
package domain

import (
    "time"
)

// Location represents a physical location/branch
type AgentLocation struct {
    ID        int64      `json:"id" gorm:"primaryKey"`
    TenantID  int64      `json:"tenant_id"`
    AgentID   *int64     `json:"agent_id"`
    Name      string     `json:"name"`
    Address   string     `json:"address,omitempty"`
    Region    string     `json:"region,omitempty"`
    Coordinates string   `json:"coordinates,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

func (AgentLocation) TableName() string {
    return "agent_location"
}

// NetworkDevice represents a managed network device
type NetworkDevice struct {
    ID                int64      `json:"id" gorm:"primaryKey"`
    TenantID          int64      `json:"tenant_id"`
    LocationID        *int64     `json:"location_id,omitempty"`
    
    // Identity
    Name              string     `json:"name"`
    DeviceType       string     `json:"device_type"`
    Vendor           string     `json:"vendor,omitempty"`
    Model            string     `json:"model,omitempty"`
    SerialNumber     string     `json:"serial_number,omitempty"`
    FirmwareVersion  string     `json:"firmware_version,omitempty"`
    
    // Network
    IPAddress        string     `json:"ip_address"`
    MacAddress       string     `json:"mac_address,omitempty"`
    SNMPPort        int        `json:"snmp_port"`
    SNMPCommunity   string     `json:"-"`
    
    // API
    APIEndpoint      string     `json:"api_endpoint,omitempty"`
    APIUsername      string     `json:"api_username,omitempty"`
    APIPassword      string     `json:"-"`
    
    // Status
    Status           string     `json:"status"`
    LastSeen         *time.Time `json:"last_seen,omitempty"`
    LastOnline       *time.Time `json:"last_online,omitempty"`
    LastOffline      *time.Time `json:"last_offline,omitempty"`
    
    // Settings
    PollingEnabled   bool       `json:"polling_enabled"`
    PollingInterval  int        `json:"polling_interval"`
    AlertOnOffline   bool       `json:"alert_on_offline"`
    
    // Metadata
    Tags             string     `json:"tags,omitempty"`
    Remark           string     `json:"remark,omitempty"`
    
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (NetworkDevice) TableName() string {
    return "network_device"
}

// NetworkDeviceMetric represents a metric reading
type NetworkDeviceMetric struct {
    ID          int64     `json:"id" gorm:"primaryKey"`
    DeviceID    int64     `json:"device_id"`
    MetricType  string    `json:"metric_type"`
    Value       float64   `json:"value"`
    Unit        string    `json:"unit,omitempty"`
    Severity    string    `json:"severity"`
    CollectedAt time.Time `json:"collected_at"`
    CreatedAt   time.Time `json:"created_at"`
}

func (NetworkDeviceMetric) TableName() string {
    return "network_device_metric"
}

// NetworkDeviceAlert represents an alert
type NetworkDeviceAlert struct {
    ID              int64      `json:"id" gorm:"primaryKey"`
    DeviceID       int64      `json:"device_id"`
    TenantID       int64      `json:"tenant_id"`
    AlertType      string     `json:"alert_type"`
    Severity       string     `json:"severity"`
    Message        string     `json:"message"`
    MetricType     string     `json:"metric_type,omitempty"`
    MetricValue    *float64   `json:"metric_value,omitempty"`
    ThresholdValue *float64   `json:"threshold_value,omitempty"`
    Status         string     `json:"status"`
    AcknowledgedBy *int64     `json:"acknowledged_by,omitempty"`
    AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
    ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
    CreatedAt      time.Time  `json:"created_at"`
}

func (NetworkDeviceAlert) TableName() string {
    return "network_device_alert"
}

// Device type constants
const (
    DeviceTypeRouter   = "router"
    DeviceTypeAP       = "ap"
    DeviceTypeBridge    = "bridge"
    DeviceTypeSwitch    = "switch"
    DeviceTypeFirewall  = "firewall"
    DeviceTypeModem    = "modem"
    DeviceTypeOther    = "other"
)

// Device status constants
const (
    DeviceStatusOnline  = "online"
    DeviceStatusOffline = "offline"
    DeviceStatusUnknown = "unknown"
)

// Alert constants
const (
    AlertTypeOffline   = "offline"
    AlertTypeOnline    = "online"
    AlertTypeThreshold = "threshold"
    AlertTypeError     = "error"
)

const (
    SeverityInfo     = "info"
    SeverityWarning  = "warning"
    SeverityCritical = "critical"
)

const (
    AlertStatusActive       = "active"
    AlertStatusAcknowledged = "acknowledged"
    AlertStatusResolved     = "resolved"
)
```

### 4.2 API Endpoints

**File:** `internal/adminapi/device.go`

```go
package adminapi

import (
    "net/http"
    "strconv"
    "time"
    
    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/webserver"
)

// Device constants
var deviceTypes = []map[string]string{
    {"id": "router", "name": "Router"},
    {"id": "ap", "name": "Access Point"},
    {"id": "bridge", "name": "Bridge"},
    {"id": "switch", "name": "Switch"},
    {"id": "firewall", "name": "Firewall"},
    {"id": "modem", "name": "Modem"},
    {"id": "other", "name": "Other"},
}

var deviceVendors = []map[string]string{
    {"id": "mikrotik", "name": "MikroTik"},
    {"id": "ubiquiti", "name": "Ubiquiti"},
    {"id": "tplink", "name": "TP-Link"},
    {"id": "cisco", "name": "Cisco"},
    {"id": "tp-link", "name": "TP-Link"},
    {"id": "other", "name": "Other"},
}

var metricTypes = []map[string]interface{}{
    {"id": "temperature", "name": "Temperature", "unit": "°C"},
    {"id": "cpu_load", "name": "CPU Load", "unit": "%"},
    {"id": "memory", "name": "Memory", "unit": "%"},
    {"id": "voltage", "name": "Voltage", "unit": "V"},
    {"id": "signal", "name": "Signal Strength", "unit": "dBm"},
    {"id": "uptime", "name": "Uptime", "unit": "hours"},
}

func registerDeviceRoutes() {
    // Devices
    webserver.ApiGET("/devices", listDevices)
    webserver.ApiPOST("/devices", createDevice)
    webserver.ApiGET("/devices/:id", getDevice)
    webserver.ApiPUT("/devices/:id", updateDevice)
    webserver.ApiDELETE("/devices/:id", deleteDevice)
    webserver.ApiGET("/devices/:id/metrics", getDeviceMetrics)
    webserver.ApiGET("/devices/:id/alerts", getDeviceAlerts)
    webserver.ApiGET("/devices/overview", getDevicesOverview)
    
    // Locations
    webserver.ApiGET("/locations", listLocations)
    webserver.ApiPOST("/locations", createLocation)
    webserver.ApiPUT("/locations/:id", updateLocation)
    webserver.ApiDELETE("/locations/:id", deleteLocation)
    
    // Constants
    webserver.ApiGET("/device-types", getDeviceTypes)
    webserver.ApiGET("/device-vendors", getDeviceVendors)
    webserver.ApiGET("/metric-types", getMetricTypes)
}

func listDevices(c echo.Context) error {
    page, pageSize := parsePagination(c)
    tenantID := GetTenantID(c)
    
    query := GetDB(c).Model(&NetworkDevice{}).Where("tenant_id = ?", tenantID)
    
    // Filters
    if deviceType := c.QueryParam("type"); deviceType != "" {
        query = query.Where("device_type = ?", deviceType)
    }
    if status := c.QueryParam("status"); status != "" {
        query = query.Where("status = ?", status)
    }
    if locationID := c.QueryParam("location_id"); locationID != "" {
        query = query.Where("location_id = ?", locationID)
    }
    
    var total int64
    query.Count(&total)
    
    var devices []NetworkDevice
    err := query.Order("updated_at DESC").
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&devices).Error
    if err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return paged(c, devices, total, page, pageSize)
}

func createDevice(c echo.Context) error {
    tenantID := GetTenantID(c)
    
    var device NetworkDevice
    if err := c.Bind(&device); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
    }
    
    device.TenantID = tenantID
    device.Status = DeviceStatusUnknown
    
    if err := GetDB(c).Create(&device).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return ok(c, device)
}

func getDevice(c echo.Context) error {
    id, err := parseIDParam(c, "id")
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
    }
    
    tenantID := GetTenantID(c)
    
    var device NetworkDevice
    if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
        return fail(c, http.StatusNotFound, "DEVICE_NOT_FOUND", err.Error())
    }
    
    return ok(c, device)
}

func updateDevice(c echo.Context) error {
    id, err := parseIDParam(c, "id")
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
    }
    
    tenantID := GetTenantID(c)
    
    var device NetworkDevice
    if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
        return fail(c, http.StatusNotFound, "DEVICE_NOT_FOUND", err.Error())
    }
    
    var updates map[string]interface{}
    if err := c.Bind(&updates); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
    }
    
    // Remove protected fields
    delete(updates, "id")
    delete(updates, "tenant_id")
    delete(updates, "created_at")
    
    if err := GetDB(c).Model(&device).Updates(updates).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return ok(c, device)
}

func deleteDevice(c echo.Context) error {
    id, err := parseIDParam(c, "id")
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
    }
    
    tenantID := GetTenantID(c)
    
    result := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&NetworkDevice{})
    if result.Error != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", result.Error.Error())
    }
    if result.RowsAffected == 0 {
        return fail(c, http.StatusNotFound, "DEVICE_NOT_FOUND", "Device not found")
    }
    
    return ok(c, map[string]bool{"success": true})
}

func getDeviceMetrics(c echo.Context) error {
    id, err := parseIDParam(c, "id")
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
    }
    
    tenantID := GetTenantID(c)
    
    // Verify device belongs to tenant
    var device NetworkDevice
    if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
        return fail(c, http.StatusNotFound, "DEVICE_NOT_FOUND", err.Error())
    }
    
    // Query params for time range
    from := c.QueryParam("from")
    to := c.QueryParam("to")
    metricType := c.QueryParam("type")
    
    query := GetDB(c).Model(&NetworkDeviceMetric{}).Where("device_id = ?", id)
    
    if metricType != "" {
        query = query.Where("metric_type = ?", metricType)
    }
    if from != "" {
        query = query.Where("collected_at >= ?", from)
    }
    if to != "" {
        query = query.Where("collected_at <= ?", to)
    }
    
    var metrics []NetworkDeviceMetric
    if err := query.Order("collected_at DESC").Limit(100).Find(&metrics).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return ok(c, map[string]interface{}{
        "data": metrics,
        "count": len(metrics),
    })
}

func getDeviceAlerts(c echo.Context) error {
    id, err := parseIDParam(c, "id")
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
    }
    
    tenantID := GetTenantID(c)
    
    var device NetworkDevice
    if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&device).Error; err != nil {
        return fail(c, http.StatusNotFound, "DEVICE_NOT_FOUND", err.Error())
    }
    
    status := c.QueryParam("status")
    page, pageSize := parsePagination(c)
    
    query := GetDB(c).Model(&NetworkDeviceAlert{}).Where("device_id = ?", id)
    if status != "" {
        query = query.Where("status = ?", status)
    }
    
    var total int64
    query.Count(&total)
    
    var alerts []NetworkDeviceAlert
    if err := query.Order("created_at DESC").
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&alerts).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return paged(c, alerts, total, page, pageSize)
}

func getDevicesOverview(c echo.Context) error {
    tenantID := GetTenantID(c)
    
    var stats struct {
        Total    int64 `json:"total"`
        Online   int64 `json:"online"`
        Offline  int64 `json:"offline"`
        Unknown  int64 `json:"unknown"`
    }
    
    db := GetDB(c).Model(&NetworkDevice{}).Where("tenant_id = ?", tenantID)
    
    db.Count(&stats.Total)
    db.Where("status = ?", DeviceStatusOnline).Count(&stats.Online)
    db.Where("status = ?", DeviceStatusOffline).Count(&stats.Offline)
    db.Where("status = ?", DeviceStatusUnknown).Count(&stats.Unknown)
    
    return ok(c, stats)
}

// Location endpoints
func listLocations(c echo.Context) error {
    tenantID := GetTenantID(c)
    
    var locations []AgentLocation
    if err := GetDB(c).Where("tenant_id = ?", tenantID).Order("name ASC").Find(&locations).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return ok(c, map[string]interface{}{
        "data": locations,
        "count": len(locations),
    })
}

func createLocation(c echo.Context) error {
    tenantID := GetTenantID(c)
    
    var location AgentLocation
    if err := c.Bind(&location); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
    }
    
    location.TenantID = tenantID
    
    if err := GetDB(c).Create(&location).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return ok(c, location)
}

func updateLocation(c echo.Context) error {
    id, err := parseIDParam(c, "id")
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
    }
    
    tenantID := GetTenantID(c)
    
    var location AgentLocation
    if err := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).First(&location).Error; err != nil {
        return fail(c, http.StatusNotFound, "LOCATION_NOT_FOUND", err.Error())
    }
    
    var updates map[string]interface{}
    if err := c.Bind(&updates); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
    }
    
    delete(updates, "id")
    delete(updates, "tenant_id")
    
    if err := GetDB(c).Model(&location).Updates(updates).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
    }
    
    return ok(c, location)
}

func deleteLocation(c echo.Context) error {
    id, err := parseIDParam(c, "id")
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
    }
    
    tenantID := GetTenantID(c)
    
    result := GetDB(c).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&AgentLocation{})
    if result.Error != nil {
        return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", result.Error.Error())
    }
    if result.RowsAffected == 0 {
        return fail(c, http.StatusNotFound, "LOCATION_NOT_FOUND", "Location not found")
    }
    
    return ok(c, map[string]bool{"success": true})
}

// Constants endpoints
func getDeviceTypes(c echo.Context) error {
    return ok(c, map[string]interface{}{"data": deviceTypes})
}

func getDeviceVendors(c echo.Context) error {
    return ok(c, map[string]interface{}{"data": deviceVendors})
}

func getMetricTypes(c echo.Context) error {
    return ok(c, map[string]interface{}{"data": metricTypes})
}
```

---

## 5. Frontend Implementation

### 5.1 Device List Page

**File:** `web/src/pages/Devices.tsx`

```tsx
import { useState } from 'react';
import {
  List,
  Datagrid,
  TextField,
  SelectField,
  useListContext,
  TopToolbar,
  RefreshButton,
  ExportButton,
  CreateButton,
  FilterButton,
  TextInput,
  SelectInput,
  useTranslate,
  useGetList,
  useRecordContext,
  ShowButton,
  EditButton,
  DeleteButton,
  useNavigate,
} from 'react-admin';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Avatar,
  Chip,
  Stack,
  IconButton,
  Tooltip,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import {
  Router as RouterIcon,
  Wifi as WifiIcon,
  Hub as HubIcon,
  SettingsEthernet as SwitchIcon,
  Lan as LanIcon,
  MoreVert as MoreIcon,
} from '@mui/icons-material';

const deviceTypeIcons: Record<string, React.ReactNode> = {
  router: <RouterIcon />,
  ap: <WifiIcon />,
  bridge: <HubIcon />,
  switch: <SwitchIcon />,
  firewall: <LanIcon />,
  other: <RouterIcon />,
};

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default'> = {
  online: 'success',
  offline: 'error',
  unknown: 'default',
};

const DeviceFilters = [
  <SelectInput
    key="type"
    source="device_type"
    label="Type"
    choices={[
      { id: 'router', name: 'Router' },
      { id: 'ap', name: 'Access Point' },
      { id: 'bridge', name: 'Bridge' },
      { id: 'switch', name: 'Switch' },
      { id: 'other', name: 'Other' },
    ]}
  />,
  <SelectInput
    key="status"
    source="status"
    label="Status"
    choices={[
      { id: 'online', name: 'Online' },
      { id: 'offline', name: 'Offline' },
      { id: 'unknown', name: 'Unknown' },
    ]}
  />,
];

const DeviceListContent = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data, total, isLoading } = useListContext();

  if (isLoading) return null;

  const getStatusIcon = (status: string) => {
    const colors = {
      online: '#4caf50',
      offline: '#f44336',
      unknown: '#9e9e9e',
    };
    return (
      <Box
        sx={{
          width: 12,
          height: 12,
          borderRadius: '50%',
          bgcolor: colors[status] || colors.unknown,
        }}
      />
    );
  };

  return (
    <Box>
      <Card
        elevation={0}
        sx={{
          borderRadius: 2,
          border: `1px solid ${theme.palette.divider}`,
          mb: 2,
        }}
      >
        <Box
          sx={{
            px: 2,
            py: 1,
            bgcolor:
              theme.palette.mode === 'dark'
                ? 'rgba(255,255,255,0.02)'
                : 'rgba(0,0,0,0.01)',
            borderBottom: `1px solid ${theme.palette.divider}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Typography variant="body2" color="text.secondary">
            Total: <strong>{total?.toLocaleString() || 0}</strong> Devices
          </Typography>
        </Box>

        {isMobile ? (
          <Box sx={{ p: 2 }}>
            {data?.map((device: any) => (
              <Card key={device.id} variant="outlined" sx={{ mb: 2, borderRadius: 3 }}>
                <CardContent>
                  <Stack direction="row" alignItems="center" spacing={2} mb={2}>
                    <Avatar sx={{ bgcolor: 'secondary.main' }}>
                      {deviceTypeIcons[device.device_type] || <RouterIcon />}
                    </Avatar>
                    <Box sx={{ flexGrow: 1 }}>
                      <Typography variant="h6">{device.name}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {device.vendor} {device.model}
                      </Typography>
                    </Box>
                    <Chip
                      label={device.status}
                      size="small"
                      color={statusColors[device.status]}
                    />
                  </Stack>
                  <Stack spacing={1}>
                    <Typography variant="caption">
                      <strong>IP:</strong> {device.ip_address}
                    </Typography>
                    <Typography variant="caption">
                      <strong>Type:</strong> {device.device_type}
                    </Typography>
                  </Stack>
                </CardContent>
              </Card>
            ))}
          </Box>
        ) : (
          <Datagrid
            rowClick="show"
            bulkActionButtons={false}
            sx={{
              '& .RaDatagrid-thead': {
                bgcolor:
                  theme.palette.mode === 'dark'
                    ? 'rgba(255,255,255,0.05)'
                    : 'rgba(0,0,0,0.02)',
              },
            }}
          >
            <TextField source="name" label="Name" />
            <SelectField
              source="device_type"
              label="Type"
              choices={[
                { id: 'router', name: 'Router' },
                { id: 'ap', name: 'AP' },
                { id: 'bridge', name: 'Bridge' },
                { id: 'switch', name: 'Switch' },
                { id: 'other', name: 'Other' },
              ]}
            />
            <TextField source="ip_address" label="IP Address" />
            <TextField source="vendor" label="Vendor" />
            <TextField source="model" label="Model" />
            <SelectField
              source="status"
              label="Status"
              choices={[
                { id: 'online', name: 'Online' },
                { id: 'offline', name: 'Offline' },
                { id: 'unknown', name: 'Unknown' },
              ]}
            />
            <ShowButton />
            <EditButton />
            <DeleteButton />
          </Datagrid>
        )}
      </Card>
    </Box>
  );
};

export const DeviceList = () => (
  <List
    actions={
      <TopToolbar>
        <RefreshButton />
        <ExportButton />
        <CreateButton />
        <FilterButton filters={DeviceFilters} />
      </TopToolbar>
    }
    filters={DeviceFilters}
    sort={{ field: 'updated_at', order: 'DESC' }}
  >
    <DeviceListContent />
  </List>
);

export const DeviceShow = () => {
  const record = useRecordContext();
  const translate = useTranslate();
  const theme = useTheme();

  if (!record) return null;

  return (
    <Card sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h5">{record.name}</Typography>
        <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2 }}>
          <Box>
            <Typography variant="caption" color="text.secondary">
              IP Address
            </Typography>
            <Typography>{record.ip_address}</Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              Status
            </Typography>
            <Chip
              label={record.status}
              size="small"
              color={statusColors[record.status]}
            />
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              Type
            </Typography>
            <Typography>{record.device_type}</Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              Vendor
            </Typography>
            <Typography>{record.vendor}</Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              Model
            </Typography>
            <Typography>{record.model}</Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">
              Last Seen
            </Typography>
            <Typography>
              {record.last_seen
                ? new Date(record.last_seen).toLocaleString()
                : 'Never'}
            </Typography>
          </Box>
        </Box>
      </Stack>
    </Card>
  );
};
```

### 5.2 Location Management

**File:** `web/src/pages/Locations.tsx`

```tsx
import {
  List,
  Datagrid,
  TextField,
  Create,
  Edit,
  SimpleForm,
  TextInput,
  useTranslate,
  useRecordContext,
  TopToolbar,
  CreateButton,
  EditButton,
  DeleteButton,
} from 'react-admin';
import { Card, CardContent, Box, Typography, Stack } from '@mui/material';

export const LocationList = () => (
  <List
    actions={
      <TopToolbar>
        <CreateButton />
      </TopToolbar>
    }
  >
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="name" label="Name" />
      <TextField source="region" label="Region" />
      <TextField source="address" label="Address" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export const LocationCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="name" label="Location Name" required />
      <TextInput source="region" label="Region" />
      <TextInput source="address" label="Address" multiline />
      <TextInput source="coordinates" label="Coordinates (lat,lng)" />
    </SimpleForm>
  </Create>
);

export const LocationEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="name" label="Location Name" required />
      <TextInput source="region" label="Region" />
      <TextInput source="address" label="Address" multiline />
      <TextInput source="coordinates" label="Coordinates (lat,lng)" />
    </SimpleForm>
  </Edit>
);
```

---

## 6. Menu Update

**File:** `web/src/components/CustomMenu.tsx`

Add new menu item:

```tsx
// Add to menu items
{
  to: '/devices',
  labelKey: 'menu.devices',
  icon: <RouterIcon />,
  permissions: ['super', 'admin'],
},
{
  to: '/locations',
  labelKey: 'menu.locations',
  icon: <LocationOnIcon />,
  permissions: ['super', 'admin'],
},
```

---

## 7. Translation Keys

### English (`web/src/i18n/en-US.ts`)

```typescript
resources: {
  devices: {
    name: 'Network Device |||| Network Devices',
    fields: {
      name: 'Device Name',
      device_type: 'Type',
      vendor: 'Vendor',
      model: 'Model',
      ip_address: 'IP Address',
      status: 'Status',
      last_seen: 'Last Seen',
    },
    types: {
      router: 'Router',
      ap: 'Access Point',
      bridge: 'Bridge',
      switch: 'Switch',
      firewall: 'Firewall',
      other: 'Other',
    },
    status: {
      online: 'Online',
      offline: 'Offline',
      unknown: 'Unknown',
    },
  },
  locations: {
    name: 'Location |||| Locations',
    fields: {
      name: 'Name',
      region: 'Region',
      address: 'Address',
      coordinates: 'Coordinates',
    },
  },
},
menu: {
  devices: 'Network Devices',
  locations: 'Locations',
},
```

### Arabic (`web/src/i18n/ar.ts`)

```typescript
resources: {
  devices: {
    name: 'جهاز الشبكة |||| أجهزة الشبكة',
    fields: {
      name: 'اسم الجهاز',
      device_type: 'النوع',
      vendor: 'الشركة المصنعة',
      model: 'الموديل',
      ip_address: 'عنوان IP',
      status: 'الحالة',
      last_seen: 'آخر ظهور',
    },
    types: {
      router: 'موجه',
      ap: 'نقطة وصول',
      bridge: 'جسر',
      switch: 'محول',
      firewall: 'جدار حماية',
      other: 'أخرى',
    },
    status: {
      online: 'متصل',
      offline: 'غير متصل',
      unknown: 'غير معروف',
    },
  },
  locations: {
    name: 'الموقع |||| المواقع',
    fields: {
      name: 'الاسم',
      region: 'المنطقة',
      address: 'العنوان',
      coordinates: 'الإحداثيات',
    },
  },
},
menu: {
  devices: 'أجهزة الشبكة',
  locations: 'المواقع',
},
```

---

## 8. Implementation Order

### Step 1: Database Migration
1. Create migration file
2. Run migration
3. Verify tables created

### Step 2: Backend Domain
1. Create domain models
2. Add to adminapi/adminapi.go registration
3. Test API endpoints

### Step 3: Frontend Core
1. Create device list page
2. Create device show page
3. Add menu items
4. Add translations

### Step 4: Location Management
1. Create location CRUD pages
2. Add to menu
3. Add translations

### Step 5: Integration
1. Connect device to location
2. Test tenant isolation
3. Verify filters work

---

## 9. Testing Checklist

- [ ] Can create device
- [ ] Can edit device
- [ ] Can delete device
- [ ] Device list shows only tenant's devices
- [ ] Can create location
- [ ] Can assign device to location
- [ ] Filters work (type, status, location)
- [ ] Translations work (EN, AR)
- [ ] Mobile responsive

---

## 10. Effort Estimate

| Component | Time |
|-----------|------|
| Database Migration | 30 min |
| Backend Domain + API | 2 hours |
| Frontend Device List | 2 hours |
| Frontend Device Detail | 1 hour |
| Location Management | 1 hour |
| Menu + Translations | 30 min |
| **Total** | **~7 hours** |
