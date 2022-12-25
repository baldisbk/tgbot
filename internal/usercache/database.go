package usercache

import (
	"golang.org/x/xerrors"
)

type StoredUser struct {
	Id       uint64
	Name     string
	Contents string
}

var noRowsError = xerrors.New("no rows found")
