# ToughRADIUS ISP Management Platform - Feature Enhancement Proposal

This document outlines comprehensive feature enhancements to transform ToughRADIUS into a top-tier ISP management platform. Each section addresses specific functional areas with implementation recommendations based on the existing codebase architecture.

---

## 1. Real-Time Network Monitoring Dashboard

### 1.1 Current State Analysis

The existing dashboard ([`internal/adminapi/dashboard.go`](internal/adminapi/dashboard.go)) provides:
- Basic user statistics (total, online, disabled, expired)
- Authentication trend (last 7 days)
- Traffic statistics (24-hour hourly data)
- Profile distribution pie chart

### 1.2 Proposed Enhancements

#### 1.2.1 Real-Time Metrics Streaming

Implement WebSocket-based real-time updates for live network metrics:

```go
// internal/adminapi/websocket/metrics.go
// MetricsBroadcaster broadcasts real-time metrics to connected dashboard clients
type MetricsBroadcaster struct {
    hub    *MetricsHub
    NASIP  string
}

type NetworkMetrics struct {
    Timestamp        time.Time `json:"timestamp"`
    ActiveSessions   int64     `json:"active_sessions"`
    TotalThroughput  struct {
        UploadMbps   float64 `json:"upload_mbps"`
        DownloadMbps float64 `json:"download_mbps"`
    } `json:"total_throughput"`
    NASMetrics       []NASMetrics `json:"nas_metrics"`
    ActiveAlerts     []Alert      `json:"active_alerts"`
}

type NASMetrics struct {
    NASIP       string  `json:"nas_ip"`
    NASName     string  `json:"nas_name"`
    OnlineUsers int64   `json:"online_users"`
    UploadMbps  float64 `json:"upload_mbps"`
    DownloadMbps float64 `json:"download_mbps"`
    LatencyMs   float64 `json:"latency_ms"`
    Status      string  `json:"status"` // online, degraded, offline
}

type Alert struct {
    ID        string    `json:"id"`
    Severity  string    `json:"severity"` // critical, warning, info
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    NASIP     string    `json:"nas_ip,omitempty"`
}
```

**Implementation Steps:**
1. Create WebSocket hub in `internal/adminapi/websocket/`
2. Add metrics collection goroutine in [`internal/app/jobs.go`](internal/app/jobs.go)
3. Implement client subscription to specific NAS or global metrics
4. Add reconnection logic and heartbeat in frontend

#### 1.2.2 Predictive Analytics Engine

Implement ML-based prediction for network issues:

```go
// internal/adminapi/analytics/predictor.go
// TrafficPredictor predicts future traffic patterns and potential issues
type TrafficPredictor struct {
    model           *timeSeriesModel
    historicalData  []TrafficDataPoint
    predictionWindow time.Duration // e.g., 24 hours
}

type TrafficDataPoint struct {
    Timestamp     time.Time
    UploadBytes   int64
    DownloadBytes int64
    SessionCount  int64
    ErrorRate     float64
}

type PredictionResult struct {
    PredictedPeakTime    time.Time `json:"predicted_peak_time"`
    PredictedPeakUpload  float64    `json:"predicted_peak_upload_mbps"`
    PredictedPeakDownload float64   `json:"predicted_peak_download_mbps"`
    ConfidenceScore      float64    `json:"confidence_score"`
    RiskFactors          []RiskFactor `json:"risk_factors"`
}

type RiskFactor struct {
    Type        string `json:"type"` // bandwidth, sessions, nas_health
    Description string `json:"description"`
    Probability float64 `json:"probability"` // 0.0 - 1.0
    Threshold   float64 `json:"threshold"`
}
```

**Prediction Models:**
- **Bandwidth Prediction**: Use exponential smoothing for traffic forecasting
- **Session Surge Detection**: Anomaly detection based on historical patterns
- **NAS Health Score**: Composite metric from latency, error rate, response time

#### 1.2.3 Network Topology Map

Add visual network topology with live status:

```typescript
// web/src/components/NetworkTopology.tsx
interface TopologyNode {
  id: string;
  type: 'nas' | 'router' | 'switch' | 'gateway';
  name: string;
  ip: string;
  status: 'online' | 'degraded' | 'offline';
  metrics: {
    cpu: number;
    memory: number;
    uptime: number;
    sessionCount: number;
  };
  position: { x: number; y: number };
}
```

### 1.3 Frontend Dashboard Enhancements

