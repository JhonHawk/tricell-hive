package management

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"tricell-hive/integrations/hostevent"
)

// This file is the safe editor for the shared JSON hook files (Claude Code's
// settings.json, Codex's hooks.json, Cursor's hooks.json). Other tools keep
// their hooks and settings there, so the editor never parses the file into a
// value and writes it back: it finds the byte spans of the values it needs and
// splices text in or out, so every byte outside its own entries stays exactly
// as it was. It recognizes as Hive's only the entries whose command is
// `'<hive>' hook <host> <event>` (hostevent.IsHiveCommand). A file it cannot
// read as the expected shape is refused, never rewritten.

// jsonValue is a parsed value with its byte span [start, end) in the source.
type jsonValue struct {
	kind       byte // '{', '[' or 'v' (string, number or literal)
	start, end int
	members    []jsonMember // objects
	elems      []*jsonValue // arrays
}

type jsonMember struct {
	key   string
	start int // offset of the key's opening quote
	val   *jsonValue
}

func (v *jsonValue) member(key string) (*jsonMember, error) {
	var found *jsonMember
	for i := range v.members {
		if v.members[i].key == key {
			if found != nil {
				return nil, fmt.Errorf("duplicate key %q", key)
			}
			found = &v.members[i]
		}
	}
	return found, nil
}

const jsonMaxDepth = 100

type spanParser struct {
	b []byte
	i int
}

func (p *spanParser) skipSpace() {
	for p.i < len(p.b) && (p.b[p.i] == ' ' || p.b[p.i] == '\t' || p.b[p.i] == '\n' || p.b[p.i] == '\r') {
		p.i++
	}
}

func (p *spanParser) str() error {
	p.i++ // opening quote
	for p.i < len(p.b) {
		switch p.b[p.i] {
		case '\\':
			p.i += 2
		case '"':
			p.i++
			return nil
		default:
			p.i++
		}
	}
	return errors.New("unterminated string")
}

func (p *spanParser) value(depth int) (*jsonValue, error) {
	if depth > jsonMaxDepth {
		return nil, errors.New("nesting too deep")
	}
	p.skipSpace()
	if p.i >= len(p.b) {
		return nil, errors.New("unexpected end")
	}
	v := &jsonValue{start: p.i}
	switch p.b[p.i] {
	case '{':
		v.kind = '{'
		p.i++
		for {
			p.skipSpace()
			if p.i < len(p.b) && p.b[p.i] == '}' {
				p.i++
				break
			}
			if p.i >= len(p.b) || p.b[p.i] != '"' {
				return nil, errors.New("expected a key")
			}
			keyStart := p.i
			if err := p.str(); err != nil {
				return nil, err
			}
			var key string
			if err := json.Unmarshal(p.b[keyStart:p.i], &key); err != nil {
				return nil, err
			}
			p.skipSpace()
			if p.i >= len(p.b) || p.b[p.i] != ':' {
				return nil, errors.New("expected a colon")
			}
			p.i++
			val, err := p.value(depth + 1)
			if err != nil {
				return nil, err
			}
			v.members = append(v.members, jsonMember{key: key, start: keyStart, val: val})
			p.skipSpace()
			if p.i < len(p.b) && p.b[p.i] == ',' {
				p.i++
			}
		}
	case '[':
		v.kind = '['
		p.i++
		for {
			p.skipSpace()
			if p.i < len(p.b) && p.b[p.i] == ']' {
				p.i++
				break
			}
			elem, err := p.value(depth + 1)
			if err != nil {
				return nil, err
			}
			v.elems = append(v.elems, elem)
			p.skipSpace()
			if p.i < len(p.b) && p.b[p.i] == ',' {
				p.i++
			}
		}
	case '"':
		v.kind = 'v'
		if err := p.str(); err != nil {
			return nil, err
		}
	default:
		v.kind = 'v'
		for p.i < len(p.b) && !strings.ContainsRune(",]} \t\r\n", rune(p.b[p.i])) {
			p.i++
		}
	}
	v.end = p.i
	return v, nil
}

// parseJSONSpans parses b, which must be one valid JSON value, into spans.
func parseJSONSpans(b []byte) (*jsonValue, error) {
	if !json.Valid(b) {
		return nil, errors.New("not valid JSON")
	}
	p := &spanParser{b: b}
	v, err := p.value(0)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type textEdit struct {
	start, end int
	text       string
}

func applyTextEdits(b []byte, edits []textEdit) ([]byte, error) {
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start > edits[j].start
		}
		return edits[i].end > edits[j].end
	})
	out := append([]byte(nil), b...)
	limit := len(b)
	for _, e := range edits {
		if e.start > e.end || e.end > limit {
			return nil, errors.New("overlapping edits")
		}
		out = append(out[:e.start], append([]byte(e.text), out[e.end:]...)...)
		limit = e.start
	}
	return out, nil
}

