// tui_views.go holds the application's views: the menu, and the generic views the
// later screens stack on top of their source view: summary with confirmation,
// notice, and recovery on open (design.md "Vistas").
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
// Text helpers shared by the views.
// ---------------------------------------------------------------------------

func padRight(s string, width int) string {
	if n := len([]rune(s)); n < width {
		return s + strings.Repeat(" ", width-n)
	}
	return s
}

// wrapLines wraps text to width columns, one entry per line.
func wrapLines(text string, width int) []string {
	if text == "" {
		return nil
	}
	return strings.Split(wrapBreakingPaths(text, max(width, 1)), "\n")
}

// wrapBreakingPaths wraps text to width, filling each line with whole words.
// A word that fits the width is never split, not even at a hyphen: it moves
// whole to the next line. A word wider than the line breaks after a "/"
// instead of in the middle of a name, keeping each slash at the end of its
// line; a segment with no slash that is still too wide, or a word with none,
// falls back to a hard break. Runs of spaces inside a line are kept.
func wrapBreakingPaths(text string, width int) string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		out = append(out, fillLine(line, width)...)
	}
	return strings.Join(out, "\n")
}

// fillLine wraps one line without newlines for wrapBreakingPaths. The line's
// leading spaces stay on a wide word's pieces, so an indented list entry keeps
// its indent; spaces that fall at a break are dropped.
func fillLine(line string, width int) []string {
	if ansi.StringWidth(line) <= width {
		return []string{line}
	}
	var lines []string
	cur, started, fresh := "", false, false
	flush := func() {
		lines = append(lines, strings.TrimRight(cur, " "))
		cur, started, fresh = "", false, true
	}
	for _, word := range strings.Split(line, " ") {
		if word == "" && fresh {
			continue // spaces at a break do not start the next line
		}
		if ansi.StringWidth(word) > width {
			indent := ""
			if strings.TrimSpace(cur) == "" && !fresh && len(lines) == 0 {
				indent = cur
				if started {
					indent += " "
				}
			} else if strings.TrimSpace(cur) != "" {
				flush()
			}
			room := width - ansi.StringWidth(indent)
			if room < 10 {
				indent, room = "", width
			}
			pieces := slashChunks(word, room)
			for i, piece := range pieces {
				for ansi.StringWidth(piece) > room {
					head := ansi.Truncate(piece, room, "")
					lines = append(lines, indent+head)
					piece = strings.TrimPrefix(piece, head)
				}
				if i < len(pieces)-1 {
					lines = append(lines, indent+piece)
					continue
				}
				cur, started, fresh = indent+piece, true, false
			}
			continue
		}
		candidate := word
		if started {
			candidate = cur + " " + word
		}
		if ansi.StringWidth(candidate) <= width {
			cur, started, fresh = candidate, true, false
			continue
		}
		flush()
		if word != "" {
			cur, started, fresh = word, true, false
		}
	}
	if started {
		flush()
	}
	return lines
}

// slashChunks splits a token into pieces of at most width columns, cutting only
// after a "/" where it can. A single segment wider than width stays whole for
// the caller's hard break.
func slashChunks(tok string, width int) []string {
	var chunks []string
	current := ""
	for _, seg := range strings.SplitAfter(tok, "/") {
		if seg == "" {
			continue
		}
		if current != "" && ansi.StringWidth(current+seg) > width {
			chunks = append(chunks, current)
			current = ""
		}
		current += seg
	}
	return append(chunks, current)
}

