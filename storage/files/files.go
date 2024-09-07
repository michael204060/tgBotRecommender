package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
	"time"
)

const defaultPerm = 0774

type Storage struct {
	basePath string
}

func NewStorage(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (stor Storage) PickRandom(userName int) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfError("cannot pick random page", err) }()
	path := filepath.Join(stor.basePath, strconv.Itoa(userName))

	files, err := os.ReadDir(path)
	if err != nil {

		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return stor.decodePage(filepath.Join(path, file.Name()))
}

func (stor Storage) Remove(p *storage.Page) error {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("removing is impossible", err)
	}
	path := filepath.Join(stor.basePath, strconv.Itoa(p.UserID), fileName)

	if err := os.Remove(path); nil != err {
		msg := fmt.Sprintf("removing file %s is impossible", path)

		return e.Wrap(msg, err)
	}
	return nil
}

func (stor Storage) Save(page *storage.Page) (err error) {
	if page == nil {
		return errors.New("page is nil")
	}

	defer func() {
		if err != nil {
			err = e.Wrap("cannot save page", err)
		}
	}()

	fPath := filepath.Join(stor.basePath, strconv.Itoa(page.UserID))

	if err = os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	fName, err := fileName(page)
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = os.Remove(fPath)
		}
		_ = file.Close()
	}()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (stor Storage) IsExist(p *storage.Page) (bool, error) {
	if p == nil {
		return false, errors.New("page is nil")
	}

	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("impossible to check if file exists", err)
	}

	path := filepath.Join(stor.basePath, strconv.Itoa(p.UserID), fileName)

	_, err = os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("checking if file %s exists is impossible", path)
		return false, e.Wrap(msg, err)
	}

	return true, nil
}

func (stor Storage) decodePage(filePath string) (page *storage.Page, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("decoding is enable", err)
	}
	defer func() { _ = file.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(file).Decode(&p); nil != err {
		return nil, e.Wrap("decoding is enable", err)
	}
	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
