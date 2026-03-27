# Network Device Management System - Design Specification

**Date:** 2026-03-26
**Status:** Draft
**Author:** System Design

---

## 1. Executive Summary

This document describes the design for a comprehensive Network Device Management System that extends the current Radio ISP platform to support multi-vendor device monitoring, inventory management, and alerting for 500,000+ users across multiple providers.

### 1.1 Goals

- **Scalability:** Support 500K users with horizontal scaling architecture
- **Multi-Tenant:** Strict tenant isolation for device visibility
- **Multi-Vendor:** Support MikroTik, Ubiquiti, TP-Link, and other SNMP/API-enabled devices
- **Real-Time:** Sub-minute status updates and alerting
- **Hybrid Monitoring:** Combine push (MikroTik) and poll (Ubiquiti/SNMP) approaches

### 1.2 Non-Goals

- TR-069/CWMP auto-provisioning (existing CPE page is optional)
- Direct firmware upgrades (out of scope for v1)
- Network topology visualization

---

## 2. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Radio Platform                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────────┐    │
│  │   NAS Devices   │     │  Network Device │     │   Agent/Location   │    │
│  │   (RADIUS)      │     │   (Monitoring)   │     │   Hierarchy        │    │
│  │                 │     │                  │     │                    │    │
│  │  MikroTik CCR  │     │  MikroTik       │     │  Provider A        │    │
│  │  MikroTik hAP  │     │  Ubiquiti       │     │    └── Location 1   │    │
│  │  etc.          │     │  TP-Link        │     │    └── Location 2   │    │
│  └────────┬────────┘     └────────┬────────┘     └─────────────────────┘    │
│           │                          │                                         │
│           │                          │                                         │
│           ▼                          ▼                                         │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                    Device Monitoring Service                        │    │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────────┐  │    │
│  │  │   SNMP        │  │   HTTP/API    │  │    Alert          │  │    │
│  │  │   Poller      │  │   Receiver    │  │    Engine         │  │    │
│  │  │               │  │   (MikroTik)  │  │                    │  │    │
│  │  │ • Ubiquiti   │  │               │  │ • Threshold       │  │    │
│  │  │ • TP-Link    │  │ • Metrics     │  │ • Offline         │  │    │
│  │  │ • Switches   │  │ • Status      │  │ • Error           │  │    │
│  │  └───────────────┘  └───────────────┘  └────────────────────┘  │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │                      Unified Device Dashboard                       │    │
│  │                                                                        │    │
│  │  • Device Inventory (all types)    • Real-time Status             │    │
│  │  • Metrics (temp, voltage, signal)  • Alerts & Notifications      │    │
│  │  • Location Grouping               • Remote Actions (reboot)       │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Database Schema

### 3.1 Tables

#### `agent_location` - Location/Branch Management

```sql
CREATE TABLE agent_location (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES net_tenant(id),
    agent_id BIGINT REFERENCES sys_opr(id),
    
    name VARCHAR(255) NOT NULL,
    address TEXT,
    region VARCHAR(100),
    coordinates VARCHAR(100),  -- "lat,lng" format
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_agent_location_tenant ON agent_location(tenant_id);
CREATE INDEX idx_agent_location_agent ON agent_location(agent_id);
```

#### `network_device` - Device Registry

