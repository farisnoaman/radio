# Provider Admin Reporting & Alerts System - Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a comprehensive reporting dashboard and alerts system for Provider Admins with real-time network status, financial reporting, device issues, fraud detection, and configurable notifications.

**Architecture:** Multi-phase implementation with:
- Phase 1: Database migrations (new tables)
- Phase 2: Backend domain models and services
- Phase 3: API endpoints
- Phase 4: Frontend dashboard and widgets
- Phase 5: Scheduled jobs for aggregation and fraud detection

**Tech Stack:** Go (backend), React Admin, Material-UI, PostgreSQL, Cron jobs

---

## File Structure Overview

### Backend (New Files)

| File | Purpose |
|------|---------|
| `internal/domain/reporting.go` | Domain models for snapshots, fraud, notifications |
| `internal/migration/reporting_tables.go` | Database migration for all new tables |
| `internal/service/reporting_service.go` | Core reporting logic |
| `internal/service/aggregation_service.go` | Daily/weekly/monthly snapshot aggregation |
| `internal/service/fraud_detection_service.go` | Fraud detection engine |
| `internal/service/provider_notification_service.go` | Alert generation |
| `internal/adminapi/reporting.go` | API handlers |
| `internal/adminapi/reporting_routes.go` | Route registration |

### Backend (Modify)

| File | Change |
|------|--------|
| `internal/app/jobs.go` | Add scheduled jobs |
| `config/config.go` | Add notification config |
| `internal/adminapi/adminapi.go` | Register routes |

### Frontend (New Files)

| File | Purpose |
|------|---------|
| `web/src/pages/Platform/ReportingDashboard.tsx` | Main dashboard page |
| `web/src/pages/Platform/NotificationSettings.tsx` | Alert settings |
| `web/src/components/Platform/SummaryCard.tsx` | Reusable metric card |
| `web/src/components/Platform/NetworkStatusWidget.tsx` | Network status |
| `web/src/components/Platform/AgentFinancialWidget.tsx` | Agent revenue |
| `web/src/components/Platform/IssuesReporterWidget.tsx` | Issues list |
| `web/src/components/Platform/FraudAlertWidget.tsx` | Fraud alerts |
| `web/src/components/ExportButton.tsx` | CSV export |

### Frontend (Modify)

| File | Change |
|------|--------|
| `web/src/App.tsx` | Add routes |
| `web/src/pages/Platform/PlatformSettings.tsx` | Add Notifications tab |
| `web/src/i18n/en-US.ts` | Translations |
| `web/src/i18n/ar.ts` | Translations |
| `web/src/i18n/zh-CN.ts` | Translations |

---

## Phase 1: Database Migrations

### Task 1: Create Reporting Domain Models

**Files:**
- Create: `internal/domain/reporting.go`

```go
package domain

import (
    "time"
)

// DailySnapshot represents aggregated metrics for a provider
type DailySnapshot struct {
    ID                     int64     `json:"id" gorm:"primaryKey"`
    ProviderID             int64     `json:"provider_id" gorm:"index"`
    SnapshotDate           time.Time `json:"snapshot_date" gorm:"type:date"`
    TotalUsers            int       `json:"total_users"`
    ActiveUsers           int       `json:"active_users"`
    NewMonthlyUsers       int       `json:"new_monthly_users"`
    NewVoucherUsers       int       `json:"new_voucher_users"`
    TotalSessions         int       `json:"total_sessions"`
    ActiveSessions        int       `json:"active_sessions"`
    MonthlyDataUsedBytes  int64     `json:"monthly_data_used_bytes"`
    VoucherDataUsedBytes  int64     `json:"voucher_data_used_bytes"`
    ActiveNodes           int       `json:"active_nodes"`
    TotalNodes            int       `json:"total_nodes"`
    ActiveServers         int       `json:"active_servers"`
    TotalServers          int       `json:"total_servers"`
    ActiveCPEs            int       `json:"active_cpes"`
    TotalCPEs             int       `json:"total_cpes"`
    TotalAgents           int       `json:"total_agents"`
    TotalBatches          int       `json:"total_batches"`
    AgentRevenue          float64   `json:"agent_revenue"`
    MRR                   float64   `json:"mrr"`
    DeviceIssues          int       `json:"device_issues"`
    NetworkIssues          int       `json:"network_issues"`
    FraudAttempts         int       `json:"fraud_attempts"`
    CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (DailySnapshot) TableName() string {
    return "reporting_daily_snapshots"
}

// FraudLog records suspicious voucher activation attempts
type FraudLog struct {
    ID         int64     `json:"id" gorm:"primaryKey"`
    ProviderID int64     `json:"provider_id" gorm:"index"`
    VoucherID  int64     `json:"voucher_id"`
    UserID     int64     `json:"user_id"`
    IPAddress  string    `json:"ip_address" gorm:"type:varchar(45)"`
    EventType  string    `json:"event_type" gorm:"type:varchar(50)"`
    Details    string    `json:"details" gorm:"type:jsonb"`
    CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (FraudLog) TableName() string {
    return "reporting_fraud_log"
}

// ProviderNotificationPreference stores alert settings
type ProviderNotificationPreference struct {
    ID                           int64     `json:"id" gorm:"primaryKey"`
    ProviderID                   int64     `json:"provider_id" gorm:"uniqueIndex"`
    AlertPercentages             string    `json:"alert_percentages" gorm:"type:varchar(50);default:'70,85,100'"`
    AlertPercentagesEnabled      bool      `json:"alert_percentages_enabled" gorm:"default:true"`
    MaxUsersThreshold            int       `json:"max_users_threshold"`
    MaxDataBytesThreshold        int64     `json:"max_data_bytes_threshold"`
    AbsoluteAlertsEnabled        bool      `json:"absolute_alerts_enabled" gorm:"default:false"`
    AnomalyDetectionEnabled      bool      `json:"anomaly_detection_enabled" gorm:"default:false"`
    AnomalyThresholdPercent      int       `json:"anomaly_threshold_percent" gorm:"default:50"`
    EmailEnabled                 bool      `json:"email_enabled" gorm:"default:true"`
    SMSEnabled                   bool      `json:"sms_enabled" gorm:"default:false"`
    CreatedAt                    time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt                    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ProviderNotificationPreference) TableName() string {
    return "provider_notification_preferences"
}

// NetworkIssue records device/network problems
type NetworkIssue struct {
    ID            int64      `json:"id" gorm:"primaryKey"`
    ProviderID    int64      `json:"provider_id" gorm:"index"`
    DeviceType   string     `json:"device_type" gorm:"type:varchar(20)"` // node, server, nas, cpe, user_device
    DeviceID     int64      `json:"device_id"`
    DeviceName   string     `json:"device_name" gorm:"type:varchar(255)"`
    IssueType    string     `json:"issue_type" gorm:"type:varchar(50)"`
    IssueDetails string     `json:"issue_details" gorm:"type:text"`
    Status       string     `json:"status" gorm:"type:varchar(20);default:'open'"` // open, resolved, ignored
    ResolvedAt   *time.Time `json:"resolved_at"`
    CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (NetworkIssue) TableName() string {
    return "reporting_network_issues"
}
```

**Steps:**
- [ ] **Step 1:** Create file `internal/domain/reporting.go`
- [ ] **Step 2:** Run test to verify it compiles: `go build ./internal/domain/...`
- [ ] **Step 3:** Commit

```bash
git add internal/domain/reporting.go
git commit -m "feat(domain): add reporting domain models

- DailySnapshot for aggregated metrics
- FraudLog for suspicious activity tracking
- ProviderNotificationPreference for alert settings
- NetworkIssue for device problems

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 2: Create Database Migration

**Files:**
- Create: `internal/migration/reporting_tables.go`

```go
package migrations

