package persistence

import (
	"context"

	"github.com/eneskaya/insider-messaging/internal/domain/entity"
	"github.com/eneskaya/insider-messaging/internal/domain/repository"
	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/persistence/model"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type messageRepositoryGorm struct {
	db        *gorm.DB
	charLimit int
}

func NewMessageRepositoryGorm(db *gorm.DB, charLimit int) repository.MessageRepository {
	return &messageRepositoryGorm{
		db:        db,
		charLimit: charLimit,
	}
}

func (r *messageRepositoryGorm) Create(ctx context.Context, message *entity.Message) error {
	messageModel := model.ToModel(message)

	result := r.db.WithContext(ctx).Create(messageModel)
	if result.Error != nil {
		logger.Get().Error("failed to create message",
			zap.Error(result.Error),
			zap.String("message_id", message.ID().String()),
		)
		return mapGormError(result.Error)
	}

	return nil
}

func (r *messageRepositoryGorm) Update(ctx context.Context, message *entity.Message) error {
	messageModel := model.ToModel(message)

	result := r.db.WithContext(ctx).
		Model(&model.MessageModel{}).
		Where("id = ?", messageModel.ID).
		Updates(messageModel)

	if result.Error != nil {
		logger.Get().Error("failed to update message",
			zap.Error(result.Error),
			zap.String("message_id", message.ID().String()),
		)
		return mapGormError(result.Error)
	}

	if err := checkRowsAffected(result, 1); err != nil {
		return err
	}

	message.IncrementVersion()
	return nil
}

func (r *messageRepositoryGorm) FindByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	var messageModel model.MessageModel

	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&messageModel)

	if result.Error != nil {
		logger.Get().Error("failed to find message by ID",
			zap.Error(result.Error),
			zap.String("message_id", id.String()),
		)
		return nil, mapGormError(result.Error)
	}

	return model.ToEntity(&messageModel, r.charLimit)
}

func (r *messageRepositoryGorm) FindPendingMessages(ctx context.Context, limit int) ([]*entity.Message, error) {
	var models []model.MessageModel

	query := `
		SELECT * FROM messages
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`

	result := r.db.WithContext(ctx).
		Raw(query, valueobject.MessageStatusPending.String(), limit).
		Scan(&models)

	if result.Error != nil {
		logger.Get().Error("failed to find pending messages", zap.Error(result.Error))
		return nil, mapGormError(result.Error)
	}

	return model.ToEntities(models, r.charLimit)
}

func (r *messageRepositoryGorm) FindSentMessages(ctx context.Context, limit, offset int) ([]*entity.Message, error) {
	var models []model.MessageModel

	result := r.db.WithContext(ctx).
		Where("status = ?", valueobject.MessageStatusSent.String()).
		Order("sent_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models)

	if result.Error != nil {
		logger.Get().Error("failed to find sent messages", zap.Error(result.Error))
		return nil, mapGormError(result.Error)
	}

	return model.ToEntities(models, r.charLimit)
}

func (r *messageRepositoryGorm) GetStats(ctx context.Context) (*repository.MessageStats, error) {
	var stats repository.MessageStats

	type statsResult struct {
		Total   int64
		Pending int64
		Sent    int64
		Failed  int64
	}

	var result statsResult

	err := r.db.WithContext(ctx).
		Model(&model.MessageModel{}).
		Select(`
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'sent') as sent,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		`).
		Scan(&result).Error

	if err != nil {
		logger.Get().Error("failed to get message stats", zap.Error(err))
		return nil, mapGormError(err)
	}

	stats.TotalMessages = result.Total
	stats.PendingMessages = result.Pending
	stats.SentMessages = result.Sent
	stats.FailedMessages = result.Failed

	return &stats, nil
}

func (r *messageRepositoryGorm) BeginTx(ctx context.Context) (repository.Transaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, mapGormError(tx.Error)
	}

	return &gormTransaction{
		tx:  tx,
		ctx: ctx,
		db:  r.db,
	}, nil
}

type gormTransaction struct {
	tx  *gorm.DB
	ctx context.Context
	db  *gorm.DB
}

func (t *gormTransaction) Commit() error {
	err := t.tx.Commit().Error
	if err != nil {
		return mapGormError(err)
	}
	return nil
}

func (t *gormTransaction) Rollback() error {
	err := t.tx.Rollback().Error
	if err != nil {
		return mapGormError(err)
	}
	return nil
}

func (t *gormTransaction) GetContext() context.Context {
	return t.ctx
}

func (r *messageRepositoryGorm) WithTx(tx *gorm.DB) repository.MessageRepository {
	return &messageRepositoryGorm{
		db:        tx,
		charLimit: r.charLimit,
	}
}
