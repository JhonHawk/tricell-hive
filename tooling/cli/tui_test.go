package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// interfaceTestOptions builds a fresh synthetic home and state directory for
// one test, mirroring the pattern install_test.go and update_test.go use.
func interfaceTestOptions(t *testing.T) options {
	t.Helper()
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	source := minimalTestSource(t)
	return options{Home: home, StateDir: stateDir, Source: source}
}

// TestMenuStatusLineNoHosts covers the empty-state line design.md "La
// interfaz" requires when nothing is registered.
func TestMenuStatusLineNoHosts(t *testing.T) {
	o := interfaceTestOptions(t)
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir}
	line, err := interfaceStatusLine(mo)
	if err != nil {
		t.Fatal(err)
	}
	if line != "No CLI hosts are registered." {
		t.Fatalf("got %q", line)
	}
}

// ---------------------------------------------------------------------------
// Full-screen application (T7): driver and tests.
// ---------------------------------------------------------------------------

// driverStepCap bounds how many messages one driver call may process. A Cmd
// that keeps rescheduling itself would otherwise hang the test; hitting the
// cap fails it instead (design.md "Pruebas").
const driverStepCap = 500

// ansiPattern matches the CSI sequences lipgloss and Bubble Tea emit.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;:?<=>]*[ -/]*[@-~]`)

func stripANSI(s string) string { return ansiPattern.ReplaceAllString(s, "") }

// colorSGR reports whether s holds an SGR sequence that sets a foreground or
// background color (30-38, 40-48, 90-97, 100-107). Attributes such as bold
// and the plain reset are not colors.
func colorSGR(s string) bool {
	for _, seq := range regexp.MustCompile(`\x1b\[([0-9;:]*)m`).FindAllStringSubmatch(s, -1) {
		for _, tok := range strings.FieldsFunc(seq[1], func(r rune) bool { return r == ';' || r == ':' }) {
			n, err := strconv.Atoi(tok)
			if err != nil {
				continue
			}
			if (n >= 30 && n <= 38) || (n >= 40 && n <= 48) || (n >= 90 && n <= 97) || (n >= 100 && n <= 107) {
				return true
			}
		}
	}
	return false
}

// appDriver drives a Bubble Tea model without a terminal: it sends messages,
// runs every returned Cmd synchronously in the test goroutine, expands
// tea.BatchMsg, drops the spinner and text-cursor tick messages, and fails
// the test when one call needs more than driverStepCap steps. With hold set
// it keeps the Cmds returned by an Update that starts a write, so a test can
// assert the intermediate state and release them later.
type appDriver struct {
	t         *testing.T
	model     tea.Model
	isWriting func() bool
	hold      bool
	held      []tea.Cmd
	quit      bool
	fail      func(format string, args ...any)
}

func newAppDriver(t *testing.T, m tea.Model) *appDriver {
	t.Helper()
	d := &appDriver{t: t, model: m, fail: func(format string, args ...any) {
		t.Helper()
		t.Fatalf(format, args...)
	}}
	if w, ok := m.(interface{ isWriting() bool }); ok {
		d.isWriting = w.isWriting
	} else {
		d.isWriting = func() bool { return false }
	}
	return d
}

// boot sends the window size, then runs Init and everything it schedules.
func (d *appDriver) boot(width, height int) {
	d.t.Helper()
	d.dispatch(tea.WindowSizeMsg{Width: width, Height: height})
	if cmd := d.model.Init(); cmd != nil {
		d.run(cmd)
	}
}

func (d *appDriver) send(msg tea.Msg) {
	d.t.Helper()
	d.dispatch(msg)
}

func (d *appDriver) resize(width, height int) {
	d.t.Helper()
	d.dispatch(tea.WindowSizeMsg{Width: width, Height: height})
}

func (d *appDriver) key(names ...string) {
	d.t.Helper()
	for _, name := range names {
		d.dispatch(keyMsg(d.t, name))
	}
}

// release runs the Cmds held while a write was in progress.
func (d *appDriver) release() {
	d.t.Helper()
	held := d.held
	d.held = nil
	for _, cmd := range held {
		d.run(cmd)
	}
}

func (d *appDriver) run(cmd tea.Cmd) {
	d.t.Helper()
	d.process([]tea.Msg{cmd()})
}

func (d *appDriver) dispatch(msg tea.Msg) {
	d.t.Helper()
	d.process([]tea.Msg{msg})
}

func (d *appDriver) process(queue []tea.Msg) {
	d.t.Helper()
	for steps := 0; len(queue) > 0; steps++ {
		if steps >= driverStepCap {
			d.fail("driver exceeded %d steps: a Cmd keeps rescheduling itself", driverStepCap)
			return
		}
		msg := queue[0]
		queue = queue[1:]
		switch m := msg.(type) {
		case nil:
			continue
		case tea.QuitMsg:
			d.quit = true
			continue
		case tea.BatchMsg:
			for _, cmd := range m {
				if cmd != nil {
					queue = append(queue, cmd())
				}
			}
			continue
		}
		if isTickNoise(msg) {
			continue
		}
		_, cmd := d.model.Update(msg)
		if cmd == nil {
			continue
		}
		if d.hold && d.isWriting() {
			d.held = append(d.held, cmd)
			continue
		}
		queue = append(queue, cmd())
	}
}

// isTickNoise recognizes the periodic messages the driver drops: the spinner's
// frames and the text input's cursor blink.
func isTickNoise(msg tea.Msg) bool {
	name := fmt.Sprintf("%T", msg)
	return name == "spinner.TickMsg" || strings.HasPrefix(name, "cursor.")
}

func keyMsg(t *testing.T, name string) tea.KeyPressMsg {
	t.Helper()
	switch name {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "pgup":
		return tea.KeyPressMsg{Code: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyPressMsg{Code: tea.KeyPgDown}
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	}
	r := []rune(name)
	if len(r) != 1 {
		t.Fatalf("unknown key name %q", name)
	}
	return tea.KeyPressMsg{Code: r[0], Text: name}
}

// raw returns the model's current View content, escape sequences included.
func (d *appDriver) raw() string {
	d.t.Helper()
	return d.model.View().Content
}

// screen returns the model's current View content as plain text.
func (d *appDriver) screen() string {
	d.t.Helper()
	return stripANSI(d.raw())
}

// lines splits the plain screen into lines.
func (d *appDriver) lines() []string { return strings.Split(d.screen(), "\n") }

func (d *appDriver) mustShow(want ...string) {
	d.t.Helper()
	s := d.screen()
	for _, w := range want {
		if !strings.Contains(s, w) {
			d.t.Fatalf("screen does not show %q:\n%s", w, s)
		}
	}
}

func (d *appDriver) mustNotShow(unwanted ...string) {
	d.t.Helper()
	s := d.screen()
	for _, w := range unwanted {
		if strings.Contains(s, w) {
			d.t.Fatalf("screen unexpectedly shows %q:\n%s", w, s)
		}
	}
}

// testAppConfig builds an application config over a synthetic home. The
// default dependencies report no pending operation and use the core-only
// adapter double.
func testAppConfig(t *testing.T) appConfig {
	t.Helper()
	o := interfaceTestOptions(t)
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir, Source: o.Source}
	// The application always receives normalized options (runApp does the
	// same): on macOS the temporary directory sits under the /var symlink.
	if _, stateDir, err := management.NormalizeOptions(mo); err != nil {
		t.Fatal(err)
	} else {
		mo.StateDir = stateDir
	}
	return appConfig{
		Options:          mo,
		ExplicitStateDir: true,
		Deps:             defaultInstallDependencies(coreOnlyAdapterFactory),
		Dark:             true,
	}
}