| Feature | Implementation | Priority |
|---------|----------------|----------|
| Live session counter | WebSocket + animated counter | High |
| Bandwidth gauge | Real-time throughput visualization | High |
| NAS status grid | Color-coded status cards | High |
| Alert notification center | Toast notifications + panel | Medium |
| Traffic heatmap | Hourly/daily traffic visualization | Medium |
| Quality of Service metrics | Jitter, latency, packet loss charts | Medium |

---

## 2. Expired Voucher Management

### 2.1 Current State Analysis

The existing voucher system ([`internal/adminapi/vouchers.go`](internal/adminapi/vouchers.go)) supports:
- Voucher batch creation with configurable expiration (fixed or first-use)
- PIN-protected vouchers
- Voucher extension ([`VoucherExtendRequest`](internal/adminapi/vouchers.go:393))
- Voucher top-up ([`VoucherTopupRequest`](internal/adminapi/vouchers.go:398))
- Auto-renewal subscriptions ([`VoucherSubscriptionRequest`](internal/adminapi/vouchers.go:412))

### 2.2 Proposed Enhancements

#### 2.2.1 Automated Expiration Handling

```go
// internal/app/voucher_expiry.go
// VoucherExpiryHandler manages expired voucher automation
type VoucherExpiryHandler struct {
    db *gorm.DB
    config *VoucherExpiryConfig
}

type VoucherExpiryConfig struct {
    ExpiryAction          string // "disable", "extend", "notify", "archive"
    GracePeriodHours      int
    AutoExtendDays        int
    NotifyBeforeDays      int
    NotifyTemplateID      string
    ArchiveAfterDays      int
}

// HandleExpiredVouchers processes vouchers that have passed their expiration
func (h *VoucherExpiryHandler) HandleExpiredVouchers() error {
    expiredVouchers, err := h.getExpiredVouchers()
    if err != nil {
        return err
    }
    
    for _, voucher := range expiredVouchers {
        switch h.config.ExpiryAction {
        case "disable":
            h.disableVoucherUser(voucher)
        case "extend":
            h.autoExtendVoucher(voucher)
        case "notify":
            h.sendExpiryNotification(voucher)
        case "archive":
            h.archiveVoucher(voucher)
        }
    }
    return nil
}
```

#### 2.2.2 Voucher Renewal Campaigns

```go
// internal/adminapi/voucher_campaigns.go
// VoucherCampaign manages bulk renewal and promotional campaigns
type VoucherCampaign struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Type        string    `json:"type"` // renewal, upgrade, promotion
    TargetFilter CampaignFilter `json:"target_filter"`
    Offer       CampaignOffer `json:"offer"`
    Schedule    CampaignSchedule `json:"schedule"`
    Status      string    `json:"status"` // draft, scheduled, active, completed
}

type CampaignFilter struct {
    ExpiryWithinDays int     `json:"expiry_within_days"`
    ProfileIDs       []int64 `json:"profile_ids"`
    MinUsageGB       int64   `json:"min_usage_gb"`
    ExcludedAgents   []int64 `json:"excluded_agents"`
}

type CampaignOffer struct {
    DiscountPercent int     `json:"discount_percent"`
    BonusDays       int     `json:"bonus_days"`
    BonusDataMB     int64   `json:"bonus_data_mb"`
    NewProductID    int64   `json:"new_product_id"`
}
```

#### 2.2.3 Voucher Transfer & Assignment

Extend the existing voucher transfer functionality ([`web/src/components/VoucherTransferDialog.tsx`](web/src/components/VoucherTransferDialog.tsx)):

- **Agent-to-Agent Transfer**: Allow agents to transfer vouchers between accounts
- **Bulk Assignment**: Assign vouchers to specific users or groups
- **Voucher Import**: CSV import for external voucher systems

#### 2.2.4 Voucher Usage Analytics

```go
// internal/adminapi/voucher_analytics.go
type VoucherAnalytics struct {
    TotalIssued      int64   `json:"total_issued"`
    TotalRedeemed    int64   `json:"total_redeemed"`
    TotalExpired     int64   `json:"total_expired"`
    RedemptionRate   float64 `json:"redemption_rate"`
    AvgUsageDays     float64 `json:"avg_usage_days"`
    TopProducts      []ProductUsage `json:"top_products"`
    ExpiryForecast   []ForecastPoint `json:"expiry_forecast"`
}
```

---

## 3. Session and Logging Management

### 3.1 Current State Analysis

