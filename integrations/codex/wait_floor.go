package codex

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// The wait floor makes Codex's multi-agent v2 `wait_agent` raise any wait
// shorter than WaitFloorMinMS to that minimum, so a parent does not wake every
// 60 seconds. The editor below is a pure function over the bytes of
// config.toml: the caller owns file I/O, fingerprints and the journal.

// byteOrderMark is the UTF-8 BOM some editors put at the start of a file.
const byteOrderMark = "\xef\xbb\xbf"

const (
	// WaitFloorMinMS is the value written for min_wait_timeout_ms.
	WaitFloorMinMS = 300000
	// WaitFloorDefaultMS is the value written for default_wait_timeout_ms; it
	// is also the lowest max_wait_timeout_ms that does not conflict.
	WaitFloorDefaultMS = 600000

	waitFloorTable      = "features.multi_agent_v2"
	waitFloorMinKey     = "min_wait_timeout_ms"
	waitFloorDefaultKey = "default_wait_timeout_ms"
	waitFloorMaxKey     = "max_wait_timeout_ms"
)

// WaitFloorState classifies a config.toml against the wait floor.
type WaitFloorState int

const (
	// WaitFloorAbsent: no [features.multi_agent_v2] table; the edit creates it.
	WaitFloorAbsent WaitFloorState = iota
	// WaitFloorApplied: the table already holds both values; no edit.
	WaitFloorApplied
	// WaitFloorMissing: the table exists without one or both keys; the edit
	// inserts only the missing keys.
	WaitFloorMissing
	// WaitFloorConflict: the user set a different value; Hive must not write.
	WaitFloorConflict
	// WaitFloorUnsupported: a TOML form this line scanner cannot read with
	// certainty; Hive must not write.
	WaitFloorUnsupported
)

// WaitFloorRefusal explains why the editor will not write. Cause is a fixed
// phrase and never holds text from the file, which may contain secrets.
type WaitFloorRefusal struct {
	State WaitFloorState
	Cause string
}

func (r *WaitFloorRefusal) Error() string {
	return "Codex config cannot be edited safely: " + r.Cause
}

// WaitFloorRecord is what Hive wrote, as the caller must keep it to undo the
// edit: the inserted lines (header first when the table was created, without
// line endings) and what Hive created.
type WaitFloorRecord struct {
	Lines        []string
	CreatedTable bool
	CreatedFile  bool
	// Separator is true when Hive inserted a blank line before the new header.
	Separator bool
}

// WaitFloorEdit is an insertion: Insert goes at byte Offset of the content the
// edit was computed from.
type WaitFloorEdit struct {
	Offset int
	Insert []byte
	Record WaitFloorRecord
}

// Apply returns content with the insertion applied. It does not modify content.
func (e WaitFloorEdit) Apply(content []byte) []byte {
	if len(e.Insert) == 0 {
		return append([]byte(nil), content...)
	}
	out := make([]byte, 0, len(content)+len(e.Insert))
	out = append(out, content[:e.Offset]...)
	out = append(out, e.Insert...)
	return append(out, content[e.Offset:]...)
}

type tomlLine struct {
	start, end int    // byte range including the line ending
	text       string // without line ending or leading BOM
	hasEOL     bool
}

func splitTOMLLines(content []byte) []tomlLine {
	var lines []tomlLine
	for pos := 0; pos < len(content); {
		nl := bytes.IndexByte(content[pos:], '\n')
		end, textEnd, hasEOL := len(content), len(content), false
		if nl >= 0 {
			end, textEnd, hasEOL = pos+nl+1, pos+nl, true
		}
		if textEnd > pos && content[textEnd-1] == '\r' && hasEOL {
			textEnd--
		}
		text := string(content[pos:textEnd])
		if pos == 0 {
			text = strings.TrimPrefix(text, byteOrderMark)
		}
		lines = append(lines, tomlLine{start: pos, end: end, text: text, hasEOL: hasEOL})
		pos = end
	}
	return lines
}

