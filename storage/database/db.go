package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
)

// const defaultPerm = 0774
const ConnStr = "user=postgres password=204060 dbname=chats sslmode=disable"

type Storage struct {
} //???
//
//func NewStorage(message storage.Message) Storage {
//	return Storage{}
//}

//func PickRandom1(userName int) (page *storage.Message, err error) {
//	defer func() { err = e.WrapIfError("cannot pick random page", err) }()
//	path := filepath.Join(stor.connStr, strconv.Itoa(userName))
//
//	files, err := os.ReadDir(path)
//	if err != nil {
//
//		return nil, err
//	}
//
//	if len(files) == 0 {
//		return nil, storage.ErrNoSavedMessages
//	}
//
//	rand.New(rand.NewSource(time.Now().UnixNano()))
//	n := rand.Intn(len(files))
//
//	file := files[n]
//
//	return stor.decodeMessage(filepath.Join(path, file.Name()))
//}

func (stor Storage) PickRandom(chatId int, db *sql.DB) (message *storage.Dialogs, err error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, e.Wrap("couldn't begin the transaction", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Printf("couldn't rollback transaction: %s", err)
		}
	}(tx)
	//usersMessages, err := tx.Query("select index, content from dialogs where sender = $1", chatId)
	//if err != nil {
	//	return nil, e.Wrap("couldn't query the users messages", err)
	//}
	//var index int
	//var content string
	//usersIndexes := make([]int, 1, 2)
	//for usersMessages.Next() {
	//	err = usersMessages.Scan(&index, &content)
	//	if err != nil {
	//		return nil, e.Wrap("couldn't read the message", err)
	//	}
	//	usersIndexes = append(usersIndexes, index)
	//}
	//rand.New(rand.NewSource(time.Now().UnixNano()))
	//index = rand.Intn(len(usersIndexes))
	//defer func(usersMessages *sql.Rows) {
	//	err := usersMessages.Close()
	//	if err != nil {
	//		return
	//
	//	}
	//}(usersMessages)
	//message.Index = index
	//return &storage.Dialogs{
	//	Index: index,
	//	Message: storage.Message{
	//		Content: content,
	//		UserID:  chatId,
	//	},
	//}, tx.Commit()

	var index int64 // Объявляем index как int64
	var content string

	row := db.QueryRow("SELECT index, content FROM dialogs WHERE sender = $1 ORDER BY RANDOM() LIMIT 1", chatId)
	err = row.Scan(&index, &content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoSavedMessages
		}
		return nil, e.Wrap("couldn't read the message", err)
	}

	return &storage.Dialogs{
		Index: int(index),
		Message: storage.Message{
			Content: content,
			UserID:  chatId,
		},
	}, tx.Commit()
}

func HadleConn(connStr string) (db *sql.DB) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

func (stor Storage) Save(message *storage.Message, db *sql.DB) error {
	if message == nil {
		return errors.New("message is nil")
	}

	_, err := db.Exec("insert into dialogs (content, sender) values ($1, $2)", message.Content, message.UserID)
	if err != nil {
		return e.Wrap("could not save message", err)
	}
	return nil
}

func (stor Storage) IsExist(message *storage.Message, db *sql.DB) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, e.Wrap("couldn't begin the transaction", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Printf("couldn't rollback transaction: %s", err)
		}
	}(tx)
	var isEixist bool
	err = tx.QueryRow("select exists(select 1 from dialogs where content = $1)", message.Content).Scan(&isEixist)
	if err != nil {
		return false, e.Wrap("couldn't query the message", err)
	}
	return isEixist, tx.Commit()
}

func (stor Storage) Remove(index int, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return e.Wrap("could not begin transaction", err)
	}
	defer func(tx *sql.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("rollback failed: %v", err)
		}

	}(tx)

	var isExist bool
	err = tx.QueryRow("select exists(select 1 from dialogs where index = $1)", index).Scan(&isExist)
	if err != nil {
		return e.Wrap("the row does not exist", err)
	}
	if !isExist {
		return fmt.Errorf("index %d does not exist\n", index)
	}
	_, err = tx.Exec("delete from dialogs where index = $1", index)
	if err != nil {
		return e.Wrap("removing is impossible", err)
	}

	_, err = tx.Exec("update dialogs set index = index - 1 where index > $1", index)
	if err != nil {
		return e.Wrap("something wrong in updating indexes", err)
	}
	return tx.Commit()
}

//func Remove1(message *storage.Message) error {
//	fileName, err := fileName(message)
//	if err != nil {
//		return e.Wrap("removing is impossible", err)
//	}
//	path := filepath.Join(stor.connStr, strconv.Itoa(message.UserID), fileName)
//
//	if err := os.Remove(path); nil != err {
//		warning := fmt.Sprintf("removing file %s is impossible", path)
//
//		return e.Wrap(warning, err)
//	}
//	return nil
//} //----

//func (stor Storage) Save(message *storage.Message) (err error) {
//	if message == nil {
//		return errors.New("message is nil")
//	}
//
//	defer func() {
//		if err != nil {
//			err = e.Wrap("cannot save message", err)
//		}
//	}()
//
//	fPath := filepath.Join(stor.connStr, strconv.Itoa(message.UserID))
//
//	if err = os.MkdirAll(fPath, defaultPerm); err != nil {
//		return err
//	}
//
//	fName, err := fileName(message)
//	if err != nil {
//		return err
//	}
//
//	fPath = filepath.Join(fPath, fName)
//
//	file, err := os.Create(fPath)
//	if err != nil {
//		return err
//	}
//
//	defer func() {
//		if err != nil {
//			_ = os.Remove(fPath)
//		}
//		_ = file.Close()
//	}()
//
//	if err := gob.NewEncoder(file).Encode(message); err != nil {
//		return err
//	}
//
//	return nil
//} //stay logic the same but rewrite realisation

//func IsExist1(message *storage.Message) (bool, error) {
//	if message == nil {
//		return false, errors.New("page is nil")
//	}
//
//	fileName, err := fileName(message)
//	if err != nil {
//		return false, e.Wrap("impossible to check if file exists", err)
//	}
//
//	path := filepath.Join(stor.connStr, strconv.Itoa(message.UserID), fileName)
//
//	_, err = os.Stat(path)
//	switch {
//	case errors.Is(err, os.ErrNotExist):
//		return false, nil
//	case err != nil:
//		warning := fmt.Sprintf("checking if file %s exists is impossible", path)
//		return false, e.Wrap(warning, err)
//	}
//
//	return true, nil
//} //rewrite logics and keep something like this

//func (stor Storage) decodeMessage(filePath string) (message *storage.Message, err error) {
//	file, err := os.Open(filePath)
//	if err != nil {
//		return nil, e.Wrap("decoding is enable", err)
//	}
//	defer func() { _ = file.Close() }()
//
//	var m storage.Message
//
//	if err := gob.NewDecoder(file).Decode(&m); nil != err {
//		return nil, e.Wrap("decoding is enable", err)
//	}
//	return &m, nil
//} //kepp something like this

//func fileName(p *storage.Message) (string, error) {
//	return p.Hash()
//} //keep something like this
