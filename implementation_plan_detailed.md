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

---------------
## Voucher Management Enhancement

### 1.1 Voucher Batch Management
# Voucher authentication parity with Mikrotik: users vs vouchers in Radius

**Session ID:** ses_32ba5ffe4ffeSQZFEKuc2gb86a
**Created:** 3/9/2026, 11:46:34 PM
**Updated:** 3/10/2026, 12:30:29 AM

---

## User

please check the current implementation of voucher, why vouchers in the radius database are not authinticated in mikrotik router while the users created in user management tab are appear normally. I want the users /vouchers to be also authinticated in mikrotik like users , how we could better do it.

---

## Assistant (Plan · minimax-m2.5-free · 6.5s)

_Thinking:_

The user is asking me to investigate an issue with their ToughRADIUS implementation where vouchers are not being authenticated in MikroTik routers, while regular users work fine. They want vouchers to authenticate like regular users.

Let me start by understanding the codebase structure and then explore how authentication works for both users and vouchers.

I need to:
1. First invoke the brainstorming skill since this is a creative/investigative task
2. Then explore the codebase to understand:
   - How users are authenticated in MikroTik
   - How vouchers are supposed to work
   - Why vouchers might not be authenticating

Let me start with the skill and then explore.

I'll investigate this authentication issue. Let me first understand the codebase structure and then explore how both users and vouchers are handled for MikroTik authentication.

**Tool: skill**

**Input:**
```json
{
  "name": "brainstorming"
}
```

**Output:**
```
<skill_content name="brainstorming">
# Skill: brainstorming

# Brainstorming Ideas Into Designs

## Overview

Help turn ideas into fully formed designs and specs through natural collaborative dialogue.

Start by understanding the current project context, then ask questions one at a time to refine the idea. Once you understand what you're building, present the design and get user approval.

<HARD-GATE>
Do NOT invoke any implementation skill, write any code, scaffold any project, or take any implementation action until you have presented a design and the user has approved it. This applies to EVERY project regardless of perceived simplicity.
</HARD-GATE>

## Anti-Pattern: "This Is Too Simple To Need A Design"

Every project goes through this process. A todo list, a single-function utility, a config change — all of them. "Simple" projects are where unexamined assumptions cause the most wasted work. The design can be short (a few sentences for truly simple projects), but you MUST present it and get approval.

## Checklist

You MUST create a task for each of these items and complete them in order:

1. **Explore project context** — check files, docs, recent commits
2. **Ask clarifying questions** — one at a time, understand purpose/constraints/success criteria
3. **Propose 2-3 approaches** — with trade-offs and your recommendation
4. **Present design** — in sections scaled to their complexity, get user approval after each section
5. **Write design doc** — save to `docs/plans/YYYY-MM-DD-<topic>-design.md` and commit
6. **Transition to implementation** — invoke writing-plans skill to create implementation plan

## Process Flow

```dot
digraph brainstorming {
    "Explore project context" [shape=box];
    "Ask clarifying questions" [shape=box];
    "Propose 2-3 approaches" [shape=box];
    "Present design sections" [shape=box];
    "User approves design?" [shape=diamond];
    "Write design doc" [shape=box];
    "Invoke writing-plans skill" [shape=doublecircle];

    "Explore project context" -> "Ask clarifying questions";
    "Ask clarifying questions" -> "Propose 2-3 approaches";
    "Propose 2-3 approaches" -> "Present design sections";
    "Present design sections" -> "User approves design?";
    "User approves design?" -> "Present design sections" [label="no, revise"];
    "User approves design?" -> "Write design doc" [label="yes"];
    "Write design doc" -> "Invoke writing-plans skill";
}
```

**The terminal state is invoking writing-plans.** Do NOT invoke frontend-design, mcp-builder, or any other implementation skill. The ONLY skill you invoke after brainstorming is writing-plans.

## The Process

**Understanding the idea:**
- Check out the current project state first (files, docs, recent commits)
- Ask questions one at a time to refine the idea
- Prefer multiple choice questions when possible, but open-ended is fine too
- Only one question per message - if a topic needs more exploration, break it into multiple questions
- Focus on understanding: purpose, constraints, success criteria

**Exploring approaches:**
- Propose 2-3 different approaches with trade-offs
- Present options conversationally with your recommendation and reasoning
- Lead with your recommended option and explain why

**Presenting the design:**
- Once you believe you understand what you're building, present the design
- Scale each section to its complexity: a few sentences if straightforward, up to 200-300 words if nuanced
- Ask after each section whether it looks right so far
- Cover: architecture, components, data flow, error handling, testing
- Be ready to go back and clarify if something doesn't make sense

## After the Design

**Documentation:**
- Write the validated design to `docs/plans/YYYY-MM-DD-<topic>-design.md`
- Use elements-of-style:writing-clearly-and-concisely skill if available
- Commit the design document to git

**Implementation:**
- Invoke the writing-plans skill to create a detailed implementation plan
- Do NOT invoke any other skill. writing-plans is the next step.

## Key Principles

- **One question at a time** - Don't overwhelm with multiple questions
- **Multiple choice preferred** - Easier to answer than open-ended when possible
- **YAGNI ruthlessly** - Remove unnecessary features from all designs
- **Explore alternatives** - Always propose 2-3 approaches before settling
- **Incremental validation** - Present design, get approval before moving on
- **Be flexible** - Go back and clarify when something doesn't make sense

Base directory for this skill: file:///home/faris/.config/opencode/skills/superpowers/brainstorming
Relative paths in this skill (e.g., scripts/, reference/) are relative to this base directory.
Note: file list is sampled.

<skill_files>

</skill_files>
</skill_content>
```

---

## Assistant (Plan · minimax-m2.5-free · 6.4s)

_Thinking:_

Now let me explore the codebase to understand the project structure and how authentication works for both users and vouchers.

**Tool: glob**

**Input:**
```json
{
  "pattern": "**/*.py"
}
```

**Output:**
```
/home/faris/Downloads/toughradius/toughradius/go/src/runtime/runtime-gdb.py
/home/faris/Downloads/toughradius/toughradius/fix_performance.py
/home/faris/Downloads/toughradius/toughradius/fix_operators.py
/home/faris/Downloads/toughradius/toughradius/fix_logs.py
/home/faris/Downloads/toughradius/toughradius/fix_config.py
/home/faris/Downloads/toughradius/toughradius/fix_profiles.py
/home/faris/Downloads/toughradius/toughradius/fix_mobile_title.py
/home/faris/Downloads/toughradius/toughradius/fix_appbar.py
/home/faris/Downloads/toughradius/toughradius/fix_build.py
/home/faris/Downloads/toughradius/toughradius/enhance_cards.py
/home/faris/Downloads/toughradius/toughradius/update_lists.py
/home/faris/Downloads/toughradius/toughradius/fix_product_show.py
```

**Tool: grep**

**Input:**
```json
{
  "pattern": "voucher",
  "include": "*.py"
}
```

**Output:**
```
Found 39 matches
/home/faris/Downloads/toughradius/toughradius/fix_performance.py:
  Line 35:                                         <TableCell align="right">{agent.total_vouchers}</TableCell>
  Line 37:                                             <Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
  Line 40:                                             <Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
  Line 67:                                                     <Typography variant="body2" fontWeight="bold">{agent.total_vouchers}</Typography>
  Line 71:                                                     <Box><Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600, height: 20 }} /></Box>
  Line 75:                                                     <Box><Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600, height: 20 }} /></Box>
  Line 105:                                         <TableCell align="right">{agent.total_vouchers}</TableCell>
  Line 107:                                             <Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
  Line 110:                                             <Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
  Line 148:                                                 <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
  Line 151:                                                 <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>
  Line 188:                                                         <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
  Line 192:                                                         <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>
  Line 227:                                                 <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
  Line 230:                                                 <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>

/home/faris/Downloads/toughradius/toughradius/fix_build.py:
  Line 23: with open('web/src/resources/vouchers.tsx', 'r') as f:
  Line 24:     vouchers_content = f.read()
  Line 27: vouchers_content = vouchers_content.replace(
  Line 32: with open('web/src/resources/vouchers.tsx', 'w') as f:
  Line 33:     f.write(vouchers_content)

/home/faris/Downloads/toughradius/toughradius/enhance_cards.py:
  Line 177: def enhance_vouchers():
  Line 178:     with open('web/src/resources/vouchers.tsx', 'r') as f:
  Line 241:     new_voucher_grid = """
  Line 281:                                         <ReferenceField source="batch_id" reference="voucher-batches"><TextField source="name" /></ReferenceField>
  Line 318:     old_voucher_grid_pattern = r"const VoucherGrid = \(\) => \{.+?return \(.+?</Box>\s*\);\s*\};\s*"
  Line 319:     if re.search(old_voucher_grid_pattern, content, flags=re.DOTALL):
  Line 320:         content = re.sub(old_voucher_grid_pattern, new_voucher_grid.strip() + "\n", content, flags=re.DOTALL)
  Line 325:     with open('web/src/resources/vouchers.tsx', 'w') as f:
  Line 327:         print("Enhanced vouchers grids")
  Line 332: enhance_vouchers()

/home/faris/Downloads/toughradius/toughradius/update_lists.py:
  Line 196: def update_vouchers():
  Line 197:     with open('web/src/resources/vouchers.tsx', 'r') as f:
  Line 293:     voucher_grid = """
  Line 308:                                 Batch: <ReferenceField source="batch_id" reference="voucher-batches"><TextField source="name" /></ReferenceField>
  Line 331:         content = content.replace('export const VoucherList =', voucher_grid + '\nexport const VoucherList =')
  Line 338:             <ReferenceField source="batch_id" reference="voucher-batches">
  Line 361:                     <ReferenceField source="batch_id" reference="voucher-batches">
  Line 377:     with open('web/src/resources/vouchers.tsx', 'w') as f:
  Line 382: update_vouchers()
```

---

## Assistant (Plan · minimax-m2.5-free · 8.7s)

_Thinking:_

Now I can see there's a vouchers.tsx file in the web frontend. Let me explore more about the voucher backend and authentication logic. Let me look for more details about the voucher implementation and how authentication works.

**Tool: glob**

**Input:**
```json
{
  "pattern": "**/voucher*.py"
}
```

**Output:**
```
No files found
```

**Tool: grep**

**Input:**
```json
{
  "pattern": "radcheck|radauth",
  "include": "*.py"
}
```

**Output:**
```
No files found
```

---

## Assistant (Plan · minimax-m2.5-free · 3.5s)

_Thinking:_

Let me look for the database models or any backend code related to vouchers and authentication.

**Tool: glob**

**Input:**
```json
{
  "pattern": "server/**/*.py"
}
```

**Output:**
```
No files found
```

**Tool: grep**

**Input:**
```json
{
  "include": "*.py",
  "pattern": "class.*User|class.*Voucher"
}
```

**Output:**
```
No files found
```

---

## Assistant (Plan · minimax-m2.5-free · 3.6s)

_Thinking:_

Let me look at the project structure first to understand where the backend code is located.

Let me explore the project structure first:

**Tool: bash**

**Input:**
```json
{
  "command": "ls -la",
  "description": "List project root files"
}
```

**Output:**
```
total 267244
drwxrwxr-x 18 faris faris     4096 Mar  4 22:29 .
drwxrwxr-x  4 faris faris     4096 Feb 20 00:09 ..
-rw-rw-r--  1 faris faris    42491 Nov 26 17:46 AGENT.md
-rw-rw-r--  1 faris faris   243320 Mar  6 02:30 backend.log
-rw-rw-r--  1 faris faris     5671 Feb 17 16:36 capabilities_report.md
-rw-rw-r--  1 faris faris     8715 Feb 17 18:29 capabilities_summery.md
drwxrwxr-x  9 faris faris     4096 Nov 26 17:46 cmd
drwxrwxr-x  2 faris faris     4096 Nov 26 17:46 config
-rw-rw-r--  1 faris faris      131 Feb 17 16:36 cookies.txt
-rw-rw-r--  1 faris faris    23000 Feb 20 01:58 coverage.out
-rw-rw-r--  1 faris faris     1197 Nov 26 17:46 Dockerfile
drwxrwxr-x  3 faris faris     4096 Nov 26 17:46 docs
-rw-rw-r--  1 faris faris      279 Nov 26 17:46 .editorconfig
-rw-rw-r--  1 faris faris    19932 Feb 20 01:40 enhance_cards.py
-rw-rw-r--  1 faris faris     1632 Nov 26 17:46 .env.example
-rw-rw-r--  1 faris faris    33962 Feb 18 00:41 features_enhancement_proposal.md
-rw-rw-r--  1 faris faris     3071 Feb 17 19:10 features_roadmap.md
-rw-rw-r--  1 faris faris     5192 Feb 20 01:54 fix_appbar.py
-rw-rw-r--  1 faris faris     1022 Feb 20 01:41 fix_build.py
-rw-rw-r--  1 faris faris     3465 Feb 20 02:05 fix_config.py
-rw-rw-r--  1 faris faris     4882 Feb 20 02:08 fix_logs.py
-rw-rw-r--  1 faris faris     2790 Feb 20 01:59 fix_mobile_title.py
-rw-rw-r--  1 faris faris     7898 Feb 20 02:11 fix_operators.py
-rw-rw-r--  1 faris faris    18681 Feb 20 02:12 fix_performance.py
-rw-rw-r--  1 faris faris    20513 Feb 20 01:29 fix_product_show.py
-rw-rw-r--  1 faris faris     7987 Feb 20 02:04 fix_profiles.py
drwxrwxr-x  8 faris faris     4096 Mar  5 23:47 .git
-rw-rw-r--  1 faris faris       81 Nov 26 17:46 .gitattributes
drwxrwxr-x  2 faris faris     4096 Nov 26 17:46 .githooks
drwxrwxr-x  4 faris faris     4096 Nov 26 17:46 .github
-rw-rw-r--  1 faris faris      489 Nov 26 17:46 .gitignore
drwxrwxr-x  9 faris faris     4096 Mar  2 01:21 go
-rw-rw-r--  1 faris faris 68988925 Feb  7  2024 go1.22.0.linux-amd64.tar.gz
-rw-rw-r--  1 faris faris 78382844 Feb 17 16:36 go1.24.0.linux-amd64.tar.gz
-rw-rw-r--  1 faris faris 78382844 Feb 11  2025 go1.24.0.linux-amd64.tar.gz.1
-rw-rw-r--  1 faris faris     1101 Nov 26 17:46 .golangci.yml
-rw-rw-r--  1 faris faris     4451 Mar  5 23:47 go.mod
-rw-rw-r--  1 faris faris    25566 Mar  5 23:47 go.sum
-rw-rw-r--  1 faris faris    39973 Feb 18 00:53 implementation_plan_detailed.md
-rw-rw-r--  1 faris faris    31221 Feb 17 16:36 instruction.md
drwxrwxr-x 12 faris faris     4096 Feb 25 22:49 internal
-rw-rw-r--  1 faris faris     1073 Nov 26 17:46 LICENSE
-rw-rw-r--  1 faris faris     3387 Feb 25 21:45 main.go
-rw-rw-r--  1 faris faris     8597 Nov 26 17:46 Makefile
drwxrwxr-x  4 faris faris     4096 Feb 17 16:36 node_modules
-rw-rw-r--  1 faris faris       73 Feb 17 16:36 package.json
-rw-rw-r--  1 faris faris      623 Feb 17 16:36 package-lock.json
drwxrwxr-x 10 faris faris     4096 Nov 26 17:46 pkg
drwxrwxr-x  2 faris faris     4096 Mar  5 02:17 plans
-rw-rw-r--  1 faris faris     6829 Nov 26 17:46 README.md
drwxrwxr-x  2 faris faris     4096 Mar  4 22:30 release
drwxr-xr-x  7 faris faris     4096 Feb 16 15:50 rundata
drwxrwxr-x  2 faris faris     4096 Nov 26 17:46 scripts
-rw-rw-r--  1 faris faris      925 Nov 26 17:46 SECURITY.md
drwxrwxr-x  2 faris faris    12288 Nov 26 17:46 share
-rw-rw-r--  1 faris faris     6710 Mar  6 02:02 start_dev.md
-rwxrwxr-x  1 faris faris     3412 Feb 28 17:39 start_dev.sh
-rwxrwxr-x  1 faris faris     1288 Feb 17 16:36 start_production.sh
-rwxrwxr-x  1 faris faris 47094724 Mar  1 23:44 toughradius
-rw-rw-r--  1 faris faris      484 Feb 17 16:36 toughradius.dev.yml
-rw-rw-r--  1 faris faris    18279 Feb 20 01:36 update_lists.py
-rw-rw-r--  1 faris faris     3773 Feb 17 19:17 voucher_management_assessment.md
drwxrwxr-x  9 faris faris     4096 Mar  6 00:50 web

```

---

## Assistant (Plan · minimax-m2.5-free · 3.7s)

_Thinking:_

This is a Go project (toughradius). Let me explore the internal directory and find where the voucher and authentication logic is.

**Tool: glob**

**Input:**
```json
{
  "pattern": "internal/**/*.go"
}
```

**Output:**
```
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/servers_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/adminapi.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/servers.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/test_helpers.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/tablename_test.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/tables.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/server.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/server_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/agents.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/agent_hierarchy.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/agents_wallet_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/agent_hierarchy_test.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/agent_hierarchy.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/vouchers.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/websocket.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/analytics.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/voucher.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/invoice.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/users.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/invoices.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/radius.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/invoices_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/vouchers_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/settings.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/product.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/dashboard_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/cpe_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/cpe.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/discovery.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/campaigns_test.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/coa/client.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/coa/client_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/coa.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/coa/vendor.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/coa/types.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/dashboard.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/lifecycle.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/system_log.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/tunnel.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/system.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/topology_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/system_log_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/network_topology.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/privacy.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/maintenance.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/backup.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/privacy_test.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/network.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/products.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/profiles.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/errors/errors.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/sessions.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/vouchers_bulk_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/helpers.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/auth.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/finance.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/financial.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/stats.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/auth_plugin_runner_test.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendor_parse_test.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/errors/errors_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/nodes.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/responses.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/eap_helper_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/accounting.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/operators.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/auth_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/profiles_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/sessions_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/accounting_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/nas_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/operators_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/nodes_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/nas.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/users_test.go
/home/faris/Downloads/toughradius/toughradius/internal/adminapi/context.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radsec_server.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius_auth.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendor_parse.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/acct_plugin_runner.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/auth_pipeline.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/simple_cache.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/radius_user_getters_test.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/marshal_test.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/radius_user_getters.go
/home/faris/Downloads/toughradius/toughradius/internal/domain/marshal.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/ikuai/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/h3c/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/registry.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/registry_test.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/mikrotik/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/aruba/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/huawei/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/microsoft/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/hillstone/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/alcatel/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/cisco/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/pfSense/generated.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/codes.go
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/vendors/juniper/generated.go

(Results are truncated: showing first 100 results. Consider using a more specific path or pattern.)
```

---

## Assistant (Plan · minimax-m2.5-free · 3.5s)

_Thinking:_

I found the key files. Let me examine:
1. `internal/domain/voucher.go` - the voucher domain model
2. `internal/radiusd/radius_auth.go` - the RADIUS authentication logic
3. `internal/domain/radius.go` - RADIUS domain models

Let me also search for how users are authenticated vs how vouchers should be authenticated.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/domain/voucher.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/domain/voucher.go</path>
<type>file</type>
<content>1: package domain
2: 
3: import (
4: 	"time"
5: )
6: 
7: // VoucherBatch represents a batch of generated vouchers that can be sold
8: // to customers. Each batch is linked to a Product which defines the actual
9: // allocations (data quota, time validity) that vouchers will inherit.
10: //
11: // The batch controls printing/validity deadlines - vouchers cannot be printed
12: // or activated after PrintExpireTime. The actual allocations come from the
13: // Product at voucher creation time.
14: //
15: // Database table: voucher_batch
16: //
17: // Lifecycle:
18: //   - Created via Admin API POST /api/v1/voucher-batches
19: //   - Generates multiple Voucher records on creation
20: //   - Can be activated/deactivated in bulk
21: //   - Supports soft delete with restore capability
22: type VoucherBatch struct {
23: 	ID           int64      `json:"id,string" form:"id"`
24: 	Name         string     `json:"name" form:"name"`
25: 	ProductID    int64      `gorm:"index" json:"product_id,string" form:"product_id"`
26: 	AgentID      int64      `gorm:"index" json:"agent_id,string" form:"agent_id"`
27: 	Count        int        `json:"count" form:"count"`
28: 	Prefix       string     `json:"prefix" form:"prefix"`
29: 	Remark      string     `json:"remark" form:"remark"`
30: 	PrintExpireTime  *time.Time `json:"print_expire_time" gorm:"column:expire_time"`  // Deadline for printing/activating vouchers (maps to existing expire_time column)
31: 	GeneratePIN bool        `json:"generate_pin" form:"generate_pin"`
32: 	PINLength    int        `json:"pin_length" form:"pin_length"`
33: 	ExpirationType string    `json:"expiration_type" form:"expiration_type"` // "fixed" or "first_use"
34: 	ValidityDays   int       `json:"validity_days" form:"validity_days"`
35: 	IsDeleted      bool      `json:"is_deleted" gorm:"default:false"`
36: 	ActivatedAt   *time.Time `json:"activated_at" gorm:"index"`  // When the batch was activated (non-nil = activated)
37: 	CreatedAt    time.Time  `gorm:"index" json:"created_at"`
38: }
39: 
40: func (VoucherBatch) TableName() string {
41: 	return "voucher_batch"
42: }
43: 
44: // Voucher represents an individual prepaid access code that customers can
45: // redeem to get internet access. Each voucher is linked to a Product which
46: // defines the allocations (data quota, time validity) that this voucher provides.
47: //
48: // Vouchers track both their allocated quota and actual usage. Expiration can
49: // occur via either: (1) time-based ExpireTime reached, or (2) quota-based
50: // when DataUsed >= DataQuota or TimeUsed >= TimeQuota.
51: //
52: // Database table: voucher
53: //
54: // Lifecycle:
55: //   - Created when batch is generated (inherits allocations from Product)
56: //   - Status transitions: unused → active → used/expired
57: //   - Quota usage tracked via RADIUS accounting
58: //   - Soft-deleted after grace period when expired
59: type Voucher struct {
60: 	ID              int64     `json:"id,string" form:"id"`
61: 	BatchID         int64     `gorm:"index" json:"batch_id,string" form:"batch_id"`
62: 	Code            string    `json:"code" gorm:"uniqueIndex" form:"code"`
63: 	RadiusUsername  string    `gorm:"index" json:"radius_username" form:"radius_username"`
64: 	Status          string    `gorm:"index" json:"status" form:"status"` // unused, active, used, expired
65: 	AgentID         int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
66: 	Price           float64   `json:"price" form:"price"`
67: 	ActivatedAt     time.Time `json:"activated_at"`
68: 	ExpireTime      time.Time `gorm:"index" json:"expire_time"`
69: 
70: 	// Allocations from Product (set at voucher creation)
71: 	DataQuota int64 `json:"data_quota" form:"data_quota"` // MB (0 = unlimited)
72: 	TimeQuota int64 `json:"time_quota" form:"time_quota"` // seconds (0 = unlimited)
73: 
74: 	// Actual usage tracking
75: 	DataUsed int64 `json:"data_used" form:"data_used"` // MB used
76: 	TimeUsed int64 `json:"time_used" form:"time_used"` // seconds used
77: 
78: 	// Grace period tracking for quota-based expiration
79: 	QuotaExpiredAt *time.Time `json:"quota_expired_at"`
80: 
81: 	ExtendedCount   int       `json:"extended_count"`
82: 	LastExtendedAt  time.Time `json:"last_extended_at"`
83: 	FirstUsedAt     time.Time `json:"first_used_at"`
84: 	PIN             string    `json:"-"`                  // Hashed PIN (never exposed via JSON)
85: 	RequirePIN      bool      `json:"require_pin" form:"require_pin"`
86: 	IsDeleted       bool      `json:"is_deleted" gorm:"default:false"`
87: 	CreatedAt       time.Time `json:"created_at"`
88: 	UpdatedAt       time.Time `json:"updated_at"`
89: }
90: 
91: func (Voucher) TableName() string {
92: 	return "voucher"
93: }
94: 
95: // VoucherTopup represents additional data quota that can be added to an active voucher
96: type VoucherTopup struct {
97: 	ID              int64     `json:"id,string" form:"id"`
98: 	VoucherID       int64     `json:"voucher_id,string" form:"voucher_id"`
99: 	VoucherCode     string    `gorm:"index" json:"voucher_code" form:"voucher_code"`
100: 	AgentID         int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
101: 	DataQuota       int64     `json:"data_quota" form:"data_quota"`      // Additional data in MB
102: 	TimeQuota       int64     `json:"time_quota" form:"time_quota"`      // Additional time in seconds
103: 	Price           float64   `json:"price" form:"price"`                // Price paid for topup
104: 	Status          string    `gorm:"index" json:"status" form:"status"`               // active, used, expired
105: 	ActivatedAt     time.Time `json:"activated_at"`
106: 	ExpireTime      time.Time `json:"expire_time"`
107: 	IsDeleted       bool      `json:"is_deleted" gorm:"default:false"`
108: 	CreatedAt       time.Time `gorm:"index" json:"created_at"`
109: 
110: 	UpdatedAt       time.Time `json:"updated_at"`
111: }
112: 
113: func (VoucherTopup) TableName() string {
114: 	return "voucher_topup"
115: }
116: 
117: // VoucherSubscription represents a recurring subscription for automatic voucher renewal
118: type VoucherSubscription struct {
119: 	ID              int64     `json:"id,string" form:"id"`
120: 	VoucherCode     string    `gorm:"index" json:"voucher_code" form:"voucher_code"`
121: 	ProductID       int64     `json:"product_id,string" form:"product_id"`
122: 	AgentID         int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
123: 	IntervalDays    int       `json:"interval_days" form:"interval_days"` // Days between renewals
124: 	Status          string    `gorm:"index" json:"status" form:"status"`                 // active, paused, cancelled
125: 
126: 	AutoRenew       bool      `json:"auto_renew" form:"auto_renew"`        // Whether to auto-renew
127: 	LastRenewalAt   time.Time `json:"last_renewal_at"`
128: 	NextRenewalAt   time.Time `json:"next_renewal_at"`
129: 	RenewalCount    int       `json:"renewal_count"`                       // Number of times renewed
130: 	IsDeleted       bool      `json:"is_deleted" gorm:"default:false"`
131: 	CreatedAt       time.Time `json:"created_at"`
132: 	UpdatedAt       time.Time `json:"updated_at"`
133: }
134: 
135: func (VoucherSubscription) TableName() string {
136: 	return "voucher_subscription"
137: }
138: 
139: // VoucherBundle represents a package of multiple vouchers sold together
140: type VoucherBundle struct {
141: 	ID          int64     `json:"id,string" form:"id"`
142: 	Name        string    `json:"name" form:"name"`
143: 	Description string    `json:"description" form:"description"`
144: 	AgentID     int64     `gorm:"index" json:"agent_id,string" form:"agent_id"`
145: 	ProductID   int64     `json:"product_id,string" form:"product_id"`
146: 	VoucherCount int      `json:"voucher_count" form:"voucher_count"` // Number of vouchers in bundle
147: 	Price       float64   `json:"price" form:"price"`                  // Bundle price
148: 	Status      string    `gorm:"index" json:"status" form:"status"`                 // active, inactive
149: 
150: 	IsDeleted   bool      `json:"is_deleted" gorm:"default:false"`
151: 	CreatedAt   time.Time `json:"created_at"`
152: 	UpdatedAt   time.Time `json:"updated_at"`
153: }
154: 
155: func (VoucherBundle) TableName() string {
156: 	return "voucher_bundle"
157: }
158: 
159: // VoucherBundleItem represents a voucher included in a bundle
160: type VoucherBundleItem struct {
161: 	ID         int64     `json:"id,string" form:"id"`
162: 	BundleID   int64     `json:"bundle_id,string" form:"bundle_id"`
163: 	VoucherID  int64     `json:"voucher_id,string" form:"voucher_id"`
164: 	VoucherCode string   `json:"voucher_code" form:"voucher_code"`
165: 	Status     string    `json:"status" form:"status"` // available, sold, used
166: 	CreatedAt  time.Time `json:"created_at"`
167: }
168: 
169: func (VoucherBundleItem) TableName() string {
170: 	return "voucher_bundle_item"
171: }

