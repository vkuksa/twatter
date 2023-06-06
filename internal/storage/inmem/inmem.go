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
func (s *Store[V]) Insert(content string) (internal.Message, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return internal.Message{}, err
	}
	msg := internal.Message{ID: id.String(), Content: content, CreatedAt: time.Now()}

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
func (s *Store[V]) GetAll() ([]V, error) {
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
