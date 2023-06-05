package inmem

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/vkuksa/twatter/internal"
)

type Storage[V any] struct {
	mut sync.RWMutex
	m   map[string][]byte
}

func NewStorage[V any]() *Storage[V] {
	return &Storage[V]{m: make(map[string][]byte)}
}

// Saves data into in-memory storage
// Returns an error, if given key is not valid or value can not be marshalled to json
func (s *Storage[V]) Insert(msg internal.Message) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	msg.ID = id.String()
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("set: %w", err)
	}

	s.mut.Lock()
	defer s.mut.Unlock()
	s.m[msg.ID] = data
	return nil
}

// Retrieves all stored values of map
// Returns resulting collection and an error, if an error occured during unmarshalling
func (s *Storage[V]) GetAll() ([]V, error) {
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
