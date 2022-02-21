package statemachine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	testCases := []struct {
		desc        string
		transitions []Transition
		startState  string

		input  interface{}
		output interface{}
		err    error
	}{
		{
			desc: "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert := require.New(t)
			sm := NewSM(tC.startState, tC.transitions)
			output, err := sm.Run(context.Background(), tC.input)
			if tC.err != nil {
				assert.NoError(err)
				assert.Equal(tC.output, output)
			} else {
				assert.Equal(tC.err, err)
			}
		})
	}
}
