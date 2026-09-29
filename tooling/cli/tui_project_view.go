// tui_project_view.go is the Project view (design.md "Vistas"): the result of
// validating the `## Hive` section of the current repository's AGENTS.md. Three
// rows are fixed (heading, AGENTS.md path, position); the verdict or findings
// and the values read scroll between them. It only reads. The load runs as a
// Cmd whose result is addressed to the view (owned) and carries a sequence
// number, so a double reload or a result that arrives after the view was left
// is dropped.
package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// projectFixedRows are the heading, the AGENTS.md path and the position row.
const projectFixedRows = 3

type projectLoadedMsg struct {
	owned
	seq   int
	check projectCheck
}

type projectView struct {
	cfg     appConfig
	dir     string // "" means the current directory
	deps    doctorDeps
	git     gitRunner
	seq     int
	loading bool
	loaded  bool
	failed  bool // the section could not be fully checked
	check   projectCheck
	box     scrollBox
	w, h    int
}

// newProjectView opens Project over the current directory. Tests use
// newProjectViewWith to choose the directory and replace the dependencies.
func newProjectView(cfg appConfig) *projectView {
	return newProjectViewWith(cfg, "", realDoctorDeps())
}

func newProjectViewWith(cfg appConfig, dir string, deps doctorDeps) *projectView {
	return &projectView{cfg: cfg, dir: dir, deps: deps, git: newGitRunner(), loading: true, box: newHangingScrollBox()}
}

func (v *projectView) Init() tea.Cmd { return v.reload() }

func (v *projectView) TextFocused() bool { return false }

func (v *projectView) NeedsSpinner() bool { return v.loading }

func (v *projectView) reload() tea.Cmd {
	v.seq++
	v.loading = true
	seq, dir, git := v.seq, v.dir, v.git
	return func() tea.Msg {
		return projectLoadedMsg{owned: owned{v}, seq: seq, check: checkProject(dir, git)}
	}
}

func (v *projectView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case projectLoadedMsg:
		if msg.seq != v.seq {
			break
		}
		v.loading, v.loaded = false, true
		v.check = msg.check
		v.failed = msg.check.Section.Err != ""
		v.box.setText(projectBoxText(msg.check.Section))
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

// projectBoxText is the scrolling text: the section's lines, after the reason
// the check could not finish when it could not.
func projectBoxText(s doctorSection) string {
	text := ""
	if s.Err != "" {
		first, rest := errLines(s.Err)
		text = "Could not check everything: " + first + "\n"
		for _, l := range rest {
			text += "  " + l + "\n"
		}
	}
	for _, l := range s.Lines {
		text += l + "\n"
	}
	return text
}

func (v *projectView) Resize(w, h int) { v.w, v.h = w, h; v.layout() }

// layout gives the scrolling text the area between the AGENTS.md path and the
// position row.
func (v *projectView) layout() {
	v.box.vp.SetWidth(v.w)
	v.box.vp.SetHeight(max(v.h-projectFixedRows, 1))
	v.box.rewrap()
}

// locationLine names what was checked: the AGENTS.md path, or the directory
// when it is not inside a repository. A path too long for the row loses its
// beginning, because the end is what tells one repository from another.
func (v *projectView) locationLine() string {
	label, path := v.check.location()
	if room := v.w - lipgloss.Width(label); v.w > 0 && lipgloss.Width(path) > room {
		if room < 2 {
			return ""
		}
		path = ansi.TruncateLeft(path, lipgloss.Width(path)-(room-1), "…")
	}
	return label + path
}

func (v *projectView) View(c viewCtx) string {
	if c.Width != v.w || c.Height != v.h {
		v.w, v.h = c.Width, c.Height
		v.layout()
	}
	th := c.Theme
	position := ""
	switch {
	case v.loading && !v.loaded:
		return th.Title.Render("Project") + "\n" + c.Spinner + " " + th.Muted.Render("Checking the ## Hive section…")
	case v.loading:
		position = c.Spinner + " " + th.Muted.Render("Refreshing…")
	case v.failed:
		position = th.Danger.Render("The check failed") + th.Muted.Render(" · r to retry")
	case v.box.scrollable():
		position = th.Muted.Render(v.box.position())
	}
	return th.Title.Render("Project") + "\n" + th.Muted.Render(v.locationLine()) + "\n" + v.box.vp.View() + "\n" + position
}

func (v *projectView) Keys() []key.Binding {
	return []key.Binding{
		scrollBinding,
		binding("r", "r", "reload"),
		binding("esc", "esc", "back"),
	}
}