// inputView draws a text input within its own width. A value that does not
// fit scrolls around the cursor, and a "…" marks each cut: at the start only
// when text is hidden to the left, at the end only when text is hidden to the
// right. The cursor is a reversed cell (a space after the last character) that
// always fits inside the width. A blurred field shows the head of its value,
// and an empty one draws the input's own placeholder.
//
// The window is computed here from the value, the cursor position and the
// width, not read back from the text input's rendering: the input keeps its
// own scroll offset privately, and a rendering can be neither measured for the
// cut nor searched for in the value reliably (repeated text, trailing spaces).
func inputView(in textinput.Model, th *appTheme) string {
	value := []rune(in.Value())
	width := in.Width()
	if len(value) == 0 || width <= 0 {
		return in.View()
	}
	focused := in.Focused()
	cursor := 0
	if focused {
		cursor = min(max(in.Position(), 0), len(value))
	}
	cellWidth := func(r rune) int { return max(lipgloss.Width(string(r)), 1) }
	span := func(from, to int) int {
		n := 0
		for _, r := range value[from:to] {
			n += cellWidth(r)
		}
		return n
	}
	// need is the columns the window [start, end) takes: its text, the cursor
	// cell after the last character, and the markers for what is cut.
	need := func(start, end int) int {
		n := span(start, end)
		if focused && cursor == len(value) && end == len(value) {
			n++
		}
		if start > 0 {
			n++
		}
		if end < len(value) {
			n++
		}
		return n
	}
	start, end := 0, len(value)
	if need(0, len(value)) > width {
		// Anchor the window on the cursor with a little context to its right,
		// fill the room to the left, then use what is left on the right.
		end = min(len(value), cursor+1+width/4)
		start = end
		for start > 0 && need(start-1, end) <= width {
			start--
		}
		for end < len(value) && need(start, end+1) <= width {
			end++
		}
		for start > 0 && need(start-1, end) <= width {
			start--
		}
		if start > cursor {
			// So narrow that not even the cursor's character fit: keep it.
			start, end = cursor, min(len(value), cursor+1)
		}
	}
	var b strings.Builder
	if start > 0 {
		b.WriteString(th.Muted.Render("…"))
	}
	text := th.Text
	for i := start; i < end; i++ {
		if focused && i == cursor {
			b.WriteString(text.Reverse(true).Render(string(value[i])))
			continue
		}
		b.WriteString(text.Render(string(value[i])))
	}
	if focused && cursor == len(value) && end == len(value) {
		b.WriteString(text.Reverse(true).Render(" "))
	}
	if end < len(value) {
		b.WriteString(th.Muted.Render("…"))
	}
	return b.String()
}

// truncateRunes returns s unchanged if it fits within width runes,
// otherwise truncates it to width RUNES with a trailing "…". It counts and
// slices by rune, not by byte: "…" (U+2026) is three UTF-8 bytes but one
// terminal column, and every caller's own width here is a column budget —
// slicing the underlying string by byte index would silently make a long
// label two bytes (not columns) over budget.
func truncateRunes(s string, width int) string {
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	if width <= 1 {
		return strings.Repeat(".", max(width, 0))
	}
	return string([]rune(s)[:width-1]) + "…"
}

// ---------------------------------------------------------------------------
// Menu.
// ---------------------------------------------------------------------------

// menuItem is one row of the main menu. A nil open means Quit; otherwise it
// builds the view the entry opens from the application's configuration.
type menuItem struct {
	label, desc string
	open        func(cfg appConfig) view
}

// mainMenuItems is the menu's fixed set, in order (design.md "Menú (D2-A)").
// Each entry but Quit opens its own view (T8 and T9; the four read-only
// views of #46).
var mainMenuItems = []menuItem{
	{"CLIs", "Install, remove and check CLI hosts", func(cfg appConfig) view { return newHostsView(cfg) }},
	{"Update", "Update Hive from a Git commit", func(cfg appConfig) view { return newUpdateView(cfg) }},
	{"Releases", "Go back to a retained release", func(cfg appConfig) view { return newReleasesView(cfg) }},
	{"Voice", "Choose the assistant voice", func(cfg appConfig) view { return newVoiceView(cfg) }},
	{"Diagnostics", "Check CLI versions, the installation and open sessions", func(cfg appConfig) view { return newDoctorView(cfg) }},
	{"Models", "See and change the model and effort of each role", func(cfg appConfig) view { return newModelsView(cfg) }},
	{"Integrations", "Check Engram, Context7, pi-subagents and agent-browser", func(cfg appConfig) view { return newIntegrationsView(cfg) }},
	{"Project", "Check or edit this repository's ## Hive section", func(cfg appConfig) view { return newProjectView(cfg) }},
	{"Quit", "Leave Hive", nil},
}

