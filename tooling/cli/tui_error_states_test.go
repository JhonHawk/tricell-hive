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
		// The raw error says the same as the note, so no Detail line repeats it.
		d.mustNotShow("Cannot read the CLIs", "Diagnostics shows", "Detail:")
		mustShowFlat(d, skill+" differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or updating. Removing may also be refused if Hive installed that file.")
		if strings.Contains(d.screen(), "legacy") {
			t.Errorf("the note mentions a legacy file:\n%s", d.screen())
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
		mustShowFlat(d, "Nothing was changed. "+env.sharedSkillPath()+" differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere")
		if n := strings.Count(squash(d.screen()), squash(env.sharedSkillPath())); n != 1 {
			t.Errorf("the path appears %d times, want once:\n%s", n, d.screen())
		}
		d.mustNotShow("modified managed skill")
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
		// The refusal says what the note said: the note gives way to it.
		if n := strings.Count(spaced(d.screen()), "differs from what Hive expects there"); n != 1 {
			t.Errorf("the changed file is explained %d times, want once:\n%s", n, d.screen())
		}
		assertFits(t, d, 80, 24)
		assertUntouched(t, env, homeBefore, stateBefore)
	})
}

// squash removes all whitespace, so a path wrapped across lines can be counted.
func squash(s string) string { return strings.Join(strings.Fields(s), "") }

// spaced collapses every run of whitespace to one space, so a sentence wrapped
// across lines can be counted.
func spaced(s string) string { return strings.Join(strings.Fields(s), " ") }

// TestHostsViewAddingCursorWithAChangedSkillRefusesOnce (#66, AC4): adding a
// host while a shared skill was edited by hand shows one plain refusal that
// names the path once, with no raw error and no duplicate notice.
func TestHostsViewAddingCursorWithAChangedSkillRefusesOnce(t *testing.T) {
	var others []string
	for _, h := range installerHosts {
		if h != "cursor" {
			others = append(others, h)
		}
	}
	env := driftedEnv(t, others)
	skill := env.sharedSkillPath()
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		d := openDriftedCLIs(t, env, size[0], size[1])
		homeBefore, stateBefore := collectFiles(t, env.home), stateBytes(t, env.stateDir)
		toggle(t, d, "cursor")
		d.key("a")
		mustShowFlat(d, "Nothing was changed. "+skill+" differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere")
		if n := strings.Count(spaced(d.screen()), "differs from what Hive expects there"); n != 1 {
			t.Errorf("%dx%d: the changed file is explained %d times, want once:\n%s", size[0], size[1], n, d.screen())
		}
		if n := strings.Count(squash(d.screen()), squash(skill)); n != 1 {
			t.Errorf("%dx%d: the path appears %d times, want once:\n%s", size[0], size[1], n, d.screen())
		}
		d.mustNotShow("press m to read all", "modified managed skill")
		assertFits(t, d, size[0], size[1])
		assertUntouched(t, env, homeBefore, stateBefore)
	}
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
	mustShowFlat(d, "Hive could not check for a legacy installation, so installing or updating may be refused. The Detail line says why.")
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
		mustShowFlat(d, path+" differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or updating. Removing may also be refused if Hive installed that file.")
		d.mustNotShow("Detail:")
		d.mustNotShow("Open Diagnostics", "Hive installed was changed", "Diagnostics shows")
		assertFits(t, d, size[0], size[1])
	}
}

