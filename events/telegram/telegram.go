package telegram

import (
	"errors"
	"fmt"
	"log"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/events"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
)

var (
	errUnknownEventType = errors.New("Unknown event type")
	errUnknownMetaType  = errors.New("Unknown meta type")
)

const ErrProcessMsg = "failed to process message"

type Processor struct {
	tg      *tgClient.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID int
}

func New(client *tgClient.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (proces *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := proces.tg.Updates(proces.offset, limit)
	if err != nil {
		return nil, e.Wrap("unable to get event from tgClient", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, upd := range updates {
		res = append(res, event(upd))
	}

	// Update the offset to the ID of the last update fetched
	proces.offset = updates[len(updates)-1].UpdateID + 1
	log.Printf("Updated offset to: %d", proces.offset) // Добавлено логирование
	return res, nil
}

func (proces *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return proces.processMessage(event)
	default:
		return e.Wrap(ErrProcessMsg, errUnknownEventType)
	}
}

func (proces *Processor) processMessage(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap(ErrProcessMsg, err)
	}

	err, info := proces.doCmd(event.Text, meta.ChatID)
	if err != nil {
		return e.Wrap(ErrProcessMsg, err)
	}
	//if info == events.Content {
	//	if err := proces.setPriority(event.Text, meta.ChatID); err != nil {
	//		return e.Wrap(ErrProcessMsg, err)
	//	}
	//}
	fmt.Print(info)
	return nil
}

func meta(event events.Event) (Meta, error) {
	result, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("cannot identify meta", errUnknownMetaType)
	}
	return result, nil
}

func event(upd tgClient.Update) events.Event {
	updType := fetchType(upd)
	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message && upd.Message != nil {
		res.Meta = Meta{
			ChatID: upd.Message.Chat.ID,
		}
	}

	return res
}

func fetchText(upd tgClient.Update) string {
	if upd.Message != nil {
		return upd.Message.Text
	}
	return ""
}

func fetchType(upd tgClient.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}
	return events.Message
}
