package usercache

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/xerrors"
)

type pgDB struct {
	pool *pgxpool.Pool
}

func NewPGDB(ctx context.Context, dbPath string) (DB, error) {
	pool, err := pgxpool.New(context.Background(), dbPath)
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}
	db := pgDB{pool: pool}
	return &db, nil
}

func (db *pgDB) tx(ctx context.Context, proc func(context.Context, pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return xerrors.Errorf("tx: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := proc(ctx, tx); err != nil {
		return xerrors.Errorf("exec: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return xerrors.Errorf("commit: %w", err)
	}
	return nil
}

func (db *pgDB) Add(ctx context.Context, user StoredUser) error {
	return db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, insertSQL, user.Id, user.Name, user.Contents); err != nil {
			return xerrors.Errorf("exec: %w", err)
		}
		return nil
	})
}

func (db *pgDB) Get(ctx context.Context, id uint64) (*StoredUser, error) {
	var rows pgx.Rows
	err := db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		rows, err = db.pool.Query(ctx, selectSQL, id)
		if err != nil {
			return xerrors.Errorf("exec: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, xerrors.Errorf("exec: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, xerrors.Errorf("res next: %w", err)
		}
		return nil, noRowsError
	}
	var name, contents string
	if err := rows.Scan(&name, &contents); err != nil {
		return nil, xerrors.Errorf("scan: %w", err)
	}
	return &StoredUser{
		Id:       id,
		Name:     name,
		Contents: contents,
	}, nil
}

func (db *pgDB) List(ctx context.Context) ([]StoredUser, error) {
	var rows pgx.Rows
	err := db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		rows, err = db.pool.Query(ctx, listSQL)
		if err != nil {
			return xerrors.Errorf("exec: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, xerrors.Errorf("exec: %w", err)
	}
	defer rows.Close()
	var users []StoredUser
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, xerrors.Errorf("res next: %w", err)
		}
		var id uint64
		var name, contents string
		if err := rows.Scan(&id, &name, &contents); err != nil {
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

func (db *pgDB) Close() {
	db.pool.Close()
}
