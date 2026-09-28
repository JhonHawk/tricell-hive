// tui_app.go is the full-screen application's root model (design.md
// "Aplicación"): one Bubble Tea program with a stack of views on the alternate
// screen. The root owns the status line, the help bar, the global keys, the
// minimum-size warning and the write flag; views own their own content and
// keys. Long operations run as tea.Cmd values that return result messages, so
// the model only changes inside Update.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"tricell-hive/tooling/management"
)

// Below this size the application shows a warning instead of the view
// (design.md "Tamaño mínimo").
const (
	minWidth  = 80
	minHeight = 24
)

// The fixed rows around a view: the status line, one blank row, and the help
// bar. A view gets the rest of the screen.
const chromeRows = 3

// navKind is what a view asks the root to do after handling a message.
type navKind int

const (
	// navDefault means the view did not consume the key: the root applies the
	// global meaning of Esc and Backspace, and ignores any other key.
	navDefault navKind = iota
	navNone
	navPush
	navPop
	navQuit
)

// action is a view's reply to a message.
//
// pops removes that many views from the top before push is added: with
// navPop it defaults to one, with navPush to none, so a view can replace
// itself (navPush with pops 1) or unwind several views at once.
//
// write marks the accompanying Cmd as a write: the root then ignores every
// key, and drops interrupts, until the Cmd's result arrives. The result goes
// to result when set (a view lower in the stack that owns the flow), and to
// the top view otherwise.
type action struct {
	nav    navKind
	push   view
	pops   int
	write  bool
	result view
}

// targetedMsg is implemented by a message that must reach the view that issued
// it, even when another view was pushed on top of that view in the meantime.
type targetedMsg interface {
	target() view
}

// revealer is implemented by a view that reacts to becoming the top again
// after the views above it were popped (the CLIs view reloads its rows).
type revealer interface {
	Reveal() tea.Cmd
}

// spinnerNeeder is implemented by a view that is loading or planning and
// wants the spinner frames to keep ticking.
type spinnerNeeder interface {
	NeedsSpinner() bool
}

// viewCtx is what a view needs to draw: its area, the theme, and whether a
// write is in progress with the current spinner frame.
type viewCtx struct {
	Width, Height int
	Theme         *appTheme
	Busy          bool
	Spinner       string
}

// view is one screen of the stack. Update handles keys the root passed on and
// the result messages of the view's own Cmds; it never runs work itself.
type view interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Cmd, action)
	View(c viewCtx) string
	// Resize gives the view the area it draws in (below the status line and
	// above the help bar).
	Resize(width, height int)
	// Keys lists the view's own keys for the help bar.
	Keys() []key.Binding
	// TextFocused reports whether a text field has the focus, which makes
	// Backspace edit instead of going back.
	TextFocused() bool
}

// Messages the root handles itself.
type (
	// pushViewMsg pushes a view, as an asynchronous result would.
	pushViewMsg struct{ v view }
	// statusLoadedMsg carries the status line the header shows. seq lets a
	// stale load be ignored.
	statusLoadedMsg struct {
		seq  int
		line string
		err  error
	}
	pendingCheckedMsg struct {
		kind management.PendingKind
		err  error
	}
	// writeDoneMsg wraps the result of a write Cmd; the root clears the write
	// flag and hands the inner message to the top view.
	writeDoneMsg struct{ msg tea.Msg }
)

// appConfig is everything the application needs; tests replace Deps.
type appConfig struct {
	// Options are normalized: StateDir is a concrete path.
	Options management.Options
	// ExplicitStateDir tells recovery phrases to name --state-dir.
	ExplicitStateDir bool
	Deps             installDependencies
	// Dark selects the dark palette; NoColor removes color entirely.
	Dark, NoColor bool
}

// copyOptions returns o with its own Hosts slice, so a Cmd never shares memory
// with the model.
func copyOptions(o management.Options) management.Options {
	o.Hosts = append([]string(nil), o.Hosts...)
	return o
}

type appModel struct {
	cfg           appConfig
	theme         appTheme
	width, height int
	stack         []view
	spin          spinner.Model
	spinning      bool
	help          help.Model
	status        string
	statusErr     string
	statusLoading bool
	statusSeq     int
	writing       bool
	// writeTarget receives the running write's result; nil means the top view.
	writeTarget view
}

