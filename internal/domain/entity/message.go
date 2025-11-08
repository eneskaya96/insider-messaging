package entity

import (
	"time"

	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	"github.com/google/uuid"
)

type Message struct {
	id                uuid.UUID
	phoneNumber       *valueobject.PhoneNumber
	content           *valueobject.MessageContent
	status            valueobject.MessageStatus
	createdAt         time.Time
	sentAt            *time.Time
	attempts          int
	maxAttempts       int
	lastError         string
	errorCode         string
	webhookMessageID  string
	webhookResponse   string
	version           int
}

func NewMessage(
	phoneNumber *valueobject.PhoneNumber,
	content *valueobject.MessageContent,
	maxAttempts int,
) (*Message, error) {
	return &Message{
		id:          uuid.New(),
		phoneNumber: phoneNumber,
		content:     content,
		status:      valueobject.MessageStatusPending,
		createdAt:   time.Now().UTC(),
		attempts:    0,
		maxAttempts: maxAttempts,
		version:     1,
	}, nil
}

func ReconstructMessage(
	id uuid.UUID,
	phoneNumber *valueobject.PhoneNumber,
	content *valueobject.MessageContent,
	status valueobject.MessageStatus,
	createdAt time.Time,
	sentAt *time.Time,
	attempts int,
	maxAttempts int,
	lastError string,
	errorCode string,
	webhookMessageID string,
	webhookResponse string,
	version int,
) *Message {
	return &Message{
		id:               id,
		phoneNumber:      phoneNumber,
		content:          content,
		status:           status,
		createdAt:        createdAt,
		sentAt:           sentAt,
		attempts:         attempts,
		maxAttempts:      maxAttempts,
		lastError:        lastError,
		errorCode:        errorCode,
		webhookMessageID: webhookMessageID,
		webhookResponse:  webhookResponse,
		version:          version,
	}
}

func (m *Message) ID() uuid.UUID {
	return m.id
}

func (m *Message) PhoneNumber() *valueobject.PhoneNumber {
	return m.phoneNumber
}

func (m *Message) Content() *valueobject.MessageContent {
	return m.content
}

func (m *Message) Status() valueobject.MessageStatus {
	return m.status
}

func (m *Message) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Message) SentAt() *time.Time {
	return m.sentAt
}

func (m *Message) Attempts() int {
	return m.attempts
}

func (m *Message) MaxAttempts() int {
	return m.maxAttempts
}

func (m *Message) LastError() string {
	return m.lastError
}

func (m *Message) ErrorCode() string {
	return m.errorCode
}

func (m *Message) WebhookMessageID() string {
	return m.webhookMessageID
}

func (m *Message) WebhookResponse() string {
	return m.webhookResponse
}

func (m *Message) Version() int {
	return m.version
}

func (m *Message) MarkAsProcessing() {
	m.status = valueobject.MessageStatusProcessing
	m.attempts++
}

func (m *Message) MarkAsSent(webhookMessageID, webhookResponse string) {
	m.status = valueobject.MessageStatusSent
	now := time.Now().UTC()
	m.sentAt = &now
	m.webhookMessageID = webhookMessageID
	m.webhookResponse = webhookResponse
	m.lastError = ""
	m.errorCode = ""
}

func (m *Message) MarkAsFailed(errorMsg, errorCode string) {
	m.lastError = errorMsg
	m.errorCode = errorCode

	if m.attempts >= m.maxAttempts {
		m.status = valueobject.MessageStatusFailed
	} else {
		m.status = valueobject.MessageStatusPending
	}
}

func (m *Message) CanRetry() bool {
	return m.attempts < m.maxAttempts && !m.status.IsSent()
}

func (m *Message) IncrementVersion() {
	m.version++
}
