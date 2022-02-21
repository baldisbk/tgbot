package tgapi

import "context"

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

type TGClient interface {
	Test(ctx context.Context) error
	GetUpdates(ctx context.Context) ([]Update, error)
	EditMessage(ctx context.Context, chat uint64, text string, msgId uint64) (uint64, error)
	SendMessage(ctx context.Context, chat uint64, text string) (uint64, error)
	AnswerCallback(ctx context.Context, callbackId string) error
	EditAnswerKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard AnswerKeyboard) (uint64, error)
	CreateAnswerKeyboard(ctx context.Context, chat uint64, text string, keyboard AnswerKeyboard) (uint64, error)
	EditInputKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard InlineKeyboard) (uint64, error)
	CreateInputKeyboard(ctx context.Context, chat uint64, text string, keyboard InlineKeyboard) (uint64, error)
	DropKeyboard(ctx context.Context, chat uint64, text string) error
}
