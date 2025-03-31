package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
	"time"
)

type Storage struct {
}

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
	var id int64
	var content string
	var priority int
	var flag int
	row := db.QueryRow("SELECT id, content, priority, flag FROM messages WHERE user_id = $1 ORDER BY RANDOM() LIMIT 1", chatId)
	err = row.Scan(&id, &content, &priority, &flag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoSavedMessages
		}
		return nil, e.Wrap("couldn't read the message", err)
	}
	return &storage.Dialogs{
		Index: int(id),
		Message: storage.Message{
			Content:  content,
			UserID:   chatId,
			Priority: priority,
			Flag:     flag,
		},
	}, tx.Commit()
}

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

func (stor Storage) SaveUser(message *storage.Message, db *sql.DB) error {
	if message == nil {
		return errors.New("message is nil")
	}
	var maxId int64
	err := db.QueryRow("select coalesce(max(id), 0) from users").Scan(&maxId)
	if err != nil {
		return e.Wrap("couldn't get the max id", err)
	}
	maxId = maxId + 1
	_, err = db.Exec("insert into users (id, sender) values ($1)", maxId, message.UserID)
	if err != nil {
		return e.Wrap("couldn't save user", err)
	}
	return nil
}

func (stor Storage) Save(message *storage.Message, db *sql.DB) error {
	if message == nil {
		return errors.New("message is nil")
	}
	var maxId int64
	err := db.QueryRow("select coalesce(max(id), 0) from messages").Scan(&maxId)
	if err != nil {
		return e.Wrap("couldn't get the max id", err)
	}
	maxId = maxId + 1
	_, err = db.Exec("insert into messages (id, content, user_id) values ($1, $2, $3)", maxId, message.Content, message.UserID)
	if err != nil {
		return e.Wrap("couldn't save the message", err)
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
	var isExist bool
	err = tx.QueryRow("select exists(select 1 from messages where content = $1)", message.Content).Scan(&isExist)
	if err != nil {
		return false, e.Wrap("couldn't query the message", err)
	}
	return isExist, tx.Commit()
}

func (stor Storage) IsUserNotExist(message *storage.Message, db *sql.DB) (bool, error) {
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
	var isExist bool
	err = tx.QueryRow("select exists(select 1 from users where sender = $1)", message.UserID).Scan(&isExist)
	if err != nil {
		return true, e.Wrap("couldn't query the message", err)
	}
	return false, tx.Commit()
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
	err = tx.QueryRow("select exists(select 1 from messages where id = $1)", index).Scan(&isExist)
	if err != nil {
		return e.Wrap("the row does not exist", err)
	}
	if !isExist {
		return fmt.Errorf("index %d does not exist\n", index)
	}
	_, err = tx.Exec("delete from messages where id = $1", index)
	if err != nil {
		return e.Wrap("removing is impossible", err)
	}
	_, err = tx.Exec("update messages set id = id - 1 where id > $1", index)
	if err != nil {
		return e.Wrap("something wrong in updating indexes", err)
	}
	return tx.Commit()
}

func (stor Storage) RemoveUser(index int, db *sql.DB) error {
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
	err = tx.QueryRow("select exists(select 1 from messages where id = $1)", index).Scan(&isExist)
	if err != nil {
		return e.Wrap("the row does not exist", err)
	}
	if !isExist {
		_, err = tx.Exec("delete from users where id = $1", index)
		if err != nil {
			return e.Wrap("removing is impossible", err)
		}
		_, err = tx.Exec("update users set id = id - 1 where id > $1", index)
		if err != nil {
			return e.Wrap("something wrong in updating indexes", err)
		}
	}
	return tx.Commit()
}
