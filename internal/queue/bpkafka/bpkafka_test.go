package bpkafka

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vkuksa/twatter/internal"
	"github.com/vkuksa/twatter/internal/livefeed"
	"go.uber.org/zap"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) InsertMessage(msg string) (internal.Message, error) {
	args := m.Called(msg)
	return args.Get(0).(internal.Message), args.Error(1)
}

func (m *MockStorage) GetStoredMessages() ([]internal.Message, error) {
	args := m.Called()
	return args.Get(0).([]internal.Message), args.Error(1)
}

type MockEventNotifier struct {
	mock.Mock
}

func (m *MockEventNotifier) RegisterObserver(observer livefeed.Observer) {
	m.Called(observer)
}

func (m *MockEventNotifier) RemoveObserver(observer livefeed.Observer) {
	m.Called(observer)
}

func (m *MockEventNotifier) Notify(msg internal.Message) {
	m.Called(msg)
}

//! Requires local running instance of kafka

func TestEnqueue(t *testing.T) {
	// Create a mock storage
	mockStorage := new(MockStorage)
	mockStorage.On("InsertMessage", "test message").Return(internal.Message{ID: "123", Content: "test message", CreatedAt: time.Now()}, nil)

	// Create a mock event notifier
	mockNotifier := new(MockEventNotifier)

	// Create a new queue
	queue, err := NewBackpressureQueue(context.Background(), zap.NewNop(), mockStorage, mockNotifier, "", 1)
	assert.NoError(t, err)

	// Enqueue a message
	queue.Enqueue(context.Background(), "test message")

	// Assert that the storage's InsertMessage method was called
	mockStorage.AssertCalled(t, "InsertMessage", "test message")

	// Assert that the event notifier's Notify method was called
	mockNotifier.AssertCalled(t, "Notify", mock.AnythingOfType("internal.Message"))
}

func TestStart(t *testing.T) {
	// Create a mock storage
	mockStorage := new(MockStorage)

	// Create a mock event notifier
	mockNotifier := new(MockEventNotifier)

	// Create a new queue
	queue, err := NewBackpressureQueue(context.Background(), zap.NewNop(), mockStorage, mockNotifier, "", 1)
	assert.NoError(t, err)

	// Start the queue
	queue.Start()

	// Wait for a short duration to allow the workers to start
	time.Sleep(100 * time.Millisecond)

	// Assert that the insertion worker goroutines have started
	assert.Equal(t, 1, runtime.NumGoroutine())

	// Shutdown the queue
	queue.Shutdown()

	// Wait for a short duration to allow the workers to exit
	time.Sleep(100 * time.Millisecond)

	// Assert that all goroutines have exited
	assert.Equal(t, 0, runtime.NumGoroutine())
}
