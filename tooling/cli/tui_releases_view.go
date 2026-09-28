// tui_releases_view.go is the Releases view (T9; design.md "Vistas"): the
// retained releases, newest first, in a scrolling list that follows the cursor,
// with a substring filter opened by `/`. Enter on a release goes to the summary
// of `plan install --release` for the registered CLIs (buildRollbackPlan) and
// its confirmation. The list is our own, not
// bubbles/list, which would add a module.
package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
)

type (
	releasesLoadedMsg struct {
		owned
		seq         int
		entries     []management.ReleaseEntry
		installedID string
		err         error
	}
	rollbackPlannedMsg struct {
		owned
		flow    int
		id      string
		plan    management.Plan
		summary string
		err     error
	}
	rollbackAppliedMsg struct {
		owned
		flow    int
		text    string
		err     error
		pending management.PendingKind
	}
)

type releasesView struct {
	cfg         appConfig
	entries     []management.ReleaseEntry
	installedID string
	seq         int
	loading     bool
	loadErr     string
	filter      textinput.Model
	filtering   bool // the filter field has the focus
	cursor      int  // index into the visible (filtered) releases
	vp          viewport.Model
	listHeight  int
	planning    string
	message     string
	messageErr  bool
	flow        int
}

func newReleasesView(cfg appConfig) *releasesView {
	f := textinput.New()
	f.Prompt = ""
	f.Placeholder = "type to filter"
	return &releasesView{cfg: cfg, filter: f, loading: true, vp: viewport.New(), listHeight: 10}
}

func (v *releasesView) options() management.Options { return copyOptions(v.cfg.Options) }

func (v *releasesView) Init() tea.Cmd { return v.reload() }

func (v *releasesView) Resize(w, _ int) { v.filter.SetWidth(max(w-12, 10)) }

func (v *releasesView) TextFocused() bool { return v.filtering }

func (v *releasesView) NeedsSpinner() bool { return v.loading || v.planning != "" }

func (v *releasesView) reload() tea.Cmd {
	v.seq++
	v.loading = true
	seq, o := v.seq, v.options()
	return func() tea.Msg {
		msg := releasesLoadedMsg{owned: owned{v}, seq: seq}
		if msg.entries, msg.err = management.Releases(o); msg.err != nil {
			return msg
		}
		if len(msg.entries) > 0 {
			msg.installedID, msg.err = installedReleaseID(o)
		}
		return msg
	}
}

// Reveal reloads the list when the views above it were popped, unless a load
// or a plan is already running.
func (v *releasesView) Reveal() tea.Cmd {
	if v.loading || v.planning != "" {
		return nil
	}
	return v.reload()
}

// visible is the releases the filter lets through, in list order.
func (v *releasesView) visible() []management.ReleaseEntry {
	needle := strings.ToLower(strings.TrimSpace(v.filter.Value()))
	if needle == "" {
		return v.entries
	}
	var out []management.ReleaseEntry
	for _, e := range v.entries {
		if strings.Contains(strings.ToLower(v.filterText(e)), needle) {
			out = append(out, e)
		}
	}
	return out
}

// filterText is what the filter searches: everything about the release, not
// only what fits in its row (the row cuts a long host list with an ellipsis).
func (v *releasesView) filterText(e management.ReleaseEntry) string {
	parts := []string{e.ID, e.LastWrittenAt}
	parts = append(parts, e.Commits...)
	parts = append(parts, releaseHostNames(e.Consumers)...)
	if e.ID == v.installedID {
		parts = append(parts, "installed")
	}
	return strings.Join(parts, " ")
}

func (v *releasesView) clampCursor() {
	v.cursor = max(min(v.cursor, len(v.visible())-1), 0)
}

func (v *releasesView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case releasesLoadedMsg:
		if msg.seq != v.seq {
			break
		}
		v.loading = false
		if msg.err != nil {
			v.loadErr, v.entries = msg.err.Error(), nil
			break
		}
		v.loadErr, v.entries, v.installedID = "", msg.entries, msg.installedID
		v.clampCursor()
	case rollbackPlannedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onPlanned(msg)
	case rollbackAppliedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onApplied(msg)
	case tea.KeyPressMsg:
		return v.onKey(msg)
	default:
		if v.filtering {
			var cmd tea.Cmd
			v.filter, cmd = v.filter.Update(msg)
			return cmd, action{nav: navNone}
		}
	}
	return nil, action{nav: navNone}
}

