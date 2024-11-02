package httputils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/baldisbk/tgbot/pkg/logging"
)

type BaseClient struct {
	Client *http.Client
	Path   string
}

func (c *BaseClient) Request(ctx context.Context, httpmethod, apimethod string, input interface{}, output interface{}) error {
	body, err := json.Marshal(input)
	if err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, httpmethod, c.Path+apimethod, bytes.NewBuffer(body))
	if err != nil {
		return xerrors.Errorf("make req: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	// TODO middleware
	logging.S(ctx).Debugf("HTTP REQ: %s, %s, %s", req.URL.String(), req.Method, string(body))
	rsp, err := c.Client.Do(req)
	if err != nil {
		return xerrors.Errorf("request: %w", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return xerrors.Errorf("http status: %d", rsp.StatusCode)
	}
	body, err = io.ReadAll(rsp.Body)
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
