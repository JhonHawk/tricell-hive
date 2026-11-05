package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"sync"
	"testing"
	"time"
	"tricell-hive/tooling/viewer"
)

// postProbe fronts the fixture's viewer and calls onPost(n, event) before each
// POST /api/host-events is forwarded (n counts from 1), and records the order in
// which the events arrived.
type postProbe struct {
	mu     sync.Mutex
	titles []string
	queued [][]string // the queue as it was when each POST arrived
}

func (p *postProbe) arrived() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string{}, p.titles...)
}

func (h *hookFixture) probeViewer(onPost func(n int)) *postProbe {
	h.t.Helper()
	real := h.f.info()
	target, _ := url.Parse(real.url())
	proxy := httputil.NewSingleHostReverseProxy(target)
	direct := proxy.Director
	proxy.Director = func(r *http.Request) { direct(r); r.Host = target.Host }
	p := &postProbe{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/host-events" {
			raw, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(raw))
			var e viewer.TimelineHostEvent
			_ = json.Unmarshal(raw, &e)
			queued := hostEventQueueNames(filepath.Join(h.f.stateDir, "hooks", "queue"))
			p.mu.Lock()
			p.titles = append(p.titles, e.Data.Title)
			p.queued = append(p.queued, queued)
			n := len(p.titles)
			p.mu.Unlock()
			if onPost != nil {
				onPost(n)
			}
		}
		proxy.ServeHTTP(w, r)
	}))
	h.t.Cleanup(server.Close)
	u, _ := url.Parse(server.URL)
	spy := real
	spy.Port = mustPort(h.t, u.Port())
	paths := newViewPaths(h.f.stateDir)
	if err := writeServerInfo(paths, spy); err != nil {
		h.t.Fatal(err)
	}
	h.t.Cleanup(func() { _ = writeServerInfo(paths, real) })
	return p
}

func (h *hookFixture) holdQueueLock() {
	h.t.Helper()
	if _, err := ensureHookDir(h.f.stateDir, "queue"); err != nil {
		h.t.Fatal(err)
	}
	l, err := acquirePrivateWorkLock(filepath.Join(h.f.stateDir, "hooks", "queue.lock"), false)
	if err != nil {
		h.t.Fatal(err)
	}
	h.t.Cleanup(func() { _ = l.Release() })
}

func (h *hookFixture) runTodos() {
	h.t.Helper()
	mustHook(h.t, h.f.stateDir, fmt.Sprintf(openCodeTodos, h.project, h.at), "opencode", "todowrite")
}

// H40: another run holds the queue lock; this run does not wait for it and
// delivers its own events, in emission order.
func TestHookDeliversDirectlyWhileAnotherRunHoldsTheQueueLock(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("busy-lock")
	probe := h.probeViewer(nil)
	h.holdQueueLock()
	start := time.Now()
	h.runTodos()
	if time.Since(start) > time.Second {
		t.Fatal("the run waited for the queue lock")
	}
	if got := probe.arrived(); fmt.Sprint(got) != "[First Second Third]" {
		t.Fatalf("events arrived %v, want all three in emission order", got)
	}
	if got := h.queued(); len(got) != 0 {
		t.Fatalf("queued %v", got)
	}
}

// H43: with the viewer up an event reaches the viewer before anything is written
// to the queue, and the queue stays empty.
func TestHookWritesNothingToTheQueueWhileTheViewerIsUp(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("no-queue-write")
	probe := h.probeViewer(nil)
	h.runTodos()
	if got := probe.arrived(); len(got) != 3 {
		t.Fatalf("%d events arrived, want 3", len(got))
	}
	for i, q := range probe.queued {
		if len(q) != 0 {
			t.Fatalf("event %d: the queue held %v while the viewer was being called", i+1, q)
		}
	}
	if got := h.queued(); len(got) != 0 {
		t.Fatalf("queued %v", got)
	}
}

// With too little budget left to try, nothing is sent and every event is queued.
func TestHookQueuesWithoutTryingWhenTheBudgetIsTooShort(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("budget-short")
	probe := h.probeViewer(nil)
	oldBudget := hookBudget
	hookBudget = 50 * time.Millisecond // below the least budget worth a delivery attempt
	t.Cleanup(func() { hookBudget = oldBudget })
	h.runTodos()
	time.Sleep(300 * time.Millisecond) // the run past its budget must not post either
	if got := probe.arrived(); len(got) != 0 {
		t.Fatalf("a run with no budget posted %v", got)
	}
	if got := h.queued(); len(got) != 3 {
		t.Fatalf("%d events queued, want 3", len(got))
	}
}

// The budget ends while the second event hangs: the first was delivered, and the
// second and third stay queued, in order.
func TestHookBudgetEndedMidRunLeavesTheUndeliveredEventsQueued(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("budget-mid-run")
	release := make(chan struct{})
	probe := h.probeViewer(func(n int) {
		if n >= 2 {
			select {
			case <-release:
			case <-time.After(5 * time.Second):
			}
		}
	})
	oldBudget := hookBudget
	hookBudget = 600 * time.Millisecond
	t.Cleanup(func() { hookBudget = oldBudget })
	h.runTodos()
	got := h.queued()
	close(release)
	time.Sleep(500 * time.Millisecond) // let the abandoned run drain before the state is removed
	if len(got) != 2 {
		t.Fatalf("%d events queued, want the 2 undelivered: %v", len(got), got)
	}
	if arrived := probe.arrived(); len(arrived) == 0 || arrived[0] != "First" {
		t.Fatalf("arrived %v", arrived)
	}
	if got[0] >= got[1] {
		t.Fatalf("queue not in emission order: %v", got)
	}
}
