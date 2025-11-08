package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxChars  int
		wantError bool
	}{
		{
			name:      "valid content",
			content:   "Hello World",
			maxChars:  160,
			wantError: false,
		},
		{
			name:      "empty content",
			content:   "",
			maxChars:  160,
			wantError: true,
		},
		{
			name:      "content exceeds limit",
			content:   strings.Repeat("a", 161),
			maxChars:  160,
			wantError: true,
		},
		{
			name:      "content at limit",
			content:   strings.Repeat("a", 160),
			maxChars:  160,
			wantError: false,
		},
		{
			name:      "unicode characters",
			content:   "Türkçe karakterler: ğüşıöç",
			maxChars:  160,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewMessageContent(tt.content, tt.maxChars)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, content)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, content)
				assert.Equal(t, tt.content, content.String())
			}
		})
	}
}

func TestMessageContentLength(t *testing.T) {
	content, _ := NewMessageContent("Hello", 160)
	assert.Equal(t, 5, content.Length())

	unicodeContent, _ := NewMessageContent("Merhaba", 160)
	assert.Equal(t, 7, unicodeContent.Length())
}
