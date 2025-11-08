package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eneskaya/insider-messaging/internal/domain/entity"
	"github.com/eneskaya/insider-messaging/internal/domain/repository"
	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	apperrors "github.com/eneskaya/insider-messaging/pkg/errors"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type messageRepositoryPostgres struct {
	db        *sql.DB
	charLimit int
}

func NewMessageRepositoryPostgres(db *sql.DB, charLimit int) repository.MessageRepository {
	return &messageRepositoryPostgres{
		db:        db,
		charLimit: charLimit,
	}
}

func (r *messageRepositoryPostgres) Create(ctx context.Context, message *entity.Message) error {
	query := `
		INSERT INTO messages (
			id, phone_number, content, status, created_at,
			attempts, max_attempts, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		message.ID(),
		message.PhoneNumber().String(),
		message.Content().String(),
		message.Status().String(),
		message.CreatedAt(),
		message.Attempts(),
		message.MaxAttempts(),
		message.Version(),
	)

	if err != nil {
		logger.Get().Error("failed to create message",
			zap.Error(err),
			zap.String("message_id", message.ID().String()),
		)
		return apperrors.NewDatabaseError(err)
	}

	return nil
}

func (r *messageRepositoryPostgres) Update(ctx context.Context, message *entity.Message) error {
	query := `
		UPDATE messages SET
			status = $1,
			sent_at = $2,
			attempts = $3,
			last_error = $4,
			error_code = $5,
			webhook_message_id = $6,
			webhook_response = $7,
			version = $8
		WHERE id = $9 AND version = $10
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		message.Status().String(),
		message.SentAt(),
		message.Attempts(),
		message.LastError(),
		message.ErrorCode(),
		message.WebhookMessageID(),
		message.WebhookResponse(),
		message.Version()+1,
		message.ID(),
		message.Version(),
	)

	if err != nil {
		logger.Get().Error("failed to update message",
			zap.Error(err),
			zap.String("message_id", message.ID().String()),
		)
		return apperrors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return apperrors.New(apperrors.ErrorCodeNotFound, "message not found or version mismatch (optimistic lock)")
	}

	message.IncrementVersion()
	return nil
}

func (r *messageRepositoryPostgres) FindByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	query := `
		SELECT
			id, phone_number, content, status, created_at, sent_at,
			attempts, max_attempts, last_error, error_code,
			webhook_message_id, webhook_response, version
		FROM messages
		WHERE id = $1
	`

	var (
		msgID            uuid.UUID
		phoneNumber      string
		content          string
		status           string
		createdAt        time.Time
		sentAt           sql.NullTime
		attempts         int
		maxAttempts      int
		lastError        sql.NullString
		errorCode        sql.NullString
		webhookMessageID sql.NullString
		webhookResponse  sql.NullString
		version          int
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msgID, &phoneNumber, &content, &status, &createdAt, &sentAt,
		&attempts, &maxAttempts, &lastError, &errorCode,
		&webhookMessageID, &webhookResponse, &version,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NewNotFoundError("message not found")
	}
	if err != nil {
		logger.Get().Error("failed to find message by ID",
			zap.Error(err),
			zap.String("message_id", id.String()),
		)
		return nil, apperrors.NewDatabaseError(err)
	}

	return r.scanMessage(
		msgID, phoneNumber, content, status, createdAt, sentAt,
		attempts, maxAttempts, lastError, errorCode,
		webhookMessageID, webhookResponse, version,
	)
}

func (r *messageRepositoryPostgres) FindPendingMessages(ctx context.Context, limit int) ([]*entity.Message, error) {
	query := `
		SELECT
			id, phone_number, content, status, created_at, sent_at,
			attempts, max_attempts, last_error, error_code,
			webhook_message_id, webhook_response, version
		FROM messages
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.QueryContext(ctx, query, valueobject.MessageStatusPending.String(), limit)
	if err != nil {
		logger.Get().Error("failed to find pending messages", zap.Error(err))
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

func (r *messageRepositoryPostgres) FindSentMessages(ctx context.Context, limit, offset int) ([]*entity.Message, error) {
	query := `
		SELECT
			id, phone_number, content, status, created_at, sent_at,
			attempts, max_attempts, last_error, error_code,
			webhook_message_id, webhook_response, version
		FROM messages
		WHERE status = $1
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, valueobject.MessageStatusSent.String(), limit, offset)
	if err != nil {
		logger.Get().Error("failed to find sent messages", zap.Error(err))
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

func (r *messageRepositoryPostgres) GetStats(ctx context.Context) (*repository.MessageStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'sent') as sent,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM messages
	`

	var stats repository.MessageStats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalMessages,
		&stats.PendingMessages,
		&stats.SentMessages,
		&stats.FailedMessages,
	)

	if err != nil {
		logger.Get().Error("failed to get message stats", zap.Error(err))
		return nil, apperrors.NewDatabaseError(err)
	}

	return &stats, nil
}

func (r *messageRepositoryPostgres) BeginTx(ctx context.Context) (repository.Transaction, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	return &postgresTransaction{tx: tx, ctx: ctx}, nil
}

func (r *messageRepositoryPostgres) scanMessages(rows *sql.Rows) ([]*entity.Message, error) {
	messages := make([]*entity.Message, 0)

	for rows.Next() {
		var (
			msgID            uuid.UUID
			phoneNumber      string
			content          string
			status           string
			createdAt        time.Time
			sentAt           sql.NullTime
			attempts         int
			maxAttempts      int
			lastError        sql.NullString
			errorCode        sql.NullString
			webhookMessageID sql.NullString
			webhookResponse  sql.NullString
			version          int
		)

		err := rows.Scan(
			&msgID, &phoneNumber, &content, &status, &createdAt, &sentAt,
			&attempts, &maxAttempts, &lastError, &errorCode,
			&webhookMessageID, &webhookResponse, &version,
		)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}

		message, err := r.scanMessage(
			msgID, phoneNumber, content, status, createdAt, sentAt,
			attempts, maxAttempts, lastError, errorCode,
			webhookMessageID, webhookResponse, version,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	return messages, nil
}

func (r *messageRepositoryPostgres) scanMessage(
	msgID uuid.UUID,
	phoneNumber string,
	content string,
	status string,
	createdAt time.Time,
	sentAt sql.NullTime,
	attempts int,
	maxAttempts int,
	lastError sql.NullString,
	errorCode sql.NullString,
	webhookMessageID sql.NullString,
	webhookResponse sql.NullString,
	version int,
) (*entity.Message, error) {
	phone, err := valueobject.NewPhoneNumber(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number in database: %w", err)
	}

	messageContent, err := valueobject.NewMessageContent(content, r.charLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid message content in database: %w", err)
	}

	messageStatus, err := valueobject.NewMessageStatus(status)
	if err != nil {
		return nil, fmt.Errorf("invalid message status in database: %w", err)
	}

	var sentAtPtr *time.Time
	if sentAt.Valid {
		sentAtPtr = &sentAt.Time
	}

	return entity.ReconstructMessage(
		msgID,
		phone,
		messageContent,
		messageStatus,
		createdAt,
		sentAtPtr,
		attempts,
		maxAttempts,
		lastError.String,
		errorCode.String,
		webhookMessageID.String,
		webhookResponse.String,
		version,
	), nil
}

type postgresTransaction struct {
	tx  *sql.Tx
	ctx context.Context
}

func (t *postgresTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *postgresTransaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *postgresTransaction) GetContext() context.Context {
	return t.ctx
}
