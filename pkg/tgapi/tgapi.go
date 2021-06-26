package tgapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"golang.org/x/xerrors"
)

const BotToken = "1607353956:AAGbM0Sp4d56fXK6zbBm9o252PRO9ON-gx4"
const TgApi = "https://api.telegram.org/"
const (
	TestCmd    = "getMe"
	SendCmd    = "sendMessage"
	ReceiveCmd = "getUpdates"
	AnswerCmd  = "answerCallbackQuery"
	EditCmd    = "editMessageText"
)

// ======== Incoming updates ========

type User struct {
	Id           uint64 `json:"id"`
	FirstName    string `json:"first_name"`
	LanguageCode string `json:"language_code"`
}

type Chat struct {
	Id        uint64 `json:"id"`
	Type      string `json:"type"`
	FirstName string `json:"first_name"`
}

type Message struct {
	MessageId uint64 `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      uint64 `json:"date"`
	Text      string `json:"text"`
}

func (m *Message) User() User                         { return m.From }
func (m *Message) Message() interface{}               { return m }
func (m *Message) PreProcess(client *TGClient) error  { return nil }
func (m *Message) PostProcess(client *TGClient) error { return nil }

type CallbackQuery struct {
	Id           string `json:"id"`
	From         User   `json:"from"`
	ChatInstance string `json:"chat_instance"`
	Data         string `json:"data"`
}

func (m *CallbackQuery) User() User                         { return m.From }
func (m *CallbackQuery) Message() interface{}               { return m }
func (m *CallbackQuery) PreProcess(client *TGClient) error  { return client.AnswerCallback(m.Id) }
func (m *CallbackQuery) PostProcess(client *TGClient) error { return nil }

type Update struct {
	UpdateId      uint64         `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type UpdateResponse struct {
	Result []Update `json:"result"`
	Ok     bool     `json:"ok"`
}

type SendResponse struct {
	Result Message `json:"result"`
	Ok     bool    `json:"ok"`
}

// ======== Outgoing requests ========

// base outgoing message
type SendParams struct {
	ChatId    uint64 `json:"chat_id"`
	Text      string `json:"text"`
	MessageId uint64 `json:"message_id,omitempty"`
}

// keyboard with answers
type AnswerKeyboardButton struct {
	Text string `json:"text"`
}

type AnswerKeyboard struct {
	Keyboard [][]AnswerKeyboardButton `json:"keyboard"`
	OneTime  bool                     `json:"one_time"`
	Resize   bool                     `json:"resize"`
}

type SendAnswerKeyboard struct {
	SendParams
	ReplyMarkup AnswerKeyboard `json:"reply_markup"`
}

// keyboard with callbacks
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type InlineKeyboard struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type SendInlineKeyboard struct {
	SendParams
	ReplyMarkup InlineKeyboard `json:"reply_markup"`
}

// drop keyboard
type DropKeyboard struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
}

type SendDropKeyboard struct {
	SendParams
	ReplyMarkup DropKeyboard `json:"reply_markup"`
}

// set webhook
type SetWebhook struct {
	URL                string `json:"url"`
	Certificate        string `json:"certificate"`
	DropPendingUpdates bool   `json:"drop_pending_updates"`
}

// get updates
type GetUpdates struct {
	Offset uint64 `json:"offset"`
}

type AnswerCallback struct {
	CallbackQueryId string `json:"callback_query_id"`
	Text            string `json:"text"`
}

// ======== Client ========

type TGClient struct {
	client *http.Client
	path   string
	offset uint64
}

func MakeCmd(address, token string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "bot"+token)
	return u.String() + "/", nil
}

func NewClient(url string, token string) (*TGClient, error) {
	cli := &TGClient{client: &http.Client{}}
	path, err := MakeCmd(url, token)
	if err != nil {
		return nil, xerrors.Errorf("make url: %w", err)
	}
	cli.path = path
	if err := cli.Test(); err != nil {
		return nil, xerrors.Errorf("test: %w", err)
	}
	return cli, nil
}

func (c *TGClient) Test() error {
	req, err := http.NewRequest(http.MethodGet, c.path+TestCmd, nil)
	if err != nil {
		return xerrors.Errorf("make req: %w", err)
	}
	_, err = c.client.Do(req)
	if err != nil {
		return xerrors.Errorf("request: %w", err)
	}
	return nil
}

