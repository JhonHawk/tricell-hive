// tui_models_picker.go is the model picker of the Models view (#46, D9-A and
// its corrections): a raised panel over the edit panel and the table, opened
// from the Model field. It lists the CLI's models, by provider on the CLIs whose
// ids carry one and flat on the rest, and ends with two entries, "release
// default" and "Other…". It draws no border: every line is padded to the view's
// width on the panel's own background, so nothing can run past the screen's edge.
package main

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// modelList is the open picker: the search text, the cursor and the scroll
// offset. The cursor counts the model rows first, then the two bottom entries.
type modelList struct {
	filter string
	cur    int
	off    int // first visible line of the scrolling area
}

const (
	currentMarker = "● "
	pickerHint    = "↑↓ enter esc"
	pickerRowsFix = 5 // title, blank, search, status or blank, and the bottom row
)

// providerTitles are the readable names of the providers OpenCode and Pi list;
// any other provider shows its own id.
var providerTitles = map[string]string{
	"github-copilot": "GitHub Copilot", "openai": "OpenAI", "anthropic": "Anthropic",
	"opencode": "OpenCode", "opencode-go": "OpenCode Go", "xai": "xAI", "google": "Google",
	"openai-codex": "OpenAI Codex", "mistral": "Mistral", "deepseek": "DeepSeek", "groq": "Groq",
	"openrouter": "OpenRouter", "amazon-bedrock": "Amazon Bedrock", "azure": "Azure",
}

func providerTitle(id string) string {
	if t, ok := providerTitles[id]; ok {
		return t
	}
	return id
}

// pickRow is one selectable model.
type pickRow struct {
	id, name, provider string
	current            bool
}

// shortID is the id as the picker shows it under a provider: without its
// "<provider>/" prefix. The value chosen stays the full id.
func (r pickRow) shortID() string {
	if r.provider != "" {
		return strings.TrimPrefix(r.id, r.provider+"/")
	}
	return r.id
}

// pickLine is one line of the scrolling area: a section header, a blank
// separator, or a model row.
type pickLine struct {
	header string
	blank  bool
	row    int
}

type pickerContent struct {
	rows  []pickRow
	lines []pickLine
}

// openPicker opens the picker over the Model field, with text as the first
// characters of the search. With no search the cursor starts on the model the
// panel stands on, or on the first model when none is (a mixed group).
func (v *modelsView) openPicker(text string) {
	p := v.panel
	p.list = &modelList{filter: text}
	cur := 0
	if text == "" {
		for i, r := range v.pickerContent().rows {
			if r.current {
				cur = i
				break
			}
		}
	}
	v.moveList(cur)
}

// currentModel is the model the panel stands on, for the "●" mark: the chosen
// one, or the effective one while nothing was chosen. Empty for a mixed group.
func (p *modelsPanel) currentModel() string { return p.modelID() }

// providerOf is the part of an OpenCode or Pi id before the first "/".
func providerOf(host, id string) string {
	if host != "opencode" && host != "pi" {
		return ""
	}
	provider, _, found := strings.Cut(id, "/")
	if !found {
		return ""
	}
	return catalogText(provider)
}

// pickerContent lays the picker out for the current search. OpenCode and Pi get
// one section per provider in order of first appearance; the other CLIs get one
// flat list. The model the panel stands on is listed even when the CLI's own
// list lacks it, so it can always be seen and kept.
func (v *modelsView) pickerContent() pickerContent {
	p, host := v.panel, v.host
	models := append([]catalogModel(nil), p.catalog...)
	current := p.currentModel()
	if current != "" {
		found := false
		for _, m := range models {
			found = found || m.ID == current
		}
		if !found {
			models = append([]catalogModel{{ID: current, Provider: providerOf(host, current)}}, models...)
		}
	}
	sections := host == "opencode" || host == "pi"
	filter := strings.ToLower(p.list.filter)
	var c pickerContent
	var order []string
	groups := map[string][]pickRow{}
	for _, m := range models {
		r := pickRow{id: m.ID, name: m.Name, provider: m.Provider, current: m.ID == current}
		if filter != "" && !strings.Contains(strings.ToLower(r.name), filter) &&
			!strings.Contains(strings.ToLower(r.id), filter) && !strings.Contains(strings.ToLower(r.provider), filter) {
			continue
		}
		key := ""
		if sections {
			key = m.Provider
		}
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], r)
	}
	for i, key := range order {
		if sections {
			if i > 0 {
				c.lines = append(c.lines, pickLine{blank: true})
			}
			header := host + " models"
			if key != "" {
				header = providerTitle(key)
			}
			c.lines = append(c.lines, pickLine{header: header})
		}
		for _, r := range groups[key] {
			c.lines = append(c.lines, pickLine{row: len(c.rows)})
			c.rows = append(c.rows, r)
		}
	}
	return c
}