func (v *releasesView) onKey(msg tea.KeyPressMsg) (tea.Cmd, action) {
	name := msg.String()
	if v.loading || v.planning != "" {
		if name == "esc" || name == "backspace" {
			v.flow++ // leaving abandons the plan; its late result is ignored
			v.planning = ""
			return nil, action{}
		}
		return nil, action{nav: navNone}
	}
	if v.filtering {
		switch name {
		case "enter":
			v.filtering = false
			v.filter.Blur()
		case "esc":
			v.clearFilter()
		case "up", "down", "pgup", "pgdown":
			v.move(name)
		default:
			var cmd tea.Cmd
			v.filter, cmd = v.filter.Update(msg)
			v.cursor = 0
			v.clampCursor()
			return cmd, action{nav: navNone}
		}
		return nil, action{nav: navNone}
	}
	switch name {
	case "esc":
		if v.filter.Value() != "" {
			v.clearFilter()
			return nil, action{nav: navNone}
		}
		return nil, action{}
	case "backspace":
		return nil, action{}
	case "/":
		v.filtering = true
		return v.filter.Focus(), action{nav: navNone}
	case "up", "down", "pgup", "pgdown", "home", "end":
		v.move(name)
	case "enter":
		return v.choose()
	}
	return nil, action{nav: navNone}
}

func (v *releasesView) clearFilter() {
	v.filter.SetValue("")
	v.filter.Blur()
	v.filtering = false
	v.cursor = 0
}

func (v *releasesView) move(name string) {
	n := len(v.visible())
	page := max(v.listHeight-1, 1)
	switch name {
	case "up":
		v.cursor--
	case "down":
		v.cursor++
	case "pgup":
		v.cursor -= page
	case "pgdown":
		v.cursor += page
	case "home":
		v.cursor = 0
	case "end":
		v.cursor = n - 1
	}
	v.cursor = max(min(v.cursor, n-1), 0)
}

// choose starts the rollback to the release under the cursor.
func (v *releasesView) choose() (tea.Cmd, action) {
	visible := v.visible()
	if len(visible) == 0 {
		return nil, action{nav: navNone}
	}
	chosen := visible[v.cursor].ID
	v.message = ""
	if v.installedID != "" && chosen == v.installedID {
		v.message, v.messageErr = "Already installed", false
		return nil, action{nav: navNone}
	}
	v.flow++
	id, o := v.flow, v.options()
	v.planning = "Planning the rollback…"
	return func() tea.Msg {
		msg := rollbackPlannedMsg{owned: owned{v}, flow: id, id: chosen}
		plan, unchanged, err := buildRollbackPlan(o, chosen)
		if err != nil {
			msg.err = err
			return msg
		}
		var summary bytes.Buffer
		showInstallSummary(&summary, plan, onboardingPreview{}, false, unchanged, false)
		msg.plan, msg.summary = plan, summary.String()
		return msg
	}, action{nav: navNone}
}

