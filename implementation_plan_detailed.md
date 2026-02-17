# ToughRADIUS Feature Enhancement - Implementation Plan

This document provides a detailed implementation plan that builds upon the existing codebase. All existing functionality is preserved - we only add new features.

---

## Part 1: Already Implemented Features

Based on code analysis of the existing codebase:

### 1.1 Voucher Management (✅ Fully Implemented)
- [x] Voucher batch creation with configurable expiration (fixed or first-use)
- [x] PIN-protected vouchers
- [x] Voucher extension ([`VoucherExtendRequest`](internal/adminapi/vouchers.go:393))
- [x] Voucher top-up ([`VoucherTopupRequest`](internal/adminapi/vouchers.go:398))
- [x] Auto-renewal subscriptions ([`VoucherSubscriptionRequest`](internal/adminapi/vouchers.go:412))
- [x] Voucher transfer between agents ([`TransferVouchers`](internal/adminapi/vouchers.go:732))
- [x] Bulk activate/deactivate vouchers ([`BulkActivateVouchers`](internal/adminapi/vouchers.go:646), [`BulkDeactivateVouchers`](internal/adminapi/vouchers.go:655))
- [x] Soft delete and restore vouchers ([`DeleteVoucherBatch`](internal/adminapi/vouchers.go:685), [`RestoreVoucherBatch`](internal/adminapi/vouchers.go:702))
- [x] Voucher bundles ([`VoucherBundle`](internal/domain/voucher.go:97))
- [x] Public voucher status check ([`PublicVoucherStatus`](internal/adminapi/vouchers.go:1357))
- [x] Financial reports ([`GetFinancialReport`](internal/adminapi/financial.go:75))
- [x] Agent stats with wallet, batches, transactions ([`GetAgentStats`](internal/adminapi/stats.go:39))

### 1.2 Sessions Management (✅ Partially Implemented)
- [x] Online session listing with filters ([`ListOnlineSessions`](internal/adminapi/sessions.go:74))
- [x] Force disconnect via CoA ([`DisconnectSession`](internal/adminapi/sessions.go:206))
- [x] Session data: username, NAS, IP, MAC, session time

### 1.3 Dashboard (✅ Partially Implemented)
- [x] Basic stats: total users, online users, disabled, expired ([`GetDashboardStats`](internal/adminapi/dashboard.go:63))
- [x] Authentication trend (7 days) ([`fetchAuthTrend`](internal/adminapi/dashboard.go:115))
- [x] Traffic stats (24-hour) ([`fetchTrafficStats`](internal/adminapi/dashboard.go:148))
- [x] Profile distribution pie chart ([`fetchProfileDistribution`](internal/adminapi/dashboard.go:189))

### 1.4 System Logs (✅ Partially Implemented)
- [x] Operation logs ([`SysOprLog`](internal/domain/system.go))
- [x] Filtering by operator, action, keyword ([`ListSystemLogs`](internal/adminapi/system_log.go:25))
- [x] Admin-only access

### 1.5 Background Jobs (✅ Partially Implemented)
- [x] System monitor (CPU, memory) every 30 seconds ([`SchedSystemMonitorTask`](internal/app/jobs.go:170))
- [x] Process monitor every 30 seconds ([`SchedProcessMonitorTask`](internal/app/jobs.go:191))
- [x] Daily log cleanup (older than 1 year)
- [x] Subscription renewal every minute ([`SchedSubscriptionRenewalTask`](internal/app/jobs.go:54))
- [x] Accounting history cleanup ([`SchedClearExpireData`](internal/app/jobs.go:217))
- [x] Expired online session cleanup (5 minutes)

### 1.6 Database (✅ Partially Implemented)
- [x] SQLite and PostgreSQL support
- [x] Connection pooling for PostgreSQL
- [x] Backup directory exists ([`GetBackupDir`](config/config.go:234))

---

## Part 2: Implementation Plan for Remaining Features

### Priority 1: Database Backup System

#### 1.1 Local Backup Implementation

**Files to Create:**
- `internal/app/backup/local_backup.go`

**Files to Modify:**
- `internal/app/app.go` - Add backup manager initialization
- `config/config.go` - Add backup configuration
- `internal/app/jobs.go` - Add scheduled backup jobs

**Implementation:**

