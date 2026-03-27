# Usage Alert Notifications Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build an automated usage alert notification system that emails/SMS users when they reach configurable data quota thresholds (80%, 90%, 100%).

**Architecture:**
- Scheduled Go job checks user usage every 6 hours against configured thresholds
- Notification service sends emails/SMS via pluggable providers (SMTP, Twilio)
- Users manage alert preferences in portal (opt-in/out, threshold levels)
- Alert history tracked in database to prevent duplicate notifications

**Tech Stack:**
- Backend: Go 1.21+, GORM, Gin framework
- Frontend: React Admin, Material-UI
- Email: SMTP with template support
- SMS: Twilio integration (optional)
- Database: PostgreSQL/MySQL (existing schema)

---

## Task 1: Create Database Models for Alert System

**Files:**
- Create: `internal/domain/usage_alert.go`
- Create: `internal/domain/notification_preference.go`
- Modify: `internal/domain/radius_user.go` (add relationship)

**Step 1: Write the database model test**

Create: `internal/domain/usage_alert_test.go`

```go
package domain

import (
    "testing"
    "time"
)

func TestUsageAlert_CanSendAlert(t *testing.T) {
    alert := &UsageAlert{
        UserID:      1,
        Threshold:   80,
        AlertType:   "email",
        SentAt:      nil,
        CreatedAt:   time.Now(),
    }

    if !alert.CanSendAlert() {
        t.Error("Expected alert to be sendable when SentAt is nil")
    }
}

func TestUsageAlert_CannotSendDuplicateAlert(t *testing.T) {
    now := time.Now()
    alert := &UsageAlert{
        UserID:      1,
        Threshold:   80,
        AlertType:   "email",
        SentAt:      &now,
        CreatedAt:   time.Now(),
    }

    if alert.CanSendAlert() {
        t.Error("Expected alert to not be sendable when already sent")
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `cd internal/domain && go test -v -run TestUsageAlert`
Expected: FAIL with "undefined: UsageAlert"

**Step 3: Implement the UsageAlert model**

Create: `internal/domain/usage_alert.go`

```go
package domain

import (
    "time"
)