// jsonStyle is the document's formatting, read once and used for new text.
type jsonStyle struct {
	multi bool   // the document has line breaks
	unit  string // one indentation step
	nl    string
}

func detectStyle(b []byte, root *jsonValue) jsonStyle {
	st := jsonStyle{unit: "  ", nl: "\n"}
	if bytes.Contains(b, []byte("\r\n")) {
		st.nl = "\r\n"
	}
	st.multi = bytes.IndexByte(b[root.start:root.end], '\n') >= 0 || len(root.members) == 0
	if first := firstChildStart(root); first >= 0 {
		gap := string(b[root.start+1 : first])
		if i := strings.LastIndexByte(gap, '\n'); i >= 0 && i+1 < len(gap) {
			st.unit = gap[i+1:]
		}
	}
	return st
}

func firstChildStart(v *jsonValue) int {
	switch {
	case v.kind == '{' && len(v.members) > 0:
		return v.members[0].start
	case v.kind == '[' && len(v.elems) > 0:
		return v.elems[0].start
	}
	return -1
}

func lastChildEnd(v *jsonValue) int {
	switch {
	case v.kind == '{' && len(v.members) > 0:
		return v.members[len(v.members)-1].val.end
	case v.kind == '[' && len(v.elems) > 0:
		return v.elems[len(v.elems)-1].end
	}
	return -1
}

// layout is how a container lays out its children, so new children look like
// the ones around them.
type layout struct {
	multi  bool
	indent string // indentation of a child's line
	sep    string // text between the opening bracket or a comma and a child
	close  string // text between the last child and the closing bracket
}

func containerLayout(b []byte, v *jsonValue, selfIndent string, st jsonStyle) layout {
	if first := firstChildStart(v); first >= 0 {
		gap := string(b[v.start+1 : first])
		l := layout{sep: gap, close: string(b[lastChildEnd(v) : v.end-1])}
		if i := strings.LastIndexByte(gap, '\n'); i >= 0 {
			l.multi, l.indent = true, gap[i+1:]
		}
		return l
	}
	if st.multi {
		indent := selfIndent + st.unit
		return layout{multi: true, indent: indent, sep: st.nl + indent, close: st.nl + selfIndent}
	}
	return layout{}
}