```go
// internal/app/backup/local_backup.go
package backup

import (
    "compress/gzip"
    "crypto/cipher"
    "crypto/aes"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/talkincode/toughradius/v9/config"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// LocalBackupManager handles local database backups
type LocalBackupManager struct {
    db        *gorm.DB
    backupDir string
    config    *BackupConfig
}

// BackupConfig holds backup configuration
type BackupConfig struct {
    MaxLocalBackups    int  // Maximum local backups to retain
    CompressBackups    bool // Compress with gzip
    EncryptBackups     bool // Encrypt with AES
    EncryptionKey      string // Key for encryption (from config)
}

// BackupMetadata represents backup metadata
type BackupMetadata struct {
    ID           int64     `json:"id"`
    Filename     string    `json:"filename"`
    SizeBytes    int64     `json:"size_bytes"`
    CreatedAt    time.Time `json:"created_at"`
    DatabaseType string    `json:"database_type"`
    Version      string    `json:"version"`
    Checksum     string    `json:"checksum"`
    IsCompressed bool      `json:"is_compressed"`
    IsEncrypted  bool      `json:"is_encrypted"`
}

// NewLocalBackupManager creates a new backup manager
func NewLocalBackupManager(db *gorm.DB, backupDir string, cfg *BackupConfig) *LocalBackupManager {
    return &LocalBackupManager{
        db:        db,
        backupDir: backupDir,
        config:    cfg,
    }
}

// CreateBackup creates a new database backup
func (m *LocalBackupManager) CreateBackup() (*BackupMetadata, error) {
    timestamp := time.Now().Format("20060102150405")
    filename := fmt.Sprintf("toughradius_backup_%s.sqlite", timestamp)
    filepath := filepath.Join(m.backupDir, filename)

    // Get database type
    dbType := "sqlite" // Determine from db
    
    // Create backup based on database type
    if err := m.createSQLiteBackup(filepath); err != nil {
        return nil, err
    }

    // Get file info
    fi, err := os.Stat(filepath)
    if err != nil {
        return nil, err
    }

    metadata := &BackupMetadata{
        Filename:     filename,
        SizeBytes:    fi.Size(),
        CreatedAt:    time.Now(),
        DatabaseType: dbType,
        Version:      "9.0.0",
    }

    // Compress if enabled
    if m.config.CompressBackups {
        if err := m.compressBackup(filepath); err != nil {
            return nil, err
        }
        metadata.Filename += ".gz"
        metadata.IsCompressed = true
    }

    // Encrypt if enabled
    if m.config.EncryptBackups && m.config.EncryptionKey != "" {
        if err := m.encryptBackup(filepath, m.config.EncryptionKey); err != nil {
            return nil, err
        }
        metadata.Filename += ".enc"
        metadata.IsEncrypted = true
    }

    // Calculate checksum
    checksum, err := m.calculateChecksum(filepath)
    if err != nil {
        return nil, err
    }
    metadata.Checksum = checksum

    // Save metadata to database
    m.saveMetadata(metadata)

    // Cleanup old backups
    m.cleanupOldBackups()

    zap.S().Infof("Backup created: %s (%.2f MB)", metadata.Filename, float64(metadata.SizeBytes)/1024/1024)

    return metadata, nil
}

// createSQLiteBackup creates a backup of SQLite database
func (m *LocalBackupManager) createSQLiteBackup(filepath string) error {
    // Use SQLite backup API or copy file
    // This is a simplified version
    sourceDB := m.db.Config().ConnPool
    // Implementation depends on database type
    return nil
}

// compressBackup compresses a backup file using gzip
func (m *LocalBackupManager) compressBackup(filepath string) error {
    // Open original file
    original, err := os.Open(filepath)
    if err != nil {
        return err
    }
    defer original.Close()

    // Create compressed file
    compressed, err := os.Create(filepath + ".gz")
    if err != nil {
        return err
    }
    defer compressed.Close()

    // Create gzip writer
    writer := gzip.NewWriter(compressed)
    defer writer.Close()

    // Copy content
    _, err = io.Copy(writer, original)
    return err
}

// encryptBackup encrypts a backup file using AES
func (m *LocalBackupManager) encryptBackup(filepath, key string) error {
    // Implementation for AES encryption
    return nil
}

// calculateChecksum calculates SHA256 checksum
func (m *LocalBackupManager) calculateChecksum(filepath string) (string, error) {
    // Implementation
    return "", nil
}

// saveMetadata saves backup metadata to database
func (m *LocalBackupManager) saveMetadata(metadata *BackupMetadata) error {
    // Save to database
    return nil
}

// cleanupOldBackups removes old backups exceeding MaxLocalBackups
func (m *LocalBackupManager) cleanupOldBackups() {
    // Implementation
}

// ListBackups returns available backups
func (m *LocalBackupManager) ListBackups() ([]BackupMetadata, error) {
    // Implementation
    return nil, nil
}

// RestoreBackup restores database from backup
func (m *LocalBackupManager) RestoreBackup(backupID int64) error {
    // Implementation
    return nil
}
```

