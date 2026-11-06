package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Tests for the session-map rules: nested hosts, pid reuse, changed_at,
// removed Cursor variable and unsafe locations.

func TestResolveViewIdentityIgnoresCursorConversationVariable(t *testing.T) {
	withIdentityDeps(t, map[string]string{"CURSOR_CONVERSATION_ID": "conv-1"})
	if id, err := resolveViewIdentity(viewIdentityDepsFn(), t.TempDir()); !errors.Is(err, errViewIdentityMissing) {
		t.Fatalf("CURSOR_CONVERSATION_ID must not be a source, got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityCursorResolvesThroughTheMap(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	withIdentityDeps(t, nil, 1000, 800)
	writePidMap(t, state, "cursor", 800, now.Add(-time.Hour), pidMapSession{ID: "conv-1", Started: now.Add(-time.Minute)})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.Host != "cursor" || id.NativeSessionID != "conv-1" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

// withProcessNames names processes for the nested-host checks; others are shells.
func withProcessNames(t *testing.T, names map[int]string) {
	t.Helper()
	inner := viewIdentityDepsFn
	viewIdentityDepsFn = func() viewIdentityDeps {
		d := inner()
		d.processName = func(pid int) (string, error) {
			if n, ok := names[pid]; ok {
				return n, nil
			}
			return "zsh", nil
		}
		return d
	}
}

func withProcessStart(t *testing.T, start func(int) time.Time) {
	t.Helper()
	inner := viewIdentityDepsFn
	viewIdentityDepsFn = func() viewIdentityDeps {
		d := inner()
		d.processStart = func(pid int) (time.Time, error) { return start(pid), nil }
		return d
	}
}

func TestResolveViewIdentityNestedHostNeverUsesTheOuterVariable(t *testing.T) {
	now := time.Now()
	// Cursor (800) started from Claude Code's shell inherits CLAUDE_CODE_SESSION_ID
	// and has no variable of its own, so it resolves through its own map.
	// self 1000 -> shell 900 -> cursor 800 -> claude's shell 700 -> claude 600
	for _, c := range []struct {
		name  string
		setup func(t *testing.T, state string)
		want  viewIdentity
		err   error
	}{
		{"nearest map is another host", func(t *testing.T, state string) {
			withProcessNames(t, map[int]string{600: "claude"})
			writePidMap(t, state, "cursor", 800, now.Add(-time.Hour), pidMapSession{ID: "inner", Started: now.Add(-time.Minute)})
		}, viewIdentity{Host: "cursor", NativeSessionID: "inner"}, nil},
		// Nothing above cursor explains CLAUDE_CODE_SESSION_ID: it may belong to
		// an inner host the walk did not recognize, so the map decides nothing.
		{"origin of the variable not found above", func(t *testing.T, state string) {
			writePidMap(t, state, "cursor", 800, now.Add(-time.Hour), pidMapSession{ID: "inner", Started: now.Add(-time.Minute)})
		}, viewIdentity{}, errViewIdentityAmbiguous},
		{"nearest host process has no map yet", func(t *testing.T, state string) {
			withProcessNames(t, map[int]string{800: "/usr/local/bin/cursor-agent", 600: "claude"})
			writePidMap(t, state, "claude-code", 600, now.Add(-time.Hour), pidMapSession{ID: "outer", Started: now.Add(-time.Minute)})
		}, viewIdentity{}, errViewIdentityMissing},
	} {
		t.Run(c.name, func(t *testing.T) {
			state := t.TempDir()
			withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "outer"}, 1000, 900, 800, 700, 600)
			c.setup(t, state)
			id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
			if id != c.want || !errors.Is(err, c.err) {
				t.Fatalf("a nested host must not publish as the outer session: got %+v, %v; want %+v, %v", id, err, c.want, c.err)
			}
		})
	}
}

func TestResolveViewIdentityVariablesStayAmbiguousWhenTheirOriginIsUnknown(t *testing.T) {
	now := time.Now()
	for _, c := range []struct {
		name  string
		setup func(t *testing.T, state string)
	}{
		// Pi runs under node, but its command line cannot be read: nothing above
		// claude explains PI_SESSION_ID, so it may belong to an inner host.
		{"inner host not recognized", func(t *testing.T, state string) {
			withProcessNames(t, map[int]string{800: "node", 600: "claude"})
		}},
		// Pi (unrecognized node) was started from a Grok shell: Grok is the
		// nearest recognized host and has no variable, but nothing explains
		// PI_SESSION_ID, so Grok's map must not speak for it.
		{"unrecognized inner host under a host without a variable", func(t *testing.T, state string) {
			withProcessNames(t, map[int]string{800: "node", 600: "grok"})
			writePidMap(t, state, "grok", 600, now.Add(-time.Hour), pidMapSession{ID: "gk", Started: now.Add(-time.Minute)})
		}},
		// A map that cannot be trusted at the nearest ancestor decides nothing.
		{"untrusted nearest map", func(t *testing.T, state string) {
			path := writePidMap(t, state, "claude-code", 800, now.Add(-time.Hour), pidMapSession{ID: "cc", Started: now.Add(-time.Minute)})
			if err := os.Chmod(path, 0666); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			state := t.TempDir()
			withIdentityDeps(t, map[string]string{"PI_SESSION_ID": "pi-1", "CLAUDE_CODE_SESSION_ID": "cc"}, 1000, 900, 800, 700, 600)
			oldArgs := hookProcessArgs
			hookProcessArgs = func(int) (string, error) { return "", errors.New("unreadable") }
			t.Cleanup(func() { hookProcessArgs = oldArgs })
			c.setup(t, state)
			if id, err := resolveViewIdentity(viewIdentityDepsFn(), state); !errors.Is(err, errViewIdentityAmbiguous) {
				t.Fatalf("got %+v, %v; want ambiguous", id, err)
			}
		})
	}
}

func TestResolveViewIdentityCodexThreadUsesItsVariableInASharedDaemon(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	// The app-server (800) runs two threads; thread-b prompted last, so the map
	// names it, but this shell belongs to thread-a.
	withIdentityDeps(t, map[string]string{"CODEX_SESSION_ID": "thread-a"}, 1000, 900, 800)
	withProcessNames(t, map[int]string{800: "/Users/u/.codex/packages/app-server-daemon/releases/0.167.0/bin/codex"})
	writePidMap(t, state, "codex", 800, now.Add(-time.Minute), pidMapSession{ID: "thread-b", Started: now.Add(-time.Minute)})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.Host != "codex" || id.NativeSessionID != "thread-a" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityNearestHostChoosesAmongVariables(t *testing.T) {
	now := time.Now()
	for _, c := range []struct {
		name  string
		env   map[string]string
		names map[int]string
		args  map[int]string
		maps  func(t *testing.T, state string)
		want  viewIdentity
	}{
		{"claude started from a codex shell",
			map[string]string{"CLAUDE_CODE_SESSION_ID": "cc", "CODEX_SESSION_ID": "cx"},
			map[int]string{800: "claude", 600: "codex"}, nil, nil,
			viewIdentity{Host: "claude-code", NativeSessionID: "cc"}},
		{"pi under node started from a claude shell",
			map[string]string{"PI_SESSION_ID": "pi-1", "CLAUDE_CODE_SESSION_ID": "cc"},
			map[int]string{800: "node", 600: "claude"}, map[int]string{800: "node /Users/u/.local/lib/pi/dist/bin/pi.js"}, nil,
			viewIdentity{Host: "pi", NativeSessionID: "pi-1"}},
		{"grok started from a codex shell",
			map[string]string{"CODEX_SESSION_ID": "cx"},
			map[int]string{800: "grok", 600: "codex"}, nil,
			func(t *testing.T, state string) {
				writePidMap(t, state, "grok", 800, now.Add(-time.Hour), pidMapSession{ID: "gk", Started: now.Add(-time.Minute)})
			},
			viewIdentity{Host: "grok", NativeSessionID: "gk"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			state := t.TempDir()
			// self 1000 -> shell 900 -> inner host 800 -> outer shell 700 -> outer host 600
			withIdentityDeps(t, c.env, 1000, 900, 800, 700, 600)
			withProcessNames(t, c.names)
			oldArgs := hookProcessArgs
			hookProcessArgs = func(pid int) (string, error) { return c.args[pid], nil }
			t.Cleanup(func() { hookProcessArgs = oldArgs })
			if c.maps != nil {
				c.maps(t, state)
			}
			id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
			if err != nil || id != c.want {
				t.Fatalf("got %+v, %v; want %+v", id, err, c.want)
			}
		})
	}
}

func TestResolveViewIdentityKeepsVariableWhenItsOwnHostIsNearest(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	// claude (800) is the nearest host; a codex (700) runs above it.
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "mine"}, 1000, 900, 800, 700)
	writePidMap(t, state, "claude-code", 800, now.Add(-time.Hour), pidMapSession{ID: "mine", Started: now.Add(-time.Minute)})
	writePidMap(t, state, "codex", 700, now.Add(-time.Hour), pidMapSession{ID: "far", Started: now.Add(-time.Minute)})
	withProcessNames(t, map[int]string{800: "claude", 700: "codex"})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.Host != "claude-code" || id.NativeSessionID != "mine" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

func TestHostOfProcessNameMapsToCanonicalHosts(t *testing.T) {
	for name, host := range map[string]string{"codex": "codex", "grok": "grok", "opencode": "opencode", "pi": "pi", "cursor-agent": "cursor", "claude": "claude-code"} {
		if got := hostOfProcessName("/opt/bin/" + name); got != host {
			t.Errorf("%s: got %q, want %q", name, got, host)
		}
	}
	if hostOfProcessName("zsh") != "" || hostOfProcessName("") != "" {
		t.Error("a shell is not a host")
	}
}

func TestResolveViewIdentityTreatsReusedPidAsNoMap(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	// The crashed host's pid 800 now belongs to a process started later; a live
	// host with a fresh map sits at 700.
	withIdentityDeps(t, nil, 1000, 800, 700)
	writePidMap(t, state, "codex", 800, now.Add(-time.Hour), pidMapSession{ID: "dead", Started: now.Add(-time.Minute)})
	writePidMap(t, state, "grok", 700, now.Add(-time.Hour), pidMapSession{ID: "alive", Started: now.Add(-time.Minute)})
	withProcessStart(t, func(pid int) time.Time {
		if pid == 800 {
			return identityTestStart.Add(3 * time.Hour)
		}
		return identityTestStart
	})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.NativeSessionID != "alive" {
		t.Fatalf("a map from a reused pid must be skipped, got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityReusedPidAloneIsMissing(t *testing.T) {
	state := t.TempDir()
	withIdentityDeps(t, nil, 1000, 800)
	writePidMap(t, state, "codex", 800, time.Now().Add(-time.Hour), pidMapSession{ID: "dead", Started: time.Now()})
	withProcessStart(t, func(int) time.Time { return identityTestStart.Add(time.Hour) })
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), state); !errors.Is(err, errViewIdentityMissing) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveViewIdentityMapWithoutProcessStartIsRejected(t *testing.T) {
	state := t.TempDir()
	withIdentityDeps(t, nil, 1000, 800)
	p := writePidMap(t, state, "codex", 800, time.Now().Add(-time.Hour), pidMapSession{ID: "x", Started: time.Now()})
	raw, _ := os.ReadFile(p)
	var doc map[string]any
	json.Unmarshal(raw, &doc)
	delete(doc, "process_started_at")
	raw, _ = json.Marshal(doc)
	os.WriteFile(p, raw, 0600)
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), state); !errors.Is(err, errViewIdentityAmbiguous) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveViewIdentityToleratesSubSecondStartRounding(t *testing.T) {
	state := t.TempDir()
	withIdentityDeps(t, nil, 1000, 800)
	writePidMap(t, state, "codex", 800, time.Now().Add(-time.Hour), pidMapSession{ID: "x", Started: time.Now()})
	withProcessStart(t, func(int) time.Time { return identityTestStart.Add(900 * time.Millisecond) })
	if id, err := resolveViewIdentity(viewIdentityDepsFn(), state); err != nil || id.NativeSessionID != "x" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

func TestParsePsLstart(t *testing.T) {
	got, err := parsePsLstart("Tue Nov  4 10:05:09 2026\n")
	want := time.Date(2026, 11, 4, 10, 5, 9, 0, time.Local)
	if err != nil || !got.Equal(want) {
		t.Fatalf("got %v, %v", got, err)
	}
	if _, err := parsePsLstart("garbage"); err == nil {
		t.Fatal("garbage must fail")
	}
}

func TestPsProcessStartOfThisProcess(t *testing.T) {
	got, err := psProcessStart(os.Getpid())
	if err != nil || got.After(time.Now().Add(time.Minute)) || got.Before(time.Now().Add(-24*time.Hour)) {
		t.Fatalf("got %v, %v", got, err)
	}
}

// changed_at moves only when a session replaces another in the same process.
func TestResolveViewIdentitySecondSessionAlongsideStaysAmbiguous(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	withIdentityDeps(t, nil, 1000, 800)
	// changed_at equals the first start and did not move when the second started.
	writePidMap(t, state, "codex", 800, now.Add(-2*time.Minute),
		pidMapSession{ID: "main", Started: now.Add(-2 * time.Minute)},
		pidMapSession{ID: "sub", Started: now.Add(-time.Minute)})
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), state); !errors.Is(err, errViewIdentityAmbiguous) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveViewIdentityResumeIntoAnotherSessionResolvesTheNewOne(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	withIdentityDeps(t, nil, 1000, 800)
	// The old session never got ended_at; the replacement moved changed_at to its own start.
	writePidMap(t, state, "claude-code", 800, now.Add(-time.Minute),
		pidMapSession{ID: "old", Started: now.Add(-time.Hour)},
		pidMapSession{ID: "new", Started: now.Add(-time.Minute)})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.NativeSessionID != "new" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityRejectsUnsafeMapPaths(t *testing.T) {
	now := time.Now()
	for name, prepare := range map[string]func(t *testing.T, state, file string){
		"map group-writable":       func(t *testing.T, state, file string) { os.Chmod(file, 0620) },
		"map world-writable":       func(t *testing.T, state, file string) { os.Chmod(file, 0602) },
		"sessions dir writable":    func(t *testing.T, state, file string) { os.Chmod(filepath.Dir(file), 0777) },
		"hooks dir group-writable": func(t *testing.T, state, file string) { os.Chmod(filepath.Join(state, "hooks"), 0770) },
		"hooks dir is a symlink": func(t *testing.T, state, file string) {
			real := filepath.Join(t.TempDir(), "hooks")
			os.Rename(filepath.Join(state, "hooks"), real)
			os.Symlink(real, filepath.Join(state, "hooks"))
		},
		"sessions dir is a symlink": func(t *testing.T, state, file string) {
			real := filepath.Join(t.TempDir(), "sessions")
			os.Rename(filepath.Dir(file), real)
			os.Symlink(real, filepath.Dir(file))
		},
	} {
		t.Run(name, func(t *testing.T) {
			state := t.TempDir()
			withIdentityDeps(t, nil, 1000, 800)
			file := writePidMap(t, state, "codex", 800, now.Add(-time.Hour), pidMapSession{ID: "x", Started: now})
			prepare(t, state, file)
			if id, err := resolveViewIdentity(viewIdentityDepsFn(), state); !errors.Is(err, errViewIdentityAmbiguous) {
				t.Fatalf("resolved %+v from an unsafe location (%v)", id, err)
			}
		})
	}
}

func TestCurrentWorkRejectsUnsafeFolders(t *testing.T) {
	for name, prepare := range map[string]func(t *testing.T, cur string){
		"current group-writable": func(t *testing.T, cur string) { os.Chmod(cur, 0770) },
		"host folder writable":   func(t *testing.T, cur string) { os.Chmod(filepath.Join(cur, "codex"), 0777) },
		"host folder symlink": func(t *testing.T, cur string) {
			real := filepath.Join(t.TempDir(), "codex")
			os.Rename(filepath.Join(cur, "codex"), real)
			os.Symlink(real, filepath.Join(cur, "codex"))
		},
	} {
		t.Run(name, func(t *testing.T) {
			state := t.TempDir()
			if err := writeCurrentWorks(state, "codex", "s1", []string{workA}); err != nil {
				t.Fatal(err)
			}
			cur := filepath.Join(state, "publisher", "v1", "current")
			prepare(t, cur)
			if _, err := readCurrentWorks(state, "codex", "s1"); err == nil {
				t.Fatal("read from an unsafe folder")
			}
			if err := writeCurrentWorks(state, "codex", "s1", []string{workB}); err == nil {
				t.Fatal("wrote into an unsafe folder")
			}
		})
	}
}
