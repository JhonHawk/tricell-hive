package main

import (
	"testing"
)

// Defect 1: a rejected non-open entry releases the real work its handle maps to,
// and not_found counts as the viewer reporting the work gone.
func TestWorkQueueRejectionReleasesTheRealWorkBehindTheHandle(t *testing.T) {
	for _, code := range []string{"session_closed", "session_deleted", "work_closed", "not_found"} {
		q := newQueueCLI(t)
		state := q.f.stateDir
		const real = "1111111111111111111111111111111111111111111111111111111111111111"
		const handle = "queued-handle"
		if err := addCurrentWork(state, queueHost, queueNative, real); err != nil {
			t.Fatal(err)
		}
		e := &viewWorkEngine{state: state, journal: newWorkJournal()}
		e.journal.Aliases = append(e.journal.Aliases, viewWorkAlias{Alias: handle, WorkID: real})
		entry := &viewWorkQueued{ID: "op-1", Action: "sync", WorkID: handle}
		c := viewWorkCommand{}
		c.Host, c.NativeSessionID = queueHost, queueNative
		if err := e.releaseRejectedWork(entry, code, c); err != nil {
			t.Fatal(err)
		}
		ids, err := readCurrentWorks(state, queueHost, queueNative)
		if err != nil || len(ids) != 0 {
			t.Fatalf("%s: the real work stayed current: %v %v", code, ids, err)
		}
	}
}
