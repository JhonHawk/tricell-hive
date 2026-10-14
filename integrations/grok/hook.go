package grok

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ReadHookName is the file Hive owns in <GrokHome>/hooks.
const ReadHookName = "hive-read-whole-references.json"

// ReadHookPath returns where the read-whole hook lives for a Grok home.
func ReadHookPath(grokHome string) string {
	return filepath.Join(grokHome, "hooks", ReadHookName)
}

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

type hookGroup struct {
	Matcher string        `json:"matcher"`
	Hooks   []hookCommand `json:"hooks"`
}

type hookFile struct {
	Hooks struct {
		PreToolUse []hookGroup `json:"PreToolUse"`
	} `json:"hooks"`
}

// RenderReadHook returns the hook file that makes Grok read Hive's reference
// files whole. Grok runs the command with sh -c, so both paths are single
// quoted; they come from a plan that may have been edited by hand, so they are
// checked here as well.
func RenderReadHook(executable, skillsDir string) ([]byte, error) {
	for name, p := range map[string]string{"executable": executable, "skills directory": skillsDir} {
		if !filepath.IsAbs(p) || strings.ContainsRune(p, 0) {
			return nil, fmt.Errorf("Grok read-whole hook %s must be an absolute path without NUL", name)
		}
	}
	var f hookFile
	f.Hooks.PreToolUse = []hookGroup{{
		Matcher: "read_file",
		Hooks: []hookCommand{{
			Type:    "command",
			Command: shellQuote(executable) + " grok-hook read-whole --skills " + shellQuote(skillsDir),
			Timeout: 10,
		}},
	}}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// shellQuote quotes s as one POSIX shell word: single quotes around it, and
// each embedded single quote closed, escaped and reopened.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
