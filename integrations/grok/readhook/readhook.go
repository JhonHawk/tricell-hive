// Package readhook implements the logic of the Grok PreToolUse hook that
// makes reads of Hive reference files whole. It does no process I/O.
//
// The hook input is model-controlled, so the path is validated (inside the
// skills folder after resolving symlinks, a regular .md file, bounded size)
// before the file content is read, and nothing from it reaches a shell.
package readhook

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxLines = 1000
	maxBytes = 64 * 1024
)

// ReadWhole returns the PreToolUse reply that removes offset and limit from a
// read_file input, or nil when the call must go through unchanged.
func ReadWhole(event []byte, skillsDir string) []byte {
	var ev struct {
		ToolInput          map[string]json.RawMessage `json:"toolInput"`
		ToolInputTruncated bool                       `json:"toolInputTruncated"`
		Cwd                string                     `json:"cwd"`
		WorkspaceRoot      string                     `json:"workspaceRoot"`
	}
	if json.Unmarshal(event, &ev) != nil || ev.ToolInput == nil || ev.ToolInputTruncated {
		return nil
	}
	_, hasOffset := ev.ToolInput["offset"]
	_, hasLimit := ev.ToolInput["limit"]
	if !hasOffset && !hasLimit {
		return nil
	}
	target := stringField(ev.ToolInput, "target_file")
	if target == "" {
		target = stringField(ev.ToolInput, "path")
	}
	if target == "" || !eligible(target, ev.WorkspaceRoot, ev.Cwd, skillsDir) {
		return nil
	}
	updated := make(map[string]json.RawMessage, len(ev.ToolInput))
	for k, v := range ev.ToolInput {
		if k != "offset" && k != "limit" {
			updated[k] = v
		}
	}
	out, err := json.Marshal(map[string]any{"hookSpecificOutput": map[string]any{
		"hookEventName": "PreToolUse",
		"updatedInput":  updated,
	}})
	if err != nil {
		return nil
	}
	return out
}

func stringField(m map[string]json.RawMessage, key string) string {
	var s string
	if raw, ok := m[key]; !ok || json.Unmarshal(raw, &s) != nil {
		return ""
	}
	return s
}

func eligible(target, workspaceRoot, cwd, skillsDir string) bool {
	if !filepath.IsAbs(skillsDir) {
		return false
	}
	if !filepath.IsAbs(target) {
		base := workspaceRoot
		if base == "" {
			base = cwd
		}
		if !filepath.IsAbs(base) {
			return false
		}
		target = filepath.Join(base, target)
	}
	root, err := filepath.EvalSymlinks(filepath.Clean(skillsDir))
	if err != nil {
		return false
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(target))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	if !strings.HasSuffix(resolved, ".md") {
		return false
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return false
	}
	f, err := os.Open(resolved)
	if err != nil {
		return false
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil || len(data) > maxBytes {
		return false
	}
	lines := bytes.Count(data, []byte{'\n'})
	if len(data) > 0 && data[len(data)-1] != '\n' {
		lines++
	}
	return lines <= maxLines
}
