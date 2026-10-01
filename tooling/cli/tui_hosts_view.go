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

	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/legacy"
	"tricell-hive/tooling/management"
)

// hostRow is one CLI the view lists. Release, Version and Drift are display
// strings, "-" when unknown, the same values `hive status` reports.
type hostRow struct {
	Name       string
	State      string // "registered", "legacy install", "editor only", "detected" or "not detected"
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
	// scanNote is set when only the legacy-installation scan failed: the rows
	// are still complete, except that a legacy install may not be labeled.
	scanNote string
	scanPath string // the file scanNote names, when it names one
	err      error
}

// legacyScanNote explains, in plain words followed by a "Detail:" line with the
// raw error, why the legacy scan failed. The scan reads every file Hive could
// have installed, and a file that differs from what Hive expects at a path it
// once used fails it; the file may be one Hive wrote or the user's own, so the
// note names the path and the ways out without saying which. It does not point
// to Diagnostics, which lists only files Hive currently manages.

func legacyScanNote(err error) string {
	raw := sanitizeLine(err.Error())
	words := "Hive could not check for a legacy installation, so installing or updating may be refused. The Detail line says why."
	var modified *legacy.ModifiedFileError
	if errors.As(err, &modified) {
		path := sanitizeLine(modified.Path)
		// The raw error says the same thing, so no Detail line repeats it.
		return path + " differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or updating. Removing may also be refused if Hive installed that file."
	}
	return words + "\nDetail: " + raw
}

// loadHostRows lists the detected, registered and legacy CLIs with what
// management.Status reports for the registered ones, and checks whether an
// operation is pending. It only reads state. A failed legacy scan does not
// hide the hosts: its note comes back beside the rows.
func loadHostRows(o management.Options, deps installDependencies) (rows []hostRow, pending management.PendingKind, scanNote, scanPath string, err error) {
	candidates, err := deps.DiscoverHosts(o)
	var scanErr *legacyScanError
	scanFailed := errors.As(err, &scanErr)
	if scanFailed {
		err = nil
	}
	if err != nil {
		return nil, management.PendingNone, "", "", err
	}
	var registered []string
	for _, c := range candidates {
		if c.Registered {
			registered = append(registered, c.Name)
		}
	}
	if scanFailed {
		scanNote = legacyScanNote(scanErr)
		var modified *legacy.ModifiedFileError
		if errors.As(scanErr, &modified) {
			scanPath = modified.Path
		}
	}
	var entries []management.StatusEntry
	if len(registered) > 0 {
		so := o
		so.Hosts = registered
		if entries, err = management.Status(so); err != nil {
			return nil, management.PendingNone, "", "", err
		}
	}
	for _, c := range candidates {
		if !c.Registered && !c.Legacy && !c.Detected && !c.Offered {
			continue
		}
		row := hostRow{Name: c.Name, Registered: c.Registered, Release: "-", Version: "-", Drift: "-"}
		switch {
		case c.Registered:
			row.State = "registered"
			row.Release, row.Version, row.Drift = hostStatusFields(c.Name, entries)
		case c.Legacy:
			row.State = "legacy install"
		case c.EditorOnly:
			row.State = "editor only"
		case !c.Detected:
			row.State = "not detected"
		default:
			row.State = "detected"
		}
		rows = append(rows, row)
	}
	pending, err = deps.Pending(o.StateDir)
	return rows, pending, scanNote, scanPath, err
}

// hostStatusFields reads one host's short release, product version and drift
// count from a Status result, the way `hive status` reports them.
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
		partial bool // the core installed, but an optional capability is pending
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
	hostsChangeW  = 9 // "→ install"
	hostsGap      = 2

	hostsVersionHeader      = "Hive version"
	hostsVersionShortHeader = "Hive"
)

type hostsView struct {
	cfg          appConfig
	rows         []hostRow
	checked      map[string]bool
	cursor       int
	seq          int
	loading      bool
	loadErr      string
	scanNote     string // why the legacy scan failed, when the rows are still shown
	scanPath     string // the file the scan note names, when it names one
	messagePath  string // the file the refusal in message names, when it names one
	planning     string
	message      string
	messageErr   bool
	messageTitle string // the read-in-full view's title; "" means Result, or Error for a failure
	messageCut   bool   // the last drawn message was cut, so `m` shows more
	offered      bool   // the recovery view was already offered for a pending operation
	dirty        bool   // the state may have changed under the view: reload when it is revealed
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
	v.dirty = false
	v.seq++
	v.loading = true
	seq, o, deps := v.seq, v.options(), v.cfg.Deps
	return func() tea.Msg {
		rows, pending, scanNote, scanPath, err := loadHostRows(o, deps)
		return hostsLoadedMsg{owned: owned{v}, seq: seq, rows: rows, pending: pending, scanNote: scanNote, scanPath: scanPath, err: err}
	}
}

