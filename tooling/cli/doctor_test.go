package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"tricell-hive/tooling/management"
)

// doctorFake replaces every doctorDeps field, so no test reads the
// developer's configuration or runs a real CLI.
type doctorFake struct {
	mu       sync.Mutex
	paths    map[string]string // binary name -> path lookPath returns
	versions map[string]string // path -> output of --version
	verErr   map[string]error
	alive    map[int]bool
	env      map[string]string
	home     string
	now      time.Time
	ran      []string
}

func newDoctorFake(home string) *doctorFake {
	return &doctorFake{
		paths: map[string]string{}, versions: map[string]string{}, verErr: map[string]error{},
		alive: map[int]bool{}, env: map[string]string{}, home: home,
		now: time.Date(2026, 9, 29, 15, 0, 0, 0, time.UTC),
	}
}

// install makes a binary detectable and gives it a --version output.
func (f *doctorFake) install(name, output string) {
	path := "/fake/bin/" + name
	f.paths[name] = path
	f.versions[path] = output
}

func (f *doctorFake) deps() doctorDeps {
	return doctorDeps{
		lookPath: func(name string) (string, error) {
			if p, ok := f.paths[name]; ok {
				return p, nil
			}
			return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
		},
		runVersion: func(bin string) (string, error) {
			f.mu.Lock()
			f.ran = append(f.ran, bin)
			f.mu.Unlock()
			if err := f.verErr[bin]; err != nil {
				return "", err
			}
			return f.versions[bin], nil
		},
		processAlive: func(pid int) bool { return f.alive[pid] },
		now:          func() time.Time { return f.now },
		userHome:     func() (string, error) { return f.home, nil },
		getenv:       func(k string) string { return f.env[k] },
	}
}

// isolateDoctorEnv points the process at home and clears every variable the
// manager reads, so a non-synthetic run touches only the test folder.
func isolateDoctorEnv(t *testing.T, home string) {
	t.Helper()
	t.Setenv("HOME", home)
	for _, k := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "GROK_HOME", "PI_CODING_AGENT_DIR", "XDG_CONFIG_HOME", "XDG_STATE_HOME"} {
		t.Setenv(k, "")
	}
}

// doctorHome installs the hosts into a fresh synthetic home and returns the
// synthetic options. Use realHomeOptions for a run that detects CLIs.
func doctorHome(t *testing.T, hosts string) (o management.Options, home, stateDir string) {
	t.Helper()
	source := minimalTestSource(t)
	home, stateDir = newHostsTestHome(t)
	installViaText(t, home, stateDir, source, hosts, "y\n", hostsTestDeps(coreOnlyAdapterFactory))
	o = management.Options{Scope: "user", Home: home, StateDir: stateDir}
	_, norm, err := management.NormalizeOptions(o)
	if err != nil {
		t.Fatal(err)
	}
	o.StateDir = norm
	return o, home, norm
}

// nonSyntheticOptions returns options that address the same home as "the real
// home" (HOME points at it), so CLIs are detected through the fake lookPath.
func nonSyntheticOptions(t *testing.T, o management.Options, home string) management.Options {
	t.Helper()
	isolateDoctorEnv(t, home)
	o.Home = ""
	return o
}

// setReleaseTime sets the write time of every retained release.
func setReleaseTime(t *testing.T, stateDir string, at time.Time) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(stateDir, "releases", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no release files: %v", err)
	}
	for _, f := range files {
		if err := os.Chtimes(f, at, at); err != nil {
			t.Fatal(err)
		}
	}
}

func sectionText(s doctorSection) string { return doctorSectionsText(s) }

func mustContain(t *testing.T, text string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(text, w) {
			t.Fatalf("missing %q in:\n%s", w, text)
		}
	}
}

func mustNotContain(t *testing.T, text string, unwanted ...string) {
	t.Helper()
	for _, w := range unwanted {
		if strings.Contains(text, w) {
			t.Fatalf("unexpected %q in:\n%s", w, text)
		}
	}
}

func releaseIDOf(t *testing.T, stateDir string) string {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(stateDir, "releases", "*.json"))
	for _, f := range files {
		base := filepath.Base(f)
		if !strings.Contains(base, ".commits") {
			return strings.TrimSuffix(base, ".json")
		}
	}
	t.Fatal("no release")
	return ""
}

