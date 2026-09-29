// tui_doctor_view.go is the Diagnostics view (design.md "Vistas"): the CLIs,
// Installation and Sessions sections of the doctor report in one scrolling
// text, with a position row underneath. It only reads: the load runs as a Cmd
// whose result is addressed to the view (owned) and carries a sequence number,
// so a double reload or a result that arrives after the view was left is
// dropped.
package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
)

// diagnosticsFixedRows are the heading and the position row; the scrolling
// text gets the rest of the view's area.
const diagnosticsFixedRows = 2

type doctorLoadedMsg struct {
	owned
	seq      int
	sections [3]doctorSection
}

type doctorView struct {
	cfg     appConfig
	deps    doctorDeps
	seq     int
	loading bool
	loaded  bool
	failed  bool // a section could not be fully checked
	box     scrollBox
	w, h    int
}

// newDoctorView opens Diagnostics over the real machine. Tests use
// newDoctorViewWith to replace the dependencies.
func newDoctorView(cfg appConfig) *doctorView { return newDoctorViewWith(cfg, realDoctorDeps()) }

func newDoctorViewWith(cfg appConfig, deps doctorDeps) *doctorView {
	return &doctorView{cfg: cfg, deps: deps, loading: true, box: newHangingScrollBox()}
}

func (v *doctorView) options() management.Options { return copyOptions(v.cfg.Options) }

func (v *doctorView) Init() tea.Cmd { return v.reload() }

func (v *doctorView) TextFocused() bool { return false }

func (v *doctorView) NeedsSpinner() bool { return v.loading }

func (v *doctorView) reload() tea.Cmd {
	v.seq++
	v.loading = true
	seq, o, deps := v.seq, v.options(), v.deps
	return func() tea.Msg {
		msg := doctorLoadedMsg{owned: owned{v}, seq: seq}
		msg.sections[0], msg.sections[1], msg.sections[2] = collectHostSections(o, deps)
		return msg
	}
}

func (v *doctorView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case doctorLoadedMsg:
		if msg.seq != v.seq {
			break
		}
		v.loading, v.loaded, v.failed = false, true, false
		for _, s := range msg.sections {
			v.failed = v.failed || s.Err != ""
		}
		v.box.setText(doctorSectionsText(msg.sections[:]...))
	case tea.KeyPressMsg:
		switch name := msg.String(); name {
		case "esc", "backspace":
			return nil, action{}
		case "r":
			if !v.loading {
				return v.reload(), action{nav: navNone}
			}
		default:
			v.box.handleKey(name)
		}
	}
	return nil, action{nav: navNone}
}

func (v *doctorView) Resize(w, h int) { v.w, v.h = w, h; v.layout() }

// layout gives the scrolling text the area between the heading and the
// position row.
func (v *doctorView) layout() {
	v.box.vp.SetWidth(v.w)
	v.box.vp.SetHeight(max(v.h-diagnosticsFixedRows, 1))
	v.box.rewrap()
}

func (v *doctorView) View(c viewCtx) string {
	if c.Width != v.w || c.Height != v.h {
		v.w, v.h = c.Width, c.Height
		v.layout()
	}
	th := c.Theme
	position := ""
	switch {
	case v.loading && !v.loaded:
		return th.Title.Render("Diagnostics") + "\n" + c.Spinner + " " + th.Muted.Render("Checking CLIs, installation and sessions…")
	case v.loading:
		position = c.Spinner + " " + th.Muted.Render("Refreshing…")
	case v.failed:
		position = th.Danger.Render("Some checks failed") + th.Muted.Render(" · r to retry")
	case v.box.scrollable():
		position = th.Muted.Render(v.box.position())
	}
	return th.Title.Render("Diagnostics") + "\n" + v.box.vp.View() + "\n" + position
}

func (v *doctorView) Keys() []key.Binding {
	return []key.Binding{
		scrollBinding,
		binding("r", "r", "reload"),
		binding("esc", "esc", "back"),
	}
}
