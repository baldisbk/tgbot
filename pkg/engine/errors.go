package engine

import "golang.org/x/xerrors"

// TODO make them types
var (
	BadStateError   = xerrors.New("bad state")   // message is ok, but state is wrong
	BadMessageError = xerrors.New("bad message") // message is not acceptable
	RetriableError  = xerrors.New("retriable")
	FatalError      = xerrors.New("fatal")
)
