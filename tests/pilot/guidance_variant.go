package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/opencode"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

// flowBuildSkillSource is the one skill source this file ever asks a host
// resolver to expand: the read whose hash design.md's step 4 records.
const flowBuildSkillSource = "content/skills/flow-build/SKILL.md"

// guidanceVariantReport is the run.json-visible record of a --guidance-source
// installation: which checkout and arm label were used, where its shadow home
// lives, and the hashes of the two guidance files the host is meant to read
// from inside that shadow home (design.md "Pilot with the runner" (archived record heading, originally in Spanish), step 4).
// GuidanceBlockPath is Codex's AGENTS.md or Grok's Claude-compatible
// CLAUDE.md, whichever that host's own resolver actually reads.
type guidanceVariantReport struct {
	Source, Arm, ShadowHome                string
	GuidanceBlockPath, GuidanceBlockHash   string
	FlowBuildSkillPath, FlowBuildSkillHash string
	CodexAuthSymlinkPreserved              bool   `json:",omitempty"`
	CodexAuthWarning                       string `json:",omitempty"`
	// ShadowCredentialStoreRemoved records that the OpenCode shadow data
	// directory, whose database holds a copy of one imported credential, was
	// deleted at the end of the run; ShadowCredentialWarning says what survived
	// when it could not be.
	ShadowCredentialStoreRemoved bool   `json:",omitempty"`
	ShadowCredentialWarning      string `json:",omitempty"`
}

// guidanceVariant is the runtime handle a single run keeps for its shadow
// home: enough to compute the child environment and to tear down the
// authentication symlink afterward. It never touches memory.go's isolation.
type guidanceVariant struct {
	report        guidanceVariantReport
	shadowHome    string
	host          string
	realGrokHome  string
	authPath      string // empty unless host == "codex"
	shadowDataDir string // empty unless host == "opencode": shadow data dir holding the imported credential
	integration   string // OpenCode provider whose credential importCredential transfers
	// credentialImported is set once the credential import succeeded, so cleanup
	// can tell an absent data directory (a credential that went somewhere
	// unexpected) from a run that never imported.
	credentialImported bool
	// mu guards the in-flight import state below: cleanup may run from the
	// signal-handler goroutine while importOpenCodeCredential is mid-run.
	mu         sync.Mutex
	cleaned    bool
	importCmds []*exec.Cmd
	importDone chan struct{}
}

