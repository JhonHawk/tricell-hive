package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

func TestValidateGuidanceVariantFlags(t *testing.T) {
	for _, c := range []struct {
		name, host, source, arm, model string
		wantErr                        bool
	}{
		{"neither flag is the ordinary run", "codex", "", "", "gpt-6", false},
		{"arm without source fails", "codex", "", "A", "gpt-6", true},
		{"source without arm fails", "codex", "/src", "", "gpt-6", true},
		{"bad arm value fails", "codex", "/src", "C", "gpt-6", true},
		{"unsupported host fails", "claude", "/src", "A", "m", true},
		{"codex arm A accepted", "codex", "/src", "A", "gpt-6", false},
		{"grok arm B accepted", "grok", "/src", "B", "grok-5", false},
		{"opencode arm A accepted", "opencode", "/src", "A", "opencode-go/m", false},
		{"opencode with variant suffix accepted", "opencode", "/src", "B", "opencode-go/m#max", false},
		{"opencode model without provider prefix fails", "opencode", "/src", "A", "deepseek-v4.1-flash", true},
		{"opencode empty provider fails", "opencode", "/src", "A", "/m", true},
		{"opencode empty model fails", "opencode", "/src", "A", "", true},
		{"pi still unsupported", "pi", "/src", "A", "p/m", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := validateGuidanceVariantFlags(c.host, c.source, c.arm, c.model)
			if (err != nil) != c.wantErr {
				t.Fatalf("host=%s source=%q arm=%q: err=%v wantErr=%v", c.host, c.source, c.arm, err, c.wantErr)
			}
		})
	}
}

func TestGuidanceReadPathsAreInsideShadowHomeForEachHost(t *testing.T) {
	shadow, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"codex", "grok", "opencode"} {
		agents, skill, err := guidanceReadPaths(host, shadow)
		if err != nil {
			t.Fatal(err)
		}
		if !under(agents, shadow) || !under(skill, shadow) {
			t.Fatalf("%s read paths escape the shadow home: %s %s", host, agents, skill)
		}
		if !strings.HasSuffix(skill, filepath.Join("skills", "flow-build", "SKILL.md")) {
			t.Fatalf("%s skill path is not flow-build's: %s", host, skill)
		}
	}
	if _, _, err := guidanceReadPaths("claude", shadow); err == nil {
		t.Fatal("expected an unsupported host to fail clearly")
	}
}

