package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eneskaya/insider-messaging/pkg/config"
	apperrors "github.com/eneskaya/insider-messaging/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSendMessage_Success(t *testing.T) {
	// Arrange - Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-auth-key", r.Header.Get("x-ins-auth-key"))

		// Verify request body
		var req WebhookRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "+905551234567", req.To)
		assert.Equal(t, "Test message", req.Content)

		// Send successful response
		resp := WebhookResponse{
			Message:   "Message sent successfully",
			MessageID: "webhook-msg-123",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     10,
		RateLimitPerSecond: 10,
	}

	client := NewWebhookClient(cfg)

	// Act
	result, err := client.SendMessage(context.Background(), "+905551234567", "Test message")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Message sent successfully", result.Message)
	assert.Equal(t, "webhook-msg-123", result.MessageID)
}

func TestSendMessage_ServerError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     10,
		RateLimitPerSecond: 10,
	}

	client := NewWebhookClient(cfg)

	// Act
	result, err := client.SendMessage(context.Background(), "+905551234567", "Test")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*apperrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrorCodeServerError, appErr.Code)
}

func TestSendMessage_BadRequest(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid phone number"))
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     10,
		RateLimitPerSecond: 10,
	}

	client := NewWebhookClient(cfg)

	// Act
	result, err := client.SendMessage(context.Background(), "invalid-phone", "Test")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*apperrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrorCodeInvalidResponse, appErr.Code)
	assert.Contains(t, err.Error(), "400")
}

func TestSendMessage_InvalidJSONResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     10,
		RateLimitPerSecond: 10,
	}

	client := NewWebhookClient(cfg)

	// Act
	result, err := client.SendMessage(context.Background(), "+905551234567", "Test")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*apperrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrorCodeInvalidResponse, appErr.Code)
	assert.Contains(t, err.Error(), "invalid JSON response")
}

func TestSendMessage_MissingMessageID(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"message": "Success",
			// messageId is missing
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     10,
		RateLimitPerSecond: 10,
	}

	client := NewWebhookClient(cfg)

	// Act
	result, err := client.SendMessage(context.Background(), "+905551234567", "Test")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "missing messageId")
}

func TestSendMessage_Timeout(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     0, // 0 seconds timeout - will cause immediate timeout
		RateLimitPerSecond: 10,
	}

	client := NewWebhookClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Act
	result, err := client.SendMessage(ctx, "+905551234567", "Test")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*apperrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrorCodeTimeout, appErr.Code)
}

func TestSendMessage_RateLimiting(t *testing.T) {
	// Arrange
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := WebhookResponse{
			Message:   "Success",
			MessageID: "msg-123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     10,
		RateLimitPerSecond: 2, // 2 requests per second
	}

	client := NewWebhookClient(cfg)

	// Act - Send 3 messages quickly
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := client.SendMessage(context.Background(), "+905551234567", "Test")
		assert.NoError(t, err)
	}
	duration := time.Since(start)

	// Assert - Should take at least 500ms due to rate limiting (3 requests at 2/sec)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(400))
	assert.Equal(t, 3, callCount)
}

func TestSendMessage_ContextCancelled(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.WebhookConfig{
		URL:                server.URL,
		AuthKey:            "test-auth-key",
		TimeoutSeconds:     10,
		RateLimitPerSecond: 10,
	}

	client := NewWebhookClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	result, err := client.SendMessage(ctx, "+905551234567", "Test")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, ok := err.(*apperrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, apperrors.ErrorCodeRateLimit, appErr.Code)
	assert.Contains(t, err.Error(), "rate limit wait cancelled")
}
