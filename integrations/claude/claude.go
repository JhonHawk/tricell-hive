// Package claude resolves Claude Code's native instruction and skill locations.
package claude

import (
	"fmt"
	"path/filepath"
	"tricell-hive/integrations/target"
)

// SkillsDir returns the directory Hive installs skills into for c. At user
// scope that is the shared store under the home directory, which Claude's own
// skills directory links into; at project scope it is the project's.
func SkillsDir(c target.Config) string {
	if c.Scope == "project" {
		return filepath.Join(c.Root, ".claude", "skills")
	}
	return filepath.Join(c.Home, ".agents", "skills")
}

func resolveBase(c target.Config) ([]target.Target, error) {
	base, context := c.ClaudeHome, c.Home
	instruction := filepath.Join(base, "CLAUDE.md")
	skills := filepath.Join(base, "skills")
	if c.Scope == "project" {
		context = c.Root
		instruction = filepath.Join(c.Root, "CLAUDE.md")
		skills = SkillsDir(c)
		return []target.Target{{Path: instruction, Kind: "block", Host: "claude", Scope: c.Scope, Context: context}, {Path: filepath.Join(skills, "workspace-conventions", "SKILL.md"), Kind: "skill", Host: "claude", Scope: c.Scope, Context: context}}, nil
	}
	if c.Scope != "user" {
		return nil, fmt.Errorf("explicit scope must be user or project")
	}
	sharedSkillDir := filepath.Join(SkillsDir(c), "workspace-conventions")
	skillLink := filepath.Join(skills, "workspace-conventions")
	linkTarget, err := filepath.Rel(filepath.Dir(skillLink), sharedSkillDir)
	if err != nil {
		return nil, err
	}
	return []target.Target{
		{Path: instruction, Kind: "block", Host: "claude", Scope: c.Scope, Context: context},
		{Path: filepath.Join(sharedSkillDir, "SKILL.md"), Kind: "skill", Host: "claude", Scope: c.Scope, Context: context},
		{Path: skillLink, Kind: "symlink", Host: "claude", Scope: c.Scope, Context: context, LinkTarget: linkTarget},
	}, nil
}

// Resolve expands native locations for the supplied source catalogue.
func Resolve(c target.Config, sources ...string) ([]target.Target, error) {
	base, err := resolveBase(c)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(c.ClaudeHome, "agents")
	if c.Scope == "project" {
		dir = filepath.Join(c.Root, ".claude", "agents")
	}
	out := target.ExpandSkills(base, sources)
	return target.ExpandAgents(out, sources, dir, ".md", base[0]), nil
}
