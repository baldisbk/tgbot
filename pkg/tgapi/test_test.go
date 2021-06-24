package tgapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeCmd(t *testing.T) {
	a := assert.New(t)
	s, err := MakeCmd("http://tg.api", "token:token")
	a.NoError(err)
	a.Equal("http://tg.api/bottoken:token/", s)
}
