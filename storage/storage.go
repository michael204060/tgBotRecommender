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
	Save(page *Page) error
	PickRandom(userName int) (page *Page, err error)
	Remove(page *Page) error
	IsExist(page *Page) (bool, error)
}

var ErrNoSavedPages = errors.New("there is not saved pages")

type Page struct {
	Url    string
	UserID int
}

func (p Page) Hash() (string, error) {
	hash := sha1.New()

	if _, err := io.WriteString(hash, p.Url); err != nil {
		return "", e.Wrap("impossible to calculate hash", err)
	}

	if _, err := io.WriteString(hash, strconv.Itoa(p.UserID)); err != nil {
		return "", e.Wrap("impossible to calculate hash", err)
	}
	//strconv.Atoi()
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