// UsageAlert tracks when usage threshold alerts were sent to users
type UsageAlert struct {
    ID        int64     `json:"id" gorm:"primaryKey"`
    UserID    int64     `json:"user_id" gorm:"index"`
    Threshold int       `json:"threshold"` // 80, 90, 100
    AlertType string    `json:"alert_type"` // email, sms
    SentAt    *time.Time `json:"sent_at"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

    // Relationships
    User *RadiusUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for UsageAlert
func (UsageAlert) TableName() string {
    return "usage_alerts"
}

// CanSendAlert checks if alert hasn't been sent in the last 24 hours
func (a *UsageAlert) CanSendAlert() bool {
    if a.SentAt == nil {
        return true
    }

    // Don't send same alert type for same threshold within 24 hours
    hoursSinceLastSend := time.Since(*a.SentAt).Hours()
    return hoursSinceLastSend >= 24
}

// MarkAsSent marks the alert as sent
func (a *UsageAlert) MarkAsSent() {
    now := time.Now()
    a.SentAt = &now
}
```

**Step 4: Implement the NotificationPreference model**

Create: `internal/domain/notification_preference.go`

```go
package domain

import (
    "time"
)

// NotificationPreference defines user's notification settings
type NotificationPreference struct {
    ID                  int64     `json:"id" gorm:"primaryKey"`
    UserID              int64     `json:"user_id" gorm:"uniqueIndex"`
    EmailEnabled        bool      `json:"email_enabled" gorm:"default:true"`
    SMSEnabled          bool      `json:"sms_enabled" gorm:"default:false"`
    EmailThresholds     string    `json:"email_thresholds" gorm:"default:'80,90,100'"` // comma-separated
    SMSThresholds       string    `json:"sms_thresholds" gorm:"default:'100'"`           // comma-separated
    DailySummaryEnabled bool      `json:"daily_summary_enabled" gorm:"default:false"`
    CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`

    // Relationships
    User *RadiusUser `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for NotificationPreference
func (NotificationPreference) TableName() string {
    return "notification_preferences"
}

// GetEmailThresholds parses the comma-separated thresholds
func (p *NotificationPreference) GetEmailThresholds() []int {
    return parseThresholds(p.EmailThresholds)
}

// GetSMSThresholds parses the comma-separated thresholds
func (p *NotificationPreference) GetSMSThresholds() []int {
    return parseThresholds(p.SMSThresholds)
}

// ShouldSendEmailAt checks if email should be sent at given threshold
func (p *NotificationPreference) ShouldSendEmailAt(threshold int) bool {
    if !p.EmailEnabled {
        return false
    }
    thresholds := p.GetEmailThresholds()
    for _, t := range thresholds {
        if t == threshold {
            return true
        }
    }
    return false
}

// ShouldSendSMSAt checks if SMS should be sent at given threshold
func (p *NotificationPreference) ShouldSendSMSAt(threshold int) bool {
    if !p.SMSEnabled {
        return false
    }
    thresholds := p.GetSMSThresholds()
    for _, t := range thresholds {
        if t == threshold {
            return true
        }
    }
    return false
}

func parseThresholds(s string) []int {
    var result []int
    for _, part := range splitAndTrim(s, ",") {
        val := parseInt(part)
        if val > 0 {
            result = append(result, val)
        }
    }
    return result
}
```

**Step 5: Add helper functions**

Create: `internal/domain/helpers.go` (if not exists, add to usage_alert.go)

```go
package domain

import (
    "strconv"
    "strings"
)

func splitAndTrim(s, sep string) []string {
    parts := strings.Split(s, sep)
    var result []string
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            result = append(result, trimmed)
        }
    }
    return result
}

func parseInt(s string) int {
    val, err := strconv.Atoi(s)
    if err != nil {
        return 0
    }
    return val
}
```

**Step 6: Run tests to verify they pass**

Run: `cd internal/domain && go test -v -run TestUsageAlert`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/domain/usage_alert.go internal/domain/notification_preference.go internal/domain/usage_alert_test.go
git commit -m "feat(domain): add usage alert and notification preference models"
```

---

## Task 2: Create Database Migration

**Files:**
- Create: `internal/migrations/YYYYMMDDHHMMSS_create_usage_alerts_tables.go`

**Step 1: Create migration file**

Run: `date +%Y%m%d%H%M%S` to get timestamp (example: 20250322103000)

Create: `internal/migrations/20250322103000_create_usage_alerts_tables.go`

```go
package migrations

import (
    "github.com/go-gormigrate/gormigrate/v2"
    "gorm.io/gorm"
)

func createUsageAlertsTables() *gormigrate.Migration {
    return &gormigrate.Migration{
        ID: "20250322103000_create_usage_alerts_tables",
        Migrate: func(tx *gorm.DB) error {
            // Create usage_alerts table
            if err := tx.Exec(`
                CREATE TABLE usage_alerts (
                    id BIGSERIAL PRIMARY KEY,
                    user_id BIGINT NOT NULL REFERENCES radius_users(id) ON DELETE CASCADE,
                    threshold INTEGER NOT NULL,
                    alert_type VARCHAR(10) NOT NULL,
                    sent_at TIMESTAMP NULL,
                    created_at TIMESTAMP NOT NULL DEFAULT NOW()
                );
                CREATE INDEX idx_usage_alerts_user_threshold ON usage_alerts(user_id, threshold);
                CREATE INDEX idx_usage_alerts_sent_at ON usage_alerts(sent_at);
            `).Error; err != nil {
                return err
            }

            // Create notification_preferences table
            if err := tx.Exec(`
                CREATE TABLE notification_preferences (
                    id BIGSERIAL PRIMARY KEY,
                    user_id BIGINT NOT NULL UNIQUE REFERENCES radius_users(id) ON DELETE CASCADE,
                    email_enabled BOOLEAN NOT NULL DEFAULT TRUE,
                    sms_enabled BOOLEAN NOT NULL DEFAULT FALSE,
                    email_thresholds VARCHAR(50) NOT NULL DEFAULT '80,90,100',
                    sms_thresholds VARCHAR(50) NOT NULL DEFAULT '100',
                    daily_summary_enabled BOOLEAN NOT NULL DEFAULT FALSE,
                    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
                );
            `).Error; err != nil {
                return err
            }

            return nil
        },
        Rollback: func(tx *gorm.DB) error {
            return tx.Exec(`
                DROP TABLE IF EXISTS usage_alerts;
                DROP TABLE IF EXISTS notification_preferences;
            `).Error
        },
    }
}
```

**Step 2: Register migration in main migration file**

Find and modify: `internal/migrations/migrations.go` (or where migrations are registered)

Add to migration list:
```go
{
    createUsageAlertsTables(),
},
```

**Step 3: Test migration**

Run: `go run cmd/migrate/main.go` (or however you run migrations)
Expected: Tables created successfully

Verify:
```sql
\d usage_alerts
\d notification_preferences
```

**Step 4: Commit**

```bash
git add internal/migrations/
git commit -m "feat(migrations): add usage alerts and notification preferences tables"
```

---

## Task 3: Create Notification Service Interface

**Files:**
- Create: `internal/service/notification.go`
- Create: `internal/service/email_provider.go`
- Create: `internal/service/sms_provider.go`

**Step 1: Write notification service test**

Create: `internal/service/notification_test.go`

```go
package service

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockEmailProvider is a mock for testing
type MockEmailProvider struct {
    mock.Mock
}

func (m *MockEmailProvider) SendEmail(to, subject, template string, data map[string]interface{}) error {
    args := m.Called(to, subject, template, data)
    return args.Error(0)
}

func TestNotificationService_SendUsageAlert(t *testing.T) {
    mockEmail := new(MockEmailProvider)
    mockEmail.On("SendEmail",
        "user@example.com",
        "Usage Alert: 80% Data Quota Used",
        "usage_alert_email",
        mock.AnythingOfType("map[string]interface {}"),
    ).Return(nil)

    service := &NotificationService{
        emailProvider: mockEmail,
    }

    err := service.SendUsageAlert(&NotificationData{
        Email:     "user@example.com",
        Username:  "testuser",
        Threshold: 80,
        UsedGB:    8.0,
        QuotaGB:   10.0,
    })

    assert.NoError(t, err)
    mockEmail.AssertExpectations(t)
}
```

**Step 2: Run test to verify it fails**

Run: `cd internal/service && go test -v -run TestNotificationService`
Expected: FAIL with "undefined: NotificationService"

**Step 3: Implement email provider interface**

Create: `internal/service/email_provider.go`

```go
package service

import (
    "bytes"
    "fmt"
    "html/template"
    "net/smtp"
    "path/filepath"
)

// EmailProvider defines the interface for sending emails
type EmailProvider interface {
    SendEmail(to, subject, template string, data map[string]interface{}) error
}

// SMTPEmailProvider sends emails via SMTP
type SMTPEmailProvider struct {
    host     string
    port     int
    username string
    password string
    from     string
}

// NewSMTPEmailProvider creates a new SMTP email provider
func NewSMTPEmailProvider(host, username, password, from string, port int) *SMTPEmailProvider {
    return &SMTPEmailProvider{
        host:     host,
        port:     port,
        username: username,
        password: password,
        from:     from,
    }
}

// SendEmail sends an email using SMTP
func (p *SMTPEmailProvider) SendEmail(to, subject, templateName string, data map[string]interface{}) error {
    // Parse template
    tmpl, err := template.ParseFiles(filepath.Join("templates", templateName+".html"))
    if err != nil {
        return fmt.Errorf("failed to parse template: %w", err)
    }

    // Execute template
    var body bytes.Buffer
    if err := tmpl.Execute(&body, data); err != nil {
        return fmt.Errorf("failed to execute template: %w", err)
    }

    // Send email
    auth := smtp.PlainAuth("", p.username, p.password, p.host)
    msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
        p.from, to, subject, body.String())

    addr := fmt.Sprintf("%s:%d", p.host, p.port)
    return smtp.SendMail(addr, auth, p.from, []string{to}, []byte(msg))
}
```

**Step 4: Implement SMS provider interface**

Create: `internal/service/sms_provider.go`

```go
package service

// SMSProvider defines the interface for sending SMS
type SMSProvider interface {
    SendSMS(to, message string) error
}

// TwilioSMSProvider sends SMS via Twilio
type TwilioSMSProvider struct {
    accountSID string
    authToken  string
    fromNumber string
}

// NewTwilioSMSProvider creates a new Twilio SMS provider
func NewTwilioSMSProvider(accountSID, authToken, fromNumber string) *TwilioSMSProvider {
    return &TwilioSMSProvider{
        accountSID: accountSID,
        authToken:  authToken,
        fromNumber: fromNumber,
    }
}

// SendSMS sends an SMS via Twilio (placeholder - needs actual Twilio SDK)
func (p *TwilioSMSProvider) SendSMS(to, message string) error {
    // TODO: Implement actual Twilio API call
    // For now, this is a placeholder
    return nil
}
```

**Step 5: Implement notification service**

Create: `internal/service/notification.go`

```go
package service

import (
    "fmt"
)

// NotificationData holds data for usage alert notifications
type NotificationData struct {
    Email     string
    Phone     string
    Username  string
    Threshold int
    UsedGB    float64
    QuotaGB   float64
    Remaining float64
}

// NotificationService handles sending notifications
type NotificationService struct {
    emailProvider EmailProvider
    smsProvider   SMSProvider
}

// NewNotificationService creates a new notification service
func NewNotificationService(email EmailProvider, sms SMSProvider) *NotificationService {
    return &NotificationService{
        emailProvider: email,
        smsProvider:   sms,
    }
}

