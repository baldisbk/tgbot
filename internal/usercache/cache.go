package usercache

import (
	"encoding/json"
	"fmt"

	"github.com/baldisbk/tgbot_sample/internal/tgapi"
	"golang.org/x/xerrors"
)

type cache struct {
	// TODO: change to LRU cache
	cache map[tgapi.User]User
	ctor  UserFactory
	db    *DB
}

func (c *cache) Get(user tgapi.User) (User, error) {
	if u, ok := c.cache[user]; ok {
		fmt.Println("\t CACHED USER", user)
		return u, nil
	} else {
		u := c.ctor(user)
		stored, err := c.db.Get(user.Id)
		if err != nil {
			if err != noRowsError {
				return nil, xerrors.Errorf("get: %w", err)
			}
			fmt.Println("\t NEW USER", user)
		} else {
			fmt.Println("\t STORED USER", user)
			if err := json.Unmarshal([]byte(stored.Contents), u); err != nil {
				return nil, xerrors.Errorf("umarshal: %w", err)
			}
		}
		c.cache[user] = u
		return u, nil
	}
}

func (c *cache) Put(user tgapi.User, state User) error {
	c.cache[user] = state
	content, err := json.Marshal(state)
	if err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}
	if err := c.db.Add(StoredUser{
		Id:       user.Id,
		Name:     user.FirstName,
		Contents: string(content),
	}); err != nil {
		return xerrors.Errorf("add: %w", err)
	}
	return nil
}

// TODO close DB connection
func (c *cache) Close() {}

type Config struct {
	Filename string
}

func NewCache(cfg Config, ctor UserFactory) (UserCache, error) {
	db, err := NewDB(cfg.Filename)
	if err != nil {
		return nil, xerrors.Errorf("new db: %w", err)
	}
	return &cache{
		db:    db,
		cache: map[tgapi.User]User{},
		ctor:  ctor,
	}, nil
}
