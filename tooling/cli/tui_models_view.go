// tui_models_view.go is the Models view (#46, design.md "Vistas" and "Edición
// en la vista Models"): the model and effort each installed agent role gets on
// each registered CLI, read from the release snapshot the role was installed
// from. A cursor selects a role; Enter opens a panel that edits its model and
// effort, and x resets a marked role. Both build the plan `hive models set` or
// `reset` builds, show its summary and confirm before writing.
package main

import (
	"bytes"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"tricell-hive/integrations/agents"
	"tricell-hive/tooling/management"
)

// modelsFixedRows are the rows around the table: title, CLI row, table
// header, position, and the two-line footer.
const modelsFixedRows = 6

// modelsPanelRows are the rows the edit panel takes above the table: the
// role, Model, Effort, an error row and a blank one.
const modelsPanelRows = 5

const (
	modelsPanelLabel  = 9 // "Model" and "Effort" plus the gap, after the two-column cursor
	effortReleaseText = "release default"
)

// Messages of the edit flow. flow lets a result that arrives after the panel
// was closed, or after another request, be dropped.
type (
	modelsPlannedMsg struct {
		owned
		flow      int
		plan      management.Plan
		summary   string
		unchanged bool
		title     string
		err       error
	}
	modelsAppliedMsg struct {
		owned
		flow    int
		err     error
		pending management.PendingKind
		// filesChanged says whether the applied plan rewrote an agent file.
		filesChanged bool
	}
)

// modelsLoadedMsg is the result of loading the registered CLIs and the rows.
type modelsLoadedMsg struct {
	owned
	seq   int
	hosts []string
	rows  []management.ModelRow
	err   error
}

// modelsPanel is the edit panel of one role.
type modelsPanel struct {
	role          string
	row           int // 0 Model, 1 Effort
	model         textinput.Model
	efforts       []string // effortReleaseText, then the values the CLI accepts; one entry when it takes none
	effort        int
	initialModel  string
	initialEffort int
	message       string
	messageOK     bool // message is information, not an error
}

func (p *modelsPanel) supportsEffort() bool { return len(p.efforts) > 1 }

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

	cur     int    // index of the selected role among the CLI's rows
	curRole string // the selected role's name, kept across reloads and CLI switches
	panel   *modelsPanel
	flow    int
	working string // what the edit flow is doing now, "" when idle
	message string // the last result, shown in the position row until a key is pressed
	msgErr  bool
}

func newModelsView(cfg appConfig) *modelsView {
	return &modelsView{cfg: cfg, loading: true, box: newScrollBox()}
}

func (v *modelsView) Init() tea.Cmd { return v.reload() }

func (v *modelsView) NeedsSpinner() bool { return v.loading || v.working != "" }

func (v *modelsView) TextFocused() bool { return v.panel != nil && v.panel.row == 0 }

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

// tableRows is the number of table rows the screen has room for.
func (v *modelsView) tableRows() int {
	rows := v.height - modelsFixedRows
	if v.panel != nil {
		rows -= modelsPanelRows
	}
	return max(rows, 1)
}

// layout redraws the table for the selected CLI and the current size, with the
// cursor on the selected role.
func (v *modelsView) layout() {
	v.box.vp.SetWidth(max(v.width, 1))
	v.box.vp.SetHeight(v.tableRows())
	rows := v.hostRows()
	v.cur = 0
	for i, r := range rows {
		if r.Role == v.curRole {
			v.cur = i
		}
	}
	if len(rows) > 0 {
		v.curRole = rows[v.cur].Role
	}
	v.render()
}

// render draws the table lines with the cursor mark and keeps the cursor in view.
func (v *modelsView) render() {
	rows := v.hostRows()
	_, lines := modelTable(rows, max(v.width-2, 1))
	for i := range lines {
		prefix := "  "
		if i == v.cur {
			prefix = "> "
		}
		lines[i] = prefix + lines[i]
	}
	v.box.setText(strings.Join(lines, "\n"))
	off, h := v.box.vp.YOffset(), v.box.vp.Height()
	switch {
	case v.cur < off:
		v.box.vp.SetYOffset(v.cur)
	case v.cur >= off+h:
		v.box.vp.SetYOffset(v.cur - h + 1)
	}
}

// moveCursor selects row i of the CLI's rows.
func (v *modelsView) moveCursor(i int) {
	rows := v.hostRows()
	if len(rows) == 0 {
		return
	}
	v.cur = min(max(i, 0), len(rows)-1)
	v.curRole = rows[v.cur].Role
	v.render()
}

// scroll applies a paging key to the table, then keeps the cursor inside the
// rows the screen shows; Home and End select the first and the last role.
func (v *modelsView) scroll(name string) {
	switch name {
	case "home":
		v.moveCursor(0)
		return
	case "end":
		v.moveCursor(len(v.hostRows()) - 1)
		return
	}
	v.box.handleKey(name)
	off, h := v.box.vp.YOffset(), v.box.vp.Height()
	v.moveCursor(min(max(v.cur, off), off+h-1))
}