```sql
CREATE TABLE network_device (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES net_tenant(id),
    location_id BIGINT REFERENCES agent_location(id),
    
    -- Device Identity
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL,  -- router, ap, bridge, switch, firewall, other
    vendor VARCHAR(100),               -- MikroTik, Ubiquiti, TP-Link, Cisco, etc.
    model VARCHAR(100),
    serial_number VARCHAR(100),
    firmware_version VARCHAR(100),
    
    -- Network
    ip_address VARCHAR(45) NOT NULL,
    mac_address VARCHAR(17),
    snmp_port INTEGER DEFAULT 161,
    snmp_community VARCHAR(100),        -- Encrypted at rest
    
    -- API Access (for MikroTik)
    api_endpoint VARCHAR(500),
    api_username VARCHAR(100),
    api_password VARCHAR(255),          -- Encrypted
    
    -- Status (updated by monitoring)
    status VARCHAR(20) DEFAULT 'unknown',  -- online, offline, unknown
    last_seen TIMESTAMP,
    last_online TIMESTAMP,
    last_offline TIMESTAMP,
    
    -- Settings
    polling_enabled BOOLEAN DEFAULT true,
    polling_interval INTEGER DEFAULT 60,  -- seconds
    alert_on_offline BOOLEAN DEFAULT true,
    
    -- Metadata
    tags TEXT,                           -- Comma-separated or JSON
    remark TEXT,
    metadata JSONB,                      -- Flexible device-specific data
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_network_device_tenant ON network_device(tenant_id);
CREATE INDEX idx_network_device_location ON network_device(location_id);
CREATE INDEX idx_network_device_status ON network_device(status);
CREATE INDEX idx_network_device_type ON network_device(device_type);
CREATE INDEX idx_network_device_vendor ON network_device(vendor);
CREATE UNIQUE INDEX idx_network_device_ip_tenant ON network_device(ip_address, tenant_id);
```

#### `network_device_metric` - Time-Series Metrics

```sql
CREATE TABLE network_device_metric (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES network_device(id) ON DELETE CASCADE,
    
    metric_type VARCHAR(50) NOT NULL,
    value DECIMAL(15,4) NOT NULL,
    unit VARCHAR(20),
    severity VARCHAR(20) DEFAULT 'normal',  -- normal, warning, critical
    
    collected_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_device_metric_device ON network_device_metric(device_id);
CREATE INDEX idx_device_metric_type ON network_device_metric(metric_type);
CREATE INDEX idx_device_metric_time ON network_device_metric(collected_at);
CREATE INDEX idx_device_metric_device_type ON network_device_metric(device_id, metric_type, collected_at);

-- For cleanup (retention policy)
CREATE INDEX idx_device_metric_cleanup ON network_device_metric(collected_at);
```

#### `network_device_alert` - Alerts

```sql
CREATE TABLE network_device_alert (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES network_device(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES net_tenant(id),
    
    alert_type VARCHAR(50) NOT NULL,    -- offline, online, threshold, error
    severity VARCHAR(20) NOT NULL,      -- info, warning, critical
    message TEXT NOT NULL,
    
    metric_type VARCHAR(50),             -- If threshold alert
    metric_value DECIMAL(15,4),         -- Current value
    threshold_value DECIMAL(15,4),      -- Threshold that triggered
    
    status VARCHAR(20) DEFAULT 'active',  -- active, acknowledged, resolved
    acknowledged_by BIGINT REFERENCES sys_opr(id),
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_device_alert_tenant ON network_device_alert(tenant_id);
CREATE INDEX idx_device_alert_device ON network_device_alert(device_id);
CREATE INDEX idx_device_alert_status ON network_device_alert(status);
CREATE INDEX idx_device_alert_created ON network_device_alert(created_at);
```

### 3.2 Enums

```sql
-- Device Types
CREATE TYPE device_type AS ENUM ('router', 'ap', 'bridge', 'switch', 'firewall', 'modem', 'other');

-- Device Status
CREATE TYPE device_status AS ENUM ('online', 'offline', 'unknown');

-- Alert Severity
CREATE TYPE alert_severity AS ENUM ('info', 'warning', 'critical');

-- Alert Status
CREATE TYPE alert_status AS ENUM ('active', 'acknowledged', 'resolved');
```

---

## 4. API Specification

### 4.1 Endpoints

All endpoints require `Authorization: Bearer <token>` header and enforce tenant isolation.

#### Devices

```
GET    /api/v1/devices
       Query: ?type=router&status=online&location_id=1&page=1&pageSize=20
       Response: { data: Device[], total: number, page: number, pageSize: number }

POST   /api/v1/devices
       Body: { name, device_type, vendor, model, ip_address, snmp_community, location_id, ... }
       Response: { data: Device }

GET    /api/v1/devices/:id
       Response: { data: Device }

PUT    /api/v1/devices/:id
       Body: { name?, ip_address?, snmp_community?, ... }
       Response: { data: Device }

DELETE /api/v1/devices/:id
       Response: { success: true }

GET    /api/v1/devices/:id/metrics
       Query: ?type=temperature&from=2026-03-26T00:00:00Z&to=2026-03-26T23:59:59Z
       Response: { data: Metric[], count: number }

GET    /api/v1/devices/:id/alerts
       Query: ?status=active&page=1&pageSize=20
       Response: { data: Alert[], total: number }

POST   /api/v1/devices/:id/ping
       Response: { success: boolean, latency_ms: number, error?: string }

POST   /api/v1/devices/:id/reboot
       Response: { success: boolean, message: string }

GET    /api/v1/devices/overview
       Response: { total: number, online: number, offline: number, warnings: number }
```

