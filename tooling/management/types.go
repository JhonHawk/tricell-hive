// Package management owns deterministic file delivery. It never launches models.
package management

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"tricell-hive/integrations/target"
)

const Begin = "<!-- === TRICELL HIVE RULES:BEGIN === -->"
const End = "<!-- === TRICELL HIVE RULES:END === -->"
const GlobalSource = "content/guidance/global.md"
const SkillSource = "content/skills/workspace-conventions/SKILL.md"

type Payload struct {
	Path string `json:"path"`
	Data []byte `json:"data"`
	Mode uint32 `json:"mode,omitempty"`
}
type Release struct {
	ID       string    `json:"id"`
	Files    []Payload `json:"files"`
	Profiles []byte    `json:"profiles,omitempty"`
	Renderer string    `json:"renderer,omitempty"`
}
type Record struct {
	Target      target.Target `json:"target"`
	Managed     []byte        `json:"managed"`
	Mode        uint32        `json:"mode,omitempty"`
	Leading     string        `json:"leading,omitempty"`
	CreatedFile bool          `json:"created_file"`
	Release     string        `json:"release"`
	Consumers   []Consumer    `json:"consumers,omitempty"`
}
type Consumer struct {
	Host    string `json:"host"`
	Scope   string `json:"scope"`
	Context string `json:"context"`
}
type State struct {
	Version     int               `json:"version"`
	Records     map[string]Record `json:"records"`
	CreatedDirs []string          `json:"created_dirs,omitempty"`
}
type Fingerprint struct {
	Exists     bool   `json:"exists"`
	Hash       string `json:"hash"`
	Mode       uint32 `json:"mode"`
	Kind       string `json:"kind,omitempty"`
	LinkTarget string `json:"link_target,omitempty"`
	FileMode   uint32 `json:"file_mode,omitempty"`
}
type Change struct {
	Target   target.Target `json:"target"`
	Expected Fingerprint   `json:"expected"`
	Before   *Record       `json:"before,omitempty"`
	After    *Record       `json:"after,omitempty"`
	Replaces *Record       `json:"replaces,omitempty"`
}
type Plan struct {
	Version   int           `json:"version"`
	Action    string        `json:"action"`
	Config    target.Config `json:"config"`
	Hosts     []string      `json:"hosts"`
	StateDir  string        `json:"state_dir"`
	StateHash string        `json:"state_hash"`
	Release   *Release      `json:"release,omitempty"`
	Changes   []Change      `json:"changes"`
	ID        string        `json:"id"`
}
type Options struct {
	Scope, Home, Root, StateDir, Source, ReleaseID string
	Hosts                                          []string
}
type StatusEntry struct {
	Path, Host, Kind, Status, Release string
	Consumers                         []Consumer
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func encode(v any) []byte {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(b, '\n')
}

// V3 releases were addressed by their file list alone. Keep that identifier
// stable so existing release snapshots remain selectable after the upgrade.
func releaseID(r Release) string {
	if r.Renderer == "" && len(r.Profiles) == 0 {
		return hash(encode(r.Files))
	}
	return hash(encode(struct {
		Files    []Payload `json:"files"`
		Profiles []byte    `json:"profiles"`
		Renderer string    `json:"renderer"`
	}{r.Files, r.Profiles, r.Renderer}))
}
func planID(p Plan) string { p.ID = ""; return hash(encode(p)) }
func emptyState() State    { return State{Version: 4, Records: map[string]Record{}} }
func normalize(o Options) (target.Config, string, error) {
	if o.Scope != "user" && o.Scope != "project" {
		return target.Config{}, "", fmt.Errorf("explicit scope must be user or project")
	}
	synthetic := o.Home != ""
	if o.Home == "" {
		var err error
		o.Home, err = os.UserHomeDir()
		if err != nil {
			return target.Config{}, "", err
		}
	}
	home, err := target.Canonical(o.Home)
	if err != nil {
		return target.Config{}, "", err
	}
	c := target.Config{Scope: o.Scope, Home: home, CodexHome: filepath.Join(home, ".codex"), ClaudeHome: filepath.Join(home, ".claude")}
	if !synthetic {
		if v := os.Getenv("CODEX_HOME"); v != "" {
			c.CodexHome = v
		}
		if v := os.Getenv("CLAUDE_CONFIG_DIR"); v != "" {
			c.ClaudeHome = v
		}
	}
	c.CodexHome, err = target.Canonical(c.CodexHome)
	if err != nil {
		return c, "", err
	}
	c.ClaudeHome, err = target.Canonical(c.ClaudeHome)
	if err != nil {
		return c, "", err
	}
	c, err = target.ExpandHostHomes(c, synthetic)
	if err != nil {
		return c, "", err
	}
	if o.Scope == "project" {
		if o.Root == "" {
			return c, "", fmt.Errorf("project scope requires --root")
		}
		c.Root, err = target.Canonical(o.Root)
		if err != nil {
			return c, "", err
		}
	}
	if o.StateDir == "" {
		o.StateDir = filepath.Join(home, "Library", "Application Support", "tricell-hive")
	}
	state, err := target.Canonical(o.StateDir)
	if err != nil {
		return c, "", err
	}
	return c, state, target.Safe(state)
}
func validateHosts(hosts []string) ([]string, error) {
	h := append([]string(nil), hosts...)
	sort.Strings(h)
	if len(h) == 0 {
		return nil, fmt.Errorf("explicit hosts required")
	}
	for i, s := range h {
		if s != "codex" && s != "claude" && s != "grok" && s != "pi" && s != "opencode" && s != "cursor" {
			return nil, fmt.Errorf("unsupported host %q", s)
		}
		if i > 0 && s == h[i-1] {
			return nil, fmt.Errorf("duplicate host %q", s)
		}
	}
	return h, nil
}
