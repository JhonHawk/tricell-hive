//go:build unix

package viewer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A lock file may be removed by its holder; a process that opened it before
// the removal must not end up holding the unlinked file while another process
// locks a new one at the same path.
func TestAcquireLockNeverHoldsAFileRemovedWhileWaiting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.lock")
	holder, err := AcquireLock(path, false)
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		lock *FileLock
		err  error
	}
	done := make(chan result, 1)
	go func() {
		l, err := AcquireLock(path, true)
		done <- result{l, err}
	}()
	time.Sleep(100 * time.Millisecond) // the waiter has opened the file and blocks on it
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	holder.Release()
	r := <-done
	if r.err != nil {
		t.Fatal(r.err)
	}
	defer r.lock.Release()
	onDisk, err := os.Stat(path)
	if err != nil {
		t.Fatalf("the waiter must hold the file at the path: %v", err)
	}
	held, err := r.lock.f.Stat()
	if err != nil || !os.SameFile(onDisk, held) {
		t.Fatalf("the waiter holds an unlinked file (%v)", err)
	}
	if _, err := AcquireLock(path, false); err != ErrLocked {
		t.Fatalf("a second holder got the lock: %v", err)
	}
}
