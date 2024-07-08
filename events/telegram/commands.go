package telegram

import (
	"errors"
	"log"
	"net/url"
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

func (proces *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got command: %s from %s", text, username)

	if isAddCmd(text) {
		return proces.savePage(chatID, text, username)
		//TODO: AddPage()
	}

	switch text {
	case RndCmd:
		return proces.sendRandom(chatID, username)
	case HelpCmd:
		return proces.sendHelp(chatID)
	case StartCmd:
		return proces.sendHello(chatID)
	default:
		return proces.tg.SendMessage(chatID, msgUnknownCommand)

	}
}

func (proces *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfError("Impossible to execute command of saving page", err) }()

	sendMsg := NewMessageSendler(chatID, proces.tg)
	page := &storage.Page{
		Url:      pageURL,
		UserName: username,
	}

	isExists, err := proces.storage.IsExist(page)
	if err != nil {
		return err
	}
	if isExists {
		return sendMsg(msgAlreadyExists)
		//return proces.tg.SendMessage(chatID, msgAlreadyExists)
	}

	if err := proces.storage.Save(page); err != nil {
		return err
	}

	if err := proces.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}
	return nil
}

func (proces *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfError("Impossible to execute random command: fail to send random ", err) }()

	page, err := proces.storage.PickRandom(username)
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

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
