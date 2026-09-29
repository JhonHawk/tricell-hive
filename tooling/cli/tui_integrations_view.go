// tui_integrations_view.go is the Integrations view (#46, design.md "Vistas"):
// a four-row list of the optional capabilities with a cursor, and under it the
// selected row's detail (local evidence, last onboarding status, source and
// next step) in a text box that wraps without cutting and scrolls. It only
// reads. The load runs as a Cmd whose result is addressed to the view (owned)
// and carries a sequence number, so a double reload or a result that arrives
// after the view was left is dropped.
package main

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// integrationsFixedRows are the title, the list header and the four list rows;
// the detail gets the rest of the view's area. scrollBox.resize subtracts the
// dialog rows, so the box's height is set directly.
const integrationsFixedRows = 6

type integrationsLoadedMsg struct {
	owned
	seq  int
	rows []integrationRow
	err  string
}

type integrationsView struct {
	cfg     appConfig
	deps    doctorDeps
	seq     int
	loading bool
	loaded  bool
	rows    []integrationRow
	loadErr string
	cursor  int
	box     scrollBox // the selected row's detail
	w, h    int
}

// newIntegrationsView opens Integrations over the real machine. Tests use
// newIntegrationsViewWith to replace the dependencies.
func newIntegrationsView(cfg appConfig) *integrationsView {
	return newIntegrationsViewWith(cfg, realDoctorDeps())
}

func newIntegrationsViewWith(cfg appConfig, deps doctorDeps) *integrationsView {
	return &integrationsView{cfg: cfg, deps: deps, loading: true, box: newHangingScrollBox()}
}

func (v *integrationsView) Init() tea.Cmd { return v.reload() }

func (v *integrationsView) TextFocused() bool { return false }

func (v *integrationsView) NeedsSpinner() bool { return v.loading }

func (v *integrationsView) reload() tea.Cmd {
	v.seq++
	v.loading = true
	seq, o, deps := v.seq, copyOptions(v.cfg.Options), v.deps
	return func() tea.Msg {
		msg := integrationsLoadedMsg{owned: owned{v}, seq: seq}
		msg.rows, msg.err = collectIntegrationRows(o, deps)
		return msg
	}
}

func (v *integrationsView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case integrationsLoadedMsg:
		if msg.seq != v.seq {
			break
		}
		v.loading, v.loaded = false, true
		v.rows, v.loadErr = msg.rows, msg.err
		v.cursor = min(v.cursor, max(len(v.rows)-1, 0))
		v.showDetail(false)
	case tea.KeyPressMsg:
		return v.onKey(msg.String())
	}
	return nil, action{nav: navNone}
}

func (v *integrationsView) onKey(name string) (tea.Cmd, action) {
	switch name {
	case "esc", "backspace":
		return nil, action{}
	case "r":
		if !v.loading {
			return v.reload(), action{nav: navNone}
		}
	case "up":
		v.moveCursor(-1)
	case "down":
		v.moveCursor(1)
	case "pgup", "pgdown", "home", "end":
		v.box.handleKey(name)
	}
	return nil, action{nav: navNone}
}

func (v *integrationsView) moveCursor(delta int) {
	next := min(max(v.cursor+delta, 0), max(len(v.rows)-1, 0))
	if next != v.cursor {
		v.cursor = next
		v.showDetail(true)
	}
}

// showDetail puts the selected row's detail in the box. A different row starts
// at the top; a reload keeps the scroll position.
func (v *integrationsView) showDetail(gotoTop bool) {
	if v.cursor >= len(v.rows) {
		v.box.setText("")
		return
	}
	v.box.setText(strings.Join(v.rows[v.cursor].detailLines(), "\n"))
	if gotoTop {
		v.box.vp.GotoTop()
	}
}

func (v *integrationsView) Resize(w, h int) { v.w, v.h = w, h; v.layout() }

func (v *integrationsView) layout() {
	v.box.vp.SetWidth(max(v.w, 1))
	v.box.vp.SetHeight(max(v.h-integrationsFixedRows, 1))
	v.box.rewrap()
}

func (v *integrationsView) View(c viewCtx) string {
	if c.Width != v.w || c.Height != v.h {
		v.w, v.h = c.Width, c.Height
		v.layout()
	}
	th := c.Theme
	title := th.Title.Render("Integrations")
	if !v.loaded {
		return title + "\n" + c.Spinner + " " + th.Muted.Render("Checking integrations…")
	}
	switch {
	case v.loading:
		title += "  " + c.Spinner + " " + th.Muted.Render("Refreshing…")
	case v.loadErr != "":
		title += "  " + th.Danger.Render("Onboarding record unreadable") + th.Muted.Render(" · r to retry")
	case v.box.scrollable():
		title += "  " + th.Muted.Render(v.box.position())
	}
	lines := []string{title, "  " + th.Muted.Render(truncateRunes(integrationColumns("Integration", "Local", "Onboarding record"), max(v.w-2, 1)))}
	for i, r := range v.rows {
		text := truncateRunes(r.summaryLine(), max(v.w-2, 1))
		if i == v.cursor {
			lines = append(lines, th.Accent.Render("> "+text))
		} else {
			lines = append(lines, "  "+th.Text.Render(text))
		}
	}
	for len(lines) < integrationsFixedRows {
		lines = append(lines, "")
	}
	lines = append(lines, v.box.vp.View())
	return strings.Join(lines, "\n")
}

func (v *integrationsView) Keys() []key.Binding {
	return []key.Binding{
		binding("up,down", "↑/↓", "select"),
		binding("pgup,pgdown", "PgUp/PgDn", "scroll detail"),
		binding("r", "r", "reload"),
		binding("esc", "esc", "back"),
	}
}
