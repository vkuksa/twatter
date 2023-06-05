package livefeed

import (
	"context"
	"fmt"
	"time"

	"github.com/vkuksa/twatter/internal"
	"go.uber.org/zap"
)

type Storage interface {
	Insert(msg internal.Message) error

	GetAll() ([]internal.Message, error)
}

type Service struct {
	logger   *zap.Logger
	storage  Storage
	addEvent AddEvent
}

func NewService(l *zap.Logger, s Storage) *Service {
	return &Service{logger: l, storage: s}
}

func (s *Service) Add(_ context.Context, content string) (internal.Message, error) {
	//TODO: implement backpressure
	msg := internal.Message{Content: content, CreatedAt: time.Now()}
	if err := s.storage.Insert(msg); err != nil {
		return msg, fmt.Errorf("add: %w", err)
	}
	s.addEvent.NotifyObservers(msg)

	return msg, nil
}

func (s *Service) GenerateFeed(ctx context.Context) (chan internal.Message, error) {
	storedMessages, err := s.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("getfeed: %w", err)
	}

	msgChan := make(chan internal.Message)

	messageAdded := MessageAddedObserver(make(chan internal.Message))
	s.addEvent.RegisterObserver(&messageAdded)

	go func() {
		defer close(msgChan)
		defer s.addEvent.RemoveObserver(&messageAdded)

		for _, msg := range storedMessages {
			msgChan <- msg
		}

		// Wait for events of adding a message, or context.Done() event
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-messageAdded:
				msgChan <- msg
			}
		}
	}()

	return msgChan, nil
}
