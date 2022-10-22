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

// ======== Client ========

type tgClient struct {
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

func NewClient(ctx context.Context, cfg Config) (*tgClient, error) {
	cli := &tgClient{client: &http.Client{}}
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

func (c *tgClient) Test(ctx context.Context) error {
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

func (c *tgClient) request(ctx context.Context, httpmethod, apimethod string, input interface{}, output interface{}) error {
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

func (c *tgClient) GetUpdates(ctx context.Context) ([]Update, error) {
	var res UpdateResponse
	err := c.request(ctx, http.MethodGet, ReceiveCmd, GetUpdates{Offset: c.offset}, &res)
	if err != nil {
		return nil, xerrors.Errorf("request: %w", err)
	}
	for _, r := range res.Result {
		if c.offset <= r.UpdateId {
			c.offset = r.UpdateId + 1
		}
		hash(r)
	}
	return res.Result, nil
}

func (c *tgClient) EditMessage(ctx context.Context, chat uint64, text string, msgId uint64) (uint64, error) {
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

func (c *tgClient) SendMessage(ctx context.Context, chat uint64, text string) (uint64, error) {
	return c.EditMessage(ctx, chat, text, 0)
}

func (c *tgClient) AnswerCallback(ctx context.Context, callbackId string) error {
	return c.request(ctx,
		http.MethodPost, AnswerCmd,
		AnswerCallback{
			CallbackQueryId: callbackId,
		}, nil)
}

func (c *tgClient) EditAnswerKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard AnswerKeyboard) (uint64, error) {
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

func (c *tgClient) CreateAnswerKeyboard(ctx context.Context, chat uint64, text string, keyboard AnswerKeyboard) (uint64, error) {
	return c.EditAnswerKeyboard(ctx, chat, text, 0, keyboard)
}

func (c *tgClient) EditInputKeyboard(ctx context.Context, chat uint64, text string, msgId uint64, keyboard InlineKeyboard) (uint64, error) {
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

func (c *tgClient) CreateInputKeyboard(ctx context.Context, chat uint64, text string, keyboard InlineKeyboard) (uint64, error) {
	return c.EditInputKeyboard(ctx, chat, text, 0, keyboard)
}

func (c *tgClient) DropKeyboard(ctx context.Context, chat uint64, text string) error {
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
