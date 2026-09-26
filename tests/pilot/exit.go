package main

import (
	"os"
	"os/signal"
	"syscall"
)

// exitCleanups holds cleanup callbacks that must run on an abnormal exit: a
// must() failure or a SIGINT/SIGTERM. Registering one (registerCleanup) is
// what guarantees a resource created mid-run — such as the guidance
// variant's shadow-home Codex auth.json symlink — is torn down on every exit
// path after its setup, not only the happy path that calls its teardown
// directly.
var exitCleanups []func()

// registerCleanup adds f to the list must() and the signal handler installed
// by startSignalCleanup run before this process exits abnormally. f must be
// safe to call more than once: the normal, successful flow may also invoke
// the same teardown directly on its own happy path, and a second must()
// failure after that would run every registered cleanup again.
func registerCleanup(f func()) {
	exitCleanups = append(exitCleanups, f)
}

// runCleanups runs every registered cleanup, most-recently-registered first,
// then clears the list so a later call is a no-op.
func runCleanups() {
	for i := len(exitCleanups) - 1; i >= 0; i-- {
		exitCleanups[i]()
	}
	exitCleanups = nil
}

// installSignalCleanup starts a goroutine that runs every registered cleanup
// and then exits once sig delivers a signal, so an operator's Ctrl-C (or a
// TERM from a supervising process) between a resource's setup and its
// happy-path teardown still leaves it cleaned up. It takes runCleanups and
// exit as parameters, rather than calling the package-level runCleanups and
// os.Exit directly, so a test can observe both without sending this test
// process a real signal.
func installSignalCleanup(sig <-chan os.Signal, runCleanups func(), exit func(int)) {
	go func() {
		s, ok := <-sig
		if !ok {
			return
		}
		runCleanups()
		code := 128
		if n, ok := s.(syscall.Signal); ok {
			code += int(n)
		}
		exit(code)
	}()
}

// startSignalCleanup is main()'s real wiring: SIGINT and SIGTERM each run
// every registered cleanup before the process exits.
func startSignalCleanup() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	installSignalCleanup(sig, runCleanups, os.Exit)
}
