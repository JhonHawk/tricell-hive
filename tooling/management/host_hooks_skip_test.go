package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A shared hook file Hive cannot edit safely is skipped, never rewritten: the
// rest of the install finishes and the caller's notes name the file and why.

// Cursor's hooks.json is the subject: Claude Code's settings.json and Codex's
// hooks.json are also read by the legacy scan (tooling/legacy), which refuses
// an invalid or symlinked file before the plan reaches the hook step.
func skipCases(t *testing.T, hp hookPaths) map[string]func() {
	return map[string]func(){
		"invalid JSON":       func() { put(t, hp.cursor, `{"hooks": `) },
		"event not an array": func() { put(t, hp.cursor, `{"version":1,"hooks":{"stop":{}}}`) },
		"symlink": func() {
			elsewhere := filepath.Join(filepath.Dir(hp.cursor), "elsewhere.json")
			put(t, elsewhere, foreignCursorHooks)
			if err := os.Remove(hp.cursor); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(elsewhere, hp.cursor); err != nil {
				t.Fatal(err)
			}
		},
		"hard link": func() {
			other := filepath.Join(filepath.Dir(hp.cursor), "linked.json")
			if err := os.Link(hp.cursor, other); err != nil {
				t.Fatal(err)
			}
		},
	}
}

func TestInstallSkipsAnUneditableSharedFileAndFinishesTheRest(t *testing.T) {
	probe, _, _ := hooksSetup(t)
	cases := skipCases(t, pathsFor(probe))
	for name := range cases {
		t.Run(name, func(t *testing.T) {
			o, hp, originals := hooksSetup(t)
			skipCases(t, hp)[name]()
			before, _ := os.Lstat(hp.cursor)
			p, err := BuildPlan("install", o)
			if err != nil {
				t.Fatalf("the plan must not fail for a skipped file: %v", err)
			}
			notes, err := HostHookNotes(p)
			if err != nil {
				t.Fatal(err)
			}
			warned := false
			for _, n := range notes {
				if strings.Contains(n, hp.cursor) && strings.Contains(strings.ToLower(n), "skip") {
					warned = true
				}
			}
			if !warned {
				t.Fatalf("no warning names the skipped file and says why: %q", notes)
			}
			if got := HostHookPaths(p); !strings.Contains(strings.Join(got, "\n"), hp.cursor) {
				t.Fatalf("the preview must still list the target: %v", got)
			}
			apply(t, p)
			after, _ := os.Lstat(hp.cursor)
			if before.Mode() != after.Mode() || before.Size() != after.Size() {
				t.Fatal("the skipped file was rewritten")
			}
			if name == "invalid JSON" && get(t, hp.cursor) != `{"hooks": ` {
				t.Fatal("the invalid file was rewritten")
			}
			if len(hiveCommands(t, hp.codex, "codex")) != 4 || len(hiveCommands(t, hp.claude, "claude")) != 7 {
				t.Fatal("the rest of the install did not finish")
			}
			for _, path := range []string{hp.grokEvents, hp.opencode, hp.pi} {
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("%s not installed: %v", path, err)
				}
			}
			if _, ok := stateFor(t, o).HostHooks[hp.cursor]; ok {
				t.Fatal("a skipped file got a record")
			}
			if unchanged, err := PlanUnchanged(plan(t, "install", o)); err != nil || !unchanged {
				t.Fatalf("a second plan must be unchanged: %v %v", unchanged, err)
			}
			apply(t, plan(t, "remove", o))
			if name == "invalid JSON" && get(t, hp.cursor) != `{"hooks": ` {
				t.Fatal("uninstall touched the skipped file")
			}
			for _, path := range []string{hp.codex, hp.claude} {
				if get(t, path) != originals[path] {
					t.Fatalf("%s not restored", path)
				}
			}
		})
	}
}

