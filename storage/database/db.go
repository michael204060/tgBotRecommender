package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"tgBotRecommender/storage"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct{}

//go:embed init.sql
var initSQL string

func HandleConn() (*sql.DB, error) {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	log.Printf("Connecting to database using DATABASE_URL")

	var db *sql.DB
	var err error

	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			log.Printf("Attempt %d: failed to open database: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		err = db.Ping()
		if err != nil {
			log.Printf("Attempt %d: failed to ping database: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		log.Printf("Successfully connected to database on Render")

		if err := initTables(db); err != nil {
			log.Printf("Warning: failed to init tables: %v", err)
		}

		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxAttempts)
}

func initTables(db *sql.DB) error {
	_, err := db.Exec(initSQL)
	if err != nil {
		return fmt.Errorf("failed to init tables: %w", err)
	}
	log.Printf("Database tables initialized successfully")
	return nil
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

func (stor Storage) HigherPriority(messageID int, userID int, db *sql.DB) error {
	_, err := db.Exec(`
        UPDATE messages 
        SET priority = 0
        WHERE id = $1 AND user_id = $2`, messageID, userID)
	return err
}
