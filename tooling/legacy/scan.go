// Package legacy recognizes only the pinned Hive distribution. It never writes.
package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"tricell-hive/integrations/target"
)

const Version = "legacy-hive-16e7d335-v1"

//go:embed catalog.json
var catalogJSON []byte

//go:embed pi-subagents-0.67.0.json
var patchJSON []byte

//go:embed pi-subagents-0.67.0.patch
var patchText string

type Edit struct {
	Path          string
	Before, After []byte
	Delete        bool
	Mode          uint32
	Hosts         []string
	Kind          string
}

// ModifiedFileError reports a file at a path Hive once used that differs from
// what Hive expects there. The file may be one Hive installed and the user
// edited, or the user's own, so the message does not call it legacy. Callers
// find the path with errors.As instead of reading the message.
type ModifiedFileError struct{ Path string }

func (e *ModifiedFileError) Error() string {
	return e.Path + " differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere"
}

type Result struct {
	Edits    []Edit
	Detected bool
	Hosts    []string
}
type fileRecord struct {
	Root, Path, Source, SHA256, Manifest string
	Data                                 []byte
}
type hookRecord struct{ Host, Event, Command string }
type catalog struct {
	Revision string
	Files    []fileRecord
	Hooks    []hookRecord
}

var catalogOnce sync.Once
var pinnedCatalog catalog

func loadCatalog() catalog {
	catalogOnce.Do(func() {
		if e := json.Unmarshal(catalogJSON, &pinnedCatalog); e != nil {
			panic(e)
		}
	})
	return pinnedCatalog
}

func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

var allHosts = []string{"claude", "codex", "grok", "opencode", "pi"}

func owners(root string) []string {
	if root == "shared" {
		return allHosts
	}
	if root == "claude" {
		return []string{"claude", "grok"}
	}
	return []string{root}
}
func roots(c target.Config) map[string][]string {
	m := map[string][]string{}
	add := func(k string, v ...string) {
		for _, p := range v {
			if p == "" {
				continue
			}
			p = filepath.Clean(p)
			found := false
			for _, x := range m[k] {
				found = found || x == p
			}
			if !found {
				m[k] = append(m[k], p)
			}
		}
	}
	add("claude", c.ClaudeHome, filepath.Join(c.Home, ".claude"))
	add("codex", c.CodexHome, filepath.Join(c.Home, ".codex"))
	add("grok", c.GrokHome, filepath.Join(c.Home, ".grok"))
	add("pi", c.PiHome, filepath.Join(c.Home, ".pi", "agent"))
	add("opencode", c.OpenCodeHome, filepath.Join(c.Home, ".config", "opencode"))
	add("shared", filepath.Join(c.Home, ".agents"))
	return m
}
func read(p string) ([]byte, os.FileMode, error) {
	if e := target.Safe(p); e != nil {
		return nil, 0, e
	}
	i, e := os.Lstat(p)
	if os.IsNotExist(e) {
		return nil, 0, nil
	}
	if e != nil {
		return nil, 0, e
	}
	if !i.Mode().IsRegular() {
		return nil, 0, fmt.Errorf("legacy target is not regular: %s", p)
	}
	b, e := os.ReadFile(p)
	return b, i.Mode().Perm(), e
}
func safeRelative(p string) bool {
	return p != "" && !filepath.IsAbs(p) && filepath.ToSlash(filepath.Clean(p)) == p && p != "." && !strings.HasPrefix(p, "../") && !strings.ContainsAny(p, "\x00\r\n\\")
}

// Scan produces a complete read-only proposal or a conflict; callers must not
// apply a partial result. All consumers of shared resources must be selected.
func Scan(c target.Config, hosts []string) (Result, error) {
	return ScanExcluding(c, hosts, nil)
}