// validateGuidanceVariantFlags is main()'s pure guard for --guidance-source
// and --arm in deployed-global: both are required together, --arm must be
// A or B, and only codex, grok and opencode are supported for now. Passing neither is the
// ordinary deployed-global run and is always allowed.
func validateGuidanceVariantFlags(host, guidanceSource, arm, model string) error {
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
	if host != "codex" && host != "grok" && host != "opencode" {
		return fmt.Errorf("--guidance-source is only supported for --host codex, grok or opencode")
	}
	if host == "opencode" {
		if _, err := openCodeIntegration(model); err != nil {
			return err
		}
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
	case "opencode":
		targets, err = opencode.Resolve(c, flowBuildSkillSource)
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

// checkOpenCodeCwdOutsideHome guards the isolation of an OpenCode variant
// run. Observed on OpenCode v2.0.22 with a shadow HOME: when the working
// directory sits under the real home, OpenCode also watches and loads the real
// ~/.claude/skills, ~/.agents/skills and ~/.opencode/skill(s) (smoke run 3
// loaded flow-research from the real ~/.agents/skills); with the same shadow
// home and a working directory outside the real home only the shadow's paths
// appear. HOME, OPENCODE_TEST_HOME and the XDG variables do not change this,
// and OPENCODE_DISABLE_PROJECT_CONFIG=1 would also drop the fixture's own
// AGENTS.md, so the working directory must be outside the real home.
func checkOpenCodeCwdOutsideHome(cwd, userHome string) error {
	dir, err := target.Canonical(cwd)
	if err != nil {
		return err
	}
	home, err := target.Canonical(userHome)
	if err != nil {
		return err
	}
	if under(dir, home) {
		return fmt.Errorf("the OpenCode working directory %s is under the real home %s: OpenCode would load the real home's skills into this variant run; use a --source checkout (flows) or TMPDIR (other suites) outside %s", cwd, userHome, userHome)
	}
	return nil
}

// fixtureParentDir is the directory main() creates the fixture root in:
// under --out for the flows suite, the system temp directory otherwise. It
// lets the isolation check refuse a run before --out is created.
func fixtureParentDir(suite, output string) string {
	if suite == "flows" {
		return output
	}
	return os.TempDir()
}

// realOpenCodeDirs returns the real OpenCode config and cache directories the
// way OpenCode resolves them (verified with `opencode debug paths`, v2.0.22):
// XDG_CONFIG_HOME/XDG_CACHE_HOME when set, else ~/.config/opencode and
// ~/.cache/opencode.
func realOpenCodeDirs(userHome string, getenv func(string) string) (config, cache string, err error) {
	resolve := func(env, fallback string) (string, error) {
		base := getenv(env)
		if base == "" {
			return filepath.Join(userHome, fallback, "opencode"), nil
		}
		if !filepath.IsAbs(base) {
			return "", fmt.Errorf("%s must be an absolute path", env)
		}
		return filepath.Join(base, "opencode"), nil
	}
	if config, err = resolve("XDG_CONFIG_HOME", ".config"); err != nil {
		return "", "", err
	}
	if cache, err = resolve("XDG_CACHE_HOME", ".cache"); err != nil {
		return "", "", err
	}
	return config, cache, nil
}

// copyOpenCodeModelCatalog copies (never links, so the run cannot modify the
// real cache) the real cache's models.json, the models.dev catalog snapshot,
// into the shadow cache. The file is a public catalog, not a credential.
func copyOpenCodeModelCatalog(shadowCacheDir, realCacheDir string) error {
	data, err := os.ReadFile(filepath.Join(realCacheDir, "models.json"))
	if err != nil {
		return fmt.Errorf("real OpenCode model catalog models.json unavailable: %w", err)
	}
	if err := os.MkdirAll(shadowCacheDir, 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(shadowCacheDir, "models.json"), data, 0600)
}

// writeShadowOpenCodeConfig generates a minimal shadow opencode.json carrying
// only the real mcp.engram server's launch command, pinned to the run's
// isolated Engram store through the server's own environment. No other key of
// the real config (plugins, other MCP servers, permissions) is read or copied,
// so only the variant's installed AGENTS.md and skills shape the run.
func writeShadowOpenCodeConfig(shadowConfigDir, realConfigDir, engramDataDir string) error {
	raw, err := os.ReadFile(filepath.Join(realConfigDir, "opencode.json"))
	if err != nil {
		return fmt.Errorf("read real OpenCode opencode.json: %w", err)
	}
	var real struct {
		Mcp map[string]struct {
			Command []string `json:"command"`
		} `json:"mcp"`
	}
	if err := json.Unmarshal(raw, &real); err != nil {
		return fmt.Errorf("parse real OpenCode opencode.json: %w", err)
	}
	engram, ok := real.Mcp["engram"]
	if !ok || len(engram.Command) == 0 {
		return fmt.Errorf("real OpenCode opencode.json has no usable mcp.engram command")
	}
	shadow := map[string]any{
		"$schema": "https://opencode.ai/config.json",
		"mcp": map[string]any{"engram": map[string]any{
			"type":        "local",
			"command":     engram.Command,
			"enabled":     true,
			"environment": map[string]string{"ENGRAM_DATA_DIR": engramDataDir, "ENGRAM_CLOUD_AUTOSYNC": "0"},
		}},
	}
	out, err := json.MarshalIndent(shadow, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(shadowConfigDir, 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(shadowConfigDir, "opencode.json"), out, 0600)
}

// openCodeIntegration returns the integration (provider) ID in an OpenCode
// model such as "opencode-go/deepseek-v4.1-flash#max".
func openCodeIntegration(model string) (string, error) {
	id, _, ok := strings.Cut(model, "/")
	if !ok || id == "" {
		return "", fmt.Errorf("OpenCode guidance variants need --model provider/model, got %q", model)
	}
	return id, nil
}

// importOpenCodeCredential gives the shadow home the one credential its model
// needs. OpenCode v2 keeps credentials in its SQLite database and imports the
// legacy auth.json only in a v1-to-v2 migration, which a fresh shadow database
// never runs (observed: auth.json symlinked, credential table empty, run fails
// with provider.no-route). So `auth export <integration>` against the real
// install (read-only, default service) is piped straight into `auth import
// --standalone` under the shadow home. The credential only ever travels through
// that pipe: it is never read, logged or put in an error, and both commands'
// output is discarded.
func importOpenCodeCredential(g *guidanceVariant, parentEnv []string, integration string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	export, imp := openCodeAuthCommands(ctx, g, parentEnv, integration)
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	defer r.Close()
	export.Stdout = w
	imp.Stdin = r
	// Registered under the lock together with the start, so a concurrent cleanup
	// either sees both processes and kills them or has already forbidden the start.
	g.mu.Lock()
	if g.cleaned {
		g.mu.Unlock()
		w.Close()
		return fmt.Errorf("OpenCode credential import refused: cleanup already ran")
	}
	g.importDone = make(chan struct{})
	defer close(g.importDone)
	if err := export.Start(); err != nil {
		g.mu.Unlock()
		w.Close()
		return fmt.Errorf("start OpenCode credential export: %w", err)
	}
	w.Close()
	if err := imp.Start(); err != nil {
		g.mu.Unlock()
		killProcessGroup(export)
		export.Wait()
		return fmt.Errorf("start OpenCode credential import: %w", err)
	}
	g.importCmds = []*exec.Cmd{export, imp}
	g.mu.Unlock()
	exportErr := export.Wait()
	importErr := imp.Wait()
	if exportErr != nil {
		return fmt.Errorf("OpenCode credential export for %q failed: %v", integration, exportErr)
	}
	if importErr != nil {
		return fmt.Errorf("OpenCode credential import into the shadow home failed: %v", importErr)
	}
	return nil
}

// openCodeAuthCommands builds the export and import commands without starting
// them. Each runs outside any project directory (export in the system temp
// dir, import in the shadow home) with PWD matching it, and in its own process
// group so cleanup can kill the CLI and any standalone server it spawned.
func openCodeAuthCommands(ctx context.Context, g *guidanceVariant, parentEnv []string, integration string) (export, imp *exec.Cmd) {
	export = exec.CommandContext(ctx, "opencode", "auth", "export", integration)
	export.Dir = os.TempDir()
	export.Env = environmentInDirectory(parentEnv, export.Dir)
	imp = exec.CommandContext(ctx, "opencode", "auth", "import", "--standalone")
	imp.Dir = g.shadowHome
	imp.Env = environmentInDirectory(g.applyEnvironment(parentEnv), imp.Dir)
	for _, c := range []*exec.Cmd{export, imp} {
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	return export, imp
}

func killProcessGroup(c *exec.Cmd) {
	if c != nil && c.Process != nil {
		syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
	}
}

// importCredential is the OpenCode credential transfer as a separate step, run
// by main() after registering cleanup: every fallible setup step then precedes
// it, and a failure or signal during or after the import still removes the
// shadow data directory. It is a no-op for other hosts.
func (g *guidanceVariant) importCredential() error {
	if g == nil || g.host != "opencode" {
		return nil
	}
	if err := importOpenCodeCredential(g, os.Environ(), g.integration); err != nil {
		return err
	}
	g.credentialImported = true
	return nil
}

// setupGuidanceVariant is the top-level orchestration for design.md's
// "Pilot with the runner" (archived record heading, originally in Spanish)
// steps 1-2: it creates the run's shadow home,
// installs source's guidance into it for host, records the read-path hashes,
// and for Codex prepares its config.toml and auth symlink.
func setupGuidanceVariant(source, arm, output, host, userHome, model, engramDataDir string) (*guidanceVariant, error) {
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
	case "opencode":
		realConfig, realCache, err := realOpenCodeDirs(userHome, os.Getenv)
		if err != nil {
			return nil, err
		}
		if g.integration, err = openCodeIntegration(model); err != nil {
			return nil, err
		}
		// Mirrors the shadow layout `opencode debug paths` reports under a
		// shadowed HOME with no XDG overrides. The credential import is not
		// done here: see importCredential.
		if err := writeShadowOpenCodeConfig(filepath.Join(shadowHome, ".config", "opencode"), realConfig, engramDataDir); err != nil {
			return nil, err
		}
		if err := copyOpenCodeModelCatalog(filepath.Join(shadowHome, ".cache", "opencode"), realCache); err != nil {
			return nil, err
		}
		g.shadowDataDir = filepath.Join(shadowHome, ".local", "share", "opencode")
	}
	return g, nil
}

// applyEnvironment sets HOME to the shadow home for both hosts (step 3): for
// Codex, CODEX_HOME follows it there too, since its skills live under
// $HOME/.agents/skills; for Grok, GROK_HOME stays the real one, a declared
// limitation (design.md "Bounded risks" (archived record heading, originally in Spanish)) since Grok then reads its
// deployed agents rather than this arm's. It only ever filters and replaces
// HOME/CODEX_HOME/GROK_HOME, so ENGRAM_DATA_DIR (already set by
// prepareMemoryIsolation) and every other entry survive untouched.
func (g *guidanceVariant) applyEnvironment(env []string) []string {
	blocked := map[string]bool{"HOME": true, "CODEX_HOME": true, "GROK_HOME": true}
	if g.host == "opencode" {
		// OpenCode honors these over HOME-relative defaults (verified with
		// `opencode debug paths`), so an inherited one would point the run back
		// at the real config, data, state, or cache directories.
		for _, k := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME", "OPENCODE_CONFIG", "OPENCODE_CONFIG_DIR", "OPENCODE_CONFIG_CONTENT"} {
			blocked[k] = true
		}
	}
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
	if g == nil {
		return
	}
	g.stopImport()
	if g.shadowDataDir != "" {
		g.removeShadowDataDir()
	}
	if g.authPath == "" {
		return
	}
	stillSymlink, warning := verifyAndCleanupAuthSymlink(g.authPath)
	g.report.CodexAuthSymlinkPreserved = stillSymlink
	g.report.CodexAuthWarning = warning
}

// stopImport forbids any later import, kills an import still in flight (its
// whole process group, since the CLI may have spawned a standalone server that
// would otherwise recreate the database after removal) and waits for
// importOpenCodeCredential to return.
func (g *guidanceVariant) stopImport() {
	g.mu.Lock()
	g.cleaned = true
	for _, c := range g.importCmds {
		killProcessGroup(c)
	}
	done := g.importDone
	g.mu.Unlock()
	if done != nil {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}
}

// removeShadowDataDir deletes the whole shadow OpenCode data directory, whose
// database, WAL, log and any other file may hold the imported credential or
// data derived from it; nothing from it is kept. If anything survives, or the
// directory is absent after a successful import, a warning goes to the report
// instead of a silent leak.
func (g *guidanceVariant) removeShadowDataDir() {
	if _, err := os.Lstat(g.shadowDataDir); err != nil {
		if g.credentialImported && !g.report.ShadowCredentialStoreRemoved {
			g.report.ShadowCredentialWarning = "the credential import succeeded but the shadow OpenCode data directory is absent, so the copy went somewhere unexpected: " + g.shadowDataDir
		}
		return
	}
	os.RemoveAll(g.shadowDataDir)
	if _, err := os.Lstat(g.shadowDataDir); err == nil {
		g.report.ShadowCredentialStoreRemoved = false
		g.report.ShadowCredentialWarning = "the shadow OpenCode data directory could not be fully removed and may still hold a copy of the imported credential: " + g.shadowDataDir
		return
	}
	g.report.ShadowCredentialStoreRemoved = true
}
