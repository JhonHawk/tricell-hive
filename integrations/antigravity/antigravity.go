// Package antigravity resolves the user-level locations shared by the agy CLI
// and the Antigravity App, which both read ~/.gemini/config. Only user scope
// exists. This package currently resolves skills; rules, agents and hooks are
// added by later changes.
package antigravity

import (
	"fmt"
	"path/filepath"

	"tricell-hive/integrations/target"
)

// SkillsDir returns the shared store the skill links of c point into.
func SkillsDir(c target.Config) string {
	return filepath.Join(c.Home, ".agents", "skills")
}

// AgentsDir is the folder Antigravity loads global agents from.
func AgentsDir(antigravityHome string) string {
	return filepath.Join(antigravityHome, "agents")
}

func resolveBase(c target.Config) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, fmt.Errorf("Antigravity project scope is unsupported")
	}
	if c.AntigravityHome == "" {
		return nil, fmt.Errorf("Antigravity home is required")
	}
	sharedSkillDir := filepath.Join(SkillsDir(c), "workspace-conventions")
	skillLink := filepath.Join(c.AntigravityHome, "skills", "workspace-conventions")
	linkTarget, err := filepath.Rel(filepath.Dir(skillLink), sharedSkillDir)
	if err != nil {
		return nil, err
	}
	return []target.Target{
		{Path: filepath.Join(sharedSkillDir, "SKILL.md"), Kind: "skill", Host: "antigravity", Scope: c.Scope, Context: c.Home},
		{Path: skillLink, Kind: "symlink", Host: "antigravity", Scope: c.Scope, Context: c.Home, LinkTarget: linkTarget},
	}, nil
}

// Resolve expands native locations for the supplied source catalogue. It
// returns no guidance block yet: the rule parts arrive with a later change.
func Resolve(c target.Config, sources ...string) ([]target.Target, error) {
	base, err := resolveBase(c)
	if err != nil {
		return nil, err
	}
	return target.ExpandSkills(base, sources), nil
}