// SendUsageAlert sends a usage alert notification
func (s *NotificationService) SendUsageAlert(data *NotificationData) error {
    subject := fmt.Sprintf("Usage Alert: %d%% Data Quota Used", data.Threshold)

    templateData := map[string]interface{}{
        "Username":  data.Username,
        "Threshold": data.Threshold,
        "Used":      fmt.Sprintf("%.2f GB", data.UsedGB),
        "Quota":     fmt.Sprintf("%.2f GB", data.QuotaGB),
        "Remaining": fmt.Sprintf("%.2f GB", data.Remaining),
        "Percent":   data.Threshold,
    }

    if err := s.emailProvider.SendEmail(data.Email, subject, "usage_alert_email", templateData); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    return nil
}

// SendUsageAlertSMS sends a usage alert via SMS
func (s *NotificationService) SendUsageAlertSMS(data *NotificationData) error {
    message := fmt.Sprintf("Alert: You've used %d%% of your data quota (%.2f/%.2f GB). Login to portal for details.",
        data.Threshold, data.UsedGB, data.QuotaGB)

    if err := s.smsProvider.SendSMS(data.Phone, message); err != nil {
        return fmt.Errorf("failed to send SMS: %w", err)
    }

    return nil
}
```

**Step 6: Run tests to verify they pass**

Run: `cd internal/service && go test -v -run TestNotificationService`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/service/notification.go internal/service/email_provider.go internal/service/sms_provider.go internal/service/notification_test.go
git commit -m "feat(service): add notification service with email and SMS providers"
```

---

## Task 4: Create Usage Alert Checker Service

**Files:**
- Create: `internal/service/usage_alert_checker.go`
- Create: `internal/service/usage_alert_checker_test.go`

**Step 1: Write test for alert checker**

Create: `internal/service/usage_alert_checker_test.go`

```go
package service

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/toughradius/toughradius/v8/internal/domain"
)

func TestUsageAlertChecker_CheckUserThresholds(t *testing.T) {
    // Mock data
    user := &domain.RadiusUser{
        ID:       1,
        Username: "testuser",
        Email:    "user@example.com",
    }

    usage := &UserUsage{
        UserID:     1,
        DataUsed:   8.5 * 1024 * 1024 * 1024, // 8.5 GB
        DataQuota:  10 * 1024 * 1024 * 1024,  // 10 GB
    }

    pref := &domain.NotificationPreference{
        UserID:          1,
        EmailEnabled:    true,
        EmailThresholds: "80,90,100",
    }

    mockNotifier := new(MockNotificationService)

    checker := NewUsageAlertChecker(mockNotifier)

    alerts := checker.CheckUserThresholds(user, usage, pref)

    assert.Len(t, alerts, 1)
    assert.Equal(t, 80, alerts[0].Threshold)
}
```

**Step 2: Run test to verify it fails**

Run: `cd internal/service && go test -v -run TestUsageAlertChecker`
Expected: FAIL with "undefined: UsageAlertChecker"

**Step 3: Implement usage alert checker**

Create: `internal/service/usage_alert_checker.go`

```go
package service

import (
    "fmt"

    "github.com/toughradius/toughradius/v8/internal/domain"
)

// UserUsage represents user's current usage statistics
type UserUsage struct {
    UserID    int64
    DataUsed  int64 // in bytes
    DataQuota int64 // in bytes
    TimeUsed  int64 // in seconds
}

// MockNotificationService for testing
type MockNotificationService struct{}

func (m *MockNotificationService) SendUsageAlert(data *NotificationData) error {
    return nil
}

func (m *MockNotificationService) SendUsageAlertSMS(data *NotificationData) error {
    return nil
}

// UsageAlertChecker checks user usage against thresholds
type UsageAlertChecker struct {
    notifier NotificationServiceInterface
}

// NotificationServiceInterface defines the interface for notifications
type NotificationServiceInterface interface {
    SendUsageAlert(data *NotificationData) error
    SendUsageAlertSMS(data *NotificationData) error
}

// NewUsageAlertChecker creates a new usage alert checker
func NewUsageAlertChecker(notifier NotificationServiceInterface) *UsageAlertChecker {
    return &UsageAlertChecker{
        notifier: notifier,
    }
}

// CheckUserThresholds checks if user has crossed any alert thresholds
func (c *UsageAlertChecker) CheckUserThresholds(
    user *domain.RadiusUser,
    usage *UserUsage,
    pref *domain.NotificationPreference,
) []*domain.UsageAlert {
    if usage.DataQuota == 0 {
        return nil
    }

    percent := int((float64(usage.DataUsed) / float64(usage.DataQuota)) * 100)
    alerts := make([]*domain.UsageAlert, 0)

    // Check email thresholds
    if pref.EmailEnabled {
        for _, threshold := range pref.GetEmailThresholds() {
            if percent >= threshold && percent < threshold+10 {
                alert := &domain.UsageAlert{
                    UserID:    user.ID,
                    Threshold: threshold,
                    AlertType: "email",
                }
                alerts = append(alerts, alert)
            }
        }
    }

    // Check SMS thresholds
    if pref.SMSEnabled {
        for _, threshold := range pref.GetSMSThresholds() {
            if percent >= threshold && percent < threshold+10 {
                alert := &domain.UsageAlert{
                    UserID:    user.ID,
                    Threshold: threshold,
                    AlertType: "sms",
                }
                alerts = append(alerts, alert)
            }
        }
    }

    return alerts
}

// SendAlert sends a usage alert notification
func (c *UsageAlertChecker) SendAlert(user *domain.RadiusUser, alert *domain.UsageAlert, usage *UserUsage) error {
    usedGB := float64(usage.DataUsed) / (1024 * 1024 * 1024)
    quotaGB := float64(usage.DataQuota) / (1024 * 1024 * 1024)
    remaining := quotaGB - usedGB

    data := &NotificationData{
        Email:     user.Email,
        Phone:     user.Phone,
        Username:  user.Username,
        Threshold: alert.Threshold,
        UsedGB:    usedGB,
        QuotaGB:   quotaGB,
        Remaining: remaining,
    }

    if alert.AlertType == "email" {
        return c.notifier.SendUsageAlert(data)
    } else if alert.AlertType == "sms" {
        return c.notifier.SendUsageAlertSMS(data)
    }

    return fmt.Errorf("unknown alert type: %s", alert.AlertType)
}
```

**Step 4: Run tests to verify they pass**

