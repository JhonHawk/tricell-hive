// Package opencode resolves OpenCode V2's native user-level guidance targets.
package opencode

import (
	"fmt"
	"path/filepath"

	"tricell-hive/integrations/target"
)

// ContextLoaderSource is the only editorial plugin payload allowed in releases.
const ContextLoaderSource = "integrations/opencode/context-loader.js"

// SkillsDir returns the directory Hive installs skills into for c. User scope
// uses the shared store under the home directory.
func SkillsDir(c target.Config) string {
	return filepath.Join(c.Home, ".agents", "skills")
}

func resolveBase(c target.Config) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, fmt.Errorf("OpenCode project scope is unsupported")
	}
	if c.OpenCodeHome == "" {
		return nil, fmt.Errorf("OpenCode home is required")
	}
	skill := filepath.Join(SkillsDir(c), "workspace-conventions", "SKILL.md")
	return []target.Target{
		{Path: filepath.Join(c.OpenCodeHome, "AGENTS.md"), Kind: "block", Host: "opencode", Scope: c.Scope, Context: c.Home},
		{Path: skill, Kind: "skill", Host: "opencode", Scope: c.Scope, Context: c.Home},
	}, nil
}

// Resolve expands native locations for the supplied source catalogue.
func Resolve(c target.Config, sources ...string) ([]target.Target, error) {
	base, err := resolveBase(c)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(c.OpenCodeHome, "agents")
	out := target.ExpandSkills(base, sources)
	out = target.ExpandAgents(out, sources, dir, ".md", base[0])
	for _, source := range sources {
		if source != ContextLoaderSource {
			continue
		}
		for _, name := range []string{"index.js", "context-loader.json"} {
			t := base[0]
			t.Kind, t.Source, t.Path = "plugin", source, filepath.Join(c.OpenCodeHome, "hive", "context-loader", name)
			out = append(out, t)
		}
	}
	return out, nil
}
