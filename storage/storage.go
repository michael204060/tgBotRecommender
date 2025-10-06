package storage

import (
	"database/sql"
	"errors"
)

type Storage interface {
	SaveWithPriority(message *Message, db *sql.DB) error
	IsPriorityExists(userID int, priority int, db *sql.DB) (bool, error)
	PickHighestPriority(userID int, db *sql.DB) (*Dialogs, error)
	NormalizePriorities(userID int, db *sql.DB) error
	LowerPriority(messageID int, userID int, db *sql.DB) error
	RemoveByMessageID(messageID int, db *sql.DB) error
	HigherPriority(messageID int, userID int, db *sql.DB) error
}

var ErrNoSavedMessages = errors.New("no saved messages")

type Message struct {
	Content  string
	UserID   int
	Priority int
}

type Dialogs struct {
	Index   int
	Message Message
}
