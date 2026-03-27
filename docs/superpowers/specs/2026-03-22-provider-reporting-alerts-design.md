# Provider Admin Dashboard & Alerts System - Design Specification

> **Date:** 2026-03-22
> **Feature:** Comprehensive Provider Admin Reporting Dashboard with Integrated Alerts

---

## 1. Overview

Build an integrated reporting and alerting system for Provider Admins that combines:
- **Reporting Dashboard**: Daily/Weekly/Monthly summaries of users, vouchers, sessions, data usage
- **Network Status Widget**: Real-time overview of nodes, servers, CPEs
- **Agent Financial Reporting**: Revenue tracking, batch management
- **Device Issues Reporter**: Categorized network/device problems
- **Fraud Detection**: Rate-limiting based voucher fraud prevention
- **Provider Notifications**: Configurable alerts for resource usage (percentage, absolute, anomaly)

---

## 2. Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Provider Admin Frontend                      │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐   │
│  │ Reporting │ │ Network   │ │ Agent     │ │ Issues &  │   │
│  │ Dashboard │ │ Status    │ │ Finance   │ │ Fraud     │   │
│  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘   │
│        └─────────────┴─────────────┴─────────────┘          │
│                            │                                 │
│                    /api/v1/reporting/*                       │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────┴────────────────────────────────┐
│                   Backend Services (Go)                       │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐ │
│  │ReportingService│  │NotificationSvc │  │ FraudDetector  │ │
│  └───────┬────────┘  └───────┬────────┘  └───────┬────────┘ │
│          │                   │                    │          │
│  ┌───────┴──────────────────┴────────────────────┴────────┐│
│  │                   Aggregation Engine                     ││
│  └────────────────────────┬───────────────────────────────┘│
│                           │                                 │
│  ┌────────────────────────┴───────────────────────────────┐│
│  │              Snapshot Tables (Hybrid Storage)             ││
│  │  - Detailed: last 30 days                               ││
│  │  - Daily/Weekly/Monthly aggregates                       ││
│  └──────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Storage Strategy (Hybrid)

| Data Type | Retention | Storage |
|-----------|-----------|---------|
| Detailed logs | 30 days | `radacct`, existing tables |
| Daily snapshots | 90 days | `reporting_daily_snapshots` |
| Weekly snapshots | 12 months | `reporting_weekly_snapshots` |
| Monthly snapshots | 5 years | `reporting_monthly_snapshots` |

---

## 3. Database Schema

### 3.1 New Tables

#### `reporting_daily_snapshots`
```sql
CREATE TABLE reporting_daily_snapshots (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL,
    snapshot_date DATE NOT NULL,
    -- User metrics
    total_users INT DEFAULT 0,
    active_users INT DEFAULT 0,
    new_monthly_users INT DEFAULT 0,
    new_voucher_users INT DEFAULT 0,
    -- Session metrics
    total_sessions INT DEFAULT 0,
    active_sessions INT DEFAULT 0,
    -- Data metrics
    monthly_data_used_bytes BIGINT DEFAULT 0,
    voucher_data_used_bytes BIGINT DEFAULT 0,
    -- Network metrics
    active_nodes INT DEFAULT 0,
    total_nodes INT DEFAULT 0,
    active_servers INT DEFAULT 0,
    total_servers INT DEFAULT 0,
    active_cpes INT DEFAULT 0,
    total_cpes INT DEFAULT 0,
    -- Agent metrics
    total_agents INT DEFAULT 0,
    total_batches INT DEFAULT 0,
    agent_revenue DECIMAL(15,2) DEFAULT 0,
    mrr DECIMAL(15,2) DEFAULT 0,
    -- Issues metrics
    device_issues INT DEFAULT 0,
    network_issues INT DEFAULT 0,
    fraud_attempts INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(provider_id, snapshot_date)
);
```

#### `reporting_fraud_log`
```sql
CREATE TABLE reporting_fraud_log (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL,
    voucher_id BIGINT,
    user_id BIGINT,
    ip_address VARCHAR(45),
    event_type VARCHAR(50),
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### `provider_notification_preferences`
```sql
CREATE TABLE provider_notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL UNIQUE,
    -- Percentage-based alerts
    alert_percentages VARCHAR(50) DEFAULT '70,85,100',
    alert_percentages_enabled BOOLEAN DEFAULT true,
    -- Absolute thresholds
    max_users_threshold INT,
    max_data_bytes_threshold BIGINT,
    absolute_alerts_enabled BOOLEAN DEFAULT false,
    -- Anomaly detection
    anomaly_detection_enabled BOOLEAN DEFAULT false,
    anomaly_threshold_percent INT DEFAULT 50,
    -- Notification channels
    email_enabled BOOLEAN DEFAULT true,
    sms_enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### `reporting_network_issues`
```sql
CREATE TABLE reporting_network_issues (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL,
    device_type VARCHAR(20), -- 'node', 'server', 'nas', 'cpe', 'user_device'
    device_id BIGINT,
    device_name VARCHAR(255),
    issue_type VARCHAR(50),
    issue_details TEXT,
    status VARCHAR(20) DEFAULT 'open', -- 'open', 'resolved', 'ignored'
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 4. API Endpoints

### 4.1 Reporting Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/reporting/summary` | Get daily/weekly/monthly summary |
| GET | `/reporting/users` | User metrics with trend |
| GET | `/reporting/sessions` | Session metrics |
| GET | `/reporting/network-status` | Network device status |
| GET | `/reporting/agents` | Agent financial summary |
| GET | `/reporting/issues` | Device/network issues |
| GET | `/reporting/fraud` | Fraud detection log |
| GET | `/reporting/export` | Export CSV |

### 4.2 Notification Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/reporting/notifications/preferences` | Get notification settings |
| PUT | `/reporting/notifications/preferences` | Update notification settings |

---

## 5. Components

### 5.1 Frontend Components

| Component | Location | Description |
|-----------|----------|-------------|
| `ProviderReportingDashboard` | `pages/Platform/` | Main reporting dashboard page |
| `NetworkStatusWidget` | `components/Platform/` | Network devices overview |
| `AgentFinancialWidget` | `components/Platform/` | Agent revenue cards |
| `IssuesReporterWidget` | `components/Platform/` | Issues list with categories |
| `FraudAlertWidget` | `components/Platform/` | Fraud detection alerts |
| `NotificationPreferences` | `pages/Platform/` | Alert settings form |
| `ExportButton` | `components/` | CSV export functionality |
| `SummaryCard` | `components/` | Reusable metric card with trend |

### 5.2 Backend Services

| Service | Location | Description |
|---------|----------|-------------|
| `ReportingService` | `service/reporting.go` | Core reporting logic |
| `AggregationService` | `service/aggregation.go` | Snapshot rollup job |
| `FraudDetectionService` | `service/fraud.go` | Fraud detection engine |
| `NotificationService` | `service/provider_notifications.go` | Alert generation |

---

## 6. Features

### 6.1 Reporting Dashboard

**Summary View:**
- Toggle: Daily / Weekly / Monthly
- Date range selector
- Key metrics cards with trend indicators
- Export to CSV button

**Metrics Displayed:**
| Metric | Daily | Weekly | Monthly |
|--------|-------|--------|---------|
| Total Users | ✓ | ✓ | ✓ |
| New Monthly Users | ✓ | ✓ | ✓ |
| New Voucher Users | ✓ | ✓ | ✓ |
| Active Sessions | ✓ | ✓ | ✓ |
| Monthly Data Used (GB) | ✓ | ✓ | ✓ |
| Voucher Data Used (GB) | ✓ | ✓ | ✓ |
| Agent Revenue | ✓ | ✓ | ✓ |
| MRR | ✓ | ✓ | ✓ |

### 6.2 Network Status Widget

**Real-time** counts with status indicators (fetched directly from database, no caching):
- Nodes: `12/15 active`
- Servers: `5/6 active`
- CPEs: `89/100 active`

Note: Network status queries the database directly for accurate real-time counts. This is acceptable because the queries are simple COUNT queries with indexes.

### 6.3 Agent Financial Widget

Per-agent summary:
- Agent name
- Total batches created
- Revenue collected
- Commission earned

### 6.4 Issues Reporter

Categorized issue list:
| Category | Issue Types |
|----------|------------|
| Node Issues | Down, High Latency, Capacity Warning |
| Server Issues | Down, Authentication Failures, Overload |
| CPE Issues | Provisioning Failed, CPE Reboot Loop |
| User Device | Authentication Failures, Session Drops |

### 6.5 Fraud Detection

Rate-limiting rules:
| Rule | Threshold | Action |
|------|-----------|--------|
| IP Activation Limit | 5 activations/IP/hour | Flag & log |
| Same Voucher Multi-Use | 2+ different IPs | Quarantine voucher |
| Rapid Successive | 10+ attempts/minute | Temporary block |

### 6.6 Provider Notifications

Configurable alerts:

**Percentage-Based:**
- 70% plan usage → Warning email
- 85% plan usage → Alert email
- 100% plan usage → Critical email

**Absolute Thresholds:**
- Max users (e.g., 1000)
- Max data (e.g., 10TB)
- Max sessions (e.g., 500)

**Anomaly Detection:**
- >50% deviation from weekly average
- Sudden drop in active sessions
- Unusual data usage spike

---

## 7. Scheduled Jobs

### 7.1 Daily Snapshot Job
- **Schedule:** `0 1 * * *` (1 AM daily)
- **Action:** Create daily snapshot record

### 7.2 Weekly Rollup Job
- **Schedule:** `0 2 * * 0` (Sunday 2 AM)
- **Action:** Aggregate daily → weekly snapshots

### 7.3 Monthly Rollup Job
- **Schedule:** `0 3 1 * *` (1st of month 3 AM)
- **Action:** Aggregate weekly → monthly snapshots

### 7.4 Data Retention Job
- **Schedule:** `0 4 1 * *` (Monthly)
- **Action:** Purge data beyond retention limits

### 7.5 Fraud Analysis Job
- **Schedule:** `*/15 * * * *` (Every 15 minutes)
- **Action:** Scan recent activations for fraud patterns

---

## 8. User Flows

### 8.1 View Reporting Dashboard

```
1. Provider Admin logs in
2. Navigate to /platform/reporting
3. Dashboard loads with default (Daily, Today)
4. Toggle to Weekly/Monthly view
5. Click on metric card for detail view
6. Use date picker for custom range
7. Click Export to download CSV
```

### 8.2 Configure Notifications

```
1. Provider Admin navigates to /platform/settings
2. Select "Notifications" tab
3. Toggle desired alert types
4. Set percentage thresholds
5. Set absolute thresholds (optional)
6. Enable anomaly detection (optional)
7. Configure email/SMS channels
8. Save preferences
9. System sends test notification
```

### 8.3 Fraud Alert Flow

```
1. System detects suspicious activity
2. Fraud event logged to reporting_fraud_log
3. If threshold exceeded:
   - Quarantine voucher
   - Flag user account
   - Send alert to provider