**Test File to Create:**
- `internal/app/backup/local_backup_test.go`

#### 1.2 Google Drive Backup Implementation

**Files to Create:**
- `internal/app/backup/gdrive_backup.go`

**Implementation:**
```go
// internal/app/backup/gdrive_backup.go
package backup

import (
    "context"
    "fmt"
    "io"
    "os"
    "time"

    "github.com/google/google-api-go-client/drive/v3"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "go.uber.org/zap"
)

// GoogleDriveBackupManager handles backups to Google Drive
type GoogleDriveBackupManager struct {
    localMgr    *LocalBackupManager
    driveService *drive.Service
    config      *GDriveConfig
}

type GDriveConfig struct {
    Enabled         bool   `json:"enabled"`
    CredentialsJSON string `json:"credentials_json"`
    TokenJSON       string `json:"token_json"`
    FolderID        string `json:"folder_id"`
    SyncSchedule    string `json:"sync_schedule"`
    KeepRemoteCount int    `json:"keep_remote_count"`
}

// NewGoogleDriveBackupManager creates a new Google Drive backup manager
func NewGoogleDriveBackupManager(localMgr *LocalBackupManager, cfg *GDriveConfig) (*GoogleDriveBackupManager, error) {
    // Initialize OAuth2 client
    ctx := context.Background()
    
    config, err := google.ConfigFromJSON([]byte(cfg.CredentialsJSON), drive.DriveScope)
    if err != nil {
        return nil, err
    }

    client := getOAuth2Client(ctx, config, cfg.TokenJSON)
    service, err := drive.New(client)
    if err != nil {
        return nil, err
    }

    return &GoogleDriveBackupManager{
        localMgr:    localMgr,
        driveService: service,
        config:      cfg,
    }, nil
}

func getOAuth2Client(ctx context.Context, config *oauth2.Config, tokenJSON string) *http.Client {
    // Implementation
    return nil
}

// UploadBackup uploads backup to Google Drive
func (m *GoogleDriveBackupManager) UploadBackup(backupID int64) error {
    // Implementation
    return nil
}

// DownloadBackup downloads backup from Google Drive
func (m *GoogleDriveBackupManager) DownloadBackup(remoteID, localPath string) error {
    // Implementation
    return nil
}

// ListRemoteBackups lists backups in Google Drive
func (m *GoogleDriveBackupManager) ListRemoteBackups() ([]GDriveBackupFile, error) {
    // Implementation
    return nil, nil
}

// SyncBackups synchronizes local and remote backups
func (m *GoogleDriveBackupManager) SyncBackups() error {
    // Implementation
    return nil
}

type GDriveBackupFile struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Size      int64     `json:"size"`
    CreatedAt time.Time `json:"created_time"`
}
```

**Test File to Create:**
- `internal/app/backup/gdrive_backup_test.go`

---

### Priority 2: Enhanced Dashboard with Real-Time Metrics

#### 2.1 WebSocket Server for Real-Time Updates

**Files to Create:**
- `internal/adminapi/websocket/hub.go`
- `internal/adminapi/websocket/client.go`
- `internal/adminapi/websocket/metrics.go`

**Files to Modify:**
- `internal/adminapi/adminapi.go` - Register WebSocket routes
- `web/src/hooks/useWebSocket.ts` - New frontend hook

**Implementation:**

```go
// internal/adminapi/websocket/hub.go
package websocket

import (
    "encoding/json"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
    // Registered clients
    clients map[*Client]bool

    // Register requests from the clients
    register chan *Client

    // Unregister requests from clients
    unregister chan *Client

    // Broadcast messages to clients
    broadcast chan []byte

    // Mutex for concurrent access
    mu sync.RWMutex
}

// NewHub creates a new hub
func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan []byte, 256),
    }
}

// Run starts the hub
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

// BroadcastMessage broadcasts a message to all clients
func (h *Hub) BroadcastMessage(msgType string, data interface{}) {
    msg := WSMessage{
        Type:      msgType,
        Timestamp: time.Now(),
        Data:      data,
    }

    jsonData, err := json.Marshal(msg)
    if err != nil {
        return
    }

    h.broadcast <- jsonData
}

// Client represents a WebSocket client
type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

// WSMessage represents a WebSocket message
type WSMessage struct {
    Type      string      `json:"type"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
    return &Client{
        hub:  hub,
        conn: conn,
        send: make(chan []byte, 256),
    }
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(512)
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }

        // Handle incoming messages (e.g., subscribe to specific metrics)
        var msg WSMessage
        if err := json.Unmarshal(message, &msg); err == nil {
            // Process subscription requests
        }
    }
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