import (
    "github.com/go-gormigrate/gormigrate/v2"
    "gorm.io/gorm"
)

func reportingTables() *gormigrate.Migration {
    return &gormigrate.Migration{
        ID: "20260322000001_reporting_tables",
        Migrate: func(tx *gorm.DB) error {
            // Daily snapshots table
            if err := tx.Exec(`
                CREATE TABLE reporting_daily_snapshots (
                    id BIGSERIAL PRIMARY KEY,
                    provider_id BIGINT NOT NULL,
                    snapshot_date DATE NOT NULL,
                    total_users INT DEFAULT 0,
                    active_users INT DEFAULT 0,
                    new_monthly_users INT DEFAULT 0,
                    new_voucher_users INT DEFAULT 0,
                    total_sessions INT DEFAULT 0,
                    active_sessions INT DEFAULT 0,
                    monthly_data_used_bytes BIGINT DEFAULT 0,
                    voucher_data_used_bytes BIGINT DEFAULT 0,
                    active_nodes INT DEFAULT 0,
                    total_nodes INT DEFAULT 0,
                    active_servers INT DEFAULT 0,
                    total_servers INT DEFAULT 0,
                    active_cpes INT DEFAULT 0,
                    total_cpes INT DEFAULT 0,
                    total_agents INT DEFAULT 0,
                    total_batches INT DEFAULT 0,
                    agent_revenue DECIMAL(15,2) DEFAULT 0,
                    mrr DECIMAL(15,2) DEFAULT 0,
                    device_issues INT DEFAULT 0,
                    network_issues INT DEFAULT 0,
                    fraud_attempts INT DEFAULT 0,
                    created_at TIMESTAMP DEFAULT NOW(),
                    UNIQUE(provider_id, snapshot_date)
                );
                CREATE INDEX idx_daily_snapshots_provider_date ON reporting_daily_snapshots(provider_id, snapshot_date);
            `).Error; err != nil {
                return err
            }

            // Fraud log table
            if err := tx.Exec(`
                CREATE TABLE reporting_fraud_log (
                    id BIGSERIAL PRIMARY KEY,
                    provider_id BIGINT NOT NULL,
                    voucher_id BIGINT,
                    user_id BIGINT,
                    ip_address VARCHAR(45),
                    event_type VARCHAR(50) NOT NULL,
                    details JSONB,
                    created_at TIMESTAMP DEFAULT NOW()
                );
                CREATE INDEX idx_fraud_log_provider ON reporting_fraud_log(provider_id);
                CREATE INDEX idx_fraud_log_ip_time ON reporting_fraud_log(ip_address, created_at);
            `).Error; err != nil {
                return err
            }

            // Provider notification preferences
            if err := tx.Exec(`
                CREATE TABLE provider_notification_preferences (
                    id BIGSERIAL PRIMARY KEY,
                    provider_id BIGINT NOT NULL UNIQUE,
                    alert_percentages VARCHAR(50) DEFAULT '70,85,100',
                    alert_percentages_enabled BOOLEAN DEFAULT TRUE,
                    max_users_threshold INT,
                    max_data_bytes_threshold BIGINT,
                    absolute_alerts_enabled BOOLEAN DEFAULT FALSE,
                    anomaly_detection_enabled BOOLEAN DEFAULT FALSE,
                    anomaly_threshold_percent INT DEFAULT 50,
                    email_enabled BOOLEAN DEFAULT TRUE,
                    sms_enabled BOOLEAN DEFAULT FALSE,
                    created_at TIMESTAMP DEFAULT NOW(),
                    updated_at TIMESTAMP DEFAULT NOW()
                );
            `).Error; err != nil {
                return err
            }

            // Network issues table
            if err := tx.Exec(`
                CREATE TABLE reporting_network_issues (
                    id BIGSERIAL PRIMARY KEY,
                    provider_id BIGINT NOT NULL,
                    device_type VARCHAR(20) NOT NULL,
                    device_id BIGINT,
                    device_name VARCHAR(255),
                    issue_type VARCHAR(50) NOT NULL,
                    issue_details TEXT,
                    status VARCHAR(20) DEFAULT 'open',
                    resolved_at TIMESTAMP,
                    created_at TIMESTAMP DEFAULT NOW()
                );
                CREATE INDEX idx_network_issues_provider ON reporting_network_issues(provider_id);
                CREATE INDEX idx_network_issues_status ON reporting_network_issues(status, created_at);
            `).Error; err != nil {
                return err
            }

            return nil
        },
        Rollback: func(tx *gorm.DB) error {
            return tx.Exec(`
                DROP TABLE IF EXISTS reporting_daily_snapshots;
                DROP TABLE IF EXISTS reporting_fraud_log;
                DROP TABLE IF EXISTS provider_notification_preferences;
                DROP TABLE IF EXISTS reporting_network_issues;
            `).Error
        },
    }
}
```

**Steps:**
- [ ] **Step 1:** Register migration in main migration file
- [ ] **Step 2:** Run migration test
- [ ] **Step 3:** Commit

---

## Phase 2: Backend Services

### Task 3: Create Reporting Service

**Files:**
- Create: `internal/service/reporting_service.go`

```go
package service

import (
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "gorm.io/gorm"
)

type ReportingService struct {
    db *gorm.DB
}

func NewReportingService(db *gorm.DB) *ReportingService {
    return &ReportingService{db: db}
}

type ReportingSummary struct {
    Period     string `json:"period"` // daily, weekly, monthly
    StartDate  string `json:"start_date"`
    EndDate    string `json:"end_date"`
    Users      UserMetrics
    Sessions   SessionMetrics
    Data       DataMetrics
    Network    NetworkMetrics
    Agents     AgentMetrics
    Issues     IssueMetrics
}

type UserMetrics struct {
    TotalUsers      int `json:"total_users"`
    ActiveUsers     int `json:"active_users"`
    NewMonthlyUsers int `json:"new_monthly_users"`
    NewVoucherUsers int `json:"new_voucher_users"`
}

type SessionMetrics struct {
    TotalSessions  int `json:"total_sessions"`
    ActiveSessions int `json:"active_sessions"`
}

type DataMetrics struct {
    MonthlyDataUsedGB float64 `json:"monthly_data_used_gb"`
    VoucherDataUsedGB float64 `json:"voucher_data_used_gb"`
}

type NetworkMetrics struct {
    Nodes       DeviceCount `json:"nodes"`
    Servers     DeviceCount `json:"servers"`
    CPEs        DeviceCount `json:"cpes"`
}

type DeviceCount struct {
    Active int `json:"active"`
    Total  int `json:"total"`
}

type AgentMetrics struct {
    TotalAgents   int     `json:"total_agents"`
    TotalBatches int     `json:"total_batches"`
    Revenue       float64 `json:"revenue"`
    MRR           float64 `json:"mrr"`
}

type IssueMetrics struct {
    DeviceIssues int `json:"device_issues"`
    NetworkIssues int `json:"network_issues"`
}

