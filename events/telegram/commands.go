package telegram

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"tgBotRecommender/clients/tgClient"
	"tgBotRecommender/lib/e"
	"tgBotRecommender/storage"
	"tgBotRecommender/storage/database"
)

type UserState struct {
	WaitingForPriority bool
	CurrentMessage     string
	CurrentPriority    int
}

var userStates = make(map[int]*UserState)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

const (
	DeleteAction = "delete"
	KeepAction   = "keep"
	LowerAction  = "lower"
	DeletePrefix = "delete_"
	KeepPrefix   = "keep_"
	LowerPrefix  = "lower_"
)

func (p *Processor) doCmd(text string, chatID int, userID int) error {
	text = strings.TrimSpace(text)
	log.Printf("got command: %s from %d", text, chatID)

	if state, exists := userStates[userID]; exists && state.WaitingForPriority {
		return p.handlePriorityInput(text, userID, chatID)
	}

	switch text {
	case RndCmd:
		return p.sendHighestPriorityMessage(chatID, userID)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		userStates[userID] = &UserState{
			WaitingForPriority: true,
			CurrentMessage:     text,
		}
		return p.tg.SendMessage(chatID, "Пожалуйста, укажите приоритет этого сообщения (уникальное положительное число):")
	}
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, `Я могу сохранять ваши сообщения с приоритетами.
Чтобы сохранить сообщение, просто отправьте его, а затем укажите уникальный приоритет (число).
Чтобы получить сообщение с наивысшим приоритетом, используйте команду /rnd.
После просмотра вы сможете удалить сообщение или изменить его приоритет.`)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, "Привет! 👋\n\n"+`Я могу сохранять ваши сообщения с приоритетами.
Чтобы сохранить сообщение, просто отправьте его, а затем укажите уникальный приоритет (число).
Чтобы получить сообщение с наивысшим приоритетом, используйте команду /rnd.
После просмотра вы сможете удалить сообщение или изменить его приоритет.`)
}

func (p *Processor) handlePriorityInput(text string, userID int, chatID int) error {
	priority, err := strconv.Atoi(text)
	if err != nil || priority <= 0 {
		return p.tg.SendMessage(chatID, "Неверный формат приоритета. Введите положительное целое число:")
	}

	state := userStates[userID]
	delete(userStates, userID)

	db, err := database.HandleConn()
	if err != nil {
		return e.Wrap("failed to connect to database", err)
	}
	defer db.Close()

	exists, err := p.storage.IsPriorityExists(userID, priority, db)
	if err != nil {
		return e.Wrap("failed to check priority", err)
	}
	if exists {
		p.tg.SendMessage(chatID, fmt.Sprintf("Приоритет %d уже существует. Введите другой уникальный приоритет:", priority))
		userStates[userID] = state
		return nil
	}

	message := &storage.Message{
		Content:  state.CurrentMessage,
		UserID:   userID,
		Priority: priority,
	}

	if err := p.storage.SaveWithPriority(message, db); err != nil {
		return e.Wrap("failed to save message", err)
	}

	if err := p.storage.NormalizePriorities(userID, db); err != nil {
		return e.Wrap("failed to normalize priorities", err)
	}

	return p.tg.SendMessage(chatID, fmt.Sprintf("Сообщение сохранено с приоритетом %d", priority))
}

func (p *Processor) sendHighestPriorityMessage(chatID int, userID int) error {
	db, err := database.HandleConn()
	if err != nil {
		return e.Wrap("failed to connect to database", err)
	}
	defer db.Close()

	message, err := p.storage.PickHighestPriority(userID, db)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedMessages) {
			return p.tg.SendMessage(chatID, "Нет сохраненных сообщений")
		}
		return e.Wrap("failed to pick highest priority message", err)
	}

	// Создаем inline-кнопки
	buttons := []tgClient.InlineButton{
		{Text: "Удалить", Data: DeletePrefix + strconv.Itoa(message.Index)},
		{Text: "Оставить", Data: KeepPrefix + strconv.Itoa(message.Index)},
	}

	msgText := fmt.Sprintf("Сообщение с наивысшим приоритетом (%d):\n\n%s", message.Message.Priority, message.Message.Content)
	return p.tg.SendInlineKeyboard(chatID, msgText, buttons)
}

func (p *Processor) handleCallback(chatID int, userID int, callbackData string) error {
	db, err := database.HandleConn()
	if err != nil {
		return e.Wrap("failed to connect to database", err)
	}
	defer db.Close()

	var messageID int
	var action string

	switch {
	case strings.HasPrefix(callbackData, DeletePrefix):
		messageID, _ = strconv.Atoi(strings.TrimPrefix(callbackData, DeletePrefix))
		action = DeleteAction
	case strings.HasPrefix(callbackData, KeepPrefix):
		messageID, _ = strconv.Atoi(strings.TrimPrefix(callbackData, KeepPrefix))
		action = KeepAction
	case strings.HasPrefix(callbackData, LowerPrefix):
		messageID, _ = strconv.Atoi(strings.TrimPrefix(callbackData, LowerPrefix))
		action = LowerAction
	default:
		return fmt.Errorf("unknown callback data: %s", callbackData)
	}

	switch action {
	case DeleteAction:
		if err := p.storage.RemoveByMessageID(messageID, db); err != nil {
			return e.Wrap("failed to remove message", err)
		}
		if err := p.storage.NormalizePriorities(userID, db); err != nil {
			return e.Wrap("failed to normalize priorities", err)
		}
		return p.tg.SendMessage(chatID, "Сообщение удалено. Приоритеты обновлены.")

	case KeepAction:
		// Предлагаем оставить или понизить приоритет
		buttons := []tgClient.InlineButton{
			{Text: "Оставить текущий", Data: KeepPrefix + strconv.Itoa(messageID)},
			{Text: "Понизить приоритет", Data: LowerPrefix + strconv.Itoa(messageID)},
		}
		return p.tg.SendInlineKeyboard(chatID, "Выберите действие с приоритетом:", buttons)

	case LowerAction:
		if err := p.storage.LowerPriority(messageID, userID, db); err != nil {
			return e.Wrap("failed to lower priority", err)
		}
		if err := p.storage.NormalizePriorities(userID, db); err != nil {
			return e.Wrap("failed to normalize priorities", err)
		}
		return p.tg.SendMessage(chatID, "Приоритет понижен. Все приоритеты обновлены.")
	}

	return nil
}
