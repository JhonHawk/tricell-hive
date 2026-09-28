// tui_update_view.go is the Update view (T9; design.md "Vistas"): the source
// and revision fields in one view, then the summary and confirmation of `hive
// update`. The commit resolves in planUpdate, the function updateWith shares,
// so the summary and the result equal the command's.
package main

import (
	"bytes"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
)

// applyInputStyles gives a text input the theme's colors. The cursor is a
// static reversed block: it needs no color and no blink ticks.
func applyInputStyles(in *textinput.Model, th *appTheme) {
	st := textinput.Styles{}
	st.Focused = textinput.StyleState{Text: th.Text, Placeholder: th.Muted, Prompt: th.Muted}
	st.Blurred = st.Focused
	in.SetStyles(st)
}

type (
	updatePlannedMsg struct {
		owned
		flow       int
		plan       management.Plan
		summary    string
		unchanged  bool
		sourceLine string
		err        error
	}
	updateAppliedMsg struct {
		owned
		flow    int
		text    string
		err     error
		dialog  bool // a confirmation view sits above the Update view
		pending management.PendingKind
	}
)

type updateView struct {
	cfg        appConfig
	source     textinput.Model
	rev        textinput.Model
	focus      int // 0 source, 1 revision
	planning   string
	message    string
	messageErr bool
	flow       int
}

func newUpdateView(cfg appConfig) *updateView {
	v := &updateView{cfg: cfg, source: textinput.New(), rev: textinput.New()}
	for _, in := range []*textinput.Model{&v.source, &v.rev} {
		in.Prompt = ""
	}
	v.source.SetValue(".")
	v.rev.SetValue("HEAD")
	return v
}

func (v *updateView) Init() tea.Cmd { return v.source.Focus() }

func (v *updateView) Resize(w, _ int) {
	v.source.SetWidth(max(w-14, 10))
	v.rev.SetWidth(max(w-14, 10))
}

func (v *updateView) TextFocused() bool { return true }

func (v *updateView) NeedsSpinner() bool { return v.planning != "" }

func (v *updateView) setFocus(i int) tea.Cmd {
	v.focus = i
	if i == 0 {
		v.rev.Blur()
		return v.source.Focus()
	}
	v.source.Blur()
	return v.rev.Focus()
}

func (v *updateView) input() *textinput.Model {
	if v.focus == 0 {
		return &v.source
	}
	return &v.rev
}

func (v *updateView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case updatePlannedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onPlanned(msg)
	case updateAppliedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onApplied(msg)
	case tea.KeyPressMsg:
		return v.onKey(msg)
	default:
		if v.planning == "" {
			cmd := v.forwardToInput(msg)
			return cmd, action{nav: navNone}
		}
	}
	return nil, action{nav: navNone}
}

func (v *updateView) forwardToInput(msg tea.Msg) tea.Cmd {
	in := v.input()
	var cmd tea.Cmd
	*in, cmd = in.Update(msg)
	return cmd
}

func (v *updateView) onKey(msg tea.KeyPressMsg) (tea.Cmd, action) {
	name := msg.String()
	if v.planning != "" {
		if name == "esc" {
			v.flow++ // leaving abandons the resolve; its late result is ignored
			v.planning = ""
			return nil, action{}
		}
		return nil, action{nav: navNone}
	}
	switch name {
	case "esc":
		return nil, action{}
	case "up":
		return v.setFocus(0), action{nav: navNone}
	case "down":
		return v.setFocus(1), action{nav: navNone}
	case "tab":
		return v.setFocus(1 - v.focus), action{nav: navNone}
	case "shift+tab":
		return v.setFocus(1 - v.focus), action{nav: navNone}
	case "enter":
		return v.resolve(), action{nav: navNone}
	}
	return v.forwardToInput(msg), action{nav: navNone}
}

// resolve runs planUpdate as a Cmd: Git and the extraction can take a while.
func (v *updateView) resolve() tea.Cmd {
	v.flow++
	id := v.flow
	f := updateFlags{Source: v.source.Value(), Rev: v.rev.Value(), Home: v.cfg.Options.Home, StateDir: v.cfg.Options.StateDir}
	v.message = ""
	v.planning = "Resolving " + f.Rev + "…"
	return func() tea.Msg {
		msg := updatePlannedMsg{owned: owned{v}, flow: id}
		plan, line, unchanged, err := planUpdate(f)
		if err != nil {
			msg.err = err
			return msg
		}
		var summary bytes.Buffer
		showInstallSummary(&summary, plan, onboardingPreview{}, false, unchanged, false)
		summary.WriteString(line + "\n")
		msg.plan, msg.summary, msg.unchanged, msg.sourceLine = plan, summary.String(), unchanged, line
		return msg
	}
}

