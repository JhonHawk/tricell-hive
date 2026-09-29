package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
)

// openIntegrationsView opens the app over o and pushes the Integrations view
// with the fake dependencies, the way the menu will.
func openIntegrationsView(t *testing.T, o management.Options, f *doctorFake, width, height int) (*appModel, *appDriver, *integrationsView) {
	t.Helper()
	cfg := testAppConfig(t)
	cfg.Options = o
	m, d := newTestApp(t, cfg, width, height)
	v := newIntegrationsViewWith(cfg, f.deps())
	d.send(pushViewMsg{v: v})
	if m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
	return m, d, v
}

// mustShowUnwrapped checks that text is on the screen even when the view wrapped
// it across lines: whitespace is ignored on both sides.
func mustShowUnwrapped(t *testing.T, d *appDriver, want string) {
	t.Helper()
	squeeze := func(s string) string { return strings.Join(strings.Fields(s), "") }
	if !strings.Contains(squeeze(d.screen()), squeeze(want)) {
		t.Fatalf("screen does not show %q, wrapped or not:\n%s", want, d.screen())
	}
}

// listRows are the screen lines of the four-row list (the cursor column, the
// name, and what follows).
func listRows(d *appDriver) []string {
	var rows []string
	for _, line := range d.lines() {
		for _, name := range []string{"Engram", "Context7", "pi-subagents", "agent-browser"} {
			if strings.HasPrefix(line, "> "+name) || strings.HasPrefix(line, "  "+name) {
				rows = append(rows, line)
			}
		}
	}
	return rows
}

