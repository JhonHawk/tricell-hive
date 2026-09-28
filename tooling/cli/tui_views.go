// tui_views.go holds the application's views: the menu, the provisional
// "Not implemented yet" view that T8 and T9 replace, and the generic views the
// later screens stack on top of their source view: summary with confirmation,
// notice, and recovery on open (design.md "Vistas").
package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"tricell-hive/tooling/management"
)

func binding(keys, helpKey, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(strings.Split(keys, ",")...), key.WithHelp(helpKey, desc))
}

// baseView supplies the parts most views share: nothing to load, no text
// field, and no size-dependent state.
type baseView struct{}

func (baseView) Init() tea.Cmd     { return nil }
func (baseView) Resize(int, int)   {}
func (baseView) TextFocused() bool { return false }

// ---------------------------------------------------------------------------
// Menu.
// ---------------------------------------------------------------------------

// menuItem is one row of the main menu. A nil open means Quit.
type menuItem struct {
	label, desc string
	open        func() view
}

// mainMenuItems is the menu's fixed set, in order (design.md "Menú (D2-A)").
// Every entry but Quit opens a provisional view until T8 and T9 replace it.
var mainMenuItems = []menuItem{
	{"CLIs", "Install, remove and check CLI hosts", func() view { return newPlaceholderView("CLIs") }},
	{"Update", "Update Hive from a Git commit", func() view { return newPlaceholderView("Update") }},
	{"Releases", "Go back to a retained release", func() view { return newPlaceholderView("Releases") }},
	{"Voice", "Choose the assistant voice", func() view { return newPlaceholderView("Voice") }},
	{"Quit", "Leave Hive", nil},
}

type menuView struct {
	baseView
	cursor int
}

func newMenuView() *menuView { return &menuView{} }

func (v *menuView) Update(msg tea.Msg) (tea.Cmd, action) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, action{nav: navNone}
	}
	switch k.String() {
	case "esc", "backspace":
		return nil, action{} // the root decides: Esc quits, Backspace does nothing
	case "up":
		v.cursor = max(v.cursor-1, 0)
	case "down":
		v.cursor = min(v.cursor+1, len(mainMenuItems)-1)
	case "enter":
		item := mainMenuItems[v.cursor]
		if item.open == nil {
			return nil, action{nav: navQuit}
		}
		return nil, action{nav: navPush, push: item.open()}
	}
	return nil, action{nav: navNone}
}

func (v *menuView) View(c viewCtx) string {
	th := c.Theme
	lines := []string{th.Title.Render("Main menu"), ""}
	labelWidth := 0
	for _, item := range mainMenuItems {
		labelWidth = max(labelWidth, len(item.label))
	}
	for i, item := range mainMenuItems {
		label := fmt.Sprintf("%-*s", labelWidth, item.label)
		if i == v.cursor {
			lines = append(lines, th.Accent.Render("> "+label)+"  "+th.Muted.Render(item.desc))
			continue
		}
		lines = append(lines, th.Text.Render("  "+label)+"  "+th.Muted.Render(item.desc))
	}
	return strings.Join(lines, "\n")
}

func (v *menuView) Keys() []key.Binding {
	return []key.Binding{
		binding("up,down", "↑/↓", "move"),
		binding("enter", "enter", "open"),
		binding("esc", "esc", "quit"),
	}
}

// ---------------------------------------------------------------------------
// Provisional view.
// ---------------------------------------------------------------------------

type placeholderView struct {
	baseView
	title string
}

func newPlaceholderView(title string) *placeholderView { return &placeholderView{title: title} }

func (v *placeholderView) Update(msg tea.Msg) (tea.Cmd, action) {
	if k, ok := msg.(tea.KeyPressMsg); ok && (k.String() == "esc" || k.String() == "backspace") {
		return nil, action{}
	}
	return nil, action{nav: navNone}
}

func (v *placeholderView) View(c viewCtx) string {
	return c.Theme.Title.Render(v.title) + "\n\n" + c.Theme.Text.Render("Not implemented yet.")
}

func (v *placeholderView) Keys() []key.Binding {
	return []key.Binding{binding("esc,backspace", "esc", "back")}
}

// ---------------------------------------------------------------------------
// Scrolling text and the dialog layout shared by the generic views.
// ---------------------------------------------------------------------------

// dialogRows is what a dialog uses besides its scrolling text: the heading,
// the position row and the button row.
const dialogRows = 3

// scrollBox shows text wrapped to its width in a viewport; the arrow, page,
// Home and End keys scroll it.
type scrollBox struct {
	raw string
	vp  viewport.Model
}

func newScrollBox() scrollBox { return scrollBox{vp: viewport.New()} }

func (s *scrollBox) setText(text string) {
	s.raw = strings.TrimRight(text, "\n")
	s.rewrap()
}

func (s *scrollBox) resize(width, height int) {
	s.vp.SetWidth(width)
	s.vp.SetHeight(max(height-dialogRows, 1))
	s.rewrap()
}