func (v *updateView) fail(text string) (tea.Cmd, action) {
	v.planning = ""
	v.message, v.messageErr = text, true
	return nil, action{nav: navNone}
}

func (v *updateView) onPlanned(msg updatePlannedMsg) (tea.Cmd, action) {
	v.planning = ""
	if msg.err != nil {
		return v.fail(msg.err.Error())
	}
	if msg.unchanged {
		// Nothing changes: apply directly, as the command does, so the source
		// commit is still recorded.
		return applyUpdateCmd(v, msg, "Hive is already up to date", false, v.cfg.Deps), action{nav: navNone, write: true, result: v}
	}
	confirm := newConfirmView(confirmOptions{
		Title:   "Update Hive",
		Summary: msg.summary,
		OnApply: func() (tea.Cmd, action) {
			return applyUpdateCmd(v, msg, "Hive updated", true, v.cfg.Deps), action{nav: navNone, write: true, result: v}
		},
		OnCancel: func() (tea.Cmd, action) {
			v.message, v.messageErr = "Cancelled. No changes applied.", false
			return nil, action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

func applyUpdateCmd(owner view, planned updatePlannedMsg, verb string, dialog bool, deps installDependencies) tea.Cmd {
	return func() tea.Msg {
		result, err := (management.Engine{}).Apply(planned.plan)
		msg := updateAppliedMsg{owned: owned{owner}, flow: planned.flow, err: err, dialog: dialog}
		if err != nil {
			msg.pending = pendingAfterFailure(deps, planned.plan.StateDir)
			return msg
		}
		var text bytes.Buffer
		reportApplyResult(&text, verb, result)
		msg.text = strings.TrimRight(text.String(), "\n")
		return msg
	}
}

func (v *updateView) onApplied(msg updateAppliedMsg) (tea.Cmd, action) {
	pops := 0
	if msg.dialog {
		pops = 1
	}
	if msg.err != nil {
		v.message, v.messageErr = msg.err.Error(), true
		if msg.pending != management.PendingNone {
			rv := newRecoveryView(msg.pending, copyOptions(v.cfg.Options), v.cfg.ExplicitStateDir, v.cfg.Deps)
			return nil, action{nav: navPush, pops: pops, push: rv}
		}
	} else {
		v.message, v.messageErr = msg.text, false
	}
	if pops > 0 {
		return nil, action{nav: navPop}
	}
	return nil, action{nav: navNone}
}

func (v *updateView) View(c viewCtx) string {
	th := c.Theme
	applyInputStyles(&v.source, th)
	applyInputStyles(&v.rev, th)
	row := func(i int, label string, in textinput.Model) string {
		prefix := "  "
		if i == v.focus {
			prefix = "> "
		}
		return prefix + padRight(label, 10) + in.View()
	}
	lines := []string{
		th.Title.Render("Update"),
		"",
		row(0, "Source", v.source),
		row(1, "Revision", v.rev),
		"",
	}
	for _, l := range wrapLines("Enter resolves the revision in a Git checkout and shows what would change.", c.Width) {
		lines = append(lines, th.Muted.Render(l))
	}
	switch {
	case v.planning != "":
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render(v.planning))
	case c.Busy:
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render("Applying…"))
	case v.message != "":
		style := th.Text
		if v.messageErr {
			style = th.Danger
		}
		lines = append(lines, "")
		room := max(c.Height-len(lines), 1)
		wrapped := wrapLines(v.message, c.Width)
		if len(wrapped) > room {
			wrapped = append(wrapped[:room-1], "…")
		}
		for _, l := range wrapped {
			lines = append(lines, style.Render(l))
		}
	}
	return strings.Join(lines, "\n")
}

func (v *updateView) Keys() []key.Binding {
	return []key.Binding{
		binding("up,down,tab", "↑/↓", "field"),
		binding("enter", "enter", "resolve"),
		binding("esc", "esc", "back"),
	}
}
