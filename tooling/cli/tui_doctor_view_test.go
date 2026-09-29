package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
)

// Tests of the Diagnostics view.

// openDoctorView opens the app over options and pushes the Diagnostics view
// with the fake dependencies, the way the menu will.
func openDoctorView(t *testing.T, o management.Options, f *doctorFake, width, height int) (*appModel, *appDriver, *doctorView) {
	t.Helper()
	cfg := testAppConfig(t)
	cfg.Options = o
	m, d := newTestApp(t, cfg, width, height)
	v := newDoctorViewWith(cfg, f.deps())
	d.send(pushViewMsg{v: v})
	if m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
	return m, d, v
}

func TestDoctorViewShowsThreeSectionsAndNotTheOthers(t *testing.T) {
	o, home, stateDir := doctorHome(t, "claude,codex")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("claude", "2.1.284\n")
	_, d, _ := openDoctorView(t, o, f, 80, 40)
	d.mustShow("Diagnostics", "CLIs", "Installation", "Sessions", "CLI 2.1.284", "Hive "+shortHash(releaseIDOf(t, stateDir)), "No problems found")
	d.mustNotShow("Integrations", "Project", "Loading", "Checking")
	assertFits(t, d, 80, 40)
}

func TestDoctorViewShowsSpinnerWhileLoading(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	cfg := testAppConfig(t)
	cfg.Options = o
	v := newDoctorViewWith(cfg, newDoctorFake(home).deps())
	if !v.NeedsSpinner() {
		t.Fatal("a loading view must ask for the spinner")
	}
	th := newAppTheme(true, false)
	out := stripANSI(v.View(viewCtx{Width: 80, Height: 21, Theme: &th, Spinner: "*"}))
	mustContain(t, out, "Diagnostics", "* Checking")
	if lines := strings.Split(out, "\n"); len(lines) > 21 {
		t.Fatalf("%d lines", len(lines))
	}
}

func TestDoctorViewReloadsWithRAndIgnoresStaleResults(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	f.install("claude", "1.0.0\n")
	_, d, v := openDoctorView(t, o, f, 80, 40)
	d.mustShow("CLI 1.0.0")

	f.versions["/fake/bin/claude"] = "2.0.0\n"
	d.key("r")
	d.mustShow("CLI 2.0.0")
	d.mustNotShow("CLI 1.0.0")

	// A result from an older load is dropped.
	stale := doctorLoadedMsg{owned: owned{v}, seq: v.seq - 1}
	stale.sections[0] = doctorSection{Title: "CLIs", Lines: []string{"STALE"}}
	d.send(stale)
	d.mustNotShow("STALE")
	d.mustShow("CLI 2.0.0")

	// A double r while loading starts one load only.
	cmd, _ := v.Update(keyMsg(t, "r"))
	if cmd == nil {
		t.Fatal("r did not reload")
	}
	seq := v.seq
	if cmd, _ := v.Update(keyMsg(t, "r")); cmd != nil || v.seq != seq {
		t.Fatal("r while loading started another load")
	}
}

func TestDoctorViewLateResultAfterLeavingIsDropped(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	cmdOwner := newDoctorFake(home)
	cfg := testAppConfig(t)
	cfg.Options = o
	m, d := newTestApp(t, cfg, 80, 24)
	v := newDoctorViewWith(cfg, cmdOwner.deps())
	d.send(pushViewMsg{v: v})
	d.key("esc")
	if _, ok := m.top().(*menuView); !ok {
		t.Fatalf("top view is %T", m.top())
	}
	late := doctorLoadedMsg{owned: owned{v}, seq: v.seq}
	late.sections[0] = doctorSection{Title: "CLIs", Lines: []string{"LATE"}}
	d.send(late)
	d.mustNotShow("LATE")
}

func TestDoctorViewEscAndBackspaceReturnToMenu(t *testing.T) {
	for _, k := range []string{"esc", "backspace"} {
		t.Run(k, func(t *testing.T) {
			o, home, _ := doctorHome(t, "claude")
			m, d, _ := openDoctorView(t, o, newDoctorFake(home), 80, 24)
			d.key(k)
			if _, ok := m.top().(*menuView); !ok {
				t.Fatalf("top view is %T after %s", m.top(), k)
			}
		})
	}
}

