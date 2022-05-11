package tgapi

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type tgMock struct {
	mock.Mock
}

func NewMock() *tgMock {
	return &tgMock{}
}

func (tg *tgMock) Test(ctx context.Context) error {
	args := tg.Called(ctx)
	return args.Error(0)
}

func (tg *tgMock) GetUpdates(ctx context.Context) ([]Update, error) {
	args := tg.Called(ctx)
	return args[0].([]Update), args.Error(1)
}

func (tg *tgMock) EditMessage(ctx context.Context, chat uint64, text string, msgId uint64) (uint64, error) {
	args := tg.Called(ctx, chat, text, msgId)
	return args[0].(uint64), args.Error(1)
}

func (tg *tgMock) SendMessage(ctx context.Context, chat uint64, text string) (uint64, error) {
	args := tg.Called(ctx, chat, text)
	return args[0].(uint64), args.Error(1)
}

func (tg *tgMock) AnswerCallback(ctx context.Context, callbackId string) error {
	args := tg.Called(ctx, callbackId)
	return args.Error(0)
}

func (tg *tgMock) EditAnswerKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard AnswerKeyboard) (uint64, error) {
	args := tg.Called(ctx, chat, text, msgId, keyboard)
	return args[0].(uint64), args.Error(1)
}

func (tg *tgMock) CreateAnswerKeyboard(ctx context.Context, chat uint64, text string, keyboard AnswerKeyboard) (uint64, error) {
	args := tg.Called(ctx, chat, text, keyboard)
	return args[0].(uint64), args.Error(1)
}

func (tg *tgMock) EditInputKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard InlineKeyboard) (uint64, error) {
	args := tg.Called(ctx, chat, text, msgId, keyboard)
	return args[0].(uint64), args.Error(1)
}

func (tg *tgMock) CreateInputKeyboard(ctx context.Context, chat uint64, text string, keyboard InlineKeyboard) (uint64, error) {
	args := tg.Called(ctx, chat, text, keyboard)
	return args[0].(uint64), args.Error(1)
}

func (tg *tgMock) DropKeyboard(ctx context.Context, chat uint64, text string) error {
	args := tg.Called(ctx, chat, text)
	return args.Error(0)
}
