package storage

import (
	"database/sql"
	"errors"
)

type Storage interface {
	Save(message *Message, db *sql.DB) error
	PickRandom(chatId int, db *sql.DB) (message *Dialogs, err error)
	Remove(index int, db *sql.DB) error
	IsExist(message *Message, db *sql.DB) (bool, error)
	IsUserNotExist(message *Message, db *sql.DB) (bool, error)
	SaveUser(message *Message, db *sql.DB) error
	RemoveUser(index int, db *sql.DB) error
}

var ErrNoSavedMessages = errors.New("there is not saved messages")

type Message struct {
	Content  string
	UserID   int
	Priority int
	Flag     int
}

type Dialogs struct {
	Index   int
	Message Message
}