// selectedRow is the row under the cursor.
func (v *modelsView) selectedRow() (management.ModelRow, bool) {
	rows := v.hostRows()
	if v.cur < 0 || v.cur >= len(rows) {
		return management.ModelRow{}, false
	}
	return rows[v.cur], true
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
			v.loadErr, v.hosts, v.rows = unreadableStateText(stateDirOf(v.cfg.Options), msg.err), nil, nil
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
	case modelsPlannedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onPlanned(msg)
	case modelsAppliedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onApplied(msg)
	case tea.KeyPressMsg:
		if v.panel != nil {
			return v.panelKey(msg)
		}
		return v.onKey(msg.String())
	default:
		if v.TextFocused() {
			var cmd tea.Cmd
			v.panel.model, cmd = v.panel.model.Update(msg)
			return cmd, action{nav: navNone}
		}
	}
	return nil, action{nav: navNone}
}

func (v *modelsView) onKey(name string) (tea.Cmd, action) {
	if v.working != "" {
		// Esc and Backspace stop waiting; the late result is dropped by the flow number.
		if name == "esc" || name == "backspace" {
			v.flow++
			v.working = ""
		}
		return nil, action{nav: navNone}
	}
	v.message = ""
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
	case "up":
		v.moveCursor(v.cur - 1)
	case "down":
		v.moveCursor(v.cur + 1)
	case "pgup", "pgdown", "home", "end":
		v.scroll(name)
	case "enter":
		v.openPanel()
	case "x":
		return v.reset()
	}
	return nil, action{nav: navNone}
}

// effectiveParts is the model without its OpenCode "#variant" and the effort,
// as the table shows them.
func effectiveParts(r management.ModelRow) (model, effort string) {
	model, effort = r.Model, r.Effort
	if r.Host == "opencode" {
		if base, variant, found := strings.Cut(model, "#"); found {
			model, effort = base, variant
		}
	}
	return model, effort
}

// effortChoices are "release default" and the efforts host accepts, asked of
// the same validator the commands use.
func effortChoices(host string) []string {
	choices := []string{effortReleaseText}
	for _, e := range []string{"low", "medium", "high", "xhigh", "max", "ultra"} {
		if agents.ValidateOverride(host, agents.ModelOverride{Effort: e}) == nil {
			choices = append(choices, e)
		}
	}
	return choices
}

// openPanel opens the edit panel on the selected role.
func (v *modelsView) openPanel() {
	row, ok := v.selectedRow()
	if !ok || v.loading {
		return
	}
	model, effort := effectiveParts(row)
	p := &modelsPanel{role: row.Role, model: textinput.New(), efforts: effortChoices(row.Host), initialModel: model}
	p.model.Prompt = ""
	p.model.SetStyles(textinput.Styles{}) // a static cursor until the first draw applies the theme
	p.model.SetValue(model)
	p.model.CursorEnd()
	for i, e := range p.efforts {
		if i > 0 && e == effort {
			p.effort = i
		}
	}
	p.initialEffort = p.effort
	v.panel = p
	v.layout()
	_ = p.model.Focus()
}

func (v *modelsView) closePanel() {
	v.panel = nil
	v.layout()
}

// panelKey handles a key while the edit panel is open.
func (v *modelsView) panelKey(msg tea.KeyPressMsg) (tea.Cmd, action) {
	p, name := v.panel, msg.String()
	if v.working != "" {
		if name == "esc" {
			v.flow++
			v.working = ""
		}
		return nil, action{nav: navNone}
	}
	switch name {
	case "esc":
		v.closePanel()
		return nil, action{nav: navNone}
	case "up":
		p.row = 0
		return p.model.Focus(), action{nav: navNone}
	case "down":
		if p.supportsEffort() {
			p.row = 1
			p.model.Blur()
		}
		return nil, action{nav: navNone}
	case "enter":
		return v.review()
	}
	if p.row == 0 {
		p.message = ""
		var cmd tea.Cmd
		p.model, cmd = p.model.Update(msg)
		return cmd, action{nav: navNone}
	}
	switch name {
	case "left", "right":
		delta := 1
		if name == "left" {
			delta = -1
		}
		p.message = ""
		p.effort = cycle(p.effort, delta, len(p.efforts))
	}
	// Backspace on Effort edits nothing and must not leave the view.
	return nil, action{nav: navNone}
}

