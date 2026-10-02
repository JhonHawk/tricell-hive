// Package grok resolves Grok Build's supported user-level guidance targets.
package grok

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"tricell-hive/integrations/target"
)

// SkillsDir returns the directory Hive installs skills into for c. User scope
// uses the shared store under the home directory.
func SkillsDir(c target.Config) string {
	return filepath.Join(c.Home, ".agents", "skills")
}

// Resolve returns Grok's Claude-compatible home instruction target and the
// shared skill target. Grok has no native skill copy in this integration.
func resolveBase(c target.Config) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, fmt.Errorf("Grok project scope is unsupported")
	}
	if c.GrokHome == "" {
		return nil, fmt.Errorf("Grok home is required")
	}
	if err := claudeAgentsEnabled(c.GrokHome, c.Synthetic); err != nil {
		return nil, err
	}
	skill := filepath.Join(SkillsDir(c), "workspace-conventions", "SKILL.md")
	return []target.Target{
		{Path: filepath.Join(c.Home, ".claude", "CLAUDE.md"), Kind: "block", Host: "grok", Scope: c.Scope, Context: c.Home},
		{Path: skill, Kind: "skill", Host: "grok", Scope: c.Scope, Context: c.Home},
	}, nil
}

// claudeAgentsEnabled resolves the documented env > TOML > default-on setting.
// This parser deliberately supports only the one boolean cell the resolver
// needs and fails closed on ambiguous representations.
func claudeAgentsEnabled(home string, synthetic bool) error {
	if value, ok := os.LookupEnv("GROK_CLAUDE_AGENTS_ENABLED"); ok && !synthetic {
		enabled, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("unsupported GROK_CLAUDE_AGENTS_ENABLED value")
		}
		if !enabled {
			return fmt.Errorf("Grok compat.claude.agents is disabled")
		}
		return nil
	}

	config := filepath.Join(home, "config.toml")
	f, err := os.Open(config)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read Grok config: %w", err)
	}
	defer f.Close()

	section := ""
	seenTable := false
	var value string
	valueSet := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := stripTOMLComment(scanner.Text())
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			if strings.HasPrefix(trimmed, "[[") || !strings.HasSuffix(trimmed, "]") {
				if isClaudeCompatTable(trimmed) {
					return fmt.Errorf("unsupported Grok compat.claude TOML table")
				}
				section = ""
				continue
			}
			section = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			if isClaudeCompatTable(section) && section != "compat.claude" {
				return fmt.Errorf("unsupported Grok compat.claude TOML table")
			}
			if section == "compat.claude" {
				if seenTable {
					return fmt.Errorf("duplicate Grok compat.claude TOML table")
				}
				seenTable = true
			}
			continue
		}
		if section == "" {
			key := strings.TrimSpace(strings.SplitN(trimmed, "=", 2)[0])
			if normalizeTOMLKey(key) == "compat" || strings.Contains(strings.ToLower(key), "compat.claude") {
				return fmt.Errorf("unsupported Grok compat.claude TOML form")
			}
			continue
		}
		if section == "compat" {
			key := strings.TrimSpace(strings.SplitN(trimmed, "=", 2)[0])
			if normalizeTOMLKey(key) == "claude" || strings.Contains(strings.ToLower(trimmed), "claude") {
				return fmt.Errorf("unsupported Grok compat.claude TOML form")
			}
			continue
		}
		if section != "compat.claude" {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("unsupported Grok compat.claude TOML entry")
		}
		key := strings.TrimSpace(parts[0])
		if normalizeTOMLKey(key) == "agents" && key != "agents" {
			return fmt.Errorf("unsupported quoted Grok compat.claude.agents key")
		}
		if key != "agents" {
			continue
		}
		if valueSet {
			return fmt.Errorf("duplicate Grok compat.claude.agents setting")
		}
		value = strings.TrimSpace(parts[1])
		if value == "" {
			return fmt.Errorf("unsupported Grok compat.claude.agents value")
		}
		valueSet = true
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read Grok config: %w", err)
	}
	if !valueSet {
		return nil
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("unsupported Grok compat.claude.agents value")
	}
	if !enabled {
		return fmt.Errorf("Grok compat.claude.agents is disabled")
	}
	return nil
}