```go
// internal/adminapi/websocket/metrics.go
package websocket

import (
    "sync"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// MetricsBroadcaster broadcasts real-time metrics to connected clients
type MetricsBroadcaster struct {
    hub     *Hub
    db      *gorm.DB
    interval time.Duration
    stopCh   chan struct{}
    wg       sync.WaitGroup
}

// NetworkMetrics represents real-time network metrics
type NetworkMetrics struct {
    Timestamp       time.Time       `json:"timestamp"`
    ActiveSessions  int64           `json:"active_sessions"`
    TotalThroughput Throughput      `json:"total_throughput"`
    NASMetrics      []NASMetrics    `json:"nas_metrics"`
    ActiveAlerts   []Alert         `json:"active_alerts"`
}

// Throughput represents bandwidth metrics
type Throughput struct {
    UploadMbps    float64 `json:"upload_mbps"`
    DownloadMbps float64 `json:"download_mbps"`
}

// NASMetrics represents per-NAS metrics
type NASMetrics struct {
    NASIP          string  `json:"nas_ip"`
    NASName        string  `json:"nas_name"`
    OnlineUsers    int64   `json:"online_users"`
    UploadMbps     float64 `json:"upload_mbps"`
    DownloadMbps   float64 `json:"download_mbps"`
    LatencyMs      float64 `json:"latency_ms"`
    Status         string  `json:"status"` // online, degraded, offline
}

// Alert represents a system alert
type Alert struct {
    ID          string    `json:"id"`
    Severity    string    `json:"severity"` // critical, warning, info
    Message     string    `json:"message"`
    Timestamp   time.Time `json:"timestamp"`
    NASIP       string    `json:"nas_ip,omitempty"`
}

// NewMetricsBroadcaster creates a new metrics broadcaster
func NewMetricsBroadcaster(hub *Hub, db *gorm.DB, interval time.Duration) *MetricsBroadcaster {
    return &MetricsBroadcaster{
        hub:      hub,
        db:       db,
        interval: interval,
        stopCh:   make(chan struct{}),
    }
}

// Start starts broadcasting metrics
func (m *MetricsBroadcaster) Start() {
    m.wg.Add(1)
    go func() {
        defer m.wg.Done()
        m.broadcastLoop()
    }()
}

// Stop stops broadcasting metrics
func (m *MetricsBroadcaster) Stop() {
    close(m.stopCh)
    m.wg.Wait()
}

func (m *MetricsBroadcaster) broadcastLoop() {
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            metrics := m.collectMetrics()
            m.hub.BroadcastMessage("metrics", metrics)

        case <-m.stopCh:
            return
        }
    }
}

func (m *MetricsBroadcaster) collectMetrics() *NetworkMetrics {
    metrics := &NetworkMetrics{
        Timestamp:      time.Now(),
        ActiveSessions: 0,
    }

    // Get active sessions count
    m.db.Model(&domain.RadiusOnline{}).Count(&metrics.ActiveSessions)

    // Get traffic stats for last minute
    var traffic struct {
        TotalInput  int64
        TotalOutput int64
    }
    oneMinuteAgo := time.Now().Add(-time.Minute)
    m.db.Model(&domain.RadiusAccounting{}).
        Select("COALESCE(SUM(acct_input_total), 0) as total_input, COALESCE(SUM(acct_output_total), 0) as total_output").
        Where("acct_start_time >= ?", oneMinuteAgo).
        Scan(&traffic)

    // Convert to Mbps (assuming 1 minute = 60 seconds)
    metrics.TotalThroughput.UploadMbps = float64(traffic.TotalInput) * 8 / 60 / 1000000
    metrics.TotalThroughput.DownloadMbps = float64(traffic.TotalOutput) * 8 / 60 / 1000000

    // Get NAS metrics
    var nasList []domain.NetNas
    m.db.Find(&nasList)

    for _, nas := range nasList {
        var onlineCount int64
        m.db.Model(&domain.RadiusOnline{}).Where("nas_addr = ?", nas.Ipaddr).Count(&onlineCount)

        nasMetrics := NASMetrics{
            NASIP:       nas.Ipaddr,
            NASName:     nas.Name,
            OnlineUsers: onlineCount,
            Status:      "online", // Would need actual health check
        }
        metrics.NASMetrics = append(metrics.NASMetrics, nasMetrics)
    }

    // Check for alerts
    metrics.ActiveAlerts = m.checkAlerts()

    return metrics
}

func (m *MetricsBroadcaster) checkAlerts() []Alert {
    var alerts []Alert

    // Check for high session count
    var sessionCount int64
    m.db.Model(&domain.RadiusOnline{}).Count(&sessionCount)
    if sessionCount > 10000 {
        alerts = append(alerts, Alert{
            ID:        "high_sessions",
            Severity:  "warning",
            Message:   "High number of active sessions",
            Timestamp: time.Now(),
        })
    }

    // Check for critical alerts
    // Add more alert conditions as needed

    return alerts
}
```

