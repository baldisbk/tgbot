package usercache

import "context"

type DB interface {
	Add(ctx context.Context, user StoredUser) error
	Get(ctx context.Context, id uint64) (*StoredUser, error)
	List(ctx context.Context) ([]StoredUser, error)
	Close()
}
