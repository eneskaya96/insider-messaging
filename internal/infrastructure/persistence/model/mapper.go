package model

import (
	"fmt"

	"github.com/eneskaya/insider-messaging/internal/domain/entity"
	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	"gorm.io/plugin/optimisticlock"
)

func ToEntity(model *MessageModel, charLimit int) (*entity.Message, error) {
	phoneNumber, err := valueobject.NewPhoneNumber(model.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number in database: %w", err)
	}

	content, err := valueobject.NewMessageContent(model.Content, charLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid message content in database: %w", err)
	}

	status, err := valueobject.NewMessageStatus(model.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid message status in database: %w", err)
	}

	return entity.ReconstructMessage(
		model.ID,
		phoneNumber,
		content,
		status,
		model.CreatedAt,
		model.SentAt,
		model.Attempts,
		model.MaxAttempts,
		model.LastError,
		model.ErrorCode,
		model.WebhookMessageID,
		model.WebhookResponse,
		int(model.Version.Int64),
	), nil
}

func ToEntities(models []MessageModel, charLimit int) ([]*entity.Message, error) {
	entities := make([]*entity.Message, 0, len(models))

	for _, model := range models {
		entity, err := ToEntity(&model, charLimit)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

func ToModel(entity *entity.Message) *MessageModel {
	return &MessageModel{
		ID:               entity.ID(),
		PhoneNumber:      entity.PhoneNumber().String(),
		Content:          entity.Content().String(),
		Status:           entity.Status().String(),
		CreatedAt:        entity.CreatedAt(),
		SentAt:           entity.SentAt(),
		Attempts:         entity.Attempts(),
		MaxAttempts:      entity.MaxAttempts(),
		LastError:        entity.LastError(),
		ErrorCode:        entity.ErrorCode(),
		WebhookMessageID: entity.WebhookMessageID(),
		WebhookResponse:  entity.WebhookResponse(),
		Version:          optimisticlock.Version{Int64: int64(entity.Version())},
	}
}

func UpdateModelFromEntity(model *MessageModel, entity *entity.Message) {
	model.Status = entity.Status().String()
	model.SentAt = entity.SentAt()
	model.Attempts = entity.Attempts()
	model.LastError = entity.LastError()
	model.ErrorCode = entity.ErrorCode()
	model.WebhookMessageID = entity.WebhookMessageID()
	model.WebhookResponse = entity.WebhookResponse()
	model.Version = optimisticlock.Version{Int64: int64(entity.Version())}
}
