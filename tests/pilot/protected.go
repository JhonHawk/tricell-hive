package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

func envRoot(name, fallback string) string {
	if p := os.Getenv(name); p != "" {
		return p
	}
	return fallback
}

// Every worker monitors shared immutable guidance, but only its own mutable
// configuration. The batch coordinator owns cross-host reconciliation.
func protectedFor(host, home string) map[string]string {
	codex := envRoot("CODEX_HOME", filepath.Join(home, ".codex"))
	claude := envRoot("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))
	grok := envRoot("GROK_HOME", filepath.Join(home, ".grok"))
	pi := envRoot("PI_CODING_AGENT_DIR", filepath.Join(home, ".pi", "agent"))
	oc := filepath.Join(envRoot("XDG_CONFIG_HOME", filepath.Join(home, ".config")), "opencode")
	return protectedPaths(host, home, codex, claude, grok, pi, oc)
}

// protectedForHome is protectedFor's env-independent counterpart, for a
// guidance variant's shadow home. protectedFor(host, userHome) is correct
// for the REAL home: an actual invocation honors the runner process's own
// CODEX_HOME/CLAUDE_CONFIG_DIR/GROK_HOME/PI_CODING_AGENT_DIR/XDG_CONFIG_HOME,
// so watching wherever those point is right. But main.go also calls
// protectedFor a second time for the shadow home guidanceVariant creates,
// and reusing the same env-aware lookup there silently collapses the shadow
// snapshot onto whatever those variables happen to be set to in the runner's
// own environment (for example a real CODEX_HOME exported in the operator's
// shell) instead of the shadow home's own layout — exactly what
// applyEnvironment sets for the child (CODEX_HOME=<shadow>/.codex, and for
// Grok the real, unshadowed GROK_HOME, per its own declared limitation).
// protectedForHome always resolves every sub-path directly under home,
// ignoring the process environment entirely, so the shadow-home audit stays
// scoped to the shadow home regardless of what the operator has exported.
func protectedForHome(host, home string) map[string]string {
	return protectedPaths(host, home, filepath.Join(home, ".codex"), filepath.Join(home, ".claude"), filepath.Join(home, ".grok"), filepath.Join(home, ".pi", "agent"), filepath.Join(home, ".config", "opencode"))
}

func protectedPaths(host, home, codex, claude, grok, pi, oc string) map[string]string {
	paths := []string{
		filepath.Join(codex, "AGENTS.md"), filepath.Join(codex, "AGENTS.override.md"),
		filepath.Join(claude, "CLAUDE.md"), filepath.Join(claude, "rules"),
		filepath.Join(grok, "AGENTS.md"), filepath.Join(grok, "rules"),
		filepath.Join(pi, "AGENTS.md"), filepath.Join(pi, "AGENTS.override.md"), filepath.Join(pi, "SYSTEM.md"), filepath.Join(pi, "APPEND_SYSTEM.md"),
		filepath.Join(oc, "AGENTS.md"),
	}
	for _, base := range []string{filepath.Join(home, ".agents"), codex, claude, grok, pi, oc} {
		for _, skill := range []string{"workspace-conventions", "flow-research", "flow-plan", "flow-build", "git-workflow"} {
			paths = append(paths, filepath.Join(base, "skills", skill))
		}
	}
	configs := map[string][]string{
		"codex":    {filepath.Join(codex, "config.toml")},
		"claude":   {filepath.Join(claude, "settings.json"), filepath.Join(claude, "settings.local.json")},
		"grok":     {filepath.Join(grok, "config.toml")},
		"pi":       {filepath.Join(pi, "settings.json"), filepath.Join(pi, "models.json")},
		"opencode": {filepath.Join(oc, "opencode.json"), filepath.Join(oc, "opencode.jsonc")},
	}
	paths = append(paths, configs[host]...)
	out := map[string]string{}
	for _, p := range paths {
		out[p] = fingerprint(p, map[string]bool{})
	}
	return out
}

// Hash both link identity and followed payload. Only hashes/statuses leave this
// function; configuration bytes are never copied to evidence.
func fingerprint(path string, active map[string]bool) string {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "absent"
	}
	if err != nil {
		return "unreadable"
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(path)
		if err != nil {
			return "unreadable-link"
		}
		real, err := filepath.EvalSymlinks(path)
		if err != nil {
			return digest([]byte("broken-link:" + link))
		}
		return digest([]byte("link:" + link + "\nreal:" + real + "\npayload:" + fingerprint(real, active)))
	}
	if info.IsDir() {
		real, err := filepath.EvalSymlinks(path)
		if err != nil {
			return "unreadable-dir"
		}
		if active[real] {
			return "cycle"
		}
		active[real] = true
		defer delete(active, real)
		entries, err := os.ReadDir(path)
		if err != nil {
			return "unreadable-dir"
		}
		parts := []string{}
		for _, entry := range entries {
			parts = append(parts, entry.Name()+":"+fingerprint(filepath.Join(path, entry.Name()), active))
		}
		b, _ := json.Marshal(parts)
		return digest(b)
	}
	if !info.Mode().IsRegular() {
		return "nonregular:" + info.Mode().String()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "unreadable"
	}
	return digest(b)
}

func changedProtected(before, after map[string]string) []string {
	set := map[string]bool{}
	for p, v := range before {
		if after[p] != v {
			set[p] = true
		}
	}
	for p, v := range after {
		if before[p] != v {
			set[p] = true
		}
	}
	paths := []string{}
	for p := range set {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