func newTestApp(t *testing.T, cfg appConfig, width, height int) (*appModel, *appDriver) {
	t.Helper()
	m := newAppModel(cfg)
	d := newAppDriver(t, m)
	d.boot(width, height)
	return m, d
}

var menuRowPattern = regexp.MustCompile(`^(> |  )(CLIs|Update|Releases|Voice|Diagnostics|Models|Integrations|Project|Quit)(?:\s|$)`)

// menuRows returns the menu entries visible on the screen, in order.
func menuRows(d *appDriver) []string {
	var rows []string
	for _, line := range d.lines() {
		if m := menuRowPattern.FindStringSubmatch(line); m != nil {
			rows = append(rows, m[2])
		}
	}
	return rows
}

// TestAppDriverFailsOnIterationCap proves the driver stops a Cmd that keeps
// rescheduling itself instead of hanging (design.md "Pruebas").
func TestAppDriverFailsOnIterationCap(t *testing.T) {
	d := newAppDriver(t, &loopModel{})
	var failure string
	type abort struct{}
	d.fail = func(format string, args ...any) {
		failure = fmt.Sprintf(format, args...)
		panic(abort{})
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(abort); !ok {
					panic(r)
				}
			}
		}()
		d.boot(80, 24)
	}()
	if !strings.Contains(failure, "keeps rescheduling") {
		t.Fatalf("driver did not fail on the iteration cap, failure = %q", failure)
	}
}

type loopMsg struct{}

func loopCmd() tea.Msg { return loopMsg{} }

// loopModel reschedules a Cmd on every message, forever.
type loopModel struct{}

func (loopModel) Init() tea.Cmd                          { return loopCmd }
func (m *loopModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, loopCmd }
func (loopModel) View() tea.View                         { return tea.NewView("") }

// TestAppMenuHasNineEntriesInOrder covers AC1: the menu lists exactly CLIs,
// Update, Releases, Voice, Diagnostics, Models, Integrations, Project and
// Quit (#46 read-only views), with the status line above it and the help bar
// below.
func TestAppMenuHasNineEntriesInOrder(t *testing.T) {
	_, d := newTestApp(t, testAppConfig(t), 80, 24)
	want := []string{"CLIs", "Update", "Releases", "Voice", "Diagnostics", "Models", "Integrations", "Project", "Quit"}
	got := menuRows(d)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("menu entries = %v, want %v\n%s", got, want, d.screen())
	}
	d.mustNotShow("Status", "Install CLIs", "Remove CLIs")
	lines := d.lines()
	if !strings.Contains(lines[0], "No CLI hosts are registered") {
		t.Fatalf("first line is not the status line: %q", lines[0])
	}
	if last := lines[len(lines)-1]; !strings.Contains(last, "ctrl+c quit") {
		t.Fatalf("help bar is missing on the last line: %q", last)
	}
	if !d.model.View().AltScreen {
		t.Fatal("the view does not request the alternate screen")
	}
}

// TestAppStatusLineLoadsWithSpinner covers design.md "Aplicación": the status
// line is computed by a Cmd, and the header shows a loading indicator until it
// arrives.
func TestAppStatusLineLoadsWithSpinner(t *testing.T) {
	m := newAppModel(testAppConfig(t))
	d := newAppDriver(t, m)
	d.dispatch(tea.WindowSizeMsg{Width: 80, Height: 24})
	d.mustShow("Loading status")
	d.run(m.Init())
	d.mustNotShow("Loading status")
	d.mustShow("No CLI hosts are registered")
}

// TestAppStatusLineShowsRegisteredState checks the status line for a home
// with a registered host (same text the earlier interface showed).
func TestAppStatusLineShowsRegisteredState(t *testing.T) {
	cfg := testAppConfig(t)
	installViaText(t, cfg.Options.Home, cfg.Options.StateDir, cfg.Options.Source, "codex", "y\n", defaultInstallDependencies(coreOnlyAdapterFactory))
	_, d := newTestApp(t, cfg, 80, 24)
	d.mustShow("1 CLI host")
	d.mustNotShow("Loading status")
}

