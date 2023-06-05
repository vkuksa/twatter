package cockroachdb

import (
	"database/sql"
	"os"

	// pq must just be registered as driver.
	_ "github.com/lib/pq"
	"github.com/vkuksa/twatter/internal"
)

// Client is a gokv.Store implementation for CockroachDB.
type Client struct {
	db *sql.DB
}

// NewClient creates a new CockroachDB client.
//
// Call Close() method as soon as you're done working with DB.
func NewClient() (*Client, error) {
	conn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	client := &Client{db: conn}

	err = conn.Ping()
	if err != nil {
		return client, err
	}

	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS messages (id STRING PRIMARY KEY, content STRING, time TIMETZ)")
	if err != nil {
		return client, err
	}

	return client, nil
}

// Close closes a connection to database
func (c *Client) Close() error {
	return c.db.Close()
}

// Insert inserts message into cockroach db
func (s *Client) Insert(msg internal.Message) error {
	if _, err := s.db.Exec("INSERT INTO messages (content, created_at) VALUES ($1, $2)", msg.Content, msg.CreatedAt); err != nil {
		return err
	}

	return nil
}

// Retrieves all stored messages from database
func (s *Client) GetAll() ([]internal.Message, error) {
	rows, err := s.db.Query("SELECT id, content, created_at FROM messages ORDER BY created_at ASC")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	msgs := []internal.Message{}

	for rows.Next() {
		msg := internal.Message{}
		rows.Scan(&msg.ID, &msg.Content, &msg.CreatedAt)
		msgs = append(msgs, msg)
	}

	return msgs, nil
}