var quitBinding = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))

func newAppModel(cfg appConfig) *appModel {
	theme := newAppTheme(cfg.Dark, cfg.NoColor)
	h := help.New()
	h.Styles = theme.Help
	m := &appModel{
		cfg:           cfg,
		theme:         theme,
		spin:          spinner.New(spinner.WithSpinner(spinner.Line), spinner.WithStyle(theme.Accent)),
		help:          h,
		statusLoading: true,
	}
	m.stack = []view{newMenuView(cfg)}
	return m
}

// optionsCopy is the Options every Cmd receives (design.md "Operaciones
// largas").
func (m *appModel) optionsCopy() management.Options { return copyOptions(m.cfg.Options) }

func (m *appModel) isWriting() bool { return m.writing }

func (m *appModel) top() view { return m.stack[len(m.stack)-1] }

func (m *appModel) tooSmall() bool {
	return m.width < minWidth || m.height < minHeight
}

func (m *appModel) needsSpinner() bool {
	if m.statusLoading || m.writing {
		return true
	}
	n, ok := m.top().(spinnerNeeder)
	return ok && n.NeedsSpinner()
}

// startSpinner starts the frame ticks unless they already run; the chain stops
// itself once nothing needs the spinner.
func (m *appModel) startSpinner() tea.Cmd {
	if m.spinning || !m.needsSpinner() {
		return nil
	}
	m.spinning = true
	return m.spin.Tick
}

func (m *appModel) loadStatus() tea.Cmd {
	m.statusSeq++
	m.statusLoading = true
	seq, o := m.statusSeq, m.optionsCopy()
	load := func() tea.Msg {
		line, err := interfaceStatusLine(o)
		return statusLoadedMsg{seq: seq, line: line, err: err}
	}
	return tea.Batch(load, m.startSpinner())
}

func (m *appModel) checkPending() tea.Cmd {
	deps, stateDir := m.cfg.Deps, m.cfg.Options.StateDir
	return func() tea.Msg {
		kind, err := deps.Pending(stateDir)
		return pendingCheckedMsg{kind: kind, err: err}
	}
}

func (m *appModel) Init() tea.Cmd {
	return tea.Batch(m.loadStatus(), m.checkPending())
}

func (m *appModel) bodySize() (int, int) { return m.width, max(m.height-chromeRows, 0) }

func (m *appModel) resizeViews() {
	w, h := m.bodySize()
	for _, v := range m.stack {
		v.Resize(w, h)
	}
}

func (m *appModel) push(v view) tea.Cmd {
	w, h := m.bodySize()
	v.Resize(w, h)
	m.stack = append(m.stack, v)
	return v.Init()
}

// apply carries out a view's reply and returns the Cmds it produces.
func (m *appModel) apply(cmd tea.Cmd, act action) tea.Cmd {
	var cmds []tea.Cmd
	if act.nav == navQuit {
		return tea.Quit
	}
	pops := act.pops
	if act.nav == navPop {
		pops = max(pops, 1)
	}
	popped := 0
	for ; popped < pops && len(m.stack) > 1; popped++ {
		m.stack = m.stack[:len(m.stack)-1]
	}
	pushed := false
	if act.nav == navPush && act.push != nil {
		cmds = append(cmds, m.push(act.push))
		pushed = true
	}
	if popped > 0 && !pushed {
		if r, ok := m.top().(revealer); ok {
			cmds = append(cmds, r.Reveal())
		}
	}
	if cmd != nil {
		if act.write {
			m.writing = true
			m.writeTarget = act.result
			cmd = wrapWrite(cmd)
		}
		cmds = append(cmds, cmd)
	}
	cmds = append(cmds, m.startSpinner())
	return tea.Batch(cmds...)
}

// wrapWrite tags a write Cmd's result so the root can tell when the write
// finished.
func wrapWrite(cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg { return writeDoneMsg{msg: cmd()} }
}

func (m *appModel) inStack(v view) bool {
	for _, s := range m.stack {
		if s == v {
			return true
		}
	}
	return false
}