func (v *releasesView) onPlanned(msg rollbackPlannedMsg) (tea.Cmd, action) {
	v.planning = ""
	if msg.err != nil {
		v.message, v.messageErr = msg.err.Error(), true
		return nil, action{nav: navNone}
	}
	confirm := newConfirmView(confirmOptions{
		Title:   "Roll back to " + shortHash(msg.id),
		Summary: msg.summary,
		OnApply: func() (tea.Cmd, action) {
			return applyRollbackCmd(v, msg, v.cfg.Deps), action{nav: navNone, write: true, result: v}
		},
		OnCancel: func() (tea.Cmd, action) {
			v.message, v.messageErr = "Cancelled. No changes applied.", false
			return nil, action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

func applyRollbackCmd(owner view, planned rollbackPlannedMsg, deps installDependencies) tea.Cmd {
	return func() tea.Msg {
		result, err := (management.Engine{}).Apply(planned.plan)
		msg := rollbackAppliedMsg{owned: owned{owner}, flow: planned.flow, err: err}
		if err != nil {
			msg.pending = pendingAfterFailure(deps, planned.plan.StateDir)
			return msg
		}
		var text bytes.Buffer
		reportApplyResult(&text, "Hive rolled back", result)
		msg.text = strings.TrimRight(text.String(), "\n")
		return msg
	}
}

func (v *releasesView) onApplied(msg rollbackAppliedMsg) (tea.Cmd, action) {
	v.message, v.messageErr = msg.text, false
	if msg.err != nil {
		v.message, v.messageErr = msg.err.Error(), true
		if msg.pending != management.PendingNone {
			rv := newRecoveryView(msg.pending, v.options(), v.cfg.ExplicitStateDir, v.cfg.Deps)
			return v.reload(), action{nav: navPush, pops: 1, push: rv}
		}
	}
	return v.reload(), action{nav: navPop}
}

func (v *releasesView) View(c viewCtx) string {
	th := c.Theme
	applyInputStyles(&v.filter, th)
	lines := []string{th.Title.Render("Releases")}
	if v.filtering || v.filter.Value() != "" {
		label := "  Filter: "
		if v.filtering {
			label = "/ Filter: " // the list cursor keeps the only ">"
		}
		lines = append(lines, label+inputView(v.filter, th))
	} else {
		lines = append(lines, th.Muted.Render("  Press / to filter."))
	}

	var status []string
	switch {
	case v.planning != "":
		status = append(status, c.Spinner+" "+th.Muted.Render(v.planning))
	case c.Busy:
		status = append(status, c.Spinner+" "+th.Muted.Render("Applying…"))
	case v.loading && len(v.entries) == 0:
		status = append(status, c.Spinner+" "+th.Muted.Render("Loading releases…"))
	case v.loading:
		status = append(status, c.Spinner+" "+th.Muted.Render("Refreshing…"))
	}
	if v.message != "" {
		style := th.Text
		if v.messageErr {
			style = th.Danger
		}
		wrapped := wrapLines(v.message, c.Width)
		if len(wrapped) > 4 {
			wrapped = append(wrapped[:3], "…")
		}
		for _, l := range wrapped {
			status = append(status, style.Render(l))
		}
	}
	reserved := len(status)
	if reserved > 0 {
		reserved++ // the blank line above the status
	}
	height := max(c.Height-len(lines)-reserved, 1)
	v.listHeight = height

	visible := v.visible()
	switch {
	case v.loadErr != "":
		for _, l := range wrapLines("Cannot read the releases: "+v.loadErr, c.Width) {
			lines = append(lines, th.Danger.Render(l))
		}
	case len(v.entries) == 0 && !v.loading:
		lines = append(lines, th.Text.Render("No releases are retained yet"))
	case len(visible) == 0 && len(v.entries) > 0:
		lines = append(lines, th.Text.Render("No release matches the filter."))
	default:
		v.clampCursor()
		rows := make([]string, len(visible))
		for i, e := range visible {
			label := formatReleaseLabel(e, e.ID == v.installedID)
			if i == v.cursor {
				rows[i] = th.Accent.Render("> " + label)
			} else {
				rows[i] = th.Text.Render("  " + label)
			}
		}
		v.vp.SetWidth(c.Width)
		v.vp.SetHeight(height)
		v.vp.SetContentLines(rows)
		off := v.vp.YOffset()
		switch {
		case v.cursor < off:
			v.vp.SetYOffset(v.cursor)
		case v.cursor >= off+height:
			v.vp.SetYOffset(v.cursor - height + 1)
		}
		lines = append(lines, v.vp.View())
	}
	if reserved > 0 {
		lines = append(lines, "")
		lines = append(lines, status...)
	}
	return strings.Join(lines, "\n")
}

func (v *releasesView) Keys() []key.Binding {
	if len(v.entries) == 0 && !v.filtering {
		return []key.Binding{binding("esc", "esc", "back")}
	}
	if v.filtering {
		return []key.Binding{binding("enter", "enter", "keep filter"), binding("esc", "esc", "clear filter")}
	}
	return []key.Binding{
		binding("up,down,pgup,pgdown", "↑/↓", "move"),
		binding("enter", "enter", "roll back"),
		binding("/", "/", "filter"),
		binding("esc", "esc", "back"),
	}
}

// installedReleaseID names the release ID currently installed for the
// registered user-scope hosts, reusing tui.go's own latestRelease the menu's
// status line already relies on. It returns "" (not an error) when no host
// is registered yet: Releases still lists retained snapshots in that case,
// it just cannot mark one "installed".
func installedReleaseID(o management.Options) (string, error) {
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return "", err
	}
	if len(hosts) == 0 {
		return "", nil
	}
	so := o
	so.Hosts = hosts
	entries, err := management.Status(so)
	if err != nil {
		return "", err
	}
	id, _ := latestRelease(entries)
	return id, nil
}

// buildRollbackPlan builds the plan the Releases view shows: the install plan pinned to releaseID for the registered hosts, and
// whether it changes nothing. A release the manager cannot validate comes back
// with the downgrade hint appended: only a problem with the release's own
// retained snapshot gets it, never an unrelated failure such as "explicit
// hosts required" or "unfinished operation: recover first".
func buildRollbackPlan(o management.Options, releaseID string) (plan management.Plan, unchanged bool, err error) {
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return management.Plan{}, false, err
	}
	ro := o
	ro.Hosts = hosts
	ro.ReleaseID = releaseID
	plan, err = management.BuildPlan("install", ro)
	if err != nil {
		if isReleaseValidationError(err) {
			return management.Plan{}, false, fmt.Errorf("%w; plan that downgrade with the manager from the commit that produced it", err)
		}
		return management.Plan{}, false, err
	}
	unchanged, err = management.PlanUnchanged(plan)
	if err != nil {
		return management.Plan{}, false, err
	}
	return plan, unchanged, nil
}

// releaseValidationErrorPrefixes are every error text tooling/management's
// own loadRelease/validateRelease (plan.go) can produce while checking a
// pinned ReleaseID's own retained snapshot — the closed set read directly
// from that function's source, current as of this change. The downgrade hint
// buildRollbackPlan adds applies only when a BuildPlan
// failure starts with one of these, never to an unrelated failure such as
// "explicit hosts required" or "unfinished operation: recover first", which
// loadRelease is never reached to produce.
var releaseValidationErrorPrefixes = []string{
	"invalid release ID",
	"invalid release fingerprint or file list",
	"invalid source ",
	"missing skill entrypoint:",
	"duplicate agent name:",
	"agent release missing renderer",
	"unsupported agent field ",
	"unsupported agent renderer",
	"reserved source delimiter",
	"nonportable personal path in ",
}

// isReleaseValidationError reports whether err's own message starts with one
// of releaseValidationErrorPrefixes.
func isReleaseValidationError(err error) bool {
	msg := err.Error()
	for _, prefix := range releaseValidationErrorPrefixes {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return false
}

// formatReleaseLabel renders one release's own Select option label at or
// under releaseLabelWidth display columns, INCLUDING the " (installed)"
// marker: its short ID, date, first commit (or "-" when the release
// predates commit tracking) and its consuming hosts, truncated to whatever
// room is left, with the marker for the one currently applied.
//
// releaseLabelWidth is 76: the Releases view puts a two-column cursor prefix
// ("> " or two spaces) before the label, so a row is at most 78 columns
// (design.md "Vistas": rows of 78 columns or less), marker included.
const releaseLabelWidth = 76

func formatReleaseLabel(e management.ReleaseEntry, installed bool) string {
	shortID := shortHash(e.ID)
	date := e.LastWrittenAt
	if len(date) > 10 {
		date = date[:10]
	}
	commit := "-"
	if len(e.Commits) > 0 {
		commit = shortHash(e.Commits[0])
	}
	marker := ""
	if installed {
		marker = " (installed)"
	}
	prefix := fmt.Sprintf("%s  %s  %s  ", shortID, date, commit)
	budget := releaseLabelWidth - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(marker)
	if budget < 0 {
		budget = 0
	}
	return prefix + truncateJoined(releaseHostNames(e.Consumers), budget) + marker
}

// releaseHostNames is a release's own unique, sorted consumer host names.
func releaseHostNames(consumers []management.Consumer) []string {
	seen := map[string]bool{}
	var names []string
	for _, c := range consumers {
		if !seen[c.Host] {
			seen[c.Host] = true
			names = append(names, c.Host)
		}
	}
	sort.Strings(names)
	return names
}

// truncateJoined joins names with ", " and truncates the result to width
// runes via truncateRunes (never splitting a name's own first character off
// with nothing to show for it).
func truncateJoined(names []string, width int) string {
	return truncateRunes(strings.Join(names, ", "), width)
}
