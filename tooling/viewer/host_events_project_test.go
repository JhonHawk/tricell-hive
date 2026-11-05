package viewer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// H41: the key the stored works bind is read under the store lock, so a work
// created for the root after the directories were resolved is still found.
func TestHostEventSeesAWorkBoundBetweenResolvingAndTheLock(t *testing.T) {
	h := newHostEventHarness(t)
	other := t.TempDir()
	if err := os.MkdirAll(filepath.Join(other, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := registerRepository(h.store.vdir, other, h.clock); err != nil {
		t.Fatal(err)
	}
	source := h.source
	source.ProjectRoot = other
	create := h.request
	create.OperationID, create.Source = "create-other", source
	creating, created, reads := false, false, 0
	var work TimelineSession
	h.store.remote = func(string) (string, error) {
		if creating {
			return "git@github.com:acme/other.git", nil
		}
		// Git fails for every read of the event, which runs outside the lock. A
		// concurrent create binds the root to its remote during the last read: the
		// event's directory is both its project root and the first ancestor, so
		// that is the second read.
		reads++
		if reads == 2 && !created {
			created, creating = true, true
			h.clock = h.clock.Add(-time.Minute)
			work = createTimeline(t, h.store, create)
			h.clock = h.clock.Add(time.Minute)
			creating = false
		}
		return "", errors.New("git unavailable")
	}
	h.advance()
	r := h.mustSend(h.eventFrom(source, "activity", TimelineHostEventData{}, "", ""))
	if work.ViewerSessionID == "" {
		t.Fatal("the concurrent create never ran")
	}
	if r.Status != "applied" {
		t.Fatalf("event after a concurrent bind: %+v", r)
	}
	if got := domainCurrent(t, h.store, work.ViewerSessionID).LastActivityAt; got == work.LastActivityAt {
		t.Fatalf("the work did not take the activity: %s", got)
	}
}

// The error of a directory that resolves to no project is the one of the
// farthest candidate, not the nearest: a directory that does not exist under an
// unregistered one ends at "not registered", as before.
func TestHostEventErrorWhenNoCandidateResolvesIsTheFarthestOne(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	missing := h.source
	missing.ProjectRoot = filepath.Join(t.TempDir(), "gone", "deeper")
	h.advance()
	event, err := DecodeTimelineHostEvent(h.eventFrom(missing, "activity", TimelineHostEventData{}, id, stageID))
	if err != nil {
		t.Fatal(err)
	}
	_, err = h.store.TimelineHostEvent(event)
	requireTimelineCode(t, err, "project_not_registered")
}
