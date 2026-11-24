package management

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// Antigravity's hooks.json keeps each tool's hooks under a top-level name. The
// fixture is a synthetic copy of the shape herdr leaves there (a flat
// PreInvocation list); the real file is never read or written.

const foreignAntigravityHooks = `{
  "herdr": {
    "PreInvocation": [
      { "command": "herdr agent-state working", "timeout": 10, "type": "command" }
    ]
  }
}
`

func agyHooksPath(o Options) string { return filepath.Join(antigravityHome(o), "hooks.json") }

// memberText is the exact text of one top-level member of a JSON object.
func memberText(t *testing.T, doc, key string) string {
	t.Helper()
	root, err := parseJSONSpans([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	m, err := root.member(key)
	if err != nil || m == nil {
		t.Fatalf("no member %q in %s (%v)", key, doc, err)
	}
	return doc[m.start:m.val.end]
}

type agyHandler struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

func agyHiveHook(t *testing.T, path string) map[string][]agyHandler {
	t.Helper()
	var doc map[string]json.RawMessage
	if err := json.Unmarshal([]byte(get(t, path)), &doc); err != nil {
		t.Fatal(err)
	}
	raw, ok := doc["hive"]
	if !ok {
		return nil
	}
	var hook map[string][]agyHandler
	if err := json.Unmarshal(raw, &hook); err != nil {
		t.Fatal(err)
	}
	return hook
}

func TestAntigravityHookRegistersTheNamedHookAndLeavesHerdrByteIdentical(t *testing.T) {
	o := antigravitySetup(t)
	path := agyHooksPath(o)
	put(t, path, foreignAntigravityHooks)
	exe := suiteExecutable(t)
	p := plan(t, "install", o)
	if len(p.HostHooks) != 1 || p.HostHooks[0].Host != "antigravity" || p.HostHooks[0].Path != path {
		t.Fatalf("host hook steps = %+v", p.HostHooks)
	}
	apply(t, p)

	hook := agyHiveHook(t, path)
	if len(hook) != 3 {
		t.Fatalf("hive hook events = %v", hook)
	}
	for _, event := range []string{"SessionStart", "PreInvocation", "Stop"} {
		hs := hook[event]
		want := agyHandler{Type: "command", Command: "'" + exe + "' hook antigravity " + event, Timeout: 5}
		if len(hs) != 1 || hs[0] != want {
			t.Fatalf("%s handlers = %+v, want %+v", event, hs, want)
		}
	}
	if _, ok := hook["PreToolUse"]; ok {
		t.Fatal("PreToolUse must not be registered")
	}
	got := get(t, path)
	if memberText(t, got, "herdr") != memberText(t, foreignAntigravityHooks, "herdr") {
		t.Fatalf("herdr changed:\n%s", got)
	}
	if !strings.HasPrefix(got, strings.TrimSuffix(foreignAntigravityHooks, "\n}\n")+",\n") {
		t.Fatalf("the bytes before Hive's key changed:\n%s", got)
	}
	if get(t, path+".hive-backup") != foreignAntigravityHooks || wfMode(t, path+".hive-backup") != 0600 {
		t.Fatal("the backup must hold the original file")
	}

	// Update with a moved binary rewrites only Hive's key.
	moved := copyExecutable(t, "hive-test-executable-moved")
	withHiveExecutable(t, moved)
	apply(t, plan(t, "install", o))
	got = get(t, path)
	if memberText(t, got, "herdr") != memberText(t, foreignAntigravityHooks, "herdr") || !strings.Contains(got, moved) || strings.Contains(got, "hive-test-executable'") {
		t.Fatalf("after the update:\n%s", got)
	}
	if n := len(agyHiveHook(t, path)["Stop"]); n != 1 {
		t.Fatalf("Stop handlers after the update = %d", n)
	}
	if again := apply(t, plan(t, "install", o)); again != "unchanged" {
		t.Fatalf("second update = %q", again)
	}

	// Remove takes the key away and gives the original bytes back.
	apply(t, plan(t, "remove", o))
	if got = get(t, path); got != foreignAntigravityHooks {
		t.Fatalf("after the removal:\n%s", got)
	}
	if get(t, path+".hive-backup") != foreignAntigravityHooks {
		t.Fatal("the backup must outlive the uninstall")
	}
	if s := stateFor(t, o); len(s.HostHooks) != 0 {
		t.Fatalf("host hook records after the removal: %+v", s.HostHooks)
	}
}

func TestAntigravityHookFileHiveCreatesIsRemovedWithIt(t *testing.T) {
	o := antigravitySetup(t)
	path := agyHooksPath(o)
	apply(t, plan(t, "install", o))
	if len(agyHiveHook(t, path)) != 3 {
		t.Fatalf("hooks.json:\n%s", get(t, path))
	}
	absent(t, path+".hive-backup")
	apply(t, plan(t, "remove", o))
	absent(t, path)
}

func TestAntigravityHookStatusRowsFollowTheFile(t *testing.T) {
	o := antigravitySetup(t)
	path := agyHooksPath(o)
	put(t, path, foreignAntigravityHooks)
	apply(t, plan(t, "install", o))
	rows := func() map[string]string {
		entries, err := Status(o)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, en := range entries {
			if en.Kind == "hook_entry" && en.Host == "antigravity" {
				m[en.Event] = en.Status
			}
		}
		return m
	}
	if m := rows(); len(m) != 3 || m["SessionStart"] != "installed" || m["PreInvocation"] != "installed" || m["Stop"] != "installed" {
		t.Fatalf("status rows = %v", m)
	}
	put(t, path, strings.Replace(get(t, path), `"Stop"`, `"StopX"`, 1))
	if m := rows(); m["Stop"] != "missing" || m["SessionStart"] != "installed" {
		t.Fatalf("status rows after a hand edit = %v", m)
	}
}

func TestAntigravityHookRefusesAFileItCannotReadAsANamedHooksObject(t *testing.T) {
	o := antigravitySetup(t)
	path := agyHooksPath(o)
	const odd = `{"hive": ["not", "an", "object"]}` + "\n"
	put(t, path, odd)
	p := plan(t, "install", o)
	if len(p.HostHooks) != 1 || p.HostHooks[0].Skipped == "" {
		t.Fatalf("the step must be skipped: %+v", p.HostHooks)
	}
	apply(t, p)
	if get(t, path) != odd {
		t.Fatal("a file Hive cannot edit safely must stay as it is")
	}
}

func TestAntigravityHookRecoversAnInterruptedInstall(t *testing.T) {
	for _, stage := range []string{"host-hooks-backup", "host-hooks-write", "host-hooks-after", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := antigravitySetup(t)
			path := agyHooksPath(o)
			put(t, path, foreignAntigravityHooks)
			if _, err := hookFail(stage).Apply(plan(t, "install", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if got := get(t, path); got != foreignAntigravityHooks {
				t.Fatalf("hooks.json after recovery:\n%s", got)
			}
			absent(t, path+".hive-backup")
			if len(stateFor(t, o).HostHooks) != 0 {
				t.Fatal("a host hooks record survived a recovery")
			}
		})
	}
}
