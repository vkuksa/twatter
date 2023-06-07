package inmem

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vkuksa/twatter/internal"
)

type Store[V any] struct {
	mut sync.RWMutex
	m   map[string][]byte
}

func NewStore[V any]() *Store[V] {
	return &Store[V]{m: make(map[string][]byte)}
}

func NewMessageStore() *Store[internal.Message] {
	return &Store[internal.Message]{m: make(map[string][]byte)}
}

// Saves data into in-memory Store
// Returns an error, if given key is not valid or value can not be marshalled to json
func (s *Store[V]) InsertMessage(content string) (internal.Message, error) {
	msg := internal.Message{Content: content, CreatedAt: time.Now()}
	id, err := uuid.NewRandom()
	if err != nil {
		return internal.Message{}, err
	}
	msg.ID = id.String()

	data, err := json.Marshal(msg)
	if err != nil {
		return msg, fmt.Errorf("set: %w", err)
	}

	s.mut.Lock()
	defer s.mut.Unlock()
	s.m[msg.ID] = data
	return msg, nil
}

// Retrieves all stored values of map
// Returns resulting collection and an error, if an error occured during unmarshalling
func (s *Store[V]) GetStoredMessages() ([]V, error) {
	s.mut.RLock()
	result := make([]V, 0, len(s.m))

	for _, data := range s.m {
		var v V
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, fmt.Errorf("get: %w", err)
		}
		result = append(result, v)
	}
	s.mut.RUnlock()

	return result, nil
}

// MutateStore modifies the internal map of the Store for testing purposes.
// It takes a map of message IDs to serialized message data and updates the store accordingly.
func (s *Store[V]) mutateStore(data map[string][]byte) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Iterate over the provided data and update the store's map
	for id, serializedData := range data {
		s.m[id] = serializedData
	}

	return nil
}