// TestAppStatusLineSaysInPlainWordsThatStateCannotBeRead covers I1: with an
// unreadable state.json the status line names what happened and where to look,
// on one line that fits 80 columns, and never prints the raw Go error (the
// Diagnostics view carries it).
func TestAppStatusLineSaysInPlainWordsThatStateCannotBeRead(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		cfg := testAppConfig(t)
		if err := os.WriteFile(filepath.Join(cfg.Options.StateDir, "state.json"), []byte("null garbage"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, d := newTestApp(t, cfg, size[0], size[1])
		d.mustShow("Status unavailable: state could not be read; see Diagnostics.")
		d.mustNotShow("invalid character", "json", "unmarshal", "…")
		header := d.lines()[0]
		if w := lipgloss.Width(header); w > 80 {
			t.Fatalf("status line is %d columns wide: %q", w, header)
		}
		assertFits(t, d, size[0], size[1])
	}
}

// TestAppKeysEscAndBackspaceReturnFromPlaceholder covers AC2: from each view
// the menu opens, Esc and Backspace both return to the menu (Backspace also in
// the Update view, whose first field has the focus only after typing: with
// nothing to delete it must still not go back), and the cursor stays on the
// entry that was opened. It keeps its T7 name.
func TestAppKeysEscAndBackspaceReturnFromPlaceholder(t *testing.T) {
	entries := []string{"CLIs", "Update", "Releases", "Voice", "Diagnostics", "Models", "Integrations", "Project"}
	marker := func(name string) string {
		return map[string]string{
			"CLIs":         "No CLI hosts were detected or registered",
			"Update":       "Revision",
			"Releases":     "No releases are retained yet",
			"Voice":        "Open CLIs to install one",
			"Diagnostics":  "Sessions",
			"Models":       "Open CLIs from the menu",
			"Integrations": "No onboarding record yet",
			"Project":      "AGENTS.md",
		}[name]
	}
	for i, name := range entries {
		for _, back := range []string{"esc", "backspace"} {
			if name == "Update" && back == "backspace" {
				continue // Backspace edits the Update view's text fields; covered by TestUpdateViewBackspaceEditsAndNeverGoesBack
			}
			t.Run(name+"/"+back, func(t *testing.T) {
				_, d := newTestApp(t, testAppConfig(t), 80, 24)
				for range i {
					d.key("down")
				}
				d.key("enter")
				d.mustShow(marker(name))
				if len(menuRows(d)) != 0 {
					t.Fatalf("the menu is still drawn under the view:\n%s", d.screen())
				}
				d.key(back)
				d.mustNotShow(marker(name))
				if got := len(menuRows(d)); got != len(mainMenuItems) {
					t.Fatalf("menu not shown after %s, rows=%d\n%s", back, got, d.screen())
				}
				d.mustShow("> " + name)
				if d.quit {
					t.Fatalf("%s from a view must not quit", back)
				}
			})
		}
	}
}

// TestAppKeysBackspaceAtMenuDoesNothing covers AC2: Backspace at the menu is
// a no-op, while Esc there leaves the application.
func TestAppKeysBackspaceAtMenuDoesNothing(t *testing.T) {
	_, d := newTestApp(t, testAppConfig(t), 80, 24)
	before := d.screen()
	d.key("backspace")
	if d.quit {
		t.Fatal("Backspace at the menu quit the application")
	}
	if after := d.screen(); after != before {
		t.Fatalf("Backspace at the menu changed the screen:\n%s\n--- was ---\n%s", after, before)
	}
	d.key("esc")
	if !d.quit {
		t.Fatal("Esc at the menu did not quit")
	}
}

// TestAppKeysCtrlCExitsFromMenuAndView covers AC2: Ctrl-C quits from the menu
// and from a view.
func TestAppKeysCtrlCExitsFromMenuAndView(t *testing.T) {
	t.Run("menu", func(t *testing.T) {
		_, d := newTestApp(t, testAppConfig(t), 80, 24)
		d.key("ctrl+c")
		if !d.quit {
			t.Fatal("Ctrl-C at the menu did not quit")
		}
	})
	t.Run("view", func(t *testing.T) {
		_, d := newTestApp(t, testAppConfig(t), 80, 24)
		d.key("down", "enter") // Update
		d.mustShow("Revision")
		d.key("ctrl+c")
		if !d.quit {
			t.Fatal("Ctrl-C inside a view did not quit")
		}
	})
	t.Run("terminal too small", func(t *testing.T) {
		_, d := newTestApp(t, testAppConfig(t), 60, 10)
		d.key("ctrl+c")
		if !d.quit {
			t.Fatal("Ctrl-C under the size warning did not quit")
		}
	})
}

// TestAppTooSmallWarnsAndKeepsStateAfterResize covers AC9: below 80x24 the
// application shows the warning instead of the view, and growing back
// restores the view with its state.
func TestAppTooSmallWarnsAndKeepsStateAfterResize(t *testing.T) {
	for _, size := range [][2]int{{79, 24}, {80, 23}, {40, 10}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			_, d := newTestApp(t, testAppConfig(t), 80, 24)
			d.key("down", "down", "enter") // Releases
			d.mustShow("No releases are retained yet")
			d.resize(size[0], size[1])
			d.mustShow("Terminal too small: needs 80×24")
			d.mustNotShow("No releases are retained yet")
			for i, line := range d.lines() {
				if w := lipgloss.Width(line); w > size[0] {
					t.Fatalf("warning line %d is %d wide, terminal is %d: %q", i, w, size[0], line)
				}
			}
			if h := len(d.lines()); h > size[1] {
				t.Fatalf("warning is %d lines, terminal is %d", h, size[1])
			}
			d.key("esc", "enter") // ignored while too small
			d.mustShow("Terminal too small")
			d.resize(80, 24)
			d.mustShow("No releases are retained yet")
			d.key("esc")
			d.mustShow("> Releases")
		})
	}
}

// pushTestView pushes v through the root's own message, as an asynchronous
// result would.
func pushTestView(d *appDriver, v view) {
	d.t.Helper()
	d.send(pushViewMsg{v: v})
}

const longLine = "/very/long/path/that/keeps/going/and/going/until/it/is/well/past/the/right/edge/of/a/one/hundred/and/twenty/column/terminal/file.md"

func longSummary(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %02d %s\n", i, longLine)
	}
	return b.String()
}

// screenFitProblem describes a line wider than the terminal or a screen taller
// than it, or returns "". It measures what the root composes, which the root
// has already clipped, so it cannot see a view that draws too much.
func screenFitProblem(d *appDriver, width, height int) string {
	lines := d.lines()
	if len(lines) > height {
		return fmt.Sprintf("screen has %d lines, terminal has %d:\n%s", len(lines), height, d.screen())
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > width {
			return fmt.Sprintf("line %d is %d columns wide, terminal has %d: %q", i, w, width, line)
		}
	}
	return ""
}

// viewFitProblem measures the top view's own View output against the body
// area it is given (the width, and the height minus the status line, the blank
// row and the help bar), before the root clips anything, or returns "".
func viewFitProblem(d *appDriver) string {
	m, ok := d.model.(*appModel)
	if !ok || m.tooSmall() {
		return ""
	}
	width, height := m.bodySize()
	body := m.top().View(viewCtx{Width: width, Height: height, Theme: &m.theme, Busy: m.writing, Spinner: m.spin.View()})
	lines := strings.Split(body, "\n")
	if len(lines) > height {
		return fmt.Sprintf("the view drew %d lines, its body has %d:\n%s", len(lines), height, stripANSI(body))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > width {
			return fmt.Sprintf("the view's line %d is %d columns wide, its body has %d: %q", i, w, width, stripANSI(line))
		}
	}
	return ""
}

// assertFits fails when the screen, or the top view's own output, does not fit
// the terminal (AC9). Checking the view itself matters: the root silently
// clips whatever a view draws beyond its area.
func assertFits(t *testing.T, d *appDriver, width, height int) {
	t.Helper()
	if p := screenFitProblem(d, width, height); p != "" {
		t.Fatal(p)
	}
	if p := viewFitProblem(d); p != "" {
		t.Fatal(p)
	}
}

// oversizedView draws more lines, or wider lines, than its body allows.
type oversizedView struct {
	baseView
	extraLines, extraColumns int
}

func (v oversizedView) Update(tea.Msg) (tea.Cmd, action) { return nil, action{nav: navNone} }
func (v oversizedView) Keys() []key.Binding              { return nil }
func (v oversizedView) View(c viewCtx) string {
	line := strings.Repeat("x", c.Width+v.extraColumns)
	lines := make([]string, c.Height+v.extraLines)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// TestFitAssertionMeasuresTheViewNotTheClippedScreen shows the fit check
// catches a deliberately oversized view: the root clips the screen so the
// screen alone still fits, but the view's own output does not.
func TestFitAssertionMeasuresTheViewNotTheClippedScreen(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    oversizedView
	}{
		{"too many lines", oversizedView{extraLines: 5}},
		{"too wide", oversizedView{extraColumns: 10}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, d := newTestApp(t, testAppConfig(t), 80, 24)
			pushTestView(d, tc.v)
			if p := screenFitProblem(d, 80, 24); p != "" {
				t.Fatalf("the root should clip an oversized view, but the screen does not fit: %s", p)
			}
			if p := viewFitProblem(d); p == "" {
				t.Fatal("the view-level check missed an oversized view")
			}
		})
	}
	t.Run("a well-sized view passes", func(t *testing.T) {
		_, d := newTestApp(t, testAppConfig(t), 80, 24)
		if p := viewFitProblem(d); p != "" {
			t.Fatal(p)
		}
	})
}

