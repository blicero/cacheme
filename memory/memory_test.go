// /home/krylon/go/src/github.com/blicero/cacheme/memory/memory_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 11. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-11-06 18:01:02 krylon>

package memory

import (
	"fmt"
	"testing"
	"time"

	"github.com/blicero/cacheme"
)

const (
	cnt     = 16
	testTTL = time.Second * 2
)

var c cacheme.Backend

func TestCreate(t *testing.T) {
	if c = New(); c == nil {
		t.Fatal("New returned a nil value")
	}
} // func TestCreate(t *testing.T)

func TestInstall(t *testing.T) {
	if c == nil {
		t.SkipNow()
	}

	for i := 0; i < cnt; i++ {
		var (
			key, val string
			err      error
		)

		key = fmt.Sprintf("Key%02d", i)
		val = fmt.Sprintf("Value%02d", i)
		if err = c.Install(key, val, testTTL); err != nil {
			t.Errorf("Failed to install key %s: %s",
				key,
				err.Error())
		}
	}
} // func TestInstall(t *testing.T)

func TestLookup(t *testing.T) {
	if c == nil {
		t.SkipNow()
	}

	for i := 0; i < cnt; i++ {
		var (
			key, val, expectedVal string
			expires               time.Time
			err                   error
			ok                    bool
			now                   = time.Now()
		)

		key = fmt.Sprintf("Key%02d", i)
		expectedVal = fmt.Sprintf("Value%02d", i)
		if val, ok, expires, err = c.Lookup(key); err != nil {
			t.Errorf("Error looking up key %s: %s",
				key,
				err.Error())
		} else if !ok {
			t.Errorf("Key %s was not found", key)
		} else if expires.Before(now) {
			t.Errorf("Key %s was returned, but has already expired?\nexpired = %s\nnow = %s",
				key,
				expires.Format(time.DateTime),
				now.Format(time.DateTime))
		} else if val != expectedVal {
			t.Errorf("Unexpected value for key %s - have: %q, want: %q",
				key,
				val,
				expectedVal)
		}
	}
} // func TestLookup(t *testing.T)

func TestExpiration(t *testing.T) {
	if c == nil {
		t.SkipNow()
	}

	time.Sleep(testTTL + time.Second)

	for i := 0; i < cnt; i++ {
		var (
			key     string
			expires time.Time
			err     error
			ok      bool
		)

		key = fmt.Sprintf("Key%02d", i)
		if _, ok, expires, err = c.Lookup(key); err != nil {
			t.Errorf("Error looking up key %s: %s",
				key,
				err.Error())
		} else if ok {
			t.Errorf("Key %s was found, should have expired by now. TTL says %s",
				key,
				expires.Format(time.DateTime))
		}
	}
} // func TestExpiration(t *testing.T)
