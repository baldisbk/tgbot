package impl

import "github.com/baldisbk/tgbot/pkg/statemachine"

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

func (u *User) timed(callbacks ...statemachine.SMCallback) statemachine.SMCallback {
	cbs := append([]statemachine.SMCallback{u.doTimeout}, callbacks...)
	return statemachine.CompositeCallback(cbs...)
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
		{ // rollback - just do nothing
			Source:      startState,
			Destination: startState,
			Predicate:   res.isRollback,
		},
		{
			Source:      startState,
			Destination: timerState,
			Predicate:   res.isTimer,
			Callback:    res.timed(res.doTimer),
		},
		{
			Source:      startState,
			Destination: listState,
			Predicate:   checkCallback(listCallback),
			Callback:    res.timed(res.doList),
		},
		{
			Source:      startState,
			Destination: addState,
			Predicate:   checkCallback(addCallback),
			Callback:    res.timed(res.doStartAdd),
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
			Callback:    res.timed(res.doReport),
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
			Callback:    res.timed(res.doListForward),
		},
		{
			Source:      listState,
			Destination: listState,
			Predicate:   checkCallback(backwardListCallback),
			Callback:    res.timed(res.doListBackward),
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
			Callback:    res.timed(res.doDisplay),
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
			Callback:    res.timed(res.doList),
		},
		{
			Source:      displayState,
			Destination: startState,
			Predicate:   checkCallback(stopListCallback),
			Callback:    res.timed(res.doStart),
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
			Callback:    res.timed(res.doAdd),
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
			Callback:    res.timed(res.doAdd),
		},
		res.DontUnderstand(addState),

		// from report
		{ // rollback - auto postpone
			Source:      reportState,
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
