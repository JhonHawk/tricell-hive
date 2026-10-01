// tui_models_picker.go is the model picker of the Models view (#46, D9-A): a
// bordered box over the edit panel and the table, opened from the Model field.
// It lists the models in use on the CLI first, then the CLI's own list by
// provider, and ends with two entries, "release default" and "Other…".
package main

import (
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
	pickerInUse   = "In use on "
	currentMarker = "● "
	pickerHint    = "↑↓ enter esc"
)

// pickRow is one selectable model.
type pickRow struct {
	id, name, provider string
	current            bool
	showProvider       bool // the provider is not the section's header
}

// pickLine is one line of the scrolling area: a section header, or a row.
type pickLine struct {
	header string
	row    int
}

type pickerContent struct {
	rows  []pickRow
	lines []pickLine
}

// openList opens the picker, with text as the first characters of the search.
func (p *modelsPanel) openList(text string) {
	p.list = &modelList{filter: text}
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

// shortID is the id as the picker shows it: without its "<provider>/" prefix
// when the row has a provider. The value chosen stays the full id.
func (r pickRow) shortID() string {
	if r.provider != "" {
		return strings.TrimPrefix(r.id, r.provider+"/")
	}
	return r.id
}

// pickerContent lays the picker out for the current search: the models in use
// on the CLI, then the CLI's models by provider, in order of first appearance.
// A model in use is not repeated below.
func (v *modelsView) pickerContent() pickerContent {
	p, host := v.panel, v.host
	byID := map[string]catalogModel{}
	for _, m := range p.catalog {
		byID[m.ID] = m
	}
	current := p.currentModel()
	filter := strings.ToLower(p.list.filter)
	match := func(r pickRow) bool {
		return filter == "" || strings.Contains(strings.ToLower(r.name), filter) ||
			strings.Contains(strings.ToLower(r.id), filter) || strings.Contains(strings.ToLower(r.provider), filter)
	}
	var c pickerContent
	add := func(header string, rows []pickRow) {
		var kept []pickRow
		for _, r := range rows {
			if match(r) {
				kept = append(kept, r)
			}
		}
		if len(kept) == 0 {
			return
		}
		c.lines = append(c.lines, pickLine{header: header})
		for _, r := range kept {
			c.lines = append(c.lines, pickLine{row: len(c.rows)})
			c.rows = append(c.rows, r)
		}
	}
	inUse := map[string]bool{}
	var used []pickRow
	for _, r := range v.hostRows() {
		model, _ := effectiveParts(r)
		if model == "" || inUse[model] {
			continue
		}
		inUse[model] = true
		row := pickRow{id: model, provider: providerOf(host, model), current: model == current}
		if m, ok := byID[model]; ok {
			row.name, row.provider = m.Name, m.Provider
		}
		row.showProvider = row.provider != ""
		used = append(used, row)
	}
	add(pickerInUse+host, used)
	var order []string
	groups := map[string][]pickRow{}
	for _, m := range p.catalog {
		if inUse[m.ID] {
			continue
		}
		key := m.Provider
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], pickRow{id: m.ID, name: m.Name, provider: m.Provider, current: m.ID == current})
	}
	for _, key := range order {
		header := key
		if key == "" {
			header = host + " models"
		}
		add(header, groups[key])
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

// pickerBoxHeight is the box's height: everything between the CLI row and the footer.
func (v *modelsView) pickerBoxHeight() int { return max(v.height-4, 5) }

// listHeight is the number of lines the scrolling area has room for: the box
// less its borders, the search row, the bottom row and the status row.
func (v *modelsView) listHeight() int {
	h := v.pickerBoxHeight() - 2 - 1 - 1
	if loading, text := v.listStatus(); loading || text != "" {
		h--
	}
	return max(h, 1)
}

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
			if ln.header == "" && ln.row == l.cur {
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
		if l.cur < n || l.cur == n {
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

// pickerLines draws the box: borders, search, status, the scrolling sections and
// the bottom entries, in exactly pickerBoxHeight lines of the view's width.
func (v *modelsView) pickerLines(c viewCtx) []string {
	th, p := c.Theme, v.panel
	W := max(c.Width, 20)
	inner := W - 4
	border := func(s string) string { return th.Muted.Render(s) }
	boxed := func(rendered string, plain int) string {
		return border("│ ") + rendered + strings.Repeat(" ", max(inner-plain, 0)) + border(" │")
	}
	who := p.role
	if p.group != "" {
		who = strings.ToUpper(p.group[:1]) + p.group[1:] + " group"
	}
	left := truncateRunes(" Select model · "+v.host+" · "+who+" ", max(W-10, 4))
	fill := max(W-2-utf8.RuneCountInString(left)-utf8.RuneCountInString(" esc "), 0)
	lines := []string{border("┌") + th.Title.Render(left) + border(strings.Repeat("─", fill)) + th.Muted.Render(" esc ") + border("┐")}

	search := p.list.filter
	if search == "" {
		lines = append(lines, boxed(th.Muted.Render("Search"), len("Search")))
	} else {
		text := truncateRunes("Search: "+search, inner)
		lines = append(lines, boxed(th.Text.Render(text), utf8.RuneCountInString(text)))
	}
	if loading, text := v.listStatus(); loading {
		text = truncateRunes(text, inner-2)
		lines = append(lines, boxed(c.Spinner+" "+th.Muted.Render(text), 2+utf8.RuneCountInString(text)))
	} else if text != "" {
		text = truncateRunes(text, inner)
		lines = append(lines, boxed(th.Danger.Render(text), utf8.RuneCountInString(text)))
	}

	v.moveList(p.list.cur) // the content or the room may have changed since the last key
	content := v.pickerContent()
	h := v.listHeight()
	shown := 0
	for i := p.list.off; i < min(p.list.off+h, len(content.lines)); i++ {
		ln := content.lines[i]
		shown++
		if ln.header != "" {
			text := truncateRunes(ln.header, inner)
			lines = append(lines, boxed(th.Accent.Render(text), utf8.RuneCountInString(text)))
			continue
		}
		rendered, plain := pickerRow(th, content.rows[ln.row], ln.row == p.list.cur, inner)
		lines = append(lines, boxed(rendered, plain))
	}
	for ; shown < h; shown++ {
		lines = append(lines, boxed("", 0))
	}
	lines = append(lines, v.pickerBottom(c, len(content.rows), inner, boxed))
	return append(lines, border("└"+strings.Repeat("─", W-2)+"┘"))
}

// pickerRow draws one model row of width inner and returns its visible width.
// The selected row is highlighted across the line, or marked ">" under NO_COLOR.
func pickerRow(th *appTheme, r pickRow, selected bool, inner int) (string, int) {
	cursor := "  "
	if selected && th.NoColor {
		cursor = "> "
	}
	mark := "  "
	if r.current {
		mark = currentMarker
	}
	right := ""
	if r.showProvider {
		right = r.provider
	}
	budget := inner - 4
	if right != "" {
		budget -= utf8.RuneCountInString(right) + 2
	}
	primary := r.name
	if primary == "" {
		primary = r.shortID()
	}
	primary = truncateRunes(primary, max(budget, 1))
	secondary := ""
	if r.name != "" {
		if room := budget - utf8.RuneCountInString(primary) - 2; room >= 4 {
			secondary = truncateRunes(r.shortID(), room)
		}
	}
	used := 4 + utf8.RuneCountInString(primary)
	if secondary != "" {
		used += 2 + utf8.RuneCountInString(secondary)
	}
	pad := max(inner-used-utf8.RuneCountInString(right), 0)
	if selected && !th.NoColor {
		line := cursor + mark + primary
		if secondary != "" {
			line += "  " + secondary
		}
		line += strings.Repeat(" ", pad) + right
		return th.ButtonOn.Render(line), utf8.RuneCountInString(line)
	}
	out := th.Text.Render(cursor + mark + primary)
	if secondary != "" {
		out += th.Muted.Render("  " + secondary)
	}
	out += strings.Repeat(" ", pad) + th.Muted.Render(right)
	return out, used + pad + utf8.RuneCountInString(right)
}

// pickerBottom draws the box's last line: the two entries that always stay, and
// the key hints on the right.
func (v *modelsView) pickerBottom(c viewCtx, n, inner int, boxed func(string, int) string) string {
	th, l := c.Theme, v.panel.list
	labels := []string{effortReleaseText, otherModelText}
	var out string
	plain := 0
	for i, label := range labels {
		sel := l.cur == n+i
		var text string
		switch {
		case sel && th.NoColor:
			text = "> " + label
			out += th.Accent.Render(text)
		case sel:
			text = " " + label + " "
			out += th.ButtonOn.Render(text)
		default:
			text = "  " + label
			out += th.Text.Render(text)
		}
		plain += utf8.RuneCountInString(text)
		out += "  "
		plain += 2
	}
	hint := pickerHint
	gap := max(inner-plain-utf8.RuneCountInString(hint), 1)
	return boxed(out+strings.Repeat(" ", gap)+th.Muted.Render(hint), plain+gap+utf8.RuneCountInString(hint))
}