// TestAppMenuAndGenericViewsFit covers AC9 for the menu and the generic views
// at 80x24 and 120x40, with long data.
func TestAppMenuAndGenericViewsFit(t *testing.T) {
	sizes := [][2]int{{80, 24}, {120, 40}}
	for _, size := range sizes {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			t.Run("menu", func(t *testing.T) {
				_, d := newTestApp(t, testAppConfig(t), width, height)
				assertFits(t, d, width, height)
			})
			t.Run("a view", func(t *testing.T) {
				_, d := newTestApp(t, testAppConfig(t), width, height)
				d.key("down", "enter") // Update
				d.mustShow("Revision")
				assertFits(t, d, width, height)
			})
			t.Run("summary and confirmation", func(t *testing.T) {
				_, d := newTestApp(t, testAppConfig(t), width, height)
				pushTestView(d, newConfirmView(confirmOptions{Title: "Install codex", Summary: longSummary(80)}))
				d.mustShow("[Apply]", "line 01")
				assertFits(t, d, width, height)
				d.key("pgdown", "pgdown", "pgdown")
				assertFits(t, d, width, height)
			})
			t.Run("notice", func(t *testing.T) {
				_, d := newTestApp(t, testAppConfig(t), width, height)
				pushTestView(d, newNoticeView("Required CLIs", longSummary(60), nil))
				d.mustShow("[Continue]")
				assertFits(t, d, width, height)
			})
			t.Run("recovery", func(t *testing.T) {
				cfg := testAppConfig(t)
				cfg.Options.StateDir = "/state/" + strings.Repeat("deeply/nested/", 12) + "dir"
				cfg.Deps.Pending = func(string) (management.PendingKind, error) { return management.PendingOnboarding, nil }
				_, d := newTestApp(t, cfg, width, height)
				d.mustShow("[Recover]", "Leave it")
				assertFits(t, d, width, height)
			})
		})
	}
}

// TestAppConfirmKeys covers the generic summary and confirmation view: the
// summary scrolls with the arrow and page keys, Left/Right choose, Enter
// accepts, y and n answer directly.
func TestAppConfirmKeys(t *testing.T) {
	newApp := func(t *testing.T, o confirmOptions) (*appDriver, *string) {
		t.Helper()
		result := new(string)
		o.Title = "Install codex"
		if o.Summary == "" {
			o.Summary = longSummary(80)
		}
		o.OnApply = func() (tea.Cmd, action) { *result = "applied"; return nil, action{nav: navPop} }
		o.OnCancel = func() (tea.Cmd, action) { *result = "cancelled"; return nil, action{nav: navPop} }
		_, d := newTestApp(t, testAppConfig(t), 80, 24)
		pushTestView(d, newConfirmView(o))
		return d, result
	}
	t.Run("starts on Apply and Enter applies", func(t *testing.T) {
		d, result := newApp(t, confirmOptions{})
		d.mustShow("[Apply]")
		d.key("enter")
		if *result != "applied" {
			t.Fatalf("result = %q, want applied", *result)
		}
		d.mustShow("> CLIs") // popped back to the menu
	})
	t.Run("left and right choose", func(t *testing.T) {
		d, result := newApp(t, confirmOptions{})
		d.key("right")
		d.mustShow("[Cancel]")
		d.mustNotShow("[Apply]")
		d.key("left")
		d.mustShow("[Apply]")
		d.key("right", "enter")
		if *result != "cancelled" {
			t.Fatalf("result = %q, want cancelled", *result)
		}
	})
	t.Run("y applies and n cancels", func(t *testing.T) {
		d, result := newApp(t, confirmOptions{})
		d.key("y")
		if *result != "applied" {
			t.Fatalf("y: result = %q", *result)
		}
		d, result = newApp(t, confirmOptions{})
		d.key("n")
		if *result != "cancelled" {
			t.Fatalf("n: result = %q", *result)
		}
	})
	t.Run("Esc and Backspace cancel", func(t *testing.T) {
		for _, k := range []string{"esc", "backspace"} {
			d, result := newApp(t, confirmOptions{})
			d.key(k)
			if *result != "cancelled" {
				t.Fatalf("%s: result = %q", k, *result)
			}
		}
	})
	t.Run("start on Cancel disables y", func(t *testing.T) {
		d, result := newApp(t, confirmOptions{StartOnCancel: true, DisableYes: true})
		d.mustShow("[Cancel]")
		d.key("y")
		if *result != "" {
			t.Fatalf("y acted although it is disabled: %q", *result)
		}
		d.mustShow("[Cancel]")
		d.key("enter")
		if *result != "cancelled" {
			t.Fatalf("Enter on Cancel: result = %q", *result)
		}
		d, result = newApp(t, confirmOptions{StartOnCancel: true, DisableYes: true})
		d.key("left", "enter")
		if *result != "applied" {
			t.Fatalf("Left then Enter: result = %q", *result)
		}
	})
	t.Run("summary scrolls", func(t *testing.T) {
		d, _ := newApp(t, confirmOptions{})
		d.mustShow("line 01")
		d.key("down", "down")
		d.mustNotShow("line 01 ")
		d.mustShow("line 03")
		d.key("up", "up")
		d.mustShow("line 01")
		d.key("pgdown")
		d.mustNotShow("line 01 ")
		for range 15 {
			d.key("pgdown")
		}
		d.mustShow("line 80")
		for range 15 {
			d.key("pgup")
		}
		d.mustShow("line 01")
		d.key("end")
		d.mustShow("line 80")
		d.key("home")
		d.mustShow("line 01")
	})
}

// TestAppNoticeView covers the generic notice: Enter and Esc both continue.
func TestAppNoticeView(t *testing.T) {
	for _, k := range []string{"enter", "esc", "backspace"} {
		t.Run(k, func(t *testing.T) {
			_, d := newTestApp(t, testAppConfig(t), 80, 24)
			pushTestView(d, newNoticeView("Heads up", "Something to read.", nil))
			d.mustShow("Something to read.", "[Continue]")
			d.key(k)
			d.mustNotShow("Something to read.")
			d.mustShow("> CLIs")
		})
	}
}

// TestAppNoColorHasNoColorSequencesAndKeepsCursorSymbol covers design.md
// "Tema": under NO_COLOR the output carries no color sequences, and the
// cursor, the checkbox-style symbols and the chosen button are symbols, not
// color.
func TestAppNoColorHasNoColorSequencesAndKeepsCursorSymbol(t *testing.T) {
	build := func(t *testing.T, noColor bool) *appDriver {
		cfg := testAppConfig(t)
		cfg.NoColor = noColor
		_, d := newTestApp(t, cfg, 80, 24)
		return d
	}
	views := func(d *appDriver) map[string]func() {
		return map[string]func(){
			"menu":      func() {},
			"confirm":   func() { pushTestView(d, newConfirmView(confirmOptions{Title: "t", Summary: "s"})) },
			"notice":    func() { pushTestView(d, newNoticeView("t", "text", nil)) },
			"too small": func() { d.resize(40, 10) },
		}
	}
	colored := build(t, false)
	if !colorSGR(colored.raw()) {
		t.Fatalf("the colored theme emits no color sequence, so the NO_COLOR check would prove nothing:\n%q", colored.raw())
	}
	for name := range views(colored) {
		t.Run(name, func(t *testing.T) {
			d := build(t, true)
			views(d)[name]()
			if raw := d.raw(); colorSGR(raw) {
				t.Fatalf("NO_COLOR view emits a color sequence: %q", raw)
			}
		})
	}
	d := build(t, true)
	d.mustShow("> CLIs")
	pushTestView(d, newConfirmView(confirmOptions{Title: "t", Summary: "s"}))
	d.mustShow("[Apply]")
}