**Test Files to Create:**
- `internal/adminapi/websocket/hub_test.go`
- `internal/adminapi/websocket/metrics_test.go`

---

### Priority 3: Maintenance Mode System

#### 3.1 Maintenance Mode Implementation

**Files to Create:**
- `internal/app/maintenance/manager.go`

**Files to Modify:**
- `internal/app/app.go` - Add maintenance manager
- `internal/adminapi/adminapi.go` - Add maintenance API routes
- `internal/domain/tables.go` - Add maintenance config table

**Implementation:**

```go
// internal/app/maintenance/manager.go
package maintenance

import (
    "sync"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// MaintenanceManager controls system maintenance mode
type MaintenanceManager struct {
    db      *gorm.DB
    config  *MaintenanceConfig
    state   *MaintenanceState
    mu      sync.RWMutex
    notify  NotificationService
}

// MaintenanceConfig holds maintenance configuration
type MaintenanceConfig struct {
    AllowUserLogin   bool          `json:"allow_user_login"`
    AllowRadiusAuth  bool          `json:"allow_radius_auth"`
    AllowRadiusAcct  bool          `json:"allow_radius_acct"`
    AllowAdminAccess bool          `json:"allow_admin_access"`
    ScheduledStart   *time.Time    `json:"scheduled_start"`
    ScheduledEnd     *time.Time    `json:"scheduled_end"`
    NotifyTemplate   string        `json:"notification_template"`
    GracePeriod      time.Duration `json:"grace_period"`
}

// MaintenanceState holds current maintenance state
type MaintenanceState struct {
    Enabled          bool       `json:"enabled"`
    StartTime        time.Time  `json:"start_time"`
    EndTime          time.Time  `json:"end_time"`
    Reason           string     `json:"reason"`
    InitiatedBy      string     `json:"initiated_by"`
    AffectedServices []string   `json:"affected_services"`
    DrainingSessions bool      `json:"draining_sessions"`
}

// NotificationService sends notifications
type NotificationService interface {
    SendNotification(recipient string, message string) error
    BroadcastToActiveUsers(message string) error
}

// NewMaintenanceManager creates a new maintenance manager
func NewMaintenanceManager(db *gorm.DB, config *MaintenanceConfig) *MaintenanceManager {
    return &MaintenanceManager{
        db:     db,
        config: config,
        state:  &MaintenanceState{},
    }
}

// EnableMaintenance enables maintenance mode
func (m *MaintenanceManager) EnableMaintenance(reason, initiatedBy string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.state.Enabled = true
    m.state.StartTime = time.Now()
    m.state.Reason = reason
    m.state.InitiatedBy = initiatedBy
    m.state.AffectedServices = m.getAffectedServices()

    // Save state to database
    if err := m.saveState(); err != nil {
        return err
    }

    // Notify active users if configured
    if m.config.NotifyTemplate != "" {
        m.notify.BroadcastToActiveUsers(m.config.NotifyTemplate)
    }

    zap.S().Infof("Maintenance mode enabled: %s", reason)

    return nil
}

// DisableMaintenance disables maintenance mode
func (m *MaintenanceManager) DisableMaintenance() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.state.Enabled = false
    m.state.EndTime = time.Now()

    if err := m.saveState(); err != nil {
        return err
    }

    zap.S().Info("Maintenance mode disabled")

    return nil
}

// IsMaintenanceMode checks if system is in maintenance mode
func (m *MaintenanceManager) IsMaintenanceMode() bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.state.Enabled
}

// GetState returns current maintenance state
func (m *MaintenanceManager) GetState() *MaintenanceState {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.state
}

// GetConfig returns maintenance configuration
func (m *MaintenanceManager) GetConfig() *MaintenanceConfig {
    return m.config
}

// UpdateConfig updates maintenance configuration
func (m *MaintenanceManager) UpdateConfig(config *MaintenanceConfig) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.config = config

    // Save to database
    return m.saveConfig()
}

// CanAcceptAuth checks if authentication is allowed
func (m *MaintenanceManager) CanAcceptAuth() bool {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if !m.state.Enabled {
        return true
    }
    return m.config.AllowRadiusAuth
}

// CanAcceptAccounting checks if accounting is allowed
func (m *MaintenanceManager) CanAcceptAccounting() bool {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if !m.state.Enabled {
        return true
    }
    return m.config.AllowRadiusAcct
}

// CanUserLogin checks if user login is allowed
func (m *MaintenanceManager) CanUserLogin() bool {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if !m.state.Enabled {
        return true
    }
    return m.config.AllowUserLogin
}

// DrainSessions drains active sessions based on strategy
func (m *MaintenanceManager) DrainSessions(strategy SessionDrainingStrategy) error {
    m.mu.Lock()
    m.state.DrainingSessions = true
    m.mu.Unlock()

    defer func() {
        m.mu.Lock()
        m.state.DrainingSessions = false
        m.mu.Unlock()
    }()

    switch strategy.Strategy {
    case "immediate":
        return m.drainImmediate()
    case "graceful":
        return m.drainGraceful(strategy.GracePeriod)
    case "notify_only":
        return m.notifyUsers()
    }

    return nil
}

type SessionDrainingStrategy struct {
    Strategy    string
    GracePeriod time.Duration
    NotifyUsers bool
}

func (m *MaintenanceManager) getAffectedServices() []string {
    var services []string

    if !m.config.AllowRadiusAuth {
        services = append(services, "radius_auth")
    }
    if !m.config.AllowRadiusAcct {
        services = append(services, "radius_accounting")
    }
    if !m.config.AllowUserLogin {
        services = append(services, "user_portal")
    }
    if !m.config.AllowAdminAccess {
        services = append(services, "admin_panel")
    }

    return services
}

func (m *MaintenanceManager) saveState() error {
    // Save to sys_config or dedicated table
    return nil
}

func (m *MaintenanceManager) saveConfig() error {
    // Save to sys_config
    return nil
}

func (m *MaintenanceManager) drainImmediate() error {
    // Get all online sessions
    var sessions []domain.RadiusOnline
    m.db.Find(&sessions)

    for _, session := range sessions {
        // Send CoA disconnect request
        // This would integrate with the existing DisconnectSession function
        zap.S().Infof("Force disconnect session: %s", session.Username)
        m.db.Delete(&session)
    }

    return nil
}

func (m *MaintenanceManager) drainGraceful(timeout time.Duration) error {
    // Stop accepting new sessions
    // Wait for existing sessions to end naturally
    // Send warnings at intervals

    ticker := time.NewTicker(1 * time.Minute)
    deadline := time.Now().Add(timeout)

    for {
        select {
        case <-ticker.C:
            if time.Now().After(deadline) {
                // Force disconnect remaining
                return m.drainImmediate()
            }

            // Check remaining sessions
            var count int64
            m.db.Model(&domain.RadiusOnline{}).Count(&count)

            if count == 0 {
                return nil
            }

            // Send warning to remaining users
            m.notifyUsers()

        default:
            // Check if all sessions ended
            var count int64
            m.db.Model(&domain.RadiusOnline{}).Count(&count)
            if count == 0 {
                return nil
            }
            time.Sleep(1 * time.Second)
        }
    }
}

func (m *MaintenanceManager) notifyUsers() error {
    // Get active sessions
    var sessions []domain.RadiusOnline
    m.db.Find(&sessions)

    for _, session := range sessions {
        // Send notification (would integrate with notification service)
        zap.S().Infof("Notifying user %s about maintenance", session.Username)
    }

    return nil
}
```

