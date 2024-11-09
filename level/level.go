// /home/krylon/go/src/github.com/blicero/cacheme/level/level.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 11. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-11-09 19:46:51 krylon>

// Package level implements the cacheme.Backend interface using LevelDB as its storage backend.
package level

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/blicero/cacheme"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var zero time.Time

// Cache is an implementation of the cacheme.Backend interface, using LevelDB
// as its backend to provide persistence.
type Cache struct {
	lock sync.RWMutex
	db   *leveldb.DB
}

// New creates a new LevelCache that stores its data at the given path.
func New(path string) (*Cache, error) {
	var (
		c   *Cache
		err error
	)

	c = new(Cache)

	if c.db, err = leveldb.OpenFile(path, nil); err != nil {
		return nil, err
	}

	return c, nil
} // func New(path string) (*LevelCache, error)

// Install adds a key-value-pair with the given TTL to the Cache. If the key
// already exists, it is silently replaced.
func (l *Cache) Install(key, val string, ttl time.Duration) error {
	var (
		err            error
		v              cacheme.Value
		valBuf, keyBuf []byte
	)

	l.lock.Lock()
	defer l.lock.Unlock()

	v.Val = val
	v.Expires = time.Now().Add(ttl)

	if valBuf, err = json.Marshal(&v); err != nil {
		return err
	}

	keyBuf = []byte(key)

	if err = l.db.Put(keyBuf, valBuf, nil); err != nil {
		return err
	}

	return nil
} // func (l *LevelCache) Install(key, val string, ttl time.Duration) error

// Lookup looks up the given key. If the key is found AND has not yet expired,
// the associated value is returned.
func (l *Cache) Lookup(key string) (string, bool, time.Time, error) {
	var (
		err  error
		jval []byte
		val  cacheme.Value
	)

	l.lock.RLock()
	defer l.lock.RUnlock()

	if jval, err = l.db.Get([]byte(key), nil); err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return "", false, zero, nil
		}

		return "", false, zero, err
	} else if err = json.Unmarshal(jval, &val); err != nil {
		return "", true, zero, err
	} else if val.Expires.Before(time.Now()) {
		return "", false, zero, nil
	}

	return val.Val, true, val.Expires, nil
} // func (l *LevelCache) Lookup(key string) (string, bool, time.Time, error)

// Delete removes the given key and its value from the Cache.
func (l *Cache) Delete(key string) error {
	var (
		err error
	)

	l.lock.Lock()
	defer l.lock.Unlock()

	if err = l.db.Delete([]byte(key), nil); err != nil {
		return err
	}

	return nil
} // func (l *LevelCache) Delete(key string) error

// Purge removes all key-value-pairs that have expired from the Cache.
func (l *Cache) Purge() (e error) {
	var (
		err    error
		status bool
		now    time.Time
		tx     *leveldb.Transaction
	)

	l.lock.Lock()
	defer l.lock.Unlock()

	if tx, err = l.db.OpenTransaction(); err != nil {
		return err
	}

	defer func() {
		var x error
		if !status {
			tx.Discard()
		} else if x = tx.Commit(); x != nil {
			e = x
		}
	}()

	iter := tx.NewIterator(nil, nil)
	defer iter.Release()
	now = time.Now()

	for iter.Next() {
		var (
			vbuf []byte
			v    cacheme.Value
		)

		vbuf = iter.Value()

		if err = json.Unmarshal(vbuf, &v); err != nil {
			return err
		} else if v.Expires.Before(now) {
			if err = tx.Delete(iter.Key(), nil); err != nil {
				return err
			}
		}
	}

	if err = l.db.CompactRange(util.Range{Start: nil, Limit: nil}); err != nil {
		return err
	}

	status = true

	return nil
} // func (l *LevelCache) Purge() (e error)

// Flush removes ALL key-value-pairs from the Cache.
func (l *Cache) Flush() (e error) {
	var (
		err    error
		tx     *leveldb.Transaction
		iter   iterator.Iterator
		status bool
	)

	l.lock.Lock()
	defer l.lock.Unlock()

	if tx, err = l.db.OpenTransaction(); err != nil {
		return err
	}

	defer func() {
		var x error
		if !status {
			tx.Discard()
		} else if x = tx.Commit(); x != nil {
			e = x
		}
	}()

	iter = tx.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		if err = tx.Delete(iter.Key(), nil); err != nil {
			return err
		}
	}

	if err = l.db.CompactRange(util.Range{Start: nil, Limit: nil}); err != nil {
		return err
	}

	status = true

	return nil
} // func (l *LevelCache) Flush() (e error)
