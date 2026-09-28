// tui_hosts_view.go is the CLIs view (T8; design.md "Vistas" and "Diff de
// CLIs"): one row per detected or registered CLI with a checkbox, and an
// apply that turns the checkboxes into removals and additions. Every step
// reuses the pieces the commands use: BuildPlan and Engine.Apply for
// removals, and the install flow's checks, plan, summary and
// applyInstallOnboarding for additions. Planning and applying run as tea.Cmd
// values; the view only changes inside Update, and it owns the flow that
// stacks the summary and confirmation views above it.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
)

// hostRow is one CLI the view lists. Release, Version and Drift are display
// strings, "-" when unknown, the same values `hive status` reports.
type hostRow struct {
	Name       string
	State      string // "registered", "legacy install" or "detected"
	Registered bool
	Release    string
	Version    string
	Drift      string
}

// owned marks a message with the view that issued it. The root delivers it to
// that view even when another view was pushed on top in the meantime.
type owned struct{ owner view }

func (o owned) target() view { return o.owner }

// hostsLoadedMsg is the result of loading the rows and checking for a pending
// operation.
type hostsLoadedMsg struct {
	owned
	seq     int
	rows    []hostRow
	pending management.PendingKind
	err     error
}

// loadHostRows lists the detected, registered and legacy CLIs with what
// management.Status reports for the registered ones, and checks whether an
// operation is pending. It only reads state.
func loadHostRows(o management.Options, deps installDependencies) ([]hostRow, management.PendingKind, error) {
	candidates, err := deps.DiscoverHosts(o)
	if err != nil {
		return nil, management.PendingNone, err
	}
	var registered []string
	for _, c := range candidates {
		if c.Registered {
			registered = append(registered, c.Name)
		}
	}
	var entries []management.StatusEntry
	if len(registered) > 0 {
		so := o
		so.Hosts = registered
		if entries, err = management.Status(so); err != nil {
			return nil, management.PendingNone, err
		}
	}
	var rows []hostRow
	for _, c := range candidates {
		if !c.Registered && !c.Legacy && !c.Detected {
			continue
		}
		row := hostRow{Name: c.Name, Registered: c.Registered, Release: "-", Version: "-", Drift: "-"}
		switch {
		case c.Registered:
			row.State = "registered"
			row.Release, row.Version, row.Drift = hostStatusFields(c.Name, entries)
		case c.Legacy:
			row.State = "legacy install"
		default:
			row.State = "detected"
		}
		rows = append(rows, row)
	}
	pending, err := deps.Pending(o.StateDir)
	return rows, pending, err
}

// hostStatusFields reads one host's short release, product version and drift
// count from a Status result, the way statusLine does.
func hostStatusFields(host string, entries []management.StatusEntry) (release, version, drift string) {
	release, version = "-", "-"
	count := 0
	for _, e := range entries {
		if e.Kind == "voice" || e.Host != host {
			continue
		}
		if e.Release != "" {
			release = shortHash(e.Release)
		}
		if e.ProductVersion != "" {
			version = e.ProductVersion
		}
		if e.Status == "drift" {
			count++
		}
	}
	return release, version, fmt.Sprint(count)
}

// Messages of the apply flow. Each carries the flow's id, so a result that
// arrives after the flow was abandoned is ignored.
type (
	removePlannedMsg struct {
		owned
		flow    int
		plan    management.Plan
		summary string
		err     error
	}
	removeAppliedMsg struct {
		owned
		flow    int
		text    string
		err     error
		pending management.PendingKind
	}
	installCheckedMsg struct {
		owned
		flow         int
		hosts        []string
		notice       string
		needsConsent bool
		err          error
	}
	installPlannedMsg struct {
		owned
		flow      int
		hosts     []string
		plan      management.Plan
		unchanged bool
		adapter   onboardingAdapter
		offers    []providerOffer
		err       error
	}
	installPreviewMsg struct {
		owned
		flow    int
		planned installPlannedMsg
		preview onboardingPreview
		summary string
		err     error
	}
	installAppliedMsg struct {
		owned
		flow    int
		text    string
		err     error
		pending management.PendingKind
	}
)

