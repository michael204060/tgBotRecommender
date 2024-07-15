package tgClient

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID int              `json:"update_id"` // Изменено имя поля на "update_id"
	Message  *IncomingMessage `json:"message"`
}

type IncomingMessage struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}

type Chat struct {
	ID int `json:"id"`
}