(End of file - total 171 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 10.0s)

_Thinking:_

Now I see the voucher model. It has `RadiusUsername` field. Let me check how the authentication is handled - let me look at the radius_auth.go and also see how users are handled.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius_auth.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius_auth.go</path>
<type>file</type>
<content>1: package radiusd
2: 
3: import (
4: 	"context"
5: 
6: 	"github.com/talkincode/toughradius/v9/internal/app"
7: 	"github.com/talkincode/toughradius/v9/internal/domain"
8: 	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
9: 	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
10: 	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
11: 	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
12: 	"go.uber.org/zap"
13: 	"layeh.com/radius"
14: 	"layeh.com/radius/rfc2865"
15: )
16: 
17: type AuthService struct {
18: 	*RadiusService
19: 	eapHelper               *EAPAuthHelper
20: 	authPipeline            *AuthPipeline
21: 	allowedEAPHandlers      map[string]struct{}
22: 	allowedEAPHandlersOrder []string
23: }
24: 
25: func NewAuthService(radiusService *RadiusService) *AuthService {
26: 	authService := &AuthService{
27: 		RadiusService: radiusService,
28: 	}
29: 	allowed := authService.initAllowedEAPHandlers()
30: 	authService.eapHelper = NewEAPAuthHelper(radiusService, allowed)
31: 	authService.authPipeline = NewAuthPipeline()
32: 	authService.registerDefaultStages()
33: 	return authService
34: }
35: 
36: func (s *AuthService) ServeRADIUS(w radius.ResponseWriter, r *radius.Request) {
37: 	// Recover from unexpected panics only (programming errors)
38: 	// Normal errors should be handled via error return values
39: 	defer func() {
40: 		if ret := recover(); ret != nil {
41: 			var err error
42: 			switch v := ret.(type) {
43: 			case error:
44: 				err = v
45: 			case string:
46: 				err = radiuserrors.NewError(v)
47: 			default:
48: 				err = radiuserrors.NewError("unknown panic")
49: 			}
50: 			zap.L().Error("radius auth unexpected panic",
51: 				zap.Error(err),
52: 				zap.String("namespace", "radius"),
53: 				zap.String("metrics", app.MetricsRadiusAuthDrop),
54: 				zap.Stack("stacktrace"),
55: 			)
56: 			s.SendReject(w, r, err)
57: 		}
58: 	}()
59: 
60: 	if r == nil {
61: 		return
62: 	}
63: 
64: 	if s.Config().Radiusd.Debug {
65: 		zap.S().Info(FmtRequest(r))
66: 	}
67: 
68: 	s.ensurePipeline()
69: 	pipelineCtx := NewAuthPipelineContext(s, w, r)
70: 	defer func() {
71: 		if pipelineCtx != nil && pipelineCtx.RateLimitChecked && pipelineCtx.Username != "" {
72: 			s.ReleaseAuthRateLimit(pipelineCtx.Username)
73: 		}
74: 	}()
75: 
76: 	if err := s.authPipeline.Execute(pipelineCtx); err != nil {
77: 		// Process error through guards and log appropriately
78: 		finalErr := s.processAuthError("auth_pipeline", r, pipelineCtx.User, pipelineCtx.NAS,
79: 			pipelineCtx.VendorRequestForPlugin, pipelineCtx.IsMacAuth,
80: 			pipelineCtx.Username, pipelineCtx.RemoteIP, err)
81: 		if finalErr != nil {
82: 			s.logAndReject(w, r, finalErr)
83: 		}
84: 	}
85: }
86: 
87: // Pipeline exposes the underlying auth pipeline for customization.
88: func (s *AuthService) Pipeline() *AuthPipeline {
89: 	s.ensurePipeline()
90: 	return s.authPipeline
91: }
92: 
93: func (s *AuthService) SendAccept(w radius.ResponseWriter, r *radius.Request, resp *radius.Packet) {
94: 	if err := w.Write(resp); err != nil {
95: 		zap.L().Error("radius write accept error",
96: 			zap.String("namespace", "radius"),
97: 			zap.String("metrics", app.MetricsRadiusAuthDrop),
98: 			zap.Error(err),
99: 		)
100: 		return
101: 	}
102: 
103: 	if s.eapHelper != nil {
104: 		s.eapHelper.CleanupState(r)
105: 	}
106: 
107: 	if s.Config().Radiusd.Debug {
108: 		zap.S().Debug(FmtResponse(resp, r.RemoteAddr))
109: 	}
110: }
111: 
112: func (s *AuthService) SendReject(w radius.ResponseWriter, r *radius.Request, err error) {
113: 	var code = radius.CodeAccessReject
114: 	var resp = r.Response(code)
115: 	if err != nil {
116: 		msg := err.Error()
117: 		if len(msg) > 253 {
118: 			msg = msg[:253]
119: 		}
120: 		_ = rfc2865.ReplyMessage_SetString(resp, msg)
121: 	}
122: 
123: 	if writeErr := w.Write(resp); writeErr != nil {
124: 		zap.L().Error("radius write reject response error",
125: 			zap.String("namespace", "radius"),
126: 			zap.String("metrics", app.MetricsRadiusAuthDrop),
127: 			zap.Error(writeErr),
128: 		)
129: 	}
130: 
131: 	if s.eapHelper != nil {
132: 		s.eapHelper.CleanupState(r)
133: 	}
134: 
135: 	// debug message
136: 	if s.Config().Radiusd.Debug {
137: 		zap.S().Info(FmtResponse(resp, r.RemoteAddr))
138: 	}
139: }
140: 
141: // logAndReject logs the error with appropriate metrics and sends reject response.
142: func (s *AuthService) logAndReject(w radius.ResponseWriter, r *radius.Request, err error) {
143: 	metricsKey := app.MetricsRadiusAuthDrop
144: 	if radiusErr, ok := radiuserrors.GetRadiusError(err); ok {
145: 		metricsKey = radiusErr.MetricsKey()
146: 	}
147: 
148: 	zap.L().Error("radius auth error",
149: 		zap.Error(err),
150: 		zap.String("namespace", "radius"),
151: 		zap.String("metrics", metricsKey),
152: 	)
153: 
154: 	s.SendReject(w, r, err)
155: }
156: 
157: // processAuthError processes authentication errors through registered guards.
158: // It returns the final error after all guards have been consulted.
159: // This replaces the old handleAuthError which used panic for flow control.
160: //
161: // Parameters:
162: //   - stage: The pipeline stage where the error occurred
163: //   - r: The RADIUS request
164: //   - user: The user being authenticated (may be nil)
165: //   - nas: The NAS device (may be nil)
166: //   - vendorReq: Vendor-specific request data
167: //   - isMacAuth: Whether this is MAC authentication
168: //   - username: The username (for logging)
169: //   - nasip: The NAS IP (for logging)
170: //   - err: The original error
171: //
172: // Returns:
173: //   - error: The final error after guard processing, or nil if suppressed
174: func (s *AuthService) processAuthError(
175: 	stage string,
176: 	r *radius.Request,
177: 	user interface{},
178: 	nas *domain.NetNas,
179: 	vendorReq *vendorparsers.VendorRequest,
180: 	isMacAuth bool,
181: 	username string,
182: 	nasip string,
183: 	err error,
184: ) error {
185: 	if err == nil {
186: 		return nil
187: 	}
188: 
189: 	var radiusUser *domain.RadiusUser
190: 	if u, ok := user.(*domain.RadiusUser); ok {
191: 		radiusUser = u
192: 	}
193: 
194: 	metadata := map[string]interface{}{
195: 		"stage": stage,
196: 	}
197: 	if appCtx := s.AppContext(); appCtx != nil {
198: 		metadata["config_mgr"] = appCtx.ConfigMgr()
199: 		metadata["profile_cache"] = appCtx.ProfileCache()
200: 	}
201: 	if username != "" {
202: 		metadata["username"] = username
203: 	}
204: 	if nasip != "" {
205: 		metadata["nas_ip"] = nasip
206: 	}
207: 
208: 	authCtx := &auth.AuthContext{
209: 		Request:       r,
210: 		User:          radiusUser,
211: 		Nas:           nas,
212: 		VendorRequest: vendorReq,
213: 		IsMacAuth:     isMacAuth,
214: 		Metadata:      metadata,
215: 	}
216: 
217: 	ctx := context.Background()
218: 	currentErr := err
219: 
220: 	// Process through all guards
221: 	for _, guard := range registry.GetAuthGuards() {
222: 		// Try new interface first
223: 		if result := guard.OnAuthError(ctx, authCtx, stage, currentErr); result != nil {
224: 			switch result.Action {
225: 			case auth.GuardActionStop:
226: 				// Use the error from guard and stop processing
227: 				if result.Err != nil {
228: 					return result.Err
229: 				}
230: 				return currentErr
231: 			case auth.GuardActionSuppress:
232: 				// Error is suppressed, treat as success
233: 				return nil
234: 			case auth.GuardActionContinue:
235: 				// Update error if guard modified it
236: 				if result.Err != nil {
237: 					currentErr = result.Err
238: 				}
239: 				continue
240: 			}
241: 		}
242: 
243: 		// Fallback to old interface for backward compatibility
244: 		if guardErr := guard.OnError(ctx, authCtx, stage, currentErr); guardErr != nil {
245: 			// Old behavior: guard returns error means replace current error
246: 			currentErr = guardErr
247: 		}
248: 	}
249: 
250: 	return currentErr
251: }

(End of file - total 251 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 16.2s)

_Thinking:_

This is the main auth service. Let me search for how it processes users and vouchers - I need to find where it queries the database for user credentials.

**Tool: grep**

**Input:**
```json
{
  "include": "*.go",
  "pattern": "GetUser|QueryUser|FindUser"
}
```

**Output:**
```
Found 43 matches
/home/faris/Downloads/toughradius/toughradius/go/src/syscall/zsyscall_windows.go:
  Line 171: 	procGetUserNameExW                     = modsecur32.NewProc("GetUserNameExW")
  Line 174: 	procGetUserProfileDirectoryW           = moduserenv.NewProc("GetUserProfileDirectoryW")
  Line 1222: func GetUserNameEx(nameFormat uint32, nameBuffre *uint16, nSize *uint32) (err error) {
  Line 1223: 	r1, _, e1 := Syscall(procGetUserNameExW.Addr(), 3, uintptr(nameFormat), uintptr(unsafe.Pointer(nameBuffre)), uintptr(unsafe.Pointer(nSize)))
  Line 1247: func GetUserProfileDirectory(t Token, dir *uint16, dirLen *uint32) (err error) {
  Line 1248: 	r1, _, e1 := Syscall(procGetUserProfileDirectoryW.Addr(), 3, uintptr(t), uintptr(unsafe.Pointer(dir)), uintptr(unsafe.Pointer(dirLen)))

/home/faris/Downloads/toughradius/toughradius/go/src/syscall/security_windows.go:
  Line 35: //sys	GetUserNameEx(nameFormat uint32, nameBuffre *uint16, nSize *uint32) (err error) [failretval&0xff==0] = secur32.GetUserNameExW
  Line 292: //sys	GetUserProfileDirectory(t Token, dir *uint16, dirLen *uint32) (err error) = userenv.GetUserProfileDirectoryW
  Line 361: // GetUserProfileDirectory retrieves path to the
  Line 363: func (t Token) GetUserProfileDirectory() (string, error) {
  Line 367: 		e := GetUserProfileDirectory(t, &b[0], &n)

/home/faris/Downloads/toughradius/toughradius/go/src/os/user/lookup_windows.go:
  Line 278: 		dir, e := t.GetUserProfileDirectory()
  Line 282: 		username, e := windows.GetUserName(syscall.NameSamCompatible)
  Line 286: 		displayName, e := windows.GetUserName(syscall.NameDisplay)

/home/faris/Downloads/toughradius/toughradius/go/src/internal/syscall/windows/security_windows.go:
  Line 161: // GetUserName retrieves the user name of the current thread
  Line 163: func GetUserName(format uint32) (string, error) {
  Line 167: 		e := syscall.GetUserNameEx(format, &b[0], &n)

/home/faris/Downloads/toughradius/toughradius/go/src/cmd/vendor/golang.org/x/tools/internal/stdlib/manifest.go:
  Line 10202: 		{"(Token).GetUserProfileDirectory", Method, 0},
  Line 11949: 		{"GetUserNameEx", Func, 0},
  Line 11950: 		{"GetUserProfileDirectory", Func, 0},

/home/faris/Downloads/toughradius/toughradius/go/src/cmd/vendor/golang.org/x/sys/windows/zsyscall_windows.go:
  Line 308: 	procGetUserPreferredUILanguages                          = modkernel32.NewProc("GetUserPreferredUILanguages")
  Line 452: 	procGetUserNameExW                                       = modsecur32.NewProc("GetUserNameExW")
  Line 504: 	procGetUserProfileDirectoryW                             = moduserenv.NewProc("GetUserProfileDirectoryW")
  Line 544: 	procWTSQueryUserToken                                    = modwtsapi32.NewProc("WTSQueryUserToken")
  Line 2660: 	r1, _, e1 := syscall.Syscall6(procGetUserPreferredUILanguages.Addr(), 4, uintptr(flags), uintptr(unsafe.Pointer(numLanguages)), uintptr(unsafe.Pointer(buf)), uintptr(unsafe.Pointer(bufSize)), 0, 0)
  Line 3880: func GetUserNameEx(nameFormat uint32, nameBuffre *uint16, nSize *uint32) (err error) {
  Line 3881: 	r1, _, e1 := syscall.Syscall(procGetUserNameExW.Addr(), 3, uintptr(nameFormat), uintptr(unsafe.Pointer(nameBuffre)), uintptr(unsafe.Pointer(nSize)))
  Line 4289: func GetUserProfileDirectory(t Token, dir *uint16, dirLen *uint32) (err error) {
  Line 4290: 	r1, _, e1 := syscall.Syscall(procGetUserProfileDirectoryW.Addr(), 3, uintptr(t), uintptr(unsafe.Pointer(dir)), uintptr(unsafe.Pointer(dirLen)))
  Line 4680: func WTSQueryUserToken(session uint32, token *Token) (err error) {
  Line 4681: 	r1, _, e1 := syscall.Syscall(procWTSQueryUserToken.Addr(), 2, uintptr(session), uintptr(unsafe.Pointer(token)), 0)

/home/faris/Downloads/toughradius/toughradius/go/src/cmd/vendor/golang.org/x/sys/windows/syscall_windows.go:
  Line 423: //sys	getUserPreferredUILanguages(flags uint32, numLanguages *uint32, buf *uint16, bufSize *uint32) (err error) = kernel32.GetUserPreferredUILanguages
  Line 1614: // GetUserPreferredUILanguages retrieves information about the user preferred UI languages.
  Line 1615: func GetUserPreferredUILanguages(flags uint32) ([]string, error) {

/home/faris/Downloads/toughradius/toughradius/go/src/cmd/vendor/golang.org/x/sys/windows/security_windows.go:
  Line 28: //sys	GetUserNameEx(nameFormat uint32, nameBuffre *uint16, nSize *uint32) (err error) [failretval&0xff==0] = secur32.GetUserNameExW
  Line 638: //sys	GetUserProfileDirectory(t Token, dir *uint16, dirLen *uint32) (err error) = userenv.GetUserProfileDirectoryW
  Line 737: // GetUserProfileDirectory retrieves path to the
  Line 739: func (t Token) GetUserProfileDirectory() (string, error) {
  Line 743: 		e := GetUserProfileDirectory(t, &b[0], &n)
  Line 888: //sys WTSQueryUserToken(session uint32, token *Token) (err error) = wtsapi32.WTSQueryUserToken

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius.go:
  Line 182: // GetUserForAcct fetches the user without validating expiration or status
  Line 184: func (s *RadiusService) GetUserForAcct(username string) (user *domain.RadiusUser, err error) {

/home/faris/Downloads/toughradius/toughradius/internal/adminapi/users_test.go:
  Line 283: func TestGetUser(t *testing.T) {
```

---

## Assistant (Plan · minimax-m2.5-free · 5.0s)

_Thinking:_