#### Locations

```
GET    /api/v1/locations
       Response: { data: Location[], total: number }

POST   /api/v1/locations
       Body: { name, address, region, coordinates? }
       Response: { data: Location }

PUT    /api/v1/locations/:id
       Body: { name?, address?, region?, coordinates? }
       Response: { data: Location }

DELETE /api/v1/locations/:id
       Response: { success: true }
```

#### Constants

```
GET    /api/v1/device-types
       Response: { data: [{ id: 'router', name: 'Router' }, ...] }

GET    /api/v1/device-vendors
       Response: { data: [{ id: 'mikrotik', name: 'MikroTik' }, ...] }

GET    /api/v1/metric-types
       Response: { data: [{ id: 'temperature', name: 'Temperature', unit: '°C' }, ...] }
```

#### Webhook (for MikroTik push)

```
POST   /api/v1/webhooks/device-metrics
       Body: { device_ip, metrics: [{ type, value, unit }], timestamp }
       Auth: API key in header
       Response: { success: boolean }
```

### 4.2 Response Types

```typescript
interface Device {
  id: number;
  tenant_id: number;
  location_id?: number;
  name: string;
  device_type: 'router' | 'ap' | 'bridge' | 'switch' | 'firewall' | 'modem' | 'other';
  vendor: string;
  model: string;
  serial_number?: string;
  firmware_version?: string;
  ip_address: string;
  mac_address?: string;
  snmp_port: number;
  status: 'online' | 'offline' | 'unknown';
  last_seen?: string;
  tags?: string[];
  remark?: string;
  created_at: string;
  updated_at: string;
}

interface DeviceMetric {
  id: number;
  device_id: number;
  metric_type: string;
  value: number;
  unit: string;
  severity: 'normal' | 'warning' | 'critical';
  collected_at: string;
}

interface DeviceAlert {
  id: number;
  device_id: number;
  device_name: string;
  tenant_id: number;
  alert_type: string;
  severity: 'info' | 'warning' | 'critical';
  message: string;
  status: 'active' | 'acknowledged' | 'resolved';
  created_at: string;
  acknowledged_at?: string;
  resolved_at?: string;
}

interface Location {
  id: number;
  tenant_id: number;
  agent_id?: number;
  name: string;
  address?: string;
  region?: string;
  coordinates?: string;
  created_at: string;
}
```

---

## 5. Monitoring Service

### 5.1 Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                    Monitoring Worker Pool                        │
│                    (Background Goroutines)                        │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  SNMP Poller   │  │  Status Checker │  │  Alert Processor│ │
│  │  (Ubiquiti,    │  │  (Ping all     │  │  (Threshold,   │ │
│  │   TP-Link)     │  │   devices)     │  │   offline)      │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘ │
│           │                      │                      │          │
│           └──────────────────────┼──────────────────────┘          │
│                                  │                                  │
│                                  ▼                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                      Metrics Database                         │ │
│  │                  (PostgreSQL + Redis Cache)                    │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                  │                                  │
│                                  ▼                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Alert Notification                         │ │
│  │                (Email, Webhook, In-App)                     │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### 5.2 Polling Strategy

| Device Type | Method | Interval | Metrics |
|-------------|--------|----------|---------|
| MikroTik | HTTP Push (webhook) | Real-time | CPU, RAM, Temp, Interface stats |
| Ubiquiti | SNMP Poll | 60s | Signal, Noise, Traffic, Uptime |
| TP-Link | SNMP Poll | 60s | CPU, Memory, Port Status |
| Generic | Ping | 30s | Online/Offline only |

### 5.3 Threshold Configuration

