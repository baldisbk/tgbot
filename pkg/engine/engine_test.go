package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/baldisbk/tgbot/pkg/usercache"
)

func TestEngine(t *testing.T) {
	preErr := xerrors.New("preprocess")
	postErr := xerrors.New("postprocess")
	runErr := xerrors.New("run")
	saveErr := xerrors.New("update")
	getErr := xerrors.New("get")
	putErr := xerrors.New("put")

	testCases := []struct {
		desc            string
		req, rsp        interface{}
		preErr, postErr error
		runErr, saveErr error
		getErr, putErr  error
		expErr          error
	}{
		{
			desc: "ok",
			req:  "A",
			rsp:  "B",
		},
		{
			desc:   "pre-err",
			req:    "A",
			preErr: preErr,
			expErr: preErr,
		},
		{
			desc:    "post-err",
			req:     "A",
			postErr: postErr,
			expErr:  postErr,
		},
		{
			desc:   "run-err",
			req:    "A",
			runErr: runErr,
			expErr: runErr,
		},
		{
			desc:    "save-err",
			req:     "A",
			saveErr: saveErr,
			expErr:  saveErr,
		},
		{
			desc:   "get-err",
			req:    "A",
			getErr: getErr,
			expErr: getErr,
		},
		{
			desc:   "put-err",
			req:    "A",
			putErr: putErr,
			expErr: putErr,
		},
	}
	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			assert := require.New(t)

			client := tgapi.NewMock()

			user := usercache.NewUserMock()
			user.On("Run", mock.Anything, c.req).Return(c.rsp, c.runErr)
			user.On("UpdateState", mock.Anything, c.rsp).Return(c.saveErr)

			cache := usercache.NewCacheMock()
			cache.On("Get", mock.Anything, tgapi.User{}).Return(user, c.getErr)
			cache.On("Put", mock.Anything, tgapi.User{}, user).Return(c.putErr)

			signal := NewSignalMock()
			signal.On("User").Return(tgapi.User{})
			signal.On("Message").Return(c.req)
			signal.On("PreProcess", mock.Anything, client).Return(c.preErr)
			signal.On("PostProcess", mock.Anything, client).Return(c.postErr)

			engine := NewEngine(client, cache)
			err := engine.Receive(context.Background(), signal)

			if c.expErr == nil {
				assert.NoError(err)
			} else {
				assert.True(xerrors.Is(err, c.expErr))
			}
		})
	}
}