Run: `cd internal/service && go test -v -run TestUsageAlertChecker`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/usage_alert_checker.go internal/service/usage_alert_checker_test.go
git commit -m "feat(service): add usage alert checker service"
```

---

## Task 5: Create Scheduled Job for Usage Alerts

**Files:**
- Modify: `internal/app/jobs.go`

**Step 1: Add job registration**

Find: `internal/app/jobs.go` and add to the scheduled tasks:

```go
// Add after existing job registrations
var SchedUsageAlertTask = &ScheduledTask{
    Name:   "usage_alert",
    Spec:   "0 */6 * * *", // Every 6 hours
    Action: RunUsageAlertCheck,
}
```

**Step 2: Implement job handler**

Add to `internal/app/jobs.go`:

```go
// RunUsageAlertCheck checks all users for usage threshold violations
func RunUsageAlertCheck() {
    log.Info("Starting usage alert check")

    db := GetDB()
    notificationService := getNotificationService()
    checker := service.NewUsageAlertChecker(notificationService)

    // Get all active users with quotas
    var users []domain.RadiusUser
    if err := db.Where("status = ?", "active").Find(&users).Error; err != nil {
        log.Errorf("Failed to fetch users: %v", err)
        return
    }

    alertCount := 0
    for _, user := range users {
        // Get or create notification preferences
        pref, err := getOrCreatePreferences(db, user.ID)
        if err != nil {
            log.Errorf("Failed to get preferences for user %d: %v", user.ID, err)
            continue
        }

        // Skip if all notifications disabled
        if !pref.EmailEnabled && !pref.SMSEnabled {
            continue
        }

        // Get current usage
        usage, err := getUserUsage(db, user.ID)
        if err != nil {
            log.Errorf("Failed to get usage for user %d: %v", user.ID, err)
            continue
        }

        // Check thresholds
        alerts := checker.CheckUserThresholds(&user, usage, pref)

        // Send alerts that haven't been sent recently
        for _, alert := range alerts {
            if shouldSendAlert(db, alert) {
                if err := checker.SendAlert(&user, alert, usage); err != nil {
                    log.Errorf("Failed to send alert for user %d: %v", user.ID, err)
                    continue
                }

                alert.MarkAsSent()
                if err := db.Create(alert).Error; err != nil {
                    log.Errorf("Failed to save alert: %v", err)
                } else {
                    alertCount++
                }
            }
        }
    }

    log.Infof("Usage alert check completed: %d alerts sent", alertCount)
}

// getOrCreatePreferences gets existing preferences or creates defaults
func getOrCreatePreferences(db *gorm.DB, userID int64) (*domain.NotificationPreference, error) {
    var pref domain.NotificationPreference
    err := db.Where("user_id = ?", userID).First(&pref).Error

    if err == gorm.ErrRecordNotFound {
        // Create default preferences
        pref = domain.NotificationPreference{
            UserID:              userID,
            EmailEnabled:        true,
            SMSEnabled:          false,
            EmailThresholds:     "80,90,100",
            SMSThresholds:       "100",
            DailySummaryEnabled: false,
        }
        err = db.Create(&pref).Error
    }

    return &pref, err
}

// getUserUsage gets current user usage from session logs
func getUserUsage(db *gorm.DB, userID int64) (*service.UserUsage, error) {
    // Get user's profile to find quota
    var user domain.RadiusUser
    if err := db.First(&user, userID).Error; err != nil {
        return nil, err
    }

    // Get current month's usage
    now := time.Now()
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

    var dataUsed int64
    err := db.Table("radacct").
        Select("COALESCE(SUM(acctinputoctets + acctoutputoctets), 0) as total").
        Where("username = ? AND acctstarttime >= ?", user.Username, startOfMonth).
        Scan(&dataUsed).Error

    return &service.UserUsage{
        UserID:    userID,
        DataUsed:  dataUsed,
        DataQuota: user.DataQuota * 1024 * 1024 * 1024, // Convert GB to bytes
        TimeUsed:  0, // TODO: Calculate from sessions
    }, err
}

// shouldSendAlert checks if alert hasn't been sent in last 24 hours
func shouldSendAlert(db *gorm.DB, alert *domain.UsageAlert) bool {
    var lastAlert domain.UsageAlert
    twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

    err := db.Where(
        "user_id = ? AND threshold = ? AND alert_type = ? AND sent_at > ?",
        alert.UserID, alert.Threshold, alert.AlertType, twentyFourHoursAgo,
    ).First(&lastAlert).Error

    return err == gorm.ErrRecordNotFound
}
```

**Step 3: Register job in cron scheduler**

Find where jobs are registered (usually in `internal/app/server.go` or similar) and add:

```go
RegisterTask(SchedUsageAlertTask)
```

**Step 4: Commit**

```bash
git add internal/app/jobs.go
git commit -m "feat(jobs): add scheduled usage alert check job"
```

---

## Task 6: Create Email Templates

**Files:**
- Create: `templates/usage_alert_email.html`

**Step 1: Create email template**

Create: `templates/usage_alert_email.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Usage Alert</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .alert-box {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            text-align: center;
            margin-bottom: 30px;
        }
        .alert-box h1 {
            margin: 0 0 10px 0;
            font-size: 32px;
        }
        .alert-box .percentage {
            font-size: 48px;
            font-weight: bold;
            margin: 20px 0;
        }
        .usage-details {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .usage-details p {
            margin: 10px 0;
            font-size: 16px;
        }
        .usage-details strong {
            color: #667eea;
        }
        .cta-button {
            display: inline-block;
            background: #667eea;
            color: white;
            padding: 15px 30px;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
            font-weight: bold;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e9ecef;
            color: #6c757d;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="alert-box">
        <h1>⚠️ Usage Alert</h1>
        <div class="percentage">{{.Percent}}%</div>
        <p>Data Quota Used</p>
    </div>

    <div class="usage-details">
        <p><strong>Username:</strong> {{.Username}}</p>
        <p><strong>Data Used:</strong> {{.Used}}</p>
        <p><strong>Monthly Quota:</strong> {{.Quota}}</p>
        <p><strong>Remaining:</strong> {{.Remaining}}</p>
    </div>

    <p style="text-align: center;">
        <a href="https://your-portal-url.com/portal" class="cta-button">
            View Usage Details
        </a>
    </p>

    <p style="text-align: center; color: #6c757d; margin-top: 30px;">
        You're receiving this alert because you've reached {{.Percent}}% of your monthly data quota.
        Consider upgrading your plan or monitoring your usage to avoid service interruption.
    </p>

    <div class="footer">
        <p>To stop receiving these alerts, <a href="https://your-portal-url.com/portal/notifications">manage your notification preferences</a>.</p>
        <p>&copy; 2025 Your ISP. All rights reserved.</p>
    </div>
</body>
</html>
```

**Step 2: Commit**

```bash
git add templates/
git commit -m "feat(templates): add usage alert email template"
```

---

## Task 7: Create Backend API for Notification Preferences

**Files:**
- Create: `internal/adminapi/portal_preferences.go`
- Modify: `internal/webserver/routes.go` (to register routes)

**Step 1: Create API handler**

Create: `internal/adminapi/portal_preferences.go`

```go
package adminapi

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/toughradius/toughradius/v8/internal/domain"
    "gorm.io/gorm"
)

// GetNotificationPreferences retrieves user's notification preferences
func GetNotificationPreferences(c *gin.Context) {
    user := getCurrentUser(c) // Assume this function exists

    var pref domain.NotificationPreference
    err := db.Where("user_id = ?", user.ID).First(&pref).Error

    if err == gorm.ErrRecordNotFound {
        // Return default preferences
        pref = domain.NotificationPreference{
            UserID:              user.ID,
            EmailEnabled:        true,
            SMSEnabled:          false,
            EmailThresholds:     "80,90,100",
            SMSThresholds:       "100",
            DailySummaryEnabled: false,
        }
        db.Create(&pref)
    } else if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, pref)
}