Let me look at the RadiusService and the GetUser function in the radius.go file:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius.go</path>
<type>file</type>
<content>1: package radiusd
2: 
3: import (
4: 	"bytes"
5: 	"context"
6: 	"crypto/md5"
7: 	"errors"
8: 	"fmt"
9: 	"net"
10: 	"os"
11: 	"strconv"
12: 	"strings"
13: 	"sync"
14: 	"time"
15: 
16: 	"github.com/panjf2000/ants/v2"
17: 	"github.com/talkincode/toughradius/v9/config"
18: 	"github.com/talkincode/toughradius/v9/internal/app"
19: 	"github.com/talkincode/toughradius/v9/internal/domain"
20: 	cachepkg "github.com/talkincode/toughradius/v9/internal/radiusd/cache"
21: 	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
22: 	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
23: 	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
24: 	repogorm "github.com/talkincode/toughradius/v9/internal/radiusd/repository/gorm"
25: 	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
26: 	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
27: 	"github.com/talkincode/toughradius/v9/pkg/common"
28: 	"go.uber.org/zap"
29: 	"gorm.io/gorm"
30: 	"layeh.com/radius"
31: 	"layeh.com/radius/rfc2865"
32: 	"layeh.com/radius/rfc2866"
33: 	"layeh.com/radius/rfc2869"
34: 	"layeh.com/radius/rfc3162"
35: 	"layeh.com/radius/rfc4818"
36: 
37: 	// Import vendor parsers for auto-registration
38: 	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
39: 	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
40: 	_ "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers/parsers"
41: )
42: 
43: const (
44: 	RadiusRejectDelayTimes = 7
45: 	RadiusAuthRateInterval = 1 // Original: 1 second rate limit
46: )
47: 
48: type VendorRequest struct {
49: 	MacAddr string
50: 	Vlanid1 int64
51: 	Vlanid2 int64
52: }
53: 
54: type AuthRateUser struct {
55: 	Username  string
56: 	Starttime time.Time
57: }
58: 
59: type EapState struct {
60: 	Username  string
61: 	Challenge []byte
62: 	StateID   string
63: 	EapMethad string
64: 	Success   bool
65: }
66: 
67: type RadiusService struct {
68: 	appCtx        app.AppContext // Use interface instead of concrete type
69: 	AuthRateCache map[string]AuthRateUser
70: 	EapStateCache map[string]EapState
71: 	TaskPool      *ants.Pool
72: 	arclock       sync.Mutex
73: 	eaplock       sync.Mutex
74: 	nasCache      *cachepkg.TTLCache[*domain.NetNas]
75: 	userCache     *cachepkg.TTLCache[*domain.RadiusUser]
76: 
77: 	// New Repository Layer (v9 refactoring)
78: 	UserRepo       repository.UserRepository
79: 	SessionRepo    repository.SessionRepository
80: 	AccountingRepo repository.AccountingRepository
81: 	NasRepo        repository.NasRepository
82: 	VoucherRepo    repository.VoucherRepository
83: }
84: 
85: func NewRadiusService(appCtx app.AppContext) *RadiusService {
86: 	poolsize, err := strconv.Atoi(os.Getenv("TOUGHRADIUS_RADIUS_POOL"))
87: 	if err != nil {
88: 		poolsize = 1024
89: 	}
90: 	pool, err := ants.NewPool(poolsize)
91: 	common.Must(err)
92: 
93: 	// Initialize all repositories using injected context
94: 	db := appCtx.DB()
95: 	s := &RadiusService{
96: 		appCtx:        appCtx,
97: 		AuthRateCache: make(map[string]AuthRateUser),
98: 		EapStateCache: make(map[string]EapState),
99: 		arclock:       sync.Mutex{},
100: 		TaskPool:      pool,
101: 		nasCache:      cachepkg.NewTTLCache[*domain.NetNas](time.Minute, 512),
102: 		userCache:     cachepkg.NewTTLCache[*domain.RadiusUser](10*time.Second, 2048),
103: 		// Initialize repository layer
104: 		UserRepo:       repogorm.NewGormUserRepository(db),
105: 		SessionRepo:    repogorm.NewGormSessionRepository(db),
106: 		AccountingRepo: repogorm.NewGormAccountingRepository(db),
107: 		NasRepo:        repogorm.NewGormNasRepository(db),
108: 		VoucherRepo:    repogorm.NewGormVoucherRepository(db),
109: 	}
110: 
111: 	// Note: Plugin initialization is done externally after service creation
112: 	// to avoid circular dependency. Call plugins.InitPlugins() from main.go.
113: 
114: 	return s
115: }
116: 
117: func (s *RadiusService) RADIUSSecret(ctx context.Context, remoteAddr net.Addr) ([]byte, error) {
118: 	return []byte("mysecret"), nil
119: }
120: 
121: // GetNas looks up a NAS device, preferring IP before ID
122: // Deprecated: Use NasRepo.GetByIPOrIdentifier instead
123: func (s *RadiusService) GetNas(ip, identifier string) (nas *domain.NetNas, err error) {
124: 	cacheKey := fmt.Sprintf("%s|%s", ip, identifier)
125: 	if cached, ok := s.nasCache.Get(cacheKey); ok {
126: 		return cached, nil
127: 	}
128: 	// Adapter: delegate to repository layer
129: 	nas, err = s.NasRepo.GetByIPOrIdentifier(context.Background(), ip, identifier)
130: 	if err != nil {
131: 		if errors.Is(err, gorm.ErrRecordNotFound) {
132: 			return nil, radiuserrors.NewUnauthorizedNasError(ip, identifier, err)
133: 		}
134: 		return nil, err
135: 	}
136: 	s.nasCache.Set(cacheKey, nas)
137: 	return nas, nil
138: }
139: 
140: // GetValidUser retrieves a valid user and performs initial checks
141: // Deprecated: Use UserRepo methods with plugin-based validation instead
142: func (s *RadiusService) GetValidUser(usernameOrMac string, macauth bool) (user *domain.RadiusUser, err error) {
143: 	cacheKey := fmt.Sprintf("%t|%s", macauth, usernameOrMac)
144: 	if cached, ok := s.userCache.Get(cacheKey); ok {
145: 		return cached, nil
146: 	}
147: 	// Adapter: delegate to repository layer
148: 	ctx := context.Background()
149: 	if macauth {
150: 		user, err = s.UserRepo.GetByMacAddr(ctx, usernameOrMac)
151: 	} else {
152: 		user, err = s.UserRepo.GetByUsername(ctx, usernameOrMac)
153: 	}
154: 
155: 	if err != nil {
156: 		if errors.Is(err, gorm.ErrRecordNotFound) {
157: 			return nil, radiuserrors.NewUserNotExistsError()
158: 		}
159: 		return nil, err
160: 	}
161: 
162: 	// Keep original validation logic for backward compatibility
163: 	if user.Status == common.DISABLED {
164: 		return nil, radiuserrors.NewUserDisabledError()
165: 	}
166: 
167: 	if user.ExpireTime.Before(time.Now()) {
168: 		return nil, radiuserrors.NewUserExpiredError()
169: 	}
170: 
171: 	// Postpaid subscription enforcement: reject users whose subscription is suspended
172: 	// due to unpaid invoices. This check is only relevant for postpaid billing users
173: 	// and does not affect prepaid (voucher) users at all.
174: 	if user.BillingType == "postpaid" && user.SubscriptionStatus == "suspended" {
175: 		return nil, radiuserrors.NewUserDisabledError()
176: 	}
177: 
178: 	s.userCache.Set(cacheKey, user)
179: 	return user, nil
180: }
181: 
182: // GetUserForAcct fetches the user without validating expiration or status
183: // Deprecated: Use UserRepo.GetByUsername instead
184: func (s *RadiusService) GetUserForAcct(username string) (user *domain.RadiusUser, err error) {
185: 	// Adapter: delegate to repository layer
186: 	user, err = s.UserRepo.GetByUsername(context.Background(), username)
187: 	if err != nil {
188: 		if errors.Is(err, gorm.ErrRecordNotFound) {
189: 			return nil, radiuserrors.NewUserNotExistsError()
190: 		}
191: 		return nil, err
192: 	}
193: 	return user, nil
194: }
195: 
196: // Deprecated: Use UserRepo.UpdateField instead
197: func (s *RadiusService) UpdateUserField(username string, field string, value interface{}) {
198: 	err := s.UserRepo.UpdateField(context.Background(), username, field, value)
199: 	if err != nil {
200: 		zap.L().Error(fmt.Sprintf("update user %s error", field), zap.Error(err), zap.String("namespace", "radius"))
201: 	}
202: }
203: 
204: // Deprecated: Use UserRepo.UpdateMacAddr instead
205: func (s *RadiusService) UpdateUserMac(username string, macaddr string) {
206: 	_ = s.UserRepo.UpdateMacAddr(context.Background(), username, macaddr)
207: }
208: 
209: // Deprecated: Use UserRepo.UpdateVlanId instead
210: func (s *RadiusService) UpdateUserVlanid1(username string, vlanid1 int) {
211: 	_ = s.UserRepo.UpdateVlanId(context.Background(), username, vlanid1, 0)
212: }
213: 
214: // Deprecated: Use UserRepo.UpdateVlanId instead
215: func (s *RadiusService) UpdateUserVlanid2(username string, vlanid2 int) {
216: 	_ = s.UserRepo.UpdateVlanId(context.Background(), username, 0, vlanid2)
217: }
218: 
219: // Deprecated: Use UserRepo.UpdateLastOnline instead
220: func (s *RadiusService) UpdateUserLastOnline(username string) {
221: 	_ = s.UserRepo.UpdateLastOnline(context.Background(), username)
222: }
223: 
224: func (s *RadiusService) GetEapMethod() string {
225: 	// Read directly from the ConfigManager (already in memory)
226: 	return s.appCtx.ConfigMgr().GetString("radius", "EapMethod")
227: }
228: 
229: // Config returns the application configuration
230: func (s *RadiusService) Config() *config.AppConfig {
231: 	return s.appCtx.Config()
232: }
233: 
234: // AppContext returns the application context
235: func (s *RadiusService) AppContext() app.AppContext {
236: 	return s.appCtx
237: }
238: 
239: func GetFramedIpv6Address(r *radius.Request, nas *domain.NetNas) string {
240: 	switch nas.VendorCode {
241: 	case vendors.CodeHuawei:
242: 		return common.IfEmptyStr(huawei.HuaweiFramedIPv6Address_Get(r.Packet).String(), common.NA)
243: 	default:
244: 		return ""
245: 	}
246: }
247: 
248: func GetNetRadiusOnlineFromRequest(r *radius.Request, vr *VendorRequest, nas *domain.NetNas, nasrip string) domain.RadiusOnline {
249: 	acctInputOctets := int(rfc2866.AcctInputOctets_Get(r.Packet))
250: 	acctInputGigawords := int(rfc2869.AcctInputGigawords_Get(r.Packet))
251: 	acctOutputOctets := int(rfc2866.AcctOutputOctets_Get(r.Packet))
252: 	acctOutputGigawords := int(rfc2869.AcctOutputGigawords_Get(r.Packet))
253: 
254: 	getAcctStartTime := func(sessionTime int) time.Time {
255: 		m, _ := time.ParseDuration(fmt.Sprintf("-%ds", sessionTime))
256: 		return time.Now().Add(m)
257: 	}
258: 	return domain.RadiusOnline{
259: 		ID:                  0,
260: 		Username:            rfc2865.UserName_GetString(r.Packet),
261: 		NasId:               common.IfEmptyStr(rfc2865.NASIdentifier_GetString(r.Packet), common.NA),
262: 		NasAddr:             nas.Ipaddr,
263: 		NasPaddr:            nasrip,
264: 		SessionTimeout:      int(rfc2865.SessionTimeout_Get(r.Packet)),
265: 		FramedIpaddr:        common.IfEmptyStr(rfc2865.FramedIPAddress_Get(r.Packet).String(), common.NA),
266: 		FramedNetmask:       common.IfEmptyStr(rfc2865.FramedIPNetmask_Get(r.Packet).String(), common.NA),
267: 		FramedIpv6Address:   GetFramedIpv6Address(r, nas),
268: 		FramedIpv6Prefix:    common.IfEmptyStr(rfc3162.FramedIPv6Prefix_Get(r.Packet).String(), common.NA),
269: 		DelegatedIpv6Prefix: common.IfEmptyStr(rfc4818.DelegatedIPv6Prefix_Get(r.Packet).String(), common.NA),
270: 		MacAddr:             common.IfEmptyStr(vr.MacAddr, common.NA),
271: 		NasPort:             0,
272: 		NasClass:            common.NA,
273: 		NasPortId:           common.IfEmptyStr(rfc2869.NASPortID_GetString(r.Packet), common.NA),
274: 		NasPortType:         0,
275: 		ServiceType:         0,
276: 		AcctSessionId:       rfc2866.AcctSessionID_GetString(r.Packet),
277: 		AcctSessionTime:     int(rfc2866.AcctSessionTime_Get(r.Packet)),
278: 		AcctInputTotal:      int64(acctInputOctets) + int64(acctInputGigawords)*4*1024*1024*1024,
279: 		AcctOutputTotal:     int64(acctOutputOctets) + int64(acctOutputGigawords)*4*1024*1024*1024,
280: 		AcctInputPackets:    int(rfc2866.AcctInputPackets_Get(r.Packet)),
281: 		AcctOutputPackets:   int(rfc2866.AcctOutputPackets_Get(r.Packet)),
282: 		AcctStartTime:       getAcctStartTime(int(rfc2866.AcctSessionTime_Get(r.Packet))),
283: 		LastUpdate:          time.Now(),
284: 	}
285: 
286: }
287: 
288: // CheckAuthRateLimit
289: // Authentication frequency detection, each user can only authenticate once every few seconds
290: func (s *RadiusService) CheckAuthRateLimit(username string) error {
291: 	s.arclock.Lock()
292: 	defer s.arclock.Unlock()
293: 	val, ok := s.AuthRateCache[username]
294: 	if ok {
295: 		if time.Now().Before(val.Starttime.Add(time.Duration(RadiusAuthRateInterval) * time.Second)) {
296: 			return radiuserrors.NewOnlineLimitError("there is a authentication still in process")
297: 		}
298: 		delete(s.AuthRateCache, username)
299: 	}
300: 	s.AuthRateCache[username] = AuthRateUser{
301: 		Username:  username,
302: 		Starttime: time.Now(),
303: 	}
304: 	return nil
305: }
306: 
307: func (s *RadiusService) ReleaseAuthRateLimit(username string) {
308: 	s.arclock.Lock()
309: 	defer s.arclock.Unlock()
310: 	delete(s.AuthRateCache, username)
311: }
312: 
313: // Deprecated: Use SessionRepo.Create instead
314: func (s *RadiusService) AddRadiusOnline(ol domain.RadiusOnline) error {
315: 	ol.ID = common.UUIDint64()
316: 	return s.SessionRepo.Create(context.Background(), &ol)
317: }
318: 
319: // Deprecated: Use AccountingRepo.Create instead
320: func (s *RadiusService) AddRadiusAccounting(ol domain.RadiusOnline, start bool) error {
321: 	accounting := domain.RadiusAccounting{
322: 		ID:                  common.UUIDint64(),
323: 		Username:            ol.Username,
324: 		AcctSessionId:       ol.AcctSessionId,
325: 		NasId:               ol.NasId,
326: 		NasAddr:             ol.NasAddr,
327: 		NasPaddr:            ol.NasPaddr,
328: 		SessionTimeout:      ol.SessionTimeout,
329: 		FramedIpaddr:        ol.FramedIpaddr,
330: 		FramedNetmask:       ol.FramedNetmask,
331: 		FramedIpv6Prefix:    ol.FramedIpv6Prefix,
332: 		FramedIpv6Address:   ol.FramedIpv6Address,
333: 		DelegatedIpv6Prefix: ol.DelegatedIpv6Prefix,
334: 		MacAddr:             ol.MacAddr,
335: 		NasPort:             ol.NasPort,
336: 		NasClass:            ol.NasClass,
337: 		NasPortId:           ol.NasPortId,
338: 		NasPortType:         ol.NasPortType,
339: 		ServiceType:         ol.ServiceType,
340: 		AcctSessionTime:     ol.AcctSessionTime,
341: 		AcctInputTotal:      ol.AcctInputTotal,
342: 		AcctOutputTotal:     ol.AcctOutputTotal,
343: 		AcctInputPackets:    ol.AcctInputPackets,
344: 		AcctOutputPackets:   ol.AcctOutputPackets,
345: 		LastUpdate:          time.Now(),
346: 		AcctStartTime:       ol.AcctStartTime,
347: 		AcctStopTime:        time.Time{},
348: 	}
349: 
350: 	if !start {
351: 		accounting.AcctStopTime = time.Now()
352: 	}
353: 	return s.AccountingRepo.Create(context.Background(), &accounting)
354: }
355: 
356: // Deprecated: Use SessionRepo.CountByUsername instead
357: func (s *RadiusService) GetRadiusOnlineCount(username string) int {
358: 	count, _ := s.SessionRepo.CountByUsername(context.Background(), username)
359: 	return count
360: }
361: 
362: // Deprecated: Use SessionRepo.Exists instead
363: func (s *RadiusService) ExistRadiusOnline(sessionId string) bool {
364: 	exists, _ := s.SessionRepo.Exists(context.Background(), sessionId)
365: 	return exists
366: }
367: 
368: // Deprecated: Use SessionRepo.Update instead
369: func (s *RadiusService) UpdateRadiusOnlineData(data domain.RadiusOnline) error {
370: 	return s.SessionRepo.Update(context.Background(), &data)
371: }
372: 
373: // Deprecated: Use AccountingRepo.UpdateStop instead
374: func (s *RadiusService) EndRadiusAccounting(online domain.RadiusOnline) error {
375: 	accounting := domain.RadiusAccounting{
376: 		AcctSessionId:     online.AcctSessionId,
377: 		AcctSessionTime:   online.AcctSessionTime,
378: 		AcctInputTotal:    online.AcctInputTotal,
379: 		AcctOutputTotal:   online.AcctOutputTotal,
380: 		AcctInputPackets:  online.AcctInputPackets,
381: 		AcctOutputPackets: online.AcctOutputPackets,
382: 	}
383: 	return s.AccountingRepo.UpdateStop(context.Background(), online.AcctSessionId, &accounting)
384: }
385: 
386: // Deprecated: Use SessionRepo.Delete instead
387: func (s *RadiusService) RemoveRadiusOnline(sessionId string) error {
388: 	return s.SessionRepo.Delete(context.Background(), sessionId)
389: }
390: 
391: // Deprecated: Use SessionRepo.BatchDelete instead
392: func (s *RadiusService) BatchClearRadiusOnline(ids string) error {
393: 	return s.SessionRepo.BatchDelete(context.Background(), strings.Split(ids, ","))
394: }
395: 
396: // Deprecated: Use SessionRepo.BatchDeleteByNas instead
397: func (s *RadiusService) BatchClearRadiusOnlineByNas(nasip, nasid string) {
398: 	_ = s.SessionRepo.BatchDeleteByNas(context.Background(), nasip, nasid)
399: }
400: 
401: func (s *RadiusService) Release() {
402: 	s.TaskPool.Running()
403: 	_ = s.TaskPool.ReleaseTimeout(time.Second * 5)
404: }
405: 
406: // ErrSecretEmpty indicates an empty RADIUS secret
407: var ErrSecretEmpty = errors.New("secret is empty")
408: 
409: // ErrSecretMismatch indicates a RADIUS secret mismatch
410: var ErrSecretMismatch = errors.New("secret mismatch")
411: 
412: // CheckRequestSecret validates the RADIUS packet authenticator against the shared secret.
413: // Returns an error if validation fails, nil on success.
414: func (s *RadiusService) CheckRequestSecret(r *radius.Packet, secret []byte) error {
415: 	request, err := r.MarshalBinary()
416: 	if err != nil {
417: 		return fmt.Errorf("failed to marshal packet: %w", err)
418: 	}
419: 
420: 	if len(secret) == 0 {
421: 		return ErrSecretEmpty
422: 	}
423: 
424: 	hash := md5.New()
425: 	hash.Write(request[:4])
426: 	var nul [16]byte
427: 	hash.Write(nul[:])
428: 	hash.Write(request[20:])
429: 	hash.Write(secret)
430: 	var sum [md5.Size]byte
431: 	if !bytes.Equal(hash.Sum(sum[:0]), request[4:20]) {
432: 		return ErrSecretMismatch
433: 	}
434: 	return nil
435: }
436: 
437: // State add
438: func (s *RadiusService) AddEapState(stateid, username string, challenge []byte, eapMethad string) {
439: 	s.eaplock.Lock()
440: 	defer s.eaplock.Unlock()
441: 	s.EapStateCache[stateid] = EapState{
442: 		Username:  username,
443: 		StateID:   stateid,
444: 		Challenge: challenge,
445: 		EapMethad: eapMethad,
446: 		Success:   false,
447: 	}
448: }
449: 
450: // State get
451: func (s *RadiusService) GetEapState(stateid string) (state *EapState, err error) {
452: 	s.eaplock.Lock()
453: 	defer s.eaplock.Unlock()
454: 	val, ok := s.EapStateCache[stateid]
455: 	if ok {
456: 		return &val, nil
457: 	}
458: 	return nil, errors.New("state not found")
459: }
460: 
461: // State delete
462: func (s *RadiusService) DeleteEapState(stateid string) {
463: 	s.eaplock.Lock()
464: 	defer s.eaplock.Unlock()
465: 	delete(s.EapStateCache, stateid)
466: }
467: 
468: func (s *AuthService) GetLocalPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
469: 	if isMacAuth {
470: 		return user.MacAddr, nil
471: 	}
472: 	return user.Password, nil
473: }
474: 
475: func (s *AuthService) UpdateBind(user *domain.RadiusUser, vendorReq *VendorRequest) {
476: 	if user.MacAddr != vendorReq.MacAddr {
477: 		s.UpdateUserMac(user.Username, vendorReq.MacAddr)
478: 	}
479: 	reqvid1 := int(vendorReq.Vlanid1)
480: 	reqvid2 := int(vendorReq.Vlanid2)
481: 	if user.Vlanid1 != reqvid1 {
482: 		s.UpdateUserVlanid2(user.Username, reqvid1)
483: 	}
484: 	if user.Vlanid2 != reqvid2 {
485: 		s.UpdateUserVlanid2(user.Username, reqvid2)
486: 	}
487: }
488: 
489: // ApplyAcceptEnhancers delivers user profile configuration via plugins
490: func (s *AuthService) ApplyAcceptEnhancers(
491: 	user *domain.RadiusUser,
492: 	nas *domain.NetNas,
493: 	vendorReq *vendorparsers.VendorRequest,
494: 	radAccept *radius.Packet,
495: ) {
496: 	authCtx := &auth.AuthContext{
497: 		User:          user,
498: 		Nas:           nas,
499: 		VendorRequest: vendorReq,
500: 		Response:      radAccept,
501: 	}
502: 
503: 	ctx := context.Background()
504: 	for _, enhancer := range registry.GetResponseEnhancers() {
505: 		if err := enhancer.Enhance(ctx, authCtx); err != nil {
506: 			zap.L().Warn("response enhancer failed",
507: 				zap.String("enhancer", enhancer.Name()),
508: 				zap.Error(err))
509: 		}
510: 	}
511: }
512: 
513: func (s *RadiusService) DoAcctDisconnect(r *radius.Request, nas *domain.NetNas, username, nasrip string) {
514: 	packet := radius.New(radius.CodeDisconnectRequest, []byte(nas.Secret))
515: 	sessionid := rfc2866.AcctSessionID_GetString(r.Packet)
516: 	if sessionid == "" {
517: 		return
518: 	}
519: 	_ = rfc2865.UserName_SetString(packet, username)
520: 	_ = rfc2866.AcctSessionID_Set(packet, []byte(sessionid))
521: 	response, err := radius.Exchange(context.Background(), packet, fmt.Sprintf("%s:%d", nasrip, nas.CoaPort))
522: 	if err != nil {
523: 		zap.L().Error("radius disconnect error",
524: 			zap.String("namespace", "radius"),
525: 			zap.String("username", username),
526: 			zap.Error(err),
527: 		)
528: 		return
529: 	}
530: 	zap.L().Info("radius disconnect done",
531: 		zap.String("namespace", "radius"),
532: 		zap.String("nasip", nasrip),
533: 		zap.Int("coaport", nas.CoaPort),
534: 		zap.String("request", FmtPacket(packet)),
535: 		zap.String("response", FmtPacket(response)),
536: 	)
537: }

(End of file - total 537 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 6.2s)

_Thinking:_

Now I can see the structure. The `RadiusService` has a `UserRepo` and a `VoucherRepo`. Let me check the repository to understand how users are queried vs vouchers. Let me also see what `GetValidUser` does - it queries the UserRepo.

The key insight: The `GetValidUser` method queries only the `UserRepo`, not the `VoucherRepo`. This is likely why vouchers are not authenticating - the authentication process doesn't look up vouchers at all.

Let me verify this by checking the auth pipeline and see where the lookup happens.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/radiusd/auth_pipeline.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/radiusd/auth_pipeline.go</path>
<type>file</type>
<content>1: package radiusd
2: 
3: import (
4: 	"context"
5: 	"fmt"
6: 	"sync"
7: 
8: 	"github.com/talkincode/toughradius/v9/internal/domain"
9: 	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
10: 	"layeh.com/radius"
11: )
12: 
13: // AuthPipelineStage represents a pluggable processing unit inside ServeRADIUS.
14: type AuthPipelineStage interface {
15: 	Name() string
16: 	Execute(ctx *AuthPipelineContext) error
17: }
18: 
19: type stageFunc struct {
20: 	name string
21: 	fn   func(ctx *AuthPipelineContext) error
22: }
23: 
24: func (s *stageFunc) Name() string {
25: 	return s.name
26: }
27: 
28: func (s *stageFunc) Execute(ctx *AuthPipelineContext) error {
29: 	return s.fn(ctx)
30: }
31: 
32: func newStage(name string, fn func(ctx *AuthPipelineContext) error) AuthPipelineStage {
33: 	return &stageFunc{name: name, fn: fn}
34: }
35: 
36: // AuthPipeline manages ordered stage execution.
37: type AuthPipeline struct {
38: 	mu     sync.RWMutex
39: 	stages []AuthPipelineStage
40: }
41: 
42: // NewAuthPipeline creates an empty pipeline instance.
43: func NewAuthPipeline() *AuthPipeline {
44: 	return &AuthPipeline{stages: make([]AuthPipelineStage, 0)}
45: }
46: 
47: // Use appends a stage to the end of the pipeline.
48: func (p *AuthPipeline) Use(stage AuthPipelineStage) *AuthPipeline {
49: 	p.mu.Lock()
50: 	defer p.mu.Unlock()
51: 	p.stages = append(p.stages, stage)
52: 	return p
53: }
54: 
55: // InsertBefore inserts a stage before the target stage name.
56: func (p *AuthPipeline) InsertBefore(target string, stage AuthPipelineStage) error {
57: 	p.mu.Lock()
58: 	defer p.mu.Unlock()
59: 	idx := p.indexOf(target)
60: 	if idx == -1 {
61: 		return fmt.Errorf("stage %s not found", target)
62: 	}
63: 	p.stages = append(p.stages[:idx], append([]AuthPipelineStage{stage}, p.stages[idx:]...)...)
64: 	return nil
65: }
66: 
67: // InsertAfter inserts a stage after the target stage name.
68: func (p *AuthPipeline) InsertAfter(target string, stage AuthPipelineStage) error {
69: 	p.mu.Lock()
70: 	defer p.mu.Unlock()
71: 	idx := p.indexOf(target)
72: 	if idx == -1 {
73: 		return fmt.Errorf("stage %s not found", target)
74: 	}
75: 	insertPos := idx + 1
76: 	p.stages = append(p.stages[:insertPos], append([]AuthPipelineStage{stage}, p.stages[insertPos:]...)...)
77: 	return nil
78: }
79: 
80: // Replace swaps the stage with the provided implementation.
81: func (p *AuthPipeline) Replace(target string, stage AuthPipelineStage) error {
82: 	p.mu.Lock()
83: 	defer p.mu.Unlock()
84: 	idx := p.indexOf(target)
85: 	if idx == -1 {
86: 		return fmt.Errorf("stage %s not found", target)
87: 	}
88: 	p.stages[idx] = stage
89: 	return nil
90: }
91: 
92: // Remove deletes a stage from the pipeline.
93: func (p *AuthPipeline) Remove(target string) error {
94: 	p.mu.Lock()
95: 	defer p.mu.Unlock()
96: 	idx := p.indexOf(target)
97: 	if idx == -1 {
98: 		return fmt.Errorf("stage %s not found", target)
99: 	}
100: 	p.stages = append(p.stages[:idx], p.stages[idx+1:]...)
101: 	return nil
102: }
103: 
104: // Stages returns a copy of the registered stages.
105: func (p *AuthPipeline) Stages() []AuthPipelineStage {
106: 	p.mu.RLock()
107: 	defer p.mu.RUnlock()
108: 	stages := make([]AuthPipelineStage, len(p.stages))
109: 	copy(stages, p.stages)
110: 	return stages
111: }
112: 
113: // Execute runs stages sequentially until completion or ctx.Stop() is invoked.
114: func (p *AuthPipeline) Execute(ctx *AuthPipelineContext) error {
115: 	p.mu.RLock()
116: 	stages := make([]AuthPipelineStage, len(p.stages))
117: 	copy(stages, p.stages)
118: 	p.mu.RUnlock()
119: 
120: 	for _, stage := range stages {
121: 		if ctx.IsStopped() {
122: 			break
123: 		}
124: 		if err := stage.Execute(ctx); err != nil {
125: 			return fmt.Errorf("stage %s failed: %w", stage.Name(), err)
126: 		}
127: 	}
128: 	return nil
129: }
130: 
131: func (p *AuthPipeline) indexOf(name string) int {
132: 	for idx, stage := range p.stages {
133: 		if stage.Name() == name {
134: 			return idx
135: 		}
136: 	}
137: 	return -1
138: }
139: 
140: // AuthPipelineContext carries per-request mutable data across stages.
141: type AuthPipelineContext struct {
142: 	Context context.Context
143: 	Service *AuthService
144: 
145: 	Writer   radius.ResponseWriter
146: 	Request  *radius.Request
147: 	Response *radius.Packet
148: 
149: 	Username         string
150: 	NasIdentifier    string
151: 	CallingStationID string
152: 	RemoteIP         string
153: 
154: 	NAS                    *domain.NetNas
155: 	VendorRequest          *VendorRequest
156: 	VendorRequestForPlugin *vendorparsers.VendorRequest
157: 	User                   *domain.RadiusUser
158: 
159: 	IsEAP            bool
160: 	EAPMethod        string
161: 	IsMacAuth        bool
162: 	RateLimitChecked bool
163: 
164: 	stop bool
165: }
166: 
167: // NewAuthPipelineContext builds a context with sane defaults.
168: func NewAuthPipelineContext(service *AuthService, w radius.ResponseWriter, r *radius.Request) *AuthPipelineContext {
169: 	return &AuthPipelineContext{
170: 		Context:                context.Background(),
171: 		Service:                service,
172: 		Writer:                 w,
173: 		Request:                r,
174: 		VendorRequest:          &VendorRequest{},
175: 		VendorRequestForPlugin: &vendorparsers.VendorRequest{},
176: 	}
177: }
178: 
179: // Stop halts further stage execution.
180: func (ctx *AuthPipelineContext) Stop() {
181: 	ctx.stop = true
182: }
183: 
184: // IsStopped reports whether execution has been halted.
185: func (ctx *AuthPipelineContext) IsStopped() bool {
186: 	return ctx.stop
187: }

(End of file - total 187 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 4.5s)

_Thinking:_

Now let me find where the pipeline is configured and see what stages process the user authentication. I need to look at where the `GetValidUser` method is called.

Now let me find where `GetValidUser` is called in the authentication pipeline:

**Tool: grep**

**Input:**
```json
{
  "include": "*.go",
  "pattern": "GetValidUser"
}
```

**Output:**
```
Found 3 matches
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius.go:
  Line 140: // GetValidUser retrieves a valid user and performs initial checks
  Line 142: func (s *RadiusService) GetValidUser(usernameOrMac string, macauth bool) (user *domain.RadiusUser, err error) {

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/auth_stages.go:
  Line 121: 	user, err := s.GetValidUser(ctx.Username, ctx.IsMacAuth)
```

---

## Assistant (Plan · minimax-m2.5-free · 20.3s)

_Thinking:_

