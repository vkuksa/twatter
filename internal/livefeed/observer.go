package livefeed

import (
	"time"

	"github.com/vkuksa/twatter/internal"
)

// Observer represents an observer that receives updates from the subject.
type Observer interface {
	Update(internal.Message)
}

// Subject represents a subject that can be observed by observers.
type Subject interface {
	RegisterObserver(Observer)
	RemoveObserver(Observer)
	NotifyObservers(internal.Message)
}

// ConcreteSubject represents a concrete subject that implements the Subject interface.
type AddEvent struct {
	observers []Observer
}

// RegisterObserver adds an observer to the list of observers.
func (s *AddEvent) RegisterObserver(observer Observer) {
	s.observers = append(s.observers, observer)
}

// RemoveObserver removes an observer from the list of observers.
func (s *AddEvent) RemoveObserver(observer Observer) {
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// NotifyObservers notifies all observers with the given message.
func (s *AddEvent) NotifyObservers(msg internal.Message) {
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
	// Case protecting from race condition, where client will not be able to read a message, but observer was not removed from the list
	case <-time.After(time.Second):
	}
}
