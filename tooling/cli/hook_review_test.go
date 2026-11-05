package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
	"tricell-hive/integrations/hostevent"
)

// A Claude Code session started from a Cursor terminal inherits Cursor's
// variables; they must not turn its hooks into Cursor's.
func TestDetectHookHostIgnoresInheritedCursorEnvironment(t *testing.T) {
	env := func(k string) string {
		if k == "CURSOR_VERSION" || k == "CURSOR_PROJECT_DIR" {
			return "inherited"
		}
		return ""
	}
	claudePayload := hostevent.Decode([]byte(`{"session_id":"s","cwd":"/p","hook_event_name":"UserPromptSubmit"}`))
	if got := detectHookHost("claude", claudePayload, env); got != "claude-code" {
		t.Fatalf("host = %q, want claude-code", got)
	}
	cursorPayload := hostevent.Decode([]byte(`{"conversation_id":"c","workspace_roots":["/p"]}`))
	if got := detectHookHost("claude", cursorPayload, func(string) string { return "" }); got != "cursor" {
		t.Fatalf("host = %q, want cursor", got)
	}
}

// slowViewer fronts the fixture's viewer with a proxy that counts the requests
// by route and, when asked, holds every POST /api/host-events until release is
// closed.
type slowViewer struct {
	mu      sync.Mutex
	counts  map[string]int
	release chan struct{}
}

func (v *slowViewer) count(route string) int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.counts[route]
}

func (h *hookFixture) frontViewer(hold bool) *slowViewer {
	h.t.Helper()
	real := h.f.info()
	target, _ := url.Parse(real.url())
	proxy := httputil.NewSingleHostReverseProxy(target)
	direct := proxy.Director
	proxy.Director = func(r *http.Request) { direct(r); r.Host = target.Host }
	v := &slowViewer{counts: map[string]int{}, release: make(chan struct{})}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := r.Method + " " + r.URL.Path
		if strings.HasPrefix(r.URL.Path, "/api/timeline/sessions/") {
			route = "GET session"
		}
		v.mu.Lock()
		v.counts[route]++
		v.mu.Unlock()
		if hold && r.URL.Path == "/api/host-events" {
			select {
			case <-v.release:
			case <-time.After(5 * time.Second):
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
	return v
}

const openCodeTodos = `{"session_id":"oc-review-1","directory":"%s","occurred_at":"%s","todos":[` +
	`{"content":"First","status":"pending"},{"content":"Second","status":"in_progress"},{"content":"Third","status":"completed"}]}`

// Finding 1: when the budget runs out with the viewer slow, every translated
// event is already in the queue; none is lost.
func TestHookQueuesEveryTranslatedEventBeforeItDelivers(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("budget")
	v := h.frontViewer(true)
	oldBudget := hookBudget
	hookBudget = 300 * time.Millisecond
	t.Cleanup(func() { hookBudget = oldBudget })
	payload := fmt.Sprintf(openCodeTodos, h.project, h.at)
	mustHook(t, h.f.stateDir, payload, "opencode", "todowrite")
	got := len(h.queued())
	close(v.release)
	time.Sleep(500 * time.Millisecond) // let the abandoned run drain before the state is removed
	if got != 3 {
		t.Fatalf("%d events queued when the budget ran out, want 3", got)
	}
}

// Finding 2: a current work the viewer reports closed or not found is dropped
// from the current-work file, so the session's one live work is named again.
func TestHookDropsCurrentWorksTheViewerReportsClosedOrGone(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("live")
	other := decodeSync(t, h.must("", "open", "--project", h.project, "--key", "finished-elsewhere", "--title", "Finished", "--summary", "Closed by someone else")).WorkID
	h.must("", "close", "--work-id", other, "--outcome", "completed", "--reason", "Done")
	gone := strings.Repeat("ab", 32)
	for _, id := range []string{other, gone} {
		if err := addCurrentWork(h.f.stateDir, "claude-code", hookSyncSession, id); err != nil {
			t.Fatal(err)
		}
	}
	h.hook("TaskCreated", h.payload(`,"task_id":"7","task_subject":"Named again"`))
	ids, err := readCurrentWorks(h.f.stateDir, "claude-code", hookSyncSession)
	if err != nil || len(ids) != 1 || ids[0] != h.work {
		t.Fatalf("current works %v %v, want only %s", ids, err, h.work)
	}
	if h.task("host:7") == nil {
		t.Fatal("the live work was not named once the others were dropped")
	}
}

// Finding 5: one health check and one work read serve a whole run.
func TestHookResolvesTheViewerAndTheWorkOncePerRun(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("once")
	v := h.frontViewer(false)
	payload := fmt.Sprintf(openCodeTodos, h.project, h.at)
	mustHook(t, h.f.stateDir, payload, "opencode", "todowrite")
	if n := v.count("GET /api/health"); n != 1 {
		t.Fatalf("%d health checks, want 1", n)
	}
	if n := v.count("GET session"); n > 1 {
		t.Fatalf("%d session reads for 3 events, want at most 1", n)
	}
	if n := v.count("POST /api/host-events"); n != 3 {
		t.Fatalf("%d events posted, want 3", n)
	}
	if len(h.queued()) != 0 {
		t.Fatalf("queue not empty: %v", h.queued())
	}
}
