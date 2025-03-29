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
	var index int64
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

		// Используем встроенный SQL
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

func (stor Storage) Save(message *storage.Message, db *sql.DB) error {
	if message == nil {
		return errors.New("message is nil")
	}
	var maxId int64
	err := db.QueryRow("select coalesce(max(index), 0) from dialogs").Scan(&maxId)
	if err != nil {
		return e.Wrap("couldn't get the max id", err)
	}
	maxId = maxId + 1
	_, err = db.Exec("insert into dialogs (index, content, sender) values ($1, $2, $3)", maxId, message.Content, message.UserID)
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