// ---------------------------------------------------------------------------
// sanitizeLine.
// ---------------------------------------------------------------------------

func TestSanitizeLine(t *testing.T) {
	for _, tc := range []struct{ name, in, want string }{
		{"plain", "claude 2.1.284", "claude 2.1.284"},
		{"clear screen", "v1\x1b[2J.0", "v1.0"},
		{"color", "\x1b[31mred\x1b[0m", "red"},
		{"osc title", "a\x1b]0;evil\x07b", "ab"},
		{"osc st", "a\x1b]8;;http://x\x1b\\link\x1b]8;;\x1b\\b", "alinkb"},
		{"control", "a\x00b\x07c\x7fd", "abcd"},
		{"c1 control", "a\u009bb", "ab"},
		{"newline and tab", "a\nb\tc\r", "a b c "},
		{"bidi override", "a‮b", "ab"},
		{"invalid utf8", "a\xffb", "ab"},
		{"lone escape", "a\x1b", "a"},
		{"unicode kept", "café — ✓", "café — ✓"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitizeLine(tc.in); got != tc.want {
				t.Fatalf("sanitizeLine(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CLIs (AC2).
// ---------------------------------------------------------------------------

func TestDoctorCLIsDetectedHostShowsVersionReleaseAndState(t *testing.T) {
	o, home, stateDir := doctorHome(t, "claude,codex")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("claude", "2.1.284 (Claude Code)\nsecond line\n")
	f.install("codex", "codex-cli 0.158.0\n")
	f.install("pi", "0.87.1\n")

	r := collectDoctor(o, t.TempDir(), f.deps())
	text := sectionText(r.CLIs)
	short := shortHash(releaseIDOf(t, stateDir))
	mustContain(t, text, "CLIs")
	lines := map[string]string{}
	for _, l := range r.CLIs.Lines {
		lines[strings.Fields(l)[0]] = l
	}
	for _, host := range installerHosts {
		if lines[host] == "" {
			t.Fatalf("no row for %s in:\n%s", host, text)
		}
	}
	// Rows follow installerHosts order.
	for i, host := range installerHosts {
		if !strings.HasPrefix(r.CLIs.Lines[i], host) {
			t.Fatalf("row %d = %q, want host %s", i, r.CLIs.Lines[i], host)
		}
	}
	mustContain(t, lines["claude"], "detected", "CLI version 2.1.284 (Claude Code)", "Hive release "+short, "(legacy)")
	mustNotContain(t, lines["claude"], "second line")
	mustContain(t, lines["codex"], "detected", "CLI version codex-cli 0.158.0", "Hive release "+short)
	mustContain(t, lines["pi"], "detected", "CLI version 0.87.1", "not installed by Hive")
	mustContain(t, lines["grok"], "not detected", "not installed by Hive")
	mustNotContain(t, lines["grok"], "version")
}

// TestDoctorCLIsLabelsTheHostVersionAndHiveRelease covers M6: the host
// binary's version and Hive's release are named apart in every row, so neither
// is read as the CLIs view's Version column (Hive's own version).
func TestDoctorCLIsLabelsTheHostVersionAndHiveRelease(t *testing.T) {
	o, home, stateDir := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("claude", "2.1.284 (Claude Code)\n")
	f.install("codex", "codex-cli 0.158.0\n")
	f.verErr["/fake/bin/codex"] = errors.New("exit status 2")
	short := shortHash(releaseIDOf(t, stateDir))
	lines := collectDoctor(o, t.TempDir(), f.deps()).CLIs.Lines
	byHost := map[string]string{}
	for _, l := range lines {
		byHost[strings.Fields(l)[0]] = l
	}
	mustContain(t, byHost["claude"], "Hive release "+short, "CLI version 2.1.284 (Claude Code)")
	mustContain(t, byHost["codex"], "CLI version unavailable: exit status 2")
	for host, l := range byHost {
		// "version" always follows "CLI " and "release" always follows "Hive ".
		if strings.Count(l, "version") != strings.Count(l, "CLI version") || strings.Count(l, "release") != strings.Count(l, "Hive release") {
			t.Errorf("%s: an unlabeled version or release in %q", host, l)
		}
	}

	// Without a readable state, the release is still named as Hive's.
	bad := o
	bad.StateDir = filepath.Join(t.TempDir(), "state-is-a-file")
	if err := os.WriteFile(bad.StateDir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, l := range collectDoctor(bad, t.TempDir(), f.deps()).CLIs.Lines {
		mustContain(t, l, "Hive release unknown")
	}
}

func TestDoctorCLIsVersionIsTrimmedAndSanitized(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("claude", "v1\x1b[2J.0 "+strings.Repeat("x", 200)+"\n")
	r := collectDoctor(o, t.TempDir(), f.deps())
	text := sectionText(r.CLIs)
	mustNotContain(t, text, "\x1b", "[2J")
	mustContain(t, text, "CLI version v1.0 ")
	for _, l := range r.CLIs.Lines {
		if strings.HasPrefix(l, "claude") {
			v := l[strings.Index(l, "CLI version ")+len("CLI version "):]
			if n := len([]rune(v)); n != versionMaxRunes {
				t.Fatalf("version has %d runes, want %d: %q", n, versionMaxRunes, v)
			}
		}
	}
}

func TestDoctorCLIsCursorRunsCursorAgent(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("cursor", "should not run\n")
	f.install("cursor-agent", "2026.09.23\n")
	r := collectDoctor(o, t.TempDir(), f.deps())
	var cursor string
	for _, l := range r.CLIs.Lines {
		if strings.HasPrefix(l, "cursor") {
			cursor = l
		}
	}
	mustContain(t, cursor, "detected", "CLI version 2026.09.23")
	if len(f.ran) != 1 || f.ran[0] != "/fake/bin/cursor-agent" {
		t.Fatalf("ran %v, want only cursor-agent", f.ran)
	}
}

func TestDoctorCLIsCursorWithoutCursorAgent(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("cursor", "x\n")
	r := collectDoctor(o, t.TempDir(), f.deps())
	mustContain(t, sectionText(r.CLIs), "CLI version unavailable: cursor-agent not found")
	if len(f.ran) != 0 {
		t.Fatalf("ran %v", f.ran)
	}
}

func TestDoctorCLIsVersionFailureShowsReason(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("claude", "")
	f.verErr["/fake/bin/claude"] = errors.New("exit status 2\x1b[2J")
	f.install("codex", "  \n")
	r := collectDoctor(o, t.TempDir(), f.deps())
	text := sectionText(r.CLIs)
	mustContain(t, text, "CLI version unavailable: exit status 2", "CLI version unavailable: no output")
	mustNotContain(t, text, "\x1b")
}

// fakeExecutable writes an executable shell script in dir and returns its path.
func fakeExecutable(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDoctorCLIsRealVersionTimesOutWithoutWaiting(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	bin := t.TempDir()
	f := newDoctorFake(home)
	f.paths["claude"] = fakeExecutable(t, bin, "claude", "sleep 5")
	f.paths["codex"] = fakeExecutable(t, bin, "codex", `echo "codex 1.0"`)
	deps := f.deps()
	deps.runVersion = runVersionCommand // the real function, over fake executables

	start := time.Now()
	r := collectDoctor(o, t.TempDir(), deps)
	if elapsed := time.Since(start); elapsed >= 5*time.Second {
		t.Fatalf("the section took %s, want under 5s", elapsed)
	}
	text := sectionText(r.CLIs)
	mustContain(t, text, "CLI version unavailable: timed out after 3s", "CLI version codex 1.0")
}

func TestDoctorUndetectedHostIsNotExecuted(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	bin, marker := t.TempDir(), filepath.Join(t.TempDir(), "ran")
	f := newDoctorFake(home)
	// cursor is not detected (no "cursor" on the path) although its version
	// binary exists: it must not run.
	f.paths["cursor-agent"] = fakeExecutable(t, bin, "cursor-agent", "touch "+marker)
	deps := f.deps()
	deps.runVersion = runVersionCommand
	r := collectDoctor(o, t.TempDir(), deps)
	mustContain(t, sectionText(r.CLIs), "not detected")
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("an undetected CLI was executed")
	}
}

func TestDoctorSyntheticHomeRunsNothing(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	bin, marker := t.TempDir(), filepath.Join(t.TempDir(), "ran")
	f := newDoctorFake(home)
	for _, h := range installerHosts {
		f.paths[h] = fakeExecutable(t, bin, h, "touch "+marker)
	}
	f.paths["cursor-agent"] = f.paths["cursor"]
	f.env["CLAUDE_CONFIG_DIR"] = "/must/not/be/read"
	deps := f.deps()
	deps.runVersion = runVersionCommand
	r := collectDoctor(o, t.TempDir(), deps) // o.Home is synthetic
	text := sectionText(r.CLIs)
	if strings.Contains(text, "detected") && !strings.Contains(text, "not detected") {
		t.Fatalf("a CLI was detected under a synthetic home:\n%s", text)
	}
	mustNotContain(t, text, "CLI version")
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("a CLI was executed under a synthetic home")
	}
	if len(f.ran) != 0 {
		t.Fatalf("runVersion was called: %v", f.ran)
	}
}

func TestDoctorNoRegisteredHosts(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	// The empty state says what to do first, in the sections and in the text
	// command that prints them.
	mustContain(t, sectionText(r.Installation), "No CLI hosts are registered. Open CLIs from the menu, or run hive install, to install Hive.")
	mustContain(t, sectionText(r.Sessions), "No CLI hosts are registered. Open CLIs from the menu, or run hive install, to install Hive.")
	for _, l := range r.CLIs.Lines {
		mustContain(t, l, "not detected", "not installed by Hive")
	}
}

// TestDoctorPendingOperationShowsWhenNoHostIsRegistered covers D1 (AC3): an
// interrupted first install leaves a pending operation and no registered CLI,
// and Installation must still point to hive recover.
func TestDoctorPendingOperationShowsWhenNoHostIsRegistered(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	if err := os.WriteFile(filepath.Join(stateDir, "pending.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	text := sectionText(r.Installation)
	mustContain(t, text, "No CLI hosts are registered", "An unfinished Hive operation is pending; run hive recover --state-dir", "to finish it")
	mustNotContain(t, text, "No problems found")
	if len(r.Installation.Lines) != 2 {
		t.Fatalf("lines = %q", r.Installation.Lines)
	}
}

// TestDoctorUnreadableStateLeadsWithWordsThenDetail covers M2: a damaged state
// file is explained in plain words with the way out, and the technical error
// follows on its own line. The explanation is written once, in the first
// section (N1); the other two say only that they were not checked.
func TestDoctorUnreadableStateLeadsWithWordsThenDetail(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	_, stateDir, err := management.NormalizeOptions(o) // the resolved path is the one shown
	if err != nil {
		t.Fatal(err)
	}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	lines := strings.Split(sectionText(r.CLIs), "\n")
	if len(lines) < 4 {
		t.Fatalf("CLIs: too few lines: %q", lines)
	}
	first, detail := lines[1], lines[2]
	mustContain(t, first, "Could not check everything: Hive's state in "+stateDir+" could not be read", "hive status")
	mustNotContain(t, first, "invalid character")
	mustContain(t, detail, "Detail: ", "invalid character")
}

// TestDoctorUnreadableStateIsExplainedOnce covers N1: the explanation and its
// Detail line appear once across CLIs, Installation and Sessions; the other two
// sections carry one short line and keep Err, so the view still offers a retry.
func TestDoctorUnreadableStateIsExplainedOnce(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	text := doctorSectionsText(r.CLIs, r.Installation, r.Sessions)
	if n := strings.Count(text, "Detail: "); n != 1 {
		t.Fatalf("Detail line appears %d times, want 1:\n%s", n, text)
	}
	if n := strings.Count(text, "could not be read"); n != 3 {
		// once in the full explanation, once in each short line
		t.Fatalf("\"could not be read\" appears %d times, want 3:\n%s", n, text)
	}
	const short = "Not checked: Hive's state could not be read (see above)."
	if n := strings.Count(text, short); n != 2 {
		t.Fatalf("short line appears %d times, want 2:\n%s", n, text)
	}
	for _, sec := range []doctorSection{r.CLIs, r.Installation, r.Sessions} {
		if sec.Err == "" {
			t.Fatalf("%s lost its Err", sec.Title)
		}
	}
	for _, sec := range []doctorSection{r.Installation, r.Sessions} {
		mustContain(t, sectionText(sec), short)
		mustNotContain(t, sectionText(sec), "Detail: ", "Could not check everything")
	}
}

func TestDoctorSectionErrorStaysInsideItsSection(t *testing.T) {
	home, _ := newHostsTestHome(t)
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A state directory that is a file cannot be read; the report still has
	// every section, and the state-dependent ones carry the error.
	o := management.Options{Scope: "user", Home: home, StateDir: blocker}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	if r.CLIs.Title == "" || r.Installation.Title == "" || r.Sessions.Title == "" {
		t.Fatalf("missing sections: %+v", r)
	}
	if r.Installation.Err == "" || r.Sessions.Err == "" || r.CLIs.Err == "" {
		t.Fatalf("errors were not kept in the sections: %+v", r)
	}
	if len(r.CLIs.Lines) != len(installerHosts) {
		t.Fatalf("CLIs rows = %d", len(r.CLIs.Lines))
	}
	mustContain(t, sectionText(r.Installation), "Not checked")
	mustContain(t, sectionText(r.CLIs), "Could not check everything")
}

// ---------------------------------------------------------------------------
// Installation (AC3).
// ---------------------------------------------------------------------------

func TestDoctorInstallationCleanSaysNoProblems(t *testing.T) {
	o, home, _ := doctorHome(t, "claude,codex")
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	if got := strings.Join(r.Installation.Lines, "|"); got != "No problems found" {
		t.Fatalf("installation = %q", got)
	}
}

func TestDoctorInstallationListsDriftAndDuplicatedMarkersWithPath(t *testing.T) {
	o, home, _ := doctorHome(t, "claude,codex")
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	codexMD := filepath.Join(home, ".codex", "AGENTS.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudeMD, bytes.Replace(data, []byte("Minimal test guidance."), []byte("Edited by hand."), 1), 0o600); err != nil {
		t.Fatal(err)
	}
	codexData, err := os.ReadFile(codexMD)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codexMD, append(append([]byte(nil), codexData...), codexData...), 0o600); err != nil {
		t.Fatal(err)
	}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	text := sectionText(r.Installation)
	canonHome, _ := filepath.EvalSymlinks(home)
	mustContain(t, text,
		"drift  claude  The installed file was changed or cannot be read",
		filepath.Join(canonHome, ".claude", "CLAUDE.md"),
		"drift  codex",
		filepath.Join(canonHome, ".codex", "AGENTS.md"))
	mustNotContain(t, text, "No problems found", "not_installed")
}

// driftRepairLine is what Installation says once a file is in drift (M4), spelled
// out here so a change to the wording is deliberate. No Hive command repairs
// drift (see TestNoCommandRepairsADriftedManagedFile), so the line says what the
// person can do.
const driftRepairLine = "Hive cannot repair a changed file by itself: hive install, hive update and hive plan remove refuse to run while it differs from what Hive wrote. " +
	"Undo the change (or fix its permissions), or restore the file from a backup, then run hive status to check."

func TestDoctorInstallationDriftSaysHowToRepairIt(t *testing.T) {
	o, home, _ := doctorHome(t, "claude,codex")
	for _, path := range []string{filepath.Join(home, ".claude", "CLAUDE.md"), filepath.Join(home, ".codex", "AGENTS.md")} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, bytes.Replace(data, []byte("Minimal test guidance."), []byte("Edited by hand."), 1), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sec := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps()).Installation
	// Two files drifted, one line: after the rows it explains.
	var at []int
	for i, l := range sec.Lines {
		if l == driftRepairLine {
			at = append(at, i)
		}
	}
	if len(at) != 1 || at[0] != len(sec.Lines)-1 {
		t.Fatalf("repair line at %v in:\n%s", at, sectionText(sec))
	}
	mustContain(t, sectionText(sec), "drift  claude", "drift  codex")
}

func TestDoctorInstallationRepairLineOnlyForDrift(t *testing.T) {
	// Other findings, and a clean installation, do not get it.
	st := doctorState{registered: []string{"claude"}, entries: []management.StatusEntry{
		{Path: "/p/unowned", Host: "claude", Status: "unowned"},
		{Path: "/p/broken", Host: "claude", Status: "unowned_or_conflicting"},
		{Path: "/p/shadowed", Host: "claude", Status: "shadowed"},
		{Path: "/p/migrate", Host: "claude", Status: "migration_required"},
	}}
	mustNotContain(t, sectionText(installationSection(management.Options{}, st)), "Hive cannot repair", "hive plan remove")
	st.entries = []management.StatusEntry{{Path: "/p/ok", Host: "claude", Status: "installed"}}
	mustNotContain(t, sectionText(installationSection(management.Options{}, st)), "Hive cannot repair")

	// A voice row in drift counts too, and its row is listed as shared.
	st.entries = []management.StatusEntry{{Path: "/p/voice", Kind: "voice", Status: "drift"}}
	text := sectionText(installationSection(management.Options{}, st))
	mustContain(t, text, "drift  shared", driftRepairLine)
}

func TestDoctorInstallationHidesNotInstalledRows(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	entries, err := management.Status(management.Options{Scope: "user", Home: home, StateDir: o.StateDir, Hosts: []string{"claude"}})
	if err != nil {
		t.Fatal(err)
	}
	seen := false
	for _, e := range entries {
		seen = seen || e.Status == "not_installed"
	}
	if !seen {
		t.Fatal("setup: Status has no not_installed row to hide")
	}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	mustNotContain(t, sectionText(r.Installation), "not_installed", "not installed")
}

func TestDoctorInstallationPhrasePerState(t *testing.T) {
	st := doctorState{registered: []string{"claude"}, entries: []management.StatusEntry{
		{Path: "/p/installed", Host: "claude", Status: "installed"},
		{Path: "/p/shared", Host: "claude", Status: "retained_shared"},
		{Path: "/p/none", Host: "claude", Status: "not_installed"},
		{Path: "/p/drift", Host: "claude", Status: "drift"},
		{Path: "/p/shadowed", Host: "claude", Status: "shadowed"},
		{Path: "/p/unowned", Host: "claude", Status: "unowned"},
		{Path: "/p/broken", Host: "claude", Status: "unowned_or_conflicting"},
		{Path: "/p/migrate", Host: "claude", Status: "migration_required"},
		{Path: "/p/recover", Host: "claude", Status: "recovery_required"},
		{Path: "/p/new\x1b[2J", Host: "claude", Status: "brand_new_state"},
	}}
	text := sectionText(installationSection(management.Options{}, st))
	mustContain(t, text,
		"unowned_or_conflicting  claude  Hive markers are missing, duplicated or broken", "/p/broken",
		"drift  claude  The installed file was changed or cannot be read", "/p/drift",
		"shadowed  claude", "/p/shadowed",
		"unowned  claude  A file exists that Hive did not write", "/p/unowned",
		"migration_required  claude", "/p/migrate",
		"recovery_required  claude  An unfinished operation must be recovered", "/p/recover",
		"brand_new_state  claude  Unknown state brand_new_state", "/p/new")
	mustNotContain(t, text, "/p/installed", "/p/shared", "/p/none", "\x1b", "No problems found")
}

func TestDoctorInstallationPendingOperationIsOneRecoverLine(t *testing.T) {
	o, home, stateDir := doctorHome(t, "claude,codex")
	if err := os.WriteFile(filepath.Join(stateDir, "pending.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := collectDoctor(o, t.TempDir(), newDoctorFake(home).deps())
	if len(r.Installation.Lines) != 1 {
		t.Fatalf("lines = %q", r.Installation.Lines)
	}
	line := r.Installation.Lines[0]
	mustContain(t, line, "hive recover", "--state-dir")
	mustNotContain(t, line, "recovery_required")
	mustNotContain(t, sectionText(r.Installation), "No problems found")
}

func TestDoctorRecoverCommandOmitsDefaultStateDir(t *testing.T) {
	home := t.TempDir()
	def, err := management.DefaultStateDir(home, true)
	if err != nil {
		t.Fatal(err)
	}
	o := management.Options{Home: home}
	if got := recoverCommand(o, doctorState{stateDir: def}); got != "run hive recover" {
		t.Fatalf("default: %q", got)
	}
	got := recoverCommand(o, doctorState{stateDir: "/some dir/state"})
	if got != "run hive recover --state-dir '/some dir/state'" {
		t.Fatalf("explicit: %q", got)
	}
}

// ---------------------------------------------------------------------------
// General.
// ---------------------------------------------------------------------------

// TestDoctorLeavesStateAndHomeUnchanged runs the whole report over a fixture
// and compares every file of the home and the state directory before and after.
func TestDoctorLeavesStateAndHomeUnchanged(t *testing.T) {
	o, home, stateDir := doctorHome(t, "claude,codex,opencode")
	if err := os.MkdirAll(filepath.Join(home, ".claude", "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "sessions", "77.json"), []byte(`{"pid":77,"cwd":"/w","startedAt":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	f := newDoctorFake(home)
	f.alive[77] = true
	beforeHome, beforeState := collectFiles(t, home), collectFiles(t, stateDir)
	collectDoctor(o, t.TempDir(), f.deps())
	afterHome, afterState := collectFiles(t, home), collectFiles(t, stateDir)
	if fmt.Sprint(beforeHome) != fmt.Sprint(afterHome) {
		t.Fatal("the home changed")
	}
	if fmt.Sprint(beforeState) != fmt.Sprint(afterState) {
		t.Fatal("the state directory changed")
	}
}

func TestDoctorRenderTextIndentsTheDetailLineOfAnError(t *testing.T) {
	var b bytes.Buffer
	renderDoctorText(doctorReport{Installation: doctorSection{Title: "Installation", Err: "words\nDetail: raw"}}, &b)
	want := "Installation\n  Could not check everything: words\n    Detail: raw\n"
	if b.String() != want {
		t.Fatalf("got %q, want %q", b.String(), want)
	}
}

func TestDoctorRenderTextPrintsTitledSectionsAndSkipsPlaceholders(t *testing.T) {
	var b bytes.Buffer
	renderDoctorText(doctorReport{
		CLIs:         doctorSection{Title: "CLIs", Lines: []string{"a"}},
		Installation: doctorSection{Title: "Installation", Err: "boom", Lines: []string{"b"}},
	}, &b)
	want := "CLIs\n  a\n\nInstallation\n  Could not check everything: boom\n  b\n"
	if b.String() != want {
		t.Fatalf("got %q, want %q", b.String(), want)
	}
}

func TestDoctorRealDepsUseRealFunctions(t *testing.T) {
	d := realDoctorDeps()
	if d.lookPath == nil || d.runVersion == nil || d.processAlive == nil || d.now == nil || d.userHome == nil || d.getenv == nil {
		t.Fatalf("a dependency is missing: %+v", d)
	}
	if !d.processAlive(os.Getpid()) {
		t.Fatal("the current process is not alive")
	}
	if d.processAlive(0) || d.processAlive(-1) {
		t.Fatal("pid 0 and -1 must never be alive")
	}
}

func TestDoctorHostInstallationPrefersOwnedRowsAndNewestRelease(t *testing.T) {
	older := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	st := doctorState{
		released: map[string]time.Time{"old": older, "new": older.Add(time.Hour), "shared": older.Add(2 * time.Hour)},
		entries: []management.StatusEntry{
			{Host: "claude", Status: "retained_shared", Release: "shared", VersionStatus: "partial"},
			{Host: "claude", Status: "installed", Release: "old", VersionStatus: "verified"},
			{Host: "claude", Status: "drift", Release: "new", VersionStatus: "drift"},
			{Host: "claude", Kind: "voice", Status: "installed", Release: "voice"},
			{Host: "codex", Status: "retained_shared", Release: "shared", VersionStatus: "partial"},
			{Host: "pi", Status: "not_installed"},
		},
	}
	if got := st.hostInstallation("claude"); got != (hostInstall{Release: "new", State: "drift"}) {
		t.Fatalf("claude = %+v", got)
	}
	if got := st.hostInstallation("codex"); got != (hostInstall{Release: "shared", State: "partial"}) {
		t.Fatalf("codex = %+v", got)
	}
	if got := st.hostInstallation("pi"); got != (hostInstall{}) {
		t.Fatalf("pi = %+v", got)
	}
}