func (s *scrollBox) rewrap() {
	if s.vp.Width() <= 0 {
		s.vp.SetContentLines(nil)
		return
	}
	s.vp.SetContentLines(strings.Split(ansi.Wrap(s.raw, s.vp.Width(), ""), "\n"))
}

// handleKey scrolls for a scrolling key and reports whether it was one.
func (s *scrollBox) handleKey(k string) bool {
	switch k {
	case "up":
		s.vp.ScrollUp(1)
	case "down":
		s.vp.ScrollDown(1)
	case "pgup":
		s.vp.PageUp()
	case "pgdown":
		s.vp.PageDown()
	case "home":
		s.vp.GotoTop()
	case "end":
		s.vp.GotoBottom()
	default:
		return false
	}
	return true
}

func (s *scrollBox) scrollable() bool { return s.vp.TotalLineCount() > s.vp.Height() }

func (s *scrollBox) position() string {
	first := s.vp.YOffset() + 1
	return fmt.Sprintf("lines %d-%d of %d", first, first+s.vp.VisibleLineCount()-1, s.vp.TotalLineCount())
}

// dialogLayout draws a heading, the scrolling text, a position row that shows
// only when the text scrolls, and the button row, or the progress indicator
// while a write runs.
func dialogLayout(c viewCtx, title string, box *scrollBox, buttons, busyText string) string {
	th := c.Theme
	position := ""
	if box.scrollable() {
		position = th.Muted.Render(box.position())
	}
	bottom := buttons
	if c.Busy {
		bottom = c.Spinner + " " + th.Muted.Render(busyText)
	}
	return strings.Join([]string{th.Title.Render(title), box.vp.View(), position, bottom}, "\n")
}

// renderButtons draws a row of buttons; the chosen one is bracketed, so the
// choice never depends on color alone.
func renderButtons(c viewCtx, chosen int, labels ...string) string {
	parts := make([]string, len(labels))
	for i, label := range labels {
		if i == chosen {
			parts[i] = c.Theme.ButtonOn.Render("[" + label + "]")
		} else {
			parts[i] = c.Theme.ButtonOff.Render(" " + label + " ")
		}
	}
	return strings.Join(parts, "  ")
}

var scrollBinding = binding("up,down,pgup,pgdown", "↑/↓", "scroll")

// ---------------------------------------------------------------------------
// Summary and confirmation.
// ---------------------------------------------------------------------------

// confirmOptions configures a summary with Apply and Cancel. OnApply and
// OnCancel return what the view does next; a nil callback pops the view.
// StartOnCancel makes Cancel the initial choice, and DisableYes removes the
// one-key shortcut for applying (both for destructive summaries).
type confirmOptions struct {
	Title, Summary string
	StartOnCancel  bool
	DisableYes     bool
	OnApply        func() (tea.Cmd, action)
	OnCancel       func() (tea.Cmd, action)
}

type confirmView struct {
	baseView
	o      confirmOptions
	box    scrollBox
	choice int // 0 Apply, 1 Cancel
}

func newConfirmView(o confirmOptions) *confirmView {
	v := &confirmView{o: o, box: newScrollBox()}
	if o.StartOnCancel {
		v.choice = 1
	}
	v.box.setText(o.Summary)
	return v
}

func (v *confirmView) Resize(w, h int) { v.box.resize(w, h) }

func (v *confirmView) settle(cb func() (tea.Cmd, action)) (tea.Cmd, action) {
	if cb == nil {
		return nil, action{nav: navPop}
	}
	return cb()
}

func (v *confirmView) Update(msg tea.Msg) (tea.Cmd, action) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, action{nav: navNone}
	}
	switch name := k.String(); name {
	case "left":
		v.choice = 0
	case "right":
		v.choice = 1
	case "enter":
		if v.choice == 0 {
			return v.settle(v.o.OnApply)
		}
		return v.settle(v.o.OnCancel)
	case "y":
		if !v.o.DisableYes {
			return v.settle(v.o.OnApply)
		}
	case "n", "esc", "backspace":
		return v.settle(v.o.OnCancel)
	default:
		v.box.handleKey(name)
	}
	return nil, action{nav: navNone}
}

func (v *confirmView) View(c viewCtx) string {
	return dialogLayout(c, v.o.Title, &v.box, renderButtons(c, v.choice, "Apply", "Cancel"), "Applying…")
}

func (v *confirmView) Keys() []key.Binding {
	answer := binding("y,n", "y/n", "answer")
	if v.o.DisableYes {
		answer = binding("n", "n", "cancel")
	}
	return []key.Binding{
		scrollBinding,
		binding("left,right", "←/→", "choose"),
		binding("enter", "enter", "accept"),
		answer,
		binding("esc", "esc", "cancel"),
	}
}

// ---------------------------------------------------------------------------
// Notice.
// ---------------------------------------------------------------------------

