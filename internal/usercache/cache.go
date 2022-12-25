package usercache

import (
	"context"
	"encoding/json"
	"strings"

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
	db      DB
}

func (c *cache) Get(ctx context.Context, user tgapi.User) (pkgcache.User, error) {
	if u, ok := c.cache[user.Id]; ok {
		logging.S(ctx).Debugf("Cached user %v %v", user, u)
		return u, nil
	} else {
		u := c.factory.MakeUser(user)
		stored, err := c.db.Get(ctx, user.Id)
		if err != nil {
			if xerrors.Is(err, noRowsError) {
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
	if err := c.db.Add(ctx, StoredUser{
		Id:       tgUser.Id,
		Name:     tgUser.FirstName,
		Contents: string(content),
	}); err != nil {
		return xerrors.Errorf("add: %w", err)
	}
	return nil
}

func (c *cache) Close() { c.db.Close() }

func (c *cache) AttachFactory(ctx context.Context, factory UserFactory) error {
	c.factory = factory
	users, err := c.db.List(ctx)
	if err != nil {
		return xerrors.Errorf("list: %w", err)
	}
	for _, user := range users {
		u := c.factory.MakeUser(tgapi.User{Id: user.Id, FirstName: user.Name})
		if err := json.Unmarshal([]byte(user.Contents), u); err != nil {
			return xerrors.Errorf("make user: %w", err)
		}
		u.Wake()
	}
	return nil
}

type Config struct {
	Driver   string `yaml:"driver" env:"TGBOT_DB_DRIVER"`
	Path     string `yaml:"path" env:"TGBOT_DB_PATH"`
	User     string `yaml:"user" env:"TGBOT_DB_USER"`
	Password string `yaml:"-" env:"TGBOT_DB_PASSWORD"`
	Database string `yaml:"database" env:"TGBOT_DB_DATABASE"`
}

func NewCache(ctx context.Context, cfg Config) (*cache, error) {
	var db DB
	var err error
	switch strings.ToLower(cfg.Driver) {
	case "sqlite":
		db, err = NewSQLiteDB(ctx, cfg)
	case "pg":
		db, err = NewPGDB(ctx, cfg)
	default:
		return nil, xerrors.Errorf("unknown driver: %q", cfg.Driver)
	}
	if err != nil {
		return nil, xerrors.Errorf("new db: %w", err)
	}
	return &cache{
		db:    db,
		cache: map[uint64]*impl.User{},
	}, nil
}
