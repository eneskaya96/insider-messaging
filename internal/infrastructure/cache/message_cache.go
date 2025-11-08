package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eneskaya/insider-messaging/pkg/logger"
	"go.uber.org/zap"
)

type CachedMessage struct {
	MessageID        string    `json:"message_id"`
	WebhookMessageID string    `json:"webhook_message_id"`
	SentAt           time.Time `json:"sent_at"`
	PhoneNumber      string    `json:"phone_number"`
}

type MessageCache interface {
	CacheSentMessage(ctx context.Context, msg *CachedMessage) error
	GetSentMessage(ctx context.Context, messageID string) (*CachedMessage, error)
	IsCached(ctx context.Context, messageID string) (bool, error)
}

type messageCache struct {
	redis *RedisCache
}

func NewMessageCache(redis *RedisCache) MessageCache {
	return &messageCache{
		redis: redis,
	}
}

func (c *messageCache) CacheSentMessage(ctx context.Context, msg *CachedMessage) error {
	key := c.buildKey(msg.MessageID)

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Get().Error("failed to marshal cached message",
			zap.Error(err),
			zap.String("message_id", msg.MessageID),
		)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := c.redis.Set(ctx, key, data); err != nil {
		logger.Get().Error("failed to cache sent message",
			zap.Error(err),
			zap.String("message_id", msg.MessageID),
		)
		return fmt.Errorf("failed to cache message: %w", err)
	}

	logger.Get().Debug("cached sent message",
		zap.String("message_id", msg.MessageID),
		zap.String("webhook_message_id", msg.WebhookMessageID),
	)

	return nil
}

func (c *messageCache) GetSentMessage(ctx context.Context, messageID string) (*CachedMessage, error) {
	key := c.buildKey(messageID)

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("message not found in cache: %w", err)
	}

	var msg CachedMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached message: %w", err)
	}

	return &msg, nil
}

func (c *messageCache) IsCached(ctx context.Context, messageID string) (bool, error) {
	key := c.buildKey(messageID)
	return c.redis.Exists(ctx, key)
}

func (c *messageCache) buildKey(messageID string) string {
	return fmt.Sprintf("message:sent:%s", messageID)
}
