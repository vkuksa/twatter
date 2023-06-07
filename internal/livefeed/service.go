package livefeed

import (
	"context"
	"fmt"

	"github.com/vkuksa/twatter/internal"
)

// Represents an interface for inserting messages to storage
type MessageInserter interface {
	InsertMessage(context.Context, string)
}

type MessageRetriever interface {
	RetrieveAllMessages() ([]internal.Message, error)
}

type MessageAddedNotifier interface {
	SetMessageAddedNotifier(EventNotifier)
}

type MessageInserterRetrieverNotifier interface {
	MessageInserter
	MessageRetriever
	MessageAddedNotifier
}

// MessageService represents a structure, that serves primary business logic: Adding message to storage, returning and live-feeding stored messages
// The instance of EventNotifier should be shared between MessageService and Storage in order to receive updated messages
type MessageService struct {
	msgStorage       MessageInserterRetrieverNotifier
	msgAddedNotifier EventNotifier
}

// Returns new instance of MessageService
// Parameters: interface of storage implementing required behavioural interfaces
func NewMessageService(irn MessageInserterRetrieverNotifier) *MessageService {
	notifier := NewMessageAddedNotifier()
	irn.SetMessageAddedNotifier(notifier)
	return &MessageService{msgStorage: irn, msgAddedNotifier: notifier}
}

// Adds message for storing by queue
func (s *MessageService) AddMessage(ctx context.Context, msg string) {
	s.msgStorage.InsertMessage(ctx, msg)
}

// Retuns a message feed, consisting of previously stored messages from storage and new messages, that are stored during streaming
// Streaming new messages gives the live feed effect
func (s *MessageService) GenerateMessageFeed(ctx context.Context) (chan internal.Message, error) {
	storedMessages, err := s.msgStorage.RetrieveAllMessages()
	if err != nil {
		return nil, fmt.Errorf("getfeed: %w", err)
	}

	msgChan := make(chan internal.Message)

	addedMessage := MessageAddedObserver(make(chan internal.Message))
	s.msgAddedNotifier.RegisterObserver(&addedMessage)

	go func() {
		defer close(msgChan)
		defer s.msgAddedNotifier.RemoveObserver(&addedMessage)

		for _, msg := range storedMessages {
			msgChan <- msg
		}

		for {
			select {
			case <-ctx.Done():
				return
			case msgChan <- <-addedMessage:
			}
		}
	}()

	return msgChan, nil
}