**Sessions** ([`internal/adminapi/sessions.go`](internal/adminapi/sessions.go)):
- Online session listing with filters
- Force disconnect via CoA (Change of Authorization)
- Session data includes: username, NAS, IP, MAC, session time

**System Logs** ([`internal/adminapi/system_log.go`](internal/adminapi/system_log.go)):
- Operation logs (SysOprLog)
- Basic filtering by operator, action, keyword
- Admin-only access

### 3.2 Proposed Enhancements

#### 3.2.1 Comprehensive Session Logging

```go
// internal/domain/session_log.go
// SessionLog records detailed session lifecycle events
type SessionLog struct {
    ID             int64     `json:"id" gorm:"primaryKey"`
    SessionID      string    `json:"session_id" gorm:"index"`
    Username       string    `json:"username" gorm:"index"`
    NASIP          string    `json:"nas_ip"`
    NASName        string    `json:"nas_name"`
    FramedIP       string    `json:"framed_ip"`
    MACAddress     string    `json:"mac_address"`
    EventType      string    `json:"event_type"` // start, stop, update, disconnect, timeout
    EventTime      time.Time `json:"event_time"`
    InputBytes     int64     `json:"input_bytes"`
    OutputBytes    int64     `json:"output_bytes"`
    Duration       int64     `json:"duration_seconds"`
    TerminateCause string    `json:"terminate_cause"`
    IPChange       bool      `json:"ip_change"` // IP address changed during session
    DataCapHit     bool      `json:"data_cap_hit"`
    TimeCapHit     bool      `json:"time_cap_hit"`
}

// SessionAudit provides session analysis
type SessionAudit struct {
    SessionID       string    `json:"session_id"`
    Username        string    `json:"username"`
    StartTime       time.Time `json:"start_time"`
    EndTime         time.Time `json:"end_time"`
    Duration        int64     `json:"duration_seconds"`
    TotalData       int64     `json:"total_bytes"`
    AvgBandwidth    float64   `json:"avg_bandwidth_mbps"`
    PeakBandwidth  float64   `json:"peak_bandwidth_mbps"`
    IPChanges       int       `json:"ip_changes"`
    DisconnectCount int       `json:"disconnect_count"`
}
```

#### 3.2.2 Enhanced Log Management System

```go
// internal/adminapi/log_management.go
// LogManager provides centralized log management
type LogManager struct {
    db *gorm.DB
}

// LogQueryRequest defines log search parameters
type LogQueryRequest struct {
    StartTime    time.Time `json:"start_time"`
    EndTime      time.Time `json:"end_time"`
    LogType      string    `json:"log_type"` // system, session, auth, accounting, api
    Level        string    `json:"level"` // debug, info, warn, error
    Source       string    `json:"source"` // nas_ip, username, operator
    Keyword      string    `json:"keyword"`
    Pagination   Pagination `json:"pagination"`
}

// LogAggregation provides statistical summaries
type LogAggregation struct {
    TotalCount      int64            `json:"total_count"`
    ByLevel          map[string]int64 `json:"by_level"`
    ByType          map[string]int64 `json:"by_type"`
    BySource        map[string]int64 `json:"by_source"`
    TimeSeries      []TimeSeriesPoint `json:"time_series"`
    TopErrors       []ErrorSummary    `json:"top_errors"`
}
```

#### 3.2.3 Log Archival and Retention

```go
// internal/app/log_archival.go
// LogArchival manages log rotation and archival
type LogArchivalConfig struct {
    ActiveRetentionDays  int  // Days to keep in main table
    ArchiveRetentionDays int  // Days to keep in archive
    ArchiveTableName     string
    CompressOldLogs      bool
    DeleteAfterArchive   bool
}

// Scheduled job to archive logs
func (a *Application) SchedLogArchivalTask() {
    // Move old logs to archive table
    // Compress if enabled
    // Delete from main table after archive
}
```

#### 3.2.4 Real-Time Log Streaming

Implement WebSocket-based log streaming for live monitoring:

```go
// internal/adminapi/websocket/logs.go
// LogStreamer streams logs in real-time to connected clients
type LogStreamer struct {
    hub      *LogHub
    filters  LogFilter
    buffer   chan *LogEntry
    bufferSize int
}
```

---

## 4. User Data and Privacy Management

### 4.1 Current State Analysis

The current system stores user data in [`domain.RadiusUser`](internal/domain/radius_user_getters.go) with:
- Username, password, status
- Expiration time
- Profile association
- Basic timestamps

