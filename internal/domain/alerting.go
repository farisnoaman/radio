package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// AlertRule defines a threshold-based alert rule.
type AlertRule struct {
	ID                    int64      `json:"id,string" gorm:"primaryKey"`
	TenantID              int64      `json:"tenant_id" gorm:"index"`
	Name                  string     `json:"name" gorm:"size:200;not null"`
	MetricName            string     `json:"metric_name" gorm:"not null;size:100"` // e.g., "device.cpu", "auth.failure_rate"
	Operator              string     `json:"operator" gorm:"not null;size:10"`       // >, <, >=, <=, ==
	Threshold             float64    `json:"threshold" gorm:"not null"`
	Duration              int        `json:"duration" gorm:"not null"`              // Seconds
	Severity              string     `json:"severity" gorm:"not null;size:20"`      // info, warning, critical
	Enabled               bool       `json:"enabled" gorm:"default:true"`
	NotificationChannels  []string   `json:"notification_channels" gorm:"serializer:json"` // email, webhook, sms
	MessageTemplate       string     `json:"message_template" gorm:"type:text"`
	LastTriggered          *time.Time `json:"last_triggered"`
	TriggerCount           int        `json:"trigger_count" gorm:"default:0"`
	CooldownSec            int        `json:"cooldown_sec" gorm:"default:300"` // Minimum time between alerts
	Remark                 string     `json:"remark" gorm:"size:500"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// TableName specifies the table name.
func (AlertRule) TableName() string {
	return "alert_rule"
}

// Validate checks if the alert rule is valid.
func (r *AlertRule) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if r.MetricName == "" {
		return errors.New("metric name is required")
	}

	validOperators := []string{">", "<", ">=", "<=", "=="}
	opValid := false
	for _, op := range validOperators {
		if r.Operator == op {
			opValid = true
			break
		}
	}
	if !opValid {
		return fmt.Errorf("invalid operator: %s (must be one of: %s)", r.Operator, strings.Join(validOperators, ", "))
	}

	validSeverities := []string{"info", "warning", "critical"}
	severityValid := false
	for _, sev := range validSeverities {
		if r.Severity == sev {
			severityValid = true
			break
		}
	}
	if !severityValid {
		return fmt.Errorf("invalid severity: %s (must be one of: %s)", r.Severity, strings.Join(validSeverities, ", "))
	}

	if r.Duration < 0 {
		return errors.New("duration cannot be negative")
	}

	if r.CooldownSec < 0 {
		return errors.New("cooldown cannot be negative")
	}

	return nil
}

// Alert represents a triggered alert instance.
type Alert struct {
	ID              int64      `json:"id,string" gorm:"primaryKey"`
	TenantID        int64      `json:"tenant_id" gorm:"index"`
	RuleID          int64      `json:"rule_id" gorm:"index"`
	RuleName        string     `json:"rule_name" gorm:"size:200;not null"`
	Severity        string     `json:"severity" gorm:"not null;size:20"`
	Message         string     `json:"message" gorm:"type:text;not null"`
	Status          string     `json:"status" gorm:"size:20;default:'active'"` // active, acknowledged, resolved
	TriggeredAt     time.Time  `json:"triggered_at" gorm:"not null;index"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at"`
	AcknowledgedBy  string     `json:"acknowledged_by" gorm:"size:100"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	ResolvedBy      string     `json:"resolved_by" gorm:"size:100"`
	RecipientEmail  string     `json:"recipient_email" gorm:"size:200"`
	NotificationSent bool      `json:"notification_sent" gorm:"default:false"`
	CreatedAt       time.Time  `json:"created_at"`
}

// TableName specifies the table name.
func (Alert) TableName() string {
	return "alert"
}

// Validate checks if the alert is valid.
func (a *Alert) Validate() error {
	if a.RuleName == "" {
		return errors.New("rule name is required")
	}

	if a.Message == "" {
		return errors.New("message is required")
	}

	validStatuses := []string{"active", "acknowledged", "resolved"}
	statusValid := false
	for _, status := range validStatuses {
		if a.Status == status {
			statusValid = true
			break
		}
	}
	if !statusValid {
		return fmt.Errorf("invalid status: %s (must be one of: %s)", a.Status, strings.Join(validStatuses, ", "))
	}

	if a.TriggeredAt.IsZero() {
		return errors.New("triggered at is required")
	}

	return nil
}

// UsageAlert represents an alert sent to a user about their usage.
// Note: This is already defined in usage_alert.go, this is a placeholder for reference.
// type UsageAlert struct { ... }

// NotificationPreference represents a user's notification preferences.
// Note: This is already defined in notification_preference.go, this is a placeholder for reference.
// type NotificationPreference struct { ... }