// forward hands a message to the top view.
func (m *appModel) forward(msg tea.Msg) tea.Cmd {
	cmd, act := m.top().Update(msg)
	return m.apply(cmd, act)
}

func (m *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeViews()
		return m, nil
	case spinner.TickMsg:
		if !m.needsSpinner() {
			m.spinning = false
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case statusLoadedMsg:
		if msg.seq != m.statusSeq {
			return m, nil
		}
		m.statusLoading = false
		m.status, m.statusErr = msg.line, ""
		if msg.err != nil {
			m.status, m.statusErr = "", msg.err.Error()
		}
		return m, nil
	case pendingCheckedMsg:
		return m, m.onPendingChecked(msg)
	case pushViewMsg:
		return m, m.apply(nil, action{nav: navPush, push: msg.v})
	case writeDoneMsg:
		m.writing = false
		target := m.writeTarget
		m.writeTarget = nil
		var cmd tea.Cmd
		if target != nil && m.inStack(target) {
			c, act := target.Update(msg.msg)
			cmd = m.apply(c, act)
		} else {
			cmd = m.forward(msg.msg)
		}
		return m, tea.Batch(cmd, m.loadStatus())
	case tea.KeyPressMsg:
		return m, m.handleKey(msg)
	}
	if t, ok := msg.(targetedMsg); ok {
		if tv := t.target(); tv != nil {
			if !m.inStack(tv) {
				return m, nil // the view was popped: its late result is dropped
			}
			cmd, act := tv.Update(msg)
			return m, m.apply(cmd, act)
		}
	}
	return m, m.forward(msg)
}

func (m *appModel) onPendingChecked(msg pendingCheckedMsg) tea.Cmd {
	switch {
	case msg.err != nil:
		text := "The interrupted-operation check failed: " + msg.err.Error()
		return m.apply(nil, action{nav: navPush, push: newNoticeView("Cannot read the state", text, nil)})
	case msg.kind == management.PendingNone:
		return nil
	}
	rv := newRecoveryView(msg.kind, m.optionsCopy(), m.cfg.ExplicitStateDir, m.cfg.Deps)
	return m.apply(nil, action{nav: navPush, push: rv})
}

// handleKey applies the global keys (design.md "Teclas globales"): every key
// is ignored while a write runs; Ctrl-C quits; a view sees each other key
// first and the root applies the default for Esc and Backspace.
func (m *appModel) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	if m.writing {
		return nil
	}
	if msg.String() == "ctrl+c" {
		return tea.Quit
	}
	if m.tooSmall() {
		return nil
	}
	top := m.top()
	cmd, act := top.Update(msg)
	if act.nav == navDefault {
		act = m.defaultAction(msg.String(), top)
	}
	return m.apply(cmd, act)
}

// defaultAction is the global meaning of a key the view did not consume: Esc
// goes back, or quits at the menu; Backspace goes back unless a text field has
// the focus, and does nothing at the menu.
func (m *appModel) defaultAction(k string, top view) action {
	switch k {
	case "esc":
		if len(m.stack) > 1 {
			return action{nav: navPop}
		}
		return action{nav: navQuit}
	case "backspace":
		if len(m.stack) > 1 && !top.TextFocused() {
			return action{nav: navPop}
		}
	}
	return action{nav: navNone}
}

// clipLine shortens a possibly styled line to w columns, ending with an
// ellipsis when it cut something.
func clipLine(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	return ansi.Truncate(s, w, "…")
}

func (m *appModel) headerLine() string {
	var rest string
	switch {
	case m.statusLoading:
		rest = m.spin.View() + " " + m.theme.Muted.Render("Loading status…")
	case m.statusErr != "":
		rest = m.theme.Danger.Render("Status unavailable: " + m.statusErr)
	default:
		rest = m.theme.Muted.Render(m.status)
	}
	return clipLine(m.theme.Title.Render("Hive")+" "+m.theme.Muted.Render("·")+" "+rest, m.width)
}

func (m *appModel) helpLine() string {
	if m.writing {
		return clipLine(m.spin.View()+" "+m.theme.Muted.Render("Working… keys are ignored until it finishes"), m.width)
	}
	bindings := append(append([]key.Binding(nil), m.top().Keys()...), quitBinding)
	m.help.SetWidth(m.width)
	return clipLine(m.help.ShortHelpView(bindings), m.width)
}