### 4.2 Proposed Enhancements

#### 4.2.1 Data Privacy Controls

```go
// internal/adminapi/privacy.go
// PrivacyManager handles user data privacy operations
type PrivacyManager struct {
    db *gorm.DB
}

// UserDataExport exports user data in standard format
type UserDataExport struct {
    UserID        string    `json:"user_id"`
    Username      string    `json:"username"`
    Profile       string    `json:"profile"`
    Sessions      []SessionSummary `json:"sessions"`
    Payments      []PaymentSummary `json:"payments"`
    DataRetention DataRetentionInfo `json:"data_retention"`
    ExportTime    time.Time `json:"export_time"`
}

// DataRetentionPolicy defines data retention rules
type DataRetentionPolicy struct {
    ID              int64  `json:"id"`
    Name            string `json:"name"`
    SessionDataDays int    `json:"session_data_days"`
    AccountingDays  int    `json:"accounting_days"`
    LogRetentionDays int   `json:"log_retention_days"`
    AnonymizeAfter  int    `json:"anonymize_after_days"`
    DeleteAfter     int    `json:"delete_after_days"`
}

// AnonymizeUserData anonymizes user PII
func (m *PrivacyManager) AnonymizeUserData(username string) error {
    // Anonymize username in logs
    // Hash MAC addresses
    // Remove personal identifiers
    // Keep aggregate data
}
```

#### 4.2.2 GDPR Compliance Features

```go
// Data Subject Request handling
type DataSubjectRequest struct {
    ID          string    `json:"id"`
    Type        string    `json:"type"` // access, rectification, erasure, portability
    UserID      int64     `json:"user_id"`
    Status      string    `json:"status"` // pending, processing, completed, rejected
    RequestedAt time.Time `json:"requested_at"`
    CompletedAt time.Time `json:"completed_at"`
    Notes       string    `json:"notes"`
}
```

#### 4.2.3 User Consent Management

```go
// UserConsent tracks user consent for data processing
type UserConsent struct {
    ID          int64     `json:"id"`
    UserID      int64     `json:"user_id" gorm:"index"`
    Purpose     string    `json:"purpose"` // marketing, analytics, third_party_sharing
    ConsentGiven bool     `json:"consent_given"`
    ConsentDate time.Time `json:"consent_date"`
    WithdrawalDate time.Time `json:"withdrawal_date"`
    IPAddress   string    `json:"ip_address"`
    UserAgent   string    `json:"user_agent"`
}
```

#### 4.2.4 Data Access Auditing

```go
// DataAccessLog tracks who accessed user data
type DataAccessLog struct {
    ID          int64     `json:"id"`
    OperatorID  int64     `json:"operator_id"`
    UserID      int64     `json:"user_id"`
    DataType    string    `json:"data_type"` // user, session, voucher, financial
    AccessType  string    `json:"access_type"` // view, export, modify, delete
    Timestamp   time.Time `json:"timestamp"`
    IPAddress   string    `json:"ip_address"`
    Reason      string    `json:"reason"`
}
```

---

## 5. Database Backup and Restore

### 5.1 Current State Analysis

The system has a backup directory configured ([`config.GetBackupDir()`](config/config.go:234)) but lacks automated backup/restore functionality.

### 5.2 Proposed Enhancements

#### 5.2.1 Local Backup System

```go
// internal/app/backup/local.go
// LocalBackupManager handles local database backups
type LocalBackupManager struct {
    db        *gorm.DB
    backupDir string
    config    *BackupConfig
}

type BackupConfig struct {
    MaxLocalBackups     int           // Maximum local backups to retain
    BackupSchedule      string        // Cron expression
    CompressBackups     bool          // Compress with gzip
    EncryptBackups      bool          // Encrypt with AES
    EncryptionKey       string        // Key for encryption (from config)
    IncludeAttachments  bool          // Include uploaded files
}

type BackupMetadata struct {
    ID            int64     `json:"id"`
    Filename      string    `json:"filename"`
    SizeBytes     int64     `json:"size_bytes"`
    CreatedAt     time.Time `json:"created_at"`
    DatabaseType  string    `json:"database_type"` // sqlite, postgres
    Version       string    `json:"version"`
    Checksum      string    `json:"checksum"`
    IsCompressed  bool      `json:"is_compressed"`
    IsEncrypted   bool      `json:"is_encrypted"`
}

// CreateBackup creates a new database backup
func (m *LocalBackupManager) CreateBackup() (*BackupMetadata, error) {
    // SQLite: Backup using backup API or file copy
    // PostgreSQL: Use pg_dump
    // Create metadata file
    // Verify checksum
    // Return backup info
}

// ListBackups returns available backups
func (m *LocalBackupManager) ListBackups() ([]BackupMetadata, error)

// RestoreBackup restores database from backup
func (m *LocalBackupManager) RestoreBackup(backupID int64) error
```

