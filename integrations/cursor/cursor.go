// Package cursor resolves Cursor CLI's supported user-level guidance targets.
// Cursor does not load ~/.cursor/AGENTS.md natively; a project pointer decides
// whether a session reads it. That pointer is distributed content, not code.
package cursor

import (
	"fmt"
	"path/filepath"

	"tricell-hive/integrations/target"
)

func resolveBase(c target.Config) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, fmt.Errorf("Cursor project scope is unsupported")
	}
	if c.CursorHome == "" {
		return nil, fmt.Errorf("Cursor home is required")
	}
	skill := filepath.Join(c.Home, ".agents", "skills", "workspace-conventions", "SKILL.md")
	return []target.Target{
		{Path: filepath.Join(c.CursorHome, "AGENTS.md"), Kind: "block", Host: "cursor", Scope: c.Scope, Context: c.Home},
		{Path: skill, Kind: "skill", Host: "cursor", Scope: c.Scope, Context: c.Home},
	}, nil
}

// Resolve expands native locations for the supplied source catalogue.
func Resolve(c target.Config, sources ...string) ([]target.Target, error) {
	base, err := resolveBase(c)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(c.CursorHome, "agents")
	out := target.ExpandSkills(base, sources)
	return target.ExpandAgents(out, sources, dir, ".md", base[0]), nil
}