// listStatus is the picker's status row: the spinner text while the query runs,
// or why the list is unavailable. It is empty when neither applies.
func (v *modelsView) listStatus() (loading bool, text string) {
	p := v.panel
	switch {
	case p.noCatalog:
		return false, ""
	case v.catalogLoading():
		return true, "Loading models…"
	case p.catalogErr != "":
		return false, "Model list unavailable: " + p.catalogErr
	}
	return false, ""
}

// pickerBoxHeight is the panel's height: everything between the CLI row and the footer.
func (v *modelsView) pickerBoxHeight() int { return max(v.height-4, pickerRowsFix+1) }

// listHeight is the number of lines the scrolling area has room for.
func (v *modelsView) listHeight() int { return max(v.pickerBoxHeight()-pickerRowsFix, 1) }

// moveList moves the cursor among the model rows and the two bottom entries,
// and keeps a model row, with its section header above it, in view.
func (v *modelsView) moveList(to int) {
	l, c := v.panel.list, v.pickerContent()
	n := len(c.rows)
	l.cur = min(max(to, 0), n+1)
	h := v.listHeight()
	if l.cur < n {
		at := 0
		for i, ln := range c.lines {
			if !ln.blank && ln.header == "" && ln.row == l.cur {
				at = i
			}
		}
		top := at
		if at > 0 && c.lines[at-1].header != "" {
			top = at - 1 // a header scrolls with its first row
		}
		switch {
		case top < l.off:
			l.off = top
		case at >= l.off+h:
			l.off = at - h + 1
		}
	}
	if l.cur >= n {
		l.off = len(c.lines) // the bottom entries come after the last model: show the end
	}
	l.off = min(max(l.off, 0), max(len(c.lines)-h, 0))
}

// listKey handles a key while the picker is open.
func (v *modelsView) listKey(msg tea.KeyPressMsg) (tea.Cmd, action) {
	p, l, name := v.panel, v.panel.list, msg.String()
	n := len(v.pickerContent().rows)
	switch name {
	case "esc":
		p.list = nil
	case "up":
		if l.cur >= n && n > 0 {
			v.moveList(n - 1)
		} else {
			v.moveList(l.cur - 1)
		}
	case "down":
		if l.cur <= n {
			v.moveList(l.cur + 1)
		}
	case "left":
		if l.cur > n {
			v.moveList(n)
		}
	case "right":
		if l.cur == n {
			v.moveList(n + 1)
		}
	case "pgup":
		v.moveList(min(l.cur, n) - v.listHeight())
	case "pgdown":
		v.moveList(min(l.cur+v.listHeight(), n))
	case "home":
		v.moveList(0)
	case "end":
		v.moveList(n + 1)
	case "backspace":
		if _, size := utf8.DecodeLastRuneInString(l.filter); size > 0 {
			l.filter = l.filter[:len(l.filter)-size]
			l.cur, l.off = 0, 0
		}
	case "enter":
		return v.choose(l.cur)
	default:
		if msg.Text != "" {
			l.filter += msg.Text
			l.cur, l.off = 0, 0
		}
	}
	return nil, action{nav: navNone}
}

// choose applies the picker's entry at cursor and returns to the panel with the
// focus on Model.
func (v *modelsView) choose(cur int) (tea.Cmd, action) {
	p := v.panel
	c := v.pickerContent()
	n := len(c.rows)
	p.list, p.row, p.message = nil, 0, ""
	switch {
	case cur == n+1: // Other…
		start := p.modelValue() // read before the field takes over
		if p.modelMixed && !p.touched {
			start = ""
		}
		p.textMode = true
		p.model.SetValue(start)
		p.model.CursorEnd()
		return p.model.Focus(), action{nav: navNone}
	case cur == n: // release default
		p.chosen, p.touched, p.textMode = "", true, false
	default:
		p.chosen, p.touched, p.textMode = c.rows[cur].id, true, false
	}
	p.trimEfforts(true)
	return nil, action{nav: navNone}
}

// pickSeg is a piece of a panel line with its style.
type pickSeg struct {
	text string
	st   lipgloss.Style
}

// paint draws one panel line of exactly width columns: one column of padding on
// each side and the surface background under everything.
func paintPanelLine(th *appTheme, width int, segs ...pickSeg) string {
	bg := th.Surface.GetBackground()
	var b strings.Builder
	used := 0
	put := func(text string, st lipgloss.Style) {
		if text == "" {
			return
		}
		b.WriteString(st.Background(bg).Render(text))
		used += utf8.RuneCountInString(text)
	}
	put(" ", th.Text)
	for _, s := range segs {
		put(s.text, s.st)
	}
	put(strings.Repeat(" ", max(width-used, 0)), th.Text)
	return b.String()
}

// spacerFor is the filler that right-aligns what follows, inside inner columns.
func spacerFor(inner int, left, right string) string {
	return strings.Repeat(" ", max(inner-utf8.RuneCountInString(left)-utf8.RuneCountInString(right), 1))
}

