package msgstor

import "github.com/vkuksa/twatter/internal"

// Represents a message storage in a system
type Storage interface {
	InsertMessage(msg string) (internal.Message, error)

	RetrieveAllMessages() ([]internal.Message, error)
}