type menuView struct {
	baseView
	cfg    appConfig
	cursor int
}

func newMenuView(cfg appConfig) *menuView { return &menuView{cfg: cfg} }

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
		return nil, action{nav: navPush, push: item.open(v.cfg)}
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
	// hanging continues a line that does not fit under its own indentation
	// plus two spaces, for the read-only views' indented rows and paths.
	hanging bool
}

func newScrollBox() scrollBox { return scrollBox{vp: viewport.New()} }

// newHangingScrollBox is a scrollBox whose wrapped lines keep their indentation.
func newHangingScrollBox() scrollBox { return scrollBox{vp: viewport.New(), hanging: true} }

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
	if !s.hanging {
		s.vp.SetContentLines(strings.Split(wrapBreakingPaths(s.raw, s.vp.Width()), "\n"))
		return
	}
	var lines []string
	for _, line := range strings.Split(s.raw, "\n") {
		lines = append(lines, wrapHanging(line, s.vp.Width())...)
	}
	s.vp.SetContentLines(lines)
}

// wrapHanging wraps one line to width. A line that does not fit continues
// under its own leading spaces plus two, so the row it belongs to stays
// visible; when that leaves too little room it falls back to plain wrapping.
func wrapHanging(line string, width int) []string {
	if ansi.StringWidth(line) <= width {
		return []string{line}
	}
	lead := len(line) - len(strings.TrimLeft(line, " "))
	room := width - lead - 2
	if room < 10 {
		return strings.Split(wrapBreakingPaths(line, width), "\n")
	}
	parts := strings.Split(wrapBreakingPaths(line[lead:], room), "\n")
	out := make([]string, len(parts))
	for i, part := range parts {
		if i == 0 {
			out[i] = strings.Repeat(" ", lead) + part
			continue
		}
		out[i] = strings.Repeat(" ", lead+2) + part
	}
	return out
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
// one-key shortcut for applying (both for destructive summaries). ApplyLabel
// replaces the "Apply" button's text (for example "Accept").
type confirmOptions struct {
	Title, Summary string
	ApplyLabel     string
	StartOnCancel  bool
	DisableYes     bool
	// Hanging continues a wrapped line under its own indentation, for a summary
	// made of indented items.
	Hanging  bool
	OnApply  func() (tea.Cmd, action)
	OnCancel func() (tea.Cmd, action)
}

type confirmView struct {
	baseView
	o      confirmOptions
	box    scrollBox
	choice int // 0 Apply, 1 Cancel
}

func newConfirmView(o confirmOptions) *confirmView {
	v := &confirmView{o: o, box: newScrollBox()}
	if o.Hanging {
		v.box = newHangingScrollBox()
	}
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
	apply := v.o.ApplyLabel
	if apply == "" {
		apply = "Apply"
	}
	return dialogLayout(c, v.o.Title, &v.box, renderButtons(c, v.choice, apply, "Cancel"), "Applying…")
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
// view. By default Esc and Backspace also continue; withBack makes them a
// separate "back" answer instead, for a notice that sits between two steps.
type noticeView struct {
	baseView
	title      string
	box        scrollBox
	onContinue func() (tea.Cmd, action)
	onBack     func() (tea.Cmd, action)
}

// withBack sets what Esc and Backspace do, apart from Continue.
func (v *noticeView) withBack(onBack func() (tea.Cmd, action)) *noticeView {
	v.onBack = onBack
	return v
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
	case "esc", "backspace":
		if v.onBack != nil {
			return v.onBack()
		}
		fallthrough
	case "enter":
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
	if v.onBack != nil {
		return []key.Binding{scrollBinding, binding("enter", "enter", "continue"), binding("esc,backspace", "esc", "stop here")}
	}
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
