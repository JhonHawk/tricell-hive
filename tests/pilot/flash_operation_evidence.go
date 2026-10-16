package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// The private operations file keeps the full text of every shell/execute call
// that the bounded grammar withheld and every glob/grep pattern, so a later
// inspection can decide from the text. It sits beside run.json with mode 0600;
// run.json carries only its name, and no publication step may copy it.
const flashPrivateOperationsFile = "operations-private.jsonl"

type flashPrivateOperation struct {
	SessionID string `json:"session_id"`
	Agent     string `json:"agent"`
	Tool      string `json:"tool"`
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	Class     string `json:"class"`
	Path      string `json:"path,omitempty"`
	Text      string `json:"text"`
	Truncated bool   `json:"truncated,omitempty"`
}

func flashPrivateKey(sessionID, id string) string { return sessionID + "\x00" + id }

// writeFlashPrivateOperations always creates the file, even without entries:
// its presence tells the inspection that withheld text is reviewable.
func writeFlashPrivateOperations(root *os.Root, children ...flashChild) (string, error) {
	var buf bytes.Buffer
	for _, c := range children {
		for _, op := range c.Operations {
			if op.PrivateText == "" {
				continue
			}
			class := "pattern"
			if strings.EqualFold(op.Tool, "execute") {
				class = "execute"
			} else if op.CommandWithheld {
				class = "outside_grammar"
			}
			text, truncated := capFlashPrivateText(op.PrivateText)
			path, _ := capFlashPrivateText(op.Path)
			line, err := json.Marshal(flashPrivateOperation{SessionID: op.SessionID, Agent: op.Agent, Tool: op.Tool, Kind: op.Kind, ID: op.ID, Class: class, Path: path, Text: text, Truncated: truncated})
			if err != nil {
				return "", fmt.Errorf("private operation encoding failed")
			}
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}
	file, err := root.OpenFile(flashPrivateOperationsFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("private operations destination unavailable")
	}
	_, writeErr := file.Write(buf.Bytes())
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return "", fmt.Errorf("private operations write failed")
	}
	return flashPrivateOperationsFile, nil
}

// The reader stops at a line over 1 MiB, so the writer bounds every stored
// field; a truncated text is marked and never auto-accepted.
const flashPrivateTextLimit = 64 << 10

func capFlashPrivateText(text string) (string, bool) {
	if len(text) <= flashPrivateTextLimit {
		return text, false
	}
	return strings.ToValidUTF8(text[:flashPrivateTextLimit], ""), true
}

func loadFlashPrivateOperations(dir, name string) (map[string]flashPrivateOperation, error) {
	out := map[string]flashPrivateOperation{}
	if dir == "" || name == "" || filepath.Base(name) != name {
		return out, fmt.Errorf("private operations file not locatable")
	}
	file, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return out, fmt.Errorf("private operations file unreadable")
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	scan.Buffer(make([]byte, 64*1024), 1024*1024)
	for scan.Scan() {
		var op flashPrivateOperation
		if json.Unmarshal(scan.Bytes(), &op) == nil && op.ID != "" {
			out[flashPrivateKey(op.SessionID, op.ID)] = op
		}
	}
	if scan.Err() != nil {
		return out, fmt.Errorf("private operations file truncated by an oversized line")
	}
	return out, nil
}

var (
	flashGitWrite       = regexp.MustCompile(`\bgit\s+(?:[^;&|\n]*\s)?(?:init|commit|push)\b`)
	flashShellSeparator = regexp.MustCompile(`&&|\|\||[|;\n]`)
	flashQuoted         = regexp.MustCompile(`'[^']*'|"[^"$` + "`" + `\\]*"`)
	// Whole tokens only: a redirect glued to more characters (2>&1x) is a different one.
	flashStderrRedirect = regexp.MustCompile(`(^|[ \t\n|;&])(?:2>[ \t]*/dev/null|2>&1)($|[ \t\n|;&])`)
)

