// tui_voice_view.go is the Voice view (T9; design.md "Vistas"): the voice, its
// address, name and intensity as rows changed in place (↑↓ moves between rows,
// ←→ changes the value), loaded from the active voice. Enter builds the same
// plan `hive voice set` or `off` builds, shows its summary and confirms.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
)

var (
	voiceAddresses   = []string{"none", "sir", "name"}
	voiceIntensities = []string{"subtle", "marked"}
)

type (
	voiceLoadedMsg struct {
		owned
		seq     int
		voices  []management.VoiceInfo
		active  *management.VoiceSetting
		noHosts bool
		err     error
		// stateErr marks an err that came from reading Hive's state, as opposed
		// to the source's voices, which have their own guidance.
		stateErr bool
	}
	voicePlannedMsg struct {
		owned
		flow      int
		plan      management.Plan
		summary   string
		unchanged bool
		off       bool
		err       error
	}
	voiceAppliedMsg struct {
		owned
		flow    int
		text    string
		err     error
		pending management.PendingKind
	}
)

// loadVoice reads what the view needs: whether any CLI is registered, the
// active voice, and the source's voices.
func loadVoice(o management.Options) (msg voiceLoadedMsg) {
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		msg.err, msg.stateErr = err, true
		return msg
	}
	if len(hosts) == 0 {
		msg.noHosts = true
		return msg
	}
	if msg.active, err = management.CurrentVoice(o); err != nil {
		msg.err, msg.stateErr = err, true
		return msg
	}
	// A source without content/voices/ would make ListVoices return a raw
	// filesystem error; give the same actionable guidance the install flow gives.
	if !sourceHasVoices(o.Source) {
		msg.err = errSourceWithoutVoices
		return msg
	}
	msg.voices, msg.err = management.ListVoices(o.Source)
	return msg
}

type sourceError string

func (e sourceError) Error() string { return string(e) }

const errSourceWithoutVoices = sourceError("Run hive from a Hive checkout or package, or pass --source")

const voiceLabelCol = 11

type voiceView struct {
	cfg        appConfig
	seq        int
	loading    bool
	ready      bool
	noHosts    bool
	loadErr    string
	stateErr   bool // loadErr is the unreadable-state text
	voices     []management.VoiceInfo
	voice      int // 0 is Off, i is voices[i-1]
	address    int
	intensity  int
	name       textinput.Model
	row        int
	planning   string
	message    string
	messageErr bool
	flow       int
}

func newVoiceView(cfg appConfig) *voiceView {
	n := textinput.New()
	n.Prompt = ""
	n.Placeholder = "a name is required"
	return &voiceView{cfg: cfg, name: n, loading: true}
}

func (v *voiceView) options() management.Options { return copyOptions(v.cfg.Options) }

func (v *voiceView) Init() tea.Cmd { return v.reload() }

func (v *voiceView) Resize(w, _ int) { v.name.SetWidth(max(w-voiceLabelCol-4, 10)) }

func (v *voiceView) NeedsSpinner() bool { return v.loading || v.planning != "" }

func (v *voiceView) reload() tea.Cmd {
	v.seq++
	v.loading = true
	seq, o := v.seq, v.options()
	return func() tea.Msg {
		msg := loadVoice(o)
		msg.owned, msg.seq = owned{v}, seq
		return msg
	}
}

// rows are the kinds of rows shown now: Name only with address "name", and
// nothing but the voice when it is Off.
func (v *voiceView) rows() []string {
	if v.voice == 0 {
		return []string{"voice"}
	}
	rows := []string{"voice", "address"}
	if v.address == 2 {
		rows = append(rows, "name")
	}
	return append(rows, "intensity")
}

func (v *voiceView) current() string {
	rows := v.rows()
	return rows[min(v.row, len(rows)-1)]
}

func (v *voiceView) TextFocused() bool { return v.ready && v.current() == "name" }