#### 5.2.2 Google Drive Backup Integration

```go
// internal/app/backup/gdrive.go
// GoogleDriveBackupManager handles backups to Google Drive
type GoogleDriveBackupManager struct {
    localMgr    *LocalBackupManager
    driveService *drive.Service
    config      *GDriveConfig
}

type GDriveConfig struct {
    Enabled         bool   `json:"enabled"`
    CredentialsJSON string `json:"credentials_json"` // OAuth2 credentials
    TokenJSON       string `json:"token_json"` // OAuth2 token
    FolderID        string `json:"folder_id"` // Specific folder ID
    SyncSchedule    string `json:"sync_schedule"` // Cron expression
    KeepRemoteCount int    `json:"keep_remote_count"` // Max backups in GDrive
}

// UploadBackup uploads backup to Google Drive
func (m *GoogleDriveBackupManager) UploadBackup(backupID int64) error {
    // Read local backup file
    // Upload to Google Drive
    // Set file description with metadata
    // Update sync status
}

// DownloadBackup downloads backup from Google Drive
func (m *GoogleDriveBackupManager) DownloadBackup(remoteID, localPath string) error

// ListRemoteBackups lists backups in Google Drive
func (m *GoogleDriveBackupManager) ListRemoteBackups() ([]GDriveBackupFile, error)

// SyncBackups synchronizes local and remote backups
func (m *GoogleDriveBackupManager) SyncBackups() error
```

#### 5.2.3 Backup Scheduling

Add to [`internal/app/jobs.go`](internal/app/jobs.go):

```go
func (a *Application) initBackupJobs() {
    // Daily backup at 2 AM
    a.sched.AddFunc("0 2 * * *", func() {
        a.LocalBackupManager.CreateBackup()
    })
    
    // Sync to Google Drive daily at 3 AM
    a.sched.AddFunc("0 3 * * *", func() {
        if a.GDriveConfig.Enabled {
            a.GoogleDriveBackupManager.SyncBackups()
        }
    })
    
    // Cleanup old local backups weekly
    a.sched.AddFunc("0 4 * * 0", func() {
        a.LocalBackupManager.CleanupOldBackups()
    })
}
```

#### 5.2.4 Restore Workflow

```go
// internal/adminapi/backup.go
// BackupAPI provides backup management endpoints
type BackupAPI struct {
    localMgr  *LocalBackupManager
    gdriveMgr *GoogleDriveBackupManager
}

// RestoreRequest defines restore parameters
type RestoreRequest struct {
    BackupID    string `json:"backup_id"` // Local ID or GDrive file ID
    Source      string `json:"source"` // "local" or "gdrive"
    TargetDB    string `json:"target_db"` // "sqlite" or "postgres"
    Force       bool   `json:"force"` // Force restore without confirmation
}

// API Endpoints:
// POST /api/v1/admin/backup - Create backup
// GET /api/v1/admin/backup - List backups
// GET /api/v1/admin/backup/:id - Get backup info
// POST /api/v1/admin/backup/:id/restore - Restore from backup
// DELETE /api/v1/admin/backup/:id - Delete backup
// POST /api/v1/admin/backup/gdrive/sync - Sync with GDrive
```

---

## 6. Remote Tunnel for Maintenance

### 6.1 Proposed Implementation

#### 6.1.1 Built-in Reverse Tunnel