4. Provider reviews in /platform/reporting/fraud
5. Provider can whitelist IP or restore voucher
```

---

## 9. Technical Considerations

### 9.1 Performance

- Use snapshot tables for fast queries (avoid aggregations on radacct)
- Implement caching (5-minute TTL for dashboard data)
- Pagination for issues and fraud logs
- Async CSV generation for large exports

### 9.2 Security

- Provider-scoped queries (WHERE provider_id = ?)
- Rate limiting on API endpoints
- Input validation on all params
- Audit logging for fraud decisions

### 9.3 Scalability

- Batch inserts for snapshot creation
- Indexed queries on provider_id + date
- Partitioning consideration for radacct table

---

## 10. Files to Create/Modify

### Backend

| File | Action |
|------|--------|
| `internal/domain/reporting.go` | New domain models |
| `internal/migration/reporting.go` | New migration |
| `internal/service/reporting.go` | Reporting service |
| `internal/service/aggregation.go` | Aggregation service |
| `internal/service/fraud.go` | Fraud detection |
| `internal/service/provider_notifications.go` | Notifications |
| `internal/adminapi/reporting.go` | API handlers |
| `internal/adminapi/reporting_routes.go` | Route registration |
| `internal/app/jobs.go` | Add scheduled jobs |
| `config/config.go` | Add config structs |

### Frontend

| File | Action |
|------|--------|
| `web/src/pages/Platform/ReportingDashboard.tsx` | New dashboard page |
| `web/src/pages/Platform/NotificationSettings.tsx` | New settings tab |
| `web/src/components/Platform/SummaryCard.tsx` | New reusable card |
| `web/src/components/Platform/NetworkStatusWidget.tsx` | New widget |
| `web/src/components/Platform/AgentFinancialWidget.tsx` | New widget |
| `web/src/components/Platform/IssuesReporterWidget.tsx` | New widget |
| `web/src/components/Platform/FraudAlertWidget.tsx` | New widget |
| `web/src/i18n/en-US.ts` | Add translations |
| `web/src/i18n/ar.ts` | Add translations |
| `web/src/i18n/zh-CN.ts` | Add translations |
| `web/src/App.tsx` | Add routes |

---

## 11. Testing Strategy

### Unit Tests
- Aggregation calculations
- Fraud detection rules
- Threshold comparisons
- Data serialization

### Integration Tests
- API endpoint responses
- Database migrations
- Scheduled job execution
- CSV export generation

### Manual Testing
- Dashboard load performance
- Notification delivery
- Fraud flagging accuracy
- Export file accuracy

---

## 12. Acceptance Criteria

1. ✅ Dashboard loads within 2 seconds with cached data
2. ✅ All metrics match source data within 1% variance
3. ✅ CSV export contains all visible data
4. ✅ Fraud detection flags at configured thresholds
5. ✅ Notifications sent within 5 minutes of threshold breach
6. ✅ Provider can only see their own data
7. ✅ Historical data viewable up to retention limits
