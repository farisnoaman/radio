package alerting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestEmailNotifier_Send_ShouldFailWithoutRecipient(t *testing.T) {
	notifier := NewEmailNotifier(EmailConfig{
		SMTPHost:     "localhost",
		SMTPPort:     25,
		SMTPUser:     "test",
		SMTPPassword: "pass",
		FromAddress:  "alerts@example.com",
	})

	alert := &domain.Alert{
		ID:         1,
		TenantID:   1,
		RuleName:   "Test Alert",
		Severity:   "warning",
		Message:    "Test message",
		Status:     "active",
		TriggeredAt: time.Now(),
		RecipientEmail: "",
	}

	err := notifier.Send(context.Background(), alert)
	if err == nil {
		t.Error("expected error for missing recipient")
	}
}

func TestWebhookNotifier_Send_ShouldSucceed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier(WebhookConfig{
		URL:     server.URL,
		Headers: map[string]string{"X-Custom": "test"},
		Timeout: 5 * time.Second,
	})

	alert := &domain.Alert{
		ID:          1,
		TenantID:    1,
		RuleName:    "Test Alert",
		Severity:    "warning",
		Message:     "Test message",
		Status:      "active",
		TriggeredAt: time.Now(),
	}

	err := notifier.Send(context.Background(), alert)
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
}

func TestWebhookNotifier_Send_ShouldFailOnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier(WebhookConfig{
		URL:     server.URL,
		Timeout: 5 * time.Second,
	})

	alert := &domain.Alert{
		ID:          1,
		TenantID:    1,
		RuleName:    "Test Alert",
		Severity:    "warning",
		Message:     "Test message",
		Status:      "active",
		TriggeredAt: time.Now(),
	}

	err := notifier.Send(context.Background(), alert)
	if err == nil {
		t.Error("expected error for failed webhook")
	}
}

func TestWebhookNotifier_Name(t *testing.T) {
	notifier := NewWebhookNotifier(WebhookConfig{URL: "http://localhost"})
	if notifier.Name() != "webhook" {
		t.Errorf("expected webhook, got %s", notifier.Name())
	}
}

func TestEmailNotifier_Name(t *testing.T) {
	notifier := NewEmailNotifier(EmailConfig{})
	if notifier.Name() != "email" {
		t.Errorf("expected email, got %s", notifier.Name())
	}
}