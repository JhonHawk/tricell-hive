package main

import (
	"regexp"
	"strings"
)

// selfFeedingLog fails when one shell command redirects a compound block, or
// the whole script with exec, into a file that a trap or a command of that
// block reads: the reader's own output lands in the file it reads and the file
// grows without end. A simple command with a redirection is not this: Bash
// undoes that redirection before a trap runs.
func selfFeedingLog(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "self_feeding_log", Status: "not_observed", Evidence: []string{}}
	seen := map[string]bool{}
	for _, e := range s.Trace.Events {
		if e.Kind != "shell" && e.Kind != "exec_item" {
			continue
		}
		text := e.Command
		if text == "" {
			text = e.CommandTemplate
		}
		if text == "" || !selfFeedingCommand(text) {
			continue
		}
		if ev := regressionEvidence(e); !seen[ev] {
			seen[ev] = true
			c.Status = "fail"
			c.Evidence = append(c.Evidence, ev)
		}
	}
	return c
}

var (
	heredocStart = regexp.MustCompile(`<<-?\s*(?:'([A-Za-z_][A-Za-z0-9_]*)'|"([A-Za-z_][A-Za-z0-9_]*)"|([A-Za-z_][A-Za-z0-9_]*))`)
	shellWord    = regexp.MustCompile(`(?:^|[\s/])(?:ba|z|da)?sh\b`)
	// A reader of a file; the command must stand where a command can start.
	readerCmd = regexp.MustCompile(`(?:^|[\n;&|(){!]|\b(?:then|do|else)\b)\s*(?:sudo\s+)?(?:tail|head|cat|grep|egrep|fgrep|rg|sed|less)\b`)
	trapCmd   = regexp.MustCompile(`(?:^|[\n;&|(){]|\b(?:then|do|else)\b)\s*trap\s+`)
	execCmd   = regexp.MustCompile(`(?:^|[\n;&|(){]|\b(?:then|do|else)\b)\s*exec\s+`)
	// One redirection operator: group 1 is &> or &>>, group 2 the file
	// descriptor, group 3 > or >>, group 4 the & of a duplication such as 2>&1.
	redirOp    = regexp.MustCompile(`^[ \t]*(?:(&>>?)|([0-9]*)(>>?)(&)?)`)
	identOnly  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	restoreOut = regexp.MustCompile(`exec\s+1?>&[0-9-]`)
)

// stdoutRedirects walks every redirection word that starts at i and returns
// the files that receive standard output, as feedTarget words, and the index
// after the last redirection. A file descriptor other than 1 and a
// duplication such as 2>&1 send no standard output to a file.
func stdoutRedirects(text string, i int) (words, keys []string, end int) {
	end = i
	for end < len(text) {
		m := redirOp.FindStringSubmatchIndex(text[end:])
		if m == nil {
			break
		}
		pos := end + m[1]
		fd := ""
		if m[4] >= 0 {
			fd = text[end+m[4] : end+m[5]]
		}
		stdout := m[2] >= 0 || fd == "" || fd == "1"
		if m[2] < 0 && m[8] >= 0 {
			for pos < len(text) && !strings.ContainsRune(" \t\n;&|<>()", rune(text[pos])) {
				pos++
			}
			end = pos
			continue
		}
		for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t') {
			pos++
		}
		word, key, next := feedTarget(text, pos)
		if next == pos {
			break
		}
		if stdout && word != "" {
			words, keys = append(words, word), append(keys, key)
		}
		end = next
	}
	return words, keys, end
}

// selfFeedingCommand reads the text of one shell command, with no
// interpretation of variables: a file is its literal path or the name of the
// variable that holds it.
func selfFeedingCommand(text string) bool {
	text = inlineShellHeredocs(text)
	mask := shellMask(text)
	traps := trapBodies(text, mask)

	for _, m := range execCmd.FindAllStringIndex(text, -1) {
		if mask[kwStart(text, m[0])] {
			continue
		}
		words, keys, end := stdoutRedirects(text, m[1])
		rest := text[end:]
		if loc := restoreOut.FindStringIndex(rest); loc != nil {
			rest = rest[:loc[0]]
		}
		for k, word := range words {
			if readsFile(traps, keys[k], word) || readsFile([]string{rest}, keys[k], word) {
				return true
			}
		}
	}
	for i := 0; i < len(text); i++ {
		if mask[i] || (text[i] != '}' && text[i] != ')') {
			continue
		}
		words, keys, _ := stdoutRedirects(text, i+1)
		if len(words) == 0 {
			continue
		}
		open := matchingOpen(text, mask, i)
		if open < 0 {
			continue
		}
		for k, word := range words {
			if readsFile(traps, keys[k], word) || readsFile([]string{text[open+1 : i]}, keys[k], word) {
				return true
			}
		}
	}
	return false
}

// kwStart is the index of the keyword a command-position match begins with:
// the match may start with a separator and spaces.
func kwStart(text string, i int) int {
	for i < len(text) && strings.ContainsRune("\n;&|(){! \t", rune(text[i])) {
		i++
	}
	if i >= len(text) {
		return len(text) - 1
	}
	return i
}

// inlineShellHeredocs keeps the body of a heredoc fed to a shell, which runs
// as part of the command, and drops any other heredoc body, such as a Python
// script, whose text is data here.
func inlineShellHeredocs(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		m := heredocStart.FindStringSubmatchIndex(line)
		if m == nil {
			out = append(out, line)
			continue
		}
		delim := ""
		for g := 1; g <= 3; g++ {
			if m[2*g] >= 0 {
				delim = line[m[2*g]:m[2*g+1]]
			}
		}
		runs := shellWord.MatchString(line[:m[0]])
		if runs {
			out = append(out, line[:m[0]])
		} else {
			out = append(out, line)
		}
		for i++; i < len(lines) && strings.TrimSpace(lines[i]) != delim; i++ {
			if runs {
				out = append(out, lines[i])
			}
		}
	}
	return strings.Join(out, "\n")
}

