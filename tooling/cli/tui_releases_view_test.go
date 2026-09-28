package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"tricell-hive/tooling/management"
)

// Tests of the Releases view (T9).

var releaseRowPattern = regexp.MustCompile(`^(> |  )([0-9a-f]{12})  (.*)$`)

type releaseRow struct {
	cursor    bool
	short     string
	installed bool
	text      string
}

func releaseRows(d *appDriver) []releaseRow {
	var rows []releaseRow
	for _, line := range d.lines() {
		if m := releaseRowPattern.FindStringSubmatch(line); m != nil {
			rows = append(rows, releaseRow{cursor: m[1] == "> ", short: m[2], installed: strings.Contains(m[3], "(installed)"), text: strings.TrimRight(line, " ")})
		}
	}
	return rows
}

// syntheticReleases builds n releases, newest first, alternating consumers.
func syntheticReleases(n int) []management.ReleaseEntry {
	var entries []management.ReleaseEntry
	for i := n; i >= 1; i-- {
		hosts := []string{"codex", "claude"}
		if i%2 == 0 {
			hosts = []string{"claude", "cursor", "grok", "opencode", "pi", "codex"}
		}
		var consumers []management.Consumer
		for _, h := range hosts {
			consumers = append(consumers, management.Consumer{Host: h, Scope: "user", Context: "global"})
		}
		entries = append(entries, management.ReleaseEntry{
			ID:            fmt.Sprintf("%012x%052x", 0xa00000+i, i),
			LastWrittenAt: fmt.Sprintf("2026-09-%02dT10:00:00Z", 1+i%28),
			Commits:       []string{fmt.Sprintf("c%039x", i)},
			Consumers:     consumers,
		})
	}
	return entries
}

// openSyntheticReleases opens the Releases view over a home with one CLI
// installed and replaces its rows with entries.
func openSyntheticReleases(t *testing.T, entries []management.ReleaseEntry, installedID string, width, height int) (*appModel, *appDriver, *releasesView) {
	t.Helper()
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", deps)
	m, d := newTestApp(t, hostsAppConfig(t, home, stateDir, source, deps), width, height)
	openMenuEntry(t, d, "Releases")
	v, ok := m.top().(*releasesView)
	if !ok {
		t.Fatalf("top view is %T", m.top())
	}
	d.send(releasesLoadedMsg{seq: v.seq, entries: entries, installedID: installedID})
	return m, d, v
}

// TestReleasesViewListsNewestFirstAndMarksInstalled covers AC7: releases from
// the newest to the oldest, the installed one marked.
func TestReleasesViewListsNewestFirstAndMarksInstalled(t *testing.T) {
	env, olderID, _ := twoReleasesEnv(t)
	o := management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir}
	entries, err := management.Releases(o)
	if err != nil || len(entries) != 2 {
		t.Fatalf("setup: %v %d", err, len(entries))
	}
	installedID, err := installedReleaseID(o)
	if err != nil {
		t.Fatal(err)
	}
	cfg := hostsAppConfig(t, env.home, env.stateDir, env.repo, defaultInstallDependencies(coreOnlyAdapterFactory))
	_, d := newTestApp(t, cfg, 80, 24)
	openMenuEntry(t, d, "Releases")
	rows := releaseRows(d)
	if len(rows) != 2 {
		t.Fatalf("rows = %+v\n%s", rows, d.screen())
	}
	for i, r := range rows {
		if r.short != shortHash(entries[i].ID) {
			t.Errorf("row %d is %s, want %s (management.Releases order)", i, r.short, shortHash(entries[i].ID))
		}
		if r.installed != (entries[i].ID == installedID) {
			t.Errorf("row %d installed marker = %v", i, r.installed)
		}
		if w := lipgloss.Width(r.text); w > 78 {
			t.Errorf("row %d is %d columns: %q", i, w, r.text)
		}
	}
	if !rows[0].cursor {
		t.Error("the cursor does not start on the first row")
	}
	_ = olderID
}

