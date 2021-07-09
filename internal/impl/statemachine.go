package impl

import "github.com/baldisbk/tgbot_sample/pkg/statemachine"

const (
	startState   = "start"
	timerState   = "timer"
	listState    = "list"
	displayState = "display"
	addState     = "add"
	reportState  = "report"
)

func (u *User) DontUnderstand(state string) statemachine.Transition {
	return statemachine.Transition{
		Source: state, Destination: state, Predicate: statemachine.NotNilPredicate, Callback: u.doNoUnderstand,
	}
}

func makeTransitions(res *User) []statemachine.Transition {
	return []statemachine.Transition{
		// from initial
		{
			Source:      startState,
			Destination: startState,
			Predicate:   res.isStart,
			Callback:    res.doStart,
		},
		{
			Source:      startState,
			Destination: timerState,
			Predicate:   res.isTimer,
			Callback:    res.doTimer,
		},
		{
			Source:      startState,
			Destination: listState,
			Predicate:   checkCallback(listCallback),
			Callback:    res.doList,
		},
		{
			Source:      startState,
			Destination: addState,
			Predicate:   checkCallback(addCallback),
			Callback:    res.doStartAdd,
		},
		res.DontUnderstand(startState),

		// from timer
		{ // rollback - auto postpone to default period
			Source:      timerState,
			Destination: startState,
			Predicate:   res.isRollback,
			Callback:    statemachine.CompositeCallback(res.doPostpone, res.doStart),
		},
		{
			Source:      timerState,
			Destination: reportState,
			Predicate:   checkCallback(reportCallback),
			Callback:    res.doReport,
		},
		{
			Source:      timerState,
			Destination: startState,
			Predicate:   checkCallback(postponeCallback),
			// TODO custom postpone time via menu, two-step
			Callback: statemachine.CompositeCallback(res.doPostpone, res.doStart),
		},
		res.DontUnderstand(timerState),

		// from list
		{ // rollback - do noting, it's just a menu
			Source:      listState,
			Destination: startState,
			Predicate:   res.isRollback,
			Callback:    res.doStart,
		},
		{
			Source:      listState,
			Destination: listState,
			Predicate:   checkCallback(forwardListCallback),
			Callback:    res.doListForward,
		},
		{
			Source: listState, Destination: listState,
			Predicate: checkCallback(backwardListCallback),
			Callback:  res.doListBackward,
		},
		{
			Source:      listState,
			Destination: startState,
			Predicate:   checkCallback(stopListCallback),
			Callback:    res.doStart,
		},
		{
			Source:      listState,
			Destination: displayState,
			Predicate:   res.isDisplay,
			Callback:    res.doDisplay,
		},
		res.DontUnderstand(listState),

		// from display
		{ // rollback - do nothing, it's just a menu
			Source:      displayState,
			Destination: startState,
			Predicate:   res.isRollback,
			Callback:    res.doStart,
		},
		{
			Source:      displayState,
			Destination: listState,
			Predicate:   checkCallback(listCallback),
			Callback:    res.doList,
		},
		{
			Source:      displayState,
			Destination: startState,
			Predicate:   checkCallback(stopListCallback),
			Callback:    res.doStart,
		},
		res.DontUnderstand(displayState),

		// from add
		{ // rollback - drop all inputs
			Source:      addState,
			Destination: startState,
			Predicate:   res.isRollback,
			Callback:    statemachine.CompositeCallback(res.dropAdd, res.doStart),
		},
		{
			Source:      addState,
			Destination: startState,
			Predicate:   checkCallback(okCallback),
			Callback:    statemachine.CompositeCallback(res.doFinishAdd, res.dropAdd, res.doStart),
		},
		{
			Source:      addState,
			Destination: addState,
			Predicate:   checkCallback(retryCallback),
			Callback:    res.doAdd,
		},
		{
			Source:      addState,
			Destination: startState,
			Predicate:   checkCallback(abortCallback),
			Callback:    statemachine.CompositeCallback(res.dropAdd, res.doStart),
		},
		{
			Source:      addState,
			Destination: addState,
			Predicate:   res.isValidInput,
			Callback:    res.doAdd,
		},
		res.DontUnderstand(addState),

		// from report
		{ // rollback - auto postpone
			Source:      addState,
			Destination: startState,
			Predicate:   res.isRollback,
			Callback:    statemachine.CompositeCallback(res.doPostpone, res.doStart),
		},
		{
			Source:      reportState,
			Destination: startState,
			Predicate:   res.isValidInput,
			Callback:    statemachine.CompositeCallback(res.doFinishReport, res.doStart),
		},
		res.DontUnderstand(reportState),
	}
}