type shellMasker struct {
	s    string
	mask []bool
}

// shellMask marks the bytes of quoted strings and comments, so that a brace or
// a keyword inside one is not read as shell structure. Command substitutions
// inside double quotes stay unmasked: they run.
func shellMask(s string) []bool {
	m := &shellMasker{s: s, mask: make([]bool, len(s))}
	m.code(0, false)
	return m.mask
}

func (m *shellMasker) fill(from, to int) {
	for k := from; k < to && k < len(m.mask); k++ {
		m.mask[k] = true
	}
}

func (m *shellMasker) code(i int, inSubst bool) int {
	depth := 0
	for i < len(m.s) {
		c := m.s[i]
		switch {
		case c == '\\':
			m.fill(i, i+2)
			i += 2
		case c == '\'':
			j := strings.IndexByte(m.s[i+1:], '\'')
			end := len(m.s)
			if j >= 0 {
				end = i + 1 + j + 1
			}
			m.fill(i, end)
			i = end
		case c == '"':
			i = m.double(i)
		case c == '#' && (i == 0 || strings.ContainsRune(" \t\n;&|(", rune(m.s[i-1]))):
			end := strings.IndexByte(m.s[i:], '\n')
			if end < 0 {
				end = len(m.s) - i
			}
			m.fill(i, i+end)
			i += end
		case inSubst && c == '(':
			depth++
			i++
		case inSubst && c == ')':
			if depth == 0 {
				return i + 1
			}
			depth--
			i++
		default:
			i++
		}
	}
	return i
}

func (m *shellMasker) double(i int) int {
	m.mask[i] = true
	i++
	for i < len(m.s) {
		switch {
		case m.s[i] == '\\':
			m.fill(i, i+2)
			i += 2
		case m.s[i] == '"':
			m.mask[i] = true
			return i + 1
		case m.s[i] == '$' && i+1 < len(m.s) && m.s[i+1] == '(':
			m.fill(i, i+2)
			i = m.code(i+2, true)
		default:
			m.mask[i] = true
			i++
		}
	}
	return i
}

// feedTarget reads the file a redirection names, starting at i: it
// returns the raw word, its key (the variable's name or the literal path) and
// the index after it. Output sinks are not files that can be read back.
func feedTarget(text string, i int) (word, key string, end int) {
	j := i
	switch {
	case j >= len(text):
		return "", "", i
	case text[j] == '"' || text[j] == '\'':
		k := strings.IndexByte(text[j+1:], text[j])
		if k < 0 {
			return "", "", i
		}
		end = j + 1 + k + 1
		word = text[j+1 : j+1+k]
	default:
		k := j
		for k < len(text) && !strings.ContainsRune(" \t\n;&|<>()", rune(text[k])) {
			k++
		}
		end, word = k, text[j:k]
	}
	if word == "" || strings.HasPrefix(word, "/dev/") || strings.HasPrefix(word, "&") {
		return "", "", end
	}
	key = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(word, "$"), "{"), "}")
	if strings.HasPrefix(word, "$") && identOnly.MatchString(key) {
		return word, key, end
	}
	return word, word, end
}

// matchingOpen finds the opening brace or parenthesis that the closer at i
// closes, ignoring masked bytes; -1 when it is not found or when the
// parenthesis starts a command substitution.
func matchingOpen(text string, mask []bool, i int) int {
	closer, opener := text[i], byte('{')
	if closer == ')' {
		opener = '('
	}
	depth := 0
	for k := i; k >= 0; k-- {
		if mask[k] {
			continue
		}
		switch text[k] {
		case closer:
			depth++
		case opener:
			depth--
			if depth == 0 {
				if opener == '(' && k > 0 && text[k-1] == '$' {
					return -1
				}
				return k
			}
		}
	}
	return -1
}

// trapBodies returns the command text of every trap in the script.
func trapBodies(text string, mask []bool) []string {
	var bodies []string
	for _, m := range trapCmd.FindAllStringIndex(text, -1) {
		if mask[kwStart(text, m[0])] {
			continue
		}
		i := m[1]
		if i >= len(text) {
			continue
		}
		switch text[i] {
		case '\'':
			if k := strings.IndexByte(text[i+1:], '\''); k >= 0 {
				bodies = append(bodies, text[i+1:i+1+k])
			}
		case '"':
			for k := i + 1; k < len(text); k++ {
				if text[k] == '\\' {
					k++
				} else if text[k] == '"' {
					bodies = append(bodies, text[i+1:k])
					break
				}
			}
		}
	}
	return bodies
}

// readsFile reports whether a reader command, standing where a command can
// start in any of the scripts, names the file.
func readsFile(scripts []string, key, word string) bool {
	for _, script := range scripts {
		mask := shellMask(script)
		for _, m := range readerCmd.FindAllStringIndex(script, -1) {
			start := kwStart(script, m[0])
			if mask[start] {
				continue
			}
			end := m[1]
			for end < len(script) {
				if !mask[end] && strings.ContainsRune("\n;&|)}", rune(script[end])) {
					break
				}
				end++
			}
			if namesFile(script[m[1]:end], key, word) {
				return true
			}
		}
	}
	return false
}

func namesFile(args, key, word string) bool {
	if identOnly.MatchString(key) && strings.HasPrefix(word, "$") {
		return regexp.MustCompile(`\$\{?` + key + `\b`).MatchString(args)
	}
	return strings.Contains(args, key)
}
