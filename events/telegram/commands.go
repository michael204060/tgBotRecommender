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
)

const (
	MikeId  = 1113360256
	SoniaId = 1470858378
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
			err := proces.savePage(chatID, text)
			if err != nil {
				return err, result
			}
		}
	}
	return nil, result
}

func (proces *Processor) setPriority(text string, chatID int) (err error) {
	switch chatID {
	case MikeId:
		if err := proces.tg.SendMessage(chatID, msgSetPriorityMike); err != nil {
			return err
		}
	case SoniaId:
		if err := proces.tg.SendMessage(chatID, msgSetPrioritySonia); err != nil {
			return err
		}
	default:
		if err := proces.tg.SendMessage(chatID, msgSetPriority); err != nil {
			return err
		}
	}

	text = strings.TrimSpace(text)
	log.Printf("got number: %s from %d", text, chatID)
	var priorityOrder []int
	number, err := strconv.Atoi(text)
	if err != nil {
		return e.Wrap(notANumber, err)
	}
	switch chatID {
	case MikeId:
		if err != nil {
			return e.Wrap(notANumber, err)
		}
	case SoniaId:
		if err != nil {
			return e.Wrap(notANumberSonia, err)
		}
	default:
		if err != nil {
			return e.Wrap(notANumber, err)
		}
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
	switch chatID {
	case MikeId:
		if isExists {
			return sendMsg(msgAlreadyExists)
		}
	case SoniaId:
		if isExists {
			return sendMsg(msgAlreadyExistsSonia)
		}
	default:
		if isExists {
			return sendMsg(msgAlreadyExists)
		}
	}
	switch chatID {
	case MikeId:
		if err := proces.tg.SendMessage(chatID, msgSaved); err != nil {
			return err
		}
	case SoniaId:
		if err := proces.tg.SendMessage(chatID, msgSavedSonia); err != nil {
			return err
		}
	default:
		if err := proces.tg.SendMessage(chatID, msgSaved); err != nil {
			return err
		}
	}
	switch chatID {
	case MikeId:
		if isExists {
			return sendMsg(msgAlreadyExists)
		}
	case SoniaId:
		if isExists {
			return sendMsg(msgAlreadyExistsSonia)
		}
	default:
		if isExists {
			return sendMsg(msgAlreadyExists)
		}
	}
	if err := proces.storage.Save(page); err != nil {
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
	switch chatID {
	case MikeId:
		if errors.Is(err, storage.ErrNoSavedPages) {
			return proces.tg.SendMessage(chatID, msgNoSavedPage)
		}
	case SoniaId:
		if errors.Is(err, storage.ErrNoSavedPages) {
			return proces.tg.SendMessage(chatID, msgExists)
		} else {
			proces.tg.SendMessage(chatID, msgNoSavedPage)
		}
	default:
		if errors.Is(err, storage.ErrNoSavedPages) {
			return proces.tg.SendMessage(chatID, msgNoSavedPage)
		}
	}

	if err := proces.tg.SendMessage(chatID, page.Url); err != nil {
		return err
	}
	return proces.storage.Remove(page)
}

func (proces *Processor) sendHelp(chatID int) error {

	switch chatID {
	case MikeId:
		return proces.tg.SendMessage(chatID, msgHelpMike)
	case SoniaId:
		return proces.tg.SendMessage(chatID, msgHelpSonia)
	default:
		return proces.tg.SendMessage(chatID, msgHelp)
	}
}

func (proces *Processor) sendHello(chatID int) error {

	switch chatID {
	case MikeId:
		return proces.tg.SendMessage(chatID, msgHello)
	case SoniaId:
		return proces.tg.SendMessage(chatID, msgHelloSonia)
	default:
		return proces.tg.SendMessage(chatID, msgHello)
	}
}

func NewMessageSendler(chatID int, tg *tgClient.Client) func(string) error {
	return func(msg string) error {
		return tg.SendMessage(chatID, msg)
	}
}
