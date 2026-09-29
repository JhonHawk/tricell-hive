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
	d.mustShow("Diagnostics", "CLIs", "Installation", "Sessions", "CLI version 2.1.284", "Hive release "+shortHash(releaseIDOf(t, stateDir)), "No problems found")
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
	d.mustShow("CLI version 1.0.0")

	f.versions["/fake/bin/claude"] = "2.0.0\n"
	d.key("r")
	d.mustShow("CLI version 2.0.0")
	d.mustNotShow("CLI version 1.0.0")

	// A result from an older load is dropped.
	stale := doctorLoadedMsg{owned: owned{v}, seq: v.seq - 1}
	stale.sections[0] = doctorSection{Title: "CLIs", Lines: []string{"STALE"}}
	d.send(stale)
	d.mustNotShow("STALE")
	d.mustShow("CLI version 2.0.0")

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
		// The app's status line above the view also carries the raw error, so
		// look only from the view's own heading down.
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