func cursorRow(t *testing.T, d *appDriver) string {
	t.Helper()
	var found []string
	for _, line := range d.lines() {
		if strings.HasPrefix(line, "> ") {
			found = append(found, line)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d cursor rows on:\n%s", len(found), d.screen())
	}
	return found[0]
}

func TestIntegrationsViewShowsTitleHeaderAndFourRows(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	_, d, _ := openIntegrationsView(t, o, f, 80, 24)
	d.mustShow("Integrations", "Onboarding record", "No onboarding record yet")
	rows := listRows(d)
	if len(rows) != 4 {
		t.Fatalf("%d list rows:\n%s", len(rows), d.screen())
	}
	if !strings.HasPrefix(rows[0], "> Engram") {
		t.Fatalf("the cursor starts on %q", rows[0])
	}
	for i, name := range []string{"Engram", "Context7", "pi-subagents", "agent-browser"} {
		if !strings.Contains(rows[i], name) {
			t.Fatalf("row %d = %q, want %s", i, rows[i], name)
		}
	}
	assertFits(t, d, 80, 24)
}

func TestIntegrationsViewArrowsMoveTheCursorAndTheDetail(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	_, d, _ := openIntegrationsView(t, o, f, 80, 24)
	d.mustShow("Source: github.com/Gentleman-Programming/engram")
	d.key("down")
	if got := cursorRow(t, d); !strings.HasPrefix(got, "> Context7") {
		t.Fatalf("cursor = %q", got)
	}
	d.mustShow("Source: context7.com", "Last onboarding: No onboarding record yet", "Next step:")
	d.mustNotShow("Gentleman-Programming")
	d.key("down", "down")
	if got := cursorRow(t, d); !strings.HasPrefix(got, "> agent-browser") {
		t.Fatalf("cursor = %q", got)
	}
	d.mustShow("Source: "+agentBrowserSource, "Not part of onboarding")
	d.key("down") // the last row stays selected
	if got := cursorRow(t, d); !strings.HasPrefix(got, "> agent-browser") {
		t.Fatalf("down past the last row moved to %q", got)
	}
	d.key("up", "up", "up", "up")
	if got := cursorRow(t, d); !strings.HasPrefix(got, "> Engram") {
		t.Fatalf("up past the first row moved to %q", got)
	}
	assertFits(t, d, 80, 24)
}

// TestIntegrationsViewShowsTheConcreteNextStepOfTheSelectedRow covers M1 in the
// view: the detail names the command, and the same text is what hive doctor prints.
func TestIntegrationsViewShowsTheConcreteNextStepOfTheSelectedRow(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	_, d, _ := openIntegrationsView(t, o, f, 80, 24)
	mustShowUnwrapped(t, d, "Install it from github.com/Gentleman-Programming/engram")
	d.mustNotShow("validation is pending")
	d.key("down")
	mustShowUnwrapped(t, d, "To install it, run npx ctx7@latest setup --cli")
	mustShowUnwrapped(t, d, "sign in")
	assertFits(t, d, 80, 24)
	d.key("down")
	mustShowUnwrapped(t, d, "If you use Pi, install it with Pi's own package manager")
	assertFits(t, d, 80, 24)
}

func TestIntegrationsViewShowsLocalEvidenceAndRecordStatus(t *testing.T) {
	o, f, home, _ := integrationsFixture(t)
	writeOnboardingRecord(t, o, "manual", "engram")
	pathDir, marker := fakePrograms(t, "engram")
	skill := filepath.Join(home, ".agents", "skills", "find-docs", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := testAppConfig(t)
	cfg.Options = o
	deps := f.deps()
	deps.lookPath = exec.LookPath
	_, d := newTestApp(t, cfg, 80, 24)
	d.send(pushViewMsg{v: newIntegrationsViewWith(cfg, deps)})
	d.mustShow("> Engram", "detected", "manual", "Last onboarding: manual")
	mustShowUnwrapped(t, d, filepath.Join(pathDir, "engram"))
	d.key("down")
	mustShowUnwrapped(t, d, skill)
	mustNotExist(t, marker)
}

func TestIntegrationsViewShowsSpinnerWhileLoading(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	cfg := testAppConfig(t)
	cfg.Options = o
	v := newIntegrationsViewWith(cfg, f.deps())
	if !v.NeedsSpinner() {
		t.Fatal("a loading view must ask for the spinner")
	}
	th := newAppTheme(true, false)
	out := stripANSI(v.View(viewCtx{Width: 80, Height: 21, Theme: &th, Spinner: "*"}))
	mustContain(t, out, "Integrations", "* Checking")
}

func TestIntegrationsViewReloadsWithRAndIgnoresStaleResults(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	_, d, v := openIntegrationsView(t, o, f, 80, 24)
	d.mustShow("No onboarding record yet")
	writeOnboardingRecord(t, o, "manual", "engram")
	d.mustShow("No onboarding record yet") // nothing changes until r
	d.key("r")
	d.mustShow("Last onboarding: manual")

	stale := integrationsLoadedMsg{owned: owned{v}, seq: v.seq - 1, rows: []integrationRow{{ID: "engram", Name: "STALE"}}}
	d.send(stale)
	d.mustNotShow("STALE")

	cmd, _ := v.Update(keyMsg(t, "r"))
	if cmd == nil {
		t.Fatal("r did not reload")
	}
	seq := v.seq
	if cmd, _ := v.Update(keyMsg(t, "r")); cmd != nil || v.seq != seq {
		t.Fatal("r while loading started another load")
	}
}

func TestIntegrationsViewKeepsTheCursorAcrossAReload(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	_, d, _ := openIntegrationsView(t, o, f, 80, 24)
	d.key("down", "down", "r")
	if got := cursorRow(t, d); !strings.HasPrefix(got, "> pi-subagents") {
		t.Fatalf("cursor = %q after reload", got)
	}
}

func TestIntegrationsViewLateResultAfterLeavingIsDropped(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	cfg := testAppConfig(t)
	cfg.Options = o
	m, d := newTestApp(t, cfg, 80, 24)
	v := newIntegrationsViewWith(cfg, f.deps())
	d.send(pushViewMsg{v: v})
	d.key("esc")
	if _, ok := m.top().(*menuView); !ok {
		t.Fatalf("top view is %T", m.top())
	}
	d.send(integrationsLoadedMsg{owned: owned{v}, seq: v.seq, rows: []integrationRow{{ID: "engram", Name: "LATE"}}})
	d.mustNotShow("LATE")
}

func TestIntegrationsViewEscAndBackspaceReturnToMenu(t *testing.T) {
	for _, k := range []string{"esc", "backspace"} {
		t.Run(k, func(t *testing.T) {
			o, f, _, _ := integrationsFixture(t)
			m, d, _ := openIntegrationsView(t, o, f, 80, 24)
			d.key(k)
			if _, ok := m.top().(*menuView); !ok {
				t.Fatalf("top view is %T after %s", m.top(), k)
			}
		})
	}
}

func TestIntegrationsViewShowsATamperedRecordAsAnErrorInside(t *testing.T) {
	o, f, _, stateDir := integrationsFixture(t)
	writeOnboardingRecord(t, o, "manual", "engram")
	tamperOnboardingRecord(t, stateDir)
	_, d, _ := openIntegrationsView(t, o, f, 80, 24)
	d.mustShow("Integrations", "r to retry after fixing it", "record unreadable", "The record is damaged or unreadable", "hive install", "Detail: cannot read the last onboarding record", "invalid onboarding journal")
	d.mustNotShow("Last onboarding: Cannot read")
	d.mustNotShow("No onboarding record yet")
	assertFits(t, d, 80, 24)
	d.key("down", "down", "down") // agent-browser is not part of onboarding, and its local checks show
	d.mustShow("Source: "+agentBrowserSource, "r to retry after fixing it")
}

func TestIntegrationsViewLeavesStateAndHomeUnchanged(t *testing.T) {
	o, f, home, stateDir := integrationsFixture(t)
	writeOnboardingRecord(t, o, "manual", "engram")
	beforeHome, beforeState := collectFiles(t, home), collectFiles(t, stateDir)
	stateJSON, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, d, _ := openIntegrationsView(t, o, f, 80, 24)
	d.key("r", "down", "up", "pgdown", "pgup", "end", "home")
	afterState, _ := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if string(stateJSON) != string(afterState) {
		t.Fatal("state.json changed")
	}
	if fmt.Sprint(beforeHome) != fmt.Sprint(collectFiles(t, home)) || fmt.Sprint(beforeState) != fmt.Sprint(collectFiles(t, stateDir)) {
		t.Fatal("files changed")
	}
}

func TestIntegrationsViewFitsAtBothSizes(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			o, f, _, _ := integrationsFixture(t)
			writeOnboardingRecord(t, o, "manual", "engram", "context7", "pi-subagents")
			_, d, _ := openIntegrationsView(t, o, f, size[0], size[1])
			assertFits(t, d, size[0], size[1])
			for _, k := range []string{"down", "down", "down", "pgdown", "pgup", "end", "home"} {
				d.key(k)
				assertFits(t, d, size[0], size[1])
			}
		})
	}
}