// ---------------------------------------------------------------------------
// Entry.
// ---------------------------------------------------------------------------

type capturedStart struct {
	called bool
	o      options
}

// stubInterfaceEntry replaces the terminal check and the app start, so run
// can be driven without a terminal and without a program.
func stubInterfaceEntry(t *testing.T, isTTY bool, home, stateDir string) *capturedStart {
	t.Helper()
	old := interfaceStdio
	t.Cleanup(func() { interfaceStdio = old })
	got := &capturedStart{}
	interfaceStdio.isTTY = func() bool { return isTTY }
	interfaceStdio.home = home
	interfaceStdio.stateDir = stateDir
	interfaceStdio.start = func(o options, _ io.Reader, _ io.Writer) error {
		got.called, got.o = true, o
		return nil
	}
	return got
}

const wantNoTerminalMessage = "hive tui needs a terminal; use the text commands: hive status, install, update, releases, voice, doctor, models, plan/apply to remove hosts, recover"

// TestUsageBareHiveWithoutTerminalKeepsUsageError covers AC1: bare hive
// without a terminal prints today's usage error and never starts the app.
func TestUsageBareHiveWithoutTerminalKeepsUsageError(t *testing.T) {
	got := stubInterfaceEntry(t, false, "", "")
	err := run(nil)
	if err == nil || err.Error() != usageMessage {
		t.Fatalf("got %v, want %q", err, usageMessage)
	}
	if got.called {
		t.Fatal("the application started without a terminal")
	}
}

// TestUsageTuiWithoutTerminalNamesTextCommands covers AC1: hive tui without a
// terminal fails with exactly the message naming the text commands.
func TestUsageTuiWithoutTerminalNamesTextCommands(t *testing.T) {
	got := stubInterfaceEntry(t, false, "", "")
	err := run([]string{"tui"})
	if err == nil || err.Error() != wantNoTerminalMessage {
		t.Fatalf("got %v, want %q", err, wantNoTerminalMessage)
	}
	if got.called {
		t.Fatal("the application started without a terminal")
	}
}

// TestAppEntryBareHiveStartsApp covers AC1: bare hive in a terminal starts the
// application with no options of its own.
func TestAppEntryBareHiveStartsApp(t *testing.T) {
	got := stubInterfaceEntry(t, true, "/h", "/s")
	if err := run(nil); err != nil {
		t.Fatal(err)
	}
	if !got.called || got.o != (options{Home: "/h", StateDir: "/s", Source: "."}) {
		t.Fatalf("start = %+v", got)
	}
}

// TestAppEntryTuiPassesFlags covers AC1: hive tui takes --home, --state-dir
// and --source.
func TestAppEntryTuiPassesFlags(t *testing.T) {
	got := stubInterfaceEntry(t, true, "", "")
	if err := run([]string{"tui", "--home", "/h", "--state-dir", "/s", "--source", "/src"}); err != nil {
		t.Fatal(err)
	}
	if !got.called || got.o != (options{Home: "/h", StateDir: "/s", Source: "/src"}) {
		t.Fatalf("start = %+v", got)
	}
	got = stubInterfaceEntry(t, true, "", "")
	if err := run([]string{"tui"}); err != nil {
		t.Fatal(err)
	}
	if got.o != (options{Source: "."}) {
		t.Fatalf("defaults = %+v", got.o)
	}
	if err := run([]string{"tui", "extra"}); err == nil {
		t.Fatal("positional arguments accepted")
	}
}

// TestAppOptionsCopyDoesNotAlias covers design.md "Operaciones largas": every
// Cmd gets its own copy of the options, Hosts slice included.
func TestAppOptionsCopyDoesNotAlias(t *testing.T) {
	cfg := testAppConfig(t)
	cfg.Options.Hosts = []string{"codex", "claude"}
	m := newAppModel(cfg)
	first, second := m.optionsCopy(), m.optionsCopy()
	if len(first.Hosts) != 2 || len(second.Hosts) != 2 {
		t.Fatalf("copies lost the hosts: %v %v", first.Hosts, second.Hosts)
	}
	first.Hosts[0] = "mutated"
	if third := m.optionsCopy(); third.Hosts[0] != "codex" || second.Hosts[0] != "codex" {
		t.Fatalf("a mutation of one copy reached the model or another copy: %v %v", third.Hosts, second.Hosts)
	}
}

// ---------------------------------------------------------------------------
// Recovery on open.
// ---------------------------------------------------------------------------

