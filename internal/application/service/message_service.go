package service

import (
	"context"
	"fmt"

	"github.com/eneskaya/insider-messaging/internal/application/dto"
	"github.com/eneskaya/insider-messaging/internal/domain/entity"
	"github.com/eneskaya/insider-messaging/internal/domain/repository"
	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/cache"
	infrahttp "github.com/eneskaya/insider-messaging/internal/infrastructure/http"
	apperrors "github.com/eneskaya/insider-messaging/pkg/errors"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MessageService interface {
	CreateMessage(ctx context.Context, req *dto.CreateMessageRequest) (*dto.MessageResponse, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*dto.MessageResponse, error)
	GetSentMessages(ctx context.Context, page, pageSize int) (*dto.MessageListResponse, error)
	GetStats(ctx context.Context) (*dto.MessageStatsResponse, error)
	ProcessPendingMessages(ctx context.Context, batchSize int) (int, error)
}

type messageService struct {
	repo          repository.MessageRepository
	webhookClient infrahttp.WebhookClient
	messageCache  cache.MessageCache
	charLimit     int
	maxRetries    int
}

func NewMessageService(
	repo repository.MessageRepository,
	webhookClient infrahttp.WebhookClient,
	messageCache cache.MessageCache,
	charLimit int,
	maxRetries int,
) MessageService {
	return &messageService{
		repo:          repo,
		webhookClient: webhookClient,
		messageCache:  messageCache,
		charLimit:     charLimit,
		maxRetries:    maxRetries,
	}
}

func (s *messageService) CreateMessage(ctx context.Context, req *dto.CreateMessageRequest) (*dto.MessageResponse, error) {
	phoneNumber, err := valueobject.NewPhoneNumber(req.PhoneNumber)
	if err != nil {
		return nil, apperrors.NewValidationError(err.Error())
	}

	content, err := valueobject.NewMessageContent(req.Content, s.charLimit)
	if err != nil {
		return nil, apperrors.NewValidationError(err.Error())
	}

	message, err := entity.NewMessage(phoneNumber, content, s.maxRetries)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	if err := s.repo.Create(ctx, message); err != nil {
		return nil, err
	}

	logger.Get().Info("message created successfully",
		zap.String("message_id", message.ID().String()),
		zap.String("phone_number", phoneNumber.String()),
	)

	return s.toDTO(message), nil
}

func (s *messageService) GetMessage(ctx context.Context, id uuid.UUID) (*dto.MessageResponse, error) {
	message, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toDTO(message), nil
}

func (s *messageService) GetSentMessages(ctx context.Context, page, pageSize int) (*dto.MessageListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	messages, err := s.repo.FindSentMessages(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	responseMsgs := make([]dto.MessageResponse, len(messages))
	for i, msg := range messages {
		responseMsgs[i] = *s.toDTO(msg)
	}

	return &dto.MessageListResponse{
		Messages:   responseMsgs,
		TotalCount: int(stats.SentMessages),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *messageService) GetStats(ctx context.Context) (*dto.MessageStatsResponse, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.MessageStatsResponse{
		TotalMessages:   stats.TotalMessages,
		PendingMessages: stats.PendingMessages,
		SentMessages:    stats.SentMessages,
		FailedMessages:  stats.FailedMessages,
	}, nil
}

func (s *messageService) ProcessPendingMessages(ctx context.Context, batchSize int) (int, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	messages, err := s.repo.FindPendingMessages(tx.GetContext(), batchSize)
	if err != nil {
		return 0, err
	}

	if len(messages) == 0 {
		return 0, nil
	}

	logger.Get().Info("processing pending messages",
		zap.Int("count", len(messages)),
		zap.Int("batch_size", batchSize),
	)

	successCount := 0
	for _, message := range messages {
		if err := s.processSingleMessage(tx.GetContext(), message); err != nil {
			logger.Get().Error("failed to process message",
				zap.Error(err),
				zap.String("message_id", message.ID().String()),
			)
			continue
		}
		successCount++
	}

	if err := tx.Commit(); err != nil {
		logger.Get().Error("failed to commit transaction", zap.Error(err))
		return 0, apperrors.NewDatabaseError(err)
	}

	logger.Get().Info("batch processing completed",
		zap.Int("total", len(messages)),
		zap.Int("successful", successCount),
		zap.Int("failed", len(messages)-successCount),
	)

	return successCount, nil
}

func (s *messageService) processSingleMessage(ctx context.Context, message *entity.Message) error {
	message.MarkAsProcessing()

	if err := s.repo.Update(ctx, message); err != nil {
		return err
	}

	webhookResp, err := s.webhookClient.SendMessage(
		ctx,
		message.PhoneNumber().String(),
		message.Content().String(),
	)

	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		errorCode := string(apperrors.ErrorCodeInternal)
		if ok {
			errorCode = string(appErr.Code)
		}

		message.MarkAsFailed(err.Error(), errorCode)
		if updateErr := s.repo.Update(ctx, message); updateErr != nil {
			logger.Get().Error("failed to update message after webhook failure",
				zap.Error(updateErr),
				zap.String("message_id", message.ID().String()),
			)
		}

		return fmt.Errorf("webhook send failed: %w", err)
	}

	responseJSON := fmt.Sprintf(`{"message": "%s", "messageId": "%s"}`, webhookResp.Message, webhookResp.MessageID)
	message.MarkAsSent(webhookResp.MessageID, responseJSON)

	if err := s.repo.Update(ctx, message); err != nil {
		return err
	}

	cachedMsg := &cache.CachedMessage{
		MessageID:        message.ID().String(),
		WebhookMessageID: webhookResp.MessageID,
		SentAt:           *message.SentAt(),
		PhoneNumber:      message.PhoneNumber().String(),
	}

	if err := s.messageCache.CacheSentMessage(ctx, cachedMsg); err != nil {
		logger.Get().Warn("failed to cache sent message (non-critical)",
			zap.Error(err),
			zap.String("message_id", message.ID().String()),
		)
	}

	logger.Get().Info("message sent successfully",
		zap.String("message_id", message.ID().String()),
		zap.String("webhook_message_id", webhookResp.MessageID),
	)

	return nil
}

func (s *messageService) toDTO(message *entity.Message) *dto.MessageResponse {
	return &dto.MessageResponse{
		ID:               message.ID().String(),
		PhoneNumber:      message.PhoneNumber().String(),
		Content:          message.Content().String(),
		Status:           message.Status().String(),
		CreatedAt:        message.CreatedAt(),
		SentAt:           message.SentAt(),
		Attempts:         message.Attempts(),
		MaxAttempts:      message.MaxAttempts(),
		LastError:        message.LastError(),
		ErrorCode:        message.ErrorCode(),
		WebhookMessageID: message.WebhookMessageID(),
	}
}