func newFixtureCheckout(t *testing.T, globalBody string) string {
	t.Helper()
	dir := t.TempDir()
	write := func(p, s string) {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(s), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(management.GlobalSource, globalBody)
	write("content/skills/flow-build/SKILL.md", "---\nname: flow-build\ndescription: Implement authorized work.\n---\nPreserve evidence.\n")
	canonical, err := target.Canonical(dir)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

func TestGuidanceVariantInstallsDifferentContentPerArmIntoDifferentShadowHomes(t *testing.T) {
	dirA := newFixtureCheckout(t, "# Rules A\nArm A body.\n")
	dirB := newFixtureCheckout(t, "# Rules B\nArm B body.\n")
	baseA, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	baseB, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	shadowA := filepath.Join(baseA, "shadow")
	shadowB := filepath.Join(baseB, "shadow")
	for _, home := range []string{shadowA, shadowB} {
		if err := os.MkdirAll(home, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := installGuidanceVariant(dirA, shadowA, "codex"); err != nil {
		t.Fatal(err)
	}
	if _, err := installGuidanceVariant(dirB, shadowB, "codex"); err != nil {
		t.Fatal(err)
	}
	agentsA, skillA, err := guidanceReadPaths("codex", shadowA)
	if err != nil {
		t.Fatal(err)
	}
	agentsB, skillB, err := guidanceReadPaths("codex", shadowB)
	if err != nil {
		t.Fatal(err)
	}
	if !under(agentsA, shadowA) || !under(agentsB, shadowB) {
		t.Fatal("installed guidance escaped its own shadow home")
	}
	hashA, hashB := hashOrAbsent(agentsA), hashOrAbsent(agentsB)
	if hashA == "absent" || hashB == "absent" {
		t.Fatalf("guidance was not installed: %s %s", hashA, hashB)
	}
	if hashA == hashB {
		t.Fatal("both arms installed identical guidance")
	}
	if hashOrAbsent(skillA) == "absent" || hashOrAbsent(skillB) == "absent" {
		t.Fatal("flow-build skill was not installed into the shadow home")
	}
}

func TestGuidanceVariantReportSerializesHashesIntoRunJSON(t *testing.T) {
	r := result{GuidanceVariant: &guidanceVariantReport{Source: "/src", Arm: "A", GuidanceBlockHash: "deadbeef", FlowBuildSkillHash: "cafef00d"}}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"GuidanceBlockHash":"deadbeef"`, `"FlowBuildSkillHash":"cafef00d"`, `"Arm":"A"`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("missing %s in %s", want, b)
		}
	}
}

func TestPlanDestinationsUnderHomeGuardsEscape(t *testing.T) {
	home := filepath.Join(string(os.PathSeparator), "shadow", "home")
	good := management.Plan{Changes: []management.Change{{Target: target.Target{Path: filepath.Join(home, ".codex", "AGENTS.md")}}}}
	if err := planDestinationsUnderHome(good, home); err != nil {
		t.Fatal(err)
	}
	bad := management.Plan{Changes: []management.Change{{Target: target.Target{Path: filepath.Join(string(os.PathSeparator), "etc", "passwd")}}}}
	if err := planDestinationsUnderHome(bad, home); err == nil {
		t.Fatal("expected an escaping destination to be rejected")
	}
}

func TestScrubManagerEnvRemovesAndRestoresCodexAndGrokHome(t *testing.T) {
	t.Setenv("CODEX_HOME", "/real/codex")
	t.Setenv("GROK_HOME", "/real/grok")
	restore := scrubManagerEnv()
	if v := os.Getenv("CODEX_HOME"); v != "" {
		t.Fatalf("CODEX_HOME not scrubbed: %s", v)
	}
	if v := os.Getenv("GROK_HOME"); v != "" {
		t.Fatalf("GROK_HOME not scrubbed: %s", v)
	}
	restore()
	if v := os.Getenv("CODEX_HOME"); v != "/real/codex" {
		t.Fatalf("CODEX_HOME not restored: %s", v)
	}
	if v := os.Getenv("GROK_HOME"); v != "/real/grok" {
		t.Fatalf("GROK_HOME not restored: %s", v)
	}
}

func TestScrubManagerEnvLeavesUnsetVariablesUnset(t *testing.T) {
	t.Setenv("CODEX_HOME", "")
	os.Unsetenv("CODEX_HOME")
	t.Setenv("GROK_HOME", "")
	os.Unsetenv("GROK_HOME")
	restore := scrubManagerEnv()
	restore()
	if _, ok := os.LookupEnv("CODEX_HOME"); ok {
		t.Fatal("CODEX_HOME was invented by restore")
	}
	if _, ok := os.LookupEnv("GROK_HOME"); ok {
		t.Fatal("GROK_HOME was invented by restore")
	}
}

func TestGuidanceVariantEnvironmentPreservesEngramIsolationAndSetsHostHomes(t *testing.T) {
	for _, tc := range []struct{ host string }{{"codex"}, {"grok"}, {"opencode"}} {
		g := &guidanceVariant{shadowHome: "/shadow", host: tc.host, realGrokHome: "/real/grok"}
		in := []string{"HOME=/old-home", "CODEX_HOME=/old-codex", "GROK_HOME=/old-grok", "ENGRAM_DATA_DIR=/isolated/store", "PATH=/bin"}
		out := g.applyEnvironment(in)
		get := func(key string) (string, bool) {
			for _, e := range out {
				if k, v, ok := strings.Cut(e, "="); ok && k == key {
					return v, true
				}
			}
			return "", false
		}
		if v, ok := get("HOME"); !ok || v != "/shadow" {
			t.Fatalf("%s: HOME not shadowed: %v %v", tc.host, v, ok)
		}
		if v, ok := get("ENGRAM_DATA_DIR"); !ok || v != "/isolated/store" {
			t.Fatalf("%s: ENGRAM_DATA_DIR lost precedence: %v %v", tc.host, v, ok)
		}
		if v, ok := get("PATH"); !ok || v != "/bin" {
			t.Fatalf("%s: unrelated env entry lost", tc.host)
		}
		count := 0
		for _, e := range out {
			if strings.HasPrefix(e, "HOME=") {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("%s: duplicate HOME entries: %d", tc.host, count)
		}
		switch tc.host {
		case "codex":
			if v, ok := get("CODEX_HOME"); !ok || v != filepath.Join("/shadow", ".codex") {
				t.Fatalf("CODEX_HOME not shadowed: %v", v)
			}
			if _, ok := get("GROK_HOME"); ok {
				t.Fatal("a codex run must not carry GROK_HOME")
			}
		case "grok":
			if v, ok := get("GROK_HOME"); !ok || v != "/real/grok" {
				t.Fatalf("GROK_HOME not set to the real, unshadowed path: %v", v)
			}
			if _, ok := get("CODEX_HOME"); ok {
				t.Fatal("a grok run must not carry CODEX_HOME")
			}
		case "opencode":
			for _, key := range []string{"CODEX_HOME", "GROK_HOME"} {
				if _, ok := get(key); ok {
					t.Fatalf("an opencode run must not carry %s", key)
				}
			}
		}
	}
}

func TestWriteShadowCodexConfigPreservesEngramServerDefinitionOnly(t *testing.T) {
	real := t.TempDir()
	realConfig := "model = \"gpt-6\"\n\n[mcp_servers.other]\ncommand = \"other-bin\"\n\n[mcp_servers.engram]\ncommand = \"npx\"\nargs = [\"-y\", \"@tricell/engram-mcp\"]\n\n[notice]\nflag = true\n"
	if err := os.WriteFile(filepath.Join(real, "config.toml"), []byte(realConfig), 0600); err != nil {
		t.Fatal(err)
	}
	shadow := t.TempDir()
	if err := writeShadowCodexConfig(shadow, real); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(shadow, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{"[mcp_servers.engram]", `command = "npx"`, `"-y"`, `"@tricell/engram-mcp"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in shadow config: %s", want, got)
		}
	}
	for _, unwanted := range []string{"other-bin", "mcp_servers.other", "notice", "gpt-6"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("shadow config copied unrelated real configuration (%q): %s", unwanted, got)
		}
	}
}

func TestWriteShadowCodexConfigFailsClearlyWithoutEngramServer(t *testing.T) {
	real := t.TempDir()
	if err := os.WriteFile(filepath.Join(real, "config.toml"), []byte("model = \"gpt-6\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeShadowCodexConfig(t.TempDir(), real); err == nil {
		t.Fatal("expected a missing Engram server definition to fail clearly")
	}
}

func TestLinkCodexAuthSymlinksRatherThanCopies(t *testing.T) {
	realHome := t.TempDir()
	if err := os.WriteFile(filepath.Join(realHome, "auth.json"), []byte(`{"token":"secret-value"}`), 0600); err != nil {
		t.Fatal(err)
	}
	shadowHome := t.TempDir()
	if err := linkCodexAuth(shadowHome, realHome); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(shadowHome, "auth.json")
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected a symlink, not a copy")
	}
	dest, err := os.Readlink(link)
	if err != nil || dest != filepath.Join(realHome, "auth.json") {
		t.Fatalf("wrong symlink target: %s", dest)
	}
}

func TestLinkCodexAuthFailsClearlyWhenRealAuthMissing(t *testing.T) {
	if err := linkCodexAuth(t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("expected a clear error when the real auth.json is absent")
	}
}

func TestVerifyAndCleanupAuthSymlinkStillSymlinkIsRemovedWithoutWarning(t *testing.T) {
	real := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(real, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	shadow := filepath.Join(t.TempDir(), "auth.json")
	if err := os.Symlink(real, shadow); err != nil {
		t.Fatal(err)
	}
	stillSymlink, warning := verifyAndCleanupAuthSymlink(shadow)
	if !stillSymlink || warning != "" {
		t.Fatalf("unexpected result: stillSymlink=%v warning=%q", stillSymlink, warning)
	}
	if _, err := os.Lstat(shadow); !os.IsNotExist(err) {
		t.Fatal("symlink was not removed")
	}
}

func TestVerifyAndCleanupAuthSymlinkRegularFileIsDeletedWithWarningAndNeverRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, []byte(`{"token":"renewed-secret"}`), 0600); err != nil {
		t.Fatal(err)
	}
	stillSymlink, warning := verifyAndCleanupAuthSymlink(path)
	if stillSymlink || warning == "" {
		t.Fatalf("expected a replacement warning, got stillSymlink=%v warning=%q", stillSymlink, warning)
	}
	if strings.Contains(warning, "renewed-secret") {
		t.Fatal("warning must never carry the file's content")
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatal("the replaced auth file was not deleted")
	}
}