```go
// internal/app/tunnel/tunnel.go
// TunnelService provides remote access via reverse tunnel
type TunnelService struct {
    config    *TunnelConfig
    serverURL string
    client    *tunnelClient
    sessions  map[string]*TunnelSession
}

type TunnelConfig struct {
    Enabled       bool   `json:"enabled"`
    ServerURL     string `json:"server_url"` // Tunnel server (e.g., tunnel.example.com)
    AuthToken     string `json:"auth_token"`
    TunnelID      string `json:"tunnel_id"` // Unique identifier for this instance
    PortForward   []PortForward `json:"port_forward"` // Local ports to expose
    HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

type PortForward struct {
    LocalPort  int    `json:"local_port"`
    RemotePort int    `json:"remote_port"` // Port on tunnel server
    Protocol   string `json:"protocol"` // tcp, udp
}

// StartTunnel establishes reverse tunnel connection
func (s *TunnelService) StartTunnel() error {
    // Connect to tunnel server
    // Register this instance
    // Start heartbeat
    // Setup port forwards for:
    // - Web admin interface (1816)
    // - RADIUS auth (1812)
    // - RADIUS acct (1813)
}

// TunnelSession represents an active tunnel session
type TunnelSession struct {
    SessionID  string
    RemoteAddr string
    Established time.Time
    LastActivity time.Time
    BytesIn    int64
    BytesOut   int64
}
```

#### 6.1.2 Alternative: Cloudflare Tunnel Integration

```go
// For deployment with Cloudflare Tunnel:
// 1. Use cloudflared binary as subprocess
// 2. Or implement tunnel protocol directly

type CloudflareTunnelConfig struct {
    AccountID  string `json:"account_id"`
    TunnelID   string `json:"tunnel_id"`
    CredentialsPath string `json:"credentials_path"`
}

// Config in toughradius.yml:
// tunnel:
//   enabled: true
//   type: cloudflare  # or "built-in"
//   cloudflare:
//     tunnel_id: $TUNNEL_ID
//     credentials_path: /etc/cloudflared/credentials.json
```

#### 6.1.3 Admin API for Tunnel Management

```go
// internal/adminapi/tunnel.go
// Tunnel management endpoints
type TunnelStatus struct {
    Connected    bool      `json:"connected"`
    TunnelID     string    `json:"tunnel_id"`
    ServerURL    string    `json:"server_url"`
    Uptime       time.Duration `json:"uptime"`
    LastHeartbeat time.Time `json:"last_heartbeat"`
    ActiveSessions int      `json:"active_sessions"`
    PortForwards  []PortForwardStatus `json:"port_forwards"`
}

// API Endpoints:
// GET /api/v1/admin/tunnel/status - Get tunnel status
// POST /api/v1/admin/tunnel/connect - Connect tunnel
// POST /api/v1/admin/tunnel/disconnect - Disconnect tunnel
// GET /api/v1/admin/tunnel/sessions - Active tunnel sessions
```

---

## 7. System Maintenance and Active Sessions

### 7.1 Current State Analysis

The system lacks:
- Graceful shutdown handling for active sessions
- Maintenance mode configuration
- User notification system

### 7.2 Proposed Enhancements

#### 7.2.1 Maintenance Mode

```go
// internal/app/maintenance/maintenance.go
// MaintenanceManager controls system maintenance mode
type MaintenanceManager struct {
    db      *gorm.DB
    config  *MaintenanceConfig
    state   *MaintenanceState
    notify  *NotificationService
}

type MaintenanceConfig struct {
    AllowUserLogin      bool        `json:"allow_user_login"`
    AllowRadiusAuth     bool        `json:"allow_radius_auth"`
    AllowRadiusAcct     bool        `json:"allow_radius_acct"`
    AllowAdminAccess    bool        `json:"allow_admin_access"`
    ScheduledStart      *time.Time  `json:"scheduled_start"`
    ScheduledEnd        *time.Time  `json:"scheduled_end"`
    NotificationTemplate string    `json:"notification_template"`
}

type MaintenanceState struct {
    Enabled         bool      `json:"enabled"`
    StartTime       time.Time `json:"start_time"`
    EndTime         time.Time `json:"end_time"`
    Reason          string    `json:"reason"`
    InitiatedBy     string    `json:"initiated_by"`
    AffectedServices []string `json:"affected_services"`
}

// EnableMaintenance enables maintenance mode
func (m *MaintenanceManager) EnableMaintenance(config *MaintenanceConfig) error {
    // Save configuration
    // Update system state
    // Send notifications to affected users
    // Begin session draining if configured
}

// DisableMaintenance disables maintenance mode
func (m *MaintenanceManager) DisableMaintenance() error
```

#### 7.2.2 Active Session Handling During Maintenance

