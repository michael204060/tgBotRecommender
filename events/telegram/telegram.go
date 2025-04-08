package telegram

import (
	"errors"
	"log"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/events"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
)

var (
	errUnknownEventType = errors.New("unknown event type")
	errUnknownMetaType  = errors.New("unknown meta type")
)

type Processor struct {
	tg      *tgClient.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID   int
	UserID   int
	IsButton bool
	Data     string
}

func New(client *tgClient.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
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

	p.offset = updates[len(updates)-1].UpdateID + 1
	log.Printf("Updated offset to: %d", p.offset)
	return res, nil
}

func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	case events.Callback:
		return p.processCallback(event)
	default:
		return e.Wrap("failed to process message", errUnknownEventType)
	}
}

func (p *Processor) processMessage(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("failed to get meta", err)
	}

	if meta.IsButton {
		return p.handleCallback(meta.ChatID, meta.UserID, meta.Data)
	}

	return p.doCmd(event.Text, meta.ChatID, meta.UserID)
}

func (p *Processor) processCallback(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("failed to get meta", err)
	}

	return p.handleCallback(meta.ChatID, meta.UserID, meta.Data)
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("cannot get meta", errUnknownMetaType)
	}
	return res, nil
}

func event(upd tgClient.Update) events.Event {
	if upd.CallbackQuery != nil {
		return events.Event{
			Type: events.Callback,
			Text: upd.CallbackQuery.Data,
			Meta: Meta{
				ChatID:   upd.CallbackQuery.Message.Chat.ID,
				UserID:   upd.CallbackQuery.From.ID,
				IsButton: true,
				Data:     upd.CallbackQuery.Data,
			},
		}
	}

	if upd.Message == nil {
		return events.Event{
			Type: events.Unknown,
		}
	}

	return events.Event{
		Type: events.Message,
		Text: upd.Message.Text,
		Meta: Meta{
			ChatID: upd.Message.Chat.ID,
			UserID: upd.Message.From.ID,
		},
	}
}