// sync keeps the name field's focus and the row index consistent with the
// rows shown.
func (v *voiceView) sync() tea.Cmd {
	v.row = min(v.row, len(v.rows())-1)
	if v.current() == "name" {
		return v.name.Focus()
	}
	v.name.Blur()
	return nil
}

func (v *voiceView) prefill(active *management.VoiceSetting) {
	v.voice, v.address, v.intensity = 0, 0, 0
	v.name.SetValue("")
	if active == nil {
		return
	}
	for i, info := range v.voices {
		if info.ID == active.ID {
			v.voice = i + 1
		}
	}
	if v.voice == 0 {
		// The active voice is not among this source's voices: show it, marked,
		// rather than Off, so reviewing never proposes turning it off.
		v.voices = append(v.voices, management.VoiceInfo{ID: active.ID, Description: "(not in this source)"})
		v.voice = len(v.voices)
	}
	for i, a := range voiceAddresses {
		if a == active.Address {
			v.address = i
		}
	}
	for i, in := range voiceIntensities {
		if in == active.Intensity {
			v.intensity = i
		}
	}
	v.name.SetValue(active.Name)
}

func (v *voiceView) Update(msg tea.Msg) (tea.Cmd, action) {
	switch msg := msg.(type) {
	case voiceLoadedMsg:
		if msg.seq != v.seq {
			break
		}
		v.loading, v.noHosts, v.loadErr, v.stateErr = false, msg.noHosts, "", msg.stateErr
		if msg.err != nil {
			v.loadErr, v.ready = msg.err.Error(), false
			if msg.stateErr {
				v.loadErr = unreadableStateText(stateDirOf(v.cfg.Options), msg.err)
			}
			break
		}
		if msg.noHosts {
			v.ready = false
			break
		}
		v.voices, v.ready = msg.voices, true
		v.prefill(msg.active)
		return v.sync(), action{nav: navNone}
	case voicePlannedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onPlanned(msg)
	case voiceAppliedMsg:
		if msg.flow != v.flow {
			break
		}
		return v.onApplied(msg)
	case tea.KeyPressMsg:
		return v.onKey(msg)
	default:
		if v.TextFocused() {
			var cmd tea.Cmd
			v.name, cmd = v.name.Update(msg)
			return cmd, action{nav: navNone}
		}
	}
	return nil, action{nav: navNone}
}

func cycle(i, delta, n int) int { return ((i+delta)%n + n) % n }

func (v *voiceView) onKey(msg tea.KeyPressMsg) (tea.Cmd, action) {
	name := msg.String()
	if v.loading || v.planning != "" {
		// Esc leaves, and so does Backspace unless a text field has the focus
		// (there it must not silently abandon the plan). Leaving drops the
		// plan's late result.
		if name == "esc" || (name == "backspace" && !v.TextFocused()) {
			v.flow++
			v.planning = ""
			return nil, action{}
		}
		return nil, action{nav: navNone}
	}
	if !v.ready {
		if name == "esc" || name == "backspace" {
			return nil, action{}
		}
		if name == "r" && v.stateErr {
			return v.reload(), action{nav: navNone} // no text field is open yet, and the state may be fixed
		}
		return nil, action{nav: navNone}
	}
	switch name {
	case "esc":
		return nil, action{}
	case "up":
		v.row = max(v.row-1, 0)
		return v.sync(), action{nav: navNone}
	case "down":
		v.row = min(v.row+1, len(v.rows())-1)
		return v.sync(), action{nav: navNone}
	case "enter":
		return v.plan()
	case "left", "right":
		if v.current() != "name" {
			delta := 1
			if name == "left" {
				delta = -1
			}
			v.message = "" // an edit: the previous result no longer describes the rows
			switch v.current() {
			case "voice":
				v.voice = cycle(v.voice, delta, len(v.voices)+1)
			case "address":
				v.address = cycle(v.address, delta, len(voiceAddresses))
			case "intensity":
				v.intensity = cycle(v.intensity, delta, len(voiceIntensities))
			}
			return v.sync(), action{nav: navNone}
		}
	}
	if v.current() == "name" {
		v.message = ""
		var cmd tea.Cmd
		v.name, cmd = v.name.Update(msg)
		return cmd, action{nav: navNone}
	}
	if name == "backspace" {
		return nil, action{}
	}
	return nil, action{nav: navNone}
}

