package cache

import (
	"github.com/dgraph-io/badger/v4"
	"time"
)

const meta byte = 1

type Cache struct {
	db *badger.DB
}

func New(path string) (*Cache, error) {
	options := badger.DefaultOptions(path).WithLogger(nil)

	if path == "" {
		options = options.WithInMemory(true)
	}

	db, err := badger.Open(options)

	if err != nil {
		return nil, err
	}

	return &Cache{
		db: db,
	}, nil
}

func (c *Cache) Close() error {
	var err error

	for err == nil {
		err = c.db.RunValueLogGC(0.7)
	}

	return c.db.Close()
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	var v []byte

	return v, c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		if err != nil {
			return err
		}

		if item.UserMeta() != meta {
			return badger.ErrKeyNotFound
		}

		return item.Value(func(val []byte) error {
			v = val
			return nil
		})
	})
}

func (c *Cache) SetWithExpiry(key []byte, value []byte, expiresAt time.Time) error {
	return c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value).WithMeta(meta)
		e.ExpiresAt = uint64(expiresAt.Unix())

		return txn.SetEntry(e)
	})
}
