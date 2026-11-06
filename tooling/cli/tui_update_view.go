// The Update view shares release and committed-source preparation with the CLI.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/version"

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
		// flags, adapter and offers carry the optional capabilities update
		// can offer for plan; offers is empty once the choice is made.
		release *releasePreparation
		flags   updateFlags
		adapter onboardingAdapter
		offers  []providerOffer
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

// The worker publishes its result before closing done. Shutdown waits on done
// before reading or cleaning that result, including a reply not yet delivered
// to the view. No lifecycle state is shared without that channel boundary.
type updatePreparationTask struct {
	cancel context.CancelFunc
	done   chan struct{}
	result *releasePreparation
}

type updateView struct {
	releaseTask       *updatePreparationTask
	cancelPreparation context.CancelFunc
	cancelRequested   bool
	pendingRelease    *management.UpdateJournal
	cfg               appConfig
	source            textinput.Model
	rev               textinput.Model
	releaseMode       bool
	releaseVersion    textinput.Model
	focus             int // -1 mode, 0 source/version, 1 source revision
	planning          string
	message           string
	messageErr        bool
	flow              int
}

func newUpdateView(cfg appConfig) *updateView {
	v := &updateView{cfg: cfg, source: textinput.New(), rev: textinput.New(), releaseVersion: textinput.New(), releaseMode: version.Current != "dev"}
	for _, in := range []*textinput.Model{&v.source, &v.rev, &v.releaseVersion} {
		in.Prompt = ""
	}
	v.releaseVersion.Placeholder = "latest stable"
	if v.releaseMode {
		v.focus = -1
	}
	v.source.SetValue(".")
	v.rev.SetValue("HEAD")
	return v
}

func (v *updateView) Init() tea.Cmd {
	if v.releaseMode {
		return nil
	}
	return v.source.Focus()
}

func (v *updateView) Resize(w, _ int) {
	v.releaseVersion.SetWidth(max(w-14, 10))
	v.source.SetWidth(max(w-14, 10))
	v.rev.SetWidth(max(w-14, 10))
}

func (v *updateView) TextFocused() bool { return v.focus != -1 }

func (v *updateView) NeedsSpinner() bool { return v.planning != "" }

func (v *updateView) setFocus(i int) tea.Cmd {
	v.focus = i
	v.releaseVersion.Blur()
	if i == -1 {
		v.source.Blur()
		v.rev.Blur()
		return nil
	}
	if v.releaseMode {
		v.source.Blur()
		v.rev.Blur()
		return v.releaseVersion.Focus()
	}
	if i == 0 {
		v.rev.Blur()
		return v.source.Focus()
	}
	v.source.Blur()
	return v.rev.Focus()
}

func (v *updateView) input() *textinput.Model {
	if v.releaseMode {
		return &v.releaseVersion
	}
	if v.focus == 0 {
		return &v.source
	}
	return &v.rev
}

