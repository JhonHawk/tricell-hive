// Package antigravity resolves the user-level locations shared by the agy CLI
// and the Antigravity App, which both read ~/.gemini/config. Only user scope
// exists. This package resolves skills and agents and names the rule files
// that carry the global guidance; hooks are added by a later change.
package antigravity

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

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

// RuleFrontmatter makes a rule file apply to every request. The manager writes
// it before the managed block when it creates a rule part.
const RuleFrontmatter = "---\ntrigger: always_on\n---\n\n"

// MaxRuleBytes is the size above which agy truncates a rule file, the global
// ones included (observed with agy 1.5.2, 2026-11-06). RuleBudget is what Hive
// plans a part to: a margin below the truncation point for edits to the guidance.
const (
	MaxRuleBytes = 24000
	RuleBudget   = 22000
	// MaxRuleParts is the most parts the two-digit file name can number.
	MaxRuleParts = 99
)

var rulePartName = regexp.MustCompile(`^hive-([0-9]{2})\.md$`)

// RulesDir is the folder Antigravity loads global rules from.
func RulesDir(antigravityHome string) string {
	return filepath.Join(antigravityHome, "rules")
}

// RulePath is the file that carries part (1-based) of the global guidance.
func RulePath(antigravityHome string, part int) string {
	return filepath.Join(RulesDir(antigravityHome), fmt.Sprintf("hive-%02d.md", part))
}

// RulePartNumber reports whether path is a Hive rule part under
// antigravityHome and, if so, its 1-based part number.
func RulePartNumber(antigravityHome, path string) (int, bool) {
	if antigravityHome == "" || filepath.Dir(path) != RulesDir(antigravityHome) {
		return 0, false
	}
	m := rulePartName.FindStringSubmatch(filepath.Base(path))
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
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
// returns no guidance block: how many rule parts exist depends on the release,
// so the manager adds them (see RulePath).
func Resolve(c target.Config, sources ...string) ([]target.Target, error) {
	base, err := resolveBase(c)
	if err != nil {
		return nil, err
	}
	out := target.ExpandSkills(base, sources)
	return target.ExpandAgents(out, sources, AgentsDir(c.AntigravityHome), ".md", base[0]), nil
}
