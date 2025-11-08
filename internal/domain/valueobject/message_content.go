package valueobject

import (
	"fmt"
	"unicode/utf8"
)

type MessageContent struct {
	value    string
	maxChars int
}

func NewMessageContent(content string, maxChars int) (*MessageContent, error) {
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	charCount := utf8.RuneCountInString(content)
	if charCount > maxChars {
		return nil, fmt.Errorf("message content exceeds maximum length of %d characters (got %d)", maxChars, charCount)
	}

	return &MessageContent{
		value:    content,
		maxChars: maxChars,
	}, nil
}

func (m *MessageContent) String() string {
	return m.value
}

func (m *MessageContent) Length() int {
	return utf8.RuneCountInString(m.value)
}

func (m *MessageContent) Equals(other *MessageContent) bool {
	if other == nil {
		return false
	}
	return m.value == other.value
}
