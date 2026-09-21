// Package codex resolves documented Codex instruction and skill locations.
package codex

import (
	"os"
	"path/filepath"
	"tricell-hive/integrations/target"
)

func Resolve(c target.Config) ([]target.Target, error) {
	base, skills, context := c.CodexHome, filepath.Join(c.Home, ".agents", "skills"), c.Home
	if c.Scope == "project" {
		base = c.Root
		skills = filepath.Join(c.Root, ".agents", "skills")
		context = c.Root
	}
	instruction := filepath.Join(base, "AGENTS.md")
	override := filepath.Join(base, "AGENTS.override.md")
	if err := target.Safe(override); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(override)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if len(b) > 0 {
		instruction = override
	}
	return []target.Target{{Path: instruction, Kind: "block", Host: "codex", Scope: c.Scope, Context: context}, {Path: filepath.Join(skills, "workspace-conventions", "SKILL.md"), Kind: "skill", Host: "codex", Scope: c.Scope, Context: context}}, nil
}