func TestIntegrationsViewLongNextStepWrapsAndPageDownReachesItsLastLine(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	words := make([]string, 0, 120)
	for i := 1; i <= 120; i++ {
		words = append(words, fmt.Sprintf("word%03d", i))
	}
	longNext := strings.Join(words, " ") + " LASTWORD"
	_, d, v := openIntegrationsView(t, o, f, 80, 24)
	rows := make([]integrationRow, len(v.rows))
	copy(rows, v.rows)
	rows[0].Next = longNext
	d.send(integrationsLoadedMsg{owned: owned{v}, seq: v.seq, rows: rows})
	d.mustShow("Next step:")
	d.mustNotShow("LASTWORD")
	for range 20 {
		d.key("pgdown")
	}
	d.mustShow("LASTWORD")
	d.mustNotShow("…")
	assertFits(t, d, 80, 24)
	// Every word appears somewhere while paging: nothing was cut.
	seen := map[string]bool{}
	d.key("home")
	for range 20 {
		for _, w := range strings.Fields(d.screen()) {
			seen[w] = true
		}
		d.key("pgdown")
	}
	for _, w := range words {
		if !seen[w] {
			t.Fatalf("%s never appeared while paging", w)
		}
	}
	d.key("home")
	d.mustShow("Source:")
}

func TestIntegrationsViewWrapsALongPathWithoutCuttingIt(t *testing.T) {
	o, f, home, _ := integrationsFixture(t)
	skill := filepath.Join(home, ".agents", "skills", "find-docs", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, d, _ := openIntegrationsView(t, o, f, 80, 24)
	d.key("down")
	mustShowUnwrapped(t, d, skill)
}

func TestIntegrationsViewIgnoresOtherMessages(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	cfg := testAppConfig(t)
	cfg.Options = o
	v := newIntegrationsViewWith(cfg, f.deps())
	if cmd, act := v.Update(tea.WindowSizeMsg{}); cmd != nil || act.nav != navNone {
		t.Fatalf("cmd=%v act=%+v", cmd, act)
	}
}

func TestIntegrationsViewUsesRealDependenciesByDefault(t *testing.T) {
	v := newIntegrationsView(testAppConfig(t))
	if v.deps.lookPath == nil || v.deps.getenv == nil || v.deps.userHome == nil {
		t.Fatal("the production view has no dependencies")
	}
}
