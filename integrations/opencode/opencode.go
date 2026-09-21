// Package opencode resolves OpenCode V2's native user-level guidance targets.
package opencode

import (
	"fmt"
	"path/filepath"

	"tricell-hive/integrations/target"
)

func Resolve(c target.Config) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, fmt.Errorf("OpenCode project scope is unsupported")
	}
	if c.OpenCodeHome == "" {
		return nil, fmt.Errorf("OpenCode home is required")
	}
	skill := filepath.Join(c.Home, ".agents", "skills", "workspace-conventions", "SKILL.md")
	return []target.Target{
		{Path: filepath.Join(c.OpenCodeHome, "AGENTS.md"), Kind: "block", Host: "opencode", Scope: c.Scope, Context: c.Home},
		{Path: skill, Kind: "skill", Host: "opencode", Scope: c.Scope, Context: c.Home},
	}, nil
}