**Test File to Create:**
- `internal/app/maintenance/manager_test.go`

---

### Priority 4: Enhanced Logging System

#### 4.1 Session Log Implementation

**Files to Create:**
- `internal/domain/session_log.go` - New domain model
- `internal/adminapi/session_log.go` - API handlers
- `internal/app/log_archival.go` - Log archival

**Files to Modify:**
- `internal/domain/tables.go` - Add new tables

**Implementation:**

```go
// internal/domain/session_log.go
package domain

import (
    "time"
)

// SessionLog records detailed session lifecycle events
type SessionLog struct {
    ID              int64     `json:"id" gorm:"primaryKey"`
    SessionID       string    `json:"session_id" gorm:"index"`
    Username        string    `json:"username" gorm:"index"`
    NASIP           string    `json:"nas_ip"`
    NASName         string    `json:"nas_name"`
    FramedIP        string    `json:"framed_ip"`
    MACAddress      string    `json:"mac_address"`
    EventType       string    `json:"event_type"` // start, stop, update, disconnect, timeout
    EventTime       time.Time `json:"event_time"`
    InputBytes      int64     `json:"input_bytes"`
    OutputBytes     int64     `json:"output_bytes"`
    Duration        int64     `json:"duration_seconds"`
    TerminateCause  string    `json:"terminate_cause"`
    IPChange        bool      `json:"ip_change"`
    DataCapHit      bool      `json:"data_cap_hit"`
    TimeCapHit      bool      `json:"time_cap_hit"`
    CreatedAt       time.Time `json:"created_at"`
}

func (SessionLog) TableName() string {
    return "radius_session_log"
}

// SystemLogArchive represents archived system logs
type SystemLogArchive struct {
    ID          int64     `json:"id" gorm:"primaryKey"`
    LogDate     time.Time `json:"log_date" gorm:"index"`
    Level       string    `json:"level"`
    Category    string    `json:"category"`
    Message     string    `json:"message"`
    Operator    string    `json:"operator"`
    IPAddress   string    `json:"ip_address"`
    Details     string    `json:"details"`
    CreatedAt   time.Time `json:"created_at"`
}

func (SystemLogArchive) TableName() string {
    return "sys_log_archive"
}
```

