package entity

import (
	"testing"

	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test message", 160)

	message, err := NewMessage(phone, content, 3)

	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, valueobject.MessageStatusPending, message.Status())
	assert.Equal(t, 0, message.Attempts())
	assert.Equal(t, 3, message.MaxAttempts())
}

func TestMessageMarkAsProcessing(t *testing.T) {
	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test message", 160)
	message, _ := NewMessage(phone, content, 3)

	message.MarkAsProcessing()

	assert.Equal(t, valueobject.MessageStatusProcessing, message.Status())
	assert.Equal(t, 1, message.Attempts())
}

func TestMessageMarkAsSent(t *testing.T) {
	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test message", 160)
	message, _ := NewMessage(phone, content, 3)

	webhookID := "webhook-123"
	response := `{"message": "sent"}`

	message.MarkAsSent(webhookID, response)

	assert.Equal(t, valueobject.MessageStatusSent, message.Status())
	assert.Equal(t, webhookID, message.WebhookMessageID())
	assert.Equal(t, response, message.WebhookResponse())
	assert.NotNil(t, message.SentAt())
	assert.Empty(t, message.LastError())
}

func TestMessageMarkAsFailed(t *testing.T) {
	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test message", 160)
	message, _ := NewMessage(phone, content, 3)

	message.MarkAsProcessing()
	message.MarkAsFailed("timeout error", "TIMEOUT")

	assert.Equal(t, valueobject.MessageStatusPending, message.Status())
	assert.Equal(t, "timeout error", message.LastError())
	assert.Equal(t, "TIMEOUT", message.ErrorCode())

	message.MarkAsProcessing()
	message.MarkAsFailed("error 2", "ERROR")
	message.MarkAsProcessing()
	message.MarkAsFailed("error 3", "ERROR")

	assert.Equal(t, valueobject.MessageStatusFailed, message.Status())
	assert.Equal(t, 3, message.Attempts())
}

func TestMessageCanRetry(t *testing.T) {
	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test message", 160)
	message, _ := NewMessage(phone, content, 3)

	assert.True(t, message.CanRetry())

	message.MarkAsProcessing()
	message.MarkAsProcessing()
	message.MarkAsProcessing()

	assert.False(t, message.CanRetry())

	message.MarkAsSent("webhook-123", "{}")
	assert.False(t, message.CanRetry())
}
