// tui_project_view.go is the Project view (design.md "Vistas"): the result of
// validating the `## Hive` section of the current repository's AGENTS.md. Two
// rows are fixed (heading, position); the AGENTS.md path, the verdict or
// findings and the values read scroll between them, so a long path wraps
// instead of being cut. Reading is the default; `e` opens the form that edits
// the section (design.md "Formulario en la vista Project"). The load runs as a
// Cmd whose result is addressed to the view (owned) and carries a sequence
// number, so a double reload or a result that arrives after the view was left
// is dropped.
package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// projectFixedRows are the heading and the position row.
const projectFixedRows = 2

type projectLoadedMsg struct {
	owned
	seq   int
	check projectCheck
}

// Messages of the form flow. flow lets a result that arrives after the form
// was closed, or after another request, be dropped.
type (
	projectFormOpenedMsg struct {
		owned
		flow        int
		items       []hiveItem
		suggestions map[string]string
		err         error
	}
	projectReviewedMsg struct {
		owned
		flow      int
		edit      projectEdit
		summary   string
		unchanged bool
		err       error
	}
	projectWrittenMsg struct {
		owned
		flow int
		err  error
	}
)

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

	form    *projectForm // nil while the form is closed
	flow    int
	working string // what the form flow is doing now, "" when idle
	notice  string // why the form did not open
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

func (v *projectView) TextFocused() bool { return v.form != nil && v.form.isText(v.form.row) }

func (v *projectView) NeedsSpinner() bool { return v.loading || v.working != "" }

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
		v.box.setText(projectBoxText(msg.check))
	case projectFormOpenedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onFormOpened(msg)
	case projectReviewedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onReviewed(msg)
	case projectWrittenMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onWritten(msg)
	case tea.KeyPressMsg:
		if v.form != nil {
			return v.formKey(msg)
		}
		switch name := msg.String(); name {
		case "esc", "backspace":
			return nil, action{}
		case "r":
			if !v.loading {
				v.notice = ""
				return v.reload(), action{nav: navNone}
			}
		case "e":
			if v.canEdit() {
				return v.openForm()
			}
		default:
			v.box.handleKey(name)
		}
	default:
		if v.TextFocused() {
			var cmd tea.Cmd
			f := v.form
			f.inputs[f.row], cmd = f.inputs[f.row].Update(msg)
			return cmd, action{nav: navNone}
		}
	}
	return nil, action{nav: navNone}
}

// projectBoxText is the scrolling text: the location line first, as hive
// doctor prints it, then the reason the check could not finish when it could
// not, then the section's lines.
func projectBoxText(c projectCheck) string {
	text := ""
	if label, path := c.location(); path != "" {
		text = label + path + "\n"
	}
	if c.Section.Err != "" {
		first, rest := errLines(c.Section.Err)
		text += "Could not check everything: " + first + "\n"
		for _, l := range rest {
			text += "  " + l + "\n"
		}
	}
	for _, l := range c.Section.Lines {
		text += l + "\n"
	}
	return text
}

func (v *projectView) Resize(w, h int) { v.w, v.h = w, h; v.layout() }

// layout gives the scrolling text the area between the heading and the
// position row.
func (v *projectView) layout() {
	v.box.vp.SetWidth(v.w)
	v.box.vp.SetHeight(max(v.h-projectFixedRows, 1))
	v.box.rewrap()
}

func (v *projectView) View(c viewCtx) string {
	if c.Width != v.w || c.Height != v.h {
		v.w, v.h = c.Width, c.Height
		v.layout()
	}
	th := c.Theme
	if v.form != nil {
		return v.formView(c)
	}
	position := ""
	switch {
	case v.notice != "":
		position = th.Danger.Render(truncateRunes(v.notice, max(c.Width, 1)))
	case v.loading && !v.loaded:
		return th.Title.Render("Project") + "\n" + c.Spinner + " " + th.Muted.Render("Checking the ## Hive section…")
	case v.loading:
		position = c.Spinner + " " + th.Muted.Render("Refreshing…")
	case v.failed:
		position = th.Danger.Render("The check failed") + th.Muted.Render(" · r to retry")
	case v.box.scrollable():
		position = th.Muted.Render(v.box.position())
	}
	return th.Title.Render("Project") + "\n" + v.box.vp.View() + "\n" + position
}

func (v *projectView) Keys() []key.Binding {
	if v.form != nil {
		return []key.Binding{
			binding("up,down", "↑/↓", "field"),
			binding("left,right", "←/→", "choice"),
			binding("enter", "enter", "review"),
			binding("esc", "esc", "close"),
		}
	}
	keys := []key.Binding{scrollBinding, binding("r", "r", "reload")}
	if v.canEdit() {
		keys = append(keys, binding("e", "e", "edit"))
	}
	return append(keys, binding("esc", "esc", "back"))
}

// canEdit reports whether the form can open: the check finished inside a Git
// repository.
func (v *projectView) canEdit() bool {
	return v.loaded && !v.loading && v.check.Root != "" && v.working == ""
}

// ---------------------------------------------------------------------------
// The form.
// ---------------------------------------------------------------------------

