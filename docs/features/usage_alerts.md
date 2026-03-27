# Usage Alert Notifications

## Overview

Automated notification system that alerts users when they reach configurable data quota thresholds (80%, 90%, 100%). Helps users monitor their usage and avoid unexpected service interruptions.

## Features

- **Email alerts** at customizable thresholds (default: 80%, 90%, 100%)
- **SMS alerts** for critical levels (optional, via Twilio)
- **User-configurable preferences** in customer portal
- **Alert history tracking** in database
- **Duplicate prevention** (24-hour cooldown per threshold)
- **Scheduled checks** every 6 hours (configurable)

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Scheduled Job  │───>│ UsageAlertChecker│───>│ SMTP/Twilio     │
│  (every 6 hours)│    │                  │    │ (Email/SMS)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │
         v                       v
┌─────────────────┐    ┌──────────────────┐
│  RadiusUser     │    │ UsageAlert       │
│  (with DataQuota)│    │ (sent history)   │
└─────────────────┘    └──────────────────┘
         │
         v
┌─────────────────┐
│ NotificationPref │
│ (user settings)  │
└─────────────────┘
```

## Database Schema

### usage_alerts table

Tracks sent alerts to prevent duplicates.

| Column     | Type        | Description                      |
|------------|-------------|----------------------------------|
| id         | BIGSERIAL   | Primary key                      |
| user_id    | BIGINT      | FK to radius_users               |
| threshold  | INTEGER     | 80, 90, or 100                   |
| alert_type | VARCHAR(10) | 'email' or 'sms'                 |
| sent_at    | TIMESTAMP   | When alert was sent              |
| created_at | TIMESTAMP   | Record creation time             |

### notification_preferences table

User notification settings.

| Column                 | Type        | Description                    |
|------------------------|-------------|--------------------------------|
| id                     | BIGSERIAL   | Primary key                    |
| user_id                | BIGINT      | FK to radius_users (unique)    |
| email_enabled          | BOOLEAN     | Enable email alerts            |
| sms_enabled            | BOOLEAN     | Enable SMS alerts              |
| email_thresholds       | VARCHAR(50) | Comma-separated: "80,90,100"  |
| sms_thresholds         | VARCHAR(50) | Comma-separated: "100"         |
| daily_summary_enabled  | BOOLEAN     | Daily usage summary            |
| created_at             | TIMESTAMP   | Record creation time           |
| updated_at             | TIMESTAMP   | Record update time             |

## Configuration

### toughradius.yml

```yaml
notification:
  enabled: true
  alert_check_cron: "0 */6 * * *"  # Every 6 hours
  smtp_host: smtp.example.com
  smtp_port: 587
  smtp_username: your-email@example.com
  smtp_password: your-app-password
  smtp_from: noreply@example.com
  # Twilio (optional)
  twilio_account_sid: ""
  twilio_auth_token: ""
  twilio_from_number: ""
```

### Environment Variables

| Variable                              | Description           |
|---------------------------------------|-----------------------|
| TOUGHRADIUS_NOTIFICATION_ENABLED      | Enable/disable feature|
| TOUGHRADIUS_NOTIFICATION_SMTP_HOST    | SMTP server           |
| TOUGHRADIUS_NOTIFICATION_SMTP_PORT    | SMTP port             |
| TOUGHRADIUS_NOTIFICATION_SMTP_USERNAME| SMTP username        |
| TOUGHRADIUS_NOTIFICATION_SMTP_PASSWORD| SMTP password        |
| TOUGHRADIUS_NOTIFICATION_SMTP_FROM    | From email address    |
| TOUGHRADIUS_NOTIFICATION_TWILIO_ACCOUNT_SID | Twilio SID    |
| TOUGHRADIUS_NOTIFICATION_TWILIO_AUTH_TOKEN  | Twilio Token |
| TOUGHRADIUS_NOTIFICATION_TWILIO_FROM_NUMBER  | Twilio number |

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

## Frontend Pages

- **Notification Preferences**: `/portal/preferences/notifications`
- **Alert History**: `/portal/alerts/history`

## Files Added

### Backend

| File | Description |
|------|-------------|
| `internal/domain/usage_alert.go` | UsageAlert model |
| `internal/domain/notification_preference.go` | NotificationPreference model |
| `internal/domain/helpers.go` | Helper functions |
| `internal/migration/usage_alerts.go` | Database migration |
| `internal/service/notification.go` | NotificationService |
| `internal/service/email_provider.go` | SMTP email provider |
| `internal/service/sms_provider.go` | Twilio SMS provider |
| `internal/service/usage_alert_checker.go` | Alert checker logic |
| `internal/app/jobs.go` | Scheduled job |
| `internal/adminapi/portal_notifications.go` | API handlers |
| `config/config.go` | Configuration struct |

### Frontend

| File | Description |
|------|-------------|
| `web/src/pages/NotificationPreferences.tsx` | Preferences page |
| `web/src/pages/AlertHistory.tsx` | Alert history page |
| `web/src/i18n/en-US.ts` | English translations |
| `web/src/i18n/ar.ts` | Arabic translations |
| `web/src/i18n/zh-CN.ts` | Chinese translations |

### Templates

| File | Description |
|------|-------------|
| `templates/usage_alert_email.html` | HTML email template |

## Testing

See [Testing Checklist](./testing/usage_alerts_checklist.md) for detailed test procedures.

## Running Tests

```bash
# Unit tests
go test ./internal/service/... -v

# Domain tests
go test ./internal/domain/... -v

# Migration tests
go test ./internal/migration/... -v
```