// hostsFlow is one apply in progress: the CLIs to remove and to install, and
// the ones step 1 already removed.
type hostsFlow struct {
	id      int
	all     bool
	remove  []string
	add     []string
	removed []string
}

func (f *hostsFlow) twoSteps() bool { return len(f.remove) > 0 && len(f.add) > 0 }

const (
	hostsNamePad  = 16
	hostsStateCol = 14 // "legacy install"
	hostsReleaseW = 12
	hostsDriftW   = 5
	hostsGap      = 2
)

type hostsView struct {
	cfg          appConfig
	rows         []hostRow
	checked      map[string]bool
	cursor       int
	seq          int
	loading      bool
	loadErr      string
	planning     string
	message      string
	messageErr   bool
	offered      bool // the recovery view was already offered for a pending operation
	stillPending bool
	flow         *hostsFlow
	flowSeq      int
}

func newHostsView(cfg appConfig) *hostsView {
	return &hostsView{cfg: cfg, checked: map[string]bool{}, loading: true}
}

func (v *hostsView) options() management.Options { return copyOptions(v.cfg.Options) }

func (v *hostsView) Init() tea.Cmd { return v.reload() }

func (v *hostsView) Resize(int, int) {}

func (v *hostsView) TextFocused() bool { return false }

func (v *hostsView) NeedsSpinner() bool { return v.loading || v.planning != "" }

// reload reads the rows and the pending check again; the checkboxes go back to
// the registered state.
func (v *hostsView) reload() tea.Cmd {
	v.seq++
	v.loading = true
	seq, o, deps := v.seq, v.options(), v.cfg.Deps
	return func() tea.Msg {
		rows, pending, err := loadHostRows(o, deps)
		return hostsLoadedMsg{owned: owned{v}, seq: seq, rows: rows, pending: pending, err: err}
	}
}

// Reveal reloads the view when the views above it were popped, unless a flow
// or a load is already in progress.
func (v *hostsView) Reveal() tea.Cmd {
	if v.flow != nil || v.loading {
		return nil
	}
	return v.reload()
}

// finish ends the flow with a message shown in the view, and reloads.
func (v *hostsView) finish(text string, isErr bool) tea.Cmd {
	v.flow, v.planning = nil, ""
	v.message, v.messageErr = text, isErr
	return v.reload()
}

func (v *hostsView) recoveryView(kind management.PendingKind) view {
	v.offered = true
	return newRecoveryView(kind, v.options(), v.cfg.ExplicitStateDir, v.cfg.Deps)
}

func (v *hostsView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case hostsLoadedMsg:
		return v.onLoaded(msg)
	case tea.KeyPressMsg:
		return v.onKey(msg.String())
	case removePlannedMsg:
		if v.stale(msg.flow) {
			break
		}
		return v.onRemovePlanned(msg)
	case removeAppliedMsg:
		if v.stale(msg.flow) {
			break
		}
		return v.onRemoveApplied(msg)
	case installCheckedMsg:
		if v.stale(msg.flow) {
			break
		}
		return v.onInstallChecked(msg)
	case installPlannedMsg:
		if v.stale(msg.flow) {
			break
		}
		return v.onInstallPlanned(msg)
	case installPreviewMsg:
		if v.stale(msg.flow) {
			break
		}
		return v.onInstallPreview(msg)
	case installAppliedMsg:
		if v.stale(msg.flow) {
			break
		}
		return v.onInstallApplied(msg)
	}
	return nil, action{nav: navNone}
}

func (v *hostsView) stale(flow int) bool { return v.flow == nil || v.flow.id != flow }

