package codex

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FeatureEnabled reports whether Codex's config.toml under home turns the
// named feature flag on. It recognizes `key = true|false` inside the
// [features] table and the dotted root form `features.<key> = true|false`. A
// missing file or a missing key reports false. Hive never writes this file
// (Codex's own `codex features enable` does), so any form this minimal line
// scanner cannot read with certainty, such as an inline `features` table, a
// duplicate or array-of-tables `features`, a repeated key, a non-boolean value
// or the key set to different values in both forms, returns an error and the
// caller must not treat the result as an answer.
func FeatureEnabled(home string, key string) (bool, error) {
	f, err := os.Open(filepath.Join(home, "config.toml"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read Codex config: %w", err)
	}
	defer f.Close()

	unsupported := fmt.Errorf("unsupported Codex features TOML form")
	dotted := "features." + key
	section := ""
	seenTable := false
	var inTable, atRoot bool
	var inTableSet, atRootSet bool

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		trimmed := strings.TrimSpace(stripTOMLComment(scanner.Text()))
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			if strings.HasPrefix(trimmed, "[[") || !strings.HasSuffix(trimmed, "]") {
				if strings.Contains(trimmed, "features") {
					return false, unsupported
				}
				section = "?"
				continue
			}
			section = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			switch {
			case section == "features":
				if seenTable {
					return false, fmt.Errorf("duplicate Codex features TOML table")
				}
				seenTable = true
			case section == dotted:
				return false, unsupported
			}
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		name := strings.TrimSpace(parts[0])
		switch section {
		case "":
			if name == "features" {
				return false, unsupported
			}
			if name != dotted {
				continue
			}
			if len(parts) != 2 {
				return false, unsupported
			}
			if atRootSet {
				return false, fmt.Errorf("duplicate Codex %s setting", dotted)
			}
			v, err := parseTOMLBool(parts[1])
			if err != nil {
				return false, err
			}
			atRoot, atRootSet = v, true
		case "features":
			if name != key {
				continue
			}
			if len(parts) != 2 {
				return false, unsupported
			}
			if inTableSet {
				return false, fmt.Errorf("duplicate Codex features.%s setting", key)
			}
			v, err := parseTOMLBool(parts[1])
			if err != nil {
				return false, err
			}
			inTable, inTableSet = v, true
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("read Codex config: %w", err)
	}
	switch {
	case inTableSet && atRootSet && inTable != atRoot:
		return false, fmt.Errorf("conflicting Codex features.%s settings", key)
	case inTableSet:
		return inTable, nil
	case atRootSet:
		return atRoot, nil
	}
	return false, nil
}

func parseTOMLBool(raw string) (bool, error) {
	switch strings.TrimSpace(raw) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("unsupported Codex feature value")
}

// stripTOMLComment drops a trailing `#` comment that is outside a basic
// string, as the equivalent helper in integrations/grok does.
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
