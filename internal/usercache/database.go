package usercache

import (
	"database/sql"

	"golang.org/x/xerrors"

	_ "github.com/mattn/go-sqlite3"
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
)

var noRowsError = xerrors.New("no rows found")

type DB struct {
	sql *sql.DB
	ins *sql.Stmt
	sel *sql.Stmt
}

func NewDB(dbFile string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}
	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		return nil, xerrors.Errorf("exec: %w", err)
	}

	ins, err := sqlDB.Prepare(insertSQL)
	if err != nil {
		return nil, xerrors.Errorf("prepare insert: %w", err)
	}
	sel, err := sqlDB.Prepare(selectSQL)
	if err != nil {
		return nil, xerrors.Errorf("prepare select: %w", err)
	}

	db := DB{
		sql: sqlDB,
		ins: ins,
		sel: sel,
	}
	return &db, nil
}

func (db *DB) Add(user StoredUser) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return xerrors.Errorf("tx: %w", err)
	}
	if _, err = tx.Stmt(db.ins).Exec(user.Id, user.Name, user.Contents); err != nil {
		tx.Rollback()
		return xerrors.Errorf("exec: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return xerrors.Errorf("commit: %w", err)
	}
	return nil
}

func (db *DB) Get(id uint64) (*StoredUser, error) {
	res, err := db.sql.Query(selectSQL, id)
	if err != nil {
		return nil, xerrors.Errorf("exec: %w", err)
	}
	defer res.Close()
	if !res.Next() {
		if err := res.Err(); err != nil {
			return nil, xerrors.Errorf("res next: %w", err)
		}
		return nil, noRowsError
	}
	var name, contents string
	if err := res.Scan(&name, &contents); err != nil {
		return nil, xerrors.Errorf("scan: %w", err)
	}
	return &StoredUser{
		Id:       id,
		Name:     name,
		Contents: contents,
	}, nil
}

func (db *DB) Close() {
	db.ins.Close()
	db.sel.Close()
	db.sql.Close()
}
