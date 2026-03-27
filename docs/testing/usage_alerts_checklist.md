# Usage Alert Notifications - Testing Checklist

## Prerequisites

1. Configure SMTP settings in `toughradius.yml`:
   ```yaml
   notification:
     enabled: true
     smtp_host: smtp.example.com
     smtp_port: 587
     smtp_username: your-email@example.com
     smtp_password: your-app-password
     smtp_from: noreply@example.com
   ```

2. Run database migrations to create tables:
   ```bash
   go run cmd/migrate/main.go
   ```

3. Verify tables exist:
   ```sql
   \d usage_alerts
   \d notification_preferences
   ```

## Test Cases

### 1. Email Alert at 80% Threshold

- [ ] Create user with 10 GB data quota
- [ ] Set user's current month usage to 8.1 GB (81%)
- [ ] Verify user has default notification preferences
- [ ] Trigger alert check job manually or wait for scheduled run
- [ ] Verify email sent to user's email address
- [ ] Check `usage_alerts` table for alert record
- [ ] Verify `sent_at` timestamp is set

### 2. No Duplicate Alerts Within 24 Hours

- [ ] Run alert check job immediately after Test 1
- [ ] Verify no duplicate email sent
- [ ] Query `usage_alerts` table - should only have 1 record for 80% threshold
- [ ] Wait 24 hours or adjust system time
- [ ] Run job again
- [ ] Verify second email sent

### 3. Multiple Thresholds Triggered

- [ ] Set user preferences to thresholds: 80%, 90%, 100%
- [ ] Set user usage to 9.5 GB (95%)
- [ ] Clear any existing alert records
- [ ] Run alert check job
- [ ] Verify 80% and 90% alerts sent (not 100%)
- [ ] Check usage_alerts table for 2 records

### 4. SMS Alert at 100% Threshold

- [ ] Enable SMS for user with Twilio credentials configured
- [ ] Set SMS threshold to 100%
- [ ] Set user usage to exceed quota (e.g., 10.1 GB)
- [ ] Run alert check job
- [ ] Verify SMS sent via Twilio (check Twilio logs/dashboard)

### 5. Disabled Notifications

- [ ] Disable email alerts for user
- [ ] Set user usage to trigger 80% threshold
- [ ] Run alert check job
- [ ] Verify NO email sent
- [ ] Verify NO record in usage_alerts table

### 6. Portal - View Notification Preferences

- [ ] Login to portal as regular user
- [ ] Navigate to `/portal/preferences/notifications`
- [ ] Verify current preferences displayed
- [ ] Verify default values if no preferences set

### 7. Portal - Update Notification Preferences

- [ ] Toggle email alerts OFF
- [ ] Change email thresholds to "70,85,100"
- [ ] Toggle SMS alerts ON
- [ ] Set SMS threshold to "100"
- [ ] Click Save
- [ ] Verify success notification
- [ ] Refresh page - verify settings persisted
- [ ] Verify via API response

### 8. Portal - Alert History

- [ ] Navigate to `/portal/alerts/history`
- [ ] Verify past alerts displayed (if any)
- [ ] Check timestamp format is readable
- [ ] Verify threshold percentage shown correctly
- [ ] Verify alert type (email/SMS) shown

### 9. Scheduled Job Execution

- [ ] Check cron schedule (default: every 6 hours)
- [ ] Review application logs for "Starting usage alert check"
- [ ] Verify log shows number of alerts sent
- [ ] Verify no errors in logs

### 10. Database Records

- [ ] Query `notification_preferences` table
- [ ] Verify user_id foreign key constraint works
- [ ] Query `usage_alerts` table
- [ ] Verify proper indexing on user_id, threshold, sent_at
- [ ] Test CASCADE delete when user deleted

## API Endpoint Tests

### GET /api/v1/portal/preferences/notifications

```bash
curl -X GET http://localhost:1816/api/v1/portal/preferences/notifications \
  -H "Authorization: Bearer USER_TOKEN"
```

Expected: JSON with notification preferences

### PUT /api/v1/portal/preferences/notifications

```bash
curl -X PUT http://localhost:1816/api/v1/portal/preferences/notifications \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email_enabled": true,
    "email_thresholds": "80,90,100",
    "sms_enabled": false,
    "sms_thresholds": "100",
    "daily_summary_enabled": false
  }'
```

Expected: Updated preferences JSON

### GET /api/v1/portal/alerts/history

```bash
curl -X GET http://localhost:1816/api/v1/portal/alerts/history \
  -H "Authorization: Bearer USER_TOKEN"
```

Expected: JSON array of alert records

## Troubleshooting

### Emails not sending
1. Check SMTP credentials in config
2. Verify SMTP port (587 for TLS, 465 for SSL)
3. Check application logs: `tail -f toughradius.log | grep -i "alert"`
4. Test SMTP connection: `telnet smtp.example.com 587`

### Job not running
1. Verify `notification.enabled: true` in config
2. Check logs for "init usage alert job error"
3. Manual trigger: not available via API yet (future enhancement)

### Alerts not appearing in portal
1. Verify user_id matches logged-in user
2. Check database has records in usage_alerts table
3. Check API response in browser DevTools
