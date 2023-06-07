package livefeed

import (
	"time"

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

// Represents concrete
type ObserverNotifier struct {
	observers []Observer
}

func NewObserverNotifier() *ObserverNotifier {
	return &ObserverNotifier{observers: make([]Observer, 0)}
}

// RegisterObserver adds an observer to the list of observers.
func (s *ObserverNotifier) RegisterObserver(observer Observer) {
	s.observers = append(s.observers, observer)
}

// RemoveObserver removes an observer from the list of observers.
func (s *ObserverNotifier) RemoveObserver(observer Observer) {
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// NotifyObservers notifies all observers with the given message.
func (s *ObserverNotifier) Notify(msg internal.Message) {
	for _, observer := range s.observers {
		observer.Update(msg)
	}
}

// MessageAddedObserver represents a concrete observer that should signal when message being added
type MessageAddedObserver chan internal.Message

// Update receives an update from the subject.
func (o *MessageAddedObserver) Update(m internal.Message) {
	select {
	case *o <- m:
	case <-time.After(1 * time.Second):
		return
		// Case protecting from race condition, where client will not be able to read a message, but observer was not removed from the list and was notified
	}
}
