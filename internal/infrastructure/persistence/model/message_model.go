package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/plugin/optimisticlock"
)

type MessageModel struct {
	ID               uuid.UUID                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PhoneNumber      string                    `gorm:"column:phone_number;type:varchar(20);not null;index:idx_messages_phone"`
	Content          string                    `gorm:"type:text;not null"`
	Status           string                    `gorm:"type:varchar(20);not null;default:'pending';index:idx_messages_status;index:idx_messages_status_created_at,priority:1"`
	CreatedAt        time.Time                 `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_messages_created_at;index:idx_messages_status_created_at,priority:2;index:idx_messages_pending_fifo,where:status = 'pending'"`
	SentAt           *time.Time                `gorm:"index:idx_messages_sent_at,where:sent_at IS NOT NULL"`
	Attempts         int                       `gorm:"not null;default:0"`
	MaxAttempts      int                       `gorm:"not null;default:3"`
	LastError        string                    `gorm:"type:text"`
	ErrorCode        string                    `gorm:"type:varchar(50)"`
	WebhookMessageID string                    `gorm:"column:webhook_message_id;type:varchar(255)"`
	WebhookResponse  string                    `gorm:"type:text"`
	Version          optimisticlock.Version    `gorm:"column:version;not null;default:0"`
}

func (MessageModel) TableName() string {
	return "messages"
}