```go
// SessionDraining manages active sessions during maintenance
type SessionDraining struct {
    Strategy    string  // "immediate", "graceful", "notify_only"
    GracePeriod int     // seconds to wait before forcing disconnect
    NotifyUsers bool    // Send notification before disconnect
}

func (m *MaintenanceManager) DrainSessions(strategy SessionDraining) {
    activeSessions, _ := m.GetActiveSessions()
    
    switch strategy.Strategy {
    case "immediate":
        // Force disconnect all sessions
        for _, session := range activeSessions {
            m.DisconnectSession(session)
        }
        
    case "graceful":
        // 1. Stop accepting new sessions
        // 2. Wait for existing sessions to end naturally (with timeout)
        // 3. Force disconnect remaining sessions
        // 4. Send periodic warnings to users
        
    case "notify_only":
        // Just notify users, keep sessions active
        for _, session := range activeSessions {
            m.notify.SendMaintenanceWarning(session.Username, m.state)
        }
    }
}
```

#### 7.2.3 User Notification System

```go
// internal/app/notifications/notification.go
// NotificationService sends notifications to users
type NotificationService struct {
    channels map[string]NotificationChannel
}

type NotificationChannel interface {
    Send(recipient, message string) error
}

type NotificationRequest struct {
    Type        string   `json:"type"` // maintenance, expiry, usage_alert
    Recipients  []string `json:"recipients"` // usernames or "all_active"
    Subject     string   `json:"subject"`
    Message     string   `json:"message"`
    Schedule    *time.Time `json:"schedule"` // Send at specific time
    Priority    string   `json:"priority"` // low, normal, high, urgent
}

// Notification channels to implement:
// - Portal notification (in-app)
// - RADIUS Attribute 77 (ConnectInfo)
// - SMS (via gateway)
// - Email (via SMTP)
```

#### 7.2.4 Graceful Shutdown

```go
// internal/app/app.go - Modify Shutdown function
func (a *Application) Shutdown(ctx context.Context) error {
    // 1. Enable maintenance mode (block new auth)
    a.MaintenanceManager.EnableMaintenance(&MaintenanceConfig{
        AllowRadiusAuth: false,
        Reason:          "System shutdown",
    })
    
    // 2. Wait for active sessions to drain (max 5 minutes)
    if err := a.DrainActiveSessions(5 * time.Minute); err != nil {
        a.Logger.Warn("Session drain timeout, forcing disconnect")
    }
    
    // 3. Stop accepting new connections
    a.radiusServer.Stop()
    a.webServer.Stop()
    
    // 4. Flush pending accounting records
    a.FlushAccountingBuffer()
    
    // 5. Close database connections
    a.db.Close()
    
    return nil
}
```

---

## 8. Database Operations Optimization

### 8.1 Current State Analysis

Current cleanup in [`internal/app/jobs.go`](internal/app/jobs.go):
- System logs older than 1 year are deleted daily
- Accounting logs cleaned based on `AccountingHistoryDays` config (default 90)
- Expired online sessions cleaned (5 minutes timeout)

### 8.2 Proposed Enhancements

#### 8.2.1 Intelligent Data Lifecycle Management

```go
// internal/app/dblifecycle/manager.go
// DataLifecycleManager implements tiered data retention
type DataLifecycleManager struct {
    db     *gorm.DB
    config *LifecycleConfig
}

type LifecycleConfig struct {
    // Tier 1: Hot data (last 7 days) - full detail, fast queries
    HotRetentionDays int `json:"hot_retention_days"`
    
    // Tier 2: Warm data (8-90 days) - aggregated, normal queries
    WarmRetentionDays int `json:"warm_retention_days"`
    
    // Tier 3: Cold data (91-365 days) - summarized, slow queries OK
    ColdRetentionDays int `json:"cold_retention_days"`
    
    // Tier 4: Archive (365+ days) - retained in separate archive table
    ArchiveRetentionDays int `json:"archive_retention_days"`
    
    // Auto-vacuum settings
    VacuumSchedule string `json:"vacuum_schedule"`
    AnalyzeSchedule string `json:"analyze_schedule"`
}

// Implement data tiering:
// - Hot: radius_online, recent radius_accounting
// - Warm: aggregated daily stats
// - Cold: monthly summaries
// - Archive: separate table for compliance
```

#### 8.2.2 Database Index Optimization

```go
// internal/app/indexes/index_manager.go
// IndexManager maintains optimal database indexes
type IndexManager struct {
    db *gormDB
}

// Recommended indexes for ToughRADIUS:
// radius_accounting:
//   - (username, acct_start_time)
//   - (nas_addr, acct_start_time)
//   - (acct_session_id)
//   - (acct_stop_time) - for cleanup queries
//
// radius_online:
//   - (username)
//   - (nas_addr)
//   - (framed_ipaddr)
//   - (mac_addr)
//   - (last_update) - for cleanup
//
// radius_user:
//   - (username) UNIQUE
//   - (status)
//   - (expire_time)
//   - (profile_id)
```

