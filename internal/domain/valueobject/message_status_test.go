package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantError bool
		expected  MessageStatus
	}{
		{
			name:      "valid pending status",
			status:    "pending",
			wantError: false,
			expected:  MessageStatusPending,
		},
		{
			name:      "valid processing status",
			status:    "processing",
			wantError: false,
			expected:  MessageStatusProcessing,
		},
		{
			name:      "valid sent status",
			status:    "sent",
			wantError: false,
			expected:  MessageStatusSent,
		},
		{
			name:      "valid failed status",
			status:    "failed",
			wantError: false,
			expected:  MessageStatusFailed,
		},
		{
			name:      "invalid status",
			status:    "unknown",
			wantError: true,
		},
		{
			name:      "empty status",
			status:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewMessageStatus(tt.status)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid message status")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, status)
			}
		})
	}
}

func TestMessageStatus_String(t *testing.T) {
	assert.Equal(t, "pending", MessageStatusPending.String())
	assert.Equal(t, "processing", MessageStatusProcessing.String())
	assert.Equal(t, "sent", MessageStatusSent.String())
	assert.Equal(t, "failed", MessageStatusFailed.String())
}

func TestMessageStatus_IsPending(t *testing.T) {
	assert.True(t, MessageStatusPending.IsPending())
	assert.False(t, MessageStatusProcessing.IsPending())
	assert.False(t, MessageStatusSent.IsPending())
	assert.False(t, MessageStatusFailed.IsPending())
}

func TestMessageStatus_IsProcessing(t *testing.T) {
	assert.False(t, MessageStatusPending.IsProcessing())
	assert.True(t, MessageStatusProcessing.IsProcessing())
	assert.False(t, MessageStatusSent.IsProcessing())
	assert.False(t, MessageStatusFailed.IsProcessing())
}

func TestMessageStatus_IsSent(t *testing.T) {
	assert.False(t, MessageStatusPending.IsSent())
	assert.False(t, MessageStatusProcessing.IsSent())
	assert.True(t, MessageStatusSent.IsSent())
	assert.False(t, MessageStatusFailed.IsSent())
}

func TestMessageStatus_IsFailed(t *testing.T) {
	assert.False(t, MessageStatusPending.IsFailed())
	assert.False(t, MessageStatusProcessing.IsFailed())
	assert.False(t, MessageStatusSent.IsFailed())
	assert.True(t, MessageStatusFailed.IsFailed())
}

func TestMessageStatus_CanProcess(t *testing.T) {
	assert.True(t, MessageStatusPending.CanProcess())
	assert.False(t, MessageStatusProcessing.CanProcess())
	assert.False(t, MessageStatusSent.CanProcess())
	assert.False(t, MessageStatusFailed.CanProcess())
}
