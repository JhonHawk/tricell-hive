package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
	"tricell-hive/tooling/legacy"
	"tricell-hive/tooling/management"
)

// The tests here cover the load-error states of the CLIs and Voice views
// (issue #46): a managed file changed by hand must not empty the CLIs view,
// and a state that cannot be read must read in plain words first.

// driftedEnv installs hosts from the update fixture's catalog, whose skill is
// a shared file every host consumes, then edits that shared skill by hand, the
// way a user's change leaves a managed file in drift.
func driftedEnv(t *testing.T, hosts []string) updateEnv {
	t.Helper()
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	installDirect(t, env.home, env.stateDir, env.repo, hosts)
	path := env.sharedSkillPath()
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(original), "Preserve evidence.", "Edited by hand.", 1)
	if edited == string(original) {
		t.Fatal("setup: nothing was edited")
	}
	if err := os.WriteFile(path, []byte(edited), 0o600); err != nil {
		t.Fatal(err)
	}
	return env
}

func openDriftedCLIs(t *testing.T, env updateEnv, width, height int) *appDriver {
	t.Helper()
	_, d := newTestApp(t, hostsAppConfig(t, env.home, env.stateDir, env.repo, hostsTestDeps(coreOnlyAdapterFactory)), width, height)
	d.key("enter")
	d.mustShow("CLIs")
	return d
}

// TestHostsViewListsHostsWhenAManagedFileWasChanged (J1): a hand-edited shared
// skill makes the legacy scan fail; the view still lists every host with its
// state and drift, says in plain words what happened and where to go, and
// keeps the raw error on its own Detail line.
func TestHostsViewListsHostsWhenAManagedFileWasChanged(t *testing.T) {
	env := driftedEnv(t, installerHosts)
	skill := env.sharedSkillPath()
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		d := openDriftedCLIs(t, env, size[0], size[1])
		d.mustNotShow("Cannot read the CLIs", "Diagnostics shows")
		for _, want := range []string{
			skill + " differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or removing.",
			"Detail: " + skill + " differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere",
		} {
			mustShowFlat(d, want)
		}
		// Neither the plain words nor the raw detail call the file legacy.
		plain, detail, _ := strings.Cut(d.screen(), "Detail:")
		if strings.Contains(plain, "legacy") || strings.Contains(detail, "legacy") {
			t.Errorf("the plain words mention a legacy file:\n%s", d.screen())
		}
		rows := hostRows(d)
		if len(rows) != len(installerHosts) {
			t.Fatalf("%dx%d: %d rows, want %d:\n%s", size[0], size[1], len(rows), len(installerHosts), d.screen())
		}
		for _, r := range rows {
			if r.state != "registered" || !r.checked || r.drift == "0" || r.drift == "-" {
				t.Errorf("%s row = %+v, want registered, checked and drift counted", r.name, r)
			}
		}
		assertFits(t, d, size[0], size[1])
	}
}

// TestHostsViewRefusesChangesWhenAManagedFileWasChanged (J1): listing the hosts
// does not let a change through; both a removal and an installation are
// refused as before, and nothing is written.
func TestHostsViewRefusesChangesWhenAManagedFileWasChanged(t *testing.T) {
	t.Run("removal", func(t *testing.T) {
		env := driftedEnv(t, installerHosts)
		d := openDriftedCLIs(t, env, 80, 24)
		homeBefore, stateBefore := collectFiles(t, env.home), stateBytes(t, env.stateDir)
		toggle(t, d, "claude")
		d.key("a")
		mustShowFlat(d, "modified managed skill")
		assertFits(t, d, 80, 24)
		assertUntouched(t, env, homeBefore, stateBefore)
	})
	t.Run("installation", func(t *testing.T) {
		env := driftedEnv(t, installerHosts[:5])
		d := openDriftedCLIs(t, env, 80, 24)
		homeBefore, stateBefore := collectFiles(t, env.home), stateBytes(t, env.stateDir)
		toggle(t, d, "pi")
		d.key("a")
		mustShowFlat(d, "differs from what Hive expects there; undo the change")
		assertFits(t, d, 80, 24)
		assertUntouched(t, env, homeBefore, stateBefore)
	})
}