// GetSummary returns aggregated metrics for the specified period
func (s *ReportingService) GetSummary(providerID int64, period string, startDate, endDate time.Time) (*ReportingSummary, error) {
    summary := &ReportingSummary{
        Period:    period,
        StartDate: startDate.Format("2006-01-02"),
        EndDate:   endDate.Format("2006-01-02"),
    }

    // Query daily snapshots for the date range
    var snapshots []domain.DailySnapshot
    err := s.db.Where("provider_id = ? AND snapshot_date BETWEEN ? AND ?",
        providerID, startDate, endDate).Find(&snapshots).Error
    if err != nil {
        return nil, err
    }

    // Aggregate from snapshots
    for _, snap := range snapshots {
        summary.Users.TotalUsers += snap.TotalUsers
        summary.Users.ActiveUsers += snap.ActiveUsers
        summary.Users.NewMonthlyUsers += snap.NewMonthlyUsers
        summary.Users.NewVoucherUsers += snap.NewVoucherUsers
        summary.Sessions.TotalSessions += snap.TotalSessions
        summary.Sessions.ActiveSessions += snap.ActiveSessions
        summary.Data.MonthlyDataUsedGB += float64(snap.MonthlyDataUsedBytes) / (1024 * 1024 * 1024)
        summary.Data.VoucherDataUsedGB += float64(snap.VoucherDataUsedBytes) / (1024 * 1024 * 1024)
        summary.Network.Nodes.Active += snap.ActiveNodes
        summary.Network.Nodes.Total += snap.TotalNodes
        summary.Network.Servers.Active += snap.ActiveServers
        summary.Network.Servers.Total += snap.TotalServers
        summary.Network.CPEs.Active += snap.ActiveCPEs
        summary.Network.CPEs.Total += snap.TotalCPEs
        summary.Agents.TotalAgents = max(summary.Agents.TotalAgents, snap.TotalAgents)
        summary.Agents.TotalBatches += snap.TotalBatches
        summary.Agents.Revenue += snap.AgentRevenue
        summary.Agents.MRR += snap.MRR
        summary.Issues.DeviceIssues += snap.DeviceIssues
        summary.Issues.NetworkIssues += snap.NetworkIssues
    }

    // If no snapshots, calculate real-time
    if len(snapshots) == 0 {
        return s.getRealTimeSummary(providerID)
    }

    return summary, nil
}

// GetRealTimeSummary calculates metrics directly from source tables
func (s *ReportingService) getRealTimeSummary(providerID int64) (*ReportingSummary, error) {
    summary := &ReportingSummary{
        Period:    "realtime",
        StartDate: time.Now().Format("2006-01-02"),
        EndDate:   time.Now().Format("2006-01-02"),
    }

    now := time.Now()
    monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

    // Users
    var userCount, activeUserCount int64
    s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ?", providerID).Count(&userCount)
    s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeUserCount)
    summary.Users.TotalUsers = int(userCount)
    summary.Users.ActiveUsers = int(activeUserCount)

    // Active sessions
    var sessionCount int64
    s.db.Model(&domain.RadiusAccounting{}).Where("tenant_id = ? AND acctstoptime IS NULL", providerID).Count(&sessionCount)
    summary.Sessions.ActiveSessions = int(sessionCount)

    // Network status (real-time)
    s.getNetworkStatus(providerID, summary)

    return summary, nil
}