func TestVerifyAndCleanupAuthSymlinkAbsentIsQuiet(t *testing.T) {
	stillSymlink, warning := verifyAndCleanupAuthSymlink(filepath.Join(t.TempDir(), "auth.json"))
	if stillSymlink || warning != "" {
		t.Fatalf("an already-absent path must be quiet: stillSymlink=%v warning=%q", stillSymlink, warning)
	}
}

func TestOpenCodeEnvironmentDropsXDGAndConfigOverrides(t *testing.T) {
	g := &guidanceVariant{shadowHome: "/shadow", host: "opencode"}
	in := []string{"HOME=/old", "XDG_CONFIG_HOME=/real/cfg", "XDG_DATA_HOME=/real/data", "XDG_STATE_HOME=/real/state", "XDG_CACHE_HOME=/real/cache", "XDG_DATA_DIRS=/usr/share", "OPENCODE_CONFIG=/real/o.json", "OPENCODE_CONFIG_DIR=/real/dir", "OPENCODE_CONFIG_CONTENT={}", "OPENCODE_EXPERIMENTAL=1", "ENGRAM_DATA_DIR=/iso"}
	out := g.applyEnvironment(in)
	have := map[string]string{}
	for _, e := range out {
		k, v, _ := strings.Cut(e, "=")
		have[k] = v
	}
	for _, gone := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME", "OPENCODE_CONFIG", "OPENCODE_CONFIG_DIR", "OPENCODE_CONFIG_CONTENT"} {
		if _, ok := have[gone]; ok {
			t.Fatalf("%s would leak the real OpenCode directories into the shadow run", gone)
		}
	}
	for k, want := range map[string]string{"HOME": "/shadow", "XDG_DATA_DIRS": "/usr/share", "OPENCODE_EXPERIMENTAL": "1", "ENGRAM_DATA_DIR": "/iso"} {
		if have[k] != want {
			t.Fatalf("%s = %q, want %q", k, have[k], want)
		}
	}
}

