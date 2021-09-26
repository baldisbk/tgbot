package tgapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"golang.org/x/xerrors"
)

const (
	TestCmd    = "getMe"
	SendCmd    = "sendMessage"
	ReceiveCmd = "getUpdates"
	AnswerCmd  = "answerCallbackQuery"
	EditCmd    = "editMessageText"
)

type Config struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}

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

func (m *Message) User() User                                              { return m.From }
func (m *Message) Message() interface{}                                    { return m }
func (m *Message) PreProcess(ctx context.Context, client *TGClient) error  { return nil }
func (m *Message) PostProcess(ctx context.Context, client *TGClient) error { return nil }

type CallbackQuery struct {
	Id           string `json:"id"`
	From         User   `json:"from"`
	ChatInstance string `json:"chat_instance"`
	Data         string `json:"data"`
}

func (m *CallbackQuery) User() User           { return m.From }
func (m *CallbackQuery) Message() interface{} { return m }
func (m *CallbackQuery) PreProcess(ctx context.Context, client *TGClient) error {
	return client.AnswerCallback(ctx, m.Id)
}
func (m *CallbackQuery) PostProcess(ctx context.Context, client *TGClient) error { return nil }

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

func makeCmd(address, token string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "bot"+token)
	return u.String() + "/", nil
}

func NewClient(ctx context.Context, cfg Config) (*TGClient, error) {
	cli := &TGClient{client: &http.Client{}}
	path, err := makeCmd(cfg.Address, cfg.Token)
	if err != nil {
		return nil, xerrors.Errorf("make url: %w", err)
	}
	cli.path = path
	if err := cli.Test(ctx); err != nil {
		return nil, xerrors.Errorf("test: %w", err)
	}
	return cli, nil
}

func (c *TGClient) Test(ctx context.Context) error {
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

func (c *TGClient) request(ctx context.Context, httpmethod, apimethod string, input interface{}, output interface{}) error {
	body, err := json.Marshal(input)
	req, err := http.NewRequest(httpmethod, c.path+apimethod, bytes.NewBuffer(body))
	if err != nil {
		return xerrors.Errorf("make req: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	// TODO middleware
	logging.S(ctx).Debugf("HTTP REQ: %s, %s, %s", req.URL.String(), req.Method, string(body))
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
	logging.S(ctx).Debugf("HTTP RSP: %s", string(body))
	if output == nil {
		return nil
	}
	err = json.Unmarshal(body, output)
	if err != nil {
		return xerrors.Errorf("parse: %w", err)
	}
	return nil
}

func (c *TGClient) GetUpdates(ctx context.Context) ([]Update, error) {
	var res UpdateResponse
	err := c.request(ctx, http.MethodGet, ReceiveCmd, GetUpdates{Offset: c.offset}, &res)
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

func (c *TGClient) EditMessage(ctx context.Context, chat uint64, text string, msgId uint64) (uint64, error) {
	var msg SendResponse
	var cmd = SendCmd
	if msgId != 0 {
		cmd = EditCmd
	}
	err := c.request(ctx,
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

func (c *TGClient) SendMessage(ctx context.Context, chat uint64, text string) (uint64, error) {
	return c.EditMessage(ctx, chat, text, 0)
}

func (c *TGClient) AnswerCallback(ctx context.Context, callbackId string) error {
	return c.request(ctx,
		http.MethodPost, AnswerCmd,
		AnswerCallback{
			CallbackQueryId: callbackId,
		}, nil)
}

func (c *TGClient) EditAnswerKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard AnswerKeyboard) (uint64, error) {
	var msg SendResponse
	var cmd = SendCmd
	if msgId != 0 {
		cmd = EditCmd
	}
	err := c.request(ctx,
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

func (c *TGClient) CreateAnswerKeyboard(ctx context.Context, chat uint64, text string, keyboard AnswerKeyboard) (uint64, error) {
	return c.EditAnswerKeyboard(ctx, chat, text, 0, keyboard)
}

func (c *TGClient) EditInputKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard InlineKeyboard) (uint64, error) {
	var msg SendResponse
	var cmd = SendCmd
	if msgId != 0 {
		cmd = EditCmd
	}
	err := c.request(ctx,
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

func (c *TGClient) CreateInputKeyboard(ctx context.Context, chat uint64, text string, keyboard InlineKeyboard) (uint64, error) {
	return c.EditInputKeyboard(ctx, chat, text, 0, keyboard)
}

func (c *TGClient) DropKeyboard(ctx context.Context, chat uint64, text string) error {
	return c.request(ctx,
		http.MethodPost, SendCmd,
		SendDropKeyboard{
			SendParams: SendParams{
				ChatId: chat,
				Text:   text,
			},
			ReplyMarkup: DropKeyboard{RemoveKeyboard: true},
		}, nil)
}