// Reveal reloads the view when the views above it were popped and the state may
// have changed (after a write or a recovery), unless a flow or a load is already
// in progress. After a read-only view (the full message) the rows and the
// user's pending marks stay as they are.
func (v *hostsView) Reveal() tea.Cmd {
	if v.flow != nil || v.loading || !v.dirty {
		return nil
	}
	return v.reload()
}

// finish ends the flow with a message shown in the view, and reloads.
func (v *hostsView) finish(text string, isErr bool) tea.Cmd {
	v.flow, v.planning = nil, ""
	v.message, v.messageErr, v.messageTitle, v.messagePath = text, isErr, "", ""
	return v.reload()
}

func (v *hostsView) recoveryView(kind management.PendingKind) view {
	v.offered = true
	v.dirty = true // a recovery may change the state
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
		v.loadErr, v.scanNote, v.scanPath, v.rows = unreadableStateText(stateDirOf(v.cfg.Options), msg.err), "", "", nil
		return nil, action{nav: navNone}
	}
	v.loadErr, v.scanNote, v.scanPath = "", msg.scanNote, msg.scanPath
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
	case "a", "enter":
		// Enter reviews the pending changes like `a`, the documented key, so
		// it never does nothing silently.
		return v.startApply()
	case "m":
		if v.message != "" {
			title := v.messageTitle
			if title == "" {
				title = "Result"
				if v.messageErr {
					title = "Error"
				}
			}
			return nil, action{nav: navPush, push: newNoticeView(title, v.message, nil)}
		}
	case "u":
		return v.startUninstallAll()
	case "r":
		if v.loadErr != "" {
			return v.reload(), action{nav: navNone} // the state may be fixed by now
		}
	}
	return nil, action{nav: navNone}
}

// pendingChanges is the diff the checkboxes stand for: the unchecked
// registered CLIs to remove, and the checked unregistered ones to install.
func (v *hostsView) pendingChanges() (remove, add []string) {
	for _, r := range v.rows {
		switch {
		case r.Registered && !v.checked[r.Name]:
			remove = append(remove, r.Name)
		case !r.Registered && v.checked[r.Name]:
			add = append(add, r.Name)
		}
	}
	return remove, add
}

