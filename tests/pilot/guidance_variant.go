package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

// flowBuildSkillSource is the one skill source this file ever asks a host
// resolver to expand: the read whose hash design.md's step 4 records.
const flowBuildSkillSource = "content/skills/flow-build/SKILL.md"

// guidanceVariantReport is the run.json-visible record of a --guidance-source
// installation: which checkout and arm label were used, where its shadow home
// lives, and the hashes of the two guidance files the host is meant to read
// from inside that shadow home (design.md "Piloto con el runner", step 4).
// GuidanceBlockPath is Codex's AGENTS.md or Grok's Claude-compatible
// CLAUDE.md, whichever that host's own resolver actually reads.
type guidanceVariantReport struct {
	Source, Arm, ShadowHome                string
	GuidanceBlockPath, GuidanceBlockHash   string
	FlowBuildSkillPath, FlowBuildSkillHash string
	CodexAuthSymlinkPreserved              bool   `json:",omitempty"`
	CodexAuthWarning                       string `json:",omitempty"`
}

// guidanceVariant is the runtime handle a single run keeps for its shadow
// home: enough to compute the child environment and to tear down the
// authentication symlink afterward. It never touches memory.go's isolation.
type guidanceVariant struct {
	report       guidanceVariantReport
	shadowHome   string
	host         string
	realGrokHome string
	authPath     string // empty unless host == "codex"
}

// validateGuidanceVariantFlags is main()'s pure guard for --guidance-source
// and --arm in deployed-global: both are required together, --arm must be
// A or B, and only codex/grok are supported for now. Passing neither is the
// ordinary deployed-global run and is always allowed.
func validateGuidanceVariantFlags(host, guidanceSource, arm string) error {
	if guidanceSource == "" && arm == "" {
		return nil
	}
	if guidanceSource == "" {
		return fmt.Errorf("--arm in deployed-global requires --guidance-source")
	}
	if arm == "" {
		return fmt.Errorf("--guidance-source requires --arm A or B")
	}
	if arm != "A" && arm != "B" {
		return fmt.Errorf("--arm must be A or B")
	}
	if host != "codex" && host != "grok" {
		return fmt.Errorf("--guidance-source is only supported for --host codex or grok")
	}
	return nil
}

// shadowHostConfig builds the same synthetic target.Config tooling/management
// would build for Options{Scope:"user", Home: shadowHome} (see
// management.normalize), without installing anything. It is the "fake host"
// surface guidanceReadPaths and installGuidanceVariant both resolve against.
func shadowHostConfig(shadowHome string) (target.Config, error) {
	home, err := target.Canonical(shadowHome)
	if err != nil {
		return target.Config{}, err
	}
	c := target.Config{Scope: "user", Home: home, CodexHome: filepath.Join(home, ".codex"), ClaudeHome: filepath.Join(home, ".claude")}
	if c.CodexHome, err = target.Canonical(c.CodexHome); err != nil {
		return c, err
	}
	if c.ClaudeHome, err = target.Canonical(c.ClaudeHome); err != nil {
		return c, err
	}
	return target.ExpandHostHomes(c, true)
}

// guidanceReadPaths returns the guidance block (Codex's AGENTS.md, Grok's
// CLAUDE.md) and flow-build/SKILL.md paths this run's design.md step 4 hash
// record refers to, by asking that host's own resolver — never a guess at
// its layout. It never runs a host process or looks at process environment:
// a host name and a shadow home are enough to compute where a faithful
// reader would look inside it.
func guidanceReadPaths(host, shadowHome string) (blockPath, skillPath string, err error) {
	c, err := shadowHostConfig(shadowHome)
	if err != nil {
		return "", "", err
	}
	var targets []target.Target
	switch host {
	case "codex":
		targets, err = codex.Resolve(c, flowBuildSkillSource)
	case "grok":
		targets, err = grok.Resolve(c, flowBuildSkillSource)
	default:
		return "", "", fmt.Errorf("guidance variant unsupported for host %q", host)
	}
	if err != nil {
		return "", "", err
	}
	for _, t := range targets {
		if t.Kind == "block" {
			blockPath = t.Path
		}
		if t.Source == flowBuildSkillSource {
			skillPath = t.Path
		}
	}
	if blockPath == "" || skillPath == "" {
		return "", "", fmt.Errorf("could not resolve guidance read paths for host %q", host)
	}
	return blockPath, skillPath, nil
}

func hashOrAbsent(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return "absent"
	}
	return digest(b)
}

// scrubManagerEnv removes CODEX_HOME/GROK_HOME from this process's own
// environment for the duration of a guidance install, per design.md step 1.
// tooling/management already ignores these for a synthetic (explicit
// Options.Home) install; this is the defense-in-depth the design calls for
// on top of that. The returned func restores whatever was there before.
func scrubManagerEnv() func() {
	codexHome, hadCodex := os.LookupEnv("CODEX_HOME")
	grokHome, hadGrok := os.LookupEnv("GROK_HOME")
	os.Unsetenv("CODEX_HOME")
	os.Unsetenv("GROK_HOME")
	return func() {
		if hadCodex {
			os.Setenv("CODEX_HOME", codexHome)
		}
		if hadGrok {
			os.Setenv("GROK_HOME", grokHome)
		}
	}
}

