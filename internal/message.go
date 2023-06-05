package internal

import (
	"fmt"
	"time"
)

// Message is a message, that was accepted by a system
// Content corresponds to the content of original message
// ID and CreatedAt fields are being set by a system
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// Validates a message
func (m Message) Validate() error {
	if m.ID == "" {
		return NewErrorf(ErrorCodeInvalidArgument, "id is required")
	}

	if m.Content == "" {
		return NewErrorf(ErrorCodeInvalidArgument, "content is required")
	}

	if m.CreatedAt == (time.Time{}) {
		return NewErrorf(ErrorCodeInvalidArgument, "createdAt is not set")
	}

	return nil
}

func (m Message) String() string {
	return fmt.Sprintf("%d:%02d:%02d %s\n", m.CreatedAt.Hour(), m.CreatedAt.Minute(), m.CreatedAt.Second(), m.Content)
}
