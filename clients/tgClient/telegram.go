package tgClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"tgBotRecommender/lib/e"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

const (
	getUpdatesMethod  = "getUpdates"
	sendMessageMethod = "sendMessage"
)

func New(host string, token string) *Client {
	return &Client{
		host:     host,
		basePath: newBasePath(token),
		client:   http.Client{},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

func (client *Client) Updates(offset int, limit int) ([]Update, error) {
	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := client.doRequest(getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	var result UpdatesResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Result, nil
}

func (client *Client) SendMessage(chatId int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatId))
	q.Add("text", text)

	_, err := client.doRequest(sendMessageMethod, q)
	if err != nil {
		return e.Wrap("cannot send a massage", err)
	}

	return nil
}

func (client *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfError("can't do the request", err) }()
	u := url.URL{
		Scheme: "https",
		Host:   client.host,
		Path:   path.Join(client.basePath, method),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func (client *Client) SendMessageWithKeyboard(chatID int, text string, keyboard [][]string) error {
	buttons := make([][]map[string]interface{}, len(keyboard))
	for i, row := range keyboard {
		buttons[i] = make([]map[string]interface{}, len(row))
		for j, btn := range row {
			buttons[i][j] = map[string]interface{}{
				"text": btn,
			}
		}
	}

	data := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
		"reply_markup": map[string]interface{}{
			"keyboard":          buttons,
			"resize_keyboard":   true,
			"one_time_keyboard": true,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("https://%s/%s/sendMessage", client.host, client.basePath),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}

func (client *Client) SendInlineKeyboard(chatID int, text string, buttons []InlineButton) error {
	keyboard := make([][]map[string]interface{}, len(buttons))
	for i, btn := range buttons {
		keyboard[i] = []map[string]interface{}{{
			"text":          btn.Text,
			"callback_data": btn.Data,
		}}
	}

	data := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
		"reply_markup": map[string]interface{}{
			"inline_keyboard": keyboard,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("https://%s/%s/sendMessage", client.host, client.basePath),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

type InlineButton struct {
	Text string
	Data string
}