Let me look at auth_stages.go to see how the authentication pipeline works and where the user is looked up.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/radiusd/auth_stages.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/radiusd/auth_stages.go</path>
<type>file</type>
<content>1: package radiusd
2: 
3: import (
4: 	"fmt"
5: 	"net"
6: 	"strings"
7: 
8: 	"github.com/talkincode/toughradius/v9/internal/app"
9: 	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
10: 	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
11: 	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
12: 	"go.uber.org/zap"
13: 	"layeh.com/radius"
14: 	"layeh.com/radius/rfc2865"
15: )
16: 
17: const (
18: 	StageRequestMetadata = "request_metadata"
19: 	StageNasLookup       = "nas_lookup"
20: 	StageRateLimit       = "auth_rate_limit"
21: 	StageVendorParsing   = "vendor_parsing"
22: 	StageLoadUser        = "load_user"
23: 	StageEAPDispatch     = "eap_dispatch"
24: 	StagePluginAuth      = "plugin_auth"
25: )
26: 
27: func (s *AuthService) registerDefaultStages() {
28: 	stages := []AuthPipelineStage{
29: 		newStage(StageRequestMetadata, s.stageRequestMetadata),
30: 		newStage(StageNasLookup, s.stageNasLookup),
31: 		newStage(StageRateLimit, s.stageRateLimit),
32: 		newStage(StageVendorParsing, s.stageVendorParsing),
33: 		newStage(StageLoadUser, s.stageLoadUser),
34: 		newStage(StageEAPDispatch, s.stageEAPDispatch),
35: 		newStage(StagePluginAuth, s.stagePluginAuth),
36: 	}
37: 
38: 	for _, stage := range stages {
39: 		s.authPipeline.Use(stage)
40: 	}
41: }
42: 
43: func (s *AuthService) stageRequestMetadata(ctx *AuthPipelineContext) error {
44: 	r := ctx.Request
45: 
46: 	preferredMethod := s.resolveEapMethod(s.GetEapMethod())
47: 	ctx.EAPMethod = preferredMethod
48: 
49: 	if _, err := eap.ParseEAPMessage(r.Packet); err == nil {
50: 		ctx.IsEAP = true
51: 	}
52: 
53: 	host, _, err := net.SplitHostPort(r.RemoteAddr.String())
54: 	if err != nil {
55: 		ctx.RemoteIP = r.RemoteAddr.String()
56: 	} else {
57: 		ctx.RemoteIP = host
58: 	}
59: 
60: 	ctx.NasIdentifier = rfc2865.NASIdentifier_GetString(r.Packet)
61: 	ctx.Username = rfc2865.UserName_GetString(r.Packet)
62: 	ctx.CallingStationID = rfc2865.CallingStationID_GetString(r.Packet)
63: 
64: 	if ctx.Username == "" {
65: 		return radiuserrors.NewAuthErrorWithStage(
66: 			app.MetricsRadiusRejectNotExists,
67: 			"username is empty of client mac",
68: 			StageRequestMetadata,
69: 		)
70: 	}
71: 
72: 	return nil
73: }
74: 
75: func (s *AuthService) stageNasLookup(ctx *AuthPipelineContext) error {
76: 	nas, err := s.GetNas(ctx.RemoteIP, ctx.NasIdentifier)
77: 	if err != nil {
78: 		return err
79: 	}
80: 	ctx.NAS = nas
81: 
82: 	if nas != nil {
83: 		secret := []byte(nas.Secret)
84: 		ctx.Request.Secret = secret
85: 		ctx.Request.Secret = secret //nolint:staticcheck
86: 		ctx.Response = ctx.Request.Response(radius.CodeAccessAccept)
87: 	}
88: 
89: 	return nil
90: }
91: 
92: func (s *AuthService) stageRateLimit(ctx *AuthPipelineContext) error {
93: 	if ctx.IsEAP {
94: 		return nil
95: 	}
96: 	if err := s.CheckAuthRateLimit(ctx.Username); err != nil {
97: 		return err
98: 	}
99: 	ctx.RateLimitChecked = true
100: 	return nil
101: }
102: 
103: func (s *AuthService) stageVendorParsing(ctx *AuthPipelineContext) error {
104: 	if ctx.NAS == nil {
105: 		return fmt.Errorf("nas should not be nil before vendor parsing")
106: 	}
107: 	vendorReq := s.ParseVendor(ctx.Request, ctx.NAS.VendorCode)
108: 	ctx.VendorRequest = vendorReq
109: 
110: 	ctx.IsMacAuth = vendorReq.MacAddr != "" && vendorReq.MacAddr == ctx.Username
111: 
112: 	ctx.VendorRequestForPlugin = &vendorparsers.VendorRequest{
113: 		MacAddr: vendorReq.MacAddr,
114: 		Vlanid1: vendorReq.Vlanid1,
115: 		Vlanid2: vendorReq.Vlanid2,
116: 	}
117: 	return nil
118: }
119: 
120: func (s *AuthService) stageLoadUser(ctx *AuthPipelineContext) error {
121: 	user, err := s.GetValidUser(ctx.Username, ctx.IsMacAuth)
122: 	if err != nil {
123: 		return err
124: 	}
125: 	ctx.User = user
126: 	return nil
127: }
128: 
129: func (s *AuthService) stageEAPDispatch(ctx *AuthPipelineContext) error {
130: 	if !ctx.IsEAP || s.eapHelper == nil {
131: 		return nil
132: 	}
133: 
134: 	handled, success, eapErr := s.eapHelper.HandleEAPAuthentication(
135: 		ctx.Writer,
136: 		ctx.Request,
137: 		ctx.User,
138: 		ctx.NAS,
139: 		ctx.VendorRequestForPlugin,
140: 		ctx.Response,
141: 		ctx.EAPMethod,
142: 	)
143: 
144: 	if eapErr != nil {
145: 		zap.L().Warn("eap handling failed",
146: 			zap.String("namespace", "radius"),
147: 			zap.Error(eapErr),
148: 		)
149: 		_ = s.eapHelper.SendEAPFailure(ctx.Writer, ctx.Request, ctx.NAS.Secret, eapErr)
150: 		s.eapHelper.CleanupState(ctx.Request)
151: 		ctx.Stop()
152: 		return nil
153: 	}
154: 
155: 	if handled {
156: 		if success {
157: 			err := s.AuthenticateUserWithPlugins(ctx.Context, ctx.Request, ctx.Response, ctx.User, ctx.NAS, ctx.VendorRequestForPlugin, ctx.IsMacAuth, SkipPasswordValidation())
158: 			if err != nil {
159: 				_ = s.eapHelper.SendEAPFailure(ctx.Writer, ctx.Request, ctx.NAS.Secret, err)
160: 				s.eapHelper.CleanupState(ctx.Request)
161: 				ctx.Stop()
162: 				return nil
163: 			}
164: 			s.sendAcceptResponse(ctx, true)
165: 		}
166: 		ctx.Stop()
167: 	}
168: 
169: 	return nil
170: }
171: 
172: func (s *AuthService) stagePluginAuth(ctx *AuthPipelineContext) error {
173: 	if ctx.IsStopped() {
174: 		return nil
175: 	}
176: 
177: 	err := s.AuthenticateUserWithPlugins(ctx.Context, ctx.Request, ctx.Response, ctx.User, ctx.NAS, ctx.VendorRequestForPlugin, ctx.IsMacAuth)
178: 	if err != nil {
179: 		return err
180: 	}
181: 
182: 	s.sendAcceptResponse(ctx, false)
183: 	ctx.Stop()
184: 	return nil
185: }
186: 
187: func (s *AuthService) sendAcceptResponse(ctx *AuthPipelineContext, isEapFlow bool) {
188: 	vendorPlugin := ctx.VendorRequestForPlugin
189: 	if vendorPlugin == nil {
190: 		vendorPlugin = &vendorparsers.VendorRequest{}
191: 	}
192: 
193: 	if ctx.NAS == nil || ctx.User == nil {
194: 		zap.L().Warn("skip accept response due to missing context",
195: 			zap.String("namespace", "radius"),
196: 			zap.Bool("is_eap", isEapFlow),
197: 		)
198: 		return
199: 	}
200: 
201: 	s.ApplyAcceptEnhancers(ctx.User, ctx.NAS, vendorPlugin, ctx.Response)
202: 
203: 	if isEapFlow && s.eapHelper != nil {
204: 		if err := s.eapHelper.SendEAPSuccess(ctx.Writer, ctx.Request, ctx.Response, ctx.NAS.Secret); err != nil {
205: 			zap.L().Error("send eap success failed",
206: 				zap.String("namespace", "radius"),
207: 				zap.Error(err),
208: 			)
209: 		}
210: 		s.eapHelper.CleanupState(ctx.Request)
211: 	} else {
212: 		s.SendAccept(ctx.Writer, ctx.Request, ctx.Response)
213: 	}
214: 
215: 	vendorReq := ctx.VendorRequest
216: 	if vendorReq == nil {
217: 		vendorReq = &VendorRequest{}
218: 	}
219: 
220: 	if ctx.User != nil {
221: 		s.UpdateBind(ctx.User, vendorReq)
222: 		s.UpdateUserLastOnline(ctx.User.Username)
223: 	}
224: 
225: 	zap.L().Info("radius auth success",
226: 		zap.String("namespace", "radius"),
227: 		zap.String("username", ctx.Username),
228: 		zap.String("nasip", ctx.RemoteIP),
229: 		zap.Bool("is_eap", isEapFlow),
230: 		zap.String("result", "success"),
231: 		zap.String("metrics", app.MetricsRadiusAccept),
232: 	)
233: }
234: 
235: func (s *AuthService) resolveEapMethod(preferred string) string {
236: 	method := strings.TrimSpace(strings.ToLower(preferred))
237: 	if method == "" {
238: 		method = "eap-md5"
239: 	}
240: 	if len(s.allowedEAPHandlers) == 0 {
241: 		return method
242: 	}
243: 	if _, ok := s.allowedEAPHandlers[method]; ok {
244: 		return method
245: 	}
246: 	for _, candidate := range s.allowedEAPHandlersOrder {
247: 		if _, ok := s.allowedEAPHandlers[candidate]; ok {
248: 			zap.L().Warn("preferred EAP method disabled, falling back",
249: 				zap.String("namespace", "radius"),
250: 				zap.String("preferred", method),
251: 				zap.String("fallback", candidate),
252: 			)
253: 			return candidate
254: 		}
255: 	}
256: 	for candidate := range s.allowedEAPHandlers {
257: 		zap.L().Warn("preferred EAP method disabled, falling back",
258: 			zap.String("namespace", "radius"),
259: 			zap.String("preferred", method),
260: 			zap.String("fallback", candidate),
261: 		)
262: 		return candidate
263: 	}
264: 	return method
265: }
266: 
267: func (s *AuthService) ensurePipeline() {
268: 	if s.authPipeline != nil {
269: 		return
270: 	}
271: 	s.authPipeline = NewAuthPipeline()
272: 	s.registerDefaultStages()
273: }
274: 
275: func (s *AuthService) buildAllowedEAPHandlers() []string {
276: 	appCtx := s.AppContext()
277: 	if appCtx == nil {
278: 		return nil
279: 	}
280: 	cfgMgr := appCtx.ConfigMgr()
281: 	if cfgMgr == nil {
282: 		return nil
283: 	}
284: 
285: 	raw := strings.TrimSpace(cfgMgr.GetString("radius", "EapEnabledHandlers"))
286: 	if raw == "" || raw == "*" {
287: 		return nil
288: 	}
289: 
290: 	parts := strings.Split(raw, ",")
291: 	ordered := make([]string, 0, len(parts))
292: 	seen := make(map[string]struct{})
293: 
294: 	for _, part := range parts {
295: 		name := strings.ToLower(strings.TrimSpace(part))
296: 		if name == "" {
297: 			continue
298: 		}
299: 		if name == "*" {
300: 			return nil
301: 		}
302: 		if _, ok := seen[name]; ok {
303: 			continue
304: 		}
305: 		seen[name] = struct{}{}
306: 		ordered = append(ordered, name)
307: 	}
308: 
309: 	return ordered
310: }
311: 
312: func (s *AuthService) initAllowedEAPHandlers() []string {
313: 	allowed := s.buildAllowedEAPHandlers()
314: 	if len(allowed) == 0 {
315: 		s.allowedEAPHandlers = nil
316: 		s.allowedEAPHandlersOrder = nil
317: 		return nil
318: 	}
319: 
320: 	s.allowedEAPHandlers = make(map[string]struct{}, len(allowed))
321: 	for _, name := range allowed {
322: 		s.allowedEAPHandlers[name] = struct{}{}
323: 	}
324: 	s.allowedEAPHandlersOrder = allowed
325: 	return allowed
326: }

(End of file - total 326 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 4.8s)

_Thinking:_

Now I can see clearly that in `stageLoadUser` (line 120-127), the system calls `s.GetValidUser(ctx.Username, ctx.IsMacAuth)` which queries only the `UserRepo`. It doesn't check vouchers at all.

Now I need to see the VoucherRepository to understand how vouchers are supposed to be used:

**Tool: glob**

**Input:**
```json
{
  "pattern": "**/voucher*repository*.go"
}
```

**Output:**
```
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/gorm/voucher_repository.go
```

---

## Assistant (Plan · minimax-m2.5-free · 4.0s)

_Thinking:_

Let me look at the voucher repository to understand how it works:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/gorm/voucher_repository.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/gorm/voucher_repository.go</path>
<type>file</type>
<content>1: package gorm
2: 
3: import (
4: 	"context"
5: 	"time"
6: 
7: 	"github.com/talkincode/toughradius/v9/internal/domain"
8: 	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
9: 	"gorm.io/gorm"
10: )
11: 
12: // GormVoucherRepository is the GORM implementation of the voucher repository
13: type GormVoucherRepository struct {
14: 	db *gorm.DB
15: }
16: 
17: // NewGormVoucherRepository creates a voucher repository instance
18: func NewGormVoucherRepository(db *gorm.DB) repository.VoucherRepository {
19: 	return &GormVoucherRepository{db: db}
20: }
21: 
22: func (r *GormVoucherRepository) GetByCode(ctx context.Context, code string) (*domain.Voucher, error) {
23: 	var voucher domain.Voucher
24: 	err := r.db.WithContext(ctx).Where("code = ?", code).First(&voucher).Error
25: 	if err != nil {
26: 		return nil, err
27: 	}
28: 	return &voucher, nil
29: }
30: 
31: func (r *GormVoucherRepository) UpdateFirstUsedAt(ctx context.Context, code string, firstUsedAt, expireTime time.Time) error {
32: 	updates := map[string]interface{}{
33: 		"first_used_at": firstUsedAt,
34: 		"expire_time":   expireTime,
35: 	}
36: 	return r.db.WithContext(ctx).
37: 		Model(&domain.Voucher{}).
38: 		Where("code = ?", code).
39: 		Updates(updates).Error
40: }
41: 
42: func (r *GormVoucherRepository) GetBatchByID(ctx context.Context, batchID int64) (*domain.VoucherBatch, error) {
43: 	var batch domain.VoucherBatch
44: 	err := r.db.WithContext(ctx).First(&batch, batchID).Error
45: 	if err != nil {
46: 		return nil, err
47: 	}
48: 	return &batch, nil
49: }

(End of file - total 49 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 3.7s)

_Thinking:_

Now I can see that there's a `VoucherRepository` with a `GetByCode` method, but it's not being used in the authentication flow. Let me check the repository interface:

**Tool: glob**

**Input:**
```json
{
  "pattern": "**/repository/*voucher*.go"
}
```

**Output:**
```
No files found
```

---

## Assistant (Plan · minimax-m2.5-free · 2.8s)

_Thinking:_

Let me search more broadly for the repository interface:

**Tool: grep**

**Input:**
```json
{
  "include": "*.go",
  "pattern": "type VoucherRepository interface"
}
```

**Output:**
```
Found 1 matches
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/interfaces.go:
  Line 82: type VoucherRepository interface {
```

---

## Assistant (Plan · minimax-m2.5-free · 3.0s)

_Thinking:_

Let me read the repository interface:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/interfaces.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/interfaces.go</path>
<type>file</type>
<content>1: package repository
2: 
3: import (
4: 	"context"
5: 	"time"
6: 
7: 	"github.com/talkincode/toughradius/v9/internal/domain"
8: )
9: 
10: // UserRepository defines user data access operations
11: type UserRepository interface {
12: 	// GetByUsername finds a user by username
13: 	GetByUsername(ctx context.Context, username string) (*domain.RadiusUser, error)
14: 
15: 	// GetByMacAddr finds a user by MAC address
16: 	GetByMacAddr(ctx context.Context, macAddr string) (*domain.RadiusUser, error)
17: 
18: 	// UpdateMacAddr updates the user's MAC address
19: 	UpdateMacAddr(ctx context.Context, username, macAddr string) error
20: 
21: 	// UpdateVlanId updates the user's VLAN ID
22: 	UpdateVlanId(ctx context.Context, username string, vlanId1, vlanId2 int) error
23: 
24: 	// UpdateLastOnline updates the last online time
25: 	UpdateLastOnline(ctx context.Context, username string) error
26: 
27: 	// UpdateField updates a specified user field
28: 	UpdateField(ctx context.Context, username string, field string, value interface{}) error
29: }
30: 
31: // SessionRepository manages online sessions
32: type SessionRepository interface {
33: 	// Create Create online session
34: 	Create(ctx context.Context, session *domain.RadiusOnline) error
35: 
36: 	// Update updates session data
37: 	Update(ctx context.Context, session *domain.RadiusOnline) error
38: 
39: 	// Delete deletes a session
40: 	Delete(ctx context.Context, sessionId string) error
41: 
42: 	// GetBySessionId finds a session by its ID
43: 	GetBySessionId(ctx context.Context, sessionId string) (*domain.RadiusOnline, error)
44: 
45: 	// CountByUsername counts online sessions per user
46: 	CountByUsername(ctx context.Context, username string) (int, error)
47: 
48: 	// Exists checks whether the session exists
49: 	Exists(ctx context.Context, sessionId string) (bool, error)
50: 
51: 	// BatchDelete deletes sessions in bulk
52: 	BatchDelete(ctx context.Context, ids []string) error
53: 
54: 	// BatchDeleteByNas deletes sessions by NAS
55: 	BatchDeleteByNas(ctx context.Context, nasAddr, nasId string) error
56: }
57: 
58: // AccountingRepository defines accounting record operations
59: type AccountingRepository interface {
60: 	// Create Create accounting record
61: 	Create(ctx context.Context, accounting *domain.RadiusAccounting) error
62: 
63: 	// UpdateStop updates stop time and traffic counters
64: 	UpdateStop(ctx context.Context, sessionId string, accounting *domain.RadiusAccounting) error
65: 	// GetTotalUsage calculates total traffic usage for a user (input + output total)
66: 	GetTotalUsage(ctx context.Context, username string) (int64, error)
67: }
68: 
69: // NasRepository manages NAS devices
70: type NasRepository interface {
71: 	// GetByIP finds a NAS by IP
72: 	GetByIP(ctx context.Context, ip string) (*domain.NetNas, error)
73: 
74: 	// GetByIdentifier finds a NAS by identifier
75: 	GetByIdentifier(ctx context.Context, identifier string) (*domain.NetNas, error)
76: 
77: 	// GetByIPOrIdentifier finds a NAS by IP or identifier
78: 	GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error)
79: }
80: 
81: // VoucherRepository manages vouchers
82: type VoucherRepository interface {
83: 	// GetByCode finds a voucher by code
84: 	GetByCode(ctx context.Context, code string) (*domain.Voucher, error)
85: 
86: 	// UpdateFirstUsedAt updates the first used time and expire time
87: 	UpdateFirstUsedAt(ctx context.Context, code string, firstUsedAt, expireTime time.Time) error
88: 
89: 	// GetBatchByID finds a voucher batch by ID
90: 	GetBatchByID(ctx context.Context, batchID int64) (*domain.VoucherBatch, error)
91: }

(End of file - total 91 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 6.8s)

_Thinking:_

