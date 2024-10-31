package engine

import (
	"context"

	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/stretchr/testify/mock"
)

// ======== Engine mock ========

type engineMock struct {
	mock.Mock
}

func NewEngineMock() *engineMock {
	return &engineMock{}
}

func (u *engineMock) Receive(ctx context.Context, signal Signal) error {
	return u.Called(ctx, signal).Error(0)
}

// ======== Signal mock ========

type signalMock struct {
	mock.Mock
}

func NewSignalMock() *signalMock {
	return &signalMock{}
}

func (u *signalMock) User() tgapi.User {
	return u.Called()[0].(tgapi.User)
}

func (u *signalMock) Message() interface{} {
	return u.Called()[0]
}

func (u *signalMock) PreProcess(ctx context.Context, client tgapi.TGClient) error {
	return u.Called(ctx, client).Error(0)
}

func (u *signalMock) PostProcess(ctx context.Context, client tgapi.TGClient) error {
	return u.Called(ctx, client).Error(0)
}
