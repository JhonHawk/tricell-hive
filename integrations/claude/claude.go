// Package claude resolves Claude Code's native instruction and skill locations.
package claude

import (
	"fmt"
	"path/filepath"
	"tricell-hive/integrations/target"
)

func Resolve(c target.Config) ([]target.Target, error) {
	base, context := c.ClaudeHome, c.Home
	instruction := filepath.Join(base, "CLAUDE.md")
	skills := filepath.Join(base, "skills")
	if c.Scope == "project" {
		context = c.Root
		instruction = filepath.Join(c.Root, "CLAUDE.md")
		skills = filepath.Join(c.Root, ".claude", "skills")
		return []target.Target{{Path: instruction, Kind: "block", Host: "claude", Scope: c.Scope, Context: context}, {Path: filepath.Join(skills, "workspace-conventions", "SKILL.md"), Kind: "skill", Host: "claude", Scope: c.Scope, Context: context}}, nil
	}
	if c.Scope != "user" {
		return nil, fmt.Errorf("explicit scope must be user or project")
	}
	sharedSkillDir := filepath.Join(c.Home, ".agents", "skills", "workspace-conventions")
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