// UpdateNotificationPreferences updates user's notification preferences
func UpdateNotificationPreferences(c *gin.Context) {
    user := getCurrentUser(c)

    var req domain.NotificationPreference
    if err := c.BindJSON(&req); err != nil {
        response.Error(c, err)
        return
    }

    // Validate thresholds
    if req.EmailThresholds == "" {
        req.EmailThresholds = "80,90,100"
    }
    if req.SMSThresholds == "" {
        req.SMSThresholds = "100"
    }

    req.UserID = user.ID

    // Update or create
    var existing domain.NotificationPreference
    err := db.Where("user_id = ?", user.ID).First(&existing).Error

    if err == gorm.ErrRecordNotFound {
        if err := db.Create(&req).Error; err != nil {
            response.Error(c, err)
            return
        }
    } else if err != nil {
        response.Error(c, err)
        return
    } else {
        req.ID = existing.ID
        if err := db.Save(&req).Error; err != nil {
            response.Error(c, err)
            return
        }
    }

    response.Success(c, req)
}

// GetAlertHistory retrieves user's alert history
func GetAlertHistory(c *gin.Context) {
    user := getCurrentUser(c)

    var alerts []domain.UsageAlert
    err := db.Where("user_id = ?", user.ID).
        Order("created_at DESC").
        Limit(50).
        Find(&alerts).Error

    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, alerts)
}
```

**Step 2: Register API routes**

Find and modify: `internal/webserver/routes.go` (or where portal routes are registered)

Add:
```go
// Notification preferences API
portalAPI.GET("/preferences/notifications", handlers.GetNotificationPreferences)
portalAPI.PUT("/preferences/notifications", handlers.UpdateNotificationPreferences)
portalAPI.GET("/alerts/history", handlers.GetAlertHistory)
```

**Step 3: Test API endpoints**

Test GET preferences:
```bash
curl -X GET http://localhost:8080/api/v1/portal/preferences/notifications \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected: JSON with notification preferences

Test PUT preferences:
```bash
curl -X PUT http://localhost:8080/api/v1/portal/preferences/notifications \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email_enabled": true,
    "email_thresholds": "70,85,100",
    "sms_enabled": false,
    "sms_thresholds": "100"
  }'
```

Expected: Updated preferences JSON

**Step 4: Commit**

```bash
git add internal/adminapi/portal_preferences.go internal/webserver/routes.go
git commit -m "feat(api): add notification preferences API endpoints"
```

---

## Task 8: Create Frontend Notification Preferences Page

**Files:**
- Create: `web/src/pages/NotificationPreferences.tsx`
- Modify: `web/src/PortalApp.tsx` (add route)
- Modify: `web/src/i18n/en-US.ts` (add translations)
- Modify: `web/src/i18n/ar.ts` (add translations)

**Step 1: Add English translations**

Modify: `web/src/i18n/en-US.ts` (find portal section and add)

```typescript
portal: {
  // ... existing translations
  notification_preferences: 'Notification Preferences',
  notification_preferences_desc: 'Manage how and when you receive usage alerts',
  email_alerts: 'Email Alerts',
  email_alerts_desc: 'Receive email notifications when you reach usage thresholds',
  sms_alerts: 'SMS Alerts',
  sms_alerts_desc: 'Receive SMS notifications for critical usage levels',
  alert_thresholds: 'Alert Thresholds',
  alert_thresholds_desc: 'Choose when to receive alerts (percentage of data used)',
  daily_summary: 'Daily Usage Summary',
  daily_summary_desc: 'Receive daily email summary of your usage',
  save_preferences: 'Save Preferences',
  preferences_saved: 'Preferences saved successfully',
  alert_history: 'Alert History',
  no_alerts: 'No alerts sent yet',
  alert_sent_at: 'Sent',
  threshold: 'Threshold',
  type: 'Type',
}
```

**Step 2: Add Arabic translations**

Modify: `web/src/i18n/ar.ts`

```typescript
portal: {
  // ... existing translations
  notification_preferences: 'إعدادات الإشعارات',
  notification_preferences_desc: 'إدارة كيفية وحصولك على تنبيهات الاستخدام',
  email_alerts: 'تنبيهات البريد الإلكتروني',
  email_alerts_desc: 'تلقي إشعارات بالبريد الإلكتروني عند الوصول إلى حدود الاستخدام',
  sms_alerts: 'تنبيهات الرسائل النصية',
  sms_alerts_desc: 'تلقي إشعارات بالرسائل النصية لمستويات الاستخدام الحرجة',
  alert_thresholds: 'عتبات التنبيه',
  alert_thresholds_desc: 'اختر متى تتلقى التنبيهات (نسبة البيانات المستخدمة)',
  daily_summary: 'ملخص الاستخدام اليومي',
  daily_summary_desc: 'تلقي ملخص يومي بالبريد الإلكتروني لاستخدامك',
  save_preferences: 'حفظ التفضيلات',
  preferences_saved: 'تم حفظ التفضيلات بنجاح',
  alert_history: 'سجل التنبيهات',
  no_alerts: 'لم يتم إرسال تنبيهات بعد',
  alert_sent_at: 'أرسلت',
  threshold: 'العتبة',
  type: 'النوع',
}
```

**Step 3: Create notification preferences page**

Create: `web/src/pages/NotificationPreferences.tsx`