// pickerLines draws the panel: title, search, status, the scrolling list and the
// bottom entries, in exactly pickerBoxHeight lines of exactly the view's width.
func (v *modelsView) pickerLines(c viewCtx) []string {
	th, p := c.Theme, v.panel
	W := max(c.Width, 20)
	inner := W - 2
	blank := func() string { return paintPanelLine(th, W) }
	who := p.role
	if p.group != "" {
		who = strings.ToUpper(p.group[:1]) + p.group[1:] + " group"
	}
	title, ctx := "Select model", truncateRunes(" · "+v.host+" · "+who, max(inner-len("Select model")-len("esc")-2, 1))
	lines := []string{paintPanelLine(th, W,
		pickSeg{title, th.Text.Bold(true)}, pickSeg{ctx, th.Muted},
		pickSeg{spacerFor(inner, title+ctx, "esc"), th.Text}, pickSeg{"esc", th.Muted})}
	lines = append(lines, blank())
	if p.list.filter == "" {
		lines = append(lines, paintPanelLine(th, W, pickSeg{"Search", th.Muted}))
	} else {
		lines = append(lines, paintPanelLine(th, W, pickSeg{truncateRunes("Search: "+p.list.filter, inner), th.Text}))
	}
	if loading, text := v.listStatus(); loading {
		lines = append(lines, paintPanelLine(th, W, pickSeg{c.Spinner + " ", th.Text}, pickSeg{truncateRunes(text, inner-2), th.Muted}))
	} else if text != "" {
		lines = append(lines, paintPanelLine(th, W, pickSeg{truncateRunes(text, inner), th.Danger}))
	} else {
		lines = append(lines, blank())
	}

	v.moveList(p.list.cur) // the content or the room may have changed since the last key
	content := v.pickerContent()
	h := v.listHeight()
	shown := 0
	for i := p.list.off; i < min(p.list.off+h, len(content.lines)); i++ {
		ln := content.lines[i]
		shown++
		switch {
		case ln.blank:
			lines = append(lines, blank())
		case ln.header != "":
			lines = append(lines, paintPanelLine(th, W, pickSeg{truncateRunes(ln.header, inner), th.Accent}))
		default:
			lines = append(lines, pickerRow(th, W, content.rows[ln.row], ln.row == p.list.cur, v.host))
		}
	}
	for ; shown < h; shown++ {
		lines = append(lines, blank())
	}
	return append(lines, v.pickerBottom(c, W, len(content.rows)))
}

// pickerRow draws one model row. The selected row is highlighted across the
// whole line, or marked ">" under NO_COLOR.
func pickerRow(th *appTheme, W int, r pickRow, selected bool, host string) string {
	inner := W - 2
	cursor := "  "
	if selected && th.NoColor {
		cursor = "> "
	}
	mark := "  "
	if r.current {
		mark = currentMarker
	}
	id := r.id
	if host == "opencode" || host == "pi" {
		id = r.shortID()
	}
	primary := id
	if r.name != "" {
		primary = r.name
	}
	budget := inner - 4
	primary = truncateRunes(primary, max(budget, 1))
	secondary := ""
	if r.name != "" {
		if room := budget - utf8.RuneCountInString(primary) - 2; room >= 4 {
			secondary = truncateRunes(id, room)
		}
	}
	if selected && !th.NoColor {
		text := cursor + mark + primary
		if secondary != "" {
			text += "  " + secondary
		}
		return th.ButtonOn.Render(" " + text + strings.Repeat(" ", max(W-1-utf8.RuneCountInString(text), 0)))
	}
	segs := []pickSeg{{cursor + mark + primary, th.Text}}
	if secondary != "" {
		segs = append(segs, pickSeg{"  " + secondary, th.Muted})
	}
	return paintPanelLine(th, W, segs...)
}

// pickerBottom draws the last line: the two entries that always stay, and the
// key hints on the right.
func (v *modelsView) pickerBottom(c viewCtx, W, n int) string {
	th, l := c.Theme, v.panel.list
	var segs []pickSeg
	used := 0
	for i, label := range []string{effortReleaseText, otherModelText} {
		sel := l.cur == n+i
		text := "  " + label
		st := th.Text
		switch {
		case sel && th.NoColor:
			text, st = "> "+label, th.Accent
		case sel:
			text, st = " "+label+" ", th.ButtonOn
		}
		segs = append(segs, pickSeg{text, st}, pickSeg{"  ", th.Text})
		used += utf8.RuneCountInString(text) + 2
	}
	gap := max(W-2-used-utf8.RuneCountInString(pickerHint), 1)
	return paintPanelLine(th, W, append(segs, pickSeg{strings.Repeat(" ", gap), th.Text}, pickSeg{pickerHint, th.Muted})...)
}