func (v *hostsView) onLoaded(msg hostsLoadedMsg) (tea.Cmd, action) {
	if msg.seq != v.seq {
		return nil, action{nav: navNone}
	}
	v.loading = false
	if msg.err != nil {
		v.loadErr, v.rows = msg.err.Error(), nil
		return nil, action{nav: navNone}
	}
	v.loadErr = ""
	v.rows = msg.rows
	v.checked = map[string]bool{}
	for _, r := range v.rows {
		v.checked[r.Name] = r.Registered
	}
	v.cursor = min(v.cursor, max(len(v.rows)-1, 0))
	v.stillPending = false
	switch {
	case msg.pending == management.PendingNone:
		v.offered = false
	case !v.offered:
		return nil, action{nav: navPush, push: v.recoveryView(msg.pending)}
	default:
		v.stillPending = true
	}
	return nil, action{nav: navNone}
}

func (v *hostsView) onKey(name string) (tea.Cmd, action) {
	if v.loading || v.planning != "" {
		if name == "esc" || name == "backspace" {
			v.flow, v.planning = nil, "" // leaving abandons the flow; late results are ignored
			return nil, action{}
		}
		return nil, action{nav: navNone}
	}
	switch name {
	case "esc", "backspace":
		return nil, action{}
	case "up":
		v.cursor = max(v.cursor-1, 0)
	case "down":
		v.cursor = min(v.cursor+1, max(len(v.rows)-1, 0))
	case "space":
		if len(v.rows) > 0 {
			r := v.rows[v.cursor]
			v.checked[r.Name] = !v.checked[r.Name]
		}
	case "a":
		return v.startApply()
	case "u":
		return v.startUninstallAll()
	}
	return nil, action{nav: navNone}
}

// startApply turns the checkboxes into the diff of design.md "Diff de CLIs":
// the unchecked registered CLIs are removed first, then the checked
// unregistered ones are installed.
func (v *hostsView) startApply() (tea.Cmd, action) {
	var remove, add []string
	for _, r := range v.rows {
		switch {
		case r.Registered && !v.checked[r.Name]:
			remove = append(remove, r.Name)
		case !r.Registered && v.checked[r.Name]:
			add = append(add, r.Name)
		}
	}
	v.message = ""
	if len(remove) == 0 && len(add) == 0 {
		v.message, v.messageErr = "Nothing to apply", false
		return nil, action{nav: navNone}
	}
	v.flowSeq++
	v.flow = &hostsFlow{id: v.flowSeq, remove: remove, add: add}
	if len(remove) > 0 {
		return v.planRemove(), action{nav: navNone}
	}
	return v.beginInstall(), action{nav: navNone}
}

// startUninstallAll removes every registered CLI, skipping the checkboxes.
func (v *hostsView) startUninstallAll() (tea.Cmd, action) {
	var registered []string
	for _, r := range v.rows {
		if r.Registered {
			registered = append(registered, r.Name)
		}
	}
	v.message = ""
	if len(registered) == 0 {
		v.message, v.messageErr = "No CLI hosts are registered.", false
		return nil, action{nav: navNone}
	}
	v.flowSeq++
	v.flow = &hostsFlow{id: v.flowSeq, all: true, remove: registered}
	return v.planRemove(), action{nav: navNone}
}

// ---------------------------------------------------------------------------
// Removal.
// ---------------------------------------------------------------------------

func (v *hostsView) planRemove() tea.Cmd {
	v.planning = "Planning the removal…"
	id, o, hosts := v.flow.id, v.options(), append([]string(nil), v.flow.remove...)
	return func() tea.Msg {
		plan, err := buildRemovePlan(o, hosts)
		var summary bytes.Buffer
		if err == nil {
			showRemoveSummary(&summary, plan)
		}
		return removePlannedMsg{owned: owned{v}, flow: id, plan: plan, summary: summary.String(), err: err}
	}
}