func assertUntouched(t *testing.T, env updateEnv, homeBefore map[string][]byte, stateBefore []byte) {
	t.Helper()
	after := collectFiles(t, env.home)
	if len(after) != len(homeBefore) {
		t.Errorf("the home has %d files, had %d", len(after), len(homeBefore))
	}
	for rel, data := range homeBefore {
		if !bytes.Equal(after[rel], data) {
			t.Errorf("%s changed", rel)
		}
	}
	if !bytes.Equal(stateBytes(t, env.stateDir), stateBefore) {
		t.Error("state.json changed")
	}
}

// TestHostsViewNamesAnUnknownLegacyScanFailureWithoutBlamingAFile (J1): a scan
// failure other than a changed file keeps the rows and gets the general note.
func TestHostsViewNamesAnUnknownLegacyScanFailureWithoutBlamingAFile(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.DiscoverHosts = func(o management.Options) ([]hostCandidate, error) {
		candidates, _ := detectInstallerHosts(o)
		for i := range candidates {
			candidates[i].Detected = true
		}
		return candidates, &legacyScanError{err: errors.New("unsupported legacy manifest: /x/manifest.json")}
	}
	_, d := openHostsApp(t, home, stateDir, minimalTestSource(t), deps)
	mustShowFlat(d, "Hive could not check for a legacy installation, so installing or removing may be refused. The Detail line says why.")
	mustShowFlat(d, "Detail: unsupported legacy manifest: /x/manifest.json")
	d.mustNotShow("differs from what Hive expects", "Open Diagnostics")
	if n := len(hostRows(d)); n != len(installerHosts) {
		t.Fatalf("%d rows, want %d:\n%s", n, len(installerHosts), d.screen())
	}
	assertFits(t, d, 80, 24)
}

// TestCLIsAndVoiceViewsExplainAnUnreadableStateInPlainWords (J2): with a
// corrupted state.json, each view leads with what happened and the way out,
// the same words as Diagnostics, and puts the raw error on its own Detail line
// after them.
func TestCLIsAndVoiceViewsExplainAnUnreadableStateInPlainWords(t *testing.T) {
	for _, tc := range []struct{ name, title, entry string }{
		{"CLIs", "The CLIs cannot be shown", "CLIs"},
		{"Voice", "The voice cannot be shown", "Voice"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newVoiceFixture(t)
			if err := os.WriteFile(filepath.Join(f.stateDir, "state.json"), []byte("{ this is not json"), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg := hostsAppConfig(t, f.home, f.stateDir, f.source, f.deps)
			_, d := newTestApp(t, cfg, 80, 24)
			openMenuEntry(t, d, tc.entry)
			mustShowFlat(d, tc.title)
			mustShowFlat(d, "Hive's state in "+cfg.Options.StateDir+" could not be read. Repair or restore its files; hive doctor shows the same problem.")
			d.mustNotShow("Cannot read the CLIs")
			detail, plain := -1, -1
			for i, l := range d.lines() {
				if plain < 0 && strings.Contains(l, "could not be read") {
					plain = i
				}
				if detail < 0 && strings.HasPrefix(l, "Detail: ") {
					detail = i
				}
			}
			if plain < 0 || detail <= plain {
				t.Fatalf("the raw error must follow the plain words on a Detail line (plain %d, detail %d):\n%s", plain, detail, d.screen())
			}
			assertFits(t, d, 80, 24)
		})
	}
}

// TestVoiceViewKeepsASourceProblemInItsOwnWords (J2): only a state that cannot
// be read gets the state wording; a source without voices keeps its guidance.
func TestVoiceViewKeepsASourceProblemInItsOwnWords(t *testing.T) {
	f := newVoiceFixture(t)
	f.source = t.TempDir()
	_, d := f.open(t, 80, 24)
	mustShowFlat(d, "Run hive from a Hive checkout or package, or pass --source")
	d.mustNotShow("could not be read")
}