Default thresholds (configurable per device):

```yaml
thresholds:
  temperature:
    warning: 50  # °C
    critical: 70 # °C
  voltage:
    warning: 22  # V
    critical: 20 # V
  signal:  # Ubiquiti
    warning: -70 # dBm
    critical: -80 # dBm
  uptime:
    offline_after: 120 # seconds without ping response
```

### 5.4 MikroTik Push Integration

MikroTik script to push metrics:

```routeros
/system scheduler
add name=metrics-push interval=60s on-event={
    :local metrics ""
    
    # CPU Load
    :local cpuLoad [/system resource get cpu-load]
    
    # Memory
    :local memUsed [/system resource get used-memory]
    :local memTotal [/system resource get total-memory]
    :local memPercent (($memUsed * 100) / $memTotal)
    
    # Temperature (if available)
    :local temp [/system health get temperature]
    
    # Build JSON
    :local json "{\"device_ip\":\"$[/ip address find interface=ether1]\",\"metrics\":["
    :set json "$json{\"type\":\"cpu_load\",\"value\":$cpuLoad,\"unit\":\"%\"}"
    :set json "$json,{\"type\":\"memory\",\"value\":$memPercent,\"unit\":\"%\"}"
    :if ([:typeof $temp] = "num") do={
        :set json "$json,{\"type\":\"temperature\",\"value\":$temp,\"unit\":\"C\"}"
    }
    :set json "$json]}"
    
    # Send to Radio
    /tool fetch url="http://YOUR_SERVER:1816/api/v1/webhooks/device-metrics" \
        method=post data=$json
}
```

---

## 6. Frontend Design

### 6.1 Device List Page (`/devices`)

**Layout:**
```
┌─────────────────────────────────────────────────────────────────────────┐
│  Devices                                    [+ Add Device] [↻ Refresh]  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Filter: [All Types ▼] [All Status ▼] [All Locations ▼] [🔍]  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌────────┬────────────────┬───────────┬──────────┬────────────────┐   │
│  │ Status │ Device        │ Type      │ Location │ Last Seen      │   │
│  ├────────┼────────────────┼───────────┼──────────┼────────────────┤   │
│  │ 🟢     │ MikroTik CCR   │ Router    │ Baghdad  │ 2 min ago     │   │
│  │ 🟢     │ NanoStation M5 │ Bridge    │ Baghdad  │ 5 min ago     │   │
│  │ 🔴     │ TP-Link C7     │ AP        │ Fallujah │ 2 hours ago   │   │
│  │ 🟢     │ LiteBeam M5    │ Bridge    │ Basra    │ 1 min ago     │   │
│  │ 🟡     │ Switch 24P     │ Switch    │ Baghdad  │ 10 min ago    │   │
│  └────────┴────────────────┴───────────┴──────────┴────────────────┘   │
│                                                                         │
│  Showing 1-5 of 47 devices              [< Prev] [1] [2] [3] [Next >]│
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Status Colors:**
- 🟢 Online (green)
- 🔴 Offline (red)
- 🟡 Warning (yellow - has critical metric)
- ⚫ Unknown (gray)

### 6.2 Device Detail Page (`/devices/:id`)

**Tabs:**
1. **Overview** - Status, info, quick actions
2. **Metrics** - Real-time and historical graphs
3. **Alerts** - Active and past alerts
4. **Settings** - Configuration

**Overview Tab:**
```
┌─────────────────────────────────────────────────────────────────────────┐
│  🟢 MikroTik CCR1009                              [Edit] [Reboot] [⋮] │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  IP Address:    192.168.1.1                                             │
│  MAC Address:  AA:BB:CC:DD:EE:FF                                      │
│  Type:         Router                                                  │
│  Vendor:       MikroTik                                                │
│  Model:        CCR1009-7G-1C-1S+                                       │
│  Firmware:     7.15.3                                                  │
│  Location:     Baghdad HQ                                               │
│                                                                         │
│  ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐    │
│  │ CPU Load         │  │ Memory           │  │ Temperature       │    │
│  │ ████████░░ 78%  │  │ █████░░░░ 52%   │  │ ██░░░░░░░ 42°C  │    │
│  │ Normal           │  │ Normal           │  │ Normal            │    │
│  └───────────────────┘  └───────────────────┘  └───────────────────┘    │
│                                                                         │
│  Uptime: 45 days 7 hours                                               │
│  Last Seen: 2 minutes ago                                              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.3 RTL Support