func (v *hostsView) onRemovePlanned(msg removePlannedMsg) (tea.Cmd, action) {
	v.planning = ""
	if msg.err != nil {
		return v.finish(msg.err.Error(), true), action{nav: navNone}
	}
	f := v.flow
	title := "Remove " + strings.Join(f.remove, ", ")
	switch {
	case f.all:
		title = "Uninstall all CLIs"
	case f.twoSteps():
		title = "Step 1 of 2: remove " + strings.Join(f.remove, ", ")
	}
	confirm := newConfirmView(confirmOptions{
		Title:         title,
		Summary:       msg.summary,
		StartOnCancel: f.all,
		DisableYes:    f.all,
		OnApply: func() (tea.Cmd, action) {
			return applyRemoveCmd(v, f.id, msg.plan, v.cfg.Deps), action{nav: navNone, write: true, result: v}
		},
		OnCancel: func() (tea.Cmd, action) {
			return v.finish("Cancelled. No changes applied.", false), action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

func applyRemoveCmd(owner view, id int, plan management.Plan, deps installDependencies) tea.Cmd {
	return func() tea.Msg {
		result, err := removeApply(plan)
		msg := removeAppliedMsg{owned: owned{owner}, flow: id, err: err}
		if err != nil {
			msg.pending = pendingAfterFailure(deps, plan.StateDir)
			return msg
		}
		var text bytes.Buffer
		reportApplyResult(&text, "Hive removed", result)
		msg.text = text.String()
		return msg
	}
}

// pendingAfterFailure reports whether a failed write left an operation or an
// onboarding pending; a check that itself fails counts as none.
func pendingAfterFailure(deps installDependencies, stateDir string) management.PendingKind {
	kind, err := deps.Pending(stateDir)
	if err != nil {
		return management.PendingNone
	}
	return kind
}

func (v *hostsView) onRemoveApplied(msg removeAppliedMsg) (tea.Cmd, action) {
	if msg.err != nil {
		// A failed removal never continues to the install; if it left an
		// operation pending, the recovery view opens.
		cmd := v.finish(msg.err.Error(), true)
		if msg.pending != management.PendingNone && !v.offered {
			return cmd, action{nav: navPush, pops: 1, push: v.recoveryView(msg.pending)}
		}
		return cmd, action{nav: navPop}
	}
	f := v.flow
	f.removed = f.remove
	if len(f.add) == 0 {
		return v.finish(strings.TrimRight(msg.text, "\n"), false), action{nav: navPop}
	}
	text := fmt.Sprintf("Removed: %s.\n\nNext: install %s.", strings.Join(f.removed, ", "), strings.Join(f.add, ", "))
	notice := newNoticeView("Step 1 of 2 done", text, func() (tea.Cmd, action) {
		return v.beginInstall(), action{nav: navPop}
	}).withBack(func() (tea.Cmd, action) {
		// Stop after step 1: the removal stays applied and the view says so.
		return v.declineInstall(), action{nav: navPop}
	})
	return nil, action{nav: navPush, pops: 1, push: notice}
}

// ---------------------------------------------------------------------------
// Installation.
// ---------------------------------------------------------------------------

func (v *hostsView) removedPrefix() string {
	if v.flow == nil || len(v.flow.removed) == 0 {
		return ""
	}
	return fmt.Sprintf("Removed: %s.", strings.Join(v.flow.removed, ", "))
}

// beginInstall starts the additions: the checks `hive install` makes before it
// plans, then the shared-resource notice.
func (v *hostsView) beginInstall() tea.Cmd {
	v.planning = "Checking the install…"
	id, o, deps := v.flow.id, v.options(), v.cfg.Deps
	o.Hosts = append([]string(nil), v.flow.add...)
	return func() tea.Msg { return checkInstall(v, id, o, deps) }
}

func checkInstall(owner view, id int, o management.Options, deps installDependencies) installCheckedMsg {
	if err := distribution.VerifyIfPackaged(o.Source); err != nil {
		return installCheckedMsg{owned: owned{owner}, flow: id, err: err}
	}
	if !sourceHasCatalog(o.Source) {
		return installCheckedMsg{owned: owned{owner}, flow: id, err: errors.New("Run hive from a Hive checkout or package, or pass --source")}
	}
	var notice bytes.Buffer
	hosts, needsConsent, err := describeRequiredHosts(&notice, o, deps)
	return installCheckedMsg{owned: owned{owner}, flow: id, hosts: hosts, notice: notice.String(), needsConsent: needsConsent, err: err}
}

// failInstall ends the flow with an error from an install step; a removal
// already applied by step 1 is named first.
func (v *hostsView) failInstall(err error) (tea.Cmd, action) {
	text := err.Error()
	if prefix := v.removedPrefix(); prefix != "" {
		text = prefix + "\n" + text
	}
	return v.finish(text, true), action{nav: navNone}
}

// declineInstall ends the flow when the operator rejects an install step.
func (v *hostsView) declineInstall() tea.Cmd {
	text := "Cancelled. No changes applied."
	if prefix := v.removedPrefix(); prefix != "" {
		text = prefix + " Install cancelled; no further changes"
	}
	return v.finish(text, false)
}

func (v *hostsView) installTitle(hosts []string) string {
	if v.flow != nil && v.flow.twoSteps() {
		return "Step 2 of 2: install " + strings.Join(hosts, ", ")
	}
	return "Install " + strings.Join(hosts, ", ")
}

func (v *hostsView) onInstallChecked(msg installCheckedMsg) (tea.Cmd, action) {
	v.planning = ""
	if msg.err != nil {
		return v.failInstall(msg.err)
	}
	if !msg.needsConsent {
		return v.planInstall(msg.hosts), action{nav: navNone}
	}
	confirm := newConfirmView(confirmOptions{
		Title:      "Required CLIs",
		Summary:    msg.notice,
		ApplyLabel: "Accept",
		OnApply: func() (tea.Cmd, action) {
			return v.planInstall(msg.hosts), action{nav: navPop}
		},
		OnCancel: func() (tea.Cmd, action) {
			return v.declineInstall(), action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

func (v *hostsView) planInstall(hosts []string) tea.Cmd {
	v.planning = "Planning the install…"
	id, o, deps := v.flow.id, v.options(), v.cfg.Deps
	o.Hosts = append([]string(nil), hosts...)
	return func() tea.Msg {
		msg := installPlannedMsg{owned: owned{v}, flow: id, hosts: o.Hosts}
		var err error
		if msg.plan, err = management.BuildPlan("install", o); err != nil {
			msg.err = err
			return msg
		}
		if msg.unchanged, err = management.PlanUnchanged(msg.plan); err != nil {
			msg.err = err
			return msg
		}
		if msg.adapter, err = deps.AdapterFactory(newOnboardingInput(o, false)); err != nil {
			msg.err = err
			return msg
		}
		msg.offers, msg.err = msg.adapter.Detect(o)
		return msg
	}
}

func (v *hostsView) onInstallPlanned(msg installPlannedMsg) (tea.Cmd, action) {
	v.planning = ""
	if msg.err != nil {
		return v.failInstall(msg.err)
	}
	if len(msg.offers) == 0 {
		return v.previewInstall(msg, nil), action{nav: navNone}
	}
	caps := newCapsView(msg.offers,
		func(requests []providerRequest) (tea.Cmd, action) {
			return v.previewInstall(msg, requests), action{nav: navPop}
		},
		func() (tea.Cmd, action) {
			return v.declineInstall(), action{nav: navPop}
		})
	return nil, action{nav: navPush, push: caps}
}

func (v *hostsView) previewInstall(planned installPlannedMsg, requests []providerRequest) tea.Cmd {
	v.planning = "Preparing the summary…"
	return func() tea.Msg {
		msg := installPreviewMsg{owned: owned{v}, flow: planned.flow, planned: planned}
		preview, err := planned.adapter.Plan(planned.plan, requests)
		if err != nil {
			msg.err = err
			return msg
		}
		msg.preview = preview
		var summary bytes.Buffer
		showInstallSummary(&summary, planned.plan, preview, false, planned.unchanged, false)
		msg.summary = summary.String()
		return msg
	}
}

func (v *hostsView) onInstallPreview(msg installPreviewMsg) (tea.Cmd, action) {
	v.planning = ""
	if msg.err != nil {
		return v.failInstall(msg.err)
	}
	planned := msg.planned
	confirm := newConfirmView(confirmOptions{
		Title:   v.installTitle(planned.hosts),
		Summary: msg.summary,
		OnApply: func() (tea.Cmd, action) {
			cmd := applyInstallCmd(v, planned, msg.preview, v.cfg)
			return cmd, action{nav: navNone, write: true, result: v}
		},
		OnCancel: func() (tea.Cmd, action) {
			return v.declineInstall(), action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

// applyInstallCmd applies the install as a write: the same
// applyInstallOnboarding and finalizeInstallResult `hive install` uses, with
// fromInterface set and the real explicit state directory.
func applyInstallCmd(owner view, planned installPlannedMsg, preview onboardingPreview, cfg appConfig) tea.Cmd {
	stateDir := cfg.Options.StateDir
	return func() tea.Msg {
		result, err := applyInstallOnboarding(planned.plan, preview, planned.adapter)
		var text bytes.Buffer
		finalErr := finalizeInstallResult(&text, result, err, false, cfg.ExplicitStateDir, stateDir, true)
		msg := installAppliedMsg{owned: owned{owner}, flow: planned.flow, text: text.String(), err: finalErr}
		if finalErr != nil {
			msg.pending = pendingAfterFailure(cfg.Deps, stateDir)
		}
		return msg
	}
}

func (v *hostsView) onInstallApplied(msg installAppliedMsg) (tea.Cmd, action) {
	text := strings.TrimRight(msg.text, "\n")
	if msg.err != nil {
		if text != "" {
			text += "\n"
		}
		text += msg.err.Error()
	}
	if prefix := v.removedPrefix(); prefix != "" {
		text = prefix + "\n" + text
	}
	cmd := v.finish(text, msg.err != nil)
	if msg.err != nil && msg.pending != management.PendingNone && !v.offered {
		return cmd, action{nav: navPush, pops: 1, push: v.recoveryView(msg.pending)}
	}
	return cmd, action{nav: navPop}
}

// ---------------------------------------------------------------------------
// Drawing.
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
	return strings.Split(ansi.Wrap(text, max(width, 1), ""), "\n")
}

// rowLines draws the header and one line per row, sized to the width: the
// version column takes what the fixed columns leave, ending with an ellipsis
// when it is longer.
func (v *hostsView) rowLines(c viewCtx) []string {
	th := c.Theme
	nameW, versionW := len("CLI"), len("Version")
	for _, r := range v.rows {
		nameW = max(nameW, min(len([]rune(r.Name)), hostsNamePad))
		versionW = max(versionW, len([]rune(r.Version)))
	}
	prefix := len("> [x] ")
	fixed := prefix + nameW + hostsGap + hostsStateCol + hostsGap + hostsReleaseW + hostsGap + hostsGap + hostsDriftW
	versionW = min(versionW, max(c.Width-fixed, len("Version")))
	format := func(name, state, release, version, drift string) string {
		gap := strings.Repeat(" ", hostsGap)
		return padRight(truncateRunes(name, nameW), nameW) + gap +
			padRight(truncateRunes(state, hostsStateCol), hostsStateCol) + gap +
			padRight(truncateRunes(release, hostsReleaseW), hostsReleaseW) + gap +
			padRight(truncateRunes(version, versionW), versionW) + gap +
			truncateRunes(drift, hostsDriftW)
	}
	lines := []string{th.Muted.Render(strings.Repeat(" ", prefix) + format("CLI", "State", "Release", "Version", "Drift"))}
	for i, r := range v.rows {
		box := "[ ]"
		if v.checked[r.Name] {
			box = "[x]"
		}
		text := format(r.Name, r.State, r.Release, r.Version, r.Drift)
		if i == v.cursor {
			lines = append(lines, th.Accent.Render("> "+box+" "+text))
			continue
		}
		lines = append(lines, th.Text.Render("  "+box+" "+text))
	}
	return lines
}

func (v *hostsView) View(c viewCtx) string {
	th := c.Theme
	lines := []string{th.Title.Render("CLIs"), ""}
	switch {
	case v.loading && len(v.rows) == 0:
		lines = append(lines, c.Spinner+" "+th.Muted.Render("Loading CLIs…"))
	case v.loadErr != "":
		for _, l := range wrapLines("Cannot read the CLIs: "+v.loadErr, c.Width) {
			lines = append(lines, th.Danger.Render(l))
		}
	case len(v.rows) == 0:
		text := "No CLI hosts were detected or registered. Install a supported CLI (claude, codex, cursor, grok, opencode, pi) and reopen this view."
		for _, l := range wrapLines(text, c.Width) {
			lines = append(lines, th.Text.Render(l))
		}
	default:
		lines = append(lines, v.rowLines(c)...)
	}
	if v.planning != "" {
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render(v.planning))
	} else if v.loading && len(v.rows) > 0 {
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render("Refreshing…"))
	}
	if v.message != "" {
		style := th.Text
		if v.messageErr {
			style = th.Danger
		}
		lines = append(lines, "")
		room := max(c.Height-len(lines)-1, 1)
		wrapped := wrapLines(v.message, c.Width)
		if len(wrapped) > room {
			wrapped = append(wrapped[:room-1], "…")
		}
		for _, l := range wrapped {
			lines = append(lines, style.Render(l))
		}
	}
	if v.stillPending {
		note := "An interrupted operation is still pending. " + recoveryTextOnOpen(v.cfg.ExplicitStateDir)(v.cfg.Options.StateDir, true) + "."
		for _, l := range wrapLines(note, c.Width) {
			lines = append(lines, th.Danger.Render(l))
		}
	}
	return strings.Join(lines, "\n")
}

func (v *hostsView) Keys() []key.Binding {
	return []key.Binding{
		binding("up,down", "↑/↓", "move"),
		binding("space", "space", "toggle"),
		binding("a", "a", "apply"),
		binding("u", "u", "uninstall all"),
		binding("esc,backspace", "esc", "back"),
	}
}

// ---------------------------------------------------------------------------
// Optional capabilities.
// ---------------------------------------------------------------------------

// capsView lists the optional capabilities the install flow's adapter offers,
// none checked. Checking one that asks for a version opens its version field
// with the focus; Enter or Tab close the field and return to the list, and
// Enter on the list goes on to the summary.
type capsView struct {
	offers   []providerOffer
	checked  []bool
	inputs   []textinput.Model
	cursor   int
	editing  bool
	errText  string
	onDone   func([]providerRequest) (tea.Cmd, action)
	onCancel func() (tea.Cmd, action)
}

func newCapsView(offers []providerOffer, onDone func([]providerRequest) (tea.Cmd, action), onCancel func() (tea.Cmd, action)) *capsView {
	v := &capsView{
		offers:   offers,
		checked:  make([]bool, len(offers)),
		inputs:   make([]textinput.Model, len(offers)),
		onDone:   onDone,
		onCancel: onCancel,
	}
	for i := range v.inputs {
		in := textinput.New()
		in.Prompt = ""
		in.Placeholder = "exact version"
		v.inputs[i] = in
	}
	return v
}

func (v *capsView) Init() tea.Cmd { return nil }

func (v *capsView) Resize(w, _ int) {
	for i := range v.inputs {
		v.inputs[i].SetWidth(max(w-16, 10))
	}
}

func (v *capsView) TextFocused() bool { return v.editing }

func (v *capsView) stopEditing() {
	v.inputs[v.cursor].Blur()
	v.editing = false
}

func (v *capsView) startEditing(i int) tea.Cmd {
	v.cursor, v.editing = i, true
	return v.inputs[i].Focus()
}

func (v *capsView) Update(msg tea.Msg) (tea.Cmd, action) {
	k, isKey := msg.(tea.KeyPressMsg)
	if v.editing {
		if isKey {
			switch k.String() {
			case "enter", "tab", "esc":
				v.stopEditing()
				return nil, action{nav: navNone}
			}
			v.errText = ""
		}
		var cmd tea.Cmd
		v.inputs[v.cursor], cmd = v.inputs[v.cursor].Update(msg)
		return cmd, action{nav: navNone}
	}
	if !isKey {
		return nil, action{nav: navNone}
	}
	switch k.String() {
	case "up":
		v.cursor = max(v.cursor-1, 0)
	case "down":
		v.cursor = min(v.cursor+1, len(v.offers)-1)
	case "space":
		return v.toggle(), action{nav: navNone}
	case "enter":
		return v.proceed()
	case "esc", "backspace":
		return v.onCancel()
	}
	return nil, action{nav: navNone}
}

func (v *capsView) toggle() tea.Cmd {
	i := v.cursor
	v.errText = ""
	v.checked[i] = !v.checked[i]
	if !v.checked[i] {
		v.inputs[i].SetValue("")
		return nil
	}
	if v.offers[i].ManualOnly {
		return nil
	}
	return v.startEditing(i)
}

// proceed collects the requests, refusing a checked capability with a blank
// version the way the install flow's own prompt does.
func (v *capsView) proceed() (tea.Cmd, action) {
	var requests []providerRequest
	for i, offer := range v.offers {
		if !v.checked[i] {
			continue
		}
		if offer.ManualOnly {
			requests = append(requests, providerRequest{ID: offer.ID})
			continue
		}
		version := strings.TrimSpace(v.inputs[i].Value())
		if version == "" {
			v.errText = "exact version required for " + offer.Name
			return v.startEditing(i), action{nav: navNone}
		}
		requests = append(requests, providerRequest{ID: offer.ID, Version: version})
	}
	return v.onDone(requests)
}

func (v *capsView) View(c viewCtx) string {
	th := c.Theme
	in := textinput.Styles{}
	in.Focused = textinput.StyleState{Text: th.Text, Placeholder: th.Muted, Prompt: th.Muted}
	in.Blurred = in.Focused
	lines := []string{th.Title.Render("Optional capabilities"), th.Muted.Render("None is fine: leave them unchecked and continue."), ""}
	for i, offer := range v.offers {
		box := "[ ]"
		if v.checked[i] {
			box = "[x]"
		}
		row := fmt.Sprintf("%s %s", box, providerOfferLabel(offer))
		if i == v.cursor && !v.editing {
			lines = append(lines, th.Accent.Render("> "+row))
		} else {
			lines = append(lines, th.Text.Render("  "+row))
		}
		if v.checked[i] && !offer.ManualOnly {
			v.inputs[i].SetStyles(in)
			label := "  Version: "
			if i == v.cursor && v.editing {
				label = "> Version: "
			}
			lines = append(lines, "    "+label+v.inputs[i].View())
		}
	}
	if v.errText != "" {
		lines = append(lines, "", th.Danger.Render(v.errText))
	}
	return strings.Join(lines, "\n")
}

func (v *capsView) Keys() []key.Binding {
	if v.editing {
		return []key.Binding{binding("enter,tab,esc", "enter", "done")}
	}
	return []key.Binding{
		binding("up,down", "↑/↓", "move"),
		binding("space", "space", "toggle"),
		binding("enter", "enter", "continue"),
		binding("esc", "esc", "cancel"),
	}
}