```go
// internal/adminapi/session_log.go
package adminapi

import (
    "net/http"
    "strconv"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/domain"
    "github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerSessionLogRoutes() {
    webserver.ApiGET("/session-logs", ListSessionLogs)
    webserver.ApiGET("/session-logs/:id", GetSessionLog)
}

// ListSessionLogs retrieves session logs
// @Summary get session logs
// @Tags Session
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param username query string false "Username"
// @Param session_id query string false "Session ID"
// @Param event_type query string false "Event type"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {object} ListResponse
// @Router /api/v1/session-logs [get]
func ListSessionLogs(c echo.Context) error {
    db := GetDB(c)

    page, _ := strconv.Atoi(c.QueryParam("page"))
    perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
    if page < 1 {
        page = 1
    }
    if perPage < 1 || perPage > 100 {
        perPage = 10
    }

    var total int64
    var logs []domain.SessionLog

    query := db.Model(&domain.SessionLog{})

    if username := c.QueryParam("username"); username != "" {
        query = query.Where("username LIKE ?", "%"+username+"%")
    }

    if sessionID := c.QueryParam("session_id"); sessionID != "" {
        query = query.Where("session_id = ?", sessionID)
    }

    if eventType := c.QueryParam("event_type"); eventType != "" {
        query = query.Where("event_type = ?", eventType)
    }

    if startTime := c.QueryParam("start_time"); startTime != "" {
        if t, err := time.Parse(time.RFC3339, startTime); err == nil {
            query = query.Where("event_time >= ?", t)
        }
    }

    if endTime := c.QueryParam("end_time"); endTime != "" {
        if t, err := time.Parse(time.RFC3339, endTime); err == nil {
            query = query.Where("event_time <= ?", t)
        }
    }

    query.Count(&total)

    offset := (page - 1) * perPage
    query.Order("id DESC").Limit(perPage).Offset(offset).Find(&logs)

    return paged(c, logs, total, page, perPage)
}

// GetSessionLog retrieves a single session log
func GetSessionLog(c echo.Context) error {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid session log ID", nil)
    }

    var logEntry domain.SessionLog
    if err := GetDB(c).First(&logEntry, id).Error; err != nil {
        return fail(c, http.StatusNotFound, "NOT_FOUND", "Session log not found", nil)
    }

    return ok(c, logEntry)
}
```

**Test Files to Create:**
- `internal/adminapi/session_log_test.go`

---

### Priority 5: Database Optimization

#### 5.1 Database Optimization Manager

**Files to Create:**
- `internal/app/database/optimizer.go`

**Files to Modify:**
- `config/config.go` - Add optimization config
- `internal/app/jobs.go` - Add optimization jobs

**Implementation:**