// lineEnding picks the ending to insert: CRLF when the first line break of the
// file is CRLF, LF otherwise.
func lineEnding(content []byte) string {
	if i := bytes.IndexByte(content, '\n'); i > 0 && content[i-1] == '\r' {
		return "\r\n"
	}
	return "\n"
}

// parseWaitFloorInt reads a TOML decimal integer, allowing `_` between digits.
func parseWaitFloorInt(raw string) (int, bool) {
	s := strings.TrimSpace(raw)
	if s == "" || strings.HasPrefix(s, "_") || strings.HasSuffix(s, "_") || strings.Contains(s, "__") {
		return 0, false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && r != '_' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(strings.ReplaceAll(s, "_", ""))
	if err != nil {
		return 0, false
	}
	return n, true
}

type waitFloorScan struct {
	tableIdx int            // index of the header line, -1 when absent
	anchor   int            // last line of the table with content (header or key)
	bodyEnd  int            // index after the last line of the table
	values   map[string]int // integer values of the wait floor keys
	keyLine  map[string]int // line index of each wait floor key
	lines    []tomlLine
}

// splitTOMLKey splits a (possibly dotted) key or table name into its segments,
// honoring quotes, so `a."b.c"` is two segments and `"a.b"` is one. ok is
// false for a malformed key.
func splitTOMLKey(name string) (segs []string, ok bool) {
	var cur strings.Builder
	var quote rune
	flush := func() {
		segs = append(segs, strings.TrimSpace(cur.String()))
		cur.Reset()
	}
	for _, r := range name {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '.':
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	if quote != 0 {
		return nil, false
	}
	flush()
	for _, s := range segs {
		if s == "" {
			return nil, false
		}
	}
	return segs, true
}

func refuse(state WaitFloorState, cause string) *WaitFloorRefusal {
	return &WaitFloorRefusal{State: state, Cause: cause}
}

// scanWaitFloor reads the file for the wait floor table. It returns a refusal
// for conflicts and for forms it cannot read with certainty. forUndo skips the
// conflict and value checks, because the inverse must read a file the user has
// edited since Hive wrote it.
func scanWaitFloor(content []byte, forUndo bool) (*waitFloorScan, *WaitFloorRefusal) {
	if len(content) > maxConfigLine {
		return nil, refuse(WaitFloorUnsupported, "the file is larger than 1 MiB")
	}
	scan := &waitFloorScan{tableIdx: -1, anchor: -1, values: map[string]int{}, keyLine: map[string]int{}, lines: splitTOMLLines(content)}
	scan.bodyEnd = len(scan.lines)
	const (
		secOther = iota
		secFeatures
		secTarget
	)
	section := secOther
	atRoot := true
	for i, ln := range scan.lines {
		code := stripTOMLComment(ln.text)
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}
		if strings.Contains(code, `"""`) || strings.Contains(code, "'''") {
			return nil, refuse(WaitFloorUnsupported, "a multi-line string is present")
		}
		if strings.HasPrefix(trimmed, "[") {
			atRoot = false
			if section == secTarget {
				scan.bodyEnd = i
			}
			section = secOther
			array := strings.HasPrefix(trimmed, "[[")
			if !strings.HasSuffix(trimmed, "]") {
				if strings.Contains(trimmed, "features") {
					return nil, refuse(WaitFloorUnsupported, "a features table header is malformed")
				}
				continue
			}
			inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]")
			if array {
				inner = strings.TrimSuffix(strings.TrimPrefix(inner, "["), "]")
			}
			segs, ok := splitTOMLKey(inner)
			if !ok {
				if strings.Contains(trimmed, "features") {
					return nil, refuse(WaitFloorUnsupported, "a features table header is malformed")
				}
				continue
			}
			if segs[0] != "features" {
				continue
			}
			switch {
			case array:
				return nil, refuse(WaitFloorUnsupported, "features is an array of tables")
			case len(segs) == 1:
				section = secFeatures
			case segs[1] != "multi_agent_v2":
			case len(segs) > 2:
				return nil, refuse(WaitFloorUnsupported, "a sub-table of "+waitFloorTable+" is present")
			case scan.tableIdx >= 0:
				return nil, refuse(WaitFloorUnsupported, waitFloorTable+" appears more than once")
			default:
				scan.tableIdx, scan.anchor, section = i, i, secTarget
			}
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if section == secTarget {
			scan.anchor = i
		}
		if len(parts) != 2 {
			if section == secTarget {
				return nil, refuse(WaitFloorUnsupported, "a line of "+waitFloorTable+" is not a key and value")
			}
			continue
		}
		segs, ok := splitTOMLKey(parts[0])
		if !ok {
			if section == secTarget {
				return nil, refuse(WaitFloorUnsupported, "a key of "+waitFloorTable+" is malformed")
			}
			continue
		}
		switch {
		case atRoot && segs[0] == "features" && (len(segs) == 1 || segs[1] == "multi_agent_v2"):
			return nil, refuse(WaitFloorUnsupported, "features or "+waitFloorTable+" is set with a dotted or inline form")
		case section == secFeatures && segs[0] == "multi_agent_v2":
			return nil, refuse(WaitFloorUnsupported, "multi_agent_v2 is set inside [features]")
		case section == secTarget && len(segs) == 1 && (segs[0] == waitFloorMinKey || segs[0] == waitFloorDefaultKey || segs[0] == waitFloorMaxKey):
			if _, dup := scan.keyLine[segs[0]]; dup {
				return nil, refuse(WaitFloorUnsupported, "a key of "+waitFloorTable+" is repeated")
			}
			scan.keyLine[segs[0]] = i
			n, ok := parseWaitFloorInt(parts[1])
			if !ok && forUndo {
				continue
			}
			if !ok {
				return nil, refuse(WaitFloorUnsupported, "a value of "+waitFloorTable+" is not an integer")
			}
			scan.values[segs[0]] = n
		}
	}
	if forUndo {
		return scan, nil
	}
	if v, ok := scan.values[waitFloorMinKey]; ok && v != WaitFloorMinMS {
		return nil, refuse(WaitFloorConflict, waitFloorMinKey+" has a different value")
	}
	if v, ok := scan.values[waitFloorDefaultKey]; ok && v != WaitFloorDefaultMS {
		return nil, refuse(WaitFloorConflict, waitFloorDefaultKey+" has a different value")
	}
	if v, ok := scan.values[waitFloorMaxKey]; ok && v < WaitFloorDefaultMS {
		return nil, refuse(WaitFloorConflict, waitFloorMaxKey+" is below the default wait")
	}
	return scan, nil
}

// WaitFloor classifies content, the bytes of Codex's config.toml, and returns
// the insertion that sets the wait floor. exists is false when the file does
// not exist; the edit then also creates it. The function never touches the
// file system and never puts text from content in an error.
func WaitFloor(content []byte, exists bool) (WaitFloorState, WaitFloorEdit, error) {
	scan, refusal := scanWaitFloor(content, false)
	if refusal != nil {
		return refusal.State, WaitFloorEdit{}, refusal
	}
	eol := lineEnding(content)
	var want []string
	if _, ok := scan.keyLine[waitFloorMinKey]; !ok {
		want = append(want, fmt.Sprintf("%s = %d", waitFloorMinKey, WaitFloorMinMS))
	}
	if _, ok := scan.keyLine[waitFloorDefaultKey]; !ok {
		want = append(want, fmt.Sprintf("%s = %d", waitFloorDefaultKey, WaitFloorDefaultMS))
	}

	if scan.tableIdx < 0 {
		rec := WaitFloorRecord{Lines: append([]string{"[" + waitFloorTable + "]"}, want...), CreatedTable: true, CreatedFile: !exists}
		var ins strings.Builder
		body := strings.TrimPrefix(string(content), byteOrderMark)
		if strings.TrimSpace(body) != "" {
			if content[len(content)-1] != '\n' {
				ins.WriteString(eol)
			}
			if last := scan.lines[len(scan.lines)-1]; strings.TrimSpace(last.text) != "" || !last.hasEOL {
				ins.WriteString(eol)
				rec.Separator = true
			}
		}
		for _, l := range rec.Lines {
			ins.WriteString(l + eol)
		}
		return WaitFloorAbsent, WaitFloorEdit{Offset: len(content), Insert: []byte(ins.String()), Record: rec}, nil
	}
	if len(want) == 0 {
		return WaitFloorApplied, WaitFloorEdit{}, nil
	}
	anchor := scan.lines[scan.anchor]
	var ins strings.Builder
	if !anchor.hasEOL {
		ins.WriteString(eol)
	}
	for _, l := range want {
		ins.WriteString(l + eol)
	}
	return WaitFloorMissing, WaitFloorEdit{Offset: anchor.end, Insert: []byte(ins.String()), Record: WaitFloorRecord{Lines: want}}, nil
}

// WaitFloorUndo is the result of removing what Hive wrote. Removed and Kept
// name registered keys, never file text.
type WaitFloorUndo struct {
	// Content is the file after the removal. When DeleteFile is true it is empty.
	Content []byte
	// Removed lists the registered keys whose line was found unchanged and removed.
	Removed []string
	// Kept lists the registered keys whose line is gone or holds another value.
	Kept []string
	// HeaderKept is true when Hive created the table but it still holds other content.
	HeaderKept bool
	// DeleteFile is true when Hive created the file and nothing is left in it.
	DeleteFile bool
	// Skipped is a fixed phrase when the table could not be found or read; the
	// content is then unchanged and every registered key is in Kept.
	Skipped string
}

// UndoWaitFloor removes, from the current content, the lines rec registers,
// and only those that still hold the value Hive wrote. It removes the header
// only when Hive created the table and nothing else remains in it, the blank
// separator only with that header, and signals deletion of the file only when
// Hive created it and it is empty. It never fails: a file Codex or the user has
// changed is reported in the result, so a recovery is never blocked by it.
func UndoWaitFloor(content []byte, rec WaitFloorRecord) WaitFloorUndo {
	res := WaitFloorUndo{Content: content}
	type regLine struct {
		key string
		val int
	}
	var reg []regLine
	for _, l := range rec.Lines {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 || strings.HasPrefix(strings.TrimSpace(l), "[") {
			continue
		}
		if v, ok := parseWaitFloorInt(parts[1]); ok {
			reg = append(reg, regLine{strings.TrimSpace(parts[0]), v})
		}
	}
	keepAll := func(cause string) WaitFloorUndo {
		res.Skipped = cause
		for _, r := range reg {
			res.Kept = append(res.Kept, r.key)
		}
		return res
	}
	scan, refusal := scanWaitFloor(content, true)
	if refusal != nil {
		return keepAll("the file can no longer be read with certainty")
	}
	if scan.tableIdx < 0 {
		return keepAll(waitFloorTable + " is no longer present")
	}
	drop := map[int]bool{}
	for _, r := range reg {
		if idx, ok := scan.keyLine[r.key]; ok {
			if v, isInt := scan.values[r.key]; isInt && v == r.val {
				drop[idx] = true
				res.Removed = append(res.Removed, r.key)
				continue
			}
		}
		res.Kept = append(res.Kept, r.key)
	}
	if rec.CreatedTable {
		remaining := false
		for i := scan.tableIdx + 1; i < scan.bodyEnd; i++ {
			if !drop[i] && strings.TrimSpace(scan.lines[i].text) != "" {
				remaining = true
			}
		}
		if remaining {
			res.HeaderKept = true
		} else {
			drop[scan.tableIdx] = true
			if i := scan.tableIdx - 1; rec.Separator && i >= 0 && strings.TrimSpace(scan.lines[i].text) == "" {
				drop[i] = true
			}
		}
	}
	if len(drop) == 0 {
		return res
	}
	out := make([]byte, 0, len(content))
	for i, ln := range scan.lines {
		if !drop[i] {
			out = append(out, content[ln.start:ln.end]...)
		}
	}
	res.Content = out
	if rec.CreatedFile && strings.TrimSpace(strings.TrimPrefix(string(out), byteOrderMark)) == "" {
		res.DeleteFile = true
		res.Content = nil
	}
	return res
}