func marshalText(v any, prefix, unit string, multi bool) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if multi {
		enc.SetIndent(prefix, unit)
	}
	if err := enc.Encode(v); err != nil {
		panic(err) // the values are Hive's own structs
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

type renderedPair struct{ key, value string }

func renderObject(pairs []renderedPair, indent string, st jsonStyle, multi bool) string {
	parts := make([]string, len(pairs))
	colon := ":"
	if multi {
		colon = ": "
	}
	for i, p := range pairs {
		parts[i] = marshalText(p.key, "", "", false) + colon + p.value
	}
	if !multi {
		return "{" + strings.Join(parts, ",") + "}"
	}
	inner := indent + st.unit
	return "{" + st.nl + inner + strings.Join(parts, ","+st.nl+inner) + st.nl + indent + "}"
}

func renderArray(items []string, indent string, st jsonStyle, multi bool) string {
	if !multi {
		return "[" + strings.Join(items, ",") + "]"
	}
	inner := indent + st.unit
	return "[" + st.nl + inner + strings.Join(items, ","+st.nl+inner) + st.nl + indent + "]"
}

// insertChildren returns the edit that appends texts to a container, each as a
// finished child (an element, or a `"key": value` member).
func insertChildren(v *jsonValue, l layout, texts []string) textEdit {
	return insertChildrenNoting(v, l, texts, nil, "", nil)
}

// insertChildrenNoting is insertChildren that also records, when the container
// had no children, the whitespace between its brackets, under key in delta.
func insertChildrenNoting(v *jsonValue, l layout, texts []string, src []byte, key string, delta *HostHooksRecord) textEdit {
	if delta != nil && len(texts) > 0 && lastChildEnd(v) < 0 {
		if w := string(src[v.start+1 : v.end-1]); w != "" {
			if delta.Interiors == nil {
				delta.Interiors = map[string]string{}
			}
			delta.Interiors[key] = w
		}
	}
	if len(texts) == 0 {
		return textEdit{}
	}
	if last := lastChildEnd(v); last >= 0 {
		var sb strings.Builder
		for _, t := range texts {
			sb.WriteString("," + l.sep + t)
		}
		return textEdit{start: last, end: last, text: sb.String()}
	}
	return textEdit{start: v.start + 1, end: v.end - 1, text: l.sep + strings.Join(texts, ","+l.sep) + l.close}
}

// removeChildren returns the edits that delete the marked children of a
// container, with their commas and the whitespace that led to them, so the
// neighbours keep their own formatting. spans are the children's [start, end).
//
// When every child goes, empty is the text between the brackets afterwards: the
// whitespace the container had before Hive filled it, so the user's `[ ]` is not
// rewritten as `[]`.
func removeChildren(v *jsonValue, spans [][2]int, remove []bool, empty string) []textEdit {
	n := len(spans)
	all := n > 0
	for _, r := range remove {
		all = all && r
	}
	if all {
		return []textEdit{{start: v.start + 1, end: v.end - 1, text: empty}}
	}
	var edits []textEdit
	for i := 0; i < n; {
		if !remove[i] {
			i++
			continue
		}
		j := i
		for j+1 < n && remove[j+1] {
			j++
		}
		if j+1 < n {
			edits = append(edits, textEdit{start: spans[i][0], end: spans[j+1][0]})
		} else {
			edits = append(edits, textEdit{start: spans[i-1][1], end: spans[j][1]})
		}
		i = j + 1
	}
	return edits
}

func memberSpans(v *jsonValue) [][2]int {
	out := make([][2]int, len(v.members))
	for i, m := range v.members {
		out[i] = [2]int{m.start, m.val.end}
	}
	return out
}

func elemSpans(v *jsonValue) [][2]int {
	out := make([][2]int, len(v.elems))
	for i, e := range v.elems {
		out[i] = [2]int{e.start, e.end}
	}
	return out
}

// hookRefusal is the cause of a refusal to edit a hook file.
type hookRefusal struct{ cause string }

func (r *hookRefusal) Error() string { return r.cause }

func refuse(format string, a ...any) error { return &hookRefusal{fmt.Sprintf(format, a...)} }

// hookEditResult is what an install edit leaves.
type hookEditResult struct {
	Data    []byte
	Changed bool
	// Delta is what this edit created or wrote, for undoing exactly it.
	Delta HostHooksRecord
	// Replaced holds, per event, the text of Hive's previous entries that the
	// edit rewrote or dropped (a moved binary), so recovery can put them back.
	Replaced map[string][]string
}

// ownEntries returns the indexes of the elements of an event array that are
// Hive's entries for the event.
func ownEntries(b []byte, arr *jsonValue, f hostevent.JSONHookFile, event string, recorded map[string]bool) []int {
	var own []int
	for i, e := range arr.elems {
		if e.kind != '{' {
			continue
		}
		var m map[string]any
		if json.Unmarshal(b[e.start:e.end], &m) != nil {
			continue
		}
		cmds := f.Commands(m)
		if len(cmds) == 1 && hostevent.IsHiveCommand(cmds[0], f.Word, event, recorded) {
			own = append(own, i)
		}
	}
	return own
}

func recordedSet(commands []string) map[string]bool {
	set := map[string]bool{}
	for _, c := range commands {
		set[c] = true
	}
	return set
}

// editJSONHooks adds, or brings up to date, Hive's entries for every event of
// f. cur is the file as it is; exists says whether the file is there at all.
func editJSONHooks(cur []byte, exists bool, f hostevent.JSONHookFile, executable string, prior *HostHooksRecord) (hookEditResult, error) {
	var recorded map[string]bool
	if prior != nil {
		recorded = recordedSet(prior.Commands)
	}
	res := hookEditResult{Data: cur}
	res.Delta = HostHooksRecord{Path: f.Path}
	commands := map[string]string{}
	for _, event := range f.Events {
		commands[event] = hostevent.Command(executable, f.Word, event)
		res.Delta.Commands = append(res.Delta.Commands, commands[event])
	}
	// A file that holds only whitespace is replaced by a new object.
	blank := len(bytes.TrimSpace(cur)) == 0
	src := cur
	if blank {
		src = []byte("{}")
		res.Delta.CreatedFile = !exists
	}
	root, err := parseJSONSpans(src)
	if err != nil {
		return res, refuse("%s", err.Error())
	}
	if root.kind != '{' {
		return res, refuse("the top level is not a JSON object")
	}
	st := detectStyle(src, root)
	rootLay := containerLayout(src, root, "", st)
	var edits []textEdit

	newArray := func(event string, indent string, multi bool) renderedPair {
		item := marshalText(f.Entry(executable, event), indent+st.unit, st.unit, multi)
		return renderedPair{key: event, value: renderArray([]string{item}, indent, st, multi)}
	}

	container := f.ContainerKey()
	hooksM, err := root.member(container)
	if err != nil {
		return res, refuse("%s in the top level object", err.Error())
	}
	switch {
	case hooksM == nil:
		var pairs []renderedPair
		hooksIndent := rootLay.indent
		for _, event := range f.Events {
			pairs = append(pairs, newArray(event, hooksIndent+st.unit, rootLay.multi))
			res.Delta.CreatedEvents = append(res.Delta.CreatedEvents, event)
		}
		colon := ":"
		if rootLay.multi {
			colon = ": "
		}
		var members []string
		if blank && f.RootVersion > 0 {
			members = append(members, marshalText("version", "", "", false)+colon+fmt.Sprint(f.RootVersion))
		}
		members = append(members, marshalText(container, "", "", false)+colon+renderObject(pairs, hooksIndent, st, rootLay.multi))
		edits = append(edits, insertChildrenNoting(root, rootLay, members, src, "root", &res.Delta))
		res.Delta.CreatedHooks = true
	case hooksM.val.kind != '{':
		return res, refuse("%q is not an object", container)
	default:
		hooks := hooksM.val
		hooksLay := containerLayout(src, hooks, rootLay.indent, st)
		var newPairs []string
		for _, event := range f.Events {
			em, err := hooks.member(event)
			if err != nil {
				return res, refuse("%s in %q", err.Error(), container)
			}
			if em == nil {
				p := newArray(event, hooksLay.indent, hooksLay.multi)
				colon := ":"
				if hooksLay.multi {
					colon = ": "
				}
				newPairs = append(newPairs, marshalText(p.key, "", "", false)+colon+p.value)
				res.Delta.CreatedEvents = append(res.Delta.CreatedEvents, event)
				continue
			}
			if em.val.kind != '[' {
				return res, refuse("\"%s.%s\" is not an array", container, event)
			}
			arr := em.val
			arrLay := containerLayout(src, arr, hooksLay.indent, st)
			want := f.Entry(executable, event)
			own := ownEntries(src, arr, f, event, recorded)
			if len(own) == 0 {
				item := marshalText(want, arrLay.indent, st.unit, arrLay.multi)
				edits = append(edits, insertChildrenNoting(arr, arrLay, []string{item}, src, "event:"+event, &res.Delta))
				continue
			}
			var current any
			e := arr.elems[own[0]]
			_ = json.Unmarshal(src[e.start:e.end], &current)
			var wantAny any
			_ = json.Unmarshal([]byte(marshalText(want, "", "", false)), &wantAny)
			rewritten := !reflect.DeepEqual(current, wantAny)
			if rewritten {
				edits = append(edits, textEdit{start: e.start, end: e.end, text: marshalText(want, arrLay.indent, st.unit, arrLay.multi)})
			}
			if rewritten || len(own) > 1 {
				if res.Replaced == nil {
					res.Replaced = map[string][]string{}
				}
				for _, i := range own {
					res.Replaced[event] = append(res.Replaced[event], string(src[arr.elems[i].start:arr.elems[i].end]))
				}
			}
			if len(own) > 1 {
				drop := make([]bool, len(arr.elems))
				for _, i := range own[1:] {
					drop[i] = true
				}
				edits = append(edits, removeChildren(arr, elemSpans(arr), drop, "")...)
			}
		}
		if len(newPairs) > 0 {
			edits = append(edits, insertChildrenNoting(hooks, hooksLay, newPairs, src, "hooks", &res.Delta))
		}
	}
	if len(edits) == 0 {
		return res, nil
	}
	out, err := applyTextEdits(src, edits)
	if err != nil {
		return res, err
	}
	if blank {
		out = append(out, '\n')
	}
	if !json.Valid(out) {
		return res, errors.New("the edit would not leave valid JSON")
	}
	res.Data, res.Changed = out, true
	return res, nil
}

// undoJSONHooks removes Hive's entries recorded in rec, restoring none.
func undoJSONHooks(cur []byte, f hostevent.JSONHookFile, rec HostHooksRecord) hookUndoResult {
	return undoJSONHooksRestoring(cur, f, rec, nil)
}

// hookUndoResult is what an uninstall edit did.
type hookUndoResult struct {
	Data       []byte
	DeleteFile bool
	Removed    int
	// Skipped is non-empty when the file could not be read as the expected
	// shape; nothing is changed then.
	Skipped string
}

// undoJSONHooks removes Hive's entries recorded in rec from the file as it is
// now. It never fails on a file it cannot read: it leaves it and says why.
//
// restore, when not nil, names per event the previous entries of Hive's that an
// interrupted install replaced: the undo puts them back where Hive's entry now
// is instead of leaving the event without one.
func undoJSONHooksRestoring(cur []byte, f hostevent.JSONHookFile, rec HostHooksRecord, restore map[string][]string) hookUndoResult {
	res := hookUndoResult{Data: cur}
	root, err := parseJSONSpans(cur)
	if err != nil {
		res.Skipped = "the file is not valid JSON"
		return res
	}
	if root.kind != '{' {
		res.Skipped = "the top level is not a JSON object"
		return res
	}
	recorded := recordedSet(rec.Commands)
	created := map[string]bool{}
	for _, e := range rec.CreatedEvents {
		created[e] = true
	}
	hooksM, err := root.member(f.ContainerKey())
	if err != nil || hooksM == nil || hooksM.val.kind != '{' {
		if err != nil {
			res.Skipped = "the file repeats a key"
		}
		return res
	}
	hooks := hooksM.val
	events := map[string]bool{}
	for _, e := range f.Events {
		events[e] = true
	}
	for _, c := range rec.Commands {
		if _, w, e, ok := hostevent.ParseCommand(c); ok && w == f.Word {
			events[e] = true
		}
	}
	var edits []textEdit
	dropMember := make([]bool, len(hooks.members))
	for mi, m := range hooks.members {
		if !events[m.key] || m.val.kind != '[' {
			continue
		}
		if _, err := hooks.member(m.key); err != nil {
			res.Skipped = "the file repeats a key"
			return res
		}
		own := ownEntries(cur, m.val, f, m.key, recorded)
		if len(own) == 0 {
			continue
		}
		res.Removed += len(own)
		if prev := restore[m.key]; len(prev) > 0 {
			// First own entry becomes the previous entries; the others go.
			lay := containerLayout(cur, m.val, "", jsonStyle{})
			first := m.val.elems[own[0]]
			edits = append(edits, textEdit{start: first.start, end: first.end, text: strings.Join(prev, ","+lay.sep)})
			if len(own) > 1 {
				drop := make([]bool, len(m.val.elems))
				for _, i := range own[1:] {
					drop[i] = true
				}
				edits = append(edits, removeChildren(m.val, elemSpans(m.val), drop, "")...)
			}
			continue
		}
		if len(own) == len(m.val.elems) && created[m.key] {
			dropMember[mi] = true
			continue
		}
		drop := make([]bool, len(m.val.elems))
		for _, i := range own {
			drop[i] = true
		}
		edits = append(edits, removeChildren(m.val, elemSpans(m.val), drop, rec.Interiors["event:"+m.key])...)
	}
	allDropped := len(hooks.members) > 0
	for _, d := range dropMember {
		allDropped = allDropped && d
	}
	switch {
	case allDropped && rec.CreatedHooks:
		// Every member of "hooks" was Hive's, so the object goes too.
		rootDrop := make([]bool, len(root.members))
		only := true
		for i, m := range root.members {
			switch {
			case m.start == hooksM.start:
				rootDrop[i] = true
			case m.key == "version" && f.RootVersion > 0 && string(cur[m.val.start:m.val.end]) == fmt.Sprint(f.RootVersion):
				// A version Hive wrote into a file it created goes with the file.
			default:
				only = false
			}
		}
		if only && rec.CreatedFile {
			res.DeleteFile = true
			return res
		}
		edits = nil
		edits = append(edits, removeChildren(root, memberSpans(root), rootDrop, rec.Interiors["root"])...)
		// Nothing else in "hooks" can have changed: every member was dropped.
	default:
		edits = append(edits, removeChildren(hooks, memberSpans(hooks), dropMember, rec.Interiors["hooks"])...)
		// removeChildren turns "all members dropped" into an empty object.
	}
	if len(edits) == 0 {
		return res
	}
	out, err := applyTextEdits(cur, edits)
	if err != nil || !json.Valid(out) {
		res.Removed, res.Skipped = 0, "the edit would not leave valid JSON"
		return res
	}
	res.Data = out
	return res
}
