// Package cursor resolves Cursor CLI's supported user-level guidance targets.
// Cursor CLI does not load ~/.cursor/AGENTS.md. It loads <dir>/.cursor/rules/*.mdc
// for every directory from the workspace up to the filesystem root, so an
// always-applied rule in ~/.cursor/rules reaches every workspace under the home
// directory (observed with cursor-agent 2026.10.22, 2026-10-31); a workspace
// outside the home directory does not get it.
package cursor

import (
	"fmt"
	"path/filepath"

	"tricell-hive/integrations/target"
)

// SkillsDir returns the directory Hive installs skills into for c. User scope
// uses the shared store under the home directory.
func SkillsDir(c target.Config) string {
	return filepath.Join(c.Home, ".agents", "skills")
}

// RuleFrontmatter makes the guidance rule apply to every request. The manager
// writes it before the managed block when it creates the rule file.
const RuleFrontmatter = "---\ndescription: Tricell Hive guidance\nalwaysApply: true\n---\n\n"

// RulePath is the always-applied rule that carries the Hive guidance.
func RulePath(cursorHome string) string {
	return filepath.Join(cursorHome, "rules", "hive.mdc")
}

func resolveBase(c target.Config) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, fmt.Errorf("Cursor project scope is unsupported")
	}
	if c.CursorHome == "" {
		return nil, fmt.Errorf("Cursor home is required")
	}
	skill := filepath.Join(SkillsDir(c), "workspace-conventions", "SKILL.md")
	return []target.Target{
		{Path: RulePath(c.CursorHome), Kind: "block", Host: "cursor", Scope: c.Scope, Context: c.Home},
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
