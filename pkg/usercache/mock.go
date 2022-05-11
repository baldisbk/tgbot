package usercache

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
)

// ======== User mock ========

type userMock struct {
	mock.Mock
}

func NewUserMock() *userMock {
	return &userMock{}
}

func (u *userMock) UpdateState(ctx context.Context, value interface{}) error {
	args := u.Called(ctx, value)
	return args.Error(0)
}

func (u *userMock) Run(ctx context.Context, input interface{}) (interface{}, error) {
	args := u.Called(ctx, input)
	return args[0], args.Error(1)
}

// ======== Cache mock ========

type cacheMock struct {
	mock.Mock
}

func NewCacheMock() *cacheMock {
	return &cacheMock{}
}

func (u *cacheMock) Get(ctx context.Context, user tgapi.User) (User, error) {
	args := u.Called(ctx, user)
	return args[0].(User), args.Error(1)
}

func (u *cacheMock) Put(ctx context.Context, tgUser tgapi.User, user User) error {
	args := u.Called(ctx, tgUser, user)
	return args.Error(0)
}

func (u *cacheMock) Close() {
	u.Called()
}
