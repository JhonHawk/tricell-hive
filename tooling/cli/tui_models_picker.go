// tui_models_picker.go is the model picker of the Models view (#46, D9-A and
// its corrections): a bordered box over the edit panel and the table, opened
// from the Model field. It lists the CLI's models, by provider on the CLIs whose
// ids carry one and flat on the rest, and ends with two entries, "release
// default" and "Other…". The box stops two columns short of the view's width.
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
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
	if text != "" {
		v.resetPickerCursor()
	}
}

// resetPickerCursor puts the cursor on the first match after the search text
// changed. With no match it rests on nothing (-1), so Enter cannot choose an
// entry the person has not reached; ↓ then goes to the bottom entries.
func (v *modelsView) resetPickerCursor() {
	l := v.panel.list
	l.cur, l.off = 0, 0
	if l.filter != "" && len(v.pickerContent().rows) == 0 {
		l.cur = -1
	}
}

// pickerEntryKey names the highlighted entry: "m:<id>" for a model, "default"
// or "other" for a bottom entry, "" for none or when the picker is closed.
func (v *modelsView) pickerEntryKey() string {
	p := v.panel
	if p == nil || p.list == nil || p.list.cur < 0 {
		return ""
	}
	rows := v.pickerContent().rows
	switch cur := p.list.cur; {
	case cur < len(rows):
		return "m:" + rows[cur].id
	case cur == len(rows):
		return "default"
	}
	return "other"
}

// restorePickerEntry puts the cursor back on the entry key names, after the
// list it is part of changed (a model list that arrived while the box was open).
func (v *modelsView) restorePickerEntry(key string) {
	if key == "" || v.panel == nil || v.panel.list == nil {
		return
	}
	rows := v.pickerContent().rows
	cur := -1
	switch key {
	case "default":
		cur = len(rows)
	case "other":
		cur = len(rows) + 1
	default:
		for i, r := range rows {
			if "m:"+r.id == key {
				cur = i
			}
		}
	}
	if cur >= 0 {
		v.moveList(cur)
	}
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
	lo := 0
	if n == 0 && l.filter != "" {
		lo = -1 // a search with no match: the cursor may rest on nothing
	}
	l.cur = min(max(to, lo), n+1)
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
			v.resetPickerCursor()
		}
	case "enter":
		if l.cur < 0 {
			return nil, action{nav: navNone} // nothing is highlighted
		}
		return v.choose(l.cur)
	default:
		if msg.Text != "" {
			l.filter += msg.Text
			v.resetPickerCursor()
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
	if reset := p.trimEfforts(true); reset != "" {
		p.message, p.messageOK = "Effort reset to release default: "+p.modelID()+" has no "+reset, true
	}
	return nil, action{nav: navNone}
}

// pickerMargin is how many columns the box leaves free at the right of the
// view. A terminal that draws an ambiguous-width character ("·", "●", "…", "▸")
// as two columns pushes a line past its measured width by a few columns; the
// margin keeps the border from wrapping when it does.
const pickerMargin = 2