func (s *ReportingService) getNetworkStatus(providerID int64, summary *ReportingSummary) error {
    // Nodes
    var totalNodes, activeNodes int64
    s.db.Model(&domain.NetNode{}).Where("tenant_id = ?", providerID).Count(&totalNodes)
    s.db.Model(&domain.NetNode{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeNodes)
    summary.Network.Nodes.Total = int(totalNodes)
    summary.Network.Nodes.Active = int(activeNodes)

    // Servers
    var totalServers, activeServers int64
    s.db.Model(&domain.Server{}).Where("tenant_id = ?", providerID).Count(&totalServers)
    s.db.Model(&domain.Server{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeServers)
    summary.Network.Servers.Total = int(totalServers)
    summary.Network.Servers.Active = int(activeServers)

    // CPEs
    var totalCPEs, activeCPEs int64
    s.db.Model(&domain.CPEDevice{}).Where("tenant_id = ?", providerID).Count(&totalCPEs)
    s.db.Model(&domain.CPEDevice{}).Where("tenant_id = ? AND status = ?", providerID, "active").Count(&activeCPEs)
    summary.Network.CPEs.Total = int(totalCPEs)
    summary.Network.CPEs.Active = int(activeCPEs)

    return nil
}
```

**Steps:**
- [ ] **Step 1:** Create file `internal/service/reporting_service.go`
- [ ] **Step 2:** Build to verify: `go build ./internal/service/...`
- [ ] **Step 3:** Commit

---

### Task 4: Create Fraud Detection Service

**Files:**
- Create: `internal/service/fraud_detection_service.go`

```go
package service

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "gorm.io/gorm"
)

type FraudDetectionService struct {
    db *gorm.DB
}

func NewFraudDetectionService(db *gorm.DB) *FraudDetectionService {
    return &FraudDetectionService{db: db}
}

type FraudRule struct {
    Name      string
    Threshold int
    Window    time.Duration
    Action    string
}

var fraudRules = []FraudRule{
    {Name: "ip_activation_limit", Threshold: 5, Window: time.Hour, Action: "flag"},
    {Name: "same_voucher_multi_use", Threshold: 2, Window: time.Hour, Action: "quarantine"},
    {Name: "rapid_successive", Threshold: 10, Window: time.Minute, Action: "block"},
}

// CheckAndRecord checks if an activation is suspicious
func (s *FraudDetectionService) CheckAndRecord(providerID int64, voucherID int64, userID int64, ipAddress string) ([]string, error) {
    var triggeredRules []string

    for _, rule := range fraudRules {
        count, err := s.countRecentEvents(providerID, ipAddress, voucherID, rule)
        if err != nil {
            return nil, err
        }

        if count >= rule.Threshold {
            triggeredRules = append(triggeredRules, rule.Name)

            // Log the fraud event
            details, _ := json.Marshal(map[string]interface{}{
                "rule":          rule.Name,
                "count":         count,
                "threshold":     rule.Threshold,
                "voucher_id":    voucherID,
                "user_id":       userID,
            })

            fraudLog := &domain.FraudLog{
                ProviderID: providerID,
                VoucherID:  voucherID,
                UserID:     userID,
                IPAddress:  ipAddress,
                EventType:  rule.Name,
                Details:    string(details),
            }
            s.db.Create(fraudLog)

            // Execute action
            switch rule.Action {
            case "quarantine":
                s.quarantineVoucher(voucherID)
            case "block":
                s.blockIP(providerID, ipAddress)
            }
        }
    }

    return triggeredRules, nil
}

func (s *FraudDetectionService) countRecentEvents(providerID int64, ipAddress string, voucherID int64, rule FraudRule) (int64, error) {
    since := time.Now().Add(-rule.Window)

    var count int64
    switch rule.Name {
    case "ip_activation_limit":
        s.db.Model(&domain.FraudLog{}).
            Where("provider_id = ? AND ip_address = ? AND created_at > ?", providerID, ipAddress, since).
            Count(&count)
    case "same_voucher_multi_use":
        s.db.Model(&domain.FraudLog{}).
            Where("provider_id = ? AND voucher_id = ? AND event_type = ? AND created_at > ?",
                providerID, voucherID, "same_voucher_multi_use", since).
            Count(&count)
    case "rapid_successive":
        s.db.Model(&domain.FraudLog{}).
            Where("provider_id = ? AND ip_address = ? AND created_at > ?", providerID, ipAddress, since).
            Count(&count)
    }

    return count, nil
}

func (s *FraudDetectionService) quarantineVoucher(voucherID int64) {
    s.db.Model(&domain.Voucher{}).Where("id = ?", voucherID).Update("status", "quarantined")
}

func (s *FraudDetectionService) blockIP(providerID int64, ipAddress string) {
    // Add to blocklist (implementation depends on existing patterns)
    fmt.Printf("Blocking IP %s for provider %d\n", ipAddress, providerID)
}

// GetFraudLogs returns recent fraud events for a provider
func (s *FraudDetectionService) GetFraudLogs(providerID int64, limit int) ([]domain.FraudLog, error) {
    var logs []domain.FraudLog
    err := s.db.Where("provider_id = ?", providerID).
        Order("created_at DESC").
        Limit(limit).
        Find(&logs).Error
    return logs, err
}
```

**Steps:**
- [ ] **Step 1:** Create file `internal/service/fraud_detection_service.go`
- [ ] **Step 2:** Build to verify
- [ ] **Step 3:** Commit

---

### Task 5: Create Provider Notification Service

**Files:**
- Create: `internal/service/provider_notification_service.go`

```go
package service

import (
    "fmt"
    "strconv"
    "strings"
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "gorm.io/gorm"
)

type ProviderNotificationService struct {
    db            *gorm.DB
    emailProvider EmailProvider
}

func NewProviderNotificationService(db *gorm.DB, emailProvider EmailProvider) *ProviderNotificationService {
    return &ProviderNotificationService{db: db, emailProvider: emailProvider}
}

type NotificationCheck struct {
    CurrentUsagePercent float64
    TotalUsers         int
    ActiveUsers        int
    DataUsedGB         float64
}

// CheckThresholds checks all notification rules for a provider
func (s *ProviderNotificationService) CheckThresholds(providerID int64, check NotificationCheck) error {
    pref, err := s.getOrCreatePreferences(providerID)
    if err != nil {
        return err
    }

    if pref.AlertPercentagesEnabled {
        if err := s.checkPercentageThresholds(providerID, pref, check.CurrentUsagePercent); err != nil {
            return err
        }
    }

    if pref.AbsoluteAlertsEnabled {
        if err := s.checkAbsoluteThresholds(providerID, pref, check); err != nil {
            return err
        }
    }

    return nil
}

func (s *ProviderNotificationService) checkPercentageThresholds(providerID int64, pref *domain.ProviderNotificationPreference, usagePercent float64) error {
    thresholds := parseThresholds(pref.AlertPercentages)

    for _, threshold := range thresholds {
        if usagePercent >= float64(threshold) {
            // Check if we already sent this alert today
            if !s.shouldSendAlert(providerID, "percentage", threshold) {
                continue
            }

            subject := fmt.Sprintf("Usage Alert: %d%% of Plan Limit", threshold)
            body := fmt.Sprintf("Your provider has used %.1f%% of its plan limit. Current usage: %.2f GB", usagePercent, usagePercent)

            if s.emailProvider != nil && pref.EmailEnabled {
                s.emailProvider.SendEmail("provider@example.com", subject, body)
            }

            s.recordAlertSent(providerID, "percentage", threshold)
        }
    }

    return nil
}

func (s *ProviderNotificationService) checkAbsoluteThresholds(providerID int64, pref *domain.ProviderNotificationPreference, check NotificationCheck) error {
    if pref.MaxUsersThreshold > 0 && check.TotalUsers >= pref.MaxUsersThreshold {
        if s.shouldSendAlert(providerID, "max_users", pref.MaxUsersThreshold) {
            subject := fmt.Sprintf("User Limit Alert: %d Users", check.TotalUsers)
            if s.emailProvider != nil && pref.EmailEnabled {
                s.emailProvider.SendEmail("provider@example.com", subject, "You have reached your maximum user limit.")
            }
            s.recordAlertSent(providerID, "max_users", pref.MaxUsersThreshold)
        }
    }

    return nil
}

func (s *ProviderNotificationService) shouldSendAlert(providerID int64, alertType string, threshold int) bool {
    var count int64
    yesterday := time.Now().Add(-24 * time.Hour)
    s.db.Model(&domain.UsageAlert{}).
        Where("user_id = ? AND alert_type = ? AND threshold = ? AND sent_at > ?",
            providerID, alertType, threshold, yesterday).
        Count(&count)
    return count == 0
}

func (s *ProviderNotificationService) recordAlertSent(providerID int64, alertType string, threshold int) {
    alert := &domain.UsageAlert{
        UserID:    providerID,
        Threshold: threshold,
        AlertType: alertType,
    }
    now := time.Now()
    alert.SentAt = &now
    s.db.Create(alert)
}

func (s *ProviderNotificationService) getOrCreatePreferences(providerID int64) (*domain.ProviderNotificationPreference, error) {
    var pref domain.ProviderNotificationPreference
    err := s.db.Where("provider_id = ?", providerID).First(&pref).Error

    if err == gorm.ErrRecordNotFound {
        pref = domain.ProviderNotificationPreference{
            ProviderID:              providerID,
            AlertPercentages:        "70,85,100",
            AlertPercentagesEnabled: true,
        }
        err = s.db.Create(&pref).Error
    }

    return &pref, err
}

func parseThresholds(s string) []int {
    if s == "" {
        return []int{}
    }
    var result []int
    for _, part := range strings.Split(s, ",") {
        val, err := strconv.Atoi(strings.TrimSpace(part))
        if err == nil && val > 0 {
            result = append(result, val)
        }
    }
    return result
}
```

**Steps:**
- [ ] **Step 1:** Create file `internal/service/provider_notification_service.go`
- [ ] **Step 2:** Build to verify
- [ ] **Step 3:** Commit

---

## Phase 3: API Endpoints

### Task 6: Create Reporting API Handlers

**Files:**
- Create: `internal/adminapi/reporting.go`
- Create: `internal/adminapi/reporting_routes.go`

```go
package adminapi

import (
    "encoding/csv"
    "fmt"
    "net/http"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/service"
)

type ReportingHandler struct {
    reportingService         *service.ReportingService
    fraudService             *service.FraudDetectionService
    notificationService      *service.ProviderNotificationService
}

func NewReportingHandler(db *gorm.DB) *ReportingHandler {
    return &ReportingHandler{
        reportingService:    service.NewReportingService(db),
        fraudService:        service.NewFraudDetectionService(db),
        notificationService: service.NewProviderNotificationService(db, nil),
    }
}

func registerReportingRoutes() {
    handler := NewReportingHandler(GetDB(nil))

    webserver.ApiGET("/reporting/summary", handler.GetSummary)
    webserver.ApiGET("/reporting/users", handler.GetUserMetrics)
    webserver.ApiGET("/reporting/sessions", handler.GetSessionMetrics)
    webserver.ApiGET("/reporting/network-status", handler.GetNetworkStatus)
    webserver.ApiGET("/reporting/agents", handler.GetAgentMetrics)
    webserver.ApiGET("/reporting/issues", handler.GetIssues)
    webserver.ApiGET("/reporting/fraud", handler.GetFraudLogs)
    webserver.ApiGET("/reporting/export", handler.ExportCSV)
    webserver.ApiGET("/reporting/notifications/preferences", handler.GetNotificationPreferences)
    webserver.ApiPUT("/reporting/notifications/preferences", handler.UpdateNotificationPreferences)
}

func (h *ReportingHandler) GetSummary(c echo.Context) error {
    providerID := GetOperatorTenantID(c)

    period := c.QueryParam("period") // daily, weekly, monthly
    if period == "" {
        period = "daily"
    }

    startDate := parseDate(c.QueryParam("start_date"), time.Now().AddDate(0, -1, 0))
    endDate := parseDate(c.QueryParam("end_date"), time.Now())

    summary, err := h.reportingService.GetSummary(providerID, period, startDate, endDate)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "REPORTING_ERROR", err.Error(), nil)
    }

    return ok(c, summary)
}

func (h *ReportingHandler) GetNetworkStatus(c echo.Context) error {
    providerID := GetOperatorTenantID(c)

    summary, err := h.reportingService.GetSummary(providerID, "realtime", time.Now(), time.Now())
    if err != nil {
        return fail(c, http.StatusInternalServerError, "REPORTING_ERROR", err.Error(), nil)
    }

    return ok(c, summary.Network)
}

