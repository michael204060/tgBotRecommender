package telegram

import (
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
	db, err := database.HandleConn()
	if err != nil {
		return e.Wrap("failed to connect to database", err)
	}
	defer db.Close()

	sendMsg := NewMessageSendler(chatID, proces.tg)
	messageInfo := &storage.Message{
		Content: message,
		UserID:  chatID,
	}

	isExists, err := proces.storage.IsExist(messageInfo, db)
	if err != nil {
		return e.Wrap("failed to check message existence", err)
	}
	if isExists {
		return sendMsg(msgAlreadyExists)
	}
	isUserNotExists, err := proces.storage.IsUserNotExist(messageInfo, db)
	if err != nil {
		return e.Wrap("failed to check user's existence", err)
	}
	if isUserNotExists {
		if err = proces.storage.Save(messageInfo, db); err != nil {
			return e.Wrap("failed to save user", err)
		}
	}

	if err := proces.storage.Save(messageInfo, db); err != nil {
		return e.Wrap("failed to save message", err)
	}

	return sendMsg(msgSaved)
}

func (proces *Processor) sendRandom(chatID int) error {
	db, err := database.HandleConn()
	if err != nil {
		return e.Wrap("failed to connect to database", err)
	}
	defer db.Close()

	message, err := proces.storage.PickRandom(chatID, db)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedMessages) {
			if err := proces.storage.RemoveUser(message.Index, db); err != nil {
				return e.Wrap("failed to remove user", err)
			}
			return proces.tg.SendMessage(chatID, msgNoSavedMessage)
		}
		return e.Wrap("failed to pick random message", err)
	}

	if err := proces.tg.SendMessage(chatID, message.Message.Content); err != nil {
		return e.Wrap("failed to send message", err)
	}

	if err := proces.storage.Remove(message.Index, db); err != nil {
		return e.Wrap("failed to remove message", err)
	}

	return nil
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