// legacyPathEnv is a home with no state and no CLI, where the user's own file
// sits at a path Hive once installed a skill to (M1, M2).
func legacyPathEnv(t *testing.T) (home, stateDir, path string) {
	t.Helper()
	home, stateDir = newHostsTestHome(t)
	home, err := target.Canonical(home) // the scan reports the resolved path
	if err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(home, ".agents", "skills", "adversarial-research", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("My own notes.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return home, stateDir, path
}

// TestHostsViewShowsTheScanNoteWithNoHosts (M1, M2): with no CLI detected or
// registered and no state, a file at a legacy path still explains itself: the
// note names the path and what to do, without an assumption about who wrote
// the file and without sending the user to a Diagnostics that has nothing to
// say about it.
func TestHostsViewShowsTheScanNoteWithNoHosts(t *testing.T) {
	home, stateDir, path := legacyPathEnv(t)
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		_, d := newTestApp(t, hostsAppConfig(t, home, stateDir, minimalTestSource(t), defaultInstallDependencies(coreOnlyAdapterFactory)), size[0], size[1])
		d.key("enter")
		d.mustShow("CLIs")
		mustShowFlat(d, "No CLI hosts were detected or registered.")
		mustShowFlat(d, path+" differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or removing.")
		mustShowFlat(d, "Detail: "+path+" differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere")
		d.mustNotShow("Open Diagnostics", "Hive installed was changed", "Diagnostics shows")
		assertFits(t, d, size[0], size[1])
	}
}

// corruptState replaces state.json with text that is not JSON and returns a
// function that puts the original back (or removes the file when there was
// none), the way a user repairs it.
func corruptState(t *testing.T, stateDir string) (repair func()) {
	t.Helper()
	path := filepath.Join(stateDir, "state.json")
	original, readErr := os.ReadFile(path)
	if err := os.WriteFile(path, []byte("{ this is not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	return func() {
		if readErr != nil {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			return
		}
		if err := os.WriteFile(path, original, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// TestCLIsAndVoiceViewsRetryAfterTheStateIsFixed (M3): like Models, both views
// tell the user how to try again, list the key in the help bar, and load once
// the state is readable.
func TestCLIsAndVoiceViewsRetryAfterTheStateIsFixed(t *testing.T) {
	for _, tc := range []struct{ name, loaded string }{
		{"CLIs", "claude"},
		{"Voice", "Voice"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newVoiceFixture(t)
			repair := corruptState(t, f.stateDir)
			_, d := newTestApp(t, hostsAppConfig(t, f.home, f.stateDir, f.source, hostsTestDeps(coreOnlyAdapterFactory)), 80, 24)
			openMenuEntry(t, d, tc.name)
			mustShowFlat(d, "Press r to retry after fixing it.")
			d.mustShow("r reload")
			d.key("r") // still broken: the error stays, nothing else happens
			mustShowFlat(d, "could not be read")
			repair()
			d.key("r")
			d.mustNotShow("cannot be shown", "Press r to retry", "Detail:")
			d.mustShow(tc.loaded)
			assertFits(t, d, 80, 24)
			// Back at the menu, the header no longer says the state is unreadable.
			d.key("esc")
			d.mustNotShow("Status unavailable")
		})
	}
}

// TestVoiceSourceProblemOffersNoRetryText (M3): only an unreadable state gets
// the retry text; a source without voices keeps its own guidance.
func TestVoiceSourceProblemOffersNoRetryText(t *testing.T) {
	f := newVoiceFixture(t)
	f.source = t.TempDir()
	_, d := f.open(t, 80, 24)
	d.mustNotShow("Press r to retry", "r reload")
}

// TestCLIsViewNamesCursorsEditorOnlyLikeDiagnostics (M4): a host detected only
// through its editor launcher reads "editor only" in the CLIs view too, and
// stays installable; the CLI's own executable reads "detected".
func TestCLIsViewNamesCursorsEditorOnlyLikeDiagnostics(t *testing.T) {
	for _, tc := range []struct{ on, state string }{
		{"cursor", "editor only"},
		{"cursor-agent", "detected"},
	} {
		t.Run(tc.on, func(t *testing.T) {
			home, stateDir := newHostsTestHome(t)
			deps := defaultInstallDependencies(coreOnlyAdapterFactory)
			deps.DiscoverHosts = func(o management.Options) ([]hostCandidate, error) {
				return discoverInstallerHosts(o, func(name string) (string, error) {
					if name == tc.on {
						return "/fake/bin/" + name, nil
					}
					return "", errors.New("not found")
				})
			}
			_, d := openHostsApp(t, home, stateDir, minimalTestSource(t), deps)
			rows := hostRows(d)
			if len(rows) != 1 || rows[0].name != "cursor" || rows[0].state != tc.state {
				t.Fatalf("rows = %+v, want one cursor row with state %q:\n%s", rows, tc.state, d.screen())
			}
			d.key("space")
			if rows = hostRows(d); !rows[0].checked {
				t.Errorf("Cursor is not installable:\n%s", d.screen())
			}
			assertFits(t, d, 80, 24)
		})
	}
}

// TestDoctorCommandExplainsAnUnreadableState (L1): the words the views end on,
// "hive doctor shows the same problem", are true of the command.
func TestDoctorCommandExplainsAnUnreadableState(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	corruptState(t, stateDir)
	var runErr error
	out := captureStdout(t, func() {
		runErr = run([]string{"doctor", "--home", home, "--state-dir", stateDir, "--project", t.TempDir()})
	})
	if runErr != nil {
		t.Fatalf("doctor failed: %v\n%s", runErr, out)
	}
	if !strings.Contains(out, "could not be read") || !strings.Contains(out, "hive doctor shows the same problem") {
		t.Errorf("doctor does not explain the state problem:\n%s", out)
	}
}

// TestScanNoteUsesTheErrorTypeNotTheMessage (K1): the note recognizes a
// modified file by its error type, so it does not depend on the message text,
// and any other scan failure keeps the generic words.
func TestScanNoteUsesTheErrorTypeNotTheMessage(t *testing.T) {
	path := "/home/u/.agents/skills/x/SKILL.md"
	typed := &legacyScanError{err: fmt.Errorf("scanning: %w", &legacy.ModifiedFileError{Path: path})}
	words, detail := errLines(legacyScanNote(typed))
	if want := path + " differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or removing."; words != want {
		t.Errorf("words = %q, want %q", words, want)
	}
	if len(detail) != 1 || !strings.HasPrefix(detail[0], "Detail: ") {
		t.Errorf("detail = %q, want one Detail line", detail)
	}
	// The old message text alone no longer selects the specific words.
	text := &legacyScanError{err: errors.New("modified legacy file requires manual resolution: " + path)}
	if words, _ := errLines(legacyScanNote(text)); strings.Contains(words, path) {
		t.Errorf("the message text alone selected the specific words: %q", words)
	}
}

// TestInstallRefusalReadsInPlainWords (K1): planning an installation over a
// file at a path Hive once used is refused with words that do not call the
// file legacy, and writes nothing.
func TestInstallRefusalReadsInPlainWords(t *testing.T) {
	home, stateDir, path := legacyPathEnv(t)
	before := collectFiles(t, home)
	err := run([]string{"plan", "install", "--scope", "user", "--home", home, "--source", minimalTestSource(t), "--hosts", "codex", "--state-dir", stateDir})
	if err == nil {
		t.Fatal("plan install accepted a modified file")
	}
	want := path + " differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere"
	if err.Error() != want {
		t.Errorf("refusal = %q, want %q", err.Error(), want)
	}
	after := collectFiles(t, home)
	if len(after) != len(before) {
		t.Errorf("the home has %d files, had %d", len(after), len(before))
	}
	for rel, data := range before {
		if !bytes.Equal(after[rel], data) {
			t.Errorf("%s changed", rel)
		}
	}
}