// request is the change the panel asks for: only what differs from the values
// the table shows. On OpenCode a new model also sends the effort shown, as
// `hive models set --model x --effort <shown>` does, because a model override
// otherwise drops the variant the release's model carried.
func (p *modelsPanel) request(host string) modelsChange {
	var c modelsChange
	text := strings.TrimSpace(p.model.Value())
	if text != p.initialModel {
		if text == "" {
			c.dropModel = true
		} else {
			c.set.Model = text
		}
	}
	if p.supportsEffort() && p.effort != p.initialEffort {
		if p.effort == 0 {
			c.dropEffort = true
		} else {
			c.set.Effort = p.efforts[p.effort]
		}
	}
	if host == "opencode" && c.set.Model != "" && c.set.Effort == "" && p.effort != 0 {
		c.set.Effort = p.efforts[p.effort]
	}
	return c
}

// review builds the plan for the panel's request and shows its summary.
func (v *modelsView) review() (tea.Cmd, action) {
	p := v.panel
	p.message = ""
	change := p.request(v.host)
	if change.empty() {
		p.message, p.messageOK = "Nothing to change", true
		return nil, action{nav: navNone}
	}
	return v.plan(p.role, change, "Change "+p.role+" on "+v.host)
}

// reset opens the reset confirmation of the selected role when it has an override.
func (v *modelsView) reset() (tea.Cmd, action) {
	row, ok := v.selectedRow()
	if !ok || !row.Override || v.loading {
		return nil, action{nav: navNone}
	}
	return v.plan(row.Role, modelsChange{dropAll: true}, "Reset "+row.Role+" on "+v.host)
}

// plan builds the plan off the UI thread, from the overrides stored now, like
// the command does.
func (v *modelsView) plan(role string, change modelsChange, title string) (tea.Cmd, action) {
	v.flow++
	flow, host, o := v.flow, v.host, copyOptions(v.cfg.Options)
	v.working = "Planning the change…"
	return func() tea.Msg {
		msg := modelsPlannedMsg{owned: owned{v}, flow: flow, title: title}
		stored, err := management.StoredModelOverrides(o)
		if err != nil {
			msg.err = err
			return msg
		}
		plan, before, err := planModelsChange(o, host, change.applyTo(stored[host], role))
		if err != nil {
			msg.err = err
			return msg
		}
		if msg.unchanged, err = management.PlanUnchanged(plan); err != nil {
			msg.err = err
			return msg
		}
		var summary bytes.Buffer
		showModelsSummary(&summary, host, plan, before, stored[host], msg.unchanged)
		msg.plan, msg.summary = plan, summary.String()
		return msg
	}, action{nav: navNone}
}

// say shows a result where the user is: in the panel when it is open, in the
// position row otherwise.
func (v *modelsView) say(text string, ok bool) {
	if v.panel != nil {
		v.panel.message, v.panel.messageOK = text, ok
		return
	}
	v.message, v.msgErr = text, !ok
}

