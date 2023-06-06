package livefeed

import (
	"github.com/vkuksa/twatter/internal"
)

// Observer represents an observer that receives updates from the subject.
type Observer interface {
	Update(internal.Message)
}

// Notifier represents a subject that notifies observers
type EventNotifier interface {
	RegisterObserver(observer Observer)
	RemoveObserver(observer Observer)
	Notify(internal.Message)
}

// MessageAdded represents a concrete subject that implements the Subject interface.
type MessageAdded struct {
	observers []Observer
}

func NewMessageAddedNotifier() EventNotifier {
	return &MessageAdded{observers: make([]Observer, 0)}
}

// RegisterObserver adds an observer to the list of observers.
func (s *MessageAdded) RegisterObserver(observer Observer) {
	s.observers = append(s.observers, observer)
}

// RemoveObserver removes an observer from the list of observers.
func (s *MessageAdded) RemoveObserver(observer Observer) {
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// NotifyObservers notifies all observers with the given message.
func (s *MessageAdded) Notify(msg internal.Message) {
	for _, observer := range s.observers {
		observer.Update(msg)
	}
}

// TODO: think of this interfaces
// MessageAddedObserver represents a concrete observer that should signal when message being added
type MessageAddedObserver chan internal.Message

// Update receives an update from the subject.
func (o *MessageAddedObserver) Update(m internal.Message) {
	select {
	case *o <- m:
	default:
		// Case protecting from race condition, where client will not be able to read a message, but observer was not removed from the list and was notified
	}
}
