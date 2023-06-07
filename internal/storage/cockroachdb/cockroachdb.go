package cockroachdb

import (
	"database/sql"
	"os"

	// pq must just be registered as driver.
	_ "github.com/lib/pq"
	"github.com/vkuksa/twatter/internal"
)

// MessageStore is a storage for messages based on cockroach db
type MessageStore struct {
	db *sql.DB
}

// NewMessageStorage creates a new storage for messages based on CockroachDB client.
//
// Call Close() method as soon as you're done working with DB.
func NewMessageStore() (*MessageStore, error) {
	conn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(); err != nil {
		return nil, err
	}

	return &MessageStore{db: conn}, nil
}

// Close closes a connection to database
func (s *MessageStore) Close() error {
	return s.db.Close()
}

// Inserts message into cockroach db
func (s *MessageStore) InsertMessage(content string) (internal.Message, error) {
	var msg internal.Message
	row := s.db.QueryRow("INSERT INTO messages (content) VALUES ($1) RETURNING id, content, created_at ", content)
	err := row.Scan(&msg.ID, &msg.Content, &msg.CreatedAt)
	return msg, err
}

// Retrieves all stored messages from database
func (s *MessageStore) RetrieveAllMessages() ([]internal.Message, error) {
	rows, err := s.db.Query("SELECT id, content, created_at FROM messages ORDER BY created_at ASC")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	msgs := []internal.Message{}

	for rows.Next() {
		msg := internal.Message{}
		err := rows.Scan(&msg.ID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}
