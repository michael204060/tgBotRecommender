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
}

var ErrNoSavedMessages = errors.New("there is not saved messages")

type Message struct {
	Content string
	UserID  int
}

type Dialogs struct {
	Index   int
	Message Message
}

//func (p Message) Hash() (string, error) {
//	hash := sha1.New()
//
//	if _, err := io.WriteString(hash, p.Content); err != nil {
//		return "", e.Wrap("impossible to calculate hash", err)
//	}
//
//	if _, err := io.WriteString(hash, strconv.Itoa(p.UserID)); err != nil {
//		return "", e.Wrap("impossible to calculate hash", err)
//	}
//	//strconv.Atoi()
//	return fmt.Sprintf("%x", hash.Sum(nil)), nil
//}
