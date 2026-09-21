// Package pi resolves Pi's native user-level guidance targets.
package pi

import (
	"fmt"
	"path/filepath"

	"tricell-hive/integrations/target"
)

func resolveBase(c target.Config) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, fmt.Errorf("Pi project scope is unsupported")
	}
	if c.PiHome == "" {
		return nil, fmt.Errorf("Pi home is required")
	}
	skill := filepath.Join(c.Home, ".agents", "skills", "workspace-conventions", "SKILL.md")
	return []target.Target{
		{Path: filepath.Join(c.PiHome, "AGENTS.md"), Kind: "block", Host: "pi", Scope: c.Scope, Context: c.Home},
		{Path: skill, Kind: "skill", Host: "pi", Scope: c.Scope, Context: c.Home},
	}, nil
}

// Resolve expands native locations for the supplied source catalogue.
func Resolve(c target.Config, sources ...string) ([]target.Target, error) {
	base, err := resolveBase(c)
	if err != nil {
		return nil, err
	}
	return target.ExpandSkills(base, sources), nil
}
