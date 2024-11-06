// /home/krylon/go/src/github.com/blicero/cacheme/cacheme.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 11. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-11-06 17:40:49 krylon>

// Package cacheme provides a cache that is safe to use from multiple
// goroutines concurrently.
// The access model is similar to what BoltDB provides (as a point of reference) or
// SQLite pre-WAL: One writer XOR multiple readers simultaneously.
package cacheme

import "time"

// Value is a helper type so make it easier to store the value and its
// expiration timestamp together.
// This should not be considered part of the public interface, except for people
// wanting to write new Backends, as it reduces some duplication of effort (hopefully).
type Value struct {
	Val     string
	Expires time.Time
}

// Backend defines the interface for the various backends.
// All operations - for the time being - are blocking.
type Backend interface {
	Install(key, val string, ttl time.Duration) error
	Lookup(key string) (string, bool, time.Time, error)
	Delete(key string) error
	Purge() error
	Flush() error
}