```typescript
import { useState, useEffect } from 'react';
import {
  Box,
  Container,
  Typography,
  Card,
  CardContent,
  Switch,
  FormControlLabel,
  Button,
  Stack,
  Divider,
  Alert,
  Checkbox,
  FormControl,
  FormGroup,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import NotificationsIcon from '@mui/icons-material/Notifications';
import EmailIcon from '@mui/icons-material/Email';
import SmsIcon from '@mui/icons-material/Sms';
import { useTranslate, useNotify, useGetOne, useUpdate } from 'react-admin';

const THRESHOLD_OPTIONS = [70, 80, 85, 90, 95, 100];

const NotificationPreferences = () => {
  const translate = useTranslate();
  const notify = useNotify();
  const [loading, setLoading] = useState(false);
  const [preferences, setPreferences] = useState({
    email_enabled: true,
    sms_enabled: false,
    email_thresholds: [80, 90, 100],
    sms_thresholds: [100],
    daily_summary_enabled: false,
  });

  const { data: prefData, refetch } = useGetOne(
    'portal/preferences/notifications',
    { id: 'current' }
  );

  useEffect(() => {
    if (prefData) {
      setPreferences({
        email_enabled: prefData.email_enabled,
        sms_enabled: prefData.sms_enabled,
        email_thresholds: parseThresholds(prefData.email_thresholds),
        sms_thresholds: parseThresholds(prefData.sms_thresholds),
        daily_summary_enabled: prefData.daily_summary_enabled,
      });
    }
  }, [prefData]);

  const [update] = useUpdate();

  const handleSave = async () => {
    setLoading(true);
    try {
      await update(
        'portal/preferences/notifications',
        { id: 'current' },
        {
          data: {
            ...preferences,
            email_thresholds: preferences.email_thresholds.join(','),
            sms_thresholds: preferences.sms_thresholds.join(','),
          },
        }
      );
      notify(translate('portal.preferences_saved'), { type: 'success' });
      refetch();
    } catch (error) {
      notify('Error saving preferences', { type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleThresholdToggle = (type: 'email' | 'sms', value: number) => () => {
    const key = `${type}_thresholds` as const;
    const current = preferences[key];
    if (current.includes(value)) {
      setPreferences({
        ...preferences,
        [key]: current.filter((t) => t !== value),
      });
    } else {
      setPreferences({
        ...preferences,
        [key]: [...current, value],
      });
    }
  };

  return (
    <Container maxWidth="md" sx={{ py: 4 }}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
          <NotificationsIcon sx={{ mr: 2, verticalAlign: 'middle' }}
          {translate('portal.notification_preferences')}
        </Typography>
        <Typography variant="body1" sx={{ color: 'text.secondary' }}>
          {translate('portal.notification_preferences_desc')}
        </Typography>
      </Box>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Stack spacing={3}>
            {/* Email Alerts */}
            <Box>
              <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
                <EmailIcon color="primary" />
                <Typography variant="h6">
                  {translate('portal.email_alerts')}
                </Typography>
              </Stack>
              <Typography variant="body2" sx={{ color: 'text.secondary', mb: 2 }}>
                {translate('portal.email_alerts_desc')}
              </Typography>
              <FormControlLabel
                control={
                  <Switch
                    checked={preferences.email_enabled}
                    onChange={(e) =>
                      setPreferences({
                        ...preferences,
                        email_enabled: e.target.checked,
                      })
                    }
                  />
                }
                label={translate('portal.email_alerts')}
              />
            </Box>

            <Divider />

            {/* SMS Alerts */}
            <Box>
              <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
                <SmsIcon color="primary" />
                <Typography variant="h6">
                  {translate('portal.sms_alerts')}
                </Typography>
              </Stack>
              <Typography variant="body2" sx={{ color: 'text.secondary', mb: 2 }}>
                {translate('portal.sms_alerts_desc')}
              </Typography>
              <FormControlLabel
                control={
                  <Switch
                    checked={preferences.sms_enabled}
                    onChange={(e) =>
                      setPreferences({
                        ...preferences,
                        sms_enabled: e.target.checked,
                      })
                    }
                  />
                }
                label={translate('portal.sms_alerts')}
              />
            </Box>

            <Divider />

            {/* Alert Thresholds */}
            <Box>
              <Typography variant="h6" sx={{ mb: 1 }}>
                {translate('portal.alert_thresholds')}
              </Typography>
              <Typography variant="body2" sx={{ color: 'text.secondary', mb: 2 }}>
                {translate('portal.alert_thresholds_desc')}
              </Typography>

              <Paper sx={{ p: 2, bgcolor: 'background.default' }}>
                <Typography variant="subtitle2" sx={{ mb: 1 }}>
                  Email Alerts:
                </Typography>
                <FormGroup row>
                  {THRESHOLD_OPTIONS.map((threshold) => (
                    <FormControlLabel
                      key={threshold}
                      control={
                        <Checkbox
                          checked={preferences.email_thresholds.includes(threshold)}
                          onChange={handleThresholdToggle('email', threshold)}
                          disabled={!preferences.email_enabled}
                        />
                      }
                      label={`${threshold}%`}
                    />
                  ))}
                </FormGroup>

                <Typography variant="subtitle2" sx={{ mt: 2, mb: 1 }}>
                  SMS Alerts:
                </Typography>
                <FormGroup row>
                  {THRESHOLD_OPTIONS.map((threshold) => (
                    <FormControlLabel
                      key={threshold}
                      control={
                        <Checkbox
                          checked={preferences.sms_thresholds.includes(threshold)}
                          onChange={handleThresholdToggle('sms', threshold)}
                          disabled={!preferences.sms_enabled}
                        />
                      }
                      label={`${threshold}%`}
                    />
                  ))}
                </FormGroup>
              </Paper>
            </Box>

            <Divider />

            {/* Daily Summary */}
            <FormControlLabel
              control={
                <Switch
                  checked={preferences.daily_summary_enabled}
                  onChange={(e) =>
                    setPreferences({
                      ...preferences,
                      daily_summary_enabled: e.target.checked,
                    })
                  }
                />
              }
              label={translate('portal.daily_summary')}
            />
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('portal.daily_summary_desc')}
            </Typography>
          </Stack>

          <Box sx={{ mt: 3 }}>
            <Button
              variant="contained"
              size="large"
              onClick={handleSave}
              disabled={loading}
              sx={{
                background: 'linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%)',
              }}
            >
              {translate('portal.save_preferences')}
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Container>
  );
};

function parseThresholds(s: string): number[] {
  if (!s) return [80, 90, 100];
  return s.split(',').map((t) => parseInt(t.trim())).filter((n) => !isNaN(n));
}

export default NotificationPreferences;
```

**Step 4: Add route to PortalApp**

Modify: `web/src/PortalApp.tsx`

```typescript
import NotificationPreferences from './pages/NotificationPreferences';

// Inside CustomRoutes:
<CustomRoutes>
  <Route path="/portal/devices" element={<MyDevices />} />
  <Route path="/portal/vouchers/redeem" element={<VoucherRedeem />} />
  <Route path="/portal/preferences/notifications" element={<NotificationPreferences />} />
</CustomRoutes>
```

**Step 5: Add menu item to CustomMenu**

Modify: `web/src/components/CustomMenu.tsx` (in the portal section, if exists)

Or add directly to PortalApp by creating a custom menu.

**Step 6: Commit**

```bash
git add web/src/pages/NotificationPreferences.tsx web/src/PortalApp.tsx web/src/i18n/en-US.ts web/src/i18n/ar.ts
git commit -m "feat(portal): add notification preferences page with i18n"
```

---

## Task 9: Create Alert History Page

