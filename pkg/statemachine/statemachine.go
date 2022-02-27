package statemachine

import (
	"context"

	"github.com/baldisbk/tgbot_sample/pkg/logging"
)

type SMPredicate func(context.Context, string, interface{}) bool
type SMCallback func(context.Context, interface{}) (interface{}, error)

func EmptyPredicate(context.Context, string, interface{}) bool                  { return true }
func NotNilPredicate(ctx context.Context, state string, input interface{}) bool { return input != nil }
func EmptyCallback(ctx context.Context, input interface{}) (interface{}, error) { return input, nil }

func CompositeCallback(callbacks ...SMCallback) SMCallback {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		var arg = input
		var err error
		for _, callback := range callbacks {
			arg, err = callback(ctx, arg)
			if err != nil {
				return arg, err
			}
		}
		return arg, nil
	}
}

func ConstCallback(output interface{}) SMCallback {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		return output, nil
	}
}

type Transition struct {
	Source      string
	Destination string
	Predicate   SMPredicate
	Callback    SMCallback
}

type Machine interface {
	Run(ctx context.Context, input interface{}) (interface{}, error)
}

type sm struct {
	transitions map[string][]Transition
	state       string
}

func (s *sm) Run(ctx context.Context, input interface{}) (interface{}, error) {
	for {
		stateCtx := logging.WithTag(ctx, "STATE", s.state)
		logging.S(stateCtx).Infof("Received input %#v", input)
		trs, ok := s.transitions[s.state]
		if !ok {
			logging.S(stateCtx).Debugf("No transitions found")
			return input, nil
		}
		found := false
		for _, tr := range trs {
			logging.S(stateCtx).Debugf("Found transition")
			if tr.Predicate == nil || tr.Predicate(ctx, s.state, input) {
				logging.S(stateCtx).Debugf("Predicate ok")
				if tr.Callback != nil {
					res, err := tr.Callback(ctx, input)
					if err != nil {
						logging.S(stateCtx).Warnf("Callback returned error: %#v", err)
						return input, err
					}
					logging.S(stateCtx).Infof("Callback returned result: %#v", res)
					input = res
				} else {
					logging.S(stateCtx).Debugf("No callback")
				}
				found = true
				s.state = tr.Destination
				break
			}
		}
		if !found {
			logging.S(stateCtx).Debugf("No relevant transitions found, stop")
			return input, nil
		}
		logging.S(stateCtx).Debugf("Switch to state %s", s.state)
	}
}

func NewSM(state string, trs []Transition) *sm {
	sm := &sm{
		state:       state,
		transitions: map[string][]Transition{},
	}
	for _, tr := range trs {
		sm.transitions[tr.Source] = append(sm.transitions[tr.Source], tr)
	}
	return sm
}
