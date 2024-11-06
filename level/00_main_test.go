// /home/krylon/go/src/github.com/blicero/cacheme/level/00_main_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 11. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-11-06 19:17:45 krylon>

package level

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var testPath string

func TestMain(m *testing.M) {
	var result int

	testPath = time.Now().Format("/tmp/cacheme_level_test_20060102_150405")

	if result = m.Run(); result == 0 {
		// If any test failed, we keep the test directory (and the
		// database inside it) around, so we can manually inspect it
		// if needed.
		// If all tests pass, OTOH, we can safely remove the directory.
		fmt.Printf("Removing BaseDir %s\n",
			testPath)
		_ = os.RemoveAll(testPath)
	} else {
		fmt.Printf(">>> TEST DIRECTORY: %s\n", testPath)
	}

	os.Exit(result)
} // func TestMain(m *testing.M)