// stagePendingOnboarding leaves a real pending onboarding journal in stateDir,
// the same way main_test.go's recover test does.
func stagePendingOnboarding(t *testing.T, home, stateDir, source string) {
	t.Helper()
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source, Hosts: []string{"codex"}}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(providers.Step{ID: "context7-manual", Provider: providers.Context7, Status: providers.Manual, ManualReason: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Onboard(p, []management.ExternalStep{{ID: "context7-manual", Payload: payload}}, stagingFailureRunner{}); err == nil {
		t.Fatal("expected the staged provider step to fail, leaving onboarding pending")
	}
	if kind, err := management.Pending(stateDir); err != nil || kind != management.PendingOnboarding {
		t.Fatalf("expected a pending onboarding: kind=%v err=%v", kind, err)
	}
}

// withoutState returns files minus everything under the state directory,
// whose journals carry random IDs that differ between twin homes.
func withoutState(files map[string][]byte) map[string][]byte {
	out := map[string][]byte{}
	for rel, data := range files {
		if rel == "state" || strings.HasPrefix(rel, "state/") {
			continue
		}
		out[rel] = data
	}
	return out
}

// TestRecoverOffersPendingOnboardingAndMatchesHiveRecover covers "Recuperación
// al abrir": a pending onboarding is offered on open, accepting it recovers,
// and the resulting home and state equal those of `hive recover` on a twin.
func TestRecoverOffersPendingOnboardingAndMatchesHiveRecover(t *testing.T) {
	source := minimalTestSource(t)
	newHome := func() (home, stateDir string) {
		home, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		return home, filepath.Join(home, "state")
	}
	homeA, stateA := newHome()
	homeB, stateB := newHome()
	stagePendingOnboarding(t, homeA, stateA, source)
	stagePendingOnboarding(t, homeB, stateB, source)

	// Twin B: the text command.
	t.Setenv("HOME", homeB)
	commandOutput := captureStdout(t, func() {
		if err := run([]string{"recover", "--state-dir", stateB}); err != nil {
			t.Fatalf("hive recover: %v", err)
		}
	})
	var commandResult struct{ ID, Phase string }
	if err := json.Unmarshal([]byte(commandOutput), &commandResult); err != nil {
		t.Fatalf("hive recover output %q: %v", commandOutput, err)
	}

	// Twin A: the application.
	cfg := appConfig{
		Options:          management.Options{Scope: "user", Home: homeA, StateDir: stateA, Source: source},
		ExplicitStateDir: true,
		Deps:             defaultInstallDependencies(nativeProviderAdapterFactory),
		Dark:             true,
	}
	_, d := newTestApp(t, cfg, 80, 24)
	d.mustShow("Optional onboarding", "[Recover]", "Leave it")
	if len(menuRows(d)) != 0 {
		t.Fatal("the menu is drawn over the recovery offer")
	}
	d.key("enter")
	// The onboarding IDs are random per staged journal, so only the phase is
	// compared with the command's own result.
	d.mustShow("Onboarding recovery: ", fmt.Sprintf(" (%s).", commandResult.Phase), "[Continue]")
	kindA, errA := management.Pending(stateA)
	kindB, errB := management.Pending(stateB)
	if errA != nil || errB != nil || kindA != kindB {
		t.Fatalf("pending state after the recovery differs from the hive recover twin: app=%v (%v) command=%v (%v)", kindA, errA, kindB, errB)
	}
	d.key("enter")
	d.mustShow("> CLIs")

	gotFiles, wantFiles := withoutState(collectFiles(t, homeA)), withoutState(collectFiles(t, homeB))
	for rel, want := range wantFiles {
		if got, ok := gotFiles[rel]; !ok || !bytes.Equal(got, want) {
			t.Errorf("file %s differs from the hive recover twin (present=%v)", rel, ok)
		}
	}
	if len(gotFiles) != len(wantFiles) {
		t.Errorf("file sets differ: %d vs %d", len(gotFiles), len(wantFiles))
	}
	assertStateJSONMatches(t, stateA, homeA, stateB, homeB)
}

// TestRecoverLeaveItKeepsPendingState covers "Leave it": nothing recovers and
// the pending state stays for `hive recover`.
func TestRecoverLeaveItKeepsPendingState(t *testing.T) {
	for _, choice := range [][]string{{"esc"}, {"backspace"}, {"right", "enter"}, {"l"}} {
		t.Run(strings.Join(choice, "+"), func(t *testing.T) {
			cfg := testAppConfig(t)
			recovered := false
			cfg.Deps.Pending = func(string) (management.PendingKind, error) { return management.PendingCore, nil }
			cfg.Deps.RecoverCore = func(string) (string, error) { recovered = true; return "x", nil }
			_, d := newTestApp(t, cfg, 80, 24)
			d.mustShow("An operation was interrupted", "[Recover]", "Leave it")
			d.key(choice...)
			if recovered {
				t.Fatal("Leave it recovered the operation")
			}
			d.mustShow("> CLIs")
			d.mustNotShow("interrupted")
			if d.quit {
				t.Fatal("leaving the offer quit the application")
			}
		})
	}
}

// TestRecoverFailureNamesHiveRecover covers "Recuperar": when the recovery
// fails, the view shows the error and names `hive recover` with the explicit
// state directory, and the pending state is untouched.
func TestRecoverFailureNamesHiveRecover(t *testing.T) {
	cfg := testAppConfig(t)
	pending := filepath.Join(cfg.Options.StateDir, "pending.json")
	if err := os.WriteFile(pending, []byte(`{"id":"core"}`), 0600); err != nil {
		t.Fatal(err)
	}
	_, d := newTestApp(t, cfg, 80, 24)
	d.mustShow("An operation was interrupted", "[Recover]")
	d.key("enter")
	d.mustShow("invalid transaction ID", "hive recover --state-dir")
	d.key("enter")
	d.mustShow("> CLIs")
	if _, err := os.Stat(pending); err != nil {
		t.Fatalf("the failed recovery removed pending.json: %v", err)
	}
}

// TestRecoverBlocksKeysWhileWriting covers AC2/AC8's write rule: while the
// recovery runs, the view shows the progress indicator and Esc, Backspace,
// Enter and Ctrl-C do nothing; releasing the Cmd shows the result.
func TestRecoverBlocksKeysWhileWriting(t *testing.T) {
	cfg := testAppConfig(t)
	done := false
	cfg.Deps.Pending = func(string) (management.PendingKind, error) {
		if done {
			return management.PendingNone, nil
		}
		return management.PendingCore, nil
	}
	cfg.Deps.RecoverCore = func(string) (string, error) { done = true; return "rec-id", nil }
	m, d := newTestApp(t, cfg, 80, 24)
	d.hold = true
	d.key("enter")
	if !m.isWriting() {
		t.Fatal("the model does not mark the recovery as a write")
	}
	d.mustShow("Recovering")
	before := d.screen()
	d.key("esc", "backspace", "enter", "ctrl+c", "left", "right", "y")
	if d.quit {
		t.Fatal("Ctrl-C quit during a write")
	}
	if after := d.screen(); after != before {
		t.Fatalf("keys changed the screen during a write:\n%s\n--- was ---\n%s", after, before)
	}
	if done {
		t.Fatal("the recovery ran before it was released")
	}
	d.hold = false
	d.release()
	if m.isWriting() {
		t.Fatal("the model still marks a write after it finished")
	}
	d.mustShow("Recovery: rec-id.", "[Continue]")
	d.mustNotShow("Recovering")
	d.key("enter")
	d.mustShow("> CLIs")
	d.key("ctrl+c")
	if !d.quit {
		t.Fatal("Ctrl-C does not quit after the write")
	}
}

// TestRecoverPendingCheckErrorShowsNotice covers a failing pending check: the
// application stays open and shows the error.
func TestRecoverPendingCheckErrorShowsNotice(t *testing.T) {
	cfg := testAppConfig(t)
	cfg.Deps.Pending = func(string) (management.PendingKind, error) {
		return management.PendingNone, errors.New("state unreadable")
	}
	_, d := newTestApp(t, cfg, 80, 24)
	d.mustShow("state unreadable", "[Continue]")
	d.key("enter")
	d.mustShow("> CLIs")
}

// ---------------------------------------------------------------------------
// The write filter and the real program.
// ---------------------------------------------------------------------------

// TestAppFilterDropsInterruptAndQuitDuringWrite covers design.md "Aplicación":
// while the root marks a write in progress, the program filter drops
// InterruptMsg and QuitMsg; otherwise it lets everything through.
func TestAppFilterDropsInterruptAndQuitDuringWrite(t *testing.T) {
	m := newAppModel(testAppConfig(t))
	for _, msg := range []tea.Msg{tea.InterruptMsg{}, tea.QuitMsg{}} {
		if got := appFilter(m, msg); got == nil {
			t.Fatalf("%T dropped while no write is running", msg)
		}
	}
	m.writing = true
	for _, msg := range []tea.Msg{tea.InterruptMsg{}, tea.QuitMsg{}} {
		if got := appFilter(m, msg); got != nil {
			t.Fatalf("%T passed the filter during a write: %v", msg, got)
		}
	}
	if got := appFilter(m, tea.KeyPressMsg{Code: 'a', Text: "a"}); got == nil {
		t.Fatal("the filter dropped an ordinary message during a write")
	}
}

// syncBuffer is a bytes.Buffer safe for the program's goroutines and the test.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// waitFor polls until cond holds or the timeout passes.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// startProgram runs the real Bubble Tea program on a pipe and a buffer sized
// 80x24, without a terminal.
func startProgram(t *testing.T, cfg appConfig) (write func(string), out *syncBuffer, p *tea.Program, done <-chan error) {
	t.Helper()
	pr, pw := io.Pipe()
	out = &syncBuffer{}
	p = newAppProgram(cfg, pr, out, tea.WithWindowSize(80, 24))
	errc := make(chan error, 1)
	go func() {
		_, err := p.Run()
		errc <- err
	}()
	t.Cleanup(func() {
		p.Kill()
		pw.Close()
	})
	return func(s string) {
		if _, err := pw.Write([]byte(s)); err != nil {
			t.Errorf("writing to the program input: %v", err)
		}
	}, out, p, errc
}

func waitExit(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("the program did not exit")
		return nil
	}
}

