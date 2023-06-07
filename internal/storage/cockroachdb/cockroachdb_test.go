package cockroachdb

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vkuksa/twatter/internal"
)

// ! Requires cockroach db node running on environment
func MustOpenDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create the twatter_test database
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS twatter_test")
	if err != nil {
		t.Fatalf("failed to create twatter_test database: %v", err)
	}

	// Switch to the twatter_test database
	_, err = db.Exec("USE twatter_test")
	if err != nil {
		t.Fatalf("failed to switch to twatter_test database: %v", err)
	}

	// Create the messages table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS public.messages (
		id UUID NOT NULL DEFAULT gen_random_uuid(),
		content STRING NULL,
		created_at TIME NULL DEFAULT current_time():::TIME,
		CONSTRAINT messages_pkey PRIMARY KEY (id ASC)
	)`)
	if err != nil {
		t.Fatalf("failed to create messages table: %v", err)
	}

	return db
}

func TestMessageStore_InsertAndRetrieveAll(t *testing.T) {
	db := MustOpenDB(t)
	defer db.Close()

	store := &MessageStore{db: db}

	// Clean up any existing test data
	_, err := db.Exec("DELETE FROM messages")
	if err != nil {
		t.Fatalf("failed to delete existing messages: %v", err)
	}

	// Insert a new message
	content := "Hello, world!"
	msg, err := store.InsertMessage(content)
	if err != nil {
		t.Fatalf("failed to insert message: %v", err)
	}

	// Verify the inserted message
	expectedMsg := internal.Message{
		ID:        msg.ID,
		Content:   content,
		CreatedAt: msg.CreatedAt,
	}

	assert.Equal(t, expectedMsg, msg, "inserted message does not match expected")

	// Retrieve all stored messages
	msgs, err := store.RetrieveAllMessages()
	if err != nil {
		t.Fatalf("failed to retrieve stored messages: %v", err)
	}

	// Verify the retrieved messages
	expectedMsgs := []internal.Message{expectedMsg}
	assert.Equal(t, expectedMsgs, msgs, "retrieved messages do not match expected")
}
