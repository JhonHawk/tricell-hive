//go:build unix

package viewer

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// ErrLocked is returned by a non-blocking AcquireLock when another open file
// description holds the lock.
var ErrLocked = errors.New("lock is held by another process")

// FileLock is an exclusive advisory flock(2) on a file. The kernel drops it
// when the holder exits, so a crashed process never leaves a stale lock.
type FileLock struct{ f *os.File }

// AcquireLock takes an exclusive flock on path, creating the file (mode
// 0600) when needed; its directory must exist. With wait it blocks until the
// lock is free (the cross-process guard of viewer/lock); without it a held
// lock returns ErrLocked at once (viewer/server.lock, "is a server running").
func AcquireLock(path string, wait bool) (*FileLock, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	how := syscall.LOCK_EX
	if !wait {
		how |= syscall.LOCK_NB
	}
	for {
		err = syscall.Flock(int(f.Fd()), how)
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		f.Close()
		if !wait && (err == syscall.EWOULDBLOCK || err == syscall.EAGAIN) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}
	return &FileLock{f: f}, nil
}

// Release drops the lock. It is safe to call more than once.
func (l *FileLock) Release() error {
	if l == nil || l.f == nil {
		return nil
	}
	f := l.f
	l.f = nil
	// Closing the descriptor releases the flock.
	return f.Close()
}
