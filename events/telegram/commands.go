package telegram

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/events"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
	"tgBotRecommender/storage/database"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (proces *Processor) doCmd(text string, chatID int) (error, int) {
	text = strings.TrimSpace(text)
	log.Printf("got command: %s from %d", text, chatID)
	result := events.Command

	switch text {
	case RndCmd:

		return proces.sendRandom(chatID), result
	case HelpCmd:
		return proces.sendHelp(chatID), result
	case StartCmd:
		return proces.sendHello(chatID), result
	default:
		{
			result = events.Content
			err := proces.saveMessage(chatID, text)
			if err != nil {
				return err, result
			}
		}
	}
	return nil, result
}

func (proces *Processor) setPriority(text string, chatID int) (err error) {
	if err := proces.tg.SendMessage(chatID, msgSetPriority); err != nil {
		return err
	}
	text = strings.TrimSpace(text)
	log.Printf("got number: %s from %d", text, chatID)
	var priorityOrder []int
	number, err := strconv.Atoi(text)
	if err != nil {
		return e.Wrap(notANumber, err)
	}
	priorityOrder = append(priorityOrder, number)
	//if err := proces.tg.SendMessage(chatID, strconv.Itoa(number)); err != nil {
	//	return err
	//}

	//for _, value := range priorityOrder {
	//	if err := proces.tg.SendMessage(chatID, strconv.Itoa(value)); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func (proces *Processor) saveMessage(chatID int, message string) (err error) {
	db := database.HadleConn(database.ConnStr)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	defer func() { err = e.WrapIfError("Impossible to execute command of saving message", err) }()
	sendMsg := NewMessageSendler(chatID, proces.tg)
	messageInfo := &storage.Message{
		Content: message,
		UserID:  chatID,
	}
	isExists, err := proces.storage.IsExist(messageInfo, db)
	if err != nil {
		return err
	}
	if isExists {
		return sendMsg(msgAlreadyExists)
	}
	if err := proces.storage.Save(messageInfo, db); err != nil {
		return err
	}
	if err := proces.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}
	return nil
}

func (proces *Processor) sendRandom(chatID int) (err error) {
	defer func() { err = e.WrapIfError("Impossible to execute random command: fail to send random ", err) }()
	db := database.HadleConn(database.ConnStr)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	message, err := proces.storage.PickRandom(chatID, db)
	if err != nil && !errors.Is(err, storage.ErrNoSavedMessages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedMessages) {
		return proces.tg.SendMessage(chatID, msgNoSavedMessage)
	}
	if err := proces.tg.SendMessage(chatID, message.Message.Content); err != nil {
		return err
	}
	return proces.storage.Remove(message.Index, db)
}

func (proces *Processor) sendHelp(chatID int) error {
	return proces.tg.SendMessage(chatID, msgHelp)
}

func (proces *Processor) sendHello(chatID int) error {
	return proces.tg.SendMessage(chatID, msgHello)
}

func NewMessageSendler(chatID int, tg *tgClient.Client) func(string) error {
	return func(msg string) error {
		return tg.SendMessage(chatID, msg)
	}
}