// flashStripQuotes replaces each quoted token with the bare word Q and returns
// the quoted contents. The shell glues a quoted span to adjacent characters
// (""/etc is /etc, .”./etc is ../etc), so a span that touches anything but
// whitespace or a separator keeps its quotes, as does a span that starts with
// "-"; the caller rejects both.
func flashStripQuotes(text string) (string, []string) {
	var contents []string
	var out strings.Builder
	last := 0
	for _, span := range flashQuoted.FindAllStringIndex(text, -1) {
		start, end := span[0], span[1]
		inner := text[start+1 : end-1]
		contents = append(contents, inner)
		out.WriteString(text[last:start])
		last = end
		if !flashQuoteBoundary(text, start-1) || !flashQuoteBoundary(text, end) || strings.HasPrefix(strings.TrimLeft(inner, " \t"), "-") {
			out.WriteString(text[start:end])
			continue
		}
		out.WriteString("Q")
	}
	out.WriteString(text[last:])
	return out.String(), contents
}

// flashQuoteBoundary reports whether the byte at i ends a word: the text edge,
// whitespace or a command separator.
func flashQuoteBoundary(text string, i int) bool {
	return i < 0 || i >= len(text) || strings.IndexByte(" \t\n|;&", text[i]) >= 0
}

// flashReadOnlyShell accepts a composition of the grammar's read-only commands
// whose arguments may be quoted, and whose paths stay inside the roots: quoting,
// pipes, && and 2>/dev/null widen the per-command grammar, but nothing else.
// Anything it does not recognize stays with the human inspection.
func flashReadOnlyShell(text, cwd string, roots []string) bool {
	if len(text) > 8192 || strings.ContainsAny(text, "$`(){}\r<") {
		return false
	}
	stripped, quoted := flashStripQuotes(text)
	// Adjacent redirects share a boundary character, so repeat until stable.
	for previous := ""; previous != stripped; {
		previous = stripped
		stripped = flashStderrRedirect.ReplaceAllString(stripped, "$1 $2")
	}
	if strings.ContainsAny(stripped, `'"`) {
		return false
	}
	segments := 0
	for _, segment := range flashShellSeparator.Split(stripped, -1) {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		if flashOperationCommand(segment) != segment {
			return false
		}
		for _, word := range strings.Fields(segment)[1:] {
			if !flashContained(word, cwd, roots) {
				return false
			}
		}
		segments++
	}
	for _, inner := range quoted {
		if !flashContained(inner, cwd, roots) {
			return false
		}
	}
	return segments > 0
}

// flashContained judges a path-like argument with the same mechanism the write
// checks use (eventPath, under): home references, absolute paths and ".." climbs
// must resolve inside a root. Plain words resolve under the working directory.
func flashContained(argument, cwd string, roots []string) bool {
	if argument == "" || strings.HasPrefix(argument, "-") {
		return true
	}
	if strings.HasPrefix(argument, "~") {
		return false
	}
	if !filepath.IsAbs(argument) && !strings.Contains(argument, "..") {
		return true
	}
	absolute := eventPath(argument, cwd)
	for _, root := range roots {
		if root != "" && under(absolute, filepath.Clean(root)) {
			return true
		}
	}
	return false
}

// flashGlobContained also rejects brace and bracket expansion: one pattern can
// then name several paths, and only the first is resolved here.
func flashGlobContained(pattern, cwd string, roots []string) bool {
	return !strings.ContainsAny(pattern, "{[") && flashContained(pattern, cwd, roots)
}

// flashGitWriteText reports a git init, commit or push that can run: the git
// subcommand itself, after any global options, not a word elsewhere in a
// read-only command such as "git log --grep commit".
func flashGitWriteText(text string) bool {
	for _, segment := range flashShellSeparator.Split(text, -1) {
		words := strings.Fields(segment)
		if len(words) < 2 || words[0] != "git" {
			continue
		}
		for i := 1; i < len(words); i++ {
			switch word := words[i]; {
			case word == "-C" || word == "-c" || word == "--git-dir" || word == "--work-tree" || word == "--namespace":
				i++
			case strings.HasPrefix(word, "-"):
			default:
				if word == "init" || word == "commit" || word == "push" {
					return true
				}
				i = len(words)
			}
		}
	}
	return false
}

