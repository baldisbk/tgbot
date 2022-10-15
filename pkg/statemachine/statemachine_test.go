package statemachine

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

var testError = xerrors.New("test error")

func concatCallback(append string) SMCallback {
	return func(c context.Context, i interface{}) (interface{}, error) {
		return fmt.Sprint(i, append), nil
	}
}

func TestSM(t *testing.T) {
	testCases := []struct {
		desc        string
		transitions []Transition
		startState  string

		finalState string
		input      interface{}
		output     interface{}
		err        error
	}{
		{
			desc:        "empty",
			transitions: []Transition{},
			startState:  "start",
			finalState:  "start",
			input:       "A",
			output:      "A",
		},
		{
			desc: "single",
			transitions: []Transition{{
				Source:      "start",
				Destination: "finish",
				Predicate:   EmptyPredicate,
				Callback:    ConstCallback("B"),
			}},
			startState: "start",
			finalState: "finish",
			input:      "A",
			output:     "B",
		},
		{
			desc: "composite",
			transitions: []Transition{{
				Source:      "start",
				Destination: "finish",
				Predicate:   EmptyPredicate,
				Callback: CompositeCallback(
					ConstCallback("B"),
					concatCallback("C"),
				),
			}},
			startState: "start",
			finalState: "finish",
			input:      "A",
			output:     "BC",
		},
		{
			desc: "first served",
			transitions: []Transition{
				{
					Source:      "start",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("B"),
				},
				{
					Source:      "start",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("C"),
				},
			},
			startState: "start",
			finalState: "finish",
			input:      "A",
			output:     "B",
		},
		{
			desc: "source",
			transitions: []Transition{
				{
					Source:      "intermediate",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("C"),
				},
				{
					Source:      "start",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("B"),
				},
			},
			startState: "start",
			finalState: "finish",
			input:      "A",
			output:     "B",
		},
		{
			desc: "multiple",
			transitions: []Transition{
				{
					Source:      "intermediate",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("C"),
				},
				{
					Source:      "start",
					Destination: "intermediate",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("B"),
				},
			},
			startState: "start",
			finalState: "finish",
			input:      "A",
			output:     "C",
		},
		{
			desc: "predicate value 1",
			transitions: []Transition{
				{
					Source:      "start",
					Destination: "finish",
					Predicate: func(c context.Context, s string, i interface{}) bool {
						if s, ok := i.(string); ok && s == "A" {
							return true
						}
						return false
					},
					Callback: ConstCallback("B"),
				},
				{
					Source:      "start",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("C"),
				},
			},
			startState: "start",
			finalState: "finish",
			input:      "A",
			output:     "B",
		},
		{
			desc: "predicate value 2",
			transitions: []Transition{
				{
					Source:      "start",
					Destination: "finish",
					Predicate: func(c context.Context, s string, i interface{}) bool {
						if s, ok := i.(string); ok && s == "A" {
							return true
						}
						return false
					},
					Callback: ConstCallback("B"),
				},
				{
					Source:      "start",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback:    ConstCallback("C"),
				},
			},
			startState: "start",
			finalState: "finish",
			input:      "B",
			output:     "C",
		},
		{
			desc: "predicate state 1",
			transitions: []Transition{
				{
					Source:      "start",
					Destination: "finish",
					Predicate: func(c context.Context, s string, i interface{}) bool {
						return s == "start"
					},
					Callback: ConstCallback("B"),
				},
			},
			startState: "start",
			finalState: "finish",
			input:      "A",
			output:     "B",
		},
		{
			desc: "predicate state 2",
			transitions: []Transition{
				{
					Source:      "not start",
					Destination: "finish",
					Predicate: func(c context.Context, s string, i interface{}) bool {
						return s == "start"
					},
					Callback: ConstCallback("B"),
				},
			},
			startState: "not start",
			finalState: "not start",
			input:      "A",
			output:     "A",
		},
		{
			desc: "error",
			transitions: []Transition{
				{
					Source:      "start",
					Destination: "finish",
					Predicate:   EmptyPredicate,
					Callback: func(c context.Context, i interface{}) (interface{}, error) {
						return nil, testError
					},
				},
			},
			startState: "start",
			finalState: "start",
			input:      "A",
			output:     "A",
			err:        testError,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert := require.New(t)
			sm := NewSM(tC.startState, tC.transitions)
			output, err := sm.Run(context.Background(), tC.input)
			if tC.err == nil {
				assert.NoError(err)
				assert.Equal(tC.output, output)
				assert.Equal(tC.finalState, sm.state)
			} else {
				assert.Equal(tC.err, err)
			}
		})
	}
}
