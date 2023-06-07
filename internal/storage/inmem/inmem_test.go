package inmem

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vkuksa/twatter/internal"
)

func TestStore_InsertMessage(t *testing.T) {
	store := NewMessageStore()
	content := "Test Message"

	msg, err := store.InsertMessage(content)

	assert.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, content, msg.Content)
	assert.NotZero(t, msg.CreatedAt)

	assert.Equal(t, len(store.m), 1)
}

func TestStore_GetStoredMessages(t *testing.T) {
	store := NewMessageStore()
	content := "Test Message"

	msg := internal.Message{
		ID:        "1",
		Content:   content,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	_ = store.mutateStore(map[string][]byte{
		"1": data,
	})

	storedMessages, err := store.GetStoredMessages()

	assert.NoError(t, err)
	assert.Len(t, storedMessages, 1)

	storedMsg := storedMessages[0]
	assert.Equal(t, msg.ID, storedMsg.ID)
	assert.Equal(t, msg.Content, storedMsg.Content)
	// We use strings.Split, because golang appends monotonic clock	https://stackoverflow.com/questions/51165616/unexpected-output-from-time-time
	got := storedMsg.CreatedAt.Format("2006-01-02 15:04:05 MST")
	assert.Equal(t, msg.CreatedAt.Format("2006-01-02 15:04:05 MST"), got)
}