func (h *ReportingHandler) GetFraudLogs(c echo.Context) error {
    providerID := GetOperatorTenantID(c)

    limit := parseInt(c.QueryParam("limit"), 50)

    logs, err := h.fraudService.GetFraudLogs(providerID, limit)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "FRAUD_ERROR", err.Error(), nil)
    }

    return ok(c, logs)
}

func (h *ReportingHandler) GetNotificationPreferences(c echo.Context) error {
    providerID := GetOperatorTenantID(c)

    pref, err := h.notificationService.GetPreferences(providerID)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "PREFERENCES_ERROR", err.Error(), nil)
    }

    return ok(c, pref)
}

func (h *ReportingHandler) UpdateNotificationPreferences(c echo.Context) error {
    providerID := GetOperatorTenantID(c)

    var req struct {
        AlertPercentages            string `json:"alert_percentages"`
        AlertPercentagesEnabled     bool   `json:"alert_percentages_enabled"`
        MaxUsersThreshold           int    `json:"max_users_threshold"`
        AbsoluteAlertsEnabled      bool   `json:"absolute_alerts_enabled"`
        AnomalyDetectionEnabled     bool   `json:"anomaly_detection_enabled"`
        AnomalyThresholdPercent     int    `json:"anomaly_threshold_percent"`
        EmailEnabled               bool   `json:"email_enabled"`
        SMSEnabled                 bool   `json:"sms_enabled"`
    }

    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
    }

    err := h.notificationService.UpdatePreferences(providerID, &req)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "UPDATE_ERROR", err.Error(), nil)
    }

    return ok(c, map[string]string{"message": "Preferences updated"})
}

func (h *ReportingHandler) ExportCSV(c echo.Context) error {
    providerID := GetOperatorTenantID(c)

    reportType := c.QueryParam("type") // users, sessions, network, agents
    startDate := parseDate(c.QueryParam("start_date"), time.Now().AddDate(0, -1, 0))
    endDate := parseDate(c.QueryParam("end_date"), time.Now())

    summary, err := h.reportingService.GetSummary(providerID, "daily", startDate, endDate)
    if err != nil {
        return fail(c, http.StatusInternalServerError, "EXPORT_ERROR", err.Error(), nil)
    }

    c.Response().Header().Set("Content-Type", "text/csv")
    c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s.csv", time.Now().Format("20060102")))

    writer := csv.NewWriter(c.Response().Writer)
    defer writer.Flush()

    // Write CSV based on report type
    writer.Write([]string{"Metric", "Value"})
    writer.Write([]string{"Total Users", fmt.Sprintf("%d", summary.Users.TotalUsers)})
    writer.Write([]string{"Active Users", fmt.Sprintf("%d", summary.Users.ActiveUsers)})
    writer.Write([]string{"New Monthly Users", fmt.Sprintf("%d", summary.Users.NewMonthlyUsers)})
    writer.Write([]string{"New Voucher Users", fmt.Sprintf("%d", summary.Users.NewVoucherUsers)})
    writer.Write([]string{"Active Sessions", fmt.Sprintf("%d", summary.Sessions.ActiveSessions)})
    writer.Write([]string{"Monthly Data (GB)", fmt.Sprintf("%.2f", summary.Data.MonthlyDataUsedGB)})
    writer.Write([]string{"Voucher Data (GB)", fmt.Sprintf("%.2f", summary.Data.VoucherDataUsedGB)})

    return nil
}

func parseDate(s string, defaultVal time.Time) time.Time {
    if s == "" {
        return defaultVal
    }
    t, err := time.Parse("2006-01-02", s)
    if err != nil {
        return defaultVal
    }
    return t
}

func parseInt(s string, defaultVal int) int {
    if s == "" {
        return defaultVal
    }
    val, err := strconv.Atoi(s)
    if err != nil {
        return defaultVal
    }
    return val
}
```

**Steps:**
- [ ] **Step 1:** Create file `internal/adminapi/reporting.go`
- [ ] **Step 2:** Create file `internal/adminapi/reporting_routes.go`
- [ ] **Step 3:** Register routes in `adminapi.go`
- [ ] **Step 4:** Build to verify
- [ ] **Step 5:** Commit

---

## Phase 4: Frontend Dashboard

### Task 7: Create Summary Card Component

**Files:**
- Create: `web/src/components/Platform/SummaryCard.tsx`

```typescript
import { Card, CardContent, Typography, Box, Skeleton } from '@mui/material';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import TrendingDownIcon from '@mui/icons-material/TrendingDown';

interface SummaryCardProps {
  title: string;
  value: string | number;
  icon?: React.ReactNode;
  trend?: {
    value: number;
    direction: 'up' | 'down';
  };
  loading?: boolean;
  color?: string;
}