func TestDoctorViewShowsRecoverLineWhenNoHostIsRegistered(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	if err := os.WriteFile(filepath.Join(stateDir, "pending.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		_, d, _ := openDoctorView(t, o, newDoctorFake(home), size[0], size[1])
		d.mustShow("No CLI hosts are registered.", "Open CLIs from the menu", "An unfinished Hive operation is pending", "run hive recover --state-dir")
		d.mustNotShow("No problems found")
		assertFits(t, d, size[0], size[1])
	}
}

func TestDoctorViewLoadErrorLeadsWithWordsAndRetryAfterFixing(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		_, d, _ := openDoctorView(t, o, newDoctorFake(home), size[0], size[1])
		d.mustShow("Hive's state in", "could not be read", "hive status", "Detail: invalid character", "r to retry after fixing it")
		// Look only from the view's own heading down: the status line above
		// it never carries the raw error, but the heading keeps the check exact.
		lines := d.lines()
		start := 0
		for i, l := range lines {
			if strings.HasPrefix(l, "Diagnostics") {
				start = i
				break
			}
		}
		first, detail := -1, -1
		for i, l := range lines[start:] {
			if first < 0 && strings.Contains(l, "Could not check everything") {
				first = i
			}
			if detail < 0 && strings.Contains(l, "invalid character") {
				detail = i
			}
		}
		if first < 0 || detail <= first {
			t.Fatalf("the technical detail must come after the plain words (first %d, detail %d):\n%s", first, detail, d.screen())
		}
		assertFits(t, d, size[0], size[1])
	}
}

// TestDoctorViewUnreadableStateKeepsSessionsInSightAt80x24 covers N1: the
// explanation is drawn once, so the Sessions heading fits without scrolling.
func TestDoctorViewUnreadableStateKeepsSessionsInSightAt80x24(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	_, d, v := openDoctorView(t, o, newDoctorFake(home), 80, 24)
	d.mustShow("CLIs", "Installation", "Sessions", "Not checked: Hive's state could not be read (see above).", "r to retry after fixing it")
	if n := strings.Count(d.screen(), "Detail: "); n != 1 {
		t.Fatalf("Detail line drawn %d times, want 1:\n%s", n, d.screen())
	}
	if v.box.scrollable() {
		t.Fatalf("the unreadable-state text must fit without scrolling:\n%s", d.screen())
	}
	if !v.failed {
		t.Fatal("the view must stay marked failed")
	}
	assertFits(t, d, 80, 24)
}

func TestDoctorViewLoadErrorIsShownInsideWithRetry(t *testing.T) {
	home, _ := newHostsTestHome(t)
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: blocker}
	_, d, _ := openDoctorView(t, o, newDoctorFake(home), 80, 24)
	d.mustShow("Could not check everything", "r to retry")
	assertFits(t, d, 80, 24)
}

// crowdedDoctorFixture has all six CLIs registered and detected and many live
// Claude Code sessions with long paths, so the text is far taller than the screen.
func crowdedDoctorFixture(t *testing.T) (management.Options, *doctorFake) {
	t.Helper()
	o, home, _ := doctorHome(t, "claude,codex,cursor,grok,opencode,pi")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	for _, h := range installerHosts {
		f.install(h, h+" 1.2.3 "+strings.Repeat("v", 60)+"\n")
	}
	f.install("cursor-agent", "2026.09.23\n")
	dir := filepath.Join(home, ".claude")
	for i := 0; i < 30; i++ {
		pid := 1000 + i
		claudeSession(t, dir, fmt.Sprintf("%d.json", pid), pid, "/very/long/working/directory/"+strings.Repeat("segment/", 12)+fmt.Sprint(i), sessionBefore)
		f.alive[pid] = true
	}
	return o, f
}

