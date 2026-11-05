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
		// Git fails for the event's read, which runs outside the lock. A concurrent
		// create binds the root to its remote while that read is in flight, that is
		// after the directory was located and before the event takes the lock.
		reads++
		if reads == 1 && !created {
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

// The directory is both the project root and the first ancestor, and it is the
// first candidate that resolves: Git is asked about it once per event.
func TestHostEventAsksGitOnceForAResolvedRoot(t *testing.T) {
	h := newHostEventHarness(t)
	reads := 0
	h.store.remote = func(string) (string, error) {
		reads++
		return "", errors.New("git unavailable")
	}
	h.advance()
	h.mustSend(h.eventFrom(h.source, "activity", TimelineHostEventData{}, "", ""))
	if reads != 1 {
		t.Fatalf("git ran %d times for one event, want 1", reads)
	}
}

// When no candidate resolves, the error is the nearest candidate's: it is the
// most specific cause. A directory that does not exist is "project_unavailable",
// even though the unregistered ancestors above it are "project_not_registered".
func TestHostEventErrorWhenNoCandidateResolvesIsTheNearestOne(t *testing.T) {
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
	requireTimelineCode(t, err, "project_unavailable")
}
