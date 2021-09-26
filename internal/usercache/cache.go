package usercache

import (
	"context"
	"encoding/json"

	"github.com/baldisbk/tgbot_sample/internal/impl"
	"github.com/baldisbk/tgbot_sample/pkg/logging"
	"github.com/baldisbk/tgbot_sample/pkg/tgapi"
	pkgcache "github.com/baldisbk/tgbot_sample/pkg/usercache"

	"golang.org/x/xerrors"
)

type UserFactory interface {
	MakeUser(tgapi.User) *impl.User
}

type cache struct {
	// TODO: change to LRU cache
	cache   map[uint64]*impl.User
	factory UserFactory
	db      *DB
}

func (c *cache) Get(ctx context.Context, user tgapi.User) (pkgcache.User, error) {
	if u, ok := c.cache[user.Id]; ok {
		logging.S(ctx).Debugf("Cached user %v %v", user, u)
		return u, nil
	} else {
		u := c.factory.MakeUser(user)
		stored, err := c.db.Get(user.Id)
		if err != nil {
			if err != noRowsError {
				return nil, xerrors.Errorf("get: %w", err)
			}
			logging.S(ctx).Debugf("New user %v", user)
		} else {
			logging.S(ctx).Debugf("DB user %v", user)
			if err := json.Unmarshal([]byte(stored.Contents), u); err != nil {
				return nil, xerrors.Errorf("umarshal: %w", err)
			}
		}
		u.Wake()
		logging.S(ctx).Debugf("Store user %v", user, u)
		c.cache[user.Id] = u
		return u, nil
	}
}

func (c *cache) Put(ctx context.Context, tgUser tgapi.User, state pkgcache.User) error {
	content, err := json.Marshal(state)
	if err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}
	if err := c.db.Add(StoredUser{
		Id:       tgUser.Id,
		Name:     tgUser.FirstName,
		Contents: string(content),
	}); err != nil {
		return xerrors.Errorf("add: %w", err)
	}
	return nil
}

// TODO close DB connection
func (c *cache) Close() {}

func (c *cache) AttachFactory(factory UserFactory) {
	c.factory = factory
	users, _ := c.db.List()
	for _, user := range users {
		u := c.factory.MakeUser(tgapi.User{Id: user.Id, FirstName: user.Name})
		if err := json.Unmarshal([]byte(user.Contents), u); err == nil {
			u.Wake()
		}
	}
}

type Config struct {
	Filename string `yaml:"database"`
}

func NewCache(cfg Config) (*cache, error) {
	db, err := NewDB(cfg.Filename)
	if err != nil {
		return nil, xerrors.Errorf("new db: %w", err)
	}
	return &cache{
		db:    db,
		cache: map[uint64]*impl.User{},
	}, nil
}
