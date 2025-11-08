package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eneskaya/insider-messaging/pkg/config"
	apperrors "github.com/eneskaya/insider-messaging/pkg/errors"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type WebhookRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type WebhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

type WebhookClient interface {
	SendMessage(ctx context.Context, phoneNumber, content string) (*WebhookResponse, error)
}

type webhookClient struct {
	client      *http.Client
	url         string
	authKey     string
	rateLimiter *rate.Limiter
}

func NewWebhookClient(cfg *config.WebhookConfig) WebhookClient {
	return &webhookClient{
		client: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
		url:         cfg.URL,
		authKey:     cfg.AuthKey,
		rateLimiter: rate.NewLimiter(rate.Limit(cfg.RateLimitPerSecond), cfg.RateLimitPerSecond),
	}
}

func (w *webhookClient) SendMessage(ctx context.Context, phoneNumber, content string) (*WebhookResponse, error) {
	if err := w.rateLimiter.Wait(ctx); err != nil {
		logger.Get().Warn("rate limiter context cancelled", zap.Error(err))
		return nil, apperrors.Wrap(apperrors.ErrorCodeRateLimit, "rate limit wait cancelled", err)
	}

	reqBody := WebhookRequest{
		To:      phoneNumber,
		Content: content,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrorCodeInternal, "failed to marshal request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrorCodeInternal, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ins-auth-key", w.authKey)

	startTime := time.Now()
	resp, err := w.client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		logger.Get().Error("webhook request failed",
			zap.Error(err),
			zap.String("phone_number", phoneNumber),
			zap.Duration("duration", duration),
		)

		if ctx.Err() == context.DeadlineExceeded {
			return nil, apperrors.Wrap(apperrors.ErrorCodeTimeout, "webhook request timeout", err)
		}
		return nil, apperrors.Wrap(apperrors.ErrorCodeNetworkError, "network error during webhook request", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrorCodeInvalidResponse, "failed to read response body", err)
	}

	logger.Get().Info("webhook request completed",
		zap.String("phone_number", phoneNumber),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", duration),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Get().Error("webhook returned error status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(responseBody)),
		)

		if resp.StatusCode >= 500 {
			return nil, apperrors.New(apperrors.ErrorCodeServerError,
				fmt.Sprintf("webhook server error: %d", resp.StatusCode))
		}

		return nil, apperrors.New(apperrors.ErrorCodeInvalidResponse,
			fmt.Sprintf("webhook returned status %d: %s", resp.StatusCode, string(responseBody)))
	}

	var webhookResp WebhookResponse
	if err := json.Unmarshal(responseBody, &webhookResp); err != nil {
		logger.Get().Error("failed to unmarshal webhook response",
			zap.Error(err),
			zap.String("response_body", string(responseBody)),
		)
		return nil, apperrors.Wrap(apperrors.ErrorCodeInvalidResponse, "invalid JSON response from webhook", err)
	}

	if webhookResp.MessageID == "" {
		return nil, apperrors.New(apperrors.ErrorCodeInvalidResponse, "webhook response missing messageId")
	}

	return &webhookResp, nil
}