export const SummaryCard = ({ title, value, icon, trend, loading, color = '#1976d2' }: SummaryCardProps) => {
  if (loading) {
    return (
      <Card sx={{ height: '100%' }}>
        <CardContent>
          <Skeleton variant="text" width="60%" />
          <Skeleton variant="text" width="40%" height={40} />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card sx={{ height: '100%', borderLeft: `4px solid ${color}` }}>
      <CardContent>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          {title}
        </Typography>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h4" sx={{ fontWeight: 'bold' }}>
            {value}
          </Typography>
          {icon}
        </Box>
        {trend && (
          <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
            {trend.direction === 'up' ? (
              <TrendingUpIcon sx={{ color: 'success.main', fontSize: 20 }} />
            ) : (
              <TrendingDownIcon sx={{ color: 'error.main', fontSize: 20 }} />
            )}
            <Typography
              variant="caption"
              sx={{ color: trend.direction === 'up' ? 'success.main' : 'error.main', ml: 0.5 }}
            >
              {trend.value}% {trend.direction === 'up' ? 'increase' : 'decrease'}
            </Typography>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};
```

**Steps:**
- [ ] **Step 1:** Create component
- [ ] **Step 2:** Export from index
- [ ] **Step 3:** Commit

---

### Task 8: Create Network Status Widget

**Files:**
- Create: `web/src/components/Platform/NetworkStatusWidget.tsx`

```typescript
import { Grid, Card, CardContent, Typography, Box, Chip } from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';
import LanIcon from '@mui/icons-material/Lan';
import DnsIcon from '@mui/icons-material/Dns';
import RouterIcon from '@mui/icons-material/Router';

interface NetworkMetrics {
  nodes: { active: number; total: number };
  servers: { active: number; total: number };
  cpes: { active: number; total: number };
}

const DeviceStatus = ({ active, total, label, icon }: { active: number; total: number; label: string; icon: React.ReactNode }) => {
  const percent = total > 0 ? (active / total) * 100 : 0;
  const color = percent >= 80 ? 'success' : percent >= 50 ? 'warning' : 'error';

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          {icon}
          <Typography variant="h6" sx={{ ml: 1 }}>
            {label}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h3" sx={{ fontWeight: 'bold' }}>
            {active}/{total}
          </Typography>
          <Chip label={`${percent.toFixed(0)}%`} color={color} size="small" />
        </Box>
        <Typography variant="caption" color="text.secondary">
          {active} active of {total} total
        </Typography>
      </CardContent>
    </Card>
  );
};

export const NetworkStatusWidget = () => {
  const { data, isLoading } = useApiQuery<NetworkMetrics>({
    path: '/api/v1/reporting/network-status',
    queryKey: ['reporting', 'network-status'],
    enabled: true,
  });

  return (
    <Box>
      <Typography variant="h6" sx={{ mb: 2 }}>
        Network Status (Real-time)
      </Typography>
      <Grid container spacing={2}>
        <Grid item xs={12} md={4}>
          <DeviceStatus
            active={data?.nodes.active ?? 0}
            total={data?.nodes.total ?? 0}
            label="Nodes"
            icon={<LanIcon color="primary" />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <DeviceStatus
            active={data?.servers.active ?? 0}
            total={data?.servers.total ?? 0}
            label="Servers"
            icon={<DnsIcon color="primary" />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <DeviceStatus
            active={data?.cpes.active ?? 0}
            total={data?.cpes.total ?? 0}
            label="CPEs"
            icon={<RouterIcon color="primary" />}
          />
        </Grid>
      </Grid>
    </Box>
  );
};
```

**Steps:**
- [ ] **Step 1:** Create component
- [ ] **Step 2:** Export from index
- [ ] **Step 3:** Commit

---

### Task 9: Create Reporting Dashboard Page

**Files:**
- Create: `web/src/pages/Platform/ReportingDashboard.tsx`

```typescript
import { useState } from 'react';
import { Box, Typography, Grid, Card, CardContent, Tabs, Tab, Button } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import { SummaryCard } from '../components/Platform/SummaryCard';
import { NetworkStatusWidget } from '../components/Platform/NetworkStatusWidget';
import { AgentFinancialWidget } from '../components/Platform/AgentFinancialWidget';
import { IssuesReporterWidget } from '../components/Platform/IssuesReporterWidget';
import { FraudAlertWidget } from '../components/Platform/FraudAlertWidget';
import { useApiQuery } from '../hooks/useApiQuery';

export default function ReportingDashboard() {
  const [period, setPeriod] = useState('daily');

  const { data: summary, isLoading } = useApiQuery({
    path: `/api/v1/reporting/summary?period=${period}`,
    queryKey: ['reporting', 'summary', period],
    enabled: true,
  });

  const handleExport = () => {
    window.open(`/api/v1/reporting/export?type=summary&period=${period}`, '_blank');
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 'bold' }}>
          Reporting Dashboard
        </Typography>
        <Box sx={{ display: 'flex', gap: 2 }}>
          <Tabs value={period} onChange={(_, v) => setPeriod(v)}>
            <Tab label="Daily" value="daily" />
            <Tab label="Weekly" value="weekly" />
            <Tab label="Monthly" value="monthly" />
          </Tabs>
          <Button startIcon={<DownloadIcon />} variant="outlined" onClick={handleExport}>
            Export CSV
          </Button>
        </Box>
      </Box>

      {/* Key Metrics */}
      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <SummaryCard title="Total Users" value={summary?.users?.total_users ?? 0} loading={isLoading} color="#1976d2" />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <SummaryCard title="New Monthly Users" value={summary?.users?.new_monthly_users ?? 0} loading={isLoading} color="#2e7d32" />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <SummaryCard title="Active Sessions" value={summary?.sessions?.active_sessions ?? 0} loading={isLoading} color="#ed6c02" />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <SummaryCard title="Monthly Data (GB)" value={`${(summary?.data?.monthly_data_used_gb ?? 0).toFixed(1)}`} loading={isLoading} color="#9c27b0" />
        </Grid>
      </Grid>

      {/* Network Status (Real-time) */}
      <Box sx={{ mb: 3 }}>
        <NetworkStatusWidget />
      </Box>

      {/* Agent Financial */}
      <Box sx={{ mb: 3 }}>
        <AgentFinancialWidget />
      </Box>

      {/* Issues & Fraud */}
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <IssuesReporterWidget />
        </Grid>
        <Grid item xs={12} md={6}>
          <FraudAlertWidget />
        </Grid>
      </Grid>
    </Box>
  );
}
```

**Steps:**
- [ ] **Step 1:** Create dashboard page
- [ ] **Step 2:** Add route in App.tsx
- [ ] **Step 3:** Add menu item in CustomMenu
- [ ] **Step 4:** Add translations
- [ ] **Step 5:** Commit

---

### Task 10: Create Remaining Widget Components

**Files:**
- Create: `web/src/components/Platform/AgentFinancialWidget.tsx`
- Create: `web/src/components/Platform/IssuesReporterWidget.tsx`
- Create: `web/src/components/Platform/FraudAlertWidget.tsx`

```typescript
// AgentFinancialWidget.tsx
import { Card, CardContent, Typography, Box, Chip } from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';

interface AgentMetrics {
  total_agents: number;
  total_batches: number;
  revenue: number;
  mrr: number;
}

export const AgentFinancialWidget = () => {
  const { data } = useApiQuery<AgentMetrics>({
    path: '/api/v1/reporting/agents',
    queryKey: ['reporting', 'agents'],
    enabled: true,
  });

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>Agent Financial Summary</Typography>
        <Box sx={{ display: 'flex', gap: 3, mb: 2 }}>
          <Box>
            <Typography variant="caption" color="text.secondary">Total Agents</Typography>
            <Typography variant="h4">{data?.total_agents ?? 0}</Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary">Revenue</Typography>
            <Typography variant="h4">${(data?.revenue ?? 0).toFixed(2)}</Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// IssuesReporterWidget.tsx
import { Card, CardContent, Typography, Box, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip } from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';

interface Issue {
  id: number;
  device_type: string;
  device_name: string;
  issue_type: string;
  status: string;
  created_at: string;
}

export const IssuesReporterWidget = () => {
  const { data: issues } = useApiQuery<Issue[]>({
    path: '/api/v1/reporting/issues?status=open',
    queryKey: ['reporting', 'issues'],
    enabled: true,
  });

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>Device & Network Issues</Typography>
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Device</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Issue</TableCell>
                <TableCell>Status</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {issues?.slice(0, 10).map((issue) => (
                <TableRow key={issue.id}>
                  <TableCell>{issue.device_name}</TableCell>
                  <TableCell>{issue.device_type}</TableCell>
                  <TableCell>{issue.issue_type}</TableCell>
                  <TableCell>
                    <Chip label={issue.status} color={issue.status === 'open' ? 'error' : 'default'} size="small" />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </CardContent>
    </Card>
  );
};

// FraudAlertWidget.tsx
import { Card, CardContent, Typography, Box, Alert } from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';

interface FraudEvent {
  id: number;
  ip_address: string;
  event_type: string;
  details: string;
  created_at: string;
}

export const FraudAlertWidget = () => {
  const { data: fraudLogs } = useApiQuery<FraudEvent[]>({
    path: '/api/v1/reporting/fraud?limit=10',
    queryKey: ['reporting', 'fraud'],
    enabled: true,
  });

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>Fraud Detection</Typography>
        {fraudLogs && fraudLogs.length > 0 ? (
          <Box>
            {fraudLogs.map((log) => (
              <Alert severity="warning" key={log.id} sx={{ mb: 1 }}>
                <Typography variant="body2">
                  <strong>{log.event_type}</strong> - IP: {log.ip_address}
                </Typography>
                <Typography variant="caption">{new Date(log.created_at).toLocaleString()}</Typography>
              </Alert>
            ))}
          </Box>
        ) : (
          <Typography variant="body2" color="text.secondary">No fraud events detected</Typography>
        )}
      </CardContent>
    </Card>
  );
};
```

**Steps:**
- [ ] **Step 1:** Create AgentFinancialWidget
- [ ] **Step 2:** Create IssuesReporterWidget
- [ ] **Step 3:** Create FraudAlertWidget
- [ ] **Step 4:** Commit

---

### Task 11: Create Notification Settings Page

**Files:**
- Create: `web/src/pages/Platform/NotificationSettings.tsx`

```typescript
import { useState, useEffect } from 'react';
import { Box, Card, CardContent, Typography, Switch, FormControlLabel, TextField, Button, Stack, Divider } from '@mui/material';
import { useNotify, useTranslate } from 'react-admin';
import { useApiQuery, useApiMutation } from '../hooks/useApiQuery';

export default function NotificationSettings() {
  const translate = useTranslate();
  const notify = useNotify();

  const [preferences, setPreferences] = useState({
    alert_percentages: '70,85,100',
    alert_percentages_enabled: true,
    max_users_threshold: 0,
    absolute_alerts_enabled: false,
    anomaly_detection_enabled: false,
    anomaly_threshold_percent: 50,
    email_enabled: true,
    sms_enabled: false,
  });

  const { data, isLoading } = useApiQuery({
    path: '/api/v1/reporting/notifications/preferences',
    queryKey: ['reporting', 'notification-preferences'],
    enabled: true,
  });

  useEffect(() => {
    if (data) {
      setPreferences(data);
    }
  }, [data]);

  const mutation = useApiMutation(
    'PUT',
    '/api/v1/reporting/notifications/preferences',
    {
      onSuccess: () => {
        notify(translate('common.save_success'), { type: 'success' });
      },
      onError: (error: Error) => {
        notify(error.message, { type: 'error' });
      },
    }
  );

  const handleSave = () => {
    mutation.mutate(preferences);
  };

  return (
    <Box sx={{ p: 3, maxWidth: 800 }}>
      <Typography variant="h5" sx={{ fontWeight: 'bold', mb: 3 }}>
        Notification Settings
      </Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Percentage-Based Alerts
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={preferences.alert_percentages_enabled}
                onChange={(e) => setPreferences({ ...preferences, alert_percentages_enabled: e.target.checked })}
              />
            }
            label="Enable percentage alerts"
          />
          <TextField
            label="Alert Percentages (comma-separated)"
            value={preferences.alert_percentages}
            onChange={(e) => setPreferences({ ...preferences, alert_percentages: e.target.value })}
            disabled={!preferences.alert_percentages_enabled}
            fullWidth
            sx={{ mt: 2 }}
            placeholder="70,85,100"
          />
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Absolute Thresholds
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={preferences.absolute_alerts_enabled}
                onChange={(e) => setPreferences({ ...preferences, absolute_alerts_enabled: e.target.checked })}
              />
            }
            label="Enable absolute threshold alerts"
          />
          <TextField
            label="Max Users Threshold"
            type="number"
            value={preferences.max_users_threshold}
            onChange={(e) => setPreferences({ ...preferences, max_users_threshold: parseInt(e.target.value) })}
            disabled={!preferences.absolute_alerts_enabled}
            fullWidth
            sx={{ mt: 2 }}
          />
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Anomaly Detection
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={preferences.anomaly_detection_enabled}
                onChange={(e) => setPreferences({ ...preferences, anomaly_detection_enabled: e.target.checked })}
              />
            }
            label="Enable anomaly detection"
          />
          <TextField
            label="Deviation Threshold (%)"
            type="number"
            value={preferences.anomaly_threshold_percent}
            onChange={(e) => setPreferences({ ...preferences, anomaly_threshold_percent: parseInt(e.target.value) })}
            disabled={!preferences.anomaly_detection_enabled}
            fullWidth
            sx={{ mt: 2 }}
          />
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Notification Channels
          </Typography>
          <Stack spacing={2}>
            <FormControlLabel
              control={
                <Switch
                  checked={preferences.email_enabled}
                  onChange={(e) => setPreferences({ ...preferences, email_enabled: e.target.checked })}
                />
              }
              label="Email notifications"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={preferences.sms_enabled}
                  onChange={(e) => setPreferences({ ...preferences, sms_enabled: e.target.checked })}
                />
              }
              label="SMS notifications"
            />
          </Stack>
        </CardContent>
      </Card>

      <Button variant="contained" size="large" onClick={handleSave}>
        Save Settings
      </Button>
    </Box>
  );
}
```

**Steps:**
- [ ] **Step 1:** Create notification settings page
- [ ] **Step 2:** Add as tab in PlatformSettings.tsx
- [ ] **Step 3:** Add translations
- [ ] **Step 4:** Commit

---

## Phase 5: Scheduled Jobs

### Task 11: Add Scheduled Jobs

**Files:**
- Modify: `internal/app/jobs.go`

```go
// Add to existing job registration in initJobs()

// Daily snapshot job - runs at 1 AM
_, err = a.sched.AddFunc("0 1 * * *", func() {
    go a.SchedDailySnapshotJob()
})
if err != nil {
    zap.S().Errorf("init daily snapshot job error: %s", err.Error())
}

// Fraud analysis job - runs every 15 minutes
_, err = a.sched.AddFunc("*/15 * * * *", func() {
    go a.SchedFraudAnalysisJob()
})
if err != nil {
    zap.S().Errorf("init fraud analysis job error: %s", err.Error())
}

// SchedDailySnapshotJob creates daily snapshots for all providers
func (a *Application) SchedDailySnapshotJob() {
    defer func() {
        if err := recover(); err != nil {
            zap.S().Error(err)
        }
    }()

    zap.S().Info("Starting daily snapshot job")

    var providers []domain.Provider
    if err := a.gormDB.Find(&providers).Error; err != nil {
        zap.S().Errorf("Failed to fetch providers: %v", err)
        return
    }

    for _, provider := range providers {
        if err := a.createDailySnapshot(&provider); err != nil {
            zap.S().Errorf("Failed to create snapshot for provider %d: %v", provider.ID, err)
        }
    }

    zap.S().Info("Daily snapshot job completed")
}

func (a *Application) createDailySnapshot(provider *domain.Provider) error {
    today := time.Now().Truncate(24 * time.Hour)

    // Check if snapshot already exists
    var existing domain.DailySnapshot
    if err := a.gormDB.Where("provider_id = ? AND snapshot_date = ?", provider.ID, today).First(&existing).Error; err == nil {
        return nil // Already exists
    }

    snapshot := &domain.DailySnapshot{
        ProviderID:     provider.ID,
        SnapshotDate:   today,
        TotalUsers:    a.countUsers(provider.ID),
        ActiveUsers:   a.countActiveUsers(provider.ID),
        NewMonthlyUsers: a.countNewMonthlyUsers(provider.ID, today),
        NewVoucherUsers: a.countNewVoucherUsers(provider.ID, today),
        ActiveSessions: a.countActiveSessions(provider.ID),
        MonthlyDataUsedBytes: a.sumMonthlyDataUsed(provider.ID),
        VoucherDataUsedBytes: a.sumVoucherDataUsed(provider.ID),
        ActiveNodes: a.countActiveNodes(provider.ID),
        TotalNodes: a.countTotalNodes(provider.ID),
        ActiveServers: a.countActiveServers(provider.ID),
        TotalServers: a.countTotalServers(provider.ID),
        TotalAgents: a.countAgents(provider.ID),
        TotalBatches: a.countBatches(provider.ID),
    }

    return a.gormDB.Create(snapshot).Error
}

// SchedFraudAnalysisJob checks recent activations for fraud
func (a *Application) SchedFraudAnalysisJob() {
    defer func() {
        if err := recover(); err != nil {
            zap.S().Error(err)
        }
    }()

    fraudService := service.NewFraudDetectionService(a.gormDB)

    // Get recent voucher activations (last 15 minutes)
    var activations []struct {
        VoucherID int64
        UserID    int64
        IPAddress string
    }

    if err := a.gormDB.Table("vouchers").
        Select("id as voucher_id, user_id, ip_address").
        Where("activated_at > ?", time.Now().Add(-15*time.Minute)).
        Find(&activations).Error; err != nil {
        zap.S().Errorf("Failed to fetch activations: %v", err)
        return
    }

    for _, act := range activations {
        if _, err := fraudService.CheckAndRecord(act.ProviderID, act.VoucherID, act.UserID, act.IPAddress); err != nil {
            zap.S().Errorf("Fraud check failed: %v", err)
        }
    }
}

// SchedWeeklyRollupJob aggregates daily -> weekly snapshots
func (a *Application) SchedWeeklyRollupJob() {
    defer func() {
        if err := recover(); err != nil {
            zap.S().Error(err)
        }
    }()

    zap.S().Info("Starting weekly rollup job")

    weekStart := getWeekStart(time.Now())

    var providers []domain.Provider
    if err := a.gormDB.Find(&providers).Error; err != nil {
        return
    }

    for _, provider := range providers {
        a.createWeeklySnapshot(&provider, weekStart)
    }
}

// SchedMonthlyRollupJob aggregates weekly -> monthly snapshots
func (a *Application) SchedMonthlyRollupJob() {
    defer func() {
        if err := recover(); err != nil {
            zap.S().Error(err)
        }
    }()

    zap.S().Info("Starting monthly rollup job")

    monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

    var providers []domain.Provider
    if err := a.gormDB.Find(&providers).Error; err != nil {
        return
    }

    for _, provider := range providers {
        a.createMonthlySnapshot(&provider, monthStart)
    }
}

// SchedDataRetentionJob purges old data beyond retention limits
func (a *Application) SchedDataRetentionJob() {
    defer func() {
        if err := recover(); err != nil {
            zap.S().Error(err)
        }
    }()

    zap.S().Info("Starting data retention job")

    ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
    a.gormDB.Where("snapshot_date < ?", ninetyDaysAgo).Delete(&domain.DailySnapshot{})

    zap.S().Info("Data retention job completed")
}

func getWeekStart(t time.Time) time.Time {
    weekday := int(t.Weekday())
    if weekday == 0 {
        weekday = 7
    }
    return time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 0, 0, 0, 0, t.Location())
}
```

**Steps:**
- [ ] **Step 1:** Add job registrations for all 5 jobs
- [ ] **Step 2:** Implement snapshot creation
- [ ] **Step 3:** Implement fraud analysis
- [ ] **Step 4:** Implement weekly/monthly rollup jobs
- [ ] **Step 5:** Implement data retention job
- [ ] **Step 6:** Build and test
- [ ] **Step 7:** Commit

---

### Task 13: Create Aggregation Service

**Files:**
- Create: `internal/service/aggregation_service.go`

```go
package service