func TestDoctorViewScrollsToTheLastRowAndBack(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			o, f := crowdedDoctorFixture(t)
			_, d, v := openDoctorView(t, o, f, size[0], size[1])
			assertFits(t, d, size[0], size[1])
			d.mustShow("CLIs")
			d.mustNotShow("Restart marks are estimates")
			if !v.box.scrollable() {
				t.Fatal("the fixture must not fit the screen")
			}
			for range 200 {
				d.key("down")
			}
			d.mustShow("Restart marks are estimates")
			d.mustNotShow("CLIs\n")
			assertFits(t, d, size[0], size[1])
			for range 200 {
				d.key("up")
			}
			d.mustShow("CLIs")
			d.key("pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown")
			d.mustShow("Restart marks are estimates")
			for range 12 {
				d.key("pgup")
			}
			d.mustShow("CLIs")
			d.mustShow("lines 1-")
		})
	}
}

// TestDoctorViewAnswersWhetherToRestartAndHowToRepairDrift covers M4 in the
// view: the Sessions summary leads its section, and Installation says how to get
// out of drift, at both sizes and all the way down the scrolled text.
func TestDoctorViewAnswersWhetherToRestartAndHowToRepairDrift(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			o, home, _ := doctorHome(t, "claude,codex")
			o = nonSyntheticOptions(t, o, home)
			f := newDoctorFake(home)
			claudeSession(t, filepath.Join(home, ".claude"), "2001.json", 2001, "/work/old", sessionBefore)
			claudeSession(t, filepath.Join(home, ".claude"), "2002.json", 2002, "/work/new", sessionAfter)
			f.alive[2001], f.alive[2002] = true, true
			setReleaseTime(t, o.StateDir, sessionRelease)
			claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
			data, err := os.ReadFile(claudeMD)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(claudeMD, []byte(strings.Replace(string(data), "Minimal test guidance.", "Edited by hand.", 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			_, d, _ := openDoctorView(t, o, f, size[0], size[1])
			assertFits(t, d, size[0], size[1])
			var seen strings.Builder
			for range 12 {
				seen.WriteString(d.screen())
				d.key("pgdown")
				assertFits(t, d, size[0], size[1])
			}
			squeezed := strings.Join(strings.Fields(seen.String()), "")
			for _, want := range []string{
				"1 open Claude Code session should be restarted",
				"claude: 2 open sessions, 1 to restart",
				driftRepairLine,
			} {
				if !strings.Contains(squeezed, strings.Join(strings.Fields(want), "")) {
					t.Fatalf("the view never showed %q:\n%s", want, seen.String())
				}
			}
		})
	}
}

// TestDoctorViewShowsTheSessionsAnswerOnTheFirstScreenAt80x24 covers M3: with
// six registered and detected CLIs, real version strings and nothing in drift,
// every CLIs row but the longest fits one line at 80 columns, so the Sessions
// headline is on the first screen with at least two lines of margin below it.
func TestDoctorViewShowsTheSessionsAnswerOnTheFirstScreenAt80x24(t *testing.T) {
	o, home, stateDir := doctorHome(t, "claude,codex,cursor,grok,opencode,pi")
	o = nonSyntheticOptions(t, o, home)
	f := newDoctorFake(home)
	versions := map[string]string{
		"claude":   "2.1.284 (Claude Code)",
		"codex":    "codex-cli 0.159.0",
		"cursor":   "2026.09.23-86fc751",
		"grok":     "grok 1.0.45 (c33bff361a6f) [alpha]",
		"opencode": "opencode v2.0.19",
		"pi":       "0.87.1",
	}
	for host, v := range versions {
		f.install(host, "unused\n")
		f.install(versionBinary(host), v+"\n")
	}
	_, d, v := openDoctorView(t, o, f, 80, 24)
	assertFits(t, d, 80, 24)
	short := shortHash(releaseIDOf(t, stateDir))
	lines := d.lines()
	for host, version := range versions {
		if host == "grok" {
			continue // its version alone is 34 characters and may wrap
		}
		found := false
		for _, l := range lines {
			l = strings.TrimSpace(l)
			found = found || (strings.HasPrefix(l, host+" ") && strings.HasSuffix(l, "CLI "+version) && strings.Contains(l, "Hive "+short))
		}
		if !found {
			t.Errorf("the %s row is not on one line at 80 columns:\n%s", host, d.screen())
		}
	}
	if v.box.vp.YOffset() != 0 {
		t.Fatalf("the view opened scrolled:\n%s", d.screen())
	}
	shown := strings.Split(stripANSI(v.box.vp.View()), "\n")
	headline := -1
	for i, l := range shown {
		if strings.Contains(l, "No open Claude Code or Grok session needs a restart") {
			headline = i
		}
	}
	if headline < 0 {
		t.Fatalf("the Sessions headline is not on the first screen:\n%s", d.screen())
	}
	if margin := len(shown) - 1 - headline; margin < 2 {
		t.Fatalf("only %d lines below the Sessions headline, want at least 2:\n%s", margin, d.screen())
	}
	d.mustShow("Hive cannot see Codex, Pi or Cursor sessions")
}