func (v *updateView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case updatePlannedMsg:
		if msg.flow != v.flow {
			if msg.release != nil {
				_ = msg.release.packageFiles.Cleanup()
			}
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
	if v.focus == -1 {
		return nil
	}
	in := v.input()
	var cmd tea.Cmd
	*in, cmd = in.Update(msg)
	return cmd
}

func (v *updateView) onKey(msg tea.KeyPressMsg) (tea.Cmd, action) {
	name := msg.String()
	if v.planning != "" {
		if name == "esc" {
			if v.cancelPreparation != nil {
				v.cancelPreparation()
				v.cancelRequested = true
				v.planning = "Cancelling preparation…"
				return nil, action{nav: navNone}
			}
			v.flow++ // leaving abandons the resolve; its late result is ignored
			v.planning = ""
			return nil, action{}
		}
		return nil, action{nav: navNone}
	}
	switch name {
	case "esc":
		return nil, action{}
	case "left", "right", "space":
		if v.focus == -1 {
			v.releaseMode = !v.releaseMode
			v.message = ""
			return v.setFocus(-1), action{nav: navNone}
		}
	case "up":
		if v.focus == 1 {
			return v.setFocus(0), action{nav: navNone}
		}
		return v.setFocus(-1), action{nav: navNone}
	case "down":
		if v.focus == -1 {
			return v.setFocus(0), action{nav: navNone}
		}
		if v.releaseMode || v.focus == 1 {
			return v.setFocus(-1), action{nav: navNone}
		}
		return v.setFocus(1), action{nav: navNone}
	case "tab":
		if v.focus == -1 {
			return v.setFocus(0), action{nav: navNone}
		}
		if v.releaseMode || v.focus == 1 {
			return v.setFocus(-1), action{nav: navNone}
		}
		return v.setFocus(1), action{nav: navNone}
	case "shift+tab":
		if v.focus == -1 {
			if v.releaseMode {
				return v.setFocus(0), action{nav: navNone}
			}
			return v.setFocus(1), action{nav: navNone}
		}
		return v.setFocus(v.focus - 1), action{nav: navNone}
	case "enter":
		if v.pendingRelease != nil {
			return recoverReleaseViewCmd(v, *v.pendingRelease), action{nav: navNone, write: true, result: v}
		}
		return v.resolve(), action{nav: navNone}
	}
	return v.forwardToInput(msg), action{nav: navNone}
}

// resolve runs planUpdate as a Cmd: Git and the extraction can take a while.
func (v *updateView) resolve() tea.Cmd {
	if v.releaseMode {
		v.flow++
		id := v.flow
		f := updateFlags{Version: strings.TrimSpace(v.releaseVersion.Value()), Home: v.cfg.Options.Home, StateDir: v.cfg.Options.StateDir}
		v.message = ""
		ctx, cancel := context.WithCancel(context.Background())
		v.cancelPreparation = cancel
		task := &updatePreparationTask{cancel: cancel, done: make(chan struct{})}
		v.releaseTask = task
		v.cancelRequested = false
		v.planning = "Checking release…"
		return func() tea.Msg {
			defer close(task.done)
			r, err := prepareReleaseUpdateContext(ctx, f)
			task.result = r
			msg := updatePlannedMsg{owned: owned{v}, flow: id, err: err, flags: f, release: r}
			if err == nil {
				var summary bytes.Buffer
				releaseUpdateSummary(&summary, r)
				msg.summary = summary.String()
				msg.unchanged = r.unchanged
			}
			return msg
		}
	}

	// The view's own checks, before any Git work: an empty field gets a clear
	// message instead of the command's error with a gap where the value goes.
	switch {
	case strings.TrimSpace(v.source.Value()) == "":
		v.message, v.messageErr = "Source is empty: enter the path of a Git checkout.", true
		return nil
	case strings.TrimSpace(v.rev.Value()) == "":
		v.message, v.messageErr = "Revision is empty: enter a revision such as HEAD.", true
		return nil
	}
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
		msg.flags = f
		msg.adapter, msg.offers, msg.err = detectUpdateOffers(plan, f, true)
		return msg
	}
}