#### 8.2.3 Query Optimization

```go
// Implement query result caching for expensive operations
type QueryCache struct {
    cache *cache.Cache
    ttl   time.Duration
}

// Cache expensive dashboard queries
func (m *DashboardStats) GetCachedStats() (*DashboardStats, error) {
    cacheKey := "dashboard:stats"
    
    if cached, found := m.cache.Get(cacheKey); found {
        return cached.(*DashboardStats), nil
    }
    
    stats := m.computeStats()
    m.cache.Set(cacheKey, stats, m.ttl)
    return stats, nil
}
```

#### 8.2.4 Database Size Management

```go
// internal/app/dbsize/manager.go
// DatabaseSizeManager monitors and controls database size
type DatabaseSizeManager struct {
    db        *gorm.DB
    thresholds *SizeThresholds
}

type SizeThresholds struct {
    WarningGB     float64 `json:"warning_gb"`     // Alert at this size
    CriticalGB    float64 `json:"critical_gb"`    // Force cleanup at this size
    MaxSizeGB     float64 `json:"max_size_gb"`    // Reject new data at this size
    TargetSizeGB  float64 `json:"target_size_gb"` // Target after cleanup
}

// Scheduled cleanup job
func (m *DatabaseSizeManager) ScheduledCleanup() {
    size := m.GetDatabaseSize()
    
    if size > m.thresholds.CriticalGB {
        m.ForceCleanup() // Aggressive cleanup
    } else if size > m.thresholds.WarningGB {
        m.StandardCleanup() // Normal cleanup
    }
}

// Functions to reduce database size:
// 1. VACUUM (SQLite) / VACUUM ANALYZE (PostgreSQL)
// 2. Delete old accounting records
// 3. Archive old system logs
// 4. Compress old session data
// 5. Remove unnecessary indexes
```

#### 8.2.5 Connection Pool Tuning

The current config ([`internal/app/database.go`](internal/app/database.go)) sets:
- SQLite: 1 connection (appropriate for single-user)
- PostgreSQL: MaxConn=100, IdleConn=10 (configurable)

Recommended enhancements:
```go
// Add to config
type DBOptimizationConfig struct {
    // Connection pool
    MaxConnLifetime time.Duration `json:"max_conn_lifetime"` // Default: 30min
    MaxConnReuse   int           `json:"max_conn_reuse"`   // Default: 1000
    
    // Query optimization
    SlowQueryThreshold time.Duration `json:"slow_query_threshold"` // Default: 1s
    EnableQueryCache   bool          `json:"enable_query_cache"`
    
    // Write optimization
    BatchSize         int  `json:"batch_size"`          // Default: 100
    EnableWAL         bool `json:"enable_wal"`          // PostgreSQL only
}
```

---

## 9. Implementation Priority Matrix

| Feature | Complexity | Impact | Priority |
|---------|------------|--------|----------|
| Database backup (local) | Medium | High | P0 |
| Maintenance mode | Medium | High | P0 |
| Active session handling | High | High | P0 |
| Enhanced dashboard metrics | Medium | High | P0 |
| Log management system | Medium | Medium | P1 |
| Google Drive backup | Medium | Medium | P1 |
| User data privacy | High | High | P1 |
| Real-time WebSocket metrics | High | Medium | P2 |
| Predictive analytics | High | Medium | P2 |
| Remote tunnel | High | Medium | P2 |
| Database optimization | Medium | High | P1 |

---

## 10. Summary

This enhancement proposal transforms ToughRADIUS from a basic RADIUS server into a comprehensive ISP management platform. The key focus areas are:

1. **Real-time monitoring** with WebSocket-based live metrics and predictive analytics
2. **Automated voucher lifecycle management** with smart expiration handling
3. **Centralized logging** with archival and retention policies
4. **Privacy compliance** with GDPR features and data access auditing
5. **Robust backup** to local storage and Google Drive
6. **Remote access** via built-in tunnel or Cloudflare integration
7. **Zero-downtime maintenance** with graceful session draining
8. **Database efficiency** through lifecycle management and optimization

Each feature builds on the existing codebase architecture and can be implemented incrementally following the priority matrix.
