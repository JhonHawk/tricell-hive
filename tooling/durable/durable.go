// Package durable holds the fsync switch shared by every writer under tooling.
// It imports no other tooling package so any of them can depend on it.
package durable

import (
	"os"
	"sync/atomic"
	"testing"
)

// skipSync, when set, stops writes from being flushed to stable storage.
// Its zero value keeps syncing on in production; only test code sets it.
var skipSync atomic.Bool

// isTestBinary reports whether the process is a test binary. It is a variable
// so a test can simulate a production binary.
var isTestBinary = testing.Testing

// DisableForTests turns off fsync for the rest of the process. Only test
// binaries call it, from TestMain: production code must keep its durability
// guarantee, so it panics anywhere else.
func DisableForTests() {
	if !isTestBinary() {
		panic("DisableForTests called outside a test binary")
	}
	skipSync.Store(true)
}

// Sync flushes a file or directory unless disk sync is disabled.
func Sync(f *os.File) error {
	if skipSync.Load() {
		return nil
	}
	return f.Sync()
}