// TestHostsViewUninstallAllWorksWithAUserFileAtALegacyPath: a user's file at a
// path Hive once used, which the current catalogue does not manage, makes the
// legacy scan fail and shows the notice, but removing is not refused: Uninstall
// all still ends with "Hive removed".
func TestHostsViewUninstallAllWorksWithAUserFileAtALegacyPath(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude,codex,pi", "y\n", deps)
	canonical, err := target.Canonical(home) // the scan reports the resolved path
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(canonical, ".agents", "skills", "adversarial-research", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("My own notes.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, d := openHostsApp(t, home, stateDir, source, deps)
	mustShowFlat(d, path+" differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or updating. Removing may also be refused if Hive installed that file.")
	d.key("u", "left", "enter")
	d.mustShow("Hive removed")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the user's file was touched: %v", err)
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
			// K3: the state is explained under the table only when a row has it.
			const note = "editor only: the Cursor editor is installed, but not its CLI (cursor-agent). Hive can still install Cursor's files."
			if tc.state == "editor only" {
				mustShowFlat(d, note)
			} else {
				d.mustNotShow("editor only")
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
	if want := path + " differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or updating. Removing may also be refused if Hive installed that file."; words != want {
		t.Errorf("words = %q, want %q", words, want)
	}
	if len(detail) != 0 {
		t.Errorf("detail = %q, want none: the raw error repeats the note", detail)
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

// TestCLIsViewEditorOnlyNoteFitsWithOtherNotes (K3): the editor-only note sits
// under the table, before the pending-change line, and the screen still fits
// 80x24 with every host listed.
func TestCLIsViewEditorOnlyNoteFitsWithOtherNotes(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	deps := defaultInstallDependencies(coreOnlyAdapterFactory)
	deps.DiscoverHosts = func(o management.Options) ([]hostCandidate, error) {
		return discoverInstallerHosts(o, func(name string) (string, error) {
			if name == "cursor" || name == "claude" || name == "codex" {
				return "/fake/bin/" + name, nil
			}
			return "", errors.New("not found")
		})
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		_, d := newTestApp(t, hostsAppConfig(t, home, stateDir, minimalTestSource(t), deps), size[0], size[1])
		d.key("enter")
		d.mustShow("CLIs")
		mustShowFlat(d, "editor only: the Cursor editor is installed, but not its CLI (cursor-agent). Hive can still install Cursor's files.")
		d.key("space")
		mustShowFlat(d, "1 pending change: press a to review it.")
		// The note stands apart from the table and from the pending line.
		lines := d.lines()
		for i, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(l), "editor only:") && (i == 0 || strings.TrimSpace(lines[i-1]) != "") {
				t.Errorf("no blank line before the editor only note:\n%s", d.screen())
			}
			if strings.HasPrefix(strings.TrimSpace(l), "1 pending change") && (i == 0 || strings.TrimSpace(lines[i-1]) != "") {
				t.Errorf("no blank line before the pending line:\n%s", d.screen())
			}
		}
		assertFits(t, d, size[0], size[1])
	}
}

// TestScanNoteHidesOnlyForARefusalAboutTheSameFile: the note under the rows is
// redundant only when the refusal below it names the same path. A refusal about
// another file, in any of the wordings, leaves the note on screen.
func TestScanNoteHidesOnlyForARefusalAboutTheSameFile(t *testing.T) {
	const scanned = "/home/u/.agents/skills/x/SKILL.md"
	const other = "/home/u/.claude/CLAUDE.md"
	note := legacyScanNote(&legacyScanError{err: &legacy.ModifiedFileError{Path: scanned}})
	th := newAppTheme(true, false)
	refusals := map[string]error{
		"changed file":  &management.ManagedFileChangedError{Path: scanned, Kind: management.ManagedFileChanged},
		"legacy file":   &legacy.ModifiedFileError{Path: scanned},
		"block changed": &management.ManagedFileChangedError{Path: scanned, Kind: management.ManagedBlockChanged, Block: "Hive"},
		"wrapped":       fmt.Errorf("planning: %w", &management.ManagedFileChangedError{Path: scanned, Kind: management.ManagedFileMissing}),
	}
	shown := func(refusal error) bool {
		v := &hostsView{scanNote: note, scanPath: scanned}
		v.finishRefused(refusal, refusedText(refusal))
		return len(v.scanNoteLines(viewCtx{Width: 80, Theme: &th})) > 0
	}
	for name, err := range refusals {
		if shown(err) {
			t.Errorf("%s: the note stays although the refusal names the same file", name)
		}
	}
	for name, err := range map[string]error{
		"other changed file": &management.ManagedFileChangedError{Path: other, Kind: management.ManagedFileChanged},
		"other block":        &management.ManagedFileChangedError{Path: other, Kind: management.ManagedBlockChanged, Block: "Hive"},
		"no path":            errors.New("unowned Hive block"),
	} {
		if !shown(err) {
			t.Errorf("%s: the note was hidden by a refusal about something else", name)
		}
	}
}