// TestAppAltScreenOpensAndCloses covers AC1: the program enters the alternate
// screen on start and leaves it on exit, and Ctrl-C ends it with success.
func TestAppAltScreenOpensAndCloses(t *testing.T) {
	const enter, leave = "\x1b[?1049h", "\x1b[?1049l"
	write, out, _, done := startProgram(t, testAppConfig(t))
	waitFor(t, "the alternate screen to open", func() bool { return strings.Contains(out.String(), enter) })
	waitFor(t, "the menu to draw", func() bool { return strings.Contains(stripANSI(out.String()), "Quit") })
	write("\x03")
	if err := waitExit(t, done); err != nil {
		t.Fatalf("Ctrl-C ended the program with %v, want nil (exit 0)", err)
	}
	s := out.String()
	if !strings.Contains(s, leave) {
		t.Fatalf("the program never left the alternate screen: %q", s)
	}
	if strings.LastIndex(s, leave) < strings.Index(s, enter) {
		t.Fatalf("the alternate screen was left before it was entered: %q", s)
	}
}

// TestAppAltScreenNoColorOutputHasNoColor covers design.md "Tema": with
// NO_COLOR the bytes the real program writes hold no color sequence.
func TestAppAltScreenNoColorOutputHasNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cfg := testAppConfig(t)
	cfg.NoColor = true
	write, out, _, done := startProgram(t, cfg)
	waitFor(t, "the menu to draw", func() bool { return strings.Contains(stripANSI(out.String()), "Quit") })
	write("\x03")
	if err := waitExit(t, done); err != nil {
		t.Fatal(err)
	}
	if colorSGR(out.String()) {
		t.Fatalf("NO_COLOR output holds a color sequence: %q", out.String())
	}
	if !strings.Contains(stripANSI(out.String()), "> CLIs") {
		t.Fatalf("the cursor is not the > symbol: %q", stripANSI(out.String()))
	}
}

// TestAppRealProgramIgnoresInterruptDuringWrite covers design.md "Pruebas": a
// SIGINT-style InterruptMsg and a Ctrl-C key during a write are ignored, and
// the write finishes; with -race it covers the real Cmd goroutine.
func TestAppRealProgramIgnoresInterruptDuringWrite(t *testing.T) {
	cfg := testAppConfig(t)
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	t.Cleanup(release)
	var mu sync.Mutex
	done := false
	cfg.Deps.Pending = func(string) (management.PendingKind, error) {
		mu.Lock()
		defer mu.Unlock()
		if done {
			return management.PendingNone, nil
		}
		return management.PendingCore, nil
	}
	cfg.Deps.RecoverCore = func(string) (string, error) {
		<-gate
		mu.Lock()
		done = true
		mu.Unlock()
		return "rec-id", nil
	}
	write, out, p, exited := startProgram(t, cfg)
	waitFor(t, "the recovery offer", func() bool { return strings.Contains(stripANSI(out.String()), "Leave it") })
	write("\r")
	waitFor(t, "the recovery to start", func() bool { return strings.Contains(stripANSI(out.String()), "Recovering") })
	p.Send(tea.InterruptMsg{})
	write("\x03")
	select {
	case err := <-exited:
		t.Fatalf("the program exited during a write: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	release()
	waitFor(t, "the recovery result", func() bool { return strings.Contains(stripANSI(out.String()), "Recovery: rec-id.") })
	write("\x03")
	if err := waitExit(t, exited); err != nil {
		t.Fatalf("Ctrl-C after the write ended the program with %v", err)
	}
}

// TestAppTooSmallWhileWritingDoesNotAdviseCtrlC covers the size warning during
// a write: Ctrl-C is ignored until the write finishes, so the warning must not
// tell the user to press it; outside a write it still does.
func TestAppTooSmallWhileWritingDoesNotAdviseCtrlC(t *testing.T) {
	cfg := testAppConfig(t)
	cfg.Deps.Pending = func(string) (management.PendingKind, error) { return management.PendingCore, nil }
	cfg.Deps.RecoverCore = func(string) (string, error) { return "rec-id", nil }
	m, d := newTestApp(t, cfg, 80, 24)
	d.resize(60, 10)
	mustShowFlat(d, "Terminal too small")
	mustShowFlat(d, "press ctrl+c to quit")
	d.resize(80, 24)
	d.hold = true
	d.key("enter")
	if !m.isWriting() {
		t.Fatal("the recovery is not marked as a write")
	}
	d.resize(60, 10)
	mustShowFlat(d, "Terminal too small")
	if strings.Contains(strings.Join(strings.Fields(d.screen()), ""), "pressctrl+c") {
		t.Fatalf("the warning offers Ctrl-C during a write:\n%s", d.screen())
	}
	mustShowFlat(d, "ignored until the operation finishes")
	d.key("ctrl+c")
	if d.quit {
		t.Fatal("Ctrl-C quit during a write")
	}
	d.hold = false
	d.release()
	mustShowFlat(d, "press ctrl+c to quit") // the write finished: Ctrl-C works again
}

// The recovery result line names `hive recover` only when something is still
// pending. These call recoverPending, the function the recovery view's Cmd
// runs, with the doubles the install tests use.

func recoverPendingOptions(t *testing.T) management.Options {
	t.Helper()
	o := interfaceTestOptions(t)
	return management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir}
}

// TestRecoverPendingCoreSuccessOmitsRecoverPhrase: a core recovery that
// succeeded must not also tell the operator to run `hive recover`.
func TestRecoverPendingCoreSuccessOmitsRecoverPhrase(t *testing.T) {
	deps := defaultInstallDependencies(coreOnlyAdapterFactory)
	recovered := false
	deps.RecoverCore = func(string) (string, error) {
		recovered = true
		return "recovered-id", nil
	}
	text, err := recoverPending(management.PendingCore, recoverPendingOptions(t), true, deps)
	if err != nil {
		t.Fatal(err)
	}
	if !recovered {
		t.Fatal("RecoverCore was not called")
	}
	if !strings.Contains(text, "Recovery: recovered-id.") {
		t.Fatalf("missing the recovery success message: %q", text)
	}
	if strings.Contains(text, "hive recover") || strings.Contains(text, "install.sh") {
		t.Fatalf("a recovery that already succeeded must not also tell the operator to recover: %q", text)
	}
}

// TestRecoverPendingOnboardingPartialNamesRecover: an onboarding recovery can
// finish "partial", and then the line names `Run hive recover`, with
// --state-dir when the app was opened against an explicit one.
func TestRecoverPendingOnboardingPartialNamesRecover(t *testing.T) {
	o := recoverPendingOptions(t)
	deps := defaultInstallDependencies(coreOnlyAdapterFactory)
	deps.RecoverOnboarding = func(string, onboardingAdapter) (management.OnboardingResult, error) {
		return management.OnboardingResult{ID: "onboarding-id", Phase: "partial"}, nil
	}
	text, err := recoverPending(management.PendingOnboarding, o, false, deps)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Onboarding recovery: onboarding-id (partial). Run hive recover.") {
		t.Fatalf("a still-partial onboarding recovery must name Run hive recover: %q", text)
	}
	text, err = recoverPending(management.PendingOnboarding, o, true, deps)
	if err != nil {
		t.Fatal(err)
	}
	if want := "Run hive recover --state-dir " + o.StateDir + "."; !strings.Contains(text, want) {
		t.Fatalf("an explicit state directory must be named: got %q, want it to contain %q", text, want)
	}
}