// flashEventText returns the text a gate decision rests on: the private
// evidence for a native-export operation, otherwise what the event itself holds.
func flashEventText(e traceEvent, reviewed map[string]flashPrivateOperation) (string, bool) {
	tool := strings.ToLower(e.Tool)
	if tool == "execute" {
		return "", false
	}
	if p, ok := reviewed[flashPrivateKey(e.SessionID, e.ID)]; ok && e.ID != "" {
		return p.Text, !p.Truncated
	}
	if e.Kind == "shell" && e.Command != "" {
		return e.Command, true
	}
	if tool == "glob" || tool == "grep" {
		var input struct {
			Pattern string `json:"pattern"`
		}
		if json.Unmarshal(e.Input, &input) == nil && input.Pattern != "" {
			return input.Pattern, true
		}
	}
	return "", false
}

// opaqueSearch makes a glob/grep event opaque: observedPathSafety does not
// treat the "search" kind of grep as opaque, so an unreviewed or escaping
// search would pass silently.
func opaqueSearch(e traceEvent) traceEvent {
	if tool := strings.ToLower(e.Tool); tool == "glob" || tool == "grep" {
		e.Kind = "tool"
	}
	return e
}

func markSearchesOpaque(events []traceEvent) []traceEvent {
	out := make([]traceEvent, len(events))
	for i, e := range events {
		out[i] = opaqueSearch(e)
	}
	return out
}

// reviewFlashOperations removes the operations whose text the inspection can
// accept and returns the remaining events and whether an operation was
// accepted. Only a git write keeps its text on a retained shell event, so the
// path safety check fails it; other unaccepted text stays opaque.
func reviewFlashOperations(events []traceEvent, reviewed map[string]flashPrivateOperation, cwd string, roots []string) ([]traceEvent, bool) {
	kept := make([]traceEvent, 0, len(events))
	accepted := false
	for _, e := range events {
		tool := strings.ToLower(e.Tool)
		if e.Kind != "shell" && tool != "glob" && tool != "grep" {
			kept = append(kept, e)
			continue
		}
		text, ok := flashEventText(e, reviewed)
		if !ok {
			kept = append(kept, opaqueSearch(e))
			continue
		}
		if e.Kind == "shell" {
			stripped, _ := flashStripQuotes(text)
			e.Command = ""
			if flashGitWriteText(stripped) {
				e.Command = stripped
			} else if flashReadOnlyShell(text, cwd, roots) {
				accepted = true
				continue
			}
		} else if (tool == "glob" && !flashGlobContained(text, cwd, roots)) || !flashGlobContained(e.Path, cwd, roots) {
			// A search that may leave the fixture is opaque, whatever its kind.
			e = opaqueSearch(e)
		} else {
			accepted = true
			continue
		}
		kept = append(kept, e)
	}
	return kept, accepted
}

// flashContainmentEvidence says why an accepted operation cannot be shown to
// have stayed inside regular files. Only a typed inventory can prove it: an
// entry without a kind predates the field, and a symlink may lead a read or
// search elsewhere.
func flashContainmentEvidence(after map[string]item) []string {
	var links, untyped []string
	for path, entry := range after {
		switch entry.Kind {
		case "file":
		case "symlink":
			links = append(links, path)
		default:
			untyped = append(untyped, path)
		}
	}
	sort.Strings(links)
	sort.Strings(untyped)
	var out []string
	if len(links) > 0 {
		out = append(out, fmt.Sprintf("An accepted operation ran over a fixture holding %d symlink entries (%s); containment is not proven.", len(links), strings.Join(links, ", ")))
	}
	if len(untyped) > 0 {
		out = append(out, fmt.Sprintf("An accepted operation ran over a fixture inventory without entry kinds (%d entries); containment is not proven.", len(untyped)))
	}
	return out
}

// recordFlashError appends to the run error so an earlier, more specific one survives.
func recordFlashError(r *flashClassReport, message string) {
	if r.Error == "" {
		r.Error = message
	} else {
		r.Error += "; " + message
	}
}