import (
    "time"

    "github.com/talkincode/toughradius/v9/internal/domain"
    "gorm.io/gorm"
)

type AggregationService struct {
    db *gorm.DB
}

func NewAggregationService(db *gorm.DB) *AggregationService {
    return &AggregationService{db: db}
}

// CreateDailySnapshot creates or updates a daily snapshot for a provider
func (s *AggregationService) CreateDailySnapshot(providerID int64, date time.Time) error {
    today := date.Truncate(24 * time.Hour)

    // Check if snapshot already exists
    var existing domain.DailySnapshot
    if err := s.db.Where("provider_id = ? AND snapshot_date = ?", providerID, today).First(&existing).Error; err == nil {
        return nil // Already exists, skip
    }

    snapshot := &domain.DailySnapshot{
        ProviderID:   providerID,
        SnapshotDate: today,
    }

    // Aggregate from daily data
    s.aggregateUserMetrics(snapshot, providerID, today)
    s.aggregateSessionMetrics(snapshot, providerID, today)
    s.aggregateDataMetrics(snapshot, providerID, today)
    s.aggregateNetworkMetrics(snapshot, providerID)
    s.aggregateAgentMetrics(snapshot, providerID)

    return s.db.Create(snapshot).Error
}

func (s *AggregationService) aggregateUserMetrics(snapshot *domain.DailySnapshot, providerID int64, date time.Time) {
    monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())

    // Total users
    s.db.Model(&domain.RadiusUser{}).Where("tenant_id = ?", providerID).Count((*int64)(&snapshot.TotalUsers))

    // Active users (had session in last 30 days)
    thirtyDaysAgo := date.AddDate(0, 0, -30)
    s.db.Model(&domain.RadiusUser{}).
        Where("tenant_id = ? AND last_login >= ?", providerID, thirtyDaysAgo).
        Count((*int64)(&snapshot.ActiveUsers))

    // New monthly users
    s.db.Model(&domain.RadiusUser{}).
        Where("tenant_id = ? AND created_at >= ?", providerID, monthStart).
        Count((*int64)(&snapshot.NewMonthlyUsers))
}

