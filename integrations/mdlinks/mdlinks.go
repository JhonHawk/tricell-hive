// Package mdlinks holds the Markdown link logic shared by release validation
// and the installer. It has no dependencies on the rest of the repository so
// that both tooling/management and integrations/agents can import it.
package mdlinks

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// fenceScan walks data line by line and reports, for each line, whether it is
// ordinary text, fenced code, or a fence delimiter. It follows CommonMark's
// rule that a fence closes only on the same character, at least as long as the
// opener, with no info string.
func fenceScan(data string, visit func(line string, kind lineKind)) {
	var fence byte
	size := 0
	for _, line := range strings.Split(data, "\n") {
		trim := strings.TrimLeft(line, " ")
		if len(line)-len(trim) <= 3 && len(trim) >= 3 && (trim[0] == '`' || trim[0] == '~') {
			n := 0
			for n < len(trim) && trim[n] == trim[0] {
				n++
			}
			if n >= 3 {
				if fence == 0 {
					fence, size = trim[0], n
					visit(line, delimiterLine)
					continue
				}
				if trim[0] == fence && n >= size && strings.TrimSpace(trim[n:]) == "" {
					fence, size = 0, 0
					visit(line, delimiterLine)
					continue
				}
			}
		}
		if fence == 0 {
			visit(line, textLine)
		} else {
			visit(line, fencedLine)
		}
	}
}

type lineKind int

const (
	textLine lineKind = iota
	fencedLine
	delimiterLine
)

// OutsideFences returns data without fenced code blocks and their delimiter
// lines, each remaining line ending in a newline. Validation uses it so a
// template's example links are not treated as release dependencies.
func OutsideFences(data string) string {
	var out strings.Builder
	fenceScan(data, func(line string, kind lineKind) {
		if kind == textLine {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	})
	return out.String()
}

// skillLink matches an inline link whose destination is a skill: locator, with
// an optional angle-bracket form and title. Reference-style definitions
// ("[x]: skill:...") are deliberately not matched.
var skillLink = regexp.MustCompile(`(\[[^\]\n]*\]\(\s*)(?:<skill:([^>\n]+)>|skill:([^\s)]+))((?:\s+"[^"\n]*")?\s*\))`)

// RewriteSkillLinks replaces the destination of every inline
// [label](skill:owner/path) link outside fenced code with <dir/owner/path>,
// where dir is the skills directory the installer used. The angle brackets keep
// a path with spaces a single destination. A locator that cannot be unescaped
// is left as written.
func RewriteSkillLinks(body, dir string) string {
	var out strings.Builder
	first := true
	fenceScan(body, func(line string, kind lineKind) {
		if !first {
			out.WriteByte('\n')
		}
		first = false
		if kind != textLine || !strings.Contains(line, "skill:") {
			out.WriteString(line)
			return
		}
		out.WriteString(skillLink.ReplaceAllStringFunc(line, func(m string) string {
			sub := skillLink.FindStringSubmatch(m)
			locator := sub[2] + sub[3]
			resource, err := url.PathUnescape(locator)
			if err != nil {
				return m
			}
			return sub[1] + "<" + filepath.Join(dir, resource) + ">" + sub[4]
		}))
	})
	return out.String()
}