const (
	projectLabelCol  = 16 // two for the cursor mark, thirteen for "Hive guidance", one space
	projectTagText   = "  suggested"
	projectFormFixed = 1 // the heading
)

// projectForm holds one field per setting of hiveSettingKeys. Text settings
// are text inputs; Delivery and Hive guidance choose among ‹ none › and
// their one valid value (plus the value the file holds now, when it is
// neither, so opening the form never drops it silently).
type projectForm struct {
	inputs    []textinput.Model
	options   [][]string // for a choice field; nil for a text field
	choice    []int
	suggested map[string]string // key → suggested value, until the user changes it
	current   map[string]string // values the section holds now, by key
	row       int
	message   string
	messageOK bool // message is information, not an error
}

func (f *projectForm) isText(i int) bool { return f.options[i] == nil }

// value is the trimmed value of field i; "" for none.
func (f *projectForm) value(i int) string {
	if f.isText(i) {
		return strings.TrimSpace(f.inputs[i].Value())
	}
	return f.options[i][f.choice[i]]
}

// newProjectForm builds the form from the items the section holds now and the
// suggestions. A suggestion fills only a text setting the section lacks.
func newProjectForm(items []hiveItem, suggestions map[string]string) *projectForm {
	n := len(hiveSettingKeys)
	f := &projectForm{
		inputs: make([]textinput.Model, n), options: make([][]string, n), choice: make([]int, n),
		suggested: map[string]string{}, current: map[string]string{},
	}
	for _, it := range items {
		if _, dup := f.current[it.Key]; !dup {
			f.current[it.Key] = it.Value
		}
	}
	for i, k := range hiveSettingKeys {
		cur, present := f.current[k.Name]
		switch k.Name {
		case "Delivery", "Hive guidance":
			want := hiveDeliveryWant
			if k.Name == "Hive guidance" {
				want = hiveGuidanceWant
			}
			f.options[i] = []string{"", want}
			if present && cur != "" && cur != want {
				f.options[i] = append(f.options[i], cur)
			}
			f.choice[i] = max(slices.Index(f.options[i], cur), 0)
		default:
			in := textinput.New()
			in.Prompt = ""
			value := cur
			if s, ok := suggestions[k.Name]; ok && !present {
				value = s
				f.suggested[k.Name] = s
			}
			in.SetValue(value)
			in.CursorEnd()
			f.inputs[i] = in
		}
	}
	return f
}

// sync keeps exactly the current text field focused.
func (f *projectForm) sync() tea.Cmd {
	var cmd tea.Cmd
	for i := range f.inputs {
		if i == f.row && f.isText(i) {
			cmd = f.inputs[i].Focus()
		} else if f.isText(i) {
			f.inputs[i].Blur()
		}
	}
	return cmd
}

func (f *projectForm) resize(w int) {
	for i := range f.inputs {
		if f.isText(i) {
			f.inputs[i].SetWidth(max(w-projectLabelCol-len(projectTagText), 10))
		}
	}
}

// openForm reads the file and the suggestions off the UI thread.
func (v *projectView) openForm() (tea.Cmd, action) {
	v.flow++
	v.notice = ""
	v.working = "Opening the form…"
	flow, root, git := v.flow, v.check.Root, v.git
	return func() tea.Msg {
		msg := projectFormOpenedMsg{owned: owned{v}, flow: flow}
		data, _, _, err := readProjectFile(filepath.Join(root, projectFileName))
		if err != nil {
			msg.err = err
			return msg
		}
		w := walkHiveSection(string(data))
		if w.Headings > 1 {
			msg.err = errors.New(hiveHeading + ": section appears " + strconv.Itoa(w.Headings) + " times; keep one")
			return msg
		}
		msg.items = w.Items
		msg.suggestions = suggestHiveValues(root, git)
		return msg
	}, action{nav: navNone}
}

func (v *projectView) onFormOpened(msg projectFormOpenedMsg) (tea.Cmd, action) {
	v.working = ""
	if msg.err != nil {
		v.notice = sanitizeLine(msg.err.Error())
		return nil, action{nav: navNone}
	}
	v.form = newProjectForm(msg.items, msg.suggestions)
	v.form.resize(v.w)
	return v.form.sync(), action{nav: navNone}
}

// formKey handles a key while the form is open.
func (v *projectView) formKey(msg tea.KeyPressMsg) (tea.Cmd, action) {
	f := v.form
	name := msg.String()
	if v.working != "" {
		// Esc cancels the pending request and closes nothing else; the late
		// result is dropped by the flow number.
		if name == "esc" {
			v.flow++
			v.working = ""
		}
		return nil, action{nav: navNone}
	}
	switch name {
	case "esc":
		v.form = nil
		return nil, action{nav: navNone}
	case "up":
		f.row = max(f.row-1, 0)
		return f.sync(), action{nav: navNone}
	case "down":
		f.row = min(f.row+1, len(f.inputs)-1)
		return f.sync(), action{nav: navNone}
	case "enter":
		return v.review()
	}
	if f.isText(f.row) {
		f.message = ""
		var cmd tea.Cmd
		f.inputs[f.row], cmd = f.inputs[f.row].Update(msg)
		return cmd, action{nav: navNone}
	}
	switch name {
	case "left", "right":
		delta := 1
		if name == "left" {
			delta = -1
		}
		f.message = ""
		f.choice[f.row] = cycle(f.choice[f.row], delta, len(f.options[f.row]))
	}
	// Backspace on a choice field edits nothing and must not leave the view.
	return nil, action{nav: navNone}
}

