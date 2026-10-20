package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

func statusSource(t *testing.T, status string) string {
	t.Helper()
	folder := filepath.Join(t.TempDir(), "source")
	copyTree(t, walkthroughFixture("valid/final"), folder)
	setStatusSource(t, folder, status)
	return folder
}

func setStatusSource(t *testing.T, folder, status string) {
	t.Helper()
	path := filepath.Join(folder, "walkthrough.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if status == "" {
		delete(m["session"].(map[string]any), "status")
	} else {
		m["session"].(map[string]any)["status"] = status
	}
	data, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestViewStatusDirectoryExactRevision(t *testing.T) {
	for _, name := range []string{"closed", "active", "historical", "excluded", "canonical-alias", "other-source", "other-session", "other-key", "other-hash", "other-status", "legacy", "resource-changed", "reopened", "mutation-during-query", "resource-mutation-during-query", "removed-during-query", "index-denied", "index-malformed", "invalid-source", "source-missing", "offline"} {
		t.Run(name, func(t *testing.T) {
			status := "closed"
			if name == "active" {
				status = "in_progress"
			}
			if name == "historical" {
				status = ""
			}
			folder := statusSource(t, status)
			fp, err := viewer.FingerprintSource(folder)
			if err != nil {
				t.Fatal(err)
			}
			sess := viewer.Session{Source: fp.Source, SessionID: fp.SessionID, Key: fp.Key, ContentHash: fp.ContentHash, SessionStatus: fp.SessionStatus}
			wantFound := name == "closed" || name == "active" || name == "historical" || name == "excluded" || name == "canonical-alias"
			switch name {
			case "other-source":
				sess.Source += "-other"
			case "other-session":
				sess.SessionID += "-other"
			case "other-key":
				sess.Key = strings.Repeat("0", 64)
			case "other-hash":
				sess.ContentHash = strings.Repeat("0", 64)
			case "other-status":
				sess.SessionStatus = "in_progress"
			case "legacy":
				sess.SessionStatus = ""
			case "reopened":
				setStatusSource(t, folder, "in_progress")
			case "resource-changed":
				mutateStatusResource(t, folder)
			case "invalid-source":
				setStatusSource(t, folder, "invalid")
			case "source-missing":
				if err := os.Remove(filepath.Join(folder, "walkthrough.json")); err != nil {
					t.Fatal(err)
				}
			}
			state, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			p := newViewPaths(state)
			if err := os.MkdirAll(p.dir, 0700); err != nil {
				t.Fatal(err)
			}
			if name != "offline" {
				lock, err := viewer.AcquireLock(p.lock, false)
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Release()
				const token = "fixture-status-access-token"
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("X-Hive-Token") != token {
						t.Error("missing authentication")
						http.Error(w, "denied", 401)
						return
					}
					if r.URL.Path == "/api/health" {
						json.NewEncoder(w).Encode(map[string]any{"pid": os.Getpid(), "state_dir": p.state})
						return
					}
					if r.URL.Path != "/api/index" {
						t.Error("unexpected request")
						http.NotFound(w, r)
						return
					}
					if name == "index-denied" {
						http.Error(w, "fixture-status-access-token", http.StatusForbidden)
						return
					}
					if name == "index-malformed" {
						w.Write([]byte("fixture-status-access-token"))
						return
					}
					if name == "mutation-during-query" {
						setStatusSource(t, folder, "in_progress")
					}
					if name == "resource-mutation-during-query" {
						mutateStatusResource(t, folder)
					}
					if name == "removed-during-query" {
						if err := os.Remove(filepath.Join(folder, "walkthrough.json")); err != nil {
							t.Error(err)
						}
					}
					v := viewer.View{Projects: []viewer.Project{{Tickets: []viewer.Ticket{{Sessions: []viewer.Session{sess}}}}}}
					if name == "excluded" {
						v.Projects = nil
						v.ExcludedSessions = []viewer.Session{sess}
					}
					json.NewEncoder(w).Encode(v)
				}))
				defer srv.Close()
				if err := writeServerInfo(p, serverInfo{PID: os.Getpid(), Port: srv.Listener.Addr().(*net.TCPAddr).Port, Token: token, StateDir: p.state}); err != nil {
					t.Fatal(err)
				}
			}
			var out bytes.Buffer
			queryDir := folder
			if name == "canonical-alias" {
				queryDir = filepath.Join(t.TempDir(), "alias")
				if err := os.Symlink(folder, queryDir); err != nil {
					t.Fatal(err)
				}
			}
			if err := runView([]string{"status", "--dir", queryDir, "--state-dir", p.state}, &out, viewDeps{}); err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			wantStatus := "unknown"
			if wantFound {
				wantStatus = fp.SessionStatus
			}
			if got["found"] != wantFound || got["session_status"] != wantStatus {
				t.Fatalf("confirmation = %v/%v, want %v/%v", got["found"], got["session_status"], wantFound, wantStatus)
			}
			if strings.Contains(out.String(), "fixture-status-access-token") || strings.Contains(out.String(), "capability") {
				t.Fatal("status exposed credentials")
			}
			if _, ok := got["running"]; !ok {
				t.Fatal("normal output missing")
			}
		})
	}
}

func mutateStatusResource(t *testing.T, folder string) {
	t.Helper()
	m, err := viewer.ValidateManifest(folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) == 0 {
		t.Fatal("missing resource fixture")
	}
	if err := os.WriteFile(filepath.Join(folder, m.Files[0]), []byte("changed resource"), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestViewStatusWithoutDirectoryKeepsOutputAndDoesNotCreateStore(t *testing.T) {
	state := filepath.Join(t.TempDir(), "unused-state")
	var out bytes.Buffer
	if err := runView([]string{"status", "--state-dir", state}, &out, viewDeps{}); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got["running"] != false {
		t.Fatal("normal offline output changed")
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatalf("status created state: %v", err)
	}
}

func TestViewStatusIndexDoesNotRedirectCredentials(t *testing.T) {
	targetCalls := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { targetCalls++ }))
	defer target.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()
	if _, err := viewIndex(serverInfo{Port: srv.Listener.Addr().(*net.TCPAddr).Port, Token: "fixture-status-access-token"}); err == nil {
		t.Fatal("index redirect accepted")
	}
	if targetCalls != 0 {
		t.Fatal("index followed redirect")
	}
}

func TestViewStatusHelpDescribesDirectoryConfirmation(t *testing.T) {
	var out bytes.Buffer
	if err := runView([]string{"--help"}, &out, viewDeps{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hive view status [--dir DIR]") || !strings.Contains(out.String(), "current source revision") {
		t.Fatal("help must describe directory confirmation")
	}
}