// plan builds the voice plan the command would: "off" for Off, "set" otherwise.
func (v *voiceView) plan() (tea.Cmd, action) {
	v.flow++
	id, o := v.flow, v.options()
	v.message = ""
	off := v.voice == 0
	setting := management.VoiceSetting{}
	if !off {
		setting = management.VoiceSetting{ID: v.voices[v.voice-1].ID, Address: voiceAddresses[v.address], Intensity: voiceIntensities[v.intensity]}
		if v.address == 2 {
			setting.Name = v.name.Value()
		}
	}
	v.planning = "Planning the voice change…"
	return func() tea.Msg {
		msg := voicePlannedMsg{owned: owned{v}, flow: id, off: off}
		kind := "set"
		if off {
			kind = "off"
		}
		plan, err := management.BuildVoicePlan(kind, o, setting)
		if err != nil {
			msg.err = err
			return msg
		}
		if msg.unchanged, err = management.PlanUnchanged(plan); err != nil {
			msg.err = err
			return msg
		}
		var summary bytes.Buffer
		showVoiceSummary(&summary, plan, msg.unchanged)
		msg.plan, msg.summary = plan, summary.String()
		return msg
	}, action{nav: navNone}
}

func (v *voiceView) onPlanned(msg voicePlannedMsg) (tea.Cmd, action) {
	v.planning = ""
	if msg.err != nil {
		v.message, v.messageErr = msg.err.Error(), true
		return nil, action{nav: navNone}
	}
	if msg.unchanged {
		v.message, v.messageErr = "Voice is already set this way", false
		return nil, action{nav: navNone}
	}
	verb, title := "Voice set", "Set the voice"
	if msg.off {
		verb, title = "Voice turned off", "Turn the voice off"
	}
	confirm := newConfirmView(confirmOptions{
		Title:   title,
		Summary: msg.summary,
		OnApply: func() (tea.Cmd, action) {
			return applyVoiceCmd(v, msg, verb, v.cfg.Deps), action{nav: navNone, write: true, result: v}
		},
		OnCancel: func() (tea.Cmd, action) {
			v.message, v.messageErr = "Cancelled. No changes applied.", false
			return nil, action{nav: navPop}
		},
	})
	return nil, action{nav: navPush, push: confirm}
}

