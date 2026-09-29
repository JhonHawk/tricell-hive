package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of the `hive doctor` and `hive models` commands (#46, T5).

// installCodexFixture installs the codex host over a characterization fixture
// and returns it.
func installCodexFixture(t *testing.T) characterizationFixture {
	t.Helper()
	f := newCharacterizationFixture(t)
	o := management.Options{Scope: "user", Home: f.home, StateDir: f.stateDir, Source: f.source, Hosts: []string{"codex"}}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestDoctorCommandPrintsFindingsAndExitsZero(t *testing.T) {
	f := installCodexFixture(t)
	block := filepath.Join(f.home, ".codex", "AGENTS.md")
	data, err := os.ReadFile(block)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(block, []byte(strings.Replace(string(data), "Keep user content.", "Edited by hand.", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	var runErr error
	out := captureStdout(t, func() {
		runErr = run([]string{"doctor", "--home", f.home, "--state-dir", f.stateDir, "--project", t.TempDir()})
	})
	if runErr != nil {
		t.Fatalf("doctor failed on a finding: %v\n%s", runErr, out)
	}
	for _, want := range []string{"CLIs\n", "Installation\n", "Sessions\n", "drift  codex", block} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// The command prints exactly what the shared renderer produces.
	o := management.Options{Scope: "user", Home: f.home, StateDir: f.stateDir}
	var want strings.Builder
	project := t.TempDir()
	renderDoctorText(collectDoctor(o, project, realDoctorDeps()), &want)
	got := captureStdout(t, func() {
		runErr = run([]string{"doctor", "--home", f.home, "--state-dir", f.stateDir, "--project", project})
	})
	if runErr != nil || got != want.String() {
		t.Errorf("doctor output differs from renderDoctorText (err %v):\n%s\n---\n%s", runErr, got, want.String())
	}
}

func TestDoctorCommandFailsOnAStateDirThatIsAFile(t *testing.T) {
	f := newCharacterizationFixture(t)
	file := filepath.Join(t.TempDir(), "state")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	var runErr error
	captureStdout(t, func() { runErr = run([]string{"doctor", "--home", f.home, "--state-dir", file}) })
	if runErr == nil {
		t.Fatal("doctor accepted a state directory that is a file")
	}
	captureStdout(t, func() { runErr = run([]string{"models", "--home", f.home, "--state-dir", file}) })
	if runErr == nil {
		t.Fatal("models accepted a state directory that is a file")
	}
}

func TestModelsCommandPrintsTheRowsAndExitsZero(t *testing.T) {
	f := installCodexFixture(t)
	var runErr error
	out := captureStdout(t, func() { runErr = run([]string{"models", "--home", f.home, "--state-dir", f.stateDir}) })
	if runErr != nil {
		t.Fatalf("models failed: %v", runErr)
	}
	o := management.Options{Scope: "user", Home: f.home, StateDir: f.stateDir}
	hosts, rows, err := collectModels(o)
	if err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	renderModelsText(hosts, rows, &want)
	if out != want.String() || !strings.HasPrefix(out, "codex\n") {
		t.Errorf("models output:\n%s\nwant:\n%s", out, want.String())
	}

	empty := newCharacterizationFixture(t)
	out = captureStdout(t, func() { runErr = run([]string{"models", "--home", empty.home, "--state-dir", empty.stateDir}) })
	if runErr != nil || out != "No CLI hosts are registered\n" {
		t.Errorf("models without hosts = %q, %v", out, runErr)
	}
}

func TestDoctorAndModelsRejectPositionalArguments(t *testing.T) {
	for _, cmd := range []string{"doctor", "models"} {
		var runErr error
		captureStdout(t, func() { runErr = run([]string{cmd, "extra"}) })
		if runErr == nil {
			t.Errorf("%s accepted a positional argument", cmd)
		}
	}
}