// TestRecoverPendingOnboardingCompletedOmitsRecoverPhrase: phase "completed"
// means nothing more is pending, so the phrase must not appear.
func TestRecoverPendingOnboardingCompletedOmitsRecoverPhrase(t *testing.T) {
	deps := defaultInstallDependencies(coreOnlyAdapterFactory)
	deps.RecoverOnboarding = func(string, onboardingAdapter) (management.OnboardingResult, error) {
		return management.OnboardingResult{ID: "onboarding-id", Phase: "completed"}, nil
	}
	text, err := recoverPending(management.PendingOnboarding, recoverPendingOptions(t), true, deps)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Onboarding recovery: onboarding-id (completed).") {
		t.Fatalf("missing the recovery result: %q", text)
	}
	if strings.Contains(text, "hive recover") {
		t.Fatalf("a completed onboarding recovery must not also tell the operator to recover: %q", text)
	}
}

// TestAppRealProgramSurvivesRepeatedSignalsDuringWrite covers AC2 with the
// real signal path (runAppWith): every SIGINT and SIGTERM sent during a write
// is ignored, however many arrive; after the write, a signal at idle quits
// cleanly, with the alternate screen left and no error. The signals go to the
// test process itself, which is safe while runAppWith has them registered.
func TestAppRealProgramSurvivesRepeatedSignalsDuringWrite(t *testing.T) {
	cfg := testAppConfig(t)
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	var mu sync.Mutex
	recovered := false
	cfg.Deps.Pending = func(string) (management.PendingKind, error) {
		mu.Lock()
		defer mu.Unlock()
		if recovered {
			return management.PendingNone, nil
		}
		return management.PendingCore, nil
	}
	cfg.Deps.RecoverCore = func(string) (string, error) {
		<-gate
		mu.Lock()
		recovered = true
		mu.Unlock()
		return "rec-id", nil
	}
	pr, pw := io.Pipe()
	out := &syncBuffer{}
	done := make(chan error, 1)
	go func() { done <- runAppWith(cfg, pr, out, tea.WithWindowSize(80, 24)) }()
	t.Cleanup(func() {
		release()
		pw.Close()
	})
	seen := func(s string) func() bool {
		return func() bool { return strings.Contains(stripANSI(out.String()), s) }
	}
	waitFor(t, "the recovery offer", seen("Leave it"))
	if _, err := pw.Write([]byte("\r")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the recovery to start", seen("Recovering"))
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGINT} {
		if err := syscall.Kill(os.Getpid(), sig); err != nil {
			t.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	select {
	case err := <-done:
		t.Fatalf("the app exited during a write: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	release()
	waitFor(t, "the recovery result", seen("Recovery: rec-id."))
	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("a signal at idle ended the app with %v, want a clean exit", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the app did not exit on a signal at idle")
	}
	if s := out.String(); !strings.Contains(s, "\x1b[?1049l") {
		t.Fatalf("the alternate screen was not left: %q", s)
	}
}

// TestAppRealProgramSigtermAtIdleQuitsCleanly covers the idle SIGTERM.
func TestAppRealProgramSigtermAtIdleQuitsCleanly(t *testing.T) {
	pr, pw := io.Pipe()
	out := &syncBuffer{}
	done := make(chan error, 1)
	go func() { done <- runAppWith(testAppConfig(t), pr, out, tea.WithWindowSize(80, 24)) }()
	t.Cleanup(func() { pw.Close() })
	waitFor(t, "the menu", func() bool { return strings.Contains(stripANSI(out.String()), "Quit") })
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("SIGTERM ended the app with %v, want a clean exit", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the app did not exit on SIGTERM")
	}
	if !strings.Contains(out.String(), "\x1b[?1049l") {
		t.Fatal("the alternate screen was not left")
	}
}

// TestAppConfigNamesTheStateDirWhenItDiffersFromTheDefault covers the recovery
// phrases: a bare `hive recover` targets the real user's state, so the app
// names the state directory whenever --state-dir or --home moved it, not only
// with --state-dir.
func TestAppConfigNamesTheStateDirWhenItDiffersFromTheDefault(t *testing.T) {
	cases := []struct {
		name     string
		o        options
		explicit bool
	}{
		{"nothing given", options{Source: "."}, false},
		{"--state-dir", options{StateDir: filepath.Join(t.TempDir(), "state"), Source: "."}, true},
		{"--home alone moves the state", options{Home: t.TempDir(), Source: "."}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := newAppConfig(tc.o)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.ExplicitStateDir != tc.explicit {
				t.Fatalf("ExplicitStateDir = %v, want %v", cfg.ExplicitStateDir, tc.explicit)
			}
		})
	}
	// End to end: the recovery offer of a --home-only run names its state.
	cfg, err := newAppConfig(options{Home: t.TempDir(), Source: "."})
	if err != nil {
		t.Fatal(err)
	}
	cfg.Deps.Pending = func(string) (management.PendingKind, error) { return management.PendingCore, nil }
	_, d := newTestApp(t, cfg, 80, 24)
	mustShowFlat(d, "Run hive recover --state-dir "+shellQuote(cfg.Options.StateDir))
}

// TestRecoveryPhrasesQuoteThePathForTheShell covers the phrases the app builds:
// a state directory with spaces (macOS "Application Support") or a quote must
// paste into a shell as one argument; a plain path stays as it is.
func TestRecoveryPhrasesQuoteThePathForTheShell(t *testing.T) {
	cases := []struct{ path, quoted string }{
		{"/tmp/plain-state_1", "/tmp/plain-state_1"},
		{"/Users/x/Library/Application Support/tricell-hive", "'/Users/x/Library/Application Support/tricell-hive'"},
		{"/tmp/it's here", `'/tmp/it'\''s here'`},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got, want := recoveryTextOnOpen(true)(tc.path, true), "Run hive recover --state-dir "+tc.quoted; got != want {
				t.Errorf("recoveryTextOnOpen = %q, want %q", got, want)
			}
			if got, want := recoveryPhraseFor(true, true, false, tc.path), "run hive recover --state-dir "+tc.quoted; got != want {
				t.Errorf("recoveryPhraseFor = %q, want %q", got, want)
			}
		})
	}
}