// planDestinationsUnderHome is the R2 guard: every destination a built plan
// would touch must resolve under the shadow home. tooling/management's own
// synthetic-home resolution already guarantees this; this is the explicit,
// separately testable check design.md's step 1 asks for.
func planDestinationsUnderHome(p management.Plan, home string) error {
	for _, ch := range p.Changes {
		if !under(ch.Target.Path, home) {
			return fmt.Errorf("planned destination escapes shadow home: %s", ch.Target.Path)
		}
	}
	return nil
}

// installGuidanceVariant installs source's Hive guidance into a shadow home
// for host only (scope user), validates every planned destination stays
// inside that home, and applies it. It never touches the real user home.
func installGuidanceVariant(source, shadowHome, host string) (management.Plan, error) {
	restore := scrubManagerEnv()
	defer restore()
	opts := management.Options{Scope: "user", Home: shadowHome, Source: source, Hosts: []string{host}}
	plan, err := management.BuildPlan("install", opts)
	if err != nil {
		return plan, err
	}
	if err := planDestinationsUnderHome(plan, shadowHome); err != nil {
		return plan, err
	}
	if _, err := (management.Engine{}).Apply(plan); err != nil {
		return plan, err
	}
	return plan, nil
}

// parseTOMLBasicString accepts a TOML basic string ("...") using JSON's
// escaping rules as an approximation. This is deliberately not a general
// TOML reader (see onlyTrustAdded's comment in main.go for the same stance):
// it only ever has to round-trip the plain ASCII command/args values a
// declared MCP server entry carries.
func parseTOMLBasicString(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", false
	}
	var out string
	if json.Unmarshal([]byte(s), &out) != nil {
		return "", false
	}
	return out, true
}

func tomlQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// splitTOMLArrayElements splits a single-line TOML array's inner text on
// commas outside quotes. Multi-line arrays are a declared, out-of-scope
// limitation: the real config this reads from writes args on one line.
func splitTOMLArrayElements(s string) []string {
	var parts []string
	var cur strings.Builder
	inQuote := false
	escaped := false
	for _, r := range s {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\' && inQuote:
			cur.WriteRune(r)
			escaped = true
		case r == '"':
			inQuote = !inQuote
			cur.WriteRune(r)
		case r == ',' && !inQuote:
			parts = append(parts, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if strings.TrimSpace(cur.String()) != "" {
		parts = append(parts, cur.String())
	}
	return parts
}

func parseTOMLStringArray(s string) ([]string, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, false
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return []string{}, true
	}
	out := []string{}
	for _, part := range splitTOMLArrayElements(inner) {
		v, ok := parseTOMLBasicString(strings.TrimSpace(part))
		if !ok {
			return nil, false
		}
		out = append(out, v)
	}
	return out, true
}

// extractEngramServerDefinition reads only the "command" and "args" keys of
// the real config.toml's [mcp_servers.engram] table: the launch definition a
// shadow config needs so Codex can actually start that MCP server, before
// codexMemoryArgs's -c mcp_servers.engram.env.* overrides its environment. No
// other table or key is read or copied, and neither value is ever printed.
func extractEngramServerDefinition(doc string) (command string, args []string, err error) {
	inSection := false
	haveCommand, haveArgs := false, false
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inSection = trimmed == "[mcp_servers.engram]"
			continue
		}
		if !inSection || trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "command":
			if s, ok := parseTOMLBasicString(value); ok {
				command, haveCommand = s, true
			}
		case "args":
			if a, ok := parseTOMLStringArray(value); ok {
				args, haveArgs = a, true
			}
		}
	}
	if !haveCommand || !haveArgs {
		return "", nil, fmt.Errorf("real Codex config.toml has no usable [mcp_servers.engram] command/args")
	}
	return command, args, nil
}

// writeShadowCodexConfig generates a minimal shadow CODEX_HOME/config.toml
// carrying only the real [mcp_servers.engram] launch definition, so the
// isolated MCP server can start; codexMemoryArgs supplies its environment
// separately via -c flags. Nothing else from the real config is copied.
func writeShadowCodexConfig(shadowCodexHome, realCodexHome string) error {
	real, err := os.ReadFile(filepath.Join(realCodexHome, "config.toml"))
	if err != nil {
		return fmt.Errorf("read real Codex config.toml: %w", err)
	}
	command, args, err := extractEngramServerDefinition(string(real))
	if err != nil {
		return err
	}
	quotedArgs := make([]string, len(args))
	for i, a := range args {
		quotedArgs[i] = tomlQuote(a)
	}
	var b strings.Builder
	b.WriteString("[mcp_servers.engram]\n")
	b.WriteString("command = " + tomlQuote(command) + "\n")
	b.WriteString("args = [" + strings.Join(quotedArgs, ", ") + "]\n")
	return os.WriteFile(filepath.Join(shadowCodexHome, "config.toml"), []byte(b.String()), 0600)
}

