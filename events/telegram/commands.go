package telegram

import (
	"errors"
	"log"
	"strings"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (proces *Processor) doCmd(text string, chatID int) error {
	text = strings.TrimSpace(text)
	log.Printf("got command: %s from %d", text, chatID)
	switch text {
	case RndCmd:
		return proces.sendRandom(chatID)
	case HelpCmd:
		return proces.sendHelp(chatID)
	case StartCmd:
		return proces.sendHello(chatID)
	default:
		{
			err := proces.savePage(chatID, text)
			if err != nil {
				return err
			}
			err = proces.setPriority(chatID)
			if err != nil {
				return err
			}
			//switch text {
			//
			//}
		}
	}
	return nil
}

func (proces *Processor) setPriority(chatID int) (err error) {
	if err := proces.tg.SendMessage(chatID, msgSetPriority); err != nil {
		return err
	}
	return nil
}

func (proces *Processor) savePage(chatID int, message string) (err error) {
	defer func() { err = e.WrapIfError("Impossible to execute command of saving page", err) }()
	sendMsg := NewMessageSendler(chatID, proces.tg)
	page := &storage.Page{
		Url:    message,
		UserID: chatID,
	}
	isExists, err := proces.storage.IsExist(page)
	if err != nil {
		return err
	}
	if isExists {
		return sendMsg(msgAlreadyExists)
	}
	if err := proces.storage.Save(page); err != nil {
		return err
	}
	if err := proces.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}
	return nil
}

func (proces *Processor) sendRandom(chatID int) (err error) {
	defer func() { err = e.WrapIfError("Impossible to execute random command: fail to send random ", err) }()
	page, err := proces.storage.PickRandom(chatID)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return proces.tg.SendMessage(chatID, msgNoSavedPage)
	}
	if err := proces.tg.SendMessage(chatID, page.Url); err != nil {
		return err
	}
	return proces.storage.Remove(page)
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