// noticeView shows a text and a Continue button. A nil onContinue pops the
// view.
type noticeView struct {
	baseView
	title      string
	box        scrollBox
	onContinue func() (tea.Cmd, action)
}

func newNoticeView(title, text string, onContinue func() (tea.Cmd, action)) *noticeView {
	v := &noticeView{title: title, box: newScrollBox(), onContinue: onContinue}
	v.box.setText(text)
	return v
}

func (v *noticeView) Resize(w, h int) { v.box.resize(w, h) }

func (v *noticeView) Update(msg tea.Msg) (tea.Cmd, action) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, action{nav: navNone}
	}
	switch name := k.String(); name {
	case "enter", "esc", "backspace":
		if v.onContinue == nil {
			return nil, action{nav: navPop}
		}
		return v.onContinue()
	default:
		v.box.handleKey(name)
	}
	return nil, action{nav: navNone}
}

func (v *noticeView) View(c viewCtx) string {
	return dialogLayout(c, v.title, &v.box, renderButtons(c, 0, "Continue"), "Working…")
}

func (v *noticeView) Keys() []key.Binding {
	return []key.Binding{scrollBinding, binding("enter,esc,backspace", "enter", "continue")}
}

// ---------------------------------------------------------------------------
// Recovery on open.
// ---------------------------------------------------------------------------

// recoverDoneMsg is the result of the recovery Cmd.
type recoverDoneMsg struct {
	text string
	err  error
}

// recoveryView offers to recover an interrupted operation or onboarding found
// when the application opens (design.md "Recuperar"). Recover runs the same
// detection and recovery functions `hive recover` and the install flow use.
type recoveryView struct {
	baseView
	kind     management.PendingKind
	opts     management.Options
	explicit bool
	deps     installDependencies
	box      scrollBox
	choice   int // 0 Recover, 1 Leave it
	done     bool
	failed   bool
}

func newRecoveryView(kind management.PendingKind, opts management.Options, explicitStateDir bool, deps installDependencies) *recoveryView {
	v := &recoveryView{kind: kind, opts: opts, explicit: explicitStateDir, deps: deps, box: newScrollBox()}
	later := recoveryTextOnOpen(explicitStateDir)(opts.StateDir, true)
	if kind == management.PendingOnboarding {
		v.box.setText(fmt.Sprintf("Optional onboarding was interrupted.\n\nRecovery directory: %s\n\n"+
			"Recover reconciles the pending optional steps. Leave it keeps them pending. %s to recover them later.", opts.StateDir, later))
	} else {
		v.box.setText(fmt.Sprintf("An operation was interrupted.\n\nRecovery directory: %s\n\n"+
			"Recover settles it from its journal. Leave it keeps the operation pending. %s to recover it later.", opts.StateDir, later))
	}
	return v
}

func (v *recoveryView) Resize(w, h int) { v.box.resize(w, h) }

func (v *recoveryView) recover() (tea.Cmd, action) {
	return recoverPendingCmd(v.kind, copyOptions(v.opts), v.explicit, v.deps), action{nav: navNone, write: true}
}

func (v *recoveryView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case recoverDoneMsg:
		v.done, v.failed = true, msg.err != nil
		text := msg.text
		if msg.err != nil {
			hint := recoveryTextOnOpen(v.explicit)(v.opts.StateDir, true)
			text = fmt.Sprintf("Recovery failed: %v\n\n%s.", msg.err, hint)
		}
		v.box.setText(text)
		return nil, action{nav: navNone}
	case tea.KeyPressMsg:
		name := msg.String()
		if v.done {
			switch name {
			case "enter", "esc", "backspace":
				return nil, action{nav: navPop}
			}
			v.box.handleKey(name)
			return nil, action{nav: navNone}
		}
		switch name {
		case "left":
			v.choice = 0
		case "right":
			v.choice = 1
		case "r":
			return v.recover()
		case "l", "esc", "backspace":
			return nil, action{nav: navPop}
		case "enter":
			if v.choice == 0 {
				return v.recover()
			}
			return nil, action{nav: navPop}
		default:
			v.box.handleKey(name)
		}
	}
	return nil, action{nav: navNone}
}

func (v *recoveryView) View(c viewCtx) string {
	switch {
	case v.done && v.failed:
		return dialogLayout(c, "Recovery failed", &v.box, renderButtons(c, 0, "Continue"), "Recovering…")
	case v.done:
		return dialogLayout(c, "Recovery finished", &v.box, renderButtons(c, 0, "Continue"), "Recovering…")
	}
	return dialogLayout(c, "Recovery needed", &v.box, renderButtons(c, v.choice, "Recover", "Leave it"), "Recovering…")
}

func (v *recoveryView) Keys() []key.Binding {
	if v.done {
		return []key.Binding{scrollBinding, binding("enter,esc,backspace", "enter", "continue")}
	}
	return []key.Binding{
		binding("left,right", "←/→", "choose"),
		binding("enter", "enter", "accept"),
		binding("esc", "esc", "leave it"),
	}
}
