// tui_models_view.go is the Models view (#46, design.md "Vistas" and "Edición
// en la vista Models"): the model and effort each installed agent role gets on
// each registered CLI, read from the release snapshot the role was installed
// from. A cursor selects a role; Enter opens a panel that edits its model and
// effort, and x resets a marked role. Roles are grouped by the folder of their
// source: a header row stands for its group, and Enter or x on it edits or
// resets every role of the group. The Model field is a list the CLI supplies
// (T9). Every change builds the plan `hive models set` or `reset` builds, shows
// its summary and confirms before writing.
package main

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
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

// modelsGroupMark opens a group header and modelsRoleIndent sets a grouped role
// under it, so the structure reads without color.
const (
	modelsGroupMark  = "▾ "
	modelsRoleIndent = "  "
)

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

// modelsPanel is the edit panel of one role, or of one group when group is set.
type modelsPanel struct {
	host      string
	role      string
	group     string                // set for a group panel
	groupRows []management.ModelRow // the roles of the group
	title     string                // the Role line
	row       int                   // 0 Model, 1 Effort

	// Model is a choice from a list until "Other…" turns it into a text field.
	model        textinput.Model
	textMode     bool
	chosen       string // the model picked from the list; "" is the release's model
	touched      bool   // a choice was made from the list
	initialModel string
	modelMixed   bool // a group whose roles have different models
	list         *modelList

	efforts       []string // effortReleaseText, then the values the CLI accepts; one entry when it takes none
	effort        int
	initialEffort string // the effort value shown when the panel opened; "" is the release default
	effortMixed   bool

	catalog    []catalogModel // the CLI's list, once loaded
	catalogErr string         // why the list is unavailable
	noCatalog  bool           // no list is ever queried here (a synthetic home)

	message   string
	messageOK bool // message is information, not an error
}

func (p *modelsPanel) supportsEffort() bool { return len(p.efforts) > 1 }

const (
	otherModelText  = "Other…"
	effortMixedText = "mixed"
)

// modelsItem is one line of the grouped table: a group header or a role.
type modelsItem struct {
	header bool
	blank  bool // the empty row that separates a group from the one before it
	group  string
	cells  modelCells            // the header's cells: the group's name and its common model and effort
	rows   []management.ModelRow // the group's roles, for a header
	row    management.ModelRow   // the role's row, for a role
}

// modelsKey identifies the item under the cursor across reloads and CLI
// switches. A header and a role never share a key.
type modelsKey struct {
	header bool
	name   string
}

func (k modelsKey) of(it modelsItem) bool {
	if it.blank {
		return false
	}
	if it.header {
		return k.header && k.name == it.group
	}
	return !k.header && k.name == it.row.Role
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

	items       []modelsItem // the selected CLI's table: headers and roles in order
	tableHeader string       // the column header line drawn for items
	cur         int          // index of the selected item
	curKey      modelsKey    // the selected item's key, kept across reloads and CLI switches
	panel       *modelsPanel
	flow        int
	working     string // what the edit flow is doing now, "" when idle
	message     string // the last result, shown in the position row until a key is pressed
	msgErr      bool
}

func newModelsView(cfg appConfig) *modelsView {
	if cfg.catalog == nil {
		cfg.catalog = newModelCatalogCache()
	}
	return &modelsView{cfg: cfg, loading: true, box: newScrollBox()}
}

func (v *modelsView) Init() tea.Cmd { return v.reload() }

// NeedsSpinner is true while the rows load, an edit works, or the open panel
// waits for its CLI's model list.
func (v *modelsView) NeedsSpinner() bool {
	return v.loading || v.working != "" || v.catalogLoading()
}

func (v *modelsView) catalogLoading() bool {
	return v.panel != nil && !v.panel.noCatalog && v.cfg.catalog.isLoading(v.host)
}

// TextFocused is true while a text field takes the keys: the model field after
// "Other…", and the open list's filter.
func (v *modelsView) TextFocused() bool {
	return v.panel != nil && (v.panel.list != nil || (v.panel.row == 0 && v.panel.textMode))
}

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

// hostRows are the rows of the selected CLI, ordered by group and then role.
func (v *modelsView) hostRows() []management.ModelRow {
	var out []management.ModelRow
	for _, r := range v.rows {
		if r.Host == v.host {
			out = append(out, r)
		}
	}
	return sortedByGroup(out)
}