func (c *TGClient) request(httpmethod, apimethod string, input interface{}, output interface{}) error {
	body, err := json.Marshal(input)
	req, err := http.NewRequest(httpmethod, c.path+apimethod, bytes.NewBuffer(body))
	if err != nil {
		return xerrors.Errorf("make req: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	fmt.Printf("\t HTTP REQ: %s, %s, %s\n", req.URL.String(), req.Method, string(body))
	rsp, err := c.client.Do(req)
	if err != nil {
		return xerrors.Errorf("request: %w", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return xerrors.Errorf("http status: %d", rsp.StatusCode)
	}
	body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return xerrors.Errorf("read rsp: %w", err)
	}
	fmt.Printf("\t HTTP RSP: %s\n", string(body))
	if output == nil {
		return nil
	}
	err = json.Unmarshal(body, output)
	if err != nil {
		return xerrors.Errorf("parse: %w", err)
	}
	return nil
}

func (c *TGClient) GetUpdates() ([]Update, error) {
	var res UpdateResponse
	err := c.request(http.MethodGet, ReceiveCmd, GetUpdates{Offset: c.offset}, &res)
	if err != nil {
		return nil, xerrors.Errorf("request: %w", err)
	}
	for _, r := range res.Result {
		if c.offset <= r.UpdateId {
			c.offset = r.UpdateId + 1
		}
	}
	return res.Result, nil
}

func (c *TGClient) EditMessage(chat uint64, text string, msgId uint64) (uint64, error) {
	var msg SendResponse
	var cmd = SendCmd
	if msgId != 0 {
		cmd = EditCmd
	}
	err := c.request(
		http.MethodPost, cmd,
		SendParams{
			ChatId:    chat,
			Text:      text,
			MessageId: msgId, // 0 will be omitted
		}, &msg)
	if err != nil {
		return 0, err
	}
	return msg.Result.MessageId, nil
}

func (c *TGClient) SendMessage(chat uint64, text string) (uint64, error) {
	return c.EditMessage(chat, text, 0)
}

func (c *TGClient) AnswerCallback(callbackId string) error {
	return c.request(
		http.MethodPost, AnswerCmd,
		AnswerCallback{
			CallbackQueryId: callbackId,
		}, nil)
}

func (c *TGClient) EditAnswerKeyboard(chat uint64, text string, msgId uint64, keyboard AnswerKeyboard) (uint64, error) {
	var msg SendResponse
	var cmd = SendCmd
	if msgId != 0 {
		cmd = EditCmd
	}
	err := c.request(
		http.MethodPost, cmd,
		SendAnswerKeyboard{
			SendParams: SendParams{
				ChatId:    chat,
				Text:      text,
				MessageId: msgId,
			},
			ReplyMarkup: keyboard,
		}, &msg)
	if err != nil {
		return 0, err
	}
	return msg.Result.MessageId, nil
}

func (c *TGClient) CreateAnswerKeyboard(chat uint64, text string, keyboard AnswerKeyboard) (uint64, error) {
	return c.EditAnswerKeyboard(chat, text, 0, keyboard)
}

func (c *TGClient) EditInputKeyboard(chat uint64, text string, msgId uint64, keyboard InlineKeyboard) (uint64, error) {
	var msg SendResponse
	var cmd = SendCmd
	if msgId != 0 {
		cmd = EditCmd
	}
	err := c.request(
		http.MethodPost, cmd,
		SendInlineKeyboard{
			SendParams: SendParams{
				ChatId:    chat,
				Text:      text,
				MessageId: msgId,
			},
			ReplyMarkup: keyboard,
		}, &msg)
	if err != nil {
		return 0, err
	}
	return msg.Result.MessageId, nil
}

func (c *TGClient) CreateInputKeyboard(chat uint64, text string, keyboard InlineKeyboard) (uint64, error) {
	return c.EditInputKeyboard(chat, text, 0, keyboard)
}

func (c *TGClient) DropKeyboard(chat uint64, text string) error {
	return c.request(
		http.MethodPost, SendCmd,
		SendDropKeyboard{
			SendParams: SendParams{
				ChatId: chat,
				Text:   text,
			},
			ReplyMarkup: DropKeyboard{RemoveKeyboard: true},
		}, nil)
}
