package storage

import (
	"errors"
)

type Storage interface {
	SaveWithPriority(message *Message) error
	IsPriorityExists(userID int, priority int) (bool, error)
	PickHighestPriority(userID int) (*Dialogs, error)
	NormalizePriorities(userID int) error
	LowerPriority(messageID int, userID int) error
	RemoveByMessageID(messageID int) error
	HigherPriority(messageID int, userID int) error
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