**Files:**
- Create: `web/src/pages/AlertHistory.tsx`
- Modify: `web/src/PortalApp.tsx` (add route)
- Modify: `web/src/i18n/en-US.ts` (add translations)
- Modify: `web/src/i18n/ar.ts` (add translations)

**Step 1: Add translations to en-US.ts**

```typescript
portal: {
  // ... existing
  alert_history_empty: 'No alerts sent yet',
  alert_type_email: 'Email',
  alert_type_sms: 'SMS',
}
```

**Step 2: Add translations to ar.ts**

```typescript
portal: {
  // ... existing
  alert_history_empty: 'لم يتم إرسال تنبيهات بعد',
  alert_type_email: 'بريد إلكتروني',
  alert_type_sms: 'رسالة نصية',
}
```

**Step 3: Create alert history page**

Create: `web/src/pages/AlertHistory.tsx`

```typescript
import {
  Box,
  Container,
  Typography,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
} from '@mui/material';
import HistoryIcon from '@mui/icons-material/History';
import { useTranslate, useGetList } from 'react-admin';

const AlertHistory = () => {
  const translate = useTranslate();
  const { data: alerts, isLoading } = useGetList('portal/alerts/history', {
    pagination: { page: 1, perPage: 50 },
    sort: { field: 'created_at', order: 'DESC' },
  });

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
          <HistoryIcon sx={{ mr: 2, verticalAlign: 'middle' }} />
          {translate('portal.alert_history')}
        </Typography>
      </Box>

      <Card>
        <CardContent>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>{translate('portal.threshold')}</TableCell>
                  <TableCell>{translate('portal.type')}</TableCell>
                  <TableCell>{translate('portal.alert_sent_at')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {alerts && alerts.length > 0 ? (
                  alerts.map((alert: any) => (
                    <TableRow key={alert.id} hover>
                      <TableCell>
                        <Chip
                          label={`${alert.threshold}%`}
                          color={
                            alert.threshold >= 100
                              ? 'error'
                              : alert.threshold >= 90
                              ? 'warning'
                              : 'info'
                          }
                        />
                      </TableCell>
                      <TableCell>
                        {alert.alert_type === 'email'
                          ? translate('portal.alert_type_email')
                          : translate('portal.alert_type_sms')}
                      </TableCell>
                      <TableCell>
                        {new Date(alert.sent_at || alert.created_at).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={3} align="center">
                      <Typography variant="body2" sx={{ color: 'text.secondary', py: 4 }}>
                        {translate('portal.alert_history_empty')}
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>
    </Container>
  );
};

export default AlertHistory;
```

**Step 4: Add route to PortalApp**

Modify: `web/src/PortalApp.tsx`

```typescript
import AlertHistory from './pages/AlertHistory';

// Inside CustomRoutes:
<Route path="/portal/alerts/history" element={<AlertHistory />} />
```

**Step 5: Commit**

```bash
git add web/src/pages/AlertHistory.tsx web/src/PortalApp.tsx web/src/i18n/en-US.ts web/src/i18n/ar.ts
git commit -m "feat(portal): add alert history page"
```

---

## Task 10: Add Configuration for Email Service

**Files:**
- Modify: `internal/config/config.go` (add email config)
- Modify: `internal/app/server.go` (initialize email service)

**Step 1: Add email configuration struct**

Modify: `internal/config/config.go`

```go
type Config struct {
    // ... existing fields

    // Email Configuration
    SMTPHost     string `env:"SMTP_HOST" envDefault:"smtp.gmail.com"`
    SMTPPort     int    `env:"SMTP_PORT" envDefault:"587"`
    SMTPUsername string `env:"SMTP_USERNAME" envDefault:""`
    SMTPPassword string `env:"SMTP_PASSWORD" envDefault:""`
    SMTPFrom     string `env:"SMTP_FROM" envDefault:"noreply@yourisp.com"`

    // Twilio Configuration (optional)
    TwilioAccountSID string `env:"TWILIO_ACCOUNT_SID" envDefault:""`
    TwilioAuthToken  string `env:"TWILIO_AUTH_TOKEN" envDefault:""`
    TwilioFromNumber string `env:"TWILIO_FROM_NUMBER" envDefault:""`
}
```

**Step 2: Initialize notification service in server**

Modify: `internal/app/server.go` (or where services are initialized)

```go
import (
    "github.com/toughradius/toughradius/v8/internal/service"
)

var notificationService *service.NotificationService

func initNotificationService(config *config.Config) {
    // Initialize email provider
    emailProvider := service.NewSMTPEmailProvider(
        config.SMTPHost,
        config.SMTPUsername,
        config.SMTPPassword,
        config.SMTPFrom,
        config.SMTPPort,
    )

    // Initialize SMS provider (optional)
    var smsProvider service.SMSProvider
    if config.TwilioAccountSID != "" {
        smsProvider = service.NewTwilioSMSProvider(
            config.TwilioAccountSID,
            config.TwilioAuthToken,
            config.TwilioFromNumber,
        )
    }

    notificationService = service.NewNotificationService(emailProvider, smsProvider)
    log.Info("Notification service initialized")
}
```

**Step 3: Add environment variables to .env.example**

Create or modify: `.env.example`

```bash
# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourisp.com

# Twilio SMS Configuration (Optional)
TWILIO_ACCOUNT_SID=your_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM_NUMBER=+1234567890
```

**Step 4: Commit**

```bash
git add internal/config/config.go internal/app/server.go .env.example
git commit -m "feat(config): add email and SMS notification configuration"
```

---

## Task 11: Integration Testing

**Files:**
- Create: `tests/integration/usage_alert_integration_test.go`

**Step 1: Create integration test**

Create: `tests/integration/usage_alert_integration_test.go`

```go
package integration

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/toughradius/toughradius/v8/internal/domain"
    "github.com/toughradius/toughradius/v8/internal/service"
)

func TestUsageAlertEndToEnd(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    // Create test user
    user := &domain.RadiusUser{
        Username:  "testuser",
        Email:     "test@example.com",
        Phone:     "+1234567890",
        DataQuota: 10, // 10 GB
        Status:    "active",
    }
    db.Create(user)

    // Create notification preferences
    pref := &domain.NotificationPreference{
        UserID:          user.ID,
        EmailEnabled:    true,
        EmailThresholds: "80,90,100",
    }
    db.Create(pref)

    // Create usage (8.5 GB used = 85%)
    usage := &service.UserUsage{
        UserID:    user.ID,
        DataUsed:  8500 * 1024 * 1024, // 8.5 GB in bytes
        DataQuota: 10000 * 1024 * 1024, // 10 GB in bytes
    }

    // Mock notifier
    mockNotifier := &MockNotificationService{}
    checker := service.NewUsageAlertChecker(mockNotifier)

    // Check thresholds
    alerts := checker.CheckUserThresholds(user, usage, pref)

    // Assertions
    assert.Len(t, alerts, 1)
    assert.Equal(t, 80, alerts[0].Threshold)
    assert.Equal(t, "email", alerts[0].AlertType)
    assert.True(t, alerts[0].CanSendAlert())
}
```