// buildItems lays the CLI's rows out as headers and roles.
func buildItems(rows []management.ModelRow) []modelsItem {
	var items []modelsItem
	for i, r := range rows {
		if r.Group != "" && (i == 0 || rows[i-1].Group != r.Group) {
			end := i
			for end < len(rows) && rows[end].Group == r.Group {
				end++
			}
			model, effort, override := groupSummary(rows[i:end])
			name := modelsGroupMark + strings.ToUpper(r.Group[:1]) + r.Group[1:]
			if len(items) > 0 {
				items = append(items, modelsItem{blank: true})
			}
			if override {
				name += modelOverrideMarker
			}
			items = append(items, modelsItem{header: true, group: r.Group, cells: modelCells{role: name, model: model, effort: effort}, rows: rows[i:end]})
		}
		items = append(items, modelsItem{group: r.Group, row: r})
	}
	return items
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
// cursor on the selected item.
func (v *modelsView) layout() {
	v.box.vp.SetWidth(max(v.width, 1))
	v.box.vp.SetHeight(v.tableRows())
	v.items = buildItems(v.hostRows())
	v.cur = 0
	for i, it := range v.items {
		if v.curKey.of(it) {
			v.cur = i
		}
	}
	if len(v.items) > 0 {
		v.curKey = keyOf(v.items[v.cur])
	}
	v.render()
}

func keyOf(it modelsItem) modelsKey {
	if it.header {
		return modelsKey{header: true, name: it.group}
	}
	return modelsKey{name: it.row.Role}
}

// render draws the table lines with the cursor mark and keeps the cursor in
// view. Group headers sit in the same columns as the roles: the group's name
// under Role, and its common model and effort under Model and Effort.
func (v *modelsView) render() {
	cells := make([]modelCells, len(v.items))
	for i, it := range v.items {
		switch {
		case it.blank:
		case it.header:
			cells[i] = it.cells
		default:
			cells[i] = modelCellsFor(it.row)
			if it.group != "" {
				cells[i].role = modelsRoleIndent + cells[i].role
			}
		}
	}
	columns := layoutModelColumns(cells, max(v.width-2, 1))
	v.tableHeader = columns.header()
	out := make([]string, len(v.items))
	for i := range v.items {
		prefix := "  "
		if i == v.cur {
			prefix = "> "
		}
		if v.items[i].blank {
			continue
		}
		out[i] = prefix + columns.line(cells[i])
	}
	v.box.setText(strings.Join(out, "\n"))
	off, h := v.box.vp.YOffset(), v.box.vp.Height()
	switch {
	case v.cur < off:
		v.box.vp.SetYOffset(v.cur)
	case v.cur >= off+h:
		v.box.vp.SetYOffset(v.cur - h + 1)
	}
}

// moveCursor selects item i of the CLI's table. A blank separator row cannot be
// selected: the cursor goes to the next item, or to the previous one at the end.
func (v *modelsView) moveCursor(i int) {
	if len(v.items) == 0 {
		return
	}
	i = min(max(i, 0), len(v.items)-1)
	if v.items[i].blank {
		if i+1 < len(v.items) {
			i++
		} else {
			i--
		}
	}
	v.cur = i
	v.curKey = keyOf(v.items[v.cur])
	v.render()
}

// step moves the cursor one selectable item up or down.
func (v *modelsView) step(delta int) {
	i := v.cur + delta
	for i >= 0 && i < len(v.items) && v.items[i].blank {
		i += delta
	}
	if i < 0 || i >= len(v.items) {
		return
	}
	v.moveCursor(i)
}

// scroll applies a paging key to the table, then keeps the cursor inside the
// rows the screen shows; Home and End select the first and the last item.
func (v *modelsView) scroll(name string) {
	switch name {
	case "home":
		v.moveCursor(0)
		return
	case "end":
		v.moveCursor(len(v.items) - 1)
		return
	}
	v.box.handleKey(name)
	off, h := v.box.vp.YOffset(), v.box.vp.Height()
	v.moveCursor(min(max(v.cur, off), off+h-1))
}

// selectedItem is the item under the cursor.
func (v *modelsView) selectedItem() (modelsItem, bool) {
	if v.cur < 0 || v.cur >= len(v.items) {
		return modelsItem{}, false
	}
	return v.items[v.cur], true
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
	case catalogLoadedMsg:
		v.onCatalog(msg)
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
		if v.panel != nil && v.panel.row == 0 && v.panel.textMode {
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
		v.step(-1)
	case "down":
		v.step(1)
	case "pgup", "pgdown", "home", "end":
		v.scroll(name)
	case "enter":
		return v.openPanel(), action{nav: navNone}
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

// commonParts are the model and effort the rows share, and whether each part
// differs between them.
func commonParts(rows []management.ModelRow) (model, effort string, modelMixed, effortMixed bool) {
	for i, r := range rows {
		m, e := effectiveParts(r)
		if i == 0 {
			model, effort = m, e
			continue
		}
		modelMixed = modelMixed || m != model
		effortMixed = effortMixed || e != effort
	}
	return model, effort, modelMixed, effortMixed
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

// openPanel opens the edit panel on the selected role or group and starts the
// CLI's model query when it was never made (or failed the last time).
func (v *modelsView) openPanel() tea.Cmd {
	it, ok := v.selectedItem()
	if !ok || v.loading {
		return nil
	}
	rows := []management.ModelRow{it.row}
	p := &modelsPanel{host: v.host, role: it.row.Role, model: textinput.New()}
	title := "  " + p.role + " on " + v.host
	if it.header {
		rows = it.rows
		p.role, p.group, p.groupRows = "", it.group, it.rows
		noun := "roles"
		if len(rows) == 1 {
			noun = "role"
		}
		title = fmt.Sprintf("%s group on %s (%d %s)", strings.ToUpper(it.group[:1])+it.group[1:], v.host, len(rows), noun)
	} else {
		title = p.role + " on " + v.host
	}
	p.title = title
	model, effort, modelMixed, effortMixed := commonParts(rows)
	p.chosen, p.initialModel, p.modelMixed = model, model, modelMixed
	p.model.Prompt = ""
	p.model.SetStyles(textinput.Styles{}) // a static cursor until the first draw applies the theme
	p.efforts = effortChoices(v.host)
	if effortMixed && len(p.efforts) > 1 {
		p.efforts = append([]string{effortMixedText}, p.efforts...)
		p.effortMixed = true
		p.initialEffort = effortMixedText
		p.effort = 0
	} else {
		for i, e := range p.efforts {
			if i > 0 && e == effort {
				p.effort = i
			}
		}
		p.initialEffort = p.efforts[p.effort]
		if p.initialEffort == effortReleaseText {
			p.initialEffort = ""
		}
	}
	v.panel = p
	v.layout()
	return v.loadCatalog()
}

// loadCatalog gives the panel the CLI's list, and starts the query when it has
// not run. Under a synthetic home nothing is ever queried, so the list offers
// only the current value and "Other…".
func (v *modelsView) loadCatalog() tea.Cmd {
	p, host := v.panel, v.host
	if v.cfg.catalogRunner == nil && v.cfg.Options.Home != "" && host != "claude" {
		p.noCatalog = true
		return nil
	}
	if models, ok := v.cfg.catalog.get(host); ok {
		p.catalog = models
		p.trimEfforts(false)
		return nil
	}
	if !v.cfg.catalog.begin(host) {
		return nil // a query is already running; its result reaches the panel
	}
	run := v.cfg.catalogRunner
	if run == nil {
		run = catalogRunnerFor(v.cfg.Options)
	}
	cache := v.cfg.catalog
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), catalogTimeout+2*catalogWaitDelay)
		defer cancel()
		models, err := listHostModels(ctx, host, run)
		return catalogLoadedMsg{cache: cache, host: host, models: models, err: err}
	}
}

// onCatalog hands a finished query to the open panel of the same CLI. The
// cache was already written by appModel.Update.
func (v *modelsView) onCatalog(msg catalogLoadedMsg) {
	p := v.panel
	if p == nil || msg.host != v.host {
		return
	}
	if msg.err != nil {
		p.catalogErr = sanitizeLine(msg.err.Error())
		return
	}
	p.catalogErr = ""
	key := v.pickerEntryKey() // the highlighted entry, by what it is, not by its place
	p.catalog = msg.models
	p.trimEfforts(false)
	v.restorePickerEntry(key)
}

// modelID is the model the panel stands on: the one chosen, or the effective
// one while nothing was chosen. It is empty when no single model is known.
func (p *modelsPanel) modelID() string {
	if p.modelChanged() {
		return p.modelValue()
	}
	if p.modelMixed {
		return ""
	}
	return p.initialModel
}

// trimEfforts cuts the effort choices to the levels the panel's model lists
// (Codex lists them per model). A chosen effort the model does not list falls
// back to "release default" when the person just chose the model, and keeps the
// list uncut otherwise, so a list that arrives late or a panel that opens does
// not move an effort nobody changed.
func (p *modelsPanel) trimEfforts(fromChoice bool) {
	if !p.supportsEffort() {
		return
	}
	var allowed []string
	for _, m := range p.catalog {
		if id := p.modelID(); id != "" && m.ID == id {
			allowed = m.Efforts
		}
	}
	current := p.efforts[p.effort]
	base := []string{effortReleaseText}
	if p.effortMixed {
		base = []string{effortMixedText, effortReleaseText}
	}
	list := base
	for _, e := range p.untrimmed() {
		if contains(base, e) {
			continue
		}
		if len(allowed) == 0 || contains(allowed, e) {
			list = append(list, e)
		}
	}
	if len(allowed) == 0 {
		list = p.untrimmed()
	}
	i := indexOf(list, current)
	if i < 0 {
		if !fromChoice {
			return
		}
		i = indexOf(list, effortReleaseText)
	}
	p.efforts, p.effort = list, i
}

// untrimmed is the full effort list of the panel's CLI.
func (p *modelsPanel) untrimmed() []string {
	list := effortChoices(p.host)
	if p.effortMixed {
		list = append([]string{effortMixedText}, list...)
	}
	return list
}

func contains(list []string, s string) bool { return indexOf(list, s) >= 0 }

func indexOf(list []string, s string) int {
	for i, x := range list {
		if x == s {
			return i
		}
	}
	return -1
}

func (v *modelsView) closePanel() {
	v.panel = nil
	v.layout()
}

// modelValue is the model the panel asks for: the text of the field after
// "Other…", or the list's choice; "" means the release's model.
func (p *modelsPanel) modelValue() string {
	if p.textMode {
		return strings.TrimSpace(p.model.Value())
	}
	return p.chosen
}

// modelChanged says whether the person asked for another model.
func (p *modelsPanel) modelChanged() bool {
	switch {
	case p.modelMixed && p.textMode:
		return strings.TrimSpace(p.model.Value()) != ""
	case p.modelMixed:
		return p.touched
	}
	return p.modelValue() != p.initialModel
}

// effortValue is the chosen effort; "" is the release default.
func (p *modelsPanel) effortValue() string {
	e := p.efforts[p.effort]
	if e == effortReleaseText {
		return ""
	}
	return e
}

// effortChanged says whether the chosen effort differs, by value, from the
// effort the panel opened with.
func (p *modelsPanel) effortChanged() bool {
	if !p.supportsEffort() {
		return false
	}
	return p.efforts[p.effort] != p.openingEffort()
}

// openingEffort is the effort entry the panel opened on.
func (p *modelsPanel) openingEffort() string {
	if p.initialEffort == "" {
		return effortReleaseText
	}
	return p.initialEffort
}

// modelLabel is how the Model row reads while it is not a text field.
func (p *modelsPanel) modelLabel() string {
	switch {
	case p.modelMixed && !p.touched:
		return effortMixedText
	case p.chosen == "" && p.initialModel == "" && !p.touched:
		return modelHostDefault
	case p.chosen == "":
		return effortReleaseText
	}
	return p.chosen
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
	if p.list != nil {
		return v.listKey(msg)
	}
	switch name {
	case "esc":
		v.closePanel()
		return nil, action{nav: navNone}
	case "up":
		p.row = 0
		if p.textMode {
			return p.model.Focus(), action{nav: navNone}
		}
		return nil, action{nav: navNone}
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
		if p.textMode {
			var cmd tea.Cmd
			p.model, cmd = p.model.Update(msg)
			p.trimEfforts(false) // typing never moves an effort
			return cmd, action{nav: navNone}
		}
		// The Model field is a choice: → or a letter opens the list, and the
		// letter starts the filter. Backspace never leaves the view.
		switch {
		case name == "right":
			v.openPicker("")
		case msg.Text != "" && msg.Text != " ":
			v.openPicker(msg.Text)
		}
		return nil, action{nav: navNone}
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

// request is the change a role panel asks for: only what differs from the
// values the table shows. (On OpenCode, agents.Resolve keeps the effort a model
// override would otherwise drop, so nothing is added here.)
func (p *modelsPanel) request(host string) modelsChange {
	var c modelsChange
	if p.modelChanged() {
		if text := p.modelValue(); text == "" {
			c.dropModel = true
		} else {
			c.set.Model = text
		}
	}
	if p.effortChanged() {
		if v := p.effortValue(); v == "" {
			c.dropEffort = true
		} else {
			c.set.Effort = v
		}
	}
	return c
}

// groupRequest is the override every role of a group gets: only the parts the
// person chose. A part left untouched is kept only when every role of the group
// already holds it as its own override, and a part left "mixed" is not sent.
func (p *modelsPanel) groupRequest(stored map[string]management.ModelOverride) management.ModelOverride {
	var set management.ModelOverride
	every := func(has func(management.ModelOverride) bool) bool {
		for _, r := range p.groupRows {
			if !has(stored[r.Role]) {
				return false
			}
		}
		return true
	}
	switch {
	case p.modelChanged():
		set.Model = p.modelValue()
	case !p.modelMixed && p.initialModel != "" && every(func(o management.ModelOverride) bool { return o.Model == p.initialModel }):
		set.Model = p.initialModel
	}
	switch {
	case p.effortChanged():
		if v := p.effortValue(); v != effortMixedText {
			set.Effort = v
		}
	case !p.effortMixed && p.initialEffort != "" && every(func(o management.ModelOverride) bool { return o.Effort == p.initialEffort }):
		set.Effort = p.initialEffort
	}
	return set
}

// review builds the plan for the panel's request and shows its summary.
func (v *modelsView) review() (tea.Cmd, action) {
	p := v.panel
	p.message = ""
	host := v.host
	if p.group != "" {
		if !p.modelChanged() && !p.effortChanged() {
			p.message, p.messageOK = "Nothing to change", true
			return nil, action{nav: navNone}
		}
		rows := p.groupRows
		group := p.group
		return v.plan(func(stored map[string]management.ModelOverride) (map[string]management.ModelOverride, []string, string) {
			set := p.groupRequest(stored)
			next, replaced := setGroup(stored, rows, set)
			if set == (management.ModelOverride{}) {
				// Both parts go back to the release's values: this is a reset.
				return next, nil, modelsTitle("Reset", "", group, host, len(rows))
			}
			return next, replaced, ""
		}, modelsTitle("Change", "", group, host, len(rows)))
	}
	change := p.request(host)
	if change.empty() {
		p.message, p.messageOK = "Nothing to change", true
		return nil, action{nav: navNone}
	}
	role := p.role
	return v.plan(func(stored map[string]management.ModelOverride) (map[string]management.ModelOverride, []string, string) {
		return change.applyTo(stored, role), nil, ""
	}, modelsTitle("Change", role, "", host, 0))
}

// reset opens the reset confirmation of the selected role or group when it has
// an override.
func (v *modelsView) reset() (tea.Cmd, action) {
	it, ok := v.selectedItem()
	if !ok || v.loading {
		return nil, action{nav: navNone}
	}
	drop := modelsChange{dropAll: true}
	if it.header {
		marked := false
		for _, r := range it.rows {
			marked = marked || r.Override
		}
		if !marked {
			return nil, action{nav: navNone}
		}
		rows := it.rows
		return v.plan(func(stored map[string]management.ModelOverride) (map[string]management.ModelOverride, []string, string) {
			next := stored
			for _, r := range rows {
				next = drop.applyTo(next, r.Role)
			}
			return next, nil, ""
		}, modelsTitle("Reset", "", it.group, v.host, len(rows)))
	}
	if !it.row.Override {
		return nil, action{nav: navNone}
	}
	role := it.row.Role
	return v.plan(func(stored map[string]management.ModelOverride) (map[string]management.ModelOverride, []string, string) {
		return drop.applyTo(stored, role), nil, ""
	}, modelsTitle("Reset", role, "", v.host, 0))
}

// modelsBuild turns the overrides stored for a CLI into the set a change asks
// for, with the roles whose own override it replaces.
type modelsBuild func(stored map[string]management.ModelOverride) (next map[string]management.ModelOverride, replaced []string, title string)

// plan builds the plan off the UI thread, from the overrides stored now, like
// the command does.
func (v *modelsView) plan(build modelsBuild, title string) (tea.Cmd, action) {
	v.flow++
	flow, host, o, width := v.flow, v.host, copyOptions(v.cfg.Options), v.width
	v.working = "Planning the change…"
	return func() tea.Msg {
		msg := modelsPlannedMsg{owned: owned{v}, flow: flow, title: title}
		stored, err := management.StoredModelOverrides(o)
		if err != nil {
			msg.err = err
			return msg
		}
		next, replaced, retitle := build(stored[host])
		if retitle != "" {
			msg.title = retitle
		}
		plan, before, err := planModelsChange(o, host, next)
		if err != nil {
			msg.err = err
			return msg
		}
		if msg.unchanged, err = management.PlanUnchanged(plan); err != nil {
			msg.err = err
			return msg
		}
		var summary bytes.Buffer
		showModelsSummary(&summary, host, plan, before, stored[host], replaced, o.Home, 0, width, msg.unchanged)
		msg.plan, msg.summary = plan, summary.String()
		return msg
	}, action{nav: navNone}
}

// overridePrefix is the technical lead of a validation error: the panel already
// shows which role and CLI the error is about, so only the actionable part stays.
var overridePrefix = regexp.MustCompile(`^model override for \S+ \S+: `)

func panelErrorText(err error) string {
	return sanitizeLine(overridePrefix.ReplaceAllString(err.Error(), ""))
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
		v.say(panelErrorText(msg.err), false)
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
		th.Muted.Render("Enter edits the selected role or group; x resets one marked *."),
		th.Muted.Render("Defaults come from integrations/agent-profiles.json in the release."),
	}
	if !v.loading && v.loadErr == "" && len(v.hosts) == 0 {
		footer = nil // with no registered CLI there is no model to change
	}
	if v.panel != nil && v.panel.list != nil && v.loadErr == "" && len(v.hosts) > 0 && len(v.items) > 0 {
		// The picker is a box over the panel and the table.
		lines := append([]string{title, v.hostRow(th)}, v.pickerLines(c)...)
		return strings.Join(append(lines, footer...), "\n")
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
		switch {
		case len(v.items) == 0:
			body = append(body, th.Text.Render("No agents installed for "+v.host))
		default:
			header = th.Muted.Render("  " + v.tableHeader)
			body = strings.Split(v.box.vp.View(), "\n")
			for i, off := 0, v.box.vp.YOffset(); i < len(body); i++ {
				if k := off + i; k < len(v.items) && v.items[k].header {
					body[i] = th.Accent.Bold(true).Render(strings.TrimRight(body[i], " "))
				}
			}
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
	p.model.SetWidth(max(c.Width-2-modelsPanelLabel-1, 10)) // one column for the cursor
	mark := func(on bool) string {
		if on {
			return "> "
		}
		return "  "
	}
	label := func(text string) string { return padRight(text, modelsPanelLabel) }
	roleLine := th.Muted.Render(truncateRunes("  "+label("Role")+p.title, c.Width))
	var modelLine string
	if p.textMode {
		modelLine = mark(p.row == 0) + label("Model") + inputView(p.model, th)
	} else {
		// "▸" says the value opens a list; the hint shows while the row has the focus.
		text := truncateRunes(mark(p.row == 0)+label("Model")+p.modelLabel()+" ▸", c.Width)
		modelLine = th.Text.Render(text)
		if hint := "  → choose"; p.row == 0 && textWidth(text)+textWidth(hint) <= c.Width {
			modelLine += th.Muted.Render(hint)
		}
	}
	lines := []string{roleLine, modelLine}
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
	if v.panel != nil && v.panel.list != nil {
		return []key.Binding{
			binding("type", "type", "search"),
			binding("up,down", "↑/↓", "model"),
			binding("enter", "enter", "choose"),
			binding("esc", "esc", "close"),
		}
	}
	if p := v.panel; p != nil && p.textMode {
		if p.row == 0 { // typing: ←/→ move the text cursor, so no effort or list keys
			return []key.Binding{
				binding("type", "type", "model"),
				binding("up,down", "↑/↓", "field"),
				binding("enter", "enter", "review"),
				binding("esc", "esc", "close"),
			}
		}
		return []key.Binding{
			binding("up,down", "↑/↓", "field"),
			binding("left,right", "←/→", "effort"),
			binding("enter", "enter", "review"),
			binding("esc", "esc", "close"),
		}
	}
	if v.panel != nil {
		return []key.Binding{
			binding("up,down", "↑/↓", "field"),
			binding("right", "→", "choose model"),
			binding("left,right", "←/→", "effort"),
			binding("enter", "enter", "review"),
			binding("esc", "esc", "close"),
		}
	}
	keys := []key.Binding{}
	if len(v.items) > 0 {
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
