package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of how the CLI reports the hook registrations of the six hosts, on a
// synthetic home. They never touch the real home.

var allHookHosts = []string{"claude", "codex", "cursor", "grok", "opencode", "pi"}

func hookTargets(home string) []string {
	return []string{
		filepath.Join(home, ".claude", "settings.json"),
		filepath.Join(home, ".codex", "hooks.json"),
		filepath.Join(home, ".cursor", "hooks.json"),
		filepath.Join(home, ".grok", "hooks", "hive-session-events.json"),
		filepath.Join(home, ".config", "opencode", "plugins", "hive-events.js"),
		filepath.Join(home, ".pi", "agent", "extensions", "hive-events.ts"),
	}
}

func TestInstallSummaryListsTheSixHookTargetsInADryRun(t *testing.T) {
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	o := management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir, Source: env.repo, Hosts: allHookHosts}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	showInstallSummary(&out, p, onboardingPreview{}, true, false, true)
	for _, path := range hookTargets(env.home) {
		if !strings.Contains(out.String(), path) {
			t.Errorf("the dry-run summary does not list %s:\n%s", path, out.String())
		}
	}
}

func TestUpdateReportsTheHookNotesAfterApplyAndKeepsTheTransactionID(t *testing.T) {
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	gitInitLocal(t, env.repo)
	gitCommitAll(t, env.repo, "init")
	installDirect(t, env.home, env.stateDir, env.repo, []string{"claude", "codex", "cursor"})
	// A hand edit drops one Hive entry in each of the two files with a note.
	for _, name := range []string{".codex", ".cursor"} {
		path := filepath.Join(env.home, name, "hooks.json")
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		edited := bytes.Replace(b, []byte("SessionStart"), []byte("SessionStartX"), 1)
		edited = bytes.Replace(edited, []byte("sessionStart"), []byte("sessionStartX"), 1)
		if bytes.Equal(edited, b) {
			t.Fatalf("setup: nothing edited in %s", path)
		}
		if err := os.WriteFile(path, edited, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{"approve", "/hooks", "Cursor CLI"} {
		if !strings.Contains(text, want) {
			t.Errorf("the report lacks %q:\n%s", want, text)
		}
	}
	if !regexp.MustCompile(`Hive updated \([0-9a-f]+\)\.\n`).MatchString(text) {
		t.Errorf("the transaction line must stay a bare ID:\n%s", text)
	}
}

func TestDoctorAndStatusShowTheHookEntriesPerHost(t *testing.T) {
	o, home, _ := doctorHome(t, "claude,codex,cursor")
	rows := func() []management.StatusEntry {
		so := o
		so.Hosts = []string{"claude", "codex", "cursor"}
		entries, err := management.Status(so)
		if err != nil {
			t.Fatal(err)
		}
		var out []management.StatusEntry
		for _, e := range entries {
			if e.Kind == "hook_entry" {
				out = append(out, e)
			}
		}
		return out
	}
	if n := len(rows()); n != 7+4+4 {
		t.Fatalf("hook entry rows = %d", n)
	}
	text := sectionText(installationSection(o, loadDoctorState(o)))
	if strings.Contains(text, "hooks") && !strings.Contains(text, "No problems found") {
		t.Fatalf("a healthy install reports a hook problem:\n%s", text)
	}
	codex := filepath.Join(home, ".codex", "hooks.json")
	b, err := os.ReadFile(codex)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codex, bytes.Replace(b, []byte("SubagentStart"), []byte("SubagentStartX"), 1), 0o600); err != nil {
		t.Fatal(err)
	}
	cursor := filepath.Join(home, ".cursor", "hooks.json")
	if err := os.WriteFile(cursor, []byte(`{"hooks": `), 0o600); err != nil {
		t.Fatal(err)
	}
	text = sectionText(installationSection(o, loadDoctorState(o)))
	for _, want := range []string{"codex", "missing", "SubagentStart", cursor, "skipped"} {
		if !strings.Contains(text, want) {
			t.Errorf("the Installation section lacks %q:\n%s", want, text)
		}
	}
}
