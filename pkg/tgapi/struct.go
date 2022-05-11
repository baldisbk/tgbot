package tgapi

import "context"

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

func (m *Message) User() User                                             { return m.From }
func (m *Message) Message() interface{}                                   { return m }
func (m *Message) PreProcess(ctx context.Context, client TGClient) error  { return nil }
func (m *Message) PostProcess(ctx context.Context, client TGClient) error { return nil }

type CallbackQuery struct {
	Id           string `json:"id"`
	From         User   `json:"from"`
	ChatInstance string `json:"chat_instance"`
	Data         string `json:"data"`
}

func (m *CallbackQuery) User() User           { return m.From }
func (m *CallbackQuery) Message() interface{} { return m }
func (m *CallbackQuery) PreProcess(ctx context.Context, client TGClient) error {
	return client.AnswerCallback(ctx, m.Id)
}
func (m *CallbackQuery) PostProcess(ctx context.Context, client TGClient) error { return nil }

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