func (s *AggregationService) aggregateSessionMetrics(snapshot *domain.DailySnapshot, providerID int64, date time.Time) {
    // Active sessions
    s.db.Model(&domain.RadiusAccounting{}).
        Where("tenant_id = ? AND acctstoptime IS NULL", providerID).
        Count((*int64)(&snapshot.ActiveSessions))
}

func (s *AggregationService) aggregateDataMetrics(snapshot *domain.DailySnapshot, providerID int64, date time.Time) {
    monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())

    // Monthly data used
    s.db.Model(&domain.RadiusAccounting{}).
        Select("COALESCE(SUM(acctinputoctets + acctoutputoctets), 0)").
        Where("tenant_id = ? AND acctstarttime >= ?", providerID, monthStart).
        Scan(&snapshot.MonthlyDataUsedBytes)
}
```

**Steps:**
- [ ] **Step 1:** Create aggregation service
- [ ] **Step 2:** Build to verify
- [ ] **Step 3:** Commit

---

## Phase 6: Translations

### Task 14: Add i18n Translations

**Files:**
- Modify: `web/src/i18n/en-US.ts`
- Modify: `web/src/i18n/ar.ts`
- Modify: `web/src/i18n/zh-CN.ts`

Add to `menu:` section:
```typescript
reporting: 'Reporting Dashboard',
notification_settings: 'Notification Settings',
```

Add to `platform:` section:
```typescript
network_status: 'Network Status',
agents: 'Agent Management',
issues: 'Issues',
fraud_alerts: 'Fraud Alerts',
```

**Steps:**
- [ ] **Step 1:** Add English translations
- [ ] **Step 2:** Add Arabic translations
- [ ] **Step 3:** Add Chinese translations
- [ ] **Step 4:** Commit

---

## Implementation Order

1. **Phase 1:** Database migration
2. **Phase 2:** Domain models
3. **Phase 3:** Backend services (reporting, fraud, notifications)
4. **Phase 4:** API endpoints
5. **Phase 5:** Frontend dashboard + widgets
6. **Phase 6:** Notification settings
7. **Phase 7:** Scheduled jobs
8. **Phase 8:** Testing & polish

---

## Testing Strategy

### Backend Tests
```bash
go test ./internal/service/... -v
go test ./internal/adminapi/... -v
```

### Frontend Tests
```bash
cd web && npm test
```

### Manual Testing Checklist
- [ ] Dashboard loads with all widgets
- [ ] Period toggle (daily/weekly/monthly) works
- [ ] CSV export downloads correct data
- [ ] Network status shows real-time counts
- [ ] Notification settings save correctly
- [ ] Fraud logs display after simulation

---

## Commit History Template

```bash
git add <files>
git commit -m "feat(reporting): <description>

- Add domain models for snapshots, fraud, notifications
- Add reporting service with aggregation
- Add fraud detection service
- Add provider notification service
- Add API endpoints for reporting
- Add frontend dashboard with widgets
- Add notification settings page
- Add scheduled jobs

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

**Plan saved to:** `docs/superpowers/plans/2026-03-22-provider-reporting-alerts-plan.md`
