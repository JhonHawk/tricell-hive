// tui_models_view.go is the Models view (#46, design.md "Vistas"): the model
// and effort each installed agent role gets on each registered CLI, read from
// the release snapshot the role was installed from. It never writes; changing
// a model means editing integrations/agent-profiles.json and running
// `hive update`, which the footer says.
package main

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
)

// modelsFixedRows are the rows around the table: title, CLI row, table
// header, position, and the two-line footer.
const modelsFixedRows = 6

// modelsLoadedMsg is the result of loading the registered CLIs and the rows.
type modelsLoadedMsg struct {
	owned
	seq   int
	hosts []string
	rows  []management.ModelRow
	err   error
}

type modelsView struct {
	cfg     appConfig
	seq     int
	loading bool
	loadErr string
	hosts   []string
	rows    []management.ModelRow
	host    string // the selected CLI
	box     scrollBox
	width   int
	height  int
}

func newModelsView(cfg appConfig) *modelsView {
	return &modelsView{cfg: cfg, loading: true, box: newScrollBox()}
}

func (v *modelsView) Init() tea.Cmd { return v.reload() }

func (v *modelsView) NeedsSpinner() bool { return v.loading }

func (v *modelsView) TextFocused() bool { return false }

func (v *modelsView) reload() tea.Cmd {
	v.seq++
	v.loading = true
	seq, o := v.seq, copyOptions(v.cfg.Options)
	return func() tea.Msg {
		msg := modelsLoadedMsg{owned: owned{v}, seq: seq}
		msg.hosts, msg.rows, msg.err = collectModels(o)
		return msg
	}
}

func (v *modelsView) Resize(w, h int) {
	v.width, v.height = w, h
	v.layout()
}

// hostRows are the rows of the selected CLI.
func (v *modelsView) hostRows() []management.ModelRow {
	var out []management.ModelRow
	for _, r := range v.rows {
		if r.Host == v.host {
			out = append(out, r)
		}
	}
	return out
}

// layout redraws the table for the selected CLI and the current size.
func (v *modelsView) layout() {
	v.box.vp.SetWidth(max(v.width, 1))
	v.box.vp.SetHeight(max(v.height-modelsFixedRows, 1))
	_, lines := modelTable(v.hostRows(), v.width)
	v.box.setText(strings.Join(lines, "\n"))
	v.box.vp.GotoTop()
}

func (v *modelsView) hostIndex() int {
	for i, h := range v.hosts {
		if h == v.host {
			return i
		}
	}
	return -1
}

func (v *modelsView) selectHost(i int) {
	if i < 0 || i >= len(v.hosts) || v.hosts[i] == v.host {
		return
	}
	v.host = v.hosts[i]
	v.layout()
}

func (v *modelsView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case modelsLoadedMsg:
		if msg.seq != v.seq {
			break
		}
		v.loading = false
		if msg.err != nil {
			v.loadErr, v.hosts, v.rows = msg.err.Error(), nil, nil
			v.layout()
			break
		}
		v.loadErr, v.hosts, v.rows = "", msg.hosts, msg.rows
		if v.hostIndex() < 0 {
			v.host = ""
			if len(v.hosts) > 0 {
				v.host = v.hosts[0]
			}
		}
		v.layout()
	case tea.KeyPressMsg:
		return v.onKey(msg.String())
	}
	return nil, action{nav: navNone}
}

func (v *modelsView) onKey(name string) (tea.Cmd, action) {
	switch name {
	case "esc", "backspace":
		return nil, action{}
	case "r":
		if !v.loading {
			return v.reload(), action{nav: navNone}
		}
	case "left":
		v.selectHost(v.hostIndex() - 1)
	case "right":
		v.selectHost(v.hostIndex() + 1)
	default:
		v.box.handleKey(name)
	}
	return nil, action{nav: navNone}
}

// hostRow draws the CLIs, the selected one in brackets so the choice does not
// depend on color.
func (v *modelsView) hostRow(th *appTheme) string {
	parts := make([]string, len(v.hosts))
	for i, h := range v.hosts {
		if h == v.host {
			parts[i] = th.Accent.Render("[" + h + "]")
		} else {
			parts[i] = th.Muted.Render(" " + h + " ")
		}
	}
	return strings.Join(parts, " ")
}

func (v *modelsView) View(c viewCtx) string {
	th := c.Theme
	title := th.Title.Render("Models")
	footer := []string{
		th.Muted.Render("To change a model or effort, edit integrations/agent-profiles.json"),
		th.Muted.Render("in the Hive checkout, then run hive update."),
	}
	var host, header, position string
	body := []string{}
	switch {
	case v.loading && len(v.rows) == 0 && len(v.hosts) == 0:
		body = append(body, c.Spinner+" "+th.Muted.Render("Loading models…"))
	case v.loadErr != "":
		for _, l := range wrapLines("Cannot read the models: "+v.loadErr, c.Width) {
			body = append(body, th.Danger.Render(l))
		}
		body = append(body, th.Muted.Render("Press r to retry."))
	case len(v.hosts) == 0:
		body = append(body, th.Text.Render("No CLI hosts are registered"))
	default:
		host = v.hostRow(th)
		if rows := v.hostRows(); len(rows) == 0 {
			body = append(body, th.Text.Render("No agents installed for "+v.host))
		} else {
			h, _ := modelTable(rows, c.Width)
			header = th.Muted.Render(h)
			body = strings.Split(v.box.vp.View(), "\n")
			if v.box.scrollable() {
				position = th.Muted.Render(v.box.position())
			}
		}
	}
	room := max(c.Height-modelsFixedRows, 1)
	if len(body) > room {
		body = body[:room]
	}
	for len(body) < room {
		body = append(body, "")
	}
	lines := append([]string{title, host, header}, body...)
	lines = append(lines, position)
	lines = append(lines, footer...)
	return strings.Join(lines, "\n")
}

func (v *modelsView) Keys() []key.Binding {
	return []key.Binding{
		binding("left,right", "←/→", "switch CLI"),
		scrollBinding,
		binding("r", "r", "reload"),
		binding("esc", "esc", "back"),
	}
}