// ScanExcluding excludes only engine-verified currently owned resources.
func ScanExcluding(c target.Config, hosts []string, trustedPaths []string) (Result, error) {
	var result Result
	if c.Scope != "user" {
		return result, nil
	}
	selected := map[string]bool{}
	for _, h := range hosts {
		selected[h] = true
	}
	rs := roots(c)
	cat := loadCatalog()
	edits := map[string]Edit{}
	required := map[string]bool{}
	claims := map[string]bool{}
	dirs := map[string][]string{}
	add := func(e Edit) error {
		for _, h := range e.Hosts {
			if !selected[h] {
				return fmt.Errorf("legacy resource %s also requires host %s", e.Path, h)
			}
			required[h] = true
		}
		if old, ok := edits[e.Path]; ok && !bytes.Equal(old.After, e.After) {
			return fmt.Errorf("conflicting legacy edits: %s", e.Path)
		}
		edits[e.Path] = e
		result.Detected = true
		return nil
	}
	enabled := func(root string) bool {
		for _, h := range owners(root) {
			if selected[h] {
				return true
			}
		}
		return false
	}
	// Validate the path-only manifest against the embedded catalog, never against
	// paths supplied by the manifest itself.
	for _, root := range rs["claude"] {
		p := filepath.Join(root, ".deploy-manifest")
		b, mode, e := read(p)
		if e != nil {
			return result, e
		}
		if b == nil {
			continue
		}
		lines := strings.Split(string(b), "\n")
		validHeader := false
		var used []string
		for _, line := range lines {
			if strings.HasPrefix(line, "# source_commit: ") {
				v := strings.TrimPrefix(line, "# source_commit: ")
				validHeader = len(v) >= 7 && strings.HasPrefix(cat.Revision, v)
			}
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !safeRelative(line) {
				return result, fmt.Errorf("invalid legacy manifest entry in %s", p)
			}
			found := false
			for _, f := range cat.Files {
				if f.Manifest != line {
					continue
				}
				found = true
				for _, r := range rs[f.Root] {
					claims[filepath.Join(r, filepath.FromSlash(f.Path))] = true
				}
				used = append(used, owners(f.Root)...)
			}
			if !found {
				return result, fmt.Errorf("unsupported legacy manifest entry %s", line)
			}
		}
		if !validHeader {
			return result, fmt.Errorf("unsupported legacy manifest revision: %s", p)
		}
		used = unique(used)
		active := false
		for _, h := range used {
			active = active || selected[h]
		}
		if active {
			if e = add(Edit{Path: p, Before: b, Delete: true, Mode: uint32(mode), Hosts: used}); e != nil {
				return result, e
			}
		}
	}
	for _, kind := range []string{"shared", "pi"} {
		if !enabled(kind) {
			continue
		}
		for _, root := range rs[kind] {
			p := filepath.Join(root, ".hive-deploy-manifest.json")
			b, mode, e := read(p)
			if e != nil {
				return result, e
			}
			if b == nil {
				continue
			}
			var m struct {
				SchemaVersion int
				Scope         string
				ManagedFiles  map[string]struct{ SHA256, Source, Kind string }
				ManagedConfig map[string]json.RawMessage
			}
			if e = uniqueJSON(b); e != nil {
				return result, fmt.Errorf("invalid legacy manifest keys: %s", p)
			}
			if e = json.Unmarshal(b, &m); e != nil {
				return result, fmt.Errorf("invalid legacy manifest: %s", p)
			}
			scope := kind
			if kind == "shared" {
				scope = "shared-skills"
			}
			if m.SchemaVersion != 1 || m.Scope != scope {
				return result, fmt.Errorf("unsupported legacy manifest: %s", p)
			}
			for rel, record := range m.ManagedFiles {
				if !safeRelative(rel) {
					return result, fmt.Errorf("unsafe legacy path in %s", p)
				}
				found := false
				for _, f := range cat.Files {
					if f.Root == kind && f.Path == rel {
						data := render(f, root)
						if record.SHA256 != digest(data) || record.Source != f.Source {
							return result, fmt.Errorf("unsupported legacy fingerprint: %s", rel)
						}
						found = true
						break
					}
				}
				if !found && kind == "pi" {
					found = knownPatch(rel, record.SHA256)
				}
				if !found {
					return result, fmt.Errorf("unsupported legacy file: %s", rel)
				}
				claims[filepath.Join(root, filepath.FromSlash(rel))] = true
			}
			for rel, managed := range m.ManagedConfig {
				if kind != "pi" || (rel != "settings.json" && rel != "mcp.json" && rel != "web-search.json" && rel != "extensions/subagent/config.json") {
					return result, fmt.Errorf("unsupported managed configuration: %s", rel)
				}
				es, err := cleanPiConfig(filepath.Join(root, filepath.FromSlash(rel)), rel, managed, root)
				if err != nil {
					return result, err
				}
				for _, edit := range es {
					if err = add(edit); err != nil {
						return result, err
					}
				}
			}
			if e = add(Edit{Path: p, Before: b, Delete: true, Mode: uint32(mode), Hosts: owners(kind)}); e != nil {
				return result, e
			}
		}
	}
	for _, f := range cat.Files {
		if !enabled(f.Root) {
			continue
		}
		if digest(f.Data) != f.SHA256 {
			return result, fmt.Errorf("embedded catalog corrupted")
		}
		for _, root := range rs[f.Root] {
			p := filepath.Join(root, filepath.FromSlash(f.Path))
			if excluded(p, trustedPaths) {
				continue
			}
			b, mode, e := read(p)
			if e != nil {
				return result, e
			}
			if b == nil {
				continue
			}
			want := render(f, root)
			if !bytes.Equal(b, want) {
				if claims[p] || (f.Path != "AGENTS.md" && f.Path != "CLAUDE.md") || legacyMarker(b, want) {
					return result, &ModifiedFileError{Path: p}
				}
				continue
			}
			if e = add(Edit{Path: p, Before: b, Delete: true, Mode: uint32(mode), Hosts: owners(f.Root)}); e != nil {
				return result, e
			}
			if strings.HasPrefix(f.Path, "skills/") {
				parts := strings.Split(f.Path, "/")
				if len(parts) >= 3 {
					dirs[filepath.Join(root, "skills", parts[1])] = owners(f.Root)
				}
			}
		}
	}
	for _, host := range []string{"claude", "codex"} {
		if !enabled(host) {
			continue
		}
		for _, root := range rs[host] {
			name := "settings.json"
			if host == "codex" {
				name = "hooks.json"
			}
			p := filepath.Join(root, name)
			e, ok, err := cleanHooks(p, host, cat.Hooks)
			if err != nil {
				return result, err
			}
			if ok {
				e.Hosts = owners(host)
				if err = add(e); err != nil {
					return result, err
				}
			}
		}
	}
	if enabled("pi") {
		for _, root := range rs["pi"] {
			for _, rel := range []string{"settings.json", "mcp.json", "web-search.json", "extensions/subagent/config.json"} {
				if _, err := cleanPiConfig(filepath.Join(root, filepath.FromSlash(rel)), rel, json.RawMessage(`{}`), root); err != nil {
					return result, err
				}
			}
			es, err := reversePatches(root, claims)
			if err != nil {
				return result, err
			}
			for _, e := range es {
				if err = add(e); err != nil {
					return result, err
				}
			}
		}
	}
	// Only remove directories proven to contain exclusively recognized files.
	for dir, hs := range dirs {
		var pending []Edit
		err := filepath.WalkDir(dir, func(p string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() {
				i, e := d.Info()
				if e != nil {
					return e
				}
				pending = append(pending, Edit{Path: p, Kind: "directory", Delete: true, Mode: uint32(i.Mode().Perm()), Hosts: hs})
				return nil
			}
			if e, ok := edits[p]; !ok || !e.Delete {
				return fmt.Errorf("unrecognized file in legacy skill directory: %s", p)
			}
			return nil
		})
		if err != nil {
			return result, err
		}
		for _, e := range pending {
			if err = add(e); err != nil {
				return result, err
			}
		}
	}
	for _, e := range edits {
		result.Edits = append(result.Edits, e)
	}
	sort.Slice(result.Edits, func(i, j int) bool {
		a, b := result.Edits[i], result.Edits[j]
		rank := func(e Edit) int {
			if !e.Delete {
				return 0
			}
			if e.Kind == "directory" {
				return 2
			}
			return 1
		}
		if rank(a) != rank(b) {
			return rank(a) < rank(b)
		}
		if a.Kind == "directory" && len(a.Path) != len(b.Path) {
			return len(a.Path) > len(b.Path)
		}
		return a.Path < b.Path
	})
	for h := range required {
		result.Hosts = append(result.Hosts, h)
	}
	sort.Strings(result.Hosts)
	return result, nil
}
func render(f fileRecord, root string) []byte {
	if f.Root == "pi" {
		return bytes.ReplaceAll(f.Data, []byte("__HIVE_PI_ROOT__"), []byte(root))
	}
	return f.Data
}
func unique(ss []string) []string {
	m := map[string]bool{}
	for _, s := range ss {
		m[s] = true
	}
	var out []string
	for s := range m {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
func legacyMarker(b, want []byte) bool {
	if bytes.Contains(b, []byte("global/core-sections")) || bytes.Contains(b, []byte("# Agent Harness Configuration")) || bytes.Contains(b, []byte("flow-core")) {
		return true
	}
	return len(want) > 200 && len(b) > 200 && bytes.Equal(b[:200], want[:200])
}

func excluded(p string, paths []string) bool {
	for _, x := range paths {
		if p == x || strings.HasPrefix(p, x+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