// Dotfile managers link settings.json (and Codex's hooks.json) into place: the
// link is skipped with a warning, and the install of every other host finishes.
func TestSymlinkedClaudeAndCodexFilesAreSkippedNotFatal(t *testing.T) {
	for _, which := range []string{"claude", "codex"} {
		t.Run(which, func(t *testing.T) {
			o, hp, originals := hooksSetup(t)
			path := map[string]string{"claude": hp.claude, "codex": hp.codex}[which]
			elsewhere := filepath.Join(o.Home, "dotfiles", filepath.Base(path))
			put(t, elsewhere, originals[path])
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(elsewhere, path); err != nil {
				t.Fatal(err)
			}
			p, err := BuildPlan("install", o)
			if err != nil {
				t.Fatalf("a linked %s file must not fail the plan: %v", which, err)
			}
			notes, err := HostHookNotes(p)
			if err != nil || !strings.Contains(strings.Join(notes, "\n"), path) {
				t.Fatalf("no warning names %s: %q %v", path, notes, err)
			}
			apply(t, p)
			if info, err := os.Lstat(path); err != nil || info.Mode()&os.ModeSymlink == 0 || get(t, elsewhere) != originals[path] {
				t.Fatal("the linked file was touched")
			}
			if len(hiveCommands(t, hp.cursor, "cursor")) != 4 {
				t.Fatal("the other hosts were not registered")
			}
			apply(t, plan(t, "remove", o))
			if get(t, elsewhere) != originals[path] {
				t.Fatal("uninstall touched the linked file")
			}
		})
	}
}

func TestRemoveOfAFileThatBecameUneditableWarnsAndLeavesIt(t *testing.T) {
	o, hp, _ := hooksSetup(t)
	apply(t, plan(t, "install", o))
	elsewhere := filepath.Join(o.Home, "elsewhere.json")
	put(t, elsewhere, get(t, hp.cursor))
	if err := os.Remove(hp.cursor); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, hp.cursor); err != nil {
		t.Fatal(err)
	}
	result := apply(t, plan(t, "remove", o))
	if !strings.Contains(result, hp.cursor) {
		t.Fatalf("remove must warn about the skipped file: %q", result)
	}
	if info, err := os.Lstat(hp.cursor); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("the symlink was replaced")
	}
	if len(hiveCommands(t, hp.codex, "codex")) != 0 || len(hiveCommands(t, hp.claude, "claude")) != 0 {
		t.Fatal("the other hosts must still be removed")
	}
}

func TestRecoverRestoresAnEntryAnInterruptedInstallReplacedInPlace(t *testing.T) {
	for _, stage := range []string{"host-hooks-write", "host-hooks-after", "state"} {
		t.Run(stage, func(t *testing.T) {
			o, hp, originals := hooksSetup(t)
			apply(t, plan(t, "install", o))
			installed := map[string]string{}
			for _, path := range []string{hp.claude, hp.codex, hp.cursor} {
				installed[path] = get(t, path)
			}
			moved := copyExecutable(t, "hive-test-executable-moved")
			withHiveExecutable(t, moved)
			if _, err := hookFail(stage).Apply(plan(t, "install", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			for path, want := range installed {
				if got := get(t, path); got != want {
					t.Fatalf("%s after recovery lost the previous entries:\n%s", path, got)
				}
			}
			apply(t, plan(t, "remove", o))
			assertForeignIntact(t, originals)
		})
	}
}

func TestStatusReportsEachHookEntryPerHost(t *testing.T) {
	o, hp, _ := hooksSetup(t)
	rows := func() map[string]StatusEntry {
		so := o
		entries, err := Status(so)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]StatusEntry{}
		for _, e := range entries {
			if e.Kind == "hook_entry" {
				out[e.Host+"/"+e.Event] = e
			}
		}
		return out
	}
	if n := len(rows()); n != 0 {
		t.Fatalf("a fresh home reports %d hook rows", n)
	}
	apply(t, plan(t, "install", o))
	r := rows()
	for _, k := range []string{"claude/SessionStart", "claude/SessionEnd", "codex/SessionEnd", "cursor/stop"} {
		if r[k].Status != "installed" || r[k].Path == "" {
			t.Fatalf("%s: %+v (rows: %d)", k, r[k], len(r))
		}
	}
	put(t, hp.claude, strings.Replace(get(t, hp.claude), "SessionEnd", "SessionEndX", 1))
	if r := rows()["claude/SessionEnd"]; r.Status != "missing" {
		t.Fatalf("after a hand edit: %+v", r)
	}
	put(t, hp.claude, `{"hooks": `)
	if r := rows()["claude/SessionStart"]; r.Status != "skipped" || r.Reason == "" {
		t.Fatalf("invalid file: %+v", r)
	}
}
