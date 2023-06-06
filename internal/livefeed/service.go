package livefeed

import (
	"context"
	"fmt"

	"github.com/vkuksa/twatter/internal"
	msgqueue "github.com/vkuksa/twatter/internal/queue"
	msgstor "github.com/vkuksa/twatter/internal/storage"
)

// MessageService represents a structure, that serves primary business logic, around returning
type MessageService struct {
	storage          msgstor.Storage
	queue            msgqueue.Queue
	msgAddedNotifier EventNotifier
}

func NewService(s msgstor.Storage, q msgqueue.Queue, notifier EventNotifier) *MessageService {
	return &MessageService{storage: s, queue: q, msgAddedNotifier: notifier}
}

// Adds message to kafka stream
func (s *MessageService) AddMessage(ctx context.Context, msg string) {
	s.queue.Enqueue(ctx, msg)
}

// Retuns a message feed, consisting of previously stored messages from storage and new messages, that are stored during streaming
// Streaming new messages gives the live feed effect
func (svc *MessageService) GenerateMessageFeed(ctx context.Context) (chan internal.Message, error) {
	storedMessages, err := svc.storage.GetStoredMessages()
	if err != nil {
		return nil, fmt.Errorf("getfeed: %w", err)
	}

	msgChan := make(chan internal.Message)

	addedMessage := MessageAddedObserver(make(chan internal.Message))
	svc.msgAddedNotifier.RegisterObserver(&addedMessage)

	go func() {
		defer close(msgChan)
		defer svc.msgAddedNotifier.RemoveObserver(&addedMessage)

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
