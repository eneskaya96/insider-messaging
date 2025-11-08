package repository

import (
	"context"

	"github.com/eneskaya/insider-messaging/internal/domain/entity"
	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, message *entity.Message) error
	Update(ctx context.Context, message *entity.Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Message, error)
	FindPendingMessages(ctx context.Context, limit int) ([]*entity.Message, error)
	FindSentMessages(ctx context.Context, limit, offset int) ([]*entity.Message, error)
	GetStats(ctx context.Context) (*MessageStats, error)
	BeginTx(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Commit() error
	Rollback() error
	GetContext() context.Context
}

type MessageStats struct {
	TotalMessages   int64
	PendingMessages int64
	SentMessages    int64
	FailedMessages  int64
}