func (v *modelsView) onPlanned(msg modelsPlannedMsg) (tea.Cmd, action) {
	v.working = ""
	if msg.err != nil {
		v.say(sanitizeLine(msg.err.Error()), false)
		return nil, action{nav: navNone}
	}
	if msg.unchanged {
		v.say("Nothing to change", true)
		return nil, action{nav: navNone}
	}
	confirm := newConfirmView(confirmOptions{
		Title:   msg.title,
		Summary: msg.summary,
		OnApply: func() (tea.Cmd, action) {
			return applyModelsCmd(v, msg, v.cfg.Deps), action{nav: navNone, write: true, result: v}
		},
		OnCancel: func() (tea.Cmd, action) {
			v.say("Cancelled. No changes applied.", true)
			return nil, action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

func applyModelsCmd(owner view, planned modelsPlannedMsg, deps installDependencies) tea.Cmd {
	return func() tea.Msg {
		_, err := (management.Engine{}).Apply(planned.plan)
		msg := modelsAppliedMsg{owned: owned{owner}, flow: planned.flow, err: err, filesChanged: modelsFilesChange(planned.plan)}
		if err != nil {
			msg.pending = pendingAfterFailure(deps, planned.plan.StateDir)
		}
		return msg
	}
}

func (v *modelsView) onApplied(msg modelsAppliedMsg) (tea.Cmd, action) {
	if msg.err != nil {
		v.say(sanitizeLine(msg.err.Error()), false)
		if msg.pending != management.PendingNone {
			rv := newRecoveryView(msg.pending, copyOptions(v.cfg.Options), v.cfg.ExplicitStateDir, v.cfg.Deps)
			return nil, action{nav: navPush, pops: 1, push: rv}
		}
		return nil, action{nav: navPop}
	}
	v.panel = nil
	v.message, v.msgErr = "Updated. Open sessions keep the previous model until they restart.", false
	if !msg.filesChanged {
		v.message = "Updated. The setting is saved; no agent file changed."
	}
	return v.reload(), action{nav: navPop}
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
	if c.Width != v.width || c.Height != v.height {
		v.width, v.height = c.Width, c.Height
		v.layout()
	}
	title := th.Title.Render("Models")
	footer := []string{
		th.Muted.Render("Enter edits the selected role; x resets one marked *."),
		th.Muted.Render("Defaults come from integrations/agent-profiles.json in the release."),
	}
	if !v.loading && v.loadErr == "" && len(v.hosts) == 0 {
		footer = nil // with no registered CLI there is no model to change
	}
	var host, header, position string
	body := []string{}
	switch {
	case v.loading && len(v.rows) == 0 && len(v.hosts) == 0:
		body = append(body, c.Spinner+" "+th.Muted.Render("Loading models…"))
	case v.loadErr != "":
		// The plain words and the way out come first; the technical error
		// follows on its own lines, so a short screen cuts it before them.
		words, detail := errLines(v.loadErr)
		for _, l := range wrapLines("The models cannot be shown: "+words, c.Width) {
			body = append(body, th.Danger.Render(l))
		}
		body = append(body, th.Muted.Render("Press r to retry after fixing it."))
		for _, d := range detail {
			for _, l := range wrapLines(d, c.Width) {
				body = append(body, th.Muted.Render(l))
			}
		}
	case len(v.hosts) == 0:
		for _, l := range wrapLines(noHostsText, c.Width) {
			body = append(body, th.Text.Render(l))
		}
	default:
		host = v.hostRow(th)
		if rows := v.hostRows(); len(rows) == 0 {
			body = append(body, th.Text.Render("No agents installed for "+v.host))
		} else {
			h, _ := modelTable(rows, max(c.Width-2, 1))
			header = th.Muted.Render("  " + h)
			body = strings.Split(v.box.vp.View(), "\n")
			if v.box.scrollable() {
				position = th.Muted.Render(v.box.position())
			}
		}
	}
	switch {
	case v.working != "":
		position = c.Spinner + " " + th.Muted.Render(v.working)
	case c.Busy:
		position = c.Spinner + " " + th.Muted.Render("Applying…")
	case v.message != "":
		style := th.Text
		if v.msgErr {
			style = th.Danger
		}
		position = style.Render(truncateRunes(v.message, max(c.Width, 1)))
	}
	room := v.tableRows()
	if v.panel == nil {
		room = max(c.Height-modelsFixedRows, 1)
	}
	if len(body) > room {
		body = body[:room]
	}
	for len(body) < room {
		body = append(body, "")
	}
	lines := []string{title, host}
	if v.panel != nil {
		lines = append(lines, v.panelLines(c)...)
	}
	lines = append(lines, header)
	lines = append(lines, body...)
	lines = append(lines, position)
	lines = append(lines, footer...)
	return strings.Join(lines, "\n")
}

// panelLines draws the edit panel's modelsPanelRows rows.
func (v *modelsView) panelLines(c viewCtx) []string {
	th, p := c.Theme, v.panel
	applyInputStyles(&p.model, th)
	p.model.SetWidth(max(c.Width-2-modelsPanelLabel, 10))
	mark := func(on bool) string {
		if on {
			return "> "
		}
		return "  "
	}
	label := func(text string) string { return padRight(text, modelsPanelLabel) }
	lines := []string{
		th.Muted.Render(truncateRunes("  "+label("Role")+p.role+" on "+v.host, c.Width)),
		mark(p.row == 0) + label("Model") + inputView(p.model, th),
	}
	switch {
	case !p.supportsEffort():
		lines = append(lines, th.Muted.Render(truncateRunes("  "+label("Effort")+"not supported by "+v.host, c.Width)))
	case p.row == 1:
		lines = append(lines, th.Accent.Render("> "+label("Effort")+"< "+p.efforts[p.effort]+" >"))
	default:
		lines = append(lines, th.Text.Render("  "+label("Effort")+"< "+p.efforts[p.effort]+" >"))
	}
	message := ""
	if p.message != "" {
		style := th.Danger
		if p.messageOK {
			style = th.Text
		}
		message = style.Render(truncateRunes(p.message, max(c.Width, 1)))
	}
	return append(lines, message, "")
}

func (v *modelsView) Keys() []key.Binding {
	if v.panel != nil {
		return []key.Binding{
			binding("up,down", "↑/↓", "field"),
			binding("left,right", "←/→", "effort"),
			binding("enter", "enter", "review"),
			binding("esc", "esc", "close"),
		}
	}
	keys := []key.Binding{}
	if len(v.hostRows()) > 0 {
		keys = append(keys,
			binding("up,down", "↑/↓", "role"),
			binding("enter", "enter", "edit"),
			binding("x", "x", "reset"),
		)
	}
	return append(keys,
		binding("left,right", "←/→", "CLI"),
		binding("r", "r", "reload"),
		binding("esc", "esc", "back"),
	)
}
