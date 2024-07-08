package telegram

import (
	"errors"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/events"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
	"tgBotRecommender/storage/files"
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
	ChatID   int
	Username string
}

func New(client *tgClient.Client, storage files.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (proces *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := proces.tg.Updates(proces.offset, limit)
	if err != nil {
		return nil, e.Wrap("enable to get event from tgClient", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, upd := range updates {
		res = append(res, event(upd))
	}

	proces.offset = updates[len(updates)-1].ID + 1
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

	if err := proces.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return e.Wrap(ErrProcessMsg, err)
	}
	return nil
}

func meta(event events.Event) (Meta, error) {
	result, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("cannot identify meta", errUnknownEventType)
	}
	return result, nil
}

func event(upd tgClient.Update) events.Event {
	updType := fetchType(upd)
	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: fetchText(upd),
		}
	}

	return res
}

func fetchText(upd tgClient.Update) string {
	if upd.Message != nil {

		return ""
	}
	return upd.Message.Text
}

func fetchType(upd tgClient.Update) events.Type {
	if upd.Message == nil {

		return events.Unknown
	}
	return events.Message
}