For Arabic interface:
- All text right-aligned
- Icons remain on correct side
- Tables: headers right, data left-aligned
- Status indicators always visible
- Monospace fonts for commands (LTR)

---

## 7. Tenant Isolation

### 7.1 Rules

1. **All queries** filter by `tenant_id` from context
2. **Super Admin** sees all devices (filter by `tenant_id` optional)
3. **Provider Admin** sees only their `tenant_id` devices
4. **Agent** sees devices in their assigned locations
5. **Locations** filtered by agent's assigned locations

### 7.2 Implementation

```go
// Middleware extracts tenant_id
func GetTenantID(c echo.Context) int64 {
    tenantID, _ := tenant.FromContext(c.Request().Context())
    return tenantID
}

// All list endpoints use TenantScope
func listDevices(c echo.Context) error {
    tenantID := GetTenantID(c)
    db := GetDB(c).Where("tenant_id = ?", tenantID)
    // ...
}
```

---

## 8. Scalability Considerations

### 8.1 Database

- **Indexes** on `tenant_id`, `status`, `device_type`, `location_id`
- **Partitioning** by `tenant_id` for very large deployments (future)
- **Metric retention:** 30 days raw, then aggregated
- **Archival:** Move old data to cold storage

### 8.2 Monitoring

- **Worker pool:** Configurable goroutines for parallel polling
- **Batch queries:** Fetch devices in batches for polling
- **Redis cache:** Cache device status for dashboard
- **Connection pooling:** Reuse SNMP/HTTP connections

### 8.3 Sharding Path (Future)

```
Phase 1: Shared DB (current)
    │
    ▼
Phase 2: Read replicas for monitoring queries
    │
    ▼
Phase 3: Shard by tenant_id
    │
    ▼
Phase 4: Per-tenant databases
```

---

## 9. Implementation Phases

### Phase 1: Core Infrastructure
- Database tables
- Basic CRUD API with tenant isolation
- Device list UI
- Location management

### Phase 2: Monitoring
- SNMP polling for Ubiquiti/TP-Link
- Ping-based status checks
- Metrics storage
- Alert generation

### Phase 3: MikroTik Integration
- Webhook receiver for MikroTik push
- MikroTik API client
- Real-time metrics display

### Phase 4: Advanced Features
- Remote reboot
- Config templates
- Notifications (email/webhook)
- Reporting

---

## 10. Configuration

### 10.1 Database Migration

```bash
# Run migration
./migrate -up -limit 1

# Migration includes:
# - agent_location table
# - network_device table
# - network_device_metric table
# - network_device_alert table
```

### 10.2 Environment Variables

```bash
# Monitoring
DEVICE_POLL_INTERVAL=60        # seconds
DEVICE_PING_INTERVAL=30       # seconds
DEVICE_METRIC_RETENTION=30    # days

# SNMP Defaults
SNMP_COMMUNITY=public
SNMP_TIMEOUT=5                # seconds
SNMP_RETRIES=3

# Alert Thresholds
ALERT_TEMP_WARNING=50
ALERT_TEMP_CRITICAL=70
ALERT_OFFLINE_THRESHOLD=120   # seconds
```

---

## 11. Security Considerations

1. **SNMP Community** stored encrypted, never logged
2. **API Credentials** encrypted at rest
3. **Tenant Isolation** enforced at middleware level
4. **Rate Limiting** on webhook endpoints
5. **Audit Logging** for device changes
6. **HTTPS Only** for all device communication

---

## 12. Open Questions

- [ ] Do you need multi-tenant DNS/naming for devices?
- [ ] Should alerts trigger notifications (email/SMS)?
- [ ] Do you need role-based permissions per device type?
- [ ] What is the expected device-to-user ratio?
- [ ] Do you need device grouping (hierarchical locations)?

---

## 13. Revision History

| Date | Version | Changes |
|------|--------|---------|
| 2026-03-26 | 1.0 | Initial draft |