```go
// internal/app/database/optimizer.go
package database

import (
    "fmt"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// Optimizer handles database optimization
type Optimizer struct {
    db *gorm.DB
}

// NewOptimizer creates a new database optimizer
func NewOptimizer(db *gorm.DB) *Optimizer {
    return &Optimizer{db: db}
}

// Vacuum optimizes database storage (PostgreSQL/SQLite)
func (o *Optimizer) Vacuum() error {
    sqlDB, err := o.db.DB()
    if err != nil {
        return err
    }

    // For SQLite
    if o.db.Name() == "sqlite" {
        _, err = sqlDB.Exec("VACUUM")
        return err
    }

    // For PostgreSQL
    _, err = sqlDB.Exec("VACUUM ANALYZE")
    return err
}

// Analyze updates table statistics for query planner
func (o *Optimizer) Analyze() error {
    sqlDB, err := o.db.DB()
    if err != nil {
        return err
    }

    if o.db.Name() == "postgres" {
        _, err = sqlDB.Exec("ANALYZE")
        return err
    }

    return nil
}

// GetDatabaseSize returns current database size
func (o *Optimizer) GetDatabaseSize() (int64, error) {
    sqlDB, err := o.db.DB()
    if err != nil {
        return 0, err
    }

    if o.db.Name() == "sqlite" {
        var size int64
        err := sqlDB.QueryRow("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").Scan(&size)
        return size, err
    }

    // PostgreSQL
    var size int64
    err = sqlDB.QueryRow("SELECT pg_database_size(current_database())").Scan(&size)
    return size, err
}

// CleanupOldData removes old data based on retention policy
func (o *Optimizer) CleanupOldData(retentionDays int) error {
    cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

    // Delete old accounting records
    result := o.db.Where("acct_stop_time < ?", cutoffTime).Delete(&domain.RadiusAccounting{})
    if result.Error != nil {
        return result.Error
    }
    zap.S().Infof("Deleted %d old accounting records", result.RowsAffected)

    // Delete old session logs
    result = o.db.Where("event_time < ?", cutoffTime).Delete(&domain.SessionLog{})
    if result.Error != nil {
        return result.Error
    }
    zap.S().Infof("Deleted %d old session logs", result.RowsAffected)

    return nil
}

// GetTableStats returns statistics for all tables
func (o *Optimizer) GetTableStats() ([]TableStats, error) {
    var stats []TableStats

    tables := []string{
        "radius_user",
        "radius_online",
        "radius_accounting",
        "voucher",
        "voucher_batch",
        "sys_opr_log",
    }

    for _, table := range tables {
        var count int64
        if err := o.db.Table(table).Count(&count).Error; err != nil {
            continue
        }

        stats = append(stats, TableStats{
            TableName: table,
            RowCount:  count,
        })
    }

    return stats, nil
}

type TableStats struct {
    TableName string
    RowCount  int64
}

// CreateIndexes creates necessary indexes if they don't exist
func (o *Optimizer) CreateIndexes() error {
    indexes := []struct {
        Table  string
        Name   string
        Column string
    }{
        {"radius_accounting", "idx_acct_username_time", "username, acct_start_time"},
        {"radius_accounting", "idx_acct_nas_time", "nas_addr, acct_start_time"},
        {"radius_accounting", "idx_acct_stop_time", "acct_stop_time"},
        {"radius_online", "idx_online_username", "username"},
        {"radius_online", "idx_online_nas", "nas_addr"},
        {"radius_online", "idx_online_ip", "framed_ipaddr"},
        {"radius_online", "idx_online_mac", "mac_addr"},
        {"radius_online", "idx_online_update", "last_update"},
        {"voucher", "idx_voucher_status", "status"},
        {"voucher", "idx_voucher_expire", "expire_time"},
    }

    for _, idx := range indexes {
        query := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", idx.Name, idx.Table, idx.Column)
        if err := o.db.Exec(query).Error; err != nil {
            zap.S().Warnf("Failed to create index %s: %v", idx.Name, err)
        }
    }

    return nil
}
```

**Test File to Create:**
- `internal/app/database/optimizer_test.go`

---

## Summary: Implementation Priority Matrix

| Priority | Feature | Files to Create | Files to Modify | Estimated Effort |
|----------|---------|-----------------|-----------------|------------------|
| P0 | Local Backup | `internal/app/backup/local_backup.go` + tests | `internal/app/app.go`, `config/config.go`, `internal/app/jobs.go` | 3 days |
| P0 | Google Drive Backup | `internal/app/backup/gdrive_backup.go` + tests | - | 2 days |
| P1 | Real-Time Dashboard (WebSocket) | `internal/adminapi/websocket/*.go` + tests | `internal/adminapi/adminapi.go`, frontend | 5 days |
| P1 | Maintenance Mode | `internal/app/maintenance/manager.go` + tests | `internal/app/app.go`, `internal/adminapi/adminapi.go`, `internal/domain/tables.go` | 4 days |
| P2 | Session Logging | `internal/domain/session_log.go`, `internal/adminapi/session_log.go` + tests | `internal/domain/tables.go` | 3 days |
| P2 | Database Optimizer | `internal/app/database/optimizer.go` + tests | `config/config.go`, `internal/app/jobs.go` | 2 days |

## Compliance with Repo Standards

All implementations will follow:
1. **TDD Approach**: Write tests first in `*_test.go` files
2. **Code Style**: Match existing patterns in [`internal/adminapi/`](internal/adminapi/) and [`internal/app/`](internal/app/)
3. **Documentation**: Add comprehensive in-code comments following the "Code is Documentation" principle
4. **No Breaking Changes**: All existing APIs remain unchanged
5. **Pure Go**: No CGO dependencies
