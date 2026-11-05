package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"tricell-hive/tooling/viewer"
)

// refuseSessionCreation fronts the fixture's viewer with a proxy that answers
// every session creation with the given rejection.
func (q *queueCLI) refuseSessionCreation(code string) {
	q.t.Helper()
	real := q.f.info()
	target, _ := url.Parse(real.url())
	proxy := httputil.NewSingleHostReverseProxy(target)
	direct := proxy.Director
	proxy.Director = func(r *http.Request) { direct(r); r.Host = target.Host }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/timeline/sessions" {
			w.Header().Set(viewer.TimelineRevisionHeader, viewer.TimelineRevisionValue)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"code":"` + code + `"}`))
			return
		}
		proxy.ServeHTTP(w, r)
	}))
	q.t.Cleanup(server.Close)
	u, _ := url.Parse(server.URL)
	spy := real
	spy.Port = mustPort(q.t, u.Port())
	paths := newViewPaths(q.f.stateDir)
	if err := writeServerInfo(paths, spy); err != nil {
		q.t.Fatal(err)
	}
	q.t.Cleanup(func() { _ = writeServerInfo(paths, real) })
}

// Finding 2: a queued open the viewer rejects when the queue is sent leaves no
// local handle behind in the current-work file.
func TestWorkQueueRejectedOpenReleasesItsLocalHandle(t *testing.T) {
	q := newQueueCLI(t)
	q.f.mustView("stop")
	queued := decodeQueued(t, q.open("doomed"))
	ids, err := readCurrentWorks(q.f.stateDir, queueHost, queueNative)
	if err != nil || len(ids) != 1 || ids[0] != queued.WorkID {
		t.Fatalf("current works after the queued open: %v %v", ids, err)
	}
	q.f.mustView("start")
	q.refuseSessionCreation("foreign_project")
	_, err = q.run("", "sync", "--work-id", queued.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "Go on")
	t.Logf("sync: %v", err)
	if rejected := q.journal().Queue; len(rejected) == 0 || rejected[0].State != "rejected" {
		t.Logf("journal queue: %d entries", len(rejected))
	}
	ids, err = readCurrentWorks(q.f.stateDir, queueHost, queueNative)
	if err != nil || len(ids) != 0 {
		t.Fatalf("a rejected open left its handle current: %v %v", ids, err)
	}
}
