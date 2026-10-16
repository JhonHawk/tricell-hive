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
			if op.CommandWithheld {
				class = "outside_grammar"
			}
			line, err := json.Marshal(flashPrivateOperation{SessionID: op.SessionID, Agent: op.Agent, Tool: op.Tool, Kind: op.Kind, ID: op.ID, Class: class, Path: op.Path, Text: op.PrivateText})
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

func loadFlashPrivateOperations(dir, name string) map[string]flashPrivateOperation {
	out := map[string]flashPrivateOperation{}
	if dir == "" || name == "" || filepath.Base(name) != name {
		return out
	}
	file, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return out
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
	return out
}

var (
	flashGitWrite       = regexp.MustCompile(`\bgit\s+(?:[^;&|\n]*\s)?(?:init|commit|push)\b`)
	flashShellSeparator = regexp.MustCompile(`&&|\|\||[|;\n]`)
	flashQuoted         = regexp.MustCompile(`'[^']*'|"[^"$` + "`" + `\\]*"`)
	flashStderrRedirect = regexp.MustCompile(`2>\s*/dev/null|2>&1`)
)

// flashReadOnlyShell accepts a composition of the grammar's read-only commands
// whose arguments may be quoted: quoting, pipes, && and 2>/dev/null widen the
// per-command grammar, but nothing else. Anything it does not recognize stays
// with the human inspection.
func flashReadOnlyShell(text string) bool {
	if len(text) > 8192 || strings.ContainsAny(text, "$`(){}\r<") {
		return false
	}
	text = flashStderrRedirect.ReplaceAllString(flashQuoted.ReplaceAllString(text, "Q"), " ")
	if strings.ContainsAny(text, `'"`) {
		return false
	}
	segments := 0
	for _, segment := range flashShellSeparator.Split(text, -1) {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		if flashOperationCommand(segment) != segment {
			return false
		}
		segments++
	}
	return segments > 0
}

func flashSearchShell(text string) bool {
	for _, segment := range flashShellSeparator.Split(text, -1) {
		words := strings.Fields(segment)
		if len(words) > 0 && (words[0] == "rg" || words[0] == "grep" || words[0] == "find" || (words[0] == "git" && len(words) > 1 && words[1] == "grep")) {
			return true
		}
	}
	return false
}

func flashHasParentSegment(value string) bool {
	for _, part := range strings.FieldsFunc(value, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return true
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
		return p.Text, true
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

// reviewFlashOperations removes the operations whose text the inspection can
// accept and returns the remaining events and whether an accepted operation
// searched the fixture. A retained shell event carries its text so the path
// safety check can see a git write.
func reviewFlashOperations(events []traceEvent, reviewed map[string]flashPrivateOperation, roots []string) ([]traceEvent, bool) {
	kept := make([]traceEvent, 0, len(events))
	searched := false
	for _, e := range events {
		tool := strings.ToLower(e.Tool)
		if e.Kind != "shell" && tool != "glob" && tool != "grep" {
			kept = append(kept, e)
			continue
		}
		text, ok := flashEventText(e, reviewed)
		if !ok {
			kept = append(kept, e)
			continue
		}
		if e.Kind == "shell" {
			e.Command = text
			if !flashGitWrite.MatchString(text) && flashReadOnlyShell(text) {
				searched = searched || flashSearchShell(text)
				continue
			}
		} else if filepath.IsAbs(text) || flashHasParentSegment(text) || flashHasParentSegment(e.Path) || !flashSearchRootAllowed(e.Path, roots) {
			// A search that may leave the fixture is opaque, whatever its kind.
			e.Kind = "tool"
		} else {
			searched = true
			continue
		}
		kept = append(kept, e)
	}
	return kept, searched
}

func flashSearchRootAllowed(path string, roots []string) bool {
	if path == "" || !filepath.IsAbs(path) {
		return true
	}
	for _, root := range roots {
		if root != "" && under(filepath.Clean(path), root) {
			return true
		}
	}
	return false
}

// flashContainmentEvidence says why a search cannot be shown to have stayed
// inside regular files. Only a typed inventory can prove it: an entry without a
// kind predates the field, and a symlink may lead a recursive search elsewhere.
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
		out = append(out, fmt.Sprintf("A search ran over a fixture holding %d symlink entries (%s); containment is not proven.", len(links), strings.Join(links, ", ")))
	}
	if len(untyped) > 0 {
		out = append(out, fmt.Sprintf("A search ran over a fixture inventory without entry kinds (%d entries); containment is not proven.", len(untyped)))
	}
	return out
}
