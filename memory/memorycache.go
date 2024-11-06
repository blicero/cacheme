// /home/krylon/go/src/github.com/blicero/cacheme/memorycache.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 11. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-11-06 17:19:40 krylon>

// Package memory provides a simple in-memory cache based on a map.
package memory

import (
	"sync"
	"time"

	"github.com/blicero/cacheme"
)

var zero time.Time

// Memory implements a cache based on a simple map.
type Memory struct {
	store map[string]cacheme.Value
	lock  sync.RWMutex
}

// New creates a fresh Memory Cache and returns it.
func New() *Memory {
	var m = &Memory{store: make(map[string]cacheme.Value)}

	return m
} // func New() *Memory

func (m *Memory) Install(key, val string, ttl time.Duration) error {
	m.lock.Lock()
	m.store[key] = cacheme.Value{
		Val:     val,
		Expires: time.Now().Add(ttl),
	}
	m.lock.Unlock()
	return nil
} // func (m *Memory) Install(key, val string, ttl time.Duration) error

func (m *Memory) Lookup(key string) (string, bool, time.Time, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var (
		val cacheme.Value
		ok  bool
	)

	if val, ok = m.store[key]; !ok {
		return "", false, zero, nil
	} else if val.Expires.Before(time.Now()) {
		return "", false, zero, nil
	} else {
		return val.Val, true, val.Expires, nil
	}
} // func (m *Memory) Lookup(key string) (string, time.Duration, error)

func (m *Memory) Delete(key string) error {
	m.lock.Lock()
	delete(m.store, key)
	m.lock.Unlock()
	return nil
} // func (m *Memory) Delete(key string) error

func (m *Memory) Purge() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var now = time.Now()

	for k, v := range m.store {
		if v.Expires.Before(now) {
			delete(m.store, k)
		}
	}

	return nil
} // func (m *Memory) Purge() error

func (m *Memory) Flush() error {
	m.lock.Lock()
	m.store = make(map[string]cacheme.Value)
	m.lock.Unlock()
	return nil
} // func (m *Memory) Flush() error