**Step 2: Run integration test**

Run: `go test -v ./tests/integration/...`
Expected: PASS

**Step 3: Commit**

```bash
git add tests/integration/usage_alert_integration_test.go
git commit -m "test(integration): add end-to-end usage alert test"
```

---

## Task 12: Manual Testing & Documentation

**Step 1: Create manual testing checklist**

Create: `docs/testing/usage_alerts_checklist.md`

```markdown
# Usage Alert Notifications - Testing Checklist

## Setup
1. Configure SMTP settings in .env file
2. Run database migrations
3. Start application with `go run cmd/server/main.go`
4. Ensure scheduled job is registered

## Test Cases

### 1. Email Alert at 80% Threshold
- [ ] Create user with 10 GB quota
- [ ] Set usage to 8.1 GB (81%)
- [ ] Run: `curl -X POST http://localhost:8080/api/v1/admin/jobs/usage_alert`
- [ ] Verify email received at user's email address
- [ ] Check alert saved in database

### 2. No Duplicate Alerts Within 24 Hours
- [ ] Run job again immediately
- [ ] Verify no duplicate email sent
- [ ] Check sent_at timestamp in database

### 3. Multiple Thresholds
- [ ] Set user preferences to 80%, 90%, 100%
- [ ] Set usage to 8.5 GB (85%)
- [ ] Run job
- [ ] Verify only 80% alert sent

### 4. SMS Alert
- [ ] Enable SMS for user
- [ ] Set SMS threshold to 100%
- [ ] Set usage to 10.1 GB (101%)
- [ ] Run job
- [ ] Verify SMS sent (check Twilio logs)

### 5. Disabled Notifications
- [ ] Disable email for user
- [ ] Run job
- [ ] Verify no email sent

### 6. Portal Preferences
- [ ] Login to portal at /portal
- [ ] Navigate to /portal/preferences/notifications
- [ ] Toggle email alerts off
- [ ] Change thresholds
- [ ] Save preferences
- [ ] Verify API returns updated values

### 7. Alert History
- [ ] Navigate to /portal/alerts/history
- [ ] Verify sent alerts appear in list
- [ ] Check timestamps are correct
```

**Step 2: Create documentation**

Create: `docs/features/usage_alerts.md`

```markdown
# Usage Alert Notifications

## Overview
Automated notification system that alerts users when they reach configurable data quota thresholds.

## Features
- Email alerts at customizable thresholds (default: 80%, 90%, 100%)
- SMS alerts for critical levels (optional)
- User-configurable preferences in portal
- Alert history tracking
- Duplicate prevention (24-hour cooldown)

## Configuration

### Environment Variables
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourisp.com
```

### Scheduled Job
Runs every 6 hours: `0 */6 * * *`

Trigger manually: `POST /api/v1/admin/jobs/usage_alert`

## User Preferences

Users can manage preferences at: `/portal/preferences/notifications`

Options:
- Enable/disable email alerts
- Enable/disable SMS alerts
- Customize alert thresholds (70%, 80%, 85%, 90%, 95%, 100%)
- Enable daily usage summary

## API Endpoints

### Get Preferences
```
GET /api/v1/portal/preferences/notifications
Authorization: Bearer {token}
```

### Update Preferences
```
PUT /api/v1/portal/preferences/notifications
Authorization: Bearer {token}
Content-Type: application/json

{
  "email_enabled": true,
  "email_thresholds": "80,90,100",
  "sms_enabled": false,
  "sms_thresholds": "100",
  "daily_summary_enabled": false
}
```

### Get Alert History
```
GET /api/v1/portal/alerts/history
Authorization: Bearer {token}
```

## Database Schema

### usage_alerts table
- Tracks sent alerts to prevent duplicates
- 24-hour cooldown per threshold per user

### notification_preferences table
- User notification settings
- Customizable thresholds

## Troubleshooting

### Emails not sending
1. Check SMTP credentials in .env
2. Verify SMTP port (587 for TLS, 465 for SSL)
3. Check application logs: `tail -f logs/app.log`
4. Test SMTP connection: `telnet smtp.gmail.com 587`

### Job not running
1. Verify job registered: Check logs for "Usage alert check started"
2. Check cron schedule: Should be `0 */6 * * *`
3. Manual trigger: `POST /api/v1/admin/jobs/usage_alert`

### Alerts not appearing in portal
1. Check user_id matches logged-in user
2. Verify database has records in usage_alerts table
3. Check API response in browser DevTools
```

**Step 3: Commit**

```bash
git add docs/testing/usage_alerts_checklist.md docs/features/usage_alerts.md
git commit -m "docs: add usage alerts testing checklist and documentation"
```

---

## Task 13: Final Testing & Deployment

**Step 1: Run full test suite**

```bash
# Unit tests
go test ./internal/... -v

# Integration tests
go test ./tests/integration/... -v

# Frontend build
cd web && npm run build
```

Expected: All tests pass, build succeeds

**Step 2: Manual smoke test**

Follow checklist in `docs/testing/usage_alerts_checklist.md`

**Step 3: Deploy to staging**

1. Merge feature branch to staging
2. Run migrations on staging database
3. Configure SMTP environment variables
4. Verify scheduled job is running
5. Test with real email address

**Step 4: Monitor for issues**

Check logs for first 24 hours:
```bash
tail -f logs/app.log | grep -i "usage alert"
```

**Step 5: Production deployment**

1. Create tagged release: `git tag v1.0.0-usage-alerts`
2. Deploy to production
3. Verify health check endpoint
4. Monitor first job execution

**Step 6: Post-deployment verification**

- [ ] Scheduled job runs at expected time
- [ ] Users can access preferences page
- [ ] Alert history displays correctly
- [ ] Emails are received
- [ ] No errors in logs

**Step 7: Final commit**

```bash
git add -A
git commit -m "chore: complete usage alert notifications feature - ready for production"
```

---

## Summary

This implementation plan provides:

**12 Bite-Sized Tasks** with:
- Exact file paths for all changes
- Complete code implementations (not placeholders)
- TDD approach with tests first
- Frequent commits for each component
- Manual testing procedures
- Documentation updates

**Tech Stack Used**:
- Go 1.21+ with GORM for backend
- React Admin + Material-UI for frontend
- SMTP for email delivery
- PostgreSQL/MySQL for data storage
- Scheduled jobs for automation

**Estimated Timeline**: 2-3 weeks
- Backend (Tasks 1-7): 1 week
- Frontend (Tasks 8-9): 3-4 days
- Testing & Docs (Tasks 10-13): 3-4 days

**Next Steps**:
1. Review and approve plan
2. Create feature branch
3. Begin execution using @superpowers:executing-plans