// startApply turns the checkboxes into the diff of design.md "Diff de CLIs":
// the unchecked registered CLIs are removed first, then the checked
// unregistered ones are installed.
func (v *hostsView) startApply() (tea.Cmd, action) {
	remove, add := v.pendingChanges()
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
		return v.finishRefused(msg.err, refusedText(msg.err)), action{nav: navNone}
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

// refusedText is the error of a step that failed while planning, before
// anything was written. The first sentence tells the user the action did not
// happen, since the refusal can read almost like the standing scan note.
func refusedText(err error) string {
	return "Nothing was changed. " + err.Error()
}

// refusedPath is the file a refusal names, when it is about a changed file.
func refusedPath(err error) string {
	var managed *management.ManagedFileChangedError
	if errors.As(err, &managed) {
		return managed.Path
	}
	var modified *legacy.ModifiedFileError
	if errors.As(err, &modified) {
		return modified.Path
	}
	return ""
}

// finishRefused ends the flow with text and remembers which file err names, so
// the scan note is hidden only when it is about that same file.
func (v *hostsView) finishRefused(err error, text string) tea.Cmd {
	cmd := v.finish(text, true)
	v.messagePath = refusedPath(err)
	return cmd
}

// failInstall ends the flow with an error from an install step; a removal
// already applied by step 1 is named first.
func (v *hostsView) failInstall(err error) (tea.Cmd, action) {
	text := refusedText(err)
	if prefix := v.removedPrefix(); prefix != "" {
		// Step 1 already wrote, so only the install step can say it changed nothing.
		text = prefix + "\n" + text
	}
	return v.finishRefused(err, text), action{nav: navNone}
}

// declineInstall ends the flow when the operator rejects an install step.
func (v *hostsView) declineInstall() tea.Cmd {
	text := "Cancelled. No changes applied."
	if prefix := v.removedPrefix(); prefix != "" {
		text = prefix + " Install cancelled; no further changes."
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
		msg := installAppliedMsg{owned: owned{owner}, flow: planned.flow, text: text.String(), err: finalErr, partial: result.Phase == "partial"}
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
	if msg.partial {
		v.messageTitle = "Result" // the core installed: a partial installation is an outcome, not an error
	}
	if msg.err != nil && msg.pending != management.PendingNone && !v.offered {
		return cmd, action{nav: navPush, pops: 1, push: v.recoveryView(msg.pending)}
	}
	return cmd, action{nav: navPop}
}

// ---------------------------------------------------------------------------
// Drawing.
// ---------------------------------------------------------------------------

// rowLines draws the header and one line per row, sized to the width: the
// version column takes what the fixed columns leave, ending with an ellipsis
// when it is longer. That column holds Hive's own product version, so its header
// says "Hive version" (the CLI's version is in Diagnostics), and shortens to
// "Hive" only when long names leave no room for the full label.
func (v *hostsView) rowLines(c viewCtx) []string {
	th := c.Theme
	nameW, versionW := len("CLI"), 0
	for _, r := range v.rows {
		nameW = max(nameW, min(len([]rune(r.Name)), hostsNamePad))
		versionW = max(versionW, len([]rune(r.Version)))
	}
	prefix := len("> [x] ")
	fixed := prefix + nameW + hostsGap + hostsStateCol + hostsGap + hostsReleaseW + hostsGap + hostsGap + hostsDriftW + hostsGap + hostsChangeW
	versionHeader := hostsVersionHeader
	if c.Width-fixed < len(versionHeader) {
		versionHeader = hostsVersionShortHeader
	}
	versionW = min(max(versionW, len(versionHeader)), max(c.Width-fixed, len(versionHeader)))
	format := func(name, state, release, version, drift string) string {
		gap := strings.Repeat(" ", hostsGap)
		return padRight(truncateRunes(name, nameW), nameW) + gap +
			padRight(truncateRunes(state, hostsStateCol), hostsStateCol) + gap +
			padRight(truncateRunes(release, hostsReleaseW), hostsReleaseW) + gap +
			padRight(truncateRunes(version, versionW), versionW) + gap +
			truncateRunes(drift, hostsDriftW)
	}
	lines := []string{th.Muted.Render(strings.Repeat(" ", prefix) + format("CLI", "State", "Release", versionHeader, "Drift"))}
	for i, r := range v.rows {
		box := "[ ]"
		if v.checked[r.Name] {
			box = "[x]"
		}
		text := format(r.Name, r.State, r.Release, r.Version, r.Drift)
		switch {
		case r.Registered && !v.checked[r.Name]:
			text += strings.Repeat(" ", hostsGap) + "→ remove"
		case !r.Registered && v.checked[r.Name]:
			text += strings.Repeat(" ", hostsGap) + "→ install"
		}
		if i == v.cursor {
			lines = append(lines, th.Accent.Render("> "+box+" "+text))
			continue
		}
		lines = append(lines, th.Text.Render("  "+box+" "+text))
	}
	return lines
}

// editorOnlyNote explains the "editor only" state, which only Cursor has: the
// editor launcher is on PATH but not the cursor-agent CLI.
const editorOnlyNote = "editor only: the Cursor editor is installed, but not its CLI (cursor-agent). Hive can still install Cursor's files."

// editorOnlyNoteLines draws the note under the table when a row says "editor
// only", and nothing otherwise.
func (v *hostsView) editorOnlyNoteLines(c viewCtx) []string {
	for _, r := range v.rows {
		if r.State != "editor only" {
			continue
		}
		lines := []string{""}
		for _, l := range wrapLines(editorOnlyNote, c.Width) {
			lines = append(lines, c.Theme.Muted.Render(l))
		}
		return lines
	}
	return nil
}

func (v *hostsView) View(c viewCtx) string {
	th := c.Theme
	v.messageCut = false
	lines := []string{th.Title.Render("CLIs"), ""}
	switch {
	case v.loading && len(v.rows) == 0:
		lines = append(lines, c.Spinner+" "+th.Muted.Render("Loading CLIs…"))
	case v.loadErr != "":
		// The plain words and the way out come first; the raw error follows
		// on its own line, so a short screen cuts it before them.
		words, detail := errLines(v.loadErr)
		for _, l := range wrapLines("The CLIs cannot be shown: "+words, c.Width) {
			lines = append(lines, th.Danger.Render(l))
		}
		lines = append(lines, th.Muted.Render("Press r to retry after fixing it."))
		for _, d := range detail {
			for _, l := range wrapLines(d, c.Width) {
				lines = append(lines, th.Muted.Render(l))
			}
		}
	case len(v.rows) == 0:
		text := "No CLI hosts were detected or registered. Install a supported CLI (claude, codex, cursor, grok, opencode, pi) and reopen this view."
		for _, l := range wrapLines(text, c.Width) {
			lines = append(lines, th.Text.Render(l))
		}
		lines = append(lines, v.scanNoteLines(c)...)
	default:
		lines = append(lines, v.rowLines(c)...)
		lines = append(lines, v.editorOnlyNoteLines(c)...)
		remove, add := v.pendingChanges()
		if n := len(remove) + len(add); n == 1 {
			lines = append(lines, "", th.Muted.Render("1 pending change: press a to review it."))
		} else if n > 1 {
			lines = append(lines, "", th.Muted.Render(fmt.Sprintf("%d pending changes: press a to review them.", n)))
		}
		lines = append(lines, v.scanNoteLines(c)...)
	}
	if v.planning != "" {
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render(v.planning))
	} else if v.loading && len(v.rows) > 0 {
		lines = append(lines, "", c.Spinner+" "+th.Muted.Render("Refreshing…"))
	}
	// The pending-operation note goes below the message, so its lines are
	// reserved first: the message gives up its own lines, ending in an
	// ellipsis, rather than the note being cut off.
	var noteLines []string
	if v.stillPending {
		note := "An interrupted operation is still pending. " + recoveryTextOnOpen(v.cfg.ExplicitStateDir)(v.cfg.Options.StateDir, true) + "."
		noteLines = wrapLines(note, c.Width)
	}
	if v.message != "" {
		style := th.Text
		if v.messageErr {
			style = th.Danger
		}
		lines = append(lines, "")
		room := max(c.Height-len(lines)-len(noteLines), 1)
		wrapped := wrapLines(v.message, c.Width)
		v.messageCut = len(wrapped) > room
		if v.messageCut {
			// Cut, but never silently: the full text opens with `m`.
			wrapped = append(wrapped[:room-1], "… press m to read all")
		}
		for _, l := range wrapped {
			lines = append(lines, style.Render(l))
		}
	}
	for _, l := range noteLines {
		lines = append(lines, th.Danger.Render(l))
	}
	return strings.Join(lines, "\n")
}

// scanNoteLines draws the note about a failed legacy scan, blank line first,
// or nothing when the scan did not fail. It belongs under the rows and under
// the empty state alike, since the file it names is the problem in both.
func (v *hostsView) scanNoteLines(c viewCtx) []string {
	if v.scanNote == "" {
		return nil
	}
	if v.messageErr && v.messagePath != "" && v.messagePath == v.scanPath {
		return nil // the refusal below says the same about the same file
	}
	th := c.Theme
	words, detail := errLines(v.scanNote)
	lines := []string{""}
	for _, l := range wrapLines(words, c.Width) {
		lines = append(lines, th.Danger.Render(l))
	}
	if v.message != "" {
		detail = nil // a result or error below needs the room more than the raw error
	}
	for _, d := range detail {
		for _, l := range wrapLines(d, c.Width) {
			lines = append(lines, th.Muted.Render(l))
		}
	}
	return lines
}

func (v *hostsView) Keys() []key.Binding {
	back := binding("esc,backspace", "esc", "back")
	more := binding("m", "m", "more") // only while the message is cut
	if len(v.rows) == 0 {
		keys := []key.Binding{}
		if v.messageCut {
			keys = append(keys, more)
		}
		if v.loadErr != "" {
			keys = append(keys, binding("r", "r", "reload"))
		}
		return append(keys, back)
	}
	keys := []key.Binding{
		binding("up,down", "↑/↓", "move"),
		binding("space", "space", "mark"),
		binding("a,enter", "a", "review"),
		binding("u", "u", "uninstall all"),
	}
	if v.messageCut {
		// Room for `m more` on an 80-column help bar: the arrows are obvious.
		keys = keys[1:]
		keys = append(keys, more)
	}
	return append(keys, back)
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
			lines = append(lines, "    "+label+inputView(v.inputs[i], th))
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

// providerOfferLabel mirrors selectProviderRequests's own per-offer label
// text (install.go): the capability's name, its source, and whether it needs a
// version.
func providerOfferLabel(o providerOffer) string {
	version := "exact version required"
	if o.ManualOnly {
		version = "manual instructions only"
	}
	return fmt.Sprintf("%s (%s; %s)", o.Name, o.Source, version)
}
