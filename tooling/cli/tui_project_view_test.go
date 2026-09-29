package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Tests of the Project view.

// openProjectView opens the app and pushes the Project view over dir, the way
// the menu will. git replaces the runner when it is not nil.
func openProjectView(t *testing.T, dir string, git gitRunner, width, height int) (*appModel, *appDriver, *projectView) {
	t.Helper()
	cfg := testAppConfig(t)
	m, d := newTestApp(t, cfg, width, height)
	v := newProjectViewWith(cfg, dir, doctorDeps{})
	if git != nil {
		v.git = git
	}
	d.send(pushViewMsg{v: v})
	if m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
	return m, d, v
}

func TestProjectViewShowsVerdictPathAndValues(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			root := projectRepoWithSpecs(t, validHiveSection)
			_, d, _ := openProjectView(t, root, nil, size[0], size[1])
			d.mustShow("Project", "Valid", "Values read:", "Base branch: main", "Specs: _support/openspec")
			d.mustShow(filepath.Join(root, "AGENTS.md")[max(0, len(filepath.Join(root, "AGENTS.md"))-size[0]+2):])
			d.mustNotShow("Checking", "Loading")
			assertFits(t, d, size[0], size[1])
		})
	}
}

func TestProjectViewShowsFindings(t *testing.T) {
	root := projectRepoWithSpecs(t, strings.Replace(validHiveSection, "Base branch: main", "Base branch: nope", 1))
	_, d, v := openProjectView(t, root, nil, 80, 24)
	d.mustShow("Base branch: nope is not a local branch or on origin")
	d.mustNotShow("Valid", "r to retry")
	if v.failed {
		t.Fatal("a finding is not a failed check")
	}
}

func TestProjectViewOutsideGitShowsTwoLinesAndTheDirectory(t *testing.T) {
	dir := t.TempDir()
	_, d, _ := openProjectView(t, dir, nil, 80, 24)
	d.mustShow(outsideGitLine, outsideGitLineTwo, "Directory:")
	assertFits(t, d, 80, 24)
}

func TestProjectViewShowsSpinnerWhileLoading(t *testing.T) {
	v := newProjectViewWith(testAppConfig(t), t.TempDir(), doctorDeps{})
	if !v.NeedsSpinner() {
		t.Fatal("a loading view must ask for the spinner")
	}
	th := newAppTheme(true, false)
	out := stripANSI(v.View(viewCtx{Width: 80, Height: 21, Theme: &th, Spinner: "*"}))
	mustContain(t, out, "Project", "* Checking")
	if lines := strings.Split(out, "\n"); len(lines) > 21 {
		t.Fatalf("%d lines", len(lines))
	}
}

func TestProjectViewReloadsWithRAndIgnoresStaleResults(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	_, d, v := openProjectView(t, root, nil, 80, 24)
	d.mustShow("Valid")

	writeAgents(t, root, strings.Replace(validHiveSection, "- Tracker: GitHub Issues · org/demo\n", "", 1))
	d.key("r")
	d.mustShow("Tracker: required value is missing")
	d.mustNotShow("Valid")

	stale := projectLoadedMsg{owned: owned{v}, seq: v.seq - 1}
	stale.check.Section = doctorSection{Title: "Project", Lines: []string{"STALE"}}
	d.send(stale)
	d.mustNotShow("STALE")
	d.mustShow("Tracker: required value is missing")

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

func TestProjectViewLateResultAfterLeavingIsDropped(t *testing.T) {
	m, d, v := openProjectView(t, t.TempDir(), nil, 80, 24)
	d.key("esc")
	if _, ok := m.top().(*menuView); !ok {
		t.Fatalf("top view is %T", m.top())
	}
	late := projectLoadedMsg{owned: owned{v}, seq: v.seq}
	late.check.Section = doctorSection{Title: "Project", Lines: []string{"LATE"}}
	d.send(late)
	d.mustNotShow("LATE")
}

func TestProjectViewEscAndBackspaceReturnToMenu(t *testing.T) {
	for _, k := range []string{"esc", "backspace"} {
		t.Run(k, func(t *testing.T) {
			m, d, _ := openProjectView(t, t.TempDir(), nil, 80, 24)
			d.key(k)
			if _, ok := m.top().(*menuView); !ok {
				t.Fatalf("top view is %T after %s", m.top(), k)
			}
		})
	}
}

func TestProjectViewLoadErrorIsShownInsideWithRetry(t *testing.T) {
	broken := func(string, ...string) (string, error) { return "", errors.New("git is broken") }
	_, d, v := openProjectView(t, t.TempDir(), broken, 80, 24)
	d.mustShow("Could not check everything", "git is broken", "r to retry")
	assertFits(t, d, 80, 24)
	if !v.failed {
		t.Fatal("the view must be marked failed")
	}
}

func TestProjectViewScrollsToTheLastFindingAndBack(t *testing.T) {
	var b strings.Builder
	b.WriteString("## Hive\n")
	for i := 0; i < 60; i++ {
		fmt.Fprintf(&b, "- Extra%02d: %s\n", i, strings.Repeat("value ", 20))
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			root := projectRepoWithSpecs(t, b.String())
			_, d, v := openProjectView(t, root, nil, size[0], size[1])
			assertFits(t, d, size[0], size[1])
			d.mustShow("Project: required value is missing")
			if !v.box.scrollable() {
				t.Fatal("the fixture must not fit the screen")
			}
			for range 400 {
				d.key("down")
			}
			d.mustShow("Extra59: value")
			d.mustNotShow("Project: required value is missing")
			assertFits(t, d, size[0], size[1])
			for range 400 {
				d.key("up")
			}
			d.mustShow("Project: required value is missing", "lines 1-")
			for range 60 {
				d.key("pgdown")
			}
			d.mustShow("Extra59: value")
			for range 60 {
				d.key("pgup")
			}
			d.mustShow("Project: required value is missing")
		})
	}
}

func TestProjectViewClipsALongPathFromTheLeft(t *testing.T) {
	// A repository whose root path is longer than the screen.
	longRoot := filepath.Join(t.TempDir(), strings.Repeat("a-long-directory-name", 6))
	if err := os.MkdirAll(longRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, longRoot, "init", "-q", "-b", "main")
	writeAgents(t, longRoot, validHiveSection)
	_, d, _ := openProjectView(t, longRoot, nil, 80, 24)
	d.mustShow("…", "/AGENTS.md")
	assertFits(t, d, 80, 24)
}

func TestProjectViewLeavesTheRepositoryUnchanged(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	before, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	filesBefore := fmt.Sprint(collectFiles(t, root))
	_, d, _ := openProjectView(t, root, nil, 80, 24)
	d.key("r", "down", "up", "pgdown", "pgup")
	after, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !bytes.Equal(before, after) || filesBefore != fmt.Sprint(collectFiles(t, root)) {
		t.Fatal("the repository changed")
	}
}

func TestProjectViewIgnoresOtherMessages(t *testing.T) {
	v := newProjectViewWith(testAppConfig(t), t.TempDir(), doctorDeps{})
	if cmd, act := v.Update(tea.WindowSizeMsg{}); cmd != nil || act.nav != navNone {
		t.Fatalf("cmd=%v act=%+v", cmd, act)
	}
}

func TestProjectViewCurrentDirectoryConstructor(t *testing.T) {
	v := newProjectView(testAppConfig(t))
	if v.dir != "" || !v.loading {
		t.Fatalf("dir %q loading %v", v.dir, v.loading)
	}
}
