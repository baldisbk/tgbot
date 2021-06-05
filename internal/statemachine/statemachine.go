package statemachine

import "fmt"

type SMPredicate func(string, interface{}) bool
type SMCallback func(interface{}) (interface{}, error)

func EmptyPredicate(string, interface{}) bool              { return true }
func EmptyCallback(input interface{}) (interface{}, error) { return input, nil }

type Transition struct {
	Source      string
	Destination string
	Predicate   SMPredicate
	Callback    SMCallback
}

type Machine interface {
	Run(input interface{}) (interface{}, error)
}

type sm struct {
	transitions map[string][]Transition
	state       string
}

func (s *sm) Run(input interface{}) (interface{}, error) {
	for {
		fmt.Printf("State %s, Received %#v\n", s.state, input)
		trs, ok := s.transitions[s.state]
		if !ok {
			fmt.Printf("No transitions found for %s\n", s.state)
			return input, nil
		}
		found := false
		for _, tr := range trs {
			fmt.Printf("Found transition for %s\n", s.state)
			if tr.Predicate == nil || tr.Predicate(s.state, input) {
				fmt.Printf("Predicate ok\n")
				if tr.Callback != nil {
					res, err := tr.Callback(input)
					if err != nil {
						fmt.Printf("Callback returned error: %#v\n", err)
						return input, err
					}
					fmt.Printf("Callback returned result: %#v\n", res)
					input = res
				} else {
					fmt.Printf("No callback\n")
				}
				found = true
				s.state = tr.Destination
				break
			}
		}
		if !found {
			fmt.Printf("None found, stop\n")
			return input, nil
		}
		fmt.Printf("New state is %s\n", s.state)
	}
}

func NewSM(state string, trs []Transition) Machine {
	sm := &sm{
		state:       state,
		transitions: map[string][]Transition{},
	}
	for _, tr := range trs {
		sm.transitions[tr.Source] = append(sm.transitions[tr.Source], tr)
	}
	return sm
}
