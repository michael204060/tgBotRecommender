package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"tgBotRecommender/storage"
	"time"
)

type Storage struct{}

//go:embed init.sql
var initSQL string

func HandleConn() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		5432,
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))

	maxAttempts := 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Attempt %d: failed to open database: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			db.Close()
			log.Printf("Attempt %d: failed to ping database: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		if _, err := db.Exec(initSQL); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create table: %v", err)
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		return db, nil
	}
	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxAttempts)
}

func (stor Storage) SaveWithPriority(message *storage.Message, db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO messages (content, user_id, priority) VALUES ($1, $2, $3)",
		message.Content, message.UserID, message.Priority,
	)
	return err
}

func (stor Storage) IsPriorityExists(userID int, priority int, db *sql.DB) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM messages WHERE user_id = $1 AND priority = $2)",
		userID, priority,
	).Scan(&exists)
	return exists, err
}

func (stor Storage) PickHighestPriority(userID int, db *sql.DB) (*storage.Dialogs, error) {
	var id int
	var content string
	var priority int

	err := db.QueryRow(`
		SELECT id, content, priority 
		FROM messages 
		WHERE user_id = $1 
		ORDER BY priority ASC 
		LIMIT 1`, userID,
	).Scan(&id, &content, &priority)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoSavedMessages
		}
		return nil, err
	}

	return &storage.Dialogs{
		Index: id,
		Message: storage.Message{
			Content:  content,
			UserID:   userID,
			Priority: priority,
		},
	}, nil
}

func (stor Storage) NormalizePriorities(userID int, db *sql.DB) error {
	_, err := db.Exec(`
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY priority) as new_priority
			FROM messages 
			WHERE user_id = $1
		)
		UPDATE messages m
		SET priority = r.new_priority
		FROM ranked r
		WHERE m.id = r.id AND m.user_id = $1`, userID)
	return err
}

func (stor Storage) LowerPriority(messageID int, userID int, db *sql.DB) error {
	_, err := db.Exec(`
		UPDATE messages 
		SET priority = (SELECT COALESCE(MAX(priority), 0) + 1 FROM messages WHERE user_id = $1)
		WHERE id = $2 AND user_id = $1`, userID, messageID)
	return err
}

func (stor Storage) RemoveByMessageID(messageID int, db *sql.DB) error {
	_, err := db.Exec("DELETE FROM messages WHERE id = $1", messageID)
	return err
}