// pickerLines draws the bordered box: the title in the top border, search,
// status, the scrolling list and the bottom entries, in exactly pickerBoxHeight
// lines, each pickerMargin columns narrower than the view.
func (v *modelsView) pickerLines(c viewCtx) []string {
	th, p := c.Theme, v.panel
	W := max(c.Width-pickerMargin, 24)
	inner := W - 4 // inside the borders and one column of padding on each side
	border := func(s string) string { return th.Muted.Render(s) }
	boxed := func(rendered string, plain int) string {
		return border("│ ") + rendered + strings.Repeat(" ", max(inner-plain, 0)) + border(" │")
	}
	who := p.role
	if p.group != "" {
		who = groupTitle(p.group) + " group"
	}
	title := "Select model"
	ctx := truncateRunes(" · "+v.host+" · "+who, max(W-len(title)-len(" esc ")-6, 1))
	fill := max(W-2-1-len(title)-textWidth(ctx)-1-len(" esc "), 0)
	lines := []string{border("┌") + th.Text.Render(" ") + th.Title.Render(title) + th.Muted.Render(ctx+" ") +
		border(strings.Repeat("─", fill)) + th.Muted.Render(" esc ") + border("┐")}

	if p.list.filter == "" {
		lines = append(lines, boxed(th.Muted.Render("Search"), len("Search")))
	} else {
		text := truncateRunes("Search: "+p.list.filter, inner)
		lines = append(lines, boxed(th.Text.Render(text), textWidth(text)))
	}
	if loading, text := v.listStatus(); loading {
		text = truncateRunes(text, inner-2)
		lines = append(lines, boxed(c.Spinner+" "+th.Muted.Render(text), 2+textWidth(text)))
	} else if text != "" {
		text = truncateRunes(text, inner)
		lines = append(lines, boxed(th.Danger.Render(text), textWidth(text)))
	} else {
		lines = append(lines, boxed("", 0))
	}

	v.moveList(p.list.cur) // the content or the room may have changed since the last key
	content := v.pickerContent()
	h := v.listHeight()
	shown := 0
	if len(content.lines) == 0 && p.list.filter != "" {
		lines = append(lines, boxed(th.Muted.Render("No matching models"), len("No matching models")))
		shown++
	}
	for i := p.list.off; i < min(p.list.off+h, len(content.lines)); i++ {
		ln := content.lines[i]
		shown++
		switch {
		case ln.blank:
			lines = append(lines, boxed("", 0))
		case ln.header != "":
			text := truncateRunes(ln.header, inner)
			lines = append(lines, boxed(th.Accent.Render(text), textWidth(text)))
		default:
			rendered, plain := pickerRow(th, inner, content.rows[ln.row], ln.row == p.list.cur, v.host)
			lines = append(lines, boxed(rendered, plain))
		}
	}
	for ; shown < h; shown++ {
		lines = append(lines, boxed("", 0))
	}
	lines = append(lines, v.pickerBottom(c, len(content.rows), inner, boxed))
	return append(lines, border("└"+strings.Repeat("─", W-2)+"┘"))
}

// pickerRow draws one model row of width inner and returns its visible width.
// The selected row is highlighted across the line, or marked ">" under NO_COLOR.
func pickerRow(th *appTheme, inner int, r pickRow, selected bool, host string) (string, int) {
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
		if room := budget - textWidth(primary) - 2; room >= 4 {
			secondary = truncateRunes(id, room)
		}
	}
	text := cursor + mark + primary
	if secondary != "" {
		text += "  " + secondary
	}
	if selected && !th.NoColor {
		text += strings.Repeat(" ", max(inner-textWidth(text), 0))
		return th.ButtonOn.Render(text), textWidth(text)
	}
	out := th.Text.Render(cursor + mark + primary)
	if secondary != "" {
		out += th.Muted.Render("  " + secondary)
	}
	return out, textWidth(text)
}

// pickerBottom draws the last line inside the box: the two entries that always
// stay, and the key hints on the right.
func (v *modelsView) pickerBottom(c viewCtx, n, inner int, boxed func(string, int) string) string {
	th, l := c.Theme, v.panel.list
	var out string
	plain := 0
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
		out += st.Render(text) + "  "
		plain += textWidth(text) + 2
	}
	hint := pickerHint
	if l.cur >= 0 && l.cur < n {
		// The cursor's place among the models, when the row has room for it.
		if with := fmt.Sprintf("%d of %d  %s", l.cur+1, n, pickerHint); inner-plain-textWidth(with) >= 1 {
			hint = with
		}
	}
	gap := max(inner-plain-textWidth(hint), 1)
	return boxed(out+strings.Repeat(" ", gap)+th.Muted.Render(hint), plain+gap+textWidth(hint))
}
