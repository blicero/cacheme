// /home/krylon/go/src/github.com/blicero/cacheme/memory/memory_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 11. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-11-07 19:46:25 krylon>

package level

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
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
	var err error

	if c, err = New(testPath); err != nil {
		t.Fatalf("Failed to open LevelDB at %s: %s",
			testPath,
			err.Error())
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

type concArgs struct {
	id       int64
	t        *testing.T
	wg       *sync.WaitGroup
	keyCnt   int
	deadline time.Time
	duration time.Duration
	jitter   [2]int64 // Range, in milliseconds, of random delay.
	errq     chan error
}

func concWorker(arg *concArgs) {
	defer arg.wg.Done()

	// We assume that time.Now().Add(arg.duration) == arg.deadline,
	// or at least so close the difference is negligible.
	// I just want to avoid having all workers perform that same computation,
	// trivial as it is.

	var (
		err    error
		ttl    = arg.duration / 50
		ticker = time.NewTicker(arg.duration)
		rng    = rand.New(rand.NewSource(time.Now().Unix() + arg.id*100))
	)

	defer ticker.Stop()

	for time.Now().Before(arg.deadline) {
		var data = make(map[string]string, arg.keyCnt)

		for i := 0; i < arg.keyCnt; i++ {
			var (
				key = fmt.Sprintf("%02d_%02d",
					arg.id,
					i)
				val = fmt.Sprintf("%016x",
					rng.Int63())
			)

			if err = c.Install(key, val, ttl); err != nil {
				var e2 = fmt.Errorf("Worker %d failed to install key %q => %q in cache: %s",
					arg.id,
					key,
					val,
					err.Error())
				var e3 = errors.Join(e2, err)
				arg.errq <- e3
				return
			}

			data[key] = val
		}

		var delay = time.Duration(rng.Int63n(arg.jitter[1]-arg.jitter[0]) + arg.jitter[0])
		time.Sleep(delay * time.Millisecond)

		for k, v := range data {
			var (
				cval string
				ok   bool
			)

			if cval, ok, _, err = c.Lookup(k); err != nil {
				var e2 = fmt.Errorf("Worker %d failed to lookup key %q: %s",
					arg.id,
					k,
					err.Error())
				var e3 = errors.Join(e2, err)
				arg.errq <- e3
				return // *should* be redundant
			} else if !ok && (delay*time.Millisecond) < ttl {
				var e2 = fmt.Errorf("Did not find key %q in cache, it should still be there for at least %s",
					k,
					(ttl - delay))
				var e3 = errors.Join(e2, err)
				arg.errq <- e3
				return // Redundant?
			} else if ok && v != cval {
				var e2 = fmt.Errorf(`unexpected value for key %q:
Expected: %q
Got:      %q`,
					k,
					v,
					cval)
				var e3 = errors.Join(e2, err)
				arg.errq <- e3
				return // Redundant?
			}
		}
	}
} // func concWorker(arg *concArgs)

func TestConcurrency(t *testing.T) {
	if c == nil {
		t.SkipNow()
	}

	const (
		concLvl  int64 = 64
		seconds        = 5
		keyCnt         = 32
		duration       = time.Second * seconds
	)

	var (
		err      error
		wg       sync.WaitGroup
		i        int64
		deadline = time.Now().Add(duration)
		errq     = make(chan error)
		timer    = time.NewTimer(duration + time.Second*2)
	)

	wg.Add(int(concLvl))

	for i = 0; i < concLvl; i++ {
		var arg = &concArgs{
			id:       i,
			t:        t,
			wg:       &wg,
			keyCnt:   keyCnt,
			deadline: deadline,
			duration: duration,
			jitter:   [2]int64{100, 300},
			errq:     errq,
		}

		go concWorker(arg)
	}

	select {
	case <-timer.C:
		// done
	case err = <-errq:
		t.Fatal(err.Error())
	}

	wg.Wait()
} // func TestConcurrency(t *testing.T)