// request turns the form into what `hive project set` takes: the settings that
// differ from the section now, and the optional ones the user emptied.
func (f *projectForm) request() (set []hiveItem, unset []string) {
	for i, k := range hiveSettingKeys {
		value := f.value(i)
		cur, present := f.current[k.Name]
		switch {
		case value != "" && (!present || value != cur):
			set = append(set, hiveItem{Key: k.Name, Value: value})
		case value == "" && present && cur != "" && !k.Required:
			unset = append(unset, k.Name)
		}
	}
	return set, unset
}

// review checks the required fields, then builds the same edit the command
// builds and shows its summary.
func (v *projectView) review() (tea.Cmd, action) {
	f := v.form
	f.message = ""
	for i, k := range hiveSettingKeys {
		if k.Required && f.value(i) == "" {
			f.row = i
			f.message, f.messageOK = k.Name+" is required", false
			return f.sync(), action{nav: navNone}
		}
	}
	set, unset := f.request()
	v.flow++
	flow, root, git := v.flow, v.check.Root, v.git
	v.working = "Preparing the change…"
	return func() tea.Msg {
		msg := projectReviewedMsg{owned: owned{v}, flow: flow}
		edit, err := prepareProjectEdit(root, set, unset, git)
		if err != nil {
			msg.err = err
			return msg
		}
		msg.edit = edit
		msg.unchanged = edit.Existed && bytes.Equal(edit.Before, edit.After)
		var summary bytes.Buffer
		showProjectSummary(&summary, edit, msg.unchanged)
		msg.summary = summary.String()
		return msg
	}, action{nav: navNone}
}

func (v *projectView) onReviewed(msg projectReviewedMsg) (tea.Cmd, action) {
	v.working = ""
	f := v.form
	if f == nil {
		return nil, action{nav: navNone}
	}
	if msg.err != nil {
		f.message, f.messageOK = sanitizeLine(msg.err.Error()), false
		return nil, action{nav: navNone}
	}
	if msg.unchanged {
		f.message, f.messageOK = "Nothing to change", true
		return nil, action{nav: navNone}
	}
	confirm := newConfirmView(confirmOptions{
		Title:   "Write AGENTS.md",
		Summary: msg.summary,
		Hanging: true,
		OnApply: func() (tea.Cmd, action) {
			flow, edit := v.flow, msg.edit
			return func() tea.Msg {
				err := writeProjectFile(edit.File, edit.Before, edit.Existed, edit.After)
				return projectWrittenMsg{owned: owned{v}, flow: flow, err: err}
			}, action{nav: navNone, write: true, result: v}
		},
		OnCancel: func() (tea.Cmd, action) {
			f.message, f.messageOK = "Cancelled. No changes applied.", true
			return nil, action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

func (v *projectView) onWritten(msg projectWrittenMsg) (tea.Cmd, action) {
	if msg.err != nil {
		if v.form != nil {
			v.form.message, v.form.messageOK = sanitizeLine(msg.err.Error()), false
		}
		return nil, action{nav: navPop}
	}
	v.form = nil
	return v.reload(), action{nav: navPop}
}

func (v *projectView) formView(c viewCtx) string {
	th, f := c.Theme, v.form
	f.resize(c.Width)
	lines := []string{th.Title.Render("Project")}
	if v.check.File != "" {
		for _, l := range wrapHanging(sanitizeLine(v.check.File), c.Width) {
			lines = append(lines, th.Muted.Render(l))
		}
	}
	lines = append(lines, "")
	for i, k := range hiveSettingKeys {
		applyInputStyles(&f.inputs[i], th)
		prefix := "  "
		if i == f.row {
			prefix = "> "
		}
		label := prefix + padRight(k.Name, projectLabelCol-2)
		style := th.Text
		if i == f.row {
			style = th.Accent
		}
		if !f.isText(i) {
			text := f.options[i][f.choice[i]]
			if text == "" {
				text = "none"
			}
			lines = append(lines, style.Render(label+"‹ "+truncateRunes(sanitizeLine(text), max(c.Width-len(label)-4, 1))+" ›"))
			continue
		}
		row := style.Render(label) + inputView(f.inputs[i], th)
		if s, ok := f.suggested[k.Name]; ok && f.value(i) == s {
			row += th.Muted.Render(projectTagText)
		}
		lines = append(lines, row)
	}
	lines = append(lines, "")
	switch {
	case v.working != "":
		lines = append(lines, c.Spinner+" "+th.Muted.Render(v.working))
	case f.message != "":
		style := th.Danger
		if f.messageOK {
			style = th.Text
		}
		lines = append(lines, style.Render(truncateRunes(f.message, max(c.Width, 1))))
	}
	return strings.Join(lines, "\n")
}