// CursorAgentsEnabled reports whether Grok loads Cursor's instruction files
// (compat.cursor.agents), resolving the documented env > TOML > default-on
// order. Hive only warns about this setting and never writes it, so any
// ambiguous or unreadable form returns an error and the caller omits the
// warning.
func CursorAgentsEnabled(home string, synthetic bool) (bool, error) {
	if value, ok := os.LookupEnv("GROK_CURSOR_AGENTS_ENABLED"); ok && !synthetic {
		enabled, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return false, fmt.Errorf("unsupported GROK_CURSOR_AGENTS_ENABLED value")
		}
		return enabled, nil
	}
	f, err := os.Open(filepath.Join(home, "config.toml"))
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read Grok config: %w", err)
	}
	defer f.Close()

	unsupported := fmt.Errorf("unsupported Grok compat.cursor TOML form")
	section := ""
	seenTable := false
	value := ""
	valueSet := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		trimmed := strings.TrimSpace(stripTOMLComment(scanner.Text()))
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			if strings.HasPrefix(trimmed, "[[") || !strings.HasSuffix(trimmed, "]") {
				if isCursorCompatTable(trimmed) {
					return false, unsupported
				}
				section = ""
				continue
			}
			section = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			if isCursorCompatTable(section) && section != "compat.cursor" {
				return false, unsupported
			}
			if section == "compat.cursor" {
				if seenTable {
					return false, fmt.Errorf("duplicate Grok compat.cursor TOML table")
				}
				seenTable = true
			}
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		key := strings.TrimSpace(parts[0])
		switch section {
		case "":
			if normalizeTOMLKey(key) == "compat" || isCursorCompatTable(key) {
				return false, unsupported
			}
		case "compat":
			if normalizeTOMLKey(key) == "cursor" || strings.Contains(strings.ToLower(trimmed), "cursor") {
				return false, unsupported
			}
		case "compat.cursor":
			if len(parts) != 2 {
				return false, unsupported
			}
			if normalizeTOMLKey(key) == "agents" && key != "agents" {
				return false, unsupported
			}
			if key != "agents" {
				continue
			}
			if valueSet {
				return false, fmt.Errorf("duplicate Grok compat.cursor.agents setting")
			}
			value = strings.TrimSpace(parts[1])
			valueSet = true
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("read Grok config: %w", err)
	}
	if !valueSet {
		return true, nil
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("unsupported Grok compat.cursor.agents value")
	}
	return enabled, nil
}

func isCursorCompatTable(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "compat") && strings.Contains(lower, "cursor")
}

func normalizeTOMLKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) >= 2 && ((key[0] == '"' && key[len(key)-1] == '"') || (key[0] == '\'' && key[len(key)-1] == '\'')) {
		return key[1 : len(key)-1]
	}
	return key
}

func isClaudeCompatTable(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "compat") && strings.Contains(lower, "claude")
}

func stripTOMLComment(line string) string {
	quoted, escaped := false, false
	for i, r := range line {
		if escaped {
			escaped = false
			continue
		}
		if quoted && r == '\\' {
			escaped = true
			continue
		}
		if r == '"' {
			quoted = !quoted
			continue
		}
		if r == '#' && !quoted {
			return line[:i]
		}
	}
	return line
}

// Resolve expands native locations for the supplied source catalogue.
func Resolve(c target.Config, sources ...string) ([]target.Target, error) {
	base, err := resolveBase(c)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(c.GrokHome, "agents")
	out := target.ExpandSkills(base, sources)
	return target.ExpandAgents(out, sources, dir, ".md", base[0]), nil
}
