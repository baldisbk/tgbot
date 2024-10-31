package tgmock

import (
	"context"
	"net/http"
	"net/url"

	"github.com/baldisbk/tgbot/pkg/httputils"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"golang.org/x/xerrors"
)

type Client struct {
	httputils.BaseClient
}

type ClientConfig struct {
	Address string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	addr, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, xerrors.Errorf("parse address: %w", err)
	}
	return &Client{BaseClient: httputils.BaseClient{
		Client: &http.Client{},
		Path:   addr.String(),
	}}, nil
}

func (c *Client) SendMessage(ctx context.Context, userID uint64, message string) error {
	req := PrivateRequest{
		UserID:  userID,
		Payload: message,
	}
	err := c.Request(ctx, http.MethodPut, privateMessagePath, req, nil)
	if err != nil {
		return xerrors.Errorf("request: %w", err)
	}
	return nil
}

func (c *Client) PushButton(ctx context.Context, userID uint64, data string) error {
	req := PrivateRequest{
		UserID:  userID,
		Payload: data,
	}
	err := c.Request(ctx, http.MethodPut, privateButtonPath, req, nil)
	if err != nil {
		return xerrors.Errorf("request: %w", err)
	}
	return nil
}

func (c *Client) GetHistory(ctx context.Context, userID uint64) ([]tgapi.Update, error) {
	req := PrivateRequest{
		UserID: userID,
	}
	var res []tgapi.Update
	err := c.Request(ctx, http.MethodGet, privateHistoryPath, req, &res)
	if err != nil {
		return nil, xerrors.Errorf("request: %w", err)
	}
	return res, nil
}
