package valueobject

import "fmt"

type MessageStatus string

const (
	MessageStatusPending    MessageStatus = "pending"
	MessageStatusProcessing MessageStatus = "processing"
	MessageStatusSent       MessageStatus = "sent"
	MessageStatusFailed     MessageStatus = "failed"
)

func NewMessageStatus(status string) (MessageStatus, error) {
	ms := MessageStatus(status)
	switch ms {
	case MessageStatusPending, MessageStatusProcessing, MessageStatusSent, MessageStatusFailed:
		return ms, nil
	default:
		return "", fmt.Errorf("invalid message status: %s", status)
	}
}

func (s MessageStatus) String() string {
	return string(s)
}

func (s MessageStatus) IsPending() bool {
	return s == MessageStatusPending
}

func (s MessageStatus) IsProcessing() bool {
	return s == MessageStatusProcessing
}

func (s MessageStatus) IsSent() bool {
	return s == MessageStatusSent
}

func (s MessageStatus) IsFailed() bool {
	return s == MessageStatusFailed
}

func (s MessageStatus) CanProcess() bool {
	return s == MessageStatusPending
}