func applyVoiceCmd(owner view, planned voicePlannedMsg, verb string, deps installDependencies) tea.Cmd {
	return func() tea.Msg {
		result, err := (management.Engine{}).Apply(planned.plan)
		msg := voiceAppliedMsg{owned: owned{owner}, flow: planned.flow, err: err}
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

func (v *voiceView) onApplied(msg voiceAppliedMsg) (tea.Cmd, action) {
	v.message, v.messageErr = msg.text, false
	if msg.err != nil {
		v.message, v.messageErr = msg.err.Error(), true
		if msg.pending != management.PendingNone {
			rv := newRecoveryView(msg.pending, v.options(), v.cfg.ExplicitStateDir, v.cfg.Deps)
			return nil, action{nav: navPush, pops: 1, push: rv}
		}
		return nil, action{nav: navPop}
	}
	return v.reload(), action{nav: navPop}
}

func (v *voiceView) value(kind string) string {
	switch kind {
	case "voice":
		if v.voice == 0 {
			return "Off"
		}
		return voiceOptionLabel(v.voices[v.voice-1])
	case "address":
		return voiceAddresses[v.address]
	default:
		return voiceIntensities[v.intensity]
	}
}

func (v *voiceView) View(c viewCtx) string {
	th := c.Theme
	applyInputStyles(&v.name, th)
	lines := []string{th.Title.Render("Voice"), ""}
	switch {
	case v.loading && !v.ready:
		lines = append(lines, c.Spinner+" "+th.Muted.Render("Loading the voice…"))
	case v.noHosts:
		lines = append(lines, th.Text.Render("No CLI hosts are registered. Open CLIs to install one."))
	case v.loadErr != "":
		// A state that cannot be read leads with plain words and the way out,
		// then the raw error on its own line; a source problem is one line.
		words, detail := errLines(v.loadErr)
		if v.stateErr {
			words = "The voice cannot be shown: " + words
		}
		for _, l := range wrapLines(words, c.Width) {
			lines = append(lines, th.Danger.Render(l))
		}
		if v.stateErr {
			lines = append(lines, th.Muted.Render("Press r to retry after fixing it."))
		}
		for _, d := range detail {
			for _, l := range wrapLines(d, c.Width) {
				lines = append(lines, th.Muted.Render(l))
			}
		}
	case v.ready:
		labels := map[string]string{"voice": "Voice", "address": "Address", "name": "Name", "intensity": "Intensity"}
		rows := v.rows()
		for i, kind := range rows {
			prefix := "  "
			if i == min(v.row, len(rows)-1) {
				prefix = "> "
			}
			text := prefix + padRight(labels[kind], voiceLabelCol)
			if kind == "name" {
				lines = append(lines, text+inputView(v.name, th))
				continue
			}
			value := "< " + truncateRunes(v.value(kind), max(c.Width-len(text)-4, 1)) + " >"
			if i == min(v.row, len(rows)-1) {
				lines = append(lines, th.Accent.Render(text+value))
			} else {
				lines = append(lines, th.Text.Render(text+value))
			}
		}
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

func (v *voiceView) Keys() []key.Binding {
	if !v.ready {
		if v.stateErr && v.loadErr != "" {
			return []key.Binding{binding("r", "r", "reload"), binding("esc", "esc", "back")}
		}
		return []key.Binding{binding("esc", "esc", "back")}
	}
	return []key.Binding{
		binding("up,down", "↑/↓", "row"),
		binding("left,right", "←/→", "value"),
		binding("enter", "enter", "review"),
		binding("esc", "esc", "back"),
	}
}

// voiceLabelWidth bounds voiceOptionLabel to one rendered line: the view
// shortens it further to what its row leaves.
const voiceLabelWidth = 76

// voiceOptionLabel renders one ListVoices entry as "id — description",
// truncating the description (never the ID) so the whole label fits within
// voiceLabelWidth columns. Before this, a full-length description (real
// voices like jarvis or mentor carry one- or two-sentence descriptions) could
// wrap to two or more rendered lines and push later options out of a fixed
// height list — confirmed directly: at width 80, jarvis and
// mentor's own full descriptions wrapped, hiding both from a fresh 80x24
// render along with everything after them.
func voiceOptionLabel(v management.VoiceInfo) string {
	prefix := v.ID + " — "
	budget := voiceLabelWidth - utf8.RuneCountInString(prefix)
	if budget < 0 {
		budget = 0
	}
	return prefix + truncateRunes(voiceDisplayDescription(v.Description), budget)
}

// sourceHasVoices reports whether source has a content/voices/ directory
// (T6 fix round F4), the same presence check sourceHasCatalog (tui_hosts.go)
// applies to management.GlobalSource. management.ListVoices' own
// os.ReadDir(dir) requires a directory, not merely an existing path, so a
// stray content/voices file (not a directory) is treated the same as a
// missing one here rather than surfacing ListVoices' own raw error either
// way.
func sourceHasVoices(source string) bool {
	info, err := os.Stat(filepath.Join(source, management.VoicesSource))
	return err == nil && info.IsDir()
}