// linkCodexAuth symlinks (never copies) the real CODEX_HOME's auth.json into
// the shadow CODEX_HOME, so Codex authenticates without a credential copy
// living in run evidence.
func linkCodexAuth(shadowCodexHome, realCodexHome string) error {
	real := filepath.Join(realCodexHome, "auth.json")
	if _, err := os.Lstat(real); err != nil {
		return fmt.Errorf("real Codex auth.json unavailable: %w", err)
	}
	shadow := filepath.Join(shadowCodexHome, "auth.json")
	if err := os.Remove(shadow); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Symlink(real, shadow)
}

// verifyAndCleanupAuthSymlink is the end-of-run auth guard: it never reads
// path's content. If Codex renewed the token by replacing the symlink with a
// regular file, that file is deleted unread and a warning is returned; a
// still-intact symlink is simply removed. An already-absent path is neither
// a warning nor a still-symlink.
func verifyAndCleanupAuthSymlink(path string) (stillSymlink bool, warning string) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, ""
	}
	stillSymlink = info.Mode()&os.ModeSymlink != 0
	if !stillSymlink {
		warning = "Codex replaced the shadow CODEX_HOME auth.json symlink with a regular file during this run; it was deleted without being read. If Codex authentication now fails, run `codex login` again to renew the real credential."
	}
	os.Remove(path)
	return stillSymlink, warning
}

// setupGuidanceVariant is the top-level orchestration for design.md's
// "Piloto con el runner" steps 1-2: it creates the run's shadow home,
// installs source's guidance into it for host, records the read-path hashes,
// and for Codex prepares its config.toml and auth symlink.
func setupGuidanceVariant(source, arm, output, host, userHome string) (*guidanceVariant, error) {
	shadowHome := filepath.Join(output, "shadow-home")
	if err := os.MkdirAll(shadowHome, 0700); err != nil {
		return nil, err
	}
	if _, err := installGuidanceVariant(source, shadowHome, host); err != nil {
		return nil, err
	}
	blockPath, skillPath, err := guidanceReadPaths(host, shadowHome)
	if err != nil {
		return nil, err
	}
	g := &guidanceVariant{
		shadowHome: shadowHome,
		host:       host,
		report: guidanceVariantReport{
			Source: source, Arm: arm, ShadowHome: shadowHome,
			GuidanceBlockPath: blockPath, GuidanceBlockHash: hashOrAbsent(blockPath),
			FlowBuildSkillPath: skillPath, FlowBuildSkillHash: hashOrAbsent(skillPath),
		},
	}
	switch host {
	case "codex":
		shadowCodexHome := filepath.Join(shadowHome, ".codex")
		realCodexHome := envRoot("CODEX_HOME", filepath.Join(userHome, ".codex"))
		if err := writeShadowCodexConfig(shadowCodexHome, realCodexHome); err != nil {
			return nil, err
		}
		if err := linkCodexAuth(shadowCodexHome, realCodexHome); err != nil {
			return nil, err
		}
		g.authPath = filepath.Join(shadowCodexHome, "auth.json")
	case "grok":
		g.realGrokHome = envRoot("GROK_HOME", filepath.Join(userHome, ".grok"))
	}
	return g, nil
}

// applyEnvironment sets HOME to the shadow home for both hosts (step 3): for
// Codex, CODEX_HOME follows it there too, since its skills live under
// $HOME/.agents/skills; for Grok, GROK_HOME stays the real one, a declared
// limitation (design.md "Riesgos acotados") since Grok then reads its
// deployed agents rather than this arm's. It only ever filters and replaces
// HOME/CODEX_HOME/GROK_HOME, so ENGRAM_DATA_DIR (already set by
// prepareMemoryIsolation) and every other entry survive untouched.
func (g *guidanceVariant) applyEnvironment(env []string) []string {
	blocked := map[string]bool{"HOME": true, "CODEX_HOME": true, "GROK_HOME": true}
	out := make([]string, 0, len(env)+2)
	for _, e := range env {
		key, _, _ := strings.Cut(e, "=")
		if !blocked[key] {
			out = append(out, e)
		}
	}
	out = append(out, "HOME="+g.shadowHome)
	switch g.host {
	case "codex":
		out = append(out, "CODEX_HOME="+filepath.Join(g.shadowHome, ".codex"))
	case "grok":
		out = append(out, "GROK_HOME="+g.realGrokHome)
	}
	return out
}

// cleanup removes the Codex auth symlink (step 6's "remove symlinks and
// shadow auth"), guarding against a renewed token left as a plain file, and
// records the outcome onto the report the caller re-embeds into run.json.
func (g *guidanceVariant) cleanup() {
	if g == nil || g.authPath == "" {
		return
	}
	stillSymlink, warning := verifyAndCleanupAuthSymlink(g.authPath)
	g.report.CodexAuthSymlinkPreserved = stillSymlink
	g.report.CodexAuthWarning = warning
}
