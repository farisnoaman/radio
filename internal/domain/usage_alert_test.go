package domain

import (
	"testing"
	"time"
)

func TestUsageAlert_CanSendAlert(t *testing.T) {
	alert := &UsageAlert{
		UserID:    1,
		Threshold: 80,
		AlertType: "email",
		SentAt:    nil,
		CreatedAt: time.Now(),
	}

	if !alert.CanSendAlert() {
		t.Error("Expected alert to be sendable when SentAt is nil")
	}
}

func TestUsageAlert_CannotSendDuplicateAlert(t *testing.T) {
	now := time.Now()
	alert := &UsageAlert{
		UserID:    1,
		Threshold: 80,
		AlertType: "email",
		SentAt:    &now,
		CreatedAt: time.Now(),
	}

	if alert.CanSendAlert() {
		t.Error("Expected alert to not be sendable when already sent")
	}
}