func TestDoctorViewLeavesStateAndHomeUnchanged(t *testing.T) {
	o, home, stateDir := doctorHome(t, "claude,codex")
	f := newDoctorFake(home)
	beforeHome, beforeState := collectFiles(t, home), collectFiles(t, stateDir)
	stateJSON, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, d, _ := openDoctorView(t, o, f, 80, 24)
	d.key("r", "down", "up", "pgdown")
	afterState, _ := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if string(stateJSON) != string(afterState) {
		t.Fatal("state.json changed")
	}
	if fmt.Sprint(beforeHome) != fmt.Sprint(collectFiles(t, home)) || fmt.Sprint(beforeState) != fmt.Sprint(collectFiles(t, stateDir)) {
		t.Fatal("files changed")
	}
}

func TestDoctorViewIgnoresOtherMessages(t *testing.T) {
	o, home, _ := doctorHome(t, "claude")
	cfg := testAppConfig(t)
	cfg.Options = o
	v := newDoctorViewWith(cfg, newDoctorFake(home).deps())
	if cmd, act := v.Update(tea.WindowSizeMsg{}); cmd != nil || act.nav != navNone {
		t.Fatalf("cmd=%v act=%+v", cmd, act)
	}
}

// screenLineOf returns the 0-based screen line that starts with prefix, or -1.
func screenLineOf(d *appDriver, prefix string) int {
	for i, l := range d.lines() {
		if strings.HasPrefix(strings.TrimSpace(l), prefix) {
			return i
		}
	}
	return -1
}

// I3 in the view: six hosts sharing one drifted file draw one row, so the view
// fits both sizes and Sessions stays reachable. At 120x40 everything fits
// without scrolling. At 80x24 the headline is on the first screen only when the
// path is short (about 70 characters leave it on the last visible line); the
// temporary path here is longer, so the test only requires that it can be reached.
func TestDoctorViewGroupsSharedDriftedFileAndKeepsSessionsReachable(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			o, home, paths := sharedSkillHome(t, "alpha")
			o = nonSyntheticOptions(t, o, home)
			driftFile(t, paths[0])
			f := newDoctorFake(home)
			for host, v := range map[string]string{"claude": "2.1.284 (Claude Code)", "codex": "codex-cli 0.159.0", "grok": "grok 1.0.45 (c33bff361a6f) [alpha]", "opencode": "opencode v2.0.19", "pi": "0.87.1"} {
				f.install(host, v+"\n")
			}
			f.install("cursor-agent", "2026.09.23-86fc751\n")
			_, d, v := openDoctorView(t, o, f, size[0], size[1])
			assertFits(t, d, size[0], size[1])
			d.mustShow("drift  claude, codex, cursor, grok, opencode, pi")
			if n := strings.Count(strings.Join(strings.Fields(d.screen()), ""), strings.Join(strings.Fields(paths[0]), "")); n > 1 {
				t.Fatalf("the path is drawn %d times:\n%s", n, d.screen())
			}
			if size[0] == 120 {
				if v.box.scrollable() || screenLineOf(d, "Sessions") < 0 {
					t.Fatalf("at 120x40 the view must fit without scrolling:\n%s", d.screen())
				}
				return
			}
			seen := d.screen()
			for range 12 {
				d.key("pgdown")
				seen += d.screen()
			}
			if !strings.Contains(seen, "Sessions") {
				t.Fatalf("Sessions never shown:\n%s", seen)
			}
		})
	}
}
