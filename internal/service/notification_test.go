package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmailProvider struct {
	mock.Mock
}

func (m *MockEmailProvider) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

type MockSMSProvider struct {
	mock.Mock
}

func (m *MockSMSProvider) SendSMS(to, message string) error {
	args := m.Called(to, message)
	return args.Error(0)
}

func TestNotificationService_SendUsageAlert(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockEmail.On("SendEmail",
		"user@example.com",
		"Usage Alert: 80% Data Quota Used",
		mock.AnythingOfType("string"),
	).Return(nil)

	service := &NotificationService{
		emailProvider: mockEmail,
		smsProvider:  nil,
	}

	data := &NotificationData{
		Email:     "user@example.com",
		Phone:     "+1234567890",
		Username:  "testuser",
		Threshold: 80,
		UsedGB:    8.0,
		QuotaGB:   10.0,
	}

	err := service.SendUsageAlertEmail(data)

	assert.NoError(t, err)
	mockEmail.AssertExpectations(t)
}

func TestNotificationService_SendUsageAlertSMS(t *testing.T) {
	mockSMS := new(MockSMSProvider)
	mockSMS.On("SendSMS",
		"+1234567890",
		mock.AnythingOfType("string"),
	).Return(nil)

	service := &NotificationService{
		emailProvider: nil,
		smsProvider:   mockSMS,
	}

	data := &NotificationData{
		Email:     "user@example.com",
		Phone:     "+1234567890",
		Username:  "testuser",
		Threshold: 100,
		UsedGB:    10.0,
		QuotaGB:   10.0,
	}

	err := service.SendUsageAlertSMS(data)

	assert.NoError(t, err)
	mockSMS.AssertExpectations(t)
}

func TestNotificationData_CalculatePercent(t *testing.T) {
	data := &NotificationData{
		UsedGB:  8.5,
		QuotaGB: 10.0,
	}

	expectedPercent := 85.0
	assert.InDelta(t, expectedPercent, data.UsedPercent(), 0.1)
}