Now I understand the problem. Let me also check if there's any other place where vouchers are supposed to be authenticated. Let me also look at the admin API for vouchers to see how they're created and how the `RadiusUsername` field is used:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/adminapi/vouchers.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/adminapi/vouchers.go</path>
<type>file</type>
<content>1: package adminapi
2: 
3: import (
4: 	"encoding/csv"
5: 	"fmt"
6: 	"net/http"
7: 
8: 	"strconv"
9: 	"strings"
10: 	"time"
11: 
12: 	"github.com/labstack/echo/v4"
13: 	"gorm.io/gorm"
14: 
15: 
16: 	"github.com/talkincode/toughradius/v9/internal/domain"
17: 	"github.com/talkincode/toughradius/v9/internal/webserver"
18: 	"github.com/talkincode/toughradius/v9/pkg/common"
19: 	"go.uber.org/zap"
20: 	"golang.org/x/time/rate"
21: )
22: 
23: // init function removed to avoid duplication/conflict if registerVoucherRoutes is used.
24: 
25: 
26: // ListVoucherBatches retrieves the voucher batch list
27: // @Summary get the voucher batch list
28: // @Tags Voucher
29: // @Param page query int false "Page number"
30: // @Param perPage query int false "Items per page"
31: // @Param sort query string false "Sort field"
32: // @Param order query string false "Sort direction"
33: // @Success 200 {object} ListResponse
34: // @Router /api/v1/voucher-batches [get]
35: func ListVoucherBatches(c echo.Context) error {
36: 	db := GetDB(c)
37: 
38: 	page, _ := strconv.Atoi(c.QueryParam("page"))
39: 	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
40: 	if page < 1 {
41: 		page = 1
42: 	}
43: 	if perPage < 1 || perPage > 100 {
44: 		perPage = 10
45: 	}
46: 
47: 	sortField := c.QueryParam("sort")
48: 	order := c.QueryParam("order")
49: 	if sortField == "" {
50: 		sortField = "id"
51: 	}
52: 	if order != "ASC" && order != "DESC" {
53: 		order = "DESC"
54: 	}
55: 
56: 	var total int64
57: 	var batches []domain.VoucherBatch
58: 
59: 	// Permission check
60: 	currentUser, err := resolveOperatorFromContext(c)
61: 	if err != nil {
62: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
63: 	}
64: 
65: 	query := db.Model(&domain.VoucherBatch{})
66: 
67: 	// Filter by agent: Agents can only see their own batches
68: 	if currentUser.Level == "agent" {
69: 		query = query.Where("agent_id = ? AND is_deleted = ?", currentUser.ID, false)
70: 	} else {
71: 		// Admin: show deleted batches if requested, or filter out by default
72: 		if c.QueryParam("is_deleted") == "true" {
73: 			query = query.Where("is_deleted = ?", true)
74: 		} else if c.QueryParam("is_deleted") == "false" {
75: 			query = query.Where("is_deleted = ?", false)
76: 		}
77: 	}
78: 
79: 	// Filter by name
80: 	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
81: 		query = query.Where("name LIKE ?", "%"+name+"%")
82: 	}
83: 
84: 	query.Count(&total)
85: 
86: 	offset := (page - 1) * perPage
87: 	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&batches)
88: 
89: 	return paged(c, batches, total, page, perPage)
90: }
91: 
92: // ListVouchers retrieves the voucher list
93: // @Summary get the voucher list
94: // @Tags Voucher
95: // @Param page query int false "Page number"
96: // @Param perPage query int false "Items per page"
97: // @Param batch_id query int false "Batch ID"
98: // @Param status query string false "Status"
99: // @Success 200 {object} ListResponse
100: // @Router /api/v1/vouchers [get]
101: func ListVouchers(c echo.Context) error {
102: 	db := GetDB(c)
103: 
104: 	page, _ := strconv.Atoi(c.QueryParam("page"))
105: 	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
106: 	if page < 1 {
107: 		page = 1
108: 	}
109: 	if perPage < 1 || perPage > 1000 {
110: 		perPage = 50
111: 	}
112: 
113: 	var total int64
114: 	var vouchers []domain.Voucher
115: 
116: 	query := db.Model(&domain.Voucher{})
117: 
118: 	if batchID := c.QueryParam("batch_id"); batchID != "" {
119: 		query = query.Where("batch_id = ?", batchID)
120: 	}
121: 
122: 	if status := c.QueryParam("status"); status != "" {
123: 		query = query.Where("status = ?", status)
124: 	}
125: 
126: 	if code := c.QueryParam("code"); code != "" {
127: 		query = query.Where("code = ?", code)
128: 	}
129: 
130: 	// Search by SN (batchid-voucherid format), voucher ID, or exact code match
131: 	if sn := c.QueryParam("sn"); sn != "" {
132: 		parts := strings.Split(sn, "-")
133: 		if len(parts) == 2 {
134: 			// SN format: batchid-voucherid
135: 			batchID, err1 := strconv.ParseInt(parts[0], 10, 64)
136: 			voucherID, err2 := strconv.ParseInt(parts[1], 10, 64)
137: 			if err1 == nil && err2 == nil {
138: 				query = query.Where("batch_id = ? AND id = ?", batchID, voucherID)
139: 			}
140: 		} else if _, err := strconv.ParseInt(sn, 10, 64); err == nil {
141: 			// Pure numeric input - search by voucher ID
142: 			voucherID, _ := strconv.ParseInt(sn, 10, 64)
143: 			query = query.Where("id = ?", voucherID)
144: 		} else {
145: 			// Exact code match
146: 			query = query.Where("code = ?", sn)
147: 		}
148: 	}
149: 
150: 	query.Count(&total)
151: 
152: 	offset := (page - 1) * perPage
153: 	query.Order("id DESC").Limit(perPage).Offset(offset).Find(&vouchers)
154: 
155: 	return paged(c, vouchers, total, page, perPage)
156: }
157: 
158: // VoucherBatchRequest
159: type VoucherBatchRequest struct {
160: 	Name         string `json:"name" validate:"required"`
161: 	ProductID    string `json:"product_id" validate:"required"`
162: 	Count        int    `json:"count" validate:"required,min=1,max=10000"`
163: 	Prefix       string `json:"prefix" validate:"omitempty,max=10"`
164: 	Length       int    `json:"length" validate:"omitempty,min=6,max=20"` // Length of random part
165: 	Type         string `json:"type"`                                      // number, alpha, mixed
166: 	Remark       string `json:"remark"`
167: 	AgentID      string `json:"agent_id" validate:"omitempty"`          // Optional, if generated by agent
168: 	ExpireTime   string `json:"expire_time"`                               // ISO8601 string
169: 	GeneratePIN  bool   `json:"generate_pin"`                             // Generate PIN for vouchers
170: 	PINLength    int    `json:"pin_length" validate:"omitempty,min=4,max=8"` // PIN length (default 4)
171: 	// First-Use Expiration options
172: 	ExpirationType string `json:"expiration_type"`                                     // "fixed" (default) or "first_use"
173: 	ValidityDays   int    `json:"validity_days" validate:"omitempty,min=1,max=365"` // Days of validity for first_use type
174: }
175: 
176: // Local generation functions removed in favor of pkg/common
177: 
178: 
179: // CreateVoucherBatch generates vouchers
180: // @Summary create voucher batch
181: // @Tags Voucher
182: // @Param batch body VoucherBatchRequest true "Batch info"
183: // @Success 201 {object} domain.VoucherBatch
184: // @Router /api/v1/voucher-batches [post]
185: func CreateVoucherBatch(c echo.Context) error {
186: 	var req VoucherBatchRequest
187: 	if err := c.Bind(&req); err != nil {
188: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
189: 	}
190: 
191: 	if err := c.Validate(&req); err != nil {
192: 		return err
193: 	}
194: 
195: 	if req.Length == 0 {
196: 		req.Length = 10
197: 	}
198: 
199: 	productID, _ := strconv.ParseInt(req.ProductID, 10, 64)
200: 
201: 	// Validate Product
202: 	var product domain.Product
203: 	if err := GetDB(c).First(&product, productID).Error; err != nil {
204: 		return fail(c, http.StatusBadRequest, "INVALID_PRODUCT", "Product not found", nil)
205: 	}
206: 
207: 	// Securely get current user to enforce AgentID
208: 	currentUser, err := resolveOperatorFromContext(c)
209: 	if err != nil {
210: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
211: 	}
212: 
213: 	// If user is an agent, force the AgentID to be their own ID
214: 	if currentUser.Level == "agent" {
215: 		req.AgentID = fmt.Sprintf("%d", currentUser.ID)
216: 	}
217: 
218: 	// Start Transaction
219: 	tx := GetDB(c).Begin()
220: 
221: 	agentID, _ := strconv.ParseInt(req.AgentID, 10, 64)
222: 	var finalCost float64
223: 
224: 	// Agent Wallet Logic
225: 	if agentID > 0 {
226: 		finalCost = product.CostPrice * float64(req.Count)
227: 		// Fallback to retail price if cost price is not set
228: 		if finalCost <= 0 && product.Price > 0 {
229: 			finalCost = product.Price * float64(req.Count)
230: 		}
231: 
232: 		var wallet domain.AgentWallet
233: 		if err := tx.Set("gorm:query_option", "FOR UPDATE").FirstOrCreate(&wallet, domain.AgentWallet{AgentID: agentID}).Error; err != nil {
234: 			tx.Rollback()
235: 			return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to lock wallet", err.Error())
236: 		}
237: 
238: 		if wallet.Balance < finalCost {
239: 			tx.Rollback()
240: 			return fail(c, http.StatusPaymentRequired, "INSUFFICIENT_FUNDS", "Insufficient wallet balance", nil)
241: 		}
242: 
243: 		// Deduct Balance
244: 		newBalance := wallet.Balance - finalCost
245: 		if err := tx.Model(&domain.AgentWallet{}).Where("agent_id = ?", agentID).Updates(map[string]interface{}{"balance": newBalance, "updated_at": time.Now()}).Error; err != nil {
246: 			tx.Rollback()
247: 			return fail(c, http.StatusInternalServerError, "WALLET_UPDATE_FAILED", "Failed to update balance", err.Error())
248: 		}
249: 
250: 		// Log Transaction
251: 		log := domain.WalletLog{
252: 			AgentID:     agentID,
253: 			Type:        "purchase",
254: 			Amount:      -finalCost,
255: 			Balance:     newBalance,
256: 			ReferenceID: "batch-" + common.UUID(),
257: 			Remark:      fmt.Sprintf("generated %d vouchers from %s", req.Count, product.Name),
258: 			CreatedAt:   time.Now(),
259: 		}
260: 
261: 		// We'll update ReferenceID with BatchID after batch creation, or just use UUID
262: 		if err := tx.Create(&log).Error; err != nil {
263: 			tx.Rollback()
264: 			return fail(c, http.StatusInternalServerError, "LOG_FAILED", "Failed to create transaction log", err.Error())
265: 		}
266: 	}
267: 
268: 	batch := domain.VoucherBatch{
269: 		Name:           req.Name,
270: 		ProductID:      productID,
271: 		AgentID:        agentID,
272: 		Count:          req.Count,
273: 		Prefix:         req.Prefix,
274: 		Remark:         req.Remark,
275: 		GeneratePIN:    req.GeneratePIN,
276: 		PINLength:      req.PINLength,
277: 		ExpirationType: req.ExpirationType,
278: 		ValidityDays:   req.ValidityDays,
279: 		CreatedAt:      time.Now(),
280: 	}
281: 
282: 	// PrintExpireTime controls when vouchers can no longer be printed/activated
283: 	// This is separate from the actual validity period which comes from Product
284: 	if req.ExpireTime != "" {
285: 		if t, err := time.Parse(time.RFC3339, req.ExpireTime); err == nil {
286: 			batch.PrintExpireTime = &t
287: 		}
288: 	}
289: 
290: 	if err := tx.Create(&batch).Error; err != nil {
291: 		tx.Rollback()
292: 		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create batch record", err.Error())
293: 	}
294: 
295: 	// Generate Vouchers
296: 	vouchers := make([]domain.Voucher, 0, req.Count)
297: 
298: 	var expireTime time.Time
299: 	// For "fixed" expiration, calculate ExpireTime from Product.ValiditySeconds
300: 	// For "first_use" expiration, ExpireTime will be set when voucher is first used
301: 	if batch.ExpirationType != "first_use" && product.ValiditySeconds > 0 {
302: 		expireTime = time.Now().Add(time.Duration(product.ValiditySeconds) * time.Second)
303: 	}
304: 
305: 	// For first_use type, calculate expiry from validity days (will be applied on first use)
306: 	if batch.ExpirationType == "first_use" && batch.ValidityDays > 0 {
307: 		_ = int64(batch.ValidityDays * 24 * 60 * 60)
308: 	}
309: 
310: 	// Determine PIN length (default to 4 if not set)
311: 	pinLength := req.PINLength
312: 	if pinLength < 4 {
313: 		pinLength = 4
314: 	}
315: 
316: 	for i := 0; i < req.Count; i++ {
317: 		code := req.Prefix + common.GenerateVoucherCode(req.Length, req.Type)
318: 
319: 		voucher := domain.Voucher{
320: 			BatchID:     batch.ID,
321: 			Code:        code,
322: 			Status:      "unused",
323: 			Price:       product.Price,
324: 			AgentID:     agentID,
325: 			ExpireTime:  expireTime,
326: 			RequirePIN:  req.GeneratePIN,
327: 			CreatedAt:   time.Now(),
328: 			UpdatedAt:   time.Now(),
329: 
330: 			// Inherit allocations from Product
331: 			DataQuota: product.DataQuota,
332: 			TimeQuota: product.ValiditySeconds,
333: 		}
334: 
335: 		// For first_use type, leave ExpireTime empty - will be set on first login
336: 		if batch.ExpirationType == "first_use" {
337: 			voucher.ExpireTime = time.Time{}
338: 		}
339: 
340: 		// Generate PIN if required
341: 		if req.GeneratePIN {
342: 			voucher.PIN = common.GeneratePIN(pinLength)
343: 		}
344: 
345: 		vouchers = append(vouchers, voucher)
346: 	}
347: 
348: 	// Batch Insert
349: 	if err := tx.CreateInBatches(vouchers, 100).Error; err != nil {
350: 		tx.Rollback()
351: 		return fail(c, http.StatusInternalServerError, "GENERATE_FAILED", "Failed to insert vouchers", err.Error())
352: 	}
353: 
354: 	tx.Commit()
355: 
356: 	LogOperation(c, "create_voucher_batch", fmt.Sprintf("Created batch %s with %d vouchers", batch.Name, batch.Count))
357: 
358: 	return ok(c, batch)
359: }
360: 
361: // VoucherRedeemRequest
362: type VoucherRedeemRequest struct {
363: 	Code string `json:"code" validate:"required"`
364: 	PIN  string `json:"pin"` // Optional PIN for PIN-protected vouchers
365: }
366: 
367: // VoucherExtendRequest - add time to voucher
368: type VoucherExtendRequest struct {
369: 	Code         string `json:"code" validate:"required"`
370: 	ValidityDays int    `json:"validity_days" validate:"required,min=1,max=365"`
371: }
372: 
373: // VoucherTopupRequest - add data/time to active voucher
374: type VoucherTopupRequest struct {
375: 	VoucherCode string `json:"voucher_code" validate:"required"`
376: 	DataQuota   int64  `json:"data_quota"`   // Data quota in MB
377: 	TimeQuota   int64  `json:"time_quota"`   // Time quota in seconds
378: 	Price       float64 `json:"price"`       // Price paid for topup
379: }
380: 
381: // VoucherTopupListRequest - list topups for a voucher
382: type VoucherTopupListRequest struct {
383: 	VoucherCode string `json:"voucher_code" validate:"required"`
384: }
385: 
386: // VoucherSubscriptionRequest - create subscription for auto-renewal
387: type VoucherSubscriptionRequest struct {
388: 	VoucherCode  string `json:"voucher_code" validate:"required"`
389: 	ProductID    string `json:"product_id" validate:"required"`
390: 	IntervalDays int    `json:"interval_days" validate:"required,min=1,max=365"`
391: 	AutoRenew    bool   `json:"auto_renew"`
392: }
393: 
394: // VoucherBundleRequest - create a bundle of vouchers
395: type VoucherBundleRequest struct {
396: 	Name        string `json:"name" validate:"required"`
397: 	Description string `json:"description"`
398: 	ProductID   string `json:"product_id" validate:"required"`
399: 	VoucherCount int  `json:"voucher_count" validate:"required,min=1,max=100"`
400: 	Price       float64 `json:"price" validate:"required,min=0"`
401: }
402: 
403: // RedeemBundleRequest - redeem a bundle and get all voucher codes
404: type RedeemBundleRequest struct {
405: 	BundleCode string `json:"bundle_code" validate:"required"`
406: 	PIN        string `json:"pin"`
407: }
408: 
409: // RedeemVoucher activates a voucher and creates a Radius user
410: // @Summary redeem a voucher
411: // @Tags Voucher
412: // @Param redeem body VoucherRedeemRequest true "Redeem info"
413: // @Success 200 {object} domain.RadiusUser
414: // @Router /api/v1/vouchers/redeem [post]
415: func RedeemVoucher(c echo.Context) error {
416: 	var req VoucherRedeemRequest
417: 	if err := c.Bind(&req); err != nil {
418: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
419: 	}
420: 
421: 	if err := c.Validate(&req); err != nil {
422: 		return err
423: 	}
424: 
425: 	tx := GetDB(c).Begin()
426: 
427: 	// 1. Find Voucher
428: 	var voucher domain.Voucher
429: 	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("code = ?", req.Code).First(&voucher).Error; err != nil {
430: 		tx.Rollback()
431: 		return fail(c, http.StatusNotFound, "INVALID_CODE", "Voucher code not found", nil)
432: 	}
433: 
434: 	if voucher.Status != "unused" {
435: 		tx.Rollback()
436: 		return fail(c, http.StatusConflict, "VOUCHER_USED", "Voucher already used or expired", nil)
437: 	}
438: 
439: 	if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
440: 		tx.Rollback()
441: 		return fail(c, http.StatusConflict, "VOUCHER_EXPIRED", "Voucher has expired", nil)
442: 	}
443: 
444: 	// 1a. Validate PIN if required
445: 	if voucher.RequirePIN {
446: 		if req.PIN == "" {
447: 			tx.Rollback()
448: 			return fail(c, http.StatusBadRequest, "PIN_REQUIRED", "PIN is required for this voucher", nil)
449: 		}
450: 		if voucher.PIN != req.PIN {
451: 			tx.Rollback()
452: 			return fail(c, http.StatusUnauthorized, "INVALID_PIN", "Invalid PIN provided", nil)
453: 		}
454: 	}
455: 
456: 	// 2. Get Product and Profile
457: 	var product domain.Product
458: 	var batch domain.VoucherBatch
459: 	if err := tx.First(&batch, voucher.BatchID).Error; err != nil {
460: 		tx.Rollback()
461: 		return fail(c, http.StatusInternalServerError, "BATCH_ERROR", "Voucher batch not found", err.Error())
462: 	}
463: 
464: 	// Check if batch has expired for printing/activation
465: 	if batch.PrintExpireTime != nil && batch.PrintExpireTime.Before(time.Now()) {
466: 		tx.Rollback()
467: 		return fail(c, http.StatusConflict, "BATCH_EXPIRED", "Voucher batch has expired for printing/activation", nil)
468: 	}
469: 
470: 	if err := tx.First(&product, batch.ProductID).Error; err != nil {
471: 		tx.Rollback()
472: 		return fail(c, http.StatusInternalServerError, "PRODUCT_ERROR", "Associated product not found", err.Error())
473: 	}
474: 
475: 	var profile domain.RadiusProfile
476: 	if err := tx.First(&profile, product.RadiusProfileID).Error; err != nil {
477: 		tx.Rollback()
478: 		return fail(c, http.StatusInternalServerError, "PROFILE_ERROR", "Associated profile not found", err.Error())
479: 	}
480: 
481: 	// 3. Create Radius User
482: 	// User/Pass is the voucher code
483: 	now := time.Now()
484: 
485: 	// Handle first_use expiration type: calculate expiry based on batch validity days
486: 	expireTime := now.AddDate(1, 0, 0) // Default 1 year for first-use
487: 	if batch.ExpirationType == "first_use" {
488: 		// First-use expiration: Since we set FirstUsedAt = now below, we start the clock now.
489: 		if batch.ValidityDays > 0 {
490: 			expireTime = now.AddDate(0, 0, batch.ValidityDays)
491: 		} else {
492: 			expireTime, _ = time.Parse("2006-01-02", "9999-12-31")
493: 		}
494: 
495: 	} else {
496: 		// Fixed expiration: use product validity or batch expiry
497: 		if !voucher.ExpireTime.IsZero() {
498: 			expireTime = voucher.ExpireTime
499: 		} else {
500: 			expireTime = now.Add(time.Duration(product.ValiditySeconds) * time.Second)
501: 		}
502: 	}
503: 
504: 	// 519: user := domain.RadiusUser{ ... }
505: 	// We need to decide which rates to use.
506: 	// Use allocations from Voucher (inherited from Product at creation time)
507: 	userUpRate := product.UpRate
508: 	userDownRate := product.DownRate
509: 
510: 	// Use DataQuota from Voucher (inherited from Product at batch creation)
511: 	// This ensures quota is tied to the specific voucher, not the product
512: 	userDataQuota := voucher.DataQuota
513: 
514: 	if userUpRate == 0 {
515: 		userUpRate = profile.UpRate
516: 	}
517: 	if userDownRate == 0 {
518: 		userDownRate = profile.DownRate
519: 	}
520: 	if userDataQuota == 0 {
521: 		userDataQuota = profile.DataQuota
522: 	}
523: 
524: 	user := domain.RadiusUser{
525: 		Username:   voucher.Code,
526: 		Password:   voucher.Code,
527: 		ProfileId:  profile.ID,
528: 		Status:     "enabled",
529: 		ExpireTime: expireTime,
530: 		CreatedAt:  now,
531: 		UpdatedAt:  now,
532: 		UpRate:     userUpRate,
533: 		DownRate:   userDownRate,
534: 		DataQuota:  userDataQuota,
535: 		AddrPool:   profile.AddrPool,
536: 	}
537: 
538: 	if err := tx.Create(&user).Error; err != nil {
539: 		tx.Rollback()
540: 		return fail(c, http.StatusInternalServerError, "USER_CREATE_FAILED", "Failed to create user", err.Error())
541: 	}
542: 
543: 	// 4. Update Voucher Status
544: 	// Status "active" means the voucher has been redeemed and is in use
545: 	// This allows tracking of data/time usage against the voucher
546: 	voucher.Status = "active"
547: 	voucher.RadiusUsername = user.Username
548: 	voucher.ActivatedAt = now
549: 	voucher.FirstUsedAt = now // Record first use time for first_use expiration
550: 
551: 	if err := tx.Save(&voucher).Error; err != nil {
552: 		tx.Rollback()
553: 		return fail(c, http.StatusInternalServerError, "VOUCHER_UPDATE_FAILED", "Failed to update voucher status", err.Error())
554: 	}
555: 
556: 	tx.Commit()
557: 
558: 	LogOperation(c, "redeem_voucher", fmt.Sprintf("Redeemed voucher %s for user %s", voucher.Code, user.Username))
559: 
560: 	// Calculate commissions for agent hierarchy
561: 	if batch.AgentID != 0 {
562: 		if err := CalculateCommissions(GetDB(c), batch.AgentID, voucher.ID, voucher.Price); err != nil {
563: 			// Log error but don't fail the redemption
564: 			zap.L().Error("failed to calculate commissions",
565: 				zap.Error(err),
566: 				zap.Int64("voucher_id", voucher.ID),
567: 				zap.Int64("agent_id", batch.AgentID))
568: 		}
569: 	}
570: 
571: 	return ok(c, user)
572: }
573: 
574: // ExtendVoucher adds time to a used voucher by extending the associated user's expiration.
575: // It retrieves the voucher, validates it's in "used" status, finds the associated RadiusUser,
576: // extends the user's expiration time by the specified days, and logs the extension.
577: //
578: // Parameters:
579: //   - Code: The voucher code to extend
580: //   - ValidityDays: Number of days to add (1-365)
581: //
582: // Returns:
583: //   - Success: Extended user object with updated expiration
584: //   - Error: Voucher not found, not used, or user not found
585: //
586: // Authorization: Admin only (check current_operator level)
587: func ExtendVoucher(c echo.Context) error {
588: 	var req VoucherExtendRequest
589: 	if err := c.Bind(&req); err != nil {
590: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
591: 	}
592: 	if err := c.Validate(&req); err != nil {
593: 		return err
594: 	}
595: 
596: 	tx := GetDB(c).Begin()
597: 
598: 	var voucher domain.Voucher
599: 	if err := tx.Where("code = ?", req.Code).First(&voucher).Error; err != nil {
600: 		tx.Rollback()
601: 		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher code not found", nil)
602: 	}
603: 
604: 
605: 
606: 
607: 	// 2. Verify voucher status is "active"
608: 	if voucher.Status != "active" {
609: 		tx.Rollback()
610: 		return fail(c, http.StatusConflict, "INVALID_STATUS", "Voucher must be in 'active' status to extend", nil)
611: 	}
612: 
613: 	// 3. Find associated RadiusUser
614: 	var user domain.RadiusUser
615: 	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("username = ?", voucher.Code).First(&user).Error; err != nil {
616: 		tx.Rollback()
617: 		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "Associated user not found", err.Error())
618: 	}
619: 
620: 	// 4. Calculate new expiration time
621: 	now := time.Now()
622: 	var currentExpire time.Time
623: 
624: 	// If user already expired, count from now; otherwise extend from current expiration
625: 	if user.ExpireTime.Before(now) {
626: 		currentExpire = now
627: 	} else {
628: 		currentExpire = user.ExpireTime
629: 	}
630: 
631: 	newExpire := currentExpire.Add(time.Duration(req.ValidityDays) * 24 * time.Hour)
632: 
633: 	// 5. Update user's ExpireTime
634: 	if err := tx.Model(&domain.RadiusUser{}).Where("id = ?", user.ID).Update("expire_time", newExpire).Error; err != nil {
635: 		tx.Rollback()
636: 		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to extend user expiration", err.Error())
637: 	}
638: 
639: 	// 6. Update voucher extension tracking
640: 	voucher.ExtendedCount++
641: 	voucher.LastExtendedAt = now
642: 	if err := tx.Save(&voucher).Error; err != nil {
643: 		tx.Rollback()
644: 		return fail(c, http.StatusInternalServerError, "VOUCHER_UPDATE_FAILED", "Failed to update voucher extension count", err.Error())
645: 	}
646: 
647: 	tx.Commit()
648: 
649: 	LogOperation(c, "extend_voucher", fmt.Sprintf("Extended voucher %s by %d days", voucher.Code, req.ValidityDays))
650: 
651: 	// Return updated user
652: 	user.ExpireTime = newExpire
653: 	return ok(c, map[string]interface{}{
654: 		"voucher":       voucher,
655: 		"user":          user,
656: 		"old_expire":    currentExpire,
657: 		"new_expire":    newExpire,
658: 		"days_added":    req.ValidityDays,
659: 	})
660: }
661: 
662: // BulkActivateVouchers activates all vouchers in a batch
663: // BulkActivateVouchers activates all vouchers in a batch
664: func BulkActivateVouchers(c echo.Context) error {
665: 	id := c.Param("id")
666: 	batchID, _ := strconv.ParseInt(id, 10, 64)
667: 
668: 	db := GetDB(c)
669: 	var batch domain.VoucherBatch
670: 	if err := db.First(&batch, batchID).Error; err != nil {
671: 		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
672: 	}
673: 
674: 	var product domain.Product
675: 	if err := db.First(&product, batch.ProductID).Error; err != nil {
676: 		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
677: 	}
678: 
679: 	now := time.Now()
680: 	zap.L().Debug("bulk activate vouchers",
681: 		zap.Int64("batch_id", batchID),
682: 		zap.String("expiration_type", batch.ExpirationType),
683: 		zap.Int64("product_id", batch.ProductID),
684: 		zap.Int64("product_validity_seconds", product.ValiditySeconds),
685: 		zap.Time("batch_print_expire_time", func() time.Time { if batch.PrintExpireTime != nil { return *batch.PrintExpireTime }; return time.Time{} }()),
686: 	)
687: 	updates := map[string]interface{}{
688: 		"status":       "active",
689: 		"activated_at": now,
690: 	}
691: 
692: 	// Calculate expiration for fixed type
693: 	if batch.ExpirationType != "first_use" {
694: 		if batch.PrintExpireTime != nil && !batch.PrintExpireTime.IsZero() {
695: 			updates["expire_time"] = *batch.PrintExpireTime
696: 		} else if product.ValiditySeconds > 0 {
697: 			updates["expire_time"] = now.Add(time.Duration(product.ValiditySeconds) * time.Second)
698: 		} else {
699: 			// Default: set expire time to 30 days from now if no validity specified
700: 			updates["expire_time"] = now.AddDate(0, 0, 30)
701: 		}
702: 	} else {
703: 		// ensure expire_time is zero for first_use
704: 		updates["expire_time"] = time.Time{}
705: 	}
706: 
707: 	zap.L().Debug("bulk activate vouchers updates",
708: 		zap.Any("updates", updates),
709: 	)
710: 
711: 	// Update vouchers
712: 	if err := db.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", batchID, "unused").Updates(updates).Error; err != nil {
713: 		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to activate vouchers", err.Error())
714: 	}
715: 
716: 	// Also update the batch activation timestamp
717: 	if err := db.Model(&domain.VoucherBatch{}).Where("id = ?", batchID).Update("activated_at", &now).Error; err != nil {
718: 		zap.L().Error("Failed to update batch activated_at", zap.Int64("batch_id", batchID), zap.Error(err))
719: 		// Don't fail the request, vouchers are already activated
720: 	}
721: 
722: 	return ok(c, nil)
723: }
724: 
725: // BulkDeactivateVouchers deactivates all active vouchers in a batch
726: func BulkDeactivateVouchers(c echo.Context) error {
727: 	id := c.Param("id")
728: 
729: 	// Find all active vouchers in this batch to get their codes (usernames)
730: 	var vouchers []domain.Voucher
731: 	if err := GetDB(c).Where("batch_id = ? AND status = ?", id, "active").Find(&vouchers).Error; err != nil {
732: 		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query vouchers", err.Error())
733: 	}
734: 
735: 	// Disconnect online sessions for these vouchers
736: 	for _, v := range vouchers {
737: 		var session domain.RadiusOnline
738: 		// Check if user is online
739: 		if err := GetDB(c).Where("username = ?", v.Code).First(&session).Error; err == nil {
740: 			// User is online, disconnect them
741: 			if err := DisconnectSession(c, session); err != nil {
742: 				// Log error but continue with other vouchers
743: 				zap.L().Error("Failed to disconnect voucher user",
744: 					zap.String("username", v.Code),
745: 					zap.Error(err))
746: 			}
747: 		}
748: 	}
749: 
750: 	if err := GetDB(c).Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", id, "active").Update("status", "unused").Error; err != nil {
751: 		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to deactivate vouchers", err.Error())
752: 	}
753: 
754: 	// Also clear the batch activation timestamp
755: 	batchID, _ := strconv.ParseInt(id, 10, 64)
756: 	if err := GetDB(c).Model(&domain.VoucherBatch{}).Where("id = ?", id).Update("activated_at", nil).Error; err != nil {
757: 		zap.L().Error("Failed to clear batch activated_at", zap.Int64("batch_id", batchID), zap.Error(err))
758: 	}
759: 
760: 	return ok(c, nil)
761: }
762: 
763: // DeleteVoucherBatch soft deletes a batch and its vouchers
764: func DeleteVoucherBatch(c echo.Context) error {
765: 	id := c.Param("id")
766: 	tx := GetDB(c).Begin()
767: 	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ?", id).Update("is_deleted", true).Error; err != nil {
768: 		tx.Rollback()
769: 		return fail(c, http.StatusInternalServerError, "DELETE_VOUCHERS_FAILED", "Failed to delete vouchers", err.Error())
770: 	}
771: 	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", id).Update("is_deleted", true).Error; err != nil {
772: 		tx.Rollback()
773: 		return fail(c, http.StatusInternalServerError, "DELETE_BATCH_FAILED", "Failed to delete batch", err.Error())
774: 	}
775: 	tx.Commit()
776: 	return ok(c, nil)
777: }
778: 
779: // RestoreVoucherBatch restores a soft-deleted batch
780: func RestoreVoucherBatch(c echo.Context) error {
781: 	id := c.Param("id")
782: 	tx := GetDB(c).Begin()
783: 	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ?", id).Update("is_deleted", false).Error; err != nil {
784: 		tx.Rollback()
785: 		return fail(c, http.StatusInternalServerError, "RESTORE_VOUCHERS_FAILED", "Failed to restore vouchers", err.Error())
786: 	}
787: 	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", id).Update("is_deleted", false).Error; err != nil {
788: 		tx.Rollback()
789: 		return fail(c, http.StatusInternalServerError, "RESTORE_BATCH_FAILED", "Failed to restore batch", err.Error())
790: 	}
791: 	tx.Commit()
792: 	return ok(c, nil)
793: }
794: 
795: // TransferVouchersRequest - transfer vouchers between agents
796: type TransferVouchersRequest struct {
797: 	BatchID   string `json:"batch_id" validate:"required"`
798: 	ToAgentID string `json:"to_agent_id" validate:"required"`
799: }
800: 
801: // TransferVouchers transfers vouchers from one agent to another.
802: // Only unused vouchers can be transferred. Admin can transfer any vouchers,
803: // while agents can only transfer their own vouchers.
804: //
805: // Parameters:
806: //   - batch_id: ID of the voucher batch to transfer
807: //   - to_agent_id: ID of the target agent
808: //
809: // Returns:
810: //   - transferred_count: Number of vouchers transferred
811: func TransferVouchers(c echo.Context) error {
812: 	var req TransferVouchersRequest
813: 	if err := c.Bind(&req); err != nil {
814: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
815: 	}
816: 	if err := c.Validate(&req); err != nil {
817: 		return err
818: 	}
819: 
820: 	// Get current user for authorization
821: 	currentUser, err := resolveOperatorFromContext(c)
822: 	if err != nil {
823: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
824: 	}
825: 
826: 	batchID, _ := strconv.ParseInt(req.BatchID, 10, 64)
827: 	toAgentID, _ := strconv.ParseInt(req.ToAgentID, 10, 64)
828: 
829: 	db := GetDB(c)
830: 
831: 	// Verify target agent exists
832: 	var targetAgent domain.SysOpr
833: 	if err := db.First(&targetAgent, toAgentID).Error; err != nil {
834: 		return fail(c, http.StatusNotFound, "AGENT_NOT_FOUND", "Target agent not found", nil)
835: 	}
836: 
837: 	// Get the batch
838: 	var batch domain.VoucherBatch
839: 	if err := db.First(&batch, batchID).Error; err != nil {
840: 		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Voucher batch not found", nil)
841: 	}
842: 
843: 	// Authorization check
844: 	if currentUser.Level == "agent" {
845: 		// Agents can only transfer their own vouchers
846: 		if batch.AgentID != currentUser.ID {
847: 			return fail(c, http.StatusForbidden, "FORBIDDEN", "You can only transfer your own vouchers", nil)
848: 		}
849: 	}
850: 
851: 	// Check if transfer is to the same agent
852: 	if batch.AgentID == toAgentID {
853: 		return fail(c, http.StatusBadRequest, "SAME_AGENT", "Source and target agents are the same", nil)
854: 	}
855: 
856: 	tx := db.Begin()
857: 
858: 	// Count unused vouchers to transfer
859: 	var count int64
860: 	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ? AND is_deleted = ?", batchID, "unused", false).Count(&count).Error; err != nil {
861: 		tx.Rollback()
862: 		return fail(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count vouchers", err.Error())
863: 	}
864: 
865: 	if count == 0 {
866: 		tx.Rollback()
867: 		return fail(c, http.StatusConflict, "NO_VOUCHERS", "No unused vouchers to transfer", nil)
868: 	}
869: 
870: 	// Update vouchers to new agent
871: 	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ? AND is_deleted = ?", batchID, "unused", false).Updates(map[string]interface{}{"agent_id": toAgentID}).Error; err != nil {
872: 		tx.Rollback()
873: 		return fail(c, http.StatusInternalServerError, "TRANSFER_FAILED", "Failed to transfer vouchers", err.Error())
874: 	}
875: 
876: 	// Update batch agent_id
877: 	if err := tx.Model(&domain.VoucherBatch{}).Where("id = ?", batchID).Update("agent_id", toAgentID).Error; err != nil {
878: 		tx.Rollback()
879: 		return fail(c, http.StatusInternalServerError, "BATCH_UPDATE_FAILED", "Failed to update batch agent", err.Error())
880: 	}
881: 
882: 	tx.Commit()
883: 
884: 	zap.L().Info("Vouchers transferred",
885: 		zap.Int64("batch_id", batchID),
886: 		zap.Int64("from_agent", batch.AgentID),
887: 		zap.Int64("to_agent", toAgentID),
888: 		zap.Int64("count", count))
889: 
890: 	return ok(c, map[string]interface{}{
891: 		"transferred_count": count,
892: 		"batch_id":        batchID,
893: 		"from_agent":     batch.AgentID,
894: 		"to_agent":       toAgentID,
895: 	})
896: }
897: 
898: // RefundUnusedVouchers refunds unused vouchers to agent wallet
899: func RefundUnusedVouchers(c echo.Context) error {
900: 	id := c.Param("id")
901: 	db := GetDB(c)
902: 	zap.L().Info("Refund attempt", zap.String("batch_id", id))
903: 
904: 	var batch domain.VoucherBatch
905: 	if err := db.First(&batch, id).Error; err != nil {
906: 		zap.L().Error("Refund failed: batch not found", zap.String("batch_id", id), zap.Error(err))
907: 		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
908: 	}
909: 
910: 	if batch.AgentID <= 0 {
911: 		zap.L().Warn("Refund failed: no agent linked", zap.String("batch_id", id), zap.Int64("agent_id", batch.AgentID))
912: 		return fail(c, http.StatusBadRequest, "NO_AGENT", "This batch was not generated by an agent", nil)
913: 	}
914: 
915: 	tx := db.Begin()
916: 
917: 	var vouchers []domain.Voucher
918: 	if err := tx.Where("batch_id = ? AND status = ?", id, "unused").Find(&vouchers).Error; err != nil {
919: 		zap.L().Error("Refund failed: voucher query error", zap.String("batch_id", id), zap.Error(err))
920: 		tx.Rollback()
921: 		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to find unused vouchers", err.Error())
922: 	}
923: 
924: 	if len(vouchers) == 0 {
925: 		zap.L().Warn("Refund failed: no unused vouchers found", zap.String("batch_id", id))
926: 		tx.Rollback()
927: 		return fail(c, http.StatusBadRequest, "NO_UNUSED", "No unused vouchers to refund", nil)
928: 	}
929: 
930: 	var totalRefund float64
931: 	// Fetch product to get the cost price (same logic as creation)
932: 	var product domain.Product
933: 	if err := tx.First(&product, batch.ProductID).Error; err != nil {
934: 		zap.L().Error("Refund failed: product not found", zap.String("batch_id", id), zap.Int64("product_id", batch.ProductID), zap.Error(err))
935: 		tx.Rollback()
936: 		return fail(c, http.StatusInternalServerError, "PRODUCT_ERROR", "Associated product not found", err.Error())
937: 	}
938: 
939: 	unitCost := product.CostPrice
940: 	if unitCost <= 0 && product.Price > 0 {
941: 		unitCost = product.Price
942: 	}
943: 
944: 	totalRefund = unitCost * float64(len(vouchers))
945: 
946: 	// 1. Mark vouchers as deleted/refunded
947: 	if err := tx.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", id, "unused").Updates(map[string]interface{}{
948: 		"status":     "refunded",
949: 		"is_deleted": true,
950: 	}).Error; err != nil {
951: 		tx.Rollback()
952: 		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update voucher status", err.Error())
953: 	}
954: 
955: 	// 2. Refund Wallet
956: 	var wallet domain.AgentWallet
957: 	if err := tx.Set("gorm:query_option", "FOR UPDATE").FirstOrCreate(&wallet, domain.AgentWallet{AgentID: batch.AgentID}).Error; err != nil {
958: 		tx.Rollback()
959: 		return fail(c, http.StatusInternalServerError, "WALLET_ERROR", "Failed to lock wallet", err.Error())
960: 	}
961: 
962: 	newBalance := wallet.Balance + totalRefund
963: 	if err := tx.Model(&domain.AgentWallet{}).Where("agent_id = ?", batch.AgentID).Update("balance", newBalance).Error; err != nil {
964: 		tx.Rollback()
965: 		return fail(c, http.StatusInternalServerError, "REFUND_FAILED", "Failed to update wallet balance", err.Error())
966: 	}
967: 
968: 	// 3. Log Transaction
969: 	log := domain.WalletLog{
970: 		AgentID:     batch.AgentID,
971: 		Type:        "refund",
972: 		Amount:      totalRefund,
973: 		Balance:     newBalance,
974: 		ReferenceID: fmt.Sprintf("refund-%d", batch.ID),
975: 		Remark:      fmt.Sprintf("refunded %d unused vouchers (unit cost: %.2f) from batch %s", len(vouchers), unitCost, batch.Name),
976: 		CreatedAt:   time.Now(),
977: 	}
978: 	if err := tx.Create(&log).Error; err != nil {
979: 		tx.Rollback()
980: 		return fail(c, http.StatusInternalServerError, "LOG_FAILED", "Failed to create log", err.Error())
981: 	}
982: 
983: 	tx.Commit()
984: 	
985: 	LogOperation(c, "refund_vouchers", fmt.Sprintf("Refunded %d vouchers from batch %s", len(vouchers), batch.Name))
986: 	
987: 	return ok(c, map[string]interface{}{"refunded_count": len(vouchers), "refund_amount": totalRefund})
988: }
989: 
990: // ExportVoucherBatch exports batch vouchers to CSV
991: // ExportVoucherBatch exports batch vouchers to CSV
992: func ExportVoucherBatch(c echo.Context) error {
993: 	id := c.Param("id")
994: 
995: 	var batch domain.VoucherBatch
996: 	if err := GetDB(c).First(&batch, id).Error; err != nil {
997: 		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
998: 	}
999: 
1000: 	var product domain.Product
1001: 	if err := GetDB(c).First(&product, batch.ProductID).Error; err != nil {
1002: 		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
1003: 	}
1004: 
1005: 	var vouchers []domain.Voucher
1006: 	if err := GetDB(c).Where("batch_id = ?", id).Find(&vouchers).Error; err != nil {
1007: 		return fail(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to find vouchers", err.Error())
1008: 	}
1009: 
1010: 	nowStr := time.Now().Format("02012006")
1011: 	filename := fmt.Sprintf("%s-%s-%d-%s.csv", batch.Name, product.Name, batch.Count, nowStr)
1012: 	// Sanitize filename
1013: 	filename = strings.ReplaceAll(filename, " ", "_")
1014: 
1015: 	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
1016: 	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
1017: 	c.Response().WriteHeader(http.StatusOK)
1018: 
1019: 	writer := csv.NewWriter(c.Response())
1020: 	defer writer.Flush()
1021: 
1022: 	header := []string{
1023: 		"batch_id", "code", "radius_user", "status", "agent_id", "price",
1024: 		"activated_at", "expire_time", "extended_count", "last_extended_at",
1025: 		"is_deleted", "created_at", "updated_at",
1026: 	}
1027: 
1028: 	if err := writer.Write(header); err != nil {
1029: 		zap.L().Error("Failed to write CSV header", zap.Error(err))
1030: 		return fail(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to write CSV header", err.Error())
1031: 	}
1032: 
1033: 	formatTime := func(t time.Time) string {
1034: 		if t.IsZero() {
1035: 			return ""
1036: 		}
1037: 		return t.Format("2006-01-02 15:04:05")
1038: 	}
1039: 
1040: 	for _, v := range vouchers {
1041: 		row := []string{
1042: 			fmt.Sprintf("%d", v.BatchID),
1043: 			v.Code,
1044: 			v.RadiusUsername,
1045: 			v.Status,
1046: 			fmt.Sprintf("%d", v.AgentID),
1047: 			fmt.Sprintf("%.2f", v.Price),
1048: 			formatTime(v.ActivatedAt),
1049: 			formatTime(v.ExpireTime),
1050: 			fmt.Sprintf("%d", v.ExtendedCount),
1051: 			formatTime(v.LastExtendedAt),
1052: 			strconv.FormatBool(v.IsDeleted),
1053: 			formatTime(v.CreatedAt),
1054: 			formatTime(v.UpdatedAt),
1055: 		}
1056: 		if err := writer.Write(row); err != nil {
1057: 			zap.L().Error("Failed to write CSV row", zap.Int64("voucher_id", v.ID), zap.Error(err))
1058: 			return fail(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to write CSV row", err.Error())
1059: 		}
1060: 	}
1061: 	return nil
1062: }
1063: 
1064: // PrintVoucherBatch returns all vouchers for a batch without pagination limits
1065: // Used specifically for printing large batches
1066: // @Summary print voucher batch
1067: // @Tags Voucher
1068: // @Param id path int true "Batch ID"
1069: // @Success 200 {array} map[string]interface{}
1070: // @Router /api/v1/voucher-batches/{id}/print [get]
1071: func PrintVoucherBatch(c echo.Context) error {
1072: 	id := c.Param("id")
1073: 
1074: 	var batch domain.VoucherBatch
1075: 	if err := GetDB(c).First(&batch, id).Error; err != nil {
1076: 		return fail(c, http.StatusNotFound, "BATCH_NOT_FOUND", "Batch not found", err.Error())
1077: 	}
1078: 
1079: 	var vouchers []domain.Voucher
1080: 	if err := GetDB(c).Where("batch_id = ? AND is_deleted = ?", id, false).Order("id ASC").Find(&vouchers).Error; err != nil {
1081: 		return fail(c, http.StatusInternalServerError, "PRINT_FAILED", "Failed to fetch vouchers", err.Error())
1082: 	}
1083: 
1084: 	result := make([]map[string]interface{}, len(vouchers))
1085: 	for i, v := range vouchers {
1086: 		result[i] = map[string]interface{}{
1087: 			"id":          v.ID,
1088: 			"code":        v.Code,
1089: 			"status":      v.Status,
1090: 			"price":       v.Price,
1091: 			"expire_time": v.ExpireTime,
1092: 			"activated_at": v.ActivatedAt,
1093: 		}
1094: 	}
1095: 
1096: 	return c.JSON(http.StatusOK, result)
1097: }
1098: 
1099: // CreateVoucherTopup adds data/time quota to an active voucher
1100: // @Summary create voucher topup
1101: // @Tags Voucher
1102: // @Param topup body VoucherTopupRequest true "Topup info"
1103: // @Success 200 {object} domain.VoucherTopup
1104: // @Router /api/v1/vouchers/topup [post]
1105: func CreateVoucherTopup(c echo.Context) error {
1106: 	var req VoucherTopupRequest
1107: 	if err := c.Bind(&req); err != nil {
1108: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
1109: 	}
1110: 
1111: 	if err := c.Validate(&req); err != nil {
1112: 		return err
1113: 	}
1114: 
1115: 	db := GetDB(c)
1116: 
1117: 	// 1. Find the voucher
1118: 	var voucher domain.Voucher
1119: 	if err := db.Where("code = ?", req.VoucherCode).First(&voucher).Error; err != nil {
1120: 		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher not found", err.Error())
1121: 	}
1122: 
1123: 	// 2. Validate voucher is active
1124: 	if voucher.Status != "active" {
1125: 		return fail(c, http.StatusConflict, "VOUCHER_NOT_ACTIVE", "Voucher must be active to add topup", nil)
1126: 	}
1127: 
1128: 	// 3. Get current user to check agent permissions
1129: 	currentUser, err := resolveOperatorFromContext(c)
1130: 	if err != nil {
1131: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
1132: 	}
1133: 
1134: 	// Determine agent ID
1135: 	var agentID int64
1136: 	if currentUser.Level == "agent" {
1137: 		agentID = currentUser.ID
1138: 		// Agent can only add topup to their own vouchers
1139: 		if voucher.AgentID != agentID {
1140: 			return fail(c, http.StatusForbidden, "FORBIDDEN", "Cannot add topup to voucher not owned by you", nil)
1141: 		}
1142: 	}
1143: 
1144: 	// 4. Create topup record
1145: 	now := time.Now()
1146: 	topup := domain.VoucherTopup{
1147: 		VoucherID:   voucher.ID,
1148: 		VoucherCode: voucher.Code,
1149: 		AgentID:     agentID,
1150: 		DataQuota:   req.DataQuota,
1151: 		TimeQuota:   req.TimeQuota,
1152: 		Price:       req.Price,
1153: 		Status:      "active",
1154: 		ActivatedAt: now,
1155: 		CreatedAt:   now,
1156: 		UpdatedAt:   now,
1157: 	}
1158: 
1159: 	// 5. If there's a time quota, extend the expiry
1160: 	if req.TimeQuota > 0 {
1161: 		newExpireTime := voucher.ExpireTime.Add(time.Duration(req.TimeQuota) * time.Second)
1162: 		topup.ExpireTime = newExpireTime
1163: 		if err := db.Model(&voucher).Update("expire_time", newExpireTime).Error; err != nil {
1164: 			return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to extend voucher expiry", err.Error())
1165: 		}
1166: 	}
1167: 
1168: 	if err := db.Create(&topup).Error; err != nil {
1169: 		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create topup", err.Error())
1170: 	}
1171: 
1172: 	zap.L().Info("Voucher topup created",
1173: 		zap.String("voucher_code", voucher.Code),
1174: 		zap.Int64("data_quota", req.DataQuota),
1175: 		zap.Int64("time_quota", req.TimeQuota))
1176: 
1177: 	return ok(c, topup)
1178: }
1179: 
1180: // ListVoucherTopups lists all topups for a voucher
1181: // @Summary list voucher topups
1182: // @Tags Voucher
1183: // @Param voucher_code query string true "Voucher code"
1184: // @Success 200 {object} ListResponse
1185: // @Router /api/v1/vouchers/topups [get]
1186: func ListVoucherTopups(c echo.Context) error {
1187: 	voucherCode := c.QueryParam("voucher_code")
1188: 	if voucherCode == "" {
1189: 		return fail(c, http.StatusBadRequest, "VOUCHER_CODE_REQUIRED", "Voucher code is required", nil)
1190: 	}
1191: 
1192: 	db := GetDB(c)
1193: 	var topups []domain.VoucherTopup
1194: 	if err := db.Where("voucher_code = ?", voucherCode).Order("created_at DESC").Find(&topups).Error; err != nil {
1195: 		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query topups", err.Error())
1196: 	}
1197: 
1198: 	return ok(c, topups)
1199: }
1200: 
1201: // CreateVoucherSubscription creates a subscription for automatic voucher renewal
1202: // @Summary create voucher subscription
1203: // @Tags Voucher
1204: // @Param subscription body VoucherSubscriptionRequest true "Subscription info"
1205: // @Success 200 {object} domain.VoucherSubscription
1206: // @Router /api/v1/vouchers/subscriptions [post]
1207: func CreateVoucherSubscription(c echo.Context) error {
1208: 	var req VoucherSubscriptionRequest
1209: 	if err := c.Bind(&req); err != nil {
1210: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
1211: 	}
1212: 
1213: 	if err := c.Validate(&req); err != nil {
1214: 		return err
1215: 	}
1216: 
1217: 	db := GetDB(c)
1218: 
1219: 	// 1. Find the voucher
1220: 	var voucher domain.Voucher
1221: 	if err := db.Where("code = ?", req.VoucherCode).First(&voucher).Error; err != nil {
1222: 		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher not found", err.Error())
1223: 	}
1224: 
1225: 	// 2. Validate voucher is active
1226: 	if voucher.Status != "active" {
1227: 		return fail(c, http.StatusConflict, "VOUCHER_NOT_ACTIVE", "Voucher must be active to create subscription", nil)
1228: 	}
1229: 
1230: 	// 3. Parse product ID
1231: 	productID, err := strconv.ParseInt(req.ProductID, 10, 64)
1232: 	if err != nil {
1233: 		return fail(c, http.StatusBadRequest, "INVALID_PRODUCT", "Invalid product ID", err.Error())
1234: 	}
1235: 
1236: 	// 4. Check product exists
1237: 	var product domain.Product
1238: 	if err := db.First(&product, productID).Error; err != nil {
1239: 		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
1240: 	}
1241: 
1242: 	// 5. Get current user
1243: 	currentUser, err := resolveOperatorFromContext(c)
1244: 	if err != nil {
1245: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
1246: 	}
1247: 
1248: 	var agentID int64
1249: 	if currentUser.Level == "agent" {
1250: 		agentID = currentUser.ID
1251: 		if voucher.AgentID != agentID {
1252: 			return fail(c, http.StatusForbidden, "FORBIDDEN", "Cannot create subscription for voucher not owned by you", nil)
1253: 		}
1254: 	}
1255: 
1256: 	// 6. Check if subscription already exists
1257: 	var existing domain.VoucherSubscription
1258: 	if err := db.Where("voucher_code = ? AND status = ?", req.VoucherCode, "active").First(&existing).Error; err == nil {
1259: 		return fail(c, http.StatusConflict, "SUBSCRIPTION_EXISTS", "Active subscription already exists for this voucher", nil)
1260: 	}
1261: 
1262: 	now := time.Now()
1263: 	subscription := domain.VoucherSubscription{
1264: 		VoucherCode:    req.VoucherCode,
1265: 		ProductID:      productID,
1266: 		AgentID:        agentID,
1267: 		IntervalDays:   req.IntervalDays,
1268: 		Status:         "active",
1269: 		AutoRenew:      req.AutoRenew,
1270: 		LastRenewalAt:  now,
1271: 		NextRenewalAt:  now.Add(time.Duration(req.IntervalDays) * 24 * time.Hour),
1272: 		RenewalCount:   0,
1273: 		CreatedAt:      now,
1274: 		UpdatedAt:      now,
1275: 	}
1276: 
1277: 	if err := db.Create(&subscription).Error; err != nil {
1278: 		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create subscription", err.Error())
1279: 	}
1280: 
1281: 	zap.L().Info("Voucher subscription created",
1282: 		zap.String("voucher_code", req.VoucherCode),
1283: 		zap.Int("interval_days", req.IntervalDays),
1284: 		zap.Bool("auto_renew", req.AutoRenew))
1285: 
1286: 	return ok(c, subscription)
1287: }
1288: 
1289: // ListVoucherSubscriptions lists subscriptions for a voucher
1290: // @Summary list voucher subscriptions
1291: // @Tags Voucher
1292: // @Param voucher_code query string true "Voucher code"
1293: // @Success 200 {object} ListResponse
1294: // @Router /api/v1/vouchers/subscriptions [get]
1295: func ListVoucherSubscriptions(c echo.Context) error {
1296: 	voucherCode := c.QueryParam("voucher_code")
1297: 	if voucherCode == "" {
1298: 		return fail(c, http.StatusBadRequest, "VOUCHER_CODE_REQUIRED", "Voucher code is required", nil)
1299: 	}
1300: 
1301: 	db := GetDB(c)
1302: 	var subscriptions []domain.VoucherSubscription
1303: 	if err := db.Where("voucher_code = ?", voucherCode).Order("created_at DESC").Find(&subscriptions).Error; err != nil {
1304: 		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query subscriptions", err.Error())
1305: 	}
1306: 
1307: 	return ok(c, subscriptions)
1308: }
1309: 
1310: // CancelVoucherSubscription cancels an active subscription
1311: // @Summary cancel voucher subscription
1312: // @Tags Voucher
1313: // @Param id path string true "Subscription ID"
1314: // @Success 200 {object} domain.VoucherSubscription
1315: // @Router /api/v1/vouchers/subscriptions/{id}/cancel [post]
1316: func CancelVoucherSubscription(c echo.Context) error {
1317: 	id := c.Param("id")
1318: 	subID, err := strconv.ParseInt(id, 10, 64)
1319: 	if err != nil {
1320: 		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID", err.Error())
1321: 	}
1322: 
1323: 	db := GetDB(c)
1324: 
1325: 	var subscription domain.VoucherSubscription
1326: 	if err := db.Where("id = ? AND status = ?", subID, "active").First(&subscription).Error; err != nil {
1327: 		return fail(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Active subscription not found", err.Error())
1328: 	}
1329: 
1330: 	// Verify ownership
1331: 	currentUser, err := resolveOperatorFromContext(c)
1332: 	if err != nil {
1333: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
1334: 	}
1335: 
1336: 	if currentUser.Level == "agent" && subscription.AgentID != currentUser.ID {
1337: 		return fail(c, http.StatusForbidden, "FORBIDDEN", "Cannot cancel subscription not owned by you", nil)
1338: 	}
1339: 
1340: 	subscription.Status = "cancelled"
1341: 	subscription.UpdatedAt = time.Now()
1342: 
1343: 	if err := db.Save(&subscription).Error; err != nil {
1344: 		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to cancel subscription", err.Error())
1345: 	}
1346: 
1347: 	zap.L().Info("Voucher subscription cancelled",
1348: 		zap.Int64("subscription_id", subID),
1349: 		zap.String("voucher_code", subscription.VoucherCode))
1350: 
1351: 	return ok(c, subscription)
1352: }
1353: 
1354: // CreateVoucherBundle creates a bundle of vouchers
1355: // @Summary create voucher bundle
1356: // @Tags Voucher
1357: // @Param bundle body VoucherBundleRequest true "Bundle info"
1358: // @Success 200 {object} domain.VoucherBundle
1359: // @Router /api/v1/voucher-bundles [post]
1360: func CreateVoucherBundle(c echo.Context) error {
1361: 	var req VoucherBundleRequest
1362: 	if err := c.Bind(&req); err != nil {
1363: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request", err.Error())
1364: 	}
1365: 
1366: 	if err := c.Validate(&req); err != nil {
1367: 		return err
1368: 	}
1369: 
1370: 	db := GetDB(c)
1371: 
1372: 	// Parse product ID
1373: 	productID, err := strconv.ParseInt(req.ProductID, 10, 64)
1374: 	if err != nil {
1375: 		return fail(c, http.StatusBadRequest, "INVALID_PRODUCT", "Invalid product ID", err.Error())
1376: 	}
1377: 
1378: 	// Check product exists
1379: 	var product domain.Product
1380: 	if err := db.First(&product, productID).Error; err != nil {
1381: 		return fail(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
1382: 	}
1383: 
1384: 	// Get current user
1385: 	currentUser, err := resolveOperatorFromContext(c)
1386: 	if err != nil {
1387: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
1388: 	}
1389: 
1390: 	var agentID int64
1391: 	if currentUser.Level == "agent" {
1392: 		agentID = currentUser.ID
1393: 	}
1394: 
1395: 	now := time.Now()
1396: 	// Generate unique bundle code
1397: 	bundleCode := "BUNDLE-" + common.UUID()[:8]
1398: 
1399: 	bundle := domain.VoucherBundle{
1400: 		Name:        req.Name,
1401: 		Description: req.Description,
1402: 		AgentID:     agentID,
1403: 		ProductID:   productID,
1404: 		VoucherCount: req.VoucherCount,
1405: 		Price:       req.Price,
1406: 		Status:      "active",
1407: 		CreatedAt:   now,
1408: 		UpdatedAt:   now,
1409: 	}
1410: 
1411: 	if err := db.Create(&bundle).Error; err != nil {
1412: 		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create bundle", err.Error())
1413: 	}
1414: 
1415: 	// Generate vouchers for the bundle
1416: 	vouchers, err := generateVouchersForBundle(db, product, bundle, req.VoucherCount, agentID)
1417: 	if err != nil {
1418: 		db.Delete(&bundle)
1419: 		return fail(c, http.StatusInternalServerError, "GENERATE_FAILED", "Failed to generate vouchers", err.Error())
1420: 	}
1421: 
1422: 	zap.L().Info("Voucher bundle created",
1423: 		zap.String("name", req.Name),
1424: 		zap.Int("voucher_count", req.VoucherCount),
1425: 		zap.Float64("price", req.Price))
1426: 
1427: 	return ok(c, map[string]interface{}{
1428: 		"bundle":      bundle,
1429: 		"bundle_code": bundleCode,
1430: 		"vouchers":    vouchers,
1431: 	})
1432: }
1433: 
1434: // generateVouchersForBundle generates vouchers for a bundle
1435: func generateVouchersForBundle(db *gorm.DB, product domain.Product, bundle domain.VoucherBundle, count int, agentID int64) ([]domain.Voucher, error) {
1436: 	vouchers := make([]domain.Voucher, 0, count)
1437: 
1438: 	for i := 0; i < count; i++ {
1439: 		code := common.GenerateVoucherCode(12, "mixed")
1440: 
1441: 		voucher := domain.Voucher{
1442: 			BatchID:    bundle.ID, // Link to bundle
1443: 			Code:       code,
1444: 			Status:     "unused",
1445: 			Price:      product.Price,
1446: 			AgentID:    agentID,
1447: 			ExpireTime: time.Time{},
1448: 			CreatedAt:  time.Now(),
1449: 			UpdatedAt:  time.Now(),
1450: 		}
1451: 		vouchers = append(vouchers, voucher)
1452: 	}
1453: 
1454: 	if err := db.CreateInBatches(vouchers, 100).Error; err != nil {
1455: 		return nil, err
1456: 	}
1457: 
1458: 	return vouchers, nil
1459: }
1460: 
1461: // ListVoucherBundles lists all voucher bundles
1462: // @Summary list voucher bundles
1463: // @Tags Voucher
1464: // @Success 200 {object} ListResponse
1465: // @Router /api/v1/voucher-bundles [get]
1466: func ListVoucherBundles(c echo.Context) error {
1467: 	db := GetDB(c)
1468: 
1469: 	var bundles []domain.VoucherBundle
1470: 	query := db.Model(&domain.VoucherBundle{})
1471: 
1472: 	// Filter by agent
1473: 	currentUser, err := resolveOperatorFromContext(c)
1474: 	if err != nil {
1475: 		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed", err.Error())
1476: 	}
1477: 
1478: 	if currentUser.Level == "agent" {
1479: 		query = query.Where("agent_id = ?", currentUser.ID)
1480: 	}
1481: 
1482: 	if err := query.Order("created_at DESC").Find(&bundles).Error; err != nil {
1483: 		return fail(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to query bundles", err.Error())
1484: 	}
1485: 
1486: 	return ok(c, bundles)
1487: }
1488: 
1489: // PublicVoucherStatus returns public voucher status without authentication
1490: // @Summary get public voucher status
1491: // @Tags Voucher
1492: // @Param code query string true "Voucher code"
1493: // @Success 200 {object} map[string]interface{}
1494: // @Router /public/vouchers/status [get]
1495: func PublicVoucherStatus(c echo.Context) error {
1496: 	code := c.QueryParam("code")
1497: 	if code == "" {
1498: 		return fail(c, http.StatusBadRequest, "CODE_REQUIRED", "Voucher code is required", nil)
1499: 	}
1500: 
1501: 	db := GetDB(c)
1502: 
1503: 	var voucher domain.Voucher
1504: 	if err := db.Where("code = ?", code).First(&voucher).Error; err != nil {
1505: 		return fail(c, http.StatusNotFound, "VOUCHER_NOT_FOUND", "Voucher not found", nil)
1506: 	}
1507: 
1508: 	// Return only public information
1509: 	return ok(c, map[string]interface{}{
1510: 		"status":       voucher.Status,
1511: 		"expire_time":  voucher.ExpireTime,
1512: 		"activated_at": voucher.ActivatedAt,
1513: 	})
1514: }
1515: 
1516: func registerVoucherRoutes() {
1517: 	webserver.ApiGET("/voucher-batches", ListVoucherBatches)

(Output capped at 50 KB. Showing lines 1-1517. Use offset=1518 to continue.)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 16.3s)

_Thinking:_

This is a very long file. Let me summarize what I've found so far:

1. The voucher system has:
   - `Voucher` model with fields like `Code`, `RadiusUsername`, `Status`, etc.
   - `VoucherBatch` model for batches
   - A redemption process (`RedeemVoucher`) that creates a `RadiusUser` when a voucher is redeemed

2. The key issue is in the `RedeemVoucher` function (line 415-572):
   - When a voucher is redeemed, it **creates a new RadiusUser** (line 524-536)
   - The RadiusUser's username and password are set to the voucher code
   - This means the voucher is converted to a user for authentication

3. The authentication flow (in `radius.go` and `auth_stages.go`):
   - `stageLoadUser` calls `GetValidUser(ctx.Username, ctx.IsMacAuth)` 
   - This only queries the `UserRepo`, NOT the `VoucherRepo`
   - Vouchers are NOT directly authenticated - they're converted to users first via redemption

Wait - I see that there are two types of vouchers:
1. **Redeemed vouchers** - these become RadiusUser entries and work fine
2. **Non-redeemed vouchers** - these are just voucher codes that haven't been converted yet

Let me check what the user means - they say vouchers don't authenticate. Let me check if there's a direct voucher authentication mode or if they need to use the "redeem" API first.

Actually, looking at the `RedeemVoucher` function, it seems that vouchers need to be **redeemed first** before they can authenticate. The user would:
1. Get a voucher code (e.g., "VOUCHER123")
2. Call the redeem API to convert it to a user
3. Then use the same code as username/password to authenticate

Let me check if there's a "direct" authentication mode for vouchers (without redemption).


The user wants vouchers to authenticate directly without needing to redeem them first. Looking at the domain model, there's a `RadiusUsername` field on vouchers that gets set during redemption, and the voucher repository has a `GetByCode` method that could theoretically support direct authentication. The problem is the current authentication pipeline doesn't use the voucher repository at all - it only checks the user repository.

The implementation would need to either modify the auth stage to check vouchers as a fallback, create a plugin to handle voucher authentication, or introduce a hybrid lookup approach. Before diving deeper into that, I should verify what the user actually wants: whether vouchers should authenticate without requiring an explicit redemption step first.


Now I have a clear picture of the issue. Let me verify my understanding by checking if there's any configuration or alternative authentication path for vouchers:

**Tool: grep**

**Input:**
```json
{
  "include": "*.go",
  "pattern": "VoucherRepo|GetByCode"
}
```

**Output:**
```
Found 20 matches
/home/faris/Downloads/toughradius/toughradius/internal/radiusd/radius.go:
  Line 82: 	VoucherRepo    repository.VoucherRepository
  Line 108: 		VoucherRepo:    repogorm.NewGormVoucherRepository(db),

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/integration_test.go:
  Line 128: 	plugins.InitPlugins(appCtx, appCtx.DB(), radiusService.SessionRepo, radiusService.AccountingRepo, radiusService.VoucherRepo, radiusService.UserRepo)

/home/faris/Downloads/toughradius/toughradius/main.go:
  Line 97: 	plugins.InitPlugins(application, application.DB(), radiusService.SessionRepo, radiusService.AccountingRepo, radiusService.VoucherRepo, radiusService.UserRepo)

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/plugins/init.go:
  Line 26: 	voucherRepo repository.VoucherRepository,

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/interfaces.go:
  Line 81: // VoucherRepository manages vouchers
  Line 82: type VoucherRepository interface {
  Line 83: 	// GetByCode finds a voucher by code
  Line 84: 	GetByCode(ctx context.Context, code string) (*domain.Voucher, error)

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/plugins/auth/checkers/first_use_activator.go:
  Line 15: 	voucherRepo repository.VoucherRepository
  Line 21: 	voucherRepo repository.VoucherRepository,
  Line 53: 	voucher, err := c.voucherRepo.GetByCode(ctx, user.Username)

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/repository/gorm/voucher_repository.go:
  Line 12: // GormVoucherRepository is the GORM implementation of the voucher repository
  Line 13: type GormVoucherRepository struct {
  Line 17: // NewGormVoucherRepository creates a voucher repository instance
  Line 18: func NewGormVoucherRepository(db *gorm.DB) repository.VoucherRepository {
  Line 19: 	return &GormVoucherRepository{db: db}
  Line 22: func (r *GormVoucherRepository) GetByCode(ctx context.Context, code string) (*domain.Voucher, error) {
  Line 31: func (r *GormVoucherRepository) UpdateFirstUsedAt(ctx context.Context, code string, firstUsedAt, expireTime time.Time) error {
  Line 42: func (r *GormVoucherRepository) GetBatchByID(ctx context.Context, batchID int64) (*domain.VoucherBatch, error) {
```

---

## Assistant (Plan · minimax-m2.5-free · 4.5s)

_Thinking:_

Interesting! There's a plugin `first_use_activator.go` that uses the VoucherRepo. Let me check what that plugin does - it might be related to voucher authentication. Also, I see that VoucherRepo exists but is only used in one plugin.

Let me check the first_use_activator plugin:

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/radiusd/plugins/auth/checkers/first_use_activator.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/radiusd/plugins/auth/checkers/first_use_activator.go</path>
<type>file</type>
<content>1: package checkers
2: 
3: import (
4: 	"context"
5: 	"time"
6: 
7: 	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
8: 	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
9: 	"go.uber.org/zap"
10: )
11: 
12: // FirstUseActivator checks if a user is logging in for the first time via a first-use voucher
13: // and activates the expiration countdown.
14: type FirstUseActivator struct {
15: 	voucherRepo repository.VoucherRepository
16: 	userRepo    repository.UserRepository
17: }
18: 
19: // NewFirstUseActivator creates a first-use activator instance
20: func NewFirstUseActivator(
21: 	voucherRepo repository.VoucherRepository,
22: 	userRepo repository.UserRepository,
23: ) *FirstUseActivator {
24: 	return &FirstUseActivator{
25: 		voucherRepo: voucherRepo,
26: 		userRepo:    userRepo,
27: 	}
28: }
29: 
30: func (c *FirstUseActivator) Name() string {
31: 	return "first_use_activator"
32: }
33: 
34: func (c *FirstUseActivator) Order() int {
35: 	// Execute before ExpireChecker (which is 10) to ensure we update expiration before checking it
36: 	return 5
37: }
38: 
39: func (c *FirstUseActivator) Check(ctx context.Context, authCtx *auth.AuthContext) error {
40: 	user := authCtx.User
41: 	if user == nil {
42: 		return nil
43: 	}
44: 
45: 	// Check if this is a "pending activation" user
46: 	// We set ExpireTime to year 9999 in RedeemVoucher for first_use vouchers
47: 	if user.ExpireTime.Year() < 9999 {
48: 		// Already activated or not a first-use voucher
49: 		return nil
50: 	}
51: 
52: 	// Double check by looking up the voucher
53: 	voucher, err := c.voucherRepo.GetByCode(ctx, user.Username)
54: 	if err != nil {
55: 		// Not found or error -> ignore, treat as normal user
56: 		return nil
57: 	}
58: 
59: 	// Get batch to confirm type and validity
60: 	batch, err := c.voucherRepo.GetBatchByID(ctx, voucher.BatchID)
61: 	if err != nil {
62: 		zap.L().Error("first_use_activator: batch not found",
63: 			zap.String("username", user.Username),
64: 			zap.Int64("batch_id", voucher.BatchID),
65: 			zap.Error(err))
66: 		return nil
67: 	}
68: 
69: 	if batch.ExpirationType != "first_use" {
70: 		// Should generally not happen if ExpireTime is 9999, but safe to ignore
71: 		return nil
72: 	}
73: 
74: 	// Calculate new expiration
75: 	now := time.Now()
76: 	validityDuration := time.Duration(batch.ValidityDays) * 24 * time.Hour
77: 	newExpire := now.Add(validityDuration)
78: 
79: 	// Update RadiusUser
80: 	if err := c.userRepo.UpdateField(ctx, user.Username, "expire_time", newExpire); err != nil {
81: 		zap.L().Error("first_use_activator: failed to update user expiration",
82: 			zap.String("username", user.Username),
83: 			zap.Error(err))
84: 		// If DB update fails, we should probably return error to prevent free access
85: 		return err
86: 	}
87: 
88: 	// Update Voucher
89: 	if err := c.voucherRepo.UpdateFirstUsedAt(ctx, voucher.Code, now, newExpire); err != nil {
90: 		zap.L().Error("first_use_activator: failed to update voucher",
91: 			zap.String("code", voucher.Code),
92: 			zap.Error(err))
93: 		// Log error but proceed since user is updated
94: 	}
95: 
96: 	// Update the user object in the current context so subsequent checkers see the correct time
97: 	user.ExpireTime = newExpire
98: 
99: 	zap.L().Info("first_use_activator: voucher activated on first login",
100: 		zap.String("username", user.Username),
101: 		zap.Time("activated_at", now),
102: 		zap.Time("new_expire", newExpire))
103: 
104: 	return nil
105: }

(End of file - total 105 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 46.6s)

_Thinking:_

Now I understand the architecture better. Let me summarize my findings:

## Current Architecture

1. **Users (from User Management tab)**: These are `RadiusUser` records in the database. They are directly queried during authentication via `UserRepo.GetByUsername()` in the `stageLoadUser` authentication stage.

2. **Vouchers**: There are two workflows:
   - **Redeemed vouchers**: When you call the `/api/v1/vouchers/redeem` API, it creates a new `RadiusUser` record with the same username/password as the voucher code. After redemption, the voucher authenticates like a normal user.
   - **Non-redeemed vouchers**: These are just `Voucher` records that have NOT been converted to `RadiusUser` yet. They don't exist in the `radcheck` table (which stores user credentials).

3. **The Problem**: The authentication pipeline ONLY looks up users from `UserRepo`. It never checks the `VoucherRepo` for direct voucher authentication. So if a user tries to authenticate with a voucher code that hasn't been redeemed, authentication fails because there's no matching `RadiusUser`.

## The Question for the User

I need to clarify what the user wants:
1. **Option A**: Keep the redemption workflow - users must redeem vouchers first before using them (current behavior). This is more controlled.
2. **Option B**: Allow direct authentication with voucher codes without redemption - the system should look up vouchers and authenticate them directly.

Based on the user's question, they want vouchers to authenticate "like users" without needing a separate redemption step. This is **Option B**.

Now let me ask a clarifying question about this.

Now I have a clear picture. Let me ask you a clarifying question:

---

## Current State Analysis

I've identified the root cause:

1. **Users from User Management**: These are stored as `RadiusUser` records and are directly queried during RADIUS authentication via `UserRepo.GetByUsername()` in `internal/radiusd/auth_stages.go:121`

2. **Vouchers**: The current system requires a **two-step process**:
   - **Step 1**: Call `/api/v1/vouchers/redeem` API to convert voucher to a `RadiusUser`
   - **Step 2**: Then authenticate with the voucher code

3. **The Problem**: The authentication pipeline (`stageLoadUser`) only queries `UserRepo` - it never checks `VoucherRepo`. So vouchers that haven't been redeemed cannot authenticate.

---

**Question**: How do you want vouchers to work?

**Option A** - Keep redemption workflow (current): Users must call the redeem API first, then authenticate. This gives you control over when vouchers become active.

**Option B** - Direct voucher authentication: Voucher codes authenticate directly without redemption. The system checks vouchers first, then falls back to regular users.

Which approach do you prefer? Option B would require modifying the authentication pipeline to also check the `VoucherRepo`.

---

## User

I want to keep control in voucher , for example i want sometime to deactivate all the voucher of a batch abc if the agent did not pay, even if it is active now. so any voucher under this abc batch has no access to my network. do you understand me ?

---

## Assistant (Plan · minimax-m2.5-free · 20.3s)

_Thinking:_

The user wants to keep control over vouchers - they want to be able to deactivate/activate entire batches of vouchers remotely (e.g., via MikroTik CoA commands) even if the vouchers have been "redeemed" (i.e., converted to RadiusUser).

This is actually a good use case. The current system already supports this partially:
1. BulkActivateVouchers - activates all vouchers in a batch
2. BulkDeactivateVouchers - deactivates all vouchers in a batch

But the user wants to be able to control access via the MikroTik router - which means they need the RADIUS authentication to respect the voucher status.

Currently, after redemption:
- Voucher status is set to "active"
- A RadiusUser is created with the same username/password as the voucher code

The problem is: once the RadiusUser is created, the RADIUS auth only checks the RadiusUser table - it doesn't check the voucher status anymore. So even if you "deactivate" vouchers in ToughRADIUS, the RadiusUser still exists and can authenticate.

The user wants to be able to:
1. Deactivate a batch of vouchers remotely
2. The MikroTik should reject authentication for those users

This is a valid requirement. The solution would be to:
1. Keep checking voucher status during authentication even after redemption
2. OR use CoA (Change of Authorization) to disconnect users when vouchers are deactivated

Let me confirm my understanding with the user and propose a solution.

The key insight is:
- Users created in User Management: pure RadiusUser records
- Vouchers: Voucher records + RadiusUser records (after redemption)

The user wants:
- Vouchers that are "active" should be able to authenticate
- Vouchers that are "deactivated" (via BulkDeactivateVouchers) should NOT authenticate

Currently, BulkDeactivateVouchers changes voucher status from "active" to "unused" in the voucher table, but the RadiusUser still exists and can authenticate.
 
The best approach would be checking both the RadiusUser table and the voucher status during authentication. If a corresponding voucher exists and its status isn't "active", reject the authentication request. This ensures deactivated vouchers can't access the network, even if their RadiusUser record remains. I'll propose this solution to the user.

Yes, I understand perfectly! You want **batch-level control** over vouchers even after they're redeemed.

Here's the issue with the current flow:

1. When a voucher is **redeemed** (`/api/v1/vouchers/redeem`):
   - Voucher status becomes "active" 
   - A `RadiusUser` is created with the same username/password as the voucher code

2. When you **deactivate a batch** (`BulkDeactivateVouchers`):
   - Voucher status changes from "active" → "unused" in the voucher table
   - **BUT** the `RadiusUser` still exists in the database!

3. **The problem**: RADIUS auth only checks `RadiusUser` table - it doesn't know the voucher was deactivated, so users can still authenticate.

---

## Solution Approach

Modify the RADIUS authentication pipeline to also verify the **voucher status** when authenticating. Here's the logic:

```
During authentication:
1. Check if username exists in RadiusUser table
2. If found, also check if this user came from a voucher
3. If it's a voucher user → verify voucher status is "active"
4. If voucher is NOT active → reject authentication
```

This way:
- When you deactivate a batch → voucher status becomes "unused"
- Users from that batch → immediately lose network access
- When you reactivate → they regain access

---

**Questions for you:**

1. Should we also handle "expired" vouchers (quota exceeded or time expired)? 
2. Do you want the existing CoA (Change of Authorization) to also work for disconnecting already-online voucher users when you deactivate a batch?

---

## User

1. yes i want  to handle them, you can suggest better way that increase the performance of netwrk and not overwhelming my mikrotik device resources, 2.for already-online voucher i want to disconnect them but with their used qouta in time or data, because if the agent pay I will activate the batch and the voucher user will continue to consume remaining qouta /time or both and not to get full qouta/time when activate again. so please prepare full roubust solution for all of this, also when make batch active it means an auto redemption should occur directly. or redemption not needed because i feel it's extra unnecesary  complications , or if you recommend to retain it please what is the beneift of it ?

---

## Assistant (Plan · minimax-m2.5-free · 62.7s)

_Thinking:_

The user wants a comprehensive solution for voucher-based authentication with:

1. **Batch-level control** - ability to deactivate entire batch (even after redemption)
2. **Handle expired vouchers** - quota exceeded or time expired
3. **Performance optimization** - don't overwhelm MikroTik
4. **Disconnect already-online users** when batch deactivated, but preserve their used quota
5. **Auto-redemption on batch activation** - no manual redemption needed
6. **Retention consideration** - whether to keep the redemption step or remove it

Let me think through the architecture:

## Current Flow:
1. Create voucher batch → vouchers created with status "unused"
2. Redeem voucher → creates RadiusUser, voucher status becomes "active"
3. Authenticate → checks RadiusUser only
4. Deactivate batch → voucher status changes to "unused", but RadiusUser remains

## Proposed Flow:
1. Create voucher batch → vouchers created with status "unused"
2. **Activate batch** → vouchers status becomes "active" (auto-redemption happens here)
3. Authenticate → checks RadiusUser AND verifies voucher status is "active"
4. Deactivate batch → voucher status becomes "unused", disconnect online users
5. Reactivate batch → voucher status becomes "active", users can reconnect with remaining quota

## Benefits of keeping redemption step vs removing it:
The user is asking if we should keep the redemption step. Let me think:

**Current workflow with redemption:**
- Customer buys voucher → gets code
- Customer goes to portal → redeems voucher → becomes user
- Admin can control via batch deactivation

**Proposed workflow without redemption (direct auth):**
- Customer buys voucher → gets code
- Customer connects to Mikrotik → authenticates with voucher code directly
- No "redemption" step needed

The user seems to prefer removing redemption since they said "auto redemption should occur directly" and "redemption not needed because i feel it's extra unnecessary complications".


Removing redemption simplifies the customer experience significantly. The primary advantage of keeping it was controlling when a voucher becomes active, but batch activation already provides this capability. With auto-redemption on batch activation, customers can authenticate immediately after purchase without needing to visit a portal first.

For implementation, I'd store voucher status in Redis for quick lookups during authentication instead of hitting the database on every request. This keeps the RadiusUser intact in the database as a persistent record while using Redis as a fast cache layer to determine if a voucher is currently active. When the user authenticates, the RADIUS server checks Redis for voucher status—if it's not cached, it queries the database and then caches the result.

This approach handles several key requirements: maintaining quota data in the voucher table even after users are deactivated so it persists across reactivation cycles, automatically disconnecting users through CoA when their voucher is deactivated, and providing batch-level controls to deactivate or reactivate entire groups of vouchers simultaneously.

However, there's a question about where the quota consumption data lives. Since the user is deactivated but still has a RadiusUser record, that record likely contains the usage counters. When reactivating, those counters need to be preserved from the moment of deactivation rather than resetting. The voucher table also tracks usage separately, so there's potential redundancy to consider.

My approach would be: keep the RadiusUser record intact but inactive during deactivation, allow the usage data to persist in both tables, and upon reactivation the user reconnects with their remaining quota. The challenge is handling the reconnection logic—does it happen automatically when they authenticate again, or do we need an explicit reactivation step? Looking at the existing implementation, the RadiusUser gets created during the redemption phase.

When BulkActivateVouchers runs, it sets the voucher status to "active" but leaves the RadiusUser untouched. The user can then authenticate and resume with their existing quota. But when BulkDeactivateVouchers executes, it disconnects any active sessions and resets the voucher status back to "unused"—the RadiusUser remains in the database but can't authenticate anymore. Reactivating the batch sets the voucher back to "active" and they regain access with their remaining quota intact.

The implementation would need to modify the authentication pipeline to verify voucher status directly rather than requiring a separate redemption step, optimize queries to cache active voucher batches, handle the activation process without creating redundant RadiusUser records, and ensure CoA properly terminates sessions during deactivation.

I understand your requirements clearly. Let me analyze and propose a comprehensive solution.

## Current Architecture Issue

Currently:
- **Redeem** = creates `RadiusUser` + activates voucher
- **Auth** = only checks `RadiusUser` (ignores voucher status)
- **Deactivate** = changes voucher status, but RadiusUser still works!

## Proposed Solution

### 1. Remove Redemption Step - Direct Authentication

**Benefit of removing redemption:**
- Simpler flow: Create batch → Activate → Users connect directly
- No customer portal needed
- Faster customer experience
- Less API complexity

The only reason to keep redemption was "control when to activate" - but **batch activation already provides this control**.

### 2. Authentication Flow (Modified)

```
RADIUS Auth Request:
1. Query RadiusUser by username
2. If found → Check if this username has a related voucher (via RadiusUsername)
3. If voucher exists → Verify voucher status is "active"
4. If voucher expired (quota/time) → Reject
5. If voucher deactivated → Reject
6. Otherwise → Allow
```

### 3. Performance Optimization

Instead of querying database on every auth request:
- **Cache active voucher batch IDs in memory** (simple in-memory cache)
- On batch activate/deactivate → Update the cache
- Auth check becomes: `if user has voucher → check if batch_id in active cache`

This is extremely fast and won't overwhelm MikroTik.

### 4. Deactivation Handling

When you deactivate a batch:
- Set voucher status = "unused" 
- **Disconnect online users via CoA** (already implemented in `BulkDeactivateVouchers`)
- Their quota data **remains in voucher record** - not lost!

When you reactivate a batch:
- Set voucher status = "active"  
- Users can reconnect - **they continue with remaining quota**

### 5. Expiration Handling

Two types of expiration:

| Type | Trigger | Action |
|------|---------|--------|
| Time-based | ExpireTime reached | Reject auth |
| Quota-based | DataUsed >= DataQuota OR TimeUsed >= TimeQuota | Reject auth |

The system should check both on each auth request.

---

## Implementation Plan

### Phase 1: Database Schema Changes
1. Add `batch_id` to `RadiusUser` table (link user to voucher batch)
2. This allows quick lookup: "is this user from a voucher? which batch?"

### Phase 2: Modify Authentication Pipeline
1. In `stageLoadUser` - after getting RadiusUser, check if voucher exists
2. If voucher exists → validate status, expiration, quota
3. Add in-memory cache for active batch IDs
4. Update cache on batch activate/deactivate

### Phase 3: Simplify Batch Operations
1. `BulkActivateVouchers` - just set status="active" (no redemption needed)
2. `BulkDeactivateVouchers` - set status="unused" + disconnect online users
3. Keep CoA disconnect logic (already works)

### Phase 4: Cleanup
1. Remove or deprecate `/redeem` API endpoint
2. Update documentation

---

## Questions:

1. **Where should quota usage be tracked?** 
   - Option A: In `voucher` table only (current)
   - Option B: In `RadiusUser` table only 
   - Option C: Both tables (current) - but need to sync on changes
   
   I recommend **Option C** - keep both, it's more robust as backup.

2. **Should we also track "first_use" expiration type?** 
   - This means: voucher validity starts on first login, not on batch activation
   - Currently supported - should we keep it?

3. **About the CoA disconnect** - should we:
   - Option A: Disconnect immediately (current behavior)
   - Option B: Give them grace period (e.g., 5 minutes) before disconnecting

Please confirm the approach and let me know your preference on the questions above, then I'll create a detailed implementation plan.

---

## User

1. option c, 2. keep it, 3- option b and give them notification why disconnection. 

---

## Assistant (Plan · minimax-m2.5-free · 33.0s)

_Thinking:_

The user has confirmed their preferences:
1. Option C - Keep quota in both voucher and RadiusUser tables
2. Keep first_use expiration type
3. Option B - Give grace period before disconnecting, with notification

Now I need to create a comprehensive implementation plan. Let me analyze the codebase more to understand:

1. The current database schema and how RadiusUser relates to Voucher
2. How the authentication pipeline works
3. How CoA disconnect works
4. How batch activation/deactivation works

Then I'll create a detailed implementation plan.

Let me first check the current RadiusUser schema to understand what fields we need to add.

Perfect! Let me check the current schema and create a comprehensive implementation plan.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/domain/radius.go"
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/domain/radius.go</path>
<type>file</type>
<content>1: package domain
2: 
3: import (
4: 	"time"
5: )
6: 
7: // RADIUS related models
8: 
9: // RadiusProfile RADIUS billing profile
10: type RadiusProfile struct {
11: 	ID             int64     `json:"id,string" form:"id"`                      // Primary key ID
12: 	NodeId         int64     `gorm:"index" json:"node_id,string" form:"node_id"`            // Node ID
13: 
14: 	Name           string    `json:"name" form:"name"`                         // Profile name
15: 	Status         string    `gorm:"index" json:"status" form:"status"`        // Profile status: 0=disabled 1=enabled
16: 	AddrPool       string    `json:"addr_pool" form:"addr_pool"`               // Address pool
17: 	ActiveNum      int       `json:"active_num" form:"active_num"`             // Concurrent sessions
18: 	UpRate         int       `json:"up_rate" form:"up_rate"`                   // Upload rate in Kb
19: 	DownRate       int       `json:"down_rate" form:"down_rate"`               // Download rate in Kb
20: 	DataQuota      int64     `json:"data_quota" form:"data_quota"`             // Data quota in MB (0 = unlimited)
21: 	Domain         string    `json:"domain" form:"domain"`                     // Domain, corresponds to NAS device domain attribute, e.g., Huawei domain_code
22: 	IPv6PrefixPool string    `json:"ipv6_prefix_pool" form:"ipv6_prefix_pool"` // IPv6 prefix pool name for NAS-side allocation
23: 	BindMac        int       `json:"bind_mac" form:"bind_mac"`                 // Bind MAC
24: 	BindVlan       int       `json:"bind_vlan" form:"bind_vlan"`               // Bind VLAN
25: 	Remark         string    `json:"remark" form:"remark"`                     // Remark
26: 	CreatedAt      time.Time `json:"created_at" form:"created_at"`
27: 	UpdatedAt      time.Time `json:"updated_at" form:"updated_at"`
28: }
29: 
30: // TableName Specify table name
31: func (RadiusProfile) TableName() string {
32: 	return "radius_profile"
33: }
34: 
35: // RadiusUser RADIUS Authentication account
36: type RadiusUser struct {
37: 	ID              int64     `json:"id,string" form:"id"`                              // Primary key ID
38: 	NodeId          int64     `json:"node_id,string" form:"node_id"`                    // Node ID
39: 	ProfileId       int64     `gorm:"index" json:"profile_id,string" form:"profile_id"` // RADIUS profile ID
40: 	Realname        string    `gorm:"index" json:"realname" form:"realname"`                         // Contact name
41: 	Email           string    `json:"email" form:"email"`                               // Email address
42: 	Mobile          string    `gorm:"index" json:"mobile" form:"mobile"`                             // Contact phone
43: 
44: 	Address         string    `json:"address" form:"address"`                           // Contact address
45: 	Username        string    `json:"username" gorm:"uniqueIndex" form:"username"`      // Account name
46: 	Password        string    `json:"password" form:"password"`                         // Password
47: 	AddrPool        string    `json:"addr_pool" form:"addr_pool"`                       // Address pool
48: 	ActiveNum       int       `gorm:"index" json:"active_num" form:"active_num"`        // Concurrent sessions
49: 	UpRate          int       `json:"up_rate" form:"up_rate"`                           // Upload rate
50: 	DownRate        int       `json:"down_rate" form:"down_rate"`                       // Download rate
51: 	DataQuota       int64     `json:"data_quota" form:"data_quota"`                     // Data quota in MB (0 = unlimited)
52: 	Vlanid1         int       `json:"vlanid1" form:"vlanid1"`                           // VLAN ID 1
53: 	Vlanid2         int       `json:"vlanid2" form:"vlanid2"`                           // VLAN ID 2
54: 	IpAddr          string    `json:"ip_addr" form:"ip_addr"`                           // Static IP
55: 	IpV6Addr        string    `json:"ipv6_addr" form:"ipv6_addr"`                       // Static IPv6 address
56: 	MacAddr         string    `json:"mac_addr" form:"mac_addr"`                         // MAC address
57: 	Domain          string    `json:"domain" form:"domain"`                             // Domain name for vendor-specific features (e.g., Huawei domain)
58: 	IPv6PrefixPool  string    `json:"ipv6_prefix_pool" form:"ipv6_prefix_pool"`         // IPv6 prefix pool name (inherited from profile or user-specific)
59: 	BindVlan        int       `json:"bind_vlan" form:"bind_vlan"`                       // Bind VLAN
60: 	BindMac         int       `json:"bind_mac" form:"bind_mac"`                         // Bind MAC
61: 	ProfileLinkMode int       `json:"profile_link_mode" form:"profile_link_mode"`       // 0=static (snapshot), 1=dynamic (real-time from profile)
62: 	ExpireTime      time.Time `gorm:"index" json:"expire_time"`                         // Expiration time
63: 	Status          string    `gorm:"index" json:"status" form:"status"`                // Status: enabled | disabled
64: 	Remark          string    `json:"remark" form:"remark"`                             // Remark
65: 
66: 	// Postpaid billing fields.
67: 	// BillingType determines whether the user is billed via prepaid vouchers or postpaid monthly invoices.
68: 	// Default is "prepaid" so existing users are unaffected by the migration.
69: 	BillingType        string    `json:"billing_type" gorm:"size:20;default:'prepaid';index" form:"billing_type"`
70: 	// SubscriptionStatus tracks the postpaid subscription lifecycle.
71: 	// Only relevant when BillingType is "postpaid".
72: 	// Values: "active", "suspended", "canceled". Default empty means N/A for prepaid users.
73: 	SubscriptionStatus string    `json:"subscription_status" gorm:"size:20" form:"subscription_status"`
74: 	// NextBillingDate is the date when the next invoice should be generated.
75: 	// The billing engine cron job checks this field daily.
76: 	NextBillingDate    time.Time `json:"next_billing_date" gorm:"index" form:"next_billing_date"`
77: 	// MonthlyFee is the recurring charge amount for the postpaid subscription.
78: 	// Inherited from the assigned RadiusProfile/Product at subscription time.
79: 	MonthlyFee         float64   `json:"monthly_fee" form:"monthly_fee"`
80: 	// PricePerGb is the cost per Gigabyte of data consumed.
81: 	// Used for consumption-based billing.
82: 	PricePerGb         float64   `json:"price_per_gb" form:"price_per_gb"`
83: 
84: 	OnlineCount     int       `json:"online_count" gorm:"-:migration;<-:false"`
85: 	LastOnline      time.Time `json:"last_online"`
86: 	CreatedAt       time.Time `gorm:"index" json:"created_at"`
87: 	UpdatedAt       time.Time `json:"updated_at"`
88: }
89: 
90: // TableName Specify table name
91: func (RadiusUser) TableName() string {
92: 	return "radius_user"
93: }
94: 
95: // RadiusOnline
96: // Radius RadiusOnline Recode
97: type RadiusOnline struct {
98: 	ID                  int64     `json:"id,string"` // Primary key ID
99: 	Username            string    `gorm:"index" json:"username"`
100: 	NasId               string    `gorm:"index" json:"nas_id"`
101: 	NasAddr             string    `json:"nas_addr"`
102: 	NasPaddr            string    `json:"nas_paddr"`
103: 	SessionTimeout      int       `json:"session_timeout"`
104: 	FramedIpaddr        string    `gorm:"index" json:"framed_ipaddr"`
105: 	FramedNetmask       string    `json:"framed_netmask"`
106: 	FramedIpv6Prefix    string    `json:"framed_ipv6_prefix"`
107: 	FramedIpv6Address   string    `json:"framed_ipv6_address"`
108: 	DelegatedIpv6Prefix string    `json:"delegated_ipv6_prefix"`
109: 	MacAddr             string    `gorm:"index" json:"mac_addr"`
110: 
111: 	NasPort             int64     `json:"nas_port,string"`
112: 	NasClass            string    `json:"nas_class"`
113: 	NasPortId           string    `json:"nas_port_id"`
114: 	NasPortType         int       `json:"nas_port_type"`
115: 	ServiceType         int       `json:"service_type"`
116: 	AcctSessionId       string    `gorm:"index" json:"acct_session_id"`
117: 	AcctSessionTime     int       `json:"acct_session_time"`
118: 	AcctInputTotal      int64     `json:"acct_input_total,string"`
119: 	AcctOutputTotal     int64     `json:"acct_output_total,string"`
120: 	AcctInputPackets    int       `json:"acct_input_packets"`
121: 	AcctOutputPackets   int       `json:"acct_output_packets"`
122: 	AcctStartTime       time.Time `gorm:"index" json:"acct_start_time"`
123: 	LastUpdate          time.Time `json:"last_update"`
124: }
125: 
126: // TableName Specify table name
127: func (RadiusOnline) TableName() string {
128: 	return "radius_online"
129: }
130: 
131: // RadiusAccounting
132: // Radius Accounting Recode
133: type RadiusAccounting struct {
134: 	ID                  int64     `json:"id,string"` // Primary key ID
135: 	Username            string    `gorm:"index" json:"username"`
136: 	AcctSessionId       string    `gorm:"index" json:"acct_session_id"`
137: 	NasId               string    `gorm:"index" json:"nas_id"`
138: 	NasAddr             string    `json:"nas_addr"`
139: 	NasPaddr            string    `json:"nas_paddr"`
140: 	SessionTimeout      int       `json:"session_timeout"`
141: 	FramedIpaddr        string    `gorm:"index" json:"framed_ipaddr"`
142: 	FramedNetmask       string    `json:"framed_netmask"`
143: 	FramedIpv6Prefix    string    `json:"framed_ipv6_prefix"`
144: 	FramedIpv6Address   string    `json:"framed_ipv6_address"`
145: 	DelegatedIpv6Prefix string    `json:"delegated_ipv6_prefix"`
146: 	MacAddr             string    `gorm:"index" json:"mac_addr"`
147: 
148: 	NasPort             int64     `json:"nas_port,string"`
149: 	NasClass            string    `json:"nas_class"`
150: 	NasPortId           string    `json:"nas_port_id"`
151: 	NasPortType         int       `json:"nas_port_type"`
152: 	ServiceType         int       `json:"service_type"`
153: 	AcctSessionTime     int       `json:"acct_session_time"`
154: 	AcctInputTotal      int64     `json:"acct_input_total,string"`
155: 	AcctOutputTotal     int64     `json:"acct_output_total,string"`
156: 	AcctInputPackets    int       `json:"acct_input_packets"`
157: 	AcctOutputPackets   int       `json:"acct_output_packets"`
158: 	LastUpdate          time.Time `json:"last_update"`
159: 	AcctStartTime       time.Time `gorm:"index" json:"acct_start_time"`
160: 	AcctStopTime        time.Time `gorm:"index" json:"acct_stop_time"`
161: }
162: 
163: // TableName Specify table name
164: func (RadiusAccounting) TableName() string {
165: 	return "radius_accounting"
166: }

(End of file - total 166 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 6.4s)

_Thinking:_

Now I understand the current schema. Let me also check how the authentication pipeline works in more detail, and how the disconnect mechanism works.

**Tool: grep**

**Input:**
```json
{
  "include": "*.go",
  "pattern": "DisconnectSession|Disconnect-Request"
}
```

**Output:**
```
Found 24 matches
/home/faris/Downloads/toughradius/toughradius/internal/app/jobs.go:
  Line 355: 			// This would require iterating and calling DisconnectSession

/home/faris/Downloads/toughradius/toughradius/internal/adminapi/vouchers.go:
  Line 741: 			if err := DisconnectSession(c, session); err != nil {

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/coa/client.go:
  Line 16: // Disconnect-Request and CoA-Request messages to NAS devices.
  Line 82: // SendDisconnect sends a Disconnect-Request (RFC 3576) to the NAS to terminate a session.
  Line 84: // This method sends a RADIUS Disconnect-Request packet to the specified NAS.
  Line 342: // buildDisconnectPacket constructs a RADIUS Disconnect-Request packet.
  Line 347: 	// Create packet with Disconnect-Request code

/home/faris/Downloads/toughradius/toughradius/internal/adminapi/coa.go:
  Line 143: // DisconnectSessionEnhanced disconnects a user session using the enhanced CoA client.
  Line 153: func DisconnectSessionEnhanced(c echo.Context) error {
  Line 232: 		zap.L().Info("Disconnect-Request ACK received",
  Line 253: 	zap.L().Warn("Disconnect-Request failed",
  Line 363: 	webserver.ApiPOST("/sessions/:id/disconnect", DisconnectSessionEnhanced)

/home/faris/Downloads/toughradius/toughradius/internal/radiusd/coa/types.go:
  Line 4: // This package provides a robust client for sending CoA-Request and Disconnect-Request
  Line 71: // This request sends a Disconnect-Request (Code 40) to the NAS, which should
  Line 78: //   - RFC 3576 Section 3: Disconnect-Request

/home/faris/Downloads/toughradius/toughradius/internal/app/maintenance/manager.go:
  Line 49: // In a real implementation this would iterate over active sessions and send CoA Disconnect-Requests

/home/faris/Downloads/toughradius/toughradius/internal/adminapi/sessions.go:
  Line 203: 	if err := DisconnectSession(c, session); err != nil {
  Line 212: // DisconnectSession sends a CoA Disconnect-Request to the NAS
  Line 213: func DisconnectSession(c echo.Context, session domain.RadiusOnline) error {
  Line 224: 	// Send CoA Disconnect-Request to NAS asynchronously (non-blocking)
  Line 229: 		// Build CoA Disconnect-Request packet
  Line 242: 			zap.L().Error("Failed to send CoA Disconnect-Request",
  Line 252: 			zap.L().Info("CoA Disconnect-Request ACK received",
  Line 257: 			zap.L().Warn("CoA Disconnect-Request NAK received",
```

---

## Assistant (Plan · minimax-m2.5-free · 4.9s)

_Thinking:_

Now let me check the adminapi/coa.go to understand how disconnect works, particularly the notification mechanism.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/adminapi/coa.go",
  "limit": 200
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/adminapi/coa.go</path>
<type>file</type>
<content>1: package adminapi
2: 
3: import (
4: 	"context"
5: 	"net/http"
6: 	"strconv"
7: 	"time"
8: 
9: 	"github.com/labstack/echo/v4"
10: 	"github.com/talkincode/toughradius/v9/internal/domain"
11: 	"github.com/talkincode/toughradius/v9/internal/radiusd/coa"
12: 	"github.com/talkincode/toughradius/v9/internal/webserver"
13: 	"go.uber.org/zap"
14: )
15: 
16: // CoARequest represents a CoA request from the API.
17: type CoARequest struct {
18: 	// SessionTimeout is the new session timeout in seconds.
19: 	SessionTimeout int `json:"session_timeout"`
20: 	// UpRate is the new upload rate in Kbps.
21: 	UpRate int `json:"up_rate"`
22: 	// DownRate is the new download rate in Kbps.
23: 	DownRate int `json:"down_rate"`
24: 	// DataQuota is the new data quota in MB.
25: 	DataQuota int64 `json:"data_quota"`
26: 	// Reason is the reason for the CoA operation.
27: 	Reason string `json:"reason"`
28: }
29: 
30: // DisconnectRequest represents a disconnect request from the API.
31: type DisconnectRequest struct {
32: 	// Reason is the reason for the disconnect.
33: 	Reason string `json:"reason"`
34: 	// NotifyUser indicates whether to notify the user.
35: 	NotifyUser bool `json:"notify_user"`
36: }
37: 
38: // CoAResponse represents the response from a CoA operation.
39: type CoAResponse struct {
40: 	Success   bool   `json:"success"`
41: 	Message   string `json:"message"`
42: 	NASResp   string `json:"nas_response,omitempty"`
43: 	Duration  int64  `json:"duration_ms,omitempty"`
44: 	RetryCnt  int    `json:"retry_count,omitempty"`
45: 	ErrorCode string `json:"error_code,omitempty"`
46: 	ErrorMsg  string `json:"error_msg,omitempty"`
47: }
48: 
49: // SendCoA sends a CoA-Request to modify an active session.
50: //
51: // @Summary Send CoA-Request to modify session
52: // @Tags CoA
53: // @Accept json
54: // @Produce json
55: // @Param id path int true "Session ID"
56: // @Param request body CoARequest true "CoA Request"
57: // @Success 200 {object} CoAResponse
58: // @Router /api/v1/sessions/{id}/coa [post]
59: func SendCoA(c echo.Context) error {
60: 	// Get session ID
61: 	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
62: 	if err != nil {
63: 		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid session ID", nil)
64: 	}
65: 
66: 	// Parse request body
67: 	var req CoARequest
68: 	if err := c.Bind(&req); err != nil {
69: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
70: 	}
71: 
72: 	// Fetch session from database
73: 	var session domain.RadiusOnline
74: 	if err := GetDB(c).First(&session, id).Error; err != nil {
75: 		return fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found", err)
76: 	}
77: 
78: 	// Fetch NAS info
79: 	var nas domain.NetNas
80: 	if err := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
81: 		return fail(c, http.StatusNotFound, "NAS_NOT_FOUND", "NAS device not found", err)
82: 	}
83: 
84: 	// Determine vendor code
85: 	vendorCode := nas.VendorCode
86: 	if vendorCode == "" {
87: 		vendorCode = coa.VendorGeneric
88: 	}
89: 
90: 	// Create CoA request
91: 	coaReq := coa.CoARequest{
92: 		NASIP:          nas.Ipaddr,
93: 		NASPort:        nas.CoaPort,
94: 		Secret:         nas.Secret,
95: 		Username:       session.Username,
96: 		AcctSessionID:  session.AcctSessionId,
97: 		VendorCode:     vendorCode,
98: 		SessionTimeout: req.SessionTimeout,
99: 		UpRate:         req.UpRate,
100: 		DownRate:       req.DownRate,
101: 		DataQuota:      req.DataQuota,
102: 		Reason:         req.Reason,
103: 	}
104: 
105: 	// Send CoA request asynchronously
106: 	go func() {
107: 		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
108: 		defer cancel()
109: 
110: 		client := coa.NewClient(coa.Config{
111: 			Timeout:    5 * time.Second,
112: 			RetryCount: 2,
113: 		})
114: 
115: 		// Build packet with vendor attributes
116: 		// Note: The client handles basic packet, vendor attributes would need to be added
117: 		// through the vendor builder interface
118: 
119: 		resp := client.SendCoA(ctx, coaReq)
120: 
121: 		if resp.Success {
122: 			zap.L().Info("CoA-Request successful",
123: 				zap.String("nas_addr", nas.Ipaddr),
124: 				zap.String("username", session.Username),
125: 				zap.Int64("session_id", id),
126: 				zap.String("namespace", "adminapi"))
127: 		} else {
128: 			zap.L().Error("CoA-Request failed",
129: 				zap.Error(resp.Error),
130: 				zap.String("nas_addr", nas.Ipaddr),
131: 				zap.String("username", session.Username),
132: 				zap.Int64("session_id", id),
133: 				zap.String("namespace", "adminapi"))
134: 		}
135: 	}()
136: 
137: 	return ok(c, CoAResponse{
138: 		Success: true,
139: 		Message: "CoA request sent successfully",
140: 	})
141: }
142: 
143: // DisconnectSessionEnhanced disconnects a user session using the enhanced CoA client.
144: //
145: // @Summary Disconnect user session (enhanced)
146: // @Tags CoA
147: // @Accept json
148: // @Produce json
149: // @Param id path int true "Session ID"
150: // @Param request body DisconnectRequest false "Disconnect Request"
151: // @Success 200 {object} CoAResponse
152: // @Router /api/v1/sessions/{id}/disconnect [post]
153: func DisconnectSessionEnhanced(c echo.Context) error {
154: 	// Get session ID
155: 	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
156: 	if err != nil {
157: 		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid session ID", nil)
158: 	}
159: 
160: 	// Parse request body (optional)
161: 	var req DisconnectRequest
162: 	if err := c.Bind(&req); err != nil {
163: 		// Use defaults if no body provided
164: 		req = DisconnectRequest{
165: 			Reason:    "admin_initiated",
166: 			NotifyUser: false,
167: 		}
168: 	}
169: 
170: 	// Fetch session from database
171: 	var session domain.RadiusOnline
172: 	if err := GetDB(c).First(&session, id).Error; err != nil {
173: 		return fail(c, http.StatusNotFound, "NOT_FOUND", "Session not found", err)
174: 	}
175: 
176: 	// Delete session from database first (best effort)
177: 	if err := GetDB(c).Delete(&domain.RadiusOnline{}, id).Error; err != nil {
178: 		zap.L().Error("Failed to delete session from database",
179: 			zap.Error(err),
180: 			zap.Int64("session_id", id),
181: 			zap.String("namespace", "adminapi"))
182: 	}
183: 
184: 	// Fetch NAS info
185: 	var nas domain.NetNas
186: 	if err := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
187: 		zap.L().Warn("NAS not found for CoA, session deleted from database only",
188: 			zap.String("nas_addr", session.NasAddr),
189: 			zap.String("username", session.Username),
190: 			zap.String("namespace", "adminapi"))
191: 		return ok(c, CoAResponse{
192: 			Success: true,
193: 			Message: "Session removed from database (NAS not found)",
194: 		})
195: 	}
196: 
197: 	// Determine vendor code
198: 	vendorCode := nas.VendorCode
199: 	if vendorCode == "" {
200: 		vendorCode = coa.VendorGeneric

(Showing lines 1-200 of 365. Use offset=201 to continue.)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 11.4s)

_Thinking:_

Now I have a good understanding of the system. Let me create a comprehensive implementation plan for the user.

Key findings:
1. RadiusUser needs a `voucher_batch_id` field to link to voucher batch
2. Authentication pipeline needs modification to check voucher status
3. Batch activation/deactivation needs modification to handle grace period disconnect
4. Need to handle notification to users during disconnect

Let me create a detailed implementation plan.

**Tool: read**

**Input:**
```json
{
  "filePath": "/home/faris/Downloads/toughradius/toughradius/internal/adminapi/coa.go",
  "limit": 170,
  "offset": 200
}
```

**Output:**
```
<path>/home/faris/Downloads/toughradius/toughradius/internal/adminapi/coa.go</path>
<type>file</type>
<content>200: 		vendorCode = coa.VendorGeneric
201: 	}
202: 
203: 	// Determine CoA port
204: 	coaPort := nas.CoaPort
205: 	if coaPort <= 0 {
206: 		coaPort = 3799 // Default CoA port
207: 	}
208: 
209: 	// Create disconnect request
210: 	discReq := coa.DisconnectRequest{
211: 		NASIP:         nas.Ipaddr,
212: 		NASPort:       coaPort,
213: 		Secret:        nas.Secret,
214: 		Username:       session.Username,
215: 		AcctSessionID: session.AcctSessionId,
216: 		VendorCode:    vendorCode,
217: 		Reason:        req.Reason,
218: 	}
219: 
220: 	// Send disconnect request synchronously with timeout
221: 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
222: 	defer cancel()
223: 
224: 	client := coa.NewClient(coa.Config{
225: 		Timeout:    5 * time.Second,
226: 		RetryCount: 2,
227: 	})
228: 
229: 	resp := client.SendDisconnect(ctx, discReq)
230: 
231: 	if resp.Success {
232: 		zap.L().Info("Disconnect-Request ACK received",
233: 			zap.String("nas_addr", nas.Ipaddr),
234: 			zap.String("username", session.Username),
235: 			zap.String("acct_session_id", session.AcctSessionId),
236: 			zap.String("namespace", "adminapi"))
237: 
238: 		return ok(c, CoAResponse{
239: 			Success:  true,
240: 			Message:  "User disconnected successfully",
241: 			NASResp:  "ACK",
242: 			Duration: resp.Duration.Milliseconds(),
243: 			RetryCnt: resp.RetryCount,
244: 		})
245: 	}
246: 
247: 	// Handle failure
248: 	errorMsg := "Unknown error"
249: 	if resp.Error != nil {
250: 		errorMsg = resp.Error.Error()
251: 	}
252: 
253: 	zap.L().Warn("Disconnect-Request failed",
254: 		zap.String("nas_addr", nas.Ipaddr),
255: 		zap.String("username", session.Username),
256: 		zap.String("error", errorMsg),
257: 		zap.String("namespace", "adminapi"))
258: 
259: 	// Even if NAS fails, session is removed from DB
260: 	return ok(c, CoAResponse{
261: 		Success:   true,
262: 		Message:   "Session removed (NAS disconnect may have failed)",
263: 		NASResp:   "NAK",
264: 		ErrorMsg:  errorMsg,
265: 		Duration:  resp.Duration.Milliseconds(),
266: 		RetryCnt:  resp.RetryCount,
267: 	})
268: }
269: 
270: // BulkDisconnect disconnects multiple user sessions.
271: //
272: // @Summary Bulk disconnect sessions
273: // @Tags CoA
274: // @Accept json
275: // @Produce json
276: // @Param request body struct{SessionIDs []int64 `json:"session_ids"`} true "Session IDs"
277: // @Success 200 {object} SuccessResponse
278: // @Router /api/v1/sessions/bulk-disconnect [post]
279: func BulkDisconnect(c echo.Context) error {
280: 	type BulkRequest struct {
281: 		SessionIDs []int64 `json:"session_ids"`
282: 	}
283: 
284: 	var req BulkRequest
285: 	if err := c.Bind(&req); err != nil {
286: 		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
287: 	}
288: 
289: 	if len(req.SessionIDs) == 0 {
290: 		return fail(c, http.StatusBadRequest, "NO_SESSIONS", "No session IDs provided", nil)
291: 	}
292: 
293: 	// Process each session
294: 	successCount := 0
295: 	failCount := 0
296: 
297: 	for _, sessionID := range req.SessionIDs {
298: 		// Fetch session
299: 		var session domain.RadiusOnline
300: 		if err := GetDB(c).First(&session, sessionID).Error; err != nil {
301: 			failCount++
302: 			continue
303: 		}
304: 
305: 		// Fetch NAS
306: 		var nas domain.NetNas
307: 		if err := GetDB(c).Where("ipaddr = ?", session.NasAddr).First(&nas).Error; err != nil {
308: 			// Delete from DB anyway
309: 			GetDB(c).Delete(&domain.RadiusOnline{}, sessionID)
310: 			successCount++
311: 			continue
312: 		}
313: 
314: 		// Determine CoA port
315: 		coaPort := nas.CoaPort
316: 		if coaPort <= 0 {
317: 			coaPort = 3799
318: 		}
319: 
320: 		// Create disconnect request
321: 		discReq := coa.DisconnectRequest{
322: 			NASIP:         nas.Ipaddr,
323: 			NASPort:       coaPort,
324: 			Secret:        nas.Secret,
325: 			Username:      session.Username,
326: 			AcctSessionID: session.AcctSessionId,
327: 		}
328: 
329: 		// Send disconnect (fire and forget for bulk)
330: 		go func() {
331: 			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
332: 			defer cancel()
333: 
334: 			client := coa.NewClient(coa.Config{
335: 				Timeout:    3 * time.Second,
336: 				RetryCount: 1,
337: 			})
338: 
339: 			resp := client.SendDisconnect(ctx, discReq)
340: 			if resp.Success {
341: 				zap.L().Info("Bulk disconnect successful",
342: 					zap.String("username", session.Username),
343: 					zap.String("namespace", "coa"))
344: 			}
345: 		}()
346: 
347: 		// Delete from database
348: 		GetDB(c).Delete(&domain.RadiusOnline{}, sessionID)
349: 		successCount++
350: 	}
351: 
352: 	return ok(c, map[string]interface{}{
353: 		"message":       "Bulk disconnect completed",
354: 		"total":        len(req.SessionIDs),
355: 		"success":      successCount,
356: 		"failed":       failCount,
357: 	})
358: }
359: 
360: // registerCoARoutes registers CoA-related routes.
361: func registerCoARoutes() {
362: 	webserver.ApiPOST("/sessions/:id/coa", SendCoA)
363: 	webserver.ApiPOST("/sessions/:id/disconnect", DisconnectSessionEnhanced)
364: 	webserver.ApiPOST("/sessions/bulk-disconnect", BulkDisconnect)
365: }

(End of file - total 365 lines)
</content>
```

---

## Assistant (Plan · minimax-m2.5-free · 44.5s)

_Thinking:_

Now I have a complete understanding of the system. Let me create a comprehensive implementation plan.

Now I have a complete picture. Let me create a comprehensive implementation plan:

---

# Implementation Plan: Voucher Batch Control with Direct Authentication

## Overview

This plan implements direct voucher authentication without redemption, with full batch-level control for activation/deactivation, quota tracking, and graceful disconnect with notifications.

---

## Phase 1: Database Schema Changes

### 1.1 Add Voucher Batch Link to RadiusUser

**File:** `internal/domain/radius.go`

```go
// Add to RadiusUser struct:
VoucherBatchID    int64     `json:"voucher_batch_id,string" gorm:"index" form:"voucher_batch_id"` // Links to voucher batch (0 = not from voucher)
VoucherCode       string    `json:"voucher_code" gorm:"index" form:"voucher_code"`                // Original voucher code
```

### 1.2 Database Migration

Create migration to add new columns:
- `voucher_batch_id` (int64, indexed)
- `voucher_code` (string, indexed)

---

## Phase 2: Voucher Authentication Plugin

### 2.1 Create VoucherAuthChecker Plugin

**New File:** `internal/radiusd/plugins/auth/checkers/voucher_auth.go`

This plugin will:
1. Check if user has `VoucherBatchID > 0`
2. If yes, validate voucher status:
   - Must be "active" status
   - Must not be expired (time-based)
   - Must not exceed quota (data/time)
3. Return appropriate error if validation fails

**Implementation:**
```go
type VoucherAuthChecker struct {
    voucherRepo repository.VoucherRepository
    // Use in-memory cache for active batch IDs
    activeBatchCache *cache.Cache
}

func (c *VoucherAuthChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error {
    user := authCtx.User
    if user.VoucherBatchID == 0 {
        return nil // Not a voucher user, skip
    }
    
    // Check if batch is active (fast cache check)
    if !c.activeBatchCache.Exists(user.VoucherBatchID) {
        return errors.New("voucher batch is not active")
    }
    
    // Get voucher for detailed validation
    voucher, err := c.voucherRepo.GetByCode(ctx, user.VoucherCode)
    if err != nil {
        return errors.New("voucher not found")
    }
    
    // Check status
    if voucher.Status != "active" {
        return errors.New("voucher is not active")
    }
    
    // Check time expiration
    if !voucher.ExpireTime.IsZero() && voucher.ExpireTime.Before(time.Now()) {
        return errors.New("voucher has expired")
    }
    
    // Check data quota
    if voucher.DataQuota > 0 && voucher.DataUsed >= voucher.DataQuota {
        return errors.New("voucher data quota exceeded")
    }
    
    // Check time quota
    if voucher.TimeQuota > 0 && voucher.TimeUsed >= voucher.TimeQuota {
        return errors.New("voucher time quota exceeded")
    }
    
    return nil
}
```

### 2.2 In-Memory Active Batch Cache

**New File:** `internal/radiusd/cache/voucher_batch_cache.go`

Simple TTL cache for active batch IDs:
- On batch activation → Add batch ID to cache
- On batch deactivation → Remove batch ID from cache
- Cache TTL: 1 minute (fast expiration if cache miss)

---

## Phase 3: Batch Activation/Deactivation Changes

### 3.1 BulkActivateVouchers Modification

**File:** `internal/adminapi/vouchers.go`

Change logic:
```go
func BulkActivateVouchers(c echo.Context) error {
    // ... existing code to find batch ...
    
    // Update vouchers to active
    updates := map[string]interface{}{
        "status": "active",
        "activated_at": now,
    }
    
    // For "first_use" type, calculate expiry on first login (keep current behavior)
    
    // Add batch ID to active cache
    voucherCache.AddBatch(batch.ID)
    
    // If "first_use" expiration: create RadiusUser records NOW (not on first login)
    // This ensures users can connect immediately after batch activation
    if batch.ExpirationType == "first_use" {
        // Create RadiusUser for each activated voucher
        go createRadiusUsersForBatch(batch.ID)
    }
    
    // Return response
}
```

### 3.2 BulkDeactivateVouchers with Grace Period

**File:** `internal/adminapi/vouchers.go`

```go
func BulkDeactivateVouchers(c echo.Context) error {
    // Parse grace period from request (default 5 minutes)
    gracePeriod := c.QueryParam("grace_period") // in seconds
    graceDuration := 5 * time.Minute
    if gp, err := strconv.Atoi(gracePeriod); err == nil {
        graceDuration = time.Duration(gp) * time.Second
    }
    
    reason := c.QueryParam("reason") // e.g., "agent_payment_pending"
    
    // 1. Update voucher status to "unused"
    updates := map[string]interface{}{
        "status": "unused",
    }
    db.Model(&domain.Voucher{}).Where("batch_id = ? AND status = ?", batchID, "active").Updates(updates)
    
    // 2. Remove from active cache
    voucherCache.RemoveBatch(batch.ID)
    
    // 3. Find all online users from this batch
    var sessions []domain.RadiusOnline
    db.Joins("JOIN radius_user ON radius_user.username = radius_online.username").
        Where("radius_user.voucher_batch_id = ?", batchID).
        Find(&sessions)
    
    // 4. Schedule grace period disconnect
    if graceDuration > 0 {
        // Send notification to users (via CoA or other mechanism)
        for _, session := range sessions {
            // Send Session-Timeout update with short timeout as warning
            sendGracePeriodWarning(session, graceDuration, reason)
        }
        
        // Schedule actual disconnect
        go func() {
            time.Sleep(graceDuration)
            disconnectBatchUsers(batch.ID, reason)
        }()
    } else {
        // Immediate disconnect
        disconnectBatchUsers(batch.ID, reason)
    }
}

func disconnectBatchUsers(batchID int64, reason string) {
    // Query all online sessions for this batch
    // Send CoA Disconnect-Request to each NAS
    // Delete sessions from database
    // Log disconnect with reason
}
```

---

## Phase 4: User Notification Mechanism

### 4.1 Grace Period Warning

**Implementation options:**

**Option A: Session-Timeout CoA** (Recommended)
- Send CoA-Request with short Session-Timeout (grace period)
- MikroTik will show notification if configured
- After timeout, user gets disconnected

**Option B: Vendor-Specific Attribute**
- Use Mikrotik's `Mikrotilk-Wireless-Notify` attribute
- Requires NAS configuration

**Option C: Custom Attribute**
- Send Disconnect-Request with `Reply-Message` attribute containing reason
- MikroTik can display this message

We'll implement **Option A** with fallback to **Option C**.

---

## Phase 5: Quota Sync on Accounting

### 5.1 Accounting Update Plugin

**File:** `internal/radiusd/plugins/auth/checkers/voucher_quota_updater.go`

On each Accounting-Request:
1. Check if user has `VoucherBatchID > 0`
2. If yes, update both:
   - `radius_user.data_used` / `radius_user.time_used`
   - `voucher.data_used` / `voucher.time_used`
3. This ensures quota is tracked in both tables

---

## Phase 6: First_Use Expiration Handling

### 6.1 FirstUseActivator Enhancement

**File:** `internal/radiusd/plugins/auth/checkers/first_use_activator.go`

Current behavior: Sets expiration on first login
New behavior:
1. On first login, calculate expiry from batch validity
2. Update voucher `ExpireTime` 
3. Update RadiusUser `ExpireTime`
4. Continue with normal auth

---

## Phase 7: API Cleanup (Optional)

### 7.1 Deprecate Redemption API

Mark `/api/v1/vouchers/redeem` as deprecated:
- Add deprecation warning to response
- Still works for backward compatibility

---

## Implementation Order

| Phase | Task | Complexity |
|-------|------|------------|
| 1 | Database schema + migration | Medium |
| 2.1 | VoucherAuthChecker plugin | High |
| 2.2 | Active batch cache | Low |
| 3.1 | Modify BulkActivateVouchers | Medium |
| 3.2 | Modify BulkDeactivateVouchers | High |
| 4 | Grace period + notification | High |
| 5 | Accounting quota sync | Medium |
| 6 | First-use activator enhancement | Low |
| 7 | Deprecate redemption API | Low |

---

## Estimated Timeline

- **Phase 1-2**: 1-2 days
- **Phase 3-4**: 2-3 days
- **Phase 5-7**: 1 day

**Total**: ~4-6 days for full implementation

---

## Testing Plan

1. **Unit Tests**: Test VoucherAuthChecker logic
2. **Integration Tests**: 
   - Activate batch → Users can connect
   - Deactivate batch → Users disconnected after grace period
   - Quota exceeded → Users rejected
   - Reactivate → Users can reconnect with remaining quota

---

Do you want me to proceed with implementation? Please let me know if you'd like to adjust any aspect of this plan.

---

## User

please proceed

---

## Assistant (Build · minimax-m2.5-free)

_Thinking:_

The user

---