func (m *appModel) sizeWarning() string {
	advice := "Enlarge the window, or press ctrl+c to quit."
	if m.writing {
		// Ctrl-C is ignored until the write finishes, so do not offer it.
		advice = "Enlarge the window. Keys, ctrl+c included, are ignored until the operation finishes."
	}
	var lines []string
	for _, l := range strings.Split(ansi.Wrap(fmt.Sprintf("Terminal too small: needs %d×%d", minWidth, minHeight), max(m.width, 1), ""), "\n") {
		lines = append(lines, m.theme.Danger.Render(l))
	}
	for _, l := range strings.Split(ansi.Wrap(fmt.Sprintf("Current size: %d×%d. %s", m.width, m.height, advice), max(m.width, 1), ""), "\n") {
		lines = append(lines, m.theme.Muted.Render(l))
	}
	return strings.Join(lines[:min(len(lines), max(m.height, 1))], "\n")
}

func (m *appModel) render() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if m.tooSmall() {
		return m.sizeWarning()
	}
	w, bodyH := m.bodySize()
	body := m.top().View(viewCtx{Width: w, Height: bodyH, Theme: &m.theme, Busy: m.writing, Spinner: m.spin.View()})
	lines := strings.Split(body, "\n")
	if len(lines) > bodyH {
		lines = lines[:bodyH]
	}
	for i, l := range lines {
		lines[i] = clipLine(l, w)
	}
	for len(lines) < bodyH {
		lines = append(lines, "")
	}
	all := append([]string{m.headerLine(), ""}, lines...)
	all = append(all, m.helpLine())
	return strings.Join(all, "\n")
}

func (m *appModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

// appFilter is the program's message filter (design.md "Aplicación"): while
// the root marks a write in progress it drops InterruptMsg (an external
// SIGINT) and QuitMsg, so nothing cuts an apply short. A SIGKILL still can;
// the operation stays pending and is offered for recovery on the next open.
func appFilter(m tea.Model, msg tea.Msg) tea.Msg {
	if am, ok := m.(*appModel); ok && am.writing {
		switch msg.(type) {
		case tea.InterruptMsg, tea.QuitMsg:
			return nil
		}
	}
	return msg
}

// newAppProgram builds the program without running it. Tests add options such
// as tea.WithWindowSize. Bubble Tea's own signal handler is off: it stops after
// forwarding the first signal, and the process would then take the next one
// with the default action, ending an apply half done with the terminal still
// on the alternate screen. runAppWith forwards every signal instead.
func newAppProgram(cfg appConfig, in io.Reader, out io.Writer, extra ...tea.ProgramOption) *tea.Program {
	opts := append([]tea.ProgramOption{tea.WithInput(in), tea.WithOutput(out), tea.WithFilter(appFilter), tea.WithoutSignalHandler()}, extra...)
	return tea.NewProgram(newAppModel(cfg), opts...)
}

// forwardSignals turns each SIGINT into an InterruptMsg and each other signal
// into a QuitMsg for the whole life of the program, until stop closes. While a
// write runs, appFilter drops both messages, however many arrive.
func forwardSignals(p *tea.Program, sigs <-chan os.Signal, stop <-chan struct{}) {
	for {
		select {
		case s := <-sigs:
			if s == syscall.SIGINT {
				p.Send(tea.InterruptMsg{})
			} else {
				p.Send(tea.QuitMsg{})
			}
		case <-stop:
			return
		}
	}
}

// runAppWith runs the application until the user quits. Ctrl-C, and an
// external SIGINT or SIGTERM outside a write, end it successfully with the
// terminal restored; during a write every signal is ignored.
func runAppWith(cfg appConfig, in io.Reader, out io.Writer, extra ...tea.ProgramOption) error {
	p := newAppProgram(cfg, in, out, extra...)
	sigs := make(chan os.Signal, 8)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	stop := make(chan struct{})
	go forwardSignals(p, sigs, stop)
	_, err := p.Run()
	signal.Stop(sigs)
	close(stop)
	if errors.Is(err, tea.ErrInterrupted) {
		return nil
	}
	return err
}
