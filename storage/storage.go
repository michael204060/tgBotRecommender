package storage

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strconv"
	"tgBotRecommender/lib/e"
)

type Storage interface {
	Save(message *Message) error
	PickRandom(userName int) (message *Message, err error)
	Remove(message *Message) error
	IsExist(message *Message) (bool, error)
}

var ErrNoSavedMessages = errors.New("there is not saved messages")

type Message struct {
	Content string
	UserID  int
}

func (p Message) Hash() (string, error) {
	hash := sha1.New()

	if _, err := io.WriteString(hash, p.Content); err != nil {
		return "", e.Wrap("impossible to calculate hash", err)
	}

	if _, err := io.WriteString(hash, strconv.Itoa(p.UserID)); err != nil {
		return "", e.Wrap("impossible to calculate hash", err)
	}
	//strconv.Atoi()
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