// TestReleasesViewRollbackMatchesCommand covers AC7: choosing an older release
// shows the summary of `plan install --release` for the registered CLIs, and
// confirming leaves the same files and state as that plan applied.
func TestReleasesViewRollbackMatchesCommand(t *testing.T) {
	twin, twinOlderID, olderBody := twoReleasesEnv(t)
	plan, err := management.BuildPlan("install", management.Options{Scope: "user", Home: twin.home, StateDir: twin.stateDir, Hosts: []string{"codex", "claude"}, ReleaseID: twinOlderID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(plan); err != nil {
		t.Fatal(err)
	}

	env, olderID, _ := twoReleasesEnv(t)
	cfg := hostsAppConfig(t, env.home, env.stateDir, env.repo, defaultInstallDependencies(coreOnlyAdapterFactory))
	_, d := newTestApp(t, cfg, 80, 24)
	openMenuEntry(t, d, "Releases")
	rows := releaseRows(d)
	target := -1
	for i, r := range rows {
		if r.short == shortHash(olderID) {
			target = i
		}
	}
	if target < 0 || rows[target].installed {
		t.Fatalf("the older release is not listed as not installed: %+v", rows)
	}
	for range target {
		d.key("down")
	}
	d.key("enter")
	d.mustShow("Roll back to "+shortHash(olderID), "Hive files", "[Apply]")
	d.key("enter")
	d.mustShow("Hive rolled back")
	got, err := os.ReadFile(env.sharedSkillPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, olderBody) {
		t.Fatalf("the rollback did not restore the older content:\n%q", got)
	}
	// The marker moved to the older release.
	for _, r := range releaseRows(d) {
		if r.installed != (r.short == shortHash(olderID)) {
			t.Fatalf("installed marker after rollback: %+v", releaseRows(d))
		}
	}
	assertTwin(t, env.home, env.stateDir, twin.home, twin.stateDir)
}

// TestReleasesViewEmptyAndInstalledStates covers the limit states of the design
// table.
func TestReleasesViewEmptyAndInstalledStates(t *testing.T) {
	t.Run("no releases", func(t *testing.T) {
		home, stateDir := newHostsTestHome(t)
		_, d := newTestApp(t, hostsAppConfig(t, home, stateDir, minimalTestSource(t), defaultInstallDependencies(coreOnlyAdapterFactory)), 80, 24)
		openMenuEntry(t, d, "Releases")
		d.mustShow("No releases are retained yet")
		if len(releaseRows(d)) != 0 {
			t.Fatal("rows shown without releases")
		}
	})
	t.Run("the installed one", func(t *testing.T) {
		source := minimalTestSource(t)
		deps := hostsTestDeps(coreOnlyAdapterFactory)
		home, stateDir := newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "claude", "y\n", deps)
		before := collectFiles(t, home)
		_, d := newTestApp(t, hostsAppConfig(t, home, stateDir, source, deps), 80, 24)
		openMenuEntry(t, d, "Releases")
		d.mustShow("(installed)")
		d.key("enter")
		d.mustShow("Already installed")
		d.mustNotShow("[Apply]")
		if got := collectFiles(t, home); len(got) != len(before) {
			t.Fatal("choosing the installed release changed the home")
		}
	})
}

// TestReleasesViewValidationErrorShowsItsMessage covers AC7: a release the
// manager cannot validate shows the planner's message and the downgrade hint.
func TestReleasesViewValidationErrorShowsItsMessage(t *testing.T) {
	bogus := syntheticReleases(3)
	bogus[0].ID = "not-a-release"
	_, d, _ := openSyntheticReleases(t, bogus, "", 80, 24)
	d.key("enter")
	mustShowFlat(d, "invalid release ID")
	mustShowFlat(d, "plan that downgrade with the manager from the commit that produced it")
	d.mustNotShow("[Apply]")
}

// TestReleasesViewDecliningKeepsFiles covers AC8: rejecting the rollback
// summary changes nothing and leaves the message inside the view.
func TestReleasesViewDecliningKeepsFiles(t *testing.T) {
	env, olderID, _ := twoReleasesEnv(t)
	before := collectFiles(t, env.home)
	cfg := hostsAppConfig(t, env.home, env.stateDir, env.repo, defaultInstallDependencies(coreOnlyAdapterFactory))
	for _, reject := range [][]string{{"n"}, {"esc"}, {"right", "enter"}} {
		_, d := newTestApp(t, cfg, 80, 24)
		openMenuEntry(t, d, "Releases")
		for i, r := range releaseRows(d) {
			if r.short == shortHash(olderID) {
				for range i {
					d.key("down")
				}
			}
		}
		d.key("enter")
		d.mustShow("[Apply]")
		d.key(reject...)
		d.mustShow("Cancelled. No changes applied.", "Releases")
		d.mustNotShow("[Apply]")
	}
	after := collectFiles(t, env.home)
	for rel, data := range before {
		if !bytes.Equal(after[rel], data) {
			t.Fatalf("%s changed", rel)
		}
	}
}

// TestReleasesViewWriteBlocksKeys covers AC2/AC8 for the rollback write.
func TestReleasesViewWriteBlocksKeys(t *testing.T) {
	env, olderID, _ := twoReleasesEnv(t)
	cfg := hostsAppConfig(t, env.home, env.stateDir, env.repo, defaultInstallDependencies(coreOnlyAdapterFactory))
	m, d := newTestApp(t, cfg, 80, 24)
	openMenuEntry(t, d, "Releases")
	for i, r := range releaseRows(d) {
		if r.short == shortHash(olderID) {
			for range i {
				d.key("down")
			}
		}
	}
	d.key("enter")
	d.hold = true
	d.key("enter")
	if !m.isWriting() {
		t.Fatal("the apply is not marked as a write")
	}
	d.mustShow("Applying")
	before := d.screen()
	d.key("esc", "backspace", "enter", "ctrl+c", "down", "/")
	if d.quit || d.screen() != before {
		t.Fatal("keys acted during a write")
	}
	d.hold = false
	d.release()
	d.mustShow("Hive rolled back")
}

// TestReleasesViewFilterAndEsc covers AC7: `/` opens the filter, which matches a
// substring without regard to case; Esc clears it, and without a filter goes
// back; Backspace inside the filter deletes.
func TestReleasesViewFilterAndEsc(t *testing.T) {
	_, d, _ := openSyntheticReleases(t, syntheticReleases(30), "", 80, 24)
	all := len(releaseRows(d))
	if all == 0 {
		t.Fatalf("no rows:\n%s", d.screen())
	}
	d.key("/")
	typeText(d, "CURSOR") // only the even releases have the cursor host, even where the row cuts the host list
	if rows := releaseRows(d); len(rows) == 0 || len(rows) >= all {
		t.Fatalf("the filter did not narrow the list: %d rows of %d\n%s", len(rows), all, d.screen())
	}
	d.mustShow("CURSOR")
	d.key("backspace")
	d.mustShow("CURSO")
	d.mustNotShow("CURSOR")
	if len(menuRows(d)) != 0 {
		t.Fatal("Backspace in the filter left the view")
	}
	d.key("enter") // accepts the filter: the list has the focus again
	d.mustShow("CURSO")
	d.key("esc") // clears it
	d.mustNotShow("CURSO")
	if len(releaseRows(d)) != all || len(menuRows(d)) != 0 {
		t.Fatalf("Esc with a filter should only clear it:\n%s", d.screen())
	}
	d.key("esc")
	d.mustShow("> Releases")

	// Esc inside the open filter also clears it without leaving.
	_, d, _ = openSyntheticReleases(t, syntheticReleases(30), "", 80, 24)
	d.key("/")
	typeText(d, "grok")
	d.key("esc")
	if len(releaseRows(d)) != all || len(menuRows(d)) != 0 {
		t.Fatalf("Esc in the open filter should clear it:\n%s", d.screen())
	}
	// The filter also searches what the row cuts: opencode is the fifth host of
	// the even releases, beyond the ellipsis.
	d.key("/")
	typeText(d, "opencode")
	if rows := releaseRows(d); len(rows) == 0 || len(rows) >= all {
		t.Fatalf("the filter missed a host hidden by the row's ellipsis: %d rows of %d\n%s", len(rows), all, d.screen())
	}
	d.key("esc")
	// A filter with no match.
	d.key("/")
	typeText(d, "zzzz")
	d.mustShow("No release matches")
}

// TestReleasesViewHundredReleasesScrollToTheLast covers AC9: 100 releases scroll
// inside the screen at both sizes, the cursor reaches the last, and every row
// fits.
func TestReleasesViewHundredReleasesScrollToTheLast(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			entries := syntheticReleases(100)
			_, d, _ := openSyntheticReleases(t, entries, entries[3].ID, width, height)
			assertFits(t, d, width, height)
			rows := releaseRows(d)
			if len(rows) == 0 || len(rows) >= 100 {
				t.Fatalf("%d rows visible; the list should scroll", len(rows))
			}
			for _, r := range rows {
				if lipgloss.Width(r.text) > 78 {
					t.Fatalf("row wider than 78 columns: %q", r.text)
				}
			}
			for range 99 {
				d.key("down")
			}
			assertFits(t, d, width, height)
			last := shortHash(entries[99].ID)
			var cursorRow releaseRow
			for _, r := range releaseRows(d) {
				if r.cursor {
					cursorRow = r
				}
			}
			if cursorRow.short != last {
				t.Fatalf("the cursor is on %q, want the last release %s:\n%s", cursorRow.short, last, d.screen())
			}
			d.key("pgup")
			if r := releaseRows(d); len(r) == 0 {
				t.Fatal("rows vanished after PgUp")
			}
			d.key("home")
			if !releaseRows(d)[0].cursor {
				t.Fatal("Home did not return to the first row")
			}
			d.key("end")
			for _, r := range releaseRows(d) {
				if r.cursor && r.short != last {
					t.Fatalf("End left the cursor on %s", r.short)
				}
			}
			if !strings.Contains(strings.Join(d.lines(), "\n"), "…") {
				t.Fatalf("long host lists were cut without an ellipsis:\n%s", d.screen())
			}
		})
	}
}
