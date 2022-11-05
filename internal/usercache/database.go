package usercache

import (
	"golang.org/x/xerrors"
)

type StoredUser struct {
	Id       uint64
	Name     string
	Contents string
}

const (
	schemaSQL = `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY ON CONFLICT REPLACE, name TEXT, contents TEXT);`
	insertSQL = `INSERT INTO users (id, name, contents) VALUES (?, ?, ?);`
	selectSQL = `SELECT name, contents FROM users WHERE id=?;`
	listSQL   = `SELECT name, contents FROM users;` // TODO paging
)

var noRowsError = xerrors.New("no rows found")
