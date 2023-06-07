package livefeed

import (
	"context"
	"fmt"

	"github.com/vkuksa/twatter/internal"
)

// Represents an interface for message inserter
type MessageInserter interface {
	InsertMessage(context.Context, string)
}

// Represents an interface for message retriever
type MessageRetriever interface {
	RetrieveAllMessages() ([]internal.Message, error)
}

// Represents an interface for notifying about creation of message through passed notifier
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

	ctx context.Context
}

// Returns new instance of MessageService
// Parameters: interface of storage implementing required behavioural interfaces
func NewMessageService(ctx context.Context, irn MessageInserterRetrieverNotifier) *MessageService {
	notifier := NewObserverNotifier()
	irn.SetMessageAddedNotifier(notifier)
	return &MessageService{ctx: ctx, msgStorage: irn, msgAddedNotifier: notifier}
}

// Adds message for storing by queue
func (s *MessageService) AddMessage(ctx context.Context, msg string) {
	s.msgStorage.InsertMessage(ctx, msg)
}

// Retuns a message feed, consisting of previously stored messages from storage and new messages, that are stored during streaming
// Streaming new messages gives the live feed effect
func (s *MessageService) GenerateMessageFeed(reqCtx context.Context) (chan internal.Message, error) {
	storedMessages, err := s.msgStorage.RetrieveAllMessages()
	if err != nil {
		return nil, fmt.Errorf("getfeed: %w", err)
	}

	msgChan := make(chan internal.Message)

	addedMessage := MessageAddedObserver(make(chan internal.Message))
	s.msgAddedNotifier.RegisterObserver(&addedMessage)

	go func() {
		defer func() {
			s.msgAddedNotifier.RemoveObserver(&addedMessage)
			close(msgChan)
		}()

		for _, msg := range storedMessages {
			select {
			case msgChan <- msg:
			case <-s.ctx.Done():
				return
			case <-reqCtx.Done():
				return
			}
		}

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-reqCtx.Done():
				return
			case msg, ok := <-addedMessage:
				// We are notified that message is added
				if !ok {
					return
				}

				select {
				// Passing the message to next receiver
				case msgChan <- msg:
				case <-s.ctx.Done():
					return
				case <-reqCtx.Done():
					return
				}
			}
		}
	}()

	return msgChan, nil
}