func TestRealOpenCodeDirsHonorXDGThenHomeDefaults(t *testing.T) {
	cfg, cache, err := realOpenCodeDirs("/home/u", func(string) string { return "" })
	if err != nil || cfg != "/home/u/.config/opencode" || cache != "/home/u/.cache/opencode" {
		t.Fatalf("defaults: %s %s %v", cfg, cache, err)
	}
	env := map[string]string{"XDG_CONFIG_HOME": "/x/c", "XDG_CACHE_HOME": "/x/k"}
	cfg, cache, err = realOpenCodeDirs("/home/u", func(k string) string { return env[k] })
	if err != nil || cfg != "/x/c/opencode" || cache != "/x/k/opencode" {
		t.Fatalf("xdg: %s %s %v", cfg, cache, err)
	}
	for _, bad := range []string{"XDG_CONFIG_HOME", "XDG_CACHE_HOME"} {
		if _, _, err := realOpenCodeDirs("/home/u", func(k string) string {
			if k == bad {
				return "relative"
			}
			return ""
		}); err == nil {
			t.Fatalf("a relative %s must be rejected", bad)
		}
	}
}

func TestWriteShadowOpenCodeConfigCarriesOnlyIsolatedEngramServer(t *testing.T) {
	real := t.TempDir()
	realConfig := `{"autoupdate":"notify","mcp":{"linear":{"type":"remote","url":"https://example.invalid/mcp"},"engram":{"command":["/opt/bin/engram","mcp","--tools=agent"],"enabled":true,"type":"local"}},"permission":{"bash":{"*":"allow"}},"plugin":["real-plugin"]}`
	if err := os.WriteFile(filepath.Join(real, "opencode.json"), []byte(realConfig), 0600); err != nil {
		t.Fatal(err)
	}
	shadow := t.TempDir()
	if err := writeShadowOpenCodeConfig(shadow, real, "/out/engram-data"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(shadow, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Mcp map[string]struct {
			Type        string
			Command     []string
			Enabled     bool
			Environment map[string]string
		}
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatalf("shadow config is not JSON: %v", err)
	}
	e, ok := cfg.Mcp["engram"]
	if len(cfg.Mcp) != 1 || !ok || e.Type != "local" || !e.Enabled || strings.Join(e.Command, " ") != "/opt/bin/engram mcp --tools=agent" {
		t.Fatalf("unexpected engram server: %+v", cfg.Mcp)
	}
	if e.Environment["ENGRAM_DATA_DIR"] != "/out/engram-data" || e.Environment["ENGRAM_CLOUD_AUTOSYNC"] != "0" {
		t.Fatalf("engram server is not pinned to the isolated store: %v", e.Environment)
	}
	for _, unwanted := range []string{"linear", "real-plugin", "permission", "autoupdate", "example.invalid"} {
		if strings.Contains(string(b), unwanted) {
			t.Fatalf("shadow config copied unrelated real configuration (%q)", unwanted)
		}
	}
}

func TestWriteShadowOpenCodeConfigFailsClearlyWithoutEngramServer(t *testing.T) {
	real := t.TempDir()
	if err := os.WriteFile(filepath.Join(real, "opencode.json"), []byte(`{"mcp":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeShadowOpenCodeConfig(t.TempDir(), real, "/out/engram-data"); err == nil {
		t.Fatal("expected a missing Engram server definition to fail clearly")
	}
}

// fakeOpenCodeOnPath installs a fake `opencode` that records how it was called:
// `auth export <id>` prints a marker document, and `auth import` stores stdin
// plus the HOME and flags it ran with, so tests never touch a real credential.
func fakeOpenCodeOnPath(t *testing.T) string {
	t.Helper()
	bin := t.TempDir()
	script := `#!/bin/sh
if [ "$1" = "auth" ] && [ "$2" = "export" ]; then
  echo "export-home=$HOME args=$*" >> "$FAKE_OPENCODE_LOG"
  printf '{"marker":"%s"}' "$3"
  exit 0
fi
if [ "$1" = "auth" ] && [ "$2" = "import" ]; then
  echo "import-home=$HOME args=$*" >> "$FAKE_OPENCODE_LOG"
  mkdir -p "$HOME/.local/share/opencode"
  cat > "$HOME/.local/share/opencode/opencode.db"
  exit 0
fi
exit 1
`
	if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(t.TempDir(), "calls.log")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_OPENCODE_LOG", logPath)
	return logPath
}

func TestImportOpenCodeCredentialPipesExportIntoShadowImport(t *testing.T) {
	logPath := fakeOpenCodeOnPath(t)
	shadow := t.TempDir()
	g := &guidanceVariant{shadowHome: shadow, host: "opencode"}
	if err := importOpenCodeCredential(g, os.Environ(), "opencode-go"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(shadow, ".local", "share", "opencode", "opencode.db"))
	if err != nil || string(b) != `{"marker":"opencode-go"}` {
		t.Fatalf("export did not reach the shadow import: %q %v", b, err)
	}
	calls, _ := os.ReadFile(logPath)
	got := string(calls)
	if !strings.Contains(got, "export-home="+os.Getenv("HOME")+" args=auth export opencode-go") {
		t.Fatalf("export must run against the real home for only that integration: %s", got)
	}
	if !strings.Contains(got, "import-home="+shadow+" args=auth import --standalone") {
		t.Fatalf("import must run standalone under the shadow home: %s", got)
	}
}

func TestImportOpenCodeCredentialFailsClearlyWhenExportFails(t *testing.T) {
	bin := t.TempDir()
	os.WriteFile(filepath.Join(bin, "opencode"), []byte("#!/bin/sh\necho SECRET-LEAK >&2\nexit 3\n"), 0700)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	g := &guidanceVariant{shadowHome: t.TempDir(), host: "opencode"}
	err := importOpenCodeCredential(g, os.Environ(), "opencode-go")
	if err == nil || strings.Contains(err.Error(), "SECRET-LEAK") {
		t.Fatalf("expected a clear error that carries no command output: %v", err)
	}
}

func TestOpenCodeIntegrationFromModel(t *testing.T) {
	for model, want := range map[string]string{"opencode-go/deepseek-v4.1-flash#max": "opencode-go", "openai/gpt-6": "openai"} {
		if got, err := openCodeIntegration(model); err != nil || got != want {
			t.Fatalf("%s: %q %v", model, got, err)
		}
	}
	for _, bad := range []string{"", "no-slash", "/x"} {
		if _, err := openCodeIntegration(bad); err == nil {
			t.Fatalf("%q must be rejected", bad)
		}
	}
}

func TestSetupGuidanceVariantOpenCodeInstallsGuidanceConfigCatalogThenImportsCredentialSeparately(t *testing.T) {
	logPath := fakeOpenCodeOnPath(t)
	userHome := t.TempDir()
	realCfg := filepath.Join(userHome, ".config", "opencode")
	for _, d := range []string{realCfg, filepath.Join(userHome, ".cache", "opencode")} {
		if err := os.MkdirAll(d, 0700); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(realCfg, "opencode.json"), []byte(`{"mcp":{"engram":{"command":["engram","mcp"],"type":"local"}}}`), 0600)
	os.WriteFile(filepath.Join(userHome, ".cache", "opencode", "models.json"), []byte(`{}`), 0600)
	for _, k := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME"} {
		t.Setenv(k, "")
	}
	out, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	g, err := setupGuidanceVariant(newFixtureCheckout(t, "# Rules A\n"), "A", out, "opencode", userHome, "opencode-go/deepseek-v4.1-flash#max", filepath.Join(out, "engram-data"))
	if err != nil {
		t.Fatal(err)
	}
	shadow := filepath.Join(out, "shadow-home")
	wantBlock := filepath.Join(shadow, ".config", "opencode", "AGENTS.md")
	if g.report.GuidanceBlockPath != wantBlock || g.report.GuidanceBlockHash == "absent" || g.report.FlowBuildSkillHash == "absent" {
		t.Fatalf("report does not record the installed guidance: %+v", g.report)
	}
	for _, f := range []string{filepath.Join(".config", "opencode", "opencode.json"), filepath.Join(".cache", "opencode", "models.json")} {
		if _, err := os.Stat(filepath.Join(shadow, f)); err != nil {
			t.Fatal(err)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(shadow, ".config", "opencode", "opencode.json")); !strings.Contains(string(b), filepath.Join(out, "engram-data")) {
		t.Fatal("the shadow config must carry the memory isolation's data directory")
	}
	if calls, _ := os.ReadFile(logPath); len(calls) != 0 {
		t.Fatalf("setup must not import the credential itself, so every fallible step can precede it: %s", calls)
	}
	if err := g.importCredential(); err != nil {
		t.Fatal(err)
	}
	if calls, _ := os.ReadFile(logPath); !strings.Contains(string(calls), "args=auth export opencode-go") {
		t.Fatalf("credential for the model's integration was not transferred: %s", calls)
	}
	data := filepath.Join(shadow, ".local", "share", "opencode")
	if g.shadowDataDir != data {
		t.Fatalf("shadow data dir not tracked: %q", g.shadowDataDir)
	}
	g.cleanup()
	if _, err := os.Lstat(data); !os.IsNotExist(err) {
		t.Fatal("credential copy left behind")
	}
}

func TestCopyOpenCodeModelCatalogCopiesRatherThanLinks(t *testing.T) {
	realCache := t.TempDir()
	if err := os.WriteFile(filepath.Join(realCache, "models.json"), []byte(`{"catalog":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	shadowCache := filepath.Join(t.TempDir(), ".cache", "opencode")
	if err := copyOpenCodeModelCatalog(shadowCache, realCache); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(shadowCache, "models.json")
	info, err := os.Lstat(dst)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("expected a regular copy: %v", err)
	}
	if b, _ := os.ReadFile(dst); string(b) != `{"catalog":1}` {
		t.Fatalf("copy differs: %s", b)
	}
	if err := os.WriteFile(dst, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(realCache, "models.json")); string(b) != `{"catalog":1}` {
		t.Fatal("writing the copy modified the real cache")
	}
	if err := copyOpenCodeModelCatalog(t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("expected a clear error when the real models.json is missing")
	}
}

func TestCheckOpenCodeCwdOutsideHome(t *testing.T) {
	home := t.TempDir()
	if err := checkOpenCodeCwdOutsideHome(filepath.Join(home, "runs", "r1", "fixture"), home); err == nil {
		t.Fatal("a working directory under the real home must be rejected: OpenCode would load the real home's skills")
	}
	if err := checkOpenCodeCwdOutsideHome(home, home); err == nil {
		t.Fatal("the real home itself must be rejected")
	}
	if err := checkOpenCodeCwdOutsideHome(filepath.Join(t.TempDir(), "not-yet"), home); err != nil {
		t.Fatalf("a working directory outside the real home is valid: %v", err)
	}
}

func TestOpenCodeFixtureParent(t *testing.T) {
	if got := fixtureParentDir("flows", "/out/run"); got != "/out/run" {
		t.Fatalf("flows fixtures live under --out: %s", got)
	}
	if got := fixtureParentDir("workspace-conventions", "/out/run"); got != os.TempDir() {
		t.Fatalf("other suites use the system temp dir: %s", got)
	}
}

func TestCleanupRemovesWholeShadowDataDirIncludingLog(t *testing.T) {
	shadow := t.TempDir()
	data := filepath.Join(shadow, ".local", "share", "opencode")
	os.MkdirAll(filepath.Join(data, "log"), 0700)
	for _, f := range []string{"opencode.db", "opencode.db-wal", "credentials-extra.json", filepath.Join("log", "opencode.log")} {
		os.WriteFile(filepath.Join(data, f), []byte("x"), 0600)
	}
	g := &guidanceVariant{shadowHome: shadow, host: "opencode", shadowDataDir: data}
	g.cleanup()
	if _, err := os.Stat(data); !os.IsNotExist(err) {
		t.Fatal("the whole shadow data directory must be removed")
	}
	entries, _ := os.ReadDir(shadow)
	for _, e := range entries {
		if e.Name() != ".local" {
			t.Fatalf("nothing from the data directory may be retained, found %s", e.Name())
		}
	}
	if !g.report.ShadowCredentialStoreRemoved || g.report.ShadowCredentialWarning != "" {
		t.Fatalf("unexpected outcome: %+v", g.report)
	}
	g.cleanup() // idempotent
	if !g.report.ShadowCredentialStoreRemoved {
		t.Fatal("a second cleanup must not undo the recorded outcome")
	}
}

func TestCleanupWarnsWhenImportSucceededButDataDirIsAbsent(t *testing.T) {
	g := &guidanceVariant{shadowHome: t.TempDir(), host: "opencode", shadowDataDir: filepath.Join(t.TempDir(), "gone"), credentialImported: true}
	g.cleanup()
	if g.report.ShadowCredentialWarning == "" || g.report.ShadowCredentialStoreRemoved {
		t.Fatalf("a credential that went somewhere unexpected must be reported: %+v", g.report)
	}
	quiet := &guidanceVariant{shadowHome: t.TempDir(), host: "opencode", shadowDataDir: filepath.Join(t.TempDir(), "gone")}
	quiet.cleanup()
	if quiet.report.ShadowCredentialWarning != "" {
		t.Fatal("no import happened, so an absent directory is not a warning")
	}
}

func TestCleanupWarnsWhenShadowDataSurvives(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions, so the removal cannot be made to fail")
	}
	shadow := t.TempDir()
	data := filepath.Join(shadow, ".local", "share", "opencode")
	os.MkdirAll(data, 0700)
	os.WriteFile(filepath.Join(data, "opencode.db"), []byte("x"), 0600)
	// A read-only parent makes the removal fail on every platform the tests run on.
	parent := filepath.Dir(data)
	os.Chmod(parent, 0500)
	defer os.Chmod(parent, 0700)
	g := &guidanceVariant{shadowHome: shadow, host: "opencode", shadowDataDir: data}
	g.cleanup()
	if g.report.ShadowCredentialStoreRemoved || g.report.ShadowCredentialWarning == "" {
		t.Fatalf("a surviving credential store must be reported: %+v", g.report)
	}
}

func TestImportOpenCodeCredentialRunsOutsideProjectDirectories(t *testing.T) {
	bin := t.TempDir()
	script := "#!/bin/sh\npwd >> \"$FAKE_OPENCODE_LOG\"\ncat > /dev/null\nprintf '{}'\n"
	os.WriteFile(filepath.Join(bin, "opencode"), []byte(script), 0700)
	logPath := filepath.Join(t.TempDir(), "pwd.log")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_OPENCODE_LOG", logPath)
	shadow, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	g := &guidanceVariant{shadowHome: shadow, host: "opencode"}
	if err := importOpenCodeCredential(g, os.Environ(), "opencode-go"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(logPath)
	dirs := strings.Fields(string(b))
	tmp, _ := target.Canonical(os.TempDir())
	if len(dirs) != 2 {
		t.Fatalf("expected export and import working directories: %v", dirs)
	}
	// The two processes run concurrently, so their log order is not defined.
	seen := map[string]bool{}
	for _, d := range dirs {
		c, _ := target.Canonical(d)
		seen[c] = true
	}
	if !seen[tmp] || !seen[shadow] {
		t.Fatalf("export must run in the system temp dir and import in the shadow home, ran in %v", dirs)
	}
}

func TestOpenCodeAuthCommandsSetDirAndMatchingPWD(t *testing.T) {
	shadow := t.TempDir()
	g := &guidanceVariant{shadowHome: shadow, host: "opencode"}
	export, imp := openCodeAuthCommands(context.Background(), g, []string{"HOME=/real", "PWD=/runner/cwd", "OLDPWD=/x"}, "opencode-go")
	for name, c := range map[string]*exec.Cmd{"export": export, "import": imp} {
		if c.Dir == "" {
			t.Fatalf("%s has no working directory", name)
		}
		pwd := 0
		for _, e := range c.Env {
			if strings.HasPrefix(e, "PWD=") {
				pwd++
				if e != "PWD="+c.Dir {
					t.Fatalf("%s: PWD %q disagrees with Dir %q", name, e, c.Dir)
				}
			}
			if strings.HasPrefix(e, "OLDPWD=") {
				t.Fatalf("%s: stale OLDPWD kept", name)
			}
		}
		if pwd != 1 {
			t.Fatalf("%s: expected exactly one PWD, got %d", name, pwd)
		}
	}
	if !strings.Contains(strings.Join(imp.Env, "\n"), "HOME="+shadow) {
		t.Fatal("import must run under the shadow HOME")
	}
}

func TestCleanupKillsInFlightImportAndKeepsDataDirRemoved(t *testing.T) {
	bin := t.TempDir()
	pidFile := filepath.Join(t.TempDir(), "pid")
	script := `#!/bin/sh
if [ "$2" = "export" ]; then printf '{}'; exit 0; fi
echo $$ > "$FAKE_OPENCODE_PID"
mkdir -p "$HOME/.local/share/opencode"
cat > /dev/null
sleep 30
echo late > "$HOME/.local/share/opencode/opencode.db"
`
	os.WriteFile(filepath.Join(bin, "opencode"), []byte(script), 0700)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_OPENCODE_PID", pidFile)
	shadow, _ := target.Canonical(t.TempDir())
	data := filepath.Join(shadow, ".local", "share", "opencode")
	g := &guidanceVariant{shadowHome: shadow, host: "opencode", shadowDataDir: data, integration: "opencode-go"}
	done := make(chan error, 1)
	go func() { done <- g.importCredential() }()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(pidFile); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the fake import never started")
		}
		time.Sleep(20 * time.Millisecond)
	}
	started := time.Now()
	g.cleanup()
	if time.Since(started) > 8*time.Second {
		t.Fatal("cleanup must not wait for the 30 s import")
	}
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("an interrupted import must report an error")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("importCredential did not return after cleanup killed the import")
	}
	raw, _ := os.ReadFile(pidFile)
	pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatal("the import process is still alive after cleanup")
	}
	if _, err := os.Lstat(data); !os.IsNotExist(err) {
		t.Fatal("the shadow data directory must be gone")
	}
	// An import starting after cleanup already ran must refuse to start.
	if err := g.importCredential(); err == nil {
		t.Fatal("no import may start after cleanup")
	}
}
