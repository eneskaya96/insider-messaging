package dto

import "time"

type CreateMessageRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

type MessageResponse struct {
	ID               string     `json:"id"`
	PhoneNumber      string     `json:"phone_number"`
	Content          string     `json:"content"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	SentAt           *time.Time `json:"sent_at,omitempty"`
	Attempts         int        `json:"attempts"`
	MaxAttempts      int        `json:"max_attempts"`
	LastError        string     `json:"last_error,omitempty"`
	ErrorCode        string     `json:"error_code,omitempty"`
	WebhookMessageID string     `json:"webhook_message_id,omitempty"`
}

type MessageListResponse struct {
	Messages   []MessageResponse `json:"messages"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

type MessageStatsResponse struct {
	TotalMessages   int64 `json:"total_messages"`
	PendingMessages int64 `json:"pending_messages"`
	SentMessages    int64 `json:"sent_messages"`
	FailedMessages  int64 `json:"failed_messages"`
}

type SchedulerStatusResponse struct {
	IsRunning       bool      `json:"is_running"`
	LastRunAt       time.Time `json:"last_run_at,omitempty"`
	TotalProcessed  int64     `json:"total_processed"`
	TotalSuccessful int64     `json:"total_successful"`
	TotalFailed     int64     `json:"total_failed"`
}
