package tgClient

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID      int              `json:"update_id"`
	Message       *IncomingMessage `json:"message"`
	CallbackQuery *CallbackQuery   `json:"callback_query"`
}

type IncomingMessage struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	From User   `json:"from"`
}

type CallbackQuery struct {
	ID      string           `json:"id"`
	From    User             `json:"from"`
	Message *IncomingMessage `json:"message"`
	Data    string           `json:"data"`
}

type Chat struct {
	ID int `json:"id"`
}

type User struct {
	ID int `json:"id"`
}
