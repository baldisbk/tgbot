package usercache

import (
	"context"
	"database/sql"

	"golang.org/x/xerrors"

	_ "github.com/mattn/go-sqlite3"
)

const (
	schemaSQLite = `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY ON CONFLICT REPLACE, name TEXT, contents TEXT);`
	insertSQLite = `INSERT INTO users (id, name, contents) VALUES (?, ?, ?);`
	selectSQLite = `SELECT name, contents FROM users WHERE id=?;`
	listSQLite   = `SELECT name, contents FROM users;` // TODO paging
)

func NewSQLiteDB(ctx context.Context, cfg Config) (DB, error) {
	sqlDB, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}
	db := sqliteDB{sql: sqlDB}
	if err := db.prepare(); err != nil {
		return nil, xerrors.Errorf("prepare: %w", err)
	}
	return &db, nil
}

type sqliteDB struct {
	sql  *sql.DB
	ins  *sql.Stmt
	sel  *sql.Stmt
	list *sql.Stmt
}

func (db *sqliteDB) prepare() error {
	var err error
	if _, err = db.sql.Exec(schemaSQLite); err != nil {
		return xerrors.Errorf("exec: %w", err)
	}
	db.ins, err = db.sql.Prepare(insertSQLite)
	if err != nil {
		return xerrors.Errorf("prepare insert: %w", err)
	}
	db.sel, err = db.sql.Prepare(selectSQLite)
	if err != nil {
		return xerrors.Errorf("prepare select: %w", err)
	}
	db.list, err = db.sql.Prepare(listSQLite)
	if err != nil {
		return xerrors.Errorf("prepare list: %w", err)
	}
	return nil
}

func (db *sqliteDB) Add(ctx context.Context, user StoredUser) error {
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

func (db *sqliteDB) Get(ctx context.Context, id uint64) (*StoredUser, error) {
	res, err := db.sql.Query(selectSQLite, id)
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

func (db *sqliteDB) List(ctx context.Context) ([]StoredUser, error) {
	res, err := db.sql.Query(listSQLite)
	if err != nil {
		return nil, xerrors.Errorf("exec: %w", err)
	}
	defer res.Close()
	var users []StoredUser
	for res.Next() {
		if err := res.Err(); err != nil {
			return nil, xerrors.Errorf("res next: %w", err)
		}
		var id uint64
		var name, contents string
		if err := res.Scan(&id, &name, &contents); err != nil {
			return nil, xerrors.Errorf("scan: %w", err)
		}
		users = append(users, StoredUser{
			Id:       id,
			Name:     name,
			Contents: contents,
		})
	}
	return users, nil
}

func (db *sqliteDB) Close() {
	db.ins.Close()
	db.sel.Close()
	db.sql.Close()
}