// choose applies the operator's choice among the offered capabilities: it
// rebuilds the plan when the choice changes what update writes and renders the
// summary of what will be applied, like `hive update` does.
func (v *updateView) choose(planned updatePlannedMsg, requests []providerRequest) tea.Cmd {
	v.planning = "Preparing the summary…"
	return func() tea.Msg {
		msg := updatePlannedMsg{owned: owned{v}, flow: planned.flow, sourceLine: planned.sourceLine, flags: planned.flags}
		plan, unchanged, preview, err := finishUpdateChoice(planned.flags, planned.plan, planned.unchanged, planned.adapter, planned.offers, requests)
		if err != nil {
			msg.err = err
			return msg
		}
		var summary bytes.Buffer
		showInstallSummary(&summary, plan, preview, false, unchanged, false)
		summary.WriteString(planned.sourceLine + "\n")
		msg.plan, msg.summary, msg.unchanged = plan, summary.String(), unchanged
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
	if v.cancelPreparation != nil {
		v.cancelPreparation()
		v.cancelPreparation = nil
	}
	if v.cancelRequested {
		v.cancelRequested = false
		if msg.release != nil {
			_ = msg.release.packageFiles.Cleanup()
		}
		v.message = "Cancelled. No changes applied."
		v.messageErr = false
		return nil, action{nav: navNone}
	}
	if msg.err != nil {
		var pending *pendingReleaseUpdateError
		if errors.As(msg.err, &pending) {
			v.pendingRelease = &pending.journal
		}
		return v.fail(msg.err.Error())
	}
	if msg.release != nil {
		if msg.unchanged {
			_ = msg.release.packageFiles.Cleanup()
			v.message = "Hive is already up to date (" + msg.release.target.ProductVersion + ")."
			v.messageErr = false
			return nil, action{nav: navNone}
		}
		confirm := newConfirmView(confirmOptions{Title: "Update Hive release", Summary: msg.summary, OnApply: func() (tea.Cmd, action) {
			return applyReleaseUpdateCmd(v, msg), action{nav: navNone, write: true, result: v}
		}, OnCancel: func() (tea.Cmd, action) {
			_ = msg.release.packageFiles.Cleanup()
			v.message = "Cancelled. No changes applied."
			v.messageErr = false
			return nil, action{nav: navPop}
		}})
		return nil, action{nav: navPush, push: confirm}
	}
	if len(msg.offers) > 0 {
		caps := newCapsView(msg.offers,
			func(requests []providerRequest) (tea.Cmd, action) {
				return v.choose(msg, requests), action{nav: navPop}
			},
			func() (tea.Cmd, action) {
				// Esc cancels the update, as dismissing the offer does in
				// `hive update`. Unchecking the offer and continuing is the
				// way to skip it and still apply.
				v.message, v.messageErr = "Cancelled. No changes applied.", false
				return nil, action{nav: navPop}
			})
		return nil, action{nav: navPush, push: caps}
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
		reportUpdateViewer(&text, planned.plan.StateDir, planned.plan.SourceCommit)
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
		if _, state, err := management.NormalizeOptions(v.cfg.Options); err == nil {
			if pending, err := management.PendingUpdate(state); err == nil && pending != nil {
				v.pendingRelease = pending
				if pops > 0 {
					return nil, action{nav: navPop}
				}
				return nil, action{nav: navNone}
			}
		}
		if msg.pending != management.PendingNone {
			rv := newRecoveryView(msg.pending, copyOptions(v.cfg.Options), v.cfg.ExplicitStateDir, v.cfg.Deps)
			return nil, action{nav: navPush, pops: pops, push: rv}
		}
	} else {
		v.pendingRelease = nil
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
	applyInputStyles(&v.releaseVersion, th)
	row := func(i int, label string, in textinput.Model) string {
		prefix := "  "
		if i == v.focus {
			prefix = "> "
		}
		return prefix + padRight(label, 10) + inputView(in, th)
	}
	mode := "Source"
	if v.releaseMode {
		mode = "Release"
	}
	prefix := "  "
	if v.focus == -1 {
		prefix = "> "
	}
	lines := []string{th.Title.Render("Update"), "", prefix + padRight("Mode", 10) + mode}
	hint := "Enter resolves the committed source and previews changes."
	if v.releaseMode {
		lines = append(lines, row(0, "Version", v.releaseVersion), "Current: "+version.Current)
		hint = "Enter checks the release and previews changes."
	} else {
		lines = append(lines, row(0, "Source", v.source), row(1, "Revision", v.rev))
	}
	lines = append(lines, "")
	for _, l := range wrapLines(hint, c.Width) {
		lines = append(lines, th.Muted.Render(l))
	}

	switch {
	case v.planning != "":
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render(v.planning))
	case c.Busy:
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render("Applying…"))
	case v.message != "":
		if v.pendingRelease != nil {
			lines = append(lines, th.Muted.Render("Enter recovers offline; Esc goes back."))
		}
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
		binding("left,right", "←/→", "mode"),
		binding("up,down,tab", "↑/↓", "field"),
		binding("enter", "enter", "resolve"),
		binding("esc", "esc", "back"),
	}
}

func applyReleaseUpdateCmd(owner view, planned updatePlannedMsg) tea.Cmd {
	return func() tea.Msg {
		defer planned.release.packageFiles.Cleanup()
		var out bytes.Buffer
		err := applyPreparedRelease(planned.release, &out)
		return updateAppliedMsg{owned: owned{owner}, flow: planned.flow, text: strings.TrimSpace(out.String()), err: err, dialog: true}
	}
}

func recoverReleaseViewCmd(v *updateView, j management.UpdateJournal) tea.Cmd {
	flow := v.flow
	return func() tea.Msg {
		payload, err := json.Marshal(management.UpdateRecoverPayload{Journal: j})
		var r distribution.UpdateResponse
		if err == nil {
			r, err = management.RouteUpdate(distribution.UpdateRequest{Protocol: j.Plan.Protocol, OperationID: j.Plan.OperationID, Action: distribution.UpdateRecover, ProductVersion: j.Plan.Package.Artifact.ProductVersion, Payload: payload}, j.Plan)
		}
		if err == nil && r.Status != distribution.UpdateActivated && r.Status != distribution.UpdateRecovered {
			err = fmt.Errorf("core committed; command activation is still pending; retry recovery with %s", releaseRecoveryCommand(j.Plan))
		}
		var text bytes.Buffer
		fmt.Fprintf(&text, "Release update recovery: %s.\n", r.Status)
		if err == nil {
			printHookNotes(&text, releaseUpdateHookReminder(j.Plan))
		}
		return updateAppliedMsg{owned: owned{v}, flow: flow, text: strings.TrimSpace(text.String()), err: err}
	}
}

// QuitCleanup is called on the model goroutine. Its returned command owns
// disposal while the root ignores late preparation replies and apply keys.
func (v *updateView) QuitCleanup() tea.Cmd {
	task := v.releaseTask
	if task == nil {
		return nil
	}
	task.cancel()
	v.planning = "Cancelling preparation…"
	return func() tea.Msg {
		<-task.done
		if task.result != nil {
			_ = task.result.packageFiles.Cleanup()
		}
		return nil
	}
}
